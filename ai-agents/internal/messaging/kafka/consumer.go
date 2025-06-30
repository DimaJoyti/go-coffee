package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go-coffee-ai-agents/internal/observability"
)

// Event type definitions for message handling
type kafkaBeverageCreatedEvent struct {
	BeverageID string `json:"beverage_id"`
	Name       string `json:"name"`
	Theme      string `json:"theme"`
	CreatedBy  string `json:"created_by"`
	CreatedAt  string `json:"created_at"`
}

type kafkaTaskCreatedEvent struct {
	TaskID     string `json:"task_id"`
	Title      string `json:"title"`
	AssigneeID string `json:"assignee_id"`
	Priority   string `json:"priority"`
	CreatedBy  string `json:"created_by"`
	CreatedAt  string `json:"created_at"`
}

// MessageHandler defines the interface for handling consumed messages
type MessageHandler interface {
	Handle(ctx context.Context, message *ConsumedMessage) error
	GetTopics() []string
	GetGroupID() string
}

// ConsumedMessage represents a message consumed from Kafka
type ConsumedMessage struct {
	Topic     string            `json:"topic"`
	Partition int               `json:"partition"`
	Offset    int64             `json:"offset"`
	Key       string            `json:"key"`
	Value     []byte            `json:"value"`
	Headers   map[string]string `json:"headers"`
	Timestamp time.Time         `json:"timestamp"`
	
	// Kafka message for committing
	kafkaMessage kafka.Message
}

// GetHeader returns a header value
func (m *ConsumedMessage) GetHeader(key string) string {
	return m.Headers[key]
}

// UnmarshalValue unmarshals the message value into the provided interface
func (m *ConsumedMessage) UnmarshalValue(v interface{}) error {
	return json.Unmarshal(m.Value, v)
}

// Consumer handles Kafka message consumption with observability
type Consumer struct {
	connectionManager *ConnectionManager
	logger            *observability.StructuredLogger
	metrics           *observability.MetricsCollector
	tracing           *observability.TracingHelper
	
	handlers map[string]MessageHandler
	readers  map[string]*kafka.Reader
	
	// Control channels
	stopChan chan struct{}
	doneChan chan struct{}
	wg       sync.WaitGroup
	
	// State
	running bool
	mutex   sync.RWMutex
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(
	connectionManager *ConnectionManager,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *Consumer {
	return &Consumer{
		connectionManager: connectionManager,
		logger:            logger,
		metrics:           metrics,
		tracing:           tracing,
		handlers:          make(map[string]MessageHandler),
		readers:           make(map[string]*kafka.Reader),
		stopChan:          make(chan struct{}),
		doneChan:          make(chan struct{}),
	}
}

// RegisterHandler registers a message handler for specific topics
func (c *Consumer) RegisterHandler(handler MessageHandler) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running {
		return fmt.Errorf("cannot register handler while consumer is running")
	}

	topics := handler.GetTopics()
	groupID := handler.GetGroupID()

	for _, topic := range topics {
		handlerKey := fmt.Sprintf("%s-%s", topic, groupID)
		c.handlers[handlerKey] = handler
		
		c.logger.Info("Registered message handler",
			"topic", topic,
			"group_id", groupID,
			"handler_type", fmt.Sprintf("%T", handler))
	}

	return nil
}

// Start starts the consumer and begins processing messages
func (c *Consumer) Start(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running {
		return fmt.Errorf("consumer is already running")
	}

	ctx, span := c.tracing.StartKafkaSpan(ctx, "START_CONSUMER", "kafka")
	defer span.End()

	c.logger.InfoContext(ctx, "Starting Kafka consumer",
		"handlers_count", len(c.handlers))

	// Create readers for each handler
	for _, handler := range c.handlers {
		topics := handler.GetTopics()
		groupID := handler.GetGroupID()

		for _, topic := range topics {
			readerKey := fmt.Sprintf("%s-%s", topic, groupID)
			reader := c.connectionManager.GetReader(topic, groupID)
			c.readers[readerKey] = reader

			// Start consumer goroutine for this topic
			c.wg.Add(1)
			go c.consumeMessages(ctx, topic, groupID, reader, handler)
		}
	}

	c.running = true
	c.tracing.RecordSuccess(span, "Kafka consumer started")
	c.logger.InfoContext(ctx, "Kafka consumer started successfully",
		"readers_count", len(c.readers))

	return nil
}

// Stop stops the consumer gracefully
func (c *Consumer) Stop(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.running {
		return nil
	}

	ctx, span := c.tracing.StartKafkaSpan(ctx, "STOP_CONSUMER", "kafka")
	defer span.End()

	c.logger.InfoContext(ctx, "Stopping Kafka consumer")

	// Signal all goroutines to stop
	close(c.stopChan)

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.logger.InfoContext(ctx, "All consumer goroutines stopped")
	case <-ctx.Done():
		c.logger.WarnContext(ctx, "Consumer stop timeout, some goroutines may still be running")
	}

	// Close all readers
	for readerKey, reader := range c.readers {
		if err := reader.Close(); err != nil {
			c.logger.ErrorContext(ctx, "Failed to close reader", err,
				"reader_key", readerKey)
		}
	}

	c.running = false
	c.tracing.RecordSuccess(span, "Kafka consumer stopped")
	c.logger.InfoContext(ctx, "Kafka consumer stopped successfully")

	return nil
}

// consumeMessages consumes messages from a specific topic
func (c *Consumer) consumeMessages(ctx context.Context, topic, groupID string, reader *kafka.Reader, handler MessageHandler) {
	defer c.wg.Done()

	c.logger.InfoContext(ctx, "Starting message consumption",
		"topic", topic,
		"group_id", groupID)

	for {
		select {
		case <-c.stopChan:
			c.logger.InfoContext(ctx, "Stopping message consumption",
				"topic", topic,
				"group_id", groupID)
			return
		default:
			// Read message with timeout
			readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			kafkaMessage, err := reader.ReadMessage(readCtx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					// Timeout is expected, continue
					continue
				}
				
				c.logger.ErrorContext(ctx, "Failed to read message", err,
					"topic", topic,
					"group_id", groupID)
				
				// Record error metric
				if c.metrics != nil {
					counters := c.metrics.GetCounters()
					if counters != nil {
						counters.KafkaMessagesError.Add(ctx, 1)
					}
				}
				
				// Wait before retrying
				time.Sleep(time.Second)
				continue
			}

			// Process the message
			if err := c.processMessage(ctx, kafkaMessage, handler); err != nil {
				c.logger.ErrorContext(ctx, "Failed to process message", err,
					"topic", topic,
					"group_id", groupID,
					"offset", kafkaMessage.Offset,
					"partition", kafkaMessage.Partition)
			}
		}
	}
}

// processMessage processes a single Kafka message
func (c *Consumer) processMessage(ctx context.Context, kafkaMessage kafka.Message, handler MessageHandler) error {
	// Extract trace context from headers
	ctx = c.extractTraceContext(ctx, kafkaMessage)
	
	ctx, span := c.tracing.StartKafkaSpan(ctx, "PROCESS_MESSAGE", kafkaMessage.Topic)
	defer span.End()

	start := time.Now()

	// Convert to consumed message
	consumedMessage := &ConsumedMessage{
		Topic:        kafkaMessage.Topic,
		Partition:    kafkaMessage.Partition,
		Offset:       kafkaMessage.Offset,
		Key:          string(kafkaMessage.Key),
		Value:        kafkaMessage.Value,
		Headers:      make(map[string]string),
		Timestamp:    kafkaMessage.Time,
		kafkaMessage: kafkaMessage,
	}

	// Extract headers
	for _, header := range kafkaMessage.Headers {
		consumedMessage.Headers[header.Key] = string(header.Value)
	}

	c.logger.DebugContext(ctx, "Processing message",
		"topic", kafkaMessage.Topic,
		"partition", kafkaMessage.Partition,
		"offset", kafkaMessage.Offset,
		"key", string(kafkaMessage.Key))

	// Handle the message
	err := handler.Handle(ctx, consumedMessage)
	duration := time.Since(start)

	// Record metrics
	if c.metrics != nil {
		counters := c.metrics.GetCounters()
		histograms := c.metrics.GetHistograms()
		if counters != nil && histograms != nil {
			if err != nil {
				counters.KafkaMessagesError.Add(ctx, 1)
			} else {
				counters.KafkaMessagesSuccess.Add(ctx, 1)
			}
			counters.KafkaMessagesTotal.Add(ctx, 1)
			histograms.KafkaConsumeDuration.Record(ctx, duration.Seconds())
		}
	}

	if err != nil {
		c.tracing.RecordError(span, err, "Message processing failed")
		c.logger.ErrorContext(ctx, "Message processing failed", err,
			"topic", kafkaMessage.Topic,
			"partition", kafkaMessage.Partition,
			"offset", kafkaMessage.Offset,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to process message: %w", err)
	}

	c.tracing.RecordSuccess(span, "Message processed successfully")
	c.logger.DebugContext(ctx, "Message processed successfully",
		"topic", kafkaMessage.Topic,
		"partition", kafkaMessage.Partition,
		"offset", kafkaMessage.Offset,
		"duration_ms", duration.Milliseconds())

	return nil
}

// extractTraceContext extracts trace context from message headers
func (c *Consumer) extractTraceContext(ctx context.Context, kafkaMessage kafka.Message) context.Context {
	var traceID, spanID, correlationID string

	for _, header := range kafkaMessage.Headers {
		switch header.Key {
		case "trace_id":
			traceID = string(header.Value)
		case "span_id":
			spanID = string(header.Value)
		case "correlation_id":
			correlationID = string(header.Value)
		}
	}

	// Inject trace context - these functions need to be implemented in observability package
	// For now, we'll store the values in context using standard context.WithValue
	if traceID != "" {
		ctx = context.WithValue(ctx, "trace_id", traceID)
	}
	if spanID != "" {
		ctx = context.WithValue(ctx, "span_id", spanID)
	}
	if correlationID != "" {
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}

	return ctx
}

// IsRunning returns whether the consumer is currently running
func (c *Consumer) IsRunning() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.running
}

// GetStats returns consumer statistics
func (c *Consumer) GetStats() ConsumerStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	connStats := c.connectionManager.GetConnectionStats()
	return ConsumerStats{
		ConnectionStats: connStats,
		HandlersCount:   len(c.handlers),
		ReadersCount:    len(c.readers),
		Running:         c.running,
	}
}

// ConsumerStats represents consumer statistics
type ConsumerStats struct {
	ConnectionStats ConnectionStats `json:"connection_stats"`
	HandlersCount   int             `json:"handlers_count"`
	ReadersCount    int             `json:"readers_count"`
	Running         bool            `json:"running"`
}

// BaseMessageHandler provides a base implementation for message handlers
type BaseMessageHandler struct {
	topics  []string
	groupID string
	logger  *observability.StructuredLogger
	tracing *observability.TracingHelper
}

// NewBaseMessageHandler creates a new base message handler
func NewBaseMessageHandler(
	topics []string,
	groupID string,
	logger *observability.StructuredLogger,
	tracing *observability.TracingHelper,
) *BaseMessageHandler {
	return &BaseMessageHandler{
		topics:  topics,
		groupID: groupID,
		logger:  logger,
		tracing: tracing,
	}
}

// GetTopics returns the topics this handler processes
func (h *BaseMessageHandler) GetTopics() []string {
	return h.topics
}

// GetGroupID returns the consumer group ID
func (h *BaseMessageHandler) GetGroupID() string {
	return h.groupID
}

// Handle provides a default implementation that logs the message
func (h *BaseMessageHandler) Handle(ctx context.Context, message *ConsumedMessage) error {
	h.logger.InfoContext(ctx, "Received message",
		"topic", message.Topic,
		"partition", message.Partition,
		"offset", message.Offset,
		"key", message.Key,
		"headers", message.Headers)
	return nil
}

// BeverageEventHandler handles beverage-related events
type BeverageEventHandler struct {
	*BaseMessageHandler
}

// NewBeverageEventHandler creates a new beverage event handler
func NewBeverageEventHandler(
	groupID string,
	logger *observability.StructuredLogger,
	tracing *observability.TracingHelper,
) *BeverageEventHandler {
	topics := []string{
		"beverage.created",
		"beverage.updated",
		"beverage.deleted",
	}
	
	return &BeverageEventHandler{
		BaseMessageHandler: NewBaseMessageHandler(topics, groupID, logger, tracing),
	}
}

// Handle processes beverage events
func (h *BeverageEventHandler) Handle(ctx context.Context, message *ConsumedMessage) error {
	ctx, span := h.tracing.StartKafkaSpan(ctx, "HANDLE_BEVERAGE_EVENT", message.Topic)
	defer span.End()

	eventType := message.GetHeader("event_type")
	entityID := message.GetHeader("entity_id")

	h.logger.InfoContext(ctx, "Processing beverage event",
		"event_type", eventType,
		"entity_id", entityID,
		"topic", message.Topic)

	switch eventType {
	case "beverage.created":
		var event kafkaBeverageCreatedEvent
		if err := message.UnmarshalValue(&event); err != nil {
			h.tracing.RecordError(span, err, "Failed to unmarshal beverage created event")
			return fmt.Errorf("failed to unmarshal beverage created event: %w", err)
		}
		return h.handleBeverageCreated(ctx, &event)
		
	case "beverage.updated":
		h.logger.InfoContext(ctx, "Beverage updated event received", "entity_id", entityID)
		
	case "beverage.deleted":
		h.logger.InfoContext(ctx, "Beverage deleted event received", "entity_id", entityID)
		
	default:
		h.logger.WarnContext(ctx, "Unknown beverage event type", "event_type", eventType)
	}

	h.tracing.RecordSuccess(span, "Beverage event processed")
	return nil
}

// handleBeverageCreated handles beverage created events
func (h *BeverageEventHandler) handleBeverageCreated(ctx context.Context, event *kafkaBeverageCreatedEvent) error {
	h.logger.InfoContext(ctx, "Processing beverage created event",
		"beverage_id", event.BeverageID,
		"name", event.Name,
		"theme", event.Theme,
		"created_by", event.CreatedBy)

	// Here you would implement business logic for beverage creation
	// For example:
	// - Send notifications
	// - Update analytics
	// - Trigger downstream processes

	return nil
}

// TaskEventHandler handles task-related events
type TaskEventHandler struct {
	*BaseMessageHandler
}

// NewTaskEventHandler creates a new task event handler
func NewTaskEventHandler(
	groupID string,
	logger *observability.StructuredLogger,
	tracing *observability.TracingHelper,
) *TaskEventHandler {
	topics := []string{
		"task.created",
		"task.updated",
		"task.completed",
	}
	
	return &TaskEventHandler{
		BaseMessageHandler: NewBaseMessageHandler(topics, groupID, logger, tracing),
	}
}

// Handle processes task events
func (h *TaskEventHandler) Handle(ctx context.Context, message *ConsumedMessage) error {
	ctx, span := h.tracing.StartKafkaSpan(ctx, "HANDLE_TASK_EVENT", message.Topic)
	defer span.End()

	eventType := message.GetHeader("event_type")
	entityID := message.GetHeader("entity_id")

	h.logger.InfoContext(ctx, "Processing task event",
		"event_type", eventType,
		"entity_id", entityID,
		"topic", message.Topic)

	switch eventType {
	case "task.created":
		var event kafkaTaskCreatedEvent
		if err := message.UnmarshalValue(&event); err != nil {
			h.tracing.RecordError(span, err, "Failed to unmarshal task created event")
			return fmt.Errorf("failed to unmarshal task created event: %w", err)
		}
		return h.handleTaskCreated(ctx, &event)
		
	case "task.updated":
		h.logger.InfoContext(ctx, "Task updated event received", "entity_id", entityID)
		
	case "task.completed":
		h.logger.InfoContext(ctx, "Task completed event received", "entity_id", entityID)
		
	default:
		h.logger.WarnContext(ctx, "Unknown task event type", "event_type", eventType)
	}

	h.tracing.RecordSuccess(span, "Task event processed")
	return nil
}

// handleTaskCreated handles task created events
func (h *TaskEventHandler) handleTaskCreated(ctx context.Context, event *kafkaTaskCreatedEvent) error {
	h.logger.InfoContext(ctx, "Processing task created event",
		"task_id", event.TaskID,
		"title", event.Title,
		"assignee_id", event.AssigneeID,
		"priority", event.Priority,
		"created_by", event.CreatedBy)

	// Here you would implement business logic for task creation
	// For example:
	// - Send notifications to assignee
	// - Update project metrics
	// - Trigger automation workflows

	return nil
}
