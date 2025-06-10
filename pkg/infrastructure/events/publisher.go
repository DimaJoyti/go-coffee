package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/redis"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	// Publish single event
	Publish(ctx context.Context, event *Event) error

	// Publish multiple events
	PublishBatch(ctx context.Context, events []*Event) error

	// Publish to specific channel
	PublishToChannel(ctx context.Context, channel string, event *Event) error

	// Start/Stop publisher
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Health check
	HealthCheck(ctx context.Context) error
}

// RedisEventPublisher implements EventPublisher using Redis pub/sub
type RedisEventPublisher struct {
	client    redis.ClientInterface
	config    *config.EventPublisherConfig
	logger    *logger.Logger
	eventChan chan *publishRequest
	stopChan  chan struct{}
	wg        sync.WaitGroup
	running   bool
	mutex     sync.RWMutex
}

// publishRequest represents a publish request
type publishRequest struct {
	event   *Event
	channel string
	ctx     context.Context
	errChan chan error
}

// NewRedisEventPublisher creates a new Redis event publisher
func NewRedisEventPublisher(client redis.ClientInterface, cfg *config.EventPublisherConfig, logger *logger.Logger) EventPublisher {
	return &RedisEventPublisher{
		client:    client,
		config:    cfg,
		logger:    logger,
		eventChan: make(chan *publishRequest, cfg.BufferSize),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the event publisher
func (p *RedisEventPublisher) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		return fmt.Errorf("publisher is already running")
	}

	p.running = true

	// Start worker goroutines
	for i := 0; i < p.config.Workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	p.logger.InfoWithFields("Event publisher started",
		logger.Int("workers", p.config.Workers),
		logger.Int("buffer_size", p.config.BufferSize))

	return nil
}

// Stop stops the event publisher
func (p *RedisEventPublisher) Stop(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return nil
	}

	p.running = false
	close(p.stopChan)

	// Wait for workers to finish
	p.wg.Wait()

	p.logger.Info("Event publisher stopped")
	return nil
}

// Publish publishes a single event
func (p *RedisEventPublisher) Publish(ctx context.Context, event *Event) error {
	return p.PublishToChannel(ctx, p.getEventChannel(event), event)
}

// PublishBatch publishes multiple events
func (p *RedisEventPublisher) PublishBatch(ctx context.Context, events []*Event) error {
	if len(events) == 0 {
		return nil
	}

	errChan := make(chan error, len(events))

	for _, event := range events {
		select {
		case p.eventChan <- &publishRequest{
			event:   event,
			channel: p.getEventChannel(event),
			ctx:     ctx,
			errChan: errChan,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Wait for all events to be processed
	var lastErr error
	for i := 0; i < len(events); i++ {
		select {
		case err := <-errChan:
			if err != nil {
				lastErr = err
				p.logger.WithError(err).Error("Failed to publish event in batch")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// PublishToChannel publishes an event to a specific channel
func (p *RedisEventPublisher) PublishToChannel(ctx context.Context, channel string, event *Event) error {
	p.mutex.RLock()
	running := p.running
	p.mutex.RUnlock()

	if !running {
		return fmt.Errorf("publisher is not running")
	}

	errChan := make(chan error, 1)

	select {
	case p.eventChan <- &publishRequest{
		event:   event,
		channel: channel,
		ctx:     ctx,
		errChan: errChan,
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// HealthCheck checks the health of the publisher
func (p *RedisEventPublisher) HealthCheck(ctx context.Context) error {
	return p.client.Ping(ctx)
}

// worker processes publish requests
func (p *RedisEventPublisher) worker(workerID int) {
	defer p.wg.Done()

	p.logger.With(logger.Int("worker_id", workerID)).Debug("Event publisher worker started")

	for {
		select {
		case req := <-p.eventChan:
			err := p.publishEvent(req.ctx, req.channel, req.event)
			select {
			case req.errChan <- err:
			case <-req.ctx.Done():
			}
		case <-p.stopChan:
			p.logger.With(logger.Int("worker_id", workerID)).Debug("Event publisher worker stopped")
			return
		}
	}
}

// publishEvent publishes a single event to Redis
func (p *RedisEventPublisher) publishEvent(ctx context.Context, channel string, event *Event) error {
	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize event
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to Redis
	if err := p.client.Publish(ctx, channel, string(eventData)); err != nil {
		return fmt.Errorf("failed to publish event to channel %s: %w", channel, err)
	}

	p.logger.With(
		logger.String("event_id", event.ID),
		logger.String("event_type", event.Type),
		logger.String("channel", channel),
	).Debug("Event published")

	return nil
}

// getEventChannel returns the channel name for an event
func (p *RedisEventPublisher) getEventChannel(event *Event) string {
	return fmt.Sprintf("events:%s", event.Type)
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
	CanHandle(eventType string) bool
	GetHandlerName() string
}

// EventSubscriber defines the interface for subscribing to events
type EventSubscriber interface {
	// Subscribe to specific event types
	Subscribe(ctx context.Context, eventTypes []string, handler EventHandler) error

	// Subscribe to all events
	SubscribeAll(ctx context.Context, handler EventHandler) error

	// Subscribe with pattern matching
	SubscribePattern(ctx context.Context, pattern string, handler EventHandler) error

	// Unsubscribe from event types
	Unsubscribe(ctx context.Context, eventTypes []string) error

	// Start/Stop subscriber
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Health check
	HealthCheck(ctx context.Context) error
}

// RedisEventSubscriber implements EventSubscriber using Redis pub/sub
type RedisEventSubscriber struct {
	client        redis.ClientInterface
	config        *config.EventSubscriberConfig
	logger        *logger.Logger
	handlers      map[string][]EventHandler
	subscriptions map[string]*redisPubSub
	stopChan      chan struct{}
	wg            sync.WaitGroup
	running       bool
	mutex         sync.RWMutex
}

// redisPubSub wraps Redis pub/sub
type redisPubSub struct {
	pubsub   interface{} // Redis PubSub interface
	channels []string
}

// NewRedisEventSubscriber creates a new Redis event subscriber
func NewRedisEventSubscriber(client redis.ClientInterface, cfg *config.EventSubscriberConfig, logger *logger.Logger) EventSubscriber {
	return &RedisEventSubscriber{
		client:        client,
		config:        cfg,
		logger:        logger,
		handlers:      make(map[string][]EventHandler),
		subscriptions: make(map[string]*redisPubSub),
		stopChan:      make(chan struct{}),
	}
}

// Start starts the event subscriber
func (s *RedisEventSubscriber) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("subscriber is already running")
	}

	s.running = true
	s.logger.Info("Event subscriber started")
	return nil
}

// Stop stops the event subscriber
func (s *RedisEventSubscriber) Stop(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	close(s.stopChan)

	// Close all subscriptions
	for _, sub := range s.subscriptions {
		// Close subscription (implementation depends on Redis client)
		_ = sub
	}

	// Wait for workers to finish
	s.wg.Wait()

	s.logger.Info("Event subscriber stopped")
	return nil
}

// Subscribe subscribes to specific event types
func (s *RedisEventSubscriber) Subscribe(ctx context.Context, eventTypes []string, handler EventHandler) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return fmt.Errorf("subscriber is not running")
	}

	// Add handler for each event type
	for _, eventType := range eventTypes {
		if s.handlers[eventType] == nil {
			s.handlers[eventType] = make([]EventHandler, 0)
		}
		s.handlers[eventType] = append(s.handlers[eventType], handler)
	}

	// Create channels for subscription
	channels := make([]string, len(eventTypes))
	for i, eventType := range eventTypes {
		channels[i] = fmt.Sprintf("events:%s", eventType)
	}

	// Subscribe to Redis channels
	pubsub := s.client.Subscribe(ctx, channels...)
	s.subscriptions[handler.GetHandlerName()] = &redisPubSub{
		pubsub:   pubsub,
		channels: channels,
	}

	// Start message processing goroutine
	s.wg.Add(1)
	go s.processMessages(ctx, pubsub, handler)

	s.logger.InfoWithFields("Subscribed to events",
		logger.Any("event_types", eventTypes),
		logger.String("handler", handler.GetHandlerName()))

	return nil
}

// SubscribeAll subscribes to all events
func (s *RedisEventSubscriber) SubscribeAll(ctx context.Context, handler EventHandler) error {
	return s.SubscribePattern(ctx, "events:*", handler)
}

// SubscribePattern subscribes with pattern matching
func (s *RedisEventSubscriber) SubscribePattern(ctx context.Context, pattern string, handler EventHandler) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return fmt.Errorf("subscriber is not running")
	}

	// Subscribe to Redis pattern
	pubsub := s.client.PSubscribe(ctx, pattern)
	s.subscriptions[handler.GetHandlerName()] = &redisPubSub{
		pubsub:   pubsub,
		channels: []string{pattern},
	}

	// Start message processing goroutine
	s.wg.Add(1)
	go s.processMessages(ctx, pubsub, handler)

	s.logger.InfoWithFields("Subscribed to pattern",
		logger.String("pattern", pattern),
		logger.String("handler", handler.GetHandlerName()))

	return nil
}

// Unsubscribe unsubscribes from event types
func (s *RedisEventSubscriber) Unsubscribe(ctx context.Context, eventTypes []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Remove handlers for event types
	for _, eventType := range eventTypes {
		delete(s.handlers, eventType)
	}

	s.logger.InfoWithFields("Unsubscribed from events", logger.Any("event_types", eventTypes))
	return nil
}

// HealthCheck checks the health of the subscriber
func (s *RedisEventSubscriber) HealthCheck(ctx context.Context) error {
	return s.client.Ping(ctx)
}

// processMessages processes incoming messages from Redis pub/sub
func (s *RedisEventSubscriber) processMessages(ctx context.Context, pubsub interface{}, handler EventHandler) {
	defer s.wg.Done()

	// This is a simplified implementation
	// In a real implementation, you would receive messages from the pubsub
	// and process them with the handler

	s.logger.With(logger.String("handler", handler.GetHandlerName())).Debug("Message processor started")

	for {
		select {
		case <-s.stopChan:
			s.logger.With(logger.String("handler", handler.GetHandlerName())).Debug("Message processor stopped")
			return
		case <-ctx.Done():
			return
		}
	}
}
