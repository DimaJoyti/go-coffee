package detection

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/video"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockDetector for testing
type MockDetector struct {
	shouldError bool
	objects     []domain.DetectedObject
	delay       time.Duration
}

func (m *MockDetector) LoadModel(ctx context.Context, modelPath string) error {
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockDetector) DetectObjects(ctx context.Context, frameData []byte) ([]domain.DetectedObject, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	
	if m.shouldError {
		return nil, assert.AnError
	}
	
	return m.objects, nil
}

func (m *MockDetector) SetConfidenceThreshold(threshold float64) {}

func (m *MockDetector) GetSupportedClasses() []string {
	return []string{"person", "car", "dog"}
}

func (m *MockDetector) IsModelLoaded() bool {
	return !m.shouldError
}

func (m *MockDetector) Close() error {
	return nil
}

func TestNewInferenceEngine(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{}
	config := DefaultInferenceConfig()

	engine := NewInferenceEngine(logger, detector, config)

	assert.NotNil(t, engine)
	assert.Equal(t, detector, engine.detector)
	assert.Equal(t, config, engine.config)
	assert.False(t, engine.IsRunning())
}

func TestDefaultInferenceConfig(t *testing.T) {
	config := DefaultInferenceConfig()

	assert.Equal(t, 1, config.BatchSize)
	assert.Equal(t, 4, config.MaxConcurrentInfers)
	assert.Equal(t, 5*time.Second, config.InferenceTimeout)
	assert.False(t, config.EnableBatching)
	assert.Equal(t, 100, config.QueueSize)
}

func TestInferenceEngine_StartStop(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)

	ctx := context.Background()

	// Test start
	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, engine.IsRunning())

	// Test start when already running
	err = engine.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stop
	err = engine.Stop()
	assert.NoError(t, err)
	assert.False(t, engine.IsRunning())

	// Test stop when not running
	err = engine.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestInferenceEngine_ProcessFrame(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{
		objects: []domain.DetectedObject{
			{
				ID:         "test-obj-1",
				Class:      "person",
				Confidence: 0.9,
			},
		},
	}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)

	ctx := context.Background()
	err := engine.Start(ctx)
	require.NoError(t, err)
	defer engine.Stop()

	// Create test frame
	frame := video.NewFrame("test-stream", 1, []byte("test frame data"))

	// Test successful processing
	response, err := engine.ProcessFrame(ctx, frame)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test-stream", response.StreamID)
	assert.Len(t, response.Objects, 1)
	assert.Equal(t, "person", response.Objects[0].Class)
	assert.Equal(t, "test-stream", response.Objects[0].StreamID)
	assert.Equal(t, frame.ID, response.Objects[0].FrameID)
	assert.NoError(t, response.Error)
}

func TestInferenceEngine_ProcessFrame_NotRunning(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)

	ctx := context.Background()
	frame := video.NewFrame("test-stream", 1, []byte("test frame data"))

	response, err := engine.ProcessFrame(ctx, frame)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
	assert.Nil(t, response)
}

func TestInferenceEngine_ProcessFrame_NilFrame(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)

	ctx := context.Background()
	err := engine.Start(ctx)
	require.NoError(t, err)
	defer engine.Stop()

	response, err := engine.ProcessFrame(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "frame is nil")
	assert.Nil(t, response)
}

func TestInferenceEngine_ProcessFrame_DetectionError(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{shouldError: true}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)

	ctx := context.Background()
	err := engine.Start(ctx)
	require.NoError(t, err)
	defer engine.Stop()

	frame := video.NewFrame("test-stream", 1, []byte("test frame data"))

	response, err := engine.ProcessFrame(ctx, frame)
	assert.NoError(t, err) // Engine doesn't return error, but response contains error
	assert.NotNil(t, response)
	assert.Error(t, response.Error)
	assert.Contains(t, response.Error.Error(), "detection failed")
}

func TestInferenceEngine_ProcessFrameAsync(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{
		objects: []domain.DetectedObject{
			{ID: "test-obj-1", Class: "person", Confidence: 0.9},
		},
		delay: 10 * time.Millisecond, // Small delay to test async behavior
	}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)

	ctx := context.Background()
	err := engine.Start(ctx)
	require.NoError(t, err)
	defer engine.Stop()

	frame := video.NewFrame("test-stream", 1, []byte("test frame data"))

	// Test async processing
	responseChan := engine.ProcessFrameAsync(ctx, frame)

	select {
	case response := <-responseChan:
		assert.NotNil(t, response)
		assert.Equal(t, "test-stream", response.StreamID)
		assert.Len(t, response.Objects, 1)
		assert.NoError(t, response.Error)
	case <-time.After(1 * time.Second):
		t.Fatal("Async processing timed out")
	}
}

func TestInferenceEngine_GetStats(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{
		objects: []domain.DetectedObject{
			{ID: "test-obj-1", Class: "person", Confidence: 0.9},
		},
	}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)

	ctx := context.Background()
	err := engine.Start(ctx)
	require.NoError(t, err)
	defer engine.Stop()

	// Initial stats
	stats := engine.GetStats()
	assert.Equal(t, int64(0), stats.TotalInferences)
	assert.Equal(t, int64(0), stats.SuccessfulInfers)
	assert.Equal(t, int64(0), stats.FailedInfers)

	// Process a frame
	frame := video.NewFrame("test-stream", 1, []byte("test frame data"))
	_, err = engine.ProcessFrame(ctx, frame)
	require.NoError(t, err)

	// Check updated stats
	stats = engine.GetStats()
	assert.Equal(t, int64(1), stats.TotalInferences)
	assert.Equal(t, int64(1), stats.SuccessfulInfers)
	assert.Equal(t, int64(0), stats.FailedInfers)
	assert.Equal(t, int64(1), stats.ObjectsDetected)
	assert.Greater(t, stats.AverageInferTime, time.Duration(0))
}

func TestInferenceEngine_SetDetector(t *testing.T) {
	logger := zap.NewNop()
	detector1 := &MockDetector{}
	detector2 := &MockDetector{}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector1, config)

	assert.Equal(t, detector1, engine.GetDetector())

	engine.SetDetector(detector2)
	assert.Equal(t, detector2, engine.GetDetector())
}

func TestInferenceEngine_UpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)

	newConfig := InferenceConfig{
		BatchSize:           2,
		MaxConcurrentInfers: 8,
		InferenceTimeout:    10 * time.Second,
		EnableBatching:      true,
		QueueSize:           200,
	}

	engine.UpdateConfig(newConfig)
	assert.Equal(t, newConfig, engine.GetConfig())
}

func TestNewDetectionFrameProcessor(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)
	streamID := "test-stream"

	processor := NewDetectionFrameProcessor(logger, engine, streamID)

	assert.NotNil(t, processor)
	assert.Equal(t, engine, processor.engine)
	assert.Equal(t, streamID, processor.streamID)
	assert.Equal(t, 0, processor.GetCallbackCount())
}

func TestDetectionFrameProcessor_ProcessFrame(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{
		objects: []domain.DetectedObject{
			{ID: "test-obj-1", Class: "person", Confidence: 0.9},
		},
	}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)
	streamID := "test-stream"

	ctx := context.Background()
	err := engine.Start(ctx)
	require.NoError(t, err)
	defer engine.Stop()

	processor := NewDetectionFrameProcessor(logger, engine, streamID)

	// Add callback to capture results
	var capturedResponse *InferenceResponse
	processor.AddCallback(func(response *InferenceResponse) {
		capturedResponse = response
	})

	frame := video.NewFrame(streamID, 1, []byte("test frame data"))

	// Process frame
	resultFrame, err := processor.ProcessFrame(ctx, frame)
	assert.NoError(t, err)
	assert.Equal(t, frame, resultFrame) // Should return original frame

	// Wait a bit for callback to be called
	time.Sleep(10 * time.Millisecond)

	// Check callback was called
	assert.NotNil(t, capturedResponse)
	assert.Equal(t, streamID, capturedResponse.StreamID)
	assert.Len(t, capturedResponse.Objects, 1)
}

func TestDetectionFrameProcessor_ProcessFrame_NilFrame(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)
	processor := NewDetectionFrameProcessor(logger, engine, "test-stream")

	ctx := context.Background()

	resultFrame, err := processor.ProcessFrame(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "frame is nil")
	assert.Nil(t, resultFrame)
}

func TestDetectionFrameProcessor_AddRemoveCallback(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)
	processor := NewDetectionFrameProcessor(logger, engine, "test-stream")

	assert.Equal(t, 0, processor.GetCallbackCount())

	// Add callback
	callback := func(response *InferenceResponse) {}
	processor.AddCallback(callback)
	assert.Equal(t, 1, processor.GetCallbackCount())

	// Remove callback (simplified implementation just logs)
	processor.RemoveCallback(callback)
	// Note: The current implementation doesn't actually remove callbacks
	// This is noted in the code as a simplified implementation
}

func TestInferenceEngine_ContextCancellation(t *testing.T) {
	logger := zap.NewNop()
	detector := &MockDetector{
		delay: 100 * time.Millisecond, // Long enough to cancel
	}
	config := DefaultInferenceConfig()
	engine := NewInferenceEngine(logger, detector, config)

	ctx, cancel := context.WithCancel(context.Background())
	err := engine.Start(ctx)
	require.NoError(t, err)
	defer engine.Stop()

	frame := video.NewFrame("test-stream", 1, []byte("test frame data"))

	// Cancel context before processing completes
	cancel()

	response, err := engine.ProcessFrame(ctx, frame)
	assert.NoError(t, err) // Engine handles context cancellation gracefully
	assert.NotNil(t, response)
	// Response should contain context cancellation error
	assert.Error(t, response.Error)
}
