package database

import (
	"context"
	"strings"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/services"
)

// MockIngredientKnowledgeBase provides mock ingredient knowledge for testing and development
type MockIngredientKnowledgeBase struct {
	profiles           map[string]*services.IngredientProfile
	compatibilityRules []*services.CompatibilityRule
	substitutions      map[string][]*services.Substitution
	flavorProfiles     map[string]*services.FlavorProfile
	chemicalProps      map[string]*services.ChemicalProperties
}

// NewMockIngredientKnowledgeBase creates a new mock ingredient knowledge base
func NewMockIngredientKnowledgeBase() *MockIngredientKnowledgeBase {
	kb := &MockIngredientKnowledgeBase{
		profiles:           make(map[string]*services.IngredientProfile),
		compatibilityRules: []*services.CompatibilityRule{},
		substitutions:      make(map[string][]*services.Substitution),
		flavorProfiles:     make(map[string]*services.FlavorProfile),
		chemicalProps:      make(map[string]*services.ChemicalProperties),
	}

	// Initialize with common beverage ingredient knowledge
	kb.initializeKnowledgeBase()

	return kb
}

// GetIngredientProfile returns comprehensive ingredient information
func (kb *MockIngredientKnowledgeBase) GetIngredientProfile(ctx context.Context, ingredient string) (*services.IngredientProfile, error) {
	ingredient = strings.ToLower(ingredient)

	if profile, exists := kb.profiles[ingredient]; exists {
		return profile, nil
	}

	// Return default profile for unknown ingredients
	return &services.IngredientProfile{
		Name:     ingredient,
		Category: "unknown",
		FlavorProfile: &services.FlavorProfile{
			Primary:     []string{"neutral"},
			Secondary:   []string{},
			Intensity:   5.0,
			Sweetness:   3.0,
			Acidity:     3.0,
			Bitterness:  3.0,
			Saltiness:   1.0,
			Umami:       1.0,
			Astringency: 2.0,
			Aromatics:   []string{"mild"},
			Mouthfeel:   "neutral",
			AfterTaste:  "clean",
		},
		ChemicalProps: &services.ChemicalProperties{
			PH:             6.5, // Slightly acidic, more typical for beverage ingredients
			Solubility:     "water",
			Stability:      "stable",
			ReactiveGroups: []string{},
			Preservatives:  []string{},
			Emulsifiers:    []string{},
			Antioxidants:   []string{},
		},
		Allergens:     []string{},
		Substitutions: []*services.Substitution{},
		StorageReqs: &services.StorageRequirements{
			Temperature: "room",
			Humidity:    "low",
			Light:       "dark",
			ShelfLife:   "6 months",
			Container:   "airtight",
		},
		SeasonalInfo: &services.SeasonalInfo{
			PeakSeason:       []string{"year-round"},
			OffSeason:        []string{},
			QualityVariation: "minimal",
			PriceVariation:   "stable",
		},
	}, nil
}

// GetCompatibilityRules returns ingredient compatibility rules
func (kb *MockIngredientKnowledgeBase) GetCompatibilityRules(ctx context.Context) ([]*services.CompatibilityRule, error) {
	return kb.compatibilityRules, nil
}

// GetSubstitutions returns possible substitutions for an ingredient
func (kb *MockIngredientKnowledgeBase) GetSubstitutions(ctx context.Context, ingredient string) ([]*services.Substitution, error) {
	ingredient = strings.ToLower(ingredient)

	if substitutions, exists := kb.substitutions[ingredient]; exists {
		return substitutions, nil
	}

	return []*services.Substitution{}, nil
}

// GetFlavorProfile returns flavor profile for an ingredient
func (kb *MockIngredientKnowledgeBase) GetFlavorProfile(ctx context.Context, ingredient string) (*services.FlavorProfile, error) {
	ingredient = strings.ToLower(ingredient)

	if profile, exists := kb.flavorProfiles[ingredient]; exists {
		return profile, nil
	}

	// Return default flavor profile
	return &services.FlavorProfile{
		Primary:     []string{"neutral"},
		Secondary:   []string{},
		Intensity:   5.0,
		Sweetness:   3.0,
		Acidity:     3.0,
		Bitterness:  3.0,
		Saltiness:   1.0,
		Umami:       1.0,
		Astringency: 2.0,
		Aromatics:   []string{"mild"},
		Mouthfeel:   "neutral",
		AfterTaste:  "clean",
	}, nil
}

// GetChemicalProperties returns chemical properties for an ingredient
func (kb *MockIngredientKnowledgeBase) GetChemicalProperties(ctx context.Context, ingredient string) (*services.ChemicalProperties, error) {
	ingredient = strings.ToLower(ingredient)

	if props, exists := kb.chemicalProps[ingredient]; exists {
		return props, nil
	}

	// Return default chemical properties with more realistic values
	return &services.ChemicalProperties{
		PH:             6.5, // Slightly acidic, more typical for beverage ingredients
		Solubility:     "water",
		Stability:      "stable",
		ReactiveGroups: []string{},
		Preservatives:  []string{},
		Emulsifiers:    []string{},
		Antioxidants:   []string{},
	}, nil
}

// initializeKnowledgeBase initializes the knowledge base with common ingredient data
func (kb *MockIngredientKnowledgeBase) initializeKnowledgeBase() {
	// Coffee flavor profile
	kb.flavorProfiles["coffee"] = &services.FlavorProfile{
		Primary:     []string{"bitter", "roasted", "earthy"},
		Secondary:   []string{"nutty", "chocolate", "caramel"},
		Intensity:   8.0,
		Sweetness:   2.0,
		Acidity:     6.0,
		Bitterness:  8.0,
		Saltiness:   1.0,
		Umami:       3.0,
		Astringency: 4.0,
		Aromatics:   []string{"roasted", "smoky", "rich"},
		Mouthfeel:   "full-bodied",
		AfterTaste:  "lingering bitter",
	}

	// Milk flavor profile
	kb.flavorProfiles["whole milk"] = &services.FlavorProfile{
		Primary:     []string{"creamy", "sweet", "mild"},
		Secondary:   []string{"rich", "smooth"},
		Intensity:   4.0,
		Sweetness:   6.0,
		Acidity:     2.0,
		Bitterness:  1.0,
		Saltiness:   2.0,
		Umami:       2.0,
		Astringency: 1.0,
		Aromatics:   []string{"fresh", "dairy"},
		Mouthfeel:   "creamy",
		AfterTaste:  "clean",
	}

	// Cinnamon flavor profile
	kb.flavorProfiles["cinnamon"] = &services.FlavorProfile{
		Primary:     []string{"sweet", "spicy", "warm"},
		Secondary:   []string{"woody", "aromatic"},
		Intensity:   9.0,
		Sweetness:   7.0,
		Acidity:     2.0,
		Bitterness:  3.0,
		Saltiness:   1.0,
		Umami:       1.0,
		Astringency: 5.0,
		Aromatics:   []string{"warm spice", "sweet", "woody"},
		Mouthfeel:   "warming",
		AfterTaste:  "spicy warmth",
	}

	// Chemical properties
	kb.chemicalProps["coffee"] = &services.ChemicalProperties{
		PH:             5.0,
		Solubility:     "water",
		Stability:      "heat stable",
		ReactiveGroups: []string{"caffeine", "chlorogenic acids"},
		Preservatives:  []string{},
		Emulsifiers:    []string{},
		Antioxidants:   []string{"chlorogenic acid", "caffeic acid"},
	}

	kb.chemicalProps["whole milk"] = &services.ChemicalProperties{
		PH:             6.7,
		Solubility:     "water",
		Stability:      "heat sensitive",
		ReactiveGroups: []string{"proteins", "lactose"},
		Preservatives:  []string{},
		Emulsifiers:    []string{"casein proteins"},
		Antioxidants:   []string{"vitamin E"},
	}

	// Compatibility rules
	kb.compatibilityRules = []*services.CompatibilityRule{
		{
			ID:          "coffee_milk_positive",
			Type:        "positive",
			Ingredient1: "coffee",
			Ingredient2: "whole milk",
			Reason:      "Classic combination - milk mellows coffee's bitterness and adds creaminess",
			Confidence:  0.95,
			Source:      "culinary tradition",
		},
		{
			ID:          "coffee_cinnamon_positive",
			Type:        "positive",
			Ingredient1: "coffee",
			Ingredient2: "cinnamon",
			Reason:      "Cinnamon complements coffee's roasted notes and adds warmth",
			Confidence:  0.85,
			Source:      "culinary tradition",
		},
		{
			ID:          "milk_cinnamon_positive",
			Type:        "positive",
			Ingredient1: "whole milk",
			Ingredient2: "cinnamon",
			Reason:      "Cinnamon pairs well with dairy, creating warm, comforting flavors",
			Confidence:  0.80,
			Source:      "culinary tradition",
		},
		{
			ID:          "coffee_lemon_negative",
			Type:        "negative",
			Ingredient1: "coffee",
			Ingredient2: "lemon juice",
			Reason:      "High acidity can curdle milk and create unpleasant sourness",
			Confidence:  0.75,
			Source:      "chemical incompatibility",
		},
	}

	// Substitutions
	kb.substitutions["whole milk"] = []*services.Substitution{
		{
			Original:          "whole milk",
			Substitute:        "almond milk",
			Ratio:             1.0,
			FlavorImpact:      "Nuttier, less creamy",
			TextureImpact:     "Thinner consistency",
			NutritionalImpact: "Lower calories, less protein",
			Confidence:        0.85,
			Notes:             "Good dairy-free alternative",
		},
		{
			Original:          "whole milk",
			Substitute:        "oat milk",
			Ratio:             1.0,
			FlavorImpact:      "Slightly sweet, oaty flavor",
			TextureImpact:     "Similar creaminess",
			NutritionalImpact: "Lower protein, more fiber",
			Confidence:        0.90,
			Notes:             "Best texture match for dairy milk",
		},
	}

	kb.substitutions["sugar"] = []*services.Substitution{
		{
			Original:          "sugar",
			Substitute:        "honey",
			Ratio:             0.75,
			FlavorImpact:      "Floral, more complex sweetness",
			TextureImpact:     "Slightly thicker",
			NutritionalImpact: "More minerals, similar calories",
			Confidence:        0.80,
			Notes:             "Natural alternative with unique flavor",
		},
		{
			Original:          "sugar",
			Substitute:        "stevia",
			Ratio:             0.1,
			FlavorImpact:      "Very sweet, slight aftertaste",
			TextureImpact:     "No texture impact",
			NutritionalImpact: "Zero calories",
			Confidence:        0.70,
			Notes:             "Calorie-free but may have aftertaste",
		},
	}

	// Add more ingredient profiles for better coverage
	kb.addCommonIngredients()
}

// addCommonIngredients adds profiles for common beverage ingredients
func (kb *MockIngredientKnowledgeBase) addCommonIngredients() {
	// Espresso
	kb.flavorProfiles["espresso"] = &services.FlavorProfile{
		Primary:     []string{"bitter", "intense", "roasted"},
		Secondary:   []string{"chocolate", "caramel", "nutty"},
		Intensity:   9.0,
		Sweetness:   1.0,
		Acidity:     7.0,
		Bitterness:  9.0,
		Saltiness:   1.0,
		Umami:       4.0,
		Astringency: 5.0,
		Aromatics:   []string{"intense roasted", "rich", "bold"},
		Mouthfeel:   "full-bodied",
		AfterTaste:  "strong bitter",
	}

	kb.chemicalProps["espresso"] = &services.ChemicalProperties{
		PH:             4.8,
		Solubility:     "water",
		Stability:      "heat stable",
		ReactiveGroups: []string{"caffeine", "chlorogenic acids", "quinides"},
		Preservatives:  []string{},
		Emulsifiers:    []string{},
		Antioxidants:   []string{"chlorogenic acid", "caffeic acid", "melanoidins"},
	}

	// Sugar
	kb.flavorProfiles["sugar"] = &services.FlavorProfile{
		Primary:     []string{"sweet"},
		Secondary:   []string{},
		Intensity:   8.0,
		Sweetness:   10.0,
		Acidity:     1.0,
		Bitterness:  1.0,
		Saltiness:   1.0,
		Umami:       1.0,
		Astringency: 1.0,
		Aromatics:   []string{"neutral"},
		Mouthfeel:   "clean",
		AfterTaste:  "sweet",
	}

	kb.chemicalProps["sugar"] = &services.ChemicalProperties{
		PH:             7.0, // Pure sugar is neutral
		Solubility:     "water",
		Stability:      "stable",
		ReactiveGroups: []string{"sucrose"},
		Preservatives:  []string{},
		Emulsifiers:    []string{},
		Antioxidants:   []string{},
	}

	// Vanilla
	kb.flavorProfiles["vanilla"] = &services.FlavorProfile{
		Primary:     []string{"sweet", "creamy", "floral"},
		Secondary:   []string{"woody", "spicy"},
		Intensity:   7.0,
		Sweetness:   8.0,
		Acidity:     2.0,
		Bitterness:  2.0,
		Saltiness:   1.0,
		Umami:       1.0,
		Astringency: 2.0,
		Aromatics:   []string{"sweet", "floral", "warm"},
		Mouthfeel:   "smooth",
		AfterTaste:  "sweet lingering",
	}

	kb.chemicalProps["vanilla"] = &services.ChemicalProperties{
		PH:             6.2,
		Solubility:     "alcohol/water",
		Stability:      "heat stable",
		ReactiveGroups: []string{"vanillin", "vanillic acid"},
		Preservatives:  []string{},
		Emulsifiers:    []string{},
		Antioxidants:   []string{"vanillic acid"},
	}

	// Lemon juice
	kb.flavorProfiles["lemon juice"] = &services.FlavorProfile{
		Primary:     []string{"sour", "citrus", "bright"},
		Secondary:   []string{"fresh", "zesty"},
		Intensity:   8.0,
		Sweetness:   2.0,
		Acidity:     9.0,
		Bitterness:  3.0,
		Saltiness:   1.0,
		Umami:       1.0,
		Astringency: 4.0,
		Aromatics:   []string{"citrus", "fresh", "bright"},
		Mouthfeel:   "sharp",
		AfterTaste:  "tart",
	}

	kb.chemicalProps["lemon juice"] = &services.ChemicalProperties{
		PH:             2.3,
		Solubility:     "water",
		Stability:      "vitamin C degrades",
		ReactiveGroups: []string{"citric acid", "ascorbic acid"},
		Preservatives:  []string{"citric acid"},
		Emulsifiers:    []string{},
		Antioxidants:   []string{"vitamin C", "limonene"},
	}

	// Add compatibility rules for new ingredients
	kb.compatibilityRules = append(kb.compatibilityRules, []*services.CompatibilityRule{
		{
			ID:          "espresso_vanilla_positive",
			Type:        "positive",
			Ingredient1: "espresso",
			Ingredient2: "vanilla",
			Reason:      "Vanilla enhances coffee's sweet notes and reduces harsh bitterness",
			Confidence:  0.88,
			Source:      "culinary tradition",
		},
		{
			ID:          "milk_lemon_negative",
			Type:        "negative",
			Ingredient1: "whole milk",
			Ingredient2: "lemon juice",
			Reason:      "Acid causes milk proteins to coagulate and curdle",
			Confidence:  0.95,
			Source:      "chemical incompatibility",
		},
		{
			ID:          "sugar_cinnamon_positive",
			Type:        "positive",
			Ingredient1: "sugar",
			Ingredient2: "cinnamon",
			Reason:      "Sugar balances cinnamon's spiciness and enhances warm flavors",
			Confidence:  0.82,
			Source:      "culinary tradition",
		},
	}...)

	// Add complete profiles
	kb.profiles["espresso"] = &services.IngredientProfile{
		Name:            "espresso",
		Category:        "beverage base",
		FlavorProfile:   kb.flavorProfiles["espresso"],
		ChemicalProps:   kb.chemicalProps["espresso"],
		NutritionalInfo: nil,
		Allergens:       []string{},
		Substitutions:   []*services.Substitution{},
		StorageReqs: &services.StorageRequirements{
			Temperature: "room",
			Humidity:    "low",
			Light:       "dark",
			ShelfLife:   "1-2 weeks (ground), 1 month (beans)",
			Container:   "airtight",
		},
		SeasonalInfo: &services.SeasonalInfo{
			PeakSeason:       []string{"year-round"},
			OffSeason:        []string{},
			QualityVariation: "origin dependent",
			PriceVariation:   "moderate seasonal variation",
		},
	}

	kb.profiles["sugar"] = &services.IngredientProfile{
		Name:            "sugar",
		Category:        "sweetener",
		FlavorProfile:   kb.flavorProfiles["sugar"],
		ChemicalProps:   kb.chemicalProps["sugar"],
		NutritionalInfo: nil,
		Allergens:       []string{},
		Substitutions:   kb.substitutions["sugar"],
		StorageReqs: &services.StorageRequirements{
			Temperature: "room",
			Humidity:    "low",
			Light:       "any",
			ShelfLife:   "indefinite",
			Container:   "airtight",
		},
		SeasonalInfo: &services.SeasonalInfo{
			PeakSeason:       []string{"year-round"},
			OffSeason:        []string{},
			QualityVariation: "minimal",
			PriceVariation:   "stable",
		},
	}
}
