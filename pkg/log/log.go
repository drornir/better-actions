package log

import (
	"context"
	"log/slog"
)

const SlogLevelTrace slog.Level = slog.LevelDebug - 4

func New(slogger *slog.Logger) *Logger {
	return &Logger{
		sl: slogger,
	}
}

type Logger struct {
	sl *slog.Logger
}

func (l *Logger) T(ctx context.Context, msg string, a ...any) {
	l.log(ctx, SlogLevelTrace, msg, a...)
}

func (l *Logger) D(ctx context.Context, msg string, a ...any) {
	l.log(ctx, slog.LevelDebug, msg, a...)
}

func (l *Logger) I(ctx context.Context, msg string, a ...any) {
	l.log(ctx, slog.LevelInfo, msg, a...)
}

func (l *Logger) W(ctx context.Context, msg string, a ...any) {
	l.log(ctx, slog.LevelWarn, msg, a...)
}

func (l *Logger) E(ctx context.Context, msg string, a ...any) {
	l.log(ctx, slog.LevelError, msg, a...)
}

func (l *Logger) Slogger() *slog.Logger {
	return l.sl
}

func (l *Logger) With(args ...any) *Logger {
	return New(l.Slogger().With(args...))
}

func (l *Logger) WithGroup(name string) *Logger {
	return New(l.Slogger().WithGroup(name))
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, a ...any) {
	l.sl.Log(ctx, level, msg, a...)
}
