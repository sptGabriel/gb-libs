package xlog

import "context"

type loggerCtxKeyType struct{}

var loggerCtxKey = loggerCtxKeyType{}

// WithContext attaches a logger instance to a context
// Returns the original context unchanged if the provided logger is nil
// Creates a context chain that propagates the logger
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

// FromContext returns the logger from the context.
// Returns nil if no logger was previously attached
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerCtxKey).(Logger); ok {
		return logger
	}
	return nil
}
