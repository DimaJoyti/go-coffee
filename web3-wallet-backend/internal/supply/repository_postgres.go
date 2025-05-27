package supply

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// PostgresRepository implements the Repository interface using PostgreSQL
type PostgresRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB, logger *logger.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger.Named("supply-postgres-repo"),
	}
}

// GetSupply gets a supply by ID
func (r *PostgresRepository) GetSupply(ctx context.Context, id string) (*Supply, error) {
	query := `
		SELECT id, user_id, currency, amount, status, created_at, updated_at
		FROM supplies
		WHERE id = $1
	`

	var supply Supply
	err := r.db.GetContext(ctx, &supply, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get supply: %w", err)
	}

	return &supply, nil
}

// CreateSupply creates a new supply
func (r *PostgresRepository) CreateSupply(ctx context.Context, supply *Supply) error {
	query := `
		INSERT INTO supplies (id, user_id, currency, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		supply.ID,
		supply.UserID,
		supply.Currency,
		supply.Amount,
		supply.Status,
		supply.CreatedAt,
		supply.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create supply: %w", err)
	}

	return nil
}

// UpdateSupply updates a supply
func (r *PostgresRepository) UpdateSupply(ctx context.Context, id string, supply *Supply) error {
	query := `
		UPDATE supplies
		SET user_id = $1, currency = $2, amount = $3, status = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		supply.UserID,
		supply.Currency,
		supply.Amount,
		supply.Status,
		supply.UpdatedAt,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update supply: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supply not found")
	}

	return nil
}

// DeleteSupply deletes a supply
func (r *PostgresRepository) DeleteSupply(ctx context.Context, id string) error {
	query := `
		DELETE FROM supplies
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete supply: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supply not found")
	}

	return nil
}

// ListSupplies lists supplies
func (r *PostgresRepository) ListSupplies(ctx context.Context, userID, currency, status string, page, pageSize int) ([]*Supply, int, error) {
	// Build query
	baseQuery := `
		SELECT id, user_id, currency, amount, status, created_at, updated_at
		FROM supplies
		WHERE 1=1
	`
	countQuery := `
		SELECT COUNT(*)
		FROM supplies
		WHERE 1=1
	`

	// Add filters
	var args []interface{}
	argIndex := 1

	if userID != "" {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if currency != "" {
		baseQuery += fmt.Sprintf(" AND currency = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND currency = $%d", argIndex)
		args = append(args, currency)
		argIndex++
	}

	if status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// Add pagination
	baseQuery += " ORDER BY created_at DESC"
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, pageSize, offset)
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args[:argIndex-1]...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get supplies
	var supplies []*Supply
	err = r.db.SelectContext(ctx, &supplies, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list supplies: %w", err)
	}

	return supplies, total, nil
}
