package policy

import (
	"strings"

	"client-ai-gateway/internal/config"
)

type Input struct {
	AppID         string
	RequestType   string
	DataLabels    []string
	Model         string
	ProviderClass string
}

type Decision struct {
	RuleID           string           `json:"rule_id"`
	RulePriority     int              `json:"rule_priority"`
	Version          string           `json:"version"`
	Allowed          bool             `json:"allowed"`
	AllowCloud       bool             `json:"allow_cloud"`
	ForceLocal       bool             `json:"force_local"`
	ConditionSummary string           `json:"condition_summary,omitempty"`
	Explanation      string           `json:"explanation"`
	RuleEvaluations  []RuleEvaluation `json:"rule_evaluations,omitempty"`
}

type RuleEvaluation struct {
	RuleID           string   `json:"rule_id"`
	EvaluationOrder  int      `json:"evaluation_order"`
	Priority         int      `json:"priority"`
	ConditionSummary string   `json:"condition_summary"`
	Matched          bool     `json:"matched"`
	MismatchFields   []string `json:"mismatch_fields,omitempty"`
}

type Engine struct {
	version string
	rules   []config.Policy
}

func NewEngine(version string, rules []config.Policy) *Engine {
	return &Engine{version: version, rules: orderedRules(rules)}
}

func (e *Engine) Evaluate(input Input) Decision {
	decision := Decision{
		RuleID:      "default-allow",
		Version:     e.version,
		Allowed:     true,
		AllowCloud:  true,
		Explanation: "Allowed by default local development policy",
	}
	for _, rule := range e.rules {
		mismatchFields := ruleMismatchFields(rule, input)
		matched := len(mismatchFields) == 0
		decision.RuleEvaluations = append(decision.RuleEvaluations, RuleEvaluation{
			RuleID:           rule.ID,
			EvaluationOrder:  len(decision.RuleEvaluations) + 1,
			Priority:         rule.Priority,
			ConditionSummary: ConditionSummary(rule),
			Matched:          matched,
			MismatchFields:   mismatchFields,
		})
		if !matched {
			continue
		}
		decision.RuleID = rule.ID
		decision.RulePriority = rule.Priority
		decision.ConditionSummary = ConditionSummary(rule)
		decision.Explanation = rule.Reason
		switch rule.Effect {
		case "allow":
			decision.Allowed = true
			decision.AllowCloud = true
			decision.ForceLocal = false
		case "deny":
			decision.Allowed = false
			decision.AllowCloud = false
			decision.ForceLocal = false
		case "deny_cloud_for_sensitive", "force_local":
			decision.Allowed = true
			decision.AllowCloud = false
			decision.ForceLocal = true
		}
		return decision
	}
	return decision
}

func ConditionSummary(rule config.Policy) string {
	parts := []string{
		conditionPart("app", rule.AppIDs),
		conditionPart("request", rule.RequestTypes),
		conditionPart("model", rule.Models),
		conditionPart("provider", rule.ProviderClasses),
		conditionPart("label", effectiveDataLabels(rule)),
	}
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return "any"
	}
	return strings.Join(out, " && ")
}

func conditionPart(name string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	return name + " in [" + strings.Join(values, ",") + "]"
}

func effectiveDataLabels(rule config.Policy) []string {
	if len(rule.DataLabels) == 0 && rule.Effect == "deny_cloud_for_sensitive" {
		return []string{"sensitive"}
	}
	return rule.DataLabels
}

func orderedRules(rules []config.Policy) []config.Policy {
	return OrderedRules(rules)
}

func OrderedRules(rules []config.Policy) []config.Policy {
	out := append([]config.Policy(nil), rules...)
	for i := 1; i < len(out); i++ {
		current := out[i]
		j := i - 1
		for j >= 0 && out[j].Priority < current.Priority {
			out[j+1] = out[j]
			j--
		}
		out[j+1] = current
	}
	return out
}

func matchesRule(rule config.Policy, input Input) bool {
	return len(ruleMismatchFields(rule, input)) == 0
}

func ruleMismatchFields(rule config.Policy, input Input) []string {
	fields := []string{}
	if !matchAny(rule.AppIDs, input.AppID) {
		fields = append(fields, "app")
	}
	if !matchAny(rule.RequestTypes, input.RequestType) {
		fields = append(fields, "request")
	}
	if !matchAny(rule.Models, input.Model) {
		fields = append(fields, "model")
	}
	if !matchesProviderClass(rule, input.ProviderClass) {
		fields = append(fields, "provider")
	}
	dataLabels := effectiveDataLabels(rule)
	if len(dataLabels) == 0 {
		return fields
	}
	for _, want := range dataLabels {
		if contains(input.DataLabels, want) {
			return fields
		}
	}
	return append(fields, "label")
}

func matchAny(patterns []string, value string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		if pattern == "*" || pattern == value {
			return true
		}
	}
	return false
}

func matchesProviderClass(rule config.Policy, value string) bool {
	if len(rule.ProviderClasses) == 0 {
		return true
	}
	if value != "" {
		return matchAny(rule.ProviderClasses, value)
	}
	switch rule.Effect {
	case "force_local", "deny_cloud_for_sensitive":
		return contains(rule.ProviderClasses, "cloud") || contains(rule.ProviderClasses, "*")
	default:
		return false
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
