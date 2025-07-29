package prediction

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultMarketPredictorConfig returns default market predictor configuration
func GetDefaultMarketPredictorConfig() MarketPredictorConfig {
	return MarketPredictorConfig{
		Enabled:        true,
		UpdateInterval: 15 * time.Minute,
		CacheTimeout:   1 * time.Hour,
		PredictionHorizons: []time.Duration{
			1 * time.Hour,
			4 * time.Hour,
			24 * time.Hour,
			7 * 24 * time.Hour,
		},
		ConfidenceThreshold: decimal.NewFromFloat(0.7),
		SentimentConfig: SentimentAnalysisConfig{
			Enabled:        true,
			Sources:        []string{"twitter", "reddit", "news", "telegram"},
			UpdateInterval: 30 * time.Minute,
			SentimentWeight:  decimal.NewFromFloat(0.3),
			NewsWeight:       decimal.NewFromFloat(0.25),
			SocialWeight:     decimal.NewFromFloat(0.25),
			InfluencerWeight: decimal.NewFromFloat(0.2),
			LanguageModels:   []string{"bert", "roberta", "finbert"},
		},
		OnChainConfig: OnChainAnalysisConfig{
			Enabled: true,
			Metrics: []string{
				"transaction_volume",
				"active_addresses",
				"network_hash_rate",
				"exchange_flows",
				"whale_movements",
				"defi_tvl",
				"staking_ratio",
			},
			UpdateInterval:      10 * time.Minute,
			TransactionWeight:   decimal.NewFromFloat(0.25),
			AddressWeight:       decimal.NewFromFloat(0.2),
			VolumeWeight:        decimal.NewFromFloat(0.2),
			LiquidityWeight:     decimal.NewFromFloat(0.15),
			DeFiWeight:          decimal.NewFromFloat(0.1),
			NetworkHealthWeight: decimal.NewFromFloat(0.1),
		},
		TechnicalConfig: TechnicalAnalysisConfig{
			Enabled: true,
			Indicators: []string{
				"sma", "ema", "rsi", "macd", "bollinger_bands",
				"stochastic", "williams_r", "cci", "atr", "obv",
			},
			Timeframes:     []string{"1m", "5m", "15m", "1h", "4h", "1d"},
			UpdateInterval: 5 * time.Minute,
			TrendWeight:     decimal.NewFromFloat(0.3),
			MomentumWeight:  decimal.NewFromFloat(0.25),
			VolatilityWeight: decimal.NewFromFloat(0.2),
			VolumeWeight:    decimal.NewFromFloat(0.15),
			SupportResistanceWeight: decimal.NewFromFloat(0.1),
		},
		MacroConfig: MacroAnalysisConfig{
			Enabled: true,
			Indicators: []string{
				"fed_rates", "inflation", "gdp", "unemployment",
				"dollar_index", "gold_price", "oil_price", "vix",
			},
			UpdateInterval:     1 * time.Hour,
			EconomicWeight:     decimal.NewFromFloat(0.4),
			MonetaryWeight:     decimal.NewFromFloat(0.3),
			GeopoliticalWeight: decimal.NewFromFloat(0.2),
			RegulatoryWeight:   decimal.NewFromFloat(0.1),
		},
		EnsembleConfig: EnsembleModelConfig{
			Enabled:           true,
			Models:            []string{"sentiment", "onchain", "technical", "macro"},
			WeightingStrategy: "performance_weighted",
			ModelWeights: map[string]decimal.Decimal{
				"sentiment": decimal.NewFromFloat(0.25),
				"onchain":   decimal.NewFromFloat(0.30),
				"technical": decimal.NewFromFloat(0.30),
				"macro":     decimal.NewFromFloat(0.15),
			},
			RebalanceInterval: 24 * time.Hour,
			PerformanceWindow: 7 * 24 * time.Hour,
		},
		DataSources: []string{
			"coingecko",
			"coinmarketcap",
			"messari",
			"glassnode",
			"santiment",
			"lunarcrush",
		},
		ModelRetrainingConfig: ModelRetrainingConfig{
			Enabled:              true,
			RetrainingInterval:   7 * 24 * time.Hour, // Weekly
			PerformanceThreshold: decimal.NewFromFloat(0.6),
			DataWindow:           30 * 24 * time.Hour, // 30 days
			ValidationSplit:      decimal.NewFromFloat(0.2),
			AutoDeploy:           false, // Manual approval required
		},
		AlertThresholds: PredictionAlertThresholds{
			HighConfidenceBull:   decimal.NewFromFloat(0.8),
			HighConfidenceBear:   decimal.NewFromFloat(0.8),
			LowConfidence:        decimal.NewFromFloat(0.4),
			VolatilitySpike:      decimal.NewFromFloat(0.3),
			TrendReversal:        decimal.NewFromFloat(0.7),
			AnomalyDetection:     decimal.NewFromFloat(0.9),
		},
	}
}

// GetConservativePredictorConfig returns conservative predictor configuration
func GetConservativePredictorConfig() MarketPredictorConfig {
	config := GetDefaultMarketPredictorConfig()
	
	// Higher confidence thresholds
	config.ConfidenceThreshold = decimal.NewFromFloat(0.8)
	config.AlertThresholds.HighConfidenceBull = decimal.NewFromFloat(0.9)
	config.AlertThresholds.HighConfidenceBear = decimal.NewFromFloat(0.9)
	config.AlertThresholds.LowConfidence = decimal.NewFromFloat(0.6)
	
	// More weight on fundamental analysis
	config.EnsembleConfig.ModelWeights = map[string]decimal.Decimal{
		"sentiment": decimal.NewFromFloat(0.15),
		"onchain":   decimal.NewFromFloat(0.40),
		"technical": decimal.NewFromFloat(0.25),
		"macro":     decimal.NewFromFloat(0.20),
	}
	
	// Longer prediction horizons
	config.PredictionHorizons = []time.Duration{
		4 * time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
		30 * 24 * time.Hour,
	}
	
	return config
}

// GetAggressivePredictorConfig returns aggressive predictor configuration
func GetAggressivePredictorConfig() MarketPredictorConfig {
	config := GetDefaultMarketPredictorConfig()
	
	// Lower confidence thresholds
	config.ConfidenceThreshold = decimal.NewFromFloat(0.6)
	config.AlertThresholds.HighConfidenceBull = decimal.NewFromFloat(0.7)
	config.AlertThresholds.HighConfidenceBear = decimal.NewFromFloat(0.7)
	config.AlertThresholds.LowConfidence = decimal.NewFromFloat(0.3)
	
	// More weight on technical and sentiment
	config.EnsembleConfig.ModelWeights = map[string]decimal.Decimal{
		"sentiment": decimal.NewFromFloat(0.35),
		"onchain":   decimal.NewFromFloat(0.20),
		"technical": decimal.NewFromFloat(0.35),
		"macro":     decimal.NewFromFloat(0.10),
	}
	
	// Shorter prediction horizons
	config.PredictionHorizons = []time.Duration{
		15 * time.Minute,
		1 * time.Hour,
		4 * time.Hour,
		24 * time.Hour,
	}
	
	// More frequent updates
	config.UpdateInterval = 5 * time.Minute
	config.TechnicalConfig.UpdateInterval = 1 * time.Minute
	config.SentimentConfig.UpdateInterval = 10 * time.Minute
	
	return config
}

// GetDayTradingPredictorConfig returns day trading focused configuration
func GetDayTradingPredictorConfig() MarketPredictorConfig {
	config := GetDefaultMarketPredictorConfig()
	
	// Short-term prediction horizons
	config.PredictionHorizons = []time.Duration{
		5 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
		4 * time.Hour,
	}
	
	// High frequency updates
	config.UpdateInterval = 1 * time.Minute
	config.TechnicalConfig.UpdateInterval = 30 * time.Second
	config.OnChainConfig.UpdateInterval = 2 * time.Minute
	
	// Focus on technical analysis
	config.EnsembleConfig.ModelWeights = map[string]decimal.Decimal{
		"sentiment": decimal.NewFromFloat(0.10),
		"onchain":   decimal.NewFromFloat(0.15),
		"technical": decimal.NewFromFloat(0.65),
		"macro":     decimal.NewFromFloat(0.10),
	}
	
	// Short-term technical indicators
	config.TechnicalConfig.Timeframes = []string{"1m", "5m", "15m", "1h"}
	config.TechnicalConfig.Indicators = []string{
		"ema", "rsi", "macd", "stochastic", "williams_r",
		"cci", "atr", "obv", "vwap", "pivot_points",
	}
	
	return config
}

// ValidateMarketPredictorConfig validates market predictor configuration
func ValidateMarketPredictorConfig(config MarketPredictorConfig) error {
	if !config.Enabled {
		return nil
	}
	
	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}
	
	if config.CacheTimeout <= 0 {
		return fmt.Errorf("cache timeout must be positive")
	}
	
	if len(config.PredictionHorizons) == 0 {
		return fmt.Errorf("at least one prediction horizon must be specified")
	}
	
	if config.ConfidenceThreshold.LessThan(decimal.Zero) || 
	   config.ConfidenceThreshold.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("confidence threshold must be between 0 and 1")
	}
	
	// Validate sentiment config
	if config.SentimentConfig.Enabled {
		if len(config.SentimentConfig.Sources) == 0 {
			return fmt.Errorf("sentiment sources cannot be empty when enabled")
		}
		if config.SentimentConfig.UpdateInterval <= 0 {
			return fmt.Errorf("sentiment update interval must be positive")
		}
	}
	
	// Validate on-chain config
	if config.OnChainConfig.Enabled {
		if len(config.OnChainConfig.Metrics) == 0 {
			return fmt.Errorf("on-chain metrics cannot be empty when enabled")
		}
		if config.OnChainConfig.UpdateInterval <= 0 {
			return fmt.Errorf("on-chain update interval must be positive")
		}
	}
	
	// Validate technical config
	if config.TechnicalConfig.Enabled {
		if len(config.TechnicalConfig.Indicators) == 0 {
			return fmt.Errorf("technical indicators cannot be empty when enabled")
		}
		if len(config.TechnicalConfig.Timeframes) == 0 {
			return fmt.Errorf("technical timeframes cannot be empty when enabled")
		}
	}
	
	// Validate ensemble config
	if config.EnsembleConfig.Enabled {
		if len(config.EnsembleConfig.Models) == 0 {
			return fmt.Errorf("ensemble models cannot be empty when enabled")
		}
		
		// Validate model weights sum to 1.0
		totalWeight := decimal.Zero
		for _, weight := range config.EnsembleConfig.ModelWeights {
			totalWeight = totalWeight.Add(weight)
		}
		if !totalWeight.Equal(decimal.NewFromFloat(1.0)) {
			return fmt.Errorf("ensemble model weights must sum to 1.0, got %s", totalWeight.String())
		}
	}
	
	// Validate alert thresholds
	thresholds := config.AlertThresholds
	if thresholds.HighConfidenceBull.LessThan(decimal.Zero) || 
	   thresholds.HighConfidenceBull.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("high confidence bull threshold must be between 0 and 1")
	}
	
	if thresholds.HighConfidenceBear.LessThan(decimal.Zero) || 
	   thresholds.HighConfidenceBear.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("high confidence bear threshold must be between 0 and 1")
	}
	
	return nil
}

// GetPredictionHorizonDescription returns description for prediction horizons
func GetPredictionHorizonDescription() map[string]string {
	return map[string]string{
		"5m":  "Ultra short-term - Scalping and high-frequency trading",
		"15m": "Very short-term - Intraday momentum trading",
		"1h":  "Short-term - Day trading and swing entry/exit",
		"4h":  "Medium-term - Swing trading and position sizing",
		"1d":  "Daily - Position trading and trend following",
		"1w":  "Weekly - Strategic positioning and portfolio allocation",
		"1M":  "Monthly - Long-term investment decisions",
	}
}

// GetAnalysisTypeDescription returns description for analysis types
func GetAnalysisTypeDescription() map[string]string {
	return map[string]string{
		"sentiment":  "Market sentiment from news, social media, and expert opinions",
		"onchain":    "Blockchain metrics including transactions, addresses, and network health",
		"technical":  "Price action analysis using technical indicators and chart patterns",
		"macro":      "Macroeconomic factors including monetary policy and global events",
		"ensemble":   "Combined prediction from multiple analysis methods",
	}
}

// GetConfidenceLevelDescription returns description for confidence levels
func GetConfidenceLevelDescription() map[string]string {
	return map[string]string{
		"very_high": "90-100% - Extremely confident prediction with strong signal consensus",
		"high":      "80-90% - High confidence with multiple confirming signals",
		"medium":    "60-80% - Moderate confidence with some conflicting signals",
		"low":       "40-60% - Low confidence with weak or mixed signals",
		"very_low":  "0-40% - Very low confidence, high uncertainty",
	}
}

// GetRiskLevelDescription returns description for risk levels
func GetRiskLevelDescription() map[string]string {
	return map[string]string{
		"low":    "Low risk - Stable prediction with low volatility expected",
		"medium": "Medium risk - Moderate uncertainty with average volatility",
		"high":   "High risk - Significant uncertainty with high volatility expected",
	}
}

// GetSupportedDataSources returns supported data sources
func GetSupportedDataSources() []string {
	return []string{
		"coingecko",
		"coinmarketcap",
		"messari",
		"glassnode",
		"santiment",
		"lunarcrush",
		"cryptoquant",
		"dune_analytics",
		"nansen",
		"chainalysis",
	}
}

// GetSupportedTechnicalIndicators returns supported technical indicators
func GetSupportedTechnicalIndicators() []string {
	return []string{
		"sma", "ema", "wma", "rsi", "macd", "bollinger_bands",
		"stochastic", "williams_r", "cci", "atr", "obv",
		"vwap", "pivot_points", "fibonacci", "ichimoku",
		"parabolic_sar", "adx", "aroon", "momentum",
	}
}

// GetSupportedTimeframes returns supported timeframes
func GetSupportedTimeframes() []string {
	return []string{
		"1m", "5m", "15m", "30m", "1h", "2h", "4h",
		"6h", "8h", "12h", "1d", "3d", "1w", "1M",
	}
}

// GetEnsembleWeightingStrategies returns available weighting strategies
func GetEnsembleWeightingStrategies() []string {
	return []string{
		"equal_weight",
		"performance_weighted",
		"confidence_weighted",
		"adaptive_weighted",
		"volatility_adjusted",
		"sharpe_weighted",
	}
}
