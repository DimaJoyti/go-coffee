package tracking

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
)

// Tracker implements object tracking across video frames
type Tracker struct {
	logger       *zap.Logger
	tracks       map[string]*Track
	nextTrackID  uint64
	config       TrackerConfig
	mutex        sync.RWMutex
	stats        *TrackerStats
}

// TrackerConfig configures the object tracker
type TrackerConfig struct {
	MaxAge           int     // Maximum frames a track can exist without detection
	MinHits          int     // Minimum detections before a track is confirmed
	IOUThreshold     float64 // IOU threshold for association
	MaxDistance      float64 // Maximum distance for association
	EnablePrediction bool    // Enable Kalman filter prediction
	TrackTimeout     time.Duration // Time after which inactive tracks are removed
}

// TrackerStats tracks performance metrics
type TrackerStats struct {
	ActiveTracks     int
	TotalTracks      int64
	ConfirmedTracks  int64
	DeletedTracks    int64
	AssociationCount int64
	PredictionCount  int64
	StartTime        time.Time
	LastUpdateTime   time.Time
	mutex            sync.RWMutex
}

// Track represents a tracked object across frames
type Track struct {
	ID               string
	Class            string
	State            TrackState
	Detections       []domain.DetectedObject
	Predictions      []Prediction
	Trajectory       []TrajectoryPoint
	FirstSeen        time.Time
	LastSeen         time.Time
	LastDetection    *domain.DetectedObject
	LastPrediction   *Prediction
	HitStreak        int  // Consecutive frames with detections
	TimeSinceUpdate  int  // Frames since last detection
	Age              int  // Total frames since creation
	Confidence       float64
	Velocity         Velocity
	KalmanFilter     *KalmanFilter
	mutex            sync.RWMutex
}

// TrackState represents the state of a track
type TrackState int

const (
	TrackStateTentative TrackState = iota // New track, not yet confirmed
	TrackStateConfirmed                   // Confirmed track with enough detections
	TrackStateDeleted                     // Track marked for deletion
)

// Prediction represents a predicted object position
type Prediction struct {
	BoundingBox domain.Rectangle
	Confidence  float64
	Timestamp   time.Time
	FrameID     string
}

// TrajectoryPoint represents a point in the object's trajectory
type TrajectoryPoint struct {
	X         float64
	Y         float64
	Timestamp time.Time
	FrameID   string
	Velocity  Velocity
}

// Velocity represents object velocity
type Velocity struct {
	VX float64 // Velocity in X direction (pixels/second)
	VY float64 // Velocity in Y direction (pixels/second)
}

// Association represents a detection-track association
type Association struct {
	DetectionIndex int
	TrackID        string
	Distance       float64
	IOU            float64
	Confidence     float64
}

// DefaultTrackerConfig returns default tracker configuration
func DefaultTrackerConfig() TrackerConfig {
	return TrackerConfig{
		MaxAge:           30,  // 30 frames without detection
		MinHits:          3,   // 3 detections to confirm track
		IOUThreshold:     0.3, // 30% IOU for association
		MaxDistance:      100, // 100 pixels maximum distance
		EnablePrediction: true,
		TrackTimeout:     5 * time.Second,
	}
}

// NewTracker creates a new object tracker
func NewTracker(logger *zap.Logger, config TrackerConfig) *Tracker {
	return &Tracker{
		logger:      logger.With(zap.String("component", "tracker")),
		tracks:      make(map[string]*Track),
		nextTrackID: 1,
		config:      config,
		stats: &TrackerStats{
			StartTime: time.Now(),
		},
	}
}

// Update updates the tracker with new detections
func (t *Tracker) Update(ctx context.Context, detections []domain.DetectedObject, frameID string) ([]*Track, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.logger.Debug("Updating tracker",
		zap.Int("detections", len(detections)),
		zap.String("frame_id", frameID),
		zap.Int("active_tracks", len(t.tracks)))

	// Predict next positions for existing tracks
	t.predict(frameID)

	// Associate detections with existing tracks
	associations, unassociatedDetections, unassociatedTracks := t.associate(detections)

	// Update associated tracks
	for _, assoc := range associations {
		track := t.tracks[assoc.TrackID]
		detection := detections[assoc.DetectionIndex]
		t.updateTrack(track, &detection, frameID)
	}

	// Create new tracks for unassociated detections
	for _, detIdx := range unassociatedDetections {
		detection := detections[detIdx]
		t.createTrack(&detection, frameID)
	}

	// Mark unassociated tracks for potential deletion
	for _, trackID := range unassociatedTracks {
		track := t.tracks[trackID]
		t.markTrackMissed(track, frameID)
	}

	// Clean up old tracks
	t.cleanupTracks()

	// Update statistics
	t.updateStats()

	// Return active tracks
	activeTracks := t.getActiveTracks()

	t.logger.Debug("Tracker update completed",
		zap.Int("active_tracks", len(activeTracks)),
		zap.Int("associations", len(associations)),
		zap.Int("new_tracks", len(unassociatedDetections)))

	return activeTracks, nil
}

// GetActiveTracks returns all active tracks
func (t *Tracker) GetActiveTracks() []*Track {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.getActiveTracks()
}

// GetTrack returns a specific track by ID
func (t *Tracker) GetTrack(trackID string) (*Track, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	track, exists := t.tracks[trackID]
	if !exists {
		return nil, false
	}

	// Return track data without copying the mutex
	return track, true
}

// GetTrackHistory returns the trajectory history for a track
func (t *Tracker) GetTrackHistory(trackID string) ([]TrajectoryPoint, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	track, exists := t.tracks[trackID]
	if !exists {
		return nil, false
	}

	// Return a copy of the trajectory
	trajectory := make([]TrajectoryPoint, len(track.Trajectory))
	copy(trajectory, track.Trajectory)
	return trajectory, true
}

// GetStats returns tracker statistics
func (t *Tracker) GetStats() *TrackerStats {
	t.stats.mutex.RLock()
	defer t.stats.mutex.RUnlock()

	// Return a copy
	return &TrackerStats{
		ActiveTracks:     t.stats.ActiveTracks,
		TotalTracks:      t.stats.TotalTracks,
		ConfirmedTracks:  t.stats.ConfirmedTracks,
		DeletedTracks:    t.stats.DeletedTracks,
		AssociationCount: t.stats.AssociationCount,
		PredictionCount:  t.stats.PredictionCount,
		StartTime:        t.stats.StartTime,
		LastUpdateTime:   t.stats.LastUpdateTime,
	}
}

// getActiveTracks returns active tracks (internal method)
func (t *Tracker) getActiveTracks() []*Track {
	var activeTracks []*Track
	for _, track := range t.tracks {
		if track.State != TrackStateDeleted {
			activeTracks = append(activeTracks, track)
		}
	}
	return activeTracks
}

// generateTrackID generates a unique track ID
func (t *Tracker) generateTrackID() string {
	id := t.nextTrackID
	t.nextTrackID++
	return generateTrackIDString(id)
}

// updateStats updates tracker statistics
func (t *Tracker) updateStats() {
	t.stats.mutex.Lock()
	defer t.stats.mutex.Unlock()

	activeTracks := 0
	confirmedTracks := int64(0)

	for _, track := range t.tracks {
		if track.State != TrackStateDeleted {
			activeTracks++
			if track.State == TrackStateConfirmed {
				confirmedTracks++
			}
		}
	}

	t.stats.ActiveTracks = activeTracks
	t.stats.ConfirmedTracks = confirmedTracks
	t.stats.LastUpdateTime = time.Now()
}

// String returns string representation of track state
func (ts TrackState) String() string {
	switch ts {
	case TrackStateTentative:
		return "tentative"
	case TrackStateConfirmed:
		return "confirmed"
	case TrackStateDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}



// generateTrackIDString generates a track ID string from a number
func generateTrackIDString(id uint64) string {
	return fmt.Sprintf("track_%d", id)
}

// predict predicts next positions for existing tracks
func (t *Tracker) predict(frameID string) {
	for _, track := range t.tracks {
		if track.State == TrackStateDeleted {
			continue
		}

		if track.KalmanFilter != nil && track.KalmanFilter.IsInitialized() {
			predictedBox := track.KalmanFilter.Predict()

			prediction := Prediction{
				BoundingBox: predictedBox,
				Confidence:  track.Confidence * 0.9, // Decay confidence over time
				Timestamp:   time.Now(),
				FrameID:     frameID,
			}

			track.LastPrediction = &prediction
			track.Predictions = append(track.Predictions, prediction)

			// Limit prediction history
			if len(track.Predictions) > 10 {
				track.Predictions = track.Predictions[1:]
			}
		}

		track.Age++
		track.TimeSinceUpdate++
	}
}

// associate associates detections with existing tracks
func (t *Tracker) associate(detections []domain.DetectedObject) ([]Association, []int, []string) {
	// Get active tracks
	activeTracks := make([]*Track, 0)
	for _, track := range t.tracks {
		if track.State != TrackStateDeleted {
			activeTracks = append(activeTracks, track)
		}
	}

	// Create associator
	associator := NewAssociator(t.logger, DefaultAssociationConfig())

	// Perform association
	return associator.Associate(detections, activeTracks)
}

// updateTrack updates a track with a new detection
func (t *Tracker) updateTrack(track *Track, detection *domain.DetectedObject, frameID string) {
	track.mutex.Lock()
	defer track.mutex.Unlock()

	// Update Kalman filter
	if track.KalmanFilter != nil {
		track.KalmanFilter.Update(*detection)
	}

	// Update track properties
	track.LastDetection = detection
	track.LastSeen = time.Now()
	track.HitStreak++
	track.TimeSinceUpdate = 0
	track.Confidence = detection.Confidence

	// Update velocity calculation
	if len(track.Detections) > 0 {
		lastDetection := track.Detections[len(track.Detections)-1]
		dt := detection.Timestamp.Sub(lastDetection.Timestamp).Seconds()
		if dt > 0 {
			// Calculate velocity based on center movement
			center1X, center1Y := lastDetection.BoundingBox.GetCenter()
			center2X, center2Y := detection.BoundingBox.GetCenter()

			track.Velocity.VX = (center2X - center1X) / dt
			track.Velocity.VY = (center2Y - center1Y) / dt
		}
	}

	// Add detection to history
	track.Detections = append(track.Detections, *detection)

	// Limit detection history
	if len(track.Detections) > 30 {
		track.Detections = track.Detections[1:]
	}

	// Add trajectory point
	centerX, centerY := detection.BoundingBox.GetCenter()
	trajectoryPoint := TrajectoryPoint{
		X:         centerX,
		Y:         centerY,
		Timestamp: detection.Timestamp,
		FrameID:   frameID,
		Velocity:  track.Velocity,
	}
	track.Trajectory = append(track.Trajectory, trajectoryPoint)

	// Limit trajectory history
	if len(track.Trajectory) > 100 {
		track.Trajectory = track.Trajectory[1:]
	}

	// Update track state
	if track.State == TrackStateTentative && track.HitStreak >= t.config.MinHits {
		track.State = TrackStateConfirmed
		t.logger.Debug("Track confirmed", zap.String("track_id", track.ID))
	}
}

// createTrack creates a new track from a detection
func (t *Tracker) createTrack(detection *domain.DetectedObject, frameID string) {
	trackID := t.generateTrackID()

	// Create Kalman filter
	var kalmanFilter *KalmanFilter
	if t.config.EnablePrediction {
		kalmanFilter = NewKalmanFilter(DefaultKalmanConfig())
		kalmanFilter.Initialize(*detection)
	}

	// Calculate initial position
	centerX, centerY := detection.BoundingBox.GetCenter()

	track := &Track{
		ID:               trackID,
		Class:            detection.Class,
		State:            TrackStateTentative,
		Detections:       []domain.DetectedObject{*detection},
		Predictions:      []Prediction{},
		Trajectory:       []TrajectoryPoint{},
		FirstSeen:        time.Now(),
		LastSeen:         time.Now(),
		LastDetection:    detection,
		HitStreak:        1,
		TimeSinceUpdate:  0,
		Age:              1,
		Confidence:       detection.Confidence,
		Velocity:         Velocity{VX: 0, VY: 0},
		KalmanFilter:     kalmanFilter,
	}

	// Add initial trajectory point
	trajectoryPoint := TrajectoryPoint{
		X:         centerX,
		Y:         centerY,
		Timestamp: detection.Timestamp,
		FrameID:   frameID,
		Velocity:  track.Velocity,
	}
	track.Trajectory = append(track.Trajectory, trajectoryPoint)

	t.tracks[trackID] = track
	t.stats.TotalTracks++

	t.logger.Debug("Track created",
		zap.String("track_id", trackID),
		zap.String("class", detection.Class),
		zap.Float64("confidence", detection.Confidence))
}

// markTrackMissed marks a track as missed (no detection)
func (t *Tracker) markTrackMissed(track *Track, frameID string) {
	track.mutex.Lock()
	defer track.mutex.Unlock()

	track.HitStreak = 0
	track.TimeSinceUpdate++
	track.Age++

	// Decay confidence
	track.Confidence *= 0.8

	t.logger.Debug("Track missed",
		zap.String("track_id", track.ID),
		zap.Int("time_since_update", track.TimeSinceUpdate))
}

// cleanupTracks removes old or invalid tracks
func (t *Tracker) cleanupTracks() {
	var tracksToDelete []string

	for trackID, track := range t.tracks {
		shouldDelete := false

		// Delete tracks that are too old
		if track.TimeSinceUpdate > t.config.MaxAge {
			shouldDelete = true
			t.logger.Debug("Track deleted due to age",
				zap.String("track_id", trackID),
				zap.Int("time_since_update", track.TimeSinceUpdate))
		}

		// Delete tracks with very low confidence
		if track.Confidence < 0.1 {
			shouldDelete = true
			t.logger.Debug("Track deleted due to low confidence",
				zap.String("track_id", trackID),
				zap.Float64("confidence", track.Confidence))
		}

		// Delete tracks that have been inactive for too long
		if time.Since(track.LastSeen) > t.config.TrackTimeout {
			shouldDelete = true
			t.logger.Debug("Track deleted due to timeout",
				zap.String("track_id", trackID),
				zap.Duration("inactive_time", time.Since(track.LastSeen)))
		}

		if shouldDelete {
			track.State = TrackStateDeleted
			tracksToDelete = append(tracksToDelete, trackID)
		}
	}

	// Remove deleted tracks
	for _, trackID := range tracksToDelete {
		delete(t.tracks, trackID)
		t.stats.DeletedTracks++
	}
}

// Close closes the tracker and cleans up resources
func (t *Tracker) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.logger.Info("Closing tracker")

	// Clean up all tracks
	for trackID, track := range t.tracks {
		if track.KalmanFilter != nil {
			// Clean up Kalman filter resources if needed
		}
		delete(t.tracks, trackID)
	}

	t.logger.Info("Tracker closed")
	return nil
}
