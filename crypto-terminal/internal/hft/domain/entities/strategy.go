package entities

import (
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Import value objects for cleaner code
type (
	StrategyType        = valueobjects.StrategyType
	StrategyStatus      = valueobjects.StrategyStatus
	RiskLimits          = valueobjects.RiskLimits
	StrategyPerformance = valueobjects.StrategyPerformance
	StrategyEvent       = valueobjects.StrategyEvent
	StrategyEventType   = valueobjects.StrategyEventType
)

// Import constants
const (
	StrategyStatusStopped = valueobjects.StrategyStatusStopped
	StrategyStatusRunning = valueobjects.StrategyStatusRunning
	StrategyStatusPaused  = valueobjects.StrategyStatusPaused
	StrategyStatusError   = valueobjects.StrategyStatusError
)

// Strategy represents a trading strategy entity in the domain
type Strategy struct {
	id          StrategyID
	name        string
	strategyType StrategyType
	status      StrategyStatus
	symbols     []Symbol
	exchanges   []Exchange
	parameters  map[string]interface{}
	riskLimits  RiskLimits
	performance StrategyPerformance
	createdAt   time.Time
	updatedAt   time.Time
	startedAt   *time.Time
	stoppedAt   *time.Time
	events      []StrategyEvent
}

// NewStrategy creates a new strategy entity
func NewStrategy(
	name string,
	strategyType StrategyType,
	symbols []Symbol,
	exchanges []Exchange,
	parameters map[string]interface{},
	riskLimits RiskLimits,
) (*Strategy, error) {
	if err := validateStrategyParameters(name, strategyType, symbols, exchanges, riskLimits); err != nil {
		return nil, fmt.Errorf("invalid strategy parameters: %w", err)
	}

	now := time.Now()
	strategy := &Strategy{
		id:           StrategyID(generateStrategyID()),
		name:         name,
		strategyType: strategyType,
		status:       StrategyStatusStopped,
		symbols:      symbols,
		exchanges:    exchanges,
		parameters:   parameters,
		riskLimits:   riskLimits,
		performance:  StrategyPerformance{},
		createdAt:    now,
		updatedAt:    now,
		events:       make([]StrategyEvent, 0),
	}

	// Record strategy creation event
	strategy.recordEvent(valueobjects.StrategyEventCreated, map[string]interface{}{
		"strategy_id":   strategy.id,
		"name":          strategy.name,
		"strategy_type": strategy.strategyType,
		"symbols":       strategy.symbols,
		"exchanges":     strategy.exchanges,
	})

	return strategy, nil
}

// GetID returns the strategy ID
func (s *Strategy) GetID() StrategyID {
	return s.id
}

// GetName returns the strategy name
func (s *Strategy) GetName() string {
	return s.name
}

// GetType returns the strategy type
func (s *Strategy) GetType() StrategyType {
	return s.strategyType
}

// GetStatus returns the strategy status
func (s *Strategy) GetStatus() StrategyStatus {
	return s.status
}

// GetSymbols returns the trading symbols
func (s *Strategy) GetSymbols() []Symbol {
	return s.symbols
}

// GetExchanges returns the exchanges
func (s *Strategy) GetExchanges() []Exchange {
	return s.exchanges
}

// GetParameters returns the strategy parameters
func (s *Strategy) GetParameters() map[string]interface{} {
	return s.parameters
}

// GetRiskLimits returns the risk limits
func (s *Strategy) GetRiskLimits() RiskLimits {
	return s.riskLimits
}

// GetPerformance returns the strategy performance
func (s *Strategy) GetPerformance() StrategyPerformance {
	return s.performance
}

// GetCreatedAt returns the creation timestamp
func (s *Strategy) GetCreatedAt() time.Time {
	return s.createdAt
}

// GetUpdatedAt returns the last update timestamp
func (s *Strategy) GetUpdatedAt() time.Time {
	return s.updatedAt
}

// GetStartedAt returns the start timestamp
func (s *Strategy) GetStartedAt() *time.Time {
	return s.startedAt
}

// GetStoppedAt returns the stop timestamp
func (s *Strategy) GetStoppedAt() *time.Time {
	return s.stoppedAt
}

// GetEvents returns the strategy events
func (s *Strategy) GetEvents() []StrategyEvent {
	return s.events
}

// Start starts the strategy
func (s *Strategy) Start() error {
	if s.status == StrategyStatusRunning {
		return fmt.Errorf("strategy is already running")
	}

	if s.status == StrategyStatusError {
		return fmt.Errorf("cannot start strategy in error state")
	}

	s.status = StrategyStatusRunning
	now := time.Now()
	s.startedAt = &now
	s.updatedAt = now

	s.recordEvent(valueobjects.StrategyEventStarted, map[string]interface{}{
		"status":     s.status,
		"started_at": s.startedAt,
	})

	return nil
}

// Stop stops the strategy
func (s *Strategy) Stop() error {
	if s.status == StrategyStatusStopped {
		return nil // Already stopped
	}

	s.status = StrategyStatusStopped
	now := time.Now()
	s.stoppedAt = &now
	s.updatedAt = now

	s.recordEvent(valueobjects.StrategyEventStopped, map[string]interface{}{
		"status":     s.status,
		"stopped_at": s.stoppedAt,
	})

	return nil
}

// Pause pauses the strategy
func (s *Strategy) Pause() error {
	if s.status != StrategyStatusRunning {
		return fmt.Errorf("can only pause running strategy, current status: %s", s.status)
	}

	s.status = StrategyStatusPaused
	s.updatedAt = time.Now()

	s.recordEvent(valueobjects.StrategyEventPaused, map[string]interface{}{
		"status": s.status,
	})

	return nil
}

// Resume resumes the strategy
func (s *Strategy) Resume() error {
	if s.status != StrategyStatusPaused {
		return fmt.Errorf("can only resume paused strategy, current status: %s", s.status)
	}

	s.status = StrategyStatusRunning
	s.updatedAt = time.Now()

	s.recordEvent(valueobjects.StrategyEventResumed, map[string]interface{}{
		"status": s.status,
	})

	return nil
}

// SetError sets the strategy to error state
func (s *Strategy) SetError(errorMessage string) {
	s.status = StrategyStatusError
	s.updatedAt = time.Now()

	s.recordEvent(valueobjects.StrategyEventError, map[string]interface{}{
		"status": s.status,
		"error":  errorMessage,
	})
}

// UpdateParameters updates the strategy parameters
func (s *Strategy) UpdateParameters(parameters map[string]interface{}) {
	s.parameters = parameters
	s.updatedAt = time.Now()

	s.recordEvent(valueobjects.StrategyEventParametersUpdated, map[string]interface{}{
		"parameters": parameters,
	})
}

// UpdateRiskLimits updates the risk limits
func (s *Strategy) UpdateRiskLimits(riskLimits RiskLimits) error {
	if err := riskLimits.Validate(); err != nil {
		return fmt.Errorf("invalid risk limits: %w", err)
	}

	s.riskLimits = riskLimits
	s.updatedAt = time.Now()

	s.recordEvent(valueobjects.StrategyEventRiskLimitsUpdated, map[string]interface{}{
		"risk_limits": riskLimits,
	})

	return nil
}

// UpdatePerformance updates the strategy performance metrics
func (s *Strategy) UpdatePerformance(performance StrategyPerformance) {
	s.performance = performance
	s.updatedAt = time.Now()

	s.recordEvent(valueobjects.StrategyEventPerformanceUpdated, map[string]interface{}{
		"performance": performance,
	})
}

// IsActive returns true if the strategy is active (running or paused)
func (s *Strategy) IsActive() bool {
	return s.status == StrategyStatusRunning || s.status == StrategyStatusPaused
}

// IsRunning returns true if the strategy is running
func (s *Strategy) IsRunning() bool {
	return s.status == StrategyStatusRunning
}

// IsStopped returns true if the strategy is stopped
func (s *Strategy) IsStopped() bool {
	return s.status == StrategyStatusStopped
}

// IsPaused returns true if the strategy is paused
func (s *Strategy) IsPaused() bool {
	return s.status == StrategyStatusPaused
}

// IsInError returns true if the strategy is in error state
func (s *Strategy) IsInError() bool {
	return s.status == StrategyStatusError
}

// recordEvent records a domain event
func (s *Strategy) recordEvent(eventType StrategyEventType, data map[string]interface{}) {
	event := StrategyEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}
	s.events = append(s.events, event)
}

// generateStrategyID generates a unique strategy ID
func generateStrategyID() string {
	return uuid.New().String()
}

// validateStrategyParameters validates strategy creation parameters
func validateStrategyParameters(
	name string,
	strategyType StrategyType,
	symbols []Symbol,
	exchanges []Exchange,
	riskLimits RiskLimits,
) error {
	if name == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}
	if !strategyType.IsValid() {
		return fmt.Errorf("invalid strategy type: %s", strategyType)
	}
	if len(symbols) == 0 {
		return fmt.Errorf("strategy must have at least one symbol")
	}
	if len(exchanges) == 0 {
		return fmt.Errorf("strategy must have at least one exchange")
	}
	if err := riskLimits.Validate(); err != nil {
		return fmt.Errorf("invalid risk limits: %w", err)
	}
	return nil
}
