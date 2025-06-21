package detection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/video"
	"go.uber.org/zap"
)

// InferenceEngine manages the inference pipeline
type InferenceEngine struct {
	logger    *zap.Logger
	detector  domain.ObjectDetector
	config    InferenceConfig
	isRunning bool
	mutex     sync.RWMutex
	stats     *InferenceStats
}

// InferenceConfig configures the inference engine
type InferenceConfig struct {
	BatchSize           int
	MaxConcurrentInfers int
	InferenceTimeout    time.Duration
	EnableBatching      bool
	QueueSize           int
}

// InferenceStats tracks inference performance
type InferenceStats struct {
	TotalInferences    int64
	SuccessfulInfers   int64
	FailedInfers       int64
	AverageInferTime   time.Duration
	TotalInferTime     time.Duration
	ObjectsDetected    int64
	StartTime          time.Time
	LastInferTime      time.Time
	InferencesPerSec   float64
	mutex              sync.RWMutex
}

// InferenceRequest represents a single inference request
type InferenceRequest struct {
	ID        string
	Frame     *video.Frame
	StreamID  string
	Timestamp time.Time
	Context   context.Context
	Response  chan *InferenceResponse
}

// InferenceResponse represents the result of an inference
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

// DefaultInferenceConfig returns default inference configuration
func DefaultInferenceConfig() InferenceConfig {
	return InferenceConfig{
		BatchSize:           1,
		MaxConcurrentInfers: 4,
		InferenceTimeout:    5 * time.Second,
		EnableBatching:      false,
		QueueSize:           100,
	}
}

// NewInferenceEngine creates a new inference engine
func NewInferenceEngine(logger *zap.Logger, detector domain.ObjectDetector, config InferenceConfig) *InferenceEngine {
	return &InferenceEngine{
		logger:   logger.With(zap.String("component", "inference_engine")),
		detector: detector,
		config:   config,
		stats: &InferenceStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the inference engine
func (ie *InferenceEngine) Start(ctx context.Context) error {
	ie.mutex.Lock()
	defer ie.mutex.Unlock()

	if ie.isRunning {
		return fmt.Errorf("inference engine is already running")
	}

	ie.logger.Info("Starting inference engine",
		zap.Int("max_concurrent", ie.config.MaxConcurrentInfers),
		zap.Int("batch_size", ie.config.BatchSize),
		zap.Bool("batching_enabled", ie.config.EnableBatching))

	ie.isRunning = true
	ie.stats.StartTime = time.Now()

	ie.logger.Info("Inference engine started")
	return nil
}

// Stop stops the inference engine
func (ie *InferenceEngine) Stop() error {
	ie.mutex.Lock()
	defer ie.mutex.Unlock()

	if !ie.isRunning {
		return fmt.Errorf("inference engine is not running")
	}

	ie.logger.Info("Stopping inference engine")
	ie.isRunning = false

	ie.logger.Info("Inference engine stopped")
	return nil
}

// ProcessFrame processes a single frame for object detection
func (ie *InferenceEngine) ProcessFrame(ctx context.Context, frame *video.Frame) (*InferenceResponse, error) {
	if !ie.IsRunning() {
		return nil, fmt.Errorf("inference engine is not running")
	}

	if frame == nil {
		return nil, fmt.Errorf("frame is nil")
	}

	startTime := time.Now()

	// Create inference request
	request := &InferenceRequest{
		ID:        fmt.Sprintf("infer_%s_%d", frame.StreamID, frame.FrameNum),
		Frame:     frame,
		StreamID:  frame.StreamID,
		Timestamp: time.Now(),
		Context:   ctx,
		Response:  make(chan *InferenceResponse, 1),
	}

	// Process the request
	response, err := ie.processRequest(request)
	if err != nil {
		ie.updateStats(startTime, false, 0)
		return nil, fmt.Errorf("failed to process inference request: %w", err)
	}

	ie.updateStats(startTime, true, len(response.Objects))
	return response, nil
}

// ProcessFrameAsync processes a frame asynchronously
func (ie *InferenceEngine) ProcessFrameAsync(ctx context.Context, frame *video.Frame) <-chan *InferenceResponse {
	responseChan := make(chan *InferenceResponse, 1)

	go func() {
		defer close(responseChan)

		response, err := ie.ProcessFrame(ctx, frame)
		if err != nil {
			response = &InferenceResponse{
				ID:        fmt.Sprintf("infer_%s_%d", frame.StreamID, frame.FrameNum),
				StreamID:  frame.StreamID,
				Error:     err,
				Timestamp: time.Now(),
			}
		}

		select {
		case responseChan <- response:
		case <-ctx.Done():
		}
	}()

	return responseChan
}

// processRequest processes a single inference request
func (ie *InferenceEngine) processRequest(request *InferenceRequest) (*InferenceResponse, error) {
	startTime := time.Now()

	// Check context cancellation
	select {
	case <-request.Context.Done():
		return nil, request.Context.Err()
	default:
	}

	// Get frame data
	frameData := request.Frame.Data
	if frameData == nil {
		return nil, fmt.Errorf("frame data is nil")
	}

	// Run object detection
	objects, err := ie.detector.DetectObjects(request.Context, frameData)
	if err != nil {
		return &InferenceResponse{
			ID:          request.ID,
			StreamID:    request.StreamID,
			Error:       fmt.Errorf("detection failed: %w", err),
			ProcessTime: time.Since(startTime),
			Timestamp:   time.Now(),
		}, nil
	}

	// Update object metadata
	for i := range objects {
		objects[i].StreamID = request.StreamID
		objects[i].FrameID = request.Frame.ID
		objects[i].Timestamp = request.Timestamp
	}

	response := &InferenceResponse{
		ID:          request.ID,
		StreamID:    request.StreamID,
		Objects:     objects,
		ProcessTime: time.Since(startTime),
		Timestamp:   time.Now(),
		FrameWidth:  request.Frame.Width,
		FrameHeight: request.Frame.Height,
	}

	return response, nil
}

// IsRunning returns whether the inference engine is running
func (ie *InferenceEngine) IsRunning() bool {
	ie.mutex.RLock()
	defer ie.mutex.RUnlock()
	return ie.isRunning
}

// GetStats returns inference statistics
func (ie *InferenceEngine) GetStats() *InferenceStats {
	ie.stats.mutex.RLock()
	defer ie.stats.mutex.RUnlock()

	// Return a copy to avoid race conditions
	return &InferenceStats{
		TotalInferences:  ie.stats.TotalInferences,
		SuccessfulInfers: ie.stats.SuccessfulInfers,
		FailedInfers:     ie.stats.FailedInfers,
		AverageInferTime: ie.stats.AverageInferTime,
		TotalInferTime:   ie.stats.TotalInferTime,
		ObjectsDetected:  ie.stats.ObjectsDetected,
		StartTime:        ie.stats.StartTime,
		LastInferTime:    ie.stats.LastInferTime,
		InferencesPerSec: ie.stats.InferencesPerSec,
	}
}

// updateStats updates inference statistics
func (ie *InferenceEngine) updateStats(startTime time.Time, success bool, objectCount int) {
	ie.stats.mutex.Lock()
	defer ie.stats.mutex.Unlock()

	inferTime := time.Since(startTime)
	ie.stats.TotalInferences++
	ie.stats.TotalInferTime += inferTime
	ie.stats.LastInferTime = time.Now()

	if success {
		ie.stats.SuccessfulInfers++
		ie.stats.ObjectsDetected += int64(objectCount)
	} else {
		ie.stats.FailedInfers++
	}

	// Calculate average inference time
	if ie.stats.TotalInferences > 0 {
		ie.stats.AverageInferTime = ie.stats.TotalInferTime / time.Duration(ie.stats.TotalInferences)
	}

	// Calculate inferences per second
	elapsed := time.Since(ie.stats.StartTime).Seconds()
	if elapsed > 0 {
		ie.stats.InferencesPerSec = float64(ie.stats.TotalInferences) / elapsed
	}
}

// GetDetector returns the underlying detector
func (ie *InferenceEngine) GetDetector() domain.ObjectDetector {
	return ie.detector
}

// SetDetector sets a new detector
func (ie *InferenceEngine) SetDetector(detector domain.ObjectDetector) {
	ie.mutex.Lock()
	defer ie.mutex.Unlock()

	ie.detector = detector
	ie.logger.Info("Detector updated")
}

// GetConfig returns the inference configuration
func (ie *InferenceEngine) GetConfig() InferenceConfig {
	return ie.config
}

// UpdateConfig updates the inference configuration
func (ie *InferenceEngine) UpdateConfig(config InferenceConfig) {
	ie.mutex.Lock()
	defer ie.mutex.Unlock()

	ie.config = config
	ie.logger.Info("Inference configuration updated",
		zap.Int("batch_size", config.BatchSize),
		zap.Int("max_concurrent", config.MaxConcurrentInfers),
		zap.Duration("timeout", config.InferenceTimeout))
}

// DetectionFrameProcessor implements video.FrameProcessor for object detection
type DetectionFrameProcessor struct {
	logger    *zap.Logger
	engine    *InferenceEngine
	streamID  string
	callbacks []DetectionCallback
	mutex     sync.RWMutex
}

// DetectionCallback is called when objects are detected
type DetectionCallback func(response *InferenceResponse)

// NewDetectionFrameProcessor creates a new detection frame processor
func NewDetectionFrameProcessor(logger *zap.Logger, engine *InferenceEngine, streamID string) *DetectionFrameProcessor {
	return &DetectionFrameProcessor{
		logger:   logger.With(zap.String("component", "detection_processor"), zap.String("stream_id", streamID)),
		engine:   engine,
		streamID: streamID,
	}
}

// ProcessFrame processes a frame for object detection
func (dfp *DetectionFrameProcessor) ProcessFrame(ctx context.Context, frame *video.Frame) (*video.Frame, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame is nil")
	}

	// Run inference
	response, err := dfp.engine.ProcessFrame(ctx, frame)
	if err != nil {
		dfp.logger.Error("Inference failed", zap.Error(err))
		return frame, nil // Return original frame on error
	}

	// Call callbacks
	dfp.mutex.RLock()
	callbacks := make([]DetectionCallback, len(dfp.callbacks))
	copy(callbacks, dfp.callbacks)
	dfp.mutex.RUnlock()

	for _, callback := range callbacks {
		go callback(response)
	}

	dfp.logger.Debug("Detection completed",
		zap.String("frame_id", frame.ID),
		zap.Int("objects_detected", len(response.Objects)),
		zap.Duration("process_time", response.ProcessTime))

	return frame, nil
}

// AddCallback adds a detection callback
func (dfp *DetectionFrameProcessor) AddCallback(callback DetectionCallback) {
	dfp.mutex.Lock()
	defer dfp.mutex.Unlock()

	dfp.callbacks = append(dfp.callbacks, callback)
}

// RemoveCallback removes a detection callback
func (dfp *DetectionFrameProcessor) RemoveCallback(callback DetectionCallback) {
	dfp.mutex.Lock()
	defer dfp.mutex.Unlock()

	// Note: This is a simplified implementation
	// In practice, you might want to use a more sophisticated callback management system
	dfp.logger.Info("Callback removal requested (simplified implementation)")
}

// GetCallbackCount returns the number of registered callbacks
func (dfp *DetectionFrameProcessor) GetCallbackCount() int {
	dfp.mutex.RLock()
	defer dfp.mutex.RUnlock()
	return len(dfp.callbacks)
}
