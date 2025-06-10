package kafka

import (
	"context"
	"log"
	"sync"

	"github.com/DimaJoyti/go-coffee/consumer/config"

	"github.com/IBM/sarama"
)

// ConsumerGroupHandler interface defines the methods for a Kafka consumer group handler
type ConsumerGroupHandler interface {
	Setup(sarama.ConsumerGroupSession) error
	Cleanup(sarama.ConsumerGroupSession) error
	ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error
}

// GroupConsumer interface defines the methods for a Kafka consumer group
type GroupConsumer interface {
	Consume(ctx context.Context, topics []string, handler ConsumerGroupHandler) error
	Close() error
}

// SaramaGroupConsumer implements the GroupConsumer interface using Sarama
type SaramaGroupConsumer struct {
	consumerGroup sarama.ConsumerGroup
}

// NewGroupConsumer creates a new Kafka consumer group
func NewGroupConsumer(config *config.Config) (GroupConsumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Enable consumer group rebalance strategy
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	consumerGroup, err := sarama.NewConsumerGroup(config.Kafka.Brokers, config.Kafka.ConsumerGroup, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &SaramaGroupConsumer{
		consumerGroup: consumerGroup,
	}, nil
}

// Consume starts consuming messages from the specified topics
func (c *SaramaGroupConsumer) Consume(ctx context.Context, topics []string, handler ConsumerGroupHandler) error {
	// Track errors
	go func() {
		for err := range c.consumerGroup.Errors() {
			log.Printf("Consumer group error: %v", err)
		}
	}()

	// Consume messages
	return c.consumerGroup.Consume(ctx, topics, handler.(sarama.ConsumerGroupHandler))
}

// Close closes the consumer group
func (c *SaramaGroupConsumer) Close() error {
	return c.consumerGroup.Close()
}

// OrderConsumerHandler handles consuming order messages
type OrderConsumerHandler struct {
	ready    chan bool
	messages chan *sarama.ConsumerMessage
	errors   chan error
	wg       sync.WaitGroup
}

// NewOrderConsumerHandler creates a new OrderConsumerHandler
func NewOrderConsumerHandler() *OrderConsumerHandler {
	return &OrderConsumerHandler{
		ready:    make(chan bool),
		messages: make(chan *sarama.ConsumerMessage),
		errors:   make(chan error),
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *OrderConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(h.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *OrderConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *OrderConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE: Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		h.messages <- message
		session.MarkMessage(message, "")
	}

	return nil
}

// Messages returns the message channel
func (h *OrderConsumerHandler) Messages() <-chan *sarama.ConsumerMessage {
	return h.messages
}

// Errors returns the error channel
func (h *OrderConsumerHandler) Errors() <-chan error {
	return h.errors
}

// WaitReady waits until the consumer is ready
func (h *OrderConsumerHandler) WaitReady() {
	<-h.ready
}

// Close closes the handler
func (h *OrderConsumerHandler) Close() {
	close(h.messages)
	close(h.errors)
}
