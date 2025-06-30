package entities

import (
	"time"

	"github.com/google/uuid"
)

// StockMovement represents a movement of inventory items
type StockMovement struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	MovementNumber    string                 `json:"movement_number" redis:"movement_number"`
	Type              MovementType           `json:"type" redis:"type"`
	Status            MovementStatus         `json:"status" redis:"status"`
	Direction         MovementDirection      `json:"direction" redis:"direction"`
	InventoryItemID   uuid.UUID              `json:"inventory_item_id" redis:"inventory_item_id"`
	InventoryItem     *InventoryItem         `json:"inventory_item,omitempty"`
	BatchID           *uuid.UUID             `json:"batch_id,omitempty" redis:"batch_id"`
	Batch             *InventoryBatch        `json:"batch,omitempty"`
	Quantity          float64                `json:"quantity" redis:"quantity"`
	Unit              MeasurementUnit        `json:"unit" redis:"unit"`
	UnitCost          Money                  `json:"unit_cost" redis:"unit_cost"`
	TotalCost         Money                  `json:"total_cost" redis:"total_cost"`
	FromLocationID    *uuid.UUID             `json:"from_location_id,omitempty" redis:"from_location_id"`
	FromLocation      *Location              `json:"from_location,omitempty"`
	ToLocationID      *uuid.UUID             `json:"to_location_id,omitempty" redis:"to_location_id"`
	ToLocation        *Location              `json:"to_location,omitempty"`
	FromZone          string                 `json:"from_zone" redis:"from_zone"`
	ToZone            string                 `json:"to_zone" redis:"to_zone"`
	ReasonCode        string                 `json:"reason_code" redis:"reason_code"`
	Reason            string                 `json:"reason" redis:"reason"`
	ReferenceType     string                 `json:"reference_type" redis:"reference_type"`
	ReferenceID       *uuid.UUID             `json:"reference_id,omitempty" redis:"reference_id"`
	ReferenceNumber   string                 `json:"reference_number" redis:"reference_number"`
	SupplierID        *uuid.UUID             `json:"supplier_id,omitempty" redis:"supplier_id"`
	Supplier          *Supplier              `json:"supplier,omitempty"`
	CustomerID        *uuid.UUID             `json:"customer_id,omitempty" redis:"customer_id"`
	QualityCheck      *QualityCheck          `json:"quality_check,omitempty"`
	ExpirationDate    *time.Time             `json:"expiration_date,omitempty" redis:"expiration_date"`
	SerialNumbers     []string               `json:"serial_numbers" redis:"serial_numbers"`
	LotNumbers        []string               `json:"lot_numbers" redis:"lot_numbers"`
	Documents         []*MovementDocument    `json:"documents,omitempty"`
	Approvals         []*MovementApproval    `json:"approvals,omitempty"`
	Attributes        map[string]interface{} `json:"attributes" redis:"attributes"`
	Tags              []string               `json:"tags" redis:"tags"`
	Notes             string                 `json:"notes" redis:"notes"`
	ProcessedAt       *time.Time             `json:"processed_at,omitempty" redis:"processed_at"`
	ProcessedBy       string                 `json:"processed_by" redis:"processed_by"`
	ScheduledAt       *time.Time             `json:"scheduled_at,omitempty" redis:"scheduled_at"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty" redis:"completed_at"`
	CancelledAt       *time.Time             `json:"cancelled_at,omitempty" redis:"cancelled_at"`
	CancellationReason string                `json:"cancellation_reason" redis:"cancellation_reason"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         string                 `json:"created_by" redis:"created_by"`
	UpdatedBy         string                 `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// MovementType defines the type of stock movement
type MovementType string

const (
	MovementTypeReceipt     MovementType = "receipt"
	MovementTypeIssue       MovementType = "issue"
	MovementTypeTransfer    MovementType = "transfer"
	MovementTypeAdjustment  MovementType = "adjustment"
	MovementTypeReturn      MovementType = "return"
	MovementTypeWaste       MovementType = "waste"
	MovementTypeDamage      MovementType = "damage"
	MovementTypeExpiry      MovementType = "expiry"
	MovementTypeProduction  MovementType = "production"
	MovementTypeConsumption MovementType = "consumption"
	MovementTypeSale        MovementType = "sale"
	MovementTypePurchase    MovementType = "purchase"
	MovementTypeReservation MovementType = "reservation"
	MovementTypeRelease     MovementType = "release"
	MovementTypeCount       MovementType = "count"
)

// MovementStatus defines the status of a stock movement
type MovementStatus string

const (
	MovementStatusPending   MovementStatus = "pending"
	MovementStatusScheduled MovementStatus = "scheduled"
	MovementStatusProcessing MovementStatus = "processing"
	MovementStatusCompleted MovementStatus = "completed"
	MovementStatusCancelled MovementStatus = "cancelled"
	MovementStatusFailed    MovementStatus = "failed"
	MovementStatusOnHold    MovementStatus = "on_hold"
	MovementStatusApproved  MovementStatus = "approved"
	MovementStatusRejected  MovementStatus = "rejected"
)

// MovementDirection defines the direction of stock movement
type MovementDirection string

const (
	DirectionIn    MovementDirection = "in"
	DirectionOut   MovementDirection = "out"
	DirectionMove  MovementDirection = "move"
	DirectionAdjust MovementDirection = "adjust"
)

// MovementDocument represents a document associated with a stock movement
type MovementDocument struct {
	ID           uuid.UUID `json:"id" redis:"id"`
	Type         string    `json:"type" redis:"type"`
	Name         string    `json:"name" redis:"name"`
	Description  string    `json:"description" redis:"description"`
	URL          string    `json:"url" redis:"url"`
	MimeType     string    `json:"mime_type" redis:"mime_type"`
	Size         int64     `json:"size" redis:"size"`
	Checksum     string    `json:"checksum" redis:"checksum"`
	IsRequired   bool      `json:"is_required" redis:"is_required"`
	UploadedBy   string    `json:"uploaded_by" redis:"uploaded_by"`
	UploadedAt   time.Time `json:"uploaded_at" redis:"uploaded_at"`
}

// MovementApproval represents an approval for a stock movement
type MovementApproval struct {
	ID           uuid.UUID  `json:"id" redis:"id"`
	ApproverID   uuid.UUID  `json:"approver_id" redis:"approver_id"`
	ApproverName string     `json:"approver_name" redis:"approver_name"`
	Status       string     `json:"status" redis:"status"`
	Comments     string     `json:"comments" redis:"comments"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty" redis:"approved_at"`
	RejectedAt   *time.Time `json:"rejected_at,omitempty" redis:"rejected_at"`
	CreatedAt    time.Time  `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" redis:"updated_at"`
}

// StockMovementSummary provides a summary of stock movements
type StockMovementSummary struct {
	InventoryItemID uuid.UUID              `json:"inventory_item_id"`
	ItemName        string                 `json:"item_name"`
	ItemSKU         string                 `json:"item_sku"`
	Period          string                 `json:"period"`
	OpeningStock    float64                `json:"opening_stock"`
	ClosingStock    float64                `json:"closing_stock"`
	TotalIn         float64                `json:"total_in"`
	TotalOut        float64                `json:"total_out"`
	NetMovement     float64                `json:"net_movement"`
	MovementsByType map[MovementType]float64 `json:"movements_by_type"`
	TotalValue      Money                  `json:"total_value"`
	AverageCost     Money                  `json:"average_cost"`
	MovementCount   int                    `json:"movement_count"`
	LastMovement    *time.Time             `json:"last_movement,omitempty"`
}

// NewStockMovement creates a new stock movement
func NewStockMovement(
	movementType MovementType,
	direction MovementDirection,
	inventoryItemID uuid.UUID,
	quantity float64,
	unit MeasurementUnit,
	reason string,
) *StockMovement {
	now := time.Now()
	return &StockMovement{
		ID:              uuid.New(),
		MovementNumber:  generateMovementNumber(),
		Type:            movementType,
		Status:          MovementStatusPending,
		Direction:       direction,
		InventoryItemID: inventoryItemID,
		Quantity:        quantity,
		Unit:            unit,
		Reason:          reason,
		UnitCost:        Money{Amount: 0, Currency: "USD"},
		TotalCost:       Money{Amount: 0, Currency: "USD"},
		Attributes:      make(map[string]interface{}),
		Tags:            []string{},
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
	}
}

// generateMovementNumber generates a unique movement number
func generateMovementNumber() string {
	return "MOV-" + time.Now().Format("20060102-150405") + "-" + uuid.New().String()[:8]
}

// SetCost sets the unit cost and calculates total cost
func (sm *StockMovement) SetCost(unitCost Money) {
	sm.UnitCost = unitCost
	sm.TotalCost = Money{
		Amount:   sm.Quantity * unitCost.Amount,
		Currency: unitCost.Currency,
	}
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// SetLocations sets the from and to locations
func (sm *StockMovement) SetLocations(fromLocationID, toLocationID *uuid.UUID) {
	sm.FromLocationID = fromLocationID
	sm.ToLocationID = toLocationID
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// SetZones sets the from and to zones
func (sm *StockMovement) SetZones(fromZone, toZone string) {
	sm.FromZone = fromZone
	sm.ToZone = toZone
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// SetReference sets the reference information
func (sm *StockMovement) SetReference(refType string, refID *uuid.UUID, refNumber string) {
	sm.ReferenceType = refType
	sm.ReferenceID = refID
	sm.ReferenceNumber = refNumber
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// SetSupplier sets the supplier information
func (sm *StockMovement) SetSupplier(supplierID uuid.UUID) {
	sm.SupplierID = &supplierID
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// SetBatch sets the batch information
func (sm *StockMovement) SetBatch(batchID uuid.UUID) {
	sm.BatchID = &batchID
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// SetExpirationDate sets the expiration date
func (sm *StockMovement) SetExpirationDate(expirationDate time.Time) {
	sm.ExpirationDate = &expirationDate
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// AddSerialNumber adds a serial number
func (sm *StockMovement) AddSerialNumber(serialNumber string) {
	sm.SerialNumbers = append(sm.SerialNumbers, serialNumber)
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// AddLotNumber adds a lot number
func (sm *StockMovement) AddLotNumber(lotNumber string) {
	sm.LotNumbers = append(sm.LotNumbers, lotNumber)
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// AddDocument adds a document to the movement
func (sm *StockMovement) AddDocument(document *MovementDocument) {
	sm.Documents = append(sm.Documents, document)
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// Schedule schedules the movement for a specific time
func (sm *StockMovement) Schedule(scheduledAt time.Time) {
	sm.ScheduledAt = &scheduledAt
	sm.Status = MovementStatusScheduled
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// Process marks the movement as being processed
func (sm *StockMovement) Process(processedBy string) {
	now := time.Now()
	sm.Status = MovementStatusProcessing
	sm.ProcessedAt = &now
	sm.ProcessedBy = processedBy
	sm.UpdatedAt = now
	sm.Version++
}

// Complete marks the movement as completed
func (sm *StockMovement) Complete() {
	now := time.Now()
	sm.Status = MovementStatusCompleted
	sm.CompletedAt = &now
	sm.UpdatedAt = now
	sm.Version++
}

// Cancel cancels the movement
func (sm *StockMovement) Cancel(reason string) {
	now := time.Now()
	sm.Status = MovementStatusCancelled
	sm.CancelledAt = &now
	sm.CancellationReason = reason
	sm.UpdatedAt = now
	sm.Version++
}

// Fail marks the movement as failed
func (sm *StockMovement) Fail(reason string) {
	sm.Status = MovementStatusFailed
	sm.CancellationReason = reason
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// Hold puts the movement on hold
func (sm *StockMovement) Hold(reason string) {
	sm.Status = MovementStatusOnHold
	sm.Notes = reason
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// Resume resumes a held movement
func (sm *StockMovement) Resume() {
	sm.Status = MovementStatusPending
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// AddApproval adds an approval to the movement
func (sm *StockMovement) AddApproval(approval *MovementApproval) {
	sm.Approvals = append(sm.Approvals, approval)
	sm.UpdatedAt = time.Now()
	sm.Version++
}

// IsApproved checks if the movement is approved
func (sm *StockMovement) IsApproved() bool {
	if len(sm.Approvals) == 0 {
		return true // No approvals required
	}
	
	for _, approval := range sm.Approvals {
		if approval.Status != "approved" {
			return false
		}
	}
	return true
}

// RequiresApproval checks if the movement requires approval
func (sm *StockMovement) RequiresApproval() bool {
	// Define business rules for when approval is required
	// For example, high-value movements, certain types, etc.
	
	// High-value movements require approval
	if sm.TotalCost.Amount > 1000 {
		return true
	}
	
	// Certain movement types require approval
	switch sm.Type {
	case MovementTypeAdjustment, MovementTypeWaste, MovementTypeDamage:
		return true
	}
	
	return false
}

// CanProcess checks if the movement can be processed
func (sm *StockMovement) CanProcess() bool {
	if sm.Status != MovementStatusPending && sm.Status != MovementStatusScheduled {
		return false
	}
	
	if sm.RequiresApproval() && !sm.IsApproved() {
		return false
	}
	
	return true
}

// IsCompleted checks if the movement is completed
func (sm *StockMovement) IsCompleted() bool {
	return sm.Status == MovementStatusCompleted
}

// IsCancelled checks if the movement is cancelled
func (sm *StockMovement) IsCancelled() bool {
	return sm.Status == MovementStatusCancelled
}

// IsFailed checks if the movement failed
func (sm *StockMovement) IsFailed() bool {
	return sm.Status == MovementStatusFailed
}

// IsInbound checks if this is an inbound movement
func (sm *StockMovement) IsInbound() bool {
	return sm.Direction == DirectionIn
}

// IsOutbound checks if this is an outbound movement
func (sm *StockMovement) IsOutbound() bool {
	return sm.Direction == DirectionOut
}

// IsTransfer checks if this is a transfer movement
func (sm *StockMovement) IsTransfer() bool {
	return sm.Direction == DirectionMove
}

// IsAdjustment checks if this is an adjustment movement
func (sm *StockMovement) IsAdjustment() bool {
	return sm.Direction == DirectionAdjust
}

// GetEffectiveQuantity returns the effective quantity considering direction
func (sm *StockMovement) GetEffectiveQuantity() float64 {
	switch sm.Direction {
	case DirectionIn:
		return sm.Quantity
	case DirectionOut:
		return -sm.Quantity
	case DirectionAdjust:
		return sm.Quantity // Can be positive or negative
	default:
		return sm.Quantity
	}
}

// Validate validates the stock movement
func (sm *StockMovement) Validate() error {
	if sm.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	
	if sm.InventoryItemID == uuid.Nil {
		return ErrInvalidInventoryItem
	}
	
	// Validate transfer movements have both locations
	if sm.Type == MovementTypeTransfer {
		if sm.FromLocationID == nil || sm.ToLocationID == nil {
			return ErrInvalidTransferLocations
		}
		if *sm.FromLocationID == *sm.ToLocationID {
			return ErrSameTransferLocation
		}
	}
	
	// Validate receipt movements have supplier or from location
	if sm.Type == MovementTypeReceipt {
		if sm.SupplierID == nil && sm.FromLocationID == nil {
			return ErrMissingReceiptSource
		}
	}
	
	return nil
}

// Domain errors for stock movements
var (
	ErrInvalidInventoryItem      = NewDomainError("INVALID_INVENTORY_ITEM", "Invalid inventory item")
	ErrInvalidTransferLocations  = NewDomainError("INVALID_TRANSFER_LOCATIONS", "Transfer movements require both from and to locations")
	ErrSameTransferLocation      = NewDomainError("SAME_TRANSFER_LOCATION", "From and to locations cannot be the same")
	ErrMissingReceiptSource      = NewDomainError("MISSING_RECEIPT_SOURCE", "Receipt movements require a supplier or source location")
	ErrMovementNotPending        = NewDomainError("MOVEMENT_NOT_PENDING", "Movement is not in pending status")
	ErrMovementAlreadyProcessed  = NewDomainError("MOVEMENT_ALREADY_PROCESSED", "Movement has already been processed")
	ErrMovementNotApproved       = NewDomainError("MOVEMENT_NOT_APPROVED", "Movement requires approval before processing")
)
