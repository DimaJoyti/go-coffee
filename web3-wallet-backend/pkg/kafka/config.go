package kafka

import (
	"time"
)

// Config represents Kafka configuration
type Config struct {
	Brokers            []string      // Kafka broker addresses
	Topic              string        // Default topic
	ConsumerGroup      string        // Consumer group ID
	AutoOffsetReset    string        // Auto offset reset (earliest, latest)
	EnableAutoCommit   bool          // Whether to enable auto commit
	AutoCommitInterval time.Duration // Auto commit interval
	SessionTimeout     time.Duration // Session timeout
	HeartbeatInterval  time.Duration // Heartbeat interval
	MaxPollInterval    time.Duration // Maximum poll interval
	MaxPollRecords     int           // Maximum poll records
	FetchMinBytes      int           // Minimum bytes to fetch
	FetchMaxBytes      int           // Maximum bytes to fetch
	FetchMaxWait       time.Duration // Maximum wait time for fetch
	RequiredAcks       string        // Required acks (none, local, all)
	RetryMax           int           // Maximum number of retries
	RetryBackoff       time.Duration // Retry backoff
	Compression        string        // Compression type (none, gzip, snappy, lz4, zstd)
	BatchSize          int           // Producer batch size
	BatchTimeout       time.Duration // Producer batch timeout
	ReadTimeout        time.Duration // Read timeout
	WriteTimeout       time.Duration // Write timeout
	DialTimeout        time.Duration // Dial timeout
	KeepAlive          time.Duration // Keep alive
}
