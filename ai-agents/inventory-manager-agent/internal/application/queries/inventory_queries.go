package queries

import (
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/entities"
)

// InventoryOverviewQuery represents a query for inventory overview
type InventoryOverviewQuery struct {
	LocationID   *uuid.UUID `json:"location_id,omitempty"`
	IncludeInactive bool    `json:"include_inactive"`
	Categories   []entities.ItemCategory `json:"categories,omitempty"`
	SupplierIDs  []uuid.UUID `json:"supplier_ids,omitempty"`
	Tags         []string    `json:"tags,omitempty"`
}

// GetInventoryItemQuery represents a query to get a specific inventory item
type GetInventoryItemQuery struct {
	ID  *uuid.UUID `json:"id,omitempty"`
	SKU *string    `json:"sku,omitempty"`
	IncludeBatches    bool `json:"include_batches"`
	IncludeMovements  bool `json:"include_movements"`
	IncludeSupplier   bool `json:"include_supplier"`
	IncludeLocation   bool `json:"include_location"`
}

// ListInventoryItemsQuery represents a query to list inventory items
type ListInventoryItemsQuery struct {
	LocationIDs    []uuid.UUID                `json:"location_ids,omitempty"`
	Categories     []entities.ItemCategory    `json:"categories,omitempty"`
	SupplierIDs    []uuid.UUID                `json:"supplier_ids,omitempty"`
	Status         []entities.ItemStatus      `json:"status,omitempty"`
	IsActive       *bool                      `json:"is_active,omitempty"`
	IsPerishable   *bool                      `json:"is_perishable,omitempty"`
	MinStock       *float64                   `json:"min_stock,omitempty"`
	MaxStock       *float64                   `json:"max_stock,omitempty"`
	MinValue       *float64                   `json:"min_value,omitempty"`
	MaxValue       *float64                   `json:"max_value,omitempty"`
	Tags           []string                   `json:"tags,omitempty"`
	SearchTerm     string                     `json:"search_term,omitempty"`
	CreatedAfter   *time.Time                 `json:"created_after,omitempty"`
	CreatedBefore  *time.Time                 `json:"created_before,omitempty"`
	UpdatedAfter   *time.Time                 `json:"updated_after,omitempty"`
	UpdatedBefore  *time.Time                 `json:"updated_before,omitempty"`
	SortBy         string                     `json:"sort_by,omitempty"`
	SortOrder      string                     `json:"sort_order,omitempty"`
	Page           int                        `json:"page" validate:"min=1"`
	PageSize       int                        `json:"page_size" validate:"min=1,max=1000"`
}

// GetStockMovementsQuery represents a query to get stock movements
type GetStockMovementsQuery struct {
	ItemIDs        []uuid.UUID                  `json:"item_ids,omitempty"`
	LocationIDs    []uuid.UUID                  `json:"location_ids,omitempty"`
	Types          []entities.MovementType      `json:"types,omitempty"`
	Directions     []entities.MovementDirection `json:"directions,omitempty"`
	Status         []entities.MovementStatus    `json:"status,omitempty"`
	SupplierIDs    []uuid.UUID                  `json:"supplier_ids,omitempty"`
	ReferenceTypes []string                     `json:"reference_types,omitempty"`
	ReferenceIDs   []uuid.UUID                  `json:"reference_ids,omitempty"`
	ProcessedBy    []string                     `json:"processed_by,omitempty"`
	CreatedAfter   *time.Time                   `json:"created_after,omitempty"`
	CreatedBefore  *time.Time                   `json:"created_before,omitempty"`
	ProcessedAfter *time.Time                   `json:"processed_after,omitempty"`
	ProcessedBefore *time.Time                  `json:"processed_before,omitempty"`
	MinQuantity    *float64                     `json:"min_quantity,omitempty"`
	MaxQuantity    *float64                     `json:"max_quantity,omitempty"`
	MinValue       *float64                     `json:"min_value,omitempty"`
	MaxValue       *float64                     `json:"max_value,omitempty"`
	Tags           []string                     `json:"tags,omitempty"`
	SortBy         string                       `json:"sort_by,omitempty"`
	SortOrder      string                       `json:"sort_order,omitempty"`
	Page           int                          `json:"page" validate:"min=1"`
	PageSize       int                          `json:"page_size" validate:"min=1,max=1000"`
}

// DemandForecastQuery represents a query for demand forecasting
type DemandForecastQuery struct {
	ItemID     uuid.UUID     `json:"item_id" validate:"required"`
	Period     time.Duration `json:"period" validate:"required"`
	LocationID *uuid.UUID    `json:"location_id,omitempty"`
	Algorithm  string        `json:"algorithm,omitempty"`
	Confidence float64       `json:"confidence" validate:"min=0,max=1"`
}

// ReorderRecommendationsQuery represents a query for reorder recommendations
type ReorderRecommendationsQuery struct {
	LocationID      *uuid.UUID              `json:"location_id,omitempty"`
	Categories      []entities.ItemCategory `json:"categories,omitempty"`
	SupplierIDs     []uuid.UUID             `json:"supplier_ids,omitempty"`
	MinUrgency      float64                 `json:"min_urgency" validate:"min=0,max=100"`
	MaxRecommendations int                  `json:"max_recommendations" validate:"min=1,max=1000"`
	IncludeCosts    bool                    `json:"include_costs"`
	IncludeSupplierInfo bool                `json:"include_supplier_info"`
}

// InventoryAnalyticsQuery represents a query for inventory analytics
type InventoryAnalyticsQuery struct {
	LocationID    *uuid.UUID              `json:"location_id,omitempty"`
	Categories    []entities.ItemCategory `json:"categories,omitempty"`
	SupplierIDs   []uuid.UUID             `json:"supplier_ids,omitempty"`
	Period        time.Duration           `json:"period" validate:"required"`
	GroupBy       string                  `json:"group_by,omitempty"`
	Metrics       []string                `json:"metrics,omitempty"`
	IncludeTrends bool                    `json:"include_trends"`
	IncludeForecasts bool                 `json:"include_forecasts"`
}

// StockAlertsQuery represents a query for stock alerts
type StockAlertsQuery struct {
	LocationID    *uuid.UUID              `json:"location_id,omitempty"`
	Categories    []entities.ItemCategory `json:"categories,omitempty"`
	AlertTypes    []string                `json:"alert_types,omitempty"`
	Severity      []string                `json:"severity,omitempty"`
	IncludeResolved bool                  `json:"include_resolved"`
	CreatedAfter  *time.Time              `json:"created_after,omitempty"`
	CreatedBefore *time.Time              `json:"created_before,omitempty"`
	SortBy        string                  `json:"sort_by,omitempty"`
	SortOrder     string                  `json:"sort_order,omitempty"`
	Page          int                     `json:"page" validate:"min=1"`
	PageSize      int                     `json:"page_size" validate:"min=1,max=1000"`
}

// ExpiringItemsQuery represents a query for expiring items
type ExpiringItemsQuery struct {
	LocationID     *uuid.UUID              `json:"location_id,omitempty"`
	Categories     []entities.ItemCategory `json:"categories,omitempty"`
	WithinDays     int                     `json:"within_days" validate:"min=1,max=365"`
	IncludeExpired bool                    `json:"include_expired"`
	MinQuantity    *float64                `json:"min_quantity,omitempty"`
	SortBy         string                  `json:"sort_by,omitempty"`
	SortOrder      string                  `json:"sort_order,omitempty"`
	Page           int                     `json:"page" validate:"min=1"`
	PageSize       int                     `json:"page_size" validate:"min=1,max=1000"`
}

// SupplierPerformanceQuery represents a query for supplier performance
type SupplierPerformanceQuery struct {
	SupplierIDs   []uuid.UUID `json:"supplier_ids,omitempty"`
	Categories    []entities.SupplierCategory `json:"categories,omitempty"`
	Period        time.Duration `json:"period" validate:"required"`
	MinOrders     int         `json:"min_orders" validate:"min=1"`
	Metrics       []string    `json:"metrics,omitempty"`
	IncludeTrends bool        `json:"include_trends"`
	SortBy        string      `json:"sort_by,omitempty"`
	SortOrder     string      `json:"sort_order,omitempty"`
	Page          int         `json:"page" validate:"min=1"`
	PageSize      int         `json:"page_size" validate:"min=1,max=1000"`
}

// LocationUtilizationQuery represents a query for location utilization
type LocationUtilizationQuery struct {
	LocationIDs   []uuid.UUID           `json:"location_ids,omitempty"`
	Types         []entities.LocationType `json:"types,omitempty"`
	IncludeZones  bool                  `json:"include_zones"`
	IncludeMetrics bool                 `json:"include_metrics"`
	Period        time.Duration         `json:"period,omitempty"`
	SortBy        string                `json:"sort_by,omitempty"`
	SortOrder     string                `json:"sort_order,omitempty"`
}

// StockValuationQuery represents a query for stock valuation
type StockValuationQuery struct {
	LocationID     *uuid.UUID              `json:"location_id,omitempty"`
	Categories     []entities.ItemCategory `json:"categories,omitempty"`
	SupplierIDs    []uuid.UUID             `json:"supplier_ids,omitempty"`
	ValuationMethod string                 `json:"valuation_method,omitempty"`
	AsOfDate       *time.Time              `json:"as_of_date,omitempty"`
	Currency       string                  `json:"currency,omitempty"`
	GroupBy        string                  `json:"group_by,omitempty"`
	IncludeDetails bool                    `json:"include_details"`
}

// MovementTrendsQuery represents a query for movement trends
type MovementTrendsQuery struct {
	ItemIDs     []uuid.UUID                  `json:"item_ids,omitempty"`
	LocationIDs []uuid.UUID                  `json:"location_ids,omitempty"`
	Categories  []entities.ItemCategory      `json:"categories,omitempty"`
	Types       []entities.MovementType      `json:"types,omitempty"`
	Period      time.Duration                `json:"period" validate:"required"`
	Granularity string                       `json:"granularity,omitempty"`
	GroupBy     string                       `json:"group_by,omitempty"`
	Metrics     []string                     `json:"metrics,omitempty"`
}

// QualityMetricsQuery represents a query for quality metrics
type QualityMetricsQuery struct {
	LocationID     *uuid.UUID              `json:"location_id,omitempty"`
	Categories     []entities.ItemCategory `json:"categories,omitempty"`
	SupplierIDs    []uuid.UUID             `json:"supplier_ids,omitempty"`
	Period         time.Duration           `json:"period" validate:"required"`
	CheckTypes     []string                `json:"check_types,omitempty"`
	IncludeTrends  bool                    `json:"include_trends"`
	GroupBy        string                  `json:"group_by,omitempty"`
}

// CostAnalysisQuery represents a query for cost analysis
type CostAnalysisQuery struct {
	LocationID      *uuid.UUID              `json:"location_id,omitempty"`
	Categories      []entities.ItemCategory `json:"categories,omitempty"`
	SupplierIDs     []uuid.UUID             `json:"supplier_ids,omitempty"`
	Period          time.Duration           `json:"period" validate:"required"`
	CostTypes       []string                `json:"cost_types,omitempty"`
	IncludeVariance bool                    `json:"include_variance"`
	IncludeTrends   bool                    `json:"include_trends"`
	Currency        string                  `json:"currency,omitempty"`
	GroupBy         string                  `json:"group_by,omitempty"`
}

// Default values and validation

// SetDefaults sets default values for ListInventoryItemsQuery
func (q *ListInventoryItemsQuery) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}
	if q.SortBy == "" {
		q.SortBy = "name"
	}
	if q.SortOrder == "" {
		q.SortOrder = "asc"
	}
}

// SetDefaults sets default values for GetStockMovementsQuery
func (q *GetStockMovementsQuery) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
	if q.SortOrder == "" {
		q.SortOrder = "desc"
	}
}

// SetDefaults sets default values for StockAlertsQuery
func (q *StockAlertsQuery) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
	if q.SortOrder == "" {
		q.SortOrder = "desc"
	}
}

// SetDefaults sets default values for ExpiringItemsQuery
func (q *ExpiringItemsQuery) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}
	if q.WithinDays == 0 {
		q.WithinDays = 7
	}
	if q.SortBy == "" {
		q.SortBy = "expiration_date"
	}
	if q.SortOrder == "" {
		q.SortOrder = "asc"
	}
}

// SetDefaults sets default values for SupplierPerformanceQuery
func (q *SupplierPerformanceQuery) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}
	if q.MinOrders == 0 {
		q.MinOrders = 1
	}
	if q.SortBy == "" {
		q.SortBy = "overall_rating"
	}
	if q.SortOrder == "" {
		q.SortOrder = "desc"
	}
}

// SetDefaults sets default values for DemandForecastQuery
func (q *DemandForecastQuery) SetDefaults() {
	if q.Confidence == 0 {
		q.Confidence = 0.95
	}
	if q.Algorithm == "" {
		q.Algorithm = "auto"
	}
}

// SetDefaults sets default values for ReorderRecommendationsQuery
func (q *ReorderRecommendationsQuery) SetDefaults() {
	if q.MaxRecommendations == 0 {
		q.MaxRecommendations = 100
	}
	if q.MinUrgency == 0 {
		q.MinUrgency = 50
	}
}

// Validate validates the query parameters
func (q *ListInventoryItemsQuery) Validate() error {
	if q.Page < 1 {
		return entities.NewDomainError("INVALID_PAGE", "Page must be greater than 0")
	}
	if q.PageSize < 1 || q.PageSize > 1000 {
		return entities.NewDomainError("INVALID_PAGE_SIZE", "Page size must be between 1 and 1000")
	}
	if q.MinStock != nil && q.MaxStock != nil && *q.MinStock > *q.MaxStock {
		return entities.NewDomainError("INVALID_STOCK_RANGE", "Min stock cannot be greater than max stock")
	}
	if q.MinValue != nil && q.MaxValue != nil && *q.MinValue > *q.MaxValue {
		return entities.NewDomainError("INVALID_VALUE_RANGE", "Min value cannot be greater than max value")
	}
	return nil
}

// Validate validates the GetStockMovementsQuery
func (q *GetStockMovementsQuery) Validate() error {
	if q.Page < 1 {
		return entities.NewDomainError("INVALID_PAGE", "Page must be greater than 0")
	}
	if q.PageSize < 1 || q.PageSize > 1000 {
		return entities.NewDomainError("INVALID_PAGE_SIZE", "Page size must be between 1 and 1000")
	}
	if q.CreatedAfter != nil && q.CreatedBefore != nil && q.CreatedAfter.After(*q.CreatedBefore) {
		return entities.NewDomainError("INVALID_DATE_RANGE", "Created after cannot be later than created before")
	}
	return nil
}

// Validate validates the DemandForecastQuery
func (q *DemandForecastQuery) Validate() error {
	if q.Period <= 0 {
		return entities.NewDomainError("INVALID_PERIOD", "Period must be greater than 0")
	}
	if q.Confidence < 0 || q.Confidence > 1 {
		return entities.NewDomainError("INVALID_CONFIDENCE", "Confidence must be between 0 and 1")
	}
	return nil
}
