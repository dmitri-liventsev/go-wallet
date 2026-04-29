package services

import (
	"context"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// BalanceRepository balance storage.
type BalanceRepository interface {
	Save(ctx context.Context, balance *entities.Balance) error
	Get(ctx context.Context, userId uint64) (*entities.Balance, error)
	AtomicAdd(ctx context.Context, userID uint64, cents int64) error
	AtomicAddIfNonNegative(ctx context.Context, userID uint64, cents int64) (bool, error)
}

// Balance service
type Balance struct {
	repo BalanceRepository
}

// ErrNegativeBalance error.
var ErrNegativeBalance = errors.New("TotalAmount cannot be negative")

// UpdateBalance atomically adds amount to the user's balance.
// Returns ErrNegativeBalance if the result would be negative (lose exceeds balance).
func (b *Balance) UpdateBalance(ctx context.Context, amount vo.Amount, userId uint64) error {
	ok, err := b.repo.AtomicAddIfNonNegative(ctx, userId, int64(amount.Cents))
	if err != nil {
		return errors.Wrap(err, "cannot update balance")
	}
	if !ok {
		return ErrNegativeBalance
	}
	return nil
}

// ForceUpdateBalance atomically adds amount without checking for negative balance.
func (b *Balance) ForceUpdateBalance(ctx context.Context, amount vo.Amount, userId uint64) error {
	if err := b.repo.AtomicAdd(ctx, userId, int64(amount.Cents)); err != nil {
		return errors.Wrap(err, "cannot update balance")
	}
	return nil
}

// NewBalanceService returns an instance of Balance service.
func NewBalanceService(db *gorm.DB) *Balance {
	return &Balance{
		repo: repositories.NewBalanceRepository(db),
	}
}
