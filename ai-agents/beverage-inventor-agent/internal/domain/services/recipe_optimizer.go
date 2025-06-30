package services

import (
	"context"
	"fmt"
	"math"
	"sort"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
)

// RecipeOptimizer provides multi-objective recipe optimization
type RecipeOptimizer struct {
	nutritionalAnalyzer *NutritionalAnalyzer
	costCalculator      *CostCalculator
	compatibilityAnalyzer *IngredientCompatibilityAnalyzer
	aiProvider          OptimizationAIProvider
}

// OptimizationAIProvider defines AI capabilities for recipe optimization
type OptimizationAIProvider interface {
	OptimizeRecipe(ctx context.Context, recipe *entities.Beverage, objectives *OptimizationObjectives) (*OptimizedRecipe, error)
	GenerateVariations(ctx context.Context, recipe *entities.Beverage, variationType string, count int) ([]*entities.Beverage, error)
	PredictCustomerPreference(ctx context.Context, recipe *entities.Beverage, demographics *CustomerDemographics) (float64, error)
	AnalyzeMarketFit(ctx context.Context, recipe *entities.Beverage, market *MarketContext) (*MarketFitAnalysis, error)
}

// OptimizationObjectives defines what to optimize for
type OptimizationObjectives struct {
	// Primary objectives (weighted 0-1, should sum to 1)
	TasteWeight      float64 `json:"taste_weight"`
	CostWeight       float64 `json:"cost_weight"`
	NutritionWeight  float64 `json:"nutrition_weight"`
	CompatibilityWeight float64 `json:"compatibility_weight"`
	
	// Target values
	MaxCostPerServing   *float64 `json:"max_cost_per_serving,omitempty"`
	MaxCalories         *float64 `json:"max_calories,omitempty"`
	MinProtein          *float64 `json:"min_protein,omitempty"`
	MaxSugar            *float64 `json:"max_sugar,omitempty"`
	TargetTasteProfile  *TasteProfile `json:"target_taste_profile,omitempty"`
	
	// Constraints
	RequiredIngredients []string `json:"required_ingredients"`
	ForbiddenIngredients []string `json:"forbidden_ingredients"`
	DietaryConstraints  *DietaryProfile `json:"dietary_constraints,omitempty"`
	
	// Optimization settings
	MaxIterations       int     `json:"max_iterations"`
	ConvergenceThreshold float64 `json:"convergence_threshold"`
	PopulationSize      int     `json:"population_size"`
}

// OptimizationRequest represents a request for recipe optimization
type OptimizationRequest struct {
	BaseRecipe    *entities.Beverage      `json:"base_recipe"`
	Objectives    *OptimizationObjectives `json:"objectives"`
	ServingSize   float64                 `json:"serving_size"`
	BatchSize     float64                 `json:"batch_size"`
	Market        *MarketContext          `json:"market,omitempty"`
	Demographics  *CustomerDemographics   `json:"demographics,omitempty"`
	OptimizationType OptimizationType     `json:"optimization_type"`
}

// OptimizationType defines the type of optimization
type OptimizationType string

const (
	OptimizationTypeSingleObjective OptimizationType = "single_objective"
	OptimizationTypeMultiObjective  OptimizationType = "multi_objective"
	OptimizationTypePareto          OptimizationType = "pareto"
	OptimizationTypeGenetic         OptimizationType = "genetic"
)

// OptimizedRecipe represents an optimized recipe
type OptimizedRecipe struct {
	Recipe          *entities.Beverage      `json:"recipe"`
	OptimizationScore float64               `json:"optimization_score"`
	Objectives      *ObjectiveScores        `json:"objectives"`
	Improvements    *ImprovementAnalysis    `json:"improvements"`
	Variations      []*RecipeVariation      `json:"variations"`
	Confidence      float64                 `json:"confidence"`
	Iterations      int                     `json:"iterations"`
	ConvergedAt     int                     `json:"converged_at"`
}

// ObjectiveScores represents scores for each optimization objective
type ObjectiveScores struct {
	TasteScore        float64 `json:"taste_score"`        // 0-100
	CostScore         float64 `json:"cost_score"`         // 0-100 (lower cost = higher score)
	NutritionScore    float64 `json:"nutrition_score"`    // 0-100
	CompatibilityScore float64 `json:"compatibility_score"` // 0-100
	OverallScore      float64 `json:"overall_score"`      // weighted average
}

// ImprovementAnalysis represents analysis of improvements made
type ImprovementAnalysis struct {
	CostReduction     float64  `json:"cost_reduction"`     // percentage
	CalorieReduction  float64  `json:"calorie_reduction"`  // percentage
	NutritionImprovement float64 `json:"nutrition_improvement"` // percentage
	TasteImprovement  float64  `json:"taste_improvement"`  // percentage
	KeyChanges        []string `json:"key_changes"`
	TradeOffs         []string `json:"trade_offs"`
}

// RecipeVariation represents a recipe variation
type RecipeVariation struct {
	Name            string             `json:"name"`
	Recipe          *entities.Beverage `json:"recipe"`
	VariationType   string             `json:"variation_type"`
	Score           float64            `json:"score"`
	Description     string             `json:"description"`
	KeyDifferences  []string           `json:"key_differences"`
}

// CustomerDemographics represents target customer demographics
type CustomerDemographics struct {
	AgeGroup        string   `json:"age_group"`        // 18-25, 26-35, etc.
	Gender          string   `json:"gender"`
	Income          string   `json:"income"`           // low, medium, high
	Lifestyle       string   `json:"lifestyle"`        // health-conscious, busy, etc.
	FlavorPreferences []string `json:"flavor_preferences"`
	DietaryRestrictions []string `json:"dietary_restrictions"`
	Pricesensitivity string   `json:"price_sensitivity"` // low, medium, high
}

// MarketContext represents market context for optimization
type MarketContext struct {
	Region          string   `json:"region"`
	Season          string   `json:"season"`
	Trends          []string `json:"trends"`
	CompetitorPricing *PriceRange `json:"competitor_pricing"`
	TargetMarket    string   `json:"target_market"`    // premium, mass, budget
	Distribution    string   `json:"distribution"`     // cafe, retail, online
}

// MarketFitAnalysis represents market fit analysis
type MarketFitAnalysis struct {
	MarketFitScore  float64  `json:"market_fit_score"`  // 0-100
	TrendAlignment  float64  `json:"trend_alignment"`   // 0-100
	PricePositioning string  `json:"price_positioning"` // competitive, premium, budget
	TargetAppeal    float64  `json:"target_appeal"`     // 0-100
	Recommendations []string `json:"recommendations"`
	Risks           []string `json:"risks"`
	Opportunities   []string `json:"opportunities"`
}

// OptimizationResult represents the complete optimization result
type OptimizationResult struct {
	BestRecipe      *OptimizedRecipe      `json:"best_recipe"`
	ParetoFront     []*OptimizedRecipe    `json:"pareto_front,omitempty"`
	AllCandidates   []*OptimizedRecipe    `json:"all_candidates,omitempty"`
	MarketAnalysis  *MarketFitAnalysis    `json:"market_analysis,omitempty"`
	Recommendations []string              `json:"recommendations"`
	OptimizationLog []OptimizationStep    `json:"optimization_log"`
	ExecutionTime   float64               `json:"execution_time"` // seconds
}

// OptimizationStep represents a step in the optimization process
type OptimizationStep struct {
	Iteration       int     `json:"iteration"`
	BestScore       float64 `json:"best_score"`
	AverageScore    float64 `json:"average_score"`
	Improvement     float64 `json:"improvement"`
	Action          string  `json:"action"`
	Description     string  `json:"description"`
}

// NewRecipeOptimizer creates a new recipe optimizer
func NewRecipeOptimizer(
	nutritionalAnalyzer *NutritionalAnalyzer,
	costCalculator *CostCalculator,
	compatibilityAnalyzer *IngredientCompatibilityAnalyzer,
	aiProvider OptimizationAIProvider,
) *RecipeOptimizer {
	return &RecipeOptimizer{
		nutritionalAnalyzer:   nutritionalAnalyzer,
		costCalculator:        costCalculator,
		compatibilityAnalyzer: compatibilityAnalyzer,
		aiProvider:            aiProvider,
	}
}

// OptimizeRecipe performs multi-objective recipe optimization
func (ro *RecipeOptimizer) OptimizeRecipe(ctx context.Context, req *OptimizationRequest) (*OptimizationResult, error) {
	result := &OptimizationResult{
		Recommendations: []string{},
		OptimizationLog: []OptimizationStep{},
	}
	
	// Validate objectives
	if err := ro.validateObjectives(req.Objectives); err != nil {
		return nil, fmt.Errorf("invalid objectives: %w", err)
	}
	
	// Evaluate base recipe
	baseScore, err := ro.evaluateRecipe(ctx, req.BaseRecipe, req.Objectives, req.ServingSize, req.BatchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate base recipe: %w", err)
	}
	
	// Perform optimization based on type
	switch req.OptimizationType {
	case OptimizationTypeSingleObjective:
		result.BestRecipe, err = ro.optimizeSingleObjective(ctx, req)
	case OptimizationTypeMultiObjective:
		result.BestRecipe, err = ro.optimizeMultiObjective(ctx, req)
	case OptimizationTypePareto:
		result.ParetoFront, err = ro.optimizePareto(ctx, req)
		if len(result.ParetoFront) > 0 {
			result.BestRecipe = result.ParetoFront[0] // Best overall score
		}
	case OptimizationTypeGenetic:
		result.BestRecipe, result.AllCandidates, err = ro.optimizeGenetic(ctx, req)
	default:
		result.BestRecipe, err = ro.optimizeMultiObjective(ctx, req)
	}
	
	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}
	
	// Analyze market fit if context provided
	if req.Market != nil && result.BestRecipe != nil {
		marketAnalysis, err := ro.aiProvider.AnalyzeMarketFit(ctx, result.BestRecipe.Recipe, req.Market)
		if err == nil {
			result.MarketAnalysis = marketAnalysis
		}
	}
	
	// Generate recommendations
	if result.BestRecipe != nil {
		result.Recommendations = ro.generateOptimizationRecommendations(baseScore, result.BestRecipe.Objectives)
	}
	
	return result, nil
}

// evaluateRecipe evaluates a recipe against optimization objectives
func (ro *RecipeOptimizer) evaluateRecipe(ctx context.Context, recipe *entities.Beverage, objectives *OptimizationObjectives, servingSize, batchSize float64) (*ObjectiveScores, error) {
	scores := &ObjectiveScores{}
	
	// Evaluate taste (compatibility-based for now)
	if objectives.TasteWeight > 0 {
		compatibilityReq := &CompatibilityAnalysisRequest{
			Ingredients:   recipe.Ingredients,
			BeverageType:  recipe.Theme,
			AnalysisLevel: CompatibilityAnalysisBasic,
		}
		
		compatibilityResult, err := ro.compatibilityAnalyzer.AnalyzeCompatibility(ctx, compatibilityReq)
		if err == nil {
			scores.TasteScore = compatibilityResult.OverallCompatibility
			scores.CompatibilityScore = compatibilityResult.OverallCompatibility
		}
	}
	
	// Evaluate cost
	if objectives.CostWeight > 0 {
		costReq := &CostCalculationRequest{
			Beverage:    recipe,
			ServingSize: servingSize,
			BatchSize:   batchSize,
		}
		
		costBreakdown, err := ro.costCalculator.CalculateCost(ctx, costReq)
		if err == nil {
			// Convert cost to score (lower cost = higher score)
			maxAcceptableCost := 10.0 // $10 per serving max
			if objectives.MaxCostPerServing != nil {
				maxAcceptableCost = *objectives.MaxCostPerServing
			}
			
			costScore := math.Max(0, (maxAcceptableCost-costBreakdown.CostPerServing)/maxAcceptableCost*100)
			scores.CostScore = math.Min(100, costScore)
		}
	}
	
	// Evaluate nutrition
	if objectives.NutritionWeight > 0 {
		nutritionReq := &NutritionalAnalysisRequest{
			Beverage:      recipe,
			ServingSize:   servingSize,
			AnalysisLevel: AnalysisLevelBasic,
		}
		
		nutritionResult, err := ro.nutritionalAnalyzer.AnalyzeNutrition(ctx, nutritionReq)
		if err == nil {
			scores.NutritionScore = nutritionResult.Score
		}
	}
	
	// Calculate overall score
	scores.OverallScore = (scores.TasteScore*objectives.TasteWeight +
		scores.CostScore*objectives.CostWeight +
		scores.NutritionScore*objectives.NutritionWeight +
		scores.CompatibilityScore*objectives.CompatibilityWeight)
	
	return scores, nil
}

// optimizeMultiObjective performs multi-objective optimization
func (ro *RecipeOptimizer) optimizeMultiObjective(ctx context.Context, req *OptimizationRequest) (*OptimizedRecipe, error) {
	// Use AI provider for sophisticated optimization
	optimized, err := ro.aiProvider.OptimizeRecipe(ctx, req.BaseRecipe, req.Objectives)
	if err != nil {
		return nil, fmt.Errorf("AI optimization failed: %w", err)
	}
	
	// Evaluate the optimized recipe
	scores, err := ro.evaluateRecipe(ctx, optimized.Recipe, req.Objectives, req.ServingSize, req.BatchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate optimized recipe: %w", err)
	}
	
	optimized.Objectives = scores
	optimized.OptimizationScore = scores.OverallScore
	
	// Generate variations
	variations, err := ro.generateVariations(ctx, optimized.Recipe, req.Objectives)
	if err == nil {
		optimized.Variations = variations
	}
	
	return optimized, nil
}

// optimizeSingleObjective optimizes for a single primary objective
func (ro *RecipeOptimizer) optimizeSingleObjective(ctx context.Context, req *OptimizationRequest) (*OptimizedRecipe, error) {
	// Determine primary objective
	primaryObjective := "taste"
	if req.Objectives.CostWeight > req.Objectives.TasteWeight &&
		req.Objectives.CostWeight > req.Objectives.NutritionWeight {
		primaryObjective = "cost"
	} else if req.Objectives.NutritionWeight > req.Objectives.TasteWeight {
		primaryObjective = "nutrition"
	}
	
	// Create focused objectives
	focusedObjectives := *req.Objectives
	switch primaryObjective {
	case "cost":
		focusedObjectives.CostWeight = 0.8
		focusedObjectives.TasteWeight = 0.1
		focusedObjectives.NutritionWeight = 0.1
	case "nutrition":
		focusedObjectives.NutritionWeight = 0.8
		focusedObjectives.TasteWeight = 0.1
		focusedObjectives.CostWeight = 0.1
	default: // taste
		focusedObjectives.TasteWeight = 0.8
		focusedObjectives.CostWeight = 0.1
		focusedObjectives.NutritionWeight = 0.1
	}
	
	// Use multi-objective optimization with focused weights
	focusedReq := *req
	focusedReq.Objectives = &focusedObjectives
	
	return ro.optimizeMultiObjective(ctx, &focusedReq)
}

// optimizePareto finds Pareto-optimal solutions
func (ro *RecipeOptimizer) optimizePareto(ctx context.Context, req *OptimizationRequest) ([]*OptimizedRecipe, error) {
	// Generate multiple candidate recipes with different objective weights
	candidates := []*OptimizedRecipe{}
	
	// Generate candidates with different weight distributions
	weightCombinations := [][]float64{
		{0.7, 0.2, 0.1, 0.0}, // Taste-focused
		{0.2, 0.7, 0.1, 0.0}, // Cost-focused
		{0.2, 0.1, 0.7, 0.0}, // Nutrition-focused
		{0.3, 0.3, 0.3, 0.1}, // Balanced
		{0.4, 0.3, 0.2, 0.1}, // Taste-cost balance
		{0.3, 0.2, 0.4, 0.1}, // Taste-nutrition balance
	}
	
	for _, weights := range weightCombinations {
		candidateReq := *req
		candidateReq.Objectives = &OptimizationObjectives{
			TasteWeight:      weights[0],
			CostWeight:       weights[1],
			NutritionWeight:  weights[2],
			CompatibilityWeight: weights[3],
			MaxIterations:    req.Objectives.MaxIterations,
		}
		
		optimized, err := ro.optimizeMultiObjective(ctx, &candidateReq)
		if err == nil {
			candidates = append(candidates, optimized)
		}
	}
	
	// Find Pareto front
	paretoFront := ro.findParetoFront(candidates)
	
	// Sort by overall score
	sort.Slice(paretoFront, func(i, j int) bool {
		return paretoFront[i].OptimizationScore > paretoFront[j].OptimizationScore
	})
	
	return paretoFront, nil
}

// optimizeGenetic performs genetic algorithm optimization
func (ro *RecipeOptimizer) optimizeGenetic(ctx context.Context, req *OptimizationRequest) (*OptimizedRecipe, []*OptimizedRecipe, error) {
	// This would implement a genetic algorithm for recipe optimization
	// For now, use multi-objective optimization as fallback
	best, err := ro.optimizeMultiObjective(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	
	// Generate population of variations
	variations, err := ro.generateVariations(ctx, best.Recipe, req.Objectives)
	if err != nil {
		return best, []*OptimizedRecipe{best}, nil
	}
	
	// Convert variations to optimized recipes
	population := []*OptimizedRecipe{best}
	for _, variation := range variations {
		scores, err := ro.evaluateRecipe(ctx, variation.Recipe, req.Objectives, req.ServingSize, req.BatchSize)
		if err == nil {
			optimized := &OptimizedRecipe{
				Recipe:            variation.Recipe,
				OptimizationScore: scores.OverallScore,
				Objectives:        scores,
				Confidence:        0.8,
			}
			population = append(population, optimized)
		}
	}
	
	return best, population, nil
}

// generateVariations generates recipe variations
func (ro *RecipeOptimizer) generateVariations(ctx context.Context, recipe *entities.Beverage, objectives *OptimizationObjectives) ([]*RecipeVariation, error) {
	variations := []*RecipeVariation{}
	
	// Generate different types of variations
	variationTypes := []string{"healthier", "cheaper", "premium", "seasonal"}
	
	for _, variationType := range variationTypes {
		aiVariations, err := ro.aiProvider.GenerateVariations(ctx, recipe, variationType, 2)
		if err != nil {
			continue
		}
		
		for i, variation := range aiVariations {
			// Evaluate variation
			scores, err := ro.evaluateRecipe(ctx, variation, objectives, 350, 1) // Default serving size
			if err != nil {
				continue
			}
			
			recipeVariation := &RecipeVariation{
				Name:          fmt.Sprintf("%s_%d", variationType, i+1),
				Recipe:        variation,
				VariationType: variationType,
				Score:         scores.OverallScore,
				Description:   fmt.Sprintf("%s variation of the original recipe", variationType),
				KeyDifferences: ro.findKeyDifferences(recipe, variation),
			}
			
			variations = append(variations, recipeVariation)
		}
	}
	
	// Sort by score
	sort.Slice(variations, func(i, j int) bool {
		return variations[i].Score > variations[j].Score
	})
	
	// Return top 5 variations
	maxVariations := 5
	if len(variations) < maxVariations {
		maxVariations = len(variations)
	}
	
	return variations[:maxVariations], nil
}

// Helper methods
func (ro *RecipeOptimizer) validateObjectives(objectives *OptimizationObjectives) error {
	totalWeight := objectives.TasteWeight + objectives.CostWeight + 
		objectives.NutritionWeight + objectives.CompatibilityWeight
	
	if math.Abs(totalWeight-1.0) > 0.01 {
		return fmt.Errorf("objective weights must sum to 1.0, got %.2f", totalWeight)
	}
	
	return nil
}

func (ro *RecipeOptimizer) findParetoFront(candidates []*OptimizedRecipe) []*OptimizedRecipe {
	paretoFront := []*OptimizedRecipe{}
	
	for _, candidate := range candidates {
		isDominated := false
		
		for _, other := range candidates {
			if ro.dominates(other.Objectives, candidate.Objectives) {
				isDominated = true
				break
			}
		}
		
		if !isDominated {
			paretoFront = append(paretoFront, candidate)
		}
	}
	
	return paretoFront
}

func (ro *RecipeOptimizer) dominates(a, b *ObjectiveScores) bool {
	return a.TasteScore >= b.TasteScore &&
		a.CostScore >= b.CostScore &&
		a.NutritionScore >= b.NutritionScore &&
		a.CompatibilityScore >= b.CompatibilityScore &&
		(a.TasteScore > b.TasteScore ||
			a.CostScore > b.CostScore ||
			a.NutritionScore > b.NutritionScore ||
			a.CompatibilityScore > b.CompatibilityScore)
}

func (ro *RecipeOptimizer) findKeyDifferences(original, variation *entities.Beverage) []string {
	differences := []string{}
	
	// Compare ingredients
	originalIngredients := make(map[string]float64)
	for _, ing := range original.Ingredients {
		originalIngredients[ing.Name] = ing.Quantity
	}
	
	for _, ing := range variation.Ingredients {
		if originalAmount, exists := originalIngredients[ing.Name]; exists {
			if math.Abs(originalAmount-ing.Quantity) > 0.1 {
				differences = append(differences, fmt.Sprintf("Changed %s amount", ing.Name))
			}
		} else {
			differences = append(differences, fmt.Sprintf("Added %s", ing.Name))
		}
	}
	
	// Check for removed ingredients
	variationIngredients := make(map[string]bool)
	for _, ing := range variation.Ingredients {
		variationIngredients[ing.Name] = true
	}
	
	for _, ing := range original.Ingredients {
		if !variationIngredients[ing.Name] {
			differences = append(differences, fmt.Sprintf("Removed %s", ing.Name))
		}
	}
	
	return differences
}

func (ro *RecipeOptimizer) generateOptimizationRecommendations(baseScore, optimizedRecipe *ObjectiveScores) []string {
	recommendations := []string{}
	
	if optimizedRecipe == nil {
		return []string{"Optimization failed - consider adjusting objectives or constraints"}
	}
	
	improvement := optimizedRecipe.OverallScore - baseScore.OverallScore
	
	if improvement > 10 {
		recommendations = append(recommendations, fmt.Sprintf("Significant improvement achieved (%.1f points)", improvement))
	} else if improvement > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Moderate improvement achieved (%.1f points)", improvement))
	} else {
		recommendations = append(recommendations, "Limited improvement possible with current constraints")
	}
	
	// Specific recommendations based on scores
	if optimizedRecipe.CostScore < 50 {
		recommendations = append(recommendations, "Consider cost reduction strategies")
	}
	if optimizedRecipe.NutritionScore < 60 {
		recommendations = append(recommendations, "Consider nutritional enhancements")
	}
	if optimizedRecipe.TasteScore < 70 {
		recommendations = append(recommendations, "Consider flavor profile improvements")
	}
	
	return recommendations
}
