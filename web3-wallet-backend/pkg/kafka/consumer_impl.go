package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// kafkaConsumer implements the Consumer interface
type kafkaConsumer struct {
	consumer *kafka.Consumer
	config   *Config
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(config *Config) (Consumer, error) {
	// Create Kafka configuration
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":        config.Brokers[0],
		"group.id":                 config.ConsumerGroup,
		"socket.keepalive.enable":  true,
		"enable.auto.commit":       config.EnableAutoCommit,
		"auto.offset.reset":        config.AutoOffsetReset,
		"session.timeout.ms":       int(config.SessionTimeout.Milliseconds()),
		"heartbeat.interval.ms":    int(config.HeartbeatInterval.Milliseconds()),
		"max.poll.interval.ms":     int(config.MaxPollInterval.Milliseconds()),
		"fetch.min.bytes":          config.FetchMinBytes,
		"fetch.max.bytes":          config.FetchMaxBytes,
		"fetch.wait.max.ms":        int(config.FetchMaxWait.Milliseconds()),
	}

	if config.EnableAutoCommit {
		kafkaConfig.SetKey("auto.commit.interval.ms", int(config.AutoCommitInterval.Milliseconds()))
	}

	if config.MaxPollRecords > 0 {
		kafkaConfig.SetKey("max.poll.records", config.MaxPollRecords)
	}

	// Create consumer
	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &kafkaConsumer{
		consumer: consumer,
		config:   config,
	}, nil
}

// Subscribe subscribes to topics
func (c *kafkaConsumer) Subscribe(topics []string) error {
	return c.consumer.SubscribeTopics(topics, nil)
}

// Poll polls for messages
func (c *kafkaConsumer) Poll(timeout time.Duration) (Message, error) {
	event := c.consumer.Poll(int(timeout.Milliseconds()))
	if event == nil {
		return nil, nil
	}

	switch e := event.(type) {
	case *kafka.Message:
		return &kafkaMessage{
			message: e,
		}, nil
	case kafka.Error:
		return nil, e
	default:
		return nil, nil
	}
}

// Commit commits offsets
func (c *kafkaConsumer) Commit() error {
	_, err := c.consumer.Commit()
	return err
}

// CommitMessage commits a message
func (c *kafkaConsumer) CommitMessage(msg Message) error {
	kafkaMsg, ok := msg.(*kafkaMessage)
	if !ok {
		return fmt.Errorf("invalid message type")
	}
	_, err := c.consumer.CommitMessage(kafkaMsg.message)
	return err
}

// Close closes the consumer
func (c *kafkaConsumer) Close() error {
	return c.consumer.Close()
}

// kafkaMessage implements the Message interface
type kafkaMessage struct {
	message *kafka.Message
}

// Topic returns the topic
func (m *kafkaMessage) Topic() string {
	return *m.message.TopicPartition.Topic
}

// Partition returns the partition
func (m *kafkaMessage) Partition() int32 {
	return m.message.TopicPartition.Partition
}

// Offset returns the offset
func (m *kafkaMessage) Offset() int64 {
	return int64(m.message.TopicPartition.Offset)
}

// Key returns the key
func (m *kafkaMessage) Key() []byte {
	return m.message.Key
}

// Value returns the value
func (m *kafkaMessage) Value() []byte {
	return m.message.Value
}

// Timestamp returns the timestamp
func (m *kafkaMessage) Timestamp() time.Time {
	return m.message.Timestamp
}

// Headers returns the headers
func (m *kafkaMessage) Headers() []Header {
	headers := make([]Header, 0, len(m.message.Headers))
	for _, h := range m.message.Headers {
		headers = append(headers, &kafkaHeader{
			header: h,
		})
	}
	return headers
}

// kafkaHeader implements the Header interface
type kafkaHeader struct {
	header kafka.Header
}

// Key returns the key
func (h *kafkaHeader) Key() string {
	return h.header.Key
}

// Value returns the value
func (h *kafkaHeader) Value() []byte {
	return h.header.Value
}
