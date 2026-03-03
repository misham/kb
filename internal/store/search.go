package store

import (
	"context"
	"fmt"
	"strings"
)

// SearchResult represents a single search result with ranking info.
type SearchResult struct {
	ID      int64
	Type    string
	Title   string
	Rank    float64
	Snippet string
}

// Search performs FTS5 full-text search with BM25 ranking.
// If docType is non-empty, results are filtered to that type.
func (s *Store) Search(query string, docType string) ([]SearchResult, error) {
	sanitized := SanitizeQuery(query)
	if sanitized == "" {
		return []SearchResult{}, nil
	}

	var (
		sqlQuery string
		args     []any
	)

	if docType != "" {
		sqlQuery = `SELECT d.id, d.type, d.title, bm25(documents_fts) as rank,
			snippet(documents_fts, 1, '»', '«', '…', 32) as snippet
			FROM documents_fts f
			JOIN documents d ON d.id = f.rowid
			WHERE documents_fts MATCH ?
			AND d.type = ?
			ORDER BY rank`
		args = []any{sanitized, docType}
	} else {
		sqlQuery = `SELECT d.id, d.type, d.title, bm25(documents_fts) as rank,
			snippet(documents_fts, 1, '»', '«', '…', 32) as snippet
			FROM documents_fts f
			JOIN documents d ON d.id = f.rowid
			WHERE documents_fts MATCH ?
			ORDER BY rank`
		args = []any{sanitized}
	}

	rows, err := s.db.QueryContext(context.Background(), sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("searching documents: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.Type, &r.Title, &r.Rank, &r.Snippet); err != nil {
			return nil, fmt.Errorf("scanning search result: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating search results: %w", err)
	}

	if results == nil {
		results = []SearchResult{}
	}
	return results, nil
}

// SanitizeQuery prepares user input for FTS5 MATCH.
// If the input contains explicit FTS5 operators, it passes through as-is.
// Otherwise, special characters are stripped to avoid parse errors.
func SanitizeQuery(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// Check for explicit FTS5 syntax — pass through for power users.
	if hasFTS5Syntax(input) {
		return input
	}

	// Strip characters that could cause FTS5 parse errors.
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case '"', '(', ')', '{', '}', ':':
			return -1
		case '*':
			return -1
		default:
			return r
		}
	}, input)

	return strings.TrimSpace(cleaned)
}

// hasFTS5Syntax checks if the query contains intentional FTS5 operators.
func hasFTS5Syntax(input string) bool {
	// Balanced quotes indicate a phrase query.
	if strings.Count(input, `"`)%2 == 0 && strings.Contains(input, `"`) {
		return true
	}

	// Check for FTS5 boolean operators (must be uppercase and surrounded by spaces or at boundaries).
	for _, op := range []string{" AND ", " OR ", " NOT ", "NOT "} {
		if strings.Contains(input, op) {
			return true
		}
	}

	// NEAR operator.
	if strings.Contains(input, " NEAR ") || strings.Contains(input, "NEAR(") {
		return true
	}

	// Prefix search (word ending in *).
	words := strings.Fields(input)
	for _, w := range words {
		if len(w) > 1 && strings.HasSuffix(w, "*") {
			return true
		}
	}

	return false
}
