package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/repository"
)

// OrderRepository implements repository.OrderRepository using PostgreSQL
type OrderRepository struct {
	db *Database
}

// NewOrderRepository creates a new PostgreSQL order repository
func NewOrderRepository(db *Database) repository.OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order
func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (
			id, account_id, status, total_amount, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		order.ID,
		order.AccountID,
		order.Status,
		order.TotalAmount,
		order.CreatedAt,
		order.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetByID retrieves an order by ID
func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	query := `
		SELECT id, account_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order models.Order
	err := r.db.GetDB().GetContext(ctx, &order, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// List retrieves all orders with optional pagination
func (r *OrderRepository) List(ctx context.Context, offset, limit int) ([]*models.Order, error) {
	query := `
		SELECT id, account_id, status, total_amount, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var orders []*models.Order
	err := r.db.GetDB().SelectContext(ctx, &orders, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, nil
}

// ListByAccount retrieves all orders for a specific account
func (r *OrderRepository) ListByAccount(ctx context.Context, accountID uuid.UUID, offset, limit int) ([]*models.Order, error) {
	query := `
		SELECT id, account_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var orders []*models.Order
	err := r.db.GetDB().SelectContext(ctx, &orders, query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders by account: %w", err)
	}

	return orders, nil
}

// ListByStatus retrieves all orders with a specific status
func (r *OrderRepository) ListByStatus(ctx context.Context, status models.OrderStatus, offset, limit int) ([]*models.Order, error) {
	query := `
		SELECT id, account_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var orders []*models.Order
	err := r.db.GetDB().SelectContext(ctx, &orders, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders by status: %w", err)
	}

	return orders, nil
}

// Update updates an existing order
func (r *OrderRepository) Update(ctx context.Context, order *models.Order) error {
	query := `
		UPDATE orders
		SET status = $1, total_amount = $2
		WHERE id = $3
	`

	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		order.Status,
		order.TotalAmount,
		order.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// Delete deletes an order by ID
func (r *OrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM orders WHERE id = $1`

	_, err := r.db.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

// Count returns the total number of orders
func (r *OrderRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM orders`

	var count int
	err := r.db.GetDB().GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}

	return count, nil
}

// CountByAccount returns the total number of orders for a specific account
func (r *OrderRepository) CountByAccount(ctx context.Context, accountID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM orders WHERE account_id = $1`

	var count int
	err := r.db.GetDB().GetContext(ctx, &count, query, accountID)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders by account: %w", err)
	}

	return count, nil
}

// CountByStatus returns the total number of orders with a specific status
func (r *OrderRepository) CountByStatus(ctx context.Context, status models.OrderStatus) (int, error) {
	query := `SELECT COUNT(*) FROM orders WHERE status = $1`

	var count int
	err := r.db.GetDB().GetContext(ctx, &count, query, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders by status: %w", err)
	}

	return count, nil
}
