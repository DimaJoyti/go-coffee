package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

// CloudWatchEvent represents a CloudWatch event (simplified version)
type CloudWatchEvent struct {
	Version    string          `json:"version"`
	ID         string          `json:"id"`
	DetailType string          `json:"detail-type"`
	Source     string          `json:"source"`
	Account    string          `json:"account"`
	Time       time.Time       `json:"time"`
	Region     string          `json:"region"`
	Detail     json.RawMessage `json:"detail"`
}

// OrderEvent represents a coffee order event
type OrderEvent struct {
	OrderID      string                 `json:"order_id"`
	CustomerName string                 `json:"customer_name"`
	CoffeeType   string                 `json:"coffee_type"`
	Quantity     int                    `json:"quantity"`
	Price        float64                `json:"price"`
	Status       string                 `json:"status"`
	Timestamp    time.Time              `json:"timestamp"`
	LocationID   string                 `json:"location_id"`
	PaymentType  string                 `json:"payment_type"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ProcessedOrder represents the processed order result
type ProcessedOrder struct {
	OrderID            string                 `json:"order_id"`
	Status             string                 `json:"status"`
	ProcessingTime     time.Duration          `json:"processing_time"`
	EstimatedReadyTime time.Time              `json:"estimated_ready_time"`
	QueuePosition      int                    `json:"queue_position"`
	Notifications      []NotificationTarget   `json:"notifications"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationTarget represents a notification destination
type NotificationTarget struct {
	Type    string `json:"type"`   // email, sms, push, slack
	Target  string `json:"target"` // email address, phone number, etc.
	Message string `json:"message"`
}

// KafkaProducer handles Kafka message publishing
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers),
			Topic:    "processed_orders",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// PublishProcessedOrder publishes a processed order to Kafka
func (kp *KafkaProducer) PublishProcessedOrder(ctx context.Context, order ProcessedOrder) error {
	orderBytes, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal processed order: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(order.OrderID),
		Value: orderBytes,
		Time:  time.Now(),
	}

	return kp.writer.WriteMessages(ctx, message)
}

// Close closes the Kafka producer
func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}

// OrderProcessor handles order processing logic
type OrderProcessor struct {
	kafkaProducer *KafkaProducer
	environment   string
}

// NewOrderProcessor creates a new order processor
func NewOrderProcessor() *OrderProcessor {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	return &OrderProcessor{
		kafkaProducer: NewKafkaProducer(kafkaBrokers),
		environment:   os.Getenv("ENVIRONMENT"),
	}
}

// ProcessOrder processes a coffee order
func (op *OrderProcessor) ProcessOrder(ctx context.Context, order OrderEvent) (*ProcessedOrder, error) {
	startTime := time.Now()

	log.Printf("Processing order: %s for customer: %s", order.OrderID, order.CustomerName)

	// Validate order
	if err := op.validateOrder(order); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// Calculate processing metrics
	processingTime := time.Since(startTime)
	estimatedReadyTime := op.calculateReadyTime(order)
	queuePosition := op.calculateQueuePosition(order.LocationID)

	// Create processed order
	processedOrder := &ProcessedOrder{
		OrderID:            order.OrderID,
		Status:             "processing",
		ProcessingTime:     processingTime,
		EstimatedReadyTime: estimatedReadyTime,
		QueuePosition:      queuePosition,
		Notifications:      op.generateNotifications(order),
		Metadata: map[string]interface{}{
			"original_order": order,
			"processor":      "serverless-lambda",
			"environment":    op.environment,
			"processed_at":   time.Now(),
		},
	}

	// Publish to Kafka
	if err := op.kafkaProducer.PublishProcessedOrder(ctx, *processedOrder); err != nil {
		log.Printf("Failed to publish processed order to Kafka: %v", err)
		// Don't fail the entire process, just log the error
	}

	log.Printf("Successfully processed order: %s in %v", order.OrderID, processingTime)
	return processedOrder, nil
}

// validateOrder validates the incoming order
func (op *OrderProcessor) validateOrder(order OrderEvent) error {
	if order.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}
	if order.CustomerName == "" {
		return fmt.Errorf("customer name is required")
	}
	if order.CoffeeType == "" {
		return fmt.Errorf("coffee type is required")
	}
	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	if order.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	return nil
}

// calculateReadyTime estimates when the order will be ready
func (op *OrderProcessor) calculateReadyTime(order OrderEvent) time.Time {
	// Base preparation time based on coffee type
	baseTime := map[string]int{
		"espresso":    2, // 2 minutes
		"americano":   3, // 3 minutes
		"latte":       5, // 5 minutes
		"cappuccino":  5, // 5 minutes
		"macchiato":   4, // 4 minutes
		"mocha":       6, // 6 minutes
		"frappuccino": 8, // 8 minutes
	}

	prepTime, exists := baseTime[order.CoffeeType]
	if !exists {
		prepTime = 5 // Default 5 minutes
	}

	// Add time based on quantity
	totalTime := prepTime + (order.Quantity-1)*2

	return time.Now().Add(time.Duration(totalTime) * time.Minute)
}

// calculateQueuePosition estimates the queue position
func (op *OrderProcessor) calculateQueuePosition(locationID string) int {
	// In a real implementation, this would query the current queue
	// For now, return a simulated position
	return 3
}

// generateNotifications creates notification targets for the order
func (op *OrderProcessor) generateNotifications(order OrderEvent) []NotificationTarget {
	notifications := []NotificationTarget{
		{
			Type:    "push",
			Target:  order.CustomerName,
			Message: fmt.Sprintf("Your %s order is being prepared! Estimated ready time: %v", order.CoffeeType, time.Now().Add(5*time.Minute).Format("15:04")),
		},
	}

	// Add SMS notification for orders over $10
	if order.Price > 10.0 {
		notifications = append(notifications, NotificationTarget{
			Type:    "sms",
			Target:  "customer-phone", // Would be actual phone number
			Message: fmt.Sprintf("Premium order confirmed! Your %s will be ready soon.", order.CoffeeType),
		})
	}

	return notifications
}

// Close closes the order processor resources
func (op *OrderProcessor) Close() error {
	return op.kafkaProducer.Close()
}

// Lambda handler function
func HandleRequest(ctx context.Context, event CloudWatchEvent) (map[string]interface{}, error) {
	log.Printf("Received CloudWatch event: %s", event.DetailType)

	// Parse the order event from CloudWatch event detail
	var orderEvent OrderEvent
	if err := json.Unmarshal(event.Detail, &orderEvent); err != nil {
		log.Printf("Failed to unmarshal order event: %v", err)
		return map[string]interface{}{
			"statusCode": 400,
			"error":      "Invalid event format",
		}, err
	}

	// Create order processor
	processor := NewOrderProcessor()
	defer processor.Close()

	// Process the order
	processedOrder, err := processor.ProcessOrder(ctx, orderEvent)
	if err != nil {
		log.Printf("Failed to process order: %v", err)
		return map[string]interface{}{
			"statusCode": 500,
			"error":      err.Error(),
		}, err
	}

	// Return success response
	return map[string]interface{}{
		"statusCode":     200,
		"processedOrder": processedOrder,
		"message":        "Order processed successfully",
	}, nil
}

// Handler for Google Cloud Functions
func Handler(ctx context.Context, m map[string]interface{}) (map[string]interface{}, error) {
	// Convert the generic map to CloudWatch event format
	eventBytes, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	var event CloudWatchEvent
	if err := json.Unmarshal(eventBytes, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	return HandleRequest(ctx, event)
}

func main() {
	// Check if running in AWS Lambda environment
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// In a real AWS Lambda environment, you would use lambda.Start(HandleRequest)
		// For now, we'll simulate the Lambda runtime
		log.Println("AWS Lambda environment detected - would start Lambda handler")

		// Simulate Lambda execution for testing
		ctx := context.Background()
		event := CloudWatchEvent{
			Version:    "0",
			ID:         "test-event",
			DetailType: "Coffee Order Processing Request",
			Source:     "go-coffee.order-processor",
			Account:    "123456789012",
			Time:       time.Now(),
			Region:     "us-east-1",
			Detail:     json.RawMessage(`{"order_id":"test-order-123","customer_name":"John Doe","coffee_type":"latte"}`),
		}

		result, err := HandleRequest(ctx, event)
		if err != nil {
			log.Printf("Handler error: %v", err)
		} else {
			log.Printf("Handler result: %+v", result)
		}
	} else {
		// For local testing or other cloud providers
		log.Println("Coffee Order Processor started")

		// Example order for testing
		testOrder := OrderEvent{
			OrderID:      "test-order-123",
			CustomerName: "John Doe",
			CoffeeType:   "latte",
			Quantity:     2,
			Price:        8.50,
			Status:       "pending",
			Timestamp:    time.Now(),
			LocationID:   "location-1",
			PaymentType:  "credit_card",
		}

		processor := NewOrderProcessor()
		defer processor.Close()

		result, err := processor.ProcessOrder(context.Background(), testOrder)
		if err != nil {
			log.Fatalf("Failed to process test order: %v", err)
		}

		log.Printf("Test order processed successfully: %+v", result)
	}
}
