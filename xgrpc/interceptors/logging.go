package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
)

func UnaryLoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		logger.Debug("request started", "method", info.FullMethod)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			logger.Error("request failed",
				"method", info.FullMethod,
				"request", req,
				"error", err,
				"duration", duration,
			)
		} else {
			logger.Debug("request succeeded",
				"method", info.FullMethod,
				"duration", duration,
			)
		}

		return resp, err
	}
}
