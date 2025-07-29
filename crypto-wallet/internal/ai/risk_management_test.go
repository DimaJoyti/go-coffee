package ai

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAIRedisClient for testing
type MockAIRedisClient struct {
	mock.Mock
}

func (m *MockAIRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockAIRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockAIRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockAIRedisClient) Exists(ctx context.Context, keys ...string) (bool, error) {
	args := m.Called(ctx, keys)
	return args.Bool(0), args.Error(1)
}

func (m *MockAIRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockAIRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAIRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAIRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockAIRedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockAIRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockAIRedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func (m *MockAIRedisClient) Pipeline() redis.Pipeline {
	m.Called()
	return nil // For testing, we can return nil
}

func (m *MockAIRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Helper function to create test logger config for risk management tests
func getTestLoggerConfigForRisk() config.LoggingConfig {
	return config.LoggingConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
	}
}

func TestNewAIRiskManager(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)

	assert.NotNil(t, manager)
	assert.Equal(t, config.RiskToleranceLevel, manager.config.RiskToleranceLevel)
	assert.Equal(t, config.MaxPortfolioRisk, manager.config.MaxPortfolioRisk)
	assert.False(t, manager.isRunning)
	assert.NotNil(t, manager.riskAssessments)
	assert.NotNil(t, manager.riskAlerts)
}

func TestAIRiskManager_Start(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()
	config.RiskAssessmentInterval = 100 * time.Millisecond // Fast for testing

	manager := NewAIRiskManager(logger, mockCache, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, manager.IsRunning())

	// Clean up
	err = manager.Stop()
	assert.NoError(t, err)
	assert.False(t, manager.IsRunning())
}

func TestAIRiskManager_StartDisabled(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()
	config.Enabled = false

	manager := NewAIRiskManager(logger, mockCache, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, manager.IsRunning()) // Should remain false when disabled
}

func TestAIRiskManager_AssessTransactionRisk(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)
	ctx := context.Background()

	// Start the manager
	err := manager.Start(ctx)
	require.NoError(t, err)

	// Create test transaction data
	transactionData := &TransactionData{
		Hash:           "0x1234567890abcdef",
		From:           "0xfrom",
		To:             "0xto",
		Value:          decimal.NewFromFloat(1000),
		GasPrice:       decimal.NewFromFloat(20000000000), // 20 gwei
		GasLimit:       21000,
		Timestamp:      time.Now(),
		BlockNumber:    12345678,
		Success:        true,
		MEVDetected:    false,
		SlippageActual: decimal.NewFromFloat(0.01),
	}

	// Assess transaction risk
	assessment, err := manager.AssessTransactionRisk(ctx, transactionData)
	assert.NoError(t, err)
	assert.NotNil(t, assessment)
	assert.Equal(t, "transaction", assessment.AssessmentType)
	assert.Equal(t, transactionData.Hash, assessment.TransactionHash)
	assert.NotEmpty(t, assessment.RiskFactors)
	assert.NotEmpty(t, assessment.Recommendations)

	// Clean up
	manager.Stop()
}

func TestAIRiskManager_AssessTransactionRisk_HighRisk(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	require.NoError(t, err)

	// Create high-risk transaction data
	transactionData := &TransactionData{
		Hash:            "0xhighrisk",
		From:            "0xfrom",
		To:              "0xto",
		Value:           decimal.NewFromFloat(100000),       // High value
		GasPrice:        decimal.NewFromFloat(100000000000), // 100 gwei - high gas
		GasLimit:        21000,
		ContractAddress: "0xcontract", // Contract interaction
		Timestamp:       time.Now(),
		BlockNumber:     12345678,
		Success:         true,
		MEVDetected:     true, // MEV detected
		SlippageActual:  decimal.NewFromFloat(0.05),
	}

	assessment, err := manager.AssessTransactionRisk(ctx, transactionData)
	assert.NoError(t, err)
	assert.NotNil(t, assessment)
	assert.True(t, assessment.OverallRiskScore.GreaterThan(decimal.NewFromFloat(0.5)))

	// Should generate alert for high risk
	alerts := manager.GetActiveAlerts()
	assert.NotEmpty(t, alerts)

	manager.Stop()
}

func TestAIRiskManager_AssessPortfolioRisk(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	require.NoError(t, err)

	// Create test portfolio data
	portfolioData := &PortfolioData{
		TotalValue: decimal.NewFromFloat(100000),
		Assets: map[string]*AssetHolding{
			"BTC": {
				Address:     "0xbtc",
				Symbol:      "BTC",
				Amount:      decimal.NewFromFloat(2),
				Value:       decimal.NewFromFloat(60000),
				Weight:      decimal.NewFromFloat(0.6),
				Price:       decimal.NewFromFloat(30000),
				PriceChange: decimal.NewFromFloat(0.02),
			},
			"ETH": {
				Address:     "0xeth",
				Symbol:      "ETH",
				Amount:      decimal.NewFromFloat(20),
				Value:       decimal.NewFromFloat(40000),
				Weight:      decimal.NewFromFloat(0.4),
				Price:       decimal.NewFromFloat(2000),
				PriceChange: decimal.NewFromFloat(0.01),
			},
		},
		Timestamp: time.Now(),
	}

	metrics, err := manager.AssessPortfolioRisk(ctx, portfolioData)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, portfolioData.TotalValue, metrics.TotalValue)
	assert.True(t, metrics.ValueAtRisk.GreaterThan(decimal.Zero))
	assert.True(t, metrics.Volatility.GreaterThan(decimal.Zero))

	manager.Stop()
}

func TestAIRiskManager_PredictMarketRisk(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	require.NoError(t, err)

	// Predict market risk for 1 hour
	predictions, err := manager.PredictMarketRisk(ctx, 1*time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, predictions)

	// Check prediction structure
	for _, prediction := range predictions {
		assert.NotEmpty(t, prediction.Scenario)
		assert.True(t, prediction.Probability.GreaterThanOrEqual(decimal.Zero))
		assert.True(t, prediction.Probability.LessThanOrEqual(decimal.NewFromFloat(1.0)))
		assert.NotEmpty(t, prediction.Description)
	}

	manager.Stop()
}

func TestAIRiskManager_OptimizePortfolio(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	require.NoError(t, err)

	portfolioData := &PortfolioData{
		TotalValue: decimal.NewFromFloat(100000),
		Assets: map[string]*AssetHolding{
			"BTC": {
				Symbol: "BTC",
				Weight: decimal.NewFromFloat(0.8), // Overweight
			},
			"ETH": {
				Symbol: "ETH",
				Weight: decimal.NewFromFloat(0.2),
			},
		},
		Timestamp: time.Now(),
	}

	constraints := &OptimizationConstraints{
		MaxWeight: decimal.NewFromFloat(0.5),
		MinWeight: decimal.NewFromFloat(0.1),
		MaxRisk:   decimal.NewFromFloat(0.2),
		MinReturn: decimal.NewFromFloat(0.05),
	}

	optimization, err := manager.OptimizePortfolio(ctx, portfolioData, constraints)
	assert.NoError(t, err)
	assert.NotNil(t, optimization)
	assert.NotEmpty(t, optimization.OptimalWeights)
	assert.True(t, optimization.ExpectedReturn.GreaterThan(decimal.Zero))
	assert.True(t, optimization.ExpectedRisk.GreaterThan(decimal.Zero))

	manager.Stop()
}

func TestAIRiskManager_GetOptimalExecutionStrategy(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	require.NoError(t, err)

	tradeRequest := &TradeRequest{
		Asset:       "BTC",
		Side:        "buy",
		Amount:      decimal.NewFromFloat(10000),
		MaxSlippage: decimal.NewFromFloat(0.01),
		Urgency:     "medium",
		Timeframe:   30 * time.Minute,
	}

	strategy, err := manager.GetOptimalExecutionStrategy(ctx, tradeRequest)
	assert.NoError(t, err)
	assert.NotNil(t, strategy)
	assert.NotEmpty(t, strategy.Strategy)
	assert.NotEmpty(t, strategy.Chunks)
	assert.True(t, strategy.EstimatedCost.GreaterThan(decimal.Zero))
	assert.True(t, strategy.EstimatedTime > 0)

	manager.Stop()
}

func TestAIRiskManager_AlertManagement(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)

	// Manually create an alert
	alert := &RiskAlert{
		ID:          "test-alert-1",
		Type:        "portfolio",
		Severity:    "high",
		Title:       "Test Alert",
		Description: "Test alert description",
		RiskScore:   decimal.NewFromFloat(0.8),
		Threshold:   decimal.NewFromFloat(0.7),
		Actions:     []string{"Test action"},
		CreatedAt:   time.Now(),
		Status:      "active",
	}

	manager.mutex.Lock()
	manager.riskAlerts[alert.ID] = alert
	manager.mutex.Unlock()

	// Test getting active alerts
	activeAlerts := manager.GetActiveAlerts()
	assert.Equal(t, 1, len(activeAlerts))
	assert.Equal(t, alert.ID, activeAlerts[0].ID)

	// Test resolving alert
	err := manager.ResolveAlert(alert.ID)
	assert.NoError(t, err)

	// Check alert status
	allAlerts := manager.GetAllAlerts()
	assert.Equal(t, 1, len(allAlerts))
	assert.Equal(t, "resolved", allAlerts[0].Status)
	assert.NotNil(t, allAlerts[0].ResolvedAt)

	// Test dismissing alert (create new one)
	alert2 := &RiskAlert{
		ID:        "test-alert-2",
		Type:      "transaction",
		Severity:  "medium",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	manager.mutex.Lock()
	manager.riskAlerts[alert2.ID] = alert2
	manager.mutex.Unlock()

	err = manager.DismissAlert(alert2.ID)
	assert.NoError(t, err)

	allAlerts = manager.GetAllAlerts()
	found := false
	for _, a := range allAlerts {
		if a.ID == alert2.ID {
			assert.Equal(t, "dismissed", a.Status)
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestAIRiskManager_ConfigManagement(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)

	// Test getting config
	currentConfig := manager.GetConfig()
	assert.Equal(t, config.RiskToleranceLevel, currentConfig.RiskToleranceLevel)

	// Test updating config
	newConfig := config
	newConfig.RiskToleranceLevel = "aggressive"
	newConfig.MaxPortfolioRisk = decimal.NewFromFloat(0.25)

	err := manager.UpdateConfig(newConfig)
	assert.NoError(t, err)

	updatedConfig := manager.GetConfig()
	assert.Equal(t, "aggressive", updatedConfig.RiskToleranceLevel)
	assert.Equal(t, decimal.NewFromFloat(0.25), updatedConfig.MaxPortfolioRisk)
}

func TestGetDefaultAIRiskConfig(t *testing.T) {
	config := GetDefaultAIRiskConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, "moderate", config.RiskToleranceLevel)
	assert.Equal(t, decimal.NewFromFloat(0.15), config.MaxPortfolioRisk)
	assert.Equal(t, decimal.NewFromFloat(0.05), config.MaxSingleTransactionRisk)
	assert.True(t, config.EnableRealTimeMonitoring)
	assert.True(t, config.EnablePredictiveAnalysis)
	assert.Equal(t, 5*time.Minute, config.RiskAssessmentInterval)
	assert.Equal(t, 1*time.Minute, config.MarketAnalysisInterval)
	assert.Equal(t, 1*time.Hour, config.ModelUpdateInterval)
	assert.Equal(t, 30, config.HistoricalDataDays)
}

// Benchmark tests
func BenchmarkAIRiskManager_AssessTransactionRisk(b *testing.B) {
	logger := logger.NewLogger(getTestLoggerConfigForRisk())
	mockCache := &MockAIRedisClient{}
	config := GetDefaultAIRiskConfig()

	manager := NewAIRiskManager(logger, mockCache, config)
	ctx := context.Background()

	err := manager.Start(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer manager.Stop()

	transactionData := &TransactionData{
		Hash:           "0xbenchmark",
		From:           "0xfrom",
		To:             "0xto",
		Value:          decimal.NewFromFloat(1000),
		GasPrice:       decimal.NewFromFloat(20000000000),
		GasLimit:       21000,
		Timestamp:      time.Now(),
		BlockNumber:    12345678,
		Success:        true,
		MEVDetected:    false,
		SlippageActual: decimal.NewFromFloat(0.01),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.AssessTransactionRisk(ctx, transactionData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
