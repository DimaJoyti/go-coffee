package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go-coffee-ai-agents/internal/config"
	"go-coffee-ai-agents/internal/observability"
)

// ConnectionManager manages Kafka connections and configuration
type ConnectionManager struct {
	config  config.KafkaConfig
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
	
	// Connection pools
	writers map[string]*kafka.Writer
	readers map[string]*kafka.Reader
}

// NewConnectionManager creates a new Kafka connection manager
func NewConnectionManager(
	config config.KafkaConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *ConnectionManager {
	return &ConnectionManager{
		config:  config,
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
		writers: make(map[string]*kafka.Writer),
		readers: make(map[string]*kafka.Reader),
	}
}

// Initialize initializes the Kafka connection manager
func (cm *ConnectionManager) Initialize(ctx context.Context) error {
	ctx, span := cm.tracing.StartKafkaSpan(ctx, "INITIALIZE", "kafka")
	defer span.End()

	cm.logger.InfoContext(ctx, "Initializing Kafka connection manager",
		"brokers", cm.config.Brokers,
		"client_id", cm.config.ClientID)

	// Test connection to Kafka brokers
	if err := cm.testConnection(ctx); err != nil {
		cm.tracing.RecordError(span, err, "Failed to connect to Kafka brokers")
		return fmt.Errorf("failed to connect to Kafka brokers: %w", err)
	}

	cm.tracing.RecordSuccess(span, "Kafka connection manager initialized")
	cm.logger.InfoContext(ctx, "Kafka connection manager initialized successfully")

	return nil
}

// testConnection tests the connection to Kafka brokers
func (cm *ConnectionManager) testConnection(ctx context.Context) error {
	ctx, span := cm.tracing.StartKafkaSpan(ctx, "TEST_CONNECTION", "kafka")
	defer span.End()

	start := time.Now()

	// Create a temporary connection to test connectivity
	conn, err := kafka.DialContext(ctx, "tcp", cm.config.Brokers[0])
	if err != nil {
		duration := time.Since(start)
		cm.tracing.RecordError(span, err, "Failed to dial Kafka broker")
		cm.logger.ErrorContext(ctx, "Failed to dial Kafka broker", err,
			"broker", cm.config.Brokers[0],
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to dial Kafka broker: %w", err)
	}
	defer conn.Close()

	// Test broker metadata
	brokers, err := conn.Brokers()
	if err != nil {
		duration := time.Since(start)
		cm.tracing.RecordError(span, err, "Failed to get broker metadata")
		cm.logger.ErrorContext(ctx, "Failed to get broker metadata", err,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to get broker metadata: %w", err)
	}

	duration := time.Since(start)
	cm.tracing.RecordSuccess(span, "Kafka connection test successful")
	cm.logger.InfoContext(ctx, "Kafka connection test successful",
		"brokers_count", len(brokers),
		"duration_ms", duration.Milliseconds())

	return nil
}

// GetWriter returns a Kafka writer for the specified topic
func (cm *ConnectionManager) GetWriter(topic string) *kafka.Writer {
	if writer, exists := cm.writers[topic]; exists {
		return writer
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cm.config.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cm.config.BatchSize,
		BatchTimeout: cm.config.BatchTimeout,
		ReadTimeout:  cm.config.ReadTimeout,
		WriteTimeout: cm.config.WriteTimeout,
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Synchronous writes for reliability
	}

	// Configure SASL if enabled
	if cm.config.EnableSASL {
		writer.Transport = cm.createTransport()
	}

	cm.writers[topic] = writer
	
	cm.logger.Debug("Created Kafka writer",
		"topic", topic,
		"brokers", cm.config.Brokers,
		"batch_size", cm.config.BatchSize)

	return writer
}

// GetReader returns a Kafka reader for the specified topic
func (cm *ConnectionManager) GetReader(topic string, groupID string) *kafka.Reader {
	readerKey := fmt.Sprintf("%s-%s", topic, groupID)
	if reader, exists := cm.readers[readerKey]; exists {
		return reader
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:        cm.config.Brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	}

	// Configure SASL if enabled
	if cm.config.EnableSASL {
		readerConfig.Dialer = &kafka.Dialer{
			Timeout:       cm.config.ConnectTimeout,
			DualStack:     true,
			TLS:           cm.createTLSConfig(),
			SASLMechanism: cm.createSASLMechanism(),
		}
	}

	reader := kafka.NewReader(readerConfig)

	cm.readers[readerKey] = reader
	
	cm.logger.Debug("Created Kafka reader",
		"topic", topic,
		"group_id", groupID,
		"brokers", cm.config.Brokers)

	return reader
}

// createTransport creates a Kafka transport with SASL and TLS configuration
func (cm *ConnectionManager) createTransport() *kafka.Transport {
	transport := &kafka.Transport{
		SASL: cm.createSASLMechanism(),
	}

	if cm.config.EnableTLS {
		transport.TLS = cm.createTLSConfig()
	}

	return transport
}

// createSASLMechanism creates the appropriate SASL mechanism
func (cm *ConnectionManager) createSASLMechanism() sasl.Mechanism {
	switch cm.config.SASLMechanism {
	case "PLAIN":
		return plain.Mechanism{
			Username: cm.config.SASLUsername,
			Password: cm.config.SASLPassword,
		}
	case "SCRAM-SHA-256":
		mechanism, err := scram.Mechanism(scram.SHA256, cm.config.SASLUsername, cm.config.SASLPassword)
		if err != nil {
			cm.logger.Error("Failed to create SCRAM-SHA-256 mechanism", err)
			return nil
		}
		return mechanism
	case "SCRAM-SHA-512":
		mechanism, err := scram.Mechanism(scram.SHA512, cm.config.SASLUsername, cm.config.SASLPassword)
		if err != nil {
			cm.logger.Error("Failed to create SCRAM-SHA-512 mechanism", err)
			return nil
		}
		return mechanism
	default:
		cm.logger.Warn("Unknown SASL mechanism, falling back to PLAIN",
			"mechanism", cm.config.SASLMechanism)
		return plain.Mechanism{
			Username: cm.config.SASLUsername,
			Password: cm.config.SASLPassword,
		}
	}
}

// createTLSConfig creates TLS configuration for Kafka connections
func (cm *ConnectionManager) createTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: cm.config.TLSSkipVerify,
	}
}

// HealthCheck performs a health check on Kafka connections
func (cm *ConnectionManager) HealthCheck(ctx context.Context) error {
	ctx, span := cm.tracing.StartKafkaSpan(ctx, "HEALTH_CHECK", "kafka")
	defer span.End()

	start := time.Now()

	// Test connection to each broker
	for i, broker := range cm.config.Brokers {
		conn, err := kafka.DialContext(ctx, "tcp", broker)
		if err != nil {
			duration := time.Since(start)
			cm.tracing.RecordError(span, err, fmt.Sprintf("Failed to connect to broker %d", i))
			cm.logger.ErrorContext(ctx, "Kafka health check failed", err,
				"broker", broker,
				"broker_index", i,
				"duration_ms", duration.Milliseconds())
			return fmt.Errorf("failed to connect to broker %s: %w", broker, err)
		}
		conn.Close()
	}

	duration := time.Since(start)
	cm.tracing.RecordSuccess(span, "Kafka health check passed")
	cm.logger.DebugContext(ctx, "Kafka health check passed",
		"brokers_checked", len(cm.config.Brokers),
		"duration_ms", duration.Milliseconds())

	return nil
}

// GetConnectionStats returns connection statistics
func (cm *ConnectionManager) GetConnectionStats() ConnectionStats {
	return ConnectionStats{
		Brokers:       cm.config.Brokers,
		ClientID:      cm.config.ClientID,
		WritersCount:  len(cm.writers),
		ReadersCount:  len(cm.readers),
		SASLEnabled:   cm.config.EnableSASL,
		TLSEnabled:    cm.config.EnableTLS,
		BatchSize:     cm.config.BatchSize,
		BatchTimeout:  cm.config.BatchTimeout,
		ReadTimeout:   cm.config.ReadTimeout,
		WriteTimeout:  cm.config.WriteTimeout,
	}
}

// Close closes all Kafka connections
func (cm *ConnectionManager) Close(ctx context.Context) error {
	ctx, span := cm.tracing.StartKafkaSpan(ctx, "CLOSE", "kafka")
	defer span.End()

	cm.logger.InfoContext(ctx, "Closing Kafka connections")

	var errors []error

	// Close all writers
	for topic, writer := range cm.writers {
		if err := writer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close writer for topic %s: %w", topic, err))
		}
	}

	// Close all readers
	for key, reader := range cm.readers {
		if err := reader.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close reader %s: %w", key, err))
		}
	}

	if len(errors) > 0 {
		err := fmt.Errorf("errors closing Kafka connections: %v", errors)
		cm.tracing.RecordError(span, err, "Failed to close some Kafka connections")
		cm.logger.ErrorContext(ctx, "Failed to close some Kafka connections", err)
		return err
	}

	cm.tracing.RecordSuccess(span, "All Kafka connections closed")
	cm.logger.InfoContext(ctx, "All Kafka connections closed successfully")

	return nil
}

// ConnectionStats represents Kafka connection statistics
type ConnectionStats struct {
	Brokers       []string      `json:"brokers"`
	ClientID      string        `json:"client_id"`
	WritersCount  int           `json:"writers_count"`
	ReadersCount  int           `json:"readers_count"`
	SASLEnabled   bool          `json:"sasl_enabled"`
	TLSEnabled    bool          `json:"tls_enabled"`
	BatchSize     int           `json:"batch_size"`
	BatchTimeout  time.Duration `json:"batch_timeout"`
	ReadTimeout   time.Duration `json:"read_timeout"`
	WriteTimeout  time.Duration `json:"write_timeout"`
}

// TopicManager manages Kafka topic operations
type TopicManager struct {
	connectionManager *ConnectionManager
	logger            *observability.StructuredLogger
	tracing           *observability.TracingHelper
}

// NewTopicManager creates a new topic manager
func NewTopicManager(
	connectionManager *ConnectionManager,
	logger *observability.StructuredLogger,
	tracing *observability.TracingHelper,
) *TopicManager {
	return &TopicManager{
		connectionManager: connectionManager,
		logger:            logger,
		tracing:           tracing,
	}
}

// CreateTopic creates a Kafka topic if it doesn't exist
func (tm *TopicManager) CreateTopic(ctx context.Context, topic string, partitions int, replicationFactor int) error {
	ctx, span := tm.tracing.StartKafkaSpan(ctx, "CREATE_TOPIC", topic)
	defer span.End()

	start := time.Now()

	tm.logger.InfoContext(ctx, "Creating Kafka topic",
		"topic", topic,
		"partitions", partitions,
		"replication_factor", replicationFactor)

	// Connect to Kafka
	conn, err := kafka.DialContext(ctx, "tcp", tm.connectionManager.config.Brokers[0])
	if err != nil {
		duration := time.Since(start)
		tm.tracing.RecordError(span, err, "Failed to connect to Kafka for topic creation")
		tm.logger.ErrorContext(ctx, "Failed to connect to Kafka for topic creation", err,
			"topic", topic,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	// Create topic configuration
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	// Create the topic
	err = conn.CreateTopics(topicConfig)
	duration := time.Since(start)

	if err != nil {
		tm.tracing.RecordError(span, err, "Failed to create Kafka topic")
		tm.logger.ErrorContext(ctx, "Failed to create Kafka topic", err,
			"topic", topic,
			"partitions", partitions,
			"replication_factor", replicationFactor,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to create topic %s: %w", topic, err)
	}

	tm.tracing.RecordSuccess(span, "Kafka topic created successfully")
	tm.logger.InfoContext(ctx, "Kafka topic created successfully",
		"topic", topic,
		"partitions", partitions,
		"replication_factor", replicationFactor,
		"duration_ms", duration.Milliseconds())

	return nil
}

// ListTopics lists all available Kafka topics
func (tm *TopicManager) ListTopics(ctx context.Context) ([]string, error) {
	ctx, span := tm.tracing.StartKafkaSpan(ctx, "LIST_TOPICS", "kafka")
	defer span.End()

	start := time.Now()

	// Connect to Kafka
	conn, err := kafka.DialContext(ctx, "tcp", tm.connectionManager.config.Brokers[0])
	if err != nil {
		duration := time.Since(start)
		tm.tracing.RecordError(span, err, "Failed to connect to Kafka for topic listing")
		tm.logger.ErrorContext(ctx, "Failed to connect to Kafka for topic listing", err,
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	// Get topic metadata
	partitions, err := conn.ReadPartitions()
	if err != nil {
		duration := time.Since(start)
		tm.tracing.RecordError(span, err, "Failed to read Kafka partitions")
		tm.logger.ErrorContext(ctx, "Failed to read Kafka partitions", err,
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	// Extract unique topic names
	topicSet := make(map[string]bool)
	for _, partition := range partitions {
		topicSet[partition.Topic] = true
	}

	topics := make([]string, 0, len(topicSet))
	for topic := range topicSet {
		topics = append(topics, topic)
	}

	duration := time.Since(start)
	tm.tracing.RecordSuccess(span, "Kafka topics listed successfully")
	tm.logger.DebugContext(ctx, "Kafka topics listed successfully",
		"topics_count", len(topics),
		"duration_ms", duration.Milliseconds())

	return topics, nil
}
