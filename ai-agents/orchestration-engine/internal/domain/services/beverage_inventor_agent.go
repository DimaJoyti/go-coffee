package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BeverageInventorAgent implements the Agent interface for beverage invention operations
type BeverageInventorAgent struct {
	baseURL string
	client  HTTPClient
	metrics *AgentMetrics
	mutex   sync.RWMutex
	logger  Logger
}

// NewBeverageInventorAgent creates a new beverage inventor agent
func NewBeverageInventorAgent(baseURL string, client HTTPClient, logger Logger) *BeverageInventorAgent {
	return &BeverageInventorAgent{
		baseURL: baseURL,
		client:  client,
		metrics: &AgentMetrics{
			LastUpdated: time.Now(),
		},
		logger: logger,
	}
}

// Execute executes an action on the beverage inventor agent
func (bia *BeverageInventorAgent) Execute(ctx context.Context, action string, input map[string]interface{}) (map[string]interface{}, error) {
	bia.updateMetrics(true, time.Now())
	defer func(start time.Time) {
		bia.updateResponseTime(time.Since(start))
	}(time.Now())

	switch action {
	case "invent_beverage":
		return bia.inventBeverage(ctx, input)
	case "optimize_recipe":
		return bia.optimizeRecipe(ctx, input)
	case "refine_recipe":
		return bia.refineRecipe(ctx, input)
	case "suggest_improvements":
		return bia.suggestImprovements(ctx, input)
	case "generate_batch_instructions":
		return bia.generateBatchInstructions(ctx, input)
	case "analyze_ingredients":
		return bia.analyzeIngredients(ctx, input)
	case "calculate_nutrition":
		return bia.calculateNutrition(ctx, input)
	case "generate_variations":
		return bia.generateVariations(ctx, input)
	case "validate_recipe":
		return bia.validateRecipe(ctx, input)
	case "estimate_complexity":
		return bia.estimateComplexity(ctx, input)
	default:
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

// GetCapabilities returns the capabilities of the agent
func (bia *BeverageInventorAgent) GetCapabilities() []string {
	return []string{
		"invent_beverage",
		"optimize_recipe",
		"refine_recipe",
		"suggest_improvements",
		"generate_batch_instructions",
		"analyze_ingredients",
		"calculate_nutrition",
		"generate_variations",
		"validate_recipe",
		"estimate_complexity",
	}
}

// GetStatus returns the current status of the agent
func (bia *BeverageInventorAgent) GetStatus() AgentStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := bia.client.Get(ctx, bia.baseURL+"/health")
	if err != nil {
		return AgentStatusOffline
	}

	return AgentStatusOnline
}

// Validate validates the input for an action
func (bia *BeverageInventorAgent) Validate(action string, input map[string]interface{}) error {
	switch action {
	case "invent_beverage":
		return bia.validateInventBeverage(input)
	case "optimize_recipe":
		return bia.validateOptimizeRecipe(input)
	case "refine_recipe":
		return bia.validateRefineRecipe(input)
	case "suggest_improvements":
		return bia.validateSuggestImprovements(input)
	case "generate_batch_instructions":
		return bia.validateGenerateBatchInstructions(input)
	case "analyze_ingredients":
		return bia.validateAnalyzeIngredients(input)
	case "calculate_nutrition":
		return bia.validateCalculateNutrition(input)
	case "generate_variations":
		return bia.validateGenerateVariations(input)
	case "validate_recipe":
		return bia.validateValidateRecipe(input)
	case "estimate_complexity":
		return bia.validateEstimateComplexity(input)
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
}

// GetMetrics returns the current metrics for the agent
func (bia *BeverageInventorAgent) GetMetrics() *AgentMetrics {
	bia.mutex.RLock()
	defer bia.mutex.RUnlock()

	metricsCopy := *bia.metrics
	return &metricsCopy
}

// Action implementations

func (bia *BeverageInventorAgent) inventBeverage(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Inventing new beverage", "input", input)

	// Prepare request payload
	payload := map[string]interface{}{
		"flavor_profile":       input["flavor_profile"],
		"dietary_requirements": input["dietary_requirements"],
		"seasonal_preferences": input["seasonal_preferences"],
		"target_market":        input["target_market"],
		"market_trends":        input["market_trends"],
		"innovation_level":     "high",
		"include_nutrition":    true,
		"include_cost_estimate": true,
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/invent", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to invent beverage: %w", err)
	}

	// Enhance result with additional metadata
	enhancedResult := map[string]interface{}{
		"recipe":           result["recipe"],
		"ingredients":      result["ingredients"],
		"instructions":     result["instructions"],
		"nutrition_info":   result["nutrition"],
		"cost_estimate":    result["cost_estimate"],
		"complexity_score": result["complexity"],
		"innovation_score": result["innovation_score"],
		"recipe_id":        result["id"],
		"created_at":       time.Now(),
		"constraints":      result["constraints"],
		"alternatives":     result["alternatives"],
	}

	bia.updateMetrics(true, time.Now())
	return enhancedResult, nil
}

func (bia *BeverageInventorAgent) optimizeRecipe(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Optimizing recipe", "input", input)

	payload := map[string]interface{}{
		"recipe":       input["recipe"],
		"budget_range": input["budget_range"],
		"constraints":  input["constraints"],
		"optimization_goals": []string{"cost", "taste", "production_efficiency"},
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/optimize", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to optimize recipe: %w", err)
	}

	optimizedResult := map[string]interface{}{
		"optimized_recipe":   result["optimized_recipe"],
		"ingredients":        result["ingredients"],
		"cost_reduction":     result["cost_reduction"],
		"complexity_score":   result["complexity"],
		"optimization_notes": result["notes"],
		"recipe_id":          result["id"],
		"original_recipe":    input["recipe"],
		"improvements":       result["improvements"],
	}

	bia.updateMetrics(true, time.Now())
	return optimizedResult, nil
}

func (bia *BeverageInventorAgent) refineRecipe(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Refining recipe", "input", input)

	payload := map[string]interface{}{
		"recipe":           input["recipe"],
		"cost_constraints": input["cost_constraints"],
		"feedback":         input["feedback"],
		"refinement_focus": "feasibility",
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/refine", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to refine recipe: %w", err)
	}

	refinedResult := map[string]interface{}{
		"refined_recipe":    result["refined_recipe"],
		"changes_made":      result["changes"],
		"feasibility_score": result["feasibility"],
		"cost_impact":       result["cost_impact"],
		"recipe_id":         result["id"],
		"refinement_notes":  result["notes"],
	}

	bia.updateMetrics(true, time.Now())
	return refinedResult, nil
}

func (bia *BeverageInventorAgent) suggestImprovements(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Suggesting improvements", "input", input)

	payload := map[string]interface{}{
		"recipe":           input["recipe"],
		"feedback_analysis": input["feedback_analysis"],
		"low_scores":       input["low_scores"],
		"improvement_areas": []string{"taste", "texture", "aroma", "presentation"},
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/improve", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to suggest improvements: %w", err)
	}

	improvementResult := map[string]interface{}{
		"suggestions":       result["suggestions"],
		"priority_changes":  result["priority_changes"],
		"alternative_ingredients": result["alternatives"],
		"expected_impact":   result["impact"],
		"implementation_difficulty": result["difficulty"],
		"cost_implications": result["cost_impact"],
	}

	bia.updateMetrics(true, time.Now())
	return improvementResult, nil
}

func (bia *BeverageInventorAgent) generateBatchInstructions(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Generating batch instructions", "input", input)

	payload := map[string]interface{}{
		"final_recipe":   input["final_recipe"],
		"batch_size":     input["batch_size"],
		"equipment_list": input["equipment_list"],
		"detail_level":   "comprehensive",
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/batch-instructions", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to generate batch instructions: %w", err)
	}

	instructionResult := map[string]interface{}{
		"step_by_step_instructions": result["instructions"],
		"equipment_setup":           result["equipment_setup"],
		"ingredient_preparation":    result["ingredient_prep"],
		"quality_checkpoints":       result["quality_checks"],
		"timing_guidelines":         result["timing"],
		"troubleshooting_guide":     result["troubleshooting"],
		"safety_notes":              result["safety"],
	}

	bia.updateMetrics(true, time.Now())
	return instructionResult, nil
}

func (bia *BeverageInventorAgent) analyzeIngredients(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Analyzing ingredients", "input", input)

	payload := map[string]interface{}{
		"ingredients": input["ingredients"],
		"analysis_type": "comprehensive",
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/analyze-ingredients", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to analyze ingredients: %w", err)
	}

	analysisResult := map[string]interface{}{
		"ingredient_analysis":  result["analysis"],
		"compatibility_matrix": result["compatibility"],
		"flavor_interactions":  result["interactions"],
		"nutritional_profile":  result["nutrition"],
		"allergen_information": result["allergens"],
		"sourcing_recommendations": result["sourcing"],
	}

	bia.updateMetrics(true, time.Now())
	return analysisResult, nil
}

func (bia *BeverageInventorAgent) calculateNutrition(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Calculating nutrition", "input", input)

	payload := map[string]interface{}{
		"recipe":      input["recipe"],
		"serving_size": input["serving_size"],
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/nutrition", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to calculate nutrition: %w", err)
	}

	nutritionResult := map[string]interface{}{
		"nutrition_facts":    result["nutrition_facts"],
		"calorie_breakdown":  result["calories"],
		"macro_nutrients":    result["macros"],
		"micro_nutrients":    result["micros"],
		"dietary_labels":     result["labels"],
		"health_benefits":    result["benefits"],
	}

	bia.updateMetrics(true, time.Now())
	return nutritionResult, nil
}

func (bia *BeverageInventorAgent) generateVariations(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Generating variations", "input", input)

	payload := map[string]interface{}{
		"base_recipe":     input["base_recipe"],
		"variation_count": input["variation_count"],
		"variation_types": []string{"seasonal", "dietary", "intensity", "temperature"},
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/variations", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to generate variations: %w", err)
	}

	variationResult := map[string]interface{}{
		"variations":        result["variations"],
		"seasonal_options":  result["seasonal"],
		"dietary_options":   result["dietary"],
		"intensity_levels":  result["intensity"],
		"temperature_variants": result["temperature"],
	}

	bia.updateMetrics(true, time.Now())
	return variationResult, nil
}

func (bia *BeverageInventorAgent) validateRecipe(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Validating recipe", "input", input)

	payload := map[string]interface{}{
		"recipe": input["recipe"],
		"validation_level": "strict",
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/validate", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to validate recipe: %w", err)
	}

	validationResult := map[string]interface{}{
		"is_valid":          result["valid"],
		"validation_errors": result["errors"],
		"warnings":          result["warnings"],
		"suggestions":       result["suggestions"],
		"compliance_check":  result["compliance"],
	}

	bia.updateMetrics(true, time.Now())
	return validationResult, nil
}

func (bia *BeverageInventorAgent) estimateComplexity(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	bia.logger.Info("Estimating complexity", "input", input)

	payload := map[string]interface{}{
		"recipe": input["recipe"],
		"factors": []string{"ingredients", "techniques", "equipment", "time"},
	}

	result, err := bia.client.Post(ctx, bia.baseURL+"/api/v1/beverages/complexity", payload)
	if err != nil {
		bia.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to estimate complexity: %w", err)
	}

	complexityResult := map[string]interface{}{
		"complexity_score":   result["score"],
		"complexity_factors": result["factors"],
		"skill_level_required": result["skill_level"],
		"time_estimate":      result["time"],
		"equipment_complexity": result["equipment"],
	}

	bia.updateMetrics(true, time.Now())
	return complexityResult, nil
}

// Validation methods

func (bia *BeverageInventorAgent) validateInventBeverage(input map[string]interface{}) error {
	required := []string{"flavor_profile"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (bia *BeverageInventorAgent) validateOptimizeRecipe(input map[string]interface{}) error {
	required := []string{"recipe"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (bia *BeverageInventorAgent) validateRefineRecipe(input map[string]interface{}) error {
	required := []string{"recipe"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (bia *BeverageInventorAgent) validateSuggestImprovements(input map[string]interface{}) error {
	required := []string{"recipe", "feedback_analysis"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (bia *BeverageInventorAgent) validateGenerateBatchInstructions(input map[string]interface{}) error {
	required := []string{"final_recipe"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (bia *BeverageInventorAgent) validateAnalyzeIngredients(input map[string]interface{}) error {
	required := []string{"ingredients"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (bia *BeverageInventorAgent) validateCalculateNutrition(input map[string]interface{}) error {
	required := []string{"recipe"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (bia *BeverageInventorAgent) validateGenerateVariations(input map[string]interface{}) error {
	required := []string{"base_recipe"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (bia *BeverageInventorAgent) validateValidateRecipe(input map[string]interface{}) error {
	required := []string{"recipe"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (bia *BeverageInventorAgent) validateEstimateComplexity(input map[string]interface{}) error {
	required := []string{"recipe"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

// Helper methods

func (bia *BeverageInventorAgent) updateMetrics(success bool, timestamp time.Time) {
	bia.mutex.Lock()
	defer bia.mutex.Unlock()

	bia.metrics.TotalRequests++
	if success {
		bia.metrics.SuccessfulRequests++
	} else {
		bia.metrics.FailedRequests++
	}
	bia.metrics.LastUpdated = timestamp
}

func (bia *BeverageInventorAgent) updateResponseTime(duration time.Duration) {
	bia.mutex.Lock()
	defer bia.mutex.Unlock()

	if bia.metrics.TotalRequests == 1 {
		bia.metrics.AverageResponseTime = duration
	} else {
		bia.metrics.AverageResponseTime = (bia.metrics.AverageResponseTime + duration) / 2
	}
}
