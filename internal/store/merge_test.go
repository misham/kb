package store_test

import (
	"encoding/json"
	"testing"

	"kb/internal/store"
)

func TestMergeDB_NewDocuments(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	// Insert a document in each side with different titles.
	_, err := ours.InsertDocument("note", "Ours Only", "ours content", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = theirs.InsertDocument("note", "Theirs Only", "theirs content", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}

	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}

	titles := map[string]bool{}
	for _, d := range docs {
		titles[d.Title] = true
	}
	if !titles["Ours Only"] || !titles["Theirs Only"] {
		t.Errorf("expected both titles, got %v", titles)
	}
}

func TestMergeDB_ConflictKeepsNewer(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	// Insert the same (type, title) in both, but theirs is newer.
	_, err := ours.InsertDocument("note", "Shared", "old content", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Manually set updated_at so we can control which is newer.
	_, err = ours.DB().Exec(`UPDATE documents SET updated_at = '2024-01-01 00:00:00' WHERE title = 'Shared'`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = theirs.InsertDocument("note", "Shared", "new content", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = theirs.DB().Exec(`UPDATE documents SET updated_at = '2025-01-01 00:00:00' WHERE title = 'Shared'`)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}

	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	if docs[0].Content != "new content" {
		t.Errorf("expected theirs content, got %q", docs[0].Content)
	}
}

func TestMergeDB_ConflictKeepsOursWhenNewer(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	_, err := ours.InsertDocument("note", "Shared", "ours content", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ours.DB().Exec(`UPDATE documents SET updated_at = '2025-06-01 00:00:00' WHERE title = 'Shared'`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = theirs.InsertDocument("note", "Shared", "theirs content", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = theirs.DB().Exec(`UPDATE documents SET updated_at = '2024-06-01 00:00:00' WHERE title = 'Shared'`)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	if docs[0].Content != "ours content" {
		t.Errorf("expected ours content preserved, got %q", docs[0].Content)
	}
}

func TestMergeDB_LinksRemapped(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	// Ours has doc A.
	_, err := ours.InsertDocument("note", "A", "a", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Theirs has docs B and C with a link B->C.
	bID, err := theirs.InsertDocument("note", "B", "b", nil)
	if err != nil {
		t.Fatal(err)
	}
	cID, err := theirs.InsertDocument("note", "C", "c", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := theirs.CreateLink(bID, cID, "related"); err != nil {
		t.Fatal(err)
	}

	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	// Ours should now have A, B, C and a link between the merged B and C.
	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}

	// Find the merged B and check its links.
	var mergedBID int64
	for _, d := range docs {
		if d.Title == "B" {
			mergedBID = d.ID
			break
		}
	}
	links, err := ours.GetLinks(mergedBID)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Title != "C" {
		t.Errorf("expected link to C, got %q", links[0].Title)
	}
	if links[0].Relationship != "related" {
		t.Errorf("expected relationship 'related', got %q", links[0].Relationship)
	}
}

func TestMergeDB_MetadataPreserved(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	meta := json.RawMessage(`{"key":"value"}`)
	_, err := theirs.InsertDocument("note", "WithMeta", "content", meta)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}

	var m map[string]string
	if err := json.Unmarshal(docs[0].Metadata, &m); err != nil {
		t.Fatalf("unmarshaling metadata: %v", err)
	}
	if m["key"] != "value" {
		t.Errorf("expected metadata key=value, got %v", m)
	}
}

func TestMergeDB_TimestampsPreservedOnInsert(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	_, err := theirs.InsertDocument("note", "Old Doc", "content", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Set specific timestamps on theirs.
	_, err = theirs.DB().Exec(
		`UPDATE documents SET created_at = '2020-03-15 10:00:00', updated_at = '2023-07-20 14:30:00' WHERE title = 'Old Doc'`,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	doc := docs[0]
	if doc.CreatedAt != "2020-03-15 10:00:00" {
		t.Errorf("expected created_at preserved, got %q", doc.CreatedAt)
	}
	if doc.UpdatedAt != "2023-07-20 14:30:00" {
		t.Errorf("expected updated_at preserved, got %q", doc.UpdatedAt)
	}
}

func TestMergeDB_DuplicateLinkIgnored(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	// Create the same documents and link on both sides.
	aOurs, err := ours.InsertDocument("note", "A", "a", nil)
	if err != nil {
		t.Fatal(err)
	}
	bOurs, err := ours.InsertDocument("note", "B", "b", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := ours.CreateLink(aOurs, bOurs, "related"); err != nil {
		t.Fatal(err)
	}

	aTheirs, err := theirs.InsertDocument("note", "A", "a", nil)
	if err != nil {
		t.Fatal(err)
	}
	bTheirs, err := theirs.InsertDocument("note", "B", "b", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := theirs.CreateLink(aTheirs, bTheirs, "related"); err != nil {
		t.Fatal(err)
	}

	// Merge should succeed without errors — duplicate link is silently ignored.
	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	links, err := ours.GetLinks(aOurs)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Errorf("expected 1 link (no duplicate), got %d", len(links))
	}
}

func TestMergeDB_EmptyDatabases(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	// Merging two empty databases should be a no-op.
	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB with empty dbs: %v", err)
	}

	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Errorf("expected 0 documents, got %d", len(docs))
	}
}

func TestMergeDB_EmptyTheirsIsNoop(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	_, err := ours.InsertDocument("note", "Existing", "content", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Errorf("expected 1 document unchanged, got %d", len(docs))
	}
}

func TestMergeDB_SameTimestampKeepsOurs(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	ts := "2025-01-01 12:00:00"

	_, err := ours.InsertDocument("note", "Tied", "ours version", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ours.DB().Exec(`UPDATE documents SET updated_at = ? WHERE title = 'Tied'`, ts)
	if err != nil {
		t.Fatal(err)
	}

	_, err = theirs.InsertDocument("note", "Tied", "theirs version", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = theirs.DB().Exec(`UPDATE documents SET updated_at = ? WHERE title = 'Tied'`, ts)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	// Same timestamp — ours wins (> not >=).
	if docs[0].Content != "ours version" {
		t.Errorf("expected ours to win on tie, got %q", docs[0].Content)
	}
}

func TestMergeDB_ConflictUpdatesMetadata(t *testing.T) {
	ours := newTestStore(t)
	theirs := newTestStore(t)

	oldMeta := json.RawMessage(`{"status":"draft"}`)
	newMeta := json.RawMessage(`{"status":"published","reviewer":"alice"}`)

	_, err := ours.InsertDocument("note", "Doc", "content", oldMeta)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ours.DB().Exec(`UPDATE documents SET updated_at = '2024-01-01 00:00:00' WHERE title = 'Doc'`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = theirs.InsertDocument("note", "Doc", "updated content", newMeta)
	if err != nil {
		t.Fatal(err)
	}
	_, err = theirs.DB().Exec(`UPDATE documents SET updated_at = '2025-01-01 00:00:00' WHERE title = 'Doc'`)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MergeDB(ours, theirs); err != nil {
		t.Fatalf("MergeDB: %v", err)
	}

	docs, err := ours.ListDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}

	var m map[string]string
	if err := json.Unmarshal(docs[0].Metadata, &m); err != nil {
		t.Fatalf("unmarshaling metadata: %v", err)
	}
	if m["status"] != "published" {
		t.Errorf("expected metadata status=published, got %q", m["status"])
	}
	if m["reviewer"] != "alice" {
		t.Errorf("expected metadata reviewer=alice, got %q", m["reviewer"])
	}
}
