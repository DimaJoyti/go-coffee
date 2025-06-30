package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SimpleDashboardServer provides HTTP endpoints for the dashboard using standard library
type SimpleDashboardServer struct {
	dashboard *RealTimeDashboard
	config    *DashboardConfig
	logger    Logger
	server    *http.Server
	mux       *http.ServeMux
}

// NewSimpleDashboardServer creates a new simple dashboard server
func NewSimpleDashboardServer(dashboard *RealTimeDashboard, config *DashboardConfig, logger Logger) *SimpleDashboardServer {
	return &SimpleDashboardServer{
		dashboard: dashboard,
		config:    config,
		logger:    logger,
		mux:       http.NewServeMux(),
	}
}

// Start starts the simple dashboard server
func (sds *SimpleDashboardServer) Start(ctx context.Context) error {
	sds.setupRoutes()
	
	sds.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", sds.config.Port),
		Handler: sds.mux,
	}

	go func() {
		sds.logger.Info("Simple dashboard server starting", "port", sds.config.Port)
		if err := sds.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sds.logger.Error("Simple dashboard server error", err)
		}
	}()

	return nil
}

// Stop stops the simple dashboard server
func (sds *SimpleDashboardServer) Stop(ctx context.Context) error {
	if sds.server != nil {
		return sds.server.Shutdown(ctx)
	}
	return nil
}

// setupRoutes sets up HTTP routes using standard library
func (sds *SimpleDashboardServer) setupRoutes() {
	// API routes
	if sds.config.EnableHTTPAPI {
		sds.mux.HandleFunc("/api/v1/metrics", sds.withMiddleware(sds.handleGetMetrics))
		sds.mux.HandleFunc("/api/v1/metrics/current", sds.withMiddleware(sds.handleGetCurrentMetrics))
		sds.mux.HandleFunc("/api/v1/events", sds.withMiddleware(sds.handleGetEvents))
		sds.mux.HandleFunc("/api/v1/events/recent", sds.withMiddleware(sds.handleGetRecentEvents))
		sds.mux.HandleFunc("/api/v1/dashboard/stats", sds.withMiddleware(sds.handleGetDashboardStats))
		sds.mux.HandleFunc("/api/v1/dashboard/health", sds.withMiddleware(sds.handleGetDashboardHealth))
		sds.mux.HandleFunc("/api/v1/clients", sds.withMiddleware(sds.handleGetClients))
		sds.mux.HandleFunc("/api/v1/stream/stats", sds.withMiddleware(sds.handleGetStreamStats))
	}

	// Static files
	if sds.config.EnableStaticFiles {
		sds.mux.Handle("/", http.FileServer(http.Dir(sds.config.StaticFilesPath)))
	}
}

// withMiddleware applies middleware to handlers
func (sds *SimpleDashboardServer) withMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apply CORS middleware if enabled
		if sds.config.CORSEnabled {
			sds.applyCORS(w, r)
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		// Apply authentication middleware if enabled
		if sds.config.AuthEnabled && !sds.isPublicEndpoint(r.URL.Path) {
			if !sds.authenticate(r) {
				sds.writeJSONResponse(w, http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"message": "Invalid or missing authentication token",
				})
				return
			}
		}

		handler(w, r)
	}
}

// HTTP Handlers

// handleGetMetrics returns current dashboard metrics
func (sds *SimpleDashboardServer) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sds.writeJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	metrics := sds.dashboard.GetCurrentMetrics()
	sds.writeJSONResponse(w, http.StatusOK, metrics)
}

// handleGetCurrentMetrics returns current metrics (alias for compatibility)
func (sds *SimpleDashboardServer) handleGetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	sds.handleGetMetrics(w, r)
}

// handleGetEvents returns recent events with optional filtering
func (sds *SimpleDashboardServer) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sds.writeJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	limit := sds.getIntQueryParam(r, "limit", 100)
	events := sds.dashboard.GetRecentEvents(limit)
	sds.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

// handleGetRecentEvents returns recent events (alias for compatibility)
func (sds *SimpleDashboardServer) handleGetRecentEvents(w http.ResponseWriter, r *http.Request) {
	sds.handleGetEvents(w, r)
}

// handleGetDashboardStats returns dashboard statistics
func (sds *SimpleDashboardServer) handleGetDashboardStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sds.writeJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	stats := sds.dashboard.GetDashboardStats()
	sds.writeJSONResponse(w, http.StatusOK, stats)
}

// handleGetDashboardHealth returns dashboard health status
func (sds *SimpleDashboardServer) handleGetDashboardHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sds.writeJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	ctx := r.Context()
	err := sds.dashboard.Health(ctx)

	health := map[string]interface{}{
		"healthy":   err == nil,
		"timestamp": time.Now(),
	}

	if err != nil {
		health["error"] = err.Error()
		sds.writeJSONResponse(w, http.StatusServiceUnavailable, health)
	} else {
		sds.writeJSONResponse(w, http.StatusOK, health)
	}
}

// handleGetClients returns connected WebSocket clients
func (sds *SimpleDashboardServer) handleGetClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sds.writeJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	clients := sds.dashboard.GetConnectedClients()
	sds.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"clients": clients,
		"count":   len(clients),
	})
}

// handleGetStreamStats returns event streaming statistics
func (sds *SimpleDashboardServer) handleGetStreamStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sds.writeJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	stats := sds.dashboard.eventStreamer.GetStats()
	sds.writeJSONResponse(w, http.StatusOK, stats)
}

// Middleware helpers

// applyCORS applies CORS headers
func (sds *SimpleDashboardServer) applyCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range sds.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}

	if allowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// authenticate checks authentication
func (sds *SimpleDashboardServer) authenticate(r *http.Request) bool {
	token := r.Header.Get("Authorization")
	if token == "" {
		token = r.URL.Query().Get("token")
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	return token == sds.config.AuthToken
}

// isPublicEndpoint checks if an endpoint is public
func (sds *SimpleDashboardServer) isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/api/v1/dashboard/health",
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}

	return false
}

// Helper methods

// writeJSONResponse writes a JSON response
func (sds *SimpleDashboardServer) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		sds.logger.Error("Failed to encode JSON response", err)
	}
}

// getIntQueryParam gets an integer query parameter with default value
func (sds *SimpleDashboardServer) getIntQueryParam(r *http.Request, param string, defaultValue int) int {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}

	return defaultValue
}

// getClientIP extracts client IP address
func (sds *SimpleDashboardServer) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}

// Health checks the health of the dashboard server
func (sds *SimpleDashboardServer) Health(ctx context.Context) error {
	if sds.server == nil {
		return fmt.Errorf("dashboard server not initialized")
	}

	// Check if dashboard is healthy
	return sds.dashboard.Health(ctx)
}

// GetServerInfo returns server information
func (sds *SimpleDashboardServer) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"server_type":    "simple_dashboard_server",
		"port":           sds.config.Port,
		"cors_enabled":   sds.config.CORSEnabled,
		"auth_enabled":   sds.config.AuthEnabled,
		"api_enabled":    sds.config.EnableHTTPAPI,
		"static_enabled": sds.config.EnableStaticFiles,
		"uptime":         time.Since(time.Now()), // This would be tracked properly in a real implementation
	}
}
