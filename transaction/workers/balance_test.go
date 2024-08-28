package workers_test

import (
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
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

	Context("transaction was locked by another process", func() {
		var lockUuid uuid.UUID

		BeforeEach(func() {
			lockUuid = uuid.New()
			createLockedTransaction(&lockUuid)
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

			It("the transactionProcessor should not be called", func() {
				mockTransactionProcessor.EXPECT().Execute(gomock.Any()).Times(0)
				err := balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("transaction was locked by same process before", func() {
		var (
			lockUuid    uuid.UUID
			transaction *entities.Transaction
		)

		BeforeEach(func() {
			lockUuid = uuid.New()
			transaction = createLockedTransaction(&lockUuid)
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
				balanceWorker.LockUuid = lockUuid
			})

			It("the transactionProcessor should be called", func() {
				mockTransactionProcessor.EXPECT().Execute(transaction)
				err := balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("three transaction are locked, transaction in the middle are locked by another process", func() {
		var (
			lockUuid    uuid.UUID
			transaction *entities.Transaction
		)

		BeforeEach(func() {
			lockUuid = uuid.New()
			transaction = createLockedTransaction(&lockUuid)
			randomUuid := uuid.New()
			_ = createLockedTransaction(&randomUuid)
			_ = createLockedTransaction(&lockUuid)
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
				balanceWorker.LockUuid = lockUuid
			})

			It("the transactionProcessor should be called only for one transaction", func() {
				mockTransactionProcessor.EXPECT().Execute(transaction).Times(1)
				err := balanceWorker.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
