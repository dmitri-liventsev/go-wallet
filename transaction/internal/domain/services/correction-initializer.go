package services

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
)

// CorrectionInitializer handles the creation and initialization of new corrections,
// using repositories for transactions and corrections.
type CorrectionInitializer struct {
	txRepo         *repositories.TransactionRepository
	correctionRepo *repositories.CorrectionRepository
}

// Execute retrieves the last 10 odd-numbered transactions, creates a new correction with their IDs,
// and saves the correction to the repository.
func (c CorrectionInitializer) Execute() error {
	doomedTransactions, err := c.txRepo.GetLastOddTransactions(10)
	if err != nil {
		return errors.Wrap(err, "couldn't get last odd transactions")
	}

	ids := make([]string, len(doomedTransactions))
	for i, tx := range doomedTransactions {
		ids[i] = tx.ID
	}

	correction := entities.NewCorrection(ids)

	err = c.correctionRepo.Save(correction)
	if err != nil {
		return errors.Wrap(err, "couldn't save correction")
	}

	return nil
}

// NewCorrectionInitializer returns CorrectionInitializer instance.
func NewCorrectionInitializer(db *gorm.DB) CorrectionInitializer {
	return CorrectionInitializer{
		txRepo:         repositories.NewTransactionRepository(db),
		correctionRepo: repositories.NewCorrectionRepository(db),
	}
}
