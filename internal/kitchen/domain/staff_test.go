package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStaff(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		staffName string
		wantErr   bool
	}{
		{
			name:      "valid staff",
			id:        "staff-01",
			staffName: "Alice Cooper",
			wantErr:   false,
		},
		{
			name:      "empty id",
			id:        "",
			staffName: "Alice Cooper",
			wantErr:   true,
		},
		{
			name:      "empty name",
			id:        "staff-01",
			staffName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			staff, err := NewStaff(tt.id, tt.staffName)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, staff)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, staff)
				assert.Equal(t, tt.id, staff.ID())
				assert.Equal(t, tt.staffName, staff.Name())
				assert.True(t, staff.IsAvailable())
				assert.Equal(t, float32(5.0), staff.SkillLevel()) // Default skill level
				assert.Equal(t, int32(0), staff.CurrentOrders())
				assert.Empty(t, staff.Specializations())
			}
		})
	}
}

func TestStaff_AddSpecialization(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	tests := []struct {
		name           string
		specialization StationType
		wantErr        bool
	}{
		{
			name:           "valid specialization",
			specialization: StationTypeEspresso,
			wantErr:        false,
		},
		{
			name:           "another valid specialization",
			specialization: StationTypeSteamer,
			wantErr:        false,
		},
		{
			name:           "duplicate specialization",
			specialization: StationTypeEspresso,
			wantErr:        false, // Should not error, just not add duplicate
		},
		{
			name:           "invalid specialization",
			specialization: StationType(999),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := staff.AddSpecialization(tt.specialization)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, staff.Specializations(), tt.specialization)
			}
		})
	}

	// Check that we have unique specializations
	assert.Len(t, staff.Specializations(), 2) // Should have Espresso and Steamer only
}

func TestStaff_RemoveSpecialization(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	// Add some specializations
	staff.AddSpecialization(StationTypeEspresso)
	staff.AddSpecialization(StationTypeSteamer)

	tests := []struct {
		name           string
		specialization StationType
		shouldExist    bool
	}{
		{
			name:           "remove existing specialization",
			specialization: StationTypeEspresso,
			shouldExist:    false,
		},
		{
			name:           "remove non-existing specialization",
			specialization: StationTypeGrinder,
			shouldExist:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			staff.RemoveSpecialization(tt.specialization)
			
			if tt.shouldExist {
				assert.Contains(t, staff.Specializations(), tt.specialization)
			} else {
				assert.NotContains(t, staff.Specializations(), tt.specialization)
			}
		})
	}
}

func TestStaff_HasSpecialization(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	staff.AddSpecialization(StationTypeEspresso)
	staff.AddSpecialization(StationTypeSteamer)

	tests := []struct {
		name           string
		specialization StationType
		expected       bool
	}{
		{
			name:           "has specialization",
			specialization: StationTypeEspresso,
			expected:       true,
		},
		{
			name:           "has another specialization",
			specialization: StationTypeSteamer,
			expected:       true,
		},
		{
			name:           "does not have specialization",
			specialization: StationTypeGrinder,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := staff.HasSpecialization(tt.specialization)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStaff_UpdateSkillLevel(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	tests := []struct {
		name       string
		skillLevel float32
		wantErr    bool
	}{
		{
			name:       "valid skill level",
			skillLevel: 8.5,
			wantErr:    false,
		},
		{
			name:       "minimum skill level",
			skillLevel: 0.0,
			wantErr:    false,
		},
		{
			name:       "maximum skill level",
			skillLevel: 10.0,
			wantErr:    false,
		},
		{
			name:       "negative skill level",
			skillLevel: -1.0,
			wantErr:    true,
		},
		{
			name:       "over maximum skill level",
			skillLevel: 11.0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := staff.UpdateSkillLevel(tt.skillLevel)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.skillLevel, staff.SkillLevel())
			}
		})
	}
}

func TestStaff_SetAvailability(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	// Test setting unavailable
	staff.SetAvailability(false)
	assert.False(t, staff.IsAvailable())

	// Test setting available
	staff.SetAvailability(true)
	assert.True(t, staff.IsAvailable())
}

func TestStaff_AssignOrder(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	orderID := "order-123"

	// Assign order
	err = staff.AssignOrder(orderID)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), staff.CurrentOrders())
	assert.Contains(t, staff.AssignedOrders(), orderID)

	// Assign another order
	err = staff.AssignOrder("order-456")
	assert.NoError(t, err)
	assert.Equal(t, int32(2), staff.CurrentOrders())

	// Try to assign duplicate order
	err = staff.AssignOrder(orderID)
	assert.Error(t, err) // Should error on duplicate
}

func TestStaff_CompleteOrder(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	orderID := "order-123"

	// Try to complete order that wasn't assigned
	err = staff.CompleteOrder(orderID)
	assert.Error(t, err)

	// Assign and then complete order
	staff.AssignOrder(orderID)
	err = staff.CompleteOrder(orderID)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), staff.CurrentOrders())
	assert.NotContains(t, staff.AssignedOrders(), orderID)
}

func TestStaff_IsOverloaded(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	// Set max orders
	staff.SetMaxOrders(3)

	// Not overloaded
	staff.AssignOrder("order-1")
	staff.AssignOrder("order-2")
	assert.False(t, staff.IsOverloaded())

	// At capacity
	staff.AssignOrder("order-3")
	assert.True(t, staff.IsOverloaded())
}

func TestStaff_CanAcceptOrder(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	staff.SetMaxOrders(2)

	tests := []struct {
		name            string
		currentOrders   int
		available       bool
		expectedResult  bool
	}{
		{
			name:           "available with capacity",
			currentOrders:  1,
			available:      true,
			expectedResult: true,
		},
		{
			name:           "available at capacity",
			currentOrders:  2,
			available:      true,
			expectedResult: false,
		},
		{
			name:           "unavailable with capacity",
			currentOrders:  1,
			available:      false,
			expectedResult: false,
		},
		{
			name:           "available with no orders",
			currentOrders:  0,
			available:      true,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset staff state
			staff.ClearOrders()
			staff.SetAvailability(tt.available)

			// Assign orders
			for i := 0; i < tt.currentOrders; i++ {
				staff.AssignOrder(fmt.Sprintf("order-%d", i))
			}

			result := staff.CanAcceptOrder()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestStaff_GetWorkload(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	staff.SetMaxOrders(5)

	tests := []struct {
		name          string
		currentOrders int
		expectedLoad  float32
	}{
		{
			name:          "no orders",
			currentOrders: 0,
			expectedLoad:  0.0,
		},
		{
			name:          "half capacity",
			currentOrders: 2,
			expectedLoad:  0.4,
		},
		{
			name:          "full capacity",
			currentOrders: 5,
			expectedLoad:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			staff.ClearOrders()
			
			for i := 0; i < tt.currentOrders; i++ {
				staff.AssignOrder(fmt.Sprintf("order-%d", i))
			}

			workload := staff.GetWorkload()
			assert.Equal(t, tt.expectedLoad, workload)
		})
	}
}

func TestStaff_Events(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	// Test assignment event
	err = staff.AssignOrder("order-123")
	assert.NoError(t, err)

	events := staff.GetEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "kitchen.staff.order_assigned", events[0].Type)
	assert.Equal(t, staff.ID(), events[0].AggregateID)

	// Clear events
	staff.ClearEvents()
	assert.Len(t, staff.GetEvents(), 0)

	// Test overload event
	staff.SetMaxOrders(1)
	staff.AssignOrder("order-456") // This should trigger overload event

	events = staff.GetEvents()
	assert.Len(t, events, 2) // Assignment + overload events
	
	// Find overload event
	var overloadEvent *DomainEvent
	for _, event := range events {
		if event.Type == "kitchen.staff.overloaded" {
			overloadEvent = event
			break
		}
	}
	assert.NotNil(t, overloadEvent)
}

func TestStaff_ClearOrders(t *testing.T) {
	staff, err := NewStaff("staff-01", "Alice Cooper")
	require.NoError(t, err)

	// Assign some orders
	staff.AssignOrder("order-1")
	staff.AssignOrder("order-2")
	assert.Equal(t, int32(2), staff.CurrentOrders())

	// Clear orders
	staff.ClearOrders()
	assert.Equal(t, int32(0), staff.CurrentOrders())
	assert.Empty(t, staff.AssignedOrders())
}

// Helper function for fmt.Sprintf in tests
import "fmt"
