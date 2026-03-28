package cqs

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/sptGabriel/defaults/xlog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func LoggingQueryInterceptor[Q any, R any](
	logger *slog.Logger,
	tracer trace.Tracer,
) QueryInterceptor[Q, R] {
	return func(ctx context.Context, q Q, next Next[Q, R]) (R, error) {
		var qType *Q
		qName := reflect.TypeOf(qType).Elem().Name()
		logger := xlog.FromContext(ctx)
		logger.Debug("query handling started", "name", qName)

		ctx, span := tracer.Start(ctx, "query-handler")
		defer span.End()

		if span.IsRecording() {
			span.SetAttributes(attribute.String("query.name", qName))
		}

		out, err := next(ctx, q)
		if err != nil {
			logger.Error("query handling failed", "name", qName, "error", err)

			span.RecordError(err)
			span.SetStatus(codes.Error, fmt.Sprintf("%s handler error", qName))

			var zero R
			return zero, err
		}

		logger.Debug("query handling finished", "name", qName)

		return out, nil
	}
}
