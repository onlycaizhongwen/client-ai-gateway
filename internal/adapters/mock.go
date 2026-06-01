package adapters

import (
	"context"
	"fmt"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Result struct {
	ProviderID string
	Model      string
	Content    string
	Usage      Usage
}

type ChatInput struct {
	TraceID  string
	Model    string
	Messages []Message
	FailMode string
}

type Provider interface {
	ID() string
	Chat(context.Context, ChatInput) (Result, error)
}

type HealthChecker interface {
	CheckHealth(context.Context) error
}

type MockProvider struct {
	id string
}

func NewMockProvider(id string) *MockProvider {
	return &MockProvider{id: id}
}

func (m *MockProvider) ID() string {
	return m.id
}

func (m *MockProvider) Chat(_ context.Context, input ChatInput) (Result, error) {
	if input.FailMode == m.id || input.FailMode == "all" {
		return Result{}, fmt.Errorf("mock provider %s unavailable", m.id)
	}
	prompt := lastUserMessage(input.Messages)
	if prompt == "" {
		prompt = "empty prompt"
	}
	content := fmt.Sprintf("mock response from %s using %s: %s", m.id, input.Model, prompt)
	return Result{
		ProviderID: m.id,
		Model:      input.Model,
		Content:    content,
		Usage: Usage{
			PromptTokens:     len(strings.Fields(prompt)),
			CompletionTokens: len(strings.Fields(content)),
			TotalTokens:      len(strings.Fields(prompt)) + len(strings.Fields(content)),
		},
	}, nil
}

func (m *MockProvider) CheckHealth(context.Context) error {
	return nil
}

func lastUserMessage(messages []Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}
