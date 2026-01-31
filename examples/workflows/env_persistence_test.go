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

func TestEnvPersistenceWorkflow(t *testing.T) {
	const filename = "env_persistence.yaml"
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
	_, err = run.RunWorkflow(ctx, wf, &types.WorkflowContexts{})
	// Check for errors during execution
	if err != nil {
		t.Logf("Console Output:\n%s", consoleBuffer.String())
		t.Fatal("failed to run workflow:", err)
	}

	output := consoleBuffer.String()

	// Assertions based on expected output from the yaml steps
	assert.Contains(t, output, "MY_VAR is correctly set to persisted")
	assert.Contains(t, output, "my-tool executed successfully")
}
