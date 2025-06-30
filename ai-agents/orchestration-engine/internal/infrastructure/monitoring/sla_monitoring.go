package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SLAMonitor monitors Service Level Agreements and Service Level Indicators
type SLAMonitor struct {
	slis            map[string]*ServiceLevelIndicator
	slos            map[string]*ServiceLevelObjective
	alertRules      map[string]*SLAAlertRule
	incidentManager *IncidentManager
	config          *SLAConfig
	logger          Logger
	mutex           sync.RWMutex
	stopCh          chan struct{}
}

// ServiceLevelIndicator defines a measurable metric
type ServiceLevelIndicator struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	MetricType  SLIMetricType          `json:"metric_type"`
	Query       string                 `json:"query"`
	DataSource  string                 `json:"data_source"`
	Unit        string                 `json:"unit"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ServiceLevelObjective defines target performance levels
type ServiceLevelObjective struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	SLIID           string                 `json:"sli_id"`
	Target          float64                `json:"target"`
	TimeWindow      time.Duration          `json:"time_window"`
	BurnRateWindow  time.Duration          `json:"burn_rate_window"`
	ErrorBudget     float64                `json:"error_budget"`
	Severity        SLOSeverity            `json:"severity"`
	Owner           string                 `json:"owner"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// SLAAlertRule defines alerting rules for SLA violations
type SLAAlertRule struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	SLOID           string                 `json:"slo_id"`
	Condition       SLACondition           `json:"condition"`
	Threshold       float64                `json:"threshold"`
	Duration        time.Duration          `json:"duration"`
	Severity        AlertSeverity          `json:"severity"`
	Enabled         bool                   `json:"enabled"`
	Recipients      []string               `json:"recipients"`
	Actions         []AlertAction          `json:"actions"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// SLIMetricType represents different types of SLI metrics
type SLIMetricType string

const (
	SLIMetricTypeAvailability SLIMetricType = "availability"
	SLIMetricTypeLatency      SLIMetricType = "latency"
	SLIMetricTypeThroughput   SLIMetricType = "throughput"
	SLIMetricTypeErrorRate    SLIMetricType = "error_rate"
	SLIMetricTypeCustom       SLIMetricType = "custom"
)

// SLOSeverity represents SLO severity levels
type SLOSeverity string

const (
	SLOSeverityLow      SLOSeverity = "low"
	SLOSeverityMedium   SLOSeverity = "medium"
	SLOSeverityHigh     SLOSeverity = "high"
	SLOSeverityCritical SLOSeverity = "critical"
)

// SLACondition represents SLA alert conditions
type SLACondition string

const (
	SLAConditionErrorBudgetExhausted SLACondition = "error_budget_exhausted"
	SLAConditionBurnRateHigh         SLACondition = "burn_rate_high"
	SLAConditionSLOViolation         SLACondition = "slo_violation"
	SLAConditionCustom               SLACondition = "custom"
)

// SLAConfig contains SLA monitoring configuration
type SLAConfig struct {
	EvaluationInterval    time.Duration `json:"evaluation_interval"`
	DefaultTimeWindow     time.Duration `json:"default_time_window"`
	DefaultErrorBudget    float64       `json:"default_error_budget"`
	EnableIncidentManager bool          `json:"enable_incident_manager"`
	AlertChannels         []string      `json:"alert_channels"`
	RetentionPeriod       time.Duration `json:"retention_period"`
}

// IncidentManager manages incident response
type IncidentManager struct {
	incidents   map[string]*Incident
	escalations map[string]*EscalationPolicy
	logger      Logger
	mutex       sync.RWMutex
}

// Incident represents a service incident
type Incident struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    IncidentSeverity       `json:"severity"`
	Status      IncidentStatus         `json:"status"`
	SLOID       string                 `json:"slo_id"`
	AlertRuleID string                 `json:"alert_rule_id"`
	Assignee    string                 `json:"assignee"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// IncidentSeverity represents incident severity levels
type IncidentSeverity string

const (
	IncidentSeverityLow      IncidentSeverity = "low"
	IncidentSeverityMedium   IncidentSeverity = "medium"
	IncidentSeverityHigh     IncidentSeverity = "high"
	IncidentSeverityCritical IncidentSeverity = "critical"
)

// IncidentStatus represents incident status
type IncidentStatus string

const (
	IncidentStatusOpen       IncidentStatus = "open"
	IncidentStatusInProgress IncidentStatus = "in_progress"
	IncidentStatusResolved   IncidentStatus = "resolved"
	IncidentStatusClosed     IncidentStatus = "closed"
)

// EscalationPolicy defines incident escalation rules
type EscalationPolicy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Rules       []*EscalationRule      `json:"rules"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// EscalationRule defines escalation timing and targets
type EscalationRule struct {
	Level     int           `json:"level"`
	Delay     time.Duration `json:"delay"`
	Targets   []string      `json:"targets"`
	Actions   []string      `json:"actions"`
	Condition string        `json:"condition"`
}

// NewSLAMonitor creates a new SLA monitor
func NewSLAMonitor(config *SLAConfig, logger Logger) *SLAMonitor {
	if config == nil {
		config = DefaultSLAConfig()
	}

	monitor := &SLAMonitor{
		slis:       make(map[string]*ServiceLevelIndicator),
		slos:       make(map[string]*ServiceLevelObjective),
		alertRules: make(map[string]*SLAAlertRule),
		config:     config,
		logger:     logger,
		stopCh:     make(chan struct{}),
	}

	if config.EnableIncidentManager {
		monitor.incidentManager = NewIncidentManager(logger)
	}

	return monitor
}

// DefaultSLAConfig returns default SLA configuration
func DefaultSLAConfig() *SLAConfig {
	return &SLAConfig{
		EvaluationInterval:    1 * time.Minute,
		DefaultTimeWindow:     24 * time.Hour,
		DefaultErrorBudget:    0.1, // 10% error budget
		EnableIncidentManager: true,
		AlertChannels:         []string{"email", "slack"},
		RetentionPeriod:       30 * 24 * time.Hour, // 30 days
	}
}

// Start starts the SLA monitor
func (sm *SLAMonitor) Start(ctx context.Context) error {
	sm.logger.Info("Starting SLA monitor")

	// Start evaluation loop
	go sm.evaluationLoop(ctx)

	// Start incident manager if enabled
	if sm.incidentManager != nil {
		go sm.incidentManager.Start(ctx)
	}

	// Create default SLIs and SLOs
	sm.createDefaultSLIs()
	sm.createDefaultSLOs()

	sm.logger.Info("SLA monitor started")
	return nil
}

// Stop stops the SLA monitor
func (sm *SLAMonitor) Stop(ctx context.Context) error {
	sm.logger.Info("Stopping SLA monitor")
	
	close(sm.stopCh)
	
	if sm.incidentManager != nil {
		sm.incidentManager.Stop(ctx)
	}
	
	sm.logger.Info("SLA monitor stopped")
	return nil
}

// AddSLI adds a Service Level Indicator
func (sm *SLAMonitor) AddSLI(sli *ServiceLevelIndicator) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sli.ID == "" {
		sli.ID = fmt.Sprintf("sli_%d", time.Now().UnixNano())
	}

	sli.CreatedAt = time.Now()
	sli.UpdatedAt = time.Now()

	sm.slis[sli.ID] = sli
	sm.logger.Info("SLI added", "id", sli.ID, "name", sli.Name)

	return nil
}

// AddSLO adds a Service Level Objective
func (sm *SLAMonitor) AddSLO(slo *ServiceLevelObjective) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if slo.ID == "" {
		slo.ID = fmt.Sprintf("slo_%d", time.Now().UnixNano())
	}

	// Validate SLI exists
	if _, exists := sm.slis[slo.SLIID]; !exists {
		return fmt.Errorf("SLI %s not found", slo.SLIID)
	}

	slo.CreatedAt = time.Now()
	slo.UpdatedAt = time.Now()

	sm.slos[slo.ID] = slo
	sm.logger.Info("SLO added", "id", slo.ID, "name", slo.Name)

	return nil
}

// AddAlertRule adds an SLA alert rule
func (sm *SLAMonitor) AddAlertRule(rule *SLAAlertRule) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if rule.ID == "" {
		rule.ID = fmt.Sprintf("alert_rule_%d", time.Now().UnixNano())
	}

	// Validate SLO exists
	if _, exists := sm.slos[rule.SLOID]; !exists {
		return fmt.Errorf("SLO %s not found", rule.SLOID)
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	sm.alertRules[rule.ID] = rule
	sm.logger.Info("SLA alert rule added", "id", rule.ID, "name", rule.Name)

	return nil
}

// evaluationLoop runs the SLA evaluation loop
func (sm *SLAMonitor) evaluationLoop(ctx context.Context) {
	ticker := time.NewTicker(sm.config.EvaluationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopCh:
			return
		case <-ticker.C:
			sm.evaluateSLOs(ctx)
		}
	}
}

// evaluateSLOs evaluates all SLOs and triggers alerts if needed
func (sm *SLAMonitor) evaluateSLOs(ctx context.Context) {
	sm.mutex.RLock()
	slos := make([]*ServiceLevelObjective, 0, len(sm.slos))
	for _, slo := range sm.slos {
		slos = append(slos, slo)
	}
	sm.mutex.RUnlock()

	for _, slo := range slos {
		go sm.evaluateSLO(ctx, slo)
	}
}

// evaluateSLO evaluates a single SLO
func (sm *SLAMonitor) evaluateSLO(ctx context.Context, slo *ServiceLevelObjective) {
	sm.logger.Debug("Evaluating SLO", "id", slo.ID, "name", slo.Name)

	// Get SLI
	sm.mutex.RLock()
	sli, exists := sm.slis[slo.SLIID]
	sm.mutex.RUnlock()

	if !exists {
		sm.logger.Error("SLI not found for SLO", fmt.Errorf("SLI %s not found", slo.SLIID), "slo_id", slo.ID)
		return
	}

	// Calculate current SLI value (simplified implementation)
	currentValue := sm.calculateSLIValue(ctx, sli)
	
	// Check SLO compliance
	isViolating := sm.checkSLOViolation(slo, currentValue)
	
	if isViolating {
		sm.handleSLOViolation(ctx, slo, sli, currentValue)
	}
}

// calculateSLIValue calculates the current SLI value
func (sm *SLAMonitor) calculateSLIValue(ctx context.Context, sli *ServiceLevelIndicator) float64 {
	// This is a simplified implementation
	// In a real system, this would query the actual data source
	
	switch sli.MetricType {
	case SLIMetricTypeAvailability:
		return 99.5 // 99.5% availability
	case SLIMetricTypeLatency:
		return 150.0 // 150ms average latency
	case SLIMetricTypeThroughput:
		return 1000.0 // 1000 requests/second
	case SLIMetricTypeErrorRate:
		return 0.5 // 0.5% error rate
	default:
		return 0.0
	}
}

// checkSLOViolation checks if an SLO is being violated
func (sm *SLAMonitor) checkSLOViolation(slo *ServiceLevelObjective, currentValue float64) bool {
	// Simplified violation check
	switch slo.SLIID {
	case "availability_sli":
		return currentValue < slo.Target // Availability should be above target
	case "latency_sli":
		return currentValue > slo.Target // Latency should be below target
	case "error_rate_sli":
		return currentValue > slo.Target // Error rate should be below target
	default:
		return false
	}
}

// handleSLOViolation handles SLO violations
func (sm *SLAMonitor) handleSLOViolation(ctx context.Context, slo *ServiceLevelObjective, sli *ServiceLevelIndicator, currentValue float64) {
	sm.logger.Warn("SLO violation detected", 
		"slo_id", slo.ID, 
		"slo_name", slo.Name,
		"current_value", currentValue,
		"target", slo.Target,
	)

	// Check for applicable alert rules
	sm.mutex.RLock()
	alertRules := make([]*SLAAlertRule, 0)
	for _, rule := range sm.alertRules {
		if rule.SLOID == slo.ID && rule.Enabled {
			alertRules = append(alertRules, rule)
		}
	}
	sm.mutex.RUnlock()

	// Trigger alerts
	for _, rule := range alertRules {
		sm.triggerAlert(ctx, rule, slo, sli, currentValue)
	}

	// Create incident if incident manager is enabled
	if sm.incidentManager != nil {
		incident := &Incident{
			ID:          fmt.Sprintf("incident_%d", time.Now().UnixNano()),
			Title:       fmt.Sprintf("SLO Violation: %s", slo.Name),
			Description: fmt.Sprintf("SLO %s is violating target. Current: %.2f, Target: %.2f", slo.Name, currentValue, slo.Target),
			Severity:    sm.mapSLOSeverityToIncident(slo.Severity),
			Status:      IncidentStatusOpen,
			SLOID:       slo.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Metadata: map[string]interface{}{
				"current_value": currentValue,
				"target":        slo.Target,
				"sli_id":        sli.ID,
			},
		}

		sm.incidentManager.CreateIncident(incident)
	}
}

// triggerAlert triggers an alert for SLA violation
func (sm *SLAMonitor) triggerAlert(ctx context.Context, rule *SLAAlertRule, slo *ServiceLevelObjective, sli *ServiceLevelIndicator, currentValue float64) {
	alert := &Alert{
		ID:          fmt.Sprintf("sla_alert_%d", time.Now().UnixNano()),
		RuleID:      rule.ID,
		Title:       fmt.Sprintf("SLA Alert: %s", rule.Name),
		Message:     fmt.Sprintf("SLO %s violation detected. Current: %.2f, Target: %.2f", slo.Name, currentValue, slo.Target),
		Severity:    rule.Severity,
		Value:       currentValue,
		Threshold:   slo.Target,
		Timestamp:   time.Now(),
		Status:      AlertStatusFiring,
		Metadata: map[string]interface{}{
			"slo_id":        slo.ID,
			"sli_id":        sli.ID,
			"rule_id":       rule.ID,
			"current_value": currentValue,
			"target":        slo.Target,
		},
	}

	sm.logger.Info("SLA alert triggered", "alert_id", alert.ID, "rule_id", rule.ID)
	
	// In a real implementation, this would send the alert through notification channels
}

// createDefaultSLIs creates default Service Level Indicators
func (sm *SLAMonitor) createDefaultSLIs() {
	// Availability SLI
	availabilitySLI := &ServiceLevelIndicator{
		ID:          "availability_sli",
		Name:        "Service Availability",
		Description: "Percentage of successful requests",
		MetricType:  SLIMetricTypeAvailability,
		Query:       "sum(rate(http_requests_total{status!~'5..'}[5m])) / sum(rate(http_requests_total[5m])) * 100",
		DataSource:  "prometheus",
		Unit:        "%",
		Tags:        []string{"availability", "http"},
	}

	// Latency SLI
	latencySLI := &ServiceLevelIndicator{
		ID:          "latency_sli",
		Name:        "Response Time",
		Description: "95th percentile response time",
		MetricType:  SLIMetricTypeLatency,
		Query:       "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
		DataSource:  "prometheus",
		Unit:        "ms",
		Tags:        []string{"latency", "performance"},
	}

	// Error Rate SLI
	errorRateSLI := &ServiceLevelIndicator{
		ID:          "error_rate_sli",
		Name:        "Error Rate",
		Description: "Percentage of failed requests",
		MetricType:  SLIMetricTypeErrorRate,
		Query:       "sum(rate(http_requests_total{status=~'5..'}[5m])) / sum(rate(http_requests_total[5m])) * 100",
		DataSource:  "prometheus",
		Unit:        "%",
		Tags:        []string{"errors", "reliability"},
	}

	sm.AddSLI(availabilitySLI)
	sm.AddSLI(latencySLI)
	sm.AddSLI(errorRateSLI)
}

// createDefaultSLOs creates default Service Level Objectives
func (sm *SLAMonitor) createDefaultSLOs() {
	// Availability SLO
	availabilitySLO := &ServiceLevelObjective{
		ID:             "availability_slo",
		Name:           "Service Availability SLO",
		Description:    "Service should be available 99.9% of the time",
		SLIID:          "availability_sli",
		Target:         99.9,
		TimeWindow:     24 * time.Hour,
		BurnRateWindow: 1 * time.Hour,
		ErrorBudget:    0.1,
		Severity:       SLOSeverityCritical,
		Owner:          "platform-team",
		Tags:           []string{"availability", "critical"},
	}

	// Latency SLO
	latencySLO := &ServiceLevelObjective{
		ID:             "latency_slo",
		Name:           "Response Time SLO",
		Description:    "95% of requests should complete within 200ms",
		SLIID:          "latency_sli",
		Target:         200.0,
		TimeWindow:     1 * time.Hour,
		BurnRateWindow: 5 * time.Minute,
		ErrorBudget:    5.0,
		Severity:       SLOSeverityHigh,
		Owner:          "platform-team",
		Tags:           []string{"latency", "performance"},
	}

	// Error Rate SLO
	errorRateSLO := &ServiceLevelObjective{
		ID:             "error_rate_slo",
		Name:           "Error Rate SLO",
		Description:    "Error rate should be below 1%",
		SLIID:          "error_rate_sli",
		Target:         1.0,
		TimeWindow:     1 * time.Hour,
		BurnRateWindow: 5 * time.Minute,
		ErrorBudget:    1.0,
		Severity:       SLOSeverityHigh,
		Owner:          "platform-team",
		Tags:           []string{"errors", "reliability"},
	}

	sm.AddSLO(availabilitySLO)
	sm.AddSLO(latencySLO)
	sm.AddSLO(errorRateSLO)
}

// mapSLOSeverityToIncident maps SLO severity to incident severity
func (sm *SLAMonitor) mapSLOSeverityToIncident(sloSeverity SLOSeverity) IncidentSeverity {
	switch sloSeverity {
	case SLOSeverityLow:
		return IncidentSeverityLow
	case SLOSeverityMedium:
		return IncidentSeverityMedium
	case SLOSeverityHigh:
		return IncidentSeverityHigh
	case SLOSeverityCritical:
		return IncidentSeverityCritical
	default:
		return IncidentSeverityMedium
	}
}

// NewIncidentManager creates a new incident manager
func NewIncidentManager(logger Logger) *IncidentManager {
	return &IncidentManager{
		incidents:   make(map[string]*Incident),
		escalations: make(map[string]*EscalationPolicy),
		logger:      logger,
	}
}

// Start starts the incident manager
func (im *IncidentManager) Start(ctx context.Context) error {
	im.logger.Info("Starting incident manager")
	return nil
}

// Stop stops the incident manager
func (im *IncidentManager) Stop(ctx context.Context) error {
	im.logger.Info("Stopping incident manager")
	return nil
}

// CreateIncident creates a new incident
func (im *IncidentManager) CreateIncident(incident *Incident) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.incidents[incident.ID] = incident
	im.logger.Info("Incident created", "id", incident.ID, "title", incident.Title, "severity", incident.Severity)

	return nil
}

// GetSLAStats returns SLA monitoring statistics
func (sm *SLAMonitor) GetSLAStats() *SLAStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	activeSLOs := 0
	violatingSLOs := 0
	for _ = range sm.slos {
		activeSLOs++
		// In a real implementation, this would check current violation status
	}

	return &SLAStats{
		TotalSLIs:     len(sm.slis),
		TotalSLOs:     len(sm.slos),
		ActiveSLOs:    activeSLOs,
		ViolatingSLOs: violatingSLOs,
		AlertRules:    len(sm.alertRules),
		LastEvaluation: time.Now(),
	}
}

// SLAStats represents SLA monitoring statistics
type SLAStats struct {
	TotalSLIs      int       `json:"total_slis"`
	TotalSLOs      int       `json:"total_slos"`
	ActiveSLOs     int       `json:"active_slos"`
	ViolatingSLOs  int       `json:"violating_slos"`
	AlertRules     int       `json:"alert_rules"`
	LastEvaluation time.Time `json:"last_evaluation"`
}
