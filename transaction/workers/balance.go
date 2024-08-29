package workers

import (
	"context"
	"github.com/google/uuid"
	"goa.design/clue/log"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	txdb "wallet/transaction/internal/infrastructure/db"
)

// RunBalanceWorker starts a background goroutine that continuously executes the balance worker,
// handling transactions and rolling back on errors until the context is done.
func RunBalanceWorker(ctx context.Context, db *gorm.DB) {
	go func(ctx context.Context) {
		worker := NewBalanceWorker(db)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := worker.Execute()
				if err != nil {
					log.Fatalf(ctx, err, "cannot Execute balance worker")
				}
			}
		}
	}(ctx)
}

// BalanceWorker is responsible for monitoring new correction requests, initiating correction processing,
// tracking new transactions, and initiating their processing.
type BalanceWorker struct {
	LockUuid uuid.UUID
	db       *gorm.DB
}

// Execute retrieves and processes the current correction if available,
// then retrieves and processes the next transaction.
func (b BalanceWorker) Execute() error {
	tx := b.db.Begin()
	// Lock transactions
	txProvider := repositories.NewTransactionRepository(tx)
	err := txProvider.LockNewTransactions(b.LockUuid)
	if err != nil {

		tx.Rollback()
		return err
	}

	//get  all locked transactions
	transactions, err := txProvider.GetLockedTransactions()
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, transaction := range transactions {
		if *transaction.LockUuid != b.LockUuid {
			_ = txdb.TryTxCommit(tx)
			return nil // next transactions to handle was booked by another process
		}

		err := services.NewTransactionProcessor(tx).Execute(&transaction)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	_ = txdb.TryTxCommit(tx)

	return nil
}

// NewBalanceWorker returns BalanceWorker instance.
func NewBalanceWorker(db *gorm.DB) BalanceWorker {
	return BalanceWorker{
		db:       db,
		LockUuid: uuid.New(),
	}
}
