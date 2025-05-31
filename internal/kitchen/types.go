package kitchen

import (
	"time"
)

// EquipmentStatus represents the status of kitchen equipment
type EquipmentStatus int32

const (
	EquipmentStatusUnknown     EquipmentStatus = 0
	EquipmentStatusAvailable   EquipmentStatus = 1
	EquipmentStatusInUse       EquipmentStatus = 2
	EquipmentStatusMaintenance EquipmentStatus = 3
	EquipmentStatusBroken      EquipmentStatus = 4
)

// StationType represents different types of kitchen stations
type StationType int32

const (
	StationTypeUnknown  StationType = 0
	StationTypeEspresso StationType = 1
	StationTypeGrinder  StationType = 2
	StationTypeSteamer  StationType = 3
	StationTypeAssembly StationType = 4
)

// OrderStatus represents the status of an order
type OrderStatus int32

const (
	OrderStatusUnknown    OrderStatus = 0
	OrderStatusPending    OrderStatus = 1
	OrderStatusProcessing OrderStatus = 2
	OrderStatusCompleted  OrderStatus = 3
	OrderStatusCancelled  OrderStatus = 4
)

// Equipment represents a piece of kitchen equipment
type Equipment struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	StationType     StationType     `json:"station_type"`
	Status          EquipmentStatus `json:"status"`
	EfficiencyScore float32         `json:"efficiency_score"`
	CurrentLoad     int32           `json:"current_load"`
	MaxCapacity     int32           `json:"max_capacity"`
	LastMaintenance time.Time       `json:"last_maintenance"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// Staff represents a kitchen staff member
type Staff struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Specializations  []StationType `json:"specializations"`
	SkillLevel       float32       `json:"skill_level"`
	IsAvailable      bool          `json:"is_available"`
	CurrentOrders    int32         `json:"current_orders"`
	MaxConcurrentOrders int32      `json:"max_concurrent_orders"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// Order represents a kitchen order
type Order struct {
	ID                string      `json:"id"`
	CustomerID        string      `json:"customer_id"`
	Items             []*OrderItem `json:"items"`
	Status            OrderStatus `json:"status"`
	Priority          int32       `json:"priority"`
	EstimatedTime     int32       `json:"estimated_time"` // in seconds
	ActualTime        int32       `json:"actual_time"`    // in seconds
	AssignedStaffID   string      `json:"assigned_staff_id"`
	AssignedEquipment []string    `json:"assigned_equipment"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	CompletedAt       *time.Time  `json:"completed_at,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Quantity     int32             `json:"quantity"`
	Instructions string            `json:"instructions"`
	Requirements []StationType     `json:"requirements"` // Required stations
	Metadata     map[string]string `json:"metadata"`
}

// WorkflowOptimization represents an optimized workflow
type WorkflowOptimization struct {
	OrderID           string                    `json:"order_id"`
	OptimizedSteps    []*WorkflowStep          `json:"optimized_steps"`
	EstimatedTime     int32                    `json:"estimated_time"`
	EfficiencyGain    float32                  `json:"efficiency_gain"`
	ResourceUtilization map[string]float32     `json:"resource_utilization"`
	Recommendations   []string                 `json:"recommendations"`
}

// WorkflowStep represents a step in the optimized workflow
type WorkflowStep struct {
	StepID          string      `json:"step_id"`
	StationType     StationType `json:"station_type"`
	EstimatedTime   int32       `json:"estimated_time"`
	RequiredSkill   float32     `json:"required_skill"`
	Dependencies    []string    `json:"dependencies"`
	CanParallelize  bool        `json:"can_parallelize"`
}

// StaffAllocation represents staff allocation for orders
type StaffAllocation struct {
	Allocations     []*StaffOrderAllocation `json:"allocations"`
	UtilizationRate float32                 `json:"utilization_rate"`
	LoadBalance     map[string]float32      `json:"load_balance"`
	Recommendations []string                `json:"recommendations"`
}

// StaffOrderAllocation represents allocation of staff to specific orders
type StaffOrderAllocation struct {
	StaffID         string    `json:"staff_id"`
	OrderID         string    `json:"order_id"`
	StationType     StationType `json:"station_type"`
	EstimatedTime   int32     `json:"estimated_time"`
	Priority        int32     `json:"priority"`
}

// QueueStatus represents the current status of the order queue
type QueueStatus struct {
	TotalOrders      int32                    `json:"total_orders"`
	ProcessingOrders int32                    `json:"processing_orders"`
	PendingOrders    int32                    `json:"pending_orders"`
	CompletedOrders  int32                    `json:"completed_orders"`
	AverageWaitTime  int32                    `json:"average_wait_time"`
	QueuesByPriority map[int32]int32          `json:"queues_by_priority"`
	StationLoad      map[StationType]float32  `json:"station_load"`
}

// KitchenMetrics represents kitchen performance metrics
type KitchenMetrics struct {
	AvgPreparationTime   float32            `json:"avg_preparation_time"`
	OrdersCompleted      int32              `json:"orders_completed"`
	OrdersInQueue        int32              `json:"orders_in_queue"`
	EfficiencyRate       float32            `json:"efficiency_rate"`
	CustomerSatisfaction float32            `json:"customer_satisfaction"`
	StaffUtilization     map[string]float32 `json:"staff_utilization"`
	EquipmentUptime      map[string]float32 `json:"equipment_uptime"`
	Timestamp            time.Time          `json:"timestamp"`
}

// OptimizationSuggestion represents an AI-generated optimization suggestion
type OptimizationSuggestion struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Confidence  float32   `json:"confidence"`
	CreatedAt   time.Time `json:"created_at"`
}

// KitchenConfiguration represents kitchen configuration settings
type KitchenConfiguration struct {
	MaxConcurrentOrders int32                      `json:"max_concurrent_orders"`
	DefaultPriority     int32                      `json:"default_priority"`
	StationCapacities   map[StationType]int32      `json:"station_capacities"`
	StaffShifts         map[string]*ShiftSchedule  `json:"staff_shifts"`
	OperatingHours      *OperatingHours            `json:"operating_hours"`
	AIOptimizationEnabled bool                     `json:"ai_optimization_enabled"`
}

// ShiftSchedule represents a staff member's shift schedule
type ShiftSchedule struct {
	StaffID   string    `json:"staff_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	DaysOfWeek []int    `json:"days_of_week"` // 0=Sunday, 1=Monday, etc.
}

// OperatingHours represents kitchen operating hours
type OperatingHours struct {
	OpenTime  string `json:"open_time"`  // "07:00"
	CloseTime string `json:"close_time"` // "22:00"
	Timezone  string `json:"timezone"`   // "UTC"
}
