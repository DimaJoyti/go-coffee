package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/config"
)

// Consumer wraps Kafka consumer functionality
type Consumer struct {
	reader *kafka.Reader
	config config.KafkaConfig
}

// MessageHandler defines the interface for handling Kafka messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, message *kafka.Message) error
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg config.KafkaConfig, topics []string) (*Consumer, error) {
	if len(topics) == 0 {
		return nil, fmt.Errorf("at least one topic must be specified")
	}

	// Note: kafka-go ReaderConfig uses Topic (singular), not Topics
	// For multiple topics, we'll use the first topic as primary
	topic := topics[0]
	if len(topics) > 1 {
		// Log warning about multiple topics - may need separate readers
		fmt.Printf("Warning: kafka-go Reader supports single topic, using: %s\n", topic)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          topic,
		GroupID:        cfg.Consumer.GroupID,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			fmt.Printf("Kafka consumer error: "+msg+"\n", args...)
		}),
	})

	return &Consumer{
		reader: reader,
		config: cfg,
	}, nil
}

// Start starts consuming messages and calls the handler for each message
func (c *Consumer) Start(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				fmt.Printf("Error reading message: %v\n", err)
				continue
			}

			// Handle the message
			if err := handler.HandleMessage(ctx, &message); err != nil {
				fmt.Printf("Error handling message: %v\n", err)
				// Continue processing other messages even if one fails
				continue
			}

			// Commit the message
			if err := c.reader.CommitMessages(ctx, message); err != nil {
				fmt.Printf("Error committing message: %v\n", err)
			}
		}
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// MessageProcessor provides utilities for processing different message types
type MessageProcessor struct{}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{}
}

// ProcessRecipeRequest processes a recipe request message
func (mp *MessageProcessor) ProcessRecipeRequest(message *kafka.Message) (*RecipeRequestEvent, error) {
	var event RecipeRequestEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipe request event: %w", err)
	}

	// Validate the event
	if err := mp.validateRecipeRequestEvent(&event); err != nil {
		return nil, fmt.Errorf("invalid recipe request event: %w", err)
	}

	return &event, nil
}

// ProcessIngredientDiscovered processes an ingredient discovered message
func (mp *MessageProcessor) ProcessIngredientDiscovered(message *kafka.Message) (*IngredientDiscoveredEvent, error) {
	var event IngredientDiscoveredEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ingredient discovered event: %w", err)
	}

	// Validate the event
	if err := mp.validateIngredientDiscoveredEvent(&event); err != nil {
		return nil, fmt.Errorf("invalid ingredient discovered event: %w", err)
	}

	return &event, nil
}

// validateRecipeRequestEvent validates a recipe request event
func (mp *MessageProcessor) validateRecipeRequestEvent(event *RecipeRequestEvent) error {
	if event.RequestID == "" {
		return fmt.Errorf("request ID is required")
	}
	if len(event.Ingredients) == 0 {
		return fmt.Errorf("at least one ingredient is required")
	}
	if event.Theme == "" {
		return fmt.Errorf("theme is required")
	}
	if event.RequestedBy == "" {
		return fmt.Errorf("requested by is required")
	}
	return nil
}

// validateIngredientDiscoveredEvent validates an ingredient discovered event
func (mp *MessageProcessor) validateIngredientDiscoveredEvent(event *IngredientDiscoveredEvent) error {
	if event.IngredientID == "" {
		return fmt.Errorf("ingredient ID is required")
	}
	if event.Name == "" {
		return fmt.Errorf("ingredient name is required")
	}
	if event.Source == "" {
		return fmt.Errorf("ingredient source is required")
	}
	return nil
}

// GetMessageMetadata extracts metadata from a Kafka message
func (mp *MessageProcessor) GetMessageMetadata(message *kafka.Message) map[string]string {
	metadata := make(map[string]string)
	
	// Extract headers
	for _, header := range message.Headers {
		metadata[header.Key] = string(header.Value)
	}
	
	// Add message metadata
	metadata["topic"] = message.Topic
	metadata["partition"] = fmt.Sprintf("%d", message.Partition)
	metadata["offset"] = fmt.Sprintf("%d", message.Offset)
	metadata["key"] = string(message.Key)
	
	return metadata
}

// Additional event types for consumption

// IngredientDiscoveredEvent represents an ingredient discovery event
type IngredientDiscoveredEvent struct {
	IngredientID   string                 `json:"ingredient_id"`
	Name           string                 `json:"name"`
	Source         string                 `json:"source"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	Availability   string                 `json:"availability"`
	Cost           float64                `json:"cost"`
	Nutritional    map[string]interface{} `json:"nutritional"`
	DiscoveredBy   string                 `json:"discovered_by"`
	DiscoveredAt   time.Time              `json:"discovered_at"`
	Properties     map[string]interface{} `json:"properties"`
	EventType      string                 `json:"event_type"`
	Version        string                 `json:"version"`
}

// FeedbackReceivedEvent represents feedback on a beverage
type FeedbackReceivedEvent struct {
	FeedbackID     string                 `json:"feedback_id"`
	BeverageID     string                 `json:"beverage_id"`
	CustomerID     string                 `json:"customer_id"`
	Rating         int                    `json:"rating"`
	Comments       string                 `json:"comments"`
	Sentiment      string                 `json:"sentiment"`
	Categories     []string               `json:"categories"`
	ReceivedAt     time.Time              `json:"received_at"`
	Source         string                 `json:"source"`
	Metadata       map[string]interface{} `json:"metadata"`
	EventType      string                 `json:"event_type"`
	Version        string                 `json:"version"`
}

// InventoryUpdateEvent represents an inventory update
type InventoryUpdateEvent struct {
	InventoryID    string                 `json:"inventory_id"`
	IngredientID   string                 `json:"ingredient_id"`
	IngredientName string                 `json:"ingredient_name"`
	Location       string                 `json:"location"`
	Quantity       float64                `json:"quantity"`
	Unit           string                 `json:"unit"`
	PreviousQty    float64                `json:"previous_quantity"`
	ChangeType     string                 `json:"change_type"` // addition, consumption, adjustment
	UpdatedBy      string                 `json:"updated_by"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Reason         string                 `json:"reason"`
	Metadata       map[string]interface{} `json:"metadata"`
	EventType      string                 `json:"event_type"`
	Version        string                 `json:"version"`
}
