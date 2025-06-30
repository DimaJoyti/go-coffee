package events

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

// EventManager manages event publishing and consumption with type safety
type EventManager struct {
	registry   *EventRegistry
	serializer *EventSerializer
	producer   *kafka.Writer
	logger     Logger
}

// Logger interface for the event manager
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewEventManager creates a new event manager
func NewEventManager(registry *EventRegistry, serializer *EventSerializer, producer *kafka.Writer, logger Logger) *EventManager {
	return &EventManager{
		registry:   registry,
		serializer: serializer,
		producer:   producer,
		logger:     logger,
	}
}

// PublishEvent publishes an event to Kafka with type safety
func (em *EventManager) PublishEvent(ctx context.Context, topic, eventType, version string, message proto.Message, metadata map[string]string) error {
	// Add correlation and trace IDs if not present
	if metadata == nil {
		metadata = make(map[string]string)
	}
	
	if metadata["correlation_id"] == "" {
		metadata["correlation_id"] = generateCorrelationID()
	}
	
	if metadata["trace_id"] == "" {
		metadata["trace_id"] = generateTraceID()
	}

	// Serialize the event
	data, err := em.serializer.SerializeEvent(eventType, version, message, metadata)
	if err != nil {
		em.logger.Error("Failed to serialize event", err, 
			"event_type", eventType, 
			"version", version,
			"topic", topic)
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Create Kafka message
	kafkaMessage := kafka.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("%s:%s", eventType, metadata["correlation_id"])),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(eventType)},
			{Key: "event-version", Value: []byte(version)},
			{Key: "content-type", Value: []byte("application/x-protobuf")},
			{Key: "correlation-id", Value: []byte(metadata["correlation_id"])},
			{Key: "trace-id", Value: []byte(metadata["trace_id"])},
			{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}

	// Publish to Kafka
	if err := em.producer.WriteMessages(ctx, kafkaMessage); err != nil {
		em.logger.Error("Failed to publish event to Kafka", err,
			"event_type", eventType,
			"version", version,
			"topic", topic,
			"correlation_id", metadata["correlation_id"])
		return fmt.Errorf("failed to publish event: %w", err)
	}

	em.logger.Info("Event published successfully",
		"event_type", eventType,
		"version", version,
		"topic", topic,
		"correlation_id", metadata["correlation_id"],
		"trace_id", metadata["trace_id"])

	return nil
}

// ConsumeEvents consumes events from Kafka with type safety
func (em *EventManager) ConsumeEvents(ctx context.Context, reader *kafka.Reader, handler EventHandler) error {
	em.logger.Info("Starting event consumption", "topic", reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			em.logger.Info("Event consumption stopped")
			return ctx.Err()
		default:
			message, err := reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				em.logger.Error("Failed to read message from Kafka", err)
				continue
			}

			if err := em.processMessage(ctx, &message, handler); err != nil {
				em.logger.Error("Failed to process message", err,
					"topic", message.Topic,
					"partition", message.Partition,
					"offset", message.Offset)
				// Continue processing other messages
				continue
			}

			// Commit the message
			if err := reader.CommitMessages(ctx, message); err != nil {
				em.logger.Error("Failed to commit message", err,
					"topic", message.Topic,
					"partition", message.Partition,
					"offset", message.Offset)
			}
		}
	}
}

// processMessage processes a single Kafka message
func (em *EventManager) processMessage(ctx context.Context, message *kafka.Message, handler EventHandler) error {
	// Extract event metadata from headers
	eventType := em.getHeaderValue(message.Headers, "event-type")
	version := em.getHeaderValue(message.Headers, "event-version")
	correlationID := em.getHeaderValue(message.Headers, "correlation-id")
	traceID := em.getHeaderValue(message.Headers, "trace-id")

	if eventType == "" || version == "" {
		return fmt.Errorf("missing event-type or event-version headers")
	}

	em.logger.Debug("Processing event",
		"event_type", eventType,
		"version", version,
		"correlation_id", correlationID,
		"trace_id", traceID,
		"topic", message.Topic)

	// Deserialize the event
	deserializedEventType, deserializedVersion, event, metadata, err := em.serializer.DeserializeEvent(message.Value)
	if err != nil {
		return fmt.Errorf("failed to deserialize event: %w", err)
	}

	// Verify event type consistency
	if deserializedEventType != eventType || deserializedVersion != version {
		em.logger.Warn("Event type mismatch between headers and payload",
			"header_type", eventType,
			"header_version", version,
			"payload_type", deserializedEventType,
			"payload_version", deserializedVersion)
	}

	// Create event context
	eventCtx := &EventContext{
		Context:       ctx,
		Topic:         message.Topic,
		Partition:     message.Partition,
		Offset:        message.Offset,
		EventType:     deserializedEventType,
		Version:       deserializedVersion,
		CorrelationID: correlationID,
		TraceID:       traceID,
		Metadata:      metadata,
		Timestamp:     time.Now(),
	}

	// Handle the event
	return handler.HandleEvent(eventCtx, event)
}

// getHeaderValue extracts a header value from Kafka message headers
func (em *EventManager) getHeaderValue(headers []kafka.Header, key string) string {
	for _, header := range headers {
		if header.Key == key {
			return string(header.Value)
		}
	}
	return ""
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	HandleEvent(ctx *EventContext, event proto.Message) error
}

// EventContext provides context information for event handling
type EventContext struct {
	Context       context.Context
	Topic         string
	Partition     int
	Offset        int64
	EventType     string
	Version       string
	CorrelationID string
	TraceID       string
	Metadata      map[string]string
	Timestamp     time.Time
}

// TypedEventHandler provides type-safe event handling
type TypedEventHandler struct {
	handlers map[string]map[string]func(*EventContext, proto.Message) error
	logger   Logger
}

// NewTypedEventHandler creates a new typed event handler
func NewTypedEventHandler(logger Logger) *TypedEventHandler {
	return &TypedEventHandler{
		handlers: make(map[string]map[string]func(*EventContext, proto.Message) error),
		logger:   logger,
	}
}

// RegisterHandler registers a handler for a specific event type and version
func (teh *TypedEventHandler) RegisterHandler(eventType, version string, handler func(*EventContext, proto.Message) error) {
	if teh.handlers[eventType] == nil {
		teh.handlers[eventType] = make(map[string]func(*EventContext, proto.Message) error)
	}
	teh.handlers[eventType][version] = handler
	
	teh.logger.Info("Registered event handler",
		"event_type", eventType,
		"version", version)
}

// HandleEvent implements the EventHandler interface
func (teh *TypedEventHandler) HandleEvent(ctx *EventContext, event proto.Message) error {
	// Find handler for this event type and version
	versionHandlers, exists := teh.handlers[ctx.EventType]
	if !exists {
		teh.logger.Warn("No handlers registered for event type", "event_type", ctx.EventType)
		return fmt.Errorf("no handlers registered for event type: %s", ctx.EventType)
	}

	handler, exists := versionHandlers[ctx.Version]
	if !exists {
		teh.logger.Warn("No handler registered for event version",
			"event_type", ctx.EventType,
			"version", ctx.Version)
		return fmt.Errorf("no handler registered for event type %s version %s", ctx.EventType, ctx.Version)
	}

	// Execute the handler
	teh.logger.Debug("Executing event handler",
		"event_type", ctx.EventType,
		"version", ctx.Version,
		"correlation_id", ctx.CorrelationID)

	return handler(ctx, event)
}

// EventMetrics tracks event processing metrics
type EventMetrics struct {
	EventsPublished   map[string]int64
	EventsConsumed    map[string]int64
	ProcessingErrors  map[string]int64
	ProcessingLatency map[string]time.Duration
}

// NewEventMetrics creates a new event metrics tracker
func NewEventMetrics() *EventMetrics {
	return &EventMetrics{
		EventsPublished:   make(map[string]int64),
		EventsConsumed:    make(map[string]int64),
		ProcessingErrors:  make(map[string]int64),
		ProcessingLatency: make(map[string]time.Duration),
	}
}

// RecordPublished records a published event
func (em *EventMetrics) RecordPublished(eventType string) {
	em.EventsPublished[eventType]++
}

// RecordConsumed records a consumed event
func (em *EventMetrics) RecordConsumed(eventType string) {
	em.EventsConsumed[eventType]++
}

// RecordError records a processing error
func (em *EventMetrics) RecordError(eventType string) {
	em.ProcessingErrors[eventType]++
}

// RecordLatency records processing latency
func (em *EventMetrics) RecordLatency(eventType string, latency time.Duration) {
	em.ProcessingLatency[eventType] = latency
}

// Utility functions for generating IDs
func generateCorrelationID() string {
	return fmt.Sprintf("corr-%d", time.Now().UnixNano())
}

func generateTraceID() string {
	return fmt.Sprintf("trace-%d", time.Now().UnixNano())
}
