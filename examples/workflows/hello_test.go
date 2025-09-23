package workflows_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/drornir/better-actions/pkg/runner"
	"github.com/drornir/better-actions/pkg/yamls"
)

func TestHelloWorkflow(t *testing.T) {
	ctx := makeContext(t, "file", "hello.yaml")
	consoleBuffer := &bytes.Buffer{}
	console := io.MultiWriter(consoleBuffer, t.Output())
	run := &runner.Runner{
		Console: console,
	}

	f, err := rootFs.Open("hello.yaml")
	if err != nil {
		t.Fatal("failed to open workflow file", err)
	}
	wf, err := yamls.ReadWorkflow(f, false)
	if err != nil {
		t.Fatal("failed to read workflow", err)
	}

	if err := run.RunWorkflow(ctx, wf); err != nil {
		t.Fatal("failed to run workflow", err)
	}

	assert.Contains(t, consoleBuffer.String(), "Hello World")
}
