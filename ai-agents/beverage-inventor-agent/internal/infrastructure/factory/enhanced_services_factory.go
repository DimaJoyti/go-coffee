package factory

import (
	"context"
	"math"
	"strings"
	"time"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/services"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/ai"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/database"
	aiManager "go-coffee-ai-agents/internal/ai"
)

// Logger interface for the factory
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// EnhancedServicesFactory creates and wires up all enhanced beverage services
type EnhancedServicesFactory struct {
	logger Logger
}

// NewEnhancedServicesFactory creates a new enhanced services factory
func NewEnhancedServicesFactory(logger Logger) *EnhancedServicesFactory {
	return &EnhancedServicesFactory{
		logger: logger,
	}
}

// EnhancedServices contains all the enhanced beverage services
type EnhancedServices struct {
	NutritionalAnalyzer   *services.NutritionalAnalyzer
	CostCalculator        *services.CostCalculator
	CompatibilityAnalyzer *services.IngredientCompatibilityAnalyzer
	RecipeOptimizer       *services.RecipeOptimizer
	BeverageAIProvider    *ai.BeverageAIProvider
}

// CreateEnhancedServices creates and wires up all enhanced services
func (f *EnhancedServicesFactory) CreateEnhancedServices(aiMgr *aiManager.Manager) *EnhancedServices {
	f.logger.Info("Creating enhanced beverage services")

	// Create mock databases
	nutritionDB := database.NewMockNutritionDatabase()
	pricingDB := database.NewMockPricingDatabase()
	ingredientKB := database.NewMockIngredientKnowledgeBase()

	// Create AI provider
	beverageAI := ai.NewBeverageAIProvider(aiMgr, f.logger)

	// Create mock supplier API and market analyzer
	supplierAPI := NewMockSupplierAPI()
	marketAnalyzer := NewMockMarketAnalyzer()

	// Create nutritional analyzer
	nutritionalAnalyzer := services.NewNutritionalAnalyzer(nutritionDB, beverageAI)

	// Create cost calculator
	costCalculator := services.NewCostCalculator(pricingDB, supplierAPI, marketAnalyzer)

	// Create compatibility analyzer
	compatibilityAnalyzer := services.NewIngredientCompatibilityAnalyzer(ingredientKB, beverageAI, NewMockFlavorDatabase())

	// Create recipe optimizer
	recipeOptimizer := services.NewRecipeOptimizer(
		nutritionalAnalyzer,
		costCalculator,
		compatibilityAnalyzer,
		beverageAI,
	)

	f.logger.Info("Enhanced beverage services created successfully")

	return &EnhancedServices{
		NutritionalAnalyzer:   nutritionalAnalyzer,
		CostCalculator:        costCalculator,
		CompatibilityAnalyzer: compatibilityAnalyzer,
		RecipeOptimizer:       recipeOptimizer,
		BeverageAIProvider:    beverageAI,
	}
}

// MockSupplierAPI provides mock supplier API functionality
type MockSupplierAPI struct{}

// NewMockSupplierAPI creates a new mock supplier API
func NewMockSupplierAPI() *MockSupplierAPI {
	return &MockSupplierAPI{}
}

// GetRealTimePricing implements the SupplierAPI interface
func (api *MockSupplierAPI) GetRealTimePricing(ctx context.Context, ingredients []string) ([]*services.IngredientPrice, error) {
	prices := make([]*services.IngredientPrice, len(ingredients))

	for i, ingredient := range ingredients {
		prices[i] = &services.IngredientPrice{
			Ingredient:    ingredient,
			Supplier:      "mock_supplier",
			Price:         5.00 + float64(i), // Varying prices
			Unit:          "unit",
			Currency:      "USD",
			MinQuantity:   1.0,
			MaxQuantity:   1000.0,
			LastUpdated:   time.Now(),
			Quality:       "standard",
			Certification: []string{},
		}
	}

	return prices, nil
}

// CheckAvailability implements the SupplierAPI interface
func (api *MockSupplierAPI) CheckAvailability(ctx context.Context, ingredient string, quantity float64) (*services.AvailabilityInfo, error) {
	return &services.AvailabilityInfo{
		Available:    true,
		Quantity:     quantity * 10,  // Mock available quantity
		LeadTime:     time.Hour * 24, // 1 day lead time
		NextRestock:  nil,
		Alternatives: []string{ingredient + "_alternative"},
	}, nil
}

// GetShippingCost implements the SupplierAPI interface
func (api *MockSupplierAPI) GetShippingCost(ctx context.Context, supplier string, location string, weight float64) (float64, error) {
	// Simple shipping cost calculation
	baseCost := 5.0
	weightCost := weight * 0.01 // $0.01 per gram
	return baseCost + weightCost, nil
}

// GetLeadTime implements the SupplierAPI interface
func (api *MockSupplierAPI) GetLeadTime(ctx context.Context, supplier string, ingredient string) (time.Duration, error) {
	// Mock lead times based on ingredient type
	switch {
	case strings.Contains(strings.ToLower(ingredient), "coffee"):
		return time.Hour * 48, nil // 2 days for coffee
	case strings.Contains(strings.ToLower(ingredient), "milk"):
		return time.Hour * 24, nil // 1 day for dairy
	default:
		return time.Hour * 72, nil // 3 days default
	}
}

// MockMarketAnalyzer provides mock market analysis functionality
type MockMarketAnalyzer struct{}

// NewMockMarketAnalyzer creates a new mock market analyzer
func NewMockMarketAnalyzer() *MockMarketAnalyzer {
	return &MockMarketAnalyzer{}
}

// PredictPriceTrends implements the MarketAnalyzer interface
func (ma *MockMarketAnalyzer) PredictPriceTrends(ctx context.Context, ingredient string, days int) (*services.PriceTrend, error) {
	currentPrice := 5.0 + float64(len(ingredient)%10) // Mock price based on ingredient name

	// Generate mock predicted prices
	predictedPrices := make([]services.PricePoint, days)
	for i := 0; i < days; i++ {
		variation := math.Sin(float64(i)*0.1) * 0.1 // Small price variations
		predictedPrices[i] = services.PricePoint{
			Date:     time.Now().AddDate(0, 0, i+1),
			Price:    currentPrice * (1.0 + variation),
			Volume:   100.0,
			Supplier: "mock_supplier",
		}
	}

	return &services.PriceTrend{
		Ingredient:      ingredient,
		CurrentPrice:    currentPrice,
		PredictedPrices: predictedPrices,
		Trend:           services.TrendFlat,
		Confidence:      0.75,
		Factors:         []string{"seasonal variation", "market demand"},
	}, nil
}

// AnalyzeMarketVolatility implements the MarketAnalyzer interface
func (ma *MockMarketAnalyzer) AnalyzeMarketVolatility(ctx context.Context, ingredient string) (*services.VolatilityAnalysis, error) {
	// Mock volatility based on ingredient type
	var volatilityScore float64
	var stability, riskLevel, action string

	switch {
	case strings.Contains(strings.ToLower(ingredient), "coffee"):
		volatilityScore = 65.0
		stability = "moderate"
		riskLevel = "medium"
		action = "monitor"
	case strings.Contains(strings.ToLower(ingredient), "milk"):
		volatilityScore = 25.0
		stability = "stable"
		riskLevel = "low"
		action = "buy_now"
	default:
		volatilityScore = 45.0
		stability = "moderate"
		riskLevel = "medium"
		action = "wait"
	}

	return &services.VolatilityAnalysis{
		Ingredient:        ingredient,
		VolatilityScore:   volatilityScore,
		PriceStability:    stability,
		RiskLevel:         riskLevel,
		RecommendedAction: action,
	}, nil
}

// GetCompetitorPricing implements the MarketAnalyzer interface
func (ma *MockMarketAnalyzer) GetCompetitorPricing(ctx context.Context, beverageType string) (*services.CompetitorAnalysis, error) {
	// Mock competitor analysis based on beverage type
	var avgPrice float64
	var priceRange services.PriceRange
	var position string

	switch strings.ToLower(beverageType) {
	case "coffee", "espresso":
		avgPrice = 4.50
		priceRange = services.PriceRange{Min: 2.50, Max: 8.00}
		position = "mid-range"
	case "tea":
		avgPrice = 3.00
		priceRange = services.PriceRange{Min: 1.50, Max: 6.00}
		position = "budget"
	default:
		avgPrice = 3.50
		priceRange = services.PriceRange{Min: 2.00, Max: 7.00}
		position = "competitive"
	}

	competitors := []services.CompetitorPrice{
		{Name: "Competitor A", Price: avgPrice * 0.8, Size: 350, PricePerUnit: (avgPrice * 0.8) / 350},
		{Name: "Competitor B", Price: avgPrice, Size: 350, PricePerUnit: avgPrice / 350},
		{Name: "Competitor C", Price: avgPrice * 1.2, Size: 350, PricePerUnit: (avgPrice * 1.2) / 350},
	}

	return &services.CompetitorAnalysis{
		BeverageType:    beverageType,
		AveragePrice:    avgPrice,
		PriceRange:      priceRange,
		MarketPosition:  position,
		Competitors:     competitors,
		Recommendations: []string{"Consider premium positioning", "Focus on quality differentiation"},
	}, nil
}

// MockFlavorDatabase provides mock flavor database functionality
type MockFlavorDatabase struct{}

// NewMockFlavorDatabase creates a new mock flavor database
func NewMockFlavorDatabase() *MockFlavorDatabase {
	return &MockFlavorDatabase{}
}

// GetFlavorCompounds implements the FlavorDatabase interface
func (db *MockFlavorDatabase) GetFlavorCompounds(ctx context.Context, ingredient string) ([]string, error) {
	// Mock flavor compounds based on ingredient
	switch strings.ToLower(ingredient) {
	case "coffee":
		return []string{"caffeine", "chlorogenic acid", "quinides", "furans"}, nil
	case "cinnamon":
		return []string{"cinnamaldehyde", "eugenol", "coumarin"}, nil
	case "vanilla":
		return []string{"vanillin", "vanillic acid", "4-hydroxybenzaldehyde"}, nil
	default:
		return []string{"generic_compound"}, nil
	}
}

// GetFlavorIntensity implements the FlavorDatabase interface
func (db *MockFlavorDatabase) GetFlavorIntensity(ctx context.Context, ingredient string) (float64, error) {
	// Mock intensity based on ingredient
	switch strings.ToLower(ingredient) {
	case "coffee", "espresso":
		return 8.0, nil
	case "cinnamon":
		return 9.0, nil
	case "vanilla":
		return 6.0, nil
	case "milk":
		return 3.0, nil
	default:
		return 5.0, nil
	}
}

// GetComplementaryFlavors implements the FlavorDatabase interface
func (db *MockFlavorDatabase) GetComplementaryFlavors(ctx context.Context, ingredient string) ([]string, error) {
	// Mock complementary flavors
	switch strings.ToLower(ingredient) {
	case "coffee":
		return []string{"chocolate", "caramel", "vanilla", "cinnamon"}, nil
	case "cinnamon":
		return []string{"apple", "vanilla", "chocolate", "coffee"}, nil
	case "vanilla":
		return []string{"chocolate", "coffee", "caramel", "fruit"}, nil
	default:
		return []string{"neutral"}, nil
	}
}

// GetConflictingFlavors implements the FlavorDatabase interface
func (db *MockFlavorDatabase) GetConflictingFlavors(ctx context.Context, ingredient string) ([]string, error) {
	// Mock conflicting flavors
	switch strings.ToLower(ingredient) {
	case "coffee":
		return []string{"citrus", "fish", "strong herbs"}, nil
	case "milk":
		return []string{"high acid", "citrus", "wine"}, nil
	default:
		return []string{}, nil
	}
}
