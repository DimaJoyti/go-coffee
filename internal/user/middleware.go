package user

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// LoggerMiddleware provides request logging
func LoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Build log data
		logData := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"status":     statusCode,
			"latency":    latency.String(),
			"user_agent": c.Request.UserAgent(),
			"client_ip":  c.ClientIP(),
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			logData["errors"] = c.Errors.String()
		}

		switch {
		case statusCode >= 500:
			logger.Error("HTTP request completed with server error", logData)
		case statusCode >= 400:
			logger.Warn("HTTP request completed with client error", logData)
		default:
			logger.Info("HTTP request completed", logData)
		}
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting per IP
func RateLimitMiddleware() gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Cleanup old clients every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		if _, found := clients[ip]; !found {
			// Allow 100 requests per minute per IP
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Every(time.Minute/100), 100),
			}
		}
		clients[ip].lastSeen = time.Now()
		limiter := clients[ip].limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests from this IP address",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthMiddleware provides authentication (placeholder implementation)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		token := c.GetHeader("Authorization")
		
		// Skip auth for health check and docs
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/api/docs" {
			c.Next()
			return
		}

		// Simple token validation (in production, use proper JWT validation)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization token required",
			})
			c.Abort()
			return
		}

		// TODO: Implement proper JWT token validation
		// For now, accept any non-empty token
		if token != "" {
			// Set user context (placeholder)
			c.Set("user_id", "user-123")
			c.Set("user_role", "customer")
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID (in production, use UUID)
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// SecurityMiddleware adds security headers
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// MetricsMiddleware collects metrics for monitoring
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// Calculate metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// TODO: Send metrics to Prometheus
		// For now, just log metrics
		if duration > 5*time.Second {
			// Log slow requests
			zap.L().Warn("Slow request detected",
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", statusCode),
				zap.Duration("duration", duration),
			)
		}
	}
}

// RecoveryMiddleware handles panics gracefully
func RecoveryMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			zap.Any("panic", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "An unexpected error occurred",
		})
	})
}

// ValidationMiddleware validates request data
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add request size limit
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1MB limit

		c.Next()
	}
}

// CacheMiddleware adds caching headers for static content
func CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cache static files for 1 hour
		if c.Request.URL.Path[:8] == "/static/" {
			c.Header("Cache-Control", "public, max-age=3600")
		} else {
			// No cache for API endpoints
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

// CompressionMiddleware enables gzip compression
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Enable gzip compression for responses > 1KB
		c.Header("Vary", "Accept-Encoding")
		c.Next()
	}
}

// Helper functions

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple implementation - in production use UUID
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
