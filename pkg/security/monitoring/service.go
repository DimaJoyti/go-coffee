package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// SecurityMonitoringService provides security monitoring and threat detection
type SecurityMonitoringService struct {
	config         *Config
	logger         *logger.Logger
	eventStore     EventStore
	alertManager   AlertManager
	threatDetector ThreatDetector
	metrics        *SecurityMetrics
	mu             sync.RWMutex
}

// Config represents monitoring configuration
type Config struct {
	EnableRealTimeMonitoring bool            `yaml:"enable_real_time_monitoring" env:"ENABLE_REAL_TIME_MONITORING" default:"true"`
	AlertThresholds          AlertThresholds `yaml:"alert_thresholds"`
	RetentionPeriod          time.Duration   `yaml:"retention_period" env:"RETENTION_PERIOD" default:"720h"`
	MaxEventsPerMinute       int             `yaml:"max_events_per_minute" env:"MAX_EVENTS_PER_MINUTE" default:"1000"`
	EnableThreatIntelligence bool            `yaml:"enable_threat_intelligence" env:"ENABLE_THREAT_INTELLIGENCE" default:"true"`
}

// AlertThresholds defines thresholds for different types of alerts
type AlertThresholds struct {
	FailedLoginAttempts    int           `yaml:"failed_login_attempts" default:"5"`
	SuspiciousIPRequests   int           `yaml:"suspicious_ip_requests" default:"100"`
	HighRiskEvents         int           `yaml:"high_risk_events" default:"10"`
	TimeWindow             time.Duration `yaml:"time_window" default:"5m"`
	CriticalEventThreshold int           `yaml:"critical_event_threshold" default:"1"`
}

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	EventTypeAuthentication      SecurityEventType = "authentication"
	EventTypeAuthorization       SecurityEventType = "authorization"
	EventTypeDataAccess          SecurityEventType = "data_access"
	EventTypeSystemAccess        SecurityEventType = "system_access"
	EventTypeNetworkActivity     SecurityEventType = "network_activity"
	EventTypeMaliciousActivity   SecurityEventType = "malicious_activity"
	EventTypeConfigChange        SecurityEventType = "config_change"
	EventTypePrivilegeEscalation SecurityEventType = "privilege_escalation"
)

// SecuritySeverity represents the severity of a security event
type SecuritySeverity string

const (
	SeverityInfo     SecuritySeverity = "info"
	SeverityLow      SecuritySeverity = "low"
	SeverityMedium   SecuritySeverity = "medium"
	SeverityHigh     SecuritySeverity = "high"
	SeverityCritical SecuritySeverity = "critical"
)

// SecurityEvent represents a security event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   SecurityEventType      `json:"event_type"`
	Severity    SecuritySeverity       `json:"severity"`
	Source      string                 `json:"source"`
	UserID      string                 `json:"user_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ThreatLevel ThreatLevel            `json:"threat_level"`
	Mitigated   bool                   `json:"mitigated"`
}

// ThreatLevel represents the threat level
type ThreatLevel string

const (
	ThreatLevelNone     ThreatLevel = "none"
	ThreatLevelLow      ThreatLevel = "low"
	ThreatLevelMedium   ThreatLevel = "medium"
	ThreatLevelHigh     ThreatLevel = "high"
	ThreatLevelCritical ThreatLevel = "critical"
)

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	AlertType   AlertType              `json:"alert_type"`
	Severity    SecuritySeverity       `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Events      []SecurityEvent        `json:"events"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      AlertStatus            `json:"status"`
	AssignedTo  string                 `json:"assigned_to,omitempty"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeThresholdExceeded AlertType = "threshold_exceeded"
	AlertTypeAnomalyDetected   AlertType = "anomaly_detected"
	AlertTypeThreatDetected    AlertType = "threat_detected"
	AlertTypeSystemCompromise  AlertType = "system_compromise"
	AlertTypeDataBreach        AlertType = "data_breach"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusOpen          AlertStatus = "open"
	AlertStatusInvestigating AlertStatus = "investigating"
	AlertStatusResolved      AlertStatus = "resolved"
	AlertStatusFalsePositive AlertStatus = "false_positive"
)

// SecurityMetrics holds security-related metrics
type SecurityMetrics struct {
	TotalEvents      int64            `json:"total_events"`
	EventsByType     map[string]int64 `json:"events_by_type"`
	EventsBySeverity map[string]int64 `json:"events_by_severity"`
	ActiveAlerts     int64            `json:"active_alerts"`
	ResolvedAlerts   int64            `json:"resolved_alerts"`
	ThreatDetections int64            `json:"threat_detections"`
	BlockedRequests  int64            `json:"blocked_requests"`
	LastUpdated      time.Time        `json:"last_updated"`
}

// EventStore interface for storing security events
type EventStore interface {
	Store(ctx context.Context, event *SecurityEvent) error
	Query(ctx context.Context, filter EventFilter) ([]*SecurityEvent, error)
	Delete(ctx context.Context, olderThan time.Time) error
}

// AlertManager interface for managing alerts
type AlertManager interface {
	CreateAlert(ctx context.Context, alert *SecurityAlert) error
	UpdateAlert(ctx context.Context, alertID string, updates map[string]interface{}) error
	GetActiveAlerts(ctx context.Context) ([]*SecurityAlert, error)
	ResolveAlert(ctx context.Context, alertID string, reason string) error
}

// ThreatDetector interface for threat detection
type ThreatDetector interface {
	AnalyzeEvent(ctx context.Context, event *SecurityEvent) (*ThreatAnalysis, error)
	UpdateThreatIntelligence(ctx context.Context) error
	GetThreatLevel(ctx context.Context, indicators []string) (ThreatLevel, error)
}

// EventFilter represents filters for querying events
type EventFilter struct {
	StartTime   *time.Time          `json:"start_time,omitempty"`
	EndTime     *time.Time          `json:"end_time,omitempty"`
	EventTypes  []SecurityEventType `json:"event_types,omitempty"`
	Severities  []SecuritySeverity  `json:"severities,omitempty"`
	UserID      string              `json:"user_id,omitempty"`
	IPAddress   string              `json:"ip_address,omitempty"`
	ThreatLevel *ThreatLevel        `json:"threat_level,omitempty"`
	Limit       int                 `json:"limit,omitempty"`
}

// ThreatAnalysis represents the result of threat analysis
type ThreatAnalysis struct {
	ThreatLevel     ThreatLevel            `json:"threat_level"`
	Confidence      float64                `json:"confidence"`
	Indicators      []string               `json:"indicators"`
	Recommendations []string               `json:"recommendations"`
	ShouldBlock     bool                   `json:"should_block"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewSecurityMonitoringService creates a new security monitoring service
func NewSecurityMonitoringService(
	config *Config,
	logger *logger.Logger,
	eventStore EventStore,
	alertManager AlertManager,
	threatDetector ThreatDetector,
) *SecurityMonitoringService {
	return &SecurityMonitoringService{
		config:         config,
		logger:         logger,
		eventStore:     eventStore,
		alertManager:   alertManager,
		threatDetector: threatDetector,
		metrics: &SecurityMetrics{
			EventsByType:     make(map[string]int64),
			EventsBySeverity: make(map[string]int64),
			LastUpdated:      time.Now(),
		},
	}
}

// LogSecurityEvent logs a security event
func (s *SecurityMonitoringService) LogSecurityEvent(ctx context.Context, event *SecurityEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
	}

	// Store the event
	if err := s.eventStore.Store(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to store security event")
		return fmt.Errorf("failed to store security event: %w", err)
	}

	// Update metrics
	s.updateMetrics(event)

	// Log the event
	s.logEvent(event)

	// Analyze for threats if enabled
	if s.config.EnableThreatIntelligence {
		go s.analyzeEventForThreats(ctx, event)
	}

	// Check for alert conditions
	if s.config.EnableRealTimeMonitoring {
		go s.checkAlertConditions(ctx, event)
	}

	return nil
}

// GetSecurityMetrics returns current security metrics
func (s *SecurityMonitoringService) GetSecurityMetrics(ctx context.Context) *SecurityMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &SecurityMetrics{
		TotalEvents:      s.metrics.TotalEvents,
		EventsByType:     make(map[string]int64),
		EventsBySeverity: make(map[string]int64),
		ActiveAlerts:     s.metrics.ActiveAlerts,
		ResolvedAlerts:   s.metrics.ResolvedAlerts,
		ThreatDetections: s.metrics.ThreatDetections,
		BlockedRequests:  s.metrics.BlockedRequests,
		LastUpdated:      time.Now(),
	}

	for k, v := range s.metrics.EventsByType {
		metrics.EventsByType[k] = v
	}

	for k, v := range s.metrics.EventsBySeverity {
		metrics.EventsBySeverity[k] = v
	}

	return metrics
}

// QueryEvents queries security events based on filter
func (s *SecurityMonitoringService) QueryEvents(ctx context.Context, filter EventFilter) ([]*SecurityEvent, error) {
	return s.eventStore.Query(ctx, filter)
}

// CreateAlert creates a new security alert
func (s *SecurityMonitoringService) CreateAlert(ctx context.Context, alert *SecurityAlert) error {
	if alert.ID == "" {
		alert.ID = fmt.Sprintf("alert_%d", time.Now().UnixNano())
	}

	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}

	if alert.Status == "" {
		alert.Status = AlertStatusOpen
	}

	if err := s.alertManager.CreateAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	s.mu.Lock()
	s.metrics.ActiveAlerts++
	s.mu.Unlock()

	s.logger.WithFields(map[string]any{
		"alert_id":   alert.ID,
		"alert_type": alert.AlertType,
		"severity":   alert.Severity,
		"title":      alert.Title,
	}).Warn("Security alert created")

	return nil
}

// Helper methods

func (s *SecurityMonitoringService) updateMetrics(event *SecurityEvent) {
	s.metrics.TotalEvents++
	s.metrics.EventsByType[string(event.EventType)]++
	s.metrics.EventsBySeverity[string(event.Severity)]++

	if event.ThreatLevel != ThreatLevelNone {
		s.metrics.ThreatDetections++
	}

	s.metrics.LastUpdated = time.Now()
}

func (s *SecurityMonitoringService) logEvent(event *SecurityEvent) {
	eventJSON, _ := json.Marshal(event)

	switch event.Severity {
	case SeverityCritical:
		s.logger.Error("Critical security event", map[string]any{
			"event": string(eventJSON),
		})
	case SeverityHigh:
		s.logger.Warn("High severity security event", map[string]any{
			"event": string(eventJSON),
		})
	case SeverityMedium:
		s.logger.Info("Medium severity security event", map[string]any{
			"event": string(eventJSON),
		})
	default:
		s.logger.Debug("Security event", map[string]any{
			"event": string(eventJSON),
		})
	}
}

func (s *SecurityMonitoringService) analyzeEventForThreats(ctx context.Context, event *SecurityEvent) {
	analysis, err := s.threatDetector.AnalyzeEvent(ctx, event)
	if err != nil {
		s.logger.WithError(err).Error("Failed to analyze event for threats")
		return
	}

	if analysis.ThreatLevel >= ThreatLevelHigh {
		alert := &SecurityAlert{
			AlertType:   AlertTypeThreatDetected,
			Severity:    SeverityHigh,
			Title:       "High-level threat detected",
			Description: fmt.Sprintf("Threat analysis detected %s level threat", analysis.ThreatLevel),
			Events:      []SecurityEvent{*event},
			Metadata: map[string]interface{}{
				"threat_analysis": analysis,
			},
		}

		if analysis.ThreatLevel == ThreatLevelCritical {
			alert.Severity = SeverityCritical
			alert.Title = "Critical threat detected"
		}

		s.CreateAlert(ctx, alert)
	}
}

func (s *SecurityMonitoringService) checkAlertConditions(ctx context.Context, event *SecurityEvent) {
	// Check for threshold-based alerts
	s.checkThresholdAlerts(ctx, event)

	// Check for anomaly-based alerts
	s.checkAnomalyAlerts(ctx, event)
}

func (s *SecurityMonitoringService) checkThresholdAlerts(ctx context.Context, event *SecurityEvent) {
	// Implementation for threshold-based alerting
	// This would check various thresholds and create alerts accordingly
}

func (s *SecurityMonitoringService) checkAnomalyAlerts(ctx context.Context, event *SecurityEvent) {
	// Implementation for anomaly-based alerting
	// This would use machine learning or statistical methods to detect anomalies
}
