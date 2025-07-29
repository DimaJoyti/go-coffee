package alerts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Mock implementations for testing and demonstration

// MockRuleEngine provides mock rule engine functionality
type MockRuleEngine struct{}

func (m *MockRuleEngine) EvaluateRule(ctx context.Context, rule *AlertRule, data map[string]interface{}) (*Alert, error) {
	// Mock rule evaluation - create alert if conditions are met
	conditionResults := []ConditionResult{}
	
	for _, condition := range rule.Conditions {
		result := &ConditionResult{
			ConditionID:   condition.ID,
			Met:           true, // Mock: assume condition is met
			ActualValue:   "mock_value",
			ExpectedValue: condition.Value,
			Confidence:    decimal.NewFromFloat(0.95),
			EvaluatedAt:   time.Now(),
			Message:       fmt.Sprintf("Condition %s evaluated successfully", condition.Type),
		}
		conditionResults = append(conditionResults, *result)
	}
	
	// Create alert if any required conditions are met
	alert := &Alert{
		ID:         fmt.Sprintf("alert_%d", time.Now().Unix()),
		UserID:     rule.UserID,
		RuleID:     rule.ID,
		Type:       rule.Type,
		Priority:   rule.Priority,
		Title:      fmt.Sprintf("Alert: %s", rule.Name),
		Message:    fmt.Sprintf("Rule '%s' has been triggered", rule.Name),
		Data:       data,
		Conditions: conditionResults,
		Actions:    rule.Actions,
		Channels:   rule.Channels,
		Status:     "pending",
		CreatedAt:  time.Now(),
		Metadata:   map[string]interface{}{"rule_name": rule.Name},
	}
	
	return alert, nil
}

func (m *MockRuleEngine) CreateRule(ctx context.Context, rule *AlertRule) error {
	// Mock rule creation
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	return nil
}

func (m *MockRuleEngine) UpdateRule(ctx context.Context, rule *AlertRule) error {
	// Mock rule update
	rule.UpdatedAt = time.Now()
	return nil
}

func (m *MockRuleEngine) DeleteRule(ctx context.Context, ruleID string) error {
	// Mock rule deletion
	return nil
}

func (m *MockRuleEngine) GetUserRules(ctx context.Context, userID string) ([]*AlertRule, error) {
	// Mock user rules
	return []*AlertRule{
		{
			ID:          "rule_1",
			UserID:      userID,
			Name:        "Price Alert",
			Description: "Alert when price changes significantly",
			Type:        "price",
			Enabled:     true,
			Conditions: []Condition{
				{
					ID:       "condition_1",
					Type:     "price",
					Field:    "price",
					Operator: "gt",
					Value:    50000,
				},
			},
			Priority:  "medium",
			Channels:  []string{"email", "push"},
			Cooldown:  15 * time.Minute,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}, nil
}

// MockConditionEvaluator provides mock condition evaluation
type MockConditionEvaluator struct{}

func (m *MockConditionEvaluator) EvaluateCondition(ctx context.Context, condition *Condition, data map[string]interface{}) (*ConditionResult, error) {
	// Mock condition evaluation
	result := &ConditionResult{
		ConditionID:   condition.ID,
		Met:           true, // Mock: assume condition is met
		ActualValue:   "mock_actual_value",
		ExpectedValue: condition.Value,
		Confidence:    decimal.NewFromFloat(0.9),
		EvaluatedAt:   time.Now(),
		Message:       fmt.Sprintf("Condition %s evaluated", condition.Type),
	}
	
	return result, nil
}

func (m *MockConditionEvaluator) EvaluateConditions(ctx context.Context, conditions []Condition, data map[string]interface{}) ([]ConditionResult, error) {
	var results []ConditionResult
	
	for _, condition := range conditions {
		result, err := m.EvaluateCondition(ctx, &condition, data)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}
	
	return results, nil
}

func (m *MockConditionEvaluator) ValidateCondition(condition *Condition) error {
	if condition.Type == "" {
		return fmt.Errorf("condition type is required")
	}
	if condition.Field == "" {
		return fmt.Errorf("condition field is required")
	}
	if condition.Operator == "" {
		return fmt.Errorf("condition operator is required")
	}
	if condition.Value == nil {
		return fmt.Errorf("condition value is required")
	}
	
	return nil
}

// MockPriorityCalculator provides mock priority calculation
type MockPriorityCalculator struct{}

func (m *MockPriorityCalculator) CalculatePriority(ctx context.Context, alert *Alert, userPreferences *SubscriptionPreferences) (string, error) {
	// Mock priority calculation based on alert type
	switch alert.Type {
	case "price":
		return "medium", nil
	case "portfolio":
		return "high", nil
	case "risk":
		return "critical", nil
	case "transaction":
		return "low", nil
	default:
		return "medium", nil
	}
}

func (m *MockPriorityCalculator) GetPriorityScore(priority string) decimal.Decimal {
	scores := map[string]decimal.Decimal{
		"low":      decimal.NewFromFloat(1.0),
		"medium":   decimal.NewFromFloat(2.0),
		"high":     decimal.NewFromFloat(3.0),
		"critical": decimal.NewFromFloat(4.0),
	}
	
	if score, exists := scores[priority]; exists {
		return score
	}
	
	return decimal.NewFromFloat(2.0) // Default to medium
}

func (m *MockPriorityCalculator) AdjustPriorityForUser(priority string, userID string) string {
	// Mock user-specific priority adjustment
	return priority
}

// MockTemplateEngine provides mock template rendering
type MockTemplateEngine struct{}

func (m *MockTemplateEngine) RenderAlert(ctx context.Context, alert *Alert, template string, format string) (string, error) {
	// Mock template rendering
	switch format {
	case "text":
		return fmt.Sprintf("ALERT: %s\n%s\nPriority: %s\nTime: %s", 
			alert.Title, alert.Message, alert.Priority, alert.CreatedAt.Format(time.RFC3339)), nil
	case "html":
		return fmt.Sprintf("<h2>%s</h2><p>%s</p><p>Priority: %s</p><p>Time: %s</p>", 
			alert.Title, alert.Message, alert.Priority, alert.CreatedAt.Format(time.RFC3339)), nil
	case "markdown":
		return fmt.Sprintf("## %s\n\n%s\n\n**Priority:** %s  \n**Time:** %s", 
			alert.Title, alert.Message, alert.Priority, alert.CreatedAt.Format(time.RFC3339)), nil
	default:
		return alert.Message, nil
	}
}

func (m *MockTemplateEngine) GetTemplate(alertType string, channel string) (string, error) {
	// Mock template retrieval
	templates := map[string]string{
		"price":       "Price Alert: {{.Title}} - {{.Message}}",
		"portfolio":   "Portfolio Alert: {{.Title}} - {{.Message}}",
		"risk":        "Risk Alert: {{.Title}} - {{.Message}}",
		"transaction": "Transaction Alert: {{.Title}} - {{.Message}}",
	}
	
	if template, exists := templates[alertType]; exists {
		return template, nil
	}
	
	return "Alert: {{.Title}} - {{.Message}}", nil
}

func (m *MockTemplateEngine) RegisterTemplate(name string, template string) error {
	// Mock template registration
	return nil
}

// MockNotificationManager provides mock notification management
type MockNotificationManager struct{}

func (m *MockNotificationManager) SendNotification(ctx context.Context, alert *Alert, channels []string) error {
	// Mock notification sending
	for _, channel := range channels {
		switch channel {
		case "email":
			// Mock email sending
		case "sms":
			// Mock SMS sending
		case "push":
			// Mock push notification
		case "webhook":
			// Mock webhook call
		}
	}
	
	return nil
}

func (m *MockNotificationManager) SendBatchNotifications(ctx context.Context, alerts []*Alert) error {
	// Mock batch notification sending
	for _, alert := range alerts {
		if err := m.SendNotification(ctx, alert, alert.Channels); err != nil {
			return err
		}
	}
	
	return nil
}

func (m *MockNotificationManager) GetDeliveryStatus(alertID string) (*DeliveryStatus, error) {
	// Mock delivery status
	return &DeliveryStatus{
		AlertID:      alertID,
		Channel:      "email",
		Status:       "delivered",
		AttemptCount: 1,
		LastAttempt:  time.Now(),
	}, nil
}

func (m *MockNotificationManager) RetryFailedNotifications(ctx context.Context) error {
	// Mock retry logic
	return nil
}

// MockChannelManager provides mock channel management
type MockChannelManager struct{}

func (m *MockChannelManager) GetChannel(channelType string) (*NotificationChannel, error) {
	// Mock channel retrieval
	channels := map[string]*NotificationChannel{
		"email": {
			Type:     "email",
			Name:     "Email Notifications",
			Enabled:  true,
			Priority: 1,
			Config: map[string]interface{}{
				"smtp_server": "smtp.example.com",
				"port":        587,
			},
		},
		"sms": {
			Type:     "sms",
			Name:     "SMS Notifications",
			Enabled:  true,
			Priority: 2,
			Config: map[string]interface{}{
				"provider": "twilio",
			},
		},
		"push": {
			Type:     "push",
			Name:     "Push Notifications",
			Enabled:  true,
			Priority: 3,
			Config: map[string]interface{}{
				"provider": "firebase",
			},
		},
		"webhook": {
			Type:     "webhook",
			Name:     "Webhook Notifications",
			Enabled:  true,
			Priority: 4,
			Config: map[string]interface{}{
				"timeout": "30s",
			},
		},
	}
	
	if channel, exists := channels[channelType]; exists {
		return channel, nil
	}
	
	return nil, fmt.Errorf("channel not found: %s", channelType)
}

func (m *MockChannelManager) GetUserChannels(userID string) ([]*NotificationChannel, error) {
	// Mock user channels
	channels := []*NotificationChannel{}
	
	// Get all available channels for mock
	channelTypes := []string{"email", "sms", "push", "webhook"}
	for _, channelType := range channelTypes {
		if channel, err := m.GetChannel(channelType); err == nil {
			channels = append(channels, channel)
		}
	}
	
	return channels, nil
}

func (m *MockChannelManager) RegisterChannel(channel *NotificationChannel) error {
	// Mock channel registration
	return nil
}

func (m *MockChannelManager) TestChannel(channelType string, config map[string]interface{}) error {
	// Mock channel testing
	return nil
}

// MockMarketDataProvider provides mock market data
type MockMarketDataProvider struct{}

func (m *MockMarketDataProvider) GetMarketData(ctx context.Context, symbols []string) (map[string]*MarketData, error) {
	// Mock market data
	data := make(map[string]*MarketData)
	
	for _, symbol := range symbols {
		data[symbol] = &MarketData{
			Symbol:                strings.ToUpper(symbol),
			Price:                 decimal.NewFromFloat(50000 + float64(len(symbol)*1000)), // Mock price
			Volume24h:             decimal.NewFromFloat(1000000000),
			MarketCap:             decimal.NewFromFloat(1000000000000),
			PriceChange24h:        decimal.NewFromFloat(1000),
			PriceChangePercent24h: decimal.NewFromFloat(2.5),
			Timestamp:             time.Now(),
			Source:                "mock_provider",
		}
	}
	
	return data, nil
}

func (m *MockMarketDataProvider) GetPriceData(ctx context.Context, symbol string) (*MarketData, error) {
	// Mock single price data
	return &MarketData{
		Symbol:                strings.ToUpper(symbol),
		Price:                 decimal.NewFromFloat(50000),
		Volume24h:             decimal.NewFromFloat(1000000000),
		MarketCap:             decimal.NewFromFloat(1000000000000),
		PriceChange24h:        decimal.NewFromFloat(1000),
		PriceChangePercent24h: decimal.NewFromFloat(2.5),
		Timestamp:             time.Now(),
		Source:                "mock_provider",
	}, nil
}

func (m *MockMarketDataProvider) SubscribeToUpdates(ctx context.Context, symbols []string, callback func(*MarketData)) error {
	// Mock subscription
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, symbol := range symbols {
					data, _ := m.GetPriceData(ctx, symbol)
					callback(data)
				}
			}
		}
	}()
	
	return nil
}

// MockPortfolioTracker provides mock portfolio tracking
type MockPortfolioTracker struct{}

func (m *MockPortfolioTracker) GetPortfolioData(ctx context.Context, userID string) (*PortfolioData, error) {
	// Mock portfolio data
	return &PortfolioData{
		UserID:                userID,
		TotalValue:            decimal.NewFromFloat(100000),
		TotalChange24h:        decimal.NewFromFloat(2500),
		TotalChangePercent24h: decimal.NewFromFloat(2.5),
		Holdings: []Holding{
			{
				Symbol:           "BTC",
				Amount:           decimal.NewFromFloat(1.5),
				Value:            decimal.NewFromFloat(75000),
				Change24h:        decimal.NewFromFloat(1500),
				ChangePercent24h: decimal.NewFromFloat(2.0),
				Weight:           decimal.NewFromFloat(75.0),
			},
			{
				Symbol:           "ETH",
				Amount:           decimal.NewFromFloat(10),
				Value:            decimal.NewFromFloat(25000),
				Change24h:        decimal.NewFromFloat(1000),
				ChangePercent24h: decimal.NewFromFloat(4.0),
				Weight:           decimal.NewFromFloat(25.0),
			},
		},
		Timestamp: time.Now(),
	}, nil
}

func (m *MockPortfolioTracker) GetHolding(ctx context.Context, userID string, symbol string) (*Holding, error) {
	// Mock single holding
	return &Holding{
		Symbol:           strings.ToUpper(symbol),
		Amount:           decimal.NewFromFloat(1.0),
		Value:            decimal.NewFromFloat(50000),
		Change24h:        decimal.NewFromFloat(1000),
		ChangePercent24h: decimal.NewFromFloat(2.0),
		Weight:           decimal.NewFromFloat(50.0),
	}, nil
}

func (m *MockPortfolioTracker) UpdatePortfolio(ctx context.Context, userID string, portfolio *PortfolioData) error {
	// Mock portfolio update
	return nil
}

// MockRiskMonitor provides mock risk monitoring
type MockRiskMonitor struct{}

func (m *MockRiskMonitor) GetRiskData(ctx context.Context, userID string) (*RiskData, error) {
	// Mock risk data
	return &RiskData{
		UserID:            userID,
		RiskScore:         decimal.NewFromFloat(6.5),
		VaR95:             decimal.NewFromFloat(5000),
		MaxDrawdown:       decimal.NewFromFloat(15.5),
		Volatility:        decimal.NewFromFloat(25.0),
		SharpeRatio:       decimal.NewFromFloat(1.2),
		LiquidationRisk:   decimal.NewFromFloat(2.0),
		ConcentrationRisk: decimal.NewFromFloat(30.0),
		Timestamp:         time.Now(),
	}, nil
}

func (m *MockRiskMonitor) CalculateRiskMetrics(ctx context.Context, portfolio *PortfolioData) (*RiskData, error) {
	// Mock risk calculation
	return m.GetRiskData(ctx, portfolio.UserID)
}

func (m *MockRiskMonitor) MonitorRiskThresholds(ctx context.Context, userID string) ([]*Alert, error) {
	// Mock risk threshold monitoring
	return []*Alert{}, nil
}

// MockAlertStore provides mock alert storage
type MockAlertStore struct{}

func (m *MockAlertStore) SaveAlert(ctx context.Context, alert *Alert) error {
	// Mock alert saving
	return nil
}

func (m *MockAlertStore) GetAlert(ctx context.Context, alertID string) (*Alert, error) {
	// Mock alert retrieval
	return &Alert{
		ID:        alertID,
		UserID:    "user_123",
		Type:      "price",
		Priority:  "medium",
		Title:     "Mock Alert",
		Message:   "This is a mock alert",
		Status:    "sent",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}, nil
}

func (m *MockAlertStore) GetUserAlerts(ctx context.Context, userID string, filters map[string]interface{}) ([]*Alert, error) {
	// Mock user alerts
	return []*Alert{
		{
			ID:        "alert_1",
			UserID:    userID,
			Type:      "price",
			Priority:  "medium",
			Title:     "Price Alert",
			Message:   "BTC price has increased by 5%",
			Status:    "sent",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "alert_2",
			UserID:    userID,
			Type:      "portfolio",
			Priority:  "high",
			Title:     "Portfolio Alert",
			Message:   "Portfolio value has decreased by 10%",
			Status:    "acknowledged",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}, nil
}

func (m *MockAlertStore) UpdateAlertStatus(ctx context.Context, alertID string, status string) error {
	// Mock status update
	return nil
}

func (m *MockAlertStore) DeleteExpiredAlerts(ctx context.Context) error {
	// Mock cleanup
	return nil
}

// MockSubscriptionManager provides mock subscription management
type MockSubscriptionManager struct{}

func (m *MockSubscriptionManager) CreateSubscription(ctx context.Context, subscription *Subscription) error {
	// Mock subscription creation
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = time.Now()
	return nil
}

func (m *MockSubscriptionManager) UpdateSubscription(ctx context.Context, subscription *Subscription) error {
	// Mock subscription update
	subscription.UpdatedAt = time.Now()
	return nil
}

func (m *MockSubscriptionManager) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	// Mock subscription deletion
	return nil
}

func (m *MockSubscriptionManager) GetUserSubscriptions(ctx context.Context, userID string) ([]*Subscription, error) {
	// Mock user subscriptions
	return []*Subscription{
		{
			ID:       "sub_1",
			UserID:   userID,
			Type:     "price_alerts",
			Channels: []string{"email", "push"},
			Preferences: SubscriptionPreferences{
				MinPriority:      "medium",
				MaxAlertsPerHour: 10,
				QuietHours: []TimeRange{
					{
						Start:      time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC),
						End:        time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
						DaysOfWeek: []int{0, 1, 2, 3, 4, 5, 6}, // All days
					},
				},
				PreferredChannels:      []string{"email"},
				GroupSimilarAlerts:     true,
				IncludeRecommendations: true,
			},
			Enabled:   true,
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}, nil
}

func (m *MockSubscriptionManager) GetSubscriptionPreferences(ctx context.Context, userID string) (*SubscriptionPreferences, error) {
	// Mock user preferences
	return &SubscriptionPreferences{
		MinPriority:      "medium",
		MaxAlertsPerHour: 10,
		QuietHours: []TimeRange{
			{
				Start:      time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC),
				End:        time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
				DaysOfWeek: []int{0, 1, 2, 3, 4, 5, 6}, // All days
			},
		},
		PreferredChannels:      []string{"email", "push"},
		GroupSimilarAlerts:     true,
		IncludeRecommendations: true,
	}, nil
}
