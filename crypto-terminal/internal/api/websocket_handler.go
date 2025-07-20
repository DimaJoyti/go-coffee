package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/trading"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// WebSocketHandler handles WebSocket connections for real-time updates
type WebSocketHandler struct {
	upgrader        websocket.Upgrader
	clients         map[*websocket.Conn]*Client
	clientsMutex    sync.RWMutex
	strategyEngine  *trading.StrategyEngine
	logger          *logrus.Logger
	
	// Channels for broadcasting
	priceUpdateChan    chan *PriceUpdate
	signalAlertChan    chan *SignalAlert
	tradeExecutionChan chan *TradeExecution
	portfolioUpdateChan chan *PortfolioUpdate
	riskAlertChan      chan *RiskAlert
	
	stopChan chan struct{}
}

// Client represents a WebSocket client
type Client struct {
	conn         *websocket.Conn
	send         chan []byte
	subscriptions map[string]bool
	userID       string
	lastPing     time.Time
}

// WebSocket message types
type MessageType string

const (
	MessageTypeSubscribe        MessageType = "subscribe"
	MessageTypeUnsubscribe      MessageType = "unsubscribe"
	MessageTypePriceUpdate      MessageType = "price_update"
	MessageTypeSignalAlert      MessageType = "signal_alert"
	MessageTypeTradeExecution   MessageType = "trade_execution"
	MessageTypePortfolioUpdate  MessageType = "portfolio_update"
	MessageTypeRiskAlert        MessageType = "risk_alert"
	MessageTypePing             MessageType = "ping"
	MessageTypePong             MessageType = "pong"
	MessageTypeError            MessageType = "error"
)

// WebSocket message structures
type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	ID        string      `json:"id,omitempty"`
}

type SubscriptionRequest struct {
	Channels []string `json:"channels"`
}

type PriceUpdate struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change24h float64 `json:"change_24h"`
	Volume    float64 `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

type SignalAlert struct {
	Strategy   string  `json:"strategy"`
	Symbol     string  `json:"symbol"`
	Signal     string  `json:"signal"`
	Confidence float64 `json:"confidence"`
	Price      float64 `json:"price"`
	Message    string  `json:"message"`
	Emoji      string  `json:"emoji"`
}

type TradeExecution struct {
	OrderID   string  `json:"order_id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Quantity  float64 `json:"quantity"`
	Price     float64 `json:"price"`
	Status    string  `json:"status"`
	Strategy  string  `json:"strategy"`
	Message   string  `json:"message"`
}

type PortfolioUpdate struct {
	TotalValue      float64            `json:"total_value"`
	AvailableBalance float64           `json:"available_balance"`
	TotalPnL        float64            `json:"total_pnl"`
	DailyPnL        float64            `json:"daily_pnl"`
	UnrealizedPnL   float64            `json:"unrealized_pnl"`
	Positions       map[string]interface{} `json:"positions"`
	ActiveStrategies []string          `json:"active_strategies"`
	Message         string             `json:"message"`
}

type RiskAlert struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Message     string  `json:"message"`
	Symbol      string  `json:"symbol,omitempty"`
	Strategy    string  `json:"strategy,omitempty"`
	RiskLevel   float64 `json:"risk_level"`
	Threshold   float64 `json:"threshold"`
	Action      string  `json:"action"`
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(strategyEngine *trading.StrategyEngine, logger *logrus.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:             make(map[*websocket.Conn]*Client),
		strategyEngine:      strategyEngine,
		logger:              logger,
		priceUpdateChan:     make(chan *PriceUpdate, 100),
		signalAlertChan:     make(chan *SignalAlert, 100),
		tradeExecutionChan:  make(chan *TradeExecution, 100),
		portfolioUpdateChan: make(chan *PortfolioUpdate, 100),
		riskAlertChan:       make(chan *RiskAlert, 100),
		stopChan:            make(chan struct{}),
	}
}

// RegisterRoutes registers WebSocket routes
func (wsh *WebSocketHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/ws/coffee-trading", wsh.HandleWebSocket)
}

// Start starts the WebSocket handler
func (wsh *WebSocketHandler) Start(ctx context.Context) error {
	wsh.logger.Info("Starting Coffee Trading WebSocket Handler")
	
	// Start background goroutines
	go wsh.broadcastLoop(ctx)
	go wsh.pingLoop(ctx)
	go wsh.cleanupLoop(ctx)
	
	wsh.logger.Info("Coffee Trading WebSocket Handler started successfully")
	return nil
}

// Stop stops the WebSocket handler
func (wsh *WebSocketHandler) Stop() error {
	wsh.logger.Info("Stopping Coffee Trading WebSocket Handler")
	
	close(wsh.stopChan)
	
	// Close all client connections
	wsh.clientsMutex.Lock()
	for conn, client := range wsh.clients {
		close(client.send)
		conn.Close()
		delete(wsh.clients, conn)
	}
	wsh.clientsMutex.Unlock()
	
	wsh.logger.Info("Coffee Trading WebSocket Handler stopped")
	return nil
}

// HandleWebSocket handles WebSocket connections
func (wsh *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := wsh.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		wsh.logger.Errorf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	client := &Client{
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
		userID:        c.Query("user_id"), // Get user ID from query params
		lastPing:      time.Now(),
	}

	wsh.clientsMutex.Lock()
	wsh.clients[conn] = client
	wsh.clientsMutex.Unlock()

	wsh.logger.Infof("New WebSocket client connected: %s", client.userID)

	// Send welcome message
	welcomeMsg := WebSocketMessage{
		Type: MessageTypeSignalAlert,
		Data: SignalAlert{
			Strategy:   "welcome",
			Symbol:     "COFFEE",
			Signal:     "CONNECTED",
			Confidence: 1.0,
			Message:    "☕ Welcome to Coffee Trading! Your real-time connection is brewing perfectly.",
			Emoji:      "☕",
		},
		Timestamp: time.Now(),
	}
	wsh.sendToClient(client, welcomeMsg)

	// Start goroutines for this client
	go wsh.writePump(client)
	go wsh.readPump(client)
}

// readPump handles reading messages from the client
func (wsh *WebSocketHandler) readPump(client *Client) {
	defer func() {
		wsh.clientsMutex.Lock()
		delete(wsh.clients, client.conn)
		wsh.clientsMutex.Unlock()
		client.conn.Close()
	}()

	client.conn.SetReadLimit(512)
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.lastPing = time.Now()
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				wsh.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			wsh.logger.Errorf("Failed to unmarshal WebSocket message: %v", err)
			continue
		}

		wsh.handleClientMessage(client, &msg)
	}
}

// writePump handles writing messages to the client
func (wsh *WebSocketHandler) writePump(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage handles messages from clients
func (wsh *WebSocketHandler) handleClientMessage(client *Client, msg *WebSocketMessage) {
	switch msg.Type {
	case MessageTypeSubscribe:
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if channels, ok := data["channels"].([]interface{}); ok {
				for _, ch := range channels {
					if channel, ok := ch.(string); ok {
						client.subscriptions[channel] = true
						wsh.logger.Infof("Client %s subscribed to channel: %s", client.userID, channel)
					}
				}
			}
		}

	case MessageTypeUnsubscribe:
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if channels, ok := data["channels"].([]interface{}); ok {
				for _, ch := range channels {
					if channel, ok := ch.(string); ok {
						delete(client.subscriptions, channel)
						wsh.logger.Infof("Client %s unsubscribed from channel: %s", client.userID, channel)
					}
				}
			}
		}

	case MessageTypePing:
		client.lastPing = time.Now()
		pongMsg := WebSocketMessage{
			Type:      MessageTypePong,
			Timestamp: time.Now(),
		}
		wsh.sendToClient(client, pongMsg)
	}
}

// Broadcasting methods
func (wsh *WebSocketHandler) BroadcastPriceUpdate(update *PriceUpdate) {
	select {
	case wsh.priceUpdateChan <- update:
	default:
		wsh.logger.Warn("Price update channel full, dropping update")
	}
}

func (wsh *WebSocketHandler) BroadcastSignalAlert(alert *SignalAlert) {
	select {
	case wsh.signalAlertChan <- alert:
	default:
		wsh.logger.Warn("Signal alert channel full, dropping alert")
	}
}

func (wsh *WebSocketHandler) BroadcastTradeExecution(execution *TradeExecution) {
	select {
	case wsh.tradeExecutionChan <- execution:
	default:
		wsh.logger.Warn("Trade execution channel full, dropping execution")
	}
}

func (wsh *WebSocketHandler) BroadcastPortfolioUpdate(update *PortfolioUpdate) {
	select {
	case wsh.portfolioUpdateChan <- update:
	default:
		wsh.logger.Warn("Portfolio update channel full, dropping update")
	}
}

func (wsh *WebSocketHandler) BroadcastRiskAlert(alert *RiskAlert) {
	select {
	case wsh.riskAlertChan <- alert:
	default:
		wsh.logger.Warn("Risk alert channel full, dropping alert")
	}
}

// broadcastLoop handles broadcasting messages to all subscribed clients
func (wsh *WebSocketHandler) broadcastLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-wsh.stopChan:
			return
		case update := <-wsh.priceUpdateChan:
			wsh.broadcastToSubscribers("price_updates", MessageTypePriceUpdate, update)
		case alert := <-wsh.signalAlertChan:
			wsh.broadcastToSubscribers("signal_alerts", MessageTypeSignalAlert, alert)
		case execution := <-wsh.tradeExecutionChan:
			wsh.broadcastToSubscribers("trade_executions", MessageTypeTradeExecution, execution)
		case update := <-wsh.portfolioUpdateChan:
			wsh.broadcastToSubscribers("portfolio_updates", MessageTypePortfolioUpdate, update)
		case alert := <-wsh.riskAlertChan:
			wsh.broadcastToSubscribers("risk_alerts", MessageTypeRiskAlert, alert)
		}
	}
}

// broadcastToSubscribers broadcasts a message to all clients subscribed to a channel
func (wsh *WebSocketHandler) broadcastToSubscribers(channel string, msgType MessageType, data interface{}) {
	msg := WebSocketMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}

	wsh.clientsMutex.RLock()
	for _, client := range wsh.clients {
		if client.subscriptions[channel] {
			wsh.sendToClient(client, msg)
		}
	}
	wsh.clientsMutex.RUnlock()
}

// sendToClient sends a message to a specific client
func (wsh *WebSocketHandler) sendToClient(client *Client, msg WebSocketMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		wsh.logger.Errorf("Failed to marshal WebSocket message: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		close(client.send)
		wsh.clientsMutex.Lock()
		delete(wsh.clients, client.conn)
		wsh.clientsMutex.Unlock()
	}
}

// pingLoop sends periodic ping messages to check client connectivity
func (wsh *WebSocketHandler) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-wsh.stopChan:
			return
		case <-ticker.C:
			wsh.clientsMutex.RLock()
			clientCount := len(wsh.clients)
			wsh.clientsMutex.RUnlock()
			
			if clientCount > 0 {
				wsh.logger.Debugf("WebSocket clients connected: %d", clientCount)
			}
		}
	}
}

// cleanupLoop removes stale client connections
func (wsh *WebSocketHandler) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-wsh.stopChan:
			return
		case <-ticker.C:
			wsh.cleanupStaleClients()
		}
	}
}

// cleanupStaleClients removes clients that haven't pinged recently
func (wsh *WebSocketHandler) cleanupStaleClients() {
	staleThreshold := time.Now().Add(-2 * time.Minute)
	
	wsh.clientsMutex.Lock()
	for conn, client := range wsh.clients {
		if client.lastPing.Before(staleThreshold) {
			wsh.logger.Infof("Removing stale client: %s", client.userID)
			close(client.send)
			conn.Close()
			delete(wsh.clients, conn)
		}
	}
	wsh.clientsMutex.Unlock()
}
