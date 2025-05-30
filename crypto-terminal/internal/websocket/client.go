package websocket

import (
	"encoding/json"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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

// readPump pumps messages from the websocket connection to the hub
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
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket error: %v", err)
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
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
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

// handleMessage handles incoming messages from the client
func (c *Client) handleMessage(message []byte) {
	var wsMessage models.WebSocketMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		logrus.Errorf("Failed to unmarshal WebSocket message: %v", err)
		return
	}

	switch wsMessage.Type {
	case "subscribe":
		c.handleSubscription(&wsMessage)
	case "unsubscribe":
		c.handleUnsubscription(&wsMessage)
	case "ping":
		c.handlePing()
	default:
		logrus.Warnf("Unknown message type: %s", wsMessage.Type)
	}
}

// handleSubscription handles subscription requests
func (c *Client) handleSubscription(message *models.WebSocketMessage) {
	var subscription models.WebSocketSubscription
	data, err := json.Marshal(message.Data)
	if err != nil {
		logrus.Errorf("Failed to marshal subscription data: %v", err)
		return
	}

	if err := json.Unmarshal(data, &subscription); err != nil {
		logrus.Errorf("Failed to unmarshal subscription: %v", err)
		return
	}

	// Subscribe to the requested channel
	c.hub.subscribeToChannel(subscription.Channel, c)

	// Send confirmation
	response := models.WebSocketMessage{
		Type:      "subscription_confirmed",
		Channel:   subscription.Channel,
		Data:      map[string]interface{}{"status": "subscribed"},
		Timestamp: time.Now(),
	}

	c.sendMessage(response)
}

// handleUnsubscription handles unsubscription requests
func (c *Client) handleUnsubscription(message *models.WebSocketMessage) {
	var subscription models.WebSocketSubscription
	data, err := json.Marshal(message.Data)
	if err != nil {
		logrus.Errorf("Failed to marshal unsubscription data: %v", err)
		return
	}

	if err := json.Unmarshal(data, &subscription); err != nil {
		logrus.Errorf("Failed to unmarshal unsubscription: %v", err)
		return
	}

	// Unsubscribe from the channel
	c.hub.unsubscribeFromChannel(subscription.Channel, c)

	// Send confirmation
	response := models.WebSocketMessage{
		Type:      "unsubscription_confirmed",
		Channel:   subscription.Channel,
		Data:      map[string]interface{}{"status": "unsubscribed"},
		Timestamp: time.Now(),
	}

	c.sendMessage(response)
}

// handlePing handles ping messages
func (c *Client) handlePing() {
	response := models.WebSocketMessage{
		Type:      "pong",
		Channel:   "",
		Data:      map[string]interface{}{"timestamp": time.Now()},
		Timestamp: time.Now(),
	}

	c.sendMessage(response)
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(message models.WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("Failed to marshal message: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		delete(c.hub.clients, c)
	}
}

// IsActive returns whether the client is still active
func (c *Client) IsActive() bool {
	return time.Since(c.lastActivity) < 5*time.Minute
}

// GetChannels returns the channels the client is subscribed to
func (c *Client) GetChannels() []string {
	var channels []string
	for channel := range c.channels {
		channels = append(channels, channel)
	}
	return channels
}

// GetUserID returns the user ID of the client
func (c *Client) GetUserID() string {
	return c.userID
}

// GetID returns the client ID
func (c *Client) GetID() string {
	return c.id
}
