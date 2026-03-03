package store

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schemaSQL string

// InitSchema executes the embedded schema SQL to create tables, indexes, and triggers.
func (s *Store) InitSchema() error {
	if _, err := s.db.ExecContext(context.Background(), schemaSQL); err != nil {
		return fmt.Errorf("initializing schema: %w", err)
	}
	return nil
}
