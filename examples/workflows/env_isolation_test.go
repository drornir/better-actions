package workflows_test

import (
	"bytes"
	"io"
	"log/slog"
	"os"
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

	// Ensure the test variable is NOT in the host environment
	os.Unsetenv("JOB_VAR")

	run := runner.New(
		console,
		runner.EnvFromOS(), // We want to test that we inherit OS env, but don't leak back
	)

	f, err := rootFs.Open(filename)
	require.NoError(t, err, "failed to open workflow file")

	wf, err := yamls.ReadWorkflow(f, false)
	require.NoError(t, err, "failed to read workflow")

	// Run workflow
	_, err = run.RunWorkflow(ctx, wf, &types.WorkflowContexts{})
	require.NoError(t, err, "failed to run workflow")

	output := consoleBuffer.String()
	t.Log(output)

	// Verify persistence within job-persistence (implicitly verified by the shell script exit code)
	// Verify isolation from host
	_, exists := os.LookupEnv("JOB_VAR")
	assert.False(t, exists, "JOB_VAR should not leak to host process")
}
