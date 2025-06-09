package domain

import (
	"errors"
	"time"
)

// OrderStatus represents the status of an order in the kitchen
type OrderStatus int32

const (
	OrderStatusUnknown    OrderStatus = 0
	OrderStatusPending    OrderStatus = 1
	OrderStatusProcessing OrderStatus = 2
	OrderStatusCompleted  OrderStatus = 3
	OrderStatusCancelled  OrderStatus = 4
)

// OrderPriority represents the priority level of an order
type OrderPriority int32

const (
	OrderPriorityLow    OrderPriority = 1
	OrderPriorityNormal OrderPriority = 2
	OrderPriorityHigh   OrderPriority = 3
	OrderPriorityUrgent OrderPriority = 4
)

// KitchenOrder represents an order in the kitchen system (Domain Entity)
type KitchenOrder struct {
	id                  string
	customerID          string
	items               []*OrderItem
	status              OrderStatus
	priority            OrderPriority
	estimatedTime       int32 // in seconds
	actualTime          int32 // in seconds
	assignedStaffID     string
	assignedEquipment   []string
	specialInstructions string
	createdAt           time.Time
	updatedAt           time.Time
	startedAt           *time.Time
	completedAt         *time.Time
}

// OrderItem represents an item in a kitchen order
type OrderItem struct {
	id           string
	name         string
	quantity     int32
	instructions string
	requirements []StationType // Required stations
	metadata     map[string]string
}

// NewKitchenOrder creates a new kitchen order with validation
func NewKitchenOrder(id, customerID string, items []*OrderItem) (*KitchenOrder, error) {
	if id == "" {
		return nil, errors.New("order ID is required")
	}
	if customerID == "" {
		return nil, errors.New("customer ID is required")
	}
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	// Validate items
	for _, item := range items {
		if err := validateOrderItem(item); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	return &KitchenOrder{
		id:                  id,
		customerID:          customerID,
		items:               items,
		status:              OrderStatusPending,
		priority:            OrderPriorityNormal,
		estimatedTime:       0,
		actualTime:          0,
		assignedStaffID:     "",
		assignedEquipment:   []string{},
		specialInstructions: "",
		createdAt:           now,
		updatedAt:           now,
	}, nil
}

// validateOrderItem validates an order item
func validateOrderItem(item *OrderItem) error {
	if item.id == "" {
		return errors.New("item ID is required")
	}
	if item.name == "" {
		return errors.New("item name is required")
	}
	if item.quantity <= 0 {
		return errors.New("item quantity must be greater than 0")
	}
	return nil
}

// NewOrderItem creates a new order item
func NewOrderItem(id, name string, quantity int32, requirements []StationType) *OrderItem {
	return &OrderItem{
		id:           id,
		name:         name,
		quantity:     quantity,
		requirements: requirements,
		metadata:     make(map[string]string),
	}
}

// Getters for OrderItem
func (oi *OrderItem) ID() string                  { return oi.id }
func (oi *OrderItem) Name() string                { return oi.name }
func (oi *OrderItem) Quantity() int32             { return oi.quantity }
func (oi *OrderItem) Instructions() string        { return oi.instructions }
func (oi *OrderItem) Requirements() []StationType { return oi.requirements }
func (oi *OrderItem) Metadata() map[string]string { return oi.metadata }

// Setters for OrderItem
func (oi *OrderItem) SetInstructions(instructions string) {
	oi.instructions = instructions
}

func (oi *OrderItem) SetMetadata(metadata map[string]string) {
	oi.metadata = metadata
}

// Getters
func (o *KitchenOrder) ID() string                  { return o.id }
func (o *KitchenOrder) CustomerID() string          { return o.customerID }
func (o *KitchenOrder) Items() []*OrderItem         { return o.items }
func (o *KitchenOrder) Status() OrderStatus         { return o.status }
func (o *KitchenOrder) Priority() OrderPriority     { return o.priority }
func (o *KitchenOrder) EstimatedTime() int32        { return o.estimatedTime }
func (o *KitchenOrder) ActualTime() int32           { return o.actualTime }
func (o *KitchenOrder) AssignedStaffID() string     { return o.assignedStaffID }
func (o *KitchenOrder) AssignedEquipment() []string { return o.assignedEquipment }
func (o *KitchenOrder) SpecialInstructions() string { return o.specialInstructions }
func (o *KitchenOrder) CreatedAt() time.Time        { return o.createdAt }
func (o *KitchenOrder) UpdatedAt() time.Time        { return o.updatedAt }
func (o *KitchenOrder) StartedAt() *time.Time       { return o.startedAt }
func (o *KitchenOrder) CompletedAt() *time.Time     { return o.completedAt }

// Business Methods

// UpdateStatus changes the order status with validation
func (o *KitchenOrder) UpdateStatus(status OrderStatus) error {
	// Business rules for status transitions
	switch o.status {
	case OrderStatusPending:
		if status != OrderStatusProcessing && status != OrderStatusCancelled {
			return errors.New("pending order can only be moved to processing or cancelled")
		}
	case OrderStatusProcessing:
		if status != OrderStatusCompleted && status != OrderStatusCancelled {
			return errors.New("processing order can only be moved to completed or cancelled")
		}
	case OrderStatusCompleted, OrderStatusCancelled:
		return errors.New("completed or cancelled orders cannot change status")
	}

	o.status = status
	o.updatedAt = time.Now()

	// Set timestamps based on status
	now := time.Now()
	switch status {
	case OrderStatusProcessing:
		if o.startedAt == nil {
			o.startedAt = &now
		}
	case OrderStatusCompleted:
		if o.completedAt == nil {
			o.completedAt = &now
			// Calculate actual time if started
			if o.startedAt != nil {
				o.actualTime = int32(now.Sub(*o.startedAt).Seconds())
			}
		}
	}

	return nil
}

// SetPriority sets the order priority
func (o *KitchenOrder) SetPriority(priority OrderPriority) {
	o.priority = priority
	o.updatedAt = time.Now()
}

// SetEstimatedTime sets the estimated preparation time
func (o *KitchenOrder) SetEstimatedTime(seconds int32) error {
	if seconds < 0 {
		return errors.New("estimated time cannot be negative")
	}
	o.estimatedTime = seconds
	o.updatedAt = time.Now()
	return nil
}

// AssignStaff assigns a staff member to the order
func (o *KitchenOrder) AssignStaff(staffID string) error {
	if staffID == "" {
		return errors.New("staff ID cannot be empty")
	}
	if o.status != OrderStatusPending {
		return errors.New("can only assign staff to pending orders")
	}

	o.assignedStaffID = staffID
	o.updatedAt = time.Now()
	return nil
}

// AssignEquipment assigns equipment to the order
func (o *KitchenOrder) AssignEquipment(equipmentIDs []string) error {
	if len(equipmentIDs) == 0 {
		return errors.New("at least one equipment must be assigned")
	}
	if o.status != OrderStatusPending {
		return errors.New("can only assign equipment to pending orders")
	}

	o.assignedEquipment = equipmentIDs
	o.updatedAt = time.Now()
	return nil
}

// SetSpecialInstructions sets special instructions for the order
func (o *KitchenOrder) SetSpecialInstructions(instructions string) {
	o.specialInstructions = instructions
	o.updatedAt = time.Now()
}

// GetRequiredStations returns all required station types for this order
func (o *KitchenOrder) GetRequiredStations() []StationType {
	stationMap := make(map[StationType]bool)

	for _, item := range o.items {
		for _, requirement := range item.requirements {
			stationMap[requirement] = true
		}
	}

	stations := make([]StationType, 0, len(stationMap))
	for station := range stationMap {
		stations = append(stations, station)
	}

	return stations
}

// GetTotalQuantity returns the total quantity of items in the order
func (o *KitchenOrder) GetTotalQuantity() int32 {
	total := int32(0)
	for _, item := range o.items {
		total += item.quantity
	}
	return total
}

// IsReadyToStart checks if the order is ready to start processing
func (o *KitchenOrder) IsReadyToStart() bool {
	return o.status == OrderStatusPending &&
		o.assignedStaffID != "" &&
		len(o.assignedEquipment) > 0
}

// GetWaitTime returns how long the order has been waiting
func (o *KitchenOrder) GetWaitTime() time.Duration {
	if o.startedAt != nil {
		return o.startedAt.Sub(o.createdAt)
	}
	return time.Since(o.createdAt)
}

// GetProcessingTime returns how long the order has been processing
func (o *KitchenOrder) GetProcessingTime() time.Duration {
	if o.startedAt == nil {
		return 0
	}
	if o.completedAt != nil {
		return o.completedAt.Sub(*o.startedAt)
	}
	return time.Since(*o.startedAt)
}

// IsOverdue checks if the order is taking longer than estimated
func (o *KitchenOrder) IsOverdue() bool {
	if o.estimatedTime == 0 || o.startedAt == nil {
		return false
	}

	expectedCompletion := o.startedAt.Add(time.Duration(o.estimatedTime) * time.Second)
	return time.Now().After(expectedCompletion) && o.status != OrderStatusCompleted
}

// ToDTO converts domain entity to data transfer object
func (o *KitchenOrder) ToDTO() *KitchenOrderDTO {
	itemDTOs := make([]*OrderItemDTO, len(o.items))
	for i, item := range o.items {
		itemDTOs[i] = item.ToDTO()
	}

	return &KitchenOrderDTO{
		ID:                  o.id,
		CustomerID:          o.customerID,
		Items:               itemDTOs,
		Status:              o.status,
		Priority:            o.priority,
		EstimatedTime:       o.estimatedTime,
		ActualTime:          o.actualTime,
		AssignedStaffID:     o.assignedStaffID,
		AssignedEquipment:   o.assignedEquipment,
		SpecialInstructions: o.specialInstructions,
		CreatedAt:           o.createdAt,
		UpdatedAt:           o.updatedAt,
		StartedAt:           o.startedAt,
		CompletedAt:         o.completedAt,
	}
}

// ToDTO converts order item to DTO
func (oi *OrderItem) ToDTO() *OrderItemDTO {
	return &OrderItemDTO{
		ID:           oi.id,
		Name:         oi.name,
		Quantity:     oi.quantity,
		Instructions: oi.instructions,
		Requirements: oi.requirements,
		Metadata:     oi.metadata,
	}
}

// DTOs
type KitchenOrderDTO struct {
	ID                  string          `json:"id"`
	CustomerID          string          `json:"customer_id"`
	Items               []*OrderItemDTO `json:"items"`
	Status              OrderStatus     `json:"status"`
	Priority            OrderPriority   `json:"priority"`
	EstimatedTime       int32           `json:"estimated_time"`
	ActualTime          int32           `json:"actual_time"`
	AssignedStaffID     string          `json:"assigned_staff_id"`
	AssignedEquipment   []string        `json:"assigned_equipment"`
	SpecialInstructions string          `json:"special_instructions"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	StartedAt           *time.Time      `json:"started_at,omitempty"`
	CompletedAt         *time.Time      `json:"completed_at,omitempty"`
}

type OrderItemDTO struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Quantity     int32             `json:"quantity"`
	Instructions string            `json:"instructions"`
	Requirements []StationType     `json:"requirements"`
	Metadata     map[string]string `json:"metadata"`
}
