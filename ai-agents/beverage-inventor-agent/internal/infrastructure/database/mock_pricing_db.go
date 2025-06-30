package database

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/services"
)

// MockPricingDatabase provides mock pricing data for testing and development
type MockPricingDatabase struct {
	priceData     map[string]*services.IngredientPrice
	historicalData map[string][]*services.PricePoint
	seasonalData  map[string]*services.SeasonalPricing
	bulkDiscounts map[string][]*services.BulkDiscount
}

// NewMockPricingDatabase creates a new mock pricing database
func NewMockPricingDatabase() *MockPricingDatabase {
	db := &MockPricingDatabase{
		priceData:     make(map[string]*services.IngredientPrice),
		historicalData: make(map[string][]*services.PricePoint),
		seasonalData:  make(map[string]*services.SeasonalPricing),
		bulkDiscounts: make(map[string][]*services.BulkDiscount),
	}
	
	// Initialize with common beverage ingredient prices
	db.initializePriceData()
	
	return db
}

// GetIngredientPrice returns pricing information for an ingredient
func (db *MockPricingDatabase) GetIngredientPrice(ctx context.Context, ingredient string, supplier string, quantity float64, unit string) (*services.IngredientPrice, error) {
	ingredient = strings.ToLower(ingredient)
	key := ingredient
	if supplier != "" {
		key = fmt.Sprintf("%s_%s", ingredient, strings.ToLower(supplier))
	}
	
	// Try to find specific supplier price first
	if price, exists := db.priceData[key]; exists {
		return price, nil
	}
	
	// Fall back to generic ingredient price
	if price, exists := db.priceData[ingredient]; exists {
		return price, nil
	}
	
	// Return default price for unknown ingredients
	return &services.IngredientPrice{
		Ingredient:   ingredient,
		Supplier:     supplier,
		Price:        5.00, // $5 per unit default
		Unit:         unit,
		Currency:     "USD",
		MinQuantity:  1.0,
		MaxQuantity:  1000.0,
		LastUpdated:  time.Now(),
		Quality:      "standard",
		Certification: []string{},
	}, nil
}

// GetHistoricalPrices returns historical pricing data
func (db *MockPricingDatabase) GetHistoricalPrices(ctx context.Context, ingredient string, days int) ([]*services.PricePoint, error) {
	ingredient = strings.ToLower(ingredient)
	
	if historical, exists := db.historicalData[ingredient]; exists {
		// Filter to requested number of days
		cutoff := time.Now().AddDate(0, 0, -days)
		filtered := []*services.PricePoint{}
		
		for _, point := range historical {
			if point.Date.After(cutoff) {
				filtered = append(filtered, point)
			}
		}
		
		return filtered, nil
	}
	
	// Generate synthetic historical data
	return db.generateSyntheticHistoricalData(ingredient, days), nil
}

// GetSeasonalPricing returns seasonal pricing information
func (db *MockPricingDatabase) GetSeasonalPricing(ctx context.Context, ingredient string) (*services.SeasonalPricing, error) {
	ingredient = strings.ToLower(ingredient)
	
	if seasonal, exists := db.seasonalData[ingredient]; exists {
		return seasonal, nil
	}
	
	// Return default seasonal pricing
	return &services.SeasonalPricing{
		Ingredient:      ingredient,
		BasePrice:       5.00,
		SeasonalFactors: map[string]float64{
			"january": 1.0, "february": 1.0, "march": 1.0,
			"april": 1.0, "may": 1.0, "june": 1.0,
			"july": 1.0, "august": 1.0, "september": 1.0,
			"october": 1.0, "november": 1.0, "december": 1.0,
		},
		PeakSeason:      "year-round",
		LowSeason:       "none",
		Volatility:      0.1,
	}, nil
}

// GetBulkDiscounts returns bulk discount information
func (db *MockPricingDatabase) GetBulkDiscounts(ctx context.Context, ingredient string, supplier string) ([]*services.BulkDiscount, error) {
	ingredient = strings.ToLower(ingredient)
	key := ingredient
	if supplier != "" {
		key = fmt.Sprintf("%s_%s", ingredient, strings.ToLower(supplier))
	}
	
	if discounts, exists := db.bulkDiscounts[key]; exists {
		return discounts, nil
	}
	
	// Return default bulk discounts
	return []*services.BulkDiscount{
		{MinQuantity: 10, MaxQuantity: 50, Discount: 5.0, Unit: "units"},
		{MinQuantity: 50, MaxQuantity: 100, Discount: 10.0, Unit: "units"},
		{MinQuantity: 100, MaxQuantity: 500, Discount: 15.0, Unit: "units"},
		{MinQuantity: 500, MaxQuantity: 1000, Discount: 20.0, Unit: "units"},
	}, nil
}

// generateSyntheticHistoricalData generates synthetic historical pricing data
func (db *MockPricingDatabase) generateSyntheticHistoricalData(ingredient string, days int) []*services.PricePoint {
	basePrice := 5.00
	if price, exists := db.priceData[ingredient]; exists {
		basePrice = price.Price
	}
	
	points := make([]*services.PricePoint, days)
	
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -days+i)
		
		// Add some random variation (±10%)
		variation := (math.Sin(float64(i)*0.1) * 0.1) + (math.Cos(float64(i)*0.05) * 0.05)
		price := basePrice * (1.0 + variation)
		
		points[i] = &services.PricePoint{
			Date:     date,
			Price:    price,
			Volume:   100.0 + float64(i%50), // Synthetic volume
			Supplier: "default",
		}
	}
	
	return points
}

// initializePriceData initializes the mock database with common ingredient prices
func (db *MockPricingDatabase) initializePriceData() {
	// Coffee and tea prices (per 100g)
	db.priceData["coffee"] = &services.IngredientPrice{
		Ingredient: "coffee", Supplier: "premium_beans", Price: 12.00, Unit: "100g",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 1000, LastUpdated: time.Now(),
		Quality: "premium", Certification: []string{"fair-trade", "organic"},
	}
	
	db.priceData["coffee_budget_roasters"] = &services.IngredientPrice{
		Ingredient: "coffee", Supplier: "budget_roasters", Price: 6.00, Unit: "100g",
		Currency: "USD", MinQuantity: 5, MaxQuantity: 500, LastUpdated: time.Now(),
		Quality: "standard", Certification: []string{},
	}
	
	db.priceData["espresso"] = &services.IngredientPrice{
		Ingredient: "espresso", Supplier: "italian_imports", Price: 15.00, Unit: "100g",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 200, LastUpdated: time.Now(),
		Quality: "premium", Certification: []string{"italian-certified"},
	}
	
	db.priceData["green tea"] = &services.IngredientPrice{
		Ingredient: "green tea", Supplier: "tea_masters", Price: 8.00, Unit: "100g",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 500, LastUpdated: time.Now(),
		Quality: "premium", Certification: []string{"organic"},
	}
	
	db.priceData["black tea"] = &services.IngredientPrice{
		Ingredient: "black tea", Supplier: "tea_masters", Price: 6.00, Unit: "100g",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 500, LastUpdated: time.Now(),
		Quality: "standard", Certification: []string{},
	}
	
	// Milk and dairy prices (per liter)
	db.priceData["whole milk"] = &services.IngredientPrice{
		Ingredient: "whole milk", Supplier: "local_dairy", Price: 3.50, Unit: "liter",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 100, LastUpdated: time.Now(),
		Quality: "standard", Certification: []string{"pasteurized"},
	}
	
	db.priceData["almond milk"] = &services.IngredientPrice{
		Ingredient: "almond milk", Supplier: "plant_based_co", Price: 4.50, Unit: "liter",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 50, LastUpdated: time.Now(),
		Quality: "premium", Certification: []string{"organic", "non-gmo"},
	}
	
	db.priceData["oat milk"] = &services.IngredientPrice{
		Ingredient: "oat milk", Supplier: "plant_based_co", Price: 4.00, Unit: "liter",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 50, LastUpdated: time.Now(),
		Quality: "standard", Certification: []string{"organic"},
	}
	
	// Sweetener prices
	db.priceData["sugar"] = &services.IngredientPrice{
		Ingredient: "sugar", Supplier: "sweet_supply", Price: 2.00, Unit: "kg",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 1000, LastUpdated: time.Now(),
		Quality: "standard", Certification: []string{},
	}
	
	db.priceData["honey"] = &services.IngredientPrice{
		Ingredient: "honey", Supplier: "local_beekeepers", Price: 12.00, Unit: "kg",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 100, LastUpdated: time.Now(),
		Quality: "premium", Certification: []string{"raw", "local"},
	}
	
	db.priceData["stevia"] = &services.IngredientPrice{
		Ingredient: "stevia", Supplier: "natural_sweeteners", Price: 25.00, Unit: "kg",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 50, LastUpdated: time.Now(),
		Quality: "premium", Certification: []string{"organic", "natural"},
	}
	
	// Spice and flavoring prices (per 100g)
	db.priceData["cinnamon"] = &services.IngredientPrice{
		Ingredient: "cinnamon", Supplier: "spice_world", Price: 8.00, Unit: "100g",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 100, LastUpdated: time.Now(),
		Quality: "premium", Certification: []string{"organic", "ceylon"},
	}
	
	db.priceData["vanilla"] = &services.IngredientPrice{
		Ingredient: "vanilla", Supplier: "flavor_extracts", Price: 15.00, Unit: "100ml",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 50, LastUpdated: time.Now(),
		Quality: "premium", Certification: []string{"pure", "madagascar"},
	}
	
	// Fruit prices
	db.priceData["lemon juice"] = &services.IngredientPrice{
		Ingredient: "lemon juice", Supplier: "fresh_citrus", Price: 6.00, Unit: "liter",
		Currency: "USD", MinQuantity: 1, MaxQuantity: 100, LastUpdated: time.Now(),
		Quality: "standard", Certification: []string{"fresh-squeezed"},
	}
	
	// Seasonal pricing for coffee
	db.seasonalData["coffee"] = &services.SeasonalPricing{
		Ingredient: "coffee",
		BasePrice:  12.00,
		SeasonalFactors: map[string]float64{
			"january": 1.1, "february": 1.1, "march": 1.0,
			"april": 0.95, "may": 0.9, "june": 0.9,
			"july": 0.95, "august": 1.0, "september": 1.05,
			"october": 1.1, "november": 1.15, "december": 1.2,
		},
		PeakSeason: "november-february",
		LowSeason:  "may-july",
		Volatility: 0.15,
	}
	
	// Bulk discounts for coffee
	db.bulkDiscounts["coffee_premium_beans"] = []*services.BulkDiscount{
		{MinQuantity: 5, MaxQuantity: 20, Discount: 5.0, Unit: "kg"},
		{MinQuantity: 20, MaxQuantity: 50, Discount: 10.0, Unit: "kg"},
		{MinQuantity: 50, MaxQuantity: 100, Discount: 15.0, Unit: "kg"},
		{MinQuantity: 100, MaxQuantity: 500, Discount: 20.0, Unit: "kg"},
	}
	
	// Historical data for coffee (last 30 days)
	db.historicalData["coffee"] = db.generateSyntheticHistoricalData("coffee", 30)
}
