package service

import (
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/errors"
	"github.com/DimaJoyti/go-coffee/pkg/kafka"
	"github.com/DimaJoyti/go-coffee/pkg/models"
	"github.com/google/uuid"

	"github.com/DimaJoyti/go-coffee/producer/internal/repository"
)

// OrderService представляє сервіс для роботи з замовленнями
type OrderService struct {
	kafkaProducer kafka.Producer
	orderRepo     repository.OrderRepository
	kafkaTopic    string
}

// NewOrderService створює новий сервіс для роботи з замовленнями
func NewOrderService(kafkaProducer kafka.Producer, orderRepo repository.OrderRepository) *OrderService {
	return &OrderService{
		kafkaProducer: kafkaProducer,
		orderRepo:     orderRepo,
		kafkaTopic:    "coffee_orders", // За замовчуванням, можна передавати через конфігурацію
	}
}

// CreateOrder створює нове замовлення
func (s *OrderService) CreateOrder(customerName, coffeeType string) (*models.Order, error) {
	// Створення нового замовлення
	now := time.Now()
	order := &models.Order{
		ID:           uuid.New().String(),
		CustomerName: customerName,
		CoffeeType:   coffeeType,
		Status:       models.OrderStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Збереження замовлення в репозиторії
	if err := s.orderRepo.Add(order); err != nil {
		return nil, errors.Wrap(err, "failed to add order to repository")
	}

	// Відправка замовлення в Kafka
	orderJSON, err := order.ToJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal order")
	}

	if err := s.kafkaProducer.PushToQueue(s.kafkaTopic, orderJSON); err != nil {
		return nil, errors.Wrap(err, "failed to send order to Kafka")
	}

	return order, nil
}

// GetOrder отримує замовлення за ID
func (s *OrderService) GetOrder(id string) (*models.Order, error) {
	order, err := s.orderRepo.Get(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get order from repository")
	}
	return order, nil
}

// UpdateOrder оновлює замовлення
func (s *OrderService) UpdateOrder(order *models.Order) error {
	// Оновлення часу оновлення
	order.UpdatedAt = time.Now()

	// Оновлення замовлення в репозиторії
	if err := s.orderRepo.Update(order); err != nil {
		return errors.Wrap(err, "failed to update order in repository")
	}

	// Відправка оновленого замовлення в Kafka
	orderJSON, err := order.ToJSON()
	if err != nil {
		return errors.Wrap(err, "failed to marshal order")
	}

	if err := s.kafkaProducer.PushToQueue(s.kafkaTopic, orderJSON); err != nil {
		return errors.Wrap(err, "failed to send updated order to Kafka")
	}

	return nil
}

// CancelOrder скасовує замовлення
func (s *OrderService) CancelOrder(id string) (*models.Order, error) {
	// Отримання замовлення
	order, err := s.orderRepo.Get(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get order from repository")
	}

	// Оновлення статусу
	order.Status = models.OrderStatusCancelled
	order.UpdatedAt = time.Now()

	// Оновлення замовлення в репозиторії
	if err := s.orderRepo.Update(order); err != nil {
		return nil, errors.Wrap(err, "failed to update order in repository")
	}

	// Відправка оновленого замовлення в Kafka
	orderJSON, err := order.ToJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal order")
	}

	if err := s.kafkaProducer.PushToQueue(s.kafkaTopic, orderJSON); err != nil {
		return nil, errors.Wrap(err, "failed to send updated order to Kafka")
	}

	return order, nil
}

// ListOrders отримує список всіх замовлень
func (s *OrderService) ListOrders() ([]*models.Order, error) {
	orders, err := s.orderRepo.List()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list orders from repository")
	}
	return orders, nil
}
