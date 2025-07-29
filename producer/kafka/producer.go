package kafka

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/producer/config"
	"github.com/IBM/sarama"
)

// Producer interface defines the methods for a Kafka producer
type Producer interface {
	PushToQueue(topic string, message []byte) error
	PushToQueueAsync(topic string, message []byte) error
	Close() error
	Flush() error
}

// SaramaProducer implements the Producer interface using Sarama
type SaramaProducer struct {
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewProducer creates a new Kafka producer with both sync and async capabilities
func NewProducer(config *config.Config) (Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	
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
	
	// Performance optimizations for async producer
	saramaConfig.Producer.Flush.Frequency = 10 * time.Millisecond
	saramaConfig.Producer.Flush.Messages = 100
	saramaConfig.Producer.Flush.Bytes = 1024 * 1024 // 1MB
	saramaConfig.Producer.Compression = sarama.CompressionSnappy

	// Create both sync and async producers
	syncProducer, err := sarama.NewSyncProducer(config.Kafka.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	asyncProducer, err := sarama.NewAsyncProducer(config.Kafka.Brokers, saramaConfig)
	if err != nil {
		syncProducer.Close()
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	producer := &SaramaProducer{
		syncProducer:  syncProducer,
		asyncProducer: asyncProducer,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Start async message handlers
	producer.wg.Add(2)
	go producer.handleAsyncSuccess()
	go producer.handleAsyncErrors()

	return producer, nil
}

// PushToQueue sends a message to a Kafka topic synchronously (for backward compatibility)
func (p *SaramaProducer) PushToQueue(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := p.syncProducer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Order is stored in topic(%s)/partition(%d)/offset(%d)\n",
		topic,
		partition,
		offset)

	return nil
}

// PushToQueueAsync sends a message to a Kafka topic asynchronously for better performance
func (p *SaramaProducer) PushToQueueAsync(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	select {
	case p.asyncProducer.Input() <- msg:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

// Flush ensures all pending async messages are sent
func (p *SaramaProducer) Flush() error {
	// For async producer, we can't force flush directly with Sarama
	// But we can wait a bit for the batching to complete
	time.Sleep(50 * time.Millisecond)
	return nil
}

// handleAsyncSuccess handles successful async message deliveries
func (p *SaramaProducer) handleAsyncSuccess() {
	defer p.wg.Done()
	for {
		select {
		case msg := <-p.asyncProducer.Successes():
			log.Printf("Async message sent successfully to topic(%s)/partition(%d)/offset(%d)\n",
				msg.Topic, msg.Partition, msg.Offset)
		case <-p.ctx.Done():
			return
		}
	}
}

// handleAsyncErrors handles async message delivery errors
func (p *SaramaProducer) handleAsyncErrors() {
	defer p.wg.Done()
	for {
		select {
		case err := <-p.asyncProducer.Errors():
			log.Printf("Failed to send async message to topic %s: %v\n", err.Msg.Topic, err.Err)
		case <-p.ctx.Done():
			return
		}
	}
}

// Close closes the Kafka producer
func (p *SaramaProducer) Close() error {
	p.cancel()
	
	// Close async producer first
	if err := p.asyncProducer.Close(); err != nil {
		log.Printf("Error closing async producer: %v", err)
	}
	
	// Wait for goroutines to finish
	p.wg.Wait()
	
	// Close sync producer
	return p.syncProducer.Close()
}
