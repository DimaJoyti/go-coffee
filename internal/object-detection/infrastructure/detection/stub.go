//go:build !opencv
// +build !opencv

package detection

import (
	"context"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
)

// Stub types for non-opencv builds

// ResultProcessor stub
type ResultProcessor struct{}

// ProcessorConfig stub
type ProcessorConfig struct {
	ConfidenceThreshold float64
	NMSThreshold        float64
	MaxDetections       int
	ClassFilter         []string
	MinObjectSize       int
	MaxObjectSize       int
}

// InferenceEngine stub
type InferenceEngine struct{}

// InferenceConfig stub
type InferenceConfig struct {
	BatchSize           int
	MaxConcurrentInfers int
	InferenceTimeout    time.Duration
	EnableBatching      bool
	QueueSize           int
}

// InferenceResponse stub
type InferenceResponse struct {
	ID           string
	StreamID     string
	Objects      []domain.DetectedObject
	ProcessTime  time.Duration
	Error        error
	Timestamp    time.Time
	FrameWidth   int
	FrameHeight  int
}

// DetectionCallback is called when objects are detected
type DetectionCallback func(response *InferenceResponse)

// ModelManager stub
type ModelManager struct{}

// ModelManagerConfig stub
type ModelManagerConfig struct {
	StoragePath     string
	MaxModelSize    int64
	AllowedFormats  []string
	ValidateOnLoad  bool
}

// DetectionFrameProcessor stub
type DetectionFrameProcessor struct{}

// Stub functions

func DefaultProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		ConfidenceThreshold: 0.5,
		NMSThreshold:        0.4,
		MaxDetections:       100,
		ClassFilter:         []string{},
		MinObjectSize:       100,
		MaxObjectSize:       0,
	}
}

func DefaultInferenceConfig() InferenceConfig {
	return InferenceConfig{
		BatchSize:           1,
		MaxConcurrentInfers: 4,
		InferenceTimeout:    5 * time.Second,
		EnableBatching:      false,
		QueueSize:           100,
	}
}

func DefaultModelManagerConfig() ModelManagerConfig {
	return ModelManagerConfig{
		StoragePath:    "./data/models",
		MaxModelSize:   500 * 1024 * 1024,
		AllowedFormats: []string{".onnx"},
		ValidateOnLoad: true,
	}
}

func DefaultDetectorConfig() DetectorConfig {
	return DetectorConfig{
		ConfidenceThreshold: 0.5,
		NMSThreshold:        0.4,
		InputSize:           640,
		EnableGPU:           false,
		Classes:             GetCOCOClasses(),
	}
}

func NewResultProcessor(logger *zap.Logger, config ProcessorConfig) *ResultProcessor {
	return &ResultProcessor{}
}

func NewInferenceEngine(logger *zap.Logger, detector domain.ObjectDetector, config InferenceConfig) *InferenceEngine {
	return &InferenceEngine{}
}

func NewModelManager(logger *zap.Logger, config ModelManagerConfig) (*ModelManager, error) {
	return &ModelManager{}, nil
}

func NewDetector(logger *zap.Logger, config DetectorConfig) *Detector {
	return &Detector{}
}

func NewDetectionFrameProcessor(logger *zap.Logger, engine *InferenceEngine, streamID string) *DetectionFrameProcessor {
	return &DetectionFrameProcessor{}
}

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

// Detector stub
type Detector struct{}

// DetectorConfig stub
type DetectorConfig struct {
	ModelPath           string
	ConfidenceThreshold float64
	NMSThreshold        float64
	InputSize           int
	Classes             []string
	EnableGPU           bool
}

// Stub methods for ResultProcessor
func (rp *ResultProcessor) ProcessDetections(ctx context.Context, detections []domain.DetectedObject) ([]domain.DetectedObject, error) {
	return detections, nil
}

func (rp *ResultProcessor) UpdateConfig(config ProcessorConfig) {}
func (rp *ResultProcessor) GetConfig() ProcessorConfig        { return ProcessorConfig{} }
func (rp *ResultProcessor) SetConfidenceThreshold(threshold float64) {}
func (rp *ResultProcessor) SetNMSThreshold(threshold float64) {}
func (rp *ResultProcessor) SetClassFilter(classes []string) {}
func (rp *ResultProcessor) GetStats() map[string]interface{} { return make(map[string]interface{}) }

// Stub methods for Detector
func (d *Detector) LoadModel(ctx context.Context, modelPath string) error                { return nil }
func (d *Detector) DetectObjects(ctx context.Context, frameData []byte) ([]domain.DetectedObject, error) { return []domain.DetectedObject{}, nil }
func (d *Detector) SetConfidenceThreshold(threshold float64) {}
func (d *Detector) GetSupportedClasses() []string { return []string{} }
func (d *Detector) IsModelLoaded() bool { return false }
func (d *Detector) GetModelPath() string { return "" }
func (d *Detector) GetModelInfo() map[string]interface{} { return make(map[string]interface{}) }
func (d *Detector) Close() error { return nil }