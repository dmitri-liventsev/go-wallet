package services_test

import (
	"context"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func createTransaction(ctx context.Context, amount int) *entities.Transaction {
	GinkgoHelper()
	return createTransactionWithStatus(ctx, amount, entities.New)
}

func createCancelledTransaction(ctx context.Context, amount int) *entities.Transaction {
	GinkgoHelper()
	return createTransactionWithStatus(ctx, amount, entities.Cancelled)
}

func createDoneTransaction(ctx context.Context, amount int) *entities.Transaction {
	GinkgoHelper()
	return createTransactionWithStatus(ctx, amount, entities.Done)
}

func createTransactionWithStatus(ctx context.Context, amount int, status string) *entities.Transaction {
	GinkgoHelper()

	action := entities.Win
	if amount < 0 {
		action = entities.Lost
	}

	transactionRepo := repositories.NewTransactionRepository(DB)
	transaction := entities.NewTransaction(uuid.New().String(), vo.NewAmount(amount), action, entities.Game, 1)
	transaction.Status = status

	err := transactionRepo.Save(ctx, transaction)
	Expect(err).ToNot(HaveOccurred())

	return transaction
}
