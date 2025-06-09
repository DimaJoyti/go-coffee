package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Mock implementations for testing

type MockRepositoryManager struct {
	mock.Mock
}

func (m *MockRepositoryManager) Equipment() domain.EquipmentRepository {
	args := m.Called()
	return args.Get(0).(domain.EquipmentRepository)
}

func (m *MockRepositoryManager) Staff() domain.StaffRepository {
	args := m.Called()
	return args.Get(0).(domain.StaffRepository)
}

func (m *MockRepositoryManager) Order() domain.OrderRepository {
	args := m.Called()
	return args.Get(0).(domain.OrderRepository)
}

func (m *MockRepositoryManager) Queue() domain.QueueRepository {
	args := m.Called()
	return args.Get(0).(domain.QueueRepository)
}

func (m *MockRepositoryManager) Workflow() domain.WorkflowRepository {
	args := m.Called()
	return args.Get(0).(domain.WorkflowRepository)
}

func (m *MockRepositoryManager) Metrics() domain.MetricsRepository {
	args := m.Called()
	return args.Get(0).(domain.MetricsRepository)
}

func (m *MockRepositoryManager) BeginTransaction(ctx context.Context) (domain.RepositoryManager, error) {
	args := m.Called(ctx)
	return args.Get(0).(domain.RepositoryManager), args.Error(1)
}

func (m *MockRepositoryManager) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRepositoryManager) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

type MockEquipmentRepository struct {
	mock.Mock
}

func (m *MockEquipmentRepository) Save(ctx context.Context, equipment *domain.Equipment) error {
	args := m.Called(ctx, equipment)
	return args.Error(0)
}

func (m *MockEquipmentRepository) FindByID(ctx context.Context, id string) (*domain.Equipment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Equipment), args.Error(1)
}

func (m *MockEquipmentRepository) FindByStationType(ctx context.Context, stationType domain.StationType) ([]*domain.Equipment, error) {
	args := m.Called(ctx, stationType)
	return args.Get(0).([]*domain.Equipment), args.Error(1)
}

func (m *MockEquipmentRepository) FindAvailable(ctx context.Context) ([]*domain.Equipment, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Equipment), args.Error(1)
}

func (m *MockEquipmentRepository) FindAll(ctx context.Context, filter *domain.EquipmentFilter) ([]*domain.Equipment, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.Equipment), args.Error(1)
}

func (m *MockEquipmentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockQueueService struct {
	mock.Mock
}

func (m *MockQueueService) AddOrder(ctx context.Context, order *domain.KitchenOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockQueueService) GetNextOrder(ctx context.Context) (*domain.KitchenOrder, error) {
	args := m.Called(ctx)
	return args.Get(0).(*domain.KitchenOrder), args.Error(1)
}

func (m *MockQueueService) GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*domain.QueueStatus), args.Error(1)
}

func (m *MockQueueService) OptimizeQueue(ctx context.Context) (*QueueOptimization, error) {
	args := m.Called(ctx)
	return args.Get(0).(*QueueOptimization), args.Error(1)
}

func (m *MockQueueService) RebalanceQueue(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockOptimizerService struct {
	mock.Mock
}

func (m *MockOptimizerService) OptimizeQueue(ctx context.Context, orders []*domain.KitchenOrder, equipment []*domain.Equipment, staff []*domain.Staff) (*QueueOptimization, error) {
	args := m.Called(ctx, orders, equipment, staff)
	return args.Get(0).(*QueueOptimization), args.Error(1)
}

func (m *MockOptimizerService) PredictCapacity(ctx context.Context, timeRange *TimePeriod) (*CapacityPrediction, error) {
	args := m.Called(ctx, timeRange)
	return args.Get(0).(*CapacityPrediction), args.Error(1)
}

func (m *MockOptimizerService) SuggestWorkflow(ctx context.Context, order *domain.KitchenOrder) (*WorkflowSuggestion, error) {
	args := m.Called(ctx, order)
	return args.Get(0).(*WorkflowSuggestion), args.Error(1)
}

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) NotifyOrderAdded(ctx context.Context, order *domain.KitchenOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyOrderStatusChanged(ctx context.Context, order *domain.KitchenOrder, oldStatus domain.OrderStatus) error {
	args := m.Called(ctx, order, oldStatus)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyOrderCompleted(ctx context.Context, order *domain.KitchenOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyOrderOverdue(ctx context.Context, order *domain.KitchenOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyStaffAssigned(ctx context.Context, staff *domain.Staff, order *domain.KitchenOrder) error {
	args := m.Called(ctx, staff, order)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyStaffOverloaded(ctx context.Context, staff *domain.Staff) error {
	args := m.Called(ctx, staff)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyEquipmentMaintenance(ctx context.Context, equipment *domain.Equipment) error {
	args := m.Called(ctx, equipment)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyEquipmentOverloaded(ctx context.Context, equipment *domain.Equipment) error {
	args := m.Called(ctx, equipment)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyQueueBacklog(ctx context.Context, queueStatus *domain.QueueStatus) error {
	args := m.Called(ctx, queueStatus)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyCapacityAlert(ctx context.Context, prediction *CapacityPrediction) error {
	args := m.Called(ctx, prediction)
	return args.Error(0)
}

type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) PublishEvent(ctx context.Context, event *domain.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventService) PublishEvents(ctx context.Context, events []*domain.DomainEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockEventService) HandleOrderEvent(ctx context.Context, event *domain.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventService) HandleEquipmentEvent(ctx context.Context, event *domain.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventService) HandleStaffEvent(ctx context.Context, event *domain.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventService) HandleQueueEvent(ctx context.Context, event *domain.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventService) SubscribeToEvents(ctx context.Context, eventTypes []string, handler domain.EventHandler) error {
	args := m.Called(ctx, eventTypes, handler)
	return args.Error(0)
}

func (m *MockEventService) UnsubscribeFromEvents(ctx context.Context, eventTypes []string) error {
	args := m.Called(ctx, eventTypes)
	return args.Error(0)
}

// Test cases

func TestKitchenService_CreateEquipment(t *testing.T) {
	// Setup
	mockRepo := &MockRepositoryManager{}
	mockEquipmentRepo := &MockEquipmentRepository{}
	mockQueueService := &MockQueueService{}
	mockOptimizerService := &MockOptimizerService{}
	mockNotificationService := &MockNotificationService{}
	mockEventService := &MockEventService{}
	logger := logger.New("test")

	service := NewKitchenService(
		mockRepo,
		mockQueueService,
		mockOptimizerService,
		mockNotificationService,
		mockEventService,
		logger,
	)

	ctx := context.Background()
	req := &CreateEquipmentRequest{
		ID:          "espresso-01",
		Name:        "Professional Espresso Machine",
		StationType: domain.StationTypeEspresso,
	}

	// Setup mocks
	mockRepo.On("Equipment").Return(mockEquipmentRepo)
	mockEquipmentRepo.On("Save", ctx, mock.AnythingOfType("*domain.Equipment")).Return(nil)
	mockEventService.On("PublishEvents", ctx, mock.AnythingOfType("[]*domain.DomainEvent")).Return(nil)

	// Execute
	equipment, err := service.CreateEquipment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, equipment)
	assert.Equal(t, req.ID, equipment.ID)
	assert.Equal(t, req.Name, equipment.Name)
	assert.Equal(t, req.StationType, equipment.StationType)

	mockRepo.AssertExpectations(t)
	mockEquipmentRepo.AssertExpectations(t)
	mockEventService.AssertExpectations(t)
}

func TestKitchenService_AddOrderToQueue(t *testing.T) {
	// Setup
	mockRepo := &MockRepositoryManager{}
	mockQueueService := &MockQueueService{}
	mockOptimizerService := &MockOptimizerService{}
	mockNotificationService := &MockNotificationService{}
	mockEventService := &MockEventService{}
	logger := logger.New("test")

	service := NewKitchenService(
		mockRepo,
		mockQueueService,
		mockOptimizerService,
		mockNotificationService,
		mockEventService,
		logger,
	)

	ctx := context.Background()
	req := &AddOrderRequest{
		ID:         "order-123",
		CustomerID: "customer-456",
		Items: []*OrderItemRequest{
			{
				ID:           "item-1",
				Name:         "Espresso",
				Quantity:     1,
				Instructions: "Extra hot",
				Requirements: []domain.StationType{domain.StationTypeEspresso},
			},
		},
		Priority: domain.OrderPriorityNormal,
	}

	// Setup mocks
	mockQueueService.On("AddOrder", ctx, mock.AnythingOfType("*domain.KitchenOrder")).Return(nil)
	mockNotificationService.On("NotifyOrderAdded", ctx, mock.AnythingOfType("*domain.KitchenOrder")).Return(nil)
	mockEventService.On("PublishEvents", ctx, mock.AnythingOfType("[]*domain.DomainEvent")).Return(nil)

	// Execute
	order, err := service.AddOrderToQueue(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, req.ID, order.ID)
	assert.Equal(t, req.CustomerID, order.CustomerID)
	assert.Equal(t, domain.OrderStatusPending, order.Status)

	mockQueueService.AssertExpectations(t)
	mockNotificationService.AssertExpectations(t)
	mockEventService.AssertExpectations(t)
}

func TestKitchenService_GetQueueStatus(t *testing.T) {
	// Setup
	mockRepo := &MockRepositoryManager{}
	mockQueueService := &MockQueueService{}
	mockOptimizerService := &MockOptimizerService{}
	mockNotificationService := &MockNotificationService{}
	mockEventService := &MockEventService{}
	logger := logger.New("test")

	service := NewKitchenService(
		mockRepo,
		mockQueueService,
		mockOptimizerService,
		mockNotificationService,
		mockEventService,
		logger,
	)

	ctx := context.Background()
	expectedStatus := &domain.QueueStatus{
		TotalOrders:     10,
		PendingOrders:   5,
		ProcessingOrders: 3,
		CompletedOrders: 2,
		AverageWaitTime: 300, // 5 minutes
		EstimatedWaitTime: 600, // 10 minutes
	}

	// Setup mocks
	mockQueueService.On("GetQueueStatus", ctx).Return(expectedStatus, nil)

	// Execute
	status, err := service.GetQueueStatus(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, expectedStatus.TotalOrders, status.TotalOrders)
	assert.Equal(t, expectedStatus.PendingOrders, status.PendingOrders)

	mockQueueService.AssertExpectations(t)
}

func TestKitchenService_OptimizeQueue(t *testing.T) {
	// Setup
	mockRepo := &MockRepositoryManager{}
	mockQueueService := &MockQueueService{}
	mockOptimizerService := &MockOptimizerService{}
	mockNotificationService := &MockNotificationService{}
	mockEventService := &MockEventService{}
	logger := logger.New("test")

	service := NewKitchenService(
		mockRepo,
		mockQueueService,
		mockOptimizerService,
		mockNotificationService,
		mockEventService,
		logger,
	)

	ctx := context.Background()
	expectedOptimization := &QueueOptimization{
		OptimizedOrders: []*OptimizedOrder{
			{
				OrderID:          "order-1",
				NewPosition:      1,
				EstimatedTime:    300,
				AssignedStaffID:  "staff-1",
				RequiredStations: []domain.StationType{domain.StationTypeEspresso},
			},
		},
		EstimatedTimeReduction: 120, // 2 minutes saved
		EfficiencyGain:        15.5, // 15.5% improvement
	}

	// Setup mocks
	mockQueueService.On("OptimizeQueue", ctx).Return(expectedOptimization, nil)

	// Execute
	optimization, err := service.OptimizeQueue(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, optimization)
	assert.Equal(t, expectedOptimization.EstimatedTimeReduction, optimization.EstimatedTimeReduction)
	assert.Equal(t, expectedOptimization.EfficiencyGain, optimization.EfficiencyGain)

	mockQueueService.AssertExpectations(t)
}

func TestKitchenService_UpdateOrderStatus(t *testing.T) {
	// Setup
	mockRepo := &MockRepositoryManager{}
	mockOrderRepo := &MockOrderRepository{}
	mockQueueService := &MockQueueService{}
	mockOptimizerService := &MockOptimizerService{}
	mockNotificationService := &MockNotificationService{}
	mockEventService := &MockEventService{}
	logger := logger.New("test")

	service := NewKitchenService(
		mockRepo,
		mockQueueService,
		mockOptimizerService,
		mockNotificationService,
		mockEventService,
		logger,
	)

	ctx := context.Background()
	orderID := "order-123"
	newStatus := domain.OrderStatusProcessing

	// Create test order
	items := []*domain.OrderItem{
		{
			ID:           "item-1",
			Name:         "Espresso",
			Quantity:     1,
			Requirements: []domain.StationType{domain.StationTypeEspresso},
		},
	}
	order, _ := domain.NewKitchenOrder(orderID, "customer-456", items)

	// Setup mocks
	mockRepo.On("Order").Return(mockOrderRepo)
	mockOrderRepo.On("FindByID", ctx, orderID).Return(order, nil)
	mockOrderRepo.On("Save", ctx, order).Return(nil)
	mockNotificationService.On("NotifyOrderStatusChanged", ctx, order, domain.OrderStatusPending).Return(nil)
	mockEventService.On("PublishEvents", ctx, mock.AnythingOfType("[]*domain.DomainEvent")).Return(nil)

	// Execute
	err := service.UpdateOrderStatus(ctx, orderID, newStatus)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, newStatus, order.Status())

	mockRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockNotificationService.AssertExpectations(t)
	mockEventService.AssertExpectations(t)
}

// Mock OrderRepository for the test
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Save(ctx context.Context, order *domain.KitchenOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id string) (*domain.KitchenOrder, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.KitchenOrder), args.Error(1)
}

func (m *MockOrderRepository) FindByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.KitchenOrder, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*domain.KitchenOrder), args.Error(1)
}

func (m *MockOrderRepository) FindByCustomerID(ctx context.Context, customerID string) ([]*domain.KitchenOrder, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]*domain.KitchenOrder), args.Error(1)
}

func (m *MockOrderRepository) FindAll(ctx context.Context, filter *domain.OrderFilter) ([]*domain.KitchenOrder, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.KitchenOrder), args.Error(1)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
