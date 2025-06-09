package domain

import (
	"errors"
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

// Equipment represents a piece of kitchen equipment (Domain Entity)
type Equipment struct {
	id              string
	name            string
	stationType     StationType
	status          EquipmentStatus
	efficiencyScore float32
	currentLoad     int32
	maxCapacity     int32
	lastMaintenance time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

// NewEquipment creates a new equipment entity with validation
func NewEquipment(id, name string, stationType StationType, maxCapacity int32) (*Equipment, error) {
	if id == "" {
		return nil, errors.New("equipment ID is required")
	}
	if name == "" {
		return nil, errors.New("equipment name is required")
	}
	if maxCapacity <= 0 {
		return nil, errors.New("max capacity must be greater than 0")
	}

	now := time.Now()
	return &Equipment{
		id:              id,
		name:            name,
		stationType:     stationType,
		status:          EquipmentStatusAvailable,
		efficiencyScore: 0.0,
		currentLoad:     0,
		maxCapacity:     maxCapacity,
		lastMaintenance: now,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// Getters
func (e *Equipment) ID() string              { return e.id }
func (e *Equipment) Name() string            { return e.name }
func (e *Equipment) StationType() StationType { return e.stationType }
func (e *Equipment) Status() EquipmentStatus { return e.status }
func (e *Equipment) EfficiencyScore() float32 { return e.efficiencyScore }
func (e *Equipment) CurrentLoad() int32      { return e.currentLoad }
func (e *Equipment) MaxCapacity() int32      { return e.maxCapacity }
func (e *Equipment) LastMaintenance() time.Time { return e.lastMaintenance }
func (e *Equipment) CreatedAt() time.Time    { return e.createdAt }
func (e *Equipment) UpdatedAt() time.Time    { return e.updatedAt }

// Business Methods

// UpdateStatus changes the equipment status with validation
func (e *Equipment) UpdateStatus(status EquipmentStatus) error {
	// Business rule: Cannot change status to in-use if at max capacity
	if status == EquipmentStatusInUse && e.currentLoad >= e.maxCapacity {
		return errors.New("equipment is at maximum capacity")
	}

	// Business rule: Cannot use broken equipment
	if e.status == EquipmentStatusBroken && status == EquipmentStatusInUse {
		return errors.New("cannot use broken equipment")
	}

	e.status = status
	e.updatedAt = time.Now()
	return nil
}

// AddLoad increases the current load with validation
func (e *Equipment) AddLoad(load int32) error {
	if load <= 0 {
		return errors.New("load must be positive")
	}
	if e.currentLoad+load > e.maxCapacity {
		return errors.New("would exceed maximum capacity")
	}
	if e.status != EquipmentStatusAvailable && e.status != EquipmentStatusInUse {
		return errors.New("equipment is not available for use")
	}

	e.currentLoad += load
	if e.currentLoad > 0 {
		e.status = EquipmentStatusInUse
	}
	e.updatedAt = time.Now()
	return nil
}

// RemoveLoad decreases the current load with validation
func (e *Equipment) RemoveLoad(load int32) error {
	if load <= 0 {
		return errors.New("load must be positive")
	}
	if e.currentLoad < load {
		return errors.New("cannot remove more load than current")
	}

	e.currentLoad -= load
	if e.currentLoad == 0 && e.status == EquipmentStatusInUse {
		e.status = EquipmentStatusAvailable
	}
	e.updatedAt = time.Now()
	return nil
}

// UpdateEfficiencyScore updates the efficiency score with validation
func (e *Equipment) UpdateEfficiencyScore(score float32) error {
	if score < 0.0 || score > 10.0 {
		return errors.New("efficiency score must be between 0.0 and 10.0")
	}

	e.efficiencyScore = score
	e.updatedAt = time.Now()
	return nil
}

// ScheduleMaintenance schedules maintenance for the equipment
func (e *Equipment) ScheduleMaintenance() error {
	if e.status == EquipmentStatusInUse {
		return errors.New("cannot schedule maintenance while equipment is in use")
	}

	e.status = EquipmentStatusMaintenance
	e.updatedAt = time.Now()
	return nil
}

// CompleteMaintenance marks maintenance as completed
func (e *Equipment) CompleteMaintenance() error {
	if e.status != EquipmentStatusMaintenance {
		return errors.New("equipment is not under maintenance")
	}

	e.status = EquipmentStatusAvailable
	e.lastMaintenance = time.Now()
	e.updatedAt = time.Now()
	return nil
}

// IsAvailable checks if equipment is available for use
func (e *Equipment) IsAvailable() bool {
	return e.status == EquipmentStatusAvailable && e.currentLoad < e.maxCapacity
}

// CanAcceptLoad checks if equipment can accept additional load
func (e *Equipment) CanAcceptLoad(load int32) bool {
	return e.IsAvailable() && e.currentLoad+load <= e.maxCapacity
}

// GetUtilizationRate returns the current utilization rate (0.0 to 1.0)
func (e *Equipment) GetUtilizationRate() float32 {
	if e.maxCapacity == 0 {
		return 0.0
	}
	return float32(e.currentLoad) / float32(e.maxCapacity)
}

// NeedsMaintenance checks if equipment needs maintenance based on business rules
func (e *Equipment) NeedsMaintenance() bool {
	// Business rule: Maintenance needed if last maintenance was more than 30 days ago
	// or efficiency score is below 7.0
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return e.lastMaintenance.Before(thirtyDaysAgo) || e.efficiencyScore < 7.0
}

// ToDTO converts domain entity to data transfer object
func (e *Equipment) ToDTO() *EquipmentDTO {
	return &EquipmentDTO{
		ID:              e.id,
		Name:            e.name,
		StationType:     e.stationType,
		Status:          e.status,
		EfficiencyScore: e.efficiencyScore,
		CurrentLoad:     e.currentLoad,
		MaxCapacity:     e.maxCapacity,
		LastMaintenance: e.lastMaintenance,
		CreatedAt:       e.createdAt,
		UpdatedAt:       e.updatedAt,
	}
}

// EquipmentDTO represents equipment data transfer object
type EquipmentDTO struct {
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
