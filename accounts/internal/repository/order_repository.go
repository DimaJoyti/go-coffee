package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
)

// OrderRepository defines the interface for order repository operations
type OrderRepository interface {
	// Create creates a new order
	Create(ctx context.Context, order *models.Order) error
	
	// GetByID retrieves an order by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	
	// List retrieves all orders with optional pagination
	List(ctx context.Context, offset, limit int) ([]*models.Order, error)
	
	// ListByAccount retrieves all orders for a specific account
	ListByAccount(ctx context.Context, accountID uuid.UUID, offset, limit int) ([]*models.Order, error)
	
	// ListByStatus retrieves all orders with a specific status
	ListByStatus(ctx context.Context, status models.OrderStatus, offset, limit int) ([]*models.Order, error)
	
	// Update updates an existing order
	Update(ctx context.Context, order *models.Order) error
	
	// Delete deletes an order by ID
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Count returns the total number of orders
	Count(ctx context.Context) (int, error)
	
	// CountByAccount returns the total number of orders for a specific account
	CountByAccount(ctx context.Context, accountID uuid.UUID) (int, error)
	
	// CountByStatus returns the total number of orders with a specific status
	CountByStatus(ctx context.Context, status models.OrderStatus) (int, error)
}
