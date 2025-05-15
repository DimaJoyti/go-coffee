package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/repository"
)

// OrderItemRepository implements repository.OrderItemRepository using PostgreSQL
type OrderItemRepository struct {
	db *Database
}

// NewOrderItemRepository creates a new PostgreSQL order item repository
func NewOrderItemRepository(db *Database) repository.OrderItemRepository {
	return &OrderItemRepository{db: db}
}

// Create creates a new order item
func (r *OrderItemRepository) Create(ctx context.Context, orderItem *models.OrderItem) error {
	query := `
		INSERT INTO order_items (
			id, order_id, product_id, quantity, unit_price, total_price, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		orderItem.ID,
		orderItem.OrderID,
		orderItem.ProductID,
		orderItem.Quantity,
		orderItem.UnitPrice,
		orderItem.TotalPrice,
		orderItem.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}

	return nil
}

// CreateMany creates multiple order items in a single transaction
func (r *OrderItemRepository) CreateMany(ctx context.Context, orderItems []*models.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Prepare the statement
	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO order_items (
			id, order_id, product_id, quantity, unit_price, total_price, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute the statement for each order item
	for _, item := range orderItems {
		_, err := stmt.ExecContext(
			ctx,
			item.ID,
			item.OrderID,
			item.ProductID,
			item.Quantity,
			item.UnitPrice,
			item.TotalPrice,
			item.CreatedAt,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves an order item by ID
func (r *OrderItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, unit_price, total_price, created_at
		FROM order_items
		WHERE id = $1
	`

	var orderItem models.OrderItem
	err := r.db.GetDB().GetContext(ctx, &orderItem, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("order item not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get order item: %w", err)
	}

	return &orderItem, nil
}

// ListByOrder retrieves all items for a specific order
func (r *OrderItemRepository) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, unit_price, total_price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	var orderItems []*models.OrderItem
	err := r.db.GetDB().SelectContext(ctx, &orderItems, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to list order items: %w", err)
	}

	return orderItems, nil
}

// Delete deletes an order item by ID
func (r *OrderItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM order_items WHERE id = $1`

	_, err := r.db.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order item: %w", err)
	}

	return nil
}

// DeleteByOrder deletes all items for a specific order
func (r *OrderItemRepository) DeleteByOrder(ctx context.Context, orderID uuid.UUID) error {
	query := `DELETE FROM order_items WHERE order_id = $1`

	_, err := r.db.GetDB().ExecContext(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	return nil
}
