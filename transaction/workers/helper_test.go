package workers_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
	"wallet/transaction/internal/domain/entities"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/services"
	"wallet/transaction/internal/domain/vo"
)

func createLockedCorrection(lockId uuid.UUID) *entities.Correction {
	GinkgoHelper()
	correction := entities.NewCorrection(uuid.MustParse(services.CorrectionId))
	correction.Lock(lockId)
	now := time.Now()
	correction.LockedAt = &now
	err := repositories.NewCorrectionRepository(DB).Save(correction)
	Expect(err).ToNot(HaveOccurred())

	return correction
}

func createReadyCorrection() *entities.Correction {
	correction := entities.NewCorrection(uuid.MustParse(services.CorrectionId))
	correction.Status = entities.Ready
	doneAt := time.Now().Add(-11 * time.Minute)
	correction.LockedAt = &doneAt
	correction.DoneAt = &doneAt

	err := repositories.NewCorrectionRepository(DB).Save(correction)
	Expect(err).ToNot(HaveOccurred())

	return correction
}

func createFrozenCorrection() *entities.Correction {
	correction := entities.NewCorrection(uuid.MustParse(services.CorrectionId))
	correction.Status = entities.Locked
	lockedAt := time.Now().Add(-11 * time.Minute)
	correction.LockedAt = &lockedAt
	doneAt := time.Now().Add(-21 * time.Minute)
	correction.DoneAt = &doneAt

	err := repositories.NewCorrectionRepository(DB).Save(correction)
	Expect(err).ToNot(HaveOccurred())

	return correction
}

func createTransaction(amount int) *entities.Transaction {
	GinkgoHelper()
	return createTransactionWithStatus(amount, entities.New)
}

func createLockedTransaction(lockId *uuid.UUID) *entities.Transaction {
	GinkgoHelper()
	transaction := createTransactionWithStatus(10, entities.New)
	transaction.LockUuid = lockId

	return transaction
}

func createTransactionWithStatus(amount int, status string) *entities.Transaction {
	GinkgoHelper()

	action := entities.Win
	if amount < 0 {
		action = entities.Lost
		amount *= -1
	}

	transactionRepo := repositories.NewTransactionRepository(DB)
	transaction := entities.NewTransaction(uuid.New().String(), vo.NewAmount(amount), action, entities.Game)
	transaction.Status = status

	err := transactionRepo.Save(transaction)
	Expect(err).ToNot(HaveOccurred())

	return transaction
}
