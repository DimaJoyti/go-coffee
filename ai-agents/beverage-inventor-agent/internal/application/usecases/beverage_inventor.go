package usecases

import (
	"context"
	"fmt"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/repositories"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/services"
)

// BeverageInventorUseCase handles the business logic for beverage invention
type BeverageInventorUseCase struct {
	beverageRepo     repositories.BeverageRepository
	eventPublisher   repositories.EventPublisher
	aiProvider       repositories.AIProvider
	taskManager      repositories.TaskManager
	notificationSvc  repositories.NotificationService
	generatorService *services.BeverageGeneratorService

	// Enhanced services
	nutritionalAnalyzer   *services.NutritionalAnalyzer
	costCalculator        *services.CostCalculator
	compatibilityAnalyzer *services.IngredientCompatibilityAnalyzer
	recipeOptimizer       *services.RecipeOptimizer

	logger Logger
}

// Logger defines the logging interface
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewBeverageInventorUseCase creates a new beverage inventor use case
func NewBeverageInventorUseCase(
	beverageRepo repositories.BeverageRepository,
	eventPublisher repositories.EventPublisher,
	aiProvider repositories.AIProvider,
	taskManager repositories.TaskManager,
	notificationSvc repositories.NotificationService,
	nutritionalAnalyzer *services.NutritionalAnalyzer,
	costCalculator *services.CostCalculator,
	compatibilityAnalyzer *services.IngredientCompatibilityAnalyzer,
	recipeOptimizer *services.RecipeOptimizer,
	logger Logger,
) *BeverageInventorUseCase {
	return &BeverageInventorUseCase{
		beverageRepo:          beverageRepo,
		eventPublisher:        eventPublisher,
		aiProvider:            aiProvider,
		taskManager:           taskManager,
		notificationSvc:       notificationSvc,
		generatorService:      services.NewBeverageGeneratorService(),
		nutritionalAnalyzer:   nutritionalAnalyzer,
		costCalculator:        costCalculator,
		compatibilityAnalyzer: compatibilityAnalyzer,
		recipeOptimizer:       recipeOptimizer,
		logger:                logger,
	}
}

// InventBeverageRequest represents a request to invent a new beverage
type InventBeverageRequest struct {
	Ingredients    []string            `json:"ingredients"`
	Theme          string              `json:"theme"`
	UseAI          bool                `json:"use_ai"`
	CreatedBy      string              `json:"created_by"`
	TargetAudience []string            `json:"target_audience,omitempty"`
	Constraints    BeverageConstraints `json:"constraints,omitempty"`

	// Enhanced options
	OptimizationGoals *OptimizationGoals       `json:"optimization_goals,omitempty"`
	ServingSize       float64                  `json:"serving_size,omitempty"`
	BatchSize         float64                  `json:"batch_size,omitempty"`
	DietaryProfile    *services.DietaryProfile `json:"dietary_profile,omitempty"`
	MarketContext     *services.MarketContext  `json:"market_context,omitempty"`
	AnalysisLevel     string                   `json:"analysis_level,omitempty"` // basic, detailed, comprehensive
}

// OptimizationGoals represents optimization goals for beverage creation
type OptimizationGoals struct {
	PrioritizeTaste     bool    `json:"prioritize_taste"`
	PrioritizeCost      bool    `json:"prioritize_cost"`
	PrioritizeNutrition bool    `json:"prioritize_nutrition"`
	TargetMargin        float64 `json:"target_margin,omitempty"` // profit margin percentage
	MaxCostPerServing   float64 `json:"max_cost_per_serving,omitempty"`
	HealthScore         float64 `json:"health_score,omitempty"` // target health score 0-100
}

// BeverageConstraints represents constraints for beverage creation
type BeverageConstraints struct {
	MaxCost       *float64 `json:"max_cost,omitempty"`
	MaxCalories   *int     `json:"max_calories,omitempty"`
	MaxPrepTime   *int     `json:"max_prep_time,omitempty"`
	RequiredTags  []string `json:"required_tags,omitempty"`
	ForbiddenTags []string `json:"forbidden_tags,omitempty"`
	AllergenFree  []string `json:"allergen_free,omitempty"`
}

// InventBeverageResponse represents the response from beverage invention
type InventBeverageResponse struct {
	Beverage    *entities.Beverage `json:"beverage"`
	TaskCreated bool               `json:"task_created"`
	TaskID      string             `json:"task_id,omitempty"`
	AIUsed      bool               `json:"ai_used"`
	Warnings    []string           `json:"warnings,omitempty"`

	// Enhanced analysis results
	NutritionalAnalysis   *services.NutritionalAnalysisResult   `json:"nutritional_analysis,omitempty"`
	CostAnalysis          *services.CostBreakdown               `json:"cost_analysis,omitempty"`
	CompatibilityAnalysis *services.CompatibilityAnalysisResult `json:"compatibility_analysis,omitempty"`
	OptimizationResult    *services.OptimizationResult          `json:"optimization_result,omitempty"`

	// Additional insights
	HealthScore     float64  `json:"health_score,omitempty"`
	EstimatedCost   float64  `json:"estimated_cost,omitempty"`
	ProfitMargin    float64  `json:"profit_margin,omitempty"`
	MarketFitScore  float64  `json:"market_fit_score,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// InventBeverage creates a new beverage based on the request
func (uc *BeverageInventorUseCase) InventBeverage(ctx context.Context, req *InventBeverageRequest) (*InventBeverageResponse, error) {
	uc.logger.Info("Starting enhanced beverage invention", "ingredients", req.Ingredients, "theme", req.Theme, "use_ai", req.UseAI)

	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Set defaults
	if req.ServingSize == 0 {
		req.ServingSize = 350 // Default 350ml serving
	}
	if req.BatchSize == 0 {
		req.BatchSize = 1 // Default single serving
	}
	if req.AnalysisLevel == "" {
		req.AnalysisLevel = "detailed"
	}

	var beverage *entities.Beverage
	var warnings []string
	var aiUsed bool

	// Generate base beverage
	if req.UseAI && uc.aiProvider != nil {
		var err error
		beverage, warnings, err = uc.generateBeverageWithAI(ctx, req)
		if err != nil {
			uc.logger.Warn("AI generation failed, falling back to rule-based generation", "error", err)
			beverage, err = uc.generateBeverageWithRules(req)
			if err != nil {
				return nil, fmt.Errorf("failed to generate beverage: %w", err)
			}
		} else {
			aiUsed = true
		}
	} else {
		var err error
		beverage, err = uc.generateBeverageWithRules(req)
		if err != nil {
			return nil, fmt.Errorf("failed to generate beverage: %w", err)
		}
	}

	// Set creator
	beverage.CreatedBy = req.CreatedBy

	// Perform enhanced analysis and optimization
	response := &InventBeverageResponse{
		Beverage:        beverage,
		AIUsed:          aiUsed,
		Warnings:        warnings,
		Recommendations: []string{},
	}

	// Enhanced analysis pipeline
	if err := uc.performEnhancedAnalysis(ctx, req, response); err != nil {
		uc.logger.Warn("Enhanced analysis failed", "error", err)
		warnings = append(warnings, fmt.Sprintf("Enhanced analysis failed: %v", err))
	}

	// Apply constraints and optimization
	if err := uc.applyConstraintsAndOptimize(ctx, req, response); err != nil {
		warnings = append(warnings, fmt.Sprintf("Constraint/optimization issue: %v", err))
	}

	// Update warnings
	response.Warnings = warnings

	// Save beverage
	if err := uc.beverageRepo.Save(ctx, response.Beverage); err != nil {
		return nil, fmt.Errorf("failed to save beverage: %w", err)
	}

	// Publish event
	if err := uc.eventPublisher.PublishBeverageCreated(ctx, response.Beverage); err != nil {
		uc.logger.Error("Failed to publish beverage created event", err, "beverage_id", response.Beverage.ID)
	}

	// Create task for barista
	taskID, taskCreated := uc.createEnhancedBaristaTask(ctx, response)
	response.TaskCreated = taskCreated
	response.TaskID = taskID

	// Send enhanced notification
	uc.sendEnhancedNotification(ctx, response)

	uc.logger.Info("Enhanced beverage invention completed",
		"beverage_id", response.Beverage.ID,
		"name", response.Beverage.Name,
		"health_score", response.HealthScore,
		"estimated_cost", response.EstimatedCost)

	return response, nil
}

// validateRequest validates the invention request
func (uc *BeverageInventorUseCase) validateRequest(req *InventBeverageRequest) error {
	if len(req.Ingredients) == 0 {
		return fmt.Errorf("at least one ingredient is required")
	}

	if req.Theme == "" {
		return fmt.Errorf("theme is required")
	}

	if req.CreatedBy == "" {
		return fmt.Errorf("creator is required")
	}

	return nil
}

// performEnhancedAnalysis performs comprehensive analysis of the beverage
func (uc *BeverageInventorUseCase) performEnhancedAnalysis(ctx context.Context, req *InventBeverageRequest, response *InventBeverageResponse) error {
	beverage := response.Beverage

	// Nutritional Analysis
	if uc.nutritionalAnalyzer != nil {
		analysisLevel := services.AnalysisLevelDetailed
		if req.AnalysisLevel == "comprehensive" {
			analysisLevel = services.AnalysisLevelComprehensive
		} else if req.AnalysisLevel == "basic" {
			analysisLevel = services.AnalysisLevelBasic
		}

		nutritionReq := &services.NutritionalAnalysisRequest{
			Beverage:       beverage,
			ServingSize:    req.ServingSize,
			DietaryProfile: req.DietaryProfile,
			AnalysisLevel:  analysisLevel,
		}

		nutritionResult, err := uc.nutritionalAnalyzer.AnalyzeNutrition(ctx, nutritionReq)
		if err == nil {
			response.NutritionalAnalysis = nutritionResult
			response.HealthScore = nutritionResult.Score
			response.Recommendations = append(response.Recommendations, nutritionResult.Recommendations...)
		}
	}

	// Cost Analysis
	if uc.costCalculator != nil {
		costReq := &services.CostCalculationRequest{
			Beverage:        beverage,
			ServingSize:     req.ServingSize,
			BatchSize:       req.BatchSize,
			IncludeShipping: true,
			IncludeLabor:    true,
			IncludeOverhead: true,
		}

		costResult, err := uc.costCalculator.CalculateCost(ctx, costReq)
		if err == nil {
			response.CostAnalysis = costResult
			response.EstimatedCost = costResult.CostPerServing

			// Calculate profit margin if optimization goals provided
			if req.OptimizationGoals != nil && req.OptimizationGoals.TargetMargin > 0 {
				suggestedPrice := costResult.CostPerServing / (1 - req.OptimizationGoals.TargetMargin/100)
				response.ProfitMargin = req.OptimizationGoals.TargetMargin
				response.Recommendations = append(response.Recommendations,
					fmt.Sprintf("Suggested selling price: $%.2f for %.1f%% margin", suggestedPrice, req.OptimizationGoals.TargetMargin))
			}
		}
	}

	// Compatibility Analysis
	if uc.compatibilityAnalyzer != nil {
		compatibilityReq := &services.CompatibilityAnalysisRequest{
			Ingredients:   beverage.Ingredients,
			BeverageType:  beverage.Theme,
			AnalysisLevel: services.CompatibilityAnalysisDetailed,
		}

		if req.DietaryProfile != nil {
			compatibilityReq.Constraints = &services.CompatibilityConstraints{
				Allergens:           req.DietaryProfile.Allergies,
				DietaryRestrictions: req.DietaryProfile.DietaryRestrictions,
				FlavorPreferences:   req.DietaryProfile.PreferredFlavors,
			}
		}

		compatibilityResult, err := uc.compatibilityAnalyzer.AnalyzeCompatibility(ctx, compatibilityReq)
		if err == nil {
			response.CompatibilityAnalysis = compatibilityResult
			response.Recommendations = append(response.Recommendations, compatibilityResult.Recommendations...)

			// Add warnings for conflicts
			for _, conflict := range compatibilityResult.Conflicts {
				if conflict.Severity == "high" {
					response.Warnings = append(response.Warnings,
						fmt.Sprintf("High severity conflict: %s", conflict.Description))
				}
			}
		}
	}

	return nil
}

// applyConstraintsAndOptimize applies constraints and performs optimization
func (uc *BeverageInventorUseCase) applyConstraintsAndOptimize(ctx context.Context, req *InventBeverageRequest, response *InventBeverageResponse) error {
	// Apply basic constraints
	if err := uc.applyConstraints(response.Beverage, req.Constraints); err != nil {
		return err
	}

	// Perform optimization if goals are provided
	if req.OptimizationGoals != nil && uc.recipeOptimizer != nil {
		objectives := &services.OptimizationObjectives{
			TasteWeight:          0.4,
			CostWeight:           0.3,
			NutritionWeight:      0.2,
			CompatibilityWeight:  0.1,
			MaxIterations:        50,
			ConvergenceThreshold: 0.01,
		}

		// Adjust weights based on goals
		if req.OptimizationGoals.PrioritizeTaste {
			objectives.TasteWeight = 0.6
			objectives.CostWeight = 0.2
			objectives.NutritionWeight = 0.2
		} else if req.OptimizationGoals.PrioritizeCost {
			objectives.CostWeight = 0.6
			objectives.TasteWeight = 0.2
			objectives.NutritionWeight = 0.2
		} else if req.OptimizationGoals.PrioritizeNutrition {
			objectives.NutritionWeight = 0.6
			objectives.TasteWeight = 0.2
			objectives.CostWeight = 0.2
		}

		// Set target values
		if req.OptimizationGoals.MaxCostPerServing > 0 {
			objectives.MaxCostPerServing = &req.OptimizationGoals.MaxCostPerServing
		}
		if req.OptimizationGoals.HealthScore > 0 {
			// Convert health score to nutrition constraints
			// This would need more sophisticated mapping
		}

		optimizationReq := &services.OptimizationRequest{
			BaseRecipe:       response.Beverage,
			Objectives:       objectives,
			ServingSize:      req.ServingSize,
			BatchSize:        req.BatchSize,
			Market:           req.MarketContext,
			OptimizationType: services.OptimizationTypeMultiObjective,
		}

		optimizationResult, err := uc.recipeOptimizer.OptimizeRecipe(ctx, optimizationReq)
		if err == nil {
			response.OptimizationResult = optimizationResult

			// Use optimized recipe if it's significantly better
			if optimizationResult.BestRecipe != nil &&
				optimizationResult.BestRecipe.OptimizationScore > 75 {
				response.Beverage = optimizationResult.BestRecipe.Recipe
				response.Recommendations = append(response.Recommendations,
					"Recipe optimized for better performance")
				response.Recommendations = append(response.Recommendations,
					optimizationResult.Recommendations...)
			}
		}
	}

	return nil
}

// generateBeverageWithAI generates a beverage using AI
func (uc *BeverageInventorUseCase) generateBeverageWithAI(ctx context.Context, req *InventBeverageRequest) (*entities.Beverage, []string, error) {
	var warnings []string

	// Analyze ingredients first
	analysis, err := uc.aiProvider.AnalyzeIngredients(ctx, req.Ingredients)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to analyze ingredients: %w", err)
	}

	if !analysis.Compatible {
		warnings = append(warnings, "AI detected potential ingredient compatibility issues")
	}

	warnings = append(warnings, analysis.Warnings...)

	// Generate base beverage using rules
	beverage, err := uc.generatorService.GenerateBeverage(req.Ingredients, req.Theme)
	if err != nil {
		return nil, warnings, err
	}

	// Enhance description with AI
	aiDescription, err := uc.aiProvider.GenerateDescription(ctx, beverage)
	if err != nil {
		uc.logger.Warn("Failed to generate AI description", "error", err)
	} else {
		beverage.Description = aiDescription
	}

	// Get AI suggestions for improvements
	suggestions, err := uc.aiProvider.SuggestImprovements(ctx, beverage)
	if err != nil {
		uc.logger.Warn("Failed to get AI suggestions", "error", err)
	} else if len(suggestions) > 0 {
		// Add suggestions as metadata
		beverage.Metadata.Tags = append(beverage.Metadata.Tags, "AI-Enhanced")
		// Store suggestions in a custom field (would need to extend metadata)
	}

	return beverage, warnings, nil
}

// generateBeverageWithRules generates a beverage using rule-based logic
func (uc *BeverageInventorUseCase) generateBeverageWithRules(req *InventBeverageRequest) (*entities.Beverage, error) {
	return uc.generatorService.GenerateBeverage(req.Ingredients, req.Theme)
}

// applyConstraints applies constraints to the beverage
func (uc *BeverageInventorUseCase) applyConstraints(beverage *entities.Beverage, constraints BeverageConstraints) error {
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

// createBaristaTask creates a task for the barista to test the new beverage
func (uc *BeverageInventorUseCase) createBaristaTask(ctx context.Context, beverage *entities.Beverage) (string, bool) {
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

	if err := uc.taskManager.CreateTask(ctx, task); err != nil {
		uc.logger.Error("Failed to create barista task", err, "beverage_id", beverage.ID)
		return "", false
	}

	return task.ID, true
}

// generateTaskDescription generates a detailed task description for the barista
func (uc *BeverageInventorUseCase) generateTaskDescription(beverage *entities.Beverage) string {
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

// sendNotification sends notifications about the new beverage
func (uc *BeverageInventorUseCase) sendNotification(ctx context.Context, beverage *entities.Beverage, taskCreated bool, taskID string) {
	if uc.notificationSvc == nil {
		return
	}

	message := fmt.Sprintf("🎉 New beverage invented: *%s*\n\n%s\n\nEstimated cost: $%.2f\nPrep time: %d minutes",
		beverage.Name, beverage.Description, beverage.Metadata.EstimatedCost, beverage.Metadata.PreparationTime)

	if taskCreated {
		message += fmt.Sprintf("\n\n📋 Task created for testing: %s", taskID)
	}

	// Send to Slack (assuming there's a general channel)
	if err := uc.notificationSvc.SendSlackMessage(ctx, "#beverage-innovation", message); err != nil {
		uc.logger.Error("Failed to send Slack notification", err, "beverage_id", beverage.ID)
	}
}

// GetBeverage retrieves a beverage by ID
func (uc *BeverageInventorUseCase) GetBeverage(ctx context.Context, id string) (*entities.Beverage, error) {
	// Parse UUID
	beverageID, err := entities.ParseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid beverage ID: %w", err)
	}

	return uc.beverageRepo.FindByID(ctx, beverageID)
}

// ListBeverages retrieves beverages with pagination
func (uc *BeverageInventorUseCase) ListBeverages(ctx context.Context, offset, limit int) ([]*entities.Beverage, error) {
	return uc.beverageRepo.List(ctx, offset, limit)
}

// UpdateBeverageStatus updates the status of a beverage
func (uc *BeverageInventorUseCase) UpdateBeverageStatus(ctx context.Context, id string, status entities.BeverageStatus) error {
	beverageID, err := entities.ParseUUID(id)
	if err != nil {
		return fmt.Errorf("invalid beverage ID: %w", err)
	}

	beverage, err := uc.beverageRepo.FindByID(ctx, beverageID)
	if err != nil {
		return fmt.Errorf("failed to find beverage: %w", err)
	}

	oldStatus := beverage.Status
	beverage.UpdateStatus(status)

	if err := uc.beverageRepo.Update(ctx, beverage); err != nil {
		return fmt.Errorf("failed to update beverage: %w", err)
	}

	// Publish status change event
	if err := uc.eventPublisher.PublishBeverageStatusChanged(ctx, beverage, oldStatus); err != nil {
		uc.logger.Error("Failed to publish status change event", err, "beverage_id", beverage.ID)
	}

	return nil
}

// createEnhancedBaristaTask creates an enhanced task for the barista with detailed analysis
func (uc *BeverageInventorUseCase) createEnhancedBaristaTask(ctx context.Context, response *InventBeverageResponse) (string, bool) {
	if uc.taskManager == nil {
		return "", false
	}

	beverage := response.Beverage

	// Generate enhanced task description
	description := uc.generateEnhancedTaskDescription(response)

	// Create task with enhanced metadata
	task := &repositories.Task{
		Title:       fmt.Sprintf("Test New Beverage: %s", beverage.Name),
		Description: description,
		Status:      repositories.TaskStatusOpen,
		Priority:    repositories.TaskPriorityHigh, // Enhanced recipes get higher priority
		Tags:        []string{"beverage-testing", "enhanced-recipe", beverage.Theme},
		CustomFields: map[string]interface{}{
			"beverage_id":      beverage.ID.String(),
			"beverage_name":    beverage.Name,
			"theme":            beverage.Theme,
			"estimated_cost":   response.EstimatedCost,
			"health_score":     response.HealthScore,
			"profit_margin":    response.ProfitMargin,
			"analysis_level":   "enhanced",
		},
	}

	// Add analysis results to custom fields if available
	if response.NutritionalAnalysis != nil {
		task.CustomFields["nutrition_score"] = response.NutritionalAnalysis.Score
		if response.NutritionalAnalysis.HealthAnalysis != nil {
			task.CustomFields["allergens"] = response.NutritionalAnalysis.HealthAnalysis.Allergens
		}
	}

	if response.CompatibilityAnalysis != nil {
		task.CustomFields["compatibility_score"] = response.CompatibilityAnalysis.OverallCompatibility
		task.CustomFields["flavor_conflicts"] = len(response.CompatibilityAnalysis.Conflicts)
	}

	if err := uc.taskManager.CreateTask(ctx, task); err != nil {
		uc.logger.Error("Failed to create enhanced barista task", err, "beverage_id", beverage.ID)
		return "", false
	}

	uc.logger.Info("Enhanced barista task created", "task_id", task.ID, "beverage_id", beverage.ID)
	return task.ID, true
}

// generateEnhancedTaskDescription generates a comprehensive task description
func (uc *BeverageInventorUseCase) generateEnhancedTaskDescription(response *InventBeverageResponse) string {
	beverage := response.Beverage

	description := fmt.Sprintf(`🎯 Enhanced Beverage Recipe Testing: %s

📋 Basic Information:
• Description: %s
• Theme: %s
• Created by: %s

🧪 Ingredients:
`, beverage.Name, beverage.Description, beverage.Theme, beverage.CreatedBy)

	for _, ingredient := range beverage.Ingredients {
		description += fmt.Sprintf("  • %s: %.1f %s\n", ingredient.Name, ingredient.Quantity, ingredient.Unit)
	}

	// Add enhanced analysis results
	if response.EstimatedCost > 0 {
		description += fmt.Sprintf("\n💰 Cost Analysis:\n")
		description += fmt.Sprintf("  • Estimated cost per serving: $%.2f\n", response.EstimatedCost)
		if response.ProfitMargin > 0 {
			description += fmt.Sprintf("  • Target profit margin: %.1f%%\n", response.ProfitMargin)
		}
	}

	if response.HealthScore > 0 {
		description += fmt.Sprintf("\n🏥 Health Analysis:\n")
		description += fmt.Sprintf("  • Health score: %.1f/100\n", response.HealthScore)
	}

	if response.NutritionalAnalysis != nil && response.NutritionalAnalysis.BasicNutrition != nil {
		nutrition := response.NutritionalAnalysis.BasicNutrition
		description += fmt.Sprintf("  • Calories: %.0f\n", nutrition.Calories)
		description += fmt.Sprintf("  • Protein: %.1fg\n", nutrition.Protein)
		description += fmt.Sprintf("  • Sugar: %.1fg\n", nutrition.Sugar)
		if nutrition.Caffeine > 0 {
			description += fmt.Sprintf("  • Caffeine: %.0fmg\n", nutrition.Caffeine)
		}
	}

	if response.CompatibilityAnalysis != nil {
		description += fmt.Sprintf("\n🔬 Compatibility Analysis:\n")
		description += fmt.Sprintf("  • Overall compatibility: %.1f/100\n", response.CompatibilityAnalysis.OverallCompatibility)

		if len(response.CompatibilityAnalysis.Conflicts) > 0 {
			description += fmt.Sprintf("  • ⚠️ Flavor conflicts detected: %d\n", len(response.CompatibilityAnalysis.Conflicts))
		}

		if len(response.CompatibilityAnalysis.Synergies) > 0 {
			description += fmt.Sprintf("  • ✨ Flavor synergies: %d\n", len(response.CompatibilityAnalysis.Synergies))
		}
	}

	// Add recommendations
	if len(response.Recommendations) > 0 {
		description += fmt.Sprintf("\n💡 Recommendations:\n")
		for _, rec := range response.Recommendations {
			description += fmt.Sprintf("  • %s\n", rec)
		}
	}

	// Add warnings
	if len(response.Warnings) > 0 {
		description += fmt.Sprintf("\n⚠️ Warnings:\n")
		for _, warning := range response.Warnings {
			description += fmt.Sprintf("  • %s\n", warning)
		}
	}

	description += fmt.Sprintf(`

📝 Testing Instructions:
1. 🛒 Gather all ingredients listed above
2. 📏 Measure ingredients precisely as specified
3. 🥤 Follow standard preparation procedures for %s theme
4. 📊 Document the actual preparation process with photos
5. 👅 Prepare samples for taste testing
6. 📝 Record taste notes, texture, aroma, and overall impression
7. 💰 Verify actual cost vs estimated cost
8. 🔄 Note any adjustments needed for improvement
9. 📋 Update this task with detailed results and feedback
10. 📸 Include photos of the final product

🎯 Success Criteria:
• Recipe produces consistent results
• Taste meets quality standards
• Cost aligns with estimates
• No major ingredient conflicts observed
• Positive feedback from taste testers

📊 Please rate the following (1-10):
• Taste: ___/10
• Appearance: ___/10
• Aroma: ___/10
• Texture: ___/10
• Overall Appeal: ___/10
• Cost Effectiveness: ___/10

Target Audience: %s
`, beverage.Theme, beverage.Metadata.TargetAudience)

	return description
}

// sendEnhancedNotification sends enhanced notifications with analysis results
func (uc *BeverageInventorUseCase) sendEnhancedNotification(ctx context.Context, response *InventBeverageResponse) {
	if uc.notificationSvc == nil {
		return
	}

	beverage := response.Beverage

	// Create rich notification message
	message := fmt.Sprintf(`🎉 *New Enhanced Beverage Recipe Created!*

🥤 *%s*
%s

📊 *Analysis Summary:*`, beverage.Name, beverage.Description)

	if response.HealthScore > 0 {
		healthEmoji := "🟢"
		if response.HealthScore < 60 {
			healthEmoji = "🟡"
		}
		if response.HealthScore < 40 {
			healthEmoji = "🔴"
		}
		message += fmt.Sprintf("\n%s Health Score: %.1f/100", healthEmoji, response.HealthScore)
	}

	if response.EstimatedCost > 0 {
		costEmoji := "💰"
		if response.EstimatedCost > 5 {
			costEmoji = "💸"
		}
		message += fmt.Sprintf("\n%s Estimated Cost: $%.2f per serving", costEmoji, response.EstimatedCost)
	}

	if response.CompatibilityAnalysis != nil {
		compatEmoji := "✅"
		if response.CompatibilityAnalysis.OverallCompatibility < 70 {
			compatEmoji = "⚠️"
		}
		message += fmt.Sprintf("\n%s Ingredient Compatibility: %.1f/100", compatEmoji, response.CompatibilityAnalysis.OverallCompatibility)
	}

	// Add key recommendations
	if len(response.Recommendations) > 0 {
		message += "\n\n💡 *Key Recommendations:*"
		for i, rec := range response.Recommendations {
			if i >= 3 { // Limit to top 3 recommendations
				break
			}
			message += fmt.Sprintf("\n• %s", rec)
		}
	}

	// Add warnings if any
	if len(response.Warnings) > 0 {
		message += "\n\n⚠️ *Warnings:*"
		for i, warning := range response.Warnings {
			if i >= 2 { // Limit to top 2 warnings
				break
			}
			message += fmt.Sprintf("\n• %s", warning)
		}
	}

	if response.TaskCreated {
		message += fmt.Sprintf("\n\n📋 *Testing Task Created:* %s", response.TaskID)
	}

	message += fmt.Sprintf("\n\n🏷️ Theme: %s | 👥 Target: %s",
		beverage.Theme, beverage.Metadata.TargetAudience)

	// Send to Slack
	if err := uc.notificationSvc.SendSlackMessage(ctx, "#beverage-innovation", message); err != nil {
		uc.logger.Error("Failed to send enhanced Slack notification", err, "beverage_id", beverage.ID)
	}
}

// AnalyzeBeverage performs comprehensive analysis of an existing beverage
func (uc *BeverageInventorUseCase) AnalyzeBeverage(ctx context.Context, beverageID string, analysisLevel string) (*InventBeverageResponse, error) {
	uc.logger.Info("Starting comprehensive beverage analysis", "beverage_id", beverageID, "analysis_level", analysisLevel)

	// Get the beverage
	beverage, err := uc.GetBeverage(ctx, beverageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get beverage: %w", err)
	}

	// Create analysis request
	req := &InventBeverageRequest{
		Ingredients:   []string{}, // Will be populated from beverage
		Theme:         beverage.Theme,
		UseAI:         true,
		CreatedBy:     beverage.CreatedBy,
		ServingSize:   350, // Default serving size
		BatchSize:     1,   // Single serving analysis
		AnalysisLevel: analysisLevel,
	}

	// Extract ingredient names
	for _, ingredient := range beverage.Ingredients {
		req.Ingredients = append(req.Ingredients, ingredient.Name)
	}

	// Create response structure
	response := &InventBeverageResponse{
		Beverage:        beverage,
		AIUsed:          true,
		Warnings:        []string{},
		Recommendations: []string{},
	}

	// Perform enhanced analysis
	if err := uc.performEnhancedAnalysis(ctx, req, response); err != nil {
		uc.logger.Warn("Enhanced analysis failed", "error", err)
		response.Warnings = append(response.Warnings, fmt.Sprintf("Enhanced analysis failed: %v", err))
	}

	uc.logger.Info("Beverage analysis completed",
		"beverage_id", beverage.ID,
		"health_score", response.HealthScore,
		"estimated_cost", response.EstimatedCost)

	return response, nil
}

// OptimizeBeverage optimizes an existing beverage recipe
func (uc *BeverageInventorUseCase) OptimizeBeverage(ctx context.Context, beverageID string, goals *OptimizationGoals) (*services.OptimizationResult, error) {
	uc.logger.Info("Starting beverage optimization", "beverage_id", beverageID)

	if uc.recipeOptimizer == nil {
		return nil, fmt.Errorf("recipe optimizer not available")
	}

	// Get the beverage
	beverage, err := uc.GetBeverage(ctx, beverageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get beverage: %w", err)
	}

	// Create optimization objectives
	objectives := &services.OptimizationObjectives{
		TasteWeight:          0.4,
		CostWeight:           0.3,
		NutritionWeight:      0.2,
		CompatibilityWeight:  0.1,
		MaxIterations:        100,
		ConvergenceThreshold: 0.01,
	}

	// Adjust weights based on goals
	if goals != nil {
		if goals.PrioritizeTaste {
			objectives.TasteWeight = 0.6
			objectives.CostWeight = 0.2
			objectives.NutritionWeight = 0.2
		} else if goals.PrioritizeCost {
			objectives.CostWeight = 0.6
			objectives.TasteWeight = 0.2
			objectives.NutritionWeight = 0.2
		} else if goals.PrioritizeNutrition {
			objectives.NutritionWeight = 0.6
			objectives.TasteWeight = 0.2
			objectives.CostWeight = 0.2
		}

		if goals.MaxCostPerServing > 0 {
			objectives.MaxCostPerServing = &goals.MaxCostPerServing
		}
	}

	// Create optimization request
	optimizationReq := &services.OptimizationRequest{
		BaseRecipe:       beverage,
		Objectives:       objectives,
		ServingSize:      350,
		BatchSize:        1,
		OptimizationType: services.OptimizationTypeMultiObjective,
	}

	// Perform optimization
	result, err := uc.recipeOptimizer.OptimizeRecipe(ctx, optimizationReq)
	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}

	uc.logger.Info("Beverage optimization completed",
		"beverage_id", beverage.ID,
		"optimization_score", result.BestRecipe.OptimizationScore)

	return result, nil
}
