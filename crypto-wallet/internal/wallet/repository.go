package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/models"
)

// Repository provides access to wallet storage
type Repository interface {
	CreateWallet(ctx context.Context, wallet *models.Wallet) error
	GetWallet(ctx context.Context, id string) (*models.Wallet, error)
	ListWallets(ctx context.Context, userID, chain, walletType string, limit, offset int) ([]models.Wallet, int, error)
	UpdateWallet(ctx context.Context, wallet *models.Wallet) error
	DeleteWallet(ctx context.Context, id string) error
	SaveKeystore(ctx context.Context, walletID, keystore string) error
	GetKeystore(ctx context.Context, walletID string) (string, error)
	DeleteKeystore(ctx context.Context, walletID string) error
}

// PostgresRepository is a PostgreSQL implementation of Repository
type PostgresRepository struct {
	db            *sqlx.DB
	logger        *logger.Logger
	keystorePath  string
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB, logger *logger.Logger, keystorePath string) *PostgresRepository {
	// Create keystore directory if it doesn't exist
	if err := os.MkdirAll(keystorePath, 0700); err != nil {
		logger.Error(fmt.Sprintf("Failed to create keystore directory: %v", err))
		panic(fmt.Sprintf("Failed to create keystore directory: %v", err))
	}

	return &PostgresRepository{
		db:           db,
		logger:       logger.Named("wallet-repository"),
		keystorePath: keystorePath,
	}
}

// CreateWallet creates a new wallet in the database
func (r *PostgresRepository) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	query := `
		INSERT INTO wallets (id, user_id, name, address, chain, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		wallet.ID,
		wallet.UserID,
		wallet.Name,
		wallet.Address,
		wallet.Chain,
		wallet.Type,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to create wallet: %v", err))
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	return nil
}

// GetWallet retrieves a wallet by ID
func (r *PostgresRepository) GetWallet(ctx context.Context, id string) (*models.Wallet, error) {
	query := `
		SELECT id, user_id, name, address, chain, type, created_at, updated_at
		FROM wallets
		WHERE id = $1
	`

	var wallet models.Wallet
	err := r.db.GetContext(ctx, &wallet, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.Error(fmt.Sprintf("Failed to get wallet: %v", err))
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

// ListWallets lists all wallets for a user
func (r *PostgresRepository) ListWallets(ctx context.Context, userID, chain, walletType string, limit, offset int) ([]models.Wallet, int, error) {
	// Build query based on filters
	query := `
		SELECT id, user_id, name, address, chain, type, created_at, updated_at
		FROM wallets
		WHERE user_id = $1
	`
	countQuery := `
		SELECT COUNT(*)
		FROM wallets
		WHERE user_id = $1
	`

	args := []interface{}{userID}
	argIndex := 2

	if chain != "" {
		query += fmt.Sprintf(" AND chain = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND chain = $%d", argIndex)
		args = append(args, chain)
		argIndex++
	}

	if walletType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, walletType)
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

	// Get wallets
	var wallets []models.Wallet
	err := r.db.SelectContext(ctx, &wallets, query, args...)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to list wallets: %v", err))
		return nil, 0, fmt.Errorf("failed to list wallets: %w", err)
	}

	// Get total count
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, args[:argIndex-1]...)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to get total wallet count: %v", err))
		return nil, 0, fmt.Errorf("failed to get total wallet count: %w", err)
	}

	return wallets, total, nil
}

// UpdateWallet updates a wallet in the database
func (r *PostgresRepository) UpdateWallet(ctx context.Context, wallet *models.Wallet) error {
	query := `
		UPDATE wallets
		SET name = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		wallet.Name,
		wallet.UpdatedAt,
		wallet.ID,
	)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to update wallet: %v", err))
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}

// DeleteWallet deletes a wallet from the database
func (r *PostgresRepository) DeleteWallet(ctx context.Context, id string) error {
	query := `
		DELETE FROM wallets
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to delete wallet: %v", err))
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	return nil
}

// SaveKeystore saves a keystore file
func (r *PostgresRepository) SaveKeystore(ctx context.Context, walletID, keystore string) error {
	// Create keystore file path
	keystorePath := filepath.Join(r.keystorePath, walletID+".json")

	// Write keystore to file
	err := os.WriteFile(keystorePath, []byte(keystore), 0600)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to save keystore: %v", err))
		return fmt.Errorf("failed to save keystore: %w", err)
	}

	return nil
}

// GetKeystore retrieves a keystore file
func (r *PostgresRepository) GetKeystore(ctx context.Context, walletID string) (string, error) {
	// Create keystore file path
	keystorePath := filepath.Join(r.keystorePath, walletID+".json")

	// Read keystore from file
	data, err := os.ReadFile(keystorePath)
	if err != nil {
		if os.IsNotExist(err) {
			r.logger.Error(fmt.Sprintf("Keystore not found: %s", walletID))
			return "", fmt.Errorf("keystore not found: %s", walletID)
		}
		r.logger.Error(fmt.Sprintf("Failed to read keystore: %v", err))
		return "", fmt.Errorf("failed to read keystore: %w", err)
	}

	return string(data), nil
}

// DeleteKeystore deletes a keystore file
func (r *PostgresRepository) DeleteKeystore(ctx context.Context, walletID string) error {
	// Create keystore file path
	keystorePath := filepath.Join(r.keystorePath, walletID+".json")

	// Delete keystore file
	err := os.Remove(keystorePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, consider it deleted
			return nil
		}
		r.logger.Error(fmt.Sprintf("Failed to delete keystore: %v", err))
		return fmt.Errorf("failed to delete keystore: %w", err)
	}

	return nil
}
