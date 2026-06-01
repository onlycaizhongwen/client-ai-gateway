package providerhealth

import (
	"testing"

	"client-ai-gateway/internal/config"
)

func TestStoreInitialViews(t *testing.T) {
	disabled := false
	store := NewStore([]config.Provider{
		{ID: "ok", Class: "local", Adapter: "mock", Models: []string{"m"}, Healthy: true},
		{ID: "off", Class: "cloud", Adapter: "mock", Models: []string{"m"}, Healthy: true, Enabled: &disabled},
		{ID: "bad", Class: "local", Adapter: "mock", Models: []string{"m"}, Healthy: false},
	})

	views := store.Views([]config.Provider{
		{ID: "ok", Class: "local", Adapter: "mock", Models: []string{"m"}, Healthy: true},
		{ID: "off", Class: "cloud", Adapter: "mock", Models: []string{"m"}, Healthy: true, Enabled: &disabled},
		{ID: "bad", Class: "local", Adapter: "mock", Models: []string{"m"}, Healthy: false},
	})
	if views[0].RuntimeStatus != StatusHealthy {
		t.Fatalf("expected healthy, got %+v", views[0])
	}
	if views[1].RuntimeStatus != StatusDisabled || views[1].Enabled {
		t.Fatalf("expected disabled view, got %+v", views[1])
	}
	if views[2].RuntimeStatus != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %+v", views[2])
	}
}

func TestStoreRoutableStatus(t *testing.T) {
	provider := config.Provider{ID: "p", Class: "local", Models: []string{"m"}, Healthy: true}
	store := NewStore([]config.Provider{provider})
	if !store.IsRoutable(provider) {
		t.Fatal("expected healthy provider routable")
	}
	store.Set("p", StatusDegraded, "slow responses")
	if !store.IsRoutable(provider) {
		t.Fatal("expected degraded provider routable")
	}
	store.Set("p", StatusUnhealthy, "probe failed")
	if store.IsRoutable(provider) {
		t.Fatal("expected unhealthy provider not routable")
	}
}
