package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/entities"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/events"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/repositories"
)

// InventoryTrackingService provides core inventory tracking functionality
type InventoryTrackingService struct {
	inventoryRepo     repositories.InventoryRepository
	movementRepo      repositories.StockMovementRepository
	locationRepo      repositories.LocationRepository
	eventPublisher    EventPublisher
	logger            Logger
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	PublishEvent(ctx context.Context, event DomainEvent) error
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// DomainEvent represents a domain event
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	EventData() map[string]interface{}
}

// NewInventoryTrackingService creates a new inventory tracking service
func NewInventoryTrackingService(
	inventoryRepo repositories.InventoryRepository,
	movementRepo repositories.StockMovementRepository,
	locationRepo repositories.LocationRepository,
	eventPublisher EventPublisher,
	logger Logger,
) *InventoryTrackingService {
	return &InventoryTrackingService{
		inventoryRepo:  inventoryRepo,
		movementRepo:   movementRepo,
		locationRepo:   locationRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// CreateInventoryItem creates a new inventory item
func (its *InventoryTrackingService) CreateInventoryItem(ctx context.Context, item *entities.InventoryItem) error {
	its.logger.Info("Creating inventory item", "sku", item.SKU, "name", item.Name)

	// Validate the item
	if err := its.validateInventoryItem(ctx, item); err != nil {
		its.logger.Error("Inventory item validation failed", err, "sku", item.SKU)
		return err
	}

	// Check for duplicate SKU
	existing, err := its.inventoryRepo.GetBySKU(ctx, item.SKU)
	if err == nil && existing != nil {
		return entities.NewDomainError("DUPLICATE_SKU", fmt.Sprintf("Item with SKU %s already exists", item.SKU))
	}

	// Create the item
	if err := its.inventoryRepo.Create(ctx, item); err != nil {
		its.logger.Error("Failed to create inventory item", err, "sku", item.SKU)
		return err
	}

	// Publish event
	event := events.NewInventoryItemCreatedEvent(item)
	if err := its.eventPublisher.PublishEvent(ctx, event); err != nil {
		its.logger.Error("Failed to publish inventory item created event", err, "item_id", item.ID)
	}

	its.logger.Info("Inventory item created successfully", "item_id", item.ID, "sku", item.SKU)
	return nil
}

// UpdateInventoryItem updates an existing inventory item
func (its *InventoryTrackingService) UpdateInventoryItem(ctx context.Context, item *entities.InventoryItem) error {
	its.logger.Info("Updating inventory item", "item_id", item.ID, "sku", item.SKU)

	// Validate the item
	if err := its.validateInventoryItem(ctx, item); err != nil {
		its.logger.Error("Inventory item validation failed", err, "item_id", item.ID)
		return err
	}

	// Get existing item for comparison
	existing, err := its.inventoryRepo.GetByID(ctx, item.ID)
	if err != nil {
		its.logger.Error("Failed to get existing inventory item", err, "item_id", item.ID)
		return err
	}

	// Update the item
	if err := its.inventoryRepo.Update(ctx, item); err != nil {
		its.logger.Error("Failed to update inventory item", err, "item_id", item.ID)
		return err
	}

	// Publish event if significant changes occurred
	if its.hasSignificantChanges(existing, item) {
		event := events.NewInventoryItemUpdatedEvent(item, existing)
		if err := its.eventPublisher.PublishEvent(ctx, event); err != nil {
			its.logger.Error("Failed to publish inventory item updated event", err, "item_id", item.ID)
		}
	}

	its.logger.Info("Inventory item updated successfully", "item_id", item.ID)
	return nil
}

// ProcessStockMovement processes a stock movement and updates inventory levels
func (its *InventoryTrackingService) ProcessStockMovement(ctx context.Context, movement *entities.StockMovement) error {
	its.logger.Info("Processing stock movement", 
		"movement_id", movement.ID,
		"type", movement.Type,
		"item_id", movement.InventoryItemID,
		"quantity", movement.Quantity)

	// Validate the movement
	if err := movement.Validate(); err != nil {
		its.logger.Error("Stock movement validation failed", err, "movement_id", movement.ID)
		return err
	}

	// Check if movement can be processed
	if !movement.CanProcess() {
		return entities.NewDomainError("MOVEMENT_CANNOT_PROCESS", "Movement cannot be processed in current state")
	}

	// Get the inventory item
	item, err := its.inventoryRepo.GetByID(ctx, movement.InventoryItemID)
	if err != nil {
		its.logger.Error("Failed to get inventory item", err, "item_id", movement.InventoryItemID)
		return err
	}

	// Process within transaction
	return its.movementRepo.WithTransaction(ctx, func(ctx context.Context, movementRepo repositories.StockMovementRepository) error {
		return its.inventoryRepo.WithTransaction(ctx, func(ctx context.Context, inventoryRepo repositories.InventoryRepository) error {
			// Mark movement as processing
			movement.Process("system")
			if err := movementRepo.Update(ctx, movement); err != nil {
				return err
			}

			// Apply the movement to inventory
			if err := its.applyMovementToInventory(ctx, item, movement, inventoryRepo); err != nil {
				movement.Fail(err.Error())
				movementRepo.Update(ctx, movement)
				return err
			}

			// Complete the movement
			movement.Complete()
			if err := movementRepo.Update(ctx, movement); err != nil {
				return err
			}

			// Update inventory item
			if err := inventoryRepo.Update(ctx, item); err != nil {
				return err
			}

			// Publish events
			its.publishMovementEvents(ctx, movement, item)

			return nil
		})
	})
}

// ReceiveStock receives stock from a supplier or transfer
func (its *InventoryTrackingService) ReceiveStock(ctx context.Context, request *ReceiveStockRequest) error {
	its.logger.Info("Receiving stock", 
		"item_id", request.ItemID,
		"quantity", request.Quantity,
		"supplier_id", request.SupplierID)

	// Get inventory item
	item, err := its.inventoryRepo.GetByID(ctx, request.ItemID)
	if err != nil {
		its.logger.Error("Failed to get inventory item", err, "item_id", request.ItemID)
		return err
	}

	// Create stock movement
	movement := entities.NewStockMovement(
		entities.MovementTypeReceipt,
		entities.DirectionIn,
		request.ItemID,
		request.Quantity,
		item.Unit,
		request.Reason,
	)

	// Set additional details
	movement.SetCost(request.UnitCost)
	if request.SupplierID != nil {
		movement.SetSupplier(*request.SupplierID)
	}
	if request.LocationID != nil {
		movement.SetLocations(nil, request.LocationID)
	}
	if request.ExpirationDate != nil {
		movement.SetExpirationDate(*request.ExpirationDate)
	}
	if request.BatchNumber != "" {
		// Create new batch
		batch := &entities.InventoryBatch{
			ID:              uuid.New(),
			BatchNumber:     request.BatchNumber,
			Quantity:        request.Quantity,
			UnitCost:        request.UnitCost,
			ExpirationDate:  request.ExpirationDate,
			ManufactureDate: request.ManufactureDate,
			ReceivedDate:    time.Now(),
			QualityStatus:   entities.QualityPending,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if request.SupplierID != nil {
			batch.SupplierID = *request.SupplierID
		}
		
		// Add batch to item and set movement batch
		item.AddBatch(batch)
		movement.SetBatch(batch.ID)
	}

	// Create movement record
	if err := its.movementRepo.Create(ctx, movement); err != nil {
		its.logger.Error("Failed to create stock movement", err, "movement_id", movement.ID)
		return err
	}

	// Process the movement
	return its.ProcessStockMovement(ctx, movement)
}

// IssueStock issues stock for consumption or transfer
func (its *InventoryTrackingService) IssueStock(ctx context.Context, request *IssueStockRequest) error {
	its.logger.Info("Issuing stock", 
		"item_id", request.ItemID,
		"quantity", request.Quantity,
		"reason", request.Reason)

	// Get inventory item
	item, err := its.inventoryRepo.GetByID(ctx, request.ItemID)
	if err != nil {
		its.logger.Error("Failed to get inventory item", err, "item_id", request.ItemID)
		return err
	}

	// Check available stock
	if item.AvailableStock < request.Quantity {
		return entities.ErrInsufficientStock
	}

	// Create stock movement
	movement := entities.NewStockMovement(
		entities.MovementTypeIssue,
		entities.DirectionOut,
		request.ItemID,
		request.Quantity,
		item.Unit,
		request.Reason,
	)

	// Set additional details
	movement.SetCost(item.UnitCost)
	if request.LocationID != nil {
		movement.SetLocations(request.LocationID, nil)
	}
	if request.ReferenceType != "" && request.ReferenceID != nil {
		movement.SetReference(request.ReferenceType, request.ReferenceID, request.ReferenceNumber)
	}

	// Create movement record
	if err := its.movementRepo.Create(ctx, movement); err != nil {
		its.logger.Error("Failed to create stock movement", err, "movement_id", movement.ID)
		return err
	}

	// Process the movement
	return its.ProcessStockMovement(ctx, movement)
}

// TransferStock transfers stock between locations
func (its *InventoryTrackingService) TransferStock(ctx context.Context, request *TransferStockRequest) error {
	its.logger.Info("Transferring stock", 
		"item_id", request.ItemID,
		"quantity", request.Quantity,
		"from_location", request.FromLocationID,
		"to_location", request.ToLocationID)

	// Validate locations are different
	if request.FromLocationID == request.ToLocationID {
		return entities.NewDomainError("SAME_LOCATION", "From and to locations cannot be the same")
	}

	// Get inventory item
	item, err := its.inventoryRepo.GetByID(ctx, request.ItemID)
	if err != nil {
		its.logger.Error("Failed to get inventory item", err, "item_id", request.ItemID)
		return err
	}

	// Check available stock at source location
	if item.AvailableStock < request.Quantity {
		return entities.ErrInsufficientStock
	}

	// Create stock movement
	movement := entities.NewStockMovement(
		entities.MovementTypeTransfer,
		entities.DirectionMove,
		request.ItemID,
		request.Quantity,
		item.Unit,
		request.Reason,
	)

	// Set locations
	movement.SetLocations(&request.FromLocationID, &request.ToLocationID)
	movement.SetZones(request.FromZone, request.ToZone)
	movement.SetCost(item.UnitCost)

	// Create movement record
	if err := its.movementRepo.Create(ctx, movement); err != nil {
		its.logger.Error("Failed to create stock movement", err, "movement_id", movement.ID)
		return err
	}

	// Process the movement
	return its.ProcessStockMovement(ctx, movement)
}

// AdjustStock adjusts stock levels for corrections
func (its *InventoryTrackingService) AdjustStock(ctx context.Context, request *AdjustStockRequest) error {
	its.logger.Info("Adjusting stock", 
		"item_id", request.ItemID,
		"adjustment", request.Adjustment,
		"reason", request.Reason)

	// Get inventory item
	item, err := its.inventoryRepo.GetByID(ctx, request.ItemID)
	if err != nil {
		its.logger.Error("Failed to get inventory item", err, "item_id", request.ItemID)
		return err
	}

	// Create stock movement
	movement := entities.NewStockMovement(
		entities.MovementTypeAdjustment,
		entities.DirectionAdjust,
		request.ItemID,
		request.Adjustment,
		item.Unit,
		request.Reason,
	)

	movement.SetCost(item.UnitCost)
	if request.LocationID != nil {
		movement.SetLocations(request.LocationID, request.LocationID)
	}

	// Create movement record
	if err := its.movementRepo.Create(ctx, movement); err != nil {
		its.logger.Error("Failed to create stock movement", err, "movement_id", movement.ID)
		return err
	}

	// Process the movement
	return its.ProcessStockMovement(ctx, movement)
}

// ReserveStock reserves stock for future use
func (its *InventoryTrackingService) ReserveStock(ctx context.Context, itemID uuid.UUID, quantity float64, reason string) error {
	its.logger.Info("Reserving stock", "item_id", itemID, "quantity", quantity)

	// Get inventory item
	item, err := its.inventoryRepo.GetByID(ctx, itemID)
	if err != nil {
		its.logger.Error("Failed to get inventory item", err, "item_id", itemID)
		return err
	}

	// Reserve stock
	if err := item.ReserveStock(quantity); err != nil {
		its.logger.Error("Failed to reserve stock", err, "item_id", itemID, "quantity", quantity)
		return err
	}

	// Update inventory
	if err := its.inventoryRepo.Update(ctx, item); err != nil {
		its.logger.Error("Failed to update inventory item", err, "item_id", itemID)
		return err
	}

	// Create movement record
	movement := entities.NewStockMovement(
		entities.MovementTypeReservation,
		entities.DirectionOut,
		itemID,
		quantity,
		item.Unit,
		reason,
	)
	movement.SetCost(item.UnitCost)

	if err := its.movementRepo.Create(ctx, movement); err != nil {
		its.logger.Error("Failed to create reservation movement", err, "movement_id", movement.ID)
		return err
	}

	// Publish event
	event := events.NewStockReservedEvent(item, quantity, reason)
	if err := its.eventPublisher.PublishEvent(ctx, event); err != nil {
		its.logger.Error("Failed to publish stock reserved event", err, "item_id", itemID)
	}

	its.logger.Info("Stock reserved successfully", "item_id", itemID, "quantity", quantity)
	return nil
}

// ReleaseStock releases reserved stock
func (its *InventoryTrackingService) ReleaseStock(ctx context.Context, itemID uuid.UUID, quantity float64, reason string) error {
	its.logger.Info("Releasing stock", "item_id", itemID, "quantity", quantity)

	// Get inventory item
	item, err := its.inventoryRepo.GetByID(ctx, itemID)
	if err != nil {
		its.logger.Error("Failed to get inventory item", err, "item_id", itemID)
		return err
	}

	// Release stock
	if err := item.ReleaseStock(quantity); err != nil {
		its.logger.Error("Failed to release stock", err, "item_id", itemID, "quantity", quantity)
		return err
	}

	// Update inventory
	if err := its.inventoryRepo.Update(ctx, item); err != nil {
		its.logger.Error("Failed to update inventory item", err, "item_id", itemID)
		return err
	}

	// Create movement record
	movement := entities.NewStockMovement(
		entities.MovementTypeRelease,
		entities.DirectionIn,
		itemID,
		quantity,
		item.Unit,
		reason,
	)
	movement.SetCost(item.UnitCost)

	if err := its.movementRepo.Create(ctx, movement); err != nil {
		its.logger.Error("Failed to create release movement", err, "movement_id", movement.ID)
		return err
	}

	// Publish event
	event := events.NewStockReleasedEvent(item, quantity, reason)
	if err := its.eventPublisher.PublishEvent(ctx, event); err != nil {
		its.logger.Error("Failed to publish stock released event", err, "item_id", itemID)
	}

	its.logger.Info("Stock released successfully", "item_id", itemID, "quantity", quantity)
	return nil
}

// validateInventoryItem validates an inventory item
func (its *InventoryTrackingService) validateInventoryItem(ctx context.Context, item *entities.InventoryItem) error {
	if item.SKU == "" {
		return entities.NewDomainError("INVALID_SKU", "SKU is required")
	}

	if item.Name == "" {
		return entities.NewDomainError("INVALID_NAME", "Name is required")
	}

	if item.LocationID != uuid.Nil {
		// Validate location exists
		_, err := its.locationRepo.GetByID(ctx, item.LocationID)
		if err != nil {
			return entities.NewDomainError("INVALID_LOCATION", "Location does not exist")
		}
	}

	return nil
}

// applyMovementToInventory applies a stock movement to inventory levels
func (its *InventoryTrackingService) applyMovementToInventory(
	ctx context.Context,
	item *entities.InventoryItem,
	movement *entities.StockMovement,
	inventoryRepo repositories.InventoryRepository,
) error {
	effectiveQuantity := movement.GetEffectiveQuantity()
	newStock := item.CurrentStock + effectiveQuantity

	// Validate the new stock level
	if newStock < 0 {
		return entities.NewDomainError("NEGATIVE_STOCK", "Stock cannot be negative")
	}

	// Update stock level
	item.UpdateStock(newStock)

	its.logger.Info("Applied stock movement", 
		"item_id", item.ID,
		"movement_type", movement.Type,
		"quantity", movement.Quantity,
		"effective_quantity", effectiveQuantity,
		"old_stock", item.CurrentStock - effectiveQuantity,
		"new_stock", item.CurrentStock)

	return nil
}

// hasSignificantChanges checks if there are significant changes between items
func (its *InventoryTrackingService) hasSignificantChanges(old, new *entities.InventoryItem) bool {
	return old.MinimumLevel != new.MinimumLevel ||
		   old.MaximumLevel != new.MaximumLevel ||
		   old.ReorderPoint != new.ReorderPoint ||
		   old.ReorderQuantity != new.ReorderQuantity ||
		   old.IsActive != new.IsActive ||
		   old.Status != new.Status
}

// publishMovementEvents publishes events related to stock movements
func (its *InventoryTrackingService) publishMovementEvents(ctx context.Context, movement *entities.StockMovement, item *entities.InventoryItem) {
	// Publish movement completed event
	event := events.NewStockMovementCompletedEvent(movement)
	if err := its.eventPublisher.PublishEvent(ctx, event); err != nil {
		its.logger.Error("Failed to publish stock movement completed event", err, "movement_id", movement.ID)
	}

	// Check for low stock and publish alert if needed
	if item.IsLowStock() {
		alertEvent := events.NewLowStockAlertEvent(item)
		if err := its.eventPublisher.PublishEvent(ctx, alertEvent); err != nil {
			its.logger.Error("Failed to publish low stock alert event", err, "item_id", item.ID)
		}
	}

	// Check for out of stock and publish alert if needed
	if item.IsOutOfStock() {
		alertEvent := events.NewOutOfStockAlertEvent(item)
		if err := its.eventPublisher.PublishEvent(ctx, alertEvent); err != nil {
			its.logger.Error("Failed to publish out of stock alert event", err, "item_id", item.ID)
		}
	}

	// Check for reorder needed and publish alert if needed
	if item.NeedsReorder() {
		alertEvent := events.NewReorderNeededEvent(item)
		if err := its.eventPublisher.PublishEvent(ctx, alertEvent); err != nil {
			its.logger.Error("Failed to publish reorder needed event", err, "item_id", item.ID)
		}
	}
}

// Request types for service operations
type ReceiveStockRequest struct {
	ItemID          uuid.UUID      `json:"item_id"`
	Quantity        float64        `json:"quantity"`
	UnitCost        entities.Money `json:"unit_cost"`
	SupplierID      *uuid.UUID     `json:"supplier_id,omitempty"`
	LocationID      *uuid.UUID     `json:"location_id,omitempty"`
	BatchNumber     string         `json:"batch_number,omitempty"`
	ExpirationDate  *time.Time     `json:"expiration_date,omitempty"`
	ManufactureDate *time.Time     `json:"manufacture_date,omitempty"`
	Reason          string         `json:"reason"`
}

type IssueStockRequest struct {
	ItemID          uuid.UUID  `json:"item_id"`
	Quantity        float64    `json:"quantity"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	ReferenceType   string     `json:"reference_type,omitempty"`
	ReferenceID     *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceNumber string     `json:"reference_number,omitempty"`
	Reason          string     `json:"reason"`
}

type TransferStockRequest struct {
	ItemID         uuid.UUID `json:"item_id"`
	Quantity       float64   `json:"quantity"`
	FromLocationID uuid.UUID `json:"from_location_id"`
	ToLocationID   uuid.UUID `json:"to_location_id"`
	FromZone       string    `json:"from_zone,omitempty"`
	ToZone         string    `json:"to_zone,omitempty"`
	Reason         string    `json:"reason"`
}

type AdjustStockRequest struct {
	ItemID     uuid.UUID  `json:"item_id"`
	Adjustment float64    `json:"adjustment"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
	Reason     string     `json:"reason"`
}
