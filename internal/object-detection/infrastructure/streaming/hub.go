package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub manages WebSocket connections and message broadcasting
type Hub struct {
	logger      *zap.Logger
	clients     map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	upgrader    websocket.Upgrader
	config      HubConfig
	isRunning   bool
	mutex       sync.RWMutex
	stats       *HubStats
}

// HubConfig configures the WebSocket hub
type HubConfig struct {
	ReadBufferSize    int           // WebSocket read buffer size
	WriteBufferSize   int           // WebSocket write buffer size
	WriteTimeout      time.Duration // Write timeout for clients
	PongTimeout       time.Duration // Pong timeout for ping/pong
	PingPeriod        time.Duration // Ping period for keep-alive
	MaxMessageSize    int64         // Maximum message size
	AllowedOrigins    []string      // Allowed origins for CORS
	MaxConnections    int           // Maximum concurrent connections
	EnableCompression bool          // Enable WebSocket compression
}

// HubStats tracks hub performance metrics
type HubStats struct {
	ConnectedClients   int
	TotalConnections   int64
	MessagesSent       int64
	MessagesReceived   int64
	BytesSent          int64
	BytesReceived      int64
	ConnectionErrors   int64
	BroadcastErrors    int64
	StartTime          time.Time
	LastActivity       time.Time
	mutex              sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      MessageType `json:"type"`
	StreamID  string      `json:"stream_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// MessageType defines the type of WebSocket message
type MessageType string

const (
	MessageTypeDetection    MessageType = "detection"
	MessageTypeTracking     MessageType = "tracking"
	MessageTypeFrame        MessageType = "frame"
	MessageTypeStatus       MessageType = "status"
	MessageTypeError        MessageType = "error"
	MessageTypeSubscribe    MessageType = "subscribe"
	MessageTypeUnsubscribe  MessageType = "unsubscribe"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
)

// DefaultHubConfig returns default hub configuration
func DefaultHubConfig() HubConfig {
	return HubConfig{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		WriteTimeout:      10 * time.Second,
		PongTimeout:       60 * time.Second,
		PingPeriod:        54 * time.Second,
		MaxMessageSize:    512 * 1024, // 512KB
		AllowedOrigins:    []string{"*"},
		MaxConnections:    1000,
		EnableCompression: true,
	}
}

// NewHub creates a new WebSocket hub
func NewHub(logger *zap.Logger, config HubConfig) *Hub {
	upgrader := websocket.Upgrader{
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		EnableCompression: config.EnableCompression,
		CheckOrigin: func(r *http.Request) bool {
			// Check allowed origins
			origin := r.Header.Get("Origin")
			if len(config.AllowedOrigins) == 0 {
				return true
			}
			
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					return true
				}
			}
			return false
		},
	}

	return &Hub{
		logger:     logger.With(zap.String("component", "websocket_hub")),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		upgrader:   upgrader,
		config:     config,
		stats: &HubStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the WebSocket hub
func (h *Hub) Start(ctx context.Context) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.isRunning {
		return fmt.Errorf("hub is already running")
	}

	h.logger.Info("Starting WebSocket hub",
		zap.Int("max_connections", h.config.MaxConnections),
		zap.Duration("ping_period", h.config.PingPeriod),
		zap.Int64("max_message_size", h.config.MaxMessageSize))

	h.isRunning = true
	h.stats.StartTime = time.Now()

	// Start hub goroutine
	go h.run(ctx)

	h.logger.Info("WebSocket hub started")
	return nil
}

// Stop stops the WebSocket hub
func (h *Hub) Stop() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.isRunning {
		return fmt.Errorf("hub is not running")
	}

	h.logger.Info("Stopping WebSocket hub")

	// Close all client connections
	for client := range h.clients {
		client.Close()
	}

	h.isRunning = false

	h.logger.Info("WebSocket hub stopped")
	return nil
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !h.IsRunning() {
		http.Error(w, "WebSocket hub not running", http.StatusServiceUnavailable)
		return
	}

	// Check connection limit
	h.mutex.RLock()
	clientCount := len(h.clients)
	h.mutex.RUnlock()

	if clientCount >= h.config.MaxConnections {
		http.Error(w, "Maximum connections reached", http.StatusTooManyRequests)
		return
	}

	// Upgrade connection
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		h.updateStats(func(stats *HubStats) {
			stats.ConnectionErrors++
		})
		return
	}

	// Create client
	client := NewClient(h, conn, r)

	// Register client
	h.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()

	h.logger.Info("WebSocket connection established",
		zap.String("client_id", client.ID),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()))
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message *Message) error {
	if !h.IsRunning() {
		return fmt.Errorf("hub is not running")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	select {
	case h.broadcast <- data:
		h.updateStats(func(stats *HubStats) {
			stats.MessagesSent++
			stats.BytesSent += int64(len(data))
			stats.LastActivity = time.Now()
		})
		return nil
	default:
		h.updateStats(func(stats *HubStats) {
			stats.BroadcastErrors++
		})
		return fmt.Errorf("broadcast channel full")
	}
}

// BroadcastToStream sends a message to clients subscribed to a specific stream
func (h *Hub) BroadcastToStream(streamID string, message *Message) error {
	if !h.IsRunning() {
		return fmt.Errorf("hub is not running")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	sentCount := 0
	for client := range h.clients {
		if client.IsSubscribedToStream(streamID) {
			select {
			case client.send <- data:
				sentCount++
			default:
				// Client send channel is full, close the client
				h.logger.Warn("Client send channel full, closing connection",
					zap.String("client_id", client.ID))
				client.Close()
			}
		}
	}

	h.updateStats(func(stats *HubStats) {
		stats.MessagesSent += int64(sentCount)
		stats.BytesSent += int64(len(data) * sentCount)
		stats.LastActivity = time.Now()
	})

	h.logger.Debug("Broadcast to stream completed",
		zap.String("stream_id", streamID),
		zap.String("message_type", string(message.Type)),
		zap.Int("clients_sent", sentCount))

	return nil
}

// GetConnectedClients returns the number of connected clients
func (h *Hub) GetConnectedClients() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// GetStats returns hub statistics
func (h *Hub) GetStats() *HubStats {
	h.stats.mutex.RLock()
	defer h.stats.mutex.RUnlock()

	h.mutex.RLock()
	connectedClients := len(h.clients)
	h.mutex.RUnlock()

	return &HubStats{
		ConnectedClients:   connectedClients,
		TotalConnections:   h.stats.TotalConnections,
		MessagesSent:       h.stats.MessagesSent,
		MessagesReceived:   h.stats.MessagesReceived,
		BytesSent:          h.stats.BytesSent,
		BytesReceived:      h.stats.BytesReceived,
		ConnectionErrors:   h.stats.ConnectionErrors,
		BroadcastErrors:    h.stats.BroadcastErrors,
		StartTime:          h.stats.StartTime,
		LastActivity:       h.stats.LastActivity,
	}
}

// IsRunning returns whether the hub is running
func (h *Hub) IsRunning() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.isRunning
}

// run is the main hub loop
func (h *Hub) run(ctx context.Context) {
	ticker := time.NewTicker(h.config.PingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Hub context cancelled, stopping")
			return

		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()

			h.updateStats(func(stats *HubStats) {
				stats.TotalConnections++
			})

			h.logger.Info("Client registered",
				zap.String("client_id", client.ID),
				zap.Int("total_clients", len(h.clients)))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()

			h.logger.Info("Client unregistered",
				zap.String("client_id", client.ID),
				zap.Int("total_clients", len(h.clients)))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client send channel is full, close the client
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mutex.RUnlock()

		case <-ticker.C:
			// Send ping to all clients
			h.sendPingToAllClients()
		}
	}
}

// sendPingToAllClients sends ping messages to all connected clients
func (h *Hub) sendPingToAllClients() {
	pingMessage := &Message{
		Type:      MessageTypePing,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(pingMessage)
	if err != nil {
		h.logger.Error("Failed to marshal ping message", zap.Error(err))
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			// Client is not responsive, close it
			h.logger.Warn("Client not responsive to ping, closing",
				zap.String("client_id", client.ID))
			client.Close()
		}
	}
}

// updateStats updates hub statistics
func (h *Hub) updateStats(updateFunc func(*HubStats)) {
	h.stats.mutex.Lock()
	defer h.stats.mutex.Unlock()
	updateFunc(h.stats)
}

// GetConfig returns the hub configuration
func (h *Hub) GetConfig() HubConfig {
	return h.config
}

// UpdateConfig updates the hub configuration
func (h *Hub) UpdateConfig(config HubConfig) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.config = config
	h.logger.Info("Hub configuration updated",
		zap.Int("max_connections", config.MaxConnections),
		zap.Duration("ping_period", config.PingPeriod))
}
