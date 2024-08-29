package tests

import (
	"context"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/gen/transaction"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"
)

var _ = Describe("transaction management", func() {
	var payload *transaction.CreatePayload
	var repo *repositories.TransactionRepository

	Context("no transactions exist in the system", func() {
		BeforeEach(func() {
			payload = &transaction.CreatePayload{
				State:         entities.Win,
				Amount:        "0",
				TransactionID: uuid.New().String(),
				SourceType:    entities.Game,
			}

			repo = repositories.NewTransactionRepository(DB)
		})
		When("a transaction with 0 amount received", func() {
			BeforeEach(func(ctx context.Context) {
				err := client.Create(ctx, payload)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should ignore the transaction", func() {
				transaction, err := repo.GetNextTransaction()

				Expect(transaction).To(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("a signal to create a transaction with positive amount is received", func() {
			BeforeEach(func(ctx context.Context) {
				payload.Amount = "10.01"
				err := client.Create(ctx, payload)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should save transaction data correctly", func() {
				transaction, err := repo.GetNextTransaction()

				Expect(transaction).ToNot(BeNil())
				Expect(err).NotTo(HaveOccurred())

				Expect(transaction.ID).To(Equal(payload.TransactionID))
				Expect(transaction.Amount.Cents).To(Equal(1001))
				Expect(transaction.Status).To(Equal(entities.New))
				Expect(transaction.Action).To(Equal(payload.State))
				Expect(transaction.SourceType).To(Equal(payload.SourceType))
			})
		})

		When("a signal to create a lost transaction with positive amount is received", func() {
			var err error
			BeforeEach(func(ctx context.Context) {
				payload.Amount = "10.01"
				payload.State = "lost"
				err = client.Create(ctx, payload)

			})

			It("error should return", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		When("a signal to create a transaction with negative amount is received", func() {
			BeforeEach(func(ctx context.Context) {
				payload.Amount = "-10.01"
				payload.State = "lost"
				err := client.Create(ctx, payload)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should save transaction data correctly", func() {
				transaction, err := repo.GetNextTransaction()

				Expect(transaction).ToNot(BeNil())
				Expect(err).NotTo(HaveOccurred())

				Expect(transaction.ID).To(Equal(payload.TransactionID))
				Expect(transaction.Amount.Cents).To(Equal(-1001))
				Expect(transaction.Status).To(Equal(entities.New))
				Expect(transaction.Action).To(Equal(payload.State))
				Expect(transaction.SourceType).To(Equal(payload.SourceType))
			})
		})

		When("a signal to create a win transaction with negative amount is received", func() {
			var err error
			BeforeEach(func(ctx context.Context) {
				payload.Amount = "-10.01"
				payload.State = "win"
				err = client.Create(ctx, payload)

			})

			It("error should return", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("a transaction are exists", func() {
		var existedTransaction *entities.Transaction
		BeforeEach(func() {
			payload = &transaction.CreatePayload{
				State:         entities.Win,
				Amount:        "10.01",
				TransactionID: uuid.New().String(),
				SourceType:    entities.Game,
			}

			repo = repositories.NewTransactionRepository(DB)

			existedTransaction = entities.NewTransaction(uuid.New().String(), vo.NewAmount(10), entities.Win, entities.Game)
			err := repo.Save(existedTransaction)
			Expect(err).NotTo(HaveOccurred())
		})

		When("a signals to create a transaction with same ID is received", func() {
			BeforeEach(func(ctx context.Context) {
				payload.TransactionID = existedTransaction.ID
				err := client.Create(ctx, payload)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should only save the first transaction and ignore subsequent ones", func() {
				transactions, err := repo.GetAllTransactions()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(transactions)).To(Equal(1))
			})
		})
	})

	Context("a cancelled transaction exists", func() {
		var existedTransaction *entities.Transaction
		BeforeEach(func() {
			payload = &transaction.CreatePayload{
				State:         entities.Win,
				Amount:        "10.01",
				TransactionID: uuid.New().String(),
				SourceType:    entities.Game,
			}

			repo = repositories.NewTransactionRepository(DB)

			existedTransaction = entities.NewTransaction(uuid.New().String(), vo.NewAmount(10), entities.Win, entities.Game)
			existedTransaction.MarkAsCancelled()
			err := repo.Save(existedTransaction)
			Expect(err).NotTo(HaveOccurred())
		})

		When("a signals to create a transaction with same ID is received", func() {
			BeforeEach(func(ctx context.Context) {
				payload.TransactionID = existedTransaction.ID
				err := client.Create(ctx, payload)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should ignore the new transaction", func() {
				transactions, err := repo.GetAllTransactions()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(transactions)).To(Equal(1))
			})
		})
	})
})
