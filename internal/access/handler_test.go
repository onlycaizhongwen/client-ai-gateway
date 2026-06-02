package access

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"client-ai-gateway/internal/adapters"
	"client-ai-gateway/internal/audit"
	"client-ai-gateway/internal/config"
	"client-ai-gateway/internal/core"
	"client-ai-gateway/internal/fallback"
	"client-ai-gateway/internal/policy"
	"client-ai-gateway/internal/providerhealth"
	"client-ai-gateway/internal/router"
	gatewayruntime "client-ai-gateway/internal/runtime"
	"client-ai-gateway/internal/trace"
	"os"
	"path/filepath"
)

func TestChatSuccessHTTP(t *testing.T) {
	handler, _ := newTestHandler()
	res := postJSON(handler, "/v1/chat/completions", "dev-token", map[string]any{
		"model":    "local-small",
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
	})

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body map[string]any
	decodeBody(t, res, &body)
	if body["trace_id"] == "" {
		t.Fatal("expected trace_id")
	}
}

func TestChatUnauthorizedHTTPIncludesTraceID(t *testing.T) {
	handler, store := newTestHandler()
	res := postJSON(handler, "/v1/chat/completions", "bad-token", map[string]any{
		"model":    "local-small",
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
	})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", res.Code, res.Body.String())
	}
	traceID := errorTraceID(t, res)
	if traceID == "" {
		t.Fatal("expected trace id")
	}
	if _, ok := store.Get(traceID); !ok {
		t.Fatal("expected trace record for unauthorized request")
	}
}

func TestChatInvalidRequestHTTP(t *testing.T) {
	handler, _ := newTestHandler()
	res := postJSON(handler, "/v1/chat/completions", "dev-token", map[string]any{
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
	})

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", res.Code, res.Body.String())
	}
	if traceID := errorTraceID(t, res); traceID != "" {
		t.Fatalf("invalid request should not reach core pipeline, got trace %s", traceID)
	}
}

func TestChatFallbackAllowedHTTP(t *testing.T) {
	handler, store := newTestHandler()
	res := postJSON(handler, "/v1/chat/completions", "dev-token", map[string]any{
		"model":    "local-small",
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
		"metadata": map[string]string{"fail_provider": "local-mock"},
	})

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body map[string]any
	decodeBody(t, res, &body)
	record, ok := store.Get(body["trace_id"].(string))
	if !ok {
		t.Fatal("expected trace record")
	}
	if record.ProviderID != "cloud-mock" || len(record.Fallbacks) != 1 {
		t.Fatalf("expected cloud fallback, got provider=%s fallbacks=%d", record.ProviderID, len(record.Fallbacks))
	}
}

func TestChatFallbackBlockedHTTPIncludesTraceID(t *testing.T) {
	handler, store := newTestHandler()
	res := postJSON(handler, "/v1/chat/completions", "dev-token", map[string]any{
		"model":       "local-small",
		"messages":    []map[string]string{{"role": "user", "content": "secret"}},
		"data_labels": []string{"sensitive"},
		"metadata":    map[string]string{"fail_provider": "local-mock"},
	})

	if res.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", res.Code, res.Body.String())
	}
	traceID := errorTraceID(t, res)
	if traceID == "" {
		t.Fatal("expected trace id")
	}
	record, ok := store.Get(traceID)
	if !ok {
		t.Fatal("expected failed trace record")
	}
	if record.Status != "failed" || len(record.Fallbacks) != 1 {
		t.Fatalf("expected failed fallback trace, got status=%s fallbacks=%d", record.Status, len(record.Fallbacks))
	}
}

func TestTraceListFiltersHTTP(t *testing.T) {
	handler, _ := newTestHandler()
	postJSON(handler, "/v1/chat/completions", "dev-token", map[string]any{
		"model":    "local-small",
		"messages": []map[string]string{{"role": "user", "content": "ok"}},
	})
	postJSON(handler, "/v1/chat/completions", "dev-token", map[string]any{
		"model":       "local-small",
		"messages":    []map[string]string{{"role": "user", "content": "bad"}},
		"data_labels": []string{"sensitive"},
		"metadata":    map[string]string{"fail_provider": "local-mock"},
	})

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/traces?status=failed", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var body struct {
		Traces []trace.Record `json:"traces"`
	}
	decodeBody(t, res, &body)
	if len(body.Traces) != 1 || body.Traces[0].Status != "failed" {
		t.Fatalf("expected one failed trace, got %+v", body.Traces)
	}
}

func TestTraceListPaginationHTTP(t *testing.T) {
	handler, _ := newTestHandler()
	for i := 0; i < 3; i++ {
		postJSON(handler, "/v1/chat/completions", "dev-token", map[string]any{
			"model":    "local-small",
			"messages": []map[string]string{{"role": "user", "content": "page"}},
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/traces?limit=1&offset=1", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var body struct {
		Traces []trace.Record `json:"traces"`
		Total  int            `json:"total"`
		Offset int            `json:"offset"`
		Limit  int            `json:"limit"`
	}
	decodeBody(t, res, &body)
	if len(body.Traces) != 1 || body.Total != 3 || body.Offset != 1 || body.Limit != 1 {
		t.Fatalf("unexpected paged trace body: %+v", body)
	}
}

func TestTraceExportHTTP(t *testing.T) {
	handler, _ := newTestHandler()
	postJSON(handler, "/v1/chat/completions", "dev-token", map[string]any{
		"model":    "local-small",
		"messages": []map[string]string{{"role": "user", "content": "export me"}},
	})

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/traces/export?limit=1", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "application/x-ndjson" {
		t.Fatalf("unexpected content type %q", got)
	}
	if !strings.Contains(res.Body.String(), `"trace_id":"trace-`) || !strings.Contains(res.Body.String(), `"status":"completed"`) {
		t.Fatalf("unexpected export body: %s", res.Body.String())
	}
}

func TestAccessLogIncludesTraceID(t *testing.T) {
	var logs bytes.Buffer
	handler, _ := newTestHandlerWithLogger(slog.New(slog.NewJSONHandler(&logs, nil)))
	res := postJSON(handler, "/v1/chat/completions", "dev-token", map[string]any{
		"model":    "local-small",
		"messages": []map[string]string{{"role": "user", "content": "logged"}},
	})
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	line := logs.String()
	if !strings.Contains(line, `"msg":"http_request"`) {
		t.Fatalf("expected access log, got %s", line)
	}
	if !strings.Contains(line, `"trace_id":"trace-`) {
		t.Fatalf("expected trace_id in log, got %s", line)
	}
	if !strings.Contains(line, `"app_id":"dev-app"`) {
		t.Fatalf("expected app_id in log, got %s", line)
	}
}

func TestProvidersHTTPIncludesRuntimeHealth(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/providers", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var body struct {
		Providers []providerhealth.View `json:"providers"`
	}
	decodeBody(t, res, &body)
	if len(body.Providers) != 2 {
		t.Fatalf("expected two providers, got %+v", body.Providers)
	}
	if body.Providers[0].RuntimeStatus == "" || body.Providers[0].Adapter == "" {
		t.Fatalf("expected runtime health view, got %+v", body.Providers[0])
	}
}

func TestModelsHTTP(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/models", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var body struct {
		Models []modelView `json:"models"`
	}
	decodeBody(t, res, &body)
	if len(body.Models) != 3 {
		t.Fatalf("expected three available model entries, got %+v", body.Models)
	}
	if body.Models[0].Model == "" || body.Models[0].ProviderID == "" {
		t.Fatalf("expected model catalog details, got %+v", body.Models[0])
	}
}

func TestRuntimeHealthHTTP(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [
	    {"id":"dev-app","token":"dev-token","grants":["chat"]},
	    {"id":"admin-app","token":"admin-token","grants":["admin"]}
	  ],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}]
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	handler := NewRuntimeHandler(manager, store).Routes()

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/runtime/health", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var body struct {
		Status          string `json:"status"`
		ProviderMonitor struct {
			Status string `json:"status"`
			Total  int    `json:"total"`
		} `json:"provider_monitor"`
		ModelRuntime struct {
			Status string `json:"status"`
		} `json:"model_runtime"`
	}
	decodeBody(t, res, &body)
	if body.Status != "healthy" || body.ProviderMonitor.Status != "running" || body.ProviderMonitor.Total != 1 {
		t.Fatalf("unexpected runtime health body: %+v", body)
	}
	if body.ModelRuntime.Status != "not_configured" {
		t.Fatalf("expected model runtime placeholder, got %+v", body.ModelRuntime)
	}
}

func TestAppsListRequiresAdmin(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/apps", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without admin token, got %d", res.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/gateway/v1/apps", nil)
	req.Header.Set("Authorization", "Bearer dev-token")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with non-admin token, got %d", res.Code)
	}
}

func TestAppsListMasksTokensAndFilters(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/apps?grant=tool&limit=1&offset=0", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "dev-token") || strings.Contains(res.Body.String(), "admin-token") {
		t.Fatalf("apps list leaked raw token: %s", res.Body.String())
	}
	var body struct {
		Apps   []appView `json:"apps"`
		Total  int       `json:"total"`
		Offset int       `json:"offset"`
		Limit  int       `json:"limit"`
	}
	decodeBody(t, res, &body)
	if body.Total != 1 || body.Offset != 0 || body.Limit != 1 || len(body.Apps) != 1 {
		t.Fatalf("unexpected app page body: %+v", body)
	}
	if body.Apps[0].ID != "dev-app" || body.Apps[0].TokenHint != "dev-...oken" || !hasGrant(body.Apps[0].Grants, "tool") {
		t.Fatalf("unexpected app view: %+v", body.Apps[0])
	}

	req = httptest.NewRequest(http.MethodGet, "/gateway/v1/apps?app_id=admin-app", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	decodeBody(t, res, &body)
	if body.Total != 1 || len(body.Apps) != 1 || body.Apps[0].ID != "admin-app" || body.Apps[0].TokenHint != "admi...oken" {
		t.Fatalf("unexpected filtered admin app body: %+v", body)
	}
}

func TestToolsListHTTP(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/tools", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Tools []toolView `json:"tools"`
		Total int        `json:"total"`
		Limit int        `json:"limit"`
	}
	decodeBody(t, res, &body)
	if body.Total != 1 || body.Limit != 100 || len(body.Tools) != 1 || body.Tools[0].ID != "gateway.runtime_health" || !body.Tools[0].ReadOnly || body.Tools[0].RiskLevel != "low" {
		t.Fatalf("unexpected tools body: %+v", body)
	}
}

func TestToolsListIncludesMCPPlaceholderManifests(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat","tool"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "tools": [{"id":"gateway.runtime_health","name":"Runtime Health","adapter":"runtime-health","read_only":true,"risk_level":"low","scopes":["runtime.read"],"enabled":true}],
	  "mcp_runtime": {"enabled":true,"mode":"manifest_only","servers":[{"id":"desktop-context","tools":[{"id":"mcp.desktop.list_context","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":true}]}]}
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	handler := NewRuntimeHandler(manager, store).Routes()

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/tools", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Tools []toolView `json:"tools"`
	}
	decodeBody(t, res, &body)
	if len(body.Tools) != 2 {
		t.Fatalf("expected builtin and mcp placeholder tools, got %+v", body.Tools)
	}
	mcpTool := body.Tools[1]
	if mcpTool.ID != "mcp.desktop.list_context" || mcpTool.Origin != "mcp" || mcpTool.ServerID != "desktop-context" || mcpTool.Adapter != "mcp-placeholder" || !mcpTool.Enabled {
		t.Fatalf("unexpected mcp tool view: %+v", mcpTool)
	}
}

func TestToolsListFiltersAndPages(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat","tool"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "tools": [{"id":"gateway.runtime_health","name":"Runtime Health","adapter":"runtime-health","read_only":true,"risk_level":"low","scopes":["runtime.read"],"enabled":true}],
	  "mcp_runtime": {"enabled":true,"mode":"manifest_only","servers":[
	    {"id":"desktop-context","tools":[
	      {"id":"mcp.desktop.list_context","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":true},
	      {"id":"mcp.desktop.disabled","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":false}
	    ]},
	    {"id":"repo-context","tools":[
	      {"id":"mcp.repo.list","read_only":true,"risk_level":"low","scopes":["repo.read"],"enabled":true}
	    ]}
	  ]}
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	handler := NewRuntimeHandler(manager, store).Routes()

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/tools?origin=mcp&server_id=desktop-context&scope=desktop.read&enabled=true&limit=1&offset=0", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Tools  []toolView `json:"tools"`
		Total  int        `json:"total"`
		Offset int        `json:"offset"`
		Limit  int        `json:"limit"`
	}
	decodeBody(t, res, &body)
	if body.Total != 1 || body.Offset != 0 || body.Limit != 1 || len(body.Tools) != 1 {
		t.Fatalf("unexpected paged tools body: %+v", body)
	}
	if body.Tools[0].ID != "mcp.desktop.list_context" {
		t.Fatalf("unexpected filtered tool: %+v", body.Tools[0])
	}
}

func TestToolsExportHTTP(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat","tool"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "tools": [{"id":"gateway.runtime_health","name":"Runtime Health","adapter":"runtime-health","read_only":true,"risk_level":"low","scopes":["runtime.read"],"enabled":true}],
	  "mcp_runtime": {"enabled":true,"mode":"manifest_only","servers":[{"id":"desktop-context","tools":[{"id":"mcp.desktop.list_context","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":true}]}]}
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	handler := NewRuntimeHandler(manager, store).Routes()

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/tools/export?origin=mcp&scope=desktop.read", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "application/x-ndjson" {
		t.Fatalf("expected jsonl content type, got %q", got)
	}
	body := res.Body.String()
	if !strings.Contains(body, `"id":"mcp.desktop.list_context"`) || strings.Contains(body, `"id":"gateway.runtime_health"`) {
		t.Fatalf("unexpected tools export body: %s", body)
	}
}

func TestMCPPlaceholderToolInvocationFailsClosed(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat","tool:desktop.read"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "tools": [{"id":"gateway.runtime_health","name":"Runtime Health","adapter":"runtime-health","read_only":true,"risk_level":"low","scopes":["runtime.read"],"enabled":true}],
	  "mcp_runtime": {"enabled":true,"mode":"manifest_only","servers":[{"id":"desktop-context","tools":[{"id":"mcp.desktop.list_context","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":true}]}]}
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	auditStore := audit.NewMemoryStore()
	handler := NewRuntimeHandler(manager, store).WithAudit(auditStore).Routes()

	res := postJSON(handler, "/gateway/v1/tools/mcp.desktop.list_context/invoke", "dev-token", map[string]any{})
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", res.Code, res.Body.String())
	}
	traceID := errorTraceID(t, res)
	record, ok := store.Get(traceID)
	if !ok {
		t.Fatalf("expected mcp placeholder trace %s to be saved", traceID)
	}
	if record.ToolID != "mcp.desktop.list_context" || record.Status != "failed" || !strings.Contains(record.Error, "manifest-only") {
		t.Fatalf("unexpected mcp placeholder trace: %+v", record)
	}
	events := auditStore.List(audit.ListQuery{Action: "tool.invoke", TraceID: traceID})
	if len(events) != 1 || events[0].Result != audit.ResultFailed || events[0].Metadata["origin"] != "mcp" || events[0].Metadata["server_id"] != "desktop-context" {
		t.Fatalf("expected mcp placeholder audit metadata, got %+v", events)
	}
}

func TestMCPServersHTTP(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "mcp_runtime": {"enabled":true,"mode":"manifest_only","servers":[{"id":"desktop-context","name":"Desktop Context","tools":[{"id":"mcp.desktop.list_context","name":"List Context","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":true}]}]}
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	handler := NewRuntimeHandler(manager, store).Routes()

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/mcp/servers", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Enabled bool            `json:"enabled"`
		Mode    string          `json:"mode"`
		Servers []mcpServerView `json:"servers"`
	}
	decodeBody(t, res, &body)
	if !body.Enabled || body.Mode != "manifest_only" || len(body.Servers) != 1 {
		t.Fatalf("unexpected mcp catalog body: %+v", body)
	}
	server := body.Servers[0]
	if server.ID != "desktop-context" || !server.Enabled || server.ToolCount != 1 || server.EnabledTools != 1 {
		t.Fatalf("unexpected mcp server view: %+v", server)
	}
	if len(server.Tools) != 1 || server.Tools[0].ID != "mcp.desktop.list_context" || server.Tools[0].Origin != "mcp" {
		t.Fatalf("unexpected mcp tool view: %+v", server.Tools)
	}
}

func TestMCPServersHTTPFilters(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "mcp_runtime": {"enabled":true,"mode":"manifest_only","servers":[
	    {"id":"desktop-context","tools":[
	      {"id":"mcp.desktop.list_context","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":true},
	      {"id":"mcp.desktop.disabled","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":false}
	    ]},
	    {"id":"repo-context","tools":[
	      {"id":"mcp.repo.list","read_only":true,"risk_level":"low","scopes":["repo.read"],"enabled":true}
	    ]}
	  ]}
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	handler := NewRuntimeHandler(manager, store).Routes()

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/mcp/servers?server_id=desktop-context&scope=desktop.read&enabled=true", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Filters map[string]string `json:"filters"`
		Servers []mcpServerView   `json:"servers"`
	}
	decodeBody(t, res, &body)
	if body.Filters["server_id"] != "desktop-context" || body.Filters["scope"] != "desktop.read" || body.Filters["enabled"] != "true" {
		t.Fatalf("unexpected filters echo: %+v", body.Filters)
	}
	if len(body.Servers) != 1 || body.Servers[0].ID != "desktop-context" {
		t.Fatalf("expected only desktop-context server, got %+v", body.Servers)
	}
	if body.Servers[0].ToolCount != 1 || body.Servers[0].EnabledTools != 1 || len(body.Servers[0].Tools) != 1 {
		t.Fatalf("expected only enabled desktop.read tool, got %+v", body.Servers[0])
	}
	if body.Servers[0].Tools[0].ID != "mcp.desktop.list_context" {
		t.Fatalf("unexpected filtered tool: %+v", body.Servers[0].Tools[0])
	}
}

func TestMCPServersHTTPPagination(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "mcp_runtime": {"enabled":true,"mode":"manifest_only","servers":[
	    {"id":"desktop-context","tools":[{"id":"mcp.desktop.list_context","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":true}]},
	    {"id":"repo-context","tools":[{"id":"mcp.repo.list","read_only":true,"risk_level":"low","scopes":["repo.read"],"enabled":true}]},
	    {"id":"browser-context","tools":[{"id":"mcp.browser.list","read_only":true,"risk_level":"low","scopes":["browser.read"],"enabled":true}]}
	  ]}
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	handler := NewRuntimeHandler(manager, store).Routes()

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/mcp/servers?limit=1&offset=1", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Servers []mcpServerView `json:"servers"`
		Total   int             `json:"total"`
		Offset  int             `json:"offset"`
		Limit   int             `json:"limit"`
	}
	decodeBody(t, res, &body)
	if body.Total != 3 || body.Offset != 1 || body.Limit != 1 || len(body.Servers) != 1 {
		t.Fatalf("unexpected paged mcp body: %+v", body)
	}
	if body.Servers[0].ID != "repo-context" {
		t.Fatalf("unexpected paged mcp server: %+v", body.Servers[0])
	}
}

func TestMCPServersExportHTTP(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "mcp_runtime": {"enabled":true,"mode":"manifest_only","servers":[{"id":"desktop-context","tools":[{"id":"mcp.desktop.list_context","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":true}]}]}
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	handler := NewRuntimeHandler(manager, store).Routes()

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/mcp/servers/export?scope=desktop.read", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "application/x-ndjson" {
		t.Fatalf("expected jsonl content type, got %q", got)
	}
	if !strings.Contains(res.Body.String(), `"id":"desktop-context"`) || !strings.Contains(res.Body.String(), `"mcp.desktop.list_context"`) {
		t.Fatalf("unexpected mcp export body: %s", res.Body.String())
	}
}

func TestToolInvokeRequiresScope(t *testing.T) {
	handler, store, cleanup := newRuntimeToolTestHandler(t, `[
	    {"id":"dev-app","token":"dev-token","grants":["chat"]},
	    {"id":"admin-app","token":"admin-token","grants":["admin"]}
	  ]`)
	defer cleanup()

	res := postJSON(handler, "/gateway/v1/tools/gateway.runtime_health/invoke", "admin-token", map[string]any{})
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", res.Code, res.Body.String())
	}
	traceID := errorTraceID(t, res)
	if traceID == "" {
		t.Fatal("expected denied tool invocation to include trace_id")
	}
	record, ok := store.Get(traceID)
	if !ok {
		t.Fatalf("expected denied tool trace %s to be saved", traceID)
	}
	if record.RequestType != "tool" || record.ToolID != "gateway.runtime_health" || record.Status != "failed" {
		t.Fatalf("unexpected denied tool trace: %+v", record)
	}
}

func TestToolInvokeAllowsScopedGrant(t *testing.T) {
	handler, _, cleanup := newRuntimeToolTestHandler(t, `[
	    {"id":"dev-app","token":"dev-token","grants":["chat","tool:runtime.read"]},
	    {"id":"admin-app","token":"admin-token","grants":["admin"]}
	  ]`)
	defer cleanup()

	res := postJSON(handler, "/gateway/v1/tools/gateway.runtime_health/invoke", "dev-token", map[string]any{})
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
}

func TestToolInvokeRuntimeHealthHTTP(t *testing.T) {
	handler, store, cleanup := newRuntimeToolTestHandler(t, `[
	    {"id":"dev-app","token":"dev-token","grants":["chat","tool"]},
	    {"id":"admin-app","token":"admin-token","grants":["admin"]}
	  ]`)
	defer cleanup()

	res := postJSON(handler, "/gateway/v1/tools/gateway.runtime_health/invoke", "dev-token", map[string]any{})
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		ToolID     string `json:"tool_id"`
		TraceID    string `json:"trace_id"`
		AppID      string `json:"app_id"`
		DurationMS int64  `json:"duration_ms"`
		Output     struct {
			Status string `json:"status"`
		} `json:"output"`
	}
	decodeBody(t, res, &body)
	if body.ToolID != "gateway.runtime_health" || body.Output.Status != "healthy" {
		t.Fatalf("unexpected tool result: %+v", body)
	}
	if body.AppID != "dev-app" || body.DurationMS < 0 {
		t.Fatalf("expected tool result observability fields, got %+v", body)
	}
	if body.TraceID == "" {
		t.Fatalf("expected tool trace_id, got %+v", body)
	}
	record, ok := store.Get(body.TraceID)
	if !ok {
		t.Fatalf("expected tool trace %s to be saved", body.TraceID)
	}
	if record.RequestType != "tool" || record.ToolID != "gateway.runtime_health" || record.AppID != "dev-app" || record.Status != "completed" {
		t.Fatalf("unexpected tool trace record: %+v", record)
	}
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/audit/events?action=tool.invoke", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	auditRes := httptest.NewRecorder()
	handler.ServeHTTP(auditRes, req)
	if auditRes.Code != http.StatusOK {
		t.Fatalf("expected audit 200, got %d: %s", auditRes.Code, auditRes.Body.String())
	}
	var auditBody struct {
		Events []audit.Event `json:"events"`
	}
	decodeBody(t, auditRes, &auditBody)
	events := auditBody.Events
	if len(events) != 1 || events[0].Result != audit.ResultSuccess || events[0].Target != "gateway.runtime_health" {
		t.Fatalf("expected tool invoke audit success, got %+v", events)
	}
	if events[0].TraceID != body.TraceID {
		t.Fatalf("expected audit trace_id %s, got %+v", body.TraceID, events[0])
	}
	if events[0].DurationMS < 0 || events[0].Metadata["adapter"] != "runtime-health" || events[0].Metadata["read_only"] != true {
		t.Fatalf("expected tool invoke audit metadata, got %+v", events[0])
	}
	if events[0].Metadata["matched_grant"] != "tool" || events[0].Metadata["sandbox_required"] != false {
		t.Fatalf("expected tool grant and sandbox audit metadata, got %+v", events[0].Metadata)
	}
	if scopes, ok := events[0].Metadata["required_scopes"].([]any); !ok || len(scopes) != 1 || scopes[0] != "runtime.read" {
		t.Fatalf("expected required scopes audit metadata, got %+v", events[0].Metadata["required_scopes"])
	}
}

func TestAuditListFiltersByTraceID(t *testing.T) {
	handler, _, cleanup := newRuntimeToolTestHandler(t, `[
	    {"id":"dev-app","token":"dev-token","grants":["chat","tool"]},
	    {"id":"admin-app","token":"admin-token","grants":["admin"]}
	  ]`)
	defer cleanup()

	first := postJSON(handler, "/gateway/v1/tools/gateway.runtime_health/invoke", "dev-token", map[string]any{})
	if first.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", first.Code, first.Body.String())
	}
	var firstBody struct {
		TraceID string `json:"trace_id"`
	}
	decodeBody(t, first, &firstBody)

	second := postJSON(handler, "/gateway/v1/tools/gateway.runtime_health/invoke", "dev-token", map[string]any{})
	if second.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", second.Code, second.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/audit/events?trace_id="+firstBody.TraceID, nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Events []audit.Event `json:"events"`
		Total  int           `json:"total"`
	}
	decodeBody(t, res, &body)
	if body.Total != 1 || len(body.Events) != 1 || body.Events[0].TraceID != firstBody.TraceID {
		t.Fatalf("expected one audit event for trace %s, got %+v", firstBody.TraceID, body)
	}
}

func newRuntimeToolTestHandler(t *testing.T, appsJSON string) (http.Handler, *trace.MemoryStore, func()) {
	t.Helper()
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": `+appsJSON+`,
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "tools": [{"id":"gateway.runtime_health","name":"Runtime Health","adapter":"runtime-health","read_only":true,"risk_level":"low","scopes":["runtime.read"],"enabled":true}]
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	auditStore := audit.NewMemoryStore()
	handler := NewRuntimeHandler(manager, store).WithAudit(auditStore).Routes()
	return handler, store, manager.Close
}

func TestConsoleIncludesExportActions(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/console", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	body := res.Body.String()
	for _, want := range []string{
		"id=\"trace-export\"",
		"id=\"trace-app-filter\"",
		"id=\"trace-provider-filter\"",
		"id=\"trace-filter-apply\"",
		"id=\"audit-export\"",
		"id=\"audit-action-filter\"",
		"id=\"audit-result-filter\"",
		"id=\"audit-app-filter\"",
		"id=\"audit-trace-filter\"",
		"id=\"audit-filter-apply\"",
		"id=\"audit-detail\"",
		"id=\"access-app-id\"",
		"id=\"access-token\"",
		"id=\"access-action\"",
		"id=\"access-tool-id\"",
		"id=\"access-dry-run\"",
		"id=\"access-result\"",
		"id=\"app-id-filter\"",
		"id=\"app-grant-filter\"",
		"id=\"app-filter-apply\"",
		"id=\"app-refresh\"",
		"id=\"app-rows\"",
		"id=\"app-prev\"",
		"id=\"app-next\"",
		"id=\"tool-select\"",
		"id=\"tool-invoke\"",
		"id=\"tool-export\"",
		"id=\"tool-rows\"",
		"id=\"tool-prev\"",
		"id=\"tool-next\"",
		"id=\"tool-origin-filter\"",
		"id=\"tool-scope-filter\"",
		"id=\"mcp-server-filter\"",
		"id=\"mcp-scope-filter\"",
		"id=\"mcp-enabled-filter\"",
		"id=\"mcp-export\"",
		"id=\"mcp-rows\"",
		"id=\"mcp-prev\"",
		"id=\"mcp-next\"",
		"function loadMCPCatalog()",
		"function renderMCPServers",
		"function exportMCPCatalog()",
		"function exportTraces()",
		"function exportAudit()",
		"function auditQuery",
		"function showAuditDetail",
		"function accessDryRun",
		"function loadApps()",
		"function appCatalogQuery()",
		"function renderApps()",
		"function loadTools()",
		"function toolCatalogQuery()",
		"function invokeTool()",
		"function exportTools()",
		"function shortTraceID",
		"auditRows.querySelectorAll(\"tr[data-audit]\")",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected console html to contain %q", want)
		}
	}
}

func TestRoutingExplainHTTP(t *testing.T) {
	handler, _ := newTestHandler()
	res := postJSON(handler, "/gateway/v1/routing/explain", "", map[string]any{
		"app_id":      "dev-app",
		"model":       "local-small",
		"data_labels": []string{"sensitive"},
	})
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Policy struct {
			AllowCloud bool `json:"allow_cloud"`
		} `json:"policy"`
		Candidates []struct {
			Provider config.Provider `json:"provider"`
		} `json:"candidates"`
		Skipped []struct {
			ProviderID string `json:"provider_id"`
			Reason     string `json:"reason"`
		} `json:"skipped"`
	}
	decodeBody(t, res, &body)
	if body.Policy.AllowCloud {
		t.Fatal("expected cloud blocked by policy")
	}
	if len(body.Candidates) != 1 || body.Candidates[0].Provider.ID != "local-mock" {
		t.Fatalf("expected local candidate, got %+v", body.Candidates)
	}
	if len(body.Skipped) == 0 || !strings.Contains(body.Skipped[0].Reason, "cloud") {
		t.Fatalf("expected skipped cloud reason, got %+v", body.Skipped)
	}
}

func TestPolicyDryRunMatchesModelRuleHTTP(t *testing.T) {
	handler, _ := newTestHandler()
	res := postJSON(handler, "/gateway/v1/policy/dry-run", "", map[string]any{
		"app_id":       "dev-app",
		"request_type": "chat",
		"model":        "cloud-smart",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Decision struct {
			RuleID  string `json:"rule_id"`
			Allowed bool   `json:"allowed"`
		} `json:"decision"`
	}
	decodeBody(t, res, &body)
	if body.Decision.Allowed || body.Decision.RuleID != "deny-cloud-smart" {
		t.Fatalf("expected deny-cloud-smart decision, got %+v", body.Decision)
	}
}

func TestAccessDryRunAllowsChatByAppID(t *testing.T) {
	handler, _ := newTestHandler()
	res := postJSON(handler, "/gateway/v1/access/dry-run", "", map[string]any{
		"app_id": "dev-app",
		"action": "chat",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Allowed      bool   `json:"allowed"`
		AppID        string `json:"app_id"`
		MatchedGrant string `json:"matched_grant"`
	}
	decodeBody(t, res, &body)
	if !body.Allowed || body.AppID != "dev-app" || body.MatchedGrant != "chat" {
		t.Fatalf("unexpected access dry-run body: %+v", body)
	}
}

func TestAccessDryRunExplainsMissingToolScope(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [
	    {"id":"dev-app","token":"dev-token","grants":["chat","tool"]},
	    {"id":"admin-app","token":"admin-token","grants":["admin"]}
	  ],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "tools": [{"id":"gateway.runtime_health","name":"Runtime Health","adapter":"runtime-health","read_only":true,"risk_level":"low","scopes":["runtime.read"],"enabled":true}]
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	auditStore := audit.NewMemoryStore()
	handler := NewRuntimeHandler(manager, store).WithAudit(auditStore).Routes()

	res := postJSON(handler, "/gateway/v1/access/dry-run", "", map[string]any{
		"token":   "admin-token",
		"action":  "tool.invoke",
		"tool_id": "gateway.runtime_health",
	})
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Allowed       bool     `json:"allowed"`
		AppID         string   `json:"app_id"`
		MissingGrants []string `json:"missing_grants"`
		Tool          struct {
			ID     string   `json:"id"`
			Scopes []string `json:"scopes"`
		} `json:"tool"`
	}
	decodeBody(t, res, &body)
	if body.Allowed || body.AppID != "admin-app" || body.Tool.ID != "gateway.runtime_health" {
		t.Fatalf("unexpected access dry-run denial: %+v", body)
	}
	if len(body.MissingGrants) != 1 || body.MissingGrants[0] != "tool:runtime.read" {
		t.Fatalf("expected missing runtime scope, got %+v", body.MissingGrants)
	}
	events := auditStore.List(audit.ListQuery{Action: "access.dry_run"})
	if len(events) != 1 || events[0].Metadata["allowed"] != false || events[0].Metadata["reason"] != "required grant is missing" {
		t.Fatalf("expected access dry-run audit metadata, got %+v", events)
	}
	if events[0].Metadata["tool_id"] != "gateway.runtime_health" || events[0].Metadata["adapter"] != "runtime-health" {
		t.Fatalf("expected tool audit metadata on dry-run, got %+v", events[0].Metadata)
	}
	if missing, ok := events[0].Metadata["missing_grants"].([]string); !ok || len(missing) != 1 || missing[0] != "tool:runtime.read" {
		t.Fatalf("expected missing grants audit metadata, got %+v", events[0].Metadata["missing_grants"])
	}
}

func TestAccessDryRunRejectsUnsupportedAction(t *testing.T) {
	handler, _ := newTestHandler()
	res := postJSON(handler, "/gateway/v1/access/dry-run", "", map[string]any{
		"app_id": "dev-app",
		"action": "provider.enabled",
	})
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", res.Code, res.Body.String())
	}
}

func TestConfigReloadHTTP(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [
	    {"id":"dev-app","token":"dev-token","grants":["chat"]},
	    {"id":"admin-app","token":"admin-token","grants":["admin"]}
	  ],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}]
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	auditStore := audit.NewMemoryStore()
	handler := NewRuntimeHandler(manager, store).WithAudit(auditStore).Routes()

	if err := os.WriteFile(path, []byte(`{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v2",
	  "apps": [
	    {"id":"dev-app","token":"dev-token","grants":["chat"]},
	    {"id":"admin-app","token":"admin-token","grants":["admin"]}
	  ],
	  "providers": [
	    {"id":"local-mock","class":"local","models":["local-small"],"healthy":true},
	    {"id":"cloud-mock","class":"cloud","models":["local-small"],"healthy":true}
	  ]
	}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/gateway/v1/config/reload", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without admin token, got %d", res.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/gateway/v1/config/reload", nil)
	req.Header.Set("Authorization", "Bearer dev-token")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with non-admin token, got %d", res.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/gateway/v1/config/reload", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body map[string]any
	decodeBody(t, res, &body)
	if body["status"] != "reloaded" || body["policy_version"] != "v2" {
		t.Fatalf("unexpected reload body: %+v", body)
	}

	req = httptest.NewRequest(http.MethodGet, "/gateway/v1/providers", nil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var providersBody struct {
		Providers []providerhealth.View `json:"providers"`
	}
	decodeBody(t, res, &providersBody)
	if len(providersBody.Providers) != 2 {
		t.Fatalf("expected reloaded providers, got %+v", providersBody.Providers)
	}

	events := auditStore.List(audit.ListQuery{Action: "config.reload"})
	if len(events) != 3 {
		t.Fatalf("expected denied, denied, success audit events, got %+v", events)
	}
	if events[0].Result != audit.ResultSuccess || events[0].AppID != "admin-app" {
		t.Fatalf("expected latest success event for admin app, got %+v", events[0])
	}
}

func TestProviderManagementHTTP(t *testing.T) {
	path := writeHandlerConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [
	    {"id":"dev-app","token":"dev-token","grants":["chat"]},
	    {"id":"admin-app","token":"admin-token","grants":["admin"]}
	  ],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}]
	}`)
	store := trace.NewMemoryStore()
	manager, err := gatewayruntime.NewManager(path, store)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	auditStore := audit.NewMemoryStore()
	handler := NewRuntimeHandler(manager, store).WithAudit(auditStore).Routes()

	res := postJSON(handler, "/gateway/v1/providers/local-mock/enabled", "dev-token", map[string]any{"enabled": false})
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for non-admin, got %d", res.Code)
	}

	res = postJSON(handler, "/gateway/v1/providers/local-mock/enabled", "admin-token", map[string]any{"enabled": false})
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/providers", nil)
	providersRes := httptest.NewRecorder()
	handler.ServeHTTP(providersRes, req)
	var providersBody struct {
		Providers []providerhealth.View `json:"providers"`
	}
	decodeBody(t, providersRes, &providersBody)
	if len(providersBody.Providers) != 1 || providersBody.Providers[0].Enabled {
		t.Fatalf("expected disabled provider view, got %+v", providersBody.Providers)
	}

	res = postJSON(handler, "/gateway/v1/providers/local-mock/probe", "admin-token", map[string]any{})
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	events := auditStore.List(audit.ListQuery{Action: "provider.enabled"})
	if len(events) == 0 || events[0].Result != audit.ResultSuccess {
		t.Fatalf("expected provider.enabled audit success, got %+v", events)
	}
}

func TestAuditListRequiresAdmin(t *testing.T) {
	handler, _ := newTestHandlerWithLogger(slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/audit/events", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestAuditListAdminHTTP(t *testing.T) {
	handler, _ := newTestHandlerWithLogger(slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))
	postJSON(handler, "/gateway/v1/policy/dry-run", "", map[string]any{
		"app_id":       "dev-app",
		"request_type": "chat",
		"model":        "local-small",
	})
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/audit/events?action=policy.dry_run", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Events []audit.Event `json:"events"`
	}
	decodeBody(t, res, &body)
	if len(body.Events) != 1 || body.Events[0].Action != "policy.dry_run" {
		t.Fatalf("expected dry-run audit event, got %+v", body.Events)
	}
}

func TestAuditListPaginationHTTP(t *testing.T) {
	handler, _ := newTestHandlerWithLogger(slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))
	for i := 0; i < 3; i++ {
		postJSON(handler, "/gateway/v1/policy/dry-run", "", map[string]any{
			"app_id":       "dev-app",
			"request_type": "chat",
			"model":        "local-small",
		})
	}
	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/audit/events?action=policy.dry_run&limit=1&offset=1", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Events []audit.Event `json:"events"`
		Total  int           `json:"total"`
		Offset int           `json:"offset"`
		Limit  int           `json:"limit"`
	}
	decodeBody(t, res, &body)
	if len(body.Events) != 1 || body.Total != 3 || body.Offset != 1 || body.Limit != 1 {
		t.Fatalf("unexpected paged audit body: %+v", body)
	}
}

func TestAuditExportRequiresAdminAndExportsJSONL(t *testing.T) {
	handler, _ := newTestHandlerWithLogger(slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))
	postJSON(handler, "/gateway/v1/policy/dry-run", "", map[string]any{
		"app_id":       "dev-app",
		"request_type": "chat",
		"model":        "local-small",
	})

	req := httptest.NewRequest(http.MethodGet, "/gateway/v1/audit/events/export?action=policy.dry_run", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/gateway/v1/audit/events/export?action=policy.dry_run", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "application/x-ndjson" {
		t.Fatalf("unexpected content type %q", got)
	}
	if !strings.Contains(res.Body.String(), `"action":"policy.dry_run"`) || !strings.Contains(res.Body.String(), `"result":"success"`) {
		t.Fatalf("unexpected audit export body: %s", res.Body.String())
	}
}

func newTestHandler() (http.Handler, *trace.MemoryStore) {
	return newTestHandlerWithLogger(slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))
}

func writeHandlerConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func newTestHandlerWithLogger(logger *slog.Logger) (http.Handler, *trace.MemoryStore) {
	cfg := config.Config{
		ListenAddr:     "127.0.0.1:0",
		TraceStorePath: "memory",
		PolicyVersion:  "test",
		Apps: []config.App{
			{ID: "dev-app", Token: "dev-token", Grants: []string{"chat", "tool"}},
			{ID: "admin-app", Token: "admin-token", Grants: []string{"admin"}},
		},
		Providers: []config.Provider{
			{ID: "local-mock", Class: "local", Models: []string{"local-small"}, Healthy: true},
			{ID: "cloud-mock", Class: "cloud", Models: []string{"local-small", "cloud-smart"}, Healthy: true},
		},
		Policies: []config.Policy{
			{ID: "deny-sensitive-cloud", Effect: "deny_cloud_for_sensitive", Reason: "Sensitive data cannot use cloud providers", DataLabels: []string{"sensitive"}},
			{ID: "deny-cloud-smart", Effect: "deny", Reason: "cloud-smart is disabled for this app", AppIDs: []string{"dev-app"}, Models: []string{"cloud-smart"}},
		},
		Tools: []config.Tool{{ID: "gateway.runtime_health", Name: "Runtime Health", Adapter: "runtime-health", ReadOnly: true, RiskLevel: "low", Scopes: []string{"runtime.read"}}},
	}
	health := providerhealth.NewStore(cfg.Providers)
	auditStore := audit.NewMemoryStore()
	registry := adapters.NewRegistry()
	registry.Register(adapters.NewMockProvider("local-mock"))
	registry.Register(adapters.NewMockProvider("cloud-mock"))
	store := trace.NewMemoryStore()
	pipeline := core.NewPipeline(core.Dependencies{
		Config:     cfg,
		Policy:     policy.NewEngine(cfg.PolicyVersion, cfg.Policies),
		Router:     router.NewWithHealth(cfg.Providers, health),
		Fallback:   fallback.NewManager(),
		Adapters:   registry,
		TraceStore: store,
	})
	return NewHandler(cfg, pipeline, store).WithProviderHealth(health).WithAudit(auditStore).WithLogger(logger).Routes(), store
}

func postJSON(handler http.Handler, path, token string, body any) *httptest.ResponseRecorder {
	encoded, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(encoded))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func decodeBody(t *testing.T, res *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		t.Fatalf("decode body: %v; body=%s", err, res.Body.String())
	}
}

func errorTraceID(t *testing.T, res *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error struct {
			TraceID string `json:"trace_id"`
		} `json:"error"`
	}
	decodeBody(t, res, &body)
	return body.Error.TraceID
}
