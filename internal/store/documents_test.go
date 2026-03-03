package store_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"kb/internal/store"
)

func TestInsertDocument(t *testing.T) {
	s := newTestStore(t)

	id, err := s.InsertDocument("research", "Test Doc", "Some content", nil)
	if err != nil {
		t.Fatalf("InsertDocument: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}
}

func TestInsertDocument_StoresFields(t *testing.T) {
	s := newTestStore(t)

	meta := json.RawMessage(`{"key":"value"}`)
	id, err := s.InsertDocument("plan", "My Plan", "Plan content", meta)
	if err != nil {
		t.Fatal(err)
	}

	var doc store.Document
	err = s.DB().QueryRow(
		"SELECT id, type, title, content, metadata FROM documents WHERE id = ?", id,
	).Scan(&doc.ID, &doc.Type, &doc.Title, &doc.Content, &doc.Metadata)
	if err != nil {
		t.Fatal(err)
	}

	if doc.Type != "plan" {
		t.Errorf("type = %q, want %q", doc.Type, "plan")
	}
	if doc.Title != "My Plan" {
		t.Errorf("title = %q, want %q", doc.Title, "My Plan")
	}
	if doc.Content != "Plan content" {
		t.Errorf("content = %q, want %q", doc.Content, "Plan content")
	}
	if string(doc.Metadata) != `{"key":"value"}` {
		t.Errorf("metadata = %s, want %s", doc.Metadata, `{"key":"value"}`)
	}
}

func TestInsertDocument_NilMetadata(t *testing.T) {
	s := newTestStore(t)

	id, err := s.InsertDocument("note", "No Meta", "Content", nil)
	if err != nil {
		t.Fatal(err)
	}

	var metadata *string
	err = s.DB().QueryRow("SELECT metadata FROM documents WHERE id = ?", id).Scan(&metadata)
	if err != nil {
		t.Fatal(err)
	}
	if metadata != nil {
		t.Errorf("expected NULL metadata, got %q", *metadata)
	}
}

func TestInsertDocument_FTSSync(t *testing.T) {
	s := newTestStore(t)

	_, err := s.InsertDocument("research", "FTS Test Title", "searchable content here", nil)
	if err != nil {
		t.Fatal(err)
	}

	var title string
	err = s.DB().QueryRow(
		"SELECT title FROM documents_fts WHERE documents_fts MATCH ?", "searchable",
	).Scan(&title)
	if err != nil {
		t.Fatalf("FTS query failed: %v", err)
	}
	if title != "FTS Test Title" {
		t.Errorf("title = %q, want %q", title, "FTS Test Title")
	}
}

// --- List tests ---

func TestListDocuments_All(t *testing.T) {
	s := newTestStore(t)

	for i := range 3 {
		_, err := s.InsertDocument("note", fmt.Sprintf("Doc %d", i), "content", nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	docs, err := s.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 3 {
		t.Errorf("got %d docs, want 3", len(docs))
	}
}

func TestListDocuments_FilterByType(t *testing.T) {
	s := newTestStore(t)

	_, err := s.InsertDocument("research", "R1", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.InsertDocument("note", "N1", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	docs, err := s.ListDocuments("research")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("got %d docs, want 1", len(docs))
	}
	if docs[0].Type != "research" {
		t.Errorf("type = %q, want %q", docs[0].Type, "research")
	}
}

func TestListDocuments_Empty(t *testing.T) {
	s := newTestStore(t)

	docs, err := s.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if docs == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(docs) != 0 {
		t.Errorf("got %d docs, want 0", len(docs))
	}
}

func TestListDocuments_Order(t *testing.T) {
	s := newTestStore(t)

	_, err := s.InsertDocument("note", "First", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.InsertDocument("note", "Second", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	docs, err := s.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("got %d docs, want 2", len(docs))
	}
	// Most recent first.
	if docs[0].Title != "Second" {
		t.Errorf("first doc = %q, want %q", docs[0].Title, "Second")
	}
}

// --- Get tests ---

func TestGetDocument(t *testing.T) {
	s := newTestStore(t)

	meta := json.RawMessage(`{"tags":["go"]}`)
	id, err := s.InsertDocument("research", "Get Test", "get content", meta)
	if err != nil {
		t.Fatal(err)
	}

	doc, err := s.GetDocument(id)
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID != id {
		t.Errorf("id = %d, want %d", doc.ID, id)
	}
	if doc.Title != "Get Test" {
		t.Errorf("title = %q, want %q", doc.Title, "Get Test")
	}
	if doc.Content != "get content" {
		t.Errorf("content = %q, want %q", doc.Content, "get content")
	}
}

func TestGetDocument_NotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.GetDocument(999)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetDocument_MetadataPreserved(t *testing.T) {
	s := newTestStore(t)

	meta := json.RawMessage(`{"complex":{"nested":"value"},"list":[1,2,3]}`)
	id, err := s.InsertDocument("note", "Meta Test", "content", meta)
	if err != nil {
		t.Fatal(err)
	}

	doc, err := s.GetDocument(id)
	if err != nil {
		t.Fatal(err)
	}

	var expected, actual any
	if err := json.Unmarshal(meta, &expected); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(doc.Metadata, &actual); err != nil {
		t.Fatal(err)
	}
	// Compare via re-marshaling for stable ordering.
	e, _ := json.Marshal(expected)
	a, _ := json.Marshal(actual)
	if string(e) != string(a) {
		t.Errorf("metadata = %s, want %s", a, e)
	}
}

// --- Delete tests ---

func TestDeleteDocument(t *testing.T) {
	s := newTestStore(t)

	id, err := s.InsertDocument("note", "To Delete", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteDocument(id); err != nil {
		t.Fatal(err)
	}

	_, err = s.GetDocument(id)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDeleteDocument_NotFound(t *testing.T) {
	s := newTestStore(t)

	err := s.DeleteDocument(999)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteDocument_CascadesLinks(t *testing.T) {
	s := newTestStore(t)

	id1, err := s.InsertDocument("note", "Doc 1", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.InsertDocument("note", "Doc 2", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.CreateLink(id1, id2, "related"); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteDocument(id1); err != nil {
		t.Fatal(err)
	}

	// Link should be gone.
	links, err := s.GetLinks(id2)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 0 {
		t.Errorf("expected 0 links after cascade delete, got %d", len(links))
	}
}

func TestDeleteDocument_RemovesFTS(t *testing.T) {
	s := newTestStore(t)

	id, err := s.InsertDocument("note", "FTS Delete Test", "unique_keyword_xyz", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Verify FTS finds it.
	results, err := s.Search("unique_keyword_xyz", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 FTS result before delete, got %d", len(results))
	}

	if err := s.DeleteDocument(id); err != nil {
		t.Fatal(err)
	}

	// FTS should no longer find it.
	results, err = s.Search("unique_keyword_xyz", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 FTS results after delete, got %d", len(results))
	}
}
