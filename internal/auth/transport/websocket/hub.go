package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// User to clients mapping for targeted messaging
	userClients map[string][]*Client

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Event handlers
	eventHandlers map[string][]EventHandler

	// Logger
	logger *logger.Logger

	// Mutex for thread safety
	mu sync.RWMutex

	// Upgrader for WebSocket connections
	upgrader websocket.Upgrader
}

// NewHub creates a new WebSocket hub
func NewHub(logger *logger.Logger) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		userClients:   make(map[string][]*Client),
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		eventHandlers: make(map[string][]EventHandler),
		logger:        logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin in development
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
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
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// Add to user clients mapping
	if client.userID != "" {
		h.userClients[client.userID] = append(h.userClients[client.userID], client)
	}

	h.logger.InfoWithFields("Client registered",
		logger.String("client_id", client.id),
		logger.String("user_id", client.userID))

	// Send welcome message
	welcomeMsg := Message{
		Type:      "connection",
		Data:      map[string]string{"status": "connected"},
		Timestamp: time.Now(),
	}
	client.send <- h.encodeMessage(welcomeMsg)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		// Remove from user clients mapping
		if client.userID != "" {
			userClients := h.userClients[client.userID]
			for i, c := range userClients {
				if c == client {
					h.userClients[client.userID] = append(userClients[:i], userClients[i+1:]...)
					break
				}
			}
			// Clean up empty slice
			if len(h.userClients[client.userID]) == 0 {
				delete(h.userClients, client.userID)
			}
		}

		h.logger.InfoWithFields("Client unregistered",
			logger.String("client_id", client.id),
			logger.String("user_id", client.userID))
	}
}

// broadcastMessage broadcasts a message to all clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// SendToUser sends a message to all clients of a specific user
func (h *Hub) SendToUser(userID string, message Message) {
	h.mu.RLock()
	clients := h.userClients[userID]
	h.mu.RUnlock()

	if len(clients) == 0 {
		return
	}

	messageBytes := h.encodeMessage(message)
	for _, client := range clients {
		select {
		case client.send <- messageBytes:
		default:
			// Client's send channel is full, skip
			h.logger.WarnWithFields("Client send channel full",
				logger.String("client_id", client.id),
				logger.String("user_id", userID))
		}
	}
}

// SendToAll sends a message to all connected clients
func (h *Hub) SendToAll(message Message) {
	messageBytes := h.encodeMessage(message)
	h.broadcast <- messageBytes
}

// HandleWebSocket handles WebSocket upgrade and client management
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.ErrorWithFields("WebSocket upgrade failed", logger.Error(err))
		return
	}

	// Extract user ID from context (set by authentication middleware)
	userID := h.getUserIDFromContext(r.Context())

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		id:       generateClientID(),
		userID:   userID,
		channels: make(map[string]bool),
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// Subscribe subscribes to auth events
func (h *Hub) Subscribe(eventType string, handler EventHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.eventHandlers[eventType] = append(h.eventHandlers[eventType], handler)
}

// HandleAuthEvent handles authentication-related events
func (h *Hub) HandleAuthEvent(event *domain.DomainEvent) {
	h.mu.RLock()
	handlers := h.eventHandlers[event.Type]
	h.mu.RUnlock()

	// Execute event handlers
	for _, handler := range handlers {
		go func(handler EventHandler, e *domain.DomainEvent) {
			if err := handler.Handle(e); err != nil {
				h.logger.ErrorWithFields("Event handler failed",
					logger.Error(err),
					logger.String("event_type", e.Type))
			}
		}(handler, event)
	}

	// Send real-time notifications based on event type
	h.handleEventNotification(event)
}

// handleEventNotification sends real-time notifications for events
func (h *Hub) handleEventNotification(event *domain.DomainEvent) {
	switch event.Type {
	case domain.EventTypeUserLoggedIn:
		h.handleUserLoginEvent(event)
	case domain.EventTypeUserLoggedOut:
		h.handleUserLogoutEvent(event)
	case domain.EventTypeSessionCreated:
		h.handleSessionCreatedEvent(event)
	case domain.EventTypeSessionExpired:
		h.handleSessionExpiredEvent(event)
	case domain.EventTypeFailedLogin:
		h.handleFailedLoginEvent(event)
	case domain.EventTypeMFAEnabled:
		h.handleMFAEnabledEvent(event)
	case domain.EventTypeUserPasswordChanged:
		h.handlePasswordChangedEvent(event)
	}
}

// Event notification handlers

func (h *Hub) handleUserLoginEvent(event *domain.DomainEvent) {
	if userID, ok := event.Data["user_id"].(string); ok {
		message := Message{
			Type: "user_login",
			Data: map[string]interface{}{
				"message":   "User logged in successfully",
				"timestamp": event.Timestamp,
			},
			Timestamp: time.Now(),
		}
		h.SendToUser(userID, message)
	}
}

func (h *Hub) handleUserLogoutEvent(event *domain.DomainEvent) {
	if userID, ok := event.Data["user_id"].(string); ok {
		message := Message{
			Type: "user_logout",
			Data: map[string]interface{}{
				"message":   "User logged out",
				"timestamp": event.Timestamp,
			},
			Timestamp: time.Now(),
		}
		h.SendToUser(userID, message)
	}
}

func (h *Hub) handleSessionCreatedEvent(event *domain.DomainEvent) {
	if userID, ok := event.Data["user_id"].(string); ok {
		message := Message{
			Type: "session_created",
			Data: map[string]interface{}{
				"message":    "New session created",
				"session_id": event.Data["session_id"],
				"timestamp":  event.Timestamp,
			},
			Timestamp: time.Now(),
		}
		h.SendToUser(userID, message)
	}
}

func (h *Hub) handleSessionExpiredEvent(event *domain.DomainEvent) {
	if userID, ok := event.Data["user_id"].(string); ok {
		message := Message{
			Type: "session_expired",
			Data: map[string]interface{}{
				"message":    "Session expired",
				"session_id": event.Data["session_id"],
				"timestamp":  event.Timestamp,
			},
			Timestamp: time.Now(),
		}
		h.SendToUser(userID, message)
	}
}

func (h *Hub) handleFailedLoginEvent(event *domain.DomainEvent) {
	// For security events, we might want to notify admins
	message := Message{
		Type: "security_alert",
		Data: map[string]interface{}{
			"type":      "failed_login",
			"message":   "Failed login attempt detected",
			"timestamp": event.Timestamp,
			"details":   event.Data,
		},
		Timestamp: time.Now(),
	}

	// Send to admin users (implementation would depend on how you identify admins)
	h.SendToAll(message)
}

func (h *Hub) handleMFAEnabledEvent(event *domain.DomainEvent) {
	if userID, ok := event.Data["user_id"].(string); ok {
		message := Message{
			Type: "mfa_enabled",
			Data: map[string]interface{}{
				"message":   "Multi-factor authentication enabled",
				"timestamp": event.Timestamp,
			},
			Timestamp: time.Now(),
		}
		h.SendToUser(userID, message)
	}
}

func (h *Hub) handlePasswordChangedEvent(event *domain.DomainEvent) {
	if userID, ok := event.Data["user_id"].(string); ok {
		message := Message{
			Type: "password_changed",
			Data: map[string]interface{}{
				"message":   "Password changed successfully",
				"timestamp": event.Timestamp,
			},
			Timestamp: time.Now(),
		}
		h.SendToUser(userID, message)
	}
}

// Helper methods

func (h *Hub) encodeMessage(message Message) []byte {
	data, _ := json.Marshal(message)
	return data
}

func (h *Hub) getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetStats returns hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"total_clients":   len(h.clients),
		"connected_users": len(h.userClients),
		"event_handlers":  len(h.eventHandlers),
	}
}

// EventHandler defines the interface for handling WebSocket events
type EventHandler interface {
	Handle(event *domain.DomainEvent) error
}

// EventHandlerFunc is a function type that implements EventHandler
type EventHandlerFunc func(event *domain.DomainEvent) error

// Handle implements the EventHandler interface
func (f EventHandlerFunc) Handle(event *domain.DomainEvent) error {
	return f(event)
}
