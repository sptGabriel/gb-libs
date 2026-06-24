package cqs

import (
	"context"
	"fmt"

	"github.com/sptGabriel/gb-libs/xlog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func LoggingCommandInterceptor[C Command](tracer trace.Tracer) CommandInterceptor[C] {
	return func(ctx context.Context, cmd C, next CommandNext[C]) error {
		cmdName := typeName[C]()

		xlog.DebugContext(
			ctx,
			"command handling started.",
			"name", cmdName,
			"values", cmd,
		)

		ctx, span := tracer.Start(ctx, "command-handler")
		defer span.End()

		if span.IsRecording() {
			span.SetAttributes(
				attribute.String("command.name", cmdName),
			)
		}

		err := next(ctx, cmd)
		if err != nil {
			xlog.ErrorContext(
				ctx,
				"command handling failed.",
				"error", err,
				"name", cmdName,
			)

			span.RecordError(err)
			span.SetStatus(codes.Error, fmt.Sprintf("%s handler error", cmdName))
			return err
		}

		xlog.DebugContext(ctx, "command handling finished.", "name", cmdName)
		return nil
	}
}
