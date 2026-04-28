package workers

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"

	"github.com/google/uuid"
	"goa.design/clue/log"
	"gorm.io/gorm"
)

const (
	balanceWorkerCycleDelay = 100 * time.Millisecond
	balanceWorkerRetryDelay = time.Second
	balanceWorkerFatalDelay = 5 * time.Second

	userLockTTL       = 30 * time.Second
	usersBatchSize    = 50
	transactionsLimit = 100
)

var (
	ErrNoWork       = errors.New("no work available")
	ErrLockConflict = errors.New("user lock owned by another worker")

	errFatalCycle = errors.New("fatal cycle")
)

type WorkerMetrics struct {
	TotalCycles    atomic.Int64
	SuccessCycles  atomic.Int64
	FailedCycles   atomic.Int64
	NoWorkCycles   atomic.Int64
	ConflictCycles atomic.Int64
	LastDurationMs atomic.Int64
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

	worker := NewBalanceWorker(db, lockUuid)
	return worker.Execute(ctx)
}

type UserLocker interface {
	TryLockUser(ctx context.Context, userID int64, lockUUID uuid.UUID, ttl time.Duration) (bool, error)
	UnlockUser(ctx context.Context, userID int64, lockUUID uuid.UUID) error
}

type UserProvider interface {
	GetUsersWithNewTransactions(ctx context.Context, limit int) ([]int64, error)
	GetUserTransactions(ctx context.Context, userID int64, limit int) ([]entities.Transaction, error)
}

type Processor interface {
	Execute(ctx context.Context, transaction *entities.Transaction) error
}

type BalanceWorker struct {
	LockUuid     uuid.UUID
	UserLocker   UserLocker
	UserProvider UserProvider
	Processor    Processor
}

func (b BalanceWorker) Execute(ctx context.Context) error {
	userIDs, err := b.UserProvider.GetUsersWithNewTransactions(ctx, usersBatchSize)
	if err != nil {
		return err
	}

	if len(userIDs) == 0 {
		return ErrNoWork
	}

	for _, userID := range userIDs {
		ok, err := b.UserLocker.TryLockUser(ctx, userID, b.LockUuid, userLockTTL)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		// IMPORTANT: process user in transaction
		err = b.processUser(ctx, userID)
		if err != nil {
			return err
		}

		// optional unlock (TTL fallback exists anyway)
		_ = b.UserLocker.UnlockUser(ctx, userID, b.LockUuid)

		return nil // process only one user per cycle (important for fairness)
	}

	return ErrLockConflict
}

func (b BalanceWorker) processUser(ctx context.Context, userID int64) error {
	txList, err := b.UserProvider.GetUserTransactions(ctx, userID, transactionsLimit)
	if err != nil {
		return err
	}

	if len(txList) == 0 {
		return nil
	}

	for _, t := range txList {
		if err := b.Processor.Execute(ctx, &t); err != nil {
			return err
		}
	}

	return nil
}

func NewBalanceWorker(db *gorm.DB, lockUuid uuid.UUID) BalanceWorker {
	transactionRepository := repositories.NewTransactionRepository(db)
	userLockRepository := repositories.NewUserLockRepository(db)

	return BalanceWorker{
		LockUuid:     lockUuid,
		UserLocker:   userLockRepository,
		UserProvider: transactionRepository,
		Processor:    services.NewTransactionProcessor(db),
	}
}
