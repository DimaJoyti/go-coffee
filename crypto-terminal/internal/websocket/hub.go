package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	config *config.Config

	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Channel subscriptions
	subscriptions map[string]map[*Client]bool

	// Mutex for thread safety
	mu sync.RWMutex

	// WebSocket upgrader
	upgrader websocket.Upgrader
}

// Client represents a WebSocket client
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Client ID
	id string

	// User ID
	userID string

	// Subscribed channels
	channels map[string]bool

	// Last activity time
	lastActivity time.Time
}

// NewHub creates a new WebSocket hub
func NewHub(config *config.Config) *Hub {
	return &Hub{
		config:        config,
		clients:       make(map[*Client]bool),
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		subscriptions: make(map[string]map[*Client]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.WebSocket.ReadBufferSize,
			WriteBufferSize: config.WebSocket.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return config.WebSocket.CheckOrigin
			},
		},
	}
}

// Run starts the WebSocket hub
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
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

// HandleWebSocket handles general WebSocket connections
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	client := &Client{
		hub:          h,
		conn:         conn,
		send:         make(chan []byte, 256),
		id:           generateClientID(),
		userID:       userID,
		channels:     make(map[string]bool),
		lastActivity: time.Now(),
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// HandleMarketWebSocket handles market data WebSocket connections
func (h *Hub) HandleMarketWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("Failed to upgrade market WebSocket connection: %v", err)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	client := &Client{
		hub:          h,
		conn:         conn,
		send:         make(chan []byte, 256),
		id:           generateClientID(),
		userID:       userID,
		channels:     map[string]bool{"market": true},
		lastActivity: time.Now(),
	}

	client.hub.register <- client
	h.subscribeToChannel("market", client)

	// Send initial market data
	h.sendInitialMarketData(client)

	go client.writePump()
	go client.readPump()
}

// HandlePortfolioWebSocket handles portfolio WebSocket connections
func (h *Hub) HandlePortfolioWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("Failed to upgrade portfolio WebSocket connection: %v", err)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	client := &Client{
		hub:          h,
		conn:         conn,
		send:         make(chan []byte, 256),
		id:           generateClientID(),
		userID:       userID,
		channels:     map[string]bool{"portfolio": true},
		lastActivity: time.Now(),
	}

	client.hub.register <- client
	h.subscribeToChannel("portfolio", client)

	go client.writePump()
	go client.readPump()
}

// HandleAlertsWebSocket handles alerts WebSocket connections
func (h *Hub) HandleAlertsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("Failed to upgrade alerts WebSocket connection: %v", err)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	client := &Client{
		hub:          h,
		conn:         conn,
		send:         make(chan []byte, 256),
		id:           generateClientID(),
		userID:       userID,
		channels:     map[string]bool{"alerts": true},
		lastActivity: time.Now(),
	}

	client.hub.register <- client
	h.subscribeToChannel("alerts", client)

	go client.writePump()
	go client.readPump()
}

// BroadcastToChannel broadcasts a message to all clients subscribed to a channel
func (h *Hub) BroadcastToChannel(channel string, message interface{}) {
	h.mu.RLock()
	clients, exists := h.subscriptions[channel]
	h.mu.RUnlock()

	if !exists {
		return
	}

	wsMessage := models.WebSocketMessage{
		Type:      "broadcast",
		Channel:   channel,
		Data:      message,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(wsMessage)
	if err != nil {
		logrus.Errorf("Failed to marshal WebSocket message: %v", err)
		return
	}

	h.mu.RLock()
	for client := range clients {
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(h.clients, client)
			delete(clients, client)
		}
	}
	h.mu.RUnlock()
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"client_id": client.id,
		"user_id":   client.userID,
	}).Info("Client connected")
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		// Remove from all channel subscriptions
		for channel := range h.subscriptions {
			delete(h.subscriptions[channel], client)
		}
	}
	h.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"client_id": client.id,
		"user_id":   client.userID,
	}).Info("Client disconnected")
}

// broadcastMessage broadcasts a message to all clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
	h.mu.RUnlock()
}

// subscribeToChannel subscribes a client to a channel
func (h *Hub) subscribeToChannel(channel string, client *Client) {
	h.mu.Lock()
	if h.subscriptions[channel] == nil {
		h.subscriptions[channel] = make(map[*Client]bool)
	}
	h.subscriptions[channel][client] = true
	client.channels[channel] = true
	h.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"client_id": client.id,
		"channel":   channel,
	}).Debug("Client subscribed to channel")
}

// unsubscribeFromChannel unsubscribes a client from a channel
func (h *Hub) unsubscribeFromChannel(channel string, client *Client) {
	h.mu.Lock()
	if clients, exists := h.subscriptions[channel]; exists {
		delete(clients, client)
	}
	delete(client.channels, channel)
	h.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"client_id": client.id,
		"channel":   channel,
	}).Debug("Client unsubscribed from channel")
}

// sendInitialMarketData sends initial market data to a client
func (h *Hub) sendInitialMarketData(client *Client) {
	// Send welcome message with initial data
	welcomeMessage := models.WebSocketMessage{
		Type:      "welcome",
		Channel:   "market",
		Data:      map[string]interface{}{"message": "Connected to market data feed"},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(welcomeMessage)
	if err != nil {
		logrus.Errorf("Failed to marshal welcome message: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
