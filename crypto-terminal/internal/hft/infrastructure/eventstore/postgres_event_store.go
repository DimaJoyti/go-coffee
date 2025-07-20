package eventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/repositories"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PostgresEventStore implements EventStore using PostgreSQL
type PostgresEventStore struct {
	db     *sql.DB
	tracer trace.Tracer
}

// NewPostgresEventStore creates a new PostgreSQL event store
func NewPostgresEventStore(db *sql.DB) repositories.EventStore {
	return &PostgresEventStore{
		db:     db,
		tracer: otel.Tracer("hft.infrastructure.event_store"),
	}
}

// SaveEvents saves domain events to the event store
func (es *PostgresEventStore) SaveEvents(ctx context.Context, aggregateID string, events []repositories.DomainEvent) error {
	ctx, span := es.tracer.Start(ctx, "PostgresEventStore.SaveEvents")
	defer span.End()

	span.SetAttributes(
		attribute.String("aggregate_id", aggregateID),
		attribute.Int("events_count", len(events)),
	)

	if len(events) == 0 {
		return nil
	}

	// Start transaction for atomic event saving
	tx, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current version for the aggregate
	currentVersion, err := es.getCurrentVersion(ctx, tx, aggregateID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Insert events with proper versioning
	query := `
		INSERT INTO hft_events (
			id, aggregate_id, event_type, event_data, timestamp, version, stream_name
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	for i, event := range events {
		eventID := uuid.New().String()
		version := currentVersion + i + 1
		streamName := fmt.Sprintf("%s-stream", aggregateID)

		eventDataJSON, err := json.Marshal(event.EventData)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		_, err = tx.ExecContext(ctx, query,
			eventID,
			aggregateID,
			event.EventType,
			eventDataJSON,
			event.Timestamp,
			version,
			streamName,
		)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to save event %s: %w", event.EventType, err)
		}

		// Update the event with the assigned ID and version
		events[i].ID = eventID
		events[i].Version = version
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	span.SetAttributes(
		attribute.Int("final_version", currentVersion+len(events)),
	)

	return nil
}

// GetEvents retrieves all events for an aggregate
func (es *PostgresEventStore) GetEvents(ctx context.Context, aggregateID string) ([]repositories.DomainEvent, error) {
	ctx, span := es.tracer.Start(ctx, "PostgresEventStore.GetEvents")
	defer span.End()

	span.SetAttributes(attribute.String("aggregate_id", aggregateID))

	query := `
		SELECT id, aggregate_id, event_type, event_data, timestamp, version
		FROM hft_events
		WHERE aggregate_id = $1
		ORDER BY version ASC`

	rows, err := es.db.QueryContext(ctx, query, aggregateID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	events, err := es.scanEvents(rows)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to scan events: %w", err)
	}

	span.SetAttributes(attribute.Int("events_count", len(events)))
	return events, nil
}

// GetEventsSince retrieves events for an aggregate since a specific time
func (es *PostgresEventStore) GetEventsSince(ctx context.Context, aggregateID string, since time.Time) ([]repositories.DomainEvent, error) {
	ctx, span := es.tracer.Start(ctx, "PostgresEventStore.GetEventsSince")
	defer span.End()

	span.SetAttributes(
		attribute.String("aggregate_id", aggregateID),
		attribute.String("since", since.Format(time.RFC3339)),
	)

	query := `
		SELECT id, aggregate_id, event_type, event_data, timestamp, version
		FROM hft_events
		WHERE aggregate_id = $1 AND timestamp > $2
		ORDER BY version ASC`

	rows, err := es.db.QueryContext(ctx, query, aggregateID, since)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query events since: %w", err)
	}
	defer rows.Close()

	events, err := es.scanEvents(rows)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to scan events: %w", err)
	}

	span.SetAttributes(attribute.Int("events_count", len(events)))
	return events, nil
}

// GetEventsFromVersion retrieves events for an aggregate from a specific version
func (es *PostgresEventStore) GetEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]repositories.DomainEvent, error) {
	ctx, span := es.tracer.Start(ctx, "PostgresEventStore.GetEventsFromVersion")
	defer span.End()

	span.SetAttributes(
		attribute.String("aggregate_id", aggregateID),
		attribute.Int("from_version", fromVersion),
	)

	query := `
		SELECT id, aggregate_id, event_type, event_data, timestamp, version
		FROM hft_events
		WHERE aggregate_id = $1 AND version >= $2
		ORDER BY version ASC`

	rows, err := es.db.QueryContext(ctx, query, aggregateID, fromVersion)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query events from version: %w", err)
	}
	defer rows.Close()

	events, err := es.scanEvents(rows)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to scan events: %w", err)
	}

	span.SetAttributes(attribute.Int("events_count", len(events)))
	return events, nil
}

// SaveSnapshot saves an aggregate snapshot
func (es *PostgresEventStore) SaveSnapshot(ctx context.Context, snapshot repositories.Snapshot) error {
	ctx, span := es.tracer.Start(ctx, "PostgresEventStore.SaveSnapshot")
	defer span.End()

	span.SetAttributes(
		attribute.String("aggregate_id", snapshot.AggregateID),
		attribute.Int("version", snapshot.Version),
	)

	query := `
		INSERT INTO hft_snapshots (aggregate_id, data, version, timestamp)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (aggregate_id) 
		DO UPDATE SET data = $2, version = $3, timestamp = $4`

	snapshotDataJSON, err := json.Marshal(snapshot.Data)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	_, err = es.db.ExecContext(ctx, query,
		snapshot.AggregateID,
		snapshotDataJSON,
		snapshot.Version,
		snapshot.Timestamp,
	)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	return nil
}

// GetLatestSnapshot retrieves the latest snapshot for an aggregate
func (es *PostgresEventStore) GetLatestSnapshot(ctx context.Context, aggregateID string) (*repositories.Snapshot, error) {
	ctx, span := es.tracer.Start(ctx, "PostgresEventStore.GetLatestSnapshot")
	defer span.End()

	span.SetAttributes(attribute.String("aggregate_id", aggregateID))

	query := `
		SELECT aggregate_id, data, version, timestamp
		FROM hft_snapshots
		WHERE aggregate_id = $1`

	row := es.db.QueryRowContext(ctx, query, aggregateID)

	var snapshot repositories.Snapshot
	var snapshotDataJSON []byte

	err := row.Scan(
		&snapshot.AggregateID,
		&snapshotDataJSON,
		&snapshot.Version,
		&snapshot.Timestamp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No snapshot found
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to scan snapshot: %w", err)
	}

	if err := json.Unmarshal(snapshotDataJSON, &snapshot.Data); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
	}

	span.SetAttributes(attribute.Int("snapshot_version", snapshot.Version))
	return &snapshot, nil
}

// GetEventStream retrieves an event stream (simplified implementation)
func (es *PostgresEventStore) GetEventStream(ctx context.Context, streamName string) (<-chan repositories.DomainEvent, error) {
	ctx, span := es.tracer.Start(ctx, "PostgresEventStore.GetEventStream")
	defer span.End()

	span.SetAttributes(attribute.String("stream_name", streamName))

	// This is a simplified implementation
	// In a production system, you might want to use PostgreSQL's LISTEN/NOTIFY
	// or implement a more sophisticated streaming mechanism
	
	eventChan := make(chan repositories.DomainEvent, 100)
	
	go func() {
		defer close(eventChan)
		
		// Query existing events in the stream
		query := `
			SELECT id, aggregate_id, event_type, event_data, timestamp, version
			FROM hft_events
			WHERE stream_name = $1
			ORDER BY version ASC`

		rows, err := es.db.QueryContext(ctx, query, streamName)
		if err != nil {
			span.RecordError(err)
			return
		}
		defer rows.Close()

		events, err := es.scanEvents(rows)
		if err != nil {
			span.RecordError(err)
			return
		}

		// Send events to channel
		for _, event := range events {
			select {
			case eventChan <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return eventChan, nil
}

// PublishEvent publishes a single event (simplified implementation)
func (es *PostgresEventStore) PublishEvent(ctx context.Context, event repositories.DomainEvent) error {
	ctx, span := es.tracer.Start(ctx, "PostgresEventStore.PublishEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("event_type", event.EventType),
		attribute.String("aggregate_id", event.AggregateID),
	)

	// In a real implementation, this might publish to a message queue
	// or use PostgreSQL's NOTIFY mechanism
	
	// For now, we'll just save the event
	return es.SaveEvents(ctx, event.AggregateID, []repositories.DomainEvent{event})
}

// Helper methods

// getCurrentVersion gets the current version for an aggregate
func (es *PostgresEventStore) getCurrentVersion(ctx context.Context, tx *sql.Tx, aggregateID string) (int, error) {
	query := `
		SELECT COALESCE(MAX(version), 0)
		FROM hft_events
		WHERE aggregate_id = $1`

	var version int
	err := tx.QueryRowContext(ctx, query, aggregateID).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}

	return version, nil
}

// scanEvents scans multiple events from database rows
func (es *PostgresEventStore) scanEvents(rows *sql.Rows) ([]repositories.DomainEvent, error) {
	var events []repositories.DomainEvent

	for rows.Next() {
		var event repositories.DomainEvent
		var eventDataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.EventType,
			&eventDataJSON,
			&event.Timestamp,
			&event.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		if err := json.Unmarshal(eventDataJSON, &event.EventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over event rows: %w", err)
	}

	return events, nil
}

// CreateTables creates the necessary database tables for event sourcing
func (es *PostgresEventStore) CreateTables(ctx context.Context) error {
	ctx, span := es.tracer.Start(ctx, "PostgresEventStore.CreateTables")
	defer span.End()

	// Create events table
	eventsTableQuery := `
		CREATE TABLE IF NOT EXISTS hft_events (
			id UUID PRIMARY KEY,
			aggregate_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			event_data JSONB NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			version INTEGER NOT NULL,
			stream_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(aggregate_id, version)
		);

		CREATE INDEX IF NOT EXISTS idx_hft_events_aggregate_id ON hft_events(aggregate_id);
		CREATE INDEX IF NOT EXISTS idx_hft_events_stream_name ON hft_events(stream_name);
		CREATE INDEX IF NOT EXISTS idx_hft_events_timestamp ON hft_events(timestamp);
		CREATE INDEX IF NOT EXISTS idx_hft_events_event_type ON hft_events(event_type);`

	// Create snapshots table
	snapshotsTableQuery := `
		CREATE TABLE IF NOT EXISTS hft_snapshots (
			aggregate_id VARCHAR(255) PRIMARY KEY,
			data JSONB NOT NULL,
			version INTEGER NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_hft_snapshots_version ON hft_snapshots(version);
		CREATE INDEX IF NOT EXISTS idx_hft_snapshots_timestamp ON hft_snapshots(timestamp);`

	// Execute table creation queries
	if _, err := es.db.ExecContext(ctx, eventsTableQuery); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create events table: %w", err)
	}

	if _, err := es.db.ExecContext(ctx, snapshotsTableQuery); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create snapshots table: %w", err)
	}

	return nil
}
