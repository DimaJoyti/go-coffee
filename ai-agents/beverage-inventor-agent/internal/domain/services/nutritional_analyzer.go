package services

import (
	"context"
	"fmt"
	"math"
	"strings"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
)

// NutritionalAnalyzer provides advanced nutritional analysis for beverages
type NutritionalAnalyzer struct {
	nutritionDB NutritionDatabase
	aiProvider  AIProvider
}

// NutritionDatabase defines the interface for nutrition data
type NutritionDatabase interface {
	GetIngredientNutrition(ctx context.Context, ingredient string, amount float64, unit string) (*entities.NutritionalInfo, error)
	GetIngredientDensity(ctx context.Context, ingredient string) (float64, error)
	GetIngredientAllergens(ctx context.Context, ingredient string) ([]string, error)
	SearchSimilarIngredients(ctx context.Context, ingredient string) ([]string, error)
}

// AIProvider defines AI capabilities for nutritional analysis
type AIProvider interface {
	AnalyzeHealthBenefits(ctx context.Context, ingredients []entities.Ingredient) (*HealthAnalysis, error)
	GenerateNutritionalRecommendations(ctx context.Context, profile *DietaryProfile, nutrition *entities.NutritionalInfo) ([]string, error)
	CalculateGlycemicIndex(ctx context.Context, ingredients []entities.Ingredient) (float64, error)
}

// HealthAnalysis represents comprehensive health analysis
type HealthAnalysis struct {
	OverallScore     float64            `json:"overall_score"`     // 0-100
	HealthBenefits   []string           `json:"health_benefits"`
	HealthConcerns   []string           `json:"health_concerns"`
	Antioxidants     float64            `json:"antioxidants"`      // mg
	Caffeine         float64            `json:"caffeine"`          // mg
	Sugar            float64            `json:"sugar"`             // g
	Allergens        []string           `json:"allergens"`
	DietaryTags      []string           `json:"dietary_tags"`      // vegan, keto, etc.
	GlycemicIndex    float64            `json:"glycemic_index"`    // 0-100
	InflammationScore float64           `json:"inflammation_score"` // -10 to +10
	Recommendations  []string           `json:"recommendations"`
}

// DietaryProfile represents a user's dietary preferences and restrictions
type DietaryProfile struct {
	Allergies        []string  `json:"allergies"`
	DietaryRestrictions []string `json:"dietary_restrictions"` // vegan, keto, etc.
	HealthGoals      []string  `json:"health_goals"`          // weight_loss, energy, etc.
	PreferredFlavors []string  `json:"preferred_flavors"`
	MaxCalories      *float64  `json:"max_calories,omitempty"`
	MaxSugar         *float64  `json:"max_sugar,omitempty"`
	MaxCaffeine      *float64  `json:"max_caffeine,omitempty"`
}

// NutritionalAnalysisRequest represents a request for nutritional analysis
type NutritionalAnalysisRequest struct {
	Beverage        *entities.Beverage `json:"beverage"`
	ServingSize     float64            `json:"serving_size"`     // ml
	DietaryProfile  *DietaryProfile    `json:"dietary_profile,omitempty"`
	AnalysisLevel   AnalysisLevel      `json:"analysis_level"`
}

// AnalysisLevel defines the depth of nutritional analysis
type AnalysisLevel string

const (
	AnalysisLevelBasic       AnalysisLevel = "basic"
	AnalysisLevelDetailed    AnalysisLevel = "detailed"
	AnalysisLevelComprehensive AnalysisLevel = "comprehensive"
)

// NutritionalAnalysisResult represents the result of nutritional analysis
type NutritionalAnalysisResult struct {
	BasicNutrition   *entities.NutritionalInfo `json:"basic_nutrition"`
	HealthAnalysis   *HealthAnalysis           `json:"health_analysis,omitempty"`
	DietaryCompatibility *DietaryCompatibility `json:"dietary_compatibility,omitempty"`
	Recommendations  []string                  `json:"recommendations"`
	Score           float64                   `json:"score"`           // Overall nutritional score 0-100
	Confidence      float64                   `json:"confidence"`      // Analysis confidence 0-1
	Warnings        []string                  `json:"warnings"`
}

// DietaryCompatibility represents compatibility with dietary restrictions
type DietaryCompatibility struct {
	IsVegan         bool     `json:"is_vegan"`
	IsVegetarian    bool     `json:"is_vegetarian"`
	IsGlutenFree    bool     `json:"is_gluten_free"`
	IsKeto          bool     `json:"is_keto"`
	IsPaleo         bool     `json:"is_paleo"`
	IsDairy         bool     `json:"is_dairy_free"`
	IsNutFree       bool     `json:"is_nut_free"`
	IsSoyFree       bool     `json:"is_soy_free"`
	IsLowSugar      bool     `json:"is_low_sugar"`
	IsLowCalorie    bool     `json:"is_low_calorie"`
	CustomTags      []string `json:"custom_tags"`
}

// NewNutritionalAnalyzer creates a new nutritional analyzer
func NewNutritionalAnalyzer(nutritionDB NutritionDatabase, aiProvider AIProvider) *NutritionalAnalyzer {
	return &NutritionalAnalyzer{
		nutritionDB: nutritionDB,
		aiProvider:  aiProvider,
	}
}

// AnalyzeNutrition performs comprehensive nutritional analysis
func (na *NutritionalAnalyzer) AnalyzeNutrition(ctx context.Context, req *NutritionalAnalysisRequest) (*NutritionalAnalysisResult, error) {
	result := &NutritionalAnalysisResult{
		Recommendations: []string{},
		Warnings:        []string{},
	}
	
	// Calculate basic nutrition
	basicNutrition, err := na.calculateBasicNutrition(ctx, req.Beverage, req.ServingSize)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate basic nutrition: %w", err)
	}
	result.BasicNutrition = basicNutrition
	
	// Perform detailed analysis if requested
	if req.AnalysisLevel == AnalysisLevelDetailed || req.AnalysisLevel == AnalysisLevelComprehensive {
		healthAnalysis, err := na.performHealthAnalysis(ctx, req.Beverage)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Health analysis failed: %v", err))
		} else {
			result.HealthAnalysis = healthAnalysis
		}
		
		// Calculate dietary compatibility
		compatibility, err := na.calculateDietaryCompatibility(ctx, req.Beverage)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Dietary compatibility analysis failed: %v", err))
		} else {
			result.DietaryCompatibility = compatibility
		}
	}
	
	// Generate personalized recommendations if profile provided
	if req.DietaryProfile != nil {
		recommendations, err := na.generatePersonalizedRecommendations(ctx, req.DietaryProfile, result)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Recommendation generation failed: %v", err))
		} else {
			result.Recommendations = append(result.Recommendations, recommendations...)
		}
	}
	
	// Calculate overall score
	result.Score = na.calculateNutritionalScore(result)
	result.Confidence = na.calculateConfidence(result)
	
	return result, nil
}

// calculateBasicNutrition calculates basic nutritional information
func (na *NutritionalAnalyzer) calculateBasicNutrition(ctx context.Context, beverage *entities.Beverage, servingSize float64) (*entities.NutritionalInfo, error) {
	totalNutrition := &entities.NutritionalInfo{}
	
	for _, ingredient := range beverage.Ingredients {
		// Get nutrition data for this ingredient
		nutrition, err := na.nutritionDB.GetIngredientNutrition(ctx, ingredient.Name, ingredient.Quantity, ingredient.Unit)
		if err != nil {
			// Log warning but continue with other ingredients
			continue
		}
		
		// Scale nutrition based on serving size
		scaleFactor := servingSize / 100.0 // Assuming nutrition data is per 100ml
		
		// Add to total
		totalNutrition.Calories += int(float64(nutrition.Calories) * scaleFactor)
		totalNutrition.Protein += nutrition.Protein * scaleFactor
		totalNutrition.Carbs += nutrition.Carbs * scaleFactor
		totalNutrition.Fat += nutrition.Fat * scaleFactor
		totalNutrition.Sugar += nutrition.Sugar * scaleFactor
		totalNutrition.Caffeine += nutrition.Caffeine * scaleFactor
		
		// Merge allergens
		for _, allergen := range nutrition.Allergens {
			totalNutrition.Allergens = append(totalNutrition.Allergens, allergen)
		}
	}
	
	return totalNutrition, nil
}

// performHealthAnalysis performs AI-powered health analysis
func (na *NutritionalAnalyzer) performHealthAnalysis(ctx context.Context, beverage *entities.Beverage) (*HealthAnalysis, error) {
	// Use AI to analyze health benefits and concerns
	healthAnalysis, err := na.aiProvider.AnalyzeHealthBenefits(ctx, beverage.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("AI health analysis failed: %w", err)
	}
	
	// Calculate glycemic index
	glycemicIndex, err := na.aiProvider.CalculateGlycemicIndex(ctx, beverage.Ingredients)
	if err != nil {
		glycemicIndex = 0 // Default if calculation fails
	}
	healthAnalysis.GlycemicIndex = glycemicIndex
	
	// Analyze allergens
	allergens := []string{}
	for _, ingredient := range beverage.Ingredients {
		ingredientAllergens, err := na.nutritionDB.GetIngredientAllergens(ctx, ingredient.Name)
		if err == nil {
			allergens = append(allergens, ingredientAllergens...)
		}
	}
	healthAnalysis.Allergens = na.removeDuplicates(allergens)
	
	// Calculate inflammation score
	healthAnalysis.InflammationScore = na.calculateInflammationScore(beverage.Ingredients)
	
	return healthAnalysis, nil
}

// calculateDietaryCompatibility determines dietary compatibility
func (na *NutritionalAnalyzer) calculateDietaryCompatibility(ctx context.Context, beverage *entities.Beverage) (*DietaryCompatibility, error) {
	compatibility := &DietaryCompatibility{
		IsVegan:      true,
		IsVegetarian: true,
		IsGlutenFree: true,
		IsKeto:       true,
		IsPaleo:      true,
		IsDairy:      true,
		IsNutFree:    true,
		IsSoyFree:    true,
		IsLowSugar:   true,
		IsLowCalorie: true,
	}
	
	// Check each ingredient against dietary restrictions
	for _, ingredient := range beverage.Ingredients {
		ingredientName := strings.ToLower(ingredient.Name)
		
		// Check for animal products
		if na.isAnimalProduct(ingredientName) {
			compatibility.IsVegan = false
			if na.isMeat(ingredientName) {
				compatibility.IsVegetarian = false
			}
		}
		
		// Check for gluten
		if na.containsGluten(ingredientName) {
			compatibility.IsGlutenFree = false
		}
		
		// Check for dairy
		if na.isDairy(ingredientName) {
			compatibility.IsDairy = false
		}
		
		// Check for nuts
		if na.isNut(ingredientName) {
			compatibility.IsNutFree = false
		}
		
		// Check for soy
		if na.isSoy(ingredientName) {
			compatibility.IsSoyFree = false
		}
		
		// Check for high carb (keto)
		if na.isHighCarb(ingredientName) {
			compatibility.IsKeto = false
		}
		
		// Check for processed foods (paleo)
		if na.isProcessed(ingredientName) {
			compatibility.IsPaleo = false
		}
	}
	
	// Calculate nutritional info to check sugar and calorie content
	nutritionalInfo, err := na.calculateBasicNutrition(ctx, beverage, 250) // Standard serving size
	if err == nil {
		if nutritionalInfo.Sugar > 10 { // > 10g sugar
			compatibility.IsLowSugar = false
		}
		if nutritionalInfo.Calories > 100 { // > 100 calories
			compatibility.IsLowCalorie = false
		}
	}
	
	return compatibility, nil
}

// generatePersonalizedRecommendations generates recommendations based on dietary profile
func (na *NutritionalAnalyzer) generatePersonalizedRecommendations(ctx context.Context, profile *DietaryProfile, result *NutritionalAnalysisResult) ([]string, error) {
	recommendations := []string{}
	
	// Check against allergies
	if result.HealthAnalysis != nil {
		for _, allergen := range result.HealthAnalysis.Allergens {
			for _, allergy := range profile.Allergies {
				if strings.EqualFold(allergen, allergy) {
					recommendations = append(recommendations, fmt.Sprintf("⚠️ Contains %s - avoid if allergic", allergen))
				}
			}
		}
	}
	
	// Check calorie limits
	if profile.MaxCalories != nil && float64(result.BasicNutrition.Calories) > *profile.MaxCalories {
		recommendations = append(recommendations, fmt.Sprintf("High in calories (%d cal) - consider reducing serving size", result.BasicNutrition.Calories))
	}
	
	// Check sugar limits
	if profile.MaxSugar != nil && result.BasicNutrition.Sugar > *profile.MaxSugar {
		recommendations = append(recommendations, fmt.Sprintf("High in sugar (%.1fg) - consider sugar-free alternatives", result.BasicNutrition.Sugar))
	}
	
	// Check caffeine limits
	if profile.MaxCaffeine != nil && result.BasicNutrition.Caffeine > *profile.MaxCaffeine {
		recommendations = append(recommendations, fmt.Sprintf("High in caffeine (%.0fmg) - consider decaf options", result.BasicNutrition.Caffeine))
	}
	
	// Use AI for personalized recommendations
	aiRecommendations, err := na.aiProvider.GenerateNutritionalRecommendations(ctx, profile, result.BasicNutrition)
	if err == nil {
		recommendations = append(recommendations, aiRecommendations...)
	}
	
	return recommendations, nil
}

// calculateNutritionalScore calculates an overall nutritional score
func (na *NutritionalAnalyzer) calculateNutritionalScore(result *NutritionalAnalysisResult) float64 {
	score := 50.0 // Base score
	
	if result.BasicNutrition == nil {
		return score
	}
	
	// Positive factors
	if result.BasicNutrition.Protein > 5 {
		score += 10 // Good protein content
	}
	
	// Negative factors
	if result.BasicNutrition.Sugar > 20 {
		score -= 15 // High sugar
	}
	if result.BasicNutrition.Calories > 300 {
		score -= 10 // High calories
	}
	
	// Health analysis factors
	if result.HealthAnalysis != nil {
		score += result.HealthAnalysis.OverallScore * 0.3 // 30% weight
		if result.HealthAnalysis.InflammationScore < 0 {
			score += 5 // Anti-inflammatory
		}
		if result.HealthAnalysis.GlycemicIndex < 55 {
			score += 5 // Low glycemic index
		}
	}
	
	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return math.Round(score*10) / 10 // Round to 1 decimal place
}

// calculateConfidence calculates confidence in the analysis
func (na *NutritionalAnalyzer) calculateConfidence(result *NutritionalAnalysisResult) float64 {
	confidence := 0.8 // Base confidence
	
	// Reduce confidence based on warnings
	confidence -= float64(len(result.Warnings)) * 0.1
	
	// Increase confidence if we have detailed analysis
	if result.HealthAnalysis != nil {
		confidence += 0.1
	}
	if result.DietaryCompatibility != nil {
		confidence += 0.1
	}
	
	// Ensure confidence is within bounds
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}
	
	return math.Round(confidence*100) / 100 // Round to 2 decimal places
}

// Helper methods for dietary compatibility checking
func (na *NutritionalAnalyzer) isAnimalProduct(ingredient string) bool {
	animalProducts := []string{"milk", "cream", "butter", "cheese", "yogurt", "honey", "egg", "meat", "chicken", "beef", "pork", "fish", "salmon", "tuna"}
	for _, product := range animalProducts {
		if strings.Contains(ingredient, product) {
			return true
		}
	}
	return false
}

func (na *NutritionalAnalyzer) isMeat(ingredient string) bool {
	meats := []string{"meat", "chicken", "beef", "pork", "fish", "salmon", "tuna", "turkey", "lamb"}
	for _, meat := range meats {
		if strings.Contains(ingredient, meat) {
			return true
		}
	}
	return false
}

func (na *NutritionalAnalyzer) containsGluten(ingredient string) bool {
	glutenSources := []string{"wheat", "barley", "rye", "oats", "flour", "bread", "pasta"}
	for _, source := range glutenSources {
		if strings.Contains(ingredient, source) {
			return true
		}
	}
	return false
}

func (na *NutritionalAnalyzer) isDairy(ingredient string) bool {
	dairyProducts := []string{"milk", "cream", "butter", "cheese", "yogurt", "whey", "casein", "lactose"}
	for _, product := range dairyProducts {
		if strings.Contains(ingredient, product) {
			return true
		}
	}
	return false
}

func (na *NutritionalAnalyzer) isNut(ingredient string) bool {
	nuts := []string{"almond", "peanut", "walnut", "cashew", "pecan", "hazelnut", "pistachio", "macadamia"}
	for _, nut := range nuts {
		if strings.Contains(ingredient, nut) {
			return true
		}
	}
	return false
}

func (na *NutritionalAnalyzer) isSoy(ingredient string) bool {
	soyProducts := []string{"soy", "tofu", "tempeh", "miso", "soybean", "edamame"}
	for _, product := range soyProducts {
		if strings.Contains(ingredient, product) {
			return true
		}
	}
	return false
}

func (na *NutritionalAnalyzer) isHighCarb(ingredient string) bool {
	highCarbIngredients := []string{"sugar", "honey", "syrup", "fruit", "banana", "apple", "orange", "rice", "potato", "bread"}
	for _, highCarbIngredient := range highCarbIngredients {
		if strings.Contains(ingredient, highCarbIngredient) {
			return true
		}
	}
	return false
}

func (na *NutritionalAnalyzer) isProcessed(ingredient string) bool {
	processedIngredients := []string{"syrup", "artificial", "preservative", "additive", "flavor", "coloring", "sweetener"}
	for _, processed := range processedIngredients {
		if strings.Contains(ingredient, processed) {
			return true
		}
	}
	return false
}

func (na *NutritionalAnalyzer) calculateInflammationScore(ingredients []entities.Ingredient) float64 {
	score := 0.0
	
	for _, ingredient := range ingredients {
		name := strings.ToLower(ingredient.Name)
		
		// Anti-inflammatory ingredients (negative score is good)
		if strings.Contains(name, "turmeric") {
			score -= 2
		}
		if strings.Contains(name, "ginger") {
			score -= 1.5
		}
		if strings.Contains(name, "green tea") {
			score -= 1
		}
		if strings.Contains(name, "cinnamon") {
			score -= 1
		}
		
		// Pro-inflammatory ingredients (positive score is bad)
		if strings.Contains(name, "sugar") {
			score += 1
		}
		if strings.Contains(name, "artificial") {
			score += 0.5
		}
	}
	
	// Clamp score between -10 and +10
	if score < -10 {
		score = -10
	}
	if score > 10 {
		score = 10
	}
	
	return math.Round(score*10) / 10
}

func (na *NutritionalAnalyzer) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}
