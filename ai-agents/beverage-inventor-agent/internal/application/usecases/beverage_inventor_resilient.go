package usecases

import (
	"context"
	"fmt"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/repositories"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/services"
	"go-coffee-ai-agents/internal/errors"
	"go-coffee-ai-agents/internal/resilience"
)

// ResilientBeverageInventorUseCase handles beverage invention with resilience patterns
type ResilientBeverageInventorUseCase struct {
	beverageRepo       repositories.BeverageRepository
	eventPublisher     repositories.EventPublisher
	aiProvider         repositories.AIProvider
	taskManager        repositories.TaskManager
	notificationSvc    repositories.NotificationService
	generatorService   *services.BeverageGeneratorService
	resilienceManager  *resilience.ResilienceManager
	logger             Logger
}

// NewResilientBeverageInventorUseCase creates a new resilient beverage inventor use case
func NewResilientBeverageInventorUseCase(
	beverageRepo repositories.BeverageRepository,
	eventPublisher repositories.EventPublisher,
	aiProvider repositories.AIProvider,
	taskManager repositories.TaskManager,
	notificationSvc repositories.NotificationService,
	resilienceManager *resilience.ResilienceManager,
	logger Logger,
) *ResilientBeverageInventorUseCase {
	return &ResilientBeverageInventorUseCase{
		beverageRepo:      beverageRepo,
		eventPublisher:    eventPublisher,
		aiProvider:        aiProvider,
		taskManager:       taskManager,
		notificationSvc:   notificationSvc,
		generatorService:  services.NewBeverageGeneratorService(),
		resilienceManager: resilienceManager,
		logger:            logger,
	}
}

// InventBeverage creates a new beverage with resilience patterns
func (uc *ResilientBeverageInventorUseCase) InventBeverage(ctx context.Context, req *InventBeverageRequest) (*InventBeverageResponse, error) {
	uc.logger.Info("Starting resilient beverage invention", 
		"ingredients", req.Ingredients, 
		"theme", req.Theme, 
		"use_ai", req.UseAI)

	// Validate request with structured error
	if err := uc.validateRequest(req); err != nil {
		return nil, errors.NewValidationError("INVALID_REQUEST", err.Error()).
			WithContext("ingredients", req.Ingredients).
			WithContext("theme", req.Theme)
	}

	var beverage *entities.Beverage
	var warnings []string
	var aiUsed bool

	if req.UseAI && uc.aiProvider != nil {
		// Use AI to generate beverage with resilience
		var err error
		beverage, warnings, err = uc.generateBeverageWithAIResilient(ctx, req)
		if err != nil {
			uc.logger.Warn("AI generation failed, falling back to rule-based generation", "error", err)
			beverage, err = uc.generateBeverageWithRules(req)
			if err != nil {
				return nil, errors.NewBusinessError("GENERATION_FAILED", "Failed to generate beverage").
					WithContext("fallback_attempted", true)
			}
		} else {
			aiUsed = true
		}
	} else {
		// Use rule-based generation
		var err error
		beverage, err = uc.generateBeverageWithRules(req)
		if err != nil {
			return nil, errors.NewBusinessError("GENERATION_FAILED", "Failed to generate beverage with rules")
		}
	}

	// Set creator
	beverage.CreatedBy = req.CreatedBy

	// Apply constraints
	if err := uc.applyConstraints(beverage, req.Constraints); err != nil {
		warnings = append(warnings, fmt.Sprintf("Constraint violation: %v", err))
	}

	// Save beverage with database resilience
	err := uc.resilienceManager.ExecuteWithResilienceConfig(ctx, "save_beverage", "database", func(ctx context.Context) error {
		return uc.beverageRepo.Save(ctx, beverage)
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "SAVE_FAILED", "Failed to save beverage").
			WithContext("beverage_id", beverage.ID.String())
	}

	// Publish event with Kafka resilience
	err = uc.resilienceManager.ExecuteWithResilienceConfig(ctx, "publish_event", "kafka", func(ctx context.Context) error {
		return uc.eventPublisher.PublishBeverageCreated(ctx, beverage)
	})
	if err != nil {
		uc.logger.Error("Failed to publish beverage created event", err, "beverage_id", beverage.ID)
		// Don't fail the entire operation for event publishing
	}

	// Create task for barista with external API resilience
	taskID, taskCreated := uc.createBaristaTaskResilient(ctx, beverage)

	// Send notification with resilience
	uc.sendNotificationResilient(ctx, beverage, taskCreated, taskID)

	uc.logger.Info("Resilient beverage invention completed", 
		"beverage_id", beverage.ID, 
		"name", beverage.Name,
		"ai_used", aiUsed,
		"task_created", taskCreated)

	return &InventBeverageResponse{
		Beverage:    beverage,
		TaskCreated: taskCreated,
		TaskID:      taskID,
		AIUsed:      aiUsed,
		Warnings:    warnings,
	}, nil
}

// generateBeverageWithAIResilient generates a beverage using AI with resilience patterns
func (uc *ResilientBeverageInventorUseCase) generateBeverageWithAIResilient(ctx context.Context, req *InventBeverageRequest) (*entities.Beverage, []string, error) {
	var warnings []string
	var analysis *repositories.IngredientAnalysis
	var err error

	// Analyze ingredients with AI provider resilience
	err = uc.resilienceManager.ExecuteWithResilienceConfig(ctx, "analyze_ingredients", "ai_provider", func(ctx context.Context) error {
		analysis, err = uc.aiProvider.AnalyzeIngredients(ctx, req.Ingredients)
		return err
	})
	if err != nil {
		return nil, nil, errors.NewAIProviderError("ingredient_analysis", err).
			WithContext("ingredients", req.Ingredients)
	}

	if !analysis.Compatible {
		warnings = append(warnings, "AI detected potential ingredient compatibility issues")
	}
	warnings = append(warnings, analysis.Warnings...)

	// Generate base beverage using rules (no external dependency)
	beverage, err := uc.generatorService.GenerateBeverage(req.Ingredients, req.Theme)
	if err != nil {
		return nil, warnings, errors.NewBusinessError("RULE_GENERATION_FAILED", err.Error())
	}

	// Enhance description with AI (with resilience)
	err = uc.resilienceManager.ExecuteWithResilienceConfig(ctx, "generate_description", "ai_provider", func(ctx context.Context) error {
		aiDescription, err := uc.aiProvider.GenerateDescription(ctx, beverage)
		if err == nil {
			beverage.Description = aiDescription
		}
		return err
	})
	if err != nil {
		uc.logger.Warn("Failed to generate AI description", "error", err)
		// Continue with rule-based description
	}

	// Get AI suggestions for improvements (with resilience)
	err = uc.resilienceManager.ExecuteWithResilienceConfig(ctx, "suggest_improvements", "ai_provider", func(ctx context.Context) error {
		suggestions, err := uc.aiProvider.SuggestImprovements(ctx, beverage)
		if err == nil && len(suggestions) > 0 {
			beverage.Metadata.Tags = append(beverage.Metadata.Tags, "AI-Enhanced")
		}
		return err
	})
	if err != nil {
		uc.logger.Warn("Failed to get AI suggestions", "error", err)
		// Continue without suggestions
	}

	return beverage, warnings, nil
}

// createBaristaTaskResilient creates a task with external API resilience
func (uc *ResilientBeverageInventorUseCase) createBaristaTaskResilient(ctx context.Context, beverage *entities.Beverage) (string, bool) {
	if uc.taskManager == nil {
		return "", false
	}

	task := &repositories.Task{
		Title:       fmt.Sprintf("Test New Beverage: %s", beverage.Name),
		Description: uc.generateTaskDescription(beverage),
		Status:      repositories.TaskStatusOpen,
		Priority:    repositories.TaskPriorityNormal,
		Tags:        []string{"beverage-testing", "new-recipe", beverage.Theme},
		CustomFields: map[string]interface{}{
			"beverage_id":    beverage.ID.String(),
			"estimated_cost": beverage.Metadata.EstimatedCost,
			"prep_time":      beverage.Metadata.PreparationTime,
		},
	}

	var taskID string
	err := uc.resilienceManager.ExecuteWithResilienceConfig(ctx, "create_task", "external_api", func(ctx context.Context) error {
		err := uc.taskManager.CreateTask(ctx, task)
		if err == nil {
			taskID = task.ID
		}
		return err
	})

	if err != nil {
		uc.logger.Error("Failed to create barista task with resilience", err, 
			"beverage_id", beverage.ID,
			"error_type", fmt.Sprintf("%T", err))
		return "", false
	}

	return taskID, true
}

// sendNotificationResilient sends notifications with resilience patterns
func (uc *ResilientBeverageInventorUseCase) sendNotificationResilient(ctx context.Context, beverage *entities.Beverage, taskCreated bool, taskID string) {
	if uc.notificationSvc == nil {
		return
	}

	message := fmt.Sprintf("🎉 New beverage invented: *%s*\n\n%s\n\nEstimated cost: $%.2f\nPrep time: %d minutes",
		beverage.Name, beverage.Description, beverage.Metadata.EstimatedCost, beverage.Metadata.PreparationTime)

	if taskCreated {
		message += fmt.Sprintf("\n\n📋 Task created for testing: %s", taskID)
	}

	// Send to Slack with resilience
	err := uc.resilienceManager.ExecuteWithResilienceConfig(ctx, "send_slack_notification", "external_api", func(ctx context.Context) error {
		return uc.notificationSvc.SendSlackMessage(ctx, "#beverage-innovation", message)
	})

	if err != nil {
		uc.logger.Error("Failed to send Slack notification with resilience", err, 
			"beverage_id", beverage.ID,
			"error_type", fmt.Sprintf("%T", err))
	}
}

// validateRequest validates the invention request with structured errors
func (uc *ResilientBeverageInventorUseCase) validateRequest(req *InventBeverageRequest) error {
	if len(req.Ingredients) == 0 {
		return fmt.Errorf("at least one ingredient is required")
	}

	if req.Theme == "" {
		return fmt.Errorf("theme is required")
	}

	if req.CreatedBy == "" {
		return fmt.Errorf("creator is required")
	}

	// Validate constraints
	if req.Constraints.MaxCost != nil && *req.Constraints.MaxCost <= 0 {
		return fmt.Errorf("max cost must be positive")
	}

	if req.Constraints.MaxCalories != nil && *req.Constraints.MaxCalories <= 0 {
		return fmt.Errorf("max calories must be positive")
	}

	if req.Constraints.MaxPrepTime != nil && *req.Constraints.MaxPrepTime <= 0 {
		return fmt.Errorf("max preparation time must be positive")
	}

	return nil
}

// generateBeverageWithRules generates a beverage using rule-based logic
func (uc *ResilientBeverageInventorUseCase) generateBeverageWithRules(req *InventBeverageRequest) (*entities.Beverage, error) {
	return uc.generatorService.GenerateBeverage(req.Ingredients, req.Theme)
}

// applyConstraints applies constraints to the beverage
func (uc *ResilientBeverageInventorUseCase) applyConstraints(beverage *entities.Beverage, constraints BeverageConstraints) error {
	if constraints.MaxCost != nil && beverage.Metadata.EstimatedCost > *constraints.MaxCost {
		return fmt.Errorf("beverage cost %.2f exceeds maximum %.2f", beverage.Metadata.EstimatedCost, *constraints.MaxCost)
	}

	if constraints.MaxCalories != nil && beverage.GetTotalCalories() > *constraints.MaxCalories {
		return fmt.Errorf("beverage calories %d exceed maximum %d", beverage.GetTotalCalories(), *constraints.MaxCalories)
	}

	if constraints.MaxPrepTime != nil && beverage.Metadata.PreparationTime > *constraints.MaxPrepTime {
		return fmt.Errorf("preparation time %d minutes exceeds maximum %d", beverage.Metadata.PreparationTime, *constraints.MaxPrepTime)
	}

	// Check allergen-free requirements
	if len(constraints.AllergenFree) > 0 {
		allergens := beverage.GetAllAllergens()
		for _, requiredFree := range constraints.AllergenFree {
			for _, allergen := range allergens {
				if allergen == requiredFree {
					return fmt.Errorf("beverage contains allergen %s which should be avoided", allergen)
				}
			}
		}
	}

	return nil
}

// generateTaskDescription generates a detailed task description for the barista
func (uc *ResilientBeverageInventorUseCase) generateTaskDescription(beverage *entities.Beverage) string {
	description := fmt.Sprintf(`New Beverage Recipe: %s

Description: %s

Ingredients:
`, beverage.Name, beverage.Description)

	for _, ingredient := range beverage.Ingredients {
		description += fmt.Sprintf("- %s: %.1f %s\n", ingredient.Name, ingredient.Quantity, ingredient.Unit)
	}

	description += fmt.Sprintf(`
Estimated Cost: $%.2f
Preparation Time: %d minutes
Difficulty: %s

Instructions for Barista:
1. Gather all ingredients listed above
2. Follow standard preparation procedures for %s theme
3. Document the actual preparation process
4. Prepare a sample for tasting and feedback
5. Note any adjustments needed
6. Update the task with results and photos

Target Audience: %s
`, beverage.Metadata.EstimatedCost, beverage.Metadata.PreparationTime, beverage.Metadata.Difficulty, beverage.Theme, beverage.Metadata.TargetAudience)

	return description
}
