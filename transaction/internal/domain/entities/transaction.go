package entities

import (
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

// New status.
const New = "new"

// Done status.
const Done = "done"

// Cancelled status.
const Cancelled = "cancelled"

// Transaction represents the Transaction entity, which stores all incoming requests for changing the user's balance.
type Transaction struct {
	ID         string    `gorm:"type:varchar(128);primaryKey"`
	Status     string    `gorm:"type:varchar(10);check:status IN ('new','done','cancelled');index"`
	SourceType string    `gorm:"type:varchar(10);check:source_type IN ('game','server','payment')"`
	Action     string    `gorm:"type:varchar(10);check:action IN ('win','lost')"`
	Amount     vo.Amount `gorm:"type:integer"`
	CreatedAt  time.Time `gorm:"type:timestamptz;default:current_timestamp;index"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;default:current_timestamp"`
}

// MarkAsDone mark Transaction as done.
func (t *Transaction) MarkAsDone() {
	t.Status = "done"
}

// MarkAsCancelled mark Transaction as cancelled
func (t *Transaction) MarkAsCancelled() {
	t.Status = "cancelled"
}

// NewTransaction returns new Transaction entity.
func NewTransaction(id string, amount vo.Amount, action string, sourceType string) *Transaction {
	return &Transaction{
		Status:     "new",
		Action:     action,
		SourceType: sourceType,
		Amount:     amount,
		ID:         id,
	}
}
