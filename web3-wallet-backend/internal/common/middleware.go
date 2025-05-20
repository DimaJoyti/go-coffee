package common

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/web3-wallet-backend/pkg/config"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// LoggerMiddleware returns a middleware that logs HTTP requests
func LoggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Get request ID
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = "unknown"
		}

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Get error if any
		var errorMessage string
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}

		// Log request
		log.Info("HTTP Request",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("error", errorMessage),
		)
	}
}

// RequestIDMiddleware returns a middleware that adds a request ID to the context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = uuid.New().String()
		}

		// Set request ID in context
		c.Set("request_id", requestID)

		// Set request ID in response header
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

// CORSMiddleware returns a middleware that handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware returns a middleware that limits request rate
func RateLimitMiddleware(cfg config.RateLimitConfig) gin.HandlerFunc {
	// Create limiter
	limiter := rate.NewLimiter(rate.Limit(cfg.RequestsPerMinute/60), cfg.Burst)

	return func(c *gin.Context) {
		// Skip if rate limiting is disabled
		if !cfg.Enabled {
			c.Next()
			return
		}

		// Check if request is allowed
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			return
		}

		c.Next()
	}
}

// AuthMiddleware returns a middleware that checks for a valid JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Remove "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// TODO: Validate token
		// For now, just set a dummy user ID in context
		c.Set("user_id", "user123")

		c.Next()
	}
}

// ErrorMiddleware returns a middleware that handles errors
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Check if it's an API error
			if apiErr, ok := err.Err.(*APIError); ok {
				c.AbortWithStatusJSON(apiErr.StatusCode, gin.H{
					"error": apiErr.Message,
					"code":  apiErr.Code,
				})
				return
			}

			// Default to internal server error
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			return
		}
	}
}

// APIError represents an API error
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

// Error returns the error message
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, code, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// BadRequestError creates a new bad request error
func BadRequestError(message string) *APIError {
	return NewAPIError(http.StatusBadRequest, "bad_request", message)
}

// UnauthorizedError creates a new unauthorized error
func UnauthorizedError(message string) *APIError {
	return NewAPIError(http.StatusUnauthorized, "unauthorized", message)
}

// ForbiddenError creates a new forbidden error
func ForbiddenError(message string) *APIError {
	return NewAPIError(http.StatusForbidden, "forbidden", message)
}

// NotFoundError creates a new not found error
func NotFoundError(message string) *APIError {
	return NewAPIError(http.StatusNotFound, "not_found", message)
}

// ConflictError creates a new conflict error
func ConflictError(message string) *APIError {
	return NewAPIError(http.StatusConflict, "conflict", message)
}

// InternalServerError creates a new internal server error
func InternalServerError(message string) *APIError {
	return NewAPIError(http.StatusInternalServerError, "internal_server_error", message)
}
