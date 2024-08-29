package services_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
)

var _ = Describe("correction record lazy creation", func() {
	Context("correction record are not exist", func() {
		var correctionProvider services.CorrectionProvider

		BeforeEach(func() {
			correctionProvider = services.NewCorrectionProvider(DB)
		})

		When("trying to provide transaction", func() {
			var (
				correction           *entities.Correction
				err                  error
				correctionRepository *repositories.CorrectionRepository
			)
			BeforeEach(func() {
				correctionRepository = repositories.NewCorrectionRepository(DB)
				correction, err = correctionProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				Expect(correction).ToNot(BeNil())
			})

			It("correction should be created", func() {
				corrections, err := correctionRepository.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
			})
		})
	})

	Context("correction record are exist", func() {
		var (
			correctionProvider services.CorrectionProvider
		)

		BeforeEach(func() {
			correctionProvider = services.NewCorrectionProvider(DB)
		})

		When("trying to provide transaction", func() {
			var (
				correctionRepository *repositories.CorrectionRepository
				correction           *entities.Correction
				err                  error
			)

			BeforeEach(func() {
				correctionRepository = repositories.NewCorrectionRepository(DB)

				correction = entities.NewCorrection(uuid.MustParse(entities.CorrectionId))
				err = correctionRepository.Save(correction)
				Expect(err).ToNot(HaveOccurred())

				correction, err = correctionProvider.Provide()
				Expect(err).ToNot(HaveOccurred())
				Expect(correction).ToNot(BeNil())
			})

			It("correction should be created", func() {
				corrections, err := correctionRepository.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(corrections).To(HaveLen(1))
			})
		})
	})

})
