package kafka

import (
	"context"
	"fmt"
	"time"

	"go-coffee-ai-agents/internal/config"
	"go-coffee-ai-agents/internal/observability"
)

// Manager manages Kafka operations including producers, consumers, and topics
type Manager struct {
	config            config.KafkaConfig
	connectionManager *ConnectionManager
	topicManager      *TopicManager
	producer          *Producer
	consumer          *Consumer
	logger            *observability.StructuredLogger
	metrics           *observability.MetricsCollector
	tracing           *observability.TracingHelper
}

// NewManager creates a new Kafka manager
func NewManager(
	config config.KafkaConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *Manager {
	return &Manager{
		config:  config,
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// Initialize initializes the Kafka manager and all its components
func (m *Manager) Initialize(ctx context.Context) error {
	ctx, span := m.tracing.StartKafkaSpan(ctx, "INITIALIZE", "kafka")
	defer span.End()

	m.logger.InfoContext(ctx, "Initializing Kafka manager",
		"brokers", m.config.Brokers,
		"client_id", m.config.ClientID)

	// Initialize connection manager
	m.connectionManager = NewConnectionManager(m.config, m.logger, m.metrics, m.tracing)
	if err := m.connectionManager.Initialize(ctx); err != nil {
		m.tracing.RecordError(span, err, "Failed to initialize connection manager")
		return fmt.Errorf("failed to initialize connection manager: %w", err)
	}

	// Initialize topic manager
	m.topicManager = NewTopicManager(m.connectionManager, m.logger, m.tracing)

	// Initialize producer
	m.producer = NewProducer(m.connectionManager, m.logger, m.metrics, m.tracing)

	// Initialize consumer
	m.consumer = NewConsumer(m.connectionManager, m.logger, m.metrics, m.tracing)

	// Create required topics
	if err := m.createRequiredTopics(ctx); err != nil {
		m.tracing.RecordError(span, err, "Failed to create required topics")
		return fmt.Errorf("failed to create required topics: %w", err)
	}

	// Register default message handlers
	if err := m.registerDefaultHandlers(); err != nil {
		m.tracing.RecordError(span, err, "Failed to register default handlers")
		return fmt.Errorf("failed to register default handlers: %w", err)
	}

	m.tracing.RecordSuccess(span, "Kafka manager initialized successfully")
	m.logger.InfoContext(ctx, "Kafka manager initialized successfully")

	return nil
}

// createRequiredTopics creates all required Kafka topics
func (m *Manager) createRequiredTopics(ctx context.Context) error {
	ctx, span := m.tracing.StartKafkaSpan(ctx, "CREATE_TOPICS", "kafka")
	defer span.End()

	topics := []struct {
		name               string
		partitions         int
		replicationFactor  int
	}{
		{m.config.Topics.BeverageCreated, 3, 1},
		{m.config.Topics.BeverageUpdated, 3, 1},
		{m.config.Topics.TaskCreated, 3, 1},
		{m.config.Topics.TaskUpdated, 3, 1},
		{m.config.Topics.NotificationSent, 3, 1},
		{m.config.Topics.AIRequestCompleted, 3, 1},
		{m.config.Topics.SystemEvent, 1, 1},
	}

	for _, topic := range topics {
		if err := m.topicManager.CreateTopic(ctx, topic.name, topic.partitions, topic.replicationFactor); err != nil {
			// Log warning but don't fail if topic already exists
			m.logger.WarnContext(ctx, "Failed to create topic (may already exist)", 
				"topic", topic.name,
				"error", err.Error())
		}
	}

	m.tracing.RecordSuccess(span, "Required topics created")
	return nil
}

// registerDefaultHandlers registers default message handlers
func (m *Manager) registerDefaultHandlers() error {
	// Register beverage event handler
	beverageHandler := NewBeverageEventHandler(
		m.config.GroupID+"-beverage",
		m.logger,
		m.tracing,
	)
	if err := m.consumer.RegisterHandler(beverageHandler); err != nil {
		return fmt.Errorf("failed to register beverage handler: %w", err)
	}

	// Register task event handler
	taskHandler := NewTaskEventHandler(
		m.config.GroupID+"-task",
		m.logger,
		m.tracing,
	)
	if err := m.consumer.RegisterHandler(taskHandler); err != nil {
		return fmt.Errorf("failed to register task handler: %w", err)
	}

	return nil
}

// StartConsumer starts the Kafka consumer
func (m *Manager) StartConsumer(ctx context.Context) error {
	return m.consumer.Start(ctx)
}

// StopConsumer stops the Kafka consumer
func (m *Manager) StopConsumer(ctx context.Context) error {
	return m.consumer.Stop(ctx)
}

// GetProducer returns the Kafka producer
func (m *Manager) GetProducer() *Producer {
	return m.producer
}

// GetConsumer returns the Kafka consumer
func (m *Manager) GetConsumer() *Consumer {
	return m.consumer
}

// GetTopicManager returns the topic manager
func (m *Manager) GetTopicManager() *TopicManager {
	return m.topicManager
}

// PublishBeverageCreated publishes a beverage created event
func (m *Manager) PublishBeverageCreated(ctx context.Context, event *BeverageCreatedEvent) error {
	return m.producer.PublishBeverageEvent(ctx, m.config.Topics.BeverageCreated, event.BeverageID, event)
}

// PublishBeverageUpdated publishes a beverage updated event
func (m *Manager) PublishBeverageUpdated(ctx context.Context, beverageID string, event interface{}) error {
	return m.producer.PublishBeverageEvent(ctx, m.config.Topics.BeverageUpdated, beverageID, event)
}

// PublishTaskCreated publishes a task created event
func (m *Manager) PublishTaskCreated(ctx context.Context, event *TaskCreatedEvent) error {
	return m.producer.PublishTaskEvent(ctx, m.config.Topics.TaskCreated, event.TaskID, event)
}

// PublishTaskUpdated publishes a task updated event
func (m *Manager) PublishTaskUpdated(ctx context.Context, taskID string, event interface{}) error {
	return m.producer.PublishTaskEvent(ctx, m.config.Topics.TaskUpdated, taskID, event)
}

// PublishNotificationSent publishes a notification sent event
func (m *Manager) PublishNotificationSent(ctx context.Context, event *NotificationSentEvent) error {
	return m.producer.PublishNotificationEvent(ctx, m.config.Topics.NotificationSent, event.NotificationID, event)
}

// PublishAIRequestCompleted publishes an AI request completed event
func (m *Manager) PublishAIRequestCompleted(ctx context.Context, event *AIRequestCompletedEvent) error {
	return m.producer.PublishEvent(ctx, m.config.Topics.AIRequestCompleted, event)
}

// PublishSystemEvent publishes a system event
func (m *Manager) PublishSystemEvent(ctx context.Context, event interface{}) error {
	return m.producer.PublishEvent(ctx, m.config.Topics.SystemEvent, event)
}

// HealthCheck performs a comprehensive health check
func (m *Manager) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	ctx, span := m.tracing.StartKafkaSpan(ctx, "HEALTH_CHECK", "kafka")
	defer span.End()

	start := time.Now()
	status := &HealthStatus{
		Timestamp: start,
		Healthy:   true,
		Details:   make(map[string]interface{}),
	}

	// Check connection manager
	if err := m.connectionManager.HealthCheck(ctx); err != nil {
		status.Healthy = false
		status.Error = err.Error()
		status.Details["connection_error"] = err.Error()
		m.tracing.RecordError(span, err, "Kafka health check failed")
		return status, nil
	}

	// Get connection statistics
	connStats := m.connectionManager.GetConnectionStats()
	status.Details["connection_stats"] = connStats

	// Get producer statistics
	if m.producer != nil {
		producerStats := m.producer.GetStats()
		status.Details["producer_stats"] = producerStats
	}

	// Get consumer statistics
	if m.consumer != nil {
		consumerStats := m.consumer.GetStats()
		status.Details["consumer_stats"] = consumerStats
	}

	// Test topic listing
	topics, err := m.topicManager.ListTopics(ctx)
	if err != nil {
		status.Details["topic_list_error"] = err.Error()
	} else {
		status.Details["available_topics"] = topics
		status.Details["topics_count"] = len(topics)
	}

	status.ResponseTime = time.Since(start)
	status.Details["response_time_ms"] = status.ResponseTime.Milliseconds()

	m.tracing.RecordSuccess(span, "Kafka health check completed")
	m.logger.DebugContext(ctx, "Kafka health check completed",
		"healthy", status.Healthy,
		"response_time_ms", status.ResponseTime.Milliseconds())

	return status, nil
}

// GetStatistics returns comprehensive Kafka statistics
func (m *Manager) GetStatistics(ctx context.Context) (*KafkaStatistics, error) {
	ctx, span := m.tracing.StartKafkaSpan(ctx, "GET_STATISTICS", "kafka")
	defer span.End()

	stats := &KafkaStatistics{
		Timestamp: time.Now(),
	}

	// Connection statistics
	stats.Connection = m.connectionManager.GetConnectionStats()

	// Producer statistics
	if m.producer != nil {
		stats.Producer = m.producer.GetStats()
	}

	// Consumer statistics
	if m.consumer != nil {
		stats.Consumer = m.consumer.GetStats()
	}

	// Topic statistics
	topics, err := m.topicManager.ListTopics(ctx)
	if err != nil {
		m.logger.WarnContext(ctx, "Failed to get topic statistics", "error", err.Error())
		stats.Topics = make([]string, 0)
	} else {
		stats.Topics = topics
	}

	m.tracing.RecordSuccess(span, "Kafka statistics retrieved")
	return stats, nil
}

// Close closes the Kafka manager and all its components
func (m *Manager) Close(ctx context.Context) error {
	ctx, span := m.tracing.StartKafkaSpan(ctx, "CLOSE", "kafka")
	defer span.End()

	m.logger.InfoContext(ctx, "Closing Kafka manager")

	var errors []error

	// Stop consumer
	if m.consumer != nil {
		if err := m.consumer.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop consumer: %w", err))
		}
	}

	// Close connection manager
	if m.connectionManager != nil {
		if err := m.connectionManager.Close(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection manager: %w", err))
		}
	}

	if len(errors) > 0 {
		err := fmt.Errorf("errors closing Kafka manager: %v", errors)
		m.tracing.RecordError(span, err, "Failed to close Kafka manager cleanly")
		m.logger.ErrorContext(ctx, "Failed to close Kafka manager cleanly", err)
		return err
	}

	m.tracing.RecordSuccess(span, "Kafka manager closed successfully")
	m.logger.InfoContext(ctx, "Kafka manager closed successfully")

	return nil
}

// HealthStatus represents the health status of Kafka
type HealthStatus struct {
	Timestamp    time.Time              `json:"timestamp"`
	Healthy      bool                   `json:"healthy"`
	Error        string                 `json:"error,omitempty"`
	ResponseTime time.Duration          `json:"response_time"`
	Details      map[string]interface{} `json:"details"`
}

// KafkaStatistics represents comprehensive Kafka statistics
type KafkaStatistics struct {
	Timestamp  time.Time       `json:"timestamp"`
	Connection ConnectionStats `json:"connection"`
	Producer   ProducerStats   `json:"producer"`
	Consumer   ConsumerStats   `json:"consumer"`
	Topics     []string        `json:"topics"`
}

// EventBus provides a high-level interface for event publishing
type EventBus struct {
	manager *Manager
	logger  *observability.StructuredLogger
	tracing *observability.TracingHelper
}

// NewEventBus creates a new event bus
func NewEventBus(manager *Manager) *EventBus {
	return &EventBus{
		manager: manager,
		logger:  manager.logger,
		tracing: manager.tracing,
	}
}

// PublishBeverageEvent publishes a beverage-related event
func (eb *EventBus) PublishBeverageEvent(ctx context.Context, eventType string, beverageID string, data interface{}) error {
	ctx, span := eb.tracing.StartKafkaSpan(ctx, "PUBLISH_BEVERAGE_EVENT", eventType)
	defer span.End()

	switch eventType {
	case "created":
		if event, ok := data.(*BeverageCreatedEvent); ok {
			return eb.manager.PublishBeverageCreated(ctx, event)
		}
	case "updated":
		return eb.manager.PublishBeverageUpdated(ctx, beverageID, data)
	default:
		err := fmt.Errorf("unknown beverage event type: %s", eventType)
		eb.tracing.RecordError(span, err, "Unknown beverage event type")
		return err
	}

	eb.tracing.RecordSuccess(span, "Beverage event published")
	return nil
}

// PublishTaskEvent publishes a task-related event
func (eb *EventBus) PublishTaskEvent(ctx context.Context, eventType string, taskID string, data interface{}) error {
	ctx, span := eb.tracing.StartKafkaSpan(ctx, "PUBLISH_TASK_EVENT", eventType)
	defer span.End()

	switch eventType {
	case "created":
		if event, ok := data.(*TaskCreatedEvent); ok {
			return eb.manager.PublishTaskCreated(ctx, event)
		}
	case "updated":
		return eb.manager.PublishTaskUpdated(ctx, taskID, data)
	default:
		err := fmt.Errorf("unknown task event type: %s", eventType)
		eb.tracing.RecordError(span, err, "Unknown task event type")
		return err
	}

	eb.tracing.RecordSuccess(span, "Task event published")
	return nil
}

// PublishNotificationEvent publishes a notification event
func (eb *EventBus) PublishNotificationEvent(ctx context.Context, event *NotificationSentEvent) error {
	return eb.manager.PublishNotificationSent(ctx, event)
}

// PublishAIEvent publishes an AI-related event
func (eb *EventBus) PublishAIEvent(ctx context.Context, event *AIRequestCompletedEvent) error {
	return eb.manager.PublishAIRequestCompleted(ctx, event)
}

// Global Kafka manager instance
var globalManager *Manager

// InitGlobalManager initializes the global Kafka manager
func InitGlobalManager(
	config config.KafkaConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) error {
	globalManager = NewManager(config, logger, metrics, tracing)
	return globalManager.Initialize(context.Background())
}

// GetGlobalManager returns the global Kafka manager
func GetGlobalManager() *Manager {
	return globalManager
}

// GetGlobalEventBus returns a global event bus instance
func GetGlobalEventBus() *EventBus {
	if globalManager == nil {
		return nil
	}
	return NewEventBus(globalManager)
}

// CloseGlobalManager closes the global Kafka manager
func CloseGlobalManager(ctx context.Context) error {
	if globalManager == nil {
		return nil
	}
	return globalManager.Close(ctx)
}
