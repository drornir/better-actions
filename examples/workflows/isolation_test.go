package workflows_test

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/drornir/better-actions/pkg/runner"
	"github.com/drornir/better-actions/pkg/types"
	"github.com/drornir/better-actions/pkg/yamls"
)

func TestIsolationWorkflow(t *testing.T) {
	const filename = "isolation_test.yaml"
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

	// Capture CWD before run
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Run workflow
	if _, err := run.RunWorkflow(ctx, wf, &types.WorkflowContexts{}); err != nil {
		t.Fatal("failed to run workflow:", err)
	}

	// Verify output contains the "Workspace:" line
	output := consoleBuffer.String()
	assert.Contains(t, output, "Workspace: ")

	// Verify artifact.txt does NOT exist in CWD
	_, err = os.Stat(filepath.Join(cwd, "artifact.txt"))
	assert.True(t, os.IsNotExist(err), "artifact.txt should not exist in CWD")

	// Clean up if it failed and created it
	if err == nil {
		os.Remove(filepath.Join(cwd, "artifact.txt"))
	}
}
