package video

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
)

// StreamManager manages video streams and their processing pipelines
type StreamManager struct {
	logger    *zap.Logger
	streams   map[string]*ManagedStream
	processor *Processor
	mutex     sync.RWMutex
}

// ManagedStream represents a managed video stream with its pipeline
type ManagedStream struct {
	Stream     *domain.VideoStream
	Handler    *StreamHandler
	Pipeline   *Pipeline
	Context    context.Context
	Cancel     context.CancelFunc
	IsActive   bool
	StartTime  time.Time
	FrameCount int64
	mutex      sync.RWMutex
}

// StreamManagerConfig configures the stream manager
type StreamManagerConfig struct {
	MaxStreams         int
	DefaultWorkers     int
	DefaultBufferSize  int
	HealthCheckInterval time.Duration
}

// DefaultStreamManagerConfig returns default configuration
func DefaultStreamManagerConfig() StreamManagerConfig {
	return StreamManagerConfig{
		MaxStreams:          10,
		DefaultWorkers:      2,
		DefaultBufferSize:   10,
		HealthCheckInterval: 30 * time.Second,
	}
}

// NewStreamManager creates a new stream manager
func NewStreamManager(logger *zap.Logger, config StreamManagerConfig) *StreamManager {
	processor := NewProcessor(logger)

	return &StreamManager{
		logger:    logger.With(zap.String("component", "stream_manager")),
		streams:   make(map[string]*ManagedStream),
		processor: processor,
	}
}

// CreateStream creates and registers a new video stream
func (sm *StreamManager) CreateStream(ctx context.Context, stream *domain.VideoStream) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.streams[stream.ID]; exists {
		return fmt.Errorf("stream already exists: %s", stream.ID)
	}

	sm.logger.Info("Creating stream",
		zap.String("stream_id", stream.ID),
		zap.String("source", stream.Source),
		zap.String("type", string(stream.Type)))

	// Create managed stream
	streamCtx, cancel := context.WithCancel(ctx)
	managedStream := &ManagedStream{
		Stream:    stream,
		Context:   streamCtx,
		Cancel:    cancel,
		IsActive:  false,
		StartTime: time.Now(),
	}

	// Create pipeline
	pipelineConfig := PipelineConfig{
		StreamID:   stream.ID,
		Workers:    2, // Default workers
		BufferSize: 10, // Default buffer size
	}

	pipeline := NewPipeline(sm.logger, pipelineConfig)
	managedStream.Pipeline = pipeline

	// Add preprocessing if configured
	if stream.Config.DetectionThreshold > 0 {
		preprocessConfig := DefaultPreprocessingConfig()
		preprocessConfig.TargetWidth = 640
		preprocessConfig.TargetHeight = 640
		
		preprocessor := NewCompositeProcessor(sm.logger, preprocessConfig)
		pipeline.AddProcessor(preprocessor)
	}

	sm.streams[stream.ID] = managedStream

	sm.logger.Info("Stream created successfully", zap.String("stream_id", stream.ID))
	return nil
}

// StartStream starts processing a video stream
func (sm *StreamManager) StartStream(ctx context.Context, streamID string) error {
	sm.mutex.Lock()
	managedStream, exists := sm.streams[streamID]
	sm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	managedStream.mutex.Lock()
	defer managedStream.mutex.Unlock()

	if managedStream.IsActive {
		return fmt.Errorf("stream is already active: %s", streamID)
	}

	sm.logger.Info("Starting stream", zap.String("stream_id", streamID))

	// Open video stream
	err := sm.processor.OpenStream(ctx, managedStream.Stream.Source, managedStream.Stream.Type)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	// Start pipeline
	err = managedStream.Pipeline.Start(managedStream.Context)
	if err != nil {
		sm.processor.CloseStream(ctx, streamID)
		return fmt.Errorf("failed to start pipeline: %w", err)
	}

	// Start frame processing goroutine
	go sm.processStreamFrames(managedStream)

	managedStream.IsActive = true
	managedStream.StartTime = time.Now()
	managedStream.Stream.Status = domain.StreamStatusActive

	sm.logger.Info("Stream started successfully", zap.String("stream_id", streamID))
	return nil
}

// StopStream stops processing a video stream
func (sm *StreamManager) StopStream(ctx context.Context, streamID string) error {
	sm.mutex.Lock()
	managedStream, exists := sm.streams[streamID]
	sm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	managedStream.mutex.Lock()
	defer managedStream.mutex.Unlock()

	if !managedStream.IsActive {
		return fmt.Errorf("stream is not active: %s", streamID)
	}

	sm.logger.Info("Stopping stream", zap.String("stream_id", streamID))

	// Cancel context to stop processing
	managedStream.Cancel()

	// Stop pipeline
	managedStream.Pipeline.Stop()

	// Close video stream
	sm.processor.CloseStream(ctx, streamID)

	managedStream.IsActive = false
	managedStream.Stream.Status = domain.StreamStatusStopped

	sm.logger.Info("Stream stopped successfully", zap.String("stream_id", streamID))
	return nil
}

// DeleteStream removes a video stream
func (sm *StreamManager) DeleteStream(ctx context.Context, streamID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	managedStream, exists := sm.streams[streamID]
	if !exists {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	sm.logger.Info("Deleting stream", zap.String("stream_id", streamID))

	// Stop stream if active
	if managedStream.IsActive {
		managedStream.Cancel()
		managedStream.Pipeline.Stop()
		sm.processor.CloseStream(ctx, streamID)
	}

	// Remove from streams map
	delete(sm.streams, streamID)

	sm.logger.Info("Stream deleted successfully", zap.String("stream_id", streamID))
	return nil
}

// GetStream retrieves a video stream
func (sm *StreamManager) GetStream(streamID string) (*domain.VideoStream, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	managedStream, exists := sm.streams[streamID]
	if !exists {
		return nil, fmt.Errorf("stream not found: %s", streamID)
	}

	return managedStream.Stream, nil
}

// GetAllStreams retrieves all video streams
func (sm *StreamManager) GetAllStreams() []*domain.VideoStream {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	streams := make([]*domain.VideoStream, 0, len(sm.streams))
	for _, managedStream := range sm.streams {
		streams = append(streams, managedStream.Stream)
	}

	return streams
}

// GetActiveStreams retrieves all active video streams
func (sm *StreamManager) GetActiveStreams() []*domain.VideoStream {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	streams := make([]*domain.VideoStream, 0)
	for _, managedStream := range sm.streams {
		if managedStream.IsActive {
			streams = append(streams, managedStream.Stream)
		}
	}

	return streams
}

// GetStreamStats retrieves statistics for a stream
func (sm *StreamManager) GetStreamStats(streamID string) (map[string]interface{}, error) {
	sm.mutex.RLock()
	managedStream, exists := sm.streams[streamID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("stream not found: %s", streamID)
	}

	managedStream.mutex.RLock()
	defer managedStream.mutex.RUnlock()

	stats := map[string]interface{}{
		"stream_id":    streamID,
		"is_active":    managedStream.IsActive,
		"start_time":   managedStream.StartTime,
		"frame_count":  managedStream.FrameCount,
		"uptime":       time.Since(managedStream.StartTime).String(),
	}

	// Add pipeline stats if available
	if managedStream.Pipeline != nil {
		pipelineStats := managedStream.Pipeline.GetStats()
		stats["pipeline"] = map[string]interface{}{
			"frames_processed": pipelineStats.FramesProcessed,
			"frames_dropped":   pipelineStats.FramesDropped,
			"avg_process_time": pipelineStats.AverageProcessTime.String(),
			"fps":              pipelineStats.FPS,
		}

		inputBuffer, outputBuffer := managedStream.Pipeline.GetBufferUsage()
		stats["buffer_usage"] = map[string]interface{}{
			"input":  inputBuffer,
			"output": outputBuffer,
		}
	}

	return stats, nil
}

// processStreamFrames processes frames from a video stream
func (sm *StreamManager) processStreamFrames(managedStream *ManagedStream) {
	streamID := managedStream.Stream.ID
	logger := sm.logger.With(zap.String("stream_id", streamID))

	logger.Info("Starting frame processing")

	frameNum := 0
	ticker := time.NewTicker(time.Duration(1000.0/float64(managedStream.Stream.Config.FPS)) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-managedStream.Context.Done():
			logger.Info("Frame processing stopped by context")
			return

		case <-ticker.C:
			// Read frame from processor
			frameData, err := sm.processor.ReadFrame(managedStream.Context, streamID)
			if err != nil {
				if err.Error() == "end of stream" {
					logger.Info("End of stream reached")
					managedStream.Stream.Status = domain.StreamStatusStopped
					return
				}
				logger.Error("Failed to read frame", zap.Error(err))
				continue
			}

			// Create frame
			frame := NewFrame(streamID, frameNum, frameData)

			// Send to pipeline
			err = managedStream.Pipeline.ProcessFrame(frame)
			if err != nil {
				logger.Warn("Failed to process frame", zap.Error(err))
				continue
			}

			managedStream.mutex.Lock()
			managedStream.FrameCount++
			frameNum++
			managedStream.mutex.Unlock()
		}
	}
}

// GetProcessedFrames returns the output channel for processed frames
func (sm *StreamManager) GetProcessedFrames(streamID string) (<-chan *Frame, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	managedStream, exists := sm.streams[streamID]
	if !exists {
		return nil, fmt.Errorf("stream not found: %s", streamID)
	}

	if managedStream.Pipeline == nil {
		return nil, fmt.Errorf("pipeline not initialized for stream: %s", streamID)
	}

	return managedStream.Pipeline.GetOutputChannel(), nil
}

// Close closes the stream manager and all streams
func (sm *StreamManager) Close() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.logger.Info("Closing stream manager")

	// Stop all streams
	for streamID, managedStream := range sm.streams {
		sm.logger.Debug("Stopping stream", zap.String("stream_id", streamID))
		
		if managedStream.IsActive {
			managedStream.Cancel()
			managedStream.Pipeline.Stop()
		}
	}

	// Close processor
	sm.processor.Close()

	// Clear streams map
	sm.streams = make(map[string]*ManagedStream)

	sm.logger.Info("Stream manager closed")
	return nil
}
