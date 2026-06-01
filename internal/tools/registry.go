package tools

import (
	"context"
	"fmt"

	"client-ai-gateway/internal/config"
)

const (
	ErrCodeUnavailable = "tool_unavailable"
	ErrCodeFailed      = "tool_failed"
)

type Manifest struct {
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
}

type Input struct {
	AppID     string         `json:"app_id"`
	ToolID    string         `json:"tool_id"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type Result struct {
	ToolID     string `json:"tool_id"`
	TraceID    string `json:"trace_id,omitempty"`
	AppID      string `json:"app_id,omitempty"`
	DurationMS int64  `json:"duration_ms"`
	Output     any    `json:"output"`
}

type Tool interface {
	ID() string
	Manifest() Manifest
	Invoke(context.Context, Input) (Result, error)
}

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewError(code, message string, err error) *Error {
	if code == "" {
		code = ErrCodeFailed
	}
	return &Error{Code: code, Message: message, Err: err}
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistryFromConfig(cfg []config.Tool, runtimeHealth func() any) (*Registry, error) {
	registry := &Registry{tools: map[string]Tool{}}
	for i, toolCfg := range cfg {
		if !toolCfg.IsEnabled() {
			continue
		}
		switch toolCfg.Adapter {
		case "runtime-health":
			registry.Register(&RuntimeHealthTool{manifest: ManifestFromConfig(toolCfg), health: runtimeHealth})
		default:
			return nil, fmt.Errorf("tools[%d] %s: unsupported adapter %q", i, toolCfg.ID, toolCfg.Adapter)
		}
	}
	return registry, nil
}

func (r *Registry) Register(tool Tool) {
	r.tools[tool.ID()] = tool
}

func (r *Registry) Get(id string) (Tool, bool) {
	tool, ok := r.tools[id]
	return tool, ok
}

func ManifestFromConfig(tool config.Tool) Manifest {
	return Manifest{
		ID:              tool.ID,
		Name:            tool.Name,
		Adapter:         tool.Adapter,
		Description:     tool.Description,
		ReadOnly:        tool.ReadOnly,
		RiskLevel:       tool.RiskLevel,
		Scopes:          append([]string(nil), tool.Scopes...),
		InputSchema:     tool.InputSchema,
		OutputSchema:    tool.OutputSchema,
		SandboxRequired: tool.SandboxRequired,
	}
}

type RuntimeHealthTool struct {
	manifest Manifest
	health   func() any
}

func (t *RuntimeHealthTool) ID() string {
	return t.manifest.ID
}

func (t *RuntimeHealthTool) Manifest() Manifest {
	return t.manifest
}

func (t *RuntimeHealthTool) Invoke(_ context.Context, input Input) (Result, error) {
	if t.health == nil {
		return Result{}, NewError(ErrCodeUnavailable, "runtime health is unavailable", nil)
	}
	return Result{
		ToolID: input.ToolID,
		Output: t.health(),
	}, nil
}
