package workers_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/workers"
)

var _ = Describe("correction worker", func() {
	Context("correction does not exists", func() {
		When("correction workers started", func() {
			var (
				correctionWorker workers.CorrectionWorker
				correctionRepo   *repositories.CorrectionRepository
				transactionRepo  *repositories.TransactionRepository
				tranasction      *entities.Transaction
			)

			BeforeEach(func() {
				correctionWorker = workers.NewCorrectionWorker(DB, uuid.New())
				correctionRepo = repositories.NewCorrectionRepository(DB)

				transactionRepo = repositories.NewTransactionRepository(DB)
				tranasction = createDoneTransaction(10)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should create new correction", func() {
				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(entities.CorrectionId))
			})

			It("correction process should not be started", func() {
				transactionAfterProcess, err := transactionRepo.FindByID(tranasction.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(transactionAfterProcess.Status).ToNot(Equal(entities.Cancelled))
			})

			It("should be ready", func() {
				correction, err := correctionRepo.FindByID(uuid.MustParse(entities.CorrectionId))
				Expect(err).ToNot(HaveOccurred())

				Expect(correction.LockUuid).ToNot(BeNil())
				Expect(correction.Status).To(Equal(entities.Ready))
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
				correctionWorker workers.CorrectionWorker
				correctionRepo   *repositories.CorrectionRepository
				transactionRepo  *repositories.TransactionRepository
				tranasction      *entities.Transaction
			)

			BeforeEach(func() {
				correctionRepo = repositories.NewCorrectionRepository(DB)

				correctionWorker = workers.NewCorrectionWorker(DB, lockUuid)

				transactionRepo = repositories.NewTransactionRepository(DB)
				tranasction = createDoneTransaction(10)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not create new correction", func() {
				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(entities.CorrectionId))
			})

			It("correction process should not be started", func() {
				transactionAfterProcess, err := transactionRepo.FindByID(tranasction.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(transactionAfterProcess.Status).ToNot(Equal(entities.Cancelled))
			})

			It("should not unlock correction", func() {
				correction, err := correctionRepo.FindByID(uuid.MustParse(entities.CorrectionId))
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
				correctionWorker workers.CorrectionWorker
				correctionRepo   *repositories.CorrectionRepository
				transactionRepo  *repositories.TransactionRepository
				tranasction      *entities.Transaction
			)

			BeforeEach(func() {
				correctionWorker = workers.NewCorrectionWorker(DB, lockUuid)
				correctionRepo = repositories.NewCorrectionRepository(DB)

				transactionRepo = repositories.NewTransactionRepository(DB)
				tranasction = createDoneTransaction(10)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not create new correction", func() {
				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(entities.CorrectionId))
			})

			It("correction process should be started", func() {
				transactionAfterProcess, err := transactionRepo.FindByID(tranasction.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(transactionAfterProcess.Status).To(Equal(entities.Cancelled))
			})

			It("should unlock correction", func() {
				correction, err := correctionRepo.FindByID(uuid.MustParse(entities.CorrectionId))
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
				correctionWorker workers.CorrectionWorker
				correctionRepo   *repositories.CorrectionRepository
				transactionRepo  *repositories.TransactionRepository
				tranasction      *entities.Transaction
			)

			BeforeEach(func() {
				correctionWorker = workers.NewCorrectionWorker(DB, uuid.New())
				correctionRepo = repositories.NewCorrectionRepository(DB)

				transactionRepo = repositories.NewTransactionRepository(DB)
				tranasction = createDoneTransaction(10)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not create new correction", func() {
				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(entities.CorrectionId))
			})

			It("correction process should not be started", func() {
				transactionAfterProcess, err := transactionRepo.FindByID(tranasction.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(transactionAfterProcess.Status).ToNot(Equal(entities.Cancelled))
			})

			It("should not unlock correction", func() {
				correction, err := correctionRepo.FindByID(uuid.MustParse(entities.CorrectionId))
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
				correctionWorker workers.CorrectionWorker
				correctionRepo   *repositories.CorrectionRepository
				transactionRepo  *repositories.TransactionRepository
				tranasction      *entities.Transaction
			)

			BeforeEach(func() {
				correctionWorker = workers.NewCorrectionWorker(DB, uuid.New())
				correctionRepo = repositories.NewCorrectionRepository(DB)

				transactionRepo = repositories.NewTransactionRepository(DB)
				tranasction = createDoneTransaction(10)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not create new correction", func() {
				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(entities.CorrectionId))
			})

			It("correction process should be started", func() {
				transactionAfterProcess, err := transactionRepo.FindByID(tranasction.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(transactionAfterProcess.Status).To(Equal(entities.Cancelled))
			})

			It("should unlock correction", func() {
				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())

				correction, err := correctionRepo.FindByID(uuid.MustParse(entities.CorrectionId))
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
				correctionWorker workers.CorrectionWorker
				correctionRepo   *repositories.CorrectionRepository
				transactionRepo  *repositories.TransactionRepository
				tranasction      *entities.Transaction
			)

			BeforeEach(func() {
				correctionWorker = workers.NewCorrectionWorker(DB, uuid.New())
				correctionRepo = repositories.NewCorrectionRepository(DB)

				transactionRepo = repositories.NewTransactionRepository(DB)
				tranasction = createDoneTransaction(10)

				err := correctionWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not create new correction", func() {
				corrections, err := correctionRepo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
				Expect(corrections[0].ID.String()).To(Equal(entities.CorrectionId))
			})

			It("correction process should be started", func() {
				transactionAfterProcess, err := transactionRepo.FindByID(tranasction.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(transactionAfterProcess.Status).To(Equal(entities.Cancelled))
			})

			It("should unlock correction", func() {
				correction, err := correctionRepo.FindByID(uuid.MustParse(entities.CorrectionId))
				Expect(err).ToNot(HaveOccurred())

				Expect(correction.LockUuid).To(BeNil())
				Expect(correction.Status).To(Equal(entities.Ready))
			})
		})
	})
})
