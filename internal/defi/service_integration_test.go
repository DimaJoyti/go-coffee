package defi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/blockchain"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/redis"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeFiServiceIntegration(t *testing.T) {
	// Create mock clients
	mockEth := blockchain.NewMockEthereumClient()
	mockBSC := blockchain.NewMockEthereumClient()
	mockPolygon := blockchain.NewMockEthereumClient()
	mockRedis := redis.NewMockRedisClient()

	// Create logger
	logger := logger.New("defi-test")

	// Create DeFi config
	defiConfig := config.DeFiConfig{
		UniswapV3Router:     "0xE592427A0AEce92De3Edee1F18E0157C05861564",
		AaveLendingPool:     "0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9",
		CompoundComptroller: "0x3d9819210A31b4961b30EF54bE2aeD79B9c9Cd3B",
		OneInchAPIKey:       "",
		ChainlinkEnabled:    true,
		ArbitrageEnabled:    true,
		YieldFarmingEnabled: true,
		TradingBotsEnabled:  true,
	}

	// Create service with the expected signature
	service := NewService(mockEth, mockBSC, mockPolygon, mockRedis, logger, defiConfig)

	// Test service creation
	assert.NotNil(t, service)
	assert.NotNil(t, service.ethClient)
	assert.NotNil(t, service.bscClient)
	assert.NotNil(t, service.polygonClient)
	assert.NotNil(t, service.cache)
	assert.NotNil(t, service.logger)

	// Test service start
	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)

	// Test token price retrieval
	t.Run("GetTokenPrice", func(t *testing.T) {
		req := &GetTokenPriceRequest{
			TokenAddress: "0x1234567890123456789012345678901234567890",
			Chain:        ChainEthereum,
		}

		resp, err := service.GetTokenPrice(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.TokenAddress, resp.Token.Address)
		assert.Equal(t, req.Chain, resp.Token.Chain)
		assert.True(t, resp.Price.GreaterThan(decimal.Zero))
	})

	// Test swap quote
	t.Run("GetSwapQuote", func(t *testing.T) {
		req := &GetSwapQuoteRequest{
			TokenIn:  "0x1234567890123456789012345678901234567890",
			TokenOut: "0x0987654321098765432109876543210987654321",
			AmountIn: decimal.NewFromFloat(100),
			Chain:    ChainEthereum,
			Slippage: decimal.NewFromFloat(0.01),
		}

		resp, err := service.GetSwapQuote(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Quote.ID)
		assert.True(t, resp.Quote.AmountOut.GreaterThan(decimal.Zero))
		assert.True(t, resp.Quote.ExpiresAt.After(time.Now()))
	})

	// Test liquidity pools
	t.Run("GetLiquidityPools", func(t *testing.T) {
		req := &GetLiquidityPoolsRequest{
			Chain:    ChainEthereum,
			Protocol: ProtocolTypeUniswap,
			MinTVL:   decimal.NewFromFloat(1000),
			Limit:    10,
		}

		resp, err := service.GetLiquidityPools(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.GreaterOrEqual(t, resp.Total, 0)
	})

	// Test arbitrage opportunities
	t.Run("GetArbitrageOpportunities", func(t *testing.T) {
		opportunities, err := service.GetArbitrageOpportunities(ctx)
		require.NoError(t, err)
		assert.NotNil(t, opportunities)
	})

	// Test yield opportunities
	t.Run("GetBestYieldOpportunities", func(t *testing.T) {
		opportunities, err := service.GetBestYieldOpportunities(ctx, 5)
		require.NoError(t, err)
		assert.NotNil(t, opportunities)
	})

	// Test trading bot creation
	t.Run("CreateTradingBot", func(t *testing.T) {
		config := TradingBotConfig{
			MaxPositionSize:   decimal.NewFromFloat(1000),
			MinProfitMargin:   decimal.NewFromFloat(0.01),
			MaxSlippage:       decimal.NewFromFloat(0.005),
			RiskTolerance:     RiskLevelMedium,
			AutoCompound:      true,
			MaxDailyTrades:    10,
			StopLossPercent:   decimal.NewFromFloat(0.05),
			TakeProfitPercent: decimal.NewFromFloat(0.15),
			ExecutionDelay:    time.Second,
		}

		bot, err := service.CreateTradingBot(ctx, "Test Bot", StrategyTypeArbitrage, config)
		require.NoError(t, err)
		assert.NotNil(t, bot)
		assert.Equal(t, "Test Bot", bot.Name)
		assert.Equal(t, StrategyTypeArbitrage, bot.Strategy)
		assert.Equal(t, BotStatusStopped, bot.Status)

		// Test bot start
		err = service.StartTradingBot(ctx, bot.ID)
		require.NoError(t, err)

		// Test bot status
		status := bot.GetStatus()
		assert.Equal(t, BotStatusActive, status)

		// Test bot performance
		performance, err := service.GetTradingBotPerformance(ctx, bot.ID)
		require.NoError(t, err)
		assert.NotNil(t, performance)

		// Test bot stop
		err = service.StopTradingBot(ctx, bot.ID)
		require.NoError(t, err)

		// Test bot deletion
		err = service.DeleteTradingBot(ctx, bot.ID)
		require.NoError(t, err)
	})

	// Test on-chain analysis
	t.Run("OnChainAnalysis", func(t *testing.T) {
		// Test market signals
		signals, err := service.GetMarketSignals(ctx)
		require.NoError(t, err)
		assert.NotNil(t, signals)

		// Test whale activity
		whales, err := service.GetWhaleActivity(ctx)
		require.NoError(t, err)
		assert.NotNil(t, whales)
	})

	// Test service stop
	service.Stop()
}

func TestDeFiServiceWithSolana(t *testing.T) {
	// Create mock clients
	mockEth := blockchain.NewMockEthereumClient()
	mockBSC := blockchain.NewMockEthereumClient()
	mockPolygon := blockchain.NewMockEthereumClient()
	mockSolana := blockchain.NewMockSolanaClient()
	mockRedis := redis.NewMockRedisClient()

	// Create logger
	logger := logger.New("defi-solana-test")

	// Create DeFi config
	defiConfig := config.DeFiConfig{
		UniswapV3Router:     "0xE592427A0AEce92De3Edee1F18E0157C05861564",
		AaveLendingPool:     "0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9",
		CompoundComptroller: "0x3d9819210A31b4961b30EF54bE2aeD79B9c9Cd3B",
		OneInchAPIKey:       "",
		ChainlinkEnabled:    true,
		ArbitrageEnabled:    true,
		YieldFarmingEnabled: true,
		TradingBotsEnabled:  true,
	}

	// Create Solana clients
	raydiumClient := NewRaydiumClient(logger)
	jupiterClient := NewJupiterClient(logger)

	// Create service with Solana support
	service := NewServiceWithSolana(
		mockEth, mockBSC, mockPolygon, mockSolana,
		raydiumClient, jupiterClient,
		mockRedis, logger, defiConfig,
	)

	// Test service creation
	assert.NotNil(t, service)
	assert.NotNil(t, service.solanaClient)
	assert.NotNil(t, service.raydiumClient)
	assert.NotNil(t, service.jupiterClient)

	// Test Solana swap quote
	ctx := context.Background()
	req := &GetSwapQuoteRequest{
		TokenIn:  "So11111111111111111111111111111111111111112", // SOL
		TokenOut: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		AmountIn: decimal.NewFromFloat(1),
		Chain:    ChainSolana,
		Slippage: decimal.NewFromFloat(0.01),
	}

	resp, err := service.GetSwapQuote(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, ChainSolana, resp.Quote.Chain)
}

func TestDeFiServiceErrorHandling(t *testing.T) {
	// Create mock clients with error behaviors
	mockEth := blockchain.NewMockEthereumClient()
	mockBSC := blockchain.NewMockEthereumClient()
	mockPolygon := blockchain.NewMockEthereumClient()
	mockRedis := redis.NewMockRedisClient()

	// Create logger
	logger := logger.New("defi-error-test")

	// Create DeFi config
	defiConfig := config.DeFiConfig{
		UniswapV3Router:     "0xE592427A0AEce92De3Edee1F18E0157C05861564",
		AaveLendingPool:     "0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9",
		CompoundComptroller: "0x3d9819210A31b4961b30EF54bE2aeD79B9c9Cd3B",
		OneInchAPIKey:       "",
		ChainlinkEnabled:    true,
		ArbitrageEnabled:    true,
		YieldFarmingEnabled: true,
		TradingBotsEnabled:  true,
	}

	// Create service
	service := NewService(mockEth, mockBSC, mockPolygon, mockRedis, logger, defiConfig)

	ctx := context.Background()

	// Test invalid chain
	t.Run("InvalidChain", func(t *testing.T) {
		req := &GetSwapQuoteRequest{
			TokenIn:  "0x1234567890123456789012345678901234567890",
			TokenOut: "0x0987654321098765432109876543210987654321",
			AmountIn: decimal.NewFromFloat(100),
			Chain:    Chain("invalid"),
		}

		_, err := service.GetSwapQuote(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported chain")
	})

	// Test non-existent trading bot
	t.Run("NonExistentBot", func(t *testing.T) {
		_, err := service.GetTradingBot(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trading bot not found")
	})

	// Test invalid protocol
	t.Run("InvalidProtocol", func(t *testing.T) {
		req := &GetLiquidityPoolsRequest{
			Chain:    ChainEthereum,
			Protocol: ProtocolType("invalid"),
		}

		_, err := service.GetLiquidityPools(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported protocol")
	})
}

func TestDeFiServiceConcurrency(t *testing.T) {
	// Create mock clients
	mockEth := blockchain.NewMockEthereumClient()
	mockBSC := blockchain.NewMockEthereumClient()
	mockPolygon := blockchain.NewMockEthereumClient()
	mockRedis := redis.NewMockRedisClient()

	// Create logger
	logger := logger.New("defi-concurrency-test")

	// Create DeFi config
	defiConfig := config.DeFiConfig{
		UniswapV3Router:     "0xE592427A0AEce92De3Edee1F18E0157C05861564",
		AaveLendingPool:     "0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9",
		CompoundComptroller: "0x3d9819210A31b4961b30EF54bE2aeD79B9c9Cd3B",
		OneInchAPIKey:       "",
		ChainlinkEnabled:    true,
		ArbitrageEnabled:    true,
		YieldFarmingEnabled: true,
		TradingBotsEnabled:  true,
	}

	// Create service
	service := NewService(mockEth, mockBSC, mockPolygon, mockRedis, logger, defiConfig)

	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)

	// Test concurrent operations
	t.Run("ConcurrentOperations", func(t *testing.T) {
		const numGoroutines = 10

		// Test concurrent price requests
		for i := 0; i < numGoroutines; i++ {
			go func() {
				req := &GetTokenPriceRequest{
					TokenAddress: "0x1234567890123456789012345678901234567890",
					Chain:        ChainEthereum,
				}
				_, err := service.GetTokenPrice(ctx, req)
				assert.NoError(t, err)
			}()
		}

		// Test concurrent trading bot creation
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				config := TradingBotConfig{
					MaxPositionSize: decimal.NewFromFloat(1000),
					MinProfitMargin: decimal.NewFromFloat(0.01),
					RiskTolerance:   RiskLevelMedium,
				}
				bot, err := service.CreateTradingBot(ctx, fmt.Sprintf("Bot-%d", index), StrategyTypeArbitrage, config)
				assert.NoError(t, err)
				assert.NotNil(t, bot)
			}(i)
		}

		// Give goroutines time to complete
		time.Sleep(100 * time.Millisecond)
	})

	service.Stop()
}
