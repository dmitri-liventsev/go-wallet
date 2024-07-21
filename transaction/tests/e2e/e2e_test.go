package e2e_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
)

var _ = Describe("transaction сreation", func() {
	Context("there are no existing transactions in the system", func() {
		When("a signal to create a transaction is received", func() {
			var (
				responseCode  int
				repo          *repositories.TransactionRepository
				transactionId string
			)

			BeforeEach(func() {
				transactionId = uuid.New().String()
				responseCode = createTx(transactionId, 10, entities.Win)
				repo = repositories.NewTransactionRepository(DB)
			})

			It("should successfully save the transaction", func() {
				transaction, err := repo.GetNextTransaction()

				Expect(transaction.ID).To(Equal(transactionId))
				Expect(err).NotTo(HaveOccurred())
			})

			It("correct http code should be returned", func() {
				Expect(responseCode).To(Equal(202))
			})
		})
	})
})
