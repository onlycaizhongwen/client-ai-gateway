package trace

import (
	"path/filepath"
	"testing"
	"time"
)

func TestJSONLStorePersistsAndReloads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "traces.jsonl")
	store, err := NewJSONLStore(path)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	record := Record{
		TraceID:   "trace-test",
		AppID:     "dev-app",
		Model:     "local-small",
		Status:    "completed",
		StartedAt: time.Now().UTC(),
		Events:    []Event{{Type: "request_started", Message: "ok", At: time.Now().UTC()}},
	}
	if err := store.Save(record); err != nil {
		t.Fatalf("save: %v", err)
	}

	reloaded, err := NewJSONLStore(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, ok := reloaded.Get(record.TraceID)
	if !ok {
		t.Fatal("expected persisted record")
	}
	if got.AppID != record.AppID || got.Status != record.Status {
		t.Fatalf("unexpected record: %+v", got)
	}
}

func TestMemoryStorePageFiltersAndOffsets(t *testing.T) {
	store := NewMemoryStore()
	base := time.Date(2026, 5, 30, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		status := "completed"
		if i%2 == 0 {
			status = "failed"
		}
		if err := store.Save(Record{
			TraceID:    "trace-page-" + string(rune('a'+i)),
			AppID:      "dev-app",
			ProviderID: "local-mock",
			Status:     status,
			StartedAt:  base.Add(time.Duration(i) * time.Minute),
		}); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	page := store.Page(ListQuery{Status: "failed", Offset: 1, Limit: 2})
	if page.Total != 3 || page.Offset != 1 || page.Limit != 2 {
		t.Fatalf("unexpected page metadata: %+v", page)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected two items, got %+v", page.Items)
	}
	if page.Items[0].TraceID != "trace-page-c" || page.Items[1].TraceID != "trace-page-a" {
		t.Fatalf("expected newest filtered page after offset, got %+v", page.Items)
	}
}

func TestJSONLStoreRetentionKeepsNewestRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "traces.jsonl")
	store, err := NewJSONLStoreWithRetention(path, 2)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	base := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		if err := store.Save(Record{
			TraceID:   "trace-retain-" + string(rune('a'+i)),
			Status:    "completed",
			StartedAt: base.Add(time.Duration(i) * time.Minute),
		}); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	if _, ok := store.Get("trace-retain-a"); ok {
		t.Fatal("expected oldest trace to be pruned")
	}
	if _, ok := store.Get("trace-retain-b"); !ok {
		t.Fatal("expected newer trace b to remain")
	}
	if _, ok := store.Get("trace-retain-c"); !ok {
		t.Fatal("expected newer trace c to remain")
	}

	reloaded, err := NewJSONLStoreWithRetention(path, 2)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	page := reloaded.Page(ListQuery{Limit: 10})
	if page.Total != 2 {
		t.Fatalf("expected two retained records, got %+v", page)
	}
}

func TestMemoryStorePageOffsetPastEnd(t *testing.T) {
	store := NewMemoryStore()
	if err := store.Save(Record{TraceID: "trace-one", Status: "completed", StartedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("save: %v", err)
	}

	page := store.Page(ListQuery{Offset: 10, Limit: 5})
	if page.Total != 1 || len(page.Items) != 0 || page.Offset != 10 {
		t.Fatalf("expected empty page past end, got %+v", page)
	}
}
