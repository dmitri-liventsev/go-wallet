package tests

import (
	"context"
	"errors"
	"math/rand"
	"time"
	"wallet/gen/wallet"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/workers"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("transaction management", func() {
	var payloads []*wallet.CreateTransactionPayload
	var repo *repositories.TransactionRepository
	var balanceRepo *repositories.BalanceRepository
	var expectedBalance int64
	var numOfTransactions = 1000

	Context("asynchronous request processing", func() {
		BeforeEach(func() {
			payloads, expectedBalance = generatePayloads(numOfTransactions)
			repo = repositories.NewTransactionRepository(DB)
			balanceRepo = repositories.NewBalanceRepository(DB)
		})

		When("all payloads are sends", func() {
			BeforeEach(func(ctx context.Context) {
				for _, payload := range payloads {
					err := client.CreateTransaction(ctx, payload)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("saves each trasnactions", func(ctx context.Context) {
				transactions, err := repo.GetAllTransactions(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(transactions)).To(Equal(numOfTransactions))
			})

			When("transactions are processed", func() {
				BeforeEach(func(ctx context.Context) {
					worker := workers.NewBalanceWorker(DB, uuid.New())
					for i := 0; i < numOfTransactions; i++ {
						err := worker.Execute(ctx)
						if errors.Is(err, workers.ErrNoWork) {
							break
						}
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("balance are correct2", func(ctx context.Context) {
					balance, err := balanceRepo.Get(ctx, 1)
					Expect(err).NotTo(HaveOccurred())

					Expect(balance.Value.Cents).To(Equal(expectedBalance * 100))
				})
			})
		})
	})
})

func generatePayloads(limit int) ([]*wallet.CreateTransactionPayload, int64) {
	payloads := make([]*wallet.CreateTransactionPayload, 0, limit)
	expectedBalance := int64(0)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < limit; i++ {
		state := entities.Win

		randomBool := rand.Intn(2) == 1
		amount := "10"
		if randomBool {
			state = entities.Lost
			if expectedBalance-10 >= 0 {
				expectedBalance -= 10
			}

			amount = "10"
		} else {
			expectedBalance += 10
		}

		payloads = append(payloads, &wallet.CreateTransactionPayload{
			State:         state,
			Amount:        amount,
			UserID:        1,
			TransactionID: uuid.New().String(),
			SourceType:    entities.Game,
		})
	}

	return payloads, expectedBalance
}
