package recording

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// Repository implements the recording repository interface
type Repository struct {
	logger *zap.Logger
	db     *sql.DB
}

// NewRepository creates a new recording repository
func NewRepository(logger *zap.Logger, db *sql.DB) *Repository {
	return &Repository{
		logger: logger.With(zap.String("component", "recording_repository")),
		db:     db,
	}
}

// CreateRecording creates a new recording
func (r *Repository) CreateRecording(recording *domain.Recording) error {
	metadataJSON, err := json.Marshal(recording.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO recordings (
			id, stream_id, type, trigger, status, quality, format, filename, file_path,
			file_size, duration, start_time, end_time, pre_buffer, post_buffer,
			alert_id, zone_id, object_id, metadata, tags, created_at, updated_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
	`

	_, err = r.db.Exec(query,
		recording.ID,
		recording.StreamID,
		string(recording.Type),
		string(recording.Trigger),
		string(recording.Status),
		string(recording.Quality),
		string(recording.Format),
		recording.Filename,
		recording.FilePath,
		recording.FileSize,
		recording.Duration.Nanoseconds(),
		recording.StartTime,
		recording.EndTime,
		recording.PreBuffer.Nanoseconds(),
		recording.PostBuffer.Nanoseconds(),
		recording.AlertID,
		recording.ZoneID,
		recording.ObjectID,
		string(metadataJSON),
		pq.Array(recording.Tags),
		recording.CreatedAt,
		recording.UpdatedAt,
		recording.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert recording: %w", err)
	}

	r.logger.Info("Recording created in database",
		zap.String("recording_id", recording.ID),
		zap.String("stream_id", recording.StreamID))

	return nil
}

// GetRecording retrieves a recording by ID
func (r *Repository) GetRecording(id string) (*domain.Recording, error) {
	query := `
		SELECT id, stream_id, type, trigger, status, quality, format, filename, file_path,
			   file_size, duration, start_time, end_time, pre_buffer, post_buffer,
			   alert_id, zone_id, object_id, metadata, tags, created_at, updated_at, expires_at
		FROM recordings
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)

	var recording domain.Recording
	var recordingType, trigger, status, quality, format string
	var metadataJSON string
	var tags pq.StringArray
	var durationNanos, preBufferNanos, postBufferNanos int64

	err := row.Scan(
		&recording.ID,
		&recording.StreamID,
		&recordingType,
		&trigger,
		&status,
		&quality,
		&format,
		&recording.Filename,
		&recording.FilePath,
		&recording.FileSize,
		&durationNanos,
		&recording.StartTime,
		&recording.EndTime,
		&preBufferNanos,
		&postBufferNanos,
		&recording.AlertID,
		&recording.ZoneID,
		&recording.ObjectID,
		&metadataJSON,
		&tags,
		&recording.CreatedAt,
		&recording.UpdatedAt,
		&recording.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recording not found: %s", id)
		}
		return nil, fmt.Errorf("failed to scan recording: %w", err)
	}

	// Parse enum fields
	recording.Type = domain.RecordingType(recordingType)
	recording.Trigger = domain.RecordingTrigger(trigger)
	recording.Status = domain.RecordingStatus(status)
	recording.Quality = domain.RecordingQuality(quality)
	recording.Format = domain.VideoFormat(format)
	recording.Tags = []string(tags)

	// Parse duration fields
	recording.Duration = time.Duration(durationNanos)
	recording.PreBuffer = time.Duration(preBufferNanos)
	recording.PostBuffer = time.Duration(postBufferNanos)

	// Parse JSON fields
	if err := json.Unmarshal([]byte(metadataJSON), &recording.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &recording, nil
}

// GetRecordings retrieves recordings with filters
func (r *Repository) GetRecordings(filters domain.RecordingFilters) ([]*domain.Recording, error) {
	query := `
		SELECT id, stream_id, type, trigger, status, quality, format, filename, file_path,
			   file_size, duration, start_time, end_time, pre_buffer, post_buffer,
			   alert_id, zone_id, object_id, metadata, tags, created_at, updated_at, expires_at
		FROM recordings
		WHERE 1=1
	`

	var args []interface{}
	argIndex := 1

	// Build WHERE clause based on filters
	if filters.StreamID != "" {
		query += fmt.Sprintf(" AND stream_id = $%d", argIndex)
		args = append(args, filters.StreamID)
		argIndex++
	}

	if filters.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, string(filters.Type))
		argIndex++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(filters.Status))
		argIndex++
	}

	if filters.Trigger != "" {
		query += fmt.Sprintf(" AND trigger = $%d", argIndex)
		args = append(args, string(filters.Trigger))
		argIndex++
	}

	if filters.AlertID != "" {
		query += fmt.Sprintf(" AND alert_id = $%d", argIndex)
		args = append(args, filters.AlertID)
		argIndex++
	}

	if filters.ZoneID != "" {
		query += fmt.Sprintf(" AND zone_id = $%d", argIndex)
		args = append(args, filters.ZoneID)
		argIndex++
	}

	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND start_time >= $%d", argIndex)
		args = append(args, *filters.StartTime)
		argIndex++
	}

	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND start_time <= $%d", argIndex)
		args = append(args, *filters.EndTime)
		argIndex++
	}

	if filters.MinDuration != nil {
		query += fmt.Sprintf(" AND duration >= $%d", argIndex)
		args = append(args, filters.MinDuration.Nanoseconds())
		argIndex++
	}

	if filters.MaxDuration != nil {
		query += fmt.Sprintf(" AND duration <= $%d", argIndex)
		args = append(args, filters.MaxDuration.Nanoseconds())
		argIndex++
	}

	if len(filters.Tags) > 0 {
		query += fmt.Sprintf(" AND tags && $%d", argIndex)
		args = append(args, pq.Array(filters.Tags))
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY start_time DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
		argIndex++
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query recordings: %w", err)
	}
	defer rows.Close()

	var recordings []*domain.Recording

	for rows.Next() {
		var recording domain.Recording
		var recordingType, trigger, status, quality, format string
		var metadataJSON string
		var tags pq.StringArray
		var durationNanos, preBufferNanos, postBufferNanos int64

		err := rows.Scan(
			&recording.ID,
			&recording.StreamID,
			&recordingType,
			&trigger,
			&status,
			&quality,
			&format,
			&recording.Filename,
			&recording.FilePath,
			&recording.FileSize,
			&durationNanos,
			&recording.StartTime,
			&recording.EndTime,
			&preBufferNanos,
			&postBufferNanos,
			&recording.AlertID,
			&recording.ZoneID,
			&recording.ObjectID,
			&metadataJSON,
			&tags,
			&recording.CreatedAt,
			&recording.UpdatedAt,
			&recording.ExpiresAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan recording: %w", err)
		}

		// Parse enum fields
		recording.Type = domain.RecordingType(recordingType)
		recording.Trigger = domain.RecordingTrigger(trigger)
		recording.Status = domain.RecordingStatus(status)
		recording.Quality = domain.RecordingQuality(quality)
		recording.Format = domain.VideoFormat(format)
		recording.Tags = []string(tags)

		// Parse duration fields
		recording.Duration = time.Duration(durationNanos)
		recording.PreBuffer = time.Duration(preBufferNanos)
		recording.PostBuffer = time.Duration(postBufferNanos)

		// Parse JSON fields
		if err := json.Unmarshal([]byte(metadataJSON), &recording.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		recordings = append(recordings, &recording)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate recordings: %w", err)
	}

	return recordings, nil
}

// UpdateRecording updates an existing recording
func (r *Repository) UpdateRecording(recording *domain.Recording) error {
	metadataJSON, err := json.Marshal(recording.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE recordings
		SET stream_id = $2, type = $3, trigger = $4, status = $5, quality = $6, format = $7,
			filename = $8, file_path = $9, file_size = $10, duration = $11, start_time = $12,
			end_time = $13, pre_buffer = $14, post_buffer = $15, alert_id = $16, zone_id = $17,
			object_id = $18, metadata = $19, tags = $20, updated_at = $21, expires_at = $22
		WHERE id = $1
	`

	result, err := r.db.Exec(query,
		recording.ID,
		recording.StreamID,
		string(recording.Type),
		string(recording.Trigger),
		string(recording.Status),
		string(recording.Quality),
		string(recording.Format),
		recording.Filename,
		recording.FilePath,
		recording.FileSize,
		recording.Duration.Nanoseconds(),
		recording.StartTime,
		recording.EndTime,
		recording.PreBuffer.Nanoseconds(),
		recording.PostBuffer.Nanoseconds(),
		recording.AlertID,
		recording.ZoneID,
		recording.ObjectID,
		string(metadataJSON),
		pq.Array(recording.Tags),
		recording.UpdatedAt,
		recording.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update recording: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recording not found: %s", recording.ID)
	}

	r.logger.Info("Recording updated in database", zap.String("recording_id", recording.ID))
	return nil
}

// DeleteRecording deletes a recording
func (r *Repository) DeleteRecording(id string) error {
	// Start transaction to delete recording and related data
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete recording clips first (foreign key constraint)
	_, err = tx.Exec("DELETE FROM recording_clips WHERE recording_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete recording clips: %w", err)
	}

	// Delete the recording
	result, err := tx.Exec("DELETE FROM recordings WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete recording: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recording not found: %s", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Recording deleted from database", zap.String("recording_id", id))
	return nil
}

// CreateRecordingClip creates a new recording clip
func (r *Repository) CreateRecordingClip(clip *domain.RecordingClip) error {
	metadataJSON, err := json.Marshal(clip.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO recording_clips (
			id, recording_id, stream_id, alert_id, title, description, filename, file_path,
			file_size, duration, start_offset, end_offset, quality, format, thumbnail,
			metadata, tags, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err = r.db.Exec(query,
		clip.ID,
		clip.RecordingID,
		clip.StreamID,
		clip.AlertID,
		clip.Title,
		clip.Description,
		clip.Filename,
		clip.FilePath,
		clip.FileSize,
		clip.Duration.Nanoseconds(),
		clip.StartOffset.Nanoseconds(),
		clip.EndOffset.Nanoseconds(),
		string(clip.Quality),
		string(clip.Format),
		clip.Thumbnail,
		string(metadataJSON),
		pq.Array(clip.Tags),
		clip.CreatedAt,
		clip.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert recording clip: %w", err)
	}

	r.logger.Info("Recording clip created in database",
		zap.String("clip_id", clip.ID),
		zap.String("recording_id", clip.RecordingID))

	return nil
}

// GetRecordingClip retrieves a recording clip by ID
func (r *Repository) GetRecordingClip(id string) (*domain.RecordingClip, error) {
	query := `
		SELECT id, recording_id, stream_id, alert_id, title, description, filename, file_path,
			   file_size, duration, start_offset, end_offset, quality, format, thumbnail,
			   metadata, tags, created_at, expires_at
		FROM recording_clips
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)

	var clip domain.RecordingClip
	var quality, format string
	var metadataJSON string
	var tags pq.StringArray
	var durationNanos, startOffsetNanos, endOffsetNanos int64

	err := row.Scan(
		&clip.ID,
		&clip.RecordingID,
		&clip.StreamID,
		&clip.AlertID,
		&clip.Title,
		&clip.Description,
		&clip.Filename,
		&clip.FilePath,
		&clip.FileSize,
		&durationNanos,
		&startOffsetNanos,
		&endOffsetNanos,
		&quality,
		&format,
		&clip.Thumbnail,
		&metadataJSON,
		&tags,
		&clip.CreatedAt,
		&clip.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recording clip not found: %s", id)
		}
		return nil, fmt.Errorf("failed to scan recording clip: %w", err)
	}

	// Parse enum fields
	clip.Quality = domain.RecordingQuality(quality)
	clip.Format = domain.VideoFormat(format)
	clip.Tags = []string(tags)

	// Parse duration fields
	clip.Duration = time.Duration(durationNanos)
	clip.StartOffset = time.Duration(startOffsetNanos)
	clip.EndOffset = time.Duration(endOffsetNanos)

	// Parse JSON fields
	if err := json.Unmarshal([]byte(metadataJSON), &clip.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &clip, nil
}

// GetRecordingClips retrieves clips for a recording
func (r *Repository) GetRecordingClips(recordingID string) ([]*domain.RecordingClip, error) {
	query := `
		SELECT id, recording_id, stream_id, alert_id, title, description, filename, file_path,
			   file_size, duration, start_offset, end_offset, quality, format, thumbnail,
			   metadata, tags, created_at, expires_at
		FROM recording_clips
		WHERE recording_id = $1
		ORDER BY start_offset ASC
	`

	rows, err := r.db.Query(query, recordingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query recording clips: %w", err)
	}
	defer rows.Close()

	var clips []*domain.RecordingClip

	for rows.Next() {
		var clip domain.RecordingClip
		var quality, format string
		var metadataJSON string
		var tags pq.StringArray
		var durationNanos, startOffsetNanos, endOffsetNanos int64

		err := rows.Scan(
			&clip.ID,
			&clip.RecordingID,
			&clip.StreamID,
			&clip.AlertID,
			&clip.Title,
			&clip.Description,
			&clip.Filename,
			&clip.FilePath,
			&clip.FileSize,
			&durationNanos,
			&startOffsetNanos,
			&endOffsetNanos,
			&quality,
			&format,
			&clip.Thumbnail,
			&metadataJSON,
			&tags,
			&clip.CreatedAt,
			&clip.ExpiresAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan recording clip: %w", err)
		}

		// Parse enum fields
		clip.Quality = domain.RecordingQuality(quality)
		clip.Format = domain.VideoFormat(format)
		clip.Tags = []string(tags)

		// Parse duration fields
		clip.Duration = time.Duration(durationNanos)
		clip.StartOffset = time.Duration(startOffsetNanos)
		clip.EndOffset = time.Duration(endOffsetNanos)

		// Parse JSON fields
		if err := json.Unmarshal([]byte(metadataJSON), &clip.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		clips = append(clips, &clip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate recording clips: %w", err)
	}

	return clips, nil
}

// DeleteRecordingClip deletes a recording clip
func (r *Repository) DeleteRecordingClip(id string) error {
	result, err := r.db.Exec("DELETE FROM recording_clips WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete recording clip: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recording clip not found: %s", id)
	}

	r.logger.Info("Recording clip deleted from database", zap.String("clip_id", id))
	return nil
}

// GetRecordingStatistics retrieves recording statistics
func (r *Repository) GetRecordingStatistics(streamID string) (*domain.RecordingStatistics, error) {
	query := `
		SELECT stream_id, total_recordings, active_recordings, total_duration, total_size,
			   recordings_by_type, recordings_by_status, recordings_by_hour, recordings_by_day,
			   average_file_size, average_duration, last_recording, start_time
		FROM recording_statistics
		WHERE stream_id = $1
	`

	row := r.db.QueryRow(query, streamID)

	var stats domain.RecordingStatistics
	var recordingsByTypeJSON, recordingsByStatusJSON, recordingsByHourJSON, recordingsByDayJSON string
	var totalDurationNanos, averageDurationNanos int64

	err := row.Scan(
		&stats.StreamID,
		&stats.TotalRecordings,
		&stats.ActiveRecordings,
		&totalDurationNanos,
		&stats.TotalSize,
		&recordingsByTypeJSON,
		&recordingsByStatusJSON,
		&recordingsByHourJSON,
		&recordingsByDayJSON,
		&stats.AverageFileSize,
		&averageDurationNanos,
		&stats.LastRecording,
		&stats.StartTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recording statistics not found: %s", streamID)
		}
		return nil, fmt.Errorf("failed to scan recording statistics: %w", err)
	}

	// Parse duration fields
	stats.TotalDuration = time.Duration(totalDurationNanos)
	stats.AverageDuration = time.Duration(averageDurationNanos)

	// Parse JSON fields
	if err := json.Unmarshal([]byte(recordingsByTypeJSON), &stats.RecordingsByType); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recordings by type: %w", err)
	}

	if err := json.Unmarshal([]byte(recordingsByStatusJSON), &stats.RecordingsByStatus); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recordings by status: %w", err)
	}

	if err := json.Unmarshal([]byte(recordingsByHourJSON), &stats.RecordingsByHour); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recordings by hour: %w", err)
	}

	if err := json.Unmarshal([]byte(recordingsByDayJSON), &stats.RecordingsByDay); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recordings by day: %w", err)
	}

	return &stats, nil
}

// UpdateRecordingStatistics updates recording statistics
func (r *Repository) UpdateRecordingStatistics(stats *domain.RecordingStatistics) error {
	recordingsByTypeJSON, err := json.Marshal(stats.RecordingsByType)
	if err != nil {
		return fmt.Errorf("failed to marshal recordings by type: %w", err)
	}

	recordingsByStatusJSON, err := json.Marshal(stats.RecordingsByStatus)
	if err != nil {
		return fmt.Errorf("failed to marshal recordings by status: %w", err)
	}

	recordingsByHourJSON, err := json.Marshal(stats.RecordingsByHour)
	if err != nil {
		return fmt.Errorf("failed to marshal recordings by hour: %w", err)
	}

	recordingsByDayJSON, err := json.Marshal(stats.RecordingsByDay)
	if err != nil {
		return fmt.Errorf("failed to marshal recordings by day: %w", err)
	}

	query := `
		INSERT INTO recording_statistics (
			stream_id, total_recordings, active_recordings, total_duration, total_size,
			recordings_by_type, recordings_by_status, recordings_by_hour, recordings_by_day,
			average_file_size, average_duration, last_recording, start_time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (stream_id) DO UPDATE SET
			total_recordings = EXCLUDED.total_recordings,
			active_recordings = EXCLUDED.active_recordings,
			total_duration = EXCLUDED.total_duration,
			total_size = EXCLUDED.total_size,
			recordings_by_type = EXCLUDED.recordings_by_type,
			recordings_by_status = EXCLUDED.recordings_by_status,
			recordings_by_hour = EXCLUDED.recordings_by_hour,
			recordings_by_day = EXCLUDED.recordings_by_day,
			average_file_size = EXCLUDED.average_file_size,
			average_duration = EXCLUDED.average_duration,
			last_recording = EXCLUDED.last_recording
	`

	_, err = r.db.Exec(query,
		stats.StreamID,
		stats.TotalRecordings,
		stats.ActiveRecordings,
		stats.TotalDuration.Nanoseconds(),
		stats.TotalSize,
		string(recordingsByTypeJSON),
		string(recordingsByStatusJSON),
		string(recordingsByHourJSON),
		string(recordingsByDayJSON),
		stats.AverageFileSize,
		stats.AverageDuration.Nanoseconds(),
		stats.LastRecording,
		stats.StartTime,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert recording statistics: %w", err)
	}

	return nil
}

// Ensure Repository implements domain.RecordingRepository
var _ domain.RecordingRepository = (*Repository)(nil)
