package application

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// QueueServiceImpl implements the QueueService interface
type QueueServiceImpl struct {
	repoManager      domain.RepositoryManager
	optimizerService OptimizerService
	eventService     EventService
	logger           *logger.Logger
	queue            *domain.OrderQueue
}

// NewQueueService creates a new queue service instance
func NewQueueService(
	repoManager domain.RepositoryManager,
	optimizerService OptimizerService,
	eventService EventService,
	logger *logger.Logger,
) QueueService {
	return &QueueServiceImpl{
		repoManager:      repoManager,
		optimizerService: optimizerService,
		eventService:     eventService,
		logger:           logger,
		queue:            domain.NewOrderQueue(),
	}
}

// AddOrder adds an order to the queue
func (s *QueueServiceImpl) AddOrder(ctx context.Context, order *domain.KitchenOrder) error {
	s.logger.WithFields(map[string]interface{}{
		"order_id": order.ID(),
		"priority": order.Priority(),
	}).Info("Adding order to queue")

	// Add order to in-memory queue
	if err := s.queue.AddOrder(order); err != nil {
		s.logger.WithError(err).Error("Failed to add order to in-memory queue")
		return fmt.Errorf("failed to add order to queue: %w", err)
	}

	// Persist queue state
	if err := s.repoManager.Queue().AddOrderToQueue(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to persist order to queue")
		// Try to remove from in-memory queue to maintain consistency
		if removeErr := s.queue.RemoveOrder(order.ID()); removeErr != nil {
			s.logger.WithError(removeErr).Error("Failed to remove order from in-memory queue")
		}
		return fmt.Errorf("failed to persist order to queue: %w", err)
	}

	// Update queue status
	status := s.queue.GetQueueStatus()
	if err := s.repoManager.Queue().SaveQueueStatus(ctx, status); err != nil {
		s.logger.WithError(err).Warn("Failed to save queue status")
	}

	// Publish event
	event := domain.NewOrderAddedToQueueEvent(order)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish order added to queue event")
	}

	s.logger.WithField("order_id", order.ID()).Info("Order added to queue successfully")
	return nil
}

// RemoveOrder removes an order from the queue
func (s *QueueServiceImpl) RemoveOrder(ctx context.Context, orderID string) error {
	s.logger.WithField("order_id", orderID).Info("Removing order from queue")

	// Remove from in-memory queue
	if err := s.queue.RemoveOrder(orderID); err != nil {
		s.logger.WithError(err).Error("Failed to remove order from in-memory queue")
		return fmt.Errorf("failed to remove order from queue: %w", err)
	}

	// Persist queue state
	if err := s.repoManager.Queue().RemoveOrderFromQueue(ctx, orderID); err != nil {
		s.logger.WithError(err).Error("Failed to remove order from persistent queue")
		return fmt.Errorf("failed to remove order from persistent queue: %w", err)
	}

	// Update queue status
	status := s.queue.GetQueueStatus()
	if err := s.repoManager.Queue().SaveQueueStatus(ctx, status); err != nil {
		s.logger.WithError(err).Warn("Failed to save queue status")
	}

	s.logger.WithField("order_id", orderID).Info("Order removed from queue successfully")
	return nil
}

// GetNextOrder returns the next order to be processed
func (s *QueueServiceImpl) GetNextOrder(ctx context.Context) *domain.KitchenOrder {
	nextOrder := s.queue.GetNextOrder()
	if nextOrder != nil {
		s.logger.WithFields(map[string]interface{}{
			"order_id": nextOrder.ID(),
			"priority": nextOrder.Priority(),
		}).Info("Retrieved next order from queue")
	} else {
		s.logger.Info("No orders available in queue")
	}
	return nextOrder
}

// UpdateOrderPriority updates the priority of an order in the queue
func (s *QueueServiceImpl) UpdateOrderPriority(ctx context.Context, orderID string, priority domain.OrderPriority) error {
	s.logger.WithFields(map[string]interface{}{
		"order_id": orderID,
		"priority": priority,
	}).Info("Updating order priority in queue")

	// Update in-memory queue
	if err := s.queue.UpdateOrderPriority(orderID, priority); err != nil {
		s.logger.WithError(err).Error("Failed to update order priority in in-memory queue")
		return fmt.Errorf("failed to update order priority in queue: %w", err)
	}

	// Get the updated order
	order := s.queue.GetOrder(orderID)
	if order == nil {
		return fmt.Errorf("order not found in queue: %s", orderID)
	}

	// Persist queue state
	if err := s.repoManager.Queue().UpdateOrderInQueue(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to update order in persistent queue")
		return fmt.Errorf("failed to update order in persistent queue: %w", err)
	}

	// Update queue status
	status := s.queue.GetQueueStatus()
	if err := s.repoManager.Queue().SaveQueueStatus(ctx, status); err != nil {
		s.logger.WithError(err).Warn("Failed to save queue status")
	}

	// Publish event
	event := domain.NewOrderPriorityChangedEvent(order)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish order priority changed event")
	}

	s.logger.WithField("order_id", orderID).Info("Order priority updated successfully")
	return nil
}

// GetQueueStatus returns the current queue status
func (s *QueueServiceImpl) GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error) {
	status := s.queue.GetQueueStatus()

	s.logger.WithFields(map[string]interface{}{
		"total_orders":      status.TotalOrders,
		"pending_orders":    status.PendingOrders,
		"processing_orders": status.ProcessingOrders,
	}).Info("Retrieved queue status")

	return status, nil
}

// GetEstimatedWaitTime calculates estimated wait time for an order
func (s *QueueServiceImpl) GetEstimatedWaitTime(ctx context.Context, order *domain.KitchenOrder) (time.Duration, error) {
	waitTime := s.queue.GetEstimatedWaitTime(order)

	s.logger.WithFields(map[string]interface{}{
		"order_id":  order.ID(),
		"wait_time": waitTime.String(),
	}).Info("Calculated estimated wait time")

	return waitTime, nil
}

// GetOverdueOrders returns orders that are overdue
func (s *QueueServiceImpl) GetOverdueOrders(ctx context.Context) ([]*domain.KitchenOrder, error) {
	overdueOrders := s.queue.GetOverdueOrders()

	s.logger.WithField("overdue_count", len(overdueOrders)).Info("Retrieved overdue orders")

	// Publish events for overdue orders
	for _, order := range overdueOrders {
		event := domain.NewOrderOverdueEvent(order)
		if err := s.eventService.PublishEvent(ctx, event); err != nil {
			s.logger.WithError(err).WithField("order_id", order.ID()).Warn("Failed to publish order overdue event")
		}
	}

	return overdueOrders, nil
}

// OptimizeQueue optimizes the current queue using AI
func (s *QueueServiceImpl) OptimizeQueue(ctx context.Context) (*domain.WorkflowOptimization, error) {
	s.logger.Info("Starting queue optimization")

	// Get all orders in queue
	orders := s.queue.GetAllOrders()
	if len(orders) == 0 {
		s.logger.Info("No orders in queue to optimize")
		return nil, fmt.Errorf("no orders in queue to optimize")
	}

	// Use AI optimizer to optimize workflow
	optimization, err := s.optimizerService.OptimizeWorkflow(ctx, orders)
	if err != nil {
		s.logger.WithError(err).Error("Failed to optimize workflow")
		return nil, fmt.Errorf("failed to optimize workflow: %w", err)
	}

	// Save optimization results
	if err := s.repoManager.Workflow().SaveOptimization(ctx, optimization); err != nil {
		s.logger.WithError(err).Warn("Failed to save optimization results")
	}

	// Publish optimization event
	event := domain.NewWorkflowOptimizedEvent(optimization)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish workflow optimized event")
	}

	s.logger.WithFields(map[string]interface{}{
		"order_id":        optimization.OrderID,
		"efficiency_gain": optimization.EfficiencyGain,
		"estimated_time":  optimization.EstimatedTime,
	}).Info("Queue optimization completed")

	return optimization, nil
}

// RebalanceQueue rebalances the queue based on current conditions
func (s *QueueServiceImpl) RebalanceQueue(ctx context.Context) error {
	s.logger.Info("Starting queue rebalancing")

	// Get current queue status
	status := s.queue.GetQueueStatus()

	// Check if rebalancing is needed
	if status.PendingOrders < 5 {
		s.logger.Info("Queue rebalancing not needed - low order count")
		return nil
	}

	// Get all pending orders
	pendingOrders := s.queue.GetOrdersByStatus(domain.OrderStatusPending)

	// Sort by priority and wait time
	// This is a simplified rebalancing - in production, you'd use more sophisticated algorithms
	for _, order := range pendingOrders {
		waitTime := order.GetWaitTime()

		// Increase priority for orders waiting too long
		if waitTime > 15*time.Minute && order.Priority() < domain.OrderPriorityHigh {
			newPriority := order.Priority() + 1
			if newPriority > domain.OrderPriorityUrgent {
				newPriority = domain.OrderPriorityUrgent
			}

			if err := s.UpdateOrderPriority(ctx, order.ID(), newPriority); err != nil {
				s.logger.WithError(err).WithField("order_id", order.ID()).Warn("Failed to update order priority during rebalancing")
			} else {
				s.logger.WithFields(map[string]interface{}{
					"order_id":     order.ID(),
					"old_priority": order.Priority(),
					"new_priority": newPriority,
					"wait_time":    waitTime.String(),
				}).Info("Increased order priority due to long wait time")
			}
		}
	}

	// Update queue status after rebalancing
	newStatus := s.queue.GetQueueStatus()
	if err := s.repoManager.Queue().SaveQueueStatus(ctx, newStatus); err != nil {
		s.logger.WithError(err).Warn("Failed to save queue status after rebalancing")
	}

	// Publish queue status changed event
	event := domain.NewQueueStatusChangedEvent(newStatus)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish queue status changed event")
	}

	s.logger.Info("Queue rebalancing completed")
	return nil
}

// LoadQueueFromPersistence loads the queue from persistent storage
func (s *QueueServiceImpl) LoadQueueFromPersistence(ctx context.Context) error {
	s.logger.Info("Loading queue from persistence")

	persistedQueue, err := s.repoManager.Queue().LoadQueue(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to load queue from persistence")
		return fmt.Errorf("failed to load queue from persistence: %w", err)
	}

	s.queue = persistedQueue

	s.logger.WithField("queue_length", s.queue.GetQueueLength()).Info("Queue loaded from persistence successfully")
	return nil
}

// SaveQueueToPersistence saves the current queue to persistent storage
func (s *QueueServiceImpl) SaveQueueToPersistence(ctx context.Context) error {
	s.logger.Info("Saving queue to persistence")

	if err := s.repoManager.Queue().SaveQueue(ctx, s.queue); err != nil {
		s.logger.WithError(err).Error("Failed to save queue to persistence")
		return fmt.Errorf("failed to save queue to persistence: %w", err)
	}

	s.logger.WithField("queue_length", s.queue.GetQueueLength()).Info("Queue saved to persistence successfully")
	return nil
}

// GetQueueMetrics returns queue performance metrics
func (s *QueueServiceImpl) GetQueueMetrics(ctx context.Context, period *TimePeriod) (*QueueMetrics, error) {
	s.logger.WithField("period", period).Info("Retrieving queue metrics")

	var start, end time.Time
	if period != nil {
		start, end = period.Start, period.End
	} else {
		// Default to last 24 hours
		end = time.Now()
		start = end.AddDate(0, 0, -1)
	}

	// Get throughput stats
	throughputStats, err := s.repoManager.Queue().GetThroughputStats(ctx, start, end)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get throughput stats")
		throughputStats = &domain.ThroughputStats{}
	}

	// Get average wait time
	avgWaitTime, err := s.repoManager.Queue().GetAverageWaitTime(ctx, start, end)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get average wait time")
		avgWaitTime = 0
	}

	// Get current status
	currentStatus := s.queue.GetQueueStatus()

	metrics := &QueueMetrics{
		CurrentQueueLength: currentStatus.TotalOrders,
		PendingOrders:      currentStatus.PendingOrders,
		ProcessingOrders:   currentStatus.ProcessingOrders,
		CompletedOrders:    currentStatus.CompletedOrders,
		AverageWaitTime:    float32(avgWaitTime),
		ThroughputPerHour:  throughputStats.OrdersPerHour,
		ThroughputPerDay:   throughputStats.OrdersPerDay,
		PeakThroughput:     throughputStats.PeakHourThroughput,
		CalculatedAt:       time.Now(),
	}

	s.logger.WithFields(map[string]interface{}{
		"queue_length":    metrics.CurrentQueueLength,
		"avg_wait_time":   metrics.AverageWaitTime,
		"throughput_hour": metrics.ThroughputPerHour,
	}).Info("Queue metrics calculated")

	return metrics, nil
}

// QueueMetrics represents queue performance metrics
type QueueMetrics struct {
	CurrentQueueLength int32     `json:"current_queue_length"`
	PendingOrders      int32     `json:"pending_orders"`
	ProcessingOrders   int32     `json:"processing_orders"`
	CompletedOrders    int32     `json:"completed_orders"`
	AverageWaitTime    float32   `json:"average_wait_time"`
	ThroughputPerHour  float32   `json:"throughput_per_hour"`
	ThroughputPerDay   float32   `json:"throughput_per_day"`
	PeakThroughput     float32   `json:"peak_throughput"`
	CalculatedAt       time.Time `json:"calculated_at"`
}
