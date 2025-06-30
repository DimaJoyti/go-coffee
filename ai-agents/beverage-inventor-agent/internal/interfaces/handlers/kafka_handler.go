package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/application/usecases"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/config"
	kafkaInfra "go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/kafka"
)

// KafkaHandler handles Kafka messages for the beverage inventor agent
type KafkaHandler struct {
	beverageUseCase *usecases.BeverageInventorUseCase
	processor       *kafkaInfra.MessageProcessor
	logger          Logger
}

// Logger interface for the handler
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewKafkaHandler creates a new Kafka handler
func NewKafkaHandler(beverageUseCase *usecases.BeverageInventorUseCase, logger Logger) *KafkaHandler {
	return &KafkaHandler{
		beverageUseCase: beverageUseCase,
		processor:       kafkaInfra.NewMessageProcessor(),
		logger:          logger,
	}
}

// StartConsumer starts consuming messages from Kafka topics
func (h *KafkaHandler) StartConsumer(ctx context.Context, cfg config.KafkaConfig) error {
	// Define topics to consume
	topics := []string{
		cfg.Topics.RecipeRequests,
		cfg.Topics.IngredientDiscovered,
	}

	// Create consumer
	consumer, err := kafkaInfra.NewConsumer(cfg, topics)
	if err != nil {
		return fmt.Errorf("failed to create Kafka consumer: %w", err)
	}
	defer consumer.Close()

	h.logger.Info("Starting Kafka consumer", "topics", topics, "group_id", cfg.Consumer.GroupID)

	// Start consuming messages
	return consumer.Start(ctx, h)
}

// HandleMessage implements the MessageHandler interface
func (h *KafkaHandler) HandleMessage(ctx context.Context, message *kafka.Message) error {
	h.logger.Debug("Received Kafka message", 
		"topic", message.Topic, 
		"partition", message.Partition, 
		"offset", message.Offset,
		"key", string(message.Key))

	// Get message metadata
	metadata := h.processor.GetMessageMetadata(message)
	
	// Route message based on topic
	switch message.Topic {
	case "recipe.requests":
		return h.handleRecipeRequest(ctx, message, metadata)
	case "ingredient.discovered":
		return h.handleIngredientDiscovered(ctx, message, metadata)
	default:
		h.logger.Warn("Unknown topic", "topic", message.Topic)
		return nil // Don't fail for unknown topics
	}
}

// handleRecipeRequest handles recipe request messages
func (h *KafkaHandler) handleRecipeRequest(ctx context.Context, message *kafka.Message, metadata map[string]string) error {
	h.logger.Info("Processing recipe request", "key", string(message.Key))

	// Parse the recipe request event
	event, err := h.processor.ProcessRecipeRequest(message)
	if err != nil {
		h.logger.Error("Failed to process recipe request", err, "key", string(message.Key))
		return err
	}

	// Convert to use case request
	request := &usecases.InventBeverageRequest{
		Ingredients: event.Ingredients,
		Theme:       event.Theme,
		UseAI:       event.UseAI,
		CreatedBy:   event.RequestedBy,
	}

	// Apply constraints if present
	if event.Constraints != nil {
		request.Constraints = h.convertConstraints(event.Constraints)
	}

	// Execute the use case
	response, err := h.beverageUseCase.InventBeverage(ctx, request)
	if err != nil {
		h.logger.Error("Failed to invent beverage", err, 
			"request_id", event.RequestID,
			"ingredients", event.Ingredients,
			"theme", event.Theme)
		return err
	}

	h.logger.Info("Successfully invented beverage", 
		"request_id", event.RequestID,
		"beverage_id", response.Beverage.ID,
		"beverage_name", response.Beverage.Name,
		"task_created", response.TaskCreated,
		"ai_used", response.AIUsed)

	// Log warnings if any
	if len(response.Warnings) > 0 {
		h.logger.Warn("Beverage invention completed with warnings", 
			"warnings", response.Warnings,
			"beverage_id", response.Beverage.ID)
	}

	return nil
}

// handleIngredientDiscovered handles ingredient discovered messages
func (h *KafkaHandler) handleIngredientDiscovered(ctx context.Context, message *kafka.Message, metadata map[string]string) error {
	h.logger.Info("Processing ingredient discovered event", "key", string(message.Key))

	// Parse the ingredient discovered event
	event, err := h.processor.ProcessIngredientDiscovered(message)
	if err != nil {
		h.logger.Error("Failed to process ingredient discovered event", err, "key", string(message.Key))
		return err
	}

	h.logger.Info("New ingredient discovered", 
		"ingredient_id", event.IngredientID,
		"name", event.Name,
		"source", event.Source,
		"category", event.Category)

	// Create a beverage idea with the new ingredient
	// Use a default theme or derive from source
	theme := h.deriveThemeFromSource(event.Source)
	
	request := &usecases.InventBeverageRequest{
		Ingredients: []string{event.Name},
		Theme:       theme,
		UseAI:       true, // Use AI for new ingredient exploration
		CreatedBy:   "system", // System-generated
	}

	// Execute the use case
	response, err := h.beverageUseCase.InventBeverage(ctx, request)
	if err != nil {
		h.logger.Error("Failed to create beverage for new ingredient", err, 
			"ingredient_id", event.IngredientID,
			"ingredient_name", event.Name)
		return err
	}

	h.logger.Info("Successfully created beverage for new ingredient", 
		"ingredient_id", event.IngredientID,
		"ingredient_name", event.Name,
		"beverage_id", response.Beverage.ID,
		"beverage_name", response.Beverage.Name)

	return nil
}

// convertConstraints converts event constraints to use case constraints
func (h *KafkaHandler) convertConstraints(eventConstraints map[string]interface{}) usecases.BeverageConstraints {
	constraints := usecases.BeverageConstraints{}

	if maxCost, ok := eventConstraints["max_cost"].(float64); ok {
		constraints.MaxCost = &maxCost
	}

	if maxCalories, ok := eventConstraints["max_calories"].(float64); ok {
		calories := int(maxCalories)
		constraints.MaxCalories = &calories
	}

	if maxPrepTime, ok := eventConstraints["max_prep_time"].(float64); ok {
		prepTime := int(maxPrepTime)
		constraints.MaxPrepTime = &prepTime
	}

	if requiredTags, ok := eventConstraints["required_tags"].([]interface{}); ok {
		tags := make([]string, len(requiredTags))
		for i, tag := range requiredTags {
			if tagStr, ok := tag.(string); ok {
				tags[i] = tagStr
			}
		}
		constraints.RequiredTags = tags
	}

	if forbiddenTags, ok := eventConstraints["forbidden_tags"].([]interface{}); ok {
		tags := make([]string, len(forbiddenTags))
		for i, tag := range forbiddenTags {
			if tagStr, ok := tag.(string); ok {
				tags[i] = tagStr
			}
		}
		constraints.ForbiddenTags = tags
	}

	if allergenFree, ok := eventConstraints["allergen_free"].([]interface{}); ok {
		allergens := make([]string, len(allergenFree))
		for i, allergen := range allergenFree {
			if allergenStr, ok := allergen.(string); ok {
				allergens[i] = allergenStr
			}
		}
		constraints.AllergenFree = allergens
	}

	return constraints
}

// deriveThemeFromSource derives a theme from the ingredient source
func (h *KafkaHandler) deriveThemeFromSource(source string) string {
	switch source {
	case "Mars Base", "Martian Greenhouse":
		return "Mars Base"
	case "Lunar Mining Corp", "Moon Base Alpha":
		return "Lunar Mining Corp"
	case "Interstellar Trade Federation", "Deep Space Station":
		return "Interstellar Trade Federation"
	case "Earth Café", "Local Farm", "Organic Supplier":
		return "Earth Café"
	default:
		return "Earth Café" // Default theme
	}
}

// HTTPHandler handles HTTP requests for the beverage inventor agent
type HTTPHandler struct {
	beverageUseCase *usecases.BeverageInventorUseCase
	logger          Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(beverageUseCase *usecases.BeverageInventorUseCase, logger Logger) *HTTPHandler {
	return &HTTPHandler{
		beverageUseCase: beverageUseCase,
		logger:          logger,
	}
}

// InventBeverageRequest represents an HTTP request to invent a beverage
type InventBeverageRequest struct {
	Ingredients    []string                   `json:"ingredients"`
	Theme          string                     `json:"theme"`
	UseAI          bool                       `json:"use_ai"`
	CreatedBy      string                     `json:"created_by"`
	TargetAudience []string                   `json:"target_audience,omitempty"`
	Constraints    map[string]interface{}     `json:"constraints,omitempty"`
}

// InventBeverageResponse represents an HTTP response for beverage invention
type InventBeverageResponse struct {
	Success     bool                       `json:"success"`
	BeverageID  string                     `json:"beverage_id,omitempty"`
	Name        string                     `json:"name,omitempty"`
	Description string                     `json:"description,omitempty"`
	TaskCreated bool                       `json:"task_created"`
	TaskID      string                     `json:"task_id,omitempty"`
	AIUsed      bool                       `json:"ai_used"`
	Warnings    []string                   `json:"warnings,omitempty"`
	Error       string                     `json:"error,omitempty"`
}

// InventBeverage handles HTTP requests to invent a beverage
func (h *HTTPHandler) InventBeverage(ctx context.Context, requestBody []byte) (*InventBeverageResponse, error) {
	var httpRequest InventBeverageRequest
	if err := json.Unmarshal(requestBody, &httpRequest); err != nil {
		return &InventBeverageResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request format: %v", err),
		}, nil
	}

	// Convert to use case request
	useCaseRequest := &usecases.InventBeverageRequest{
		Ingredients:    httpRequest.Ingredients,
		Theme:          httpRequest.Theme,
		UseAI:          httpRequest.UseAI,
		CreatedBy:      httpRequest.CreatedBy,
		TargetAudience: httpRequest.TargetAudience,
	}

	// Convert constraints if present
	if httpRequest.Constraints != nil {
		useCaseRequest.Constraints = h.convertHTTPConstraints(httpRequest.Constraints)
	}

	// Execute the use case
	response, err := h.beverageUseCase.InventBeverage(ctx, useCaseRequest)
	if err != nil {
		h.logger.Error("Failed to invent beverage via HTTP", err, 
			"ingredients", httpRequest.Ingredients,
			"theme", httpRequest.Theme)
		return &InventBeverageResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &InventBeverageResponse{
		Success:     true,
		BeverageID:  response.Beverage.ID.String(),
		Name:        response.Beverage.Name,
		Description: response.Beverage.Description,
		TaskCreated: response.TaskCreated,
		TaskID:      response.TaskID,
		AIUsed:      response.AIUsed,
		Warnings:    response.Warnings,
	}, nil
}

// convertHTTPConstraints converts HTTP constraints to use case constraints
func (h *HTTPHandler) convertHTTPConstraints(httpConstraints map[string]interface{}) usecases.BeverageConstraints {
	constraints := usecases.BeverageConstraints{}

	if maxCost, ok := httpConstraints["max_cost"].(float64); ok {
		constraints.MaxCost = &maxCost
	}

	if maxCalories, ok := httpConstraints["max_calories"].(float64); ok {
		calories := int(maxCalories)
		constraints.MaxCalories = &calories
	}

	if maxPrepTime, ok := httpConstraints["max_prep_time"].(float64); ok {
		prepTime := int(maxPrepTime)
		constraints.MaxPrepTime = &prepTime
	}

	if requiredTags, ok := httpConstraints["required_tags"].([]interface{}); ok {
		tags := make([]string, len(requiredTags))
		for i, tag := range requiredTags {
			if tagStr, ok := tag.(string); ok {
				tags[i] = tagStr
			}
		}
		constraints.RequiredTags = tags
	}

	if forbiddenTags, ok := httpConstraints["forbidden_tags"].([]interface{}); ok {
		tags := make([]string, len(forbiddenTags))
		for i, tag := range forbiddenTags {
			if tagStr, ok := tag.(string); ok {
				tags[i] = tagStr
			}
		}
		constraints.ForbiddenTags = tags
	}

	if allergenFree, ok := httpConstraints["allergen_free"].([]interface{}); ok {
		allergens := make([]string, len(allergenFree))
		for i, allergen := range allergenFree {
			if allergenStr, ok := allergen.(string); ok {
				allergens[i] = allergenStr
			}
		}
		constraints.AllergenFree = allergens
	}

	return constraints
}
