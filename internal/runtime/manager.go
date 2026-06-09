package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"client-ai-gateway/internal/adapters"
	"client-ai-gateway/internal/config"
	"client-ai-gateway/internal/core"
	"client-ai-gateway/internal/fallback"
	"client-ai-gateway/internal/policy"
	"client-ai-gateway/internal/providerhealth"
	"client-ai-gateway/internal/quota"
	"client-ai-gateway/internal/router"
	"client-ai-gateway/internal/tools"
	"client-ai-gateway/internal/trace"
)

type Snapshot struct {
	Config   config.Config
	Pipeline *core.Pipeline
	Health   *providerhealth.Store
	Registry *adapters.Registry
	Tools    *tools.Registry
}

type Manager struct {
	mu            sync.RWMutex
	configWriteMu sync.Mutex
	configPath    string
	traceStore    trace.Store
	current       Snapshot
	stopHealth    context.CancelFunc
	startedAt     time.Time
	reloadedAt    time.Time
	reloadCount   int
}

type RPMQuotaChange struct {
	OldRequestsPerMinute int
	RequestsPerMinute    int
}

func NewManager(configPath string, traceStore trace.Store) (*Manager, error) {
	manager := &Manager{configPath: configPath, traceStore: traceStore, startedAt: time.Now().UTC()}
	if err := manager.Reload(); err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *Manager) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

func (m *Manager) Reload() error {
	m.configWriteMu.Lock()
	defer m.configWriteMu.Unlock()
	return m.reloadLocked()
}

func (m *Manager) reloadLocked() error {
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return err
	}
	snapshot, stopHealth, err := m.buildSnapshot(cfg)
	if err != nil {
		return err
	}

	m.mu.Lock()
	oldStop := m.stopHealth
	m.current = snapshot
	m.stopHealth = stopHealth
	m.reloadedAt = time.Now().UTC()
	m.reloadCount++
	m.mu.Unlock()
	if oldStop != nil {
		oldStop()
	}
	return nil
}

func (m *Manager) SetProviderEnabled(providerID string, enabled bool) error {
	return m.updateConfigAndReload(func(cfg *config.Config) error {
		for i := range cfg.Providers {
			if cfg.Providers[i].ID == providerID {
				cfg.Providers[i].Enabled = &enabled
				return nil
			}
		}
		return fmt.Errorf("provider %q not found", providerID)
	})
}

func (m *Manager) SetProviderRPMQuota(providerID string, requestsPerMinute int) (RPMQuotaChange, error) {
	if requestsPerMinute < 0 {
		return RPMQuotaChange{}, fmt.Errorf("provider %q requests_per_minute must be >= 0", providerID)
	}
	change := RPMQuotaChange{RequestsPerMinute: requestsPerMinute}
	err := m.updateConfigAndReload(func(cfg *config.Config) error {
		if !providerExists(cfg.Providers, providerID) {
			return fmt.Errorf("provider %q not found", providerID)
		}
		change.OldRequestsPerMinute = providerQuotaRPM(cfg.Quotas.Providers, providerID)
		cfg.Quotas.Providers = setProviderQuotaRPM(cfg.Quotas.Providers, providerID, requestsPerMinute)
		return nil
	})
	return change, err
}

func (m *Manager) SetAppRPMQuota(appID string, requestsPerMinute int) (RPMQuotaChange, error) {
	if requestsPerMinute < 0 {
		return RPMQuotaChange{}, fmt.Errorf("app %q requests_per_minute must be >= 0", appID)
	}
	change := RPMQuotaChange{RequestsPerMinute: requestsPerMinute}
	err := m.updateConfigAndReload(func(cfg *config.Config) error {
		if !appExists(cfg.Apps, appID) {
			return fmt.Errorf("app %q not found", appID)
		}
		change.OldRequestsPerMinute = appQuotaRPM(cfg.Quotas.Apps, appID)
		cfg.Quotas.Apps = setAppQuotaRPM(cfg.Quotas.Apps, appID, requestsPerMinute)
		return nil
	})
	return change, err
}

func (m *Manager) updateConfigAndReload(update func(*config.Config) error) error {
	m.configWriteMu.Lock()
	defer m.configWriteMu.Unlock()

	cfg, err := config.Load(m.configPath)
	if err != nil {
		return err
	}
	if err := update(&cfg); err != nil {
		return err
	}
	return m.writeConfigAndReload(cfg)
}

func appExists(apps []config.App, appID string) bool {
	for _, app := range apps {
		if app.ID == appID {
			return true
		}
	}
	return false
}

func providerExists(providers []config.Provider, providerID string) bool {
	for _, provider := range providers {
		if provider.ID == providerID {
			return true
		}
	}
	return false
}

func appQuotaRPM(quotas []config.AppQuota, appID string) int {
	for _, quota := range quotas {
		if quota.AppID == appID {
			return quota.RequestsPerMinute
		}
	}
	return 0
}

func providerQuotaRPM(quotas []config.ProviderQuota, providerID string) int {
	for _, quota := range quotas {
		if quota.ProviderID == providerID {
			return quota.RequestsPerMinute
		}
	}
	return 0
}

func setAppQuotaRPM(quotas []config.AppQuota, appID string, requestsPerMinute int) []config.AppQuota {
	for i := range quotas {
		if quotas[i].AppID != appID {
			continue
		}
		if requestsPerMinute == 0 {
			return append(quotas[:i], quotas[i+1:]...)
		}
		quotas[i].RequestsPerMinute = requestsPerMinute
		return quotas
	}
	if requestsPerMinute > 0 {
		return append(quotas, config.AppQuota{
			AppID:             appID,
			RequestsPerMinute: requestsPerMinute,
		})
	}
	return quotas
}

func setProviderQuotaRPM(quotas []config.ProviderQuota, providerID string, requestsPerMinute int) []config.ProviderQuota {
	for i := range quotas {
		if quotas[i].ProviderID != providerID {
			continue
		}
		if requestsPerMinute == 0 {
			return append(quotas[:i], quotas[i+1:]...)
		}
		quotas[i].RequestsPerMinute = requestsPerMinute
		return quotas
	}
	if requestsPerMinute > 0 {
		return append(quotas, config.ProviderQuota{
			ProviderID:        providerID,
			RequestsPerMinute: requestsPerMinute,
		})
	}
	return quotas
}

func (m *Manager) ProbeProvider(ctx context.Context, providerID string) (providerhealth.View, error) {
	snapshot := m.Snapshot()
	for _, provider := range snapshot.Config.Providers {
		if provider.ID == providerID {
			err := providerhealth.CheckProvider(ctx, provider, snapshot.Registry, snapshot.Health)
			for _, view := range snapshot.Health.Views(snapshot.Config.Providers) {
				if view.ID == providerID {
					return view, err
				}
			}
			return providerhealth.View{}, err
		}
	}
	return providerhealth.View{}, fmt.Errorf("provider %q not found", providerID)
}

func (m *Manager) writeConfigAndReload(cfg config.Config) error {
	encoded, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(m.configPath, append(encoded, '\n'), 0644); err != nil {
		return err
	}
	return m.reloadLocked()
}

func (m *Manager) Close() {
	m.mu.Lock()
	stopHealth := m.stopHealth
	m.stopHealth = nil
	m.mu.Unlock()
	if stopHealth != nil {
		stopHealth()
	}
}

func (m *Manager) buildSnapshot(cfg config.Config) (Snapshot, context.CancelFunc, error) {
	adapterRegistry, err := adapters.NewRegistryFromConfig(cfg.Providers)
	if err != nil {
		return Snapshot{}, nil, fmt.Errorf("build provider adapters: %w", err)
	}
	healthStore := providerhealth.NewStore(cfg.Providers)
	healthMonitor := providerhealth.NewMonitor(cfg.Providers, adapterRegistry, healthStore, providerhealth.MonitorConfig{})
	monitorCtx, stopMonitor := context.WithCancel(context.Background())
	go healthMonitor.Start(monitorCtx)

	toolRegistry, err := tools.NewRegistryFromConfig(cfg.Tools, func() any {
		return m.Health()
	})
	if err != nil {
		stopMonitor()
		return Snapshot{}, nil, fmt.Errorf("build tools: %w", err)
	}

	pipeline := core.NewPipeline(core.Dependencies{
		Config:     cfg,
		Policy:     policy.NewEngine(cfg.PolicyVersion, cfg.Policies),
		Router:     router.NewWithHealth(cfg.Providers, healthStore),
		Fallback:   fallback.NewManager(),
		Adapters:   adapterRegistry,
		Quota:      quota.NewLimiter(cfg.Quotas),
		TraceStore: m.traceStore,
	})
	return Snapshot{
		Config:   cfg,
		Pipeline: pipeline,
		Health:   healthStore,
		Registry: adapterRegistry,
		Tools:    toolRegistry,
	}, stopMonitor, nil
}
