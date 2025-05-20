package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
	"github.com/yourusername/web3-wallet-backend/pkg/models"
)

// Repository provides access to transaction storage
type Repository interface {
	CreateTransaction(ctx context.Context, transaction *models.Transaction) error
	GetTransaction(ctx context.Context, id string) (*models.Transaction, error)
	GetTransactionByHash(ctx context.Context, hash, chain string) (*models.Transaction, error)
	ListTransactions(ctx context.Context, userID, walletID, status, chain string, limit, offset int) ([]models.Transaction, int, error)
	UpdateTransaction(ctx context.Context, transaction *models.Transaction) error
	DeleteTransaction(ctx context.Context, id string) error
}

// PostgresRepository is a PostgreSQL implementation of Repository
type PostgresRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB, logger *logger.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger.Named("transaction-repository"),
	}
}

// CreateTransaction creates a new transaction in the database
func (r *PostgresRepository) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, user_id, wallet_id, hash, from_address, to_address, value, gas, gas_price, nonce, data, chain, status,
			block_number, block_hash, confirmations, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		transaction.ID,
		transaction.UserID,
		transaction.WalletID,
		transaction.Hash,
		transaction.From,
		transaction.To,
		transaction.Value,
		transaction.Gas,
		transaction.GasPrice,
		transaction.Nonce,
		transaction.Data,
		transaction.Chain,
		transaction.Status,
		transaction.BlockNumber,
		transaction.BlockHash,
		transaction.Confirmations,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to create transaction: %v", err))
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetTransaction retrieves a transaction by ID
func (r *PostgresRepository) GetTransaction(ctx context.Context, id string) (*models.Transaction, error) {
	query := `
		SELECT id, user_id, wallet_id, hash, from_address, to_address, value, gas, gas_price, nonce, data, chain, status,
			block_number, block_hash, confirmations, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	var transaction models.Transaction
	err := r.db.GetContext(ctx, &transaction, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.Error(fmt.Sprintf("Failed to get transaction: %v", err))
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

// GetTransactionByHash retrieves a transaction by hash
func (r *PostgresRepository) GetTransactionByHash(ctx context.Context, hash, chain string) (*models.Transaction, error) {
	query := `
		SELECT id, user_id, wallet_id, hash, from_address, to_address, value, gas, gas_price, nonce, data, chain, status,
			block_number, block_hash, confirmations, created_at, updated_at
		FROM transactions
		WHERE hash = $1 AND chain = $2
	`

	var transaction models.Transaction
	err := r.db.GetContext(ctx, &transaction, query, hash, chain)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.Error(fmt.Sprintf("Failed to get transaction by hash: %v", err))
		return nil, fmt.Errorf("failed to get transaction by hash: %w", err)
	}

	return &transaction, nil
}

// ListTransactions lists all transactions for a user
func (r *PostgresRepository) ListTransactions(ctx context.Context, userID, walletID, status, chain string, limit, offset int) ([]models.Transaction, int, error) {
	// Build query based on filters
	query := `
		SELECT id, user_id, wallet_id, hash, from_address, to_address, value, gas, gas_price, nonce, data, chain, status,
			block_number, block_hash, confirmations, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
	`
	countQuery := `
		SELECT COUNT(*)
		FROM transactions
		WHERE user_id = $1
	`

	args := []interface{}{userID}
	argIndex := 2

	if walletID != "" {
		query += fmt.Sprintf(" AND wallet_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND wallet_id = $%d", argIndex)
		args = append(args, walletID)
		argIndex++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if chain != "" {
		query += fmt.Sprintf(" AND chain = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND chain = $%d", argIndex)
		args = append(args, chain)
		argIndex++
	}

	// Add order by and pagination
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	// Get transactions
	var transactions []models.Transaction
	err := r.db.SelectContext(ctx, &transactions, query, args...)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to list transactions: %v", err))
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Get total count
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, args[:argIndex-1]...)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to get total transaction count: %v", err))
		return nil, 0, fmt.Errorf("failed to get total transaction count: %w", err)
	}

	return transactions, total, nil
}

// UpdateTransaction updates a transaction in the database
func (r *PostgresRepository) UpdateTransaction(ctx context.Context, transaction *models.Transaction) error {
	query := `
		UPDATE transactions
		SET status = $1, block_number = $2, block_hash = $3, confirmations = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		transaction.Status,
		transaction.BlockNumber,
		transaction.BlockHash,
		transaction.Confirmations,
		transaction.UpdatedAt,
		transaction.ID,
	)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to update transaction: %v", err))
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// DeleteTransaction deletes a transaction from the database
func (r *PostgresRepository) DeleteTransaction(ctx context.Context, id string) error {
	query := `
		DELETE FROM transactions
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to delete transaction: %v", err))
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}
