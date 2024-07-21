package entities

import (
	"github.com/google/uuid"
	"wallet/transaction/internal/domain/vo"
)

// BalanceID the balance entry id.
const BalanceID = "0f31adad-bfb6-41d1-aeff-c110ca13cbfa" //Hardcoded %(

// Balance represents the Balance entity, responsible for storing the current state of the balance.
type Balance struct {
	ID    uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Value vo.TotalAmount `gorm:"type:bigint;not null"`
}

// NewBalance returns new Balance entity instance.
func NewBalance(value vo.TotalAmount) *Balance {
	return &Balance{
		ID:    uuid.MustParse(BalanceID),
		Value: value,
	}
}
