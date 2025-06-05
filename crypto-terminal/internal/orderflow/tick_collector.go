package orderflow

import (
	"context"
	"encoding/json"
	"fmt"

	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// TickCollector collects real-time tick data from exchanges
type TickCollector struct {
	config      *config.Config
	redis       *redis.Client
	connections map[string]*websocket.Conn
	subscribers map[string][]chan *models.Tick
	mu          sync.RWMutex
	stopChan    chan struct{}
	isRunning   bool
}

// BinanceTradeData represents Binance trade stream data
type BinanceTradeData struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	TradeID   int64  `json:"t"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
	BuyerOrderID  int64 `json:"b"`
	SellerOrderID int64 `json:"a"`
	TradeTime     int64 `json:"T"`
	IsBuyerMaker  bool  `json:"m"`
}

// CoinbaseTradeData represents Coinbase trade data
type CoinbaseTradeData struct {
	Type      string `json:"type"`
	TradeID   int64  `json:"trade_id"`
	Sequence  int64  `json:"sequence"`
	MakerOrderID string `json:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id"`
	Time      string `json:"time"`
	ProductID string `json:"product_id"`
	Size      string `json:"size"`
	Price     string `json:"price"`
	Side      string `json:"side"`
}

// NewTickCollector creates a new tick collector
func NewTickCollector(config *config.Config, redis *redis.Client) (*TickCollector, error) {
	return &TickCollector{
		config:      config,
		redis:       redis,
		connections: make(map[string]*websocket.Conn),
		subscribers: make(map[string][]chan *models.Tick),
		stopChan:    make(chan struct{}),
	}, nil
}

// Start starts the tick collection
func (tc *TickCollector) Start(ctx context.Context) error {
	if tc.isRunning {
		return fmt.Errorf("tick collector is already running")
	}

	logrus.Info("Starting tick collector")

	// Start Binance tick collection
	go tc.startBinanceCollection(ctx)

	// Start Coinbase tick collection
	go tc.startCoinbaseCollection(ctx)

	// Start tick processing
	go tc.startTickProcessing(ctx)

	tc.isRunning = true
	logrus.Info("Tick collector started")
	return nil
}

// Stop stops the tick collection
func (tc *TickCollector) Stop() error {
	if !tc.isRunning {
		return nil
	}

	logrus.Info("Stopping tick collector")
	close(tc.stopChan)

	// Close all WebSocket connections
	tc.mu.Lock()
	for exchange, conn := range tc.connections {
		if conn != nil {
			conn.Close()
			logrus.Infof("Closed %s WebSocket connection", exchange)
		}
	}
	tc.connections = make(map[string]*websocket.Conn)
	tc.mu.Unlock()

	tc.isRunning = false
	logrus.Info("Tick collector stopped")
	return nil
}

// SubscribeToTicks subscribes to tick updates for a symbol
func (tc *TickCollector) SubscribeToTicks(symbol string) <-chan *models.Tick {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	ch := make(chan *models.Tick, 1000) // Large buffer for high-frequency data
	if tc.subscribers[symbol] == nil {
		tc.subscribers[symbol] = make([]chan *models.Tick, 0)
	}
	tc.subscribers[symbol] = append(tc.subscribers[symbol], ch)

	return ch
}

// UnsubscribeFromTicks unsubscribes from tick updates
func (tc *TickCollector) UnsubscribeFromTicks(symbol string, ch <-chan *models.Tick) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if subscribers, exists := tc.subscribers[symbol]; exists {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				tc.subscribers[symbol] = append(subscribers[:i], subscribers[i+1:]...)
				close(subscriber)
				break
			}
		}
	}
}

// startBinanceCollection starts collecting ticks from Binance
func (tc *TickCollector) startBinanceCollection(ctx context.Context) {
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "SOLUSDT"}
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-tc.stopChan:
			return
		default:
			if err := tc.connectToBinance(symbols); err != nil {
				logrus.Errorf("Failed to connect to Binance: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
		}
	}
}

// connectToBinance establishes WebSocket connection to Binance
func (tc *TickCollector) connectToBinance(symbols []string) error {
	// Create stream names for all symbols
	streams := make([]string, len(symbols))
	for i, symbol := range symbols {
		streams[i] = strings.ToLower(symbol) + "@trade"
	}
	
	streamParam := strings.Join(streams, "/")
	wsURL := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s", streamParam)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Binance WebSocket: %w", err)
	}

	tc.mu.Lock()
	tc.connections["binance"] = conn
	tc.mu.Unlock()

	logrus.Info("Connected to Binance WebSocket")

	// Read messages
	for {
		select {
		case <-tc.stopChan:
			return nil
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				logrus.Errorf("Binance WebSocket read error: %v", err)
				return err
			}

			if err := tc.processBinanceTrade(message); err != nil {
				logrus.Errorf("Failed to process Binance trade: %v", err)
			}
		}
	}
}

// processBinanceTrade processes a Binance trade message
func (tc *TickCollector) processBinanceTrade(message []byte) error {
	var trade BinanceTradeData
	if err := json.Unmarshal(message, &trade); err != nil {
		return fmt.Errorf("failed to unmarshal Binance trade: %w", err)
	}

	// Convert to our tick format
	price, err := decimal.NewFromString(trade.Price)
	if err != nil {
		return fmt.Errorf("failed to parse price: %w", err)
	}

	volume, err := decimal.NewFromString(trade.Quantity)
	if err != nil {
		return fmt.Errorf("failed to parse volume: %w", err)
	}

	// Determine trade side based on buyer maker flag
	side := "BUY"
	if trade.IsBuyerMaker {
		side = "SELL" // If buyer is maker, then this is a sell order hitting the bid
	}

	tick := &models.Tick{
		ID:          fmt.Sprintf("binance_%d", trade.TradeID),
		Symbol:      tc.convertBinanceSymbol(trade.Symbol),
		Price:       price,
		Volume:      volume,
		Side:        side,
		TradeID:     strconv.FormatInt(trade.TradeID, 10),
		Exchange:    "binance",
		Timestamp:   time.Unix(trade.TradeTime/1000, (trade.TradeTime%1000)*1000000),
		IsAggressor: !trade.IsBuyerMaker,
		Sequence:    trade.TradeID,
	}

	// Store tick and broadcast
	tc.storeTick(tick)
	tc.broadcastTick(tick)

	return nil
}

// startCoinbaseCollection starts collecting ticks from Coinbase
func (tc *TickCollector) startCoinbaseCollection(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-tc.stopChan:
			return
		default:
			if err := tc.connectToCoinbase(); err != nil {
				logrus.Errorf("Failed to connect to Coinbase: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
		}
	}
}

// connectToCoinbase establishes WebSocket connection to Coinbase
func (tc *TickCollector) connectToCoinbase() error {
	wsURL := "wss://ws-feed.exchange.coinbase.com"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Coinbase WebSocket: %w", err)
	}

	tc.mu.Lock()
	tc.connections["coinbase"] = conn
	tc.mu.Unlock()

	// Subscribe to trade channels
	subscribeMsg := map[string]interface{}{
		"type": "subscribe",
		"product_ids": []string{"BTC-USD", "ETH-USD", "ADA-USD", "SOL-USD"},
		"channels": []string{"matches"},
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		return fmt.Errorf("failed to subscribe to Coinbase channels: %w", err)
	}

	logrus.Info("Connected to Coinbase WebSocket")

	// Read messages
	for {
		select {
		case <-tc.stopChan:
			return nil
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				logrus.Errorf("Coinbase WebSocket read error: %v", err)
				return err
			}

			if err := tc.processCoinbaseTrade(message); err != nil {
				logrus.Errorf("Failed to process Coinbase trade: %v", err)
			}
		}
	}
}

// processCoinbaseTrade processes a Coinbase trade message
func (tc *TickCollector) processCoinbaseTrade(message []byte) error {
	var trade CoinbaseTradeData
	if err := json.Unmarshal(message, &trade); err != nil {
		return fmt.Errorf("failed to unmarshal Coinbase trade: %w", err)
	}

	// Only process match messages
	if trade.Type != "match" {
		return nil
	}

	price, err := decimal.NewFromString(trade.Price)
	if err != nil {
		return fmt.Errorf("failed to parse price: %w", err)
	}

	volume, err := decimal.NewFromString(trade.Size)
	if err != nil {
		return fmt.Errorf("failed to parse volume: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339, trade.Time)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	tick := &models.Tick{
		ID:          fmt.Sprintf("coinbase_%d", trade.TradeID),
		Symbol:      tc.convertCoinbaseSymbol(trade.ProductID),
		Price:       price,
		Volume:      volume,
		Side:        strings.ToUpper(trade.Side),
		TradeID:     strconv.FormatInt(trade.TradeID, 10),
		Exchange:    "coinbase",
		Timestamp:   timestamp,
		IsAggressor: true, // Coinbase matches are always aggressive
		Sequence:    trade.Sequence,
	}

	// Store tick and broadcast
	tc.storeTick(tick)
	tc.broadcastTick(tick)

	return nil
}

// storeTick stores a tick in Redis for caching
func (tc *TickCollector) storeTick(tick *models.Tick) {
	ctx := context.Background()
	
	// Store in Redis with expiration
	tickKey := fmt.Sprintf("tick:%s:%d", tick.Symbol, tick.Timestamp.UnixNano())
	tickData, err := json.Marshal(tick)
	if err != nil {
		logrus.Errorf("Failed to marshal tick: %v", err)
		return
	}

	if err := tc.redis.Set(ctx, tickKey, tickData, 24*time.Hour).Err(); err != nil {
		logrus.Errorf("Failed to store tick in Redis: %v", err)
	}

	// Add to symbol's tick list
	listKey := fmt.Sprintf("ticks:%s", tick.Symbol)
	if err := tc.redis.LPush(ctx, listKey, tickKey).Err(); err != nil {
		logrus.Errorf("Failed to add tick to list: %v", err)
	}

	// Trim list to keep only recent ticks (last 10000)
	if err := tc.redis.LTrim(ctx, listKey, 0, 9999).Err(); err != nil {
		logrus.Errorf("Failed to trim tick list: %v", err)
	}
}

// broadcastTick broadcasts a tick to subscribers
func (tc *TickCollector) broadcastTick(tick *models.Tick) {
	tc.mu.RLock()
	subscribers := tc.subscribers[tick.Symbol]
	tc.mu.RUnlock()

	for _, ch := range subscribers {
		select {
		case ch <- tick:
		default:
			// Channel is full, skip this subscriber
			logrus.Warnf("Tick subscriber channel full for symbol %s", tick.Symbol)
		}
	}
}

// startTickProcessing starts processing collected ticks
func (tc *TickCollector) startTickProcessing(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tc.stopChan:
			return
		case <-ticker.C:
			tc.processTickStatistics()
		}
	}
}

// processTickStatistics processes tick statistics
func (tc *TickCollector) processTickStatistics() {
	// Implementation placeholder - would calculate tick statistics
	// like ticks per second, volume per second, etc.
}

// Helper methods for symbol conversion
func (tc *TickCollector) convertBinanceSymbol(symbol string) string {
	symbolMap := map[string]string{
		"BTCUSDT":  "BTC",
		"ETHUSDT":  "ETH",
		"BNBUSDT":  "BNB",
		"ADAUSDT":  "ADA",
		"SOLUSDT":  "SOL",
		"DOTUSDT":  "DOT",
		"DOGEUSDT": "DOGE",
		"AVAXUSDT": "AVAX",
		"MATICUSDT": "MATIC",
		"LINKUSDT": "LINK",
	}

	if converted, exists := symbolMap[symbol]; exists {
		return converted
	}
	return strings.TrimSuffix(symbol, "USDT")
}

func (tc *TickCollector) convertCoinbaseSymbol(productID string) string {
	symbolMap := map[string]string{
		"BTC-USD": "BTC",
		"ETH-USD": "ETH",
		"ADA-USD": "ADA",
		"SOL-USD": "SOL",
		"DOT-USD": "DOT",
		"DOGE-USD": "DOGE",
		"AVAX-USD": "AVAX",
		"MATIC-USD": "MATIC",
		"LINK-USD": "LINK",
	}

	if converted, exists := symbolMap[productID]; exists {
		return converted
	}
	return strings.Split(productID, "-")[0]
}
