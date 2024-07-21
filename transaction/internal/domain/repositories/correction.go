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

// GetNewestCorrection returns new newest correction fo LIFO processing.
func (repo CorrectionRepository) GetNewestCorrection() (*entities.Correction, error) {
	var correction entities.Correction

	if err := repo.db.Order("created_at DESC").Limit(1).First(&correction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &correction, nil
}

// NewCorrectionRepository returns CorrectionRepository instance.
func NewCorrectionRepository(db *gorm.DB) *CorrectionRepository {
	return &CorrectionRepository{db: db}
}
