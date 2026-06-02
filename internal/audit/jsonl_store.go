package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type JSONLStore struct {
	mu       sync.RWMutex
	path     string
	maxItems int
	events   []Event
}

func NewJSONLStore(path string) (*JSONLStore, error) {
	return NewJSONLStoreWithRetention(path, 0)
}

func NewJSONLStoreWithRetention(path string, maxItems int) (*JSONLStore, error) {
	if maxItems < 0 {
		return nil, fmt.Errorf("max items must be >= 0")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	store := &JSONLStore{path: path, maxItems: maxItems}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *JSONLStore) Save(event Event) {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.ID == "" {
		event.ID = "audit-" + event.CreatedAt.Format("20060102150405.000000000")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, event)
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	_ = json.NewEncoder(file).Encode(event)
	_ = s.enforceRetentionLocked()
}

func (s *JSONLStore) List(query ListQuery) []Event {
	return s.Page(query).Items
}

func (s *JSONLStore) Page(query ListQuery) Page {
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
		if !matchesQuery(event, query) {
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

func (s *JSONLStore) load() error {
	file, err := os.Open(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		s.events = append(s.events, event)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return s.enforceRetentionLocked()
}

func (s *JSONLStore) enforceRetentionLocked() error {
	if s.maxItems <= 0 || len(s.events) <= s.maxItems {
		return nil
	}
	s.events = append([]Event(nil), s.events[len(s.events)-s.maxItems:]...)
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, event := range s.events {
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}
