package xlog

import (
	"context"
)

// Level defines the severity classification for logging events
type Level int16

const (
	// LevelDebug captures detailed diagnostic information
	LevelDebug Level = iota + 1
	// LevelInfo tracks standard operational events
	LevelInfo
	// LevelWarn indicates potential issues worth monitoring
	LevelWarn
	// LevelError indicates failures requiring attention
	LevelError
)

// Logger provides structured logging capabilities with multiple severity levels
// and contextual attribute support to enhance log entries with additional metadata
type Logger interface {
	// LogLevel emits a log entry with specified severity level and optional metadata
	Level(level Level, msg string, attrs ...any)
	// Info emits standard operational information
	Info(msg string, attrs ...any)
	// Error emits failure notifications and exceptional conditions
	Error(msg string, attrs ...any)
	// Debug emits detailed diagnostic information for troubleshooting
	Debug(msg string, attrs ...any)
	// With creates a derived logger containing permanent attributes
	With(attrs ...any) Logger
}

// LogLevelContext retrieves the logger from context and emits a message
// at the specified severity level with the given attributes
func LogLevelContext(ctx context.Context, level Level, msg string, attrs ...any) {
	logger := FromContext(ctx)
	if logger == nil {
		return
	}
	logger.Level(level, msg, attrs...)
}

// WarnContext emits failure notifications using the context-bound logger
// The operation is silently skipped if no logger exists in the provided context
func WarnContext(ctx context.Context, msg string, attrs ...any) {
	LogLevelContext(ctx, LevelWarn, msg, attrs...)
}

// InfoContext emits standard operational information using the context-bound logger
// The operation is silently skipped if no logger exists in the provided context
func InfoContext(ctx context.Context, msg string, attrs ...any) {
	LogLevelContext(ctx, LevelInfo, msg, attrs...)
}

// ErrorContext emits failure notifications using the context-bound logger
// The operation is silently skipped if no logger exists in the provided context
func ErrorContext(ctx context.Context, msg string, attrs ...any) {
	LogLevelContext(ctx, LevelError, msg, attrs...)
}

// DebugContext emits detailed diagnostic information using the context-bound logger
// The operation is silently skipped if no logger exists in the provided context
func DebugContext(ctx context.Context, msg string, attrs ...any) {
	LogLevelContext(ctx, LevelDebug, msg, attrs...)
}

// SetAttributes enriches the context with additional logging metadata
// Returns the original context unchanged if no logger is present
func SetAttributes(ctx context.Context, attrs ...any) context.Context {
	logger := FromContext(ctx)
	if logger == nil {
		return ctx
	}

	return WithContext(ctx, logger.With(attrs...))
}
