package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/entities"
)

// InventoryRepository defines the interface for inventory data access
type InventoryRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, item *entities.InventoryItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.InventoryItem, error)
	GetBySKU(ctx context.Context, sku string) (*entities.InventoryItem, error)
	Update(ctx context.Context, item *entities.InventoryItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Listing and filtering
	List(ctx context.Context, filter *InventoryFilter) ([]*entities.InventoryItem, error)
	ListByLocation(ctx context.Context, locationID uuid.UUID, filter *InventoryFilter) ([]*entities.InventoryItem, error)
	ListByCategory(ctx context.Context, category entities.ItemCategory, filter *InventoryFilter) ([]*entities.InventoryItem, error)
	ListBySupplier(ctx context.Context, supplierID uuid.UUID, filter *InventoryFilter) ([]*entities.InventoryItem, error)
	
	// Stock level operations
	UpdateStock(ctx context.Context, itemID uuid.UUID, newStock float64) error
	ReserveStock(ctx context.Context, itemID uuid.UUID, quantity float64) error
	ReleaseStock(ctx context.Context, itemID uuid.UUID, quantity float64) error
	AdjustStock(ctx context.Context, itemID uuid.UUID, adjustment float64, reason string) error
	
	// Batch operations
	AddBatch(ctx context.Context, itemID uuid.UUID, batch *entities.InventoryBatch) error
	UpdateBatch(ctx context.Context, batchID uuid.UUID, batch *entities.InventoryBatch) error
	GetBatch(ctx context.Context, batchID uuid.UUID) (*entities.InventoryBatch, error)
	ListBatches(ctx context.Context, itemID uuid.UUID) ([]*entities.InventoryBatch, error)
	GetExpiringBatches(ctx context.Context, within time.Duration) ([]*entities.InventoryBatch, error)
	
	// Stock level queries
	GetLowStockItems(ctx context.Context, locationID *uuid.UUID) ([]*entities.InventoryItem, error)
	GetOutOfStockItems(ctx context.Context, locationID *uuid.UUID) ([]*entities.InventoryItem, error)
	GetItemsNeedingReorder(ctx context.Context, locationID *uuid.UUID) ([]*entities.InventoryItem, error)
	GetExpiringItems(ctx context.Context, within time.Duration, locationID *uuid.UUID) ([]*entities.InventoryItem, error)
	
	// Analytics and reporting
	GetStockValue(ctx context.Context, locationID *uuid.UUID, category *entities.ItemCategory) (*entities.Money, error)
	GetTurnoverRate(ctx context.Context, itemID uuid.UUID, period time.Duration) (float64, error)
	GetStockMovementSummary(ctx context.Context, itemID uuid.UUID, period time.Duration) (*entities.StockMovementSummary, error)
	GetInventoryMetrics(ctx context.Context, locationID *uuid.UUID, period time.Duration) (*InventoryMetrics, error)
	
	// Search and advanced queries
	Search(ctx context.Context, query string, filter *InventoryFilter) ([]*entities.InventoryItem, error)
	GetItemsByTags(ctx context.Context, tags []string) ([]*entities.InventoryItem, error)
	GetItemsByAttributes(ctx context.Context, attributes map[string]interface{}) ([]*entities.InventoryItem, error)
	
	// Bulk operations
	BulkCreate(ctx context.Context, items []*entities.InventoryItem) error
	BulkUpdate(ctx context.Context, items []*entities.InventoryItem) error
	BulkUpdateStock(ctx context.Context, updates []StockUpdate) error
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo InventoryRepository) error) error
}

// InventoryFilter defines filtering options for inventory queries
type InventoryFilter struct {
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
	CreatedAfter   *time.Time                 `json:"created_after,omitempty"`
	CreatedBefore  *time.Time                 `json:"created_before,omitempty"`
	UpdatedAfter   *time.Time                 `json:"updated_after,omitempty"`
	UpdatedBefore  *time.Time                 `json:"updated_before,omitempty"`
	SortBy         string                     `json:"sort_by,omitempty"`
	SortOrder      string                     `json:"sort_order,omitempty"`
	Limit          int                        `json:"limit,omitempty"`
	Offset         int                        `json:"offset,omitempty"`
}

// StockUpdate represents a stock update operation
type StockUpdate struct {
	ItemID    uuid.UUID `json:"item_id"`
	NewStock  float64   `json:"new_stock"`
	Reason    string    `json:"reason,omitempty"`
	Reference string    `json:"reference,omitempty"`
}

// InventoryMetrics contains inventory metrics for a period
type InventoryMetrics struct {
	Period              string                     `json:"period"`
	LocationID          *uuid.UUID                 `json:"location_id,omitempty"`
	TotalItems          int                        `json:"total_items"`
	TotalValue          entities.Money             `json:"total_value"`
	AverageValue        entities.Money             `json:"average_value"`
	TotalMovements      int                        `json:"total_movements"`
	InboundMovements    int                        `json:"inbound_movements"`
	OutboundMovements   int                        `json:"outbound_movements"`
	AdjustmentMovements int                        `json:"adjustment_movements"`
	LowStockItems       int                        `json:"low_stock_items"`
	OutOfStockItems     int                        `json:"out_of_stock_items"`
	ExpiringItems       int                        `json:"expiring_items"`
	TurnoverRate        float64                    `json:"turnover_rate"`
	StockAccuracy       float64                    `json:"stock_accuracy"`
	CategoryBreakdown   map[string]CategoryMetrics `json:"category_breakdown"`
	TopMovingItems      []ItemMovementMetric       `json:"top_moving_items"`
	SlowMovingItems     []ItemMovementMetric       `json:"slow_moving_items"`
	GeneratedAt         time.Time                  `json:"generated_at"`
}

// CategoryMetrics contains metrics for a specific category
type CategoryMetrics struct {
	Category      entities.ItemCategory `json:"category"`
	ItemCount     int                   `json:"item_count"`
	TotalValue    entities.Money        `json:"total_value"`
	AverageValue  entities.Money        `json:"average_value"`
	TurnoverRate  float64               `json:"turnover_rate"`
	LowStockCount int                   `json:"low_stock_count"`
}

// ItemMovementMetric contains movement metrics for an item
type ItemMovementMetric struct {
	ItemID       uuid.UUID `json:"item_id"`
	ItemName     string    `json:"item_name"`
	ItemSKU      string    `json:"item_sku"`
	MovementCount int      `json:"movement_count"`
	TotalQuantity float64  `json:"total_quantity"`
	TurnoverRate  float64  `json:"turnover_rate"`
	LastMovement  time.Time `json:"last_movement"`
}

// StockMovementRepository defines the interface for stock movement data access
type StockMovementRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, movement *entities.StockMovement) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.StockMovement, error)
	GetByMovementNumber(ctx context.Context, movementNumber string) (*entities.StockMovement, error)
	Update(ctx context.Context, movement *entities.StockMovement) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Listing and filtering
	List(ctx context.Context, filter *MovementFilter) ([]*entities.StockMovement, error)
	ListByItem(ctx context.Context, itemID uuid.UUID, filter *MovementFilter) ([]*entities.StockMovement, error)
	ListByLocation(ctx context.Context, locationID uuid.UUID, filter *MovementFilter) ([]*entities.StockMovement, error)
	ListByType(ctx context.Context, movementType entities.MovementType, filter *MovementFilter) ([]*entities.StockMovement, error)
	ListByStatus(ctx context.Context, status entities.MovementStatus, filter *MovementFilter) ([]*entities.StockMovement, error)
	
	// Processing operations
	GetPendingMovements(ctx context.Context, locationID *uuid.UUID) ([]*entities.StockMovement, error)
	GetScheduledMovements(ctx context.Context, before time.Time) ([]*entities.StockMovement, error)
	GetMovementsRequiringApproval(ctx context.Context) ([]*entities.StockMovement, error)
	
	// Analytics and reporting
	GetMovementSummary(ctx context.Context, itemID uuid.UUID, period time.Duration) (*entities.StockMovementSummary, error)
	GetMovementsByPeriod(ctx context.Context, start, end time.Time, filter *MovementFilter) ([]*entities.StockMovement, error)
	GetMovementTrends(ctx context.Context, period time.Duration, groupBy string) (map[string]float64, error)
	
	// Bulk operations
	BulkCreate(ctx context.Context, movements []*entities.StockMovement) error
	BulkUpdate(ctx context.Context, movements []*entities.StockMovement) error
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo StockMovementRepository) error) error
}

// MovementFilter defines filtering options for stock movement queries
type MovementFilter struct {
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
	Limit          int                          `json:"limit,omitempty"`
	Offset         int                          `json:"offset,omitempty"`
}

// SupplierRepository defines the interface for supplier data access
type SupplierRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, supplier *entities.Supplier) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Supplier, error)
	GetByCode(ctx context.Context, code string) (*entities.Supplier, error)
	Update(ctx context.Context, supplier *entities.Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Listing and filtering
	List(ctx context.Context, filter *SupplierFilter) ([]*entities.Supplier, error)
	ListByCategory(ctx context.Context, category entities.SupplierCategory) ([]*entities.Supplier, error)
	ListByType(ctx context.Context, supplierType entities.SupplierType) ([]*entities.Supplier, error)
	GetActiveSuppliers(ctx context.Context) ([]*entities.Supplier, error)
	GetPreferredSuppliers(ctx context.Context) ([]*entities.Supplier, error)
	
	// Product and pricing
	GetSupplierProducts(ctx context.Context, supplierID uuid.UUID) ([]*entities.SupplierProduct, error)
	GetSuppliersForProduct(ctx context.Context, internalSKU string) ([]*entities.Supplier, error)
	GetBestPriceSupplier(ctx context.Context, internalSKU string, quantity float64) (*entities.Supplier, error)
	
	// Performance tracking
	UpdatePerformance(ctx context.Context, supplierID uuid.UUID, performance *entities.SupplierPerformance) error
	GetTopPerformingSuppliers(ctx context.Context, limit int) ([]*entities.Supplier, error)
	
	// Search and advanced queries
	Search(ctx context.Context, query string, filter *SupplierFilter) ([]*entities.Supplier, error)
	GetSuppliersByTags(ctx context.Context, tags []string) ([]*entities.Supplier, error)
	
	// Bulk operations
	BulkCreate(ctx context.Context, suppliers []*entities.Supplier) error
	BulkUpdate(ctx context.Context, suppliers []*entities.Supplier) error
}

// SupplierFilter defines filtering options for supplier queries
type SupplierFilter struct {
	Types         []entities.SupplierType     `json:"types,omitempty"`
	Categories    []entities.SupplierCategory `json:"categories,omitempty"`
	Status        []entities.SupplierStatus   `json:"status,omitempty"`
	IsActive      *bool                       `json:"is_active,omitempty"`
	IsPreferred   *bool                       `json:"is_preferred,omitempty"`
	MinRating     *float64                    `json:"min_rating,omitempty"`
	Countries     []string                    `json:"countries,omitempty"`
	Tags          []string                    `json:"tags,omitempty"`
	CreatedAfter  *time.Time                  `json:"created_after,omitempty"`
	CreatedBefore *time.Time                  `json:"created_before,omitempty"`
	SortBy        string                      `json:"sort_by,omitempty"`
	SortOrder     string                      `json:"sort_order,omitempty"`
	Limit         int                         `json:"limit,omitempty"`
	Offset        int                         `json:"offset,omitempty"`
}

// LocationRepository defines the interface for location data access
type LocationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, location *entities.Location) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Location, error)
	GetByCode(ctx context.Context, code string) (*entities.Location, error)
	Update(ctx context.Context, location *entities.Location) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Listing and filtering
	List(ctx context.Context, filter *LocationFilter) ([]*entities.Location, error)
	ListByType(ctx context.Context, locationType entities.LocationType) ([]*entities.Location, error)
	GetActiveLocations(ctx context.Context) ([]*entities.Location, error)
	GetDefaultLocation(ctx context.Context) (*entities.Location, error)
	
	// Hierarchy operations
	GetChildLocations(ctx context.Context, parentID uuid.UUID) ([]*entities.Location, error)
	GetParentLocation(ctx context.Context, childID uuid.UUID) (*entities.Location, error)
	GetLocationHierarchy(ctx context.Context, rootID uuid.UUID) ([]*entities.Location, error)
	
	// Storage zones
	AddStorageZone(ctx context.Context, locationID uuid.UUID, zone *entities.StorageZone) error
	UpdateStorageZone(ctx context.Context, zoneID uuid.UUID, zone *entities.StorageZone) error
	GetStorageZone(ctx context.Context, zoneID uuid.UUID) (*entities.StorageZone, error)
	ListStorageZones(ctx context.Context, locationID uuid.UUID) ([]*entities.StorageZone, error)
	
	// Capacity management
	UpdateCapacity(ctx context.Context, locationID uuid.UUID, capacity *entities.LocationCapacity) error
	GetCapacityUtilization(ctx context.Context, locationID uuid.UUID) (float64, error)
	GetLocationsWithCapacity(ctx context.Context, requiredCapacity float64) ([]*entities.Location, error)
	
	// Search and advanced queries
	Search(ctx context.Context, query string, filter *LocationFilter) ([]*entities.Location, error)
	GetLocationsByTags(ctx context.Context, tags []string) ([]*entities.Location, error)
	GetNearbyLocations(ctx context.Context, latitude, longitude, radiusKm float64) ([]*entities.Location, error)
}

// LocationFilter defines filtering options for location queries
type LocationFilter struct {
	Types         []entities.LocationType   `json:"types,omitempty"`
	Status        []entities.LocationStatus `json:"status,omitempty"`
	IsActive      *bool                     `json:"is_active,omitempty"`
	IsDefault     *bool                     `json:"is_default,omitempty"`
	ParentID      *uuid.UUID                `json:"parent_id,omitempty"`
	Countries     []string                  `json:"countries,omitempty"`
	States        []string                  `json:"states,omitempty"`
	Cities        []string                  `json:"cities,omitempty"`
	Tags          []string                  `json:"tags,omitempty"`
	CreatedAfter  *time.Time                `json:"created_after,omitempty"`
	CreatedBefore *time.Time                `json:"created_before,omitempty"`
	SortBy        string                    `json:"sort_by,omitempty"`
	SortOrder     string                    `json:"sort_order,omitempty"`
	Limit         int                       `json:"limit,omitempty"`
	Offset        int                       `json:"offset,omitempty"`
}

// PurchaseOrderRepository defines the interface for purchase order data access
type PurchaseOrderRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, order *entities.PurchaseOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.PurchaseOrder, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*entities.PurchaseOrder, error)
	Update(ctx context.Context, order *entities.PurchaseOrder) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Listing and filtering
	List(ctx context.Context, filter *PurchaseOrderFilter) ([]*entities.PurchaseOrder, error)
	ListBySupplier(ctx context.Context, supplierID uuid.UUID, filter *PurchaseOrderFilter) ([]*entities.PurchaseOrder, error)
	ListByStatus(ctx context.Context, status entities.PurchaseOrderStatus, filter *PurchaseOrderFilter) ([]*entities.PurchaseOrder, error)
	ListByBuyer(ctx context.Context, buyerID uuid.UUID, filter *PurchaseOrderFilter) ([]*entities.PurchaseOrder, error)
	
	// Status-based queries
	GetPendingOrders(ctx context.Context) ([]*entities.PurchaseOrder, error)
	GetOverdueOrders(ctx context.Context) ([]*entities.PurchaseOrder, error)
	GetOrdersRequiringApproval(ctx context.Context) ([]*entities.PurchaseOrder, error)
	GetOrdersReadyToReceive(ctx context.Context) ([]*entities.PurchaseOrder, error)
	
	// Analytics and reporting
	GetOrderMetrics(ctx context.Context, period time.Duration) (*PurchaseOrderMetrics, error)
	GetSupplierOrderSummary(ctx context.Context, supplierID uuid.UUID, period time.Duration) (*SupplierOrderSummary, error)
	GetOrderTrends(ctx context.Context, period time.Duration, groupBy string) (map[string]interface{}, error)
	
	// Bulk operations
	BulkCreate(ctx context.Context, orders []*entities.PurchaseOrder) error
	BulkUpdate(ctx context.Context, orders []*entities.PurchaseOrder) error
}

// PurchaseOrderFilter defines filtering options for purchase order queries
type PurchaseOrderFilter struct {
	SupplierIDs    []uuid.UUID                      `json:"supplier_ids,omitempty"`
	LocationIDs    []uuid.UUID                      `json:"location_ids,omitempty"`
	BuyerIDs       []uuid.UUID                      `json:"buyer_ids,omitempty"`
	Status         []entities.PurchaseOrderStatus   `json:"status,omitempty"`
	Types          []entities.PurchaseOrderType     `json:"types,omitempty"`
	Priorities     []entities.OrderPriority         `json:"priorities,omitempty"`
	RequestedBy    []string                         `json:"requested_by,omitempty"`
	ApprovedBy     []string                         `json:"approved_by,omitempty"`
	MinAmount      *float64                         `json:"min_amount,omitempty"`
	MaxAmount      *float64                         `json:"max_amount,omitempty"`
	CreatedAfter   *time.Time                       `json:"created_after,omitempty"`
	CreatedBefore  *time.Time                       `json:"created_before,omitempty"`
	OrderedAfter   *time.Time                       `json:"ordered_after,omitempty"`
	OrderedBefore  *time.Time                       `json:"ordered_before,omitempty"`
	ExpectedAfter  *time.Time                       `json:"expected_after,omitempty"`
	ExpectedBefore *time.Time                       `json:"expected_before,omitempty"`
	Tags           []string                         `json:"tags,omitempty"`
	SortBy         string                           `json:"sort_by,omitempty"`
	SortOrder      string                           `json:"sort_order,omitempty"`
	Limit          int                              `json:"limit,omitempty"`
	Offset         int                              `json:"offset,omitempty"`
}

// PurchaseOrderMetrics contains purchase order metrics
type PurchaseOrderMetrics struct {
	Period           string             `json:"period"`
	TotalOrders      int                `json:"total_orders"`
	TotalValue       entities.Money     `json:"total_value"`
	AverageValue     entities.Money     `json:"average_value"`
	PendingOrders    int                `json:"pending_orders"`
	ApprovedOrders   int                `json:"approved_orders"`
	CompletedOrders  int                `json:"completed_orders"`
	CancelledOrders  int                `json:"cancelled_orders"`
	OverdueOrders    int                `json:"overdue_orders"`
	OnTimeDelivery   float64            `json:"on_time_delivery"`
	AverageLeadTime  time.Duration      `json:"average_lead_time"`
	TopSuppliers     []SupplierMetric   `json:"top_suppliers"`
	CategoryBreakdown map[string]float64 `json:"category_breakdown"`
	GeneratedAt      time.Time          `json:"generated_at"`
}

// SupplierOrderSummary contains order summary for a supplier
type SupplierOrderSummary struct {
	SupplierID      uuid.UUID      `json:"supplier_id"`
	SupplierName    string         `json:"supplier_name"`
	Period          string         `json:"period"`
	OrderCount      int            `json:"order_count"`
	TotalValue      entities.Money `json:"total_value"`
	AverageValue    entities.Money `json:"average_value"`
	OnTimeDelivery  float64        `json:"on_time_delivery"`
	AverageLeadTime time.Duration  `json:"average_lead_time"`
	QualityRating   float64        `json:"quality_rating"`
	LastOrderDate   *time.Time     `json:"last_order_date,omitempty"`
}

// SupplierMetric contains metrics for a supplier
type SupplierMetric struct {
	SupplierID   uuid.UUID      `json:"supplier_id"`
	SupplierName string         `json:"supplier_name"`
	OrderCount   int            `json:"order_count"`
	TotalValue   entities.Money `json:"total_value"`
	Performance  float64        `json:"performance"`
}
