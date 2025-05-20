package supply

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourusername/web3-wallet-backend/pkg/kafka"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
	"github.com/yourusername/web3-wallet-backend/pkg/redis"
)

// Service represents the supply service
type Service struct {
	repo      Repository
	cache     redis.Client
	producer  kafka.Producer
	logger    *logger.Logger
	cacheTTL  time.Duration
}

// NewService creates a new supply service
func NewService(repo Repository, cache redis.Client, producer kafka.Producer, logger *logger.Logger) *Service {
	return &Service{
		repo:      repo,
		cache:     cache,
		producer:  producer,
		logger:    logger.Named("supply-service"),
		cacheTTL:  time.Hour,
	}
}

// GetSupply gets supply by ID
func (s *Service) GetSupply(ctx context.Context, id string) (*Supply, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("supply:%s", id)
	data, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var supply Supply
		if err := json.Unmarshal([]byte(data), &supply); err == nil {
			s.logger.Debug("Supply cache hit", "id", id)
			return &supply, nil
		}
	}

	// Cache miss, get from database
	supply, err := s.repo.GetSupply(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get supply: %w", err)
	}

	// Cache the result
	if supply != nil {
		data, err := json.Marshal(supply)
		if err == nil {
			if err := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); err != nil {
				s.logger.Warn("Failed to cache supply", "id", id, "error", err)
			}
		}
	}

	return supply, nil
}

// CreateSupply creates a new supply
func (s *Service) CreateSupply(ctx context.Context, supply *Supply) error {
	// Create in database
	if err := s.repo.CreateSupply(ctx, supply); err != nil {
		return fmt.Errorf("failed to create supply: %w", err)
	}

	// Cache the result
	cacheKey := fmt.Sprintf("supply:%s", supply.ID)
	data, err := json.Marshal(supply)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache supply", "id", supply.ID, "error", err)
		}
	}

	// Publish event
	event := SupplyCreatedEvent{
		ID:        supply.ID,
		UserID:    supply.UserID,
		Currency:  supply.Currency,
		Amount:    supply.Amount,
		Status:    supply.Status,
		CreatedAt: supply.CreatedAt,
	}
	eventData, err := json.Marshal(event)
	if err == nil {
		if err := s.producer.Produce("supply-events", []byte(supply.ID), eventData); err != nil {
			s.logger.Warn("Failed to publish supply created event", "id", supply.ID, "error", err)
		}
	}

	return nil
}

// UpdateSupply updates a supply
func (s *Service) UpdateSupply(ctx context.Context, id string, supply *Supply) error {
	// Update in database
	if err := s.repo.UpdateSupply(ctx, id, supply); err != nil {
		return fmt.Errorf("failed to update supply: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("supply:%s", id)
	data, err := json.Marshal(supply)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache supply", "id", id, "error", err)
		}
	}

	// Publish event
	event := SupplyUpdatedEvent{
		ID:        supply.ID,
		UserID:    supply.UserID,
		Currency:  supply.Currency,
		Amount:    supply.Amount,
		Status:    supply.Status,
		UpdatedAt: supply.UpdatedAt,
	}
	eventData, err := json.Marshal(event)
	if err == nil {
		if err := s.producer.Produce("supply-events", []byte(supply.ID), eventData); err != nil {
			s.logger.Warn("Failed to publish supply updated event", "id", supply.ID, "error", err)
		}
	}

	return nil
}

// DeleteSupply deletes a supply
func (s *Service) DeleteSupply(ctx context.Context, id string) error {
	// Delete from database
	if err := s.repo.DeleteSupply(ctx, id); err != nil {
		return fmt.Errorf("failed to delete supply: %w", err)
	}

	// Delete from cache
	cacheKey := fmt.Sprintf("supply:%s", id)
	if err := s.cache.Del(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to delete supply from cache", "id", id, "error", err)
	}

	// Publish event
	event := SupplyDeletedEvent{
		ID:        id,
		DeletedAt: time.Now(),
	}
	eventData, err := json.Marshal(event)
	if err == nil {
		if err := s.producer.Produce("supply-events", []byte(id), eventData); err != nil {
			s.logger.Warn("Failed to publish supply deleted event", "id", id, "error", err)
		}
	}

	return nil
}

// ListSupplies lists supplies
func (s *Service) ListSupplies(ctx context.Context, userID, currency, status string, page, pageSize int) ([]*Supply, int, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("supplies:user:%s:currency:%s:status:%s:page:%d:pageSize:%d", 
		userID, currency, status, page, pageSize)
	
	data, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var result struct {
			Supplies []*Supply `json:"supplies"`
			Total    int       `json:"total"`
		}
		if err := json.Unmarshal([]byte(data), &result); err == nil {
			s.logger.Debug("Supplies cache hit", "key", cacheKey)
			return result.Supplies, result.Total, nil
		}
	}

	// Cache miss, get from database
	supplies, total, err := s.repo.ListSupplies(ctx, userID, currency, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list supplies: %w", err)
	}

	// Cache the result
	result := struct {
		Supplies []*Supply `json:"supplies"`
		Total    int       `json:"total"`
	}{
		Supplies: supplies,
		Total:    total,
	}
	
	data, err = json.Marshal(result)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache supplies", "key", cacheKey, "error", err)
		}
	}

	return supplies, total, nil
}
