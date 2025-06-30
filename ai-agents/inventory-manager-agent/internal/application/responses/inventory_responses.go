package responses

import (
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/entities"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/repositories"
)

// InventoryItemResponse represents a complete inventory item response
type InventoryItemResponse struct {
	ID                uuid.UUID                     `json:"id"`
	SKU               string                        `json:"sku"`
	Name              string                        `json:"name"`
	Description       string                        `json:"description"`
	Category          entities.ItemCategory         `json:"category"`
	SubCategory       string                        `json:"sub_category"`
	Unit              entities.MeasurementUnit      `json:"unit"`
	CurrentStock      float64                       `json:"current_stock"`
	ReservedStock     float64                       `json:"reserved_stock"`
	AvailableStock    float64                       `json:"available_stock"`
	MinimumLevel      float64                       `json:"minimum_level"`
	MaximumLevel      float64                       `json:"maximum_level"`
	ReorderPoint      float64                       `json:"reorder_point"`
	ReorderQuantity   float64                       `json:"reorder_quantity"`
	SafetyStock       float64                       `json:"safety_stock"`
	UnitCost          entities.Money                `json:"unit_cost"`
	TotalValue        entities.Money                `json:"total_value"`
	AverageCost       entities.Money                `json:"average_cost"`
	LastCost          entities.Money                `json:"last_cost"`
	SupplierID        uuid.UUID                     `json:"supplier_id"`
	LocationID        uuid.UUID                     `json:"location_id"`
	Status            entities.ItemStatus           `json:"status"`
	IsActive          bool                          `json:"is_active"`
	IsPerishable      bool                          `json:"is_perishable"`
	ShelfLife         *time.Duration                `json:"shelf_life,omitempty"`
	StorageConditions *entities.StorageRequirements `json:"storage_conditions,omitempty"`
	Attributes        map[string]interface{}        `json:"attributes"`
	Tags              []string                      `json:"tags"`
	CreatedAt         time.Time                     `json:"created_at"`
	UpdatedAt         time.Time                     `json:"updated_at"`
	CreatedBy         string                        `json:"created_by"`
	UpdatedBy         string                        `json:"updated_by"`
	Version           int64                         `json:"version"`
	
	// Optional related data
	Supplier *SupplierSummary   `json:"supplier,omitempty"`
	Location *LocationSummary   `json:"location,omitempty"`
	Batches  []*BatchSummary    `json:"batches,omitempty"`
	RecentMovements []*MovementSummary `json:"recent_movements,omitempty"`
}

// InventoryItemSummary represents a summary of an inventory item
type InventoryItemSummary struct {
	ID             uuid.UUID             `json:"id"`
	SKU            string                `json:"sku"`
	Name           string                `json:"name"`
	Category       entities.ItemCategory `json:"category"`
	CurrentStock   float64               `json:"current_stock"`
	AvailableStock float64               `json:"available_stock"`
	MinimumLevel   float64               `json:"minimum_level"`
	ReorderPoint   float64               `json:"reorder_point"`
	Status         entities.ItemStatus   `json:"status"`
	TotalValue     entities.Money        `json:"total_value"`
	IsLowStock     bool                  `json:"is_low_stock"`
	IsOutOfStock   bool                  `json:"is_out_of_stock"`
	NeedsReorder   bool                  `json:"needs_reorder"`
}

// StockMovementResponse represents a stock movement response
type StockMovementResponse struct {
	ID              uuid.UUID                    `json:"id"`
	MovementNumber  string                       `json:"movement_number"`
	Type            entities.MovementType        `json:"type"`
	Status          entities.MovementStatus      `json:"status"`
	Direction       entities.MovementDirection   `json:"direction"`
	InventoryItemID uuid.UUID                    `json:"inventory_item_id"`
	Quantity        float64                      `json:"quantity"`
	Unit            entities.MeasurementUnit     `json:"unit"`
	UnitCost        entities.Money               `json:"unit_cost"`
	TotalCost       entities.Money               `json:"total_cost"`
	FromLocationID  *uuid.UUID                   `json:"from_location_id,omitempty"`
	ToLocationID    *uuid.UUID                   `json:"to_location_id,omitempty"`
	Reason          string                       `json:"reason"`
	ProcessedAt     *time.Time                   `json:"processed_at,omitempty"`
	ProcessedBy     string                       `json:"processed_by"`
	CompletedAt     *time.Time                   `json:"completed_at,omitempty"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
	CreatedBy       string                       `json:"created_by"`
	UpdatedBy       string                       `json:"updated_by"`
	
	// Optional related data
	InventoryItem *InventoryItemSummary `json:"inventory_item,omitempty"`
	FromLocation  *LocationSummary      `json:"from_location,omitempty"`
	ToLocation    *LocationSummary      `json:"to_location,omitempty"`
	Supplier      *SupplierSummary      `json:"supplier,omitempty"`
}

// InventoryOverviewResponse represents an inventory overview response
type InventoryOverviewResponse struct {
	TotalItems      int                      `json:"total_items"`
	TotalValue      entities.Money           `json:"total_value"`
	LowStockCount   int                      `json:"low_stock_count"`
	OutOfStockCount int                      `json:"out_of_stock_count"`
	ReorderCount    int                      `json:"reorder_count"`
	ExpiringCount   int                      `json:"expiring_count"`
	Items           []*InventoryItemSummary  `json:"items"`
	Metrics         *InventoryMetricsResponse `json:"metrics"`
	Alerts          []*InventoryAlert        `json:"alerts"`
	GeneratedAt     time.Time                `json:"generated_at"`
}

// InventoryMetricsResponse represents inventory metrics
type InventoryMetricsResponse struct {
	Period              string                                    `json:"period"`
	TotalItems          int                                       `json:"total_items"`
	TotalValue          entities.Money                            `json:"total_value"`
	AverageValue        entities.Money                            `json:"average_value"`
	TotalMovements      int                                       `json:"total_movements"`
	InboundMovements    int                                       `json:"inbound_movements"`
	OutboundMovements   int                                       `json:"outbound_movements"`
	AdjustmentMovements int                                       `json:"adjustment_movements"`
	LowStockItems       int                                       `json:"low_stock_items"`
	OutOfStockItems     int                                       `json:"out_of_stock_items"`
	ExpiringItems       int                                       `json:"expiring_items"`
	TurnoverRate        float64                                   `json:"turnover_rate"`
	StockAccuracy       float64                                   `json:"stock_accuracy"`
	CategoryBreakdown   map[string]repositories.CategoryMetrics  `json:"category_breakdown"`
	TopMovingItems      []repositories.ItemMovementMetric        `json:"top_moving_items"`
	SlowMovingItems     []repositories.ItemMovementMetric        `json:"slow_moving_items"`
	GeneratedAt         time.Time                                 `json:"generated_at"`
}

// InventoryAlert represents an inventory alert
type InventoryAlert struct {
	Type     string    `json:"type"`
	ItemID   uuid.UUID `json:"item_id"`
	SKU      string    `json:"sku"`
	Name     string    `json:"name"`
	Message  string    `json:"message"`
	Severity string    `json:"severity"`
}

// DemandForecastResponse represents a demand forecast response
type DemandForecastResponse struct {
	ItemID             uuid.UUID            `json:"item_id"`
	Period             time.Duration        `json:"period"`
	Forecast           *DemandForecast      `json:"forecast"`
	ConsumptionPattern *ConsumptionPattern  `json:"consumption_pattern"`
	StockoutPrediction *StockoutPrediction  `json:"stockout_prediction"`
	GeneratedAt        time.Time            `json:"generated_at"`
}

// DemandForecast represents demand forecast data
type DemandForecast struct {
	PredictedDemand    float64                `json:"predicted_demand"`
	ConfidenceInterval *ConfidenceInterval    `json:"confidence_interval"`
	SeasonalFactors    map[string]float64     `json:"seasonal_factors"`
	TrendFactors       map[string]float64     `json:"trend_factors"`
	ForecastPoints     []*ForecastPoint       `json:"forecast_points"`
	Accuracy           float64                `json:"accuracy"`
	Algorithm          string                 `json:"algorithm"`
	ModelVersion       string                 `json:"model_version"`
}

// ConsumptionPattern represents consumption pattern data
type ConsumptionPattern struct {
	AverageDaily       float64            `json:"average_daily"`
	AverageWeekly      float64            `json:"average_weekly"`
	AverageMonthly     float64            `json:"average_monthly"`
	Volatility         float64            `json:"volatility"`
	Seasonality        map[string]float64 `json:"seasonality"`
	DayOfWeekPattern   map[string]float64 `json:"day_of_week_pattern"`
	MonthlyPattern     map[string]float64 `json:"monthly_pattern"`
	TrendDirection     string             `json:"trend_direction"`
	TrendStrength      float64            `json:"trend_strength"`
}

// StockoutPrediction represents stockout prediction data
type StockoutPrediction struct {
	ProbabilityPercent float64    `json:"probability_percent"`
	PredictedDate      *time.Time `json:"predicted_date,omitempty"`
	DaysUntilStockout  *int       `json:"days_until_stockout,omitempty"`
	RecommendedAction  string     `json:"recommended_action"`
	UrgencyLevel       string     `json:"urgency_level"`
	Confidence         float64    `json:"confidence"`
}

// ReorderRecommendationsResponse represents reorder recommendations
type ReorderRecommendationsResponse struct {
	LocationID      *uuid.UUID              `json:"location_id,omitempty"`
	Recommendations []*ReorderRecommendation `json:"recommendations"`
	TotalItems      int                     `json:"total_items"`
	TotalCost       entities.Money          `json:"total_cost"`
	GeneratedAt     time.Time               `json:"generated_at"`
}

// ReorderRecommendation represents a single reorder recommendation
type ReorderRecommendation struct {
	ItemID           uuid.UUID      `json:"item_id"`
	SKU              string         `json:"sku"`
	Name             string         `json:"name"`
	CurrentStock     float64        `json:"current_stock"`
	ReorderPoint     float64        `json:"reorder_point"`
	RecommendedQty   float64        `json:"recommended_quantity"`
	EstimatedCost    entities.Money `json:"estimated_cost"`
	PreferredSupplier *SupplierSummary `json:"preferred_supplier,omitempty"`
	UrgencyScore     float64        `json:"urgency_score"`
	LeadTimeDays     int            `json:"lead_time_days"`
	StockoutRisk     float64        `json:"stockout_risk"`
	Reason           string         `json:"reason"`
	Priority         string         `json:"priority"`
}

// AutoReorderProcessingResponse represents the result of automatic reorder processing
type AutoReorderProcessingResponse struct {
	ProcessedRecommendations int                      `json:"processed_recommendations"`
	CreatedOrders           int                      `json:"created_orders"`
	Orders                  []*entities.PurchaseOrder `json:"orders"`
	TotalValue              entities.Money           `json:"total_value"`
	ProcessedAt             time.Time                `json:"processed_at"`
	Errors                  []string                 `json:"errors,omitempty"`
}

// Supporting response types

// SupplierSummary represents a summary of a supplier
type SupplierSummary struct {
	ID       uuid.UUID                 `json:"id"`
	Code     string                    `json:"code"`
	Name     string                    `json:"name"`
	Type     entities.SupplierType     `json:"type"`
	Category entities.SupplierCategory `json:"category"`
	Status   entities.SupplierStatus   `json:"status"`
	Rating   float64                   `json:"rating"`
	IsActive bool                      `json:"is_active"`
	IsPreferred bool                   `json:"is_preferred"`
}

// LocationSummary represents a summary of a location
type LocationSummary struct {
	ID       uuid.UUID             `json:"id"`
	Code     string                `json:"code"`
	Name     string                `json:"name"`
	Type     entities.LocationType `json:"type"`
	Status   entities.LocationStatus `json:"status"`
	IsActive bool                  `json:"is_active"`
	IsDefault bool                 `json:"is_default"`
}

// BatchSummary represents a summary of an inventory batch
type BatchSummary struct {
	ID             uuid.UUID             `json:"id"`
	BatchNumber    string                `json:"batch_number"`
	Quantity       float64               `json:"quantity"`
	UnitCost       entities.Money        `json:"unit_cost"`
	ExpirationDate *time.Time            `json:"expiration_date,omitempty"`
	ReceivedDate   time.Time             `json:"received_date"`
	QualityStatus  entities.QualityStatus `json:"quality_status"`
	IsActive       bool                  `json:"is_active"`
}

// MovementSummary represents a summary of a stock movement
type MovementSummary struct {
	ID             uuid.UUID                  `json:"id"`
	MovementNumber string                     `json:"movement_number"`
	Type           entities.MovementType      `json:"type"`
	Direction      entities.MovementDirection `json:"direction"`
	Quantity       float64                    `json:"quantity"`
	TotalCost      entities.Money             `json:"total_cost"`
	Reason         string                     `json:"reason"`
	ProcessedAt    *time.Time                 `json:"processed_at,omitempty"`
	ProcessedBy    string                     `json:"processed_by"`
}

// ConfidenceInterval represents a confidence interval
type ConfidenceInterval struct {
	Lower      float64 `json:"lower"`
	Upper      float64 `json:"upper"`
	Confidence float64 `json:"confidence"`
}

// ForecastPoint represents a single point in a forecast
type ForecastPoint struct {
	Date       time.Time `json:"date"`
	Value      float64   `json:"value"`
	LowerBound float64   `json:"lower_bound"`
	UpperBound float64   `json:"upper_bound"`
	Confidence float64   `json:"confidence"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// InventoryItemListResponse represents a paginated list of inventory items
type InventoryItemListResponse struct {
	Items      []*InventoryItemSummary `json:"items"`
	Pagination *PaginatedResponse      `json:"pagination"`
}

// StockMovementListResponse represents a paginated list of stock movements
type StockMovementListResponse struct {
	Movements  []*StockMovementResponse `json:"movements"`
	Pagination *PaginatedResponse       `json:"pagination"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Helper functions for creating responses

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(page, pageSize, totalItems int) *PaginatedResponse {
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	
	return &PaginatedResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error, code string) *ErrorResponse {
	response := &ErrorResponse{
		Error:   err.Error(),
		Message: err.Error(),
	}
	
	if code != "" {
		response.Code = code
	}
	
	// If it's a domain error, extract the code
	if domainErr, ok := err.(*entities.DomainError); ok {
		response.Code = domainErr.Code
	}
	
	return response
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}
