package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// EventIntegrationService handles cross-service event integration
type EventIntegrationService struct {
	redisClient         *redis.Client
	kitchenService      application.KitchenService
	orderServiceClient  *OrderServiceClient
	logger              *logger.Logger
	subscriptions       map[string]chan *IntegrationEvent
	stopChannels        map[string]chan struct{}
}

// IntegrationEvent represents an event for cross-service integration
type IntegrationEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	AggregateID string                 `json:"aggregate_id"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
}

// NewEventIntegrationService creates a new event integration service
func NewEventIntegrationService(
	redisClient *redis.Client,
	kitchenService application.KitchenService,
	orderServiceClient *OrderServiceClient,
	logger *logger.Logger,
) *EventIntegrationService {
	return &EventIntegrationService{
		redisClient:        redisClient,
		kitchenService:     kitchenService,
		orderServiceClient: orderServiceClient,
		logger:             logger,
		subscriptions:      make(map[string]chan *IntegrationEvent),
		stopChannels:       make(map[string]chan struct{}),
	}
}

// Start starts the event integration service
func (s *EventIntegrationService) Start(ctx context.Context) error {
	s.logger.Info("Starting event integration service")

	// Subscribe to order service events
	if err := s.subscribeToOrderEvents(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to order events: %w", err)
	}

	// Subscribe to kitchen events for outbound integration
	if err := s.subscribeToKitchenEvents(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to kitchen events: %w", err)
	}

	s.logger.Info("Event integration service started successfully")
	return nil
}

// Stop stops the event integration service
func (s *EventIntegrationService) Stop() {
	s.logger.Info("Stopping event integration service")

	// Stop all subscriptions
	for channel, stopChan := range s.stopChannels {
		close(stopChan)
		s.logger.WithField("channel", channel).Info("Stopped event subscription")
	}

	s.logger.Info("Event integration service stopped")
}

// subscribeToOrderEvents subscribes to events from the order service
func (s *EventIntegrationService) subscribeToOrderEvents(ctx context.Context) error {
	eventTypes := []string{
		"order.created",
		"order.updated",
		"order.cancelled",
		"order.payment_confirmed",
	}

	for _, eventType := range eventTypes {
		channel := fmt.Sprintf("events:order:%s", eventType)
		eventChan := make(chan *IntegrationEvent, 100)
		stopChan := make(chan struct{})

		s.subscriptions[channel] = eventChan
		s.stopChannels[channel] = stopChan

		// Start Redis subscription
		go s.subscribeToRedisChannel(ctx, channel, eventChan, stopChan)

		// Start event handler
		go s.handleOrderEvents(ctx, eventType, eventChan, stopChan)

		s.logger.WithField("event_type", eventType).Info("Subscribed to order events")
	}

	return nil
}

// subscribeToKitchenEvents subscribes to kitchen events for outbound integration
func (s *EventIntegrationService) subscribeToKitchenEvents(ctx context.Context) error {
	eventTypes := []string{
		"kitchen.order.status_changed",
		"kitchen.order.completed",
		"kitchen.order.overdue",
		"kitchen.queue.status_changed",
	}

	for _, eventType := range eventTypes {
		channel := fmt.Sprintf("events:kitchen:%s", eventType)
		eventChan := make(chan *IntegrationEvent, 100)
		stopChan := make(chan struct{})

		s.subscriptions[channel] = eventChan
		s.stopChannels[channel] = stopChan

		// Start Redis subscription
		go s.subscribeToRedisChannel(ctx, channel, eventChan, stopChan)

		// Start event handler
		go s.handleKitchenEvents(ctx, eventType, eventChan, stopChan)

		s.logger.WithField("event_type", eventType).Info("Subscribed to kitchen events")
	}

	return nil
}

// subscribeToRedisChannel subscribes to a Redis channel and forwards events
func (s *EventIntegrationService) subscribeToRedisChannel(ctx context.Context, channel string, eventChan chan *IntegrationEvent, stopChan chan struct{}) {
	pubsub := s.redisClient.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-stopChan:
			s.logger.WithField("channel", channel).Info("Stopping Redis subscription")
			return
		case msg := <-ch:
			var event IntegrationEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				s.logger.WithError(err).WithField("channel", channel).Error("Failed to unmarshal event")
				continue
			}

			select {
			case eventChan <- &event:
			default:
				s.logger.WithField("channel", channel).Warn("Event channel full, dropping event")
			}
		case <-ctx.Done():
			return
		}
	}
}

// handleOrderEvents handles events from the order service
func (s *EventIntegrationService) handleOrderEvents(ctx context.Context, eventType string, eventChan chan *IntegrationEvent, stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		case event := <-eventChan:
			if err := s.processOrderEvent(ctx, event); err != nil {
				s.logger.WithError(err).WithFields(map[string]interface{}{
					"event_type": eventType,
					"event_id":   event.ID,
				}).Error("Failed to process order event")
			}
		case <-ctx.Done():
			return
		}
	}
}

// handleKitchenEvents handles events from the kitchen service for outbound integration
func (s *EventIntegrationService) handleKitchenEvents(ctx context.Context, eventType string, eventChan chan *IntegrationEvent, stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		case event := <-eventChan:
			if err := s.processKitchenEvent(ctx, event); err != nil {
				s.logger.WithError(err).WithFields(map[string]interface{}{
					"event_type": eventType,
					"event_id":   event.ID,
				}).Error("Failed to process kitchen event")
			}
		case <-ctx.Done():
			return
		}
	}
}

// processOrderEvent processes an event from the order service
func (s *EventIntegrationService) processOrderEvent(ctx context.Context, event *IntegrationEvent) error {
	s.logger.WithFields(map[string]interface{}{
		"event_type":   event.Type,
		"event_id":     event.ID,
		"aggregate_id": event.AggregateID,
	}).Info("Processing order event")

	switch event.Type {
	case "order.created":
		return s.handleOrderCreated(ctx, event)
	case "order.updated":
		return s.handleOrderUpdated(ctx, event)
	case "order.cancelled":
		return s.handleOrderCancelled(ctx, event)
	case "order.payment_confirmed":
		return s.handleOrderPaymentConfirmed(ctx, event)
	default:
		s.logger.WithField("event_type", event.Type).Warn("Unknown order event type")
		return nil
	}
}

// processKitchenEvent processes an event from the kitchen service
func (s *EventIntegrationService) processKitchenEvent(ctx context.Context, event *IntegrationEvent) error {
	s.logger.WithFields(map[string]interface{}{
		"event_type":   event.Type,
		"event_id":     event.ID,
		"aggregate_id": event.AggregateID,
	}).Info("Processing kitchen event")

	switch event.Type {
	case "kitchen.order.status_changed":
		return s.handleKitchenOrderStatusChanged(ctx, event)
	case "kitchen.order.completed":
		return s.handleKitchenOrderCompleted(ctx, event)
	case "kitchen.order.overdue":
		return s.handleKitchenOrderOverdue(ctx, event)
	case "kitchen.queue.status_changed":
		return s.handleKitchenQueueStatusChanged(ctx, event)
	default:
		s.logger.WithField("event_type", event.Type).Warn("Unknown kitchen event type")
		return nil
	}
}

// Order event handlers

func (s *EventIntegrationService) handleOrderCreated(ctx context.Context, event *IntegrationEvent) error {
	orderID := event.AggregateID

	// Get order details from order service
	orderInfo, err := s.orderServiceClient.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order details: %w", err)
	}

	// Convert to kitchen order and add to queue
	kitchenOrderReq, err := s.convertToKitchenOrder(orderInfo)
	if err != nil {
		return fmt.Errorf("failed to convert order: %w", err)
	}

	_, err = s.kitchenService.AddOrderToQueue(ctx, kitchenOrderReq)
	if err != nil {
		return fmt.Errorf("failed to add order to kitchen queue: %w", err)
	}

	s.logger.WithField("order_id", orderID).Info("Order added to kitchen queue from order service")
	return nil
}

func (s *EventIntegrationService) handleOrderUpdated(ctx context.Context, event *IntegrationEvent) error {
	// Handle order updates if needed
	s.logger.WithField("order_id", event.AggregateID).Info("Order updated in order service")
	return nil
}

func (s *EventIntegrationService) handleOrderCancelled(ctx context.Context, event *IntegrationEvent) error {
	orderID := event.AggregateID

	// Update order status in kitchen
	err := s.kitchenService.UpdateOrderStatus(ctx, orderID, domain.OrderStatusCancelled)
	if err != nil {
		return fmt.Errorf("failed to cancel order in kitchen: %w", err)
	}

	s.logger.WithField("order_id", orderID).Info("Order cancelled in kitchen")
	return nil
}

func (s *EventIntegrationService) handleOrderPaymentConfirmed(ctx context.Context, event *IntegrationEvent) error {
	orderID := event.AggregateID

	// Start order processing in kitchen
	err := s.kitchenService.StartOrderProcessing(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to start order processing: %w", err)
	}

	s.logger.WithField("order_id", orderID).Info("Order processing started after payment confirmation")
	return nil
}

// Kitchen event handlers

func (s *EventIntegrationService) handleKitchenOrderStatusChanged(ctx context.Context, event *IntegrationEvent) error {
	orderID := event.AggregateID
	
	// Extract status from event data
	statusData, ok := event.Data["new_status"]
	if !ok {
		return fmt.Errorf("missing status in event data")
	}

	status, ok := statusData.(float64) // JSON numbers are float64
	if !ok {
		return fmt.Errorf("invalid status type in event data")
	}

	// Update order status in order service
	err := s.orderServiceClient.UpdateOrderStatus(ctx, orderID, domain.OrderStatus(status))
	if err != nil {
		return fmt.Errorf("failed to update order status in order service: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"order_id": orderID,
		"status":   status,
	}).Info("Order status updated in order service")

	return nil
}

func (s *EventIntegrationService) handleKitchenOrderCompleted(ctx context.Context, event *IntegrationEvent) error {
	orderID := event.AggregateID

	// Extract actual time from event data
	actualTimeData, ok := event.Data["actual_time"]
	if !ok {
		return fmt.Errorf("missing actual_time in event data")
	}

	actualTime, ok := actualTimeData.(float64)
	if !ok {
		return fmt.Errorf("invalid actual_time type in event data")
	}

	// Notify order service of completion
	err := s.orderServiceClient.NotifyOrderCompleted(ctx, orderID, int32(actualTime))
	if err != nil {
		return fmt.Errorf("failed to notify order completion: %w", err)
	}

	s.logger.WithField("order_id", orderID).Info("Order completion notified to order service")
	return nil
}

func (s *EventIntegrationService) handleKitchenOrderOverdue(ctx context.Context, event *IntegrationEvent) error {
	orderID := event.AggregateID

	// Log overdue order (could also notify external systems)
	s.logger.WithField("order_id", orderID).Warn("Order is overdue in kitchen")

	// Could implement additional logic like:
	// - Sending alerts to management
	// - Updating customer notifications
	// - Escalating priority

	return nil
}

func (s *EventIntegrationService) handleKitchenQueueStatusChanged(ctx context.Context, event *IntegrationEvent) error {
	// Handle queue status changes (could be used for capacity planning)
	s.logger.Info("Kitchen queue status changed")

	// Could implement logic like:
	// - Updating capacity predictions
	// - Adjusting order acceptance rates
	// - Triggering staff scheduling

	return nil
}

// Helper methods

func (s *EventIntegrationService) convertToKitchenOrder(orderInfo *OrderInfo) (*application.AddOrderRequest, error) {
	items := make([]*application.OrderItemRequest, len(orderInfo.Items))
	for i, item := range orderInfo.Items {
		// Map order items to kitchen requirements
		requirements := s.mapItemToStationRequirements(item.Name)
		
		items[i] = &application.OrderItemRequest{
			ID:           item.ID,
			Name:         item.Name,
			Quantity:     item.Quantity,
			Instructions: item.Instructions,
			Requirements: requirements,
			Metadata:     item.Metadata,
		}
	}

	return &application.AddOrderRequest{
		ID:         orderInfo.ID,
		CustomerID: orderInfo.CustomerID,
		Items:      items,
		Priority:   orderInfo.Priority,
	}, nil
}

func (s *EventIntegrationService) mapItemToStationRequirements(itemName string) []domain.StationType {
	// Simplified mapping - in production, this would be more sophisticated
	itemMappings := map[string][]domain.StationType{
		"espresso":    {domain.StationTypeEspresso, domain.StationTypeGrinder},
		"cappuccino":  {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer},
		"latte":       {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer},
		"americano":   {domain.StationTypeEspresso, domain.StationTypeGrinder},
		"macchiato":   {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer},
		"mocha":       {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer, domain.StationTypeAssembly},
		"frappuccino": {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer, domain.StationTypeAssembly},
	}

	// Check for exact matches first
	if requirements, exists := itemMappings[itemName]; exists {
		return requirements
	}

	// Default requirements for unknown items
	return []domain.StationType{domain.StationTypeAssembly}
}

// PublishEvent publishes an integration event to Redis
func (s *EventIntegrationService) PublishEvent(ctx context.Context, event *IntegrationEvent) error {
	channel := fmt.Sprintf("events:%s:%s", event.Source, event.Type)
	
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = s.redisClient.Publish(ctx, channel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"event_type": event.Type,
		"event_id":   event.ID,
		"channel":    channel,
	}).Info("Event published")

	return nil
}
