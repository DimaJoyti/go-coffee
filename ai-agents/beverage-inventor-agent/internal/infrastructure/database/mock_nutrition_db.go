package database

import (
	"context"
	"strings"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
)

// MockNutritionDatabase provides mock nutrition data for testing and development
type MockNutritionDatabase struct {
	nutritionData map[string]*entities.NutritionalInfo
	allergenData  map[string][]string
	densityData   map[string]float64
}

// NewMockNutritionDatabase creates a new mock nutrition database
func NewMockNutritionDatabase() *MockNutritionDatabase {
	db := &MockNutritionDatabase{
		nutritionData: make(map[string]*entities.NutritionalInfo),
		allergenData:  make(map[string][]string),
		densityData:   make(map[string]float64),
	}

	// Initialize with common beverage ingredients
	db.initializeData()

	return db
}

// GetIngredientNutrition returns nutritional information for an ingredient
func (db *MockNutritionDatabase) GetIngredientNutrition(ctx context.Context, ingredient string, amount float64, unit string) (*entities.NutritionalInfo, error) {
	ingredient = strings.ToLower(ingredient)

	// Get base nutrition data
	baseNutrition, exists := db.nutritionData[ingredient]
	if !exists {
		// Return default nutrition for unknown ingredients
		baseNutrition = &entities.NutritionalInfo{
			Calories: 50,
			Protein:  1.0,
			Carbs:    10.0,
			Fat:      0.5,
			Sugar:    8.0,
			Caffeine: 0.0,
		}
	}

	// Scale nutrition based on amount and unit
	scaleFactor := db.calculateScaleFactor(amount, unit)

	scaledNutrition := &entities.NutritionalInfo{
		Calories: int(float64(baseNutrition.Calories) * scaleFactor),
		Protein:  baseNutrition.Protein * scaleFactor,
		Carbs:    baseNutrition.Carbs * scaleFactor,
		Fat:      baseNutrition.Fat * scaleFactor,
		Sugar:    baseNutrition.Sugar * scaleFactor,
		Caffeine: baseNutrition.Caffeine * scaleFactor,
	}

	return scaledNutrition, nil
}

// GetIngredientDensity returns the density of an ingredient
func (db *MockNutritionDatabase) GetIngredientDensity(ctx context.Context, ingredient string) (float64, error) {
	ingredient = strings.ToLower(ingredient)

	if density, exists := db.densityData[ingredient]; exists {
		return density, nil
	}

	// Default density for liquids (g/ml)
	return 1.0, nil
}

// GetIngredientAllergens returns allergens for an ingredient
func (db *MockNutritionDatabase) GetIngredientAllergens(ctx context.Context, ingredient string) ([]string, error) {
	ingredient = strings.ToLower(ingredient)

	if allergens, exists := db.allergenData[ingredient]; exists {
		return allergens, nil
	}

	return []string{}, nil
}

// SearchSimilarIngredients returns similar ingredients
func (db *MockNutritionDatabase) SearchSimilarIngredients(ctx context.Context, ingredient string) ([]string, error) {
	ingredient = strings.ToLower(ingredient)
	similar := []string{}

	// Simple similarity search based on common patterns
	for key := range db.nutritionData {
		if strings.Contains(key, ingredient) || strings.Contains(ingredient, key) {
			if key != ingredient {
				similar = append(similar, key)
			}
		}
	}

	// Add some generic alternatives
	switch {
	case strings.Contains(ingredient, "milk"):
		similar = append(similar, "almond milk", "soy milk", "oat milk")
	case strings.Contains(ingredient, "sugar"):
		similar = append(similar, "honey", "stevia", "maple syrup")
	case strings.Contains(ingredient, "coffee"):
		similar = append(similar, "espresso", "cold brew", "decaf coffee")
	}

	return similar, nil
}

// calculateScaleFactor calculates scaling factor based on amount and unit
func (db *MockNutritionDatabase) calculateScaleFactor(amount float64, unit string) float64 {
	// Base nutrition data is per 100ml/100g
	switch strings.ToLower(unit) {
	case "ml", "milliliter", "milliliters":
		return amount / 100.0
	case "l", "liter", "liters":
		return (amount * 1000) / 100.0
	case "g", "gram", "grams":
		return amount / 100.0
	case "kg", "kilogram", "kilograms":
		return (amount * 1000) / 100.0
	case "tsp", "teaspoon", "teaspoons":
		return (amount * 5) / 100.0 // 1 tsp = 5ml
	case "tbsp", "tablespoon", "tablespoons":
		return (amount * 15) / 100.0 // 1 tbsp = 15ml
	case "cup", "cups":
		return (amount * 240) / 100.0 // 1 cup = 240ml
	case "oz", "ounce", "ounces":
		return (amount * 30) / 100.0 // 1 oz = 30ml
	default:
		return amount / 100.0 // Default to ml
	}
}

// initializeData initializes the mock database with common beverage ingredients
func (db *MockNutritionDatabase) initializeData() {
	// Coffee and tea
	db.nutritionData["coffee"] = &entities.NutritionalInfo{
		Calories: 2,
		Protein:  0.3,
		Carbs:    0.0,
		Fat:      0.0,
		Sugar:    0.0,
		Caffeine: 95.0,
	}

	db.nutritionData["espresso"] = &entities.NutritionalInfo{
		Calories: 9,
		Protein:  0.5,
		Carbs:    1.7,
		Fat:      0.2,
		Sugar:    0.0,
		Caffeine: 212.0,
	}

	db.nutritionData["green tea"] = &entities.NutritionalInfo{
		Calories: 2,
		Protein:  0.2,
		Carbs:    0.0,
		Fat:      0.0,
		Sugar:    0.0,
		Caffeine: 25.0,
	}

	db.nutritionData["black tea"] = &entities.NutritionalInfo{
		Calories: 2,
		Protein:  0.0,
		Carbs:    0.7,
		Fat:      0.0,
		Sugar:    0.0,
		Caffeine: 47.0,
	}

	// Milk and dairy
	db.nutritionData["whole milk"] = &entities.NutritionalInfo{
		Calories: 61,
		Protein:  3.2,
		Carbs:    4.8,
		Fat:      3.3,
		Sugar:    5.1,
		Caffeine: 0.0,
	}

	db.nutritionData["almond milk"] = &entities.NutritionalInfo{
		Calories: 17,
		Protein:  0.6,
		Carbs:    1.5,
		Fat:      1.1,
		Sugar:    1.3,
		Caffeine: 0.0,
	}

	db.nutritionData["oat milk"] = &entities.NutritionalInfo{
		Calories: 47,
		Protein:  1.0,
		Carbs:    7.0,
		Fat:      1.5,
		Sugar:    4.0,
		Caffeine: 0.0,
	}

	// Sweeteners
	db.nutritionData["sugar"] = &entities.NutritionalInfo{
		Calories: 387,
		Protein:  0.0,
		Carbs:    100.0,
		Fat:      0.0,
		Sugar:    100.0,
		Caffeine: 0.0,
	}

	db.nutritionData["honey"] = &entities.NutritionalInfo{
		Calories: 304,
		Protein:  0.3,
		Carbs:    82.4,
		Fat:      0.0,
		Sugar:    82.1,
		Caffeine: 0.0,
	}

	db.nutritionData["stevia"] = &entities.NutritionalInfo{
		Calories: 0,
		Protein:  0.0,
		Carbs:    0.0,
		Fat:      0.0,
		Sugar:    0.0,
		Caffeine: 0.0,
	}

	// Spices and flavorings
	db.nutritionData["cinnamon"] = &entities.NutritionalInfo{
		Calories: 247,
		Protein:  4.0,
		Carbs:    81.0,
		Fat:      1.2,
		Sugar:    2.2,
		Caffeine: 0.0,
	}

	db.nutritionData["vanilla"] = &entities.NutritionalInfo{
		Calories: 288,
		Protein:  0.1,
		Carbs:    12.7,
		Fat:      0.1,
		Sugar:    12.7,
		Caffeine: 0.0,
	}

	// Fruits
	db.nutritionData["lemon juice"] = &entities.NutritionalInfo{
		Calories: 22,
		Protein:  0.4,
		Carbs:    6.9,
		Fat:      0.2,
		Sugar:    2.5,
		Caffeine: 0.0,
	}

	// Allergen data
	db.allergenData["whole milk"] = []string{"milk", "dairy"}
	db.allergenData["almond milk"] = []string{"tree nuts", "almonds"}
	db.allergenData["soy milk"] = []string{"soy"}

	// Density data (g/ml)
	db.densityData["coffee"] = 1.0
	db.densityData["whole milk"] = 1.03
	db.densityData["almond milk"] = 1.01
	db.densityData["honey"] = 1.4
	db.densityData["sugar"] = 0.85 // granulated
}
