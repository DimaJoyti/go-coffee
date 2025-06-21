package tracking

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewTracker(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()

	tracker := NewTracker(logger, config)

	assert.NotNil(t, tracker)
	assert.Equal(t, config, tracker.config)
	assert.Empty(t, tracker.tracks)
	assert.Equal(t, uint64(1), tracker.nextTrackID)
}

func TestDefaultTrackerConfig(t *testing.T) {
	config := DefaultTrackerConfig()

	assert.Equal(t, 30, config.MaxAge)
	assert.Equal(t, 3, config.MinHits)
	assert.Equal(t, 0.3, config.IOUThreshold)
	assert.Equal(t, 100.0, config.MaxDistance)
	assert.True(t, config.EnablePrediction)
	assert.Equal(t, 5*time.Second, config.TrackTimeout)
}

func TestTracker_Update_EmptyDetections(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	ctx := context.Background()
	detections := []domain.DetectedObject{}

	tracks, err := tracker.Update(ctx, detections, "frame_1")

	assert.NoError(t, err)
	assert.Empty(t, tracks)
	assert.Empty(t, tracker.tracks)
}

func TestTracker_Update_NewDetections(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	ctx := context.Background()
	detections := []domain.DetectedObject{
		{
			ID:         "det_1",
			Class:      "person",
			Confidence: 0.9,
			BoundingBox: domain.Rectangle{
				X: 10, Y: 20, Width: 50, Height: 100,
			},
			Timestamp: time.Now(),
		},
		{
			ID:         "det_2",
			Class:      "car",
			Confidence: 0.8,
			BoundingBox: domain.Rectangle{
				X: 100, Y: 200, Width: 80, Height: 60,
			},
			Timestamp: time.Now(),
		},
	}

	tracks, err := tracker.Update(ctx, detections, "frame_1")

	assert.NoError(t, err)
	assert.Len(t, tracks, 2)
	assert.Len(t, tracker.tracks, 2)

	// Check track properties
	for _, track := range tracks {
		assert.Equal(t, TrackStateTentative, track.State)
		assert.Equal(t, 1, track.HitStreak)
		assert.Equal(t, 0, track.TimeSinceUpdate)
		assert.Equal(t, 1, track.Age)
		assert.Len(t, track.Detections, 1)
		assert.Len(t, track.Trajectory, 1)
	}
}

func TestTracker_Update_TrackConfirmation(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	config.MinHits = 2 // Reduce for testing
	tracker := NewTracker(logger, config)

	ctx := context.Background()
	detection := domain.DetectedObject{
		ID:         "det_1",
		Class:      "person",
		Confidence: 0.9,
		BoundingBox: domain.Rectangle{
			X: 10, Y: 20, Width: 50, Height: 100,
		},
		Timestamp: time.Now(),
	}

	// First detection - creates tentative track
	tracks, err := tracker.Update(ctx, []domain.DetectedObject{detection}, "frame_1")
	require.NoError(t, err)
	require.Len(t, tracks, 1)
	assert.Equal(t, TrackStateTentative, tracks[0].State)

	// Second detection - should confirm track
	detection.Timestamp = time.Now().Add(100 * time.Millisecond)
	tracks, err = tracker.Update(ctx, []domain.DetectedObject{detection}, "frame_2")
	require.NoError(t, err)
	require.Len(t, tracks, 1)
	assert.Equal(t, TrackStateConfirmed, tracks[0].State)
	assert.Equal(t, 2, tracks[0].HitStreak)
}

func TestTracker_Update_TrackAssociation(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	ctx := context.Background()

	// First frame - create track
	detection1 := domain.DetectedObject{
		ID:         "det_1",
		Class:      "person",
		Confidence: 0.9,
		BoundingBox: domain.Rectangle{
			X: 10, Y: 20, Width: 50, Height: 100,
		},
		Timestamp: time.Now(),
	}

	tracks, err := tracker.Update(ctx, []domain.DetectedObject{detection1}, "frame_1")
	require.NoError(t, err)
	require.Len(t, tracks, 1)
	trackID := tracks[0].ID

	// Second frame - similar detection should associate with existing track
	detection2 := domain.DetectedObject{
		ID:         "det_2",
		Class:      "person",
		Confidence: 0.85,
		BoundingBox: domain.Rectangle{
			X: 15, Y: 25, Width: 50, Height: 100, // Slightly moved
		},
		Timestamp: time.Now().Add(100 * time.Millisecond),
	}

	tracks, err = tracker.Update(ctx, []domain.DetectedObject{detection2}, "frame_2")
	require.NoError(t, err)
	require.Len(t, tracks, 1)
	assert.Equal(t, trackID, tracks[0].ID) // Same track ID
	assert.Equal(t, 2, tracks[0].HitStreak)
	assert.Len(t, tracks[0].Detections, 2)
	assert.Len(t, tracks[0].Trajectory, 2)
}

func TestTracker_Update_TrackDeletion(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	config.MaxAge = 2 // Reduce for testing
	tracker := NewTracker(logger, config)

	ctx := context.Background()

	// Create track
	detection := domain.DetectedObject{
		ID:         "det_1",
		Class:      "person",
		Confidence: 0.9,
		BoundingBox: domain.Rectangle{
			X: 10, Y: 20, Width: 50, Height: 100,
		},
		Timestamp: time.Now(),
	}

	tracks, err := tracker.Update(ctx, []domain.DetectedObject{detection}, "frame_1")
	require.NoError(t, err)
	require.Len(t, tracks, 1)

	// Update without detections - track should age
	tracks, err = tracker.Update(ctx, []domain.DetectedObject{}, "frame_2")
	require.NoError(t, err)
	require.Len(t, tracks, 1)
	assert.Equal(t, 1, tracks[0].TimeSinceUpdate)

	// Another update without detections - track should age more
	tracks, err = tracker.Update(ctx, []domain.DetectedObject{}, "frame_3")
	require.NoError(t, err)
	require.Len(t, tracks, 1)
	assert.Equal(t, 2, tracks[0].TimeSinceUpdate)

	// Third update without detections - track should be deleted
	tracks, err = tracker.Update(ctx, []domain.DetectedObject{}, "frame_4")
	require.NoError(t, err)
	assert.Empty(t, tracks) // Track should be deleted
}

func TestTracker_GetActiveTracks(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	ctx := context.Background()
	detections := []domain.DetectedObject{
		{
			ID:         "det_1",
			Class:      "person",
			Confidence: 0.9,
			BoundingBox: domain.Rectangle{
				X: 10, Y: 20, Width: 50, Height: 100,
			},
			Timestamp: time.Now(),
		},
	}

	// Create track
	_, err := tracker.Update(ctx, detections, "frame_1")
	require.NoError(t, err)

	// Get active tracks
	activeTracks := tracker.GetActiveTracks()
	assert.Len(t, activeTracks, 1)
	assert.Equal(t, TrackStateTentative, activeTracks[0].State)
}

func TestTracker_GetTrack(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	ctx := context.Background()
	detection := domain.DetectedObject{
		ID:         "det_1",
		Class:      "person",
		Confidence: 0.9,
		BoundingBox: domain.Rectangle{
			X: 10, Y: 20, Width: 50, Height: 100,
		},
		Timestamp: time.Now(),
	}

	// Create track
	tracks, err := tracker.Update(ctx, []domain.DetectedObject{detection}, "frame_1")
	require.NoError(t, err)
	require.Len(t, tracks, 1)
	trackID := tracks[0].ID

	// Get specific track
	track, exists := tracker.GetTrack(trackID)
	assert.True(t, exists)
	assert.NotNil(t, track)
	assert.Equal(t, trackID, track.ID)

	// Get non-existent track
	_, exists = tracker.GetTrack("non-existent")
	assert.False(t, exists)
}

func TestTracker_GetTrackHistory(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	ctx := context.Background()
	detection := domain.DetectedObject{
		ID:         "det_1",
		Class:      "person",
		Confidence: 0.9,
		BoundingBox: domain.Rectangle{
			X: 10, Y: 20, Width: 50, Height: 100,
		},
		Timestamp: time.Now(),
	}

	// Create track
	tracks, err := tracker.Update(ctx, []domain.DetectedObject{detection}, "frame_1")
	require.NoError(t, err)
	require.Len(t, tracks, 1)
	trackID := tracks[0].ID

	// Get track history
	trajectory, exists := tracker.GetTrackHistory(trackID)
	assert.True(t, exists)
	assert.Len(t, trajectory, 1)

	// Get history for non-existent track
	_, exists = tracker.GetTrackHistory("non-existent")
	assert.False(t, exists)
}

func TestTracker_GetStats(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	// Initial stats
	stats := tracker.GetStats()
	assert.Equal(t, 0, stats.ActiveTracks)
	assert.Equal(t, int64(0), stats.TotalTracks)

	ctx := context.Background()
	detection := domain.DetectedObject{
		ID:         "det_1",
		Class:      "person",
		Confidence: 0.9,
		BoundingBox: domain.Rectangle{
			X: 10, Y: 20, Width: 50, Height: 100,
		},
		Timestamp: time.Now(),
	}

	// Create track
	_, err := tracker.Update(ctx, []domain.DetectedObject{detection}, "frame_1")
	require.NoError(t, err)

	// Updated stats
	stats = tracker.GetStats()
	assert.Equal(t, 1, stats.ActiveTracks)
	assert.Equal(t, int64(1), stats.TotalTracks)
}

func TestTracker_Close(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	ctx := context.Background()
	detection := domain.DetectedObject{
		ID:         "det_1",
		Class:      "person",
		Confidence: 0.9,
		BoundingBox: domain.Rectangle{
			X: 10, Y: 20, Width: 50, Height: 100,
		},
		Timestamp: time.Now(),
	}

	// Create track
	_, err := tracker.Update(ctx, []domain.DetectedObject{detection}, "frame_1")
	require.NoError(t, err)
	assert.Len(t, tracker.tracks, 1)

	// Close tracker
	err = tracker.Close()
	assert.NoError(t, err)
	assert.Empty(t, tracker.tracks)
}

func TestTrackState_String(t *testing.T) {
	assert.Equal(t, "tentative", TrackStateTentative.String())
	assert.Equal(t, "confirmed", TrackStateConfirmed.String())
	assert.Equal(t, "deleted", TrackStateDeleted.String())
}

func TestRectangle_GetCenter(t *testing.T) {
	rect := domain.Rectangle{
		X: 10, Y: 20, Width: 50, Height: 100,
	}

	centerX, centerY := rect.GetCenter()
	assert.Equal(t, 35.0, centerX) // 10 + 50/2
	assert.Equal(t, 70.0, centerY) // 20 + 100/2
}

func TestTracker_VelocityCalculation(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	ctx := context.Background()

	// First detection
	detection1 := domain.DetectedObject{
		ID:         "det_1",
		Class:      "person",
		Confidence: 0.9,
		BoundingBox: domain.Rectangle{
			X: 10, Y: 20, Width: 50, Height: 100,
		},
		Timestamp: time.Now(),
	}

	tracks, err := tracker.Update(ctx, []domain.DetectedObject{detection1}, "frame_1")
	require.NoError(t, err)
	require.Len(t, tracks, 1)

	// Second detection - moved position
	detection2 := domain.DetectedObject{
		ID:         "det_2",
		Class:      "person",
		Confidence: 0.85,
		BoundingBox: domain.Rectangle{
			X: 20, Y: 30, Width: 50, Height: 100, // Moved 10 pixels in each direction
		},
		Timestamp: detection1.Timestamp.Add(1 * time.Second), // 1 second later
	}

	tracks, err = tracker.Update(ctx, []domain.DetectedObject{detection2}, "frame_2")
	require.NoError(t, err)
	require.Len(t, tracks, 1)

	// Check velocity calculation
	track := tracks[0]
	assert.Equal(t, 10.0, track.Velocity.VX) // 10 pixels/second in X
	assert.Equal(t, 10.0, track.Velocity.VY) // 10 pixels/second in Y
}

func TestTracker_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultTrackerConfig()
	tracker := NewTracker(logger, config)

	ctx := context.Background()
	detection := domain.DetectedObject{
		ID:         "det_1",
		Class:      "person",
		Confidence: 0.9,
		BoundingBox: domain.Rectangle{
			X: 10, Y: 20, Width: 50, Height: 100,
		},
		Timestamp: time.Now(),
	}

	// Create track
	_, err := tracker.Update(ctx, []domain.DetectedObject{detection}, "frame_1")
	require.NoError(t, err)

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Concurrent reads
			tracker.GetActiveTracks()
			tracker.GetStats()

			// Concurrent track lookup
			tracks := tracker.GetActiveTracks()
			if len(tracks) > 0 {
				tracker.GetTrack(tracks[0].ID)
				tracker.GetTrackHistory(tracks[0].ID)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	err = tracker.Close()
	assert.NoError(t, err)
}
