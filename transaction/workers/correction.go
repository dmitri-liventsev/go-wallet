package workers

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"goa.design/clue/log"
	"gorm.io/gorm"
	"time"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
)

// RunCorrectionWorker starts a background goroutine that continuously executes the correction worker, handling
// transactions and rolling back on errors until the context is done.
// RunCorrectionWorker запускает рабочий процесс для коррекции
func RunCorrectionWorker(ctx context.Context, db *gorm.DB) {
	// lests generate new correction if it does not exist,
	// we can swallow error here because even at it returns error, we will try again after
	_, _ = services.NewCorrectionProvider(db).Provide()

	go func(ctx context.Context) {
		lockUuid := uuid.New()
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf(ctx, "Recovered from panic")
					}
				}()

				for {
					select {
					case <-ctx.Done():
						return
					default:
						tx := db.Begin()
						err := NewCorrectionWorker(tx, lockUuid).Execute()
						if err != nil {
							tx.Rollback()
							log.Errorf(ctx, err, "Cannot execute correction worker")
							continue
						}

						_ = tx.Commit()
					}
				}
			}()
		}
	}(ctx)
}

// CorrectionSaver saves correction
type CorrectionSaver interface {
	Save(correction *entities.Correction) error
}

// CorrectionWorker monitors the creation of the latest correction and adds a new one if more than 10 minutes have
// passed since the last correction was processed.
type CorrectionWorker struct {
	Saver     CorrectionSaver
	LockUuid  uuid.UUID
	Provider  CorrectionProvider
	Locker    CorrectionLocker
	Processor CorrectionProcessor
}

type CorrectionProvider interface {
	Provide() (*entities.Correction, error)
}

type CorrectionLocker interface {
	Lock(lockUuid uuid.UUID) error
}

type CorrectionProcessor interface {
	Execute() error
}

// Execute retrieves the newest correction and, if none exists or the latest correction was processed more than
// 10 minutes ago, initializes a new correction.
func (c CorrectionWorker) Execute() error {
	err := c.Locker.Lock(c.LockUuid)
	if err != nil {
		return err
	}
	correction, err := c.Provider.Provide()

	if err != nil {
		return errors.Wrap(err, "cant get provider")
	}

	if correction.LockUuid == nil || *correction.LockUuid != c.LockUuid {
		return nil
	}

	err = c.Processor.Execute()
	if err != nil {
		return errors.Wrap(err, "cant execute processor")
	}

	err = c.unLock(correction)
	if err != nil {
		return errors.Wrap(err, "cant unlock correction")
	}

	return nil
}

func (c CorrectionWorker) unLock(correction *entities.Correction) error {
	now := time.Now()
	correction.DoneAt = &now
	correction.Status = entities.Ready
	correction.LockUuid = nil

	err := c.Saver.Save(correction)
	if err != nil {
		return err
	}

	return nil
}

// NewCorrectionWorker returns CorrectionWorker instance.
func NewCorrectionWorker(db *gorm.DB, lockUuid uuid.UUID) CorrectionWorker {
	correctionRepository := repositories.NewCorrectionRepository(db)

	return CorrectionWorker{
		LockUuid:  lockUuid,
		Provider:  services.NewCorrectionProvider(db),
		Locker:    correctionRepository,
		Saver:     correctionRepository,
		Processor: services.NewCorrectionProcessor(db),
	}
}
