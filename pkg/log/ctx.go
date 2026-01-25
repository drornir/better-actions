package log

import "context"

type ctxkey string

var ctxval ctxkey = "logger"

func (l *Logger) ContextWithLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxval, l)
}

func FromContext(ctx context.Context) *Logger {
	l := ctx.Value(ctxval)
	if l == nil {
		return GetGlobal()
	}
	return l.(*Logger)
}
