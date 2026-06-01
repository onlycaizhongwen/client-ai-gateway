package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type OpenAICompatibleConfig struct {
	ID         string
	BaseURL    string
	APIKey     string
	APIKeyEnv  string
	HTTPClient *http.Client
}

type OpenAICompatibleProvider struct {
	id        string
	baseURL   string
	apiKey    string
	apiKeyEnv string
	client    *http.Client
}

func NewOpenAICompatibleProvider(cfg OpenAICompatibleConfig) (*OpenAICompatibleProvider, error) {
	if strings.TrimSpace(cfg.ID) == "" {
		return nil, fmt.Errorf("provider id is required")
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, fmt.Errorf("base url is required")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return &OpenAICompatibleProvider{
		id:        cfg.ID,
		baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:    cfg.APIKey,
		apiKeyEnv: strings.TrimSpace(cfg.APIKeyEnv),
		client:    client,
	}, nil
}

func (p *OpenAICompatibleProvider) ID() string {
	return p.id
}

func (p *OpenAICompatibleProvider) Chat(ctx context.Context, input ChatInput) (Result, error) {
	if err := p.requireCredential(); err != nil {
		return Result{}, err
	}
	body := openAIChatRequest{
		Model:    input.Model,
		Messages: input.Messages,
		Stream:   false,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return Result{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	if input.TraceID != "" {
		req.Header.Set("X-Trace-ID", input.TraceID)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return Result{}, p.classifyRequestError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, p.statusError(resp.StatusCode, "chat completion failed")
	}

	var decoded openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return Result{}, p.providerError(ErrorInvalidResponse, "upstream returned invalid chat completion response", 0, err)
	}
	if len(decoded.Choices) == 0 {
		return Result{}, p.providerError(ErrorEmptyResponse, "upstream returned no choices", 0, nil)
	}
	model := decoded.Model
	if model == "" {
		model = input.Model
	}
	return Result{
		ProviderID: p.id,
		Model:      model,
		Content:    decoded.Choices[0].Message.Content,
		Usage: Usage{
			PromptTokens:     decoded.Usage.PromptTokens,
			CompletionTokens: decoded.Usage.CompletionTokens,
			TotalTokens:      decoded.Usage.TotalTokens,
		},
	}, nil
}

func (p *OpenAICompatibleProvider) CheckHealth(ctx context.Context) error {
	if err := p.requireCredential(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/healthz", nil)
	if err != nil {
		return err
	}
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return p.classifyRequestError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return p.statusError(resp.StatusCode, "health check failed")
	}
	return nil
}

func (p *OpenAICompatibleProvider) requireCredential() error {
	if p.apiKeyEnv != "" && strings.TrimSpace(p.apiKey) == "" {
		return p.providerError(ErrorMissingCredential, "api key environment variable "+p.apiKeyEnv+" is not set", 0, nil)
	}
	return nil
}

func (p *OpenAICompatibleProvider) classifyRequestError(err error) error {
	code := ErrorConnectionFailed
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() {
		code = ErrorTimeout
	}
	return p.providerError(code, "upstream request failed", 0, err)
}

func (p *OpenAICompatibleProvider) statusError(statusCode int, action string) error {
	code := ErrorUpstreamStatus
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		code = ErrorUnauthorized
	case http.StatusTooManyRequests:
		code = ErrorRateLimited
	}
	return p.providerError(code, action, statusCode, nil)
}

func (p *OpenAICompatibleProvider) providerError(code, message string, statusCode int, err error) error {
	return &ProviderError{
		ProviderID: p.id,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

type openAIChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type openAIChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
}
