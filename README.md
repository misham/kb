# kb

A knowledge base CLI backed by SQLite. Store markdown documents with YAML frontmatter, search them with full-text search, and link them together.

## Features

- **Full-text search** — FTS5 with BM25 ranking and snippet extraction
- **Frontmatter metadata** — YAML/TOML/JSON frontmatter stored as queryable JSON
- **Document linking** — many-to-many links with optional relationship labels
- **Styled output** — colored terminal output with `--plain` fallback
- **No CGO** — pure-Go SQLite driver, single static binary

## Install

```bash
go install kb@latest
```

Or build from source:

```bash
make build    # produces ./kb
```

## Quick Start

```bash
# Create a new knowledge base
kb init

# Import markdown files
kb import docs/research/auth-flow.md -t research
kb import docs/plans/auth-flow.md -t plan

# Link related documents
kb link 1 2 -r pair

# Search
kb search goroutines
kb search sqlite -t research

# Browse
kb list
kb list -t plan
kb get 1
```

## Commands

| Command | Description | Example |
|---------|-------------|---------|
| `init` | Create a new knowledge base | `kb init` |
| `import <file> -t <type>` | Import a markdown file | `kb import notes.md -t research` |
| `search <query> [-t type]` | Full-text search | `kb search goroutines -t research` |
| `list [-t type]` | List documents | `kb list -t plan` |
| `get <id>` | Display a document | `kb get 1` |
| `delete <id> [-f]` | Delete a document | `kb delete 1 -f` |
| `link <id1> <id2> [-r rel]` | Link two documents | `kb link 1 2 -r related` |
| `links <id>` | Show linked documents | `kb links 1` |

## Configuration

| Option | Description | Default |
|--------|-------------|---------|
| `--db <path>` | Database file path | `kb.db` |
| `KB_DB` env var | Database file path (alternative) | `kb.db` |
| `--plain` | Disable styled output | off |

## Search Syntax

Plain search terms work out of the box. Power users can use FTS5 syntax:

```bash
kb search goroutines                  # simple term
kb search '"exact phrase"'            # phrase match
kb search 'sqlite AND performance'    # boolean AND
kb search 'redis OR memcached'        # boolean OR
kb search 'NOT deprecated'            # boolean NOT
kb search 'auth*'                     # prefix match
kb search 'foo NEAR bar'             # proximity search
```

## Bulk Import

```bash
for f in docs/research/*.md; do kb import "$f" -t research; done
for f in docs/plans/*.md; do kb import "$f" -t plan; done
```

## Development

```bash
make install-tools  # install golangci-lint and gofumpt
make test           # run tests with race detector
make lint           # run golangci-lint
make fmt            # format with gofumpt
make check          # run all checks (fmt + vet + lint + test)
```

Run a single test:

```bash
go test -race -count=1 -run TestSearchName ./internal/store/
```

## Ad-hoc Queries

The database is a standard SQLite file. Query it directly for anything the CLI doesn't cover:

```bash
sqlite3 kb.db "SELECT title FROM documents WHERE json_extract(metadata, '$.topic') = 'auth'"
sqlite3 kb.db "SELECT * FROM documents_fts WHERE documents_fts MATCH 'sqlite'"
```
