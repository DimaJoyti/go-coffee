package main

import (
	"fmt"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/exchanges/kraken"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
)

func main() {
	fmt.Println("🐙 Kraken Exchange Integration Demo")
	fmt.Println("===================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create Kraken configuration
	krakenConfig := kraken.GetDefaultKrakenConfig()
	krakenConfig.APIKey = "your-kraken-api-key"       // Replace with actual API key
	krakenConfig.APISecret = "your-kraken-api-secret" // Replace with actual API secret
	krakenConfig.WebSocketConfig.Enabled = true
	krakenConfig.RateLimit.RequestsPerSecond = 1 // Conservative rate limiting

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Base URL: %s\n", krakenConfig.BaseURL)
	fmt.Printf("  WebSocket URL: %s\n", krakenConfig.WebSocketURL)
	fmt.Printf("  Rate Limit: %d req/sec\n", krakenConfig.RateLimit.RequestsPerSecond)
	fmt.Printf("  WebSocket Enabled: %v\n", krakenConfig.WebSocketConfig.Enabled)
	fmt.Printf("  Timeout: %v\n", krakenConfig.Timeout)
	fmt.Println()

	// Create Kraken client
	client := kraken.NewKrakenClient(logger, krakenConfig)

	// Add event handlers for WebSocket data
	addEventHandlers(client)

	fmt.Println("🏗️ Kraken client created successfully!")
	fmt.Println()

	// Demonstrate symbol mapping
	fmt.Println("💱 Symbol Mapping:")
	symbolMapping := kraken.GetKrakenSymbolMapping()
	fmt.Printf("  Standard -> Kraken Format (showing first 5):\n")
	count := 0
	for standard, krakenSymbol := range symbolMapping {
		if count >= 5 {
			fmt.Printf("  ... and %d more mappings\n", len(symbolMapping)-5)
			break
		}
		fmt.Printf("    %s -> %s\n", standard, krakenSymbol)
		count++
	}
	fmt.Println()

	// Demonstrate interval mapping
	fmt.Println("⏰ Interval Mapping:")
	intervalMapping := kraken.GetKrakenIntervalMapping()
	fmt.Printf("  Standard -> Kraken Format:\n")
	for standard, krakenInterval := range intervalMapping {
		fmt.Printf("    %s -> %s\n", standard, krakenInterval)
	}
	fmt.Println()

	// Demonstrate trading pairs
	fmt.Println("📊 Available Trading Pairs:")
	tradingPairs := kraken.GetKrakenTradingPairs()
	fmt.Printf("  Total Pairs: %d\n", len(tradingPairs))
	for i, pair := range tradingPairs {
		if i >= 3 {
			fmt.Printf("  ... and %d more pairs\n", len(tradingPairs)-3)
			break
		}
		fmt.Printf("  %d. %s (%s/%s)\n", i+1, pair.Symbol, pair.Base, pair.Quote)
		fmt.Printf("     Min Size: %s | Taker Fee: %s\n", pair.MinSize, pair.TakerFee)
	}
	fmt.Println()

	// Demonstrate asset information
	fmt.Println("🪙 Asset Information:")
	assetInfo := kraken.GetKrakenAssetInfo()
	for symbol, info := range assetInfo {
		fmt.Printf("  %s (%s): Decimals: %d, Trading Fee: %s\n", 
			symbol, info.Name, info.Decimals, info.TradingFee)
	}
	fmt.Println()

	// Demonstrate order types and sides
	fmt.Println("📋 Supported Order Types:")
	orderTypes := kraken.GetKrakenOrderTypes()
	fmt.Printf("  Order Types: %v\n", orderTypes)
	
	orderSides := kraken.GetKrakenOrderSides()
	fmt.Printf("  Order Sides: %v\n", orderSides)
	fmt.Println()

	// Demonstrate WebSocket channels
	fmt.Println("📡 WebSocket Channels:")
	wsChannels := kraken.GetKrakenWebSocketChannels()
	fmt.Printf("  Available Channels: %v\n", wsChannels)
	fmt.Println()

	// Demonstrate API limits
	fmt.Println("⚡ API Rate Limits:")
	apiLimits := kraken.GetKrakenAPILimits()
	for category, limits := range apiLimits {
		fmt.Printf("  %s:\n", category)
		if limitMap, ok := limits.(map[string]interface{}); ok {
			for key, value := range limitMap {
				fmt.Printf("    %s: %v\n", key, value)
			}
		}
	}
	fmt.Println()

	// Demonstrate features
	fmt.Println("🎯 Supported Features:")
	features := kraken.GetKrakenFeatures()
	enabledCount := 0
	disabledCount := 0
	
	for _, enabled := range features {
		if enabled {
			enabledCount++
		} else {
			disabledCount++
		}
	}
	
	fmt.Printf("  ✅ Enabled Features: %d\n", enabledCount)
	fmt.Printf("  ❌ Disabled Features: %d\n", disabledCount)
	fmt.Println()

	// Demonstrate mock API calls
	fmt.Println("🔄 Mock API Operations:")
	fmt.Println("  📈 Getting Ticker (BTC/USD): Ask: $50,000 | Bid: $49,999")
	fmt.Println("  📚 Getting Order Book (ETH/USD): 10 asks, 10 bids")
	fmt.Println("  💰 Getting Balances: BTC: 1.5, ETH: 10.25, USD: 10,000")
	fmt.Println("  📤 Placing Order: 0.1 BTC at $49,500 (Order ID: 12345)")
	fmt.Println()

	// Demonstrate WebSocket subscriptions
	fmt.Println("📡 WebSocket Subscription Examples:")
	fmt.Println("  📊 Ticker: Real-time price updates")
	fmt.Println("  📖 Order Book: Live order book changes")
	fmt.Println("  💱 Trades: New trade notifications")
	fmt.Println("  📈 OHLCV: Candlestick updates")
	fmt.Println()

	// Demonstrate configuration validation
	fmt.Println("✅ Configuration Validation:")
	if err := kraken.ValidateKrakenConfig(krakenConfig); err != nil {
		fmt.Printf("  ❌ Configuration Error: %v\n", err)
	} else {
		fmt.Printf("  ✅ Configuration is valid\n")
	}
	fmt.Println()

	// Show client status
	fmt.Println("📊 Client Status:")
	fmt.Printf("  Running: %v\n", client.IsRunning())
	fmt.Printf("  Subscriptions: %d\n", len(client.GetSubscriptions()))
	fmt.Printf("  Configuration Valid: %v\n", kraken.ValidateKrakenConfig(krakenConfig) == nil)
	fmt.Println()

	fmt.Println("🎉 Kraken Exchange integration demo completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Kraken API client configuration")
	fmt.Println("  ✅ Symbol and interval mapping")
	fmt.Println("  ✅ Trading pairs and asset information")
	fmt.Println("  ✅ Market data and trading API methods")
	fmt.Println("  ✅ WebSocket real-time data streaming")
	fmt.Println("  ✅ Rate limiting and error handling")
	fmt.Println("  ✅ Comprehensive configuration validation")
	fmt.Println()
	fmt.Println("Note: This demo shows the API without executing actual requests.")
	fmt.Println("Configure real API credentials to execute live operations.")
}

// addEventHandlers sets up event handlers for the Kraken client
func addEventHandlers(client *kraken.KrakenClient) {
	// Ticker event handler
	client.AddEventHandler(kraken.EventTypeTicker, func(event *kraken.WSMessage) error {
		fmt.Printf("🔔 Ticker Event: %s - %v\n", event.Symbol, event.Data)
		return nil
	})

	// Order book event handler
	client.AddEventHandler(kraken.EventTypeOrderBook, func(event *kraken.WSMessage) error {
		fmt.Printf("🔔 Order Book Event: %s - Updated\n", event.Symbol)
		return nil
	})

	// Trade event handler
	client.AddEventHandler(kraken.EventTypeTrade, func(event *kraken.WSMessage) error {
		fmt.Printf("🔔 Trade Event: %s - New trade\n", event.Symbol)
		return nil
	})

	// OHLCV event handler
	client.AddEventHandler(kraken.EventTypeOHLCV, func(event *kraken.WSMessage) error {
		fmt.Printf("🔔 OHLCV Event: %s - Candle update\n", event.Symbol)
		return nil
	})

	// Error handler
	client.AddEventHandler(kraken.EventTypeError, func(event *kraken.WSMessage) error {
		fmt.Printf("🔔 Error Event: %s\n", event.Error)
		return nil
	})
}
