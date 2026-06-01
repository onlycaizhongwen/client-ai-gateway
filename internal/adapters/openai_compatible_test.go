package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAICompatibleProviderChat(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotTrace string
	var gotModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotTrace = r.Header.Get("X-Trace-ID")
		var req openAIChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotModel = req.Model
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "model": "served-model",
		  "choices": [{"message": {"role": "assistant", "content": "hello from upstream"}}],
		  "usage": {"prompt_tokens": 2, "completion_tokens": 3, "total_tokens": 5}
		}`))
	}))
	defer server.Close()

	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		ID:         "openai-like",
		BaseURL:    server.URL + "/",
		APIKey:     "test-key",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	result, err := provider.Chat(context.Background(), ChatInput{
		TraceID: "trace-123",
		Model:   "local-small",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("unexpected path %q", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("unexpected auth %q", gotAuth)
	}
	if gotTrace != "trace-123" {
		t.Fatalf("unexpected trace header %q", gotTrace)
	}
	if gotModel != "local-small" {
		t.Fatalf("unexpected request model %q", gotModel)
	}
	if result.ProviderID != "openai-like" || result.Model != "served-model" || result.Content != "hello from upstream" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Usage.TotalTokens != 5 {
		t.Fatalf("unexpected usage: %+v", result.Usage)
	}
}

func TestOpenAICompatibleProviderHealth(t *testing.T) {
	var gotPath string
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		ID:         "openai-like",
		BaseURL:    server.URL,
		APIKey:     "test-key",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	if err := provider.CheckHealth(context.Background()); err != nil {
		t.Fatalf("health: %v", err)
	}
	if gotPath != "/healthz" {
		t.Fatalf("unexpected path %q", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("unexpected auth %q", gotAuth)
	}
}

func TestOpenAICompatibleProviderClassifiesStatusErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantCode   string
	}{
		{name: "unauthorized", statusCode: http.StatusUnauthorized, wantCode: ErrorUnauthorized},
		{name: "forbidden", statusCode: http.StatusForbidden, wantCode: ErrorUnauthorized},
		{name: "rate limited", statusCode: http.StatusTooManyRequests, wantCode: ErrorRateLimited},
		{name: "upstream status", statusCode: http.StatusBadGateway, wantCode: ErrorUpstreamStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			provider := newTestOpenAIProvider(t, server.URL, server.Client())
			_, err := provider.Chat(context.Background(), ChatInput{
				Model:    "local-small",
				Messages: []Message{{Role: "user", Content: "hello"}},
			})
			assertProviderError(t, err, tt.wantCode, tt.statusCode)
		})
	}
}

func TestOpenAICompatibleProviderClassifiesInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{`))
	}))
	defer server.Close()

	provider := newTestOpenAIProvider(t, server.URL, server.Client())
	_, err := provider.Chat(context.Background(), ChatInput{
		Model:    "local-small",
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	assertProviderError(t, err, ErrorInvalidResponse, 0)
}

func TestOpenAICompatibleProviderClassifiesEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"served-model","choices":[]}`))
	}))
	defer server.Close()

	provider := newTestOpenAIProvider(t, server.URL, server.Client())
	_, err := provider.Chat(context.Background(), ChatInput{
		Model:    "local-small",
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	assertProviderError(t, err, ErrorEmptyResponse, 0)
}

func TestOpenAICompatibleProviderMissingCredential(t *testing.T) {
	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		ID:        "openai-like",
		BaseURL:   "http://127.0.0.1:1",
		APIKeyEnv: "LOCAL_OPENAI_API_KEY",
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	err = provider.CheckHealth(context.Background())
	assertProviderError(t, err, ErrorMissingCredential, 0)
}

func newTestOpenAIProvider(t *testing.T, baseURL string, client *http.Client) *OpenAICompatibleProvider {
	t.Helper()
	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		ID:         "openai-like",
		BaseURL:    baseURL,
		HTTPClient: client,
	})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	return provider
}

func assertProviderError(t *testing.T, err error, wantCode string, wantStatus int) {
	t.Helper()
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.Code != wantCode {
		t.Fatalf("expected code %q, got %q", wantCode, providerErr.Code)
	}
	if providerErr.StatusCode != wantStatus {
		t.Fatalf("expected status %d, got %d", wantStatus, providerErr.StatusCode)
	}
	if providerErr.ProviderID != "openai-like" {
		t.Fatalf("unexpected provider id %q", providerErr.ProviderID)
	}
}
