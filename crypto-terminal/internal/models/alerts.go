package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Alert represents a user-defined alert
type Alert struct {
	ID          string          `json:"id" db:"id"`
	UserID      string          `json:"user_id" db:"user_id"`
	Type        string          `json:"type" db:"type"` // PRICE, VOLUME, TECHNICAL, NEWS, DEFI
	Symbol      string          `json:"symbol" db:"symbol"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Condition   AlertCondition  `json:"condition" db:"condition"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	IsTriggered bool            `json:"is_triggered" db:"is_triggered"`
	TriggerCount int            `json:"trigger_count" db:"trigger_count"`
	MaxTriggers int             `json:"max_triggers" db:"max_triggers"`
	Cooldown    time.Duration   `json:"cooldown" db:"cooldown"`
	LastTriggered *time.Time    `json:"last_triggered" db:"last_triggered"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
	ExpiresAt   *time.Time      `json:"expires_at" db:"expires_at"`
	Channels    []string        `json:"channels" db:"channels"` // EMAIL, SMS, PUSH, WEBHOOK
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// AlertCondition represents the condition for triggering an alert
type AlertCondition struct {
	Operator    string          `json:"operator" db:"operator"` // ABOVE, BELOW, EQUALS, CROSSES_ABOVE, CROSSES_BELOW
	Value       decimal.Decimal `json:"value" db:"value"`
	Percentage  decimal.Decimal `json:"percentage,omitempty" db:"percentage"`
	Timeframe   string          `json:"timeframe,omitempty" db:"timeframe"`
	Indicator   string          `json:"indicator,omitempty" db:"indicator"`
	Parameters  map[string]interface{} `json:"parameters,omitempty" db:"parameters"`
}

// PriceAlert represents a price-based alert
type PriceAlert struct {
	Alert
	TargetPrice     decimal.Decimal `json:"target_price"`
	CurrentPrice    decimal.Decimal `json:"current_price"`
	PriceChange24h  decimal.Decimal `json:"price_change_24h"`
	VolumeThreshold decimal.Decimal `json:"volume_threshold,omitempty"`
}

// TechnicalAlert represents a technical analysis alert
type TechnicalAlert struct {
	Alert
	Indicator       string                 `json:"indicator"`
	Timeframe       string                 `json:"timeframe"`
	CurrentValue    decimal.Decimal        `json:"current_value"`
	TargetValue     decimal.Decimal        `json:"target_value"`
	SignalStrength  decimal.Decimal        `json:"signal_strength"`
	IndicatorValues map[string]interface{} `json:"indicator_values"`
}

// VolumeAlert represents a volume-based alert
type VolumeAlert struct {
	Alert
	VolumeThreshold decimal.Decimal `json:"volume_threshold"`
	CurrentVolume   decimal.Decimal `json:"current_volume"`
	VolumeChange    decimal.Decimal `json:"volume_change"`
	AverageVolume   decimal.Decimal `json:"average_volume"`
}

// NewsAlert represents a news-based alert
type NewsAlert struct {
	Alert
	Keywords    []string `json:"keywords"`
	Sources     []string `json:"sources"`
	Sentiment   string   `json:"sentiment"` // POSITIVE, NEGATIVE, NEUTRAL
	Relevance   decimal.Decimal `json:"relevance"`
	NewsCount   int      `json:"news_count"`
}

// DeFiAlert represents a DeFi-related alert
type DeFiAlert struct {
	Alert
	Protocol        string          `json:"protocol"`
	Pool            string          `json:"pool"`
	APYThreshold    decimal.Decimal `json:"apy_threshold"`
	CurrentAPY      decimal.Decimal `json:"current_apy"`
	TVLThreshold    decimal.Decimal `json:"tvl_threshold"`
	CurrentTVL      decimal.Decimal `json:"current_tvl"`
	ImpermanentLoss decimal.Decimal `json:"impermanent_loss"`
}

// AlertTrigger represents an alert trigger event
type AlertTrigger struct {
	ID          string          `json:"id" db:"id"`
	AlertID     string          `json:"alert_id" db:"alert_id"`
	TriggerValue decimal.Decimal `json:"trigger_value" db:"trigger_value"`
	ActualValue decimal.Decimal `json:"actual_value" db:"actual_value"`
	Message     string          `json:"message" db:"message"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	TriggeredAt time.Time       `json:"triggered_at" db:"triggered_at"`
	NotifiedAt  *time.Time      `json:"notified_at" db:"notified_at"`
	Status      string          `json:"status" db:"status"` // PENDING, SENT, FAILED
}

// AlertNotification represents a notification sent for an alert
type AlertNotification struct {
	ID          string    `json:"id" db:"id"`
	AlertID     string    `json:"alert_id" db:"alert_id"`
	TriggerID   string    `json:"trigger_id" db:"trigger_id"`
	Channel     string    `json:"channel" db:"channel"` // EMAIL, SMS, PUSH, WEBHOOK
	Recipient   string    `json:"recipient" db:"recipient"`
	Subject     string    `json:"subject" db:"subject"`
	Message     string    `json:"message" db:"message"`
	Status      string    `json:"status" db:"status"` // PENDING, SENT, DELIVERED, FAILED
	SentAt      *time.Time `json:"sent_at" db:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at" db:"delivered_at"`
	Error       string    `json:"error" db:"error"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AlertTemplate represents a pre-defined alert template
type AlertTemplate struct {
	ID          string         `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description" db:"description"`
	Type        string         `json:"type" db:"type"`
	Category    string         `json:"category" db:"category"`
	Template    AlertCondition `json:"template" db:"template"`
	IsPublic    bool           `json:"is_public" db:"is_public"`
	UsageCount  int            `json:"usage_count" db:"usage_count"`
	CreatedBy   string         `json:"created_by" db:"created_by"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// AlertGroup represents a group of related alerts
type AlertGroup struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	AlertCount  int       `json:"alert_count" db:"alert_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Alerts      []Alert   `json:"alerts,omitempty"`
}

// AlertStatistics represents alert statistics for a user
type AlertStatistics struct {
	UserID           string          `json:"user_id"`
	TotalAlerts      int             `json:"total_alerts"`
	ActiveAlerts     int             `json:"active_alerts"`
	TriggeredToday   int             `json:"triggered_today"`
	TriggeredWeek    int             `json:"triggered_week"`
	TriggeredMonth   int             `json:"triggered_month"`
	SuccessRate      decimal.Decimal `json:"success_rate"`
	MostTriggeredType string         `json:"most_triggered_type"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastCalculated   time.Time       `json:"last_calculated"`
}

// SmartAlert represents an AI-powered smart alert
type SmartAlert struct {
	Alert
	AIModel         string          `json:"ai_model"`
	Confidence      decimal.Decimal `json:"confidence"`
	PredictionType  string          `json:"prediction_type"` // PRICE_MOVEMENT, TREND_REVERSAL, BREAKOUT
	TimeHorizon     time.Duration   `json:"time_horizon"`
	Probability     decimal.Decimal `json:"probability"`
	HistoricalAccuracy decimal.Decimal `json:"historical_accuracy"`
	FeatureImportance map[string]decimal.Decimal `json:"feature_importance"`
	LastModelUpdate time.Time       `json:"last_model_update"`
}

// AlertWebhook represents a webhook configuration for alerts
type AlertWebhook struct {
	ID          string            `json:"id" db:"id"`
	UserID      string            `json:"user_id" db:"user_id"`
	Name        string            `json:"name" db:"name"`
	URL         string            `json:"url" db:"url"`
	Method      string            `json:"method" db:"method"` // POST, PUT
	Headers     map[string]string `json:"headers" db:"headers"`
	Template    string            `json:"template" db:"template"`
	IsActive    bool              `json:"is_active" db:"is_active"`
	RetryCount  int               `json:"retry_count" db:"retry_count"`
	Timeout     time.Duration     `json:"timeout" db:"timeout"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// AlertBacktest represents backtesting results for an alert
type AlertBacktest struct {
	ID              string          `json:"id" db:"id"`
	AlertID         string          `json:"alert_id" db:"alert_id"`
	StartDate       time.Time       `json:"start_date" db:"start_date"`
	EndDate         time.Time       `json:"end_date" db:"end_date"`
	TotalTriggers   int             `json:"total_triggers" db:"total_triggers"`
	SuccessfulTriggers int          `json:"successful_triggers" db:"successful_triggers"`
	FalsePositives  int             `json:"false_positives" db:"false_positives"`
	SuccessRate     decimal.Decimal `json:"success_rate" db:"success_rate"`
	AverageReturn   decimal.Decimal `json:"average_return" db:"average_return"`
	MaxReturn       decimal.Decimal `json:"max_return" db:"max_return"`
	MinReturn       decimal.Decimal `json:"min_return" db:"min_return"`
	Sharpe          decimal.Decimal `json:"sharpe" db:"sharpe"`
	MaxDrawdown     decimal.Decimal `json:"max_drawdown" db:"max_drawdown"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// CreateAlertRequest represents a request to create an alert
type CreateAlertRequest struct {
	Type        string         `json:"type" binding:"required"`
	Symbol      string         `json:"symbol" binding:"required"`
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	Condition   AlertCondition `json:"condition" binding:"required"`
	Channels    []string       `json:"channels" binding:"required"`
	MaxTriggers int            `json:"max_triggers"`
	Cooldown    time.Duration  `json:"cooldown"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateAlertRequest represents a request to update an alert
type UpdateAlertRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Condition   AlertCondition `json:"condition"`
	IsActive    *bool          `json:"is_active"`
	Channels    []string       `json:"channels"`
	MaxTriggers int            `json:"max_triggers"`
	Cooldown    time.Duration  `json:"cooldown"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}
