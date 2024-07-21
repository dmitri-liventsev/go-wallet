package transaction

import (
	"errors"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"
)

// AddTransaction represents a transaction to be added, including source type, action, amount, and an identifier.
type AddTransaction struct {
	SourceType string
	Action     string
	Amount     vo.Amount
	ID         string
}

// TransactionStorage defines an interface for storing transactions with a method to create a new transaction.
type TransactionStorage interface {
	Create(transaction *entities.Transaction) error
}

// Execute creates and stores a new transaction using the provided TransactionStorage repository,
// skipping if the amount is zero.
func (a *AddTransaction) Execute(repo TransactionStorage) error {
	if a.Amount.Equal(vo.NewAmount(0)) {
		return nil
	}
	transaction := entities.NewTransaction(a.ID, a.Amount, a.Action, a.SourceType)

	err := repo.Create(transaction)
	if err != nil && !errors.Is(err, repositories.ErrDuplicateKey) {
		return err
	}

	return nil
}
