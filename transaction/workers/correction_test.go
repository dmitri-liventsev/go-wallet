package workers_test

import (
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	"wallet/transaction/workers"
	"wallet/transaction/workers/internal/mocks"
)

var _ = Describe("correction worker", func() {
	Context("correction does not exists", func() {
		When("correction workers started", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionProcessor
				correctionRepo          *repositories.CorrectionRepository
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionProcessor(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionProcessor = mockCorrectionProcessor
				correctionRepo = repositories.NewCorrectionRepository(DB)
			})

			It("should create new correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(services.CorrectionId))
			})

			It("correction process should not be started", func() {
				mockCorrectionProcessor.EXPECT().Execute().Times(0)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not unlock correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				correction, err := correctionRepo.FindByID(uuid.MustParse(services.CorrectionId))
				Expect(err).ToNot(HaveOccurred())

				Expect(correction.LockUuid).ToNot(BeNil())
				Expect(correction.Status).ToNot(Equal(entities.Ready))
			})
		})
	})

	Context("correction exists and locked by another process", func() {
		var lockUuid uuid.UUID

		BeforeEach(func() {
			lockUuid = uuid.New()
			createLockedCorrection(uuid.New())
		})

		When("correction workers started", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionProcessor
				correctionRepo          *repositories.CorrectionRepository
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })
				correctionRepo = repositories.NewCorrectionRepository(DB)

				mockCorrectionProcessor = mocks.NewMockCorrectionProcessor(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionProcessor = mockCorrectionProcessor
				correctionWorker.LockUuid = lockUuid
			})

			It("should not create new correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(services.CorrectionId))
			})

			It("correction process should not be started", func() {
				mockCorrectionProcessor.EXPECT().Execute().Times(0)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not unlock correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				correction, err := correctionRepo.FindByID(uuid.MustParse(services.CorrectionId))
				Expect(err).ToNot(HaveOccurred())

				Expect(correction.LockUuid).ToNot(BeNil())
				Expect(correction.Status).ToNot(Equal(entities.Ready))
			})
		})
	})

	Context("correction exists and locked by same process", func() {
		var lockUuid uuid.UUID

		BeforeEach(func() {
			lockUuid = uuid.New()
			createLockedCorrection(lockUuid)
		})

		When("correction workers started", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionProcessor
				correctionRepo          *repositories.CorrectionRepository
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionProcessor(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionProcessor = mockCorrectionProcessor
				correctionRepo = repositories.NewCorrectionRepository(DB)
			})

			It("should not create new correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(services.CorrectionId))
			})

			It("correction process should be started", func() {
				mockCorrectionProcessor.EXPECT().Execute().Times(1)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should unlock correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				correction, err := correctionRepo.FindByID(uuid.MustParse(services.CorrectionId))
				Expect(err).ToNot(HaveOccurred())

				Expect(correction.LockUuid).To(BeNil())
				Expect(correction.Status).To(Equal(entities.Ready))
			})
		})
	})

	Context("correction exists and its not time to do start new correction", func() {
		var lockUuid uuid.UUID

		BeforeEach(func() {
			lockUuid = uuid.New()
			createLockedCorrection(lockUuid)
		})

		When("correction workers started", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionProcessor
				correctionRepo          *repositories.CorrectionRepository
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionProcessor(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionProcessor = mockCorrectionProcessor
				correctionRepo = repositories.NewCorrectionRepository(DB)
			})

			It("should not create new correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(services.CorrectionId))
			})

			It("correction process should not be started", func() {
				mockCorrectionProcessor.EXPECT().Execute().Times(0)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not unlock correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				correction, err := correctionRepo.FindByID(uuid.MustParse(services.CorrectionId))
				Expect(err).ToNot(HaveOccurred())

				Expect(correction.LockUuid).ToNot(BeNil())
				Expect(correction.Status).ToNot(Equal(entities.Ready))
			})
		})
	})

	Context("correction exists and its time to do start new correction", func() {
		BeforeEach(func() {
			createReadyCorrection()
		})

		When("correction workers started", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionProcessor
				correctionRepo          *repositories.CorrectionRepository
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionProcessor(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionProcessor = mockCorrectionProcessor
				correctionRepo = repositories.NewCorrectionRepository(DB)
			})

			It("should not create new correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(services.CorrectionId))
			})

			It("correction process should be started", func() {
				mockCorrectionProcessor.EXPECT().Execute().Times(1)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should unlock correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				correction, err := correctionRepo.FindByID(uuid.MustParse(services.CorrectionId))
				Expect(err).ToNot(HaveOccurred())

				Expect(correction.LockUuid).To(BeNil())
				Expect(correction.Status).To(Equal(entities.Ready))
			})
		})
	})

	Context("correction exists and frozen", func() {
		BeforeEach(func() {
			createFrozenCorrection()
		})

		When("correction workers started", func() {
			var (
				correctionWorker        workers.CorrectionWorker
				ctrl                    *gomock.Controller
				mockCorrectionProcessor *mocks.MockCorrectionProcessor
				correctionRepo          *repositories.CorrectionRepository
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				DeferCleanup(func() { ctrl.Finish() })

				mockCorrectionProcessor = mocks.NewMockCorrectionProcessor(ctrl)
				correctionWorker = workers.NewCorrectionWorker(DB)
				correctionWorker.CorrectionProcessor = mockCorrectionProcessor
				correctionRepo = repositories.NewCorrectionRepository(DB)
			})

			It("should not create new correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(services.CorrectionId))
			})

			It("correction process should be started", func() {
				mockCorrectionProcessor.EXPECT().Execute().Times(1)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should unlock correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				correction, err := correctionRepo.FindByID(uuid.MustParse(services.CorrectionId))
				Expect(err).ToNot(HaveOccurred())

				Expect(correction.LockUuid).To(BeNil())
				Expect(correction.Status).To(Equal(entities.Ready))
			})
		})
	})
})
