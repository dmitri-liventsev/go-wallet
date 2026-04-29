package services

import (
	"context"
	"errors"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"

	"gorm.io/gorm"
)

// TransactionProcessor handles the processing of transactions,
// including repository operations and balance aggregation.
type TransactionProcessor struct {
	db *gorm.DB
}

// Execute processes the given transaction atomically:
// the balance update and the transaction status save happen in a single DB transaction,
// so a crash between the two steps cannot produce an inconsistent state.
func (t TransactionProcessor) Execute(ctx context.Context, transaction *entities.Transaction) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
