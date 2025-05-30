package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// SecurityMonitor provides real-time security monitoring
type SecurityMonitor struct {
	logger          *zap.Logger
	auditor         *SecurityAuditor
	metrics         *SecurityMetrics
	alertManager    *AlertManager
	running         bool
	stopChan        chan struct{}
	wg              sync.WaitGroup
	mutex           sync.RWMutex
}

// SecurityMetrics tracks security-related metrics
type SecurityMetrics struct {
	TotalEvents       int64           `json:"total_events"`
	SuspiciousEvents  int64           `json:"suspicious_events"`
	BlockedEvents     int64           `json:"blocked_events"`
	AlertsSent        int64           `json:"alerts_sent"`
	AverageRisk       decimal.Decimal `json:"average_risk"`
	LastUpdate        time.Time       `json:"last_update"`
	EventsByCategory  map[AuditCategory]int64 `json:"events_by_category"`
	EventsBySeverity  map[SeverityLevel]int64 `json:"events_by_severity"`
	mutex             sync.RWMutex
}

// AlertManager manages security alerts
type AlertManager struct {
	logger       *zap.Logger
	handlers     []AlertHandler
	alertHistory []AlertRecord
	mutex        sync.RWMutex
}

// AlertRecord represents a historical alert
type AlertRecord struct {
	ID          string        `json:"id"`
	Event       SuspiciousEvent `json:"event"`
	Handled     bool          `json:"handled"`
	HandledAt   *time.Time    `json:"handled_at,omitempty"`
	Response    string        `json:"response,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// EmailAlertHandler sends alerts via email
type EmailAlertHandler struct {
	logger    *zap.Logger
	smtpHost  string
	smtpPort  int
	username  string
	password  string
	recipients []string
}

// SlackAlertHandler sends alerts to Slack
type SlackAlertHandler struct {
	logger     *zap.Logger
	webhookURL string
	channel    string
}

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor(logger *zap.Logger, auditor *SecurityAuditor) *SecurityMonitor {
	return &SecurityMonitor{
		logger:       logger,
		auditor:      auditor,
		metrics:      NewSecurityMetrics(),
		alertManager: NewAlertManager(logger),
		stopChan:     make(chan struct{}),
	}
}

// NewSecurityMetrics creates new security metrics
func NewSecurityMetrics() *SecurityMetrics {
	return &SecurityMetrics{
		EventsByCategory: make(map[AuditCategory]int64),
		EventsBySeverity: make(map[SeverityLevel]int64),
		LastUpdate:       time.Now(),
	}
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *zap.Logger) *AlertManager {
	return &AlertManager{
		logger:       logger,
		handlers:     []AlertHandler{},
		alertHistory: []AlertRecord{},
	}
}

// Start starts the security monitor
func (sm *SecurityMonitor) Start(ctx context.Context) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.running {
		return fmt.Errorf("security monitor already running")
	}

	sm.running = true
	sm.wg.Add(1)

	go sm.monitorLoop(ctx)

	sm.logger.Info("Security monitor started")
	return nil
}

// Stop stops the security monitor
func (sm *SecurityMonitor) Stop() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if !sm.running {
		return fmt.Errorf("security monitor not running")
	}

	close(sm.stopChan)
	sm.wg.Wait()
	sm.running = false

	sm.logger.Info("Security monitor stopped")
	return nil
}

// monitorLoop is the main monitoring loop
func (sm *SecurityMonitor) monitorLoop(ctx context.Context) {
	defer sm.wg.Done()

	suspiciousEvents := sm.auditor.GetSuspiciousEvents()
	ticker := time.NewTicker(time.Minute) // Update metrics every minute
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopChan:
			return
		case <-ctx.Done():
			return
		case event := <-suspiciousEvents:
			sm.handleSuspiciousEvent(ctx, event)
		case <-ticker.C:
			sm.updateMetrics()
		}
	}
}

// handleSuspiciousEvent handles a suspicious event
func (sm *SecurityMonitor) handleSuspiciousEvent(ctx context.Context, event SuspiciousEvent) {
	sm.logger.Warn("Suspicious event detected",
		zap.String("event_id", event.ID),
		zap.String("rule_id", event.RuleID),
		zap.String("severity", string(event.Severity)),
		zap.String("description", event.Description),
	)

	// Update metrics
	sm.metrics.mutex.Lock()
	sm.metrics.SuspiciousEvents++
	sm.metrics.EventsBySeverity[event.Severity]++
	sm.metrics.mutex.Unlock()

	// Send alert
	if err := sm.alertManager.SendAlert(ctx, event); err != nil {
		sm.logger.Error("Failed to send alert",
			zap.String("event_id", event.ID),
			zap.Error(err),
		)
	}
}

// updateMetrics updates security metrics
func (sm *SecurityMonitor) updateMetrics() {
	sm.metrics.mutex.Lock()
	defer sm.metrics.mutex.Unlock()

	sm.metrics.LastUpdate = time.Now()

	// Calculate average risk
	if sm.metrics.SuspiciousEvents > 0 {
		// This is a simplified calculation
		sm.metrics.AverageRisk = decimal.NewFromFloat(5.0) // Placeholder
	}

	sm.logger.Debug("Security metrics updated",
		zap.Int64("total_events", sm.metrics.TotalEvents),
		zap.Int64("suspicious_events", sm.metrics.SuspiciousEvents),
		zap.Int64("blocked_events", sm.metrics.BlockedEvents),
	)
}

// GetMetrics returns current security metrics
func (sm *SecurityMonitor) GetMetrics() SecurityMetrics {
	sm.metrics.mutex.RLock()
	defer sm.metrics.mutex.RUnlock()

	// Create a copy to avoid race conditions
	metrics := SecurityMetrics{
		TotalEvents:      sm.metrics.TotalEvents,
		SuspiciousEvents: sm.metrics.SuspiciousEvents,
		BlockedEvents:    sm.metrics.BlockedEvents,
		AlertsSent:       sm.metrics.AlertsSent,
		AverageRisk:      sm.metrics.AverageRisk,
		LastUpdate:       sm.metrics.LastUpdate,
		EventsByCategory: make(map[AuditCategory]int64),
		EventsBySeverity: make(map[SeverityLevel]int64),
	}

	for k, v := range sm.metrics.EventsByCategory {
		metrics.EventsByCategory[k] = v
	}

	for k, v := range sm.metrics.EventsBySeverity {
		metrics.EventsBySeverity[k] = v
	}

	return metrics
}

// AddAlertHandler adds an alert handler
func (sm *SecurityMonitor) AddAlertHandler(handler AlertHandler) {
	sm.alertManager.AddHandler(handler)
}

// SendAlert sends an alert
func (am *AlertManager) SendAlert(ctx context.Context, event SuspiciousEvent) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Create alert record
	record := AlertRecord{
		ID:        event.ID,
		Event:     event,
		Handled:   false,
		CreatedAt: time.Now(),
	}

	// Send to all handlers
	var lastErr error
	for _, handler := range am.handlers {
		if err := handler.HandleAlert(ctx, event); err != nil {
			am.logger.Error("Alert handler failed",
				zap.String("event_id", event.ID),
				zap.Error(err),
			)
			lastErr = err
		}
	}

	if lastErr == nil {
		record.Handled = true
		now := time.Now()
		record.HandledAt = &now
		record.Response = "Alert sent successfully"
	}

	// Store alert record
	am.alertHistory = append(am.alertHistory, record)

	// Keep only last 1000 alerts
	if len(am.alertHistory) > 1000 {
		am.alertHistory = am.alertHistory[len(am.alertHistory)-1000:]
	}

	return lastErr
}

// AddHandler adds an alert handler
func (am *AlertManager) AddHandler(handler AlertHandler) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.handlers = append(am.handlers, handler)
	am.logger.Info("Alert handler added")
}

// GetAlertHistory returns alert history
func (am *AlertManager) GetAlertHistory(limit int) []AlertRecord {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	if limit <= 0 || limit > len(am.alertHistory) {
		limit = len(am.alertHistory)
	}

	// Return last N alerts
	start := len(am.alertHistory) - limit
	if start < 0 {
		start = 0
	}

	result := make([]AlertRecord, limit)
	copy(result, am.alertHistory[start:])
	return result
}

// NewEmailAlertHandler creates a new email alert handler
func NewEmailAlertHandler(logger *zap.Logger, smtpHost string, smtpPort int, username, password string, recipients []string) *EmailAlertHandler {
	return &EmailAlertHandler{
		logger:     logger,
		smtpHost:   smtpHost,
		smtpPort:   smtpPort,
		username:   username,
		password:   password,
		recipients: recipients,
	}
}

// HandleAlert handles an alert via email
func (eah *EmailAlertHandler) HandleAlert(ctx context.Context, event SuspiciousEvent) error {
	subject := fmt.Sprintf("Security Alert: %s", event.Description)
	body := fmt.Sprintf(`
Security Alert Detected

Event ID: %s
Rule ID: %s
Severity: %s
Risk Score: %s
Description: %s
Timestamp: %s

Event Details:
- User ID: %s
- Amount: %s
- Token: %s
- Chain: %s
- Protocol: %s

Please investigate immediately.
`, 
		event.ID,
		event.RuleID,
		event.Severity,
		event.Risk.String(),
		event.Description,
		event.Timestamp.Format(time.RFC3339),
		event.Event.UserID,
		event.Event.Amount.String(),
		event.Event.Token,
		event.Event.Chain,
		event.Event.Protocol,
	)

	// In a real implementation, you would send the email here
	eah.logger.Info("Email alert would be sent",
		zap.String("subject", subject),
		zap.Strings("recipients", eah.recipients),
	)

	return nil
}

// NewSlackAlertHandler creates a new Slack alert handler
func NewSlackAlertHandler(logger *zap.Logger, webhookURL, channel string) *SlackAlertHandler {
	return &SlackAlertHandler{
		logger:     logger,
		webhookURL: webhookURL,
		channel:    channel,
	}
}

// HandleAlert handles an alert via Slack
func (sah *SlackAlertHandler) HandleAlert(ctx context.Context, event SuspiciousEvent) error {
	message := fmt.Sprintf(`🚨 *Security Alert*

*Event ID:* %s
*Severity:* %s
*Risk Score:* %s
*Description:* %s

*Event Details:*
• User ID: %s
• Amount: %s %s
• Chain: %s
• Protocol: %s

*Timestamp:* %s`,
		event.ID,
		event.Severity,
		event.Risk.String(),
		event.Description,
		event.Event.UserID,
		event.Event.Amount.String(),
		event.Event.Token,
		event.Event.Chain,
		event.Event.Protocol,
		event.Timestamp.Format(time.RFC3339),
	)

	// In a real implementation, you would send to Slack here
	sah.logger.Info("Slack alert would be sent",
		zap.String("channel", sah.channel),
		zap.String("message", message),
	)

	return nil
}
