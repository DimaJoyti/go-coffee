package application

import (
	"context"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
)

// KitchenService defines the main kitchen service interface
type KitchenService interface {
	// Equipment management
	CreateEquipment(ctx context.Context, req *CreateEquipmentRequest) (*EquipmentResponse, error)
	GetEquipment(ctx context.Context, id string) (*EquipmentResponse, error)
	UpdateEquipmentStatus(ctx context.Context, id string, status domain.EquipmentStatus) error
	ListEquipment(ctx context.Context, filter *EquipmentFilter) ([]*EquipmentResponse, error)
	ScheduleEquipmentMaintenance(ctx context.Context, id string) error

	// Staff management
	CreateStaff(ctx context.Context, req *CreateStaffRequest) (*StaffResponse, error)
	GetStaff(ctx context.Context, id string) (*StaffResponse, error)
	UpdateStaffAvailability(ctx context.Context, id string, available bool) error
	ListStaff(ctx context.Context, filter *StaffFilter) ([]*StaffResponse, error)
	UpdateStaffSkill(ctx context.Context, id string, skillLevel float32) error

	// Order management
	AddOrderToQueue(ctx context.Context, req *AddOrderRequest) (*OrderResponse, error)
	GetOrder(ctx context.Context, id string) (*OrderResponse, error)
	UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) error
	UpdateOrderPriority(ctx context.Context, id string, priority domain.OrderPriority) error
	AssignOrderToStaff(ctx context.Context, orderID, staffID string) error
	StartOrderProcessing(ctx context.Context, orderID string) error
	CompleteOrder(ctx context.Context, orderID string) error

	// Queue management
	GetQueueStatus(ctx context.Context) (*QueueStatusResponse, error)
	GetNextOrder(ctx context.Context) (*OrderResponse, error)
	OptimizeQueue(ctx context.Context) (*OptimizationResponse, error)

	// Analytics and metrics
	GetKitchenMetrics(ctx context.Context, period *TimePeriod) (*MetricsResponse, error)
	GetPerformanceReport(ctx context.Context, period *TimePeriod) (*PerformanceReportResponse, error)
}

// QueueService defines the queue management service interface
type QueueService interface {
	// Queue operations
	AddOrder(ctx context.Context, order *domain.KitchenOrder) error
	RemoveOrder(ctx context.Context, orderID string) error
	GetNextOrder(ctx context.Context) *domain.KitchenOrder
	UpdateOrderPriority(ctx context.Context, orderID string, priority domain.OrderPriority) error

	// Queue status
	GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error)
	GetEstimatedWaitTime(ctx context.Context, order *domain.KitchenOrder) (time.Duration, error)
	GetOverdueOrders(ctx context.Context) ([]*domain.KitchenOrder, error)

	// Queue optimization
	OptimizeQueue(ctx context.Context) (*domain.WorkflowOptimization, error)
	RebalanceQueue(ctx context.Context) error
}

// OptimizerService defines the AI optimization service interface
type OptimizerService interface {
	// Workflow optimization
	OptimizeWorkflow(ctx context.Context, orders []*domain.KitchenOrder) (*domain.WorkflowOptimization, error)
	PredictPreparationTime(ctx context.Context, order *domain.KitchenOrder) (int32, error)

	// Staff allocation
	AllocateStaff(ctx context.Context, orders []*domain.KitchenOrder, staff []*domain.Staff) (*domain.StaffAllocation, error)
	OptimizeStaffSchedule(ctx context.Context, staff []*domain.Staff, timeWindow time.Duration) (*StaffScheduleOptimization, error)

	// Equipment optimization
	OptimizeEquipmentUsage(ctx context.Context, equipment []*domain.Equipment, orders []*domain.KitchenOrder) (*EquipmentOptimization, error)
	PredictEquipmentLoad(ctx context.Context, equipment *domain.Equipment, timeWindow time.Duration) (float32, error)

	// Capacity planning
	PredictCapacity(ctx context.Context, timeWindow time.Duration) (*CapacityPrediction, error)
	AnalyzeBottlenecks(ctx context.Context) (*BottleneckAnalysis, error)

	// Performance optimization
	AnalyzePerformance(ctx context.Context, period *TimePeriod) (*PerformanceAnalysis, error)
	GenerateRecommendations(ctx context.Context, context *OptimizationContext) ([]*Recommendation, error)
}

// NotificationService defines the notification service interface
type NotificationService interface {
	// Order notifications
	NotifyOrderAdded(ctx context.Context, order *domain.KitchenOrder) error
	NotifyOrderStatusChanged(ctx context.Context, order *domain.KitchenOrder, oldStatus domain.OrderStatus) error
	NotifyOrderOverdue(ctx context.Context, order *domain.KitchenOrder) error
	NotifyOrderCompleted(ctx context.Context, order *domain.KitchenOrder) error

	// Staff notifications
	NotifyStaffAssigned(ctx context.Context, staff *domain.Staff, order *domain.KitchenOrder) error
	NotifyStaffOverloaded(ctx context.Context, staff *domain.Staff) error

	// Equipment notifications
	NotifyEquipmentMaintenance(ctx context.Context, equipment *domain.Equipment) error
	NotifyEquipmentOverloaded(ctx context.Context, equipment *domain.Equipment) error

	// Queue notifications
	NotifyQueueBacklog(ctx context.Context, queueStatus *domain.QueueStatus) error
	NotifyCapacityAlert(ctx context.Context, prediction *CapacityPrediction) error
}

// EventService defines the event handling service interface
type EventService interface {
	// Event publishing
	PublishEvent(ctx context.Context, event *domain.DomainEvent) error
	PublishEvents(ctx context.Context, events []*domain.DomainEvent) error

	// Event handling
	HandleOrderEvent(ctx context.Context, event *domain.DomainEvent) error
	HandleEquipmentEvent(ctx context.Context, event *domain.DomainEvent) error
	HandleStaffEvent(ctx context.Context, event *domain.DomainEvent) error
	HandleQueueEvent(ctx context.Context, event *domain.DomainEvent) error

	// Event subscription
	SubscribeToEvents(ctx context.Context, eventTypes []string, handler domain.EventHandler) error
	UnsubscribeFromEvents(ctx context.Context, eventTypes []string) error
}

// External Service Interfaces

// OrderServiceClient defines the interface for communicating with the order service
type OrderServiceClient interface {
	GetOrder(ctx context.Context, orderID string) (*ExternalOrder, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
	NotifyOrderReady(ctx context.Context, orderID string) error
}

// PaymentServiceClient defines the interface for communicating with the payment service
type PaymentServiceClient interface {
	ValidatePayment(ctx context.Context, orderID string) (bool, error)
	ProcessRefund(ctx context.Context, orderID string, amount float64) error
}

// InventoryServiceClient defines the interface for communicating with the inventory service
type InventoryServiceClient interface {
	CheckIngredientAvailability(ctx context.Context, ingredients []string) (map[string]bool, error)
	ReserveIngredients(ctx context.Context, orderID string, ingredients map[string]int32) error
	ReleaseIngredients(ctx context.Context, orderID string) error
}

// AIServiceClient defines the interface for communicating with AI services
type AIServiceClient interface {
	PredictDemand(ctx context.Context, timeWindow time.Duration) (*DemandPrediction, error)
	OptimizeSchedule(ctx context.Context, constraints *ScheduleConstraints) (*ScheduleOptimization, error)
	AnalyzeCustomerBehavior(ctx context.Context, customerID string) (*CustomerBehaviorAnalysis, error)
}

// Supporting Data Structures

// TimePeriod represents a time period for analytics
type TimePeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OptimizationContext provides context for optimization algorithms
type OptimizationContext struct {
	CurrentQueue       []*domain.KitchenOrder `json:"current_queue"`
	AvailableStaff     []*domain.Staff        `json:"available_staff"`
	AvailableEquipment []*domain.Equipment    `json:"available_equipment"`
	TimeWindow         time.Duration          `json:"time_window"`
	Constraints        map[string]interface{} `json:"constraints"`
}

// Recommendation represents an optimization recommendation
type Recommendation struct {
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

// StaffScheduleOptimization represents optimized staff schedule
type StaffScheduleOptimization struct {
	Schedule        map[string]*StaffShift `json:"schedule"`
	EfficiencyGain  float32                `json:"efficiency_gain"`
	CostReduction   float32                `json:"cost_reduction"`
	Recommendations []string               `json:"recommendations"`
	CreatedAt       time.Time              `json:"created_at"`
}

// StaffShift represents a staff shift
type StaffShift struct {
	StaffID   string             `json:"staff_id"`
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`
	Station   domain.StationType `json:"station"`
	Load      float32            `json:"expected_load"`
}

// EquipmentOptimization represents optimized equipment usage
type EquipmentOptimization struct {
	Allocations     map[string]*EquipmentAllocation `json:"allocations"`
	UtilizationGain float32                         `json:"utilization_gain"`
	Recommendations []string                        `json:"recommendations"`
	CreatedAt       time.Time                       `json:"created_at"`
}

// EquipmentAllocation represents equipment allocation
type EquipmentAllocation struct {
	EquipmentID string    `json:"equipment_id"`
	OrderIDs    []string  `json:"order_ids"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Load        float32   `json:"expected_load"`
}

// CapacityPrediction represents capacity prediction
type CapacityPrediction struct {
	TimeWindow       time.Duration `json:"time_window"`
	PredictedOrders  int32         `json:"predicted_orders"`
	CurrentCapacity  int32         `json:"current_capacity"`
	RequiredCapacity int32         `json:"required_capacity"`
	CapacityGap      int32         `json:"capacity_gap"`
	Confidence       float32       `json:"confidence"`
	Recommendations  []string      `json:"recommendations"`
	PredictedAt      time.Time     `json:"predicted_at"`
}

// BottleneckAnalysis represents bottleneck analysis
type BottleneckAnalysis struct {
	Bottlenecks     []*Bottleneck `json:"bottlenecks"`
	OverallImpact   float32       `json:"overall_impact"`
	Recommendations []string      `json:"recommendations"`
	AnalyzedAt      time.Time     `json:"analyzed_at"`
}

// Bottleneck represents a system bottleneck
type Bottleneck struct {
	Type        string   `json:"type"` // "staff", "equipment", "process"
	ResourceID  string   `json:"resource_id"`
	Severity    float32  `json:"severity"` // 0.0 to 1.0
	Impact      string   `json:"impact"`
	Suggestions []string `json:"suggestions"`
}

// PerformanceAnalysis represents performance analysis
type PerformanceAnalysis struct {
	Period          *TimePeriod          `json:"period"`
	Metrics         map[string]float32   `json:"metrics"`
	Trends          map[string][]float32 `json:"trends"`
	Comparisons     map[string]float32   `json:"comparisons"`
	Insights        []string             `json:"insights"`
	Recommendations []*Recommendation    `json:"recommendations"`
	AnalyzedAt      time.Time            `json:"analyzed_at"`
}

// External Data Structures

// ExternalOrder represents an order from the external order service
type ExternalOrder struct {
	ID          string                 `json:"id"`
	CustomerID  string                 `json:"customer_id"`
	Items       []*ExternalOrderItem   `json:"items"`
	Status      string                 `json:"status"`
	TotalAmount float64                `json:"total_amount"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ExternalOrderItem represents an item from the external order service
type ExternalOrderItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity int32   `json:"quantity"`
	Price    float64 `json:"price"`
}

// DemandPrediction represents demand prediction from AI service
type DemandPrediction struct {
	TimeWindow      time.Duration          `json:"time_window"`
	PredictedDemand map[string]int32       `json:"predicted_demand"` // item_name -> quantity
	Confidence      float32                `json:"confidence"`
	Factors         map[string]interface{} `json:"factors"`
	PredictedAt     time.Time              `json:"predicted_at"`
}

// ScheduleOptimization represents schedule optimization from AI service
type ScheduleOptimization struct {
	Schedule        map[string]interface{} `json:"schedule"`
	EfficiencyGain  float32                `json:"efficiency_gain"`
	Recommendations []string               `json:"recommendations"`
	OptimizedAt     time.Time              `json:"optimized_at"`
}

// ScheduleConstraints represents constraints for schedule optimization
type ScheduleConstraints struct {
	StaffAvailability map[string][]time.Time `json:"staff_availability"`
	EquipmentLimits   map[string]int32       `json:"equipment_limits"`
	OrderDeadlines    map[string]time.Time   `json:"order_deadlines"`
	BusinessRules     map[string]interface{} `json:"business_rules"`
}

// CustomerBehaviorAnalysis represents customer behavior analysis
type CustomerBehaviorAnalysis struct {
	CustomerID      string                 `json:"customer_id"`
	OrderPatterns   map[string]interface{} `json:"order_patterns"`
	Preferences     map[string]interface{} `json:"preferences"`
	PredictedOrders []*PredictedOrder      `json:"predicted_orders"`
	AnalyzedAt      time.Time              `json:"analyzed_at"`
}

// PredictedOrder represents a predicted order
type PredictedOrder struct {
	Items       []string      `json:"items"`
	Probability float32       `json:"probability"`
	TimeWindow  time.Duration `json:"time_window"`
}
