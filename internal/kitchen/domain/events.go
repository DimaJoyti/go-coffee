package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event
type DomainEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	AggregateID string                 `json:"aggregate_id"`
	Data        map[string]interface{} `json:"data"`
	OccurredAt  time.Time              `json:"occurred_at"`
	Version     int                    `json:"version"`
}

// NewDomainEvent creates a new domain event
func NewDomainEvent(eventType, aggregateID string, data map[string]interface{}) *DomainEvent {
	return &DomainEvent{
		ID:          uuid.New().String(),
		Type:        eventType,
		AggregateID: aggregateID,
		Data:        data,
		OccurredAt:  time.Now(),
		Version:     1,
	}
}

// ToJSON converts the event to JSON
func (e *DomainEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Event Types
const (
	// Order Events
	OrderAddedToQueueEvent    = "kitchen.order.added_to_queue"
	OrderStatusChangedEvent   = "kitchen.order.status_changed"
	OrderAssignedEvent        = "kitchen.order.assigned"
	OrderStartedEvent         = "kitchen.order.started"
	OrderCompletedEvent       = "kitchen.order.completed"
	OrderCancelledEvent       = "kitchen.order.cancelled"
	OrderOverdueEvent         = "kitchen.order.overdue"
	OrderPriorityChangedEvent = "kitchen.order.priority_changed"

	// Equipment Events
	EquipmentStatusChangedEvent     = "kitchen.equipment.status_changed"
	EquipmentMaintenanceScheduled   = "kitchen.equipment.maintenance_scheduled"
	EquipmentMaintenanceCompleted   = "kitchen.equipment.maintenance_completed"
	EquipmentEfficiencyUpdatedEvent = "kitchen.equipment.efficiency_updated"
	EquipmentOverloadedEvent        = "kitchen.equipment.overloaded"

	// Staff Events
	StaffAssignedEvent            = "kitchen.staff.assigned"
	StaffAvailabilityChangedEvent = "kitchen.staff.availability_changed"
	StaffOverloadedEvent          = "kitchen.staff.overloaded"
	StaffSkillUpdatedEvent        = "kitchen.staff.skill_updated"

	// Queue Events
	QueueStatusChangedEvent = "kitchen.queue.status_changed"
	QueueOptimizedEvent     = "kitchen.queue.optimized"

	// Workflow Events
	WorkflowOptimizedEvent = "kitchen.workflow.optimized"
	WorkflowStepCompleted  = "kitchen.workflow.step_completed"
)

// Order Events

// NewOrderAddedToQueueEvent creates an event when an order is added to the queue
func NewOrderAddedToQueueEvent(order *KitchenOrder) *DomainEvent {
	data := map[string]interface{}{
		"order_id":    order.ID(),
		"customer_id": order.CustomerID(),
		"priority":    order.Priority(),
		"items_count": len(order.Items()),
		"created_at":  order.CreatedAt(),
	}
	return NewDomainEvent(OrderAddedToQueueEvent, order.ID(), data)
}

// NewOrderStatusChangedEvent creates an event when an order status changes
func NewOrderStatusChangedEvent(order *KitchenOrder, oldStatus OrderStatus) *DomainEvent {
	data := map[string]interface{}{
		"order_id":   order.ID(),
		"old_status": oldStatus,
		"new_status": order.Status(),
		"changed_at": order.UpdatedAt(),
	}
	return NewDomainEvent(OrderStatusChangedEvent, order.ID(), data)
}

// NewOrderAssignedEvent creates an event when an order is assigned to staff/equipment
func NewOrderAssignedEvent(order *KitchenOrder) *DomainEvent {
	data := map[string]interface{}{
		"order_id":           order.ID(),
		"assigned_staff_id":  order.AssignedStaffID(),
		"assigned_equipment": order.AssignedEquipment(),
		"assigned_at":        order.UpdatedAt(),
	}
	return NewDomainEvent(OrderAssignedEvent, order.ID(), data)
}

// NewOrderStartedEvent creates an event when an order starts processing
func NewOrderStartedEvent(order *KitchenOrder) *DomainEvent {
	data := map[string]interface{}{
		"order_id":       order.ID(),
		"started_at":     order.StartedAt(),
		"estimated_time": order.EstimatedTime(),
		"assigned_staff": order.AssignedStaffID(),
	}
	return NewDomainEvent(OrderStartedEvent, order.ID(), data)
}

// NewOrderCompletedEvent creates an event when an order is completed
func NewOrderCompletedEvent(order *KitchenOrder) *DomainEvent {
	data := map[string]interface{}{
		"order_id":       order.ID(),
		"completed_at":   order.CompletedAt(),
		"actual_time":    order.ActualTime(),
		"estimated_time": order.EstimatedTime(),
	}
	return NewDomainEvent(OrderCompletedEvent, order.ID(), data)
}

// NewOrderOverdueEvent creates an event when an order becomes overdue
func NewOrderOverdueEvent(order *KitchenOrder) *DomainEvent {
	data := map[string]interface{}{
		"order_id":        order.ID(),
		"estimated_time":  order.EstimatedTime(),
		"processing_time": order.GetProcessingTime().Seconds(),
		"overdue_by":      order.GetProcessingTime().Seconds() - float64(order.EstimatedTime()),
	}
	return NewDomainEvent(OrderOverdueEvent, order.ID(), data)
}

// NewOrderPriorityChangedEvent creates an event when order priority changes
func NewOrderPriorityChangedEvent(order *KitchenOrder) *DomainEvent {
	data := map[string]interface{}{
		"order_id":     order.ID(),
		"new_priority": order.Priority(),
		"changed_at":   order.UpdatedAt(),
	}
	return NewDomainEvent(OrderPriorityChangedEvent, order.ID(), data)
}

// Equipment Events

// NewEquipmentStatusChangedEvent creates an event when equipment status changes
func NewEquipmentStatusChangedEvent(equipment *Equipment, oldStatus EquipmentStatus) *DomainEvent {
	data := map[string]interface{}{
		"equipment_id": equipment.ID(),
		"old_status":   oldStatus,
		"new_status":   equipment.Status(),
		"changed_at":   equipment.UpdatedAt(),
	}
	return NewDomainEvent(EquipmentStatusChangedEvent, equipment.ID(), data)
}

// NewEquipmentMaintenanceScheduledEvent creates an event when maintenance is scheduled
func NewEquipmentMaintenanceScheduledEvent(equipment *Equipment) *DomainEvent {
	data := map[string]interface{}{
		"equipment_id":     equipment.ID(),
		"scheduled_at":     equipment.UpdatedAt(),
		"last_maintenance": equipment.LastMaintenance(),
	}
	return NewDomainEvent(EquipmentMaintenanceScheduled, equipment.ID(), data)
}

// NewEquipmentOverloadedEvent creates an event when equipment becomes overloaded
func NewEquipmentOverloadedEvent(equipment *Equipment) *DomainEvent {
	data := map[string]interface{}{
		"equipment_id":     equipment.ID(),
		"current_load":     equipment.CurrentLoad(),
		"max_capacity":     equipment.MaxCapacity(),
		"utilization_rate": equipment.GetUtilizationRate(),
	}
	return NewDomainEvent(EquipmentOverloadedEvent, equipment.ID(), data)
}

// Staff Events

// NewStaffAssignedEvent creates an event when staff is assigned to an order
func NewStaffAssignedEvent(staff *Staff, orderID string) *DomainEvent {
	data := map[string]interface{}{
		"staff_id":       staff.ID(),
		"order_id":       orderID,
		"assigned_at":    time.Now(),
		"current_orders": staff.CurrentOrders(),
		"workload":       staff.GetWorkload(),
	}
	return NewDomainEvent(StaffAssignedEvent, staff.ID(), data)
}

// NewStaffAvailabilityChangedEvent creates an event when staff availability changes
func NewStaffAvailabilityChangedEvent(staff *Staff, oldAvailability bool) *DomainEvent {
	data := map[string]interface{}{
		"staff_id":         staff.ID(),
		"old_availability": oldAvailability,
		"new_availability": staff.IsAvailable(),
		"changed_at":       staff.UpdatedAt(),
	}
	return NewDomainEvent(StaffAvailabilityChangedEvent, staff.ID(), data)
}

// NewStaffOverloadedEvent creates an event when staff becomes overloaded
func NewStaffOverloadedEvent(staff *Staff) *DomainEvent {
	data := map[string]interface{}{
		"staff_id":       staff.ID(),
		"current_orders": staff.CurrentOrders(),
		"max_orders":     staff.MaxConcurrentOrders(),
		"workload":       staff.GetWorkload(),
	}
	return NewDomainEvent(StaffOverloadedEvent, staff.ID(), data)
}

// NewStaffSkillUpdatedEvent creates an event when staff skill is updated
func NewStaffSkillUpdatedEvent(staff *Staff) *DomainEvent {
	data := map[string]interface{}{
		"staff_id":    staff.ID(),
		"skill_level": staff.SkillLevel(),
		"updated_at":  staff.UpdatedAt(),
	}
	return NewDomainEvent(StaffSkillUpdatedEvent, staff.ID(), data)
}

// Queue Events

// NewQueueStatusChangedEvent creates an event when queue status changes
func NewQueueStatusChangedEvent(status *QueueStatus) *DomainEvent {
	data := map[string]interface{}{
		"total_orders":      status.TotalOrders,
		"pending_orders":    status.PendingOrders,
		"processing_orders": status.ProcessingOrders,
		"completed_orders":  status.CompletedOrders,
		"average_wait_time": status.AverageWaitTime,
		"updated_at":        status.UpdatedAt,
	}
	return NewDomainEvent(QueueStatusChangedEvent, "kitchen_queue", data)
}

// Workflow Events

// NewWorkflowOptimizedEvent creates an event when a workflow is optimized
func NewWorkflowOptimizedEvent(optimization *WorkflowOptimization) *DomainEvent {
	data := map[string]interface{}{
		"order_id":        optimization.OrderID,
		"estimated_time":  optimization.EstimatedTime,
		"efficiency_gain": optimization.EfficiencyGain,
		"steps_count":     len(optimization.OptimizedSteps),
		"optimized_at":    optimization.CreatedAt,
	}
	return NewDomainEvent(WorkflowOptimizedEvent, optimization.OrderID, data)
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(event *DomainEvent) error
	PublishBatch(events []*DomainEvent) error
}

// EventSubscriber defines the interface for subscribing to domain events
type EventSubscriber interface {
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string) error
}

// EventHandler defines the interface for handling domain events
type EventHandler interface {
	Handle(event *DomainEvent) error
}

// EventHandlerFunc is a function type that implements EventHandler
type EventHandlerFunc func(event *DomainEvent) error

// Handle implements EventHandler interface
func (f EventHandlerFunc) Handle(event *DomainEvent) error {
	return f(event)
}

// EventStore defines the interface for storing domain events
type EventStore interface {
	Store(event *DomainEvent) error
	GetEvents(aggregateID string) ([]*DomainEvent, error)
	GetEventsByType(eventType string) ([]*DomainEvent, error)
	GetEventsAfter(timestamp time.Time) ([]*DomainEvent, error)
}
