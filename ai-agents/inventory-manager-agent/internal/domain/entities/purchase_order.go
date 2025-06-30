package entities

import (
	"time"

	"github.com/google/uuid"
)

// PurchaseOrder represents a purchase order for inventory items
type PurchaseOrder struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	OrderNumber       string                 `json:"order_number" redis:"order_number"`
	Status            PurchaseOrderStatus    `json:"status" redis:"status"`
	Type              PurchaseOrderType      `json:"type" redis:"type"`
	Priority          OrderPriority          `json:"priority" redis:"priority"`
	SupplierID        uuid.UUID              `json:"supplier_id" redis:"supplier_id"`
	Supplier          *Supplier              `json:"supplier,omitempty"`
	LocationID        uuid.UUID              `json:"location_id" redis:"location_id"`
	Location          *Location              `json:"location,omitempty"`
	RequestedBy       string                 `json:"requested_by" redis:"requested_by"`
	ApprovedBy        string                 `json:"approved_by" redis:"approved_by"`
	BuyerID           uuid.UUID              `json:"buyer_id" redis:"buyer_id"`
	BuyerName         string                 `json:"buyer_name" redis:"buyer_name"`
	Items             []*PurchaseOrderItem   `json:"items,omitempty"`
	SubTotal          Money                  `json:"sub_total" redis:"sub_total"`
	TaxAmount         Money                  `json:"tax_amount" redis:"tax_amount"`
	ShippingCost      Money                  `json:"shipping_cost" redis:"shipping_cost"`
	DiscountAmount    Money                  `json:"discount_amount" redis:"discount_amount"`
	TotalAmount       Money                  `json:"total_amount" redis:"total_amount"`
	Currency          string                 `json:"currency" redis:"currency"`
	PaymentTerms      *PaymentTerms          `json:"payment_terms,omitempty"`
	DeliveryTerms     *DeliveryTerms         `json:"delivery_terms,omitempty"`
	ShippingAddress   *Address               `json:"shipping_address,omitempty"`
	BillingAddress    *Address               `json:"billing_address,omitempty"`
	RequestedDate     time.Time              `json:"requested_date" redis:"requested_date"`
	ExpectedDate      *time.Time             `json:"expected_date,omitempty" redis:"expected_date"`
	PromisedDate      *time.Time             `json:"promised_date,omitempty" redis:"promised_date"`
	OrderDate         *time.Time             `json:"order_date,omitempty" redis:"order_date"`
	ConfirmedDate     *time.Time             `json:"confirmed_date,omitempty" redis:"confirmed_date"`
	ShippedDate       *time.Time             `json:"shipped_date,omitempty" redis:"shipped_date"`
	DeliveredDate     *time.Time             `json:"delivered_date,omitempty" redis:"delivered_date"`
	CancelledDate     *time.Time             `json:"cancelled_date,omitempty" redis:"cancelled_date"`
	CancellationReason string                `json:"cancellation_reason" redis:"cancellation_reason"`
	SpecialInstructions string               `json:"special_instructions" redis:"special_instructions"`
	InternalNotes     string                 `json:"internal_notes" redis:"internal_notes"`
	SupplierNotes     string                 `json:"supplier_notes" redis:"supplier_notes"`
	Documents         []*OrderDocument       `json:"documents,omitempty"`
	Approvals         []*OrderApproval       `json:"approvals,omitempty"`
	Receipts          []*OrderReceipt        `json:"receipts,omitempty"`
	Invoices          []*OrderInvoice        `json:"invoices,omitempty"`
	Attributes        map[string]interface{} `json:"attributes" redis:"attributes"`
	Tags              []string               `json:"tags" redis:"tags"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         string                 `json:"created_by" redis:"created_by"`
	UpdatedBy         string                 `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// PurchaseOrderStatus defines the status of a purchase order
type PurchaseOrderStatus string

const (
	POStatusDraft      PurchaseOrderStatus = "draft"
	POStatusPending    PurchaseOrderStatus = "pending"
	POStatusApproved   PurchaseOrderStatus = "approved"
	POStatusRejected   PurchaseOrderStatus = "rejected"
	POStatusSent       PurchaseOrderStatus = "sent"
	POStatusConfirmed  PurchaseOrderStatus = "confirmed"
	POStatusPartial    PurchaseOrderStatus = "partial"
	POStatusReceived   PurchaseOrderStatus = "received"
	POStatusClosed     PurchaseOrderStatus = "closed"
	POStatusCancelled  PurchaseOrderStatus = "cancelled"
	POStatusOnHold     PurchaseOrderStatus = "on_hold"
)

// PurchaseOrderType defines the type of purchase order
type PurchaseOrderType string

const (
	POTypeStandard   PurchaseOrderType = "standard"
	POTypeEmergency  PurchaseOrderType = "emergency"
	POTypeBlanket    PurchaseOrderType = "blanket"
	POTypeContract   PurchaseOrderType = "contract"
	POTypeService    PurchaseOrderType = "service"
	POTypeConsignment PurchaseOrderType = "consignment"
)

// OrderPriority defines the priority of an order
type OrderPriority string

const (
	PriorityLow      OrderPriority = "low"
	PriorityNormal   OrderPriority = "normal"
	PriorityHigh     OrderPriority = "high"
	PriorityUrgent   OrderPriority = "urgent"
	PriorityCritical OrderPriority = "critical"
)

// PurchaseOrderItem represents an item in a purchase order
type PurchaseOrderItem struct {
	ID                uuid.UUID       `json:"id" redis:"id"`
	LineNumber        int             `json:"line_number" redis:"line_number"`
	InventoryItemID   uuid.UUID       `json:"inventory_item_id" redis:"inventory_item_id"`
	InventoryItem     *InventoryItem  `json:"inventory_item,omitempty"`
	SupplierSKU       string          `json:"supplier_sku" redis:"supplier_sku"`
	InternalSKU       string          `json:"internal_sku" redis:"internal_sku"`
	Description       string          `json:"description" redis:"description"`
	Quantity          float64         `json:"quantity" redis:"quantity"`
	Unit              MeasurementUnit `json:"unit" redis:"unit"`
	UnitPrice         Money           `json:"unit_price" redis:"unit_price"`
	LineTotal         Money           `json:"line_total" redis:"line_total"`
	TaxRate           float64         `json:"tax_rate" redis:"tax_rate"`
	TaxAmount         Money           `json:"tax_amount" redis:"tax_amount"`
	DiscountPercent   float64         `json:"discount_percent" redis:"discount_percent"`
	DiscountAmount    Money           `json:"discount_amount" redis:"discount_amount"`
	RequestedDate     time.Time       `json:"requested_date" redis:"requested_date"`
	PromisedDate      *time.Time      `json:"promised_date,omitempty" redis:"promised_date"`
	ReceivedQuantity  float64         `json:"received_quantity" redis:"received_quantity"`
	RemainingQuantity float64         `json:"remaining_quantity" redis:"remaining_quantity"`
	Status            string          `json:"status" redis:"status"`
	Notes             string          `json:"notes" redis:"notes"`
	CreatedAt         time.Time       `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" redis:"updated_at"`
}

// OrderDocument represents a document associated with a purchase order
type OrderDocument struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	Type        string    `json:"type" redis:"type"`
	Name        string    `json:"name" redis:"name"`
	Description string    `json:"description" redis:"description"`
	URL         string    `json:"url" redis:"url"`
	MimeType    string    `json:"mime_type" redis:"mime_type"`
	Size        int64     `json:"size" redis:"size"`
	Checksum    string    `json:"checksum" redis:"checksum"`
	UploadedBy  string    `json:"uploaded_by" redis:"uploaded_by"`
	UploadedAt  time.Time `json:"uploaded_at" redis:"uploaded_at"`
}

// OrderApproval represents an approval for a purchase order
type OrderApproval struct {
	ID           uuid.UUID  `json:"id" redis:"id"`
	ApproverID   uuid.UUID  `json:"approver_id" redis:"approver_id"`
	ApproverName string     `json:"approver_name" redis:"approver_name"`
	Level        int        `json:"level" redis:"level"`
	Status       string     `json:"status" redis:"status"`
	Comments     string     `json:"comments" redis:"comments"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty" redis:"approved_at"`
	RejectedAt   *time.Time `json:"rejected_at,omitempty" redis:"rejected_at"`
	CreatedAt    time.Time  `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" redis:"updated_at"`
}

// OrderReceipt represents a receipt for a purchase order
type OrderReceipt struct {
	ID            uuid.UUID              `json:"id" redis:"id"`
	ReceiptNumber string                 `json:"receipt_number" redis:"receipt_number"`
	ReceivedDate  time.Time              `json:"received_date" redis:"received_date"`
	ReceivedBy    string                 `json:"received_by" redis:"received_by"`
	Items         []*ReceiptItem         `json:"items,omitempty"`
	QualityCheck  *QualityCheck          `json:"quality_check,omitempty"`
	Notes         string                 `json:"notes" redis:"notes"`
	Attributes    map[string]interface{} `json:"attributes" redis:"attributes"`
	CreatedAt     time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" redis:"updated_at"`
}

// ReceiptItem represents an item in a receipt
type ReceiptItem struct {
	ID                uuid.UUID       `json:"id" redis:"id"`
	PurchaseOrderItemID uuid.UUID     `json:"purchase_order_item_id" redis:"purchase_order_item_id"`
	InventoryItemID   uuid.UUID       `json:"inventory_item_id" redis:"inventory_item_id"`
	ReceivedQuantity  float64         `json:"received_quantity" redis:"received_quantity"`
	Unit              MeasurementUnit `json:"unit" redis:"unit"`
	BatchNumber       string          `json:"batch_number" redis:"batch_number"`
	LotNumber         string          `json:"lot_number" redis:"lot_number"`
	ExpirationDate    *time.Time      `json:"expiration_date,omitempty" redis:"expiration_date"`
	ManufactureDate   *time.Time      `json:"manufacture_date,omitempty" redis:"manufacture_date"`
	SerialNumbers     []string        `json:"serial_numbers" redis:"serial_numbers"`
	QualityStatus     QualityStatus   `json:"quality_status" redis:"quality_status"`
	QualityNotes      string          `json:"quality_notes" redis:"quality_notes"`
	DamageQuantity    float64         `json:"damage_quantity" redis:"damage_quantity"`
	DamageNotes       string          `json:"damage_notes" redis:"damage_notes"`
	StorageLocation   string          `json:"storage_location" redis:"storage_location"`
	CreatedAt         time.Time       `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" redis:"updated_at"`
}

// OrderInvoice represents an invoice for a purchase order
type OrderInvoice struct {
	ID            uuid.UUID              `json:"id" redis:"id"`
	InvoiceNumber string                 `json:"invoice_number" redis:"invoice_number"`
	InvoiceDate   time.Time              `json:"invoice_date" redis:"invoice_date"`
	DueDate       time.Time              `json:"due_date" redis:"due_date"`
	Amount        Money                  `json:"amount" redis:"amount"`
	TaxAmount     Money                  `json:"tax_amount" redis:"tax_amount"`
	TotalAmount   Money                  `json:"total_amount" redis:"total_amount"`
	Status        string                 `json:"status" redis:"status"`
	PaidDate      *time.Time             `json:"paid_date,omitempty" redis:"paid_date"`
	PaidAmount    Money                  `json:"paid_amount" redis:"paid_amount"`
	Notes         string                 `json:"notes" redis:"notes"`
	Attributes    map[string]interface{} `json:"attributes" redis:"attributes"`
	CreatedAt     time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" redis:"updated_at"`
}

// NewPurchaseOrder creates a new purchase order
func NewPurchaseOrder(
	supplierID, locationID, buyerID uuid.UUID,
	requestedBy, buyerName string,
	orderType PurchaseOrderType,
	priority OrderPriority,
) *PurchaseOrder {
	now := time.Now()
	return &PurchaseOrder{
		ID:              uuid.New(),
		OrderNumber:     generateOrderNumber(),
		Status:          POStatusDraft,
		Type:            orderType,
		Priority:        priority,
		SupplierID:      supplierID,
		LocationID:      locationID,
		RequestedBy:     requestedBy,
		BuyerID:         buyerID,
		BuyerName:       buyerName,
		Items:           []*PurchaseOrderItem{},
		SubTotal:        Money{Amount: 0, Currency: "USD"},
		TaxAmount:       Money{Amount: 0, Currency: "USD"},
		ShippingCost:    Money{Amount: 0, Currency: "USD"},
		DiscountAmount:  Money{Amount: 0, Currency: "USD"},
		TotalAmount:     Money{Amount: 0, Currency: "USD"},
		Currency:        "USD",
		RequestedDate:   now,
		Documents:       []*OrderDocument{},
		Approvals:       []*OrderApproval{},
		Receipts:        []*OrderReceipt{},
		Invoices:        []*OrderInvoice{},
		Attributes:      make(map[string]interface{}),
		Tags:            []string{},
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
	}
}

// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	return "PO-" + time.Now().Format("20060102-150405") + "-" + uuid.New().String()[:8]
}

// AddItem adds an item to the purchase order
func (po *PurchaseOrder) AddItem(item *PurchaseOrderItem) {
	item.LineNumber = len(po.Items) + 1
	po.Items = append(po.Items, item)
	po.recalculateTotals()
	po.UpdatedAt = time.Now()
	po.Version++
}

// RemoveItem removes an item from the purchase order
func (po *PurchaseOrder) RemoveItem(itemID uuid.UUID) {
	for i, item := range po.Items {
		if item.ID == itemID {
			po.Items = append(po.Items[:i], po.Items[i+1:]...)
			po.renumberItems()
			po.recalculateTotals()
			po.UpdatedAt = time.Now()
			po.Version++
			break
		}
	}
}

// UpdateItem updates an item in the purchase order
func (po *PurchaseOrder) UpdateItem(itemID uuid.UUID, quantity float64, unitPrice Money) {
	for _, item := range po.Items {
		if item.ID == itemID {
			item.Quantity = quantity
			item.UnitPrice = unitPrice
			item.calculateLineTotal()
			item.UpdatedAt = time.Now()
			break
		}
	}
	po.recalculateTotals()
	po.UpdatedAt = time.Now()
	po.Version++
}

// renumberItems renumbers the line items
func (po *PurchaseOrder) renumberItems() {
	for i, item := range po.Items {
		item.LineNumber = i + 1
	}
}

// recalculateTotals recalculates the order totals
func (po *PurchaseOrder) recalculateTotals() {
	subTotal := 0.0
	taxTotal := 0.0
	
	for _, item := range po.Items {
		subTotal += item.LineTotal.Amount
		taxTotal += item.TaxAmount.Amount
	}
	
	po.SubTotal.Amount = subTotal
	po.TaxAmount.Amount = taxTotal
	po.TotalAmount.Amount = subTotal + taxTotal + po.ShippingCost.Amount - po.DiscountAmount.Amount
}

// SetShippingCost sets the shipping cost
func (po *PurchaseOrder) SetShippingCost(cost Money) {
	po.ShippingCost = cost
	po.recalculateTotals()
	po.UpdatedAt = time.Now()
	po.Version++
}

// SetDiscount sets the discount amount
func (po *PurchaseOrder) SetDiscount(discount Money) {
	po.DiscountAmount = discount
	po.recalculateTotals()
	po.UpdatedAt = time.Now()
	po.Version++
}

// Submit submits the purchase order for approval
func (po *PurchaseOrder) Submit() error {
	if po.Status != POStatusDraft {
		return ErrOrderNotDraft
	}
	
	if len(po.Items) == 0 {
		return ErrOrderNoItems
	}
	
	po.Status = POStatusPending
	po.UpdatedAt = time.Now()
	po.Version++
	return nil
}

// Approve approves the purchase order
func (po *PurchaseOrder) Approve(approvedBy string) error {
	if po.Status != POStatusPending {
		return ErrOrderNotPending
	}
	
	po.Status = POStatusApproved
	po.ApprovedBy = approvedBy
	po.UpdatedAt = time.Now()
	po.Version++
	return nil
}

// Reject rejects the purchase order
func (po *PurchaseOrder) Reject(reason string) error {
	if po.Status != POStatusPending {
		return ErrOrderNotPending
	}
	
	po.Status = POStatusRejected
	po.CancellationReason = reason
	po.UpdatedAt = time.Now()
	po.Version++
	return nil
}

// Send sends the purchase order to the supplier
func (po *PurchaseOrder) Send() error {
	if po.Status != POStatusApproved {
		return ErrOrderNotApproved
	}
	
	now := time.Now()
	po.Status = POStatusSent
	po.OrderDate = &now
	po.UpdatedAt = now
	po.Version++
	return nil
}

// Confirm confirms the purchase order from the supplier
func (po *PurchaseOrder) Confirm(promisedDate *time.Time) error {
	if po.Status != POStatusSent {
		return ErrOrderNotSent
	}
	
	now := time.Now()
	po.Status = POStatusConfirmed
	po.ConfirmedDate = &now
	po.PromisedDate = promisedDate
	po.UpdatedAt = now
	po.Version++
	return nil
}

// Cancel cancels the purchase order
func (po *PurchaseOrder) Cancel(reason string) error {
	if po.Status == POStatusReceived || po.Status == POStatusClosed || po.Status == POStatusCancelled {
		return ErrOrderCannotCancel
	}
	
	now := time.Now()
	po.Status = POStatusCancelled
	po.CancelledDate = &now
	po.CancellationReason = reason
	po.UpdatedAt = now
	po.Version++
	return nil
}

// Close closes the purchase order
func (po *PurchaseOrder) Close() error {
	if po.Status != POStatusReceived && po.Status != POStatusPartial {
		return ErrOrderCannotClose
	}
	
	po.Status = POStatusClosed
	po.UpdatedAt = time.Now()
	po.Version++
	return nil
}

// AddReceipt adds a receipt to the purchase order
func (po *PurchaseOrder) AddReceipt(receipt *OrderReceipt) {
	po.Receipts = append(po.Receipts, receipt)
	po.updateReceiptStatus()
	po.UpdatedAt = time.Now()
	po.Version++
}

// updateReceiptStatus updates the order status based on receipts
func (po *PurchaseOrder) updateReceiptStatus() {
	if po.Status != POStatusConfirmed && po.Status != POStatusPartial {
		return
	}
	
	totalReceived := 0.0
	totalOrdered := 0.0
	
	for _, item := range po.Items {
		totalOrdered += item.Quantity
		totalReceived += item.ReceivedQuantity
		item.RemainingQuantity = item.Quantity - item.ReceivedQuantity
	}
	
	if totalReceived >= totalOrdered {
		po.Status = POStatusReceived
		now := time.Now()
		po.DeliveredDate = &now
	} else if totalReceived > 0 {
		po.Status = POStatusPartial
	}
}

// GetCompletionPercentage returns the completion percentage of the order
func (po *PurchaseOrder) GetCompletionPercentage() float64 {
	if len(po.Items) == 0 {
		return 0.0
	}
	
	totalOrdered := 0.0
	totalReceived := 0.0
	
	for _, item := range po.Items {
		totalOrdered += item.Quantity
		totalReceived += item.ReceivedQuantity
	}
	
	if totalOrdered == 0 {
		return 0.0
	}
	
	return (totalReceived / totalOrdered) * 100
}

// IsOverdue checks if the order is overdue
func (po *PurchaseOrder) IsOverdue() bool {
	if po.PromisedDate == nil {
		return false
	}
	
	return time.Now().After(*po.PromisedDate) && po.Status != POStatusReceived && po.Status != POStatusClosed
}

// CanReceive checks if the order can receive items
func (po *PurchaseOrder) CanReceive() bool {
	return po.Status == POStatusConfirmed || po.Status == POStatusPartial
}

// NewPurchaseOrderItem creates a new purchase order item
func NewPurchaseOrderItem(
	inventoryItemID uuid.UUID,
	supplierSKU, internalSKU, description string,
	quantity float64,
	unit MeasurementUnit,
	unitPrice Money,
) *PurchaseOrderItem {
	now := time.Now()
	item := &PurchaseOrderItem{
		ID:                uuid.New(),
		InventoryItemID:   inventoryItemID,
		SupplierSKU:       supplierSKU,
		InternalSKU:       internalSKU,
		Description:       description,
		Quantity:          quantity,
		Unit:              unit,
		UnitPrice:         unitPrice,
		TaxRate:           0.0,
		DiscountPercent:   0.0,
		RequestedDate:     now,
		ReceivedQuantity:  0.0,
		RemainingQuantity: quantity,
		Status:            "pending",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	
	item.calculateLineTotal()
	return item
}

// calculateLineTotal calculates the line total for the item
func (poi *PurchaseOrderItem) calculateLineTotal() {
	lineTotal := poi.Quantity * poi.UnitPrice.Amount
	
	// Apply discount
	if poi.DiscountPercent > 0 {
		discountAmount := lineTotal * (poi.DiscountPercent / 100)
		poi.DiscountAmount = Money{Amount: discountAmount, Currency: poi.UnitPrice.Currency}
		lineTotal -= discountAmount
	}
	
	poi.LineTotal = Money{Amount: lineTotal, Currency: poi.UnitPrice.Currency}
	
	// Calculate tax
	if poi.TaxRate > 0 {
		taxAmount := lineTotal * (poi.TaxRate / 100)
		poi.TaxAmount = Money{Amount: taxAmount, Currency: poi.UnitPrice.Currency}
	}
}

// Domain errors for purchase orders
var (
	ErrOrderNotDraft     = NewDomainError("ORDER_NOT_DRAFT", "Order is not in draft status")
	ErrOrderNoItems      = NewDomainError("ORDER_NO_ITEMS", "Order must have at least one item")
	ErrOrderNotPending   = NewDomainError("ORDER_NOT_PENDING", "Order is not in pending status")
	ErrOrderNotApproved  = NewDomainError("ORDER_NOT_APPROVED", "Order is not approved")
	ErrOrderNotSent      = NewDomainError("ORDER_NOT_SENT", "Order has not been sent")
	ErrOrderCannotCancel = NewDomainError("ORDER_CANNOT_CANCEL", "Order cannot be cancelled in current status")
	ErrOrderCannotClose  = NewDomainError("ORDER_CANNOT_CLOSE", "Order cannot be closed in current status")
)
