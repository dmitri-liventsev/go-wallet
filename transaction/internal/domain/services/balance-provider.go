package services

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"
)

// BalanceCalculator actual balance calculator
type BalanceCalculator interface {
	CalculateBalance() (int64, error)
}

type BalanceProvider struct {
	repo       *repositories.BalanceRepository
	calculator BalanceCalculator
}

func (b BalanceProvider) Provide() (*entities.Balance, error) {
	balance, err := b.repo.Get()
	if err != nil {
		return nil, errors.Wrap(err, "cannot provide balance")
	}

	if balance != nil {
		return balance, nil
	}

	calculatedValue, err := b.calculator.CalculateBalance()
	if err != nil {
		return nil, errors.Wrap(err, "cannot provide balance")
	}

	balance = entities.NewBalance(vo.NewTotalAmount(calculatedValue))
	err = b.repo.Save(balance)
	if err != nil {
		return nil, errors.Wrap(err, "cannot provide balance")
	}

	return balance, err
}

func NewBalanceProvider(db *gorm.DB) BalanceProvider {
	return BalanceProvider{
		repo:       repositories.NewBalanceRepository(db),
		calculator: repositories.NewTransactionRepository(db),
	}
}
