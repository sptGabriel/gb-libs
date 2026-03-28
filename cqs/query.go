package cqs

import "context"

type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}

type Next[Q any, R any] func(ctx context.Context, q Q) (R, error)

type QueryInterceptor[Q any, R any] func(ctx context.Context, q Q, next Next[Q, R]) (R, error)

type decoratedQueryHandler[Q any, R any] struct {
	handler     QueryHandler[Q, R]
	interceptor QueryInterceptor[Q, R]
}

func (d decoratedQueryHandler[Q, R]) Handle(ctx context.Context, q Q) (R, error) {
	return d.interceptor(ctx, q, d.handler.Handle)
}

func NewQueryHandler[Q any, R any](
	h QueryHandler[Q, R],
	interceptors ...QueryInterceptor[Q, R],
) QueryHandler[Q, R] {
	if len(interceptors) == 0 {
		return h
	}

	var decorated QueryHandler[Q, R] = h
	for i := len(interceptors) - 1; i >= 0; i-- {
		decorated = decoratedQueryHandler[Q, R]{
			handler:     decorated,
			interceptor: interceptors[i],
		}
	}

	return decorated
}
