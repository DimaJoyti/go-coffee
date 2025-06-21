package zones

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
)

// Repository implements the zone repository interface
type Repository struct {
	logger *zap.Logger
	db     *sql.DB
}

// NewRepository creates a new zone repository
func NewRepository(logger *zap.Logger, db *sql.DB) *Repository {
	return &Repository{
		logger: logger.With(zap.String("component", "zone_repository")),
		db:     db,
	}
}

// CreateZone creates a new detection zone
func (r *Repository) CreateZone(zone *domain.DetectionZone) error {
	polygonJSON, err := json.Marshal(zone.Polygon)
	if err != nil {
		return fmt.Errorf("failed to marshal polygon: %w", err)
	}

	rulesJSON, err := json.Marshal(zone.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	query := `
		INSERT INTO detection_zones (
			id, stream_id, name, description, type, polygon, rules, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Exec(query,
		zone.ID,
		zone.StreamID,
		zone.Name,
		zone.Description,
		string(zone.Type),
		string(polygonJSON),
		string(rulesJSON),
		zone.IsActive,
		zone.CreatedAt,
		zone.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert zone: %w", err)
	}

	r.logger.Info("Zone created in database",
		zap.String("zone_id", zone.ID),
		zap.String("stream_id", zone.StreamID))

	return nil
}

// GetZone retrieves a zone by ID
func (r *Repository) GetZone(id string) (*domain.DetectionZone, error) {
	query := `
		SELECT id, stream_id, name, description, type, polygon, rules, is_active, created_at, updated_at
		FROM detection_zones
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)

	var zone domain.DetectionZone
	var polygonJSON, rulesJSON string
	var zoneType string

	err := row.Scan(
		&zone.ID,
		&zone.StreamID,
		&zone.Name,
		&zone.Description,
		&zoneType,
		&polygonJSON,
		&rulesJSON,
		&zone.IsActive,
		&zone.CreatedAt,
		&zone.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("zone not found: %s", id)
		}
		return nil, fmt.Errorf("failed to scan zone: %w", err)
	}

	// Parse JSON fields
	zone.Type = domain.ZoneType(zoneType)

	if err := json.Unmarshal([]byte(polygonJSON), &zone.Polygon); err != nil {
		return nil, fmt.Errorf("failed to unmarshal polygon: %w", err)
	}

	if err := json.Unmarshal([]byte(rulesJSON), &zone.Rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	return &zone, nil
}

// GetZonesByStream retrieves all zones for a stream
func (r *Repository) GetZonesByStream(streamID string) ([]*domain.DetectionZone, error) {
	query := `
		SELECT id, stream_id, name, description, type, polygon, rules, is_active, created_at, updated_at
		FROM detection_zones
		WHERE stream_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query zones: %w", err)
	}
	defer rows.Close()

	var zones []*domain.DetectionZone

	for rows.Next() {
		var zone domain.DetectionZone
		var polygonJSON, rulesJSON string
		var zoneType string

		err := rows.Scan(
			&zone.ID,
			&zone.StreamID,
			&zone.Name,
			&zone.Description,
			&zoneType,
			&polygonJSON,
			&rulesJSON,
			&zone.IsActive,
			&zone.CreatedAt,
			&zone.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}

		// Parse JSON fields
		zone.Type = domain.ZoneType(zoneType)

		if err := json.Unmarshal([]byte(polygonJSON), &zone.Polygon); err != nil {
			return nil, fmt.Errorf("failed to unmarshal polygon: %w", err)
		}

		if err := json.Unmarshal([]byte(rulesJSON), &zone.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}

		zones = append(zones, &zone)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate zones: %w", err)
	}

	return zones, nil
}

// UpdateZone updates an existing zone
func (r *Repository) UpdateZone(zone *domain.DetectionZone) error {
	polygonJSON, err := json.Marshal(zone.Polygon)
	if err != nil {
		return fmt.Errorf("failed to marshal polygon: %w", err)
	}

	rulesJSON, err := json.Marshal(zone.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	query := `
		UPDATE detection_zones
		SET name = $2, description = $3, type = $4, polygon = $5, rules = $6, is_active = $7, updated_at = $8
		WHERE id = $1
	`

	result, err := r.db.Exec(query,
		zone.ID,
		zone.Name,
		zone.Description,
		string(zone.Type),
		string(polygonJSON),
		string(rulesJSON),
		zone.IsActive,
		zone.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update zone: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("zone not found: %s", zone.ID)
	}

	r.logger.Info("Zone updated in database", zap.String("zone_id", zone.ID))
	return nil
}

// DeleteZone deletes a zone
func (r *Repository) DeleteZone(id string) error {
	// Start transaction to delete zone and related data
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete zone events first (foreign key constraint)
	_, err = tx.Exec("DELETE FROM zone_events WHERE zone_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete zone events: %w", err)
	}

	// Delete zone statistics
	_, err = tx.Exec("DELETE FROM zone_statistics WHERE zone_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete zone statistics: %w", err)
	}

	// Delete the zone
	result, err := tx.Exec("DELETE FROM detection_zones WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("zone not found: %s", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Zone deleted from database", zap.String("zone_id", id))
	return nil
}

// ListZones lists all zones with pagination
func (r *Repository) ListZones(limit, offset int) ([]*domain.DetectionZone, error) {
	query := `
		SELECT id, stream_id, name, description, type, polygon, rules, is_active, created_at, updated_at
		FROM detection_zones
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query zones: %w", err)
	}
	defer rows.Close()

	var zones []*domain.DetectionZone

	for rows.Next() {
		var zone domain.DetectionZone
		var polygonJSON, rulesJSON string
		var zoneType string

		err := rows.Scan(
			&zone.ID,
			&zone.StreamID,
			&zone.Name,
			&zone.Description,
			&zoneType,
			&polygonJSON,
			&rulesJSON,
			&zone.IsActive,
			&zone.CreatedAt,
			&zone.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}

		// Parse JSON fields
		zone.Type = domain.ZoneType(zoneType)

		if err := json.Unmarshal([]byte(polygonJSON), &zone.Polygon); err != nil {
			return nil, fmt.Errorf("failed to unmarshal polygon: %w", err)
		}

		if err := json.Unmarshal([]byte(rulesJSON), &zone.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}

		zones = append(zones, &zone)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate zones: %w", err)
	}

	return zones, nil
}

// CreateZoneEvent creates a new zone event
func (r *Repository) CreateZoneEvent(event *domain.ZoneEvent) error {
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	boundingBoxJSON, err := json.Marshal(event.BoundingBox)
	if err != nil {
		return fmt.Errorf("failed to marshal bounding box: %w", err)
	}

	positionJSON, err := json.Marshal(event.Position)
	if err != nil {
		return fmt.Errorf("failed to marshal position: %w", err)
	}

	query := `
		INSERT INTO zone_events (
			id, zone_id, stream_id, event_type, object_id, object_class, confidence,
			position, bounding_box, dwell_time, metadata, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.db.Exec(query,
		event.ID,
		event.ZoneID,
		event.StreamID,
		string(event.EventType),
		event.ObjectID,
		event.ObjectClass,
		event.Confidence,
		string(positionJSON),
		string(boundingBoxJSON),
		event.DwellTime.Nanoseconds(),
		string(metadataJSON),
		event.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to insert zone event: %w", err)
	}

	return nil
}

// GetZoneEvents retrieves zone events with pagination
func (r *Repository) GetZoneEvents(zoneID string, limit, offset int) ([]*domain.ZoneEvent, error) {
	query := `
		SELECT id, zone_id, stream_id, event_type, object_id, object_class, confidence,
			   position, bounding_box, dwell_time, metadata, timestamp
		FROM zone_events
		WHERE zone_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, zoneID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query zone events: %w", err)
	}
	defer rows.Close()

	var events []*domain.ZoneEvent

	for rows.Next() {
		var event domain.ZoneEvent
		var eventType string
		var positionJSON, boundingBoxJSON, metadataJSON string
		var dwellTimeNanos int64

		err := rows.Scan(
			&event.ID,
			&event.ZoneID,
			&event.StreamID,
			&eventType,
			&event.ObjectID,
			&event.ObjectClass,
			&event.Confidence,
			&positionJSON,
			&boundingBoxJSON,
			&dwellTimeNanos,
			&metadataJSON,
			&event.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan zone event: %w", err)
		}

		// Parse fields
		event.EventType = domain.ZoneEventType(eventType)
		event.DwellTime = time.Duration(dwellTimeNanos)

		if err := json.Unmarshal([]byte(positionJSON), &event.Position); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position: %w", err)
		}

		if err := json.Unmarshal([]byte(boundingBoxJSON), &event.BoundingBox); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bounding box: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &event.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate zone events: %w", err)
	}

	return events, nil
}

// GetZoneEventsByTimeRange retrieves zone events within a time range
func (r *Repository) GetZoneEventsByTimeRange(zoneID string, start, end time.Time) ([]*domain.ZoneEvent, error) {
	query := `
		SELECT id, zone_id, stream_id, event_type, object_id, object_class, confidence,
			   position, bounding_box, dwell_time, metadata, timestamp
		FROM zone_events
		WHERE zone_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC
	`

	rows, err := r.db.Query(query, zoneID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query zone events: %w", err)
	}
	defer rows.Close()

	var events []*domain.ZoneEvent

	for rows.Next() {
		var event domain.ZoneEvent
		var eventType string
		var positionJSON, boundingBoxJSON, metadataJSON string
		var dwellTimeNanos int64

		err := rows.Scan(
			&event.ID,
			&event.ZoneID,
			&event.StreamID,
			&eventType,
			&event.ObjectID,
			&event.ObjectClass,
			&event.Confidence,
			&positionJSON,
			&boundingBoxJSON,
			&dwellTimeNanos,
			&metadataJSON,
			&event.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan zone event: %w", err)
		}

		// Parse fields
		event.EventType = domain.ZoneEventType(eventType)
		event.DwellTime = time.Duration(dwellTimeNanos)

		if err := json.Unmarshal([]byte(positionJSON), &event.Position); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position: %w", err)
		}

		if err := json.Unmarshal([]byte(boundingBoxJSON), &event.BoundingBox); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bounding box: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &event.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate zone events: %w", err)
	}

	return events, nil
}

// GetZoneStatistics retrieves zone statistics
func (r *Repository) GetZoneStatistics(zoneID string) (*domain.ZoneStatistics, error) {
	query := `
		SELECT zone_id, stream_id, total_entries, total_exits, current_occupancy, max_occupancy,
			   average_dwell_time, object_counts, hourly_stats, daily_stats, last_activity, start_time
		FROM zone_statistics
		WHERE zone_id = $1
	`

	row := r.db.QueryRow(query, zoneID)

	var stats domain.ZoneStatistics
	var objectCountsJSON, hourlyStatsJSON, dailyStatsJSON string
	var avgDwellTimeNanos int64

	err := row.Scan(
		&stats.ZoneID,
		&stats.StreamID,
		&stats.TotalEntries,
		&stats.TotalExits,
		&stats.CurrentOccupancy,
		&stats.MaxOccupancy,
		&avgDwellTimeNanos,
		&objectCountsJSON,
		&hourlyStatsJSON,
		&dailyStatsJSON,
		&stats.LastActivity,
		&stats.StartTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("zone statistics not found: %s", zoneID)
		}
		return nil, fmt.Errorf("failed to scan zone statistics: %w", err)
	}

	// Parse JSON fields
	stats.AverageDwellTime = time.Duration(avgDwellTimeNanos)

	if err := json.Unmarshal([]byte(objectCountsJSON), &stats.ObjectCounts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object counts: %w", err)
	}

	if err := json.Unmarshal([]byte(hourlyStatsJSON), &stats.HourlyStats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hourly stats: %w", err)
	}

	if err := json.Unmarshal([]byte(dailyStatsJSON), &stats.DailyStats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal daily stats: %w", err)
	}

	return &stats, nil
}

// UpdateZoneStatistics updates zone statistics
func (r *Repository) UpdateZoneStatistics(stats *domain.ZoneStatistics) error {
	objectCountsJSON, err := json.Marshal(stats.ObjectCounts)
	if err != nil {
		return fmt.Errorf("failed to marshal object counts: %w", err)
	}

	hourlyStatsJSON, err := json.Marshal(stats.HourlyStats)
	if err != nil {
		return fmt.Errorf("failed to marshal hourly stats: %w", err)
	}

	dailyStatsJSON, err := json.Marshal(stats.DailyStats)
	if err != nil {
		return fmt.Errorf("failed to marshal daily stats: %w", err)
	}

	query := `
		INSERT INTO zone_statistics (
			zone_id, stream_id, total_entries, total_exits, current_occupancy, max_occupancy,
			average_dwell_time, object_counts, hourly_stats, daily_stats, last_activity, start_time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (zone_id) DO UPDATE SET
			total_entries = EXCLUDED.total_entries,
			total_exits = EXCLUDED.total_exits,
			current_occupancy = EXCLUDED.current_occupancy,
			max_occupancy = EXCLUDED.max_occupancy,
			average_dwell_time = EXCLUDED.average_dwell_time,
			object_counts = EXCLUDED.object_counts,
			hourly_stats = EXCLUDED.hourly_stats,
			daily_stats = EXCLUDED.daily_stats,
			last_activity = EXCLUDED.last_activity
	`

	_, err = r.db.Exec(query,
		stats.ZoneID,
		stats.StreamID,
		stats.TotalEntries,
		stats.TotalExits,
		stats.CurrentOccupancy,
		stats.MaxOccupancy,
		stats.AverageDwellTime.Nanoseconds(),
		string(objectCountsJSON),
		string(hourlyStatsJSON),
		string(dailyStatsJSON),
		stats.LastActivity,
		stats.StartTime,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert zone statistics: %w", err)
	}

	return nil
}
