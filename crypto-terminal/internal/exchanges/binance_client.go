package exchanges

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// BinanceClient implements the ExchangeClient interface for Binance
type BinanceClient struct {
	apiKey     string
	secretKey  string
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

// BinanceConfig represents Binance client configuration
type BinanceConfig struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
	BaseURL   string `json:"base_url"`
	WSURL     string `json:"ws_url"`
	Testnet   bool   `json:"testnet"`
}

// NewBinanceClient creates a new Binance client
func NewBinanceClient(config *BinanceConfig, logger *logrus.Logger) *BinanceClient {
	baseURL := "https://api.binance.com"
	wsURL := "wss://stream.binance.com:9443/ws"
	
	if config.Testnet {
		baseURL = "https://testnet.binance.vision"
		wsURL = "wss://testnet.binance.vision/ws"
	}
	
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}
	if config.WSURL != "" {
		wsURL = config.WSURL
	}
	
	return &BinanceClient{
		apiKey:     config.APIKey,
		secretKey:  config.SecretKey,
		baseURL:    baseURL,
		wsURL:      wsURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
		wsChannels: make(map[string]chan interface{}),
		lastReset:  time.Now(),
	}
}

// GetExchangeType returns the exchange type
func (b *BinanceClient) GetExchangeType() ExchangeType {
	return ExchangeBinance
}

// Connect establishes connection to Binance
func (b *BinanceClient) Connect(ctx context.Context) error {
	// Test API connectivity
	if err := b.Ping(ctx); err != nil {
		b.lastError = err
		return fmt.Errorf("failed to connect to Binance: %w", err)
	}
	
	b.connected = true
	b.logger.Info("Connected to Binance API")
	return nil
}

// Disconnect closes connection to Binance
func (b *BinanceClient) Disconnect(ctx context.Context) error {
	b.connected = false
	
	// Close WebSocket connection if active
	if b.wsConn != nil {
		b.wsMutex.Lock()
		b.wsConn.Close()
		b.wsConn = nil
		b.wsMutex.Unlock()
	}
	
	b.streaming = false
	b.logger.Info("Disconnected from Binance API")
	return nil
}

// IsConnected returns connection status
func (b *BinanceClient) IsConnected() bool {
	return b.connected
}

// GetStatus returns current status
func (b *BinanceClient) GetStatus() string {
	if b.connected {
		return "connected"
	}
	return "disconnected"
}

// GetLastError returns the last error
func (b *BinanceClient) GetLastError() error {
	return b.lastError
}

// Ping tests connectivity to Binance API
func (b *BinanceClient) Ping(ctx context.Context) error {
	resp, err := b.makeRequest(ctx, "GET", "/api/v3/ping", nil, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// GetServerTime gets server time from Binance
func (b *BinanceClient) GetServerTime(ctx context.Context) (time.Time, error) {
	resp, err := b.makeRequest(ctx, "GET", "/api/v3/time", nil, false)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()
	
	var result struct {
		ServerTime int64 `json:"serverTime"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return time.Time{}, err
	}
	
	return time.Unix(0, result.ServerTime*int64(time.Millisecond)), nil
}

// GetExchangeInfo gets exchange information
func (b *BinanceClient) GetExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	resp, err := b.makeRequest(ctx, "GET", "/api/v3/exchangeInfo", nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Timezone   string `json:"timezone"`
		ServerTime int64  `json:"serverTime"`
		Symbols    []struct {
			Symbol string `json:"symbol"`
			Status string `json:"status"`
			BaseAsset  string `json:"baseAsset"`
			QuoteAsset string `json:"quoteAsset"`
		} `json:"symbols"`
		RateLimits []struct {
			RateLimitType string `json:"rateLimitType"`
			Interval      string `json:"interval"`
			Limit         int    `json:"limit"`
		} `json:"rateLimits"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Convert to our format
	symbols := make([]Symbol, len(result.Symbols))
	for i, s := range result.Symbols {
		symbols[i] = Symbol{
			Base:   s.BaseAsset,
			Quote:  s.QuoteAsset,
			Symbol: s.Symbol,
		}
	}
	
	rateLimits := make([]RateLimit, len(result.RateLimits))
	for i, rl := range result.RateLimits {
		rateLimits[i] = RateLimit{
			Type:     rl.RateLimitType,
			Interval: rl.Interval,
			Limit:    rl.Limit,
		}
	}
	
	return &ExchangeInfo{
		Exchange:   ExchangeBinance,
		Name:       "Binance",
		Status:     "operational",
		Symbols:    symbols,
		ServerTime: time.Unix(0, result.ServerTime*int64(time.Millisecond)),
		RateLimits: rateLimits,
	}, nil
}

// GetSymbols gets available trading symbols
func (b *BinanceClient) GetSymbols(ctx context.Context) ([]Symbol, error) {
	info, err := b.GetExchangeInfo(ctx)
	if err != nil {
		return nil, err
	}
	return info.Symbols, nil
}

// GetTicker gets ticker for a specific symbol
func (b *BinanceClient) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	
	resp, err := b.makeRequest(ctx, "GET", "/api/v3/ticker/24hr", params, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Symbol             string `json:"symbol"`
		LastPrice          string `json:"lastPrice"`
		BidPrice           string `json:"bidPrice"`
		AskPrice           string `json:"askPrice"`
		Volume             string `json:"volume"`
		QuoteVolume        string `json:"quoteVolume"`
		PriceChange        string `json:"priceChange"`
		PriceChangePercent string `json:"priceChangePercent"`
		HighPrice          string `json:"highPrice"`
		LowPrice           string `json:"lowPrice"`
		OpenPrice          string `json:"openPrice"`
		Count              int64  `json:"count"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return b.convertToTicker(&result)
}

// GetTickers gets tickers for multiple symbols
func (b *BinanceClient) GetTickers(ctx context.Context, symbols []string) ([]*Ticker, error) {
	// For multiple symbols, we'll use the all tickers endpoint and filter
	allTickers, err := b.GetAllTickers(ctx)
	if err != nil {
		return nil, err
	}
	
	symbolMap := make(map[string]bool)
	for _, symbol := range symbols {
		symbolMap[strings.ToUpper(symbol)] = true
	}
	
	var result []*Ticker
	for _, ticker := range allTickers {
		if symbolMap[ticker.Symbol] {
			result = append(result, ticker)
		}
	}
	
	return result, nil
}

// GetAllTickers gets all tickers
func (b *BinanceClient) GetAllTickers(ctx context.Context) ([]*Ticker, error) {
	resp, err := b.makeRequest(ctx, "GET", "/api/v3/ticker/24hr", nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var results []struct {
		Symbol             string `json:"symbol"`
		LastPrice          string `json:"lastPrice"`
		BidPrice           string `json:"bidPrice"`
		AskPrice           string `json:"askPrice"`
		Volume             string `json:"volume"`
		QuoteVolume        string `json:"quoteVolume"`
		PriceChange        string `json:"priceChange"`
		PriceChangePercent string `json:"priceChangePercent"`
		HighPrice          string `json:"highPrice"`
		LowPrice           string `json:"lowPrice"`
		OpenPrice          string `json:"openPrice"`
		Count              int64  `json:"count"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	
	tickers := make([]*Ticker, len(results))
	for i, result := range results {
		ticker, err := b.convertToTicker(&result)
		if err != nil {
			b.logger.Warnf("Failed to convert ticker for %s: %v", result.Symbol, err)
			continue
		}
		tickers[i] = ticker
	}
	
	return tickers, nil
}

// convertToTicker converts Binance ticker response to our Ticker format
func (b *BinanceClient) convertToTicker(result *struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	BidPrice           string `json:"bidPrice"`
	AskPrice           string `json:"askPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	OpenPrice          string `json:"openPrice"`
	Count              int64  `json:"count"`
}) (*Ticker, error) {
	lastPrice, _ := decimal.NewFromString(result.LastPrice)
	bidPrice, _ := decimal.NewFromString(result.BidPrice)
	askPrice, _ := decimal.NewFromString(result.AskPrice)
	volume, _ := decimal.NewFromString(result.Volume)
	quoteVolume, _ := decimal.NewFromString(result.QuoteVolume)
	change, _ := decimal.NewFromString(result.PriceChange)
	changePercent, _ := decimal.NewFromString(result.PriceChangePercent)
	high, _ := decimal.NewFromString(result.HighPrice)
	low, _ := decimal.NewFromString(result.LowPrice)
	open, _ := decimal.NewFromString(result.OpenPrice)
	
	return &Ticker{
		Exchange:         ExchangeBinance,
		Symbol:           result.Symbol,
		LastPrice:        lastPrice,
		BidPrice:         bidPrice,
		AskPrice:         askPrice,
		Volume24h:        volume,
		VolumeQuote24h:   quoteVolume,
		Change24h:        change,
		ChangePercent24h: changePercent,
		High24h:          high,
		Low24h:           low,
		OpenPrice:        open,
		Timestamp:        time.Now(),
		Count:            result.Count,
	}, nil
}

// makeRequest makes HTTP request to Binance API
func (b *BinanceClient) makeRequest(ctx context.Context, method, endpoint string, params url.Values, signed bool) (*http.Response, error) {
	// Rate limiting
	if err := b.checkRateLimit(); err != nil {
		return nil, err
	}
	
	var reqURL string
	if params != nil {
		reqURL = fmt.Sprintf("%s%s?%s", b.baseURL, endpoint, params.Encode())
	} else {
		reqURL = fmt.Sprintf("%s%s", b.baseURL, endpoint)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, err
	}
	
	// Add API key header
	if b.apiKey != "" {
		req.Header.Set("X-MBX-APIKEY", b.apiKey)
	}
	
	// Add signature for signed requests
	if signed && b.secretKey != "" {
		if params == nil {
			params = url.Values{}
		}
		params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
		
		signature := b.sign(params.Encode())
		params.Set("signature", signature)
		
		req.URL.RawQuery = params.Encode()
	}
	
	return b.httpClient.Do(req)
}

// sign creates HMAC SHA256 signature
func (b *BinanceClient) sign(payload string) string {
	h := hmac.New(sha256.New, []byte(b.secretKey))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// checkRateLimit implements basic rate limiting
func (b *BinanceClient) checkRateLimit() error {
	b.rateMutex.Lock()
	defer b.rateMutex.Unlock()

	now := time.Now()
	if now.Sub(b.lastReset) >= time.Minute {
		b.requestCount = 0
		b.lastReset = now
	}

	if b.requestCount >= 1200 { // Binance limit is 1200 requests per minute
		return fmt.Errorf("rate limit exceeded")
	}

	b.requestCount++
	return nil
}

// GetOrderBook gets order book for a symbol
func (b *BinanceClient) GetOrderBook(ctx context.Context, symbol string, limit int) (*OrderBook, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := b.makeRequest(ctx, "GET", "/api/v3/depth", params, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		LastUpdateId int64      `json:"lastUpdateId"`
		Bids         [][]string `json:"bids"`
		Asks         [][]string `json:"asks"`
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
		Exchange:     ExchangeBinance,
		Symbol:       strings.ToUpper(symbol),
		Bids:         bids,
		Asks:         asks,
		Timestamp:    time.Now(),
		LastUpdateID: result.LastUpdateId,
	}, nil
}

// GetRecentTrades gets recent trades for a symbol
func (b *BinanceClient) GetRecentTrades(ctx context.Context, symbol string, limit int) ([]*Trade, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := b.makeRequest(ctx, "GET", "/api/v3/trades", params, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []struct {
		ID           int64  `json:"id"`
		Price        string `json:"price"`
		Qty          string `json:"qty"`
		Time         int64  `json:"time"`
		IsBuyerMaker bool   `json:"isBuyerMaker"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	trades := make([]*Trade, len(results))
	for i, result := range results {
		price, _ := decimal.NewFromString(result.Price)
		quantity, _ := decimal.NewFromString(result.Qty)

		side := OrderSideBuy
		if result.IsBuyerMaker {
			side = OrderSideSell
		}

		trades[i] = &Trade{
			ID:        strconv.FormatInt(result.ID, 10),
			Exchange:  ExchangeBinance,
			Symbol:    strings.ToUpper(symbol),
			Price:     price,
			Quantity:  quantity,
			Side:      side,
			Timestamp: time.Unix(0, result.Time*int64(time.Millisecond)),
			IsMaker:   result.IsBuyerMaker,
		}
	}

	return trades, nil
}

// GetKlines gets candlestick data
func (b *BinanceClient) GetKlines(ctx context.Context, symbol, interval string, startTime, endTime *time.Time, limit int) ([]*Kline, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("interval", interval)

	if startTime != nil {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if endTime != nil {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := b.makeRequest(ctx, "GET", "/api/v3/klines", params, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	klines := make([]*Kline, len(results))
	for i, result := range results {
		openTime := int64(result[0].(float64))
		closeTime := int64(result[6].(float64))

		open, _ := decimal.NewFromString(result[1].(string))
		high, _ := decimal.NewFromString(result[2].(string))
		low, _ := decimal.NewFromString(result[3].(string))
		close, _ := decimal.NewFromString(result[4].(string))
		volume, _ := decimal.NewFromString(result[5].(string))
		quoteVolume, _ := decimal.NewFromString(result[7].(string))
		tradeCount := int64(result[8].(float64))
		takerBuyBaseVolume, _ := decimal.NewFromString(result[9].(string))
		takerBuyQuoteVolume, _ := decimal.NewFromString(result[10].(string))

		klines[i] = &Kline{
			Exchange:            ExchangeBinance,
			Symbol:              strings.ToUpper(symbol),
			Interval:            interval,
			OpenTime:            time.Unix(0, openTime*int64(time.Millisecond)),
			CloseTime:           time.Unix(0, closeTime*int64(time.Millisecond)),
			Open:                open,
			High:                high,
			Low:                 low,
			Close:               close,
			Volume:              volume,
			QuoteVolume:         quoteVolume,
			TradeCount:          tradeCount,
			TakerBuyBaseVolume:  takerBuyBaseVolume,
			TakerBuyQuoteVolume: takerBuyQuoteVolume,
		}
	}

	return klines, nil
}

// GetBalances gets account balances (requires authentication)
func (b *BinanceClient) GetBalances(ctx context.Context) ([]*Balance, error) {
	if b.apiKey == "" || b.secretKey == "" {
		return nil, fmt.Errorf("authentication required for account data")
	}

	resp, err := b.makeRequest(ctx, "GET", "/api/v3/account", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Balances []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	balances := make([]*Balance, 0, len(result.Balances))
	for _, balance := range result.Balances {
		free, _ := decimal.NewFromString(balance.Free)
		locked, _ := decimal.NewFromString(balance.Locked)
		total := free.Add(locked)

		// Only include non-zero balances
		if !total.IsZero() {
			balances = append(balances, &Balance{
				Exchange:  ExchangeBinance,
				Asset:     balance.Asset,
				Free:      free,
				Locked:    locked,
				Total:     total,
				Timestamp: time.Now(),
			})
		}
	}

	return balances, nil
}

// GetBalance gets balance for a specific asset
func (b *BinanceClient) GetBalance(ctx context.Context, asset string) (*Balance, error) {
	balances, err := b.GetBalances(ctx)
	if err != nil {
		return nil, err
	}

	for _, balance := range balances {
		if balance.Asset == strings.ToUpper(asset) {
			return balance, nil
		}
	}

	// Return zero balance if not found
	return &Balance{
		Exchange:  ExchangeBinance,
		Asset:     strings.ToUpper(asset),
		Free:      decimal.Zero,
		Locked:    decimal.Zero,
		Total:     decimal.Zero,
		Timestamp: time.Now(),
	}, nil
}

// PlaceOrder places a new order (requires authentication)
func (b *BinanceClient) PlaceOrder(ctx context.Context, order *Order) (*Order, error) {
	if b.apiKey == "" || b.secretKey == "" {
		return nil, fmt.Errorf("authentication required for trading")
	}

	params := url.Values{}
	params.Set("symbol", strings.ToUpper(order.Symbol))
	params.Set("side", strings.ToUpper(string(order.Side)))
	params.Set("type", strings.ToUpper(string(order.Type)))
	params.Set("quantity", order.Quantity.String())

	if order.Type == OrderTypeLimit {
		params.Set("price", order.Price.String())
		params.Set("timeInForce", "GTC") // Good Till Canceled
	}

	if order.ClientOrderID != "" {
		params.Set("newClientOrderId", order.ClientOrderID)
	}

	resp, err := b.makeRequest(ctx, "POST", "/api/v3/order", params, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Symbol              string `json:"symbol"`
		OrderId             int64  `json:"orderId"`
		ClientOrderId       string `json:"clientOrderId"`
		TransactTime        int64  `json:"transactTime"`
		Price               string `json:"price"`
		OrigQty             string `json:"origQty"`
		ExecutedQty         string `json:"executedQty"`
		CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
		Status              string `json:"status"`
		TimeInForce         string `json:"timeInForce"`
		Type                string `json:"type"`
		Side                string `json:"side"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	price, _ := decimal.NewFromString(result.Price)
	quantity, _ := decimal.NewFromString(result.OrigQty)
	filledQuantity, _ := decimal.NewFromString(result.ExecutedQty)

	return &Order{
		ID:                strconv.FormatInt(result.OrderId, 10),
		ClientOrderID:     result.ClientOrderId,
		Exchange:          ExchangeBinance,
		Symbol:            result.Symbol,
		Type:              OrderType(strings.ToLower(result.Type)),
		Side:              OrderSide(strings.ToLower(result.Side)),
		Status:            OrderStatus(strings.ToLower(result.Status)),
		Price:             price,
		Quantity:          quantity,
		FilledQuantity:    filledQuantity,
		RemainingQuantity: quantity.Sub(filledQuantity),
		TimeInForce:       result.TimeInForce,
		CreatedAt:         time.Unix(0, result.TransactTime*int64(time.Millisecond)),
		UpdatedAt:         time.Now(),
	}, nil
}

// CancelOrder cancels an existing order
func (b *BinanceClient) CancelOrder(ctx context.Context, symbol, orderID string) error {
	if b.apiKey == "" || b.secretKey == "" {
		return fmt.Errorf("authentication required for trading")
	}

	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("orderId", orderID)

	resp, err := b.makeRequest(ctx, "DELETE", "/api/v3/order", params, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to cancel order: %s", string(body))
	}

	return nil
}

// GetOrder gets order information
func (b *BinanceClient) GetOrder(ctx context.Context, symbol, orderID string) (*Order, error) {
	if b.apiKey == "" || b.secretKey == "" {
		return nil, fmt.Errorf("authentication required for order data")
	}

	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("orderId", orderID)

	resp, err := b.makeRequest(ctx, "GET", "/api/v3/order", params, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Symbol              string `json:"symbol"`
		OrderId             int64  `json:"orderId"`
		ClientOrderId       string `json:"clientOrderId"`
		Price               string `json:"price"`
		OrigQty             string `json:"origQty"`
		ExecutedQty         string `json:"executedQty"`
		CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
		Status              string `json:"status"`
		TimeInForce         string `json:"timeInForce"`
		Type                string `json:"type"`
		Side                string `json:"side"`
		StopPrice           string `json:"stopPrice"`
		Time                int64  `json:"time"`
		UpdateTime          int64  `json:"updateTime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	price, _ := decimal.NewFromString(result.Price)
	quantity, _ := decimal.NewFromString(result.OrigQty)
	filledQuantity, _ := decimal.NewFromString(result.ExecutedQty)
	stopPrice, _ := decimal.NewFromString(result.StopPrice)

	return &Order{
		ID:                strconv.FormatInt(result.OrderId, 10),
		ClientOrderID:     result.ClientOrderId,
		Exchange:          ExchangeBinance,
		Symbol:            result.Symbol,
		Type:              OrderType(strings.ToLower(result.Type)),
		Side:              OrderSide(strings.ToLower(result.Side)),
		Status:            OrderStatus(strings.ToLower(result.Status)),
		Price:             price,
		Quantity:          quantity,
		FilledQuantity:    filledQuantity,
		RemainingQuantity: quantity.Sub(filledQuantity),
		StopPrice:         stopPrice,
		TimeInForce:       result.TimeInForce,
		CreatedAt:         time.Unix(0, result.Time*int64(time.Millisecond)),
		UpdatedAt:         time.Unix(0, result.UpdateTime*int64(time.Millisecond)),
	}, nil
}

// GetOpenOrders gets open orders for a symbol
func (b *BinanceClient) GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error) {
	if b.apiKey == "" || b.secretKey == "" {
		return nil, fmt.Errorf("authentication required for order data")
	}

	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", strings.ToUpper(symbol))
	}

	resp, err := b.makeRequest(ctx, "GET", "/api/v3/openOrders", params, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []struct {
		Symbol              string `json:"symbol"`
		OrderId             int64  `json:"orderId"`
		ClientOrderId       string `json:"clientOrderId"`
		Price               string `json:"price"`
		OrigQty             string `json:"origQty"`
		ExecutedQty         string `json:"executedQty"`
		Status              string `json:"status"`
		TimeInForce         string `json:"timeInForce"`
		Type                string `json:"type"`
		Side                string `json:"side"`
		StopPrice           string `json:"stopPrice"`
		Time                int64  `json:"time"`
		UpdateTime          int64  `json:"updateTime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	orders := make([]*Order, len(results))
	for i, result := range results {
		price, _ := decimal.NewFromString(result.Price)
		quantity, _ := decimal.NewFromString(result.OrigQty)
		filledQuantity, _ := decimal.NewFromString(result.ExecutedQty)
		stopPrice, _ := decimal.NewFromString(result.StopPrice)

		orders[i] = &Order{
			ID:                strconv.FormatInt(result.OrderId, 10),
			ClientOrderID:     result.ClientOrderId,
			Exchange:          ExchangeBinance,
			Symbol:            result.Symbol,
			Type:              OrderType(strings.ToLower(result.Type)),
			Side:              OrderSide(strings.ToLower(result.Side)),
			Status:            OrderStatus(strings.ToLower(result.Status)),
			Price:             price,
			Quantity:          quantity,
			FilledQuantity:    filledQuantity,
			RemainingQuantity: quantity.Sub(filledQuantity),
			StopPrice:         stopPrice,
			TimeInForce:       result.TimeInForce,
			CreatedAt:         time.Unix(0, result.Time*int64(time.Millisecond)),
			UpdatedAt:         time.Unix(0, result.UpdateTime*int64(time.Millisecond)),
		}
	}

	return orders, nil
}

// GetOrderHistory gets order history for a symbol
func (b *BinanceClient) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error) {
	if b.apiKey == "" || b.secretKey == "" {
		return nil, fmt.Errorf("authentication required for order data")
	}

	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := b.makeRequest(ctx, "GET", "/api/v3/allOrders", params, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []struct {
		Symbol              string `json:"symbol"`
		OrderId             int64  `json:"orderId"`
		ClientOrderId       string `json:"clientOrderId"`
		Price               string `json:"price"`
		OrigQty             string `json:"origQty"`
		ExecutedQty         string `json:"executedQty"`
		Status              string `json:"status"`
		TimeInForce         string `json:"timeInForce"`
		Type                string `json:"type"`
		Side                string `json:"side"`
		StopPrice           string `json:"stopPrice"`
		Time                int64  `json:"time"`
		UpdateTime          int64  `json:"updateTime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	orders := make([]*Order, len(results))
	for i, result := range results {
		price, _ := decimal.NewFromString(result.Price)
		quantity, _ := decimal.NewFromString(result.OrigQty)
		filledQuantity, _ := decimal.NewFromString(result.ExecutedQty)
		stopPrice, _ := decimal.NewFromString(result.StopPrice)

		orders[i] = &Order{
			ID:                strconv.FormatInt(result.OrderId, 10),
			ClientOrderID:     result.ClientOrderId,
			Exchange:          ExchangeBinance,
			Symbol:            result.Symbol,
			Type:              OrderType(strings.ToLower(result.Type)),
			Side:              OrderSide(strings.ToLower(result.Side)),
			Status:            OrderStatus(strings.ToLower(result.Status)),
			Price:             price,
			Quantity:          quantity,
			FilledQuantity:    filledQuantity,
			RemainingQuantity: quantity.Sub(filledQuantity),
			StopPrice:         stopPrice,
			TimeInForce:       result.TimeInForce,
			CreatedAt:         time.Unix(0, result.Time*int64(time.Millisecond)),
			UpdatedAt:         time.Unix(0, result.UpdateTime*int64(time.Millisecond)),
		}
	}

	return orders, nil
}

// WebSocket Streaming Methods

// SubscribeToTicker subscribes to ticker updates via WebSocket
func (b *BinanceClient) SubscribeToTicker(ctx context.Context, symbol string, callback func(*Ticker)) error {
	stream := fmt.Sprintf("%s@ticker", strings.ToLower(symbol))
	return b.subscribeToStream(ctx, stream, func(data interface{}) {
		if tickerData, ok := data.(map[string]interface{}); ok {
			ticker := b.parseTickerFromWS(tickerData)
			if ticker != nil {
				callback(ticker)
			}
		}
	})
}

// SubscribeToTrades subscribes to trade updates via WebSocket
func (b *BinanceClient) SubscribeToTrades(ctx context.Context, symbol string, callback func(*Trade)) error {
	stream := fmt.Sprintf("%s@trade", strings.ToLower(symbol))
	return b.subscribeToStream(ctx, stream, func(data interface{}) {
		if tradeData, ok := data.(map[string]interface{}); ok {
			trade := b.parseTradeFromWS(tradeData)
			if trade != nil {
				callback(trade)
			}
		}
	})
}

// SubscribeToOrderBook subscribes to order book updates via WebSocket
func (b *BinanceClient) SubscribeToOrderBook(ctx context.Context, symbol string, callback func(*OrderBook)) error {
	stream := fmt.Sprintf("%s@depth", strings.ToLower(symbol))
	return b.subscribeToStream(ctx, stream, func(data interface{}) {
		if depthData, ok := data.(map[string]interface{}); ok {
			orderBook := b.parseOrderBookFromWS(depthData)
			if orderBook != nil {
				callback(orderBook)
			}
		}
	})
}

// SubscribeToKlines subscribes to kline updates via WebSocket
func (b *BinanceClient) SubscribeToKlines(ctx context.Context, symbol, interval string, callback func(*Kline)) error {
	stream := fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval)
	return b.subscribeToStream(ctx, stream, func(data interface{}) {
		if klineData, ok := data.(map[string]interface{}); ok {
			kline := b.parseKlineFromWS(klineData)
			if kline != nil {
				callback(kline)
			}
		}
	})
}

// UnsubscribeAll unsubscribes from all WebSocket streams
func (b *BinanceClient) UnsubscribeAll(ctx context.Context) error {
	b.wsMutex.Lock()
	defer b.wsMutex.Unlock()

	if b.wsConn != nil {
		b.wsConn.Close()
		b.wsConn = nil
	}

	b.streaming = false
	b.subscriptions = nil

	// Close all channels
	for _, ch := range b.wsChannels {
		close(ch)
	}
	b.wsChannels = make(map[string]chan interface{})

	return nil
}

// subscribeToStream subscribes to a specific WebSocket stream
func (b *BinanceClient) subscribeToStream(ctx context.Context, stream string, callback func(interface{})) error {
	b.wsMutex.Lock()
	defer b.wsMutex.Unlock()

	// Connect to WebSocket if not already connected
	if b.wsConn == nil {
		if err := b.connectWebSocket(ctx); err != nil {
			return err
		}
	}

	// Create channel for this stream
	ch := make(chan interface{}, 100)
	b.wsChannels[stream] = ch

	// Start goroutine to handle messages for this stream
	go func() {
		for data := range ch {
			callback(data)
		}
	}()

	// Send subscription message
	subMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{stream},
		"id":     time.Now().Unix(),
	}

	return b.wsConn.WriteJSON(subMsg)
}

// connectWebSocket establishes WebSocket connection
func (b *BinanceClient) connectWebSocket(ctx context.Context) error {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(b.wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	b.wsConn = conn
	b.streaming = true

	// Start message handler
	go b.handleWebSocketMessages()

	b.logger.Info("Connected to Binance WebSocket")
	return nil
}

// handleWebSocketMessages handles incoming WebSocket messages
func (b *BinanceClient) handleWebSocketMessages() {
	defer func() {
		b.wsMutex.Lock()
		if b.wsConn != nil {
			b.wsConn.Close()
			b.wsConn = nil
		}
		b.streaming = false
		b.wsMutex.Unlock()
	}()

	for {
		var msg map[string]interface{}
		if err := b.wsConn.ReadJSON(&msg); err != nil {
			b.logger.Errorf("WebSocket read error: %v", err)
			break
		}

		// Handle different message types
		if stream, ok := msg["stream"].(string); ok {
			if data, ok := msg["data"].(map[string]interface{}); ok {
				b.wsMutex.RLock()
				if ch, exists := b.wsChannels[stream]; exists {
					select {
					case ch <- data:
					default:
						// Channel full, skip message
					}
				}
				b.wsMutex.RUnlock()
			}
		}
	}
}

// WebSocket data parsing methods

// parseTickerFromWS parses ticker data from WebSocket
func (b *BinanceClient) parseTickerFromWS(data map[string]interface{}) *Ticker {
	symbol, _ := data["s"].(string)
	lastPrice, _ := decimal.NewFromString(data["c"].(string))
	bidPrice, _ := decimal.NewFromString(data["b"].(string))
	askPrice, _ := decimal.NewFromString(data["a"].(string))
	volume, _ := decimal.NewFromString(data["v"].(string))
	quoteVolume, _ := decimal.NewFromString(data["q"].(string))
	change, _ := decimal.NewFromString(data["P"].(string))
	high, _ := decimal.NewFromString(data["h"].(string))
	low, _ := decimal.NewFromString(data["l"].(string))
	open, _ := decimal.NewFromString(data["o"].(string))
	count := int64(data["n"].(float64))

	return &Ticker{
		Exchange:         ExchangeBinance,
		Symbol:           symbol,
		LastPrice:        lastPrice,
		BidPrice:         bidPrice,
		AskPrice:         askPrice,
		Volume24h:        volume,
		VolumeQuote24h:   quoteVolume,
		ChangePercent24h: change,
		High24h:          high,
		Low24h:           low,
		OpenPrice:        open,
		Count:            count,
		Timestamp:        time.Now(),
	}
}

// parseTradeFromWS parses trade data from WebSocket
func (b *BinanceClient) parseTradeFromWS(data map[string]interface{}) *Trade {
	symbol, _ := data["s"].(string)
	tradeID := strconv.FormatInt(int64(data["t"].(float64)), 10)
	price, _ := decimal.NewFromString(data["p"].(string))
	quantity, _ := decimal.NewFromString(data["q"].(string))
	timestamp := time.Unix(0, int64(data["T"].(float64))*int64(time.Millisecond))
	isMaker := data["m"].(bool)

	side := OrderSideBuy
	if isMaker {
		side = OrderSideSell
	}

	return &Trade{
		ID:        tradeID,
		Exchange:  ExchangeBinance,
		Symbol:    symbol,
		Price:     price,
		Quantity:  quantity,
		Side:      side,
		Timestamp: timestamp,
		IsMaker:   isMaker,
	}
}

// parseOrderBookFromWS parses order book data from WebSocket
func (b *BinanceClient) parseOrderBookFromWS(data map[string]interface{}) *OrderBook {
	symbol, _ := data["s"].(string)
	lastUpdateID := int64(data["u"].(float64))

	bidsData, _ := data["b"].([]interface{})
	asksData, _ := data["a"].([]interface{})

	bids := make([]OrderBookLevel, len(bidsData))
	for i, bidData := range bidsData {
		bid := bidData.([]interface{})
		price, _ := decimal.NewFromString(bid[0].(string))
		quantity, _ := decimal.NewFromString(bid[1].(string))
		bids[i] = OrderBookLevel{
			Price:    price,
			Quantity: quantity,
		}
	}

	asks := make([]OrderBookLevel, len(asksData))
	for i, askData := range asksData {
		ask := askData.([]interface{})
		price, _ := decimal.NewFromString(ask[0].(string))
		quantity, _ := decimal.NewFromString(ask[1].(string))
		asks[i] = OrderBookLevel{
			Price:    price,
			Quantity: quantity,
		}
	}

	return &OrderBook{
		Exchange:     ExchangeBinance,
		Symbol:       symbol,
		Bids:         bids,
		Asks:         asks,
		Timestamp:    time.Now(),
		LastUpdateID: lastUpdateID,
	}
}

// parseKlineFromWS parses kline data from WebSocket
func (b *BinanceClient) parseKlineFromWS(data map[string]interface{}) *Kline {
	klineData, _ := data["k"].(map[string]interface{})

	symbol, _ := klineData["s"].(string)
	interval, _ := klineData["i"].(string)
	openTime := time.Unix(0, int64(klineData["t"].(float64))*int64(time.Millisecond))
	closeTime := time.Unix(0, int64(klineData["T"].(float64))*int64(time.Millisecond))

	open, _ := decimal.NewFromString(klineData["o"].(string))
	high, _ := decimal.NewFromString(klineData["h"].(string))
	low, _ := decimal.NewFromString(klineData["l"].(string))
	close, _ := decimal.NewFromString(klineData["c"].(string))
	volume, _ := decimal.NewFromString(klineData["v"].(string))
	quoteVolume, _ := decimal.NewFromString(klineData["q"].(string))
	tradeCount := int64(klineData["n"].(float64))
	takerBuyBaseVolume, _ := decimal.NewFromString(klineData["V"].(string))
	takerBuyQuoteVolume, _ := decimal.NewFromString(klineData["Q"].(string))

	return &Kline{
		Exchange:            ExchangeBinance,
		Symbol:              symbol,
		Interval:            interval,
		OpenTime:            openTime,
		CloseTime:           closeTime,
		Open:                open,
		High:                high,
		Low:                 low,
		Close:               close,
		Volume:              volume,
		QuoteVolume:         quoteVolume,
		TradeCount:          tradeCount,
		TakerBuyBaseVolume:  takerBuyBaseVolume,
		TakerBuyQuoteVolume: takerBuyQuoteVolume,
	}
}
