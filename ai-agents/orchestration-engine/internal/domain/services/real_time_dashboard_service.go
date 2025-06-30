package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/config"
	"go-coffee-ai-agents/orchestration-engine/internal/infrastructure/analytics"
)

// RealTimeDashboardService provides comprehensive real-time analytics and dashboard functionality
type RealTimeDashboardService struct {
	realTimeDashboard  *analytics.RealTimeDashboard
	metricsCollector   *analytics.MetricsCollector
	eventStreamer      *analytics.EventStreamer
	dashboardServer    *analytics.SimpleDashboardServer
	analyticsService   *AnalyticsService
	
	config             *config.DashboardConfig
	logger             Logger
	
	// Dashboard state
	dashboardMetrics   *DashboardServiceMetrics
	
	// Control
	mutex              sync.RWMutex
	stopCh             chan struct{}
}

// DashboardServiceMetrics tracks dashboard service metrics
type DashboardServiceMetrics struct {
	EventsProcessed     int64     `json:"events_processed"`
	MetricsCollected    int64     `json:"metrics_collected"`
	DashboardViews      int64     `json:"dashboard_views"`
	APIRequests         int64     `json:"api_requests"`
	WebSocketConnections int64    `json:"websocket_connections"`
	ErrorsEncountered   int64     `json:"errors_encountered"`
	LastEventProcessed  time.Time `json:"last_event_processed"`
	LastMetricCollected time.Time `json:"last_metric_collected"`
	ServiceUptime       time.Duration `json:"service_uptime"`
	LastUpdated         time.Time `json:"last_updated"`
}

// DashboardReport represents a comprehensive dashboard report
type DashboardReport struct {
	Timestamp           time.Time                    `json:"timestamp"`
	ServiceHealth       string                       `json:"service_health"`
	DashboardMetrics    *DashboardServiceMetrics     `json:"dashboard_metrics"`
	RealTimeMetrics     *analytics.DashboardMetrics  `json:"real_time_metrics"`
	EventStreamStats    *analytics.EventStreamStats  `json:"event_stream_stats"`
	DashboardStats      *analytics.DashboardStats    `json:"dashboard_stats"`
	AnalyticsData       *DashboardData               `json:"analytics_data"`
	Recommendations     []string                     `json:"recommendations"`
	ComponentHealth     map[string]string            `json:"component_health"`
}

// NewRealTimeDashboardService creates a new real-time dashboard service
func NewRealTimeDashboardService(
	analyticsService *AnalyticsService,
	config *config.DashboardConfig,
	logger Logger,
) *RealTimeDashboardService {
	
	// Create dashboard configuration
	dashboardConfig := &analytics.DashboardConfig{
		Port:              config.Port,
		UpdateInterval:    time.Duration(config.UpdateIntervalSeconds) * time.Second,
		MaxEvents:         config.MaxEvents,
		EnableWebSocket:   config.EnableWebSocket,
		EnableHTTPAPI:     config.EnableHTTPAPI,
		EnableStaticFiles: config.EnableStaticFiles,
		StaticFilesPath:   config.StaticFilesPath,
		CORSEnabled:       config.CORSEnabled,
		AllowedOrigins:    config.AllowedOrigins,
		AuthEnabled:       config.AuthEnabled,
		AuthToken:         config.AuthToken,
	}

	// Create real-time dashboard
	realTimeDashboard := analytics.NewRealTimeDashboard(dashboardConfig, logger)

	// Create dashboard server
	dashboardServer := analytics.NewSimpleDashboardServer(realTimeDashboard, dashboardConfig, logger)

	rds := &RealTimeDashboardService{
		realTimeDashboard: realTimeDashboard,
		metricsCollector:  analytics.NewMetricsCollector(logger),
		eventStreamer:     analytics.NewEventStreamer(logger),
		dashboardServer:   dashboardServer,
		analyticsService:  analyticsService,
		config:           config,
		logger:           logger,
		dashboardMetrics: &DashboardServiceMetrics{
			LastUpdated: time.Now(),
		},
		stopCh: make(chan struct{}),
	}

	return rds
}

// Start starts the real-time dashboard service
func (rds *RealTimeDashboardService) Start(ctx context.Context) error {
	rds.logger.Info("Starting real-time dashboard service")

	// Start real-time dashboard
	if err := rds.realTimeDashboard.Start(ctx); err != nil {
		return fmt.Errorf("failed to start real-time dashboard: %w", err)
	}

	// Start event streamer
	if err := rds.eventStreamer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event streamer: %w", err)
	}

	// Start dashboard server
	if err := rds.dashboardServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start dashboard server: %w", err)
	}

	// Start dashboard monitoring
	go rds.dashboardMonitoringLoop(ctx)

	// Start analytics integration
	go rds.analyticsIntegrationLoop(ctx)

	// Create default event filters and processors
	rds.eventStreamer.CreateDefaultFilters()
	rds.eventStreamer.CreateDefaultProcessors()

	rds.logger.Info("Real-time dashboard service started successfully")
	return nil
}

// Stop stops the real-time dashboard service
func (rds *RealTimeDashboardService) Stop(ctx context.Context) error {
	rds.logger.Info("Stopping real-time dashboard service")
	
	close(rds.stopCh)
	
	// Stop dashboard server
	if err := rds.dashboardServer.Stop(ctx); err != nil {
		rds.logger.Error("Failed to stop dashboard server", err)
	}
	
	// Stop event streamer
	if err := rds.eventStreamer.Stop(ctx); err != nil {
		rds.logger.Error("Failed to stop event streamer", err)
	}
	
	// Stop real-time dashboard
	if err := rds.realTimeDashboard.Stop(ctx); err != nil {
		rds.logger.Error("Failed to stop real-time dashboard", err)
	}
	
	rds.logger.Info("Real-time dashboard service stopped")
	return nil
}

// RecordWorkflowEvent records a workflow-related event
func (rds *RealTimeDashboardService) RecordWorkflowEvent(workflowID, eventType, title string, data map[string]interface{}) {
	event := analytics.CreateWorkflowEvent(workflowID, eventType, title, data)
	rds.recordEvent(event)
}

// RecordAgentEvent records an agent-related event
func (rds *RealTimeDashboardService) RecordAgentEvent(agentType, eventType, title string, data map[string]interface{}) {
	event := analytics.CreateAgentEvent(agentType, eventType, title, data)
	rds.recordEvent(event)
}

// RecordSystemEvent records a system-related event
func (rds *RealTimeDashboardService) RecordSystemEvent(eventType, title, description string, data map[string]interface{}) {
	event := analytics.CreateSystemEvent(eventType, title, description, data)
	rds.recordEvent(event)
}

// RecordErrorEvent records an error event
func (rds *RealTimeDashboardService) RecordErrorEvent(component, errorType, message string, err error) {
	event := analytics.CreateErrorEvent(component, errorType, message, err)
	rds.recordEvent(event)
}

// recordEvent records an event and updates metrics
func (rds *RealTimeDashboardService) recordEvent(event *analytics.AnalyticsEvent) {
	// Record event in dashboard
	rds.realTimeDashboard.RecordEvent(event)

	// Publish event to streamer
	rds.eventStreamer.PublishEvent(event)

	// Update dashboard metrics
	rds.mutex.Lock()
	rds.dashboardMetrics.EventsProcessed++
	rds.dashboardMetrics.LastEventProcessed = time.Now()
	rds.dashboardMetrics.LastUpdated = time.Now()
	rds.mutex.Unlock()

	rds.logger.Debug("Dashboard event recorded", "type", event.Type, "category", event.Category)
}

// UpdateMetrics updates dashboard metrics from analytics service
func (rds *RealTimeDashboardService) UpdateMetrics(ctx context.Context) {
	// Get analytics data
	dashboardData, err := rds.analyticsService.GetDashboardData(ctx)
	if err != nil {
		rds.logger.Error("Failed to get analytics data", err)
		return
	}

	// Convert to dashboard metrics
	dashboardMetrics := rds.convertToDashboardMetrics(dashboardData)

	// Update real-time dashboard
	rds.realTimeDashboard.UpdateMetrics(dashboardMetrics)

	// Update metrics collection count
	rds.mutex.Lock()
	rds.dashboardMetrics.MetricsCollected++
	rds.dashboardMetrics.LastMetricCollected = time.Now()
	rds.dashboardMetrics.LastUpdated = time.Now()
	rds.mutex.Unlock()
}

// convertToDashboardMetrics converts analytics data to dashboard metrics
func (rds *RealTimeDashboardService) convertToDashboardMetrics(data *DashboardData) *analytics.DashboardMetrics {
	return &analytics.DashboardMetrics{
		Timestamp: time.Now(),
		
		// System metrics
		SystemHealth:   rds.calculateSystemHealth(data),
		CPUUsage:       data.System.CPUUsage,
		MemoryUsage:    data.System.MemoryUsage,
		DiskUsage:      data.System.DiskUsage,
		NetworkLatency: 50 * time.Millisecond, // Mock value
		
		// Workflow metrics
		ActiveWorkflows:    data.Workflows.ActiveWorkflows,
		CompletedWorkflows: data.Workflows.SuccessfulExecutions,
		FailedWorkflows:    data.Workflows.FailedExecutions,
		AvgWorkflowTime:    data.Workflows.AverageExecutionTime,
		WorkflowThroughput: data.Workflows.ThroughputPerHour,
		
		// Agent metrics
		ActiveAgents:         data.Agents.OnlineAgents,
		AgentCalls:           data.Agents.TotalRequests,
		AgentErrors:          data.Agents.FailedRequests,
		AvgAgentResponseTime: data.Agents.AverageResponseTime,
		AgentUtilization:     75.0, // Mock value
		
		// Performance metrics
		RequestsPerSecond: 125.0, // Mock value
		ErrorRate:         data.System.ErrorRate,
		ResponseTime:      data.System.ResponseTime,
		CacheHitRate:      data.System.CacheHitRate,
		
		// Security metrics (mock values)
		AuthAttempts:       1000,
		FailedLogins:       25,
		BlockedRequests:    10,
		SecurityViolations: 2,
		
		// Business metrics (mock values)
		TotalUsers:     500,
		ActiveSessions: data.System.ActiveConnections,
		DataProcessed:  data.System.NetworkIO,
		
		// Detailed breakdowns
		WorkflowsByType: rds.getWorkflowsByType(data),
		AgentsByType:    rds.getAgentsByType(data),
		ErrorsByType:    rds.getErrorsByType(data),
		ResponseTimesByEndpoint: rds.getResponseTimesByEndpoint(),
	}
}

// GetDashboardReport generates a comprehensive dashboard report
func (rds *RealTimeDashboardService) GetDashboardReport(ctx context.Context) (*DashboardReport, error) {
	rds.mutex.RLock()
	dashboardMetrics := *rds.dashboardMetrics
	rds.mutex.RUnlock()

	// Get real-time metrics
	realTimeMetrics := rds.realTimeDashboard.GetCurrentMetrics()

	// Get event stream stats
	eventStreamStats := rds.eventStreamer.GetStats()

	// Get dashboard stats
	dashboardStats := rds.realTimeDashboard.GetDashboardStats()

	// Get analytics data
	analyticsData, err := rds.analyticsService.GetDashboardData(ctx)
	if err != nil {
		rds.logger.Error("Failed to get analytics data for report", err)
		analyticsData = &DashboardData{Timestamp: time.Now()}
	}

	// Check component health
	componentHealth := rds.checkComponentHealth(ctx)

	// Calculate service health
	serviceHealth := rds.calculateSystemHealth(analyticsData)

	// Generate recommendations
	recommendations := rds.generateRecommendations(&dashboardMetrics, componentHealth)

	report := &DashboardReport{
		Timestamp:        time.Now(),
		ServiceHealth:    serviceHealth,
		DashboardMetrics: &dashboardMetrics,
		RealTimeMetrics:  realTimeMetrics,
		EventStreamStats: eventStreamStats,
		DashboardStats:   dashboardStats,
		AnalyticsData:    analyticsData,
		Recommendations:  recommendations,
		ComponentHealth:  componentHealth,
	}

	return report, nil
}

// GetDashboardMetrics returns current dashboard metrics
func (rds *RealTimeDashboardService) GetDashboardMetrics() *DashboardServiceMetrics {
	rds.mutex.RLock()
	defer rds.mutex.RUnlock()
	
	metricsCopy := *rds.dashboardMetrics
	return &metricsCopy
}

// GetDashboardURL returns the dashboard URL
func (rds *RealTimeDashboardService) GetDashboardURL() string {
	return fmt.Sprintf("http://localhost:%d", rds.config.Port)
}

// SubscribeToEvents creates an event subscription
func (rds *RealTimeDashboardService) SubscribeToEvents(id, name string, filter *analytics.EventFilter) *analytics.EventSubscriber {
	return rds.eventStreamer.Subscribe(id, name, filter)
}

// UnsubscribeFromEvents removes an event subscription
func (rds *RealTimeDashboardService) UnsubscribeFromEvents(id string) {
	rds.eventStreamer.Unsubscribe(id)
}

// dashboardMonitoringLoop monitors dashboard service health
func (rds *RealTimeDashboardService) dashboardMonitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rds.stopCh:
			return
		case <-ticker.C:
			rds.updateServiceMetrics(startTime)
		}
	}
}

// analyticsIntegrationLoop integrates with analytics service
func (rds *RealTimeDashboardService) analyticsIntegrationLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rds.stopCh:
			return
		case <-ticker.C:
			rds.UpdateMetrics(ctx)
		}
	}
}

// updateServiceMetrics updates service-level metrics
func (rds *RealTimeDashboardService) updateServiceMetrics(startTime time.Time) {
	rds.mutex.Lock()
	rds.dashboardMetrics.ServiceUptime = time.Since(startTime)
	rds.dashboardMetrics.LastUpdated = time.Now()
	rds.mutex.Unlock()
}

// Helper methods

func (rds *RealTimeDashboardService) calculateSystemHealth(data *DashboardData) string {
	if data.System == nil {
		return "unknown"
	}

	score := 100.0
	if data.System.CPUUsage > 80 {
		score -= 20
	}
	if data.System.MemoryUsage > 85 {
		score -= 20
	}
	if data.System.ErrorRate > 5 {
		score -= 25
	}

	if score >= 90 {
		return "excellent"
	} else if score >= 75 {
		return "good"
	} else if score >= 60 {
		return "fair"
	} else if score >= 40 {
		return "poor"
	} else {
		return "critical"
	}
}

func (rds *RealTimeDashboardService) getWorkflowsByType(data *DashboardData) map[string]int64 {
	// Mock implementation - in real scenario, this would analyze workflow types
	return map[string]int64{
		"data_processing": data.Workflows.ActiveWorkflows / 3,
		"research":        data.Workflows.ActiveWorkflows / 4,
		"analysis":        data.Workflows.ActiveWorkflows / 5,
		"reporting":       data.Workflows.ActiveWorkflows / 6,
	}
}

func (rds *RealTimeDashboardService) getAgentsByType(data *DashboardData) map[string]int64 {
	agentsByType := make(map[string]int64)
	for _, agent := range data.Agents.AgentMetrics {
		agentsByType[agent.AgentType]++
	}
	return agentsByType
}

func (rds *RealTimeDashboardService) getErrorsByType(data *DashboardData) map[string]int64 {
	// Mock implementation
	return map[string]int64{
		"validation_error":     data.Workflows.FailedExecutions / 3,
		"timeout_error":        data.Workflows.FailedExecutions / 4,
		"connection_error":     data.Workflows.FailedExecutions / 5,
		"authentication_error": data.Workflows.FailedExecutions / 6,
	}
}

func (rds *RealTimeDashboardService) getResponseTimesByEndpoint() map[string]time.Duration {
	return map[string]time.Duration{
		"/api/v1/workflows":     180 * time.Millisecond,
		"/api/v1/agents":        220 * time.Millisecond,
		"/api/v1/analytics":     150 * time.Millisecond,
		"/api/v1/dashboard":     100 * time.Millisecond,
	}
}

func (rds *RealTimeDashboardService) checkComponentHealth(ctx context.Context) map[string]string {
	health := make(map[string]string)

	// Check real-time dashboard
	if err := rds.realTimeDashboard.Health(ctx); err != nil {
		health["real_time_dashboard"] = "unhealthy"
	} else {
		health["real_time_dashboard"] = "healthy"
	}

	// Check event streamer
	if err := rds.eventStreamer.Health(ctx); err != nil {
		health["event_streamer"] = "unhealthy"
	} else {
		health["event_streamer"] = "healthy"
	}

	// Check dashboard server
	if err := rds.dashboardServer.Health(ctx); err != nil {
		health["dashboard_server"] = "unhealthy"
	} else {
		health["dashboard_server"] = "healthy"
	}

	return health
}

func (rds *RealTimeDashboardService) generateRecommendations(metrics *DashboardServiceMetrics, componentHealth map[string]string) []string {
	var recommendations []string

	// Check event processing
	if time.Since(metrics.LastEventProcessed) > time.Hour {
		recommendations = append(recommendations, "No events processed recently. Verify event generation and processing pipeline.")
	}

	// Check metrics collection
	if time.Since(metrics.LastMetricCollected) > time.Hour {
		recommendations = append(recommendations, "No metrics collected recently. Verify metrics collection is functioning properly.")
	}

	// Check component health
	unhealthyComponents := 0
	for _, health := range componentHealth {
		if health != "healthy" {
			unhealthyComponents++
		}
	}

	if unhealthyComponents > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d dashboard components are unhealthy. Review component status and resolve issues.", unhealthyComponents))
	}

	return recommendations
}

// Health checks the health of the real-time dashboard service
func (rds *RealTimeDashboardService) Health(ctx context.Context) error {
	// Check if service is running
	if time.Since(rds.dashboardMetrics.LastUpdated) > 5*time.Minute {
		return fmt.Errorf("dashboard service not updating metrics")
	}

	// Check component health
	componentHealth := rds.checkComponentHealth(ctx)
	for component, health := range componentHealth {
		if health != "healthy" {
			return fmt.Errorf("component %s is unhealthy", component)
		}
	}

	return nil
}
