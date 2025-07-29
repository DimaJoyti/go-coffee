package kraken

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// KrakenClient provides Kraken exchange API integration
type KrakenClient struct {
	logger *logger.Logger
	config KrakenConfig

	// HTTP client
	httpClient *http.Client

	// WebSocket connections
	wsPublic  *websocket.Conn
	wsPrivate *websocket.Conn
	wsMutex   sync.RWMutex

	// Subscriptions
	subscriptions map[string]*Subscription
	subMutex      sync.RWMutex

	// Event handlers
	eventHandlers map[EventType][]EventHandler
	handlerMutex  sync.RWMutex

	// Rate limiting
	rateLimiter *RateLimiter

	// State management
	isRunning bool
	stopChan  chan struct{}
	mutex     sync.RWMutex
}

// KrakenConfig holds configuration for Kraken client
type KrakenConfig struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	APIKey          string        `json:"api_key" yaml:"api_key"`
	APISecret       string        `json:"api_secret" yaml:"api_secret"`
	BaseURL         string        `json:"base_url" yaml:"base_url"`
	WebSocketURL    string        `json:"websocket_url" yaml:"websocket_url"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RateLimit       RateConfig    `json:"rate_limit" yaml:"rate_limit"`
	RetryConfig     RetryConfig   `json:"retry_config" yaml:"retry_config"`
	WebSocketConfig WSConfig      `json:"websocket_config" yaml:"websocket_config"`
}

// RateConfig holds rate limiting configuration
type RateConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	CounterDecay      time.Duration `json:"counter_decay"`
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// WSConfig holds WebSocket configuration
type WSConfig struct {
	Enabled           bool          `json:"enabled"`
	ReconnectInterval time.Duration `json:"reconnect_interval"`
	PingInterval      time.Duration `json:"ping_interval"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	BufferSize        int           `json:"buffer_size"`
}

// Market data structures
type Ticker struct {
	Symbol    string          `json:"symbol"`
	Ask       decimal.Decimal `json:"ask"`
	Bid       decimal.Decimal `json:"bid"`
	Last      decimal.Decimal `json:"last"`
	Volume    decimal.Decimal `json:"volume"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Open      decimal.Decimal `json:"open"`
	Change    decimal.Decimal `json:"change"`
	ChangePct decimal.Decimal `json:"change_pct"`
	Timestamp time.Time       `json:"timestamp"`
}

type OrderBook struct {
	Symbol    string      `json:"symbol"`
	Asks      []BookEntry `json:"asks"`
	Bids      []BookEntry `json:"bids"`
	Timestamp time.Time   `json:"timestamp"`
}

type BookEntry struct {
	Price     decimal.Decimal `json:"price"`
	Volume    decimal.Decimal `json:"volume"`
	Timestamp time.Time       `json:"timestamp"`
}

type Trade struct {
	ID        string          `json:"id"`
	Symbol    string          `json:"symbol"`
	Price     decimal.Decimal `json:"price"`
	Volume    decimal.Decimal `json:"volume"`
	Side      string          `json:"side"`
	Timestamp time.Time       `json:"timestamp"`
}

type OHLCV struct {
	Symbol    string          `json:"symbol"`
	Open      decimal.Decimal `json:"open"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Close     decimal.Decimal `json:"close"`
	Volume    decimal.Decimal `json:"volume"`
	Timestamp time.Time       `json:"timestamp"`
	Interval  string          `json:"interval"`
}

// Trading structures
type Order struct {
	ID              string                 `json:"id"`
	ClientOrderID   string                 `json:"client_order_id,omitempty"`
	Symbol          string                 `json:"symbol"`
	Side            string                 `json:"side"`
	Type            string                 `json:"type"`
	Amount          decimal.Decimal        `json:"amount"`
	Price           decimal.Decimal        `json:"price,omitempty"`
	Status          string                 `json:"status"`
	FilledAmount    decimal.Decimal        `json:"filled_amount"`
	RemainingAmount decimal.Decimal        `json:"remaining_amount"`
	AveragePrice    decimal.Decimal        `json:"average_price"`
	Fee             decimal.Decimal        `json:"fee"`
	FeeCurrency     string                 `json:"fee_currency"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type Balance struct {
	Currency  string          `json:"currency"`
	Available decimal.Decimal `json:"available"`
	Reserved  decimal.Decimal `json:"reserved"`
	Total     decimal.Decimal `json:"total"`
}

type Position struct {
	Symbol        string          `json:"symbol"`
	Side          string          `json:"side"`
	Size          decimal.Decimal `json:"size"`
	EntryPrice    decimal.Decimal `json:"entry_price"`
	MarkPrice     decimal.Decimal `json:"mark_price"`
	UnrealizedPnL decimal.Decimal `json:"unrealized_pnl"`
	Margin        decimal.Decimal `json:"margin"`
	Leverage      decimal.Decimal `json:"leverage"`
	Timestamp     time.Time       `json:"timestamp"`
}

// WebSocket structures
type Subscription struct {
	ID       string                 `json:"id"`
	Channel  string                 `json:"channel"`
	Symbol   string                 `json:"symbol,omitempty"`
	Interval string                 `json:"interval,omitempty"`
	Depth    int                    `json:"depth,omitempty"`
	Handler  EventHandler           `json:"-"`
	Metadata map[string]interface{} `json:"metadata"`
}

type WSMessage struct {
	Event     string                 `json:"event,omitempty"`
	Channel   string                 `json:"channel,omitempty"`
	Symbol    string                 `json:"symbol,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	RequestID string                 `json:"reqid,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type EventHandler func(event *WSMessage) error

// Enums
type EventType int

const (
	EventTypeTicker EventType = iota
	EventTypeOrderBook
	EventTypeTrade
	EventTypeOHLCV
	EventTypeOrderUpdate
	EventTypeBalanceUpdate
	EventTypePositionUpdate
	EventTypeError
	EventTypeConnection
	EventTypeDisconnection
)

// Rate limiter
type RateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewKrakenClient creates a new Kraken client
func NewKrakenClient(logger *logger.Logger, config KrakenConfig) *KrakenClient {
	return &KrakenClient{
		logger:        logger.Named("kraken-client"),
		config:        config,
		httpClient:    &http.Client{Timeout: config.Timeout},
		subscriptions: make(map[string]*Subscription),
		eventHandlers: make(map[EventType][]EventHandler),
		rateLimiter:   NewRateLimiter(config.RateLimit),
		stopChan:      make(chan struct{}),
	}
}

// Start starts the Kraken client
func (kc *KrakenClient) Start(ctx context.Context) error {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	if kc.isRunning {
		return fmt.Errorf("Kraken client is already running")
	}

	if !kc.config.Enabled {
		kc.logger.Info("Kraken client is disabled")
		return nil
	}

	kc.logger.Info("Starting Kraken client",
		zap.String("base_url", kc.config.BaseURL),
		zap.String("websocket_url", kc.config.WebSocketURL),
		zap.Bool("websocket_enabled", kc.config.WebSocketConfig.Enabled))

	// Test API connectivity
	if err := kc.testConnectivity(); err != nil {
		return fmt.Errorf("failed to test API connectivity: %w", err)
	}

	// Start WebSocket connections if enabled
	if kc.config.WebSocketConfig.Enabled {
		if err := kc.startWebSocketConnections(ctx); err != nil {
			return fmt.Errorf("failed to start WebSocket connections: %w", err)
		}
	}

	kc.isRunning = true
	kc.logger.Info("Kraken client started successfully")
	return nil
}

// Stop stops the Kraken client
func (kc *KrakenClient) Stop() error {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	if !kc.isRunning {
		return nil
	}

	kc.logger.Info("Stopping Kraken client")

	// Close WebSocket connections
	kc.closeWebSocketConnections()

	kc.isRunning = false
	close(kc.stopChan)

	kc.logger.Info("Kraken client stopped")
	return nil
}

// Market Data API methods

// GetTicker gets ticker information for a symbol
func (kc *KrakenClient) GetTicker(symbol string) (*Ticker, error) {
	kc.logger.Debug("Getting ticker", zap.String("symbol", symbol))

	if err := kc.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	endpoint := "/0/public/Ticker"
	params := url.Values{}
	params.Set("pair", symbol)

	response, err := kc.makePublicRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	ticker := kc.parseTickerResponse(symbol, response)
	return ticker, nil
}

// GetOrderBook gets order book for a symbol
func (kc *KrakenClient) GetOrderBook(symbol string, depth int) (*OrderBook, error) {
	kc.logger.Debug("Getting order book",
		zap.String("symbol", symbol),
		zap.Int("depth", depth))

	if err := kc.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	endpoint := "/0/public/Depth"
	params := url.Values{}
	params.Set("pair", symbol)
	if depth > 0 {
		params.Set("count", strconv.Itoa(depth))
	}

	response, err := kc.makePublicRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}

	orderBook := kc.parseOrderBookResponse(symbol, response)
	return orderBook, nil
}

// GetRecentTrades gets recent trades for a symbol
func (kc *KrakenClient) GetRecentTrades(symbol string, limit int) ([]Trade, error) {
	kc.logger.Debug("Getting recent trades",
		zap.String("symbol", symbol),
		zap.Int("limit", limit))

	if err := kc.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	endpoint := "/0/public/Trades"
	params := url.Values{}
	params.Set("pair", symbol)
	if limit > 0 {
		params.Set("count", strconv.Itoa(limit))
	}

	response, err := kc.makePublicRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent trades: %w", err)
	}

	trades := kc.parseTradesResponse(symbol, response)
	return trades, nil
}

// GetOHLCV gets OHLCV data for a symbol
func (kc *KrakenClient) GetOHLCV(symbol string, interval string, since time.Time) ([]OHLCV, error) {
	kc.logger.Debug("Getting OHLCV data",
		zap.String("symbol", symbol),
		zap.String("interval", interval))

	if err := kc.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	endpoint := "/0/public/OHLC"
	params := url.Values{}
	params.Set("pair", symbol)
	params.Set("interval", interval)
	if !since.IsZero() {
		params.Set("since", strconv.FormatInt(since.Unix(), 10))
	}

	response, err := kc.makePublicRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get OHLCV data: %w", err)
	}

	ohlcv := kc.parseOHLCVResponse(symbol, interval, response)
	return ohlcv, nil
}

// Trading API methods

// GetBalances gets account balances
func (kc *KrakenClient) GetBalances() ([]Balance, error) {
	kc.logger.Debug("Getting account balances")

	if err := kc.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	endpoint := "/0/private/Balance"
	params := url.Values{}

	response, err := kc.makePrivateRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get balances: %w", err)
	}

	balances := kc.parseBalancesResponse(response)
	return balances, nil
}

// PlaceOrder places a new order
func (kc *KrakenClient) PlaceOrder(symbol, side, orderType string, amount, price decimal.Decimal, options map[string]interface{}) (*Order, error) {
	kc.logger.Info("Placing order",
		zap.String("symbol", symbol),
		zap.String("side", side),
		zap.String("type", orderType),
		zap.String("amount", amount.String()),
		zap.String("price", price.String()))

	if err := kc.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	endpoint := "/0/private/AddOrder"
	params := url.Values{}
	params.Set("pair", symbol)
	params.Set("type", side)
	params.Set("ordertype", orderType)
	params.Set("volume", amount.String())

	if orderType != "market" && !price.IsZero() {
		params.Set("price", price.String())
	}

	// Add optional parameters
	if options != nil {
		for key, value := range options {
			params.Set(key, fmt.Sprintf("%v", value))
		}
	}

	response, err := kc.makePrivateRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	order := kc.parseOrderResponse(response)
	return order, nil
}

// CancelOrder cancels an existing order
func (kc *KrakenClient) CancelOrder(orderID string) error {
	kc.logger.Info("Cancelling order", zap.String("order_id", orderID))

	if err := kc.rateLimiter.Wait(); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	endpoint := "/0/private/CancelOrder"
	params := url.Values{}
	params.Set("txid", orderID)

	_, err := kc.makePrivateRequest(endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	kc.logger.Info("Order cancelled successfully", zap.String("order_id", orderID))
	return nil
}

// GetOpenOrders gets open orders
func (kc *KrakenClient) GetOpenOrders() ([]Order, error) {
	kc.logger.Debug("Getting open orders")

	if err := kc.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	endpoint := "/0/private/OpenOrders"
	params := url.Values{}

	response, err := kc.makePrivateRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}

	orders := kc.parseOrdersResponse(response)
	return orders, nil
}

// GetOrderHistory gets order history
func (kc *KrakenClient) GetOrderHistory(limit int, offset int) ([]Order, error) {
	kc.logger.Debug("Getting order history",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	if err := kc.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	endpoint := "/0/private/ClosedOrders"
	params := url.Values{}
	if limit > 0 {
		params.Set("count", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("ofs", strconv.Itoa(offset))
	}

	response, err := kc.makePrivateRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get order history: %w", err)
	}

	orders := kc.parseOrdersResponse(response)
	return orders, nil
}

// WebSocket methods

// SubscribeToTicker subscribes to ticker updates
func (kc *KrakenClient) SubscribeToTicker(symbol string, handler EventHandler) (*Subscription, error) {
	return kc.subscribe("ticker", symbol, "", 0, handler)
}

// SubscribeToOrderBook subscribes to order book updates
func (kc *KrakenClient) SubscribeToOrderBook(symbol string, depth int, handler EventHandler) (*Subscription, error) {
	return kc.subscribe("book", symbol, "", depth, handler)
}

// SubscribeToTrades subscribes to trade updates
func (kc *KrakenClient) SubscribeToTrades(symbol string, handler EventHandler) (*Subscription, error) {
	return kc.subscribe("trade", symbol, "", 0, handler)
}

// SubscribeToOHLCV subscribes to OHLCV updates
func (kc *KrakenClient) SubscribeToOHLCV(symbol string, interval string, handler EventHandler) (*Subscription, error) {
	return kc.subscribe("ohlc", symbol, interval, 0, handler)
}

// Unsubscribe unsubscribes from a channel
func (kc *KrakenClient) Unsubscribe(subscriptionID string) error {
	kc.subMutex.Lock()
	defer kc.subMutex.Unlock()

	subscription, exists := kc.subscriptions[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	// Send unsubscribe message
	unsubMsg := map[string]interface{}{
		"event": "unsubscribe",
		"pair":  []string{subscription.Symbol},
		"subscription": map[string]interface{}{
			"name": subscription.Channel,
		},
	}

	if err := kc.sendWebSocketMessage(unsubMsg); err != nil {
		return fmt.Errorf("failed to send unsubscribe message: %w", err)
	}

	delete(kc.subscriptions, subscriptionID)
	kc.logger.Info("Unsubscribed from channel",
		zap.String("subscription_id", subscriptionID),
		zap.String("channel", subscription.Channel),
		zap.String("symbol", subscription.Symbol))

	return nil
}

// Helper methods

// testConnectivity tests API connectivity
func (kc *KrakenClient) testConnectivity() error {
	endpoint := "/0/public/Time"
	params := url.Values{}

	_, err := kc.makePublicRequest(endpoint, params)
	if err != nil {
		return fmt.Errorf("connectivity test failed: %w", err)
	}

	kc.logger.Info("API connectivity test successful")
	return nil
}

// startWebSocketConnections starts WebSocket connections
func (kc *KrakenClient) startWebSocketConnections(ctx context.Context) error {
	kc.logger.Info("Starting WebSocket connections")

	// Connect to public WebSocket
	if err := kc.connectPublicWebSocket(); err != nil {
		return fmt.Errorf("failed to connect to public WebSocket: %w", err)
	}

	// Connect to private WebSocket if credentials are provided
	if kc.config.APIKey != "" && kc.config.APISecret != "" {
		if err := kc.connectPrivateWebSocket(); err != nil {
			kc.logger.Warn("Failed to connect to private WebSocket", zap.Error(err))
		}
	}

	// Start WebSocket message handlers
	go kc.handleWebSocketMessages(ctx)

	return nil
}

// closeWebSocketConnections closes WebSocket connections
func (kc *KrakenClient) closeWebSocketConnections() {
	kc.wsMutex.Lock()
	defer kc.wsMutex.Unlock()

	if kc.wsPublic != nil {
		kc.wsPublic.Close()
		kc.wsPublic = nil
	}

	if kc.wsPrivate != nil {
		kc.wsPrivate.Close()
		kc.wsPrivate = nil
	}

	kc.logger.Info("WebSocket connections closed")
}

// connectPublicWebSocket connects to public WebSocket
func (kc *KrakenClient) connectPublicWebSocket() error {
	dialer := websocket.DefaultDialer
	dialer.ReadBufferSize = kc.config.WebSocketConfig.BufferSize
	dialer.WriteBufferSize = kc.config.WebSocketConfig.BufferSize

	conn, _, err := dialer.Dial(kc.config.WebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	kc.wsMutex.Lock()
	kc.wsPublic = conn
	kc.wsMutex.Unlock()

	kc.logger.Info("Connected to public WebSocket")
	return nil
}

// connectPrivateWebSocket connects to private WebSocket
func (kc *KrakenClient) connectPrivateWebSocket() error {
	// For Kraken, private WebSocket requires authentication token
	// This is a simplified implementation
	dialer := websocket.DefaultDialer
	dialer.ReadBufferSize = kc.config.WebSocketConfig.BufferSize
	dialer.WriteBufferSize = kc.config.WebSocketConfig.BufferSize

	conn, _, err := dialer.Dial(kc.config.WebSocketURL+"-auth", nil)
	if err != nil {
		return fmt.Errorf("failed to dial private WebSocket: %w", err)
	}

	kc.wsMutex.Lock()
	kc.wsPrivate = conn
	kc.wsMutex.Unlock()

	kc.logger.Info("Connected to private WebSocket")
	return nil
}

// handleWebSocketMessages handles incoming WebSocket messages
func (kc *KrakenClient) handleWebSocketMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-kc.stopChan:
			return
		default:
			kc.processWebSocketMessage()
		}
	}
}

// processWebSocketMessage processes a single WebSocket message
func (kc *KrakenClient) processWebSocketMessage() {
	kc.wsMutex.RLock()
	conn := kc.wsPublic
	kc.wsMutex.RUnlock()

	if conn == nil {
		return
	}

	conn.SetReadDeadline(time.Now().Add(kc.config.WebSocketConfig.ReadTimeout))

	var message WSMessage
	if err := conn.ReadJSON(&message); err != nil {
		kc.logger.Error("Failed to read WebSocket message", zap.Error(err))
		return
	}

	message.Timestamp = time.Now()
	kc.handleWebSocketEvent(&message)
}

// subscribe creates a new subscription
func (kc *KrakenClient) subscribe(channel, symbol, interval string, depth int, handler EventHandler) (*Subscription, error) {
	subscriptionID := fmt.Sprintf("%s_%s_%d", channel, symbol, time.Now().UnixNano())

	subscription := &Subscription{
		ID:       subscriptionID,
		Channel:  channel,
		Symbol:   symbol,
		Interval: interval,
		Depth:    depth,
		Handler:  handler,
		Metadata: make(map[string]interface{}),
	}

	// Send subscribe message
	subMsg := map[string]interface{}{
		"event": "subscribe",
		"pair":  []string{symbol},
		"subscription": map[string]interface{}{
			"name": channel,
		},
	}

	if interval != "" {
		subMsg["subscription"].(map[string]interface{})["interval"] = interval
	}
	if depth > 0 {
		subMsg["subscription"].(map[string]interface{})["depth"] = depth
	}

	if err := kc.sendWebSocketMessage(subMsg); err != nil {
		return nil, fmt.Errorf("failed to send subscribe message: %w", err)
	}

	kc.subMutex.Lock()
	kc.subscriptions[subscriptionID] = subscription
	kc.subMutex.Unlock()

	kc.logger.Info("Subscribed to channel",
		zap.String("subscription_id", subscriptionID),
		zap.String("channel", channel),
		zap.String("symbol", symbol))

	return subscription, nil
}

// sendWebSocketMessage sends a message via WebSocket
func (kc *KrakenClient) sendWebSocketMessage(message interface{}) error {
	kc.wsMutex.RLock()
	conn := kc.wsPublic
	kc.wsMutex.RUnlock()

	if conn == nil {
		return fmt.Errorf("WebSocket connection not available")
	}

	conn.SetWriteDeadline(time.Now().Add(kc.config.WebSocketConfig.WriteTimeout))
	return conn.WriteJSON(message)
}

// handleWebSocketEvent handles WebSocket events
func (kc *KrakenClient) handleWebSocketEvent(message *WSMessage) {
	// Determine event type based on channel
	var eventType EventType
	switch message.Channel {
	case "ticker":
		eventType = EventTypeTicker
	case "book":
		eventType = EventTypeOrderBook
	case "trade":
		eventType = EventTypeTrade
	case "ohlc":
		eventType = EventTypeOHLCV
	default:
		eventType = EventTypeError
	}

	// Call registered handlers
	kc.handlerMutex.RLock()
	handlers := kc.eventHandlers[eventType]
	kc.handlerMutex.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(message); err != nil {
				kc.logger.Error("Event handler error",
					zap.String("event_type", kc.getEventTypeString(eventType)),
					zap.Error(err))
			}
		}(handler)
	}

	// Call subscription-specific handler
	kc.subMutex.RLock()
	for _, subscription := range kc.subscriptions {
		if subscription.Channel == message.Channel && subscription.Symbol == message.Symbol {
			go func(sub *Subscription) {
				if err := sub.Handler(message); err != nil {
					kc.logger.Error("Subscription handler error",
						zap.String("subscription_id", sub.ID),
						zap.Error(err))
				}
			}(subscription)
		}
	}
	kc.subMutex.RUnlock()
}

// HTTP request methods

// makePublicRequest makes a public API request
func (kc *KrakenClient) makePublicRequest(endpoint string, params url.Values) (map[string]interface{}, error) {
	url := kc.config.BaseURL + endpoint
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Kraken-Go-Client/1.0")

	resp, err := kc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if errors, exists := result["error"].([]interface{}); exists && len(errors) > 0 {
		return nil, fmt.Errorf("API error: %v", errors[0])
	}

	return result, nil
}

// makePrivateRequest makes a private API request
func (kc *KrakenClient) makePrivateRequest(endpoint string, params url.Values) (map[string]interface{}, error) {
	if kc.config.APIKey == "" || kc.config.APISecret == "" {
		return nil, fmt.Errorf("API credentials not configured")
	}

	// Add nonce
	nonce := strconv.FormatInt(time.Now().UnixNano(), 10)
	params.Set("nonce", nonce)

	// Create signature
	signature, err := kc.createSignature(endpoint, params, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}

	url := kc.config.BaseURL + endpoint
	req, err := http.NewRequest("POST", url, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("API-Key", kc.config.APIKey)
	req.Header.Set("API-Sign", signature)
	req.Header.Set("User-Agent", "Kraken-Go-Client/1.0")

	resp, err := kc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if errors, exists := result["error"].([]interface{}); exists && len(errors) > 0 {
		return nil, fmt.Errorf("API error: %v", errors[0])
	}

	return result, nil
}
