package services

import (
	"errors"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
)

// TransactionProcessor handles the processing of transactions,
// including repository operations and balance aggregation.
type TransactionProcessor struct {
	TxRepo         *repositories.TransactionRepository
	BalanceService *Balance
}

// Execute processes the given transaction by updating the balance and marking the transaction as done
// or cancelled based on the outcome. Internal transactions ignoring negative balance validation
func (t TransactionProcessor) Execute(transaction *entities.Transaction) error {
	var err error

	if transaction.IsInternal() {
		err = t.BalanceService.ForceUpdateBalance(transaction.Amount)
	} else {
		err = t.BalanceService.UpdateBalance(transaction.Amount)
	}

	if err != nil && errors.Is(err, ErrNegativeBalance) {
		transaction.MarkAsCancelled()
	} else if err != nil {
		return err
	} else {
		transaction.MarkAsDone()
	}

	err = t.TxRepo.Save(transaction)
	if err != nil {
		return err
	}

	return nil
}

// NewTransactionProcessor returns TransactionProcessor instance.
func NewTransactionProcessor(db *gorm.DB) TransactionProcessor {
	return TransactionProcessor{
		TxRepo:         repositories.NewTransactionRepository(db),
		BalanceService: NewBalanceService(db),
	}
}
