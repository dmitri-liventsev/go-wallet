package interfaces

import (
	"context"
	"errors"
	"fmt"
	balancesvc "wallet/gen/wallet"
	txsvc "wallet/gen/wallet"
	"wallet/transaction"
	"wallet/transaction/internal/domain/repositories"
	"wallet/transaction/internal/domain/vo"

	"gorm.io/gorm"
)

type txController struct {
	txRepo      *repositories.TransactionRepository
	balanceRepo *repositories.BalanceRepository
}

func (t txController) GetBalance(ctx context.Context, payload *txsvc.GetBalancePayload) (*txsvc.GetBalanceResult, error) {
	userID := payload.UserID
	balance, err := t.balanceRepo.Get(userID)

	if err != nil {
		return nil, err
	}

	if balance == nil {
		return nil, fmt.Errorf("balance not found for user %d", userID)
	}

	res := &txsvc.GetBalanceResult{
		UserID:  userID,
		Balance: balance.Value.HumanReadable(),
	}

	return res, nil
}

func (t txController) CreateTransaction(ctx context.Context, payload *txsvc.CreateTransactionPayload) error {
	amount, err := vo.NewAmountFromString(payload.Amount)
	userID := payload.UserID

	if err != nil {
		return err
	}

	if amount.LessThenZero() && payload.State == "win" {
		return errors.New("win amount must be greater than zero")
	}
	if amount.GreaterThenZero() && payload.State == "lose" {
		return errors.New("lose amount must be less than or equal to zero")
	}

	command := transaction.AddTransaction{
		SourceType: payload.SourceType,
		Action:     payload.State,
		Amount:     amount,
		UserID:     userID,
		ID:         payload.TransactionID,
	}

	return command.Execute(t.txRepo)
}

func (t txController) Healthcheck(ctx context.Context) (*balancesvc.HealthcheckResult, error) {
	res := balancesvc.HealthcheckResult{
		Status: "ok",
	}

	return &res, nil
}

func NewTxController(db *gorm.DB) txsvc.Service {
	return txController{
		txRepo:      repositories.NewTransactionRepository(db),
		balanceRepo: repositories.NewBalanceRepository(db),
	}
}
