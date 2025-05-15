package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Order represents a coffee order
type Order struct {
	ID           string    `json:"id"`
	CustomerName string    `json:"customer_name"`
	CoffeeType   string    `json:"coffee_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewOrder creates a new order with default values
func NewOrder(customerName, coffeeType string) *Order {
	now := time.Now()
	return &Order{
		ID:           uuid.New().String(),
		CustomerName: customerName,
		CoffeeType:   coffeeType,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// FromJSON creates an Order from JSON
func FromJSON(data []byte) (*Order, error) {
	var order Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// ToJSON converts an Order to JSON
func (o *Order) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// UpdateStatus updates the status of an order
func (o *Order) UpdateStatus(status string) {
	o.Status = status
	o.UpdatedAt = time.Now()
}

// ProcessedOrder represents a processed coffee order
type ProcessedOrder struct {
	OrderID      string    `json:"order_id"`
	CustomerName string    `json:"customer_name"`
	CoffeeType   string    `json:"coffee_type"`
	Status       string    `json:"status"`
	ProcessedAt  time.Time `json:"processed_at"`
	PreparationTime int    `json:"preparation_time"` // in seconds
}

// NewProcessedOrder creates a new processed order from an Order
func NewProcessedOrder(order *Order) *ProcessedOrder {
	// Calculate preparation time based on coffee type
	preparationTime := calculatePreparationTime(order.CoffeeType)
	
	return &ProcessedOrder{
		OrderID:      order.ID,
		CustomerName: order.CustomerName,
		CoffeeType:   order.CoffeeType,
		Status:       "processing",
		ProcessedAt:  time.Now(),
		PreparationTime: preparationTime,
	}
}

// FromJSONProcessed creates a ProcessedOrder from JSON
func FromJSONProcessed(data []byte) (*ProcessedOrder, error) {
	var order ProcessedOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// ToJSON converts a ProcessedOrder to JSON
func (o *ProcessedOrder) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// calculatePreparationTime calculates the preparation time based on coffee type
func calculatePreparationTime(coffeeType string) int {
	switch coffeeType {
	case "Espresso":
		return 30
	case "Latte":
		return 60
	case "Cappuccino":
		return 90
	case "Americano":
		return 45
	default:
		return 60
	}
}
