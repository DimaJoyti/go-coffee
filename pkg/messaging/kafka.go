package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// MessageBus interface defines messaging operations
type MessageBus interface {
	Publish(ctx context.Context, topic string, message *Message) error
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Close() error
}

// Message represents a message in the system
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Headers   map[string]string      `json:"headers,omitempty"`
}

// MessageHandler defines a function to handle messages
type MessageHandler func(ctx context.Context, message *Message) error

// KafkaMessageBus implements MessageBus using Apache Kafka
type KafkaMessageBus struct {
	brokers   []string
	groupID   string
	writers   map[string]*kafka.Writer
	readers   map[string]*kafka.Reader
	config    *KafkaConfig
}

// KafkaConfig represents Kafka configuration
type KafkaConfig struct {
	Brokers       []string
	GroupID       string
	BatchSize     int
	BatchTimeout  time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
}

// NewKafkaMessageBus creates a new Kafka message bus
func NewKafkaMessageBus(config *KafkaConfig) *KafkaMessageBus {
	return &KafkaMessageBus{
		brokers: config.Brokers,
		groupID: config.GroupID,
		writers: make(map[string]*kafka.Writer),
		readers: make(map[string]*kafka.Reader),
		config:  config,
	}
}

// Publish publishes a message to a topic
func (kmb *KafkaMessageBus) Publish(ctx context.Context, topic string, message *Message) error {
	writer := kmb.getWriter(topic)

	// Serialize message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create Kafka message
	kafkaMessage := kafka.Message{
		Key:   []byte(message.ID),
		Value: data,
		Time:  message.Timestamp,
	}

	// Add headers
	for key, value := range message.Headers {
		kafkaMessage.Headers = append(kafkaMessage.Headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	// Publish with retry
	for attempt := 0; attempt < kmb.config.RetryAttempts; attempt++ {
		err = writer.WriteMessages(ctx, kafkaMessage)
		if err == nil {
			return nil
		}

		if attempt < kmb.config.RetryAttempts-1 {
			time.Sleep(kmb.config.RetryDelay)
		}
	}

	return fmt.Errorf("failed to publish message after %d attempts: %w", kmb.config.RetryAttempts, err)
}

// Subscribe subscribes to a topic and handles messages
func (kmb *KafkaMessageBus) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	reader := kmb.getReader(topic)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Read message
				kafkaMessage, err := reader.ReadMessage(ctx)
				if err != nil {
					// Log error and continue
					continue
				}

				// Deserialize message
				var message Message
				if err := json.Unmarshal(kafkaMessage.Value, &message); err != nil {
					// Log error and continue
					continue
				}

				// Handle message
				if err := handler(ctx, &message); err != nil {
					// Log error but don't stop processing
					continue
				}
			}
		}
	}()

	return nil
}

// Close closes all Kafka connections
func (kmb *KafkaMessageBus) Close() error {
	// Close all writers
	for _, writer := range kmb.writers {
		if err := writer.Close(); err != nil {
			return err
		}
	}

	// Close all readers
	for _, reader := range kmb.readers {
		if err := reader.Close(); err != nil {
			return err
		}
	}

	return nil
}

// getWriter returns a writer for the topic
func (kmb *KafkaMessageBus) getWriter(topic string) *kafka.Writer {
	if writer, exists := kmb.writers[topic]; exists {
		return writer
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(kmb.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    kmb.config.BatchSize,
		BatchTimeout: kmb.config.BatchTimeout,
	}

	kmb.writers[topic] = writer
	return writer
}

// getReader returns a reader for the topic
func (kmb *KafkaMessageBus) getReader(topic string) *kafka.Reader {
	if reader, exists := kmb.readers[topic]; exists {
		return reader
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: kmb.brokers,
		Topic:   topic,
		GroupID: kmb.groupID,
	})

	kmb.readers[topic] = reader
	return reader
}

// Event types for the coffee shop

// OrderEvent represents order-related events
type OrderEvent struct {
	OrderID     string                 `json:"order_id"`
	UserID      string                 `json:"user_id"`
	Action      string                 `json:"action"` // created, updated, cancelled, completed
	Items       []OrderItem            `json:"items"`
	TotalAmount float64                `json:"total_amount"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// PaymentEvent represents payment-related events
type PaymentEvent struct {
	PaymentID   string                 `json:"payment_id"`
	OrderID     string                 `json:"order_id"`
	UserID      string                 `json:"user_id"`
	Action      string                 `json:"action"` // initiated, completed, failed, refunded
	Amount      float64                `json:"amount"`
	Currency    string                 `json:"currency"`
	Method      string                 `json:"method"` // bitcoin, ethereum, credit_card
	TxHash      string                 `json:"tx_hash,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// KitchenEvent represents kitchen-related events
type KitchenEvent struct {
	OrderID      string                 `json:"order_id"`
	Action       string                 `json:"action"` // received, started, completed, delayed
	EstimatedTime int                   `json:"estimated_time_minutes"`
	ActualTime   int                    `json:"actual_time_minutes,omitempty"`
	StaffID      string                 `json:"staff_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UserEvent represents user-related events
type UserEvent struct {
	UserID   string                 `json:"user_id"`
	Action   string                 `json:"action"` // registered, login, logout, updated, deleted
	Email    string                 `json:"email,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EventPublisher provides high-level event publishing
type EventPublisher struct {
	messageBus MessageBus
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(messageBus MessageBus) *EventPublisher {
	return &EventPublisher{messageBus: messageBus}
}

// PublishOrderEvent publishes an order event
func (ep *EventPublisher) PublishOrderEvent(ctx context.Context, event *OrderEvent) error {
	message := &Message{
		ID:        fmt.Sprintf("order_%s_%d", event.OrderID, time.Now().UnixNano()),
		Type:      "order_event",
		Source:    "order_service",
		Data:      map[string]interface{}{"event": event},
		Timestamp: time.Now(),
	}

	return ep.messageBus.Publish(ctx, "orders", message)
}

// PublishPaymentEvent publishes a payment event
func (ep *EventPublisher) PublishPaymentEvent(ctx context.Context, event *PaymentEvent) error {
	message := &Message{
		ID:        fmt.Sprintf("payment_%s_%d", event.PaymentID, time.Now().UnixNano()),
		Type:      "payment_event",
		Source:    "payment_service",
		Data:      map[string]interface{}{"event": event},
		Timestamp: time.Now(),
	}

	return ep.messageBus.Publish(ctx, "payments", message)
}

// PublishKitchenEvent publishes a kitchen event
func (ep *EventPublisher) PublishKitchenEvent(ctx context.Context, event *KitchenEvent) error {
	message := &Message{
		ID:        fmt.Sprintf("kitchen_%s_%d", event.OrderID, time.Now().UnixNano()),
		Type:      "kitchen_event",
		Source:    "kitchen_service",
		Data:      map[string]interface{}{"event": event},
		Timestamp: time.Now(),
	}

	return ep.messageBus.Publish(ctx, "kitchen", message)
}

// PublishUserEvent publishes a user event
func (ep *EventPublisher) PublishUserEvent(ctx context.Context, event *UserEvent) error {
	message := &Message{
		ID:        fmt.Sprintf("user_%s_%d", event.UserID, time.Now().UnixNano()),
		Type:      "user_event",
		Source:    "auth_service",
		Data:      map[string]interface{}{"event": event},
		Timestamp: time.Now(),
	}

	return ep.messageBus.Publish(ctx, "users", message)
}

// EventSubscriber provides high-level event subscription
type EventSubscriber struct {
	messageBus MessageBus
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(messageBus MessageBus) *EventSubscriber {
	return &EventSubscriber{messageBus: messageBus}
}

// SubscribeToOrders subscribes to order events
func (es *EventSubscriber) SubscribeToOrders(ctx context.Context, handler func(*OrderEvent) error) error {
	return es.messageBus.Subscribe(ctx, "orders", func(ctx context.Context, message *Message) error {
		if message.Type != "order_event" {
			return nil
		}

		eventData, ok := message.Data["event"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid order event data")
		}

		// Convert to OrderEvent (simplified)
		var event OrderEvent
		data, _ := json.Marshal(eventData)
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}

		return handler(&event)
	})
}

// SubscribeToPayments subscribes to payment events
func (es *EventSubscriber) SubscribeToPayments(ctx context.Context, handler func(*PaymentEvent) error) error {
	return es.messageBus.Subscribe(ctx, "payments", func(ctx context.Context, message *Message) error {
		if message.Type != "payment_event" {
			return nil
		}

		eventData, ok := message.Data["event"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid payment event data")
		}

		var event PaymentEvent
		data, _ := json.Marshal(eventData)
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}

		return handler(&event)
	})
}

// DefaultKafkaConfig returns default Kafka configuration
func DefaultKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers:       []string{"localhost:9092"},
		GroupID:       "go_coffee",
		BatchSize:     100,
		BatchTimeout:  time.Millisecond * 10,
		RetryAttempts: 3,
		RetryDelay:    time.Second,
	}
}
