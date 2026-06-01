package adapters

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"client-ai-gateway/internal/config"
)

type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: map[string]Provider{}}
}

func NewRegistryFromConfig(providers []config.Provider) (*Registry, error) {
	registry := NewRegistry()
	for i, provider := range providers {
		adapter, err := NewProvider(provider)
		if err != nil {
			return nil, fmt.Errorf("providers[%d] %s: %w", i, provider.ID, err)
		}
		registry.Register(adapter)
	}
	return registry, nil
}

func NewProvider(provider config.Provider) (Provider, error) {
	adapter := provider.Adapter
	if adapter == "" {
		adapter = "mock"
	}
	switch adapter {
	case "mock":
		return NewMockProvider(provider.ID), nil
	case "openai-compatible":
		apiKey := ""
		if provider.APIKeyEnv != "" {
			apiKey = os.Getenv(provider.APIKeyEnv)
		}
		return NewOpenAICompatibleProvider(OpenAICompatibleConfig{
			ID:         provider.ID,
			BaseURL:    provider.BaseURL,
			APIKey:     apiKey,
			APIKeyEnv:  provider.APIKeyEnv,
			HTTPClient: &http.Client{Timeout: 60 * time.Second},
		})
	default:
		return nil, fmt.Errorf("unsupported adapter %q", adapter)
	}
}

func (r *Registry) Register(provider Provider) {
	r.providers[provider.ID()] = provider
}

func (r *Registry) Get(id string) (Provider, bool) {
	provider, ok := r.providers[id]
	return provider, ok
}

func (r *Registry) All() []Provider {
	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}
