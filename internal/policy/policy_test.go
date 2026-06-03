package policy

import (
	"testing"

	"client-ai-gateway/internal/config"
)

func TestEngineDeniesCloudForSensitiveData(t *testing.T) {
	engine := NewEngine("v1", []config.Policy{{
		ID:         "sensitive-local",
		Effect:     "deny_cloud_for_sensitive",
		Reason:     "sensitive data cannot use cloud",
		DataLabels: []string{"sensitive"},
	}})

	decision := engine.Evaluate(Input{AppID: "dev-app", RequestType: "chat", Model: "local-small", DataLabels: []string{"sensitive"}})
	if !decision.Allowed || decision.AllowCloud || !decision.ForceLocal || decision.RuleID != "sensitive-local" {
		t.Fatalf("unexpected decision: %+v", decision)
	}
}

func TestEngineKeepsLegacySensitiveRuleDefault(t *testing.T) {
	engine := NewEngine("v1", []config.Policy{{
		ID:     "legacy-sensitive-local",
		Effect: "deny_cloud_for_sensitive",
		Reason: "sensitive data cannot use cloud",
	}})

	normal := engine.Evaluate(Input{RequestType: "chat", Model: "local-small"})
	if !normal.Allowed || !normal.AllowCloud || normal.RuleID != "default-allow" {
		t.Fatalf("unexpected normal decision: %+v", normal)
	}

	sensitive := engine.Evaluate(Input{RequestType: "chat", Model: "local-small", DataLabels: []string{"sensitive"}})
	if !sensitive.Allowed || sensitive.AllowCloud || sensitive.RuleID != "legacy-sensitive-local" {
		t.Fatalf("unexpected sensitive decision: %+v", sensitive)
	}
}

func TestEngineMatchesAppModelAndProviderClass(t *testing.T) {
	engine := NewEngine("v1", []config.Policy{{
		ID:              "desktop-local",
		Effect:          "force_local",
		Reason:          "desktop app must stay local",
		AppIDs:          []string{"desktop-app"},
		Models:          []string{"local-small"},
		ProviderClasses: []string{"cloud"},
	}})

	decision := engine.Evaluate(Input{
		AppID:         "desktop-app",
		RequestType:   "chat",
		Model:         "local-small",
		ProviderClass: "cloud",
	})
	if !decision.Allowed || decision.AllowCloud || !decision.ForceLocal || decision.RuleID != "desktop-local" {
		t.Fatalf("unexpected matching decision: %+v", decision)
	}

	defaultDecision := engine.Evaluate(Input{
		AppID:         "desktop-app",
		RequestType:   "chat",
		Model:         "cloud-smart",
		ProviderClass: "cloud",
	})
	if !defaultDecision.Allowed || !defaultDecision.AllowCloud || defaultDecision.RuleID != "default-allow" {
		t.Fatalf("unexpected default decision: %+v", defaultDecision)
	}
}

func TestEngineAppliesCloudBlockingBeforeProviderSelection(t *testing.T) {
	engine := NewEngine("v1", []config.Policy{{
		ID:              "local-large-local-only",
		Effect:          "force_local",
		Reason:          "local-large must stay local",
		Models:          []string{"local-large"},
		ProviderClasses: []string{"cloud"},
	}})

	decision := engine.Evaluate(Input{RequestType: "chat", Model: "local-large"})
	if !decision.Allowed || decision.AllowCloud || !decision.ForceLocal || decision.RuleID != "local-large-local-only" {
		t.Fatalf("unexpected decision before provider selection: %+v", decision)
	}
}

func TestEngineDenyEffect(t *testing.T) {
	engine := NewEngine("v1", []config.Policy{{
		ID:     "deny-cloud-smart",
		Effect: "deny",
		Reason: "model disabled",
		Models: []string{"cloud-smart"},
	}})

	decision := engine.Evaluate(Input{RequestType: "chat", Model: "cloud-smart"})
	if decision.Allowed || decision.AllowCloud || decision.ForceLocal || decision.RuleID != "deny-cloud-smart" {
		t.Fatalf("unexpected deny decision: %+v", decision)
	}
}

func TestEngineEvaluatesHigherPriorityRuleFirst(t *testing.T) {
	engine := NewEngine("v1", []config.Policy{
		{ID: "low-allow", Priority: 10, Effect: "allow", Reason: "low allow", Models: []string{"local-small"}},
		{ID: "high-deny", Priority: 100, Effect: "deny", Reason: "high deny", Models: []string{"local-small"}},
	})

	decision := engine.Evaluate(Input{RequestType: "chat", Model: "local-small"})
	if decision.Allowed || decision.RuleID != "high-deny" || decision.RulePriority != 100 {
		t.Fatalf("expected high priority deny, got %+v", decision)
	}
	if decision.ConditionSummary != "model in [local-small]" {
		t.Fatalf("unexpected condition summary %q", decision.ConditionSummary)
	}
}

func TestEngineKeepsConfigOrderForSamePriority(t *testing.T) {
	engine := NewEngine("v1", []config.Policy{
		{ID: "first", Priority: 10, Effect: "deny", Reason: "first", Models: []string{"local-small"}},
		{ID: "second", Priority: 10, Effect: "allow", Reason: "second", Models: []string{"local-small"}},
	})

	decision := engine.Evaluate(Input{RequestType: "chat", Model: "local-small"})
	if decision.RuleID != "first" {
		t.Fatalf("expected same priority to keep config order, got %+v", decision)
	}
}

func TestConditionSummaryUsesLegacySensitiveLabel(t *testing.T) {
	summary := ConditionSummary(config.Policy{ID: "sensitive", Effect: "deny_cloud_for_sensitive"})
	if summary != "label in [sensitive]" {
		t.Fatalf("unexpected summary %q", summary)
	}
}
