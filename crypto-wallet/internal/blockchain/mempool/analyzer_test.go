package mempool

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLoggerForMempool() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

// Helper function to create test transaction
func createTestTransaction(gasPrice *big.Int, gasLimit uint64, nonce uint64) *types.Transaction {
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	value := big.NewInt(1000000000000000000) // 1 ETH
	data := []byte{}

	return types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)
}

// Helper function to create EIP-1559 transaction
func createEIP1559Transaction(gasFeeCap, gasTipCap *big.Int, gasLimit uint64, nonce uint64) *types.Transaction {
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	value := big.NewInt(1000000000000000000) // 1 ETH
	data := []byte{}

	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(1),
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      data,
	})
}

func TestNewMempoolAnalyzer(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)

	assert.NotNil(t, analyzer)
	assert.Equal(t, config.Enabled, analyzer.config.Enabled)
	assert.Equal(t, config.UpdateInterval, analyzer.config.UpdateInterval)
	assert.False(t, analyzer.IsRunning())
	assert.NotNil(t, analyzer.gasTracker)
	assert.NotNil(t, analyzer.congestionModel)
	assert.NotNil(t, analyzer.gasPredictor)
	assert.NotNil(t, analyzer.timeEstimator)
	assert.NotNil(t, analyzer.priorityAnalyzer)
}

func TestMempoolAnalyzer_StartStop(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)
	ctx := context.Background()

	err := analyzer.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, analyzer.IsRunning())

	err = analyzer.Stop()
	assert.NoError(t, err)
	assert.False(t, analyzer.IsRunning())
}

func TestMempoolAnalyzer_StartDisabled(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()
	config.Enabled = false

	analyzer := NewMempoolAnalyzer(logger, config)
	ctx := context.Background()

	err := analyzer.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, analyzer.IsRunning()) // Should remain false when disabled
}

func TestMempoolAnalyzer_AddTransaction(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Create test transaction
	tx := createTestTransaction(big.NewInt(20000000000), 21000, 1) // 20 gwei

	// Add transaction
	err = analyzer.AddTransaction(tx)
	assert.NoError(t, err)

	// Check that transaction was added
	analyzer.dataMutex.RLock()
	assert.Len(t, analyzer.transactions, 1)
	assert.Contains(t, analyzer.transactions, tx.Hash())
	analyzer.dataMutex.RUnlock()
}

func TestMempoolAnalyzer_AddEIP1559Transaction(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Create EIP-1559 transaction
	tx := createEIP1559Transaction(
		big.NewInt(30000000000), // 30 gwei fee cap
		big.NewInt(2000000000),  // 2 gwei tip cap
		21000, 1)

	// Add transaction
	err = analyzer.AddTransaction(tx)
	assert.NoError(t, err)

	// Check that transaction was added with correct type
	analyzer.dataMutex.RLock()
	mempoolTx := analyzer.transactions[tx.Hash()]
	assert.Equal(t, "eip1559", mempoolTx.TransactionType)
	assert.Equal(t, decimal.NewFromInt(30000000000), mempoolTx.GasFeeCap)
	assert.Equal(t, decimal.NewFromInt(2000000000), mempoolTx.GasTipCap)
	analyzer.dataMutex.RUnlock()
}

func TestMempoolAnalyzer_RemoveTransaction(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Add transaction
	tx := createTestTransaction(big.NewInt(20000000000), 21000, 1)
	err = analyzer.AddTransaction(tx)
	require.NoError(t, err)

	// Remove transaction
	analyzer.RemoveTransaction(tx.Hash())

	// Check that transaction was removed
	analyzer.dataMutex.RLock()
	assert.Len(t, analyzer.transactions, 0)
	analyzer.dataMutex.RUnlock()
}

func TestMempoolAnalyzer_AnalyzeMempool(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Add multiple transactions with different gas prices
	transactions := []*types.Transaction{
		createTestTransaction(big.NewInt(10000000000), 21000, 1), // 10 gwei
		createTestTransaction(big.NewInt(20000000000), 21000, 2), // 20 gwei
		createTestTransaction(big.NewInt(30000000000), 21000, 3), // 30 gwei
		createTestTransaction(big.NewInt(40000000000), 21000, 4), // 40 gwei
		createTestTransaction(big.NewInt(50000000000), 21000, 5), // 50 gwei
	}

	for _, tx := range transactions {
		err = analyzer.AddTransaction(tx)
		require.NoError(t, err)
	}

	// Analyze mempool
	analysis, err := analyzer.AnalyzeMempool(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)

	// Validate analysis result
	assert.Equal(t, len(transactions), analysis.TotalTransactions)
	assert.Equal(t, len(transactions), analysis.PendingTransactions)
	assert.NotNil(t, analysis.GasStatistics)
	assert.Contains(t, []string{"low", "medium", "high"}, analysis.CongestionLevel)
	assert.True(t, analysis.CongestionScore.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, analysis.OptimalGasPrice.GreaterThan(decimal.Zero))
	assert.True(t, analysis.EstimatedWaitTime > 0)
	assert.NotNil(t, analysis.GasPredictions)
	assert.NotNil(t, analysis.Recommendations)
	assert.NotNil(t, analysis.TopTransactions)
}

func TestMempoolAnalyzer_GetMetrics(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Get metrics
	metrics := analyzer.GetMetrics()
	assert.NotNil(t, metrics)

	// Validate metrics
	assert.Contains(t, metrics, "total_transactions")
	assert.Contains(t, metrics, "is_running")
	assert.Contains(t, metrics, "gas_tracker_enabled")
	assert.Contains(t, metrics, "congestion_model_enabled")
	assert.Contains(t, metrics, "gas_predictor_enabled")
	assert.Contains(t, metrics, "time_estimator_enabled")
	assert.Contains(t, metrics, "priority_analyzer_enabled")

	assert.Equal(t, true, metrics["is_running"])
	assert.Equal(t, true, metrics["gas_tracker_enabled"])
	assert.Equal(t, true, metrics["congestion_model_enabled"])
	assert.Equal(t, true, metrics["gas_predictor_enabled"])
	assert.Equal(t, true, metrics["time_estimator_enabled"])
	assert.Equal(t, true, metrics["priority_analyzer_enabled"])
}

func TestGasTracker_CalculateStatistics(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)

	// Create test transactions
	transactions := []*MempoolTransaction{
		{GasPrice: decimal.NewFromInt(10000000000)}, // 10 gwei
		{GasPrice: decimal.NewFromInt(20000000000)}, // 20 gwei
		{GasPrice: decimal.NewFromInt(30000000000)}, // 30 gwei
		{GasPrice: decimal.NewFromInt(40000000000)}, // 40 gwei
		{GasPrice: decimal.NewFromInt(50000000000)}, // 50 gwei
	}

	// Calculate statistics
	stats := analyzer.gasTracker.CalculateStatistics(transactions)
	assert.NotNil(t, stats)

	// Check mean (should be 30 gwei)
	expectedMean := decimal.NewFromInt(30000000000)
	assert.True(t, stats.Mean.Equal(expectedMean))

	// Check median (should be 30 gwei)
	assert.True(t, stats.Median.Equal(expectedMean))

	// Check percentiles
	assert.NotEmpty(t, stats.Percentiles)
	assert.Contains(t, stats.Percentiles, 50) // Median should be in percentiles

	// Check that statistics were updated
	assert.False(t, stats.LastUpdated.IsZero())
}

func TestCongestionModel_AnalyzeCongestion(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)

	// Test low congestion (few transactions, low gas prices)
	lowCongestionTxs := []*MempoolTransaction{
		{GasPrice: decimal.NewFromInt(10000000000)}, // 10 gwei
		{GasPrice: decimal.NewFromInt(15000000000)}, // 15 gwei
	}

	level, score := analyzer.congestionModel.AnalyzeCongestion(lowCongestionTxs)
	assert.Equal(t, "low", level)
	assert.True(t, score.LessThan(decimal.NewFromFloat(0.5)))

	// Test high congestion (many transactions, high gas prices)
	highCongestionTxs := make([]*MempoolTransaction, 1500) // > 1000 transactions
	for i := range highCongestionTxs {
		highCongestionTxs[i] = &MempoolTransaction{
			GasPrice: decimal.NewFromInt(100000000000), // 100 gwei
		}
	}

	level, score = analyzer.congestionModel.AnalyzeCongestion(highCongestionTxs)
	assert.Equal(t, "high", level)
	assert.True(t, score.GreaterThan(decimal.NewFromFloat(0.7)))
}

func TestTimeEstimator_EstimateConfirmationTime(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)

	// Test different gas prices
	testCases := []struct {
		gasPrice     decimal.Decimal
		expectedTime time.Duration
	}{
		{decimal.NewFromInt(100000000000), 1 * time.Minute},  // 100 gwei -> fast
		{decimal.NewFromInt(30000000000), 3 * time.Minute},   // 30 gwei -> medium
		{decimal.NewFromInt(15000000000), 5 * time.Minute},   // 15 gwei -> slow
		{decimal.NewFromInt(5000000000), 10 * time.Minute},   // 5 gwei -> very slow
	}

	for _, tc := range testCases {
		estimatedTime := analyzer.timeEstimator.EstimateConfirmationTime(tc.gasPrice)
		assert.Equal(t, tc.expectedTime, estimatedTime)
	}
}

func TestPriorityAnalyzer_CalculatePriority(t *testing.T) {
	logger := createTestLoggerForMempool()
	config := GetDefaultMempoolAnalyzerConfig()

	analyzer := NewMempoolAnalyzer(logger, config)
	ctx := context.Background()

	// Start the analyzer to initialize weights
	err := analyzer.Start(ctx)
	require.NoError(t, err)
	defer analyzer.Stop()

	// Create test transaction
	tx := &MempoolTransaction{
		GasPrice:         decimal.NewFromInt(50000000000), // 50 gwei
		GasTipCap:        decimal.NewFromInt(2000000000),  // 2 gwei
		Size:             250,                             // 250 bytes
		FirstSeen:        time.Now().Add(-5 * time.Minute), // 5 minutes old
		IsReplacement:    true,
		ReplacementCount: 2,
	}

	priority := analyzer.priorityAnalyzer.CalculatePriority(tx)
	assert.True(t, priority.GreaterThan(decimal.Zero))
}

func TestGetDefaultMempoolAnalyzerConfig(t *testing.T) {
	config := GetDefaultMempoolAnalyzerConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 30*time.Second, config.UpdateInterval)
	assert.Equal(t, 1*time.Hour, config.DataRetentionPeriod)
	assert.Equal(t, 10000, config.MaxTransactions)

	// Check gas tracker config
	assert.True(t, config.GasTrackerConfig.Enabled)
	assert.NotEmpty(t, config.GasTrackerConfig.PercentileTargets)

	// Check congestion model config
	assert.True(t, config.CongestionModelConfig.Enabled)
	assert.NotEmpty(t, config.CongestionModelConfig.CongestionThresholds)

	// Check gas predictor config
	assert.True(t, config.GasPredictorConfig.Enabled)
	assert.NotEmpty(t, config.GasPredictorConfig.PredictionMethods)

	// Check time estimator config
	assert.True(t, config.TimeEstimatorConfig.Enabled)
	assert.NotEmpty(t, config.TimeEstimatorConfig.EstimationMethod)

	// Check priority analyzer config
	assert.True(t, config.PriorityAnalyzerConfig.Enabled)
	assert.NotEmpty(t, config.PriorityAnalyzerConfig.PriorityFactors)
}

func TestValidateMempoolAnalyzerConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultMempoolAnalyzerConfig()
	err := ValidateMempoolAnalyzerConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultMempoolAnalyzerConfig()
	disabledConfig.Enabled = false
	err = ValidateMempoolAnalyzerConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []MempoolAnalyzerConfig{
		// Invalid update interval
		{
			Enabled:        true,
			UpdateInterval: 0,
		},
		// Invalid data retention period
		{
			Enabled:             true,
			UpdateInterval:      30 * time.Second,
			DataRetentionPeriod: 0,
		},
		// Invalid max transactions
		{
			Enabled:             true,
			UpdateInterval:      30 * time.Second,
			DataRetentionPeriod: 1 * time.Hour,
			MaxTransactions:     0,
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateMempoolAnalyzerConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestConfigVariants(t *testing.T) {
	// Test high frequency config
	hfConfig := GetHighFrequencyConfig()
	assert.True(t, hfConfig.UpdateInterval < GetDefaultMempoolAnalyzerConfig().UpdateInterval)
	assert.True(t, hfConfig.MaxTransactions > GetDefaultMempoolAnalyzerConfig().MaxTransactions)

	// Test low latency config
	llConfig := GetLowLatencyConfig()
	assert.True(t, llConfig.UpdateInterval < GetDefaultMempoolAnalyzerConfig().UpdateInterval)
	assert.True(t, llConfig.MaxTransactions < GetDefaultMempoolAnalyzerConfig().MaxTransactions)

	// Test analytics config
	analyticsConfig := GetAnalyticsConfig()
	assert.True(t, analyticsConfig.DataRetentionPeriod > GetDefaultMempoolAnalyzerConfig().DataRetentionPeriod)
	assert.True(t, analyticsConfig.MaxTransactions > GetDefaultMempoolAnalyzerConfig().MaxTransactions)

	// Validate all configs
	assert.NoError(t, ValidateMempoolAnalyzerConfig(hfConfig))
	assert.NoError(t, ValidateMempoolAnalyzerConfig(llConfig))
	assert.NoError(t, ValidateMempoolAnalyzerConfig(analyticsConfig))
}

func TestUtilityFunctions(t *testing.T) {
	// Test supported prediction methods
	methods := GetSupportedPredictionMethods()
	assert.NotEmpty(t, methods)
	assert.Contains(t, methods, "moving_average")
	assert.Contains(t, methods, "neural_network")

	// Test supported estimation methods
	estimationMethods := GetSupportedEstimationMethods()
	assert.NotEmpty(t, estimationMethods)
	assert.Contains(t, estimationMethods, "historical_analysis")

	// Test supported weighting methods
	weightingMethods := GetSupportedWeightingMethods()
	assert.NotEmpty(t, weightingMethods)
	assert.Contains(t, weightingMethods, "dynamic")

	// Test supported priority factors
	priorityFactors := GetSupportedPriorityFactors()
	assert.NotEmpty(t, priorityFactors)
	assert.Contains(t, priorityFactors, "gas_price")

	// Test congestion level descriptions
	congestionDescriptions := GetCongestionLevelDescription()
	assert.NotEmpty(t, congestionDescriptions)
	assert.Contains(t, congestionDescriptions, "low")
	assert.Contains(t, congestionDescriptions, "high")

	// Test optimal config for use case
	config, err := GetOptimalConfigForUseCase("high_frequency_trading")
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Test invalid use case
	_, err = GetOptimalConfigForUseCase("invalid_use_case")
	assert.Error(t, err)
}
