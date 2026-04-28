package services_test

import (
	"context"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	"wallet/transaction/internal/domain/vo"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("check balance initialization and providing", func() {
	var (
		balanceProvider services.BalanceProvider
	)

	BeforeEach(func() {
		balanceProvider = services.NewBalanceProvider(DB)
	})

	Context("balance does not exists", func() {
		Context("system does not receive any transactions yet", func() {
			When("systeem try to get balance", func() {
				var (
					balance *entities.Balance
					err     error
				)

				BeforeEach(func(ctx context.Context) {
					balance, err = balanceProvider.Provide(ctx, 1)
					Expect(err).ToNot(HaveOccurred())
				})

				It("balance should be zero", func() {
					Expect(balance).ToNot(BeNil())
					Expect(balance.Value.Value()).To(Equal(int64(0)))
				})
			})
		})
		Context("system has done transactions", func() {
			var (
				balance *entities.Balance
				err     error
			)

			BeforeEach(func(ctx context.Context) {
				repo := repositories.NewTransactionRepository(DB)
				transaction := entities.NewTransaction(uuid.New().String(), vo.NewAmount(10), entities.Win, entities.Game, 1)
				transaction.MarkAsDone()

				err := repo.Create(ctx, transaction)
				Expect(err).ToNot(HaveOccurred())

				balance, err = balanceProvider.Provide(ctx, 1)
			})

			BeforeEach(func(ctx context.Context) {
				balance, err = balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
			})

			It("balance should be sum of each transactions", func() {
				Expect(balance).ToNot(BeNil())
				Expect(balance.Value.Value()).To(Equal(int64(10)))
			})
		})
	})

	Context("balance already exists", func() {
		var (
			balance *entities.Balance
			err     error
		)

		BeforeEach(func(ctx context.Context) {
			repo := repositories.NewBalanceRepository(DB)
			err := repo.Save(ctx, entities.NewBalance(vo.NewTotalAmount(int64(11)), 1))
			Expect(err).ToNot(HaveOccurred())
		})

		BeforeEach(func(ctx context.Context) {
			balance, err = balanceProvider.Provide(ctx, 1)
			Expect(err).ToNot(HaveOccurred())
		})

		It("system should use existed balance", func() {
			Expect(balance.Value.Value()).To(Equal(int64(11)))
		})
	})
})

var _ = Describe("check balance updating", func() {
	var (
		balanceService  *services.Balance
		balanceProvider services.BalanceProvider
	)

	BeforeEach(func() {
		balanceService = services.NewBalanceService(DB)
		balanceProvider = services.NewBalanceProvider(DB)
	})

	Context("balance are zero", func() {
		When("positive transaction received", func() {
			BeforeEach(func(ctx context.Context) {
				err := balanceService.UpdateBalance(ctx, vo.NewAmount(10), 1)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should increment balance", func(ctx context.Context) {
				balance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(balance.Value.Value()).To(Equal(int64(10)))
			})
		})
		When("negatiove transaction received", func() {
			var err error
			BeforeEach(func(ctx context.Context) {
				err = balanceService.UpdateBalance(ctx, vo.NewAmount(-10), 1)
			})

			It("error should be rised", func() {
				Expect(err).To(HaveOccurred())
			})

			It("should not increment balance", func(ctx context.Context) {
				balance, err := balanceProvider.Provide(ctx, 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(balance.Value.Value()).To(Equal(int64(0)))
			})
		})
	})
})
