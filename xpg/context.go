package xpg

import (
	"context"
	"errors"
)

type ctxConnKey struct{}

var ctxConn = ctxConnKey{}

var ErrConnNotInContext = errors.New("no pool on context")

func withQuerier(ctx context.Context, conn Querier) context.Context {
	return context.WithValue(ctx, ctxConn, conn)
}

func querierFromContext(ctx context.Context) Querier {
	return ctx.Value(ctxConn).(Querier)
}
