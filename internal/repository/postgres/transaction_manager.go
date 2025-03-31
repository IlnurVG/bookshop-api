package postgres

import (
	"context"
	"fmt"

	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Context key for storing the transaction
type txKey struct{}

// TransactionManager implements repositories.TransactionManager
type TransactionManager struct {
	db *pgxpool.Pool
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *pgxpool.Pool) repositories.TransactionManager {
	return &TransactionManager{db: db}
}

// BeginTx starts a new transaction
func (m *TransactionManager) BeginTx(ctx context.Context) (context.Context, error) {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Store the transaction in the context
	return context.WithValue(ctx, txKey{}, tx), nil
}

// CommitTx commits the current transaction
func (m *TransactionManager) CommitTx(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction in context")
	}

	return tx.Commit(ctx)
}

// RollbackTx rolls back the current transaction
func (m *TransactionManager) RollbackTx(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction in context")
	}

	return tx.Rollback(ctx)
}

// WithTransaction executes a function within a transaction
func (m *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Begin transaction
	txCtx, err := m.BeginTx(ctx)
	if err != nil {
		return err
	}

	// Ensure rollback in case of panic
	defer func() {
		if p := recover(); p != nil {
			// Roll back the transaction and re-panic
			_ = m.RollbackTx(txCtx)
			panic(p)
		}
	}()

	// Execute the function
	err = fn(txCtx)
	if err != nil {
		// Roll back transaction in case of error
		rbErr := m.RollbackTx(txCtx)
		if rbErr != nil {
			return fmt.Errorf("error in transaction: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	// Commit the transaction
	if err := m.CommitTx(txCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTx returns the transaction from context or nil if no transaction exists
func GetTx(ctx context.Context) pgx.Tx {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return nil
	}
	return tx
}

// GetQueryer returns the transaction from context or connection pool if no transaction exists
func (m *TransactionManager) GetQueryer(ctx context.Context) pgxQuerier {
	tx := GetTx(ctx)
	if tx != nil {
		return tx
	}
	return m.db
}

// pgxQuerier defines an interface for executing queries
type pgxQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}
