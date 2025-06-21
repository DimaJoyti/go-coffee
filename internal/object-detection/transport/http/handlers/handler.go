package handlers

import (
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Handler represents the HTTP handler
type Handler struct {
	logger *zap.Logger
	config *config.Config
}

// NewHandler creates a new HTTP handler
func NewHandler(logger *zap.Logger, cfg *config.Config) *Handler {
	return &Handler{
		logger: logger,
		config: cfg,
	}
}

// SetupRoutes sets up all HTTP routes
func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Health check routes
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// Metrics endpoint
	if h.config.Monitoring.Enabled {
		router.GET(h.config.Monitoring.MetricsPath, gin.WrapH(promhttp.Handler()))
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Stream management
		streams := v1.Group("/streams")
		{
			streams.POST("", h.CreateStream)
			streams.GET("", h.GetStreams)
			streams.GET("/:id", h.GetStream)
			streams.PUT("/:id", h.UpdateStream)
			streams.DELETE("/:id", h.DeleteStream)
			streams.POST("/:id/start", h.StartStream)
			streams.POST("/:id/stop", h.StopStream)
		}

		// Detection management
		detection := v1.Group("/detection")
		{
			detection.POST("/start", h.StartDetection)
			detection.POST("/stop", h.StopDetection)
			detection.GET("/results", h.GetDetectionResults)
			detection.GET("/stats", h.GetDetectionStats)
		}

		// Model management
		models := v1.Group("/models")
		{
			models.POST("", h.UploadModel)
			models.GET("", h.GetModels)
			models.GET("/:id", h.GetModel)
			models.PUT("/:id/activate", h.ActivateModel)
			models.DELETE("/:id", h.DeleteModel)
		}

		// Alert management
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", h.GetAlerts)
			alerts.PUT("/:id/acknowledge", h.AcknowledgeAlert)
		}

		// Tracking management
		tracking := v1.Group("/tracking")
		{
			tracking.GET("/active/:stream_id", h.GetActiveTracking)
			tracking.GET("/history/:tracking_id", h.GetTrackingHistory)
		}
	}

	// WebSocket routes
	if h.config.WebSocket.Enabled {
		router.GET("/ws/detections", h.WebSocketDetections)
		router.GET("/ws/alerts", h.WebSocketAlerts)
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	h.logger.Debug("Health check requested")
	
	response := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "object-detection-service",
		"version":   "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessCheck handles readiness check requests
func (h *Handler) ReadinessCheck(c *gin.Context) {
	h.logger.Debug("Readiness check requested")
	
	// TODO: Add actual readiness checks (database, redis, model loading, etc.)
	ready := true
	status := http.StatusOK
	
	if !ready {
		status = http.StatusServiceUnavailable
	}

	response := gin.H{
		"status":    map[bool]string{true: "ready", false: "not ready"}[ready],
		"timestamp": time.Now().UTC(),
		"service":   "object-detection-service",
		"checks": gin.H{
			"database": "ok", // TODO: Implement actual checks
			"redis":    "ok",
			"model":    "ok",
		},
	}

	c.JSON(status, response)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// sendError sends an error response
func (h *Handler) sendError(c *gin.Context, statusCode int, err string, message string) {
	h.logger.Error("HTTP error response", 
		zap.Int("status_code", statusCode),
		zap.String("error", err),
		zap.String("message", message))

	c.JSON(statusCode, ErrorResponse{
		Error:   err,
		Message: message,
	})
}

// sendSuccess sends a success response
func (h *Handler) sendSuccess(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}
