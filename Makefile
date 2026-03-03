.PHONY: build test lint fmt vet check clean install-tools

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Build
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o kb .

# Test with race detector
test:
	go test -race -count=1 ./...

# Test with coverage
test-cover:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Lint (golangci-lint runs staticcheck, errcheck, govet, gosec, and more)
lint:
	golangci-lint run ./...

# Format (gofumpt is a strict superset of gofmt)
fmt:
	gofumpt -w -modpath kb .

# Format check (CI — fails if files need formatting)
fmt-check:
	@test -z "$$(gofumpt -l -modpath kb .)" || (echo "files need formatting:"; gofumpt -l -modpath kb .; exit 1)

# Go vet
vet:
	go vet ./...

# Run all checks (what CI would run)
check: fmt-check vet lint test

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest

# Clean build artifacts
clean:
	rm -f kb coverage.out
