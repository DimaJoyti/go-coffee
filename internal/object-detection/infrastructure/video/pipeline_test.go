package video

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

func TestNewPipeline(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID:   "test-stream",
		Workers:    2,
		BufferSize: 5,
	}

	pipeline := NewPipeline(logger, config)

	assert.NotNil(t, pipeline)
	assert.Equal(t, "test-stream", pipeline.streamID)
	assert.Equal(t, 2, pipeline.workers)
	assert.NotNil(t, pipeline.input)
	assert.NotNil(t, pipeline.output)
	assert.NotNil(t, pipeline.stats)
	assert.False(t, pipeline.isRunning)
}

func TestPipeline_DefaultConfig(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID: "test-stream",
		// Workers and BufferSize not set
	}

	pipeline := NewPipeline(logger, config)

	assert.Equal(t, 1, pipeline.workers)     // Default workers
	assert.Equal(t, 10, cap(pipeline.input)) // Default buffer size
}

func TestPipeline_AddProcessor(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID: "test-stream",
	}

	pipeline := NewPipeline(logger, config)
	
	// Create a mock processor
	processor := &MockProcessor{}
	
	pipeline.AddProcessor(processor)
	
	assert.Len(t, pipeline.processors, 1)
}

func TestPipeline_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID: "test-stream",
		Workers:  1,
	}

	pipeline := NewPipeline(logger, config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test start
	err := pipeline.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, pipeline.IsRunning())

	// Test start when already running
	err = pipeline.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stop
	err = pipeline.Stop()
	assert.NoError(t, err)
	assert.False(t, pipeline.IsRunning())

	// Test stop when not running
	err = pipeline.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestPipeline_ProcessFrame(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID:   "test-stream",
		Workers:    1,
		BufferSize: 2,
	}

	pipeline := NewPipeline(logger, config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test process frame when not running
	frame := NewFrame("test-stream", 1, []byte("test"))
	err := pipeline.ProcessFrame(frame)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// Start pipeline
	err = pipeline.Start(ctx)
	require.NoError(t, err)

	// Test process frame when running
	err = pipeline.ProcessFrame(frame)
	assert.NoError(t, err)

	// Fill buffer to test dropping
	for i := 0; i < 5; i++ {
		frame := NewFrame("test-stream", i+2, []byte("test"))
		pipeline.ProcessFrame(frame) // May drop frames
	}

	// Stop pipeline
	pipeline.Stop()
}

func TestPipeline_GetStats(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID: "test-stream",
	}

	pipeline := NewPipeline(logger, config)
	
	stats := pipeline.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.FramesProcessed)
	assert.Equal(t, int64(0), stats.FramesDropped)
	assert.Equal(t, time.Duration(0), stats.AverageProcessTime)
}

func TestPipeline_GetBufferUsage(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID:   "test-stream",
		BufferSize: 5,
	}

	pipeline := NewPipeline(logger, config)
	
	input, output := pipeline.GetBufferUsage()
	assert.Equal(t, 0, input)
	assert.Equal(t, 0, output)
}

func TestNewFrame(t *testing.T) {
	data := []byte("test frame data")
	frame := NewFrame("test-stream", 42, data)

	assert.Equal(t, "test-stream_frame_42", frame.ID)
	assert.Equal(t, "test-stream", frame.StreamID)
	assert.Equal(t, data, frame.Data)
	assert.Equal(t, 42, frame.FrameNum)
	assert.NotZero(t, frame.Timestamp)
}

func TestNewFrameFromMat(t *testing.T) {
	// Create a test mat
	mat := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer mat.Close()

	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	require.NoError(t, err)
	require.NotNil(t, frame)

	assert.Equal(t, "test-stream_frame_1", frame.ID)
	assert.Equal(t, "test-stream", frame.StreamID)
	assert.Equal(t, 1, frame.FrameNum)
	assert.Equal(t, 100, frame.Width)
	assert.Equal(t, 100, frame.Height)
	assert.NotEmpty(t, frame.Data)
	assert.NotNil(t, frame.Mat)
}

func TestNewFrameFromMat_EmptyMat(t *testing.T) {
	mat := gocv.NewMat()
	defer mat.Close()

	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	assert.Error(t, err)
	assert.Nil(t, frame)
	assert.Contains(t, err.Error(), "invalid or empty mat")
}

func TestFrame_Clone(t *testing.T) {
	original := NewFrame("test-stream", 1, []byte("test data"))
	original.Width = 640
	original.Height = 480

	clone := original.Clone()

	assert.Equal(t, original.ID, clone.ID)
	assert.Equal(t, original.StreamID, clone.StreamID)
	assert.Equal(t, original.FrameNum, clone.FrameNum)
	assert.Equal(t, original.Width, clone.Width)
	assert.Equal(t, original.Height, clone.Height)
	assert.Equal(t, original.Timestamp, clone.Timestamp)

	// Data should be copied, not shared
	assert.Equal(t, original.Data, clone.Data)
	if len(original.Data) > 0 {
		assert.NotSame(t, &original.Data[0], &clone.Data[0])
	}
}

func TestFrame_Release(t *testing.T) {
	// Create frame with mat
	mat := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	require.NoError(t, err)

	// Release frame
	frame.Release()

	assert.Nil(t, frame.Mat)
	assert.Nil(t, frame.Data)
}

func TestFrame_GetSize(t *testing.T) {
	data := []byte("test frame data")
	frame := NewFrame("test-stream", 1, data)

	size := frame.GetSize()
	assert.Equal(t, len(data), size)

	// Test with nil data
	frame.Data = nil
	size = frame.GetSize()
	assert.Equal(t, 0, size)
}

func TestPipelineStats_UpdateStats(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID: "test-stream",
	}

	pipeline := NewPipeline(logger, config)
	
	// Simulate processing
	pipeline.updateStats(10 * time.Millisecond)
	pipeline.updateStats(20 * time.Millisecond)

	stats := pipeline.GetStats()
	assert.Equal(t, int64(2), stats.FramesProcessed)
	assert.Equal(t, 15*time.Millisecond, stats.AverageProcessTime)
	assert.Greater(t, stats.FPS, 0.0)
}

// MockProcessor for testing
type MockProcessor struct {
	processCount int
	shouldError  bool
}

func (m *MockProcessor) ProcessFrame(ctx context.Context, frame *Frame) (*Frame, error) {
	m.processCount++
	
	if m.shouldError {
		return nil, assert.AnError
	}
	
	// Return the same frame (no processing)
	return frame, nil
}

func TestPipeline_WithMockProcessor(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID:   "test-stream",
		Workers:    1,
		BufferSize: 5,
	}

	pipeline := NewPipeline(logger, config)
	processor := &MockProcessor{}
	pipeline.AddProcessor(processor)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := pipeline.Start(ctx)
	require.NoError(t, err)

	// Process a frame
	frame := NewFrame("test-stream", 1, []byte("test"))
	err = pipeline.ProcessFrame(frame)
	assert.NoError(t, err)

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Check if processor was called
	assert.Greater(t, processor.processCount, 0)

	pipeline.Stop()
}

func TestPipeline_ProcessorError(t *testing.T) {
	logger := zap.NewNop()
	config := PipelineConfig{
		StreamID:   "test-stream",
		Workers:    1,
		BufferSize: 5,
	}

	pipeline := NewPipeline(logger, config)
	processor := &MockProcessor{shouldError: true}
	pipeline.AddProcessor(processor)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := pipeline.Start(ctx)
	require.NoError(t, err)

	// Process a frame
	frame := NewFrame("test-stream", 1, []byte("test"))
	err = pipeline.ProcessFrame(frame)
	assert.NoError(t, err)

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Processor should have been called despite error
	assert.Greater(t, processor.processCount, 0)

	pipeline.Stop()
}
