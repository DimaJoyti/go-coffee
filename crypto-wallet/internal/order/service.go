package order

import (
	"go.uber.org/zap"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/kafka"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
)

// Service represents the order service
type Service struct {
	repo      Repository
	cache     redis.Client
	producer  kafka.Producer
	logger    *logger.Logger
	cacheTTL  time.Duration
}

// NewService creates a new order service
func NewService(repo Repository, cache redis.Client, producer kafka.Producer, logger *logger.Logger) *Service {
	return &Service{
		repo:      repo,
		cache:     cache,
		producer:  producer,
		logger:    logger.Named("order-service"),
		cacheTTL:  time.Hour,
	}
}

// GetOrder gets an order by ID
func (s *Service) GetOrder(ctx context.Context, id string) (*Order, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("order:%s", id)
	data, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var order Order
		if err := json.Unmarshal([]byte(data), &order); err == nil {
			// Get order items
			items, err := s.GetOrderItems(ctx, id)
			if err != nil {
				s.logger.Warn("Failed to get order items", zap.String("id", id), zap.Error(err))
			} else {
				order.Items = items
			}
			
			s.logger.Debug("Order cache hit", zap.String("id", id))
			return &order, nil
		}
	}

	// Cache miss, get from database
	order, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order != nil {
		// Get order items
		items, err := s.GetOrderItems(ctx, id)
		if err != nil {
			s.logger.Warn("Failed to get order items", zap.String("id", id), zap.Error(err))
		} else {
			order.Items = items
		}

		// Cache the order
		orderData, err := json.Marshal(order)
		if err == nil {
			if err := s.cache.Set(ctx, cacheKey, orderData, s.cacheTTL); err != nil {
				s.logger.Warn("Failed to cache order", zap.String("id", id), zap.Error(err))
			}
		}
	}

	return order, nil
}

// GetOrderItems gets order items by order ID
func (s *Service) GetOrderItems(ctx context.Context, orderID string) ([]*OrderItem, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("order:%s:items", orderID)
	data, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var items []*OrderItem
		if err := json.Unmarshal([]byte(data), &items); err == nil {
			s.logger.Debug("Order items cache hit", zap.String("orderID", orderID))
			return items, nil
		}
	}

	// Cache miss, get from database
	items, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	// Cache the items
	if items != nil {
		itemsData, err := json.Marshal(items)
		if err == nil {
			if err := s.cache.Set(ctx, cacheKey, itemsData, s.cacheTTL); err != nil {
				s.logger.Warn("Failed to cache order items", zap.String("orderID", orderID), zap.Error(err))
			}
		}
	}

	return items, nil
}

// CreateOrder creates a new order
func (s *Service) CreateOrder(ctx context.Context, order *Order) error {
	// Create in database
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	for _, item := range order.Items {
		item.OrderID = order.ID
		if err := s.repo.CreateOrderItem(ctx, item); err != nil {
			s.logger.Error("Failed to create order item", zap.String("orderID", order.ID), zap.Error(err))
			// Continue with other items
		}
	}

	// Cache the order
	cacheKey := fmt.Sprintf("order:%s", order.ID)
	orderData, err := json.Marshal(order)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, orderData, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache order", zap.String("id", order.ID), zap.Error(err))
		}
	}

	// Cache the order items
	cacheKey = fmt.Sprintf("order:%s:items", order.ID)
	itemsData, err := json.Marshal(order.Items)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, itemsData, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache order items", zap.String("orderID", order.ID), zap.Error(err))
		}
	}

	// Publish event
	itemEvents := make([]OrderItemEvent, 0, len(order.Items))
	for _, item := range order.Items {
		itemEvents = append(itemEvents, OrderItemEvent{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}
	
	event := OrderCreatedEvent{
		ID:        order.ID,
		UserID:    order.UserID,
		Currency:  order.Currency,
		Amount:    order.Amount,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
		Items:     itemEvents,
	}
	
	eventData, err := json.Marshal(event)
	if err == nil {
		if err := s.producer.Produce("order-events", []byte(order.ID), eventData); err != nil {
			s.logger.Warn("Failed to publish order created event", zap.String("id", order.ID), zap.Error(err))
		}
	}

	return nil
}

// UpdateOrder updates an order
func (s *Service) UpdateOrder(ctx context.Context, id string, order *Order) error {
	// Update in database
	if err := s.repo.UpdateOrder(ctx, id, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("order:%s", id)
	orderData, err := json.Marshal(order)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, orderData, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache order", zap.String("id", id), zap.Error(err))
		}
	}

	// Publish event
	event := OrderUpdatedEvent{
		ID:        order.ID,
		UserID:    order.UserID,
		Status:    order.Status,
		UpdatedAt: order.UpdatedAt,
	}
	
	eventData, err := json.Marshal(event)
	if err == nil {
		if err := s.producer.Produce("order-events", []byte(order.ID), eventData); err != nil {
			s.logger.Warn("Failed to publish order updated event", zap.String("id", order.ID), zap.Error(err))
		}
	}

	return nil
}

// DeleteOrder deletes an order
func (s *Service) DeleteOrder(ctx context.Context, id string) error {
	// Delete order items first
	if err := s.repo.DeleteOrderItems(ctx, id); err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	// Delete order
	if err := s.repo.DeleteOrder(ctx, id); err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	// Delete from cache
	cacheKey := fmt.Sprintf("order:%s", id)
	if err := s.cache.Del(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to delete order from cache", zap.String("id", id), zap.Error(err))
	}

	// Delete order items from cache
	cacheKey = fmt.Sprintf("order:%s:items", id)
	if err := s.cache.Del(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to delete order items from cache", "orderID", id, zap.Error(err))
	}

	// Publish event
	event := OrderDeletedEvent{
		ID:        id,
		DeletedAt: time.Now(),
	}
	
	eventData, err := json.Marshal(event)
	if err == nil {
		if err := s.producer.Produce("order-events", []byte(id), eventData); err != nil {
			s.logger.Warn("Failed to publish order deleted event", zap.String("id", id), zap.Error(err))
		}
	}

	return nil
}

// ListOrders lists orders
func (s *Service) ListOrders(ctx context.Context, userID, status string, page, pageSize int) ([]*Order, int, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("orders:user:%s:status:%s:page:%d:pageSize:%d", 
		userID, status, page, pageSize)
	
	data, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var result struct {
			Orders []*Order `json:"orders"`
			Total  int      `json:"total"`
		}
		if err := json.Unmarshal([]byte(data), &result); err == nil {
			// Get order items for each order
			for _, order := range result.Orders {
				items, err := s.GetOrderItems(ctx, order.ID)
				if err != nil {
					s.logger.Warn("Failed to get order items", zap.String("id", order.ID), zap.Error(err))
				} else {
					order.Items = items
				}
			}
			
			s.logger.Debug("Orders cache hit", zap.String("key", cacheKey))
			return result.Orders, result.Total, nil
		}
	}

	// Cache miss, get from database
	orders, total, err := s.repo.ListOrders(ctx, userID, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	// Get order items for each order
	for _, order := range orders {
		items, err := s.GetOrderItems(ctx, order.ID)
		if err != nil {
			s.logger.Warn("Failed to get order items", zap.String("id", order.ID), zap.Error(err))
		} else {
			order.Items = items
		}
	}

	// Cache the result
	result := struct {
		Orders []*Order `json:"orders"`
		Total  int      `json:"total"`
	}{
		Orders: orders,
		Total:  total,
	}
	
	data, err = json.Marshal(result)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache orders", zap.String("key", cacheKey), zap.Error(err))
		}
	}

	return orders, total, nil
}
