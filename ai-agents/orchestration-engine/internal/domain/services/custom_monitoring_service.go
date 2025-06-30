package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/config"
	"go-coffee-ai-agents/orchestration-engine/internal/infrastructure/monitoring"
)

// CustomMonitoringService provides comprehensive custom monitoring dashboard functionality
type CustomMonitoringService struct {
	dashboardManager *monitoring.DashboardManager
	alertManager     *monitoring.DashboardAlertManager
	dataProviders    map[string]monitoring.DataProvider
	
	config           *config.MonitoringConfig
	logger           Logger
	
	// Service state
	serviceMetrics   *MonitoringServiceMetrics
	
	// Control
	mutex            sync.RWMutex
	stopCh           chan struct{}
}

// MonitoringServiceMetrics tracks monitoring service metrics
type MonitoringServiceMetrics struct {
	DashboardsCreated   int64     `json:"dashboards_created"`
	WidgetsCreated      int64     `json:"widgets_created"`
	AlertsTriggered     int64     `json:"alerts_triggered"`
	DataQueriesExecuted int64     `json:"data_queries_executed"`
	TemplatesUsed       int64     `json:"templates_used"`
	ErrorsEncountered   int64     `json:"errors_encountered"`
	LastDashboardAccess time.Time `json:"last_dashboard_access"`
	LastAlertTriggered  time.Time `json:"last_alert_triggered"`
	ServiceUptime       time.Duration `json:"service_uptime"`
	LastUpdated         time.Time `json:"last_updated"`
}

// MonitoringReport represents a comprehensive monitoring report
type MonitoringReport struct {
	Timestamp           time.Time                        `json:"timestamp"`
	ServiceHealth       string                           `json:"service_health"`
	ServiceMetrics      *MonitoringServiceMetrics        `json:"service_metrics"`
	DashboardStats      *monitoring.ManagerStats         `json:"dashboard_stats"`
	AlertStats          *monitoring.AlertStats           `json:"alert_stats"`
	ActiveDashboards    []*monitoring.DashboardInfo      `json:"active_dashboards"`
	RecentAlerts        []*monitoring.Alert              `json:"recent_alerts"`
	DataProviderStatus  map[string]string                `json:"data_provider_status"`
	Recommendations     []string                         `json:"recommendations"`
	ComponentHealth     map[string]string                `json:"component_health"`
}

// NewCustomMonitoringService creates a new custom monitoring service
func NewCustomMonitoringService(config *config.MonitoringConfig, logger Logger) *CustomMonitoringService {
	
	// Create dashboard manager configuration
	dashboardConfig := &monitoring.DashboardManagerConfig{
		MaxDashboards:      config.MaxDashboards,
		DefaultRefreshRate: time.Duration(config.DefaultRefreshRateSeconds) * time.Second,
		EnableAutoRefresh:  config.EnableAutoRefresh,
		EnableAlerts:       config.EnableAlerts,
		EnableTemplates:    config.EnableTemplates,
		StorageBackend:     config.StorageBackend,
		CacheTimeout:       time.Duration(config.CacheTimeoutSeconds) * time.Second,
		MaxWidgetsPerDash:  config.MaxWidgetsPerDashboard,
		EnableSharing:      config.EnableSharing,
		EnableExport:       config.EnableExport,
	}

	// Create dashboard manager
	dashboardManager := monitoring.NewDashboardManager(dashboardConfig, logger)

	// Create alert manager
	alertManager := monitoring.NewDashboardAlertManager(logger)

	cms := &CustomMonitoringService{
		dashboardManager: dashboardManager,
		alertManager:     alertManager,
		dataProviders:    make(map[string]monitoring.DataProvider),
		config:          config,
		logger:          logger,
		serviceMetrics: &MonitoringServiceMetrics{
			LastUpdated: time.Now(),
		},
		stopCh: make(chan struct{}),
	}

	// Initialize data providers
	cms.initializeDataProviders()

	return cms
}

// Start starts the custom monitoring service
func (cms *CustomMonitoringService) Start(ctx context.Context) error {
	cms.logger.Info("Starting custom monitoring service")

	// Start dashboard manager
	if err := cms.dashboardManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start dashboard manager: %w", err)
	}

	// Start alert manager
	if err := cms.alertManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start alert manager: %w", err)
	}

	// Start monitoring loop
	go cms.monitoringLoop(ctx)

	// Create default dashboards
	cms.createDefaultDashboards()

	// Setup default alert channels
	cms.setupDefaultAlertChannels()

	cms.logger.Info("Custom monitoring service started successfully")
	return nil
}

// Stop stops the custom monitoring service
func (cms *CustomMonitoringService) Stop(ctx context.Context) error {
	cms.logger.Info("Stopping custom monitoring service")
	
	close(cms.stopCh)
	
	// Stop alert manager
	if err := cms.alertManager.Stop(ctx); err != nil {
		cms.logger.Error("Failed to stop alert manager", err)
	}
	
	// Stop dashboard manager
	if err := cms.dashboardManager.Stop(ctx); err != nil {
		cms.logger.Error("Failed to stop dashboard manager", err)
	}
	
	cms.logger.Info("Custom monitoring service stopped")
	return nil
}

// CreateDashboard creates a new custom dashboard
func (cms *CustomMonitoringService) CreateDashboard(id, name, owner string) (*monitoring.CustomDashboard, error) {
	dashboard, err := cms.dashboardManager.CreateDashboard(id, name, owner)
	if err != nil {
		return nil, err
	}

	// Update metrics
	cms.mutex.Lock()
	cms.serviceMetrics.DashboardsCreated++
	cms.serviceMetrics.LastUpdated = time.Now()
	cms.mutex.Unlock()

	cms.logger.Info("Custom dashboard created", "id", id, "name", name, "owner", owner)
	return dashboard, nil
}

// CreateDashboardFromTemplate creates a dashboard from a template
func (cms *CustomMonitoringService) CreateDashboardFromTemplate(templateID, dashboardID, name, owner string, variables map[string]interface{}) (*monitoring.CustomDashboard, error) {
	dashboard, err := cms.dashboardManager.CreateDashboardFromTemplate(templateID, dashboardID, name, owner, variables)
	if err != nil {
		return nil, err
	}

	// Update metrics
	cms.mutex.Lock()
	cms.serviceMetrics.DashboardsCreated++
	cms.serviceMetrics.TemplatesUsed++
	cms.serviceMetrics.LastUpdated = time.Now()
	cms.mutex.Unlock()

	cms.logger.Info("Dashboard created from template", "template_id", templateID, "dashboard_id", dashboardID)
	return dashboard, nil
}

// GetDashboard returns a dashboard by ID
func (cms *CustomMonitoringService) GetDashboard(id string) (*monitoring.CustomDashboard, error) {
	dashboard, err := cms.dashboardManager.GetDashboard(id)
	if err != nil {
		return nil, err
	}

	// Update access metrics
	cms.mutex.Lock()
	cms.serviceMetrics.LastDashboardAccess = time.Now()
	cms.serviceMetrics.LastUpdated = time.Now()
	cms.mutex.Unlock()

	return dashboard, nil
}

// ListDashboards returns all dashboards
func (cms *CustomMonitoringService) ListDashboards() []*monitoring.DashboardInfo {
	return cms.dashboardManager.ListDashboards()
}

// DeleteDashboard deletes a dashboard
func (cms *CustomMonitoringService) DeleteDashboard(id string) error {
	return cms.dashboardManager.DeleteDashboard(id)
}

// AddWidgetToDashboard adds a widget to a dashboard
func (cms *CustomMonitoringService) AddWidgetToDashboard(dashboardID string, widget *monitoring.DashboardWidget) error {
	dashboard, err := cms.dashboardManager.GetDashboard(dashboardID)
	if err != nil {
		return err
	}

	if err := dashboard.AddWidget(widget); err != nil {
		return err
	}

	// Update metrics
	cms.mutex.Lock()
	cms.serviceMetrics.WidgetsCreated++
	cms.serviceMetrics.LastUpdated = time.Now()
	cms.mutex.Unlock()

	cms.logger.Info("Widget added to dashboard", "dashboard_id", dashboardID, "widget_id", widget.ID)
	return nil
}

// RefreshDashboard refreshes all widgets in a dashboard
func (cms *CustomMonitoringService) RefreshDashboard(ctx context.Context, dashboardID string) (map[string]*monitoring.DataResult, error) {
	results, err := cms.dashboardManager.RefreshDashboard(ctx, dashboardID)
	if err != nil {
		return nil, err
	}

	// Update metrics
	cms.mutex.Lock()
	cms.serviceMetrics.DataQueriesExecuted += int64(len(results))
	cms.serviceMetrics.LastUpdated = time.Now()
	cms.mutex.Unlock()

	return results, nil
}

// AddAlertRule adds an alert rule
func (cms *CustomMonitoringService) AddAlertRule(rule *monitoring.AlertRule) error {
	return cms.alertManager.AddRule(rule)
}

// GetAlertRules returns all alert rules
func (cms *CustomMonitoringService) GetAlertRules() []*monitoring.AlertRule {
	return cms.alertManager.ListRules()
}

// GetTemplates returns all dashboard templates
func (cms *CustomMonitoringService) GetTemplates() []*monitoring.DashboardTemplate {
	return cms.dashboardManager.ListTemplates()
}

// GetMonitoringReport generates a comprehensive monitoring report
func (cms *CustomMonitoringService) GetMonitoringReport(ctx context.Context) (*MonitoringReport, error) {
	cms.mutex.RLock()
	serviceMetrics := *cms.serviceMetrics
	cms.mutex.RUnlock()

	// Get dashboard stats
	dashboardStats := cms.dashboardManager.GetManagerStats()

	// Get alert stats
	alertStats := cms.alertManager.GetAlertStats()

	// Get active dashboards
	activeDashboards := cms.dashboardManager.ListDashboards()

	// Check data provider status
	dataProviderStatus := cms.checkDataProviderStatus(ctx)

	// Check component health
	componentHealth := cms.checkComponentHealth(ctx)

	// Calculate service health
	serviceHealth := cms.calculateServiceHealth(componentHealth)

	// Generate recommendations
	recommendations := cms.generateRecommendations(&serviceMetrics, componentHealth)

	report := &MonitoringReport{
		Timestamp:          time.Now(),
		ServiceHealth:      serviceHealth,
		ServiceMetrics:     &serviceMetrics,
		DashboardStats:     dashboardStats,
		AlertStats:         alertStats,
		ActiveDashboards:   activeDashboards,
		RecentAlerts:       []*monitoring.Alert{}, // Would be populated from alert history
		DataProviderStatus: dataProviderStatus,
		Recommendations:    recommendations,
		ComponentHealth:    componentHealth,
	}

	return report, nil
}

// GetServiceMetrics returns current service metrics
func (cms *CustomMonitoringService) GetServiceMetrics() *MonitoringServiceMetrics {
	cms.mutex.RLock()
	defer cms.mutex.RUnlock()
	
	metricsCopy := *cms.serviceMetrics
	return &metricsCopy
}

// initializeDataProviders initializes data providers
func (cms *CustomMonitoringService) initializeDataProviders() {
	// System metrics provider
	systemProvider := monitoring.NewSystemMetricsDataProvider(cms.logger)
	cms.dataProviders["system_metrics"] = systemProvider
	cms.dashboardManager.AddDataProvider("system_metrics", systemProvider)

	// Workflow metrics provider
	workflowProvider := monitoring.NewWorkflowMetricsDataProvider(cms.logger)
	cms.dataProviders["workflow_metrics"] = workflowProvider
	cms.dashboardManager.AddDataProvider("workflow_metrics", workflowProvider)

	// Agent metrics provider
	agentProvider := monitoring.NewAgentMetricsDataProvider(cms.logger)
	cms.dataProviders["agent_metrics"] = agentProvider
	cms.dashboardManager.AddDataProvider("agent_metrics", agentProvider)

	cms.logger.Info("Data providers initialized", "count", len(cms.dataProviders))
}

// createDefaultDashboards creates default monitoring dashboards
func (cms *CustomMonitoringService) createDefaultDashboards() {
	// Create system overview dashboard
	systemDashboard, err := cms.dashboardManager.CreateDashboardFromTemplate(
		"system_overview",
		"default_system_overview",
		"System Overview",
		"system",
		map[string]interface{}{},
	)
	if err != nil {
		cms.logger.Error("Failed to create system overview dashboard", err)
	} else {
		cms.logger.Info("Created default system overview dashboard", "id", systemDashboard.GetDashboardInfo().ID)
	}

	// Create workflow analytics dashboard
	workflowDashboard, err := cms.dashboardManager.CreateDashboardFromTemplate(
		"workflow_analytics",
		"default_workflow_analytics",
		"Workflow Analytics",
		"system",
		map[string]interface{}{},
	)
	if err != nil {
		cms.logger.Error("Failed to create workflow analytics dashboard", err)
	} else {
		cms.logger.Info("Created default workflow analytics dashboard", "id", workflowDashboard.GetDashboardInfo().ID)
	}
}

// setupDefaultAlertChannels sets up default alert notification channels
func (cms *CustomMonitoringService) setupDefaultAlertChannels() {
	// Email notification channel
	emailChannel := monitoring.NewEmailNotificationChannel(
		"smtp.example.com:587",
		"alerts@example.com",
		"password",
		cms.logger,
	)
	cms.alertManager.AddNotificationChannel("email", emailChannel)

	// Slack notification channel
	slackChannel := monitoring.NewSlackNotificationChannel(
		"https://hooks.slack.com/services/...",
		"#alerts",
		cms.logger,
	)
	cms.alertManager.AddNotificationChannel("slack", slackChannel)

	// Webhook notification channel
	webhookChannel := monitoring.NewWebhookNotificationChannel(
		"https://api.example.com/alerts",
		map[string]string{"Authorization": "Bearer token"},
		cms.logger,
	)
	cms.alertManager.AddNotificationChannel("webhook", webhookChannel)

	cms.logger.Info("Default alert channels configured")
}

// monitoringLoop monitors service health and metrics
func (cms *CustomMonitoringService) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cms.stopCh:
			return
		case <-ticker.C:
			cms.updateServiceMetrics(startTime)
		}
	}
}

// updateServiceMetrics updates service-level metrics
func (cms *CustomMonitoringService) updateServiceMetrics(startTime time.Time) {
	cms.mutex.Lock()
	cms.serviceMetrics.ServiceUptime = time.Since(startTime)
	cms.serviceMetrics.LastUpdated = time.Now()
	cms.mutex.Unlock()
}

// checkDataProviderStatus checks the status of all data providers
func (cms *CustomMonitoringService) checkDataProviderStatus(ctx context.Context) map[string]string {
	status := make(map[string]string)

	for name, provider := range cms.dataProviders {
		// Test data provider by making a simple query
		_, err := provider.GetData(ctx, "SELECT 1", nil)
		if err != nil {
			status[name] = "unhealthy"
			cms.logger.Error("Data provider health check failed", err, "provider", name)
		} else {
			status[name] = "healthy"
		}
	}

	return status
}

// checkComponentHealth checks the health of all components
func (cms *CustomMonitoringService) checkComponentHealth(ctx context.Context) map[string]string {
	health := make(map[string]string)

	// Check dashboard manager
	dashboardStats := cms.dashboardManager.GetManagerStats()
	if dashboardStats.TotalDashboards >= 0 {
		health["dashboard_manager"] = "healthy"
	} else {
		health["dashboard_manager"] = "unhealthy"
	}

	// Check alert manager
	alertStats := cms.alertManager.GetAlertStats()
	if alertStats.TotalRules >= 0 {
		health["alert_manager"] = "healthy"
	} else {
		health["alert_manager"] = "unhealthy"
	}

	// Check data providers
	dataProviderStatus := cms.checkDataProviderStatus(ctx)
	healthyProviders := 0
	for _, status := range dataProviderStatus {
		if status == "healthy" {
			healthyProviders++
		}
	}

	if healthyProviders == len(dataProviderStatus) {
		health["data_providers"] = "healthy"
	} else if healthyProviders > 0 {
		health["data_providers"] = "degraded"
	} else {
		health["data_providers"] = "unhealthy"
	}

	return health
}

// calculateServiceHealth calculates overall service health
func (cms *CustomMonitoringService) calculateServiceHealth(componentHealth map[string]string) string {
	healthyCount := 0
	totalCount := len(componentHealth)

	for _, health := range componentHealth {
		if health == "healthy" {
			healthyCount++
		}
	}

	if totalCount == 0 {
		return "unknown"
	}

	healthPercentage := float64(healthyCount) / float64(totalCount) * 100

	if healthPercentage >= 100 {
		return "excellent"
	} else if healthPercentage >= 80 {
		return "good"
	} else if healthPercentage >= 60 {
		return "fair"
	} else if healthPercentage >= 40 {
		return "poor"
	} else {
		return "critical"
	}
}

// generateRecommendations generates monitoring recommendations
func (cms *CustomMonitoringService) generateRecommendations(metrics *MonitoringServiceMetrics, componentHealth map[string]string) []string {
	var recommendations []string

	// Check dashboard usage
	if metrics.DashboardsCreated == 0 {
		recommendations = append(recommendations, "No dashboards created yet. Consider creating dashboards to monitor system performance.")
	}

	// Check widget usage
	if metrics.WidgetsCreated < 5 {
		recommendations = append(recommendations, "Few widgets created. Add more widgets to dashboards for comprehensive monitoring.")
	}

	// Check alert configuration
	if metrics.AlertsTriggered == 0 && time.Since(metrics.LastAlertTriggered) > 24*time.Hour {
		recommendations = append(recommendations, "No recent alerts triggered. Verify alert rules are configured correctly.")
	}

	// Check component health
	unhealthyComponents := 0
	for _, health := range componentHealth {
		if health != "healthy" {
			unhealthyComponents++
		}
	}

	if unhealthyComponents > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d monitoring components are unhealthy. Review component status and resolve issues.", unhealthyComponents))
	}

	// Check data query frequency
	if metrics.DataQueriesExecuted < 100 {
		recommendations = append(recommendations, "Low data query activity. Ensure dashboards are being actively used and refreshed.")
	}

	return recommendations
}

// Health checks the health of the custom monitoring service
func (cms *CustomMonitoringService) Health(ctx context.Context) error {
	// Check if service is running
	if time.Since(cms.serviceMetrics.LastUpdated) > 5*time.Minute {
		return fmt.Errorf("monitoring service not updating metrics")
	}

	// Check component health
	componentHealth := cms.checkComponentHealth(ctx)
	for component, health := range componentHealth {
		if health == "unhealthy" {
			return fmt.Errorf("component %s is unhealthy", component)
		}
	}

	return nil
}
