package xotel

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type tracerCtxKeyType struct{}

var tracerCtxKey = tracerCtxKeyType{}

// WithTracer attaches a tracer instance to a context.
func WithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerCtxKey, tracer)
}

// TracerFromContext returns the tracer from the context.
// Returns a noop tracer if none was previously attached.
func TracerFromContext(ctx context.Context) trace.Tracer {
	if tracer, ok := ctx.Value(tracerCtxKey).(trace.Tracer); ok {
		return tracer
	}
	return noop.NewTracerProvider().Tracer("")
}
