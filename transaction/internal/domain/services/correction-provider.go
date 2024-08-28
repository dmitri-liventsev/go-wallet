package services

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
)

const CorrectionId = "3d8e7990-7a74-4613-9ed4-154dbba1d3b5"

// CorrectionProvider checks at correction are exist otherwise creating it
type CorrectionProvider struct {
	correctionRepo *repositories.CorrectionRepository
}

func (c CorrectionProvider) Provide() (*entities.Correction, error) {
	correctionId := uuid.MustParse(CorrectionId)
	correction, err := c.correctionRepo.FindByID(correctionId)
	if err != nil {
		return nil, err
	}

	if correction != nil {
		return correction, nil
	}

	correction = entities.NewCorrection(correctionId)
	err = c.correctionRepo.Save(correction)
	if err != nil {
		return nil, err
	}

	return correction, nil
}

// NewCorrectionProvider returns new CorrectionProvider instance
func NewCorrectionProvider(db *gorm.DB) CorrectionProvider {
	return CorrectionProvider{
		correctionRepo: repositories.NewCorrectionRepository(db),
	}
}
