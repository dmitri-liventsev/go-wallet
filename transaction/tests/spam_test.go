package tests

import (
	"context"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"math/rand"
	"time"
	"wallet/gen/transaction"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/workers"
)

var _ = Describe("transaction management", func() {
	var payloads []*transaction.CreatePayload
	var repo *repositories.TransactionRepository
	var balanceRepo *repositories.BalanceRepository
	var expectedBalance int64
	var numOfTransactions = 10000

	Context("asynchronous request processing", func() {
		BeforeEach(func() {
			payloads, expectedBalance = generatePayloads(numOfTransactions)
			repo = repositories.NewTransactionRepository(DB)
			balanceRepo = repositories.NewBalanceRepository(DB)
		})

		When("all payloads are sends", func() {
			BeforeEach(func(ctx context.Context) {
				for _, payload := range payloads {
					err := client.Create(ctx, payload)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("saves each trasnactions", func() {
				transactions, err := repo.GetAllTransactions()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(transactions)).To(Equal(1))
			})

			When("transactions are processed", func() {
				BeforeEach(func(ctx context.Context) {
					worker := workers.NewBalanceWorker(DB)
					for i := 0; i < numOfTransactions; i++ {
						err := worker.Execute()
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("balance are correct2", func() {
					balance, err := balanceRepo.Get()
					Expect(err).NotTo(HaveOccurred())

					Expect(balance.Value.Cents).To(Equal(expectedBalance * 100))
				})
			})
		})
	})
})

func generatePayloads(limit int) ([]*transaction.CreatePayload, int64) {
	payloads := make([]*transaction.CreatePayload, 0, limit)
	expectedBalance := int64(0)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < limit; i++ {
		state := entities.Win

		randomBool := rand.Intn(2) == 1
		if randomBool {
			state = entities.Lost
			if expectedBalance-10 >= 0 {
				expectedBalance -= 10
			}
		} else {
			expectedBalance += 10
		}

		payloads = append(payloads, &transaction.CreatePayload{
			State:         state,
			Amount:        "10",
			TransactionID: uuid.New().String(),
			SourceType:    entities.Game,
		})
	}

	return payloads, expectedBalance
}
