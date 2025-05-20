package order

import (
	"context"
)

// Repository represents an order repository
type Repository interface {
	// GetOrder gets an order by ID
	GetOrder(ctx context.Context, id string) (*Order, error)
	
	// CreateOrder creates a new order
	CreateOrder(ctx context.Context, order *Order) error
	
	// UpdateOrder updates an order
	UpdateOrder(ctx context.Context, id string, order *Order) error
	
	// DeleteOrder deletes an order
	DeleteOrder(ctx context.Context, id string) error
	
	// ListOrders lists orders
	ListOrders(ctx context.Context, userID, status string, page, pageSize int) ([]*Order, int, error)
	
	// GetOrderItems gets order items by order ID
	GetOrderItems(ctx context.Context, orderID string) ([]*OrderItem, error)
	
	// CreateOrderItem creates a new order item
	CreateOrderItem(ctx context.Context, item *OrderItem) error
	
	// DeleteOrderItems deletes order items by order ID
	DeleteOrderItems(ctx context.Context, orderID string) error
}
