package domain

import (
	"fmt"
	"time"
)

// Alert represents a system alert triggered by detection events
type Alert struct {
	ID          string            `json:"id" db:"id"`
	Type        AlertType         `json:"type" db:"type"`
	Severity    AlertSeverity     `json:"severity" db:"severity"`
	Title       string            `json:"title" db:"title"`
	Message     string            `json:"message" db:"message"`
	StreamID    string            `json:"stream_id" db:"stream_id"`
	ZoneID      string            `json:"zone_id,omitempty" db:"zone_id"`
	ObjectID    string            `json:"object_id,omitempty" db:"object_id"`
	ObjectClass string            `json:"object_class,omitempty" db:"object_class"`
	Confidence  float64           `json:"confidence,omitempty" db:"confidence"`
	Position    *Point            `json:"position,omitempty" db:"position"`
	BoundingBox *Rectangle        `json:"bounding_box,omitempty" db:"bounding_box"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Status      AlertStatus       `json:"status" db:"status"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	ResolvedAt  *time.Time        `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy  string            `json:"resolved_by,omitempty" db:"resolved_by"`
	Tags        []string          `json:"tags" db:"tags"`
}

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeZoneViolation    AlertType = "zone_violation"
	AlertTypeLoitering        AlertType = "loitering"
	AlertTypeCrowding         AlertType = "crowding"
	AlertTypeUnauthorizedAccess AlertType = "unauthorized_access"
	AlertTypeAbandonedObject  AlertType = "abandoned_object"
	AlertTypeUnusualBehavior  AlertType = "unusual_behavior"
	AlertTypeObjectDetection  AlertType = "object_detection"
	AlertTypeSystemError      AlertType = "system_error"
	AlertTypeStreamOffline    AlertType = "stream_offline"
	AlertTypeModelError       AlertType = "model_error"
	AlertTypeStorageError     AlertType = "storage_error"
	AlertTypeCustom           AlertType = "custom"
)

// AlertSeverity defines the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus defines the current status of an alert
type AlertStatus string

const (
	AlertStatusActive     AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved   AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
	AlertStatusExpired    AlertStatus = "expired"
)

// AlertRule defines rules for generating alerts
type AlertRule struct {
	ID              string            `json:"id" db:"id"`
	Name            string            `json:"name" db:"name"`
	Description     string            `json:"description" db:"description"`
	StreamID        string            `json:"stream_id" db:"stream_id"`
	ZoneID          string            `json:"zone_id,omitempty" db:"zone_id"`
	Type            AlertType         `json:"type" db:"type"`
	Severity        AlertSeverity     `json:"severity" db:"severity"`
	Conditions      AlertConditions   `json:"conditions" db:"conditions"`
	Actions         []AlertAction     `json:"actions" db:"actions"`
	Cooldown        time.Duration     `json:"cooldown" db:"cooldown"`
	IsActive        bool              `json:"is_active" db:"is_active"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
	LastTriggered   *time.Time        `json:"last_triggered,omitempty" db:"last_triggered"`
	TriggerCount    int64             `json:"trigger_count" db:"trigger_count"`
	Tags            []string          `json:"tags" db:"tags"`
}

// AlertConditions defines the conditions that trigger an alert
type AlertConditions struct {
	ObjectClasses     []string      `json:"object_classes,omitempty"`
	MinConfidence     float64       `json:"min_confidence,omitempty"`
	MaxConfidence     float64       `json:"max_confidence,omitempty"`
	MinObjects        int           `json:"min_objects,omitempty"`
	MaxObjects        int           `json:"max_objects,omitempty"`
	DwellTimeMin      time.Duration `json:"dwell_time_min,omitempty"`
	DwellTimeMax      time.Duration `json:"dwell_time_max,omitempty"`
	TimeWindows       []TimeWindow  `json:"time_windows,omitempty"`
	RequiredZoneEvents []ZoneEventType `json:"required_zone_events,omitempty"`
	CustomConditions  map[string]interface{} `json:"custom_conditions,omitempty"`
}

// AlertAction defines an action to take when an alert is triggered
type AlertAction struct {
	Type       AlertActionType        `json:"type"`
	Config     map[string]interface{} `json:"config"`
	IsEnabled  bool                   `json:"is_enabled"`
	RetryCount int                    `json:"retry_count"`
	Timeout    time.Duration          `json:"timeout"`
}

// AlertActionType defines the type of action to take
type AlertActionType string

const (
	AlertActionEmail     AlertActionType = "email"
	AlertActionSMS       AlertActionType = "sms"
	AlertActionWebhook   AlertActionType = "webhook"
	AlertActionSlack     AlertActionType = "slack"
	AlertActionDiscord   AlertActionType = "discord"
	AlertActionTelegram  AlertActionType = "telegram"
	AlertActionPushover  AlertActionType = "pushover"
	AlertActionRecord    AlertActionType = "record"
	AlertActionSnapshot  AlertActionType = "snapshot"
	AlertActionCustom    AlertActionType = "custom"
)

// Notification represents a notification sent for an alert
type Notification struct {
	ID          string                 `json:"id" db:"id"`
	AlertID     string                 `json:"alert_id" db:"alert_id"`
	Type        AlertActionType        `json:"type" db:"type"`
	Recipient   string                 `json:"recipient" db:"recipient"`
	Subject     string                 `json:"subject" db:"subject"`
	Content     string                 `json:"content" db:"content"`
	Status      NotificationStatus     `json:"status" db:"status"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	SentAt      *time.Time             `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt *time.Time             `json:"delivered_at,omitempty" db:"delivered_at"`
	FailedAt    *time.Time             `json:"failed_at,omitempty" db:"failed_at"`
	Error       string                 `json:"error,omitempty" db:"error"`
	RetryCount  int                    `json:"retry_count" db:"retry_count"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// NotificationStatus defines the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSending   NotificationStatus = "sending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusRetrying  NotificationStatus = "retrying"
)

// AlertTemplate defines a template for alert messages
type AlertTemplate struct {
	ID          string            `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Type        AlertType         `json:"type" db:"type"`
	Language    string            `json:"language" db:"language"`
	Subject     string            `json:"subject" db:"subject"`
	Body        string            `json:"body" db:"body"`
	Variables   []string          `json:"variables" db:"variables"`
	IsDefault   bool              `json:"is_default" db:"is_default"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// AlertStatistics tracks alert statistics
type AlertStatistics struct {
	StreamID           string                    `json:"stream_id"`
	ZoneID             string                    `json:"zone_id,omitempty"`
	TotalAlerts        int64                     `json:"total_alerts"`
	ActiveAlerts       int64                     `json:"active_alerts"`
	ResolvedAlerts     int64                     `json:"resolved_alerts"`
	AlertsByType       map[AlertType]int64       `json:"alerts_by_type"`
	AlertsBySeverity   map[AlertSeverity]int64   `json:"alerts_by_severity"`
	AlertsByHour       map[int]int64             `json:"alerts_by_hour"`
	AlertsByDay        map[string]int64          `json:"alerts_by_day"`
	AverageResolutionTime time.Duration          `json:"average_resolution_time"`
	LastAlert          *time.Time                `json:"last_alert"`
	StartTime          time.Time                 `json:"start_time"`
}

// Repository interfaces

// AlertRepository defines the interface for alert data access
type AlertRepository interface {
	// Alert management
	CreateAlert(alert *Alert) error
	GetAlert(id string) (*Alert, error)
	GetAlerts(filters AlertFilters) ([]*Alert, error)
	UpdateAlert(alert *Alert) error
	DeleteAlert(id string) error

	// Alert rules
	CreateAlertRule(rule *AlertRule) error
	GetAlertRule(id string) (*AlertRule, error)
	GetAlertRules(streamID string) ([]*AlertRule, error)
	UpdateAlertRule(rule *AlertRule) error
	DeleteAlertRule(id string) error

	// Notifications
	CreateNotification(notification *Notification) error
	GetNotification(id string) (*Notification, error)
	GetNotifications(alertID string) ([]*Notification, error)
	UpdateNotification(notification *Notification) error

	// Templates
	CreateAlertTemplate(template *AlertTemplate) error
	GetAlertTemplate(id string) (*AlertTemplate, error)
	GetAlertTemplates(alertType AlertType, language string) ([]*AlertTemplate, error)
	UpdateAlertTemplate(template *AlertTemplate) error
	DeleteAlertTemplate(id string) error

	// Statistics
	GetAlertStatistics(streamID, zoneID string) (*AlertStatistics, error)
	UpdateAlertStatistics(stats *AlertStatistics) error
}

// AlertService defines the interface for alert business logic
type AlertService interface {
	// Alert management
	CreateAlert(alert *Alert) error
	GetAlert(id string) (*Alert, error)
	GetAlerts(filters AlertFilters) ([]*Alert, error)
	UpdateAlert(alert *Alert) error
	ResolveAlert(id, resolvedBy string) error
	AcknowledgeAlert(id, acknowledgedBy string) error
	SuppressAlert(id string, duration time.Duration) error

	// Alert rules
	CreateAlertRule(rule *AlertRule) error
	GetAlertRule(id string) (*AlertRule, error)
	GetAlertRules(streamID string) ([]*AlertRule, error)
	UpdateAlertRule(rule *AlertRule) error
	DeleteAlertRule(id string) error

	// Alert processing
	ProcessZoneEvent(event *ZoneEvent) ([]*Alert, error)
	ProcessDetection(detection *DetectedObject) ([]*Alert, error)
	EvaluateAlertRules(streamID string, context AlertContext) ([]*Alert, error)

	// Notifications
	SendNotification(alert *Alert, action AlertAction) error
	GetNotifications(alertID string) ([]*Notification, error)
	RetryFailedNotifications() error

	// Templates
	RenderAlert(alert *Alert, template *AlertTemplate) (string, string, error)
	GetAlertTemplate(alertType AlertType, language string) (*AlertTemplate, error)

	// Statistics
	GetAlertStatistics(streamID, zoneID string) (*AlertStatistics, error)
	GenerateAlertReport(filters AlertFilters, reportType ReportType) (*AlertReport, error)
}

// NotificationService defines the interface for notification delivery
type AlertNotificationService interface {
	SendEmail(notification *Notification) error
	SendSMS(notification *Notification) error
	SendWebhook(notification *Notification) error
	SendSlack(notification *Notification) error
	SendDiscord(notification *Notification) error
	SendTelegram(notification *Notification) error
	SendPushover(notification *Notification) error
	SendCustom(notification *Notification) error
}

// Supporting types

// AlertFilters defines filters for querying alerts
type AlertFilters struct {
	StreamID    string        `json:"stream_id,omitempty"`
	ZoneID      string        `json:"zone_id,omitempty"`
	Type        AlertType     `json:"type,omitempty"`
	Severity    AlertSeverity `json:"severity,omitempty"`
	Status      AlertStatus   `json:"status,omitempty"`
	ObjectClass string        `json:"object_class,omitempty"`
	StartTime   *time.Time    `json:"start_time,omitempty"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	Tags        []string      `json:"tags,omitempty"`
	Limit       int           `json:"limit,omitempty"`
	Offset      int           `json:"offset,omitempty"`
}

// AlertContext provides context for alert rule evaluation
type AlertContext struct {
	StreamID     string                 `json:"stream_id"`
	ZoneID       string                 `json:"zone_id,omitempty"`
	Detections   []*DetectedObject      `json:"detections,omitempty"`
	ZoneEvents   []*ZoneEvent           `json:"zone_events,omitempty"`
	Occupancy    int                    `json:"occupancy,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// AlertReport represents a generated alert report
type AlertReport struct {
	ID          string                 `json:"id"`
	ReportType  ReportType             `json:"report_type"`
	Filters     AlertFilters           `json:"filters"`
	Data        map[string]interface{} `json:"data"`
	GeneratedAt time.Time              `json:"generated_at"`
	GeneratedBy string                 `json:"generated_by"`
}

// Validation methods

// Validate validates an alert
func (a *Alert) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("alert ID is required")
	}
	if a.Type == "" {
		return fmt.Errorf("alert type is required")
	}
	if a.Severity == "" {
		return fmt.Errorf("alert severity is required")
	}
	if a.Title == "" {
		return fmt.Errorf("alert title is required")
	}
	if a.StreamID == "" {
		return fmt.Errorf("stream ID is required")
	}
	return nil
}

// Validate validates an alert rule
func (ar *AlertRule) Validate() error {
	if ar.ID == "" {
		return fmt.Errorf("alert rule ID is required")
	}
	if ar.Name == "" {
		return fmt.Errorf("alert rule name is required")
	}
	if ar.StreamID == "" {
		return fmt.Errorf("stream ID is required")
	}
	if ar.Type == "" {
		return fmt.Errorf("alert type is required")
	}
	if ar.Severity == "" {
		return fmt.Errorf("alert severity is required")
	}
	if len(ar.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}
	return ar.Conditions.Validate()
}

// Validate validates alert conditions
func (ac *AlertConditions) Validate() error {
	if ac.MinConfidence < 0 || ac.MinConfidence > 1 {
		return fmt.Errorf("min confidence must be between 0 and 1")
	}
	if ac.MaxConfidence < 0 || ac.MaxConfidence > 1 {
		return fmt.Errorf("max confidence must be between 0 and 1")
	}
	if ac.MaxConfidence > 0 && ac.MaxConfidence < ac.MinConfidence {
		return fmt.Errorf("max confidence must be greater than min confidence")
	}
	if ac.MinObjects < 0 {
		return fmt.Errorf("min objects cannot be negative")
	}
	if ac.MaxObjects > 0 && ac.MaxObjects < ac.MinObjects {
		return fmt.Errorf("max objects must be greater than min objects")
	}
	if ac.DwellTimeMin < 0 {
		return fmt.Errorf("min dwell time cannot be negative")
	}
	if ac.DwellTimeMax > 0 && ac.DwellTimeMax < ac.DwellTimeMin {
		return fmt.Errorf("max dwell time must be greater than min dwell time")
	}
	return nil
}

// IsExpired checks if an alert has expired based on its creation time and type
func (a *Alert) IsExpired(expirationRules map[AlertType]time.Duration) bool {
	if expiration, exists := expirationRules[a.Type]; exists {
		return time.Since(a.CreatedAt) > expiration
	}
	return false
}

// GetPriority returns a numeric priority for the alert based on severity
func (a *Alert) GetPriority() int {
	switch a.Severity {
	case AlertSeverityCritical:
		return 4
	case AlertSeverityHigh:
		return 3
	case AlertSeverityMedium:
		return 2
	case AlertSeverityLow:
		return 1
	default:
		return 0
	}
}

// CanTrigger checks if an alert rule can be triggered based on cooldown
func (ar *AlertRule) CanTrigger() bool {
	if ar.LastTriggered == nil {
		return true
	}
	return time.Since(*ar.LastTriggered) >= ar.Cooldown
}

// MatchesConditions checks if the given context matches the alert conditions
func (ac *AlertConditions) MatchesConditions(context AlertContext) bool {
	// Object class filter
	if len(ac.ObjectClasses) > 0 && len(context.Detections) > 0 {
		hasMatchingClass := false
		for _, detection := range context.Detections {
			for _, class := range ac.ObjectClasses {
				if detection.Class == class {
					hasMatchingClass = true
					break
				}
			}
			if hasMatchingClass {
				break
			}
		}
		if !hasMatchingClass {
			return false
		}
	}

	// Object count filters
	objectCount := len(context.Detections)
	if ac.MinObjects > 0 && objectCount < ac.MinObjects {
		return false
	}
	if ac.MaxObjects > 0 && objectCount > ac.MaxObjects {
		return false
	}

	// Confidence filters
	if len(context.Detections) > 0 {
		for _, detection := range context.Detections {
			if ac.MinConfidence > 0 && detection.Confidence < ac.MinConfidence {
				return false
			}
			if ac.MaxConfidence > 0 && detection.Confidence > ac.MaxConfidence {
				return false
			}
		}
	}

	// Zone event filters
	if len(ac.RequiredZoneEvents) > 0 && len(context.ZoneEvents) > 0 {
		hasRequiredEvent := false
		for _, event := range context.ZoneEvents {
			for _, requiredType := range ac.RequiredZoneEvents {
				if event.EventType == requiredType {
					hasRequiredEvent = true
					break
				}
			}
			if hasRequiredEvent {
				break
			}
		}
		if !hasRequiredEvent {
			return false
		}
	}

	// Time window filters
	if len(ac.TimeWindows) > 0 {
		inTimeWindow := false
		for _, window := range ac.TimeWindows {
			// Check if current time is in any of the time windows
			now := context.Timestamp
			weekday := int(now.Weekday())
			timeStr := now.Format("15:04")

			dayMatch := false
			for _, day := range window.Days {
				if day == weekday {
					dayMatch = true
					break
				}
			}

			if dayMatch && timeStr >= window.StartTime && timeStr <= window.EndTime {
				inTimeWindow = true
				break
			}
		}
		if !inTimeWindow {
			return false
		}
	}

	return true
}
