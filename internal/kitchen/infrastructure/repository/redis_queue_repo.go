package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RedisQueueRepository implements domain.QueueRepository using Redis
type RedisQueueRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisQueueRepository creates a new Redis queue repository
func NewRedisQueueRepository(client *redis.Client, logger *logger.Logger) domain.QueueRepository {
	return &RedisQueueRepository{
		client: client,
		logger: logger,
	}
}

const (
	queueKey        = "kitchen:queue"
	queueStatusKey  = "kitchen:queue:status"
	queueHistoryKey = "kitchen:queue:history"
	queueMetricsKey = "kitchen:queue:metrics"
)

// SaveQueue saves the entire queue to Redis
func (r *RedisQueueRepository) SaveQueue(ctx context.Context, queue *domain.OrderQueue) error {
	orders := queue.GetAllOrders()
	orderIDs := make([]string, len(orders))
	
	for i, order := range orders {
		orderIDs[i] = order.ID()
	}

	// Store queue as a list of order IDs (maintaining order)
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, queueKey) // Clear existing queue
	
	if len(orderIDs) > 0 {
		pipe.LPush(ctx, queueKey, orderIDs) // Push in reverse order to maintain priority
	}
	
	// Store queue metadata
	metadata := map[string]interface{}{
		"updated_at": queue.GetUpdatedAt().Unix(),
		"length":     len(orders),
	}
	
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal queue metadata: %w", err)
	}
	
	pipe.Set(ctx, queueKey+":metadata", metadataJSON, 0)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).Error("Failed to save queue")
		return fmt.Errorf("failed to save queue: %w", err)
	}

	r.logger.WithField("queue_length", len(orders)).Info("Queue saved successfully")
	return nil
}

// LoadQueue loads the queue from Redis
func (r *RedisQueueRepository) LoadQueue(ctx context.Context) (*domain.OrderQueue, error) {
	// Get order IDs from queue
	orderIDs, err := r.client.LRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to load queue order IDs")
		return nil, fmt.Errorf("failed to load queue order IDs: %w", err)
	}

	queue := domain.NewOrderQueue()
	
	// Load each order and add to queue
	for _, orderID := range orderIDs {
		// Get order from order repository (we need access to it)
		// For now, we'll create a placeholder - in real implementation,
		// this would use the order repository
		orderKey := "kitchen:order:" + orderID
		orderData, err := r.client.Get(ctx, orderKey).Result()
		if err != nil {
			r.logger.WithError(err).WithField("order_id", orderID).Warn("Failed to load order for queue")
			continue
		}

		var dto domain.KitchenOrderDTO
		if err := json.Unmarshal([]byte(orderData), &dto); err != nil {
			r.logger.WithError(err).WithField("order_id", orderID).Error("Failed to unmarshal order")
			continue
		}

		order, err := r.dtoToOrder(&dto)
		if err != nil {
			r.logger.WithError(err).WithField("order_id", orderID).Error("Failed to convert DTO to order")
			continue
		}

		queue.AddOrder(order)
	}

	r.logger.WithField("queue_length", len(orderIDs)).Info("Queue loaded successfully")
	return queue, nil
}

// AddOrderToQueue adds an order to the queue
func (r *RedisQueueRepository) AddOrderToQueue(ctx context.Context, order *domain.KitchenOrder) error {
	// Add order ID to the queue list
	// Position depends on priority - higher priority goes to front
	priority := int(order.Priority())
	
	if priority >= int(domain.OrderPriorityHigh) {
		// High priority - add to front
		err := r.client.LPush(ctx, queueKey, order.ID()).Err()
		if err != nil {
			r.logger.WithError(err).WithField("order_id", order.ID()).Error("Failed to add high priority order to queue")
			return fmt.Errorf("failed to add order to queue: %w", err)
		}
	} else {
		// Normal/low priority - add to back
		err := r.client.RPush(ctx, queueKey, order.ID()).Err()
		if err != nil {
			r.logger.WithError(err).WithField("order_id", order.ID()).Error("Failed to add order to queue")
			return fmt.Errorf("failed to add order to queue: %w", err)
		}
	}

	r.logger.WithFields(map[string]interface{}{
		"order_id": order.ID(),
		"priority": order.Priority(),
	}).Info("Order added to queue")
	
	return nil
}

// RemoveOrderFromQueue removes an order from the queue
func (r *RedisQueueRepository) RemoveOrderFromQueue(ctx context.Context, orderID string) error {
	// Remove order from queue list
	removed, err := r.client.LRem(ctx, queueKey, 0, orderID).Result()
	if err != nil {
		r.logger.WithError(err).WithField("order_id", orderID).Error("Failed to remove order from queue")
		return fmt.Errorf("failed to remove order from queue: %w", err)
	}

	if removed == 0 {
		return fmt.Errorf("order not found in queue: %s", orderID)
	}

	r.logger.WithField("order_id", orderID).Info("Order removed from queue")
	return nil
}

// UpdateOrderInQueue updates an order in the queue (re-prioritization)
func (r *RedisQueueRepository) UpdateOrderInQueue(ctx context.Context, order *domain.KitchenOrder) error {
	// Remove and re-add to update position based on new priority
	if err := r.RemoveOrderFromQueue(ctx, order.ID()); err != nil {
		return err
	}
	
	return r.AddOrderToQueue(ctx, order)
}

// GetQueueStatus returns the current queue status
func (r *RedisQueueRepository) GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error) {
	// Get queue length
	queueLength, err := r.client.LLen(ctx, queueKey).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get queue length")
		return nil, fmt.Errorf("failed to get queue length: %w", err)
	}

	// Get cached status if available
	statusData, err := r.client.Get(ctx, queueStatusKey).Result()
	if err != nil && err != redis.Nil {
		r.logger.WithError(err).Error("Failed to get cached queue status")
	}

	var status *domain.QueueStatus
	if err == nil {
		// Use cached status
		if err := json.Unmarshal([]byte(statusData), &status); err != nil {
			r.logger.WithError(err).Error("Failed to unmarshal queue status")
		}
	}

	if status == nil {
		// Create new status
		status = &domain.QueueStatus{
			TotalOrders:      int32(queueLength),
			QueuesByPriority: make(map[domain.OrderPriority]int32),
			StationLoad:      make(map[domain.StationType]float32),
			UpdatedAt:        time.Now(),
		}
	}

	// Update with current queue length
	status.TotalOrders = int32(queueLength)
	status.UpdatedAt = time.Now()

	return status, nil
}

// SaveQueueStatus saves queue status to Redis
func (r *RedisQueueRepository) SaveQueueStatus(ctx context.Context, status *domain.QueueStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		r.logger.WithError(err).Error("Failed to marshal queue status")
		return fmt.Errorf("failed to marshal queue status: %w", err)
	}

	// Save current status
	err = r.client.Set(ctx, queueStatusKey, data, time.Hour).Err() // Cache for 1 hour
	if err != nil {
		r.logger.WithError(err).Error("Failed to save queue status")
		return fmt.Errorf("failed to save queue status: %w", err)
	}

	// Add to history for analytics
	timestamp := time.Now().Unix()
	err = r.client.ZAdd(ctx, queueHistoryKey, &redis.Z{
		Score:  float64(timestamp),
		Member: string(data),
	}).Err()
	if err != nil {
		r.logger.WithError(err).Warn("Failed to save queue status to history")
	}

	// Keep only last 24 hours of history
	dayAgo := time.Now().AddDate(0, 0, -1).Unix()
	r.client.ZRemRangeByScore(ctx, queueHistoryKey, "0", fmt.Sprintf("%d", dayAgo))

	return nil
}

// GetQueueHistory returns queue status history
func (r *RedisQueueRepository) GetQueueHistory(ctx context.Context, start, end time.Time) ([]*domain.QueueStatus, error) {
	// Get history from sorted set
	results, err := r.client.ZRangeByScore(ctx, queueHistoryKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", start.Unix()),
		Max: fmt.Sprintf("%d", end.Unix()),
	}).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get queue history")
		return nil, fmt.Errorf("failed to get queue history: %w", err)
	}

	history := make([]*domain.QueueStatus, 0, len(results))
	for _, result := range results {
		var status domain.QueueStatus
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			r.logger.WithError(err).Error("Failed to unmarshal queue status from history")
			continue
		}
		history = append(history, &status)
	}

	return history, nil
}

// GetAverageWaitTime returns average wait time
func (r *RedisQueueRepository) GetAverageWaitTime(ctx context.Context, start, end time.Time) (float64, error) {
	history, err := r.GetQueueHistory(ctx, start, end)
	if err != nil {
		return 0, err
	}

	if len(history) == 0 {
		return 0, nil
	}

	var totalWaitTime float64
	var count int

	for _, status := range history {
		if status.AverageWaitTime > 0 {
			totalWaitTime += float64(status.AverageWaitTime)
			count++
		}
	}

	if count == 0 {
		return 0, nil
	}

	return totalWaitTime / float64(count), nil
}

// GetThroughputStats returns throughput statistics
func (r *RedisQueueRepository) GetThroughputStats(ctx context.Context, start, end time.Time) (*domain.ThroughputStats, error) {
	history, err := r.GetQueueHistory(ctx, start, end)
	if err != nil {
		return nil, err
	}

	stats := &domain.ThroughputStats{
		CalculatedAt: time.Now(),
	}

	if len(history) == 0 {
		return stats, nil
	}

	var totalCompleted int32
	var maxQueueLength int32
	var totalQueueLength int32

	for _, status := range history {
		totalCompleted += status.CompletedOrders
		if status.TotalOrders > maxQueueLength {
			maxQueueLength = status.TotalOrders
		}
		totalQueueLength += status.TotalOrders
	}

	duration := end.Sub(start)
	hours := duration.Hours()
	days := duration.Hours() / 24

	if hours > 0 {
		stats.OrdersPerHour = float32(totalCompleted) / float32(hours)
	}
	if days > 0 {
		stats.OrdersPerDay = float32(totalCompleted) / float32(days)
	}

	stats.MaxQueueLength = maxQueueLength
	if len(history) > 0 {
		stats.AverageQueueLength = float32(totalQueueLength) / float32(len(history))
	}

	// Calculate peak hour throughput (simplified)
	stats.PeakHourThroughput = stats.OrdersPerHour * 1.5 // Estimate

	return stats, nil
}

// Helper method to convert DTO to order (simplified version)
func (r *RedisQueueRepository) dtoToOrder(dto *domain.KitchenOrderDTO) (*domain.KitchenOrder, error) {
	// Convert DTO items to domain items
	items := make([]*domain.OrderItem, len(dto.Items))
	for i, itemDTO := range dto.Items {
		items[i] = domain.NewOrderItem(itemDTO.ID, itemDTO.Name, itemDTO.Quantity, itemDTO.Requirements)
		items[i].SetInstructions(itemDTO.Instructions)
		items[i].SetMetadata(itemDTO.Metadata)
	}

	order, err := domain.NewKitchenOrder(dto.ID, dto.CustomerID, items)
	if err != nil {
		return nil, err
	}

	// Set additional fields
	order.SetPriority(dto.Priority)
	order.SetSpecialInstructions(dto.SpecialInstructions)
	order.SetEstimatedTime(dto.EstimatedTime)
	
	if dto.AssignedStaffID != "" {
		order.AssignStaff(dto.AssignedStaffID)
	}
	
	if len(dto.AssignedEquipment) > 0 {
		order.AssignEquipment(dto.AssignedEquipment)
	}
	
	// Update status (this will set timestamps appropriately)
	if dto.Status != domain.OrderStatusPending {
		order.UpdateStatus(dto.Status)
	}

	return order, nil
}
