package repositories

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
	"wallet/transaction/internal/domain/entities"
)

// CorrectionRepository correction repository.
type CorrectionRepository struct {
	db *gorm.DB
}

// Save saves the correction entity to the database.
func (repo CorrectionRepository) Save(correction *entities.Correction) error {
	correction.UpdatedAt = time.Now()

	return repo.db.Save(correction).Error
}

// FindByID finds a correction by its ID in the database and returns the correction entity.
func (repo CorrectionRepository) FindByID(id uuid.UUID) (*entities.Correction, error) {
	var correction entities.Correction
	if err := repo.db.First(&correction, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &correction, nil
}

// GetActualCorrection returns the oldest unprocessed correction for FIFO processing.
func (repo CorrectionRepository) GetActualCorrection() (*entities.Correction, error) {
	var correction entities.Correction
	if err := repo.db.Where("done_at IS NULL").Order("created_at ASC").First(&correction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &correction, nil
}

// Lock locks correction which is ready to execute or which is frozen
func (repo CorrectionRepository) Lock(lockUuid uuid.UUID) error {
	threshold1 := time.Now().Add(-10 * time.Minute)
	thresholdForFrozen := time.Now().Add(-10 * time.Minute)

	result := repo.db.Model(&entities.Correction{}).
		Where("(status = ? AND done_at < ?) OR locked_at < ?", "ready", threshold1, thresholdForFrozen).
		Updates(map[string]interface{}{
			"status":    "locked",
			"lock_uuid": lockUuid,
			"locked_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// FindAll returns all transactions
func (repo CorrectionRepository) FindAll() ([]entities.Correction, error) {
	var corrections []entities.Correction

	result := repo.db.Find(&corrections)
	if result.Error != nil {
		return nil, result.Error
	}

	return corrections, nil
}

// NewCorrectionRepository returns CorrectionRepository instance.
func NewCorrectionRepository(db *gorm.DB) *CorrectionRepository {
	return &CorrectionRepository{db: db}
}
