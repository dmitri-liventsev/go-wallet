package workers_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/workers"
	"wallet/transaction/workers/internal/mocks"
)

var _ = Describe("balance worker processing", func() {
	Context("an unprocessed correction not exists", func() {
		When("the worker starts", func() {
			var (
				balanceWorker            workers.BalanceWorker
				ctrl                     *gomock.Controller
				mockTransactionProcessor *mocks.MockTransactionProcessor
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockTransactionProcessor = mocks.NewMockTransactionProcessor(ctrl)
				balanceWorker = workers.NewBalanceWorker(DB)
				balanceWorker.TransactionProcessor = mockTransactionProcessor
			})

			It("the correctionProcessor should not be called", func() {
				mockTransactionProcessor.EXPECT().Execute(gomock.Any()).Times(0)
				err := balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("an unprocessed transaction exists", func() {
		var transaction *entities.Transaction

		BeforeEach(func() {
			transaction = createTransaction(1)
		})

		When("the worker starts", func() {
			var (
				balanceWorker            workers.BalanceWorker
				ctrl                     *gomock.Controller
				mockTransactionProcessor *mocks.MockTransactionProcessor
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockTransactionProcessor = mocks.NewMockTransactionProcessor(ctrl)
				balanceWorker = workers.NewBalanceWorker(DB)
				balanceWorker.TransactionProcessor = mockTransactionProcessor
			})

			It("the transactionProcessor should be called", func() {
				mockTransactionProcessor.EXPECT().Execute(transaction)
				err := balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("an unprocessed correction does not exists", func() {
		When("the worker starts", func() {
			var (
				balanceWorker           workers.BalanceWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionProcessor
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionProcessor(ctrl)
				balanceWorker = workers.NewBalanceWorker(DB)
				balanceWorker.CorrectionProcessor = mockCorrectionProcessor
			})

			It("the correctionProcessor should not be called", func() {
				mockCorrectionProcessor.EXPECT().Execute(gomock.Any()).Times(0)
				err := balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("an processed correction exists", func() {
		BeforeEach(func() {
			transaction := createTransaction(1)
			correction := entities.NewCorrection([]string{transaction.ID})
			correction.MarkAsDone()
			err := repositories.NewCorrectionRepository(DB).Save(correction)
			Expect(err).ToNot(HaveOccurred())
		})

		When("the worker starts", func() {
			var (
				balanceWorker           workers.BalanceWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionProcessor
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionProcessor(ctrl)
				balanceWorker = workers.NewBalanceWorker(DB)
				balanceWorker.CorrectionProcessor = mockCorrectionProcessor
			})

			It("the correctionProcessor should not be called", func() {
				mockCorrectionProcessor.EXPECT().Execute(gomock.Any()).Times(0)
				err := balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("an unprocessed correction exists", func() {
		var correction *entities.Correction
		BeforeEach(func() {
			transaction := createTransaction(1)
			correction = entities.NewCorrection([]string{transaction.ID})
			err := repositories.NewCorrectionRepository(DB).Save(correction)
			Expect(err).ToNot(HaveOccurred())
		})

		When("the worker starts", func() {
			var (
				balanceWorker           workers.BalanceWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionProcessor
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionProcessor(ctrl)
				balanceWorker = workers.NewBalanceWorker(DB)
				balanceWorker.CorrectionProcessor = mockCorrectionProcessor
			})

			It("the correctionProcessor should be called", func() {
				mockCorrectionProcessor.EXPECT().Execute(correction)
				err := balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
