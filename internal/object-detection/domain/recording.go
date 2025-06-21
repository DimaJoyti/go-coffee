package domain

import (
	"fmt"
	"time"
)

// Recording represents a video recording session
type Recording struct {
	ID          string            `json:"id" db:"id"`
	StreamID    string            `json:"stream_id" db:"stream_id"`
	Type        RecordingType     `json:"type" db:"type"`
	Trigger     RecordingTrigger  `json:"trigger" db:"trigger"`
	Status      RecordingStatus   `json:"status" db:"status"`
	Quality     RecordingQuality  `json:"quality" db:"quality"`
	Format      VideoFormat       `json:"format" db:"format"`
	Filename    string            `json:"filename" db:"filename"`
	FilePath    string            `json:"file_path" db:"file_path"`
	FileSize    int64             `json:"file_size" db:"file_size"`
	Duration    time.Duration     `json:"duration" db:"duration"`
	StartTime   time.Time         `json:"start_time" db:"start_time"`
	EndTime     *time.Time        `json:"end_time,omitempty" db:"end_time"`
	PreBuffer   time.Duration     `json:"pre_buffer" db:"pre_buffer"`
	PostBuffer  time.Duration     `json:"post_buffer" db:"post_buffer"`
	AlertID     string            `json:"alert_id,omitempty" db:"alert_id"`
	ZoneID      string            `json:"zone_id,omitempty" db:"zone_id"`
	ObjectID    string            `json:"object_id,omitempty" db:"object_id"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Tags        []string          `json:"tags" db:"tags"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty" db:"expires_at"`
}

// RecordingType defines the type of recording
type RecordingType string

const (
	RecordingTypeContinuous    RecordingType = "continuous"
	RecordingTypeEventTriggered RecordingType = "event_triggered"
	RecordingTypeScheduled     RecordingType = "scheduled"
	RecordingTypeManual        RecordingType = "manual"
	RecordingTypeClip          RecordingType = "clip"
	RecordingTypeSnapshot      RecordingType = "snapshot"
)

// RecordingTrigger defines what triggered the recording
type RecordingTrigger string

const (
	RecordingTriggerAlert         RecordingTrigger = "alert"
	RecordingTriggerZoneViolation RecordingTrigger = "zone_violation"
	RecordingTriggerMotion        RecordingTrigger = "motion"
	RecordingTriggerSchedule      RecordingTrigger = "schedule"
	RecordingTriggerManual        RecordingTrigger = "manual"
	RecordingTriggerAPI           RecordingTrigger = "api"
	RecordingTriggerSystem        RecordingTrigger = "system"
)

// RecordingStatus defines the current status of a recording
type RecordingStatus string

const (
	RecordingStatusPending    RecordingStatus = "pending"
	RecordingStatusRecording  RecordingStatus = "recording"
	RecordingStatusCompleted  RecordingStatus = "completed"
	RecordingStatusFailed     RecordingStatus = "failed"
	RecordingStatusCancelled  RecordingStatus = "cancelled"
	RecordingStatusProcessing RecordingStatus = "processing"
	RecordingStatusArchived   RecordingStatus = "archived"
	RecordingStatusExpired    RecordingStatus = "expired"
)

// RecordingQuality defines the quality settings for recording
type RecordingQuality string

const (
	RecordingQualityLow    RecordingQuality = "low"
	RecordingQualityMedium RecordingQuality = "medium"
	RecordingQualityHigh   RecordingQuality = "high"
	RecordingQualityUltra  RecordingQuality = "ultra"
	RecordingQualityCustom RecordingQuality = "custom"
)

// VideoFormat defines the video file format
type VideoFormat string

const (
	VideoFormatMP4  VideoFormat = "mp4"
	VideoFormatAVI  VideoFormat = "avi"
	VideoFormatMOV  VideoFormat = "mov"
	VideoFormatMKV  VideoFormat = "mkv"
	VideoFormatWEBM VideoFormat = "webm"
)

// RecordingConfig defines configuration for recording
type RecordingConfig struct {
	StreamID         string            `json:"stream_id" yaml:"stream_id"`
	Type             RecordingType     `json:"type" yaml:"type"`
	Quality          RecordingQuality  `json:"quality" yaml:"quality"`
	Format           VideoFormat       `json:"format" yaml:"format"`
	MaxDuration      time.Duration     `json:"max_duration" yaml:"max_duration"`
	PreBuffer        time.Duration     `json:"pre_buffer" yaml:"pre_buffer"`
	PostBuffer       time.Duration     `json:"post_buffer" yaml:"post_buffer"`
	StoragePath      string            `json:"storage_path" yaml:"storage_path"`
	RetentionPolicy  RetentionPolicy   `json:"retention_policy" yaml:"retention_policy"`
	CompressionLevel int               `json:"compression_level" yaml:"compression_level"`
	FrameRate        float64           `json:"frame_rate" yaml:"frame_rate"`
	Resolution       Resolution        `json:"resolution" yaml:"resolution"`
	Bitrate          int               `json:"bitrate" yaml:"bitrate"`
	EnableAudio      bool              `json:"enable_audio" yaml:"enable_audio"`
	EnableOverlays   bool              `json:"enable_overlays" yaml:"enable_overlays"`
	WatermarkConfig  *WatermarkConfig  `json:"watermark_config,omitempty" yaml:"watermark_config"`
	Triggers         []TriggerConfig   `json:"triggers" yaml:"triggers"`
}

// RetentionPolicy defines how long recordings should be kept
type RetentionPolicy struct {
	MaxAge          time.Duration `json:"max_age" yaml:"max_age"`
	MaxSize         int64         `json:"max_size" yaml:"max_size"`
	MaxCount        int           `json:"max_count" yaml:"max_count"`
	AutoArchive     bool          `json:"auto_archive" yaml:"auto_archive"`
	ArchiveAfter    time.Duration `json:"archive_after" yaml:"archive_after"`
	CompressionType string        `json:"compression_type" yaml:"compression_type"`
	DeleteAfter     time.Duration `json:"delete_after" yaml:"delete_after"`
}

// WatermarkConfig defines watermark settings
type WatermarkConfig struct {
	Enabled     bool    `json:"enabled" yaml:"enabled"`
	Text        string  `json:"text" yaml:"text"`
	Position    string  `json:"position" yaml:"position"` // top-left, top-right, bottom-left, bottom-right, center
	FontSize    int     `json:"font_size" yaml:"font_size"`
	FontColor   string  `json:"font_color" yaml:"font_color"`
	Opacity     float64 `json:"opacity" yaml:"opacity"`
	ImagePath   string  `json:"image_path,omitempty" yaml:"image_path"`
	ImageScale  float64 `json:"image_scale" yaml:"image_scale"`
}

// TriggerConfig defines recording trigger configuration
type TriggerConfig struct {
	Type        RecordingTrigger  `json:"type" yaml:"type"`
	Conditions  TriggerConditions `json:"conditions" yaml:"conditions"`
	PreBuffer   time.Duration     `json:"pre_buffer" yaml:"pre_buffer"`
	PostBuffer  time.Duration     `json:"post_buffer" yaml:"post_buffer"`
	MaxDuration time.Duration     `json:"max_duration" yaml:"max_duration"`
	Cooldown    time.Duration     `json:"cooldown" yaml:"cooldown"`
	IsEnabled   bool              `json:"is_enabled" yaml:"is_enabled"`
}

// TriggerConditions defines conditions for recording triggers
type TriggerConditions struct {
	AlertTypes      []AlertType       `json:"alert_types,omitempty" yaml:"alert_types"`
	AlertSeverities []AlertSeverity   `json:"alert_severities,omitempty" yaml:"alert_severities"`
	ZoneIDs         []string          `json:"zone_ids,omitempty" yaml:"zone_ids"`
	ObjectClasses   []string          `json:"object_classes,omitempty" yaml:"object_classes"`
	MinConfidence   float64           `json:"min_confidence,omitempty" yaml:"min_confidence"`
	TimeWindows     []TimeWindow      `json:"time_windows,omitempty" yaml:"time_windows"`
	CustomFilters   map[string]interface{} `json:"custom_filters,omitempty" yaml:"custom_filters"`
}

// RecordingClip represents a short video clip extracted from a recording
type RecordingClip struct {
	ID          string            `json:"id" db:"id"`
	RecordingID string            `json:"recording_id" db:"recording_id"`
	StreamID    string            `json:"stream_id" db:"stream_id"`
	AlertID     string            `json:"alert_id,omitempty" db:"alert_id"`
	Title       string            `json:"title" db:"title"`
	Description string            `json:"description" db:"description"`
	Filename    string            `json:"filename" db:"filename"`
	FilePath    string            `json:"file_path" db:"file_path"`
	FileSize    int64             `json:"file_size" db:"file_size"`
	Duration    time.Duration     `json:"duration" db:"duration"`
	StartOffset time.Duration     `json:"start_offset" db:"start_offset"`
	EndOffset   time.Duration     `json:"end_offset" db:"end_offset"`
	Quality     RecordingQuality  `json:"quality" db:"quality"`
	Format      VideoFormat       `json:"format" db:"format"`
	Thumbnail   string            `json:"thumbnail,omitempty" db:"thumbnail"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Tags        []string          `json:"tags" db:"tags"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty" db:"expires_at"`
}

// RecordingStatistics tracks recording statistics
type RecordingStatistics struct {
	StreamID           string                    `json:"stream_id"`
	TotalRecordings    int64                     `json:"total_recordings"`
	ActiveRecordings   int64                     `json:"active_recordings"`
	TotalDuration      time.Duration             `json:"total_duration"`
	TotalSize          int64                     `json:"total_size"`
	RecordingsByType   map[RecordingType]int64   `json:"recordings_by_type"`
	RecordingsByStatus map[RecordingStatus]int64 `json:"recordings_by_status"`
	RecordingsByHour   map[int]int64             `json:"recordings_by_hour"`
	RecordingsByDay    map[string]int64          `json:"recordings_by_day"`
	AverageFileSize    int64                     `json:"average_file_size"`
	AverageDuration    time.Duration             `json:"average_duration"`
	LastRecording      *time.Time                `json:"last_recording"`
	StartTime          time.Time                 `json:"start_time"`
}

// Repository interfaces

// RecordingRepository defines the interface for recording data access
type RecordingRepository interface {
	// Recording management
	CreateRecording(recording *Recording) error
	GetRecording(id string) (*Recording, error)
	GetRecordings(filters RecordingFilters) ([]*Recording, error)
	UpdateRecording(recording *Recording) error
	DeleteRecording(id string) error

	// Recording clips
	CreateRecordingClip(clip *RecordingClip) error
	GetRecordingClip(id string) (*RecordingClip, error)
	GetRecordingClips(recordingID string) ([]*RecordingClip, error)
	DeleteRecordingClip(id string) error

	// Statistics
	GetRecordingStatistics(streamID string) (*RecordingStatistics, error)
	UpdateRecordingStatistics(stats *RecordingStatistics) error
}

// RecordingService defines the interface for recording business logic
type RecordingService interface {
	// Recording management
	StartRecording(config RecordingConfig) (*Recording, error)
	StopRecording(id string) error
	GetRecording(id string) (*Recording, error)
	GetRecordings(filters RecordingFilters) ([]*Recording, error)
	DeleteRecording(id string) error

	// Event-triggered recording
	TriggerRecording(trigger RecordingTrigger, context RecordingContext) (*Recording, error)
	ProcessAlert(alert *Alert) error
	ProcessZoneEvent(event *ZoneEvent) error

	// Clip management
	CreateClip(recordingID string, startOffset, duration time.Duration, title string) (*RecordingClip, error)
	GetRecordingClips(recordingID string) ([]*RecordingClip, error)
	DeleteClip(id string) error

	// Storage management
	CleanupExpiredRecordings() error
	ArchiveOldRecordings() error
	GetStorageUsage(streamID string) (*StorageUsage, error)
	OptimizeStorage() error

	// Statistics
	GetRecordingStatistics(streamID string) (*RecordingStatistics, error)
	GenerateRecordingReport(filters RecordingFilters, reportType ReportType) (*RecordingReport, error)
}

// VideoRecorder defines the interface for video recording operations
type VideoRecorder interface {
	StartRecording(config RecordingConfig) error
	StopRecording() error
	IsRecording() bool
	GetCurrentRecording() *Recording
	AddFrame(frame *Frame) error
	SetQuality(quality RecordingQuality) error
	GetRecordingStats() *RecordingStats
}

// Supporting types

// RecordingFilters defines filters for querying recordings
type RecordingFilters struct {
	StreamID    string          `json:"stream_id,omitempty"`
	Type        RecordingType   `json:"type,omitempty"`
	Status      RecordingStatus `json:"status,omitempty"`
	Trigger     RecordingTrigger `json:"trigger,omitempty"`
	AlertID     string          `json:"alert_id,omitempty"`
	ZoneID      string          `json:"zone_id,omitempty"`
	StartTime   *time.Time      `json:"start_time,omitempty"`
	EndTime     *time.Time      `json:"end_time,omitempty"`
	MinDuration *time.Duration  `json:"min_duration,omitempty"`
	MaxDuration *time.Duration  `json:"max_duration,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Limit       int             `json:"limit,omitempty"`
	Offset      int             `json:"offset,omitempty"`
}

// RecordingContext provides context for recording triggers
type RecordingContext struct {
	StreamID    string                 `json:"stream_id"`
	AlertID     string                 `json:"alert_id,omitempty"`
	ZoneID      string                 `json:"zone_id,omitempty"`
	ObjectID    string                 `json:"object_id,omitempty"`
	ObjectClass string                 `json:"object_class,omitempty"`
	Confidence  float64                `json:"confidence,omitempty"`
	Position    *Point                 `json:"position,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// StorageUsage tracks storage usage statistics
type StorageUsage struct {
	StreamID        string        `json:"stream_id"`
	TotalSize       int64         `json:"total_size"`
	UsedSize        int64         `json:"used_size"`
	AvailableSize   int64         `json:"available_size"`
	RecordingCount  int64         `json:"recording_count"`
	OldestRecording *time.Time    `json:"oldest_recording"`
	NewestRecording *time.Time    `json:"newest_recording"`
	AverageFileSize int64         `json:"average_file_size"`
	CompressionRatio float64      `json:"compression_ratio"`
	LastCleanup     *time.Time    `json:"last_cleanup"`
}

// RecordingStats tracks real-time recording statistics
type RecordingStats struct {
	FramesRecorded   int64         `json:"frames_recorded"`
	BytesWritten     int64         `json:"bytes_written"`
	Duration         time.Duration `json:"duration"`
	AverageFrameRate float64       `json:"average_frame_rate"`
	CurrentBitrate   int64         `json:"current_bitrate"`
	DroppedFrames    int64         `json:"dropped_frames"`
	LastFrameTime    time.Time     `json:"last_frame_time"`
}

// RecordingReport represents a generated recording report
type RecordingReport struct {
	ID          string                 `json:"id"`
	ReportType  ReportType             `json:"report_type"`
	Filters     RecordingFilters       `json:"filters"`
	Data        map[string]interface{} `json:"data"`
	GeneratedAt time.Time              `json:"generated_at"`
	GeneratedBy string                 `json:"generated_by"`
}

// Validation methods

// Validate validates a recording
func (r *Recording) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("recording ID is required")
	}
	if r.StreamID == "" {
		return fmt.Errorf("stream ID is required")
	}
	if r.Type == "" {
		return fmt.Errorf("recording type is required")
	}
	if r.Status == "" {
		return fmt.Errorf("recording status is required")
	}
	if r.Filename == "" {
		return fmt.Errorf("filename is required")
	}
	if r.FilePath == "" {
		return fmt.Errorf("file path is required")
	}
	return nil
}

// Validate validates a recording config
func (rc *RecordingConfig) Validate() error {
	if rc.StreamID == "" {
		return fmt.Errorf("stream ID is required")
	}
	if rc.Type == "" {
		return fmt.Errorf("recording type is required")
	}
	if rc.StoragePath == "" {
		return fmt.Errorf("storage path is required")
	}
	if rc.MaxDuration <= 0 {
		return fmt.Errorf("max duration must be positive")
	}
	if rc.FrameRate <= 0 {
		return fmt.Errorf("frame rate must be positive")
	}
	if rc.Resolution.Width <= 0 || rc.Resolution.Height <= 0 {
		return fmt.Errorf("resolution must have positive width and height")
	}
	return nil
}

// IsExpired checks if a recording has expired
func (r *Recording) IsExpired() bool {
	if r.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiresAt)
}

// GetDuration returns the duration of the recording
func (r *Recording) GetDuration() time.Duration {
	if r.EndTime != nil {
		return r.EndTime.Sub(r.StartTime)
	}
	if r.Status == RecordingStatusRecording {
		return time.Since(r.StartTime)
	}
	return r.Duration
}

// GetSizeInMB returns the file size in megabytes
func (r *Recording) GetSizeInMB() float64 {
	return float64(r.FileSize) / (1024 * 1024)
}

// MatchesTriggerConditions checks if the context matches trigger conditions
func (tc *TriggerConditions) MatchesTriggerConditions(context RecordingContext) bool {
	// Object class filter
	if len(tc.ObjectClasses) > 0 && context.ObjectClass != "" {
		hasMatchingClass := false
		for _, class := range tc.ObjectClasses {
			if class == context.ObjectClass {
				hasMatchingClass = true
				break
			}
		}
		if !hasMatchingClass {
			return false
		}
	}

	// Confidence filter
	if tc.MinConfidence > 0 && context.Confidence < tc.MinConfidence {
		return false
	}

	// Zone filter
	if len(tc.ZoneIDs) > 0 && context.ZoneID != "" {
		hasMatchingZone := false
		for _, zoneID := range tc.ZoneIDs {
			if zoneID == context.ZoneID {
				hasMatchingZone = true
				break
			}
		}
		if !hasMatchingZone {
			return false
		}
	}

	// Time window filter
	if len(tc.TimeWindows) > 0 {
		inTimeWindow := false
		for _, window := range tc.TimeWindows {
			now := context.Timestamp
			weekday := int(now.Weekday())
			timeStr := now.Format("15:04")

			dayMatch := false
			for _, day := range window.Days {
				if day == weekday {
					dayMatch = true
					break
				}
			}

			if dayMatch && timeStr >= window.StartTime && timeStr <= window.EndTime {
				inTimeWindow = true
				break
			}
		}
		if !inTimeWindow {
			return false
		}
	}

	return true
}
