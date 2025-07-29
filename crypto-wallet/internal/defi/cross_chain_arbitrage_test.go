package defi

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRedisClient for testing
type MockCrossChainRedisClient struct {
	mock.Mock
}

func (m *MockCrossChainRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCrossChainRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCrossChainRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCrossChainRedisClient) Exists(ctx context.Context, keys ...string) (bool, error) {
	args := m.Called(ctx, keys)
	return args.Bool(0), args.Error(1)
}

func (m *MockCrossChainRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCrossChainRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockCrossChainRedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockCrossChainRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockCrossChainRedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func (m *MockCrossChainRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockCrossChainRedisClient) Pipeline() redis.Pipeline {
	args := m.Called()
	return args.Get(0).(redis.Pipeline)
}

func (m *MockCrossChainRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCrossChainRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewCrossChainArbitrageEngine(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled:                true,
		MinProfitThreshold:     decimal.NewFromFloat(0.01),
		MaxBridgeAmount:        decimal.NewFromFloat(1000),
		BridgeTimeoutMinutes:   30,
		ScanInterval:           30 * time.Second,
		MaxConcurrentArbitrage: 3,
		EnabledChains:          []string{"ethereum", "polygon", "arbitrum"},
		EnabledBridges:         []string{"polygon", "arbitrum"},
		SlippageTolerance:      decimal.NewFromFloat(0.05),
		AutoExecute:            false,
		RiskLevel:              "medium",
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)

	assert.NotNil(t, engine)
	assert.Equal(t, config.MinProfitThreshold, engine.config.MinProfitThreshold)
	assert.Equal(t, config.EnabledChains, engine.config.EnabledChains)
	assert.False(t, engine.isRunning)
	assert.NotNil(t, engine.opportunities)
	assert.NotNil(t, engine.activeArbitrages)
	assert.NotNil(t, engine.executionHistory)
}

func TestCrossChainArbitrageEngine_Start(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled:            true,
		EnabledChains:      []string{"ethereum", "polygon"},
		EnabledBridges:     []string{"polygon"},
		ScanInterval:       1 * time.Second,
		MinProfitThreshold: decimal.NewFromFloat(0.01),
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, engine.isRunning)

	// Clean up
	engine.Stop()
	assert.False(t, engine.isRunning)
}

func TestCrossChainArbitrageEngine_StartDisabled(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled: false,
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, engine.isRunning) // Should remain false when disabled
}

func TestCrossChainArbitrageEngine_ScanForOpportunities(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled:            true,
		EnabledChains:      []string{"ethereum", "polygon"},
		EnabledBridges:     []string{"polygon"},
		MinProfitThreshold: decimal.NewFromFloat(0.01),
		MaxBridgeAmount:    decimal.NewFromFloat(1000),
		SlippageTolerance:  decimal.NewFromFloat(0.05),
		RiskLevel:          "medium",
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)
	ctx := context.Background()

	// Initialize the engine
	err := engine.Start(ctx)
	require.NoError(t, err)

	// Scan for opportunities
	opportunities, err := engine.ScanForOpportunities(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, opportunities)

	// Clean up
	engine.Stop()
}

func TestCrossChainArbitrageEngine_ExecuteOpportunity(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled:                true,
		EnabledChains:          []string{"ethereum", "polygon"},
		EnabledBridges:         []string{"polygon"},
		MinProfitThreshold:     decimal.NewFromFloat(0.01),
		MaxBridgeAmount:        decimal.NewFromFloat(1000),
		BridgeTimeoutMinutes:   30,
		MaxConcurrentArbitrage: 3,
		SlippageTolerance:      decimal.NewFromFloat(0.05),
		AutoExecute:            false,
		RiskLevel:              "medium",
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)
	ctx := context.Background()

	// Initialize the engine
	err := engine.Start(ctx)
	require.NoError(t, err)

	// Create a mock opportunity
	opportunity := &CrossChainOpportunity{
		ID:              "test_opp_1",
		TokenAddress:    "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
		TokenSymbol:     "USDC",
		SourceChain:     "ethereum",
		TargetChain:     "polygon",
		SourceExchange:  "uniswap",
		TargetExchange:  "quickswap",
		BridgeProtocol:  "polygon",
		Amount:          decimal.NewFromFloat(100),
		SourcePrice:     decimal.NewFromFloat(1.0),
		TargetPrice:     decimal.NewFromFloat(1.02),
		ProfitMargin:    decimal.NewFromFloat(0.02),
		EstimatedProfit: decimal.NewFromFloat(2.0),
		BridgeFee:       decimal.NewFromFloat(0.1),
		GasCosts:        decimal.NewFromFloat(0.5),
		NetProfit:       decimal.NewFromFloat(1.4),
		BridgeTime:      10 * time.Minute,
		Confidence:      decimal.NewFromFloat(0.85),
		RiskScore:       decimal.NewFromFloat(0.3),
		DetectedAt:      time.Now(),
		ExpiresAt:       time.Now().Add(5 * time.Minute),
		Status:          "detected",
	}

	// Add opportunity to engine
	engine.mutex.Lock()
	engine.opportunities[opportunity.ID] = opportunity
	engine.mutex.Unlock()

	// Execute the opportunity
	execution, err := engine.ExecuteOpportunity(ctx, opportunity.ID)
	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, opportunity.ID, execution.OpportunityID)
	assert.Equal(t, opportunity.SourceChain, execution.SourceChain)
	assert.Equal(t, opportunity.TargetChain, execution.TargetChain)

	// Clean up
	engine.Stop()
}

func TestCrossChainArbitrageEngine_ExecuteOpportunity_NotFound(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled: true,
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)
	ctx := context.Background()

	// Try to execute non-existent opportunity
	_, err := engine.ExecuteOpportunity(ctx, "non_existent_id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "opportunity not found")
}

func TestCrossChainArbitrageEngine_ExecuteOpportunity_Expired(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled: true,
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)
	ctx := context.Background()

	// Create an expired opportunity
	opportunity := &CrossChainOpportunity{
		ID:        "expired_opp",
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired 1 minute ago
	}

	engine.mutex.Lock()
	engine.opportunities[opportunity.ID] = opportunity
	engine.mutex.Unlock()

	// Try to execute expired opportunity
	_, err := engine.ExecuteOpportunity(ctx, opportunity.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "opportunity expired")
}

func TestCrossChainArbitrageEngine_GetOpportunities(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled: true,
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)

	// Add some opportunities
	opp1 := &CrossChainOpportunity{ID: "opp1", TokenSymbol: "USDC"}
	opp2 := &CrossChainOpportunity{ID: "opp2", TokenSymbol: "DAI"}

	engine.mutex.Lock()
	engine.opportunities[opp1.ID] = opp1
	engine.opportunities[opp2.ID] = opp2
	engine.mutex.Unlock()

	opportunities := engine.GetOpportunities()
	assert.Equal(t, 2, len(opportunities))

	// Check that both opportunities are present
	foundOpp1, foundOpp2 := false, false
	for _, opp := range opportunities {
		if opp.ID == "opp1" {
			foundOpp1 = true
		}
		if opp.ID == "opp2" {
			foundOpp2 = true
		}
	}
	assert.True(t, foundOpp1)
	assert.True(t, foundOpp2)
}

func TestCrossChainArbitrageEngine_GetMetrics(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled: true,
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)

	metrics := engine.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalOpportunities)
	assert.Equal(t, int64(0), metrics.ExecutedArbitrages)
	assert.Equal(t, int64(0), metrics.SuccessfulArbitrages)
	assert.Equal(t, int64(0), metrics.FailedArbitrages)
	assert.True(t, metrics.TotalProfit.IsZero())
	assert.True(t, metrics.NetProfit.IsZero())
}

func TestCrossChainArbitrageEngine_UpdateConfig(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled:            true,
		MinProfitThreshold: decimal.NewFromFloat(0.01),
		AutoExecute:        false,
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)

	// Update configuration
	newConfig := CrossChainConfig{
		Enabled:            true,
		MinProfitThreshold: decimal.NewFromFloat(0.02),
		AutoExecute:        true,
		EnabledChains:      []string{"ethereum", "polygon", "arbitrum"},
	}

	err := engine.UpdateConfig(newConfig)
	assert.NoError(t, err)
	assert.Equal(t, newConfig.MinProfitThreshold, engine.config.MinProfitThreshold)
	assert.Equal(t, newConfig.AutoExecute, engine.config.AutoExecute)
	assert.Equal(t, newConfig.EnabledChains, engine.config.EnabledChains)
}

func TestBridgeClients(t *testing.T) {
	logger := logger.New("test")

	// Test Polygon Bridge
	polygonBridge := &PolygonBridgeClient{logger: logger}
	result, err := polygonBridge.Bridge(context.Background(), "0xtest", decimal.NewFromFloat(100), "ethereum")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, decimal.NewFromFloat(100), result.Amount)
	assert.Equal(t, "pending", result.Status)

	// Test Arbitrum Bridge
	arbitrumBridge := &ArbitrumBridgeClient{logger: logger}
	result, err = arbitrumBridge.Bridge(context.Background(), "0xtest", decimal.NewFromFloat(100), "ethereum")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, decimal.NewFromFloat(100), result.Amount)

	// Test Optimism Bridge
	optimismBridge := &OptimismBridgeClient{logger: logger}
	result, err = optimismBridge.Bridge(context.Background(), "0xtest", decimal.NewFromFloat(100), "ethereum")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test Avalanche Bridge
	avalancheBridge := &AvalancheBridgeClient{logger: logger}
	result, err = avalancheBridge.Bridge(context.Background(), "0xtest", decimal.NewFromFloat(100), "ethereum")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMockExchangeClient(t *testing.T) {
	logger := logger.New("test")
	exchange := &MockExchangeClient{logger: logger}

	// Test GetTokenPrice
	price, err := exchange.GetTokenPrice(context.Background(), "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1")
	assert.NoError(t, err)
	assert.Equal(t, decimal.NewFromFloat(1.0), price)

	// Test GetLiquidity
	liquidity, err := exchange.GetLiquidity(context.Background(), "0xtest")
	assert.NoError(t, err)
	assert.Equal(t, decimal.NewFromFloat(1000000.0), liquidity)

	// Test ExecuteTrade
	tradeParams := &TradeParams{
		TokenIn:      "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
		TokenOut:     "0x6B175474E89094C44Da98b954EedeAC495271d0F",
		AmountIn:     decimal.NewFromFloat(100),
		MinAmountOut: decimal.NewFromFloat(95),
		Slippage:     decimal.NewFromFloat(0.05),
		Deadline:     time.Now().Add(10 * time.Minute),
	}

	result, err := exchange.ExecuteTrade(context.Background(), tradeParams)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, tradeParams.AmountIn, result.AmountIn)

	// Test GetSupportedTokens
	tokens := exchange.GetSupportedTokens()
	assert.NotEmpty(t, tokens)
	assert.Contains(t, tokens, "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1")
}

// Benchmark tests
func BenchmarkCrossChainArbitrageEngine_ScanForOpportunities(b *testing.B) {
	logger := logger.New("benchmark")
	mockCache := &MockCrossChainRedisClient{}

	config := CrossChainConfig{
		Enabled:            true,
		EnabledChains:      []string{"ethereum", "polygon"},
		MinProfitThreshold: decimal.NewFromFloat(0.01),
		MaxBridgeAmount:    decimal.NewFromFloat(1000),
		SlippageTolerance:  decimal.NewFromFloat(0.05),
		RiskLevel:          "medium",
	}

	engine := NewCrossChainArbitrageEngine(logger, mockCache, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer engine.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.ScanForOpportunities(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
