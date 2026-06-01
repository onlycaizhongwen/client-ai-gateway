package tools

import (
	"context"
	"errors"
	"testing"

	"client-ai-gateway/internal/config"
)

func TestRegistryExposesToolManifestFromConfig(t *testing.T) {
	registry, err := NewRegistryFromConfig([]config.Tool{{
		ID:          "gateway.runtime_health",
		Name:        "Runtime Health",
		Adapter:     "runtime-health",
		Description: "runtime snapshot",
		ReadOnly:    true,
		RiskLevel:   "low",
		Scopes:      []string{"runtime.read"},
		OutputSchema: map[string]any{
			"type": "object",
		},
	}}, func() any { return map[string]any{"status": "healthy"} })
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	tool, ok := registry.Get("gateway.runtime_health")
	if !ok {
		t.Fatal("expected runtime health tool")
	}
	manifest := tool.Manifest()
	if manifest.ID != "gateway.runtime_health" || manifest.Adapter != "runtime-health" || !manifest.ReadOnly || manifest.RiskLevel != "low" {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
	if len(manifest.Scopes) != 1 || manifest.Scopes[0] != "runtime.read" {
		t.Fatalf("unexpected scopes: %+v", manifest.Scopes)
	}
	if manifest.OutputSchema["type"] != "object" {
		t.Fatalf("unexpected output schema: %+v", manifest.OutputSchema)
	}
}

func TestRuntimeHealthToolUnavailableErrorContract(t *testing.T) {
	registry, err := NewRegistryFromConfig([]config.Tool{{
		ID:        "gateway.runtime_health",
		Adapter:   "runtime-health",
		ReadOnly:  true,
		RiskLevel: "low",
		Scopes:    []string{"runtime.read"},
	}}, nil)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	tool, ok := registry.Get("gateway.runtime_health")
	if !ok {
		t.Fatal("expected runtime health tool")
	}

	_, err = tool.Invoke(context.Background(), Input{ToolID: "gateway.runtime_health"})
	var toolErr *Error
	if !errors.As(err, &toolErr) || toolErr.Code != ErrCodeUnavailable {
		t.Fatalf("expected unavailable tool error, got %v", err)
	}
}
