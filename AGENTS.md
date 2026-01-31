# Agent Guidelines

## Code tools

run: `go run ./cmd/bact`
build: `go build -o bin/bact ./cmd/bact` - to check if it compiles
test: `go test ./...`
format: `goimports -local github.com/drornir/better-actions" -w <file> && gofumpt -extra -w <file>`, where <file> can be a `.`.
lint: `go vet ./...`, `staticcheck`

debug: use dlv

## Extra task specific guidelines

When writing code, consult `.agents/code-style.md`

When exploring the codebase, consult `.agents/codebase-structure.md`

## Other Notes

README.md usually won't have anything interesting for you.

## Workflow

Before committing, format, lint, build, test.
