package workers

import (
	"context"
	"github.com/google/uuid"
	"goa.design/clue/log"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	txdb "wallet/transaction/internal/infrastructure/db"
)

// RunBalanceWorker starts a background goroutine that continuously executes the balance worker,
// handling transactions and rolling back on errors until the context is done.
func RunBalanceWorker(ctx context.Context, db *gorm.DB) {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				tx := db.Begin()
				err := NewBalanceWorker(tx).Execute()
				if err != nil {
					log.Fatalf(ctx, err, "cannot Execute balance worker")
					tx.Rollback()
				}

				_ = txdb.TryTxCommit(tx)
			}
		}
	}(ctx)
}

// TransactionProcessor defines an interface for processing transactions.
type TransactionProcessor interface {
	Execute(transaction *entities.Transaction) error
}

// TransactionProvider defines an interface for retrieving the next transaction to be processed.
type TransactionProvider interface {
	LockNewTransactions(lockUuid uuid.UUID) error
	GetLockedTransactions() ([]entities.Transaction, error)
}

// BalanceWorker is responsible for monitoring new correction requests, initiating correction processing,
// tracking new transactions, and initiating their processing.
type BalanceWorker struct {
	txProvider           TransactionProvider
	TransactionProcessor TransactionProcessor
	LockUuid             uuid.UUID
}

// Execute retrieves and processes the current correction if available,
// then retrieves and processes the next transaction.
func (b BalanceWorker) Execute() error {
	// Lock transactions
	err := b.txProvider.LockNewTransactions(b.LockUuid)
	if err != nil {
		return err
	}

	//get  all locked transactions
	transactions, err := b.txProvider.GetLockedTransactions()
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		if *transaction.LockUuid != b.LockUuid {
			return nil // next transactions to handle was booked by another process
		}

		err := b.TransactionProcessor.Execute(&transaction)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewBalanceWorker returns BalanceWorker instance.
func NewBalanceWorker(db *gorm.DB) BalanceWorker {
	return BalanceWorker{
		txProvider:           repositories.NewTransactionRepository(db),
		TransactionProcessor: services.NewTransactionProcessor(db),
		LockUuid:             uuid.New(),
	}
}
