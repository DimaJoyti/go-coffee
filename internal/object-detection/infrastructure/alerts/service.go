package alerts

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service implements the alert service
type Service struct {
	logger               *zap.Logger
	repository           domain.AlertRepository
	notificationService  domain.AlertNotificationService
	config               ServiceConfig

	// In-memory tracking for performance
	ruleCache            map[string]*domain.AlertRule
	activeAlerts         map[string]*domain.Alert
	alertStatistics      map[string]*domain.AlertStatistics
	mutex                sync.RWMutex
}

// Ensure Service implements domain.AlertService
var _ domain.AlertService = (*Service)(nil)

// ServiceConfig configures the alert service
type ServiceConfig struct {
	EnableCaching          bool                              `yaml:"enable_caching"`
	CacheRefreshInterval   time.Duration                     `yaml:"cache_refresh_interval"`
	MaxActiveAlerts        int                               `yaml:"max_active_alerts"`
	DefaultCooldown        time.Duration                     `yaml:"default_cooldown"`
	ExpirationRules        map[domain.AlertType]time.Duration `yaml:"expiration_rules"`
	RetryAttempts          int                               `yaml:"retry_attempts"`
	RetryDelay             time.Duration                     `yaml:"retry_delay"`
	EnableBatching         bool                              `yaml:"enable_batching"`
	BatchSize              int                               `yaml:"batch_size"`
	BatchTimeout           time.Duration                     `yaml:"batch_timeout"`
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		EnableCaching:        true,
		CacheRefreshInterval: 30 * time.Second,
		MaxActiveAlerts:      1000,
		DefaultCooldown:      5 * time.Minute,
		ExpirationRules: map[domain.AlertType]time.Duration{
			domain.AlertTypeZoneViolation:     24 * time.Hour,
			domain.AlertTypeLoitering:         12 * time.Hour,
			domain.AlertTypeCrowding:          6 * time.Hour,
			domain.AlertTypeUnauthorizedAccess: 48 * time.Hour,
			domain.AlertTypeSystemError:       1 * time.Hour,
		},
		RetryAttempts:  3,
		RetryDelay:     30 * time.Second,
		EnableBatching: true,
		BatchSize:      10,
		BatchTimeout:   5 * time.Second,
	}
}

// NewService creates a new alert service
func NewService(logger *zap.Logger, repository domain.AlertRepository, notificationService domain.AlertNotificationService, config ServiceConfig) *Service {
	return &Service{
		logger:              logger.With(zap.String("component", "alert_service")),
		repository:          repository,
		notificationService: notificationService,
		config:              config,
		ruleCache:           make(map[string]*domain.AlertRule),
		activeAlerts:        make(map[string]*domain.Alert),
		alertStatistics:     make(map[string]*domain.AlertStatistics),
	}
}

// CreateAlert creates a new alert
func (s *Service) CreateAlert(alert *domain.Alert) error {
	// Validate alert
	if err := alert.Validate(); err != nil {
		return fmt.Errorf("invalid alert: %w", err)
	}

	// Set timestamps
	now := time.Now()
	alert.CreatedAt = now
	alert.UpdatedAt = now

	// Generate ID if not provided
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	// Set default status
	if alert.Status == "" {
		alert.Status = domain.AlertStatusActive
	}

	// Create in repository
	if err := s.repository.CreateAlert(alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// Update cache
	s.mutex.Lock()
	s.activeAlerts[alert.ID] = alert
	s.mutex.Unlock()

	// Update statistics
	s.updateAlertStatistics(alert.StreamID, alert.ZoneID, alert.Type, alert.Severity)

	s.logger.Info("Alert created",
		zap.String("alert_id", alert.ID),
		zap.String("type", string(alert.Type)),
		zap.String("severity", string(alert.Severity)),
		zap.String("stream_id", alert.StreamID))

	return nil
}

// GetAlert retrieves an alert by ID
func (s *Service) GetAlert(id string) (*domain.Alert, error) {
	// Check cache first
	if s.config.EnableCaching {
		s.mutex.RLock()
		if alert, exists := s.activeAlerts[id]; exists {
			s.mutex.RUnlock()
			return alert, nil
		}
		s.mutex.RUnlock()
	}

	// Get from repository
	return s.repository.GetAlert(id)
}

// GetAlerts retrieves alerts with filters
func (s *Service) GetAlerts(filters domain.AlertFilters) ([]*domain.Alert, error) {
	return s.repository.GetAlerts(filters)
}

// UpdateAlert updates an existing alert
func (s *Service) UpdateAlert(alert *domain.Alert) error {
	// Validate alert
	if err := alert.Validate(); err != nil {
		return fmt.Errorf("invalid alert: %w", err)
	}

	// Set update timestamp
	alert.UpdatedAt = time.Now()

	// Update in repository
	if err := s.repository.UpdateAlert(alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	// Update cache
	s.mutex.Lock()
	s.activeAlerts[alert.ID] = alert
	s.mutex.Unlock()

	s.logger.Info("Alert updated", zap.String("alert_id", alert.ID))
	return nil
}

// ResolveAlert resolves an alert
func (s *Service) ResolveAlert(id, resolvedBy string) error {
	alert, err := s.GetAlert(id)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	now := time.Now()
	alert.Status = domain.AlertStatusResolved
	alert.ResolvedAt = &now
	alert.ResolvedBy = resolvedBy
	alert.UpdatedAt = now

	if err := s.repository.UpdateAlert(alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	// Remove from active alerts cache
	s.mutex.Lock()
	delete(s.activeAlerts, id)
	s.mutex.Unlock()

	s.logger.Info("Alert resolved",
		zap.String("alert_id", id),
		zap.String("resolved_by", resolvedBy))

	return nil
}

// AcknowledgeAlert acknowledges an alert
func (s *Service) AcknowledgeAlert(id, acknowledgedBy string) error {
	alert, err := s.GetAlert(id)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	alert.Status = domain.AlertStatusAcknowledged
	alert.UpdatedAt = time.Now()
	
	// Add acknowledgment to metadata
	if alert.Metadata == nil {
		alert.Metadata = make(map[string]interface{})
	}
	alert.Metadata["acknowledged_by"] = acknowledgedBy
	alert.Metadata["acknowledged_at"] = time.Now()

	if err := s.repository.UpdateAlert(alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	// Update cache
	s.mutex.Lock()
	s.activeAlerts[alert.ID] = alert
	s.mutex.Unlock()

	s.logger.Info("Alert acknowledged",
		zap.String("alert_id", id),
		zap.String("acknowledged_by", acknowledgedBy))

	return nil
}

// SuppressAlert suppresses an alert for a specified duration
func (s *Service) SuppressAlert(id string, duration time.Duration) error {
	alert, err := s.GetAlert(id)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	alert.Status = domain.AlertStatusSuppressed
	alert.UpdatedAt = time.Now()
	
	// Add suppression info to metadata
	if alert.Metadata == nil {
		alert.Metadata = make(map[string]interface{})
	}
	alert.Metadata["suppressed_until"] = time.Now().Add(duration)
	alert.Metadata["suppression_duration"] = duration.String()

	if err := s.repository.UpdateAlert(alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	// Update cache
	s.mutex.Lock()
	s.activeAlerts[alert.ID] = alert
	s.mutex.Unlock()

	s.logger.Info("Alert suppressed",
		zap.String("alert_id", id),
		zap.Duration("duration", duration))

	return nil
}

// CreateAlertRule creates a new alert rule
func (s *Service) CreateAlertRule(rule *domain.AlertRule) error {
	// Validate rule
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("invalid alert rule: %w", err)
	}

	// Set timestamps
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	// Generate ID if not provided
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	// Set default cooldown
	if rule.Cooldown == 0 {
		rule.Cooldown = s.config.DefaultCooldown
	}

	// Create in repository
	if err := s.repository.CreateAlertRule(rule); err != nil {
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	// Update cache
	s.mutex.Lock()
	s.ruleCache[rule.ID] = rule
	s.mutex.Unlock()

	s.logger.Info("Alert rule created",
		zap.String("rule_id", rule.ID),
		zap.String("name", rule.Name),
		zap.String("stream_id", rule.StreamID))

	return nil
}

// GetAlertRule retrieves an alert rule by ID
func (s *Service) GetAlertRule(id string) (*domain.AlertRule, error) {
	// Check cache first
	if s.config.EnableCaching {
		s.mutex.RLock()
		if rule, exists := s.ruleCache[id]; exists {
			s.mutex.RUnlock()
			return rule, nil
		}
		s.mutex.RUnlock()
	}

	// Get from repository
	return s.repository.GetAlertRule(id)
}

// GetAlertRules retrieves alert rules for a stream
func (s *Service) GetAlertRules(streamID string) ([]*domain.AlertRule, error) {
	rules, err := s.repository.GetAlertRules(streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rules: %w", err)
	}

	// Update cache
	if s.config.EnableCaching {
		s.mutex.Lock()
		for _, rule := range rules {
			s.ruleCache[rule.ID] = rule
		}
		s.mutex.Unlock()
	}

	return rules, nil
}

// UpdateAlertRule updates an existing alert rule
func (s *Service) UpdateAlertRule(rule *domain.AlertRule) error {
	// Validate rule
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("invalid alert rule: %w", err)
	}

	// Set update timestamp
	rule.UpdatedAt = time.Now()

	// Update in repository
	if err := s.repository.UpdateAlertRule(rule); err != nil {
		return fmt.Errorf("failed to update alert rule: %w", err)
	}

	// Update cache
	s.mutex.Lock()
	s.ruleCache[rule.ID] = rule
	s.mutex.Unlock()

	s.logger.Info("Alert rule updated", zap.String("rule_id", rule.ID))
	return nil
}

// DeleteAlertRule deletes an alert rule
func (s *Service) DeleteAlertRule(id string) error {
	// Delete from repository
	if err := s.repository.DeleteAlertRule(id); err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	// Remove from cache
	s.mutex.Lock()
	delete(s.ruleCache, id)
	s.mutex.Unlock()

	s.logger.Info("Alert rule deleted", zap.String("rule_id", id))
	return nil
}

// ProcessZoneEvent processes a zone event and generates alerts
func (s *Service) ProcessZoneEvent(event *domain.ZoneEvent) ([]*domain.Alert, error) {
	context := domain.AlertContext{
		StreamID:   event.StreamID,
		ZoneID:     event.ZoneID,
		ZoneEvents: []*domain.ZoneEvent{event},
		Timestamp:  event.Timestamp,
		Metadata: map[string]interface{}{
			"event_type": string(event.EventType),
			"object_id":  event.ObjectID,
			"object_class": event.ObjectClass,
		},
	}

	return s.EvaluateAlertRules(event.StreamID, context)
}

// ProcessDetection processes a detection and generates alerts
func (s *Service) ProcessDetection(detection *domain.DetectedObject) ([]*domain.Alert, error) {
	context := domain.AlertContext{
		StreamID:   detection.StreamID,
		Detections: []*domain.DetectedObject{detection},
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"detection_id": detection.ID,
			"object_class": detection.Class,
			"confidence":   detection.Confidence,
		},
	}

	return s.EvaluateAlertRules(detection.StreamID, context)
}

// EvaluateAlertRules evaluates alert rules for a given context
func (s *Service) EvaluateAlertRules(streamID string, context domain.AlertContext) ([]*domain.Alert, error) {
	// Get alert rules for the stream
	rules, err := s.GetAlertRules(streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rules: %w", err)
	}

	var alerts []*domain.Alert

	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}

		// Check if rule can be triggered (cooldown)
		if !rule.CanTrigger() {
			continue
		}

		// Check if conditions match
		if !rule.Conditions.MatchesConditions(context) {
			continue
		}

		// Create alert
		alert := &domain.Alert{
			ID:          uuid.New().String(),
			Type:        rule.Type,
			Severity:    rule.Severity,
			Title:       s.generateAlertTitle(rule, context),
			Message:     s.generateAlertMessage(rule, context),
			StreamID:    streamID,
			ZoneID:      context.ZoneID,
			Status:      domain.AlertStatusActive,
			Metadata:    context.Metadata,
			Tags:        rule.Tags,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Add detection-specific fields if available
		if len(context.Detections) > 0 {
			detection := context.Detections[0] // Use first detection
			alert.ObjectID = detection.ID
			alert.ObjectClass = detection.Class
			alert.Confidence = detection.Confidence
			alert.Position = &domain.Point{
				X: float64(detection.BoundingBox.X + detection.BoundingBox.Width/2),
				Y: float64(detection.BoundingBox.Y + detection.BoundingBox.Height/2),
			}
			alert.BoundingBox = &detection.BoundingBox
		}

		// Create the alert
		if err := s.CreateAlert(alert); err != nil {
			s.logger.Error("Failed to create alert",
				zap.String("rule_id", rule.ID),
				zap.Error(err))
			continue
		}

		alerts = append(alerts, alert)

		// Update rule trigger info
		now := time.Now()
		rule.LastTriggered = &now
		rule.TriggerCount++
		if err := s.repository.UpdateAlertRule(rule); err != nil {
			s.logger.Error("Failed to update rule trigger info",
				zap.String("rule_id", rule.ID),
				zap.Error(err))
		}

		// Execute alert actions
		for _, action := range rule.Actions {
			if action.IsEnabled {
				if err := s.SendNotification(alert, action); err != nil {
					s.logger.Error("Failed to send notification",
						zap.String("alert_id", alert.ID),
						zap.String("action_type", string(action.Type)),
						zap.Error(err))
				}
			}
		}

		s.logger.Info("Alert triggered",
			zap.String("alert_id", alert.ID),
			zap.String("rule_id", rule.ID),
			zap.String("type", string(alert.Type)),
			zap.String("severity", string(alert.Severity)))
	}

	return alerts, nil
}

// SendNotification sends a notification for an alert
func (s *Service) SendNotification(alert *domain.Alert, action domain.AlertAction) error {
	// Create notification record
	notification := &domain.Notification{
		ID:        uuid.New().String(),
		AlertID:   alert.ID,
		Type:      action.Type,
		Status:    domain.NotificationStatusPending,
		Metadata:  action.Config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Get alert template and render content
	template, err := s.GetAlertTemplate(alert.Type, "en")
	if err != nil {
		s.logger.Warn("Failed to get alert template, using default",
			zap.String("alert_type", string(alert.Type)),
			zap.Error(err))
	}

	if template != nil {
		subject, content, err := s.RenderAlert(alert, template)
		if err != nil {
			s.logger.Error("Failed to render alert template", zap.Error(err))
		} else {
			notification.Subject = subject
			notification.Content = content
		}
	}

	// Set default content if template rendering failed
	if notification.Subject == "" {
		notification.Subject = alert.Title
	}
	if notification.Content == "" {
		notification.Content = alert.Message
	}

	// Extract recipient from action config
	if recipient, ok := action.Config["recipient"].(string); ok {
		notification.Recipient = recipient
	}

	// Create notification in repository
	if err := s.repository.CreateNotification(notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification based on type
	notification.Status = domain.NotificationStatusSending
	notification.UpdatedAt = time.Now()

	var sendErr error
	switch action.Type {
	case domain.AlertActionEmail:
		sendErr = s.notificationService.SendEmail(notification)
	case domain.AlertActionSMS:
		sendErr = s.notificationService.SendSMS(notification)
	case domain.AlertActionWebhook:
		sendErr = s.notificationService.SendWebhook(notification)
	case domain.AlertActionSlack:
		sendErr = s.notificationService.SendSlack(notification)
	case domain.AlertActionDiscord:
		sendErr = s.notificationService.SendDiscord(notification)
	case domain.AlertActionTelegram:
		sendErr = s.notificationService.SendTelegram(notification)
	case domain.AlertActionPushover:
		sendErr = s.notificationService.SendPushover(notification)
	case domain.AlertActionCustom:
		sendErr = s.notificationService.SendCustom(notification)
	default:
		sendErr = fmt.Errorf("unsupported notification type: %s", action.Type)
	}

	// Update notification status
	if sendErr != nil {
		notification.Status = domain.NotificationStatusFailed
		notification.Error = sendErr.Error()
		notification.FailedAt = &notification.UpdatedAt
	} else {
		notification.Status = domain.NotificationStatusSent
		notification.SentAt = &notification.UpdatedAt
	}

	notification.UpdatedAt = time.Now()
	if err := s.repository.UpdateNotification(notification); err != nil {
		s.logger.Error("Failed to update notification status", zap.Error(err))
	}

	return sendErr
}

// GetNotifications retrieves notifications for an alert
func (s *Service) GetNotifications(alertID string) ([]*domain.Notification, error) {
	return s.repository.GetNotifications(alertID)
}

// RetryFailedNotifications retries failed notifications
func (s *Service) RetryFailedNotifications() error {
	// This would typically query for failed notifications and retry them
	// Implementation depends on specific requirements
	s.logger.Info("Retrying failed notifications")
	return nil
}

// RenderAlert renders an alert using a template
func (s *Service) RenderAlert(alert *domain.Alert, template *domain.AlertTemplate) (string, string, error) {
	// Simple template rendering - in production, use a proper template engine
	subject := template.Subject
	body := template.Body

	// Replace common variables
	replacements := map[string]string{
		"{{.AlertID}}":     alert.ID,
		"{{.AlertType}}":   string(alert.Type),
		"{{.Severity}}":    string(alert.Severity),
		"{{.Title}}":       alert.Title,
		"{{.Message}}":     alert.Message,
		"{{.StreamID}}":    alert.StreamID,
		"{{.ZoneID}}":      alert.ZoneID,
		"{{.ObjectClass}}": alert.ObjectClass,
		"{{.Timestamp}}":   alert.CreatedAt.Format(time.RFC3339),
	}

	for placeholder, value := range replacements {
		subject = strings.ReplaceAll(subject, placeholder, value)
		body = strings.ReplaceAll(body, placeholder, value)
	}

	return subject, body, nil
}

// GetAlertTemplate retrieves an alert template
func (s *Service) GetAlertTemplate(alertType domain.AlertType, language string) (*domain.AlertTemplate, error) {
	templates, err := s.repository.GetAlertTemplates(alertType, language)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert templates: %w", err)
	}

	// Return the first default template or any template if no default
	for _, template := range templates {
		if template.IsDefault {
			return template, nil
		}
	}

	if len(templates) > 0 {
		return templates[0], nil
	}

	return nil, fmt.Errorf("no template found for alert type %s and language %s", alertType, language)
}

// GetAlertStatistics retrieves alert statistics
func (s *Service) GetAlertStatistics(streamID, zoneID string) (*domain.AlertStatistics, error) {
	key := streamID
	if zoneID != "" {
		key = streamID + ":" + zoneID
	}

	s.mutex.RLock()
	stats, exists := s.alertStatistics[key]
	s.mutex.RUnlock()

	if exists {
		return stats, nil
	}

	// Get from repository
	return s.repository.GetAlertStatistics(streamID, zoneID)
}

// GenerateAlertReport generates an alert report
func (s *Service) GenerateAlertReport(filters domain.AlertFilters, reportType domain.ReportType) (*domain.AlertReport, error) {
	alerts, err := s.repository.GetAlerts(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	report := &domain.AlertReport{
		ID:          uuid.New().String(),
		ReportType:  reportType,
		Filters:     filters,
		GeneratedAt: time.Now(),
		Data:        make(map[string]interface{}),
	}

	// Generate report data based on type
	switch reportType {
	case domain.ReportTypeSummary:
		report.Data["total_alerts"] = len(alerts)
		report.Data["alerts_by_type"] = s.groupAlertsByType(alerts)
		report.Data["alerts_by_severity"] = s.groupAlertsBySeverity(alerts)
		report.Data["alerts_by_status"] = s.groupAlertsByStatus(alerts)
	}

	return report, nil
}

// Helper methods

func (s *Service) generateAlertTitle(rule *domain.AlertRule, context domain.AlertContext) string {
	switch rule.Type {
	case domain.AlertTypeZoneViolation:
		return fmt.Sprintf("Zone Violation in %s", context.ZoneID)
	case domain.AlertTypeLoitering:
		return "Loitering Detected"
	case domain.AlertTypeCrowding:
		return "Crowding Alert"
	case domain.AlertTypeUnauthorizedAccess:
		return "Unauthorized Access Detected"
	default:
		return fmt.Sprintf("%s Alert", rule.Type)
	}
}

func (s *Service) generateAlertMessage(rule *domain.AlertRule, context domain.AlertContext) string {
	return fmt.Sprintf("Alert rule '%s' triggered at %s", rule.Name, context.Timestamp.Format(time.RFC3339))
}

func (s *Service) updateAlertStatistics(streamID, zoneID string, alertType domain.AlertType, severity domain.AlertSeverity) {
	key := streamID
	if zoneID != "" {
		key = streamID + ":" + zoneID
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	stats, exists := s.alertStatistics[key]
	if !exists {
		stats = &domain.AlertStatistics{
			StreamID:         streamID,
			ZoneID:           zoneID,
			AlertsByType:     make(map[domain.AlertType]int64),
			AlertsBySeverity: make(map[domain.AlertSeverity]int64),
			AlertsByHour:     make(map[int]int64),
			AlertsByDay:      make(map[string]int64),
			StartTime:        time.Now(),
		}
		s.alertStatistics[key] = stats
	}

	now := time.Now()
	stats.TotalAlerts++
	stats.ActiveAlerts++
	stats.AlertsByType[alertType]++
	stats.AlertsBySeverity[severity]++
	stats.AlertsByHour[now.Hour()]++
	stats.AlertsByDay[now.Format("2006-01-02")]++
	stats.LastAlert = &now
}

func (s *Service) groupAlertsByType(alerts []*domain.Alert) map[domain.AlertType]int64 {
	groups := make(map[domain.AlertType]int64)
	for _, alert := range alerts {
		groups[alert.Type]++
	}
	return groups
}

func (s *Service) groupAlertsBySeverity(alerts []*domain.Alert) map[domain.AlertSeverity]int64 {
	groups := make(map[domain.AlertSeverity]int64)
	for _, alert := range alerts {
		groups[alert.Severity]++
	}
	return groups
}

func (s *Service) groupAlertsByStatus(alerts []*domain.Alert) map[domain.AlertStatus]int64 {
	groups := make(map[domain.AlertStatus]int64)
	for _, alert := range alerts {
		groups[alert.Status]++
	}
	return groups
}
