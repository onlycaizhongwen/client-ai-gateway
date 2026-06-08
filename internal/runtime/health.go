package runtime

import (
	"time"

	"client-ai-gateway/internal/config"
	"client-ai-gateway/internal/providerhealth"
)

type HealthView struct {
	Status          string                `json:"status"`
	Daemon          DaemonHealth          `json:"daemon"`
	Config          ConfigHealth          `json:"config"`
	Stores          StoresHealth          `json:"stores"`
	ProviderMonitor ProviderMonitorHealth `json:"provider_monitor"`
	QuotaRuntime    QuotaRuntimeHealth    `json:"quota_runtime"`
	ModelRuntime    ComponentHealth       `json:"model_runtime"`
	MCPRuntime      ComponentHealth       `json:"mcp_runtime"`
}

type DaemonHealth struct {
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
}

type ConfigHealth struct {
	Status         string    `json:"status"`
	ListenAddr     string    `json:"listen_addr"`
	PolicyVersion  string    `json:"policy_version"`
	AppCount       int       `json:"app_count"`
	ProviderCount  int       `json:"provider_count"`
	LastReloadedAt time.Time `json:"last_reloaded_at"`
	ReloadCount    int       `json:"reload_count"`
}

type StoresHealth struct {
	Trace StoreHealth `json:"trace"`
	Audit StoreHealth `json:"audit"`
}

type StoreHealth struct {
	Status string `json:"status"`
	Path   string `json:"path"`
}

type ProviderMonitorHealth struct {
	Status    string `json:"status"`
	Total     int    `json:"total"`
	Healthy   int    `json:"healthy"`
	Degraded  int    `json:"degraded"`
	Unhealthy int    `json:"unhealthy"`
	Disabled  int    `json:"disabled"`
}

type QuotaRuntimeHealth struct {
	Status          string `json:"status"`
	Mode            string `json:"mode"`
	AppQuotaCount   int    `json:"app_quota_count"`
	EnabledAppRPM   int    `json:"enabled_app_rpm"`
	TotalAppRPM     int    `json:"total_app_rpm"`
	ProviderBudgets int    `json:"provider_budgets"`
	TokenLedgers    int    `json:"token_ledgers"`
	Reason          string `json:"reason,omitempty"`
}

type ComponentHealth struct {
	Status         string `json:"status"`
	Reason         string `json:"reason,omitempty"`
	Mode           string `json:"mode,omitempty"`
	ServerCount    int    `json:"server_count,omitempty"`
	EnabledServers int    `json:"enabled_servers,omitempty"`
	ToolCount      int    `json:"tool_count,omitempty"`
	EnabledTools   int    `json:"enabled_tools,omitempty"`
}

func (m *Manager) Health() HealthView {
	m.mu.RLock()
	snapshot := m.current
	startedAt := m.startedAt
	reloadedAt := m.reloadedAt
	reloadCount := m.reloadCount
	monitorRunning := m.stopHealth != nil
	m.mu.RUnlock()

	providerMonitor := providerMonitorHealth(snapshot, monitorRunning)
	status := "healthy"
	if providerMonitor.Unhealthy > 0 {
		status = "degraded"
	}
	view := HealthView{
		Status: status,
		Daemon: DaemonHealth{
			Status:    "healthy",
			StartedAt: startedAt,
		},
		Config: ConfigHealth{
			Status:         "loaded",
			ListenAddr:     snapshot.Config.ListenAddr,
			PolicyVersion:  snapshot.Config.PolicyVersion,
			AppCount:       len(snapshot.Config.Apps),
			ProviderCount:  len(snapshot.Config.Providers),
			LastReloadedAt: reloadedAt,
			ReloadCount:    reloadCount,
		},
		Stores: StoresHealth{
			Trace: StoreHealth{Status: "available", Path: snapshot.Config.TraceStorePath},
			Audit: StoreHealth{Status: "available", Path: snapshot.Config.AuditStorePath},
		},
		ProviderMonitor: providerMonitor,
		QuotaRuntime:    quotaRuntimeHealth(snapshot.Config.Quotas),
		ModelRuntime: ComponentHealth{
			Status: "not_configured",
			Reason: "model runtime management is planned for a later phase",
		},
		MCPRuntime: ComponentHealth{
			Status: "not_configured",
			Reason: "MCP runtime is not configured",
		},
	}
	view.MCPRuntime = mcpRuntimeHealth(snapshot.Config.MCPRuntime)
	return view
}

func quotaRuntimeHealth(quotas config.Quotas) QuotaRuntimeHealth {
	health := QuotaRuntimeHealth{
		Status:        "not_configured",
		Mode:          "app_rpm_in_memory",
		AppQuotaCount: len(quotas.Apps),
		Reason:        "quota runtime is not configured",
	}
	for _, app := range quotas.Apps {
		if app.RequestsPerMinute > 0 {
			health.EnabledAppRPM++
			health.TotalAppRPM += app.RequestsPerMinute
		}
	}
	if health.AppQuotaCount == 0 {
		return health
	}
	if health.EnabledAppRPM == 0 {
		health.Status = "configured"
		health.Reason = "app quotas are configured but RPM limits are disabled"
		return health
	}
	health.Status = "configured"
	health.Reason = "App RPM is enforced before provider routing; counters reset on daemon restart or config reload"
	return health
}

func providerMonitorHealth(snapshot Snapshot, running bool) ProviderMonitorHealth {
	status := "running"
	if !running || snapshot.Health == nil {
		status = "unavailable"
	}
	view := ProviderMonitorHealth{Status: status}
	if snapshot.Health == nil {
		return view
	}
	views := snapshot.Health.Views(snapshot.Config.Providers)
	view.Total = len(views)
	for _, provider := range views {
		switch provider.RuntimeStatus {
		case providerhealth.StatusHealthy:
			view.Healthy++
		case providerhealth.StatusDegraded:
			view.Degraded++
		case providerhealth.StatusUnhealthy:
			view.Unhealthy++
		case providerhealth.StatusDisabled:
			view.Disabled++
		}
	}
	return view
}

func mcpRuntimeHealth(runtime config.MCPRuntime) ComponentHealth {
	mode := runtime.Mode
	if mode == "" {
		mode = "manifest_only"
	}
	health := ComponentHealth{
		Status:         "not_configured",
		Reason:         "MCP runtime is not configured",
		Mode:           mode,
		ServerCount:    len(runtime.Servers),
		EnabledServers: countEnabledMCPServers(runtime),
		ToolCount:      countMCPTools(runtime),
		EnabledTools:   countEnabledMCPTools(runtime),
	}
	if !runtime.Enabled {
		if len(runtime.Servers) > 0 {
			health.Reason = "MCP runtime is configured but disabled"
		}
		return health
	}
	if len(runtime.Servers) == 0 {
		health.Status = "not_configured"
		health.Reason = "MCP runtime is enabled but no servers are configured"
		return health
	}
	health.Status = "configured"
	health.Reason = "MCP tool manifests are loaded; command execution is not enabled yet"
	if health.EnabledServers == 0 || health.EnabledTools == 0 {
		health.Status = "degraded"
		health.Reason = "MCP runtime has no enabled server or no enabled read-only tool"
	}
	return health
}

func countEnabledMCPServers(runtime config.MCPRuntime) int {
	total := 0
	for _, server := range runtime.Servers {
		if server.IsEnabled() {
			total++
		}
	}
	return total
}

func countMCPTools(runtime config.MCPRuntime) int {
	total := 0
	for _, server := range runtime.Servers {
		total += len(server.Tools)
	}
	return total
}

func countEnabledMCPTools(runtime config.MCPRuntime) int {
	total := 0
	for _, server := range runtime.Servers {
		if !server.IsEnabled() {
			continue
		}
		for _, tool := range server.Tools {
			if tool.IsEnabled() {
				total++
			}
		}
	}
	return total
}
