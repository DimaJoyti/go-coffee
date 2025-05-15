package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/config"
)

// Consumer represents a Kafka consumer
type Consumer interface {
	// Start starts the consumer
	Start(ctx context.Context) error

	// Close closes the consumer
	Close() error

	// RegisterHandler registers a handler for a specific event type
	RegisterHandler(eventType EventType, handler EventHandler)
}

// EventHandler is a function that handles an event
type EventHandler func(event Event) error

// KafkaConsumer implements the Consumer interface using Kafka
type KafkaConsumer struct {
	consumer   sarama.ConsumerGroup
	topics     []string
	handlers   map[EventType]EventHandler
	handlersWG sync.WaitGroup
	handlersMu sync.RWMutex
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(cfg *config.Config) (Consumer, error) {
	// Create Kafka consumer configuration
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Create Kafka consumer group
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, "accounts-service", kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	return &KafkaConsumer{
		consumer: consumerGroup,
		topics:   []string{cfg.Kafka.Topic},
		handlers: make(map[EventType]EventHandler),
	}, nil
}

// Start starts the consumer
func (c *KafkaConsumer) Start(ctx context.Context) error {
	// Create a new context with cancellation
	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle signals for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming in a goroutine
	go func() {
		for {
			// Consume messages
			handler := &consumerGroupHandler{
				handlers:   c.handlers,
				handlersMu: &c.handlersMu,
				handlersWG: &c.handlersWG,
			}

			err := c.consumer.Consume(consumerCtx, c.topics, handler)
			if err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			// Check if context was cancelled, signaling that the consumer should stop
			if consumerCtx.Err() != nil {
				return
			}
		}
	}()

	// Wait for signals
	select {
	case <-signals:
		log.Println("Received signal, shutting down consumer...")
		cancel()
	case <-ctx.Done():
		log.Println("Context done, shutting down consumer...")
	}

	// Wait for all handlers to finish
	c.handlersWG.Wait()

	return nil
}

// Close closes the consumer
func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}

// RegisterHandler registers a handler for a specific event type
func (c *KafkaConsumer) RegisterHandler(eventType EventType, handler EventHandler) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	c.handlers[eventType] = handler
}

// consumerGroupHandler implements the sarama.ConsumerGroupHandler interface
type consumerGroupHandler struct {
	handlers   map[EventType]EventHandler
	handlersMu *sync.RWMutex
	handlersWG *sync.WaitGroup
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Parse the event
		var event Event
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			session.MarkMessage(message, "")
			continue
		}

		// Get the handler for this event type
		h.handlersMu.RLock()
		handler, ok := h.handlers[event.Type]
		h.handlersMu.RUnlock()

		if !ok {
			log.Printf("No handler registered for event type: %s", event.Type)
			session.MarkMessage(message, "")
			continue
		}

		// Handle the event in a goroutine
		h.handlersWG.Add(1)
		go func(msg *sarama.ConsumerMessage, evt Event, hdlr EventHandler) {
			defer h.handlersWG.Done()

			// Handle the event
			if err := hdlr(evt); err != nil {
				log.Printf("Failed to handle event: %v", err)
			}

			// Mark the message as processed
			session.MarkMessage(msg, "")
		}(message, event, handler)
	}

	return nil
}
