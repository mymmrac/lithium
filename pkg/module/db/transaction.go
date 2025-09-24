package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/bun"
)

// Transaction replaces regular Bun DB with transactional one
/*
Example:
	// Logic level
	ctx, err := tx.Begin(ctx)
	if err != nil { ... }
	defer func() { _ = tx.Rollback(ctx) }()

	// Repository level
	tx.Extract(ctx).ExecContext(ctx, ...)

	// Logic level
	if err = tx.Commit(ctx); err != nil { ... }
*/
type Transaction interface {
	DB() *bun.DB
	Extract(ctx context.Context) bun.IDB
	Begin(ctx context.Context, opts ...*sql.TxOptions) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Detach(ctx context.Context) context.Context
}

type transaction struct {
	db *bun.DB
}

// transactionKey key for transaction in context.
//
//nolint:gochecknoglobals
var transactionKey = transactionKeyType{}

type transactionKeyType struct{}

func (a *transaction) DB() *bun.DB {
	return a.db
}

func (a *transaction) Extract(ctx context.Context) bun.IDB {
	if tx, ok := ctx.Value(transactionKey).(bun.Tx); ok {
		return tx
	}
	return a.db
}

func (a *transaction) Begin(ctx context.Context, opts ...*sql.TxOptions) (context.Context, error) {
	var opt *sql.TxOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	tx, err := a.db.BeginTx(ctx, opt)
	if err != nil {
		return ctx, fmt.Errorf("begin tx: %w", err)
	}

	return context.WithValue(ctx, transactionKey, tx), nil
}

func (a *transaction) Commit(ctx context.Context) error {
	if tx, ok := ctx.Value(transactionKey).(bun.Tx); ok {
		return tx.Commit()
	}
	return errors.New("no transaction to commit")
}

func (a *transaction) Rollback(ctx context.Context) error {
	if tx, ok := ctx.Value(transactionKey).(bun.Tx); ok {
		return tx.Rollback()
	}
	return errors.New("no transaction to rollback")
}

func (a *transaction) Detach(ctx context.Context) context.Context {
	return context.WithValue(ctx, transactionKey, nil)
}
