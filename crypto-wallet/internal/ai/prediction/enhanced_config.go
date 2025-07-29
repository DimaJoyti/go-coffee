package prediction

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultEnhancedPredictorConfig returns default enhanced predictor configuration
func GetDefaultEnhancedPredictorConfig() EnhancedPredictorConfig {
	return EnhancedPredictorConfig{
		Enabled:        true,
		UpdateInterval: 5 * time.Minute,
		PredictionHorizons: []time.Duration{
			1 * time.Hour,
			4 * time.Hour,
			24 * time.Hour,
			7 * 24 * time.Hour,
			30 * 24 * time.Hour,
		},
		CacheRetentionPeriod: 1 * time.Hour,
		SentimentConfig: AdvancedSentimentConfig{
			Enabled:        true,
			Sources:        []string{"twitter", "reddit", "news", "telegram"},
			LanguageModels: []string{"bert", "roberta", "finbert"},
			SentimentWeights: map[string]decimal.Decimal{
				"twitter":  decimal.NewFromFloat(0.3),
				"reddit":   decimal.NewFromFloat(0.25),
				"news":     decimal.NewFromFloat(0.3),
				"telegram": decimal.NewFromFloat(0.15),
			},
			InfluencerWeights: map[string]decimal.Decimal{
				"high_influence":   decimal.NewFromFloat(0.5),
				"medium_influence": decimal.NewFromFloat(0.3),
				"low_influence":    decimal.NewFromFloat(0.2),
			},
			UpdateFrequency: 1 * time.Minute,
			HistoryWindow:   24 * time.Hour,
			EmotionAnalysis: true,
			TopicModeling:   true,
		},
		OnChainConfig: AdvancedOnChainConfig{
			Enabled: true,
			Metrics: []string{
				"transaction_volume", "active_addresses", "network_utilization",
				"whale_activity", "defi_tvl", "nft_volume", "staking_ratio",
			},
			Chains: []string{"ethereum", "bitcoin", "polygon", "arbitrum", "optimism"},
			MetricWeights: map[string]decimal.Decimal{
				"transaction_volume":  decimal.NewFromFloat(0.2),
				"active_addresses":    decimal.NewFromFloat(0.2),
				"network_utilization": decimal.NewFromFloat(0.15),
				"whale_activity":      decimal.NewFromFloat(0.15),
				"defi_tvl":            decimal.NewFromFloat(0.15),
				"nft_volume":          decimal.NewFromFloat(0.1),
				"staking_ratio":       decimal.NewFromFloat(0.05),
			},
			UpdateFrequency: 5 * time.Minute,
			HistoryWindow:   7 * 24 * time.Hour,
			WhaleTracking:   true,
			DeFiIntegration: true,
			NFTAnalysis:     true,
		},
		TechnicalConfig: AdvancedTechnicalConfig{
			Enabled: true,
			Indicators: []string{
				"rsi", "macd", "bollinger_bands", "moving_averages",
				"volume_profile", "fibonacci", "support_resistance",
			},
			Timeframes: []string{"1m", "5m", "15m", "1h", "4h", "1d"},
			IndicatorWeights: map[string]decimal.Decimal{
				"rsi":                decimal.NewFromFloat(0.15),
				"macd":               decimal.NewFromFloat(0.15),
				"bollinger_bands":    decimal.NewFromFloat(0.15),
				"moving_averages":    decimal.NewFromFloat(0.2),
				"volume_profile":     decimal.NewFromFloat(0.15),
				"fibonacci":          decimal.NewFromFloat(0.1),
				"support_resistance": decimal.NewFromFloat(0.1),
			},
			UpdateFrequency:    1 * time.Minute,
			PatternRecognition: true,
			VolumeAnalysis:     true,
			SupportResistance:  true,
		},
		MacroConfig: AdvancedMacroConfig{
			Enabled: true,
			Factors: []string{
				"interest_rates", "inflation", "usd_index", "stock_market",
				"geopolitical_events", "monetary_policy", "economic_indicators",
			},
			FactorWeights: map[string]decimal.Decimal{
				"interest_rates":      decimal.NewFromFloat(0.2),
				"inflation":           decimal.NewFromFloat(0.15),
				"usd_index":           decimal.NewFromFloat(0.15),
				"stock_market":        decimal.NewFromFloat(0.2),
				"geopolitical_events": decimal.NewFromFloat(0.1),
				"monetary_policy":     decimal.NewFromFloat(0.1),
				"economic_indicators": decimal.NewFromFloat(0.1),
			},
			UpdateFrequency:    1 * time.Hour,
			EconomicIndicators: true,
			GeopoliticalEvents: true,
			MonetaryPolicy:     true,
		},
		MLConfig: MachineLearningConfig{
			Enabled: true,
			Models: []string{
				"lstm", "transformer", "random_forest", "xgboost",
				"neural_network", "svm", "linear_regression",
			},
			ModelWeights: map[string]decimal.Decimal{
				"lstm":              decimal.NewFromFloat(0.25),
				"transformer":       decimal.NewFromFloat(0.2),
				"random_forest":     decimal.NewFromFloat(0.15),
				"xgboost":           decimal.NewFromFloat(0.15),
				"neural_network":    decimal.NewFromFloat(0.1),
				"svm":               decimal.NewFromFloat(0.1),
				"linear_regression": decimal.NewFromFloat(0.05),
			},
			TrainingFrequency:    24 * time.Hour,
			ValidationSplit:      decimal.NewFromFloat(0.2),
			FeatureSelection:     true,
			HyperparameterTuning: true,
			EnsembleMethods:      []string{"voting", "stacking", "bagging"},
		},
		EnsembleConfig: AdvancedEnsembleConfig{
			Enabled:           true,
			CombinationMethod: "weighted_average",
			WeightingStrategy: "performance_based",
			PerformanceWeights: map[string]decimal.Decimal{
				"sentiment": decimal.NewFromFloat(0.2),
				"onchain":   decimal.NewFromFloat(0.25),
				"technical": decimal.NewFromFloat(0.2),
				"macro":     decimal.NewFromFloat(0.15),
				"ml":        decimal.NewFromFloat(0.2),
			},
			ConfidenceThreshold: decimal.NewFromFloat(0.7),
			DiversityBonus:      decimal.NewFromFloat(0.1),
			AdaptiveLearning:    true,
		},
		DataAggregatorConfig: DataAggregatorConfig{
			Enabled: true,
			Sources: []string{
				"coinbase", "binance", "kraken", "coingecko", "coinmarketcap",
				"messari", "glassnode", "santiment", "lunarcrush",
			},
			UpdateFrequency:   1 * time.Minute,
			DataRetention:     7 * 24 * time.Hour,
			QualityFiltering:  true,
			RealTimeStreaming: true,
		},
		FeatureExtractorConfig: FeatureExtractorConfig{
			Enabled: true,
			FeatureTypes: []string{
				"price_features", "volume_features", "sentiment_features",
				"onchain_features", "technical_features", "macro_features",
			},
			ExtractionMethods: []string{
				"statistical", "fourier_transform", "wavelet_transform",
				"pca", "autoencoder", "feature_engineering",
			},
			FeatureSelection:        true,
			DimensionalityReduction: true,
			FeatureEngineering:      true,
		},
		ModelManagerConfig: ModelManagerConfig{
			Enabled: true,
			ModelTypes: []string{
				"regression", "classification", "time_series",
				"deep_learning", "ensemble", "reinforcement_learning",
			},
			AutoRetraining:      true,
			PerformanceTracking: true,
			ModelVersioning:     true,
			A_BTesting:          true,
		},
	}
}

// GetHighFrequencyConfig returns configuration optimized for high-frequency predictions
func GetHighFrequencyConfig() EnhancedPredictorConfig {
	config := GetDefaultEnhancedPredictorConfig()

	// More frequent updates
	config.UpdateInterval = 30 * time.Second
	config.CacheRetentionPeriod = 5 * time.Minute

	// Shorter prediction horizons
	config.PredictionHorizons = []time.Duration{
		5 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
		4 * time.Hour,
	}

	// More frequent sentiment updates
	config.SentimentConfig.UpdateFrequency = 10 * time.Second
	config.SentimentConfig.HistoryWindow = 2 * time.Hour

	// More frequent on-chain updates
	config.OnChainConfig.UpdateFrequency = 1 * time.Minute
	config.OnChainConfig.HistoryWindow = 24 * time.Hour

	// More frequent technical analysis
	config.TechnicalConfig.UpdateFrequency = 10 * time.Second
	config.TechnicalConfig.Timeframes = []string{"1m", "5m", "15m", "1h"}

	// Less frequent macro updates (not as relevant for HF)
	config.MacroConfig.UpdateFrequency = 6 * time.Hour

	// More frequent ML retraining
	config.MLConfig.TrainingFrequency = 6 * time.Hour

	// More frequent data updates
	config.DataAggregatorConfig.UpdateFrequency = 10 * time.Second
	config.DataAggregatorConfig.DataRetention = 24 * time.Hour

	return config
}

// GetLongTermConfig returns configuration optimized for long-term predictions
func GetLongTermConfig() EnhancedPredictorConfig {
	config := GetDefaultEnhancedPredictorConfig()

	// Less frequent updates
	config.UpdateInterval = 1 * time.Hour
	config.CacheRetentionPeriod = 6 * time.Hour

	// Longer prediction horizons
	config.PredictionHorizons = []time.Duration{
		24 * time.Hour,
		7 * 24 * time.Hour,
		30 * 24 * time.Hour,
		90 * 24 * time.Hour,
		365 * 24 * time.Hour,
	}

	// Less frequent sentiment updates
	config.SentimentConfig.UpdateFrequency = 15 * time.Minute
	config.SentimentConfig.HistoryWindow = 30 * 24 * time.Hour

	// Less frequent on-chain updates
	config.OnChainConfig.UpdateFrequency = 1 * time.Hour
	config.OnChainConfig.HistoryWindow = 90 * 24 * time.Hour

	// Less frequent technical analysis with longer timeframes
	config.TechnicalConfig.UpdateFrequency = 15 * time.Minute
	config.TechnicalConfig.Timeframes = []string{"4h", "1d", "1w", "1M"}

	// More emphasis on macro factors
	config.MacroConfig.UpdateFrequency = 6 * time.Hour
	config.EnsembleConfig.PerformanceWeights["macro"] = decimal.NewFromFloat(0.25)
	config.EnsembleConfig.PerformanceWeights["technical"] = decimal.NewFromFloat(0.15)

	// Less frequent ML retraining
	config.MLConfig.TrainingFrequency = 7 * 24 * time.Hour

	// Less frequent data updates
	config.DataAggregatorConfig.UpdateFrequency = 15 * time.Minute
	config.DataAggregatorConfig.DataRetention = 90 * 24 * time.Hour

	return config
}

// GetResearchConfig returns configuration optimized for research and analysis
func GetResearchConfig() EnhancedPredictorConfig {
	config := GetDefaultEnhancedPredictorConfig()

	// Comprehensive prediction horizons
	config.PredictionHorizons = []time.Duration{
		15 * time.Minute,
		1 * time.Hour,
		4 * time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
		30 * 24 * time.Hour,
		90 * 24 * time.Hour,
	}

	// Extended data retention
	config.CacheRetentionPeriod = 24 * time.Hour
	config.SentimentConfig.HistoryWindow = 90 * 24 * time.Hour
	config.OnChainConfig.HistoryWindow = 365 * 24 * time.Hour
	config.DataAggregatorConfig.DataRetention = 365 * 24 * time.Hour

	// All timeframes for technical analysis
	config.TechnicalConfig.Timeframes = []string{"1m", "5m", "15m", "1h", "4h", "1d", "1w", "1M"}

	// All available models
	config.MLConfig.Models = append(config.MLConfig.Models,
		"gru", "attention", "cnn", "autoencoder", "gan")

	// Enhanced feature extraction
	config.FeatureExtractorConfig.ExtractionMethods = append(
		config.FeatureExtractorConfig.ExtractionMethods,
		"ica", "tsne", "umap", "kernel_pca")

	return config
}

// ValidateEnhancedPredictorConfig validates enhanced predictor configuration
func ValidateEnhancedPredictorConfig(config EnhancedPredictorConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if len(config.PredictionHorizons) == 0 {
		return fmt.Errorf("at least one prediction horizon must be specified")
	}

	if config.CacheRetentionPeriod <= 0 {
		return fmt.Errorf("cache retention period must be positive")
	}

	// Validate sentiment config
	if config.SentimentConfig.Enabled {
		if len(config.SentimentConfig.Sources) == 0 {
			return fmt.Errorf("at least one sentiment source must be specified")
		}
		if config.SentimentConfig.UpdateFrequency <= 0 {
			return fmt.Errorf("sentiment update frequency must be positive")
		}
		if config.SentimentConfig.HistoryWindow <= 0 {
			return fmt.Errorf("sentiment history window must be positive")
		}
	}

	// Validate on-chain config
	if config.OnChainConfig.Enabled {
		if len(config.OnChainConfig.Metrics) == 0 {
			return fmt.Errorf("at least one on-chain metric must be specified")
		}
		if len(config.OnChainConfig.Chains) == 0 {
			return fmt.Errorf("at least one blockchain must be specified")
		}
		if config.OnChainConfig.UpdateFrequency <= 0 {
			return fmt.Errorf("on-chain update frequency must be positive")
		}
	}

	// Validate technical config
	if config.TechnicalConfig.Enabled {
		if len(config.TechnicalConfig.Indicators) == 0 {
			return fmt.Errorf("at least one technical indicator must be specified")
		}
		if len(config.TechnicalConfig.Timeframes) == 0 {
			return fmt.Errorf("at least one timeframe must be specified")
		}
		if config.TechnicalConfig.UpdateFrequency <= 0 {
			return fmt.Errorf("technical update frequency must be positive")
		}
	}

	// Validate ML config
	if config.MLConfig.Enabled {
		if len(config.MLConfig.Models) == 0 {
			return fmt.Errorf("at least one ML model must be specified")
		}
		if config.MLConfig.TrainingFrequency <= 0 {
			return fmt.Errorf("training frequency must be positive")
		}
		if config.MLConfig.ValidationSplit.LessThan(decimal.Zero) ||
			config.MLConfig.ValidationSplit.GreaterThan(decimal.NewFromFloat(1)) {
			return fmt.Errorf("validation split must be between 0 and 1")
		}
	}

	// Validate ensemble config
	if config.EnsembleConfig.Enabled {
		if config.EnsembleConfig.ConfidenceThreshold.LessThan(decimal.Zero) ||
			config.EnsembleConfig.ConfidenceThreshold.GreaterThan(decimal.NewFromFloat(1)) {
			return fmt.Errorf("confidence threshold must be between 0 and 1")
		}
	}

	return nil
}

// GetOptimalConfigForUseCase returns optimal configuration for specific use cases
func GetOptimalConfigForUseCase(useCase string) (EnhancedPredictorConfig, error) {
	switch useCase {
	case "high_frequency":
		return GetHighFrequencyConfig(), nil
	case "long_term":
		return GetLongTermConfig(), nil
	case "research":
		return GetResearchConfig(), nil
	case "default":
		return GetDefaultEnhancedPredictorConfig(), nil
	default:
		return EnhancedPredictorConfig{}, fmt.Errorf("unsupported use case: %s", useCase)
	}
}

// GetSupportedSentimentSources returns supported sentiment analysis sources
func GetSupportedSentimentSources() []string {
	return []string{
		"twitter", "reddit", "news", "telegram", "discord",
		"youtube", "medium", "substack", "github", "stackoverflow",
	}
}

// GetSupportedOnChainMetrics returns supported on-chain metrics
func GetSupportedOnChainMetrics() []string {
	return []string{
		"transaction_volume", "active_addresses", "network_utilization",
		"whale_activity", "defi_tvl", "nft_volume", "staking_ratio",
		"hash_rate", "difficulty", "fees", "miner_revenue",
	}
}

// Note: GetSupportedTechnicalIndicators function is defined in prediction_config.go

// GetSupportedMLModels returns supported machine learning models
func GetSupportedMLModels() []string {
	return []string{
		"lstm", "gru", "transformer", "attention", "cnn",
		"random_forest", "xgboost", "neural_network", "svm",
		"linear_regression", "autoencoder", "gan",
	}
}
