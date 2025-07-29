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

// MockYieldRedisClient for testing
type MockYieldRedisClient struct {
	mock.Mock
}

func (m *MockYieldRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockYieldRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockYieldRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockYieldRedisClient) Exists(ctx context.Context, keys ...string) (bool, error) {
	args := m.Called(ctx, keys)
	return args.Bool(0), args.Error(1)
}

func (m *MockYieldRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockYieldRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockYieldRedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockYieldRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockYieldRedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func (m *MockYieldRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockYieldRedisClient) Pipeline() redis.Pipeline {
	args := m.Called()
	return args.Get(0).(redis.Pipeline)
}

func (m *MockYieldRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockYieldRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewYieldFarmingAutomation(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()

	automation := NewYieldFarmingAutomation(logger, mockCache, config)

	assert.NotNil(t, automation)
	assert.Equal(t, config.AutoCompoundingEnabled, automation.config.AutoCompoundingEnabled)
	assert.Equal(t, config.SupportedProtocols, automation.config.SupportedProtocols)
	assert.False(t, automation.isRunning)
	assert.NotNil(t, automation.activeFarms)
	assert.NotNil(t, automation.yieldOpportunities)
	assert.NotNil(t, automation.farmingStrategies)
}

func TestYieldFarmingAutomation_Start(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()
	config.OpportunityCheckInterval = 100 * time.Millisecond // Fast for testing

	automation := NewYieldFarmingAutomation(logger, mockCache, config)
	ctx := context.Background()

	err := automation.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, automation.IsRunning())

	// Check that strategies were created
	strategies := automation.GetFarmingStrategies()
	assert.NotEmpty(t, strategies)
	assert.Len(t, strategies, 3) // Conservative, moderate, aggressive

	// Clean up
	err = automation.Stop()
	assert.NoError(t, err)
	assert.False(t, automation.IsRunning())
}

func TestYieldFarmingAutomation_StartDisabled(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()
	config.Enabled = false

	automation := NewYieldFarmingAutomation(logger, mockCache, config)
	ctx := context.Background()

	err := automation.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, automation.IsRunning()) // Should remain false when disabled
}

func TestYieldFarmingAutomation_ScanForYieldOpportunities(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()

	automation := NewYieldFarmingAutomation(logger, mockCache, config)
	ctx := context.Background()

	// Start the automation
	err := automation.Start(ctx)
	require.NoError(t, err)

	// Scan for opportunities
	opportunities, err := automation.ScanForYieldOpportunities(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, opportunities)

	// Check opportunity structure
	for _, opp := range opportunities {
		assert.NotEmpty(t, opp.ID)
		assert.NotEmpty(t, opp.Protocol)
		assert.NotEmpty(t, opp.PoolAddress)
		assert.True(t, opp.CurrentAPY.GreaterThan(decimal.Zero))
		assert.True(t, opp.TVL.GreaterThan(decimal.Zero))
		assert.True(t, opp.Confidence.GreaterThan(decimal.Zero))
	}

	// Clean up
	automation.Stop()
}

func TestYieldFarmingAutomation_EnterFarm(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()

	automation := NewYieldFarmingAutomation(logger, mockCache, config)
	ctx := context.Background()

	err := automation.Start(ctx)
	require.NoError(t, err)

	// Get opportunities first
	opportunities, err := automation.ScanForYieldOpportunities(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, opportunities)

	// Get strategies
	strategies := automation.GetFarmingStrategies()
	require.NotEmpty(t, strategies)

	// Enter farm
	amount := decimal.NewFromFloat(1000)
	farm, err := automation.EnterFarm(ctx, opportunities[0].ID, amount, strategies[0].ID)
	assert.NoError(t, err)
	assert.NotNil(t, farm)
	assert.Equal(t, amount, farm.LiquidityAmount)
	assert.Equal(t, "active", farm.Status)
	assert.Equal(t, strategies[0].ID, farm.Strategy)

	// Check that farm is tracked
	activeFarms := automation.GetActiveFarms()
	assert.Len(t, activeFarms, 1)
	assert.Equal(t, farm.ID, activeFarms[0].ID)

	automation.Stop()
}

func TestYieldFarmingAutomation_ExitFarm(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()

	automation := NewYieldFarmingAutomation(logger, mockCache, config)
	ctx := context.Background()

	err := automation.Start(ctx)
	require.NoError(t, err)

	// Enter farm first
	opportunities, err := automation.ScanForYieldOpportunities(ctx)
	require.NoError(t, err)
	strategies := automation.GetFarmingStrategies()
	require.NotEmpty(t, strategies)

	farm, err := automation.EnterFarm(ctx, opportunities[0].ID, decimal.NewFromFloat(1000), strategies[0].ID)
	require.NoError(t, err)

	// Exit farm partially
	result, err := automation.ExitFarm(ctx, farm.ID, decimal.NewFromFloat(0.5))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	// Check farm is still active but with reduced amount
	updatedFarm, err := automation.GetActiveFarm(farm.ID)
	assert.NoError(t, err)
	assert.Equal(t, decimal.NewFromFloat(500), updatedFarm.LiquidityAmount)

	// Exit farm completely
	result, err = automation.ExitFarm(ctx, farm.ID, decimal.NewFromFloat(1.0))
	assert.NoError(t, err)
	assert.True(t, result.Success)

	// Check farm is no longer active
	_, err = automation.GetActiveFarm(farm.ID)
	assert.Error(t, err)

	automation.Stop()
}

func TestYieldFarmingAutomation_CompoundRewards(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()

	automation := NewYieldFarmingAutomation(logger, mockCache, config)
	ctx := context.Background()

	err := automation.Start(ctx)
	require.NoError(t, err)

	// Enter farm first
	opportunities, err := automation.ScanForYieldOpportunities(ctx)
	require.NoError(t, err)
	strategies := automation.GetFarmingStrategies()
	require.NotEmpty(t, strategies)

	farm, err := automation.EnterFarm(ctx, opportunities[0].ID, decimal.NewFromFloat(1000), strategies[0].ID)
	require.NoError(t, err)

	// Simulate earned rewards
	automation.mutex.Lock()
	farm.RewardsEarned = decimal.NewFromFloat(50) // $50 in rewards
	automation.mutex.Unlock()

	// Compound rewards
	result, err := automation.CompoundRewards(ctx, farm.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	// Check that rewards were compounded
	updatedFarm, err := automation.GetActiveFarm(farm.ID)
	assert.NoError(t, err)
	assert.Equal(t, decimal.Zero, updatedFarm.RewardsEarned)                 // Rewards should be zero after compounding
	assert.Equal(t, decimal.NewFromFloat(1050), updatedFarm.LiquidityAmount) // Should include compounded rewards

	automation.Stop()
}

func TestYieldFarmingAutomation_FarmingStrategies(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()

	automation := NewYieldFarmingAutomation(logger, mockCache, config)

	// Test creating a new strategy
	newStrategy := &FarmingStrategy{
		Name:                 "Test Strategy",
		Description:          "Test strategy for unit tests",
		TargetAPY:            decimal.NewFromFloat(0.2), // 20%
		MaxImpermanentLoss:   decimal.NewFromFloat(0.1), // 10%
		PreferredProtocols:   []string{"uniswap"},
		PreferredChains:      []string{"ethereum"},
		RiskLevel:            "moderate",
		CompoundingFrequency: 6 * time.Hour,
		AllocationPercentage: decimal.NewFromFloat(0.3), // 30%
		IsActive:             true,
	}

	err := automation.CreateFarmingStrategy(newStrategy)
	assert.NoError(t, err)
	assert.NotEmpty(t, newStrategy.ID)
	assert.NotZero(t, newStrategy.CreatedAt)

	// Test getting the strategy
	retrievedStrategy, err := automation.GetFarmingStrategy(newStrategy.ID)
	assert.NoError(t, err)
	assert.Equal(t, newStrategy.Name, retrievedStrategy.Name)
	assert.Equal(t, newStrategy.TargetAPY, retrievedStrategy.TargetAPY)

	// Test updating the strategy
	updates := &FarmingStrategy{
		Name:      "Updated Test Strategy",
		TargetAPY: decimal.NewFromFloat(0.25), // 25%
	}

	err = automation.UpdateFarmingStrategy(newStrategy.ID, updates)
	assert.NoError(t, err)

	updatedStrategy, err := automation.GetFarmingStrategy(newStrategy.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Strategy", updatedStrategy.Name)
	assert.Equal(t, decimal.NewFromFloat(0.25), updatedStrategy.TargetAPY)

	// Test deleting the strategy
	err = automation.DeleteFarmingStrategy(newStrategy.ID)
	assert.NoError(t, err)

	_, err = automation.GetFarmingStrategy(newStrategy.ID)
	assert.Error(t, err)
}

func TestYieldFarmingAutomation_PerformanceMetrics(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()

	automation := NewYieldFarmingAutomation(logger, mockCache, config)
	ctx := context.Background()

	err := automation.Start(ctx)
	require.NoError(t, err)

	// Get initial metrics
	metrics := automation.GetPerformanceMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, decimal.Zero, metrics.TotalValueLocked)
	assert.Equal(t, int64(0), metrics.ActiveFarmsCount)

	// Enter a farm
	opportunities, err := automation.ScanForYieldOpportunities(ctx)
	require.NoError(t, err)
	strategies := automation.GetFarmingStrategies()
	require.NotEmpty(t, strategies)

	amount := decimal.NewFromFloat(1000)
	farm, err := automation.EnterFarm(ctx, opportunities[0].ID, amount, strategies[0].ID)
	require.NoError(t, err)

	// Check updated metrics
	updatedMetrics := automation.GetPerformanceMetrics()
	assert.Equal(t, amount, updatedMetrics.TotalValueLocked)
	assert.Equal(t, int64(1), updatedMetrics.ActiveFarmsCount)
	assert.Contains(t, updatedMetrics.ProtocolDistribution, farm.Protocol)

	automation.Stop()
}

func TestGetDefaultYieldFarmingConfig(t *testing.T) {
	config := GetDefaultYieldFarmingConfig()

	assert.True(t, config.Enabled)
	assert.True(t, config.AutoCompoundingEnabled)
	assert.True(t, config.PoolMigrationEnabled)
	assert.True(t, config.ImpermanentLossProtection)
	assert.Equal(t, decimal.NewFromFloat(0.05), config.MinYieldThreshold)
	assert.Equal(t, decimal.NewFromFloat(0.01), config.MaxSlippageTolerance)
	assert.Equal(t, 12*time.Hour, config.CompoundingInterval)
	assert.Equal(t, 5*time.Minute, config.OpportunityCheckInterval)
	assert.NotEmpty(t, config.SupportedProtocols)
	assert.NotEmpty(t, config.SupportedChains)
	assert.Equal(t, "moderate", config.RiskLevel)
}

// Benchmark tests
func BenchmarkYieldFarmingAutomation_ScanForYieldOpportunities(b *testing.B) {
	logger := logger.New("benchmark")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()

	automation := NewYieldFarmingAutomation(logger, mockCache, config)
	ctx := context.Background()

	err := automation.Start(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer automation.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := automation.ScanForYieldOpportunities(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYieldFarmingAutomation_EnterFarm(b *testing.B) {
	logger := logger.New("benchmark")
	mockCache := &MockYieldRedisClient{}
	config := GetDefaultYieldFarmingConfig()

	automation := NewYieldFarmingAutomation(logger, mockCache, config)
	ctx := context.Background()

	err := automation.Start(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer automation.Stop()

	// Get opportunities and strategies
	opportunities, err := automation.ScanForYieldOpportunities(ctx)
	if err != nil {
		b.Fatal(err)
	}
	strategies := automation.GetFarmingStrategies()
	if len(strategies) == 0 {
		b.Fatal("No strategies available")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different opportunity for each iteration to avoid conflicts
		oppIndex := i % len(opportunities)
		amount := decimal.NewFromFloat(1000)

		_, err := automation.EnterFarm(ctx, opportunities[oppIndex].ID, amount, strategies[0].ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}
