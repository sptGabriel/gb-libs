package cqs

import "context"

// Tx runs fn inside a transaction, committing when fn returns nil and rolling
// back when it returns an error.
//
// It lives here, and not in the database packages, because nothing about it is
// database-specific: xpg.Tx and the port an application declares for itself are
// the same one method. Go interfaces are structural, so any of those satisfies
// this one without importing it — what this declaration buys is a single home
// for the interceptor below.
type Tx interface {
	With(ctx context.Context, fn func(context.Context) error) error
}

// TransactionCommandInterceptor wraps the command handler execution in a
// transaction: it commits when the handler returns nil and rolls back when it
// returns an error.
//
// It is the third of this package's interceptors, next to logging and
// validation, and belongs with them for the same reason: the body knows only
// about commands. Which database is underneath is the caller's business, and
// arrives as the Tx it passes in.
func TransactionCommandInterceptor[C Command](tx Tx) CommandInterceptor[C] {
	return func(ctx context.Context, cmd C, next CommandNext[C]) error {
		return tx.With(ctx, func(txCtx context.Context) error {
			return next(txCtx, cmd)
		})
	}
}
