package interfaces

import (
	"context"
	"errors"
	"gorm.io/gorm"
	balancesvc "wallet/gen/transaction"
	txsvc "wallet/gen/transaction"
	"wallet/transaction"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"
)

type txController struct {
	repo *repositories.TransactionRepository
}

func (t txController) Create(ctx context.Context, payload *balancesvc.CreatePayload) error {
	amount, err := vo.NewAmountFromString(payload.Amount)
	if err != nil {
		return err
	}

	if amount.LessThenZero() && payload.State == "win" {
		return errors.New("win amount must be greater than zero")
	}
	if amount.GreaterThenZero() && payload.State == "lost" {
		return errors.New("win amount must be greater than zero")
	}

	command := transaction.AddTransaction{
		SourceType: payload.SourceType,
		Action:     payload.State,
		Amount:     amount,
		ID:         payload.TransactionID,
	}

	return command.Execute(t.repo)
}

func (t txController) Healthcheck(ctx context.Context) (*balancesvc.HealthcheckResult, error) {
	res := balancesvc.HealthcheckResult{
		Status: "ok",
	}

	return &res, nil
}

func NewTxController(db *gorm.DB) txsvc.Service {
	return txController{
		repo: repositories.NewTransactionRepository(db),
	}
}
