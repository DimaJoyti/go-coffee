package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DimaJoyti/go-coffee/internal/security-gateway/application"
	"github.com/DimaJoyti/go-coffee/internal/security-gateway/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/security/monitoring"
)

// LoggingMiddleware provides request/response logging
func LoggingMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get status code
		statusCode := c.Writer.Status()

		// Get request size
		requestSize := c.Request.ContentLength

		// Get response size
		responseSize := c.Writer.Size()

		// Build log fields
		fields := map[string]any{
			"status_code":    statusCode,
			"latency":        latency,
			"client_ip":      clientIP,
			"method":         c.Request.Method,
			"path":           path,
			"request_size":   requestSize,
			"response_size":  responseSize,
			"user_agent":     c.Request.UserAgent(),
			"request_id":     c.GetHeader("X-Request-ID"),
			"correlation_id": c.GetHeader("X-Correlation-ID"),
		}

		if raw != "" {
			fields["query"] = raw
		}

		// Log based on status code
		switch {
		case statusCode >= 500:
			logger.Error("Server error", fields)
		case statusCode >= 400:
			logger.Warn("Client error", fields)
		case statusCode >= 300:
			logger.Info("Redirect", fields)
		default:
			logger.Info("Request completed", fields)
		}
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		if len(config.AllowedOrigins) > 0 && !contains(config.AllowedOrigins, "*") {
			if !contains(config.AllowedOrigins, origin) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// Set CORS headers
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if contains(config.AllowedOrigins, "*") {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", joinStrings(config.ExposedHeaders, ", "))
		c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self' https:; frame-ancestors 'none';")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		// Remove server information
		c.Header("Server", "")
		
		c.Next()
	}
}

// RateLimitMiddleware applies rate limiting
func RateLimitMiddleware(rateLimitService *application.RateLimitService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		// Create security request
		securityRequest := domain.NewSecurityRequest(c.Request)
		
		// Check multiple rate limits
		allowed, infos, err := rateLimitService.CheckMultipleRateLimits(ctx, securityRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Rate limit check failed",
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		if len(infos) > 0 {
			// Use the most restrictive rate limit info
			info := infos[0]
			for _, i := range infos {
				if i.Remaining < info.Remaining {
					info = i
				}
			}
			
			c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded",
				"retry_after": "60",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// WAFMiddleware applies Web Application Firewall rules
func WAFMiddleware(wafService *application.WAFService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		// Read request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		
		// Create security request
		securityRequest := domain.NewSecurityRequest(c.Request)
		securityRequest.Body = bodyBytes
		
		// Check WAF rules
		wafResult, err := wafService.CheckRequest(ctx, securityRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "WAF check failed",
			})
			c.Abort()
			return
		}

		// Add WAF headers
		c.Header("X-WAF-Score", strconv.FormatFloat(wafResult.Score, 'f', 2, 64))
		
		if wafResult.Blocked {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": wafResult.Reason,
				"rule":    wafResult.RuleMatched,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GatewayMiddleware processes requests through the security gateway
func GatewayMiddleware(gatewayService *application.SecurityGatewayService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		// Process request through security gateway
		response, err := gatewayService.ProcessRequest(ctx, c.Request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Gateway processing failed",
			})
			c.Abort()
			return
		}

		// Add security headers from response
		for name, value := range response.Headers {
			c.Header(name, value)
		}

		// Check if request was blocked
		if response.StatusCode == http.StatusForbidden {
			c.JSON(response.StatusCode, gin.H{
				"error":   "Forbidden",
				"message": "Request blocked by security gateway",
				"request_id": response.RequestID,
			})
			c.Abort()
			return
		}

		// Add security check results to context
		c.Set("security_checks", response.SecurityChecks)
		c.Set("security_response", response)

		c.Next()
	}
}

// RequestIDMiddleware adds request ID to requests
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Request.Header.Set("X-Request-ID", requestID)
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		
		c.Next()
	}
}

// CorrelationIDMiddleware adds correlation ID to requests
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = generateRequestID()
			c.Request.Header.Set("X-Correlation-ID", correlationID)
		}
		
		c.Header("X-Correlation-ID", correlationID)
		c.Set("correlation_id", correlationID)
		
		c.Next()
	}
}

// TenantMiddleware extracts tenant information
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID != "" {
			c.Set("tenant_id", tenantID)
		}
		
		c.Next()
	}
}

// MetricsMiddleware collects metrics
func MetricsMiddleware(monitoringService *monitoring.SecurityMonitoringService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		// Collect metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		
		// Log security event for metrics
		event := &monitoring.SecurityEvent{
			EventType:   monitoring.EventTypeNetworkActivity,
			Severity:    monitoring.SeverityInfo,
			Source:      "security-gateway",
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Description: "HTTP request processed",
			Metadata: map[string]interface{}{
				"method":        c.Request.Method,
				"path":          c.Request.URL.Path,
				"status_code":   statusCode,
				"duration_ms":   duration.Milliseconds(),
				"request_size":  c.Request.ContentLength,
				"response_size": c.Writer.Size(),
			},
		}
		
		ctx := context.Background()
		monitoringService.LogSecurityEvent(ctx, event)
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func joinStrings(slice []string, separator string) string {
	if len(slice) == 0 {
		return ""
	}
	
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += separator + slice[i]
	}
	return result
}

func generateRequestID() string {
	return "req_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
