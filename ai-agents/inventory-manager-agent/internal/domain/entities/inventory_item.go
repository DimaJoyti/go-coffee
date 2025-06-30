package entities

import (
	"time"

	"github.com/google/uuid"
)

// InventoryItem represents a comprehensive inventory item with all necessary attributes
type InventoryItem struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	SKU               string                 `json:"sku" redis:"sku"`
	Name              string                 `json:"name" redis:"name"`
	Description       string                 `json:"description" redis:"description"`
	Category          ItemCategory           `json:"category" redis:"category"`
	SubCategory       string                 `json:"sub_category" redis:"sub_category"`
	Unit              MeasurementUnit        `json:"unit" redis:"unit"`
	CurrentStock      float64                `json:"current_stock" redis:"current_stock"`
	ReservedStock     float64                `json:"reserved_stock" redis:"reserved_stock"`
	AvailableStock    float64                `json:"available_stock" redis:"available_stock"`
	MinimumLevel      float64                `json:"minimum_level" redis:"minimum_level"`
	MaximumLevel      float64                `json:"maximum_level" redis:"maximum_level"`
	ReorderPoint      float64                `json:"reorder_point" redis:"reorder_point"`
	ReorderQuantity   float64                `json:"reorder_quantity" redis:"reorder_quantity"`
	SafetyStock       float64                `json:"safety_stock" redis:"safety_stock"`
	UnitCost          Money                  `json:"unit_cost" redis:"unit_cost"`
	TotalValue        Money                  `json:"total_value" redis:"total_value"`
	AverageCost       Money                  `json:"average_cost" redis:"average_cost"`
	LastCost          Money                  `json:"last_cost" redis:"last_cost"`
	SupplierID        uuid.UUID              `json:"supplier_id" redis:"supplier_id"`
	PrimarySupplier   *Supplier              `json:"primary_supplier,omitempty"`
	AlternateSuppliers []*Supplier           `json:"alternate_suppliers,omitempty"`
	LocationID        uuid.UUID              `json:"location_id" redis:"location_id"`
	Location          *Location              `json:"location,omitempty"`
	Batches           []*InventoryBatch      `json:"batches,omitempty"`
	QualityInfo       *QualityInformation    `json:"quality_info,omitempty"`
	StorageConditions *StorageRequirements   `json:"storage_conditions,omitempty"`
	Attributes        map[string]interface{} `json:"attributes" redis:"attributes"`
	Tags              []string               `json:"tags" redis:"tags"`
	Status            ItemStatus             `json:"status" redis:"status"`
	IsActive          bool                   `json:"is_active" redis:"is_active"`
	IsPerishable      bool                   `json:"is_perishable" redis:"is_perishable"`
	ShelfLife         *time.Duration         `json:"shelf_life,omitempty" redis:"shelf_life"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         string                 `json:"created_by" redis:"created_by"`
	UpdatedBy         string                 `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// ItemCategory defines the category of inventory items
type ItemCategory string

const (
	CategoryRawMaterials   ItemCategory = "raw_materials"
	CategoryIngredients    ItemCategory = "ingredients"
	CategoryPackaging      ItemCategory = "packaging"
	CategorySupplies       ItemCategory = "supplies"
	CategoryEquipment      ItemCategory = "equipment"
	CategoryFinishedGoods  ItemCategory = "finished_goods"
	CategoryMaintenance    ItemCategory = "maintenance"
	CategoryCleaning       ItemCategory = "cleaning"
	CategoryOffice         ItemCategory = "office"
	CategoryMarketing      ItemCategory = "marketing"
)

// MeasurementUnit defines the unit of measurement for inventory items
type MeasurementUnit string

const (
	UnitKilogram    MeasurementUnit = "kg"
	UnitGram        MeasurementUnit = "g"
	UnitLiter       MeasurementUnit = "l"
	UnitMilliliter  MeasurementUnit = "ml"
	UnitPiece       MeasurementUnit = "pcs"
	UnitBox         MeasurementUnit = "box"
	UnitPack        MeasurementUnit = "pack"
	UnitBottle      MeasurementUnit = "bottle"
	UnitCan         MeasurementUnit = "can"
	UnitBag         MeasurementUnit = "bag"
	UnitCup         MeasurementUnit = "cup"
	UnitOunce       MeasurementUnit = "oz"
	UnitPound       MeasurementUnit = "lb"
)

// ItemStatus defines the status of an inventory item
type ItemStatus string

const (
	StatusActive      ItemStatus = "active"
	StatusInactive    ItemStatus = "inactive"
	StatusDiscontinued ItemStatus = "discontinued"
	StatusLowStock    ItemStatus = "low_stock"
	StatusOutOfStock  ItemStatus = "out_of_stock"
	StatusExpiring    ItemStatus = "expiring"
	StatusExpired     ItemStatus = "expired"
	StatusDamaged     ItemStatus = "damaged"
	StatusQuarantine  ItemStatus = "quarantine"
)

// Money represents monetary values with currency
type Money struct {
	Amount   float64 `json:"amount" redis:"amount"`
	Currency string  `json:"currency" redis:"currency"`
}

// InventoryBatch represents a batch of inventory items with tracking information
type InventoryBatch struct {
	ID              uuid.UUID              `json:"id" redis:"id"`
	BatchNumber     string                 `json:"batch_number" redis:"batch_number"`
	Quantity        float64                `json:"quantity" redis:"quantity"`
	UnitCost        Money                  `json:"unit_cost" redis:"unit_cost"`
	ExpirationDate  *time.Time             `json:"expiration_date,omitempty" redis:"expiration_date"`
	ManufactureDate *time.Time             `json:"manufacture_date,omitempty" redis:"manufacture_date"`
	ReceivedDate    time.Time              `json:"received_date" redis:"received_date"`
	SupplierID      uuid.UUID              `json:"supplier_id" redis:"supplier_id"`
	PurchaseOrderID *uuid.UUID             `json:"purchase_order_id,omitempty" redis:"purchase_order_id"`
	QualityStatus   QualityStatus          `json:"quality_status" redis:"quality_status"`
	QualityNotes    string                 `json:"quality_notes" redis:"quality_notes"`
	StorageLocation string                 `json:"storage_location" redis:"storage_location"`
	Attributes      map[string]interface{} `json:"attributes" redis:"attributes"`
	IsActive        bool                   `json:"is_active" redis:"is_active"`
	CreatedAt       time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" redis:"updated_at"`
}

// QualityStatus defines the quality status of inventory batches
type QualityStatus string

const (
	QualityPending   QualityStatus = "pending"
	QualityApproved  QualityStatus = "approved"
	QualityRejected  QualityStatus = "rejected"
	QualityQuarantine QualityStatus = "quarantine"
	QualityExpired   QualityStatus = "expired"
)

// QualityInformation contains quality-related information for inventory items
type QualityInformation struct {
	QualityGrade        string                 `json:"quality_grade" redis:"quality_grade"`
	CertificationLevel  string                 `json:"certification_level" redis:"certification_level"`
	Certifications      []string               `json:"certifications" redis:"certifications"`
	QualityChecks       []*QualityCheck        `json:"quality_checks,omitempty"`
	LastQualityCheck    *time.Time             `json:"last_quality_check,omitempty" redis:"last_quality_check"`
	NextQualityCheck    *time.Time             `json:"next_quality_check,omitempty" redis:"next_quality_check"`
	QualityScore        float64                `json:"quality_score" redis:"quality_score"`
	QualityNotes        string                 `json:"quality_notes" redis:"quality_notes"`
	Allergens           []string               `json:"allergens" redis:"allergens"`
	NutritionalInfo     map[string]interface{} `json:"nutritional_info" redis:"nutritional_info"`
	OriginCountry       string                 `json:"origin_country" redis:"origin_country"`
	OrganicCertified    bool                   `json:"organic_certified" redis:"organic_certified"`
	FairTradeCertified  bool                   `json:"fair_trade_certified" redis:"fair_trade_certified"`
}

// QualityCheck represents a quality check performed on inventory
type QualityCheck struct {
	ID          uuid.UUID              `json:"id" redis:"id"`
	CheckType   string                 `json:"check_type" redis:"check_type"`
	CheckDate   time.Time              `json:"check_date" redis:"check_date"`
	CheckedBy   string                 `json:"checked_by" redis:"checked_by"`
	Result      QualityStatus          `json:"result" redis:"result"`
	Score       float64                `json:"score" redis:"score"`
	Notes       string                 `json:"notes" redis:"notes"`
	Images      []string               `json:"images" redis:"images"`
	Attributes  map[string]interface{} `json:"attributes" redis:"attributes"`
	CreatedAt   time.Time              `json:"created_at" redis:"created_at"`
}

// StorageRequirements defines storage requirements for inventory items
type StorageRequirements struct {
	Temperature     *TemperatureRange      `json:"temperature,omitempty"`
	Humidity        *HumidityRange         `json:"humidity,omitempty"`
	LightConditions string                 `json:"light_conditions" redis:"light_conditions"`
	Ventilation     string                 `json:"ventilation" redis:"ventilation"`
	SpecialHandling []string               `json:"special_handling" redis:"special_handling"`
	StorageType     string                 `json:"storage_type" redis:"storage_type"`
	Hazardous       bool                   `json:"hazardous" redis:"hazardous"`
	HazardClass     string                 `json:"hazard_class" redis:"hazard_class"`
	Attributes      map[string]interface{} `json:"attributes" redis:"attributes"`
}

// TemperatureRange defines temperature storage requirements
type TemperatureRange struct {
	MinCelsius float64 `json:"min_celsius" redis:"min_celsius"`
	MaxCelsius float64 `json:"max_celsius" redis:"max_celsius"`
	Unit       string  `json:"unit" redis:"unit"`
}

// HumidityRange defines humidity storage requirements
type HumidityRange struct {
	MinPercent float64 `json:"min_percent" redis:"min_percent"`
	MaxPercent float64 `json:"max_percent" redis:"max_percent"`
}

// NewInventoryItem creates a new inventory item with default values
func NewInventoryItem(name, sku string, category ItemCategory, unit MeasurementUnit) *InventoryItem {
	now := time.Now()
	return &InventoryItem{
		ID:              uuid.New(),
		SKU:             sku,
		Name:            name,
		Category:        category,
		Unit:            unit,
		CurrentStock:    0,
		ReservedStock:   0,
		AvailableStock:  0,
		MinimumLevel:    0,
		MaximumLevel:    0,
		ReorderPoint:    0,
		ReorderQuantity: 0,
		SafetyStock:     0,
		UnitCost:        Money{Amount: 0, Currency: "USD"},
		TotalValue:      Money{Amount: 0, Currency: "USD"},
		AverageCost:     Money{Amount: 0, Currency: "USD"},
		LastCost:        Money{Amount: 0, Currency: "USD"},
		Attributes:      make(map[string]interface{}),
		Tags:            []string{},
		Status:          StatusActive,
		IsActive:        true,
		IsPerishable:    false,
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
	}
}

// UpdateStock updates the current stock level and recalculates derived values
func (item *InventoryItem) UpdateStock(newStock float64) {
	item.CurrentStock = newStock
	item.AvailableStock = item.CurrentStock - item.ReservedStock
	item.TotalValue = Money{
		Amount:   item.CurrentStock * item.UnitCost.Amount,
		Currency: item.UnitCost.Currency,
	}
	item.UpdatedAt = time.Now()
	item.Version++
	
	// Update status based on stock levels
	item.updateStatus()
}

// ReserveStock reserves a quantity of stock
func (item *InventoryItem) ReserveStock(quantity float64) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	
	if item.AvailableStock < quantity {
		return ErrInsufficientStock
	}
	
	item.ReservedStock += quantity
	item.AvailableStock = item.CurrentStock - item.ReservedStock
	item.UpdatedAt = time.Now()
	item.Version++
	
	return nil
}

// ReleaseStock releases reserved stock
func (item *InventoryItem) ReleaseStock(quantity float64) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	
	if item.ReservedStock < quantity {
		return ErrInvalidReservation
	}
	
	item.ReservedStock -= quantity
	item.AvailableStock = item.CurrentStock - item.ReservedStock
	item.UpdatedAt = time.Now()
	item.Version++
	
	return nil
}

// IsLowStock checks if the item is below minimum level
func (item *InventoryItem) IsLowStock() bool {
	return item.CurrentStock <= item.MinimumLevel
}

// IsOutOfStock checks if the item is out of stock
func (item *InventoryItem) IsOutOfStock() bool {
	return item.CurrentStock <= 0
}

// NeedsReorder checks if the item needs to be reordered
func (item *InventoryItem) NeedsReorder() bool {
	return item.CurrentStock <= item.ReorderPoint
}

// IsExpiring checks if any batch is expiring within the given duration
func (item *InventoryItem) IsExpiring(within time.Duration) bool {
	if !item.IsPerishable {
		return false
	}
	
	cutoff := time.Now().Add(within)
	for _, batch := range item.Batches {
		if batch.ExpirationDate != nil && batch.ExpirationDate.Before(cutoff) {
			return true
		}
	}
	return false
}

// GetExpiringBatches returns batches that are expiring within the given duration
func (item *InventoryItem) GetExpiringBatches(within time.Duration) []*InventoryBatch {
	if !item.IsPerishable {
		return nil
	}
	
	cutoff := time.Now().Add(within)
	var expiring []*InventoryBatch
	
	for _, batch := range item.Batches {
		if batch.ExpirationDate != nil && batch.ExpirationDate.Before(cutoff) {
			expiring = append(expiring, batch)
		}
	}
	
	return expiring
}

// updateStatus updates the item status based on current conditions
func (item *InventoryItem) updateStatus() {
	if !item.IsActive {
		item.Status = StatusInactive
		return
	}
	
	if item.IsOutOfStock() {
		item.Status = StatusOutOfStock
	} else if item.IsLowStock() {
		item.Status = StatusLowStock
	} else if item.IsExpiring(7 * 24 * time.Hour) { // 7 days
		item.Status = StatusExpiring
	} else {
		item.Status = StatusActive
	}
}

// CalculateAverageCost calculates the weighted average cost of all batches
func (item *InventoryItem) CalculateAverageCost() Money {
	if len(item.Batches) == 0 {
		return item.UnitCost
	}
	
	totalValue := 0.0
	totalQuantity := 0.0
	currency := item.UnitCost.Currency
	
	for _, batch := range item.Batches {
		if batch.IsActive {
			totalValue += batch.Quantity * batch.UnitCost.Amount
			totalQuantity += batch.Quantity
			if currency == "" {
				currency = batch.UnitCost.Currency
			}
		}
	}
	
	if totalQuantity == 0 {
		return Money{Amount: 0, Currency: currency}
	}
	
	return Money{
		Amount:   totalValue / totalQuantity,
		Currency: currency,
	}
}

// AddBatch adds a new batch to the inventory item
func (item *InventoryItem) AddBatch(batch *InventoryBatch) {
	item.Batches = append(item.Batches, batch)
	item.CurrentStock += batch.Quantity
	item.AvailableStock = item.CurrentStock - item.ReservedStock
	item.LastCost = batch.UnitCost
	item.AverageCost = item.CalculateAverageCost()
	item.TotalValue = Money{
		Amount:   item.CurrentStock * item.AverageCost.Amount,
		Currency: item.AverageCost.Currency,
	}
	item.UpdatedAt = time.Now()
	item.Version++
	item.updateStatus()
}

// Domain errors
var (
	ErrInvalidQuantity    = NewDomainError("INVALID_QUANTITY", "Quantity must be greater than zero")
	ErrInsufficientStock  = NewDomainError("INSUFFICIENT_STOCK", "Insufficient stock available")
	ErrInvalidReservation = NewDomainError("INVALID_RESERVATION", "Invalid reservation quantity")
	ErrItemNotFound       = NewDomainError("ITEM_NOT_FOUND", "Inventory item not found")
	ErrItemInactive       = NewDomainError("ITEM_INACTIVE", "Inventory item is inactive")
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}
