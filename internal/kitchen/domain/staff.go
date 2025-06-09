package domain

import (
	"errors"
	"time"
)

// Staff represents a kitchen staff member (Domain Entity)
type Staff struct {
	id                  string
	name                string
	specializations     []StationType
	skillLevel          float32
	isAvailable         bool
	currentOrders       int32
	maxConcurrentOrders int32
	createdAt           time.Time
	updatedAt           time.Time
}

// NewStaff creates a new staff entity with validation
func NewStaff(id, name string, specializations []StationType, skillLevel float32, maxConcurrentOrders int32) (*Staff, error) {
	if id == "" {
		return nil, errors.New("staff ID is required")
	}
	if name == "" {
		return nil, errors.New("staff name is required")
	}
	if skillLevel < 0.0 || skillLevel > 10.0 {
		return nil, errors.New("skill level must be between 0.0 and 10.0")
	}
	if maxConcurrentOrders <= 0 {
		return nil, errors.New("max concurrent orders must be greater than 0")
	}
	if len(specializations) == 0 {
		return nil, errors.New("staff must have at least one specialization")
	}

	now := time.Now()
	return &Staff{
		id:                  id,
		name:                name,
		specializations:     specializations,
		skillLevel:          skillLevel,
		isAvailable:         true,
		currentOrders:       0,
		maxConcurrentOrders: maxConcurrentOrders,
		createdAt:           now,
		updatedAt:           now,
	}, nil
}

// Getters
func (s *Staff) ID() string                  { return s.id }
func (s *Staff) Name() string                { return s.name }
func (s *Staff) Specializations() []StationType { return s.specializations }
func (s *Staff) SkillLevel() float32         { return s.skillLevel }
func (s *Staff) IsAvailable() bool           { return s.isAvailable }
func (s *Staff) CurrentOrders() int32        { return s.currentOrders }
func (s *Staff) MaxConcurrentOrders() int32  { return s.maxConcurrentOrders }
func (s *Staff) CreatedAt() time.Time        { return s.createdAt }
func (s *Staff) UpdatedAt() time.Time        { return s.updatedAt }

// Business Methods

// SetAvailability sets the staff availability
func (s *Staff) SetAvailability(available bool) {
	s.isAvailable = available
	s.updatedAt = time.Now()
}

// AssignOrder assigns an order to the staff member
func (s *Staff) AssignOrder() error {
	if !s.isAvailable {
		return errors.New("staff member is not available")
	}
	if s.currentOrders >= s.maxConcurrentOrders {
		return errors.New("staff member has reached maximum concurrent orders")
	}

	s.currentOrders++
	s.updatedAt = time.Now()
	return nil
}

// CompleteOrder marks an order as completed for this staff member
func (s *Staff) CompleteOrder() error {
	if s.currentOrders <= 0 {
		return errors.New("no orders to complete")
	}

	s.currentOrders--
	s.updatedAt = time.Now()
	return nil
}

// CanHandleStation checks if staff can handle a specific station type
func (s *Staff) CanHandleStation(stationType StationType) bool {
	for _, specialization := range s.specializations {
		if specialization == stationType {
			return true
		}
	}
	return false
}

// CanAcceptOrder checks if staff can accept a new order
func (s *Staff) CanAcceptOrder() bool {
	return s.isAvailable && s.currentOrders < s.maxConcurrentOrders
}

// GetWorkload returns the current workload as a percentage (0.0 to 1.0)
func (s *Staff) GetWorkload() float32 {
	if s.maxConcurrentOrders == 0 {
		return 0.0
	}
	return float32(s.currentOrders) / float32(s.maxConcurrentOrders)
}

// UpdateSkillLevel updates the staff skill level with validation
func (s *Staff) UpdateSkillLevel(skillLevel float32) error {
	if skillLevel < 0.0 || skillLevel > 10.0 {
		return errors.New("skill level must be between 0.0 and 10.0")
	}

	s.skillLevel = skillLevel
	s.updatedAt = time.Now()
	return nil
}

// AddSpecialization adds a new specialization to the staff member
func (s *Staff) AddSpecialization(stationType StationType) error {
	// Check if already specialized
	for _, existing := range s.specializations {
		if existing == stationType {
			return errors.New("staff already has this specialization")
		}
	}

	s.specializations = append(s.specializations, stationType)
	s.updatedAt = time.Now()
	return nil
}

// RemoveSpecialization removes a specialization from the staff member
func (s *Staff) RemoveSpecialization(stationType StationType) error {
	if len(s.specializations) <= 1 {
		return errors.New("staff must have at least one specialization")
	}

	for i, specialization := range s.specializations {
		if specialization == stationType {
			s.specializations = append(s.specializations[:i], s.specializations[i+1:]...)
			s.updatedAt = time.Now()
			return nil
		}
	}

	return errors.New("specialization not found")
}

// GetEfficiencyForStation calculates efficiency for a specific station type
func (s *Staff) GetEfficiencyForStation(stationType StationType) float32 {
	if !s.CanHandleStation(stationType) {
		return 0.0
	}

	// Base efficiency is skill level
	efficiency := s.skillLevel

	// Reduce efficiency based on current workload
	workloadPenalty := s.GetWorkload() * 2.0 // Up to 2 points penalty
	efficiency -= workloadPenalty

	// Ensure efficiency is not negative
	if efficiency < 0.0 {
		efficiency = 0.0
	}

	return efficiency
}

// IsOverloaded checks if staff member is overloaded
func (s *Staff) IsOverloaded() bool {
	return s.GetWorkload() > 0.8 // 80% or more is considered overloaded
}

// ToDTO converts domain entity to data transfer object
func (s *Staff) ToDTO() *StaffDTO {
	return &StaffDTO{
		ID:                  s.id,
		Name:                s.name,
		Specializations:     s.specializations,
		SkillLevel:          s.skillLevel,
		IsAvailable:         s.isAvailable,
		CurrentOrders:       s.currentOrders,
		MaxConcurrentOrders: s.maxConcurrentOrders,
		CreatedAt:           s.createdAt,
		UpdatedAt:           s.updatedAt,
	}
}

// StaffDTO represents staff data transfer object
type StaffDTO struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	Specializations     []StationType `json:"specializations"`
	SkillLevel          float32       `json:"skill_level"`
	IsAvailable         bool          `json:"is_available"`
	CurrentOrders       int32         `json:"current_orders"`
	MaxConcurrentOrders int32         `json:"max_concurrent_orders"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// StaffAllocation represents staff allocation for orders
type StaffAllocation struct {
	Allocations     []*StaffOrderAllocation `json:"allocations"`
	UtilizationRate float32                 `json:"utilization_rate"`
	LoadBalance     map[string]float32      `json:"load_balance"`
	Recommendations []string                `json:"recommendations"`
}

// StaffOrderAllocation represents allocation of a staff member to an order
type StaffOrderAllocation struct {
	StaffID          string        `json:"staff_id"`
	OrderID          string        `json:"order_id"`
	StationType      StationType   `json:"station_type"`
	EstimatedTime    int32         `json:"estimated_time"`
	EfficiencyScore  float32       `json:"efficiency_score"`
	AllocationReason string        `json:"allocation_reason"`
	AllocatedAt      time.Time     `json:"allocated_at"`
}

// NewStaffOrderAllocation creates a new staff order allocation
func NewStaffOrderAllocation(staffID, orderID string, stationType StationType, estimatedTime int32, efficiencyScore float32, reason string) *StaffOrderAllocation {
	return &StaffOrderAllocation{
		StaffID:          staffID,
		OrderID:          orderID,
		StationType:      stationType,
		EstimatedTime:    estimatedTime,
		EfficiencyScore:  efficiencyScore,
		AllocationReason: reason,
		AllocatedAt:      time.Now(),
	}
}
