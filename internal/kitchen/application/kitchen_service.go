package application

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// KitchenServiceImpl implements the KitchenService interface
type KitchenServiceImpl struct {
	repoManager         domain.RepositoryManager
	queueService        QueueService
	optimizerService    OptimizerService
	notificationService NotificationService
	eventService        EventService
	logger              *logger.Logger
}

// NewKitchenService creates a new kitchen service instance
func NewKitchenService(
	repoManager domain.RepositoryManager,
	queueService QueueService,
	optimizerService OptimizerService,
	notificationService NotificationService,
	eventService EventService,
	logger *logger.Logger,
) KitchenService {
	return &KitchenServiceImpl{
		repoManager:         repoManager,
		queueService:        queueService,
		optimizerService:    optimizerService,
		notificationService: notificationService,
		eventService:        eventService,
		logger:              logger,
	}
}

// Equipment Management

// CreateEquipment creates a new piece of kitchen equipment
func (s *KitchenServiceImpl) CreateEquipment(ctx context.Context, req *CreateEquipmentRequest) (*EquipmentResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"equipment_id": req.ID,
		"name":         req.Name,
		"station_type": req.StationType,
	}).Info("Creating new equipment")

	// Create domain entity
	equipment, err := domain.NewEquipment(req.ID, req.Name, req.StationType, req.MaxCapacity)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create equipment entity")
		return nil, fmt.Errorf("invalid equipment data: %w", err)
	}

	// Save to repository
	if err := s.repoManager.Equipment().Create(ctx, equipment); err != nil {
		s.logger.WithError(err).Error("Failed to save equipment")
		return nil, fmt.Errorf("failed to create equipment: %w", err)
	}

	// Publish event
	event := domain.NewEquipmentStatusChangedEvent(equipment, domain.EquipmentStatusUnknown)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish equipment created event")
	}

	s.logger.WithField("equipment_id", equipment.ID()).Info("Equipment created successfully")
	return s.equipmentToResponse(equipment), nil
}

// GetEquipment retrieves equipment by ID
func (s *KitchenServiceImpl) GetEquipment(ctx context.Context, id string) (*EquipmentResponse, error) {
	equipment, err := s.repoManager.Equipment().GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("equipment_id", id).Error("Failed to get equipment")
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	return s.equipmentToResponse(equipment), nil
}

// UpdateEquipmentStatus updates the status of equipment
func (s *KitchenServiceImpl) UpdateEquipmentStatus(ctx context.Context, id string, status domain.EquipmentStatus) error {
	equipment, err := s.repoManager.Equipment().GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}

	oldStatus := equipment.Status()
	if err := equipment.UpdateStatus(status); err != nil {
		return fmt.Errorf("failed to update equipment status: %w", err)
	}

	if err := s.repoManager.Equipment().Update(ctx, equipment); err != nil {
		return fmt.Errorf("failed to save equipment: %w", err)
	}

	// Publish event
	event := domain.NewEquipmentStatusChangedEvent(equipment, oldStatus)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish equipment status changed event")
	}

	// Send notification if equipment needs attention
	if status == domain.EquipmentStatusMaintenance || status == domain.EquipmentStatusBroken {
		if err := s.notificationService.NotifyEquipmentMaintenance(ctx, equipment); err != nil {
			s.logger.WithError(err).Warn("Failed to send equipment maintenance notification")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"equipment_id": id,
		"old_status":   oldStatus,
		"new_status":   status,
	}).Info("Equipment status updated")

	return nil
}

// ListEquipment lists equipment with optional filtering
func (s *KitchenServiceImpl) ListEquipment(ctx context.Context, filter *EquipmentFilter) ([]*EquipmentResponse, error) {
	var equipment []*domain.Equipment
	var err error

	if filter != nil {
		// Apply filters
		if filter.StationType != nil {
			equipment, err = s.repoManager.Equipment().GetByStationType(ctx, *filter.StationType)
		} else if filter.Status != nil {
			equipment, err = s.repoManager.Equipment().GetByStatus(ctx, *filter.Status)
		} else if filter.Available != nil && *filter.Available {
			equipment, err = s.repoManager.Equipment().GetAvailable(ctx)
		} else {
			equipment, err = s.repoManager.Equipment().GetAll(ctx)
		}
	} else {
		equipment, err = s.repoManager.Equipment().GetAll(ctx)
	}

	if err != nil {
		s.logger.WithError(err).Error("Failed to list equipment")
		return nil, fmt.Errorf("failed to list equipment: %w", err)
	}

	responses := make([]*EquipmentResponse, len(equipment))
	for i, eq := range equipment {
		responses[i] = s.equipmentToResponse(eq)
	}

	return responses, nil
}

// ScheduleEquipmentMaintenance schedules maintenance for equipment
func (s *KitchenServiceImpl) ScheduleEquipmentMaintenance(ctx context.Context, id string) error {
	equipment, err := s.repoManager.Equipment().GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}

	if err := equipment.ScheduleMaintenance(); err != nil {
		return fmt.Errorf("failed to schedule maintenance: %w", err)
	}

	if err := s.repoManager.Equipment().Update(ctx, equipment); err != nil {
		return fmt.Errorf("failed to save equipment: %w", err)
	}

	// Publish event
	event := domain.NewEquipmentMaintenanceScheduledEvent(equipment)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish maintenance scheduled event")
	}

	// Send notification
	if err := s.notificationService.NotifyEquipmentMaintenance(ctx, equipment); err != nil {
		s.logger.WithError(err).Warn("Failed to send maintenance notification")
	}

	s.logger.WithField("equipment_id", id).Info("Equipment maintenance scheduled")
	return nil
}

// Staff Management

// CreateStaff creates a new staff member
func (s *KitchenServiceImpl) CreateStaff(ctx context.Context, req *CreateStaffRequest) (*StaffResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"staff_id":        req.ID,
		"name":            req.Name,
		"specializations": req.Specializations,
		"skill_level":     req.SkillLevel,
	}).Info("Creating new staff member")

	// Create domain entity
	staff, err := domain.NewStaff(req.ID, req.Name, req.Specializations, req.SkillLevel, req.MaxConcurrentOrders)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create staff entity")
		return nil, fmt.Errorf("invalid staff data: %w", err)
	}

	// Save to repository
	if err := s.repoManager.Staff().Create(ctx, staff); err != nil {
		s.logger.WithError(err).Error("Failed to save staff")
		return nil, fmt.Errorf("failed to create staff: %w", err)
	}

	s.logger.WithField("staff_id", staff.ID()).Info("Staff member created successfully")
	return s.staffToResponse(staff), nil
}

// GetStaff retrieves staff by ID
func (s *KitchenServiceImpl) GetStaff(ctx context.Context, id string) (*StaffResponse, error) {
	staff, err := s.repoManager.Staff().GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("staff_id", id).Error("Failed to get staff")
		return nil, fmt.Errorf("failed to get staff: %w", err)
	}

	return s.staffToResponse(staff), nil
}

// UpdateStaffAvailability updates staff availability
func (s *KitchenServiceImpl) UpdateStaffAvailability(ctx context.Context, id string, available bool) error {
	staff, err := s.repoManager.Staff().GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get staff: %w", err)
	}

	oldAvailability := staff.IsAvailable()
	staff.SetAvailability(available)

	if err := s.repoManager.Staff().Update(ctx, staff); err != nil {
		return fmt.Errorf("failed to save staff: %w", err)
	}

	// Publish event
	event := domain.NewStaffAvailabilityChangedEvent(staff, oldAvailability)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish staff availability changed event")
	}

	s.logger.WithFields(map[string]interface{}{
		"staff_id":         id,
		"old_availability": oldAvailability,
		"new_availability": available,
	}).Info("Staff availability updated")

	return nil
}

// ListStaff lists staff with optional filtering
func (s *KitchenServiceImpl) ListStaff(ctx context.Context, filter *StaffFilter) ([]*StaffResponse, error) {
	var staff []*domain.Staff
	var err error

	if filter != nil {
		// Apply filters
		if filter.Specialization != nil {
			if filter.Available != nil && *filter.Available {
				staff, err = s.repoManager.Staff().GetAvailableBySpecialization(ctx, *filter.Specialization)
			} else {
				staff, err = s.repoManager.Staff().GetBySpecialization(ctx, *filter.Specialization)
			}
		} else if filter.Available != nil && *filter.Available {
			staff, err = s.repoManager.Staff().GetAvailable(ctx)
		} else {
			staff, err = s.repoManager.Staff().GetAll(ctx)
		}
	} else {
		staff, err = s.repoManager.Staff().GetAll(ctx)
	}

	if err != nil {
		s.logger.WithError(err).Error("Failed to list staff")
		return nil, fmt.Errorf("failed to list staff: %w", err)
	}

	responses := make([]*StaffResponse, len(staff))
	for i, st := range staff {
		responses[i] = s.staffToResponse(st)
	}

	return responses, nil
}

// UpdateStaffSkill updates staff skill level
func (s *KitchenServiceImpl) UpdateStaffSkill(ctx context.Context, id string, skillLevel float32) error {
	staff, err := s.repoManager.Staff().GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get staff: %w", err)
	}

	if err := staff.UpdateSkillLevel(skillLevel); err != nil {
		return fmt.Errorf("failed to update skill level: %w", err)
	}

	if err := s.repoManager.Staff().Update(ctx, staff); err != nil {
		return fmt.Errorf("failed to save staff: %w", err)
	}

	// Publish event
	event := domain.NewStaffSkillUpdatedEvent(staff)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish staff skill updated event")
	}

	s.logger.WithFields(map[string]interface{}{
		"staff_id":    id,
		"skill_level": skillLevel,
	}).Info("Staff skill level updated")

	return nil
}

// Helper methods for converting domain entities to responses

func (s *KitchenServiceImpl) equipmentToResponse(equipment *domain.Equipment) *EquipmentResponse {
	return &EquipmentResponse{
		ID:               equipment.ID(),
		Name:             equipment.Name(),
		StationType:      equipment.StationType(),
		Status:           equipment.Status(),
		EfficiencyScore:  equipment.EfficiencyScore(),
		CurrentLoad:      equipment.CurrentLoad(),
		MaxCapacity:      equipment.MaxCapacity(),
		UtilizationRate:  equipment.GetUtilizationRate(),
		IsAvailable:      equipment.IsAvailable(),
		NeedsMaintenance: equipment.NeedsMaintenance(),
		LastMaintenance:  equipment.LastMaintenance(),
		CreatedAt:        equipment.CreatedAt(),
		UpdatedAt:        equipment.UpdatedAt(),
	}
}

func (s *KitchenServiceImpl) staffToResponse(staff *domain.Staff) *StaffResponse {
	return &StaffResponse{
		ID:                  staff.ID(),
		Name:                staff.Name(),
		Specializations:     staff.Specializations(),
		SkillLevel:          staff.SkillLevel(),
		IsAvailable:         staff.IsAvailable(),
		CurrentOrders:       staff.CurrentOrders(),
		MaxConcurrentOrders: staff.MaxConcurrentOrders(),
		Workload:            staff.GetWorkload(),
		IsOverloaded:        staff.IsOverloaded(),
		CreatedAt:           staff.CreatedAt(),
		UpdatedAt:           staff.UpdatedAt(),
	}
}

// Order Management

// AddOrderToQueue adds an order to the kitchen queue
func (s *KitchenServiceImpl) AddOrderToQueue(ctx context.Context, req *AddOrderRequest) (*OrderResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"order_id":    req.ID,
		"customer_id": req.CustomerID,
		"items_count": len(req.Items),
		"priority":    req.Priority,
	}).Info("Adding order to kitchen queue")

	// Convert request items to domain items
	items := make([]*domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.NewOrderItem(item.ID, item.Name, item.Quantity, item.Requirements)
		items[i].SetInstructions(item.Instructions)
		items[i].SetMetadata(item.Metadata)
	}

	// Create domain entity
	order, err := domain.NewKitchenOrder(req.ID, req.CustomerID, items)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create kitchen order entity")
		return nil, fmt.Errorf("invalid order data: %w", err)
	}

	// Set priority and special instructions
	if req.Priority != 0 {
		order.SetPriority(req.Priority)
	}
	if req.SpecialInstructions != "" {
		order.SetSpecialInstructions(req.SpecialInstructions)
	}

	// Predict preparation time using AI
	estimatedTime, err := s.optimizerService.PredictPreparationTime(ctx, order)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to predict preparation time, using default")
		estimatedTime = 300 // 5 minutes default
	}
	order.SetEstimatedTime(estimatedTime)

	// Save order to repository
	if err := s.repoManager.Order().Create(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to save order")
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Add to queue
	if err := s.queueService.AddOrder(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to add order to queue")
		return nil, fmt.Errorf("failed to add order to queue: %w", err)
	}

	// Publish event
	event := domain.NewOrderAddedToQueueEvent(order)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish order added event")
	}

	// Send notification
	if err := s.notificationService.NotifyOrderAdded(ctx, order); err != nil {
		s.logger.WithError(err).Warn("Failed to send order added notification")
	}

	s.logger.WithField("order_id", order.ID()).Info("Order added to kitchen queue successfully")
	return s.orderToResponse(order), nil
}

// GetOrder retrieves an order by ID
func (s *KitchenServiceImpl) GetOrder(ctx context.Context, id string) (*OrderResponse, error) {
	order, err := s.repoManager.Order().GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("order_id", id).Error("Failed to get order")
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return s.orderToResponse(order), nil
}

// UpdateOrderStatus updates the status of an order
func (s *KitchenServiceImpl) UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	order, err := s.repoManager.Order().GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	oldStatus := order.Status()
	if err := order.UpdateStatus(status); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if err := s.repoManager.Order().Update(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Publish event
	event := domain.NewOrderStatusChangedEvent(order, oldStatus)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish order status changed event")
	}

	// Send notifications based on status
	switch status {
	case domain.OrderStatusProcessing:
		event := domain.NewOrderStartedEvent(order)
		s.eventService.PublishEvent(ctx, event)
	case domain.OrderStatusCompleted:
		event := domain.NewOrderCompletedEvent(order)
		s.eventService.PublishEvent(ctx, event)
		s.notificationService.NotifyOrderCompleted(ctx, order)
	}

	s.logger.WithFields(map[string]interface{}{
		"order_id":   id,
		"old_status": oldStatus,
		"new_status": status,
	}).Info("Order status updated")

	return nil
}

// UpdateOrderPriority updates the priority of an order
func (s *KitchenServiceImpl) UpdateOrderPriority(ctx context.Context, id string, priority domain.OrderPriority) error {
	order, err := s.repoManager.Order().GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	oldPriority := order.Priority()
	order.SetPriority(priority)

	if err := s.repoManager.Order().Update(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Update queue priority
	if err := s.queueService.UpdateOrderPriority(ctx, id, priority); err != nil {
		s.logger.WithError(err).Warn("Failed to update order priority in queue")
	}

	s.logger.WithFields(map[string]interface{}{
		"order_id":     id,
		"old_priority": oldPriority,
		"new_priority": priority,
	}).Info("Order priority updated")

	return nil
}

// AssignOrderToStaff assigns an order to a staff member
func (s *KitchenServiceImpl) AssignOrderToStaff(ctx context.Context, orderID, staffID string) error {
	// Get order and staff
	order, err := s.repoManager.Order().GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	staff, err := s.repoManager.Staff().GetByID(ctx, staffID)
	if err != nil {
		return fmt.Errorf("failed to get staff: %w", err)
	}

	// Validate assignment
	if !staff.CanAcceptOrder() {
		return fmt.Errorf("staff member cannot accept more orders")
	}

	// Check if staff can handle required stations
	requiredStations := order.GetRequiredStations()
	canHandle := false
	for _, station := range requiredStations {
		if staff.CanHandleStation(station) {
			canHandle = true
			break
		}
	}
	if !canHandle {
		return fmt.Errorf("staff member cannot handle required stations")
	}

	// Assign order to staff
	if err := order.AssignStaff(staffID); err != nil {
		return fmt.Errorf("failed to assign staff to order: %w", err)
	}

	// Assign order to staff
	if err := staff.AssignOrder(); err != nil {
		return fmt.Errorf("failed to assign order to staff: %w", err)
	}

	// Save both entities
	if err := s.repoManager.Order().Update(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	if err := s.repoManager.Staff().Update(ctx, staff); err != nil {
		return fmt.Errorf("failed to save staff: %w", err)
	}

	// Publish events
	orderEvent := domain.NewOrderAssignedEvent(order)
	staffEvent := domain.NewStaffAssignedEvent(staff, orderID)

	s.eventService.PublishEvent(ctx, orderEvent)
	s.eventService.PublishEvent(ctx, staffEvent)

	// Send notification
	if err := s.notificationService.NotifyStaffAssigned(ctx, staff, order); err != nil {
		s.logger.WithError(err).Warn("Failed to send staff assignment notification")
	}

	s.logger.WithFields(map[string]interface{}{
		"order_id": orderID,
		"staff_id": staffID,
	}).Info("Order assigned to staff successfully")

	return nil
}

// StartOrderProcessing starts processing an order
func (s *KitchenServiceImpl) StartOrderProcessing(ctx context.Context, orderID string) error {
	order, err := s.repoManager.Order().GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if !order.IsReadyToStart() {
		return fmt.Errorf("order is not ready to start processing")
	}

	if err := order.UpdateStatus(domain.OrderStatusProcessing); err != nil {
		return fmt.Errorf("failed to start order processing: %w", err)
	}

	if err := s.repoManager.Order().Update(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Publish event
	event := domain.NewOrderStartedEvent(order)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish order started event")
	}

	s.logger.WithField("order_id", orderID).Info("Order processing started")
	return nil
}

// CompleteOrder completes an order
func (s *KitchenServiceImpl) CompleteOrder(ctx context.Context, orderID string) error {
	order, err := s.repoManager.Order().GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status() != domain.OrderStatusProcessing {
		return fmt.Errorf("order is not being processed")
	}

	if err := order.UpdateStatus(domain.OrderStatusCompleted); err != nil {
		return fmt.Errorf("failed to complete order: %w", err)
	}

	if err := s.repoManager.Order().Update(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Release staff assignment
	if order.AssignedStaffID() != "" {
		staff, err := s.repoManager.Staff().GetByID(ctx, order.AssignedStaffID())
		if err == nil {
			staff.CompleteOrder()
			s.repoManager.Staff().Update(ctx, staff)
		}
	}

	// Publish event
	event := domain.NewOrderCompletedEvent(order)
	if err := s.eventService.PublishEvent(ctx, event); err != nil {
		s.logger.WithError(err).Warn("Failed to publish order completed event")
	}

	// Send notification
	if err := s.notificationService.NotifyOrderCompleted(ctx, order); err != nil {
		s.logger.WithError(err).Warn("Failed to send order completed notification")
	}

	s.logger.WithField("order_id", orderID).Info("Order completed successfully")
	return nil
}

// Queue Management

// GetQueueStatus returns the current queue status
func (s *KitchenServiceImpl) GetQueueStatus(ctx context.Context) (*QueueStatusResponse, error) {
	status, err := s.queueService.GetQueueStatus(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get queue status")
		return nil, fmt.Errorf("failed to get queue status: %w", err)
	}

	// Get overdue orders
	overdueOrders, err := s.queueService.GetOverdueOrders(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get overdue orders")
		overdueOrders = []*domain.KitchenOrder{}
	}

	// Get next order
	nextOrder := s.queueService.GetNextOrder(ctx)

	response := &QueueStatusResponse{
		TotalOrders:      status.TotalOrders,
		ProcessingOrders: status.ProcessingOrders,
		PendingOrders:    status.PendingOrders,
		CompletedOrders:  status.CompletedOrders,
		AverageWaitTime:  status.AverageWaitTime,
		QueuesByPriority: status.QueuesByPriority,
		StationLoad:      status.StationLoad,
		UpdatedAt:        status.UpdatedAt,
	}

	// Convert overdue orders to responses
	response.OverdueOrders = make([]*OrderResponse, len(overdueOrders))
	for i, order := range overdueOrders {
		response.OverdueOrders[i] = s.orderToResponse(order)
	}

	// Convert next order to response
	if nextOrder != nil {
		response.NextOrder = s.orderToResponse(nextOrder)
	}

	return response, nil
}

// GetNextOrder returns the next order to be processed
func (s *KitchenServiceImpl) GetNextOrder(ctx context.Context) (*OrderResponse, error) {
	order := s.queueService.GetNextOrder(ctx)
	if order == nil {
		return nil, nil
	}

	return s.orderToResponse(order), nil
}

// OptimizeQueue optimizes the current queue
func (s *KitchenServiceImpl) OptimizeQueue(ctx context.Context) (*OptimizationResponse, error) {
	optimization, err := s.queueService.OptimizeQueue(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to optimize queue")
		return nil, fmt.Errorf("failed to optimize queue: %w", err)
	}

	return s.optimizationToResponse(optimization), nil
}

// Analytics and Metrics

// GetKitchenMetrics returns kitchen performance metrics
func (s *KitchenServiceImpl) GetKitchenMetrics(ctx context.Context, period *TimePeriod) (*MetricsResponse, error) {
	metrics, err := s.repoManager.Metrics().GetLatestMetrics(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get kitchen metrics")
		return nil, fmt.Errorf("failed to get kitchen metrics: %w", err)
	}

	// Get additional stats if period is provided
	var throughputStats *ThroughputStatsResponse
	var completionStats *CompletionStatsResponse

	if period != nil {
		if tStats, err := s.repoManager.Queue().GetThroughputStats(ctx, period.Start, period.End); err == nil {
			throughputStats = &ThroughputStatsResponse{
				OrdersPerHour:      tStats.OrdersPerHour,
				OrdersPerDay:       tStats.OrdersPerDay,
				PeakHourThroughput: tStats.PeakHourThroughput,
				AverageQueueLength: tStats.AverageQueueLength,
				MaxQueueLength:     tStats.MaxQueueLength,
				CalculatedAt:       tStats.CalculatedAt,
			}
		}

		if cStats, err := s.repoManager.Order().GetCompletionStats(ctx, period.Start, period.End); err == nil {
			completionStats = &CompletionStatsResponse{
				TotalOrders:     cStats.TotalOrders,
				CompletedOrders: cStats.CompletedOrders,
				CancelledOrders: cStats.CancelledOrders,
				AverageTime:     cStats.AverageTime,
				MedianTime:      cStats.MedianTime,
				CompletionRate:  cStats.CompletionRate,
				OnTimeRate:      cStats.OnTimeRate,
				CalculatedAt:    cStats.CalculatedAt,
			}
		}
	}

	return &MetricsResponse{
		AvgPreparationTime:   metrics.AvgPreparationTime,
		OrdersCompleted:      metrics.OrdersCompleted,
		OrdersInQueue:        metrics.OrdersInQueue,
		EfficiencyRate:       metrics.EfficiencyRate,
		CustomerSatisfaction: metrics.CustomerSatisfaction,
		StaffUtilization:     metrics.StaffUtilization,
		EquipmentUptime:      metrics.EquipmentUptime,
		ThroughputStats:      throughputStats,
		CompletionStats:      completionStats,
		Timestamp:            metrics.Timestamp,
	}, nil
}

// GetPerformanceReport returns a comprehensive performance report
func (s *KitchenServiceImpl) GetPerformanceReport(ctx context.Context, period *TimePeriod) (*PerformanceReportResponse, error) {
	// This would be implemented with more complex analytics
	// For now, return a basic implementation
	return &PerformanceReportResponse{
		Period:      period,
		Summary:     &PerformanceSummaryResponse{},
		Trends:      make(map[string][]float32),
		Comparisons: make(map[string]float32),
		TopPerformers: &TopPerformersResponse{
			Staff:     []*StaffPerformanceResponse{},
			Equipment: []*EquipmentPerformanceResponse{},
			Stations:  []*StationPerformanceResponse{},
		},
		Bottlenecks:     []*BottleneckResponse{},
		Insights:        []string{},
		Recommendations: []*RecommendationResponse{},
		GeneratedAt:     time.Now(),
	}, nil
}

// Helper methods for converting domain entities to responses

func (s *KitchenServiceImpl) orderToResponse(order *domain.KitchenOrder) *OrderResponse {
	itemResponses := make([]*OrderItemResponse, len(order.Items()))
	for i, item := range order.Items() {
		itemResponses[i] = &OrderItemResponse{
			ID:           item.ID(),
			Name:         item.Name(),
			Quantity:     item.Quantity(),
			Instructions: item.Instructions(),
			Requirements: item.Requirements(),
			Metadata:     item.Metadata(),
		}
	}

	return &OrderResponse{
		ID:                  order.ID(),
		CustomerID:          order.CustomerID(),
		Items:               itemResponses,
		Status:              order.Status(),
		Priority:            order.Priority(),
		EstimatedTime:       order.EstimatedTime(),
		ActualTime:          order.ActualTime(),
		AssignedStaffID:     order.AssignedStaffID(),
		AssignedEquipment:   order.AssignedEquipment(),
		SpecialInstructions: order.SpecialInstructions(),
		RequiredStations:    order.GetRequiredStations(),
		TotalQuantity:       order.GetTotalQuantity(),
		WaitTime:            int64(order.GetWaitTime().Seconds()),
		ProcessingTime:      int64(order.GetProcessingTime().Seconds()),
		IsOverdue:           order.IsOverdue(),
		IsReadyToStart:      order.IsReadyToStart(),
		CreatedAt:           order.CreatedAt(),
		UpdatedAt:           order.UpdatedAt(),
		StartedAt:           order.StartedAt(),
		CompletedAt:         order.CompletedAt(),
	}
}

func (s *KitchenServiceImpl) optimizationToResponse(optimization *domain.WorkflowOptimization) *OptimizationResponse {
	stepResponses := make([]*WorkflowStepResponse, len(optimization.OptimizedSteps))
	for i, step := range optimization.OptimizedSteps {
		stepResponses[i] = &WorkflowStepResponse{
			StepID:         step.StepID,
			StationType:    step.StationType,
			EstimatedTime:  step.EstimatedTime,
			RequiredSkill:  step.RequiredSkill,
			Dependencies:   step.Dependencies,
			CanParallelize: step.CanParallelize,
			EquipmentID:    step.EquipmentID,
			StaffID:        step.StaffID,
		}
	}

	return &OptimizationResponse{
		OrderID:              optimization.OrderID,
		OptimizedSteps:       stepResponses,
		EstimatedTime:        optimization.EstimatedTime,
		EfficiencyGain:       optimization.EfficiencyGain,
		ResourceUtilization:  optimization.ResourceUtilization,
		Recommendations:      optimization.Recommendations,
		StaffAllocations:     []*StaffAllocationResponse{},     // Would be populated from actual allocations
		EquipmentAllocations: []*EquipmentAllocationResponse{}, // Would be populated from actual allocations
		CreatedAt:            optimization.CreatedAt,
	}
}
