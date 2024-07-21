package service_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/service"
	"wallet/transaction/internal/domain/vo"
)

var _ = Describe("Balance Correction", func() {

	Context("Balance is zero", func() {
		var (
			balanceRepo    *repositories.BalanceRepository
			correctionRepo *repositories.CorrectionRepository
		)

		BeforeEach(func() {
			balanceRepo = repositories.NewBalanceRepository(DB)
			correctionRepo = repositories.NewCorrectionRepository(DB)

			balance := entities.NewBalance(vo.NewTotalAmount(0))
			err := balanceRepo.Save(balance)
			Expect(err).ToNot(HaveOccurred())
		})
		Context("the correction contains an empty list of transactions", func() {
			var correction *entities.Correction

			BeforeEach(func() {
				correction = entities.NewCorrection([]string{})
				err := correctionRepo.Save(correction)
				Expect(err).ToNot(HaveOccurred())
			})

			When("correction are processed", func() {
				BeforeEach(func() {
					err := service.NewCorrectionProcessor(DB).Execute(correction)
					Expect(err).ToNot(HaveOccurred())
				})

				It("the balance remains unchanged", func() {
					balance, err := balanceRepo.Get()
					Expect(err).ToNot(HaveOccurred())
					Expect(balance.Value.Cents).To(Equal(int64(0)))
				})

				It("the correction is marked as processed", func() {
					var err error
					correction, err = correctionRepo.GetNewestCorrection()
					Expect(err).ToNot(HaveOccurred())
					Expect(correction.DoneAt).ToNot(BeNil())
				})
			})
		})

		Context("the correction contains one transaction with a positive amount", func() {
			var correction *entities.Correction

			BeforeEach(func() {
				transaction := createDoneTransaction(10)
				correction = entities.NewCorrection([]string{transaction.ID})
				err := correctionRepo.Save(correction)
				Expect(err).ToNot(HaveOccurred())
			})

			When("correction are processed", func() {
				BeforeEach(func() {
					err := service.NewCorrectionProcessor(DB).Execute(correction)
					Expect(err).ToNot(HaveOccurred())
				})

				It("the balance decreases by the amount and becomes negative", func() {
					balance, err := balanceRepo.Get()
					Expect(err).ToNot(HaveOccurred())
					Expect(balance.Value.Cents).To(Equal(int64(-10)))
				})

				It("the correction is marked as processed", func() {
					var err error
					correction, err = correctionRepo.GetNewestCorrection()
					Expect(err).ToNot(HaveOccurred())
					Expect(correction.DoneAt).ToNot(BeNil())
				})
			})

		})

		Context("the correction contains one transaction with a negative amount", func() {
			var correction *entities.Correction

			BeforeEach(func() {
				transaction := createDoneTransaction(-10)
				correction = entities.NewCorrection([]string{transaction.ID})
				err := correctionRepo.Save(correction)
				Expect(err).ToNot(HaveOccurred())
			})

			When("correction are processed", func() {
				BeforeEach(func() {
					err := service.NewCorrectionProcessor(DB).Execute(correction)
					Expect(err).ToNot(HaveOccurred())
				})

				It("the balance increases by the amount", func() {
					balance, err := balanceRepo.Get()
					Expect(err).ToNot(HaveOccurred())
					Expect(balance.Value.Cents).To(Equal(int64(10)))
				})

				It("the correction is marked as processed", func() {
					var err error
					correction, err = correctionRepo.GetNewestCorrection()
					Expect(err).ToNot(HaveOccurred())
					Expect(correction.DoneAt).ToNot(BeNil())
				})
			})

		})

		Context("the correction contains two transactions", func() {
			var correction *entities.Correction

			BeforeEach(func() {
				transaction := createDoneTransaction(-10)
				transaction2 := createDoneTransaction(-1)
				correction = entities.NewCorrection([]string{transaction.ID, transaction2.ID})
				err := correctionRepo.Save(correction)
				Expect(err).ToNot(HaveOccurred())
			})

			When("correction are processed", func() {
				BeforeEach(func() {
					err := service.NewCorrectionProcessor(DB).Execute(correction)
					Expect(err).ToNot(HaveOccurred())
				})

				It("the balance decreases by the sum of all amounts", func() {
					balance, err := balanceRepo.Get()
					Expect(err).ToNot(HaveOccurred())
					Expect(balance.Value.Cents).To(Equal(int64(11)))
				})

				It("the correction is marked as processed", func() {
					var err error
					correction, err = correctionRepo.GetNewestCorrection()
					Expect(err).ToNot(HaveOccurred())
					Expect(correction.DoneAt).ToNot(BeNil())
				})
			})
		})
	})
})
