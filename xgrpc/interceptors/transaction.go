package interceptors

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
)

type Tx interface {
	With(ctx context.Context, fn func(context.Context) error) error
}

func UnaryPGXTransactionInterceptor(tx Tx, logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		logger.Debug("SQL transaction started", "method", info.FullMethod)

		var resp any

		err := tx.With(ctx, func(ctx context.Context) error {
			var hErr error
			resp, hErr = handler(ctx, req)
			return hErr
		})
		if err != nil {
			logger.Debug("SQL transaction rollback", "method", info.FullMethod, "error", err)
			return resp, err
		}

		logger.Debug("SQL transaction committed", "method", info.FullMethod)
		return resp, nil
	}
}
