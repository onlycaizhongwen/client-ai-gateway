package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"client-ai-gateway/internal/adapters"
	"client-ai-gateway/internal/config"
	"client-ai-gateway/internal/fallback"
	"client-ai-gateway/internal/policy"
	"client-ai-gateway/internal/quota"
	"client-ai-gateway/internal/router"
	"client-ai-gateway/internal/trace"
)

var ErrUnauthorized = errors.New("unauthorized")

type Dependencies struct {
	Config     config.Config
	Policy     *policy.Engine
	Router     *router.Router
	Fallback   *fallback.Manager
	Adapters   *adapters.Registry
	Quota      *quota.Limiter
	TraceStore trace.Store
}

type Pipeline struct {
	deps Dependencies
}

func NewPipeline(deps Dependencies) *Pipeline {
	return &Pipeline{deps: deps}
}

func (p *Pipeline) Chat(ctx context.Context, token string, req ChatRequest) (ChatResponse, error) {
	app, ok := p.deps.Config.AppByToken(token)
	traceID := trace.NewID()
	env := RequestEnvelope{
		TraceID:    traceID,
		Model:      req.Model,
		Messages:   req.Messages,
		DataLabels: req.DataLabels,
		Metadata:   req.Metadata,
		StartedAt:  time.Now().UTC(),
	}
	if ok {
		env.AppID = app.ID
	}

	record := trace.Record{
		TraceID:     traceID,
		RequestType: "chat",
		AppID:       env.AppID,
		Model:       req.Model,
		Request:     p.traceRequestSnapshot(req),
		Status:      "started",
		StartedAt:   env.StartedAt,
		Events:      []trace.Event{{Type: "request_started", Message: "request accepted by access layer", At: env.StartedAt}},
	}

	if !ok {
		record.Status = "failed"
		record.Error = "unknown app token"
		record.FinishedAt = time.Now().UTC()
		record.Events = append(record.Events, trace.Event{Type: "auth_failed", Message: "unknown app token", At: record.FinishedAt})
		_ = p.deps.TraceStore.Save(record)
		return ChatResponse{}, gatewayError(traceID, env.AppID, "unauthorized", "unknown app token", ErrUnauthorized)
	}
	if !hasGrant(app.Grants, "chat") {
		record.Status = "failed"
		record.Error = "app lacks chat grant"
		record.FinishedAt = time.Now().UTC()
		record.Events = append(record.Events, trace.Event{Type: "auth_failed", Message: "app lacks chat grant", At: record.FinishedAt})
		_ = p.deps.TraceStore.Save(record)
		return ChatResponse{}, gatewayError(traceID, env.AppID, "unauthorized", "app lacks chat grant", ErrUnauthorized)
	}

	decision := p.deps.Policy.Evaluate(policy.Input{
		AppID:       app.ID,
		DataLabels:  env.DataLabels,
		RequestType: "chat",
		Model:       env.Model,
	})
	record.Policy = trace.PolicyDecision{
		RuleID:      decision.RuleID,
		Version:     decision.Version,
		Allowed:     decision.Allowed,
		AllowCloud:  decision.AllowCloud,
		Explanation: decision.Explanation,
	}
	record.Events = append(record.Events, trace.Event{Type: "policy_evaluated", Message: decision.Explanation, At: time.Now().UTC()})
	if !decision.Allowed {
		record.Status = "failed"
		record.Error = decision.Explanation
		record.FinishedAt = time.Now().UTC()
		_ = p.deps.TraceStore.Save(record)
		return ChatResponse{}, gatewayError(traceID, env.AppID, "policy_denied", "policy denied: "+decision.Explanation, nil)
	}
	if p.deps.Quota != nil {
		quotaDecision := p.deps.Quota.AllowAppRequest(app.ID)
		if !quotaDecision.Allowed {
			record.Status = "failed"
			record.Error = quotaDecision.Reason
			record.FinishedAt = time.Now().UTC()
			record.Events = append(record.Events, trace.Event{Type: "quota_rejected", Message: quotaDecision.Reason, At: record.FinishedAt})
			_ = p.deps.TraceStore.Save(record)
			return ChatResponse{}, gatewayError(traceID, env.AppID, "rate_limited", quotaDecision.Reason, nil)
		}
		if quotaDecision.Limit > 0 {
			record.Events = append(record.Events, trace.Event{Type: "quota_checked", Message: fmt.Sprintf("app rpm remaining %d/%d", quotaDecision.Remaining, quotaDecision.Limit), At: time.Now().UTC()})
		}
	}

	routePlan, err := p.deps.Router.Plan(router.Input{
		Model:      env.Model,
		AllowCloud: decision.AllowCloud,
	})
	if err != nil {
		record.Status = "failed"
		record.Error = err.Error()
		record.FinishedAt = time.Now().UTC()
		record.Events = append(record.Events, trace.Event{Type: "route_failed", Message: err.Error(), At: record.FinishedAt})
		_ = p.deps.TraceStore.Save(record)
		return ChatResponse{}, gatewayError(traceID, env.AppID, "route_failed", err.Error(), err)
	}

	var lastErr error
	var result adapters.Result
	for idx, candidate := range routePlan.Candidates {
		record.Routes = append(record.Routes, trace.RouteAttempt{
			ProviderID: candidate.Provider.ID,
			Model:      candidate.Model,
			Reason:     candidate.Reason,
			At:         time.Now().UTC(),
		})
		record.Events = append(record.Events, trace.Event{Type: "route_selected", Message: candidate.Reason, At: time.Now().UTC()})

		adapter, ok := p.deps.Adapters.Get(candidate.Provider.ID)
		if !ok {
			lastErr = fmt.Errorf("adapter not found for provider %s", candidate.Provider.ID)
		} else {
			result, lastErr = adapter.Chat(ctx, adapters.ChatInput{
				TraceID:  traceID,
				Model:    candidate.Model,
				Messages: toAdapterMessages(req.Messages),
				FailMode: req.Metadata["fail_provider"],
			})
		}

		if lastErr == nil {
			record.Status = "completed"
			record.ProviderID = result.ProviderID
			record.FinalModel = result.Model
			record.FinishedAt = time.Now().UTC()
			record.Events = append(record.Events, trace.Event{Type: "request_completed", Message: "provider returned response", At: record.FinishedAt})
			_ = p.deps.TraceStore.Save(record)
			return toOpenAIResponse(traceID, env.AppID, result), nil
		}

		action := p.deps.Fallback.Decide(fallback.Input{
			Attempt: idx,
			Total:   len(routePlan.Candidates),
			Err:     lastErr,
		})
		record.Fallbacks = append(record.Fallbacks, trace.FallbackAttempt{
			FromProviderID: candidate.Provider.ID,
			Reason:         lastErr.Error(),
			Action:         action.Action,
			At:             time.Now().UTC(),
		})
		record.Events = append(record.Events, trace.Event{Type: "fallback_decision", Message: action.Explanation, At: time.Now().UTC()})
		if action.Action == fallback.ActionFail {
			break
		}
	}

	record.Status = "failed"
	record.Error = lastErr.Error()
	record.FinishedAt = time.Now().UTC()
	_ = p.deps.TraceStore.Save(record)
	return ChatResponse{}, gatewayError(traceID, env.AppID, providerErrorCode(lastErr), lastErr.Error(), lastErr)
}

func gatewayError(traceID, appID, code, message string, err error) *GatewayError {
	if err == nil {
		err = errors.New(message)
	}
	return &GatewayError{TraceID: traceID, AppID: appID, Code: code, Message: message, Err: err}
}

func providerErrorCode(err error) string {
	var providerErr *adapters.ProviderError
	if errors.As(err, &providerErr) && providerErr.Code != "" {
		return "provider_" + providerErr.Code
	}
	return "provider_failed"
}

func hasGrant(grants []string, want string) bool {
	for _, grant := range grants {
		if grant == want {
			return true
		}
	}
	return false
}

func toOpenAIResponse(traceID, appID string, result adapters.Result) ChatResponse {
	created := time.Now().Unix()
	return ChatResponse{
		ID:      "chatcmpl-" + strings.ReplaceAll(traceID, "trace-", ""),
		Object:  "chat.completion",
		Created: created,
		Model:   result.Model,
		TraceID: traceID,
		AppID:   appID,
		Choices: []Choice{{
			Index: 0,
			Message: Message{
				Role:    "assistant",
				Content: result.Content,
			},
			FinishReason: "stop",
		}},
		Usage: Usage{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		},
	}
}

func toAdapterMessages(messages []Message) []adapters.Message {
	out := make([]adapters.Message, 0, len(messages))
	for _, message := range messages {
		out = append(out, adapters.Message{Role: message.Role, Content: message.Content})
	}
	return out
}

func (p *Pipeline) traceRequestSnapshot(req ChatRequest) trace.RequestSnapshot {
	if !p.deps.Config.IsTraceSnapshotEnabled() {
		return trace.RequestSnapshot{}
	}
	redact := shouldRedactTraceSnapshot(req.DataLabels, p.deps.Config.EffectiveTraceRedactLabels())
	maxChars := p.deps.Config.TraceSnapshotMaxChars
	return trace.RequestSnapshot{
		Model:      req.Model,
		Messages:   toTraceMessages(req.Messages, redact, maxChars),
		Metadata:   traceMetadata(req.Metadata, redact, maxChars),
		DataLabels: append([]string(nil), req.DataLabels...),
	}
}

func toTraceMessages(messages []Message, redact bool, maxChars int) []trace.MessageSnapshot {
	out := make([]trace.MessageSnapshot, 0, len(messages))
	for _, message := range messages {
		content := traceSnapshotValue(message.Content, redact, maxChars)
		out = append(out, trace.MessageSnapshot{Role: message.Role, Content: content})
	}
	return out
}

func traceMetadata(values map[string]string, redact bool, maxChars int) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = traceSnapshotValue(value, redact, maxChars)
	}
	return out
}

func traceSnapshotValue(value string, redact bool, maxChars int) string {
	if redact {
		return "[redacted]"
	}
	return truncateTraceSnapshotValue(value, maxChars)
}

func truncateTraceSnapshotValue(value string, maxChars int) string {
	if maxChars <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= maxChars {
		return value
	}
	return string(runes[:maxChars]) + "...[truncated]"
}

func shouldRedactTraceSnapshot(dataLabels []string, redactLabels []string) bool {
	for _, label := range dataLabels {
		for _, redactLabel := range redactLabels {
			if strings.EqualFold(strings.TrimSpace(label), strings.TrimSpace(redactLabel)) {
				return true
			}
		}
	}
	return false
}
