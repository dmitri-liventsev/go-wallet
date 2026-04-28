package services_test

import (
	"context"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	"wallet/transaction/internal/domain/vo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction service", func() {
	var (
		transactionProcessor services.TransactionProcessor
		transactionRepo      *repositories.TransactionRepository
		balanceRepo          *repositories.BalanceRepository
	)

	BeforeEach(func() {
		transactionRepo = repositories.NewTransactionRepository(DB)
		balanceRepo = repositories.NewBalanceRepository(DB)
		transactionProcessor = services.NewTransactionProcessor(DB)
	})

	Describe("Processing transactions", func() {
		Context("the current balance is zero", func() {
			Context("an unprocessed transaction with positive amount are exist", func() {
				var transaction *entities.Transaction

				BeforeEach(func(ctx context.Context) {
					transaction = createTransaction(ctx, 10)
				})

				When("transaction are procesed", func() {
					BeforeEach(func(ctx context.Context) {
						err := transactionProcessor.Execute(ctx, transaction)
						Expect(err).ToNot(HaveOccurred())
					})

					It("balance should be increased to the amount", func(ctx context.Context) {
						balance, err := balanceRepo.Get(ctx, 1)
						Expect(err).ToNot(HaveOccurred())
						Expect(balance.Value.Cents).To(Equal(int64(10)))
					})
				})
			})

			Context("an unprocessed transaction with negative amount are exist", func() {
				var (
					transaction *entities.Transaction
					err         error
				)

				BeforeEach(func(ctx context.Context) {
					transaction = createTransaction(ctx, -10)
				})

				When("transaction are procesed", func() {
					BeforeEach(func(ctx context.Context) {
						err := transactionProcessor.Execute(ctx, transaction)
						Expect(err).ToNot(HaveOccurred())
					})

					It("balance should not be decreased to the amount", func(ctx context.Context) {
						balance, err := balanceRepo.Get(ctx, 1)
						Expect(err).ToNot(HaveOccurred())
						Expect(balance.Value.Cents).To(Equal(int64(0)))
					})

					It("transaction should be cancelled", func(ctx context.Context) {
						transaction, err = transactionRepo.FindByID(ctx, transaction.ID)
						Expect(err).ToNot(HaveOccurred())
						Expect(transaction.Status).To(Equal(entities.Cancelled))
					})
				})
			})
		})

		When("the current balance is 100", func() {
			BeforeEach(func(ctx context.Context) {
				balance := entities.NewBalance(vo.NewTotalAmount(100), 1)
				err := balanceRepo.Save(ctx, balance)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("an unprocessed transaction with negative amount are exist", func() {
				var (
					transaction *entities.Transaction
					err         error
				)

				BeforeEach(func(ctx context.Context) {
					transaction = createTransaction(ctx, -10)
				})

				When("transaction are procesed", func() {
					BeforeEach(func(ctx context.Context) {
						err := transactionProcessor.Execute(ctx, transaction)
						Expect(err).ToNot(HaveOccurred())
					})

					It("balance should be decreased to the amount", func(ctx context.Context) {
						balance, err := balanceRepo.Get(ctx, 1)
						Expect(err).ToNot(HaveOccurred())
						Expect(balance.Value.Cents).To(Equal(int64(90)))
					})

					It("transaction should be processed", func(ctx context.Context) {
						transaction, err = transactionRepo.FindByID(ctx, transaction.ID)
						Expect(err).ToNot(HaveOccurred())
						Expect(transaction.Status).To(Equal(entities.Done))
					})
				})
			})
		})
	})
})
