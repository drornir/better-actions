package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCommandKeyValueContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		expect    map[string]string
		expectErr string
	}{
		{
			name:    "simple variable",
			content: "FOO=bar\n",
			expect: map[string]string{
				"FOO": "bar",
			},
		},
		{
			name:    "heredoc value",
			content: "MULTI<<EOF\nhello\nworld\nEOF\n",
			expect: map[string]string{
				"MULTI": "hello\nworld",
			},
		},
		{
			name:    "heredoc value without newline at the end",
			content: "MULTI<<EOF\nhello\nworld\nEOF",
			expect: map[string]string{
				"MULTI": "hello\nworld",
			},
		},
		{
			name:      "missing newline before delimiter",
			content:   "BAD<<EOF\nvalue",
			expectErr: "invalid value: matching delimiter not found \"EOF\"",
		},
		{
			name:      "blocked variable",
			content:   "NODE_OPTIONS=value\n",
			expectErr: "can't store NODE_OPTIONS output parameter using '$GITHUB_ENV' command",
		},
		{
			name:      "invalid format",
			content:   "INVALID",
			expectErr: "invalid format \"INVALID\"",
		},
		{
			name:      "empty name",
			content:   "=value\n",
			expectErr: "invalid format \"=value\": name must not be empty",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pairs, err := parseCommandKeyValueContent(tc.content, GithubEnv)
			if tc.expectErr != "" {
				require.EqualError(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expect, pairs)
		})
	}
}

func TestParseEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, GithubEnv.FileName())

	require.NoError(t, os.WriteFile(path, []byte("FOO=bar\n"), 0o644))

	pairs, err := parseCommandKeyValueFile(path, GithubEnv)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"FOO": "bar"}, pairs)

	emptyPairs, err := parseCommandKeyValueFile(filepath.Join(dir, "missing.txt"), GithubEnv)
	require.NoError(t, err)
	require.Nil(t, emptyPairs)
}

func TestProcessWorkflowCommandFilesEnv(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, GithubEnv.FileName())
	require.NoError(t, os.WriteFile(path, []byte("FOO=bar\n"), 0o644))

	job := &Job{
		InitialEnv:    map[string]string{"EXISTING": "1"},
		stepsEnv:      make(map[string]string),
		stepsPath:     nil,
		stepOutputs:   make(map[string]map[string]string),
		stepStates:    make(map[string]map[string]string),
		stepSummaries: make(map[string]string),
	}
	stepCtx := &StepContext{
		Env:    map[string]string{GithubEnv.EnvVarName(): path},
		StepID: "0_test",
	}

	require.NoError(t, job.loadWFCmdFilesAfterStep(ctx, stepCtx))
	require.Equal(t, "1", job.InitialEnv["EXISTING"])
	require.Equal(t, "bar", job.stepsEnv["FOO"])
}
