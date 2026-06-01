package providerhealth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"client-ai-gateway/internal/adapters"
	"client-ai-gateway/internal/config"
)

func TestMonitorMarksUnhealthyAfterConsecutiveFailures(t *testing.T) {
	provider := config.Provider{ID: "p", Class: "local", Models: []string{"m"}, Healthy: true}
	registry := adapters.NewRegistry()
	registry.Register(&probeProvider{id: "p", err: fmt.Errorf("down")})
	store := NewStore([]config.Provider{provider})
	monitor := NewMonitor([]config.Provider{provider}, registry, store, MonitorConfig{
		Timeout:        time.Second,
		UnhealthyAfter: 2,
	})

	monitor.CheckOnce(context.Background())
	if got := store.Views([]config.Provider{provider})[0].RuntimeStatus; got != StatusDegraded {
		t.Fatalf("expected degraded after first failure, got %s", got)
	}
	monitor.CheckOnce(context.Background())
	if got := store.Views([]config.Provider{provider})[0].RuntimeStatus; got != StatusUnhealthy {
		t.Fatalf("expected unhealthy after second failure, got %s", got)
	}
}

func TestMonitorRestoresHealthyAfterSuccess(t *testing.T) {
	provider := config.Provider{ID: "p", Class: "local", Models: []string{"m"}, Healthy: true}
	probe := &probeProvider{id: "p", err: fmt.Errorf("down")}
	registry := adapters.NewRegistry()
	registry.Register(probe)
	store := NewStore([]config.Provider{provider})
	monitor := NewMonitor([]config.Provider{provider}, registry, store, MonitorConfig{UnhealthyAfter: 1})

	monitor.CheckOnce(context.Background())
	if got := store.Views([]config.Provider{provider})[0].RuntimeStatus; got != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", got)
	}
	probe.err = nil
	monitor.CheckOnce(context.Background())
	view := store.Views([]config.Provider{provider})[0]
	if view.RuntimeStatus != StatusHealthy {
		t.Fatalf("expected healthy recovery, got %+v", view)
	}
}

type probeProvider struct {
	id  string
	err error
}

func (p *probeProvider) ID() string {
	return p.id
}

func (p *probeProvider) Chat(context.Context, adapters.ChatInput) (adapters.Result, error) {
	return adapters.Result{}, nil
}

func (p *probeProvider) CheckHealth(context.Context) error {
	return p.err
}
