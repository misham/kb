package store_test

import (
	"testing"

	"kb/internal/store"
)

func TestSearch_FindsDocument(t *testing.T) {
	s := newTestStore(t)

	_, err := s.InsertDocument("research", "Go Concurrency", "Goroutines and channels are fundamental", nil)
	if err != nil {
		t.Fatal(err)
	}

	results, err := s.Search("goroutines", "")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Title != "Go Concurrency" {
		t.Errorf("title = %q, want %q", results[0].Title, "Go Concurrency")
	}
}

func TestSearch_RanksResults(t *testing.T) {
	s := newTestStore(t)

	// Doc with term in both title and content should rank higher.
	_, err := s.InsertDocument("research", "SQLite Performance", "SQLite is fast and reliable", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.InsertDocument("note", "Database Notes", "Some notes about SQLite usage", nil)
	if err != nil {
		t.Fatal(err)
	}

	results, err := s.Search("SQLite", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) < 2 {
		t.Fatalf("got %d results, want >= 2", len(results))
	}
	// The doc with "SQLite" in title+content should rank first.
	if results[0].Title != "SQLite Performance" {
		t.Errorf("first result = %q, want %q", results[0].Title, "SQLite Performance")
	}
}

func TestSearch_TypeFilter(t *testing.T) {
	s := newTestStore(t)

	_, err := s.InsertDocument("research", "Research Doc", "testing content", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.InsertDocument("note", "Note Doc", "testing content", nil)
	if err != nil {
		t.Fatal(err)
	}

	results, err := s.Search("testing", "research")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Type != "research" {
		t.Errorf("type = %q, want %q", results[0].Type, "research")
	}
}

func TestSearch_NoResults(t *testing.T) {
	s := newTestStore(t)

	results, err := s.Search("nonexistent", "")
	if err != nil {
		t.Fatal(err)
	}
	if results == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestSearch_UnbalancedQuotes(t *testing.T) {
	s := newTestStore(t)

	_, err := s.InsertDocument("note", "Test", "some content", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Should not error on unbalanced quotes.
	_, err = s.Search(`"unbalanced`, "")
	if err != nil {
		t.Errorf("unexpected error on unbalanced quotes: %v", err)
	}
}

func TestSearch_FTS5Syntax(t *testing.T) {
	s := newTestStore(t)

	_, err := s.InsertDocument("note", "Exact Phrase Doc", "this is an exact phrase test", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.InsertDocument("note", "Partial Match", "this is a test", nil)
	if err != nil {
		t.Fatal(err)
	}

	results, err := s.Search(`"exact phrase"`, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Title != "Exact Phrase Doc" {
		t.Errorf("title = %q, want %q", results[0].Title, "Exact Phrase Doc")
	}
}

func TestSearch_SpecialCharacters(t *testing.T) {
	s := newTestStore(t)

	_, err := s.InsertDocument("note", "Test", "some content", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Should not error on special characters.
	for _, q := range []string{"*", "(", ")", "foo(bar)"} {
		_, err = s.Search(q, "")
		if err != nil {
			t.Errorf("unexpected error on %q: %v", q, err)
		}
	}
}

func TestSanitizeQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple term", "hello", "hello"},
		{"multiple terms", "hello world", "hello world"},
		{"quoted phrase passthrough", `"exact match"`, `"exact match"`},
		{"AND passthrough", "foo AND bar", "foo AND bar"},
		{"OR passthrough", "foo OR bar", "foo OR bar"},
		{"NOT passthrough", "NOT foo", "NOT foo"},
		{"unbalanced quote", `"hello`, "hello"},
		{"stray parens", "foo(bar)", "foobar"},
		{"stray star alone", "*", ""},
		{"prefix star passthrough", "hel*", "hel*"},
		{"NEAR with parens passthrough", "NEAR(foo bar)", "NEAR(foo bar)"},
		{"NEAR with spaces passthrough", "foo NEAR bar", "foo NEAR bar"},
		{"nearby not treated as FTS5", "nearby", "nearby"},
		{"linear not treated as FTS5", "linear", "linear"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := store.SanitizeQuery(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeQuery(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
