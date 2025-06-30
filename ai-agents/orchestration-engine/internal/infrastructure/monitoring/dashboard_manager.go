package monitoring

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// DashboardManager manages multiple custom dashboards
type DashboardManager struct {
	dashboards    map[string]*CustomDashboard
	templates     map[string]*DashboardTemplate
	dataProviders map[string]DataProvider
	alertManager  *DashboardAlertManager
	config        *DashboardManagerConfig
	logger        Logger
	mutex         sync.RWMutex
	stopCh        chan struct{}
}

// DashboardManagerConfig contains dashboard manager configuration
type DashboardManagerConfig struct {
	MaxDashboards      int           `json:"max_dashboards"`
	DefaultRefreshRate time.Duration `json:"default_refresh_rate"`
	EnableAutoRefresh  bool          `json:"enable_auto_refresh"`
	EnableAlerts       bool          `json:"enable_alerts"`
	EnableTemplates    bool          `json:"enable_templates"`
	StorageBackend     string        `json:"storage_backend"`
	CacheTimeout       time.Duration `json:"cache_timeout"`
	MaxWidgetsPerDash  int           `json:"max_widgets_per_dashboard"`
	EnableSharing      bool          `json:"enable_sharing"`
	EnableExport       bool          `json:"enable_export"`
}

// DashboardTemplate represents a dashboard template
type DashboardTemplate struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Category    string               `json:"category"`
	Tags        []string             `json:"tags"`
	Widgets     []*WidgetTemplate    `json:"widgets"`
	Layout      *DashboardLayout     `json:"layout"`
	Config      *DashboardConfig     `json:"config"`
	Variables   map[string]*Variable `json:"variables"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Author      string               `json:"author"`
	Version     string               `json:"version"`
}

// WidgetTemplate represents a widget template
type WidgetTemplate struct {
	Type          WidgetType             `json:"type"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Position      *WidgetPosition        `json:"position"`
	Size          *WidgetSize            `json:"size"`
	DataSource    string                 `json:"data_source"`
	Query         string                 `json:"query"`
	Visualization *VisualizationConfig   `json:"visualization"`
	Options       map[string]interface{} `json:"options"`
}

// Variable represents a dashboard variable
type Variable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Label        string      `json:"label"`
	DefaultValue interface{} `json:"default_value"`
	Options      []string    `json:"options"`
	Query        string      `json:"query"`
	Required     bool        `json:"required"`
}

// NotificationChannel interface for alert notifications
type NotificationChannel interface {
	Send(ctx context.Context, alert *Alert) error
	GetType() string
	IsEnabled() bool
}

// Alert represents an alert instance
type Alert struct {
	ID          string                 `json:"id"`
	RuleID      string                 `json:"rule_id"`
	DashboardID string                 `json:"dashboard_id"`
	WidgetID    string                 `json:"widget_id"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Severity    AlertSeverity          `json:"severity"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      AlertStatus            `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertStatus represents alert status
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusSilenced AlertStatus = "silenced"
)

// NewDashboardManager creates a new dashboard manager
func NewDashboardManager(config *DashboardManagerConfig, logger Logger) *DashboardManager {
	if config == nil {
		config = DefaultDashboardManagerConfig()
	}

	dm := &DashboardManager{
		dashboards:    make(map[string]*CustomDashboard),
		templates:     make(map[string]*DashboardTemplate),
		dataProviders: make(map[string]DataProvider),
		config:        config,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}

	// Initialize alert manager if enabled
	if config.EnableAlerts {
		dm.alertManager = NewDashboardAlertManager(logger)
	}

	return dm
}

// DefaultDashboardManagerConfig returns default configuration
func DefaultDashboardManagerConfig() *DashboardManagerConfig {
	return &DashboardManagerConfig{
		MaxDashboards:      100,
		DefaultRefreshRate: 30 * time.Second,
		EnableAutoRefresh:  true,
		EnableAlerts:       true,
		EnableTemplates:    true,
		StorageBackend:     "memory",
		CacheTimeout:       5 * time.Minute,
		MaxWidgetsPerDash:  50,
		EnableSharing:      true,
		EnableExport:       true,
	}
}

// Start starts the dashboard manager
func (dm *DashboardManager) Start(ctx context.Context) error {
	dm.logger.Info("Starting dashboard manager")

	// Start auto-refresh if enabled
	if dm.config.EnableAutoRefresh {
		go dm.autoRefreshLoop(ctx)
	}

	// Start alert manager if enabled
	if dm.alertManager != nil {
		go dm.alertManager.Start(ctx)
	}

	// Load default templates
	dm.loadDefaultTemplates()

	dm.logger.Info("Dashboard manager started")
	return nil
}

// Stop stops the dashboard manager
func (dm *DashboardManager) Stop(ctx context.Context) error {
	dm.logger.Info("Stopping dashboard manager")

	close(dm.stopCh)

	if dm.alertManager != nil {
		dm.alertManager.Stop(ctx)
	}

	dm.logger.Info("Dashboard manager stopped")
	return nil
}

// CreateDashboard creates a new dashboard
func (dm *DashboardManager) CreateDashboard(id, name, owner string) (*CustomDashboard, error) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if len(dm.dashboards) >= dm.config.MaxDashboards {
		return nil, fmt.Errorf("maximum number of dashboards (%d) reached", dm.config.MaxDashboards)
	}

	if _, exists := dm.dashboards[id]; exists {
		return nil, fmt.Errorf("dashboard %s already exists", id)
	}

	dashboard := NewCustomDashboard(id, name, owner, dm.logger)

	// Add data providers to dashboard
	for name, provider := range dm.dataProviders {
		dashboard.AddDataProvider(name, provider)
	}

	dm.dashboards[id] = dashboard
	dm.logger.Info("Dashboard created", "id", id, "name", name, "owner", owner)

	return dashboard, nil
}

// CreateDashboardFromTemplate creates a dashboard from a template
func (dm *DashboardManager) CreateDashboardFromTemplate(templateID, dashboardID, name, owner string, variables map[string]interface{}) (*CustomDashboard, error) {
	dm.mutex.RLock()
	template, exists := dm.templates[templateID]
	dm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("template %s not found", templateID)
	}

	dashboard, err := dm.CreateDashboard(dashboardID, name, owner)
	if err != nil {
		return nil, err
	}

	// Apply template
	if err := dm.applyTemplate(dashboard, template, variables); err != nil {
		dm.DeleteDashboard(dashboardID)
		return nil, fmt.Errorf("failed to apply template: %w", err)
	}

	dm.logger.Info("Dashboard created from template", "template_id", templateID, "dashboard_id", dashboardID)
	return dashboard, nil
}

// GetDashboard returns a dashboard by ID
func (dm *DashboardManager) GetDashboard(id string) (*CustomDashboard, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	dashboard, exists := dm.dashboards[id]
	if !exists {
		return nil, fmt.Errorf("dashboard %s not found", id)
	}

	return dashboard, nil
}

// ListDashboards returns all dashboards
func (dm *DashboardManager) ListDashboards() []*DashboardInfo {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	dashboards := make([]*DashboardInfo, 0, len(dm.dashboards))
	for _, dashboard := range dm.dashboards {
		dashboards = append(dashboards, dashboard.GetDashboardInfo())
	}

	return dashboards
}

// DeleteDashboard deletes a dashboard
func (dm *DashboardManager) DeleteDashboard(id string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if _, exists := dm.dashboards[id]; !exists {
		return fmt.Errorf("dashboard %s not found", id)
	}

	delete(dm.dashboards, id)
	dm.logger.Info("Dashboard deleted", "id", id)

	return nil
}

// AddDataProvider adds a data provider
func (dm *DashboardManager) AddDataProvider(name string, provider DataProvider) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.dataProviders[name] = provider

	// Add to existing dashboards
	for _, dashboard := range dm.dashboards {
		dashboard.AddDataProvider(name, provider)
	}

	dm.logger.Info("Data provider added", "name", name)
}

// AddTemplate adds a dashboard template
func (dm *DashboardManager) AddTemplate(template *DashboardTemplate) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if template.ID == "" {
		template.ID = fmt.Sprintf("template_%d", time.Now().UnixNano())
	}

	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	dm.templates[template.ID] = template
	dm.logger.Info("Template added", "id", template.ID, "name", template.Name)

	return nil
}

// GetTemplate returns a template by ID
func (dm *DashboardManager) GetTemplate(id string) (*DashboardTemplate, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	template, exists := dm.templates[id]
	if !exists {
		return nil, fmt.Errorf("template %s not found", id)
	}

	return template, nil
}

// ListTemplates returns all templates
func (dm *DashboardManager) ListTemplates() []*DashboardTemplate {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	templates := make([]*DashboardTemplate, 0, len(dm.templates))
	for _, template := range dm.templates {
		templates = append(templates, template)
	}

	return templates
}

// RefreshDashboard refreshes all widgets in a dashboard
func (dm *DashboardManager) RefreshDashboard(ctx context.Context, dashboardID string) (map[string]*DataResult, error) {
	dashboard, err := dm.GetDashboard(dashboardID)
	if err != nil {
		return nil, err
	}

	return dashboard.RefreshAllWidgets(ctx)
}

// autoRefreshLoop automatically refreshes dashboards
func (dm *DashboardManager) autoRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(dm.config.DefaultRefreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dm.stopCh:
			return
		case <-ticker.C:
			dm.refreshAllDashboards(ctx)
		}
	}
}

// refreshAllDashboards refreshes all active dashboards
func (dm *DashboardManager) refreshAllDashboards(ctx context.Context) {
	dm.mutex.RLock()
	dashboards := make([]*CustomDashboard, 0, len(dm.dashboards))
	for _, dashboard := range dm.dashboards {
		if dashboard.isActive {
			dashboards = append(dashboards, dashboard)
		}
	}
	dm.mutex.RUnlock()

	for _, dashboard := range dashboards {
		go func(d *CustomDashboard) {
			if _, err := d.RefreshAllWidgets(ctx); err != nil {
				dm.logger.Error("Failed to refresh dashboard", err, "dashboard_id", d.id)
			}
		}(dashboard)
	}
}

// applyTemplate applies a template to a dashboard
func (dm *DashboardManager) applyTemplate(dashboard *CustomDashboard, template *DashboardTemplate, variables map[string]interface{}) error {
	// Apply layout
	if template.Layout != nil {
		dashboard.UpdateLayout(template.Layout)
	}

	// Apply config
	if template.Config != nil {
		dashboard.UpdateConfig(template.Config)
	}

	// Apply widgets
	for _, widgetTemplate := range template.Widgets {
		widget := &DashboardWidget{
			Type:            widgetTemplate.Type,
			Title:           dm.substituteVariables(widgetTemplate.Title, variables),
			Description:     dm.substituteVariables(widgetTemplate.Description, variables),
			Position:        widgetTemplate.Position,
			Size:            widgetTemplate.Size,
			DataSource:      widgetTemplate.DataSource,
			Query:           dm.substituteVariables(widgetTemplate.Query, variables),
			Visualization:   widgetTemplate.Visualization,
			RefreshInterval: 30 * time.Second,
			Filters:         make(map[string]interface{}),
			Options:         widgetTemplate.Options,
			IsVisible:       true,
		}

		if err := dashboard.AddWidget(widget); err != nil {
			return fmt.Errorf("failed to add widget: %w", err)
		}
	}

	return nil
}

// substituteVariables substitutes template variables in strings
func (dm *DashboardManager) substituteVariables(text string, variables map[string]interface{}) string {
	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// loadDefaultTemplates loads default dashboard templates
func (dm *DashboardManager) loadDefaultTemplates() {
	// System Overview Template
	systemTemplate := &DashboardTemplate{
		ID:          "system_overview",
		Name:        "System Overview",
		Description: "Comprehensive system monitoring dashboard",
		Category:    "system",
		Tags:        []string{"system", "monitoring", "overview"},
		Layout:      DefaultDashboardLayout(),
		Config:      DefaultDashboardConfig(),
		Variables:   make(map[string]*Variable),
		Author:      "system",
		Version:     "1.0",
		Widgets: []*WidgetTemplate{
			{
				Type:        WidgetTypeMetric,
				Title:       "CPU Usage",
				Description: "Current CPU utilization",
				Position:    &WidgetPosition{X: 0, Y: 0},
				Size:        &WidgetSize{Width: 3, Height: 2},
				DataSource:  "system_metrics",
				Query:       "SELECT cpu_usage FROM system_metrics ORDER BY timestamp DESC LIMIT 1",
				Visualization: &VisualizationConfig{
					ChartType: "gauge",
					Threshold: &ThresholdConfig{Warning: 70, Critical: 90, Unit: "%", Operator: ">"},
				},
			},
			{
				Type:        WidgetTypeChart,
				Title:       "Memory Usage Over Time",
				Description: "Memory utilization trend",
				Position:    &WidgetPosition{X: 3, Y: 0},
				Size:        &WidgetSize{Width: 6, Height: 4},
				DataSource:  "system_metrics",
				Query:       "SELECT timestamp, memory_usage FROM system_metrics WHERE timestamp > NOW() - INTERVAL 1 HOUR",
				Visualization: &VisualizationConfig{
					ChartType:   "line",
					ShowLegend:  true,
					ShowGrid:    true,
					TimeRange:   "1h",
					Aggregation: "avg",
				},
			},
		},
	}

	// Workflow Analytics Template
	workflowTemplate := &DashboardTemplate{
		ID:          "workflow_analytics",
		Name:        "Workflow Analytics",
		Description: "Workflow execution monitoring and analytics",
		Category:    "workflow",
		Tags:        []string{"workflow", "analytics", "execution"},
		Layout:      DefaultDashboardLayout(),
		Config:      DefaultDashboardConfig(),
		Variables:   make(map[string]*Variable),
		Author:      "system",
		Version:     "1.0",
		Widgets: []*WidgetTemplate{
			{
				Type:        WidgetTypeMetric,
				Title:       "Active Workflows",
				Description: "Number of currently running workflows",
				Position:    &WidgetPosition{X: 0, Y: 0},
				Size:        &WidgetSize{Width: 3, Height: 2},
				DataSource:  "workflow_metrics",
				Query:       "SELECT COUNT(*) FROM workflows WHERE status = 'running'",
				Visualization: &VisualizationConfig{
					ChartType: "metric",
				},
			},
			{
				Type:        WidgetTypeChart,
				Title:       "Workflow Success Rate",
				Description: "Workflow execution success rate over time",
				Position:    &WidgetPosition{X: 3, Y: 0},
				Size:        &WidgetSize{Width: 6, Height: 4},
				DataSource:  "workflow_metrics",
				Query:       "SELECT DATE(created_at) as date, (COUNT(CASE WHEN status = 'completed' THEN 1 END) * 100.0 / COUNT(*)) as success_rate FROM workflows GROUP BY DATE(created_at)",
				Visualization: &VisualizationConfig{
					ChartType:   "bar",
					ShowLegend:  true,
					ShowGrid:    true,
					TimeRange:   "7d",
					Aggregation: "avg",
				},
			},
		},
	}

	dm.AddTemplate(systemTemplate)
	dm.AddTemplate(workflowTemplate)
}

// GetManagerStats returns dashboard manager statistics
func (dm *DashboardManager) GetManagerStats() *ManagerStats {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	activeDashboards := 0
	totalWidgets := 0
	for _, dashboard := range dm.dashboards {
		if dashboard.isActive {
			activeDashboards++
		}
		totalWidgets += len(dashboard.widgets)
	}

	return &ManagerStats{
		TotalDashboards:  len(dm.dashboards),
		ActiveDashboards: activeDashboards,
		TotalTemplates:   len(dm.templates),
		TotalWidgets:     totalWidgets,
		DataProviders:    len(dm.dataProviders),
		LastRefresh:      time.Now(),
	}
}

// ManagerStats represents dashboard manager statistics
type ManagerStats struct {
	TotalDashboards  int       `json:"total_dashboards"`
	ActiveDashboards int       `json:"active_dashboards"`
	TotalTemplates   int       `json:"total_templates"`
	TotalWidgets     int       `json:"total_widgets"`
	DataProviders    int       `json:"data_providers"`
	LastRefresh      time.Time `json:"last_refresh"`
}
