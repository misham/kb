package store

import "errors"

// Sentinel errors for store operations.
var (
	ErrNotFound      = errors.New("document not found")
	ErrDuplicateLink = errors.New("link already exists")
)
