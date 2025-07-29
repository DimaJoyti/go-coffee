package monitoring

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLoggerForMonitoring() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

// Helper function to create test transaction
func createTestTransactionForMonitoring(gasPrice *big.Int, gasLimit uint64, nonce uint64) *types.Transaction {
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	value := big.NewInt(1000000000000000000) // 1 ETH
	data := []byte{}

	return types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)
}

func TestNewTransactionMonitor(t *testing.T) {
	logger := createTestLoggerForMonitoring()
	config := GetDefaultTransactionMonitorConfig()

	monitor := NewTransactionMonitor(logger, config)

	assert.NotNil(t, monitor)
	assert.Equal(t, config.Enabled, monitor.config.Enabled)
	assert.Equal(t, config.UpdateInterval, monitor.config.UpdateInterval)
	assert.False(t, monitor.IsRunning())
	assert.NotNil(t, monitor.confirmationTracker)
	assert.NotNil(t, monitor.failureDetector)
	assert.NotNil(t, monitor.retryManager)
	assert.NotNil(t, monitor.alertManager)
}

func TestTransactionMonitor_StartStop(t *testing.T) {
	logger := createTestLoggerForMonitoring()
	config := GetDefaultTransactionMonitorConfig()

	monitor := NewTransactionMonitor(logger, config)
	ctx := context.Background()

	err := monitor.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, monitor.IsRunning())

	err = monitor.Stop()
	assert.NoError(t, err)
	assert.False(t, monitor.IsRunning())
}

func TestTransactionMonitor_StartDisabled(t *testing.T) {
	logger := createTestLoggerForMonitoring()
	config := GetDefaultTransactionMonitorConfig()
	config.Enabled = false

	monitor := NewTransactionMonitor(logger, config)
	ctx := context.Background()

	err := monitor.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, monitor.IsRunning()) // Should remain false when disabled
}

func TestTransactionMonitor_TrackTransaction(t *testing.T) {
	logger := createTestLoggerForMonitoring()
	config := GetDefaultTransactionMonitorConfig()

	monitor := NewTransactionMonitor(logger, config)
	ctx := context.Background()

	// Start the monitor
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop()

	// Create test transaction
	tx := createTestTransactionForMonitoring(big.NewInt(20000000000), 21000, 1) // 20 gwei

	// Track transaction
	metadata := map[string]interface{}{
		"user_id": "test_user",
		"purpose": "test_transfer",
	}
	err = monitor.TrackTransaction(tx, metadata)
	assert.NoError(t, err)

	// Check that transaction is being tracked
	trackedTx, err := monitor.GetTransactionStatus(tx.Hash())
	assert.NoError(t, err)
	assert.NotNil(t, trackedTx)
	assert.Equal(t, tx.Hash(), trackedTx.Hash)
	assert.Equal(t, StatusPending, trackedTx.Status)
	assert.Equal(t, metadata["user_id"], trackedTx.Metadata["user_id"])
}

func TestTransactionMonitor_TrackDuplicateTransaction(t *testing.T) {
	logger := createTestLoggerForMonitoring()
	config := GetDefaultTransactionMonitorConfig()

	monitor := NewTransactionMonitor(logger, config)
	ctx := context.Background()

	// Start the monitor
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop()

	// Create test transaction
	tx := createTestTransactionForMonitoring(big.NewInt(20000000000), 21000, 1)

	// Track transaction first time
	err = monitor.TrackTransaction(tx, nil)
	assert.NoError(t, err)

	// Try to track same transaction again
	err = monitor.TrackTransaction(tx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already being tracked")
}

func TestTransactionMonitor_StopTracking(t *testing.T) {
	logger := createTestLoggerForMonitoring()
	config := GetDefaultTransactionMonitorConfig()

	monitor := NewTransactionMonitor(logger, config)
	ctx := context.Background()

	// Start the monitor
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop()

	// Create and track transaction
	tx := createTestTransactionForMonitoring(big.NewInt(20000000000), 21000, 1)
	err = monitor.TrackTransaction(tx, nil)
	require.NoError(t, err)

	// Stop tracking
	err = monitor.StopTracking(tx.Hash())
	assert.NoError(t, err)

	// Verify transaction is no longer tracked
	_, err = monitor.GetTransactionStatus(tx.Hash())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not being tracked")
}

func TestTransactionMonitor_GetMonitoringResult(t *testing.T) {
	logger := createTestLoggerForMonitoring()
	config := GetDefaultTransactionMonitorConfig()

	monitor := NewTransactionMonitor(logger, config)
	ctx := context.Background()

	// Start the monitor
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop()

	// Track multiple transactions
	transactions := []*types.Transaction{
		createTestTransactionForMonitoring(big.NewInt(20000000000), 21000, 1),
		createTestTransactionForMonitoring(big.NewInt(25000000000), 21000, 2),
		createTestTransactionForMonitoring(big.NewInt(30000000000), 21000, 3),
	}

	for _, tx := range transactions {
		err = monitor.TrackTransaction(tx, nil)
		require.NoError(t, err)
	}

	// Get monitoring result
	result := monitor.GetMonitoringResult()
	assert.NotNil(t, result)
	assert.Equal(t, len(transactions), result.TotalTracked)
	assert.Equal(t, len(transactions), result.PendingCount) // All should be pending initially
	assert.NotNil(t, result.PerformanceMetrics)
}

func TestTransactionMonitor_GetMetrics(t *testing.T) {
	logger := createTestLoggerForMonitoring()
	config := GetDefaultTransactionMonitorConfig()

	monitor := NewTransactionMonitor(logger, config)
	ctx := context.Background()

	// Start the monitor
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop()

	metrics := monitor.GetMetrics()
	assert.NotNil(t, metrics)

	// Validate metrics structure
	assert.Contains(t, metrics, "is_running")
	assert.Contains(t, metrics, "total_tracked")
	assert.Contains(t, metrics, "pending_count")
	assert.Contains(t, metrics, "confirmed_count")
	assert.Contains(t, metrics, "failed_count")
	assert.Contains(t, metrics, "confirmation_enabled")
	assert.Contains(t, metrics, "failure_detection_enabled")
	assert.Contains(t, metrics, "retry_enabled")
	assert.Contains(t, metrics, "alert_enabled")

	assert.Equal(t, true, metrics["is_running"])
	assert.Equal(t, true, metrics["confirmation_enabled"])
	assert.Equal(t, true, metrics["failure_detection_enabled"])
	assert.Equal(t, true, metrics["retry_enabled"])
	assert.Equal(t, true, metrics["alert_enabled"])
}

func TestMockConfirmationTracker(t *testing.T) {
	tracker := &MockConfirmationTracker{}
	ctx := context.Background()

	// Create test tracked transaction
	tx := &TrackedTransaction{
		Hash:        common.HexToHash("0x123"),
		Status:      StatusPending,
		SubmittedAt: time.Now(),
	}

	// Test confirmation tracking
	err := tracker.TrackConfirmations(ctx, tx)
	assert.NoError(t, err)
	assert.Equal(t, StatusConfirming, tx.Status)
	assert.Equal(t, 1, tx.Confirmations)

	// Track again to increment confirmations
	err = tracker.TrackConfirmations(ctx, tx)
	assert.NoError(t, err)
	assert.Equal(t, 2, tx.Confirmations)

	// Track until confirmed
	err = tracker.TrackConfirmations(ctx, tx)
	assert.NoError(t, err)
	assert.Equal(t, StatusConfirmed, tx.Status)
	assert.Equal(t, 3, tx.Confirmations)
	assert.NotNil(t, tx.ConfirmedAt)
}

func TestMockFailureDetector(t *testing.T) {
	detector := &MockFailureDetector{}
	ctx := context.Background()

	// Test recent transaction (should not fail)
	recentTx := &TrackedTransaction{
		Hash:        common.HexToHash("0x123"),
		Status:      StatusPending,
		SubmittedAt: time.Now(),
	}

	analysis, err := detector.DetectFailures(ctx, recentTx)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.False(t, analysis.IsFailed)

	// Test old transaction (should fail)
	oldTx := &TrackedTransaction{
		Hash:        common.HexToHash("0x456"),
		Status:      StatusPending,
		SubmittedAt: time.Now().Add(-2 * time.Hour),
	}

	analysis, err = detector.DetectFailures(ctx, oldTx)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.True(t, analysis.IsFailed)
	assert.True(t, analysis.IsRetryable)
	assert.NotEmpty(t, analysis.FailureReason)
}

func TestMockRetryManager(t *testing.T) {
	manager := &MockRetryManager{}

	// Test retry decision
	tx := &TrackedTransaction{
		Hash:          common.HexToHash("0x123"),
		RetryAttempts: 1,
	}

	failure := &FailureAnalysis{
		IsFailed:    true,
		IsRetryable: true,
	}

	shouldRetry := manager.ShouldRetry(tx, failure)
	assert.True(t, shouldRetry)

	// Test retry scheduling
	err := manager.ScheduleRetry(tx)
	assert.NoError(t, err)
	assert.Equal(t, 2, tx.RetryAttempts)
	assert.NotNil(t, tx.LastRetryAt)
	assert.Equal(t, StatusPending, tx.Status)
}

func TestMockAlertManager(t *testing.T) {
	manager := &MockAlertManager{}

	tx := &TrackedTransaction{
		Hash: common.HexToHash("0x123"),
	}

	// Test alert creation
	alert := manager.CreateAlert(tx, "test_alert", "warning", "Test alert message")
	assert.NotNil(t, alert)
	assert.Equal(t, tx.Hash, alert.TransactionHash)
	assert.Equal(t, "test_alert", alert.Type)
	assert.Equal(t, "warning", alert.Severity)
	assert.Equal(t, "Test alert message", alert.Message)
	assert.False(t, alert.Acknowledged)
	assert.True(t, alert.ActionRequired)

	// Test alert sending
	err := manager.SendAlert(alert)
	assert.NoError(t, err)
}

func TestGetDefaultTransactionMonitorConfig(t *testing.T) {
	config := GetDefaultTransactionMonitorConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 10*time.Second, config.UpdateInterval)
	assert.Equal(t, 1000, config.MaxTrackedTransactions)
	assert.Equal(t, 24*time.Hour, config.HistoryRetentionPeriod)

	// Check confirmation config
	assert.True(t, config.ConfirmationConfig.Enabled)
	assert.Equal(t, 3, config.ConfirmationConfig.RequiredConfirmations)

	// Check failure config
	assert.True(t, config.FailureConfig.Enabled)
	assert.NotEmpty(t, config.FailureConfig.DetectionMethods)

	// Check retry config
	assert.True(t, config.RetryConfig.Enabled)
	assert.Equal(t, 3, config.RetryConfig.MaxRetryAttempts)

	// Check alert config
	assert.True(t, config.AlertConfig.Enabled)
	assert.NotEmpty(t, config.AlertConfig.AlertChannels)
}

func TestValidateTransactionMonitorConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultTransactionMonitorConfig()
	err := ValidateTransactionMonitorConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultTransactionMonitorConfig()
	disabledConfig.Enabled = false
	err = ValidateTransactionMonitorConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid configs
	invalidConfigs := []TransactionMonitorConfig{
		// Invalid update interval
		{
			Enabled:        true,
			UpdateInterval: 0,
		},
		// Invalid max tracked transactions
		{
			Enabled:                true,
			UpdateInterval:         10 * time.Second,
			MaxTrackedTransactions: 0,
		},
		// Invalid history retention period
		{
			Enabled:                true,
			UpdateInterval:         10 * time.Second,
			MaxTrackedTransactions: 1000,
			HistoryRetentionPeriod: 0,
		},
	}

	for i, config := range invalidConfigs {
		err := ValidateTransactionMonitorConfig(config)
		assert.Error(t, err, "Config %d should be invalid", i)
	}
}

func TestConfigVariants(t *testing.T) {
	// Test high frequency config
	hfConfig := GetHighFrequencyConfig()
	assert.True(t, hfConfig.UpdateInterval < GetDefaultTransactionMonitorConfig().UpdateInterval)
	assert.True(t, hfConfig.MaxTrackedTransactions > GetDefaultTransactionMonitorConfig().MaxTrackedTransactions)

	// Test low latency config
	llConfig := GetLowLatencyConfig()
	assert.True(t, llConfig.UpdateInterval < GetDefaultTransactionMonitorConfig().UpdateInterval)
	assert.Equal(t, 1, llConfig.ConfirmationConfig.RequiredConfirmations)

	// Test robust config
	robustConfig := GetRobustConfig()
	assert.True(t, robustConfig.UpdateInterval > GetDefaultTransactionMonitorConfig().UpdateInterval)
	assert.True(t, robustConfig.ConfirmationConfig.RequiredConfirmations > GetDefaultTransactionMonitorConfig().ConfirmationConfig.RequiredConfirmations)

	// Validate all configs
	assert.NoError(t, ValidateTransactionMonitorConfig(hfConfig))
	assert.NoError(t, ValidateTransactionMonitorConfig(llConfig))
	assert.NoError(t, ValidateTransactionMonitorConfig(robustConfig))
}

func TestUtilityFunctions(t *testing.T) {
	// Test supported failure detection methods
	methods := GetSupportedFailureDetectionMethods()
	assert.NotEmpty(t, methods)
	assert.Contains(t, methods, "timeout")
	assert.Contains(t, methods, "gas_limit")

	// Test supported retry strategies
	strategies := GetSupportedRetryStrategies()
	assert.NotEmpty(t, strategies)
	assert.Contains(t, strategies, "increase_gas_price")

	// Test supported alert channels
	channels := GetSupportedAlertChannels()
	assert.NotEmpty(t, channels)
	assert.Contains(t, channels, "log")
	assert.Contains(t, channels, "webhook")

	// Test optimal config for use case
	config, err := GetOptimalConfigForUseCase("high_frequency")
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Test invalid use case
	_, err = GetOptimalConfigForUseCase("invalid_use_case")
	assert.Error(t, err)

	// Test status descriptions
	statusDescriptions := GetTransactionStatusDescription()
	assert.NotEmpty(t, statusDescriptions)
	assert.Contains(t, statusDescriptions, StatusPending)
	assert.Contains(t, statusDescriptions, StatusConfirmed)

	// Test failure type descriptions
	failureDescriptions := GetFailureTypeDescription()
	assert.NotEmpty(t, failureDescriptions)
	assert.Contains(t, failureDescriptions, "timeout")
	assert.Contains(t, failureDescriptions, "gas_limit")
}
