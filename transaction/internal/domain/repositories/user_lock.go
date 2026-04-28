package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserLockRepository struct {
	db *gorm.DB
}

func (r *UserLockRepository) TryLockUser(ctx context.Context, userId int64, uuid uuid.UUID, ttl time.Duration) (bool, error) {
	res := r.db.WithContext(ctx).Exec(`
		INSERT INTO user_lock (user_id, lock_uuid, expires_at, updated_at)
		VALUES (?, ?, NOW() + INTERVAL '? seconds', NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET 
		    lock_uuid = EXCLUDED.lock_uuid, 
		    expires_at =  EXCLUDED.expires_at,
		    updated_at = NOW()
	`, userId, uuid, int(ttl.Seconds()))

	if res.Error != nil {
		return false, res.Error
	}

	return res.RowsAffected > 0, nil
}

func (r *UserLockRepository) UnlockUser(ctx context.Context, userId int64, uuid uuid.UUID) error {
	return r.db.WithContext(ctx).
		Exec(`
				DELETE FROM user_lock WHERE user_id = ? AND lock_uuid = ?
		`, userId, uuid).Error
}

// NewUserLockRepository returns UserLockRepository instance.
func NewUserLockRepository(db *gorm.DB) *UserLockRepository {
	return &UserLockRepository{db: db}
}
