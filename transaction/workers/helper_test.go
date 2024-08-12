package workers_test

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

func createTransactionWithStatus(amount int, status string) *entities.Transaction {
	GinkgoHelper()

	action := entities.Win
	if amount < 0 {
		action = entities.Lost
		amount *= -1
	}

	transactionRepo := repositories.NewTransactionRepository(DB)
	transaction := entities.NewTransaction(uuid.New().String(), vo.NewAmount(amount), action, entities.Game)
	transaction.Status = status

	err := transactionRepo.Save(transaction)
	Expect(err).ToNot(HaveOccurred())

	return transaction
}
