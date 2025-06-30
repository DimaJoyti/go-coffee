package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/services"
)

// KafkaEventPublisher implements the EventPublisher interface using Kafka
type KafkaEventPublisher struct {
	writer *kafka.Writer
	logger services.Logger
	topics map[string]string
}

// NewKafkaEventPublisher creates a new Kafka event publisher
func NewKafkaEventPublisher(brokers string, logger services.Logger) (*KafkaEventPublisher, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		Async:        true,
		ErrorLogger:  kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Error(fmt.Sprintf("Kafka error: "+msg, args...), nil)
		}),
	}

	topics := map[string]string{
		"content.created":     "social-media-content-created",
		"content.updated":     "social-media-content-updated",
		"content.published":   "social-media-content-published",
		"content.scheduled":   "social-media-content-scheduled",
		"content.failed":      "social-media-content-failed",
		"campaign.created":    "social-media-campaign-created",
		"campaign.updated":    "social-media-campaign-updated",
		"campaign.started":    "social-media-campaign-started",
		"campaign.completed":  "social-media-campaign-completed",
		"analytics.updated":   "social-media-analytics-updated",
		"post.published":      "social-media-post-published",
		"post.failed":         "social-media-post-failed",
		"engagement.received": "social-media-engagement-received",
	}

	return &KafkaEventPublisher{
		writer: writer,
		logger: logger,
		topics: topics,
	}, nil
}

// PublishEvent publishes a single domain event
func (p *KafkaEventPublisher) PublishEvent(ctx context.Context, event services.DomainEvent) error {
	return p.PublishEvents(ctx, []services.DomainEvent{event})
}

// PublishEvents publishes multiple domain events
func (p *KafkaEventPublisher) PublishEvents(ctx context.Context, events []services.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(events))

	for _, event := range events {
		topic, exists := p.topics[event.GetEventType()]
		if !exists {
			p.logger.Warn("Unknown event type, using default topic", "event_type", event.GetEventType())
			topic = "social-media-events"
		}

		// Create event envelope
		envelope := EventEnvelope{
			ID:          uuid.New().String(),
			Type:        event.GetEventType(),
			AggregateID: event.GetAggregateID().String(),
			Version:     event.GetVersion(),
			Timestamp:   event.GetTimestamp(),
			Data:        event.GetEventData(),
			Metadata: map[string]interface{}{
				"source":      "social-media-content-agent",
				"schema_version": "1.0",
				"correlation_id": p.getCorrelationID(ctx),
			},
		}

		// Serialize event
		data, err := json.Marshal(envelope)
		if err != nil {
			p.logger.Error("Failed to marshal event", err, "event_type", event.GetEventType())
			continue
		}

		message := kafka.Message{
			Topic: topic,
			Key:   []byte(event.GetAggregateID().String()),
			Value: data,
			Headers: []kafka.Header{
				{Key: "event-type", Value: []byte(event.GetEventType())},
				{Key: "aggregate-id", Value: []byte(event.GetAggregateID().String())},
				{Key: "timestamp", Value: []byte(event.GetTimestamp().Format(time.RFC3339))},
				{Key: "source", Value: []byte("social-media-content-agent")},
			},
			Time: event.GetTimestamp(),
		}

		messages = append(messages, message)
	}

	if len(messages) == 0 {
		return fmt.Errorf("no valid messages to publish")
	}

	// Write messages to Kafka
	err := p.writer.WriteMessages(ctx, messages...)
	if err != nil {
		p.logger.Error("Failed to publish events to Kafka", err, "message_count", len(messages))
		return fmt.Errorf("failed to publish events: %w", err)
	}

	p.logger.Info("Successfully published events", "count", len(messages))
	return nil
}

// Close closes the Kafka writer
func (p *KafkaEventPublisher) Close() error {
	return p.writer.Close()
}

// getCorrelationID extracts correlation ID from context
func (p *KafkaEventPublisher) getCorrelationID(ctx context.Context) string {
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}

// EventEnvelope represents the standard event envelope
type EventEnvelope struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	AggregateID string                 `json:"aggregate_id"`
	Version     int                    `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Domain Events Implementation

// ContentCreatedEvent represents a content created event
type ContentCreatedEvent struct {
	id          string
	aggregateID uuid.UUID
	contentID   uuid.UUID
	brandID     uuid.UUID
	creatorID   uuid.UUID
	title       string
	contentType string
	timestamp   time.Time
	version     int
}

// NewContentCreatedEvent creates a new content created event
func NewContentCreatedEvent(contentID, brandID, creatorID uuid.UUID, title, contentType string) *ContentCreatedEvent {
	return &ContentCreatedEvent{
		id:          uuid.New().String(),
		aggregateID: contentID,
		contentID:   contentID,
		brandID:     brandID,
		creatorID:   creatorID,
		title:       title,
		contentType: contentType,
		timestamp:   time.Now(),
		version:     1,
	}
}

func (e *ContentCreatedEvent) GetEventType() string {
	return "content.created"
}

func (e *ContentCreatedEvent) GetAggregateID() uuid.UUID {
	return e.aggregateID
}

func (e *ContentCreatedEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"content_id":   e.contentID.String(),
		"brand_id":     e.brandID.String(),
		"creator_id":   e.creatorID.String(),
		"title":        e.title,
		"content_type": e.contentType,
	}
}

func (e *ContentCreatedEvent) GetTimestamp() time.Time {
	return e.timestamp
}

func (e *ContentCreatedEvent) GetVersion() int {
	return e.version
}

// ContentPublishedEvent represents a content published event
type ContentPublishedEvent struct {
	id          string
	aggregateID uuid.UUID
	contentID   uuid.UUID
	postID      uuid.UUID
	platform    string
	publishedAt time.Time
	timestamp   time.Time
	version     int
}

// NewContentPublishedEvent creates a new content published event
func NewContentPublishedEvent(contentID, postID uuid.UUID, platform string, publishedAt time.Time) *ContentPublishedEvent {
	return &ContentPublishedEvent{
		id:          uuid.New().String(),
		aggregateID: contentID,
		contentID:   contentID,
		postID:      postID,
		platform:    platform,
		publishedAt: publishedAt,
		timestamp:   time.Now(),
		version:     1,
	}
}

func (e *ContentPublishedEvent) GetEventType() string {
	return "content.published"
}

func (e *ContentPublishedEvent) GetAggregateID() uuid.UUID {
	return e.aggregateID
}

func (e *ContentPublishedEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"content_id":   e.contentID.String(),
		"post_id":      e.postID.String(),
		"platform":     e.platform,
		"published_at": e.publishedAt.Format(time.RFC3339),
	}
}

func (e *ContentPublishedEvent) GetTimestamp() time.Time {
	return e.timestamp
}

func (e *ContentPublishedEvent) GetVersion() int {
	return e.version
}

// CampaignStartedEvent represents a campaign started event
type CampaignStartedEvent struct {
	id          string
	aggregateID uuid.UUID
	campaignID  uuid.UUID
	brandID     uuid.UUID
	managerID   uuid.UUID
	name        string
	startDate   time.Time
	timestamp   time.Time
	version     int
}

// NewCampaignStartedEvent creates a new campaign started event
func NewCampaignStartedEvent(campaignID, brandID, managerID uuid.UUID, name string, startDate time.Time) *CampaignStartedEvent {
	return &CampaignStartedEvent{
		id:          uuid.New().String(),
		aggregateID: campaignID,
		campaignID:  campaignID,
		brandID:     brandID,
		managerID:   managerID,
		name:        name,
		startDate:   startDate,
		timestamp:   time.Now(),
		version:     1,
	}
}

func (e *CampaignStartedEvent) GetEventType() string {
	return "campaign.started"
}

func (e *CampaignStartedEvent) GetAggregateID() uuid.UUID {
	return e.aggregateID
}

func (e *CampaignStartedEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"campaign_id": e.campaignID.String(),
		"brand_id":    e.brandID.String(),
		"manager_id":  e.managerID.String(),
		"name":        e.name,
		"start_date":  e.startDate.Format(time.RFC3339),
	}
}

func (e *CampaignStartedEvent) GetTimestamp() time.Time {
	return e.timestamp
}

func (e *CampaignStartedEvent) GetVersion() int {
	return e.version
}

// AnalyticsUpdatedEvent represents an analytics updated event
type AnalyticsUpdatedEvent struct {
	id          string
	aggregateID uuid.UUID
	entityID    uuid.UUID
	entityType  string
	metrics     map[string]interface{}
	timestamp   time.Time
	version     int
}

// NewAnalyticsUpdatedEvent creates a new analytics updated event
func NewAnalyticsUpdatedEvent(entityID uuid.UUID, entityType string, metrics map[string]interface{}) *AnalyticsUpdatedEvent {
	return &AnalyticsUpdatedEvent{
		id:          uuid.New().String(),
		aggregateID: entityID,
		entityID:    entityID,
		entityType:  entityType,
		metrics:     metrics,
		timestamp:   time.Now(),
		version:     1,
	}
}

func (e *AnalyticsUpdatedEvent) GetEventType() string {
	return "analytics.updated"
}

func (e *AnalyticsUpdatedEvent) GetAggregateID() uuid.UUID {
	return e.aggregateID
}

func (e *AnalyticsUpdatedEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"entity_id":   e.entityID.String(),
		"entity_type": e.entityType,
		"metrics":     e.metrics,
	}
}

func (e *AnalyticsUpdatedEvent) GetTimestamp() time.Time {
	return e.timestamp
}

func (e *AnalyticsUpdatedEvent) GetVersion() int {
	return e.version
}

// EngagementReceivedEvent represents an engagement received event
type EngagementReceivedEvent struct {
	id             string
	aggregateID    uuid.UUID
	postID         uuid.UUID
	platform       string
	engagementType string
	userID         string
	timestamp      time.Time
	version        int
}

// NewEngagementReceivedEvent creates a new engagement received event
func NewEngagementReceivedEvent(postID uuid.UUID, platform, engagementType, userID string) *EngagementReceivedEvent {
	return &EngagementReceivedEvent{
		id:             uuid.New().String(),
		aggregateID:    postID,
		postID:         postID,
		platform:       platform,
		engagementType: engagementType,
		userID:         userID,
		timestamp:      time.Now(),
		version:        1,
	}
}

func (e *EngagementReceivedEvent) GetEventType() string {
	return "engagement.received"
}

func (e *EngagementReceivedEvent) GetAggregateID() uuid.UUID {
	return e.aggregateID
}

func (e *EngagementReceivedEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"post_id":         e.postID.String(),
		"platform":        e.platform,
		"engagement_type": e.engagementType,
		"user_id":         e.userID,
	}
}

func (e *EngagementReceivedEvent) GetTimestamp() time.Time {
	return e.timestamp
}

func (e *EngagementReceivedEvent) GetVersion() int {
	return e.version
}

// KafkaEventConsumer handles consuming events from Kafka
type KafkaEventConsumer struct {
	reader *kafka.Reader
	logger services.Logger
}

// NewKafkaEventConsumer creates a new Kafka event consumer
func NewKafkaEventConsumer(brokers, topic, groupID string, logger services.Logger) *KafkaEventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{brokers},
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		StartOffset: kafka.LastOffset,
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Error(fmt.Sprintf("Kafka consumer error: "+msg, args...), nil)
		}),
	})

	return &KafkaEventConsumer{
		reader: reader,
		logger: logger,
	}
}

// ConsumeEvents consumes events from Kafka
func (c *KafkaEventConsumer) ConsumeEvents(ctx context.Context, handler func(EventEnvelope) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("Failed to read message from Kafka", err)
				continue
			}

			var envelope EventEnvelope
			if err := json.Unmarshal(message.Value, &envelope); err != nil {
				c.logger.Error("Failed to unmarshal event envelope", err)
				continue
			}

			if err := handler(envelope); err != nil {
				c.logger.Error("Failed to handle event", err, "event_type", envelope.Type)
				continue
			}

			c.logger.Debug("Successfully processed event", "event_type", envelope.Type, "aggregate_id", envelope.AggregateID)
		}
	}
}

// Close closes the Kafka reader
func (c *KafkaEventConsumer) Close() error {
	return c.reader.Close()
}
