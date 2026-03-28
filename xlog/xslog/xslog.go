package xslog

import (
	"context"
	"log/slog"

	"github.com/sptGabriel/defaults/xlog"
)

type Logger struct {
	logger *slog.Logger
}

func (sl *Logger) With(attrs ...any) xlog.Logger {
	newLogger := sl.logger.With(attrs...)
	return NewSlog(newLogger)
}

func (sl *Logger) Level(level xlog.Level, msg string, attrs ...any) {
	slogLevel := toSlogLevel(level)
	if !sl.logger.Enabled(context.Background(), slogLevel) {
		return
	}

	sl.logger.Log(context.Background(), slogLevel, msg, attrs...)
}

func (sl *Logger) Info(msg string, attrs ...any) {
	sl.logger.Info(msg, attrs...)
}

func (sl *Logger) Error(msg string, attrs ...any) {
	sl.logger.Error(msg, attrs...)
}

func (sl *Logger) Debug(msg string, attrs ...any) {
	sl.logger.Debug(msg, attrs...)
}

// Funções auxiliares
func toSlogLevel(level xlog.Level) slog.Level {
	switch level {
	case xlog.LevelDebug:
		return slog.LevelDebug
	case xlog.LevelInfo:
		return slog.LevelInfo
	case xlog.LevelWarn:
		return slog.LevelWarn
	case xlog.LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func NewSlog(logger *slog.Logger) *Logger {
	return &Logger{logger: logger}
}
