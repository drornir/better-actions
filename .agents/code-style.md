# Better Action Code Style

## Naming Conventions

- Follow idiomatic Go: tabs for indentation, exported identifiers documented, and files named `kebab-case.go` with mirrored `kebab-case_test.go` companions.
- Always apply `gofumpt` (or `gofumpt -w .`) and organize imports with `goimports` before committing.

## Logging, Context, and Error Handling

- Extract dependencies once per function: `ctx context.Context`, then `logger := log.FromContext(ctx)` and `oopser := oops.FromContext(ctx)`.
- Pass the received context through call chains; never create `context.Background()` inside workflows or runners.
- Wrap returned errors with `oops` for context, prefer returning `(value, error)` over `log.Fatal`, and keep messages lowercase (e.g., `failed to parse`).
- Never use `fmt.Errorf` or `errors.New` directly; always wrap errors with `oops`, even if context is not injected.
- Always wrap errors that are coming from a 3rd party library with `oops` to provide context and ensure consistent error handling.
