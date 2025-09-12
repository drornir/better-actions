package workflows_test

import (
	"testing"

	"github.com/drornir/better-actions/pkg/runner"
	"github.com/drornir/better-actions/pkg/yamls"
)

func TestHelloWorkflow(t *testing.T) {
	ctx := makeContext(t, "file", "hello.yaml")
	f, err := rootFs.Open("hello.yaml")
	if err != nil {
		t.Fatal("failed to open workflow file", err)
	}
	wf, err := yamls.ReadWorkflow(f, false)
	if err != nil {
		t.Fatal("failed to read workflow", err)
	}

	if err := runner.RunWorkflow(ctx, wf); err != nil {
		t.Fatal("failed to run workflow", err)
	}
}
