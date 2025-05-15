package kafka

import (
	"log"

	"github.com/IBM/sarama"
	"kafka_producer/config"
)

// Producer interface defines the methods for a Kafka producer
type Producer interface {
	PushToQueue(topic string, message []byte) error
	Close() error
}

// SaramaProducer implements the Producer interface using Sarama
type SaramaProducer struct {
	producer sarama.SyncProducer
}

// NewProducer creates a new Kafka producer
func NewProducer(config *config.Config) (Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	
	// Set required acks based on config
	switch config.Kafka.RequiredAcks {
	case "none":
		saramaConfig.Producer.RequiredAcks = sarama.NoResponse
	case "local":
		saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal
	case "all":
		saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	default:
		saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	}
	
	saramaConfig.Producer.Retry.Max = config.Kafka.RetryMax

	producer, err := sarama.NewSyncProducer(config.Kafka.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &SaramaProducer{
		producer: producer,
	}, nil
}

// PushToQueue sends a message to a Kafka topic
func (p *SaramaProducer) PushToQueue(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Order is stored in topic(%s)/partition(%d)/offset(%d)\n",
		topic,
		partition,
		offset)

	return nil
}

// Close closes the Kafka producer
func (p *SaramaProducer) Close() error {
	return p.producer.Close()
}
