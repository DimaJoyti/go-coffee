package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/inventory-manager-agent/internal/application/commands"
	"go-coffee-ai-agents/inventory-manager-agent/internal/application/queries"
	"go-coffee-ai-agents/inventory-manager-agent/internal/application/responses"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/entities"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/repositories"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/services"
)

// InventoryManagementUseCase handles inventory management operations
type InventoryManagementUseCase struct {
	inventoryRepo       repositories.InventoryRepository
	movementRepo        repositories.StockMovementRepository
	locationRepo        repositories.LocationRepository
	supplierRepo        repositories.SupplierRepository
	trackingService     *services.InventoryTrackingService
	demandForecastSvc   DemandForecastingService
	autoReorderSvc      AutoReorderService
	qualityMgmtSvc      QualityManagementService
	logger              services.Logger
}

// DemandForecastingService defines demand forecasting capabilities
type DemandForecastingService interface {
	ForecastDemand(ctx context.Context, itemID uuid.UUID, period time.Duration) (*services.DemandForecast, error)
	GetConsumptionPattern(ctx context.Context, itemID uuid.UUID, period time.Duration) (*services.ConsumptionPattern, error)
	PredictStockout(ctx context.Context, itemID uuid.UUID) (*services.StockoutPrediction, error)
}

// AutoReorderService defines automatic reordering capabilities
type AutoReorderService interface {
	EvaluateReorderNeeds(ctx context.Context, locationID *uuid.UUID) ([]*services.ReorderRecommendation, error)
	CreateAutomaticPurchaseOrder(ctx context.Context, recommendations []*services.ReorderRecommendation) ([]*entities.PurchaseOrder, error)
	GetReorderRules(ctx context.Context, itemID uuid.UUID) (*services.ReorderRules, error)
	UpdateReorderRules(ctx context.Context, itemID uuid.UUID, rules *services.ReorderRules) error
}

// QualityManagementService defines quality management capabilities
type QualityManagementService interface {
	PerformQualityCheck(ctx context.Context, batchID uuid.UUID, checkType string) (*entities.QualityCheck, error)
	GetExpiringItems(ctx context.Context, within time.Duration, locationID *uuid.UUID) ([]*services.ExpiringItem, error)
	ProcessExpiredItems(ctx context.Context, locationID *uuid.UUID) (*services.ExpirationProcessingResult, error)
	GetQualityMetrics(ctx context.Context, period time.Duration) (*services.QualityMetrics, error)
}

// NewInventoryManagementUseCase creates a new inventory management use case
func NewInventoryManagementUseCase(
	inventoryRepo repositories.InventoryRepository,
	movementRepo repositories.StockMovementRepository,
	locationRepo repositories.LocationRepository,
	supplierRepo repositories.SupplierRepository,
	trackingService *services.InventoryTrackingService,
	demandForecastSvc DemandForecastingService,
	autoReorderSvc AutoReorderService,
	qualityMgmtSvc QualityManagementService,
	logger services.Logger,
) *InventoryManagementUseCase {
	return &InventoryManagementUseCase{
		inventoryRepo:     inventoryRepo,
		movementRepo:      movementRepo,
		locationRepo:      locationRepo,
		supplierRepo:      supplierRepo,
		trackingService:   trackingService,
		demandForecastSvc: demandForecastSvc,
		autoReorderSvc:    autoReorderSvc,
		qualityMgmtSvc:    qualityMgmtSvc,
		logger:            logger,
	}
}

// CreateInventoryItem creates a new inventory item with validation
func (uc *InventoryManagementUseCase) CreateInventoryItem(ctx context.Context, cmd *commands.CreateInventoryItemCommand) (*responses.InventoryItemResponse, error) {
	uc.logger.Info("Creating inventory item", "sku", cmd.SKU, "name", cmd.Name)

	// Create inventory item entity
	item := entities.NewInventoryItem(cmd.Name, cmd.SKU, cmd.Category, cmd.Unit)
	item.Description = cmd.Description
	item.SubCategory = cmd.SubCategory
	item.MinimumLevel = cmd.MinimumLevel
	item.MaximumLevel = cmd.MaximumLevel
	item.ReorderPoint = cmd.ReorderPoint
	item.ReorderQuantity = cmd.ReorderQuantity
	item.SafetyStock = cmd.SafetyStock
	item.UnitCost = cmd.UnitCost
	item.LocationID = cmd.LocationID
	item.SupplierID = cmd.SupplierID
	item.IsPerishable = cmd.IsPerishable
	item.ShelfLife = cmd.ShelfLife
	item.StorageConditions = cmd.StorageConditions
	item.Attributes = cmd.Attributes
	item.Tags = cmd.Tags
	item.CreatedBy = cmd.CreatedBy

	// Create the item using domain service
	if err := uc.trackingService.CreateInventoryItem(ctx, item); err != nil {
		uc.logger.Error("Failed to create inventory item", err, "sku", cmd.SKU)
		return nil, err
	}

	return uc.mapToInventoryItemResponse(item), nil
}

// UpdateInventoryItem updates an existing inventory item
func (uc *InventoryManagementUseCase) UpdateInventoryItem(ctx context.Context, cmd *commands.UpdateInventoryItemCommand) (*responses.InventoryItemResponse, error) {
	uc.logger.Info("Updating inventory item", "item_id", cmd.ID)

	// Get existing item
	item, err := uc.inventoryRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		uc.logger.Error("Failed to get inventory item", err, "item_id", cmd.ID)
		return nil, err
	}

	// Update fields
	if cmd.Name != nil {
		item.Name = *cmd.Name
	}
	if cmd.Description != nil {
		item.Description = *cmd.Description
	}
	if cmd.MinimumLevel != nil {
		item.MinimumLevel = *cmd.MinimumLevel
	}
	if cmd.MaximumLevel != nil {
		item.MaximumLevel = *cmd.MaximumLevel
	}
	if cmd.ReorderPoint != nil {
		item.ReorderPoint = *cmd.ReorderPoint
	}
	if cmd.ReorderQuantity != nil {
		item.ReorderQuantity = *cmd.ReorderQuantity
	}
	if cmd.SafetyStock != nil {
		item.SafetyStock = *cmd.SafetyStock
	}
	if cmd.UnitCost != nil {
		item.UnitCost = *cmd.UnitCost
	}
	if cmd.SupplierID != nil {
		item.SupplierID = *cmd.SupplierID
	}
	if cmd.StorageConditions != nil {
		item.StorageConditions = cmd.StorageConditions
	}
	if cmd.Attributes != nil {
		item.Attributes = cmd.Attributes
	}
	if cmd.Tags != nil {
		item.Tags = cmd.Tags
	}
	if cmd.IsActive != nil {
		item.IsActive = *cmd.IsActive
	}
	item.UpdatedBy = cmd.UpdatedBy

	// Update using domain service
	if err := uc.trackingService.UpdateInventoryItem(ctx, item); err != nil {
		uc.logger.Error("Failed to update inventory item", err, "item_id", cmd.ID)
		return nil, err
	}

	return uc.mapToInventoryItemResponse(item), nil
}

// ReceiveStock processes stock receipt from suppliers or transfers
func (uc *InventoryManagementUseCase) ReceiveStock(ctx context.Context, cmd *commands.ReceiveStockCommand) (*responses.StockMovementResponse, error) {
	uc.logger.Info("Receiving stock", "item_id", cmd.ItemID, "quantity", cmd.Quantity)

	// Create receive stock request
	request := &services.ReceiveStockRequest{
		ItemID:          cmd.ItemID,
		Quantity:        cmd.Quantity,
		UnitCost:        cmd.UnitCost,
		SupplierID:      cmd.SupplierID,
		LocationID:      cmd.LocationID,
		BatchNumber:     cmd.BatchNumber,
		ExpirationDate:  cmd.ExpirationDate,
		ManufactureDate: cmd.ManufactureDate,
		Reason:          cmd.Reason,
	}

	// Process using domain service
	if err := uc.trackingService.ReceiveStock(ctx, request); err != nil {
		uc.logger.Error("Failed to receive stock", err, "item_id", cmd.ItemID)
		return nil, err
	}

	// Get the created movement for response
	movements, err := uc.movementRepo.ListByItem(ctx, cmd.ItemID, &repositories.MovementFilter{
		Types:  []entities.MovementType{entities.MovementTypeReceipt},
		Limit:  1,
		SortBy: "created_at",
		SortOrder: "desc",
	})
	if err != nil || len(movements) == 0 {
		return nil, fmt.Errorf("failed to get created movement")
	}

	return uc.mapToStockMovementResponse(movements[0]), nil
}

// IssueStock processes stock issuance for consumption or transfer
func (uc *InventoryManagementUseCase) IssueStock(ctx context.Context, cmd *commands.IssueStockCommand) (*responses.StockMovementResponse, error) {
	uc.logger.Info("Issuing stock", "item_id", cmd.ItemID, "quantity", cmd.Quantity)

	// Create issue stock request
	request := &services.IssueStockRequest{
		ItemID:          cmd.ItemID,
		Quantity:        cmd.Quantity,
		LocationID:      cmd.LocationID,
		ReferenceType:   cmd.ReferenceType,
		ReferenceID:     cmd.ReferenceID,
		ReferenceNumber: cmd.ReferenceNumber,
		Reason:          cmd.Reason,
	}

	// Process using domain service
	if err := uc.trackingService.IssueStock(ctx, request); err != nil {
		uc.logger.Error("Failed to issue stock", err, "item_id", cmd.ItemID)
		return nil, err
	}

	// Get the created movement for response
	movements, err := uc.movementRepo.ListByItem(ctx, cmd.ItemID, &repositories.MovementFilter{
		Types:  []entities.MovementType{entities.MovementTypeIssue},
		Limit:  1,
		SortBy: "created_at",
		SortOrder: "desc",
	})
	if err != nil || len(movements) == 0 {
		return nil, fmt.Errorf("failed to get created movement")
	}

	return uc.mapToStockMovementResponse(movements[0]), nil
}

// TransferStock processes stock transfer between locations
func (uc *InventoryManagementUseCase) TransferStock(ctx context.Context, cmd *commands.TransferStockCommand) (*responses.StockMovementResponse, error) {
	uc.logger.Info("Transferring stock", 
		"item_id", cmd.ItemID,
		"quantity", cmd.Quantity,
		"from_location", cmd.FromLocationID,
		"to_location", cmd.ToLocationID)

	// Create transfer stock request
	request := &services.TransferStockRequest{
		ItemID:         cmd.ItemID,
		Quantity:       cmd.Quantity,
		FromLocationID: cmd.FromLocationID,
		ToLocationID:   cmd.ToLocationID,
		FromZone:       cmd.FromZone,
		ToZone:         cmd.ToZone,
		Reason:         cmd.Reason,
	}

	// Process using domain service
	if err := uc.trackingService.TransferStock(ctx, request); err != nil {
		uc.logger.Error("Failed to transfer stock", err, "item_id", cmd.ItemID)
		return nil, err
	}

	// Get the created movement for response
	movements, err := uc.movementRepo.ListByItem(ctx, cmd.ItemID, &repositories.MovementFilter{
		Types:  []entities.MovementType{entities.MovementTypeTransfer},
		Limit:  1,
		SortBy: "created_at",
		SortOrder: "desc",
	})
	if err != nil || len(movements) == 0 {
		return nil, fmt.Errorf("failed to get created movement")
	}

	return uc.mapToStockMovementResponse(movements[0]), nil
}

// AdjustStock processes stock adjustments for corrections
func (uc *InventoryManagementUseCase) AdjustStock(ctx context.Context, cmd *commands.AdjustStockCommand) (*responses.StockMovementResponse, error) {
	uc.logger.Info("Adjusting stock", "item_id", cmd.ItemID, "adjustment", cmd.Adjustment)

	// Create adjust stock request
	request := &services.AdjustStockRequest{
		ItemID:     cmd.ItemID,
		Adjustment: cmd.Adjustment,
		LocationID: cmd.LocationID,
		Reason:     cmd.Reason,
	}

	// Process using domain service
	if err := uc.trackingService.AdjustStock(ctx, request); err != nil {
		uc.logger.Error("Failed to adjust stock", err, "item_id", cmd.ItemID)
		return nil, err
	}

	// Get the created movement for response
	movements, err := uc.movementRepo.ListByItem(ctx, cmd.ItemID, &repositories.MovementFilter{
		Types:  []entities.MovementType{entities.MovementTypeAdjustment},
		Limit:  1,
		SortBy: "created_at",
		SortOrder: "desc",
	})
	if err != nil || len(movements) == 0 {
		return nil, fmt.Errorf("failed to get created movement")
	}

	return uc.mapToStockMovementResponse(movements[0]), nil
}

// GetInventoryOverview provides a comprehensive inventory overview
func (uc *InventoryManagementUseCase) GetInventoryOverview(ctx context.Context, query *queries.InventoryOverviewQuery) (*responses.InventoryOverviewResponse, error) {
	uc.logger.Info("Getting inventory overview", "location_id", query.LocationID)

	// Get inventory items
	filter := &repositories.InventoryFilter{
		IsActive: &[]bool{true}[0],
		Limit:    1000,
		SortBy:   "name",
		SortOrder: "asc",
	}
	if query.LocationID != nil {
		filter.LocationIDs = []uuid.UUID{*query.LocationID}
	}

	items, err := uc.inventoryRepo.List(ctx, filter)
	if err != nil {
		uc.logger.Error("Failed to get inventory items", err)
		return nil, err
	}

	// Get metrics
	metrics, err := uc.inventoryRepo.GetInventoryMetrics(ctx, query.LocationID, 30*24*time.Hour)
	if err != nil {
		uc.logger.Error("Failed to get inventory metrics", err)
		return nil, err
	}

	// Get alerts
	lowStockItems, _ := uc.inventoryRepo.GetLowStockItems(ctx, query.LocationID)
	outOfStockItems, _ := uc.inventoryRepo.GetOutOfStockItems(ctx, query.LocationID)
	reorderItems, _ := uc.inventoryRepo.GetItemsNeedingReorder(ctx, query.LocationID)
	expiringItems, _ := uc.inventoryRepo.GetExpiringItems(ctx, 7*24*time.Hour, query.LocationID)

	// Build response
	response := &responses.InventoryOverviewResponse{
		TotalItems:      len(items),
		TotalValue:      metrics.TotalValue,
		LowStockCount:   len(lowStockItems),
		OutOfStockCount: len(outOfStockItems),
		ReorderCount:    len(reorderItems),
		ExpiringCount:   len(expiringItems),
		Items:           make([]*responses.InventoryItemSummary, len(items)),
		Metrics:         uc.mapToInventoryMetricsResponse(metrics),
		Alerts:          uc.buildInventoryAlerts(lowStockItems, outOfStockItems, reorderItems, expiringItems),
	}

	for i, item := range items {
		response.Items[i] = uc.mapToInventoryItemSummary(item)
	}

	return response, nil
}

// GetDemandForecast provides demand forecasting for inventory planning
func (uc *InventoryManagementUseCase) GetDemandForecast(ctx context.Context, query *queries.DemandForecastQuery) (*responses.DemandForecastResponse, error) {
	uc.logger.Info("Getting demand forecast", "item_id", query.ItemID, "period", query.Period)

	// Get demand forecast
	forecast, err := uc.demandForecastSvc.ForecastDemand(ctx, query.ItemID, query.Period)
	if err != nil {
		uc.logger.Error("Failed to get demand forecast", err, "item_id", query.ItemID)
		return nil, err
	}

	// Get consumption pattern
	pattern, err := uc.demandForecastSvc.GetConsumptionPattern(ctx, query.ItemID, query.Period)
	if err != nil {
		uc.logger.Error("Failed to get consumption pattern", err, "item_id", query.ItemID)
		return nil, err
	}

	// Get stockout prediction
	stockoutPrediction, err := uc.demandForecastSvc.PredictStockout(ctx, query.ItemID)
	if err != nil {
		uc.logger.Error("Failed to get stockout prediction", err, "item_id", query.ItemID)
		return nil, err
	}

	return &responses.DemandForecastResponse{
		ItemID:             query.ItemID,
		Period:             query.Period,
		Forecast:           uc.convertDemandForecastToResponse(forecast),
		ConsumptionPattern: uc.convertConsumptionPatternToResponse(pattern),
		StockoutPrediction: uc.convertStockoutPredictionToResponse(stockoutPrediction),
		GeneratedAt:        time.Now(),
	}, nil
}

// GetReorderRecommendations provides automatic reorder recommendations
func (uc *InventoryManagementUseCase) GetReorderRecommendations(ctx context.Context, query *queries.ReorderRecommendationsQuery) (*responses.ReorderRecommendationsResponse, error) {
	uc.logger.Info("Getting reorder recommendations", "location_id", query.LocationID)

	// Get reorder recommendations
	recommendations, err := uc.autoReorderSvc.EvaluateReorderNeeds(ctx, query.LocationID)
	if err != nil {
		uc.logger.Error("Failed to get reorder recommendations", err)
		return nil, err
	}

	return &responses.ReorderRecommendationsResponse{
		LocationID:      query.LocationID,
		Recommendations: uc.convertReorderRecommendationsToResponse(recommendations),
		TotalItems:      len(recommendations),
		GeneratedAt:     time.Now(),
	}, nil
}

// ProcessAutomaticReorders creates purchase orders based on reorder recommendations
func (uc *InventoryManagementUseCase) ProcessAutomaticReorders(ctx context.Context, cmd *commands.ProcessAutomaticReordersCommand) (*responses.AutoReorderProcessingResponse, error) {
	uc.logger.Info("Processing automatic reorders", "location_id", cmd.LocationID)

	// Get reorder recommendations
	recommendations, err := uc.autoReorderSvc.EvaluateReorderNeeds(ctx, cmd.LocationID)
	if err != nil {
		uc.logger.Error("Failed to get reorder recommendations", err)
		return nil, err
	}

	// Filter recommendations based on criteria
	filteredRecommendations := uc.filterReorderRecommendations(recommendations, cmd.Criteria)

	// Create purchase orders
	orders, err := uc.autoReorderSvc.CreateAutomaticPurchaseOrder(ctx, filteredRecommendations)
	if err != nil {
		uc.logger.Error("Failed to create automatic purchase orders", err)
		return nil, err
	}

	return &responses.AutoReorderProcessingResponse{
		ProcessedRecommendations: len(filteredRecommendations),
		CreatedOrders:           len(orders),
		Orders:                  orders,
		ProcessedAt:             time.Now(),
	}, nil
}

// Helper methods for mapping and filtering

func (uc *InventoryManagementUseCase) mapToInventoryItemResponse(item *entities.InventoryItem) *responses.InventoryItemResponse {
	return &responses.InventoryItemResponse{
		ID:                item.ID,
		SKU:               item.SKU,
		Name:              item.Name,
		Description:       item.Description,
		Category:          item.Category,
		SubCategory:       item.SubCategory,
		Unit:              item.Unit,
		CurrentStock:      item.CurrentStock,
		ReservedStock:     item.ReservedStock,
		AvailableStock:    item.AvailableStock,
		MinimumLevel:      item.MinimumLevel,
		MaximumLevel:      item.MaximumLevel,
		ReorderPoint:      item.ReorderPoint,
		ReorderQuantity:   item.ReorderQuantity,
		SafetyStock:       item.SafetyStock,
		UnitCost:          item.UnitCost,
		TotalValue:        item.TotalValue,
		AverageCost:       item.AverageCost,
		LastCost:          item.LastCost,
		SupplierID:        item.SupplierID,
		LocationID:        item.LocationID,
		Status:            item.Status,
		IsActive:          item.IsActive,
		IsPerishable:      item.IsPerishable,
		ShelfLife:         item.ShelfLife,
		StorageConditions: item.StorageConditions,
		Attributes:        item.Attributes,
		Tags:              item.Tags,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		CreatedBy:         item.CreatedBy,
		UpdatedBy:         item.UpdatedBy,
		Version:           item.Version,
	}
}

func (uc *InventoryManagementUseCase) mapToStockMovementResponse(movement *entities.StockMovement) *responses.StockMovementResponse {
	return &responses.StockMovementResponse{
		ID:              movement.ID,
		MovementNumber:  movement.MovementNumber,
		Type:            movement.Type,
		Status:          movement.Status,
		Direction:       movement.Direction,
		InventoryItemID: movement.InventoryItemID,
		Quantity:        movement.Quantity,
		Unit:            movement.Unit,
		UnitCost:        movement.UnitCost,
		TotalCost:       movement.TotalCost,
		FromLocationID:  movement.FromLocationID,
		ToLocationID:    movement.ToLocationID,
		Reason:          movement.Reason,
		ProcessedAt:     movement.ProcessedAt,
		ProcessedBy:     movement.ProcessedBy,
		CompletedAt:     movement.CompletedAt,
		CreatedAt:       movement.CreatedAt,
		UpdatedAt:       movement.UpdatedAt,
		CreatedBy:       movement.CreatedBy,
		UpdatedBy:       movement.UpdatedBy,
	}
}

func (uc *InventoryManagementUseCase) mapToInventoryItemSummary(item *entities.InventoryItem) *responses.InventoryItemSummary {
	return &responses.InventoryItemSummary{
		ID:             item.ID,
		SKU:            item.SKU,
		Name:           item.Name,
		Category:       item.Category,
		CurrentStock:   item.CurrentStock,
		AvailableStock: item.AvailableStock,
		MinimumLevel:   item.MinimumLevel,
		ReorderPoint:   item.ReorderPoint,
		Status:         item.Status,
		TotalValue:     item.TotalValue,
		IsLowStock:     item.IsLowStock(),
		IsOutOfStock:   item.IsOutOfStock(),
		NeedsReorder:   item.NeedsReorder(),
	}
}

func (uc *InventoryManagementUseCase) mapToInventoryMetricsResponse(metrics *repositories.InventoryMetrics) *responses.InventoryMetricsResponse {
	return &responses.InventoryMetricsResponse{
		Period:              metrics.Period,
		TotalItems:          metrics.TotalItems,
		TotalValue:          metrics.TotalValue,
		AverageValue:        metrics.AverageValue,
		TotalMovements:      metrics.TotalMovements,
		InboundMovements:    metrics.InboundMovements,
		OutboundMovements:   metrics.OutboundMovements,
		AdjustmentMovements: metrics.AdjustmentMovements,
		LowStockItems:       metrics.LowStockItems,
		OutOfStockItems:     metrics.OutOfStockItems,
		ExpiringItems:       metrics.ExpiringItems,
		TurnoverRate:        metrics.TurnoverRate,
		StockAccuracy:       metrics.StockAccuracy,
		CategoryBreakdown:   metrics.CategoryBreakdown,
		TopMovingItems:      metrics.TopMovingItems,
		SlowMovingItems:     metrics.SlowMovingItems,
		GeneratedAt:         metrics.GeneratedAt,
	}
}

func (uc *InventoryManagementUseCase) buildInventoryAlerts(lowStock, outOfStock, reorder, expiring []*entities.InventoryItem) []*responses.InventoryAlert {
	var alerts []*responses.InventoryAlert

	for _, item := range lowStock {
		alerts = append(alerts, &responses.InventoryAlert{
			Type:    "low_stock",
			ItemID:  item.ID,
			SKU:     item.SKU,
			Name:    item.Name,
			Message: fmt.Sprintf("Low stock: %v %s remaining (minimum: %v)", item.CurrentStock, item.Unit, item.MinimumLevel),
			Severity: "warning",
		})
	}

	for _, item := range outOfStock {
		alerts = append(alerts, &responses.InventoryAlert{
			Type:    "out_of_stock",
			ItemID:  item.ID,
			SKU:     item.SKU,
			Name:    item.Name,
			Message: "Item is out of stock",
			Severity: "critical",
		})
	}

	for _, item := range reorder {
		alerts = append(alerts, &responses.InventoryAlert{
			Type:    "reorder_needed",
			ItemID:  item.ID,
			SKU:     item.SKU,
			Name:    item.Name,
			Message: fmt.Sprintf("Reorder needed: %v %s remaining (reorder point: %v)", item.CurrentStock, item.Unit, item.ReorderPoint),
			Severity: "info",
		})
	}

	for _, item := range expiring {
		alerts = append(alerts, &responses.InventoryAlert{
			Type:    "expiring_soon",
			ItemID:  item.ID,
			SKU:     item.SKU,
			Name:    item.Name,
			Message: "Items expiring within 7 days",
			Severity: "warning",
		})
	}

	return alerts
}

func (uc *InventoryManagementUseCase) filterReorderRecommendations(recommendations []*services.ReorderRecommendation, criteria *commands.AutoReorderCriteria) []*services.ReorderRecommendation {
	if criteria == nil {
		return recommendations
	}

	var filtered []*services.ReorderRecommendation
	for _, rec := range recommendations {
		if criteria.MaxOrderValue != nil && rec.EstimatedCost.Amount > *criteria.MaxOrderValue {
			continue
		}
		if criteria.MinUrgencyScore != nil && rec.UrgencyScore < *criteria.MinUrgencyScore {
			continue
		}
		if criteria.RequireApproval != nil && *criteria.RequireApproval && rec.EstimatedCost.Amount > 1000 {
			continue // Skip high-value orders if approval is required
		}
		filtered = append(filtered, rec)
	}
	return filtered
}

// Conversion methods between services domain types and response DTOs

func (uc *InventoryManagementUseCase) convertDemandForecastToResponse(forecast *services.DemandForecast) *responses.DemandForecast {
	if forecast == nil {
		return nil
	}
	
	return &responses.DemandForecast{
		PredictedDemand:    forecast.PredictedDemand,
		ConfidenceInterval: uc.convertConfidenceIntervalToResponse(forecast.ConfidenceInterval),
		SeasonalFactors:    forecast.SeasonalFactors,
		TrendFactors:       forecast.TrendFactors,
		ForecastPoints:     uc.convertForecastPointsToResponse(forecast.ForecastPoints),
		Accuracy:           forecast.Accuracy,
		Algorithm:          forecast.Algorithm,
		ModelVersion:       forecast.ModelVersion,
	}
}

func (uc *InventoryManagementUseCase) convertConsumptionPatternToResponse(pattern *services.ConsumptionPattern) *responses.ConsumptionPattern {
	if pattern == nil {
		return nil
	}
	
	return &responses.ConsumptionPattern{
		AverageDaily:     pattern.AverageDaily,
		AverageWeekly:    pattern.AverageWeekly,
		AverageMonthly:   pattern.AverageMonthly,
		Volatility:       pattern.Volatility,
		Seasonality:      pattern.Seasonality,
		DayOfWeekPattern: pattern.DayOfWeekPattern,
		MonthlyPattern:   pattern.MonthlyPattern,
		TrendDirection:   pattern.TrendDirection,
		TrendStrength:    pattern.TrendStrength,
	}
}

func (uc *InventoryManagementUseCase) convertStockoutPredictionToResponse(prediction *services.StockoutPrediction) *responses.StockoutPrediction {
	if prediction == nil {
		return nil
	}
	
	return &responses.StockoutPrediction{
		ProbabilityPercent: prediction.ProbabilityPercent,
		PredictedDate:      prediction.PredictedDate,
		DaysUntilStockout:  prediction.DaysUntilStockout,
		RecommendedAction:  prediction.RecommendedAction,
		UrgencyLevel:       prediction.UrgencyLevel,
		Confidence:         prediction.Confidence,
	}
}

func (uc *InventoryManagementUseCase) convertReorderRecommendationsToResponse(recommendations []*services.ReorderRecommendation) []*responses.ReorderRecommendation {
	result := make([]*responses.ReorderRecommendation, len(recommendations))
	for i, rec := range recommendations {
		result[i] = uc.convertReorderRecommendationToResponse(rec)
	}
	return result
}

func (uc *InventoryManagementUseCase) convertReorderRecommendationToResponse(rec *services.ReorderRecommendation) *responses.ReorderRecommendation {
	if rec == nil {
		return nil
	}
	
	return &responses.ReorderRecommendation{
		ItemID:         rec.ItemID,
		SKU:            rec.SKU,
		Name:           rec.Name,
		CurrentStock:   rec.CurrentStock,
		ReorderPoint:   rec.ReorderPoint,
		RecommendedQty: rec.RecommendedQty,
		EstimatedCost:  rec.EstimatedCost,
		PreferredSupplier: uc.convertSupplierRecommendationToSummary(rec.PreferredSupplier),
		UrgencyScore:   rec.UrgencyScore,
		LeadTimeDays:   rec.LeadTimeDays,
		StockoutRisk:   rec.StockoutRisk,
		Reason:         rec.Reason,
		Priority:       rec.Priority,
	}
}

// Helper conversion methods

func (uc *InventoryManagementUseCase) convertConfidenceIntervalToResponse(ci *services.ConfidenceInterval) *responses.ConfidenceInterval {
	if ci == nil {
		return nil
	}
	return &responses.ConfidenceInterval{
		Lower:      ci.Lower,
		Upper:      ci.Upper,
		Confidence: ci.Confidence,
	}
}

func (uc *InventoryManagementUseCase) convertForecastPointsToResponse(points []*services.ForecastPoint) []*responses.ForecastPoint {
	result := make([]*responses.ForecastPoint, len(points))
	for i, point := range points {
		result[i] = &responses.ForecastPoint{
			Date:       point.Date,
			Value:      point.Value,
			LowerBound: point.LowerBound,
			UpperBound: point.UpperBound,
			Confidence: point.Confidence,
		}
	}
	return result
}

func (uc *InventoryManagementUseCase) convertSupplierRecommendationToSummary(rec *services.SupplierRecommendation) *responses.SupplierSummary {
	if rec == nil {
		return nil
	}
	return &responses.SupplierSummary{
		ID:          rec.SupplierID,
		Name:        rec.SupplierName,
		Rating:      rec.QualityRating,
		IsPreferred: rec.IsPreferred,
		IsActive:    true, // Default to active for recommendations
	}
}
