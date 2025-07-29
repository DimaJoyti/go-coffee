package alerts

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultAlertSystemConfig returns default alert system configuration
func GetDefaultAlertSystemConfig() AlertSystemConfig {
	return AlertSystemConfig{
		Enabled:              true,
		ProcessingInterval:   30 * time.Second,
		MaxConcurrentAlerts:  10,
		AlertRetentionPeriod: 30 * 24 * time.Hour, // 30 days
		RuleEngineConfig: RuleEngineConfig{
			Enabled:               true,
			MaxRulesPerUser:       50,
			RuleEvaluationTimeout: 10 * time.Second,
			EnableCustomRules:     true,
			PrebuiltRules: []string{
				"price_change_alert", "portfolio_value_alert", "risk_threshold_alert",
				"transaction_alert", "liquidation_warning", "yield_opportunity",
			},
		},
		ConditionEvaluatorConfig: ConditionEvaluatorConfig{
			Enabled: true,
			SupportedOperators: []string{
				"gt", "gte", "lt", "lte", "eq", "ne", "between", "contains",
				"starts_with", "ends_with", "regex", "in", "not_in",
			},
			SupportedDataTypes: []string{
				"number", "string", "boolean", "datetime", "percentage",
				"currency", "address", "hash",
			},
			MaxConditionsPerRule: 10,
			CacheResults:         true,
		},
		PriorityCalculatorConfig: PriorityCalculatorConfig{
			Enabled:        true,
			PriorityLevels: []string{"low", "medium", "high", "critical"},
			PriorityWeights: map[string]decimal.Decimal{
				"market_impact":     decimal.NewFromFloat(0.3),
				"user_preference":   decimal.NewFromFloat(0.25),
				"risk_level":        decimal.NewFromFloat(0.2),
				"portfolio_impact":  decimal.NewFromFloat(0.15),
				"time_sensitivity":  decimal.NewFromFloat(0.1),
			},
			UserPreferenceWeight: decimal.NewFromFloat(0.4),
			MarketImpactWeight:   decimal.NewFromFloat(0.6),
		},
		TemplateEngineConfig: TemplateEngineConfig{
			Enabled: true,
			DefaultTemplates: map[string]string{
				"price_alert":       "Price Alert: {{.Symbol}} is now {{.Price}} ({{.ChangePercent}}%)",
				"portfolio_alert":   "Portfolio Alert: Total value is {{.TotalValue}} ({{.ChangePercent}}%)",
				"risk_alert":        "Risk Alert: Risk score is {{.RiskScore}} - {{.Message}}",
				"transaction_alert": "Transaction Alert: {{.Type}} transaction {{.Status}}",
			},
			CustomTemplates:  map[string]string{},
			SupportedFormats: []string{"text", "html", "markdown", "json"},
			IncludeMarkdown:  true,
		},
		NotificationConfig: NotificationManagerConfig{
			Enabled:      true,
			MaxRetries:   3,
			RetryDelay:   5 * time.Second,
			BatchSize:    100,
			RateLimiting: true,
		},
		ChannelConfig: ChannelManagerConfig{
			Enabled:           true,
			SupportedChannels: []string{"email", "sms", "push", "webhook", "slack", "discord", "telegram"},
			ChannelPriorities: map[string]int{
				"push":     1, // Highest priority
				"sms":      2,
				"email":    3,
				"webhook":  4,
				"slack":    5,
				"discord":  6,
				"telegram": 7,
			},
			FallbackChannels: []string{"email", "push"},
			ChannelConfigs: map[string]interface{}{
				"email": map[string]interface{}{
					"smtp_server": "smtp.gmail.com",
					"port":        587,
					"use_tls":     true,
				},
				"sms": map[string]interface{}{
					"provider": "twilio",
					"timeout":  "30s",
				},
				"push": map[string]interface{}{
					"provider": "firebase",
					"timeout":  "10s",
				},
				"webhook": map[string]interface{}{
					"timeout":     "30s",
					"max_retries": 3,
				},
			},
		},
		DataSourceConfig: DataSourceConfig{
			Enabled:           true,
			MarketDataSources: []string{"coingecko", "coinmarketcap", "binance", "kraken"},
			PortfolioSources:  []string{"internal", "debank", "zapper"},
			RiskDataSources:   []string{"internal", "defipulse", "messari"},
			UpdateInterval:    1 * time.Minute,
		},
		StorageConfig: AlertStorageConfig{
			Enabled:            true,
			StorageType:        "postgresql",
			ConnectionString:   "postgres://localhost/crypto_wallet",
			RetentionPeriod:    90 * 24 * time.Hour, // 90 days
			CompressionEnabled: true,
		},
	}
}

// GetHighFrequencyAlertConfig returns configuration for high-frequency trading alerts
func GetHighFrequencyAlertConfig() AlertSystemConfig {
	config := GetDefaultAlertSystemConfig()
	
	// High-frequency settings
	config.ProcessingInterval = 1 * time.Second
	config.MaxConcurrentAlerts = 50
	
	// More aggressive rule engine
	config.RuleEngineConfig.MaxRulesPerUser = 200
	config.RuleEngineConfig.RuleEvaluationTimeout = 1 * time.Second
	
	// More conditions per rule for complex strategies
	config.ConditionEvaluatorConfig.MaxConditionsPerRule = 20
	
	// Faster data updates
	config.DataSourceConfig.UpdateInterval = 5 * time.Second
	
	// Larger batch sizes for efficiency
	config.NotificationConfig.BatchSize = 500
	config.NotificationConfig.RetryDelay = 1 * time.Second
	
	return config
}

// GetConservativeAlertConfig returns configuration for conservative users
func GetConservativeAlertConfig() AlertSystemConfig {
	config := GetDefaultAlertSystemConfig()
	
	// Conservative settings
	config.ProcessingInterval = 5 * time.Minute
	config.MaxConcurrentAlerts = 5
	
	// Limited rules for simplicity
	config.RuleEngineConfig.MaxRulesPerUser = 10
	config.RuleEngineConfig.EnableCustomRules = false
	config.RuleEngineConfig.PrebuiltRules = []string{
		"price_change_alert", "portfolio_value_alert", "risk_threshold_alert",
	}
	
	// Simpler conditions
	config.ConditionEvaluatorConfig.MaxConditionsPerRule = 3
	config.ConditionEvaluatorConfig.SupportedOperators = []string{
		"gt", "lt", "eq", "between",
	}
	
	// Adjust priority weights for conservative approach
	config.PriorityCalculatorConfig.PriorityWeights = map[string]decimal.Decimal{
		"risk_level":       decimal.NewFromFloat(0.4),
		"user_preference":  decimal.NewFromFloat(0.3),
		"market_impact":    decimal.NewFromFloat(0.2),
		"portfolio_impact": decimal.NewFromFloat(0.1),
	}
	
	// Fewer notification channels
	config.ChannelConfig.SupportedChannels = []string{"email", "push"}
	config.ChannelConfig.FallbackChannels = []string{"email"}
	
	// Less frequent data updates
	config.DataSourceConfig.UpdateInterval = 5 * time.Minute
	
	return config
}

// GetEnterpriseAlertConfig returns configuration for enterprise users
func GetEnterpriseAlertConfig() AlertSystemConfig {
	config := GetDefaultAlertSystemConfig()
	
	// Enterprise settings
	config.ProcessingInterval = 10 * time.Second
	config.MaxConcurrentAlerts = 100
	config.AlertRetentionPeriod = 365 * 24 * time.Hour // 1 year
	
	// Enterprise rule engine
	config.RuleEngineConfig.MaxRulesPerUser = 1000
	config.RuleEngineConfig.RuleEvaluationTimeout = 30 * time.Second
	config.RuleEngineConfig.PrebuiltRules = append(config.RuleEngineConfig.PrebuiltRules,
		"compliance_alert", "audit_alert", "performance_alert", "anomaly_detection",
		"correlation_alert", "volume_spike_alert", "whale_movement_alert",
	)
	
	// Advanced conditions
	config.ConditionEvaluatorConfig.MaxConditionsPerRule = 50
	config.ConditionEvaluatorConfig.SupportedOperators = append(
		config.ConditionEvaluatorConfig.SupportedOperators,
		"correlation", "moving_average", "bollinger_bands", "rsi", "macd",
	)
	config.ConditionEvaluatorConfig.SupportedDataTypes = append(
		config.ConditionEvaluatorConfig.SupportedDataTypes,
		"technical_indicator", "sentiment_score", "social_metric",
	)
	
	// All notification channels
	config.ChannelConfig.SupportedChannels = append(
		config.ChannelConfig.SupportedChannels,
		"pagerduty", "opsgenie", "msteams", "jira",
	)
	
	// Enterprise data sources
	config.DataSourceConfig.MarketDataSources = append(
		config.DataSourceConfig.MarketDataSources,
		"bloomberg", "refinitiv", "alpha_vantage", "quandl",
	)
	config.DataSourceConfig.RiskDataSources = append(
		config.DataSourceConfig.RiskDataSources,
		"riskmetrics", "msci", "factset",
	)
	
	// Enterprise storage
	config.StorageConfig.StorageType = "postgresql_cluster"
	config.StorageConfig.RetentionPeriod = 7 * 365 * 24 * time.Hour // 7 years
	config.StorageConfig.CompressionEnabled = true
	
	return config
}

// ValidateAlertSystemConfig validates alert system configuration
func ValidateAlertSystemConfig(config AlertSystemConfig) error {
	if !config.Enabled {
		return nil
	}
	
	// Validate processing interval
	if config.ProcessingInterval <= 0 {
		return fmt.Errorf("processing interval must be positive")
	}
	
	// Validate concurrent alerts
	if config.MaxConcurrentAlerts <= 0 {
		return fmt.Errorf("max concurrent alerts must be positive")
	}
	
	// Validate retention period
	if config.AlertRetentionPeriod <= 0 {
		return fmt.Errorf("alert retention period must be positive")
	}
	
	// Validate rule engine config
	if config.RuleEngineConfig.Enabled {
		if config.RuleEngineConfig.MaxRulesPerUser <= 0 {
			return fmt.Errorf("max rules per user must be positive")
		}
		
		if config.RuleEngineConfig.RuleEvaluationTimeout <= 0 {
			return fmt.Errorf("rule evaluation timeout must be positive")
		}
	}
	
	// Validate condition evaluator config
	if config.ConditionEvaluatorConfig.Enabled {
		if config.ConditionEvaluatorConfig.MaxConditionsPerRule <= 0 {
			return fmt.Errorf("max conditions per rule must be positive")
		}
		
		if len(config.ConditionEvaluatorConfig.SupportedOperators) == 0 {
			return fmt.Errorf("at least one supported operator must be specified")
		}
		
		if len(config.ConditionEvaluatorConfig.SupportedDataTypes) == 0 {
			return fmt.Errorf("at least one supported data type must be specified")
		}
	}
	
	// Validate priority calculator config
	if config.PriorityCalculatorConfig.Enabled {
		if len(config.PriorityCalculatorConfig.PriorityLevels) == 0 {
			return fmt.Errorf("at least one priority level must be specified")
		}
		
		// Validate priority weights sum to approximately 1.0
		totalWeight := decimal.Zero
		for _, weight := range config.PriorityCalculatorConfig.PriorityWeights {
			totalWeight = totalWeight.Add(weight)
		}
		if totalWeight.LessThan(decimal.NewFromFloat(0.9)) || totalWeight.GreaterThan(decimal.NewFromFloat(1.1)) {
			return fmt.Errorf("priority weights should sum to approximately 1.0, got %s", totalWeight.String())
		}
	}
	
	// Validate template engine config
	if config.TemplateEngineConfig.Enabled {
		if len(config.TemplateEngineConfig.SupportedFormats) == 0 {
			return fmt.Errorf("at least one supported format must be specified")
		}
	}
	
	// Validate notification config
	if config.NotificationConfig.Enabled {
		if config.NotificationConfig.MaxRetries < 0 {
			return fmt.Errorf("max retries cannot be negative")
		}
		
		if config.NotificationConfig.RetryDelay <= 0 {
			return fmt.Errorf("retry delay must be positive")
		}
		
		if config.NotificationConfig.BatchSize <= 0 {
			return fmt.Errorf("batch size must be positive")
		}
	}
	
	// Validate channel config
	if config.ChannelConfig.Enabled {
		if len(config.ChannelConfig.SupportedChannels) == 0 {
			return fmt.Errorf("at least one supported channel must be specified")
		}
		
		// Validate fallback channels are in supported channels
		for _, fallback := range config.ChannelConfig.FallbackChannels {
			found := false
			for _, supported := range config.ChannelConfig.SupportedChannels {
				if fallback == supported {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("fallback channel %s is not in supported channels", fallback)
			}
		}
	}
	
	// Validate data source config
	if config.DataSourceConfig.Enabled {
		if config.DataSourceConfig.UpdateInterval <= 0 {
			return fmt.Errorf("data source update interval must be positive")
		}
	}
	
	// Validate storage config
	if config.StorageConfig.Enabled {
		if config.StorageConfig.StorageType == "" {
			return fmt.Errorf("storage type must be specified")
		}
		
		if config.StorageConfig.RetentionPeriod <= 0 {
			return fmt.Errorf("storage retention period must be positive")
		}
	}
	
	return nil
}

// GetSupportedAlertTypes returns supported alert types
func GetSupportedAlertTypes() []string {
	return []string{
		"price", "portfolio", "risk", "transaction", "market",
		"liquidation", "yield", "governance", "security", "compliance",
		"performance", "anomaly", "correlation", "volume", "whale_movement",
	}
}

// GetSupportedConditionOperators returns supported condition operators
func GetSupportedConditionOperators() []string {
	return []string{
		"gt", "gte", "lt", "lte", "eq", "ne", "between", "contains",
		"starts_with", "ends_with", "regex", "in", "not_in",
		"correlation", "moving_average", "bollinger_bands", "rsi", "macd",
	}
}

// GetSupportedDataTypes returns supported data types for conditions
func GetSupportedDataTypes() []string {
	return []string{
		"number", "string", "boolean", "datetime", "percentage",
		"currency", "address", "hash", "technical_indicator",
		"sentiment_score", "social_metric",
	}
}

// GetSupportedNotificationChannels returns supported notification channels
func GetSupportedNotificationChannels() []string {
	return []string{
		"email", "sms", "push", "webhook", "slack", "discord", "telegram",
		"pagerduty", "opsgenie", "msteams", "jira",
	}
}

// GetSupportedPriorityLevels returns supported priority levels
func GetSupportedPriorityLevels() []string {
	return []string{
		"low", "medium", "high", "critical",
	}
}

// GetDefaultPriorityWeights returns default priority calculation weights
func GetDefaultPriorityWeights() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"market_impact":     decimal.NewFromFloat(0.3),
		"user_preference":   decimal.NewFromFloat(0.25),
		"risk_level":        decimal.NewFromFloat(0.2),
		"portfolio_impact":  decimal.NewFromFloat(0.15),
		"time_sensitivity":  decimal.NewFromFloat(0.1),
	}
}

// GetDefaultTemplates returns default alert templates
func GetDefaultTemplates() map[string]string {
	return map[string]string{
		"price_alert":       "🚨 Price Alert: {{.Symbol}} is now ${{.Price}} ({{.ChangePercent}}% change)",
		"portfolio_alert":   "📊 Portfolio Alert: Total value is ${{.TotalValue}} ({{.ChangePercent}}% change)",
		"risk_alert":        "⚠️ Risk Alert: Risk score is {{.RiskScore}}/10 - {{.Message}}",
		"transaction_alert": "💳 Transaction Alert: {{.Type}} transaction {{.Status}} - {{.Hash}}",
		"liquidation_alert": "🔴 Liquidation Warning: Position at risk - {{.Message}}",
		"yield_alert":       "💰 Yield Opportunity: {{.Protocol}} offering {{.APY}}% APY",
		"governance_alert":  "🗳️ Governance Alert: New proposal for {{.Protocol}} - {{.Title}}",
		"security_alert":    "🔒 Security Alert: {{.Type}} detected - {{.Message}}",
		"market_alert":      "📈 Market Alert: {{.Event}} detected - {{.Description}}",
		"whale_alert":       "🐋 Whale Alert: Large {{.Type}} transaction detected - {{.Amount}}",
	}
}

// GetChannelConfigurations returns default channel configurations
func GetChannelConfigurations() map[string]interface{} {
	return map[string]interface{}{
		"email": map[string]interface{}{
			"smtp_server": "smtp.gmail.com",
			"port":        587,
			"use_tls":     true,
			"timeout":     "30s",
		},
		"sms": map[string]interface{}{
			"provider":    "twilio",
			"timeout":     "30s",
			"max_length":  160,
		},
		"push": map[string]interface{}{
			"provider": "firebase",
			"timeout":  "10s",
			"priority": "high",
		},
		"webhook": map[string]interface{}{
			"timeout":     "30s",
			"max_retries": 3,
			"headers": map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "CryptoWallet-AlertSystem/1.0",
			},
		},
		"slack": map[string]interface{}{
			"timeout":     "15s",
			"max_retries": 2,
		},
		"discord": map[string]interface{}{
			"timeout":     "15s",
			"max_retries": 2,
		},
		"telegram": map[string]interface{}{
			"timeout":     "15s",
			"max_retries": 2,
			"parse_mode":  "Markdown",
		},
	}
}

// GetOptimalConfigForUseCase returns optimal configuration for specific use cases
func GetOptimalConfigForUseCase(useCase string) (AlertSystemConfig, error) {
	switch useCase {
	case "high_frequency":
		return GetHighFrequencyAlertConfig(), nil
	case "conservative":
		return GetConservativeAlertConfig(), nil
	case "enterprise":
		return GetEnterpriseAlertConfig(), nil
	case "default":
		return GetDefaultAlertSystemConfig(), nil
	default:
		return AlertSystemConfig{}, fmt.Errorf("unsupported use case: %s", useCase)
	}
}

// GetUseCaseDescriptions returns descriptions for use cases
func GetUseCaseDescriptions() map[string]string {
	return map[string]string{
		"high_frequency": "Optimized for high-frequency trading with fast processing and many concurrent alerts",
		"conservative":   "Simplified configuration for conservative users with basic alert functionality",
		"enterprise":     "Comprehensive configuration for enterprise users with advanced features and compliance",
		"default":        "Balanced configuration suitable for most users with standard alert functionality",
	}
}

// GetAlertTypeDescriptions returns descriptions for alert types
func GetAlertTypeDescriptions() map[string]string {
	return map[string]string{
		"price":         "Price movement alerts for specific cryptocurrencies",
		"portfolio":     "Portfolio value and performance alerts",
		"risk":          "Risk management and threshold alerts",
		"transaction":   "Transaction status and confirmation alerts",
		"market":        "Market condition and trend alerts",
		"liquidation":   "Liquidation risk and margin call alerts",
		"yield":         "Yield farming and staking opportunity alerts",
		"governance":    "Governance proposal and voting alerts",
		"security":      "Security incident and vulnerability alerts",
		"compliance":    "Regulatory and compliance alerts",
		"performance":   "System and strategy performance alerts",
		"anomaly":       "Anomaly detection and unusual activity alerts",
		"correlation":   "Asset correlation and market relationship alerts",
		"volume":        "Trading volume and liquidity alerts",
		"whale_movement": "Large transaction and whale movement alerts",
	}
}
