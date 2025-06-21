package streaming

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	subscriptions map[string]bool // Stream subscriptions
	metadata      ClientMetadata
	isActive      bool
	lastPong      time.Time
	mutex         sync.RWMutex
	logger        *zap.Logger
}

// ClientMetadata contains client information
type ClientMetadata struct {
	UserAgent    string            `json:"user_agent"`
	RemoteAddr   string            `json:"remote_addr"`
	ConnectedAt  time.Time         `json:"connected_at"`
	LastActivity time.Time         `json:"last_activity"`
	Headers      map[string]string `json:"headers"`
	QueryParams  map[string]string `json:"query_params"`
}

// SubscriptionRequest represents a stream subscription request
type SubscriptionRequest struct {
	Action   string `json:"action"`   // "subscribe" or "unsubscribe"
	StreamID string `json:"stream_id"`
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, r *http.Request) *Client {
	clientID := uuid.New().String()

	// Extract headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Extract query parameters
	queryParams := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	metadata := ClientMetadata{
		UserAgent:    r.UserAgent(),
		RemoteAddr:   r.RemoteAddr,
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		Headers:      headers,
		QueryParams:  queryParams,
	}

	client := &Client{
		ID:            clientID,
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
		metadata:      metadata,
		isActive:      true,
		lastPong:      time.Now(),
		logger:        hub.logger.With(zap.String("client_id", clientID)),
	}

	// Set connection parameters
	conn.SetReadLimit(hub.config.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(hub.config.PongTimeout))
	conn.SetPongHandler(func(string) error {
		client.mutex.Lock()
		client.lastPong = time.Now()
		client.metadata.LastActivity = time.Now()
		client.mutex.Unlock()
		
		conn.SetReadDeadline(time.Now().Add(hub.config.PongTimeout))
		return nil
	})

	return client
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket read error", zap.Error(err))
			}
			break
		}

		c.mutex.Lock()
		c.metadata.LastActivity = time.Now()
		c.mutex.Unlock()

		// Update hub stats
		c.hub.updateStats(func(stats *HubStats) {
			stats.MessagesReceived++
			stats.BytesReceived += int64(len(messageData))
		})

		// Process message
		c.processMessage(messageData)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(c.hub.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteTimeout))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processMessage processes incoming messages from the client
func (c *Client) processMessage(messageData []byte) {
	var message Message
	if err := json.Unmarshal(messageData, &message); err != nil {
		c.logger.Error("Failed to unmarshal message", zap.Error(err))
		c.sendError("Invalid message format")
		return
	}

	c.logger.Debug("Received message",
		zap.String("type", string(message.Type)),
		zap.String("stream_id", message.StreamID))

	switch message.Type {
	case MessageTypeSubscribe:
		c.handleSubscription(message, true)
	case MessageTypeUnsubscribe:
		c.handleSubscription(message, false)
	case MessageTypePong:
		c.handlePong()
	default:
		c.logger.Warn("Unknown message type", zap.String("type", string(message.Type)))
		c.sendError("Unknown message type")
	}
}

// handleSubscription handles stream subscription/unsubscription
func (c *Client) handleSubscription(message Message, subscribe bool) {
	streamID := message.StreamID
	if streamID == "" {
		c.sendError("Stream ID is required for subscription")
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if subscribe {
		c.subscriptions[streamID] = true
		c.logger.Info("Client subscribed to stream", zap.String("stream_id", streamID))
		
		// Send confirmation
		c.sendMessage(&Message{
			Type:      MessageTypeStatus,
			StreamID:  streamID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"status":  "subscribed",
				"message": "Successfully subscribed to stream",
			},
		})
	} else {
		delete(c.subscriptions, streamID)
		c.logger.Info("Client unsubscribed from stream", zap.String("stream_id", streamID))
		
		// Send confirmation
		c.sendMessage(&Message{
			Type:      MessageTypeStatus,
			StreamID:  streamID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"status":  "unsubscribed",
				"message": "Successfully unsubscribed from stream",
			},
		})
	}
}

// handlePong handles pong messages
func (c *Client) handlePong() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.lastPong = time.Now()
	c.metadata.LastActivity = time.Now()
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(message *Message) {
	if !c.isActive {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		c.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	select {
	case c.send <- data:
	default:
		c.logger.Warn("Client send channel full, closing connection")
		c.Close()
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errorMsg string) {
	c.sendMessage(&Message{
		Type:      MessageTypeError,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"error": errorMsg,
		},
	})
}

// IsSubscribedToStream checks if the client is subscribed to a stream
func (c *Client) IsSubscribedToStream(streamID string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.subscriptions[streamID]
}

// GetSubscriptions returns all stream subscriptions
func (c *Client) GetSubscriptions() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	subscriptions := make([]string, 0, len(c.subscriptions))
	for streamID := range c.subscriptions {
		subscriptions = append(subscriptions, streamID)
	}
	return subscriptions
}

// GetMetadata returns client metadata
func (c *Client) GetMetadata() ClientMetadata {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.metadata
}

// IsActive returns whether the client is active
func (c *Client) IsActive() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isActive
}

// GetLastPong returns the last pong time
func (c *Client) GetLastPong() time.Time {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.lastPong
}

// Close closes the client connection
func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isActive {
		return
	}

	c.isActive = false
	c.logger.Info("Closing client connection")

	// Close the connection
	c.conn.Close()
}

// GetConnectionDuration returns how long the client has been connected
func (c *Client) GetConnectionDuration() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return time.Since(c.metadata.ConnectedAt)
}

// GetIdleDuration returns how long the client has been idle
func (c *Client) GetIdleDuration() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return time.Since(c.metadata.LastActivity)
}

// GetInfo returns client information
func (c *Client) GetInfo() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	subscriptions := make([]string, 0, len(c.subscriptions))
	for streamID := range c.subscriptions {
		subscriptions = append(subscriptions, streamID)
	}

	return map[string]interface{}{
		"id":                c.ID,
		"is_active":         c.isActive,
		"connected_at":      c.metadata.ConnectedAt,
		"last_activity":     c.metadata.LastActivity,
		"last_pong":         c.lastPong,
		"connection_duration": time.Since(c.metadata.ConnectedAt),
		"idle_duration":     time.Since(c.metadata.LastActivity),
		"subscriptions":     subscriptions,
		"remote_addr":       c.metadata.RemoteAddr,
		"user_agent":        c.metadata.UserAgent,
	}
}
