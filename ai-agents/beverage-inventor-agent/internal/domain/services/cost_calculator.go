package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
)

// CostCalculator provides advanced cost calculation for beverages
type CostCalculator struct {
	pricingDB      PricingDatabase
	supplierAPI    SupplierAPI
	marketAnalyzer MarketAnalyzer
}

// PricingDatabase defines the interface for pricing data
type PricingDatabase interface {
	GetIngredientPrice(ctx context.Context, ingredient string, supplier string, quantity float64, unit string) (*IngredientPrice, error)
	GetHistoricalPrices(ctx context.Context, ingredient string, days int) ([]*PricePoint, error)
	GetSeasonalPricing(ctx context.Context, ingredient string) (*SeasonalPricing, error)
	GetBulkDiscounts(ctx context.Context, ingredient string, supplier string) ([]*BulkDiscount, error)
}

// SupplierAPI defines the interface for supplier integration
type SupplierAPI interface {
	GetRealTimePricing(ctx context.Context, ingredients []string) ([]*IngredientPrice, error)
	CheckAvailability(ctx context.Context, ingredient string, quantity float64) (*AvailabilityInfo, error)
	GetShippingCost(ctx context.Context, supplier string, location string, weight float64) (float64, error)
	GetLeadTime(ctx context.Context, supplier string, ingredient string) (time.Duration, error)
}

// MarketAnalyzer defines the interface for market analysis
type MarketAnalyzer interface {
	PredictPriceTrends(ctx context.Context, ingredient string, days int) (*PriceTrend, error)
	AnalyzeMarketVolatility(ctx context.Context, ingredient string) (*VolatilityAnalysis, error)
	GetCompetitorPricing(ctx context.Context, beverageType string) (*CompetitorAnalysis, error)
}

// IngredientPrice represents the price of an ingredient
type IngredientPrice struct {
	Ingredient   string    `json:"ingredient"`
	Supplier     string    `json:"supplier"`
	Price        float64   `json:"price"`
	Unit         string    `json:"unit"`
	Currency     string    `json:"currency"`
	MinQuantity  float64   `json:"min_quantity"`
	MaxQuantity  float64   `json:"max_quantity"`
	LastUpdated  time.Time `json:"last_updated"`
	Quality      string    `json:"quality"`      // organic, premium, standard
	Certification []string `json:"certification"` // fair-trade, organic, etc.
}

// PricePoint represents a historical price point
type PricePoint struct {
	Date     time.Time `json:"date"`
	Price    float64   `json:"price"`
	Volume   float64   `json:"volume"`
	Supplier string    `json:"supplier"`
}

// SeasonalPricing represents seasonal price variations
type SeasonalPricing struct {
	Ingredient      string                    `json:"ingredient"`
	BasePrice       float64                   `json:"base_price"`
	SeasonalFactors map[string]float64        `json:"seasonal_factors"` // month -> multiplier
	PeakSeason      string                    `json:"peak_season"`
	LowSeason       string                    `json:"low_season"`
	Volatility      float64                   `json:"volatility"`
}

// BulkDiscount represents bulk pricing discounts
type BulkDiscount struct {
	MinQuantity float64 `json:"min_quantity"`
	MaxQuantity float64 `json:"max_quantity"`
	Discount    float64 `json:"discount"`    // percentage
	Unit        string  `json:"unit"`
}

// AvailabilityInfo represents ingredient availability
type AvailabilityInfo struct {
	Available    bool          `json:"available"`
	Quantity     float64       `json:"quantity"`
	LeadTime     time.Duration `json:"lead_time"`
	NextRestock  *time.Time    `json:"next_restock,omitempty"`
	Alternatives []string      `json:"alternatives"`
}

// PriceTrend represents predicted price trends
type PriceTrend struct {
	Ingredient    string        `json:"ingredient"`
	CurrentPrice  float64       `json:"current_price"`
	PredictedPrices []PricePoint `json:"predicted_prices"`
	Trend         TrendDirection `json:"trend"`
	Confidence    float64       `json:"confidence"`
	Factors       []string      `json:"factors"`
}

// TrendDirection represents price trend direction
type TrendDirection string

const (
	TrendUp    TrendDirection = "up"
	TrendDown  TrendDirection = "down"
	TrendFlat  TrendDirection = "flat"
)

// VolatilityAnalysis represents market volatility analysis
type VolatilityAnalysis struct {
	Ingredient       string  `json:"ingredient"`
	VolatilityScore  float64 `json:"volatility_score"`  // 0-100
	PriceStability   string  `json:"price_stability"`   // stable, moderate, volatile
	RiskLevel        string  `json:"risk_level"`        // low, medium, high
	RecommendedAction string `json:"recommended_action"` // buy_now, wait, hedge
}

// CompetitorAnalysis represents competitor pricing analysis
type CompetitorAnalysis struct {
	BeverageType     string             `json:"beverage_type"`
	AveragePrice     float64            `json:"average_price"`
	PriceRange       PriceRange         `json:"price_range"`
	MarketPosition   string             `json:"market_position"` // premium, mid-range, budget
	Competitors      []CompetitorPrice  `json:"competitors"`
	Recommendations  []string           `json:"recommendations"`
}

// PriceRange represents a price range
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// CompetitorPrice represents a competitor's pricing
type CompetitorPrice struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Size     float64 `json:"size"`
	PricePerUnit float64 `json:"price_per_unit"`
}

// CostCalculationRequest represents a request for cost calculation
type CostCalculationRequest struct {
	Beverage       *entities.Beverage `json:"beverage"`
	ServingSize    float64            `json:"serving_size"`    // ml
	BatchSize      float64            `json:"batch_size"`      // number of servings
	Location       string             `json:"location"`
	PreferredSuppliers []string        `json:"preferred_suppliers"`
	QualityLevel   string             `json:"quality_level"`   // standard, premium, organic
	IncludeShipping bool              `json:"include_shipping"`
	IncludeLabor   bool               `json:"include_labor"`
	IncludeOverhead bool              `json:"include_overhead"`
}

// CostBreakdown represents detailed cost breakdown
type CostBreakdown struct {
	IngredientCosts  []*IngredientCost `json:"ingredient_costs"`
	TotalIngredientCost float64        `json:"total_ingredient_cost"`
	LaborCost        float64           `json:"labor_cost"`
	OverheadCost     float64           `json:"overhead_cost"`
	ShippingCost     float64           `json:"shipping_cost"`
	PackagingCost    float64           `json:"packaging_cost"`
	TotalCost        float64           `json:"total_cost"`
	CostPerServing   float64           `json:"cost_per_serving"`
	Currency         string            `json:"currency"`
	CalculatedAt     time.Time         `json:"calculated_at"`
	Confidence       float64           `json:"confidence"`
	Warnings         []string          `json:"warnings"`
}

// IngredientCost represents the cost of a single ingredient
type IngredientCost struct {
	Ingredient   string  `json:"ingredient"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	UnitPrice    float64 `json:"unit_price"`
	TotalCost    float64 `json:"total_cost"`
	Supplier     string  `json:"supplier"`
	Quality      string  `json:"quality"`
	Availability string  `json:"availability"`
	LeadTime     string  `json:"lead_time"`
}

// ProfitabilityAnalysis represents profitability analysis
type ProfitabilityAnalysis struct {
	CostBreakdown    *CostBreakdown     `json:"cost_breakdown"`
	SuggestedPrice   float64            `json:"suggested_price"`
	ProfitMargin     float64            `json:"profit_margin"`     // percentage
	BreakEvenPrice   float64            `json:"break_even_price"`
	CompetitorAnalysis *CompetitorAnalysis `json:"competitor_analysis"`
	PricingStrategy  string             `json:"pricing_strategy"`
	Recommendations  []string           `json:"recommendations"`
}

// NewCostCalculator creates a new cost calculator
func NewCostCalculator(pricingDB PricingDatabase, supplierAPI SupplierAPI, marketAnalyzer MarketAnalyzer) *CostCalculator {
	return &CostCalculator{
		pricingDB:      pricingDB,
		supplierAPI:    supplierAPI,
		marketAnalyzer: marketAnalyzer,
	}
}

// CalculateCost calculates the total cost of a beverage
func (cc *CostCalculator) CalculateCost(ctx context.Context, req *CostCalculationRequest) (*CostBreakdown, error) {
	breakdown := &CostBreakdown{
		IngredientCosts: []*IngredientCost{},
		Currency:        "USD",
		CalculatedAt:    time.Now(),
		Warnings:        []string{},
	}
	
	// Calculate ingredient costs
	totalIngredientCost := 0.0
	for _, ingredient := range req.Beverage.Ingredients {
		cost, err := cc.calculateIngredientCost(ctx, ingredient, req)
		if err != nil {
			breakdown.Warnings = append(breakdown.Warnings, fmt.Sprintf("Failed to calculate cost for %s: %v", ingredient.Name, err))
			continue
		}
		
		breakdown.IngredientCosts = append(breakdown.IngredientCosts, cost)
		totalIngredientCost += cost.TotalCost
	}
	
	breakdown.TotalIngredientCost = totalIngredientCost
	
	// Calculate additional costs
	if req.IncludeLabor {
		breakdown.LaborCost = cc.calculateLaborCost(req.BatchSize)
	}
	
	if req.IncludeOverhead {
		breakdown.OverheadCost = cc.calculateOverheadCost(totalIngredientCost)
	}
	
	if req.IncludeShipping {
		shippingCost, err := cc.calculateShippingCost(ctx, req)
		if err != nil {
			breakdown.Warnings = append(breakdown.Warnings, fmt.Sprintf("Failed to calculate shipping cost: %v", err))
		} else {
			breakdown.ShippingCost = shippingCost
		}
	}
	
	// Calculate packaging cost (estimated)
	breakdown.PackagingCost = cc.calculatePackagingCost(req.ServingSize, req.BatchSize)
	
	// Calculate totals
	breakdown.TotalCost = breakdown.TotalIngredientCost + breakdown.LaborCost + 
		breakdown.OverheadCost + breakdown.ShippingCost + breakdown.PackagingCost
	
	if req.BatchSize > 0 {
		breakdown.CostPerServing = breakdown.TotalCost / req.BatchSize
	}
	
	// Calculate confidence
	breakdown.Confidence = cc.calculateCostConfidence(breakdown)
	
	return breakdown, nil
}

// AnalyzeProfitability analyzes the profitability of a beverage
func (cc *CostCalculator) AnalyzeProfitability(ctx context.Context, req *CostCalculationRequest, targetMargin float64) (*ProfitabilityAnalysis, error) {
	// Calculate cost breakdown
	costBreakdown, err := cc.CalculateCost(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cost: %w", err)
	}
	
	analysis := &ProfitabilityAnalysis{
		CostBreakdown: costBreakdown,
		Recommendations: []string{},
	}
	
	// Calculate suggested price based on target margin
	if targetMargin > 0 {
		analysis.SuggestedPrice = costBreakdown.CostPerServing / (1 - targetMargin/100)
		analysis.ProfitMargin = targetMargin
	} else {
		// Default 30% margin
		analysis.SuggestedPrice = costBreakdown.CostPerServing * 1.3
		analysis.ProfitMargin = 30.0
	}
	
	analysis.BreakEvenPrice = costBreakdown.CostPerServing
	
	// Get competitor analysis
	competitorAnalysis, err := cc.marketAnalyzer.GetCompetitorPricing(ctx, req.Beverage.Theme)
	if err == nil {
		analysis.CompetitorAnalysis = competitorAnalysis
		
		// Determine pricing strategy
		if analysis.SuggestedPrice < competitorAnalysis.PriceRange.Min {
			analysis.PricingStrategy = "budget"
			analysis.Recommendations = append(analysis.Recommendations, "Consider premium ingredients to justify higher pricing")
		} else if analysis.SuggestedPrice > competitorAnalysis.PriceRange.Max {
			analysis.PricingStrategy = "premium"
			analysis.Recommendations = append(analysis.Recommendations, "Ensure premium positioning is justified by quality")
		} else {
			analysis.PricingStrategy = "competitive"
			analysis.Recommendations = append(analysis.Recommendations, "Price is competitive with market")
		}
	}
	
	// Add cost optimization recommendations
	analysis.Recommendations = append(analysis.Recommendations, cc.generateCostOptimizationRecommendations(costBreakdown)...)
	
	return analysis, nil
}

// calculateIngredientCost calculates the cost of a single ingredient
func (cc *CostCalculator) calculateIngredientCost(ctx context.Context, ingredient entities.Ingredient, req *CostCalculationRequest) (*IngredientCost, error) {
	// Try preferred suppliers first
	var bestPrice *IngredientPrice
	var err error
	
	for _, supplier := range req.PreferredSuppliers {
		price, priceErr := cc.pricingDB.GetIngredientPrice(ctx, ingredient.Name, supplier, ingredient.Quantity, ingredient.Unit)
		if priceErr == nil && (bestPrice == nil || price.Price < bestPrice.Price) {
			bestPrice = price
		}
	}
	
	// If no preferred supplier found, get best available price
	if bestPrice == nil {
		bestPrice, err = cc.pricingDB.GetIngredientPrice(ctx, ingredient.Name, "", ingredient.Quantity, ingredient.Unit)
		if err != nil {
			return nil, fmt.Errorf("failed to get price for %s: %w", ingredient.Name, err)
		}
	}
	
	// Calculate total cost for the batch
	quantityNeeded := ingredient.Quantity * req.BatchSize
	totalCost := (quantityNeeded / bestPrice.MinQuantity) * bestPrice.Price
	
	// Apply bulk discounts if applicable
	discounts, err := cc.pricingDB.GetBulkDiscounts(ctx, ingredient.Name, bestPrice.Supplier)
	if err == nil {
		for _, discount := range discounts {
			if quantityNeeded >= discount.MinQuantity && quantityNeeded <= discount.MaxQuantity {
				totalCost *= (1 - discount.Discount/100)
				break
			}
		}
	}
	
	// Check availability
	availability, err := cc.supplierAPI.CheckAvailability(ctx, ingredient.Name, quantityNeeded)
	availabilityStatus := "unknown"
	leadTime := "unknown"
	if err == nil {
		if availability.Available {
			availabilityStatus = "available"
		} else {
			availabilityStatus = "limited"
		}
		leadTime = availability.LeadTime.String()
	}
	
	return &IngredientCost{
		Ingredient:   ingredient.Name,
		Quantity:     quantityNeeded,
		Unit:         ingredient.Unit,
		UnitPrice:    bestPrice.Price,
		TotalCost:    totalCost,
		Supplier:     bestPrice.Supplier,
		Quality:      bestPrice.Quality,
		Availability: availabilityStatus,
		LeadTime:     leadTime,
	}, nil
}

// calculateLaborCost calculates labor cost based on batch size
func (cc *CostCalculator) calculateLaborCost(batchSize float64) float64 {
	// Estimate 15 minutes per batch + 2 minutes per serving
	timeMinutes := 15 + (batchSize * 2)
	hourlyRate := 15.0 // $15/hour
	return (timeMinutes / 60) * hourlyRate
}

// calculateOverheadCost calculates overhead cost as percentage of ingredient cost
func (cc *CostCalculator) calculateOverheadCost(ingredientCost float64) float64 {
	// 20% overhead on ingredient cost
	return ingredientCost * 0.20
}

// calculateShippingCost calculates shipping cost
func (cc *CostCalculator) calculateShippingCost(ctx context.Context, req *CostCalculationRequest) (float64, error) {
	totalWeight := 0.0
	
	// Estimate weight based on ingredients
	for _, ingredient := range req.Beverage.Ingredients {
		// Rough estimate: 1ml = 1g for liquids, 0.5g for powders
		if ingredient.Unit == "ml" {
			totalWeight += ingredient.Quantity * req.BatchSize
		} else {
			totalWeight += ingredient.Quantity * req.BatchSize * 0.5
		}
	}
	
	// Get shipping cost from primary supplier
	if len(req.PreferredSuppliers) > 0 {
		return cc.supplierAPI.GetShippingCost(ctx, req.PreferredSuppliers[0], req.Location, totalWeight)
	}
	
	// Default shipping estimate
	return math.Max(5.0, totalWeight*0.01), nil // $5 minimum or $0.01 per gram
}

// calculatePackagingCost calculates packaging cost
func (cc *CostCalculator) calculatePackagingCost(servingSize, batchSize float64) float64 {
	// Estimate packaging cost based on serving size
	costPerServing := 0.10 // $0.10 per serving for basic packaging
	if servingSize > 500 { // Large servings need more expensive packaging
		costPerServing = 0.15
	}
	return costPerServing * batchSize
}

// calculateCostConfidence calculates confidence in cost calculation
func (cc *CostCalculator) calculateCostConfidence(breakdown *CostBreakdown) float64 {
	confidence := 0.9 // Base confidence
	
	// Reduce confidence for each warning
	confidence -= float64(len(breakdown.Warnings)) * 0.1
	
	// Reduce confidence if many ingredients have unknown availability
	unknownCount := 0
	for _, cost := range breakdown.IngredientCosts {
		if cost.Availability == "unknown" {
			unknownCount++
		}
	}
	confidence -= float64(unknownCount) * 0.05
	
	// Ensure confidence is within bounds
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}
	
	return math.Round(confidence*100) / 100
}

// generateCostOptimizationRecommendations generates cost optimization recommendations
func (cc *CostCalculator) generateCostOptimizationRecommendations(breakdown *CostBreakdown) []string {
	recommendations := []string{}
	
	// Find most expensive ingredients
	if len(breakdown.IngredientCosts) > 0 {
		maxCost := 0.0
		var expensiveIngredient *IngredientCost
		
		for _, cost := range breakdown.IngredientCosts {
			if cost.TotalCost > maxCost {
				maxCost = cost.TotalCost
				expensiveIngredient = cost
			}
		}
		
		if expensiveIngredient != nil && expensiveIngredient.TotalCost > breakdown.TotalIngredientCost*0.3 {
			recommendations = append(recommendations, fmt.Sprintf("Consider alternatives to %s (%.1f%% of ingredient cost)", 
				expensiveIngredient.Ingredient, (expensiveIngredient.TotalCost/breakdown.TotalIngredientCost)*100))
		}
	}
	
	// Check labor efficiency
	if breakdown.LaborCost > breakdown.TotalIngredientCost*0.5 {
		recommendations = append(recommendations, "High labor cost - consider batch optimization or automation")
	}
	
	// Check shipping efficiency
	if breakdown.ShippingCost > breakdown.TotalIngredientCost*0.2 {
		recommendations = append(recommendations, "High shipping cost - consider local suppliers or bulk ordering")
	}
	
	return recommendations
}
