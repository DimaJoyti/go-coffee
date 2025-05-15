package kafka

import (
	"github.com/IBM/sarama"
)

// Consumer інтерфейс визначає методи для Kafka consumer
type Consumer interface {
	// Consume починає споживання повідомлень з теми
	Consume(topic string, handler MessageHandler) error
	// Close закриває з'єднання з Kafka
	Close() error
}

// MessageHandler функція обробки повідомлень
type MessageHandler func(message []byte) error

// ConsumerConfig містить конфігурацію для Kafka consumer
type ConsumerConfig struct {
	// Brokers список адрес Kafka brokers
	Brokers []string
	// Topic тема Kafka за замовчуванням
	Topic string
	// ConsumerGroup група споживачів
	ConsumerGroup string
	// AutoOffsetReset стратегія скидання зміщення (earliest, latest)
	AutoOffsetReset string
}

// NewConsumer створює новий Kafka consumer
func NewConsumer(config *ConsumerConfig) (Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	
	// Встановлення auto offset reset на основі конфігурації
	switch config.AutoOffsetReset {
	case "earliest":
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "latest":
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	consumer, err := sarama.NewConsumer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &saramaConsumer{
		consumer: consumer,
		config:   config,
	}, nil
}

// saramaConsumer реалізує інтерфейс Consumer використовуючи Sarama
type saramaConsumer struct {
	consumer sarama.Consumer
	config   *ConsumerConfig
}

// Consume починає споживання повідомлень з теми
func (c *saramaConsumer) Consume(topic string, handler MessageHandler) error {
	partitions, err := c.consumer.Partitions(topic)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			for message := range pc.Messages() {
				if err := handler(message.Value); err != nil {
					// Тут можна додати логування помилок
				}
			}
		}(pc)
	}

	return nil
}

// Close закриває з'єднання з Kafka
func (c *saramaConsumer) Close() error {
	return c.consumer.Close()
}
