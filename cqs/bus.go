package cqs

import (
	"context"
	"errors"
	"sync"
)

var ErrHandlerNotFound = errors.New("no handler registered for this type")

type CommandBus struct {
	mu       sync.RWMutex
	handlers map[string]any
}

func NewCommandBus() *CommandBus {
	return &CommandBus{handlers: make(map[string]any)}
}

func Register[C Command](b *CommandBus, h CommandHandler[C]) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[typeName[C]()] = h
}

func Dispatch[C Command](b *CommandBus, ctx context.Context, cmd C) error {
	b.mu.RLock()
	h, ok := b.handlers[typeName[C]()]
	b.mu.RUnlock()
	if !ok {
		return ErrHandlerNotFound
	}
	return h.(CommandHandler[C]).Handle(ctx, cmd)
}

type QueryBus struct {
	mu       sync.RWMutex
	handlers map[string]any
}

func NewQueryBus() *QueryBus {
	return &QueryBus{handlers: make(map[string]any)}
}

func RegisterQuery[Q any, R any](b *QueryBus, h QueryHandler[Q, R]) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[typeName[Q]()] = h
}

func Query[Q any, R any](b *QueryBus, ctx context.Context, q Q) (R, error) {
	b.mu.RLock()
	h, ok := b.handlers[typeName[Q]()]
	b.mu.RUnlock()
	if !ok {
		var zero R
		return zero, ErrHandlerNotFound
	}
	return h.(QueryHandler[Q, R]).Handle(ctx, q)
}
