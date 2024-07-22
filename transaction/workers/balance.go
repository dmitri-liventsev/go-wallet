package workers

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
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
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Fatalf(ctx, fmt.Errorf("recovered in goroutine: %v", r), "recovering from panic")
			}
		}()

		runBalanceWorker(ctx, db)
	}()
}

func runBalanceWorker(ctx context.Context, db *gorm.DB) {
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

// CorrectionProcessor defines an interface for processing corrections.
type CorrectionProcessor interface {
	Execute(correction *entities.Correction) error
}

// TransactionProcessor defines an interface for processing transactions.
type TransactionProcessor interface {
	Execute(transaction *entities.Transaction) error
}

// CorrectionProvider defines an interface for retrieving corrections.
type CorrectionProvider interface {
	GetActualCorrection() (*entities.Correction, error)
	GetNewestCorrection() (*entities.Correction, error)
}

// TransactionProvider defines an interface for retrieving the next transaction to be processed.
type TransactionProvider interface {
	GetNextTransaction() (*entities.Transaction, error)
}

// BalanceWorker is responsible for monitoring new correction requests, initiating correction processing,
// tracking new transactions, and initiating their processing.
type BalanceWorker struct {
	repo                 CorrectionProvider
	txRepo               TransactionProvider
	CorrectionProcessor  CorrectionProcessor
	TransactionProcessor TransactionProcessor
}

// Execute retrieves and processes the current correction if available,
// then retrieves and processes the next transaction.
func (b BalanceWorker) Execute() error {
	{
		correction, err := b.repo.GetActualCorrection()
		if err != nil {
			return errors.Wrap(err, "cannot get last correction")
		}

		if correction != nil {
			err := b.CorrectionProcessor.Execute(correction)
			if err != nil {
				return errors.Wrap(err, "cannot Execute correction processor")
			}
		}
	}

	{
		nextTx, err := b.txRepo.GetNextTransaction()
		if err != nil {
			return errors.Wrap(err, "can not get next transaction")
		}
		if nextTx == nil {
			return nil
		}

		err = b.TransactionProcessor.Execute(nextTx)
		if err != nil {
			return errors.Wrap(err, "can not Execute transaction processor")
		}
	}

	return nil
}

// NewBalanceWorker returns BalanceWorker instance.
func NewBalanceWorker(db *gorm.DB) BalanceWorker {
	return BalanceWorker{
		repo:                 repositories.NewCorrectionRepository(db),
		txRepo:               repositories.NewTransactionRepository(db),
		CorrectionProcessor:  services.NewCorrectionProcessor(db),
		TransactionProcessor: services.NewTransactionProcessor(db),
	}
}
