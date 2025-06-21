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

// DetectionService integrates object detection with video processing
type DetectionService struct {
	logger         *zap.Logger
	detector       domain.ObjectDetector
	inferenceEngine *InferenceEngine
	resultProcessor *ResultProcessor
	modelManager   *ModelManager
	streamManagers map[string]*StreamDetectionManager
	config         DetectionServiceConfig
	mutex          sync.RWMutex
}

// DetectionServiceConfig configures the detection service
type DetectionServiceConfig struct {
	InferenceConfig InferenceConfig
	ProcessorConfig ProcessorConfig
	ModelConfig     ModelManagerConfig
	EnableAsync     bool
	ResultCallback  func(*InferenceResponse)
}

// StreamDetectionManager manages detection for a single stream
type StreamDetectionManager struct {
	streamID        string
	logger          *zap.Logger
	detectionEngine *InferenceEngine
	processor       *DetectionFrameProcessor
	isActive        bool
	stats           *StreamDetectionStats
	mutex           sync.RWMutex
}

// StreamDetectionStats tracks detection statistics per stream
type StreamDetectionStats struct {
	StreamID           string
	TotalFrames        int64
	ProcessedFrames    int64
	DetectedObjects    int64
	AverageProcessTime time.Duration
	LastDetectionTime  time.Time
	FPS                float64
	StartTime          time.Time
	mutex              sync.RWMutex
}

// DefaultDetectionServiceConfig returns default detection service configuration
func DefaultDetectionServiceConfig() DetectionServiceConfig {
	return DetectionServiceConfig{
		InferenceConfig: DefaultInferenceConfig(),
		ProcessorConfig: DefaultProcessorConfig(),
		ModelConfig:     DefaultModelManagerConfig(),
		EnableAsync:     true,
	}
}

// NewDetectionService creates a new detection service
func NewDetectionService(logger *zap.Logger, config DetectionServiceConfig) (*DetectionService, error) {
	// Create detector
	detectorConfig := DefaultDetectorConfig()
	detector := NewDetector(logger, detectorConfig)

	// Create inference engine
	inferenceEngine := NewInferenceEngine(logger, detector, config.InferenceConfig)

	// Create result processor
	resultProcessor := NewResultProcessor(logger, config.ProcessorConfig)

	// Create model manager
	modelManager, err := NewModelManager(logger, config.ModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create model manager: %w", err)
	}

	// Set detector in model manager
	modelManager.SetDetector(detector)

	return &DetectionService{
		logger:          logger.With(zap.String("component", "detection_service")),
		detector:        detector,
		inferenceEngine: inferenceEngine,
		resultProcessor: resultProcessor,
		modelManager:    modelManager,
		streamManagers:  make(map[string]*StreamDetectionManager),
		config:          config,
	}, nil
}

// Start starts the detection service
func (ds *DetectionService) Start(ctx context.Context) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.logger.Info("Starting detection service")

	// Start inference engine
	if err := ds.inferenceEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start inference engine: %w", err)
	}

	ds.logger.Info("Detection service started")
	return nil
}

// Stop stops the detection service
func (ds *DetectionService) Stop() error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.logger.Info("Stopping detection service")

	// Stop all stream managers
	for streamID, manager := range ds.streamManagers {
		ds.logger.Debug("Stopping stream detection manager", zap.String("stream_id", streamID))
		manager.Stop()
	}

	// Stop inference engine
	if err := ds.inferenceEngine.Stop(); err != nil {
		ds.logger.Error("Failed to stop inference engine", zap.Error(err))
	}

	// Close detector
	if err := ds.detector.Close(); err != nil {
		ds.logger.Error("Failed to close detector", zap.Error(err))
	}

	ds.logger.Info("Detection service stopped")
	return nil
}

// StartStreamDetection starts detection for a video stream
func (ds *DetectionService) StartStreamDetection(ctx context.Context, streamID string, pipeline *video.Pipeline) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if _, exists := ds.streamManagers[streamID]; exists {
		return fmt.Errorf("detection already started for stream: %s", streamID)
	}

	ds.logger.Info("Starting stream detection", zap.String("stream_id", streamID))

	// Create stream detection manager
	manager := &StreamDetectionManager{
		streamID:        streamID,
		logger:          ds.logger.With(zap.String("stream_id", streamID)),
		detectionEngine: ds.inferenceEngine,
		isActive:        false,
		stats: &StreamDetectionStats{
			StreamID:  streamID,
			StartTime: time.Now(),
		},
	}

	// Create detection frame processor
	processor := NewDetectionFrameProcessor(ds.logger, ds.inferenceEngine, streamID)

	// Add result callback
	processor.AddCallback(func(response *InferenceResponse) {
		ds.handleDetectionResult(response)
		manager.updateStats(response)
	})

	manager.processor = processor

	// Add processor to pipeline
	pipeline.AddProcessor(processor)

	manager.isActive = true
	ds.streamManagers[streamID] = manager

	ds.logger.Info("Stream detection started", zap.String("stream_id", streamID))
	return nil
}

// StopStreamDetection stops detection for a video stream
func (ds *DetectionService) StopStreamDetection(streamID string) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	manager, exists := ds.streamManagers[streamID]
	if !exists {
		return fmt.Errorf("no detection running for stream: %s", streamID)
	}

	ds.logger.Info("Stopping stream detection", zap.String("stream_id", streamID))

	manager.Stop()
	delete(ds.streamManagers, streamID)

	ds.logger.Info("Stream detection stopped", zap.String("stream_id", streamID))
	return nil
}

// GetStreamDetectionStats returns detection statistics for a stream
func (ds *DetectionService) GetStreamDetectionStats(streamID string) (*StreamDetectionStats, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	manager, exists := ds.streamManagers[streamID]
	if !exists {
		return nil, fmt.Errorf("no detection running for stream: %s", streamID)
	}

	return manager.GetStats(), nil
}

// GetAllStreamStats returns detection statistics for all streams
func (ds *DetectionService) GetAllStreamStats() map[string]*StreamDetectionStats {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	stats := make(map[string]*StreamDetectionStats)
	for streamID, manager := range ds.streamManagers {
		stats[streamID] = manager.GetStats()
	}

	return stats
}

// GetModelManager returns the model manager
func (ds *DetectionService) GetModelManager() *ModelManager {
	return ds.modelManager
}

// GetInferenceEngine returns the inference engine
func (ds *DetectionService) GetInferenceEngine() *InferenceEngine {
	return ds.inferenceEngine
}

// GetResultProcessor returns the result processor
func (ds *DetectionService) GetResultProcessor() *ResultProcessor {
	return ds.resultProcessor
}

// handleDetectionResult processes detection results
func (ds *DetectionService) handleDetectionResult(response *InferenceResponse) {
	if response.Error != nil {
		ds.logger.Error("Detection error", 
			zap.String("stream_id", response.StreamID),
			zap.Error(response.Error))
		return
	}

	// Process results through result processor
	processedObjects, err := ds.resultProcessor.ProcessDetections(context.Background(), response.Objects)
	if err != nil {
		ds.logger.Error("Failed to process detection results", zap.Error(err))
		return
	}

	// Update response with processed objects
	response.Objects = processedObjects

	// Call configured callback if available
	if ds.config.ResultCallback != nil {
		go ds.config.ResultCallback(response)
	}

	ds.logger.Debug("Detection result processed",
		zap.String("stream_id", response.StreamID),
		zap.Int("objects", len(processedObjects)),
		zap.Duration("process_time", response.ProcessTime))
}

// Stop stops the stream detection manager
func (sdm *StreamDetectionManager) Stop() {
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	sdm.isActive = false
	sdm.logger.Info("Stream detection manager stopped")
}

// GetStats returns stream detection statistics
func (sdm *StreamDetectionManager) GetStats() *StreamDetectionStats {
	sdm.stats.mutex.RLock()
	defer sdm.stats.mutex.RUnlock()

	// Return a copy
	return &StreamDetectionStats{
		StreamID:           sdm.stats.StreamID,
		TotalFrames:        sdm.stats.TotalFrames,
		ProcessedFrames:    sdm.stats.ProcessedFrames,
		DetectedObjects:    sdm.stats.DetectedObjects,
		AverageProcessTime: sdm.stats.AverageProcessTime,
		LastDetectionTime:  sdm.stats.LastDetectionTime,
		FPS:                sdm.stats.FPS,
		StartTime:          sdm.stats.StartTime,
	}
}

// updateStats updates stream detection statistics
func (sdm *StreamDetectionManager) updateStats(response *InferenceResponse) {
	sdm.stats.mutex.Lock()
	defer sdm.stats.mutex.Unlock()

	sdm.stats.TotalFrames++
	if response.Error == nil {
		sdm.stats.ProcessedFrames++
		sdm.stats.DetectedObjects += int64(len(response.Objects))
	}

	sdm.stats.LastDetectionTime = time.Now()

	// Calculate average process time
	if sdm.stats.ProcessedFrames > 0 {
		totalTime := time.Duration(sdm.stats.ProcessedFrames) * sdm.stats.AverageProcessTime
		totalTime += response.ProcessTime
		sdm.stats.AverageProcessTime = totalTime / time.Duration(sdm.stats.ProcessedFrames)
	} else {
		sdm.stats.AverageProcessTime = response.ProcessTime
	}

	// Calculate FPS
	elapsed := time.Since(sdm.stats.StartTime).Seconds()
	if elapsed > 0 {
		sdm.stats.FPS = float64(sdm.stats.ProcessedFrames) / elapsed
	}
}

// IsActive returns whether the stream detection manager is active
func (sdm *StreamDetectionManager) IsActive() bool {
	sdm.mutex.RLock()
	defer sdm.mutex.RUnlock()
	return sdm.isActive
}

// GetActiveStreams returns a list of streams with active detection
func (ds *DetectionService) GetActiveStreams() []string {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	streams := make([]string, 0, len(ds.streamManagers))
	for streamID, manager := range ds.streamManagers {
		if manager.IsActive() {
			streams = append(streams, streamID)
		}
	}

	return streams
}

// GetServiceStats returns overall detection service statistics
func (ds *DetectionService) GetServiceStats() map[string]interface{} {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	activeStreams := 0
	totalFrames := int64(0)
	totalObjects := int64(0)

	for _, manager := range ds.streamManagers {
		if manager.IsActive() {
			activeStreams++
		}
		stats := manager.GetStats()
		totalFrames += stats.TotalFrames
		totalObjects += stats.DetectedObjects
	}

	inferenceStats := ds.inferenceEngine.GetStats()
	modelStats := ds.modelManager.GetStats()

	return map[string]interface{}{
		"active_streams":     activeStreams,
		"total_streams":      len(ds.streamManagers),
		"total_frames":       totalFrames,
		"total_objects":      totalObjects,
		"inference_stats":    inferenceStats,
		"model_stats":        modelStats,
		"processor_config":   ds.resultProcessor.GetStats(),
	}
}
