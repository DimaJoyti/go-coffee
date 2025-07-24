package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/middleware"
		"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MessageType represents different types of WebSocket messages
type MessageType string

const (
	MessageTypeMarketData     MessageType = "market_data"
	MessageTypePortfolio      MessageType = "portfolio"
	MessageTypeAlert          MessageType = "alert"
	MessageTypeOrderFlow      MessageType = "order_flow"
	MessageTypeHFT            MessageType = "hft"
	MessageTypeNews           MessageType = "news"
	MessageTypeSubscribe      MessageType = "subscribe"
	MessageTypeUnsubscribe    MessageType = "unsubscribe"
	MessageTypeHeartbeat      MessageType = "heartbeat"
	MessageTypeError          MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Channel   string          `json:"channel,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	RequestID string          `json:"request_id,omitempty"`
}

// Subscription represents a client subscription
type Subscription struct {
	Channel   string            `json:"channel"`
	Filters   map[string]string `json:"filters,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// EnhancedClient represents an enhanced WebSocket client with subscriptions
type EnhancedClient struct {
	ID            string
	Conn          *websocket.Conn
	Send          chan Message
	Hub           *EnhancedHub
	Subscriptions map[string]*Subscription
	UserID        string
	LastSeen      time.Time
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	metrics       *middleware.WebSocketMetrics
}

// EnhancedHub maintains the set of active clients and broadcasts messages to them
type EnhancedHub struct {
	// Registered clients
	clients map[*EnhancedClient]bool

	// Inbound messages from the clients
	broadcast chan Message

	// Register requests from the clients
	register chan *EnhancedClient

	// Unregister requests from clients
	unregister chan *EnhancedClient

	// Channel subscriptions
	subscriptions map[string]map[*EnhancedClient]bool

	// Message history for replay
	messageHistory map[string][]Message
	historyLimit   int

	// Metrics and monitoring
	tracer  trace.Tracer
	metrics *middleware.WebSocketMetrics

	mu sync.RWMutex
}

// NewEnhancedHub creates a new enhanced WebSocket hub
func NewEnhancedHub(ctx context.Context) *EnhancedHub {
	return &EnhancedHub{
		clients:        make(map[*EnhancedClient]bool),
		broadcast:      make(chan Message, 1000),
		register:       make(chan *EnhancedClient),
		unregister:     make(chan *EnhancedClient),
		subscriptions:  make(map[string]map[*EnhancedClient]bool),
		messageHistory: make(map[string][]Message),
		historyLimit:   100,
		tracer:         otel.Tracer("crypto-terminal-websocket"),
		metrics:        middleware.NewWebSocketMetrics(ctx),
	}
}

// Run starts the hub
func (h *EnhancedHub) Run(ctx context.Context) {
	logrus.Info("Starting enhanced WebSocket hub")

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping enhanced WebSocket hub")
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *EnhancedHub) registerClient(client *EnhancedClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	h.metrics.ConnectionOpened()

	logrus.WithFields(logrus.Fields{
		"client_id": client.ID,
		"user_id":   client.UserID,
	}).Info("Client registered")

	// Send welcome message
	welcomeMsg := Message{
		Type:      MessageTypeHeartbeat,
		Data:      json.RawMessage(`{"status":"connected"}`),
		Timestamp: time.Now(),
	}

	select {
	case client.Send <- welcomeMsg:
	default:
		close(client.Send)
		delete(h.clients, client)
		h.metrics.ConnectionClosed()
	}
}

// unregisterClient unregisters a client
func (h *EnhancedHub) unregisterClient(client *EnhancedClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
		h.metrics.ConnectionClosed()

		// Remove from all subscriptions
		for channel, subscribers := range h.subscriptions {
			if _, subscribed := subscribers[client]; subscribed {
				delete(subscribers, client)
				if len(subscribers) == 0 {
					delete(h.subscriptions, channel)
				}
			}
		}

		logrus.WithFields(logrus.Fields{
			"client_id": client.ID,
			"user_id":   client.UserID,
		}).Info("Client unregistered")
	}
}

// broadcastMessage broadcasts a message to subscribed clients
func (h *EnhancedHub) broadcastMessage(message Message) {
	_, span := h.tracer.Start(context.Background(), "websocket.broadcast")
	defer span.End()

	span.SetAttributes(
		attribute.String("message.type", string(message.Type)),
		attribute.String("message.channel", message.Channel),
	)

	h.mu.RLock()
	subscribers, exists := h.subscriptions[message.Channel]
	if !exists {
		h.mu.RUnlock()
		return
	}

	// Store message in history
	h.storeMessageInHistory(message)

	// Broadcast to subscribers
	for client := range subscribers {
		select {
		case client.Send <- message:
			h.metrics.MessageSent(string(message.Type))
		default:
			// Client's send channel is full, remove client
			delete(h.clients, client)
			delete(subscribers, client)
			close(client.Send)
			h.metrics.ConnectionClosed()
		}
	}
	h.mu.RUnlock()

	span.SetAttributes(attribute.Int("subscribers.count", len(subscribers)))
}

// storeMessageInHistory stores a message in the channel history
func (h *EnhancedHub) storeMessageInHistory(message Message) {
	if message.Channel == "" {
		return
	}

	history := h.messageHistory[message.Channel]
	history = append(history, message)

	// Limit history size
	if len(history) > h.historyLimit {
		history = history[len(history)-h.historyLimit:]
	}

	h.messageHistory[message.Channel] = history
}

// Subscribe subscribes a client to a channel
func (h *EnhancedHub) Subscribe(client *EnhancedClient, channel string, filters map[string]string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subscriptions[channel] == nil {
		h.subscriptions[channel] = make(map[*EnhancedClient]bool)
	}

	h.subscriptions[channel][client] = true

	// Add to client's subscriptions
	client.mu.Lock()
	client.Subscriptions[channel] = &Subscription{
		Channel:   channel,
		Filters:   filters,
		CreatedAt: time.Now(),
	}
	client.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"client_id": client.ID,
		"channel":   channel,
		"filters":   filters,
	}).Info("Client subscribed to channel")

	// Send recent message history
	h.sendMessageHistory(client, channel)
}

// Unsubscribe unsubscribes a client from a channel
func (h *EnhancedHub) Unsubscribe(client *EnhancedClient, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subscribers, exists := h.subscriptions[channel]; exists {
		delete(subscribers, client)
		if len(subscribers) == 0 {
			delete(h.subscriptions, channel)
		}
	}

	// Remove from client's subscriptions
	client.mu.Lock()
	delete(client.Subscriptions, channel)
	client.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"client_id": client.ID,
		"channel":   channel,
	}).Info("Client unsubscribed from channel")
}

// sendMessageHistory sends recent message history to a client
func (h *EnhancedHub) sendMessageHistory(client *EnhancedClient, channel string) {
	history, exists := h.messageHistory[channel]
	if !exists || len(history) == 0 {
		return
	}

	// Send last 10 messages
	start := len(history) - 10
	if start < 0 {
		start = 0
	}

	for _, msg := range history[start:] {
		select {
		case client.Send <- msg:
		default:
			// Client's send channel is full, skip history
			return
		}
	}
}

// BroadcastToChannel broadcasts a message to a specific channel
func (h *EnhancedHub) BroadcastToChannel(channel string, messageType MessageType, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("Failed to marshal broadcast data: %v", err)
		return
	}

	message := Message{
		Type:      messageType,
		Channel:   channel,
		Data:      jsonData,
		Timestamp: time.Now(),
	}

	select {
	case h.broadcast <- message:
	default:
		logrus.Warn("Broadcast channel is full, dropping message")
	}
}

// GetClientCount returns the number of connected clients
func (h *EnhancedHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetSubscriptionCount returns the number of active subscriptions
func (h *EnhancedHub) GetSubscriptionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	count := 0
	for _, subscribers := range h.subscriptions {
		count += len(subscribers)
	}
	return count
}

// GetChannels returns a list of active channels
func (h *EnhancedHub) GetChannels() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	channels := make([]string, 0, len(h.subscriptions))
	for channel := range h.subscriptions {
		channels = append(channels, channel)
	}
	return channels
}
