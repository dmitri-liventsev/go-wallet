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
	txdb "wallet/transaction/internal/infrastructure/db"
)

// RunCorrectionWorker starts a background goroutine that continuously executes the correction worker, handling
// transactions and rolling back on errors until the context is done.
// RunCorrectionWorker запускает рабочий процесс для коррекции
func RunCorrectionWorker(ctx context.Context, db *gorm.DB) {
	// lests generate new correction if it does not exist,
	// we can swallow error here because even at it returns error, we will try again after
	_, _ = services.NewCorrectionProvider(db).Provide()

	go func(ctx context.Context) {
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
						err := NewCorrectionWorker(db).Execute()
						if err != nil {
							log.Errorf(ctx, err, "Cannot execute correction worker")
							continue
						}
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
	Saver    CorrectionSaver
	LockUuid uuid.UUID
	db       *gorm.DB
}

// Execute retrieves the newest correction and, if none exists or the latest correction was processed more than
// 10 minutes ago, initializes a new correction.
func (c CorrectionWorker) Execute() error {
	tx := c.db.Begin()
	correctionProvider := services.NewCorrectionProvider(tx)
	correctionRepository := repositories.NewCorrectionRepository(tx)

	err := correctionRepository.Lock(c.LockUuid)
	if err != nil {
		tx.Rollback()
		return err
	}
	correction, err := correctionProvider.Provide()

	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "cant get provider")
	}

	if correction.LockUuid == nil || *correction.LockUuid != c.LockUuid {
		_ = txdb.TryTxCommit(tx)
		return nil
	}

	err = services.NewCorrectionProcessor(tx).Execute()
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "cant execute processor")
	}

	err = c.unLock(correction, correctionRepository)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "cant unlock correction")
	}

	_ = txdb.TryTxCommit(tx)

	return nil
}

func (c CorrectionWorker) unLock(correction *entities.Correction, repo *repositories.CorrectionRepository) error {
	now := time.Now()
	correction.DoneAt = &now
	correction.Status = entities.Ready
	correction.LockUuid = nil

	err := repo.Save(correction)
	if err != nil {
		return err
	}

	return nil
}

// NewCorrectionWorker returns CorrectionWorker instance.
func NewCorrectionWorker(db *gorm.DB) CorrectionWorker {
	//correctionRepository := repositories.NewCorrectionRepository(db)

	return CorrectionWorker{
		LockUuid: uuid.New(),
		db:       db,
	}
}
