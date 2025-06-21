package domain

import (
	"context"
	"time"
)

// StreamRepository defines the interface for video stream persistence
type StreamRepository interface {
	// Create creates a new video stream
	Create(ctx context.Context, stream *VideoStream) error
	
	// GetByID retrieves a video stream by ID
	GetByID(ctx context.Context, id string) (*VideoStream, error)
	
	// GetAll retrieves all video streams with optional filtering
	GetAll(ctx context.Context, filters StreamFilters) ([]*VideoStream, error)
	
	// Update updates an existing video stream
	Update(ctx context.Context, stream *VideoStream) error
	
	// Delete deletes a video stream by ID
	Delete(ctx context.Context, id string) error
	
	// UpdateStatus updates the status of a video stream
	UpdateStatus(ctx context.Context, id string, status StreamStatus) error
	
	// UpdateLastFrameTime updates the last frame timestamp
	UpdateLastFrameTime(ctx context.Context, id string, timestamp time.Time) error
}

// DetectionRepository defines the interface for detection result persistence
type DetectionRepository interface {
	// SaveResult saves a detection result
	SaveResult(ctx context.Context, result *DetectionResult) error
	
	// GetResults retrieves detection results with filtering and pagination
	GetResults(ctx context.Context, filters DetectionFilters) ([]*DetectionResult, error)
	
	// GetResultsByStream retrieves detection results for a specific stream
	GetResultsByStream(ctx context.Context, streamID string, limit int, offset int) ([]*DetectionResult, error)
	
	// GetLatestResult retrieves the latest detection result for a stream
	GetLatestResult(ctx context.Context, streamID string) (*DetectionResult, error)
	
	// DeleteOldResults deletes detection results older than the specified duration
	DeleteOldResults(ctx context.Context, olderThan time.Duration) error
	
	// GetStats retrieves processing statistics for a stream
	GetStats(ctx context.Context, streamID string) (*ProcessingStats, error)
}

// TrackingRepository defines the interface for object tracking persistence
type TrackingRepository interface {
	// SaveTrackingData saves tracking data for an object
	SaveTrackingData(ctx context.Context, data *TrackingData) error
	
	// GetTrackingData retrieves tracking data by ID
	GetTrackingData(ctx context.Context, id string) (*TrackingData, error)
	
	// GetActiveTracking retrieves all active tracking data for a stream
	GetActiveTracking(ctx context.Context, streamID string) ([]*TrackingData, error)
	
	// UpdateTrackingData updates existing tracking data
	UpdateTrackingData(ctx context.Context, data *TrackingData) error
	
	// DeactivateTracking marks tracking data as inactive
	DeactivateTracking(ctx context.Context, id string) error
	
	// CleanupInactiveTracking removes old inactive tracking data
	CleanupInactiveTracking(ctx context.Context, olderThan time.Duration) error
}

// DetectionAlertRepository defines the interface for detection alert persistence
type DetectionAlertRepository interface {
	// SaveAlert saves a detection alert
	SaveAlert(ctx context.Context, alert *DetectionAlert) error
	
	// GetAlerts retrieves alerts with filtering and pagination
	GetAlerts(ctx context.Context, filters DetectionAlertFilters) ([]*DetectionAlert, error)
	
	// GetUnacknowledgedAlerts retrieves unacknowledged alerts
	GetUnacknowledgedAlerts(ctx context.Context, streamID string) ([]*DetectionAlert, error)
	
	// AcknowledgeAlert marks an alert as acknowledged
	AcknowledgeAlert(ctx context.Context, id string) error
	
	// DeleteOldAlerts deletes alerts older than the specified duration
	DeleteOldAlerts(ctx context.Context, olderThan time.Duration) error
}

// ModelRepository defines the interface for detection model persistence
type ModelRepository interface {
	// SaveModel saves a detection model
	SaveModel(ctx context.Context, model *DetectionModel) error
	
	// GetModel retrieves a model by ID
	GetModel(ctx context.Context, id string) (*DetectionModel, error)
	
	// GetActiveModel retrieves the currently active model
	GetActiveModel(ctx context.Context) (*DetectionModel, error)
	
	// GetAllModels retrieves all available models
	GetAllModels(ctx context.Context) ([]*DetectionModel, error)
	
	// ActivateModel activates a model and deactivates others
	ActivateModel(ctx context.Context, id string) error
	
	// DeleteModel deletes a model by ID
	DeleteModel(ctx context.Context, id string) error
}

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	// Set stores a value in cache with expiration
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, error)
	
	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error
	
	// SetStreamStatus caches stream status
	SetStreamStatus(ctx context.Context, streamID string, status StreamStatus) error
	
	// GetStreamStatus retrieves cached stream status
	GetStreamStatus(ctx context.Context, streamID string) (StreamStatus, error)
	
	// SetDetectionResult caches latest detection result
	SetDetectionResult(ctx context.Context, streamID string, result *DetectionResult) error
	
	// GetDetectionResult retrieves cached detection result
	GetDetectionResult(ctx context.Context, streamID string) (*DetectionResult, error)
}

// Filter types for repository queries
type StreamFilters struct {
	Status []StreamStatus `json:"status,omitempty"`
	Type   []StreamType   `json:"type,omitempty"`
	Limit  int            `json:"limit,omitempty"`
	Offset int            `json:"offset,omitempty"`
}

type DetectionFilters struct {
	StreamID  string    `json:"stream_id,omitempty"`
	Class     []string  `json:"class,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Offset    int       `json:"offset,omitempty"`
}

type DetectionAlertFilters struct {
	StreamID      string    `json:"stream_id,omitempty"`
	AlertType     []string  `json:"alert_type,omitempty"`
	Severity      []string  `json:"severity,omitempty"`
	Acknowledged  *bool     `json:"acknowledged,omitempty"`
	StartTime     time.Time `json:"start_time,omitempty"`
	EndTime       time.Time `json:"end_time,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Offset        int       `json:"offset,omitempty"`
}
