package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/repository"
	"github.com/google/uuid"
)

// AccountRepository implements repository.AccountRepository using PostgreSQL
type AccountRepository struct {
	db *Database
}

// NewAccountRepository creates a new PostgreSQL account repository
func NewAccountRepository(db *Database) repository.AccountRepository {
	return &AccountRepository{db: db}
}

// Create creates a new account
func (r *AccountRepository) Create(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO accounts (
			id, username, email, password_hash, first_name, last_name, is_active, is_admin, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		account.ID,
		account.Username,
		account.Email,
		account.PasswordHash,
		account.FirstName,
		account.LastName,
		account.IsActive,
		account.IsAdmin,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetByID retrieves an account by ID
func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, is_active, is_admin, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	var account models.Account
	err := r.db.GetDB().GetContext(ctx, &account, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// GetByUsername retrieves an account by username
func (r *AccountRepository) GetByUsername(ctx context.Context, username string) (*models.Account, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, is_active, is_admin, created_at, updated_at
		FROM accounts
		WHERE username = $1
	`

	var account models.Account
	err := r.db.GetDB().GetContext(ctx, &account, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// GetByEmail retrieves an account by email
func (r *AccountRepository) GetByEmail(ctx context.Context, email string) (*models.Account, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, is_active, is_admin, created_at, updated_at
		FROM accounts
		WHERE email = $1
	`

	var account models.Account
	err := r.db.GetDB().GetContext(ctx, &account, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// List retrieves all accounts with optional pagination
func (r *AccountRepository) List(ctx context.Context, offset, limit int) ([]*models.Account, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, is_active, is_admin, created_at, updated_at
		FROM accounts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var accounts []*models.Account
	err := r.db.GetDB().SelectContext(ctx, &accounts, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	return accounts, nil
}

// Update updates an existing account
func (r *AccountRepository) Update(ctx context.Context, account *models.Account) error {
	query := `
		UPDATE accounts
		SET username = $1, email = $2, password_hash = $3, first_name = $4, last_name = $5, is_active = $6, is_admin = $7
		WHERE id = $8
	`

	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		account.Username,
		account.Email,
		account.PasswordHash,
		account.FirstName,
		account.LastName,
		account.IsActive,
		account.IsAdmin,
		account.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// Delete deletes an account by ID
func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM accounts WHERE id = $1`

	_, err := r.db.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

// Count returns the total number of accounts
func (r *AccountRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM accounts`

	var count int
	err := r.db.GetDB().GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count accounts: %w", err)
	}

	return count, nil
}
