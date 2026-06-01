package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	path := writeConfig(t, `{
	  "listen_addr": "127.0.0.1:18765",
	  "trace_store_path": "data/traces.jsonl",
	  "policy_version": "test",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "policies": [{"id":"deny-sensitive-cloud","effect":"deny_cloud_for_sensitive","reason":"no cloud"}]
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load valid config: %v", err)
	}
	if cfg.PolicyVersion != "test" {
		t.Fatalf("unexpected policy version %q", cfg.PolicyVersion)
	}
}

func TestLoadValidScopedToolGrant(t *testing.T) {
	path := writeConfig(t, `{
	  "listen_addr": "127.0.0.1:18765",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat","tool:runtime.read"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "tools": [{"id":"gateway.runtime_health","adapter":"runtime-health","read_only":true,"risk_level":"low","scopes":["runtime.read"]}]
	}`)

	if _, err := Load(path); err != nil {
		t.Fatalf("load valid scoped tool grant: %v", err)
	}
}

func TestLoadValidMCPRuntimePlaceholder(t *testing.T) {
	path := writeConfig(t, `{
	  "listen_addr": "127.0.0.1:18765",
	  "apps": [{"id":"dev-app","token":"dev-token","grants":["chat","tool:desktop.read"]}],
	  "providers": [{"id":"local-mock","class":"local","models":["local-small"],"healthy":true}],
	  "mcp_runtime": {
	    "enabled": true,
	    "mode": "manifest_only",
	    "servers": [{
	      "id": "desktop-context",
	      "tools": [{
	        "id":"mcp.desktop.list_context",
	        "read_only":true,
	        "risk_level":"low",
	        "scopes":["desktop.read"]
	      }]
	    }]
	  }
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load valid mcp runtime placeholder: %v", err)
	}
	if !cfg.MCPRuntime.Enabled || len(cfg.MCPRuntime.Servers) != 1 || len(cfg.MCPRuntime.Servers[0].Tools) != 1 {
		t.Fatalf("unexpected mcp runtime config: %+v", cfg.MCPRuntime)
	}
}

func TestProviderEnabledDefaultsToTrue(t *testing.T) {
	provider := Provider{ID: "p", Class: "local", Models: []string{"m"}, Healthy: true}
	if !provider.IsEnabled() {
		t.Fatal("expected omitted enabled to default true")
	}
	disabled := false
	provider.Enabled = &disabled
	if provider.IsEnabled() {
		t.Fatal("expected explicit false to disable provider")
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "duplicate app token",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"same","grants":["chat"]},{"id":"b","token":"same","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}]
			}`,
			want: "duplicate app token",
		},
		{
			name: "bad provider class",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"edge","models":["m"]}]
			}`,
			want: "class must be local or cloud",
		},
		{
			name: "negative trace retention",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "trace_retention_max": -1,
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}]
			}`,
			want: "trace_retention_max",
		},
		{
			name: "negative audit retention",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "audit_retention_max": -1,
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}]
			}`,
			want: "audit_retention_max",
		},
		{
			name: "openai compatible provider requires base url",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","adapter":"openai-compatible","models":["m"]}]
			}`,
			want: "base_url is required",
		},
		{
			name: "bad policy effect",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "policies": [{"id":"x","effect":"allow_all","reason":"bad"}]
			}`,
			want: "unsupported",
		},
		{
			name: "bad tool adapter",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "tools": [{"id":"tool.bad","adapter":"shell","read_only":true}]
			}`,
			want: "unsupported",
		},
		{
			name: "tool must be read only",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "tools": [{"id":"tool.write","adapter":"runtime-health","read_only":false}]
			}`,
			want: "read_only",
		},
		{
			name: "tool scopes required",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "tools": [{"id":"tool.no_scope","adapter":"runtime-health","read_only":true}]
			}`,
			want: "scopes must not be empty",
		},
		{
			name: "bad tool scope",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "tools": [{"id":"tool.bad_scope","adapter":"runtime-health","read_only":true,"scopes":["Runtime Read"]}]
			}`,
			want: "invalid scope",
		},
		{
			name: "bad tool risk level",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "tools": [{"id":"tool.bad_risk","adapter":"runtime-health","read_only":true,"risk_level":"critical","scopes":["runtime.read"]}]
			}`,
			want: "risk_level",
		},
		{
			name: "sandbox required unsupported",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "tools": [{"id":"tool.sandbox","adapter":"runtime-health","read_only":true,"scopes":["runtime.read"],"sandbox_required":true}]
			}`,
			want: "sandbox_required",
		},
		{
			name: "mcp execution mode unsupported before sandbox",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "mcp_runtime": {"enabled":true,"mode":"stdio","servers":[{"id":"s","tools":[]}]}
			}`,
			want: "mcp_runtime.mode",
		},
		{
			name: "mcp disabled mode cannot be enabled",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "mcp_runtime": {"enabled":true,"mode":"disabled","servers":[{"id":"s","tools":[]}]}
			}`,
			want: "disabled requires enabled=false",
		},
		{
			name: "mcp server id required",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "mcp_runtime": {"enabled":true,"servers":[{"tools":[]}]}
			}`,
			want: "mcp_runtime.servers[0].id",
		},
		{
			name: "mcp tool must be read only",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "mcp_runtime": {"enabled":true,"servers":[{"id":"s","tools":[{"id":"mcp.write","read_only":false,"scopes":["desktop.read"]}]}]}
			}`,
			want: "read_only",
		},
		{
			name: "mcp tool scopes required",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "mcp_runtime": {"enabled":true,"servers":[{"id":"s","tools":[{"id":"mcp.no_scope","read_only":true}]}]}
			}`,
			want: "scopes must not be empty",
		},
		{
			name: "mcp sandbox required unsupported",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "mcp_runtime": {"enabled":true,"servers":[{"id":"s","tools":[{"id":"mcp.sandbox","read_only":true,"scopes":["desktop.read"],"sandbox_required":true}]}]}
			}`,
			want: "sandbox_required",
		},
		{
			name: "bad policy provider class",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "policies": [{"id":"x","effect":"force_local","reason":"bad","provider_classes":["edge"]}]
			}`,
			want: "unsupported provider class",
		},
		{
			name: "bad policy request type",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["chat"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}],
			  "policies": [{"id":"x","effect":"deny","reason":"bad","request_types":["image"]}]
			}`,
			want: "unsupported request type",
		},
		{
			name: "bad grant",
			body: `{
			  "listen_addr": "127.0.0.1:18765",
			  "apps": [{"id":"a","token":"t","grants":["root"]}],
			  "providers": [{"id":"p","class":"local","models":["m"]}]
			}`,
			want: "unsupported grant",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(writeConfig(t, tt.body))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q in %q", tt.want, err.Error())
			}
		})
	}
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
