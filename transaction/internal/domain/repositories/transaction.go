package repositories

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"time"
	"wallet/transaction/internal/domain/entities"
)

type TransactionRepository struct {
	db *gorm.DB
}

// Custom error for duplicate key
var ErrDuplicateKey = errors.New("duplicate key value violates unique constraint")

// Save saves the transaction entity to the database.
func (repo TransactionRepository) Save(transaction *entities.Transaction) error {
	transaction.UpdatedAt = time.Now()

	return repo.db.Save(transaction).Error
}

// Create creates the transaction entity to the database if it does not exist.
func (repo TransactionRepository) Create(transaction *entities.Transaction) error {
	transaction.UpdatedAt = time.Now()
	err := repo.db.Create(transaction).Error

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrDuplicateKey
		}

		return err
	}

	return nil
}

// FindByID finds a transaction by its ID in the database and returns the correction entity.
func (repo TransactionRepository) FindByID(id string) (*entities.Transaction, error) {
	var tx entities.Transaction
	if err := repo.db.First(&tx, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &tx, nil
}

// FindByIDs finds a transactions by list of IDs in the database and returns the correction entity.
func (repo TransactionRepository) FindByIDs(ids []string) ([]entities.Transaction, error) {
	var transactions []entities.Transaction
	if err := repo.db.Where("id IN ?", ids).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetNextTransaction Returns the most recent 'old' transaction for processing.
func (repo TransactionRepository) GetNextTransaction() (*entities.Transaction, error) {
	var transaction entities.Transaction

	if err := repo.db.Where("status = ?", entities.New).Order("created_at ASC").Limit(1).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &transaction, nil
}

// GetAllTransactions retrieves all records from the `transactions`.
func (repo TransactionRepository) GetAllTransactions() ([]entities.Transaction, error) {
	var transactions []entities.Transaction
	if err := repo.db.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetLastOddTransactions retrieves the most recent odd-numbered 'new' or 'done' transactions up to the specified limit.
func (repo TransactionRepository) GetLastOddTransactions(limit int) ([]entities.Transaction, error) {
	var transactions []entities.Transaction
	err := repo.db.
		Table("transactions").
		Where("status IN (?, ?)", entities.New, entities.Done).
		Order("created_at DESC").
		Limit(limit * 2).
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	var result []entities.Transaction
	for i, tx := range transactions {
		if (i+1)%2 != 0 {
			result = append(result, tx)
		}
	}

	return result, nil
}

// CalculateBalance calculates the total balance from 'done' transactions.
func (repo TransactionRepository) CalculateBalance() (int64, error) {
	var totalAmount *int64

	result := repo.db.Model(&entities.Transaction{}).
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

// LockNewTransactions locks all new and frozen(transactions which have lockedAt time more than 1 min ago) transactions
func (repo TransactionRepository) LockNewTransactions(lockUuid uuid.UUID) error {
	now := time.Now()
	threshold := now.Add(-1 * time.Minute)

	result := repo.db.Model(&entities.Transaction{}).
		Where("(status = ?) OR (locked_at < ? AND locked_at IS NOT NULL)", "new", threshold).
		Updates(map[string]interface{}{
			"status":    "locked",
			"lock_uuid": lockUuid,
			"locked_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetLockedTransactions returns all processing transactions ordered by created at ASC
func (repo TransactionRepository) GetLockedTransactions() ([]entities.Transaction, error) {
	var transactions []entities.Transaction

	result := repo.db.Where("status = ?", "locked").
		Order("created_at ASC").
		Find(&transactions)

	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}

// FindAll returns all transactions
func (repo TransactionRepository) FindAll() ([]entities.Transaction, error) {
	var transactions []entities.Transaction

	result := repo.db.Find(&transactions)
	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}

// NewTransactionRepository returns TransactionRepository instance.
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}
