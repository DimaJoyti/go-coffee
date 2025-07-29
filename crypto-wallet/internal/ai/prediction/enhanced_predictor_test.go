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
func createTestLoggerForEnhancedPredictor() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

func TestNewEnhancedMarketPredictor(t *testing.T) {
	logger := createTestLoggerForEnhancedPredictor()
	config := GetDefaultEnhancedPredictorConfig()

	predictor := NewEnhancedMarketPredictor(logger, config)

	assert.NotNil(t, predictor)
	assert.Equal(t, config.Enabled, predictor.config.Enabled)
	assert.Equal(t, config.UpdateInterval, predictor.config.UpdateInterval)
	assert.False(t, predictor.IsRunning())
	assert.NotNil(t, predictor.sentimentEngine)
	assert.NotNil(t, predictor.onChainEngine)
	assert.NotNil(t, predictor.technicalEngine)
	assert.NotNil(t, predictor.macroEngine)
	assert.NotNil(t, predictor.mlEngine)
	assert.NotNil(t, predictor.ensembleEngine)
	assert.NotNil(t, predictor.dataAggregator)
	assert.NotNil(t, predictor.featureExtractor)
	assert.NotNil(t, predictor.modelManager)
}

func TestEnhancedMarketPredictor_StartStop(t *testing.T) {
	logger := createTestLoggerForEnhancedPredictor()
	config := GetDefaultEnhancedPredictorConfig()

	predictor := NewEnhancedMarketPredictor(logger, config)
	ctx := context.Background()

	err := predictor.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, predictor.IsRunning())

	err = predictor.Stop()
	assert.NoError(t, err)
	assert.False(t, predictor.IsRunning())
}

func TestEnhancedMarketPredictor_StartDisabled(t *testing.T) {
	logger := createTestLoggerForEnhancedPredictor()
	config := GetDefaultEnhancedPredictorConfig()
	config.Enabled = false

	predictor := NewEnhancedMarketPredictor(logger, config)
	ctx := context.Background()

	err := predictor.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, predictor.IsRunning()) // Should remain false when disabled
}

func TestEnhancedMarketPredictor_PredictMarketEnhanced(t *testing.T) {
	logger := createTestLoggerForEnhancedPredictor()
	config := GetDefaultEnhancedPredictorConfig()

	predictor := NewEnhancedMarketPredictor(logger, config)
	ctx := context.Background()

	// Start the predictor
	err := predictor.Start(ctx)
	require.NoError(t, err)
	defer predictor.Stop()

	// Test prediction
	asset := "BTC"
	currentPrice := decimal.NewFromFloat(50000)

	prediction, err := predictor.PredictMarketEnhanced(ctx, asset, currentPrice)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)

	// Validate prediction structure
	assert.Equal(t, asset, prediction.Asset)
	assert.False(t, prediction.Timestamp.IsZero())
	assert.NotEmpty(t, prediction.Horizons)
	assert.Contains(t, []string{"bullish", "bearish", "neutral"}, prediction.OverallDirection)
	assert.True(t, prediction.OverallConfidence.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, prediction.OverallConfidence.LessThanOrEqual(decimal.NewFromFloat(1)))
	assert.Contains(t, []string{"low", "medium", "high"}, prediction.RiskLevel)
	assert.True(t, prediction.Volatility.GreaterThanOrEqual(decimal.Zero))
	assert.NotNil(t, prediction.FeatureImportance)
	assert.NotNil(t, prediction.Scenarios)
	assert.NotNil(t, prediction.Alerts)
	assert.NotNil(t, prediction.Metadata)

	// Validate horizon predictions
	for horizon, horizonPred := range prediction.Horizons {
		assert.True(t, horizon > 0)
		assert.True(t, horizonPred.PredictedPrice.GreaterThan(decimal.Zero))
		assert.True(t, horizonPred.PriceRange.Low.LessThanOrEqual(horizonPred.PredictedPrice))
		assert.True(t, horizonPred.PriceRange.High.GreaterThanOrEqual(horizonPred.PredictedPrice))
		assert.Contains(t, []string{"bullish", "bearish", "neutral"}, horizonPred.Direction)
		assert.True(t, horizonPred.Confidence.GreaterThanOrEqual(decimal.Zero))
		assert.True(t, horizonPred.Confidence.LessThanOrEqual(decimal.NewFromFloat(1)))
	}

	// Validate scenarios
	for _, scenario := range prediction.Scenarios {
		assert.NotEmpty(t, scenario.Name)
		assert.True(t, scenario.Probability.GreaterThanOrEqual(decimal.Zero))
		assert.True(t, scenario.Probability.LessThanOrEqual(decimal.NewFromFloat(1)))
		assert.True(t, scenario.PriceTarget.GreaterThan(decimal.Zero))
		assert.True(t, scenario.TimeToTarget > 0)
	}
}

func TestEnhancedMarketPredictor_GetMetrics(t *testing.T) {
	logger := createTestLoggerForEnhancedPredictor()
	config := GetDefaultEnhancedPredictorConfig()

	predictor := NewEnhancedMarketPredictor(logger, config)
	ctx := context.Background()

	// Start the predictor
	err := predictor.Start(ctx)
	require.NoError(t, err)
	defer predictor.Stop()

	metrics := predictor.GetMetrics()
	assert.NotNil(t, metrics)

	// Validate metrics structure
	assert.Contains(t, metrics, "is_running")
	assert.Contains(t, metrics, "cache_size")
	assert.Contains(t, metrics, "sentiment_enabled")
	assert.Contains(t, metrics, "onchain_enabled")
	assert.Contains(t, metrics, "technical_enabled")
	assert.Contains(t, metrics, "macro_enabled")
	assert.Contains(t, metrics, "ml_enabled")
	assert.Contains(t, metrics, "ensemble_enabled")
	assert.Contains(t, metrics, "prediction_horizons")

	assert.Equal(t, true, metrics["is_running"])
	assert.Equal(t, true, metrics["sentiment_enabled"])
	assert.Equal(t, true, metrics["onchain_enabled"])
	assert.Equal(t, true, metrics["technical_enabled"])
	assert.Equal(t, true, metrics["macro_enabled"])
	assert.Equal(t, true, metrics["ml_enabled"])
	assert.Equal(t, true, metrics["ensemble_enabled"])
}

func TestMockAdvancedSentimentEngine(t *testing.T) {
	engine := &MockAdvancedSentimentEngine{}
	ctx := context.Background()

	score, err := engine.AnalyzeSentiment(ctx, "BTC")
	assert.NoError(t, err)
	assert.True(t, score.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, score.LessThanOrEqual(decimal.NewFromFloat(1)))

	trends, err := engine.GetSentimentTrends("BTC", 24*time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, trends)
}

func TestMockAdvancedOnChainEngine(t *testing.T) {
	engine := &MockAdvancedOnChainEngine{}
	ctx := context.Background()

	score, err := engine.AnalyzeOnChain(ctx, "BTC")
	assert.NoError(t, err)
	assert.True(t, score.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, score.LessThanOrEqual(decimal.NewFromFloat(1)))

	metrics, err := engine.GetOnChainMetrics("BTC")
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.True(t, metrics.TransactionVolume.GreaterThan(decimal.Zero))
	assert.True(t, metrics.ActiveAddresses > 0)
}

func TestMockMachineLearningEngine(t *testing.T) {
	engine := &MockMachineLearningEngine{}
	ctx := context.Background()

	features := []decimal.Decimal{
		decimal.NewFromFloat(0.1),
		decimal.NewFromFloat(0.2),
		decimal.NewFromFloat(0.3),
	}

	prediction, err := engine.Predict(ctx, features)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.True(t, prediction.Prediction.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, prediction.Confidence.GreaterThanOrEqual(decimal.Zero))
	assert.NotEmpty(t, prediction.FeatureImportance)

	performance, err := engine.GetModelPerformance()
	assert.NoError(t, err)
	assert.NotNil(t, performance)
	assert.True(t, performance.Accuracy.GreaterThanOrEqual(decimal.Zero))
}

func TestMockDataAggregator(t *testing.T) {
	aggregator := &MockDataAggregator{}
	ctx := context.Background()

	data, err := aggregator.AggregateData(ctx, "BTC")
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotEmpty(t, data.PriceData)
	assert.NotEmpty(t, data.VolumeData)
	assert.NotEmpty(t, data.SentimentData)
	assert.NotNil(t, data.OnChainData)
	assert.NotNil(t, data.MacroData)

	quality := aggregator.GetDataQuality("test_source")
	assert.True(t, quality.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, quality.LessThanOrEqual(decimal.NewFromFloat(1)))
}

func TestGetDefaultEnhancedPredictorConfig(t *testing.T) {
	config := GetDefaultEnhancedPredictorConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 5*time.Minute, config.UpdateInterval)
	assert.NotEmpty(t, config.PredictionHorizons)
	assert.Equal(t, 1*time.Hour, config.CacheRetentionPeriod)

	// Check sentiment config
	assert.True(t, config.SentimentConfig.Enabled)
	assert.NotEmpty(t, config.SentimentConfig.Sources)
	assert.NotEmpty(t, config.SentimentConfig.LanguageModels)

	// Check on-chain config
	assert.True(t, config.OnChainConfig.Enabled)
	assert.NotEmpty(t, config.OnChainConfig.Metrics)
	assert.NotEmpty(t, config.OnChainConfig.Chains)

	// Check technical config
	assert.True(t, config.TechnicalConfig.Enabled)
	assert.NotEmpty(t, config.TechnicalConfig.Indicators)
	assert.NotEmpty(t, config.TechnicalConfig.Timeframes)

	// Check macro config
	assert.True(t, config.MacroConfig.Enabled)
	assert.NotEmpty(t, config.MacroConfig.Factors)

	// Check ML config
	assert.True(t, config.MLConfig.Enabled)
	assert.NotEmpty(t, config.MLConfig.Models)

	// Check ensemble config
	assert.True(t, config.EnsembleConfig.Enabled)
	assert.NotEmpty(t, config.EnsembleConfig.CombinationMethod)
}

func TestValidateEnhancedPredictorConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultEnhancedPredictorConfig()
	err := ValidateEnhancedPredictorConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultEnhancedPredictorConfig()
	disabledConfig.Enabled = false
	err = ValidateEnhancedPredictorConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []EnhancedPredictorConfig{
		// Invalid update interval
		{
			Enabled:        true,
			UpdateInterval: 0,
		},
		// No prediction horizons
		{
			Enabled:            true,
			UpdateInterval:     5 * time.Minute,
			PredictionHorizons: []time.Duration{},
		},
		// Invalid cache retention period
		{
			Enabled:              true,
			UpdateInterval:       5 * time.Minute,
			PredictionHorizons:   []time.Duration{1 * time.Hour},
			CacheRetentionPeriod: 0,
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateEnhancedPredictorConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestConfigVariants(t *testing.T) {
	// Test high frequency config
	hfConfig := GetHighFrequencyConfig()
	assert.True(t, hfConfig.UpdateInterval < GetDefaultEnhancedPredictorConfig().UpdateInterval)
	assert.True(t, len(hfConfig.PredictionHorizons) <= len(GetDefaultEnhancedPredictorConfig().PredictionHorizons))

	// Test long term config
	ltConfig := GetLongTermConfig()
	assert.True(t, ltConfig.UpdateInterval > GetDefaultEnhancedPredictorConfig().UpdateInterval)
	assert.True(t, len(ltConfig.PredictionHorizons) >= len(GetDefaultEnhancedPredictorConfig().PredictionHorizons))

	// Test research config
	researchConfig := GetResearchConfig()
	assert.True(t, len(researchConfig.PredictionHorizons) > len(GetDefaultEnhancedPredictorConfig().PredictionHorizons))
	assert.True(t, researchConfig.CacheRetentionPeriod > GetDefaultEnhancedPredictorConfig().CacheRetentionPeriod)

	// Validate all configs
	assert.NoError(t, ValidateEnhancedPredictorConfig(hfConfig))
	assert.NoError(t, ValidateEnhancedPredictorConfig(ltConfig))
	assert.NoError(t, ValidateEnhancedPredictorConfig(researchConfig))
}

func TestUtilityFunctions(t *testing.T) {
	// Test supported sentiment sources
	sources := GetSupportedSentimentSources()
	assert.NotEmpty(t, sources)
	assert.Contains(t, sources, "twitter")
	assert.Contains(t, sources, "reddit")

	// Test supported on-chain metrics
	metrics := GetSupportedOnChainMetrics()
	assert.NotEmpty(t, metrics)
	assert.Contains(t, metrics, "transaction_volume")
	assert.Contains(t, metrics, "active_addresses")

	// Test supported technical indicators
	indicators := GetSupportedTechnicalIndicators()
	assert.NotEmpty(t, indicators)
	assert.Contains(t, indicators, "rsi")
	assert.Contains(t, indicators, "macd")

	// Test supported ML models
	models := GetSupportedMLModels()
	assert.NotEmpty(t, models)
	assert.Contains(t, models, "lstm")
	assert.Contains(t, models, "transformer")

	// Test optimal config for use case
	config, err := GetOptimalConfigForUseCase("high_frequency")
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Test invalid use case
	_, err = GetOptimalConfigForUseCase("invalid_use_case")
	assert.Error(t, err)
}

func TestCacheManagement(t *testing.T) {
	logger := createTestLoggerForEnhancedPredictor()
	config := GetDefaultEnhancedPredictorConfig()
	config.CacheRetentionPeriod = 100 * time.Millisecond // Short retention for testing

	predictor := NewEnhancedMarketPredictor(logger, config)
	ctx := context.Background()

	// Start the predictor
	err := predictor.Start(ctx)
	require.NoError(t, err)
	defer predictor.Stop()

	// Create first prediction
	asset := "BTC"
	currentPrice := decimal.NewFromFloat(50000)

	prediction1, err := predictor.PredictMarketEnhanced(ctx, asset, currentPrice)
	assert.NoError(t, err)
	assert.NotNil(t, prediction1)

	// Create second prediction (should be cached)
	prediction2, err := predictor.PredictMarketEnhanced(ctx, asset, currentPrice)
	assert.NoError(t, err)
	assert.NotNil(t, prediction2)

	// Should be the same prediction (from cache)
	assert.Equal(t, prediction1.Timestamp, prediction2.Timestamp)

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Create third prediction (should be new)
	prediction3, err := predictor.PredictMarketEnhanced(ctx, asset, currentPrice)
	assert.NoError(t, err)
	assert.NotNil(t, prediction3)

	// Should be a new prediction
	assert.True(t, prediction3.Timestamp.After(prediction1.Timestamp))
}
