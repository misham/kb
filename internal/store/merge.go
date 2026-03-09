package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type docKey struct {
	Type  string
	Title string
}

// MergeDB merges documents and links from the "theirs" database into "ours".
// Documents are matched by (type, title). When both sides have the same
// document, the one with the later updated_at timestamp wins (ties keep ours).
// New documents from "theirs" are inserted with their original timestamps
// preserved, and links are remapped to the merged IDs. The entire merge runs
// in a single transaction so a failure leaves "ours" unchanged.
//
// Links are matched by (source_id, target_id). If both sides have the same
// link with different relationships, the "ours" relationship is kept.
func MergeDB(ours, theirs *Store) error {
	ctx := context.Background()

	// 1. Read all documents from theirs.
	theirDocs, err := allDocuments(ctx, theirs)
	if err != nil {
		return fmt.Errorf("reading theirs documents: %w", err)
	}

	// 2. Read all documents from ours and build a lookup by (type, title).
	ourDocs, err := allDocuments(ctx, ours)
	if err != nil {
		return fmt.Errorf("reading ours documents: %w", err)
	}

	// 3. Read all links from theirs before starting the transaction.
	theirLinks, err := allLinks(ctx, theirs)
	if err != nil {
		return fmt.Errorf("reading theirs links: %w", err)
	}

	// 4. Run all mutations inside a transaction.
	tx, err := ours.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	ourByKey := make(map[docKey]Document, len(ourDocs))
	for _, d := range ourDocs {
		ourByKey[docKey{d.Type, d.Title}] = d
	}

	// idMap maps theirs document ID → ours document ID after merge.
	idMap := make(map[int64]int64)

	for _, td := range theirDocs {
		newID, err := mergeDocument(ctx, tx, td, ourByKey)
		if err != nil {
			return err
		}
		idMap[td.ID] = newID
	}

	// 5. Merge links from theirs, remapping IDs.
	if err := mergeLinks(ctx, tx, theirLinks, idMap); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing merge: %w", err)
	}
	return nil
}

// mergeDocument merges a single document from theirs into ours. It returns the
// ID of the document in the ours database after the merge.
func mergeDocument(ctx context.Context, tx *sql.Tx, td Document, ourByKey map[docKey]Document) (int64, error) {
	key := docKey{td.Type, td.Title}
	od, exists := ourByKey[key]
	if exists {
		// Both sides have this document. Keep the newer one.
		if td.UpdatedAt > od.UpdatedAt {
			_, err := tx.ExecContext(ctx,
				`UPDATE documents SET content = ?, metadata = ?, updated_at = ? WHERE id = ?`,
				td.Content, td.Metadata, td.UpdatedAt, od.ID,
			)
			if err != nil {
				return 0, fmt.Errorf("updating document %d: %w", od.ID, err)
			}
		}
		return od.ID, nil
	}

	// New document from theirs — insert with original timestamps.
	result, err := tx.ExecContext(ctx,
		`INSERT INTO documents (type, title, content, metadata, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		td.Type, td.Title, td.Content, td.Metadata, td.CreatedAt, td.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting document from theirs: %w", err)
	}
	newID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return newID, nil
}

// mergeLinks copies links from theirs into ours, remapping document IDs.
// Links that already exist in ours are silently skipped.
func mergeLinks(ctx context.Context, tx *sql.Tx, theirLinks []Link, idMap map[int64]int64) error {
	for _, l := range theirLinks {
		srcID, srcOK := idMap[l.SourceID]
		tgtID, tgtOK := idMap[l.TargetID]
		if !srcOK || !tgtOK {
			continue
		}

		var rel *string
		if l.Relationship != "" {
			rel = &l.Relationship
		}

		_, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO document_links (source_id, target_id, relationship)
			 VALUES (?, ?, ?)`,
			srcID, tgtID, rel,
		)
		if err != nil {
			return fmt.Errorf("creating link %d->%d: %w", srcID, tgtID, err)
		}
	}
	return nil
}

// allDocuments returns every document in the store.
func allDocuments(ctx context.Context, s *Store) ([]Document, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, type, title, content, metadata, created_at, updated_at FROM documents ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		var meta *[]byte
		if err := rows.Scan(&d.ID, &d.Type, &d.Title, &d.Content, &meta, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		if meta != nil {
			d.Metadata = json.RawMessage(*meta)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// allLinks returns every link in the store.
func allLinks(ctx context.Context, s *Store) ([]Link, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT source_id, target_id, COALESCE(relationship, '') FROM document_links`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.SourceID, &l.TargetID, &l.Relationship); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}
