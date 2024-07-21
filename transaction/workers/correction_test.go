package workers_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/workers"
	"wallet/transaction/workers/internal/mocks"
)

var _ = Describe("correction worker", func() {

	Context("no corrections exist in the system", func() {
		When("the Correction Worker starts", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionInitializer
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionInitializer(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionInitializer = mockCorrectionProcessor
			})

			It("the CorrectionInitializer should be called", func() {
				mockCorrectionProcessor.EXPECT().Execute()
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("corrections exist in the system and are unprocessed", func() {
		BeforeEach(func() {
			transaction := createTransaction(1)
			correction := entities.NewCorrection([]string{transaction.ID})
			err := repositories.NewCorrectionRepository(DB).Save(correction)
			Expect(err).ToNot(HaveOccurred())
		})

		When("the Correction Worker starts", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionInitializer
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionInitializer(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionInitializer = mockCorrectionProcessor
			})

			It("the CorrectionInitializer should NOT be called", func() {
				mockCorrectionProcessor.EXPECT().Execute().Times(0)
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("corrections exist in the system and were processed less than 10 minutes ago", func() {
		BeforeEach(func() {
			transaction := createTransaction(1)
			correction := entities.NewCorrection([]string{transaction.ID})
			correction.MarkAsDone()
			err := repositories.NewCorrectionRepository(DB).Save(correction)
			Expect(err).ToNot(HaveOccurred())
		})

		When("the Correction Worker starts", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionInitializer
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionInitializer(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionInitializer = mockCorrectionProcessor
			})

			It("the CorrectionInitializer should NOT be called", func() {
				mockCorrectionProcessor.EXPECT().Execute().Times(0)
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("Corrections exist in the system and were processed more than 10 minutes ago", func() {
		BeforeEach(func() {
			transaction := createTransaction(1)
			correction := entities.NewCorrection([]string{transaction.ID})
			now := time.Now()
			oneHourAgo := now.Add(-1 * time.Hour)
			correction.DoneAt = &oneHourAgo
			err := repositories.NewCorrectionRepository(DB).Save(correction)
			Expect(err).ToNot(HaveOccurred())
		})

		When("the correction Worker starts", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionInitializer
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionInitializer(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionInitializer = mockCorrectionProcessor
			})

			It("the correctionInitializer should be called", func() {
				mockCorrectionProcessor.EXPECT().Execute()
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
