package services

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
)

// BeverageGeneratorService handles the business logic for generating beverages
type BeverageGeneratorService struct {
	adjectives    []string
	beverageTypes []string
	themes        map[string][]string
}

// NewBeverageGeneratorService creates a new beverage generator service
func NewBeverageGeneratorService() *BeverageGeneratorService {
	return &BeverageGeneratorService{
		adjectives: []string{
			"Exotic", "Mystical", "Vibrant", "Bold", "Smooth", "Crisp", 
			"Sparkling", "Dreamy", "Enchanted", "Galactic", "Cosmic", 
			"Ethereal", "Radiant", "Sublime", "Celestial",
		},
		beverageTypes: []string{
			"Latte", "Elixir", "Brew", "Infusion", "Concoction", "Nectar", 
			"Blend", "Quencher", "Potion", "Essence", "Fusion", "Delight",
		},
		themes: map[string][]string{
			"Mars Base": {
				"perfect for a Martian sunrise",
				"a taste of the red planet",
				"fuels your interplanetary journey",
				"inspired by the Martian landscape",
				"crafted for space explorers",
			},
			"Lunar Mining Corp": {
				"a true taste of the lunar surface",
				"energizes your moon rock excavation",
				"crafted for the lunar explorer",
				"reflects the stark beauty of the moon",
				"powers your lunar operations",
			},
			"Interstellar Trade Federation": {
				"transcends galaxies",
				"a cosmic concoction",
				"for the discerning space traveler",
				"unites flavors from across the cosmos",
				"bridges worlds through taste",
			},
			"Earth Café": {
				"brings you back to Earth",
				"a taste of home",
				"comfort in every sip",
				"grounded in tradition",
				"earthly pleasures await",
			},
		},
	}
}

// GenerateBeverage creates a new beverage based on ingredients and theme
func (s *BeverageGeneratorService) GenerateBeverage(ingredientNames []string, theme string) (*entities.Beverage, error) {
	if len(ingredientNames) == 0 {
		return nil, fmt.Errorf("at least one ingredient is required")
	}

	// Generate creative name
	name := s.generateCreativeName(ingredientNames, theme)
	
	// Generate description
	description := s.generateDescription(ingredientNames, theme)
	
	// Convert ingredient names to ingredient entities
	ingredients := s.createIngredients(ingredientNames)
	
	// Create beverage
	beverage := entities.NewBeverage(name, description, theme, ingredients)
	
	// Set metadata
	s.enrichMetadata(beverage, theme)
	
	return beverage, nil
}

// generateCreativeName creates a creative name for the beverage
func (s *BeverageGeneratorService) generateCreativeName(ingredientNames []string, theme string) string {
	rand.Seed(time.Now().UnixNano())
	
	adj := s.adjectives[rand.Intn(len(s.adjectives))]
	bevType := s.beverageTypes[rand.Intn(len(s.beverageTypes))]
	
	// Use the primary ingredient (first one) in the name
	primaryIngredient := strings.Title(ingredientNames[0])
	
	// Create variations based on theme
	switch theme {
	case "Mars Base":
		return fmt.Sprintf("Martian %s %s", primaryIngredient, bevType)
	case "Lunar Mining Corp":
		return fmt.Sprintf("Lunar %s %s", adj, bevType)
	case "Interstellar Trade Federation":
		return fmt.Sprintf("Galactic %s %s", primaryIngredient, bevType)
	default:
		return fmt.Sprintf("%s %s %s", adj, primaryIngredient, bevType)
	}
}

// generateDescription creates a description for the beverage
func (s *BeverageGeneratorService) generateDescription(ingredientNames []string, theme string) string {
	rand.Seed(time.Now().UnixNano())
	
	bevType := s.beverageTypes[rand.Intn(len(s.beverageTypes))]
	adj := strings.ToLower(s.adjectives[rand.Intn(len(s.adjectives))])
	
	// Get theme-specific phrases
	themePhrases, ok := s.themes[theme]
	if !ok || len(themePhrases) == 0 {
		themePhrases = []string{"a delightful new addition to our menu", "a unique blend"}
	}
	themePhrase := themePhrases[rand.Intn(len(themePhrases))]
	
	// Create ingredient list
	ingredientList := strings.Join(ingredientNames, ", ")
	
	return fmt.Sprintf("A %s %s featuring %s, %s. This creation %s and delivers an unforgettable experience.", 
		adj, strings.ToLower(bevType), ingredientList, themePhrase, "combines innovative flavors")
}

// createIngredients converts ingredient names to ingredient entities
func (s *BeverageGeneratorService) createIngredients(ingredientNames []string) []entities.Ingredient {
	ingredients := make([]entities.Ingredient, len(ingredientNames))
	
	for i, name := range ingredientNames {
		ingredients[i] = entities.Ingredient{
			Name:     name,
			Quantity: s.getDefaultQuantity(name),
			Unit:     s.getDefaultUnit(name),
			Source:   "Local Supplier", // Default source
			Cost:     s.estimateCost(name),
			Nutritional: s.getDefaultNutrition(name),
		}
	}
	
	return ingredients
}

// getDefaultQuantity returns a default quantity for an ingredient
func (s *BeverageGeneratorService) getDefaultQuantity(ingredient string) float64 {
	// Simple heuristics for common ingredients
	switch strings.ToLower(ingredient) {
	case "espresso", "coffee":
		return 2.0 // shots
	case "milk", "water":
		return 200.0 // ml
	case "sugar", "honey":
		return 10.0 // grams
	case "vanilla", "cinnamon":
		return 1.0 // teaspoon
	default:
		return 50.0 // default grams
	}
}

// getDefaultUnit returns a default unit for an ingredient
func (s *BeverageGeneratorService) getDefaultUnit(ingredient string) string {
	switch strings.ToLower(ingredient) {
	case "espresso", "coffee":
		return "shots"
	case "milk", "water":
		return "ml"
	case "vanilla", "cinnamon":
		return "tsp"
	default:
		return "g"
	}
}

// estimateCost provides a rough cost estimate for ingredients
func (s *BeverageGeneratorService) estimateCost(ingredient string) float64 {
	// Simple cost estimation (in cents per unit)
	switch strings.ToLower(ingredient) {
	case "espresso", "coffee":
		return 0.50
	case "milk":
		return 0.01
	case "sugar":
		return 0.005
	case "vanilla":
		return 0.10
	default:
		return 0.05
	}
}

// getDefaultNutrition provides default nutritional information
func (s *BeverageGeneratorService) getDefaultNutrition(ingredient string) entities.NutritionalInfo {
	// Simplified nutritional data per 100g/100ml
	switch strings.ToLower(ingredient) {
	case "espresso", "coffee":
		return entities.NutritionalInfo{
			Calories: 2,
			Protein:  0.1,
			Carbs:    0.0,
			Fat:      0.0,
			Sugar:    0.0,
			Caffeine: 64.0,
			Allergens: []string{},
		}
	case "milk":
		return entities.NutritionalInfo{
			Calories: 42,
			Protein:  3.4,
			Carbs:    5.0,
			Fat:      1.0,
			Sugar:    5.0,
			Caffeine: 0.0,
			Allergens: []string{"dairy"},
		}
	case "sugar":
		return entities.NutritionalInfo{
			Calories: 387,
			Protein:  0.0,
			Carbs:    100.0,
			Fat:      0.0,
			Sugar:    100.0,
			Caffeine: 0.0,
			Allergens: []string{},
		}
	default:
		return entities.NutritionalInfo{
			Calories: 20,
			Protein:  1.0,
			Carbs:    4.0,
			Fat:      0.1,
			Sugar:    2.0,
			Caffeine: 0.0,
			Allergens: []string{},
		}
	}
}

// enrichMetadata adds additional metadata to the beverage
func (s *BeverageGeneratorService) enrichMetadata(beverage *entities.Beverage, theme string) {
	beverage.Metadata.PreparationTime = s.estimatePreparationTime(beverage.Ingredients)
	beverage.Metadata.Difficulty = s.assessDifficulty(beverage.Ingredients)
	beverage.Metadata.Tags = s.generateTags(beverage, theme)
	beverage.Metadata.TargetAudience = s.determineTargetAudience(theme)
	beverage.Metadata.SeasonalAvailability = []string{"All Year"}
	
	// Calculate cost
	beverage.CalculateTotalCost()
}

// estimatePreparationTime estimates how long the beverage takes to prepare
func (s *BeverageGeneratorService) estimatePreparationTime(ingredients []entities.Ingredient) int {
	baseTime := 3 // minutes
	complexityTime := len(ingredients) * 2
	return baseTime + complexityTime
}

// assessDifficulty determines the difficulty level
func (s *BeverageGeneratorService) assessDifficulty(ingredients []entities.Ingredient) string {
	if len(ingredients) <= 3 {
		return "Easy"
	} else if len(ingredients) <= 5 {
		return "Medium"
	}
	return "Hard"
}

// generateTags creates relevant tags for the beverage
func (s *BeverageGeneratorService) generateTags(beverage *entities.Beverage, theme string) []string {
	tags := []string{"AI-Generated", "Innovative"}
	
	// Add theme-based tags
	switch theme {
	case "Mars Base":
		tags = append(tags, "Space-Themed", "Futuristic")
	case "Lunar Mining Corp":
		tags = append(tags, "Lunar", "Industrial")
	case "Interstellar Trade Federation":
		tags = append(tags, "Cosmic", "Premium")
	}
	
	// Add ingredient-based tags
	for _, ingredient := range beverage.Ingredients {
		if ingredient.Nutritional.Caffeine > 0 {
			tags = append(tags, "Caffeinated")
			break
		}
	}
	
	return tags
}

// determineTargetAudience determines who might enjoy this beverage
func (s *BeverageGeneratorService) determineTargetAudience(theme string) []string {
	switch theme {
	case "Mars Base":
		return []string{"Space Enthusiasts", "Sci-Fi Fans", "Adventurous Drinkers"}
	case "Lunar Mining Corp":
		return []string{"Workers", "Night Shift", "Energy Seekers"}
	case "Interstellar Trade Federation":
		return []string{"Business Travelers", "Luxury Seekers", "Connoisseurs"}
	default:
		return []string{"General Public", "Coffee Lovers"}
	}
}
