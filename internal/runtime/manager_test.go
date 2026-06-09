package runtime

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"client-ai-gateway/internal/config"
	"client-ai-gateway/internal/providerhealth"
	"client-ai-gateway/internal/trace"
)

func TestManagerReloadSwapsSnapshot(t *testing.T) {
	path := writeRuntimeConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [{"id":"local","class":"local","models":["m"],"healthy":true}]
	}`)
	manager, err := NewManager(path, trace.NewMemoryStore())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()
	if got := len(manager.Snapshot().Config.Providers); got != 1 {
		t.Fatalf("expected one provider, got %d", got)
	}

	if err := os.WriteFile(path, []byte(`{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v2",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [
	    {"id":"local","class":"local","models":["m"],"healthy":true},
	    {"id":"cloud","class":"cloud","models":["m"],"healthy":true}
	  ]
	}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := manager.Reload(); err != nil {
		t.Fatalf("reload: %v", err)
	}
	snapshot := manager.Snapshot()
	if snapshot.Config.PolicyVersion != "v2" || len(snapshot.Config.Providers) != 2 {
		t.Fatalf("unexpected snapshot: %+v", snapshot.Config)
	}
}

func TestManagerReloadKeepsSnapshotOnInvalidConfig(t *testing.T) {
	path := writeRuntimeConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [{"id":"local","class":"local","models":["m"],"healthy":true}]
	}`)
	manager, err := NewManager(path, trace.NewMemoryStore())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()

	if err := os.WriteFile(path, []byte(`{"listen_addr":"127.0.0.1:0"}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := manager.Reload(); err == nil {
		t.Fatal("expected reload failure")
	}
	if got := len(manager.Snapshot().Config.Providers); got != 1 {
		t.Fatalf("expected old snapshot to remain, got %d providers", got)
	}
}

func TestManagerSetProviderEnabled(t *testing.T) {
	path := writeRuntimeConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [{"id":"local","class":"local","models":["m"],"healthy":true}]
	}`)
	manager, err := NewManager(path, trace.NewMemoryStore())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()

	if err := manager.SetProviderEnabled("local", false); err != nil {
		t.Fatalf("set enabled: %v", err)
	}
	views := manager.Snapshot().Health.Views(manager.Snapshot().Config.Providers)
	if len(views) != 1 || views[0].Enabled {
		t.Fatalf("expected disabled provider after reload, got %+v", views)
	}
	if views[0].RuntimeStatus != providerhealth.StatusDisabled {
		t.Fatalf("expected disabled runtime status, got %+v", views[0])
	}
}

func TestManagerSetProviderRPMQuota(t *testing.T) {
	path := writeRuntimeConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [{"id":"local","class":"local","models":["m"],"healthy":true}]
	}`)
	manager, err := NewManager(path, trace.NewMemoryStore())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()

	change, err := manager.SetProviderRPMQuota("local", 9)
	if err != nil {
		t.Fatalf("set provider rpm quota: %v", err)
	}
	if change.OldRequestsPerMinute != 0 || change.RequestsPerMinute != 9 {
		t.Fatalf("expected provider quota change 0 -> 9, got %+v", change)
	}
	quotas := manager.Snapshot().Config.Quotas.Providers
	if len(quotas) != 1 || quotas[0].ProviderID != "local" || quotas[0].RequestsPerMinute != 9 {
		t.Fatalf("expected provider quota after reload, got %+v", quotas)
	}

	change, err = manager.SetProviderRPMQuota("local", 3)
	if err != nil {
		t.Fatalf("update provider rpm quota: %v", err)
	}
	if change.OldRequestsPerMinute != 9 || change.RequestsPerMinute != 3 {
		t.Fatalf("expected provider quota change 9 -> 3, got %+v", change)
	}
	quotas = manager.Snapshot().Config.Quotas.Providers
	if len(quotas) != 1 || quotas[0].RequestsPerMinute != 3 {
		t.Fatalf("expected updated provider quota, got %+v", quotas)
	}

	change, err = manager.SetProviderRPMQuota("local", 0)
	if err != nil {
		t.Fatalf("disable provider rpm quota: %v", err)
	}
	if change.OldRequestsPerMinute != 3 || change.RequestsPerMinute != 0 {
		t.Fatalf("expected provider quota change 3 -> 0, got %+v", change)
	}
	if got := manager.Snapshot().Config.Quotas.Providers; len(got) != 0 {
		t.Fatalf("expected provider quota to be removed, got %+v", got)
	}

	if _, err := manager.SetProviderRPMQuota("local", -1); err == nil {
		t.Fatal("expected negative provider rpm quota to fail")
	}
}

func TestManagerSetAppRPMQuota(t *testing.T) {
	path := writeRuntimeConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [{"id":"local","class":"local","models":["m"],"healthy":true}]
	}`)
	manager, err := NewManager(path, trace.NewMemoryStore())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()

	change, err := manager.SetAppRPMQuota("app", 11)
	if err != nil {
		t.Fatalf("set app rpm quota: %v", err)
	}
	if change.OldRequestsPerMinute != 0 || change.RequestsPerMinute != 11 {
		t.Fatalf("expected app quota change 0 -> 11, got %+v", change)
	}
	quotas := manager.Snapshot().Config.Quotas.Apps
	if len(quotas) != 1 || quotas[0].AppID != "app" || quotas[0].RequestsPerMinute != 11 {
		t.Fatalf("expected app quota after reload, got %+v", quotas)
	}

	change, err = manager.SetAppRPMQuota("app", 4)
	if err != nil {
		t.Fatalf("update app rpm quota: %v", err)
	}
	if change.OldRequestsPerMinute != 11 || change.RequestsPerMinute != 4 {
		t.Fatalf("expected app quota change 11 -> 4, got %+v", change)
	}
	quotas = manager.Snapshot().Config.Quotas.Apps
	if len(quotas) != 1 || quotas[0].RequestsPerMinute != 4 {
		t.Fatalf("expected updated app quota, got %+v", quotas)
	}

	change, err = manager.SetAppRPMQuota("app", 0)
	if err != nil {
		t.Fatalf("disable app rpm quota: %v", err)
	}
	if change.OldRequestsPerMinute != 4 || change.RequestsPerMinute != 0 {
		t.Fatalf("expected app quota change 4 -> 0, got %+v", change)
	}
	if got := manager.Snapshot().Config.Quotas.Apps; len(got) != 0 {
		t.Fatalf("expected app quota to be removed, got %+v", got)
	}

	if _, err := manager.SetAppRPMQuota("app", -1); err == nil {
		t.Fatal("expected negative app rpm quota to fail")
	}
}

func TestManagerConfigUpdatesAreSerialized(t *testing.T) {
	path := writeRuntimeConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [{"id":"local","class":"local","models":["m"],"healthy":true}]
	}`)
	manager, err := NewManager(path, trace.NewMemoryStore())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := manager.SetAppRPMQuota("app", 13)
		errs <- err
	}()
	go func() {
		defer wg.Done()
		errs <- manager.SetProviderEnabled("local", false)
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent config update: %v", err)
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.Quotas.Apps) != 1 || cfg.Quotas.Apps[0].AppID != "app" || cfg.Quotas.Apps[0].RequestsPerMinute != 13 {
		t.Fatalf("expected app quota to survive concurrent update, got %+v", cfg.Quotas.Apps)
	}
	if len(cfg.Providers) != 1 || cfg.Providers[0].Enabled == nil || *cfg.Providers[0].Enabled {
		t.Fatalf("expected provider enabled update to survive concurrent update, got %+v", cfg.Providers)
	}
}

func TestManagerHealthView(t *testing.T) {
	path := writeRuntimeConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [{"id":"local","class":"local","models":["m"],"healthy":true}],
	  "quotas": {
	    "apps": [{"app_id":"app","requests_per_minute":2}],
	    "providers": [{"provider_id":"local","requests_per_minute":3}]
	  }
	}`)
	manager, err := NewManager(path, trace.NewMemoryStore())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()

	health := manager.Health()
	if health.Status != "healthy" || health.Daemon.Status != "healthy" {
		t.Fatalf("unexpected runtime health: %+v", health)
	}
	if health.Config.ProviderCount != 1 || health.Config.ReloadCount == 0 {
		t.Fatalf("expected config metadata, got %+v", health.Config)
	}
	if health.ProviderMonitor.Status != "running" || health.ProviderMonitor.Total != 1 {
		t.Fatalf("expected running provider monitor, got %+v", health.ProviderMonitor)
	}
	if health.QuotaRuntime.Status != "configured" || health.QuotaRuntime.AppQuotaCount != 1 || health.QuotaRuntime.EnabledAppRPM != 1 || health.QuotaRuntime.TotalAppRPM != 2 {
		t.Fatalf("expected quota runtime metadata, got %+v", health.QuotaRuntime)
	}
	if health.QuotaRuntime.ProviderQuotaCount != 1 || health.QuotaRuntime.EnabledProviderRPM != 1 || health.QuotaRuntime.TotalProviderRPM != 3 {
		t.Fatalf("expected quota runtime metadata, got %+v", health.QuotaRuntime)
	}
	if health.ModelRuntime.Status != "not_configured" || health.MCPRuntime.Status != "not_configured" {
		t.Fatalf("expected placeholder runtime statuses, got %+v", health)
	}
}

func TestManagerHealthViewWithoutQuotaRuntime(t *testing.T) {
	path := writeRuntimeConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [{"id":"local","class":"local","models":["m"],"healthy":true}]
	}`)
	manager, err := NewManager(path, trace.NewMemoryStore())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()

	health := manager.Health()
	if health.QuotaRuntime.Status != "not_configured" || health.QuotaRuntime.AppQuotaCount != 0 {
		t.Fatalf("expected empty quota runtime metadata, got %+v", health.QuotaRuntime)
	}
}

func TestManagerHealthViewWithMCPRuntimePlaceholder(t *testing.T) {
	path := writeRuntimeConfig(t, `{
	  "listen_addr": "127.0.0.1:0",
	  "trace_store_path": "memory",
	  "audit_store_path": "memory",
	  "policy_version": "v1",
	  "apps": [{"id":"app","token":"token","grants":["chat"]}],
	  "providers": [{"id":"local","class":"local","models":["m"],"healthy":true}],
	  "mcp_runtime": {
	    "enabled": true,
	    "mode": "manifest_only",
	    "servers": [{
	      "id": "desktop-context",
	      "enabled": true,
	      "tools": [{"id":"mcp.desktop.list_context","read_only":true,"risk_level":"low","scopes":["desktop.read"],"enabled":true}]
	    }]
	  }
	}`)
	manager, err := NewManager(path, trace.NewMemoryStore())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer manager.Close()

	health := manager.Health()
	if health.MCPRuntime.Status != "configured" {
		t.Fatalf("expected configured mcp runtime, got %+v", health.MCPRuntime)
	}
	if health.MCPRuntime.Mode != "manifest_only" {
		t.Fatalf("expected manifest_only mode, got %+v", health.MCPRuntime)
	}
	if health.MCPRuntime.ServerCount != 1 || health.MCPRuntime.EnabledServers != 1 || health.MCPRuntime.ToolCount != 1 || health.MCPRuntime.EnabledTools != 1 {
		t.Fatalf("unexpected mcp runtime counts: %+v", health.MCPRuntime)
	}
}

func writeRuntimeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
