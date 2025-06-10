package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

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

// Client represents a WebSocket client
type Client struct {
	// The WebSocket connection
	conn *websocket.Conn

	// The hub that manages this client
	hub *Hub

	// Buffered channel of outbound messages
	send chan []byte

	// Client ID
	id string

	// User ID (if authenticated)
	userID string

	// Subscribed channels
	channels map[string]bool

	// Last activity time
	lastActivity time.Time
}

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// IncomingMessage represents an incoming WebSocket message
type IncomingMessage struct {
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
	Channel string          `json:"channel,omitempty"`
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.lastActivity = time.Now()
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.ErrorWithFields("WebSocket error", 
					logger.Error(err),
					logger.String("client_id", c.id))
			}
			break
		}

		c.lastActivity = time.Now()
		c.handleIncomingMessage(messageBytes)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
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
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleIncomingMessage handles incoming messages from the client
func (c *Client) handleIncomingMessage(messageBytes []byte) {
	var msg IncomingMessage
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		c.hub.logger.ErrorWithFields("Failed to unmarshal incoming message", 
			logger.Error(err),
			logger.String("client_id", c.id))
		c.sendError("Invalid message format")
		return
	}

	c.hub.logger.InfoWithFields("Incoming WebSocket message", 
		logger.String("client_id", c.id),
		logger.String("user_id", c.userID),
		logger.String("type", msg.Type))

	switch msg.Type {
	case "ping":
		c.handlePing()
	case "subscribe":
		c.handleSubscribe(msg.Data)
	case "unsubscribe":
		c.handleUnsubscribe(msg.Data)
	case "get_stats":
		c.handleGetStats()
	case "get_user_sessions":
		c.handleGetUserSessions()
	case "get_security_events":
		c.handleGetSecurityEvents()
	default:
		c.sendError(fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

// handlePing handles ping messages
func (c *Client) handlePing() {
	response := Message{
		Type:      "pong",
		Data:      map[string]string{"status": "ok"},
		Timestamp: time.Now(),
	}
	c.sendMessage(response)
}

// handleSubscribe handles subscription requests
func (c *Client) handleSubscribe(data json.RawMessage) {
	var subscribeData struct {
		Channel string `json:"channel"`
	}

	if err := json.Unmarshal(data, &subscribeData); err != nil {
		c.sendError("Invalid subscribe data")
		return
	}

	if subscribeData.Channel == "" {
		c.sendError("Channel is required for subscription")
		return
	}

	// Add channel to client's subscriptions
	c.channels[subscribeData.Channel] = true

	response := Message{
		Type: "subscribed",
		Data: map[string]string{
			"channel": subscribeData.Channel,
			"status":  "subscribed",
		},
		Timestamp: time.Now(),
	}
	c.sendMessage(response)

	c.hub.logger.InfoWithFields("Client subscribed to channel", 
		logger.String("client_id", c.id),
		logger.String("channel", subscribeData.Channel))
}

// handleUnsubscribe handles unsubscription requests
func (c *Client) handleUnsubscribe(data json.RawMessage) {
	var unsubscribeData struct {
		Channel string `json:"channel"`
	}

	if err := json.Unmarshal(data, &unsubscribeData); err != nil {
		c.sendError("Invalid unsubscribe data")
		return
	}

	if unsubscribeData.Channel == "" {
		c.sendError("Channel is required for unsubscription")
		return
	}

	// Remove channel from client's subscriptions
	delete(c.channels, unsubscribeData.Channel)

	response := Message{
		Type: "unsubscribed",
		Data: map[string]string{
			"channel": unsubscribeData.Channel,
			"status":  "unsubscribed",
		},
		Timestamp: time.Now(),
	}
	c.sendMessage(response)

	c.hub.logger.InfoWithFields("Client unsubscribed from channel", 
		logger.String("client_id", c.id),
		logger.String("channel", unsubscribeData.Channel))
}

// handleGetStats handles stats requests
func (c *Client) handleGetStats() {
	stats := c.hub.GetStats()
	
	response := Message{
		Type:      "stats",
		Data:      stats,
		Timestamp: time.Now(),
	}
	c.sendMessage(response)
}

// handleGetUserSessions handles user sessions requests
func (c *Client) handleGetUserSessions() {
	if c.userID == "" {
		c.sendError("Authentication required")
		return
	}

	// This would typically call the query bus to get user sessions
	// For now, return a placeholder response
	response := Message{
		Type: "user_sessions",
		Data: map[string]interface{}{
			"sessions": []interface{}{}, // Placeholder
			"count":    0,
		},
		Timestamp: time.Now(),
	}
	c.sendMessage(response)
}

// handleGetSecurityEvents handles security events requests
func (c *Client) handleGetSecurityEvents() {
	if c.userID == "" {
		c.sendError("Authentication required")
		return
	}

	// This would typically call the query bus to get security events
	// For now, return a placeholder response
	response := Message{
		Type: "security_events",
		Data: map[string]interface{}{
			"events": []interface{}{}, // Placeholder
			"count":  0,
		},
		Timestamp: time.Now(),
	}
	c.sendMessage(response)
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(message Message) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		c.hub.logger.ErrorWithFields("Failed to marshal outgoing message", 
			logger.Error(err),
			logger.String("client_id", c.id))
		return
	}

	select {
	case c.send <- messageBytes:
	default:
		// Client's send channel is full, close the connection
		close(c.send)
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errorMsg string) {
	errorResponse := Message{
		Type: "error",
		Data: map[string]string{
			"error": errorMsg,
		},
		Timestamp: time.Now(),
	}
	c.sendMessage(errorResponse)
}

// IsSubscribedTo checks if the client is subscribed to a channel
func (c *Client) IsSubscribedTo(channel string) bool {
	return c.channels[channel]
}

// GetInfo returns client information
func (c *Client) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":            c.id,
		"user_id":       c.userID,
		"channels":      c.getChannelList(),
		"last_activity": c.lastActivity,
		"connected_at":  time.Now(), // This should be set when client connects
	}
}

// getChannelList returns a list of subscribed channels
func (c *Client) getChannelList() []string {
	channels := make([]string, 0, len(c.channels))
	for channel := range c.channels {
		channels = append(channels, channel)
	}
	return channels
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return uuid.New().String()
}

// Notification types for different auth events
const (
	NotificationTypeLogin          = "login"
	NotificationTypeLogout         = "logout"
	NotificationTypeSessionExpired = "session_expired"
	NotificationTypePasswordChange = "password_change"
	NotificationTypeMFAEnabled     = "mfa_enabled"
	NotificationTypeSecurityAlert  = "security_alert"
)

// SendNotification sends a typed notification to the client
func (c *Client) SendNotification(notificationType string, data interface{}) {
	notification := Message{
		Type:      "notification",
		Data: map[string]interface{}{
			"type": notificationType,
			"data": data,
		},
		Timestamp: time.Now(),
	}
	c.sendMessage(notification)
}
