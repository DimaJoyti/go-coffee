package prediction

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLoggerForPrediction() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

func TestNewMarketPredictor(t *testing.T) {
	logger := createTestLoggerForPrediction()
	config := GetDefaultMarketPredictorConfig()

	predictor := NewMarketPredictor(logger, config)

	assert.NotNil(t, predictor)
	assert.Equal(t, config.Enabled, predictor.config.Enabled)
	assert.Equal(t, config.UpdateInterval, predictor.config.UpdateInterval)
	assert.False(t, predictor.IsRunning())
	assert.NotNil(t, predictor.sentimentAnalyzer)
	assert.NotNil(t, predictor.onChainAnalyzer)
	assert.NotNil(t, predictor.technicalAnalyzer)
	assert.NotNil(t, predictor.macroAnalyzer)
	assert.NotNil(t, predictor.ensembleModel)
}

func TestMarketPredictor_StartStop(t *testing.T) {
	logger := createTestLoggerForPrediction()
	config := GetDefaultMarketPredictorConfig()

	predictor := NewMarketPredictor(logger, config)
	ctx := context.Background()

	err := predictor.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, predictor.IsRunning())

	err = predictor.Stop()
	assert.NoError(t, err)
	assert.False(t, predictor.IsRunning())
}

func TestMarketPredictor_StartDisabled(t *testing.T) {
	logger := createTestLoggerForPrediction()
	config := GetDefaultMarketPredictorConfig()
	config.Enabled = false

	predictor := NewMarketPredictor(logger, config)
	ctx := context.Background()

	err := predictor.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, predictor.IsRunning()) // Should remain false when disabled
}

func TestMarketPredictor_PredictMarket(t *testing.T) {
	logger := createTestLoggerForPrediction()
	config := GetDefaultMarketPredictorConfig()

	predictor := NewMarketPredictor(logger, config)
	ctx := context.Background()

	// Start the predictor
	err := predictor.Start(ctx)
	require.NoError(t, err)
	defer predictor.Stop()

	// Test prediction for Bitcoin
	asset := "BTC"
	currentPrice := decimal.NewFromFloat(30000)

	prediction, err := predictor.PredictMarket(ctx, asset, currentPrice)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)

	// Validate prediction result
	assert.Equal(t, asset, prediction.Asset)
	assert.NotEmpty(t, prediction.ID)
	assert.Equal(t, currentPrice, prediction.CurrentPrice)
	assert.True(t, prediction.PredictedPrice.GreaterThan(decimal.Zero))
	assert.Contains(t, []string{"bullish", "bearish", "neutral"}, prediction.Direction)
	assert.True(t, prediction.Confidence.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, prediction.Confidence.LessThanOrEqual(decimal.NewFromFloat(1)))
	assert.Contains(t, []string{"low", "medium", "high"}, prediction.RiskLevel)

	// Check prediction components
	assert.True(t, prediction.SentimentScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, prediction.OnChainScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, prediction.TechnicalScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, prediction.MacroScore.GreaterThanOrEqual(decimal.Zero))

	// Check prediction elements
	assert.NotNil(t, prediction.Signals)
	assert.NotNil(t, prediction.Factors)
	assert.NotNil(t, prediction.Scenarios)
	assert.NotNil(t, prediction.Alerts)
}

func TestMarketPredictor_PredictMultipleAssets(t *testing.T) {
	logger := createTestLoggerForPrediction()
	config := GetDefaultMarketPredictorConfig()

	predictor := NewMarketPredictor(logger, config)
	ctx := context.Background()

	// Start the predictor
	err := predictor.Start(ctx)
	require.NoError(t, err)
	defer predictor.Stop()

	// Test predictions for multiple assets
	assets := map[string]decimal.Decimal{
		"BTC":  decimal.NewFromFloat(30000),
		"ETH":  decimal.NewFromFloat(2000),
		"SOL":  decimal.NewFromFloat(25),
		"USDC": decimal.NewFromFloat(1),
	}

	predictions, err := predictor.PredictMultipleAssets(ctx, assets)
	assert.NoError(t, err)
	assert.NotNil(t, predictions)

	// Should have predictions for all assets
	for asset := range assets {
		prediction, exists := predictions[asset]
		assert.True(t, exists, "Should have prediction for %s", asset)
		assert.NotNil(t, prediction)
		assert.Equal(t, asset, prediction.Asset)
	}
}

func TestMarketPredictor_GetPredictionMetrics(t *testing.T) {
	logger := createTestLoggerForPrediction()
	config := GetDefaultMarketPredictorConfig()

	predictor := NewMarketPredictor(logger, config)
	ctx := context.Background()

	// Start the predictor
	err := predictor.Start(ctx)
	require.NoError(t, err)
	defer predictor.Stop()

	// Get prediction metrics
	metrics := predictor.GetPredictionMetrics()
	assert.NotNil(t, metrics)

	// Validate metrics
	assert.Contains(t, metrics, "cached_predictions")
	assert.Contains(t, metrics, "is_running")
	assert.Contains(t, metrics, "sentiment_analyzer")
	assert.Contains(t, metrics, "onchain_analyzer")
	assert.Contains(t, metrics, "technical_analyzer")
	assert.Contains(t, metrics, "macro_analyzer")
	assert.Contains(t, metrics, "ensemble_model")
	assert.Contains(t, metrics, "cached_models")

	assert.Equal(t, true, metrics["is_running"])
	assert.Equal(t, true, metrics["sentiment_analyzer"])
	assert.Equal(t, true, metrics["onchain_analyzer"])
	assert.Equal(t, true, metrics["technical_analyzer"])
	assert.Equal(t, true, metrics["macro_analyzer"])
	assert.Equal(t, true, metrics["ensemble_model"])
}

func TestMarketPredictor_Caching(t *testing.T) {
	logger := createTestLoggerForPrediction()
	config := GetDefaultMarketPredictorConfig()
	config.CacheTimeout = 1 * time.Hour // Long cache timeout

	predictor := NewMarketPredictor(logger, config)
	ctx := context.Background()

	// Start the predictor
	err := predictor.Start(ctx)
	require.NoError(t, err)
	defer predictor.Stop()

	asset := "BTC"
	currentPrice := decimal.NewFromFloat(30000)

	// First prediction
	prediction1, err := predictor.PredictMarket(ctx, asset, currentPrice)
	assert.NoError(t, err)
	assert.NotNil(t, prediction1)

	// Second prediction should return cached result
	prediction2, err := predictor.PredictMarket(ctx, asset, currentPrice)
	assert.NoError(t, err)
	assert.NotNil(t, prediction2)

	// Should be the same result (cached)
	assert.Equal(t, prediction1.ID, prediction2.ID)
	assert.Equal(t, prediction1.Timestamp, prediction2.Timestamp)
}

func TestGetDefaultMarketPredictorConfig(t *testing.T) {
	config := GetDefaultMarketPredictorConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 15*time.Minute, config.UpdateInterval)
	assert.Equal(t, 1*time.Hour, config.CacheTimeout)
	assert.NotEmpty(t, config.PredictionHorizons)
	assert.True(t, config.ConfidenceThreshold.GreaterThan(decimal.Zero))

	// Check sentiment config
	assert.True(t, config.SentimentConfig.Enabled)
	assert.NotEmpty(t, config.SentimentConfig.Sources)
	assert.NotEmpty(t, config.SentimentConfig.LanguageModels)

	// Check on-chain config
	assert.True(t, config.OnChainConfig.Enabled)
	assert.NotEmpty(t, config.OnChainConfig.Metrics)

	// Check technical config
	assert.True(t, config.TechnicalConfig.Enabled)
	assert.NotEmpty(t, config.TechnicalConfig.Indicators)
	assert.NotEmpty(t, config.TechnicalConfig.Timeframes)

	// Check macro config
	assert.True(t, config.MacroConfig.Enabled)
	assert.NotEmpty(t, config.MacroConfig.Indicators)

	// Check ensemble config
	assert.True(t, config.EnsembleConfig.Enabled)
	assert.NotEmpty(t, config.EnsembleConfig.Models)
	assert.NotEmpty(t, config.EnsembleConfig.ModelWeights)

	// Check alert thresholds
	assert.True(t, config.AlertThresholds.HighConfidenceBull.GreaterThan(decimal.Zero))
	assert.True(t, config.AlertThresholds.HighConfidenceBear.GreaterThan(decimal.Zero))

	// Check data sources
	assert.NotEmpty(t, config.DataSources)
}

func TestValidateMarketPredictorConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultMarketPredictorConfig()
	err := ValidateMarketPredictorConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultMarketPredictorConfig()
	disabledConfig.Enabled = false
	err = ValidateMarketPredictorConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []MarketPredictorConfig{
		// Invalid update interval
		{
			Enabled:        true,
			UpdateInterval: 0,
		},
		// Invalid cache timeout
		{
			Enabled:        true,
			UpdateInterval: 15 * time.Minute,
			CacheTimeout:   0,
		},
		// No prediction horizons
		{
			Enabled:            true,
			UpdateInterval:     15 * time.Minute,
			CacheTimeout:       1 * time.Hour,
			PredictionHorizons: []time.Duration{},
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateMarketPredictorConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestPredictorConfigVariants(t *testing.T) {
	// Test conservative config
	conservativeConfig := GetConservativePredictorConfig()
	assert.True(t, conservativeConfig.ConfidenceThreshold.GreaterThan(
		GetDefaultMarketPredictorConfig().ConfidenceThreshold))

	// Test aggressive config
	aggressiveConfig := GetAggressivePredictorConfig()
	assert.True(t, aggressiveConfig.ConfidenceThreshold.LessThan(
		GetDefaultMarketPredictorConfig().ConfidenceThreshold))

	// Test day trading config
	dayTradingConfig := GetDayTradingPredictorConfig()
	assert.True(t, dayTradingConfig.UpdateInterval < GetDefaultMarketPredictorConfig().UpdateInterval)

	// Validate all configs
	assert.NoError(t, ValidateMarketPredictorConfig(conservativeConfig))
	assert.NoError(t, ValidateMarketPredictorConfig(aggressiveConfig))
	assert.NoError(t, ValidateMarketPredictorConfig(dayTradingConfig))
}

func TestPredictionEngines(t *testing.T) {
	logger := createTestLoggerForPrediction()

	// Test sentiment analyzer
	sentimentConfig := SentimentAnalysisConfig{Enabled: true}
	sentimentAnalyzer := NewSentimentAnalyzer(logger, sentimentConfig)
	assert.NotNil(t, sentimentAnalyzer)

	ctx := context.Background()
	err := sentimentAnalyzer.Start(ctx)
	assert.NoError(t, err)

	sentimentScore, err := sentimentAnalyzer.AnalyzeSentiment(ctx, "BTC")
	assert.NoError(t, err)
	assert.True(t, sentimentScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, sentimentScore.LessThanOrEqual(decimal.NewFromFloat(1)))

	err = sentimentAnalyzer.Stop()
	assert.NoError(t, err)

	// Test on-chain analyzer
	onChainConfig := OnChainAnalysisConfig{Enabled: true}
	onChainAnalyzer := NewOnChainAnalyzer(logger, onChainConfig)
	assert.NotNil(t, onChainAnalyzer)

	err = onChainAnalyzer.Start(ctx)
	assert.NoError(t, err)

	onChainScore, err := onChainAnalyzer.AnalyzeOnChain(ctx, "BTC")
	assert.NoError(t, err)
	assert.True(t, onChainScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, onChainScore.LessThanOrEqual(decimal.NewFromFloat(1)))

	err = onChainAnalyzer.Stop()
	assert.NoError(t, err)

	// Test technical analyzer
	technicalConfig := TechnicalAnalysisConfig{Enabled: true}
	technicalAnalyzer := NewTechnicalAnalyzer(logger, technicalConfig)
	assert.NotNil(t, technicalAnalyzer)

	err = technicalAnalyzer.Start(ctx)
	assert.NoError(t, err)

	technicalScore, err := technicalAnalyzer.AnalyzeTechnical(ctx, "BTC", decimal.NewFromFloat(30000))
	assert.NoError(t, err)
	assert.True(t, technicalScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, technicalScore.LessThanOrEqual(decimal.NewFromFloat(1)))

	err = technicalAnalyzer.Stop()
	assert.NoError(t, err)
}

func TestMarketPredictorUtilityFunctions(t *testing.T) {
	// Test prediction horizon descriptions
	horizonDescriptions := GetPredictionHorizonDescription()
	assert.NotEmpty(t, horizonDescriptions)
	assert.Contains(t, horizonDescriptions, "1h")
	assert.Contains(t, horizonDescriptions, "1d")

	// Test analysis type descriptions
	analysisDescriptions := GetAnalysisTypeDescription()
	assert.NotEmpty(t, analysisDescriptions)
	assert.Contains(t, analysisDescriptions, "sentiment")
	assert.Contains(t, analysisDescriptions, "technical")

	// Test confidence level descriptions
	confidenceDescriptions := GetConfidenceLevelDescription()
	assert.NotEmpty(t, confidenceDescriptions)
	assert.Contains(t, confidenceDescriptions, "high")
	assert.Contains(t, confidenceDescriptions, "low")

	// Test supported data sources
	dataSources := GetSupportedDataSources()
	assert.NotEmpty(t, dataSources)
	assert.Contains(t, dataSources, "coingecko")
	assert.Contains(t, dataSources, "glassnode")

	// Test supported technical indicators
	indicators := GetSupportedTechnicalIndicators()
	assert.NotEmpty(t, indicators)
	assert.Contains(t, indicators, "rsi")
	assert.Contains(t, indicators, "macd")

	// Test supported timeframes
	timeframes := GetSupportedTimeframes()
	assert.NotEmpty(t, timeframes)
	assert.Contains(t, timeframes, "1h")
	assert.Contains(t, timeframes, "1d")

	// Test ensemble weighting strategies
	strategies := GetEnsembleWeightingStrategies()
	assert.NotEmpty(t, strategies)
	assert.Contains(t, strategies, "performance_weighted")
	assert.Contains(t, strategies, "confidence_weighted")
}
