package xpg

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pgx "github.com/jackc/pgx/v5"
	"github.com/sptGabriel/gb-libs/xlog"
	"go.opentelemetry.io/otel/trace"
)

// Tx runs fn inside a transaction, committing when fn returns nil and
// rolling back when it returns an error.
type Tx interface {
	With(ctx context.Context, fn func(context.Context) error) error
}

// TxHandle controls the lifecycle of a transaction opened with Begin.
type TxHandle interface {
	Commit(ctx context.Context) error
	// Rollback is a no-op after a successful Commit, so callers should
	// defer it right after Begin to cover error and panic paths.
	Rollback(ctx context.Context) error
}

// TxManager is the full transaction contract: the callback API plus the
// manual Begin primitive it is built on.
type TxManager interface {
	Tx
	Begin(ctx context.Context) (context.Context, TxHandle, error)
}

type transactioner struct {
	db *Database
}

var _ TxManager = (*transactioner)(nil)

// Begin opens a transaction and hands lifecycle control to the caller.
// The returned context carries the transaction querier and must be used
// on every statement meant to run inside it. When ctx already carries a
// transaction, a nested one backed by a savepoint is created.
func (t *transactioner) Begin(ctx context.Context) (context.Context, TxHandle, error) {
	var span trace.Span
	if t.db.tracer != nil {
		ctx, span = t.db.tracer.Start(ctx, transactionTraceSpan)
	}

	nTx, err := t.begin(ctx)
	if err != nil {
		if span != nil {
			span.End()
		}

		return ctx, nil, fmt.Errorf("on begin transaction: %w", err)
	}

	return withQuerier(ctx, nTx), &txHandle{tx: nTx, span: span}, nil
}

func (t *transactioner) begin(ctx context.Context) (pgx.Tx, error) {
	if outer, ok := querierFromContext(ctx).(pgx.Tx); ok {
		return outer.Begin(ctx) // nested: pgx uses a savepoint
	}

	return t.db.rawConnection().Begin(ctx)
}

// With joins the transaction already present in ctx, if any; otherwise it
// opens one via Begin and commits or rolls back around fn.
func (t *transactioner) With(ctx context.Context, fn func(context.Context) error) error {
	if _, hasTx := querierFromContext(ctx).(pgx.Tx); hasTx {
		return fn(ctx)
	}

	txCtx, nTx, err := t.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if pErr := recover(); pErr != nil {
			_ = nTx.Rollback(ctx)

			panic(pErr)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := nTx.Rollback(ctx); rbErr != nil {
			xlog.ErrorContext(ctx, "tx rolling back error", xlog.String("error", rbErr.Error()))
		}

		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nTx.Commit(ctx)
}

// WithResult is With for callers that produce a value inside the
// transaction boundary.
func WithResult[T any](ctx context.Context, tx Tx, fn func(context.Context) (T, error)) (T, error) {
	var out T

	err := tx.With(ctx, func(txCtx context.Context) error {
		var err error
		out, err = fn(txCtx)

		return err
	})

	return out, err
}

type txHandle struct {
	tx   pgx.Tx
	span trace.Span
	once sync.Once
}

func (h *txHandle) Commit(ctx context.Context) error {
	defer h.end()

	if err := h.tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (h *txHandle) Rollback(ctx context.Context) error {
	defer h.end()

	if err := h.tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return fmt.Errorf("transaction rollback failed: %w", err)
	}

	return nil
}

func (h *txHandle) end() {
	h.once.Do(func() {
		if h.span != nil {
			h.span.End()
		}
	})
}

func NewTransactioner(db *Database) *transactioner {
	return &transactioner{db: db}
}
