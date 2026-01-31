# Project Structure & Module Organization

This codebase structure is always evolving.

- `cmd/bact` exposes the CLI entrypoint (`main.go`) and thin workflow wiring; commands should be lean and delegate to packages.
- `pkg/` holds reusable modules: `config` loads `bact.yaml`, `runner` and `shell` execute steps, `yamls` load and validate workflow YAML, and `log` centralizes structured logging.
- `examples/workflows/` contains runnable YAML samples with table-driven tests.
- `scripts/dev.sh` runs an end-to-end workflow against `examples/workflows/hello.yaml`; use it for quick smoke checks.
- `~/code/forks/runner` mirrors the official GitHub runner with local analysis notes for reference.
