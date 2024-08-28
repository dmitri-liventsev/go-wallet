package services_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	"wallet/transaction/internal/domain/vo"
)

var _ = Describe("Balance Correction", func() {

	Context("Balance is zero", func() {
		var (
			balanceRepo     *repositories.BalanceRepository
			transactionRepo *repositories.TransactionRepository
		)

		BeforeEach(func() {
			balanceRepo = repositories.NewBalanceRepository(DB)
			transactionRepo = repositories.NewTransactionRepository(DB)

			balance := entities.NewBalance(vo.NewTotalAmount(0))
			err := balanceRepo.Save(balance)
			Expect(err).ToNot(HaveOccurred())
		})
		Context("no done transactions are available", func() {
			When("correction are processed", func() {
				BeforeEach(func() {
					err := services.NewCorrectionProcessor(DB).Execute()
					Expect(err).ToNot(HaveOccurred())
				})

				It("balance remains unchanged", func() {
					balance, err := balanceRepo.Get()
					Expect(err).ToNot(HaveOccurred())
					Expect(balance.Value.Cents).To(Equal(int64(0)))
				})

				It("no new transactions are added", func() {
					transactions, err := transactionRepo.FindAll()
					Expect(err).ToNot(HaveOccurred())
					Expect(transactions).To(HaveLen(0))
				})
			})
		})

		Context("one transaction with positive amount exists", func() {
			BeforeEach(func() {
				_ = createDoneTransaction(10)
			})

			When("correction are processed", func() {
				var transactions []entities.Transaction

				BeforeEach(func() {
					err := services.NewCorrectionProcessor(DB).Execute()
					Expect(err).ToNot(HaveOccurred())

					transactions, err = transactionRepo.FindAll()
					Expect(err).ToNot(HaveOccurred())
				})

				It("balance remains unchanged", func() {
					balance, err := balanceRepo.Get()
					Expect(err).ToNot(HaveOccurred())
					Expect(balance.Value.Cents).To(Equal(int64(0)))
				})

				It("new transaction was added", func() {
					Expect(transactions).To(HaveLen(2))
				})

				It("correction transaction has negative amount", func() {
					Expect(transactions[1].Amount.Value()).To(Equal(-10))
				})
			})

		})

		Context("one transaction with negative amount exists", func() {
			var transactions []entities.Transaction

			BeforeEach(func() {
				_ = createDoneTransaction(-10)
			})

			When("correction are processed", func() {
				BeforeEach(func() {
					err := services.NewCorrectionProcessor(DB).Execute()
					Expect(err).ToNot(HaveOccurred())

					transactions, err = transactionRepo.FindAll()
					Expect(err).ToNot(HaveOccurred())
				})

				It("balance remains unchanged", func() {
					balance, err := balanceRepo.Get()
					Expect(err).ToNot(HaveOccurred())
					Expect(balance.Value.Cents).To(Equal(int64(0)))
				})

				It("new transaction was added", func() {
					Expect(transactions).To(HaveLen(2))
				})

				It("correction transaction has positive amount", func() {
					Expect(transactions[1].Amount.Value()).To(Equal(10))
				})
			})

		})

		Context("three transactions are exists", func() {
			var transactions []entities.Transaction

			BeforeEach(func() {
				_ = createDoneTransaction(-10)
				_ = createDoneTransaction(-1)
				_ = createDoneTransaction(-100)
			})

			When("correction are processed", func() {
				BeforeEach(func() {
					err := services.NewCorrectionProcessor(DB).Execute()
					Expect(err).ToNot(HaveOccurred())

					transactions, err = transactionRepo.FindAll()
					Expect(err).ToNot(HaveOccurred())
				})

				It("new transaction was added", func() {
					Expect(transactions).To(HaveLen(4))
				})

				It("correction transaction has sum of odds transactions", func() {
					Expect(transactions[3].Amount.Value()).To(Equal(110))
				})
			})
		})
	})
})
