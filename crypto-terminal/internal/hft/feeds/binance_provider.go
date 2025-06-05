package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// BinanceProvider implements ultra-low latency Binance WebSocket feeds
type BinanceProvider struct {
	config        *config.Config
	conn          *websocket.Conn
	tickChan      chan *models.MarketDataTick
	orderBookChan chan *models.OrderBook
	isConnected   bool
	mu            sync.RWMutex
	stopChan      chan struct{}
	wg            sync.WaitGroup
	subscriptions map[string]bool
	lastPingTime  time.Time
	latency       time.Duration
}

// BinanceTickerMessage represents Binance ticker WebSocket message
type BinanceTickerMessage struct {
	Stream string `json:"stream"`
	Data   struct {
		EventType   string `json:"e"`
		EventTime   int64  `json:"E"`
		Symbol      string `json:"s"`
		PriceChange string `json:"p"`
		Price       string `json:"c"`
		Volume      string `json:"v"`
		BidPrice    string `json:"b"`
		BidQty      string `json:"B"`
		AskPrice    string `json:"a"`
		AskQty      string `json:"A"`
	} `json:"data"`
}

// BinanceDepthMessage represents Binance order book WebSocket message
type BinanceDepthMessage struct {
	Stream string `json:"stream"`
	Data   struct {
		EventType     string     `json:"e"`
		EventTime     int64      `json:"E"`
		Symbol        string     `json:"s"`
		FirstUpdateID int64      `json:"U"`
		FinalUpdateID int64      `json:"u"`
		Bids          [][]string `json:"b"`
		Asks          [][]string `json:"a"`
	} `json:"data"`
}

// NewBinanceProvider creates a new Binance market data provider
func NewBinanceProvider(cfg *config.Config) (*BinanceProvider, error) {
	return &BinanceProvider{
		config:        cfg,
		tickChan:      make(chan *models.MarketDataTick, 10000),
		orderBookChan: make(chan *models.OrderBook, 1000),
		stopChan:      make(chan struct{}),
		subscriptions: make(map[string]bool),
	}, nil
}

// Connect establishes WebSocket connection to Binance
func (p *BinanceProvider) Connect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isConnected {
		return fmt.Errorf("already connected")
	}

	// Connect to Binance WebSocket
	wsURL := "wss://stream.binance.com:9443/ws"
	
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second
	
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Binance WebSocket: %w", err)
	}

	p.conn = conn
	p.isConnected = true

	// Start message processing
	p.wg.Add(2)
	go p.readMessages(ctx)
	go p.pingPong(ctx)

	logrus.Info("Connected to Binance WebSocket feed")
	return nil
}

// Disconnect closes the WebSocket connection
func (p *BinanceProvider) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isConnected {
		return nil
	}

	close(p.stopChan)
	
	if p.conn != nil {
		p.conn.Close()
	}

	p.wg.Wait()
	p.isConnected = false

	logrus.Info("Disconnected from Binance WebSocket feed")
	return nil
}

// Subscribe subscribes to market data for specified symbols
func (p *BinanceProvider) Subscribe(symbols []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isConnected {
		return fmt.Errorf("not connected")
	}

	var streams []string
	for _, symbol := range symbols {
		symbol = strings.ToLower(symbol)
		
		// Subscribe to ticker stream
		tickerStream := fmt.Sprintf("%s@ticker", symbol)
		streams = append(streams, tickerStream)
		
		// Subscribe to depth stream
		depthStream := fmt.Sprintf("%s@depth@100ms", symbol)
		streams = append(streams, depthStream)
		
		p.subscriptions[symbol] = true
	}

	// Send subscription message
	subscribeMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": streams,
		"id":     time.Now().UnixNano(),
	}

	return p.conn.WriteJSON(subscribeMsg)
}

// Unsubscribe unsubscribes from market data for specified symbols
func (p *BinanceProvider) Unsubscribe(symbols []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isConnected {
		return fmt.Errorf("not connected")
	}

	var streams []string
	for _, symbol := range symbols {
		symbol = strings.ToLower(symbol)
		
		tickerStream := fmt.Sprintf("%s@ticker", symbol)
		streams = append(streams, tickerStream)
		
		depthStream := fmt.Sprintf("%s@depth@100ms", symbol)
		streams = append(streams, depthStream)
		
		delete(p.subscriptions, symbol)
	}

	// Send unsubscription message
	unsubscribeMsg := map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": streams,
		"id":     time.Now().UnixNano(),
	}

	return p.conn.WriteJSON(unsubscribeMsg)
}

// GetTickChannel returns the tick data channel
func (p *BinanceProvider) GetTickChannel() <-chan *models.MarketDataTick {
	return p.tickChan
}

// GetOrderBookChannel returns the order book data channel
func (p *BinanceProvider) GetOrderBookChannel() <-chan *models.OrderBook {
	return p.orderBookChan
}

// IsConnected returns the connection status
func (p *BinanceProvider) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isConnected
}

// GetLatency returns the current latency
func (p *BinanceProvider) GetLatency() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.latency
}

// readMessages reads and processes WebSocket messages
func (p *BinanceProvider) readMessages(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		default:
			var rawMsg json.RawMessage
			err := p.conn.ReadJSON(&rawMsg)
			if err != nil {
				logrus.WithError(err).Error("Failed to read Binance WebSocket message")
				return
			}

			receiveTime := time.Now()
			p.processMessage(rawMsg, receiveTime)
		}
	}
}

// processMessage processes incoming WebSocket messages
func (p *BinanceProvider) processMessage(rawMsg json.RawMessage, receiveTime time.Time) {
	// Try to parse as ticker message
	var tickerMsg BinanceTickerMessage
	if err := json.Unmarshal(rawMsg, &tickerMsg); err == nil && tickerMsg.Data.EventType == "24hrTicker" {
		p.processTicker(tickerMsg, receiveTime)
		return
	}

	// Try to parse as depth message
	var depthMsg BinanceDepthMessage
	if err := json.Unmarshal(rawMsg, &depthMsg); err == nil && depthMsg.Data.EventType == "depthUpdate" {
		p.processDepth(depthMsg, receiveTime)
		return
	}
}

// processTicker processes ticker messages and creates market data ticks
func (p *BinanceProvider) processTicker(msg BinanceTickerMessage, receiveTime time.Time) {
	price, err := decimal.NewFromString(msg.Data.Price)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse price")
		return
	}

	volume, err := decimal.NewFromString(msg.Data.Volume)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse volume")
		return
	}

	bidPrice, err := decimal.NewFromString(msg.Data.BidPrice)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse bid price")
		return
	}

	bidQty, err := decimal.NewFromString(msg.Data.BidQty)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse bid quantity")
		return
	}

	askPrice, err := decimal.NewFromString(msg.Data.AskPrice)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse ask price")
		return
	}

	askQty, err := decimal.NewFromString(msg.Data.AskQty)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse ask quantity")
		return
	}

	tick := &models.MarketDataTick{
		Symbol:      strings.ToUpper(msg.Data.Symbol),
		Exchange:    "binance",
		Price:       price,
		Quantity:    volume,
		Side:        models.OrderSideBuy, // Default, would need trade data for actual side
		BidPrice:    bidPrice,
		BidQuantity: bidQty,
		AskPrice:    askPrice,
		AskQuantity: askQty,
		Timestamp:   time.Unix(0, msg.Data.EventTime*int64(time.Millisecond)),
		ReceiveTime: receiveTime,
		SequenceNum: uint64(msg.Data.EventTime),
	}

	select {
	case p.tickChan <- tick:
	default:
		// Channel is full, skip
	}
}

// processDepth processes order book depth messages
func (p *BinanceProvider) processDepth(msg BinanceDepthMessage, receiveTime time.Time) {
	var bids []models.OrderBookLevel
	var asks []models.OrderBookLevel

	// Process bids
	for _, bid := range msg.Data.Bids {
		if len(bid) >= 2 {
			price, err := decimal.NewFromString(bid[0])
			if err != nil {
				continue
			}
			quantity, err := decimal.NewFromString(bid[1])
			if err != nil {
				continue
			}

			bids = append(bids, models.OrderBookLevel{
				Price:    price,
				Quantity: quantity,
				Count:    1,
			})
		}
	}

	// Process asks
	for _, ask := range msg.Data.Asks {
		if len(ask) >= 2 {
			price, err := decimal.NewFromString(ask[0])
			if err != nil {
				continue
			}
			quantity, err := decimal.NewFromString(ask[1])
			if err != nil {
				continue
			}

			asks = append(asks, models.OrderBookLevel{
				Price:    price,
				Quantity: quantity,
				Count:    1,
			})
		}
	}

	orderBook := &models.OrderBook{
		Symbol:      strings.ToUpper(msg.Data.Symbol),
		Exchange:    "binance",
		Bids:        bids,
		Asks:        asks,
		Timestamp:   time.Unix(0, msg.Data.EventTime*int64(time.Millisecond)),
		ReceiveTime: receiveTime,
		SequenceNum: uint64(msg.Data.FinalUpdateID),
	}

	select {
	case p.orderBookChan <- orderBook:
	default:
		// Channel is full, skip
	}
}

// pingPong handles WebSocket ping/pong to maintain connection
func (p *BinanceProvider) pingPong(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.mu.Lock()
			if p.isConnected && p.conn != nil {
				pingTime := time.Now()
				err := p.conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					logrus.WithError(err).Error("Failed to send ping to Binance")
					p.mu.Unlock()
					return
				}
				p.lastPingTime = pingTime
			}
			p.mu.Unlock()
		}
	}
}
