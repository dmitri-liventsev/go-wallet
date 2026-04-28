package workers_test

import (
	"context"
	"time"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func createTransaction(ctx context.Context) *entities.Transaction {
	GinkgoHelper()
	return createTransactionWithStatus(ctx, 10, entities.New)
}

func createDoneTransaction(ctx context.Context) *entities.Transaction {
	GinkgoHelper()
	return createTransactionWithStatus(ctx, 10, entities.Done)
}

func createLockedTransaction(ctx context.Context, lockId *uuid.UUID) *entities.Transaction {
	GinkgoHelper()
	transaction := createTransactionWithStatus(ctx, 10, entities.Locked)
	transaction.LockUuid = lockId
	now := time.Now()
	transaction.LockedAt = &now

	err := repositories.NewTransactionRepository(DB).Save(ctx, transaction)
	Expect(err).ToNot(HaveOccurred())

	return transaction
}

func createTransactionWithStatus(ctx context.Context, amount int, status string) *entities.Transaction {
	GinkgoHelper()

	action := entities.Win
	if amount < 0 {
		action = entities.Lost
		amount *= -1
	}

	transactionRepo := repositories.NewTransactionRepository(DB)
	transaction := entities.NewTransaction(uuid.New().String(), vo.NewAmount(amount), action, entities.Game, 1)
	transaction.Status = status

	err := transactionRepo.Save(ctx, transaction)
	Expect(err).ToNot(HaveOccurred())

	return transaction
}
