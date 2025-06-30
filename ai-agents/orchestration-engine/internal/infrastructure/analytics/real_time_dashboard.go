package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// RealTimeDashboard provides real-time analytics and monitoring
type RealTimeDashboard struct {
	metricsCollector   *MetricsCollector
	eventStreamer      *EventStreamer
	dashboardServer    *DashboardServer
	config             *DashboardConfig
	logger             Logger
	
	// Real-time data
	currentMetrics     *DashboardMetrics
	recentEvents       []*AnalyticsEvent
	connectedClients   map[string]*WebSocketClient
	
	// Control
	mutex              sync.RWMutex
	stopCh             chan struct{}
}

// DashboardConfig contains dashboard configuration
type DashboardConfig struct {
	Port                int           `json:"port"`
	UpdateInterval      time.Duration `json:"update_interval"`
	MaxEvents           int           `json:"max_events"`
	EnableWebSocket     bool          `json:"enable_websocket"`
	EnableHTTPAPI       bool          `json:"enable_http_api"`
	EnableStaticFiles   bool          `json:"enable_static_files"`
	StaticFilesPath     string        `json:"static_files_path"`
	CORSEnabled         bool          `json:"cors_enabled"`
	AllowedOrigins      []string      `json:"allowed_origins"`
	AuthEnabled         bool          `json:"auth_enabled"`
	AuthToken           string        `json:"auth_token"`
}

// DashboardMetrics represents real-time dashboard metrics
type DashboardMetrics struct {
	Timestamp           time.Time                    `json:"timestamp"`
	
	// System metrics
	SystemHealth        string                       `json:"system_health"`
	CPUUsage            float64                      `json:"cpu_usage"`
	MemoryUsage         float64                      `json:"memory_usage"`
	DiskUsage           float64                      `json:"disk_usage"`
	NetworkLatency      time.Duration                `json:"network_latency"`
	
	// Workflow metrics
	ActiveWorkflows     int64                        `json:"active_workflows"`
	CompletedWorkflows  int64                        `json:"completed_workflows"`
	FailedWorkflows     int64                        `json:"failed_workflows"`
	AvgWorkflowTime     time.Duration                `json:"avg_workflow_time"`
	WorkflowThroughput  float64                      `json:"workflow_throughput"`
	
	// Agent metrics
	ActiveAgents        int64                        `json:"active_agents"`
	AgentCalls          int64                        `json:"agent_calls"`
	AgentErrors         int64                        `json:"agent_errors"`
	AvgAgentResponseTime time.Duration               `json:"avg_agent_response_time"`
	AgentUtilization    float64                      `json:"agent_utilization"`
	
	// Performance metrics
	RequestsPerSecond   float64                      `json:"requests_per_second"`
	ErrorRate           float64                      `json:"error_rate"`
	ResponseTime        time.Duration                `json:"response_time"`
	CacheHitRate        float64                      `json:"cache_hit_rate"`
	
	// Security metrics
	AuthAttempts        int64                        `json:"auth_attempts"`
	FailedLogins        int64                        `json:"failed_logins"`
	BlockedRequests     int64                        `json:"blocked_requests"`
	SecurityViolations  int64                        `json:"security_violations"`
	
	// Business metrics
	TotalUsers          int64                        `json:"total_users"`
	ActiveSessions      int64                        `json:"active_sessions"`
	DataProcessed       int64                        `json:"data_processed"`
	
	// Detailed breakdowns
	WorkflowsByType     map[string]int64             `json:"workflows_by_type"`
	AgentsByType        map[string]int64             `json:"agents_by_type"`
	ErrorsByType        map[string]int64             `json:"errors_by_type"`
	ResponseTimesByEndpoint map[string]time.Duration `json:"response_times_by_endpoint"`
}

// AnalyticsEvent represents a real-time analytics event
type AnalyticsEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Type        string                 `json:"type"`
	Category    string                 `json:"category"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Data        map[string]interface{} `json:"data"`
	Tags        []string               `json:"tags"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID         string          `json:"id"`
	Connection *websocket.Conn `json:"-"`
	UserID     string          `json:"user_id"`
	IPAddress  string          `json:"ip_address"`
	UserAgent  string          `json:"user_agent"`
	ConnectedAt time.Time      `json:"connected_at"`
	LastSeen   time.Time       `json:"last_seen"`
	Subscriptions []string     `json:"subscriptions"`
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewRealTimeDashboard creates a new real-time dashboard
func NewRealTimeDashboard(config *DashboardConfig, logger Logger) *RealTimeDashboard {
	if config == nil {
		config = DefaultDashboardConfig()
	}

	dashboard := &RealTimeDashboard{
		config:           config,
		logger:           logger,
		currentMetrics:   &DashboardMetrics{Timestamp: time.Now()},
		recentEvents:     make([]*AnalyticsEvent, 0, config.MaxEvents),
		connectedClients: make(map[string]*WebSocketClient),
		stopCh:           make(chan struct{}),
	}

	// Initialize components
	dashboard.metricsCollector = NewMetricsCollector(logger)
	dashboard.eventStreamer = NewEventStreamer(logger)
	dashboard.dashboardServer = NewDashboardServer(dashboard, config, logger)

	return dashboard
}

// DefaultDashboardConfig returns default dashboard configuration
func DefaultDashboardConfig() *DashboardConfig {
	return &DashboardConfig{
		Port:            8080,
		UpdateInterval:  1 * time.Second,
		MaxEvents:       1000,
		EnableWebSocket: true,
		EnableHTTPAPI:   true,
		EnableStaticFiles: true,
		StaticFilesPath: "./web/dashboard",
		CORSEnabled:     true,
		AllowedOrigins:  []string{"*"},
		AuthEnabled:     false,
		AuthToken:       "",
	}
}

// Start starts the real-time dashboard
func (rtd *RealTimeDashboard) Start(ctx context.Context) error {
	rtd.logger.Info("Starting real-time dashboard", "port", rtd.config.Port)

	// Start metrics collection
	go rtd.metricsCollectionLoop(ctx)

	// Start event streaming
	go rtd.eventStreamingLoop(ctx)

	// Start dashboard server
	if err := rtd.dashboardServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start dashboard server: %w", err)
	}

	rtd.logger.Info("Real-time dashboard started successfully")
	return nil
}

// Stop stops the real-time dashboard
func (rtd *RealTimeDashboard) Stop(ctx context.Context) error {
	rtd.logger.Info("Stopping real-time dashboard")
	
	close(rtd.stopCh)
	
	// Stop dashboard server
	if err := rtd.dashboardServer.Stop(ctx); err != nil {
		rtd.logger.Error("Failed to stop dashboard server", err)
	}
	
	// Close all WebSocket connections
	rtd.closeAllConnections()
	
	rtd.logger.Info("Real-time dashboard stopped")
	return nil
}

// RecordEvent records a new analytics event
func (rtd *RealTimeDashboard) RecordEvent(event *AnalyticsEvent) {
	rtd.mutex.Lock()
	defer rtd.mutex.Unlock()

	// Add event to recent events
	rtd.recentEvents = append(rtd.recentEvents, event)

	// Trim events if exceeding max
	if len(rtd.recentEvents) > rtd.config.MaxEvents {
		rtd.recentEvents = rtd.recentEvents[1:]
	}

	// Broadcast event to connected clients
	rtd.broadcastEvent(event)

	rtd.logger.Debug("Analytics event recorded", "type", event.Type, "category", event.Category)
}

// UpdateMetrics updates the current dashboard metrics
func (rtd *RealTimeDashboard) UpdateMetrics(metrics *DashboardMetrics) {
	rtd.mutex.Lock()
	defer rtd.mutex.Unlock()

	rtd.currentMetrics = metrics
	rtd.currentMetrics.Timestamp = time.Now()

	// Broadcast metrics to connected clients
	rtd.broadcastMetrics(metrics)
}

// GetCurrentMetrics returns the current dashboard metrics
func (rtd *RealTimeDashboard) GetCurrentMetrics() *DashboardMetrics {
	rtd.mutex.RLock()
	defer rtd.mutex.RUnlock()

	metricsCopy := *rtd.currentMetrics
	return &metricsCopy
}

// GetRecentEvents returns recent analytics events
func (rtd *RealTimeDashboard) GetRecentEvents(limit int) []*AnalyticsEvent {
	rtd.mutex.RLock()
	defer rtd.mutex.RUnlock()

	if limit <= 0 || limit > len(rtd.recentEvents) {
		limit = len(rtd.recentEvents)
	}

	events := make([]*AnalyticsEvent, limit)
	copy(events, rtd.recentEvents[len(rtd.recentEvents)-limit:])
	return events
}

// AddWebSocketClient adds a new WebSocket client
func (rtd *RealTimeDashboard) AddWebSocketClient(client *WebSocketClient) {
	rtd.mutex.Lock()
	defer rtd.mutex.Unlock()

	rtd.connectedClients[client.ID] = client
	rtd.logger.Info("WebSocket client connected", "client_id", client.ID, "user_id", client.UserID)

	// Send current metrics to new client
	rtd.sendMetricsToClient(client, rtd.currentMetrics)
}

// RemoveWebSocketClient removes a WebSocket client
func (rtd *RealTimeDashboard) RemoveWebSocketClient(clientID string) {
	rtd.mutex.Lock()
	defer rtd.mutex.Unlock()

	if client, exists := rtd.connectedClients[clientID]; exists {
		client.Connection.Close()
		delete(rtd.connectedClients, clientID)
		rtd.logger.Info("WebSocket client disconnected", "client_id", clientID)
	}
}

// GetConnectedClients returns information about connected clients
func (rtd *RealTimeDashboard) GetConnectedClients() []*WebSocketClient {
	rtd.mutex.RLock()
	defer rtd.mutex.RUnlock()

	clients := make([]*WebSocketClient, 0, len(rtd.connectedClients))
	for _, client := range rtd.connectedClients {
		clientCopy := *client
		clientCopy.Connection = nil // Don't include connection in response
		clients = append(clients, &clientCopy)
	}

	return clients
}

// metricsCollectionLoop collects metrics at regular intervals
func (rtd *RealTimeDashboard) metricsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(rtd.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rtd.stopCh:
			return
		case <-ticker.C:
			rtd.collectMetrics(ctx)
		}
	}
}

// eventStreamingLoop handles event streaming
func (rtd *RealTimeDashboard) eventStreamingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-rtd.stopCh:
			return
		case event := <-rtd.eventStreamer.EventChannel():
			rtd.RecordEvent(event)
		}
	}
}

// collectMetrics collects current system metrics
func (rtd *RealTimeDashboard) collectMetrics(ctx context.Context) {
	metrics := rtd.metricsCollector.CollectMetrics(ctx)
	rtd.UpdateMetrics(metrics)
}

// broadcastEvent broadcasts an event to all connected clients
func (rtd *RealTimeDashboard) broadcastEvent(event *AnalyticsEvent) {
	message := map[string]interface{}{
		"type": "event",
		"data": event,
	}

	rtd.broadcastMessage(message)
}

// broadcastMetrics broadcasts metrics to all connected clients
func (rtd *RealTimeDashboard) broadcastMetrics(metrics *DashboardMetrics) {
	message := map[string]interface{}{
		"type": "metrics",
		"data": metrics,
	}

	rtd.broadcastMessage(message)
}

// broadcastMessage broadcasts a message to all connected clients
func (rtd *RealTimeDashboard) broadcastMessage(message map[string]interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		rtd.logger.Error("Failed to marshal broadcast message", err)
		return
	}

	for clientID, client := range rtd.connectedClients {
		if err := client.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
			rtd.logger.Error("Failed to send message to client", err, "client_id", clientID)
			rtd.RemoveWebSocketClient(clientID)
		}
	}
}

// sendMetricsToClient sends current metrics to a specific client
func (rtd *RealTimeDashboard) sendMetricsToClient(client *WebSocketClient, metrics *DashboardMetrics) {
	message := map[string]interface{}{
		"type": "metrics",
		"data": metrics,
	}

	data, err := json.Marshal(message)
	if err != nil {
		rtd.logger.Error("Failed to marshal metrics message", err)
		return
	}

	if err := client.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
		rtd.logger.Error("Failed to send metrics to client", err, "client_id", client.ID)
		rtd.RemoveWebSocketClient(client.ID)
	}
}

// closeAllConnections closes all WebSocket connections
func (rtd *RealTimeDashboard) closeAllConnections() {
	rtd.mutex.Lock()
	defer rtd.mutex.Unlock()

	for clientID, client := range rtd.connectedClients {
		client.Connection.Close()
		delete(rtd.connectedClients, clientID)
	}
}

// GetDashboardStats returns dashboard statistics
func (rtd *RealTimeDashboard) GetDashboardStats() *DashboardStats {
	rtd.mutex.RLock()
	defer rtd.mutex.RUnlock()

	return &DashboardStats{
		ConnectedClients: len(rtd.connectedClients),
		TotalEvents:      len(rtd.recentEvents),
		LastUpdate:       rtd.currentMetrics.Timestamp,
		Uptime:           time.Since(rtd.currentMetrics.Timestamp),
	}
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	ConnectedClients int           `json:"connected_clients"`
	TotalEvents      int           `json:"total_events"`
	LastUpdate       time.Time     `json:"last_update"`
	Uptime           time.Duration `json:"uptime"`
}

// Health checks the health of the dashboard
func (rtd *RealTimeDashboard) Health(ctx context.Context) error {
	// Check if metrics are being updated
	if time.Since(rtd.currentMetrics.Timestamp) > 2*rtd.config.UpdateInterval {
		return fmt.Errorf("metrics not updated recently")
	}

	// Check if dashboard server is running
	if rtd.dashboardServer == nil {
		return fmt.Errorf("dashboard server not initialized")
	}

	return nil
}

// CreateSystemEvent creates a system analytics event
func CreateSystemEvent(eventType, title, description string, data map[string]interface{}) *AnalyticsEvent {
	return &AnalyticsEvent{
		ID:          fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        eventType,
		Category:    "system",
		Severity:    "info",
		Title:       title,
		Description: description,
		Source:      "orchestration_engine",
		Data:        data,
		Tags:        []string{"system", eventType},
	}
}

// CreateWorkflowEvent creates a workflow analytics event
func CreateWorkflowEvent(workflowID, eventType, title string, data map[string]interface{}) *AnalyticsEvent {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["workflow_id"] = workflowID

	return &AnalyticsEvent{
		ID:          fmt.Sprintf("workflow_%s_%d", workflowID, time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        eventType,
		Category:    "workflow",
		Severity:    "info",
		Title:       title,
		Description: fmt.Sprintf("Workflow %s: %s", workflowID, title),
		Source:      "workflow_engine",
		Data:        data,
		Tags:        []string{"workflow", eventType},
	}
}

// CreateAgentEvent creates an agent analytics event
func CreateAgentEvent(agentType, eventType, title string, data map[string]interface{}) *AnalyticsEvent {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["agent_type"] = agentType

	return &AnalyticsEvent{
		ID:          fmt.Sprintf("agent_%s_%d", agentType, time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        eventType,
		Category:    "agent",
		Severity:    "info",
		Title:       title,
		Description: fmt.Sprintf("Agent %s: %s", agentType, title),
		Source:      "agent_service",
		Data:        data,
		Tags:        []string{"agent", agentType, eventType},
	}
}

// CreateErrorEvent creates an error analytics event
func CreateErrorEvent(component, errorType, message string, err error) *AnalyticsEvent {
	data := map[string]interface{}{
		"component":   component,
		"error_type":  errorType,
		"error_message": message,
	}

	if err != nil {
		data["error_details"] = err.Error()
	}

	return &AnalyticsEvent{
		ID:          fmt.Sprintf("error_%s_%d", component, time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        "error",
		Category:    "error",
		Severity:    "error",
		Title:       fmt.Sprintf("Error in %s", component),
		Description: message,
		Source:      component,
		Data:        data,
		Tags:        []string{"error", component, errorType},
	}
}
