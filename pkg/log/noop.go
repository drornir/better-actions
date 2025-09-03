package log

import (
	"context"
	"log/slog"
)

func NoopSLogger() *slog.Logger {
	return slog.New(noopSloggerHandler{})
}

type noopSloggerHandler struct{}

func (n noopSloggerHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return false
}

func (n noopSloggerHandler) Handle(ctx context.Context, r slog.Record) error {
	return nil
}

func (n noopSloggerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return n
}

func (n noopSloggerHandler) WithGroup(name string) slog.Handler {
	return n
}
