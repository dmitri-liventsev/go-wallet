package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"goa.design/clue/log"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
)

const (
	correctionWorkerCycleDelay = 100 * time.Millisecond
	correctionWorkerRetryDelay = time.Second
	correctionWorkerFatalDelay = 5 * time.Second
)

// RunCorrectionWorker starts a background goroutine that continuously executes the correction worker.
// It returns a *WorkerMetrics pointer that is safe to read concurrently at any time.
func RunCorrectionWorker(ctx context.Context, db *gorm.DB) *WorkerMetrics {
	// Seed the first correction if none exists; error is intentionally ignored —
	// the worker will retry on the next cycle.
	_, _ = services.NewCorrectionProvider(db).Provide()

	m := &WorkerMetrics{}

	go func(ctx context.Context) {
		lockUuid := uuid.New()
		timer := time.NewTimer(0)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				// Reset is called only here, right after reading from timer.C,
				// so the channel is always drained — no stale-tick drain needed.
				start := time.Now()
				err := executeCorrectionCycle(ctx, db, lockUuid)
				m.LastDurationMs.Store(time.Since(start).Milliseconds())
				m.TotalCycles.Add(1)

				var delay time.Duration
				switch {
				case err == nil:
					m.SuccessCycles.Add(1)
					delay = correctionWorkerCycleDelay
				case errors.Is(err, ErrNoWork):
					m.NoWorkCycles.Add(1)
					delay = correctionWorkerCycleDelay
				case errors.Is(err, ErrLockConflict):
					m.ConflictCycles.Add(1)
					delay = correctionWorkerCycleDelay
				case errors.Is(err, errFatalCycle):
					m.FailedCycles.Add(1)
					log.Printf(ctx, "correction worker fatal failure lockUuid=%s metrics=[%s]", lockUuid, m)
					delay = correctionWorkerFatalDelay
				default:
					m.FailedCycles.Add(1)
					log.Printf(ctx, "correction worker retryable failure lockUuid=%s metrics=[%s]", lockUuid, m)
					delay = correctionWorkerRetryDelay
				}

				timer.Reset(delay)
			}
		}
	}(ctx)

	return m
}

func executeCorrectionCycle(ctx context.Context, db *gorm.DB, lockUuid uuid.UUID) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf(ctx, "panic recovered in correction worker lockUuid=%s: %v", lockUuid, r)
			err = fmt.Errorf("%w: %v", errFatalCycle, r)
		}
	}()

	tx := db.Begin()
	if tx.Error != nil {
		log.Errorf(ctx, tx.Error, "cannot begin correction worker transaction lockUuid=%s", lockUuid)
		return tx.Error
	}

	execErr := NewCorrectionWorker(tx, lockUuid).Execute()
	if execErr != nil && !errors.Is(execErr, ErrNoWork) && !errors.Is(execErr, ErrLockConflict) {
		tx.Rollback()
		log.Errorf(ctx, execErr, "cannot execute correction worker lockUuid=%s", lockUuid)
		return execErr
	}

	// ErrNoWork and ErrLockConflict are soft outcomes: commit to preserve any lock
	// updates made by the locker before returning without an error delay.
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Errorf(ctx, err, "cannot commit correction worker transaction lockUuid=%s", lockUuid)
		return err
	}

	return nil
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
		return fmt.Errorf("cannot provide correction: %w", err)
	}

	if correction == nil {
		return ErrNoWork
	}

	if correction.LockUuid == nil || *correction.LockUuid != c.LockUuid {
		return ErrLockConflict
	}

	if err = c.Processor.Execute(); err != nil {
		return fmt.Errorf("cannot execute correction processor: %w", err)
	}

	if err = c.unLock(correction); err != nil {
		return fmt.Errorf("cannot unlock correction: %w", err)
	}

	return nil
}

func (c CorrectionWorker) unLock(correction *entities.Correction) error {
	now := time.Now()
	correction.DoneAt = &now
	correction.Status = entities.Ready
	correction.LockUuid = nil

	return c.Saver.Save(correction)
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
