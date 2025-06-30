package kafka

import (
	"context"
	"time"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/repositories"
)

// EventPublisherAdapter adapts the Kafka producer to implement the EventPublisher interface
type EventPublisherAdapter struct {
	producer *Producer
}

// NewEventPublisherAdapter creates a new event publisher adapter
func NewEventPublisherAdapter(producer *Producer) repositories.EventPublisher {
	return &EventPublisherAdapter{
		producer: producer,
	}
}

// PublishBeverageCreated publishes a beverage created event
func (e *EventPublisherAdapter) PublishBeverageCreated(ctx context.Context, beverage *entities.Beverage) error {
	event := &BeverageCreatedEvent{
		BeverageID:    beverage.ID.String(),
		Name:          beverage.Name,
		Description:   beverage.Description,
		Theme:         beverage.Theme,
		Ingredients:   e.convertIngredients(beverage.Ingredients),
		CreatedBy:     beverage.CreatedBy,
		CreatedAt:     beverage.CreatedAt,
		EstimatedCost: beverage.Metadata.EstimatedCost,
		Metadata:      e.convertMetadata(beverage.Metadata),
		EventType:     "beverage.created",
		Version:       "1.0",
	}

	return e.producer.PublishBeverageCreated(ctx, event)
}

// PublishBeverageUpdated publishes a beverage updated event
func (e *EventPublisherAdapter) PublishBeverageUpdated(ctx context.Context, beverage *entities.Beverage) error {
	event := &BeverageUpdatedEvent{
		BeverageID:  beverage.ID.String(),
		Name:        beverage.Name,
		Description: beverage.Description,
		Status:      string(beverage.Status),
		UpdatedBy:   beverage.CreatedBy, // Using CreatedBy as UpdatedBy for now
		UpdatedAt:   time.Now(),
		Changes:     map[string]interface{}{}, // Would need to track changes
		EventType:   "beverage.updated",
		Version:     "1.0",
	}

	return e.producer.PublishBeverageUpdated(ctx, event)
}

// PublishBeverageStatusChanged publishes a beverage status change event
func (e *EventPublisherAdapter) PublishBeverageStatusChanged(ctx context.Context, beverage *entities.Beverage, oldStatus entities.BeverageStatus) error {
	event := &BeverageUpdatedEvent{
		BeverageID:  beverage.ID.String(),
		Name:        beverage.Name,
		Description: beverage.Description,
		Status:      string(beverage.Status),
		UpdatedBy:   beverage.CreatedBy,
		UpdatedAt:   time.Now(),
		Changes: map[string]interface{}{
			"status": map[string]interface{}{
				"old": string(oldStatus),
				"new": string(beverage.Status),
			},
		},
		EventType: "beverage.status_changed",
		Version:   "1.0",
	}

	return e.producer.PublishBeverageUpdated(ctx, event)
}

// convertIngredients converts domain ingredients to event ingredients
func (e *EventPublisherAdapter) convertIngredients(ingredients []entities.Ingredient) []IngredientEvent {
	eventIngredients := make([]IngredientEvent, len(ingredients))
	for i, ingredient := range ingredients {
		eventIngredients[i] = IngredientEvent{
			Name:     ingredient.Name,
			Quantity: ingredient.Quantity,
			Unit:     ingredient.Unit,
			Source:   ingredient.Source,
			Cost:     ingredient.Cost,
		}
	}
	return eventIngredients
}

// convertMetadata converts domain metadata to event metadata
func (e *EventPublisherAdapter) convertMetadata(metadata entities.BeverageMetadata) map[string]interface{} {
	return map[string]interface{}{
		"estimated_cost":        metadata.EstimatedCost,
		"preparation_time":      metadata.PreparationTime,
		"difficulty":            metadata.Difficulty,
		"tags":                  metadata.Tags,
		"target_audience":       metadata.TargetAudience,
		"seasonal_availability": metadata.SeasonalAvailability,
	}
}
