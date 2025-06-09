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

// RedisEquipmentRepository implements domain.EquipmentRepository using Redis
type RedisEquipmentRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisEquipmentRepository creates a new Redis equipment repository
func NewRedisEquipmentRepository(client *redis.Client, logger *logger.Logger) domain.EquipmentRepository {
	return &RedisEquipmentRepository{
		client: client,
		logger: logger,
	}
}

const (
	equipmentKeyPrefix = "kitchen:equipment:"
	equipmentSetKey    = "kitchen:equipment:all"
	equipmentByType    = "kitchen:equipment:by_type:"
	equipmentByStatus  = "kitchen:equipment:by_status:"
)

// Create saves a new equipment to Redis
func (r *RedisEquipmentRepository) Create(ctx context.Context, equipment *domain.Equipment) error {
	key := equipmentKeyPrefix + equipment.ID()
	
	// Convert equipment to DTO for storage
	dto := equipment.ToDTO()
	data, err := json.Marshal(dto)
	if err != nil {
		r.logger.WithError(err).Error("Failed to marshal equipment")
		return fmt.Errorf("failed to marshal equipment: %w", err)
	}

	pipe := r.client.TxPipeline()
	
	// Store equipment data
	pipe.Set(ctx, key, data, 0)
	
	// Add to equipment set
	pipe.SAdd(ctx, equipmentSetKey, equipment.ID())
	
	// Add to type-based set
	typeKey := equipmentByType + strconv.Itoa(int(equipment.StationType()))
	pipe.SAdd(ctx, typeKey, equipment.ID())
	
	// Add to status-based set
	statusKey := equipmentByStatus + strconv.Itoa(int(equipment.Status()))
	pipe.SAdd(ctx, statusKey, equipment.ID())
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("equipment_id", equipment.ID()).Error("Failed to create equipment")
		return fmt.Errorf("failed to create equipment: %w", err)
	}

	r.logger.WithField("equipment_id", equipment.ID()).Info("Equipment created successfully")
	return nil
}

// GetByID retrieves equipment by ID
func (r *RedisEquipmentRepository) GetByID(ctx context.Context, id string) (*domain.Equipment, error) {
	key := equipmentKeyPrefix + id
	
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("equipment not found: %s", id)
		}
		r.logger.WithError(err).WithField("equipment_id", id).Error("Failed to get equipment")
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	var dto domain.EquipmentDTO
	if err := json.Unmarshal([]byte(data), &dto); err != nil {
		r.logger.WithError(err).Error("Failed to unmarshal equipment")
		return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
	}

	return r.dtoToEquipment(&dto)
}

// Update updates existing equipment
func (r *RedisEquipmentRepository) Update(ctx context.Context, equipment *domain.Equipment) error {
	// Get existing equipment to check for status changes
	existing, err := r.GetByID(ctx, equipment.ID())
	if err != nil {
		return fmt.Errorf("equipment not found: %w", err)
	}

	key := equipmentKeyPrefix + equipment.ID()
	
	// Convert equipment to DTO for storage
	dto := equipment.ToDTO()
	data, err := json.Marshal(dto)
	if err != nil {
		r.logger.WithError(err).Error("Failed to marshal equipment")
		return fmt.Errorf("failed to marshal equipment: %w", err)
	}

	pipe := r.client.TxPipeline()
	
	// Update equipment data
	pipe.Set(ctx, key, data, 0)
	
	// Update status-based sets if status changed
	if existing.Status() != equipment.Status() {
		oldStatusKey := equipmentByStatus + strconv.Itoa(int(existing.Status()))
		newStatusKey := equipmentByStatus + strconv.Itoa(int(equipment.Status()))
		
		pipe.SRem(ctx, oldStatusKey, equipment.ID())
		pipe.SAdd(ctx, newStatusKey, equipment.ID())
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("equipment_id", equipment.ID()).Error("Failed to update equipment")
		return fmt.Errorf("failed to update equipment: %w", err)
	}

	r.logger.WithField("equipment_id", equipment.ID()).Info("Equipment updated successfully")
	return nil
}

// Delete removes equipment from Redis
func (r *RedisEquipmentRepository) Delete(ctx context.Context, id string) error {
	equipment, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	key := equipmentKeyPrefix + id
	
	pipe := r.client.TxPipeline()
	
	// Delete equipment data
	pipe.Del(ctx, key)
	
	// Remove from all sets
	pipe.SRem(ctx, equipmentSetKey, id)
	
	typeKey := equipmentByType + strconv.Itoa(int(equipment.StationType()))
	pipe.SRem(ctx, typeKey, id)
	
	statusKey := equipmentByStatus + strconv.Itoa(int(equipment.Status()))
	pipe.SRem(ctx, statusKey, id)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).WithField("equipment_id", id).Error("Failed to delete equipment")
		return fmt.Errorf("failed to delete equipment: %w", err)
	}

	r.logger.WithField("equipment_id", id).Info("Equipment deleted successfully")
	return nil
}

// GetAll retrieves all equipment
func (r *RedisEquipmentRepository) GetAll(ctx context.Context) ([]*domain.Equipment, error) {
	ids, err := r.client.SMembers(ctx, equipmentSetKey).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get equipment IDs")
		return nil, fmt.Errorf("failed to get equipment IDs: %w", err)
	}

	return r.getEquipmentByIDs(ctx, ids)
}

// GetByStationType retrieves equipment by station type
func (r *RedisEquipmentRepository) GetByStationType(ctx context.Context, stationType domain.StationType) ([]*domain.Equipment, error) {
	typeKey := equipmentByType + strconv.Itoa(int(stationType))
	ids, err := r.client.SMembers(ctx, typeKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("station_type", stationType).Error("Failed to get equipment IDs by type")
		return nil, fmt.Errorf("failed to get equipment IDs by type: %w", err)
	}

	return r.getEquipmentByIDs(ctx, ids)
}

// GetByStatus retrieves equipment by status
func (r *RedisEquipmentRepository) GetByStatus(ctx context.Context, status domain.EquipmentStatus) ([]*domain.Equipment, error) {
	statusKey := equipmentByStatus + strconv.Itoa(int(status))
	ids, err := r.client.SMembers(ctx, statusKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("status", status).Error("Failed to get equipment IDs by status")
		return nil, fmt.Errorf("failed to get equipment IDs by status: %w", err)
	}

	return r.getEquipmentByIDs(ctx, ids)
}

// GetAvailable retrieves available equipment
func (r *RedisEquipmentRepository) GetAvailable(ctx context.Context) ([]*domain.Equipment, error) {
	return r.GetByStatus(ctx, domain.EquipmentStatusAvailable)
}

// GetAvailableByStationType retrieves available equipment by station type
func (r *RedisEquipmentRepository) GetAvailableByStationType(ctx context.Context, stationType domain.StationType) ([]*domain.Equipment, error) {
	typeKey := equipmentByType + strconv.Itoa(int(stationType))
	availableKey := equipmentByStatus + strconv.Itoa(int(domain.EquipmentStatusAvailable))
	
	// Intersect type and available sets
	tempKey := fmt.Sprintf("temp:available_by_type:%d:%d", stationType, time.Now().UnixNano())
	
	pipe := r.client.TxPipeline()
	pipe.SInterStore(ctx, tempKey, typeKey, availableKey)
	pipe.Expire(ctx, tempKey, 10*time.Second) // Expire temp key
	pipe.SMembers(ctx, tempKey)
	
	results, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get available equipment by type")
		return nil, fmt.Errorf("failed to get available equipment by type: %w", err)
	}

	ids := results[2].(*redis.StringSliceCmd).Val()
	return r.getEquipmentByIDs(ctx, ids)
}

// UpdateStatus updates equipment status
func (r *RedisEquipmentRepository) UpdateStatus(ctx context.Context, id string, status domain.EquipmentStatus) error {
	equipment, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := equipment.UpdateStatus(status); err != nil {
		return err
	}

	return r.Update(ctx, equipment)
}

// UpdateLoad updates equipment load
func (r *RedisEquipmentRepository) UpdateLoad(ctx context.Context, id string, currentLoad int32) error {
	equipment, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Calculate load difference and apply
	loadDiff := currentLoad - equipment.CurrentLoad()
	if loadDiff > 0 {
		if err := equipment.AddLoad(loadDiff); err != nil {
			return err
		}
	} else if loadDiff < 0 {
		if err := equipment.RemoveLoad(-loadDiff); err != nil {
			return err
		}
	}

	return r.Update(ctx, equipment)
}

// UpdateEfficiencyScore updates equipment efficiency score
func (r *RedisEquipmentRepository) UpdateEfficiencyScore(ctx context.Context, id string, score float32) error {
	equipment, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := equipment.UpdateEfficiencyScore(score); err != nil {
		return err
	}

	return r.Update(ctx, equipment)
}

// GetNeedingMaintenance retrieves equipment needing maintenance
func (r *RedisEquipmentRepository) GetNeedingMaintenance(ctx context.Context) ([]*domain.Equipment, error) {
	allEquipment, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var needingMaintenance []*domain.Equipment
	for _, equipment := range allEquipment {
		if equipment.NeedsMaintenance() {
			needingMaintenance = append(needingMaintenance, equipment)
		}
	}

	return needingMaintenance, nil
}

// GetOverloaded retrieves overloaded equipment
func (r *RedisEquipmentRepository) GetOverloaded(ctx context.Context) ([]*domain.Equipment, error) {
	allEquipment, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var overloaded []*domain.Equipment
	for _, equipment := range allEquipment {
		if equipment.GetUtilizationRate() >= 1.0 {
			overloaded = append(overloaded, equipment)
		}
	}

	return overloaded, nil
}

// GetUtilizationStats returns utilization statistics
func (r *RedisEquipmentRepository) GetUtilizationStats(ctx context.Context) (map[string]float32, error) {
	allEquipment, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]float32)
	for _, equipment := range allEquipment {
		stats[equipment.ID()] = equipment.GetUtilizationRate()
	}

	return stats, nil
}

// GetEfficiencyStats returns efficiency statistics
func (r *RedisEquipmentRepository) GetEfficiencyStats(ctx context.Context) (map[string]float32, error) {
	allEquipment, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]float32)
	for _, equipment := range allEquipment {
		stats[equipment.ID()] = equipment.EfficiencyScore()
	}

	return stats, nil
}

// Helper methods

func (r *RedisEquipmentRepository) getEquipmentByIDs(ctx context.Context, ids []string) ([]*domain.Equipment, error) {
	if len(ids) == 0 {
		return []*domain.Equipment{}, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = equipmentKeyPrefix + id
	}

	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get equipment data")
		return nil, fmt.Errorf("failed to get equipment data: %w", err)
	}

	equipment := make([]*domain.Equipment, 0, len(results))
	for i, result := range results {
		if result == nil {
			r.logger.WithField("equipment_id", ids[i]).Warn("Equipment data not found")
			continue
		}

		var dto domain.EquipmentDTO
		if err := json.Unmarshal([]byte(result.(string)), &dto); err != nil {
			r.logger.WithError(err).WithField("equipment_id", ids[i]).Error("Failed to unmarshal equipment")
			continue
		}

		eq, err := r.dtoToEquipment(&dto)
		if err != nil {
			r.logger.WithError(err).WithField("equipment_id", ids[i]).Error("Failed to convert DTO to equipment")
			continue
		}

		equipment = append(equipment, eq)
	}

	return equipment, nil
}

func (r *RedisEquipmentRepository) dtoToEquipment(dto *domain.EquipmentDTO) (*domain.Equipment, error) {
	equipment, err := domain.NewEquipment(dto.ID, dto.Name, dto.StationType, dto.MaxCapacity)
	if err != nil {
		return nil, err
	}

	// Set additional fields that can't be set through constructor
	equipment.UpdateStatus(dto.Status)
	equipment.UpdateEfficiencyScore(dto.EfficiencyScore)
	
	// Add/remove load to match current load
	if dto.CurrentLoad > 0 {
		equipment.AddLoad(dto.CurrentLoad)
	}

	return equipment, nil
}
