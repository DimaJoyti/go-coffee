package domain

import (
	"encoding/json"
	"errors"
	"time"
)

// MessageType represents different types of messages
type MessageType string

const (
	// Service-to-service messages
	MessageTypeOrderCreated     MessageType = "order.created"
	MessageTypeOrderStatusChanged MessageType = "order.status_changed"
	MessageTypePaymentCompleted MessageType = "payment.completed"
	MessageTypeKitchenUpdate    MessageType = "kitchen.update"
	
	// Real-time notifications
	MessageTypeNotification     MessageType = "notification"
	MessageTypeAlert           MessageType = "alert"
	MessageTypeBroadcast       MessageType = "broadcast"
	
	// System messages
	MessageTypeHealthCheck     MessageType = "health.check"
	MessageTypeServiceDiscovery MessageType = "service.discovery"
	MessageTypeMetrics         MessageType = "metrics"
)

// MessagePriority represents message priority levels
type MessagePriority int32

const (
	MessagePriorityLow    MessagePriority = 0
	MessagePriorityNormal MessagePriority = 1
	MessagePriorityHigh   MessagePriority = 2
	MessagePriorityCritical MessagePriority = 3
)

// DeliveryMode represents message delivery modes
type DeliveryMode int32

const (
	DeliveryModeFireAndForget DeliveryMode = 0 // No acknowledgment required
	DeliveryModeAtLeastOnce   DeliveryMode = 1 // Requires acknowledgment
	DeliveryModeExactlyOnce   DeliveryMode = 2 // Requires deduplication
)

// Message represents a communication message
type Message struct {
	ID           string                 `json:"id"`
	Type         MessageType            `json:"type"`
	Source       string                 `json:"source"`       // Source service
	Target       string                 `json:"target"`       // Target service or topic
	Priority     MessagePriority        `json:"priority"`
	DeliveryMode DeliveryMode           `json:"delivery_mode"`
	Payload      map[string]interface{} `json:"payload"`
	Headers      map[string]string      `json:"headers"`
	Metadata     map[string]string      `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	RetryCount   int32                  `json:"retry_count"`
	MaxRetries   int32                  `json:"max_retries"`
	CorrelationID string                `json:"correlation_id,omitempty"`
	ReplyTo      string                 `json:"reply_to,omitempty"`
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, source, target string, payload map[string]interface{}) (*Message, error) {
	if msgType == "" {
		return nil, errors.New("message type is required")
	}
	
	if source == "" {
		return nil, errors.New("source is required")
	}
	
	if target == "" {
		return nil, errors.New("target is required")
	}

	return &Message{
		ID:           generateMessageID(),
		Type:         msgType,
		Source:       source,
		Target:       target,
		Priority:     MessagePriorityNormal,
		DeliveryMode: DeliveryModeAtLeastOnce,
		Payload:      payload,
		Headers:      make(map[string]string),
		Metadata:     make(map[string]string),
		CreatedAt:    time.Now(),
		RetryCount:   0,
		MaxRetries:   3,
	}, nil
}

// SetPriority sets the message priority
func (m *Message) SetPriority(priority MessagePriority) {
	m.Priority = priority
}

// SetDeliveryMode sets the delivery mode
func (m *Message) SetDeliveryMode(mode DeliveryMode) {
	m.DeliveryMode = mode
}

// SetExpiration sets the message expiration time
func (m *Message) SetExpiration(duration time.Duration) {
	expiresAt := m.CreatedAt.Add(duration)
	m.ExpiresAt = &expiresAt
}

// SetCorrelationID sets the correlation ID for request-response patterns
func (m *Message) SetCorrelationID(correlationID string) {
	m.CorrelationID = correlationID
}

// SetReplyTo sets the reply-to address for request-response patterns
func (m *Message) SetReplyTo(replyTo string) {
	m.ReplyTo = replyTo
}

// AddHeader adds a header to the message
func (m *Message) AddHeader(key, value string) {
	m.Headers[key] = value
}

// AddMetadata adds metadata to the message
func (m *Message) AddMetadata(key, value string) {
	m.Metadata[key] = value
}

// IsExpired checks if the message has expired
func (m *Message) IsExpired() bool {
	if m.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*m.ExpiresAt)
}

// CanRetry checks if the message can be retried
func (m *Message) CanRetry() bool {
	return m.RetryCount < m.MaxRetries
}

// IncrementRetry increments the retry count
func (m *Message) IncrementRetry() {
	m.RetryCount++
}

// ToJSON converts the message to JSON
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON creates a message from JSON
func FromJSON(data []byte) (*Message, error) {
	var message Message
	err := json.Unmarshal(data, &message)
	return &message, err
}

// GetPayloadString gets a string value from the payload
func (m *Message) GetPayloadString(key string) (string, bool) {
	if value, exists := m.Payload[key]; exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetPayloadInt gets an int value from the payload
func (m *Message) GetPayloadInt(key string) (int64, bool) {
	if value, exists := m.Payload[key]; exists {
		switch v := value.(type) {
		case int64:
			return v, true
		case int:
			return int64(v), true
		case float64:
			return int64(v), true
		}
	}
	return 0, false
}

// GetPayloadBool gets a boolean value from the payload
func (m *Message) GetPayloadBool(key string) (bool, bool) {
	if value, exists := m.Payload[key]; exists {
		if b, ok := value.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// Clone creates a copy of the message
func (m *Message) Clone() *Message {
	// Deep copy payload
	payloadCopy := make(map[string]interface{})
	for k, v := range m.Payload {
		payloadCopy[k] = v
	}

	// Deep copy headers
	headersCopy := make(map[string]string)
	for k, v := range m.Headers {
		headersCopy[k] = v
	}

	// Deep copy metadata
	metadataCopy := make(map[string]string)
	for k, v := range m.Metadata {
		metadataCopy[k] = v
	}

	clone := &Message{
		ID:            m.ID,
		Type:          m.Type,
		Source:        m.Source,
		Target:        m.Target,
		Priority:      m.Priority,
		DeliveryMode:  m.DeliveryMode,
		Payload:       payloadCopy,
		Headers:       headersCopy,
		Metadata:      metadataCopy,
		CreatedAt:     m.CreatedAt,
		RetryCount:    m.RetryCount,
		MaxRetries:    m.MaxRetries,
		CorrelationID: m.CorrelationID,
		ReplyTo:       m.ReplyTo,
	}

	if m.ExpiresAt != nil {
		expiresAt := *m.ExpiresAt
		clone.ExpiresAt = &expiresAt
	}

	return clone
}

// MessageBuilder provides a fluent interface for building messages
type MessageBuilder struct {
	message *Message
}

// NewMessageBuilder creates a new message builder
func NewMessageBuilder(msgType MessageType, source, target string) *MessageBuilder {
	message, _ := NewMessage(msgType, source, target, make(map[string]interface{}))
	return &MessageBuilder{message: message}
}

// WithPayload sets the payload
func (b *MessageBuilder) WithPayload(payload map[string]interface{}) *MessageBuilder {
	b.message.Payload = payload
	return b
}

// WithPriority sets the priority
func (b *MessageBuilder) WithPriority(priority MessagePriority) *MessageBuilder {
	b.message.SetPriority(priority)
	return b
}

// WithDeliveryMode sets the delivery mode
func (b *MessageBuilder) WithDeliveryMode(mode DeliveryMode) *MessageBuilder {
	b.message.SetDeliveryMode(mode)
	return b
}

// WithExpiration sets the expiration
func (b *MessageBuilder) WithExpiration(duration time.Duration) *MessageBuilder {
	b.message.SetExpiration(duration)
	return b
}

// WithCorrelationID sets the correlation ID
func (b *MessageBuilder) WithCorrelationID(correlationID string) *MessageBuilder {
	b.message.SetCorrelationID(correlationID)
	return b
}

// WithReplyTo sets the reply-to address
func (b *MessageBuilder) WithReplyTo(replyTo string) *MessageBuilder {
	b.message.SetReplyTo(replyTo)
	return b
}

// WithHeader adds a header
func (b *MessageBuilder) WithHeader(key, value string) *MessageBuilder {
	b.message.AddHeader(key, value)
	return b
}

// WithMetadata adds metadata
func (b *MessageBuilder) WithMetadata(key, value string) *MessageBuilder {
	b.message.AddMetadata(key, value)
	return b
}

// Build returns the built message
func (b *MessageBuilder) Build() *Message {
	return b.message
}

// Helper functions

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return "msg_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// Message Factory Functions

// NewOrderCreatedMessage creates an order created message
func NewOrderCreatedMessage(source string, orderID, customerID string, totalAmount int64) *Message {
	payload := map[string]interface{}{
		"order_id":     orderID,
		"customer_id":  customerID,
		"total_amount": totalAmount,
		"timestamp":    time.Now(),
	}

	message, _ := NewMessage(MessageTypeOrderCreated, source, "kitchen-service", payload)
	message.SetPriority(MessagePriorityHigh)
	message.AddHeader("event_type", "order_created")
	message.AddMetadata("service", source)

	return message
}

// NewPaymentCompletedMessage creates a payment completed message
func NewPaymentCompletedMessage(source string, paymentID, orderID string, amount int64) *Message {
	payload := map[string]interface{}{
		"payment_id": paymentID,
		"order_id":   orderID,
		"amount":     amount,
		"timestamp":  time.Now(),
	}

	message, _ := NewMessage(MessageTypePaymentCompleted, source, "order-service", payload)
	message.SetPriority(MessagePriorityHigh)
	message.AddHeader("event_type", "payment_completed")
	message.AddMetadata("service", source)

	return message
}

// NewNotificationMessage creates a notification message
func NewNotificationMessage(source, target string, title, body string, userID string) *Message {
	payload := map[string]interface{}{
		"title":     title,
		"body":      body,
		"user_id":   userID,
		"timestamp": time.Now(),
	}

	message, _ := NewMessage(MessageTypeNotification, source, target, payload)
	message.SetPriority(MessagePriorityNormal)
	message.SetExpiration(24 * time.Hour) // Notifications expire after 24 hours
	message.AddHeader("notification_type", "user")
	message.AddMetadata("service", source)

	return message
}
