package kraken

import (
	"fmt"
	"time"
)

// GetDefaultKrakenConfig returns default Kraken configuration
func GetDefaultKrakenConfig() KrakenConfig {
	return KrakenConfig{
		Enabled:      true,
		APIKey:       "", // Must be provided by user
		APISecret:    "", // Must be provided by user
		BaseURL:      "https://api.kraken.com",
		WebSocketURL: "wss://ws.kraken.com",
		Timeout:      30 * time.Second,
		RateLimit: RateConfig{
			RequestsPerSecond: 1, // Kraken allows 1 request per second for most endpoints
			BurstSize:         5, // Allow burst of 5 requests
			CounterDecay:      60 * time.Second,
		},
		RetryConfig: RetryConfig{
			MaxRetries:    3,
			InitialDelay:  1 * time.Second,
			MaxDelay:      30 * time.Second,
			BackoffFactor: 2.0,
		},
		WebSocketConfig: WSConfig{
			Enabled:           true,
			ReconnectInterval: 30 * time.Second,
			PingInterval:      30 * time.Second,
			ReadTimeout:       60 * time.Second,
			WriteTimeout:      10 * time.Second,
			BufferSize:        1024,
		},
	}
}

// ValidateKrakenConfig validates Kraken configuration
func ValidateKrakenConfig(config KrakenConfig) error {
	if !config.Enabled {
		return nil // Skip validation if disabled
	}

	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if config.WebSocketConfig.Enabled && config.WebSocketURL == "" {
		return fmt.Errorf("WebSocket URL is required when WebSocket is enabled")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if config.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("requests per second must be positive")
	}

	if config.RateLimit.BurstSize <= 0 {
		return fmt.Errorf("burst size must be positive")
	}

	if config.RetryConfig.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if config.RetryConfig.BackoffFactor <= 1.0 {
		return fmt.Errorf("backoff factor must be greater than 1.0")
	}

	return nil
}

// GetKrakenSymbolMapping returns symbol mapping for Kraken
func GetKrakenSymbolMapping() map[string]string {
	return map[string]string{
		"BTC/USD":   "XBTUSD",
		"ETH/USD":   "ETHUSD",
		"BTC/EUR":   "XBTEUR",
		"ETH/EUR":   "ETHEUR",
		"ADA/USD":   "ADAUSD",
		"DOT/USD":   "DOTUSD",
		"LINK/USD":  "LINKUSD",
		"UNI/USD":   "UNIUSD",
		"AAVE/USD":  "AAVEUSD",
		"SOL/USD":   "SOLUSD",
		"MATIC/USD": "MATICUSD",
		"AVAX/USD":  "AVAXUSD",
		"ATOM/USD":  "ATOMUSD",
		"XRP/USD":   "XRPUSD",
		"LTC/USD":   "LTCUSD",
		"BCH/USD":   "BCHUSD",
		"XLM/USD":   "XLMUSD",
		"EOS/USD":   "EOSUSD",
		"TRX/USD":   "TRXUSD",
		"XTZ/USD":   "XTZUSD",
	}
}

// GetKrakenIntervalMapping returns interval mapping for Kraken
func GetKrakenIntervalMapping() map[string]string {
	return map[string]string{
		"1m":  "1",
		"5m":  "5",
		"15m": "15",
		"30m": "30",
		"1h":  "60",
		"4h":  "240",
		"1d":  "1440",
		"1w":  "10080",
		"2w":  "21600",
	}
}

// GetKrakenOrderTypes returns supported order types
func GetKrakenOrderTypes() []string {
	return []string{
		"market",
		"limit",
		"stop-loss",
		"take-profit",
		"stop-loss-limit",
		"take-profit-limit",
		"settle-position",
	}
}

// GetKrakenOrderSides returns supported order sides
func GetKrakenOrderSides() []string {
	return []string{
		"buy",
		"sell",
	}
}

// GetKrakenAssetInfo returns asset information
func GetKrakenAssetInfo() map[string]AssetInfo {
	return map[string]AssetInfo{
		"XXBT": {
			Symbol:      "BTC",
			Name:        "Bitcoin",
			Decimals:    8,
			MinAmount:   "0.0001",
			MaxAmount:   "1000000",
			TradingFee:  "0.0026",
			WithdrawFee: "0.0005",
		},
		"XETH": {
			Symbol:      "ETH",
			Name:        "Ethereum",
			Decimals:    8,
			MinAmount:   "0.001",
			MaxAmount:   "100000",
			TradingFee:  "0.0026",
			WithdrawFee: "0.005",
		},
		"ZUSD": {
			Symbol:      "USD",
			Name:        "US Dollar",
			Decimals:    4,
			MinAmount:   "1",
			MaxAmount:   "10000000",
			TradingFee:  "0.0026",
			WithdrawFee: "5",
		},
		"ZEUR": {
			Symbol:      "EUR",
			Name:        "Euro",
			Decimals:    4,
			MinAmount:   "1",
			MaxAmount:   "10000000",
			TradingFee:  "0.0026",
			WithdrawFee: "5",
		},
	}
}

// AssetInfo holds information about an asset
type AssetInfo struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Decimals    int    `json:"decimals"`
	MinAmount   string `json:"min_amount"`
	MaxAmount   string `json:"max_amount"`
	TradingFee  string `json:"trading_fee"`
	WithdrawFee string `json:"withdraw_fee"`
}

// GetKrakenTradingPairs returns available trading pairs
func GetKrakenTradingPairs() []TradingPair {
	return []TradingPair{
		{
			Symbol:    "XBTUSD",
			Base:      "XBT",
			Quote:     "USD",
			MinSize:   "0.0001",
			MaxSize:   "1000000",
			SizeStep:  "0.0001",
			MinPrice:  "0.1",
			MaxPrice:  "1000000",
			PriceStep: "0.1",
			TakerFee:  "0.0026",
			MakerFee:  "0.0016",
			Active:    true,
		},
		{
			Symbol:    "ETHUSD",
			Base:      "ETH",
			Quote:     "USD",
			MinSize:   "0.001",
			MaxSize:   "100000",
			SizeStep:  "0.001",
			MinPrice:  "0.01",
			MaxPrice:  "100000",
			PriceStep: "0.01",
			TakerFee:  "0.0026",
			MakerFee:  "0.0016",
			Active:    true,
		},
		{
			Symbol:    "ADAUSD",
			Base:      "ADA",
			Quote:     "USD",
			MinSize:   "1",
			MaxSize:   "10000000",
			SizeStep:  "1",
			MinPrice:  "0.0001",
			MaxPrice:  "1000",
			PriceStep: "0.0001",
			TakerFee:  "0.0026",
			MakerFee:  "0.0016",
			Active:    true,
		},
		{
			Symbol:    "DOTUSD",
			Base:      "DOT",
			Quote:     "USD",
			MinSize:   "0.1",
			MaxSize:   "1000000",
			SizeStep:  "0.1",
			MinPrice:  "0.001",
			MaxPrice:  "10000",
			PriceStep: "0.001",
			TakerFee:  "0.0026",
			MakerFee:  "0.0016",
			Active:    true,
		},
	}
}

// TradingPair holds information about a trading pair
type TradingPair struct {
	Symbol    string `json:"symbol"`
	Base      string `json:"base"`
	Quote     string `json:"quote"`
	MinSize   string `json:"min_size"`
	MaxSize   string `json:"max_size"`
	SizeStep  string `json:"size_step"`
	MinPrice  string `json:"min_price"`
	MaxPrice  string `json:"max_price"`
	PriceStep string `json:"price_step"`
	TakerFee  string `json:"taker_fee"`
	MakerFee  string `json:"maker_fee"`
	Active    bool   `json:"active"`
}

// GetKrakenWebSocketChannels returns available WebSocket channels
func GetKrakenWebSocketChannels() []string {
	return []string{
		"ticker",
		"ohlc",
		"trade",
		"book",
		"spread",
		"ownTrades",
		"openOrders",
	}
}

// GetKrakenAPILimits returns API rate limits
func GetKrakenAPILimits() map[string]interface{} {
	return map[string]interface{}{
		"public_endpoints": map[string]interface{}{
			"requests_per_second": 1,
			"burst_size":          5,
			"counter_decay":       "60s",
		},
		"private_endpoints": map[string]interface{}{
			"requests_per_second": 1,
			"burst_size":          3,
			"counter_decay":       "60s",
			"order_rate_limit":    "60 orders per minute",
		},
		"websocket": map[string]interface{}{
			"max_subscriptions": 50,
			"max_connections":   1,
			"ping_interval":     "30s",
		},
	}
}

// GetKrakenFeatures returns supported features
func GetKrakenFeatures() map[string]bool {
	return map[string]bool{
		"spot_trading":       true,
		"margin_trading":     true,
		"futures_trading":    true,
		"options_trading":    false,
		"staking":            true,
		"lending":            false,
		"websocket_public":   true,
		"websocket_private":  true,
		"order_book":         true,
		"trade_history":      true,
		"ohlcv_data":         true,
		"ticker_data":        true,
		"balance_info":       true,
		"order_management":   true,
		"stop_orders":        true,
		"conditional_orders": true,
		"api_keys":           true,
		"withdrawal":         true,
		"deposit":            true,
	}
}
