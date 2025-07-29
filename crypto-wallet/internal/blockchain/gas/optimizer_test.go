package gas

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
func createTestLoggerForGas() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

// Helper function to create test optimization request
func createTestOptimizationRequest() *OptimizationRequest {
	return &OptimizationRequest{
		TransactionType:   "transfer",
		Priority:          "medium",
		TargetConfirmTime: 5 * time.Minute,
		MaxGasPrice:       decimal.NewFromFloat(100), // 100 gwei
		GasLimit:          21000,
		Value:             decimal.NewFromFloat(1), // 1 ETH
		IsReplacement:     false,
		UserPreferences: UserPreferences{
			CostOptimization:  true,
			SpeedOptimization: false,
			MaxCostTolerance:  decimal.NewFromFloat(0.01), // 0.01 ETH
			RiskTolerance:     "medium",
		},
	}
}

func TestNewGasOptimizer(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig()

	optimizer := NewGasOptimizer(logger, config)

	assert.NotNil(t, optimizer)
	assert.Equal(t, config.Enabled, optimizer.config.Enabled)
	assert.Equal(t, config.UpdateInterval, optimizer.config.UpdateInterval)
	assert.False(t, optimizer.IsRunning())
	assert.NotNil(t, optimizer.eip1559Optimizer)
	assert.NotNil(t, optimizer.historicalAnalyzer)
	assert.NotNil(t, optimizer.congestionMonitor)
	assert.NotNil(t, optimizer.predictionEngine)
	assert.NotNil(t, optimizer.networkMetrics)
}

func TestGasOptimizer_StartStop(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig()

	optimizer := NewGasOptimizer(logger, config)
	ctx := context.Background()

	err := optimizer.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, optimizer.IsRunning())

	err = optimizer.Stop()
	assert.NoError(t, err)
	assert.False(t, optimizer.IsRunning())
}

func TestGasOptimizer_StartDisabled(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig()
	config.Enabled = false

	optimizer := NewGasOptimizer(logger, config)
	ctx := context.Background()

	err := optimizer.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, optimizer.IsRunning()) // Should remain false when disabled
}

func TestGasOptimizer_OptimizeGasPrice(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig()

	optimizer := NewGasOptimizer(logger, config)
	ctx := context.Background()

	// Start the optimizer
	err := optimizer.Start(ctx)
	require.NoError(t, err)
	defer optimizer.Stop()

	// Wait for initial network metrics update
	time.Sleep(100 * time.Millisecond)

	// Create optimization request
	request := createTestOptimizationRequest()

	// Optimize gas price
	result, err := optimizer.OptimizeGasPrice(ctx, request)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Validate result
	assert.NotEmpty(t, result.Strategy)
	assert.True(t, result.GasPrice.GreaterThan(decimal.Zero))
	assert.True(t, result.EstimatedCost.GreaterThan(decimal.Zero))
	assert.True(t, result.EstimatedTime > 0)
	assert.True(t, result.Confidence.GreaterThan(decimal.Zero))
	assert.NotEmpty(t, result.Reasoning)
	assert.False(t, result.Timestamp.IsZero())
}

func TestGasOptimizer_OptimizeGasPriceDifferentPriorities(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig()

	optimizer := NewGasOptimizer(logger, config)
	ctx := context.Background()

	// Start the optimizer
	err := optimizer.Start(ctx)
	require.NoError(t, err)
	defer optimizer.Stop()

	// Wait for initial network metrics update
	time.Sleep(100 * time.Millisecond)

	priorities := []string{"low", "medium", "high", "urgent"}
	var results []*OptimizationResult

	for _, priority := range priorities {
		request := createTestOptimizationRequest()
		request.Priority = priority

		result, err := optimizer.OptimizeGasPrice(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		results = append(results, result)
	}

	// Verify that higher priority results in higher gas prices
	for i := 1; i < len(results); i++ {
		assert.True(t, results[i].GasPrice.GreaterThanOrEqual(results[i-1].GasPrice),
			"Higher priority should result in higher or equal gas price")
	}
}

func TestGasOptimizer_GetNetworkMetrics(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig()

	optimizer := NewGasOptimizer(logger, config)
	ctx := context.Background()

	// Start the optimizer
	err := optimizer.Start(ctx)
	require.NoError(t, err)
	defer optimizer.Stop()

	// Wait for initial network metrics update
	time.Sleep(100 * time.Millisecond)

	metrics := optimizer.GetNetworkMetrics()
	assert.NotNil(t, metrics)
	assert.True(t, metrics.CurrentBaseFee.GreaterThan(decimal.Zero))
	assert.True(t, metrics.RecommendedPriority.GreaterThan(decimal.Zero))
	assert.True(t, metrics.NetworkUtilization.GreaterThanOrEqual(decimal.Zero))
	assert.Contains(t, []string{"low", "medium", "high"}, metrics.CongestionLevel)
	assert.True(t, metrics.AverageConfirmTime > 0)
	assert.False(t, metrics.LastUpdated.IsZero())
}

func TestGasOptimizer_GetGasHistory(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig()

	optimizer := NewGasOptimizer(logger, config)
	ctx := context.Background()

	// Start the optimizer
	err := optimizer.Start(ctx)
	require.NoError(t, err)
	defer optimizer.Stop()

	// Wait for some history to accumulate
	time.Sleep(200 * time.Millisecond)

	history := optimizer.GetGasHistory(10)
	assert.NotNil(t, history)
	assert.True(t, len(history) > 0)

	// Validate history entries
	for _, entry := range history {
		assert.True(t, entry.BaseFee.GreaterThan(decimal.Zero))
		assert.True(t, entry.PriorityFee.GreaterThan(decimal.Zero))
		assert.True(t, entry.GasPrice.GreaterThan(decimal.Zero))
		assert.False(t, entry.Timestamp.IsZero())
	}
}

func TestGasOptimizer_GetMetrics(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig()

	optimizer := NewGasOptimizer(logger, config)
	ctx := context.Background()

	// Start the optimizer
	err := optimizer.Start(ctx)
	require.NoError(t, err)
	defer optimizer.Stop()

	metrics := optimizer.GetMetrics()
	assert.NotNil(t, metrics)

	// Validate metrics structure
	assert.Contains(t, metrics, "is_running")
	assert.Contains(t, metrics, "gas_history_size")
	assert.Contains(t, metrics, "cache_size")
	assert.Contains(t, metrics, "eip1559_enabled")
	assert.Contains(t, metrics, "historical_enabled")
	assert.Contains(t, metrics, "congestion_enabled")
	assert.Contains(t, metrics, "prediction_enabled")
	assert.Contains(t, metrics, "optimization_strategies")

	assert.Equal(t, true, metrics["is_running"])
	assert.Equal(t, true, metrics["eip1559_enabled"])
	assert.Equal(t, true, metrics["historical_enabled"])
	assert.Equal(t, true, metrics["congestion_enabled"])
	assert.Equal(t, true, metrics["prediction_enabled"])
}

func TestEIP1559Optimizer_Optimize(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig().EIP1559Config

	optimizer := &EIP1559Optimizer{
		logger: logger,
		config: config,
	}

	request := createTestOptimizationRequest()
	metrics := &NetworkMetrics{
		CurrentBaseFee:      decimal.NewFromFloat(15), // 15 gwei
		RecommendedPriority: decimal.NewFromFloat(2),  // 2 gwei
		NetworkUtilization:  decimal.NewFromFloat(0.7),
		CongestionLevel:     "medium",
	}

	result, err := optimizer.Optimize(context.Background(), request, metrics)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "eip1559", result.Strategy)
	assert.True(t, result.MaxFeePerGas.GreaterThan(metrics.CurrentBaseFee))
	assert.True(t, result.MaxPriorityFeePerGas.GreaterThan(decimal.Zero))
	assert.True(t, result.EstimatedCost.GreaterThan(decimal.Zero))
	assert.True(t, result.EstimatedTime > 0)
	assert.True(t, result.Confidence.GreaterThan(decimal.Zero))
	assert.NotEmpty(t, result.Reasoning)
}

func TestHistoricalAnalyzer_Optimize(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig().HistoricalConfig

	analyzer := &HistoricalAnalyzer{
		logger: logger,
		config: config,
	}

	request := createTestOptimizationRequest()
	
	// Create mock historical data
	history := []GasDataPoint{
		{
			Timestamp:        time.Now().Add(-10 * time.Minute),
			GasPrice:         decimal.NewFromFloat(20),
			ConfirmationTime: 3 * time.Minute,
		},
		{
			Timestamp:        time.Now().Add(-5 * time.Minute),
			GasPrice:         decimal.NewFromFloat(25),
			ConfirmationTime: 2 * time.Minute,
		},
		{
			Timestamp:        time.Now().Add(-1 * time.Minute),
			GasPrice:         decimal.NewFromFloat(30),
			ConfirmationTime: 1 * time.Minute,
		},
	}

	result, err := analyzer.Optimize(context.Background(), request, history)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "historical", result.Strategy)
	assert.True(t, result.GasPrice.GreaterThan(decimal.Zero))
	assert.True(t, result.EstimatedCost.GreaterThan(decimal.Zero))
	assert.True(t, result.EstimatedTime > 0)
	assert.True(t, result.Confidence.GreaterThan(decimal.Zero))
	assert.NotEmpty(t, result.Reasoning)
}

func TestCongestionMonitor_Optimize(t *testing.T) {
	logger := createTestLoggerForGas()
	config := GetDefaultGasOptimizerConfig().CongestionConfig

	monitor := &CongestionMonitor{
		logger: logger,
		config: config,
		metrics: &NetworkMetrics{
			CurrentBaseFee:      decimal.NewFromFloat(15),
			RecommendedPriority: decimal.NewFromFloat(2),
			NetworkUtilization:  decimal.NewFromFloat(0.8),
			CongestionLevel:     "high",
			PendingTransactions: 75000,
		},
	}

	request := createTestOptimizationRequest()

	result, err := monitor.Optimize(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "congestion_based", result.Strategy)
	assert.True(t, result.GasPrice.GreaterThan(decimal.Zero))
	assert.True(t, result.EstimatedCost.GreaterThan(decimal.Zero))
	assert.True(t, result.EstimatedTime > 0)
	assert.True(t, result.Confidence.GreaterThan(decimal.Zero))
	assert.NotEmpty(t, result.Reasoning)
}

func TestGetDefaultGasOptimizerConfig(t *testing.T) {
	config := GetDefaultGasOptimizerConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 30*time.Second, config.UpdateInterval)
	assert.Equal(t, 2*time.Hour, config.HistoryRetentionPeriod)
	assert.Equal(t, 1000, config.MaxHistorySize)
	assert.NotEmpty(t, config.OptimizationStrategies)

	// Check EIP-1559 config
	assert.True(t, config.EIP1559Config.Enabled)
	assert.True(t, config.EIP1559Config.BaseFeeMultiplier.GreaterThan(decimal.NewFromFloat(1)))

	// Check historical config
	assert.True(t, config.HistoricalConfig.Enabled)
	assert.True(t, config.HistoricalConfig.AnalysisWindow > 0)

	// Check congestion config
	assert.True(t, config.CongestionConfig.Enabled)
	assert.NotEmpty(t, config.CongestionConfig.CongestionThresholds)

	// Check prediction config
	assert.True(t, config.PredictionConfig.Enabled)
	assert.NotEmpty(t, config.PredictionConfig.PredictionMethods)

	// Check safety margins
	assert.True(t, config.SafetyMargins.MinGasPrice.GreaterThan(decimal.Zero))
	assert.True(t, config.SafetyMargins.MaxGasPrice.GreaterThan(config.SafetyMargins.MinGasPrice))
}

func TestValidateGasOptimizerConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultGasOptimizerConfig()
	err := ValidateGasOptimizerConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultGasOptimizerConfig()
	disabledConfig.Enabled = false
	err = ValidateGasOptimizerConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []GasOptimizerConfig{
		// Invalid update interval
		{
			Enabled:        true,
			UpdateInterval: 0,
		},
		// Invalid history retention period
		{
			Enabled:                true,
			UpdateInterval:         30 * time.Second,
			HistoryRetentionPeriod: 0,
		},
		// Invalid max history size
		{
			Enabled:                true,
			UpdateInterval:         30 * time.Second,
			HistoryRetentionPeriod: 2 * time.Hour,
			MaxHistorySize:         0,
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateGasOptimizerConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestConfigVariants(t *testing.T) {
	// Test high frequency config
	hfConfig := GetHighFrequencyConfig()
	assert.True(t, hfConfig.UpdateInterval < GetDefaultGasOptimizerConfig().UpdateInterval)
	assert.True(t, hfConfig.EIP1559Config.BaseFeeMultiplier.GreaterThan(GetDefaultGasOptimizerConfig().EIP1559Config.BaseFeeMultiplier))

	// Test cost optimized config
	costConfig := GetCostOptimizedConfig()
	assert.True(t, costConfig.UpdateInterval > GetDefaultGasOptimizerConfig().UpdateInterval)
	assert.True(t, costConfig.EIP1559Config.BaseFeeMultiplier.LessThan(GetDefaultGasOptimizerConfig().EIP1559Config.BaseFeeMultiplier))

	// Test balanced config
	balancedConfig := GetBalancedConfig()
	assert.NotNil(t, balancedConfig)

	// Validate all configs
	assert.NoError(t, ValidateGasOptimizerConfig(hfConfig))
	assert.NoError(t, ValidateGasOptimizerConfig(costConfig))
	assert.NoError(t, ValidateGasOptimizerConfig(balancedConfig))
}

func TestUtilityFunctions(t *testing.T) {
	// Test supported strategies
	strategies := GetSupportedOptimizationStrategies()
	assert.NotEmpty(t, strategies)
	assert.Contains(t, strategies, "eip1559")
	assert.Contains(t, strategies, "hybrid")

	// Test supported priority fee strategies
	priorityStrategies := GetSupportedPriorityFeeStrategies()
	assert.NotEmpty(t, priorityStrategies)
	assert.Contains(t, priorityStrategies, "dynamic")

	// Test supported aggressiveness levels
	levels := GetSupportedAggressivenessLevels()
	assert.NotEmpty(t, levels)
	assert.Contains(t, levels, "moderate")

	// Test optimal config for use case
	config, err := GetOptimalConfigForUseCase("high_frequency_trading")
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Test invalid use case
	_, err = GetOptimalConfigForUseCase("invalid_use_case")
	assert.Error(t, err)
}
