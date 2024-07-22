package entities

import (
	"fmt"
	"github.com/google/uuid"
	"time"
	"wallet/transaction/internal/domain/vo"
)

// Correction represents the Correction entity, which is a scheduled operation for reversing a list of transactions.
type Correction struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey"`
	TransactionIDs vo.TransactionIds `gorm:"type:jsonb"`
	DoneAt         *time.Time
	CreatedAt      time.Time `gorm:"type:timestamptz;default:current_timestamp"`
	UpdatedAt      time.Time `gorm:"type:timestamptz;default:current_timestamp"`
}

// MarkAsDone mark correction as done.
func (c *Correction) MarkAsDone() {
	now := time.Now()
	c.DoneAt = &now
}

// GetTransactionIDs returns transaction ids.
func (c *Correction) GetTransactionIDs() ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for _, idStr := range c.TransactionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID string %s: %w", idStr, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// IsOutOFDate returns true if correction is too old
func (c *Correction) IsOutOFDate() bool {
	if c.DoneAt == nil {
		return false
	}

	doneUnix := c.DoneAt.Unix()
	nowUnix := time.Now().Unix()

	return nowUnix-doneUnix > 600
}

// NewCorrection returns new Correction entity instance.
func NewCorrection(transactionIDs []string) *Correction {
	return &Correction{
		ID:             uuid.New(),
		TransactionIDs: transactionIDs,
	}
}
