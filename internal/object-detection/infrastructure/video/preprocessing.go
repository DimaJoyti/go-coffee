//go:build opencv
// +build opencv

package video

import (
	"context"
	"fmt"
	"image"

	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

// PreprocessingConfig defines preprocessing parameters
type PreprocessingConfig struct {
	TargetWidth    int
	TargetHeight   int
	MaintainAspect bool
	Normalize      bool
	Mean           []float32 // RGB mean values for normalization
	Std            []float32 // RGB standard deviation values for normalization
	SwapRB         bool      // Swap Red and Blue channels (BGR to RGB)
	Crop           bool      // Crop to target size instead of resize
	Quality        int       // JPEG quality for encoding (1-100)
}

// DefaultPreprocessingConfig returns default preprocessing configuration
func DefaultPreprocessingConfig() PreprocessingConfig {
	return PreprocessingConfig{
		TargetWidth:    640,
		TargetHeight:   640,
		MaintainAspect: true,
		Normalize:      false,
		Mean:           []float32{0.485, 0.456, 0.406}, // ImageNet mean
		Std:            []float32{0.229, 0.224, 0.225}, // ImageNet std
		SwapRB:         false,
		Crop:           false,
		Quality:        85,
	}
}

// ResizeProcessor resizes frames to target dimensions
type ResizeProcessor struct {
	logger *zap.Logger
	config PreprocessingConfig
}

// NewResizeProcessor creates a new resize processor
func NewResizeProcessor(logger *zap.Logger, config PreprocessingConfig) *ResizeProcessor {
	return &ResizeProcessor{
		logger: logger.With(zap.String("processor", "resize")),
		config: config,
	}
}

// ProcessFrame resizes the input frame
func (r *ResizeProcessor) ProcessFrame(ctx context.Context, frame *Frame) (*Frame, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame is nil")
	}

	// Get Mat from frame
	mat, err := frame.ToMat()
	if err != nil {
		return nil, fmt.Errorf("failed to get mat from frame: %w", err)
	}

	// Calculate target size
	targetSize := r.calculateTargetSize(mat.Cols(), mat.Rows())

	// Resize the image
	resized := gocv.NewMat()
	defer resized.Close()

	gocv.Resize(*mat, &resized, targetSize, 0, 0, gocv.InterpolationLinear)

	// Create new frame with resized image
	newFrame, err := NewFrameFromMat(frame.StreamID, frame.FrameNum, &resized)
	if err != nil {
		return nil, fmt.Errorf("failed to create frame from resized mat: %w", err)
	}

	// Copy metadata
	newFrame.ID = frame.ID
	newFrame.Timestamp = frame.Timestamp

	return newFrame, nil
}

// calculateTargetSize calculates the target size based on configuration
func (r *ResizeProcessor) calculateTargetSize(width, height int) image.Point {
	targetWidth := r.config.TargetWidth
	targetHeight := r.config.TargetHeight

	if !r.config.MaintainAspect {
		return image.Point{X: targetWidth, Y: targetHeight}
	}

	// Maintain aspect ratio
	aspectRatio := float64(width) / float64(height)

	if aspectRatio > 1 {
		// Landscape
		targetHeight = int(float64(targetWidth) / aspectRatio)
	} else {
		// Portrait
		targetWidth = int(float64(targetHeight) * aspectRatio)
	}

	return image.Point{X: targetWidth, Y: targetHeight}
}

// NormalizationProcessor normalizes pixel values
type NormalizationProcessor struct {
	logger *zap.Logger
	config PreprocessingConfig
}

// NewNormalizationProcessor creates a new normalization processor
func NewNormalizationProcessor(logger *zap.Logger, config PreprocessingConfig) *NormalizationProcessor {
	return &NormalizationProcessor{
		logger: logger.With(zap.String("processor", "normalization")),
		config: config,
	}
}

// ProcessFrame normalizes the input frame
func (n *NormalizationProcessor) ProcessFrame(ctx context.Context, frame *Frame) (*Frame, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame is nil")
	}

	if !n.config.Normalize {
		return frame, nil
	}

	// Get Mat from frame
	mat, err := frame.ToMat()
	if err != nil {
		return nil, fmt.Errorf("failed to get mat from frame: %w", err)
	}

	// Convert to float32
	floatMat := gocv.NewMat()
	defer floatMat.Close()

	mat.ConvertTo(&floatMat, gocv.MatTypeCV32F)

	// Normalize to [0, 1]
	floatMat.DivideFloat(255.0)

	// Apply mean and std normalization if provided
	if len(n.config.Mean) == 3 && len(n.config.Std) == 3 {
		n.applyMeanStdNormalization(&floatMat)
	}

	// Convert back to uint8 for encoding
	normalizedMat := gocv.NewMat()
	defer normalizedMat.Close()

	floatMat.MultiplyFloat(255.0)
	floatMat.ConvertTo(&normalizedMat, gocv.MatTypeCV8U)

	// Create new frame
	newFrame, err := NewFrameFromMat(frame.StreamID, frame.FrameNum, &normalizedMat)
	if err != nil {
		return nil, fmt.Errorf("failed to create frame from normalized mat: %w", err)
	}

	// Copy metadata
	newFrame.ID = frame.ID
	newFrame.Timestamp = frame.Timestamp

	return newFrame, nil
}

// applyMeanStdNormalization applies mean and standard deviation normalization
func (n *NormalizationProcessor) applyMeanStdNormalization(mat *gocv.Mat) {
	// Split channels
	channels := gocv.Split(*mat)
	defer func() {
		for _, ch := range channels {
			ch.Close()
		}
	}()

	// Apply normalization to each channel
	for i, ch := range channels {
		if i < len(n.config.Mean) && i < len(n.config.Std) {
			// (pixel - mean) / std
			ch.SubtractFloat(float32(n.config.Mean[i]))
			ch.DivideFloat(float32(n.config.Std[i]))
		}
	}

	// Merge channels back
	gocv.Merge(channels, mat)
}

// ColorSpaceProcessor converts between color spaces
type ColorSpaceProcessor struct {
	logger *zap.Logger
	config PreprocessingConfig
}

// NewColorSpaceProcessor creates a new color space processor
func NewColorSpaceProcessor(logger *zap.Logger, config PreprocessingConfig) *ColorSpaceProcessor {
	return &ColorSpaceProcessor{
		logger: logger.With(zap.String("processor", "colorspace")),
		config: config,
	}
}

// ProcessFrame converts color space of the input frame
func (c *ColorSpaceProcessor) ProcessFrame(ctx context.Context, frame *Frame) (*Frame, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame is nil")
	}

	// Get Mat from frame
	mat, err := frame.ToMat()
	if err != nil {
		return nil, fmt.Errorf("failed to get mat from frame: %w", err)
	}

	// Apply color space conversion if needed
	if c.config.SwapRB {
		convertedMat := gocv.NewMat()
		defer convertedMat.Close()

		gocv.CvtColor(*mat, &convertedMat, gocv.ColorBGRToRGB)

		// Create new frame
		newFrame, err := NewFrameFromMat(frame.StreamID, frame.FrameNum, &convertedMat)
		if err != nil {
			return nil, fmt.Errorf("failed to create frame from converted mat: %w", err)
		}

		// Copy metadata
		newFrame.ID = frame.ID
		newFrame.Timestamp = frame.Timestamp

		return newFrame, nil
	}

	return frame, nil
}

// CropProcessor crops frames to target dimensions
type CropProcessor struct {
	logger *zap.Logger
	config PreprocessingConfig
}

// NewCropProcessor creates a new crop processor
func NewCropProcessor(logger *zap.Logger, config PreprocessingConfig) *CropProcessor {
	return &CropProcessor{
		logger: logger.With(zap.String("processor", "crop")),
		config: config,
	}
}

// ProcessFrame crops the input frame
func (c *CropProcessor) ProcessFrame(ctx context.Context, frame *Frame) (*Frame, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame is nil")
	}

	if !c.config.Crop {
		return frame, nil
	}

	// Get Mat from frame
	mat, err := frame.ToMat()
	if err != nil {
		return nil, fmt.Errorf("failed to get mat from frame: %w", err)
	}

	// Calculate crop region
	cropRect := c.calculateCropRegion(mat.Cols(), mat.Rows())

	// Crop the image
	cropped := mat.Region(cropRect)
	defer cropped.Close()

	// Create new frame
	newFrame, err := NewFrameFromMat(frame.StreamID, frame.FrameNum, &cropped)
	if err != nil {
		return nil, fmt.Errorf("failed to create frame from cropped mat: %w", err)
	}

	// Copy metadata
	newFrame.ID = frame.ID
	newFrame.Timestamp = frame.Timestamp

	return newFrame, nil
}

// calculateCropRegion calculates the crop region to maintain aspect ratio
func (c *CropProcessor) calculateCropRegion(width, height int) image.Rectangle {
	targetAspect := float64(c.config.TargetWidth) / float64(c.config.TargetHeight)
	currentAspect := float64(width) / float64(height)

	var cropWidth, cropHeight int
	var x, y int

	if currentAspect > targetAspect {
		// Image is wider, crop width
		cropHeight = height
		cropWidth = int(float64(height) * targetAspect)
		x = (width - cropWidth) / 2
		y = 0
	} else {
		// Image is taller, crop height
		cropWidth = width
		cropHeight = int(float64(width) / targetAspect)
		x = 0
		y = (height - cropHeight) / 2
	}

	return image.Rectangle{
		Min: image.Point{X: x, Y: y},
		Max: image.Point{X: x + cropWidth, Y: y + cropHeight},
	}
}

// CompositeProcessor combines multiple preprocessing steps
type CompositeProcessor struct {
	logger     *zap.Logger
	processors []FrameProcessor
}

// NewCompositeProcessor creates a new composite processor
func NewCompositeProcessor(logger *zap.Logger, config PreprocessingConfig) *CompositeProcessor {
	processors := make([]FrameProcessor, 0)

	// Add processors based on configuration
	if config.Crop {
		processors = append(processors, NewCropProcessor(logger, config))
	}

	if config.TargetWidth > 0 && config.TargetHeight > 0 {
		processors = append(processors, NewResizeProcessor(logger, config))
	}

	if config.SwapRB {
		processors = append(processors, NewColorSpaceProcessor(logger, config))
	}

	if config.Normalize {
		processors = append(processors, NewNormalizationProcessor(logger, config))
	}

	return &CompositeProcessor{
		logger:     logger.With(zap.String("processor", "composite")),
		processors: processors,
	}
}

// ProcessFrame processes the frame through all configured processors
func (c *CompositeProcessor) ProcessFrame(ctx context.Context, frame *Frame) (*Frame, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame is nil")
	}

	currentFrame := frame

	for i, processor := range c.processors {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		processedFrame, err := processor.ProcessFrame(ctx, currentFrame)
		if err != nil {
			return nil, fmt.Errorf("processor %d failed: %w", i, err)
		}

		// Release previous frame if it's not the original
		if currentFrame != frame {
			currentFrame.Release()
		}

		currentFrame = processedFrame
	}

	return currentFrame, nil
}

// GetProcessorCount returns the number of processors in the composite
func (c *CompositeProcessor) GetProcessorCount() int {
	return len(c.processors)
}
