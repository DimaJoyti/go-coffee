# Event Schema Management System

This package provides a comprehensive event schema management system for the Go Coffee AI Agents using Protocol Buffers for type-safe, versioned event handling.

## Overview

The event system provides:

- **Type-Safe Events**: Protocol Buffer definitions ensure compile-time type safety
- **Schema Registry**: Centralized management of event types and versions
- **Versioning Support**: Backward-compatible event evolution
- **Serialization**: Multiple formats (Protocol Buffers, JSON)
- **Event Manager**: High-level API for publishing and consuming events
- **Metrics & Observability**: Built-in tracking and monitoring

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Event Types   │    │ Schema Registry │    │  Event Manager  │
│  (Proto Files)  │───▶│   & Validator   │───▶│   & Publisher   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Serializer    │    │ Kafka Producer/ │
                       │ (Proto/JSON)    │    │    Consumer     │
                       └─────────────────┘    └─────────────────┘
```

## Event Categories

### 1. Beverage Events (`beverage_events.proto`)
- `BeverageCreatedEvent`: New beverage recipe created
- `BeverageUpdatedEvent`: Beverage recipe updated
- `BeverageStatusChangedEvent`: Status change (draft → approved → production)
- `RecipeRequestEvent`: Request for new recipe creation
- `IngredientDiscoveredEvent`: New ingredient discovered

### 2. Task Events (`task_events.proto`)
- `TaskCreatedEvent`: New task created in task management system
- `TaskUpdatedEvent`: Task details updated
- `TaskStatusChangedEvent`: Task status changed
- `TaskAssignedEvent`: Task assigned to user
- `TaskCompletedEvent`: Task marked as completed

### 3. Notification Events (`notification_events.proto`)
- `NotificationSentEvent`: Notification successfully sent
- `NotificationFailedEvent`: Notification delivery failed
- `SlackMessageSentEvent`: Slack message sent
- `EmailSentEvent`: Email notification sent
- `WebhookSentEvent`: Webhook notification sent
- `AlertTriggeredEvent`: System alert triggered
- `AlertResolvedEvent`: System alert resolved

### 4. Social Media Events (`social_media_events.proto`)
- `SocialMediaPostCreatedEvent`: Social media post created
- `SocialMediaPostPublishedEvent`: Post published to platform
- `SocialMediaEngagementEvent`: User engagement (likes, shares, comments)
- `ContentGenerationRequestEvent`: AI content generation requested
- `ContentGeneratedEvent`: AI content generated
- `InfluencerMentionEvent`: Brand mentioned by influencer
- `SocialMediaAnalyticsEvent`: Analytics data collected

### 5. Common Events (`common.proto`)
- `ErrorEvent`: System error occurred
- `HealthCheckEvent`: Service health check result
- `MetricEvent`: System metric measurement
- `AuditEvent`: Audit log entry
- `ConfigurationChangedEvent`: Configuration updated
- `ServiceStartedEvent`: Service startup
- `ServiceStoppedEvent`: Service shutdown

## Usage

### 1. Generate Protocol Buffer Code

```bash
# Generate Go code from proto files
make proto

# Or run the script directly
./scripts/generate-proto.sh
```

### 2. Initialize Event Registry

```go
import "go-coffee-ai-agents/internal/events"

// Initialize the event registry with all known event types
err := events.InitializeEventRegistry()
if err != nil {
    log.Fatal("Failed to initialize event registry:", err)
}
```

### 3. Create Event Manager

```go
// Create event registry and serializer
registry := events.NewEventRegistry()
serializer := events.NewEventSerializer(registry, events.FormatProtobuf)

// Create Kafka producer
producer := &kafka.Writer{
    Addr:     kafka.TCP("localhost:9092"),
    Balancer: &kafka.LeastBytes{},
}

// Create event manager
eventManager := events.NewEventManager(registry, serializer, producer, logger)
```

### 4. Publishing Events

```go
// Create a beverage created event
event := &events.BeverageCreatedEvent{
    BeverageId:    "beverage-123",
    Name:          "Cosmic Coffee Blend",
    Description:   "A stellar coffee experience",
    Theme:         "Mars Base",
    CreatedBy:     "user-456",
    CreatedAt:     timestamppb.Now(),
    EstimatedCost: 4.50,
}

// Publish the event
metadata := map[string]string{
    "correlation_id": "corr-123",
    "trace_id":      "trace-456",
}

err := eventManager.PublishEvent(
    ctx,
    "beverage.created",  // topic
    "beverage.created",  // event type
    "1.0",              // version
    event,              // message
    metadata,           // metadata
)
```

### 5. Consuming Events

```go
// Create typed event handler
handler := events.NewTypedEventHandler(logger)

// Register handlers for specific event types
handler.RegisterHandler("beverage.created", "1.0", func(ctx *events.EventContext, event proto.Message) error {
    beverageEvent := event.(*events.BeverageCreatedEvent)
    
    log.Printf("New beverage created: %s", beverageEvent.Name)
    
    // Process the event
    return processBeverageCreated(beverageEvent)
})

// Create Kafka reader
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"localhost:9092"},
    Topics:  []string{"beverage.created"},
    GroupID: "beverage-processor",
})

// Start consuming events
err := eventManager.ConsumeEvents(ctx, reader, handler)
```

### 6. Event Validation

```go
// Validate an event against its schema
err := events.ValidateEvent("beverage.created", "1.0", event)
if err != nil {
    log.Printf("Event validation failed: %v", err)
}
```

### 7. Event Versioning

```go
// Check version compatibility
versionManager := events.NewEventVersionManager(registry)
compatible, err := versionManager.IsCompatible("beverage.created", "1.0", "1.1")

// Migrate event to new version
newEvent, err := versionManager.MigrateEvent("beverage.created", "1.0", "1.1", oldEvent)
```

## Event Topic Mapping

Events are automatically routed to appropriate Kafka topics:

| Event Type | Kafka Topic |
|------------|-------------|
| `beverage.created` | `beverage.created` |
| `beverage.updated` | `beverage.updated` |
| `task.created` | `task.created` |
| `notification.sent` | `notifications` |
| `social_media.post_created` | `social_media.posts` |
| `system.error` | `system.errors` |

## Schema Evolution

### Adding Fields
- New optional fields can be added without breaking compatibility
- Use `optional` keyword in proto3 for optional fields

### Removing Fields
- Mark fields as deprecated before removal
- Maintain backward compatibility for at least one major version

### Changing Field Types
- Avoid changing field types as it breaks compatibility
- Create new fields with new types and deprecate old ones

### Version Numbering
- Use semantic versioning: `MAJOR.MINOR.PATCH`
- Increment MAJOR for breaking changes
- Increment MINOR for backward-compatible additions
- Increment PATCH for bug fixes

## Best Practices

### 1. Event Design
- Keep events immutable and focused on a single business event
- Include all necessary context in the event
- Use descriptive field names
- Add correlation and trace IDs for observability

### 2. Schema Management
- Always validate events before publishing
- Register all event types in the schema registry
- Document event schemas and their purpose
- Plan for schema evolution from the beginning

### 3. Error Handling
- Handle deserialization errors gracefully
- Implement retry mechanisms for transient failures
- Log detailed error information for debugging
- Use dead letter queues for failed messages

### 4. Performance
- Use Protocol Buffers for high-throughput scenarios
- Batch events when possible
- Monitor event processing latency
- Implement backpressure mechanisms

### 5. Observability
- Include correlation IDs in all events
- Track event processing metrics
- Implement distributed tracing
- Monitor schema registry health

## Development Workflow

1. **Define Event Schema**: Create or update `.proto` files
2. **Generate Code**: Run `make proto` to generate Go code
3. **Register Events**: Add event registration in `init.go`
4. **Implement Handlers**: Create event handlers for new event types
5. **Test**: Write unit and integration tests
6. **Deploy**: Deploy with proper schema migration

## Testing

```go
// Test event serialization/deserialization
func TestEventSerialization(t *testing.T) {
    registry := events.NewEventRegistry()
    serializer := events.NewEventSerializer(registry, events.FormatProtobuf)
    
    event := &events.BeverageCreatedEvent{
        BeverageId: "test-123",
        Name:       "Test Beverage",
    }
    
    // Serialize
    data, err := serializer.SerializeEvent("beverage.created", "1.0", event, nil)
    assert.NoError(t, err)
    
    // Deserialize
    eventType, version, deserializedEvent, metadata, err := serializer.DeserializeEvent(data)
    assert.NoError(t, err)
    assert.Equal(t, "beverage.created", eventType)
    assert.Equal(t, "1.0", version)
}
```

## Monitoring

The event system provides built-in metrics:

- Events published/consumed per type
- Processing latency per event type
- Error rates and types
- Schema registry health
- Kafka consumer lag

Access metrics through the `EventMetrics` interface or integrate with Prometheus/Grafana for visualization.
