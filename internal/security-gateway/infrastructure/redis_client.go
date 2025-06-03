package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RedisConfig represents Redis configuration
type RedisConfig struct {
	URL              string        `yaml:"url"`
	DB               int           `yaml:"db"`
	Password         string        `yaml:"password"`
	MaxRetries       int           `yaml:"max_retries" default:"3"`
	PoolSize         int           `yaml:"pool_size" default:"10"`
	MinIdleConns     int           `yaml:"min_idle_conns" default:"5"`
	DialTimeout      time.Duration `yaml:"dial_timeout" default:"5s"`
	ReadTimeout      time.Duration `yaml:"read_timeout" default:"3s"`
	WriteTimeout     time.Duration `yaml:"write_timeout" default:"3s"`
	PoolTimeout      time.Duration `yaml:"pool_timeout" default:"4s"`
	IdleTimeout      time.Duration `yaml:"idle_timeout" default:"5m"`
	IdleCheckFrequency time.Duration `yaml:"idle_check_frequency" default:"1m"`
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config *RedisConfig) (*redis.Client, error) {
	// Parse Redis URL if provided
	var opts *redis.Options
	var err error

	if config.URL != "" {
		opts, err = redis.ParseURL(config.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
	} else {
		opts = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	// Override with config values
	if config.DB != 0 {
		opts.DB = config.DB
	}
	if config.Password != "" {
		opts.Password = config.Password
	}
	if config.MaxRetries != 0 {
		opts.MaxRetries = config.MaxRetries
	}
	if config.PoolSize != 0 {
		opts.PoolSize = config.PoolSize
	}
	if config.MinIdleConns != 0 {
		opts.MinIdleConns = config.MinIdleConns
	}
	if config.DialTimeout != 0 {
		opts.DialTimeout = config.DialTimeout
	}
	if config.ReadTimeout != 0 {
		opts.ReadTimeout = config.ReadTimeout
	}
	if config.WriteTimeout != 0 {
		opts.WriteTimeout = config.WriteTimeout
	}
	if config.PoolTimeout != 0 {
		opts.PoolTimeout = config.PoolTimeout
	}
	if config.IdleTimeout != 0 {
		opts.IdleTimeout = config.IdleTimeout
	}
	if config.IdleCheckFrequency != 0 {
		opts.IdleCheckFrequency = config.IdleCheckFrequency
	}

	// Create client
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// RedisEventStore implements EventStore using Redis
type RedisEventStore struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisEventStore creates a new Redis event store
func NewRedisEventStore(client *redis.Client, logger *logger.Logger) *RedisEventStore {
	return &RedisEventStore{
		client: client,
		logger: logger,
	}
}

// Store stores a security event in Redis
func (r *RedisEventStore) Store(ctx context.Context, event *SecurityEvent) error {
	// Convert event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Store in Redis with expiration
	key := fmt.Sprintf("security:events:%s", event.ID)
	err = r.client.Set(ctx, key, eventJSON, 30*24*time.Hour).Err() // 30 days retention
	if err != nil {
		return fmt.Errorf("failed to store event in Redis: %w", err)
	}

	// Add to time-based index for efficient querying
	timeKey := fmt.Sprintf("security:events:time:%s", event.Timestamp.Format("2006-01-02"))
	err = r.client.ZAdd(ctx, timeKey, &redis.Z{
		Score:  float64(event.Timestamp.Unix()),
		Member: event.ID,
	}).Err()
	if err != nil {
		r.logger.WithError(err).Error("Failed to add event to time index")
	}

	// Add to type-based index
	typeKey := fmt.Sprintf("security:events:type:%s", event.EventType)
	err = r.client.ZAdd(ctx, typeKey, &redis.Z{
		Score:  float64(event.Timestamp.Unix()),
		Member: event.ID,
	}).Err()
	if err != nil {
		r.logger.WithError(err).Error("Failed to add event to type index")
	}

	// Add to severity-based index
	severityKey := fmt.Sprintf("security:events:severity:%s", event.Severity)
	err = r.client.ZAdd(ctx, severityKey, &redis.Z{
		Score:  float64(event.Timestamp.Unix()),
		Member: event.ID,
	}).Err()
	if err != nil {
		r.logger.WithError(err).Error("Failed to add event to severity index")
	}

	return nil
}

// Query queries security events from Redis
func (r *RedisEventStore) Query(ctx context.Context, filter EventFilter) ([]*SecurityEvent, error) {
	var eventIDs []string
	var err error

	// Build query based on filter
	if filter.StartTime != nil || filter.EndTime != nil {
		// Time-based query
		eventIDs, err = r.queryByTime(ctx, filter)
	} else if len(filter.EventTypes) > 0 {
		// Type-based query
		eventIDs, err = r.queryByType(ctx, filter)
	} else if len(filter.Severities) > 0 {
		// Severity-based query
		eventIDs, err = r.queryBySeverity(ctx, filter)
	} else {
		// Get all recent events
		eventIDs, err = r.queryRecent(ctx, filter)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query event IDs: %w", err)
	}

	// Fetch events by IDs
	events := make([]*SecurityEvent, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		event, err := r.getEventByID(ctx, eventID)
		if err != nil {
			r.logger.WithError(err).Error("Failed to get event by ID", map[string]any{
				"event_id": eventID,
			})
			continue
		}
		if event != nil {
			events = append(events, event)
		}
	}

	// Apply additional filters
	filteredEvents := r.applyFilters(events, filter)

	// Apply limit
	if filter.Limit > 0 && len(filteredEvents) > filter.Limit {
		filteredEvents = filteredEvents[:filter.Limit]
	}

	return filteredEvents, nil
}

// Delete deletes old events from Redis
func (r *RedisEventStore) Delete(ctx context.Context, olderThan time.Time) error {
	// Find events older than the specified time
	pattern := "security:events:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get event keys: %w", err)
	}

	deletedCount := 0
	for _, key := range keys {
		// Skip index keys
		if contains(key, ":time:") || contains(key, ":type:") || contains(key, ":severity:") {
			continue
		}

		// Get event to check timestamp
		eventJSON, err := r.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue // Key doesn't exist
			}
			r.logger.WithError(err).Error("Failed to get event for deletion check", map[string]any{
				"key": key,
			})
			continue
		}

		var event SecurityEvent
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			r.logger.WithError(err).Error("Failed to unmarshal event for deletion check", map[string]any{
				"key": key,
			})
			continue
		}

		// Delete if older than threshold
		if event.Timestamp.Before(olderThan) {
			err := r.client.Del(ctx, key).Err()
			if err != nil {
				r.logger.WithError(err).Error("Failed to delete old event", map[string]any{
					"key": key,
				})
			} else {
				deletedCount++
			}
		}
	}

	r.logger.Info("Deleted old security events", map[string]any{
		"deleted_count": deletedCount,
		"older_than":    olderThan,
	})

	return nil
}

// Helper methods

func (r *RedisEventStore) queryByTime(ctx context.Context, filter EventFilter) ([]string, error) {
	// Use time-based index
	var startScore, endScore string
	
	if filter.StartTime != nil {
		startScore = fmt.Sprintf("%d", filter.StartTime.Unix())
	} else {
		startScore = "-inf"
	}
	
	if filter.EndTime != nil {
		endScore = fmt.Sprintf("%d", filter.EndTime.Unix())
	} else {
		endScore = "+inf"
	}

	// Query from daily time indexes
	var allEventIDs []string
	
	// Determine date range
	startDate := time.Now().AddDate(0, 0, -30) // Default to last 30 days
	if filter.StartTime != nil {
		startDate = *filter.StartTime
	}
	
	endDate := time.Now()
	if filter.EndTime != nil {
		endDate = *filter.EndTime
	}

	// Query each day in the range
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		timeKey := fmt.Sprintf("security:events:time:%s", d.Format("2006-01-02"))
		eventIDs, err := r.client.ZRangeByScore(ctx, timeKey, &redis.ZRangeBy{
			Min: startScore,
			Max: endScore,
		}).Result()
		if err != nil && err != redis.Nil {
			return nil, fmt.Errorf("failed to query time index: %w", err)
		}
		allEventIDs = append(allEventIDs, eventIDs...)
	}

	return allEventIDs, nil
}

func (r *RedisEventStore) queryByType(ctx context.Context, filter EventFilter) ([]string, error) {
	var allEventIDs []string
	
	for _, eventType := range filter.EventTypes {
		typeKey := fmt.Sprintf("security:events:type:%s", eventType)
		eventIDs, err := r.client.ZRevRange(ctx, typeKey, 0, -1).Result()
		if err != nil && err != redis.Nil {
			return nil, fmt.Errorf("failed to query type index: %w", err)
		}
		allEventIDs = append(allEventIDs, eventIDs...)
	}

	return allEventIDs, nil
}

func (r *RedisEventStore) queryBySeverity(ctx context.Context, filter EventFilter) ([]string, error) {
	var allEventIDs []string
	
	for _, severity := range filter.Severities {
		severityKey := fmt.Sprintf("security:events:severity:%s", severity)
		eventIDs, err := r.client.ZRevRange(ctx, severityKey, 0, -1).Result()
		if err != nil && err != redis.Nil {
			return nil, fmt.Errorf("failed to query severity index: %w", err)
		}
		allEventIDs = append(allEventIDs, eventIDs...)
	}

	return allEventIDs, nil
}

func (r *RedisEventStore) queryRecent(ctx context.Context, filter EventFilter) ([]string, error) {
	// Get recent events from today's time index
	today := time.Now().Format("2006-01-02")
	timeKey := fmt.Sprintf("security:events:time:%s", today)
	
	limit := int64(100) // Default limit
	if filter.Limit > 0 {
		limit = int64(filter.Limit)
	}
	
	eventIDs, err := r.client.ZRevRange(ctx, timeKey, 0, limit-1).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to query recent events: %w", err)
	}

	return eventIDs, nil
}

func (r *RedisEventStore) getEventByID(ctx context.Context, eventID string) (*SecurityEvent, error) {
	key := fmt.Sprintf("security:events:%s", eventID)
	eventJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Event not found
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	var event SecurityEvent
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	return &event, nil
}

func (r *RedisEventStore) applyFilters(events []*SecurityEvent, filter EventFilter) []*SecurityEvent {
	var filtered []*SecurityEvent

	for _, event := range events {
		// Apply filters
		if filter.UserID != "" && event.UserID != filter.UserID {
			continue
		}
		if filter.IPAddress != "" && event.IPAddress != filter.IPAddress {
			continue
		}
		if filter.ThreatLevel != nil && event.ThreatLevel != *filter.ThreatLevel {
			continue
		}

		filtered = append(filtered, event)
	}

	return filtered
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || 
		   len(s) >= len(substr) && s[:len(substr)] == substr ||
		   (len(s) > len(substr) && s[len(s)/2-len(substr)/2:len(s)/2+len(substr)/2] == substr)
}
