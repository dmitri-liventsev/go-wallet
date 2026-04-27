package repositories

import (
	"errors"
	"wallet/transaction/internal/domain/entities"

	"gorm.io/gorm"
)

// BalanceRepository Balance repository
type BalanceRepository struct {
	db *gorm.DB
}

// Save saves the balance entity to the database and returns any encountered error.
func (repo BalanceRepository) Save(balance *entities.Balance) error {
	return repo.db.Save(balance).Error
}

// Get retrieves a balance entity from the database by its ID.
func (repo BalanceRepository) Get(userId uint64) (*entities.Balance, error) {
	var balance entities.Balance

	if err := repo.db.Where("user_id = ?", userId).Limit(1).First(&balance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &balance, nil
}

// NewBalanceRepository returns BalanceRepository instance.
func NewBalanceRepository(db *gorm.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}
