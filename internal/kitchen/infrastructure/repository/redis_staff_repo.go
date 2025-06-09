package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RedisStaffRepository implements domain.StaffRepository using Redis
type RedisStaffRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisStaffRepository creates a new Redis staff repository
func NewRedisStaffRepository(client *redis.Client, logger *logger.Logger) domain.StaffRepository {
	return &RedisStaffRepository{
		client: client,
		logger: logger,
	}
}

const (
	staffKeyPrefix      = "kitchen:staff:"
	staffSetKey         = "kitchen:staff:all"
	staffBySpecialization = "kitchen:staff:by_specialization:"
	staffAvailableKey   = "kitchen:staff:available"
)

// Create saves a new staff member to Redis
func (r *RedisStaffRepository) Create(ctx context.Context, staff *domain.Staff) error {
	key := staffKeyPrefix + staff.ID()
	
	// Convert staff to DTO for storage
	dto := staff.ToDTO()
	data, err := json.Marshal(dto)
	if err != nil {
		r.logger.WithError(err).Error("Failed to marshal staff")
		return fmt.Errorf("failed to marshal staff: %w", err)
	}

	pipe := r.client.TxPipeline()
	
	// Store staff data
	pipe.Set(ctx, key, data, 0)
	
	// Add to staff set
	pipe.SAdd(ctx, staffSetKey, staff.ID())
	
	// Add to specialization-based sets
	for _, specialization := range staff.Specializations() {
		specKey := staffBySpecialization + strconv.Itoa(int(specialization))
		pipe.SAdd(ctx, specKey, staff.ID())
	}
	
	// Add to available set if available
	if staff.IsAvailable() {
		pipe.SAdd(ctx, staffAvailableKey, staff.ID())
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("staff_id", staff.ID()).Error("Failed to create staff")
		return fmt.Errorf("failed to create staff: %w", err)
	}

	r.logger.WithField("staff_id", staff.ID()).Info("Staff created successfully")
	return nil
}

// GetByID retrieves staff by ID
func (r *RedisStaffRepository) GetByID(ctx context.Context, id string) (*domain.Staff, error) {
	key := staffKeyPrefix + id
	
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("staff not found: %s", id)
		}
		r.logger.WithError(err).WithField("staff_id", id).Error("Failed to get staff")
		return nil, fmt.Errorf("failed to get staff: %w", err)
	}

	var dto domain.StaffDTO
	if err := json.Unmarshal([]byte(data), &dto); err != nil {
		r.logger.WithError(err).Error("Failed to unmarshal staff")
		return nil, fmt.Errorf("failed to unmarshal staff: %w", err)
	}

	return r.dtoToStaff(&dto)
}

// Update updates existing staff
func (r *RedisStaffRepository) Update(ctx context.Context, staff *domain.Staff) error {
	// Get existing staff to check for availability changes
	existing, err := r.GetByID(ctx, staff.ID())
	if err != nil {
		return fmt.Errorf("staff not found: %w", err)
	}

	key := staffKeyPrefix + staff.ID()
	
	// Convert staff to DTO for storage
	dto := staff.ToDTO()
	data, err := json.Marshal(dto)
	if err != nil {
		r.logger.WithError(err).Error("Failed to marshal staff")
		return fmt.Errorf("failed to marshal staff: %w", err)
	}

	pipe := r.client.TxPipeline()
	
	// Update staff data
	pipe.Set(ctx, key, data, 0)
	
	// Update availability set if availability changed
	if existing.IsAvailable() != staff.IsAvailable() {
		if staff.IsAvailable() {
			pipe.SAdd(ctx, staffAvailableKey, staff.ID())
		} else {
			pipe.SRem(ctx, staffAvailableKey, staff.ID())
		}
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("staff_id", staff.ID()).Error("Failed to update staff")
		return fmt.Errorf("failed to update staff: %w", err)
	}

	r.logger.WithField("staff_id", staff.ID()).Info("Staff updated successfully")
	return nil
}

// Delete removes staff from Redis
func (r *RedisStaffRepository) Delete(ctx context.Context, id string) error {
	staff, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	key := staffKeyPrefix + id
	
	pipe := r.client.TxPipeline()
	
	// Delete staff data
	pipe.Del(ctx, key)
	
	// Remove from all sets
	pipe.SRem(ctx, staffSetKey, id)
	pipe.SRem(ctx, staffAvailableKey, id)
	
	// Remove from specialization sets
	for _, specialization := range staff.Specializations() {
		specKey := staffBySpecialization + strconv.Itoa(int(specialization))
		pipe.SRem(ctx, specKey, id)
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("staff_id", id).Error("Failed to delete staff")
		return fmt.Errorf("failed to delete staff: %w", err)
	}

	r.logger.WithField("staff_id", id).Info("Staff deleted successfully")
	return nil
}

// GetAll retrieves all staff
func (r *RedisStaffRepository) GetAll(ctx context.Context) ([]*domain.Staff, error) {
	ids, err := r.client.SMembers(ctx, staffSetKey).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get staff IDs")
		return nil, fmt.Errorf("failed to get staff IDs: %w", err)
	}

	return r.getStaffByIDs(ctx, ids)
}

// GetAvailable retrieves available staff
func (r *RedisStaffRepository) GetAvailable(ctx context.Context) ([]*domain.Staff, error) {
	ids, err := r.client.SMembers(ctx, staffAvailableKey).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get available staff IDs")
		return nil, fmt.Errorf("failed to get available staff IDs: %w", err)
	}

	return r.getStaffByIDs(ctx, ids)
}

// GetBySpecialization retrieves staff by specialization
func (r *RedisStaffRepository) GetBySpecialization(ctx context.Context, stationType domain.StationType) ([]*domain.Staff, error) {
	specKey := staffBySpecialization + strconv.Itoa(int(stationType))
	ids, err := r.client.SMembers(ctx, specKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("specialization", stationType).Error("Failed to get staff IDs by specialization")
		return nil, fmt.Errorf("failed to get staff IDs by specialization: %w", err)
	}

	return r.getStaffByIDs(ctx, ids)
}

// GetAvailableBySpecialization retrieves available staff by specialization
func (r *RedisStaffRepository) GetAvailableBySpecialization(ctx context.Context, stationType domain.StationType) ([]*domain.Staff, error) {
	specKey := staffBySpecialization + strconv.Itoa(int(stationType))
	
	// Intersect specialization and available sets
	tempKey := fmt.Sprintf("temp:available_by_spec:%d:%d", stationType, time.Now().UnixNano())
	
	pipe := r.client.TxPipeline()
	pipe.SInterStore(ctx, tempKey, specKey, staffAvailableKey)
	pipe.Expire(ctx, tempKey, 10*time.Second) // Expire temp key
	pipe.SMembers(ctx, tempKey)
	
	results, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get available staff by specialization")
		return nil, fmt.Errorf("failed to get available staff by specialization: %w", err)
	}

	ids := results[2].(*redis.StringSliceCmd).Val()
	return r.getStaffByIDs(ctx, ids)
}

// UpdateAvailability updates staff availability
func (r *RedisStaffRepository) UpdateAvailability(ctx context.Context, id string, available bool) error {
	staff, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	staff.SetAvailability(available)
	return r.Update(ctx, staff)
}

// UpdateCurrentOrders updates staff current orders count
func (r *RedisStaffRepository) UpdateCurrentOrders(ctx context.Context, id string, currentOrders int32) error {
	staff, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Calculate order difference and apply
	orderDiff := currentOrders - staff.CurrentOrders()
	if orderDiff > 0 {
		for i := int32(0); i < orderDiff; i++ {
			if err := staff.AssignOrder(); err != nil {
				return err
			}
		}
	} else if orderDiff < 0 {
		for i := int32(0); i < -orderDiff; i++ {
			if err := staff.CompleteOrder(); err != nil {
				return err
			}
		}
	}

	return r.Update(ctx, staff)
}

// UpdateSkillLevel updates staff skill level
func (r *RedisStaffRepository) UpdateSkillLevel(ctx context.Context, id string, skillLevel float32) error {
	staff, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := staff.UpdateSkillLevel(skillLevel); err != nil {
		return err
	}

	return r.Update(ctx, staff)
}

// GetOverloaded retrieves overloaded staff
func (r *RedisStaffRepository) GetOverloaded(ctx context.Context) ([]*domain.Staff, error) {
	allStaff, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var overloaded []*domain.Staff
	for _, staff := range allStaff {
		if staff.IsOverloaded() {
			overloaded = append(overloaded, staff)
		}
	}

	return overloaded, nil
}

// GetWorkloadStats returns workload statistics
func (r *RedisStaffRepository) GetWorkloadStats(ctx context.Context) (map[string]float32, error) {
	allStaff, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]float32)
	for _, staff := range allStaff {
		stats[staff.ID()] = staff.GetWorkload()
	}

	return stats, nil
}

// GetSkillStats returns skill statistics
func (r *RedisStaffRepository) GetSkillStats(ctx context.Context) (map[string]float32, error) {
	allStaff, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]float32)
	for _, staff := range allStaff {
		stats[staff.ID()] = staff.SkillLevel()
	}

	return stats, nil
}

// Helper methods

func (r *RedisStaffRepository) getStaffByIDs(ctx context.Context, ids []string) ([]*domain.Staff, error) {
	if len(ids) == 0 {
		return []*domain.Staff{}, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = staffKeyPrefix + id
	}

	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get staff data")
		return nil, fmt.Errorf("failed to get staff data: %w", err)
	}

	staff := make([]*domain.Staff, 0, len(results))
	for i, result := range results {
		if result == nil {
			r.logger.WithField("staff_id", ids[i]).Warn("Staff data not found")
			continue
		}

		var dto domain.StaffDTO
		if err := json.Unmarshal([]byte(result.(string)), &dto); err != nil {
			r.logger.WithError(err).WithField("staff_id", ids[i]).Error("Failed to unmarshal staff")
			continue
		}

		st, err := r.dtoToStaff(&dto)
		if err != nil {
			r.logger.WithError(err).WithField("staff_id", ids[i]).Error("Failed to convert DTO to staff")
			continue
		}

		staff = append(staff, st)
	}

	return staff, nil
}

func (r *RedisStaffRepository) dtoToStaff(dto *domain.StaffDTO) (*domain.Staff, error) {
	staff, err := domain.NewStaff(dto.ID, dto.Name, dto.Specializations, dto.SkillLevel, dto.MaxConcurrentOrders)
	if err != nil {
		return nil, err
	}

	// Set additional fields that can't be set through constructor
	staff.SetAvailability(dto.IsAvailable)
	
	// Assign orders to match current count
	for i := int32(0); i < dto.CurrentOrders; i++ {
		staff.AssignOrder()
	}

	return staff, nil
}
