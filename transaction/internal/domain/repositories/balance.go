package repositories

import (
	"context"
	"errors"
	"wallet/transaction/internal/domain/entities"

	"gorm.io/gorm"
)

// BalanceRepository Balance repository
type BalanceRepository struct {
	db *gorm.DB
}

// Save saves the balance entity to the database and returns any encountered error.
func (repo BalanceRepository) Save(ctx context.Context, balance *entities.Balance) error {
	return repo.db.WithContext(ctx).Save(balance).Error
}

// Get retrieves a balance entity from the database by its ID.
func (repo BalanceRepository) Get(ctx context.Context, userId uint64) (*entities.Balance, error) {
	var balance entities.Balance

	if err := repo.db.WithContext(ctx).Where("user_id = ?", userId).Limit(1).First(&balance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &balance, nil
}

// ensureBalanceRow inserts a zero-balance row for the user if one doesn't exist yet.
// Must be called as a separate statement before any UPDATE so the UPDATE can see the row
// (PostgreSQL CTE data-modifying statements execute with the same snapshot and cannot
// see rows inserted by their own CTE, so the ensure and update must be separate calls).
func (repo BalanceRepository) ensureBalanceRow(ctx context.Context, userID uint64) error {
	return repo.db.WithContext(ctx).Exec(`
		INSERT INTO balances (id, user_id, value)
		VALUES (gen_random_uuid(), ?, 0)
		ON CONFLICT (user_id) DO NOTHING
	`, userID).Error
}

// AtomicAdd atomically adds cents to the user's balance.
// Concurrent UPDATEs are serialised by PostgreSQL row-level locking:
// the second UPDATE blocks until the first commits, then sees the committed value.
func (repo BalanceRepository) AtomicAdd(ctx context.Context, userID uint64, cents int64) error {
	if err := repo.ensureBalanceRow(ctx, userID); err != nil {
		return err
	}
	return repo.db.WithContext(ctx).Exec(`
		UPDATE balances SET value = value + ? WHERE user_id = ?
	`, cents, userID).Error
}

// AtomicAddIfNonNegative atomically adds cents only when the result stays >= 0.
// Returns (false, nil) when the constraint is violated (lose would exceed balance).
// Concurrent callers are serialised by PostgreSQL row-level locking on the balances row.
func (repo BalanceRepository) AtomicAddIfNonNegative(ctx context.Context, userID uint64, cents int64) (bool, error) {
	if err := repo.ensureBalanceRow(ctx, userID); err != nil {
		return false, err
	}
	result := repo.db.WithContext(ctx).Exec(`
		UPDATE balances SET value = value + ?
		WHERE user_id = ? AND (value + ? >= 0 OR ? >= 0)
	`, cents, userID, cents, cents)

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// NewBalanceRepository returns BalanceRepository instance.
func NewBalanceRepository(db *gorm.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}
