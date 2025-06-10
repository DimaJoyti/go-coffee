package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/DimaJoyti/go-coffee/internal/security-gateway/application"
	"github.com/DimaJoyti/go-coffee/pkg/security/monitoring"
	"github.com/DimaJoyti/go-coffee/pkg/security/validation"
)

// Helper functions for clean HTTP handlers

// respondWithJSON writes a JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondWithError writes an error response
func respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	response := map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": message,
	}
	if err != nil {
		response["details"] = err.Error()
	}
	respondWithJSON(w, statusCode, response)
}

// decodeJSON decodes JSON from request body
func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// getPathParam gets path parameter from mux
func getPathParam(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}

// getQueryParam gets query parameter
func getQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// getQueryParamWithDefault gets query parameter with default value
func getQueryParamWithDefault(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// ValidateHandler handles input validation requests (Clean HTTP Handler)
func ValidateHandler(validationService *validation.ValidationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ValidationRequest
		if err := decodeJSON(r, &req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request format", err)
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
			respondWithError(w, http.StatusBadRequest, "Invalid validation type", nil)
			return
		}

		response := ValidationResponse{
			Valid:          result.IsValid,
			Errors:         result.Errors,
			Warnings:       result.Warnings,
			SanitizedValue: result.SanitizedValue,
			ThreatLevel:    result.ThreatLevel,
		}

		respondWithJSON(w, http.StatusOK, response)
	}
}

// SecurityMetricsHandler returns security metrics (Clean HTTP Handler)
func SecurityMetricsHandler(monitoringService *monitoring.SecurityMonitoringService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		metrics := monitoringService.GetSecurityMetrics(ctx)

		response := SecurityMetricsResponse{
			TotalEvents:         metrics.TotalEvents,
			BlockedRequests:     metrics.BlockedRequests,
			AllowedRequests:     metrics.TotalEvents - metrics.BlockedRequests, // Calculate allowed requests
			ThreatDetections:    metrics.ThreatDetections,
			RateLimitViolations: 0,                      // Not available in current metrics
			WAFBlocks:           0,                      // Not available in current metrics
			RequestsByCountry:   make(map[string]int64), // Not available in current metrics
			AverageResponseTime: "0ms",                  // Not available in current metrics
			LastUpdated:         metrics.LastUpdated,
		}

		respondWithJSON(w, http.StatusOK, response)
	}
}

// AlertsHandler returns security alerts (Clean HTTP Handler)
func AlertsHandler(monitoringService *monitoring.SecurityMonitoringService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Parse query parameters
		limitStr := getQueryParamWithDefault(r, "limit", "50")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 1000 {
			limit = 50
		}

		statusFilter := getQueryParam(r, "status")
		severityFilter := getQueryParam(r, "severity")

		// Build filter
		filter := monitoring.EventFilter{
			Limit: limit,
		}

		// Add time range if specified
		if startTime := getQueryParam(r, "start_time"); startTime != "" {
			if t, err := time.Parse(time.RFC3339, startTime); err == nil {
				filter.StartTime = &t
			}
		}

		if endTime := getQueryParam(r, "end_time"); endTime != "" {
			if t, err := time.Parse(time.RFC3339, endTime); err == nil {
				filter.EndTime = &t
			}
		}

		// Query events (alerts are stored as events)
		events, err := monitoringService.QueryEvents(ctx, filter)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to query alerts", err)
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

		response := AlertsResponse{
			Alerts: alerts,
			Total:  len(alerts),
		}

		respondWithJSON(w, http.StatusOK, response)
	}
}

// MetricsHandler returns Prometheus-style metrics (Clean HTTP Handler)
func MetricsHandler(monitoringService *monitoring.SecurityMonitoringService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		metrics := monitoringService.GetSecurityMetrics(ctx)

		// Generate Prometheus-style metrics
		prometheusMetrics := generatePrometheusMetrics(metrics)

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(prometheusMetrics))
	}
}

// ProxyHandler proxies requests to backend services (Clean HTTP Handler)
func ProxyHandler(serviceName string, gatewayService *application.SecurityGatewayService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Remove the gateway prefix from the path
		originalPath := r.URL.Path
		pathParam := getPathParam(r, "path")
		if pathParam != "" {
			r.URL.Path = pathParam
		}

		// Proxy the request
		err := gatewayService.ProxyRequest(serviceName, w, r)
		if err != nil {
			response := map[string]interface{}{
				"error":   "Bad Gateway",
				"message": "Failed to proxy request to backend service",
				"service": serviceName,
			}
			respondWithJSON(w, http.StatusBadGateway, response)
			return
		}

		// Restore original path
		r.URL.Path = originalPath
	}
}

// HealthHandler returns health status (Clean HTTP Handler)
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "security-gateway",
			"version":   "1.0.0",
		}
		respondWithJSON(w, http.StatusOK, response)
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
	TotalEvents         int64            `json:"total_events"`
	BlockedRequests     int64            `json:"blocked_requests"`
	AllowedRequests     int64            `json:"allowed_requests"`
	ThreatDetections    int64            `json:"threat_detections"`
	RateLimitViolations int64            `json:"rate_limit_violations"`
	WAFBlocks           int64            `json:"waf_blocks"`
	RequestsByCountry   map[string]int64 `json:"requests_by_country"`
	AverageResponseTime string           `json:"average_response_time"`
	LastUpdated         time.Time        `json:"last_updated"`
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
security_gateway_allowed_requests ` + strconv.FormatInt(metrics.TotalEvents-metrics.BlockedRequests, 10) + `

# HELP security_gateway_threat_detections Total number of threat detections
# TYPE security_gateway_threat_detections counter
security_gateway_threat_detections ` + strconv.FormatInt(metrics.ThreatDetections, 10) + `

# HELP security_gateway_active_alerts Total number of active alerts
# TYPE security_gateway_active_alerts gauge
security_gateway_active_alerts ` + strconv.FormatInt(metrics.ActiveAlerts, 10) + `

# HELP security_gateway_resolved_alerts Total number of resolved alerts
# TYPE security_gateway_resolved_alerts counter
security_gateway_resolved_alerts ` + strconv.FormatInt(metrics.ResolvedAlerts, 10) + `

`

	// Add events by type metrics
	for eventType, count := range metrics.EventsByType {
		prometheusMetrics += `# HELP security_gateway_events_by_type Events by type
# TYPE security_gateway_events_by_type counter
security_gateway_events_by_type{type="` + eventType + `"} ` + strconv.FormatInt(count, 10) + `

`
	}

	// Add events by severity metrics
	for severity, count := range metrics.EventsBySeverity {
		prometheusMetrics += `# HELP security_gateway_events_by_severity Events by severity
# TYPE security_gateway_events_by_severity counter
security_gateway_events_by_severity{severity="` + severity + `"} ` + strconv.FormatInt(count, 10) + `

`
	}

	return prometheusMetrics
}
