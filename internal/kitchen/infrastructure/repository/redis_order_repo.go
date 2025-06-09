package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RedisOrderRepository implements domain.OrderRepository using Redis
type RedisOrderRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisOrderRepository creates a new Redis order repository
func NewRedisOrderRepository(client *redis.Client, logger *logger.Logger) domain.OrderRepository {
	return &RedisOrderRepository{
		client: client,
		logger: logger,
	}
}

const (
	orderKeyPrefix     = "kitchen:order:"
	orderSetKey        = "kitchen:order:all"
	orderByStatus      = "kitchen:order:by_status:"
	orderByPriority    = "kitchen:order:by_priority:"
	orderByCustomer    = "kitchen:order:by_customer:"
	orderByStaff       = "kitchen:order:by_staff:"
	orderByDate        = "kitchen:order:by_date:"
	orderOverdueKey    = "kitchen:order:overdue"
)

// Create saves a new order to Redis
func (r *RedisOrderRepository) Create(ctx context.Context, order *domain.KitchenOrder) error {
	key := orderKeyPrefix + order.ID()
	
	// Convert order to DTO for storage
	dto := order.ToDTO()
	data, err := json.Marshal(dto)
	if err != nil {
		r.logger.WithError(err).Error("Failed to marshal order")
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	pipe := r.client.TxPipeline()
	
	// Store order data
	pipe.Set(ctx, key, data, 0)
	
	// Add to order set
	pipe.SAdd(ctx, orderSetKey, order.ID())
	
	// Add to status-based set
	statusKey := orderByStatus + strconv.Itoa(int(order.Status()))
	pipe.SAdd(ctx, statusKey, order.ID())
	
	// Add to priority-based set
	priorityKey := orderByPriority + strconv.Itoa(int(order.Priority()))
	pipe.SAdd(ctx, priorityKey, order.ID())
	
	// Add to customer-based set
	customerKey := orderByCustomer + order.CustomerID()
	pipe.SAdd(ctx, customerKey, order.ID())
	
	// Add to date-based sorted set (for time-based queries)
	dateKey := orderByDate + order.CreatedAt().Format("2006-01-02")
	pipe.ZAdd(ctx, dateKey, &redis.Z{
		Score:  float64(order.CreatedAt().Unix()),
		Member: order.ID(),
	})
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("order_id", order.ID()).Error("Failed to create order")
		return fmt.Errorf("failed to create order: %w", err)
	}

	r.logger.WithField("order_id", order.ID()).Info("Order created successfully")
	return nil
}

// GetByID retrieves order by ID
func (r *RedisOrderRepository) GetByID(ctx context.Context, id string) (*domain.KitchenOrder, error) {
	key := orderKeyPrefix + id
	
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("order not found: %s", id)
		}
		r.logger.WithError(err).WithField("order_id", id).Error("Failed to get order")
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	var dto domain.KitchenOrderDTO
	if err := json.Unmarshal([]byte(data), &dto); err != nil {
		r.logger.WithError(err).Error("Failed to unmarshal order")
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return r.dtoToOrder(&dto)
}

// Update updates existing order
func (r *RedisOrderRepository) Update(ctx context.Context, order *domain.KitchenOrder) error {
	// Get existing order to check for changes
	existing, err := r.GetByID(ctx, order.ID())
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	key := orderKeyPrefix + order.ID()
	
	// Convert order to DTO for storage
	dto := order.ToDTO()
	data, err := json.Marshal(dto)
	if err != nil {
		r.logger.WithError(err).Error("Failed to marshal order")
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	pipe := r.client.TxPipeline()
	
	// Update order data
	pipe.Set(ctx, key, data, 0)
	
	// Update status-based sets if status changed
	if existing.Status() != order.Status() {
		oldStatusKey := orderByStatus + strconv.Itoa(int(existing.Status()))
		newStatusKey := orderByStatus + strconv.Itoa(int(order.Status()))
		
		pipe.SRem(ctx, oldStatusKey, order.ID())
		pipe.SAdd(ctx, newStatusKey, order.ID())
	}
	
	// Update priority-based sets if priority changed
	if existing.Priority() != order.Priority() {
		oldPriorityKey := orderByPriority + strconv.Itoa(int(existing.Priority()))
		newPriorityKey := orderByPriority + strconv.Itoa(int(order.Priority()))
		
		pipe.SRem(ctx, oldPriorityKey, order.ID())
		pipe.SAdd(ctx, newPriorityKey, order.ID())
	}
	
	// Update staff assignment if changed
	if existing.AssignedStaffID() != order.AssignedStaffID() {
		if existing.AssignedStaffID() != "" {
			oldStaffKey := orderByStaff + existing.AssignedStaffID()
			pipe.SRem(ctx, oldStaffKey, order.ID())
		}
		if order.AssignedStaffID() != "" {
			newStaffKey := orderByStaff + order.AssignedStaffID()
			pipe.SAdd(ctx, newStaffKey, order.ID())
		}
	}
	
	// Update overdue status
	if order.IsOverdue() {
		pipe.SAdd(ctx, orderOverdueKey, order.ID())
	} else {
		pipe.SRem(ctx, orderOverdueKey, order.ID())
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("order_id", order.ID()).Error("Failed to update order")
		return fmt.Errorf("failed to update order: %w", err)
	}

	r.logger.WithField("order_id", order.ID()).Info("Order updated successfully")
	return nil
}

// Delete removes order from Redis
func (r *RedisOrderRepository) Delete(ctx context.Context, id string) error {
	order, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	key := orderKeyPrefix + id
	
	pipe := r.client.TxPipeline()
	
	// Delete order data
	pipe.Del(ctx, key)
	
	// Remove from all sets
	pipe.SRem(ctx, orderSetKey, id)
	pipe.SRem(ctx, orderOverdueKey, id)
	
	statusKey := orderByStatus + strconv.Itoa(int(order.Status()))
	pipe.SRem(ctx, statusKey, id)
	
	priorityKey := orderByPriority + strconv.Itoa(int(order.Priority()))
	pipe.SRem(ctx, priorityKey, id)
	
	customerKey := orderByCustomer + order.CustomerID()
	pipe.SRem(ctx, customerKey, id)
	
	if order.AssignedStaffID() != "" {
		staffKey := orderByStaff + order.AssignedStaffID()
		pipe.SRem(ctx, staffKey, id)
	}
	
	dateKey := orderByDate + order.CreatedAt().Format("2006-01-02")
	pipe.ZRem(ctx, dateKey, id)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("order_id", id).Error("Failed to delete order")
		return fmt.Errorf("failed to delete order: %w", err)
	}

	r.logger.WithField("order_id", id).Info("Order deleted successfully")
	return nil
}

// GetAll retrieves all orders
func (r *RedisOrderRepository) GetAll(ctx context.Context) ([]*domain.KitchenOrder, error) {
	ids, err := r.client.SMembers(ctx, orderSetKey).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get order IDs")
		return nil, fmt.Errorf("failed to get order IDs: %w", err)
	}

	return r.getOrdersByIDs(ctx, ids)
}

// GetByStatus retrieves orders by status
func (r *RedisOrderRepository) GetByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.KitchenOrder, error) {
	statusKey := orderByStatus + strconv.Itoa(int(status))
	ids, err := r.client.SMembers(ctx, statusKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("status", status).Error("Failed to get order IDs by status")
		return nil, fmt.Errorf("failed to get order IDs by status: %w", err)
	}

	return r.getOrdersByIDs(ctx, ids)
}

// GetByPriority retrieves orders by priority
func (r *RedisOrderRepository) GetByPriority(ctx context.Context, priority domain.OrderPriority) ([]*domain.KitchenOrder, error) {
	priorityKey := orderByPriority + strconv.Itoa(int(priority))
	ids, err := r.client.SMembers(ctx, priorityKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("priority", priority).Error("Failed to get order IDs by priority")
		return nil, fmt.Errorf("failed to get order IDs by priority: %w", err)
	}

	return r.getOrdersByIDs(ctx, ids)
}

// GetByCustomerID retrieves orders by customer ID
func (r *RedisOrderRepository) GetByCustomerID(ctx context.Context, customerID string) ([]*domain.KitchenOrder, error) {
	customerKey := orderByCustomer + customerID
	ids, err := r.client.SMembers(ctx, customerKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("customer_id", customerID).Error("Failed to get order IDs by customer")
		return nil, fmt.Errorf("failed to get order IDs by customer: %w", err)
	}

	return r.getOrdersByIDs(ctx, ids)
}

// GetByStaffID retrieves orders by staff ID
func (r *RedisOrderRepository) GetByStaffID(ctx context.Context, staffID string) ([]*domain.KitchenOrder, error) {
	staffKey := orderByStaff + staffID
	ids, err := r.client.SMembers(ctx, staffKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("staff_id", staffID).Error("Failed to get order IDs by staff")
		return nil, fmt.Errorf("failed to get order IDs by staff: %w", err)
	}

	return r.getOrdersByIDs(ctx, ids)
}

// GetByDateRange retrieves orders by date range
func (r *RedisOrderRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*domain.KitchenOrder, error) {
	var allIDs []string
	
	// Iterate through each day in the range
	current := start
	for current.Before(end) || current.Equal(end) {
		dateKey := orderByDate + current.Format("2006-01-02")
		
		// Get orders for this day within time range
		ids, err := r.client.ZRangeByScore(ctx, dateKey, &redis.ZRangeBy{
			Min: strconv.FormatInt(start.Unix(), 10),
			Max: strconv.FormatInt(end.Unix(), 10),
		}).Result()
		
		if err != nil && err != redis.Nil {
			r.logger.WithError(err).WithField("date", current.Format("2006-01-02")).Error("Failed to get order IDs by date")
			return nil, fmt.Errorf("failed to get order IDs by date: %w", err)
		}
		
		allIDs = append(allIDs, ids...)
		current = current.AddDate(0, 0, 1)
	}

	return r.getOrdersByIDs(ctx, allIDs)
}

// UpdateStatus updates order status
func (r *RedisOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	order, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := order.UpdateStatus(status); err != nil {
		return err
	}

	return r.Update(ctx, order)
}

// UpdatePriority updates order priority
func (r *RedisOrderRepository) UpdatePriority(ctx context.Context, id string, priority domain.OrderPriority) error {
	order, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	order.SetPriority(priority)
	return r.Update(ctx, order)
}

// AssignStaff assigns staff to order
func (r *RedisOrderRepository) AssignStaff(ctx context.Context, id string, staffID string) error {
	order, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := order.AssignStaff(staffID); err != nil {
		return err
	}

	return r.Update(ctx, order)
}

// AssignEquipment assigns equipment to order
func (r *RedisOrderRepository) AssignEquipment(ctx context.Context, id string, equipmentIDs []string) error {
	order, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := order.AssignEquipment(equipmentIDs); err != nil {
		return err
	}

	return r.Update(ctx, order)
}

// GetOverdue retrieves overdue orders
func (r *RedisOrderRepository) GetOverdue(ctx context.Context) ([]*domain.KitchenOrder, error) {
	ids, err := r.client.SMembers(ctx, orderOverdueKey).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get overdue order IDs")
		return nil, fmt.Errorf("failed to get overdue order IDs: %w", err)
	}

	return r.getOrdersByIDs(ctx, ids)
}

// GetByRequiredStation retrieves orders requiring a specific station
func (r *RedisOrderRepository) GetByRequiredStation(ctx context.Context, stationType domain.StationType) ([]*domain.KitchenOrder, error) {
	// This requires scanning all orders since we don't index by required stations
	// In a production system, you might want to add this indexing
	allOrders, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.KitchenOrder
	for _, order := range allOrders {
		requiredStations := order.GetRequiredStations()
		for _, required := range requiredStations {
			if required == stationType {
				filtered = append(filtered, order)
				break
			}
		}
	}

	return filtered, nil
}

// GetCompletionStats returns completion statistics
func (r *RedisOrderRepository) GetCompletionStats(ctx context.Context, start, end time.Time) (*domain.OrderCompletionStats, error) {
	orders, err := r.GetByDateRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	stats := &domain.OrderCompletionStats{
		CalculatedAt: time.Now(),
	}

	var totalTime float64
	var times []float64

	for _, order := range orders {
		stats.TotalOrders++
		
		switch order.Status() {
		case domain.OrderStatusCompleted:
			stats.CompletedOrders++
			if order.ActualTime() > 0 {
				time := float64(order.ActualTime())
				totalTime += time
				times = append(times, time)
				
				// Check if on time
				if order.EstimatedTime() > 0 && order.ActualTime() <= order.EstimatedTime() {
					// On time
				}
			}
		case domain.OrderStatusCancelled:
			stats.CancelledOrders++
		}
	}

	if stats.CompletedOrders > 0 {
		stats.AverageTime = totalTime / float64(stats.CompletedOrders)
		stats.CompletionRate = float32(stats.CompletedOrders) / float32(stats.TotalOrders)
		
		// Calculate median (simplified)
		if len(times) > 0 {
			// Sort times and get median
			stats.MedianTime = times[len(times)/2]
		}
	}

	return stats, nil
}

// GetAverageProcessingTime returns average processing time
func (r *RedisOrderRepository) GetAverageProcessingTime(ctx context.Context, start, end time.Time) (float64, error) {
	stats, err := r.GetCompletionStats(ctx, start, end)
	if err != nil {
		return 0, err
	}
	return stats.AverageTime, nil
}

// GetOrderCountByStatus returns order count by status
func (r *RedisOrderRepository) GetOrderCountByStatus(ctx context.Context) (map[domain.OrderStatus]int32, error) {
	counts := make(map[domain.OrderStatus]int32)
	
	statuses := []domain.OrderStatus{
		domain.OrderStatusPending,
		domain.OrderStatusProcessing,
		domain.OrderStatusCompleted,
		domain.OrderStatusCancelled,
	}

	for _, status := range statuses {
		statusKey := orderByStatus + strconv.Itoa(int(status))
		count, err := r.client.SCard(ctx, statusKey).Result()
		if err != nil {
			r.logger.WithError(err).WithField("status", status).Error("Failed to get order count by status")
			continue
		}
		counts[status] = int32(count)
	}

	return counts, nil
}

// Helper methods

func (r *RedisOrderRepository) getOrdersByIDs(ctx context.Context, ids []string) ([]*domain.KitchenOrder, error) {
	if len(ids) == 0 {
		return []*domain.KitchenOrder{}, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = orderKeyPrefix + id
	}

	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get order data")
		return nil, fmt.Errorf("failed to get order data: %w", err)
	}

	orders := make([]*domain.KitchenOrder, 0, len(results))
	for i, result := range results {
		if result == nil {
			r.logger.WithField("order_id", ids[i]).Warn("Order data not found")
			continue
		}

		var dto domain.KitchenOrderDTO
		if err := json.Unmarshal([]byte(result.(string)), &dto); err != nil {
			r.logger.WithError(err).WithField("order_id", ids[i]).Error("Failed to unmarshal order")
			continue
		}

		order, err := r.dtoToOrder(&dto)
		if err != nil {
			r.logger.WithError(err).WithField("order_id", ids[i]).Error("Failed to convert DTO to order")
			continue
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *RedisOrderRepository) dtoToOrder(dto *domain.KitchenOrderDTO) (*domain.KitchenOrder, error) {
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
