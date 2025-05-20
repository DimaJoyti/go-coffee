package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// kafkaProducer implements the Producer interface
type kafkaProducer struct {
	producer *kafka.Producer
	config   *Config
}

// NewProducer creates a new Kafka producer
func NewProducer(config *Config) (Producer, error) {
	// Create Kafka configuration
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":        config.Brokers[0],
		"go.delivery.reports":      true,
		"go.events.channel.enable": true,
		"socket.keepalive.enable":  true,
	}

	// Set required acks
	switch config.RequiredAcks {
	case "none":
		kafkaConfig.SetKey("request.required.acks", 0)
	case "local":
		kafkaConfig.SetKey("request.required.acks", 1)
	case "all":
		kafkaConfig.SetKey("request.required.acks", -1)
	default:
		kafkaConfig.SetKey("request.required.acks", -1)
	}

	// Set compression
	if config.Compression != "" && config.Compression != "none" {
		kafkaConfig.SetKey("compression.type", config.Compression)
	}

	// Set batch size
	if config.BatchSize > 0 {
		kafkaConfig.SetKey("batch.size", config.BatchSize)
	}

	// Set batch timeout
	if config.BatchTimeout > 0 {
		kafkaConfig.SetKey("linger.ms", int(config.BatchTimeout.Milliseconds()))
	}

	// Set retry
	if config.RetryMax > 0 {
		kafkaConfig.SetKey("retries", config.RetryMax)
	}

	if config.RetryBackoff > 0 {
		kafkaConfig.SetKey("retry.backoff.ms", int(config.RetryBackoff.Milliseconds()))
	}

	// Create producer
	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Start event handler
	go func() {
		for event := range producer.Events() {
			switch ev := event.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition.Error)
				}
			}
		}
	}()

	return &kafkaProducer{
		producer: producer,
		config:   config,
	}, nil
}

// Produce produces a message to Kafka
func (p *kafkaProducer) Produce(topic string, key []byte, value []byte) error {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   key,
		Value: value,
	}

	return p.producer.Produce(message, nil)
}

// ProduceAsync produces a message to Kafka asynchronously
func (p *kafkaProducer) ProduceAsync(topic string, key []byte, value []byte, callback func(error)) {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   key,
		Value: value,
	}

	p.producer.Produce(message, func(delivery *kafka.Message) {
		if delivery.TopicPartition.Error != nil {
			callback(delivery.TopicPartition.Error)
		} else {
			callback(nil)
		}
	})
}

// Flush flushes the producer
func (p *kafkaProducer) Flush(timeout time.Duration) error {
	remaining := p.producer.Flush(int(timeout.Milliseconds()))
	if remaining > 0 {
		return fmt.Errorf("failed to flush all messages, %d remaining", remaining)
	}
	return nil
}

// Close closes the producer
func (p *kafkaProducer) Close() error {
	p.producer.Close()
	return nil
}
