package risk

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultRiskManagerConfig returns default risk manager configuration
func GetDefaultRiskManagerConfig() RiskManagerConfig {
	return RiskManagerConfig{
		Enabled:        true,
		UpdateInterval: 5 * time.Minute,
		CacheTimeout:   30 * time.Minute,
		AlertThresholds: AlertThresholds{
			TransactionRisk:   decimal.NewFromFloat(70),
			PortfolioRisk:     decimal.NewFromFloat(65),
			VolatilityRisk:    decimal.NewFromFloat(60),
			ContractRisk:      decimal.NewFromFloat(75),
			MarketRisk:        decimal.NewFromFloat(70),
			ConcentrationRisk: decimal.NewFromFloat(0.7),
			LiquidityRisk:     decimal.NewFromFloat(0.8),
		},
		TransactionRiskConfig: TransactionRiskConfig{
			Enabled:         true,
			ModelPath:       "/models/transaction_risk.pkl",
			ThresholdHigh:   decimal.NewFromFloat(80),
			ThresholdMedium: decimal.NewFromFloat(50),
			UpdateInterval:  1 * time.Hour,
			FeatureWeights: map[string]float64{
				"address_age":        0.2,
				"transaction_amount": 0.3,
				"gas_price":          0.1,
				"recipient_risk":     0.25,
				"time_of_day":        0.05,
				"frequency":          0.1,
			},
			AddressWhitelist: []string{
				"0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1", // Example trusted address
			},
			AddressBlacklist: []string{
				"0x0000000000000000000000000000000000000000", // Null address
			},
			DataSources: []string{"etherscan", "chainalysis", "internal"},
		},
		VolatilityConfig: VolatilityConfig{
			Enabled:         true,
			WindowSize:      30, // 30 days
			ConfidenceLevel: decimal.NewFromFloat(0.95),
			UpdateInterval:  1 * time.Hour,
			DataSources:     []string{"coingecko", "binance", "coinbase"},
			VolatilityModels: []string{"garch", "ewma", "historical"},
			RiskThresholds: map[string]decimal.Decimal{
				"low":    decimal.NewFromFloat(0.2),
				"medium": decimal.NewFromFloat(0.4),
				"high":   decimal.NewFromFloat(0.6),
			},
		},
		ContractAuditConfig: ContractAuditConfig{
			Enabled:         true,
			UpdateInterval:  24 * time.Hour,
			TimeoutDuration: 5 * time.Minute,
			MaxCodeSize:     1000000, // 1MB
			AuditRules: []string{
				"reentrancy",
				"integer_overflow",
				"access_control",
				"unchecked_calls",
				"gas_limit",
				"timestamp_dependence",
			},
			SecurityPatterns: []string{
				"checks_effects_interactions",
				"pull_over_push",
				"rate_limiting",
				"circuit_breaker",
				"access_control",
			},
			VulnerabilityDB: "https://api.mythx.io/v1/",
		},
		PortfolioConfig: PortfolioRiskConfig{
			Enabled:           true,
			UpdateInterval:    1 * time.Hour,
			CorrelationWindow: 90, // 90 days
			VaRConfidence:     decimal.NewFromFloat(0.95),
			RiskModels:        []string{"var", "cvar", "monte_carlo", "historical"},
			DataSources:       []string{"coingecko", "defipulse", "internal"},
			RiskFactors: []string{
				"concentration",
				"correlation",
				"liquidity",
				"volatility",
				"market_cap",
				"smart_contract",
			},
		},
		MarketPredictionConfig: MarketPredictionConfig{
			Enabled:             true,
			UpdateInterval:      30 * time.Minute,
			PredictionHorizon:   24 * time.Hour,
			ConfidenceThreshold: decimal.NewFromFloat(0.7),
			ModelTypes:          []string{"lstm", "transformer", "ensemble"},
			DataSources:         []string{"coingecko", "fear_greed", "social_sentiment", "on_chain"},
			Features: []string{
				"price_history",
				"volume",
				"market_cap",
				"social_sentiment",
				"fear_greed_index",
				"on_chain_metrics",
				"technical_indicators",
			},
		},
		MLModelPaths: map[string]string{
			"transaction_risk": "/models/transaction_risk.pkl",
			"volatility":       "/models/volatility.pkl",
			"market_prediction": "/models/market_prediction.pkl",
			"portfolio_risk":   "/models/portfolio_risk.pkl",
		},
		DataSources: map[string]DataSourceConfig{
			"coingecko": {
				Enabled:   true,
				URL:       "https://api.coingecko.com/api/v3",
				RateLimit: 50,
				Timeout:   30 * time.Second,
				Priority:  1,
			},
			"etherscan": {
				Enabled:   true,
				URL:       "https://api.etherscan.io/api",
				APIKey:    "YOUR_ETHERSCAN_API_KEY",
				RateLimit: 5,
				Timeout:   30 * time.Second,
				Priority:  2,
			},
			"chainalysis": {
				Enabled:   false, // Requires enterprise license
				URL:       "https://api.chainalysis.com",
				APIKey:    "YOUR_CHAINALYSIS_API_KEY",
				RateLimit: 10,
				Timeout:   30 * time.Second,
				Priority:  3,
			},
			"defipulse": {
				Enabled:   true,
				URL:       "https://data-api.defipulse.com/api/v1",
				APIKey:    "YOUR_DEFIPULSE_API_KEY",
				RateLimit: 100,
				Timeout:   30 * time.Second,
				Priority:  4,
			},
		},
	}
}

// ValidateRiskManagerConfig validates risk manager configuration
func ValidateRiskManagerConfig(config RiskManagerConfig) error {
	if !config.Enabled {
		return nil // Skip validation if disabled
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if config.CacheTimeout <= 0 {
		return fmt.Errorf("cache timeout must be positive")
	}

	// Validate alert thresholds
	if err := validateAlertThresholds(config.AlertThresholds); err != nil {
		return fmt.Errorf("invalid alert thresholds: %w", err)
	}

	// Validate component configs
	if err := validateTransactionRiskConfig(config.TransactionRiskConfig); err != nil {
		return fmt.Errorf("invalid transaction risk config: %w", err)
	}

	if err := validateVolatilityConfig(config.VolatilityConfig); err != nil {
		return fmt.Errorf("invalid volatility config: %w", err)
	}

	if err := validateContractAuditConfig(config.ContractAuditConfig); err != nil {
		return fmt.Errorf("invalid contract audit config: %w", err)
	}

	if err := validatePortfolioConfig(config.PortfolioConfig); err != nil {
		return fmt.Errorf("invalid portfolio config: %w", err)
	}

	if err := validateMarketPredictionConfig(config.MarketPredictionConfig); err != nil {
		return fmt.Errorf("invalid market prediction config: %w", err)
	}

	return nil
}

// validateAlertThresholds validates alert thresholds
func validateAlertThresholds(thresholds AlertThresholds) error {
	if thresholds.TransactionRisk.LessThan(decimal.Zero) || thresholds.TransactionRisk.GreaterThan(decimal.NewFromFloat(100)) {
		return fmt.Errorf("transaction risk threshold must be between 0 and 100")
	}

	if thresholds.PortfolioRisk.LessThan(decimal.Zero) || thresholds.PortfolioRisk.GreaterThan(decimal.NewFromFloat(100)) {
		return fmt.Errorf("portfolio risk threshold must be between 0 and 100")
	}

	if thresholds.VolatilityRisk.LessThan(decimal.Zero) || thresholds.VolatilityRisk.GreaterThan(decimal.NewFromFloat(100)) {
		return fmt.Errorf("volatility risk threshold must be between 0 and 100")
	}

	if thresholds.ConcentrationRisk.LessThan(decimal.Zero) || thresholds.ConcentrationRisk.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("concentration risk threshold must be between 0 and 1")
	}

	if thresholds.LiquidityRisk.LessThan(decimal.Zero) || thresholds.LiquidityRisk.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("liquidity risk threshold must be between 0 and 1")
	}

	return nil
}

// validateTransactionRiskConfig validates transaction risk configuration
func validateTransactionRiskConfig(config TransactionRiskConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if config.ThresholdHigh.LessThanOrEqual(config.ThresholdMedium) {
		return fmt.Errorf("high threshold must be greater than medium threshold")
	}

	if len(config.DataSources) == 0 {
		return fmt.Errorf("at least one data source must be specified")
	}

	return nil
}

// validateVolatilityConfig validates volatility configuration
func validateVolatilityConfig(config VolatilityConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.WindowSize <= 0 {
		return fmt.Errorf("window size must be positive")
	}

	if config.ConfidenceLevel.LessThan(decimal.Zero) || config.ConfidenceLevel.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("confidence level must be between 0 and 1")
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	return nil
}

// validateContractAuditConfig validates contract audit configuration
func validateContractAuditConfig(config ContractAuditConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if config.TimeoutDuration <= 0 {
		return fmt.Errorf("timeout duration must be positive")
	}

	if config.MaxCodeSize <= 0 {
		return fmt.Errorf("max code size must be positive")
	}

	if len(config.AuditRules) == 0 {
		return fmt.Errorf("at least one audit rule must be specified")
	}

	return nil
}

// validatePortfolioConfig validates portfolio configuration
func validatePortfolioConfig(config PortfolioRiskConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if config.CorrelationWindow <= 0 {
		return fmt.Errorf("correlation window must be positive")
	}

	if config.VaRConfidence.LessThan(decimal.Zero) || config.VaRConfidence.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("VaR confidence must be between 0 and 1")
	}

	return nil
}

// validateMarketPredictionConfig validates market prediction configuration
func validateMarketPredictionConfig(config MarketPredictionConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if config.PredictionHorizon <= 0 {
		return fmt.Errorf("prediction horizon must be positive")
	}

	if config.ConfidenceThreshold.LessThan(decimal.Zero) || config.ConfidenceThreshold.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("confidence threshold must be between 0 and 1")
	}

	return nil
}

// GetRiskLevelDescription returns description for risk levels
func GetRiskLevelDescription() map[string]string {
	return map[string]string{
		"low":      "Low risk - Normal market conditions with minimal threats",
		"medium":   "Medium risk - Some elevated risks requiring monitoring",
		"high":     "High risk - Significant risks requiring immediate attention",
		"critical": "Critical risk - Severe risks requiring urgent action",
	}
}

// GetAlertSeverityDescription returns description for alert severities
func GetAlertSeverityDescription() map[string]string {
	return map[string]string{
		"low":      "Low severity - Informational alert",
		"medium":   "Medium severity - Warning that requires attention",
		"high":     "High severity - Important issue requiring action",
		"critical": "Critical severity - Urgent issue requiring immediate action",
	}
}

// GetSupportedRiskFactors returns supported risk factors
func GetSupportedRiskFactors() []string {
	return []string{
		"concentration",
		"correlation",
		"liquidity",
		"volatility",
		"market_cap",
		"smart_contract",
		"regulatory",
		"operational",
		"counterparty",
		"technology",
	}
}

// GetSupportedMLModels returns supported ML models
func GetSupportedMLModels() []string {
	return []string{
		"linear_regression",
		"random_forest",
		"gradient_boosting",
		"neural_network",
		"lstm",
		"transformer",
		"ensemble",
		"svm",
		"naive_bayes",
		"decision_tree",
	}
}
