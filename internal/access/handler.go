package access

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
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
	mux.HandleFunc("GET /gateway/v1/models", h.models)
	mux.HandleFunc("GET /gateway/v1/runtime/health", h.runtimeHealth)
	mux.HandleFunc("GET /gateway/v1/tools", h.toolsList)
	mux.HandleFunc("POST /gateway/v1/tools/", h.toolInvoke)
	mux.HandleFunc("POST /gateway/v1/providers/", h.providerAction)
	mux.HandleFunc("GET /gateway/v1/audit/events", h.auditList)
	mux.HandleFunc("GET /gateway/v1/audit/events/export", h.auditExport)
	mux.HandleFunc("POST /gateway/v1/config/reload", h.configReload)
	mux.HandleFunc("POST /gateway/v1/policy/dry-run", h.policyDryRun)
	mux.HandleFunc("POST /gateway/v1/routing/explain", h.routingExplain)
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

func (h *Handler) providers(w http.ResponseWriter, _ *http.Request) {
	snapshot := h.snapshot()
	if snapshot.Health != nil {
		writeJSON(w, http.StatusOK, map[string]any{"providers": snapshot.Health.Views(snapshot.Config.Providers)})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"providers": snapshot.Config.Providers})
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
	snapshot := h.snapshot()
	onlyAvailable := r.URL.Query().Get("all") != "true"
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
	writeJSON(w, http.StatusOK, map[string]any{"models": models})
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
	Description     string         `json:"description,omitempty"`
	ReadOnly        bool           `json:"read_only"`
	RiskLevel       string         `json:"risk_level,omitempty"`
	Scopes          []string       `json:"scopes,omitempty"`
	InputSchema     map[string]any `json:"input_schema,omitempty"`
	OutputSchema    map[string]any `json:"output_schema,omitempty"`
	SandboxRequired bool           `json:"sandbox_required"`
	Enabled         bool           `json:"enabled"`
}

func (h *Handler) toolsList(w http.ResponseWriter, _ *http.Request) {
	snapshot := h.snapshot()
	views := make([]toolView, 0, len(snapshot.Config.Tools))
	for _, tool := range snapshot.Config.Tools {
		manifest := tools.ManifestFromConfig(tool)
		if snapshot.Tools != nil {
			if registered, ok := snapshot.Tools.Get(tool.ID); ok {
				manifest = registered.Manifest()
			}
		}
		views = append(views, toolView{
			ID:              manifest.ID,
			Name:            manifest.Name,
			Adapter:         manifest.Adapter,
			Description:     manifest.Description,
			ReadOnly:        manifest.ReadOnly,
			RiskLevel:       manifest.RiskLevel,
			Scopes:          manifest.Scopes,
			InputSchema:     manifest.InputSchema,
			OutputSchema:    manifest.OutputSchema,
			SandboxRequired: manifest.SandboxRequired,
			Enabled:         tool.IsEnabled(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"tools": views})
}

type toolInvokeRequest struct {
	Arguments map[string]any `json:"arguments,omitempty"`
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
	toolCfg, ok := findTool(snapshot.Config.Tools, toolID)
	if !ok || !toolCfg.IsEnabled() {
		h.finishToolTrace(record, app.ID, "failed", "tool not found or disabled", startedAt)
		h.saveAudit(audit.Event{TraceID: traceID, AppID: app.ID, Action: "tool.invoke", Target: toolID, Result: audit.ResultFailed, Error: "tool not found or disabled"})
		writeError(w, http.StatusNotFound, traceID, "not_found", "tool not found")
		return
	}
	if !hasToolScope(app.Grants, toolCfg.Scopes) {
		h.finishToolTrace(record, app.ID, "failed", "tool scope is required", startedAt)
		h.saveAudit(audit.Event{
			TraceID: traceID,
			AppID:   app.ID,
			Action:  "tool.invoke",
			Target:  toolID,
			Result:  audit.ResultDenied,
			Error:   "tool scope is required",
			Metadata: map[string]any{
				"required_scopes": toolCfg.Scopes,
			},
		})
		writeError(w, http.StatusForbidden, traceID, "tool_scope_denied", "tool scope is required")
		return
	}
	if !toolCfg.ReadOnly {
		h.finishToolTrace(record, app.ID, "failed", "only read-only tools are allowed", startedAt)
		h.saveAudit(audit.Event{TraceID: traceID, AppID: app.ID, Action: "tool.invoke", Target: toolID, Result: audit.ResultDenied, Error: "only read-only tools are allowed"})
		writeError(w, http.StatusForbidden, traceID, "tool_denied", "only read-only tools are allowed")
		return
	}
	if snapshot.Tools == nil {
		h.finishToolTrace(record, app.ID, "failed", "tool registry is not configured", startedAt)
		writeError(w, http.StatusServiceUnavailable, traceID, "tool_unavailable", "tool registry is not configured")
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
			Metadata: map[string]any{
				"adapter":   toolCfg.Adapter,
				"read_only": toolCfg.ReadOnly,
			},
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
		Metadata: map[string]any{
			"adapter":   toolCfg.Adapter,
			"read_only": toolCfg.ReadOnly,
		},
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
		AppID:  req.AppID,
		Action: "policy.dry_run",
		Target: req.Model,
		Result: audit.ResultSuccess,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"decision": decision,
		"input":    req,
	})
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
	writeJSON(w, http.StatusOK, map[string]any{
		"input":      req,
		"policy":     decision,
		"candidates": explanation.Candidates,
		"skipped":    explanation.Skipped,
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
		Offset:  offset,
		Limit:   limit,
		Action:  r.URL.Query().Get("action"),
		Result:  r.URL.Query().Get("result"),
		AppID:   r.URL.Query().Get("app_id"),
		TraceID: r.URL.Query().Get("trace_id"),
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
		Offset:  offset,
		Limit:   limit,
		Action:  r.URL.Query().Get("action"),
		Result:  r.URL.Query().Get("result"),
		AppID:   r.URL.Query().Get("app_id"),
		TraceID: r.URL.Query().Get("trace_id"),
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

func findTool(tools []config.Tool, toolID string) (config.Tool, bool) {
	for _, tool := range tools {
		if tool.ID == toolID {
			return tool, true
		}
	}
	return config.Tool{}, false
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
