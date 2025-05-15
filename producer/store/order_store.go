package store

import (
	"errors"
	"sync"
	"time"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	// OrderStatusPending represents a pending order
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusProcessing represents a processing order
	OrderStatusProcessing OrderStatus = "processing"
	// OrderStatusCompleted represents a completed order
	OrderStatusCompleted OrderStatus = "completed"
	// OrderStatusCancelled represents a cancelled order
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order represents a coffee order
type Order struct {
	ID           string      `json:"id"`
	CustomerName string      `json:"customer_name"`
	CoffeeType   string      `json:"coffee_type"`
	Status       OrderStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// OrderStore is an interface for storing and retrieving orders
type OrderStore interface {
	Add(order *Order) error
	Get(id string) (*Order, error)
	Update(order *Order) error
	Delete(id string) error
	List() ([]*Order, error)
	ListByStatus(status OrderStatus) ([]*Order, error)
	ListByCustomer(customerName string) ([]*Order, error)
}

// InMemoryOrderStore is an in-memory implementation of OrderStore
type InMemoryOrderStore struct {
	orders map[string]*Order
	mutex  sync.RWMutex
}

// NewInMemoryOrderStore creates a new InMemoryOrderStore
func NewInMemoryOrderStore() *InMemoryOrderStore {
	return &InMemoryOrderStore{
		orders: make(map[string]*Order),
	}
}

// Add adds an order to the store
func (s *InMemoryOrderStore) Add(order *Order) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.orders[order.ID]; exists {
		return errors.New("order already exists")
	}

	s.orders[order.ID] = order
	return nil
}

// Get retrieves an order from the store
func (s *InMemoryOrderStore) Get(id string) (*Order, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	order, exists := s.orders[id]
	if !exists {
		return nil, errors.New("order not found")
	}

	return order, nil
}

// Update updates an order in the store
func (s *InMemoryOrderStore) Update(order *Order) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.orders[order.ID]; !exists {
		return errors.New("order not found")
	}

	order.UpdatedAt = time.Now()
	s.orders[order.ID] = order
	return nil
}

// Delete deletes an order from the store
func (s *InMemoryOrderStore) Delete(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.orders[id]; !exists {
		return errors.New("order not found")
	}

	delete(s.orders, id)
	return nil
}

// List returns all orders in the store
func (s *InMemoryOrderStore) List() ([]*Order, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	orders := make([]*Order, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, order)
	}

	return orders, nil
}

// ListByStatus returns all orders with the specified status
func (s *InMemoryOrderStore) ListByStatus(status OrderStatus) ([]*Order, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	orders := make([]*Order, 0)
	for _, order := range s.orders {
		if order.Status == status {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// ListByCustomer returns all orders for the specified customer
func (s *InMemoryOrderStore) ListByCustomer(customerName string) ([]*Order, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	orders := make([]*Order, 0)
	for _, order := range s.orders {
		if order.CustomerName == customerName {
			orders = append(orders, order)
		}
	}

	return orders, nil
}
