package valueobjects

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// StrategyType represents the type of trading strategy
type StrategyType string

const (
	StrategyTypeMarketMaking StrategyType = "market_making"
	StrategyTypeArbitrage    StrategyType = "arbitrage"
	StrategyTypeMomentum     StrategyType = "momentum"
	StrategyTypeMeanRevert   StrategyType = "mean_revert"
	StrategyTypeStatArb      StrategyType = "stat_arb"
	StrategyTypeCustom       StrategyType = "custom"
)

// String returns the string representation
func (st StrategyType) String() string {
	return string(st)
}

// IsValid validates the strategy type
func (st StrategyType) IsValid() bool {
	validTypes := []StrategyType{
		StrategyTypeMarketMaking, StrategyTypeArbitrage, StrategyTypeMomentum,
		StrategyTypeMeanRevert, StrategyTypeStatArb, StrategyTypeCustom,
	}
	
	for _, validType := range validTypes {
		if st == validType {
			return true
		}
	}
	return false
}

// StrategyStatus represents the status of a strategy
type StrategyStatus string

const (
	StrategyStatusStopped StrategyStatus = "stopped"
	StrategyStatusRunning StrategyStatus = "running"
	StrategyStatusPaused  StrategyStatus = "paused"
	StrategyStatusError   StrategyStatus = "error"
)

// String returns the string representation
func (ss StrategyStatus) String() string {
	return string(ss)
}

// IsValid validates the strategy status
func (ss StrategyStatus) IsValid() bool {
	validStatuses := []StrategyStatus{
		StrategyStatusStopped, StrategyStatusRunning, StrategyStatusPaused, StrategyStatusError,
	}
	
	for _, validStatus := range validStatuses {
		if ss == validStatus {
			return true
		}
	}
	return false
}

// IsActive returns true if the strategy is in an active state
func (ss StrategyStatus) IsActive() bool {
	return ss == StrategyStatusRunning || ss == StrategyStatusPaused
}

// RiskLimits represents risk management limits for a strategy
type RiskLimits struct {
	MaxPositionSize    decimal.Decimal `json:"max_position_size"`
	MaxDailyLoss       decimal.Decimal `json:"max_daily_loss"`
	MaxDrawdown        decimal.Decimal `json:"max_drawdown"`
	MaxOrderSize       decimal.Decimal `json:"max_order_size"`
	MaxOrdersPerSecond int             `json:"max_orders_per_second"`
	MaxExposure        decimal.Decimal `json:"max_exposure"`
	StopLossPercent    decimal.Decimal `json:"stop_loss_percent"`
	TakeProfitPercent  decimal.Decimal `json:"take_profit_percent"`
}

// NewRiskLimits creates new risk limits with validation
func NewRiskLimits(
	maxPositionSize, maxDailyLoss, maxDrawdown, maxOrderSize, maxExposure decimal.Decimal,
	maxOrdersPerSecond int,
	stopLossPercent, takeProfitPercent decimal.Decimal,
) (RiskLimits, error) {
	rl := RiskLimits{
		MaxPositionSize:    maxPositionSize,
		MaxDailyLoss:       maxDailyLoss,
		MaxDrawdown:        maxDrawdown,
		MaxOrderSize:       maxOrderSize,
		MaxOrdersPerSecond: maxOrdersPerSecond,
		MaxExposure:        maxExposure,
		StopLossPercent:    stopLossPercent,
		TakeProfitPercent:  takeProfitPercent,
	}
	
	if err := rl.Validate(); err != nil {
		return RiskLimits{}, err
	}
	
	return rl, nil
}

// Validate validates the risk limits
func (rl RiskLimits) Validate() error {
	if rl.MaxPositionSize.IsNegative() {
		return fmt.Errorf("max position size cannot be negative")
	}
	if rl.MaxDailyLoss.IsNegative() {
		return fmt.Errorf("max daily loss cannot be negative")
	}
	if rl.MaxDrawdown.IsNegative() || rl.MaxDrawdown.GreaterThan(decimal.NewFromFloat(1.0)) {
		return fmt.Errorf("max drawdown must be between 0 and 1")
	}
	if rl.MaxOrderSize.IsNegative() {
		return fmt.Errorf("max order size cannot be negative")
	}
	if rl.MaxOrdersPerSecond < 0 {
		return fmt.Errorf("max orders per second cannot be negative")
	}
	if rl.MaxExposure.IsNegative() {
		return fmt.Errorf("max exposure cannot be negative")
	}
	if rl.StopLossPercent.IsNegative() || rl.StopLossPercent.GreaterThan(decimal.NewFromFloat(1.0)) {
		return fmt.Errorf("stop loss percent must be between 0 and 1")
	}
	if rl.TakeProfitPercent.IsNegative() {
		return fmt.Errorf("take profit percent cannot be negative")
	}
	return nil
}

// StrategyPerformance represents performance metrics for a strategy
type StrategyPerformance struct {
	TotalPnL        decimal.Decimal `json:"total_pnl"`
	DailyPnL        decimal.Decimal `json:"daily_pnl"`
	TotalTrades     int64           `json:"total_trades"`
	WinningTrades   int64           `json:"winning_trades"`
	LosingTrades    int64           `json:"losing_trades"`
	WinRate         decimal.Decimal `json:"win_rate"`
	AvgWin          decimal.Decimal `json:"avg_win"`
	AvgLoss         decimal.Decimal `json:"avg_loss"`
	ProfitFactor    decimal.Decimal `json:"profit_factor"`
	SharpeRatio     decimal.Decimal `json:"sharpe_ratio"`
	MaxDrawdown     decimal.Decimal `json:"max_drawdown"`
	VolumeTraded    decimal.Decimal `json:"volume_traded"`
	AvgLatency      time.Duration   `json:"avg_latency"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// NewStrategyPerformance creates a new strategy performance instance
func NewStrategyPerformance() StrategyPerformance {
	return StrategyPerformance{
		TotalPnL:        decimal.Zero,
		DailyPnL:        decimal.Zero,
		TotalTrades:     0,
		WinningTrades:   0,
		LosingTrades:    0,
		WinRate:         decimal.Zero,
		AvgWin:          decimal.Zero,
		AvgLoss:         decimal.Zero,
		ProfitFactor:    decimal.Zero,
		SharpeRatio:     decimal.Zero,
		MaxDrawdown:     decimal.Zero,
		VolumeTraded:    decimal.Zero,
		AvgLatency:      0,
		LastUpdated:     time.Now(),
	}
}

// UpdateMetrics updates performance metrics
func (sp *StrategyPerformance) UpdateMetrics(
	pnl decimal.Decimal,
	isWin bool,
	volume decimal.Decimal,
	latency time.Duration,
) {
	sp.TotalPnL = sp.TotalPnL.Add(pnl)
	sp.TotalTrades++
	sp.VolumeTraded = sp.VolumeTraded.Add(volume)
	
	if isWin {
		sp.WinningTrades++
		if sp.WinningTrades == 1 {
			sp.AvgWin = pnl
		} else {
			sp.AvgWin = sp.AvgWin.Add(pnl.Sub(sp.AvgWin).Div(decimal.NewFromInt(sp.WinningTrades)))
		}
	} else {
		sp.LosingTrades++
		if sp.LosingTrades == 1 {
			sp.AvgLoss = pnl.Abs()
		} else {
			sp.AvgLoss = sp.AvgLoss.Add(pnl.Abs().Sub(sp.AvgLoss).Div(decimal.NewFromInt(sp.LosingTrades)))
		}
	}
	
	// Update win rate
	if sp.TotalTrades > 0 {
		sp.WinRate = decimal.NewFromInt(sp.WinningTrades).Div(decimal.NewFromInt(sp.TotalTrades))
	}
	
	// Update profit factor
	if sp.LosingTrades > 0 && !sp.AvgLoss.IsZero() {
		totalWins := sp.AvgWin.Mul(decimal.NewFromInt(sp.WinningTrades))
		totalLosses := sp.AvgLoss.Mul(decimal.NewFromInt(sp.LosingTrades))
		sp.ProfitFactor = totalWins.Div(totalLosses)
	}
	
	// Update average latency
	if sp.TotalTrades == 1 {
		sp.AvgLatency = latency
	} else {
		avgNanos := int64(sp.AvgLatency) + (int64(latency)-int64(sp.AvgLatency))/sp.TotalTrades
		sp.AvgLatency = time.Duration(avgNanos)
	}
	
	sp.LastUpdated = time.Now()
}

// StrategyEvent represents a domain event for strategies
type StrategyEvent struct {
	Type      StrategyEventType      `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// StrategyEventType represents the type of strategy event
type StrategyEventType string

const (
	StrategyEventCreated             StrategyEventType = "strategy_created"
	StrategyEventStarted             StrategyEventType = "strategy_started"
	StrategyEventStopped             StrategyEventType = "strategy_stopped"
	StrategyEventPaused              StrategyEventType = "strategy_paused"
	StrategyEventResumed             StrategyEventType = "strategy_resumed"
	StrategyEventError               StrategyEventType = "strategy_error"
	StrategyEventParametersUpdated   StrategyEventType = "strategy_parameters_updated"
	StrategyEventRiskLimitsUpdated   StrategyEventType = "strategy_risk_limits_updated"
	StrategyEventPerformanceUpdated  StrategyEventType = "strategy_performance_updated"
)

// String returns the string representation
func (set StrategyEventType) String() string {
	return string(set)
}

// IsValid validates the strategy event type
func (set StrategyEventType) IsValid() bool {
	validTypes := []StrategyEventType{
		StrategyEventCreated, StrategyEventStarted, StrategyEventStopped,
		StrategyEventPaused, StrategyEventResumed, StrategyEventError,
		StrategyEventParametersUpdated, StrategyEventRiskLimitsUpdated,
		StrategyEventPerformanceUpdated,
	}
	
	for _, validType := range validTypes {
		if set == validType {
			return true
		}
	}
	return false
}
