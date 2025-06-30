package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/entities"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/events"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/repositories"
)

// AutoReorderService provides intelligent automatic reordering capabilities
type AutoReorderService struct {
	inventoryRepo     repositories.InventoryRepository
	supplierRepo      repositories.SupplierRepository
	purchaseOrderRepo repositories.PurchaseOrderRepository
	demandForecastSvc *DemandForecastingService
	eventPublisher    EventPublisher
	logger            Logger
}

// ReorderRecommendation represents a recommendation to reorder an item
type ReorderRecommendation struct {
	ItemID            uuid.UUID                `json:"item_id"`
	SKU               string                   `json:"sku"`
	Name              string                   `json:"name"`
	Category          entities.ItemCategory    `json:"category"`
	CurrentStock      float64                  `json:"current_stock"`
	ReorderPoint      float64                  `json:"reorder_point"`
	RecommendedQty    float64                  `json:"recommended_quantity"`
	EstimatedCost     entities.Money           `json:"estimated_cost"`
	PreferredSupplier *SupplierRecommendation  `json:"preferred_supplier,omitempty"`
	AlternateSuppliers []*SupplierRecommendation `json:"alternate_suppliers,omitempty"`
	UrgencyScore      float64                  `json:"urgency_score"`
	LeadTimeDays      int                      `json:"lead_time_days"`
	StockoutRisk      float64                  `json:"stockout_risk"`
	Reason            string                   `json:"reason"`
	Priority          string                   `json:"priority"`
	ReorderRules      *ReorderRules            `json:"reorder_rules,omitempty"`
	ForecastData      *DemandForecast          `json:"forecast_data,omitempty"`
	CreatedAt         time.Time                `json:"created_at"`
}

// SupplierRecommendation represents a supplier recommendation
type SupplierRecommendation struct {
	SupplierID       uuid.UUID      `json:"supplier_id"`
	SupplierName     string         `json:"supplier_name"`
	UnitPrice        entities.Money `json:"unit_price"`
	MinOrderQty      float64        `json:"min_order_quantity"`
	LeadTimeDays     int            `json:"lead_time_days"`
	QualityRating    float64        `json:"quality_rating"`
	ReliabilityScore float64        `json:"reliability_score"`
	TotalCost        entities.Money `json:"total_cost"`
	IsPreferred      bool           `json:"is_preferred"`
	LastOrderDate    *time.Time     `json:"last_order_date,omitempty"`
}

// ReorderRules defines rules for automatic reordering
type ReorderRules struct {
	ItemID              uuid.UUID     `json:"item_id"`
	IsEnabled           bool          `json:"is_enabled"`
	ReorderMethod       string        `json:"reorder_method"`       // fixed, economic_order_quantity, forecast_based
	MinOrderQuantity    float64       `json:"min_order_quantity"`
	MaxOrderQuantity    float64       `json:"max_order_quantity"`
	PreferredSuppliers  []uuid.UUID   `json:"preferred_suppliers"`
	MaxLeadTime         int           `json:"max_lead_time"`        // days
	MaxUnitCost         entities.Money `json:"max_unit_cost"`
	RequireApproval     bool          `json:"require_approval"`
	ApprovalThreshold   entities.Money `json:"approval_threshold"`
	SeasonalAdjustment  bool          `json:"seasonal_adjustment"`
	SafetyStockDays     int           `json:"safety_stock_days"`
	ReviewCycleDays     int           `json:"review_cycle_days"`
	LastReviewDate      *time.Time    `json:"last_review_date,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	CreatedBy           string        `json:"created_by"`
	UpdatedBy           string        `json:"updated_by"`
}

// AutoReorderCriteria defines criteria for automatic reorder processing
type AutoReorderCriteria struct {
	MaxOrderValue      *float64    `json:"max_order_value,omitempty"`
	MinUrgencyScore    *float64    `json:"min_urgency_score,omitempty"`
	RequireApproval    *bool       `json:"require_approval,omitempty"`
	PreferredSuppliers []uuid.UUID `json:"preferred_suppliers,omitempty"`
	MaxLeadTime        *int        `json:"max_lead_time,omitempty"`
	MinQualityRating   *float64    `json:"min_quality_rating,omitempty"`
	Categories         []entities.ItemCategory `json:"categories,omitempty"`
	ExcludeItems       []uuid.UUID `json:"exclude_items,omitempty"`
}

// NewAutoReorderService creates a new auto reorder service
func NewAutoReorderService(
	inventoryRepo repositories.InventoryRepository,
	supplierRepo repositories.SupplierRepository,
	purchaseOrderRepo repositories.PurchaseOrderRepository,
	demandForecastSvc *DemandForecastingService,
	eventPublisher EventPublisher,
	logger Logger,
) *AutoReorderService {
	return &AutoReorderService{
		inventoryRepo:     inventoryRepo,
		supplierRepo:      supplierRepo,
		purchaseOrderRepo: purchaseOrderRepo,
		demandForecastSvc: demandForecastSvc,
		eventPublisher:    eventPublisher,
		logger:            logger,
	}
}

// EvaluateReorderNeeds evaluates which items need to be reordered
func (ars *AutoReorderService) EvaluateReorderNeeds(ctx context.Context, locationID *uuid.UUID) ([]*ReorderRecommendation, error) {
	ars.logger.Info("Evaluating reorder needs", "location_id", locationID)

	// Get items that need reordering
	items, err := ars.inventoryRepo.GetItemsNeedingReorder(ctx, locationID)
	if err != nil {
		ars.logger.Error("Failed to get items needing reorder", err)
		return nil, err
	}

	var recommendations []*ReorderRecommendation

	for _, item := range items {
		recommendation, err := ars.createReorderRecommendation(ctx, item)
		if err != nil {
			ars.logger.Error("Failed to create reorder recommendation", err, "item_id", item.ID)
			continue
		}

		if recommendation != nil {
			recommendations = append(recommendations, recommendation)
		}
	}

	// Sort by urgency score (highest first)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].UrgencyScore > recommendations[j].UrgencyScore
	})

	ars.logger.Info("Reorder evaluation completed", 
		"total_items", len(items),
		"recommendations", len(recommendations))

	return recommendations, nil
}

// CreateAutomaticPurchaseOrder creates purchase orders based on recommendations
func (ars *AutoReorderService) CreateAutomaticPurchaseOrder(ctx context.Context, recommendations []*ReorderRecommendation) ([]*entities.PurchaseOrder, error) {
	ars.logger.Info("Creating automatic purchase orders", "recommendations", len(recommendations))

	// Group recommendations by supplier
	supplierGroups := ars.groupRecommendationsBySupplier(recommendations)

	var createdOrders []*entities.PurchaseOrder

	for supplierID, supplierRecs := range supplierGroups {
		order, err := ars.createPurchaseOrderForSupplier(ctx, supplierID, supplierRecs)
		if err != nil {
			ars.logger.Error("Failed to create purchase order", err, "supplier_id", supplierID)
			continue
		}

		createdOrders = append(createdOrders, order)
	}

	ars.logger.Info("Automatic purchase orders created", "orders", len(createdOrders))

	return createdOrders, nil
}

// GetReorderRules gets reorder rules for an item
func (ars *AutoReorderService) GetReorderRules(ctx context.Context, itemID uuid.UUID) (*ReorderRules, error) {
	// In a real implementation, this would fetch from a repository
	// For now, return default rules
	return &ReorderRules{
		ItemID:             itemID,
		IsEnabled:          true,
		ReorderMethod:      "forecast_based",
		MinOrderQuantity:   1,
		MaxOrderQuantity:   1000,
		MaxLeadTime:        30,
		RequireApproval:    false,
		ApprovalThreshold:  entities.Money{Amount: 1000, Currency: "USD"},
		SeasonalAdjustment: true,
		SafetyStockDays:    7,
		ReviewCycleDays:    30,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}, nil
}

// UpdateReorderRules updates reorder rules for an item
func (ars *AutoReorderService) UpdateReorderRules(ctx context.Context, itemID uuid.UUID, rules *ReorderRules) error {
	ars.logger.Info("Updating reorder rules", "item_id", itemID)

	rules.ItemID = itemID
	rules.UpdatedAt = time.Now()

	// In a real implementation, this would save to a repository
	// For now, just log the update
	ars.logger.Info("Reorder rules updated", "item_id", itemID)

	return nil
}

// Helper methods

func (ars *AutoReorderService) createReorderRecommendation(ctx context.Context, item *entities.InventoryItem) (*ReorderRecommendation, error) {
	// Get reorder rules
	rules, err := ars.GetReorderRules(ctx, item.ID)
	if err != nil {
		ars.logger.Warn("Failed to get reorder rules, using defaults", "item_id", item.ID)
		rules = &ReorderRules{IsEnabled: true, ReorderMethod: "fixed"}
	}

	if !rules.IsEnabled {
		return nil, nil // Skip disabled items
	}

	// Calculate recommended quantity
	recommendedQty, err := ars.calculateRecommendedQuantity(ctx, item, rules)
	if err != nil {
		ars.logger.Error("Failed to calculate recommended quantity", err, "item_id", item.ID)
		recommendedQty = item.ReorderQuantity // Fallback to configured reorder quantity
	}

	// Get supplier recommendations
	supplierRecs, err := ars.getSupplierRecommendations(ctx, item, recommendedQty)
	if err != nil {
		ars.logger.Error("Failed to get supplier recommendations", err, "item_id", item.ID)
		return nil, err
	}

	if len(supplierRecs) == 0 {
		ars.logger.Warn("No suppliers available for item", "item_id", item.ID)
		return nil, nil
	}

	// Select preferred supplier (first one is best)
	preferredSupplier := supplierRecs[0]
	alternateSuppliers := supplierRecs[1:]

	// Calculate urgency score
	urgencyScore := ars.calculateUrgencyScore(ctx, item)

	// Determine priority
	priority := ars.determinePriority(urgencyScore, item.CurrentStock, item.MinimumLevel)

	// Get forecast data if available
	var forecastData *DemandForecast
	if rules.ReorderMethod == "forecast_based" {
		forecast, err := ars.demandForecastSvc.ForecastDemand(ctx, item.ID, 30*24*time.Hour)
		if err == nil {
			forecastData = forecast
		}
	}

	// Calculate stockout risk
	stockoutRisk := ars.calculateStockoutRisk(item, forecastData)

	recommendation := &ReorderRecommendation{
		ItemID:             item.ID,
		SKU:                item.SKU,
		Name:               item.Name,
		Category:           item.Category,
		CurrentStock:       item.CurrentStock,
		ReorderPoint:       item.ReorderPoint,
		RecommendedQty:     recommendedQty,
		EstimatedCost:      entities.Money{Amount: recommendedQty * preferredSupplier.UnitPrice.Amount, Currency: preferredSupplier.UnitPrice.Currency},
		PreferredSupplier:  preferredSupplier,
		AlternateSuppliers: alternateSuppliers,
		UrgencyScore:       urgencyScore,
		LeadTimeDays:       preferredSupplier.LeadTimeDays,
		StockoutRisk:       stockoutRisk,
		Reason:             ars.generateReorderReason(item, urgencyScore),
		Priority:           priority,
		ReorderRules:       rules,
		ForecastData:       forecastData,
		CreatedAt:          time.Now(),
	}

	return recommendation, nil
}

func (ars *AutoReorderService) calculateRecommendedQuantity(ctx context.Context, item *entities.InventoryItem, rules *ReorderRules) (float64, error) {
	switch rules.ReorderMethod {
	case "fixed":
		return item.ReorderQuantity, nil
		
	case "economic_order_quantity":
		return ars.calculateEOQ(ctx, item)
		
	case "forecast_based":
		return ars.calculateForecastBasedQuantity(ctx, item, rules)
		
	default:
		return item.ReorderQuantity, nil
	}
}

func (ars *AutoReorderService) calculateEOQ(ctx context.Context, item *entities.InventoryItem) (float64, error) {
	// Economic Order Quantity formula: EOQ = sqrt((2 * D * S) / H)
	// D = Annual demand, S = Setup cost, H = Holding cost
	
	// Get annual demand from consumption pattern
	pattern, err := ars.demandForecastSvc.GetConsumptionPattern(ctx, item.ID, 365*24*time.Hour)
	if err != nil {
		return item.ReorderQuantity, err
	}

	annualDemand := pattern.AverageDaily * 365
	setupCost := 50.0 // Simplified setup cost
	holdingCostRate := 0.2 // 20% of item value per year
	holdingCost := item.UnitCost.Amount * holdingCostRate

	if holdingCost <= 0 {
		return item.ReorderQuantity, nil
	}

	eoq := math.Sqrt((2 * annualDemand * setupCost) / holdingCost)
	
	// Apply constraints
	if eoq < item.ReorderQuantity {
		eoq = item.ReorderQuantity
	}
	
	return eoq, nil
}

func (ars *AutoReorderService) calculateForecastBasedQuantity(ctx context.Context, item *entities.InventoryItem, rules *ReorderRules) (float64, error) {
	// Get demand forecast for lead time + safety stock period
	forecastPeriod := time.Duration(rules.SafetyStockDays+30) * 24 * time.Hour // 30 days lead time + safety stock
	
	forecast, err := ars.demandForecastSvc.ForecastDemand(ctx, item.ID, forecastPeriod)
	if err != nil {
		return item.ReorderQuantity, err
	}

	// Calculate quantity needed to cover forecast demand
	forecastDemand := forecast.PredictedDemand
	
	// Add safety stock
	safetyStock := forecastDemand * 0.2 // 20% safety stock
	
	totalNeeded := forecastDemand + safetyStock
	
	// Subtract current available stock
	quantityToOrder := totalNeeded - item.AvailableStock
	
	if quantityToOrder < item.ReorderQuantity {
		quantityToOrder = item.ReorderQuantity
	}
	
	return quantityToOrder, nil
}

func (ars *AutoReorderService) getSupplierRecommendations(ctx context.Context, item *entities.InventoryItem, quantity float64) ([]*SupplierRecommendation, error) {
	// Get suppliers for this item
	suppliers, err := ars.supplierRepo.GetSuppliersForProduct(ctx, item.SKU)
	if err != nil {
		return nil, err
	}

	var recommendations []*SupplierRecommendation

	for _, supplier := range suppliers {
		if !supplier.IsActive || supplier.Status != entities.SupplierStatusActive {
			continue
		}

		// Get pricing for this quantity
		price := supplier.GetCurrentPrice(item.SKU, quantity)
		if price == nil {
			continue
		}

		// Get supplier product info
		product := supplier.GetProductBySKU(item.SKU)
		if product == nil {
			continue
		}

		recommendation := &SupplierRecommendation{
			SupplierID:       supplier.ID,
			SupplierName:     supplier.Name,
			UnitPrice:        *price,
			MinOrderQty:      product.MinOrderQuantity,
			LeadTimeDays:     product.LeadTimeDays,
			QualityRating:    supplier.Rating,
			ReliabilityScore: ars.calculateReliabilityScore(supplier),
			TotalCost:        entities.Money{Amount: quantity * price.Amount, Currency: price.Currency},
			IsPreferred:      supplier.IsPreferred,
		}

		if supplier.Performance != nil && supplier.Performance.LastOrderDate != nil {
			recommendation.LastOrderDate = supplier.Performance.LastOrderDate
		}

		recommendations = append(recommendations, recommendation)
	}

	// Sort by preference and cost
	sort.Slice(recommendations, func(i, j int) bool {
		// Preferred suppliers first
		if recommendations[i].IsPreferred != recommendations[j].IsPreferred {
			return recommendations[i].IsPreferred
		}
		// Then by reliability score
		if recommendations[i].ReliabilityScore != recommendations[j].ReliabilityScore {
			return recommendations[i].ReliabilityScore > recommendations[j].ReliabilityScore
		}
		// Finally by total cost
		return recommendations[i].TotalCost.Amount < recommendations[j].TotalCost.Amount
	})

	return recommendations, nil
}

func (ars *AutoReorderService) calculateUrgencyScore(ctx context.Context, item *entities.InventoryItem) float64 {
	// Base urgency on stock level relative to minimum
	stockRatio := item.CurrentStock / item.MinimumLevel
	
	var urgency float64
	if stockRatio <= 0 {
		urgency = 100 // Out of stock
	} else if stockRatio <= 0.5 {
		urgency = 90 // Very low stock
	} else if stockRatio <= 1.0 {
		urgency = 70 // Below minimum
	} else if stockRatio <= 1.5 {
		urgency = 50 // Approaching minimum
	} else {
		urgency = 20 // Adequate stock
	}

	// Adjust for consumption velocity
	pattern, err := ars.demandForecastSvc.GetConsumptionPattern(ctx, item.ID, 30*24*time.Hour)
	if err == nil && pattern.AverageDaily > 0 {
		daysOfStock := item.CurrentStock / pattern.AverageDaily
		if daysOfStock < 7 {
			urgency += 20
		} else if daysOfStock < 14 {
			urgency += 10
		}
	}

	if urgency > 100 {
		urgency = 100
	}

	return urgency
}

func (ars *AutoReorderService) calculateReliabilityScore(supplier *entities.Supplier) float64 {
	if supplier.Performance == nil {
		return 50.0 // Default score for new suppliers
	}

	// Combine various performance metrics
	score := 0.0
	score += supplier.Performance.OnTimeDeliveryRate * 40  // 40% weight
	score += (1 - supplier.Performance.QualityRejectRate) * 30 // 30% weight
	score += supplier.Performance.OrderFulfillmentRate * 20    // 20% weight
	score += (supplier.Rating / 5.0) * 10                     // 10% weight

	return score
}

func (ars *AutoReorderService) calculateStockoutRisk(item *entities.InventoryItem, forecast *DemandForecast) float64 {
	if forecast == nil {
		// Simple calculation based on current stock and minimum level
		if item.CurrentStock <= 0 {
			return 100.0
		}
		if item.CurrentStock <= item.MinimumLevel {
			return 80.0
		}
		return 20.0
	}

	// Use forecast confidence interval to calculate risk
	if forecast.ConfidenceInterval != nil {
		upperBound := forecast.ConfidenceInterval.Upper
		if item.CurrentStock < upperBound {
			return 75.0
		}
	}

	return 25.0
}

func (ars *AutoReorderService) determinePriority(urgencyScore, currentStock, minimumLevel float64) string {
	if urgencyScore >= 90 || currentStock <= 0 {
		return "critical"
	} else if urgencyScore >= 70 || currentStock <= minimumLevel {
		return "high"
	} else if urgencyScore >= 50 {
		return "medium"
	}
	return "low"
}

func (ars *AutoReorderService) generateReorderReason(item *entities.InventoryItem, urgencyScore float64) string {
	if item.CurrentStock <= 0 {
		return "Item is out of stock"
	} else if item.CurrentStock <= item.MinimumLevel {
		return fmt.Sprintf("Stock below minimum level (%.2f < %.2f)", item.CurrentStock, item.MinimumLevel)
	} else if item.CurrentStock <= item.ReorderPoint {
		return fmt.Sprintf("Stock at reorder point (%.2f <= %.2f)", item.CurrentStock, item.ReorderPoint)
	} else if urgencyScore >= 70 {
		return "High consumption rate detected"
	}
	return "Proactive reorder based on forecast"
}

func (ars *AutoReorderService) groupRecommendationsBySupplier(recommendations []*ReorderRecommendation) map[uuid.UUID][]*ReorderRecommendation {
	groups := make(map[uuid.UUID][]*ReorderRecommendation)
	
	for _, rec := range recommendations {
		if rec.PreferredSupplier != nil {
			supplierID := rec.PreferredSupplier.SupplierID
			groups[supplierID] = append(groups[supplierID], rec)
		}
	}
	
	return groups
}

func (ars *AutoReorderService) createPurchaseOrderForSupplier(ctx context.Context, supplierID uuid.UUID, recommendations []*ReorderRecommendation) (*entities.PurchaseOrder, error) {
	// Get supplier details
	supplier, err := ars.supplierRepo.GetByID(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	// Determine location (use first recommendation's location or default)
	var locationID uuid.UUID
	if len(recommendations) > 0 {
		// In a real implementation, get location from item
		locationID = uuid.New() // Placeholder
	}

	// Create purchase order
	order := entities.NewPurchaseOrder(
		supplierID,
		locationID,
		uuid.New(), // Buyer ID
		"auto_reorder_system",
		"Auto Reorder System",
		entities.POTypeStandard,
		entities.PriorityNormal,
	)

	// Add items to order
	for _, rec := range recommendations {
		item := entities.NewPurchaseOrderItem(
			rec.ItemID,
			rec.SKU,
			rec.SKU,
			rec.Name,
			rec.RecommendedQty,
			entities.UnitPiece, // Simplified
			rec.PreferredSupplier.UnitPrice,
		)
		
		order.AddItem(item)
	}

	// Set payment and delivery terms from supplier
	if supplier.PaymentTerms != nil {
		order.PaymentTerms = supplier.PaymentTerms
	}
	if supplier.DeliveryTerms != nil {
		order.DeliveryTerms = supplier.DeliveryTerms
	}

	// Create the order
	if err := ars.purchaseOrderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Publish event
	event := events.NewPurchaseOrderCreatedEvent(order)
	if err := ars.eventPublisher.PublishEvent(ctx, event); err != nil {
		ars.logger.Error("Failed to publish purchase order created event", err, "order_id", order.ID)
	}

	return order, nil
}
