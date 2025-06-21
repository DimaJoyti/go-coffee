package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
)

// NotificationService implements the notification service
type NotificationService struct {
	logger *zap.Logger
	config NotificationConfig
	client *http.Client
}

// NotificationConfig configures the notification service
type NotificationConfig struct {
	// Email configuration
	SMTPHost     string `yaml:"smtp_host"`
	SMTPPort     int    `yaml:"smtp_port"`
	SMTPUsername string `yaml:"smtp_username"`
	SMTPPassword string `yaml:"smtp_password"`
	FromEmail    string `yaml:"from_email"`
	FromName     string `yaml:"from_name"`

	// SMS configuration
	SMSProvider string `yaml:"sms_provider"`
	SMSAPIKey   string `yaml:"sms_api_key"`
	SMSFrom     string `yaml:"sms_from"`

	// Webhook configuration
	WebhookTimeout time.Duration `yaml:"webhook_timeout"`
	WebhookRetries int           `yaml:"webhook_retries"`

	// Slack configuration
	SlackWebhookURL string `yaml:"slack_webhook_url"`
	SlackChannel    string `yaml:"slack_channel"`
	SlackUsername   string `yaml:"slack_username"`

	// Discord configuration
	DiscordWebhookURL string `yaml:"discord_webhook_url"`

	// Telegram configuration
	TelegramBotToken string `yaml:"telegram_bot_token"`
	TelegramChatID   string `yaml:"telegram_chat_id"`

	// Pushover configuration
	PushoverAppToken  string `yaml:"pushover_app_token"`
	PushoverUserKey   string `yaml:"pushover_user_key"`
}

// DefaultNotificationConfig returns default configuration
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{
		SMTPPort:       587,
		WebhookTimeout: 30 * time.Second,
		WebhookRetries: 3,
		SlackUsername:  "AlertBot",
	}
}

// NewNotificationService creates a new notification service
func NewNotificationService(logger *zap.Logger, config NotificationConfig) *NotificationService {
	return &NotificationService{
		logger: logger.With(zap.String("component", "notification_service")),
		config: config,
		client: &http.Client{
			Timeout: config.WebhookTimeout,
		},
	}
}

// SendEmail sends an email notification
func (ns *NotificationService) SendEmail(notification *domain.Notification) error {
	if ns.config.SMTPHost == "" {
		return fmt.Errorf("SMTP configuration not provided")
	}

	// Get recipient from notification
	recipient := notification.Recipient
	if recipient == "" {
		if recipientConfig, ok := notification.Metadata["recipient"].(string); ok {
			recipient = recipientConfig
		} else {
			return fmt.Errorf("no recipient specified for email notification")
		}
	}

	// Prepare email message
	from := fmt.Sprintf("%s <%s>", ns.config.FromName, ns.config.FromEmail)
	to := recipient
	subject := notification.Subject
	body := notification.Content

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body)

	// SMTP authentication
	auth := smtp.PlainAuth("", ns.config.SMTPUsername, ns.config.SMTPPassword, ns.config.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%d", ns.config.SMTPHost, ns.config.SMTPPort)
	err := smtp.SendMail(addr, auth, ns.config.FromEmail, []string{recipient}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	ns.logger.Info("Email notification sent",
		zap.String("notification_id", notification.ID),
		zap.String("recipient", recipient),
		zap.String("subject", subject))

	return nil
}

// SendSMS sends an SMS notification
func (ns *NotificationService) SendSMS(notification *domain.Notification) error {
	if ns.config.SMSProvider == "" || ns.config.SMSAPIKey == "" {
		return fmt.Errorf("SMS configuration not provided")
	}

	// Get recipient phone number
	recipient := notification.Recipient
	if recipient == "" {
		if recipientConfig, ok := notification.Metadata["recipient"].(string); ok {
			recipient = recipientConfig
		} else {
			return fmt.Errorf("no recipient specified for SMS notification")
		}
	}

	// This is a simplified SMS implementation
	// In production, you would integrate with actual SMS providers like Twilio, AWS SNS, etc.
	ns.logger.Info("SMS notification would be sent",
		zap.String("notification_id", notification.ID),
		zap.String("recipient", recipient),
		zap.String("provider", ns.config.SMSProvider))

	return nil
}

// SendWebhook sends a webhook notification
func (ns *NotificationService) SendWebhook(notification *domain.Notification) error {
	// Get webhook URL from notification metadata
	webhookURL, ok := notification.Metadata["webhook_url"].(string)
	if !ok {
		return fmt.Errorf("webhook URL not specified in notification metadata")
	}

	// Prepare webhook payload
	payload := map[string]interface{}{
		"notification_id": notification.ID,
		"alert_id":        notification.AlertID,
		"type":            notification.Type,
		"subject":         notification.Subject,
		"content":         notification.Content,
		"timestamp":       notification.CreatedAt,
		"metadata":        notification.Metadata,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Send webhook with retries
	var lastErr error
	for attempt := 0; attempt <= ns.config.WebhookRetries; attempt++ {
		resp, err := ns.client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			lastErr = err
			ns.logger.Warn("Webhook attempt failed",
				zap.String("notification_id", notification.ID),
				zap.String("webhook_url", webhookURL),
				zap.Int("attempt", attempt+1),
				zap.Error(err))
			
			if attempt < ns.config.WebhookRetries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
			}
			continue
		}

		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			ns.logger.Info("Webhook notification sent",
				zap.String("notification_id", notification.ID),
				zap.String("webhook_url", webhookURL),
				zap.Int("status_code", resp.StatusCode))
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status code %d", resp.StatusCode)
		ns.logger.Warn("Webhook returned non-success status",
			zap.String("notification_id", notification.ID),
			zap.String("webhook_url", webhookURL),
			zap.Int("status_code", resp.StatusCode),
			zap.Int("attempt", attempt+1))

		if attempt < ns.config.WebhookRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", ns.config.WebhookRetries+1, lastErr)
}

// SendSlack sends a Slack notification
func (ns *NotificationService) SendSlack(notification *domain.Notification) error {
	if ns.config.SlackWebhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	// Get channel from notification metadata or use default
	channel := ns.config.SlackChannel
	if channelConfig, ok := notification.Metadata["channel"].(string); ok {
		channel = channelConfig
	}

	// Prepare Slack payload
	payload := map[string]interface{}{
		"channel":    channel,
		"username":   ns.config.SlackUsername,
		"text":       notification.Subject,
		"attachments": []map[string]interface{}{
			{
				"color":     ns.getSlackColor(notification),
				"title":     notification.Subject,
				"text":      notification.Content,
				"timestamp": notification.CreatedAt.Unix(),
				"fields": []map[string]interface{}{
					{
						"title": "Alert ID",
						"value": notification.AlertID,
						"short": true,
					},
					{
						"title": "Type",
						"value": string(notification.Type),
						"short": true,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	resp, err := ns.client.Post(ns.config.SlackWebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status code %d", resp.StatusCode)
	}

	ns.logger.Info("Slack notification sent",
		zap.String("notification_id", notification.ID),
		zap.String("channel", channel))

	return nil
}

// SendDiscord sends a Discord notification
func (ns *NotificationService) SendDiscord(notification *domain.Notification) error {
	if ns.config.DiscordWebhookURL == "" {
		return fmt.Errorf("Discord webhook URL not configured")
	}

	// Prepare Discord payload
	payload := map[string]interface{}{
		"content": notification.Subject,
		"embeds": []map[string]interface{}{
			{
				"title":       notification.Subject,
				"description": notification.Content,
				"color":       ns.getDiscordColor(notification),
				"timestamp":   notification.CreatedAt.Format(time.RFC3339),
				"fields": []map[string]interface{}{
					{
						"name":   "Alert ID",
						"value":  notification.AlertID,
						"inline": true,
					},
					{
						"name":   "Type",
						"value":  string(notification.Type),
						"inline": true,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	resp, err := ns.client.Post(ns.config.DiscordWebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send Discord notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Discord webhook returned status code %d", resp.StatusCode)
	}

	ns.logger.Info("Discord notification sent",
		zap.String("notification_id", notification.ID))

	return nil
}

// SendTelegram sends a Telegram notification
func (ns *NotificationService) SendTelegram(notification *domain.Notification) error {
	if ns.config.TelegramBotToken == "" || ns.config.TelegramChatID == "" {
		return fmt.Errorf("Telegram configuration not provided")
	}

	// Get chat ID from notification metadata or use default
	chatID := ns.config.TelegramChatID
	if chatIDConfig, ok := notification.Metadata["chat_id"].(string); ok {
		chatID = chatIDConfig
	}

	// Prepare Telegram message
	message := fmt.Sprintf("*%s*\n\n%s\n\nAlert ID: `%s`", 
		notification.Subject, notification.Content, notification.AlertID)

	// Prepare Telegram payload
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Telegram payload: %w", err)
	}

	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", ns.config.TelegramBotToken)
	resp, err := ns.client.Post(telegramURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send Telegram notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API returned status code %d", resp.StatusCode)
	}

	ns.logger.Info("Telegram notification sent",
		zap.String("notification_id", notification.ID),
		zap.String("chat_id", chatID))

	return nil
}

// SendPushover sends a Pushover notification
func (ns *NotificationService) SendPushover(notification *domain.Notification) error {
	if ns.config.PushoverAppToken == "" || ns.config.PushoverUserKey == "" {
		return fmt.Errorf("Pushover configuration not provided")
	}

	// Prepare Pushover payload
	payload := map[string]interface{}{
		"token":   ns.config.PushoverAppToken,
		"user":    ns.config.PushoverUserKey,
		"title":   notification.Subject,
		"message": notification.Content,
		"priority": ns.getPushoverPriority(notification),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Pushover payload: %w", err)
	}

	resp, err := ns.client.Post("https://api.pushover.net/1/messages.json", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send Pushover notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Pushover API returned status code %d", resp.StatusCode)
	}

	ns.logger.Info("Pushover notification sent",
		zap.String("notification_id", notification.ID))

	return nil
}

// SendCustom sends a custom notification
func (ns *NotificationService) SendCustom(notification *domain.Notification) error {
	// Custom notification implementation would go here
	// This could involve calling external APIs, writing to files, etc.
	ns.logger.Info("Custom notification would be sent",
		zap.String("notification_id", notification.ID),
		zap.Any("metadata", notification.Metadata))

	return nil
}

// Helper methods

func (ns *NotificationService) getSlackColor(notification *domain.Notification) string {
	// Determine color based on alert severity or type
	if severity, ok := notification.Metadata["severity"].(string); ok {
		switch severity {
		case "critical":
			return "danger"
		case "high":
			return "warning"
		case "medium":
			return "good"
		case "low":
			return "#36a64f"
		}
	}
	return "good"
}

func (ns *NotificationService) getDiscordColor(notification *domain.Notification) int {
	// Determine color based on alert severity or type
	if severity, ok := notification.Metadata["severity"].(string); ok {
		switch severity {
		case "critical":
			return 0xFF0000 // Red
		case "high":
			return 0xFF8C00 // Orange
		case "medium":
			return 0xFFFF00 // Yellow
		case "low":
			return 0x00FF00 // Green
		}
	}
	return 0x0099FF // Blue (default)
}

func (ns *NotificationService) getPushoverPriority(notification *domain.Notification) int {
	// Determine priority based on alert severity
	if severity, ok := notification.Metadata["severity"].(string); ok {
		switch severity {
		case "critical":
			return 2 // Emergency
		case "high":
			return 1 // High
		case "medium":
			return 0 // Normal
		case "low":
			return -1 // Low
		}
	}
	return 0 // Normal (default)
}

// Ensure NotificationService implements domain.AlertNotificationService
var _ domain.AlertNotificationService = (*NotificationService)(nil)
