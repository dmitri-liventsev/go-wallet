package workers_test

import (
	"context"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	"wallet/transaction/internal/domain/vo"
	"wallet/transaction/workers"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("balance worker processing", func() {
	Context("an unprocessed transactions not exists", func() {
		When("the worker starts", func() {
			var (
				balanceWorker   workers.BalanceWorker
				startBalance    vo.TotalAmount
				balanceProvider services.BalanceProvider
			)

			BeforeEach(func(ctx context.Context) {
				balanceWorker = workers.NewBalanceWorker(DB, uuid.New())

				balanceProvider = services.NewBalanceProvider(DB)
				balance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				startBalance = balance.Value

				err = balanceWorker.Execute(ctx)
				Expect(err).To(Or(BeNil(), MatchError(workers.ErrNoWork), MatchError(workers.ErrLockConflict)))
			})

			It("balance should not be changed", func(ctx context.Context) {
				newBalance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(newBalance.Value.String()).To(Equal(startBalance.String()))
			})
		})
	})

	Context("an unprocessed transaction exists", func() {
		BeforeEach(func(ctx context.Context) {
			_ = createTransaction(ctx)
		})

		When("the worker starts", func() {
			var (
				balanceWorker   workers.BalanceWorker
				startBalance    vo.TotalAmount
				balanceProvider services.BalanceProvider
			)

			BeforeEach(func(ctx context.Context) {
				balanceProvider = services.NewBalanceProvider(DB)
				balance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				startBalance = balance.Value

				balanceWorker = workers.NewBalanceWorker(DB, uuid.New())

				err = balanceWorker.Execute(ctx)
				Expect(err).ToNot(HaveOccurred())
			})

			It("balance should be updated", func(ctx context.Context) {
				newBalance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(newBalance.Value.String()).ToNot(Equal(startBalance.String()))
			})
		})
	})

	Context("transaction was locked by another process", func() {
		var lockUuid uuid.UUID

		BeforeEach(func(ctx context.Context) {
			lockUuid = uuid.New()
			createLockedTransaction(ctx, &lockUuid)
		})

		When("the worker starts", func() {
			var (
				balanceWorker   workers.BalanceWorker
				startBalance    vo.TotalAmount
				balanceProvider services.BalanceProvider
			)

			BeforeEach(func(ctx context.Context) {
				balanceProvider = services.NewBalanceProvider(DB)
				balance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				startBalance = balance.Value

				balanceWorker = workers.NewBalanceWorker(DB, uuid.New())

				err = balanceWorker.Execute(ctx)
				Expect(err).To(Or(BeNil(), MatchError(workers.ErrNoWork), MatchError(workers.ErrLockConflict)))
			})

			It("balance should not be changed", func(ctx context.Context) {
				newBalance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(newBalance.Value.String()).To(Equal(startBalance.String()))
			})
		})
	})

	Context("user has a new transaction pending processing", func() {
		BeforeEach(func(ctx context.Context) {
			_ = createTransaction(ctx)
		})

		When("the worker starts with any lock uuid", func() {
			var (
				balanceWorker   workers.BalanceWorker
				startBalance    vo.TotalAmount
				balanceProvider services.BalanceProvider
			)

			BeforeEach(func(ctx context.Context) {
				balanceProvider = services.NewBalanceProvider(DB)
				balance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				startBalance = balance.Value

				balanceWorker = workers.NewBalanceWorker(DB, uuid.New())

				err = balanceWorker.Execute(ctx)
				Expect(err).ToNot(HaveOccurred())
			})

			It("balance should be changed", func(ctx context.Context) {
				newBalance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(newBalance.Value.String()).ToNot(Equal(startBalance.String()))
			})
		})
	})

	Context("user has multiple new transactions pending processing", func() {
		var (
			transaction1 *entities.Transaction
			transaction2 *entities.Transaction
			transaction3 *entities.Transaction
		)

		BeforeEach(func(ctx context.Context) {
			transaction1 = createTransaction(ctx)
			transaction2 = createTransaction(ctx)
			transaction3 = createTransaction(ctx)
		})

		When("the worker starts", func() {
			var (
				balanceWorker         workers.BalanceWorker
				transactionRepository *repositories.TransactionRepository
			)

			BeforeEach(func(ctx context.Context) {
				balanceWorker = workers.NewBalanceWorker(DB, uuid.New())
				transactionRepository = repositories.NewTransactionRepository(DB)

				err := balanceWorker.Execute(ctx)
				Expect(err).ToNot(HaveOccurred())
			})

			It("all transactions should be processed", func(ctx context.Context) {
				var err error
				transaction1, err = transactionRepository.FindByID(ctx, transaction1.ID)
				Expect(err).ToNot(HaveOccurred())
				transaction2, err = transactionRepository.FindByID(ctx, transaction2.ID)
				Expect(err).ToNot(HaveOccurred())
				transaction3, err = transactionRepository.FindByID(ctx, transaction3.ID)
				Expect(err).ToNot(HaveOccurred())

				Expect(transaction1.Status).To(Equal(entities.Done))
				Expect(transaction2.Status).To(Equal(entities.Done))
				Expect(transaction3.Status).To(Equal(entities.Done))
			})
		})
	})
})
