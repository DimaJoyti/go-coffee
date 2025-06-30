# Kafka Message Queue Integration

This package provides comprehensive Kafka integration for the Go Coffee AI agents, featuring producer/consumer patterns, event-driven architecture, observability, and resilient message processing.

## Overview

The Kafka messaging system implements:

1. **Connection Management**: Thread-safe Kafka connections with SASL/TLS support
2. **Producer/Consumer Pattern**: High-performance message publishing and consumption
3. **Event-Driven Architecture**: Domain events for beverage, task, and notification systems
4. **Observability Integration**: Full tracing, metrics, and logging for all operations
5. **Topic Management**: Automatic topic creation and management
6. **Resilient Processing**: Error handling, retries, and graceful degradation

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Kafka Manager   │    │ Connection      │    │ Topic Manager   │
│ • Coordination  │───▶│ Manager         │    │ • Topic Ops     │
│ • Lifecycle     │    │ • SASL/TLS      │    │ • Creation      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Producer        │    │ Consumer        │    │ Event Bus       │
│ • Publishing    │    │ • Consumption   │    │ • High-level    │
│ • Batching      │    │ • Handlers      │    │ • Domain Events │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Components

### 1. Connection Manager (`connection.go`)

Manages Kafka connections with security and observability:

```go
// Connection manager with SASL/TLS support
connectionManager := NewConnectionManager(kafkaConfig, logger, metrics, tracing)
if err := connectionManager.Initialize(ctx); err != nil {
    log.Fatal("Failed to initialize Kafka:", err)
}

// Get writer for topic
writer := connectionManager.GetWriter("beverage.created")

// Get reader for topic and group
reader := connectionManager.GetReader("beverage.created", "beverage-processor")
```

#### Security Features
- **SASL Authentication**: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
- **TLS Encryption**: Configurable TLS with certificate validation
- **Connection Pooling**: Efficient connection reuse and management

### 2. Producer (`producer.go`)

High-performance message publishing with observability:

```go
// Create producer
producer := NewProducer(connectionManager, logger, metrics, tracing)

// Publish single message
message := NewMessage("beverage.created", beverageData).
    WithKey(beverageID).
    WithHeader("event_type", "beverage.created").
    WithMetadata("source", "beverage-inventor-agent")

if err := producer.Publish(ctx, message); err != nil {
    return fmt.Errorf("failed to publish message: %w", err)
}

// Publish batch
messages := []*Message{message1, message2, message3}
if err := producer.PublishBatch(ctx, messages); err != nil {
    return fmt.Errorf("failed to publish batch: %w", err)
}
```

#### Producer Features
- **Message Batching**: Efficient batch publishing for high throughput
- **Automatic Tracing**: Trace context injection for distributed tracing
- **Retry Logic**: Built-in retry mechanisms with exponential backoff
- **Metrics Collection**: Comprehensive metrics for monitoring

### 3. Consumer (`consumer.go`)

Resilient message consumption with handler pattern:

```go
// Create consumer
consumer := NewConsumer(connectionManager, logger, metrics, tracing)

// Register message handler
handler := NewBeverageEventHandler("beverage-processor", logger, tracing)
if err := consumer.RegisterHandler(handler); err != nil {
    return fmt.Errorf("failed to register handler: %w", err)
}

// Start consuming
if err := consumer.Start(ctx); err != nil {
    return fmt.Errorf("failed to start consumer: %w", err)
}

// Graceful shutdown
if err := consumer.Stop(ctx); err != nil {
    log.Printf("Error stopping consumer: %v", err)
}
```

#### Consumer Features
- **Handler Pattern**: Clean separation of message processing logic
- **Concurrent Processing**: Multiple goroutines for parallel processing
- **Error Recovery**: Automatic error handling and recovery
- **Graceful Shutdown**: Clean shutdown with message completion

### 4. Kafka Manager (`manager.go`)

Central coordinator for all Kafka operations:

```go
// Initialize Kafka manager
manager := NewManager(kafkaConfig, logger, metrics, tracing)
if err := manager.Initialize(ctx); err != nil {
    log.Fatal("Failed to initialize Kafka manager:", err)
}

// Start consumer
if err := manager.StartConsumer(ctx); err != nil {
    log.Fatal("Failed to start consumer:", err)
}

// Publish domain events
beverageEvent := &BeverageCreatedEvent{
    BeverageID:  "123e4567-e89b-12d3-a456-426614174000",
    Name:        "Cosmic Coffee",
    Theme:       "Mars Base",
    CreatedBy:   "beverage-inventor-agent",
    CreatedAt:   time.Now(),
}

if err := manager.PublishBeverageCreated(ctx, beverageEvent); err != nil {
    return fmt.Errorf("failed to publish beverage event: %w", err)
}
```

## Event Types & Schemas

### Beverage Events

#### BeverageCreatedEvent
```go
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
```

### Task Events

#### TaskCreatedEvent
```go
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
```

### Notification Events

#### NotificationSentEvent
```go
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
```

### AI Events

#### AIRequestCompletedEvent
```go
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
```

## Message Handlers

### Custom Message Handler

```go
// Implement MessageHandler interface
type CustomEventHandler struct {
    *BaseMessageHandler
    businessLogic *BusinessService
}

func NewCustomEventHandler(groupID string, businessLogic *BusinessService) *CustomEventHandler {
    topics := []string{"custom.event.topic"}
    return &CustomEventHandler{
        BaseMessageHandler: NewBaseMessageHandler(topics, groupID, logger, tracing),
        businessLogic:      businessLogic,
    }
}

func (h *CustomEventHandler) Handle(ctx context.Context, message *ConsumedMessage) error {
    ctx, span := h.tracing.StartKafkaSpan(ctx, "HANDLE_CUSTOM_EVENT", message.Topic)
    defer span.End()

    // Extract event data
    var event CustomEvent
    if err := message.UnmarshalValue(&event); err != nil {
        h.tracing.RecordError(span, err, "Failed to unmarshal event")
        return fmt.Errorf("failed to unmarshal event: %w", err)
    }

    // Process business logic
    if err := h.businessLogic.ProcessEvent(ctx, &event); err != nil {
        h.tracing.RecordError(span, err, "Failed to process event")
        return fmt.Errorf("failed to process event: %w", err)
    }

    h.tracing.RecordSuccess(span, "Event processed successfully")
    return nil
}
```

## Configuration

### Kafka Configuration

```yaml
kafka:
  brokers:
    - localhost:9092
    - localhost:9093
    - localhost:9094
  client_id: beverage-inventor-agent
  group_id: beverage-inventor-group
  
  # Security
  enable_sasl: true
  sasl_mechanism: SCRAM-SHA-256
  sasl_username: ${KAFKA_USERNAME}
  sasl_password: ${KAFKA_PASSWORD}
  enable_tls: true
  tls_skip_verify: false
  
  # Performance
  connect_timeout: 10s
  read_timeout: 10s
  write_timeout: 10s
  batch_size: 1000
  batch_timeout: 100ms
  retry_max: 3
  retry_backoff: 100ms
  
  # Topics
  topics:
    beverage_created: beverage.created
    beverage_updated: beverage.updated
    task_created: task.created
    task_updated: task.updated
    notification_sent: notification.sent
    ai_request_completed: ai.request.completed
    system_event: system.event
```

### Environment Variables

```bash
# Kafka brokers
GOCOFFEE_KAFKA_BROKERS=broker1:9092,broker2:9092,broker3:9092

# Security
GOCOFFEE_KAFKA_ENABLE_SASL=true
GOCOFFEE_KAFKA_SASL_USERNAME=your-username
GOCOFFEE_KAFKA_SASL_PASSWORD=your-password
GOCOFFEE_KAFKA_ENABLE_TLS=true

# Performance tuning
GOCOFFEE_KAFKA_BATCH_SIZE=1000
GOCOFFEE_KAFKA_BATCH_TIMEOUT=100ms
GOCOFFEE_KAFKA_RETRY_MAX=3
```

## Usage Examples

### Publishing Events

```go
// Initialize Kafka manager
kafkaManager := kafka.GetGlobalManager()

// Publish beverage created event
beverageEvent := &kafka.BeverageCreatedEvent{
    BeverageID:  uuid.New().String(),
    Name:        "Cosmic Coffee",
    Description: "A stellar blend for space explorers",
    Theme:       "Mars Base",
    Ingredients: []interface{}{
        map[string]interface{}{
            "name":     "Coffee Beans",
            "quantity": "200g",
            "type":     "Base",
        },
    },
    CreatedBy: "beverage-inventor-agent",
    CreatedAt: time.Now(),
}

if err := kafkaManager.PublishBeverageCreated(ctx, beverageEvent); err != nil {
    return fmt.Errorf("failed to publish beverage event: %w", err)
}

// Publish task created event
taskEvent := &kafka.TaskCreatedEvent{
    TaskID:      uuid.New().String(),
    Title:       "Review Cosmic Coffee Recipe",
    Description: "Review and approve the new Cosmic Coffee recipe",
    AssigneeID:  "user-123",
    ProjectID:   "project-456",
    Priority:    "high",
    Status:      "open",
    CreatedBy:   "beverage-inventor-agent",
    CreatedAt:   time.Now(),
}

if err := kafkaManager.PublishTaskCreated(ctx, taskEvent); err != nil {
    return fmt.Errorf("failed to publish task event: %w", err)
}
```

### Event Bus Pattern

```go
// Get global event bus
eventBus := kafka.GetGlobalEventBus()

// Publish beverage event
if err := eventBus.PublishBeverageEvent(ctx, "created", beverageID, beverageEvent); err != nil {
    return fmt.Errorf("failed to publish beverage event: %w", err)
}

// Publish notification event
notificationEvent := &kafka.NotificationSentEvent{
    NotificationID: uuid.New().String(),
    Type:           "email",
    Channel:        "email",
    Recipient:      "user@example.com",
    Subject:        "New Beverage Created",
    Message:        "A new beverage 'Cosmic Coffee' has been created",
    Status:         "sent",
    SentAt:         time.Now(),
}

if err := eventBus.PublishNotificationEvent(ctx, notificationEvent); err != nil {
    return fmt.Errorf("failed to publish notification event: %w", err)
}
```

## Observability

### Metrics

Kafka operations are automatically instrumented with metrics:

- `kafka_messages_total`: Total messages processed
- `kafka_messages_success_total`: Successful message operations
- `kafka_messages_error_total`: Failed message operations
- `kafka_publish_duration_seconds`: Message publishing duration
- `kafka_consume_duration_seconds`: Message consumption duration

### Tracing

All Kafka operations are traced with distributed tracing:

```go
// Automatic span creation
ctx, span := tracing.StartKafkaSpan(ctx, "PUBLISH", topic)
defer span.End()

// Trace context propagation
message.WithHeader("trace_id", traceID)
message.WithHeader("span_id", spanID)
message.WithHeader("correlation_id", correlationID)
```

### Logging

Structured logging with trace correlation:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "message": "Message published successfully",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "message_id": "123e4567-e89b-12d3-a456-426614174000",
  "topic": "beverage.created",
  "key": "beverage-123",
  "duration_ms": 25
}
```

## Health Monitoring

```go
// Kafka health check
healthStatus, err := kafkaManager.HealthCheck(ctx)
if err != nil {
    log.Printf("Health check failed: %v", err)
    return
}

if !healthStatus.Healthy {
    log.Printf("Kafka unhealthy: %s", healthStatus.Error)
    return
}

// Get Kafka statistics
stats, err := kafkaManager.GetStatistics(ctx)
if err != nil {
    log.Printf("Failed to get statistics: %v", err)
    return
}

log.Printf("Kafka stats: %d topics, %d writers, %d readers",
    len(stats.Topics),
    stats.Producer.WritersCount,
    stats.Consumer.ReadersCount)
```

## Best Practices

### 1. Message Design
- Use consistent event schemas
- Include correlation IDs for tracing
- Add metadata for debugging
- Version your events

### 2. Error Handling
- Implement idempotent message processing
- Use dead letter queues for failed messages
- Log errors with context
- Monitor error rates

### 3. Performance
- Use message batching for high throughput
- Configure appropriate batch sizes
- Monitor consumer lag
- Scale consumers based on load

### 4. Security
- Enable SASL authentication
- Use TLS encryption
- Rotate credentials regularly
- Implement proper access controls

This Kafka integration provides a robust, observable, and scalable messaging foundation for the Go Coffee AI agent ecosystem, enabling event-driven architecture and reliable inter-service communication.
