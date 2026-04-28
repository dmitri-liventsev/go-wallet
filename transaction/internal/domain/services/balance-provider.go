package services

import (
	"context"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// BalanceCalculator actual balance calculator
type BalanceCalculator interface {
	CalculateBalance(ctx context.Context) (int64, error)
}

type BalanceProvider struct {
	repo       *repositories.BalanceRepository
	calculator BalanceCalculator
}

func (b BalanceProvider) Provide(ctx context.Context, userId uint64) (*entities.Balance, error) {
	balance, err := b.repo.Get(ctx, userId)
	if err != nil {
		return nil, errors.Wrap(err, "cannot provide balance")
	}

	if balance != nil {
		return balance, nil
	}

	calculatedValue, err := b.calculator.CalculateBalance(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot provide balance")
	}

	balance = entities.NewBalance(vo.NewTotalAmount(calculatedValue), userId)
	err = b.repo.Save(ctx, balance)
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
