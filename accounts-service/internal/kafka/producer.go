package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/config"
)

// EventType represents the type of event
type EventType string

const (
	// Account events
	EventTypeAccountCreated EventType = "account.created"
	EventTypeAccountUpdated EventType = "account.updated"
	EventTypeAccountDeleted EventType = "account.deleted"

	// Vendor events
	EventTypeVendorCreated EventType = "vendor.created"
	EventTypeVendorUpdated EventType = "vendor.updated"
	EventTypeVendorDeleted EventType = "vendor.deleted"

	// Product events
	EventTypeProductCreated EventType = "product.created"
	EventTypeProductUpdated EventType = "product.updated"
	EventTypeProductDeleted EventType = "product.deleted"

	// Order events
	EventTypeOrderCreated     EventType = "order.created"
	EventTypeOrderStatusChanged EventType = "order.status_changed"
	EventTypeOrderDeleted     EventType = "order.deleted"
)

// Event represents a Kafka event
type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// Producer represents a Kafka producer
type Producer interface {
	// Publish publishes an event to Kafka
	Publish(eventType EventType, payload interface{}) error

	// Close closes the producer
	Close() error
}

// KafkaProducer implements the Producer interface using Kafka
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(cfg *config.Config) (Producer, error) {
	// Create Kafka producer configuration
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.RequiredAcks(getRequiredAcks(cfg.Kafka.RequiredAcks))
	kafkaConfig.Producer.Retry.Max = cfg.Kafka.RetryMax
	kafkaConfig.Producer.Return.Successes = true

	// Create Kafka producer
	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer: producer,
		topic:    cfg.Kafka.Topic,
	}, nil
}

// Publish publishes an event to Kafka
func (p *KafkaProducer) Publish(eventType EventType, payload interface{}) error {
	// Create event
	event := Event{
		ID:        generateID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	// Marshal event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(eventJSON),
		Key:   sarama.StringEncoder(string(eventType)),
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Event published to topic %s, partition %d, offset %d", p.topic, partition, offset)
	return nil
}

// Close closes the producer
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

// getRequiredAcks converts a string to sarama.RequiredAcks
func getRequiredAcks(acks string) sarama.RequiredAcks {
	switch acks {
	case "no":
		return sarama.NoResponse
	case "local":
		return sarama.WaitForLocal
	case "all":
		return sarama.WaitForAll
	default:
		return sarama.WaitForAll
	}
}

// generateID generates a unique ID for an event
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
