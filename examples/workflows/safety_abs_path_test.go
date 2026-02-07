package workflows_test

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/drornir/better-actions/pkg/runner"
	"github.com/drornir/better-actions/pkg/types"
	"github.com/drornir/better-actions/pkg/yamls"
)

func TestWorkingDirectoryParityWorkflow(t *testing.T) {
	const filename = "safety_abs_path.yaml"
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

	_, err = run.RunWorkflow(ctx, wf, &types.WorkflowContexts{})
	if err != nil {
		t.Fatal("failed to run workflow:", err)
	}
	assert.Contains(t, consoleBuffer.String(), "absolute working-directory ok")
	assert.Contains(t, consoleBuffer.String(), "relative working-directory ok")
}
