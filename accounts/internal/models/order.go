package models

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order represents an order in the system
type Order struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	AccountID   uuid.UUID   `json:"account_id" db:"account_id"`
	Status      OrderStatus `json:"status" db:"status"`
	TotalAmount float64     `json:"total_amount" db:"total_amount"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	
	// Relationships (not stored in the database)
	Account     *Account    `json:"account,omitempty" db:"-"`
	Items       []OrderItem `json:"items,omitempty" db:"-"`
}

// OrderInput represents the input for creating an order
type OrderInput struct {
	AccountID uuid.UUID      `json:"account_id"`
	Items     []OrderItemInput `json:"items"`
}

// NewOrder creates a new order with default values
func NewOrder(input OrderInput, totalAmount float64) *Order {
	now := time.Now()

	return &Order{
		ID:          uuid.New(),
		AccountID:   input.AccountID,
		Status:      OrderStatusPending,
		TotalAmount: totalAmount,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// UpdateStatus updates the status of an order
func (o *Order) UpdateStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}
