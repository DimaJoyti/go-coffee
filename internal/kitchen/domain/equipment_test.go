package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEquipment(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		equipName   string
		stationType StationType
		wantErr     bool
	}{
		{
			name:        "valid equipment",
			id:          "espresso-01",
			equipName:   "Professional Espresso Machine",
			stationType: StationTypeEspresso,
			wantErr:     false,
		},
		{
			name:        "empty id",
			id:          "",
			equipName:   "Test Equipment",
			stationType: StationTypeEspresso,
			wantErr:     true,
		},
		{
			name:        "empty name",
			id:          "test-01",
			equipName:   "",
			stationType: StationTypeEspresso,
			wantErr:     true,
		},
		{
			name:        "invalid station type",
			id:          "test-01",
			equipName:   "Test Equipment",
			stationType: StationType(999),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equipment, err := NewEquipment(tt.id, tt.equipName, tt.stationType)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, equipment)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, equipment)
				assert.Equal(t, tt.id, equipment.ID())
				assert.Equal(t, tt.equipName, equipment.Name())
				assert.Equal(t, tt.stationType, equipment.StationType())
				assert.Equal(t, EquipmentStatusAvailable, equipment.Status())
				assert.Equal(t, float32(0), equipment.EfficiencyScore())
				assert.Equal(t, int32(0), equipment.CurrentLoad())
				assert.Equal(t, int32(1), equipment.MaxCapacity())
				assert.True(t, equipment.IsAvailable())
			}
		})
	}
}

func TestEquipment_UpdateStatus(t *testing.T) {
	equipment, err := NewEquipment("test-01", "Test Equipment", StationTypeEspresso)
	require.NoError(t, err)

	tests := []struct {
		name      string
		newStatus EquipmentStatus
		wantErr   bool
	}{
		{
			name:      "valid status change",
			newStatus: EquipmentStatusBusy,
			wantErr:   false,
		},
		{
			name:      "maintenance status",
			newStatus: EquipmentStatusMaintenance,
			wantErr:   false,
		},
		{
			name:      "offline status",
			newStatus: EquipmentStatusOffline,
			wantErr:   false,
		},
		{
			name:      "invalid status",
			newStatus: EquipmentStatus(999),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := equipment.UpdateStatus(tt.newStatus)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newStatus, equipment.Status())
			}
		})
	}
}

func TestEquipment_UpdateLoad(t *testing.T) {
	equipment, err := NewEquipment("test-01", "Test Equipment", StationTypeEspresso)
	require.NoError(t, err)
	
	// Set max capacity
	equipment.SetMaxCapacity(5)

	tests := []struct {
		name    string
		load    int32
		wantErr bool
	}{
		{
			name:    "valid load",
			load:    3,
			wantErr: false,
		},
		{
			name:    "zero load",
			load:    0,
			wantErr: false,
		},
		{
			name:    "max capacity load",
			load:    5,
			wantErr: false,
		},
		{
			name:    "negative load",
			load:    -1,
			wantErr: true,
		},
		{
			name:    "over capacity load",
			load:    6,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := equipment.UpdateLoad(tt.load)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.load, equipment.CurrentLoad())
			}
		})
	}
}

func TestEquipment_UpdateEfficiencyScore(t *testing.T) {
	equipment, err := NewEquipment("test-01", "Test Equipment", StationTypeEspresso)
	require.NoError(t, err)

	tests := []struct {
		name  string
		score float32
		valid bool
	}{
		{
			name:  "valid score",
			score: 8.5,
			valid: true,
		},
		{
			name:  "minimum score",
			score: 0.0,
			valid: true,
		},
		{
			name:  "maximum score",
			score: 10.0,
			valid: true,
		},
		{
			name:  "negative score",
			score: -1.0,
			valid: false,
		},
		{
			name:  "over maximum score",
			score: 11.0,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := equipment.UpdateEfficiencyScore(tt.score)
			
			if tt.valid {
				assert.NoError(t, err)
				assert.Equal(t, tt.score, equipment.EfficiencyScore())
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestEquipment_ScheduleMaintenance(t *testing.T) {
	equipment, err := NewEquipment("test-01", "Test Equipment", StationTypeEspresso)
	require.NoError(t, err)

	// Schedule maintenance
	err = equipment.ScheduleMaintenance()
	assert.NoError(t, err)
	assert.Equal(t, EquipmentStatusMaintenance, equipment.Status())
	assert.NotNil(t, equipment.LastMaintenanceAt())
	assert.False(t, equipment.IsAvailable())
}

func TestEquipment_IsOverloaded(t *testing.T) {
	equipment, err := NewEquipment("test-01", "Test Equipment", StationTypeEspresso)
	require.NoError(t, err)
	
	equipment.SetMaxCapacity(5)

	// Test not overloaded
	equipment.UpdateLoad(3)
	assert.False(t, equipment.IsOverloaded())

	// Test at capacity
	equipment.UpdateLoad(5)
	assert.True(t, equipment.IsOverloaded())
}

func TestEquipment_CanAcceptLoad(t *testing.T) {
	equipment, err := NewEquipment("test-01", "Test Equipment", StationTypeEspresso)
	require.NoError(t, err)
	
	equipment.SetMaxCapacity(5)
	equipment.UpdateLoad(3)

	tests := []struct {
		name           string
		additionalLoad int32
		expected       bool
	}{
		{
			name:           "can accept load",
			additionalLoad: 1,
			expected:       true,
		},
		{
			name:           "can accept remaining capacity",
			additionalLoad: 2,
			expected:       true,
		},
		{
			name:           "cannot accept over capacity",
			additionalLoad: 3,
			expected:       false,
		},
		{
			name:           "zero additional load",
			additionalLoad: 0,
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := equipment.CanAcceptLoad(tt.additionalLoad)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEquipment_GetUtilizationRate(t *testing.T) {
	equipment, err := NewEquipment("test-01", "Test Equipment", StationTypeEspresso)
	require.NoError(t, err)
	
	equipment.SetMaxCapacity(10)

	tests := []struct {
		name         string
		currentLoad  int32
		expectedRate float32
	}{
		{
			name:         "no load",
			currentLoad:  0,
			expectedRate: 0.0,
		},
		{
			name:         "half load",
			currentLoad:  5,
			expectedRate: 0.5,
		},
		{
			name:         "full load",
			currentLoad:  10,
			expectedRate: 1.0,
		},
		{
			name:         "quarter load",
			currentLoad:  2,
			expectedRate: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equipment.UpdateLoad(tt.currentLoad)
			rate := equipment.GetUtilizationRate()
			assert.Equal(t, tt.expectedRate, rate)
		})
	}
}

func TestEquipment_NeedsMaintenance(t *testing.T) {
	equipment, err := NewEquipment("test-01", "Test Equipment", StationTypeEspresso)
	require.NoError(t, err)

	// New equipment doesn't need maintenance
	assert.False(t, equipment.NeedsMaintenance())

	// Set last maintenance to old date
	oldTime := time.Now().Add(-25 * time.Hour) // More than 24 hours ago
	equipment.SetLastMaintenanceAt(&oldTime)
	assert.True(t, equipment.NeedsMaintenance())

	// Recent maintenance
	recentTime := time.Now().Add(-1 * time.Hour)
	equipment.SetLastMaintenanceAt(&recentTime)
	assert.False(t, equipment.NeedsMaintenance())
}

func TestEquipment_Events(t *testing.T) {
	equipment, err := NewEquipment("test-01", "Test Equipment", StationTypeEspresso)
	require.NoError(t, err)

	// Test status change event
	err = equipment.UpdateStatus(EquipmentStatusBusy)
	assert.NoError(t, err)

	events := equipment.GetEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "kitchen.equipment.status_changed", events[0].Type)
	assert.Equal(t, equipment.ID(), events[0].AggregateID)

	// Clear events
	equipment.ClearEvents()
	assert.Len(t, equipment.GetEvents(), 0)
}

func TestStationType_String(t *testing.T) {
	tests := []struct {
		stationType StationType
		expected    string
	}{
		{StationTypeEspresso, "ESPRESSO"},
		{StationTypeGrinder, "GRINDER"},
		{StationTypeSteamer, "STEAMER"},
		{StationTypeAssembly, "ASSEMBLY"},
		{StationType(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.stationType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEquipmentStatus_String(t *testing.T) {
	tests := []struct {
		status   EquipmentStatus
		expected string
	}{
		{EquipmentStatusAvailable, "AVAILABLE"},
		{EquipmentStatusBusy, "BUSY"},
		{EquipmentStatusMaintenance, "MAINTENANCE"},
		{EquipmentStatusOffline, "OFFLINE"},
		{EquipmentStatus(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
