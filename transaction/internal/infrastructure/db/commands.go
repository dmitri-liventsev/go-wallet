package db

import (
	"wallet/transaction/internal/domain/entities"

	"gorm.io/gorm"
)

// Truncate clears all records from the existed tables in the database.
func Truncate(db *gorm.DB) {
	tables := []interface{}{&entities.Transaction{}, &entities.Balance{}, &entities.UserLock{}}
	for _, table := range tables {
		_ = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error
	}
}
