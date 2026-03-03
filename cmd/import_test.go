package cmd

import "testing"

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
