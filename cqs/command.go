package cqs

import "context"

type Command interface {
	Validate() error
}

type CommandHandler[C Command] interface {
	Handle(ctx context.Context, cmd C) error
}

type CommandNext[C Command] func(ctx context.Context, cmd C) error

type CommandInterceptor[C Command] func(ctx context.Context, cmd C, next CommandNext[C]) error

type decoratedCommandHandler[C Command] struct {
	handler     CommandHandler[C]
	interceptor CommandInterceptor[C]
}

func (d decoratedCommandHandler[C]) Handle(ctx context.Context, cmd C) error {
	return d.interceptor(ctx, cmd, d.handler.Handle)
}

func NewCommandHandler[C Command](
	h CommandHandler[C],
	interceptors ...CommandInterceptor[C],
) CommandHandler[C] {
	if len(interceptors) == 0 {
		return h
	}

	var decorated CommandHandler[C] = h
	for i := len(interceptors) - 1; i >= 0; i-- {
		decorated = decoratedCommandHandler[C]{
			handler:     decorated,
			interceptor: interceptors[i],
		}
	}

	return decorated
}
