package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

// Type aliases to avoid duplication with monitoring package
// These will be replaced with proper imports once the monitoring package is fixed

// SecurityEvent represents a security event (local definition for Redis operations)
type SecurityEvent struct {
	ID          string                 `json:"id"`
	EventType   string                 `json:"event_type"`
	Severity    string                 `json:"severity"`
	UserID      string                 `json:"user_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	ThreatLevel string                 `json:"threat_level"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Source      string                 `json:"source,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Description string                 `json:"description,omitempty"`
	Mitigated   bool                   `json:"mitigated"`
}

// EventFilter represents filters for querying security events
type EventFilter struct {
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	EventTypes   []string   `json:"event_types,omitempty"`
	Severities   []string   `json:"severities,omitempty"`
	UserID       string     `json:"user_id,omitempty"`
	IPAddress    string     `json:"ip_address,omitempty"`
	ThreatLevel  *string    `json:"threat_level,omitempty"`
	Limit        int        `json:"limit,omitempty"`
}

// SecurityAlert represents a security alert (local definition for Redis operations)
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	EventIDs    []string               `json:"event_ids,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	AlertType   string                 `json:"alert_type"`
	AssignedTo  string                 `json:"assigned_to,omitempty"`
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

// Store stores a security event in Redis (implements monitoring.EventStore)
func (r *RedisEventStore) Store(ctx context.Context, event interface{}) error {
	// Convert to our internal SecurityEvent type
	var secEvent *SecurityEvent
	switch e := event.(type) {
	case *SecurityEvent:
		secEvent = e
	default:
		return fmt.Errorf("unsupported event type: %T", event)
	}

	// Convert event to JSON
	eventJSON, err := json.Marshal(secEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Store in Redis with expiration
	key := fmt.Sprintf("security:events:%s", secEvent.ID)
	err = r.client.Set(ctx, key, eventJSON, 30*24*time.Hour).Err() // 30 days retention
	if err != nil {
		return fmt.Errorf("failed to store event in Redis: %w", err)
	}

	// Add to time-based index for efficient querying
	timeKey := fmt.Sprintf("security:events:time:%s", secEvent.Timestamp.Format("2006-01-02"))
	err = r.client.ZAdd(ctx, timeKey, &redis.Z{
		Score:  float64(secEvent.Timestamp.Unix()),
		Member: secEvent.ID,
	}).Err()
	if err != nil {
		r.logger.WithError(err).Error("Failed to add event to time index", map[string]any{
			"event_id": secEvent.ID,
		})
	}

	// Add to type-based index
	typeKey := fmt.Sprintf("security:events:type:%s", secEvent.EventType)
	err = r.client.ZAdd(ctx, typeKey, &redis.Z{
		Score:  float64(secEvent.Timestamp.Unix()),
		Member: secEvent.ID,
	}).Err()
	if err != nil {
		r.logger.WithError(err).Error("Failed to add event to type index", map[string]any{
			"event_id": secEvent.ID,
		})
	}

	// Add to severity-based index
	severityKey := fmt.Sprintf("security:events:severity:%s", secEvent.Severity)
	err = r.client.ZAdd(ctx, severityKey, &redis.Z{
		Score:  float64(secEvent.Timestamp.Unix()),
		Member: secEvent.ID,
	}).Err()
	if err != nil {
		r.logger.WithError(err).Error("Failed to add event to severity index", map[string]any{
			"event_id": secEvent.ID,
		})
	}

	return nil
}

// Query queries security events from Redis (implements monitoring.EventStore)
func (r *RedisEventStore) Query(ctx context.Context, filter interface{}) (interface{}, error) {
	// Convert filter to our internal EventFilter type
	var eventFilter EventFilter
	switch f := filter.(type) {
	case EventFilter:
		eventFilter = f
	default:
		return nil, fmt.Errorf("unsupported filter type: %T", filter)
	}

	var eventIDs []string
	var err error

	// Build query based on filter
	if eventFilter.StartTime != nil || eventFilter.EndTime != nil {
		// Time-based query
		eventIDs, err = r.queryByTime(ctx, eventFilter)
	} else if len(eventFilter.EventTypes) > 0 {
		// Type-based query
		eventIDs, err = r.queryByType(ctx, eventFilter)
	} else if len(eventFilter.Severities) > 0 {
		// Severity-based query
		eventIDs, err = r.queryBySeverity(ctx, eventFilter)
	} else {
		// Get all recent events
		eventIDs, err = r.queryRecent(ctx, eventFilter)
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
	filteredEvents := r.applyFilters(events, eventFilter)

	// Apply limit
	if eventFilter.Limit > 0 && len(filteredEvents) > eventFilter.Limit {
		filteredEvents = filteredEvents[:eventFilter.Limit]
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
		"older_than":    olderThan.Format(time.RFC3339),
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

// RedisAlertManager implements AlertManager using Redis
type RedisAlertManager struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisAlertManager creates a new Redis alert manager
func NewRedisAlertManager(client *redis.Client, logger *logger.Logger) *RedisAlertManager {
	return &RedisAlertManager{
		client: client,
		logger: logger,
	}
}

// CreateAlert creates a new security alert in Redis
func (r *RedisAlertManager) CreateAlert(ctx context.Context, alert *SecurityAlert) error {
	// Convert alert to JSON
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	// Store in Redis
	key := fmt.Sprintf("security:alerts:%s", alert.ID)
	err = r.client.Set(ctx, key, alertJSON, 0).Err() // No expiration for alerts
	if err != nil {
		return fmt.Errorf("failed to store alert in Redis: %w", err)
	}

	// Add to active alerts set if not resolved
	if alert.Status != "resolved" {
		err = r.client.SAdd(ctx, "security:alerts:active", alert.ID).Err()
		if err != nil {
			r.logger.WithError(err).Error("Failed to add alert to active set", map[string]any{
				"alert_id": alert.ID,
			})
		}
	}

	// Add to severity-based index
	severityKey := fmt.Sprintf("security:alerts:severity:%s", alert.Severity)
	err = r.client.ZAdd(ctx, severityKey, &redis.Z{
		Score:  float64(alert.CreatedAt.Unix()),
		Member: alert.ID,
	}).Err()
	if err != nil {
		r.logger.WithError(err).Error("Failed to add alert to severity index", map[string]any{
			"alert_id": alert.ID,
		})
	}

	return nil
}

// UpdateAlert updates an existing security alert in Redis
func (r *RedisAlertManager) UpdateAlert(ctx context.Context, alertID string, updates map[string]interface{}) error {
	// Get existing alert
	key := fmt.Sprintf("security:alerts:%s", alertID)
	alertJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("alert not found: %s", alertID)
		}
		return fmt.Errorf("failed to get alert: %w", err)
	}

	var alert SecurityAlert
	if err := json.Unmarshal([]byte(alertJSON), &alert); err != nil {
		return fmt.Errorf("failed to unmarshal alert: %w", err)
	}

	// Apply updates
	alert.UpdatedAt = time.Now()

	if status, ok := updates["status"].(string); ok {
		alert.Status = status
		if status == "resolved" {
			now := time.Now()
			alert.ResolvedAt = &now
			// Remove from active alerts
			r.client.SRem(ctx, "security:alerts:active", alertID)
		}
	}

	if title, ok := updates["title"].(string); ok {
		alert.Title = title
	}

	if description, ok := updates["description"].(string); ok {
		alert.Description = description
	}

	if severity, ok := updates["severity"].(string); ok {
		alert.Severity = severity
	}

	// Store updated alert
	updatedJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal updated alert: %w", err)
	}

	err = r.client.Set(ctx, key, updatedJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to update alert in Redis: %w", err)
	}

	return nil
}

// GetActiveAlerts retrieves all active security alerts from Redis
func (r *RedisAlertManager) GetActiveAlerts(ctx context.Context) ([]*SecurityAlert, error) {
	// Get active alert IDs
	alertIDs, err := r.client.SMembers(ctx, "security:alerts:active").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active alert IDs: %w", err)
	}

	// Fetch alerts
	alerts := make([]*SecurityAlert, 0, len(alertIDs))
	for _, alertID := range alertIDs {
		alert, err := r.getAlertByID(ctx, alertID)
		if err != nil {
			r.logger.WithError(err).Error("Failed to get alert by ID", map[string]any{
				"alert_id": alertID,
			})
			continue
		}
		if alert != nil {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// ResolveAlert resolves a security alert
func (r *RedisAlertManager) ResolveAlert(ctx context.Context, alertID string, reason string) error {
	updates := map[string]interface{}{
		"status": "resolved",
	}

	if reason != "" {
		updates["resolution_reason"] = reason
	}

	return r.UpdateAlert(ctx, alertID, updates)
}

// getAlertByID retrieves an alert by ID
func (r *RedisAlertManager) getAlertByID(ctx context.Context, alertID string) (*SecurityAlert, error) {
	key := fmt.Sprintf("security:alerts:%s", alertID)
	alertJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Alert not found
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	var alert SecurityAlert
	if err := json.Unmarshal([]byte(alertJSON), &alert); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alert: %w", err)
	}

	return &alert, nil
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
