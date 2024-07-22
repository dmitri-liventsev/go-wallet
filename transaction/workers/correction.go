package workers

import (
	"context"
	"github.com/pkg/errors"
	"goa.design/clue/log"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	txdb "wallet/transaction/internal/infrastructure/db"
)

// RunCorrectionWorker starts a background goroutine that continuously executes the correction worker, handling
// transactions and rolling back on errors until the context is done.
func RunCorrectionWorker(ctx context.Context, db *gorm.DB) {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				tx := db.Begin()
				err := NewCorrectionWorker(tx).Execute()
				if err != nil {
					log.Fatalf(ctx, err, "cannot Execute correction worker")

					tx.Rollback()
				}
				_ = txdb.TryTxCommit(tx)
			}
		}
	}(ctx)
}

// CorrectionInitializer create a new correction order
type CorrectionInitializer interface {
	Execute() error
}

// CorrectionWorker monitors the creation of the latest correction and adds a new one if more than 10 minutes have
// passed since the last correction was processed.
type CorrectionWorker struct {
	CorrectionInitializer CorrectionInitializer
	cRepo                 CorrectionProvider
}

// Execute retrieves the newest correction and, if none exists or the latest correction was processed more than
// 10 minutes ago, initializes a new correction.
func (c CorrectionWorker) Execute() error {
	newestCorrection, err := c.cRepo.GetNewestCorrection()
	if err != nil {
		return errors.Wrap(err, "cannot GetNewestCorrection")
	}

	// If there are no correction yet (system just started) or if newest correction are processed already,
	// and it was more than 10 min ago
	if newestCorrection == nil || newestCorrection.IsOutOFDate() {
		err := c.CorrectionInitializer.Execute()
		if err != nil {
			return errors.Wrap(err, "cant create a new correction")
		}
	}

	return nil
}

// NewCorrectionWorker returns CorrectionWorker instance.
func NewCorrectionWorker(db *gorm.DB) CorrectionWorker {
	return CorrectionWorker{
		CorrectionInitializer: services.NewCorrectionInitializer(db),
		cRepo:                 repositories.NewCorrectionRepository(db),
	}
}
