package router

import (
	"fmt"

	"client-ai-gateway/internal/config"
	"client-ai-gateway/internal/providerhealth"
)

type Input struct {
	Model      string
	AllowCloud bool
}

type Candidate struct {
	Provider config.Provider `json:"provider"`
	Model    string          `json:"model"`
	Reason   string          `json:"reason"`
}

type Plan struct {
	Candidates []Candidate
}

type Explanation struct {
	Candidates []Candidate
	Skipped    []SkippedProvider
}

type SkippedProvider struct {
	ProviderID string `json:"provider_id"`
	Class      string `json:"class"`
	Reason     string `json:"reason"`
}

type Router struct {
	providers []config.Provider
	health    *providerhealth.Store
}

func New(providers []config.Provider) *Router {
	return &Router{providers: providers}
}

func NewWithHealth(providers []config.Provider, health *providerhealth.Store) *Router {
	return &Router{providers: providers, health: health}
}

func (r *Router) Plan(input Input) (Plan, error) {
	explanation := r.Explain(input)
	if len(explanation.Candidates) == 0 {
		return Plan{}, fmt.Errorf("no enabled healthy provider supports model %q under current policy", input.Model)
	}
	return Plan{Candidates: explanation.Candidates}, nil
}

func (r *Router) Explain(input Input) Explanation {
	var candidates []Candidate
	var skipped []SkippedProvider
	for _, provider := range r.providers {
		if !provider.IsEnabled() {
			skipped = append(skipped, skip(provider, "provider disabled by config"))
			continue
		}
		if r.health != nil && !r.health.IsRoutable(provider) {
			skipped = append(skipped, skip(provider, "provider is not runtime routable"))
			continue
		}
		if r.health == nil && !provider.Healthy {
			skipped = append(skipped, skip(provider, "provider marked unhealthy in config"))
			continue
		}
		if provider.Class == "cloud" && !input.AllowCloud {
			skipped = append(skipped, skip(provider, "cloud provider blocked by policy"))
			continue
		}
		matched := false
		for _, model := range provider.Models {
			if input.Model == "" || input.Model == model {
				matched = true
				candidates = append(candidates, Candidate{
					Provider: provider,
					Model:    model,
					Reason:   fmt.Sprintf("provider %s supports requested model %s", provider.ID, model),
				})
			}
		}
		if !matched {
			skipped = append(skipped, skip(provider, fmt.Sprintf("provider does not support requested model %q", input.Model)))
		}
	}
	return Explanation{Candidates: candidates, Skipped: skipped}
}

func skip(provider config.Provider, reason string) SkippedProvider {
	return SkippedProvider{ProviderID: provider.ID, Class: provider.Class, Reason: reason}
}
