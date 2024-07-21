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
				responseCode int
				repo         *repositories.TransactionRepository
			)

			BeforeEach(func() {
				responseCode = createTx(uuid.New(), 10, entities.Win)
				repo = repositories.NewTransactionRepository(DB)
			})

			It("should successfully save the transaction", func() {
				transaction, err := repo.GetNextTransaction()

				Expect(transaction).ToNot(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})

			It("correct http code should be returned", func() {
				Expect(responseCode).To(Equal(202))
			})
		})
	})
})
