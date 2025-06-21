package video

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewProcessor(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(logger)

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.streams)
	assert.Equal(t, logger, processor.logger)
}

func TestProcessor_OpenStream_InvalidType(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(logger)
	ctx := context.Background()

	err := processor.OpenStream(ctx, "test", "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported stream type")
}

func TestProcessor_CloseStream_NotFound(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(logger)
	ctx := context.Background()

	err := processor.CloseStream(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream not found")
}

func TestProcessor_ReadFrame_NotFound(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(logger)
	ctx := context.Background()

	_, err := processor.ReadFrame(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream not found")
}

func TestProcessor_GetFrameRate_NotFound(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(logger)
	ctx := context.Background()

	_, err := processor.GetFrameRate(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream not found")
}

func TestProcessor_GetResolution_NotFound(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(logger)
	ctx := context.Background()

	_, _, err := processor.GetResolution(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream not found")
}

func TestProcessor_IsStreamActive_NotFound(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(logger)
	ctx := context.Background()

	active := processor.IsStreamActive(ctx, "nonexistent")
	assert.False(t, active)
}

func TestProcessor_GetActiveStreams_Empty(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(logger)

	streams := processor.GetActiveStreams()
	assert.Empty(t, streams)
}

func TestProcessor_Close(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(logger)

	err := processor.Close()
	assert.NoError(t, err)
}

func TestGenerateStreamID(t *testing.T) {
	source1 := "test_source"
	source2 := "different_source"
	streamType := domain.StreamTypeFile

	id1 := generateStreamID(source1, streamType)
	id2 := generateStreamID(source2, streamType)
	id3 := generateStreamID(source1, streamType)

	// IDs should be different due to timestamp
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id1, id3)

	// IDs should contain stream type
	assert.Contains(t, id1, string(streamType))
	assert.Contains(t, id2, string(streamType))
}

func TestHashString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test", "364492"},
		{"hello", "5994471"},
		{"", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := hashString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Integration tests (require actual video devices/files)
func TestProcessor_Integration_Webcam(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	processor := NewProcessor(logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to open webcam (may not be available in CI)
	err := processor.OpenStream(ctx, "0", domain.StreamTypeWebcam)
	if err != nil {
		t.Skipf("Webcam not available: %v", err)
		return
	}

	// Get stream ID
	streams := processor.GetActiveStreams()
	require.Len(t, streams, 1)
	streamID := streams[0]

	// Test stream properties
	fps, err := processor.GetFrameRate(ctx, streamID)
	assert.NoError(t, err)
	assert.Greater(t, fps, 0.0)

	width, height, err := processor.GetResolution(ctx, streamID)
	assert.NoError(t, err)
	assert.Greater(t, width, 0)
	assert.Greater(t, height, 0)

	// Test frame reading
	frameData, err := processor.ReadFrame(ctx, streamID)
	if err == nil {
		assert.NotEmpty(t, frameData)
	}

	// Test stream status
	active := processor.IsStreamActive(ctx, streamID)
	assert.True(t, active)

	// Close stream
	err = processor.CloseStream(ctx, streamID)
	assert.NoError(t, err)

	// Verify stream is closed
	active = processor.IsStreamActive(ctx, streamID)
	assert.False(t, active)
}

func TestStreamHandler_GetStats(t *testing.T) {
	logger := zap.NewNop()
	handler := &StreamHandler{
		ID:        "test-stream",
		Source:    "/dev/video0",
		Type:      domain.StreamTypeWebcam,
		Status:    domain.StreamStatusActive,
		IsActive:  true,
		FrameRate: 30.0,
		Width:     640,
		Height:    480,
		Logger:    logger,
	}

	stats := handler.GetStats()

	assert.Equal(t, "test-stream", stats["id"])
	assert.Equal(t, "/dev/video0", stats["source"])
	assert.Equal(t, "webcam", stats["type"])
	assert.Equal(t, "active", stats["status"])
	assert.Equal(t, true, stats["is_active"])
	assert.Equal(t, 30.0, stats["frame_rate"])
	assert.Equal(t, 640, stats["width"])
	assert.Equal(t, 480, stats["height"])
}

func TestStreamHandler_SetFrameRate(t *testing.T) {
	logger := zap.NewNop()
	handler := &StreamHandler{
		Logger: logger,
	}

	handler.SetFrameRate(25.0)
	assert.Equal(t, 25.0, handler.FrameRate)

	// Test invalid frame rate
	handler.SetFrameRate(0)
	assert.Equal(t, 25.0, handler.FrameRate) // Should remain unchanged
}

func TestStreamHandler_GetFrameInterval(t *testing.T) {
	logger := zap.NewNop()
	handler := &StreamHandler{
		FrameRate: 30.0,
		Logger:    logger,
	}

	interval := handler.GetFrameInterval()
	expectedMs := int64(math.Round(1000.0 / 30.0))
	expected := time.Duration(expectedMs) * time.Millisecond
	assert.Equal(t, expected, interval)

	// Test with zero frame rate
	handler.FrameRate = 0
	interval = handler.GetFrameInterval()
	assert.Equal(t, 33*time.Millisecond, interval) // Default ~30 FPS
}

func TestStreamHandler_Stop(t *testing.T) {
	logger := zap.NewNop()
	handler := &StreamHandler{
		IsActive:    true,
		Status:      domain.StreamStatusActive,
		StopChannel: make(chan bool, 1),
		Logger:      logger,
	}

	handler.stop()

	assert.False(t, handler.IsActive)
	assert.Equal(t, domain.StreamStatusStopped, handler.Status)

	// Verify stop signal was sent
	select {
	case <-handler.StopChannel:
		// Expected
	default:
		t.Error("Stop signal not sent")
	}
}
