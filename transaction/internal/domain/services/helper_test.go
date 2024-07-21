package services_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"
)

func createTransaction(amount int) *entities.Transaction {
	GinkgoHelper()
	return createTransactionWithStatus(amount, entities.New)
}

func createCancelledTransaction(amount int) *entities.Transaction {
	GinkgoHelper()
	return createTransactionWithStatus(amount, entities.Cancelled)
}

func createDoneTransaction(amount int) *entities.Transaction {
	GinkgoHelper()
	return createTransactionWithStatus(amount, entities.Done)
}

func createTransactionWithStatus(amount int, status string) *entities.Transaction {
	GinkgoHelper()

	action := entities.Win
	if amount < 0 {
		action = entities.Lost
	}

	transactionRepo := repositories.NewTransactionRepository(DB)
	transaction := entities.NewTransaction(uuid.New().String(), vo.NewAmount(amount), action, entities.Game)
	transaction.Status = status

	err := transactionRepo.Save(transaction)
	Expect(err).ToNot(HaveOccurred())

	return transaction
}
