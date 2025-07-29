package gas

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GasOptimizerConfig represents the configuration for the gas optimizer
type GasOptimizerConfig struct {
	Enabled                bool                     `json:"enabled" yaml:"enabled"`
	UpdateInterval         time.Duration            `json:"update_interval" yaml:"update_interval"`
	HistoryRetentionPeriod time.Duration            `json:"history_retention_period" yaml:"history_retention_period"`
	MaxHistorySize         int                      `json:"max_history_size" yaml:"max_history_size"`
	OptimizationStrategies []string                 `json:"optimization_strategies" yaml:"optimization_strategies"`
	EIP1559Config          EIP1559OptimizerConfig   `json:"eip1559_config" yaml:"eip1559_config"`
	HistoricalConfig       HistoricalAnalyzerConfig `json:"historical_config" yaml:"historical_config"`
	CongestionConfig       CongestionMonitorConfig  `json:"congestion_config" yaml:"congestion_config"`
	PredictionConfig       PredictionEngineConfig   `json:"prediction_config" yaml:"prediction_config"`
	SafetyMargins          SafetyMarginsConfig      `json:"safety_margins" yaml:"safety_margins"`
}

// EIP1559OptimizerConfig represents the configuration for EIP-1559 optimization
type EIP1559OptimizerConfig struct {
	Enabled                bool            `json:"enabled" yaml:"enabled"`
	BaseFeeMultiplier      decimal.Decimal `json:"base_fee_multiplier" yaml:"base_fee_multiplier"`
	PriorityFeeStrategy    string          `json:"priority_fee_strategy" yaml:"priority_fee_strategy"`
	MaxFeeCapMultiplier    decimal.Decimal `json:"max_fee_cap_multiplier" yaml:"max_fee_cap_multiplier"`
	TargetConfirmationTime time.Duration   `json:"target_confirmation_time" yaml:"target_confirmation_time"`
	AggressivenessLevel    string          `json:"aggressiveness_level" yaml:"aggressiveness_level"`
}

// HistoricalAnalyzerConfig represents the configuration for historical analysis
type HistoricalAnalyzerConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	AnalysisWindow    time.Duration `json:"analysis_window" yaml:"analysis_window"`
	SampleSize        int           `json:"sample_size" yaml:"sample_size"`
	WeightingStrategy string        `json:"weighting_strategy" yaml:"weighting_strategy"`
	TrendAnalysis     bool          `json:"trend_analysis" yaml:"trend_analysis"`
}

// CongestionMonitorConfig represents the configuration for congestion monitoring
type CongestionMonitorConfig struct {
	Enabled              bool                       `json:"enabled" yaml:"enabled"`
	MonitoringInterval   time.Duration              `json:"monitoring_interval" yaml:"monitoring_interval"`
	CongestionThresholds map[string]decimal.Decimal `json:"congestion_thresholds" yaml:"congestion_thresholds"`
	AdjustmentFactors    map[string]decimal.Decimal `json:"adjustment_factors" yaml:"adjustment_factors"`
}

// PredictionEngineConfig represents the configuration for prediction engine
type PredictionEngineConfig struct {
	Enabled             bool            `json:"enabled" yaml:"enabled"`
	PredictionMethods   []string        `json:"prediction_methods" yaml:"prediction_methods"`
	ConfidenceThreshold decimal.Decimal `json:"confidence_threshold" yaml:"confidence_threshold"`
	ModelUpdateInterval time.Duration   `json:"model_update_interval" yaml:"model_update_interval"`
	TimeHorizons        []time.Duration `json:"time_horizons" yaml:"time_horizons"`
}

// SafetyMarginsConfig represents the configuration for safety margins
type SafetyMarginsConfig struct {
	MinGasPrice         decimal.Decimal `json:"min_gas_price" yaml:"min_gas_price"`
	MaxGasPrice         decimal.Decimal `json:"max_gas_price" yaml:"max_gas_price"`
	SafetyMultiplier    decimal.Decimal `json:"safety_multiplier" yaml:"safety_multiplier"`
	EmergencyMultiplier decimal.Decimal `json:"emergency_multiplier" yaml:"emergency_multiplier"`
	MaxCostIncrease     decimal.Decimal `json:"max_cost_increase" yaml:"max_cost_increase"`
}

// GetDefaultGasOptimizerConfig returns default gas optimizer configuration
func GetDefaultGasOptimizerConfig() GasOptimizerConfig {
	return GasOptimizerConfig{
		Enabled:                true,
		UpdateInterval:         30 * time.Second,
		HistoryRetentionPeriod: 2 * time.Hour,
		MaxHistorySize:         1000,
		OptimizationStrategies: []string{"eip1559", "historical", "congestion_based", "hybrid"},
		EIP1559Config: EIP1559OptimizerConfig{
			Enabled:                true,
			BaseFeeMultiplier:      decimal.NewFromFloat(1.125), // 12.5% above base fee
			PriorityFeeStrategy:    "dynamic",
			MaxFeeCapMultiplier:    decimal.NewFromFloat(2.0),
			TargetConfirmationTime: 3 * time.Minute,
			AggressivenessLevel:    "moderate",
		},
		HistoricalConfig: HistoricalAnalyzerConfig{
			Enabled:           true,
			AnalysisWindow:    30 * time.Minute,
			SampleSize:        100,
			WeightingStrategy: "weighted_recent",
			TrendAnalysis:     true,
		},
		CongestionConfig: CongestionMonitorConfig{
			Enabled:            true,
			MonitoringInterval: 1 * time.Minute,
			CongestionThresholds: map[string]decimal.Decimal{
				"low":    decimal.NewFromFloat(0.3),
				"medium": decimal.NewFromFloat(0.6),
				"high":   decimal.NewFromFloat(0.8),
			},
			AdjustmentFactors: map[string]decimal.Decimal{
				"low":    decimal.NewFromFloat(0.9),
				"medium": decimal.NewFromFloat(1.0),
				"high":   decimal.NewFromFloat(1.3),
			},
		},
		PredictionConfig: PredictionEngineConfig{
			Enabled: true,
			PredictionMethods: []string{
				"moving_average",
				"exponential_smoothing",
				"linear_regression",
			},
			TimeHorizons: []time.Duration{
				1 * time.Minute,
				5 * time.Minute,
				15 * time.Minute,
				30 * time.Minute,
			},
			ConfidenceThreshold: decimal.NewFromFloat(0.7),
		},
		SafetyMargins: SafetyMarginsConfig{
			MinGasPrice:         decimal.NewFromFloat(1),    // 1 gwei minimum
			MaxGasPrice:         decimal.NewFromFloat(500),  // 500 gwei maximum
			SafetyMultiplier:    decimal.NewFromFloat(1.05), // 5% safety margin
			EmergencyMultiplier: decimal.NewFromFloat(1.2),  // 20% for urgent transactions
		},
	}
}

// GetHighFrequencyConfig returns configuration optimized for high-frequency trading
func GetHighFrequencyConfig() GasOptimizerConfig {
	config := GetDefaultGasOptimizerConfig()

	// More frequent updates
	config.UpdateInterval = 5 * time.Second
	config.HistoryRetentionPeriod = 30 * time.Minute
	config.MaxHistorySize = 2000

	// Aggressive EIP-1559 settings
	config.EIP1559Config.BaseFeeMultiplier = decimal.NewFromFloat(1.2)
	config.EIP1559Config.PriorityFeeStrategy = "aggressive"
	config.EIP1559Config.AggressivenessLevel = "aggressive"
	config.EIP1559Config.TargetConfirmationTime = 1 * time.Minute

	// Shorter analysis windows
	config.HistoricalConfig.AnalysisWindow = 5 * time.Minute
	config.HistoricalConfig.SampleSize = 50

	// More frequent congestion monitoring
	config.CongestionConfig.MonitoringInterval = 10 * time.Second
	config.CongestionConfig.AdjustmentFactors = map[string]decimal.Decimal{
		"low":    decimal.NewFromFloat(0.95),
		"medium": decimal.NewFromFloat(1.1),
		"high":   decimal.NewFromFloat(1.5),
	}

	// Shorter prediction horizons
	config.PredictionConfig.TimeHorizons = []time.Duration{
		10 * time.Second,
		30 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
	}

	// Higher safety margins for speed
	config.SafetyMargins.SafetyMultiplier = decimal.NewFromFloat(1.1)
	config.SafetyMargins.EmergencyMultiplier = decimal.NewFromFloat(1.3)

	return config
}

// GetCostOptimizedConfig returns configuration optimized for cost efficiency
func GetCostOptimizedConfig() GasOptimizerConfig {
	config := GetDefaultGasOptimizerConfig()

	// Longer analysis periods for better cost optimization
	config.UpdateInterval = 2 * time.Minute
	config.HistoryRetentionPeriod = 6 * time.Hour
	config.MaxHistorySize = 5000

	// Conservative EIP-1559 settings
	config.EIP1559Config.BaseFeeMultiplier = decimal.NewFromFloat(1.05)
	config.EIP1559Config.PriorityFeeStrategy = "fixed"
	config.EIP1559Config.AggressivenessLevel = "conservative"
	config.EIP1559Config.TargetConfirmationTime = 10 * time.Minute

	// Longer analysis windows
	config.HistoricalConfig.AnalysisWindow = 2 * time.Hour
	config.HistoricalConfig.SampleSize = 500
	config.HistoricalConfig.WeightingStrategy = "simple_average"

	// Less aggressive congestion adjustments
	config.CongestionConfig.AdjustmentFactors = map[string]decimal.Decimal{
		"low":    decimal.NewFromFloat(0.8),
		"medium": decimal.NewFromFloat(0.95),
		"high":   decimal.NewFromFloat(1.1),
	}

	// Longer prediction horizons
	config.PredictionConfig.TimeHorizons = []time.Duration{
		5 * time.Minute,
		15 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
		2 * time.Hour,
	}

	// Lower safety margins for cost efficiency
	config.SafetyMargins.SafetyMultiplier = decimal.NewFromFloat(1.02)
	config.SafetyMargins.EmergencyMultiplier = decimal.NewFromFloat(1.1)

	return config
}

// GetBalancedConfig returns a balanced configuration
func GetBalancedConfig() GasOptimizerConfig {
	config := GetDefaultGasOptimizerConfig()

	// Balanced settings between speed and cost
	config.EIP1559Config.BaseFeeMultiplier = decimal.NewFromFloat(1.1)
	config.EIP1559Config.PriorityFeeStrategy = "dynamic"
	config.EIP1559Config.AggressivenessLevel = "moderate"
	config.EIP1559Config.TargetConfirmationTime = 5 * time.Minute

	// Moderate analysis windows
	config.HistoricalConfig.AnalysisWindow = 1 * time.Hour
	config.HistoricalConfig.SampleSize = 200

	// Balanced congestion adjustments
	config.CongestionConfig.AdjustmentFactors = map[string]decimal.Decimal{
		"low":    decimal.NewFromFloat(0.9),
		"medium": decimal.NewFromFloat(1.05),
		"high":   decimal.NewFromFloat(1.2),
	}

	return config
}

// ValidateGasOptimizerConfig validates gas optimizer configuration
func ValidateGasOptimizerConfig(config GasOptimizerConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if config.HistoryRetentionPeriod <= 0 {
		return fmt.Errorf("history retention period must be positive")
	}

	if config.MaxHistorySize <= 0 {
		return fmt.Errorf("max history size must be positive")
	}

	if len(config.OptimizationStrategies) == 0 {
		return fmt.Errorf("at least one optimization strategy must be specified")
	}

	// Validate EIP-1559 config
	if config.EIP1559Config.Enabled {
		if config.EIP1559Config.BaseFeeMultiplier.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("base fee multiplier must be positive")
		}
		if config.EIP1559Config.MaxFeeCapMultiplier.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("max fee cap multiplier must be positive")
		}
		if config.EIP1559Config.TargetConfirmationTime <= 0 {
			return fmt.Errorf("target confirmation time must be positive")
		}

		validStrategies := []string{"fixed", "dynamic", "aggressive"}
		isValid := false
		for _, strategy := range validStrategies {
			if config.EIP1559Config.PriorityFeeStrategy == strategy {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid priority fee strategy: %s", config.EIP1559Config.PriorityFeeStrategy)
		}

		validLevels := []string{"conservative", "moderate", "aggressive"}
		isValid = false
		for _, level := range validLevels {
			if config.EIP1559Config.AggressivenessLevel == level {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid aggressiveness level: %s", config.EIP1559Config.AggressivenessLevel)
		}
	}

	// Validate historical config
	if config.HistoricalConfig.Enabled {
		if config.HistoricalConfig.AnalysisWindow <= 0 {
			return fmt.Errorf("analysis window must be positive")
		}
		if config.HistoricalConfig.SampleSize <= 0 {
			return fmt.Errorf("sample size must be positive")
		}

		validWeightingStrategies := []string{"simple_average", "weighted_recent"}
		isValid := false
		for _, strategy := range validWeightingStrategies {
			if config.HistoricalConfig.WeightingStrategy == strategy {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid weighting strategy: %s", config.HistoricalConfig.WeightingStrategy)
		}
	}

	// Validate congestion config
	if config.CongestionConfig.Enabled {
		if config.CongestionConfig.MonitoringInterval <= 0 {
			return fmt.Errorf("monitoring interval must be positive")
		}

		for level, threshold := range config.CongestionConfig.CongestionThresholds {
			if threshold.LessThan(decimal.Zero) || threshold.GreaterThan(decimal.NewFromFloat(1)) {
				return fmt.Errorf("congestion threshold for %s must be between 0 and 1", level)
			}
		}

		for level, factor := range config.CongestionConfig.AdjustmentFactors {
			if factor.LessThanOrEqual(decimal.Zero) {
				return fmt.Errorf("adjustment factor for %s must be positive", level)
			}
		}
	}

	// Validate prediction config
	if config.PredictionConfig.Enabled {
		if len(config.PredictionConfig.PredictionMethods) == 0 {
			return fmt.Errorf("at least one prediction method must be specified")
		}
		if len(config.PredictionConfig.TimeHorizons) == 0 {
			return fmt.Errorf("at least one time horizon must be specified")
		}
		if config.PredictionConfig.ConfidenceThreshold.LessThan(decimal.Zero) ||
			config.PredictionConfig.ConfidenceThreshold.GreaterThan(decimal.NewFromFloat(1)) {
			return fmt.Errorf("confidence threshold must be between 0 and 1")
		}
	}

	// Validate safety margins
	if config.SafetyMargins.MinGasPrice.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("min gas price must be positive")
	}
	if config.SafetyMargins.MaxGasPrice.LessThanOrEqual(config.SafetyMargins.MinGasPrice) {
		return fmt.Errorf("max gas price must be greater than min gas price")
	}
	if config.SafetyMargins.SafetyMultiplier.LessThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("safety multiplier must be at least 1.0")
	}
	if config.SafetyMargins.EmergencyMultiplier.LessThan(config.SafetyMargins.SafetyMultiplier) {
		return fmt.Errorf("emergency multiplier must be at least as large as safety multiplier")
	}

	return nil
}

// GetSupportedOptimizationStrategies returns supported optimization strategies
func GetSupportedOptimizationStrategies() []string {
	return []string{
		"eip1559",
		"historical",
		"congestion_based",
		"prediction_based",
		"hybrid",
	}
}

// GetSupportedPriorityFeeStrategies returns supported priority fee strategies
func GetSupportedPriorityFeeStrategies() []string {
	return []string{
		"fixed",
		"dynamic",
		"aggressive",
	}
}

// GetSupportedAggressivenessLevels returns supported aggressiveness levels
func GetSupportedAggressivenessLevels() []string {
	return []string{
		"conservative",
		"moderate",
		"aggressive",
	}
}

// GetSupportedWeightingStrategies returns supported weighting strategies
func GetSupportedWeightingStrategies() []string {
	return []string{
		"simple_average",
		"weighted_recent",
		"exponential_decay",
	}
}

// GetSupportedPredictionMethods returns supported prediction methods
func GetSupportedPredictionMethods() []string {
	return []string{
		"moving_average",
		"exponential_smoothing",
		"linear_regression",
		"polynomial_regression",
		"neural_network",
		"ensemble",
	}
}

// GetOptimalConfigForUseCase returns optimal configuration for specific use cases
func GetOptimalConfigForUseCase(useCase string) (GasOptimizerConfig, error) {
	switch useCase {
	case "high_frequency_trading":
		return GetHighFrequencyConfig(), nil
	case "cost_optimized":
		return GetCostOptimizedConfig(), nil
	case "balanced":
		return GetBalancedConfig(), nil
	case "default":
		return GetDefaultGasOptimizerConfig(), nil
	default:
		return GasOptimizerConfig{}, fmt.Errorf("unsupported use case: %s", useCase)
	}
}

// GetRecommendedTimeHorizons returns recommended time horizons for different priorities
func GetRecommendedTimeHorizons() map[string][]time.Duration {
	return map[string][]time.Duration{
		"urgent": {
			10 * time.Second,
			30 * time.Second,
			1 * time.Minute,
		},
		"high": {
			30 * time.Second,
			1 * time.Minute,
			5 * time.Minute,
		},
		"medium": {
			1 * time.Minute,
			5 * time.Minute,
			15 * time.Minute,
		},
		"low": {
			5 * time.Minute,
			15 * time.Minute,
			30 * time.Minute,
			1 * time.Hour,
		},
	}
}

// GetStrategyDescription returns descriptions for optimization strategies
func GetStrategyDescription() map[string]string {
	return map[string]string{
		"eip1559":          "EIP-1559 based optimization using base fee and priority fee calculations",
		"historical":       "Historical data analysis for gas price trends and patterns",
		"congestion_based": "Network congestion monitoring and dynamic adjustment",
		"prediction_based": "Machine learning predictions for future gas prices",
		"hybrid":           "Combination of multiple strategies for optimal results",
	}
}
