# Local Semi-Isolated Environment Execution Spec

## Goal

Enable executing workflows in a local, semi-isolated environment without using Docker. The primary goal is **file system isolation**, providing each job with a dedicated workspace to perform file operations (e.g., git clones, file creation) without polluting the source repository or the user's home directory.

**Note:** The execution environment (the machine running `bact`) is assumed to be the trusted, desired environment. Jobs should inherit the tools and configuration of this host, but operate within their own file workspaces.

## Requirements

### 1. Filesystem Isolation

- Each job execution MUST have its own dedicated temporary directory (the **Job Root**).
- Inside the Job Root, there MUST be a dedicated **Workspace Directory**.
- The `GITHUB_WORKSPACE` environment variable MUST point to this **Workspace Directory**.
- Steps MUST execute within this Workspace by default.
- Steps that create files without an absolute path MUST create them inside the Workspace.
- The Job Root SHOULD be cleaned up after completion (unless in debug mode).

**Directory Structure:**

```text
/tmp/bact-job-<jobName>-<random>/  <-- Job Root (jobFilesRoot)
├── workspace/                      <-- GITHUB_WORKSPACE
│   └── (user files...)
└── steps/                          <-- Runner Metadata
    ├── step-0-setup/
    │   ├── GITHUB_ENV
    │   ├── GITHUB_PATH
    │   └── ...
    └── step-1-build/
        └── script.sh
```

**Constraints:**
- **Working Directory:** Steps MUST NOT use absolute paths for `working-directory` configuration. If an absolute path is provided, the runner MUST fail the step to ensure isolation. Relative paths in `working-directory` are resolved relative to `GITHUB_WORKSPACE`.

### 2. Environment Inheritance

- **Principle:** Jobs SHOULD inherit the environment of the host `bact` process. We assume the host is configured with the necessary tools and variables for the workflow.
- **Job Isolation:** While inheriting the host environment, changes made _during_ a job (e.g., adding to `GITHUB_ENV`, modifying `PATH` via workflow commands) MUST NOT leak to other parallel jobs or back to the host process.
- **Mechanism:**
  1.  Start with a copy of `os.Environ()`.
  2.  Apply job-specific variables (`GITHUB_WORKSPACE`, inputs, secrets).
  3.  Execute steps.
  4.  Discard the modified environment after the job finishes.
