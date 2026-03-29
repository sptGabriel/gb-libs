package interceptors

import (
	"context"
	"fmt"

	"github.com/sptGabriel/gb-libs/xotel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func UnaryOtelTracingInterceptor(tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx = xotel.ExtractOtelContextFromMetadata(ctx)

		ctx, span := tracer.Start(
			ctx,
			"grpc.method",
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(
				codes.Error,
				fmt.Sprintf("%s grpc method error", info.FullMethod),
			)
		}

		if span.IsRecording() {
			span.SetAttributes(
				attribute.String("method", info.FullMethod),
			)
			span.AddEvent("request concluded")
		}

		return resp, err
	}
}
