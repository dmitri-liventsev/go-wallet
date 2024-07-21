package db

import (
	"gorm.io/gorm"
)

// TryTxCommit attempts to commit a transaction up to three times, rolling back and retrying in case of failure.
func TryTxCommit(tx *gorm.DB) error {
	var err error
	for attempts := 0; attempts < 3; attempts++ {
		if err = tx.Commit().Error; err == nil {
			return nil
		}

		tx.Rollback()
	}

	return err
}
