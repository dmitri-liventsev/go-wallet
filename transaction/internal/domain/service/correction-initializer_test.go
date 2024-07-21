package service_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/service"
)

var _ = Describe("correction initializer", func() {
	var (
		correctionInitializer service.CorrectionInitializer
		//transactionRepo       *repositories.TransactionRepository
		correctionnRepo *repositories.CorrectionRepository
	)

	BeforeEach(func() {
		correctionInitializer = service.NewCorrectionInitializer(DB)
		//transactionRepo = repositories.NewTransactionRepository(DB)
		correctionnRepo = repositories.NewCorrectionRepository(DB)
	})

	Describe("running the correction initializer works", func() {
		Context("with no transactions in the system", func() {
			When("the correction worker runs", func() {
				var correction *entities.Correction
				BeforeEach(func() {
					err := correctionInitializer.Execute()
					Expect(err).ToNot(HaveOccurred())
					correction, err = correctionnRepo.GetNewestCorrection()
					Expect(err).ToNot(HaveOccurred())
				})

				It("should create a new unprocessed correction with an empty list of transaction IDs", func() {
					Expect(correction.DoneAt).To(BeNil())
					Expect(len(correction.TransactionIDs)).To(Equal(0))
				})
			})
		})

		Context("when only cancelled transactions exist in the system", func() {
			BeforeEach(func() {
				createCancelledTransaction(10)
			})

			When("the correction worker runs", func() {
				var correction *entities.Correction
				BeforeEach(func() {
					err := correctionInitializer.Execute()
					Expect(err).ToNot(HaveOccurred())
					correction, err = correctionnRepo.GetNewestCorrection()
					Expect(err).ToNot(HaveOccurred())
				})

				It("should create a new unprocessed correction with an empty list of transaction IDs", func() {
					Expect(correction.DoneAt).To(BeNil())
					Expect(len(correction.TransactionIDs)).To(Equal(0))
				})
			})
		})

		Context("with two processed transactions in the system", func() {
			var oddTransaction *entities.Transaction

			BeforeEach(func() {
				createDoneTransaction(10)
				oddTransaction = createDoneTransaction(10)
			})

			When("the correction worker runs", func() {
				var correction *entities.Correction

				BeforeEach(func() {
					err := correctionInitializer.Execute()
					Expect(err).ToNot(HaveOccurred())
					correction, err = correctionnRepo.GetNewestCorrection()
					Expect(err).ToNot(HaveOccurred())
				})

				It("should create a new unprocessed correction with a list containing only the odd transaction", func() {
					Expect(correction.DoneAt).To(BeNil())
					Expect(len(correction.TransactionIDs)).To(Equal(1))
					Expect(correction.TransactionIDs[0]).To(Equal(oddTransaction.ID))
				})
			})
		})

		Context("with more than 20 processed transactions in the system", func() {
			var oddTransactionIds [10]string

			BeforeEach(func() {
				createDoneTransaction(10)
				createDoneTransaction(10)
				oddTransactionIds[0] = createDoneTransaction(10).ID
				createDoneTransaction(10)
				oddTransactionIds[1] = createDoneTransaction(10).ID
				createDoneTransaction(10)
				oddTransactionIds[2] = createDoneTransaction(10).ID
				createDoneTransaction(10)
				oddTransactionIds[3] = createDoneTransaction(10).ID
				createDoneTransaction(10)
				oddTransactionIds[4] = createDoneTransaction(10).ID
				createDoneTransaction(10)
				oddTransactionIds[5] = createDoneTransaction(10).ID
				createDoneTransaction(10)
				oddTransactionIds[6] = createDoneTransaction(10).ID
				createDoneTransaction(10)
				oddTransactionIds[7] = createDoneTransaction(10).ID
				createDoneTransaction(10)
				oddTransactionIds[8] = createDoneTransaction(10).ID
				createDoneTransaction(10)
				oddTransactionIds[9] = createDoneTransaction(10).ID
			})

			When("the correction worker runs", func() {
				var correction *entities.Correction

				BeforeEach(func() {
					err := correctionInitializer.Execute()
					Expect(err).ToNot(HaveOccurred())
					correction, err = correctionnRepo.GetNewestCorrection()
					Expect(err).ToNot(HaveOccurred())
				})

				It("should create a new unprocessed correction with a list containing only 10 odd transactions", func() {
					Expect(len(correction.TransactionIDs)).To(Equal(10))
				})

				It("should contain only odd transactions", func() {
					transactionIds := correction.TransactionIDs

					Expect(oddTransactionIds).To(ContainElement(transactionIds[0]))
					Expect(oddTransactionIds).To(ContainElement(transactionIds[1]))
					Expect(oddTransactionIds).To(ContainElement(transactionIds[2]))
					Expect(oddTransactionIds).To(ContainElement(transactionIds[3]))
					Expect(oddTransactionIds).To(ContainElement(transactionIds[4]))
					Expect(oddTransactionIds).To(ContainElement(transactionIds[5]))
					Expect(oddTransactionIds).To(ContainElement(transactionIds[6]))
					Expect(oddTransactionIds).To(ContainElement(transactionIds[7]))
					Expect(oddTransactionIds).To(ContainElement(transactionIds[8]))
					Expect(oddTransactionIds).To(ContainElement(transactionIds[9]))
				})
			})
		})
	})
})
