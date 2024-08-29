package workers

import (
	"context"
	"github.com/google/uuid"
	"goa.design/clue/log"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
)

// RunBalanceWorker starts a background goroutine that continuously executes the balance worker,
// handling transactions and rolling back on errors until the context is done.
func RunBalanceWorker(ctx context.Context, db *gorm.DB) {
	go func(ctx context.Context) {
		lockUuid := uuid.New()
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf(ctx, "Recovered from panic")
					}
				}()
			}()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					tx := db.Begin()
					worker := NewBalanceWorker(tx, lockUuid)
					err := worker.Execute()
					if err != nil {
						tx.Rollback()
						log.Fatalf(ctx, err, "cannot Execute balance worker")
						continue
					}
					_ = tx.Commit().Error
				}
			}
		}
	}(ctx)
}

type Locker interface {
	LockNewTransactions(lockUuid uuid.UUID) error
}

type Provider interface {
	GetLockedTransactions() ([]entities.Transaction, error)
}

type Processor interface {
	Execute(*entities.Transaction) error
}

// BalanceWorker is responsible for monitoring new correction requests, initiating correction processing,
// tracking new transactions, and initiating their processing.
type BalanceWorker struct {
	LockUuid  uuid.UUID
	Locker    Locker
	Provider  Provider
	Processor Processor
}

// Execute retrieves and processes the current correction if available,
// then retrieves and processes the next transaction.
func (b BalanceWorker) Execute() error {
	// Lock transactions
	err := b.Locker.LockNewTransactions(b.LockUuid)
	if err != nil {
		return err
	}

	//get  all locked transactions
	transactions, err := b.Provider.GetLockedTransactions()
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		if *transaction.LockUuid != b.LockUuid {
			return nil // next transactions to handle was booked by another process
		}

		err := b.Processor.Execute(&transaction)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewBalanceWorker returns BalanceWorker instance.
func NewBalanceWorker(db *gorm.DB, lockUuid uuid.UUID) BalanceWorker {
	transactionRepository := repositories.NewTransactionRepository(db)

	return BalanceWorker{
		LockUuid:  lockUuid,
		Locker:    transactionRepository,
		Provider:  transactionRepository,
		Processor: services.NewTransactionProcessor(db),
	}
}
