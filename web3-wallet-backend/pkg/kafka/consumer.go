package kafka

import (
	"time"
)

// Consumer represents a Kafka consumer
type Consumer interface {
	// Subscribe subscribes to topics
	Subscribe(topics []string) error
	
	// Poll polls for messages
	Poll(timeout time.Duration) (Message, error)
	
	// Commit commits offsets
	Commit() error
	
	// CommitMessage commits a message
	CommitMessage(msg Message) error
	
	// Close closes the consumer
	Close() error
}

// Message represents a Kafka message
type Message interface {
	// Topic returns the topic
	Topic() string
	
	// Partition returns the partition
	Partition() int32
	
	// Offset returns the offset
	Offset() int64
	
	// Key returns the key
	Key() []byte
	
	// Value returns the value
	Value() []byte
	
	// Timestamp returns the timestamp
	Timestamp() time.Time
	
	// Headers returns the headers
	Headers() []Header
}

// Header represents a Kafka message header
type Header interface {
	// Key returns the key
	Key() string
	
	// Value returns the value
	Value() []byte
}
