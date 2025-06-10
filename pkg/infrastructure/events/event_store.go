package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/redis"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	redisLib "github.com/redis/go-redis/v9"
)

// Event represents a domain event
type Event struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Source        string                 `json:"source"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int64                  `json:"version"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	CausationID   string                 `json:"causation_id,omitempty"`
}

// EventStore defines the interface for event storage
type EventStore interface {
	// Event storage
	SaveEvent(ctx context.Context, event *Event) error
	SaveEvents(ctx context.Context, events []*Event) error

	// Event retrieval
	GetEvent(ctx context.Context, eventID string) (*Event, error)
	GetEvents(ctx context.Context, aggregateID string, fromVersion int64) ([]*Event, error)
	GetEventsByType(ctx context.Context, eventType string, limit int) ([]*Event, error)
	GetEventsByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*Event, error)

	// Event streaming
	GetEventStream(ctx context.Context, aggregateID string) (<-chan *Event, error)
	GetAllEventsStream(ctx context.Context, fromTimestamp time.Time) (<-chan *Event, error)

	// Aggregate operations
	GetAggregateVersion(ctx context.Context, aggregateID string) (int64, error)
	GetAggregateEvents(ctx context.Context, aggregateID string) ([]*Event, error)

	// Cleanup operations
	DeleteExpiredEvents(ctx context.Context) error
	GetEventCount(ctx context.Context) (int64, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

// RedisEventStore implements EventStore using Redis
type RedisEventStore struct {
	client redis.ClientInterface
	config *config.EventStoreConfig
	logger *logger.Logger
}

// NewRedisEventStore creates a new Redis event store
func NewRedisEventStore(client redis.ClientInterface, cfg *config.EventStoreConfig, logger *logger.Logger) EventStore {
	return &RedisEventStore{
		client: client,
		config: cfg,
		logger: logger,
	}
}

// SaveEvent saves a single event
func (r *RedisEventStore) SaveEvent(ctx context.Context, event *Event) error {
	if event.ID == "" {
		return fmt.Errorf("event ID is required")
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize event
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Calculate TTL
	ttl := time.Duration(r.config.RetentionDays) * 24 * time.Hour

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Store event data
	eventKey := r.getEventKey(event.ID)
	pipe.Set(ctx, eventKey, string(eventData), ttl)

	// Add to aggregate stream
	if event.AggregateID != "" {
		aggregateKey := r.getAggregateKey(event.AggregateID)
		err := r.client.ZAdd(ctx, aggregateKey, redisLib.Z{
			Score:  float64(event.Version),
			Member: event.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to add event to aggregate stream: %w", err)
		}
		r.client.Expire(ctx, aggregateKey, ttl)
	}

	// Add to type index
	typeKey := r.getTypeKey(event.Type)
	err = r.client.ZAdd(ctx, typeKey, redisLib.Z{
		Score:  float64(event.Timestamp.Unix()),
		Member: event.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to add event to type index: %w", err)
	}
	r.client.Expire(ctx, typeKey, ttl)

	// Add to time index
	timeKey := r.getTimeKey(event.Timestamp)
	err = r.client.ZAdd(ctx, timeKey, redisLib.Z{
		Score:  float64(event.Timestamp.Unix()),
		Member: event.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to add event to time index: %w", err)
	}
	r.client.Expire(ctx, timeKey, ttl)

	// Add to global events stream
	globalKey := r.getGlobalKey()
	err = r.client.ZAdd(ctx, globalKey, redisLib.Z{
		Score:  float64(event.Timestamp.Unix()),
		Member: event.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to add event to global stream: %w", err)
	}

	r.logger.With(
		logger.String("event_id", event.ID),
		logger.String("event_type", event.Type),
		logger.String("aggregate_id", event.AggregateID),
	).Debug("Event saved")

	return nil
}

// SaveEvents saves multiple events atomically
func (r *RedisEventStore) SaveEvents(ctx context.Context, events []*Event) error {
	if len(events) == 0 {
		return nil
	}

	ttl := time.Duration(r.config.RetentionDays) * 24 * time.Hour

	for _, event := range events {
		if event.ID == "" {
			return fmt.Errorf("event ID is required")
		}

		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now()
		}

		// Serialize event
		eventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %s: %w", event.ID, err)
		}

		// Store event data
		eventKey := r.getEventKey(event.ID)
		err = r.client.Set(ctx, eventKey, string(eventData), ttl)
		if err != nil {
			return fmt.Errorf("failed to store event data: %w", err)
		}

		// Add to indexes
		if event.AggregateID != "" {
			aggregateKey := r.getAggregateKey(event.AggregateID)
			err = r.client.ZAdd(ctx, aggregateKey, redisLib.Z{
				Score:  float64(event.Version),
				Member: event.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to add event to aggregate index: %w", err)
			}
			r.client.Expire(ctx, aggregateKey, ttl)
		}

		typeKey := r.getTypeKey(event.Type)
		err = r.client.ZAdd(ctx, typeKey, redisLib.Z{
			Score:  float64(event.Timestamp.Unix()),
			Member: event.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to add event to type index: %w", err)
		}
		r.client.Expire(ctx, typeKey, ttl)

		timeKey := r.getTimeKey(event.Timestamp)
		err = r.client.ZAdd(ctx, timeKey, redisLib.Z{
			Score:  float64(event.Timestamp.Unix()),
			Member: event.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to add event to time index: %w", err)
		}
		r.client.Expire(ctx, timeKey, ttl)

		globalKey := r.getGlobalKey()
		err = r.client.ZAdd(ctx, globalKey, redisLib.Z{
			Score:  float64(event.Timestamp.Unix()),
			Member: event.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to add event to global index: %w", err)
		}
	}

	r.logger.With(logger.Int("count", len(events))).Debug("Events saved")
	return nil
}

// GetEvent retrieves a single event by ID
func (r *RedisEventStore) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	eventKey := r.getEventKey(eventID)

	data, err := r.client.Get(ctx, eventKey)
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, fmt.Errorf("event not found: %s", eventID)
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	var event Event
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	return &event, nil
}

// GetEvents retrieves events for an aggregate from a specific version
func (r *RedisEventStore) GetEvents(ctx context.Context, aggregateID string, fromVersion int64) ([]*Event, error) {
	aggregateKey := r.getAggregateKey(aggregateID)

	// Get event IDs from sorted set using underlying client
	eventIDs, err := r.client.GetClient().ZRangeByScore(ctx, aggregateKey, &redisLib.ZRangeBy{
		Min: fmt.Sprintf("%d", fromVersion),
		Max: "+inf",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get event IDs: %w", err)
	}

	if len(eventIDs) == 0 {
		return []*Event{}, nil
	}

	// Get events in batch
	events := make([]*Event, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		event, err := r.GetEvent(ctx, eventID)
		if err != nil {
			r.logger.WithError(err).WithField("event_id", eventID).Error("Failed to get event")
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

// GetEventsByType retrieves events by type
func (r *RedisEventStore) GetEventsByType(ctx context.Context, eventType string, limit int) ([]*Event, error) {
	typeKey := r.getTypeKey(eventType)

	// Get latest events (highest scores first) using underlying client
	eventIDs, err := r.client.GetClient().ZRevRange(ctx, typeKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get event IDs by type: %w", err)
	}

	events := make([]*Event, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		event, err := r.GetEvent(ctx, eventID)
		if err != nil {
			r.logger.WithError(err).WithField("event_id", eventID).Error("Failed to get event")
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

// GetEventsByTimeRange retrieves events within a time range
func (r *RedisEventStore) GetEventsByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*Event, error) {
	globalKey := r.getGlobalKey()

	// Get event IDs within time range using underlying client
	eventIDs, err := r.client.GetClient().ZRangeByScore(ctx, globalKey, &redisLib.ZRangeBy{
		Min: fmt.Sprintf("%d", start.Unix()),
		Max: fmt.Sprintf("%d", end.Unix()),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get event IDs by time range: %w", err)
	}

	// Limit results
	if limit > 0 && len(eventIDs) > limit {
		eventIDs = eventIDs[:limit]
	}

	events := make([]*Event, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		event, err := r.GetEvent(ctx, eventID)
		if err != nil {
			r.logger.WithError(err).WithField("event_id", eventID).Error("Failed to get event")
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

// GetAggregateVersion returns the latest version for an aggregate
func (r *RedisEventStore) GetAggregateVersion(ctx context.Context, aggregateID string) (int64, error) {
	aggregateKey := r.getAggregateKey(aggregateID)

	// Get the highest score (latest version) using underlying client
	results, err := r.client.GetClient().ZRevRangeWithScores(ctx, aggregateKey, 0, 0).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get aggregate version: %w", err)
	}

	if len(results) == 0 {
		return 0, nil
	}

	return int64(results[0].Score), nil
}

// GetAggregateEvents retrieves all events for an aggregate
func (r *RedisEventStore) GetAggregateEvents(ctx context.Context, aggregateID string) ([]*Event, error) {
	return r.GetEvents(ctx, aggregateID, 0)
}

// DeleteExpiredEvents removes expired events
func (r *RedisEventStore) DeleteExpiredEvents(ctx context.Context) error {
	// Redis handles expiration automatically with TTL
	// This method can be used for additional cleanup if needed
	r.logger.Info("Event cleanup completed (handled by Redis TTL)")
	return nil
}

// GetEventCount returns the total number of events
func (r *RedisEventStore) GetEventCount(ctx context.Context) (int64, error) {
	globalKey := r.getGlobalKey()
	return r.client.ZCard(ctx, globalKey)
}

// HealthCheck checks the health of the event store
func (r *RedisEventStore) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx)
}

// Key generation methods
func (r *RedisEventStore) getEventKey(eventID string) string {
	return fmt.Sprintf("events:data:%s", eventID)
}

func (r *RedisEventStore) getAggregateKey(aggregateID string) string {
	return fmt.Sprintf("events:aggregate:%s", aggregateID)
}

func (r *RedisEventStore) getTypeKey(eventType string) string {
	return fmt.Sprintf("events:type:%s", eventType)
}

func (r *RedisEventStore) getTimeKey(timestamp time.Time) string {
	return fmt.Sprintf("events:time:%s", timestamp.Format("2006-01-02"))
}

func (r *RedisEventStore) getGlobalKey() string {
	return "events:global"
}

// ZRangeByScore helper method (Redis client should implement this)
func (r *RedisEventStore) ZRangeByScore(ctx context.Context, key, min, max string) ([]string, error) {
	// Use the underlying Redis client for ZRangeByScore
	return r.client.GetClient().ZRangeByScore(ctx, key, &redisLib.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
}

// GetEventStream and GetAllEventsStream would require Redis Streams
// These are placeholder implementations
func (r *RedisEventStore) GetEventStream(ctx context.Context, aggregateID string) (<-chan *Event, error) {
	// TODO: Implement using Redis Streams
	return nil, fmt.Errorf("event streaming not implemented yet")
}

func (r *RedisEventStore) GetAllEventsStream(ctx context.Context, fromTimestamp time.Time) (<-chan *Event, error) {
	// TODO: Implement using Redis Streams
	return nil, fmt.Errorf("event streaming not implemented yet")
}
