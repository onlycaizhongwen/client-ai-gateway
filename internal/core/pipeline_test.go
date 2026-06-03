package core_test

import (
	"context"
	"errors"
	"testing"

	"client-ai-gateway/internal/adapters"
	"client-ai-gateway/internal/config"
	"client-ai-gateway/internal/core"
	"client-ai-gateway/internal/fallback"
	"client-ai-gateway/internal/policy"
	"client-ai-gateway/internal/router"
	"client-ai-gateway/internal/trace"
)

func TestPipelineChatSuccess(t *testing.T) {
	pipeline, _ := newTestPipeline()

	resp, err := pipeline.Chat(context.Background(), "dev-token", core.ChatRequest{
		Model:    "local-small",
		Messages: []core.Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if resp.TraceID == "" {
		t.Fatal("expected trace id")
	}
	if resp.Choices[0].Message.Content == "" {
		t.Fatal("expected mock content")
	}
}

func TestPipelineFallbackAllowed(t *testing.T) {
	pipeline, store := newTestPipeline()

	resp, err := pipeline.Chat(context.Background(), "dev-token", core.ChatRequest{
		Model:    "local-small",
		Messages: []core.Message{{Role: "user", Content: "hello"}},
		Metadata: map[string]string{"fail_provider": "local-mock"},
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	record, ok := store.Get(resp.TraceID)
	if !ok {
		t.Fatal("expected trace record")
	}
	if len(record.Fallbacks) == 0 {
		t.Fatal("expected fallback attempt")
	}
	if record.ProviderID != "cloud-mock" {
		t.Fatalf("expected cloud fallback, got %s", record.ProviderID)
	}
	if record.RequestType != "chat" || record.Request.Model != "local-small" {
		t.Fatalf("expected chat request snapshot, got %+v", record)
	}
	if len(record.Request.Messages) != 1 || record.Request.Messages[0].Content != "hello" {
		t.Fatalf("expected request messages in trace snapshot, got %+v", record.Request.Messages)
	}
	if record.Request.Metadata["fail_provider"] != "local-mock" {
		t.Fatalf("expected request metadata in trace snapshot, got %+v", record.Request.Metadata)
	}
}

func TestPipelineSensitiveBlocksCloudFallbackRedactsTraceSnapshot(t *testing.T) {
	pipeline, store := newTestPipeline()

	_, err := pipeline.Chat(context.Background(), "dev-token", core.ChatRequest{
		Model:      "local-small",
		Messages:   []core.Message{{Role: "user", Content: "secret"}},
		DataLabels: []string{"sensitive"},
		Metadata:   map[string]string{"fail_provider": "local-mock"},
	})
	if err == nil {
		t.Fatal("expected failure because cloud fallback is blocked")
	}
	var gatewayErr *core.GatewayError
	if !errors.As(err, &gatewayErr) {
		t.Fatal("expected gateway error")
	}
	if gatewayErr.TraceID == "" {
		t.Fatal("expected trace id on failure")
	}
	record, ok := store.Get(gatewayErr.TraceID)
	if !ok {
		t.Fatal("expected failed trace record")
	}
	if len(record.Request.Messages) != 1 || record.Request.Messages[0].Content != "[redacted]" {
		t.Fatalf("expected sensitive message to be redacted, got %+v", record.Request.Messages)
	}
	if record.Request.Metadata["fail_provider"] != "[redacted]" {
		t.Fatalf("expected sensitive metadata to be redacted, got %+v", record.Request.Metadata)
	}
	if len(record.Request.DataLabels) != 1 || record.Request.DataLabels[0] != "sensitive" {
		t.Fatalf("expected data labels to be preserved, got %+v", record.Request.DataLabels)
	}
}

func TestPipelineTraceSnapshotUsesConfiguredRedactLabels(t *testing.T) {
	pipeline, store := newTestPipelineWithConfig(config.Config{
		TraceRedactLabels: []string{"private"},
	})

	resp, err := pipeline.Chat(context.Background(), "dev-token", core.ChatRequest{
		Model:      "local-small",
		Messages:   []core.Message{{Role: "user", Content: "private message"}},
		DataLabels: []string{"private"},
		Metadata:   map[string]string{"ticket": "secret-123"},
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	record, ok := store.Get(resp.TraceID)
	if !ok {
		t.Fatal("expected trace record")
	}
	if record.Request.Messages[0].Content != "[redacted]" || record.Request.Metadata["ticket"] != "[redacted]" {
		t.Fatalf("expected configured label to redact snapshot, got %+v", record.Request)
	}
}

func TestPipelineTraceSnapshotMaxChars(t *testing.T) {
	pipeline, store := newTestPipelineWithConfig(config.Config{TraceSnapshotMaxChars: 4})

	resp, err := pipeline.Chat(context.Background(), "dev-token", core.ChatRequest{
		Model:    "local-small",
		Messages: []core.Message{{Role: "user", Content: "你好abcdef"}},
		Metadata: map[string]string{"long": "metadata-value"},
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	record, ok := store.Get(resp.TraceID)
	if !ok {
		t.Fatal("expected trace record")
	}
	if record.Request.Messages[0].Content != "你好ab...[truncated]" {
		t.Fatalf("expected message truncation, got %+v", record.Request.Messages)
	}
	if record.Request.Metadata["long"] != "meta...[truncated]" {
		t.Fatalf("expected metadata truncation, got %+v", record.Request.Metadata)
	}
}

func TestPipelineTraceSnapshotCanBeDisabled(t *testing.T) {
	enabled := false
	pipeline, store := newTestPipelineWithConfig(config.Config{TraceSnapshotEnabled: &enabled})

	resp, err := pipeline.Chat(context.Background(), "dev-token", core.ChatRequest{
		Model:      "local-small",
		Messages:   []core.Message{{Role: "user", Content: "hello"}},
		DataLabels: []string{"sensitive"},
		Metadata:   map[string]string{"k": "v"},
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	record, ok := store.Get(resp.TraceID)
	if !ok {
		t.Fatal("expected trace record")
	}
	if record.Request.Model != "" || len(record.Request.Messages) != 0 || len(record.Request.Metadata) != 0 || len(record.Request.DataLabels) != 0 {
		t.Fatalf("expected empty request snapshot, got %+v", record.Request)
	}
}

func newTestPipeline() (*core.Pipeline, *trace.MemoryStore) {
	return newTestPipelineWithConfig(config.Config{})
}

func newTestPipelineWithConfig(overrides config.Config) (*core.Pipeline, *trace.MemoryStore) {
	cfg := config.Config{
		ListenAddr:            "127.0.0.1:0",
		PolicyVersion:         "test",
		TraceSnapshotEnabled:  overrides.TraceSnapshotEnabled,
		TraceRedactLabels:     overrides.TraceRedactLabels,
		TraceSnapshotMaxChars: overrides.TraceSnapshotMaxChars,
		Apps:                  []config.App{{ID: "dev-app", Token: "dev-token", Grants: []string{"chat"}}},
		Providers: []config.Provider{
			{ID: "local-mock", Class: "local", Models: []string{"local-small"}, Healthy: true},
			{ID: "cloud-mock", Class: "cloud", Models: []string{"local-small", "cloud-smart"}, Healthy: true},
		},
		Policies: []config.Policy{{ID: "deny-sensitive-cloud", Effect: "deny_cloud_for_sensitive", Reason: "Sensitive data cannot use cloud providers"}},
	}
	registry := adapters.NewRegistry()
	registry.Register(adapters.NewMockProvider("local-mock"))
	registry.Register(adapters.NewMockProvider("cloud-mock"))
	store := trace.NewMemoryStore()
	return core.NewPipeline(core.Dependencies{
		Config:     cfg,
		Policy:     policy.NewEngine(cfg.PolicyVersion, cfg.Policies),
		Router:     router.New(cfg.Providers),
		Fallback:   fallback.NewManager(),
		Adapters:   registry,
		TraceStore: store,
	}), store
}
