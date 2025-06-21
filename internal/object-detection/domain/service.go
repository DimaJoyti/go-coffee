package domain

import (
	"context"
	"io"
)

// DetectionService defines the interface for object detection operations
type DetectionService interface {
	// StartDetection starts object detection on a video stream
	StartDetection(ctx context.Context, streamID string) error
	
	// StopDetection stops object detection on a video stream
	StopDetection(ctx context.Context, streamID string) error
	
	// ProcessFrame processes a single frame for object detection
	ProcessFrame(ctx context.Context, streamID string, frameData []byte) (*DetectionResult, error)
	
	// GetDetectionResults retrieves detection results with filtering
	GetDetectionResults(ctx context.Context, filters DetectionFilters) ([]*DetectionResult, error)
	
	// GetProcessingStats retrieves processing statistics for a stream
	GetProcessingStats(ctx context.Context, streamID string) (*ProcessingStats, error)
	
	// SetDetectionModel sets the active detection model
	SetDetectionModel(ctx context.Context, modelID string) error
	
	// GetActiveModel retrieves the currently active detection model
	GetActiveModel(ctx context.Context) (*DetectionModel, error)
}

// StreamService defines the interface for video stream management
type StreamService interface {
	// CreateStream creates a new video stream
	CreateStream(ctx context.Context, stream *VideoStream) error
	
	// GetStream retrieves a video stream by ID
	GetStream(ctx context.Context, id string) (*VideoStream, error)
	
	// GetAllStreams retrieves all video streams with optional filtering
	GetAllStreams(ctx context.Context, filters StreamFilters) ([]*VideoStream, error)
	
	// UpdateStream updates an existing video stream
	UpdateStream(ctx context.Context, stream *VideoStream) error
	
	// DeleteStream deletes a video stream
	DeleteStream(ctx context.Context, id string) error
	
	// StartStream starts processing a video stream
	StartStream(ctx context.Context, id string) error
	
	// StopStream stops processing a video stream
	StopStream(ctx context.Context, id string) error
	
	// GetStreamStatus retrieves the current status of a stream
	GetStreamStatus(ctx context.Context, id string) (StreamStatus, error)
}

// TrackingService defines the interface for object tracking operations
type TrackingService interface {
	// StartTracking starts tracking objects in a video stream
	StartTracking(ctx context.Context, streamID string) error
	
	// StopTracking stops tracking objects in a video stream
	StopTracking(ctx context.Context, streamID string) error
	
	// UpdateTracking updates tracking data with new detection results
	UpdateTracking(ctx context.Context, streamID string, objects []DetectedObject) error
	
	// GetActiveTracking retrieves all active tracking data for a stream
	GetActiveTracking(ctx context.Context, streamID string) ([]*TrackingData, error)
	
	// GetTrackingHistory retrieves tracking history for an object
	GetTrackingHistory(ctx context.Context, trackingID string) (*TrackingData, error)
	
	// CleanupInactiveTracking removes old inactive tracking data
	CleanupInactiveTracking(ctx context.Context) error
}

// DetectionAlertService defines the interface for detection alert management
type DetectionAlertService interface {
	// CreateAlert creates a new detection alert
	CreateAlert(ctx context.Context, alert *DetectionAlert) error
	
	// GetAlerts retrieves alerts with filtering
	GetAlerts(ctx context.Context, filters AlertFilters) ([]*DetectionAlert, error)
	
	// AcknowledgeAlert marks an alert as acknowledged
	AcknowledgeAlert(ctx context.Context, id string) error
	
	// ProcessDetectionForAlerts processes detection results to generate alerts
	ProcessDetectionForAlerts(ctx context.Context, result *DetectionResult) error
	
	// GetUnacknowledgedAlerts retrieves unacknowledged alerts for a stream
	GetUnacknowledgedAlerts(ctx context.Context, streamID string) ([]*DetectionAlert, error)
}

// ModelService defines the interface for detection model management
type ModelService interface {
	// UploadModel uploads a new detection model
	UploadModel(ctx context.Context, model *DetectionModel, data io.Reader) error
	
	// GetModel retrieves a model by ID
	GetModel(ctx context.Context, id string) (*DetectionModel, error)
	
	// GetAllModels retrieves all available models
	GetAllModels(ctx context.Context) ([]*DetectionModel, error)
	
	// ActivateModel activates a model for detection
	ActivateModel(ctx context.Context, id string) error
	
	// DeleteModel deletes a model
	DeleteModel(ctx context.Context, id string) error
	
	// ValidateModel validates a model file
	ValidateModel(ctx context.Context, data io.Reader) error
}

// VideoProcessor defines the interface for video processing operations
type VideoProcessor interface {
	// OpenStream opens a video stream for processing
	OpenStream(ctx context.Context, source string, streamType StreamType) error
	
	// CloseStream closes a video stream
	CloseStream(ctx context.Context, streamID string) error
	
	// ReadFrame reads the next frame from a video stream
	ReadFrame(ctx context.Context, streamID string) ([]byte, error)
	
	// GetFrameRate retrieves the frame rate of a video stream
	GetFrameRate(ctx context.Context, streamID string) (float64, error)
	
	// GetResolution retrieves the resolution of a video stream
	GetResolution(ctx context.Context, streamID string) (width, height int, err error)
	
	// IsStreamActive checks if a video stream is active
	IsStreamActive(ctx context.Context, streamID string) bool
}

// ObjectDetector defines the interface for object detection algorithms
type ObjectDetector interface {
	// LoadModel loads a detection model
	LoadModel(ctx context.Context, modelPath string) error
	
	// DetectObjects detects objects in an image frame
	DetectObjects(ctx context.Context, frameData []byte) ([]DetectedObject, error)
	
	// SetConfidenceThreshold sets the confidence threshold for detection
	SetConfidenceThreshold(threshold float64)
	
	// GetSupportedClasses retrieves the classes supported by the current model
	GetSupportedClasses() []string
	
	// IsModelLoaded checks if a model is currently loaded
	IsModelLoaded() bool
	
	// Close closes the detector and releases resources
	Close() error
}

// ObjectTracker defines the interface for object tracking algorithms
type ObjectTracker interface {
	// InitializeTracker initializes the tracking system
	InitializeTracker(ctx context.Context) error
	
	// UpdateTracker updates the tracker with new detections
	UpdateTracker(ctx context.Context, detections []DetectedObject) ([]*TrackingData, error)
	
	// GetActiveTracking retrieves all active tracking data
	GetActiveTracking(ctx context.Context) ([]*TrackingData, error)
	
	// RemoveInactiveTracking removes tracking data for objects not seen recently
	RemoveInactiveTracking(ctx context.Context, maxAge int64) error
	
	// Reset resets the tracking system
	Reset(ctx context.Context) error
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	// SendAlert sends an alert notification
	SendAlert(ctx context.Context, alert *DetectionAlert) error
	
	// SendDetectionUpdate sends a detection update notification
	SendDetectionUpdate(ctx context.Context, result *DetectionResult) error
	
	// SendStreamStatusUpdate sends a stream status update notification
	SendStreamStatusUpdate(ctx context.Context, streamID string, status StreamStatus) error
	
	// Subscribe subscribes a client to notifications
	Subscribe(ctx context.Context, clientID string, topics []string) error
	
	// Unsubscribe unsubscribes a client from notifications
	Unsubscribe(ctx context.Context, clientID string) error
}
