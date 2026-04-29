package services

import (
	"context"
	"errors"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TransactionProcessor handles the processing of transactions,
// including repository operations and balance aggregation.
type TransactionProcessor struct {
	db *gorm.DB
}

// Execute processes the given transaction atomically.
//
// At the start of the DB transaction it acquires a row-level lock on the
// transactions row (SELECT FOR UPDATE WHERE status = 'new'). This prevents
// two workers from double-processing the same transaction when a user lock
// expires mid-cycle: the second worker blocks on the lock, then sees
// status != 'new' after the first worker commits and skips the row.
func (t TransactionProcessor) Execute(ctx context.Context, transaction *entities.Transaction) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the row for the duration of this DB transaction.
		// If another worker already holds the lock, we block here until it
		// commits, after which status will no longer be 'new' and we skip.
		var current entities.Transaction
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", transaction.ID, entities.New).
			First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // already processed by another worker
		}
		if err != nil {
			return err
		}

		balanceService := &Balance{repo: repositories.NewBalanceRepository(tx)}
		txRepo := repositories.NewTransactionRepository(tx)

		var balanceErr error
		if transaction.IsInternal() {
			balanceErr = balanceService.ForceUpdateBalance(ctx, transaction.Amount, transaction.UserID)
		} else {
			balanceErr = balanceService.UpdateBalance(ctx, transaction.Amount, transaction.UserID)
		}

		if balanceErr != nil && errors.Is(balanceErr, ErrNegativeBalance) {
			transaction.MarkAsCancelled()
		} else if balanceErr != nil {
			return balanceErr
		} else {
			transaction.MarkAsDone()
		}

		return txRepo.Save(ctx, transaction)
	})
}

// NewTransactionProcessor returns TransactionProcessor instance.
func NewTransactionProcessor(db *gorm.DB) TransactionProcessor {
	return TransactionProcessor{db: db}
}
