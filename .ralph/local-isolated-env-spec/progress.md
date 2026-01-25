# Implementation Plan & Progress

## Plan

### 1. Update `Job` Struct (`pkg/runner/job.go`)

- [x] Add `WorkspaceDir string` to the `Job` struct.
- [x] This will store the absolute path to the temporary workspace directory created for the job.

### 2. Implement Workspace Creation (`pkg/runner/job.go`)

- [x] In `prepareJob`:
  - [x] Create a temporary directory using `os.MkdirTemp` (pattern: `bact-workspace-<jobName>-`).
  - [x] Assign this path to `j.WorkspaceDir`.
  - [x] Add a cleanup function to `os.RemoveAll` this directory (deferred until job completion).
  - [x] Set the `GITHUB_WORKSPACE` environment variable in `j.stepsEnv` (or `InitialEnv`) to this path.

### 3. Propagate Workspace to Steps (`pkg/runner/job.go`, `pkg/runner/step.go`)

- [x] Update `StepContext` struct in `pkg/runner/step.go` to include `WorkspaceDir string`.
- [x] Update `newStepContext` in `pkg/runner/job.go` to populate `WorkspaceDir` from the `Job`.

### 4. Execute Steps in Workspace (`pkg/runner/step-run.go`)

- [x] In `StepRun.Run`:
  - [x] Determine the working directory for the shell command.
  - [x] If `step.WorkingDirectory` is explicitly set in YAML, use it (resolved relative to workspace?). _Refinement: GitHub Actions defaults to GITHUB_WORKSPACE. If `working-directory` is relative, it's relative to GITHUB_WORKSPACE._
  - [x] If `step.WorkingDirectory` is empty, use `step.Context.WorkspaceDir`.
  - [x] Pass this directory to `shell.NewCommand` via `CommandOpts.Dir`.

### 5. Verification

- [x] Create a new test workflow `examples/workflows/isolation_test.yaml` (or similar) that:
  - [x] Writes a file to the current directory (`echo "hello" > artifact.txt`).

  - [x] Prints the current working directory (`pwd`).

- [x] Create a Go test `examples/workflows/isolation_test.go` that:
  - [x] Runs the workflow.

  - [x] Asserts that `artifact.txt` was created in the temp dir, **not** in the repo root.

  - [x] Asserts that `GITHUB_WORKSPACE` was set correctly.

- [ ] Create `examples/workflows/env_isolation_test.yaml` and `examples/workflows/env_isolation_test.go` to verify:
  - [ ] `GITHUB_ENV` updates persist across steps within a job.

  - [ ] `GITHUB_ENV` updates do NOT leak to other jobs.

  - [ ] `GITHUB_ENV` updates do NOT leak to the host process.

## Task List

- [x] Update `Job` struct

- [x] Implement `prepareJob` workspace creation

- [x] Update `StepContext`

- [x] Update `StepRun` execution logic

- [x] Add verification test

- [x] Add environment isolation verification tests
