package entities

import (
	"time"

	"github.com/google/uuid"
)

type UserLock struct {
	UserId    int64      `gorm:"primary_key;column:user_id"`
	LockUuid  *uuid.UUID `gorm:"type:uuid;default:null;column:lock_uuid"`
	ExpiresAt time.Time  `gorm:"type:timestamp;default:null"`
	UpdatedAt time.Time  `gorm:"type:timestamp;default:null"`
}

// NewTransaction returns new Transaction entity.
func NewUserLock(userId int64, lockUuid *uuid.UUID, expiresAt time.Time, updatedAt time.Time) *UserLock {
	return &UserLock{
		UserId:    userId,
		LockUuid:  lockUuid,
		ExpiresAt: expiresAt,
		UpdatedAt: updatedAt,
	}
}
