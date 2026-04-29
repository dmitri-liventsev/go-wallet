package repositories

import (
	"context"
	"errors"
	"time"
	"wallet/transaction/internal/domain/entities"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

// Custom error for duplicate key
var ErrDuplicateKey = errors.New("duplicate key value violates unique constraint")

// Save saves the transaction entity to the database.
func (repo TransactionRepository) Save(ctx context.Context, transaction *entities.Transaction) error {
	transaction.UpdatedAt = time.Now()

	return repo.db.WithContext(ctx).Save(transaction).Error
}

// Create creates the transaction entity to the database if it does not exist.
func (repo TransactionRepository) Create(ctx context.Context, transaction *entities.Transaction) error {
	transaction.UpdatedAt = time.Now()
	err := repo.db.WithContext(ctx).Create(transaction).Error

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrDuplicateKey
		}

		return err
	}

	return nil
}

// FindByID finds a transaction by its ID in the database and returns the correction entity.
func (repo TransactionRepository) FindByID(ctx context.Context, id string) (*entities.Transaction, error) {
	var tx entities.Transaction
	if err := repo.db.WithContext(ctx).First(&tx, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &tx, nil
}

// GetNextTransaction Returns the most recent 'old' transaction for processing.
func (repo TransactionRepository) GetNextTransaction(ctx context.Context) (*entities.Transaction, error) {
	var transaction entities.Transaction

	if err := repo.db.WithContext(ctx).Where("status = ?", entities.New).Order("created_at ASC").Limit(1).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &transaction, nil
}

// GetAllTransactions retrieves all records from the `transactions`.
func (repo TransactionRepository) GetAllTransactions(ctx context.Context) ([]entities.Transaction, error) {
	var transactions []entities.Transaction
	if err := repo.db.WithContext(ctx).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// CalculateBalance calculates the total balance from 'done' transactions.
func (repo TransactionRepository) CalculateBalance(ctx context.Context) (int64, error) {
	var totalAmount *int64

	result := repo.db.WithContext(ctx).Model(&entities.Transaction{}).
		Where("status = ?", entities.Done).
		Select("SUM(amount)").
		Scan(&totalAmount)

	if result.Error != nil {
		return 0, result.Error
	}

	if totalAmount == nil {
		return 0, nil
	}

	return *totalAmount, nil
}

// GetUsersWithNewTransactions returns distinct user IDs that have new transactions
func (repo TransactionRepository) GetUsersWithNewTransactions(ctx context.Context, limit int) ([]int64, error) {
	var userIds []int64

	err := repo.db.WithContext(ctx).
		Table("transactions").
		Distinct("user_id").
		Where("status = ?", entities.New).
		Limit(limit).
		Pluck("user_id", &userIds).Error

	if err != nil {
		return nil, err
	}

	return userIds, nil
}

// GetUserTransactions returns unprocessed transactions for a specific user ordered by creation date ASC
func (repo TransactionRepository) GetUserTransactions(ctx context.Context, userID int64, limit int) ([]entities.Transaction, error) {
	var transactions []entities.Transaction

	err := repo.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, entities.New).
		Order("created_at ASC").
		Limit(limit).
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// NewTransactionRepository returns TransactionRepository instance.
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}
