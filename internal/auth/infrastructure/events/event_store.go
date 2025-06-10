package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// EventStore defines the interface for storing and retrieving domain events
type EventStore interface {
	SaveEvents(ctx context.Context, aggregateID string, events []domain.DomainEvent, expectedVersion int) error
	GetEvents(ctx context.Context, aggregateID string) ([]domain.DomainEvent, error)
	GetEventsSince(ctx context.Context, aggregateID string, version int) ([]domain.DomainEvent, error)
	GetAllEvents(ctx context.Context, limit int, offset int) ([]domain.DomainEvent, error)
	GetEventsByType(ctx context.Context, eventType string, limit int, offset int) ([]domain.DomainEvent, error)
}

// RedisEventStore implements EventStore using Redis
type RedisEventStore struct {
	client *redis.Client
	logger *logger.Logger
	prefix string
}

// NewRedisEventStore creates a new Redis event store
func NewRedisEventStore(client *redis.Client, logger *logger.Logger) EventStore {
	return &RedisEventStore{
		client: client,
		logger: logger,
		prefix: "auth:events:",
	}
}

// SaveEvents saves events to the event store
func (es *RedisEventStore) SaveEvents(ctx context.Context, aggregateID string, events []domain.DomainEvent, expectedVersion int) error {
	if len(events) == 0 {
		return nil
	}

	// Start transaction
	pipe := es.client.TxPipeline()

	// Check current version for optimistic concurrency control
	versionKey := es.getVersionKey(aggregateID)
	currentVersion, err := es.client.Get(ctx, versionKey).Int()
	if err != nil && err != redis.Nil {
		es.logger.ErrorWithFields("Failed to get current version", logger.Error(err), logger.String("aggregate_id", aggregateID))
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion != expectedVersion {
		return fmt.Errorf("concurrency conflict: expected version %d, got %d", expectedVersion, currentVersion)
	}

	// Save each event
	for i, event := range events {
		eventKey := es.getEventKey(aggregateID, currentVersion+i+1)
		
		// Serialize event
		eventData, err := json.Marshal(event)
		if err != nil {
			es.logger.ErrorWithFields("Failed to marshal event", logger.Error(err), logger.String("event_id", event.ID))
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		// Store event
		pipe.Set(ctx, eventKey, eventData, 0)

		// Add to aggregate event list
		aggregateEventsKey := es.getAggregateEventsKey(aggregateID)
		pipe.LPush(ctx, aggregateEventsKey, eventKey)

		// Add to global event stream
		globalEventsKey := es.getGlobalEventsKey()
		pipe.LPush(ctx, globalEventsKey, eventKey)

		// Add to event type index
		eventTypeKey := es.getEventTypeKey(event.Type)
		pipe.LPush(ctx, eventTypeKey, eventKey)

		// Add to timestamp index (sorted set)
		timestampKey := es.getTimestampKey()
		pipe.ZAdd(ctx, timestampKey, &redis.Z{
			Score:  float64(event.Timestamp.Unix()),
			Member: eventKey,
		})
	}

	// Update version
	newVersion := currentVersion + len(events)
	pipe.Set(ctx, versionKey, newVersion, 0)

	// Execute transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		es.logger.ErrorWithFields("Failed to save events", logger.Error(err), logger.String("aggregate_id", aggregateID))
		return fmt.Errorf("failed to save events: %w", err)
	}

	es.logger.InfoWithFields("Events saved successfully", 
		logger.String("aggregate_id", aggregateID),
		logger.Int("event_count", len(events)),
		logger.Int("new_version", newVersion))

	return nil
}

// GetEvents retrieves all events for an aggregate
func (es *RedisEventStore) GetEvents(ctx context.Context, aggregateID string) ([]domain.DomainEvent, error) {
	aggregateEventsKey := es.getAggregateEventsKey(aggregateID)
	
	// Get all event keys for the aggregate
	eventKeys, err := es.client.LRange(ctx, aggregateEventsKey, 0, -1).Result()
	if err != nil {
		es.logger.ErrorWithFields("Failed to get event keys", logger.Error(err), logger.String("aggregate_id", aggregateID))
		return nil, fmt.Errorf("failed to get event keys: %w", err)
	}

	// Reverse the order (LPUSH stores in reverse chronological order)
	for i, j := 0, len(eventKeys)-1; i < j; i, j = i+1, j-1 {
		eventKeys[i], eventKeys[j] = eventKeys[j], eventKeys[i]
	}

	return es.getEventsByKeys(ctx, eventKeys)
}

// GetEventsSince retrieves events for an aggregate since a specific version
func (es *RedisEventStore) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]domain.DomainEvent, error) {
	aggregateEventsKey := es.getAggregateEventsKey(aggregateID)
	
	// Get event keys since version (Redis lists are 0-indexed)
	start := int64(version)
	eventKeys, err := es.client.LRange(ctx, aggregateEventsKey, start, -1).Result()
	if err != nil {
		es.logger.ErrorWithFields("Failed to get event keys since version", 
			logger.Error(err), 
			logger.String("aggregate_id", aggregateID),
			logger.Int("version", version))
		return nil, fmt.Errorf("failed to get event keys since version: %w", err)
	}

	// Reverse the order
	for i, j := 0, len(eventKeys)-1; i < j; i, j = i+1, j-1 {
		eventKeys[i], eventKeys[j] = eventKeys[j], eventKeys[i]
	}

	return es.getEventsByKeys(ctx, eventKeys)
}

// GetAllEvents retrieves all events with pagination
func (es *RedisEventStore) GetAllEvents(ctx context.Context, limit int, offset int) ([]domain.DomainEvent, error) {
	globalEventsKey := es.getGlobalEventsKey()
	
	// Get event keys with pagination
	start := int64(offset)
	stop := int64(offset + limit - 1)
	eventKeys, err := es.client.LRange(ctx, globalEventsKey, start, stop).Result()
	if err != nil {
		es.logger.ErrorWithFields("Failed to get all event keys", logger.Error(err))
		return nil, fmt.Errorf("failed to get all event keys: %w", err)
	}

	return es.getEventsByKeys(ctx, eventKeys)
}

// GetEventsByType retrieves events by type with pagination
func (es *RedisEventStore) GetEventsByType(ctx context.Context, eventType string, limit int, offset int) ([]domain.DomainEvent, error) {
	eventTypeKey := es.getEventTypeKey(eventType)
	
	// Get event keys with pagination
	start := int64(offset)
	stop := int64(offset + limit - 1)
	eventKeys, err := es.client.LRange(ctx, eventTypeKey, start, stop).Result()
	if err != nil {
		es.logger.ErrorWithFields("Failed to get event keys by type", 
			logger.Error(err), 
			logger.String("event_type", eventType))
		return nil, fmt.Errorf("failed to get event keys by type: %w", err)
	}

	return es.getEventsByKeys(ctx, eventKeys)
}

// Helper methods

func (es *RedisEventStore) getEventsByKeys(ctx context.Context, eventKeys []string) ([]domain.DomainEvent, error) {
	if len(eventKeys) == 0 {
		return []domain.DomainEvent{}, nil
	}

	// Get all events in batch
	pipe := es.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(eventKeys))
	
	for i, key := range eventKeys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		es.logger.ErrorWithFields("Failed to get events batch", logger.Error(err))
		return nil, fmt.Errorf("failed to get events batch: %w", err)
	}

	// Deserialize events
	events := make([]domain.DomainEvent, 0, len(eventKeys))
	for i, cmd := range cmds {
		eventData, err := cmd.Result()
		if err != nil {
			es.logger.ErrorWithFields("Failed to get event data", 
				logger.Error(err), 
				logger.String("event_key", eventKeys[i]))
			continue
		}

		var event domain.DomainEvent
		if err := json.Unmarshal([]byte(eventData), &event); err != nil {
			es.logger.ErrorWithFields("Failed to unmarshal event", 
				logger.Error(err), 
				logger.String("event_key", eventKeys[i]))
			continue
		}

		events = append(events, event)
	}

	return events, nil
}

// Key generation methods

func (es *RedisEventStore) getEventKey(aggregateID string, version int) string {
	return fmt.Sprintf("%sevents:%s:%d", es.prefix, aggregateID, version)
}

func (es *RedisEventStore) getVersionKey(aggregateID string) string {
	return fmt.Sprintf("%sversion:%s", es.prefix, aggregateID)
}

func (es *RedisEventStore) getAggregateEventsKey(aggregateID string) string {
	return fmt.Sprintf("%saggregate:%s", es.prefix, aggregateID)
}

func (es *RedisEventStore) getGlobalEventsKey() string {
	return fmt.Sprintf("%sglobal", es.prefix)
}

func (es *RedisEventStore) getEventTypeKey(eventType string) string {
	return fmt.Sprintf("%stype:%s", es.prefix, eventType)
}

func (es *RedisEventStore) getTimestampKey() string {
	return fmt.Sprintf("%stimestamp", es.prefix)
}

// GetEventsByTimeRange retrieves events within a time range
func (es *RedisEventStore) GetEventsByTimeRange(ctx context.Context, start, end time.Time, limit int, offset int) ([]domain.DomainEvent, error) {
	timestampKey := es.getTimestampKey()
	
	// Get event keys by timestamp range
	eventKeys, err := es.client.ZRangeByScore(ctx, timestampKey, &redis.ZRangeBy{
		Min:    fmt.Sprintf("%d", start.Unix()),
		Max:    fmt.Sprintf("%d", end.Unix()),
		Offset: int64(offset),
		Count:  int64(limit),
	}).Result()
	
	if err != nil {
		es.logger.ErrorWithFields("Failed to get events by time range", logger.Error(err))
		return nil, fmt.Errorf("failed to get events by time range: %w", err)
	}

	return es.getEventsByKeys(ctx, eventKeys)
}
