package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/config"
)

// Producer wraps Kafka producer functionality
type Producer struct {
	writer *kafka.Writer
	config config.KafkaConfig
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg config.KafkaConfig) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: time.Duration(cfg.Producer.BatchTimeout) * time.Millisecond,
		RequiredAcks: kafka.RequiredAcks(cfg.Producer.RequiredAcks),
		MaxAttempts:  cfg.Producer.RetryMax,
		ErrorLogger:  kafka.LoggerFunc(func(msg string, args ...interface{}) {
			fmt.Printf("Kafka producer error: "+msg+"\n", args...)
		}),
	}

	return &Producer{
		writer: writer,
		config: cfg,
	}, nil
}

// PublishMessage publishes a message to the specified topic
func (p *Producer) PublishMessage(ctx context.Context, topic string, key string, value interface{}) error {
	// Serialize the value to JSON
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message value: %w", err)
	}

	message := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: valueBytes,
		Headers: []kafka.Header{
			{
				Key:   "content-type",
				Value: []byte("application/json"),
			},
			{
				Key:   "timestamp",
				Value: []byte(time.Now().Format(time.RFC3339)),
			},
			{
				Key:   "producer",
				Value: []byte("beverage-inventor-agent"),
			},
		},
	}

	return p.writer.WriteMessages(ctx, message)
}

// PublishBeverageCreated publishes a beverage created event
func (p *Producer) PublishBeverageCreated(ctx context.Context, event *BeverageCreatedEvent) error {
	return p.PublishMessage(ctx, p.config.Topics.BeverageCreated, event.BeverageID, event)
}

// PublishBeverageUpdated publishes a beverage updated event
func (p *Producer) PublishBeverageUpdated(ctx context.Context, event *BeverageUpdatedEvent) error {
	return p.PublishMessage(ctx, p.config.Topics.BeverageUpdated, event.BeverageID, event)
}

// PublishTaskCreated publishes a task created event
func (p *Producer) PublishTaskCreated(ctx context.Context, event *TaskCreatedEvent) error {
	return p.PublishMessage(ctx, p.config.Topics.TaskCreated, event.TaskID, event)
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Event types for Kafka messages

// BeverageCreatedEvent represents a beverage creation event
type BeverageCreatedEvent struct {
	BeverageID   string                 `json:"beverage_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Theme        string                 `json:"theme"`
	Ingredients  []IngredientEvent      `json:"ingredients"`
	CreatedBy    string                 `json:"created_by"`
	CreatedAt    time.Time              `json:"created_at"`
	EstimatedCost float64               `json:"estimated_cost"`
	Metadata     map[string]interface{} `json:"metadata"`
	EventType    string                 `json:"event_type"`
	Version      string                 `json:"version"`
}

// BeverageUpdatedEvent represents a beverage update event
type BeverageUpdatedEvent struct {
	BeverageID   string                 `json:"beverage_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"`
	UpdatedBy    string                 `json:"updated_by"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Changes      map[string]interface{} `json:"changes"`
	EventType    string                 `json:"event_type"`
	Version      string                 `json:"version"`
}

// TaskCreatedEvent represents a task creation event
type TaskCreatedEvent struct {
	TaskID       string                 `json:"task_id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Priority     string                 `json:"priority"`
	Assignee     string                 `json:"assignee"`
	Tags         []string               `json:"tags"`
	BeverageID   string                 `json:"beverage_id,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	EventType    string                 `json:"event_type"`
	Version      string                 `json:"version"`
}

// IngredientEvent represents an ingredient in events
type IngredientEvent struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	Source   string  `json:"source"`
	Cost     float64 `json:"cost"`
}

// RecipeRequestEvent represents a recipe request event
type RecipeRequestEvent struct {
	RequestID    string    `json:"request_id"`
	Ingredients  []string  `json:"ingredients"`
	Theme        string    `json:"theme"`
	RequestedBy  string    `json:"requested_by"`
	RequestedAt  time.Time `json:"requested_at"`
	UseAI        bool      `json:"use_ai"`
	Constraints  map[string]interface{} `json:"constraints,omitempty"`
	EventType    string    `json:"event_type"`
	Version      string    `json:"version"`
}
