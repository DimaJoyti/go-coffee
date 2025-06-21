//go:build opencv
// +build opencv

package detection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/yalue/onnxruntime_go"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

// Detector implements the ObjectDetector interface using ONNX Runtime
type Detector struct {
	logger            *zap.Logger
	session           *onnxruntime_go.Session[float32]
	inputShape        []int64
	outputShapes      [][]int64
	inputName         string
	outputNames       []string
	classes           []string
	confidenceThreshold float64
	nmsThreshold      float64
	modelPath         string
	modelLoaded       bool
	mutex             sync.RWMutex
}

// DetectorConfig configures the object detector
type DetectorConfig struct {
	ModelPath           string
	ConfidenceThreshold float64
	NMSThreshold        float64
	InputSize           int
	Classes             []string
	EnableGPU           bool
}

// DefaultDetectorConfig returns default detector configuration
func DefaultDetectorConfig() DetectorConfig {
	return DetectorConfig{
		ConfidenceThreshold: 0.5,
		NMSThreshold:        0.4,
		InputSize:           640,
		EnableGPU:           false,
		Classes:             GetCOCOClasses(),
	}
}

// NewDetector creates a new object detector
func NewDetector(logger *zap.Logger, config DetectorConfig) *Detector {
	return &Detector{
		logger:              logger.With(zap.String("component", "detector")),
		confidenceThreshold: config.ConfidenceThreshold,
		nmsThreshold:        config.NMSThreshold,
		classes:             config.Classes,
		modelLoaded:         false,
	}
}

// LoadModel loads a detection model
func (d *Detector) LoadModel(ctx context.Context, modelPath string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.logger.Info("Loading detection model", zap.String("path", modelPath))

	// Validate model file extension first
	ext := filepath.Ext(modelPath)
	if ext != ".onnx" {
		return fmt.Errorf("unsupported model format: %s (expected .onnx)", ext)
	}

	// Check if model file exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file does not exist: %s", modelPath)
	}

	// Initialize ONNX Runtime if not already done
	if err := d.initializeONNXRuntime(); err != nil {
		return fmt.Errorf("failed to initialize ONNX Runtime: %w", err)
	}

	// Create session options
	options, err := onnxruntime_go.NewSessionOptions()
	if err != nil {
		return fmt.Errorf("failed to create session options: %w", err)
	}
	defer options.Destroy()

	// Note: Execution provider setup will use default CPU provider

	// Create session
	session, err := onnxruntime_go.NewSession(modelPath, []string{}, []string{}, []*onnxruntime_go.Tensor[float32]{}, []*onnxruntime_go.Tensor[float32]{})
	if err != nil {
		return fmt.Errorf("failed to create ONNX session: %w", err)
	}

	// Close previous session if exists
	if d.session != nil {
		d.session.Destroy()
	}

	d.session = session
	d.modelPath = modelPath

	// Get model input/output information
	if err := d.getModelInfo(); err != nil {
		d.session.Destroy()
		d.session = nil
		return fmt.Errorf("failed to get model info: %w", err)
	}

	d.modelLoaded = true

	d.logger.Info("Model loaded successfully",
		zap.String("path", modelPath),
		zap.String("input_name", d.inputName),
		zap.Strings("output_names", d.outputNames),
		zap.Int64s("input_shape", d.inputShape))

	return nil
}

// DetectObjects detects objects in an image frame
func (d *Detector) DetectObjects(ctx context.Context, frameData []byte) ([]domain.DetectedObject, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.modelLoaded || d.session == nil {
		return nil, fmt.Errorf("model not loaded")
	}

	startTime := time.Now()

	// Decode image
	mat, err := gocv.IMDecode(frameData, gocv.IMReadColor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	defer mat.Close()

	// Preprocess image
	inputTensor, originalSize, err := d.preprocessImage(&mat)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess image: %w", err)
	}
	defer inputTensor.Destroy()

	// Run inference
	outputs, err := d.runInference(inputTensor)
	if err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}
	defer func() {
		for _, output := range outputs {
			output.Destroy()
		}
	}()

	// Post-process results
	detections, err := d.postprocessResults(outputs, originalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to post-process results: %w", err)
	}

	processingTime := time.Since(startTime)
	d.logger.Debug("Detection completed",
		zap.Int("detections", len(detections)),
		zap.Duration("processing_time", processingTime))

	return detections, nil
}

// SetConfidenceThreshold sets the confidence threshold for detection
func (d *Detector) SetConfidenceThreshold(threshold float64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if threshold >= 0 && threshold <= 1 {
		d.confidenceThreshold = threshold
		d.logger.Info("Confidence threshold updated", zap.Float64("threshold", threshold))
	}
}

// GetSupportedClasses retrieves the classes supported by the current model
func (d *Detector) GetSupportedClasses() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Return a copy to avoid race conditions
	classes := make([]string, len(d.classes))
	copy(classes, d.classes)
	return classes
}

// IsModelLoaded checks if a model is currently loaded
func (d *Detector) IsModelLoaded() bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.modelLoaded
}

// GetModelPath returns the path of the currently loaded model
func (d *Detector) GetModelPath() string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.modelPath
}

// GetModelInfo returns information about the loaded model
func (d *Detector) GetModelInfo() map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return map[string]interface{}{
		"model_path":           d.modelPath,
		"model_loaded":         d.modelLoaded,
		"input_name":           d.inputName,
		"input_shape":          d.inputShape,
		"output_names":         d.outputNames,
		"output_shapes":        d.outputShapes,
		"classes":              d.classes,
		"confidence_threshold": d.confidenceThreshold,
		"nms_threshold":        d.nmsThreshold,
	}
}

// Close closes the detector and releases resources
func (d *Detector) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.logger.Info("Closing detector")

	if d.session != nil {
		d.session.Destroy()
		d.session = nil
	}

	d.modelLoaded = false

	d.logger.Info("Detector closed")
	return nil
}

// initializeONNXRuntime initializes the ONNX Runtime
func (d *Detector) initializeONNXRuntime() error {
	// Initialize ONNX Runtime (this is typically done once globally)
	if err := onnxruntime_go.InitializeEnvironment(); err != nil {
		return fmt.Errorf("failed to initialize ONNX Runtime environment: %w", err)
	}
	return nil
}

// getModelInfo extracts input/output information from the loaded model
func (d *Detector) getModelInfo() error {
	if d.session == nil {
		return fmt.Errorf("session not initialized")
	}

	// For now, use default YOLO model configuration
	// TODO: Implement proper model introspection when API is stable
	d.inputName = "images"
	d.inputShape = []int64{1, 3, 640, 640} // Typical YOLO input shape
	d.outputNames = []string{"output0"}
	d.outputShapes = [][]int64{{1, 25200, 85}} // Typical YOLOv5 output shape

	d.logger.Info("Using default model configuration",
		zap.String("input_name", d.inputName),
		zap.Int64s("input_shape", d.inputShape),
		zap.Strings("output_names", d.outputNames))

	return nil
}

// GetCOCOClasses returns the COCO dataset class names
func GetCOCOClasses() []string {
	return []string{
		"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck", "boat",
		"traffic light", "fire hydrant", "stop sign", "parking meter", "bench", "bird", "cat",
		"dog", "horse", "sheep", "cow", "elephant", "bear", "zebra", "giraffe", "backpack",
		"umbrella", "handbag", "tie", "suitcase", "frisbee", "skis", "snowboard", "sports ball",
		"kite", "baseball bat", "baseball glove", "skateboard", "surfboard", "tennis racket",
		"bottle", "wine glass", "cup", "fork", "knife", "spoon", "bowl", "banana", "apple",
		"sandwich", "orange", "broccoli", "carrot", "hot dog", "pizza", "donut", "cake", "chair",
		"couch", "potted plant", "bed", "dining table", "toilet", "tv", "laptop", "mouse",
		"remote", "keyboard", "cell phone", "microwave", "oven", "toaster", "sink", "refrigerator",
		"book", "clock", "vase", "scissors", "teddy bear", "hair drier", "toothbrush",
	}
}
