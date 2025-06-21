//go:build opencv
// +build opencv

package detection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewDetector(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()

	detector := NewDetector(logger, config)

	assert.NotNil(t, detector)
	assert.Equal(t, config.ConfidenceThreshold, detector.confidenceThreshold)
	assert.Equal(t, config.NMSThreshold, detector.nmsThreshold)
	assert.Equal(t, config.Classes, detector.classes)
	assert.False(t, detector.modelLoaded)
}

func TestDefaultDetectorConfig(t *testing.T) {
	config := DefaultDetectorConfig()

	assert.Equal(t, 0.5, config.ConfidenceThreshold)
	assert.Equal(t, 0.4, config.NMSThreshold)
	assert.Equal(t, 640, config.InputSize)
	assert.False(t, config.EnableGPU)
	assert.NotEmpty(t, config.Classes)
}

func TestDetector_LoadModel_FileNotExists(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	ctx := context.Background()
	err := detector.LoadModel(ctx, "/nonexistent/model.onnx")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model file does not exist")
	assert.False(t, detector.IsModelLoaded())
}

func TestDetector_LoadModel_InvalidFormat(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	ctx := context.Background()
	err := detector.LoadModel(ctx, "model.txt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported model format")
	assert.False(t, detector.IsModelLoaded())
}

func TestDetector_DetectObjects_NoModel(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	ctx := context.Background()
	frameData := []byte("fake image data")

	objects, err := detector.DetectObjects(ctx, frameData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model not loaded")
	assert.Nil(t, objects)
}

func TestDetector_SetConfidenceThreshold(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	// Test valid threshold
	detector.SetConfidenceThreshold(0.7)
	assert.Equal(t, 0.7, detector.confidenceThreshold)

	// Test invalid thresholds (should not change)
	detector.SetConfidenceThreshold(-0.1)
	assert.Equal(t, 0.7, detector.confidenceThreshold)

	detector.SetConfidenceThreshold(1.1)
	assert.Equal(t, 0.7, detector.confidenceThreshold)
}

func TestDetector_GetSupportedClasses(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	classes := detector.GetSupportedClasses()

	assert.NotEmpty(t, classes)
	assert.Equal(t, config.Classes, classes)

	// Verify it returns a copy (modifying returned slice shouldn't affect original)
	classes[0] = "modified"
	originalClasses := detector.GetSupportedClasses()
	assert.NotEqual(t, "modified", originalClasses[0])
}

func TestDetector_GetModelInfo(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	info := detector.GetModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "", info["model_path"])
	assert.Equal(t, false, info["model_loaded"])
	assert.Equal(t, config.ConfidenceThreshold, info["confidence_threshold"])
	assert.Equal(t, config.NMSThreshold, info["nms_threshold"])
	assert.Equal(t, config.Classes, info["classes"])
}

func TestDetector_Close(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	err := detector.Close()
	assert.NoError(t, err)
	assert.False(t, detector.IsModelLoaded())
}

func TestGetCOCOClasses(t *testing.T) {
	classes := GetCOCOClasses()

	assert.NotEmpty(t, classes)
	assert.Equal(t, 80, len(classes)) // COCO has 80 classes
	assert.Equal(t, "person", classes[0])
	assert.Contains(t, classes, "car")
	assert.Contains(t, classes, "dog")
	assert.Contains(t, classes, "cat")
}

func TestNewYOLODetector(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultYOLOv5Config()

	detector := NewYOLODetector(logger, config)

	assert.NotNil(t, detector)
	assert.NotNil(t, detector.Detector)
	assert.Equal(t, 640, detector.inputSize)
	assert.Equal(t, YOLOv5Format, detector.outputFormat)
	assert.NotEmpty(t, detector.anchors)
	assert.NotEmpty(t, detector.strides)
}

func TestDefaultYOLOv5Config(t *testing.T) {
	config := DefaultYOLOv5Config()

	assert.Equal(t, 640, config.InputSize)
	assert.Equal(t, YOLOv5Format, config.OutputFormat)
	assert.Len(t, config.Strides, 3)
	assert.Len(t, config.Anchors, 3)
	assert.Equal(t, []int{8, 16, 32}, config.Strides)
}

func TestDefaultYOLOv8Config(t *testing.T) {
	config := DefaultYOLOv8Config()

	assert.Equal(t, 640, config.InputSize)
	assert.Equal(t, YOLOv8Format, config.OutputFormat)
	assert.Len(t, config.Strides, 3)
	assert.Equal(t, []int{8, 16, 32}, config.Strides)
}

func TestDetection_CalculateIOU(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	tests := []struct {
		name     string
		det1     Detection
		det2     Detection
		expected float64
	}{
		{
			name: "no overlap",
			det1: Detection{X: 10, Y: 10, Width: 20, Height: 20},
			det2: Detection{X: 50, Y: 50, Width: 20, Height: 20},
			expected: 0.0,
		},
		{
			name: "complete overlap",
			det1: Detection{X: 10, Y: 10, Width: 20, Height: 20},
			det2: Detection{X: 10, Y: 10, Width: 20, Height: 20},
			expected: 1.0,
		},
		{
			name: "partial overlap",
			det1: Detection{X: 10, Y: 10, Width: 20, Height: 20},
			det2: Detection{X: 15, Y: 15, Width: 20, Height: 20},
			expected: 0.39, // Corrected expected value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iou := detector.calculateIOU(tt.det1, tt.det2)
			assert.InDelta(t, tt.expected, iou, 0.01)
		})
	}
}

func TestDetection_FilterByConfidence(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	config.ConfidenceThreshold = 0.5
	detector := NewDetector(logger, config)

	detections := []Detection{
		{Confidence: 0.3, ClassID: 0},
		{Confidence: 0.6, ClassID: 1},
		{Confidence: 0.8, ClassID: 2},
		{Confidence: 0.4, ClassID: 3},
	}

	filtered := detector.filterByConfidence(detections)

	assert.Len(t, filtered, 2)
	assert.Equal(t, float32(0.6), filtered[0].Confidence)
	assert.Equal(t, float32(0.8), filtered[1].Confidence)
}

func TestDetection_ApplyNMS(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	config.NMSThreshold = 0.5
	detector := NewDetector(logger, config)

	// Create overlapping detections of the same class
	detections := []Detection{
		{X: 10, Y: 10, Width: 20, Height: 20, Confidence: 0.9, ClassID: 0},
		{X: 12, Y: 12, Width: 20, Height: 20, Confidence: 0.8, ClassID: 0}, // Overlapping
		{X: 50, Y: 50, Width: 20, Height: 20, Confidence: 0.7, ClassID: 0}, // Non-overlapping
	}

	result := detector.applyNMS(detections)

	// Should keep the highest confidence detection and the non-overlapping one
	assert.Len(t, result, 2)
	assert.Equal(t, float32(0.9), result[0].Confidence)
	assert.Equal(t, float32(0.7), result[1].Confidence)
}

// Integration test (requires ONNX Runtime to be installed)
func TestDetector_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	// This test would require an actual ONNX model file
	// For now, we just test that the detector can be created and closed
	assert.NotNil(t, detector)
	assert.False(t, detector.IsModelLoaded())

	err := detector.Close()
	assert.NoError(t, err)
}

func TestDetector_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultDetectorConfig()
	detector := NewDetector(logger, config)

	// Test concurrent access to detector methods
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Test concurrent reads
			detector.IsModelLoaded()
			detector.GetSupportedClasses()
			detector.GetModelInfo()

			// Test concurrent writes
			detector.SetConfidenceThreshold(0.6)
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

	err := detector.Close()
	assert.NoError(t, err)
}
