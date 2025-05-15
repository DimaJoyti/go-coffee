package kafka

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kafka_streams/config"
	"kafka_streams/models"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// StreamProcessor is responsible for processing Kafka streams
type StreamProcessor struct {
	consumer *kafka.Consumer
	producer *kafka.Producer
	config   *config.Config
	running  bool
}

// NewStreamProcessor creates a new Kafka stream processor
func NewStreamProcessor(cfg *config.Config) (*StreamProcessor, error) {
	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        cfg.Kafka.Brokers[0],
		"group.id":                 cfg.Kafka.ApplicationID,
		"auto.offset.reset":        cfg.Kafka.AutoOffsetReset,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"session.timeout.ms":       30000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Create Kafka producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka.Brokers[0],
		"acks":              "all",
	})
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &StreamProcessor{
		consumer: consumer,
		producer: producer,
		config:   cfg,
		running:  false,
	}, nil
}

// Start starts the stream processor
func (sp *StreamProcessor) Start() error {
	if sp.running {
		return fmt.Errorf("stream processor is already running")
	}

	// Subscribe to input topic
	err := sp.consumer.Subscribe(sp.config.Kafka.InputTopic, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	// Set up signal handler for graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Start processing
	sp.running = true
	go sp.processMessages(sigchan)

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

// processMessages processes messages from the input topic
func (sp *StreamProcessor) processMessages(sigchan chan os.Signal) {
	defer func() {
		sp.running = false
		sp.consumer.Close()
		sp.producer.Close()
	}()

	for sp.running {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating\n", sig)
			return
		default:
			// Poll for messages
			msg, err := sp.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Timeout or error
				if err.(kafka.Error).Code() != kafka.ErrTimedOut {
					log.Printf("Error reading message: %v\n", err)
				}
				continue
			}

			// Process the message
			if err := sp.processOrder(msg); err != nil {
				log.Printf("Error processing message: %v\n", err)
			}
		}
	}
}

// processOrder processes a single order message
func (sp *StreamProcessor) processOrder(msg *kafka.Message) error {
	// Parse the order
	order, err := models.FromJSON(msg.Value)
	if err != nil {
		return fmt.Errorf("error parsing order: %w", err)
	}

	log.Printf("Processing order: %s for %s\n", order.ID, order.CustomerName)

	// Create a processed order
	processedOrder := models.NewProcessedOrder(order)

	// Convert to JSON
	processedOrderJSON, err := processedOrder.ToJSON()
	if err != nil {
		return fmt.Errorf("error serializing processed order: %w", err)
	}

	// Send to output topic
	err = sp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &sp.config.Kafka.OutputTopic,
			Partition: kafka.PartitionAny,
		},
		Value: processedOrderJSON,
		Key:   msg.Key, // Preserve the original key
	}, nil)
	if err != nil {
		return fmt.Errorf("error producing message: %w", err)
	}

	return nil
}

// DeliveryReportHandler handles delivery reports from the producer
func (sp *StreamProcessor) DeliveryReportHandler() {
	for e := range sp.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				log.Printf("Delivery failed: %v\n", ev.TopicPartition.Error)
			} else {
				log.Printf("Delivered message to topic %s [%d] at offset %v\n",
					*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
			}
		}
	}
}
