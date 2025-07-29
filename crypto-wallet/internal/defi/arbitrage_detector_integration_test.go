package defi

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/blockchain"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArbitrageDetector_RealAPI tests the arbitrage detector with real APIs
func TestArbitrageDetector_RealAPI(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if no API keys provided
	oneInchAPIKey := os.Getenv("ONEINCH_API_KEY")
	if oneInchAPIKey == "" {
		t.Skip("Skipping integration test: ONEINCH_API_KEY not set")
	}

	// Setup
	logger := logger.New("integration-test")

	// Create real Redis client (or mock if Redis not available)
	var redisClient redis.Client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		redisConfig := &redis.Config{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		}
		var err error
		redisClient, err = redis.NewClient(redisConfig)
		if err != nil {
			t.Logf("Failed to create Redis client, using mock: %v", err)
			redisClient = &MockRedisClient{}
		}
	} else {
		// Use mock Redis for testing
		redisClient = &MockRedisClient{}
	}

	// Create real blockchain clients
	ethRPC := os.Getenv("ETHEREUM_RPC_URL")
	if ethRPC == "" {
		ethRPC = "https://eth-mainnet.alchemyapi.io/v2/demo" // Demo endpoint
	}

	ethConfig := config.BlockchainNetworkConfig{
		Network:            "mainnet",
		RPCURL:             ethRPC,
		ChainID:            1,
		GasLimit:           21000,
		ConfirmationBlocks: 12,
	}

	ethClient, err := blockchain.NewEthereumClient(ethConfig, logger)
	require.NoError(t, err, "Failed to create Ethereum client")

	// Create real protocol clients
	uniswapClient := NewUniswapClient(ethClient, logger)
	oneInchClient := NewOneInchClient(oneInchAPIKey, logger)
	chainlinkClient := NewChainlinkClient(ethClient, logger)

	// Create real price providers
	priceProviders := []PriceProvider{
		NewUniswapPriceProvider(uniswapClient),
		NewOneInchPriceProvider(oneInchClient),
		NewChainlinkPriceProvider(chainlinkClient),
	}

	// Create arbitrage detector
	detector := NewArbitrageDetector(logger, redisClient, priceProviders)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with real tokens
	testTokens := []Token{
		{
			Address:  "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
			Symbol:   "WETH",
			Name:     "Wrapped Ether",
			Decimals: 18,
			Chain:    ChainEthereum,
		},
		{
			Address:  "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1", // USDC
			Symbol:   "USDC",
			Name:     "USD Coin",
			Decimals: 6,
			Chain:    ChainEthereum,
		},
	}

	for _, token := range testTokens {
		t.Run("Token_"+token.Symbol, func(t *testing.T) {
			// Test price provider health
			for i, provider := range priceProviders {
				t.Run("Provider_Health_"+provider.GetExchangeInfo().Name, func(t *testing.T) {
					healthy := provider.IsHealthy(ctx)
					t.Logf("Provider %d (%s) healthy: %v", i, provider.GetExchangeInfo().Name, healthy)

					if healthy {
						// If provider is healthy, try to get price
						price, err := provider.GetPrice(ctx, token)
						if err != nil {
							t.Logf("Warning: Failed to get price from %s: %v", provider.GetExchangeInfo().Name, err)
						} else {
							t.Logf("Price from %s for %s: %s", provider.GetExchangeInfo().Name, token.Symbol, price.String())
							assert.True(t, price.GreaterThan(decimal.Zero), "Price should be positive")
						}
					}
				})
			}

			// Test arbitrage detection
			t.Run("Arbitrage_Detection", func(t *testing.T) {
				opportunities, err := detector.DetectArbitrageForToken(ctx, token)

				// Don't require success since real markets might not have arbitrage opportunities
				if err != nil {
					t.Logf("Arbitrage detection failed for %s: %v", token.Symbol, err)
					return
				}

				t.Logf("Found %d arbitrage opportunities for %s", len(opportunities), token.Symbol)

				for i, opp := range opportunities {
					t.Logf("Opportunity %d:", i+1)
					t.Logf("  Source: %s (Price: %s)", opp.SourceExchange.Name, opp.SourcePrice.String())
					t.Logf("  Target: %s (Price: %s)", opp.TargetExchange.Name, opp.TargetPrice.String())
					t.Logf("  Profit Margin: %s%%", opp.ProfitMargin.Mul(decimal.NewFromFloat(100)).String())
					t.Logf("  Net Profit: %s", opp.NetProfit.String())
					t.Logf("  Confidence: %s", opp.Confidence.String())
					t.Logf("  Risk: %s", opp.Risk)

					// Validate opportunity structure
					assert.NotEmpty(t, opp.ID)
					assert.Equal(t, token.Address, opp.Token.Address)
					assert.True(t, opp.SourcePrice.GreaterThan(decimal.Zero))
					assert.True(t, opp.TargetPrice.GreaterThan(decimal.Zero))
					assert.True(t, opp.Volume.GreaterThan(decimal.Zero))
					assert.True(t, opp.Confidence.GreaterThanOrEqual(decimal.Zero))
					assert.True(t, opp.Confidence.LessThanOrEqual(decimal.NewFromFloat(1.0)))
					assert.Contains(t, []RiskLevel{RiskLevelLow, RiskLevelMedium, RiskLevelHigh}, opp.Risk)
					assert.Equal(t, OpportunityStatusDetected, opp.Status)
				}
			})
		})
	}
}

// TestPriceProviders_RealAPI tests individual price providers with real APIs
func TestPriceProviders_RealAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logger.New("price-provider-test")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Test token
	wethToken := Token{
		Address:  "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		Symbol:   "WETH",
		Name:     "Wrapped Ether",
		Decimals: 18,
		Chain:    ChainEthereum,
	}

	t.Run("Chainlink_Provider", func(t *testing.T) {
		ethRPC := os.Getenv("ETHEREUM_RPC_URL")
		if ethRPC == "" {
			ethRPC = "https://eth-mainnet.alchemyapi.io/v2/demo"
		}

		ethConfig := config.BlockchainNetworkConfig{
			Network:            "mainnet",
			RPCURL:             ethRPC,
			ChainID:            1,
			GasLimit:           21000,
			ConfirmationBlocks: 12,
		}

		ethClient, err := blockchain.NewEthereumClient(ethConfig, logger)
		if err != nil {
			t.Skipf("Failed to create Ethereum client: %v", err)
		}

		chainlinkClient := NewChainlinkClient(ethClient, logger)
		provider := NewChainlinkPriceProvider(chainlinkClient)

		// Test health check
		healthy := provider.IsHealthy(ctx)
		t.Logf("Chainlink provider healthy: %v", healthy)

		if healthy {
			// Test price retrieval
			price, err := provider.GetPrice(ctx, wethToken)
			if err != nil {
				t.Logf("Failed to get price from Chainlink: %v", err)
			} else {
				t.Logf("WETH price from Chainlink: %s", price.String())
				assert.True(t, price.GreaterThan(decimal.Zero))
			}
		}

		// Test exchange info
		exchangeInfo := provider.GetExchangeInfo()
		assert.Equal(t, "chainlink-oracle", exchangeInfo.ID)
		assert.Equal(t, "Chainlink Price Feeds", exchangeInfo.Name)
	})

	t.Run("OneInch_Provider", func(t *testing.T) {
		oneInchAPIKey := os.Getenv("ONEINCH_API_KEY")
		if oneInchAPIKey == "" {
			t.Skip("Skipping 1inch test: ONEINCH_API_KEY not set")
		}

		oneInchClient := NewOneInchClient(oneInchAPIKey, logger)
		provider := NewOneInchPriceProvider(oneInchClient)

		// Test health check
		healthy := provider.IsHealthy(ctx)
		t.Logf("1inch provider healthy: %v", healthy)

		if healthy {
			// Test price retrieval
			price, err := provider.GetPrice(ctx, wethToken)
			if err != nil {
				t.Logf("Failed to get price from 1inch: %v", err)
			} else {
				t.Logf("WETH price from 1inch: %s", price.String())
				assert.True(t, price.GreaterThan(decimal.Zero))
			}
		}

		// Test exchange info
		exchangeInfo := provider.GetExchangeInfo()
		assert.Equal(t, "1inch-aggregator", exchangeInfo.ID)
		assert.Equal(t, "1inch Aggregator", exchangeInfo.Name)
	})
}

// TestArbitrageDetector_Performance tests performance with real APIs
func TestArbitrageDetector_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// This test measures how long it takes to detect arbitrage opportunities
	logger := logger.New("performance-test")
	mockRedis := &MockRedisClient{}

	// Use mock providers for performance testing to avoid API rate limits
	mockUniswap := NewMockPriceProvider("uniswap-v3", "Uniswap V3")
	mockOneInch := NewMockPriceProvider("1inch-aggregator", "1inch Aggregator")

	priceProviders := []PriceProvider{mockUniswap, mockOneInch}
	detector := NewArbitrageDetector(logger, mockRedis, priceProviders)

	ctx := context.Background()
	testToken := Token{
		Address:  "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
		Symbol:   "WETH",
		Decimals: 18,
		Chain:    ChainEthereum,
	}

	// Mock price responses
	mockUniswap.On("IsHealthy", ctx).Return(true)
	mockUniswap.On("GetPrice", ctx, testToken).Return(decimal.NewFromFloat(2500.0), nil)

	mockOneInch.On("IsHealthy", ctx).Return(true)
	mockOneInch.On("GetPrice", ctx, testToken).Return(decimal.NewFromFloat(2530.0), nil)

	// Measure performance
	start := time.Now()
	opportunities, err := detector.DetectArbitrageForToken(ctx, testToken)
	duration := time.Since(start)

	require.NoError(t, err)
	t.Logf("Arbitrage detection took: %v", duration)
	t.Logf("Found %d opportunities", len(opportunities))

	// Performance should be reasonable (< 5 seconds for mock providers)
	assert.Less(t, duration, 5*time.Second, "Arbitrage detection should complete within 5 seconds")

	mockUniswap.AssertExpectations(t)
	mockOneInch.AssertExpectations(t)
}
