package cqs

import (
	"context"
	"errors"
	"sync"
)

var ErrHandlerNotFound = errors.New("no handler registered for this type")

// CommandBus dispatches commands to their registered handlers.
// Register and Dispatch are generic methods (Go 1.27).
type CommandBus struct {
	mu       sync.RWMutex
	handlers map[string]any
}

func NewCommandBus() *CommandBus {
	return &CommandBus{handlers: make(map[string]any)}
}

func (b *CommandBus) Register[C Command](h CommandHandler[C]) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[typeName[C]()] = h
}

func (b *CommandBus) Dispatch[C Command](ctx context.Context, cmd C) error {
	b.mu.RLock()
	h, ok := b.handlers[typeName[C]()]
	b.mu.RUnlock()
	if !ok {
		return ErrHandlerNotFound
	}
	return h.(CommandHandler[C]).Handle(ctx, cmd)
}

// QueryBus routes queries to registered handlers.
// Register and Query are generic methods (Go 1.27).
type QueryBus struct {
	mu       sync.RWMutex
	handlers map[string]any
}

func NewQueryBus() *QueryBus {
	return &QueryBus{handlers: make(map[string]any)}
}

func (b *QueryBus) Register[Q any, R any](h QueryHandler[Q, R]) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[typeName[Q]()] = h
}

func (b *QueryBus) Query[Q any, R any](ctx context.Context, q Q) (R, error) {
	b.mu.RLock()
	h, ok := b.handlers[typeName[Q]()]
	b.mu.RUnlock()
	if !ok {
		var zero R
		return zero, ErrHandlerNotFound
	}
	return h.(QueryHandler[Q, R]).Handle(ctx, q)
}
