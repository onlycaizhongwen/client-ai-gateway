package core

import "time"

type ChatRequest struct {
	Model      string            `json:"model"`
	Messages   []Message         `json:"messages"`
	Stream     bool              `json:"stream,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	DataLabels []string          `json:"data_labels,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestEnvelope struct {
	TraceID    string
	AppID      string
	Model      string
	Messages   []Message
	DataLabels []string
	Metadata   map[string]string
	StartedAt  time.Time
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	TraceID string   `json:"trace_id"`
	AppID   string   `json:"-"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ProviderResult struct {
	ProviderID string
	Model      string
	Content    string
	Usage      Usage
}

type GatewayError struct {
	TraceID string
	AppID   string
	Code    string
	Message string
	Err     error
}

func (e *GatewayError) Error() string {
	return e.Message
}

func (e *GatewayError) Unwrap() error {
	return e.Err
}
