package defi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Mock types removed since NewYieldAggregator expects concrete types
// In a real integration test, actual client instances would be used

func TestYieldAggregator_GetBestOpportunities(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}

	// Use nil for clients since NewYieldAggregator expects concrete types
	// In a real integration test, these would be actual client instances
	aggregator := NewYieldAggregator(logger, mockRedis, nil, nil)

	// Populate test opportunities
	aggregator.opportunities = map[string]*YieldFarmingOpportunity{
		"opp1": {
			ID:       "opp1",
			Protocol: ProtocolTypeUniswap,
			APY:      decimal.NewFromFloat(0.15), // 15%
			TVL:      decimal.NewFromFloat(1000000),
			Risk:     RiskLevelMedium,
			Active:   true,
		},
		"opp2": {
			ID:       "opp2",
			Protocol: ProtocolTypeAave,
			APY:      decimal.NewFromFloat(0.08), // 8%
			TVL:      decimal.NewFromFloat(5000000),
			Risk:     RiskLevelLow,
			Active:   true,
		},
		"opp3": {
			ID:       "opp3",
			Protocol: ProtocolType("coffee"),
			APY:      decimal.NewFromFloat(0.12), // 12%
			TVL:      decimal.NewFromFloat(2000000),
			Risk:     RiskLevelMedium,
			Active:   true,
		},
		"opp4": {
			ID:       "opp4",
			Protocol: ProtocolTypeUniswap,
			APY:      decimal.NewFromFloat(0.05), // 5%
			TVL:      decimal.NewFromFloat(500000),
			Risk:     RiskLevelLow,
			Active:   false, // Inactive
		},
	}

	ctx := context.Background()

	// Act
	opportunities, err := aggregator.GetBestOpportunities(ctx, 3)

	// Assert
	require.NoError(t, err)
	assert.Len(t, opportunities, 3, "Should return exactly 3 opportunities")

	// Verify sorting by APY (descending)
	assert.True(t, opportunities[0].APY.GreaterThanOrEqual(opportunities[1].APY))
	assert.True(t, opportunities[1].APY.GreaterThanOrEqual(opportunities[2].APY))

	// Verify all returned opportunities are active
	for _, opp := range opportunities {
		assert.True(t, opp.Active, "All returned opportunities should be active")
	}

	// Verify highest APY is first
	assert.Equal(t, decimal.NewFromFloat(0.15), opportunities[0].APY)
}

func TestYieldAggregator_GetOptimalStrategy(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}

	// Use nil for clients since NewYieldAggregator expects concrete types
	aggregator := NewYieldAggregator(logger, mockRedis, nil, nil)

	// Populate test opportunities
	aggregator.opportunities = map[string]*YieldFarmingOpportunity{
		"low_risk": {
			ID:         "low_risk",
			Protocol:   ProtocolTypeAave,
			APY:        decimal.NewFromFloat(0.06), // 6%
			Risk:       RiskLevelLow,
			MinDeposit: decimal.NewFromFloat(100),
			Active:     true,
		},
		"medium_risk": {
			ID:         "medium_risk",
			Protocol:   ProtocolTypeUniswap,
			APY:        decimal.NewFromFloat(0.12), // 12%
			Risk:       RiskLevelMedium,
			MinDeposit: decimal.NewFromFloat(500),
			Active:     true,
		},
		"high_risk": {
			ID:         "high_risk",
			Protocol:   ProtocolType("defi_protocol"),
			APY:        decimal.NewFromFloat(0.25), // 25%
			Risk:       RiskLevelHigh,
			MinDeposit: decimal.NewFromFloat(1000),
			Active:     true,
		},
	}

	ctx := context.Background()

	// Test case 1: Conservative strategy
	conservativeReq := &OptimalStrategyRequest{
		InvestmentAmount: decimal.NewFromFloat(10000),
		RiskTolerance:    RiskLevelLow,
		MinAPY:           decimal.NewFromFloat(0.05), // 5%
		Diversification:  false,
	}

	strategy, err := aggregator.GetOptimalStrategy(ctx, conservativeReq)

	require.NoError(t, err)
	require.NotNil(t, strategy)
	assert.Equal(t, YieldStrategyTypeConservative, strategy.Type)
	assert.Len(t, strategy.Opportunities, 1)
	assert.Equal(t, RiskLevelLow, strategy.Opportunities[0].Risk)

	// Test case 2: Aggressive strategy with diversification
	aggressiveReq := &OptimalStrategyRequest{
		InvestmentAmount: decimal.NewFromFloat(50000),
		RiskTolerance:    RiskLevelHigh,
		MinAPY:           decimal.NewFromFloat(0.08), // 8%
		Diversification:  true,
	}

	strategy2, err := aggregator.GetOptimalStrategy(ctx, aggressiveReq)

	require.NoError(t, err)
	require.NotNil(t, strategy2)
	assert.Equal(t, YieldStrategyTypeAggressive, strategy2.Type)
	assert.True(t, len(strategy2.Opportunities) > 1, "Should diversify across multiple opportunities")
}

func TestYieldAggregator_CalculatePoolAPY(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}

	aggregator := NewYieldAggregator(logger, mockRedis, nil, nil)

	testCases := []struct {
		name        string
		pool        LiquidityPool
		expectedMin decimal.Decimal
		expectedMax decimal.Decimal
	}{
		{
			name: "High TVL Pool",
			pool: LiquidityPool{
				Fee: decimal.NewFromFloat(0.003),    // 0.3%
				TVL: decimal.NewFromFloat(50000000), // $50M
			},
			expectedMin: decimal.NewFromFloat(0.5), // 50% of base APY
			expectedMax: decimal.NewFromFloat(1.5), // 150% of base APY
		},
		{
			name: "Medium TVL Pool",
			pool: LiquidityPool{
				Fee: decimal.NewFromFloat(0.003),   // 0.3%
				TVL: decimal.NewFromFloat(5000000), // $5M
			},
			expectedMin: decimal.NewFromFloat(0.7), // 70% of base APY
			expectedMax: decimal.NewFromFloat(1.8), // 180% of base APY
		},
		{
			name: "Low TVL Pool",
			pool: LiquidityPool{
				Fee: decimal.NewFromFloat(0.01),   // 1%
				TVL: decimal.NewFromFloat(100000), // $100k
			},
			expectedMin: decimal.NewFromFloat(2.0), // 200% of base APY
			expectedMax: decimal.NewFromFloat(4.0), // 400% of base APY (capped at 200%)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			apy := aggregator.calculatePoolAPY(tc.pool)

			// Assert
			assert.True(t, apy.GreaterThan(decimal.Zero), "APY should be positive")
			assert.True(t, apy.LessThanOrEqual(decimal.NewFromFloat(2.0)), "APY should be capped at 200%")
		})
	}
}

func TestYieldAggregator_CalculatePoolRisk(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}

	aggregator := NewYieldAggregator(logger, mockRedis, nil, nil)

	testCases := []struct {
		name         string
		pool         LiquidityPool
		expectedRisk RiskLevel
	}{
		{
			name: "Stable-Stable Pair High TVL",
			pool: LiquidityPool{
				Token0: Token{Symbol: "USDC"},
				Token1: Token{Symbol: "USDT"},
				TVL:    decimal.NewFromFloat(10000000), // $10M
			},
			expectedRisk: RiskLevelLow,
		},
		{
			name: "Stable-Volatile Pair Medium TVL",
			pool: LiquidityPool{
				Token0: Token{Symbol: "USDC"},
				Token1: Token{Symbol: "ETH"},
				TVL:    decimal.NewFromFloat(5000000), // $5M
			},
			expectedRisk: RiskLevelMedium,
		},
		{
			name: "Volatile-Volatile Pair Low TVL",
			pool: LiquidityPool{
				Token0: Token{Symbol: "ETH"},
				Token1: Token{Symbol: "BTC"},
				TVL:    decimal.NewFromFloat(50000), // $50k
			},
			expectedRisk: RiskLevelHigh,
		},
		{
			name: "Unknown Tokens Low TVL",
			pool: LiquidityPool{
				Token0: Token{Symbol: "UNKNOWN1"},
				Token1: Token{Symbol: "UNKNOWN2"},
				TVL:    decimal.NewFromFloat(10000), // $10k
			},
			expectedRisk: RiskLevelHigh,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			risk := aggregator.calculatePoolRisk(tc.pool)

			// Assert
			assert.Equal(t, tc.expectedRisk, risk)
		})
	}
}

func TestYieldAggregator_FilterOpportunities(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}

	aggregator := NewYieldAggregator(logger, mockRedis, nil, nil)

	opportunities := []*YieldFarmingOpportunity{
		{
			ID:         "low_apy",
			APY:        decimal.NewFromFloat(0.03), // 3%
			Risk:       RiskLevelLow,
			MinDeposit: decimal.NewFromFloat(100),
			LockPeriod: 0,
		},
		{
			ID:         "good_apy",
			APY:        decimal.NewFromFloat(0.08), // 8%
			Risk:       RiskLevelMedium,
			MinDeposit: decimal.NewFromFloat(500),
			LockPeriod: time.Hour * 24 * 7, // 1 week
		},
		{
			ID:         "high_risk",
			APY:        decimal.NewFromFloat(0.15), // 15%
			Risk:       RiskLevelHigh,
			MinDeposit: decimal.NewFromFloat(1000),
			LockPeriod: time.Hour * 24 * 30, // 1 month
		},
		{
			ID:         "high_min_deposit",
			APY:        decimal.NewFromFloat(0.10), // 10%
			Risk:       RiskLevelMedium,
			MinDeposit: decimal.NewFromFloat(20000), // Too high
			LockPeriod: 0,
		},
	}

	req := &OptimalStrategyRequest{
		InvestmentAmount: decimal.NewFromFloat(10000),
		RiskTolerance:    RiskLevelMedium,
		MinAPY:           decimal.NewFromFloat(0.05), // 5%
		MaxLockPeriod:    time.Hour * 24 * 14,        // 2 weeks
	}

	// Act
	filtered := aggregator.filterOpportunities(opportunities, req)

	// Assert
	assert.Len(t, filtered, 1, "Should filter to only one opportunity")
	assert.Equal(t, "good_apy", filtered[0].ID)
}

func TestYieldAggregator_CalculateImpermanentLoss(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}

	aggregator := NewYieldAggregator(logger, mockRedis, nil, nil)

	testCases := []struct {
		name        string
		pool        LiquidityPool
		expectedMin decimal.Decimal
		expectedMax decimal.Decimal
	}{
		{
			name: "Low Risk Pool",
			pool: LiquidityPool{
				Token0: Token{Symbol: "USDC"},
				Token1: Token{Symbol: "USDT"},
				TVL:    decimal.NewFromFloat(10000000),
			},
			expectedMin: decimal.NewFromFloat(0.005), // 0.5%
			expectedMax: decimal.NewFromFloat(0.02),  // 2%
		},
		{
			name: "Medium Risk Pool",
			pool: LiquidityPool{
				Token0: Token{Symbol: "USDC"},
				Token1: Token{Symbol: "ETH"},
				TVL:    decimal.NewFromFloat(5000000),
			},
			expectedMin: decimal.NewFromFloat(0.03), // 3%
			expectedMax: decimal.NewFromFloat(0.08), // 8%
		},
		{
			name: "High Risk Pool",
			pool: LiquidityPool{
				Token0: Token{Symbol: "ETH"},
				Token1: Token{Symbol: "ALTCOIN"},
				TVL:    decimal.NewFromFloat(100000),
			},
			expectedMin: decimal.NewFromFloat(0.10), // 10%
			expectedMax: decimal.NewFromFloat(0.20), // 20%
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			il := aggregator.calculateImpermanentLoss(tc.pool)

			// Assert
			assert.True(t, il.GreaterThanOrEqual(tc.expectedMin),
				"IL should be >= %s, got %s", tc.expectedMin, il)
			assert.True(t, il.LessThanOrEqual(tc.expectedMax),
				"IL should be <= %s, got %s", tc.expectedMax, il)
		})
	}
}

// Benchmark tests
func BenchmarkYieldAggregator_FilterOpportunities(b *testing.B) {
	logger := logger.New("benchmark")
	mockRedis := &MockRedisClient{}

	aggregator := NewYieldAggregator(logger, mockRedis, nil, nil)

	// Create 100 test opportunities
	opportunities := make([]*YieldFarmingOpportunity, 100)
	for i := 0; i < 100; i++ {
		opportunities[i] = &YieldFarmingOpportunity{
			ID:         fmt.Sprintf("opp_%d", i),
			APY:        decimal.NewFromFloat(0.05 + float64(i)*0.001),
			Risk:       RiskLevel(i % 3), // Cycle through risk levels
			MinDeposit: decimal.NewFromFloat(100 + float64(i)*10),
			LockPeriod: time.Duration(i) * time.Hour,
		}
	}

	req := &OptimalStrategyRequest{
		InvestmentAmount: decimal.NewFromFloat(10000),
		RiskTolerance:    RiskLevelMedium,
		MinAPY:           decimal.NewFromFloat(0.05),
		MaxLockPeriod:    time.Hour * 24 * 30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aggregator.filterOpportunities(opportunities, req)
	}
}

func TestYieldAggregator_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	logger := logger.New("integration-test")
	mockRedis := &MockRedisClient{}

	// Use nil for clients since NewYieldAggregator expects concrete types
	// In a real integration test, these would be actual client instances
	aggregator := NewYieldAggregator(logger, mockRedis, nil, nil)

	ctx := context.Background()

	// Note: Since we're using nil clients, we can't start the aggregator
	// as it would try to call methods on nil clients. Instead, we test
	// the core functionality that doesn't require external clients.

	// Test optimal strategy with pre-populated opportunities
	aggregator.opportunities = map[string]*YieldFarmingOpportunity{
		"test_opp": {
			ID:         "test_opp",
			Protocol:   ProtocolTypeUniswap,
			APY:        decimal.NewFromFloat(0.10),
			Risk:       RiskLevelMedium,
			MinDeposit: decimal.NewFromFloat(100),
			Active:     true,
		},
	}

	req := &OptimalStrategyRequest{
		InvestmentAmount: decimal.NewFromFloat(10000),
		RiskTolerance:    RiskLevelMedium,
		MinAPY:           decimal.NewFromFloat(0.05),
		Diversification:  false,
	}

	strategy, err := aggregator.GetOptimalStrategy(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, strategy)

	// Test getting best opportunities
	opportunities, err := aggregator.GetBestOpportunities(ctx, 5)
	require.NoError(t, err)
	assert.NotEmpty(t, opportunities, "Should find yield opportunities")
}
