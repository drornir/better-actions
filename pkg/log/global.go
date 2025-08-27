package log

import (
	"context"
	"sync"
)

var (
	globalLogger     *Logger
	globalLoggerLock sync.RWMutex
)

func init() {
	globalLogger = New(NoopSlogger)
}

func GG() *Logger {
	return GetGlobal()
}

func GetGlobal() *Logger {
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()
	return globalLogger
}

func SetGlobal(l *Logger) {
	globalLoggerLock.Lock()
	defer globalLoggerLock.Unlock()
	globalLogger = l

	l.T(context.Background(), "log level set to 'trace'")
}

// T logs a trace message using the global logger
func T(ctx context.Context, msg string, a ...any) {
	GetGlobal().T(ctx, msg, a...)
}

// D logs a debug message using the global logger
func D(ctx context.Context, msg string, a ...any) {
	GetGlobal().D(ctx, msg, a...)
}

// I logs an info message using the global logger
func I(ctx context.Context, msg string, a ...any) {
	GetGlobal().I(ctx, msg, a...)
}

// W logs a warning message using the global logger
func W(ctx context.Context, msg string, a ...any) {
	GetGlobal().W(ctx, msg, a...)
}

// E logs an error message using the global logger
func E(ctx context.Context, msg string, a ...any) {
	GetGlobal().E(ctx, msg, a...)
}
