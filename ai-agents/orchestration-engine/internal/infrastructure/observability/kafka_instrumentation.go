package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	
	"go-coffee-ai-agents/orchestration-engine/internal/common"
)

// Logger interface for logging
type Logger = common.Logger

// KafkaInstrumentation provides OpenTelemetry instrumentation for Kafka
type KafkaInstrumentation struct {
	telemetry *TelemetryManager
	logger    Logger
}

// NewKafkaInstrumentation creates a new Kafka instrumentation
func NewKafkaInstrumentation(telemetry *TelemetryManager, logger Logger) *KafkaInstrumentation {
	return &KafkaInstrumentation{
		telemetry: telemetry,
		logger:    logger,
	}
}

// InstrumentedProducer wraps a Kafka producer with instrumentation
type InstrumentedProducer struct {
	producer      sarama.SyncProducer
	instrumentation *KafkaInstrumentation
}

// InstrumentedConsumer wraps a Kafka consumer with instrumentation
type InstrumentedConsumer struct {
	consumer        sarama.Consumer
	instrumentation *KafkaInstrumentation
}

// InstrumentedConsumerGroup wraps a Kafka consumer group with instrumentation
type InstrumentedConsumerGroup struct {
	consumerGroup   sarama.ConsumerGroup
	instrumentation *KafkaInstrumentation
}

// NewInstrumentedProducer creates an instrumented Kafka producer
func (ki *KafkaInstrumentation) NewInstrumentedProducer(producer sarama.SyncProducer) *InstrumentedProducer {
	return &InstrumentedProducer{
		producer:        producer,
		instrumentation: ki,
	}
}

// NewInstrumentedConsumer creates an instrumented Kafka consumer
func (ki *KafkaInstrumentation) NewInstrumentedConsumer(consumer sarama.Consumer) *InstrumentedConsumer {
	return &InstrumentedConsumer{
		consumer:        consumer,
		instrumentation: ki,
	}
}

// NewInstrumentedConsumerGroup creates an instrumented Kafka consumer group
func (ki *KafkaInstrumentation) NewInstrumentedConsumerGroup(consumerGroup sarama.ConsumerGroup) *InstrumentedConsumerGroup {
	return &InstrumentedConsumerGroup{
		consumerGroup:   consumerGroup,
		instrumentation: ki,
	}
}

// SendMessage sends a message with instrumentation
func (ip *InstrumentedProducer) SendMessage(ctx context.Context, msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	start := time.Now()
	
	// Start span for producer operation
	spanName := fmt.Sprintf("kafka.produce %s", msg.Topic)
	ctx, span := ip.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", msg.Topic),
			attribute.String("messaging.operation", "publish"),
		),
	)
	defer span.End()

	// Inject trace context into message headers
	if msg.Headers == nil {
		msg.Headers = make([]sarama.RecordHeader, 0)
	}

	// Create a map to hold the propagated context
	carrier := make(propagation.MapCarrier)
	ip.instrumentation.telemetry.config.ResourceAttributes = make(map[string]string)
	
	// Inject the trace context
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	propagator.Inject(ctx, carrier)

	// Add trace context to Kafka headers
	for key, value := range carrier {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	// Add message attributes to span
	if msg.Key != nil {
		keyBytes, err := msg.Key.Encode()
		if err == nil {
			span.SetAttributes(attribute.String("messaging.kafka.message_key", string(keyBytes)))
		}
	}
	if msg.Partition != -1 {
		span.SetAttributes(attribute.Int("messaging.kafka.partition", int(msg.Partition)))
	}

	// Send the message
	partition, offset, err = ip.producer.SendMessage(msg)
	
	duration := time.Since(start)

	// Record metrics
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ip.instrumentation.telemetry.RecordError(ctx, "kafka_producer_error", "kafka")
	}

	// Update span with result
	span.SetAttributes(
		attribute.Int("messaging.kafka.partition", int(partition)),
		attribute.Int64("messaging.kafka.offset", offset),
		attribute.String("messaging.kafka.status", status),
		attribute.Float64("messaging.kafka.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	// Record custom metrics
	ip.instrumentation.recordProducerMetrics(ctx, msg.Topic, status, duration)

	return partition, offset, err
}

// ConsumePartition consumes from a partition with instrumentation
func (ic *InstrumentedConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	partitionConsumer, err := ic.consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return nil, err
	}

	return &InstrumentedPartitionConsumer{
		partitionConsumer: partitionConsumer,
		instrumentation:   ic.instrumentation,
		topic:            topic,
		partition:        partition,
	}, nil
}

// InstrumentedPartitionConsumer wraps a partition consumer with instrumentation
type InstrumentedPartitionConsumer struct {
	partitionConsumer sarama.PartitionConsumer
	instrumentation   *KafkaInstrumentation
	topic            string
	partition        int32
}

// Messages returns the read channel for the partition consumer with instrumentation
func (ipc *InstrumentedPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	originalChan := ipc.partitionConsumer.Messages()
	instrumentedChan := make(chan *sarama.ConsumerMessage)

	go func() {
		defer close(instrumentedChan)
		for msg := range originalChan {
			// Instrument the message consumption
			ctx := ipc.instrumentMessage(msg)
			
			// Add context to message (if possible through custom wrapper)
			instrumentedChan <- msg
			
			// Record consumption metrics
			ipc.instrumentation.recordConsumerMetrics(ctx, msg.Topic, "success", time.Since(msg.Timestamp))
		}
	}()

	return instrumentedChan
}

// instrumentMessage creates instrumentation for a consumed message
func (ipc *InstrumentedPartitionConsumer) instrumentMessage(msg *sarama.ConsumerMessage) context.Context {
	// Extract trace context from message headers
	carrier := make(propagation.MapCarrier)
	for _, header := range msg.Headers {
		carrier[string(header.Key)] = string(header.Value)
	}

	// Extract context
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	ctx := propagator.Extract(context.Background(), carrier)

	// Start span for consumer operation
	spanName := fmt.Sprintf("kafka.consume %s", msg.Topic)
	ctx, span := ipc.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", msg.Topic),
			attribute.String("messaging.operation", "receive"),
			attribute.Int("messaging.kafka.partition", int(msg.Partition)),
			attribute.Int64("messaging.kafka.offset", msg.Offset),
			attribute.String("messaging.kafka.consumer_group", ""), // Add if available
		),
	)

	// Add message attributes
	if msg.Key != nil {
		span.SetAttributes(attribute.String("messaging.kafka.message_key", string(msg.Key)))
	}
	span.SetAttributes(
		attribute.Int64("messaging.kafka.message_size", int64(len(msg.Value))),
		attribute.String("messaging.kafka.timestamp", msg.Timestamp.Format(time.RFC3339)),
	)

	// The span should be ended by the message processor
	// For now, we'll end it after a short delay to capture the consumption
	go func() {
		time.Sleep(100 * time.Millisecond)
		span.End()
	}()

	return ctx
}

// Errors returns the errors channel with instrumentation
func (ipc *InstrumentedPartitionConsumer) Errors() <-chan *sarama.ConsumerError {
	originalChan := ipc.partitionConsumer.Errors()
	instrumentedChan := make(chan *sarama.ConsumerError)

	go func() {
		defer close(instrumentedChan)
		for err := range originalChan {
			// Record error metrics
			ctx := context.Background()
			ipc.instrumentation.telemetry.RecordError(ctx, "kafka_consumer_error", "kafka")
			ipc.instrumentation.recordConsumerMetrics(ctx, err.Topic, "error", 0)
			
			instrumentedChan <- err
		}
	}()

	return instrumentedChan
}

// HighWaterMarkOffset returns the high water mark offset
func (ipc *InstrumentedPartitionConsumer) HighWaterMarkOffset() int64 {
	return ipc.partitionConsumer.HighWaterMarkOffset()
}

// Close closes the partition consumer
func (ipc *InstrumentedPartitionConsumer) Close() error {
	return ipc.partitionConsumer.Close()
}

// AsyncClose initiates a shutdown of the partition consumer
func (ipc *InstrumentedPartitionConsumer) AsyncClose() {
	ipc.partitionConsumer.AsyncClose()
}

// IsPaused returns whether the partition consumer is paused
func (ipc *InstrumentedPartitionConsumer) IsPaused() bool {
	return ipc.partitionConsumer.IsPaused()
}

// Pause pauses the partition consumer
func (ipc *InstrumentedPartitionConsumer) Pause() {
	ipc.partitionConsumer.Pause()
}

// Resume resumes the partition consumer
func (ipc *InstrumentedPartitionConsumer) Resume() {
	ipc.partitionConsumer.Resume()
}

// Consume consumes messages from topics with instrumentation
func (icg *InstrumentedConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	// Wrap the handler with instrumentation
	instrumentedHandler := &InstrumentedConsumerGroupHandler{
		handler:         handler,
		instrumentation: icg.instrumentation,
	}

	return icg.consumerGroup.Consume(ctx, topics, instrumentedHandler)
}

// InstrumentedConsumerGroupHandler wraps a consumer group handler with instrumentation
type InstrumentedConsumerGroupHandler struct {
	handler         sarama.ConsumerGroupHandler
	instrumentation *KafkaInstrumentation
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (icgh *InstrumentedConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	ctx := context.Background()
	_, span := icgh.instrumentation.telemetry.StartSpan(ctx, "kafka.consumer_group.setup",
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.kafka.consumer_group", session.Context().Value("group_id").(string)),
		),
	)
	defer span.End()

	return icgh.handler.Setup(session)
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (icgh *InstrumentedConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	ctx := context.Background()
	_, span := icgh.instrumentation.telemetry.StartSpan(ctx, "kafka.consumer_group.cleanup",
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.kafka.consumer_group", session.Context().Value("group_id").(string)),
		),
	)
	defer span.End()

	return icgh.handler.Cleanup(session)
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (icgh *InstrumentedConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()
	_, span := icgh.instrumentation.telemetry.StartSpan(ctx, "kafka.consumer_group.consume_claim",
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", claim.Topic()),
			attribute.Int("messaging.kafka.partition", int(claim.Partition())),
			attribute.String("messaging.kafka.consumer_group", session.Context().Value("group_id").(string)),
		),
	)
	defer span.End()

	// Wrap the claim with instrumentation
	instrumentedClaim := &InstrumentedConsumerGroupClaim{
		claim:           claim,
		instrumentation: icgh.instrumentation,
	}

	return icgh.handler.ConsumeClaim(session, instrumentedClaim)
}

// InstrumentedConsumerGroupClaim wraps a consumer group claim with instrumentation
type InstrumentedConsumerGroupClaim struct {
	claim           sarama.ConsumerGroupClaim
	instrumentation *KafkaInstrumentation
}

// Topic returns the consumed topic name
func (icgc *InstrumentedConsumerGroupClaim) Topic() string {
	return icgc.claim.Topic()
}

// Partition returns the consumed partition
func (icgc *InstrumentedConsumerGroupClaim) Partition() int32 {
	return icgc.claim.Partition()
}

// InitialOffset returns the initial offset
func (icgc *InstrumentedConsumerGroupClaim) InitialOffset() int64 {
	return icgc.claim.InitialOffset()
}

// HighWaterMarkOffset returns the high water mark offset
func (icgc *InstrumentedConsumerGroupClaim) HighWaterMarkOffset() int64 {
	return icgc.claim.HighWaterMarkOffset()
}

// Messages returns the read channel for the partition consumer with instrumentation
func (icgc *InstrumentedConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	originalChan := icgc.claim.Messages()
	instrumentedChan := make(chan *sarama.ConsumerMessage)

	go func() {
		defer close(instrumentedChan)
		for msg := range originalChan {
			start := time.Now()
			
			// Extract trace context from message headers
			carrier := make(propagation.MapCarrier)
			for _, header := range msg.Headers {
				carrier[string(header.Key)] = string(header.Value)
			}

			// Extract context
			propagator := propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			)
			ctx := propagator.Extract(context.Background(), carrier)

			// Start span for message processing
			spanName := fmt.Sprintf("kafka.process %s", msg.Topic)
			ctx, span := icgc.instrumentation.telemetry.StartSpan(ctx, spanName,
				trace.WithAttributes(
					attribute.String("messaging.system", "kafka"),
					attribute.String("messaging.destination", msg.Topic),
					attribute.String("messaging.operation", "process"),
					attribute.Int("messaging.kafka.partition", int(msg.Partition)),
					attribute.Int64("messaging.kafka.offset", msg.Offset),
				),
			)

			// Add message to instrumented channel
			instrumentedChan <- msg

			// Record metrics
			duration := time.Since(start)
			icgc.instrumentation.recordConsumerMetrics(ctx, msg.Topic, "success", duration)

			span.End()
		}
	}()

	return instrumentedChan
}

// recordProducerMetrics records producer-specific metrics
func (ki *KafkaInstrumentation) recordProducerMetrics(_ context.Context, topic, status string, duration time.Duration) {
	if !ki.telemetry.config.MetricsEnabled {
		return
	}

	// Record using the telemetry manager's custom metrics
	// Log the Kafka producer operation for now
	ki.logger.Debug("Kafka producer operation recorded",
		"topic", topic,
		"operation", "produce",
		"status", status,
		"duration_ms", duration.Milliseconds(),
	)
}

// recordConsumerMetrics records consumer-specific metrics
func (ki *KafkaInstrumentation) recordConsumerMetrics(_ context.Context, topic, status string, duration time.Duration) {
	if !ki.telemetry.config.MetricsEnabled {
		return
	}

	// Record using the telemetry manager's custom metrics
	// Log the Kafka consumer operation for now
	ki.logger.Debug("Kafka consumer operation recorded",
		"topic", topic,
		"operation", "consume",
		"status", status,
		"duration_ms", duration.Milliseconds(),
	)
}

// Close closes the instrumented producer
func (ip *InstrumentedProducer) Close() error {
	return ip.producer.Close()
}

// Close closes the instrumented consumer
func (ic *InstrumentedConsumer) Close() error {
	return ic.consumer.Close()
}

// Close closes the instrumented consumer group
func (icg *InstrumentedConsumerGroup) Close() error {
	return icg.consumerGroup.Close()
}
