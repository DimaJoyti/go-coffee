package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// KafkaMessage represents a message from/to Kafka
type KafkaMessage struct {
	Topic     string                 `json:"topic"`
	Key       string                 `json:"key,omitempty"`
	Value     json.RawMessage        `json:"value"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageHandler defines the interface for handling Kafka messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, message *KafkaMessage) error
	GetTopics() []string
}

// KafkaService provides Kafka integration for Go Coffee ecosystem
type KafkaService struct {
	config   *config.Config
	handlers map[string]MessageHandler
	
	// Mock Kafka client (in real implementation, use a proper Kafka client)
	isConnected bool
	mu          sync.RWMutex
	
	// Channels for message processing
	incomingMessages chan *KafkaMessage
	outgoingMessages chan *KafkaMessage
	
	// Metrics and tracing
	tracer trace.Tracer
	
	// Message callbacks
	onMarketData     func(*models.MarketDataUpdate)
	onTradingSignal  func(*models.TradingSignal)
	onPortfolioUpdate func(*models.PortfolioUpdate)
	onAlert          func(*models.AlertNotification)
}

// NewKafkaService creates a new Kafka service
func NewKafkaService(cfg *config.Config) *KafkaService {
	return &KafkaService{
		config:           cfg,
		handlers:         make(map[string]MessageHandler),
		incomingMessages: make(chan *KafkaMessage, 1000),
		outgoingMessages: make(chan *KafkaMessage, 1000),
		tracer:           otel.Tracer("crypto-terminal-kafka"),
	}
}

// Start starts the Kafka service
func (s *KafkaService) Start(ctx context.Context) error {
	logrus.Info("Starting Kafka integration service")

	// Connect to Kafka (mock implementation)
	if err := s.connect(); err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}

	// Start message processing goroutines
	go s.processIncomingMessages(ctx)
	go s.processOutgoingMessages(ctx)
	go s.simulateIncomingMessages(ctx) // Mock message simulation

	logrus.Info("Kafka integration service started")
	return nil
}

// Stop stops the Kafka service
func (s *KafkaService) Stop() error {
	logrus.Info("Stopping Kafka integration service")
	
	s.mu.Lock()
	s.isConnected = false
	s.mu.Unlock()
	
	close(s.incomingMessages)
	close(s.outgoingMessages)
	
	return nil
}

// connect establishes connection to Kafka (mock implementation)
func (s *KafkaService) connect() error {
	// In a real implementation, this would connect to Kafka brokers
	logrus.WithField("brokers", s.config.Integrations.GoCoffee.KafkaBrokers).Info("Connecting to Kafka")
	
	s.mu.Lock()
	s.isConnected = true
	s.mu.Unlock()
	
	return nil
}

// IsConnected returns the connection status
func (s *KafkaService) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isConnected
}

// RegisterHandler registers a message handler for specific topics
func (s *KafkaService) RegisterHandler(handler MessageHandler) {
	for _, topic := range handler.GetTopics() {
		s.handlers[topic] = handler
		logrus.WithField("topic", topic).Info("Registered Kafka message handler")
	}
}

// SetMarketDataCallback sets the callback for market data updates
func (s *KafkaService) SetMarketDataCallback(callback func(*models.MarketDataUpdate)) {
	s.onMarketData = callback
}

// SetTradingSignalCallback sets the callback for trading signals
func (s *KafkaService) SetTradingSignalCallback(callback func(*models.TradingSignal)) {
	s.onTradingSignal = callback
}

// SetPortfolioUpdateCallback sets the callback for portfolio updates
func (s *KafkaService) SetPortfolioUpdateCallback(callback func(*models.PortfolioUpdate)) {
	s.onPortfolioUpdate = callback
}

// SetAlertCallback sets the callback for alert notifications
func (s *KafkaService) SetAlertCallback(callback func(*models.AlertNotification)) {
	s.onAlert = callback
}

// PublishMarketData publishes market data to Kafka
func (s *KafkaService) PublishMarketData(ctx context.Context, data *models.MarketDataUpdate) error {
	return s.publishMessage(ctx, s.config.Integrations.GoCoffee.KafkaTopics.MarketData, data)
}

// PublishTradingSignal publishes a trading signal to Kafka
func (s *KafkaService) PublishTradingSignal(ctx context.Context, signal *models.TradingSignal) error {
	return s.publishMessage(ctx, s.config.Integrations.GoCoffee.KafkaTopics.TradingSignals, signal)
}

// PublishPortfolioUpdate publishes a portfolio update to Kafka
func (s *KafkaService) PublishPortfolioUpdate(ctx context.Context, update *models.PortfolioUpdate) error {
	return s.publishMessage(ctx, s.config.Integrations.GoCoffee.KafkaTopics.PortfolioUpdates, update)
}

// PublishAlert publishes an alert notification to Kafka
func (s *KafkaService) PublishAlert(ctx context.Context, alert *models.AlertNotification) error {
	return s.publishMessage(ctx, s.config.Integrations.GoCoffee.KafkaTopics.Alerts, alert)
}

// publishMessage publishes a message to a Kafka topic
func (s *KafkaService) publishMessage(ctx context.Context, topic string, data interface{}) error {
	ctx, span := s.tracer.Start(ctx, "kafka.publish")
	defer span.End()

	span.SetAttributes(
		attribute.String("kafka.topic", topic),
		attribute.String("message.type", fmt.Sprintf("%T", data)),
	)

	if !s.IsConnected() {
		return fmt.Errorf("Kafka service is not connected")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	message := &KafkaMessage{
		Topic:     topic,
		Value:     jsonData,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"source":      "crypto-terminal",
			"content-type": "application/json",
		},
	}

	select {
	case s.outgoingMessages <- message:
		logrus.WithFields(logrus.Fields{
			"topic": topic,
			"size":  len(jsonData),
		}).Debug("Message queued for publishing")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("outgoing message queue is full")
	}
}

// processIncomingMessages processes incoming Kafka messages
func (s *KafkaService) processIncomingMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-s.incomingMessages:
			s.handleIncomingMessage(ctx, message)
		}
	}
}

// processOutgoingMessages processes outgoing Kafka messages
func (s *KafkaService) processOutgoingMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-s.outgoingMessages:
			s.handleOutgoingMessage(ctx, message)
		}
	}
}

// handleIncomingMessage handles an incoming Kafka message
func (s *KafkaService) handleIncomingMessage(ctx context.Context, message *KafkaMessage) {
	ctx, span := s.tracer.Start(ctx, "kafka.handle_incoming")
	defer span.End()

	span.SetAttributes(
		attribute.String("kafka.topic", message.Topic),
		attribute.String("kafka.key", message.Key),
	)

	// Route message to appropriate handler
	if handler, exists := s.handlers[message.Topic]; exists {
		if err := handler.HandleMessage(ctx, message); err != nil {
			span.RecordError(err)
			logrus.WithFields(logrus.Fields{
				"topic": message.Topic,
				"error": err,
			}).Error("Failed to handle Kafka message")
		}
		return
	}

	// Handle built-in message types
	switch message.Topic {
	case s.config.Integrations.GoCoffee.KafkaTopics.MarketData:
		s.handleMarketDataMessage(ctx, message)
	case s.config.Integrations.GoCoffee.KafkaTopics.TradingSignals:
		s.handleTradingSignalMessage(ctx, message)
	case s.config.Integrations.GoCoffee.KafkaTopics.PortfolioUpdates:
		s.handlePortfolioUpdateMessage(ctx, message)
	case s.config.Integrations.GoCoffee.KafkaTopics.Alerts:
		s.handleAlertMessage(ctx, message)
	default:
		logrus.WithField("topic", message.Topic).Warn("No handler for Kafka topic")
	}
}

// handleOutgoingMessage handles an outgoing Kafka message
func (s *KafkaService) handleOutgoingMessage(ctx context.Context, message *KafkaMessage) {
	ctx, span := s.tracer.Start(ctx, "kafka.send")
	defer span.End()

	span.SetAttributes(
		attribute.String("kafka.topic", message.Topic),
		attribute.Int("message.size", len(message.Value)),
	)

	// In a real implementation, this would send the message to Kafka
	logrus.WithFields(logrus.Fields{
		"topic": message.Topic,
		"size":  len(message.Value),
	}).Debug("Sending message to Kafka")

	// Simulate sending delay
	time.Sleep(10 * time.Millisecond)
}

// handleMarketDataMessage handles market data messages
func (s *KafkaService) handleMarketDataMessage(ctx context.Context, message *KafkaMessage) {
	if s.onMarketData == nil {
		return
	}

	var update models.MarketDataUpdate
	if err := json.Unmarshal(message.Value, &update); err != nil {
		logrus.Errorf("Failed to unmarshal market data message: %v", err)
		return
	}

	s.onMarketData(&update)
}

// handleTradingSignalMessage handles trading signal messages
func (s *KafkaService) handleTradingSignalMessage(ctx context.Context, message *KafkaMessage) {
	if s.onTradingSignal == nil {
		return
	}

	var signal models.TradingSignal
	if err := json.Unmarshal(message.Value, &signal); err != nil {
		logrus.Errorf("Failed to unmarshal trading signal message: %v", err)
		return
	}

	s.onTradingSignal(&signal)
}

// handlePortfolioUpdateMessage handles portfolio update messages
func (s *KafkaService) handlePortfolioUpdateMessage(ctx context.Context, message *KafkaMessage) {
	if s.onPortfolioUpdate == nil {
		return
	}

	var update models.PortfolioUpdate
	if err := json.Unmarshal(message.Value, &update); err != nil {
		logrus.Errorf("Failed to unmarshal portfolio update message: %v", err)
		return
	}

	s.onPortfolioUpdate(&update)
}

// handleAlertMessage handles alert notification messages
func (s *KafkaService) handleAlertMessage(ctx context.Context, message *KafkaMessage) {
	if s.onAlert == nil {
		return
	}

	var alert models.AlertNotification
	if err := json.Unmarshal(message.Value, &alert); err != nil {
		logrus.Errorf("Failed to unmarshal alert message: %v", err)
		return
	}

	s.onAlert(&alert)
}

// simulateIncomingMessages simulates incoming messages for demo purposes
func (s *KafkaService) simulateIncomingMessages(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simulate incoming trading signal
			signal := &models.TradingSignal{
				Symbol:     "BTC",
				Signal:     "BUY",
				Confidence: decimal.NewFromFloat(0.85),
				Source:     "ai-agent",
				Reasoning:  "Strong bullish momentum detected",
				CreatedAt:  time.Now(),
			}

			data, _ := json.Marshal(signal)
			message := &KafkaMessage{
				Topic:     s.config.Integrations.GoCoffee.KafkaTopics.TradingSignals,
				Value:     data,
				Timestamp: time.Now(),
			}

			select {
			case s.incomingMessages <- message:
			default:
				// Queue is full, skip
			}
		}
	}
}
