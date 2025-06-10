package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"go.uber.org/zap"
)

// SimpleProducer interface for Kafka producer
type SimpleProducer interface {
	Produce(topic string, key []byte, value []byte) error
	ProduceJSON(topic string, key string, value interface{}) error
	Close() error
}

// SimpleConfig represents simple Kafka producer configuration
type SimpleConfig struct {
	Brokers     []string      `yaml:"brokers"`
	Timeout     time.Duration `yaml:"timeout"`
	Compression string        `yaml:"compression"`
	BatchSize   int           `yaml:"batch_size"`
	BatchTimeout time.Duration `yaml:"batch_timeout"`
}

// SimpleMockProducer implements SimpleProducer interface for testing
type SimpleMockProducer struct {
	messages []SimpleMockMessage
	logger   *logger.Logger
}

// SimpleMockMessage represents a mock Kafka message
type SimpleMockMessage struct {
	Topic string
	Key   string
	Value string
	Time  time.Time
}

// NewSimpleProducer creates a new simple mock producer
func NewSimpleProducer(config SimpleConfig, logger *logger.Logger) (SimpleProducer, error) {
	return &SimpleMockProducer{
		messages: make([]SimpleMockMessage, 0),
		logger:   logger,
	}, nil
}

// Produce sends a message to mock Kafka
func (m *SimpleMockProducer) Produce(topic string, key []byte, value []byte) error {
	message := SimpleMockMessage{
		Topic: topic,
		Key:   string(key),
		Value: string(value),
		Time:  time.Now(),
	}

	m.messages = append(m.messages, message)

	m.logger.Debug("Mock message produced",
		zap.String("topic", topic),
		zap.String("key", string(key)),
	)

	return nil
}

// ProduceJSON sends a JSON message to mock Kafka
func (m *SimpleMockProducer) ProduceJSON(topic string, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return m.Produce(topic, []byte(key), jsonValue)
}

// Close mock implementation
func (m *SimpleMockProducer) Close() error {
	m.logger.Debug("Mock producer closed")
	return nil
}

// GetMessages returns all mock messages (for testing)
func (m *SimpleMockProducer) GetMessages() []SimpleMockMessage {
	return m.messages
}

// Clear clears all mock messages (for testing)
func (m *SimpleMockProducer) Clear() {
	m.messages = make([]SimpleMockMessage, 0)
}
