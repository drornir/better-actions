package log

import (
	"io"
	"log/slog"
)

type LoggerOptions struct {
	Level  string
	Writer io.Writer
	Format string
}

func MakeSLogger(opts LoggerOptions) *slog.Logger {
	var handler slog.Handler
	hopts := slog.HandlerOptions{
		Level:       ParseLevel(opts.Level),
		ReplaceAttr: SlogReplacerMinimal(),
	}
	if opts.Format == "json" {
		handler = slog.NewJSONHandler(opts.Writer, &hopts)
	} else {
		handler = slog.NewTextHandler(opts.Writer, &hopts)
	}

	return slog.New(handler)
}
