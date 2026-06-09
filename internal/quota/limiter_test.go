package quota

import (
	"testing"
	"time"

	"client-ai-gateway/internal/config"
)

func TestLimiterAllowsWithinWindowAndRejectsOverflow(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	limiter := NewLimiterWithClock(config.Quotas{Apps: []config.AppQuota{{
		AppID:             "app",
		RequestsPerMinute: 2,
	}}}, func() time.Time { return now })

	if decision := limiter.AllowAppRequest("app"); !decision.Allowed || decision.Remaining != 1 {
		t.Fatalf("expected first request allowed, got %+v", decision)
	}
	if decision := limiter.AllowAppRequest("app"); !decision.Allowed || decision.Remaining != 0 {
		t.Fatalf("expected second request allowed, got %+v", decision)
	}
	if decision := limiter.AllowAppRequest("app"); decision.Allowed || decision.Limit != 2 || decision.Reason == "" {
		t.Fatalf("expected third request rejected, got %+v", decision)
	}
}

func TestLimiterResetsAfterWindow(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	limiter := NewLimiterWithClock(config.Quotas{Apps: []config.AppQuota{{
		AppID:             "app",
		RequestsPerMinute: 1,
	}}}, func() time.Time { return now })

	if decision := limiter.AllowAppRequest("app"); !decision.Allowed {
		t.Fatalf("expected initial request allowed, got %+v", decision)
	}
	if decision := limiter.AllowAppRequest("app"); decision.Allowed {
		t.Fatalf("expected same-window request rejected, got %+v", decision)
	}
	now = now.Add(time.Minute)
	if decision := limiter.AllowAppRequest("app"); !decision.Allowed || decision.Remaining != 0 {
		t.Fatalf("expected request after reset allowed, got %+v", decision)
	}
}

func TestLimiterIgnoresUnconfiguredApps(t *testing.T) {
	limiter := NewLimiter(config.Quotas{})
	if decision := limiter.AllowAppRequest("app"); !decision.Allowed {
		t.Fatalf("expected unconfigured app allowed, got %+v", decision)
	}
}

func TestLimiterAllowsProviderWithinWindowAndRejectsOverflow(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	limiter := NewLimiterWithClock(config.Quotas{Providers: []config.ProviderQuota{{
		ProviderID:        "provider",
		RequestsPerMinute: 1,
	}}}, func() time.Time { return now })

	if decision := limiter.AllowProviderRequest("provider"); !decision.Allowed || decision.Remaining != 0 {
		t.Fatalf("expected first provider request allowed, got %+v", decision)
	}
	if decision := limiter.AllowProviderRequest("provider"); decision.Allowed || decision.Limit != 1 || decision.Reason != "provider request rate limit exceeded" {
		t.Fatalf("expected second provider request rejected, got %+v", decision)
	}
	now = now.Add(time.Minute)
	if decision := limiter.AllowProviderRequest("provider"); !decision.Allowed || decision.Remaining != 0 {
		t.Fatalf("expected provider request after reset allowed, got %+v", decision)
	}
}

func TestLimiterKeepsAppAndProviderBucketsSeparate(t *testing.T) {
	limiter := NewLimiter(config.Quotas{
		Apps:      []config.AppQuota{{AppID: "same", RequestsPerMinute: 1}},
		Providers: []config.ProviderQuota{{ProviderID: "same", RequestsPerMinute: 1}},
	})

	if decision := limiter.AllowAppRequest("same"); !decision.Allowed {
		t.Fatalf("expected app request allowed, got %+v", decision)
	}
	if decision := limiter.AllowProviderRequest("same"); !decision.Allowed {
		t.Fatalf("expected provider request to use separate bucket, got %+v", decision)
	}
}
