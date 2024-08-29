package services

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
)

// CorrectionProvider checks at correction are exist otherwise creating it
type CorrectionProvider struct {
	correctionRepo *repositories.CorrectionRepository
}

func (c CorrectionProvider) Provide() (*entities.Correction, error) {
	correction, err := c.correctionRepo.Get()
	if err != nil {
		return nil, errors.Wrap(err, "cannot provide correction")
	}

	if correction != nil {
		return correction, nil
	}
	correction = entities.NewCorrection(uuid.MustParse(entities.CorrectionId))

	err = c.correctionRepo.Save(correction)
	if err != nil {
		return nil, errors.Wrap(err, "cannot save correction")
	}

	return correction, nil
}

// NewCorrectionProvider returns new CorrectionProvider instance
func NewCorrectionProvider(db *gorm.DB) CorrectionProvider {
	return CorrectionProvider{
		correctionRepo: repositories.NewCorrectionRepository(db),
	}
}
