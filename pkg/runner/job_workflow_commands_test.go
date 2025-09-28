package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessWorkflowCommandFilesAll(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	envPath := filepath.Join(dir, GithubEnv.FileName())
	require.NoError(t, os.WriteFile(envPath, []byte("FOO=bar\n"), 0o644))

	pathPath := filepath.Join(dir, GithubPath.FileName())
	require.NoError(t, os.WriteFile(pathPath, []byte("/tmp/bin\n"), 0o644))

	outputPath := filepath.Join(dir, GithubOutput.FileName())
	require.NoError(t, os.WriteFile(outputPath, []byte("RESULT=42\n"), 0o644))

	statePath := filepath.Join(dir, GithubState.FileName())
	require.NoError(t, os.WriteFile(statePath, []byte("KEY=value\n"), 0o644))

	summaryPath := filepath.Join(dir, GithubStepSummary.FileName())
	require.NoError(t, os.WriteFile(summaryPath, []byte("## summary\n"), 0o644))

	job := &Job{
		RunnerEnv:     map[string]string{"PATH": "/usr/bin"},
		stepsEnv:      make(map[string]string),
		stepsPath:     nil,
		stepOutputs:   make(map[string]map[string]string),
		stepStates:    make(map[string]map[string]string),
		stepSummaries: make(map[string]string),
	}

	stepCtx := &StepContext{
		Env: map[string]string{
			GithubEnv.EnvVarName():         envPath,
			GithubPath.EnvVarName():        pathPath,
			GithubOutput.EnvVarName():      outputPath,
			GithubState.EnvVarName():       statePath,
			GithubStepSummary.EnvVarName(): summaryPath,
		},
		StepID: "0_test",
	}

	require.NoError(t, job.processWorkflowCommandFiles(ctx, stepCtx))

	require.Equal(t, "bar", job.stepsEnv["FOO"])
	require.Contains(t, job.stepsPath, "/tmp/bin")

	outputs := job.stepOutputs["0_test"]
	require.NotNil(t, outputs)
	require.Equal(t, "42", outputs["RESULT"])

	state := job.stepStates["0_test"]
	require.NotNil(t, state)
	require.Equal(t, "value", state["KEY"])

	summary, ok := job.stepSummaries["0_test"]
	require.True(t, ok)
	require.Equal(t, "## summary\n", summary)
}
