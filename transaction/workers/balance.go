package workers

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"goa.design/clue/log"
	"gorm.io/gorm"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
)

const (
	balanceWorkerCycleDelay = 100 * time.Millisecond
	balanceWorkerRetryDelay = time.Second     // transient DB / processing error
	balanceWorkerFatalDelay = 5 * time.Second // panic or unexpected system failure
)

var (
	// ErrNoWork and ErrLockConflict are soft outcomes — normal-speed retry, no error log.
	// They are exported so callers that invoke Execute() directly (e.g. tests, one-off tools)
	// can distinguish them from real failures.
	ErrNoWork       = errors.New("no work available")
	ErrLockConflict = errors.New("lock owned by another worker")

	// errFatalCycle wraps panics. Triggers a longer backoff to avoid thrashing after
	// an unexpected runtime failure. Kept unexported — it is an internal loop concern.
	errFatalCycle = errors.New("fatal cycle")
)

// WorkerMetrics holds lightweight in-memory counters for the balance worker.
// All fields are safe for concurrent access via atomic operations.
type WorkerMetrics struct {
	TotalCycles    atomic.Int64
	SuccessCycles  atomic.Int64
	FailedCycles   atomic.Int64 // hard failures only (not no-work or conflict)
	NoWorkCycles   atomic.Int64
	ConflictCycles atomic.Int64
	LastDurationMs atomic.Int64 // wall-clock duration of the most recent cycle
}

func (m *WorkerMetrics) String() string {
	return fmt.Sprintf(
		"total=%d success=%d failed=%d no_work=%d conflict=%d last_ms=%d",
		m.TotalCycles.Load(),
		m.SuccessCycles.Load(),
		m.FailedCycles.Load(),
		m.NoWorkCycles.Load(),
		m.ConflictCycles.Load(),
		m.LastDurationMs.Load(),
	)
}

// RunBalanceWorker starts a background goroutine that continuously executes the balance worker.
// It returns a *WorkerMetrics pointer that is safe to read concurrently at any time.
func RunBalanceWorker(ctx context.Context, db *gorm.DB) *WorkerMetrics {
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
				err := executeBalanceCycle(ctx, db, lockUuid)
				m.LastDurationMs.Store(time.Since(start).Milliseconds())
				m.TotalCycles.Add(1)

				var delay time.Duration
				switch {
				case err == nil:
					m.SuccessCycles.Add(1)
					delay = balanceWorkerCycleDelay
				case errors.Is(err, ErrNoWork):
					m.NoWorkCycles.Add(1)
					delay = balanceWorkerCycleDelay
				case errors.Is(err, ErrLockConflict):
					m.ConflictCycles.Add(1)
					delay = balanceWorkerCycleDelay
				case errors.Is(err, errFatalCycle):
					m.FailedCycles.Add(1)
					log.Printf(ctx, "balance worker fatal failure lockUuid=%s metrics=[%s]", lockUuid, m)
					delay = balanceWorkerFatalDelay
				default:
					m.FailedCycles.Add(1)
					log.Printf(ctx, "balance worker retryable failure lockUuid=%s metrics=[%s]", lockUuid, m)
					delay = balanceWorkerRetryDelay
				}

				timer.Reset(delay)
			}
		}
	}(ctx)

	return m
}

func executeBalanceCycle(ctx context.Context, db *gorm.DB, lockUuid uuid.UUID) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf(ctx, "panic recovered in balance worker lockUuid=%s: %v", lockUuid, r)
			err = fmt.Errorf("%w: %v", errFatalCycle, r)
		}
	}()

	tx := db.Begin()
	if tx.Error != nil {
		log.Errorf(ctx, tx.Error, "cannot begin balance worker transaction lockUuid=%s", lockUuid)
		return tx.Error
	}

	execErr := NewBalanceWorker(tx, lockUuid).Execute()
	if execErr != nil && !errors.Is(execErr, ErrNoWork) && !errors.Is(execErr, ErrLockConflict) {
		tx.Rollback()
		log.Errorf(ctx, execErr, "cannot execute balance worker lockUuid=%s", lockUuid)
		return execErr
	}

	// ErrNoWork and ErrLockConflict are soft outcomes: commit to preserve any lock
	// updates made by LockNewTransactions before returning without an error delay.
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Errorf(ctx, err, "cannot commit balance worker transaction lockUuid=%s", lockUuid)
		return err
	}

	return nil
}

type Locker interface {
	LockNewTransactions(lockUuid uuid.UUID) error
}

type Provider interface {
	GetLockedTransactions() ([]entities.Transaction, error)
}

type Processor interface {
	Execute(*entities.Transaction) error
}

// BalanceWorker is responsible for monitoring new correction requests, initiating correction processing,
// tracking new transactions, and initiating their processing.
type BalanceWorker struct {
	LockUuid  uuid.UUID
	Locker    Locker
	Provider  Provider
	Processor Processor
}

// Execute retrieves and processes the current correction if available,
// then retrieves and processes the next transaction.
func (b BalanceWorker) Execute() error {
	err := b.Locker.LockNewTransactions(b.LockUuid)
	if err != nil {
		return err
	}

	transactions, err := b.Provider.GetLockedTransactions()
	if err != nil {
		return err
	}

	if len(transactions) == 0 {
		return ErrNoWork
	}

	for _, transaction := range transactions {
		if *transaction.LockUuid != b.LockUuid {
			return ErrLockConflict
		}

		if err = b.Processor.Execute(&transaction); err != nil {
			return err
		}
	}

	return nil
}

// NewBalanceWorker returns BalanceWorker instance.
func NewBalanceWorker(db *gorm.DB, lockUuid uuid.UUID) BalanceWorker {
	transactionRepository := repositories.NewTransactionRepository(db)

	return BalanceWorker{
		LockUuid:  lockUuid,
		Locker:    transactionRepository,
		Provider:  transactionRepository,
		Processor: services.NewTransactionProcessor(db),
	}
}
