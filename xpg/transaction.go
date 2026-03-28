package xpg

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/sptGabriel/defaults/xlog"
)

type Tx interface {
	With(ctx context.Context, fn func(context.Context) error) error
}

type transactioner struct {
	db *Database
}

func (t *transactioner) With(ctx context.Context, fn func(context.Context) error) error {
	q := querierFromContext(ctx)

	_, hasTx := q.(pgx.Tx)
	if hasTx {
		return fn(ctx)
	}

	if t.db.tracer != nil {
		newctx, span := t.db.tracer.Start(ctx, transactionTraceSpan)
		defer span.End()

		ctx = newctx
	}

	nTx, err := t.db.rawConnection().Begin(ctx)
	if err != nil {
		return fmt.Errorf("on begin transaction: %w", err)
	}

	defer func() {
		if pErr := recover(); pErr != nil {
			_ = nTx.Rollback(ctx)

			panic(pErr)
		}
	}()

	if err = fn(withQuerier(ctx, nTx)); err != nil {
		if rbErr := nTx.Rollback(ctx); rbErr != nil {
			xlog.ErrorContext(ctx, "tx rolling back error", xlog.String("error", rbErr.Error()))
		}

		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	if err := nTx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func NewTransactioner(db *Database) *transactioner {
	return &transactioner{db: db}
}
