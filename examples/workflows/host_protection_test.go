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

func TestHostProtectionWorkflow(t *testing.T) {
	cases := []struct {
		name     string
		filename string
	}{
		{
			name:     "host env remains unchanged",
			filename: "host_protection.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := makeContext(t, slog.LevelDebug, "file", tc.filename)
			consoleBuffer := &bytes.Buffer{}
			console := io.MultiWriter(consoleBuffer, t.Output())
			originalPath := os.Getenv("PATH")

			run := runner.New(
				console,
				runner.EnvFromOS(),
			)

			f, err := rootFs.Open(tc.filename)
			require.NoError(t, err, "failed to open workflow file")

			wf, err := yamls.ReadWorkflow(f, false)
			require.NoError(t, err, "failed to read workflow")

			_, err = run.RunWorkflow(ctx, wf, &types.WorkflowContexts{})
			require.NoError(t, err, "failed to run workflow")

			output := consoleBuffer.String()
			assert.Contains(t, output, "PATH updated in job")
			assert.Equal(t, originalPath, os.Getenv("PATH"), "PATH should not change in host process")
		})
	}
}
