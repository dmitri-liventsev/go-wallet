package repositories

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
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
func (repo BalanceRepository) Get() (*entities.Balance, error) {
	var balance entities.Balance
	balanceID := uuid.MustParse(entities.BalanceID)

	if err := repo.db.Where("id = ?", balanceID).Limit(1).First(&balance).Error; err != nil {
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
