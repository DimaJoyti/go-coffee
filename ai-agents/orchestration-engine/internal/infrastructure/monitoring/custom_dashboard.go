package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CustomDashboard provides customizable monitoring dashboards
type CustomDashboard struct {
	id              string
	name            string
	description     string
	widgets         map[string]*DashboardWidget
	layout          *DashboardLayout
	config          *DashboardConfig
	dataProviders   map[string]DataProvider
	alertRules      map[string]*AlertRule
	refreshInterval time.Duration
	isActive        bool
	createdAt       time.Time
	updatedAt       time.Time
	owner           string
	permissions     *DashboardPermissions
	mutex           sync.RWMutex
	logger          Logger
}

// DashboardWidget represents a dashboard widget
type DashboardWidget struct {
	ID              string                 `json:"id"`
	Type            WidgetType             `json:"type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Position        *WidgetPosition        `json:"position"`
	Size            *WidgetSize            `json:"size"`
	DataSource      string                 `json:"data_source"`
	Query           string                 `json:"query"`
	Visualization   *VisualizationConfig   `json:"visualization"`
	RefreshInterval time.Duration          `json:"refresh_interval"`
	Filters         map[string]interface{} `json:"filters"`
	Options         map[string]interface{} `json:"options"`
	IsVisible       bool                   `json:"is_visible"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// WidgetType represents different types of dashboard widgets
type WidgetType string

const (
	WidgetTypeMetric      WidgetType = "metric"
	WidgetTypeChart       WidgetType = "chart"
	WidgetTypeTable       WidgetType = "table"
	WidgetTypeGauge       WidgetType = "gauge"
	WidgetTypeHeatmap     WidgetType = "heatmap"
	WidgetTypeTimeSeries  WidgetType = "timeseries"
	WidgetTypeAlert       WidgetType = "alert"
	WidgetTypeLog         WidgetType = "log"
	WidgetTypeStatus      WidgetType = "status"
	WidgetTypeCustom      WidgetType = "custom"
)

// WidgetPosition represents widget position on dashboard
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WidgetSize represents widget dimensions
type WidgetSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// VisualizationConfig contains visualization settings
type VisualizationConfig struct {
	ChartType    string                 `json:"chart_type"`
	ColorScheme  string                 `json:"color_scheme"`
	ShowLegend   bool                   `json:"show_legend"`
	ShowGrid     bool                   `json:"show_grid"`
	Aggregation  string                 `json:"aggregation"`
	TimeRange    string                 `json:"time_range"`
	Threshold    *ThresholdConfig       `json:"threshold"`
	CustomConfig map[string]interface{} `json:"custom_config"`
}

// ThresholdConfig defines alert thresholds for widgets
type ThresholdConfig struct {
	Warning  float64 `json:"warning"`
	Critical float64 `json:"critical"`
	Unit     string  `json:"unit"`
	Operator string  `json:"operator"` // >, <, >=, <=, ==, !=
}

// DashboardLayout defines dashboard layout configuration
type DashboardLayout struct {
	Type        string                 `json:"type"` // grid, flex, custom
	Columns     int                    `json:"columns"`
	RowHeight   int                    `json:"row_height"`
	Margin      int                    `json:"margin"`
	Padding     int                    `json:"padding"`
	Responsive  bool                   `json:"responsive"`
	CustomCSS   string                 `json:"custom_css"`
	Theme       string                 `json:"theme"`
	Properties  map[string]interface{} `json:"properties"`
}

// DashboardConfig contains dashboard configuration
type DashboardConfig struct {
	AutoRefresh     bool          `json:"auto_refresh"`
	RefreshInterval time.Duration `json:"refresh_interval"`
	TimeZone        string        `json:"time_zone"`
	DateFormat      string        `json:"date_format"`
	EnableExport    bool          `json:"enable_export"`
	EnableSharing   bool          `json:"enable_sharing"`
	EnableAlerts    bool          `json:"enable_alerts"`
	MaxDataPoints   int           `json:"max_data_points"`
	CacheTimeout    time.Duration `json:"cache_timeout"`
}

// DashboardPermissions defines access control for dashboards
type DashboardPermissions struct {
	Owner       string   `json:"owner"`
	Viewers     []string `json:"viewers"`
	Editors     []string `json:"editors"`
	Admins      []string `json:"admins"`
	IsPublic    bool     `json:"is_public"`
	ShareToken  string   `json:"share_token"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// DataProvider interface for dashboard data sources
type DataProvider interface {
	GetData(ctx context.Context, query string, filters map[string]interface{}) (*DataResult, error)
	GetSchema(ctx context.Context) (*DataSchema, error)
	ValidateQuery(query string) error
	GetSupportedAggregations() []string
}

// DataResult represents query result data
type DataResult struct {
	Data      []map[string]interface{} `json:"data"`
	Columns   []string                 `json:"columns"`
	Types     map[string]string        `json:"types"`
	Count     int                      `json:"count"`
	Timestamp time.Time                `json:"timestamp"`
	Metadata  map[string]interface{}   `json:"metadata"`
}

// DataSchema represents data source schema
type DataSchema struct {
	Tables  map[string]*TableSchema `json:"tables"`
	Metrics map[string]*MetricSchema `json:"metrics"`
}

// TableSchema represents table structure
type TableSchema struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Columns     map[string]*ColumnSchema  `json:"columns"`
}

// ColumnSchema represents column information
type ColumnSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Nullable    bool   `json:"nullable"`
}

// MetricSchema represents metric information
type MetricSchema struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Unit        string   `json:"unit"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// AlertRule defines alerting rules for dashboard widgets
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	WidgetID    string                 `json:"widget_id"`
	Condition   string                 `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Operator    string                 `json:"operator"`
	Severity    AlertSeverity          `json:"severity"`
	Enabled     bool                   `json:"enabled"`
	Frequency   time.Duration          `json:"frequency"`
	Recipients  []string               `json:"recipients"`
	Message     string                 `json:"message"`
	Actions     []AlertAction          `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertAction represents actions to take when alert triggers
type AlertAction struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
}


// NewCustomDashboard creates a new custom dashboard
func NewCustomDashboard(id, name, owner string, logger Logger) *CustomDashboard {
	return &CustomDashboard{
		id:              id,
		name:            name,
		widgets:         make(map[string]*DashboardWidget),
		layout:          DefaultDashboardLayout(),
		config:          DefaultDashboardConfig(),
		dataProviders:   make(map[string]DataProvider),
		alertRules:      make(map[string]*AlertRule),
		refreshInterval: 30 * time.Second,
		isActive:        true,
		createdAt:       time.Now(),
		updatedAt:       time.Now(),
		owner:           owner,
		permissions:     DefaultDashboardPermissions(owner),
		logger:          logger,
	}
}

// DefaultDashboardLayout returns default dashboard layout
func DefaultDashboardLayout() *DashboardLayout {
	return &DashboardLayout{
		Type:       "grid",
		Columns:    12,
		RowHeight:  100,
		Margin:     10,
		Padding:    15,
		Responsive: true,
		Theme:      "default",
		Properties: make(map[string]interface{}),
	}
}

// DefaultDashboardConfig returns default dashboard configuration
func DefaultDashboardConfig() *DashboardConfig {
	return &DashboardConfig{
		AutoRefresh:     true,
		RefreshInterval: 30 * time.Second,
		TimeZone:        "UTC",
		DateFormat:      "2006-01-02 15:04:05",
		EnableExport:    true,
		EnableSharing:   true,
		EnableAlerts:    true,
		MaxDataPoints:   1000,
		CacheTimeout:    5 * time.Minute,
	}
}

// DefaultDashboardPermissions returns default dashboard permissions
func DefaultDashboardPermissions(owner string) *DashboardPermissions {
	return &DashboardPermissions{
		Owner:    owner,
		Viewers:  make([]string, 0),
		Editors:  make([]string, 0),
		Admins:   []string{owner},
		IsPublic: false,
	}
}

// AddWidget adds a widget to the dashboard
func (cd *CustomDashboard) AddWidget(widget *DashboardWidget) error {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	if widget.ID == "" {
		widget.ID = fmt.Sprintf("widget_%d", time.Now().UnixNano())
	}

	widget.CreatedAt = time.Now()
	widget.UpdatedAt = time.Now()

	cd.widgets[widget.ID] = widget
	cd.updatedAt = time.Now()

	cd.logger.Info("Widget added to dashboard", "dashboard_id", cd.id, "widget_id", widget.ID, "widget_type", widget.Type)
	return nil
}

// RemoveWidget removes a widget from the dashboard
func (cd *CustomDashboard) RemoveWidget(widgetID string) error {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	if _, exists := cd.widgets[widgetID]; !exists {
		return fmt.Errorf("widget %s not found", widgetID)
	}

	delete(cd.widgets, widgetID)
	cd.updatedAt = time.Now()

	cd.logger.Info("Widget removed from dashboard", "dashboard_id", cd.id, "widget_id", widgetID)
	return nil
}

// UpdateWidget updates an existing widget
func (cd *CustomDashboard) UpdateWidget(widgetID string, updates map[string]interface{}) error {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	widget, exists := cd.widgets[widgetID]
	if !exists {
		return fmt.Errorf("widget %s not found", widgetID)
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok {
		widget.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		widget.Description = description
	}
	if query, ok := updates["query"].(string); ok {
		widget.Query = query
	}
	if visible, ok := updates["is_visible"].(bool); ok {
		widget.IsVisible = visible
	}

	widget.UpdatedAt = time.Now()
	cd.updatedAt = time.Now()

	cd.logger.Info("Widget updated", "dashboard_id", cd.id, "widget_id", widgetID)
	return nil
}

// GetWidget returns a specific widget
func (cd *CustomDashboard) GetWidget(widgetID string) (*DashboardWidget, error) {
	cd.mutex.RLock()
	defer cd.mutex.RUnlock()

	widget, exists := cd.widgets[widgetID]
	if !exists {
		return nil, fmt.Errorf("widget %s not found", widgetID)
	}

	// Return a copy
	widgetCopy := *widget
	return &widgetCopy, nil
}

// GetWidgets returns all widgets
func (cd *CustomDashboard) GetWidgets() map[string]*DashboardWidget {
	cd.mutex.RLock()
	defer cd.mutex.RUnlock()

	widgets := make(map[string]*DashboardWidget)
	for id, widget := range cd.widgets {
		widgetCopy := *widget
		widgets[id] = &widgetCopy
	}

	return widgets
}

// AddDataProvider adds a data provider to the dashboard
func (cd *CustomDashboard) AddDataProvider(name string, provider DataProvider) {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	cd.dataProviders[name] = provider
	cd.logger.Info("Data provider added", "dashboard_id", cd.id, "provider", name)
}

// GetDataProvider returns a data provider
func (cd *CustomDashboard) GetDataProvider(name string) (DataProvider, error) {
	cd.mutex.RLock()
	defer cd.mutex.RUnlock()

	provider, exists := cd.dataProviders[name]
	if !exists {
		return nil, fmt.Errorf("data provider %s not found", name)
	}

	return provider, nil
}

// RefreshWidget refreshes data for a specific widget
func (cd *CustomDashboard) RefreshWidget(ctx context.Context, widgetID string) (*DataResult, error) {
	cd.mutex.RLock()
	widget, exists := cd.widgets[widgetID]
	if !exists {
		cd.mutex.RUnlock()
		return nil, fmt.Errorf("widget %s not found", widgetID)
	}

	provider, providerExists := cd.dataProviders[widget.DataSource]
	cd.mutex.RUnlock()

	if !providerExists {
		return nil, fmt.Errorf("data provider %s not found for widget %s", widget.DataSource, widgetID)
	}

	return provider.GetData(ctx, widget.Query, widget.Filters)
}

// RefreshAllWidgets refreshes data for all widgets
func (cd *CustomDashboard) RefreshAllWidgets(ctx context.Context) (map[string]*DataResult, error) {
	cd.mutex.RLock()
	widgets := make(map[string]*DashboardWidget)
	for id, widget := range cd.widgets {
		if widget.IsVisible {
			widgetCopy := *widget
			widgets[id] = &widgetCopy
		}
	}
	cd.mutex.RUnlock()

	results := make(map[string]*DataResult)
	for widgetID := range widgets {
		result, err := cd.RefreshWidget(ctx, widgetID)
		if err != nil {
			cd.logger.Error("Failed to refresh widget", err, "widget_id", widgetID)
			continue
		}
		results[widgetID] = result
	}

	return results, nil
}

// UpdateLayout updates the dashboard layout
func (cd *CustomDashboard) UpdateLayout(layout *DashboardLayout) {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	cd.layout = layout
	cd.updatedAt = time.Now()

	cd.logger.Info("Dashboard layout updated", "dashboard_id", cd.id)
}

// UpdateConfig updates the dashboard configuration
func (cd *CustomDashboard) UpdateConfig(config *DashboardConfig) {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	cd.config = config
	cd.updatedAt = time.Now()

	cd.logger.Info("Dashboard config updated", "dashboard_id", cd.id)
}

// AddAlertRule adds an alert rule to the dashboard
func (cd *CustomDashboard) AddAlertRule(rule *AlertRule) error {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	if rule.ID == "" {
		rule.ID = fmt.Sprintf("alert_%d", time.Now().UnixNano())
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	cd.alertRules[rule.ID] = rule
	cd.updatedAt = time.Now()

	cd.logger.Info("Alert rule added", "dashboard_id", cd.id, "rule_id", rule.ID)
	return nil
}

// RemoveAlertRule removes an alert rule
func (cd *CustomDashboard) RemoveAlertRule(ruleID string) error {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	if _, exists := cd.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule %s not found", ruleID)
	}

	delete(cd.alertRules, ruleID)
	cd.updatedAt = time.Now()

	cd.logger.Info("Alert rule removed", "dashboard_id", cd.id, "rule_id", ruleID)
	return nil
}

// GetDashboardInfo returns dashboard information
func (cd *CustomDashboard) GetDashboardInfo() *DashboardInfo {
	cd.mutex.RLock()
	defer cd.mutex.RUnlock()

	return &DashboardInfo{
		ID:              cd.id,
		Name:            cd.name,
		Description:     cd.description,
		Owner:           cd.owner,
		WidgetCount:     len(cd.widgets),
		AlertRuleCount:  len(cd.alertRules),
		IsActive:        cd.isActive,
		CreatedAt:       cd.createdAt,
		UpdatedAt:       cd.updatedAt,
		RefreshInterval: cd.refreshInterval,
		Layout:          cd.layout,
		Config:          cd.config,
		Permissions:     cd.permissions,
	}
}

// DashboardInfo represents dashboard metadata
type DashboardInfo struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Owner           string                `json:"owner"`
	WidgetCount     int                   `json:"widget_count"`
	AlertRuleCount  int                   `json:"alert_rule_count"`
	IsActive        bool                  `json:"is_active"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	RefreshInterval time.Duration         `json:"refresh_interval"`
	Layout          *DashboardLayout      `json:"layout"`
	Config          *DashboardConfig      `json:"config"`
	Permissions     *DashboardPermissions `json:"permissions"`
}

// ExportDashboard exports dashboard configuration
func (cd *CustomDashboard) ExportDashboard() ([]byte, error) {
	cd.mutex.RLock()
	defer cd.mutex.RUnlock()

	export := map[string]interface{}{
		"dashboard_info": cd.GetDashboardInfo(),
		"widgets":        cd.widgets,
		"alert_rules":    cd.alertRules,
		"exported_at":    time.Now(),
		"version":        "1.0",
	}

	return json.MarshalIndent(export, "", "  ")
}

// ImportDashboard imports dashboard configuration
func (cd *CustomDashboard) ImportDashboard(data []byte) error {
	var importData map[string]interface{}
	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to parse import data: %w", err)
	}

	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	// Import widgets
	if widgetsData, ok := importData["widgets"].(map[string]interface{}); ok {
		for widgetID, widgetData := range widgetsData {
			widgetBytes, _ := json.Marshal(widgetData)
			var widget DashboardWidget
			if err := json.Unmarshal(widgetBytes, &widget); err == nil {
				widget.ID = widgetID
				cd.widgets[widgetID] = &widget
			}
		}
	}

	// Import alert rules
	if alertsData, ok := importData["alert_rules"].(map[string]interface{}); ok {
		for ruleID, ruleData := range alertsData {
			ruleBytes, _ := json.Marshal(ruleData)
			var rule AlertRule
			if err := json.Unmarshal(ruleBytes, &rule); err == nil {
				rule.ID = ruleID
				cd.alertRules[ruleID] = &rule
			}
		}
	}

	cd.updatedAt = time.Now()
	cd.logger.Info("Dashboard imported", "dashboard_id", cd.id)

	return nil
}

// Clone creates a copy of the dashboard
func (cd *CustomDashboard) Clone(newID, newName, newOwner string) *CustomDashboard {
	cd.mutex.RLock()
	defer cd.mutex.RUnlock()

	clone := NewCustomDashboard(newID, newName, newOwner, cd.logger)
	clone.description = cd.description + " (Clone)"

	// Copy widgets
	for id, widget := range cd.widgets {
		widgetCopy := *widget
		widgetCopy.ID = fmt.Sprintf("%s_clone", id)
		clone.widgets[widgetCopy.ID] = &widgetCopy
	}

	// Copy layout and config
	layoutCopy := *cd.layout
	clone.layout = &layoutCopy

	configCopy := *cd.config
	clone.config = &configCopy

	// Copy alert rules
	for id, rule := range cd.alertRules {
		ruleCopy := *rule
		ruleCopy.ID = fmt.Sprintf("%s_clone", id)
		clone.alertRules[ruleCopy.ID] = &ruleCopy
	}

	cd.logger.Info("Dashboard cloned", "original_id", cd.id, "clone_id", newID)
	return clone
}
