package services

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"
)

// CorrectionProcessor handles the creation and initialization of new corrections,
// using repositories for transactions and corrections.
type CorrectionProcessor struct {
	txRepo         *repositories.TransactionRepository
	correctionRepo *repositories.CorrectionRepository
}

// Execute retrieves the last 10 odd-numbered transactions cancel it and add new transaction with inversed sum of cancelled transactions
func (c CorrectionProcessor) Execute() error {
	doomedTransactions, err := c.txRepo.GetLastOddTransactions(10)
	if err != nil {
		return errors.Wrap(err, "couldn't get last odd transactions")
	}

	ids := make([]string, len(doomedTransactions))
	delta := vo.NewAmount(0)
	for i, tx := range doomedTransactions {
		ids[i] = tx.ID
		delta = delta.Add(tx.Amount)
		tx.MarkAsCancelled()
		err := c.txRepo.Save(&tx)
		if err != nil {
			return errors.Wrap(err, "unable to save doomed transaction")
		}
	}

	if delta.IsZero() {
		return nil
	}

	delta = delta.Inverse()
	action := entities.Win
	if delta.LessThenZero() {
		action = entities.Lost
	}

	correctionTransaction := entities.NewTransaction(uuid.New().String(), delta, action, entities.Internal)
	err = c.txRepo.Save(correctionTransaction)
	if err != nil {
		return errors.Wrap(err, "unable to save correction transaction")
	}

	return nil
}

// NewCorrectionInitializer returns CorrectionProcessor instance.
func NewCorrectionProcessor(db *gorm.DB) CorrectionProcessor {
	return CorrectionProcessor{
		txRepo:         repositories.NewTransactionRepository(db),
		correctionRepo: repositories.NewCorrectionRepository(db),
	}
}
