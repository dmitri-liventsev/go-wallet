package services_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	"wallet/transaction/internal/domain/vo"
)

var _ = Describe("check balance initialization and providing", func() {
	var balanceService *services.Balance

	BeforeEach(func() {
		balanceService = services.NewBalanceService(DB)
	})

	Context("balance does not exists", func() {
		Context("system does not receive any transactions yet", func() {
			When("systeem try to get balance", func() {
				var (
					balance *entities.Balance
					err     error
				)

				BeforeEach(func() {
					balance, err = balanceService.ProvideBalance()
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

			BeforeEach(func() {
				repo := repositories.NewTransactionRepository(DB)
				transaction := entities.NewTransaction(uuid.New().String(), vo.NewAmount(10), entities.Win, entities.Game)
				transaction.MarkAsDone()

				err := repo.Create(transaction)
				Expect(err).ToNot(HaveOccurred())

				balance, err = balanceService.ProvideBalance()
			})

			BeforeEach(func() {
				balance, err = balanceService.ProvideBalance()
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

		BeforeEach(func() {
			repo := repositories.NewBalanceRepository(DB)
			err := repo.Save(entities.NewBalance(vo.NewTotalAmount(int64(11))))
			Expect(err).ToNot(HaveOccurred())
		})

		BeforeEach(func() {
			balance, err = balanceService.ProvideBalance()
			Expect(err).ToNot(HaveOccurred())
		})

		It("system should use existed balance", func() {
			Expect(balance.Value.Value()).To(Equal(int64(11)))
		})
	})
})

var _ = Describe("check balance updating", func() {
	var balanceService *services.Balance

	BeforeEach(func() {
		balanceService = services.NewBalanceService(DB)
	})

	Context("balance are zero", func() {
		When("positive transaction received", func() {
			BeforeEach(func() {
				err := balanceService.UpdateBalance(vo.NewAmount(10))
				Expect(err).ToNot(HaveOccurred())
			})

			It("should increment balance", func() {
				balance, err := balanceService.ProvideBalance()
				Expect(err).ToNot(HaveOccurred())
				Expect(balance.Value.Value()).To(Equal(int64(10)))
			})
		})
		When("negatiove transaction received", func() {
			var err error
			BeforeEach(func() {
				err = balanceService.UpdateBalance(vo.NewAmount(-10))
			})

			It("error should be rised", func() {
				Expect(err).To(HaveOccurred())
			})

			It("should not increment balance", func() {
				balance, err := balanceService.ProvideBalance()
				Expect(err).ToNot(HaveOccurred())
				Expect(balance.Value.Value()).To(Equal(int64(0)))
			})
		})
	})
})
