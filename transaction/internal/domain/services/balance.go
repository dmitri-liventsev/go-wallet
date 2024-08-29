package services

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

// Balance service
type Balance struct {
	repo            BalanceRepository
	balanceProvider BalanceProvider
}

// ErrNegativeBalance error.
var ErrNegativeBalance = errors.New("TotalAmount cannot be negative")

// UpdateBalance updates the current balance by adding the specified amount,
// returns an error if the balance becomes negative or if any operation fails.
func (b *Balance) UpdateBalance(amount vo.Amount) error {
	balance, err := b.balanceProvider.Provide()
	if err != nil {
		return errors.Wrap(err, "cannot update balance")
	}

	balance.Value = balance.Value.AddAmount(amount)
	if balance.Value.LessThanZero() && amount.LessThenZero() {
		return ErrNegativeBalance
	}

	return b.repo.Save(balance)
}

// ForceUpdateBalance updates the current balance by adding the specified amount
// and saves the updated balance without checking for negative values.
func (b *Balance) ForceUpdateBalance(amount vo.Amount) error {
	balance, err := b.balanceProvider.Provide()
	if err != nil {
		return errors.Wrap(err, "cannot update balance")
	}
	balance.Value = balance.Value.AddAmount(amount)

	return b.repo.Save(balance)
}

// NewBalanceService returns an instance of Balance service.
func NewBalanceService(db *gorm.DB) *Balance {
	return &Balance{
		repo:            repositories.NewBalanceRepository(db),
		balanceProvider: NewBalanceProvider(db),
	}
}
