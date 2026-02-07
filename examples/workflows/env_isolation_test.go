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

func TestEnvIsolationWorkflow(t *testing.T) {
	const filename = "env_isolation_test.yaml"
	ctx := makeContext(t, slog.LevelDebug, "file", filename)
	consoleBuffer := &bytes.Buffer{}
	console := io.MultiWriter(consoleBuffer, t.Output())

	run := runner.New(
		console,
		runner.EnvFromEmpty(),
	)

	f, err := rootFs.Open(filename)
	require.NoError(t, err, "failed to open workflow file")

	wf, err := yamls.ReadWorkflow(f, false)
	require.NoError(t, err, "failed to read workflow")

	// Run workflow
	_, err = run.RunWorkflow(ctx, wf, &types.WorkflowContexts{})
	require.NoError(t, err, "failed to run workflow")

	output := consoleBuffer.String()
	assert.Contains(t, output, "same job env persistence ok")
	assert.Contains(t, output, "cross-job env isolation ok")
}
