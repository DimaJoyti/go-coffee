package repository

import (
	"sync"

	"github.com/DimaJoyti/go-coffee/pkg/models"
)

// OrderRepository інтерфейс для роботи з репозиторієм замовлень
type OrderRepository interface {
	Add(order *models.Order) error
	Get(id string) (*models.Order, error)
	Update(order *models.Order) error
	Delete(id string) error
	List() ([]*models.Order, error)
	ListByStatus(status models.OrderStatus) ([]*models.Order, error)
}

// InMemoryOrderRepository реалізує OrderRepository в пам'яті
type InMemoryOrderRepository struct {
	orders map[string]*models.Order
	mutex  sync.RWMutex
}

// NewOrderRepository створює новий репозиторій замовлень
func NewOrderRepository() OrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[string]*models.Order),
	}
}

// Add додає замовлення до репозиторію
func (r *InMemoryOrderRepository) Add(order *models.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.orders[order.ID]; exists {
		return errors.New("order already exists").WithCode("DUPLICATE_ORDER")
	}

	r.orders[order.ID] = order
	return nil
}

// Get отримує замовлення з репозиторію за ID
func (r *InMemoryOrderRepository) Get(id string) (*models.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, errors.New("order not found").WithCode("ORDER_NOT_FOUND")
	}

	return order, nil
}

// Update оновлює замовлення в репозиторії
func (r *InMemoryOrderRepository) Update(order *models.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.orders[order.ID]; !exists {
		return errors.New("order not found").WithCode("ORDER_NOT_FOUND")
	}

	r.orders[order.ID] = order
	return nil
}

// Delete видаляє замовлення з репозиторію
func (r *InMemoryOrderRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.orders[id]; !exists {
		return errors.New("order not found").WithCode("ORDER_NOT_FOUND")
	}

	delete(r.orders, id)
	return nil
}

// List повертає список всіх замовлень
func (r *InMemoryOrderRepository) List() ([]*models.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	orders := make([]*models.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}

	return orders, nil
}

// ListByStatus повертає список замовлень за статусом
func (r *InMemoryOrderRepository) ListByStatus(status models.OrderStatus) ([]*models.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var orders []*models.Order
	for _, order := range r.orders {
		if order.Status == status {
			orders = append(orders, order)
		}
	}

	return orders, nil
}
