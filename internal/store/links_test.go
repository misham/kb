package store_test

import (
	"errors"
	"testing"

	"kb/internal/store"
)

func TestCreateLink(t *testing.T) {
	s := newTestStore(t)

	id1, err := s.InsertDocument("note", "Doc 1", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.InsertDocument("note", "Doc 2", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.CreateLink(id1, id2, ""); err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
}

func TestCreateLink_WithRelationship(t *testing.T) {
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

	links, err := s.GetLinks(id1)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1", len(links))
	}
	if links[0].Relationship != "related" {
		t.Errorf("relationship = %q, want %q", links[0].Relationship, "related")
	}
}

func TestCreateLink_DuplicateError(t *testing.T) {
	s := newTestStore(t)

	id1, err := s.InsertDocument("note", "Doc 1", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.InsertDocument("note", "Doc 2", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.CreateLink(id1, id2, ""); err != nil {
		t.Fatal(err)
	}

	err = s.CreateLink(id1, id2, "")
	if !errors.Is(err, store.ErrDuplicateLink) {
		t.Errorf("expected ErrDuplicateLink, got %v", err)
	}
}

func TestCreateLink_InvalidDocument(t *testing.T) {
	s := newTestStore(t)

	id1, err := s.InsertDocument("note", "Doc 1", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = s.CreateLink(id1, 999, "")
	if err == nil {
		t.Fatal("expected error for non-existent target, got nil")
	}
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetLinks_Outgoing(t *testing.T) {
	s := newTestStore(t)

	id1, err := s.InsertDocument("note", "Source", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.InsertDocument("note", "Target", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.CreateLink(id1, id2, "refs"); err != nil {
		t.Fatal(err)
	}

	links, err := s.GetLinks(id1)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1", len(links))
	}
	if links[0].Title != "Target" {
		t.Errorf("linked doc = %q, want %q", links[0].Title, "Target")
	}
	if links[0].Direction != "outgoing" {
		t.Errorf("direction = %q, want %q", links[0].Direction, "outgoing")
	}
}

func TestGetLinks_Incoming(t *testing.T) {
	s := newTestStore(t)

	id1, err := s.InsertDocument("note", "Source", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.InsertDocument("note", "Target", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.CreateLink(id1, id2, "refs"); err != nil {
		t.Fatal(err)
	}

	links, err := s.GetLinks(id2)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1", len(links))
	}
	if links[0].Title != "Source" {
		t.Errorf("linked doc = %q, want %q", links[0].Title, "Source")
	}
	if links[0].Direction != "incoming" {
		t.Errorf("direction = %q, want %q", links[0].Direction, "incoming")
	}
}

func TestGetLinks_Both(t *testing.T) {
	s := newTestStore(t)

	id1, err := s.InsertDocument("note", "A", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.InsertDocument("note", "B", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	id3, err := s.InsertDocument("note", "C", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	// B -> A (incoming to A), A -> C (outgoing from A).
	if err := s.CreateLink(id2, id1, ""); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateLink(id1, id3, ""); err != nil {
		t.Fatal(err)
	}

	links, err := s.GetLinks(id1)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 {
		t.Fatalf("got %d links, want 2", len(links))
	}
}

func TestGetLinks_NoLinks(t *testing.T) {
	s := newTestStore(t)

	id, err := s.InsertDocument("note", "Lonely", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	links, err := s.GetLinks(id)
	if err != nil {
		t.Fatal(err)
	}
	if links == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(links) != 0 {
		t.Errorf("got %d links, want 0", len(links))
	}
}
