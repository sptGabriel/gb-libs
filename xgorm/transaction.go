package xgorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// context key privada para evitar colisões
type txCtxKey struct{}

var (
	ErrGetTxController         = errors.New("could not retrieve transaction controller from context")
	ErrTransactionNotInitiated = errors.New("transaction was not initiated")
	ErrTransactionAlreadyInit  = errors.New("transaction already initiated")
)

type DBController any

type TransactionController[C DBController] interface {
	QueryDB(context.Context) C
	DB(context.Context) (C, error)
	Begin(context.Context) (C, error)
	Commit(context.Context) error
	Rollback(context.Context) error
}

type GORMTransactionController struct {
	db *gorm.DB
	tx *gorm.DB
}

func NewGORMTransactionController(db *gorm.DB) *GORMTransactionController {
	return &GORMTransactionController{db: db}
}

// Retorna o DB sem transação
func (c *GORMTransactionController) QueryDB(ctx context.Context) *gorm.DB {
	return c.db.WithContext(ctx)
}

// Retorna o DB dentro da transação
func (c *GORMTransactionController) DB(ctx context.Context) (*gorm.DB, error) {
	if c.tx == nil {
		return nil, ErrTransactionNotInitiated
	}
	return c.tx.WithContext(ctx), nil
}

// Inicia uma transação
func (c *GORMTransactionController) Begin(ctx context.Context) (*gorm.DB, error) {
	if c.tx != nil {
		return nil, ErrTransactionAlreadyInit
	}
	c.tx = c.db.WithContext(ctx).Begin()
	if c.tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", c.tx.Error)
	}
	return c.tx, nil
}

// Commit da transação
func (c *GORMTransactionController) Commit(ctx context.Context) error {
	if c.tx == nil {
		return ErrTransactionNotInitiated
	}
	if err := c.tx.WithContext(ctx).Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	c.tx = nil
	return nil
}

// Rollback da transação
func (c *GORMTransactionController) Rollback(ctx context.Context) error {
	if c.tx == nil {
		return ErrTransactionNotInitiated
	}

	if err := c.tx.WithContext(ctx).Rollback().Error; err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	c.tx = nil
	return nil
}

func NewTransactioner(db *gorm.DB) TransactionController[*gorm.DB] {
	return NewGORMTransactionController(db)
}

func TransactionerWithContext(ctx context.Context) (TransactionController[*gorm.DB], error) {
	txAny := ctx.Value(txCtxKey{})
	if txAny == nil {
		return nil, ErrGetTxController
	}

	txCtrl, ok := txAny.(TransactionController[*gorm.DB])
	if !ok {
		return nil, ErrGetTxController
	}

	return txCtrl, nil
}

func TransactionerFromContext[T DBController](ctx context.Context) (TransactionController[T], error) {
	txAny := ctx.Value(txCtxKey{})
	if txAny == nil {
		return nil, ErrGetTxController
	}

	txCtrl, ok := txAny.(TransactionController[T])
	if !ok {
		return nil, ErrGetTxController
	}

	return txCtrl, nil
}

func ContextWithTransactioner(ctx context.Context, txCtrl TransactionController[*gorm.DB]) context.Context {
	return context.WithValue(ctx, txCtxKey{}, txCtrl)
}
