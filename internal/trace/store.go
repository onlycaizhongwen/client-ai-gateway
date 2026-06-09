package trace

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Record struct {
	TraceID     string            `json:"trace_id"`
	RequestType string            `json:"request_type,omitempty"`
	AppID       string            `json:"app_id"`
	Model       string            `json:"requested_model"`
	Request     RequestSnapshot   `json:"request,omitempty"`
	ToolID      string            `json:"tool_id,omitempty"`
	ProviderID  string            `json:"provider_id,omitempty"`
	FinalModel  string            `json:"final_model,omitempty"`
	Status      string            `json:"status"`
	Usage       *Usage            `json:"usage,omitempty"`
	Policy      PolicyDecision    `json:"policy"`
	Routes      []RouteAttempt    `json:"routes"`
	Fallbacks   []FallbackAttempt `json:"fallbacks"`
	Events      []Event           `json:"events"`
	Error       string            `json:"error,omitempty"`
	DurationMS  int64             `json:"duration_ms,omitempty"`
	StartedAt   time.Time         `json:"started_at"`
	FinishedAt  time.Time         `json:"finished_at,omitempty"`
}

type RequestSnapshot struct {
	Model      string            `json:"model,omitempty"`
	Messages   []MessageSnapshot `json:"messages,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	DataLabels []string          `json:"data_labels,omitempty"`
}

type MessageSnapshot struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	Source           string `json:"source"`
}

type PolicyDecision struct {
	RuleID      string `json:"rule_id"`
	Version     string `json:"version"`
	Allowed     bool   `json:"allowed"`
	AllowCloud  bool   `json:"allow_cloud"`
	Explanation string `json:"explanation"`
}

type RouteAttempt struct {
	ProviderID string    `json:"provider_id"`
	Model      string    `json:"model"`
	Reason     string    `json:"reason"`
	At         time.Time `json:"at"`
}

type FallbackAttempt struct {
	FromProviderID string    `json:"from_provider_id"`
	Reason         string    `json:"reason"`
	Action         string    `json:"action"`
	At             time.Time `json:"at"`
}

type Event struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	At      time.Time   `json:"at"`
	Quota   *QuotaEvent `json:"quota,omitempty"`
}

type QuotaEvent struct {
	Subject   string    `json:"subject"`
	ID        string    `json:"id"`
	Window    string    `json:"window,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Remaining int       `json:"remaining,omitempty"`
	ResetAt   time.Time `json:"reset_at,omitempty"`
}

type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]Record
}

type Store interface {
	Save(record Record) error
	Get(traceID string) (Record, bool)
	List(query ListQuery) []Record
	Page(query ListQuery) Page
	UsageSummary(query UsageSummaryQuery) UsageSummary
}

type ListQuery struct {
	Offset     int
	Limit      int
	Status     string
	AppID      string
	ProviderID string
	EventType  string
}

type Page struct {
	Items  []Record
	Total  int
	Offset int
	Limit  int
}

type UsageSummaryQuery struct {
	ListQuery
	GroupBy string
}

type UsageSummary struct {
	GroupBy          string             `json:"group_by"`
	Items            []UsageSummaryItem `json:"items"`
	TotalRecords     int                `json:"total_records"`
	UsageRecords     int                `json:"usage_records"`
	PromptTokens     int                `json:"prompt_tokens"`
	CompletionTokens int                `json:"completion_tokens"`
	TotalTokens      int                `json:"total_tokens"`
}

type UsageSummaryItem struct {
	Key              string `json:"key"`
	Records          int    `json:"records"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	UnknownUsage     int    `json:"unknown_usage"`
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: map[string]Record{}}
}

func (s *MemoryStore) Save(record Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[record.TraceID] = record
	return nil
}

func (s *MemoryStore) Get(traceID string) (Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[traceID]
	return record, ok
}

func (s *MemoryStore) List(query ListQuery) []Record {
	return s.Page(query).Items
}

func (s *MemoryStore) Page(query ListQuery) Page {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return pageRecords(s.records, query)
}

func (s *MemoryStore) UsageSummary(query UsageSummaryQuery) UsageSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return summarizeUsage(s.records, query)
}

type JSONLStore struct {
	mu        sync.RWMutex
	path      string
	maxItems  int
	records   map[string]Record
	writeSeq  int64
	sequences map[string]int64
}

func NewJSONLStore(path string) (*JSONLStore, error) {
	return NewJSONLStoreWithRetention(path, 0)
}

func NewJSONLStoreWithRetention(path string, maxItems int) (*JSONLStore, error) {
	if maxItems < 0 {
		return nil, fmt.Errorf("max items must be >= 0")
	}
	store := &JSONLStore{
		path:      path,
		maxItems:  maxItems,
		records:   map[string]Record{},
		sequences: map[string]int64{},
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *JSONLStore) Save(record Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeSeq++
	s.records[record.TraceID] = record
	s.sequences[record.TraceID] = s.writeSeq
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoded, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return err
	}
	return s.enforceRetentionLocked()
}

func (s *JSONLStore) Get(traceID string) (Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[traceID]
	return record, ok
}

func (s *JSONLStore) List(query ListQuery) []Record {
	return s.Page(query).Items
}

func (s *JSONLStore) Page(query ListQuery) Page {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return pageRecords(s.records, query)
}

func (s *JSONLStore) UsageSummary(query UsageSummaryQuery) UsageSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return summarizeUsage(s.records, query)
}

func (s *JSONLStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, line := range splitLines(data) {
		var record Record
		if err := json.Unmarshal(line, &record); err != nil {
			return err
		}
		s.writeSeq++
		s.records[record.TraceID] = record
		s.sequences[record.TraceID] = s.writeSeq
	}
	return s.enforceRetentionLocked()
}

func (s *JSONLStore) enforceRetentionLocked() error {
	if s.maxItems <= 0 || len(s.records) <= s.maxItems {
		return nil
	}
	records := sortedRecordsBySequence(s.records, s.sequences)
	removeCount := len(records) - s.maxItems
	for _, record := range records[:removeCount] {
		delete(s.records, record.TraceID)
		delete(s.sequences, record.TraceID)
	}
	kept := records[removeCount:]
	return s.rewriteLocked(kept)
}

func (s *JSONLStore) rewriteLocked(records []Record) error {
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}
	return nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

func pageRecords(records map[string]Record, query ListQuery) Page {
	limit := query.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	out := make([]Record, 0, len(records))
	for _, record := range records {
		if !recordMatchesListQuery(record, query) {
			continue
		}
		out = append(out, record)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StartedAt.After(out[j].StartedAt)
	})
	total := len(out)
	if offset >= total {
		return Page{Items: []Record{}, Total: total, Offset: offset, Limit: limit}
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return Page{Items: out[offset:end], Total: total, Offset: offset, Limit: limit}
}

func recordHasEventType(record Record, eventType string) bool {
	for _, event := range record.Events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func summarizeUsage(records map[string]Record, query UsageSummaryQuery) UsageSummary {
	groupBy := normalizedUsageGroup(query.GroupBy)
	summary := UsageSummary{GroupBy: groupBy}
	byKey := map[string]*UsageSummaryItem{}
	for _, record := range records {
		if !recordMatchesListQuery(record, query.ListQuery) {
			continue
		}
		summary.TotalRecords++
		if record.Usage == nil {
			continue
		}
		summary.UsageRecords++
		key := usageGroupKey(record, groupBy)
		item := byKey[key]
		if item == nil {
			item = &UsageSummaryItem{Key: key}
			byKey[key] = item
		}
		item.Records++
		item.PromptTokens += record.Usage.PromptTokens
		item.CompletionTokens += record.Usage.CompletionTokens
		item.TotalTokens += record.Usage.TotalTokens
		if record.Usage.Source == "unknown" {
			item.UnknownUsage++
		}
		summary.PromptTokens += record.Usage.PromptTokens
		summary.CompletionTokens += record.Usage.CompletionTokens
		summary.TotalTokens += record.Usage.TotalTokens
	}
	summary.Items = make([]UsageSummaryItem, 0, len(byKey))
	for _, item := range byKey {
		summary.Items = append(summary.Items, *item)
	}
	sort.Slice(summary.Items, func(i, j int) bool {
		if summary.Items[i].TotalTokens != summary.Items[j].TotalTokens {
			return summary.Items[i].TotalTokens > summary.Items[j].TotalTokens
		}
		return summary.Items[i].Key < summary.Items[j].Key
	})
	return summary
}

func recordMatchesListQuery(record Record, query ListQuery) bool {
	if query.Status != "" && record.Status != query.Status {
		return false
	}
	if query.AppID != "" && record.AppID != query.AppID {
		return false
	}
	if query.ProviderID != "" && record.ProviderID != query.ProviderID {
		return false
	}
	if query.EventType != "" && !recordHasEventType(record, query.EventType) {
		return false
	}
	return true
}

func normalizedUsageGroup(groupBy string) string {
	switch groupBy {
	case "app", "provider", "model":
		return groupBy
	default:
		return "provider"
	}
}

func usageGroupKey(record Record, groupBy string) string {
	switch groupBy {
	case "app":
		if record.AppID != "" {
			return record.AppID
		}
	case "model":
		if record.FinalModel != "" {
			return record.FinalModel
		}
		if record.Model != "" {
			return record.Model
		}
	default:
		if record.ProviderID != "" {
			return record.ProviderID
		}
	}
	return "unknown"
}

func sortedRecordsBySequence(records map[string]Record, sequences map[string]int64) []Record {
	out := make([]Record, 0, len(records))
	for _, record := range records {
		out = append(out, record)
	}
	sort.Slice(out, func(i, j int) bool {
		return sequences[out[i].TraceID] < sequences[out[j].TraceID]
	})
	return out
}

func NewID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	return "trace-" + hex.EncodeToString(b[:])
}
