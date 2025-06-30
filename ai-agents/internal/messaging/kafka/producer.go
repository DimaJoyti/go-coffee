package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go-coffee-ai-agents/internal/observability"
)

// Producer handles Kafka message production with observability
type Producer struct {
	connectionManager *ConnectionManager
	logger            *observability.StructuredLogger
	metrics           *observability.MetricsCollector
	tracing           *observability.TracingHelper
}

// NewProducer creates a new Kafka producer
func NewProducer(
	connectionManager *ConnectionManager,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *Producer {
	return &Producer{
		connectionManager: connectionManager,
		logger:            logger,
		metrics:           metrics,
		tracing:           tracing,
	}
}

// Message represents a Kafka message with metadata
type Message struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Key       string                 `json:"key,omitempty"`
	Value     interface{}            `json:"value"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewMessage creates a new message with generated ID and timestamp
func NewMessage(topic string, value interface{}) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Topic:     topic,
		Value:     value,
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// WithKey sets the message key
func (m *Message) WithKey(key string) *Message {
	m.Key = key
	return m
}

// WithHeader adds a header to the message
func (m *Message) WithHeader(key, value string) *Message {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	m.Headers[key] = value
	return m
}

// WithMetadata adds metadata to the message
func (m *Message) WithMetadata(key string, value interface{}) *Message {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
	return m
}

// Publish publishes a single message to Kafka
func (p *Producer) Publish(ctx context.Context, message *Message) error {
	ctx, span := p.tracing.StartKafkaSpan(ctx, "PUBLISH", message.Topic)
	defer span.End()

	start := time.Now()

	// Add tracing headers
	p.injectTraceHeaders(ctx, message)

	// Serialize message value
	valueBytes, err := p.serializeValue(message.Value)
	if err != nil {
		p.tracing.RecordError(span, err, "Failed to serialize message value")
		return fmt.Errorf("failed to serialize message value: %w", err)
	}

	// Create Kafka message
	kafkaMessage := kafka.Message{
		Key:   []byte(message.Key),
		Value: valueBytes,
		Time:  message.Timestamp,
	}

	// Add headers
	for key, value := range message.Headers {
		kafkaMessage.Headers = append(kafkaMessage.Headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	// Get writer for topic
	writer := p.connectionManager.GetWriter(message.Topic)

	// Publish message
	err = writer.WriteMessages(ctx, kafkaMessage)
	duration := time.Since(start)

	// Record metrics
	if p.metrics != nil {
		counters := p.metrics.GetCounters()
		histograms := p.metrics.GetHistograms()
		if counters != nil && histograms != nil {
			if err != nil {
				counters.KafkaMessagesError.Add(ctx, 1)
			} else {
				counters.KafkaMessagesSuccess.Add(ctx, 1)
			}
			counters.KafkaMessagesTotal.Add(ctx, 1)
			histograms.KafkaPublishDuration.Record(ctx, duration.Seconds())
		}
	}

	if err != nil {
		p.tracing.RecordError(span, err, "Failed to publish message to Kafka")
		p.logger.ErrorContext(ctx, "Failed to publish message to Kafka", err,
			"message_id", message.ID,
			"topic", message.Topic,
			"key", message.Key,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to publish message to topic %s: %w", message.Topic, err)
	}

	p.tracing.RecordSuccess(span, "Message published successfully")
	p.logger.InfoContext(ctx, "Message published successfully",
		"message_id", message.ID,
		"topic", message.Topic,
		"key", message.Key,
		"duration_ms", duration.Milliseconds())

	return nil
}

// PublishBatch publishes multiple messages to Kafka in a batch
func (p *Producer) PublishBatch(ctx context.Context, messages []*Message) error {
	if len(messages) == 0 {
		return nil
	}

	// Group messages by topic
	messagesByTopic := make(map[string][]*Message)
	for _, message := range messages {
		messagesByTopic[message.Topic] = append(messagesByTopic[message.Topic], message)
	}

	// Publish each topic's messages
	for topic, topicMessages := range messagesByTopic {
		if err := p.publishBatchToTopic(ctx, topic, topicMessages); err != nil {
			return fmt.Errorf("failed to publish batch to topic %s: %w", topic, err)
		}
	}

	return nil
}

// publishBatchToTopic publishes a batch of messages to a specific topic
func (p *Producer) publishBatchToTopic(ctx context.Context, topic string, messages []*Message) error {
	ctx, span := p.tracing.StartKafkaSpan(ctx, "PUBLISH_BATCH", topic)
	defer span.End()

	start := time.Now()

	p.logger.InfoContext(ctx, "Publishing message batch",
		"topic", topic,
		"message_count", len(messages))

	// Convert to Kafka messages
	kafkaMessages := make([]kafka.Message, len(messages))
	for i, message := range messages {
		// Add tracing headers
		p.injectTraceHeaders(ctx, message)

		// Serialize message value
		valueBytes, err := p.serializeValue(message.Value)
		if err != nil {
			p.tracing.RecordError(span, err, "Failed to serialize message value")
			return fmt.Errorf("failed to serialize message %s: %w", message.ID, err)
		}

		kafkaMessages[i] = kafka.Message{
			Key:   []byte(message.Key),
			Value: valueBytes,
			Time:  message.Timestamp,
		}

		// Add headers
		for key, value := range message.Headers {
			kafkaMessages[i].Headers = append(kafkaMessages[i].Headers, kafka.Header{
				Key:   key,
				Value: []byte(value),
			})
		}
	}

	// Get writer for topic
	writer := p.connectionManager.GetWriter(topic)

	// Publish batch
	err := writer.WriteMessages(ctx, kafkaMessages...)
	duration := time.Since(start)

	// Record metrics
	if p.metrics != nil {
		counters := p.metrics.GetCounters()
		histograms := p.metrics.GetHistograms()
		if counters != nil && histograms != nil {
			if err != nil {
				counters.KafkaMessagesError.Add(ctx, int64(len(messages)))
			} else {
				counters.KafkaMessagesSuccess.Add(ctx, int64(len(messages)))
			}
			counters.KafkaMessagesTotal.Add(ctx, int64(len(messages)))
			histograms.KafkaPublishDuration.Record(ctx, duration.Seconds())
		}
	}

	if err != nil {
		p.tracing.RecordError(span, err, "Failed to publish message batch")
		p.logger.ErrorContext(ctx, "Failed to publish message batch", err,
			"topic", topic,
			"message_count", len(messages),
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to publish batch to topic %s: %w", topic, err)
	}

	p.tracing.RecordSuccess(span, "Message batch published successfully")
	p.logger.InfoContext(ctx, "Message batch published successfully",
		"topic", topic,
		"message_count", len(messages),
		"duration_ms", duration.Milliseconds())

	return nil
}

// PublishEvent publishes a domain event to Kafka
func (p *Producer) PublishEvent(ctx context.Context, eventType string, payload interface{}) error {
	message := NewMessage(eventType, payload).
		WithKey(uuid.New().String()).
		WithHeader("event_type", eventType).
		WithHeader("content_type", "application/json").
		WithMetadata("source", "beverage-inventor-agent").
		WithMetadata("version", "1.0")

	return p.Publish(ctx, message)
}

// PublishBeverageEvent publishes a beverage-related event
func (p *Producer) PublishBeverageEvent(ctx context.Context, eventType string, beverageID string, payload interface{}) error {
	message := NewMessage(eventType, payload).
		WithKey(beverageID).
		WithHeader("event_type", eventType).
		WithHeader("entity_type", "beverage").
		WithHeader("entity_id", beverageID).
		WithHeader("content_type", "application/json").
		WithMetadata("source", "beverage-inventor-agent").
		WithMetadata("version", "1.0")

	return p.Publish(ctx, message)
}

// PublishTaskEvent publishes a task-related event
func (p *Producer) PublishTaskEvent(ctx context.Context, eventType string, taskID string, payload interface{}) error {
	message := NewMessage(eventType, payload).
		WithKey(taskID).
		WithHeader("event_type", eventType).
		WithHeader("entity_type", "task").
		WithHeader("entity_id", taskID).
		WithHeader("content_type", "application/json").
		WithMetadata("source", "beverage-inventor-agent").
		WithMetadata("version", "1.0")

	return p.Publish(ctx, message)
}

// PublishNotificationEvent publishes a notification event
func (p *Producer) PublishNotificationEvent(ctx context.Context, eventType string, notificationID string, payload interface{}) error {
	message := NewMessage(eventType, payload).
		WithKey(notificationID).
		WithHeader("event_type", eventType).
		WithHeader("entity_type", "notification").
		WithHeader("entity_id", notificationID).
		WithHeader("content_type", "application/json").
		WithMetadata("source", "beverage-inventor-agent").
		WithMetadata("version", "1.0")

	return p.Publish(ctx, message)
}

// injectTraceHeaders injects tracing headers into the message
func (p *Producer) injectTraceHeaders(ctx context.Context, message *Message) {
	// Extract trace context and add to headers using standard context values
	// These functions need to be implemented in observability package
	if traceID := getContextString(ctx, "trace_id"); traceID != "" {
		message.WithHeader("trace_id", traceID)
	}
	if spanID := getContextString(ctx, "span_id"); spanID != "" {
		message.WithHeader("span_id", spanID)
	}
	
	// Add correlation ID if present
	if correlationID := getContextString(ctx, "correlation_id"); correlationID != "" {
		message.WithHeader("correlation_id", correlationID)
	}
}

// getContextString retrieves a string value from context
func getContextString(ctx context.Context, key string) string {
	if value := ctx.Value(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// serializeValue serializes the message value to bytes
func (p *Producer) serializeValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}

// GetStats returns producer statistics
func (p *Producer) GetStats() ProducerStats {
	connStats := p.connectionManager.GetConnectionStats()
	return ProducerStats{
		ConnectionStats: connStats,
		WritersCount:    connStats.WritersCount,
	}
}

// ProducerStats represents producer statistics
type ProducerStats struct {
	ConnectionStats ConnectionStats `json:"connection_stats"`
	WritersCount    int             `json:"writers_count"`
}

// EventPayload represents a generic event payload
type EventPayload struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
}

// NewEventPayload creates a new event payload
func NewEventPayload(eventType, source, subject string, data interface{}) *EventPayload {
	return &EventPayload{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    source,
		Subject:   subject,
		Data:      data,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
		Version:   "1.0",
	}
}

// WithMetadata adds metadata to the event payload
func (e *EventPayload) WithMetadata(key string, value interface{}) *EventPayload {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// BeverageCreatedEvent represents a beverage created event
type BeverageCreatedEvent struct {
	BeverageID  string                 `json:"beverage_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Theme       string                 `json:"theme"`
	Ingredients []interface{}          `json:"ingredients"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TaskCreatedEvent represents a task created event
type TaskCreatedEvent struct {
	TaskID      string                 `json:"task_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	AssigneeID  string                 `json:"assignee_id"`
	ProjectID   string                 `json:"project_id"`
	Priority    string                 `json:"priority"`
	Status      string                 `json:"status"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationSentEvent represents a notification sent event
type NotificationSentEvent struct {
	NotificationID string                 `json:"notification_id"`
	Type           string                 `json:"type"`
	Channel        string                 `json:"channel"`
	Recipient      string                 `json:"recipient"`
	Subject        string                 `json:"subject"`
	Message        string                 `json:"message"`
	Status         string                 `json:"status"`
	SentAt         time.Time              `json:"sent_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// AIRequestCompletedEvent represents an AI request completed event
type AIRequestCompletedEvent struct {
	RequestID    string                 `json:"request_id"`
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
	Operation    string                 `json:"operation"`
	TokensUsed   int                    `json:"tokens_used"`
	Duration     time.Duration          `json:"duration"`
	Success      bool                   `json:"success"`
	Error        string                 `json:"error,omitempty"`
	CompletedAt  time.Time              `json:"completed_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}
