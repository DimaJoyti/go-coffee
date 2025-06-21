package tracking

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/detection"
	"go.uber.org/zap"
)

// TrackingService integrates object tracking with detection pipeline
type TrackingService struct {
	logger             *zap.Logger
	tracker            *Tracker
	trajectoryRecorder *TrajectoryRecorder
	idManager          *IDManager
	streamTrackers     map[string]*StreamTracker
	config             TrackingServiceConfig
	isRunning          bool
	mutex              sync.RWMutex
	stats              *TrackingServiceStats
}

// TrackingServiceConfig configures the tracking service
type TrackingServiceConfig struct {
	TrackerConfig     TrackerConfig
	TrajectoryConfig  TrajectoryConfig
	IDManagerConfig   IDManagerConfig
	EnableTrajectories bool
	MaxStreams        int
	TrackingCallback  func(*TrackingUpdate)
}

// StreamTracker manages tracking for a single video stream
type StreamTracker struct {
	streamID           string
	logger             *zap.Logger
	tracker            *Tracker
	trajectoryRecorder *TrajectoryRecorder
	isActive           bool
	lastUpdate         time.Time
	frameCount         int64
	mutex              sync.RWMutex
}

// TrackingServiceStats tracks service performance
type TrackingServiceStats struct {
	ActiveStreams     int
	TotalTracks       int64
	ActiveTracks      int
	TotalDetections   int64
	ProcessedFrames   int64
	AverageTracksPerFrame float64
	StartTime         time.Time
	LastUpdate        time.Time
	mutex             sync.RWMutex
}

// TrackingUpdate represents a tracking update event
type TrackingUpdate struct {
	StreamID    string
	FrameID     string
	Tracks      []*Track
	NewTracks   []*Track
	LostTracks  []*Track
	Timestamp   time.Time
	ProcessTime time.Duration
}

// TrackingCallback is called when tracking updates occur
type TrackingCallback func(*TrackingUpdate)

// DefaultTrackingServiceConfig returns default tracking service configuration
func DefaultTrackingServiceConfig() TrackingServiceConfig {
	return TrackingServiceConfig{
		TrackerConfig:      DefaultTrackerConfig(),
		TrajectoryConfig:   DefaultTrajectoryConfig(),
		IDManagerConfig:    DefaultIDManagerConfig(),
		EnableTrajectories: true,
		MaxStreams:         10,
	}
}

// NewTrackingService creates a new tracking service
func NewTrackingService(logger *zap.Logger, config TrackingServiceConfig) *TrackingService {
	// Create ID manager
	idManager := NewIDManager(logger, config.IDManagerConfig)

	// Create global tracker (can be used for cross-stream tracking)
	tracker := NewTracker(logger, config.TrackerConfig)

	// Create trajectory recorder
	var trajectoryRecorder *TrajectoryRecorder
	if config.EnableTrajectories {
		trajectoryRecorder = NewTrajectoryRecorder(logger, config.TrajectoryConfig)
	}

	return &TrackingService{
		logger:             logger.With(zap.String("component", "tracking_service")),
		tracker:            tracker,
		trajectoryRecorder: trajectoryRecorder,
		idManager:          idManager,
		streamTrackers:     make(map[string]*StreamTracker),
		config:             config,
		stats: &TrackingServiceStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the tracking service
func (ts *TrackingService) Start(ctx context.Context) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if ts.isRunning {
		return fmt.Errorf("tracking service is already running")
	}

	ts.logger.Info("Starting tracking service",
		zap.Int("max_streams", ts.config.MaxStreams),
		zap.Bool("trajectories_enabled", ts.config.EnableTrajectories))

	ts.isRunning = true
	ts.stats.StartTime = time.Now()

	ts.logger.Info("Tracking service started")
	return nil
}

// Stop stops the tracking service
func (ts *TrackingService) Stop() error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if !ts.isRunning {
		return fmt.Errorf("tracking service is not running")
	}

	ts.logger.Info("Stopping tracking service")

	// Stop all stream trackers
	for streamID, streamTracker := range ts.streamTrackers {
		ts.logger.Debug("Stopping stream tracker", zap.String("stream_id", streamID))
		streamTracker.Stop()
	}

	// Close components
	if ts.tracker != nil {
		ts.tracker.Close()
	}
	if ts.trajectoryRecorder != nil {
		ts.trajectoryRecorder.Close()
	}
	if ts.idManager != nil {
		ts.idManager.Close()
	}

	ts.isRunning = false

	ts.logger.Info("Tracking service stopped")
	return nil
}

// StartStreamTracking starts tracking for a video stream
func (ts *TrackingService) StartStreamTracking(streamID string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if !ts.isRunning {
		return fmt.Errorf("tracking service is not running")
	}

	if _, exists := ts.streamTrackers[streamID]; exists {
		return fmt.Errorf("tracking already started for stream: %s", streamID)
	}

	if len(ts.streamTrackers) >= ts.config.MaxStreams {
		return fmt.Errorf("maximum number of streams reached: %d", ts.config.MaxStreams)
	}

	ts.logger.Info("Starting stream tracking", zap.String("stream_id", streamID))

	// Create stream-specific tracker
	streamTracker := &StreamTracker{
		streamID:   streamID,
		logger:     ts.logger.With(zap.String("stream_id", streamID)),
		tracker:    NewTracker(ts.logger, ts.config.TrackerConfig),
		isActive:   true,
		lastUpdate: time.Now(),
	}

	// Create stream-specific trajectory recorder if enabled
	if ts.config.EnableTrajectories {
		streamTracker.trajectoryRecorder = NewTrajectoryRecorder(ts.logger, ts.config.TrajectoryConfig)
	}

	ts.streamTrackers[streamID] = streamTracker

	ts.logger.Info("Stream tracking started", zap.String("stream_id", streamID))
	return nil
}

// StopStreamTracking stops tracking for a video stream
func (ts *TrackingService) StopStreamTracking(streamID string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	streamTracker, exists := ts.streamTrackers[streamID]
	if !exists {
		return fmt.Errorf("no tracking running for stream: %s", streamID)
	}

	ts.logger.Info("Stopping stream tracking", zap.String("stream_id", streamID))

	streamTracker.Stop()
	delete(ts.streamTrackers, streamID)

	ts.logger.Info("Stream tracking stopped", zap.String("stream_id", streamID))
	return nil
}

// ProcessDetections processes detection results and updates tracking
func (ts *TrackingService) ProcessDetections(ctx context.Context, streamID, frameID string, detections []domain.DetectedObject) (*TrackingUpdate, error) {
	ts.mutex.RLock()
	streamTracker, exists := ts.streamTrackers[streamID]
	ts.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no tracking running for stream: %s", streamID)
	}

	startTime := time.Now()

	// Get previous tracks for comparison
	previousTracks := streamTracker.tracker.GetActiveTracks()
	previousTrackIDs := make(map[string]bool)
	for _, track := range previousTracks {
		previousTrackIDs[track.ID] = true
	}

	// Update tracker with new detections
	updatedTracks, err := streamTracker.tracker.Update(ctx, detections, frameID)
	if err != nil {
		return nil, fmt.Errorf("failed to update tracker: %w", err)
	}

	// Identify new and lost tracks
	var newTracks, lostTracks []*Track
	currentTrackIDs := make(map[string]bool)

	for _, track := range updatedTracks {
		currentTrackIDs[track.ID] = true
		if !previousTrackIDs[track.ID] {
			newTracks = append(newTracks, track)
		}
	}

	for _, track := range previousTracks {
		if !currentTrackIDs[track.ID] {
			lostTracks = append(lostTracks, track)
		}
	}

	// Record trajectory points if enabled
	if streamTracker.trajectoryRecorder != nil {
		for _, track := range updatedTracks {
			if track.LastDetection != nil {
				centerX, centerY := track.LastDetection.BoundingBox.GetCenter()
				point := TrajectoryPoint{
					X:         centerX,
					Y:         centerY,
					Timestamp: track.LastDetection.Timestamp,
					FrameID:   frameID,
					Velocity:  track.Velocity,
				}
				streamTracker.trajectoryRecorder.RecordPoint(track.ID, point)
			}
		}

		// Remove trajectories for lost tracks
		for _, track := range lostTracks {
			streamTracker.trajectoryRecorder.RemoveTrajectory(track.ID)
		}
	}

	// Update stream tracker stats
	streamTracker.mutex.Lock()
	streamTracker.lastUpdate = time.Now()
	streamTracker.frameCount++
	streamTracker.mutex.Unlock()

	// Create tracking update
	update := &TrackingUpdate{
		StreamID:    streamID,
		FrameID:     frameID,
		Tracks:      updatedTracks,
		NewTracks:   newTracks,
		LostTracks:  lostTracks,
		Timestamp:   time.Now(),
		ProcessTime: time.Since(startTime),
	}

	// Call tracking callback if configured
	if ts.config.TrackingCallback != nil {
		go ts.config.TrackingCallback(update)
	}

	// Update service statistics
	ts.updateStats(len(updatedTracks), len(detections))

	ts.logger.Debug("Tracking update completed",
		zap.String("stream_id", streamID),
		zap.String("frame_id", frameID),
		zap.Int("detections", len(detections)),
		zap.Int("tracks", len(updatedTracks)),
		zap.Int("new_tracks", len(newTracks)),
		zap.Int("lost_tracks", len(lostTracks)),
		zap.Duration("process_time", update.ProcessTime))

	return update, nil
}

// GetStreamTracks returns active tracks for a stream
func (ts *TrackingService) GetStreamTracks(streamID string) ([]*Track, error) {
	ts.mutex.RLock()
	streamTracker, exists := ts.streamTrackers[streamID]
	ts.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no tracking running for stream: %s", streamID)
	}

	return streamTracker.tracker.GetActiveTracks(), nil
}

// GetStreamTrajectory returns trajectory data for a track in a stream
func (ts *TrackingService) GetStreamTrajectory(streamID, trackID string) (*TrajectoryData, error) {
	ts.mutex.RLock()
	streamTracker, exists := ts.streamTrackers[streamID]
	ts.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no tracking running for stream: %s", streamID)
	}

	if streamTracker.trajectoryRecorder == nil {
		return nil, fmt.Errorf("trajectories not enabled for stream: %s", streamID)
	}

	trajectory, exists := streamTracker.trajectoryRecorder.GetTrajectory(trackID)
	if !exists {
		return nil, fmt.Errorf("trajectory not found for track: %s", trackID)
	}

	return trajectory, nil
}

// GetActiveStreams returns list of streams with active tracking
func (ts *TrackingService) GetActiveStreams() []string {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	streams := make([]string, 0, len(ts.streamTrackers))
	for streamID, streamTracker := range ts.streamTrackers {
		if streamTracker.IsActive() {
			streams = append(streams, streamID)
		}
	}
	return streams
}

// GetStats returns tracking service statistics
func (ts *TrackingService) GetStats() *TrackingServiceStats {
	ts.stats.mutex.RLock()
	defer ts.stats.mutex.RUnlock()

	ts.mutex.RLock()
	activeStreams := 0
	activeTracks := 0

	for _, streamTracker := range ts.streamTrackers {
		if streamTracker.IsActive() {
			activeStreams++
			tracks := streamTracker.tracker.GetActiveTracks()
			activeTracks += len(tracks)
		}
	}
	ts.mutex.RUnlock()

	return &TrackingServiceStats{
		ActiveStreams:         activeStreams,
		TotalTracks:           ts.stats.TotalTracks,
		ActiveTracks:          activeTracks,
		TotalDetections:       ts.stats.TotalDetections,
		ProcessedFrames:       ts.stats.ProcessedFrames,
		AverageTracksPerFrame: ts.stats.AverageTracksPerFrame,
		StartTime:             ts.stats.StartTime,
		LastUpdate:            ts.stats.LastUpdate,
	}
}

// CreateDetectionCallback creates a callback for detection results
func (ts *TrackingService) CreateDetectionCallback(streamID string) detection.DetectionCallback {
	return func(response *detection.InferenceResponse) {
		if response.Error != nil {
			ts.logger.Error("Detection error in tracking callback",
				zap.String("stream_id", streamID),
				zap.Error(response.Error))
			return
		}

		// Process detections through tracking
		_, err := ts.ProcessDetections(context.Background(), streamID, response.ID, response.Objects)
		if err != nil {
			ts.logger.Error("Failed to process detections in tracking",
				zap.String("stream_id", streamID),
				zap.Error(err))
		}
	}
}

// updateStats updates service statistics
func (ts *TrackingService) updateStats(trackCount, detectionCount int) {
	ts.stats.mutex.Lock()
	defer ts.stats.mutex.Unlock()

	ts.stats.TotalDetections += int64(detectionCount)
	ts.stats.ProcessedFrames++
	ts.stats.LastUpdate = time.Now()

	// Calculate average tracks per frame
	if ts.stats.ProcessedFrames > 0 {
		ts.stats.AverageTracksPerFrame = float64(ts.stats.TotalTracks) / float64(ts.stats.ProcessedFrames)
	}
}

// Stop stops the stream tracker
func (st *StreamTracker) Stop() {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.isActive = false

	if st.tracker != nil {
		st.tracker.Close()
	}
	if st.trajectoryRecorder != nil {
		st.trajectoryRecorder.Close()
	}

	st.logger.Info("Stream tracker stopped")
}

// IsActive returns whether the stream tracker is active
func (st *StreamTracker) IsActive() bool {
	st.mutex.RLock()
	defer st.mutex.RUnlock()
	return st.isActive
}

// GetFrameCount returns the number of processed frames
func (st *StreamTracker) GetFrameCount() int64 {
	st.mutex.RLock()
	defer st.mutex.RUnlock()
	return st.frameCount
}

// GetLastUpdate returns the last update time
func (st *StreamTracker) GetLastUpdate() time.Time {
	st.mutex.RLock()
	defer st.mutex.RUnlock()
	return st.lastUpdate
}
