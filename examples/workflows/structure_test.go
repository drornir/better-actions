package workflows_test

import (
	"bytes"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/drornir/better-actions/pkg/runner"
	"github.com/drornir/better-actions/pkg/types"
	"github.com/drornir/better-actions/pkg/yamls"
)

func TestStructureWorkflow(t *testing.T) {
	const filename = "structure_test.yaml"
	ctx := makeContext(t, slog.LevelDebug, "file", filename)
	consoleBuffer := &bytes.Buffer{}
	console := io.MultiWriter(consoleBuffer, t.Output())
	run := runner.New(
		console,
		runner.EnvFromEmpty(),
	)

	f, err := rootFs.Open(filename)
	if err != nil {
		t.Fatal("failed to open workflow file:", err)
	}
	wf, err := yamls.ReadWorkflow(f, false)
	if err != nil {
		t.Fatal("failed to read workflow:", err)
	}

	// Run workflow
	wfState, err := run.RunWorkflow(ctx, wf, &types.WorkflowContexts{})
	if err != nil {
		t.Fatal("failed to run workflow:", err)
	}

	job, ok := wfState.Jobs["check-locations"]
	require.True(t, ok, "job check-locations not found")

	workspaceDir := job.WorkspaceDir
	assert.NotEmpty(t, workspaceDir, "WorkspaceDir should not be empty")

	stepOutputs := job.StepOutputsCopy()
	// Step ID format: index_slug
	// Step has explicit id: "expose" -> slug: "expose"
	// Index: 0
	outputs, ok := stepOutputs["0_expose"]
	if !ok {
		// Fallback debugging if slug generation differs
		var keys []string
		for k := range stepOutputs {
			keys = append(keys, k)
		}
		t.Fatalf("Step output not found. Available keys: %v", keys)
	}

	githubEnv := outputs["ENV_PATH"]
	assert.NotEmpty(t, githubEnv, "GITHUB_ENV path should not be empty")

	// Check 1: WorkspaceDir should end with "workspace"
	assert.Equal(t, "workspace", filepath.Base(workspaceDir), "WorkspaceDir should be named 'workspace'")

	// Check 2: GITHUB_ENV should be inside a 'steps' directory
	// Structure: <root>/steps/<stepID>/GITHUB_ENV
	assert.Contains(t, githubEnv, "/steps/", "GITHUB_ENV should be inside a 'steps' directory")

	// Check 3: GITHUB_ENV is NOT inside GITHUB_WORKSPACE
	assert.False(t, strings.HasPrefix(githubEnv, workspaceDir),
		"GITHUB_ENV (%s) should NOT be inside WorkspaceDir (%s)", githubEnv, workspaceDir)

	// Check 4: They share a common root
	jobRootFromWorkspace := filepath.Dir(workspaceDir)

	stepsIndex := strings.LastIndex(githubEnv, "/steps/")
	require.NotEqual(t, -1, stepsIndex, "steps directory not found in GITHUB_ENV path")

	jobRootFromEnv := githubEnv[:stepsIndex]

	assert.Equal(t, jobRootFromWorkspace, jobRootFromEnv, "Workspace and Steps should share the same job root")
}
