package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DimaJoyti/go-coffee/domain/shared"
)

// Event types for order domain
const (
	OrderCreatedEventType                    = "order.created"
	OrderStatusChangedEventType              = "order.status_changed"
	OrderPriorityChangedEventType            = "order.priority_changed"
	OrderItemAddedEventType                  = "order.item_added"
	OrderItemRemovedEventType                = "order.item_removed"
	OrderDiscountAppliedEventType            = "order.discount_applied"
	OrderEstimatedCompletionUpdatedEventType = "order.estimated_completion_updated"
	OrderAIInsightsUpdatedEventType          = "order.ai_insights_updated"
	OrderCompletedEventType                  = "order.completed"
	OrderCancelledEventType                  = "order.cancelled"
)

// OrderCreatedEvent is raised when a new order is created
type OrderCreatedEvent struct {
	*shared.BaseDomainEvent
}

// NewOrderCreatedEvent creates a new OrderCreatedEvent
func NewOrderCreatedEvent(order *Order) *OrderCreatedEvent {
	data := map[string]interface{}{
		"order_number":         order.OrderNumber(),
		"customer_id":          order.Customer().ID().Value(),
		"customer_name":        order.Customer().name,
		"customer_email":       order.Customer().email.Value(),
		"location_id":          order.LocationID().Value(),
		"status":               order.Status().String(),
		"priority":             order.Priority().String(),
		"total_amount":         order.TotalAmount().ToFloat(),
		"currency":             order.TotalAmount().Currency(),
		"estimated_completion": order.EstimatedCompletion(),
		"item_count":           len(order.Items()),
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		OrderCreatedEventType,
		order.ID(),
		order.GetTenantID(),
		data,
	)

	return &OrderCreatedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// OrderStatusChangedEvent is raised when an order status changes
type OrderStatusChangedEvent struct {
	*shared.BaseDomainEvent
}

// NewOrderStatusChangedEvent creates a new OrderStatusChangedEvent
func NewOrderStatusChangedEvent(order *Order, oldStatus, newStatus OrderStatus, reason string) *OrderStatusChangedEvent {
	data := map[string]interface{}{
		"order_number": order.OrderNumber(),
		"customer_id":  order.Customer().ID().Value(),
		"location_id":  order.LocationID().Value(),
		"old_status":   oldStatus.String(),
		"new_status":   newStatus.String(),
		"reason":       reason,
	}

	// Add completion time if order is completed
	if newStatus == OrderStatusCompleted && order.ActualCompletion() != nil {
		data["completed_at"] = *order.ActualCompletion()
		data["preparation_time_minutes"] = order.ActualCompletion().Sub(order.CreatedAt()).Minutes()
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		OrderStatusChangedEventType,
		order.ID(),
		order.GetTenantID(),
		data,
	)

	return &OrderStatusChangedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// OrderPriorityChangedEvent is raised when an order priority changes
type OrderPriorityChangedEvent struct {
	*shared.BaseDomainEvent
}

// NewOrderPriorityChangedEvent creates a new OrderPriorityChangedEvent
func NewOrderPriorityChangedEvent(order *Order, oldPriority, newPriority OrderPriority, reason string) *OrderPriorityChangedEvent {
	data := map[string]interface{}{
		"order_number": order.OrderNumber(),
		"customer_id":  order.Customer().ID().Value(),
		"location_id":  order.LocationID().Value(),
		"old_priority": oldPriority.String(),
		"new_priority": newPriority.String(),
		"reason":       reason,
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		OrderPriorityChangedEventType,
		order.ID(),
		order.GetTenantID(),
		data,
	)

	return &OrderPriorityChangedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// OrderItemAddedEvent is raised when an item is added to an order
type OrderItemAddedEvent struct {
	*shared.BaseDomainEvent
}

// NewOrderItemAddedEvent creates a new OrderItemAddedEvent
func NewOrderItemAddedEvent(order *Order, item *OrderItem) *OrderItemAddedEvent {
	data := map[string]interface{}{
		"order_number":   order.OrderNumber(),
		"customer_id":    order.Customer().ID().Value(),
		"location_id":    order.LocationID().Value(),
		"item_id":        item.ID().Value(),
		"product_id":     item.productID.Value(),
		"product_name":   item.productName,
		"quantity":       item.quantity,
		"unit_price":     item.unitPrice.ToFloat(),
		"total_price":    item.totalPrice.ToFloat(),
		"currency":       item.unitPrice.Currency(),
		"customizations": item.customizations,
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		OrderItemAddedEventType,
		order.ID(),
		order.GetTenantID(),
		data,
	)

	return &OrderItemAddedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// OrderItemRemovedEvent is raised when an item is removed from an order
type OrderItemRemovedEvent struct {
	*shared.BaseDomainEvent
}

// NewOrderItemRemovedEvent creates a new OrderItemRemovedEvent
func NewOrderItemRemovedEvent(order *Order, item *OrderItem) *OrderItemRemovedEvent {
	data := map[string]interface{}{
		"order_number": order.OrderNumber(),
		"customer_id":  order.Customer().ID().Value(),
		"location_id":  order.LocationID().Value(),
		"item_id":      item.ID().Value(),
		"product_id":   item.productID.Value(),
		"product_name": item.productName,
		"quantity":     item.quantity,
		"total_price":  item.totalPrice.ToFloat(),
		"currency":     item.unitPrice.Currency(),
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		OrderItemRemovedEventType,
		order.ID(),
		order.GetTenantID(),
		data,
	)

	return &OrderItemRemovedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// OrderDiscountAppliedEvent is raised when a discount is applied to an order
type OrderDiscountAppliedEvent struct {
	*shared.BaseDomainEvent
}

// NewOrderDiscountAppliedEvent creates a new OrderDiscountAppliedEvent
func NewOrderDiscountAppliedEvent(order *Order, discountAmount shared.Money, reason string) *OrderDiscountAppliedEvent {
	data := map[string]interface{}{
		"order_number":    order.OrderNumber(),
		"customer_id":     order.Customer().ID().Value(),
		"location_id":     order.LocationID().Value(),
		"discount_amount": discountAmount.ToFloat(),
		"currency":        discountAmount.Currency(),
		"reason":          reason,
		"total_amount":    order.TotalAmount().ToFloat(),
		"final_amount":    order.FinalAmount().ToFloat(),
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		OrderDiscountAppliedEventType,
		order.ID(),
		order.GetTenantID(),
		data,
	)

	return &OrderDiscountAppliedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// OrderEstimatedCompletionUpdatedEvent is raised when estimated completion time is updated
type OrderEstimatedCompletionUpdatedEvent struct {
	*shared.BaseDomainEvent
}

// NewOrderEstimatedCompletionUpdatedEvent creates a new OrderEstimatedCompletionUpdatedEvent
func NewOrderEstimatedCompletionUpdatedEvent(order *Order, oldTime, newTime time.Time) *OrderEstimatedCompletionUpdatedEvent {
	data := map[string]interface{}{
		"order_number":              order.OrderNumber(),
		"customer_id":               order.Customer().ID().Value(),
		"location_id":               order.LocationID().Value(),
		"old_estimated_completion":  oldTime,
		"new_estimated_completion":  newTime,
		"time_difference_minutes":   newTime.Sub(oldTime).Minutes(),
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		OrderEstimatedCompletionUpdatedEventType,
		order.ID(),
		order.GetTenantID(),
		data,
	)

	return &OrderEstimatedCompletionUpdatedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// OrderAIInsightsUpdatedEvent is raised when AI insights are updated
type OrderAIInsightsUpdatedEvent struct {
	*shared.BaseDomainEvent
}

// NewOrderAIInsightsUpdatedEvent creates a new OrderAIInsightsUpdatedEvent
func NewOrderAIInsightsUpdatedEvent(order *Order, insights *OrderAIInsights) *OrderAIInsightsUpdatedEvent {
	data := map[string]interface{}{
		"order_number": order.OrderNumber(),
		"customer_id":  order.Customer().ID().Value(),
		"location_id":  order.LocationID().Value(),
	}

	if insights != nil {
		data["complexity_score"] = insights.complexityScore.Value()
		data["estimated_prep_time_minutes"] = insights.estimatedPrepTime.Minutes()
		data["revenue_impact"] = insights.revenueImpact.ToFloat()
		data["satisfaction_prediction"] = insights.customerSatisfactionPrediction.Value()
		data["upsell_opportunities"] = insights.upsellOpportunities
		data["cross_sell_suggestions"] = insights.crossSellSuggestions
		data["risk_factors"] = insights.riskFactors
		data["optimization_suggestions"] = insights.optimizationSuggestions
	}

	baseEvent := shared.NewBaseDomainEvent(
		uuid.New().String(),
		OrderAIInsightsUpdatedEventType,
		order.ID(),
		order.GetTenantID(),
		data,
	)

	return &OrderAIInsightsUpdatedEvent{
		BaseDomainEvent: baseEvent,
	}
}

// OrderEventHandler handles order domain events
type OrderEventHandler struct {
	name string
}

// NewOrderEventHandler creates a new order event handler
func NewOrderEventHandler() *OrderEventHandler {
	return &OrderEventHandler{
		name: "OrderEventHandler",
	}
}

// HandlerName returns the handler name
func (h *OrderEventHandler) HandlerName() string {
	return h.name
}

// CanHandle checks if the handler can process the event type
func (h *OrderEventHandler) CanHandle(eventType string) bool {
	switch eventType {
	case OrderCreatedEventType,
		OrderStatusChangedEventType,
		OrderPriorityChangedEventType,
		OrderItemAddedEventType,
		OrderItemRemovedEventType,
		OrderDiscountAppliedEventType,
		OrderEstimatedCompletionUpdatedEventType,
		OrderAIInsightsUpdatedEventType:
		return true
	default:
		return false
	}
}

// Handle processes order domain events
func (h *OrderEventHandler) Handle(ctx context.Context, event shared.DomainEvent) error {
	switch event.EventType() {
	case OrderCreatedEventType:
		return h.handleOrderCreated(ctx, event)
	case OrderStatusChangedEventType:
		return h.handleOrderStatusChanged(ctx, event)
	case OrderPriorityChangedEventType:
		return h.handleOrderPriorityChanged(ctx, event)
	case OrderItemAddedEventType:
		return h.handleOrderItemAdded(ctx, event)
	case OrderItemRemovedEventType:
		return h.handleOrderItemRemoved(ctx, event)
	case OrderDiscountAppliedEventType:
		return h.handleOrderDiscountApplied(ctx, event)
	case OrderEstimatedCompletionUpdatedEventType:
		return h.handleOrderEstimatedCompletionUpdated(ctx, event)
	case OrderAIInsightsUpdatedEventType:
		return h.handleOrderAIInsightsUpdated(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType())
	}
}

func (h *OrderEventHandler) handleOrderCreated(ctx context.Context, event shared.DomainEvent) error {
	// Handle order creation logic (e.g., send to kitchen, notify customer)
	return nil
}

func (h *OrderEventHandler) handleOrderStatusChanged(ctx context.Context, event shared.DomainEvent) error {
	// Handle status change logic (e.g., update kitchen queue, notify customer)
	return nil
}

func (h *OrderEventHandler) handleOrderPriorityChanged(ctx context.Context, event shared.DomainEvent) error {
	// Handle priority change logic (e.g., reorder kitchen queue)
	return nil
}

func (h *OrderEventHandler) handleOrderItemAdded(ctx context.Context, event shared.DomainEvent) error {
	// Handle item addition logic (e.g., update inventory, recalculate prep time)
	return nil
}

func (h *OrderEventHandler) handleOrderItemRemoved(ctx context.Context, event shared.DomainEvent) error {
	// Handle item removal logic (e.g., update inventory, recalculate prep time)
	return nil
}

func (h *OrderEventHandler) handleOrderDiscountApplied(ctx context.Context, event shared.DomainEvent) error {
	// Handle discount logic (e.g., update billing, track promotions)
	return nil
}

func (h *OrderEventHandler) handleOrderEstimatedCompletionUpdated(ctx context.Context, event shared.DomainEvent) error {
	// Handle completion time update logic (e.g., notify customer, update displays)
	return nil
}

func (h *OrderEventHandler) handleOrderAIInsightsUpdated(ctx context.Context, event shared.DomainEvent) error {
	// Handle AI insights update logic (e.g., trigger recommendations, update analytics)
	return nil
}
