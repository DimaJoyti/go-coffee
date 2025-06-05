package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/segmentio/kafka-go"
)

// Producer interface for Kafka producer
type Producer interface {
	Produce(topic string, key []byte, value []byte) error
	ProduceJSON(topic string, key string, value interface{}) error
	Close() error
}

// Config represents Kafka producer configuration
type Config struct {
	Brokers     []string      `yaml:"brokers"`
	Timeout     time.Duration `yaml:"timeout"`
	Compression string        `yaml:"compression"`
	BatchSize   int           `yaml:"batch_size"`
	BatchTimeout time.Duration `yaml:"batch_timeout"`
}

// KafkaProducer implements Producer interface
type KafkaProducer struct {
	config Config
	logger *logger.Logger
	writer *kafka.Writer
}

// MockProducer implements Producer interface for testing
type MockProducer struct {
	messages []MockMessage
	logger   *logger.Logger
}

// MockMessage represents a mock Kafka message
type MockMessage struct {
	Topic string
	Key   string
	Value string
	Time  time.Time
}

// NewProducer creates a new Kafka producer
func NewProducer(config Config, logger *logger.Logger) (Producer, error) {
	if len(config.Brokers) == 0 {
		config.Brokers = []string{"localhost:9092"}
	}

	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	if config.BatchTimeout == 0 {
		config.BatchTimeout = 1 * time.Second
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: config.BatchTimeout,
		WriteTimeout: config.Timeout,
	}

	if config.BatchSize > 0 {
		writer.BatchSize = config.BatchSize
	}

	// Set compression
	switch config.Compression {
	case "gzip":
		writer.Compression = kafka.Gzip
	case "snappy":
		writer.Compression = kafka.Snappy
	case "lz4":
		writer.Compression = kafka.Lz4
	case "zstd":
		writer.Compression = kafka.Zstd
	default:
		writer.Compression = kafka.Snappy // Default compression
	}

	return &KafkaProducer{
		config: config,
		logger: logger,
		writer: writer,
	}, nil
}

// Produce sends a message to Kafka
func (p *KafkaProducer) Produce(topic string, key []byte, value []byte) error {
	message := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.config.Timeout)
	defer cancel()

	err := p.writer.WriteMessages(ctx, message)
	if err != nil {
		p.logger.Error("Failed to produce message to Kafka",
			zap.Error(err),
			zap.String("topic", topic))
		return fmt.Errorf("failed to produce message: %w", err)
	}

	p.logger.Debug("Message produced to Kafka",
		zap.String("topic", topic),
		zap.String("key", string(key)),
		zap.Int("value_size", len(value)))

	return nil
}

// ProduceJSON sends a JSON message to Kafka
func (p *KafkaProducer) ProduceJSON(topic string, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return p.Produce(topic, []byte(key), jsonValue)
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// NewMockProducer creates a new mock producer
func NewMockProducer(logger *logger.Logger) Producer {
	return &MockProducer{
		messages: make([]MockMessage, 0),
		logger:   logger,
	}
}

// Produce mock implementation
func (m *MockProducer) Produce(topic string, key []byte, value []byte) error {
	message := MockMessage{
		Topic: topic,
		Key:   string(key),
		Value: string(value),
		Time:  time.Now(),
	}

	m.messages = append(m.messages, message)

	m.logger.Debug("Mock message produced",
		zap.String("topic", topic),
		zap.String("key", string(key)))

	return nil
}

// ProduceJSON mock implementation
func (m *MockProducer) ProduceJSON(topic string, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return m.Produce(topic, []byte(key), jsonValue)
}

// Close mock implementation
func (m *MockProducer) Close() error {
	m.logger.Debug("Mock producer closed")
	return nil
}

// GetMessages returns all mock messages (for testing)
func (m *MockProducer) GetMessages() []MockMessage {
	return m.messages
}

// Clear clears all mock messages (for testing)
func (m *MockProducer) Clear() {
	m.messages = make([]MockMessage, 0)
}
