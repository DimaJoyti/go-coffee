// +build integration

package kitchen

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/infrastructure/ai"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/infrastructure/repository"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// KitchenIntegrationTestSuite provides integration tests for the kitchen service
type KitchenIntegrationTestSuite struct {
	suite.Suite
	redisClient     *redis.Client
	repoManager     domain.RepositoryManager
	kitchenService  application.KitchenService
	queueService    application.QueueService
	optimizerService application.OptimizerService
	logger          *logger.Logger
}

// SetupSuite sets up the test suite
func (suite *KitchenIntegrationTestSuite) SetupSuite() {
	// Initialize logger
	suite.logger = logger.New("integration-test")

	// Initialize Redis client for testing
	suite.redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15, // Use a different DB for testing
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := suite.redisClient.Ping(ctx).Result()
	require.NoError(suite.T(), err, "Redis should be available for integration tests")

	// Initialize repository manager
	suite.repoManager = repository.NewRedisRepositoryManager(suite.redisClient, suite.logger)

	// Initialize AI optimizer service
	suite.optimizerService = ai.NewOptimizerService(suite.logger)

	// Initialize mock services for integration testing
	notificationService := &MockNotificationService{}
	eventService := &MockEventService{}

	// Initialize queue service
	suite.queueService = application.NewQueueService(
		suite.repoManager,
		suite.optimizerService,
		eventService,
		suite.logger,
	)

	// Initialize kitchen service
	suite.kitchenService = application.NewKitchenService(
		suite.repoManager,
		suite.queueService,
		suite.optimizerService,
		notificationService,
		eventService,
		suite.logger,
	)
}

// TearDownSuite cleans up after the test suite
func (suite *KitchenIntegrationTestSuite) TearDownSuite() {
	if suite.redisClient != nil {
		// Clean up test data
		ctx := context.Background()
		suite.redisClient.FlushDB(ctx)
		suite.redisClient.Close()
	}
}

// SetupTest sets up each test
func (suite *KitchenIntegrationTestSuite) SetupTest() {
	// Clean up test data before each test
	ctx := context.Background()
	suite.redisClient.FlushDB(ctx)
}

// TestCompleteOrderWorkflow tests the complete order workflow
func (suite *KitchenIntegrationTestSuite) TestCompleteOrderWorkflow() {
	ctx := context.Background()

	// Step 1: Create equipment
	equipmentReq := &application.CreateEquipmentRequest{
		ID:          "espresso-01",
		Name:        "Professional Espresso Machine",
		StationType: domain.StationTypeEspresso,
	}

	equipment, err := suite.kitchenService.CreateEquipment(ctx, equipmentReq)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), equipmentReq.ID, equipment.ID)

	// Step 2: Create staff
	staffReq := &application.CreateStaffRequest{
		ID:   "staff-01",
		Name: "Alice Cooper",
		Specializations: []domain.StationType{
			domain.StationTypeEspresso,
			domain.StationTypeGrinder,
		},
		SkillLevel: 8.5,
	}

	staff, err := suite.kitchenService.CreateStaff(ctx, staffReq)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), staffReq.ID, staff.ID)

	// Step 3: Add order to queue
	orderReq := &application.AddOrderRequest{
		ID:         "order-123",
		CustomerID: "customer-456",
		Items: []*application.OrderItemRequest{
			{
				ID:           "item-1",
				Name:         "Espresso",
				Quantity:     2,
				Instructions: "Extra hot",
				Requirements: []domain.StationType{domain.StationTypeEspresso, domain.StationTypeGrinder},
			},
		},
		Priority: domain.OrderPriorityNormal,
	}

	order, err := suite.kitchenService.AddOrderToQueue(ctx, orderReq)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), orderReq.ID, order.ID)
	assert.Equal(suite.T(), domain.OrderStatusPending, order.Status)

	// Step 4: Get queue status
	queueStatus, err := suite.kitchenService.GetQueueStatus(ctx)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int32(1), queueStatus.TotalOrders)
	assert.Equal(suite.T(), int32(1), queueStatus.PendingOrders)

	// Step 5: Get next order
	nextOrder, err := suite.kitchenService.GetNextOrder(ctx)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), order.ID, nextOrder.ID)

	// Step 6: Assign staff to order
	err = suite.kitchenService.AssignOrderToStaff(ctx, order.ID, staff.ID)
	require.NoError(suite.T(), err)

	// Step 7: Start order processing
	err = suite.kitchenService.StartOrderProcessing(ctx, order.ID)
	require.NoError(suite.T(), err)

	// Verify order status changed
	updatedOrder, err := suite.kitchenService.GetOrder(ctx, order.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), domain.OrderStatusProcessing, updatedOrder.Status)

	// Step 8: Complete order
	err = suite.kitchenService.CompleteOrder(ctx, order.ID)
	require.NoError(suite.T(), err)

	// Verify order is completed
	completedOrder, err := suite.kitchenService.GetOrder(ctx, order.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), domain.OrderStatusCompleted, completedOrder.Status)
	assert.NotNil(suite.T(), completedOrder.CompletedAt)
}

// TestQueueOptimization tests queue optimization functionality
func (suite *KitchenIntegrationTestSuite) TestQueueOptimization() {
	ctx := context.Background()

	// Create equipment
	equipmentTypes := []domain.StationType{
		domain.StationTypeEspresso,
		domain.StationTypeGrinder,
		domain.StationTypeSteamer,
	}

	for i, stationType := range equipmentTypes {
		equipmentReq := &application.CreateEquipmentRequest{
			ID:          fmt.Sprintf("equipment-%d", i+1),
			Name:        fmt.Sprintf("Equipment %d", i+1),
			StationType: stationType,
		}

		_, err := suite.kitchenService.CreateEquipment(ctx, equipmentReq)
		require.NoError(suite.T(), err)
	}

	// Create staff
	staffReq := &application.CreateStaffRequest{
		ID:   "staff-01",
		Name: "Alice Cooper",
		Specializations: equipmentTypes,
		SkillLevel: 9.0,
	}

	_, err := suite.kitchenService.CreateStaff(ctx, staffReq)
	require.NoError(suite.T(), err)

	// Add multiple orders with different priorities
	orders := []*application.AddOrderRequest{
		{
			ID:         "order-1",
			CustomerID: "customer-1",
			Items: []*application.OrderItemRequest{
				{
					ID:           "item-1",
					Name:         "Espresso",
					Quantity:     1,
					Requirements: []domain.StationType{domain.StationTypeEspresso},
				},
			},
			Priority: domain.OrderPriorityLow,
		},
		{
			ID:         "order-2",
			CustomerID: "customer-2",
			Items: []*application.OrderItemRequest{
				{
					ID:           "item-2",
					Name:         "Cappuccino",
					Quantity:     1,
					Requirements: []domain.StationType{domain.StationTypeEspresso, domain.StationTypeSteamer},
				},
			},
			Priority: domain.OrderPriorityUrgent,
		},
		{
			ID:         "order-3",
			CustomerID: "customer-3",
			Items: []*application.OrderItemRequest{
				{
					ID:           "item-3",
					Name:         "Americano",
					Quantity:     1,
					Requirements: []domain.StationType{domain.StationTypeEspresso, domain.StationTypeGrinder},
				},
			},
			Priority: domain.OrderPriorityHigh,
		},
	}

	// Add orders to queue
	for _, orderReq := range orders {
		_, err := suite.kitchenService.AddOrderToQueue(ctx, orderReq)
		require.NoError(suite.T(), err)
	}

	// Optimize queue
	optimization, err := suite.kitchenService.OptimizeQueue(ctx)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), optimization)
	assert.Len(suite.T(), optimization.OptimizedOrders, 3)

	// Verify urgent order is prioritized
	urgentOrder := optimization.OptimizedOrders[0]
	assert.Equal(suite.T(), "order-2", urgentOrder.OrderID) // Urgent priority should be first
}

// TestEquipmentManagement tests equipment management functionality
func (suite *KitchenIntegrationTestSuite) TestEquipmentManagement() {
	ctx := context.Background()

	// Create equipment
	equipmentReq := &application.CreateEquipmentRequest{
		ID:          "espresso-01",
		Name:        "Professional Espresso Machine",
		StationType: domain.StationTypeEspresso,
	}

	equipment, err := suite.kitchenService.CreateEquipment(ctx, equipmentReq)
	require.NoError(suite.T(), err)

	// Get equipment
	retrievedEquipment, err := suite.kitchenService.GetEquipment(ctx, equipment.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), equipment.ID, retrievedEquipment.ID)

	// Update equipment status
	err = suite.kitchenService.UpdateEquipmentStatus(ctx, equipment.ID, domain.EquipmentStatusBusy)
	require.NoError(suite.T(), err)

	// Verify status update
	updatedEquipment, err := suite.kitchenService.GetEquipment(ctx, equipment.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), domain.EquipmentStatusBusy, updatedEquipment.Status)

	// Schedule maintenance
	err = suite.kitchenService.ScheduleEquipmentMaintenance(ctx, equipment.ID)
	require.NoError(suite.T(), err)

	// Verify maintenance status
	maintenanceEquipment, err := suite.kitchenService.GetEquipment(ctx, equipment.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), domain.EquipmentStatusMaintenance, maintenanceEquipment.Status)
}

// TestStaffManagement tests staff management functionality
func (suite *KitchenIntegrationTestSuite) TestStaffManagement() {
	ctx := context.Background()

	// Create staff
	staffReq := &application.CreateStaffRequest{
		ID:   "staff-01",
		Name: "Alice Cooper",
		Specializations: []domain.StationType{
			domain.StationTypeEspresso,
			domain.StationTypeSteamer,
		},
		SkillLevel: 8.5,
	}

	staff, err := suite.kitchenService.CreateStaff(ctx, staffReq)
	require.NoError(suite.T(), err)

	// Get staff
	retrievedStaff, err := suite.kitchenService.GetStaff(ctx, staff.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), staff.ID, retrievedStaff.ID)

	// Update staff availability
	err = suite.kitchenService.UpdateStaffAvailability(ctx, staff.ID, false)
	require.NoError(suite.T(), err)

	// Verify availability update
	updatedStaff, err := suite.kitchenService.GetStaff(ctx, staff.ID)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), updatedStaff.Available)

	// Update staff skill
	newSkillLevel := float32(9.0)
	err = suite.kitchenService.UpdateStaffSkill(ctx, staff.ID, newSkillLevel)
	require.NoError(suite.T(), err)

	// Verify skill update
	skilledStaff, err := suite.kitchenService.GetStaff(ctx, staff.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), newSkillLevel, skilledStaff.SkillLevel)
}

// TestMetricsAndAnalytics tests metrics and analytics functionality
func (suite *KitchenIntegrationTestSuite) TestMetricsAndAnalytics() {
	ctx := context.Background()

	// Create some test data
	suite.createTestData(ctx)

	// Get kitchen metrics
	timePeriod := &application.TimePeriod{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	metrics, err := suite.kitchenService.GetKitchenMetrics(ctx, timePeriod)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), metrics)

	// Get performance report
	report, err := suite.kitchenService.GetPerformanceReport(ctx, timePeriod)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), report)
}

// Helper method to create test data
func (suite *KitchenIntegrationTestSuite) createTestData(ctx context.Context) {
	// Create equipment
	equipmentReq := &application.CreateEquipmentRequest{
		ID:          "test-equipment",
		Name:        "Test Equipment",
		StationType: domain.StationTypeEspresso,
	}
	suite.kitchenService.CreateEquipment(ctx, equipmentReq)

	// Create staff
	staffReq := &application.CreateStaffRequest{
		ID:              "test-staff",
		Name:            "Test Staff",
		Specializations: []domain.StationType{domain.StationTypeEspresso},
		SkillLevel:      8.0,
	}
	suite.kitchenService.CreateStaff(ctx, staffReq)

	// Create order
	orderReq := &application.AddOrderRequest{
		ID:         "test-order",
		CustomerID: "test-customer",
		Items: []*application.OrderItemRequest{
			{
				ID:           "test-item",
				Name:         "Test Item",
				Quantity:     1,
				Requirements: []domain.StationType{domain.StationTypeEspresso},
			},
		},
		Priority: domain.OrderPriorityNormal,
	}
	suite.kitchenService.AddOrderToQueue(ctx, orderReq)
}

// Mock services for integration testing

type MockNotificationService struct{}

func (m *MockNotificationService) NotifyOrderAdded(ctx context.Context, order *domain.KitchenOrder) error {
	return nil
}

func (m *MockNotificationService) NotifyOrderStatusChanged(ctx context.Context, order *domain.KitchenOrder, oldStatus domain.OrderStatus) error {
	return nil
}

func (m *MockNotificationService) NotifyOrderCompleted(ctx context.Context, order *domain.KitchenOrder) error {
	return nil
}

func (m *MockNotificationService) NotifyOrderOverdue(ctx context.Context, order *domain.KitchenOrder) error {
	return nil
}

func (m *MockNotificationService) NotifyStaffAssigned(ctx context.Context, staff *domain.Staff, order *domain.KitchenOrder) error {
	return nil
}

func (m *MockNotificationService) NotifyStaffOverloaded(ctx context.Context, staff *domain.Staff) error {
	return nil
}

func (m *MockNotificationService) NotifyEquipmentMaintenance(ctx context.Context, equipment *domain.Equipment) error {
	return nil
}

func (m *MockNotificationService) NotifyEquipmentOverloaded(ctx context.Context, equipment *domain.Equipment) error {
	return nil
}

func (m *MockNotificationService) NotifyQueueBacklog(ctx context.Context, queueStatus *domain.QueueStatus) error {
	return nil
}

func (m *MockNotificationService) NotifyCapacityAlert(ctx context.Context, prediction *application.CapacityPrediction) error {
	return nil
}

type MockEventService struct{}

func (m *MockEventService) PublishEvent(ctx context.Context, event *domain.DomainEvent) error {
	return nil
}

func (m *MockEventService) PublishEvents(ctx context.Context, events []*domain.DomainEvent) error {
	return nil
}

func (m *MockEventService) HandleOrderEvent(ctx context.Context, event *domain.DomainEvent) error {
	return nil
}

func (m *MockEventService) HandleEquipmentEvent(ctx context.Context, event *domain.DomainEvent) error {
	return nil
}

func (m *MockEventService) HandleStaffEvent(ctx context.Context, event *domain.DomainEvent) error {
	return nil
}

func (m *MockEventService) HandleQueueEvent(ctx context.Context, event *domain.DomainEvent) error {
	return nil
}

func (m *MockEventService) SubscribeToEvents(ctx context.Context, eventTypes []string, handler domain.EventHandler) error {
	return nil
}

func (m *MockEventService) UnsubscribeFromEvents(ctx context.Context, eventTypes []string) error {
	return nil
}

// TestKitchenIntegration runs the integration test suite
func TestKitchenIntegration(t *testing.T) {
	suite.Run(t, new(KitchenIntegrationTestSuite))
}

// Helper function for fmt.Sprintf
import "fmt"
