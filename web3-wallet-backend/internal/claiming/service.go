package claiming

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourusername/web3-wallet-backend/pkg/kafka"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
	"github.com/yourusername/web3-wallet-backend/pkg/redis"
)

// Service represents the claiming service
type Service struct {
	repo      Repository
	cache     redis.Client
	producer  kafka.Producer
	logger    *logger.Logger
	cacheTTL  time.Duration
}

// NewService creates a new claiming service
func NewService(repo Repository, cache redis.Client, producer kafka.Producer, logger *logger.Logger) *Service {
	return &Service{
		repo:      repo,
		cache:     cache,
		producer:  producer,
		logger:    logger.Named("claiming-service"),
		cacheTTL:  time.Hour,
	}
}

// ClaimOrder claims an order
func (s *Service) ClaimOrder(ctx context.Context, orderID string, userID string) (*Claim, error) {
	// Check if order is already claimed
	cacheKey := fmt.Sprintf("claim:order:%s", orderID)
	exists, err := s.cache.Exists(ctx, cacheKey)
	if err == nil && exists {
		// Get the claim from cache
		data, err := s.cache.Get(ctx, cacheKey)
		if err == nil {
			var claim Claim
			if err := json.Unmarshal([]byte(data), &claim); err == nil {
				return nil, fmt.Errorf("order already claimed by user %s", claim.UserID)
			}
		}
		
		// If we can't get the claim from cache, check the database
		claim, err := s.repo.GetClaimByOrderID(ctx, orderID)
		if err == nil && claim != nil {
			return nil, fmt.Errorf("order already claimed by user %s", claim.UserID)
		}
	}

	// Try to claim the order
	claim, err := s.repo.ClaimOrder(ctx, orderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to claim order: %w", err)
	}

	// Cache the claim
	data, err := json.Marshal(claim)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache claim", "orderID", orderID, "error", err)
		}
		
		// Also cache by claim ID
		claimKey := fmt.Sprintf("claim:%s", claim.ID)
		if err := s.cache.Set(ctx, claimKey, data, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache claim by ID", "claimID", claim.ID, "error", err)
		}
	}

	// Publish event
	event := OrderClaimedEvent{
		ClaimID:   claim.ID,
		OrderID:   orderID,
		UserID:    userID,
		Status:    claim.Status,
		ClaimedAt: claim.ClaimedAt,
	}
	eventData, err := json.Marshal(event)
	if err == nil {
		if err := s.producer.Produce("claim-events", []byte(claim.ID), eventData); err != nil {
			s.logger.Warn("Failed to publish order claimed event", "claimID", claim.ID, "error", err)
		}
	}

	return claim, nil
}

// GetClaim gets a claim by ID
func (s *Service) GetClaim(ctx context.Context, id string) (*Claim, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("claim:%s", id)
	data, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var claim Claim
		if err := json.Unmarshal([]byte(data), &claim); err == nil {
			s.logger.Debug("Claim cache hit", "id", id)
			return &claim, nil
		}
	}

	// Cache miss, get from database
	claim, err := s.repo.GetClaim(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}

	// Cache the result
	if claim != nil {
		data, err := json.Marshal(claim)
		if err == nil {
			if err := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); err != nil {
				s.logger.Warn("Failed to cache claim", "id", id, "error", err)
			}
			
			// Also cache by order ID
			orderKey := fmt.Sprintf("claim:order:%s", claim.OrderID)
			if err := s.cache.Set(ctx, orderKey, data, s.cacheTTL); err != nil {
				s.logger.Warn("Failed to cache claim by order ID", "orderID", claim.OrderID, "error", err)
			}
		}
	}

	return claim, nil
}

// ProcessClaim processes a claim
func (s *Service) ProcessClaim(ctx context.Context, id string, status string) (*Claim, error) {
	// Get the claim
	claim, err := s.GetClaim(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}

	if claim == nil {
		return nil, fmt.Errorf("claim not found")
	}

	// Update the claim
	claim.Status = status
	claim.ProcessedAt = time.Now()

	// Update in database
	if err := s.repo.UpdateClaim(ctx, id, claim); err != nil {
		return nil, fmt.Errorf("failed to update claim: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("claim:%s", id)
	data, err := json.Marshal(claim)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache claim", "id", id, "error", err)
		}
		
		// Also update by order ID
		orderKey := fmt.Sprintf("claim:order:%s", claim.OrderID)
		if err := s.cache.Set(ctx, orderKey, data, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache claim by order ID", "orderID", claim.OrderID, "error", err)
		}
	}

	// Publish event
	event := ClaimProcessedEvent{
		ClaimID:     claim.ID,
		OrderID:     claim.OrderID,
		UserID:      claim.UserID,
		Status:      claim.Status,
		ProcessedAt: claim.ProcessedAt,
	}
	eventData, err := json.Marshal(event)
	if err == nil {
		if err := s.producer.Produce("claim-events", []byte(claim.ID), eventData); err != nil {
			s.logger.Warn("Failed to publish claim processed event", "claimID", claim.ID, "error", err)
		}
	}

	return claim, nil
}

// ListClaims lists claims
func (s *Service) ListClaims(ctx context.Context, userID, orderID, status string, page, pageSize int) ([]*Claim, int, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("claims:user:%s:order:%s:status:%s:page:%d:pageSize:%d", 
		userID, orderID, status, page, pageSize)
	
	data, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var result struct {
			Claims []*Claim `json:"claims"`
			Total  int      `json:"total"`
		}
		if err := json.Unmarshal([]byte(data), &result); err == nil {
			s.logger.Debug("Claims cache hit", "key", cacheKey)
			return result.Claims, result.Total, nil
		}
	}

	// Cache miss, get from database
	claims, total, err := s.repo.ListClaims(ctx, userID, orderID, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list claims: %w", err)
	}

	// Cache the result
	result := struct {
		Claims []*Claim `json:"claims"`
		Total  int      `json:"total"`
	}{
		Claims: claims,
		Total:  total,
	}
	
	data, err = json.Marshal(result)
	if err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache claims", "key", cacheKey, "error", err)
		}
	}

	return claims, total, nil
}
