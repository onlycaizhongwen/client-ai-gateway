package tools

import (
	"context"
	"fmt"

	"client-ai-gateway/internal/config"
)

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
	Invoke(context.Context, Input) (Result, error)
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
			registry.Register(&RuntimeHealthTool{id: toolCfg.ID, health: runtimeHealth})
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

type RuntimeHealthTool struct {
	id     string
	health func() any
}

func (t *RuntimeHealthTool) ID() string {
	return t.id
}

func (t *RuntimeHealthTool) Invoke(_ context.Context, input Input) (Result, error) {
	if t.health == nil {
		return Result{}, fmt.Errorf("runtime health is unavailable")
	}
	return Result{
		ToolID: input.ToolID,
		Output: t.health(),
	}, nil
}
