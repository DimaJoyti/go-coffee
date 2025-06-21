package streaming

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/tracking"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/video"
	"go.uber.org/zap"
)

// FrameStreamer streams video frames with detection overlays to WebSocket clients
type FrameStreamer struct {
	logger         *zap.Logger
	hub            *Hub
	config         FrameStreamerConfig
	qualityAdapter *QualityAdapter
	isRunning      bool
	mutex          sync.RWMutex
	stats          *FrameStreamerStats
}

// FrameStreamerConfig configures the frame streamer
type FrameStreamerConfig struct {
	EnableFrameStreaming   bool          // Enable frame streaming
	EnableDetectionOverlay bool          // Draw detection overlays
	EnableTrackingOverlay  bool          // Draw tracking overlays
	EnableTrajectories     bool          // Draw trajectory paths
	MaxFrameRate          float64       // Maximum frame rate for streaming
	JpegQuality           int           // JPEG compression quality (1-100)
	MaxFrameSize          int           // Maximum frame size in bytes
	OverlayConfig         OverlayConfig // Overlay drawing configuration
	AdaptiveQuality       bool          // Enable adaptive quality based on bandwidth
	BufferSize            int           // Frame buffer size
}

// OverlayConfig configures detection and tracking overlays
type OverlayConfig struct {
	BoundingBoxColor     ColorRGBA `json:"bounding_box_color"`
	TrackingBoxColor     ColorRGBA `json:"tracking_box_color"`
	TrajectoryColor      ColorRGBA `json:"trajectory_color"`
	TextColor            ColorRGBA `json:"text_color"`
	BackgroundColor      ColorRGBA `json:"background_color"`
	LineThickness        int       `json:"line_thickness"`
	FontScale            float64   `json:"font_scale"`
	ShowConfidence       bool      `json:"show_confidence"`
	ShowTrackID          bool      `json:"show_track_id"`
	ShowClass            bool      `json:"show_class"`
	ShowTrajectoryLength int       `json:"show_trajectory_length"`
}

// ColorRGBA represents an RGBA color
type ColorRGBA struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

// FrameStreamerStats tracks frame streaming performance
type FrameStreamerStats struct {
	FramesStreamed     int64
	BytesStreamed      int64
	StreamErrors       int64
	FramesDropped      int64
	AverageFrameSize   int64
	AverageProcessTime time.Duration
	CurrentFPS         float64
	StartTime          time.Time
	LastFrameTime      time.Time
	mutex              sync.RWMutex
}

// FrameStreamData represents frame data for streaming
type FrameStreamData struct {
	StreamID     string    `json:"stream_id"`
	FrameID      string    `json:"frame_id"`
	Timestamp    time.Time `json:"timestamp"`
	ImageData    string    `json:"image_data"`    // Base64 encoded JPEG
	ImageFormat  string    `json:"image_format"`  // "jpeg"
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Quality      int       `json:"quality"`
	Size         int       `json:"size"`          // Size in bytes
	ProcessTime  time.Duration `json:"process_time"`
	HasOverlays  bool      `json:"has_overlays"`
}

// DefaultFrameStreamerConfig returns default configuration
func DefaultFrameStreamerConfig() FrameStreamerConfig {
	return FrameStreamerConfig{
		EnableFrameStreaming:   true,
		EnableDetectionOverlay: true,
		EnableTrackingOverlay:  true,
		EnableTrajectories:     true,
		MaxFrameRate:          15.0, // 15 FPS max
		JpegQuality:           75,
		MaxFrameSize:          500 * 1024, // 500KB
		AdaptiveQuality:       true,
		BufferSize:            10,
		OverlayConfig: OverlayConfig{
			BoundingBoxColor:     ColorRGBA{R: 0, G: 255, B: 0, A: 255},     // Green
			TrackingBoxColor:     ColorRGBA{R: 255, G: 0, B: 0, A: 255},     // Red
			TrajectoryColor:      ColorRGBA{R: 0, G: 0, B: 255, A: 255},     // Blue
			TextColor:            ColorRGBA{R: 255, G: 255, B: 255, A: 255}, // White
			BackgroundColor:      ColorRGBA{R: 0, G: 0, B: 0, A: 128},       // Semi-transparent black
			LineThickness:        2,
			FontScale:            0.6,
			ShowConfidence:       true,
			ShowTrackID:          true,
			ShowClass:            true,
			ShowTrajectoryLength: 20,
		},
	}
}

// NewFrameStreamer creates a new frame streamer
func NewFrameStreamer(logger *zap.Logger, hub *Hub, config FrameStreamerConfig) *FrameStreamer {
	return &FrameStreamer{
		logger: logger.With(zap.String("component", "frame_streamer")),
		hub:    hub,
		config: config,
		stats: &FrameStreamerStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the frame streamer
func (fs *FrameStreamer) Start(ctx context.Context) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if fs.isRunning {
		return fmt.Errorf("frame streamer is already running")
	}

	fs.logger.Info("Starting frame streamer",
		zap.Bool("frame_streaming_enabled", fs.config.EnableFrameStreaming),
		zap.Bool("detection_overlay_enabled", fs.config.EnableDetectionOverlay),
		zap.Bool("tracking_overlay_enabled", fs.config.EnableTrackingOverlay),
		zap.Float64("max_frame_rate", fs.config.MaxFrameRate),
		zap.Int("jpeg_quality", fs.config.JpegQuality))

	fs.isRunning = true
	fs.stats.StartTime = time.Now()

	fs.logger.Info("Frame streamer started")
	return nil
}

// Stop stops the frame streamer
func (fs *FrameStreamer) Stop() error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if !fs.isRunning {
		return fmt.Errorf("frame streamer is not running")
	}

	fs.logger.Info("Stopping frame streamer")
	fs.isRunning = false

	fs.logger.Info("Frame streamer stopped")
	return nil
}

// StreamFrame streams a video frame with overlays to clients
func (fs *FrameStreamer) StreamFrame(streamID string, frame *video.Frame, detections []domain.DetectedObject, tracks []*tracking.Track) error {
	if !fs.IsRunning() || !fs.config.EnableFrameStreaming {
		return nil
	}

	startTime := time.Now()

	// Check frame rate limiting
	if !fs.shouldStreamFrame() {
		fs.updateStats(func(stats *FrameStreamerStats) {
			stats.FramesDropped++
		})
		return nil
	}

	// For now, just pass through the frame data without OpenCV processing
	// In a full implementation, this would decode, add overlays, and re-encode
	imageData := frame.Data
	hasOverlays := false

	// Apply quality adaptation if available
	if fs.qualityAdapter != nil {
		// Get optimal quality for each client (simplified for now)
		// In a real implementation, this would be per-client
		quality := fs.qualityAdapter.GetOptimalQuality("default")
		if quality != nil {
			fs.config.JpegQuality = quality.JpegQuality
			fs.config.MaxFrameRate = quality.MaxFrameRate
		}
	}

	// Apply adaptive quality if frame is too large
	if fs.config.AdaptiveQuality && len(imageData) > fs.config.MaxFrameSize {
		// Simplified quality reduction - in practice would re-encode with lower quality
		fs.logger.Debug("Frame size exceeds limit, would apply quality reduction",
			zap.Int("current_size", len(imageData)),
			zap.Int("max_size", fs.config.MaxFrameSize))
	}

	// Create stream data
	streamData := FrameStreamData{
		StreamID:    streamID,
		FrameID:     frame.ID,
		Timestamp:   time.Now(),
		ImageData:   base64.StdEncoding.EncodeToString(imageData),
		ImageFormat: "jpeg",
		Width:       640,  // Default dimensions
		Height:      480,
		Quality:     fs.config.JpegQuality,
		Size:        len(imageData),
		ProcessTime: time.Since(startTime),
		HasOverlays: hasOverlays,
	}

	// Create message
	message := &Message{
		Type:      MessageTypeFrame,
		StreamID:  streamID,
		Timestamp: time.Now(),
		Data:      streamData,
	}

	// Stream to clients
	err := fs.hub.BroadcastToStream(streamID, message)
	if err != nil {
		fs.updateStats(func(stats *FrameStreamerStats) {
			stats.StreamErrors++
		})
		return fmt.Errorf("failed to stream frame: %w", err)
	}

	// Update stats
	fs.updateStats(func(stats *FrameStreamerStats) {
		stats.FramesStreamed++
		stats.BytesStreamed += int64(len(imageData))
		stats.LastFrameTime = time.Now()

		// Update average frame size
		if stats.FramesStreamed > 0 {
			stats.AverageFrameSize = stats.BytesStreamed / stats.FramesStreamed
		}

		// Update average process time
		if stats.FramesStreamed > 0 {
			totalTime := time.Duration(stats.FramesStreamed) * stats.AverageProcessTime
			totalTime += time.Since(startTime)
			stats.AverageProcessTime = totalTime / time.Duration(stats.FramesStreamed)
		} else {
			stats.AverageProcessTime = time.Since(startTime)
		}

		// Calculate current FPS
		elapsed := time.Since(stats.StartTime).Seconds()
		if elapsed > 0 {
			stats.CurrentFPS = float64(stats.FramesStreamed) / elapsed
		}
	})

	fs.logger.Debug("Frame streamed",
		zap.String("stream_id", streamID),
		zap.String("frame_id", frame.ID),
		zap.Int("frame_size", len(imageData)),
		zap.Bool("has_overlays", hasOverlays),
		zap.Duration("process_time", time.Since(startTime)))

	return nil
}

// shouldStreamFrame checks if a frame should be streamed based on rate limiting
func (fs *FrameStreamer) shouldStreamFrame() bool {
	fs.stats.mutex.RLock()
	lastFrameTime := fs.stats.LastFrameTime
	fs.stats.mutex.RUnlock()

	if lastFrameTime.IsZero() {
		return true
	}

	minInterval := time.Duration(1000.0/fs.config.MaxFrameRate) * time.Millisecond
	return time.Since(lastFrameTime) >= minInterval
}

// encodeFrame encodes frame data to JPEG (simplified version)
func (fs *FrameStreamer) encodeFrame(frameData []byte) ([]byte, error) {
	// For now, just return the original frame data
	// In a full implementation, this would decode, process, and re-encode
	return frameData, nil
}

// Simplified overlay methods (OpenCV-dependent methods removed for now)
// In a full implementation, these would use image processing libraries to draw overlays

func (fs *FrameStreamer) drawDetectionOverlays(frameData []byte, detections []domain.DetectedObject) []byte {
	// Placeholder for detection overlay drawing
	// Would decode image, draw bounding boxes and labels, then re-encode
	fs.logger.Debug("Would draw detection overlays", zap.Int("detections", len(detections)))
	return frameData
}

func (fs *FrameStreamer) drawTrackingOverlays(frameData []byte, tracks []*tracking.Track) []byte {
	// Placeholder for tracking overlay drawing
	// Would decode image, draw tracking boxes and trajectories, then re-encode
	fs.logger.Debug("Would draw tracking overlays", zap.Int("tracks", len(tracks)))
	return frameData
}

// IsRunning returns whether the streamer is running
func (fs *FrameStreamer) IsRunning() bool {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	return fs.isRunning
}

// GetStats returns streamer statistics
func (fs *FrameStreamer) GetStats() *FrameStreamerStats {
	fs.stats.mutex.RLock()
	defer fs.stats.mutex.RUnlock()

	return &FrameStreamerStats{
		FramesStreamed:     fs.stats.FramesStreamed,
		BytesStreamed:      fs.stats.BytesStreamed,
		StreamErrors:       fs.stats.StreamErrors,
		FramesDropped:      fs.stats.FramesDropped,
		AverageFrameSize:   fs.stats.AverageFrameSize,
		AverageProcessTime: fs.stats.AverageProcessTime,
		CurrentFPS:         fs.stats.CurrentFPS,
		StartTime:          fs.stats.StartTime,
		LastFrameTime:      fs.stats.LastFrameTime,
	}
}

// updateStats updates streamer statistics
func (fs *FrameStreamer) updateStats(updateFunc func(*FrameStreamerStats)) {
	fs.stats.mutex.Lock()
	defer fs.stats.mutex.Unlock()
	updateFunc(fs.stats)
}

// GetConfig returns the streamer configuration
func (fs *FrameStreamer) GetConfig() FrameStreamerConfig {
	return fs.config
}

// UpdateConfig updates the streamer configuration
func (fs *FrameStreamer) UpdateConfig(config FrameStreamerConfig) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.config = config
	fs.logger.Info("Frame streamer configuration updated",
		zap.Bool("frame_streaming_enabled", config.EnableFrameStreaming),
		zap.Float64("max_frame_rate", config.MaxFrameRate),
		zap.Int("jpeg_quality", config.JpegQuality))
}
