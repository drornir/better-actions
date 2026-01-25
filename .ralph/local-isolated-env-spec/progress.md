# Implementation Plan & Progress

## Plan

### 1. Update `Job` Struct (`pkg/runner/job.go`)
- [ ] Add `WorkspaceDir string` to the `Job` struct.
- [ ] This will store the absolute path to the temporary workspace directory created for the job.

### 2. Implement Workspace Creation (`pkg/runner/job.go`)
- [ ] In `prepareJob`:
    - [ ] Create a temporary directory using `os.MkdirTemp` (pattern: `bact-workspace-<jobName>-`).
    - [ ] Assign this path to `j.WorkspaceDir`.
    - [ ] Add a cleanup function to `os.RemoveAll` this directory (deferred until job completion).
    - [ ] Set the `GITHUB_WORKSPACE` environment variable in `j.stepsEnv` (or `InitialEnv`) to this path.

### 3. Propagate Workspace to Steps (`pkg/runner/job.go`, `pkg/runner/step.go`)
- [ ] Update `StepContext` struct in `pkg/runner/step.go` to include `WorkspaceDir string`.
- [ ] Update `newStepContext` in `pkg/runner/job.go` to populate `WorkspaceDir` from the `Job`.

### 4. Execute Steps in Workspace (`pkg/runner/step-run.go`)
- [ ] In `StepRun.Run`:
    - [ ] Determine the working directory for the shell command.
    - [ ] If `step.WorkingDirectory` is explicitly set in YAML, use it (resolved relative to workspace?). *Refinement: GitHub Actions defaults to GITHUB_WORKSPACE. If `working-directory` is relative, it's relative to GITHUB_WORKSPACE.*
    - [ ] If `step.WorkingDirectory` is empty, use `step.Context.WorkspaceDir`.
    - [ ] Pass this directory to `shell.NewCommand` via `CommandOpts.Dir`.

### 5. Verification
- [ ] Create a new test workflow `examples/workflows/isolation_test.yaml` (or similar) that:
    - [ ] Writes a file to the current directory (`echo "hello" > artifact.txt`).
    - [ ] Prints the current working directory (`pwd`).
- [ ] Create a Go test `examples/workflows/isolation_test.go` that:
    - [ ] Runs the workflow.
    - [ ] Asserts that `artifact.txt` was created in the temp dir, **not** in the repo root.
    - [ ] Asserts that `GITHUB_WORKSPACE` was set correctly.

## Task List
- [ ] Update `Job` struct
- [ ] Implement `prepareJob` workspace creation
- [ ] Update `StepContext`
- [ ] Update `StepRun` execution logic
- [ ] Add verification test
