package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DimaJoyti/go-coffee/streams/config"
	"github.com/DimaJoyti/go-coffee/streams/models"

	"github.com/IBM/sarama"
)

// StreamProcessor is responsible for processing Kafka streams
type StreamProcessor struct {
	consumer sarama.ConsumerGroup
	producer sarama.SyncProducer
	config   *config.Config
	running  bool
	ready    chan bool
}

// NewStreamProcessor creates a new Kafka stream processor
func NewStreamProcessor(cfg *config.Config) (*StreamProcessor, error) {
	// Configure Sarama
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3

	// Create consumer group
	consumer, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.ApplicationID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	// Create producer
	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, config)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &StreamProcessor{
		consumer: consumer,
		producer: producer,
		config:   cfg,
		running:  false,
		ready:    make(chan bool),
	}, nil
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	processor *StreamProcessor
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.processor.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Process the message
			if err := h.processor.processOrder(message); err != nil {
				log.Printf("Error processing message: %v", err)
			} else {
				session.MarkMessage(message, "")
			}

		case <-session.Context().Done():
			return nil
		}
	}
}

// Start starts the stream processor
func (sp *StreamProcessor) Start() error {
	if sp.running {
		return fmt.Errorf("stream processor is already running")
	}

	sp.running = true

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create consumer group handler
	handler := &ConsumerGroupHandler{processor: sp}

	// Start consuming
	go func() {
		for {
			if err := sp.consumer.Consume(ctx, []string{sp.config.Kafka.InputTopic}, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
				return
			}

			if ctx.Err() != nil {
				return
			}

			sp.ready = make(chan bool)
		}
	}()

	// Wait for consumer to be ready
	<-sp.ready
	log.Println("Stream processor started successfully")

	// Set up signal handling
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigchan
	log.Println("Terminating stream processor...")
	cancel()

	return nil
}

// Stop stops the stream processor
func (sp *StreamProcessor) Stop() {
	if !sp.running {
		return
	}

	sp.running = false
	sp.consumer.Close()
	sp.producer.Close()
}

// processOrder processes a single order message
func (sp *StreamProcessor) processOrder(msg *sarama.ConsumerMessage) error {
	// Parse the order
	order, err := models.FromJSON(msg.Value)
	if err != nil {
		return fmt.Errorf("error parsing order: %w", err)
	}

	log.Printf("Processing order: %s for %s from topic %s partition %d offset %d\n",
		order.ID, order.CustomerName, msg.Topic, msg.Partition, msg.Offset)

	// Create a processed order
	processedOrder := models.NewProcessedOrder(order)

	// Convert to JSON
	processedOrderJSON, err := processedOrder.ToJSON()
	if err != nil {
		return fmt.Errorf("error serializing processed order: %w", err)
	}

	// Send to output topic
	producerMsg := &sarama.ProducerMessage{
		Topic: sp.config.Kafka.OutputTopic,
		Key:   sarama.ByteEncoder(msg.Key), // Preserve the original key
		Value: sarama.ByteEncoder(processedOrderJSON),
	}

	partition, offset, err := sp.producer.SendMessage(producerMsg)
	if err != nil {
		return fmt.Errorf("error producing message: %w", err)
	}

	log.Printf("Delivered processed order to topic %s [%d] at offset %d\n",
		sp.config.Kafka.OutputTopic, partition, offset)

	return nil
}
