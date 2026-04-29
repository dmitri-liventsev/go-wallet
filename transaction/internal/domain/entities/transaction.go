package entities

import (
	"time"
	"wallet/transaction/internal/domain/vo"

	"github.com/google/uuid"
)

// Win action.
const Win = "win"

// Lost action.
const Lost = "lose"

// Game source type.
const Game = "game"

// Server source type.
const Server = "server"

// Payment source type.
const Payment = "payment"

// Internal source type
const Internal = "internal"

// New status.
const New = "new"

// Done status.
const Done = "done"

// Cancelled status.
const Cancelled = "cancelled"

// Locked status.
const Locked = "locked"

// Transaction represents the Transaction entity, which stores all incoming requests for changing the user's balance.
type Transaction struct {
	ID         string     `gorm:"type:varchar(128);primaryKey"`
	SourceType string     `gorm:"type:varchar(10);check:source_type IN ('game','server','payment', 'internal')"`
	Action     string     `gorm:"type:varchar(10);check:action IN ('win','lose')"`
	Amount     vo.Amount  `gorm:"type:integer"`
	UserID     uint64     `gorm:"type:integer;index"`
	Status     string     `gorm:"type:varchar(10);index;check:status IN ('new','done','cancelled', 'locked');index"`
	LockUuid   *uuid.UUID `gorm:"type:uuid;default:null"`
	LockedAt   *time.Time `gorm:"type:timestamptz;default:null"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;default:current_timestamp;index"`
	UpdatedAt  time.Time  `gorm:"type:timestamptz;default:current_timestamp"`
}

// MarkAsDone mark Transaction as done.
func (t *Transaction) MarkAsDone() {
	t.Status = Done
}

// MarkAsCancelled mark Transaction as cancelled
func (t *Transaction) MarkAsCancelled() {
	t.Status = Cancelled
}

// IsInternal returns true if the transaction originates from an internal source,
// bypassing the negative-balance check in the processor.
func (t *Transaction) IsInternal() bool {
	return t.SourceType == Internal
}

// NewTransaction returns new Transaction entity.
func NewTransaction(id string, amount vo.Amount, action string, sourceType string, userID uint64) *Transaction {
	return &Transaction{
		Status:     New,
		Action:     action,
		SourceType: sourceType,
		Amount:     amount,
		UserID:     userID,
		ID:         id,
	}
}
