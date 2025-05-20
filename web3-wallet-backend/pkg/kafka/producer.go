package kafka

import (
	"time"
)

// Producer represents a Kafka producer
type Producer interface {
	// Produce produces a message to Kafka
	Produce(topic string, key []byte, value []byte) error
	
	// ProduceAsync produces a message to Kafka asynchronously
	ProduceAsync(topic string, key []byte, value []byte, callback func(error))
	
	// Flush flushes the producer
	Flush(timeout time.Duration) error
	
	// Close closes the producer
	Close() error
}
