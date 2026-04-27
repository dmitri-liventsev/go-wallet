package entities

import (
	"wallet/transaction/internal/domain/vo"

	"github.com/google/uuid"
)

// Balance represents the Balance entity, responsible for storing the current state of the balance.
type Balance struct {
	ID     uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID uint64         `gorm:"type:integer"`
	Value  vo.TotalAmount `gorm:"type:bigint;not null"`
}

// NewBalance returns new Balance entity instance.
func NewBalance(value vo.TotalAmount, userId uint64) *Balance {
	return &Balance{
		ID:     uuid.New(),
		UserID: userId,
		Value:  value,
	}
}
