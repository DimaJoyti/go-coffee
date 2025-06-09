package application

import (
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
)

// Request DTOs

// CreateEquipmentRequest represents a request to create equipment
type CreateEquipmentRequest struct {
	ID          string                `json:"id" validate:"required"`
	Name        string                `json:"name" validate:"required"`
	StationType domain.StationType    `json:"station_type" validate:"required"`
	MaxCapacity int32                 `json:"max_capacity" validate:"required,min=1"`
}

// CreateStaffRequest represents a request to create staff
type CreateStaffRequest struct {
	ID                  string                `json:"id" validate:"required"`
	Name                string                `json:"name" validate:"required"`
	Specializations     []domain.StationType  `json:"specializations" validate:"required,min=1"`
	SkillLevel          float32               `json:"skill_level" validate:"required,min=0,max=10"`
	MaxConcurrentOrders int32                 `json:"max_concurrent_orders" validate:"required,min=1"`
}

// AddOrderRequest represents a request to add an order to the kitchen queue
type AddOrderRequest struct {
	ID                  string                `json:"id" validate:"required"`
	CustomerID          string                `json:"customer_id" validate:"required"`
	Items               []*OrderItemRequest   `json:"items" validate:"required,min=1"`
	Priority            domain.OrderPriority  `json:"priority"`
	SpecialInstructions string                `json:"special_instructions"`
}

// OrderItemRequest represents an order item in a request
type OrderItemRequest struct {
	ID           string                `json:"id" validate:"required"`
	Name         string                `json:"name" validate:"required"`
	Quantity     int32                 `json:"quantity" validate:"required,min=1"`
	Instructions string                `json:"instructions"`
	Requirements []domain.StationType  `json:"requirements" validate:"required,min=1"`
	Metadata     map[string]string     `json:"metadata"`
}

// UpdateOrderStatusRequest represents a request to update order status
type UpdateOrderStatusRequest struct {
	OrderID string              `json:"order_id" validate:"required"`
	Status  domain.OrderStatus  `json:"status" validate:"required"`
}

// UpdateOrderPriorityRequest represents a request to update order priority
type UpdateOrderPriorityRequest struct {
	OrderID  string                `json:"order_id" validate:"required"`
	Priority domain.OrderPriority  `json:"priority" validate:"required"`
}

// AssignOrderRequest represents a request to assign an order
type AssignOrderRequest struct {
	OrderID     string   `json:"order_id" validate:"required"`
	StaffID     string   `json:"staff_id" validate:"required"`
	EquipmentIDs []string `json:"equipment_ids" validate:"required,min=1"`
}

// Response DTOs

// EquipmentResponse represents equipment data in responses
type EquipmentResponse struct {
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	StationType     domain.StationType      `json:"station_type"`
	Status          domain.EquipmentStatus  `json:"status"`
	EfficiencyScore float32                 `json:"efficiency_score"`
	CurrentLoad     int32                   `json:"current_load"`
	MaxCapacity     int32                   `json:"max_capacity"`
	UtilizationRate float32                 `json:"utilization_rate"`
	IsAvailable     bool                    `json:"is_available"`
	NeedsMaintenance bool                   `json:"needs_maintenance"`
	LastMaintenance time.Time               `json:"last_maintenance"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// StaffResponse represents staff data in responses
type StaffResponse struct {
	ID                  string                `json:"id"`
	Name                string                `json:"name"`
	Specializations     []domain.StationType  `json:"specializations"`
	SkillLevel          float32               `json:"skill_level"`
	IsAvailable         bool                  `json:"is_available"`
	CurrentOrders       int32                 `json:"current_orders"`
	MaxConcurrentOrders int32                 `json:"max_concurrent_orders"`
	Workload            float32               `json:"workload"`
	IsOverloaded        bool                  `json:"is_overloaded"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
}

// OrderResponse represents order data in responses
type OrderResponse struct {
	ID                  string                  `json:"id"`
	CustomerID          string                  `json:"customer_id"`
	Items               []*OrderItemResponse    `json:"items"`
	Status              domain.OrderStatus      `json:"status"`
	Priority            domain.OrderPriority    `json:"priority"`
	EstimatedTime       int32                   `json:"estimated_time"`
	ActualTime          int32                   `json:"actual_time"`
	AssignedStaffID     string                  `json:"assigned_staff_id"`
	AssignedEquipment   []string                `json:"assigned_equipment"`
	SpecialInstructions string                  `json:"special_instructions"`
	RequiredStations    []domain.StationType    `json:"required_stations"`
	TotalQuantity       int32                   `json:"total_quantity"`
	WaitTime            int64                   `json:"wait_time_seconds"`
	ProcessingTime      int64                   `json:"processing_time_seconds"`
	IsOverdue           bool                    `json:"is_overdue"`
	IsReadyToStart      bool                    `json:"is_ready_to_start"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
	StartedAt           *time.Time              `json:"started_at,omitempty"`
	CompletedAt         *time.Time              `json:"completed_at,omitempty"`
}

// OrderItemResponse represents order item data in responses
type OrderItemResponse struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Quantity     int32                 `json:"quantity"`
	Instructions string                `json:"instructions"`
	Requirements []domain.StationType  `json:"requirements"`
	Metadata     map[string]string     `json:"metadata"`
}

// QueueStatusResponse represents queue status in responses
type QueueStatusResponse struct {
	TotalOrders      int32                           `json:"total_orders"`
	ProcessingOrders int32                           `json:"processing_orders"`
	PendingOrders    int32                           `json:"pending_orders"`
	CompletedOrders  int32                           `json:"completed_orders"`
	AverageWaitTime  int32                           `json:"average_wait_time"`
	QueuesByPriority map[domain.OrderPriority]int32  `json:"queues_by_priority"`
	StationLoad      map[domain.StationType]float32  `json:"station_load"`
	OverdueOrders    []*OrderResponse                `json:"overdue_orders"`
	NextOrder        *OrderResponse                  `json:"next_order,omitempty"`
	UpdatedAt        time.Time                       `json:"updated_at"`
}

// OptimizationResponse represents optimization results in responses
type OptimizationResponse struct {
	OrderID             string                          `json:"order_id"`
	OptimizedSteps      []*WorkflowStepResponse         `json:"optimized_steps"`
	EstimatedTime       int32                           `json:"estimated_time"`
	EfficiencyGain      float32                         `json:"efficiency_gain"`
	ResourceUtilization map[string]float32              `json:"resource_utilization"`
	Recommendations     []string                        `json:"recommendations"`
	StaffAllocations    []*StaffAllocationResponse      `json:"staff_allocations"`
	EquipmentAllocations []*EquipmentAllocationResponse `json:"equipment_allocations"`
	CreatedAt           time.Time                       `json:"created_at"`
}

// WorkflowStepResponse represents workflow step in responses
type WorkflowStepResponse struct {
	StepID          string              `json:"step_id"`
	StationType     domain.StationType  `json:"station_type"`
	EstimatedTime   int32               `json:"estimated_time"`
	RequiredSkill   float32             `json:"required_skill"`
	Dependencies    []string            `json:"dependencies"`
	CanParallelize  bool                `json:"can_parallelize"`
	EquipmentID     string              `json:"equipment_id,omitempty"`
	StaffID         string              `json:"staff_id,omitempty"`
}

// StaffAllocationResponse represents staff allocation in responses
type StaffAllocationResponse struct {
	StaffID          string              `json:"staff_id"`
	OrderID          string              `json:"order_id"`
	StationType      domain.StationType  `json:"station_type"`
	EstimatedTime    int32               `json:"estimated_time"`
	EfficiencyScore  float32             `json:"efficiency_score"`
	AllocationReason string              `json:"allocation_reason"`
	AllocatedAt      time.Time           `json:"allocated_at"`
}

// EquipmentAllocationResponse represents equipment allocation in responses
type EquipmentAllocationResponse struct {
	EquipmentID string    `json:"equipment_id"`
	OrderIDs    []string  `json:"order_ids"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Load        float32   `json:"expected_load"`
}

// MetricsResponse represents kitchen metrics in responses
type MetricsResponse struct {
	AvgPreparationTime   float32                         `json:"avg_preparation_time"`
	OrdersCompleted      int32                           `json:"orders_completed"`
	OrdersInQueue        int32                           `json:"orders_in_queue"`
	EfficiencyRate       float32                         `json:"efficiency_rate"`
	CustomerSatisfaction float32                         `json:"customer_satisfaction"`
	StaffUtilization     map[string]float32              `json:"staff_utilization"`
	EquipmentUptime      map[string]float32              `json:"equipment_uptime"`
	ThroughputStats      *ThroughputStatsResponse        `json:"throughput_stats"`
	CompletionStats      *CompletionStatsResponse        `json:"completion_stats"`
	Timestamp            time.Time                       `json:"timestamp"`
}

// ThroughputStatsResponse represents throughput statistics in responses
type ThroughputStatsResponse struct {
	OrdersPerHour      float32   `json:"orders_per_hour"`
	OrdersPerDay       float32   `json:"orders_per_day"`
	PeakHourThroughput float32   `json:"peak_hour_throughput"`
	AverageQueueLength float32   `json:"average_queue_length"`
	MaxQueueLength     int32     `json:"max_queue_length"`
	CalculatedAt       time.Time `json:"calculated_at"`
}

// CompletionStatsResponse represents completion statistics in responses
type CompletionStatsResponse struct {
	TotalOrders     int32     `json:"total_orders"`
	CompletedOrders int32     `json:"completed_orders"`
	CancelledOrders int32     `json:"cancelled_orders"`
	AverageTime     float64   `json:"average_time"`
	MedianTime      float64   `json:"median_time"`
	CompletionRate  float32   `json:"completion_rate"`
	OnTimeRate      float32   `json:"on_time_rate"`
	CalculatedAt    time.Time `json:"calculated_at"`
}

// PerformanceReportResponse represents performance report in responses
type PerformanceReportResponse struct {
	Period          *TimePeriod                     `json:"period"`
	Summary         *PerformanceSummaryResponse     `json:"summary"`
	Trends          map[string][]float32            `json:"trends"`
	Comparisons     map[string]float32              `json:"comparisons"`
	TopPerformers   *TopPerformersResponse          `json:"top_performers"`
	Bottlenecks     []*BottleneckResponse           `json:"bottlenecks"`
	Insights        []string                        `json:"insights"`
	Recommendations []*RecommendationResponse       `json:"recommendations"`
	GeneratedAt     time.Time                       `json:"generated_at"`
}

// PerformanceSummaryResponse represents performance summary in responses
type PerformanceSummaryResponse struct {
	TotalOrders         int32   `json:"total_orders"`
	CompletedOrders     int32   `json:"completed_orders"`
	AverageTime         float64 `json:"average_time"`
	EfficiencyRate      float32 `json:"efficiency_rate"`
	CustomerSatisfaction float32 `json:"customer_satisfaction"`
	StaffUtilization    float32 `json:"staff_utilization"`
	EquipmentUtilization float32 `json:"equipment_utilization"`
}

// TopPerformersResponse represents top performers in responses
type TopPerformersResponse struct {
	Staff     []*StaffPerformanceResponse     `json:"staff"`
	Equipment []*EquipmentPerformanceResponse `json:"equipment"`
	Stations  []*StationPerformanceResponse   `json:"stations"`
}

// StaffPerformanceResponse represents staff performance in responses
type StaffPerformanceResponse struct {
	StaffID         string  `json:"staff_id"`
	Name            string  `json:"name"`
	OrdersCompleted int32   `json:"orders_completed"`
	AverageTime     float64 `json:"average_time"`
	EfficiencyScore float32 `json:"efficiency_score"`
	Rank            int32   `json:"rank"`
}

// EquipmentPerformanceResponse represents equipment performance in responses
type EquipmentPerformanceResponse struct {
	EquipmentID     string  `json:"equipment_id"`
	Name            string  `json:"name"`
	UtilizationRate float32 `json:"utilization_rate"`
	EfficiencyScore float32 `json:"efficiency_score"`
	UptimeRate      float32 `json:"uptime_rate"`
	Rank            int32   `json:"rank"`
}

// StationPerformanceResponse represents station performance in responses
type StationPerformanceResponse struct {
	StationType     domain.StationType `json:"station_type"`
	OrdersProcessed int32              `json:"orders_processed"`
	AverageTime     float64            `json:"average_time"`
	UtilizationRate float32            `json:"utilization_rate"`
	Rank            int32              `json:"rank"`
}

// BottleneckResponse represents bottleneck in responses
type BottleneckResponse struct {
	Type        string   `json:"type"`
	ResourceID  string   `json:"resource_id"`
	Severity    float32  `json:"severity"`
	Impact      string   `json:"impact"`
	Suggestions []string `json:"suggestions"`
}

// RecommendationResponse represents recommendation in responses
type RecommendationResponse struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    int32                  `json:"priority"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`
	Effort      string                 `json:"effort"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Filter DTOs

// EquipmentFilter represents equipment filtering options
type EquipmentFilter struct {
	StationType *domain.StationType    `json:"station_type,omitempty"`
	Status      *domain.EquipmentStatus `json:"status,omitempty"`
	Available   *bool                   `json:"available,omitempty"`
	Limit       int32                   `json:"limit,omitempty"`
	Offset      int32                   `json:"offset,omitempty"`
}

// StaffFilter represents staff filtering options
type StaffFilter struct {
	Specialization *domain.StationType `json:"specialization,omitempty"`
	Available      *bool               `json:"available,omitempty"`
	MinSkillLevel  *float32            `json:"min_skill_level,omitempty"`
	Limit          int32               `json:"limit,omitempty"`
	Offset         int32               `json:"offset,omitempty"`
}

// OrderFilter represents order filtering options
type OrderFilter struct {
	Status     *domain.OrderStatus   `json:"status,omitempty"`
	Priority   *domain.OrderPriority `json:"priority,omitempty"`
	CustomerID *string               `json:"customer_id,omitempty"`
	StaffID    *string               `json:"staff_id,omitempty"`
	StartDate  *time.Time            `json:"start_date,omitempty"`
	EndDate    *time.Time            `json:"end_date,omitempty"`
	Limit      int32                 `json:"limit,omitempty"`
	Offset     int32                 `json:"offset,omitempty"`
}
