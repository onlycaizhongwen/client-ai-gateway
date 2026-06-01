package providerhealth

import (
	"context"
	"fmt"
	"time"

	"client-ai-gateway/internal/adapters"
	"client-ai-gateway/internal/config"
)

type Monitor struct {
	providers        []config.Provider
	registry         *adapters.Registry
	store            *Store
	interval         time.Duration
	timeout          time.Duration
	unhealthyAfter   int
	consecutiveFails map[string]int
}

type MonitorConfig struct {
	Interval       time.Duration
	Timeout        time.Duration
	UnhealthyAfter int
}

func NewMonitor(providers []config.Provider, registry *adapters.Registry, store *Store, cfg MonitorConfig) *Monitor {
	interval := cfg.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	unhealthyAfter := cfg.UnhealthyAfter
	if unhealthyAfter <= 0 {
		unhealthyAfter = 3
	}
	return &Monitor{
		providers:        providers,
		registry:         registry,
		store:            store,
		interval:         interval,
		timeout:          timeout,
		unhealthyAfter:   unhealthyAfter,
		consecutiveFails: map[string]int{},
	}
}

func (m *Monitor) Start(ctx context.Context) {
	m.CheckOnce(ctx)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.CheckOnce(ctx)
		}
	}
}

func (m *Monitor) CheckOnce(ctx context.Context) {
	for _, provider := range m.providers {
		err := CheckProviderWithTimeout(ctx, provider, m.registry, m.store, m.timeout)
		if err != nil {
			m.recordFailure(provider.ID, err)
			continue
		}
		m.consecutiveFails[provider.ID] = 0
		m.store.Set(provider.ID, StatusHealthy, "last probe succeeded")
	}
}

func (m *Monitor) recordFailure(providerID string, err error) {
	m.consecutiveFails[providerID]++
	reason := fmt.Sprintf("probe failed %d/%d: %v", m.consecutiveFails[providerID], m.unhealthyAfter, err)
	status := StatusDegraded
	if m.consecutiveFails[providerID] >= m.unhealthyAfter {
		status = StatusUnhealthy
	}
	m.store.Set(providerID, status, reason)
}

func CheckProvider(ctx context.Context, provider config.Provider, registry *adapters.Registry, store *Store) error {
	return CheckProviderWithTimeout(ctx, provider, registry, store, 3*time.Second)
}

func CheckProviderWithTimeout(ctx context.Context, provider config.Provider, registry *adapters.Registry, store *Store, timeout time.Duration) error {
	if !provider.IsEnabled() {
		store.Set(provider.ID, StatusDisabled, "provider disabled by config")
		return nil
	}
	if !provider.Healthy {
		store.Set(provider.ID, StatusUnhealthy, "provider marked unhealthy in config")
		return nil
	}
	adapter, ok := registry.Get(provider.ID)
	if !ok {
		return fmt.Errorf("adapter not registered")
	}
	checker, ok := adapter.(adapters.HealthChecker)
	if !ok {
		store.Set(provider.ID, StatusDegraded, "adapter does not implement health check")
		return nil
	}
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	err := checker.CheckHealth(checkCtx)
	cancel()
	if err != nil {
		return err
	}
	store.Set(provider.ID, StatusHealthy, "last probe succeeded")
	return nil
}
