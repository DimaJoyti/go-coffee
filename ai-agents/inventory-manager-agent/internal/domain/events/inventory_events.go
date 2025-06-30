package events

import (
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/entities"
)

// BaseEvent provides common event functionality
type BaseEvent struct {
	ID           uuid.UUID              `json:"id"`
	Type         string                 `json:"type"`
	AggregateID_ string                 `json:"aggregate_id"`
	OccurredAt_  time.Time              `json:"occurred_at"`
	Version      int                    `json:"version"`
	Data         map[string]interface{} `json:"data"`
}

// EventType returns the event type
func (e *BaseEvent) EventType() string {
	return e.Type
}

// AggregateID returns the aggregate ID
func (e *BaseEvent) AggregateID() string {
	return e.AggregateID_
}

// OccurredAt returns when the event occurred
func (e *BaseEvent) OccurredAt() time.Time {
	return e.OccurredAt_
}

// EventData returns the event data
func (e *BaseEvent) EventData() map[string]interface{} {
	return e.Data
}

// InventoryItemCreatedEvent is published when an inventory item is created
type InventoryItemCreatedEvent struct {
	BaseEvent
	Item *entities.InventoryItem `json:"item"`
}

// NewInventoryItemCreatedEvent creates a new inventory item created event
func NewInventoryItemCreatedEvent(item *entities.InventoryItem) *InventoryItemCreatedEvent {
	return &InventoryItemCreatedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "inventory.item.created",
			AggregateID_: item.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"item_id":     item.ID.String(),
				"sku":         item.SKU,
				"name":        item.Name,
				"category":    item.Category,
				"location_id": item.LocationID.String(),
				"created_by":  item.CreatedBy,
			},
		},
		Item: item,
	}
}

// InventoryItemUpdatedEvent is published when an inventory item is updated
type InventoryItemUpdatedEvent struct {
	BaseEvent
	Item        *entities.InventoryItem `json:"item"`
	PreviousItem *entities.InventoryItem `json:"previous_item"`
}

// NewInventoryItemUpdatedEvent creates a new inventory item updated event
func NewInventoryItemUpdatedEvent(item, previousItem *entities.InventoryItem) *InventoryItemUpdatedEvent {
	changes := make(map[string]interface{})
	
	if item.MinimumLevel != previousItem.MinimumLevel {
		changes["minimum_level"] = map[string]interface{}{
			"old": previousItem.MinimumLevel,
			"new": item.MinimumLevel,
		}
	}
	
	if item.MaximumLevel != previousItem.MaximumLevel {
		changes["maximum_level"] = map[string]interface{}{
			"old": previousItem.MaximumLevel,
			"new": item.MaximumLevel,
		}
	}
	
	if item.ReorderPoint != previousItem.ReorderPoint {
		changes["reorder_point"] = map[string]interface{}{
			"old": previousItem.ReorderPoint,
			"new": item.ReorderPoint,
		}
	}
	
	if item.Status != previousItem.Status {
		changes["status"] = map[string]interface{}{
			"old": previousItem.Status,
			"new": item.Status,
		}
	}

	return &InventoryItemUpdatedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "inventory.item.updated",
			AggregateID_: item.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"item_id":    item.ID.String(),
				"sku":        item.SKU,
				"name":       item.Name,
				"changes":    changes,
				"updated_by": item.UpdatedBy,
			},
		},
		Item:         item,
		PreviousItem: previousItem,
	}
}

// StockMovementCompletedEvent is published when a stock movement is completed
type StockMovementCompletedEvent struct {
	BaseEvent
	Movement *entities.StockMovement `json:"movement"`
}

// NewStockMovementCompletedEvent creates a new stock movement completed event
func NewStockMovementCompletedEvent(movement *entities.StockMovement) *StockMovementCompletedEvent {
	return &StockMovementCompletedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "inventory.movement.completed",
			AggregateID_: movement.InventoryItemID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"movement_id":       movement.ID.String(),
				"movement_number":   movement.MovementNumber,
				"movement_type":     movement.Type,
				"direction":         movement.Direction,
				"inventory_item_id": movement.InventoryItemID.String(),
				"quantity":          movement.Quantity,
				"unit":              movement.Unit,
				"total_cost":        movement.TotalCost,
				"from_location_id":  movement.FromLocationID,
				"to_location_id":    movement.ToLocationID,
				"reason":            movement.Reason,
				"processed_by":      movement.ProcessedBy,
				"completed_at":      movement.CompletedAt,
			},
		},
		Movement: movement,
	}
}

// LowStockAlertEvent is published when an item reaches low stock level
type LowStockAlertEvent struct {
	BaseEvent
	Item *entities.InventoryItem `json:"item"`
}

// NewLowStockAlertEvent creates a new low stock alert event
func NewLowStockAlertEvent(item *entities.InventoryItem) *LowStockAlertEvent {
	return &LowStockAlertEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "inventory.alert.low_stock",
			AggregateID_: item.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"item_id":        item.ID.String(),
				"sku":            item.SKU,
				"name":           item.Name,
				"current_stock":  item.CurrentStock,
				"minimum_level":  item.MinimumLevel,
				"location_id":    item.LocationID.String(),
				"category":       item.Category,
				"alert_level":    "low_stock",
			},
		},
		Item: item,
	}
}

// OutOfStockAlertEvent is published when an item is out of stock
type OutOfStockAlertEvent struct {
	BaseEvent
	Item *entities.InventoryItem `json:"item"`
}

// NewOutOfStockAlertEvent creates a new out of stock alert event
func NewOutOfStockAlertEvent(item *entities.InventoryItem) *OutOfStockAlertEvent {
	return &OutOfStockAlertEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "inventory.alert.out_of_stock",
			AggregateID_: item.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"item_id":       item.ID.String(),
				"sku":           item.SKU,
				"name":          item.Name,
				"current_stock": item.CurrentStock,
				"location_id":   item.LocationID.String(),
				"category":      item.Category,
				"alert_level":   "out_of_stock",
			},
		},
		Item: item,
	}
}

// ReorderNeededEvent is published when an item needs to be reordered
type ReorderNeededEvent struct {
	BaseEvent
	Item *entities.InventoryItem `json:"item"`
}

// NewReorderNeededEvent creates a new reorder needed event
func NewReorderNeededEvent(item *entities.InventoryItem) *ReorderNeededEvent {
	return &ReorderNeededEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "inventory.alert.reorder_needed",
			AggregateID_: item.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"item_id":          item.ID.String(),
				"sku":              item.SKU,
				"name":             item.Name,
				"current_stock":    item.CurrentStock,
				"reorder_point":    item.ReorderPoint,
				"reorder_quantity": item.ReorderQuantity,
				"location_id":      item.LocationID.String(),
				"category":         item.Category,
				"supplier_id":      item.SupplierID.String(),
				"alert_level":      "reorder_needed",
			},
		},
		Item: item,
	}
}

// StockReservedEvent is published when stock is reserved
type StockReservedEvent struct {
	BaseEvent
	Item     *entities.InventoryItem `json:"item"`
	Quantity float64                 `json:"quantity"`
	Reason   string                  `json:"reason"`
}

// NewStockReservedEvent creates a new stock reserved event
func NewStockReservedEvent(item *entities.InventoryItem, quantity float64, reason string) *StockReservedEvent {
	return &StockReservedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "inventory.stock.reserved",
			AggregateID_: item.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"item_id":          item.ID.String(),
				"sku":              item.SKU,
				"name":             item.Name,
				"reserved_quantity": quantity,
				"total_reserved":   item.ReservedStock,
				"available_stock":  item.AvailableStock,
				"reason":           reason,
				"location_id":      item.LocationID.String(),
			},
		},
		Item:     item,
		Quantity: quantity,
		Reason:   reason,
	}
}

// StockReleasedEvent is published when reserved stock is released
type StockReleasedEvent struct {
	BaseEvent
	Item     *entities.InventoryItem `json:"item"`
	Quantity float64                 `json:"quantity"`
	Reason   string                  `json:"reason"`
}

// NewStockReleasedEvent creates a new stock released event
func NewStockReleasedEvent(item *entities.InventoryItem, quantity float64, reason string) *StockReleasedEvent {
	return &StockReleasedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "inventory.stock.released",
			AggregateID_: item.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"item_id":          item.ID.String(),
				"sku":              item.SKU,
				"name":             item.Name,
				"released_quantity": quantity,
				"total_reserved":   item.ReservedStock,
				"available_stock":  item.AvailableStock,
				"reason":           reason,
				"location_id":      item.LocationID.String(),
			},
		},
		Item:     item,
		Quantity: quantity,
		Reason:   reason,
	}
}

// ExpirationAlertEvent is published when items are approaching expiration
type ExpirationAlertEvent struct {
	BaseEvent
	Item  *entities.InventoryItem `json:"item"`
	Batch *entities.InventoryBatch `json:"batch"`
	DaysUntilExpiration int        `json:"days_until_expiration"`
}

// NewExpirationAlertEvent creates a new expiration alert event
func NewExpirationAlertEvent(item *entities.InventoryItem, batch *entities.InventoryBatch, daysUntilExpiration int) *ExpirationAlertEvent {
	return &ExpirationAlertEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "inventory.alert.expiration",
			AggregateID_: item.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"item_id":               item.ID.String(),
				"sku":                   item.SKU,
				"name":                  item.Name,
				"batch_id":              batch.ID.String(),
				"batch_number":          batch.BatchNumber,
				"batch_quantity":        batch.Quantity,
				"expiration_date":       batch.ExpirationDate,
				"days_until_expiration": daysUntilExpiration,
				"location_id":           item.LocationID.String(),
				"alert_level":           "expiration_warning",
			},
		},
		Item:                item,
		Batch:               batch,
		DaysUntilExpiration: daysUntilExpiration,
	}
}

// PurchaseOrderCreatedEvent is published when a purchase order is created
type PurchaseOrderCreatedEvent struct {
	BaseEvent
	Order *entities.PurchaseOrder `json:"order"`
}

// NewPurchaseOrderCreatedEvent creates a new purchase order created event
func NewPurchaseOrderCreatedEvent(order *entities.PurchaseOrder) *PurchaseOrderCreatedEvent {
	return &PurchaseOrderCreatedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "procurement.order.created",
			AggregateID_: order.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"order_id":      order.ID.String(),
				"order_number":  order.OrderNumber,
				"supplier_id":   order.SupplierID.String(),
				"location_id":   order.LocationID.String(),
				"total_amount":  order.TotalAmount,
				"item_count":    len(order.Items),
				"requested_by":  order.RequestedBy,
				"priority":      order.Priority,
				"type":          order.Type,
			},
		},
		Order: order,
	}
}

// PurchaseOrderApprovedEvent is published when a purchase order is approved
type PurchaseOrderApprovedEvent struct {
	BaseEvent
	Order *entities.PurchaseOrder `json:"order"`
}

// NewPurchaseOrderApprovedEvent creates a new purchase order approved event
func NewPurchaseOrderApprovedEvent(order *entities.PurchaseOrder) *PurchaseOrderApprovedEvent {
	return &PurchaseOrderApprovedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "procurement.order.approved",
			AggregateID_: order.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"order_id":     order.ID.String(),
				"order_number": order.OrderNumber,
				"supplier_id":  order.SupplierID.String(),
				"total_amount": order.TotalAmount,
				"approved_by":  order.ApprovedBy,
			},
		},
		Order: order,
	}
}

// PurchaseOrderReceivedEvent is published when a purchase order is received
type PurchaseOrderReceivedEvent struct {
	BaseEvent
	Order   *entities.PurchaseOrder `json:"order"`
	Receipt *entities.OrderReceipt  `json:"receipt"`
}

// NewPurchaseOrderReceivedEvent creates a new purchase order received event
func NewPurchaseOrderReceivedEvent(order *entities.PurchaseOrder, receipt *entities.OrderReceipt) *PurchaseOrderReceivedEvent {
	return &PurchaseOrderReceivedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "procurement.order.received",
			AggregateID_: order.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"order_id":       order.ID.String(),
				"order_number":   order.OrderNumber,
				"receipt_id":     receipt.ID.String(),
				"receipt_number": receipt.ReceiptNumber,
				"received_by":    receipt.ReceivedBy,
				"received_date":  receipt.ReceivedDate,
				"item_count":     len(receipt.Items),
			},
		},
		Order:   order,
		Receipt: receipt,
	}
}

// SupplierPerformanceUpdatedEvent is published when supplier performance is updated
type SupplierPerformanceUpdatedEvent struct {
	BaseEvent
	Supplier    *entities.Supplier            `json:"supplier"`
	Performance *entities.SupplierPerformance `json:"performance"`
}

// NewSupplierPerformanceUpdatedEvent creates a new supplier performance updated event
func NewSupplierPerformanceUpdatedEvent(supplier *entities.Supplier, performance *entities.SupplierPerformance) *SupplierPerformanceUpdatedEvent {
	return &SupplierPerformanceUpdatedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New(),
			Type:        "supplier.performance.updated",
			AggregateID_: supplier.ID.String(),
			OccurredAt_:  time.Now(),
			Version:     1,
			Data: map[string]interface{}{
				"supplier_id":           supplier.ID.String(),
				"supplier_name":         supplier.Name,
				"overall_rating":        performance.OverallRating,
				"quality_rating":        performance.QualityRating,
				"delivery_rating":       performance.DeliveryRating,
				"on_time_delivery_rate": performance.OnTimeDeliveryRate,
				"quality_reject_rate":   performance.QualityRejectRate,
				"total_orders":          performance.TotalOrders,
				"total_value":           performance.TotalValue,
			},
		},
		Supplier:    supplier,
		Performance: performance,
	}
}
