package mempool

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultMempoolAnalyzerConfig returns default mempool analyzer configuration
func GetDefaultMempoolAnalyzerConfig() MempoolAnalyzerConfig {
	return MempoolAnalyzerConfig{
		Enabled:             true,
		UpdateInterval:      30 * time.Second,
		DataRetentionPeriod: 1 * time.Hour,
		MaxTransactions:     10000,
		GasTrackerConfig: GasTrackerConfig{
			Enabled:         true,
			TrackingWindow:  30 * time.Minute,
			SampleSize:      1000,
			PercentileTargets: []int{10, 25, 50, 75, 90, 95, 99},
			UpdateFrequency: 1 * time.Minute,
		},
		CongestionModelConfig: CongestionModelConfig{
			Enabled:        true,
			AnalysisWindow: 15 * time.Minute,
			CongestionThresholds: map[string]decimal.Decimal{
				"low":    decimal.NewFromFloat(0.3),
				"medium": decimal.NewFromFloat(0.6),
				"high":   decimal.NewFromFloat(0.8),
			},
			PredictionHorizon: 30 * time.Minute,
		},
		GasPredictorConfig: GasPredictorConfig{
			Enabled: true,
			PredictionMethods: []string{
				"moving_average",
				"exponential_smoothing",
				"linear_regression",
			},
			ConfidenceLevels: []decimal.Decimal{
				decimal.NewFromFloat(0.8),
				decimal.NewFromFloat(0.9),
				decimal.NewFromFloat(0.95),
			},
			TimeHorizons: []time.Duration{
				5 * time.Minute,
				15 * time.Minute,
				30 * time.Minute,
				1 * time.Hour,
			},
		},
		TimeEstimatorConfig: TimeEstimatorConfig{
			Enabled:          true,
			EstimationMethod: "historical_analysis",
			HistoryWindow:    24 * time.Hour,
			AccuracyTarget:   decimal.NewFromFloat(0.8),
		},
		PriorityAnalyzerConfig: PriorityAnalyzerConfig{
			Enabled: true,
			PriorityFactors: []string{
				"gas_price",
				"gas_tip",
				"transaction_size",
				"age",
				"replacement",
			},
			WeightingMethod: "dynamic",
		},
	}
}

// GetHighFrequencyConfig returns high-frequency trading configuration
func GetHighFrequencyConfig() MempoolAnalyzerConfig {
	config := GetDefaultMempoolAnalyzerConfig()
	
	// More frequent updates for high-frequency trading
	config.UpdateInterval = 5 * time.Second
	config.DataRetentionPeriod = 30 * time.Minute
	config.MaxTransactions = 50000
	
	// More aggressive gas tracking
	config.GasTrackerConfig.TrackingWindow = 5 * time.Minute
	config.GasTrackerConfig.UpdateFrequency = 10 * time.Second
	config.GasTrackerConfig.SampleSize = 5000
	
	// Shorter analysis windows
	config.CongestionModelConfig.AnalysisWindow = 2 * time.Minute
	config.CongestionModelConfig.PredictionHorizon = 5 * time.Minute
	
	// More prediction methods
	config.GasPredictorConfig.PredictionMethods = append(
		config.GasPredictorConfig.PredictionMethods,
		"neural_network",
		"ensemble",
	)
	
	// Shorter time horizons
	config.GasPredictorConfig.TimeHorizons = []time.Duration{
		30 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
	}
	
	// Faster time estimation
	config.TimeEstimatorConfig.HistoryWindow = 2 * time.Hour
	
	return config
}

// GetLowLatencyConfig returns low-latency configuration
func GetLowLatencyConfig() MempoolAnalyzerConfig {
	config := GetDefaultMempoolAnalyzerConfig()
	
	// Optimized for low latency
	config.UpdateInterval = 1 * time.Second
	config.DataRetentionPeriod = 10 * time.Minute
	config.MaxTransactions = 5000
	
	// Minimal gas tracking
	config.GasTrackerConfig.TrackingWindow = 1 * time.Minute
	config.GasTrackerConfig.UpdateFrequency = 1 * time.Second
	config.GasTrackerConfig.SampleSize = 100
	config.GasTrackerConfig.PercentileTargets = []int{50, 75, 90, 95}
	
	// Fast congestion analysis
	config.CongestionModelConfig.AnalysisWindow = 30 * time.Second
	config.CongestionModelConfig.PredictionHorizon = 1 * time.Minute
	
	// Simple prediction methods
	config.GasPredictorConfig.PredictionMethods = []string{"moving_average"}
	config.GasPredictorConfig.TimeHorizons = []time.Duration{
		10 * time.Second,
		30 * time.Second,
		1 * time.Minute,
	}
	
	// Fast time estimation
	config.TimeEstimatorConfig.HistoryWindow = 30 * time.Minute
	
	return config
}

// GetAnalyticsConfig returns analytics-focused configuration
func GetAnalyticsConfig() MempoolAnalyzerConfig {
	config := GetDefaultMempoolAnalyzerConfig()
	
	// Extended data retention for analytics
	config.DataRetentionPeriod = 24 * time.Hour
	config.MaxTransactions = 100000
	
	// Comprehensive gas tracking
	config.GasTrackerConfig.TrackingWindow = 4 * time.Hour
	config.GasTrackerConfig.SampleSize = 10000
	config.GasTrackerConfig.PercentileTargets = []int{
		1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50,
		55, 60, 65, 70, 75, 80, 85, 90, 95, 99,
	}
	
	// Extended analysis windows
	config.CongestionModelConfig.AnalysisWindow = 2 * time.Hour
	config.CongestionModelConfig.PredictionHorizon = 4 * time.Hour
	
	// All prediction methods
	config.GasPredictorConfig.PredictionMethods = []string{
		"moving_average",
		"exponential_smoothing",
		"linear_regression",
		"polynomial_regression",
		"neural_network",
		"ensemble",
		"arima",
	}
	
	// Extended time horizons
	config.GasPredictorConfig.TimeHorizons = []time.Duration{
		5 * time.Minute,
		15 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
		2 * time.Hour,
		4 * time.Hour,
		8 * time.Hour,
		24 * time.Hour,
	}
	
	// Extended history for time estimation
	config.TimeEstimatorConfig.HistoryWindow = 7 * 24 * time.Hour // 1 week
	
	return config
}

// ValidateMempoolAnalyzerConfig validates mempool analyzer configuration
func ValidateMempoolAnalyzerConfig(config MempoolAnalyzerConfig) error {
	if !config.Enabled {
		return nil
	}
	
	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}
	
	if config.DataRetentionPeriod <= 0 {
		return fmt.Errorf("data retention period must be positive")
	}
	
	if config.MaxTransactions <= 0 {
		return fmt.Errorf("max transactions must be positive")
	}
	
	// Validate gas tracker config
	if config.GasTrackerConfig.Enabled {
		if config.GasTrackerConfig.TrackingWindow <= 0 {
			return fmt.Errorf("gas tracker tracking window must be positive")
		}
		if config.GasTrackerConfig.SampleSize <= 0 {
			return fmt.Errorf("gas tracker sample size must be positive")
		}
		if config.GasTrackerConfig.UpdateFrequency <= 0 {
			return fmt.Errorf("gas tracker update frequency must be positive")
		}
		if len(config.GasTrackerConfig.PercentileTargets) == 0 {
			return fmt.Errorf("gas tracker percentile targets cannot be empty")
		}
		for _, p := range config.GasTrackerConfig.PercentileTargets {
			if p < 0 || p > 100 {
				return fmt.Errorf("percentile target must be between 0 and 100: %d", p)
			}
		}
	}
	
	// Validate congestion model config
	if config.CongestionModelConfig.Enabled {
		if config.CongestionModelConfig.AnalysisWindow <= 0 {
			return fmt.Errorf("congestion model analysis window must be positive")
		}
		if config.CongestionModelConfig.PredictionHorizon <= 0 {
			return fmt.Errorf("congestion model prediction horizon must be positive")
		}
		for level, threshold := range config.CongestionModelConfig.CongestionThresholds {
			if threshold.LessThan(decimal.Zero) || threshold.GreaterThan(decimal.NewFromFloat(1)) {
				return fmt.Errorf("congestion threshold for %s must be between 0 and 1", level)
			}
		}
	}
	
	// Validate gas predictor config
	if config.GasPredictorConfig.Enabled {
		if len(config.GasPredictorConfig.PredictionMethods) == 0 {
			return fmt.Errorf("gas predictor methods cannot be empty")
		}
		if len(config.GasPredictorConfig.ConfidenceLevels) == 0 {
			return fmt.Errorf("gas predictor confidence levels cannot be empty")
		}
		for _, confidence := range config.GasPredictorConfig.ConfidenceLevels {
			if confidence.LessThan(decimal.Zero) || confidence.GreaterThan(decimal.NewFromFloat(1)) {
				return fmt.Errorf("confidence level must be between 0 and 1")
			}
		}
		if len(config.GasPredictorConfig.TimeHorizons) == 0 {
			return fmt.Errorf("gas predictor time horizons cannot be empty")
		}
	}
	
	// Validate time estimator config
	if config.TimeEstimatorConfig.Enabled {
		if config.TimeEstimatorConfig.HistoryWindow <= 0 {
			return fmt.Errorf("time estimator history window must be positive")
		}
		if config.TimeEstimatorConfig.AccuracyTarget.LessThan(decimal.Zero) || 
		   config.TimeEstimatorConfig.AccuracyTarget.GreaterThan(decimal.NewFromFloat(1)) {
			return fmt.Errorf("time estimator accuracy target must be between 0 and 1")
		}
	}
	
	// Validate priority analyzer config
	if config.PriorityAnalyzerConfig.Enabled {
		if len(config.PriorityAnalyzerConfig.PriorityFactors) == 0 {
			return fmt.Errorf("priority analyzer factors cannot be empty")
		}
	}
	
	return nil
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
		"arima",
		"lstm",
		"random_forest",
	}
}

// GetSupportedEstimationMethods returns supported time estimation methods
func GetSupportedEstimationMethods() []string {
	return []string{
		"historical_analysis",
		"regression_analysis",
		"machine_learning",
		"statistical_model",
		"hybrid_approach",
	}
}

// GetSupportedWeightingMethods returns supported priority weighting methods
func GetSupportedWeightingMethods() []string {
	return []string{
		"static",
		"dynamic",
		"adaptive",
		"machine_learning",
	}
}

// GetSupportedPriorityFactors returns supported priority factors
func GetSupportedPriorityFactors() []string {
	return []string{
		"gas_price",
		"gas_tip",
		"transaction_size",
		"age",
		"replacement",
		"sender_reputation",
		"contract_interaction",
		"value_transfer",
	}
}

// GetCongestionLevelDescription returns congestion level descriptions
func GetCongestionLevelDescription() map[string]string {
	return map[string]string{
		"low":    "Low network congestion - fast confirmations and low fees",
		"medium": "Moderate network congestion - average confirmations and fees",
		"high":   "High network congestion - slow confirmations and high fees",
	}
}

// GetPredictionMethodDescription returns prediction method descriptions
func GetPredictionMethodDescription() map[string]string {
	return map[string]string{
		"moving_average":        "Simple moving average of recent gas prices",
		"exponential_smoothing": "Exponentially weighted moving average",
		"linear_regression":     "Linear regression on historical data",
		"polynomial_regression": "Polynomial regression for non-linear trends",
		"neural_network":        "Neural network-based prediction",
		"ensemble":              "Combination of multiple prediction methods",
		"arima":                 "AutoRegressive Integrated Moving Average",
		"lstm":                  "Long Short-Term Memory neural network",
		"random_forest":         "Random forest ensemble method",
	}
}

// GetOptimalConfigForUseCase returns optimal configuration for specific use cases
func GetOptimalConfigForUseCase(useCase string) (MempoolAnalyzerConfig, error) {
	switch useCase {
	case "high_frequency_trading":
		return GetHighFrequencyConfig(), nil
	case "low_latency":
		return GetLowLatencyConfig(), nil
	case "analytics":
		return GetAnalyticsConfig(), nil
	case "default":
		return GetDefaultMempoolAnalyzerConfig(), nil
	default:
		return MempoolAnalyzerConfig{}, fmt.Errorf("unsupported use case: %s", useCase)
	}
}

// GetRecommendedPercentiles returns recommended percentiles for different use cases
func GetRecommendedPercentiles() map[string][]int {
	return map[string][]int{
		"basic":      {25, 50, 75, 90, 95},
		"detailed":   {10, 25, 50, 75, 90, 95, 99},
		"analytics":  {1, 5, 10, 25, 50, 75, 90, 95, 99},
		"trading":    {50, 75, 90, 95, 99},
		"monitoring": {25, 50, 75, 95},
	}
}

// GetRecommendedTimeHorizons returns recommended time horizons for different use cases
func GetRecommendedTimeHorizons() map[string][]time.Duration {
	return map[string][]time.Duration{
		"scalping": {
			10 * time.Second,
			30 * time.Second,
			1 * time.Minute,
		},
		"day_trading": {
			1 * time.Minute,
			5 * time.Minute,
			15 * time.Minute,
			30 * time.Minute,
		},
		"swing_trading": {
			15 * time.Minute,
			30 * time.Minute,
			1 * time.Hour,
			4 * time.Hour,
		},
		"long_term": {
			1 * time.Hour,
			4 * time.Hour,
			12 * time.Hour,
			24 * time.Hour,
		},
		"analytics": {
			5 * time.Minute,
			15 * time.Minute,
			30 * time.Minute,
			1 * time.Hour,
			4 * time.Hour,
			12 * time.Hour,
			24 * time.Hour,
		},
	}
}
