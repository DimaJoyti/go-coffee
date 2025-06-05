package risk

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// RiskChecker interface for risk validation
type RiskChecker interface {
	ValidateOrder(ctx context.Context, order *models.Order) error
	ValidatePosition(ctx context.Context, position *models.Position) error
	CheckExposure(ctx context.Context, strategyID string) error
	CheckDrawdown(ctx context.Context, strategyID string) error
}

// Service manages risk controls and monitoring
type Service struct {
	config      *config.Config
	db          *sql.DB
	redis       *redis.Client
	riskChecker RiskChecker
	
	// Risk monitoring
	riskEvents     map[string]*models.RiskEvent
	exposureMap    map[string]decimal.Decimal
	drawdownMap    map[string]decimal.Decimal
	
	// Event channels
	riskEventChan  chan *models.RiskEvent
	violationChan  chan *models.RiskEvent
	
	// State management
	isRunning bool
	mu        sync.RWMutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
	
	// Performance metrics
	totalChecks     uint64
	violations      uint64
	blockedOrders   uint64
	riskScore       decimal.Decimal
}

// NewService creates a new risk management service
func NewService(cfg *config.Config, db *sql.DB, redis *redis.Client) (*Service, error) {
	s := &Service{
		config:        cfg,
		db:            db,
		redis:         redis,
		riskEvents:    make(map[string]*models.RiskEvent),
		exposureMap:   make(map[string]decimal.Decimal),
		drawdownMap:   make(map[string]decimal.Decimal),
		riskEventChan: make(chan *models.RiskEvent, 1000),
		violationChan: make(chan *models.RiskEvent, 1000),
		stopChan:      make(chan struct{}),
	}

	// Initialize risk checker
	s.riskChecker = NewRiskChecker(cfg, db, redis)

	return s, nil
}

// Start starts the risk management service
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("risk management service is already running")
	}

	logrus.Info("Starting HFT Risk Management Service")

	// Start monitoring goroutines
	s.wg.Add(3)
	go s.processRiskEvents(ctx)
	go s.monitorExposure(ctx)
	go s.monitorDrawdown(ctx)

	s.isRunning = true
	logrus.Info("HFT Risk Management Service started successfully")

	return nil
}

// Stop stops the risk management service
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	logrus.Info("Stopping HFT Risk Management Service")

	// Signal stop
	close(s.stopChan)

	// Wait for goroutines to finish
	s.wg.Wait()

	s.isRunning = false
	logrus.Info("HFT Risk Management Service stopped")

	return nil
}

// ValidateOrder validates an order against risk limits
func (s *Service) ValidateOrder(ctx context.Context, order *models.Order) error {
	s.mu.Lock()
	s.totalChecks++
	s.mu.Unlock()

	// Pre-trade risk checks
	if err := s.riskChecker.ValidateOrder(ctx, order); err != nil {
		s.mu.Lock()
		s.violations++
		s.blockedOrders++
		s.mu.Unlock()

		// Create risk event
		riskEvent := &models.RiskEvent{
			ID:          fmt.Sprintf("risk_%d", time.Now().UnixNano()),
			Type:        "order_validation",
			Severity:    "high",
			StrategyID:  order.StrategyID,
			Symbol:      order.Symbol,
			Description: fmt.Sprintf("Order validation failed: %s", err.Error()),
			Data: map[string]any{
				"order_id": order.ID,
				"symbol":   order.Symbol,
				"quantity": order.Quantity,
				"price":    order.Price,
				"error":    err.Error(),
			},
			Action:    "block_order",
			Resolved:  false,
			CreatedAt: time.Now(),
		}

		s.emitRiskEvent(riskEvent)
		return fmt.Errorf("order blocked by risk management: %w", err)
	}

	return nil
}

// ValidatePosition validates a position against risk limits
func (s *Service) ValidatePosition(ctx context.Context, position *models.Position) error {
	s.mu.Lock()
	s.totalChecks++
	s.mu.Unlock()

	if err := s.riskChecker.ValidatePosition(ctx, position); err != nil {
		s.mu.Lock()
		s.violations++
		s.mu.Unlock()

		// Create risk event
		riskEvent := &models.RiskEvent{
			ID:          fmt.Sprintf("risk_%d", time.Now().UnixNano()),
			Type:        "position_validation",
			Severity:    "medium",
			StrategyID:  position.StrategyID,
			Symbol:      position.Symbol,
			Description: fmt.Sprintf("Position validation failed: %s", err.Error()),
			Data: map[string]any{
				"position_id": position.ID,
				"symbol":      position.Symbol,
				"size":        position.Size,
				"pnl":         position.UnrealizedPnL,
				"error":       err.Error(),
			},
			Action:    "monitor_position",
			Resolved:  false,
			CreatedAt: time.Now(),
		}

		s.emitRiskEvent(riskEvent)
		return err
	}

	return nil
}

// CheckExposure checks total exposure for a strategy
func (s *Service) CheckExposure(ctx context.Context, strategyID string) error {
	if err := s.riskChecker.CheckExposure(ctx, strategyID); err != nil {
		s.mu.Lock()
		s.violations++
		s.mu.Unlock()

		// Create risk event
		riskEvent := &models.RiskEvent{
			ID:          fmt.Sprintf("risk_%d", time.Now().UnixNano()),
			Type:        "exposure_limit",
			Severity:    "high",
			StrategyID:  strategyID,
			Description: fmt.Sprintf("Exposure limit exceeded: %s", err.Error()),
			Data: map[string]any{
				"strategy_id": strategyID,
				"error":       err.Error(),
			},
			Action:    "reduce_exposure",
			Resolved:  false,
			CreatedAt: time.Now(),
		}

		s.emitRiskEvent(riskEvent)
		return err
	}

	return nil
}

// CheckDrawdown checks drawdown limits for a strategy
func (s *Service) CheckDrawdown(ctx context.Context, strategyID string) error {
	if err := s.riskChecker.CheckDrawdown(ctx, strategyID); err != nil {
		s.mu.Lock()
		s.violations++
		s.mu.Unlock()

		// Create risk event
		riskEvent := &models.RiskEvent{
			ID:          fmt.Sprintf("risk_%d", time.Now().UnixNano()),
			Type:        "drawdown_limit",
			Severity:    "critical",
			StrategyID:  strategyID,
			Description: fmt.Sprintf("Drawdown limit exceeded: %s", err.Error()),
			Data: map[string]any{
				"strategy_id": strategyID,
				"error":       err.Error(),
			},
			Action:    "stop_strategy",
			Resolved:  false,
			CreatedAt: time.Now(),
		}

		s.emitRiskEvent(riskEvent)
		return err
	}

	return nil
}

// GetRiskEvents returns recent risk events
func (s *Service) GetRiskEvents(limit int) []*models.RiskEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*models.RiskEvent
	count := 0
	for _, event := range s.riskEvents {
		if count >= limit {
			break
		}
		events = append(events, event)
		count++
	}

	return events
}

// GetRiskEventChannel returns the risk event channel
func (s *Service) GetRiskEventChannel() <-chan *models.RiskEvent {
	return s.riskEventChan
}

// GetViolationChannel returns the violation channel
func (s *Service) GetViolationChannel() <-chan *models.RiskEvent {
	return s.violationChan
}

// IsHealthy returns the health status of the service
func (s *Service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetMetrics returns risk management metrics
func (s *Service) GetMetrics() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	violationRate := decimal.Zero
	if s.totalChecks > 0 {
		violationRate = decimal.NewFromInt(int64(s.violations)).Div(decimal.NewFromInt(int64(s.totalChecks)))
	}

	return map[string]any{
		"total_checks":    s.totalChecks,
		"violations":      s.violations,
		"blocked_orders":  s.blockedOrders,
		"violation_rate":  violationRate,
		"risk_score":      s.riskScore,
		"active_events":   len(s.riskEvents),
	}
}

// emitRiskEvent emits a risk event
func (s *Service) emitRiskEvent(event *models.RiskEvent) {
	// Store event
	s.mu.Lock()
	s.riskEvents[event.ID] = event
	s.mu.Unlock()

	// Send to channels
	select {
	case s.riskEventChan <- event:
	default:
		// Channel is full, skip
	}

	if event.Severity == "high" || event.Severity == "critical" {
		select {
		case s.violationChan <- event:
		default:
			// Channel is full, skip
		}
	}

	logrus.WithFields(logrus.Fields{
		"event_id":    event.ID,
		"type":        event.Type,
		"severity":    event.Severity,
		"strategy_id": event.StrategyID,
		"description": event.Description,
	}).Warn("Risk event generated")
}

// processRiskEvents processes risk events
func (s *Service) processRiskEvents(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case event := <-s.riskEventChan:
			s.handleRiskEvent(event)
		}
	}
}

// monitorExposure monitors exposure levels
func (s *Service) monitorExposure(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAllExposures(ctx)
		}
	}
}

// monitorDrawdown monitors drawdown levels
func (s *Service) monitorDrawdown(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAllDrawdowns(ctx)
		}
	}
}

// handleRiskEvent handles a risk event
func (s *Service) handleRiskEvent(event *models.RiskEvent) {
	logrus.WithFields(logrus.Fields{
		"event_id": event.ID,
		"type":     event.Type,
		"severity": event.Severity,
		"action":   event.Action,
	}).Info("Handling risk event")

	// Store event in database
	s.storeRiskEvent(event)

	// Take action based on event type and severity
	switch event.Action {
	case "block_order":
		// Order already blocked, just log
		logrus.WithField("event_id", event.ID).Info("Order blocked by risk management")
	case "reduce_exposure":
		// Would trigger position reduction
		logrus.WithField("event_id", event.ID).Warn("Exposure reduction required")
	case "stop_strategy":
		// Would trigger strategy stop
		logrus.WithField("event_id", event.ID).Error("Strategy stop required")
	case "monitor_position":
		// Increase monitoring frequency
		logrus.WithField("event_id", event.ID).Info("Position monitoring increased")
	}
}

// checkAllExposures checks exposure for all strategies
func (s *Service) checkAllExposures(ctx context.Context) {
	// Placeholder implementation - would check all strategy exposures
	logrus.Debug("Checking all strategy exposures")
}

// checkAllDrawdowns checks drawdown for all strategies
func (s *Service) checkAllDrawdowns(ctx context.Context) {
	// Placeholder implementation - would check all strategy drawdowns
	logrus.Debug("Checking all strategy drawdowns")
}

// storeRiskEvent stores risk event in database
func (s *Service) storeRiskEvent(event *models.RiskEvent) {
	// Placeholder implementation - would store in database
	logrus.WithField("event_id", event.ID).Debug("Storing risk event")
}
