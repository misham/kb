package store

import (
	"context"
	"fmt"
	"strings"
)

// Link represents a directional link between two documents.
type Link struct {
	SourceID     int64
	TargetID     int64
	Relationship string
}

// LinkedDocument represents a document linked to another, with relationship info.
type LinkedDocument struct {
	Document
	Relationship string
	Direction    string // "outgoing" or "incoming".
}

// CreateLink creates a link between two documents with an optional relationship label.
func (s *Store) CreateLink(sourceID, targetID int64, relationship string) error {
	var rel *string
	if relationship != "" {
		rel = &relationship
	}

	_, err := s.db.ExecContext(context.Background(),
		"INSERT INTO document_links (source_id, target_id, relationship) VALUES (?, ?, ?)",
		sourceID, targetID, rel,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "PRIMARY KEY") {
			return fmt.Errorf("link %d->%d: %w", sourceID, targetID, ErrDuplicateLink)
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return fmt.Errorf("one or both documents do not exist: %w", ErrNotFound)
		}
		return fmt.Errorf("creating link: %w", err)
	}
	return nil
}

// GetLinks returns all documents linked to the given document ID (both directions).
func (s *Store) GetLinks(documentID int64) ([]LinkedDocument, error) {
	ctx := context.Background()

	query := `
		SELECT d.id, d.type, d.title, d.content, d.metadata, d.created_at, d.updated_at,
			   l.relationship, 'outgoing' as direction
		FROM document_links l
		JOIN documents d ON d.id = l.target_id
		WHERE l.source_id = ?
		UNION ALL
		SELECT d.id, d.type, d.title, d.content, d.metadata, d.created_at, d.updated_at,
			   l.relationship, 'incoming' as direction
		FROM document_links l
		JOIN documents d ON d.id = l.source_id
		WHERE l.target_id = ?
	`

	rows, err := s.db.QueryContext(ctx, query, documentID, documentID)
	if err != nil {
		return nil, fmt.Errorf("getting links for document %d: %w", documentID, err)
	}
	defer rows.Close()

	var links []LinkedDocument
	for rows.Next() {
		var ld LinkedDocument
		var rel *string
		var meta *[]byte
		if err := rows.Scan(
			&ld.ID, &ld.Type, &ld.Title, &ld.Content, &meta,
			&ld.CreatedAt, &ld.UpdatedAt, &rel, &ld.Direction,
		); err != nil {
			return nil, fmt.Errorf("scanning linked document: %w", err)
		}
		if meta != nil {
			ld.Metadata = *meta
		}
		if rel != nil {
			ld.Relationship = *rel
		}
		links = append(links, ld)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating linked documents: %w", err)
	}

	if links == nil {
		links = []LinkedDocument{}
	}
	return links, nil
}
