package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DimaJoyti/go-coffee/internal/security-gateway/application"
	"github.com/DimaJoyti/go-coffee/pkg/security/monitoring"
	"github.com/DimaJoyti/go-coffee/pkg/security/validation"
)

// ValidateHandler handles input validation requests
func ValidateHandler(validationService *validation.ValidationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ValidationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		var result *validation.ValidationResult

		switch req.Type {
		case "email":
			result = validationService.ValidateEmail(req.Value)
		case "password":
			result = validationService.ValidatePassword(req.Value)
		case "url":
			result = validationService.ValidateURL(req.Value)
		case "ip":
			result = validationService.ValidateIP(req.Value)
		case "input":
			result = validationService.ValidateInput(req.Value)
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Invalid validation type",
			})
			return
		}

		c.JSON(http.StatusOK, ValidationResponse{
			Valid:          result.IsValid,
			Errors:         result.Errors,
			Warnings:       result.Warnings,
			SanitizedValue: result.SanitizedValue,
			ThreatLevel:    result.ThreatLevel,
		})
	}
}

// SecurityMetricsHandler returns security metrics
func SecurityMetricsHandler(monitoringService *monitoring.SecurityMonitoringService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		metrics := monitoringService.GetSecurityMetrics(ctx)
		
		c.JSON(http.StatusOK, SecurityMetricsResponse{
			TotalEvents:         metrics.TotalEvents,
			BlockedRequests:     metrics.BlockedRequests,
			AllowedRequests:     metrics.AllowedRequests,
			ThreatDetections:    metrics.ThreatDetections,
			RateLimitViolations: metrics.RateLimitViolations,
			WAFBlocks:           metrics.WAFBlocks,
			RequestsByCountry:   metrics.RequestsByCountry,
			AverageResponseTime: metrics.AverageResponseTime.String(),
			LastUpdated:         metrics.LastUpdated,
		})
	}
}

// AlertsHandler returns security alerts
func AlertsHandler(monitoringService *monitoring.SecurityMonitoringService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		// Parse query parameters
		limitStr := c.DefaultQuery("limit", "50")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 1000 {
			limit = 50
		}

		statusFilter := c.Query("status")
		severityFilter := c.Query("severity")

		// Build filter
		filter := monitoring.EventFilter{
			Limit: limit,
		}

		// Add time range if specified
		if startTime := c.Query("start_time"); startTime != "" {
			if t, err := time.Parse(time.RFC3339, startTime); err == nil {
				filter.StartTime = &t
			}
		}

		if endTime := c.Query("end_time"); endTime != "" {
			if t, err := time.Parse(time.RFC3339, endTime); err == nil {
				filter.EndTime = &t
			}
		}

		// Query events (alerts are stored as events)
		events, err := monitoringService.QueryEvents(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to query alerts",
			})
			return
		}

		// Convert events to alert format
		alerts := make([]AlertResponse, 0, len(events))
		for _, event := range events {
			// Filter by status and severity if specified
			if statusFilter != "" && string(event.Severity) != statusFilter {
				continue
			}
			if severityFilter != "" && string(event.Severity) != severityFilter {
				continue
			}

			alerts = append(alerts, AlertResponse{
				ID:          event.ID,
				Type:        string(event.EventType),
				Severity:    string(event.Severity),
				Title:       event.Description,
				Description: event.Description,
				Source:      event.Source,
				IPAddress:   event.IPAddress,
				UserID:      event.UserID,
				Timestamp:   event.Timestamp,
				Status:      "open", // Default status
				Metadata:    event.Metadata,
			})
		}

		c.JSON(http.StatusOK, AlertsResponse{
			Alerts: alerts,
			Total:  len(alerts),
		})
	}
}

// MetricsHandler returns Prometheus-style metrics
func MetricsHandler(monitoringService *monitoring.SecurityMonitoringService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		metrics := monitoringService.GetSecurityMetrics(ctx)
		
		// Generate Prometheus-style metrics
		prometheusMetrics := generatePrometheusMetrics(metrics)
		
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.String(http.StatusOK, prometheusMetrics)
	}
}

// ProxyHandler proxies requests to backend services
func ProxyHandler(serviceName string, gatewayService *application.SecurityGatewayService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Remove the gateway prefix from the path
		originalPath := c.Request.URL.Path
		c.Request.URL.Path = c.Param("path")
		
		// Proxy the request
		err := gatewayService.ProxyRequest(serviceName, c.Writer, c.Request)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "Bad Gateway",
				"message": "Failed to proxy request to backend service",
				"service": serviceName,
			})
			return
		}
		
		// Restore original path
		c.Request.URL.Path = originalPath
	}
}

// HealthHandler returns health status
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "security-gateway",
			"version":   "1.0.0",
		})
	}
}

// Request/Response types

type ValidationRequest struct {
	Type  string `json:"type" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type ValidationResponse struct {
	Valid          bool     `json:"valid"`
	Errors         []string `json:"errors,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	SanitizedValue string   `json:"sanitized_value,omitempty"`
	ThreatLevel    string   `json:"threat_level,omitempty"`
}

type SecurityMetricsResponse struct {
	TotalEvents         int64             `json:"total_events"`
	BlockedRequests     int64             `json:"blocked_requests"`
	AllowedRequests     int64             `json:"allowed_requests"`
	ThreatDetections    int64             `json:"threat_detections"`
	RateLimitViolations int64             `json:"rate_limit_violations"`
	WAFBlocks           int64             `json:"waf_blocks"`
	RequestsByCountry   map[string]int64  `json:"requests_by_country"`
	AverageResponseTime string            `json:"average_response_time"`
	LastUpdated         time.Time         `json:"last_updated"`
}

type AlertResponse struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type AlertsResponse struct {
	Alerts []AlertResponse `json:"alerts"`
	Total  int             `json:"total"`
}

// Helper functions

func generatePrometheusMetrics(metrics *monitoring.SecurityMetrics) string {
	prometheusMetrics := `# HELP security_gateway_total_events Total number of security events
# TYPE security_gateway_total_events counter
security_gateway_total_events ` + strconv.FormatInt(metrics.TotalEvents, 10) + `

# HELP security_gateway_blocked_requests Total number of blocked requests
# TYPE security_gateway_blocked_requests counter
security_gateway_blocked_requests ` + strconv.FormatInt(metrics.BlockedRequests, 10) + `

# HELP security_gateway_allowed_requests Total number of allowed requests
# TYPE security_gateway_allowed_requests counter
security_gateway_allowed_requests ` + strconv.FormatInt(metrics.AllowedRequests, 10) + `

# HELP security_gateway_threat_detections Total number of threat detections
# TYPE security_gateway_threat_detections counter
security_gateway_threat_detections ` + strconv.FormatInt(metrics.ThreatDetections, 10) + `

# HELP security_gateway_rate_limit_violations Total number of rate limit violations
# TYPE security_gateway_rate_limit_violations counter
security_gateway_rate_limit_violations ` + strconv.FormatInt(metrics.RateLimitViolations, 10) + `

# HELP security_gateway_waf_blocks Total number of WAF blocks
# TYPE security_gateway_waf_blocks counter
security_gateway_waf_blocks ` + strconv.FormatInt(metrics.WAFBlocks, 10) + `

# HELP security_gateway_average_response_time Average response time in milliseconds
# TYPE security_gateway_average_response_time gauge
security_gateway_average_response_time ` + strconv.FormatFloat(float64(metrics.AverageResponseTime.Milliseconds()), 'f', 2, 64) + `

`

	// Add requests by country metrics
	for country, count := range metrics.RequestsByCountry {
		prometheusMetrics += `# HELP security_gateway_requests_by_country Requests by country
# TYPE security_gateway_requests_by_country counter
security_gateway_requests_by_country{country="` + country + `"} ` + strconv.FormatInt(count, 10) + `

`
	}

	return prometheusMetrics
}
