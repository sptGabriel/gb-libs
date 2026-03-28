package errs

import (
	"strings"

	"github.com/sptGabriel/defaults/iterables"
)

type validationErrMessage struct {
	Message string
	Context map[string]any
}

type ValidationErr struct {
	Messages []validationErrMessage
}

func (e ValidationErr) Error() string {
	return strings.Join(
		iterables.Map(e.Messages, func(m validationErrMessage) string { return m.Message }), "\n",
	)
}

func (e *ValidationErr) With(msg string) {
	e.Messages = append(e.Messages, validationErrMessage{Message: msg})
}

func (e *ValidationErr) WithContext(msg string, ctx map[string]any) {
	e.Messages = append(e.Messages, validationErrMessage{Message: msg, Context: ctx})
}

func (e ValidationErr) Empty() bool {
	return len(e.Messages) == 0
}
