package interceptors

import (
	"context"
	"log/slog"

	"github.com/sptGabriel/gb-libs/cqs"
	"google.golang.org/grpc"
)

// Tx is the same contract the command interceptor consumes — same one method,
// same meaning — so it is taken from cqs instead of declared a third time.
type Tx = cqs.Tx

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
