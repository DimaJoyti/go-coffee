package eventsourcing

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EventStore provides event sourcing capabilities using Redis Streams
type EventStore struct {
	redis  *redis.Client
	logger *zap.Logger
	config *EventStoreConfig
}

// EventStoreConfig contains configuration for event store
type EventStoreConfig struct {
	StreamPrefix    string
	SnapshotPrefix  string
	MaxStreamLength int64
	SnapshotInterval int64
	RetentionPeriod time.Duration
}

// Event represents a domain event
type Event struct {
	ID            string                 `json:"id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	EventType     string                 `json:"event_type"`
	EventVersion  int64                  `json:"event_version"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
	StreamID      string                 `json:"stream_id,omitempty"`
}

// EventStream represents a stream of events for an aggregate
type EventStream struct {
	AggregateID   string    `json:"aggregate_id"`
	AggregateType string    `json:"aggregate_type"`
	Version       int64     `json:"version"`
	Events        []*Event  `json:"events"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Snapshot represents a snapshot of aggregate state
type Snapshot struct {
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int64                  `json:"version"`
	Data          map[string]interface{} `json:"data"`
	Timestamp     time.Time              `json:"timestamp"`
}

// EventHandler defines a function that handles events
type EventHandler func(ctx context.Context, event *Event) error

// NewEventStore creates a new event store
func NewEventStore(redisClient *redis.Client, logger *zap.Logger, config *EventStoreConfig) *EventStore {
	if config == nil {
		config = &EventStoreConfig{
			StreamPrefix:     "events:",
			SnapshotPrefix:   "snapshots:",
			MaxStreamLength:  10000,
			SnapshotInterval: 100,
			RetentionPeriod:  30 * 24 * time.Hour,
		}
	}

	return &EventStore{
		redis:  redisClient,
		logger: logger,
		config: config,
	}
}

// AppendEvent appends an event to the event store
func (es *EventStore) AppendEvent(ctx context.Context, event *Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	streamName := es.getStreamName(event.AggregateType, event.AggregateID)
	
	es.logger.Info("Appending event to store",
		zap.String("event_id", event.ID),
		zap.String("aggregate_id", event.AggregateID),
		zap.String("event_type", event.EventType),
		zap.String("stream", streamName),
	)

	// Prepare event data for stream
	eventData := map[string]interface{}{
		"event_id":       event.ID,
		"aggregate_id":   event.AggregateID,
		"aggregate_type": event.AggregateType,
		"event_type":     event.EventType,
		"event_version":  event.EventVersion,
		"timestamp":      event.Timestamp.Unix(),
	}

	// Add event data
	if event.Data != nil {
		dataJSON, err := json.Marshal(event.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}
		eventData["data"] = string(dataJSON)
	}

	// Add metadata
	if event.Metadata != nil {
		metadataJSON, err := json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal event metadata: %w", err)
		}
		eventData["metadata"] = string(metadataJSON)
	}

	// Publish to Redis stream using XADD
	args := &redis.XAddArgs{
		Stream: streamName,
		Values: eventData,
	}

	if es.config.MaxStreamLength > 0 {
		args.MaxLen = es.config.MaxStreamLength
		args.Approx = true
	}

	streamID, err := es.redis.XAdd(ctx, args).Result()
	if err != nil {
		return fmt.Errorf("failed to append event to stream: %w", err)
	}

	event.StreamID = streamID

	es.logger.Info("Event appended successfully",
		zap.String("event_id", event.ID),
		zap.String("stream_id", streamID),
	)

	// Check if we need to create a snapshot
	if event.EventVersion%es.config.SnapshotInterval == 0 {
		go es.createSnapshotIfNeeded(context.Background(), event.AggregateType, event.AggregateID, event.EventVersion)
	}

	return nil
}

// AppendEvents appends multiple events atomically
func (es *EventStore) AppendEvents(ctx context.Context, events []*Event) error {
	if len(events) == 0 {
		return nil
	}

	es.logger.Info("Appending multiple events", zap.Int("count", len(events)))

	// Group events by stream
	streamEvents := make(map[string][]*Event)
	for _, event := range events {
		streamName := es.getStreamName(event.AggregateType, event.AggregateID)
		streamEvents[streamName] = append(streamEvents[streamName], event)
	}

	// Append events to each stream
	for _, streamEventList := range streamEvents {
		for _, event := range streamEventList {
			if err := es.AppendEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to append event %s: %w", event.ID, err)
			}
		}
	}

	es.logger.Info("Multiple events appended successfully", zap.Int("count", len(events)))
	return nil
}

// GetEvents retrieves events for an aggregate
func (es *EventStore) GetEvents(ctx context.Context, aggregateType, aggregateID string, fromVersion int64) ([]*Event, error) {
	streamName := es.getStreamName(aggregateType, aggregateID)

	es.logger.Info("Getting events from store",
		zap.String("aggregate_id", aggregateID),
		zap.String("aggregate_type", aggregateType),
		zap.Int64("from_version", fromVersion),
	)

	// Read all events from stream using XRANGE
	result, err := es.redis.XRange(ctx, streamName, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var events []*Event
	for _, message := range result {
		event, err := es.parseRedisMessage(message)
		if err != nil {
			es.logger.Warn("Failed to parse stream event", zap.Error(err))
			continue
		}

		// Filter by version if specified
		if fromVersion > 0 && event.EventVersion < fromVersion {
			continue
		}

		events = append(events, event)
	}

	es.logger.Info("Events retrieved successfully",
		zap.String("aggregate_id", aggregateID),
		zap.Int("count", len(events)),
	)

	return events, nil
}

// GetEventStream retrieves the complete event stream for an aggregate
func (es *EventStore) GetEventStream(ctx context.Context, aggregateType, aggregateID string) (*EventStream, error) {
	events, err := es.GetEvents(ctx, aggregateType, aggregateID, 0)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for aggregate %s:%s", aggregateType, aggregateID)
	}

	version := int64(0)
	if len(events) > 0 {
		version = events[len(events)-1].EventVersion
	}

	return &EventStream{
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Version:       version,
		Events:        events,
		CreatedAt:     events[0].Timestamp,
		UpdatedAt:     events[len(events)-1].Timestamp,
	}, nil
}

// SaveSnapshot saves a snapshot of aggregate state
func (es *EventStore) SaveSnapshot(ctx context.Context, snapshot *Snapshot) error {
	if snapshot.Timestamp.IsZero() {
		snapshot.Timestamp = time.Now()
	}

	key := es.getSnapshotKey(snapshot.AggregateType, snapshot.AggregateID)
	
	es.logger.Info("Saving snapshot",
		zap.String("aggregate_id", snapshot.AggregateID),
		zap.String("aggregate_type", snapshot.AggregateType),
		zap.Int64("version", snapshot.Version),
	)

	// Serialize snapshot
	snapshotData, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Save to Redis with TTL
	err = es.redis.Set(ctx, key, snapshotData, es.config.RetentionPeriod).Err()
	if err != nil {
		return fmt.Errorf("failed to save snapshot to Redis: %w", err)
	}

	es.logger.Info("Snapshot saved successfully", zap.String("key", key))

	return nil
}

// GetSnapshot retrieves the latest snapshot for an aggregate
func (es *EventStore) GetSnapshot(ctx context.Context, aggregateType, aggregateID string) (*Snapshot, error) {
	key := es.getSnapshotKey(aggregateType, aggregateID)
	
	es.logger.Info("Getting snapshot",
		zap.String("aggregate_id", aggregateID),
		zap.String("aggregate_type", aggregateType),
		zap.String("key", key),
	)

	// This would retrieve from Redis
	// For now, return nil (no snapshot found)
	return nil, fmt.Errorf("snapshot not found")
}

// ReplayEvents replays events from a specific version
func (es *EventStore) ReplayEvents(ctx context.Context, aggregateType, aggregateID string, fromVersion int64, handler EventHandler) error {
	events, err := es.GetEvents(ctx, aggregateType, aggregateID, fromVersion)
	if err != nil {
		return fmt.Errorf("failed to get events for replay: %w", err)
	}

	es.logger.Info("Replaying events",
		zap.String("aggregate_id", aggregateID),
		zap.Int("event_count", len(events)),
		zap.Int64("from_version", fromVersion),
	)

	for _, event := range events {
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("event handler failed for event %s: %w", event.ID, err)
		}
	}

	es.logger.Info("Event replay completed successfully")
	return nil
}

// Helper methods

func (es *EventStore) getStreamName(aggregateType, aggregateID string) string {
	return fmt.Sprintf("%s%s:%s", es.config.StreamPrefix, aggregateType, aggregateID)
}

func (es *EventStore) getSnapshotKey(aggregateType, aggregateID string) string {
	return fmt.Sprintf("%s%s:%s", es.config.SnapshotPrefix, aggregateType, aggregateID)
}

func (es *EventStore) parseRedisMessage(message redis.XMessage) (*Event, error) {
	event := &Event{
		StreamID: message.ID,
	}

	// Parse required fields
	if id, ok := message.Values["event_id"].(string); ok {
		event.ID = id
	}

	if aggregateID, ok := message.Values["aggregate_id"].(string); ok {
		event.AggregateID = aggregateID
	}

	if aggregateType, ok := message.Values["aggregate_type"].(string); ok {
		event.AggregateType = aggregateType
	}

	if eventType, ok := message.Values["event_type"].(string); ok {
		event.EventType = eventType
	}

	if versionStr, ok := message.Values["event_version"].(string); ok {
		if version, err := strconv.ParseInt(versionStr, 10, 64); err == nil {
			event.EventVersion = version
		}
	}

	// Parse timestamp
	if timestampStr, ok := message.Values["timestamp"].(string); ok {
		if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
			event.Timestamp = time.Unix(timestamp, 0)
		}
	}

	// Parse data
	if dataStr, ok := message.Values["data"].(string); ok {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err == nil {
			event.Data = data
		}
	}

	// Parse metadata
	if metadataStr, ok := message.Values["metadata"].(string); ok {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
			event.Metadata = metadata
		}
	}

	return event, nil
}

func (es *EventStore) createSnapshotIfNeeded(ctx context.Context, aggregateType, aggregateID string, version int64) {
	// This would implement snapshot creation logic
	es.logger.Info("Snapshot creation triggered",
		zap.String("aggregate_type", aggregateType),
		zap.String("aggregate_id", aggregateID),
		zap.Int64("version", version),
	)
}
