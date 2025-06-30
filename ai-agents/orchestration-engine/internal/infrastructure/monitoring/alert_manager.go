package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DashboardAlertManager handles dashboard alerts
type DashboardAlertManager struct {
	rules       map[string]*AlertRule
	evaluator   *AlertEvaluator
	notifier    *AlertNotifier
	logger      Logger
	mutex       sync.RWMutex
	stopCh      chan struct{}
}

// AlertEvaluator evaluates alert conditions
type AlertEvaluator struct {
	logger Logger
}

// AlertNotifier sends alert notifications
type AlertNotifier struct {
	channels map[string]NotificationChannel
	logger   Logger
}

// NewDashboardAlertManager creates a new dashboard alert manager
func NewDashboardAlertManager(logger Logger) *DashboardAlertManager {
	return &DashboardAlertManager{
		rules:     make(map[string]*AlertRule),
		evaluator: &AlertEvaluator{logger: logger},
		notifier:  &AlertNotifier{channels: make(map[string]NotificationChannel), logger: logger},
		logger:    logger,
		stopCh:    make(chan struct{}),
	}
}

// Start starts the alert manager
func (am *DashboardAlertManager) Start(ctx context.Context) error {
	am.logger.Info("Starting alert manager")
	
	// Start alert evaluation loop
	go am.evaluationLoop(ctx)
	
	am.logger.Info("Alert manager started")
	return nil
}

// Stop stops the alert manager
func (am *DashboardAlertManager) Stop(ctx context.Context) error {
	am.logger.Info("Stopping alert manager")
	
	close(am.stopCh)
	
	am.logger.Info("Alert manager stopped")
	return nil
}

// AddRule adds an alert rule
func (am *DashboardAlertManager) AddRule(rule *AlertRule) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	am.rules[rule.ID] = rule
	am.logger.Info("Alert rule added", "rule_id", rule.ID, "name", rule.Name)

	return nil
}

// RemoveRule removes an alert rule
func (am *DashboardAlertManager) RemoveRule(ruleID string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if _, exists := am.rules[ruleID]; !exists {
		return fmt.Errorf("alert rule %s not found", ruleID)
	}

	delete(am.rules, ruleID)
	am.logger.Info("Alert rule removed", "rule_id", ruleID)

	return nil
}

// GetRule returns an alert rule
func (am *DashboardAlertManager) GetRule(ruleID string) (*AlertRule, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	rule, exists := am.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("alert rule %s not found", ruleID)
	}

	ruleCopy := *rule
	return &ruleCopy, nil
}

// ListRules returns all alert rules
func (am *DashboardAlertManager) ListRules() []*AlertRule {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		ruleCopy := *rule
		rules = append(rules, &ruleCopy)
	}

	return rules
}

// AddNotificationChannel adds a notification channel
func (am *DashboardAlertManager) AddNotificationChannel(name string, channel NotificationChannel) {
	am.notifier.channels[name] = channel
	am.logger.Info("Notification channel added", "name", name, "type", channel.GetType())
}

// RemoveNotificationChannel removes a notification channel
func (am *DashboardAlertManager) RemoveNotificationChannel(name string) {
	delete(am.notifier.channels, name)
	am.logger.Info("Notification channel removed", "name", name)
}

// evaluationLoop runs the alert evaluation loop
func (am *DashboardAlertManager) evaluationLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-am.stopCh:
			return
		case <-ticker.C:
			am.evaluateRules(ctx)
		}
	}
}

// evaluateRules evaluates all alert rules
func (am *DashboardAlertManager) evaluateRules(ctx context.Context) {
	am.mutex.RLock()
	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	am.mutex.RUnlock()

	for _, rule := range rules {
		go am.evaluateRule(ctx, rule)
	}
}

// evaluateRule evaluates a single alert rule
func (am *DashboardAlertManager) evaluateRule(ctx context.Context, rule *AlertRule) {
	// This is a simplified implementation
	// In a real system, this would query the data source and evaluate conditions
	
	am.logger.Debug("Evaluating alert rule", "rule_id", rule.ID, "name", rule.Name)
	
	// Mock evaluation - in reality, this would fetch data and check conditions
	shouldAlert := am.evaluator.evaluate(ctx, rule)
	
	if shouldAlert {
		alert := &Alert{
			ID:          fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			RuleID:      rule.ID,
			DashboardID: "", // Would be set based on rule context
			WidgetID:    rule.WidgetID,
			Title:       rule.Name,
			Message:     rule.Message,
			Severity:    rule.Severity,
			Value:       0.0, // Would be actual measured value
			Threshold:   rule.Threshold,
			Timestamp:   time.Now(),
			Status:      AlertStatusFiring,
			Metadata:    make(map[string]interface{}),
		}
		
		am.sendAlert(ctx, alert)
	}
}

// sendAlert sends an alert through notification channels
func (am *DashboardAlertManager) sendAlert(ctx context.Context, alert *Alert) {
	am.logger.Info("Sending alert", "alert_id", alert.ID, "severity", alert.Severity)
	
	for name, channel := range am.notifier.channels {
		if channel.IsEnabled() {
			go func(channelName string, ch NotificationChannel) {
				if err := ch.Send(ctx, alert); err != nil {
					am.logger.Error("Failed to send alert", err, "channel", channelName, "alert_id", alert.ID)
				}
			}(name, channel)
		}
	}
}

// evaluate evaluates an alert condition (simplified implementation)
func (ae *AlertEvaluator) evaluate(ctx context.Context, rule *AlertRule) bool {
	// This is a mock implementation
	// In a real system, this would:
	// 1. Query the data source using the rule's condition
	// 2. Apply the threshold and operator
	// 3. Return true if alert should fire
	
	ae.logger.Debug("Evaluating rule condition", "rule_id", rule.ID, "condition", rule.Condition)
	
	// Mock: randomly trigger alerts for demonstration
	return time.Now().Unix()%10 == 0 // Trigger alert every ~10 evaluations
}

// EmailNotificationChannel implements email notifications
type EmailNotificationChannel struct {
	smtpServer string
	username   string
	password   string
	enabled    bool
	logger     Logger
}

// NewEmailNotificationChannel creates a new email notification channel
func NewEmailNotificationChannel(smtpServer, username, password string, logger Logger) *EmailNotificationChannel {
	return &EmailNotificationChannel{
		smtpServer: smtpServer,
		username:   username,
		password:   password,
		enabled:    true,
		logger:     logger,
	}
}

// Send sends an alert via email
func (enc *EmailNotificationChannel) Send(ctx context.Context, alert *Alert) error {
	// Mock implementation - in reality, this would send an actual email
	enc.logger.Info("Sending email alert", 
		"alert_id", alert.ID, 
		"severity", alert.Severity, 
		"title", alert.Title,
	)
	
	// Simulate email sending delay
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// GetType returns the notification channel type
func (enc *EmailNotificationChannel) GetType() string {
	return "email"
}

// IsEnabled returns whether the channel is enabled
func (enc *EmailNotificationChannel) IsEnabled() bool {
	return enc.enabled
}

// SlackNotificationChannel implements Slack notifications
type SlackNotificationChannel struct {
	webhookURL string
	channel    string
	enabled    bool
	logger     Logger
}

// NewSlackNotificationChannel creates a new Slack notification channel
func NewSlackNotificationChannel(webhookURL, channel string, logger Logger) *SlackNotificationChannel {
	return &SlackNotificationChannel{
		webhookURL: webhookURL,
		channel:    channel,
		enabled:    true,
		logger:     logger,
	}
}

// Send sends an alert to Slack
func (snc *SlackNotificationChannel) Send(ctx context.Context, alert *Alert) error {
	// Mock implementation - in reality, this would send to Slack webhook
	snc.logger.Info("Sending Slack alert", 
		"alert_id", alert.ID, 
		"severity", alert.Severity, 
		"title", alert.Title,
		"channel", snc.channel,
	)
	
	// Simulate Slack API call delay
	time.Sleep(200 * time.Millisecond)
	
	return nil
}

// GetType returns the notification channel type
func (snc *SlackNotificationChannel) GetType() string {
	return "slack"
}

// IsEnabled returns whether the channel is enabled
func (snc *SlackNotificationChannel) IsEnabled() bool {
	return snc.enabled
}

// WebhookNotificationChannel implements webhook notifications
type WebhookNotificationChannel struct {
	url     string
	headers map[string]string
	enabled bool
	logger  Logger
}

// NewWebhookNotificationChannel creates a new webhook notification channel
func NewWebhookNotificationChannel(url string, headers map[string]string, logger Logger) *WebhookNotificationChannel {
	return &WebhookNotificationChannel{
		url:     url,
		headers: headers,
		enabled: true,
		logger:  logger,
	}
}

// Send sends an alert via webhook
func (wnc *WebhookNotificationChannel) Send(ctx context.Context, alert *Alert) error {
	// Mock implementation - in reality, this would make an HTTP POST request
	wnc.logger.Info("Sending webhook alert", 
		"alert_id", alert.ID, 
		"severity", alert.Severity, 
		"title", alert.Title,
		"url", wnc.url,
	)
	
	// Simulate HTTP request delay
	time.Sleep(150 * time.Millisecond)
	
	return nil
}

// GetType returns the notification channel type
func (wnc *WebhookNotificationChannel) GetType() string {
	return "webhook"
}

// IsEnabled returns whether the channel is enabled
func (wnc *WebhookNotificationChannel) IsEnabled() bool {
	return wnc.enabled
}

// GetAlertStats returns alert manager statistics
func (am *DashboardAlertManager) GetAlertStats() *AlertStats {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	enabledRules := 0
	for _, rule := range am.rules {
		if rule.Enabled {
			enabledRules++
		}
	}

	return &AlertStats{
		TotalRules:         len(am.rules),
		EnabledRules:       enabledRules,
		NotificationChannels: len(am.notifier.channels),
		LastEvaluation:     time.Now(),
	}
}

// AlertStats represents alert manager statistics
type AlertStats struct {
	TotalRules           int       `json:"total_rules"`
	EnabledRules         int       `json:"enabled_rules"`
	NotificationChannels int       `json:"notification_channels"`
	LastEvaluation       time.Time `json:"last_evaluation"`
}
