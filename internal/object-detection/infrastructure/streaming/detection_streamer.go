package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/detection"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/tracking"
	"go.uber.org/zap"
)

// DetectionStreamer streams detection and tracking results to WebSocket clients
type DetectionStreamer struct {
	logger    *zap.Logger
	hub       *Hub
	config    DetectionStreamerConfig
	isRunning bool
	mutex     sync.RWMutex
	stats     *DetectionStreamerStats
}

// DetectionStreamerConfig configures the detection streamer
type DetectionStreamerConfig struct {
	EnableDetections     bool          // Stream detection results
	EnableTracking       bool          // Stream tracking updates
	EnableFrameOverlays  bool          // Include frame overlay data
	MaxQueueSize         int           // Maximum queue size for buffering
	StreamInterval       time.Duration // Minimum interval between streams
	IncludeConfidence    bool          // Include confidence scores
	IncludeTrajectories  bool          // Include trajectory data
	FilterLowConfidence  bool          // Filter low confidence detections
	MinConfidenceThreshold float64     // Minimum confidence for streaming
}

// DetectionStreamerStats tracks streaming performance
type DetectionStreamerStats struct {
	DetectionsSent     int64
	TrackingUpdatesSent int64
	FramesSent         int64
	BytesSent          int64
	StreamErrors       int64
	QueueOverflows     int64
	StartTime          time.Time
	LastStreamTime     time.Time
	mutex              sync.RWMutex
}

// DetectionStreamData represents detection data for streaming
type DetectionStreamData struct {
	StreamID    string                    `json:"stream_id"`
	FrameID     string                    `json:"frame_id"`
	Timestamp   time.Time                 `json:"timestamp"`
	Detections  []StreamedDetection       `json:"detections"`
	ProcessTime time.Duration             `json:"process_time"`
	FrameSize   *FrameSize                `json:"frame_size,omitempty"`
}

// TrackingStreamData represents tracking data for streaming
type TrackingStreamData struct {
	StreamID     string           `json:"stream_id"`
	FrameID      string           `json:"frame_id"`
	Timestamp    time.Time        `json:"timestamp"`
	Tracks       []StreamedTrack  `json:"tracks"`
	NewTracks    []StreamedTrack  `json:"new_tracks"`
	LostTracks   []StreamedTrack  `json:"lost_tracks"`
	ProcessTime  time.Duration    `json:"process_time"`
}

// StreamedDetection represents a detection optimized for streaming
type StreamedDetection struct {
	ID          string                 `json:"id"`
	Class       string                 `json:"class"`
	Confidence  float64                `json:"confidence"`
	BoundingBox StreamedBoundingBox    `json:"bounding_box"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// StreamedTrack represents a track optimized for streaming
type StreamedTrack struct {
	ID           string                 `json:"id"`
	Class        string                 `json:"class"`
	State        string                 `json:"state"`
	Confidence   float64                `json:"confidence"`
	BoundingBox  StreamedBoundingBox    `json:"bounding_box"`
	Velocity     StreamedVelocity       `json:"velocity"`
	Age          int                    `json:"age"`
	HitStreak    int                    `json:"hit_streak"`
	Trajectory   []StreamedTrajectoryPoint `json:"trajectory,omitempty"`
	Predictions  []StreamedTrajectoryPoint `json:"predictions,omitempty"`
	FirstSeen    time.Time              `json:"first_seen"`
	LastSeen     time.Time              `json:"last_seen"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// StreamedBoundingBox represents a bounding box for streaming
type StreamedBoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// StreamedVelocity represents velocity for streaming
type StreamedVelocity struct {
	VX float64 `json:"vx"`
	VY float64 `json:"vy"`
}

// StreamedTrajectoryPoint represents a trajectory point for streaming
type StreamedTrajectoryPoint struct {
	X         float64          `json:"x"`
	Y         float64          `json:"y"`
	Timestamp time.Time        `json:"timestamp"`
	Velocity  StreamedVelocity `json:"velocity"`
}

// FrameSize represents frame dimensions
type FrameSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DefaultDetectionStreamerConfig returns default configuration
func DefaultDetectionStreamerConfig() DetectionStreamerConfig {
	return DetectionStreamerConfig{
		EnableDetections:       true,
		EnableTracking:         true,
		EnableFrameOverlays:    true,
		MaxQueueSize:           1000,
		StreamInterval:         50 * time.Millisecond, // 20 FPS max
		IncludeConfidence:      true,
		IncludeTrajectories:    true,
		FilterLowConfidence:    true,
		MinConfidenceThreshold: 0.3,
	}
}

// NewDetectionStreamer creates a new detection streamer
func NewDetectionStreamer(logger *zap.Logger, hub *Hub, config DetectionStreamerConfig) *DetectionStreamer {
	return &DetectionStreamer{
		logger: logger.With(zap.String("component", "detection_streamer")),
		hub:    hub,
		config: config,
		stats: &DetectionStreamerStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the detection streamer
func (ds *DetectionStreamer) Start(ctx context.Context) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if ds.isRunning {
		return fmt.Errorf("detection streamer is already running")
	}

	ds.logger.Info("Starting detection streamer",
		zap.Bool("detections_enabled", ds.config.EnableDetections),
		zap.Bool("tracking_enabled", ds.config.EnableTracking),
		zap.Duration("stream_interval", ds.config.StreamInterval))

	ds.isRunning = true
	ds.stats.StartTime = time.Now()

	ds.logger.Info("Detection streamer started")
	return nil
}

// Stop stops the detection streamer
func (ds *DetectionStreamer) Stop() error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if !ds.isRunning {
		return fmt.Errorf("detection streamer is not running")
	}

	ds.logger.Info("Stopping detection streamer")
	ds.isRunning = false

	ds.logger.Info("Detection streamer stopped")
	return nil
}

// StreamDetectionResults streams detection results to clients
func (ds *DetectionStreamer) StreamDetectionResults(response *detection.InferenceResponse) error {
	if !ds.IsRunning() || !ds.config.EnableDetections {
		return nil
	}

	// Filter detections if configured
	detections := response.Objects
	if ds.config.FilterLowConfidence {
		detections = ds.filterDetections(detections)
	}

	// Convert to streaming format
	streamedDetections := ds.convertDetections(detections)

	// Create stream data
	streamData := DetectionStreamData{
		StreamID:    response.StreamID,
		FrameID:     response.ID,
		Timestamp:   response.Timestamp,
		Detections:  streamedDetections,
		ProcessTime: response.ProcessTime,
	}

	// Add frame size if available
	if response.FrameWidth > 0 && response.FrameHeight > 0 {
		streamData.FrameSize = &FrameSize{
			Width:  response.FrameWidth,
			Height: response.FrameHeight,
		}
	}

	// Create message
	message := &Message{
		Type:      MessageTypeDetection,
		StreamID:  response.StreamID,
		Timestamp: time.Now(),
		Data:      streamData,
	}

	// Stream to clients
	err := ds.hub.BroadcastToStream(response.StreamID, message)
	if err != nil {
		ds.updateStats(func(stats *DetectionStreamerStats) {
			stats.StreamErrors++
		})
		return fmt.Errorf("failed to stream detection results: %w", err)
	}

	// Update stats
	ds.updateStats(func(stats *DetectionStreamerStats) {
		stats.DetectionsSent++
		stats.LastStreamTime = time.Now()
		
		// Estimate bytes sent
		data, _ := json.Marshal(streamData)
		stats.BytesSent += int64(len(data))
	})

	ds.logger.Debug("Detection results streamed",
		zap.String("stream_id", response.StreamID),
		zap.String("frame_id", response.ID),
		zap.Int("detections", len(streamedDetections)))

	return nil
}

// StreamTrackingUpdate streams tracking updates to clients
func (ds *DetectionStreamer) StreamTrackingUpdate(update *tracking.TrackingUpdate) error {
	if !ds.IsRunning() || !ds.config.EnableTracking {
		return nil
	}

	// Convert to streaming format
	streamedTracks := ds.convertTracks(update.Tracks)
	streamedNewTracks := ds.convertTracks(update.NewTracks)
	streamedLostTracks := ds.convertTracks(update.LostTracks)

	// Create stream data
	streamData := TrackingStreamData{
		StreamID:    update.StreamID,
		FrameID:     update.FrameID,
		Timestamp:   update.Timestamp,
		Tracks:      streamedTracks,
		NewTracks:   streamedNewTracks,
		LostTracks:  streamedLostTracks,
		ProcessTime: update.ProcessTime,
	}

	// Create message
	message := &Message{
		Type:      MessageTypeTracking,
		StreamID:  update.StreamID,
		Timestamp: time.Now(),
		Data:      streamData,
	}

	// Stream to clients
	err := ds.hub.BroadcastToStream(update.StreamID, message)
	if err != nil {
		ds.updateStats(func(stats *DetectionStreamerStats) {
			stats.StreamErrors++
		})
		return fmt.Errorf("failed to stream tracking update: %w", err)
	}

	// Update stats
	ds.updateStats(func(stats *DetectionStreamerStats) {
		stats.TrackingUpdatesSent++
		stats.LastStreamTime = time.Now()
		
		// Estimate bytes sent
		data, _ := json.Marshal(streamData)
		stats.BytesSent += int64(len(data))
	})

	ds.logger.Debug("Tracking update streamed",
		zap.String("stream_id", update.StreamID),
		zap.String("frame_id", update.FrameID),
		zap.Int("tracks", len(streamedTracks)),
		zap.Int("new_tracks", len(streamedNewTracks)),
		zap.Int("lost_tracks", len(streamedLostTracks)))

	return nil
}

// filterDetections filters detections based on configuration
func (ds *DetectionStreamer) filterDetections(detections []domain.DetectedObject) []domain.DetectedObject {
	if !ds.config.FilterLowConfidence {
		return detections
	}

	filtered := make([]domain.DetectedObject, 0, len(detections))
	for _, detection := range detections {
		if detection.Confidence >= ds.config.MinConfidenceThreshold {
			filtered = append(filtered, detection)
		}
	}
	return filtered
}

// convertDetections converts domain detections to streaming format
func (ds *DetectionStreamer) convertDetections(detections []domain.DetectedObject) []StreamedDetection {
	streamedDetections := make([]StreamedDetection, len(detections))
	
	for i, detection := range detections {
		streamedDetections[i] = StreamedDetection{
			ID:         detection.ID,
			Class:      detection.Class,
			Confidence: detection.Confidence,
			BoundingBox: StreamedBoundingBox{
				X:      detection.BoundingBox.X,
				Y:      detection.BoundingBox.Y,
				Width:  detection.BoundingBox.Width,
				Height: detection.BoundingBox.Height,
			},
			Timestamp: detection.Timestamp,
		}

		// Add metadata if configured
		if ds.config.IncludeConfidence {
			if streamedDetections[i].Metadata == nil {
				streamedDetections[i].Metadata = make(map[string]interface{})
			}
			streamedDetections[i].Metadata["stream_id"] = detection.StreamID
			streamedDetections[i].Metadata["frame_id"] = detection.FrameID
		}
	}

	return streamedDetections
}

// convertTracks converts domain tracks to streaming format
func (ds *DetectionStreamer) convertTracks(tracks []*tracking.Track) []StreamedTrack {
	streamedTracks := make([]StreamedTrack, len(tracks))
	
	for i, track := range tracks {
		streamedTrack := StreamedTrack{
			ID:          track.ID,
			Class:       track.Class,
			State:       track.State.String(),
			Confidence:  track.Confidence,
			Age:         track.Age,
			HitStreak:   track.HitStreak,
			FirstSeen:   track.FirstSeen,
			LastSeen:    track.LastSeen,
			Velocity: StreamedVelocity{
				VX: track.Velocity.VX,
				VY: track.Velocity.VY,
			},
		}

		// Add bounding box from last detection
		if track.LastDetection != nil {
			streamedTrack.BoundingBox = StreamedBoundingBox{
				X:      track.LastDetection.BoundingBox.X,
				Y:      track.LastDetection.BoundingBox.Y,
				Width:  track.LastDetection.BoundingBox.Width,
				Height: track.LastDetection.BoundingBox.Height,
			}
		}

		// Add trajectory if configured
		if ds.config.IncludeTrajectories && len(track.Trajectory) > 0 {
			streamedTrack.Trajectory = ds.convertTrajectoryPoints(track.Trajectory)
		}

		streamedTracks[i] = streamedTrack
	}

	return streamedTracks
}

// convertTrajectoryPoints converts trajectory points to streaming format
func (ds *DetectionStreamer) convertTrajectoryPoints(points []tracking.TrajectoryPoint) []StreamedTrajectoryPoint {
	streamedPoints := make([]StreamedTrajectoryPoint, len(points))
	
	for i, point := range points {
		streamedPoints[i] = StreamedTrajectoryPoint{
			X:         point.X,
			Y:         point.Y,
			Timestamp: point.Timestamp,
			Velocity: StreamedVelocity{
				VX: point.Velocity.VX,
				VY: point.Velocity.VY,
			},
		}
	}

	return streamedPoints
}

// IsRunning returns whether the streamer is running
func (ds *DetectionStreamer) IsRunning() bool {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.isRunning
}

// GetStats returns streamer statistics
func (ds *DetectionStreamer) GetStats() *DetectionStreamerStats {
	ds.stats.mutex.RLock()
	defer ds.stats.mutex.RUnlock()

	return &DetectionStreamerStats{
		DetectionsSent:      ds.stats.DetectionsSent,
		TrackingUpdatesSent: ds.stats.TrackingUpdatesSent,
		FramesSent:          ds.stats.FramesSent,
		BytesSent:           ds.stats.BytesSent,
		StreamErrors:        ds.stats.StreamErrors,
		QueueOverflows:      ds.stats.QueueOverflows,
		StartTime:           ds.stats.StartTime,
		LastStreamTime:      ds.stats.LastStreamTime,
	}
}

// updateStats updates streamer statistics
func (ds *DetectionStreamer) updateStats(updateFunc func(*DetectionStreamerStats)) {
	ds.stats.mutex.Lock()
	defer ds.stats.mutex.Unlock()
	updateFunc(ds.stats)
}

// GetConfig returns the streamer configuration
func (ds *DetectionStreamer) GetConfig() DetectionStreamerConfig {
	return ds.config
}

// UpdateConfig updates the streamer configuration
func (ds *DetectionStreamer) UpdateConfig(config DetectionStreamerConfig) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.config = config
	ds.logger.Info("Detection streamer configuration updated",
		zap.Bool("detections_enabled", config.EnableDetections),
		zap.Bool("tracking_enabled", config.EnableTracking),
		zap.Duration("stream_interval", config.StreamInterval))
}
