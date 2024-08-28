package entities

import (
	"github.com/google/uuid"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"time"
	"wallet/transaction/internal/domain/vo"
)

// Win action.
const Win = "win"

// Lost action.
const Lost = "lost"

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
	Status     string     `gorm:"type:varchar(10);check:status IN ('new','done','cancelled', 'locked');index"`
	SourceType string     `gorm:"type:varchar(10);check:source_type IN ('game','server','payment', 'internal')"`
	Action     string     `gorm:"type:varchar(10);check:action IN ('win','lost')"`
	Amount     vo.Amount  `gorm:"type:integer"`
	LockUuid   *uuid.UUID `gorm:"type:uuid;default:null"`
	LockedAt   *time.Time `gorm:"type:timestamptz;default:null"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;default:current_timestamp;index"`
	UpdatedAt  time.Time  `gorm:"type:timestamptz;default:current_timestamp"`
}

// MarkAsDone mark Transaction as done.
func (t *Transaction) MarkAsDone() {
	t.Status = Done
	t.LockUuid = nil
}

// MarkAsCancelled mark Transaction as cancelled
func (t *Transaction) MarkAsCancelled() {
	t.Status = Cancelled
}

func (t *Transaction) IsInternal() bool {
	return t.Status == Internal
}

// Lock book transaction by worker process
func (t *Transaction) Lock(lockUuid uuid.UUID) {
	t.Status = Locked
	t.LockUuid = &lockUuid
	*t.LockedAt = time.Now()
}

// NewTransaction returns new Transaction entity.
func NewTransaction(id string, amount vo.Amount, action string, sourceType string) *Transaction {
	return &Transaction{
		Status:     New,
		Action:     action,
		SourceType: sourceType,
		Amount:     amount,
		ID:         id,
	}
}
