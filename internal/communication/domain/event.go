package domain

import (
	"encoding/json"
	"errors"
	"time"
)

// EventType represents different types of events
type EventType string

const (
	// System events
	EventTypeServiceStarted   EventType = "service.started"
	EventTypeServiceStopped   EventType = "service.stopped"
	EventTypeServiceHealthy   EventType = "service.healthy"
	EventTypeServiceUnhealthy EventType = "service.unhealthy"
	
	// Business events
	EventTypeOrderPlaced     EventType = "order.placed"
	EventTypeOrderConfirmed  EventType = "order.confirmed"
	EventTypeOrderPreparing  EventType = "order.preparing"
	EventTypeOrderReady      EventType = "order.ready"
	EventTypeOrderCompleted  EventType = "order.completed"
	EventTypeOrderCancelled  EventType = "order.cancelled"
	
	EventTypePaymentInitiated EventType = "payment.initiated"
	EventTypePaymentCompleted EventType = "payment.completed"
	EventTypePaymentFailed    EventType = "payment.failed"
	EventTypePaymentRefunded  EventType = "payment.refunded"
	
	// User events
	EventTypeUserRegistered EventType = "user.registered"
	EventTypeUserLoggedIn   EventType = "user.logged_in"
	EventTypeUserLoggedOut  EventType = "user.logged_out"
	
	// Kitchen events
	EventTypeKitchenOrderReceived EventType = "kitchen.order_received"
	EventTypeKitchenOrderStarted  EventType = "kitchen.order_started"
	EventTypeKitchenOrderFinished EventType = "kitchen.order_finished"
	
	// Notification events
	EventTypeNotificationSent     EventType = "notification.sent"
	EventTypeNotificationDelivered EventType = "notification.delivered"
	EventTypeNotificationFailed   EventType = "notification.failed"
)

// EventStatus represents the status of an event
type EventStatus int32

const (
	EventStatusPending   EventStatus = 0
	EventStatusProcessed EventStatus = 1
	EventStatusFailed    EventStatus = 2
	EventStatusRetrying  EventStatus = 3
	EventStatusExpired   EventStatus = 4
)

// Event represents a domain event in the communication system
type Event struct {
	ID            string                 `json:"id"`
	Type          EventType              `json:"type"`
	Source        string                 `json:"source"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int64                  `json:"version"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]string      `json:"metadata"`
	Status        EventStatus            `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
	ProcessedAt   *time.Time             `json:"processed_at,omitempty"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
	RetryCount    int32                  `json:"retry_count"`
	MaxRetries    int32                  `json:"max_retries"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, source, aggregateID, aggregateType string, data map[string]interface{}) (*Event, error) {
	if eventType == "" {
		return nil, errors.New("event type is required")
	}
	
	if source == "" {
		return nil, errors.New("source is required")
	}
	
	if aggregateID == "" {
		return nil, errors.New("aggregate ID is required")
	}

	return &Event{
		ID:            generateEventID(),
		Type:          eventType,
		Source:        source,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Version:       1,
		Data:          data,
		Metadata:      make(map[string]string),
		Status:        EventStatusPending,
		CreatedAt:     time.Now(),
		RetryCount:    0,
		MaxRetries:    3,
	}, nil
}

// SetVersion sets the event version
func (e *Event) SetVersion(version int64) {
	e.Version = version
}

// SetExpiration sets the event expiration time
func (e *Event) SetExpiration(duration time.Duration) {
	expiresAt := e.CreatedAt.Add(duration)
	e.ExpiresAt = &expiresAt
}

// SetCorrelationID sets the correlation ID
func (e *Event) SetCorrelationID(correlationID string) {
	e.CorrelationID = correlationID
}

// AddMetadata adds metadata to the event
func (e *Event) AddMetadata(key, value string) {
	e.Metadata[key] = value
}

// MarkAsProcessed marks the event as processed
func (e *Event) MarkAsProcessed() {
	e.Status = EventStatusProcessed
	now := time.Now()
	e.ProcessedAt = &now
}

// MarkAsFailed marks the event as failed
func (e *Event) MarkAsFailed(errorMessage string) {
	e.Status = EventStatusFailed
	e.ErrorMessage = errorMessage
}

// MarkAsRetrying marks the event as retrying
func (e *Event) MarkAsRetrying() {
	e.Status = EventStatusRetrying
	e.RetryCount++
}

// IsExpired checks if the event has expired
func (e *Event) IsExpired() bool {
	if e.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*e.ExpiresAt)
}

// CanRetry checks if the event can be retried
func (e *Event) CanRetry() bool {
	return e.RetryCount < e.MaxRetries && !e.IsExpired()
}

// ToJSON converts the event to JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an event from JSON
func EventFromJSON(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	return &event, err
}

// GetDataString gets a string value from the data
func (e *Event) GetDataString(key string) (string, bool) {
	if value, exists := e.Data[key]; exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetDataInt gets an int value from the data
func (e *Event) GetDataInt(key string) (int64, bool) {
	if value, exists := e.Data[key]; exists {
		switch v := value.(type) {
		case int64:
			return v, true
		case int:
			return int64(v), true
		case float64:
			return int64(v), true
		}
	}
	return 0, false
}

// GetDataBool gets a boolean value from the data
func (e *Event) GetDataBool(key string) (bool, bool) {
	if value, exists := e.Data[key]; exists {
		if b, ok := value.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// EventBuilder provides a fluent interface for building events
type EventBuilder struct {
	event *Event
}

// NewEventBuilder creates a new event builder
func NewEventBuilder(eventType EventType, source, aggregateID, aggregateType string) *EventBuilder {
	event, _ := NewEvent(eventType, source, aggregateID, aggregateType, make(map[string]interface{}))
	return &EventBuilder{event: event}
}

// WithData sets the event data
func (b *EventBuilder) WithData(data map[string]interface{}) *EventBuilder {
	b.event.Data = data
	return b
}

// WithVersion sets the event version
func (b *EventBuilder) WithVersion(version int64) *EventBuilder {
	b.event.SetVersion(version)
	return b
}

// WithExpiration sets the expiration
func (b *EventBuilder) WithExpiration(duration time.Duration) *EventBuilder {
	b.event.SetExpiration(duration)
	return b
}

// WithCorrelationID sets the correlation ID
func (b *EventBuilder) WithCorrelationID(correlationID string) *EventBuilder {
	b.event.SetCorrelationID(correlationID)
	return b
}

// WithMetadata adds metadata
func (b *EventBuilder) WithMetadata(key, value string) *EventBuilder {
	b.event.AddMetadata(key, value)
	return b
}

// WithMaxRetries sets the maximum retry count
func (b *EventBuilder) WithMaxRetries(maxRetries int32) *EventBuilder {
	b.event.MaxRetries = maxRetries
	return b
}

// Build returns the built event
func (b *EventBuilder) Build() *Event {
	return b.event
}

// Helper functions

// generateEventID generates a unique event ID
func generateEventID() string {
	return "evt_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// Event Factory Functions

// NewOrderPlacedEvent creates an order placed event
func NewOrderPlacedEvent(source, orderID, customerID string, totalAmount int64) *Event {
	data := map[string]interface{}{
		"order_id":     orderID,
		"customer_id":  customerID,
		"total_amount": totalAmount,
		"timestamp":    time.Now(),
	}

	event, _ := NewEvent(EventTypeOrderPlaced, source, orderID, "order", data)
	event.AddMetadata("service", source)
	event.AddMetadata("event_category", "business")

	return event
}

// NewPaymentCompletedEvent creates a payment completed event
func NewPaymentCompletedEvent(source, paymentID, orderID string, amount int64) *Event {
	data := map[string]interface{}{
		"payment_id": paymentID,
		"order_id":   orderID,
		"amount":     amount,
		"timestamp":  time.Now(),
	}

	event, _ := NewEvent(EventTypePaymentCompleted, source, paymentID, "payment", data)
	event.AddMetadata("service", source)
	event.AddMetadata("event_category", "business")

	return event
}

// NewUserRegisteredEvent creates a user registered event
func NewUserRegisteredEvent(source, userID, email string) *Event {
	data := map[string]interface{}{
		"user_id":   userID,
		"email":     email,
		"timestamp": time.Now(),
	}

	event, _ := NewEvent(EventTypeUserRegistered, source, userID, "user", data)
	event.AddMetadata("service", source)
	event.AddMetadata("event_category", "user")

	return event
}

// NewServiceHealthEvent creates a service health event
func NewServiceHealthEvent(source string, isHealthy bool, details map[string]interface{}) *Event {
	eventType := EventTypeServiceHealthy
	if !isHealthy {
		eventType = EventTypeServiceUnhealthy
	}

	data := map[string]interface{}{
		"is_healthy": isHealthy,
		"details":    details,
		"timestamp":  time.Now(),
	}

	event, _ := NewEvent(eventType, source, source, "service", data)
	event.AddMetadata("service", source)
	event.AddMetadata("event_category", "system")

	return event
}

// NewKitchenOrderEvent creates a kitchen order event
func NewKitchenOrderEvent(source, orderID string, eventType EventType, estimatedTime int32) *Event {
	data := map[string]interface{}{
		"order_id":       orderID,
		"estimated_time": estimatedTime,
		"timestamp":      time.Now(),
	}

	event, _ := NewEvent(eventType, source, orderID, "kitchen_order", data)
	event.AddMetadata("service", source)
	event.AddMetadata("event_category", "kitchen")

	return event
}

// NewNotificationEvent creates a notification event
func NewNotificationEvent(source, notificationID, userID string, eventType EventType, details map[string]interface{}) *Event {
	data := map[string]interface{}{
		"notification_id": notificationID,
		"user_id":         userID,
		"details":         details,
		"timestamp":       time.Now(),
	}

	event, _ := NewEvent(eventType, source, notificationID, "notification", data)
	event.AddMetadata("service", source)
	event.AddMetadata("event_category", "notification")

	return event
}
