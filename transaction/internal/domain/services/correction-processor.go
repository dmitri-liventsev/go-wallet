package services

import (
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
)

// CorrectionProcessor handles the processing of corrections.
type CorrectionProcessor struct {
	TxRepo         *repositories.TransactionRepository
	CorrectionRepo *repositories.CorrectionRepository
	BalanceService *Balance
}

// Execute processes the provided correction by finding and canceling transactions listed in the correction,
// updating the balance for each transaction.
func (c CorrectionProcessor) Execute(correction *entities.Correction) error {
	if len(correction.TransactionIDs) > 0 {
		err := c.doCorrection(correction.TransactionIDs)
		if err != nil {
			return err
		}
	}

	correction.MarkAsDone()
	err := c.CorrectionRepo.Save(correction)
	if err != nil {
		return err
	}

	return nil
}

func (c CorrectionProcessor) doCorrection(transactionIDs []string) error {
	transactions, err := c.TxRepo.FindByIDs(transactionIDs)
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		err := c.BalanceService.ForceUpdateBalance(transaction.Amount.Inverse())
		if err != nil {
			return err
		}

		transaction.MarkAsCancelled()
		err = c.TxRepo.Save(&transaction)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewCorrectionProcessor returns CorrectionProcessor instance.
func NewCorrectionProcessor(db *gorm.DB) CorrectionProcessor {
	return CorrectionProcessor{
		TxRepo:         repositories.NewTransactionRepository(db),
		CorrectionRepo: repositories.NewCorrectionRepository(db),
		BalanceService: NewBalanceService(db),
	}
}
