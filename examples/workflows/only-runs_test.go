package workflows_test

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/drornir/better-actions/pkg/log"
	"github.com/drornir/better-actions/pkg/runner"
	"github.com/drornir/better-actions/pkg/yamls"
)

func TestOnlyRuns(t *testing.T) {
	const filename = "only-runs.yaml"
	ctx := makeContext(t, "file", filename)
	logger := log.FromContext(ctx).WithLevel(slog.LevelInfo)
	ctx = logger.WithContext(ctx)

	consoleBuffer := &bytes.Buffer{}
	console := io.MultiWriter(consoleBuffer, t.Output())
	run := runner.New(console, runner.EnvFromEnviron([]string{
		"MY_SPECIAL_ENV_VAR=my special value",
	}))

	f, err := rootFs.Open(filename)
	if err != nil {
		t.Fatal("failed to open workflow file:", err)
	}
	wf, err := yamls.ReadWorkflow(f, false)
	if err != nil {
		t.Fatal("failed to read workflow:", err)
	}

	if err := run.RunWorkflow(ctx, wf); err != nil {
		t.Fatal("failed to run workflow:", errParse(err))
	}

	assert.Contains(t, consoleBuffer.String(), "value is my special value")
}
