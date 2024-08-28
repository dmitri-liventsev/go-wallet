package workers

import (
	"context"
	"github.com/google/uuid"
	"goa.design/clue/log"
	"gorm.io/gorm"
	"time"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	txdb "wallet/transaction/internal/infrastructure/db"
)

// RunCorrectionWorker starts a background goroutine that continuously executes the correction worker, handling
// transactions and rolling back on errors until the context is done.
func RunCorrectionWorker(ctx context.Context, db *gorm.DB) {
	// lests generate new correction if it does not exist,
	// we can swallow error here because even at it returns error, we will try again after
	_, _ = services.NewCorrectionProvider(db).Provide()

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

// CorrectionProvider provide correction
type CorrectionProvider interface {
	Provide() (*entities.Correction, error)
}

// CorrectionLocker lock correction
type CorrectionLocker interface {
	Lock(lockUuid uuid.UUID) error
}

// CorrectionProcessor doing correction
type CorrectionProcessor interface {
	Execute() error
}

// CorrectionSaver saves correction
type CorrectionSaver interface {
	Save(correction *entities.Correction) error
}

// CorrectionWorker monitors the creation of the latest correction and adds a new one if more than 10 minutes have
// passed since the last correction was processed.
type CorrectionWorker struct {
	CorrectionProvider  CorrectionProvider
	CorrectionProcessor CorrectionProcessor
	Locker              CorrectionLocker
	Saver               CorrectionSaver
	LockUuid            uuid.UUID
}

// Execute retrieves the newest correction and, if none exists or the latest correction was processed more than
// 10 minutes ago, initializes a new correction.
func (c CorrectionWorker) Execute() error {
	err := c.Locker.Lock(c.LockUuid)
	if err != nil {
		return err
	}
	correction, err := c.CorrectionProvider.Provide()
	if err != nil {
		return err
	}

	if *correction.LockUuid != c.LockUuid {
		return nil
	}

	err = c.CorrectionProcessor.Execute()
	if err != nil {
		return err
	}

	err = c.unLock(correction)
	if err != nil {
		return err
	}

	return nil
}

func (c CorrectionWorker) unLock(correction *entities.Correction) error {
	now := time.Now()
	correction.DoneAt = &now
	correction.LockUuid = nil

	err := c.Saver.Save(correction)
	if err != nil {
		return err
	}

	return nil
}

// NewCorrectionWorker returns CorrectionWorker instance.
func NewCorrectionWorker(db *gorm.DB) CorrectionWorker {
	correctionRepository := repositories.NewCorrectionRepository(db)

	return CorrectionWorker{
		CorrectionProvider:  services.NewCorrectionProvider(db),
		CorrectionProcessor: services.NewCorrectionProcessor(db),
		Locker:              correctionRepository,
		Saver:               correctionRepository,
		LockUuid:            uuid.New(),
	}
}
