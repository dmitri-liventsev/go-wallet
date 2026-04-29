package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"goa.design/clue/log"
	"gorm.io/gorm"
)

type UserLockRepository struct {
	db *gorm.DB
}

func (r *UserLockRepository) TryLockUser(
	ctx context.Context,
	userId int64,
	lockID uuid.UUID,
	ttl time.Duration,
) (bool, error) {

	expiresAt := time.Now().Add(ttl)

	res := r.db.WithContext(ctx).Exec(`
		INSERT INTO user_locks (user_id, lock_uuid, expires_at, updated_at)
		VALUES (?, ?, ?, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET lock_uuid = EXCLUDED.lock_uuid,
		    expires_at = EXCLUDED.expires_at,
		    updated_at = NOW()
		WHERE user_locks.expires_at < NOW()
		   OR user_locks.lock_uuid = EXCLUDED.lock_uuid
	`, userId, lockID, expiresAt)

	if res.Error != nil {
		log.Printf(ctx, "user_lock try_lock userID=%d lockID=%s result=error err=%v", userId, lockID, res.Error)
		return false, res.Error
	}

	locked := res.RowsAffected > 0
	log.Printf(ctx, "user_lock try_lock userID=%d lockID=%s rowsAffected=%d locked=%v", userId, lockID, res.RowsAffected, locked)
	return locked, nil
}

func (r *UserLockRepository) UnlockUser(ctx context.Context, userId int64, lockID uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Exec(`DELETE FROM user_locks WHERE user_id = ? AND lock_uuid = ?`, userId, lockID).Error
	log.Printf(ctx, "user_lock unlock userID=%d lockID=%s err=%v", userId, lockID, err)
	return err
}

// NewUserLockRepository returns UserLockRepository instance.
func NewUserLockRepository(db *gorm.DB) *UserLockRepository {
	return &UserLockRepository{db: db}
}
