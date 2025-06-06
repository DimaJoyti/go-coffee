package exchanges

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// CoinbaseClient implements the ExchangeClient interface for Coinbase Pro
type CoinbaseClient struct {
	apiKey     string
	secretKey  string
	passphrase string
	baseURL    string
	wsURL      string
	httpClient *http.Client
	logger     *logrus.Logger
	
	// WebSocket connections
	wsConn     *websocket.Conn
	wsMutex    sync.RWMutex
	wsChannels map[string]chan interface{}
	
	// State management
	connected    bool
	streaming    bool
	lastError    error
	subscriptions []*SubscriptionRequest
	
	// Rate limiting
	requestCount int
	lastReset    time.Time
	rateMutex    sync.Mutex
}

// CoinbaseConfig represents Coinbase client configuration
type CoinbaseConfig struct {
	APIKey     string `json:"api_key"`
	SecretKey  string `json:"secret_key"`
	Passphrase string `json:"passphrase"`
	BaseURL    string `json:"base_url"`
	WSURL      string `json:"ws_url"`
	Sandbox    bool   `json:"sandbox"`
}

// NewCoinbaseClient creates a new Coinbase client
func NewCoinbaseClient(config *CoinbaseConfig, logger *logrus.Logger) *CoinbaseClient {
	baseURL := "https://api.exchange.coinbase.com"
	wsURL := "wss://ws-feed.exchange.coinbase.com"
	
	if config.Sandbox {
		baseURL = "https://api-public.sandbox.exchange.coinbase.com"
		wsURL = "wss://ws-feed-public.sandbox.exchange.coinbase.com"
	}
	
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}
	if config.WSURL != "" {
		wsURL = config.WSURL
	}
	
	return &CoinbaseClient{
		apiKey:     config.APIKey,
		secretKey:  config.SecretKey,
		passphrase: config.Passphrase,
		baseURL:    baseURL,
		wsURL:      wsURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
		wsChannels: make(map[string]chan interface{}),
		lastReset:  time.Now(),
	}
}

// GetExchangeType returns the exchange type
func (c *CoinbaseClient) GetExchangeType() ExchangeType {
	return ExchangeCoinbase
}

// Connect establishes connection to Coinbase
func (c *CoinbaseClient) Connect(ctx context.Context) error {
	// Test API connectivity
	if err := c.Ping(ctx); err != nil {
		c.lastError = err
		return fmt.Errorf("failed to connect to Coinbase: %w", err)
	}
	
	c.connected = true
	c.logger.Info("Connected to Coinbase API")
	return nil
}

// Disconnect closes connection to Coinbase
func (c *CoinbaseClient) Disconnect(ctx context.Context) error {
	c.connected = false
	
	// Close WebSocket connection if active
	if c.wsConn != nil {
		c.wsMutex.Lock()
		c.wsConn.Close()
		c.wsConn = nil
		c.wsMutex.Unlock()
	}
	
	c.streaming = false
	c.logger.Info("Disconnected from Coinbase API")
	return nil
}

// IsConnected returns connection status
func (c *CoinbaseClient) IsConnected() bool {
	return c.connected
}

// GetStatus returns current status
func (c *CoinbaseClient) GetStatus() string {
	if c.connected {
		return "connected"
	}
	return "disconnected"
}

// GetLastError returns the last error
func (c *CoinbaseClient) GetLastError() error {
	return c.lastError
}

// Ping tests connectivity to Coinbase API
func (c *CoinbaseClient) Ping(ctx context.Context) error {
	resp, err := c.makeRequest(ctx, "GET", "/time", nil, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// GetServerTime gets server time from Coinbase
func (c *CoinbaseClient) GetServerTime(ctx context.Context) (time.Time, error) {
	resp, err := c.makeRequest(ctx, "GET", "/time", nil, false)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()
	
	var result struct {
		ISO   string  `json:"iso"`
		Epoch float64 `json:"epoch"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return time.Time{}, err
	}
	
	return time.Unix(int64(result.Epoch), 0), nil
}

// GetExchangeInfo gets exchange information
func (c *CoinbaseClient) GetExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	resp, err := c.makeRequest(ctx, "GET", "/products", nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var products []struct {
		ID             string `json:"id"`
		BaseCurrency   string `json:"base_currency"`
		QuoteCurrency  string `json:"quote_currency"`
		Status         string `json:"status"`
		TradingDisabled bool  `json:"trading_disabled"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, err
	}
	
	// Convert to our format
	symbols := make([]Symbol, 0, len(products))
	for _, p := range products {
		if !p.TradingDisabled && p.Status == "online" {
			symbols = append(symbols, Symbol{
				Base:   p.BaseCurrency,
				Quote:  p.QuoteCurrency,
				Symbol: p.ID,
			})
		}
	}
	
	return &ExchangeInfo{
		Exchange:   ExchangeCoinbase,
		Name:       "Coinbase Pro",
		Status:     "operational",
		Symbols:    symbols,
		ServerTime: time.Now(),
		RateLimits: []RateLimit{
			{
				Type:     "REQUEST",
				Interval: "SECOND",
				Limit:    10,
			},
		},
	}, nil
}

// GetSymbols gets available trading symbols
func (c *CoinbaseClient) GetSymbols(ctx context.Context) ([]Symbol, error) {
	info, err := c.GetExchangeInfo(ctx)
	if err != nil {
		return nil, err
	}
	return info.Symbols, nil
}

// GetTicker gets ticker for a specific symbol
func (c *CoinbaseClient) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	// Coinbase uses different endpoint for ticker
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/products/%s/ticker", symbol), nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Price  string    `json:"price"`
		Bid    string    `json:"bid"`
		Ask    string    `json:"ask"`
		Volume string    `json:"volume"`
		Time   time.Time `json:"time"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Get 24h stats for additional data
	statsResp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/products/%s/stats", symbol), nil, false)
	if err != nil {
		return nil, err
	}
	defer statsResp.Body.Close()
	
	var stats struct {
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Volume string `json:"volume"`
	}
	
	json.NewDecoder(statsResp.Body).Decode(&stats)
	
	price, _ := decimal.NewFromString(result.Price)
	bid, _ := decimal.NewFromString(result.Bid)
	ask, _ := decimal.NewFromString(result.Ask)
	volume, _ := decimal.NewFromString(result.Volume)
	open, _ := decimal.NewFromString(stats.Open)
	high, _ := decimal.NewFromString(stats.High)
	low, _ := decimal.NewFromString(stats.Low)
	
	change := price.Sub(open)
	changePercent := decimal.Zero
	if !open.IsZero() {
		changePercent = change.Div(open).Mul(decimal.NewFromInt(100))
	}
	
	return &Ticker{
		Exchange:         ExchangeCoinbase,
		Symbol:           symbol,
		LastPrice:        price,
		BidPrice:         bid,
		AskPrice:         ask,
		Volume24h:        volume,
		Change24h:        change,
		ChangePercent24h: changePercent,
		High24h:          high,
		Low24h:           low,
		OpenPrice:        open,
		Timestamp:        result.Time,
	}, nil
}

// GetTickers gets tickers for multiple symbols
func (c *CoinbaseClient) GetTickers(ctx context.Context, symbols []string) ([]*Ticker, error) {
	tickers := make([]*Ticker, 0, len(symbols))
	
	for _, symbol := range symbols {
		ticker, err := c.GetTicker(ctx, symbol)
		if err != nil {
			c.logger.Warnf("Failed to get ticker for %s: %v", symbol, err)
			continue
		}
		tickers = append(tickers, ticker)
	}
	
	return tickers, nil
}

// GetAllTickers gets all tickers
func (c *CoinbaseClient) GetAllTickers(ctx context.Context) ([]*Ticker, error) {
	symbols, err := c.GetSymbols(ctx)
	if err != nil {
		return nil, err
	}
	
	symbolNames := make([]string, len(symbols))
	for i, symbol := range symbols {
		symbolNames[i] = symbol.Symbol
	}
	
	return c.GetTickers(ctx, symbolNames)
}

// makeRequest makes HTTP request to Coinbase API
func (c *CoinbaseClient) makeRequest(ctx context.Context, method, endpoint string, params url.Values, signed bool) (*http.Response, error) {
	// Rate limiting
	if err := c.checkRateLimit(); err != nil {
		return nil, err
	}

	var reqURL string
	if params != nil {
		reqURL = fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, params.Encode())
	} else {
		reqURL = fmt.Sprintf("%s%s", c.baseURL, endpoint)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, err
	}

	// Add headers for signed requests
	if signed && c.apiKey != "" && c.secretKey != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		message := timestamp + method + endpoint
		if params != nil {
			message += "?" + params.Encode()
		}

		signature := c.sign(message)

		req.Header.Set("CB-ACCESS-KEY", c.apiKey)
		req.Header.Set("CB-ACCESS-SIGN", signature)
		req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
		req.Header.Set("CB-ACCESS-PASSPHRASE", c.passphrase)
	}

	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// sign creates HMAC SHA256 signature for Coinbase
func (c *CoinbaseClient) sign(message string) string {
	key, _ := base64.StdEncoding.DecodeString(c.secretKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// checkRateLimit implements basic rate limiting
func (c *CoinbaseClient) checkRateLimit() error {
	c.rateMutex.Lock()
	defer c.rateMutex.Unlock()

	now := time.Now()
	if now.Sub(c.lastReset) >= time.Second {
		c.requestCount = 0
		c.lastReset = now
	}

	if c.requestCount >= 10 { // Coinbase limit is 10 requests per second
		return fmt.Errorf("rate limit exceeded")
	}

	c.requestCount++
	return nil
}

// GetOrderBook gets order book for a symbol
func (c *CoinbaseClient) GetOrderBook(ctx context.Context, symbol string, limit int) (*OrderBook, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("level", "2") // Level 2 provides aggregated order book
	}

	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/products/%s/book", symbol), params, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Sequence int64      `json:"sequence"`
		Bids     [][]string `json:"bids"`
		Asks     [][]string `json:"asks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert bids and asks
	bids := make([]OrderBookLevel, len(result.Bids))
	for i, bid := range result.Bids {
		price, _ := decimal.NewFromString(bid[0])
		quantity, _ := decimal.NewFromString(bid[1])
		bids[i] = OrderBookLevel{
			Price:    price,
			Quantity: quantity,
		}
	}

	asks := make([]OrderBookLevel, len(result.Asks))
	for i, ask := range result.Asks {
		price, _ := decimal.NewFromString(ask[0])
		quantity, _ := decimal.NewFromString(ask[1])
		asks[i] = OrderBookLevel{
			Price:    price,
			Quantity: quantity,
		}
	}

	return &OrderBook{
		Exchange:     ExchangeCoinbase,
		Symbol:       symbol,
		Bids:         bids,
		Asks:         asks,
		Timestamp:    time.Now(),
		LastUpdateID: result.Sequence,
	}, nil
}

// GetRecentTrades gets recent trades for a symbol
func (c *CoinbaseClient) GetRecentTrades(ctx context.Context, symbol string, limit int) ([]*Trade, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/products/%s/trades", symbol), params, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []struct {
		TradeID int64     `json:"trade_id"`
		Price   string    `json:"price"`
		Size    string    `json:"size"`
		Side    string    `json:"side"`
		Time    time.Time `json:"time"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	trades := make([]*Trade, len(results))
	for i, result := range results {
		price, _ := decimal.NewFromString(result.Price)
		quantity, _ := decimal.NewFromString(result.Size)

		side := OrderSideBuy
		if result.Side == "sell" {
			side = OrderSideSell
		}

		trades[i] = &Trade{
			ID:        strconv.FormatInt(result.TradeID, 10),
			Exchange:  ExchangeCoinbase,
			Symbol:    symbol,
			Price:     price,
			Quantity:  quantity,
			Side:      side,
			Timestamp: result.Time,
			IsMaker:   false, // Coinbase doesn't provide this info in public trades
		}
	}

	return trades, nil
}

// GetKlines gets candlestick data
func (c *CoinbaseClient) GetKlines(ctx context.Context, symbol, interval string, startTime, endTime *time.Time, limit int) ([]*Kline, error) {
	params := url.Values{}

	// Convert interval to Coinbase format
	granularity := c.convertInterval(interval)
	params.Set("granularity", strconv.Itoa(granularity))

	if startTime != nil {
		params.Set("start", startTime.Format(time.RFC3339))
	}
	if endTime != nil {
		params.Set("end", endTime.Format(time.RFC3339))
	}

	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/products/%s/candles", symbol), params, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results [][]float64
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	klines := make([]*Kline, len(results))
	for i, result := range results {
		timestamp := time.Unix(int64(result[0]), 0)
		low := decimal.NewFromFloat(result[1])
		high := decimal.NewFromFloat(result[2])
		open := decimal.NewFromFloat(result[3])
		close := decimal.NewFromFloat(result[4])
		volume := decimal.NewFromFloat(result[5])

		klines[i] = &Kline{
			Exchange:    ExchangeCoinbase,
			Symbol:      symbol,
			Interval:    interval,
			OpenTime:    timestamp,
			CloseTime:   timestamp.Add(time.Duration(granularity) * time.Second),
			Open:        open,
			High:        high,
			Low:         low,
			Close:       close,
			Volume:      volume,
			QuoteVolume: volume.Mul(close), // Approximate quote volume
		}
	}

	return klines, nil
}

// convertInterval converts standard interval to Coinbase granularity
func (c *CoinbaseClient) convertInterval(interval string) int {
	switch interval {
	case "1m":
		return 60
	case "5m":
		return 300
	case "15m":
		return 900
	case "1h":
		return 3600
	case "6h":
		return 21600
	case "1d":
		return 86400
	default:
		return 3600 // Default to 1 hour
	}
}

// Account and Trading Methods (stubs for interface compliance)

// GetBalances gets account balances (requires authentication)
func (c *CoinbaseClient) GetBalances(ctx context.Context) ([]*Balance, error) {
	return nil, fmt.Errorf("authentication required for account data")
}

// GetBalance gets balance for a specific asset
func (c *CoinbaseClient) GetBalance(ctx context.Context, asset string) (*Balance, error) {
	return nil, fmt.Errorf("authentication required for account data")
}

// PlaceOrder places a new order (requires authentication)
func (c *CoinbaseClient) PlaceOrder(ctx context.Context, order *Order) (*Order, error) {
	return nil, fmt.Errorf("authentication required for trading")
}

// CancelOrder cancels an existing order
func (c *CoinbaseClient) CancelOrder(ctx context.Context, symbol, orderID string) error {
	return fmt.Errorf("authentication required for trading")
}

// GetOrder gets order information
func (c *CoinbaseClient) GetOrder(ctx context.Context, symbol, orderID string) (*Order, error) {
	return nil, fmt.Errorf("authentication required for order data")
}

// GetOpenOrders gets open orders for a symbol
func (c *CoinbaseClient) GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error) {
	return nil, fmt.Errorf("authentication required for order data")
}

// GetOrderHistory gets order history for a symbol
func (c *CoinbaseClient) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error) {
	return nil, fmt.Errorf("authentication required for order data")
}

// WebSocket Streaming Methods (stubs for interface compliance)

// SubscribeToTicker subscribes to ticker updates via WebSocket
func (c *CoinbaseClient) SubscribeToTicker(ctx context.Context, symbol string, callback func(*Ticker)) error {
	return fmt.Errorf("WebSocket streaming not implemented for Coinbase client")
}

// SubscribeToTrades subscribes to trade updates via WebSocket
func (c *CoinbaseClient) SubscribeToTrades(ctx context.Context, symbol string, callback func(*Trade)) error {
	return fmt.Errorf("WebSocket streaming not implemented for Coinbase client")
}

// SubscribeToOrderBook subscribes to order book updates via WebSocket
func (c *CoinbaseClient) SubscribeToOrderBook(ctx context.Context, symbol string, callback func(*OrderBook)) error {
	return fmt.Errorf("WebSocket streaming not implemented for Coinbase client")
}

// SubscribeToKlines subscribes to kline updates via WebSocket
func (c *CoinbaseClient) SubscribeToKlines(ctx context.Context, symbol, interval string, callback func(*Kline)) error {
	return fmt.Errorf("WebSocket streaming not implemented for Coinbase client")
}

// UnsubscribeAll unsubscribes from all WebSocket streams
func (c *CoinbaseClient) UnsubscribeAll(ctx context.Context) error {
	c.wsMutex.Lock()
	defer c.wsMutex.Unlock()

	if c.wsConn != nil {
		c.wsConn.Close()
		c.wsConn = nil
	}

	c.streaming = false
	c.subscriptions = nil

	// Close all channels
	for _, ch := range c.wsChannels {
		close(ch)
	}
	c.wsChannels = make(map[string]chan interface{})

	return nil
}
