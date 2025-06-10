package defi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeFiService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	logger := logger.New("integration-test")
	mockRedis := &MockRedisClient{}
	mockEth := &MockEthereumClient{}
	mockBSC := &MockEthereumClient{}
	mockPolygon := &MockEthereumClient{}

	defiConfig := config.DeFiConfig{
		OneInch: config.OneInchConfig{
			APIKey: "test-api-key",
		},
	}

	service := NewService(mockEth, mockBSC, mockPolygon, mockRedis, logger, defiConfig)

	ctx := context.Background()

	// Mock Redis operations
	mockRedis.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockRedis.On("Get", ctx, mock.Anything).Return("", nil).Maybe()

	// Act - Start service
	err := service.Start(ctx)
	require.NoError(t, err)

	// Test 1: Arbitrage Detection
	t.Run("ArbitrageDetection", func(t *testing.T) {
		opportunities, err := service.GetArbitrageOpportunities(ctx)
		assert.NoError(t, err)
		// May be empty initially, that's ok
		assert.NotNil(t, opportunities)
	})

	// Test 2: Yield Farming
	t.Run("YieldFarming", func(t *testing.T) {
		yields, err := service.GetBestYieldOpportunities(ctx, 5)
		assert.NoError(t, err)
		assert.NotNil(t, yields)

		// Test optimal strategy
		req := &OptimalStrategyRequest{
			InvestmentAmount: decimal.NewFromFloat(10000),
			RiskTolerance:    RiskLevelMedium,
			MinAPY:           decimal.NewFromFloat(0.05),
			Diversification:  true,
		}

		strategy, err := service.GetOptimalYieldStrategy(ctx, req)
		// May return nil if no opportunities meet criteria
		assert.NoError(t, err)
		if strategy != nil {
			assert.NotEmpty(t, strategy.Name)
		}
	})

	// Test 3: On-Chain Analytics
	t.Run("OnChainAnalytics", func(t *testing.T) {
		// Test metrics (may not exist initially)
		_, err := service.GetOnChainMetrics(ctx, "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1")
		// Error is expected if metrics don't exist yet
		assert.Error(t, err)

		// Test market signals
		signals, err := service.GetMarketSignals(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, signals)

		// Test whale activity
		whales, err := service.GetWhaleActivity(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, whales)
	})

	// Test 4: Trading Bots
	t.Run("TradingBots", func(t *testing.T) {
		// Create bot
		config := TradingBotConfig{
			MaxPositionSize:   decimal.NewFromFloat(1000),
			MinProfitMargin:   decimal.NewFromFloat(0.01),
			MaxSlippage:       decimal.NewFromFloat(0.005),
			RiskTolerance:     RiskLevelMedium,
			AutoCompound:      true,
			MaxDailyTrades:    5,
			StopLossPercent:   decimal.NewFromFloat(0.05),
			TakeProfitPercent: decimal.NewFromFloat(0.15),
			ExecutionDelay:    time.Second,
		}

		bot, err := service.CreateTradingBot(ctx, "Integration Test Bot", StrategyTypeArbitrage, config)
		require.NoError(t, err)
		assert.NotNil(t, bot)
		assert.NotEmpty(t, bot.ID)

		// Test getting bot
		retrievedBot, err := service.GetTradingBot(ctx, bot.ID)
		assert.NoError(t, err)
		assert.Equal(t, bot.ID, retrievedBot.ID)

		// Test getting all bots
		allBots, err := service.GetAllTradingBots(ctx)
		assert.NoError(t, err)
		assert.Len(t, allBots, 1)

		// Test starting bot
		err = service.StartTradingBot(ctx, bot.ID)
		assert.NoError(t, err)

		// Wait briefly
		time.Sleep(50 * time.Millisecond)

		// Test performance
		performance, err := service.GetTradingBotPerformance(ctx, bot.ID)
		assert.NoError(t, err)
		assert.NotNil(t, performance)

		// Test stopping bot
		err = service.StopTradingBot(ctx, bot.ID)
		assert.NoError(t, err)

		// Test deleting bot
		err = service.DeleteTradingBot(ctx, bot.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = service.GetTradingBot(ctx, bot.ID)
		assert.Error(t, err)
	})

	// Cleanup
	service.Stop()
}

func TestDeFiService_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// Arrange
	logger := logger.New("concurrent-test")
	mockRedis := &MockRedisClient{}
	mockEth := &MockEthereumClient{}
	mockBSC := &MockEthereumClient{}
	mockPolygon := &MockEthereumClient{}

	defiConfig := config.DeFiConfig{
		OneInch: config.OneInchConfig{
			APIKey: "test-api-key",
		},
	}

	service := NewService(mockEth, mockBSC, mockPolygon, mockRedis, logger, defiConfig)

	ctx := context.Background()

	// Mock Redis operations
	mockRedis.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockRedis.On("Get", ctx, mock.Anything).Return("", nil).Maybe()

	err := service.Start(ctx)
	require.NoError(t, err)
	defer service.Stop()

	// Test concurrent bot creation
	t.Run("ConcurrentBotCreation", func(t *testing.T) {
		const numBots = 5
		botChan := make(chan *TradingBot, numBots)
		errChan := make(chan error, numBots)

		config := TradingBotConfig{
			MaxPositionSize: decimal.NewFromFloat(1000),
			MinProfitMargin: decimal.NewFromFloat(0.01),
			RiskTolerance:   RiskLevelMedium,
		}

		// Create bots concurrently
		for i := 0; i < numBots; i++ {
			go func(id int) {
				bot, err := service.CreateTradingBot(ctx,
					fmt.Sprintf("Concurrent Bot %d", id),
					StrategyTypeArbitrage,
					config)
				if err != nil {
					errChan <- err
					return
				}
				botChan <- bot
			}(i)
		}

		// Collect results
		var bots []*TradingBot
		for i := 0; i < numBots; i++ {
			select {
			case bot := <-botChan:
				bots = append(bots, bot)
			case err := <-errChan:
				t.Errorf("Error creating bot: %v", err)
			case <-time.After(time.Second * 5):
				t.Error("Timeout waiting for bot creation")
			}
		}

		assert.Len(t, bots, numBots)

		// Verify all bots have unique IDs
		botIDs := make(map[string]bool)
		for _, bot := range bots {
			assert.False(t, botIDs[bot.ID], "Duplicate bot ID: %s", bot.ID)
			botIDs[bot.ID] = true
		}

		// Cleanup
		for _, bot := range bots {
			service.DeleteTradingBot(ctx, bot.ID)
		}
	})

	// Test concurrent API calls
	t.Run("ConcurrentAPICalls", func(t *testing.T) {
		const numCalls = 10
		resultChan := make(chan bool, numCalls)

		// Make concurrent calls to different endpoints
		for i := 0; i < numCalls; i++ {
			go func() {
				// Test different endpoints
				_, err1 := service.GetArbitrageOpportunities(ctx)
				_, err2 := service.GetBestYieldOpportunities(ctx, 5)
				_, err3 := service.GetMarketSignals(ctx)
				_, err4 := service.GetWhaleActivity(ctx)

				// All should succeed (or fail gracefully)
				success := err1 == nil || err2 == nil || err3 == nil || err4 == nil
				resultChan <- success
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < numCalls; i++ {
			select {
			case success := <-resultChan:
				if success {
					successCount++
				}
			case <-time.After(time.Second * 10):
				t.Error("Timeout waiting for API calls")
			}
		}

		// At least some calls should succeed
		assert.True(t, successCount > 0, "No API calls succeeded")
	})
}

func TestDeFiService_ErrorHandling(t *testing.T) {
	// Arrange
	logger := logger.New("error-test")
	mockRedis := &MockRedisClient{}
	mockEth := &MockEthereumClient{}
	mockBSC := &MockEthereumClient{}
	mockPolygon := &MockEthereumClient{}

	defiConfig := config.DeFiConfig{
		OneInch: config.OneInchConfig{
			APIKey: "test-api-key",
		},
	}

	service := NewService(mockEth, mockBSC, mockPolygon, mockRedis, logger, defiConfig)

	ctx := context.Background()

	// Test operations without starting service
	t.Run("OperationsWithoutStart", func(t *testing.T) {
		// These should work even without Start() being called
		opportunities, err := service.GetArbitrageOpportunities(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, opportunities)

		yields, err := service.GetBestYieldOpportunities(ctx, 5)
		assert.NoError(t, err)
		assert.NotNil(t, yields)
	})

	// Test invalid bot operations
	t.Run("InvalidBotOperations", func(t *testing.T) {
		// Try to get non-existent bot
		_, err := service.GetTradingBot(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Try to start non-existent bot
		err = service.StartTradingBot(ctx, "non-existent-id")
		assert.Error(t, err)

		// Try to stop non-existent bot
		err = service.StopTradingBot(ctx, "non-existent-id")
		assert.Error(t, err)

		// Try to delete non-existent bot
		err = service.DeleteTradingBot(ctx, "non-existent-id")
		assert.Error(t, err)

		// Try to get performance of non-existent bot
		_, err = service.GetTradingBotPerformance(ctx, "non-existent-id")
		assert.Error(t, err)
	})

	// Test invalid metrics requests
	t.Run("InvalidMetricsRequests", func(t *testing.T) {
		// Try to get metrics for non-existent token
		_, err := service.GetOnChainMetrics(ctx, "0xInvalidAddress")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestDeFiService_PerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Arrange
	logger := logger.New("performance-test")
	mockRedis := &MockRedisClient{}
	mockEth := &MockEthereumClient{}
	mockBSC := &MockEthereumClient{}
	mockPolygon := &MockEthereumClient{}

	defiConfig := config.DeFiConfig{
		OneInch: config.OneInchConfig{
			APIKey: "test-api-key",
		},
	}

	service := NewService(mockEth, mockBSC, mockPolygon, mockRedis, logger, defiConfig)

	ctx := context.Background()

	// Mock Redis operations
	mockRedis.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockRedis.On("Get", ctx, mock.Anything).Return("", nil).Maybe()

	err := service.Start(ctx)
	require.NoError(t, err)
	defer service.Stop()

	// Test high-frequency API calls
	t.Run("HighFrequencyAPICalls", func(t *testing.T) {
		const numCalls = 100
		const concurrency = 10

		start := time.Now()

		semaphore := make(chan struct{}, concurrency)
		done := make(chan bool, numCalls)

		for i := 0; i < numCalls; i++ {
			go func() {
				semaphore <- struct{}{}        // Acquire
				defer func() { <-semaphore }() // Release

				// Make API call
				_, err := service.GetArbitrageOpportunities(ctx)
				done <- err == nil
			}()
		}

		// Wait for all calls to complete
		successCount := 0
		for i := 0; i < numCalls; i++ {
			select {
			case success := <-done:
				if success {
					successCount++
				}
			case <-time.After(time.Second * 30):
				t.Error("Timeout waiting for API calls")
				return
			}
		}

		duration := time.Since(start)

		// Performance assertions
		assert.True(t, successCount > numCalls*0.8, "Success rate should be > 80%")
		assert.True(t, duration < time.Second*10, "Should complete within 10 seconds")

		t.Logf("Completed %d calls in %v (success rate: %.1f%%)",
			numCalls, duration, float64(successCount)/float64(numCalls)*100)
	})
}
