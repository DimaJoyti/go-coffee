package alerts

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// IntelligentAlertSystem provides smart, context-aware alerts for crypto operations
type IntelligentAlertSystem struct {
	logger *logger.Logger
	config AlertSystemConfig

	// Alert processors
	ruleEngine         RuleEngine
	conditionEvaluator ConditionEvaluator
	priorityCalculator PriorityCalculator
	templateEngine     TemplateEngine

	// Notification channels
	notificationManager NotificationManager
	channelManager      ChannelManager

	// Data sources
	marketDataProvider MarketDataProvider
	portfolioTracker   PortfolioTracker
	riskMonitor        RiskMonitor

	// Alert management
	alertStore          AlertStore
	subscriptionManager SubscriptionManager

	// State management
	activeAlerts      map[string]*Alert
	alertRules        map[string]*AlertRule
	userSubscriptions map[string][]*Subscription

	// Processing state
	isRunning        bool
	processingTicker *time.Ticker
	stopChan         chan struct{}
	mutex            sync.RWMutex
	alertMutex       sync.RWMutex
}

// AlertSystemConfig holds configuration for the intelligent alert system
type AlertSystemConfig struct {
	Enabled                  bool                      `json:"enabled" yaml:"enabled"`
	ProcessingInterval       time.Duration             `json:"processing_interval" yaml:"processing_interval"`
	MaxConcurrentAlerts      int                       `json:"max_concurrent_alerts" yaml:"max_concurrent_alerts"`
	AlertRetentionPeriod     time.Duration             `json:"alert_retention_period" yaml:"alert_retention_period"`
	RuleEngineConfig         RuleEngineConfig          `json:"rule_engine_config" yaml:"rule_engine_config"`
	ConditionEvaluatorConfig ConditionEvaluatorConfig  `json:"condition_evaluator_config" yaml:"condition_evaluator_config"`
	PriorityCalculatorConfig PriorityCalculatorConfig  `json:"priority_calculator_config" yaml:"priority_calculator_config"`
	TemplateEngineConfig     TemplateEngineConfig      `json:"template_engine_config" yaml:"template_engine_config"`
	NotificationConfig       NotificationManagerConfig `json:"notification_config" yaml:"notification_config"`
	ChannelConfig            ChannelManagerConfig      `json:"channel_config" yaml:"channel_config"`
	DataSourceConfig         DataSourceConfig          `json:"data_source_config" yaml:"data_source_config"`
	StorageConfig            AlertStorageConfig        `json:"storage_config" yaml:"storage_config"`
}

// Component configurations
type RuleEngineConfig struct {
	Enabled               bool          `json:"enabled" yaml:"enabled"`
	MaxRulesPerUser       int           `json:"max_rules_per_user" yaml:"max_rules_per_user"`
	RuleEvaluationTimeout time.Duration `json:"rule_evaluation_timeout" yaml:"rule_evaluation_timeout"`
	EnableCustomRules     bool          `json:"enable_custom_rules" yaml:"enable_custom_rules"`
	PrebuiltRules         []string      `json:"prebuilt_rules" yaml:"prebuilt_rules"`
}

type ConditionEvaluatorConfig struct {
	Enabled              bool     `json:"enabled" yaml:"enabled"`
	SupportedOperators   []string `json:"supported_operators" yaml:"supported_operators"`
	SupportedDataTypes   []string `json:"supported_data_types" yaml:"supported_data_types"`
	MaxConditionsPerRule int      `json:"max_conditions_per_rule" yaml:"max_conditions_per_rule"`
	CacheResults         bool     `json:"cache_results" yaml:"cache_results"`
}

type PriorityCalculatorConfig struct {
	Enabled              bool                       `json:"enabled" yaml:"enabled"`
	PriorityLevels       []string                   `json:"priority_levels" yaml:"priority_levels"`
	PriorityWeights      map[string]decimal.Decimal `json:"priority_weights" yaml:"priority_weights"`
	UserPreferenceWeight decimal.Decimal            `json:"user_preference_weight" yaml:"user_preference_weight"`
	MarketImpactWeight   decimal.Decimal            `json:"market_impact_weight" yaml:"market_impact_weight"`
}

type TemplateEngineConfig struct {
	Enabled          bool              `json:"enabled" yaml:"enabled"`
	DefaultTemplates map[string]string `json:"default_templates" yaml:"default_templates"`
	CustomTemplates  map[string]string `json:"custom_templates" yaml:"custom_templates"`
	SupportedFormats []string          `json:"supported_formats" yaml:"supported_formats"`
	IncludeMarkdown  bool              `json:"include_markdown" yaml:"include_markdown"`
}

type NotificationManagerConfig struct {
	Enabled      bool          `json:"enabled" yaml:"enabled"`
	MaxRetries   int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay" yaml:"retry_delay"`
	BatchSize    int           `json:"batch_size" yaml:"batch_size"`
	RateLimiting bool          `json:"rate_limiting" yaml:"rate_limiting"`
}

type ChannelManagerConfig struct {
	Enabled           bool                   `json:"enabled" yaml:"enabled"`
	SupportedChannels []string               `json:"supported_channels" yaml:"supported_channels"`
	ChannelPriorities map[string]int         `json:"channel_priorities" yaml:"channel_priorities"`
	FallbackChannels  []string               `json:"fallback_channels" yaml:"fallback_channels"`
	ChannelConfigs    map[string]interface{} `json:"channel_configs" yaml:"channel_configs"`
}

type DataSourceConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	MarketDataSources []string      `json:"market_data_sources" yaml:"market_data_sources"`
	PortfolioSources  []string      `json:"portfolio_sources" yaml:"portfolio_sources"`
	RiskDataSources   []string      `json:"risk_data_sources" yaml:"risk_data_sources"`
	UpdateInterval    time.Duration `json:"update_interval" yaml:"update_interval"`
}

type AlertStorageConfig struct {
	Enabled            bool          `json:"enabled" yaml:"enabled"`
	StorageType        string        `json:"storage_type" yaml:"storage_type"`
	ConnectionString   string        `json:"connection_string" yaml:"connection_string"`
	RetentionPeriod    time.Duration `json:"retention_period" yaml:"retention_period"`
	CompressionEnabled bool          `json:"compression_enabled" yaml:"compression_enabled"`
}

// Data structures

// Alert represents a generated alert
type Alert struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	RuleID         string                 `json:"rule_id"`
	Type           string                 `json:"type"`     // "price", "portfolio", "risk", "market", "transaction"
	Priority       string                 `json:"priority"` // "low", "medium", "high", "critical"
	Title          string                 `json:"title"`
	Message        string                 `json:"message"`
	Data           map[string]interface{} `json:"data"`
	Conditions     []ConditionResult      `json:"conditions"`
	Actions        []AlertAction          `json:"actions"`
	Channels       []string               `json:"channels"`
	Status         string                 `json:"status"` // "pending", "sent", "acknowledged", "resolved"
	CreatedAt      time.Time              `json:"created_at"`
	SentAt         *time.Time             `json:"sent_at"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at"`
	ResolvedAt     *time.Time             `json:"resolved_at"`
	ExpiresAt      *time.Time             `json:"expires_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// AlertRule defines conditions for generating alerts
type AlertRule struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Type          string                 `json:"type"`
	Enabled       bool                   `json:"enabled"`
	Conditions    []Condition            `json:"conditions"`
	Actions       []AlertAction          `json:"actions"`
	Priority      string                 `json:"priority"`
	Channels      []string               `json:"channels"`
	Cooldown      time.Duration          `json:"cooldown"`
	MaxTriggers   int                    `json:"max_triggers"`
	TriggerCount  int                    `json:"trigger_count"`
	LastTriggered *time.Time             `json:"last_triggered"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Condition defines a single condition for alert rules
type Condition struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"` // "price", "volume", "balance", "percentage", "time"
	Field          string                 `json:"field"`
	Operator       string                 `json:"operator"` // "gt", "lt", "eq", "gte", "lte", "between", "contains"
	Value          interface{}            `json:"value"`
	SecondaryValue interface{}            `json:"secondary_value"` // For "between" operator
	DataSource     string                 `json:"data_source"`
	Parameters     map[string]interface{} `json:"parameters"`
	Weight         decimal.Decimal        `json:"weight"`
	Required       bool                   `json:"required"`
}

// ConditionResult represents the result of evaluating a condition
type ConditionResult struct {
	ConditionID   string          `json:"condition_id"`
	Met           bool            `json:"met"`
	ActualValue   interface{}     `json:"actual_value"`
	ExpectedValue interface{}     `json:"expected_value"`
	Confidence    decimal.Decimal `json:"confidence"`
	EvaluatedAt   time.Time       `json:"evaluated_at"`
	Message       string          `json:"message"`
}

// AlertAction defines actions to take when an alert is triggered
type AlertAction struct {
	Type       string                 `json:"type"` // "notify", "execute", "webhook", "email", "sms"
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
	Priority   int                    `json:"priority"`
	Conditions []string               `json:"conditions"` // Condition IDs that must be met
}

// Subscription represents a user's subscription to alerts
type Subscription struct {
	ID          string                  `json:"id"`
	UserID      string                  `json:"user_id"`
	Type        string                  `json:"type"`
	Channels    []string                `json:"channels"`
	Filters     map[string]interface{}  `json:"filters"`
	Preferences SubscriptionPreferences `json:"preferences"`
	Enabled     bool                    `json:"enabled"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// SubscriptionPreferences holds user preferences for alerts
type SubscriptionPreferences struct {
	MinPriority            string      `json:"min_priority"`
	MaxAlertsPerHour       int         `json:"max_alerts_per_hour"`
	QuietHours             []TimeRange `json:"quiet_hours"`
	PreferredChannels      []string    `json:"preferred_channels"`
	GroupSimilarAlerts     bool        `json:"group_similar_alerts"`
	IncludeRecommendations bool        `json:"include_recommendations"`
}

// TimeRange represents a time range for quiet hours
type TimeRange struct {
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Timezone   string    `json:"timezone"`
	DaysOfWeek []int     `json:"days_of_week"` // 0=Sunday, 1=Monday, etc.
}

// NotificationChannel represents a notification delivery channel
type NotificationChannel struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	RateLimit   *RateLimit             `json:"rate_limit"`
	RetryPolicy *RetryPolicy           `json:"retry_policy"`
}

// RateLimit defines rate limiting for notification channels
type RateLimit struct {
	MaxRequests int           `json:"max_requests"`
	TimeWindow  time.Duration `json:"time_window"`
	BurstSize   int           `json:"burst_size"`
}

// RetryPolicy defines retry behavior for failed notifications
type RetryPolicy struct {
	MaxRetries        int             `json:"max_retries"`
	InitialDelay      time.Duration   `json:"initial_delay"`
	MaxDelay          time.Duration   `json:"max_delay"`
	BackoffMultiplier decimal.Decimal `json:"backoff_multiplier"`
}

// MarketData represents market data for alert evaluation
type MarketData struct {
	Symbol                string          `json:"symbol"`
	Price                 decimal.Decimal `json:"price"`
	Volume24h             decimal.Decimal `json:"volume_24h"`
	MarketCap             decimal.Decimal `json:"market_cap"`
	PriceChange24h        decimal.Decimal `json:"price_change_24h"`
	PriceChangePercent24h decimal.Decimal `json:"price_change_percent_24h"`
	Timestamp             time.Time       `json:"timestamp"`
	Source                string          `json:"source"`
}

// PortfolioData represents portfolio data for alert evaluation
type PortfolioData struct {
	UserID                string          `json:"user_id"`
	TotalValue            decimal.Decimal `json:"total_value"`
	TotalChange24h        decimal.Decimal `json:"total_change_24h"`
	TotalChangePercent24h decimal.Decimal `json:"total_change_percent_24h"`
	Holdings              []Holding       `json:"holdings"`
	Timestamp             time.Time       `json:"timestamp"`
}

// Holding represents a single asset holding
type Holding struct {
	Symbol           string          `json:"symbol"`
	Amount           decimal.Decimal `json:"amount"`
	Value            decimal.Decimal `json:"value"`
	Change24h        decimal.Decimal `json:"change_24h"`
	ChangePercent24h decimal.Decimal `json:"change_percent_24h"`
	Weight           decimal.Decimal `json:"weight"`
}

// RiskData represents risk metrics for alert evaluation
type RiskData struct {
	UserID            string          `json:"user_id"`
	RiskScore         decimal.Decimal `json:"risk_score"`
	VaR95             decimal.Decimal `json:"var_95"`
	MaxDrawdown       decimal.Decimal `json:"max_drawdown"`
	Volatility        decimal.Decimal `json:"volatility"`
	SharpeRatio       decimal.Decimal `json:"sharpe_ratio"`
	LiquidationRisk   decimal.Decimal `json:"liquidation_risk"`
	ConcentrationRisk decimal.Decimal `json:"concentration_risk"`
	Timestamp         time.Time       `json:"timestamp"`
}

// Component interfaces
type RuleEngine interface {
	EvaluateRule(ctx context.Context, rule *AlertRule, data map[string]interface{}) (*Alert, error)
	CreateRule(ctx context.Context, rule *AlertRule) error
	UpdateRule(ctx context.Context, rule *AlertRule) error
	DeleteRule(ctx context.Context, ruleID string) error
	GetUserRules(ctx context.Context, userID string) ([]*AlertRule, error)
}

type ConditionEvaluator interface {
	EvaluateCondition(ctx context.Context, condition *Condition, data map[string]interface{}) (*ConditionResult, error)
	EvaluateConditions(ctx context.Context, conditions []Condition, data map[string]interface{}) ([]ConditionResult, error)
	ValidateCondition(condition *Condition) error
}

type PriorityCalculator interface {
	CalculatePriority(ctx context.Context, alert *Alert, userPreferences *SubscriptionPreferences) (string, error)
	GetPriorityScore(priority string) decimal.Decimal
	AdjustPriorityForUser(priority string, userID string) string
}

type TemplateEngine interface {
	RenderAlert(ctx context.Context, alert *Alert, template string, format string) (string, error)
	GetTemplate(alertType string, channel string) (string, error)
	RegisterTemplate(name string, template string) error
}

type NotificationManager interface {
	SendNotification(ctx context.Context, alert *Alert, channels []string) error
	SendBatchNotifications(ctx context.Context, alerts []*Alert) error
	GetDeliveryStatus(alertID string) (*DeliveryStatus, error)
	RetryFailedNotifications(ctx context.Context) error
}

type ChannelManager interface {
	GetChannel(channelType string) (*NotificationChannel, error)
	GetUserChannels(userID string) ([]*NotificationChannel, error)
	RegisterChannel(channel *NotificationChannel) error
	TestChannel(channelType string, config map[string]interface{}) error
}

type MarketDataProvider interface {
	GetMarketData(ctx context.Context, symbols []string) (map[string]*MarketData, error)
	GetPriceData(ctx context.Context, symbol string) (*MarketData, error)
	SubscribeToUpdates(ctx context.Context, symbols []string, callback func(*MarketData)) error
}

type PortfolioTracker interface {
	GetPortfolioData(ctx context.Context, userID string) (*PortfolioData, error)
	GetHolding(ctx context.Context, userID string, symbol string) (*Holding, error)
	UpdatePortfolio(ctx context.Context, userID string, portfolio *PortfolioData) error
}

type RiskMonitor interface {
	GetRiskData(ctx context.Context, userID string) (*RiskData, error)
	CalculateRiskMetrics(ctx context.Context, portfolio *PortfolioData) (*RiskData, error)
	MonitorRiskThresholds(ctx context.Context, userID string) ([]*Alert, error)
}

type AlertStore interface {
	SaveAlert(ctx context.Context, alert *Alert) error
	GetAlert(ctx context.Context, alertID string) (*Alert, error)
	GetUserAlerts(ctx context.Context, userID string, filters map[string]interface{}) ([]*Alert, error)
	UpdateAlertStatus(ctx context.Context, alertID string, status string) error
	DeleteExpiredAlerts(ctx context.Context) error
}

type SubscriptionManager interface {
	CreateSubscription(ctx context.Context, subscription *Subscription) error
	UpdateSubscription(ctx context.Context, subscription *Subscription) error
	DeleteSubscription(ctx context.Context, subscriptionID string) error
	GetUserSubscriptions(ctx context.Context, userID string) ([]*Subscription, error)
	GetSubscriptionPreferences(ctx context.Context, userID string) (*SubscriptionPreferences, error)
}

// Supporting types
type DeliveryStatus struct {
	AlertID      string     `json:"alert_id"`
	Channel      string     `json:"channel"`
	Status       string     `json:"status"` // "pending", "sent", "failed", "delivered"
	AttemptCount int        `json:"attempt_count"`
	LastAttempt  time.Time  `json:"last_attempt"`
	NextRetry    *time.Time `json:"next_retry"`
	Error        string     `json:"error"`
}

// NewIntelligentAlertSystem creates a new intelligent alert system
func NewIntelligentAlertSystem(logger *logger.Logger, config AlertSystemConfig) *IntelligentAlertSystem {
	ias := &IntelligentAlertSystem{
		logger:            logger.Named("intelligent-alert-system"),
		config:            config,
		activeAlerts:      make(map[string]*Alert),
		alertRules:        make(map[string]*AlertRule),
		userSubscriptions: make(map[string][]*Subscription),
		stopChan:          make(chan struct{}),
	}

	// Initialize components (mock implementations for this example)
	ias.initializeComponents()

	return ias
}

// initializeComponents initializes all alert system components
func (ias *IntelligentAlertSystem) initializeComponents() {
	// Initialize components with mock implementations
	// In production, these would be real implementations
	ias.ruleEngine = &MockRuleEngine{}
	ias.conditionEvaluator = &MockConditionEvaluator{}
	ias.priorityCalculator = &MockPriorityCalculator{}
	ias.templateEngine = &MockTemplateEngine{}
	ias.notificationManager = &MockNotificationManager{}
	ias.channelManager = &MockChannelManager{}
	ias.marketDataProvider = &MockMarketDataProvider{}
	ias.portfolioTracker = &MockPortfolioTracker{}
	ias.riskMonitor = &MockRiskMonitor{}
	ias.alertStore = &MockAlertStore{}
	ias.subscriptionManager = &MockSubscriptionManager{}
}

// Start starts the intelligent alert system
func (ias *IntelligentAlertSystem) Start(ctx context.Context) error {
	ias.mutex.Lock()
	defer ias.mutex.Unlock()

	if ias.isRunning {
		return fmt.Errorf("intelligent alert system is already running")
	}

	if !ias.config.Enabled {
		ias.logger.Info("Intelligent alert system is disabled")
		return nil
	}

	ias.logger.Info("Starting intelligent alert system",
		zap.Duration("processing_interval", ias.config.ProcessingInterval),
		zap.Int("max_concurrent_alerts", ias.config.MaxConcurrentAlerts))

	// Start processing routine
	ias.processingTicker = time.NewTicker(ias.config.ProcessingInterval)
	go ias.processingLoop(ctx)

	// Start cleanup routine
	go ias.cleanupLoop(ctx)

	ias.isRunning = true
	ias.logger.Info("Intelligent alert system started successfully")
	return nil
}

// Stop stops the intelligent alert system
func (ias *IntelligentAlertSystem) Stop() error {
	ias.mutex.Lock()
	defer ias.mutex.Unlock()

	if !ias.isRunning {
		return nil
	}

	ias.logger.Info("Stopping intelligent alert system")

	// Stop processing
	if ias.processingTicker != nil {
		ias.processingTicker.Stop()
	}
	close(ias.stopChan)

	ias.isRunning = false
	ias.logger.Info("Intelligent alert system stopped")
	return nil
}

// CreateAlertRule creates a new alert rule
func (ias *IntelligentAlertSystem) CreateAlertRule(ctx context.Context, rule *AlertRule) error {
	ias.logger.Debug("Creating alert rule",
		zap.String("rule_id", rule.ID),
		zap.String("user_id", rule.UserID),
		zap.String("type", rule.Type))

	// Validate rule
	if err := ias.validateAlertRule(rule); err != nil {
		return fmt.Errorf("invalid alert rule: %w", err)
	}

	// Check user rule limits
	userRules, err := ias.ruleEngine.GetUserRules(ctx, rule.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user rules: %w", err)
	}

	if len(userRules) >= ias.config.RuleEngineConfig.MaxRulesPerUser {
		return fmt.Errorf("user has reached maximum number of rules (%d)", ias.config.RuleEngineConfig.MaxRulesPerUser)
	}

	// Create rule
	if err := ias.ruleEngine.CreateRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}

	// Cache rule
	ias.alertMutex.Lock()
	ias.alertRules[rule.ID] = rule
	ias.alertMutex.Unlock()

	ias.logger.Info("Alert rule created successfully",
		zap.String("rule_id", rule.ID),
		zap.String("user_id", rule.UserID))

	return nil
}

// UpdateAlertRule updates an existing alert rule
func (ias *IntelligentAlertSystem) UpdateAlertRule(ctx context.Context, rule *AlertRule) error {
	ias.logger.Debug("Updating alert rule",
		zap.String("rule_id", rule.ID),
		zap.String("user_id", rule.UserID))

	// Validate rule
	if err := ias.validateAlertRule(rule); err != nil {
		return fmt.Errorf("invalid alert rule: %w", err)
	}

	// Update rule
	if err := ias.ruleEngine.UpdateRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	// Update cache
	ias.alertMutex.Lock()
	ias.alertRules[rule.ID] = rule
	ias.alertMutex.Unlock()

	ias.logger.Info("Alert rule updated successfully",
		zap.String("rule_id", rule.ID))

	return nil
}

// DeleteAlertRule deletes an alert rule
func (ias *IntelligentAlertSystem) DeleteAlertRule(ctx context.Context, ruleID string) error {
	ias.logger.Debug("Deleting alert rule", zap.String("rule_id", ruleID))

	// Delete rule
	if err := ias.ruleEngine.DeleteRule(ctx, ruleID); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	// Remove from cache
	ias.alertMutex.Lock()
	delete(ias.alertRules, ruleID)
	ias.alertMutex.Unlock()

	ias.logger.Info("Alert rule deleted successfully", zap.String("rule_id", ruleID))
	return nil
}

// GetUserAlertRules gets all alert rules for a user
func (ias *IntelligentAlertSystem) GetUserAlertRules(ctx context.Context, userID string) ([]*AlertRule, error) {
	ias.logger.Debug("Getting user alert rules", zap.String("user_id", userID))

	rules, err := ias.ruleEngine.GetUserRules(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rules: %w", err)
	}

	return rules, nil
}

// CreateSubscription creates a new alert subscription
func (ias *IntelligentAlertSystem) CreateSubscription(ctx context.Context, subscription *Subscription) error {
	ias.logger.Debug("Creating subscription",
		zap.String("subscription_id", subscription.ID),
		zap.String("user_id", subscription.UserID),
		zap.String("type", subscription.Type))

	// Validate subscription
	if err := ias.validateSubscription(subscription); err != nil {
		return fmt.Errorf("invalid subscription: %w", err)
	}

	// Create subscription
	if err := ias.subscriptionManager.CreateSubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	// Update cache
	ias.alertMutex.Lock()
	if ias.userSubscriptions[subscription.UserID] == nil {
		ias.userSubscriptions[subscription.UserID] = []*Subscription{}
	}
	ias.userSubscriptions[subscription.UserID] = append(ias.userSubscriptions[subscription.UserID], subscription)
	ias.alertMutex.Unlock()

	ias.logger.Info("Subscription created successfully",
		zap.String("subscription_id", subscription.ID),
		zap.String("user_id", subscription.UserID))

	return nil
}

// ProcessAlerts processes all active alert rules
func (ias *IntelligentAlertSystem) ProcessAlerts(ctx context.Context) error {
	startTime := time.Now()

	ias.logger.Debug("Processing alerts")

	// Get all active rules
	allRules := ias.getAllActiveRules()
	if len(allRules) == 0 {
		return nil
	}

	// Gather data for evaluation
	data, err := ias.gatherEvaluationData(ctx, allRules)
	if err != nil {
		ias.logger.Warn("Failed to gather evaluation data", zap.Error(err))
		return err
	}

	// Process rules concurrently
	alertChan := make(chan *Alert, ias.config.MaxConcurrentAlerts)
	errorChan := make(chan error, len(allRules))

	// Start workers
	workerCount := min(ias.config.MaxConcurrentAlerts, len(allRules))
	ruleChan := make(chan *AlertRule, len(allRules))

	for i := 0; i < workerCount; i++ {
		go ias.alertWorker(ctx, ruleChan, alertChan, errorChan, data)
	}

	// Send rules to workers
	for _, rule := range allRules {
		ruleChan <- rule
	}
	close(ruleChan)

	// Collect results
	var generatedAlerts []*Alert
	var processingErrors []error

	for i := 0; i < len(allRules); i++ {
		select {
		case alert := <-alertChan:
			if alert != nil {
				generatedAlerts = append(generatedAlerts, alert)
			}
		case err := <-errorChan:
			if err != nil {
				processingErrors = append(processingErrors, err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Send generated alerts
	if len(generatedAlerts) > 0 {
		if err := ias.sendAlerts(ctx, generatedAlerts); err != nil {
			ias.logger.Error("Failed to send alerts", zap.Error(err))
		}
	}

	// Log processing results
	ias.logger.Info("Alert processing completed",
		zap.Int("rules_processed", len(allRules)),
		zap.Int("alerts_generated", len(generatedAlerts)),
		zap.Int("errors", len(processingErrors)),
		zap.Duration("processing_time", time.Since(startTime)))

	return nil
}

// SendAlert sends a single alert
func (ias *IntelligentAlertSystem) SendAlert(ctx context.Context, alert *Alert) error {
	ias.logger.Debug("Sending alert",
		zap.String("alert_id", alert.ID),
		zap.String("user_id", alert.UserID),
		zap.String("type", alert.Type),
		zap.String("priority", alert.Priority))

	// Calculate priority
	userPrefs, err := ias.subscriptionManager.GetSubscriptionPreferences(ctx, alert.UserID)
	if err != nil {
		ias.logger.Warn("Failed to get user preferences", zap.Error(err))
		userPrefs = &SubscriptionPreferences{} // Use defaults
	}

	priority, err := ias.priorityCalculator.CalculatePriority(ctx, alert, userPrefs)
	if err != nil {
		ias.logger.Warn("Failed to calculate priority", zap.Error(err))
	} else {
		alert.Priority = priority
	}

	// Check if alert meets user's minimum priority
	if !ias.meetsMinimumPriority(alert.Priority, userPrefs.MinPriority) {
		ias.logger.Debug("Alert below minimum priority threshold",
			zap.String("alert_priority", alert.Priority),
			zap.String("min_priority", userPrefs.MinPriority))
		return nil
	}

	// Check quiet hours
	if ias.isInQuietHours(userPrefs.QuietHours) {
		ias.logger.Debug("Alert suppressed due to quiet hours")
		// Store for later delivery
		alert.Status = "pending"
		return ias.alertStore.SaveAlert(ctx, alert)
	}

	// Render alert message
	if err := ias.renderAlertMessage(ctx, alert); err != nil {
		ias.logger.Warn("Failed to render alert message", zap.Error(err))
	}

	// Send notification
	if err := ias.notificationManager.SendNotification(ctx, alert, alert.Channels); err != nil {
		ias.logger.Error("Failed to send notification", zap.Error(err))
		alert.Status = "failed"
	} else {
		alert.Status = "sent"
		now := time.Now()
		alert.SentAt = &now
	}

	// Save alert
	if err := ias.alertStore.SaveAlert(ctx, alert); err != nil {
		ias.logger.Error("Failed to save alert", zap.Error(err))
	}

	// Update active alerts
	ias.alertMutex.Lock()
	ias.activeAlerts[alert.ID] = alert
	ias.alertMutex.Unlock()

	return nil
}

// GetUserAlerts gets alerts for a user with optional filters
func (ias *IntelligentAlertSystem) GetUserAlerts(ctx context.Context, userID string, filters map[string]interface{}) ([]*Alert, error) {
	ias.logger.Debug("Getting user alerts",
		zap.String("user_id", userID),
		zap.Any("filters", filters))

	alerts, err := ias.alertStore.GetUserAlerts(ctx, userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get user alerts: %w", err)
	}

	return alerts, nil
}

// AcknowledgeAlert acknowledges an alert
func (ias *IntelligentAlertSystem) AcknowledgeAlert(ctx context.Context, alertID string, userID string) error {
	ias.logger.Debug("Acknowledging alert",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID))

	// Get alert
	alert, err := ias.alertStore.GetAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	// Verify ownership
	if alert.UserID != userID {
		return fmt.Errorf("alert does not belong to user")
	}

	// Update status
	if err := ias.alertStore.UpdateAlertStatus(ctx, alertID, "acknowledged"); err != nil {
		return fmt.Errorf("failed to update alert status: %w", err)
	}

	// Update cache
	ias.alertMutex.Lock()
	if cachedAlert, exists := ias.activeAlerts[alertID]; exists {
		cachedAlert.Status = "acknowledged"
		now := time.Now()
		cachedAlert.AcknowledgedAt = &now
	}
	ias.alertMutex.Unlock()

	ias.logger.Info("Alert acknowledged successfully", zap.String("alert_id", alertID))
	return nil
}

// Helper methods

func (ias *IntelligentAlertSystem) validateAlertRule(rule *AlertRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if rule.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if len(rule.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}
	if len(rule.Conditions) > ias.config.ConditionEvaluatorConfig.MaxConditionsPerRule {
		return fmt.Errorf("too many conditions (max: %d)", ias.config.ConditionEvaluatorConfig.MaxConditionsPerRule)
	}

	// Validate conditions
	for _, condition := range rule.Conditions {
		if err := ias.conditionEvaluator.ValidateCondition(&condition); err != nil {
			return fmt.Errorf("invalid condition: %w", err)
		}
	}

	return nil
}

func (ias *IntelligentAlertSystem) validateSubscription(subscription *Subscription) error {
	if subscription.ID == "" {
		return fmt.Errorf("subscription ID is required")
	}
	if subscription.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if subscription.Type == "" {
		return fmt.Errorf("subscription type is required")
	}
	if len(subscription.Channels) == 0 {
		return fmt.Errorf("at least one channel is required")
	}

	// Validate channels
	for _, channel := range subscription.Channels {
		if _, err := ias.channelManager.GetChannel(channel); err != nil {
			return fmt.Errorf("invalid channel %s: %w", channel, err)
		}
	}

	return nil
}

func (ias *IntelligentAlertSystem) getAllActiveRules() []*AlertRule {
	ias.alertMutex.RLock()
	defer ias.alertMutex.RUnlock()

	var activeRules []*AlertRule
	for _, rule := range ias.alertRules {
		if rule.Enabled && !ias.isRuleOnCooldown(rule) {
			activeRules = append(activeRules, rule)
		}
	}

	return activeRules
}

func (ias *IntelligentAlertSystem) isRuleOnCooldown(rule *AlertRule) bool {
	if rule.LastTriggered == nil || rule.Cooldown == 0 {
		return false
	}

	return time.Since(*rule.LastTriggered) < rule.Cooldown
}

func (ias *IntelligentAlertSystem) gatherEvaluationData(ctx context.Context, rules []*AlertRule) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Collect unique symbols and user IDs
	symbolsSet := make(map[string]bool)
	userIDsSet := make(map[string]bool)

	for _, rule := range rules {
		userIDsSet[rule.UserID] = true
		for _, condition := range rule.Conditions {
			if symbol, ok := condition.Parameters["symbol"].(string); ok {
				symbolsSet[symbol] = true
			}
		}
	}

	// Convert sets to slices
	var symbols []string
	var userIDs []string
	for symbol := range symbolsSet {
		symbols = append(symbols, symbol)
	}
	for userID := range userIDsSet {
		userIDs = append(userIDs, userID)
	}

	// Gather market data
	if len(symbols) > 0 {
		marketData, err := ias.marketDataProvider.GetMarketData(ctx, symbols)
		if err != nil {
			ias.logger.Warn("Failed to get market data", zap.Error(err))
		} else {
			data["market_data"] = marketData
		}
	}

	// Gather portfolio data
	portfolioData := make(map[string]*PortfolioData)
	for _, userID := range userIDs {
		portfolio, err := ias.portfolioTracker.GetPortfolioData(ctx, userID)
		if err != nil {
			ias.logger.Warn("Failed to get portfolio data", zap.String("user_id", userID), zap.Error(err))
		} else {
			portfolioData[userID] = portfolio
		}
	}
	data["portfolio_data"] = portfolioData

	// Gather risk data
	riskData := make(map[string]*RiskData)
	for _, userID := range userIDs {
		risk, err := ias.riskMonitor.GetRiskData(ctx, userID)
		if err != nil {
			ias.logger.Warn("Failed to get risk data", zap.String("user_id", userID), zap.Error(err))
		} else {
			riskData[userID] = risk
		}
	}
	data["risk_data"] = riskData

	return data, nil
}

func (ias *IntelligentAlertSystem) alertWorker(ctx context.Context, ruleChan <-chan *AlertRule, alertChan chan<- *Alert, errorChan chan<- error, data map[string]interface{}) {
	for rule := range ruleChan {
		alert, err := ias.ruleEngine.EvaluateRule(ctx, rule, data)
		if err != nil {
			errorChan <- fmt.Errorf("failed to evaluate rule %s: %w", rule.ID, err)
			continue
		}

		alertChan <- alert
	}
}

func (ias *IntelligentAlertSystem) sendAlerts(ctx context.Context, alerts []*Alert) error {
	for _, alert := range alerts {
		if err := ias.SendAlert(ctx, alert); err != nil {
			ias.logger.Error("Failed to send alert",
				zap.String("alert_id", alert.ID),
				zap.Error(err))
		}
	}
	return nil
}

func (ias *IntelligentAlertSystem) meetsMinimumPriority(alertPriority, minPriority string) bool {
	priorityLevels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	alertLevel, ok1 := priorityLevels[alertPriority]
	minLevel, ok2 := priorityLevels[minPriority]

	if !ok1 || !ok2 {
		return true // Default to allowing if priority levels are unknown
	}

	return alertLevel >= minLevel
}

func (ias *IntelligentAlertSystem) isInQuietHours(quietHours []TimeRange) bool {
	now := time.Now()

	for _, timeRange := range quietHours {
		// Check if current day of week is in the range
		currentDay := int(now.Weekday())
		dayMatches := false

		if len(timeRange.DaysOfWeek) == 0 {
			dayMatches = true // No specific days means all days
		} else {
			for _, day := range timeRange.DaysOfWeek {
				if day == currentDay {
					dayMatches = true
					break
				}
			}
		}

		if !dayMatches {
			continue
		}

		// Check time range
		currentTime := now.Format("15:04")
		startTime := timeRange.Start.Format("15:04")
		endTime := timeRange.End.Format("15:04")

		if startTime <= endTime {
			// Same day range
			if currentTime >= startTime && currentTime <= endTime {
				return true
			}
		} else {
			// Overnight range
			if currentTime >= startTime || currentTime <= endTime {
				return true
			}
		}
	}

	return false
}

func (ias *IntelligentAlertSystem) renderAlertMessage(ctx context.Context, alert *Alert) error {
	// Get appropriate template
	template, err := ias.templateEngine.GetTemplate(alert.Type, alert.Channels[0])
	if err != nil {
		ias.logger.Warn("Failed to get template", zap.Error(err))
		return nil // Use default message
	}

	// Render message
	renderedMessage, err := ias.templateEngine.RenderAlert(ctx, alert, template, "text")
	if err != nil {
		ias.logger.Warn("Failed to render alert message", zap.Error(err))
		return nil // Use default message
	}

	alert.Message = renderedMessage
	return nil
}

func (ias *IntelligentAlertSystem) processingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ias.stopChan:
			return
		case <-ias.processingTicker.C:
			if err := ias.ProcessAlerts(ctx); err != nil {
				ias.logger.Error("Error processing alerts", zap.Error(err))
			}
		}
	}
}

func (ias *IntelligentAlertSystem) cleanupLoop(ctx context.Context) {
	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ias.stopChan:
			return
		case <-cleanupTicker.C:
			if err := ias.alertStore.DeleteExpiredAlerts(ctx); err != nil {
				ias.logger.Error("Error cleaning up expired alerts", zap.Error(err))
			}
		}
	}
}

// IsRunning returns whether the alert system is running
func (ias *IntelligentAlertSystem) IsRunning() bool {
	ias.mutex.RLock()
	defer ias.mutex.RUnlock()
	return ias.isRunning
}

// GetMetrics returns alert system metrics
func (ias *IntelligentAlertSystem) GetMetrics() map[string]interface{} {
	ias.alertMutex.RLock()
	defer ias.alertMutex.RUnlock()

	return map[string]interface{}{
		"is_running":                   ias.IsRunning(),
		"active_alerts_count":          len(ias.activeAlerts),
		"alert_rules_count":            len(ias.alertRules),
		"user_subscriptions_count":     len(ias.userSubscriptions),
		"processing_interval":          ias.config.ProcessingInterval.String(),
		"max_concurrent_alerts":        ias.config.MaxConcurrentAlerts,
		"alert_retention_period":       ias.config.AlertRetentionPeriod.String(),
		"rule_engine_enabled":          ias.config.RuleEngineConfig.Enabled,
		"notification_manager_enabled": ias.config.NotificationConfig.Enabled,
		"channel_manager_enabled":      ias.config.ChannelConfig.Enabled,
		"data_sources_enabled":         ias.config.DataSourceConfig.Enabled,
		"storage_enabled":              ias.config.StorageConfig.Enabled,
	}
}

// Utility function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
