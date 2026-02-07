package workflows_test

import (
	"bytes"
	"fmt"
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

func runWorkflowFromExample(t *testing.T, filename string, env []string) (string, *runner.WorkflowState, error) {
	t.Helper()

	ctx := makeContext(t, slog.LevelDebug, "file", filename)
	consoleBuffer := &bytes.Buffer{}
	console := io.MultiWriter(consoleBuffer, t.Output())
	run := runner.New(console, runner.EnvFromEnviron(env))

	f, err := rootFs.Open(filename)
	require.NoError(t, err, "failed to open workflow file")

	wf, err := yamls.ReadWorkflow(f, false)
	require.NoError(t, err, "failed to read workflow")

	wfState, err := run.RunWorkflow(ctx, wf, &types.WorkflowContexts{})
	return consoleBuffer.String(), wfState, err
}

func TestWorkflowCommands(t *testing.T) {
	const filename = "workflow-commands.yaml"
	output, wfState, err := runWorkflowFromExample(t, filename, []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		"MY_SPECIAL_ENV_VAR=my special value",
		"ACTIONS_ALLOW_UNSECURE_COMMANDS=true", // checks set-env etc.
	})
	if err != nil {
		t.Fatal("failed to run workflow:", errParse(err))
	}

	assert.Contains(t, output, "value is my special value")

	assert.Contains(t, output, "hello from custom_executable")

	assert.Contains(t, output, "WAS_SET_BY_INLINE_COMMAND=true")

	assert.Contains(t, output, "my secret is ***")
	assert.NotContains(t, output, "xx-VERY-SECRET-VALUE-xx")

	// TODO instead of inspecting the state of the job I want to print the value of the output using templating and print it
	//   waiting until I implement templating
	if assert.Contains(t, wfState.Jobs, "hello") {
		j := wfState.Jobs["hello"]
		outputs := j.StepOutputsCopy()
		if assert.Contains(t, outputs, "5_inline-setters") {
			stepOutputs := outputs["5_inline-setters"]
			if assert.Contains(t, stepOutputs, "was_set_by_inline_command") {
				assert.Equal(t, stepOutputs["was_set_by_inline_command"], "true")
			}
		}
	}
}

func TestWorkflowCommandsSetEnvRequiresOptIn(t *testing.T) {
	const filename = "workflow-commands.yaml"
	_, _, err := runWorkflowFromExample(t, filename, []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		"MY_SPECIAL_ENV_VAR=my special value",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "The set-env command is disabled")
}

func TestWorkflowCommandsAddPathRequiresOptIn(t *testing.T) {
	const filename = "workflow-commands-add-path.yaml"
	_, _, err := runWorkflowFromExample(t, filename, []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "The add-path command is disabled")
}

func TestWorkflowCommandsAddPathWithOptIn(t *testing.T) {
	const filename = "workflow-commands-add-path.yaml"
	output, _, err := runWorkflowFromExample(t, filename, []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		"ACTIONS_ALLOW_UNSECURE_COMMANDS=true",
	})
	require.NoError(t, err)
	assert.Contains(t, output, "legacy path tool")
}
