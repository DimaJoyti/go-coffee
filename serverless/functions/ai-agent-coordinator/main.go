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
	Version    string                 `json:"version"`
	ID         string                 `json:"id"`
	DetailType string                 `json:"detail-type"`
	Source     string                 `json:"source"`
	Account    string                 `json:"account"`
	Time       time.Time              `json:"time"`
	Region     string                 `json:"region"`
	Detail     map[string]interface{} `json:"detail"`
}

// AgentTask represents a task for AI agents
type AgentTask struct {
	TaskID     string                 `json:"task_id"`
	AgentType  string                 `json:"agent_type"`
	TaskType   string                 `json:"task_type"`
	Priority   int                    `json:"priority"`
	Payload    map[string]interface{} `json:"payload"`
	CreatedAt  time.Time              `json:"created_at"`
	DeadlineAt *time.Time             `json:"deadline_at,omitempty"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	Status     string                 `json:"status"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AgentResponse represents a response from an AI agent
type AgentResponse struct {
	TaskID         string                 `json:"task_id"`
	AgentType      string                 `json:"agent_type"`
	Status         string                 `json:"status"`
	Result         map[string]interface{} `json:"result,omitempty"`
	Error          string                 `json:"error,omitempty"`
	ProcessingTime time.Duration          `json:"processing_time"`
	CompletedAt    time.Time              `json:"completed_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CoordinationResult represents the result of task coordination
type CoordinationResult struct {
	TaskID              string                 `json:"task_id"`
	AssignedAgent       string                 `json:"assigned_agent"`
	CoordinationTime    time.Duration          `json:"coordination_time"`
	NextActions         []string               `json:"next_actions"`
	Dependencies        []string               `json:"dependencies"`
	EstimatedCompletion time.Time              `json:"estimated_completion"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// AgentCoordinator handles AI agent task coordination
type AgentCoordinator struct {
	kafkaProducer     *KafkaProducer
	environment       string
	agentCapabilities map[string]AgentCapability
}

// AgentCapability defines what an agent can do
type AgentCapability struct {
	AgentType             string        `json:"agent_type"`
	SupportedTasks        []string      `json:"supported_tasks"`
	MaxConcurrency        int           `json:"max_concurrency"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	Priority              int           `json:"priority"`
	Available             bool          `json:"available"`
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
			Topic:    "agent_tasks",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// PublishAgentTask publishes a task to the appropriate agent
func (kp *KafkaProducer) PublishAgentTask(ctx context.Context, task AgentTask) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal agent task: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(task.TaskID),
		Value: taskBytes,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "agent_type", Value: []byte(task.AgentType)},
			{Key: "task_type", Value: []byte(task.TaskType)},
			{Key: "priority", Value: []byte(fmt.Sprintf("%d", task.Priority))},
		},
	}

	return kp.writer.WriteMessages(ctx, message)
}

// Close closes the Kafka producer
func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}

// NewAgentCoordinator creates a new agent coordinator
func NewAgentCoordinator() *AgentCoordinator {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	coordinator := &AgentCoordinator{
		kafkaProducer:     NewKafkaProducer(kafkaBrokers),
		environment:       os.Getenv("ENVIRONMENT"),
		agentCapabilities: make(map[string]AgentCapability),
	}

	// Initialize agent capabilities
	coordinator.initializeAgentCapabilities()

	return coordinator
}

// initializeAgentCapabilities sets up the capabilities of each agent
func (ac *AgentCoordinator) initializeAgentCapabilities() {
	ac.agentCapabilities = map[string]AgentCapability{
		"beverage-inventor": {
			AgentType:             "beverage-inventor",
			SupportedTasks:        []string{"create_recipe", "analyze_ingredients", "suggest_variations"},
			MaxConcurrency:        3,
			AverageProcessingTime: 45 * time.Second,
			Priority:              1,
			Available:             true,
		},
		"inventory-manager": {
			AgentType:             "inventory-manager",
			SupportedTasks:        []string{"check_inventory", "predict_demand", "order_supplies", "optimize_stock"},
			MaxConcurrency:        5,
			AverageProcessingTime: 15 * time.Second,
			Priority:              2,
			Available:             true,
		},
		"task-manager": {
			AgentType:             "task-manager",
			SupportedTasks:        []string{"create_task", "update_task", "assign_task", "track_progress"},
			MaxConcurrency:        10,
			AverageProcessingTime: 5 * time.Second,
			Priority:              3,
			Available:             true,
		},
		"notifier": {
			AgentType:             "notifier",
			SupportedTasks:        []string{"send_notification", "send_alert", "send_reminder"},
			MaxConcurrency:        20,
			AverageProcessingTime: 2 * time.Second,
			Priority:              4,
			Available:             true,
		},
		"feedback-analyst": {
			AgentType:             "feedback-analyst",
			SupportedTasks:        []string{"analyze_feedback", "sentiment_analysis", "generate_insights"},
			MaxConcurrency:        3,
			AverageProcessingTime: 30 * time.Second,
			Priority:              2,
			Available:             true,
		},
		"scheduler": {
			AgentType:             "scheduler",
			SupportedTasks:        []string{"schedule_task", "optimize_schedule", "handle_conflicts"},
			MaxConcurrency:        5,
			AverageProcessingTime: 10 * time.Second,
			Priority:              3,
			Available:             true,
		},
		"inter-location-coordinator": {
			AgentType:             "inter-location-coordinator",
			SupportedTasks:        []string{"coordinate_locations", "transfer_resources", "sync_inventory"},
			MaxConcurrency:        2,
			AverageProcessingTime: 60 * time.Second,
			Priority:              1,
			Available:             true,
		},
		"tasting-coordinator": {
			AgentType:             "tasting-coordinator",
			SupportedTasks:        []string{"schedule_tasting", "coordinate_participants", "collect_feedback"},
			MaxConcurrency:        2,
			AverageProcessingTime: 90 * time.Second,
			Priority:              1,
			Available:             true,
		},
		"social-media-content": {
			AgentType:             "social-media-content",
			SupportedTasks:        []string{"generate_content", "schedule_posts", "analyze_engagement"},
			MaxConcurrency:        3,
			AverageProcessingTime: 20 * time.Second,
			Priority:              2,
			Available:             true,
		},
	}
}

// CoordinateTask coordinates a task with the appropriate AI agent
func (ac *AgentCoordinator) CoordinateTask(ctx context.Context, task AgentTask) (*CoordinationResult, error) {
	startTime := time.Now()

	log.Printf("Coordinating task: %s of type: %s", task.TaskID, task.TaskType)

	// Find the best agent for this task
	assignedAgent, err := ac.findBestAgent(task)
	if err != nil {
		return nil, fmt.Errorf("failed to find suitable agent: %w", err)
	}

	// Update task with assigned agent
	task.AgentType = assignedAgent
	task.Status = "assigned"

	// Calculate estimated completion time
	capability := ac.agentCapabilities[assignedAgent]
	estimatedCompletion := time.Now().Add(capability.AverageProcessingTime)

	// Determine next actions and dependencies
	nextActions := ac.determineNextActions(task)
	dependencies := ac.analyzeDependencies(task)

	// Publish task to Kafka for the assigned agent
	if err := ac.kafkaProducer.PublishAgentTask(ctx, task); err != nil {
		log.Printf("Failed to publish task to Kafka: %v", err)
		// Don't fail the entire coordination, just log the error
	}

	coordinationTime := time.Since(startTime)

	result := &CoordinationResult{
		TaskID:              task.TaskID,
		AssignedAgent:       assignedAgent,
		CoordinationTime:    coordinationTime,
		NextActions:         nextActions,
		Dependencies:        dependencies,
		EstimatedCompletion: estimatedCompletion,
		Metadata: map[string]interface{}{
			"original_task":    task,
			"coordinator":      "serverless-lambda",
			"environment":      ac.environment,
			"coordinated_at":   time.Now(),
			"agent_capability": capability,
		},
	}

	log.Printf("Successfully coordinated task: %s to agent: %s in %v", task.TaskID, assignedAgent, coordinationTime)
	return result, nil
}

// findBestAgent finds the most suitable agent for a task
func (ac *AgentCoordinator) findBestAgent(task AgentTask) (string, error) {
	var bestAgent string
	var bestScore int

	for agentType, capability := range ac.agentCapabilities {
		if !capability.Available {
			continue
		}

		// Check if agent supports this task type
		supported := false
		for _, supportedTask := range capability.SupportedTasks {
			if supportedTask == task.TaskType {
				supported = true
				break
			}
		}

		if !supported {
			continue
		}

		// Calculate score based on priority, processing time, and concurrency
		score := capability.Priority * 100
		score += int(capability.MaxConcurrency) * 10
		score -= int(capability.AverageProcessingTime.Seconds())

		// Adjust score based on task priority
		score += task.Priority * 50

		if bestAgent == "" || score > bestScore {
			bestAgent = agentType
			bestScore = score
		}
	}

	if bestAgent == "" {
		return "", fmt.Errorf("no suitable agent found for task type: %s", task.TaskType)
	}

	return bestAgent, nil
}

// determineNextActions determines what actions should follow this task
func (ac *AgentCoordinator) determineNextActions(task AgentTask) []string {
	nextActions := []string{}

	switch task.TaskType {
	case "create_recipe":
		nextActions = []string{"schedule_tasting", "update_inventory", "create_task"}
	case "check_inventory":
		nextActions = []string{"order_supplies", "send_notification"}
	case "analyze_feedback":
		nextActions = []string{"generate_content", "create_task"}
	case "schedule_tasting":
		nextActions = []string{"send_notification", "coordinate_participants"}
	case "generate_content":
		nextActions = []string{"schedule_posts", "send_notification"}
	default:
		nextActions = []string{"send_notification"}
	}

	return nextActions
}

// analyzeDependencies analyzes task dependencies
func (ac *AgentCoordinator) analyzeDependencies(task AgentTask) []string {
	dependencies := []string{}

	switch task.TaskType {
	case "create_recipe":
		dependencies = []string{"check_inventory"}
	case "schedule_tasting":
		dependencies = []string{"create_recipe", "coordinate_participants"}
	case "order_supplies":
		dependencies = []string{"check_inventory", "predict_demand"}
	case "generate_content":
		dependencies = []string{"analyze_feedback"}
	}

	return dependencies
}

// Close closes the coordinator resources
func (ac *AgentCoordinator) Close() error {
	return ac.kafkaProducer.Close()
}

// Lambda handler function
func HandleRequest(ctx context.Context, event CloudWatchEvent) (map[string]interface{}, error) {
	log.Printf("Received CloudWatch event: %s", event.DetailType)

	// Parse the agent task from CloudWatch event detail
	var agentTask AgentTask
	detailBytes, err := json.Marshal(event.Detail)
	if err != nil {
		log.Printf("Failed to marshal event detail: %v", err)
		return map[string]interface{}{
			"statusCode": 400,
			"body":       "Invalid event detail format",
		}, nil
	}

	if err := json.Unmarshal(detailBytes, &agentTask); err != nil {
		log.Printf("Failed to unmarshal agent task: %v", err)
		return map[string]interface{}{
			"statusCode": 400,
			"error":      "Invalid event format",
		}, err
	}

	// Create agent coordinator
	coordinator := NewAgentCoordinator()
	defer coordinator.Close()

	// Coordinate the task
	result, err := coordinator.CoordinateTask(ctx, agentTask)
	if err != nil {
		log.Printf("Failed to coordinate task: %v", err)
		return map[string]interface{}{
			"statusCode": 500,
			"error":      err.Error(),
		}, err
	}

	// Return success response
	return map[string]interface{}{
		"statusCode":         200,
		"coordinationResult": result,
		"message":            "Task coordinated successfully",
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
			DetailType: "AI Agent Coordination Request",
			Source:     "go-coffee.ai-coordinator",
			Account:    "123456789012",
			Time:       time.Now(),
			Region:     "us-east-1",
			Detail: map[string]interface{}{
				"task_type": "order_optimization",
				"priority":  "high",
			},
		}

		result, err := HandleRequest(ctx, event)
		if err != nil {
			log.Printf("Handler error: %v", err)
		} else {
			log.Printf("Handler result: %+v", result)
		}
	} else {
		// For local testing or other cloud providers
		log.Println("AI Agent Coordinator started")

		// Example task for testing
		testTask := AgentTask{
			TaskID:    "test-task-123",
			AgentType: "", // Will be determined by coordinator
			TaskType:  "create_recipe",
			Priority:  1,
			Payload: map[string]interface{}{
				"ingredients": []string{"espresso", "steamed milk", "vanilla syrup"},
				"theme":       "autumn",
				"location":    "downtown-shop",
			},
			CreatedAt:  time.Now(),
			RetryCount: 0,
			MaxRetries: 3,
			Status:     "pending",
		}

		coordinator := NewAgentCoordinator()
		defer coordinator.Close()

		result, err := coordinator.CoordinateTask(context.Background(), testTask)
		if err != nil {
			log.Fatalf("Failed to coordinate test task: %v", err)
		}

		log.Printf("Test task coordinated successfully: %+v", result)
	}
}
