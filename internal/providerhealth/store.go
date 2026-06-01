package providerhealth

import (
	"sync"
	"time"

	"client-ai-gateway/internal/config"
)

const (
	StatusHealthy   = "healthy"
	StatusDegraded  = "degraded"
	StatusUnhealthy = "unhealthy"
	StatusDisabled  = "disabled"
)

type State struct {
	ProviderID    string    `json:"provider_id"`
	Status        string    `json:"status"`
	Reason        string    `json:"reason,omitempty"`
	LastCheckedAt time.Time `json:"last_checked_at"`
}

type View struct {
	ID                string    `json:"id"`
	Name              string    `json:"name,omitempty"`
	Class             string    `json:"class"`
	Adapter           string    `json:"adapter"`
	BaseURL           string    `json:"base_url,omitempty"`
	Models            []string  `json:"models"`
	Enabled           bool      `json:"enabled"`
	ConfiguredHealthy bool      `json:"configured_healthy"`
	RuntimeStatus     string    `json:"runtime_status"`
	DegradedReason    string    `json:"degraded_reason,omitempty"`
	LastCheckedAt     time.Time `json:"last_checked_at"`
}

type Store struct {
	mu     sync.RWMutex
	states map[string]State
}

func NewStore(providers []config.Provider) *Store {
	store := &Store{states: map[string]State{}}
	now := time.Now().UTC()
	for _, provider := range providers {
		store.states[provider.ID] = initialState(provider, now)
	}
	return store
}

func (s *Store) IsRoutable(provider config.Provider) bool {
	if !provider.IsEnabled() {
		return false
	}
	s.mu.RLock()
	state, ok := s.states[provider.ID]
	s.mu.RUnlock()
	if !ok {
		return provider.Healthy
	}
	return state.Status == StatusHealthy || state.Status == StatusDegraded
}

func (s *Store) Set(providerID, status, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[providerID] = State{
		ProviderID:    providerID,
		Status:        status,
		Reason:        reason,
		LastCheckedAt: time.Now().UTC(),
	}
}

func (s *Store) Views(providers []config.Provider) []View {
	s.mu.RLock()
	defer s.mu.RUnlock()

	views := make([]View, 0, len(providers))
	for _, provider := range providers {
		state, ok := s.states[provider.ID]
		if !ok {
			state = initialState(provider, time.Now().UTC())
		}
		adapter := provider.Adapter
		if adapter == "" {
			adapter = "mock"
		}
		views = append(views, View{
			ID:                provider.ID,
			Name:              provider.Name,
			Class:             provider.Class,
			Adapter:           adapter,
			BaseURL:           provider.BaseURL,
			Models:            provider.Models,
			Enabled:           provider.IsEnabled(),
			ConfiguredHealthy: provider.Healthy,
			RuntimeStatus:     state.Status,
			DegradedReason:    state.Reason,
			LastCheckedAt:     state.LastCheckedAt,
		})
	}
	return views
}

func initialState(provider config.Provider, at time.Time) State {
	state := State{
		ProviderID:    provider.ID,
		Status:        StatusHealthy,
		Reason:        "configured healthy",
		LastCheckedAt: at,
	}
	if !provider.IsEnabled() {
		state.Status = StatusDisabled
		state.Reason = "provider disabled by config"
		return state
	}
	if !provider.Healthy {
		state.Status = StatusUnhealthy
		state.Reason = "provider marked unhealthy in config"
	}
	return state
}
