package xhttp

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sptGabriel/gb-libs/xlog"
)

// WithLogger injects the logger into each request context so handlers and
// interceptors can retrieve it via xlog.FromContext
func WithLogger(logger xlog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(xlog.WithContext(r.Context(), logger))
			next.ServeHTTP(w, r)
		})
	}
}

// RequestData resolves the client address and the request id (X-Request-ID,
// generated when absent), echoes the id back in the response header and
// stores everything in the request context
func RequestData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, port := readUserIP(r)

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set("X-Request-ID", requestID)
		ctx := WithRequestData(r.Context(), RequestInfo{
			IP:        ip,
			Port:      port,
			RequestID: requestID,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Recover turns panics into a 500 response instead of killing the connection
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				xlog.ErrorContext(r.Context(), "panic serving request",
					"panic", fmt.Sprint(rec),
					"method", r.Method,
					"path", r.URL.Path,
				)
				sendJSON(w, InternalServerErrorPayload, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// AccessLog emits one log line per request with method, path, status,
// duration and request id
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).String(),
		}
		if data, ok := RequestDataFromContext(r.Context()); ok {
			attrs = append(attrs, "request_id", data.RequestID, "ip", data.IP)
		}

		xlog.InfoContext(r.Context(), "http request", attrs...)
	})
}

func readUserIP(r *http.Request) (string, string) {
	address := r.Header.Get("X-Real-Ip")
	if address == "" {
		address = r.Header.Get("X-Forwarded-For")
	}
	if address == "" {
		address = r.RemoteAddr
	}

	ip, port := "unknown", "unknown"
	if parts := strings.Split(address, ":"); len(parts) == 2 {
		ip, port = parts[0], parts[1]
	}
	return ip, port
}
