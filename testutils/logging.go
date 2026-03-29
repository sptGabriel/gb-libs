package testutils

import (
	"bytes"
	"log/slog"
	"strings"

	"github.com/sptGabriel/gb-libs/xlog"
	"github.com/sptGabriel/gb-libs/xlog/xslog"
)

type TestLogger struct {
	Logger xlog.Logger
	Buffer *bytes.Buffer
}

func (l *TestLogger) String() string {
	return l.Buffer.String()
}

func (l *TestLogger) LogLines() []string {
	return strings.Split(strings.TrimSpace(l.Buffer.String()), "\n")
}

func (l *TestLogger) Reset() {
	l.Buffer.Reset()
}

func NewTestLogger() *TestLogger {
	buf := &bytes.Buffer{}
	logger := xslog.NewSlog(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	})))

	return &TestLogger{
		Logger: logger,
		Buffer: buf,
	}
}
