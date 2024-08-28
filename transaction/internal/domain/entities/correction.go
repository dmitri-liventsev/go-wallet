package entities

import (
	"github.com/google/uuid"
	"time"
)

const Ready = "ready"

// Correction represents the Correction entity, which is a scheduled operation for reversing a list of transactions.
type Correction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	DoneAt    *time.Time
	Status    string     `gorm:"type:varchar(10);check:status IN ('ready', 'locked');index"`
	LockUuid  *uuid.UUID `gorm:"type:uuid;default:null"`
	LockedAt  *time.Time `gorm:"type:timestamptz;default:null"`
	CreatedAt time.Time  `gorm:"type:timestamptz;default:current_timestamp"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:current_timestamp"`
}

// MarkAsDone mark correction as done.
func (c *Correction) MarkAsDone() {
	now := time.Now()
	c.DoneAt = &now
	c.LockUuid = nil
}

func (c *Correction) Lock(lockUuid uuid.UUID) {
	c.Status = Locked
	c.LockUuid = &lockUuid
	*c.LockedAt = time.Now()
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
func NewCorrection(id uuid.UUID) *Correction {
	now := time.Now()
	lockUuid := uuid.New()
	return &Correction{
		ID:     id,
		Status: Ready,
		// avoid correction on service starting:
		LockedAt: &now,
		LockUuid: &lockUuid,
	}
}
