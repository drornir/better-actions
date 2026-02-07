package workflows_test

import (
	"bytes"
	"io"
	"log/slog"
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

	job, ok := wfState.Jobs["env-and-output"]
	require.True(t, ok, "job env-and-output not found")

	stepOutputs := job.StepOutputsCopy()
	outputs, ok := stepOutputs["0_write"]
	if !ok {
		var keys []string
		for k := range stepOutputs {
			keys = append(keys, k)
		}
		t.Fatalf("Step output not found. Available keys: %v", keys)
	}

	assert.Equal(t, "from-output-file", outputs["from_output_file"])
	assert.Contains(t, consoleBuffer.String(), "env propagation ok")
}
