package workflows_test

import (
	"context"
	"embed"
	"log/slog"
	"testing"

	"github.com/drornir/better-actions/pkg/log"
)

//go:embed *.yaml
var rootFs embed.FS

func makeContext(t *testing.T, logAttrs ...any) context.Context {
	ctx := t.Context()
	ctx = log.New(slog.New(slog.NewTextHandler(
		t.Output(),
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	).With(logAttrs...).WithContext(ctx)

	return ctx
}
