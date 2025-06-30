package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/entities"
)

// AnalyticsService provides real-time analytics and insights for workflows and agents
type AnalyticsService struct {
	workflowRepo   WorkflowRepository
	executionRepo  ExecutionRepository
	agentRegistry  AgentRegistry
	eventPublisher EventPublisher
	logger         Logger
	
	// Real-time metrics cache
	metricsCache   map[string]*CachedMetrics
	cacheMutex     sync.RWMutex
	cacheExpiry    time.Duration
	
	// Analytics aggregators
	workflowAnalytics *WorkflowAnalytics
	agentAnalytics    *AgentAnalytics
	systemAnalytics   *SystemAnalytics
	
	// Real-time data streams
	metricsStream  chan *MetricsUpdate
	alertsStream   chan *Alert
	stopChan       chan struct{}
}

// CachedMetrics represents cached analytics data
type CachedMetrics struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// MetricsUpdate represents a real-time metrics update
type MetricsUpdate struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Alert represents a system alert
type Alert struct {
	ID          uuid.UUID              `json:"id"`
	Type        AlertType              `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Source      string                 `json:"source"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Acknowledged bool                  `json:"acknowledged"`
}

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeWorkflowFailure    AlertType = "workflow_failure"
	AlertTypeAgentDown          AlertType = "agent_down"
	AlertTypeHighErrorRate      AlertType = "high_error_rate"
	AlertTypePerformanceDegradation AlertType = "performance_degradation"
	AlertTypeResourceExhaustion AlertType = "resource_exhaustion"
	AlertTypeSecurityBreach     AlertType = "security_breach"
	AlertTypeThresholdExceeded  AlertType = "threshold_exceeded"
)

// AlertSeverity defines the severity of an alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// Analytics aggregators

type WorkflowAnalytics struct {
	TotalWorkflows       int64                    `json:"total_workflows"`
	ActiveWorkflows      int64                    `json:"active_workflows"`
	TotalExecutions      int64                    `json:"total_executions"`
	RunningExecutions    int64                    `json:"running_executions"`
	SuccessfulExecutions int64                    `json:"successful_executions"`
	FailedExecutions     int64                    `json:"failed_executions"`
	SuccessRate          float64                  `json:"success_rate"`
	AverageExecutionTime time.Duration            `json:"average_execution_time"`
	ThroughputPerHour    float64                  `json:"throughput_per_hour"`
	TopWorkflows         []*WorkflowMetricsSummary `json:"top_workflows"`
	RecentFailures       []*ExecutionFailure      `json:"recent_failures"`
	LastUpdated          time.Time                `json:"last_updated"`
}

type AgentAnalytics struct {
	TotalAgents        int64                  `json:"total_agents"`
	OnlineAgents       int64                  `json:"online_agents"`
	OfflineAgents      int64                  `json:"offline_agents"`
	BusyAgents         int64                  `json:"busy_agents"`
	TotalRequests      int64                  `json:"total_requests"`
	SuccessfulRequests int64                  `json:"successful_requests"`
	FailedRequests     int64                  `json:"failed_requests"`
	AverageResponseTime time.Duration         `json:"average_response_time"`
	AgentMetrics       []*AgentMetricsSummary `json:"agent_metrics"`
	HealthStatus       map[string]*AgentHealth `json:"health_status"`
	LastUpdated        time.Time              `json:"last_updated"`
}

type SystemAnalytics struct {
	SystemUptime        time.Duration          `json:"system_uptime"`
	CPUUsage            float64                `json:"cpu_usage"`
	MemoryUsage         float64                `json:"memory_usage"`
	DiskUsage           float64                `json:"disk_usage"`
	NetworkIO           int64                  `json:"network_io"`
	ActiveConnections   int64                  `json:"active_connections"`
	QueueDepth          int64                  `json:"queue_depth"`
	CacheHitRate        float64                `json:"cache_hit_rate"`
	ErrorRate           float64                `json:"error_rate"`
	ResponseTime        time.Duration          `json:"response_time"`
	LastUpdated         time.Time              `json:"last_updated"`
}

type WorkflowMetricsSummary struct {
	WorkflowID      uuid.UUID     `json:"workflow_id"`
	Name            string        `json:"name"`
	ExecutionCount  int64         `json:"execution_count"`
	SuccessRate     float64       `json:"success_rate"`
	AverageTime     time.Duration `json:"average_time"`
	LastExecution   time.Time     `json:"last_execution"`
}

type AgentMetricsSummary struct {
	AgentType       string        `json:"agent_type"`
	Status          AgentStatus   `json:"status"`
	RequestCount    int64         `json:"request_count"`
	SuccessRate     float64       `json:"success_rate"`
	ResponseTime    time.Duration `json:"response_time"`
	ErrorRate       float64       `json:"error_rate"`
	Load            float64       `json:"load"`
	LastSeen        time.Time     `json:"last_seen"`
}

type ExecutionFailure struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkflowName string   `json:"workflow_name"`
	Error       string    `json:"error"`
	FailedAt    time.Time `json:"failed_at"`
	Duration    time.Duration `json:"duration"`
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(
	workflowRepo WorkflowRepository,
	executionRepo ExecutionRepository,
	agentRegistry AgentRegistry,
	eventPublisher EventPublisher,
	logger Logger,
) *AnalyticsService {
	return &AnalyticsService{
		workflowRepo:   workflowRepo,
		executionRepo:  executionRepo,
		agentRegistry:  agentRegistry,
		eventPublisher: eventPublisher,
		logger:         logger,
		metricsCache:   make(map[string]*CachedMetrics),
		cacheExpiry:    5 * time.Minute,
		workflowAnalytics: &WorkflowAnalytics{},
		agentAnalytics:    &AgentAnalytics{},
		systemAnalytics:   &SystemAnalytics{},
		metricsStream:     make(chan *MetricsUpdate, 1000),
		alertsStream:      make(chan *Alert, 100),
		stopChan:          make(chan struct{}),
	}
}

// Start starts the analytics service
func (as *AnalyticsService) Start(ctx context.Context) error {
	as.logger.Info("Starting analytics service")

	// Start metrics collection
	go as.collectMetrics(ctx)
	
	// Start alert monitoring
	go as.monitorAlerts(ctx)
	
	// Start cache cleanup
	go as.cleanupCache(ctx)

	as.logger.Info("Analytics service started")
	return nil
}

// Stop stops the analytics service
func (as *AnalyticsService) Stop(ctx context.Context) error {
	as.logger.Info("Stopping analytics service")
	
	close(as.stopChan)
	close(as.metricsStream)
	close(as.alertsStream)
	
	as.logger.Info("Analytics service stopped")
	return nil
}

// GetDashboardData returns comprehensive dashboard data
func (as *AnalyticsService) GetDashboardData(ctx context.Context) (*DashboardData, error) {
	// Check cache first
	if cached := as.getCachedMetrics("dashboard"); cached != nil {
		if data, ok := cached.Data.(*DashboardData); ok {
			return data, nil
		}
	}

	// Collect fresh data
	workflowAnalytics, err := as.collectWorkflowAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect workflow analytics: %w", err)
	}

	agentAnalytics, err := as.collectAgentAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect agent analytics: %w", err)
	}

	systemAnalytics := as.collectSystemAnalytics(ctx)

	dashboardData := &DashboardData{
		Workflows: workflowAnalytics,
		Agents:    agentAnalytics,
		System:    systemAnalytics,
		Alerts:    as.getRecentAlerts(10),
		Timestamp: time.Now(),
	}

	// Cache the data
	as.setCachedMetrics("dashboard", dashboardData, 1*time.Minute)

	return dashboardData, nil
}

// GetWorkflowMetrics returns detailed workflow metrics
func (as *AnalyticsService) GetWorkflowMetrics(ctx context.Context, workflowID uuid.UUID) (*entities.WorkflowMetrics, error) {
	cacheKey := fmt.Sprintf("workflow_metrics_%s", workflowID)
	
	if cached := as.getCachedMetrics(cacheKey); cached != nil {
		if metrics, ok := cached.Data.(*entities.WorkflowMetrics); ok {
			return metrics, nil
		}
	}

	executions, err := as.executionRepo.List(ctx, &ExecutionFilter{
		WorkflowID: []uuid.UUID{workflowID},
		Limit:      1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	metrics := as.calculateWorkflowMetrics(executions)
	as.setCachedMetrics(cacheKey, metrics, 2*time.Minute)

	return metrics, nil
}

// GetAgentMetrics returns detailed agent metrics
func (as *AnalyticsService) GetAgentMetrics(ctx context.Context, agentType string) (*AgentMetricsSummary, error) {
	cacheKey := fmt.Sprintf("agent_metrics_%s", agentType)
	
	if cached := as.getCachedMetrics(cacheKey); cached != nil {
		if metrics, ok := cached.Data.(*AgentMetricsSummary); ok {
			return metrics, nil
		}
	}

	agent, err := as.agentRegistry.GetAgent(agentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	agentMetrics := agent.GetMetrics()
	health, _ := as.agentRegistry.GetAgentHealth(agentType)

	summary := &AgentMetricsSummary{
		AgentType:    agentType,
		Status:       agent.GetStatus(),
		RequestCount: agentMetrics.TotalRequests,
		SuccessRate:  float64(agentMetrics.SuccessfulRequests) / float64(agentMetrics.TotalRequests) * 100,
		ResponseTime: agentMetrics.AverageResponseTime,
		ErrorRate:    float64(agentMetrics.FailedRequests) / float64(agentMetrics.TotalRequests) * 100,
		Load:         agentMetrics.CurrentLoad,
		LastSeen:     agentMetrics.LastUpdated,
	}

	if health != nil {
		summary.LastSeen = health.LastSeen
	}

	as.setCachedMetrics(cacheKey, summary, 30*time.Second)
	return summary, nil
}

// GetRealTimeMetrics returns real-time metrics stream
func (as *AnalyticsService) GetRealTimeMetrics() <-chan *MetricsUpdate {
	return as.metricsStream
}

// GetAlerts returns alerts stream
func (as *AnalyticsService) GetAlerts() <-chan *Alert {
	return as.alertsStream
}

// collectMetrics collects metrics periodically
func (as *AnalyticsService) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-as.stopChan:
			return
		case <-ticker.C:
			as.updateMetrics(ctx)
		}
	}
}

// updateMetrics updates all metrics
func (as *AnalyticsService) updateMetrics(ctx context.Context) {
	// Update workflow analytics
	workflowAnalytics, err := as.collectWorkflowAnalytics(ctx)
	if err != nil {
		as.logger.Error("Failed to collect workflow analytics", err)
	} else {
		as.workflowAnalytics = workflowAnalytics
		as.publishMetricsUpdate("workflow", map[string]interface{}{
			"analytics": workflowAnalytics,
		})
	}

	// Update agent analytics
	agentAnalytics, err := as.collectAgentAnalytics(ctx)
	if err != nil {
		as.logger.Error("Failed to collect agent analytics", err)
	} else {
		as.agentAnalytics = agentAnalytics
		as.publishMetricsUpdate("agent", map[string]interface{}{
			"analytics": agentAnalytics,
		})
	}

	// Update system analytics
	systemAnalytics := as.collectSystemAnalytics(ctx)
	as.systemAnalytics = systemAnalytics
	as.publishMetricsUpdate("system", map[string]interface{}{
		"analytics": systemAnalytics,
	})
}

// collectWorkflowAnalytics collects workflow analytics
func (as *AnalyticsService) collectWorkflowAnalytics(ctx context.Context) (*WorkflowAnalytics, error) {
	workflows, err := as.workflowRepo.List(ctx, &WorkflowFilter{Limit: 1000})
	if err != nil {
		return nil, err
	}

	executions, err := as.executionRepo.List(ctx, &ExecutionFilter{Limit: 10000})
	if err != nil {
		return nil, err
	}

	analytics := &WorkflowAnalytics{
		TotalWorkflows:    int64(len(workflows)),
		TotalExecutions:   int64(len(executions)),
		TopWorkflows:      []*WorkflowMetricsSummary{},
		RecentFailures:    []*ExecutionFailure{},
		LastUpdated:       time.Now(),
	}

	// Count active workflows
	for _, workflow := range workflows {
		if workflow.IsActive {
			analytics.ActiveWorkflows++
		}
	}

	// Analyze executions
	var totalDuration time.Duration
	runningCount := int64(0)
	successCount := int64(0)
	failedCount := int64(0)

	for _, execution := range executions {
		switch execution.Status {
		case entities.WorkflowStatusRunning:
			runningCount++
		case entities.WorkflowStatusCompleted:
			successCount++
			totalDuration += execution.Duration
		case entities.WorkflowStatusFailed:
			failedCount++
			analytics.RecentFailures = append(analytics.RecentFailures, &ExecutionFailure{
				ExecutionID:  execution.ID,
				WorkflowID:   execution.WorkflowID,
				WorkflowName: "Unknown", // Would need to join with workflow data
				Error:        execution.Error.Message,
				FailedAt:     *execution.CompletedAt,
				Duration:     execution.Duration,
			})
		}
	}

	analytics.RunningExecutions = runningCount
	analytics.SuccessfulExecutions = successCount
	analytics.FailedExecutions = failedCount

	if analytics.TotalExecutions > 0 {
		analytics.SuccessRate = float64(successCount) / float64(analytics.TotalExecutions) * 100
	}

	if successCount > 0 {
		analytics.AverageExecutionTime = totalDuration / time.Duration(successCount)
	}

	// Calculate throughput (executions per hour)
	if len(executions) > 0 {
		timeSpan := time.Since(executions[len(executions)-1].StartedAt)
		if timeSpan > 0 {
			analytics.ThroughputPerHour = float64(len(executions)) / timeSpan.Hours()
		}
	}

	return analytics, nil
}

// collectAgentAnalytics collects agent analytics
func (as *AnalyticsService) collectAgentAnalytics(ctx context.Context) (*AgentAnalytics, error) {
	agents := as.agentRegistry.ListAgents()
	
	analytics := &AgentAnalytics{
		TotalAgents:   int64(len(agents)),
		AgentMetrics:  []*AgentMetricsSummary{},
		HealthStatus:  make(map[string]*AgentHealth),
		LastUpdated:   time.Now(),
	}

	var totalRequests, successfulRequests, failedRequests int64
	var totalResponseTime time.Duration
	agentCount := 0

	for agentType, agent := range agents {
		status := agent.GetStatus()
		metrics := agent.GetMetrics()
		health, _ := as.agentRegistry.GetAgentHealth(agentType)

		switch status {
		case AgentStatusOnline:
			analytics.OnlineAgents++
		case AgentStatusOffline:
			analytics.OfflineAgents++
		case AgentStatusBusy:
			analytics.BusyAgents++
		}

		totalRequests += metrics.TotalRequests
		successfulRequests += metrics.SuccessfulRequests
		failedRequests += metrics.FailedRequests
		totalResponseTime += metrics.AverageResponseTime
		agentCount++

		summary := &AgentMetricsSummary{
			AgentType:    agentType,
			Status:       status,
			RequestCount: metrics.TotalRequests,
			ResponseTime: metrics.AverageResponseTime,
			Load:         metrics.CurrentLoad,
			LastSeen:     metrics.LastUpdated,
		}

		if metrics.TotalRequests > 0 {
			summary.SuccessRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100
			summary.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100
		}

		analytics.AgentMetrics = append(analytics.AgentMetrics, summary)
		analytics.HealthStatus[agentType] = health
	}

	analytics.TotalRequests = totalRequests
	analytics.SuccessfulRequests = successfulRequests
	analytics.FailedRequests = failedRequests

	if agentCount > 0 {
		analytics.AverageResponseTime = totalResponseTime / time.Duration(agentCount)
	}

	return analytics, nil
}

// collectSystemAnalytics collects system analytics
func (as *AnalyticsService) collectSystemAnalytics(ctx context.Context) *SystemAnalytics {
	// In a real implementation, this would collect actual system metrics
	return &SystemAnalytics{
		SystemUptime:      time.Since(time.Now().Add(-24 * time.Hour)), // Mock uptime
		CPUUsage:          45.2,  // Mock CPU usage
		MemoryUsage:       67.8,  // Mock memory usage
		DiskUsage:         23.4,  // Mock disk usage
		NetworkIO:         1024 * 1024 * 100, // Mock network I/O
		ActiveConnections: 150,   // Mock active connections
		QueueDepth:        5,     // Mock queue depth
		CacheHitRate:      89.5,  // Mock cache hit rate
		ErrorRate:         0.2,   // Mock error rate
		ResponseTime:      50 * time.Millisecond, // Mock response time
		LastUpdated:       time.Now(),
	}
}

// Helper methods

func (as *AnalyticsService) getCachedMetrics(key string) *CachedMetrics {
	as.cacheMutex.RLock()
	defer as.cacheMutex.RUnlock()

	cached, exists := as.metricsCache[key]
	if !exists {
		return nil
	}

	if time.Since(cached.Timestamp) > cached.TTL {
		return nil
	}

	return cached
}

func (as *AnalyticsService) setCachedMetrics(key string, data interface{}, ttl time.Duration) {
	as.cacheMutex.Lock()
	defer as.cacheMutex.Unlock()

	as.metricsCache[key] = &CachedMetrics{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

func (as *AnalyticsService) publishMetricsUpdate(updateType string, data map[string]interface{}) {
	update := &MetricsUpdate{
		Type:      updateType,
		Source:    "analytics-service",
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case as.metricsStream <- update:
	default:
		// Channel full, skip this update
	}
}

func (as *AnalyticsService) calculateWorkflowMetrics(executions []*entities.WorkflowExecution) *entities.WorkflowMetrics {
	if len(executions) == 0 {
		return &entities.WorkflowMetrics{LastUpdated: time.Now()}
	}

	var totalExecutions, successful, failed int64
	var totalDuration time.Duration
	var lastExecution *time.Time

	for _, exec := range executions {
		totalExecutions++
		
		if exec.Status == entities.WorkflowStatusCompleted {
			successful++
		} else if exec.Status == entities.WorkflowStatusFailed {
			failed++
		}

		totalDuration += exec.Duration
		
		if lastExecution == nil || exec.StartedAt.After(*lastExecution) {
			lastExecution = &exec.StartedAt
		}
	}

	successRate := float64(successful) / float64(totalExecutions) * 100
	errorRate := float64(failed) / float64(totalExecutions) * 100
	avgExecutionTime := totalDuration / time.Duration(totalExecutions)

	return &entities.WorkflowMetrics{
		TotalExecutions:      totalExecutions,
		SuccessfulExecutions: successful,
		FailedExecutions:     failed,
		AverageExecutionTime: avgExecutionTime,
		LastExecutionTime:    lastExecution,
		SuccessRate:          successRate,
		ErrorRate:            errorRate,
		LastUpdated:          time.Now(),
	}
}

func (as *AnalyticsService) getRecentAlerts(limit int) []*Alert {
	// Mock implementation - in production this would fetch from alert storage
	return []*Alert{}
}

func (as *AnalyticsService) monitorAlerts(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-as.stopChan:
			return
		case <-ticker.C:
			as.checkAlertConditions(ctx)
		}
	}
}

func (as *AnalyticsService) checkAlertConditions(ctx context.Context) {
	// Check for high error rates
	if as.workflowAnalytics.SuccessRate < 90.0 {
		alert := &Alert{
			ID:        uuid.New(),
			Type:      AlertTypeHighErrorRate,
			Severity:  AlertSeverityHigh,
			Title:     "High Workflow Error Rate",
			Message:   fmt.Sprintf("Workflow success rate is %.1f%%, below threshold of 90%%", as.workflowAnalytics.SuccessRate),
			Source:    "analytics-service",
			Timestamp: time.Now(),
		}
		
		select {
		case as.alertsStream <- alert:
		default:
		}
	}

	// Check for agent failures
	for _, agentMetrics := range as.agentAnalytics.AgentMetrics {
		if agentMetrics.Status == AgentStatusOffline {
			alert := &Alert{
				ID:        uuid.New(),
				Type:      AlertTypeAgentDown,
				Severity:  AlertSeverityCritical,
				Title:     "Agent Offline",
				Message:   fmt.Sprintf("Agent %s is offline", agentMetrics.AgentType),
				Source:    "analytics-service",
				Data:      map[string]interface{}{"agent_type": agentMetrics.AgentType},
				Timestamp: time.Now(),
			}
			
			select {
			case as.alertsStream <- alert:
			default:
			}
		}
	}
}

func (as *AnalyticsService) cleanupCache(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-as.stopChan:
			return
		case <-ticker.C:
			as.cacheMutex.Lock()
			for key, cached := range as.metricsCache {
				if time.Since(cached.Timestamp) > cached.TTL {
					delete(as.metricsCache, key)
				}
			}
			as.cacheMutex.Unlock()
		}
	}
}

// DashboardData represents comprehensive dashboard data
type DashboardData struct {
	Workflows *WorkflowAnalytics `json:"workflows"`
	Agents    *AgentAnalytics    `json:"agents"`
	System    *SystemAnalytics   `json:"system"`
	Alerts    []*Alert           `json:"alerts"`
	Timestamp time.Time          `json:"timestamp"`
}
