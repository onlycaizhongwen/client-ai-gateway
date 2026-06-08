package access

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"client-ai-gateway/internal/audit"
	"client-ai-gateway/internal/config"
	"client-ai-gateway/internal/core"
	"client-ai-gateway/internal/policy"
	"client-ai-gateway/internal/providerhealth"
	"client-ai-gateway/internal/router"
	gatewayruntime "client-ai-gateway/internal/runtime"
	"client-ai-gateway/internal/tools"
	"client-ai-gateway/internal/trace"
)

type Handler struct {
	cfg            config.Config
	pipeline       *core.Pipeline
	traces         trace.Store
	providerHealth *providerhealth.Store
	runtime        *gatewayruntime.Manager
	audit          audit.Store
	logger         *slog.Logger
}

func NewHandler(cfg config.Config, pipeline *core.Pipeline, traces trace.Store) *Handler {
	return &Handler{cfg: cfg, pipeline: pipeline, traces: traces, logger: slog.Default()}
}

func NewRuntimeHandler(manager *gatewayruntime.Manager, traces trace.Store) *Handler {
	snapshot := manager.Snapshot()
	return &Handler{
		cfg:            snapshot.Config,
		pipeline:       snapshot.Pipeline,
		traces:         traces,
		providerHealth: snapshot.Health,
		runtime:        manager,
		logger:         slog.Default(),
	}
}

func (h *Handler) WithAudit(auditStore audit.Store) *Handler {
	h.audit = auditStore
	return h
}

func (h *Handler) WithProviderHealth(health *providerhealth.Store) *Handler {
	h.providerHealth = health
	return h
}

func (h *Handler) WithLogger(logger *slog.Logger) *Handler {
	if logger != nil {
		h.logger = logger
	}
	return h
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /v1/chat/completions", h.chat)
	mux.HandleFunc("GET /console", h.console)
	mux.HandleFunc("GET /gateway/v1/traces", h.traceList)
	mux.HandleFunc("GET /gateway/v1/traces/export", h.traceExport)
	mux.HandleFunc("GET /gateway/v1/traces/", h.traceByID)
	mux.HandleFunc("GET /gateway/v1/providers", h.providers)
	mux.HandleFunc("GET /gateway/v1/providers/export", h.providersExport)
	mux.HandleFunc("GET /gateway/v1/models", h.models)
	mux.HandleFunc("GET /gateway/v1/models/export", h.modelsExport)
	mux.HandleFunc("GET /gateway/v1/runtime/health", h.runtimeHealth)
	mux.HandleFunc("GET /gateway/v1/apps", h.appsList)
	mux.HandleFunc("GET /gateway/v1/apps/export", h.appsExport)
	mux.HandleFunc("GET /gateway/v1/grants", h.grantsList)
	mux.HandleFunc("GET /gateway/v1/grants/export", h.grantsExport)
	mux.HandleFunc("GET /gateway/v1/policies", h.policiesList)
	mux.HandleFunc("GET /gateway/v1/policies/export", h.policiesExport)
	mux.HandleFunc("GET /gateway/v1/policies/", h.policyByID)
	mux.HandleFunc("GET /gateway/v1/tools", h.toolsList)
	mux.HandleFunc("GET /gateway/v1/tools/export", h.toolsExport)
	mux.HandleFunc("POST /gateway/v1/tools/", h.toolInvoke)
	mux.HandleFunc("GET /gateway/v1/mcp/servers", h.mcpServers)
	mux.HandleFunc("GET /gateway/v1/mcp/servers/export", h.mcpServersExport)
	mux.HandleFunc("POST /gateway/v1/providers/", h.providerAction)
	mux.HandleFunc("GET /gateway/v1/audit/events", h.auditList)
	mux.HandleFunc("GET /gateway/v1/audit/events/export", h.auditExport)
	mux.HandleFunc("POST /gateway/v1/config/reload", h.configReload)
	mux.HandleFunc("POST /gateway/v1/policy/dry-run", h.policyDryRun)
	mux.HandleFunc("POST /gateway/v1/routing/explain", h.routingExplain)
	mux.HandleFunc("POST /gateway/v1/access/dry-run", h.accessDryRun)
	return h.accessLog(mux)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) chat(w http.ResponseWriter, r *http.Request) {
	var req core.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "", "invalid_request", err.Error())
		return
	}
	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "", "invalid_request", "model is required")
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "", "invalid_request", "messages are required")
		return
	}

	snapshot := h.snapshot()
	resp, err := snapshot.Pipeline.Chat(r.Context(), bearerToken(r.Header.Get("Authorization")), req)
	logInfo := getLogInfo(r)
	if err != nil {
		status := http.StatusBadGateway
		code := "gateway_error"
		traceID := resp.TraceID
		if errors.Is(err, core.ErrUnauthorized) {
			status = http.StatusUnauthorized
			code = "unauthorized"
		}
		var gatewayErr *core.GatewayError
		if errors.As(err, &gatewayErr) {
			traceID = gatewayErr.TraceID
			logInfo.traceID = gatewayErr.TraceID
			logInfo.appID = gatewayErr.AppID
			code = gatewayErr.Code
			if gatewayErr.Code == "policy_denied" {
				status = http.StatusForbidden
			} else if gatewayErr.Code == "rate_limited" {
				status = http.StatusTooManyRequests
			}
		}
		writeError(w, status, traceID, code, err.Error())
		return
	}
	logInfo.traceID = resp.TraceID
	logInfo.appID = resp.AppID
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) traceByID(w http.ResponseWriter, r *http.Request) {
	traceID := strings.TrimPrefix(r.URL.Path, "/gateway/v1/traces/")
	if traceID == "" {
		writeError(w, http.StatusBadRequest, "", "invalid_request", "trace id is required")
		return
	}
	record, ok := h.traces.Get(traceID)
	if !ok {
		writeError(w, http.StatusNotFound, traceID, "not_found", "trace not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (h *Handler) traceList(w http.ResponseWriter, r *http.Request) {
	limit, ok := intQuery(w, r, "limit", 100)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	page := h.traces.Page(trace.ListQuery{
		Offset:     offset,
		Limit:      limit,
		Status:     r.URL.Query().Get("status"),
		AppID:      r.URL.Query().Get("app_id"),
		ProviderID: r.URL.Query().Get("provider_id"),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"traces": page.Items,
		"total":  page.Total,
		"offset": page.Offset,
		"limit":  page.Limit,
	})
}

func (h *Handler) traceExport(w http.ResponseWriter, r *http.Request) {
	limit, ok := intQuery(w, r, "limit", 500)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	page := h.traces.Page(trace.ListQuery{
		Offset:     offset,
		Limit:      limit,
		Status:     r.URL.Query().Get("status"),
		AppID:      r.URL.Query().Get("app_id"),
		ProviderID: r.URL.Query().Get("provider_id"),
	})
	writeJSONL(w, "traces.jsonl", page.Items)
}

func (h *Handler) providers(w http.ResponseWriter, r *http.Request) {
	limit, ok := intQuery(w, r, "limit", 100)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	views := h.providerViews(r)
	total := len(views)
	offset, limit = normalizePage(offset, limit)
	pagedProviders := []providerhealth.View{}
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		pagedProviders = views[offset:end]
	}
	query := r.URL.Query()
	writeJSON(w, http.StatusOK, map[string]any{
		"providers": pagedProviders,
		"total":     total,
		"offset":    offset,
		"limit":     limit,
		"filters": map[string]string{
			"provider_id":    query.Get("provider_id"),
			"class":          query.Get("class"),
			"enabled":        query.Get("enabled"),
			"runtime_status": query.Get("runtime_status"),
		},
	})
}

func (h *Handler) providerViews(r *http.Request) []providerhealth.View {
	snapshot := h.snapshot()
	var views []providerhealth.View
	if snapshot.Health != nil {
		views = snapshot.Health.Views(snapshot.Config.Providers)
	} else {
		views = providerhealth.NewStore(snapshot.Config.Providers).Views(snapshot.Config.Providers)
	}
	query := r.URL.Query()
	providerFilter := query.Get("provider_id")
	classFilter := query.Get("class")
	enabledFilter := query.Get("enabled")
	statusFilter := query.Get("runtime_status")
	filtered := make([]providerhealth.View, 0, len(views))
	for _, view := range views {
		if providerFilter != "" && view.ID != providerFilter {
			continue
		}
		if classFilter != "" && view.Class != classFilter {
			continue
		}
		if enabledFilter != "" && strconv.FormatBool(view.Enabled) != enabledFilter {
			continue
		}
		if statusFilter != "" && view.RuntimeStatus != statusFilter {
			continue
		}
		filtered = append(filtered, view)
	}
	return filtered
}

func (h *Handler) providersExport(w http.ResponseWriter, r *http.Request) {
	writeJSONL(w, "providers.jsonl", h.providerViews(r))
}

type providerEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

type modelView struct {
	Model         string `json:"model"`
	ProviderID    string `json:"provider_id"`
	ProviderClass string `json:"provider_class"`
	Enabled       bool   `json:"enabled"`
	RuntimeStatus string `json:"runtime_status"`
	Available     bool   `json:"available"`
}

func (h *Handler) models(w http.ResponseWriter, r *http.Request) {
	limit, ok := intQuery(w, r, "limit", 100)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	models := h.modelViews(r)
	total := len(models)
	offset, limit = normalizePage(offset, limit)
	pagedModels := []modelView{}
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		pagedModels = models[offset:end]
	}
	query := r.URL.Query()
	writeJSON(w, http.StatusOK, map[string]any{
		"models": pagedModels,
		"total":  total,
		"offset": offset,
		"limit":  limit,
		"filters": map[string]string{
			"model":          query.Get("model"),
			"provider_id":    query.Get("provider_id"),
			"provider_class": query.Get("provider_class"),
			"enabled":        query.Get("enabled"),
			"available":      query.Get("available"),
			"all":            query.Get("all"),
		},
	})
}

func (h *Handler) modelViews(r *http.Request) []modelView {
	snapshot := h.snapshot()
	onlyAvailable := r.URL.Query().Get("all") != "true"
	query := r.URL.Query()
	modelFilter := query.Get("model")
	providerFilter := query.Get("provider_id")
	classFilter := query.Get("provider_class")
	enabledFilter := query.Get("enabled")
	availableFilter := query.Get("available")
	healthByID := map[string]providerhealth.View{}
	if snapshot.Health != nil {
		for _, view := range snapshot.Health.Views(snapshot.Config.Providers) {
			healthByID[view.ID] = view
		}
	}

	models := []modelView{}
	for _, provider := range snapshot.Config.Providers {
		view, ok := healthByID[provider.ID]
		enabled := provider.IsEnabled()
		status := ""
		if ok {
			enabled = view.Enabled
			status = view.RuntimeStatus
		}
		available := enabled && (status == "" || status == providerhealth.StatusHealthy || status == providerhealth.StatusDegraded)
		if onlyAvailable && !available {
			continue
		}
		for _, model := range provider.Models {
			if modelFilter != "" && model != modelFilter {
				continue
			}
			if providerFilter != "" && provider.ID != providerFilter {
				continue
			}
			if classFilter != "" && provider.Class != classFilter {
				continue
			}
			if enabledFilter != "" && strconv.FormatBool(enabled) != enabledFilter {
				continue
			}
			if availableFilter != "" && strconv.FormatBool(available) != availableFilter {
				continue
			}
			models = append(models, modelView{
				Model:         model,
				ProviderID:    provider.ID,
				ProviderClass: provider.Class,
				Enabled:       enabled,
				RuntimeStatus: status,
				Available:     available,
			})
		}
	}
	return models
}

func (h *Handler) modelsExport(w http.ResponseWriter, r *http.Request) {
	writeJSONL(w, "models.jsonl", h.modelViews(r))
}

func (h *Handler) runtimeHealth(w http.ResponseWriter, _ *http.Request) {
	if h.runtime == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "degraded",
			"daemon": map[string]string{
				"status": "standalone",
			},
			"provider_monitor": map[string]string{
				"status": "unavailable",
			},
		})
		return
	}
	writeJSON(w, http.StatusOK, h.runtime.Health())
}

type toolView struct {
	ID              string         `json:"id"`
	Name            string         `json:"name,omitempty"`
	Adapter         string         `json:"adapter"`
	Origin          string         `json:"origin,omitempty"`
	ServerID        string         `json:"server_id,omitempty"`
	Description     string         `json:"description,omitempty"`
	ReadOnly        bool           `json:"read_only"`
	RiskLevel       string         `json:"risk_level,omitempty"`
	Scopes          []string       `json:"scopes,omitempty"`
	InputSchema     map[string]any `json:"input_schema,omitempty"`
	OutputSchema    map[string]any `json:"output_schema,omitempty"`
	SandboxRequired bool           `json:"sandbox_required"`
	Enabled         bool           `json:"enabled"`
}

type mcpServerView struct {
	ID           string     `json:"id"`
	Name         string     `json:"name,omitempty"`
	Enabled      bool       `json:"enabled"`
	ToolCount    int        `json:"tool_count"`
	EnabledTools int        `json:"enabled_tools"`
	Tools        []toolView `json:"tools,omitempty"`
}

type appView struct {
	ID        string   `json:"id"`
	Name      string   `json:"name,omitempty"`
	TokenHint string   `json:"token_hint"`
	Grants    []string `json:"grants"`
	Quota     appQuota `json:"quota,omitempty"`
}

type appQuota struct {
	RequestsPerMinute int  `json:"requests_per_minute"`
	Enabled           bool `json:"enabled"`
}

type grantView struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Apps        []string `json:"apps,omitempty"`
	Tools       []string `json:"tools,omitempty"`
	Servers     []string `json:"servers,omitempty"`
}

type policyView struct {
	ID               string                 `json:"id"`
	EvaluationOrder  int                    `json:"evaluation_order"`
	Priority         int                    `json:"priority"`
	Effect           string                 `json:"effect"`
	EffectSemantics  policy.EffectSemantics `json:"effect_semantics"`
	Reason           string                 `json:"reason"`
	ConditionSummary string                 `json:"condition_summary"`
	AppIDs           []string               `json:"app_ids,omitempty"`
	RequestTypes     []string               `json:"request_types,omitempty"`
	Models           []string               `json:"models,omitempty"`
	ProviderClasses  []string               `json:"provider_classes,omitempty"`
	DataLabels       []string               `json:"data_labels,omitempty"`
}

func (h *Handler) appsList(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	limit, ok := intQuery(w, r, "limit", 100)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	views := h.appViews(r)
	total := len(views)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	pagedApps := []appView{}
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		pagedApps = views[offset:end]
	}
	query := r.URL.Query()
	writeJSON(w, http.StatusOK, map[string]any{
		"apps":   pagedApps,
		"total":  total,
		"offset": offset,
		"limit":  limit,
		"filters": map[string]string{
			"app_id":        query.Get("app_id"),
			"grant":         query.Get("grant"),
			"quota_enabled": query.Get("quota_enabled"),
		},
	})
}

func (h *Handler) appViews(r *http.Request) []appView {
	snapshot := h.snapshot()
	query := r.URL.Query()
	appFilter := query.Get("app_id")
	grantFilter := query.Get("grant")
	quotaEnabledFilter := query.Get("quota_enabled")
	quotas := appQuotaMap(snapshot.Config.Quotas)
	views := make([]appView, 0, len(snapshot.Config.Apps))
	for _, app := range snapshot.Config.Apps {
		if appFilter != "" && app.ID != appFilter {
			continue
		}
		if grantFilter != "" && !hasGrant(app.Grants, grantFilter) {
			continue
		}
		quotaView := quotas[app.ID]
		if quotaEnabledFilter == "true" && !quotaView.Enabled {
			continue
		}
		if quotaEnabledFilter == "false" && quotaView.Enabled {
			continue
		}
		views = append(views, appView{
			ID:        app.ID,
			Name:      app.Name,
			TokenHint: tokenHint(app.Token),
			Grants:    append([]string(nil), app.Grants...),
			Quota:     quotaView,
		})
	}
	return views
}

func appQuotaMap(quotas config.Quotas) map[string]appQuota {
	views := make(map[string]appQuota, len(quotas.Apps))
	for _, quota := range quotas.Apps {
		views[quota.AppID] = appQuota{
			RequestsPerMinute: quota.RequestsPerMinute,
			Enabled:           quota.RequestsPerMinute > 0,
		}
	}
	return views
}

func (h *Handler) appsExport(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	writeJSONL(w, "apps.jsonl", h.appViews(r))
}

func (h *Handler) grantsList(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	limit, ok := intQuery(w, r, "limit", 100)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	views := h.grantViews(r)
	total := len(views)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	pagedGrants := []grantView{}
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		pagedGrants = views[offset:end]
	}
	query := r.URL.Query()
	writeJSON(w, http.StatusOK, map[string]any{
		"grants": pagedGrants,
		"total":  total,
		"offset": offset,
		"limit":  limit,
		"filters": map[string]string{
			"grant":   query.Get("grant"),
			"type":    query.Get("type"),
			"app_id":  query.Get("app_id"),
			"tool_id": query.Get("tool_id"),
		},
	})
}

func (h *Handler) grantViews(r *http.Request) []grantView {
	snapshot := h.snapshot()
	grants := map[string]*grantView{}
	for _, grant := range []grantView{
		{ID: "chat", Type: "core", Description: "Allow OpenAI-compatible chat completions."},
		{ID: "embedding", Type: "core", Description: "Reserved grant for embedding requests."},
		{ID: "tool", Type: "tool_broad", Description: "Allow invoking any enabled read-only tool."},
		{ID: "admin", Type: "admin", Description: "Allow management APIs such as audit, config reload, providers, apps and grants."},
	} {
		item := grant
		grants[item.ID] = &item
	}
	for _, tool := range snapshot.Config.Tools {
		for _, scope := range tool.Scopes {
			view := ensureGrantView(grants, "tool:"+scope, "tool_scope", "Allow invoking tools that require scope "+scope+".")
			view.Tools = appendUnique(view.Tools, tool.ID)
		}
	}
	if snapshot.Config.MCPRuntime.Enabled {
		for _, server := range snapshot.Config.MCPRuntime.Servers {
			for _, tool := range server.Tools {
				for _, scope := range tool.Scopes {
					view := ensureGrantView(grants, "tool:"+scope, "tool_scope", "Allow invoking tools that require scope "+scope+".")
					view.Tools = appendUnique(view.Tools, tool.ID)
					view.Servers = appendUnique(view.Servers, server.ID)
				}
			}
		}
	}
	for _, app := range snapshot.Config.Apps {
		for _, grant := range app.Grants {
			view := ensureGrantView(grants, grant, grantType(grant), grantDescription(grant))
			view.Apps = appendUnique(view.Apps, app.ID)
		}
	}

	query := r.URL.Query()
	grantFilter := query.Get("grant")
	typeFilter := query.Get("type")
	appFilter := query.Get("app_id")
	toolFilter := query.Get("tool_id")
	views := make([]grantView, 0, len(grants))
	for _, view := range grants {
		if grantFilter != "" && view.ID != grantFilter {
			continue
		}
		if typeFilter != "" && view.Type != typeFilter {
			continue
		}
		if appFilter != "" && !containsString(view.Apps, appFilter) {
			continue
		}
		if toolFilter != "" && !containsString(view.Tools, toolFilter) {
			continue
		}
		views = append(views, *view)
	}
	sortGrantViews(views)
	return views
}

func (h *Handler) grantsExport(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	writeJSONL(w, "grants.jsonl", h.grantViews(r))
}

func (h *Handler) policiesList(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	limit, ok := intQuery(w, r, "limit", 100)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	views := h.policyViews(r)
	total := len(views)
	offset, limit = normalizePage(offset, limit)
	pagedPolicies := []policyView{}
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		pagedPolicies = views[offset:end]
	}
	query := r.URL.Query()
	writeJSON(w, http.StatusOK, map[string]any{
		"policies": pagedPolicies,
		"total":    total,
		"offset":   offset,
		"limit":    limit,
		"filters": map[string]string{
			"policy_id":      query.Get("policy_id"),
			"effect":         query.Get("effect"),
			"app_id":         query.Get("app_id"),
			"request_type":   query.Get("request_type"),
			"model":          query.Get("model"),
			"provider_class": query.Get("provider_class"),
			"data_label":     query.Get("data_label"),
		},
	})
}

func (h *Handler) policyViews(r *http.Request) []policyView {
	snapshot := h.snapshot()
	query := r.URL.Query()
	policyFilter := query.Get("policy_id")
	effectFilter := query.Get("effect")
	appFilter := query.Get("app_id")
	requestTypeFilter := query.Get("request_type")
	modelFilter := query.Get("model")
	providerClassFilter := query.Get("provider_class")
	dataLabelFilter := query.Get("data_label")
	rules := policy.OrderedRules(snapshot.Config.Policies)
	views := make([]policyView, 0, len(rules))
	for index, rule := range rules {
		dataLabels := policy.EffectiveDataLabels(rule)
		if policyFilter != "" && rule.ID != policyFilter {
			continue
		}
		if effectFilter != "" && rule.Effect != effectFilter {
			continue
		}
		if appFilter != "" && !containsString(rule.AppIDs, appFilter) {
			continue
		}
		if requestTypeFilter != "" && !containsString(rule.RequestTypes, requestTypeFilter) {
			continue
		}
		if modelFilter != "" && !containsString(rule.Models, modelFilter) {
			continue
		}
		if providerClassFilter != "" && !containsString(rule.ProviderClasses, providerClassFilter) {
			continue
		}
		if dataLabelFilter != "" && !containsString(dataLabels, dataLabelFilter) {
			continue
		}
		views = append(views, policyView{
			ID:               rule.ID,
			EvaluationOrder:  index + 1,
			Priority:         rule.Priority,
			Effect:           rule.Effect,
			EffectSemantics:  policy.EffectSemanticsFor(rule.Effect),
			Reason:           rule.Reason,
			ConditionSummary: policy.ConditionSummary(rule),
			AppIDs:           append([]string(nil), rule.AppIDs...),
			RequestTypes:     append([]string(nil), rule.RequestTypes...),
			Models:           append([]string(nil), rule.Models...),
			ProviderClasses:  append([]string(nil), rule.ProviderClasses...),
			DataLabels:       dataLabels,
		})
	}
	return views
}

func (h *Handler) policiesExport(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	writeJSONL(w, "policies.jsonl", h.policyViews(r))
}

func (h *Handler) policyByID(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	policyID := strings.TrimPrefix(r.URL.Path, "/gateway/v1/policies/")
	if policyID == "" || strings.Contains(policyID, "/") {
		writeError(w, http.StatusNotFound, "", "not_found", "policy not found")
		return
	}
	query := r.URL.Query()
	query.Set("policy_id", policyID)
	r.URL.RawQuery = query.Encode()
	views := h.policyViews(r)
	if len(views) == 0 {
		writeError(w, http.StatusNotFound, "", "not_found", "policy not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"policy": views[0]})
}

func (h *Handler) toolsList(w http.ResponseWriter, r *http.Request) {
	limit, ok := intQuery(w, r, "limit", 100)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	views := h.toolViews(r)
	total := len(views)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset >= total {
		writeJSON(w, http.StatusOK, map[string]any{"tools": []toolView{}, "total": total, "offset": offset, "limit": limit})
		return
	}
	end := offset + limit
	if end > total {
		end = total
	}
	writeJSON(w, http.StatusOK, map[string]any{"tools": views[offset:end], "total": total, "offset": offset, "limit": limit})
}

func (h *Handler) toolsExport(w http.ResponseWriter, r *http.Request) {
	writeJSONL(w, "tools.jsonl", h.toolViews(r))
}

func (h *Handler) toolViews(r *http.Request) []toolView {
	snapshot := h.snapshot()
	query := r.URL.Query()
	toolFilter := query.Get("tool_id")
	originFilter := query.Get("origin")
	serverFilter := query.Get("server_id")
	scopeFilter := query.Get("scope")
	enabledFilter := query.Get("enabled")
	views := make([]toolView, 0, len(snapshot.Config.Tools))
	for _, tool := range snapshot.Config.Tools {
		manifest := tools.ManifestFromConfig(tool)
		if snapshot.Tools != nil {
			if registered, ok := snapshot.Tools.Get(tool.ID); ok {
				manifest = registered.Manifest()
			}
		}
		view := toolView{
			ID:              manifest.ID,
			Name:            manifest.Name,
			Adapter:         manifest.Adapter,
			Origin:          "builtin",
			Description:     manifest.Description,
			ReadOnly:        manifest.ReadOnly,
			RiskLevel:       manifest.RiskLevel,
			Scopes:          manifest.Scopes,
			InputSchema:     manifest.InputSchema,
			OutputSchema:    manifest.OutputSchema,
			SandboxRequired: manifest.SandboxRequired,
			Enabled:         tool.IsEnabled(),
		}
		if includeToolView(view, toolFilter, originFilter, serverFilter, scopeFilter, enabledFilter) {
			views = append(views, view)
		}
	}
	if snapshot.Config.MCPRuntime.Enabled {
		for _, server := range snapshot.Config.MCPRuntime.Servers {
			for _, tool := range server.Tools {
				view := toolView{
					ID:              tool.ID,
					Name:            tool.Name,
					Adapter:         "mcp-placeholder",
					Origin:          "mcp",
					ServerID:        server.ID,
					Description:     tool.Description,
					ReadOnly:        tool.ReadOnly,
					RiskLevel:       tool.RiskLevel,
					Scopes:          append([]string(nil), tool.Scopes...),
					InputSchema:     tool.InputSchema,
					OutputSchema:    tool.OutputSchema,
					SandboxRequired: tool.SandboxRequired,
					Enabled:         server.IsEnabled() && tool.IsEnabled(),
				}
				if includeToolView(view, toolFilter, originFilter, serverFilter, scopeFilter, enabledFilter) {
					views = append(views, view)
				}
			}
		}
	}
	return views
}

func includeToolView(view toolView, toolFilter, originFilter, serverFilter, scopeFilter, enabledFilter string) bool {
	if toolFilter != "" && view.ID != toolFilter {
		return false
	}
	if originFilter != "" && view.Origin != originFilter {
		return false
	}
	if serverFilter != "" && view.ServerID != serverFilter {
		return false
	}
	if scopeFilter != "" && !hasScope(view.Scopes, scopeFilter) {
		return false
	}
	if enabledFilter != "" {
		if enabledFilter == "true" && !view.Enabled {
			return false
		}
		if enabledFilter == "false" && view.Enabled {
			return false
		}
	}
	return true
}

func (h *Handler) mcpServers(w http.ResponseWriter, r *http.Request) {
	limit, ok := intQuery(w, r, "limit", 100)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	runtime, servers := h.mcpServerViews(r)
	total := len(servers)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	pagedServers := []mcpServerView{}
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		pagedServers = servers[offset:end]
	}
	mode := runtime.Mode
	if mode == "" {
		mode = "manifest_only"
	}
	query := r.URL.Query()
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": runtime.Enabled,
		"mode":    mode,
		"servers": pagedServers,
		"total":   total,
		"offset":  offset,
		"limit":   limit,
		"filters": map[string]string{
			"server_id": query.Get("server_id"),
			"scope":     query.Get("scope"),
			"enabled":   query.Get("enabled"),
		},
	})
}

func (h *Handler) mcpServersExport(w http.ResponseWriter, r *http.Request) {
	_, servers := h.mcpServerViews(r)
	writeJSONL(w, "mcp-servers.jsonl", servers)
}

func (h *Handler) mcpServerViews(r *http.Request) (config.MCPRuntime, []mcpServerView) {
	snapshot := h.snapshot()
	runtime := snapshot.Config.MCPRuntime
	query := r.URL.Query()
	serverFilter := query.Get("server_id")
	scopeFilter := query.Get("scope")
	enabledFilter := query.Get("enabled")
	servers := make([]mcpServerView, 0, len(runtime.Servers))
	for _, server := range runtime.Servers {
		if serverFilter != "" && server.ID != serverFilter {
			continue
		}
		view := mcpServerView{
			ID:      server.ID,
			Name:    server.Name,
			Enabled: runtime.Enabled && server.IsEnabled(),
			Tools:   make([]toolView, 0, len(server.Tools)),
		}
		for _, tool := range server.Tools {
			enabled := view.Enabled && tool.IsEnabled()
			if enabledFilter != "" {
				if enabledFilter == "true" && !enabled {
					continue
				}
				if enabledFilter == "false" && enabled {
					continue
				}
			}
			if scopeFilter != "" && !hasScope(tool.Scopes, scopeFilter) {
				continue
			}
			view.ToolCount++
			if enabled {
				view.EnabledTools++
			}
			view.Tools = append(view.Tools, toolView{
				ID:              tool.ID,
				Name:            tool.Name,
				Adapter:         "mcp-placeholder",
				Origin:          "mcp",
				ServerID:        server.ID,
				Description:     tool.Description,
				ReadOnly:        tool.ReadOnly,
				RiskLevel:       tool.RiskLevel,
				Scopes:          append([]string(nil), tool.Scopes...),
				InputSchema:     tool.InputSchema,
				OutputSchema:    tool.OutputSchema,
				SandboxRequired: tool.SandboxRequired,
				Enabled:         enabled,
			})
		}
		if scopeFilter != "" || enabledFilter != "" {
			if len(view.Tools) == 0 {
				continue
			}
		}
		servers = append(servers, view)
	}
	return runtime, servers
}

type toolInvokeRequest struct {
	Arguments map[string]any `json:"arguments,omitempty"`
}

type toolConfigRef struct {
	ID              string
	Adapter         string
	Origin          string
	ServerID        string
	ReadOnly        bool
	Scopes          []string
	SandboxRequired bool
	Enabled         bool
}

func (h *Handler) toolInvoke(w http.ResponseWriter, r *http.Request) {
	traceID := trace.NewID()
	startedAt := time.Now().UTC()
	record := trace.Record{
		TraceID:     traceID,
		RequestType: "tool",
		ToolID:      strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/gateway/v1/tools/"), "/invoke"),
		Status:      "started",
		StartedAt:   startedAt,
		Events:      []trace.Event{{Type: "tool_invocation_started", Message: "tool invocation accepted by access layer", At: startedAt}},
	}
	if h.runtime == nil {
		h.finishToolTrace(record, "", "failed", "runtime manager is not configured", startedAt)
		writeError(w, http.StatusServiceUnavailable, traceID, "runtime_unavailable", "runtime manager is not configured")
		return
	}
	toolID := record.ToolID
	if toolID == "" || !strings.HasSuffix(r.URL.Path, "/invoke") {
		record.ToolID = ""
		h.finishToolTrace(record, "", "failed", "tool action not found", startedAt)
		writeError(w, http.StatusNotFound, traceID, "not_found", "tool action not found")
		return
	}
	snapshot := h.snapshot()
	app, ok := snapshot.Config.AppByToken(bearerToken(r.Header.Get("Authorization")))
	if ok {
		record.AppID = app.ID
	}
	if !ok {
		h.finishToolTrace(record, "", "failed", "tool grant is required", startedAt)
		h.saveAudit(audit.Event{TraceID: traceID, Action: "tool.invoke", Target: toolID, Result: audit.ResultDenied, Error: "tool grant is required"})
		writeError(w, http.StatusUnauthorized, traceID, "unauthorized", "tool grant is required")
		return
	}
	var req toolInvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.finishToolTrace(record, app.ID, "failed", err.Error(), startedAt)
		writeError(w, http.StatusBadRequest, traceID, "invalid_request", err.Error())
		return
	}
	toolCfg, ok := findTool(snapshot.Config, toolID)
	if !ok || !toolCfg.Enabled {
		h.finishToolTrace(record, app.ID, "failed", "tool not found or disabled", startedAt)
		h.saveAudit(audit.Event{TraceID: traceID, AppID: app.ID, Action: "tool.invoke", Target: toolID, Result: audit.ResultFailed, Error: "tool not found or disabled"})
		writeError(w, http.StatusNotFound, traceID, "not_found", "tool not found")
		return
	}
	if !hasToolScope(app.Grants, toolCfg.Scopes) {
		h.finishToolTrace(record, app.ID, "failed", "tool scope is required", startedAt)
		h.saveAudit(audit.Event{
			TraceID:  traceID,
			AppID:    app.ID,
			Action:   "tool.invoke",
			Target:   toolID,
			Result:   audit.ResultDenied,
			Error:    "tool scope is required",
			Metadata: toolAuditMetadata(toolCfg, "", missingToolGrants(toolCfg.Scopes)),
		})
		writeError(w, http.StatusForbidden, traceID, "tool_scope_denied", "tool scope is required")
		return
	}
	if !toolCfg.ReadOnly {
		h.finishToolTrace(record, app.ID, "failed", "only read-only tools are allowed", startedAt)
		h.saveAudit(audit.Event{TraceID: traceID, AppID: app.ID, Action: "tool.invoke", Target: toolID, Result: audit.ResultDenied, Error: "only read-only tools are allowed", Metadata: toolAuditMetadata(toolCfg, matchedToolGrant(app.Grants, toolCfg.Scopes), nil)})
		writeError(w, http.StatusForbidden, traceID, "tool_denied", "only read-only tools are allowed")
		return
	}
	if snapshot.Tools == nil {
		h.finishToolTrace(record, app.ID, "failed", "tool registry is not configured", startedAt)
		writeError(w, http.StatusServiceUnavailable, traceID, "tool_unavailable", "tool registry is not configured")
		return
	}
	if toolCfg.Origin == "mcp" {
		message := "MCP tool execution is unavailable in manifest-only mode"
		h.finishToolTrace(record, app.ID, "failed", message, startedAt)
		h.saveAudit(audit.Event{
			TraceID:  traceID,
			AppID:    app.ID,
			Action:   "tool.invoke",
			Target:   toolID,
			Result:   audit.ResultFailed,
			Error:    message,
			Metadata: toolAuditMetadata(toolCfg, matchedToolGrant(app.Grants, toolCfg.Scopes), nil),
		})
		writeError(w, http.StatusServiceUnavailable, traceID, tools.ErrCodeUnavailable, message)
		return
	}
	tool, ok := snapshot.Tools.Get(toolID)
	if !ok {
		h.finishToolTrace(record, app.ID, "failed", "tool adapter not registered", startedAt)
		h.saveAudit(audit.Event{TraceID: traceID, AppID: app.ID, Action: "tool.invoke", Target: toolID, Result: audit.ResultFailed, Error: "tool adapter not registered"})
		writeError(w, http.StatusNotFound, traceID, "not_found", "tool not found")
		return
	}
	result, err := tool.Invoke(r.Context(), tools.Input{AppID: app.ID, ToolID: toolID, Arguments: req.Arguments})
	finishedAt := time.Now().UTC()
	durationMS := finishedAt.Sub(startedAt).Milliseconds()
	if err != nil {
		record.FinishedAt = finishedAt
		record.DurationMS = durationMS
		record.Status = "failed"
		record.Error = err.Error()
		record.Events = append(record.Events, trace.Event{Type: "tool_invocation_failed", Message: err.Error(), At: finishedAt})
		_ = h.traces.Save(record)
		h.saveAudit(audit.Event{
			TraceID:    traceID,
			AppID:      app.ID,
			Action:     "tool.invoke",
			Target:     toolID,
			Result:     audit.ResultFailed,
			Error:      err.Error(),
			DurationMS: durationMS,
			Metadata:   toolAuditMetadata(toolCfg, matchedToolGrant(app.Grants, toolCfg.Scopes), nil),
		})
		code, status := toolErrorResponse(err)
		writeError(w, status, traceID, code, err.Error())
		return
	}
	record.FinishedAt = finishedAt
	record.DurationMS = durationMS
	record.Status = "completed"
	record.Events = append(record.Events, trace.Event{Type: "tool_invocation_completed", Message: "tool returned response", At: finishedAt})
	_ = h.traces.Save(record)
	result.TraceID = traceID
	result.AppID = app.ID
	result.DurationMS = durationMS
	h.saveAudit(audit.Event{
		TraceID:    traceID,
		AppID:      app.ID,
		Action:     "tool.invoke",
		Target:     toolID,
		Result:     audit.ResultSuccess,
		DurationMS: durationMS,
		Metadata:   toolAuditMetadata(toolCfg, matchedToolGrant(app.Grants, toolCfg.Scopes), nil),
	})
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) console(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(consoleHTML))
}

type policyDryRunRequest struct {
	AppID         string   `json:"app_id"`
	RequestType   string   `json:"request_type"`
	DataLabels    []string `json:"data_labels"`
	Model         string   `json:"model"`
	ProviderClass string   `json:"provider_class"`
}

type routingExplainRequest struct {
	AppID       string   `json:"app_id"`
	Model       string   `json:"model"`
	DataLabels  []string `json:"data_labels"`
	RequestType string   `json:"request_type"`
}

type accessDryRunRequest struct {
	AppID  string `json:"app_id"`
	Token  string `json:"token"`
	Action string `json:"action"`
	ToolID string `json:"tool_id"`
}

func (h *Handler) policyDryRun(w http.ResponseWriter, r *http.Request) {
	var req policyDryRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "", "invalid_request", err.Error())
		return
	}
	if req.RequestType == "" {
		req.RequestType = "chat"
	}
	snapshot := h.snapshot()
	engine := policy.NewEngine(snapshot.Config.PolicyVersion, snapshot.Config.Policies)
	decision := engine.Evaluate(policy.Input{
		AppID:         req.AppID,
		RequestType:   req.RequestType,
		DataLabels:    req.DataLabels,
		Model:         req.Model,
		ProviderClass: req.ProviderClass,
	})
	h.saveAudit(audit.Event{
		AppID:    req.AppID,
		Action:   "policy.dry_run",
		Target:   req.Model,
		Result:   audit.ResultSuccess,
		Metadata: policyDryRunAuditMetadata(req, decision),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"decision":      decision,
		"explain_chain": policyExplainChain(req, decision),
		"input":         req,
	})
}

func policyDryRunAuditMetadata(req policyDryRunRequest, decision policy.Decision) map[string]any {
	return map[string]any{
		"request_type":     req.RequestType,
		"data_labels":      strings.Join(req.DataLabels, ","),
		"provider_class":   req.ProviderClass,
		"policy_rule_id":   decision.RuleID,
		"rule_priority":    strconv.Itoa(decision.RulePriority),
		"condition":        decision.ConditionSummary,
		"rule_evaluations": decision.RuleEvaluations,
		"allow_cloud":      strconv.FormatBool(decision.AllowCloud),
		"force_local":      strconv.FormatBool(decision.ForceLocal),
		"explain_chain":    policyExplainChain(req, decision),
	}
}

func (h *Handler) accessDryRun(w http.ResponseWriter, r *http.Request) {
	var req accessDryRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "", "invalid_request", err.Error())
		return
	}
	if req.Action == "" {
		req.Action = "chat"
	}
	snapshot := h.snapshot()
	app, ok := resolveAccessDryRunApp(snapshot.Config, req)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"allowed": false,
			"reason":  "app not found",
			"input":   req,
		})
		return
	}

	allowed := false
	reason := ""
	matchedGrant := ""
	missingGrants := []string{}
	var toolRef *toolConfigRef

	switch req.Action {
	case "chat":
		allowed = hasGrant(app.Grants, "chat")
		matchedGrant = grantIfAllowed(allowed, "chat")
		missingGrants = grantsIfDenied(allowed, "chat")
	case "admin":
		allowed = hasGrant(app.Grants, "admin")
		matchedGrant = grantIfAllowed(allowed, "admin")
		missingGrants = grantsIfDenied(allowed, "admin")
	case "tool.invoke":
		if req.ToolID == "" {
			writeError(w, http.StatusBadRequest, "", "invalid_request", "tool_id is required for tool.invoke")
			return
		}
		toolCfg, found := findTool(snapshot.Config, req.ToolID)
		if !found {
			reason = "tool not found"
			missingGrants = []string{"tool"}
			break
		}
		toolRef = &toolCfg
		if !toolCfg.Enabled {
			reason = "tool is disabled"
			missingGrants = []string{"enabled tool"}
			break
		}
		allowed = hasToolScope(app.Grants, toolCfg.Scopes)
		if allowed {
			matchedGrant = matchedToolGrant(app.Grants, toolCfg.Scopes)
		} else {
			missingGrants = missingToolGrants(toolCfg.Scopes)
		}
	default:
		writeError(w, http.StatusBadRequest, "", "invalid_request", "action must be chat, admin, or tool.invoke")
		return
	}
	if reason == "" {
		if allowed {
			reason = "required grant is present"
		} else {
			reason = "required grant is missing"
		}
	}
	explainChain := accessExplainChain(req.Action, allowed, reason, matchedGrant, missingGrants, toolRef)

	response := map[string]any{
		"allowed":        allowed,
		"reason":         reason,
		"app_id":         app.ID,
		"action":         req.Action,
		"matched_grant":  matchedGrant,
		"missing_grants": missingGrants,
		"explain_chain":  explainChain,
		"grants":         app.Grants,
		"input":          req,
	}
	if toolRef != nil {
		response["tool"] = map[string]any{
			"id":               toolRef.ID,
			"origin":           toolRef.Origin,
			"server_id":        toolRef.ServerID,
			"adapter":          toolRef.Adapter,
			"read_only":        toolRef.ReadOnly,
			"scopes":           toolRef.Scopes,
			"sandbox_required": toolRef.SandboxRequired,
			"enabled":          toolRef.Enabled,
		}
	}
	h.saveAudit(audit.Event{
		AppID:    app.ID,
		Action:   "access.dry_run",
		Target:   req.Action,
		Result:   audit.ResultSuccess,
		Metadata: accessDryRunAuditMetadata(req.Action, allowed, reason, matchedGrant, missingGrants, toolRef),
	})
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) routingExplain(w http.ResponseWriter, r *http.Request) {
	var req routingExplainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "", "invalid_request", err.Error())
		return
	}
	if req.RequestType == "" {
		req.RequestType = "chat"
	}
	snapshot := h.snapshot()
	engine := policy.NewEngine(snapshot.Config.PolicyVersion, snapshot.Config.Policies)
	decision := engine.Evaluate(policy.Input{
		AppID:       req.AppID,
		RequestType: req.RequestType,
		DataLabels:  req.DataLabels,
		Model:       req.Model,
	})
	explanation := router.NewWithHealth(snapshot.Config.Providers, snapshot.Health).Explain(router.Input{
		Model:      req.Model,
		AllowCloud: decision.AllowCloud,
	})
	explainChain := routingExplainChain(req, decision, len(explanation.Candidates), len(explanation.Skipped))
	h.saveAudit(audit.Event{
		AppID:  req.AppID,
		Action: "routing.explain",
		Target: req.Model,
		Result: audit.ResultSuccess,
		Metadata: map[string]any{
			"request_type":     req.RequestType,
			"data_labels":      req.DataLabels,
			"policy_rule_id":   decision.RuleID,
			"rule_priority":    decision.RulePriority,
			"condition":        decision.ConditionSummary,
			"rule_evaluations": decision.RuleEvaluations,
			"allow_cloud":      decision.AllowCloud,
			"candidate_count":  len(explanation.Candidates),
			"skipped_count":    len(explanation.Skipped),
			"explain_chain":    explainChain,
		},
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"input":         req,
		"policy":        decision,
		"explain_chain": explainChain,
		"candidates":    explanation.Candidates,
		"skipped":       explanation.Skipped,
	})
}

func (h *Handler) configReload(w http.ResponseWriter, r *http.Request) {
	if h.runtime == nil {
		writeError(w, http.StatusServiceUnavailable, "", "reload_unavailable", "runtime manager is not configured")
		return
	}
	snapshot := h.snapshot()
	app, ok := snapshot.Config.AppByToken(bearerToken(r.Header.Get("Authorization")))
	if !ok || !hasGrant(app.Grants, "admin") {
		appID := ""
		if ok {
			appID = app.ID
		}
		h.saveAudit(audit.Event{
			AppID:  appID,
			Action: "config.reload",
			Target: "runtime",
			Result: audit.ResultDenied,
			Error:  "admin grant is required",
		})
		writeError(w, http.StatusUnauthorized, "", "unauthorized", "admin grant is required")
		return
	}
	if err := h.runtime.Reload(); err != nil {
		h.saveAudit(audit.Event{
			AppID:  app.ID,
			Action: "config.reload",
			Target: "runtime",
			Result: audit.ResultFailed,
			Error:  err.Error(),
		})
		writeError(w, http.StatusBadRequest, "", "reload_failed", err.Error())
		return
	}
	snapshot = h.runtime.Snapshot()
	h.saveAudit(audit.Event{
		AppID:  app.ID,
		Action: "config.reload",
		Target: "runtime",
		Result: audit.ResultSuccess,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "reloaded",
		"providers":      len(snapshot.Config.Providers),
		"policy_version": snapshot.Config.PolicyVersion,
	})
}

func (h *Handler) providerAction(w http.ResponseWriter, r *http.Request) {
	if h.runtime == nil {
		writeError(w, http.StatusServiceUnavailable, "", "runtime_unavailable", "runtime manager is not configured")
		return
	}
	providerID, action, ok := parseProviderAction(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "", "not_found", "provider action not found")
		return
	}
	app, ok := h.requireAdmin(w, r)
	if !ok {
		h.saveAudit(audit.Event{
			Action: "provider." + action,
			Target: providerID,
			Result: audit.ResultDenied,
			Error:  "admin grant is required",
		})
		return
	}
	switch action {
	case "enabled":
		var req providerEnabledRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "", "invalid_request", err.Error())
			return
		}
		if err := h.runtime.SetProviderEnabled(providerID, req.Enabled); err != nil {
			h.saveAudit(audit.Event{AppID: app.ID, Action: "provider.enabled", Target: providerID, Result: audit.ResultFailed, Error: err.Error()})
			writeError(w, http.StatusBadRequest, "", "provider_update_failed", err.Error())
			return
		}
		h.saveAudit(audit.Event{AppID: app.ID, Action: "provider.enabled", Target: providerID, Result: audit.ResultSuccess})
		writeJSON(w, http.StatusOK, map[string]any{"status": "updated", "provider_id": providerID, "enabled": req.Enabled})
	case "probe":
		view, err := h.runtime.ProbeProvider(r.Context(), providerID)
		if err != nil {
			h.saveAudit(audit.Event{AppID: app.ID, Action: "provider.probe", Target: providerID, Result: audit.ResultFailed, Error: err.Error()})
			writeError(w, http.StatusBadGateway, "", "provider_probe_failed", err.Error())
			return
		}
		h.saveAudit(audit.Event{AppID: app.ID, Action: "provider.probe", Target: providerID, Result: audit.ResultSuccess})
		writeJSON(w, http.StatusOK, map[string]any{"status": "probed", "provider": view})
	default:
		writeError(w, http.StatusNotFound, "", "not_found", "provider action not found")
	}
}

func (h *Handler) auditList(w http.ResponseWriter, r *http.Request) {
	if h.audit == nil {
		writeError(w, http.StatusServiceUnavailable, "", "audit_unavailable", "audit store is not configured")
		return
	}
	snapshot := h.snapshot()
	app, ok := snapshot.Config.AppByToken(bearerToken(r.Header.Get("Authorization")))
	if !ok || !hasGrant(app.Grants, "admin") {
		writeError(w, http.StatusUnauthorized, "", "unauthorized", "admin grant is required")
		return
	}
	limit, ok := intQuery(w, r, "limit", 100)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	page := h.audit.Page(audit.ListQuery{
		Offset:        offset,
		Limit:         limit,
		Action:        r.URL.Query().Get("action"),
		Result:        r.URL.Query().Get("result"),
		AppID:         r.URL.Query().Get("app_id"),
		TraceID:       r.URL.Query().Get("trace_id"),
		Target:        r.URL.Query().Get("target"),
		MetadataKey:   r.URL.Query().Get("metadata_key"),
		MetadataValue: r.URL.Query().Get("metadata_value"),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"events": page.Items,
		"total":  page.Total,
		"offset": page.Offset,
		"limit":  page.Limit,
	})
}

func (h *Handler) auditExport(w http.ResponseWriter, r *http.Request) {
	if h.audit == nil {
		writeError(w, http.StatusServiceUnavailable, "", "audit_unavailable", "audit store is not configured")
		return
	}
	snapshot := h.snapshot()
	app, ok := snapshot.Config.AppByToken(bearerToken(r.Header.Get("Authorization")))
	if !ok || !hasGrant(app.Grants, "admin") {
		writeError(w, http.StatusUnauthorized, "", "unauthorized", "admin grant is required")
		return
	}
	limit, ok := intQuery(w, r, "limit", 500)
	if !ok {
		return
	}
	offset, ok := intQuery(w, r, "offset", 0)
	if !ok {
		return
	}
	page := h.audit.Page(audit.ListQuery{
		Offset:        offset,
		Limit:         limit,
		Action:        r.URL.Query().Get("action"),
		Result:        r.URL.Query().Get("result"),
		AppID:         r.URL.Query().Get("app_id"),
		TraceID:       r.URL.Query().Get("trace_id"),
		Target:        r.URL.Query().Get("target"),
		MetadataKey:   r.URL.Query().Get("metadata_key"),
		MetadataValue: r.URL.Query().Get("metadata_value"),
	})
	writeJSONL(w, "audit-events.jsonl", page.Items)
}

func intQuery(w http.ResponseWriter, r *http.Request, name string, defaultValue int) (int, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return defaultValue, true
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "", "invalid_request", name+" must be a number")
		return 0, false
	}
	return parsed, true
}

func normalizePage(offset, limit int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return offset, limit
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) (config.App, bool) {
	snapshot := h.snapshot()
	app, ok := snapshot.Config.AppByToken(bearerToken(r.Header.Get("Authorization")))
	if !ok || !hasGrant(app.Grants, "admin") {
		writeError(w, http.StatusUnauthorized, "", "unauthorized", "admin grant is required")
		return config.App{}, false
	}
	return app, true
}

func parseProviderAction(path string) (string, string, bool) {
	rest := strings.TrimPrefix(path, "/gateway/v1/providers/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func findTool(cfg config.Config, toolID string) (toolConfigRef, bool) {
	for _, tool := range cfg.Tools {
		if tool.ID == toolID {
			return toolConfigRef{
				ID:              tool.ID,
				Adapter:         tool.Adapter,
				Origin:          "builtin",
				ReadOnly:        tool.ReadOnly,
				Scopes:          append([]string(nil), tool.Scopes...),
				SandboxRequired: tool.SandboxRequired,
				Enabled:         tool.IsEnabled(),
			}, true
		}
	}
	if cfg.MCPRuntime.Enabled {
		for _, server := range cfg.MCPRuntime.Servers {
			for _, tool := range server.Tools {
				if tool.ID == toolID {
					return toolConfigRef{
						ID:              tool.ID,
						Adapter:         "mcp-placeholder",
						Origin:          "mcp",
						ServerID:        server.ID,
						ReadOnly:        tool.ReadOnly,
						Scopes:          append([]string(nil), tool.Scopes...),
						SandboxRequired: tool.SandboxRequired,
						Enabled:         server.IsEnabled() && tool.IsEnabled(),
					}, true
				}
			}
		}
	}
	return toolConfigRef{}, false
}

func (h *Handler) saveAudit(event audit.Event) {
	if h.audit != nil {
		h.audit.Save(event)
	}
}

func (h *Handler) finishToolTrace(record trace.Record, appID, status, message string, startedAt time.Time) {
	finishedAt := time.Now().UTC()
	record.AppID = appID
	record.Status = status
	record.Error = message
	record.FinishedAt = finishedAt
	record.DurationMS = finishedAt.Sub(startedAt).Milliseconds()
	eventType := "tool_invocation_failed"
	if status == "completed" {
		eventType = "tool_invocation_completed"
		record.Error = ""
	}
	record.Events = append(record.Events, trace.Event{Type: eventType, Message: message, At: finishedAt})
	_ = h.traces.Save(record)
}

func (h *Handler) snapshot() gatewayruntime.Snapshot {
	if h.runtime != nil {
		return h.runtime.Snapshot()
	}
	return gatewayruntime.Snapshot{
		Config:   h.cfg,
		Pipeline: h.pipeline,
		Health:   h.providerHealth,
	}
}

func hasGrant(grants []string, want string) bool {
	for _, grant := range grants {
		if grant == want {
			return true
		}
	}
	return false
}

func hasToolScope(grants []string, scopes []string) bool {
	if hasGrant(grants, "tool") {
		return true
	}
	for _, scope := range scopes {
		if !hasGrant(grants, "tool:"+scope) {
			return false
		}
	}
	return len(scopes) > 0
}

func resolveAccessDryRunApp(cfg config.Config, req accessDryRunRequest) (config.App, bool) {
	if req.Token != "" {
		return cfg.AppByToken(req.Token)
	}
	for _, app := range cfg.Apps {
		if app.ID == req.AppID {
			return app, true
		}
	}
	return config.App{}, false
}

func grantIfAllowed(allowed bool, grant string) string {
	if allowed {
		return grant
	}
	return ""
}

func grantsIfDenied(allowed bool, grant string) []string {
	if allowed {
		return []string{}
	}
	return []string{grant}
}

func matchedToolGrant(grants []string, scopes []string) string {
	if hasGrant(grants, "tool") {
		return "tool"
	}
	for _, scope := range scopes {
		if hasGrant(grants, "tool:"+scope) {
			return "tool:" + scope
		}
	}
	return ""
}

func missingToolGrants(scopes []string) []string {
	missing := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		missing = append(missing, "tool:"+scope)
	}
	return missing
}

func toolAuditMetadata(toolCfg toolConfigRef, matchedGrant string, missingGrants []string) map[string]any {
	return map[string]any{
		"adapter":          toolCfg.Adapter,
		"origin":           toolCfg.Origin,
		"server_id":        toolCfg.ServerID,
		"read_only":        toolCfg.ReadOnly,
		"required_scopes":  toolCfg.Scopes,
		"sandbox_required": toolCfg.SandboxRequired,
		"matched_grant":    matchedGrant,
		"missing_grants":   missingGrants,
	}
}

type explainChain struct {
	Stage         string   `json:"stage"`
	Decision      string   `json:"decision"`
	Reason        string   `json:"reason"`
	PolicyRuleID  string   `json:"policy_rule_id,omitempty"`
	RulePriority  int      `json:"rule_priority,omitempty"`
	PolicyVersion string   `json:"policy_version,omitempty"`
	Condition     string   `json:"condition,omitempty"`
	AllowCloud    bool     `json:"allow_cloud,omitempty"`
	ForceLocal    bool     `json:"force_local,omitempty"`
	MatchedGrant  string   `json:"matched_grant,omitempty"`
	MissingGrants []string `json:"missing_grants,omitempty"`
	ToolID        string   `json:"tool_id,omitempty"`
	Candidates    int      `json:"candidate_count,omitempty"`
	Skipped       int      `json:"skipped_count,omitempty"`
	NextAction    string   `json:"next_action"`
}

func policyExplainChain(req policyDryRunRequest, decision policy.Decision) explainChain {
	decisionLabel := "allow"
	nextAction := "route"
	if !decision.Allowed {
		decisionLabel = "deny"
		nextAction = "stop"
	} else if decision.ForceLocal {
		decisionLabel = "force_local"
		nextAction = "route_local_only"
	} else if !decision.AllowCloud {
		decisionLabel = "deny_cloud"
		nextAction = "route_without_cloud"
	}
	return explainChain{
		Stage:         "policy",
		Decision:      decisionLabel,
		Reason:        decision.Explanation,
		PolicyRuleID:  decision.RuleID,
		RulePriority:  decision.RulePriority,
		PolicyVersion: decision.Version,
		Condition:     decision.ConditionSummary,
		AllowCloud:    decision.AllowCloud,
		ForceLocal:    decision.ForceLocal,
		NextAction:    nextAction,
	}
}

func routingExplainChain(req routingExplainRequest, decision policy.Decision, candidates, skipped int) explainChain {
	chain := policyExplainChain(policyDryRunRequest{
		AppID:       req.AppID,
		RequestType: req.RequestType,
		DataLabels:  req.DataLabels,
		Model:       req.Model,
	}, decision)
	chain.Stage = "routing"
	chain.Candidates = candidates
	chain.Skipped = skipped
	if !decision.Allowed {
		chain.NextAction = "stop"
	} else if candidates == 0 {
		chain.Decision = "no_route"
		chain.NextAction = "fix_provider_or_policy"
	} else {
		chain.NextAction = "invoke_provider"
	}
	return chain
}

func accessExplainChain(action string, allowed bool, reason, matchedGrant string, missingGrants []string, toolRef *toolConfigRef) explainChain {
	decision := "deny"
	nextAction := "fix_grants"
	if allowed {
		decision = "allow"
		nextAction = "continue"
	}
	chain := explainChain{
		Stage:         "access",
		Decision:      decision,
		Reason:        reason,
		MatchedGrant:  matchedGrant,
		MissingGrants: missingGrants,
		NextAction:    nextAction,
	}
	if toolRef != nil {
		chain.ToolID = toolRef.ID
	}
	return chain
}

func accessDryRunAuditMetadata(action string, allowed bool, reason, matchedGrant string, missingGrants []string, toolRef *toolConfigRef) map[string]any {
	metadata := map[string]any{
		"allowed":        allowed,
		"reason":         reason,
		"matched_grant":  matchedGrant,
		"missing_grants": missingGrants,
		"explain_chain":  accessExplainChain(action, allowed, reason, matchedGrant, missingGrants, toolRef),
	}
	if toolRef != nil {
		for key, value := range toolAuditMetadata(*toolRef, matchedGrant, missingGrants) {
			metadata[key] = value
		}
		metadata["tool_id"] = toolRef.ID
	}
	return metadata
}

func hasScope(scopes []string, want string) bool {
	for _, scope := range scopes {
		if scope == want {
			return true
		}
	}
	return false
}

func toolErrorResponse(err error) (string, int) {
	var toolErr *tools.Error
	if errors.As(err, &toolErr) && toolErr.Code != "" {
		switch toolErr.Code {
		case tools.ErrCodeUnavailable:
			return toolErr.Code, http.StatusServiceUnavailable
		default:
			return toolErr.Code, http.StatusBadGateway
		}
	}
	return "tool_failed", http.StatusBadGateway
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(header, prefix))
	}
	return ""
}

func tokenHint(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func ensureGrantView(grants map[string]*grantView, id, grantType, description string) *grantView {
	if view, ok := grants[id]; ok {
		return view
	}
	view := &grantView{ID: id, Type: grantType, Description: description}
	grants[id] = view
	return view
}

func grantType(grant string) string {
	if strings.HasPrefix(grant, "tool:") {
		return "tool_scope"
	}
	switch grant {
	case "tool":
		return "tool_broad"
	case "admin":
		return "admin"
	default:
		return "core"
	}
}

func grantDescription(grant string) string {
	switch {
	case grant == "chat":
		return "Allow OpenAI-compatible chat completions."
	case grant == "embedding":
		return "Reserved grant for embedding requests."
	case grant == "tool":
		return "Allow invoking any enabled read-only tool."
	case grant == "admin":
		return "Allow management APIs such as audit, config reload, providers, apps and grants."
	case strings.HasPrefix(grant, "tool:"):
		scope := strings.TrimPrefix(grant, "tool:")
		return "Allow invoking tools that require scope " + scope + "."
	default:
		return "Configured application grant."
	}
}

func appendUnique(values []string, value string) []string {
	if containsString(values, value) {
		return values
	}
	return append(values, value)
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func sortGrantViews(views []grantView) {
	sort.Slice(views, func(i, j int) bool {
		if views[i].Type != views[j].Type {
			return grantTypeRank(views[i].Type) < grantTypeRank(views[j].Type)
		}
		return views[i].ID < views[j].ID
	})
	for i := range views {
		sort.Strings(views[i].Apps)
		sort.Strings(views[i].Tools)
		sort.Strings(views[i].Servers)
	}
}

func grantTypeRank(value string) int {
	switch value {
	case "core":
		return 0
	case "tool_broad":
		return 1
	case "tool_scope":
		return 2
	case "admin":
		return 3
	default:
		return 4
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONL(w http.ResponseWriter, filename string, items any) {
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	switch values := items.(type) {
	case []trace.Record:
		for _, item := range values {
			_ = json.NewEncoder(w).Encode(item)
		}
	case []audit.Event:
		for _, item := range values {
			_ = json.NewEncoder(w).Encode(item)
		}
	case []providerhealth.View:
		for _, item := range values {
			_ = json.NewEncoder(w).Encode(item)
		}
	case []modelView:
		for _, item := range values {
			_ = json.NewEncoder(w).Encode(item)
		}
	case []mcpServerView:
		for _, item := range values {
			_ = json.NewEncoder(w).Encode(item)
		}
	case []toolView:
		for _, item := range values {
			_ = json.NewEncoder(w).Encode(item)
		}
	case []appView:
		for _, item := range values {
			_ = json.NewEncoder(w).Encode(item)
		}
	case []grantView:
		for _, item := range values {
			_ = json.NewEncoder(w).Encode(item)
		}
	case []policyView:
		for _, item := range values {
			_ = json.NewEncoder(w).Encode(item)
		}
	}
}

func writeError(w http.ResponseWriter, status int, traceID, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":     code,
			"message":  message,
			"trace_id": traceID,
		},
	})
}
