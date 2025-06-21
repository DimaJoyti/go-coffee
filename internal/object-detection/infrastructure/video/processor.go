package video

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

// Processor implements the VideoProcessor interface using GoCV
type Processor struct {
	logger  *zap.Logger
	streams map[string]*StreamHandler
	mutex   sync.RWMutex
}

// StreamHandler manages a single video stream
type StreamHandler struct {
	ID          string
	Source      string
	Type        domain.StreamType
	Status      domain.StreamStatus
	Capture     *gocv.VideoCapture
	FrameRate   float64
	Width       int
	Height      int
	IsActive    bool
	StopChannel chan bool
	Logger      *zap.Logger
	mutex       sync.RWMutex
}

// NewProcessor creates a new video processor
func NewProcessor(logger *zap.Logger) *Processor {
	return &Processor{
		logger:  logger,
		streams: make(map[string]*StreamHandler),
	}
}

// OpenStream opens a video stream for processing
func (p *Processor) OpenStream(ctx context.Context, source string, streamType domain.StreamType) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Generate stream ID
	streamID := generateStreamID(source, streamType)

	// Check if stream already exists
	if _, exists := p.streams[streamID]; exists {
		return fmt.Errorf("stream already exists: %s", streamID)
	}

	p.logger.Info("Opening video stream",
		zap.String("stream_id", streamID),
		zap.String("source", source),
		zap.String("type", string(streamType)))

	// Create stream handler
	handler := &StreamHandler{
		ID:          streamID,
		Source:      source,
		Type:        streamType,
		Status:      domain.StreamStatusIdle,
		IsActive:    false,
		StopChannel: make(chan bool, 1),
		Logger:      p.logger.With(zap.String("stream_id", streamID)),
	}

	// Open video capture based on stream type
	var err error
	switch streamType {
	case domain.StreamTypeWebcam:
		err = handler.openWebcam()
	case domain.StreamTypeFile:
		err = handler.openFile()
	case domain.StreamTypeRTMP, domain.StreamTypeHTTP:
		err = handler.openNetwork()
	default:
		return fmt.Errorf("unsupported stream type: %s", streamType)
	}

	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	// Get stream properties
	if err := handler.getStreamProperties(); err != nil {
		handler.close()
		return fmt.Errorf("failed to get stream properties: %w", err)
	}

	handler.Status = domain.StreamStatusActive
	p.streams[streamID] = handler

	p.logger.Info("Video stream opened successfully",
		zap.String("stream_id", streamID),
		zap.Float64("fps", handler.FrameRate),
		zap.Int("width", handler.Width),
		zap.Int("height", handler.Height))

	return nil
}

// CloseStream closes a video stream
func (p *Processor) CloseStream(ctx context.Context, streamID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	handler, exists := p.streams[streamID]
	if !exists {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	p.logger.Info("Closing video stream", zap.String("stream_id", streamID))

	// Stop the stream
	handler.stop()

	// Remove from streams map
	delete(p.streams, streamID)

	p.logger.Info("Video stream closed", zap.String("stream_id", streamID))
	return nil
}

// ReadFrame reads the next frame from a video stream
func (p *Processor) ReadFrame(ctx context.Context, streamID string) ([]byte, error) {
	p.mutex.RLock()
	handler, exists := p.streams[streamID]
	p.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("stream not found: %s", streamID)
	}

	return handler.readFrame(ctx)
}

// GetFrameRate retrieves the frame rate of a video stream
func (p *Processor) GetFrameRate(ctx context.Context, streamID string) (float64, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	handler, exists := p.streams[streamID]
	if !exists {
		return 0, fmt.Errorf("stream not found: %s", streamID)
	}

	return handler.FrameRate, nil
}

// GetResolution retrieves the resolution of a video stream
func (p *Processor) GetResolution(ctx context.Context, streamID string) (width, height int, err error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	handler, exists := p.streams[streamID]
	if !exists {
		return 0, 0, fmt.Errorf("stream not found: %s", streamID)
	}

	return handler.Width, handler.Height, nil
}

// IsStreamActive checks if a video stream is active
func (p *Processor) IsStreamActive(ctx context.Context, streamID string) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	handler, exists := p.streams[streamID]
	if !exists {
		return false
	}

	return handler.IsActive
}

// GetActiveStreams returns a list of active stream IDs
func (p *Processor) GetActiveStreams() []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var activeStreams []string
	for id, handler := range p.streams {
		if handler.IsActive {
			activeStreams = append(activeStreams, id)
		}
	}

	return activeStreams
}

// Close closes all streams and cleans up resources
func (p *Processor) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.logger.Info("Closing video processor")

	for streamID, handler := range p.streams {
		p.logger.Debug("Closing stream", zap.String("stream_id", streamID))
		handler.stop()
	}

	// Clear streams map
	p.streams = make(map[string]*StreamHandler)

	p.logger.Info("Video processor closed")
	return nil
}

// generateStreamID generates a unique stream ID
func generateStreamID(source string, streamType domain.StreamType) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%s_%d", streamType, hashString(source), timestamp)
}

// hashString creates a simple hash of a string
func hashString(s string) string {
	hash := uint32(0)
	for _, c := range s {
		hash = hash*31 + uint32(c)
	}
	return fmt.Sprintf("%x", hash)
}
