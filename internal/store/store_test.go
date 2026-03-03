package store_test

import (
	"testing"

	"kb/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.InitSchema(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestOpen(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
}

func TestOpen_PRAGMAs(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	tests := []struct {
		pragma string
		want   int
	}{
		{"foreign_keys", 1},
		{"busy_timeout", 5000},
		{"synchronous", 1}, // NORMAL = 1
	}

	for _, tt := range tests {
		var got int
		if err := s.DB().QueryRow("PRAGMA " + tt.pragma).Scan(&got); err != nil {
			t.Fatalf("PRAGMA %s: %v", tt.pragma, err)
		}
		if got != tt.want {
			t.Errorf("PRAGMA %s = %d, want %d", tt.pragma, got, tt.want)
		}
	}
}

func TestInitSchema_CreatesTables(t *testing.T) {
	s := newTestStore(t)

	tables := []string{"documents", "document_links", "documents_fts"}
	for _, table := range tables {
		var name string
		err := s.DB().QueryRow(
			"SELECT name FROM sqlite_master WHERE type IN ('table', 'shadow') AND name = ?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestInitSchema_CreatesIndexes(t *testing.T) {
	s := newTestStore(t)

	indexes := []string{"idx_doc_type", "idx_link_source", "idx_link_target"}
	for _, idx := range indexes {
		var name string
		err := s.DB().QueryRow(
			"SELECT name FROM sqlite_master WHERE type = 'index' AND name = ?", idx,
		).Scan(&name)
		if err != nil {
			t.Errorf("index %q not found: %v", idx, err)
		}
	}
}

func TestInitSchema_CreatesTriggers(t *testing.T) {
	s := newTestStore(t)

	triggers := []string{"documents_ai", "documents_ad", "documents_au"}
	for _, trig := range triggers {
		var name string
		err := s.DB().QueryRow(
			"SELECT name FROM sqlite_master WHERE type = 'trigger' AND name = ?", trig,
		).Scan(&name)
		if err != nil {
			t.Errorf("trigger %q not found: %v", trig, err)
		}
	}
}

func TestInitSchema_Idempotent(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	if err := s.InitSchema(); err != nil {
		t.Fatalf("first InitSchema: %v", err)
	}
	if err := s.InitSchema(); err != nil {
		t.Fatalf("second InitSchema: %v", err)
	}
}
