package kafka

import (
	"kafka_worker/config"

	"github.com/IBM/sarama"
)

// Consumer interface defines the methods for a Kafka consumer
type Consumer interface {
	ConsumePartition(topic string, partition int32, offset int64) (PartitionConsumer, error)
	Close() error
}

// PartitionConsumer interface defines the methods for a Kafka partition consumer
type PartitionConsumer interface {
	Messages() <-chan *sarama.ConsumerMessage
	Errors() <-chan error
	Close() error
}

// SaramaConsumer implements the Consumer interface using Sarama
type SaramaConsumer struct {
	consumer sarama.Consumer
}

// SaramaPartitionConsumer implements the PartitionConsumer interface using Sarama
type SaramaPartitionConsumer struct {
	consumer sarama.PartitionConsumer
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(config *config.Config) (Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(config.Kafka.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &SaramaConsumer{
		consumer: consumer,
	}, nil
}

// ConsumePartition creates a new partition consumer
func (c *SaramaConsumer) ConsumePartition(topic string, partition int32, offset int64) (PartitionConsumer, error) {
	consumer, err := c.consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return nil, err
	}

	return &SaramaPartitionConsumer{
		consumer: consumer,
	}, nil
}

// Close closes the consumer
func (c *SaramaConsumer) Close() error {
	return c.consumer.Close()
}

// Messages returns the message channel
func (pc *SaramaPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return pc.consumer.Messages()
}

// Errors returns the error channel
func (pc *SaramaPartitionConsumer) Errors() <-chan error {
	// Convert from <-chan *sarama.ConsumerError to <-chan error
	errorChan := make(chan error)
	go func() {
		for err := range pc.consumer.Errors() {
			errorChan <- err
		}
		close(errorChan)
	}()
	return errorChan
}

// Close closes the partition consumer
func (pc *SaramaPartitionConsumer) Close() error {
	return pc.consumer.Close()
}
