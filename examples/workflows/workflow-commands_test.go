package workflows_test

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/drornir/better-actions/pkg/runner"
	"github.com/drornir/better-actions/pkg/yamls"
)

func TestWorkflowCommands(t *testing.T) {
	const filename = "workflow-commands.yaml"
	ctx := makeContext(t, slog.LevelDebug, "file", filename)

	consoleBuffer := &bytes.Buffer{}
	console := io.MultiWriter(consoleBuffer, t.Output())
	run := runner.New(console, runner.EnvFromEnviron([]string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		"MY_SPECIAL_ENV_VAR=my special value",
		"ACTIONS_ALLOW_UNSECURE_COMMANDS=true", // checks set-env etc.
	}))

	f, err := rootFs.Open(filename)
	if err != nil {
		t.Fatal("failed to open workflow file:", err)
	}
	wf, err := yamls.ReadWorkflow(f, false)
	if err != nil {
		t.Fatal("failed to read workflow:", err)
	}
	wfState, err := run.RunWorkflow(ctx, wf)
	if err != nil {
		t.Fatal("failed to run workflow:", errParse(err))
	}

	assert.Contains(t, consoleBuffer.String(), "value is my special value")

	assert.Contains(t, consoleBuffer.String(), "hello from custom_executable")

	assert.Contains(t, consoleBuffer.String(), "WAS_SET_BY_INLINE_COMMAND=true")

	assert.Contains(t, consoleBuffer.String(), "my secret is ***")
	assert.NotContains(t, consoleBuffer.String(), "xx-VERY-SECRET-VALUE-xx")

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
