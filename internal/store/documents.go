package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// Document represents a knowledge base document.
type Document struct {
	ID        int64
	Type      string
	Title     string
	Content   string
	Metadata  json.RawMessage
	CreatedAt string
	UpdatedAt string
}

// InsertDocument inserts a new document and returns its ID.
func (s *Store) InsertDocument(docType, title, content string, metadata json.RawMessage) (int64, error) {
	result, err := s.db.ExecContext(context.Background(),
		"INSERT INTO documents (type, title, content, metadata) VALUES (?, ?, ?, ?)",
		docType, title, content, metadata,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting document: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// ListDocuments returns all documents, optionally filtered by type.
// Results are ordered by id descending (most recent first).
func (s *Store) ListDocuments(docType string) ([]Document, error) {
	ctx := context.Background()

	var (
		query string
		args  []any
	)

	if docType != "" {
		query = "SELECT id, type, title, content, metadata, created_at, updated_at FROM documents WHERE type = ? ORDER BY id DESC"
		args = []any{docType}
	} else {
		query = "SELECT id, type, title, content, metadata, created_at, updated_at FROM documents ORDER BY id DESC"
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing documents: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		var meta *[]byte
		if err := rows.Scan(&d.ID, &d.Type, &d.Title, &d.Content, &meta, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning document: %w", err)
		}
		if meta != nil {
			d.Metadata = *meta
		}
		docs = append(docs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating documents: %w", err)
	}

	if docs == nil {
		docs = []Document{}
	}
	return docs, nil
}

// GetDocument returns a single document by ID.
func (s *Store) GetDocument(id int64) (*Document, error) {
	var d Document
	var meta *[]byte
	err := s.db.QueryRowContext(context.Background(),
		"SELECT id, type, title, content, metadata, created_at, updated_at FROM documents WHERE id = ?", id,
	).Scan(&d.ID, &d.Type, &d.Title, &d.Content, &meta, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("document %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("querying document %d: %w", id, err)
	}
	if meta != nil {
		d.Metadata = *meta
	}
	return &d, nil
}

// DeleteDocument deletes a document by ID.
func (s *Store) DeleteDocument(id int64) error {
	result, err := s.db.ExecContext(context.Background(),
		"DELETE FROM documents WHERE id = ?", id,
	)
	if err != nil {
		return fmt.Errorf("deleting document %d: %w", id, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("document %d: %w", id, ErrNotFound)
	}
	return nil
}
