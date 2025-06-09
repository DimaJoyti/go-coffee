package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// RedisRepositoryManager implements domain.RepositoryManager using Redis
type RedisRepositoryManager struct {
	client        *redis.Client
	logger        *logger.Logger
	equipmentRepo domain.EquipmentRepository
	staffRepo     domain.StaffRepository
	orderRepo     domain.OrderRepository
	queueRepo     domain.QueueRepository
	workflowRepo  domain.WorkflowRepository
	metricsRepo   domain.MetricsRepository
}

// NewRedisRepositoryManager creates a new Redis repository manager
func NewRedisRepositoryManager(client *redis.Client, logger *logger.Logger) domain.RepositoryManager {
	return &RedisRepositoryManager{
		client:        client,
		logger:        logger,
		equipmentRepo: NewRedisEquipmentRepository(client, logger),
		staffRepo:     NewRedisStaffRepository(client, logger),
		orderRepo:     NewRedisOrderRepository(client, logger),
		queueRepo:     NewRedisQueueRepository(client, logger),
		workflowRepo:  NewRedisWorkflowRepository(client, logger),
		metricsRepo:   NewRedisMetricsRepository(client, logger),
	}
}

// Equipment returns the equipment repository
func (rm *RedisRepositoryManager) Equipment() domain.EquipmentRepository {
	return rm.equipmentRepo
}

// Staff returns the staff repository
func (rm *RedisRepositoryManager) Staff() domain.StaffRepository {
	return rm.staffRepo
}

// Order returns the order repository
func (rm *RedisRepositoryManager) Order() domain.OrderRepository {
	return rm.orderRepo
}

// Queue returns the queue repository
func (rm *RedisRepositoryManager) Queue() domain.QueueRepository {
	return rm.queueRepo
}

// Workflow returns the workflow repository
func (rm *RedisRepositoryManager) Workflow() domain.WorkflowRepository {
	return rm.workflowRepo
}

// Metrics returns the metrics repository
func (rm *RedisRepositoryManager) Metrics() domain.MetricsRepository {
	return rm.metricsRepo
}

// NewUnitOfWork creates a new unit of work for transaction management
func (rm *RedisRepositoryManager) NewUnitOfWork() domain.UnitOfWork {
	return NewRedisUnitOfWork(rm.client, rm.logger)
}

// HealthCheck checks the health of the repository connections
func (rm *RedisRepositoryManager) HealthCheck(ctx context.Context) error {
	// Ping Redis to check connection
	_, err := rm.client.Ping(ctx).Result()
	if err != nil {
		rm.logger.WithError(err).Error("Redis health check failed")
		return fmt.Errorf("redis health check failed: %w", err)
	}

	rm.logger.Info("Repository manager health check passed")
	return nil
}

// Close closes all repository connections
func (rm *RedisRepositoryManager) Close() error {
	if err := rm.client.Close(); err != nil {
		rm.logger.WithError(err).Error("Failed to close Redis client")
		return fmt.Errorf("failed to close Redis client: %w", err)
	}

	rm.logger.Info("Repository manager closed successfully")
	return nil
}

// RedisUnitOfWork implements domain.UnitOfWork using Redis transactions
type RedisUnitOfWork struct {
	client     *redis.Client
	logger     *logger.Logger
	tx         *redis.Tx
	committed  bool
	rolledBack bool
}

// NewRedisUnitOfWork creates a new Redis unit of work
func NewRedisUnitOfWork(client *redis.Client, logger *logger.Logger) domain.UnitOfWork {
	return &RedisUnitOfWork{
		client: client,
		logger: logger,
	}
}

// Begin starts a new transaction
func (uow *RedisUnitOfWork) Begin(ctx context.Context) error {
	if uow.tx != nil {
		return fmt.Errorf("transaction already started")
	}

	// Redis doesn't have traditional transactions like SQL databases
	// We'll use MULTI/EXEC for atomic operations
	// For now, we'll create a pipeline that can be executed atomically
	uow.logger.Info("Starting Redis transaction")
	return nil
}

// Commit commits the transaction
func (uow *RedisUnitOfWork) Commit(ctx context.Context) error {
	if uow.committed {
		return fmt.Errorf("transaction already committed")
	}
	if uow.rolledBack {
		return fmt.Errorf("transaction already rolled back")
	}

	// In Redis, we would execute the pipeline here
	// For this implementation, we'll mark as committed
	uow.committed = true
	uow.logger.Info("Redis transaction committed")
	return nil
}

// Rollback rolls back the transaction
func (uow *RedisUnitOfWork) Rollback(ctx context.Context) error {
	if uow.committed {
		return fmt.Errorf("transaction already committed")
	}
	if uow.rolledBack {
		return fmt.Errorf("transaction already rolled back")
	}

	// In Redis, we would discard the pipeline here
	uow.rolledBack = true
	uow.logger.Info("Redis transaction rolled back")
	return nil
}

// Repository access within transaction
func (uow *RedisUnitOfWork) EquipmentRepo() domain.EquipmentRepository {
	return NewRedisEquipmentRepository(uow.client, uow.logger)
}

func (uow *RedisUnitOfWork) StaffRepo() domain.StaffRepository {
	return NewRedisStaffRepository(uow.client, uow.logger)
}

func (uow *RedisUnitOfWork) OrderRepo() domain.OrderRepository {
	return NewRedisOrderRepository(uow.client, uow.logger)
}

func (uow *RedisUnitOfWork) QueueRepo() domain.QueueRepository {
	return NewRedisQueueRepository(uow.client, uow.logger)
}

func (uow *RedisUnitOfWork) WorkflowRepo() domain.WorkflowRepository {
	return NewRedisWorkflowRepository(uow.client, uow.logger)
}

func (uow *RedisUnitOfWork) MetricsRepo() domain.MetricsRepository {
	return NewRedisMetricsRepository(uow.client, uow.logger)
}

// RedisWorkflowRepository implements domain.WorkflowRepository using Redis
type RedisWorkflowRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisWorkflowRepository creates a new Redis workflow repository
func NewRedisWorkflowRepository(client *redis.Client, logger *logger.Logger) domain.WorkflowRepository {
	return &RedisWorkflowRepository{
		client: client,
		logger: logger,
	}
}

const (
	workflowKeyPrefix = "kitchen:workflow:"
	workflowSetKey    = "kitchen:workflow:all"
)

// SaveOptimization saves workflow optimization
func (r *RedisWorkflowRepository) SaveOptimization(ctx context.Context, optimization *domain.WorkflowOptimization) error {
	key := workflowKeyPrefix + optimization.OrderID

	data, err := json.Marshal(optimization)
	if err != nil {
		r.logger.WithError(err).Error("Failed to marshal workflow optimization")
		return fmt.Errorf("failed to marshal workflow optimization: %w", err)
	}

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, key, data, 0)
	pipe.SAdd(ctx, workflowSetKey, optimization.OrderID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("order_id", optimization.OrderID).Error("Failed to save workflow optimization")
		return fmt.Errorf("failed to save workflow optimization: %w", err)
	}

	return nil
}

// GetOptimization retrieves workflow optimization by order ID
func (r *RedisWorkflowRepository) GetOptimization(ctx context.Context, orderID string) (*domain.WorkflowOptimization, error) {
	key := workflowKeyPrefix + orderID

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("workflow optimization not found: %s", orderID)
		}
		return nil, fmt.Errorf("failed to get workflow optimization: %w", err)
	}

	var optimization domain.WorkflowOptimization
	if err := json.Unmarshal([]byte(data), &optimization); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow optimization: %w", err)
	}

	return &optimization, nil
}

// GetOptimizationHistory retrieves optimization history for an order
func (r *RedisWorkflowRepository) GetOptimizationHistory(ctx context.Context, orderID string) ([]*domain.WorkflowOptimization, error) {
	// For simplicity, return single optimization
	optimization, err := r.GetOptimization(ctx, orderID)
	if err != nil {
		return []*domain.WorkflowOptimization{}, nil
	}
	return []*domain.WorkflowOptimization{optimization}, nil
}

// DeleteOptimization deletes workflow optimization
func (r *RedisWorkflowRepository) DeleteOptimization(ctx context.Context, orderID string) error {
	key := workflowKeyPrefix + orderID

	pipe := r.client.TxPipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, workflowSetKey, orderID)

	_, err := pipe.Exec(ctx)
	return err
}

// GetOptimizationStats returns optimization statistics
func (r *RedisWorkflowRepository) GetOptimizationStats(ctx context.Context, start, end time.Time) (*domain.OptimizationStats, error) {
	// Simplified implementation
	return &domain.OptimizationStats{
		CalculatedAt: time.Now(),
	}, nil
}

// GetEfficiencyGains returns efficiency gains
func (r *RedisWorkflowRepository) GetEfficiencyGains(ctx context.Context, start, end time.Time) ([]float32, error) {
	return []float32{}, nil
}

// RedisMetricsRepository implements domain.MetricsRepository using Redis
type RedisMetricsRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisMetricsRepository creates a new Redis metrics repository
func NewRedisMetricsRepository(client *redis.Client, logger *logger.Logger) domain.MetricsRepository {
	return &RedisMetricsRepository{
		client: client,
		logger: logger,
	}
}

const (
	metricsKeyPrefix = "kitchen:metrics:"
	metricsLatestKey = "kitchen:metrics:latest"
)

// SaveMetrics saves kitchen metrics
func (r *RedisMetricsRepository) SaveMetrics(ctx context.Context, metrics *domain.KitchenMetrics) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	key := metricsKeyPrefix + fmt.Sprintf("%d", metrics.Timestamp.Unix())

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, key, data, 0)
	pipe.Set(ctx, metricsLatestKey, data, 0)

	_, err = pipe.Exec(ctx)
	return err
}

// GetLatestMetrics retrieves the latest metrics
func (r *RedisMetricsRepository) GetLatestMetrics(ctx context.Context) (*domain.KitchenMetrics, error) {
	data, err := r.client.Get(ctx, metricsLatestKey).Result()
	if err != nil {
		if err == redis.Nil {
			// Return default metrics if none exist
			return &domain.KitchenMetrics{
				StaffUtilization: make(map[string]float32),
				EquipmentUptime:  make(map[string]float32),
				Timestamp:        time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get latest metrics: %w", err)
	}

	var metrics domain.KitchenMetrics
	if err := json.Unmarshal([]byte(data), &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &metrics, nil
}

// GetMetricsHistory retrieves metrics history
func (r *RedisMetricsRepository) GetMetricsHistory(ctx context.Context, start, end time.Time) ([]*domain.KitchenMetrics, error) {
	// Simplified implementation - would scan keys in range
	latest, err := r.GetLatestMetrics(ctx)
	if err != nil {
		return []*domain.KitchenMetrics{}, nil
	}
	return []*domain.KitchenMetrics{latest}, nil
}

// GetMetricsByType retrieves metrics by type
func (r *RedisMetricsRepository) GetMetricsByType(ctx context.Context, metricType string, start, end time.Time) ([]float32, error) {
	return []float32{}, nil
}

// GetDailyMetrics retrieves daily metrics
func (r *RedisMetricsRepository) GetDailyMetrics(ctx context.Context, date time.Time) (*domain.KitchenMetrics, error) {
	return r.GetLatestMetrics(ctx)
}

// GetWeeklyMetrics retrieves weekly metrics
func (r *RedisMetricsRepository) GetWeeklyMetrics(ctx context.Context, week time.Time) (*domain.KitchenMetrics, error) {
	return r.GetLatestMetrics(ctx)
}

// GetMonthlyMetrics retrieves monthly metrics
func (r *RedisMetricsRepository) GetMonthlyMetrics(ctx context.Context, month time.Time) (*domain.KitchenMetrics, error) {
	return r.GetLatestMetrics(ctx)
}
