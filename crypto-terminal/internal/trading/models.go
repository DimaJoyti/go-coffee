package trading

import (
	"time"

	"github.com/shopspring/decimal"
)

// CoffeeStrategyType represents different coffee-themed trading strategies
type CoffeeStrategyType string

const (
	StrategyEspresso   CoffeeStrategyType = "espresso"   // High-frequency scalping
	StrategyLatte      CoffeeStrategyType = "latte"      // Smooth swing trading
	StrategyColdBrew   CoffeeStrategyType = "cold_brew"  // Patient position trading
	StrategyCappuccino CoffeeStrategyType = "cappuccino" // Frothy momentum trading
)

// StrategyStatus represents the status of a trading strategy
type StrategyStatus string

const (
	StatusActive   StrategyStatus = "active"
	StatusPaused   StrategyStatus = "paused"
	StatusStopped  StrategyStatus = "stopped"
	StatusError    StrategyStatus = "error"
)

// SignalType represents different types of trading signals
type SignalType string

const (
	SignalBuy    SignalType = "BUY"
	SignalSell   SignalType = "SELL"
	SignalHold   SignalType = "HOLD"
	SignalClose  SignalType = "CLOSE"
)

// SignalSource represents the source of a trading signal
type SignalSource string

const (
	SourceTradingView SignalSource = "tradingview"
	SourceBinance     SignalSource = "binance"
	SourceTechnical   SignalSource = "technical"
	SourceSentiment   SignalSource = "sentiment"
	SourceCombined    SignalSource = "combined"
)

// CoffeeStrategy represents a coffee-themed trading strategy
type CoffeeStrategy struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Type              CoffeeStrategyType `json:"type"`
	Description       string             `json:"description"`
	Status            StrategyStatus     `json:"status"`
	Config            StrategyConfig     `json:"config"`
	Performance       StrategyPerformance `json:"performance"`
	RiskManagement    RiskConfig         `json:"risk_management"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	LastExecutedAt    *time.Time         `json:"last_executed_at,omitempty"`
}

// StrategyConfig holds configuration for a trading strategy
type StrategyConfig struct {
	Symbol              string          `json:"symbol"`
	Timeframe           string          `json:"timeframe"`           // 1m, 5m, 15m, 1h, 4h, 1d
	MaxPositionSize     decimal.Decimal `json:"max_position_size"`   // Percentage of portfolio
	StopLossPercent     decimal.Decimal `json:"stop_loss_percent"`   // Stop loss percentage
	TakeProfitPercent   decimal.Decimal `json:"take_profit_percent"` // Take profit percentage
	MinConfidence       decimal.Decimal `json:"min_confidence"`      // Minimum signal confidence (0-1)
	UseTrailingStop     bool            `json:"use_trailing_stop"`
	TrailingStopPercent decimal.Decimal `json:"trailing_stop_percent"`
	MaxDailyTrades      int             `json:"max_daily_trades"`
	TradingHours        TradingHours    `json:"trading_hours"`
	Indicators          []string        `json:"indicators"`          // RSI, MACD, BB, EMA, etc.
	CoffeeCorrelation   bool            `json:"coffee_correlation"`  // Use coffee shop data
}

// TradingHours defines when the strategy should be active
type TradingHours struct {
	Enabled   bool   `json:"enabled"`
	StartHour int    `json:"start_hour"` // 0-23
	EndHour   int    `json:"end_hour"`   // 0-23
	Timezone  string `json:"timezone"`   // UTC, EST, etc.
	Weekdays  []int  `json:"weekdays"`   // 0=Sunday, 1=Monday, etc.
}

// StrategyPerformance tracks strategy performance metrics
type StrategyPerformance struct {
	TotalTrades       int             `json:"total_trades"`
	WinningTrades     int             `json:"winning_trades"`
	LosingTrades      int             `json:"losing_trades"`
	WinRate           decimal.Decimal `json:"win_rate"`           // Percentage
	TotalPnL          decimal.Decimal `json:"total_pnl"`          // Total profit/loss
	RealizedPnL       decimal.Decimal `json:"realized_pnl"`       // Realized profit/loss
	UnrealizedPnL     decimal.Decimal `json:"unrealized_pnl"`     // Unrealized profit/loss
	MaxDrawdown       decimal.Decimal `json:"max_drawdown"`       // Maximum drawdown
	SharpeRatio       decimal.Decimal `json:"sharpe_ratio"`       // Risk-adjusted return
	AverageWin        decimal.Decimal `json:"average_win"`        // Average winning trade
	AverageLoss       decimal.Decimal `json:"average_loss"`       // Average losing trade
	ProfitFactor      decimal.Decimal `json:"profit_factor"`      // Gross profit / Gross loss
	LastUpdated       time.Time       `json:"last_updated"`
}

// RiskConfig defines risk management parameters
type RiskConfig struct {
	MaxPortfolioRisk    decimal.Decimal `json:"max_portfolio_risk"`    // Max % of portfolio at risk
	MaxPositionRisk     decimal.Decimal `json:"max_position_risk"`     // Max % per position
	MaxCorrelation      decimal.Decimal `json:"max_correlation"`       // Max correlation between positions
	EmergencyStopLoss   decimal.Decimal `json:"emergency_stop_loss"`   // Emergency portfolio stop loss
	DailyLossLimit      decimal.Decimal `json:"daily_loss_limit"`      // Max daily loss
	MaxLeverage         decimal.Decimal `json:"max_leverage"`          // Maximum leverage
	RequiredMargin      decimal.Decimal `json:"required_margin"`       // Required margin percentage
	RiskRewardRatio     decimal.Decimal `json:"risk_reward_ratio"`     // Minimum risk/reward ratio
}

// TradingSignal represents a trading signal generated by a strategy
type TradingSignal struct {
	ID               string          `json:"id"`
	StrategyID       string          `json:"strategy_id"`
	Symbol           string          `json:"symbol"`
	Type             SignalType      `json:"type"`
	Source           SignalSource    `json:"source"`
	Confidence       decimal.Decimal `json:"confidence"`       // 0-1 confidence score
	Price            decimal.Decimal `json:"price"`            // Signal price
	TargetPrice      decimal.Decimal `json:"target_price"`     // Target price
	StopLoss         decimal.Decimal `json:"stop_loss"`        // Stop loss price
	Quantity         decimal.Decimal `json:"quantity"`         // Suggested quantity
	Timeframe        string          `json:"timeframe"`        // Signal timeframe
	Indicators       map[string]interface{} `json:"indicators"` // Technical indicators
	Metadata         map[string]interface{} `json:"metadata"`   // Additional data
	CreatedAt        time.Time       `json:"created_at"`
	ExpiresAt        *time.Time      `json:"expires_at,omitempty"`
	ExecutedAt       *time.Time      `json:"executed_at,omitempty"`
	Status           string          `json:"status"`           // pending, executed, expired, cancelled
}

// TradeExecution represents an executed trade
type TradeExecution struct {
	ID                string          `json:"id"`
	SignalID          string          `json:"signal_id"`
	StrategyID        string          `json:"strategy_id"`
	BinanceOrderID    string          `json:"binance_order_id"`
	Symbol            string          `json:"symbol"`
	Side              string          `json:"side"`             // BUY, SELL
	Type              string          `json:"type"`             // MARKET, LIMIT, STOP, etc.
	Quantity          decimal.Decimal `json:"quantity"`
	Price             decimal.Decimal `json:"price"`
	ExecutedQuantity  decimal.Decimal `json:"executed_quantity"`
	ExecutedPrice     decimal.Decimal `json:"executed_price"`
	Commission        decimal.Decimal `json:"commission"`
	CommissionAsset   string          `json:"commission_asset"`
	Status            string          `json:"status"`           // NEW, FILLED, CANCELLED, etc.
	PnL               decimal.Decimal `json:"pnl"`              // Profit/Loss
	CreatedAt         time.Time       `json:"created_at"`
	ExecutedAt        *time.Time      `json:"executed_at,omitempty"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// Portfolio represents the current portfolio state
type Portfolio struct {
	ID                 string                    `json:"id"`
	Name               string                    `json:"name"`
	TotalValue         decimal.Decimal           `json:"total_value"`         // Total portfolio value
	AvailableBalance   decimal.Decimal           `json:"available_balance"`   // Available for trading
	InvestedAmount     decimal.Decimal           `json:"invested_amount"`     // Currently invested
	TotalPnL           decimal.Decimal           `json:"total_pnl"`           // Total profit/loss
	DailyPnL           decimal.Decimal           `json:"daily_pnl"`           // Daily profit/loss
	UnrealizedPnL      decimal.Decimal           `json:"unrealized_pnl"`      // Unrealized profit/loss
	RealizedPnL        decimal.Decimal           `json:"realized_pnl"`        // Realized profit/loss
	Positions          map[string]Position       `json:"positions"`           // Current positions
	ActiveStrategies   []string                  `json:"active_strategies"`   // Active strategy IDs
	RiskMetrics        PortfolioRiskMetrics      `json:"risk_metrics"`
	Performance        PortfolioPerformance      `json:"performance"`
	LastUpdated        time.Time                 `json:"last_updated"`
}

// Position represents a trading position
type Position struct {
	Symbol            string          `json:"symbol"`
	Side              string          `json:"side"`             // LONG, SHORT
	Size              decimal.Decimal `json:"size"`             // Position size
	EntryPrice        decimal.Decimal `json:"entry_price"`      // Average entry price
	CurrentPrice      decimal.Decimal `json:"current_price"`    // Current market price
	UnrealizedPnL     decimal.Decimal `json:"unrealized_pnl"`   // Unrealized profit/loss
	RealizedPnL       decimal.Decimal `json:"realized_pnl"`     // Realized profit/loss
	MarginUsed        decimal.Decimal `json:"margin_used"`      // Margin used
	Leverage          decimal.Decimal `json:"leverage"`         // Position leverage
	StopLoss          decimal.Decimal `json:"stop_loss"`        // Stop loss price
	TakeProfit        decimal.Decimal `json:"take_profit"`      // Take profit price
	OpenTime          time.Time       `json:"open_time"`        // Position open time
	LastUpdated       time.Time       `json:"last_updated"`
}

// PortfolioRiskMetrics represents portfolio risk metrics
type PortfolioRiskMetrics struct {
	VaR95             decimal.Decimal `json:"var_95"`             // Value at Risk 95%
	VaR99             decimal.Decimal `json:"var_99"`             // Value at Risk 99%
	MaxDrawdown       decimal.Decimal `json:"max_drawdown"`       // Maximum drawdown
	PortfolioVolatility decimal.Decimal `json:"portfolio_volatility"` // Portfolio volatility
	RiskScore         decimal.Decimal `json:"risk_score"`         // Overall risk score (0-100)
	Correlation       decimal.Decimal `json:"correlation"`        // Average position correlation
	Leverage          decimal.Decimal `json:"leverage"`           // Portfolio leverage
	MarginUsage       decimal.Decimal `json:"margin_usage"`       // Margin usage percentage
}

// PortfolioPerformance represents portfolio performance metrics
type PortfolioPerformance struct {
	TotalReturn       decimal.Decimal `json:"total_return"`       // Total return percentage
	DailyReturn       decimal.Decimal `json:"daily_return"`       // Daily return percentage
	WeeklyReturn      decimal.Decimal `json:"weekly_return"`      // Weekly return percentage
	MonthlyReturn     decimal.Decimal `json:"monthly_return"`     // Monthly return percentage
	YearlyReturn      decimal.Decimal `json:"yearly_return"`      // Yearly return percentage
	SharpeRatio       decimal.Decimal `json:"sharpe_ratio"`       // Sharpe ratio
	SortinoRatio      decimal.Decimal `json:"sortino_ratio"`      // Sortino ratio
	CalmarRatio       decimal.Decimal `json:"calmar_ratio"`       // Calmar ratio
	Alpha             decimal.Decimal `json:"alpha"`              // Alpha
	Beta              decimal.Decimal `json:"beta"`               // Beta
	MaxDrawdown       decimal.Decimal `json:"max_drawdown"`       // Maximum drawdown
	WinRate           decimal.Decimal `json:"win_rate"`           // Win rate percentage
	ProfitFactor      decimal.Decimal `json:"profit_factor"`      // Profit factor
	LastUpdated       time.Time       `json:"last_updated"`
}

// CoffeeShopData represents coffee shop business data for correlation analysis
type CoffeeShopData struct {
	ShopID          string          `json:"shop_id"`
	Date            time.Time       `json:"date"`
	DailySales      decimal.Decimal `json:"daily_sales"`      // Daily sales amount
	CustomerCount   int             `json:"customer_count"`   // Number of customers
	AverageOrder    decimal.Decimal `json:"average_order"`    // Average order value
	PopularDrinks   []string        `json:"popular_drinks"`   // Most popular drinks
	WeatherCondition string         `json:"weather_condition"` // Weather impact
	TradingProfit   decimal.Decimal `json:"trading_profit"`   // Trading profits for the day
	Correlation     decimal.Decimal `json:"correlation"`      // Sales-trading correlation
}
