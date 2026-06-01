package router

import (
	"testing"

	"client-ai-gateway/internal/config"
	"client-ai-gateway/internal/providerhealth"
)

func TestPlanSkipsDisabledProvider(t *testing.T) {
	disabled := false
	router := New([]config.Provider{
		{ID: "local-disabled", Class: "local", Models: []string{"local-small"}, Healthy: true, Enabled: &disabled},
		{ID: "cloud-enabled", Class: "cloud", Models: []string{"local-small"}, Healthy: true},
	})

	plan, err := router.Plan(Input{Model: "local-small", AllowCloud: true})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(plan.Candidates) != 1 {
		t.Fatalf("expected one candidate, got %d", len(plan.Candidates))
	}
	if plan.Candidates[0].Provider.ID != "cloud-enabled" {
		t.Fatalf("expected cloud-enabled, got %s", plan.Candidates[0].Provider.ID)
	}
}

func TestPlanFailsWhenOnlyProviderDisabled(t *testing.T) {
	disabled := false
	router := New([]config.Provider{
		{ID: "local-disabled", Class: "local", Models: []string{"local-small"}, Healthy: true, Enabled: &disabled},
	})

	_, err := router.Plan(Input{Model: "local-small", AllowCloud: true})
	if err == nil {
		t.Fatal("expected no route")
	}
}

func TestPlanSkipsRuntimeUnhealthyProvider(t *testing.T) {
	providers := []config.Provider{
		{ID: "local-runtime-unhealthy", Class: "local", Models: []string{"local-small"}, Healthy: true},
		{ID: "cloud-healthy", Class: "cloud", Models: []string{"local-small"}, Healthy: true},
	}
	health := providerhealth.NewStore(providers)
	health.Set("local-runtime-unhealthy", providerhealth.StatusUnhealthy, "probe failed")
	router := NewWithHealth(providers, health)

	plan, err := router.Plan(Input{Model: "local-small", AllowCloud: true})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(plan.Candidates) != 1 || plan.Candidates[0].Provider.ID != "cloud-healthy" {
		t.Fatalf("expected only cloud-healthy, got %+v", plan.Candidates)
	}
}

func TestExplainIncludesSkippedReasons(t *testing.T) {
	disabled := false
	router := New([]config.Provider{
		{ID: "local-disabled", Class: "local", Models: []string{"local-small"}, Healthy: true, Enabled: &disabled},
		{ID: "cloud-blocked", Class: "cloud", Models: []string{"local-small"}, Healthy: true},
		{ID: "local-mismatch", Class: "local", Models: []string{"other"}, Healthy: true},
		{ID: "local-ok", Class: "local", Models: []string{"local-small"}, Healthy: true},
	})

	explanation := router.Explain(Input{Model: "local-small", AllowCloud: false})
	if len(explanation.Candidates) != 1 || explanation.Candidates[0].Provider.ID != "local-ok" {
		t.Fatalf("unexpected candidates: %+v", explanation.Candidates)
	}
	if len(explanation.Skipped) != 3 {
		t.Fatalf("expected three skipped providers, got %+v", explanation.Skipped)
	}
	if explanation.Skipped[0].Reason == "" {
		t.Fatalf("expected skipped reason, got %+v", explanation.Skipped[0])
	}
}
