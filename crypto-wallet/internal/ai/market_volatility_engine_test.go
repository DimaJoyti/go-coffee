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

// MockVolatilityRedisClient for testing
type MockVolatilityRedisClient struct {
	mock.Mock
}

func (m *MockVolatilityRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockVolatilityRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockVolatilityRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockVolatilityRedisClient) Exists(ctx context.Context, keys ...string) (bool, error) {
	args := m.Called(ctx, keys)
	return args.Bool(0), args.Error(1)
}

func (m *MockVolatilityRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockVolatilityRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockVolatilityRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockVolatilityRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockVolatilityRedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockVolatilityRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockVolatilityRedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func (m *MockVolatilityRedisClient) Pipeline() redis.Pipeline {
	m.Called()
	return nil // For testing, we can return nil
}

func (m *MockVolatilityRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Helper function to create test logger config
func getTestLoggerConfig() config.LoggingConfig {
	return config.LoggingConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
	}
}

func TestNewMarketVolatilityEngine(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()

	engine := NewMarketVolatilityEngine(logger, mockCache, config)

	assert.NotNil(t, engine)
	assert.Equal(t, config.AnalysisInterval, engine.config.AnalysisInterval)
	assert.Equal(t, config.MonitoredAssets, engine.config.MonitoredAssets)
	assert.False(t, engine.isRunning)
	assert.NotNil(t, engine.assetVolatilities)
	assert.NotNil(t, engine.correlationMatrix)
	assert.NotNil(t, engine.positionSizes)
}

func TestMarketVolatilityEngine_Start(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()
	config.AnalysisInterval = 100 * time.Millisecond // Fast for testing

	engine := NewMarketVolatilityEngine(logger, mockCache, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, engine.IsRunning())

	// Clean up
	err = engine.Stop()
	assert.NoError(t, err)
	assert.False(t, engine.IsRunning())
}

func TestMarketVolatilityEngine_StartDisabled(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()
	config.Enabled = false

	engine := NewMarketVolatilityEngine(logger, mockCache, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, engine.IsRunning()) // Should remain false when disabled
}

func TestMarketVolatilityEngine_AnalyzeAssetVolatility(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()

	engine := NewMarketVolatilityEngine(logger, mockCache, config)
	ctx := context.Background()

	// Start the engine
	err := engine.Start(ctx)
	require.NoError(t, err)

	// Create test price data
	priceData := []PriceMovement{
		{
			Timestamp: time.Now().Add(-3 * time.Hour),
			Price:     decimal.NewFromFloat(1000),
			Return:    decimal.Zero,
		},
		{
			Timestamp: time.Now().Add(-2 * time.Hour),
			Price:     decimal.NewFromFloat(1020),
			Return:    decimal.NewFromFloat(0.02),
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Price:     decimal.NewFromFloat(1010),
			Return:    decimal.NewFromFloat(-0.0098),
		},
		{
			Timestamp: time.Now(),
			Price:     decimal.NewFromFloat(1030),
			Return:    decimal.NewFromFloat(0.0198),
		},
	}

	// Analyze asset volatility
	volatility, err := engine.AnalyzeAssetVolatility(ctx, "BTC", priceData)
	assert.NoError(t, err)
	assert.NotNil(t, volatility)
	assert.Equal(t, "BTC", volatility.Asset)
	assert.True(t, volatility.CurrentVolatility.GreaterThan(decimal.Zero))
	assert.NotEmpty(t, volatility.VolatilityRegime)
	assert.NotEmpty(t, volatility.VolatilityTrend)

	// Clean up
	engine.Stop()
}

func TestMarketVolatilityEngine_CalculateCorrelationMatrix(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()
	config.MonitoredAssets = []string{"BTC", "ETH"}

	engine := NewMarketVolatilityEngine(logger, mockCache, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	require.NoError(t, err)

	// Add some volatility data first
	priceData := engine.generateMockPriceData("BTC", 10)
	_, err = engine.AnalyzeAssetVolatility(ctx, "BTC", priceData)
	require.NoError(t, err)

	priceData = engine.generateMockPriceData("ETH", 10)
	_, err = engine.AnalyzeAssetVolatility(ctx, "ETH", priceData)
	require.NoError(t, err)

	// Calculate correlation matrix
	correlations, err := engine.CalculateCorrelationMatrix(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, correlations)
	assert.Contains(t, correlations, "BTC")
	assert.Contains(t, correlations, "ETH")
	assert.Equal(t, decimal.NewFromFloat(1.0), correlations["BTC"]["BTC"])
	assert.Equal(t, decimal.NewFromFloat(1.0), correlations["ETH"]["ETH"])

	engine.Stop()
}

func TestMarketVolatilityEngine_GeneratePositionSizeRecommendation(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()

	engine := NewMarketVolatilityEngine(logger, mockCache, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	require.NoError(t, err)

	// Add volatility data first
	priceData := engine.generateMockPriceData("BTC", 10)
	_, err = engine.AnalyzeAssetVolatility(ctx, "BTC", priceData)
	require.NoError(t, err)

	// Generate position size recommendation
	portfolioValue := decimal.NewFromFloat(100000)
	recommendation, err := engine.GeneratePositionSizeRecommendation(ctx, "BTC", portfolioValue)
	assert.NoError(t, err)
	assert.NotNil(t, recommendation)
	assert.Equal(t, "BTC", recommendation.Asset)
	assert.True(t, recommendation.RecommendedSize.GreaterThan(decimal.Zero))
	assert.True(t, recommendation.MaxSafeSize.GreaterThan(decimal.Zero))
	assert.True(t, recommendation.Confidence.GreaterThan(decimal.Zero))
	assert.NotEmpty(t, recommendation.Reasoning)

	engine.Stop()
}

func TestMarketVolatilityEngine_VolatilityCalculations(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()

	engine := NewMarketVolatilityEngine(logger, mockCache, config)

	// Test return calculation
	priceData := []PriceMovement{
		{Price: decimal.NewFromFloat(100)},
		{Price: decimal.NewFromFloat(110)},
		{Price: decimal.NewFromFloat(105)},
		{Price: decimal.NewFromFloat(115)},
	}

	returns := engine.calculateReturns(priceData)
	assert.Equal(t, 3, len(returns))
	assert.True(t, returns[0].GreaterThan(decimal.Zero)) // 10% increase
	assert.True(t, returns[1].LessThan(decimal.Zero))    // ~4.5% decrease
	assert.True(t, returns[2].GreaterThan(decimal.Zero)) // ~9.5% increase

	// Test volatility calculation
	volatility := engine.calculateVolatility(returns)
	assert.True(t, volatility.GreaterThan(decimal.Zero))
}

func TestMarketVolatilityEngine_PositionSizingMethods(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()

	engine := NewMarketVolatilityEngine(logger, mockCache, config)

	// Test Kelly position size calculation
	portfolioValue := decimal.NewFromFloat(100000)
	kellySize := engine.calculateKellyPositionSize("BTC", portfolioValue)
	assert.True(t, kellySize.GreaterThan(decimal.Zero))
	assert.True(t, kellySize.LessThanOrEqual(portfolioValue.Mul(decimal.NewFromFloat(0.25)))) // Max 25%

	// Test volatility adjustment
	volatility := &AssetVolatility{
		VolatilityRegime: "high",
	}
	adjustment := engine.calculateVolatilityAdjustment(volatility)
	assert.True(t, adjustment.LessThan(decimal.NewFromFloat(1.0))) // Should reduce position for high volatility

	volatility.VolatilityRegime = "low"
	adjustment = engine.calculateVolatilityAdjustment(volatility)
	assert.True(t, adjustment.GreaterThan(decimal.NewFromFloat(1.0))) // Should increase position for low volatility
}

func TestMarketVolatilityEngine_AlertManagement(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()

	engine := NewMarketVolatilityEngine(logger, mockCache, config)

	// Manually create a volatility alert
	alert := &VolatilityAlert{
		ID:          "test-vol-alert-1",
		Asset:       "BTC",
		AlertType:   "high_volatility",
		Severity:    "high",
		Title:       "Test Volatility Alert",
		Description: "Test alert description",
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	engine.mutex.Lock()
	engine.volatilityAlerts[alert.ID] = alert
	engine.mutex.Unlock()

	// Test getting active alerts
	activeAlerts := engine.GetActiveVolatilityAlerts()
	assert.Equal(t, 1, len(activeAlerts))
	assert.Equal(t, alert.ID, activeAlerts[0].ID)

	// Test resolving alert
	err := engine.ResolveVolatilityAlert(alert.ID)
	assert.NoError(t, err)

	// Check alert status
	allAlerts := engine.GetVolatilityAlerts()
	assert.Equal(t, 1, len(allAlerts))
	assert.Equal(t, "resolved", allAlerts[0].Status)
	assert.NotNil(t, allAlerts[0].ResolvedAt)

	// Test dismissing alert (create new one)
	alert2 := &VolatilityAlert{
		ID:        "test-vol-alert-2",
		Asset:     "ETH",
		AlertType: "low_volatility",
		Severity:  "medium",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	engine.mutex.Lock()
	engine.volatilityAlerts[alert2.ID] = alert2
	engine.mutex.Unlock()

	err = engine.DismissVolatilityAlert(alert2.ID)
	assert.NoError(t, err)

	allAlerts = engine.GetVolatilityAlerts()
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

func TestMarketVolatilityEngine_ConfigManagement(t *testing.T) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()

	engine := NewMarketVolatilityEngine(logger, mockCache, config)

	// Test getting config
	currentConfig := engine.GetConfig()
	assert.Equal(t, config.AnalysisInterval, currentConfig.AnalysisInterval)

	// Test updating config
	newConfig := config
	newConfig.AnalysisInterval = 2 * time.Minute
	newConfig.MonitoredAssets = []string{"BTC", "ETH", "ADA"}

	err := engine.UpdateConfig(newConfig)
	assert.NoError(t, err)

	updatedConfig := engine.GetConfig()
	assert.Equal(t, 2*time.Minute, updatedConfig.AnalysisInterval)
	assert.Equal(t, []string{"BTC", "ETH", "ADA"}, updatedConfig.MonitoredAssets)
}

func TestGetDefaultVolatilityConfig(t *testing.T) {
	config := GetDefaultVolatilityConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 1*time.Minute, config.AnalysisInterval)
	assert.Equal(t, 5*time.Minute, config.CorrelationUpdateInterval)
	assert.Equal(t, 24*time.Hour, config.VolatilityWindow)
	assert.Equal(t, decimal.NewFromFloat(0.5), config.HighVolatilityThreshold)
	assert.Equal(t, decimal.NewFromFloat(0.1), config.LowVolatilityThreshold)
	assert.True(t, config.EnableRealTimeAlerts)
	assert.True(t, config.EnablePositionSizing)
	assert.NotEmpty(t, config.MonitoredAssets)
}

// Benchmark tests
func BenchmarkMarketVolatilityEngine_AnalyzeAssetVolatility(b *testing.B) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()

	engine := NewMarketVolatilityEngine(logger, mockCache, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer engine.Stop()

	priceData := engine.generateMockPriceData("BTC", 30)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.AnalyzeAssetVolatility(ctx, "BTC", priceData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarketVolatilityEngine_CalculateCorrelationMatrix(b *testing.B) {
	logger := logger.NewLogger(getTestLoggerConfig())
	mockCache := &MockVolatilityRedisClient{}
	config := GetDefaultVolatilityConfig()

	engine := NewMarketVolatilityEngine(logger, mockCache, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer engine.Stop()

	// Add some volatility data
	for _, asset := range config.MonitoredAssets {
		priceData := engine.generateMockPriceData(asset, 30)
		_, err := engine.AnalyzeAssetVolatility(ctx, asset, priceData)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.CalculateCorrelationMatrix(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
