package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DomainEvent represents a domain event that occurred in the system
type DomainEvent interface {
	// EventID returns the unique identifier of the event
	EventID() string
	
	// EventType returns the type of the event
	EventType() string
	
	// AggregateID returns the ID of the aggregate that generated the event
	AggregateID() AggregateID
	
	// TenantID returns the tenant ID associated with the event
	TenantID() TenantID
	
	// OccurredAt returns when the event occurred
	OccurredAt() time.Time
	
	// Version returns the event version
	Version() int
	
	// Data returns the event data
	Data() map[string]interface{}
	
	// Metadata returns event metadata
	Metadata() map[string]interface{}
}

// BaseDomainEvent provides common functionality for domain events
type BaseDomainEvent struct {
	eventID     string
	eventType   string
	aggregateID AggregateID
	tenantID    TenantID
	occurredAt  time.Time
	version     int
	data        map[string]interface{}
	metadata    map[string]interface{}
}

// NewBaseDomainEvent creates a new base domain event
func NewBaseDomainEvent(
	eventID string,
	eventType string,
	aggregateID AggregateID,
	tenantID TenantID,
	data map[string]interface{},
) *BaseDomainEvent {
	return &BaseDomainEvent{
		eventID:     eventID,
		eventType:   eventType,
		aggregateID: aggregateID,
		tenantID:    tenantID,
		occurredAt:  time.Now(),
		version:     1,
		data:        data,
		metadata:    make(map[string]interface{}),
	}
}

// EventID returns the event ID
func (bde *BaseDomainEvent) EventID() string {
	return bde.eventID
}

// EventType returns the event type
func (bde *BaseDomainEvent) EventType() string {
	return bde.eventType
}

// AggregateID returns the aggregate ID
func (bde *BaseDomainEvent) AggregateID() AggregateID {
	return bde.aggregateID
}

// TenantID returns the tenant ID
func (bde *BaseDomainEvent) TenantID() TenantID {
	return bde.tenantID
}

// OccurredAt returns when the event occurred
func (bde *BaseDomainEvent) OccurredAt() time.Time {
	return bde.occurredAt
}

// Version returns the event version
func (bde *BaseDomainEvent) Version() int {
	return bde.version
}

// Data returns the event data
func (bde *BaseDomainEvent) Data() map[string]interface{} {
	return bde.data
}

// Metadata returns the event metadata
func (bde *BaseDomainEvent) Metadata() map[string]interface{} {
	return bde.metadata
}

// SetMetadata sets metadata for the event
func (bde *BaseDomainEvent) SetMetadata(key string, value interface{}) {
	bde.metadata[key] = value
}

// GetMetadata gets metadata from the event
func (bde *BaseDomainEvent) GetMetadata(key string) (interface{}, bool) {
	value, exists := bde.metadata[key]
	return value, exists
}

// ToJSON serializes the event to JSON
func (bde *BaseDomainEvent) ToJSON() ([]byte, error) {
	eventData := map[string]interface{}{
		"event_id":     bde.eventID,
		"event_type":   bde.eventType,
		"aggregate_id": bde.aggregateID.Value(),
		"tenant_id":    bde.tenantID.Value(),
		"occurred_at":  bde.occurredAt,
		"version":      bde.version,
		"data":         bde.data,
		"metadata":     bde.metadata,
	}
	return json.Marshal(eventData)
}

// DomainEventHandler handles domain events
type DomainEventHandler interface {
	// Handle processes a domain event
	Handle(ctx context.Context, event DomainEvent) error
	
	// CanHandle checks if the handler can process the event type
	CanHandle(eventType string) bool
	
	// HandlerName returns the name of the handler
	HandlerName() string
}

// DomainEventPublisher publishes domain events
type DomainEventPublisher interface {
	// Publish publishes a single domain event
	Publish(ctx context.Context, event DomainEvent) error
	
	// PublishBatch publishes multiple domain events
	PublishBatch(ctx context.Context, events []DomainEvent) error
	
	// Subscribe subscribes a handler to specific event types
	Subscribe(handler DomainEventHandler, eventTypes ...string) error
	
	// Unsubscribe removes a handler subscription
	Unsubscribe(handlerName string) error
}

// InMemoryDomainEventPublisher is an in-memory implementation of DomainEventPublisher
type InMemoryDomainEventPublisher struct {
	handlers map[string][]DomainEventHandler
	mutex    sync.RWMutex
}

// NewInMemoryDomainEventPublisher creates a new in-memory event publisher
func NewInMemoryDomainEventPublisher() *InMemoryDomainEventPublisher {
	return &InMemoryDomainEventPublisher{
		handlers: make(map[string][]DomainEventHandler),
	}
}

// Publish publishes a single domain event
func (p *InMemoryDomainEventPublisher) Publish(ctx context.Context, event DomainEvent) error {
	p.mutex.RLock()
	handlers, exists := p.handlers[event.EventType()]
	p.mutex.RUnlock()
	
	if !exists {
		return nil // No handlers for this event type
	}
	
	for _, handler := range handlers {
		if handler.CanHandle(event.EventType()) {
			if err := handler.Handle(ctx, event); err != nil {
				return fmt.Errorf("handler %s failed to process event %s: %w", 
					handler.HandlerName(), event.EventID(), err)
			}
		}
	}
	
	return nil
}

// PublishBatch publishes multiple domain events
func (p *InMemoryDomainEventPublisher) PublishBatch(ctx context.Context, events []DomainEvent) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe subscribes a handler to specific event types
func (p *InMemoryDomainEventPublisher) Subscribe(handler DomainEventHandler, eventTypes ...string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	for _, eventType := range eventTypes {
		if _, exists := p.handlers[eventType]; !exists {
			p.handlers[eventType] = make([]DomainEventHandler, 0)
		}
		p.handlers[eventType] = append(p.handlers[eventType], handler)
	}
	
	return nil
}

// Unsubscribe removes a handler subscription
func (p *InMemoryDomainEventPublisher) Unsubscribe(handlerName string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	for eventType, handlers := range p.handlers {
		newHandlers := make([]DomainEventHandler, 0)
		for _, handler := range handlers {
			if handler.HandlerName() != handlerName {
				newHandlers = append(newHandlers, handler)
			}
		}
		p.handlers[eventType] = newHandlers
	}
	
	return nil
}

// DomainEventStore stores domain events for event sourcing
type DomainEventStore interface {
	// SaveEvents saves events for an aggregate
	SaveEvents(aggregateID AggregateID, tenantID TenantID, events []DomainEvent, expectedVersion int) error
	
	// GetEvents retrieves events for an aggregate
	GetEvents(aggregateID AggregateID, tenantID TenantID, fromVersion int) ([]DomainEvent, error)
	
	// GetAllEvents retrieves all events for a tenant
	GetAllEvents(tenantID TenantID, fromTimestamp time.Time) ([]DomainEvent, error)
	
	// GetEventsByType retrieves events by type for a tenant
	GetEventsByType(tenantID TenantID, eventType string, fromTimestamp time.Time) ([]DomainEvent, error)
}

// EventSourcingRepository provides event sourcing capabilities
type EventSourcingRepository interface {
	// Save saves an aggregate using event sourcing
	Save(aggregate interface{}, expectedVersion int) error
	
	// GetByID reconstructs an aggregate from events
	GetByID(aggregateID AggregateID, tenantID TenantID) (interface{}, error)
	
	// GetVersion gets the current version of an aggregate
	GetVersion(aggregateID AggregateID, tenantID TenantID) (int, error)
}

// AggregateFactory creates aggregates from events
type AggregateFactory interface {
	// CreateAggregate creates a new aggregate instance
	CreateAggregate(aggregateID AggregateID, tenantID TenantID) interface{}
	
	// ApplyEvent applies an event to an aggregate
	ApplyEvent(aggregate interface{}, event DomainEvent) error
}

// EventBus provides event publishing and subscription capabilities
type EventBus interface {
	DomainEventPublisher
	
	// PublishAndWait publishes an event and waits for all handlers to complete
	PublishAndWait(ctx context.Context, event DomainEvent) error
	
	// PublishAsync publishes an event asynchronously
	PublishAsync(ctx context.Context, event DomainEvent) error
	
	// GetSubscriptions returns all current subscriptions
	GetSubscriptions() map[string][]string
}

// TenantAwareDomainEvent extends DomainEvent with tenant-specific functionality
type TenantAwareDomainEvent interface {
	DomainEvent
	
	// IsCrossTenant checks if the event should be published across tenants
	IsCrossTenant() bool
	
	// GetTargetTenants returns specific tenants that should receive this event
	GetTargetTenants() []TenantID
	
	// GetEventScope returns the scope of the event (tenant, global, etc.)
	GetEventScope() EventScope
}

// EventScope defines the scope of an event
type EventScope string

const (
	// EventScopeTenant - event is scoped to a single tenant
	EventScopeTenant EventScope = "tenant"
	
	// EventScopeGlobal - event is global across all tenants
	EventScopeGlobal EventScope = "global"
	
	// EventScopeMultiTenant - event targets specific tenants
	EventScopeMultiTenant EventScope = "multi_tenant"
)

// String returns the string representation of EventScope
func (es EventScope) String() string {
	return string(es)
}

// TenantAwareEventPublisher publishes events with tenant awareness
type TenantAwareEventPublisher interface {
	DomainEventPublisher
	
	// PublishToTenant publishes an event to a specific tenant
	PublishToTenant(ctx context.Context, event DomainEvent, tenantID TenantID) error
	
	// PublishToAllTenants publishes an event to all tenants
	PublishToAllTenants(ctx context.Context, event DomainEvent) error
	
	// PublishToTenants publishes an event to specific tenants
	PublishToTenants(ctx context.Context, event DomainEvent, tenantIDs []TenantID) error
}
