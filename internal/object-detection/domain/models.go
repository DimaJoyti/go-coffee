package domain

import (
	"time"
)

// StreamType represents the type of video stream
type StreamType string

const (
	StreamTypeWebcam StreamType = "webcam"
	StreamTypeFile   StreamType = "file"
	StreamTypeRTMP   StreamType = "rtmp"
	StreamTypeHTTP   StreamType = "http"
)

// StreamStatus represents the current status of a video stream
type StreamStatus string

const (
	StreamStatusIdle       StreamStatus = "idle"
	StreamStatusActive     StreamStatus = "active"
	StreamStatusProcessing StreamStatus = "processing"
	StreamStatusError      StreamStatus = "error"
	StreamStatusStopped    StreamStatus = "stopped"
)

// Rectangle represents a bounding box for detected objects
type Rectangle struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// GetCenter returns the center point of the rectangle
func (r Rectangle) GetCenter() (float64, float64) {
	centerX := float64(r.X) + float64(r.Width)/2
	centerY := float64(r.Y) + float64(r.Height)/2
	return centerX, centerY
}

// Resolution represents video resolution
type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Frame represents a video frame with metadata
type Frame struct {
	ID        string    `json:"id"`
	StreamID  string    `json:"stream_id"`
	Data      []byte    `json:"data"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	FrameNum  int       `json:"frame_num"`
}

// DetectedObject represents an object detected in a video frame
type DetectedObject struct {
	ID          string     `json:"id" db:"id"`
	Class       string     `json:"class" db:"class"`
	Confidence  float64    `json:"confidence" db:"confidence"`
	BoundingBox Rectangle  `json:"bounding_box" db:"bounding_box"`
	Timestamp   time.Time  `json:"timestamp" db:"timestamp"`
	TrackingID  *string    `json:"tracking_id,omitempty" db:"tracking_id"`
	StreamID    string     `json:"stream_id" db:"stream_id"`
	FrameID     string     `json:"frame_id" db:"frame_id"`
}

// TrackingData represents tracking information for an object across frames
type TrackingData struct {
	ID           string      `json:"id" db:"id"`
	ObjectClass  string      `json:"object_class" db:"object_class"`
	FirstSeen    time.Time   `json:"first_seen" db:"first_seen"`
	LastSeen     time.Time   `json:"last_seen" db:"last_seen"`
	Trajectory   []Rectangle `json:"trajectory" db:"trajectory"`
	Velocity     *Velocity   `json:"velocity,omitempty" db:"velocity"`
	IsActive     bool        `json:"is_active" db:"is_active"`
	StreamID     string      `json:"stream_id" db:"stream_id"`
}

// Velocity represents the movement velocity of a tracked object
type Velocity struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// StreamConfig represents configuration for a video stream
type StreamConfig struct {
	FPS                int     `json:"fps" yaml:"fps"`
	Resolution         string  `json:"resolution" yaml:"resolution"`
	DetectionThreshold float64 `json:"detection_threshold" yaml:"detection_threshold"`
	TrackingEnabled    bool    `json:"tracking_enabled" yaml:"tracking_enabled"`
	RecordingEnabled   bool    `json:"recording_enabled" yaml:"recording_enabled"`
	AlertsEnabled      bool    `json:"alerts_enabled" yaml:"alerts_enabled"`
}

// VideoStream represents a video stream source
type VideoStream struct {
	ID          string       `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Source      string       `json:"source" db:"source"`
	Type        StreamType   `json:"type" db:"type"`
	Status      StreamStatus `json:"status" db:"status"`
	Config      StreamConfig `json:"config" db:"config"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	LastFrameAt *time.Time   `json:"last_frame_at,omitempty" db:"last_frame_at"`
}

// DetectionResult represents the result of object detection on a frame
type DetectionResult struct {
	ID          string           `json:"id" db:"id"`
	StreamID    string           `json:"stream_id" db:"stream_id"`
	FrameID     string           `json:"frame_id" db:"frame_id"`
	Objects     []DetectedObject `json:"objects" db:"objects"`
	ProcessTime time.Duration    `json:"process_time" db:"process_time"`
	Timestamp   time.Time        `json:"timestamp" db:"timestamp"`
	FrameWidth  int              `json:"frame_width" db:"frame_width"`
	FrameHeight int              `json:"frame_height" db:"frame_height"`
}

// DetectionAlert represents an alert triggered by object detection
type DetectionAlert struct {
	ID          string    `json:"id" db:"id"`
	StreamID    string    `json:"stream_id" db:"stream_id"`
	ObjectID    string    `json:"object_id" db:"object_id"`
	AlertType   string    `json:"alert_type" db:"alert_type"`
	Message     string    `json:"message" db:"message"`
	Severity    string    `json:"severity" db:"severity"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Acknowledged bool     `json:"acknowledged" db:"acknowledged"`
}

// DetectionModel represents a machine learning model for object detection
type DetectionModel struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Version     string    `json:"version" db:"version"`
	Type        string    `json:"type" db:"type"` // yolo, ssd, etc.
	FilePath    string    `json:"file_path" db:"file_path"`
	FileHash    string    `json:"file_hash" db:"file_hash"`
	Classes     []string  `json:"classes" db:"classes"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ProcessingStats represents statistics for detection processing
type ProcessingStats struct {
	StreamID           string        `json:"stream_id"`
	TotalFrames        int64         `json:"total_frames"`
	ProcessedFrames    int64         `json:"processed_frames"`
	DetectedObjects    int64         `json:"detected_objects"`
	AverageProcessTime time.Duration `json:"average_process_time"`
	FPS                float64       `json:"fps"`
	LastProcessedAt    time.Time     `json:"last_processed_at"`
}
