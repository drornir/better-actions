package log

import (
	"context"
	"log/slog"
	"strings"
)

// ParseLevel is a more forgiving version for parsing a string into an slog.Level
func ParseLevel(s string) slog.Level {
	s = strings.TrimSpace(s)
	if s == "" {
		s = "info"
	}
	var logLevel slog.Level
	switch strings.ToLower(s)[0] {
	case 't':
		logLevel = SlogLevelTrace
	case 'd':
		logLevel = slog.LevelDebug
	case 'i':
		logLevel = slog.LevelInfo
	case 'w':
		logLevel = slog.LevelWarn
	case 'e':
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return logLevel
}

func SloggerWithLevel(old *slog.Logger, level slog.Level) *slog.Logger {
	return slog.New(&levelRestricterHandler{
		old:   old.Handler(),
		level: level,
	})
}

type levelRestricterHandler struct {
	old   slog.Handler
	level slog.Level
}

func (h *levelRestricterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level && h.old.Enabled(ctx, level)
}

func (h *levelRestricterHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= h.level {
		return h.old.Handle(ctx, r)
	}
	return nil
}

func (h *levelRestricterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelRestricterHandler{
		old:   h.old.WithAttrs(attrs),
		level: h.level,
	}
}

func (h *levelRestricterHandler) WithGroup(name string) slog.Handler {
	return &levelRestricterHandler{
		old:   h.old.WithGroup(name),
		level: h.level,
	}
}
