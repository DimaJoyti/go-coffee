package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OrderStatus представляє статус замовлення
type OrderStatus string

const (
	// OrderStatusPending представляє статус "очікує обробки"
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusProcessing представляє статус "в обробці"
	OrderStatusProcessing OrderStatus = "processing"
	// OrderStatusCompleted представляє статус "виконано"
	OrderStatusCompleted OrderStatus = "completed"
	// OrderStatusCancelled представляє статус "скасовано"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order представляє замовлення кави
type Order struct {
	ID           string      `json:"id"`
	CustomerName string      `json:"customer_name"`
	CoffeeType   string      `json:"coffee_type"`
	Status       OrderStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// NewOrder створює нове замовлення з типовими значеннями
func NewOrder(customerName, coffeeType string) *Order {
	now := time.Now()
	return &Order{
		ID:           uuid.New().String(),
		CustomerName: customerName,
		CoffeeType:   coffeeType,
		Status:       OrderStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// FromJSON створює Order з JSON
func FromJSON(data []byte) (*Order, error) {
	var order Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// ToJSON конвертує Order в JSON
func (o *Order) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// UpdateStatus оновлює статус замовлення
func (o *Order) UpdateStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}

// ProcessedOrder представляє оброблене замовлення кави
type ProcessedOrder struct {
	OrderID         string      `json:"order_id"`
	CustomerName    string      `json:"customer_name"`
	CoffeeType      string      `json:"coffee_type"`
	Status          OrderStatus `json:"status"`
	ProcessedAt     time.Time   `json:"processed_at"`
	PreparationTime int         `json:"preparation_time"` // в секундах
}

// NewProcessedOrder створює нове оброблене замовлення з Order
func NewProcessedOrder(order *Order) *ProcessedOrder {
	// Розрахунок часу приготування на основі типу кави
	preparationTime := calculatePreparationTime(order.CoffeeType)
	
	return &ProcessedOrder{
		OrderID:         order.ID,
		CustomerName:    order.CustomerName,
		CoffeeType:      order.CoffeeType,
		Status:          OrderStatusProcessing,
		ProcessedAt:     time.Now(),
		PreparationTime: preparationTime,
	}
}

// FromJSONProcessed створює ProcessedOrder з JSON
func FromJSONProcessed(data []byte) (*ProcessedOrder, error) {
	var order ProcessedOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// ToJSON конвертує ProcessedOrder в JSON
func (o *ProcessedOrder) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// calculatePreparationTime розраховує час приготування на основі типу кави
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
