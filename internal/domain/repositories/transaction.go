package repositories

import "context"

// TransactionManager defines methods for managing database transactions
type TransactionManager interface {
	// BeginTx starts a new transaction
	BeginTx(ctx context.Context) (context.Context, error)

	// CommitTx commits the current transaction
	CommitTx(ctx context.Context) error

	// RollbackTx rolls back the current transaction
	RollbackTx(ctx context.Context) error

	// WithTransaction executes a function within a transaction
	// If the function returns an error, the transaction is rolled back
	// Otherwise, the transaction is committed
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
