package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Config struct {
	ListenAddr            string     `json:"listen_addr"`
	TraceStorePath        string     `json:"trace_store_path"`
	AuditStorePath        string     `json:"audit_store_path"`
	TraceRetentionMax     int        `json:"trace_retention_max,omitempty"`
	AuditRetentionMax     int        `json:"audit_retention_max,omitempty"`
	TraceSnapshotEnabled  *bool      `json:"trace_snapshot_enabled,omitempty"`
	TraceRedactLabels     []string   `json:"trace_redact_labels,omitempty"`
	TraceSnapshotMaxChars int        `json:"trace_snapshot_max_chars,omitempty"`
	PolicyVersion         string     `json:"policy_version"`
	Apps                  []App      `json:"apps"`
	Providers             []Provider `json:"providers"`
	Tools                 []Tool     `json:"tools,omitempty"`
	MCPRuntime            MCPRuntime `json:"mcp_runtime,omitempty"`
	Quotas                Quotas     `json:"quotas,omitempty"`
	Policies              []Policy   `json:"policies"`
}

type App struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Token  string   `json:"token"`
	Grants []string `json:"grants"`
}

type Provider struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Class     string   `json:"class"`
	Adapter   string   `json:"adapter,omitempty"`
	BaseURL   string   `json:"base_url,omitempty"`
	APIKeyEnv string   `json:"api_key_env,omitempty"`
	Models    []string `json:"models"`
	Healthy   bool     `json:"healthy"`
	Enabled   *bool    `json:"enabled,omitempty"`
}

type Tool struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Adapter         string         `json:"adapter"`
	Description     string         `json:"description,omitempty"`
	ReadOnly        bool           `json:"read_only"`
	RiskLevel       string         `json:"risk_level,omitempty"`
	Scopes          []string       `json:"scopes,omitempty"`
	InputSchema     map[string]any `json:"input_schema,omitempty"`
	OutputSchema    map[string]any `json:"output_schema,omitempty"`
	SandboxRequired bool           `json:"sandbox_required,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
}

type MCPRuntime struct {
	Enabled bool        `json:"enabled"`
	Mode    string      `json:"mode,omitempty"`
	Servers []MCPServer `json:"servers,omitempty"`
}

type MCPServer struct {
	ID      string    `json:"id"`
	Name    string    `json:"name,omitempty"`
	Enabled *bool     `json:"enabled,omitempty"`
	Tools   []MCPTool `json:"tools,omitempty"`
}

type MCPTool struct {
	ID              string         `json:"id"`
	Name            string         `json:"name,omitempty"`
	Description     string         `json:"description,omitempty"`
	ReadOnly        bool           `json:"read_only"`
	RiskLevel       string         `json:"risk_level,omitempty"`
	Scopes          []string       `json:"scopes,omitempty"`
	InputSchema     map[string]any `json:"input_schema,omitempty"`
	OutputSchema    map[string]any `json:"output_schema,omitempty"`
	SandboxRequired bool           `json:"sandbox_required,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
}

type Quotas struct {
	Apps      []AppQuota      `json:"apps,omitempty"`
	Providers []ProviderQuota `json:"providers,omitempty"`
}

type AppQuota struct {
	AppID             string `json:"app_id"`
	RequestsPerMinute int    `json:"requests_per_minute,omitempty"`
}

type ProviderQuota struct {
	ProviderID        string `json:"provider_id"`
	RequestsPerMinute int    `json:"requests_per_minute,omitempty"`
}

type Policy struct {
	ID              string   `json:"id"`
	Priority        int      `json:"priority,omitempty"`
	Effect          string   `json:"effect"`
	Reason          string   `json:"reason"`
	AppIDs          []string `json:"app_ids,omitempty"`
	RequestTypes    []string `json:"request_types,omitempty"`
	Models          []string `json:"models,omitempty"`
	ProviderClasses []string `json:"provider_classes,omitempty"`
	DataLabels      []string `json:"data_labels,omitempty"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	clean := strings.TrimPrefix(string(data), "\ufeff")
	if err := json.Unmarshal([]byte(clean), &cfg); err != nil {
		return Config{}, err
	}
	if cfg.ListenAddr == "" {
		return Config{}, fmt.Errorf("listen_addr is required")
	}
	if cfg.PolicyVersion == "" {
		cfg.PolicyVersion = "local-dev"
	}
	if cfg.TraceStorePath == "" {
		cfg.TraceStorePath = "data/traces.jsonl"
	}
	if cfg.AuditStorePath == "" {
		cfg.AuditStorePath = "data/audit.jsonl"
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) AppByToken(token string) (App, bool) {
	for _, app := range c.Apps {
		if app.Token == token {
			return app, true
		}
	}
	return App{}, false
}

func (c Config) Validate() error {
	if _, _, err := net.SplitHostPort(c.ListenAddr); err != nil {
		return fmt.Errorf("listen_addr must be host:port: %w", err)
	}
	if strings.TrimSpace(c.TraceStorePath) == "" {
		return fmt.Errorf("trace_store_path is required")
	}
	if strings.TrimSpace(c.AuditStorePath) == "" {
		return fmt.Errorf("audit_store_path is required")
	}
	if c.TraceRetentionMax < 0 {
		return fmt.Errorf("trace_retention_max must be >= 0")
	}
	if c.AuditRetentionMax < 0 {
		return fmt.Errorf("audit_retention_max must be >= 0")
	}
	if c.TraceSnapshotMaxChars < 0 {
		return fmt.Errorf("trace_snapshot_max_chars must be >= 0")
	}
	redactLabels := map[string]struct{}{}
	for i, label := range c.TraceRedactLabels {
		normalized := strings.TrimSpace(label)
		if normalized == "" {
			return fmt.Errorf("trace_redact_labels[%d] must not be empty", i)
		}
		key := strings.ToLower(normalized)
		if _, ok := redactLabels[key]; ok {
			return fmt.Errorf("trace_redact_labels contains duplicate label %q", normalized)
		}
		redactLabels[key] = struct{}{}
	}
	if len(c.Apps) == 0 {
		return fmt.Errorf("at least one app is required")
	}
	appIDs := map[string]struct{}{}
	tokens := map[string]struct{}{}
	for i, app := range c.Apps {
		if app.ID == "" {
			return fmt.Errorf("apps[%d].id is required", i)
		}
		if _, ok := appIDs[app.ID]; ok {
			return fmt.Errorf("duplicate app id %q", app.ID)
		}
		appIDs[app.ID] = struct{}{}
		if app.Token == "" {
			return fmt.Errorf("apps[%d].token is required", i)
		}
		if _, ok := tokens[app.Token]; ok {
			return fmt.Errorf("duplicate app token for app %q", app.ID)
		}
		tokens[app.Token] = struct{}{}
		if len(app.Grants) == 0 {
			return fmt.Errorf("apps[%d].grants must not be empty", i)
		}
		for _, grant := range app.Grants {
			if !validGrant(grant) {
				return fmt.Errorf("apps[%d].grants contains unsupported grant %q", i, grant)
			}
		}
	}
	if len(c.Providers) == 0 {
		return fmt.Errorf("at least one provider is required")
	}
	providerIDs := map[string]struct{}{}
	for i, provider := range c.Providers {
		if provider.ID == "" {
			return fmt.Errorf("providers[%d].id is required", i)
		}
		if _, ok := providerIDs[provider.ID]; ok {
			return fmt.Errorf("duplicate provider id %q", provider.ID)
		}
		providerIDs[provider.ID] = struct{}{}
		if provider.Class != "local" && provider.Class != "cloud" {
			return fmt.Errorf("providers[%d].class must be local or cloud", i)
		}
		if provider.Adapter != "" && provider.Adapter != "mock" && provider.Adapter != "openai-compatible" {
			return fmt.Errorf("providers[%d].adapter %q is unsupported", i, provider.Adapter)
		}
		if provider.Adapter == "openai-compatible" && strings.TrimSpace(provider.BaseURL) == "" {
			return fmt.Errorf("providers[%d].base_url is required for openai-compatible adapter", i)
		}
		if len(provider.Models) == 0 {
			return fmt.Errorf("providers[%d].models must not be empty", i)
		}
		models := map[string]struct{}{}
		for _, model := range provider.Models {
			if model == "" {
				return fmt.Errorf("providers[%d].models contains empty model", i)
			}
			if _, ok := models[model]; ok {
				return fmt.Errorf("providers[%d].models contains duplicate model %q", i, model)
			}
			models[model] = struct{}{}
		}
	}
	toolIDs := map[string]struct{}{}
	for i, tool := range c.Tools {
		if tool.ID == "" {
			return fmt.Errorf("tools[%d].id is required", i)
		}
		if _, ok := toolIDs[tool.ID]; ok {
			return fmt.Errorf("duplicate tool id %q", tool.ID)
		}
		toolIDs[tool.ID] = struct{}{}
		if tool.Adapter == "" {
			return fmt.Errorf("tools[%d].adapter is required", i)
		}
		if tool.Adapter != "runtime-health" {
			return fmt.Errorf("tools[%d].adapter %q is unsupported", i, tool.Adapter)
		}
		if !tool.ReadOnly {
			return fmt.Errorf("tools[%d].read_only must be true for phase 2 read-only MVP", i)
		}
		if tool.RiskLevel == "" {
			tool.RiskLevel = "low"
		}
		if !validToolRiskLevel(tool.RiskLevel) {
			return fmt.Errorf("tools[%d].risk_level %q is unsupported", i, tool.RiskLevel)
		}
		if len(tool.Scopes) == 0 {
			return fmt.Errorf("tools[%d].scopes must not be empty", i)
		}
		scopes := map[string]struct{}{}
		for _, scope := range tool.Scopes {
			if !validScopeName(scope) {
				return fmt.Errorf("tools[%d].scopes contains invalid scope %q", i, scope)
			}
			if _, ok := scopes[scope]; ok {
				return fmt.Errorf("tools[%d].scopes contains duplicate scope %q", i, scope)
			}
			scopes[scope] = struct{}{}
		}
		if tool.SandboxRequired {
			return fmt.Errorf("tools[%d].sandbox_required is not supported before sandbox runtime is enabled", i)
		}
	}
	if err := validateMCPRuntime(c.MCPRuntime); err != nil {
		return err
	}
	if err := validateQuotas(c.Quotas, appIDs, providerIDs); err != nil {
		return err
	}
	policyIDs := map[string]struct{}{}
	for i, rule := range c.Policies {
		if rule.ID == "" {
			return fmt.Errorf("policies[%d].id is required", i)
		}
		if _, ok := policyIDs[rule.ID]; ok {
			return fmt.Errorf("duplicate policy id %q", rule.ID)
		}
		policyIDs[rule.ID] = struct{}{}
		if !validPolicyEffect(rule.Effect) {
			return fmt.Errorf("policies[%d].effect %q is unsupported", i, rule.Effect)
		}
		if rule.Reason == "" {
			return fmt.Errorf("policies[%d].reason is required", i)
		}
		if err := validatePolicyValues(i, rule); err != nil {
			return err
		}
	}
	return nil
}

func validateQuotas(quotas Quotas, appIDs map[string]struct{}, providerIDs map[string]struct{}) error {
	seenApps := map[string]struct{}{}
	for i, quota := range quotas.Apps {
		if quota.AppID == "" {
			return fmt.Errorf("quotas.apps[%d].app_id is required", i)
		}
		if _, ok := appIDs[quota.AppID]; !ok {
			return fmt.Errorf("quotas.apps[%d].app_id %q does not reference a configured app", i, quota.AppID)
		}
		if _, ok := seenApps[quota.AppID]; ok {
			return fmt.Errorf("duplicate app quota for app %q", quota.AppID)
		}
		seenApps[quota.AppID] = struct{}{}
		if quota.RequestsPerMinute < 0 {
			return fmt.Errorf("quotas.apps[%d].requests_per_minute must be >= 0", i)
		}
	}
	seenProviders := map[string]struct{}{}
	for i, quota := range quotas.Providers {
		if quota.ProviderID == "" {
			return fmt.Errorf("quotas.providers[%d].provider_id is required", i)
		}
		if _, ok := providerIDs[quota.ProviderID]; !ok {
			return fmt.Errorf("quotas.providers[%d].provider_id %q does not reference a configured provider", i, quota.ProviderID)
		}
		if _, ok := seenProviders[quota.ProviderID]; ok {
			return fmt.Errorf("duplicate provider quota for provider %q", quota.ProviderID)
		}
		seenProviders[quota.ProviderID] = struct{}{}
		if quota.RequestsPerMinute < 0 {
			return fmt.Errorf("quotas.providers[%d].requests_per_minute must be >= 0", i)
		}
	}
	return nil
}

func (c Config) IsTraceSnapshotEnabled() bool {
	return c.TraceSnapshotEnabled == nil || *c.TraceSnapshotEnabled
}

func (c Config) EffectiveTraceRedactLabels() []string {
	if len(c.TraceRedactLabels) == 0 {
		return []string{"sensitive"}
	}
	out := make([]string, 0, len(c.TraceRedactLabels))
	for _, label := range c.TraceRedactLabels {
		label = strings.TrimSpace(label)
		if label != "" {
			out = append(out, label)
		}
	}
	return out
}

func validGrant(grant string) bool {
	if strings.HasPrefix(grant, "tool:") {
		return validScopeName(strings.TrimPrefix(grant, "tool:"))
	}
	switch grant {
	case "chat", "embedding", "tool", "admin":
		return true
	default:
		return false
	}
}

func validToolRiskLevel(level string) bool {
	switch level {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

func validScopeName(scope string) bool {
	if scope == "" {
		return false
	}
	for _, ch := range scope {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '.' || ch == '_' || ch == '-' {
			continue
		}
		return false
	}
	return true
}

func (p Provider) IsEnabled() bool {
	return p.Enabled == nil || *p.Enabled
}

func (t Tool) IsEnabled() bool {
	return t.Enabled == nil || *t.Enabled
}

func (s MCPServer) IsEnabled() bool {
	return s.Enabled == nil || *s.Enabled
}

func (t MCPTool) IsEnabled() bool {
	return t.Enabled == nil || *t.Enabled
}

func validPolicyEffect(effect string) bool {
	switch effect {
	case "allow", "deny", "deny_cloud_for_sensitive", "force_local":
		return true
	default:
		return false
	}
}

func validateMCPRuntime(runtime MCPRuntime) error {
	if !runtime.Enabled && len(runtime.Servers) == 0 {
		return nil
	}
	mode := runtime.Mode
	if mode == "" {
		mode = "manifest_only"
	}
	if mode != "manifest_only" && mode != "disabled" {
		return fmt.Errorf("mcp_runtime.mode %q is not supported before sandbox runtime is enabled", runtime.Mode)
	}
	if runtime.Enabled && mode == "disabled" {
		return fmt.Errorf("mcp_runtime.mode disabled requires enabled=false")
	}
	serverIDs := map[string]struct{}{}
	toolIDs := map[string]struct{}{}
	for i, server := range runtime.Servers {
		if server.ID == "" {
			return fmt.Errorf("mcp_runtime.servers[%d].id is required", i)
		}
		if _, ok := serverIDs[server.ID]; ok {
			return fmt.Errorf("duplicate mcp server id %q", server.ID)
		}
		serverIDs[server.ID] = struct{}{}
		for j, tool := range server.Tools {
			if tool.ID == "" {
				return fmt.Errorf("mcp_runtime.servers[%d].tools[%d].id is required", i, j)
			}
			if _, ok := toolIDs[tool.ID]; ok {
				return fmt.Errorf("duplicate mcp tool id %q", tool.ID)
			}
			toolIDs[tool.ID] = struct{}{}
			if !tool.ReadOnly {
				return fmt.Errorf("mcp_runtime.servers[%d].tools[%d].read_only must be true for read-only MCP placeholder", i, j)
			}
			if tool.RiskLevel == "" {
				tool.RiskLevel = "low"
			}
			if !validToolRiskLevel(tool.RiskLevel) {
				return fmt.Errorf("mcp_runtime.servers[%d].tools[%d].risk_level %q is unsupported", i, j, tool.RiskLevel)
			}
			if len(tool.Scopes) == 0 {
				return fmt.Errorf("mcp_runtime.servers[%d].tools[%d].scopes must not be empty", i, j)
			}
			scopes := map[string]struct{}{}
			for _, scope := range tool.Scopes {
				if !validScopeName(scope) {
					return fmt.Errorf("mcp_runtime.servers[%d].tools[%d].scopes contains invalid scope %q", i, j, scope)
				}
				if _, ok := scopes[scope]; ok {
					return fmt.Errorf("mcp_runtime.servers[%d].tools[%d].scopes contains duplicate scope %q", i, j, scope)
				}
				scopes[scope] = struct{}{}
			}
			if tool.SandboxRequired {
				return fmt.Errorf("mcp_runtime.servers[%d].tools[%d].sandbox_required is not supported before sandbox runtime is enabled", i, j)
			}
		}
	}
	return nil
}

func validatePolicyValues(index int, rule Policy) error {
	for _, requestType := range rule.RequestTypes {
		if requestType != "chat" && requestType != "embedding" && requestType != "tool" {
			return fmt.Errorf("policies[%d].request_types contains unsupported request type %q", index, requestType)
		}
	}
	for _, providerClass := range rule.ProviderClasses {
		if providerClass != "local" && providerClass != "cloud" {
			return fmt.Errorf("policies[%d].provider_classes contains unsupported provider class %q", index, providerClass)
		}
	}
	return nil
}
