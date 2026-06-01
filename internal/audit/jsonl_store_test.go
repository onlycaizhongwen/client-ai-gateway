package audit

import (
	"path/filepath"
	"testing"
)

func TestJSONLStorePersistsEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	store, err := NewJSONLStore(path)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	store.Save(Event{
		AppID:  "admin-app",
		Action: "config.reload",
		Target: "runtime",
		Result: ResultSuccess,
	})

	reopened, err := NewJSONLStore(path)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	events := reopened.List(ListQuery{Action: "config.reload"})
	if len(events) != 1 {
		t.Fatalf("expected one event, got %+v", events)
	}
	if events[0].AppID != "admin-app" || events[0].Result != ResultSuccess {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}

func TestJSONLStoreFilters(t *testing.T) {
	store, err := NewJSONLStore(filepath.Join(t.TempDir(), "audit.jsonl"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	store.Save(Event{AppID: "a", Action: "config.reload", Result: ResultSuccess})
	store.Save(Event{AppID: "b", Action: "policy.dry_run", Result: ResultDenied})

	events := store.List(ListQuery{Result: ResultDenied})
	if len(events) != 1 || events[0].Action != "policy.dry_run" {
		t.Fatalf("unexpected filtered events: %+v", events)
	}
}

func TestJSONLStoreRetentionKeepsNewestEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	store, err := NewJSONLStoreWithRetention(path, 2)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	store.Save(Event{AppID: "a", Action: "first", Result: ResultSuccess})
	store.Save(Event{AppID: "b", Action: "second", Result: ResultSuccess})
	store.Save(Event{AppID: "c", Action: "third", Result: ResultSuccess})

	events := store.List(ListQuery{Limit: 10})
	if len(events) != 2 || events[0].Action != "third" || events[1].Action != "second" {
		t.Fatalf("unexpected retained events: %+v", events)
	}

	reopened, err := NewJSONLStoreWithRetention(path, 2)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	events = reopened.List(ListQuery{Limit: 10})
	if len(events) != 2 || events[0].Action != "third" || events[1].Action != "second" {
		t.Fatalf("unexpected retained events after reload: %+v", events)
	}
}

func TestMemoryStorePageFiltersAndOffsets(t *testing.T) {
	store := NewMemoryStore()
	store.Save(Event{AppID: "a", Action: "config.reload", Result: ResultSuccess})
	store.Save(Event{AppID: "b", Action: "policy.dry_run", Result: ResultDenied})
	store.Save(Event{AppID: "c", Action: "provider.probe", Result: ResultSuccess})
	store.Save(Event{AppID: "d", Action: "provider.enabled", Result: ResultSuccess})

	page := store.Page(ListQuery{Result: ResultSuccess, Offset: 1, Limit: 2})
	if page.Total != 3 || page.Offset != 1 || page.Limit != 2 {
		t.Fatalf("unexpected page metadata: %+v", page)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected two events, got %+v", page.Items)
	}
	if page.Items[0].Action != "provider.probe" || page.Items[1].Action != "config.reload" {
		t.Fatalf("expected reverse chronological filtered page, got %+v", page.Items)
	}
}
