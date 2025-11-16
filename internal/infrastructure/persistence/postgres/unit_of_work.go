package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UnitOfWork manages database transactions.
type UnitOfWork struct {
	pool *pgxpool.Pool
}

// WithinTransaction executes a function within a database transaction with Repeatable Read isolation level.
func (uow *UnitOfWork) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := uow.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)

	err = fn(txCtx)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// txOrPool is an interface pgx.Tx and Connection.
type txOrPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// txKey is a key for storing transaction in context.
type txKey struct{}

// getTx extracts transaction from context or returns pool.
func getTx(ctx context.Context, pool *pgxpool.Pool) txOrPool {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}
