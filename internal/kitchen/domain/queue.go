package domain

import (
	"errors"
	"sort"
	"time"
)

// QueueStatus represents the current status of the order queue
type QueueStatus struct {
	TotalOrders      int32                    `json:"total_orders"`
	ProcessingOrders int32                    `json:"processing_orders"`
	PendingOrders    int32                    `json:"pending_orders"`
	CompletedOrders  int32                    `json:"completed_orders"`
	AverageWaitTime  int32                    `json:"average_wait_time"`
	QueuesByPriority map[OrderPriority]int32  `json:"queues_by_priority"`
	StationLoad      map[StationType]float32  `json:"station_load"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

// OrderQueue represents the kitchen order queue (Domain Entity)
type OrderQueue struct {
	orders    []*KitchenOrder
	updatedAt time.Time
}

// NewOrderQueue creates a new order queue
func NewOrderQueue() *OrderQueue {
	return &OrderQueue{
		orders:    make([]*KitchenOrder, 0),
		updatedAt: time.Now(),
	}
}

// AddOrder adds an order to the queue with priority-based positioning
func (q *OrderQueue) AddOrder(order *KitchenOrder) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}
	if order.Status() != OrderStatusPending {
		return errors.New("only pending orders can be added to queue")
	}

	// Check if order already exists
	for _, existingOrder := range q.orders {
		if existingOrder.ID() == order.ID() {
			return errors.New("order already exists in queue")
		}
	}

	q.orders = append(q.orders, order)
	q.sortByPriority()
	q.updatedAt = time.Now()
	return nil
}

// RemoveOrder removes an order from the queue
func (q *OrderQueue) RemoveOrder(orderID string) error {
	for i, order := range q.orders {
		if order.ID() == orderID {
			q.orders = append(q.orders[:i], q.orders[i+1:]...)
			q.updatedAt = time.Now()
			return nil
		}
	}
	return errors.New("order not found in queue")
}

// GetNextOrder returns the next order to be processed (highest priority, oldest first)
func (q *OrderQueue) GetNextOrder() *KitchenOrder {
	if len(q.orders) == 0 {
		return nil
	}

	// Find the first pending order (queue is already sorted by priority)
	for _, order := range q.orders {
		if order.Status() == OrderStatusPending && order.IsReadyToStart() {
			return order
		}
	}

	return nil
}

// GetOrder retrieves an order by ID
func (q *OrderQueue) GetOrder(orderID string) *KitchenOrder {
	for _, order := range q.orders {
		if order.ID() == orderID {
			return order
		}
	}
	return nil
}

// GetAllOrders returns all orders in the queue
func (q *OrderQueue) GetAllOrders() []*KitchenOrder {
	// Return a copy to prevent external modification
	orders := make([]*KitchenOrder, len(q.orders))
	copy(orders, q.orders)
	return orders
}

// GetOrdersByStatus returns orders filtered by status
func (q *OrderQueue) GetOrdersByStatus(status OrderStatus) []*KitchenOrder {
	var filtered []*KitchenOrder
	for _, order := range q.orders {
		if order.Status() == status {
			filtered = append(filtered, order)
		}
	}
	return filtered
}

// GetOrdersByPriority returns orders filtered by priority
func (q *OrderQueue) GetOrdersByPriority(priority OrderPriority) []*KitchenOrder {
	var filtered []*KitchenOrder
	for _, order := range q.orders {
		if order.Priority() == priority {
			filtered = append(filtered, order)
		}
	}
	return filtered
}

// UpdateOrderPriority updates the priority of an order and re-sorts the queue
func (q *OrderQueue) UpdateOrderPriority(orderID string, priority OrderPriority) error {
	order := q.GetOrder(orderID)
	if order == nil {
		return errors.New("order not found in queue")
	}

	order.SetPriority(priority)
	q.sortByPriority()
	q.updatedAt = time.Now()
	return nil
}

// sortByPriority sorts orders by priority (highest first) and then by creation time (oldest first)
func (q *OrderQueue) sortByPriority() {
	sort.Slice(q.orders, func(i, j int) bool {
		orderI, orderJ := q.orders[i], q.orders[j]
		
		// First sort by priority (higher priority first)
		if orderI.Priority() != orderJ.Priority() {
			return orderI.Priority() > orderJ.Priority()
		}
		
		// If same priority, sort by creation time (older first)
		return orderI.CreatedAt().Before(orderJ.CreatedAt())
	})
}

// GetQueueStatus returns the current status of the queue
func (q *OrderQueue) GetQueueStatus() *QueueStatus {
	status := &QueueStatus{
		QueuesByPriority: make(map[OrderPriority]int32),
		StationLoad:      make(map[StationType]float32),
		UpdatedAt:        time.Now(),
	}

	var totalWaitTime time.Duration
	var waitingOrders int32

	// Count orders by status and priority
	for _, order := range q.orders {
		status.TotalOrders++
		
		switch order.Status() {
		case OrderStatusPending:
			status.PendingOrders++
			status.QueuesByPriority[order.Priority()]++
			
			// Calculate wait time for pending orders
			waitTime := order.GetWaitTime()
			totalWaitTime += waitTime
			waitingOrders++
			
		case OrderStatusProcessing:
			status.ProcessingOrders++
			
		case OrderStatusCompleted:
			status.CompletedOrders++
		}

		// Calculate station load
		if order.Status() == OrderStatusPending || order.Status() == OrderStatusProcessing {
			requiredStations := order.GetRequiredStations()
			for _, station := range requiredStations {
				status.StationLoad[station] += float32(order.GetTotalQuantity())
			}
		}
	}

	// Calculate average wait time
	if waitingOrders > 0 {
		status.AverageWaitTime = int32(totalWaitTime.Seconds()) / waitingOrders
	}

	return status
}

// GetEstimatedWaitTime calculates estimated wait time for a new order
func (q *OrderQueue) GetEstimatedWaitTime(newOrder *KitchenOrder) time.Duration {
	var totalEstimatedTime time.Duration

	// Calculate time for all orders with higher or equal priority that are ahead
	for _, order := range q.orders {
		if order.Priority() >= newOrder.Priority() && 
		   (order.Status() == OrderStatusPending || order.Status() == OrderStatusProcessing) {
			
			if order.EstimatedTime() > 0 {
				totalEstimatedTime += time.Duration(order.EstimatedTime()) * time.Second
			} else {
				// Default estimation if no time set
				totalEstimatedTime += 5 * time.Minute
			}
		}
	}

	return totalEstimatedTime
}

// GetOverdueOrders returns orders that are taking longer than estimated
func (q *OrderQueue) GetOverdueOrders() []*KitchenOrder {
	var overdue []*KitchenOrder
	for _, order := range q.orders {
		if order.IsOverdue() {
			overdue = append(overdue, order)
		}
	}
	return overdue
}

// GetOrdersRequiringStation returns orders that require a specific station type
func (q *OrderQueue) GetOrdersRequiringStation(stationType StationType) []*KitchenOrder {
	var filtered []*KitchenOrder
	for _, order := range q.orders {
		requiredStations := order.GetRequiredStations()
		for _, required := range requiredStations {
			if required == stationType {
				filtered = append(filtered, order)
				break
			}
		}
	}
	return filtered
}

// GetQueueLength returns the number of orders in the queue
func (q *OrderQueue) GetQueueLength() int {
	return len(q.orders)
}

// GetPendingOrdersCount returns the number of pending orders
func (q *OrderQueue) GetPendingOrdersCount() int {
	count := 0
	for _, order := range q.orders {
		if order.Status() == OrderStatusPending {
			count++
		}
	}
	return count
}

// GetProcessingOrdersCount returns the number of processing orders
func (q *OrderQueue) GetProcessingOrdersCount() int {
	count := 0
	for _, order := range q.orders {
		if order.Status() == OrderStatusProcessing {
			count++
		}
	}
	return count
}

// IsEmpty checks if the queue is empty
func (q *OrderQueue) IsEmpty() bool {
	return len(q.orders) == 0
}

// Clear removes all orders from the queue
func (q *OrderQueue) Clear() {
	q.orders = make([]*KitchenOrder, 0)
	q.updatedAt = time.Now()
}

// GetUpdatedAt returns when the queue was last updated
func (q *OrderQueue) GetUpdatedAt() time.Time {
	return q.updatedAt
}

// WorkflowOptimization represents an optimized workflow
type WorkflowOptimization struct {
	OrderID             string                    `json:"order_id"`
	OptimizedSteps      []*WorkflowStep          `json:"optimized_steps"`
	EstimatedTime       int32                    `json:"estimated_time"`
	EfficiencyGain      float32                  `json:"efficiency_gain"`
	ResourceUtilization map[string]float32       `json:"resource_utilization"`
	Recommendations     []string                 `json:"recommendations"`
	CreatedAt           time.Time                `json:"created_at"`
}

// WorkflowStep represents a step in the optimized workflow
type WorkflowStep struct {
	StepID          string      `json:"step_id"`
	StationType     StationType `json:"station_type"`
	EstimatedTime   int32       `json:"estimated_time"`
	RequiredSkill   float32     `json:"required_skill"`
	Dependencies    []string    `json:"dependencies"`
	CanParallelize  bool        `json:"can_parallelize"`
	EquipmentID     string      `json:"equipment_id,omitempty"`
	StaffID         string      `json:"staff_id,omitempty"`
}

// NewWorkflowOptimization creates a new workflow optimization
func NewWorkflowOptimization(orderID string) *WorkflowOptimization {
	return &WorkflowOptimization{
		OrderID:             orderID,
		OptimizedSteps:      make([]*WorkflowStep, 0),
		ResourceUtilization: make(map[string]float32),
		Recommendations:     make([]string, 0),
		CreatedAt:           time.Now(),
	}
}

// AddStep adds a workflow step
func (wo *WorkflowOptimization) AddStep(step *WorkflowStep) {
	wo.OptimizedSteps = append(wo.OptimizedSteps, step)
}

// CalculateEstimatedTime calculates total estimated time for the workflow
func (wo *WorkflowOptimization) CalculateEstimatedTime() {
	var totalTime int32
	for _, step := range wo.OptimizedSteps {
		totalTime += step.EstimatedTime
	}
	wo.EstimatedTime = totalTime
}

// AddRecommendation adds a recommendation to the workflow
func (wo *WorkflowOptimization) AddRecommendation(recommendation string) {
	wo.Recommendations = append(wo.Recommendations, recommendation)
}
