package xhttp

import (
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"github.com/sptGabriel/gb-libs/xlog"
)

// Handle adapts a Response-returning function into an http.HandlerFunc,
// centralizing error logging, span recording and JSON encoding
func Handle(handler func(r *http.Request) Response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())

		response := handler(r)
		if response.Error != nil {
			xlog.ErrorContext(r.Context(), "application error",
				"error", response.Error.Error(),
			)
			span.RecordError(response.Error)
		}

		if h := response.Header(); h != nil {
			copyHeaders(w, h)
		}

		if err := sendJSON(w, response.Body, response.Status); err != nil {
			xlog.ErrorContext(r.Context(), "on send json", "error", err.Error())
			span.RecordError(err)
		}
	}
}

// ParseBody decodes the request body into T
func ParseBody[T any](r *http.Request) (T, error) {
	var target, zero T
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		return zero, err
	}
	return target, nil
}

func sendJSON(w http.ResponseWriter, payload any, statusCode int) error {
	if payload == nil {
		w.WriteHeader(statusCode)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(payload)
}

func copyHeaders(w http.ResponseWriter, h http.Header) {
	wh := w.Header()
	for header, values := range h {
		for _, value := range values {
			wh.Add(header, value)
		}
	}
}
