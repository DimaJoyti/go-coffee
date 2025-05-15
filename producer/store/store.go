package store

import (
	"errors"
	"sync"
	"time"
)

// Order представляє замовлення кави
type Order struct {
	ID           string    `json:"id"`
	CustomerName string    `json:"customer_name"`
	CoffeeType   string    `json:"coffee_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OrderStore інтерфейс для сховища замовлень
type OrderStore interface {
	Add(order *Order) error
	Get(id string) (*Order, error)
	Update(order *Order) error
	Delete(id string) error
	List() ([]*Order, error)
}

// InMemoryOrderStore реалізує OrderStore в пам'яті
type InMemoryOrderStore struct {
	orders map[string]*Order
	mu     sync.RWMutex
}

// NewInMemoryOrderStore створює нове сховище замовлень в пам'яті
func NewInMemoryOrderStore() *InMemoryOrderStore {
	return &InMemoryOrderStore{
		orders: make(map[string]*Order),
	}
}

// Add додає замовлення до сховища
func (s *InMemoryOrderStore) Add(order *Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orders[order.ID]; exists {
		return errors.New("order already exists")
	}

	s.orders[order.ID] = order
	return nil
}

// Get отримує замовлення зі сховища за ID
func (s *InMemoryOrderStore) Get(id string) (*Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[id]
	if !exists {
		return nil, errors.New("order not found")
	}

	return order, nil
}

// Update оновлює замовлення в сховищі
func (s *InMemoryOrderStore) Update(order *Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orders[order.ID]; !exists {
		return errors.New("order not found")
	}

	s.orders[order.ID] = order
	return nil
}

// Delete видаляє замовлення зі сховища
func (s *InMemoryOrderStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orders[id]; !exists {
		return errors.New("order not found")
	}

	delete(s.orders, id)
	return nil
}

// List повертає список всіх замовлень
func (s *InMemoryOrderStore) List() ([]*Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*Order, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, order)
	}

	return orders, nil
}
