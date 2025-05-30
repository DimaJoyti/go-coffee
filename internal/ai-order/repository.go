package aiorder

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/DimaJoyti/go-coffee/api/proto/ai_order"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Repository defines the interface for order data operations
type Repository interface {
	CreateOrder(ctx context.Context, order *pb.Order) error
	GetOrder(ctx context.Context, orderID string) (*pb.Order, error)
	UpdateOrder(ctx context.Context, order *pb.Order) error
	ListOrders(ctx context.Context, filter *ListOrdersFilter) ([]*pb.Order, int32, error)
	DeleteOrder(ctx context.Context, orderID string) error
}

// ListOrdersFilter defines filtering options for listing orders
type ListOrdersFilter struct {
	CustomerID string
	LocationID string
	Status     pb.OrderStatus
	FromDate   *timestamppb.Timestamp
	ToDate     *timestamppb.Timestamp
	PageSize   int32
	PageToken  string
}

// RedisOrderRepository implements Repository using Redis
type RedisOrderRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisOrderRepository creates a new Redis-based order repository
func NewRedisOrderRepository(client *redis.Client, logger *logger.Logger) Repository {
	return &RedisOrderRepository{
		client: client,
		logger: logger,
	}
}

// CreateOrder stores a new order in Redis
func (r *RedisOrderRepository) CreateOrder(ctx context.Context, order *pb.Order) error {
	r.logger.Info("Creating order in Redis", zap.String("order_id", order.Id))

	// Serialize order to JSON
	orderData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Store order data
	orderKey := fmt.Sprintf("ai:order:%s", order.Id)
	pipe.Set(ctx, orderKey, orderData, 24*time.Hour) // TTL: 24 hours

	// Add to customer orders index
	if order.Customer != nil && order.Customer.Id != "" {
		customerOrdersKey := fmt.Sprintf("ai:customer:%s:orders", order.Customer.Id)
		pipe.SAdd(ctx, customerOrdersKey, order.Id)
		pipe.Expire(ctx, customerOrdersKey, 7*24*time.Hour) // TTL: 7 days
	}

	// Add to location orders index
	if order.LocationId != "" {
		locationOrdersKey := fmt.Sprintf("ai:location:%s:orders", order.LocationId)
		pipe.ZAdd(ctx, locationOrdersKey, &redis.Z{
			Score:  float64(order.CreatedAt.Seconds),
			Member: order.Id,
		})
		pipe.Expire(ctx, locationOrdersKey, 7*24*time.Hour) // TTL: 7 days
	}

	// Add to status index
	statusOrdersKey := fmt.Sprintf("ai:status:%s:orders", order.Status.String())
	pipe.ZAdd(ctx, statusOrdersKey, &redis.Z{
		Score:  float64(order.CreatedAt.Seconds),
		Member: order.Id,
	})
	pipe.Expire(ctx, statusOrdersKey, 7*24*time.Hour) // TTL: 7 days

	// Add to global orders index
	pipe.ZAdd(ctx, "ai:orders:all", &redis.Z{
		Score:  float64(order.CreatedAt.Seconds),
		Member: order.Id,
	})

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}

	r.logger.Info("Order created successfully in Redis", zap.String("order_id", order.Id))
	return nil
}

// GetOrder retrieves an order from Redis by ID
func (r *RedisOrderRepository) GetOrder(ctx context.Context, orderID string) (*pb.Order, error) {
	r.logger.Info("Getting order from Redis", zap.String("order_id", orderID))

	orderKey := fmt.Sprintf("ai:order:%s", orderID)
	orderData, err := r.client.Get(ctx, orderKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("order not found: %s", orderID)
		}
		return nil, fmt.Errorf("failed to get order from Redis: %w", err)
	}

	var order pb.Order
	if err := json.Unmarshal([]byte(orderData), &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

// UpdateOrder updates an existing order in Redis
func (r *RedisOrderRepository) UpdateOrder(ctx context.Context, order *pb.Order) error {
	r.logger.Info("Updating order in Redis", zap.String("order_id", order.Id))

	// Check if order exists
	orderKey := fmt.Sprintf("ai:order:%s", order.Id)
	exists, err := r.client.Exists(ctx, orderKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check order existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("order not found: %s", order.Id)
	}

	// Serialize updated order
	orderData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Update order data
	pipe.Set(ctx, orderKey, orderData, 24*time.Hour)

	// Update status index (remove from old status, add to new status)
	// Note: In a production system, you'd want to track the previous status
	statusOrdersKey := fmt.Sprintf("ai:status:%s:orders", order.Status.String())
	pipe.ZAdd(ctx, statusOrdersKey, &redis.Z{
		Score:  float64(order.UpdatedAt.Seconds),
		Member: order.Id,
	})

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}

	r.logger.Info("Order updated successfully in Redis", zap.String("order_id", order.Id))
	return nil
}

// ListOrders retrieves orders based on filter criteria
func (r *RedisOrderRepository) ListOrders(ctx context.Context, filter *ListOrdersFilter) ([]*pb.Order, int32, error) {
	r.logger.Info("Listing orders from Redis",
		zap.String("customer_id", filter.CustomerID),
		zap.String("location_id", filter.LocationID),
		zap.String("status", filter.Status.String()),
	)

	var orderIDs []string
	var err error

	// Determine which index to use based on filter
	switch {
	case filter.CustomerID != "":
		orderIDs, err = r.getOrderIDsByCustomer(ctx, filter.CustomerID)
	case filter.LocationID != "":
		orderIDs, err = r.getOrderIDsByLocation(ctx, filter.LocationID, filter.FromDate, filter.ToDate)
	case filter.Status != pb.OrderStatus_ORDER_STATUS_UNSPECIFIED:
		orderIDs, err = r.getOrderIDsByStatus(ctx, filter.Status, filter.FromDate, filter.ToDate)
	default:
		orderIDs, err = r.getAllOrderIDs(ctx, filter.FromDate, filter.ToDate)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order IDs: %w", err)
	}

	// Apply pagination
	totalCount := int32(len(orderIDs))
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	// Simple pagination implementation
	start := 0
	if filter.PageToken != "" {
		if pageNum, err := strconv.Atoi(filter.PageToken); err == nil {
			start = pageNum * int(pageSize)
		}
	}

	end := start + int(pageSize)
	if end > len(orderIDs) {
		end = len(orderIDs)
	}

	if start >= len(orderIDs) {
		return []*pb.Order{}, totalCount, nil
	}

	paginatedOrderIDs := orderIDs[start:end]

	// Fetch orders
	orders := make([]*pb.Order, 0, len(paginatedOrderIDs))
	for _, orderID := range paginatedOrderIDs {
		order, err := r.GetOrder(ctx, orderID)
		if err != nil {
			r.logger.Warn("Failed to get order", zap.String("order_id", orderID), zap.Error(err))
			continue
		}
		orders = append(orders, order)
	}

	return orders, totalCount, nil
}

// DeleteOrder removes an order from Redis
func (r *RedisOrderRepository) DeleteOrder(ctx context.Context, orderID string) error {
	r.logger.Info("Deleting order from Redis", zap.String("order_id", orderID))

	// Get order first to clean up indexes
	order, err := r.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order for deletion: %w", err)
	}

	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Delete order data
	orderKey := fmt.Sprintf("ai:order:%s", orderID)
	pipe.Del(ctx, orderKey)

	// Remove from customer orders index
	if order.Customer != nil && order.Customer.Id != "" {
		customerOrdersKey := fmt.Sprintf("ai:customer:%s:orders", order.Customer.Id)
		pipe.SRem(ctx, customerOrdersKey, orderID)
	}

	// Remove from location orders index
	if order.LocationId != "" {
		locationOrdersKey := fmt.Sprintf("ai:location:%s:orders", order.LocationId)
		pipe.ZRem(ctx, locationOrdersKey, orderID)
	}

	// Remove from status index
	statusOrdersKey := fmt.Sprintf("ai:status:%s:orders", order.Status.String())
	pipe.ZRem(ctx, statusOrdersKey, orderID)

	// Remove from global orders index
	pipe.ZRem(ctx, "ai:orders:all", orderID)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}

	r.logger.Info("Order deleted successfully from Redis", zap.String("order_id", orderID))
	return nil
}

// Helper methods for different query patterns

func (r *RedisOrderRepository) getOrderIDsByCustomer(ctx context.Context, customerID string) ([]string, error) {
	customerOrdersKey := fmt.Sprintf("ai:customer:%s:orders", customerID)
	orderIDs, err := r.client.SMembers(ctx, customerOrdersKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get customer orders: %w", err)
	}
	return orderIDs, nil
}

func (r *RedisOrderRepository) getOrderIDsByLocation(ctx context.Context, locationID string, fromDate, toDate *timestamppb.Timestamp) ([]string, error) {
	locationOrdersKey := fmt.Sprintf("ai:location:%s:orders", locationID)
	
	min := "-inf"
	max := "+inf"
	
	if fromDate != nil {
		min = fmt.Sprintf("%d", fromDate.Seconds)
	}
	if toDate != nil {
		max = fmt.Sprintf("%d", toDate.Seconds)
	}

	orderIDs, err := r.client.ZRangeByScore(ctx, locationOrdersKey, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get location orders: %w", err)
	}
	return orderIDs, nil
}

func (r *RedisOrderRepository) getOrderIDsByStatus(ctx context.Context, status pb.OrderStatus, fromDate, toDate *timestamppb.Timestamp) ([]string, error) {
	statusOrdersKey := fmt.Sprintf("ai:status:%s:orders", status.String())
	
	min := "-inf"
	max := "+inf"
	
	if fromDate != nil {
		min = fmt.Sprintf("%d", fromDate.Seconds)
	}
	if toDate != nil {
		max = fmt.Sprintf("%d", toDate.Seconds)
	}

	orderIDs, err := r.client.ZRangeByScore(ctx, statusOrdersKey, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get status orders: %w", err)
	}
	return orderIDs, nil
}

func (r *RedisOrderRepository) getAllOrderIDs(ctx context.Context, fromDate, toDate *timestamppb.Timestamp) ([]string, error) {
	min := "-inf"
	max := "+inf"
	
	if fromDate != nil {
		min = fmt.Sprintf("%d", fromDate.Seconds)
	}
	if toDate != nil {
		max = fmt.Sprintf("%d", toDate.Seconds)
	}

	orderIDs, err := r.client.ZRangeByScore(ctx, "ai:orders:all", &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}
	return orderIDs, nil
}
