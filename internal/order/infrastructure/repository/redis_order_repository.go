package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"time"

	"github.com/go-redis/redis/v8"

	"github.com/DimaJoyti/go-coffee/internal/order/application"
	"github.com/DimaJoyti/go-coffee/internal/order/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RedisOrderRepository implements OrderRepository using Redis
type RedisOrderRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisOrderRepository creates a new Redis-based order repository
func NewRedisOrderRepository(client *redis.Client, logger *logger.Logger) *RedisOrderRepository {
	return &RedisOrderRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new order in Redis
func (r *RedisOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	// Serialize order to JSON
	orderData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Redis keys
	orderKey := r.getOrderKey(order.ID)
	customerOrdersKey := r.getCustomerOrdersKey(order.CustomerID)
	ordersByStatusKey := r.getOrdersByStatusKey(order.Status)
	allOrdersKey := r.getAllOrdersKey()

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Store order data
	pipe.Set(ctx, orderKey, orderData, 0)

	// Add to customer orders (sorted set by creation time)
	pipe.ZAdd(ctx, customerOrdersKey, &redis.Z{
		Score:  float64(order.CreatedAt.Unix()),
		Member: order.ID,
	})

	// Add to orders by status (sorted set by creation time)
	pipe.ZAdd(ctx, ordersByStatusKey, &redis.Z{
		Score:  float64(order.CreatedAt.Unix()),
		Member: order.ID,
	})

	// Add to all orders (sorted set by creation time)
	pipe.ZAdd(ctx, allOrdersKey, &redis.Z{
		Score:  float64(order.CreatedAt.Unix()),
		Member: order.ID,
	})

	// Set expiration for customer orders (30 days)
	pipe.Expire(ctx, customerOrdersKey, 30*24*time.Hour)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create order in Redis: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"order_id":    order.ID,
		"customer_id": order.CustomerID,
		"status":      order.Status.String(),
	}).Info("Order created in Redis")

	return nil
}

// GetByID retrieves an order by ID
func (r *RedisOrderRepository) GetByID(ctx context.Context, orderID string) (*domain.Order, error) {
	orderKey := r.getOrderKey(orderID)

	// Get order data
	orderData, err := r.client.Get(ctx, orderKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("order not found: %s", orderID)
		}
		return nil, fmt.Errorf("failed to get order from Redis: %w", err)
	}

	// Deserialize order
	var order domain.Order
	if err := json.Unmarshal([]byte(orderData), &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

// GetByCustomerID retrieves orders for a specific customer
func (r *RedisOrderRepository) GetByCustomerID(ctx context.Context, customerID string, limit, offset int) ([]*domain.Order, error) {
	customerOrdersKey := r.getCustomerOrdersKey(customerID)

	// Get order IDs (sorted by creation time, newest first)
	orderIDs, err := r.client.ZRevRange(ctx, customerOrdersKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get customer orders from Redis: %w", err)
	}

	// Get orders
	orders := make([]*domain.Order, 0, len(orderIDs))
	for _, orderID := range orderIDs {
		order, err := r.GetByID(ctx, orderID)
		if err != nil {
			r.logger.WithError(err).WithField("order_id", orderID).Warn("Failed to get order")
			continue
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// Update updates an existing order
func (r *RedisOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	// Get existing order to check status change
	existingOrder, err := r.GetByID(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing order: %w", err)
	}

	// Serialize order to JSON
	orderData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Redis keys
	orderKey := r.getOrderKey(order.ID)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Update order data
	pipe.Set(ctx, orderKey, orderData, 0)

	// If status changed, update status indexes
	if existingOrder.Status != order.Status {
		// Remove from old status set
		oldStatusKey := r.getOrdersByStatusKey(existingOrder.Status)
		pipe.ZRem(ctx, oldStatusKey, order.ID)

		// Add to new status set
		newStatusKey := r.getOrdersByStatusKey(order.Status)
		pipe.ZAdd(ctx, newStatusKey, &redis.Z{
			Score:  float64(order.UpdatedAt.Unix()),
			Member: order.ID,
		})
	}

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update order in Redis: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"order_id":    order.ID,
		"customer_id": order.CustomerID,
		"status":      order.Status.String(),
	}).Info("Order updated in Redis")

	return nil
}

// Delete deletes an order
func (r *RedisOrderRepository) Delete(ctx context.Context, orderID string) error {
	// Get order first to get metadata for cleanup
	order, err := r.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order for deletion: %w", err)
	}

	// Redis keys
	orderKey := r.getOrderKey(orderID)
	customerOrdersKey := r.getCustomerOrdersKey(order.CustomerID)
	ordersByStatusKey := r.getOrdersByStatusKey(order.Status)
	allOrdersKey := r.getAllOrdersKey()

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Delete order data
	pipe.Del(ctx, orderKey)

	// Remove from indexes
	pipe.ZRem(ctx, customerOrdersKey, orderID)
	pipe.ZRem(ctx, ordersByStatusKey, orderID)
	pipe.ZRem(ctx, allOrdersKey, orderID)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete order from Redis: %w", err)
	}

	r.logger.WithField("order_id", orderID).Info("Order deleted from Redis")

	return nil
}

// List retrieves orders based on filters
func (r *RedisOrderRepository) List(ctx context.Context, filters application.OrderFilters) ([]*domain.Order, error) {
	var orderIDs []string
	var err error

	// Determine which index to use based on filters
	if filters.CustomerID != "" {
		// Use customer-specific index
		customerOrdersKey := r.getCustomerOrdersKey(filters.CustomerID)
		orderIDs, err = r.client.ZRevRange(ctx, customerOrdersKey, int64(filters.Offset), int64(filters.Offset+filters.Limit-1)).Result()
	} else if filters.Status != domain.OrderStatusUnknown {
		// Use status-specific index
		ordersByStatusKey := r.getOrdersByStatusKey(filters.Status)
		orderIDs, err = r.client.ZRevRange(ctx, ordersByStatusKey, int64(filters.Offset), int64(filters.Offset+filters.Limit-1)).Result()
	} else {
		// Use all orders index
		allOrdersKey := r.getAllOrdersKey()
		orderIDs, err = r.client.ZRevRange(ctx, allOrdersKey, int64(filters.Offset), int64(filters.Offset+filters.Limit-1)).Result()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get order IDs from Redis: %w", err)
	}

	// Get orders and apply additional filters
	orders := make([]*domain.Order, 0, len(orderIDs))
	for _, orderID := range orderIDs {
		order, err := r.GetByID(ctx, orderID)
		if err != nil {
			r.logger.WithError(err).WithField("order_id", orderID).Warn("Failed to get order")
			continue
		}

		// Apply additional filters
		if r.matchesFilters(order, filters) {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// Helper methods

// matchesFilters checks if an order matches the given filters
func (r *RedisOrderRepository) matchesFilters(order *domain.Order, filters application.OrderFilters) bool {
	// Date range filter
	if filters.StartDate != nil && order.CreatedAt.Before(*filters.StartDate) {
		return false
	}
	if filters.EndDate != nil && order.CreatedAt.After(*filters.EndDate) {
		return false
	}

	// Amount range filter
	if filters.MinAmount != nil && order.TotalAmount < *filters.MinAmount {
		return false
	}
	if filters.MaxAmount != nil && order.TotalAmount > *filters.MaxAmount {
		return false
	}

	// Delivery filter
	if filters.IsDelivery != nil && order.IsDelivery != *filters.IsDelivery {
		return false
	}

	return true
}

// Redis key generators

func (r *RedisOrderRepository) getOrderKey(orderID string) string {
	return fmt.Sprintf("order:orders:%s", orderID)
}

func (r *RedisOrderRepository) getCustomerOrdersKey(customerID string) string {
	return fmt.Sprintf("order:customer_orders:%s", customerID)
}

func (r *RedisOrderRepository) getOrdersByStatusKey(status domain.OrderStatus) string {
	return fmt.Sprintf("order:orders_by_status:%s", status.String())
}

func (r *RedisOrderRepository) getAllOrdersKey() string {
	return "order:all_orders"
}

// GetOrderStats retrieves order statistics (bonus method)
func (r *RedisOrderRepository) GetOrderStats(ctx context.Context, customerID string) (*OrderStats, error) {
	customerOrdersKey := r.getCustomerOrdersKey(customerID)

	// Get total order count
	totalOrders, err := r.client.ZCard(ctx, customerOrdersKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get order count: %w", err)
	}

	// Get recent orders to calculate stats
	recentOrderIDs, err := r.client.ZRevRange(ctx, customerOrdersKey, 0, 99).Result() // Last 100 orders
	if err != nil {
		return nil, fmt.Errorf("failed to get recent orders: %w", err)
	}

	var totalAmount int64
	statusCounts := make(map[string]int)

	for _, orderID := range recentOrderIDs {
		order, err := r.GetByID(ctx, orderID)
		if err != nil {
			continue
		}

		totalAmount += order.TotalAmount
		statusCounts[order.Status.String()]++
	}

	var averageAmount int64
	if len(recentOrderIDs) > 0 {
		averageAmount = totalAmount / int64(len(recentOrderIDs))
	}

	return &OrderStats{
		TotalOrders:   totalOrders,
		AverageAmount: averageAmount,
		StatusCounts:  statusCounts,
	}, nil
}

// OrderStats represents order statistics
type OrderStats struct {
	TotalOrders   int64          `json:"total_orders"`
	AverageAmount int64          `json:"average_amount"`
	StatusCounts  map[string]int `json:"status_counts"`
}
