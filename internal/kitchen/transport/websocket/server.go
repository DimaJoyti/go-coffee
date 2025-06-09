package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Server represents the WebSocket server for real-time kitchen updates
type Server struct {
	kitchenService      application.KitchenService
	queueService        application.QueueService
	eventService        application.EventService
	logger              *logger.Logger
	upgrader            websocket.Upgrader
	clients             map[*Client]bool
	clientsMutex        sync.RWMutex
	broadcast           chan []byte
	register            chan *Client
	unregister          chan *Client
	eventSubscriptions  map[string][]string // client_id -> event_types
	subscriptionsMutex  sync.RWMutex
}

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	conn     *websocket.Conn
	send     chan []byte
	server   *Server
	userType string // "kitchen_staff", "manager", "customer"
	userID   string
}

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	Event     string                 `json:"event,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	ClientID  string                 `json:"client_id,omitempty"`
}

// NewServer creates a new WebSocket server
func NewServer(
	kitchenService application.KitchenService,
	queueService application.QueueService,
	eventService application.EventService,
	logger *logger.Logger,
) *Server {
	return &Server{
		kitchenService: kitchenService,
		queueService:   queueService,
		eventService:   eventService,
		logger:         logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:            make(map[*Client]bool),
		broadcast:          make(chan []byte),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		eventSubscriptions: make(map[string][]string),
	}
}

// Start starts the WebSocket server
func (s *Server) Start() {
	go s.handleConnections()
	s.logger.Info("WebSocket server started")
}

// Stop stops the WebSocket server
func (s *Server) Stop() {
	close(s.broadcast)
	close(s.register)
	close(s.unregister)
	s.logger.Info("WebSocket server stopped")
}

// HandleWebSocket handles WebSocket connections
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}

	// Extract client information from query parameters or headers
	clientID := r.URL.Query().Get("client_id")
	userType := r.URL.Query().Get("user_type")
	userID := r.URL.Query().Get("user_id")

	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
	}

	client := &Client{
		ID:       clientID,
		conn:     conn,
		send:     make(chan []byte, 256),
		server:   s,
		userType: userType,
		userID:   userID,
	}

	s.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()

	s.logger.WithFields(map[string]interface{}{
		"client_id": clientID,
		"user_type": userType,
		"user_id":   userID,
	}).Info("New WebSocket client connected")
}

// handleConnections manages client connections and broadcasts
func (s *Server) handleConnections() {
	for {
		select {
		case client := <-s.register:
			s.clientsMutex.Lock()
			s.clients[client] = true
			s.clientsMutex.Unlock()

			// Send welcome message
			welcomeMsg := Message{
				Type:      "connection",
				Event:     "connected",
				Data:      map[string]interface{}{"client_id": client.ID},
				Timestamp: time.Now(),
			}
			client.sendMessage(welcomeMsg)

			// Send current queue status
			s.sendQueueStatusToClient(client)

		case client := <-s.unregister:
			s.clientsMutex.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.clientsMutex.Unlock()

			// Remove event subscriptions
			s.subscriptionsMutex.Lock()
			delete(s.eventSubscriptions, client.ID)
			s.subscriptionsMutex.Unlock()

			s.logger.WithField("client_id", client.ID).Info("WebSocket client disconnected")

		case message := <-s.broadcast:
			s.clientsMutex.RLock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.clientsMutex.RUnlock()
		}
	}
}

// BroadcastMessage broadcasts a message to all connected clients
func (s *Server) BroadcastMessage(message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal broadcast message")
		return
	}

	s.broadcast <- data
}

// BroadcastToUserType broadcasts a message to clients of a specific user type
func (s *Server) BroadcastToUserType(message Message, userType string) {
	data, err := json.Marshal(message)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal message")
		return
	}

	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	for client := range s.clients {
		if client.userType == userType {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(s.clients, client)
			}
		}
	}
}

// SendToClient sends a message to a specific client
func (s *Server) SendToClient(clientID string, message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal message")
		return
	}

	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	for client := range s.clients {
		if client.ID == clientID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(s.clients, client)
			}
			break
		}
	}
}

// Event Handlers

// HandleOrderEvent handles order-related events
func (s *Server) HandleOrderEvent(event *domain.DomainEvent) {
	message := Message{
		Type:      "event",
		Event:     event.Type,
		Data:      event.Data,
		Timestamp: event.OccurredAt,
	}

	switch event.Type {
	case domain.OrderAddedToQueueEvent:
		s.BroadcastToUserType(message, "kitchen_staff")
		s.BroadcastToUserType(message, "manager")
	case domain.OrderStatusChangedEvent:
		s.BroadcastMessage(message)
	case domain.OrderCompletedEvent:
		s.BroadcastMessage(message)
	case domain.OrderOverdueEvent:
		s.BroadcastToUserType(message, "kitchen_staff")
		s.BroadcastToUserType(message, "manager")
	}

	s.logger.WithFields(map[string]interface{}{
		"event_type":   event.Type,
		"aggregate_id": event.AggregateID,
	}).Info("Broadcasted order event")
}

// HandleEquipmentEvent handles equipment-related events
func (s *Server) HandleEquipmentEvent(event *domain.DomainEvent) {
	message := Message{
		Type:      "event",
		Event:     event.Type,
		Data:      event.Data,
		Timestamp: event.OccurredAt,
	}

	// Equipment events are mainly for kitchen staff and managers
	s.BroadcastToUserType(message, "kitchen_staff")
	s.BroadcastToUserType(message, "manager")

	s.logger.WithFields(map[string]interface{}{
		"event_type":   event.Type,
		"aggregate_id": event.AggregateID,
	}).Info("Broadcasted equipment event")
}

// HandleStaffEvent handles staff-related events
func (s *Server) HandleStaffEvent(event *domain.DomainEvent) {
	message := Message{
		Type:      "event",
		Event:     event.Type,
		Data:      event.Data,
		Timestamp: event.OccurredAt,
	}

	// Staff events are mainly for managers
	s.BroadcastToUserType(message, "manager")

	s.logger.WithFields(map[string]interface{}{
		"event_type":   event.Type,
		"aggregate_id": event.AggregateID,
	}).Info("Broadcasted staff event")
}

// HandleQueueEvent handles queue-related events
func (s *Server) HandleQueueEvent(event *domain.DomainEvent) {
	message := Message{
		Type:      "event",
		Event:     event.Type,
		Data:      event.Data,
		Timestamp: event.OccurredAt,
	}

	// Queue events are for all users
	s.BroadcastMessage(message)

	s.logger.WithFields(map[string]interface{}{
		"event_type":   event.Type,
		"aggregate_id": event.AggregateID,
	}).Info("Broadcasted queue event")
}

// Helper methods

func (s *Server) sendQueueStatusToClient(client *Client) {
	ctx := context.Background()
	status, err := s.kitchenService.GetQueueStatus(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get queue status for new client")
		return
	}

	message := Message{
		Type:      "queue_status",
		Data:      map[string]interface{}{"status": status},
		Timestamp: time.Now(),
	}

	client.sendMessage(message)
}

// Client methods

func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.server.logger.WithError(err).Error("WebSocket error")
			}
			break
		}

		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			c.server.logger.WithError(err).Error("Failed to unmarshal WebSocket message")
			continue
		}

		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendMessage(message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		c.server.logger.WithError(err).Error("Failed to marshal message")
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
	}
}

func (c *Client) handleMessage(message Message) {
	c.server.logger.WithFields(map[string]interface{}{
		"client_id":    c.ID,
		"message_type": message.Type,
	}).Info("Received WebSocket message")

	switch message.Type {
	case "subscribe":
		c.handleSubscription(message)
	case "unsubscribe":
		c.handleUnsubscription(message)
	case "ping":
		c.handlePing()
	case "get_queue_status":
		c.handleGetQueueStatus()
	case "get_next_order":
		c.handleGetNextOrder()
	default:
		c.server.logger.WithField("message_type", message.Type).Warn("Unknown message type")
	}
}

func (c *Client) handleSubscription(message Message) {
	eventTypes, ok := message.Data["event_types"].([]interface{})
	if !ok {
		c.server.logger.Error("Invalid event_types in subscription message")
		return
	}

	c.server.subscriptionsMutex.Lock()
	defer c.server.subscriptionsMutex.Unlock()

	subscriptions := make([]string, len(eventTypes))
	for i, eventType := range eventTypes {
		if str, ok := eventType.(string); ok {
			subscriptions[i] = str
		}
	}

	c.server.eventSubscriptions[c.ID] = subscriptions

	response := Message{
		Type:      "subscription_confirmed",
		Data:      map[string]interface{}{"event_types": subscriptions},
		Timestamp: time.Now(),
	}
	c.sendMessage(response)
}

func (c *Client) handleUnsubscription(message Message) {
	c.server.subscriptionsMutex.Lock()
	defer c.server.subscriptionsMutex.Unlock()

	delete(c.server.eventSubscriptions, c.ID)

	response := Message{
		Type:      "unsubscription_confirmed",
		Timestamp: time.Now(),
	}
	c.sendMessage(response)
}

func (c *Client) handlePing() {
	response := Message{
		Type:      "pong",
		Timestamp: time.Now(),
	}
	c.sendMessage(response)
}

func (c *Client) handleGetQueueStatus() {
	c.server.sendQueueStatusToClient(c)
}

func (c *Client) handleGetNextOrder() {
	ctx := context.Background()
	order, err := c.server.kitchenService.GetNextOrder(ctx)
	if err != nil {
		c.server.logger.WithError(err).Error("Failed to get next order")
		return
	}

	response := Message{
		Type:      "next_order",
		Data:      map[string]interface{}{"order": order},
		Timestamp: time.Now(),
	}
	c.sendMessage(response)
}
