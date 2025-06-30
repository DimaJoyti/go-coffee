package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// DashboardServer provides HTTP and WebSocket endpoints for the dashboard
type DashboardServer struct {
	dashboard *RealTimeDashboard
	config    *DashboardConfig
	logger    Logger
	server    *http.Server
	upgrader  websocket.Upgrader
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(dashboard *RealTimeDashboard, config *DashboardConfig, logger Logger) *DashboardServer {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if !config.CORSEnabled {
				return true
			}

			origin := r.Header.Get("Origin")
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					return true
				}
			}
			return false
		},
	}

	return &DashboardServer{
		dashboard: dashboard,
		config:    config,
		logger:    logger,
		upgrader:  upgrader,
	}
}

// Start starts the dashboard server
func (ds *DashboardServer) Start(ctx context.Context) error {
	router := ds.setupRoutes()

	ds.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ds.config.Port),
		Handler: router,
	}

	go func() {
		ds.logger.Info("Dashboard server starting", "port", ds.config.Port)
		if err := ds.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ds.logger.Error("Dashboard server error", err)
		}
	}()

	return nil
}

// Stop stops the dashboard server
func (ds *DashboardServer) Stop(ctx context.Context) error {
	if ds.server != nil {
		return ds.server.Shutdown(ctx)
	}
	return nil
}

// setupRoutes sets up HTTP routes
func (ds *DashboardServer) setupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply CORS middleware if enabled
	if ds.config.CORSEnabled {
		router.Use(ds.corsMiddleware)
	}

	// Apply authentication middleware if enabled
	if ds.config.AuthEnabled {
		router.Use(ds.authMiddleware)
	}

	// API routes
	if ds.config.EnableHTTPAPI {
		api := router.PathPrefix("/api/v1").Subrouter()
		ds.setupAPIRoutes(api)
	}

	// WebSocket endpoint
	if ds.config.EnableWebSocket {
		router.HandleFunc("/ws", ds.handleWebSocket)
	}

	// Static files
	if ds.config.EnableStaticFiles {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir(ds.config.StaticFilesPath)))
	}

	return router
}

// setupAPIRoutes sets up API routes
func (ds *DashboardServer) setupAPIRoutes(router *mux.Router) {
	// Metrics endpoints
	router.HandleFunc("/metrics", ds.handleGetMetrics).Methods("GET")
	router.HandleFunc("/metrics/current", ds.handleGetCurrentMetrics).Methods("GET")

	// Events endpoints
	router.HandleFunc("/events", ds.handleGetEvents).Methods("GET")
	router.HandleFunc("/events/recent", ds.handleGetRecentEvents).Methods("GET")

	// Dashboard endpoints
	router.HandleFunc("/dashboard/stats", ds.handleGetDashboardStats).Methods("GET")
	router.HandleFunc("/dashboard/health", ds.handleGetDashboardHealth).Methods("GET")

	// Client management endpoints
	router.HandleFunc("/clients", ds.handleGetClients).Methods("GET")
	router.HandleFunc("/clients/{id}", ds.handleDisconnectClient).Methods("DELETE")

	// Event streaming endpoints
	router.HandleFunc("/stream/subscribers", ds.handleGetSubscribers).Methods("GET")
	router.HandleFunc("/stream/filters", ds.handleGetFilters).Methods("GET")
	router.HandleFunc("/stream/processors", ds.handleGetProcessors).Methods("GET")
	router.HandleFunc("/stream/stats", ds.handleGetStreamStats).Methods("GET")
}

// HTTP Handlers

// handleGetMetrics returns current dashboard metrics
func (ds *DashboardServer) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := ds.dashboard.GetCurrentMetrics()
	ds.writeJSONResponse(w, http.StatusOK, metrics)
}

// handleGetCurrentMetrics returns current metrics (alias for compatibility)
func (ds *DashboardServer) handleGetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	ds.handleGetMetrics(w, r)
}

// handleGetEvents returns recent events with optional filtering
func (ds *DashboardServer) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	limit := ds.getIntQueryParam(r, "limit", 100)
	events := ds.dashboard.GetRecentEvents(limit)
	ds.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

// handleGetRecentEvents returns recent events (alias for compatibility)
func (ds *DashboardServer) handleGetRecentEvents(w http.ResponseWriter, r *http.Request) {
	ds.handleGetEvents(w, r)
}

// handleGetDashboardStats returns dashboard statistics
func (ds *DashboardServer) handleGetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats := ds.dashboard.GetDashboardStats()
	ds.writeJSONResponse(w, http.StatusOK, stats)
}

// handleGetDashboardHealth returns dashboard health status
func (ds *DashboardServer) handleGetDashboardHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := ds.dashboard.Health(ctx)

	health := map[string]interface{}{
		"healthy":   err == nil,
		"timestamp": time.Now(),
	}

	if err != nil {
		health["error"] = err.Error()
		ds.writeJSONResponse(w, http.StatusServiceUnavailable, health)
	} else {
		ds.writeJSONResponse(w, http.StatusOK, health)
	}
}

// handleGetClients returns connected WebSocket clients
func (ds *DashboardServer) handleGetClients(w http.ResponseWriter, r *http.Request) {
	clients := ds.dashboard.GetConnectedClients()
	ds.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"clients": clients,
		"count":   len(clients),
	})
}

// handleDisconnectClient disconnects a specific WebSocket client
func (ds *DashboardServer) handleDisconnectClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	ds.dashboard.RemoveWebSocketClient(clientID)
	ds.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":   "Client disconnected",
		"client_id": clientID,
	})
}

// handleGetSubscribers returns event stream subscribers
func (ds *DashboardServer) handleGetSubscribers(w http.ResponseWriter, r *http.Request) {
	subscribers := ds.dashboard.eventStreamer.GetSubscribers()
	ds.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"subscribers": subscribers,
		"count":       len(subscribers),
	})
}

// handleGetFilters returns event filters
func (ds *DashboardServer) handleGetFilters(w http.ResponseWriter, r *http.Request) {
	filters := ds.dashboard.eventStreamer.GetEventFilters()
	ds.writeJSONResponse(w, http.StatusOK, filters)
}

// handleGetProcessors returns event processors
func (ds *DashboardServer) handleGetProcessors(w http.ResponseWriter, r *http.Request) {
	processors := ds.dashboard.eventStreamer.GetEventProcessors()
	ds.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"processors": processors,
		"count":      len(processors),
	})
}

// handleGetStreamStats returns event streaming statistics
func (ds *DashboardServer) handleGetStreamStats(w http.ResponseWriter, r *http.Request) {
	stats := ds.dashboard.eventStreamer.GetStats()
	ds.writeJSONResponse(w, http.StatusOK, stats)
}

// WebSocket Handler

// handleWebSocket handles WebSocket connections
func (ds *DashboardServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ds.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ds.logger.Error("WebSocket upgrade failed", err)
		return
	}

	// Create client
	client := &WebSocketClient{
		ID:            fmt.Sprintf("client_%d", time.Now().UnixNano()),
		Connection:    conn,
		UserID:        ds.getUserID(r),
		IPAddress:     ds.getClientIP(r),
		UserAgent:     r.UserAgent(),
		ConnectedAt:   time.Now(),
		LastSeen:      time.Now(),
		Subscriptions: []string{"metrics", "events"},
	}

	// Add client to dashboard
	ds.dashboard.AddWebSocketClient(client)

	// Handle client messages
	go ds.handleWebSocketClient(client)
}

// handleWebSocketClient handles messages from a WebSocket client
func (ds *DashboardServer) handleWebSocketClient(client *WebSocketClient) {
	defer func() {
		ds.dashboard.RemoveWebSocketClient(client.ID)
	}()

	// Set read deadline
	client.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Set pong handler
	client.Connection.SetPongHandler(func(string) error {
		client.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastSeen = time.Now()
		return nil
	})

	// Start ping routine
	go ds.pingClient(client)

	// Read messages
	for {
		messageType, message, err := client.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ds.logger.Error("WebSocket error", err, "client_id", client.ID)
			}
			break
		}

		if messageType == websocket.TextMessage {
			ds.handleWebSocketMessage(client, message)
		}
	}
}

// handleWebSocketMessage handles a message from a WebSocket client
func (ds *DashboardServer) handleWebSocketMessage(client *WebSocketClient, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		ds.logger.Error("Failed to parse WebSocket message", err, "client_id", client.ID)
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		ds.logger.Warn("WebSocket message missing type", "client_id", client.ID)
		return
	}

	switch msgType {
	case "ping":
		ds.sendWebSocketMessage(client, map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now(),
		})
	case "subscribe":
		// Handle subscription changes
		if subscriptions, ok := msg["subscriptions"].([]interface{}); ok {
			client.Subscriptions = make([]string, len(subscriptions))
			for i, sub := range subscriptions {
				if subStr, ok := sub.(string); ok {
					client.Subscriptions[i] = subStr
				}
			}
			ds.logger.Info("Client subscriptions updated", "client_id", client.ID, "subscriptions", client.Subscriptions)
		}
	default:
		ds.logger.Warn("Unknown WebSocket message type", "type", msgType, "client_id", client.ID)
	}
}

// pingClient sends periodic ping messages to keep connection alive
func (ds *DashboardServer) pingClient(client *WebSocketClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := client.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendWebSocketMessage sends a message to a WebSocket client
func (ds *DashboardServer) sendWebSocketMessage(client *WebSocketClient, message map[string]interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		ds.logger.Error("Failed to marshal WebSocket message", err)
		return
	}

	if err := client.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
		ds.logger.Error("Failed to send WebSocket message", err, "client_id", client.ID)
	}
}

// Middleware

// corsMiddleware handles CORS headers
func (ds *DashboardServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range ds.config.AllowedOrigins {
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

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authMiddleware handles authentication
func (ds *DashboardServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check and WebSocket upgrade
		if r.URL.Path == "/api/v1/dashboard/health" || r.URL.Path == "/ws" {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			token = r.URL.Query().Get("token")
		}

		if token != ds.config.AuthToken {
			ds.writeJSONResponse(w, http.StatusUnauthorized, map[string]interface{}{
				"error":   "Unauthorized",
				"message": "Invalid or missing authentication token",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper methods

// writeJSONResponse writes a JSON response
func (ds *DashboardServer) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		ds.logger.Error("Failed to encode JSON response", err)
	}
}

// getIntQueryParam gets an integer query parameter with default value
func (ds *DashboardServer) getIntQueryParam(r *http.Request, param string, defaultValue int) int {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}

	return defaultValue
}

// getUserID extracts user ID from request (simplified)
func (ds *DashboardServer) getUserID(r *http.Request) string {
	// In a real implementation, this would extract from JWT or session
	return "anonymous"
}

// getClientIP extracts client IP address
func (ds *DashboardServer) getClientIP(r *http.Request) string {
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
