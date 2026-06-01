package audit

import (
	"sync"
	"time"
)

const (
	ResultSuccess = "success"
	ResultDenied  = "denied"
	ResultFailed  = "failed"
)

type Event struct {
	ID         string         `json:"id"`
	TraceID    string         `json:"trace_id,omitempty"`
	AppID      string         `json:"app_id,omitempty"`
	Action     string         `json:"action"`
	Target     string         `json:"target,omitempty"`
	Result     string         `json:"result"`
	Error      string         `json:"error,omitempty"`
	DurationMS int64          `json:"duration_ms"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type ListQuery struct {
	Offset  int
	Limit   int
	Action  string
	Result  string
	AppID   string
	TraceID string
}

type Page struct {
	Items  []Event
	Total  int
	Offset int
	Limit  int
}

type Store interface {
	Save(Event)
	List(ListQuery) []Event
	Page(ListQuery) Page
}

type MemoryStore struct {
	mu     sync.RWMutex
	events []Event
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Save(event Event) {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.ID == "" {
		event.ID = "audit-" + event.CreatedAt.Format("20060102150405.000000000")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

func (s *MemoryStore) List(query ListQuery) []Event {
	return s.Page(query).Items
}

func (s *MemoryStore) Page(query ListQuery) Page {
	limit := query.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Event, 0, len(s.events))
	for i := len(s.events) - 1; i >= 0; i-- {
		event := s.events[i]
		if query.Action != "" && event.Action != query.Action {
			continue
		}
		if query.Result != "" && event.Result != query.Result {
			continue
		}
		if query.AppID != "" && event.AppID != query.AppID {
			continue
		}
		if query.TraceID != "" && event.TraceID != query.TraceID {
			continue
		}
		out = append(out, event)
	}
	total := len(out)
	if offset >= total {
		return Page{Items: []Event{}, Total: total, Offset: offset, Limit: limit}
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return Page{Items: out[offset:end], Total: total, Offset: offset, Limit: limit}
}
