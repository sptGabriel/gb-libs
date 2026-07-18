package xpg

import (
	"context"

	"github.com/sptGabriel/gb-libs/cqs"
)

// TransactionCommandInterceptor wraps the command handler execution in a
// database transaction. If no querier exists in the context, the
// transactioner falls back to its internal pool connection automatically.
func TransactionCommandInterceptor[C cqs.Command](tx Tx) cqs.CommandInterceptor[C] {
	return func(ctx context.Context, cmd C, next cqs.CommandNext[C]) error {
		return tx.With(ctx, func(txCtx context.Context) error {
			return next(txCtx, cmd)
		})
	}
}
