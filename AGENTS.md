# Repository Guidelines

## Project Structure & Module Organization

- `cmd/bact` exposes the CLI entrypoint (`main.go`) and thin workflow wiring; new commands should delegate to packages.
- `pkg/` holds reusable modules: `config` loads `bact.yaml`, `runner` and `shell` execute steps, `yamls` validates workflow YAML, and `log` centralizes structured logging.
- `examples/workflows/` contains runnable YAML samples with table-driven tests—extend these when changing workflow semantics.
- `scripts/dev.sh` runs an end-to-end workflow against `examples/workflows/hello.yaml`; use it for quick smoke checks.
- `~/code/forks/runner` mirrors the official GitHub runner with local analysis notes for reference. ALWAYS READ ITS AGENTS.md BEFORE USING IT.

## CLI Command Safety

- Treat the surrounding system as production: do not run commands that could mutate external services (Kubernetes clusters, AWS resources, GitHub CLI state, etc.) unless explicitly cleared by the user.

## Build, Test, and Development Commands

- `go build ./cmd/bact` compiles the CLI.
- `go run ./cmd/bact workflow run -f examples/workflows/hello.yaml` exercises the default workflow; swap the file path to reproduce issues.
- `go test ./...` runs all unit and integration tests; add `-cover` when auditing new logic.
- `scripts/dev.sh` wraps the run command with guardrails and logging for manual verification.

## Coding Style & Naming Conventions

- Follow idiomatic Go: tabs for indentation, exported identifiers documented, and files named `snake_case.go` with mirrored `_test.go` companions.
- Always apply `gofumpt` (or `gofumpt -w .`) and organize imports with `goimports` before committing.

## Logging, Context, and Error Handling

- Extract dependencies once per function: `ctx context.Context`, then `logger := log.FromContext(ctx)` and `oopser := oops.FromContext(ctx)`.
- Pass the received context through call chains; never create `context.Background()` inside workflows or runners.
- Wrap returned errors with `oops` for context, prefer returning `(value, error)` over `log.Fatal`, and keep messages lowercase (e.g., `failed to parse`).
- Never use `fmt.Errorf` or `errors.New` directly; always wrap errors with `oops`, even if context is not injected.
- Always wrap errors that are coming from a 3rd party library with `oops` to provide context and ensure consistent error handling.

## Testing Guidelines

- Co-locate tests with implementation under `pkg/` or alongside workflow fixtures in `examples/workflows/`.
- Use subtests (`t.Run("scenario", ...)`) and `testify` assertions to capture intent.
- Extend example workflows when adding features so automated and manual checks share coverage.
- Always use table testing, even if you only have one case. This will allow easier expansion
  of the test later.

## Commit & Pull Request Guidelines

- Write short, imperative commit subjects (`create workflow command files`).
- Ensure `go test ./...`, formatting, and linting. See settings for linter in `.zed/settings.json` under gopls.

## Security & Configuration Tips

- Keep secrets out of `bact.yaml`; rely on environment lookups handled in `pkg/config`.
- The local replace `github.com/drornir/factor3 => ../../factor3` expects the sibling repo—update it during development and drop the replace before publishing modules.
