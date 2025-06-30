package commands

import (
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/entities"
)

// CreateInventoryItemCommand represents a command to create a new inventory item
type CreateInventoryItemCommand struct {
	SKU               string                        `json:"sku" validate:"required,min=1,max=50"`
	Name              string                        `json:"name" validate:"required,min=1,max=200"`
	Description       string                        `json:"description,omitempty" validate:"max=1000"`
	Category          entities.ItemCategory         `json:"category" validate:"required"`
	SubCategory       string                        `json:"sub_category,omitempty" validate:"max=100"`
	Unit              entities.MeasurementUnit      `json:"unit" validate:"required"`
	MinimumLevel      float64                       `json:"minimum_level" validate:"min=0"`
	MaximumLevel      float64                       `json:"maximum_level" validate:"min=0"`
	ReorderPoint      float64                       `json:"reorder_point" validate:"min=0"`
	ReorderQuantity   float64                       `json:"reorder_quantity" validate:"min=0"`
	SafetyStock       float64                       `json:"safety_stock" validate:"min=0"`
	UnitCost          entities.Money                `json:"unit_cost"`
	LocationID        uuid.UUID                     `json:"location_id" validate:"required"`
	SupplierID        uuid.UUID                     `json:"supplier_id" validate:"required"`
	IsPerishable      bool                          `json:"is_perishable"`
	ShelfLife         *time.Duration                `json:"shelf_life,omitempty"`
	StorageConditions *entities.StorageRequirements `json:"storage_conditions,omitempty"`
	Attributes        map[string]interface{}        `json:"attributes,omitempty"`
	Tags              []string                      `json:"tags,omitempty"`
	CreatedBy         string                        `json:"created_by" validate:"required"`
}

// UpdateInventoryItemCommand represents a command to update an inventory item
type UpdateInventoryItemCommand struct {
	ID                uuid.UUID                     `json:"id" validate:"required"`
	Name              *string                       `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Description       *string                       `json:"description,omitempty" validate:"omitempty,max=1000"`
	MinimumLevel      *float64                      `json:"minimum_level,omitempty" validate:"omitempty,min=0"`
	MaximumLevel      *float64                      `json:"maximum_level,omitempty" validate:"omitempty,min=0"`
	ReorderPoint      *float64                      `json:"reorder_point,omitempty" validate:"omitempty,min=0"`
	ReorderQuantity   *float64                      `json:"reorder_quantity,omitempty" validate:"omitempty,min=0"`
	SafetyStock       *float64                      `json:"safety_stock,omitempty" validate:"omitempty,min=0"`
	UnitCost          *entities.Money               `json:"unit_cost,omitempty"`
	SupplierID        *uuid.UUID                    `json:"supplier_id,omitempty"`
	StorageConditions *entities.StorageRequirements `json:"storage_conditions,omitempty"`
	Attributes        map[string]interface{}        `json:"attributes,omitempty"`
	Tags              []string                      `json:"tags,omitempty"`
	IsActive          *bool                         `json:"is_active,omitempty"`
	UpdatedBy         string                        `json:"updated_by" validate:"required"`
}

// ReceiveStockCommand represents a command to receive stock
type ReceiveStockCommand struct {
	ItemID          uuid.UUID      `json:"item_id" validate:"required"`
	Quantity        float64        `json:"quantity" validate:"required,gt=0"`
	UnitCost        entities.Money `json:"unit_cost" validate:"required"`
	SupplierID      *uuid.UUID     `json:"supplier_id,omitempty"`
	LocationID      *uuid.UUID     `json:"location_id,omitempty"`
	BatchNumber     string         `json:"batch_number,omitempty" validate:"max=100"`
	ExpirationDate  *time.Time     `json:"expiration_date,omitempty"`
	ManufactureDate *time.Time     `json:"manufacture_date,omitempty"`
	Reason          string         `json:"reason" validate:"required,min=1,max=500"`
	ReceivedBy      string         `json:"received_by" validate:"required"`
	QualityCheck    *QualityCheckData `json:"quality_check,omitempty"`
	Documents       []DocumentData `json:"documents,omitempty"`
}

// IssueStockCommand represents a command to issue stock
type IssueStockCommand struct {
	ItemID          uuid.UUID  `json:"item_id" validate:"required"`
	Quantity        float64    `json:"quantity" validate:"required,gt=0"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	ReferenceType   string     `json:"reference_type,omitempty" validate:"max=50"`
	ReferenceID     *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceNumber string     `json:"reference_number,omitempty" validate:"max=100"`
	Reason          string     `json:"reason" validate:"required,min=1,max=500"`
	IssuedBy        string     `json:"issued_by" validate:"required"`
	IssuedTo        string     `json:"issued_to,omitempty" validate:"max=200"`
	Notes           string     `json:"notes,omitempty" validate:"max=1000"`
}

// TransferStockCommand represents a command to transfer stock between locations
type TransferStockCommand struct {
	ItemID         uuid.UUID `json:"item_id" validate:"required"`
	Quantity       float64   `json:"quantity" validate:"required,gt=0"`
	FromLocationID uuid.UUID `json:"from_location_id" validate:"required"`
	ToLocationID   uuid.UUID `json:"to_location_id" validate:"required,nefield=FromLocationID"`
	FromZone       string    `json:"from_zone,omitempty" validate:"max=100"`
	ToZone         string    `json:"to_zone,omitempty" validate:"max=100"`
	Reason         string    `json:"reason" validate:"required,min=1,max=500"`
	TransferredBy  string    `json:"transferred_by" validate:"required"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	Priority       string    `json:"priority,omitempty" validate:"omitempty,oneof=low normal high urgent"`
	Notes          string    `json:"notes,omitempty" validate:"max=1000"`
}

// AdjustStockCommand represents a command to adjust stock levels
type AdjustStockCommand struct {
	ItemID     uuid.UUID  `json:"item_id" validate:"required"`
	Adjustment float64    `json:"adjustment" validate:"required,ne=0"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
	Reason     string     `json:"reason" validate:"required,min=1,max=500"`
	AdjustedBy string     `json:"adjusted_by" validate:"required"`
	CountDate  *time.Time `json:"count_date,omitempty"`
	CountedBy  string     `json:"counted_by,omitempty" validate:"max=200"`
	Notes      string     `json:"notes,omitempty" validate:"max=1000"`
}

// ReserveStockCommand represents a command to reserve stock
type ReserveStockCommand struct {
	ItemID          uuid.UUID  `json:"item_id" validate:"required"`
	Quantity        float64    `json:"quantity" validate:"required,gt=0"`
	ReservedFor     string     `json:"reserved_for" validate:"required,min=1,max=200"`
	ReferenceType   string     `json:"reference_type,omitempty" validate:"max=50"`
	ReferenceID     *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceNumber string     `json:"reference_number,omitempty" validate:"max=100"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	Priority        string     `json:"priority,omitempty" validate:"omitempty,oneof=low normal high urgent"`
	Notes           string     `json:"notes,omitempty" validate:"max=1000"`
	ReservedBy      string     `json:"reserved_by" validate:"required"`
}

// ReleaseStockCommand represents a command to release reserved stock
type ReleaseStockCommand struct {
	ItemID        uuid.UUID  `json:"item_id" validate:"required"`
	Quantity      float64    `json:"quantity" validate:"required,gt=0"`
	ReservationID *uuid.UUID `json:"reservation_id,omitempty"`
	Reason        string     `json:"reason" validate:"required,min=1,max=500"`
	ReleasedBy    string     `json:"released_by" validate:"required"`
	Notes         string     `json:"notes,omitempty" validate:"max=1000"`
}

// ProcessAutomaticReordersCommand represents a command to process automatic reorders
type ProcessAutomaticReordersCommand struct {
	LocationID  *uuid.UUID            `json:"location_id,omitempty"`
	Criteria    *AutoReorderCriteria  `json:"criteria,omitempty"`
	ProcessedBy string                `json:"processed_by" validate:"required"`
	DryRun      bool                  `json:"dry_run"`
	Categories  []entities.ItemCategory `json:"categories,omitempty"`
	SupplierIDs []uuid.UUID           `json:"supplier_ids,omitempty"`
}

// BulkUpdateStockCommand represents a command to update multiple stock levels
type BulkUpdateStockCommand struct {
	Updates   []StockUpdateItem `json:"updates" validate:"required,min=1,max=1000"`
	Reason    string            `json:"reason" validate:"required,min=1,max=500"`
	UpdatedBy string            `json:"updated_by" validate:"required"`
	Source    string            `json:"source,omitempty" validate:"max=100"`
	Notes     string            `json:"notes,omitempty" validate:"max=1000"`
}

// StockUpdateItem represents a single stock update in a bulk operation
type StockUpdateItem struct {
	ItemID    uuid.UUID `json:"item_id" validate:"required"`
	NewStock  float64   `json:"new_stock" validate:"min=0"`
	Reason    string    `json:"reason,omitempty" validate:"max=500"`
	Reference string    `json:"reference,omitempty" validate:"max=100"`
}

// CreateLocationCommand represents a command to create a new location
type CreateLocationCommand struct {
	Code        string                        `json:"code" validate:"required,min=1,max=50"`
	Name        string                        `json:"name" validate:"required,min=1,max=200"`
	Type        entities.LocationType         `json:"type" validate:"required"`
	ParentID    *uuid.UUID                    `json:"parent_id,omitempty"`
	Address     *entities.Address             `json:"address,omitempty"`
	Coordinates *entities.Coordinates         `json:"coordinates,omitempty"`
	Capacity    *entities.LocationCapacity    `json:"capacity,omitempty"`
	Environment *entities.EnvironmentConditions `json:"environment,omitempty"`
	Security    *entities.SecurityInformation `json:"security,omitempty"`
	Attributes  map[string]interface{}        `json:"attributes,omitempty"`
	Tags        []string                      `json:"tags,omitempty"`
	CreatedBy   string                        `json:"created_by" validate:"required"`
}

// UpdateLocationCommand represents a command to update a location
type UpdateLocationCommand struct {
	ID          uuid.UUID                     `json:"id" validate:"required"`
	Name        *string                       `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Status      *entities.LocationStatus      `json:"status,omitempty"`
	Address     *entities.Address             `json:"address,omitempty"`
	Coordinates *entities.Coordinates         `json:"coordinates,omitempty"`
	Capacity    *entities.LocationCapacity    `json:"capacity,omitempty"`
	Environment *entities.EnvironmentConditions `json:"environment,omitempty"`
	Security    *entities.SecurityInformation `json:"security,omitempty"`
	Attributes  map[string]interface{}        `json:"attributes,omitempty"`
	Tags        []string                      `json:"tags,omitempty"`
	IsActive    *bool                         `json:"is_active,omitempty"`
	UpdatedBy   string                        `json:"updated_by" validate:"required"`
}

// CreateSupplierCommand represents a command to create a new supplier
type CreateSupplierCommand struct {
	Code            string                      `json:"code" validate:"required,min=1,max=50"`
	Name            string                      `json:"name" validate:"required,min=1,max=200"`
	LegalName       string                      `json:"legal_name,omitempty" validate:"max=200"`
	Type            entities.SupplierType       `json:"type" validate:"required"`
	Category        entities.SupplierCategory   `json:"category" validate:"required"`
	ContactInfo     *entities.ContactInformation `json:"contact_info,omitempty"`
	Address         *entities.Address           `json:"address,omitempty"`
	BillingAddress  *entities.Address           `json:"billing_address,omitempty"`
	PaymentTerms    *entities.PaymentTerms      `json:"payment_terms,omitempty"`
	DeliveryTerms   *entities.DeliveryTerms     `json:"delivery_terms,omitempty"`
	TaxInfo         *entities.TaxInformation    `json:"tax_info,omitempty"`
	BankInfo        *entities.BankInformation   `json:"bank_info,omitempty"`
	Certifications  []*entities.Certification   `json:"certifications,omitempty"`
	Attributes      map[string]interface{}      `json:"attributes,omitempty"`
	Tags            []string                    `json:"tags,omitempty"`
	Notes           string                      `json:"notes,omitempty" validate:"max=2000"`
	CreatedBy       string                      `json:"created_by" validate:"required"`
}

// UpdateSupplierCommand represents a command to update a supplier
type UpdateSupplierCommand struct {
	ID              uuid.UUID                   `json:"id" validate:"required"`
	Name            *string                     `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	LegalName       *string                     `json:"legal_name,omitempty" validate:"omitempty,max=200"`
	Status          *entities.SupplierStatus    `json:"status,omitempty"`
	Rating          *float64                    `json:"rating,omitempty" validate:"omitempty,min=0,max=5"`
	ContactInfo     *entities.ContactInformation `json:"contact_info,omitempty"`
	Address         *entities.Address           `json:"address,omitempty"`
	BillingAddress  *entities.Address           `json:"billing_address,omitempty"`
	PaymentTerms    *entities.PaymentTerms      `json:"payment_terms,omitempty"`
	DeliveryTerms   *entities.DeliveryTerms     `json:"delivery_terms,omitempty"`
	TaxInfo         *entities.TaxInformation    `json:"tax_info,omitempty"`
	BankInfo        *entities.BankInformation   `json:"bank_info,omitempty"`
	Attributes      map[string]interface{}      `json:"attributes,omitempty"`
	Tags            []string                    `json:"tags,omitempty"`
	Notes           *string                     `json:"notes,omitempty" validate:"omitempty,max=2000"`
	IsActive        *bool                       `json:"is_active,omitempty"`
	IsPreferred     *bool                       `json:"is_preferred,omitempty"`
	UpdatedBy       string                      `json:"updated_by" validate:"required"`
}

// Supporting data structures

// QualityCheckData represents quality check information
type QualityCheckData struct {
	CheckType   string                 `json:"check_type" validate:"required"`
	CheckedBy   string                 `json:"checked_by" validate:"required"`
	Result      entities.QualityStatus `json:"result" validate:"required"`
	Score       float64                `json:"score" validate:"min=0,max=100"`
	Notes       string                 `json:"notes,omitempty" validate:"max=1000"`
	Images      []string               `json:"images,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// DocumentData represents document information
type DocumentData struct {
	Type        string `json:"type" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url" validate:"required,url"`
	MimeType    string `json:"mime_type,omitempty"`
	Size        int64  `json:"size,omitempty" validate:"min=0"`
	Checksum    string `json:"checksum,omitempty"`
}

// AutoReorderCriteria represents criteria for automatic reordering
type AutoReorderCriteria struct {
	MaxOrderValue     *float64 `json:"max_order_value,omitempty" validate:"omitempty,gt=0"`
	MinUrgencyScore   *float64 `json:"min_urgency_score,omitempty" validate:"omitempty,min=0,max=100"`
	RequireApproval   *bool    `json:"require_approval,omitempty"`
	PreferredSuppliers []uuid.UUID `json:"preferred_suppliers,omitempty"`
	MaxLeadTime       *int     `json:"max_lead_time,omitempty" validate:"omitempty,gt=0"`
	MinQualityRating  *float64 `json:"min_quality_rating,omitempty" validate:"omitempty,min=0,max=5"`
}

// Validation helper methods

// Validate validates the CreateInventoryItemCommand
func (cmd *CreateInventoryItemCommand) Validate() error {
	if cmd.MaximumLevel > 0 && cmd.MinimumLevel >= cmd.MaximumLevel {
		return entities.NewDomainError("INVALID_LEVELS", "Maximum level must be greater than minimum level")
	}
	
	if cmd.ReorderPoint > 0 && cmd.ReorderPoint <= cmd.MinimumLevel {
		return entities.NewDomainError("INVALID_REORDER_POINT", "Reorder point should be greater than minimum level")
	}
	
	if cmd.UnitCost.Amount < 0 {
		return entities.NewDomainError("INVALID_COST", "Unit cost cannot be negative")
	}
	
	return nil
}

// Validate validates the TransferStockCommand
func (cmd *TransferStockCommand) Validate() error {
	if cmd.FromLocationID == cmd.ToLocationID {
		return entities.NewDomainError("SAME_LOCATION", "From and to locations cannot be the same")
	}
	
	return nil
}

// Validate validates the BulkUpdateStockCommand
func (cmd *BulkUpdateStockCommand) Validate() error {
	if len(cmd.Updates) == 0 {
		return entities.NewDomainError("NO_UPDATES", "At least one update is required")
	}
	
	if len(cmd.Updates) > 1000 {
		return entities.NewDomainError("TOO_MANY_UPDATES", "Maximum 1000 updates allowed per batch")
	}
	
	// Check for duplicate item IDs
	seen := make(map[uuid.UUID]bool)
	for _, update := range cmd.Updates {
		if seen[update.ItemID] {
			return entities.NewDomainError("DUPLICATE_ITEM", "Duplicate item ID in updates")
		}
		seen[update.ItemID] = true
	}
	
	return nil
}
