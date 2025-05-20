package order

import (
	"time"
)

// Order represents an order
type Order struct {
	ID        string      `json:"id" db:"id"`
	UserID    string      `json:"user_id" db:"user_id"`
	Currency  string      `json:"currency" db:"currency"`
	Amount    float64     `json:"amount" db:"amount"`
	Status    string      `json:"status" db:"status"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
	Items     []*OrderItem `json:"items"`
}

// OrderItem represents an order item
type OrderItem struct {
	ID          string  `json:"id" db:"id"`
	OrderID     string  `json:"order_id" db:"order_id"`
	ProductID   string  `json:"product_id" db:"product_id"`
	ProductName string  `json:"product_name" db:"product_name"`
	Price       float64 `json:"price" db:"price"`
	Quantity    int     `json:"quantity" db:"quantity"`
}

// OrderCreatedEvent represents an order created event
type OrderCreatedEvent struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Currency  string    `json:"currency"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Items     []OrderItemEvent `json:"items"`
}

// OrderItemEvent represents an order item event
type OrderItemEvent struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

// OrderUpdatedEvent represents an order updated event
type OrderUpdatedEvent struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderDeletedEvent represents an order deleted event
type OrderDeletedEvent struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
}
