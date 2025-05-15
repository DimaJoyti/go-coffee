package kafka

import (
	"github.com/IBM/sarama"
)

// Producer інтерфейс визначає методи для Kafka producer
type Producer interface {
	// PushToQueue надсилає повідомлення до теми Kafka
	PushToQueue(topic string, message []byte) error
	// Close закриває з'єднання з Kafka
	Close() error
}

// ProducerConfig містить конфігурацію для Kafka producer
type ProducerConfig struct {
	// Brokers список адрес Kafka brokers
	Brokers []string
	// Topic тема Kafka за замовчуванням
	Topic string
	// RetryMax максимальна кількість спроб
	RetryMax int
	// RequiredAcks рівень підтвердження (none, local, all)
	RequiredAcks string
}

// NewProducer створює новий Kafka producer
func NewProducer(config *ProducerConfig) (Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	
	// Встановлення required acks на основі конфігурації
	switch config.RequiredAcks {
	case "none":
		saramaConfig.Producer.RequiredAcks = sarama.NoResponse
	case "local":
		saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal
	case "all":
		saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	default:
		saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	}
	
	saramaConfig.Producer.Retry.Max = config.RetryMax

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &saramaProducer{
		producer: producer,
		config:   config,
	}, nil
}

// saramaProducer реалізує інтерфейс Producer використовуючи Sarama
type saramaProducer struct {
	producer sarama.SyncProducer
	config   *ProducerConfig
}

// PushToQueue надсилає повідомлення до теми Kafka
func (p *saramaProducer) PushToQueue(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	_, _, err := p.producer.SendMessage(msg)
	return err
}

// Close закриває з'єднання з Kafka
func (p *saramaProducer) Close() error {
	return p.producer.Close()
}
