package trading

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// StrategyEngine manages coffee-themed trading strategies
type StrategyEngine struct {
	strategies      map[string]*CoffeeStrategy
	activeSignals   map[string]*TradingSignal
	portfolio       *Portfolio
	riskManager     *RiskManager
	signalProcessor *SignalProcessor
	logger          *logrus.Logger
	mutex           sync.RWMutex

	// Channels for communication
	signalChan    chan *TradingSignal
	executionChan chan *TradeExecution
	stopChan      chan struct{}

	// Configuration
	config *EngineConfig
}

// EngineConfig holds configuration for the strategy engine
type EngineConfig struct {
	MaxConcurrentStrategies int             `json:"max_concurrent_strategies"`
	SignalBufferSize        int             `json:"signal_buffer_size"`
	ExecutionBufferSize     int             `json:"execution_buffer_size"`
	TickInterval            time.Duration   `json:"tick_interval"`
	MaxPortfolioRisk        decimal.Decimal `json:"max_portfolio_risk"`
	EmergencyStopEnabled    bool            `json:"emergency_stop_enabled"`
	CoffeeCorrelationWeight decimal.Decimal `json:"coffee_correlation_weight"`
}

// NewStrategyEngine creates a new strategy engine
func NewStrategyEngine(config *EngineConfig, logger *logrus.Logger) *StrategyEngine {
	if config == nil {
		config = &EngineConfig{
			MaxConcurrentStrategies: 10,
			SignalBufferSize:        100,
			ExecutionBufferSize:     100,
			TickInterval:            time.Second * 5,
			MaxPortfolioRisk:        decimal.NewFromFloat(0.02), // 2%
			EmergencyStopEnabled:    true,
			CoffeeCorrelationWeight: decimal.NewFromFloat(0.1), // 10%
		}
	}

	return &StrategyEngine{
		strategies:    make(map[string]*CoffeeStrategy),
		activeSignals: make(map[string]*TradingSignal),
		portfolio:     &Portfolio{},
		config:        config,
		logger:        logger,
		signalChan:    make(chan *TradingSignal, config.SignalBufferSize),
		executionChan: make(chan *TradeExecution, config.ExecutionBufferSize),
		stopChan:      make(chan struct{}),
	}
}

// Start starts the strategy engine
func (se *StrategyEngine) Start(ctx context.Context) error {
	se.logger.Info("Starting Coffee Trading Strategy Engine")

	// Initialize components
	se.riskManager = NewRiskManager(se.config.MaxPortfolioRisk, se.logger)
	se.signalProcessor = NewSignalProcessor(se.logger)

	// Start background goroutines
	go se.signalProcessingLoop(ctx)
	go se.executionProcessingLoop(ctx)
	go se.strategyMonitoringLoop(ctx)
	go se.portfolioUpdateLoop(ctx)

	se.logger.Info("Coffee Trading Strategy Engine started successfully")
	return nil
}

// Stop stops the strategy engine
func (se *StrategyEngine) Stop() error {
	se.logger.Info("Stopping Coffee Trading Strategy Engine")

	close(se.stopChan)

	// Stop all active strategies
	se.mutex.Lock()
	for _, strategy := range se.strategies {
		if strategy.Status == StatusActive {
			strategy.Status = StatusStopped
		}
	}
	se.mutex.Unlock()

	se.logger.Info("Coffee Trading Strategy Engine stopped")
	return nil
}

// AddStrategy adds a new coffee strategy
func (se *StrategyEngine) AddStrategy(strategy *CoffeeStrategy) error {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	if strategy.ID == "" {
		strategy.ID = uuid.New().String()
	}

	// Validate strategy configuration
	if err := se.validateStrategy(strategy); err != nil {
		return fmt.Errorf("invalid strategy configuration: %w", err)
	}

	strategy.CreatedAt = time.Now()
	strategy.UpdatedAt = time.Now()
	strategy.Status = StatusPaused

	se.strategies[strategy.ID] = strategy
	se.logger.Infof("Added %s strategy: %s (%s)", strategy.Type, strategy.Name, strategy.ID)

	return nil
}

// StartStrategy starts a specific strategy
func (se *StrategyEngine) StartStrategy(strategyID string) error {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	strategy, exists := se.strategies[strategyID]
	if !exists {
		return fmt.Errorf("strategy not found: %s", strategyID)
	}

	// Check if we can start more strategies
	activeCount := se.countActiveStrategies()
	if activeCount >= se.config.MaxConcurrentStrategies {
		return fmt.Errorf("maximum concurrent strategies reached: %d", se.config.MaxConcurrentStrategies)
	}

	strategy.Status = StatusActive
	strategy.UpdatedAt = time.Now()

	se.logger.Infof("Started %s strategy: %s", strategy.Type, strategy.Name)
	return nil
}

// StopStrategy stops a specific strategy
func (se *StrategyEngine) StopStrategy(strategyID string) error {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	strategy, exists := se.strategies[strategyID]
	if !exists {
		return fmt.Errorf("strategy not found: %s", strategyID)
	}

	strategy.Status = StatusStopped
	strategy.UpdatedAt = time.Now()

	se.logger.Infof("Stopped %s strategy: %s", strategy.Type, strategy.Name)
	return nil
}

// GetStrategy gets a strategy by ID
func (se *StrategyEngine) GetStrategy(strategyID string) (*CoffeeStrategy, error) {
	se.mutex.RLock()
	defer se.mutex.RUnlock()

	strategy, exists := se.strategies[strategyID]
	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", strategyID)
	}

	return strategy, nil
}

// GetAllStrategies gets all strategies
func (se *StrategyEngine) GetAllStrategies() []*CoffeeStrategy {
	se.mutex.RLock()
	defer se.mutex.RUnlock()

	strategies := make([]*CoffeeStrategy, 0, len(se.strategies))
	for _, strategy := range se.strategies {
		strategies = append(strategies, strategy)
	}

	return strategies
}

// ProcessSignal processes a trading signal
func (se *StrategyEngine) ProcessSignal(signal *TradingSignal) error {
	select {
	case se.signalChan <- signal:
		return nil
	default:
		return fmt.Errorf("signal channel full, dropping signal")
	}
}

// GetPortfolio gets the current portfolio
func (se *StrategyEngine) GetPortfolio() *Portfolio {
	se.mutex.RLock()
	defer se.mutex.RUnlock()

	return se.portfolio
}

// signalProcessingLoop processes incoming trading signals
func (se *StrategyEngine) signalProcessingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-se.stopChan:
			return
		case signal := <-se.signalChan:
			se.processSignalInternal(signal)
		}
	}
}

// processSignalInternal processes a signal internally
func (se *StrategyEngine) processSignalInternal(signal *TradingSignal) {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	// Validate signal
	if err := se.validateSignal(signal); err != nil {
		se.logger.Errorf("Invalid signal: %v", err)
		return
	}

	// Check risk management
	if !se.riskManager.ValidateSignal(signal, se.portfolio) {
		se.logger.Warnf("Signal rejected by risk management: %s %s", signal.Symbol, signal.Type)
		return
	}

	// Store active signal
	se.activeSignals[signal.ID] = signal

	se.logger.Infof("Processed %s signal for %s: %s (confidence: %.2f)",
		signal.Type, signal.Symbol, signal.Source, signal.Confidence)
}

// executionProcessingLoop processes trade executions
func (se *StrategyEngine) executionProcessingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-se.stopChan:
			return
		case execution := <-se.executionChan:
			se.processExecutionInternal(execution)
		}
	}
}

// processExecutionInternal processes a trade execution
func (se *StrategyEngine) processExecutionInternal(execution *TradeExecution) {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	// Update portfolio
	se.updatePortfolioFromExecution(execution)

	// Update strategy performance
	if strategy, exists := se.strategies[execution.StrategyID]; exists {
		se.updateStrategyPerformance(strategy, execution)
	}

	se.logger.Infof("Processed execution: %s %s %s @ %s",
		execution.Side, execution.Quantity, execution.Symbol, execution.ExecutedPrice)
}

// strategyMonitoringLoop monitors strategy performance and health
func (se *StrategyEngine) strategyMonitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(se.config.TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-se.stopChan:
			return
		case <-ticker.C:
			se.monitorStrategies()
		}
	}
}

// monitorStrategies monitors all active strategies
func (se *StrategyEngine) monitorStrategies() {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	for _, strategy := range se.strategies {
		if strategy.Status == StatusActive {
			se.checkStrategyHealth(strategy)
		}
	}
}

// portfolioUpdateLoop updates portfolio metrics
func (se *StrategyEngine) portfolioUpdateLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-se.stopChan:
			return
		case <-ticker.C:
			se.updatePortfolioMetrics()
		}
	}
}

// Helper methods

func (se *StrategyEngine) validateStrategy(strategy *CoffeeStrategy) error {
	if strategy.Name == "" {
		return fmt.Errorf("strategy name is required")
	}

	if strategy.Config.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if strategy.Config.MaxPositionSize.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("max position size must be greater than 0")
	}

	return nil
}

func (se *StrategyEngine) validateSignal(signal *TradingSignal) error {
	if signal.Symbol == "" {
		return fmt.Errorf("signal symbol is required")
	}

	if signal.Confidence.LessThan(decimal.Zero) || signal.Confidence.GreaterThan(decimal.NewFromInt(1)) {
		return fmt.Errorf("signal confidence must be between 0 and 1")
	}

	return nil
}

func (se *StrategyEngine) countActiveStrategies() int {
	count := 0
	for _, strategy := range se.strategies {
		if strategy.Status == StatusActive {
			count++
		}
	}
	return count
}

func (se *StrategyEngine) checkStrategyHealth(strategy *CoffeeStrategy) {
	// Check for emergency stop conditions
	if se.config.EmergencyStopEnabled {
		if strategy.Performance.MaxDrawdown.GreaterThan(decimal.NewFromFloat(0.1)) { // 10% drawdown
			strategy.Status = StatusPaused
			se.logger.Warnf("Emergency stop triggered for strategy %s: max drawdown %.2f%%",
				strategy.Name, strategy.Performance.MaxDrawdown.Mul(decimal.NewFromInt(100)))
		}
	}
}

func (se *StrategyEngine) updatePortfolioFromExecution(execution *TradeExecution) {
	// Update portfolio based on trade execution
	// This would integrate with the actual portfolio management system
	se.portfolio.LastUpdated = time.Now()
}

func (se *StrategyEngine) updateStrategyPerformance(strategy *CoffeeStrategy, execution *TradeExecution) {
	// Update strategy performance metrics
	strategy.Performance.TotalTrades++
	strategy.Performance.LastUpdated = time.Now()
	strategy.LastExecutedAt = &execution.CreatedAt
}

func (se *StrategyEngine) updatePortfolioMetrics() {
	// Update portfolio risk and performance metrics
	se.portfolio.LastUpdated = time.Now()
}

// RiskManager handles risk management for trading strategies
type RiskManager struct {
	maxPortfolioRisk decimal.Decimal
	logger           *logrus.Logger
}

// NewRiskManager creates a new risk manager
func NewRiskManager(maxPortfolioRisk decimal.Decimal, logger *logrus.Logger) *RiskManager {
	return &RiskManager{
		maxPortfolioRisk: maxPortfolioRisk,
		logger:           logger,
	}
}

// ValidateSignal validates a trading signal against risk parameters
func (rm *RiskManager) ValidateSignal(signal *TradingSignal, portfolio *Portfolio) bool {
	// Check minimum confidence
	if signal.Confidence.LessThan(decimal.NewFromFloat(0.5)) {
		rm.logger.Warnf("Signal confidence too low: %.2f", signal.Confidence)
		return false
	}

	// Check portfolio risk
	currentRisk := rm.calculatePortfolioRisk(portfolio)
	if currentRisk.GreaterThan(rm.maxPortfolioRisk) {
		rm.logger.Warnf("Portfolio risk too high: %.2f%%", currentRisk.Mul(decimal.NewFromInt(100)))
		return false
	}

	return true
}

// calculatePortfolioRisk calculates current portfolio risk
func (rm *RiskManager) calculatePortfolioRisk(portfolio *Portfolio) decimal.Decimal {
	// Simplified risk calculation
	if portfolio.TotalValue.IsZero() {
		return decimal.Zero
	}

	totalRisk := decimal.Zero
	for _, position := range portfolio.Positions {
		positionRisk := position.Size.Mul(position.CurrentPrice).Div(portfolio.TotalValue)
		totalRisk = totalRisk.Add(positionRisk)
	}

	return totalRisk
}

// SignalProcessor processes and validates trading signals
type SignalProcessor struct {
	logger *logrus.Logger
}

// NewSignalProcessor creates a new signal processor
func NewSignalProcessor(logger *logrus.Logger) *SignalProcessor {
	return &SignalProcessor{
		logger: logger,
	}
}

// ProcessSignal processes a trading signal
func (sp *SignalProcessor) ProcessSignal(signal *TradingSignal) (*TradingSignal, error) {
	// Validate signal structure
	if signal.Symbol == "" {
		return nil, fmt.Errorf("signal symbol is required")
	}

	if signal.ID == "" {
		signal.ID = uuid.New().String()
	}

	if signal.CreatedAt.IsZero() {
		signal.CreatedAt = time.Now()
	}

	// Set default expiration if not provided
	if signal.ExpiresAt == nil {
		expiration := signal.CreatedAt.Add(time.Hour)
		signal.ExpiresAt = &expiration
	}

	signal.Status = "pending"

	sp.logger.Infof("Processed signal: %s %s %s", signal.Type, signal.Symbol, signal.Source)
	return signal, nil
}
