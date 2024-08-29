package workers_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	"wallet/transaction/internal/domain/vo"
	"wallet/transaction/workers"
)

var _ = Describe("balance worker processing", func() {
	Context("an unprocessed transactions not exists", func() {
		When("the worker starts", func() {
			var (
				balanceWorker   workers.BalanceWorker
				startBalance    vo.TotalAmount
				balanceProvider services.BalanceProvider
			)

			BeforeEach(func() {
				balanceWorker = workers.NewBalanceWorker(DB)

				balanceProvider = services.NewBalanceProvider(DB)
				balance, err := balanceProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				startBalance = balance.Value

				err = balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("balance should not be changed", func() {
				newBalance, err := balanceProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				Expect(newBalance.Value.String()).To(Equal(startBalance.String()))
			})
		})
	})

	Context("an unprocessed transaction exists", func() {
		BeforeEach(func() {
			_ = createTransaction(1)
		})

		When("the worker starts", func() {
			var (
				balanceWorker   workers.BalanceWorker
				startBalance    vo.TotalAmount
				balanceProvider services.BalanceProvider
			)

			BeforeEach(func() {
				balanceProvider = services.NewBalanceProvider(DB)
				balance, err := balanceProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				startBalance = balance.Value

				balanceWorker = workers.NewBalanceWorker(DB)

				err = balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("balance should be updated", func() {
				newBalance, err := balanceProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				Expect(newBalance.Value.String()).ToNot(Equal(startBalance.String()))
			})
		})
	})

	Context("transaction was locked by another process", func() {
		var lockUuid uuid.UUID

		BeforeEach(func() {
			lockUuid = uuid.New()
			createLockedTransaction(&lockUuid)
		})

		When("the worker starts", func() {
			var (
				balanceWorker   workers.BalanceWorker
				startBalance    vo.TotalAmount
				balanceProvider services.BalanceProvider
			)

			BeforeEach(func() {
				balanceProvider = services.NewBalanceProvider(DB)
				balance, err := balanceProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				startBalance = balance.Value

				balanceWorker = workers.NewBalanceWorker(DB)

				err = balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("balance should not be changed", func() {
				newBalance, err := balanceProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				Expect(newBalance.Value.String()).To(Equal(startBalance.String()))
			})
		})
	})

	Context("transaction was locked by same process before", func() {
		var (
			lockUuid uuid.UUID
		)

		BeforeEach(func() {
			lockUuid = uuid.New()
			_ = createLockedTransaction(&lockUuid)
		})

		When("the worker starts", func() {
			var (
				balanceWorker   workers.BalanceWorker
				startBalance    vo.TotalAmount
				balanceProvider services.BalanceProvider
			)

			BeforeEach(func() {
				balanceProvider = services.NewBalanceProvider(DB)
				balance, err := balanceProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				startBalance = balance.Value

				balanceWorker = workers.NewBalanceWorker(DB)
				balanceWorker.LockUuid = lockUuid

				err = balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("balance should be changed", func() {
				newBalance, err := balanceProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				Expect(newBalance.Value.String()).ToNot(Equal(startBalance.String()))
			})
		})
	})

	Context("three transaction are locked, transaction in the middle are locked by another process", func() {
		var (
			lockUuid     uuid.UUID
			transaction1 *entities.Transaction
			transaction2 *entities.Transaction
			transaction3 *entities.Transaction
		)

		BeforeEach(func() {
			lockUuid = uuid.New()
			transaction1 = createLockedTransaction(&lockUuid)
			randomUuid := uuid.New()
			transaction2 = createLockedTransaction(&randomUuid)
			transaction3 = createLockedTransaction(&lockUuid)
		})

		When("the worker starts", func() {
			var (
				balanceWorker         workers.BalanceWorker
				transactionRepository *repositories.TransactionRepository
			)

			BeforeEach(func() {
				balanceWorker = workers.NewBalanceWorker(DB)
				balanceWorker.LockUuid = lockUuid

				transactionRepository = repositories.NewTransactionRepository(DB)

				err := balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("only first transaction should be processed", func() {
				var err error
				transaction1, err = transactionRepository.FindByID(transaction1.ID)
				Expect(err).ToNot(HaveOccurred())
				transaction2, err = transactionRepository.FindByID(transaction2.ID)
				Expect(err).ToNot(HaveOccurred())
				transaction3, err = transactionRepository.FindByID(transaction3.ID)
				Expect(err).ToNot(HaveOccurred())

				Expect(transaction1.Status).To(Equal(entities.Done))
				Expect(transaction2.Status).To(Equal(entities.Locked))
				Expect(transaction3.Status).To(Equal(entities.Locked))
			})
		})
	})
})
