package models

import (
	"time"

	"github.com/google/uuid"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID         uuid.UUID `json:"id" db:"id"`
	OrderID    uuid.UUID `json:"order_id" db:"order_id"`
	ProductID  uuid.UUID `json:"product_id" db:"product_id"`
	Quantity   int       `json:"quantity" db:"quantity"`
	UnitPrice  float64   `json:"unit_price" db:"unit_price"`
	TotalPrice float64   `json:"total_price" db:"total_price"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	
	// Relationships (not stored in the database)
	Product    *Product  `json:"product,omitempty" db:"-"`
}

// OrderItemInput represents the input for creating an order item
type OrderItemInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// NewOrderItem creates a new order item
func NewOrderItem(orderID uuid.UUID, productID uuid.UUID, quantity int, unitPrice float64) *OrderItem {
	totalPrice := float64(quantity) * unitPrice

	return &OrderItem{
		ID:         uuid.New(),
		OrderID:    orderID,
		ProductID:  productID,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		TotalPrice: totalPrice,
		CreatedAt:  time.Now(),
	}
}
