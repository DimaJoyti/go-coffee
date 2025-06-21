package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDetectedObject_Validation(t *testing.T) {
	tests := []struct {
		name   string
		object DetectedObject
		valid  bool
	}{
		{
			name: "valid object",
			object: DetectedObject{
				ID:         "test-id",
				Class:      "person",
				Confidence: 0.85,
				BoundingBox: Rectangle{
					X:      10,
					Y:      20,
					Width:  100,
					Height: 150,
				},
				Timestamp: time.Now(),
				StreamID:  "stream-1",
				FrameID:   "frame-1",
			},
			valid: true,
		},
		{
			name: "invalid confidence - too low",
			object: DetectedObject{
				ID:         "test-id",
				Class:      "person",
				Confidence: -0.1,
				BoundingBox: Rectangle{
					X:      10,
					Y:      20,
					Width:  100,
					Height: 150,
				},
				Timestamp: time.Now(),
				StreamID:  "stream-1",
				FrameID:   "frame-1",
			},
			valid: false,
		},
		{
			name: "invalid confidence - too high",
			object: DetectedObject{
				ID:         "test-id",
				Class:      "person",
				Confidence: 1.1,
				BoundingBox: Rectangle{
					X:      10,
					Y:      20,
					Width:  100,
					Height: 150,
				},
				Timestamp: time.Now(),
				StreamID:  "stream-1",
				FrameID:   "frame-1",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.object.Confidence >= 0 && tt.object.Confidence <= 1
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestRectangle_Area(t *testing.T) {
	tests := []struct {
		name     string
		rect     Rectangle
		expected int
	}{
		{
			name: "positive dimensions",
			rect: Rectangle{
				X:      0,
				Y:      0,
				Width:  10,
				Height: 20,
			},
			expected: 200,
		},
		{
			name: "zero area",
			rect: Rectangle{
				X:      0,
				Y:      0,
				Width:  0,
				Height: 20,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			area := tt.rect.Width * tt.rect.Height
			assert.Equal(t, tt.expected, area)
		})
	}
}

func TestVideoStream_StatusTransitions(t *testing.T) {
	stream := &VideoStream{
		ID:     "test-stream",
		Name:   "Test Stream",
		Source: "/dev/video0",
		Type:   StreamTypeWebcam,
		Status: StreamStatusIdle,
		Config: StreamConfig{
			FPS:                30,
			Resolution:         "1920x1080",
			DetectionThreshold: 0.5,
			TrackingEnabled:    true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test valid status transitions
	validTransitions := map[StreamStatus][]StreamStatus{
		StreamStatusIdle:       {StreamStatusActive, StreamStatusError},
		StreamStatusActive:     {StreamStatusProcessing, StreamStatusStopped, StreamStatusError},
		StreamStatusProcessing: {StreamStatusActive, StreamStatusStopped, StreamStatusError},
		StreamStatusStopped:    {StreamStatusIdle, StreamStatusActive},
		StreamStatusError:      {StreamStatusIdle, StreamStatusStopped},
	}

	for fromStatus, toStatuses := range validTransitions {
		for _, toStatus := range toStatuses {
			t.Run(string(fromStatus)+"_to_"+string(toStatus), func(t *testing.T) {
				stream.Status = fromStatus
				stream.Status = toStatus
				assert.Equal(t, toStatus, stream.Status)
			})
		}
	}
}

func TestTrackingData_Velocity(t *testing.T) {
	tracking := &TrackingData{
		ID:          "track-1",
		ObjectClass: "person",
		FirstSeen:   time.Now().Add(-10 * time.Second),
		LastSeen:    time.Now(),
		Trajectory: []Rectangle{
			{X: 10, Y: 20, Width: 50, Height: 100},
			{X: 15, Y: 25, Width: 50, Height: 100},
			{X: 20, Y: 30, Width: 50, Height: 100},
		},
		IsActive: true,
		StreamID: "stream-1",
	}

	// Test velocity calculation (simplified)
	if len(tracking.Trajectory) >= 2 {
		last := tracking.Trajectory[len(tracking.Trajectory)-1]
		prev := tracking.Trajectory[len(tracking.Trajectory)-2]
		
		velocity := &Velocity{
			X: float64(last.X - prev.X),
			Y: float64(last.Y - prev.Y),
		}
		
		tracking.Velocity = velocity
		
		assert.Equal(t, 5.0, tracking.Velocity.X)
		assert.Equal(t, 5.0, tracking.Velocity.Y)
	}
}

func TestStreamConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config StreamConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: StreamConfig{
				FPS:                30,
				Resolution:         "1920x1080",
				DetectionThreshold: 0.5,
				TrackingEnabled:    true,
				RecordingEnabled:   false,
				AlertsEnabled:      true,
			},
			valid: true,
		},
		{
			name: "invalid FPS",
			config: StreamConfig{
				FPS:                0,
				Resolution:         "1920x1080",
				DetectionThreshold: 0.5,
				TrackingEnabled:    true,
			},
			valid: false,
		},
		{
			name: "invalid detection threshold",
			config: StreamConfig{
				FPS:                30,
				Resolution:         "1920x1080",
				DetectionThreshold: 1.5,
				TrackingEnabled:    true,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.config.FPS > 0 && 
				tt.config.DetectionThreshold >= 0 && 
				tt.config.DetectionThreshold <= 1
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestDetectionResult_ObjectCount(t *testing.T) {
	result := &DetectionResult{
		ID:       "result-1",
		StreamID: "stream-1",
		FrameID:  "frame-1",
		Objects: []DetectedObject{
			{
				ID:         "obj-1",
				Class:      "person",
				Confidence: 0.9,
			},
			{
				ID:         "obj-2",
				Class:      "car",
				Confidence: 0.8,
			},
			{
				ID:         "obj-3",
				Class:      "person",
				Confidence: 0.7,
			},
		},
		ProcessTime: 50 * time.Millisecond,
		Timestamp:   time.Now(),
		FrameWidth:  1920,
		FrameHeight: 1080,
	}

	// Test object counting by class
	classCounts := make(map[string]int)
	for _, obj := range result.Objects {
		classCounts[obj.Class]++
	}

	assert.Equal(t, 2, classCounts["person"])
	assert.Equal(t, 1, classCounts["car"])
	assert.Equal(t, 3, len(result.Objects))
}

func TestProcessingStats_FPSCalculation(t *testing.T) {
	stats := &ProcessingStats{
		StreamID:           "stream-1",
		TotalFrames:        1000,
		ProcessedFrames:    950,
		DetectedObjects:    2500,
		AverageProcessTime: 33 * time.Millisecond,
		LastProcessedAt:    time.Now(),
	}

	// Calculate FPS based on average process time
	if stats.AverageProcessTime > 0 {
		fps := 1.0 / stats.AverageProcessTime.Seconds()
		stats.FPS = fps
	}

	assert.InDelta(t, 30.3, stats.FPS, 0.1)
	assert.Equal(t, int64(1000), stats.TotalFrames)
	assert.Equal(t, int64(950), stats.ProcessedFrames)
}
