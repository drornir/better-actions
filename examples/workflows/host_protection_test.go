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

func TestHostProtectionWorkflow(t *testing.T) {
	cases := []struct {
		name     string
		filename string
	}{
		{
			name:     "path updates are scoped to a job",
			filename: "host_protection.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := makeContext(t, slog.LevelDebug, "file", tc.filename)
			consoleBuffer := &bytes.Buffer{}
			console := io.MultiWriter(consoleBuffer, t.Output())

			run := runner.New(
				console,
				runner.EnvFromEmptyWithBasicPath(),
			)

			f, err := rootFs.Open(tc.filename)
			require.NoError(t, err, "failed to open workflow file")

			wf, err := yamls.ReadWorkflow(f, false)
			require.NoError(t, err, "failed to read workflow")

			_, err = run.RunWorkflow(ctx, wf, &types.WorkflowContexts{})
			require.NoError(t, err, "failed to run workflow")

			output := consoleBuffer.String()
			assert.Contains(t, output, "same job path propagation ok")
			assert.Contains(t, output, "cross-job path isolation ok")
		})
	}
}
