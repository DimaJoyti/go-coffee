package aggregators

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLogger() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

func TestNewDEXAggregator(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultDEXAggregatorConfig()

	aggregator := NewDEXAggregator(logger, config)

	assert.NotNil(t, aggregator)
	assert.Equal(t, config.DefaultSlippage, aggregator.config.DefaultSlippage)
	assert.Equal(t, config.MaxSlippage, aggregator.config.MaxSlippage)
	assert.False(t, aggregator.isRunning)
	assert.NotNil(t, aggregator.healthStatus)
	assert.NotNil(t, aggregator.priceCache)
	assert.NotNil(t, aggregator.routingEngine)
}

func TestDEXAggregator_StartStop(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultDEXAggregatorConfig()

	aggregator := NewDEXAggregator(logger, config)
	ctx := context.Background()

	err := aggregator.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, aggregator.isRunning)

	err = aggregator.Stop()
	assert.NoError(t, err)
	assert.False(t, aggregator.isRunning)
}

func TestDEXAggregator_StartDisabled(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultDEXAggregatorConfig()
	config.Enabled = false

	aggregator := NewDEXAggregator(logger, config)
	ctx := context.Background()

	err := aggregator.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, aggregator.isRunning) // Should remain false when disabled
}

func TestDEXAggregator_GetAggregatorStatus(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultDEXAggregatorConfig()

	aggregator := NewDEXAggregator(logger, config)

	status := aggregator.GetAggregatorStatus()
	assert.NotNil(t, status)
	assert.Contains(t, status, "is_running")
	assert.Contains(t, status, "last_update")
	assert.Contains(t, status, "cache_size")
	assert.Contains(t, status, "health_status")
	assert.Contains(t, status, "aggregators")

	assert.False(t, status["is_running"].(bool))
	assert.Equal(t, 0, status["cache_size"].(int))
}

func TestDEXAggregator_CacheOperations(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultDEXAggregatorConfig()

	aggregator := NewDEXAggregator(logger, config)

	// Create test quote request
	req := &QuoteRequest{
		TokenIn: Token{
			Address:  "0xA0b86a33E6441E6C8C07C4c4c8e8B0E8E8E8E8E8",
			Symbol:   "USDC",
			Decimals: 6,
			Chain:    "ethereum",
		},
		TokenOut: Token{
			Address:  "0x0000000000000000000000000000000000000000",
			Symbol:   "ETH",
			Decimals: 18,
			Chain:    "ethereum",
		},
		AmountIn: decimal.NewFromFloat(1000),
		Chain:    "ethereum",
		Slippage: decimal.NewFromFloat(0.01),
	}

	// Generate cache key
	cacheKey := aggregator.generateCacheKey(req)
	assert.NotEmpty(t, cacheKey)

	// Test cache miss
	cached := aggregator.getCachedQuote(cacheKey)
	assert.Nil(t, cached)

	// Create and cache quote
	quote := &AggregatedQuote{
		ID:        "test-quote-1",
		TokenIn:   req.TokenIn,
		TokenOut:  req.TokenOut,
		AmountIn:  req.AmountIn,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	aggregator.cacheQuote(cacheKey, quote)

	// Test cache hit
	cached = aggregator.getCachedQuote(cacheKey)
	assert.NotNil(t, cached)
	assert.Equal(t, quote.ID, cached.ID)
}

func TestRoutingEngine_SelectBestQuote(t *testing.T) {
	logger := createTestLogger()
	config := RoutingConfig{
		Strategy: RoutingStrategyBestPrice,
	}

	engine := NewRoutingEngine(logger, config)

	// Create test quotes
	quotes := []*SwapQuote{
		{
			ID:        "quote-1",
			Aggregator: "1inch",
			AmountOut: decimal.NewFromFloat(1.5),
			GasCost:   decimal.NewFromFloat(0.01),
		},
		{
			ID:        "quote-2",
			Aggregator: "paraswap",
			AmountOut: decimal.NewFromFloat(1.6), // Best price
			GasCost:   decimal.NewFromFloat(0.02),
		},
		{
			ID:        "quote-3",
			Aggregator: "0x",
			AmountOut: decimal.NewFromFloat(1.4),
			GasCost:   decimal.NewFromFloat(0.005), // Lowest gas
		},
	}

	req := &QuoteRequest{
		TokenIn:  Token{Symbol: "USDC"},
		TokenOut: Token{Symbol: "ETH"},
		AmountIn: decimal.NewFromFloat(1000),
	}

	// Test best price strategy
	bestQuote := engine.SelectBestQuote(quotes, req)
	assert.NotNil(t, bestQuote)
	assert.Equal(t, "quote-2", bestQuote.ID)
	assert.Equal(t, "paraswap", bestQuote.Aggregator)

	// Test lowest gas strategy
	engine.config.Strategy = RoutingStrategyLowestGas
	bestQuote = engine.SelectBestQuote(quotes, req)
	assert.Equal(t, "quote-3", bestQuote.ID)
	assert.Equal(t, "0x", bestQuote.Aggregator)
}

func TestRoutingEngine_GenerateRecommendation(t *testing.T) {
	logger := createTestLogger()
	config := RoutingConfig{
		Strategy: RoutingStrategyBestPrice,
	}

	engine := NewRoutingEngine(logger, config)

	quotes := []*SwapQuote{
		{
			ID:        "quote-1",
			Aggregator: "1inch",
			AmountOut: decimal.NewFromFloat(1.5),
			GasCost:   decimal.NewFromFloat(0.01),
			PriceImpact: decimal.NewFromFloat(0.01),
		},
		{
			ID:        "quote-2",
			Aggregator: "paraswap",
			AmountOut: decimal.NewFromFloat(1.6),
			GasCost:   decimal.NewFromFloat(0.02),
			PriceImpact: decimal.NewFromFloat(0.015),
		},
	}

	bestQuote := quotes[1] // paraswap quote

	req := &QuoteRequest{
		TokenIn:  Token{Symbol: "USDC"},
		TokenOut: Token{Symbol: "ETH"},
		AmountIn: decimal.NewFromFloat(1000),
	}

	recommendation := engine.GenerateRecommendation(quotes, bestQuote, req)

	assert.NotNil(t, recommendation)
	assert.Equal(t, RoutingStrategyBestPrice, recommendation.Strategy)
	assert.NotEmpty(t, recommendation.Reason)
	assert.True(t, recommendation.Confidence.GreaterThan(decimal.Zero))
	assert.NotEmpty(t, recommendation.RiskLevel)
	assert.Len(t, recommendation.Alternatives, 1)
}

func TestGetDefaultDEXAggregatorConfig(t *testing.T) {
	config := GetDefaultDEXAggregatorConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, decimal.NewFromFloat(0.01), config.DefaultSlippage)
	assert.Equal(t, decimal.NewFromFloat(0.05), config.MaxSlippage)
	assert.Equal(t, 30*time.Second, config.QuoteTimeout)
	assert.Equal(t, 5*time.Minute, config.CacheTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Minute, config.HealthCheckInterval)

	// Check aggregator configs
	assert.Contains(t, config.Aggregators, "1inch")
	assert.Contains(t, config.Aggregators, "paraswap")
	assert.Contains(t, config.Aggregators, "0x")
	assert.Contains(t, config.Aggregators, "matcha")

	oneInchConfig := config.Aggregators["1inch"]
	assert.True(t, oneInchConfig.Enabled)
	assert.Equal(t, "https://api.1inch.io", oneInchConfig.BaseURL)
	assert.Equal(t, 1, oneInchConfig.Priority)
	assert.Contains(t, oneInchConfig.Chains, "ethereum")

	// Check routing config
	assert.Equal(t, RoutingStrategyBestPrice, config.RoutingConfig.Strategy)
	assert.True(t, config.RoutingConfig.ParallelQuotes)
	assert.True(t, config.RoutingConfig.FallbackEnabled)
}

func TestGetSupportedChains(t *testing.T) {
	chains := GetSupportedChains()

	assert.NotEmpty(t, chains)
	assert.Contains(t, chains, "ethereum")
	assert.Contains(t, chains, "bsc")
	assert.Contains(t, chains, "polygon")
	assert.Contains(t, chains, "optimism")
	assert.Contains(t, chains, "arbitrum")
}

func TestGetChainInfo(t *testing.T) {
	chainInfo := GetChainInfo()

	assert.NotEmpty(t, chainInfo)
	
	ethInfo, exists := chainInfo["ethereum"]
	assert.True(t, exists)
	assert.Equal(t, 1, ethInfo.ChainID)
	assert.Equal(t, "Ethereum", ethInfo.Name)
	assert.Equal(t, "ETH", ethInfo.Symbol)
	assert.Contains(t, ethInfo.Aggregators, "1inch")

	bscInfo, exists := chainInfo["bsc"]
	assert.True(t, exists)
	assert.Equal(t, 56, bscInfo.ChainID)
	assert.Equal(t, "Binance Smart Chain", bscInfo.Name)
	assert.Equal(t, "BNB", bscInfo.Symbol)
}

func TestGetCommonTokens(t *testing.T) {
	tokens := GetCommonTokens()

	assert.NotEmpty(t, tokens)
	
	ethTokens, exists := tokens["ethereum"]
	assert.True(t, exists)
	assert.NotEmpty(t, ethTokens)

	// Find ETH token
	var ethToken *Token
	for _, token := range ethTokens {
		if token.Symbol == "ETH" {
			ethToken = &token
			break
		}
	}

	require.NotNil(t, ethToken)
	assert.Equal(t, "Ethereum", ethToken.Name)
	assert.Equal(t, 18, ethToken.Decimals)
	assert.Equal(t, "ethereum", ethToken.Chain)
}

func TestValidateDEXAggregatorConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultDEXAggregatorConfig()
	err := ValidateDEXAggregatorConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultDEXAggregatorConfig()
	disabledConfig.Enabled = false
	err = ValidateDEXAggregatorConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid slippage
	invalidConfig := GetDefaultDEXAggregatorConfig()
	invalidConfig.DefaultSlippage = decimal.NewFromFloat(-0.1)
	err = ValidateDEXAggregatorConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default slippage must be between 0 and 1")

	// Test invalid max slippage
	invalidConfig = GetDefaultDEXAggregatorConfig()
	invalidConfig.MaxSlippage = decimal.NewFromFloat(0.005) // Less than default
	err = ValidateDEXAggregatorConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max slippage must be greater than or equal to default slippage")

	// Test invalid timeout
	invalidConfig = GetDefaultDEXAggregatorConfig()
	invalidConfig.QuoteTimeout = 0
	err = ValidateDEXAggregatorConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quote timeout must be positive")
}

func TestGetRoutingStrategies(t *testing.T) {
	strategies := GetRoutingStrategies()

	assert.NotEmpty(t, strategies)
	assert.Contains(t, strategies, RoutingStrategyBestPrice)
	assert.Contains(t, strategies, RoutingStrategyLowestGas)
	assert.Contains(t, strategies, RoutingStrategyBestValue)
	assert.Contains(t, strategies, RoutingStrategyFastest)
	assert.Contains(t, strategies, RoutingStrategyMostLiquid)
	assert.Contains(t, strategies, RoutingStrategyBalanced)
}

func TestGetStrategyDescription(t *testing.T) {
	descriptions := GetStrategyDescription()

	assert.NotEmpty(t, descriptions)
	assert.Contains(t, descriptions, RoutingStrategyBestPrice)
	assert.NotEmpty(t, descriptions[RoutingStrategyBestPrice])
	assert.Contains(t, descriptions[RoutingStrategyBestPrice], "output token amount")
}
