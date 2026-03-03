# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development

```bash
make build          # Build binary (CGO_ENABLED=0)
make test           # Run tests with race detector
make test-cover     # Tests with coverage report
make lint           # golangci-lint (requires: make install-tools)
make fmt            # Format with gofumpt (requires: make install-tools)
make check          # Run all checks (fmt-check + vet + lint + test)
make install-tools  # Install golangci-lint and gofumpt
```

Run a single test:
```bash
go test -race -count=1 -run TestSearchName ./internal/store/
```

## Architecture

SQLite-backed CLI for storing and searching markdown documents with YAML frontmatter metadata.

### Two-layer structure

- **`cmd/`** — CLI commands using [Kong](https://github.com/alecthomas/kong). Each command is a struct with a `Run(*CLI)` method. The `CLI` struct in `root.go` defines the top-level flags (`--db`, `--plain`) and all subcommands.
- **`internal/store/`** — Data layer. `Store` wraps `*sql.DB` with methods for documents, search, and links. Schema is embedded via `go:embed` from `schema.sql`.

### Key patterns

- **Pure-Go SQLite** via `modernc.org/sqlite` — no CGO required, builds with `CGO_ENABLED=0`.
- **FTS5 full-text search** with BM25 ranking. FTS index kept in sync via SQL triggers (insert/update/delete).
- **Frontmatter parsing** with `github.com/adrg/frontmatter` — YAML/TOML/JSON frontmatter stored as opaque JSON in the `metadata` column.
- **Document linking** — many-to-many `document_links` table with optional relationship labels, cascade deletes.
- **Styled output** via Lipgloss/Glamour with `--plain` flag for plain text mode. `PlainOutput` is set before `kong.Parse` due to help printer firing during parse.
- **Goroutine leak detection** — both test packages use `go.uber.org/goleak` in `TestMain`.

### Testing

Tests use in-memory SQLite (`:memory:`) via the `newTestStore(t)` helper in `internal/store/store_test.go`. No test fixtures or external dependencies needed.

### Database

Default path: `kb.db` (override with `--db` flag or `KB_DB` env var). Schema uses `IF NOT EXISTS` for idempotent `kb init`.
