package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKitchenOrder(t *testing.T) {
	items := []*OrderItem{
		{
			ID:           "item-1",
			Name:         "Espresso",
			Quantity:     2,
			Instructions: "Extra hot",
			Requirements: []StationType{StationTypeEspresso, StationTypeGrinder},
		},
	}

	tests := []struct {
		name       string
		id         string
		customerID string
		items      []*OrderItem
		wantErr    bool
	}{
		{
			name:       "valid order",
			id:         "order-123",
			customerID: "customer-456",
			items:      items,
			wantErr:    false,
		},
		{
			name:       "empty id",
			id:         "",
			customerID: "customer-456",
			items:      items,
			wantErr:    true,
		},
		{
			name:       "empty customer id",
			id:         "order-123",
			customerID: "",
			items:      items,
			wantErr:    true,
		},
		{
			name:       "no items",
			id:         "order-123",
			customerID: "customer-456",
			items:      []*OrderItem{},
			wantErr:    true,
		},
		{
			name:       "nil items",
			id:         "order-123",
			customerID: "customer-456",
			items:      nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewKitchenOrder(tt.id, tt.customerID, tt.items)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, order)
				assert.Equal(t, tt.id, order.ID())
				assert.Equal(t, tt.customerID, order.CustomerID())
				assert.Equal(t, tt.items, order.Items())
				assert.Equal(t, OrderStatusPending, order.Status())
				assert.Equal(t, OrderPriorityNormal, order.Priority())
				assert.NotNil(t, order.CreatedAt())
				assert.Nil(t, order.StartedAt())
				assert.Nil(t, order.CompletedAt())
			}
		})
	}
}

func TestKitchenOrder_UpdateStatus(t *testing.T) {
	order := createTestOrder(t)

	tests := []struct {
		name      string
		newStatus OrderStatus
		wantErr   bool
	}{
		{
			name:      "pending to processing",
			newStatus: OrderStatusProcessing,
			wantErr:   false,
		},
		{
			name:      "processing to completed",
			newStatus: OrderStatusCompleted,
			wantErr:   false,
		},
		{
			name:      "invalid status",
			newStatus: OrderStatus(999),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStatus := order.Status()
			err := order.UpdateStatus(tt.newStatus)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newStatus, order.Status())
				
				// Check events
				events := order.GetEvents()
				found := false
				for _, event := range events {
					if event.Type == "kitchen.order.status_changed" {
						found = true
						assert.Equal(t, order.ID(), event.AggregateID)
						break
					}
				}
				assert.True(t, found, "Status change event should be generated")
			}
		})
	}
}

func TestKitchenOrder_SetPriority(t *testing.T) {
	order := createTestOrder(t)

	tests := []struct {
		name     string
		priority OrderPriority
		wantErr  bool
	}{
		{
			name:     "normal priority",
			priority: OrderPriorityNormal,
			wantErr:  false,
		},
		{
			name:     "high priority",
			priority: OrderPriorityHigh,
			wantErr:  false,
		},
		{
			name:     "urgent priority",
			priority: OrderPriorityUrgent,
			wantErr:  false,
		},
		{
			name:     "low priority",
			priority: OrderPriorityLow,
			wantErr:  false,
		},
		{
			name:     "invalid priority",
			priority: OrderPriority(999),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := order.SetPriority(tt.priority)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.priority, order.Priority())
			}
		})
	}
}

func TestKitchenOrder_AssignStaff(t *testing.T) {
	order := createTestOrder(t)

	tests := []struct {
		name    string
		staffID string
		wantErr bool
	}{
		{
			name:    "valid staff assignment",
			staffID: "staff-123",
			wantErr: false,
		},
		{
			name:    "empty staff id",
			staffID: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := order.AssignStaff(tt.staffID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.staffID, order.AssignedStaffID())
			}
		})
	}
}

func TestKitchenOrder_Start(t *testing.T) {
	order := createTestOrder(t)

	// Start the order
	err := order.Start()
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusProcessing, order.Status())
	assert.NotNil(t, order.StartedAt())
	assert.True(t, order.StartedAt().After(order.CreatedAt()))

	// Try to start already started order
	err = order.Start()
	assert.Error(t, err) // Should error if already started
}

func TestKitchenOrder_Complete(t *testing.T) {
	order := createTestOrder(t)

	// Try to complete without starting
	err := order.Complete()
	assert.Error(t, err) // Should error if not started

	// Start and then complete
	order.Start()
	err = order.Complete()
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusCompleted, order.Status())
	assert.NotNil(t, order.CompletedAt())
	assert.True(t, order.CompletedAt().After(*order.StartedAt()))

	// Check completion event
	events := order.GetEvents()
	found := false
	for _, event := range events {
		if event.Type == "kitchen.order.completed" {
			found = true
			break
		}
	}
	assert.True(t, found, "Completion event should be generated")
}

func TestKitchenOrder_Cancel(t *testing.T) {
	order := createTestOrder(t)

	// Cancel the order
	err := order.Cancel()
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusCancelled, order.Status())

	// Try to cancel already cancelled order
	err = order.Cancel()
	assert.Error(t, err) // Should error if already cancelled
}

func TestKitchenOrder_EstimatedTime(t *testing.T) {
	order := createTestOrder(t)

	// Set estimated time
	estimatedTime := int32(300) // 5 minutes
	order.SetEstimatedTime(estimatedTime)
	assert.Equal(t, estimatedTime, order.EstimatedTime())
}

func TestKitchenOrder_ActualTime(t *testing.T) {
	order := createTestOrder(t)

	// Start and complete order to calculate actual time
	order.Start()
	time.Sleep(10 * time.Millisecond) // Small delay
	order.Complete()

	actualTime := order.ActualTime()
	assert.Greater(t, actualTime, int32(0))
}

func TestKitchenOrder_IsOverdue(t *testing.T) {
	order := createTestOrder(t)

	// Set estimated time to 1 second
	order.SetEstimatedTime(1)
	order.Start()

	// Initially not overdue
	assert.False(t, order.IsOverdue())

	// Wait and check if overdue
	time.Sleep(2 * time.Second)
	assert.True(t, order.IsOverdue())
}

func TestKitchenOrder_GetRequiredStations(t *testing.T) {
	items := []*OrderItem{
		{
			ID:           "item-1",
			Name:         "Espresso",
			Quantity:     1,
			Requirements: []StationType{StationTypeEspresso, StationTypeGrinder},
		},
		{
			ID:           "item-2",
			Name:         "Cappuccino",
			Quantity:     1,
			Requirements: []StationType{StationTypeEspresso, StationTypeSteamer},
		},
	}

	order, err := NewKitchenOrder("order-123", "customer-456", items)
	require.NoError(t, err)

	requiredStations := order.GetRequiredStations()
	
	// Should have unique stations
	expectedStations := []StationType{
		StationTypeEspresso,
		StationTypeGrinder,
		StationTypeSteamer,
	}

	assert.Len(t, requiredStations, len(expectedStations))
	for _, station := range expectedStations {
		assert.Contains(t, requiredStations, station)
	}
}

func TestKitchenOrder_GetComplexity(t *testing.T) {
	tests := []struct {
		name               string
		items              []*OrderItem
		expectedComplexity int
	}{
		{
			name: "simple order",
			items: []*OrderItem{
				{
					ID:           "item-1",
					Name:         "Espresso",
					Quantity:     1,
					Requirements: []StationType{StationTypeEspresso},
				},
			},
			expectedComplexity: 1,
		},
		{
			name: "complex order",
			items: []*OrderItem{
				{
					ID:           "item-1",
					Name:         "Cappuccino",
					Quantity:     2,
					Requirements: []StationType{StationTypeEspresso, StationTypeSteamer},
				},
				{
					ID:           "item-2",
					Name:         "Mocha",
					Quantity:     1,
					Requirements: []StationType{StationTypeEspresso, StationTypeSteamer, StationTypeAssembly},
				},
			},
			expectedComplexity: 7, // (2*2) + (1*3) = 7
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewKitchenOrder("order-123", "customer-456", tt.items)
			require.NoError(t, err)

			complexity := order.GetComplexity()
			assert.Equal(t, tt.expectedComplexity, complexity)
		})
	}
}

func TestKitchenOrder_CanBeCombinedWith(t *testing.T) {
	order1 := createTestOrder(t)
	order2 := createTestOrder(t)
	order2.id = "order-456" // Different ID

	// Same customer - can be combined
	assert.True(t, order1.CanBeCombinedWith(order2))

	// Different customer - cannot be combined
	order2.customerID = "different-customer"
	assert.False(t, order1.CanBeCombinedWith(order2))

	// Different status - cannot be combined
	order2.customerID = order1.customerID // Reset customer
	order2.UpdateStatus(OrderStatusProcessing)
	assert.False(t, order1.CanBeCombinedWith(order2))
}

func TestOrderStatus_String(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected string
	}{
		{OrderStatusPending, "PENDING"},
		{OrderStatusProcessing, "PROCESSING"},
		{OrderStatusCompleted, "COMPLETED"},
		{OrderStatusCancelled, "CANCELLED"},
		{OrderStatus(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrderPriority_String(t *testing.T) {
	tests := []struct {
		priority OrderPriority
		expected string
	}{
		{OrderPriorityLow, "LOW"},
		{OrderPriorityNormal, "NORMAL"},
		{OrderPriorityHigh, "HIGH"},
		{OrderPriorityUrgent, "URGENT"},
		{OrderPriority(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.priority.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a test order
func createTestOrder(t *testing.T) *KitchenOrder {
	items := []*OrderItem{
		{
			ID:           "item-1",
			Name:         "Espresso",
			Quantity:     1,
			Instructions: "Extra hot",
			Requirements: []StationType{StationTypeEspresso, StationTypeGrinder},
		},
	}

	order, err := NewKitchenOrder("order-123", "customer-456", items)
	require.NoError(t, err)
	return order
}
