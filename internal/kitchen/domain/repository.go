package domain

import (
	"context"
	"time"
)

// EquipmentRepository defines the interface for equipment data operations
type EquipmentRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, equipment *Equipment) error
	GetByID(ctx context.Context, id string) (*Equipment, error)
	Update(ctx context.Context, equipment *Equipment) error
	Delete(ctx context.Context, id string) error

	// Query operations
	GetAll(ctx context.Context) ([]*Equipment, error)
	GetByStationType(ctx context.Context, stationType StationType) ([]*Equipment, error)
	GetByStatus(ctx context.Context, status EquipmentStatus) ([]*Equipment, error)
	GetAvailable(ctx context.Context) ([]*Equipment, error)
	GetAvailableByStationType(ctx context.Context, stationType StationType) ([]*Equipment, error)

	// Business operations
	UpdateStatus(ctx context.Context, id string, status EquipmentStatus) error
	UpdateLoad(ctx context.Context, id string, currentLoad int32) error
	UpdateEfficiencyScore(ctx context.Context, id string, score float32) error
	GetNeedingMaintenance(ctx context.Context) ([]*Equipment, error)
	GetOverloaded(ctx context.Context) ([]*Equipment, error)

	// Analytics operations
	GetUtilizationStats(ctx context.Context) (map[string]float32, error)
	GetEfficiencyStats(ctx context.Context) (map[string]float32, error)
}

// StaffRepository defines the interface for staff data operations
type StaffRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, staff *Staff) error
	GetByID(ctx context.Context, id string) (*Staff, error)
	Update(ctx context.Context, staff *Staff) error
	Delete(ctx context.Context, id string) error

	// Query operations
	GetAll(ctx context.Context) ([]*Staff, error)
	GetAvailable(ctx context.Context) ([]*Staff, error)
	GetBySpecialization(ctx context.Context, stationType StationType) ([]*Staff, error)
	GetAvailableBySpecialization(ctx context.Context, stationType StationType) ([]*Staff, error)

	// Business operations
	UpdateAvailability(ctx context.Context, id string, available bool) error
	UpdateCurrentOrders(ctx context.Context, id string, currentOrders int32) error
	UpdateSkillLevel(ctx context.Context, id string, skillLevel float32) error
	GetOverloaded(ctx context.Context) ([]*Staff, error)

	// Analytics operations
	GetWorkloadStats(ctx context.Context) (map[string]float32, error)
	GetSkillStats(ctx context.Context) (map[string]float32, error)
}

// OrderRepository defines the interface for kitchen order data operations
type OrderRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, order *KitchenOrder) error
	GetByID(ctx context.Context, id string) (*KitchenOrder, error)
	Update(ctx context.Context, order *KitchenOrder) error
	Delete(ctx context.Context, id string) error

	// Query operations
	GetAll(ctx context.Context) ([]*KitchenOrder, error)
	GetByStatus(ctx context.Context, status OrderStatus) ([]*KitchenOrder, error)
	GetByPriority(ctx context.Context, priority OrderPriority) ([]*KitchenOrder, error)
	GetByCustomerID(ctx context.Context, customerID string) ([]*KitchenOrder, error)
	GetByStaffID(ctx context.Context, staffID string) ([]*KitchenOrder, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*KitchenOrder, error)

	// Business operations
	UpdateStatus(ctx context.Context, id string, status OrderStatus) error
	UpdatePriority(ctx context.Context, id string, priority OrderPriority) error
	AssignStaff(ctx context.Context, id string, staffID string) error
	AssignEquipment(ctx context.Context, id string, equipmentIDs []string) error
	GetOverdue(ctx context.Context) ([]*KitchenOrder, error)
	GetByRequiredStation(ctx context.Context, stationType StationType) ([]*KitchenOrder, error)

	// Analytics operations
	GetCompletionStats(ctx context.Context, start, end time.Time) (*OrderCompletionStats, error)
	GetAverageProcessingTime(ctx context.Context, start, end time.Time) (float64, error)
	GetOrderCountByStatus(ctx context.Context) (map[OrderStatus]int32, error)
}

// QueueRepository defines the interface for queue data operations
type QueueRepository interface {
	// Queue operations
	SaveQueue(ctx context.Context, queue *OrderQueue) error
	LoadQueue(ctx context.Context) (*OrderQueue, error)
	AddOrderToQueue(ctx context.Context, order *KitchenOrder) error
	RemoveOrderFromQueue(ctx context.Context, orderID string) error
	UpdateOrderInQueue(ctx context.Context, order *KitchenOrder) error

	// Queue status operations
	GetQueueStatus(ctx context.Context) (*QueueStatus, error)
	SaveQueueStatus(ctx context.Context, status *QueueStatus) error

	// Queue analytics
	GetQueueHistory(ctx context.Context, start, end time.Time) ([]*QueueStatus, error)
	GetAverageWaitTime(ctx context.Context, start, end time.Time) (float64, error)
	GetThroughputStats(ctx context.Context, start, end time.Time) (*ThroughputStats, error)
}

// WorkflowRepository defines the interface for workflow optimization data operations
type WorkflowRepository interface {
	// Workflow operations
	SaveOptimization(ctx context.Context, optimization *WorkflowOptimization) error
	GetOptimization(ctx context.Context, orderID string) (*WorkflowOptimization, error)
	GetOptimizationHistory(ctx context.Context, orderID string) ([]*WorkflowOptimization, error)
	DeleteOptimization(ctx context.Context, orderID string) error

	// Analytics operations
	GetOptimizationStats(ctx context.Context, start, end time.Time) (*OptimizationStats, error)
	GetEfficiencyGains(ctx context.Context, start, end time.Time) ([]float32, error)
}

// MetricsRepository defines the interface for kitchen metrics data operations
type MetricsRepository interface {
	// Metrics operations
	SaveMetrics(ctx context.Context, metrics *KitchenMetrics) error
	GetLatestMetrics(ctx context.Context) (*KitchenMetrics, error)
	GetMetricsHistory(ctx context.Context, start, end time.Time) ([]*KitchenMetrics, error)
	GetMetricsByType(ctx context.Context, metricType string, start, end time.Time) ([]float32, error)

	// Aggregated metrics
	GetDailyMetrics(ctx context.Context, date time.Time) (*KitchenMetrics, error)
	GetWeeklyMetrics(ctx context.Context, week time.Time) (*KitchenMetrics, error)
	GetMonthlyMetrics(ctx context.Context, month time.Time) (*KitchenMetrics, error)
}

// Analytics Data Structures

// OrderCompletionStats represents order completion statistics
type OrderCompletionStats struct {
	TotalOrders        int32     `json:"total_orders"`
	CompletedOrders    int32     `json:"completed_orders"`
	CancelledOrders    int32     `json:"cancelled_orders"`
	AverageTime        float64   `json:"average_time"`
	MedianTime         float64   `json:"median_time"`
	CompletionRate     float32   `json:"completion_rate"`
	OnTimeRate         float32   `json:"on_time_rate"`
	CalculatedAt       time.Time `json:"calculated_at"`
}

// ThroughputStats represents queue throughput statistics
type ThroughputStats struct {
	OrdersPerHour      float32   `json:"orders_per_hour"`
	OrdersPerDay       float32   `json:"orders_per_day"`
	PeakHourThroughput float32   `json:"peak_hour_throughput"`
	AverageQueueLength float32   `json:"average_queue_length"`
	MaxQueueLength     int32     `json:"max_queue_length"`
	CalculatedAt       time.Time `json:"calculated_at"`
}

// OptimizationStats represents workflow optimization statistics
type OptimizationStats struct {
	TotalOptimizations   int32     `json:"total_optimizations"`
	AverageEfficiencyGain float32   `json:"average_efficiency_gain"`
	MaxEfficiencyGain    float32   `json:"max_efficiency_gain"`
	TimeSaved            float64   `json:"time_saved_seconds"`
	OptimizationRate     float32   `json:"optimization_rate"`
	CalculatedAt         time.Time `json:"calculated_at"`
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

// Repository Transaction Interface

// UnitOfWork defines the interface for managing transactions across repositories
type UnitOfWork interface {
	// Transaction management
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	// Repository access within transaction
	EquipmentRepo() EquipmentRepository
	StaffRepo() StaffRepository
	OrderRepo() OrderRepository
	QueueRepo() QueueRepository
	WorkflowRepo() WorkflowRepository
	MetricsRepo() MetricsRepository
}

// RepositoryManager defines the interface for managing all repositories
type RepositoryManager interface {
	// Repository access
	Equipment() EquipmentRepository
	Staff() StaffRepository
	Order() OrderRepository
	Queue() QueueRepository
	Workflow() WorkflowRepository
	Metrics() MetricsRepository

	// Transaction management
	NewUnitOfWork() UnitOfWork

	// Health check
	HealthCheck(ctx context.Context) error

	// Close resources
	Close() error
}
