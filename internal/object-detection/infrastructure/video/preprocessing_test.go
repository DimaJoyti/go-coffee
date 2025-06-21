package video

import (
	"context"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

func TestDefaultPreprocessingConfig(t *testing.T) {
	config := DefaultPreprocessingConfig()

	assert.Equal(t, 640, config.TargetWidth)
	assert.Equal(t, 640, config.TargetHeight)
	assert.True(t, config.MaintainAspect)
	assert.False(t, config.Normalize)
	assert.Len(t, config.Mean, 3)
	assert.Len(t, config.Std, 3)
	assert.False(t, config.SwapRB)
	assert.False(t, config.Crop)
	assert.Equal(t, 85, config.Quality)
}

func TestResizeProcessor_ProcessFrame(t *testing.T) {
	logger := zap.NewNop()
	config := PreprocessingConfig{
		TargetWidth:    320,
		TargetHeight:   240,
		MaintainAspect: false,
	}

	processor := NewResizeProcessor(logger, config)

	// Create test frame with mat
	mat := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3) // 640x480
	defer mat.Close()

	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	require.NoError(t, err)
	defer frame.Release()

	ctx := context.Background()
	processedFrame, err := processor.ProcessFrame(ctx, frame)
	require.NoError(t, err)
	require.NotNil(t, processedFrame)
	defer processedFrame.Release()

	// Check dimensions
	assert.Equal(t, 320, processedFrame.Width)
	assert.Equal(t, 240, processedFrame.Height)
}

func TestResizeProcessor_ProcessFrame_MaintainAspect(t *testing.T) {
	logger := zap.NewNop()
	config := PreprocessingConfig{
		TargetWidth:    640,
		TargetHeight:   640,
		MaintainAspect: true,
	}

	processor := NewResizeProcessor(logger, config)

	// Create test frame with mat (landscape)
	mat := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3) // 640x480
	defer mat.Close()

	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	require.NoError(t, err)
	defer frame.Release()

	ctx := context.Background()
	processedFrame, err := processor.ProcessFrame(ctx, frame)
	require.NoError(t, err)
	require.NotNil(t, processedFrame)
	defer processedFrame.Release()

	// Should maintain aspect ratio (landscape)
	assert.Equal(t, 640, processedFrame.Width)
	assert.Equal(t, 480, processedFrame.Height) // Calculated to maintain aspect
}

func TestResizeProcessor_CalculateTargetSize(t *testing.T) {
	logger := zap.NewNop()
	
	tests := []struct {
		name           string
		config         PreprocessingConfig
		inputWidth     int
		inputHeight    int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name: "no aspect ratio maintenance",
			config: PreprocessingConfig{
				TargetWidth:    320,
				TargetHeight:   240,
				MaintainAspect: false,
			},
			inputWidth:     640,
			inputHeight:    480,
			expectedWidth:  320,
			expectedHeight: 240,
		},
		{
			name: "maintain aspect ratio - landscape",
			config: PreprocessingConfig{
				TargetWidth:    640,
				TargetHeight:   640,
				MaintainAspect: true,
			},
			inputWidth:     800,
			inputHeight:    600,
			expectedWidth:  640,
			expectedHeight: 480,
		},
		{
			name: "maintain aspect ratio - portrait",
			config: PreprocessingConfig{
				TargetWidth:    640,
				TargetHeight:   640,
				MaintainAspect: true,
			},
			inputWidth:     600,
			inputHeight:    800,
			expectedWidth:  480,
			expectedHeight: 640,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewResizeProcessor(logger, tt.config)
			result := processor.calculateTargetSize(tt.inputWidth, tt.inputHeight)
			
			assert.Equal(t, tt.expectedWidth, result.X)
			assert.Equal(t, tt.expectedHeight, result.Y)
		})
	}
}

func TestNormalizationProcessor_ProcessFrame_Disabled(t *testing.T) {
	logger := zap.NewNop()
	config := PreprocessingConfig{
		Normalize: false,
	}

	processor := NewNormalizationProcessor(logger, config)

	// Create test frame
	mat := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer mat.Close()

	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	require.NoError(t, err)
	defer frame.Release()

	ctx := context.Background()
	processedFrame, err := processor.ProcessFrame(ctx, frame)
	require.NoError(t, err)

	// Should return the same frame when normalization is disabled
	assert.Equal(t, frame, processedFrame)
}

func TestColorSpaceProcessor_ProcessFrame_NoSwap(t *testing.T) {
	logger := zap.NewNop()
	config := PreprocessingConfig{
		SwapRB: false,
	}

	processor := NewColorSpaceProcessor(logger, config)

	// Create test frame
	mat := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer mat.Close()

	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	require.NoError(t, err)
	defer frame.Release()

	ctx := context.Background()
	processedFrame, err := processor.ProcessFrame(ctx, frame)
	require.NoError(t, err)

	// Should return the same frame when SwapRB is false
	assert.Equal(t, frame, processedFrame)
}

func TestColorSpaceProcessor_ProcessFrame_SwapRB(t *testing.T) {
	logger := zap.NewNop()
	config := PreprocessingConfig{
		SwapRB: true,
	}

	processor := NewColorSpaceProcessor(logger, config)

	// Create test frame
	mat := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer mat.Close()

	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	require.NoError(t, err)
	defer frame.Release()

	ctx := context.Background()
	processedFrame, err := processor.ProcessFrame(ctx, frame)
	require.NoError(t, err)
	require.NotNil(t, processedFrame)
	defer processedFrame.Release()

	// Should return a different frame when SwapRB is true
	assert.NotEqual(t, frame, processedFrame)
	assert.Equal(t, frame.ID, processedFrame.ID)
	assert.Equal(t, frame.StreamID, processedFrame.StreamID)
}

func TestCropProcessor_ProcessFrame_Disabled(t *testing.T) {
	logger := zap.NewNop()
	config := PreprocessingConfig{
		Crop: false,
	}

	processor := NewCropProcessor(logger, config)

	// Create test frame
	mat := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer mat.Close()

	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	require.NoError(t, err)
	defer frame.Release()

	ctx := context.Background()
	processedFrame, err := processor.ProcessFrame(ctx, frame)
	require.NoError(t, err)

	// Should return the same frame when cropping is disabled
	assert.Equal(t, frame, processedFrame)
}

func TestCropProcessor_CalculateCropRegion(t *testing.T) {
	logger := zap.NewNop()
	
	tests := []struct {
		name         string
		config       PreprocessingConfig
		inputWidth   int
		inputHeight  int
		expectedRect image.Rectangle
	}{
		{
			name: "crop wider image",
			config: PreprocessingConfig{
				TargetWidth:  640,
				TargetHeight: 640,
			},
			inputWidth:  800,
			inputHeight: 600,
			expectedRect: image.Rectangle{
				Min: image.Point{X: 100, Y: 0},
				Max: image.Point{X: 700, Y: 600},
			},
		},
		{
			name: "crop taller image",
			config: PreprocessingConfig{
				TargetWidth:  640,
				TargetHeight: 640,
			},
			inputWidth:  600,
			inputHeight: 800,
			expectedRect: image.Rectangle{
				Min: image.Point{X: 0, Y: 100},
				Max: image.Point{X: 600, Y: 700},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewCropProcessor(logger, tt.config)
			result := processor.calculateCropRegion(tt.inputWidth, tt.inputHeight)
			
			assert.Equal(t, tt.expectedRect, result)
		})
	}
}

func TestCompositeProcessor_ProcessFrame(t *testing.T) {
	logger := zap.NewNop()
	config := PreprocessingConfig{
		TargetWidth:  320,
		TargetHeight: 240,
		SwapRB:       true,
		Crop:         false,
		Normalize:    false,
	}

	processor := NewCompositeProcessor(logger, config)

	// Should have resize and color space processors
	assert.Greater(t, processor.GetProcessorCount(), 0)

	// Create test frame
	mat := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer mat.Close()

	frame, err := NewFrameFromMat("test-stream", 1, &mat)
	require.NoError(t, err)
	defer frame.Release()

	ctx := context.Background()
	processedFrame, err := processor.ProcessFrame(ctx, frame)
	require.NoError(t, err)
	require.NotNil(t, processedFrame)
	defer processedFrame.Release()

	// Should be resized
	assert.Equal(t, 320, processedFrame.Width)
	assert.Equal(t, 240, processedFrame.Height)
}

func TestCompositeProcessor_GetProcessorCount(t *testing.T) {
	logger := zap.NewNop()
	
	tests := []struct {
		name           string
		config         PreprocessingConfig
		expectedCount  int
	}{
		{
			name: "no processing",
			config: PreprocessingConfig{},
			expectedCount: 0,
		},
		{
			name: "resize only",
			config: PreprocessingConfig{
				TargetWidth:  640,
				TargetHeight: 480,
			},
			expectedCount: 1,
		},
		{
			name: "resize and color swap",
			config: PreprocessingConfig{
				TargetWidth:  640,
				TargetHeight: 480,
				SwapRB:       true,
			},
			expectedCount: 2,
		},
		{
			name: "all processing",
			config: PreprocessingConfig{
				TargetWidth:  640,
				TargetHeight: 480,
				Crop:         true,
				SwapRB:       true,
				Normalize:    true,
			},
			expectedCount: 4, // crop, resize, color swap, normalize
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewCompositeProcessor(logger, tt.config)
			count := processor.GetProcessorCount()
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestProcessFrame_NilFrame(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultPreprocessingConfig()

	processors := []FrameProcessor{
		NewResizeProcessor(logger, config),
		NewNormalizationProcessor(logger, config),
		NewColorSpaceProcessor(logger, config),
		NewCropProcessor(logger, config),
	}

	ctx := context.Background()

	for _, processor := range processors {
		_, err := processor.ProcessFrame(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "frame is nil")
	}
}
