package service

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"
)

// BalanceRepository balance storage.
type BalanceRepository interface {
	Save(balance *entities.Balance) error
	Get() (*entities.Balance, error)
}

// BalanceCalculator actual balance calculator
type BalanceCalculator interface {
	CalculateBalance() (int64, error)
}

// Balance service
type Balance struct {
	repo       BalanceRepository
	calculator BalanceCalculator
}

// ProvideBalance returns the current balance if it exists, otherwise calculates a new balance,
// creates a new balance object, saves it, and returns the newly created balance.
func (b *Balance) ProvideBalance() (*entities.Balance, error) {
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

// ErrNegativeBalance error.
var ErrNegativeBalance = errors.New("TotalAmount cannot be negative")

// UpdateBalance updates the current balance by adding the specified amount,
// returns an error if the balance becomes negative or if any operation fails.
func (b *Balance) UpdateBalance(amount vo.Amount) error {
	balance, err := b.ProvideBalance()
	if err != nil {
		return errors.Wrap(err, "cannot update balance")
	}

	balance.Value = balance.Value.AddAmount(amount)
	if balance.Value.LessThanZero() {
		return ErrNegativeBalance
	}

	return b.repo.Save(balance)
}

// ForceUpdateBalance updates the current balance by adding the specified amount
// and saves the updated balance without checking for negative values.
func (b *Balance) ForceUpdateBalance(amount vo.Amount) error {
	balance, err := b.ProvideBalance()
	if err != nil {
		return errors.Wrap(err, "cannot update balance")
	}
	balance.Value = balance.Value.AddAmount(amount)

	return b.repo.Save(balance)
}

// NewBalanceService returns an instance of Balance service.
func NewBalanceService(db *gorm.DB) *Balance {
	return &Balance{
		repo:       repositories.NewBalanceRepository(db),
		calculator: repositories.NewTransactionRepository(db),
	}
}
