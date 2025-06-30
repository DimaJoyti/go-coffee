package services

import (
	"context"
	"fmt"
	"sync"
	"time"


)

// MonitoringService provides real-time monitoring and alerting for beverage operations
type MonitoringService struct {
	alertManager    AlertManager
	metricsCollector MetricsCollector
	notificationSvc NotificationService
	thresholds      *MonitoringThresholds
	activeAlerts    map[string]*Alert
	alertMutex      sync.RWMutex
	logger          Logger
}

// AlertManager defines the interface for alert management
type AlertManager interface {
	CreateAlert(ctx context.Context, alert *Alert) error
	UpdateAlert(ctx context.Context, alertID string, status AlertStatus) error
	GetActiveAlerts(ctx context.Context) ([]*Alert, error)
	GetAlertHistory(ctx context.Context, timeframe time.Duration) ([]*Alert, error)
	AcknowledgeAlert(ctx context.Context, alertID string, acknowledgedBy string) error
}

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	CollectSystemMetrics(ctx context.Context) (*SystemMetrics, error)
	CollectBeverageMetrics(ctx context.Context, beverageID string) (*BeverageMetrics, error)
	CollectQualityMetrics(ctx context.Context) (*QualityMetrics, error)
	CollectInventoryMetrics(ctx context.Context) (*InventoryMetrics, error)
	CollectPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error)
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendAlert(ctx context.Context, alert *Alert, channels []string) error
	SendDashboardUpdate(ctx context.Context, update *DashboardUpdate) error
	SendReport(ctx context.Context, report *MonitoringReport, recipients []string) error
}

// MonitoringThresholds contains threshold values for monitoring
type MonitoringThresholds struct {
	QualityScore        ThresholdConfig `json:"quality_score"`
	CustomerRating      ThresholdConfig `json:"customer_rating"`
	ProductionCost      ThresholdConfig `json:"production_cost"`
	InventoryLevel      ThresholdConfig `json:"inventory_level"`
	SystemPerformance   ThresholdConfig `json:"system_performance"`
	ErrorRate           ThresholdConfig `json:"error_rate"`
	ResponseTime        ThresholdConfig `json:"response_time"`
	SalesVolume         ThresholdConfig `json:"sales_volume"`
	ProfitMargin        ThresholdConfig `json:"profit_margin"`
	CustomerSatisfaction ThresholdConfig `json:"customer_satisfaction"`
}

// ThresholdConfig defines threshold configuration
type ThresholdConfig struct {
	Critical float64 `json:"critical"`
	Warning  float64 `json:"warning"`
	Info     float64 `json:"info"`
	Enabled  bool    `json:"enabled"`
}

// Alert represents a monitoring alert
type Alert struct {
	ID            string                 `json:"id"`
	Type          AlertType              `json:"type"`
	Severity      AlertSeverity          `json:"severity"`
	Status        AlertStatus            `json:"status"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Source        string                 `json:"source"`
	BeverageID    string                 `json:"beverage_id,omitempty"`
	MetricName    string                 `json:"metric_name"`
	CurrentValue  float64                `json:"current_value"`
	ThresholdValue float64               `json:"threshold_value"`
	Timestamp     time.Time              `json:"timestamp"`
	AcknowledgedAt *time.Time            `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string                `json:"acknowledged_by,omitempty"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy    string                 `json:"resolved_by,omitempty"`
	Tags          map[string]string      `json:"tags"`
	Metadata      map[string]interface{} `json:"metadata"`
	Actions       []AlertAction          `json:"actions"`
}

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeQuality     AlertType = "quality"
	AlertTypeInventory   AlertType = "inventory"
	AlertTypePerformance AlertType = "performance"
	AlertTypeSystem      AlertType = "system"
	AlertTypeSales       AlertType = "sales"
	AlertTypeCost        AlertType = "cost"
	AlertTypeCustomer    AlertType = "customer"
)

// AlertSeverity defines the severity of an alert
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// AlertStatus defines the status of an alert
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusSuppressed   AlertStatus = "suppressed"
)

// AlertAction represents an action that can be taken for an alert
type AlertAction struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`        // manual, automatic
	Handler     string                 `json:"handler"`     // function or service to call
	Parameters  map[string]interface{} `json:"parameters"`
	Enabled     bool                   `json:"enabled"`
}

// SystemMetrics contains system-level metrics
type SystemMetrics struct {
	Timestamp       time.Time `json:"timestamp"`
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	DiskUsage       float64   `json:"disk_usage"`
	NetworkLatency  float64   `json:"network_latency"`
	ActiveSessions  int       `json:"active_sessions"`
	RequestsPerSec  float64   `json:"requests_per_sec"`
	ErrorRate       float64   `json:"error_rate"`
	ResponseTime    float64   `json:"response_time"`
	DatabaseConnections int   `json:"database_connections"`
	QueueLength     int       `json:"queue_length"`
}

// InventoryMetrics contains inventory-related metrics
type InventoryMetrics struct {
	Timestamp        time.Time                  `json:"timestamp"`
	TotalItems       int                        `json:"total_items"`
	LowStockItems    int                        `json:"low_stock_items"`
	OutOfStockItems  int                        `json:"out_of_stock_items"`
	ExpiringItems    int                        `json:"expiring_items"`
	TurnoverRate     float64                    `json:"turnover_rate"`
	StockValue       float64                    `json:"stock_value"`
	WastePercentage  float64                    `json:"waste_percentage"`
	IngredientLevels map[string]IngredientLevel `json:"ingredient_levels"`
}

// IngredientLevel contains level information for an ingredient
type IngredientLevel struct {
	Name         string    `json:"name"`
	CurrentLevel float64   `json:"current_level"`
	MinLevel     float64   `json:"min_level"`
	MaxLevel     float64   `json:"max_level"`
	Unit         string    `json:"unit"`
	Status       string    `json:"status"`
	LastUpdated  time.Time `json:"last_updated"`
}

// DashboardUpdate contains real-time dashboard updates
type DashboardUpdate struct {
	Type        string                 `json:"type"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	BeverageID  string                 `json:"beverage_id,omitempty"`
	Severity    string                 `json:"severity,omitempty"`
	Message     string                 `json:"message,omitempty"`
}

// MonitoringReport contains a comprehensive monitoring report
type MonitoringReport struct {
	ReportID      string                 `json:"report_id"`
	Timeframe     time.Duration          `json:"timeframe"`
	GeneratedAt   time.Time              `json:"generated_at"`
	Summary       *MonitoringSummary     `json:"summary"`
	Alerts        []*Alert               `json:"alerts"`
	Metrics       *AggregatedMetrics     `json:"metrics"`
	Trends        *MonitoringTrends      `json:"trends"`
	Recommendations []string             `json:"recommendations"`
	HealthScore   float64                `json:"health_score"`
}

// MonitoringSummary contains a summary of monitoring data
type MonitoringSummary struct {
	TotalAlerts       int                    `json:"total_alerts"`
	CriticalAlerts    int                    `json:"critical_alerts"`
	WarningAlerts     int                    `json:"warning_alerts"`
	ResolvedAlerts    int                    `json:"resolved_alerts"`
	AverageResolutionTime time.Duration     `json:"average_resolution_time"`
	TopIssues         []string               `json:"top_issues"`
	SystemUptime      float64                `json:"system_uptime"`
	PerformanceScore  float64                `json:"performance_score"`
	QualityScore      float64                `json:"quality_score"`
}

// AggregatedMetrics contains aggregated metrics over time
type AggregatedMetrics struct {
	System    *SystemMetrics    `json:"system"`
	Quality   *QualityMetrics   `json:"quality"`
	Inventory *InventoryMetrics `json:"inventory"`
	Beverages map[string]*BeverageMetrics `json:"beverages"`
}

// MonitoringTrends contains trend analysis
type MonitoringTrends struct {
	AlertFrequency    map[string]int     `json:"alert_frequency"`    // alert type -> count
	MetricTrends      map[string]float64 `json:"metric_trends"`      // metric -> trend percentage
	QualityTrend      string             `json:"quality_trend"`      // improving, declining, stable
	PerformanceTrend  string             `json:"performance_trend"`
	InventoryTrend    string             `json:"inventory_trend"`
	PredictedIssues   []string           `json:"predicted_issues"`
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(
	alertManager AlertManager,
	metricsCollector MetricsCollector,
	notificationSvc NotificationService,
	thresholds *MonitoringThresholds,
	logger Logger,
) *MonitoringService {
	return &MonitoringService{
		alertManager:    alertManager,
		metricsCollector: metricsCollector,
		notificationSvc: notificationSvc,
		thresholds:      thresholds,
		activeAlerts:    make(map[string]*Alert),
		logger:          logger,
	}
}

// StartMonitoring starts the monitoring process
func (ms *MonitoringService) StartMonitoring(ctx context.Context, interval time.Duration) error {
	ms.logger.Info("Starting monitoring service", "interval", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ms.logger.Info("Monitoring service stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := ms.performMonitoringCycle(ctx); err != nil {
				ms.logger.Error("Monitoring cycle failed", err)
			}
		}
	}
}

// performMonitoringCycle performs a single monitoring cycle
func (ms *MonitoringService) performMonitoringCycle(ctx context.Context) error {
	ms.logger.Debug("Performing monitoring cycle")

	// Collect system metrics
	systemMetrics, err := ms.metricsCollector.CollectSystemMetrics(ctx)
	if err != nil {
		ms.logger.Error("Failed to collect system metrics", err)
	} else {
		ms.checkSystemThresholds(ctx, systemMetrics)
	}

	// Collect quality metrics
	qualityMetrics, err := ms.metricsCollector.CollectQualityMetrics(ctx)
	if err != nil {
		ms.logger.Error("Failed to collect quality metrics", err)
	} else {
		ms.checkQualityThresholds(ctx, qualityMetrics)
	}

	// Collect inventory metrics
	inventoryMetrics, err := ms.metricsCollector.CollectInventoryMetrics(ctx)
	if err != nil {
		ms.logger.Error("Failed to collect inventory metrics", err)
	} else {
		ms.checkInventoryThresholds(ctx, inventoryMetrics)
	}

	// Send dashboard updates
	ms.sendDashboardUpdates(ctx, systemMetrics, qualityMetrics, inventoryMetrics)

	return nil
}

// checkSystemThresholds checks system metrics against thresholds
func (ms *MonitoringService) checkSystemThresholds(ctx context.Context, metrics *SystemMetrics) {
	// Check CPU usage
	if ms.thresholds.SystemPerformance.Enabled {
		if metrics.CPUUsage >= ms.thresholds.SystemPerformance.Critical {
			ms.createAlert(ctx, &Alert{
				Type:           AlertTypeSystem,
				Severity:       AlertSeverityCritical,
				Title:          "High CPU Usage",
				Description:    fmt.Sprintf("CPU usage is %.1f%%, exceeding critical threshold of %.1f%%", metrics.CPUUsage, ms.thresholds.SystemPerformance.Critical),
				Source:         "system_monitor",
				MetricName:     "cpu_usage",
				CurrentValue:   metrics.CPUUsage,
				ThresholdValue: ms.thresholds.SystemPerformance.Critical,
				Tags:           map[string]string{"metric": "cpu", "component": "system"},
			})
		}
	}

	// Check error rate
	if ms.thresholds.ErrorRate.Enabled {
		if metrics.ErrorRate >= ms.thresholds.ErrorRate.Warning {
			severity := AlertSeverityWarning
			if metrics.ErrorRate >= ms.thresholds.ErrorRate.Critical {
				severity = AlertSeverityCritical
			}

			ms.createAlert(ctx, &Alert{
				Type:           AlertTypeSystem,
				Severity:       severity,
				Title:          "High Error Rate",
				Description:    fmt.Sprintf("Error rate is %.2f%%, exceeding threshold", metrics.ErrorRate),
				Source:         "system_monitor",
				MetricName:     "error_rate",
				CurrentValue:   metrics.ErrorRate,
				ThresholdValue: ms.thresholds.ErrorRate.Warning,
				Tags:           map[string]string{"metric": "errors", "component": "system"},
			})
		}
	}

	// Check response time
	if ms.thresholds.ResponseTime.Enabled {
		if metrics.ResponseTime >= ms.thresholds.ResponseTime.Warning {
			severity := AlertSeverityWarning
			if metrics.ResponseTime >= ms.thresholds.ResponseTime.Critical {
				severity = AlertSeverityCritical
			}

			ms.createAlert(ctx, &Alert{
				Type:           AlertTypePerformance,
				Severity:       severity,
				Title:          "High Response Time",
				Description:    fmt.Sprintf("Response time is %.0fms, exceeding threshold", metrics.ResponseTime),
				Source:         "performance_monitor",
				MetricName:     "response_time",
				CurrentValue:   metrics.ResponseTime,
				ThresholdValue: ms.thresholds.ResponseTime.Warning,
				Tags:           map[string]string{"metric": "latency", "component": "api"},
			})
		}
	}
}

// checkQualityThresholds checks quality metrics against thresholds
func (ms *MonitoringService) checkQualityThresholds(ctx context.Context, metrics *QualityMetrics) {
	if !ms.thresholds.QualityScore.Enabled {
		return
	}

	if metrics.OverallScore <= ms.thresholds.QualityScore.Critical {
		ms.createAlert(ctx, &Alert{
			Type:           AlertTypeQuality,
			Severity:       AlertSeverityCritical,
			Title:          "Low Quality Score",
			Description:    fmt.Sprintf("Overall quality score is %.1f, below critical threshold of %.1f", metrics.OverallScore, ms.thresholds.QualityScore.Critical),
			Source:         "quality_monitor",
			MetricName:     "quality_score",
			CurrentValue:   metrics.OverallScore,
			ThresholdValue: ms.thresholds.QualityScore.Critical,
			Tags:           map[string]string{"metric": "quality", "component": "production"},
		})
	}
}

// checkInventoryThresholds checks inventory metrics against thresholds
func (ms *MonitoringService) checkInventoryThresholds(ctx context.Context, metrics *InventoryMetrics) {
	if !ms.thresholds.InventoryLevel.Enabled {
		return
	}

	// Check for low stock items
	if metrics.LowStockItems > 0 {
		ms.createAlert(ctx, &Alert{
			Type:        AlertTypeInventory,
			Severity:    AlertSeverityWarning,
			Title:       "Low Stock Alert",
			Description: fmt.Sprintf("%d ingredients are running low on stock", metrics.LowStockItems),
			Source:      "inventory_monitor",
			MetricName:  "low_stock_items",
			CurrentValue: float64(metrics.LowStockItems),
			Tags:        map[string]string{"metric": "inventory", "component": "stock"},
		})
	}

	// Check for out of stock items
	if metrics.OutOfStockItems > 0 {
		ms.createAlert(ctx, &Alert{
			Type:        AlertTypeInventory,
			Severity:    AlertSeverityCritical,
			Title:       "Out of Stock Alert",
			Description: fmt.Sprintf("%d ingredients are out of stock", metrics.OutOfStockItems),
			Source:      "inventory_monitor",
			MetricName:  "out_of_stock_items",
			CurrentValue: float64(metrics.OutOfStockItems),
			Tags:        map[string]string{"metric": "inventory", "component": "stock"},
		})
	}
}

// createAlert creates and processes a new alert
func (ms *MonitoringService) createAlert(ctx context.Context, alert *Alert) {
	alert.ID = fmt.Sprintf("alert_%d", time.Now().UnixNano())
	alert.Status = AlertStatusActive
	alert.Timestamp = time.Now()

	ms.alertMutex.Lock()
	ms.activeAlerts[alert.ID] = alert
	ms.alertMutex.Unlock()

	// Store alert
	if err := ms.alertManager.CreateAlert(ctx, alert); err != nil {
		ms.logger.Error("Failed to create alert", err, "alert_id", alert.ID)
	}

	// Send notifications
	channels := ms.getNotificationChannels(alert.Severity)
	if err := ms.notificationSvc.SendAlert(ctx, alert, channels); err != nil {
		ms.logger.Error("Failed to send alert notification", err, "alert_id", alert.ID)
	}

	ms.logger.Info("Alert created", 
		"alert_id", alert.ID,
		"type", alert.Type,
		"severity", alert.Severity,
		"title", alert.Title)
}

// sendDashboardUpdates sends real-time updates to dashboards
func (ms *MonitoringService) sendDashboardUpdates(ctx context.Context, system *SystemMetrics, quality *QualityMetrics, inventory *InventoryMetrics) {
	updates := []*DashboardUpdate{
		{
			Type:      "system_metrics",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"cpu_usage":    system.CPUUsage,
				"memory_usage": system.MemoryUsage,
				"error_rate":   system.ErrorRate,
				"response_time": system.ResponseTime,
			},
		},
		{
			Type:      "quality_metrics",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"overall_score":    quality.OverallScore,
				"taste_score":      quality.TasteScore,
				"appearance_score": quality.AppearanceScore,
			},
		},
		{
			Type:      "inventory_metrics",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"low_stock_items":   inventory.LowStockItems,
				"out_of_stock_items": inventory.OutOfStockItems,
				"turnover_rate":     inventory.TurnoverRate,
			},
		},
	}

	for _, update := range updates {
		if err := ms.notificationSvc.SendDashboardUpdate(ctx, update); err != nil {
			ms.logger.Error("Failed to send dashboard update", err, "type", update.Type)
		}
	}
}

// getNotificationChannels returns notification channels based on alert severity
func (ms *MonitoringService) getNotificationChannels(severity AlertSeverity) []string {
	switch severity {
	case AlertSeverityCritical:
		return []string{"slack", "email", "sms", "webhook"}
	case AlertSeverityWarning:
		return []string{"slack", "email"}
	case AlertSeverityInfo:
		return []string{"slack"}
	default:
		return []string{"slack"}
	}
}

// AcknowledgeAlert acknowledges an alert
func (ms *MonitoringService) AcknowledgeAlert(ctx context.Context, alertID string, acknowledgedBy string) error {
	ms.alertMutex.Lock()
	defer ms.alertMutex.Unlock()

	alert, exists := ms.activeAlerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.Status = AlertStatusAcknowledged
	alert.AcknowledgedAt = &now
	alert.AcknowledgedBy = acknowledgedBy

	return ms.alertManager.UpdateAlert(ctx, alertID, AlertStatusAcknowledged)
}

// ResolveAlert resolves an alert
func (ms *MonitoringService) ResolveAlert(ctx context.Context, alertID string, resolvedBy string) error {
	ms.alertMutex.Lock()
	defer ms.alertMutex.Unlock()

	alert, exists := ms.activeAlerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.Status = AlertStatusResolved
	alert.ResolvedAt = &now
	alert.ResolvedBy = resolvedBy

	delete(ms.activeAlerts, alertID)

	return ms.alertManager.UpdateAlert(ctx, alertID, AlertStatusResolved)
}

// GetActiveAlerts returns all active alerts
func (ms *MonitoringService) GetActiveAlerts(ctx context.Context) ([]*Alert, error) {
	return ms.alertManager.GetActiveAlerts(ctx)
}

// GenerateMonitoringReport generates a comprehensive monitoring report
func (ms *MonitoringService) GenerateMonitoringReport(ctx context.Context, timeframe time.Duration) (*MonitoringReport, error) {
	ms.logger.Info("Generating monitoring report", "timeframe", timeframe)

	report := &MonitoringReport{
		ReportID:    fmt.Sprintf("monitoring_report_%d", time.Now().Unix()),
		Timeframe:   timeframe,
		GeneratedAt: time.Now(),
	}

	// Get alert history
	alerts, err := ms.alertManager.GetAlertHistory(ctx, timeframe)
	if err == nil {
		report.Alerts = alerts
		report.Summary = ms.generateMonitoringSummary(alerts)
	}

	// Collect current metrics
	systemMetrics, _ := ms.metricsCollector.CollectSystemMetrics(ctx)
	qualityMetrics, _ := ms.metricsCollector.CollectQualityMetrics(ctx)
	inventoryMetrics, _ := ms.metricsCollector.CollectInventoryMetrics(ctx)

	report.Metrics = &AggregatedMetrics{
		System:    systemMetrics,
		Quality:   qualityMetrics,
		Inventory: inventoryMetrics,
		Beverages: make(map[string]*BeverageMetrics),
	}

	// Generate trends and recommendations
	report.Trends = ms.generateMonitoringTrends(alerts)
	report.Recommendations = ms.generateRecommendations(report)
	report.HealthScore = ms.calculateHealthScore(report)

	ms.logger.Info("Monitoring report generated", "report_id", report.ReportID)

	return report, nil
}

// generateMonitoringSummary generates a summary from alerts
func (ms *MonitoringService) generateMonitoringSummary(alerts []*Alert) *MonitoringSummary {
	summary := &MonitoringSummary{
		TopIssues: []string{},
	}

	if len(alerts) == 0 {
		return summary
	}

	summary.TotalAlerts = len(alerts)

	var totalResolutionTime time.Duration
	resolvedCount := 0

	for _, alert := range alerts {
		switch alert.Severity {
		case AlertSeverityCritical:
			summary.CriticalAlerts++
		case AlertSeverityWarning:
			summary.WarningAlerts++
		}

		if alert.Status == AlertStatusResolved {
			summary.ResolvedAlerts++
			if alert.ResolvedAt != nil {
				resolutionTime := alert.ResolvedAt.Sub(alert.Timestamp)
				totalResolutionTime += resolutionTime
				resolvedCount++
			}
		}
	}

	if resolvedCount > 0 {
		summary.AverageResolutionTime = totalResolutionTime / time.Duration(resolvedCount)
	}

	// Calculate scores (simplified)
	summary.SystemUptime = 99.5 // Would be calculated from actual uptime data
	summary.PerformanceScore = 85.0
	summary.QualityScore = 90.0

	return summary
}

// generateMonitoringTrends generates trend analysis
func (ms *MonitoringService) generateMonitoringTrends(alerts []*Alert) *MonitoringTrends {
	trends := &MonitoringTrends{
		AlertFrequency:   make(map[string]int),
		MetricTrends:     make(map[string]float64),
		PredictedIssues:  []string{},
	}

	// Count alert frequency by type
	for _, alert := range alerts {
		trends.AlertFrequency[string(alert.Type)]++
	}

	// Set trend directions (simplified)
	trends.QualityTrend = "stable"
	trends.PerformanceTrend = "improving"
	trends.InventoryTrend = "stable"

	return trends
}

// generateRecommendations generates recommendations based on monitoring data
func (ms *MonitoringService) generateRecommendations(report *MonitoringReport) []string {
	recommendations := []string{}

	if report.Summary.CriticalAlerts > 0 {
		recommendations = append(recommendations, "Address critical alerts immediately to prevent service disruption")
	}

	if report.Summary.AverageResolutionTime > time.Hour {
		recommendations = append(recommendations, "Improve alert response procedures to reduce resolution time")
	}

	if report.Metrics.Inventory != nil && report.Metrics.Inventory.LowStockItems > 5 {
		recommendations = append(recommendations, "Review inventory management processes to prevent stockouts")
	}

	return recommendations
}

// calculateHealthScore calculates an overall health score
func (ms *MonitoringService) calculateHealthScore(report *MonitoringReport) float64 {
	score := 100.0

	// Deduct points for alerts
	score -= float64(report.Summary.CriticalAlerts) * 10
	score -= float64(report.Summary.WarningAlerts) * 5

	// Deduct points for poor performance
	if report.Summary.PerformanceScore < 80 {
		score -= (80 - report.Summary.PerformanceScore)
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}

	return score
}
