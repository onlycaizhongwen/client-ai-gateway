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
	mu     sync.Mutex
	clock  func() time.Time
	apps   map[string]appLimit
	bucket map[string]appWindow
}

type appLimit struct {
	requestsPerMinute int
}

type appWindow struct {
	start time.Time
	used  int
}

func NewLimiter(cfg config.Quotas) *Limiter {
	limiter := &Limiter{
		clock:  time.Now,
		apps:   map[string]appLimit{},
		bucket: map[string]appWindow{},
	}
	for _, app := range cfg.Apps {
		if app.RequestsPerMinute > 0 {
			limiter.apps[app.AppID] = appLimit{requestsPerMinute: app.RequestsPerMinute}
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
	l.mu.Lock()
	defer l.mu.Unlock()

	limit, ok := l.apps[appID]
	if !ok || limit.requestsPerMinute <= 0 {
		return Decision{Allowed: true, Subject: appID}
	}
	now := l.clock().UTC()
	window := l.bucket[appID]
	if window.start.IsZero() || now.Sub(window.start) >= time.Minute {
		window = appWindow{start: now}
	}
	if window.used >= limit.requestsPerMinute {
		return Decision{
			Allowed:   false,
			Subject:   appID,
			Limit:     limit.requestsPerMinute,
			Remaining: 0,
			ResetAt:   window.start.Add(time.Minute),
			Reason:    "app request rate limit exceeded",
		}
	}
	window.used++
	l.bucket[appID] = window
	return Decision{
		Allowed:   true,
		Subject:   appID,
		Limit:     limit.requestsPerMinute,
		Remaining: limit.requestsPerMinute - window.used,
		ResetAt:   window.start.Add(time.Minute),
	}
}
