package quota

import (
	"sync"
	"time"

	"client-ai-gateway/internal/config"
)

type Decision struct {
	Allowed   bool
	Subject   string
	Limit     int
	Remaining int
	ResetAt   time.Time
	Reason    string
}

type Limiter struct {
	mu        sync.Mutex
	clock     func() time.Time
	apps      map[string]limit
	providers map[string]limit
	buckets   map[string]window
}

type limit struct {
	requestsPerMinute int
}

type window struct {
	start time.Time
	used  int
}

func NewLimiter(cfg config.Quotas) *Limiter {
	limiter := &Limiter{
		clock:     time.Now,
		apps:      map[string]limit{},
		providers: map[string]limit{},
		buckets:   map[string]window{},
	}
	for _, app := range cfg.Apps {
		if app.RequestsPerMinute > 0 {
			limiter.apps[app.AppID] = limit{requestsPerMinute: app.RequestsPerMinute}
		}
	}
	for _, provider := range cfg.Providers {
		if provider.RequestsPerMinute > 0 {
			limiter.providers[provider.ProviderID] = limit{requestsPerMinute: provider.RequestsPerMinute}
		}
	}
	return limiter
}

func NewLimiterWithClock(cfg config.Quotas, clock func() time.Time) *Limiter {
	limiter := NewLimiter(cfg)
	if clock != nil {
		limiter.clock = clock
	}
	return limiter
}

func (l *Limiter) AllowAppRequest(appID string) Decision {
	if l == nil {
		return Decision{Allowed: true}
	}
	return l.allow("app", appID, l.apps, "app request rate limit exceeded")
}

func (l *Limiter) AllowProviderRequest(providerID string) Decision {
	if l == nil {
		return Decision{Allowed: true}
	}
	return l.allow("provider", providerID, l.providers, "provider request rate limit exceeded")
}

func (l *Limiter) allow(subjectType, subjectID string, limits map[string]limit, reason string) Decision {
	l.mu.Lock()
	defer l.mu.Unlock()

	quotaLimit, ok := limits[subjectID]
	if !ok || quotaLimit.requestsPerMinute <= 0 {
		return Decision{Allowed: true, Subject: subjectID}
	}
	now := l.clock().UTC()
	key := subjectType + ":" + subjectID
	current := l.buckets[key]
	if current.start.IsZero() || now.Sub(current.start) >= time.Minute {
		current = window{start: now}
	}
	if current.used >= quotaLimit.requestsPerMinute {
		return Decision{
			Allowed:   false,
			Subject:   subjectID,
			Limit:     quotaLimit.requestsPerMinute,
			Remaining: 0,
			ResetAt:   current.start.Add(time.Minute),
			Reason:    reason,
		}
	}
	current.used++
	l.buckets[key] = current
	return Decision{
		Allowed:   true,
		Subject:   subjectID,
		Limit:     quotaLimit.requestsPerMinute,
		Remaining: quotaLimit.requestsPerMinute - current.used,
		ResetAt:   current.start.Add(time.Minute),
	}
}
