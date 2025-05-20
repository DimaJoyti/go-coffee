package failover

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/web3-wallet-backend/pkg/config"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
)

// Service represents a failover service
type Service struct {
	config      *config.MultiRegionConfig
	logger      *logger.Logger
	healthCheck HealthCheck
	mutex       sync.RWMutex
	status      map[string]RegionStatus
	stopCh      chan struct{}
}

// RegionStatus represents the status of a region
type RegionStatus struct {
	Healthy           bool
	FailureCount      int
	RecoveryCount     int
	LastCheck         time.Time
	LastStatusChange  time.Time
	ConsecutiveErrors int
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context, region *config.Region) error

// NewService creates a new failover service
func NewService(config *config.MultiRegionConfig, logger *logger.Logger, healthCheck HealthCheck) *Service {
	status := make(map[string]RegionStatus)
	for _, region := range config.Regions {
		status[region.Name] = RegionStatus{
			Healthy:           true,
			FailureCount:      0,
			RecoveryCount:     0,
			LastCheck:         time.Now(),
			LastStatusChange:  time.Now(),
			ConsecutiveErrors: 0,
		}
	}

	return &Service{
		config:      config,
		logger:      logger.Named("failover"),
		healthCheck: healthCheck,
		mutex:       sync.RWMutex{},
		status:      status,
		stopCh:      make(chan struct{}),
	}
}

// Start starts the failover service
func (s *Service) Start() {
	if !s.config.Failover.Enabled {
		s.logger.Info("Failover is disabled")
		return
	}

	s.logger.Info("Starting failover service")
	go s.run()
}

// Stop stops the failover service
func (s *Service) Stop() {
	if !s.config.Failover.Enabled {
		return
	}

	s.logger.Info("Stopping failover service")
	close(s.stopCh)
}

// GetActiveRegion gets the active region
func (s *Service) GetActiveRegion() *config.Region {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if current region is healthy
	currentRegion := s.config.GetCurrentRegion()
	if currentRegion != nil {
		status, ok := s.status[currentRegion.Name]
		if ok && status.Healthy {
			return currentRegion
		}
	}

	// Find the highest priority healthy region
	var activeRegion *config.Region
	for _, region := range s.config.Regions {
		status, ok := s.status[region.Name]
		if ok && status.Healthy {
			if activeRegion == nil || region.Priority < activeRegion.Priority {
				activeRegion = &region
			}
		}
	}

	return activeRegion
}

// run runs the failover service
func (s *Service) run() {
	ticker := time.NewTicker(s.config.Failover.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkRegions()
		case <-s.stopCh:
			return
		}
	}
}

// checkRegions checks the health of all regions
func (s *Service) checkRegions() {
	for _, region := range s.config.Regions {
		s.checkRegion(&region)
	}
}

// checkRegion checks the health of a region
func (s *Service) checkRegion(region *config.Region) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Failover.Timeout)
	defer cancel()

	err := s.healthCheck(ctx, region)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	status := s.status[region.Name]
	status.LastCheck = time.Now()

	if err != nil {
		s.logger.Warn("Region health check failed", "region", region.Name, "error", err)
		status.ConsecutiveErrors++
		if status.Healthy && status.ConsecutiveErrors >= s.config.Failover.FailureThreshold {
			status.Healthy = false
			status.LastStatusChange = time.Now()
			status.FailureCount++
			status.RecoveryCount = 0
			s.logger.Error("Region is now unhealthy", "region", region.Name)
		}
	} else {
		status.ConsecutiveErrors = 0
		if !status.Healthy {
			status.RecoveryCount++
			if status.RecoveryCount >= s.config.Failover.RecoveryThreshold {
				status.Healthy = true
				status.LastStatusChange = time.Now()
				status.RecoveryCount = 0
				status.FailureCount = 0
				s.logger.Info("Region is now healthy", "region", region.Name)
			}
		}
	}

	s.status[region.Name] = status
}
