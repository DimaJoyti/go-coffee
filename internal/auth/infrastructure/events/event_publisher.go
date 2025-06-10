package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event *domain.DomainEvent) error
	PublishBatch(ctx context.Context, events []*domain.DomainEvent) error
	Subscribe(ctx context.Context, eventType string, handler domain.EventHandler) error
	Unsubscribe(ctx context.Context, eventType string) error
	Close() error
}

// RedisEventPublisher implements EventPublisher using Redis pub/sub
type RedisEventPublisher struct {
	client      *redis.Client
	logger      *logger.Logger
	subscribers map[string][]domain.EventHandler
	pubsub      *redis.PubSub
	mu          sync.RWMutex
	done        chan struct{}
	wg          sync.WaitGroup
}

// NewRedisEventPublisher creates a new Redis event publisher
func NewRedisEventPublisher(client *redis.Client, logger *logger.Logger) EventPublisher {
	return &RedisEventPublisher{
		client:      client,
		logger:      logger,
		subscribers: make(map[string][]domain.EventHandler),
		done:        make(chan struct{}),
	}
}

// Publish publishes a single domain event
func (ep *RedisEventPublisher) Publish(ctx context.Context, event *domain.DomainEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Serialize event
	eventData, err := json.Marshal(event)
	if err != nil {
		ep.logger.ErrorWithFields("Failed to marshal event for publishing",
			logger.Error(err),
			logger.String("event_id", event.ID),
			logger.String("event_type", event.Type))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to specific event type channel
	eventChannel := ep.getEventChannel(event.Type)
	err = ep.client.Publish(ctx, eventChannel, eventData).Err()
	if err != nil {
		ep.logger.ErrorWithFields("Failed to publish event",
			logger.Error(err),
			logger.String("event_id", event.ID),
			logger.String("event_type", event.Type),
			logger.String("channel", eventChannel))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Also publish to global events channel
	globalChannel := ep.getGlobalChannel()
	err = ep.client.Publish(ctx, globalChannel, eventData).Err()
	if err != nil {
		ep.logger.ErrorWithFields("Failed to publish event to global channel",
			logger.Error(err),
			logger.String("event_id", event.ID),
			logger.String("channel", globalChannel))
		// Don't return error for global channel failure
	}

	ep.logger.InfoWithFields("Event published successfully",
		logger.String("event_id", event.ID),
		logger.String("event_type", event.Type),
		logger.String("aggregate_id", event.AggregateID))

	return nil
}

// PublishBatch publishes multiple domain events
func (ep *RedisEventPublisher) PublishBatch(ctx context.Context, events []*domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Use pipeline for batch publishing
	pipe := ep.client.Pipeline()

	for _, event := range events {
		if event == nil {
			continue
		}

		// Serialize event
		eventData, err := json.Marshal(event)
		if err != nil {
			ep.logger.ErrorWithFields("Failed to marshal event in batch",
				logger.Error(err),
				logger.String("event_id", event.ID))
			continue
		}

		// Add to pipeline
		eventChannel := ep.getEventChannel(event.Type)
		pipe.Publish(ctx, eventChannel, eventData)

		globalChannel := ep.getGlobalChannel()
		pipe.Publish(ctx, globalChannel, eventData)
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		ep.logger.ErrorWithFields("Failed to publish event batch",
			logger.Error(err),
			logger.Int("event_count", len(events)))
		return fmt.Errorf("failed to publish event batch: %w", err)
	}

	ep.logger.InfoWithFields("Event batch published successfully",
		logger.Int("event_count", len(events)))

	return nil
}

// Subscribe subscribes to events of a specific type
func (ep *RedisEventPublisher) Subscribe(ctx context.Context, eventType string, handler domain.EventHandler) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	// Add handler to subscribers
	if ep.subscribers[eventType] == nil {
		ep.subscribers[eventType] = make([]domain.EventHandler, 0)
	}
	ep.subscribers[eventType] = append(ep.subscribers[eventType], handler)

	// If this is the first subscriber for this event type, start listening
	if len(ep.subscribers[eventType]) == 1 {
		if err := ep.startListening(eventType); err != nil {
			// Remove the handler if we failed to start listening
			ep.subscribers[eventType] = ep.subscribers[eventType][:len(ep.subscribers[eventType])-1]
			return fmt.Errorf("failed to start listening for event type %s: %w", eventType, err)
		}
	}

	ep.logger.InfoWithFields("Event handler subscribed",
		logger.String("event_type", eventType),
		logger.Int("handler_count", len(ep.subscribers[eventType])))

	return nil
}

// Unsubscribe unsubscribes from events of a specific type
func (ep *RedisEventPublisher) Unsubscribe(ctx context.Context, eventType string) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	delete(ep.subscribers, eventType)

	// Stop listening if no more subscribers
	if ep.pubsub != nil {
		eventChannel := ep.getEventChannel(eventType)
		err := ep.pubsub.Unsubscribe(ctx, eventChannel)
		if err != nil {
			ep.logger.ErrorWithFields("Failed to unsubscribe from channel",
				logger.Error(err),
				logger.String("channel", eventChannel))
			return fmt.Errorf("failed to unsubscribe from channel: %w", err)
		}
	}

	ep.logger.InfoWithFields("Unsubscribed from event type",
		logger.String("event_type", eventType))

	return nil
}

// Close closes the event publisher and all subscriptions
func (ep *RedisEventPublisher) Close() error {
	close(ep.done)
	ep.wg.Wait()

	ep.mu.Lock()
	defer ep.mu.Unlock()

	if ep.pubsub != nil {
		err := ep.pubsub.Close()
		if err != nil {
			ep.logger.ErrorWithFields("Failed to close pubsub", logger.Error(err))
			return fmt.Errorf("failed to close pubsub: %w", err)
		}
		ep.pubsub = nil
	}

	ep.logger.Info("Event publisher closed")
	return nil
}

// startListening starts listening for events of a specific type
func (ep *RedisEventPublisher) startListening(eventType string) error {
	if ep.pubsub == nil {
		ep.pubsub = ep.client.Subscribe(context.Background())
	}

	eventChannel := ep.getEventChannel(eventType)
	err := ep.pubsub.Subscribe(context.Background(), eventChannel)
	if err != nil {
		return fmt.Errorf("failed to subscribe to channel %s: %w", eventChannel, err)
	}

	// Start message processing goroutine if not already started
	if len(ep.subscribers) == 1 {
		ep.wg.Add(1)
		go ep.processMessages()
	}

	return nil
}

// processMessages processes incoming messages from Redis pub/sub
func (ep *RedisEventPublisher) processMessages() {
	defer ep.wg.Done()

	ch := ep.pubsub.Channel()

	for {
		select {
		case <-ep.done:
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			ep.handleMessage(msg)
		}
	}
}

// handleMessage handles a single message from Redis pub/sub
func (ep *RedisEventPublisher) handleMessage(msg *redis.Message) {
	// Parse event type from channel name
	eventType := ep.parseEventTypeFromChannel(msg.Channel)
	if eventType == "" {
		ep.logger.WarnWithFields("Unknown channel format",
			logger.String("channel", msg.Channel))
		return
	}

	// Deserialize event
	var event domain.DomainEvent
	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		ep.logger.ErrorWithFields("Failed to unmarshal event from message",
			logger.Error(err),
			logger.String("channel", msg.Channel))
		return
	}

	// Get handlers for this event type
	ep.mu.RLock()
	handlers := ep.subscribers[eventType]
	ep.mu.RUnlock()

	// Execute handlers
	for _, handler := range handlers {
		go func(h domain.EventHandler, e domain.DomainEvent) {
			defer func() {
				if r := recover(); r != nil {
					ep.logger.ErrorWithFields("Event handler panicked",
						logger.Any("panic", r),
						logger.String("event_id", e.ID),
						logger.String("event_type", e.Type))
				}
			}()

			if err := h.Handle(&e); err != nil {
				ep.logger.ErrorWithFields("Event handler failed",
					logger.Error(err),
					logger.String("event_id", e.ID),
					logger.String("event_type", e.Type))
			}
		}(handler, event)
	}
}

// Helper methods

func (ep *RedisEventPublisher) getEventChannel(eventType string) string {
	return fmt.Sprintf("auth:events:%s", eventType)
}

func (ep *RedisEventPublisher) getGlobalChannel() string {
	return "auth:events:global"
}

func (ep *RedisEventPublisher) parseEventTypeFromChannel(channel string) string {
	// Expected format: "auth:events:{eventType}"
	const prefix = "auth:events:"
	if len(channel) <= len(prefix) {
		return ""
	}

	eventType := channel[len(prefix):]
	if eventType == "global" {
		return ""
	}

	return eventType
}

// EventHandlerFunc is a convenience type for creating event handlers from functions
type EventHandlerFunc func(event *domain.DomainEvent) error

// Handle implements the EventHandler interface
func (f EventHandlerFunc) Handle(event *domain.DomainEvent) error {
	return f(event)
}

// CreateEventHandler creates an event handler from a function
func CreateEventHandler(fn func(event *domain.DomainEvent) error) domain.EventHandler {
	return EventHandlerFunc(fn)
}
