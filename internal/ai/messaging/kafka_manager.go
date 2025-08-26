package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaManager manages Kafka messaging for AI agents
type KafkaManager struct {
	config   config.KafkaConfig
	logger   *zap.Logger
	producer *kafka.Writer
	consumer *kafka.Reader
	
	// State management
	mutex   sync.RWMutex
	running bool
	
	// Message handlers
	messageHandlers map[string]MessageHandler
	
	// Channels
	incomingMessages chan *Message
	outgoingMessages chan *Message
	stopChan         chan struct{}
}

// Message represents a message in the AI system
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Content   map[string]interface{} `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeAgentCommunication MessageType = "agent_communication"
	MessageTypeTaskAssignment     MessageType = "task_assignment"
	MessageTypeTaskCompletion     MessageType = "task_completion"
	MessageTypeWorkflowEvent      MessageType = "workflow_event"
	MessageTypeSystemNotification MessageType = "system_notification"
	MessageTypeExternalIntegration MessageType = "external_integration"
)

// MessageHandler defines the interface for handling messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, message *Message) error
}

// MessageHandlerFunc is a function adapter for MessageHandler
type MessageHandlerFunc func(ctx context.Context, message *Message) error

func (f MessageHandlerFunc) HandleMessage(ctx context.Context, message *Message) error {
	return f(ctx, message)
}

// NewKafkaManager creates a new Kafka manager
func NewKafkaManager(config config.KafkaConfig, logger *zap.Logger) (*KafkaManager, error) {
	return &KafkaManager{
		config:           config,
		logger:           logger,
		messageHandlers:  make(map[string]MessageHandler),
		incomingMessages: make(chan *Message, 1000),
		outgoingMessages: make(chan *Message, 1000),
		stopChan:         make(chan struct{}),
	}, nil
}

// Start starts the Kafka manager
func (km *KafkaManager) Start(ctx context.Context) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	if km.running {
		return fmt.Errorf("Kafka manager is already running")
	}

	km.logger.Info("Starting Kafka manager...")

	// Initialize producer
	km.producer = &kafka.Writer{
		Addr:         kafka.TCP(km.config.Brokers...),
		Topic:        km.config.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	// Initialize consumer
	km.consumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     km.config.Brokers,
		Topic:       km.config.Topic,
		GroupID:     km.config.ConsumerGroup,
		StartOffset: kafka.LastOffset,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
	})

	// Start background workers
	go km.messageProducer(ctx)
	go km.messageConsumer(ctx)
	go km.messageProcessor(ctx)

	km.running = true
	km.logger.Info("Kafka manager started successfully")

	return nil
}

// Stop stops the Kafka manager
func (km *KafkaManager) Stop() {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	if !km.running {
		return
	}

	km.logger.Info("Stopping Kafka manager...")

	// Signal stop to all workers
	close(km.stopChan)

	// Close producer and consumer
	if km.producer != nil {
		km.producer.Close()
	}
	if km.consumer != nil {
		km.consumer.Close()
	}

	km.running = false
	km.logger.Info("Kafka manager stopped")
}

// PublishMessage publishes a message to Kafka
func (km *KafkaManager) PublishMessage(ctx context.Context, message *Message) error {
	if !km.running {
		return fmt.Errorf("Kafka manager is not running")
	}

	select {
	case km.outgoingMessages <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("outgoing message queue is full")
	}
}

// PublishAgentMessage publishes an agent message
func (km *KafkaManager) PublishAgentMessage(ctx context.Context, agentMessage interface{}) error {
	message := &Message{
		Type:      MessageTypeAgentCommunication,
		Content:   map[string]interface{}{"agent_message": agentMessage},
		Timestamp: time.Now(),
	}

	return km.PublishMessage(ctx, message)
}

// PublishTaskEvent publishes a task-related event
func (km *KafkaManager) PublishTaskEvent(ctx context.Context, eventType string, taskData interface{}) error {
	message := &Message{
		Type:      MessageTypeTaskAssignment,
		Content:   map[string]interface{}{"event_type": eventType, "task_data": taskData},
		Timestamp: time.Now(),
	}

	return km.PublishMessage(ctx, message)
}

// PublishWorkflowEvent publishes a workflow-related event
func (km *KafkaManager) PublishWorkflowEvent(ctx context.Context, eventType string, workflowData interface{}) error {
	message := &Message{
		Type:      MessageTypeWorkflowEvent,
		Content:   map[string]interface{}{"event_type": eventType, "workflow_data": workflowData},
		Timestamp: time.Now(),
	}

	return km.PublishMessage(ctx, message)
}

// PublishExternalIntegrationEvent publishes an external integration event
func (km *KafkaManager) PublishExternalIntegrationEvent(ctx context.Context, integration string, eventData interface{}) error {
	message := &Message{
		Type:      MessageTypeExternalIntegration,
		Source:    integration,
		Content:   map[string]interface{}{"integration": integration, "event_data": eventData},
		Timestamp: time.Now(),
	}

	return km.PublishMessage(ctx, message)
}

// RegisterMessageHandler registers a message handler for a specific message type
func (km *KafkaManager) RegisterMessageHandler(messageType string, handler MessageHandler) {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	km.messageHandlers[messageType] = handler
	km.logger.Info("Message handler registered", zap.String("message_type", messageType))
}

// messageProducer handles outgoing messages
func (km *KafkaManager) messageProducer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-km.stopChan:
			return
		case message := <-km.outgoingMessages:
			if err := km.sendMessage(ctx, message); err != nil {
				km.logger.Error("Failed to send message", zap.Error(err))
			}
		}
	}
}

// messageConsumer handles incoming messages
func (km *KafkaManager) messageConsumer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-km.stopChan:
			return
		default:
			if err := km.receiveMessage(ctx); err != nil {
				km.logger.Error("Failed to receive message", zap.Error(err))
				time.Sleep(1 * time.Second) // Backoff on error
			}
		}
	}
}

// messageProcessor processes incoming messages
func (km *KafkaManager) messageProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-km.stopChan:
			return
		case message := <-km.incomingMessages:
			km.processMessage(ctx, message)
		}
	}
}

// sendMessage sends a message to Kafka
func (km *KafkaManager) sendMessage(ctx context.Context, message *Message) error {
	// Serialize message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create Kafka message
	kafkaMessage := kafka.Message{
		Key:   []byte(message.ID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "message_type", Value: []byte(message.Type)},
			{Key: "source", Value: []byte(message.Source)},
			{Key: "target", Value: []byte(message.Target)},
		},
	}

	// Send message
	if err := km.producer.WriteMessages(ctx, kafkaMessage); err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	km.logger.Debug("Message sent to Kafka",
		zap.String("message_id", message.ID),
		zap.String("message_type", string(message.Type)),
	)

	return nil
}

// receiveMessage receives a message from Kafka
func (km *KafkaManager) receiveMessage(ctx context.Context) error {
	kafkaMessage, err := km.consumer.ReadMessage(ctx)
	if err != nil {
		return fmt.Errorf("failed to read message from Kafka: %w", err)
	}

	// Deserialize message
	var message Message
	if err := json.Unmarshal(kafkaMessage.Value, &message); err != nil {
		km.logger.Error("Failed to unmarshal message", zap.Error(err))
		return nil // Don't return error to avoid stopping the consumer
	}

	// Add to processing queue
	select {
	case km.incomingMessages <- &message:
		km.logger.Debug("Message received from Kafka",
			zap.String("message_id", message.ID),
			zap.String("message_type", string(message.Type)),
		)
	default:
		km.logger.Warn("Incoming message queue is full, dropping message",
			zap.String("message_id", message.ID),
		)
	}

	return nil
}

// processMessage processes an incoming message
func (km *KafkaManager) processMessage(ctx context.Context, message *Message) {
	km.mutex.RLock()
	handler, exists := km.messageHandlers[string(message.Type)]
	km.mutex.RUnlock()

	if !exists {
		km.logger.Debug("No handler registered for message type",
			zap.String("message_type", string(message.Type)),
		)
		return
	}

	if err := handler.HandleMessage(ctx, message); err != nil {
		km.logger.Error("Message handler failed",
			zap.String("message_id", message.ID),
			zap.String("message_type", string(message.Type)),
			zap.Error(err),
		)
	}
}

// GetStats returns Kafka manager statistics
func (km *KafkaManager) GetStats() map[string]interface{} {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	return map[string]interface{}{
		"running":                km.running,
		"registered_handlers":    len(km.messageHandlers),
		"incoming_queue_size":    len(km.incomingMessages),
		"outgoing_queue_size":    len(km.outgoingMessages),
		"incoming_queue_capacity": cap(km.incomingMessages),
		"outgoing_queue_capacity": cap(km.outgoingMessages),
	}
}
