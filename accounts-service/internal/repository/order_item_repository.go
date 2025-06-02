package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
)

// OrderItemRepository defines the interface for order item repository operations
type OrderItemRepository interface {
	// Create creates a new order item
	Create(ctx context.Context, orderItem *models.OrderItem) error
	
	// CreateMany creates multiple order items in a single transaction
	CreateMany(ctx context.Context, orderItems []*models.OrderItem) error
	
	// GetByID retrieves an order item by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.OrderItem, error)
	
	// ListByOrder retrieves all items for a specific order
	ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error)
	
	// Delete deletes an order item by ID
	Delete(ctx context.Context, id uuid.UUID) error
	
	// DeleteByOrder deletes all items for a specific order
	DeleteByOrder(ctx context.Context, orderID uuid.UUID) error
}
