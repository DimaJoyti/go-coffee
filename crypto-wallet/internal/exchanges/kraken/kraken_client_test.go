package kraken

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
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

func TestNewKrakenClient(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultKrakenConfig()

	client := NewKrakenClient(logger, config)

	assert.NotNil(t, client)
	assert.Equal(t, config.BaseURL, client.config.BaseURL)
	assert.Equal(t, config.WebSocketURL, client.config.WebSocketURL)
	assert.False(t, client.isRunning)
	assert.NotNil(t, client.subscriptions)
	assert.NotNil(t, client.eventHandlers)
	assert.NotNil(t, client.rateLimiter)
}

func TestKrakenClient_StartStop(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultKrakenConfig()
	config.WebSocketConfig.Enabled = false // Disable WebSocket for testing

	client := NewKrakenClient(logger, config)
	ctx := context.Background()

	// Test starting (will fail due to no real API, but that's expected)
	err := client.Start(ctx)
	// We expect this to fail since we're not using real API credentials
	assert.Error(t, err)

	// Test stopping
	err = client.Stop()
	assert.NoError(t, err)
	assert.False(t, client.IsRunning())
}

func TestKrakenClient_StartDisabled(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultKrakenConfig()
	config.Enabled = false

	client := NewKrakenClient(logger, config)
	ctx := context.Background()

	err := client.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, client.IsRunning()) // Should remain false when disabled
}

func TestKrakenClient_Configuration(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultKrakenConfig()

	client := NewKrakenClient(logger, config)

	retrievedConfig := client.GetConfig()
	assert.Equal(t, config.BaseURL, retrievedConfig.BaseURL)
	assert.Equal(t, config.WebSocketURL, retrievedConfig.WebSocketURL)
	assert.Equal(t, config.Timeout, retrievedConfig.Timeout)
}

func TestKrakenClient_EventHandlers(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultKrakenConfig()

	client := NewKrakenClient(logger, config)

	// Add event handler
	handler := func(event *WSMessage) error {
		return nil
	}

	client.AddEventHandler(EventTypeTicker, handler)

	// Verify handler was added
	assert.Len(t, client.eventHandlers[EventTypeTicker], 1)
}

func TestRateLimiter(t *testing.T) {
	config := RateConfig{
		RequestsPerSecond: 2,
		BurstSize:         3,
		CounterDecay:      time.Second,
	}

	limiter := NewRateLimiter(config)

	// Should allow initial burst
	for i := 0; i < 3; i++ {
		err := limiter.Wait()
		assert.NoError(t, err)
	}

	// Should hit rate limit
	err := limiter.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

func TestKrakenSymbolMapping(t *testing.T) {
	mapping := GetKrakenSymbolMapping()

	assert.NotEmpty(t, mapping)
	assert.Equal(t, "XBTUSD", mapping["BTC/USD"])
	assert.Equal(t, "ETHUSD", mapping["ETH/USD"])
	assert.Equal(t, "ADAUSD", mapping["ADA/USD"])
}

func TestKrakenIntervalMapping(t *testing.T) {
	mapping := GetKrakenIntervalMapping()

	assert.NotEmpty(t, mapping)
	assert.Equal(t, "1", mapping["1m"])
	assert.Equal(t, "60", mapping["1h"])
	assert.Equal(t, "1440", mapping["1d"])
}

func TestKrakenOrderTypes(t *testing.T) {
	orderTypes := GetKrakenOrderTypes()

	assert.NotEmpty(t, orderTypes)
	assert.Contains(t, orderTypes, "market")
	assert.Contains(t, orderTypes, "limit")
	assert.Contains(t, orderTypes, "stop-loss")
}

func TestKrakenAssetInfo(t *testing.T) {
	assetInfo := GetKrakenAssetInfo()

	assert.NotEmpty(t, assetInfo)

	btcInfo, exists := assetInfo["XXBT"]
	assert.True(t, exists)
	assert.Equal(t, "BTC", btcInfo.Symbol)
	assert.Equal(t, "Bitcoin", btcInfo.Name)
	assert.Equal(t, 8, btcInfo.Decimals)

	ethInfo, exists := assetInfo["XETH"]
	assert.True(t, exists)
	assert.Equal(t, "ETH", ethInfo.Symbol)
	assert.Equal(t, "Ethereum", ethInfo.Name)
}

func TestKrakenTradingPairs(t *testing.T) {
	pairs := GetKrakenTradingPairs()

	assert.NotEmpty(t, pairs)

	// Find BTC/USD pair
	var btcPair *TradingPair
	for _, pair := range pairs {
		if pair.Symbol == "XBTUSD" {
			btcPair = &pair
			break
		}
	}

	require.NotNil(t, btcPair)
	assert.Equal(t, "XBT", btcPair.Base)
	assert.Equal(t, "USD", btcPair.Quote)
	assert.True(t, btcPair.Active)
}

func TestKrakenWebSocketChannels(t *testing.T) {
	channels := GetKrakenWebSocketChannels()

	assert.NotEmpty(t, channels)
	assert.Contains(t, channels, "ticker")
	assert.Contains(t, channels, "book")
	assert.Contains(t, channels, "trade")
	assert.Contains(t, channels, "ohlc")
}

func TestKrakenAPILimits(t *testing.T) {
	limits := GetKrakenAPILimits()

	assert.NotEmpty(t, limits)
	assert.Contains(t, limits, "public_endpoints")
	assert.Contains(t, limits, "private_endpoints")
	assert.Contains(t, limits, "websocket")
}

func TestKrakenFeatures(t *testing.T) {
	features := GetKrakenFeatures()

	assert.NotEmpty(t, features)
	assert.True(t, features["spot_trading"])
	assert.True(t, features["margin_trading"])
	assert.True(t, features["websocket_public"])
	assert.True(t, features["order_management"])
}

func TestValidateKrakenConfig(t *testing.T) {
	// Test valid config
	validConfig := GetDefaultKrakenConfig()
	err := ValidateKrakenConfig(validConfig)
	assert.NoError(t, err)

	// Test disabled config
	disabledConfig := GetDefaultKrakenConfig()
	disabledConfig.Enabled = false
	err = ValidateKrakenConfig(disabledConfig)
	assert.NoError(t, err)

	// Test invalid base URL
	invalidConfig := GetDefaultKrakenConfig()
	invalidConfig.BaseURL = ""
	err = ValidateKrakenConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base URL is required")

	// Test invalid timeout
	invalidConfig = GetDefaultKrakenConfig()
	invalidConfig.Timeout = 0
	err = ValidateKrakenConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout must be positive")

	// Test invalid rate limit
	invalidConfig = GetDefaultKrakenConfig()
	invalidConfig.RateLimit.RequestsPerSecond = 0
	err = ValidateKrakenConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requests per second must be positive")
}

func TestTickerParsing(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultKrakenConfig()
	client := NewKrakenClient(logger, config)

	// Mock response data
	response := map[string]interface{}{
		"result": map[string]interface{}{
			"XBTUSD": map[string]interface{}{
				"a": []interface{}{"50000.0", "1", "1.000"},
				"b": []interface{}{"49999.0", "2", "2.000"},
				"c": []interface{}{"50000.5", "0.1"},
				"v": []interface{}{"100.5", "200.5"},
				"h": []interface{}{"51000.0", "52000.0"},
				"l": []interface{}{"49000.0", "48000.0"},
				"o": "49500.0",
			},
		},
	}

	ticker := client.parseTickerResponse("XBTUSD", response)
	assert.NotNil(t, ticker)
	assert.Equal(t, "XBTUSD", ticker.Symbol)
	assert.Equal(t, "50000", ticker.Ask.String())
	assert.Equal(t, "49999", ticker.Bid.String())
	assert.Equal(t, "50000.5", ticker.Last.String())
	assert.Equal(t, "100.5", ticker.Volume.String())
	assert.Equal(t, "51000", ticker.High.String())
	assert.Equal(t, "49000", ticker.Low.String())
	assert.Equal(t, "49500", ticker.Open.String())
}

func TestOrderBookParsing(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultKrakenConfig()
	client := NewKrakenClient(logger, config)

	// Mock response data
	response := map[string]interface{}{
		"result": map[string]interface{}{
			"XBTUSD": map[string]interface{}{
				"asks": []interface{}{
					[]interface{}{"50000.0", "1.0", "1234567890"},
					[]interface{}{"50001.0", "2.0", "1234567891"},
				},
				"bids": []interface{}{
					[]interface{}{"49999.0", "1.5", "1234567890"},
					[]interface{}{"49998.0", "2.5", "1234567891"},
				},
			},
		},
	}

	orderBook := client.parseOrderBookResponse("XBTUSD", response)
	assert.NotNil(t, orderBook)
	assert.Equal(t, "XBTUSD", orderBook.Symbol)
	assert.Len(t, orderBook.Asks, 2)
	assert.Len(t, orderBook.Bids, 2)

	assert.Equal(t, "50000", orderBook.Asks[0].Price.String())
	assert.Equal(t, "1", orderBook.Asks[0].Volume.String())
	assert.Equal(t, "49999", orderBook.Bids[0].Price.String())
	assert.Equal(t, "1.5", orderBook.Bids[0].Volume.String())
}

func TestBalanceParsing(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultKrakenConfig()
	client := NewKrakenClient(logger, config)

	// Mock response data
	response := map[string]interface{}{
		"result": map[string]interface{}{
			"XXBT": "1.5000000000",
			"ZUSD": "10000.0000",
			"XETH": "10.2500000000",
		},
	}

	balances := client.parseBalancesResponse(response)
	assert.NotNil(t, balances)
	assert.Len(t, balances, 3)

	// Find BTC balance
	var btcBalance *Balance
	for _, balance := range balances {
		if balance.Currency == "XXBT" {
			btcBalance = &balance
			break
		}
	}

	require.NotNil(t, btcBalance)
	assert.Equal(t, "1.5", btcBalance.Available.String())
	assert.Equal(t, "1.5", btcBalance.Total.String())
}

func TestEventTypeString(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultKrakenConfig()
	client := NewKrakenClient(logger, config)

	assert.Equal(t, "ticker", client.getEventTypeString(EventTypeTicker))
	assert.Equal(t, "orderbook", client.getEventTypeString(EventTypeOrderBook))
	assert.Equal(t, "trade", client.getEventTypeString(EventTypeTrade))
	assert.Equal(t, "ohlcv", client.getEventTypeString(EventTypeOHLCV))
	assert.Equal(t, "error", client.getEventTypeString(EventTypeError))
	assert.Equal(t, "unknown", client.getEventTypeString(EventType(999)))
}

func TestGetDefaultKrakenConfig(t *testing.T) {
	config := GetDefaultKrakenConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, "https://api.kraken.com", config.BaseURL)
	assert.Equal(t, "wss://ws.kraken.com", config.WebSocketURL)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 1, config.RateLimit.RequestsPerSecond)
	assert.Equal(t, 5, config.RateLimit.BurstSize)
	assert.True(t, config.WebSocketConfig.Enabled)
	assert.Equal(t, 3, config.RetryConfig.MaxRetries)
	assert.Equal(t, 2.0, config.RetryConfig.BackoffFactor)
}
