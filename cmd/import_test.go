package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"kb/internal/store"
)

func TestExtractTitle_FromTopic(t *testing.T) {
	meta := map[string]interface{}{"topic": "My Topic"}
	got := extractTitle(meta, "# Heading\nContent", "2024-01-01-file.md")
	if got != "My Topic" {
		t.Errorf("got %q, want %q", got, "My Topic")
	}
}

func TestExtractTitle_FromTitle(t *testing.T) {
	meta := map[string]interface{}{"title": "My Title"}
	got := extractTitle(meta, "# Heading\nContent", "file.md")
	if got != "My Title" {
		t.Errorf("got %q, want %q", got, "My Title")
	}
}

func TestExtractTitle_FromHeading(t *testing.T) {
	meta := map[string]interface{}{}
	got := extractTitle(meta, "# My Heading\nContent", "file.md")
	if got != "My Heading" {
		t.Errorf("got %q, want %q", got, "My Heading")
	}
}

func TestExtractTitle_FromFilename(t *testing.T) {
	meta := map[string]interface{}{}
	got := extractTitle(meta, "No heading here", "2024-01-01-my-document.md")
	if got != "my-document" {
		t.Errorf("got %q, want %q", got, "my-document")
	}
}

func TestExtractTitle_Priority(t *testing.T) {
	meta := map[string]interface{}{"topic": "Topic Wins", "title": "Title Loses"}
	got := extractTitle(meta, "# Heading\nContent", "file.md")
	if got != "Topic Wins" {
		t.Errorf("got %q, want %q", got, "Topic Wins")
	}
}

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

func writeTempMD(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestImportFile_Success(t *testing.T) {
	s := newTestStore(t)
	cmd := &ImportCmd{Type: "research"}
	path := writeTempMD(t, "# My Title\n\nSome content here.")

	id, title, err := cmd.importFile(s, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("got id %d, want 1", id)
	}
	if title != "My Title" {
		t.Errorf("got title %q, want %q", title, "My Title")
	}
}

func TestImportFile_WithFrontmatter(t *testing.T) {
	s := newTestStore(t)
	cmd := &ImportCmd{Type: "plan"}
	path := writeTempMD(t, "---\ntitle: FM Title\nauthor: test\n---\n\nBody content.")

	id, title, err := cmd.importFile(s, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("got id %d, want 1", id)
	}
	if title != "FM Title" {
		t.Errorf("got title %q, want %q", title, "FM Title")
	}
}

func TestImportFile_ReadError(t *testing.T) {
	s := newTestStore(t)
	cmd := &ImportCmd{Type: "research"}

	_, _, err := cmd.importFile(s, "/nonexistent/path.md")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestImportRun_PartialFailure(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.InitSchema(); err != nil {
		t.Fatal(err)
	}
	s.Close()

	goodFile := writeTempMD(t, "# Good Doc\n\nValid content.")
	badFile := filepath.Join(t.TempDir(), "nonexistent.md")

	cmd := &ImportCmd{
		Files: []string{goodFile, badFile},
		Type:  "research",
	}
	cli := &CLI{DB: dbPath}

	err = cmd.Run(cli)
	if err == nil {
		t.Fatal("expected non-nil error for partial failure")
	}

	// Verify the good file was imported.
	s2, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	docs, err := s2.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("got %d documents, want 1", len(docs))
	}
	if docs[0].Title != "Good Doc" {
		t.Errorf("got title %q, want %q", docs[0].Title, "Good Doc")
	}
}
