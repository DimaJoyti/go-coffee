package defi

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: MockEthereumClient removed since NewOnChainAnalyzer expects concrete types
// In a real integration test, actual blockchain client instances would be used

func TestOnChainAnalyzer_Creation(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}

	// Act
	// Use nil for blockchain clients since NewOnChainAnalyzer expects concrete types
	// In a real integration test, these would be actual client instances
	analyzer := NewOnChainAnalyzer(logger, mockRedis, nil, nil, nil)

	// Assert
	assert.NotNil(t, analyzer)
	assert.Equal(t, time.Minute*2, analyzer.scanInterval)
	assert.Equal(t, uint64(100), analyzer.blockRange)
	assert.NotNil(t, analyzer.metrics)
	assert.NotNil(t, analyzer.whaleWatches)
	assert.NotNil(t, analyzer.liquidityEvents)
}

func TestOnChainAnalyzer_GetMetrics(t *testing.T) {
	// Arrange
	analyzer := createTestAnalyzer(t)
	ctx := context.Background()

	// Add test metrics
	testToken := "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1"
	testMetrics := &OnChainMetrics{
		Token: Token{
			Address: testToken,
			Symbol:  "USDC",
			Chain:   ChainEthereum,
		},
		Price:           decimal.NewFromFloat(1.0),
		Volume24h:       decimal.NewFromFloat(1000000),
		Liquidity:       decimal.NewFromFloat(5000000),
		MarketCap:       decimal.NewFromFloat(50000000000),
		Holders:         100000,
		Transactions24h: 50000,
		Volatility:      decimal.NewFromFloat(0.02),
		Timestamp:       time.Now(),
	}

	analyzer.mutex.Lock()
	analyzer.metrics[testToken] = testMetrics
	analyzer.mutex.Unlock()

	// Act
	metrics, err := analyzer.GetMetrics(ctx, testToken)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, testMetrics.Token.Symbol, metrics.Token.Symbol)
	assert.Equal(t, testMetrics.Price, metrics.Price)
	assert.Equal(t, testMetrics.Volume24h, metrics.Volume24h)
}

func TestOnChainAnalyzer_GetMetrics_NotFound(t *testing.T) {
	// Arrange
	analyzer := createTestAnalyzer(t)
	ctx := context.Background()

	// Act
	metrics, err := analyzer.GetMetrics(ctx, "0xNonExistent")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "metrics not found")
}

func TestOnChainAnalyzer_GetWhaleActivity(t *testing.T) {
	// Arrange
	analyzer := createTestAnalyzer(t)
	ctx := context.Background()

	// Add test whale data
	testWhale := &WhaleWatch{
		Address:    "0x1234567890123456789012345678901234567890",
		Label:      "Test Whale",
		Chain:      ChainEthereum,
		Balance:    decimal.NewFromFloat(100000000),
		TxCount24h: 5,
		Volume24h:  decimal.NewFromFloat(1000000),
		Active:     true,
	}

	analyzer.mutex.Lock()
	analyzer.whaleWatches[testWhale.Address] = testWhale
	analyzer.mutex.Unlock()

	// Act
	whales, err := analyzer.GetWhaleActivity(ctx)

	// Assert
	require.NoError(t, err)
	assert.Len(t, whales, 1)
	assert.Equal(t, testWhale.Address, whales[0].Address)
	assert.Equal(t, testWhale.TxCount24h, whales[0].TxCount24h)
}

func TestOnChainAnalyzer_GenerateMarketSignals(t *testing.T) {
	// Arrange
	analyzer := createTestAnalyzer(t)
	ctx := context.Background()

	// Add test data for signal generation
	// High volume token
	highVolumeToken := &OnChainMetrics{
		Token: Token{
			Address: "0x1",
			Symbol:  "HIGHVOL",
			Chain:   ChainEthereum,
		},
		Volume24h: decimal.NewFromFloat(100000000), // $100M (high volume)
		Timestamp: time.Now(),
	}

	// Active whale
	activeWhale := &WhaleWatch{
		Address:    "0xWhale1",
		Label:      "Active Whale",
		Chain:      ChainEthereum,
		TxCount24h: 10, // High activity
		Volume24h:  decimal.NewFromFloat(50000000),
		Active:     true,
	}

	analyzer.mutex.Lock()
	analyzer.metrics["0x1"] = highVolumeToken
	analyzer.whaleWatches["0xWhale1"] = activeWhale
	analyzer.mutex.Unlock()

	// Act
	signals := analyzer.generateMarketSignals(ctx)

	// Assert
	assert.NotEmpty(t, signals)

	// Check for volume spike signal
	volumeSignalFound := false
	whaleSignalFound := false

	for _, signal := range signals {
		if signal.Type == SignalTypeVolumeSpike {
			volumeSignalFound = true
			assert.Equal(t, "HIGHVOL", signal.Token.Symbol)
			assert.Equal(t, SignalDirectionBullish, signal.Direction)
		}
		if signal.Type == SignalTypeWhaleMovement {
			whaleSignalFound = true
			assert.True(t, signal.Strength.GreaterThan(decimal.Zero))
		}
	}

	assert.True(t, volumeSignalFound, "Should generate volume spike signal")
	assert.True(t, whaleSignalFound, "Should generate whale movement signal")
}

func TestOnChainAnalyzer_ProcessLargeTransfer(t *testing.T) {
	// Arrange
	analyzer := createTestAnalyzer(t)
	ctx := context.Background()

	// Add whale to watch list
	whaleAddress := "0x1234567890123456789012345678901234567890"
	whale := &WhaleWatch{
		Address:    whaleAddress,
		Label:      "Test Whale",
		Chain:      ChainEthereum,
		TxCount24h: 0,
		Volume24h:  decimal.Zero,
		Active:     true,
	}

	analyzer.mutex.Lock()
	analyzer.whaleWatches[whaleAddress] = whale
	analyzer.mutex.Unlock()

	// Create large transfer event
	event := &BlockchainEvent{
		ID:          "transfer1",
		Type:        EventTypeLargeTransfer,
		Chain:       ChainEthereum,
		BlockNumber: 12345,
		TxHash:      "0xabc123",
		Token: Token{
			Address: "0xToken1",
			Symbol:  "TEST",
			Chain:   ChainEthereum,
		},
		Amount:    decimal.NewFromFloat(1000000), // $1M
		From:      whaleAddress,
		To:        "0xRecipient",
		Timestamp: time.Now(),
	}

	// Act
	analyzer.processLargeTransfer(ctx, event)

	// Assert
	analyzer.mutex.RLock()
	updatedWhale := analyzer.whaleWatches[whaleAddress]
	analyzer.mutex.RUnlock()

	assert.Equal(t, 1, updatedWhale.TxCount24h)
	assert.True(t, updatedWhale.Volume24h.Equal(decimal.NewFromFloat(1000000)))
}

func TestOnChainAnalyzer_CalculateTokenScore(t *testing.T) {
	// Arrange
	analyzer := createTestAnalyzer(t)

	testCases := []struct {
		name        string
		metrics     *OnChainMetrics
		expectedMin decimal.Decimal
		expectedMax decimal.Decimal
	}{
		{
			name: "High Quality Token",
			metrics: &OnChainMetrics{
				Volume24h:  decimal.NewFromFloat(50000000),  // $50M
				Liquidity:  decimal.NewFromFloat(100000000), // $100M
				Holders:    50000,
				Volatility: decimal.NewFromFloat(0.05), // 5%
			},
			expectedMin: decimal.NewFromFloat(80),
			expectedMax: decimal.NewFromFloat(100),
		},
		{
			name: "Medium Quality Token",
			metrics: &OnChainMetrics{
				Volume24h:  decimal.NewFromFloat(5000000),  // $5M
				Liquidity:  decimal.NewFromFloat(20000000), // $20M
				Holders:    5000,
				Volatility: decimal.NewFromFloat(0.15), // 15%
			},
			expectedMin: decimal.NewFromFloat(40),
			expectedMax: decimal.NewFromFloat(70),
		},
		{
			name: "Low Quality Token",
			metrics: &OnChainMetrics{
				Volume24h:  decimal.NewFromFloat(100000), // $100k
				Liquidity:  decimal.NewFromFloat(500000), // $500k
				Holders:    100,
				Volatility: decimal.NewFromFloat(0.50), // 50%
			},
			expectedMin: decimal.NewFromFloat(10),
			expectedMax: decimal.NewFromFloat(50),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			score := analyzer.calculateTokenScore(tc.metrics)

			// Assert
			assert.True(t, score.GreaterThanOrEqual(tc.expectedMin),
				"Score should be >= %s, got %s", tc.expectedMin, score)
			assert.True(t, score.LessThanOrEqual(tc.expectedMax),
				"Score should be <= %s, got %s", tc.expectedMax, score)
		})
	}
}

func TestOnChainAnalyzer_GenerateRecommendation(t *testing.T) {
	// Arrange
	analyzer := createTestAnalyzer(t)

	testCases := []struct {
		name           string
		metrics        *OnChainMetrics
		expectedPhrase string
	}{
		{
			name: "Strong Buy",
			metrics: &OnChainMetrics{
				Volume24h:  decimal.NewFromFloat(100000000),
				Liquidity:  decimal.NewFromFloat(200000000),
				Holders:    100000,
				Volatility: decimal.NewFromFloat(0.05),
			},
			expectedPhrase: "Strong Buy",
		},
		{
			name: "Buy",
			metrics: &OnChainMetrics{
				Volume24h:  decimal.NewFromFloat(20000000),
				Liquidity:  decimal.NewFromFloat(80000000),
				Holders:    20000,
				Volatility: decimal.NewFromFloat(0.10),
			},
			expectedPhrase: "Buy",
		},
		{
			name: "Hold",
			metrics: &OnChainMetrics{
				Volume24h:  decimal.NewFromFloat(5000000),
				Liquidity:  decimal.NewFromFloat(20000000),
				Holders:    5000,
				Volatility: decimal.NewFromFloat(0.15),
			},
			expectedPhrase: "Hold",
		},
		{
			name: "Caution",
			metrics: &OnChainMetrics{
				Volume24h:  decimal.NewFromFloat(100000),
				Liquidity:  decimal.NewFromFloat(1000000),
				Holders:    100,
				Volatility: decimal.NewFromFloat(0.50),
			},
			expectedPhrase: "Caution",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			recommendation := analyzer.generateRecommendation(tc.metrics)

			// Assert
			assert.Contains(t, recommendation, tc.expectedPhrase)
		})
	}
}

func TestOnChainAnalyzer_DetermineWhaleDirection(t *testing.T) {
	// Arrange
	analyzer := createTestAnalyzer(t)

	testCases := []struct {
		name              string
		whale             *WhaleWatch
		expectedDirection SignalDirection
	}{
		{
			name: "Bearish - High Volume",
			whale: &WhaleWatch{
				Volume24h:  decimal.NewFromFloat(20000000), // > $10M
				TxCount24h: 5,
			},
			expectedDirection: SignalDirectionBearish,
		},
		{
			name: "Bullish - High Activity",
			whale: &WhaleWatch{
				Volume24h:  decimal.NewFromFloat(5000000), // < $10M
				TxCount24h: 15,                            // > 10 transactions
			},
			expectedDirection: SignalDirectionBullish,
		},
		{
			name: "Neutral - Low Activity",
			whale: &WhaleWatch{
				Volume24h:  decimal.NewFromFloat(1000000), // < $10M
				TxCount24h: 3,                             // < 10 transactions
			},
			expectedDirection: SignalDirectionNeutral,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			direction := analyzer.determineWhaleDirection(tc.whale)

			// Assert
			assert.Equal(t, tc.expectedDirection, direction)
		})
	}
}

// Helper function to create test analyzer
func createTestAnalyzer(t *testing.T) *OnChainAnalyzer {
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}

	// Use nil for blockchain clients since NewOnChainAnalyzer expects concrete types
	// In a real integration test, these would be actual client instances
	return NewOnChainAnalyzer(logger, mockRedis, nil, nil, nil)
}

// Benchmark tests
func BenchmarkOnChainAnalyzer_CalculateTokenScore(b *testing.B) {
	analyzer := createTestAnalyzer(&testing.T{})

	metrics := &OnChainMetrics{
		Volume24h:  decimal.NewFromFloat(10000000),
		Liquidity:  decimal.NewFromFloat(50000000),
		Holders:    10000,
		Volatility: decimal.NewFromFloat(0.10),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.calculateTokenScore(metrics)
	}
}

func TestOnChainAnalyzer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	analyzer := createTestAnalyzer(t)
	ctx := context.Background()

	// Note: In a real integration test, we would set up actual blockchain clients
	// For this test, we'll work with the analyzer's internal functionality

	// Start analyzer
	err := analyzer.Start(ctx)
	require.NoError(t, err)

	// Wait for initial scan
	time.Sleep(100 * time.Millisecond)

	// Test metrics calculation
	analyzer.calculateMetrics(ctx)

	// Test getting metrics
	metrics, err := analyzer.GetMetrics(ctx, "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1")
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	// Test whale activity
	whales, err := analyzer.GetWhaleActivity(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, whales) // Should have initialized whales

	// Test market signals
	_, err = analyzer.GetMarketSignals(ctx)
	assert.NoError(t, err)
	// Signals may be empty initially, that's ok

	// Cleanup
	analyzer.Stop()
}
