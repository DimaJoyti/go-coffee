package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/domain/services"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Analytics service for real-time data
	analyticsService *services.AnalyticsService

	// Logger
	logger Logger

	// Mutex for thread safety
	mutex sync.RWMutex

	// Stop channel
	stopChan chan struct{}
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// Client represents a WebSocket client
type Client struct {
	// The websocket connection
	conn WebSocketConnection

	// Buffered channel of outbound messages
	send chan []byte

	// Client ID
	id string

	// Subscriptions
	subscriptions map[string]bool

	// Last activity
	lastActivity time.Time

	// Hub reference
	hub *Hub
}

// WebSocketConnection interface for WebSocket connections
type WebSocketConnection interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	ClientID  string                 `json:"client_id,omitempty"`
}

// Subscription represents a client subscription
type Subscription struct {
	Channel string `json:"channel"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// NewHub creates a new WebSocket hub
func NewHub(analyticsService *services.AnalyticsService, logger Logger) *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		broadcast:        make(chan []byte, 256),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		analyticsService: analyticsService,
		logger:           logger,
		stopChan:         make(chan struct{}),
	}
}

// Start starts the WebSocket hub
func (h *Hub) Start(ctx context.Context) {
	h.logger.Info("Starting WebSocket hub")

	go h.run(ctx)
	go h.broadcastMetrics(ctx)
	go h.broadcastAlerts(ctx)

	h.logger.Info("WebSocket hub started")
}

// Stop stops the WebSocket hub
func (h *Hub) Stop() {
	h.logger.Info("Stopping WebSocket hub")
	close(h.stopChan)
	
	// Close all client connections
	h.mutex.Lock()
	for client := range h.clients {
		close(client.send)
		client.conn.Close()
	}
	h.mutex.Unlock()
	
	h.logger.Info("WebSocket hub stopped")
}

// run handles the main hub loop
func (h *Hub) run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopChan:
			return
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		case <-ticker.C:
			h.cleanupInactiveClients()
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	h.clients[client] = true
	h.mutex.Unlock()

	h.logger.Info("Client registered", "client_id", client.id, "total_clients", len(h.clients))

	// Send welcome message
	welcome := Message{
		Type:      "welcome",
		Data:      map[string]interface{}{"client_id": client.id},
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(welcome); err == nil {
		select {
		case client.send <- data:
		default:
			h.unregisterClient(client)
		}
	}
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
		client.conn.Close()
	}
	h.mutex.Unlock()

	h.logger.Info("Client unregistered", "client_id", client.id, "total_clients", len(h.clients))
}

// broadcastMessage broadcasts a message to all subscribed clients
func (h *Hub) broadcastMessage(message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		h.logger.Error("Failed to unmarshal broadcast message", err)
		return
	}

	h.mutex.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mutex.RUnlock()

	for _, client := range clients {
		if h.shouldSendToClient(client, &msg) {
			select {
			case client.send <- message:
			default:
				h.unregisterClient(client)
			}
		}
	}
}

// shouldSendToClient determines if a message should be sent to a specific client
func (h *Hub) shouldSendToClient(client *Client, msg *Message) bool {
	if msg.Channel == "" {
		return true // Broadcast to all clients
	}

	return client.subscriptions[msg.Channel]
}

// broadcastMetrics broadcasts real-time metrics
func (h *Hub) broadcastMetrics(ctx context.Context) {
	if h.analyticsService == nil {
		return
	}

	metricsStream := h.analyticsService.GetRealTimeMetrics()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopChan:
			return
		case update := <-metricsStream:
			if update == nil {
				continue
			}

			message := Message{
				Type:      "metrics_update",
				Channel:   "metrics",
				Data:      update.Data,
				Timestamp: update.Timestamp,
			}

			if data, err := json.Marshal(message); err == nil {
				select {
				case h.broadcast <- data:
				default:
					// Broadcast channel full, skip this update
				}
			}
		}
	}
}

// broadcastAlerts broadcasts real-time alerts
func (h *Hub) broadcastAlerts(ctx context.Context) {
	if h.analyticsService == nil {
		return
	}

	alertsStream := h.analyticsService.GetAlerts()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopChan:
			return
		case alert := <-alertsStream:
			if alert == nil {
				continue
			}

			message := Message{
				Type:    "alert",
				Channel: "alerts",
				Data: map[string]interface{}{
					"id":       alert.ID,
					"type":     alert.Type,
					"severity": alert.Severity,
					"title":    alert.Title,
					"message":  alert.Message,
					"source":   alert.Source,
					"data":     alert.Data,
				},
				Timestamp: alert.Timestamp,
			}

			if data, err := json.Marshal(message); err == nil {
				select {
				case h.broadcast <- data:
				default:
					// Broadcast channel full, skip this alert
				}
			}
		}
	}
}

// cleanupInactiveClients removes inactive clients
func (h *Hub) cleanupInactiveClients() {
	cutoff := time.Now().Add(-5 * time.Minute)

	h.mutex.Lock()
	for client := range h.clients {
		if client.lastActivity.Before(cutoff) {
			delete(h.clients, client)
			close(client.send)
			client.conn.Close()
			h.logger.Info("Removed inactive client", "client_id", client.id)
		}
	}
	h.mutex.Unlock()
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// BroadcastMessage broadcasts a message to all clients
func (h *Hub) BroadcastMessage(msgType, channel string, data map[string]interface{}) {
	message := Message{
		Type:      msgType,
		Channel:   channel,
		Data:      data,
		Timestamp: time.Now(),
	}

	if msgData, err := json.Marshal(message); err == nil {
		select {
		case h.broadcast <- msgData:
		default:
			h.logger.Warn("Broadcast channel full, message dropped")
		}
	}
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn WebSocketConnection, clientID string) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		id:            clientID,
		subscriptions: make(map[string]bool),
		lastActivity:  time.Now(),
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if !isExpectedCloseError(err) {
				c.hub.logger.Error("WebSocket read error", err, "client_id", c.id)
			}
			break
		}

		c.lastActivity = time.Now()
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(8, []byte{}) // Close message
				return
			}

			if err := c.conn.WriteMessage(1, message); err != nil { // Text message
				c.hub.logger.Error("WebSocket write error", err, "client_id", c.id)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(9, nil); err != nil { // Ping message
				return
			}
		}
	}
}

// handleMessage handles incoming messages from the client
func (c *Client) handleMessage(message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		c.hub.logger.Error("Failed to unmarshal client message", err, "client_id", c.id)
		return
	}

	switch msg.Type {
	case "subscribe":
		c.handleSubscribe(&msg)
	case "unsubscribe":
		c.handleUnsubscribe(&msg)
	case "ping":
		c.handlePing()
	default:
		c.hub.logger.Warn("Unknown message type", "type", msg.Type, "client_id", c.id)
	}
}

// handleSubscribe handles subscription requests
func (c *Client) handleSubscribe(msg *Message) {
	channel, ok := msg.Data["channel"].(string)
	if !ok {
		c.hub.logger.Warn("Invalid subscribe message: missing channel", "client_id", c.id)
		return
	}

	c.subscriptions[channel] = true
	c.hub.logger.Info("Client subscribed", "client_id", c.id, "channel", channel)

	// Send subscription confirmation
	response := Message{
		Type:      "subscribed",
		Channel:   channel,
		Data:      map[string]interface{}{"status": "success"},
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(response); err == nil {
		select {
		case c.send <- data:
		default:
			// Send channel full
		}
	}
}

// handleUnsubscribe handles unsubscription requests
func (c *Client) handleUnsubscribe(msg *Message) {
	channel, ok := msg.Data["channel"].(string)
	if !ok {
		c.hub.logger.Warn("Invalid unsubscribe message: missing channel", "client_id", c.id)
		return
	}

	delete(c.subscriptions, channel)
	c.hub.logger.Info("Client unsubscribed", "client_id", c.id, "channel", channel)

	// Send unsubscription confirmation
	response := Message{
		Type:      "unsubscribed",
		Channel:   channel,
		Data:      map[string]interface{}{"status": "success"},
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(response); err == nil {
		select {
		case c.send <- data:
		default:
			// Send channel full
		}
	}
}

// handlePing handles ping messages
func (c *Client) handlePing() {
	response := Message{
		Type:      "pong",
		Data:      map[string]interface{}{"timestamp": time.Now()},
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(response); err == nil {
		select {
		case c.send <- data:
		default:
			// Send channel full
		}
	}
}

// Start starts the client's read and write pumps
func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
}

// isExpectedCloseError checks if the error is an expected WebSocket close error
func isExpectedCloseError(err error) bool {
	// This is a simplified check - in a real implementation,
	// you would check for specific WebSocket close codes
	return err != nil && (err.Error() == "websocket: close 1001 (going away)" ||
		err.Error() == "websocket: close 1000 (normal)")
}

// WebSocketHandler handles WebSocket upgrade requests
func (h *Hub) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, you would use a WebSocket library like gorilla/websocket
	// For now, we'll just return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"message": "WebSocket endpoint - use a WebSocket client to connect",
		"endpoints": map[string]string{
			"websocket": "ws://localhost:8080/ws",
		},
		"channels": []string{
			"metrics",
			"alerts",
			"workflows",
			"executions",
			"agents",
		},
		"example_subscribe": map[string]interface{}{
			"type": "subscribe",
			"data": map[string]string{
				"channel": "metrics",
			},
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// GetStats returns hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	subscriptions := make(map[string]int)
	for client := range h.clients {
		for channel := range client.subscriptions {
			subscriptions[channel]++
		}
	}

	return map[string]interface{}{
		"connected_clients": len(h.clients),
		"subscriptions":     subscriptions,
		"uptime":           time.Since(time.Now().Add(-time.Hour)).String(), // Mock uptime
	}
}
