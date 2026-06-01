package policy

import "client-ai-gateway/internal/config"

type Input struct {
	AppID         string
	RequestType   string
	DataLabels    []string
	Model         string
	ProviderClass string
}

type Decision struct {
	RuleID      string `json:"rule_id"`
	Version     string `json:"version"`
	Allowed     bool   `json:"allowed"`
	AllowCloud  bool   `json:"allow_cloud"`
	ForceLocal  bool   `json:"force_local"`
	Explanation string `json:"explanation"`
}

type Engine struct {
	version string
	rules   []config.Policy
}

func NewEngine(version string, rules []config.Policy) *Engine {
	return &Engine{version: version, rules: rules}
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
		if !matchesRule(rule, input) {
			continue
		}
		decision.RuleID = rule.ID
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

func matchesRule(rule config.Policy, input Input) bool {
	if !matchAny(rule.AppIDs, input.AppID) {
		return false
	}
	if !matchAny(rule.RequestTypes, input.RequestType) {
		return false
	}
	if !matchAny(rule.Models, input.Model) {
		return false
	}
	if !matchesProviderClass(rule, input.ProviderClass) {
		return false
	}
	dataLabels := rule.DataLabels
	if len(dataLabels) == 0 && rule.Effect == "deny_cloud_for_sensitive" {
		dataLabels = []string{"sensitive"}
	}
	if len(dataLabels) == 0 {
		return true
	}
	for _, want := range dataLabels {
		if contains(input.DataLabels, want) {
			return true
		}
	}
	return false
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
