package streaming

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/detection"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/tracking"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/infrastructure/video"
	"go.uber.org/zap"
)

// StreamingService manages real-time streaming of detection and tracking results
type StreamingService struct {
	logger            *zap.Logger
	hub               *Hub
	detectionStreamer *DetectionStreamer
	frameStreamer     *FrameStreamer
	config            StreamingServiceConfig
	isRunning         bool
	mutex             sync.RWMutex
	stats             *StreamingServiceStats
}

// StreamingServiceConfig configures the streaming service
type StreamingServiceConfig struct {
	HubConfig               HubConfig
	DetectionStreamerConfig DetectionStreamerConfig
	FrameStreamerConfig     FrameStreamerConfig
	EnableAuthentication    bool
	AuthTokenHeader         string
	ValidTokens             []string
	EnableRateLimit         bool
	RateLimitPerSecond      int
	EnableMetrics           bool
	MetricsPath             string
}

// StreamingServiceStats tracks service performance
type StreamingServiceStats struct {
	TotalConnections    int64
	ActiveConnections   int
	TotalMessages       int64
	TotalBytes          int64
	AuthFailures        int64
	RateLimitExceeded   int64
	StartTime           time.Time
	LastActivity        time.Time
	mutex               sync.RWMutex
}

// DefaultStreamingServiceConfig returns default configuration
func DefaultStreamingServiceConfig() StreamingServiceConfig {
	return StreamingServiceConfig{
		HubConfig:               DefaultHubConfig(),
		DetectionStreamerConfig: DefaultDetectionStreamerConfig(),
		FrameStreamerConfig:     DefaultFrameStreamerConfig(),
		EnableAuthentication:    false,
		AuthTokenHeader:         "Authorization",
		ValidTokens:             []string{},
		EnableRateLimit:         true,
		RateLimitPerSecond:      100,
		EnableMetrics:           true,
		MetricsPath:             "/metrics",
	}
}

// NewStreamingService creates a new streaming service
func NewStreamingService(logger *zap.Logger, config StreamingServiceConfig) *StreamingService {
	// Create WebSocket hub
	hub := NewHub(logger, config.HubConfig)

	// Create detection streamer
	detectionStreamer := NewDetectionStreamer(logger, hub, config.DetectionStreamerConfig)

	// Create frame streamer
	frameStreamer := NewFrameStreamer(logger, hub, config.FrameStreamerConfig)

	return &StreamingService{
		logger:            logger.With(zap.String("component", "streaming_service")),
		hub:               hub,
		detectionStreamer: detectionStreamer,
		frameStreamer:     frameStreamer,
		config:            config,
		stats: &StreamingServiceStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the streaming service
func (ss *StreamingService) Start(ctx context.Context) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	if ss.isRunning {
		return fmt.Errorf("streaming service is already running")
	}

	ss.logger.Info("Starting streaming service",
		zap.Bool("authentication_enabled", ss.config.EnableAuthentication),
		zap.Bool("rate_limit_enabled", ss.config.EnableRateLimit),
		zap.Int("rate_limit_per_second", ss.config.RateLimitPerSecond))

	// Start hub
	if err := ss.hub.Start(ctx); err != nil {
		return fmt.Errorf("failed to start WebSocket hub: %w", err)
	}

	// Start detection streamer
	if err := ss.detectionStreamer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start detection streamer: %w", err)
	}

	// Start frame streamer
	if err := ss.frameStreamer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start frame streamer: %w", err)
	}

	ss.isRunning = true
	ss.stats.StartTime = time.Now()

	ss.logger.Info("Streaming service started")
	return nil
}

// Stop stops the streaming service
func (ss *StreamingService) Stop() error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	if !ss.isRunning {
		return fmt.Errorf("streaming service is not running")
	}

	ss.logger.Info("Stopping streaming service")

	// Stop components
	if err := ss.frameStreamer.Stop(); err != nil {
		ss.logger.Error("Failed to stop frame streamer", zap.Error(err))
	}

	if err := ss.detectionStreamer.Stop(); err != nil {
		ss.logger.Error("Failed to stop detection streamer", zap.Error(err))
	}

	if err := ss.hub.Stop(); err != nil {
		ss.logger.Error("Failed to stop WebSocket hub", zap.Error(err))
	}

	ss.isRunning = false

	ss.logger.Info("Streaming service stopped")
	return nil
}

// HandleWebSocket handles WebSocket connection requests
func (ss *StreamingService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !ss.IsRunning() {
		http.Error(w, "Streaming service not running", http.StatusServiceUnavailable)
		return
	}

	// Authentication check
	if ss.config.EnableAuthentication {
		if !ss.authenticateRequest(r) {
			ss.updateStats(func(stats *StreamingServiceStats) {
				stats.AuthFailures++
			})
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}
	}

	// Rate limiting check
	if ss.config.EnableRateLimit {
		if !ss.checkRateLimit(r) {
			ss.updateStats(func(stats *StreamingServiceStats) {
				stats.RateLimitExceeded++
			})
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
	}

	// Update stats
	ss.updateStats(func(stats *StreamingServiceStats) {
		stats.TotalConnections++
		stats.LastActivity = time.Now()
	})

	// Handle WebSocket upgrade
	ss.hub.HandleWebSocket(w, r)
}

// StreamDetectionResults streams detection results to connected clients
func (ss *StreamingService) StreamDetectionResults(response *detection.InferenceResponse) error {
	if !ss.IsRunning() {
		return fmt.Errorf("streaming service is not running")
	}

	return ss.detectionStreamer.StreamDetectionResults(response)
}

// StreamTrackingUpdate streams tracking updates to connected clients
func (ss *StreamingService) StreamTrackingUpdate(update *tracking.TrackingUpdate) error {
	if !ss.IsRunning() {
		return fmt.Errorf("streaming service is not running")
	}

	return ss.detectionStreamer.StreamTrackingUpdate(update)
}

// StreamFrame streams video frames with overlays to connected clients
func (ss *StreamingService) StreamFrame(streamID string, frame *video.Frame, detections []domain.DetectedObject, tracks []*tracking.Track) error {
	if !ss.IsRunning() {
		return fmt.Errorf("streaming service is not running")
	}

	return ss.frameStreamer.StreamFrame(streamID, frame, detections, tracks)
}

// CreateDetectionCallback creates a callback for detection results
func (ss *StreamingService) CreateDetectionCallback() detection.DetectionCallback {
	return func(response *detection.InferenceResponse) {
		if err := ss.StreamDetectionResults(response); err != nil {
			ss.logger.Error("Failed to stream detection results",
				zap.String("stream_id", response.StreamID),
				zap.Error(err))
		}
	}
}

// CreateTrackingCallback creates a callback for tracking updates
func (ss *StreamingService) CreateTrackingCallback() tracking.TrackingCallback {
	return func(update *tracking.TrackingUpdate) {
		if err := ss.StreamTrackingUpdate(update); err != nil {
			ss.logger.Error("Failed to stream tracking update",
				zap.String("stream_id", update.StreamID),
				zap.Error(err))
		}
	}
}

// GetConnectedClients returns the number of connected clients
func (ss *StreamingService) GetConnectedClients() int {
	return ss.hub.GetConnectedClients()
}

// GetClientInfo returns information about connected clients
func (ss *StreamingService) GetClientInfo() []map[string]interface{} {
	// This would require extending the hub to expose client information
	// For now, return basic stats
	return []map[string]interface{}{
		{
			"connected_clients": ss.hub.GetConnectedClients(),
			"hub_stats":         ss.hub.GetStats(),
		},
	}
}

// GetStats returns streaming service statistics
func (ss *StreamingService) GetStats() *StreamingServiceStats {
	ss.stats.mutex.RLock()
	defer ss.stats.mutex.RUnlock()

	hubStats := ss.hub.GetStats()
	detectionStats := ss.detectionStreamer.GetStats()
	frameStats := ss.frameStreamer.GetStats()

	return &StreamingServiceStats{
		TotalConnections:  ss.stats.TotalConnections,
		ActiveConnections: hubStats.ConnectedClients,
		TotalMessages:     hubStats.MessagesSent + detectionStats.DetectionsSent + frameStats.FramesStreamed,
		TotalBytes:        hubStats.BytesSent + detectionStats.BytesSent + frameStats.BytesStreamed,
		AuthFailures:      ss.stats.AuthFailures,
		RateLimitExceeded: ss.stats.RateLimitExceeded,
		StartTime:         ss.stats.StartTime,
		LastActivity:      ss.stats.LastActivity,
	}
}

// GetDetailedStats returns detailed statistics from all components
func (ss *StreamingService) GetDetailedStats() map[string]interface{} {
	return map[string]interface{}{
		"service":           ss.GetStats(),
		"hub":               ss.hub.GetStats(),
		"detection_streamer": ss.detectionStreamer.GetStats(),
		"frame_streamer":    ss.frameStreamer.GetStats(),
	}
}

// IsRunning returns whether the service is running
func (ss *StreamingService) IsRunning() bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.isRunning
}

// authenticateRequest validates authentication for incoming requests
func (ss *StreamingService) authenticateRequest(r *http.Request) bool {
	if !ss.config.EnableAuthentication {
		return true
	}

	token := r.Header.Get(ss.config.AuthTokenHeader)
	if token == "" {
		// Check query parameter as fallback
		token = r.URL.Query().Get("token")
	}

	if token == "" {
		return false
	}

	// Check against valid tokens
	for _, validToken := range ss.config.ValidTokens {
		if token == validToken {
			return true
		}
	}

	return false
}

// checkRateLimit checks if the request is within rate limits
func (ss *StreamingService) checkRateLimit(r *http.Request) bool {
	if !ss.config.EnableRateLimit {
		return true
	}

	// Simple rate limiting implementation
	// In production, you'd want to use a more sophisticated rate limiter
	// like token bucket or sliding window with Redis
	
	// For now, just return true (rate limiting would be implemented here)
	return true
}

// updateStats updates service statistics
func (ss *StreamingService) updateStats(updateFunc func(*StreamingServiceStats)) {
	ss.stats.mutex.Lock()
	defer ss.stats.mutex.Unlock()
	updateFunc(ss.stats)
}

// GetConfig returns the service configuration
func (ss *StreamingService) GetConfig() StreamingServiceConfig {
	return ss.config
}

// UpdateConfig updates the service configuration
func (ss *StreamingService) UpdateConfig(config StreamingServiceConfig) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ss.config = config

	// Update component configurations
	ss.hub.UpdateConfig(config.HubConfig)
	ss.detectionStreamer.UpdateConfig(config.DetectionStreamerConfig)
	ss.frameStreamer.UpdateConfig(config.FrameStreamerConfig)

	ss.logger.Info("Streaming service configuration updated",
		zap.Bool("authentication_enabled", config.EnableAuthentication),
		zap.Bool("rate_limit_enabled", config.EnableRateLimit))
}

// BroadcastMessage broadcasts a custom message to all connected clients
func (ss *StreamingService) BroadcastMessage(message *Message) error {
	if !ss.IsRunning() {
		return fmt.Errorf("streaming service is not running")
	}

	return ss.hub.Broadcast(message)
}

// BroadcastToStream broadcasts a message to clients subscribed to a specific stream
func (ss *StreamingService) BroadcastToStream(streamID string, message *Message) error {
	if !ss.IsRunning() {
		return fmt.Errorf("streaming service is not running")
	}

	return ss.hub.BroadcastToStream(streamID, message)
}

// SendStatus sends a status message to all clients
func (ss *StreamingService) SendStatus(status string, data interface{}) error {
	message := &Message{
		Type:      MessageTypeStatus,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status": status,
			"data":   data,
		},
	}

	return ss.BroadcastMessage(message)
}

// SendError sends an error message to all clients
func (ss *StreamingService) SendError(errorMsg string) error {
	message := &Message{
		Type:      MessageTypeError,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"error": errorMsg,
		},
	}

	return ss.BroadcastMessage(message)
}
