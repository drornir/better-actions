# Project: Local Semi-Isolated Environment

## Implementation Plan

### Phase 1: Core Workspace Isolation (Completed)

#### 1. Update `Job` Struct (`pkg/runner/job.go`)
- [x] Add `WorkspaceDir string` to the `Job` struct.
- [x] This stores the absolute path to the temporary workspace directory.

#### 2. Implement Workspace Creation (`pkg/runner/job.go`)
- [x] In `prepareJob`:
  - [x] Create temp directory (`bact-workspace-<jobName>-`).
  - [x] Assign to `j.WorkspaceDir`.
  - [x] Add cleanup (deferred `os.RemoveAll`).
  - [x] Set `GITHUB_WORKSPACE` in env.

#### 3. Propagate Workspace to Steps (`pkg/runner/job.go`, `pkg/runner/step.go`)
- [x] Update `StepContext` to include `WorkspaceDir`.
- [x] Update `newStepContext` to populate `WorkspaceDir`.

#### 4. Execute Steps in Workspace (`pkg/runner/step-run.go`)
- [x] In `StepRun.Run`:
  - [x] Default working directory to `s.Context.WorkspaceDir`.
  - [x] Resolve relative `working-directory` against `WorkspaceDir`.
  - [x] Pass directory to `shell.NewCommand`.

#### 5. Basic Verification
- [x] Create `examples/workflows/isolation_test.yaml` (write file, print pwd).
- [x] Create `examples/workflows/isolation_test.go` (assert file location, `GITHUB_WORKSPACE`).

### Phase 2: Environment Isolation (Pending)

#### 6. Verify Environment Variables (Intra-Job)
- [x] **Test: Persistence** (`examples/workflows/env_persistence.yaml`)
  - [x] Step 1: `echo "MY_VAR=persisted" >> $GITHUB_ENV`
  - [x] Step 2: Verify `$MY_VAR` is "persisted".
- [x] **Test: Path Modification**
  - [x] Step 1: Add directory to `$GITHUB_PATH`.
  - [x] Step 2: Verify binary in that directory is executable.

#### 7. Verify Job Isolation (Inter-Job & Future Proofing)
*Note: Even if jobs run sequentially now, these tests ensure future concurrency doesn't break isolation.*
- [ ] **Test: Environment Isolation** (`examples/workflows/job_isolation_env.yaml`)
  - [ ] Job A: Export `JOB_VAR=A`.
  - [ ] Job B: Verify `JOB_VAR` is unset/empty.
- [ ] **Test: File System Isolation** (`examples/workflows/job_isolation_fs.yaml`)
  - [ ] Job A: Create `workspace_file.txt` with content "Job A".
  - [ ] Job B: Verify `workspace_file.txt` does **not** exist.
  - [ ] Job B: Create `workspace_file.txt` with content "Job B" (ensure no collision).

#### 8. Verify Host Isolation
- [ ] **Test: Host Protection** (`examples/workflows/host_protection.yaml`)
  - [ ] Attempt to modify `PATH` or important env vars in a job.
  - [ ] Assert host process environment remains unchanged.

### Phase 3: Refinement & Safety (Pending)

#### 9. Directory Structure & Safety
- [ ] **Refactor: Validate `working-directory`**
  - [ ] In `StepRun.Run`, check if `s.Config.WorkingDirectory` is absolute.
  - [ ] If absolute, return an error.
  - [ ] **Test:** `examples/workflows/safety_abs_path.yaml` (should fail).

- [ ] **Refactor: Split `jobFilesRoot`**
  - [ ] In `prepareJob`, create `workspace/` and `steps/` subdirectories inside `jobFilesRoot`.
  - [ ] Point `j.WorkspaceDir` and `GITHUB_WORKSPACE` to `.../workspace`.
  - [ ] Update `newStepContext` to create step files inside `.../steps/<stepID>`.
  - [ ] **Test:** Verify `GITHUB_ENV` file path is NOT inside `GITHUB_WORKSPACE`.
