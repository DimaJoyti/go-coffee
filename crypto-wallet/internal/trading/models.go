package trading

import (
	"time"
)

// Order represents a trading order
type Order struct {
	ID              string                 `json:"id" db:"id"`
	AccountID       string                 `json:"account_id" db:"account_id"`
	WalletID        string                 `json:"wallet_id" db:"wallet_id"`
	ExchangeID      string                 `json:"exchange_id" db:"exchange_id"`
	Symbol          string                 `json:"symbol" db:"symbol"`
	BaseAsset       string                 `json:"base_asset" db:"base_asset"`
	QuoteAsset      string                 `json:"quote_asset" db:"quote_asset"`
	OrderType       OrderType              `json:"order_type" db:"order_type"`
	Side            OrderSide              `json:"side" db:"side"`
	Quantity        string                 `json:"quantity" db:"quantity"`
	Price           string                 `json:"price" db:"price"`
	StopPrice       string                 `json:"stop_price" db:"stop_price"`
	FilledQuantity  string                 `json:"filled_quantity" db:"filled_quantity"`
	RemainingQuantity string               `json:"remaining_quantity" db:"remaining_quantity"`
	AveragePrice    string                 `json:"average_price" db:"average_price"`
	TotalValue      string                 `json:"total_value" db:"total_value"`
	Fee             string                 `json:"fee" db:"fee"`
	FeeAsset        string                 `json:"fee_asset" db:"fee_asset"`
	Status          OrderStatus            `json:"status" db:"status"`
	TimeInForce     TimeInForce            `json:"time_in_force" db:"time_in_force"`
	ExpiresAt       *time.Time             `json:"expires_at" db:"expires_at"`
	ExecutedAt      *time.Time             `json:"executed_at" db:"executed_at"`
	CancelledAt     *time.Time             `json:"cancelled_at" db:"cancelled_at"`
	ExchangeOrderID string                 `json:"exchange_order_id" db:"exchange_order_id"`
	ClientOrderID   string                 `json:"client_order_id" db:"client_order_id"`
	StrategyID      string                 `json:"strategy_id" db:"strategy_id"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// OrderType represents the type of trading order
type OrderType string

const (
	OrderTypeMarket    OrderType = "market"
	OrderTypeLimit     OrderType = "limit"
	OrderTypeStopLoss  OrderType = "stop_loss"
	OrderTypeStopLimit OrderType = "stop_limit"
	OrderTypeTakeProfit OrderType = "take_profit"
	OrderTypeTrailingStop OrderType = "trailing_stop"
)

// OrderSide represents the side of a trading order
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderStatus represents the status of a trading order
type OrderStatus string

const (
	OrderStatusPending       OrderStatus = "pending"
	OrderStatusOpen          OrderStatus = "open"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled        OrderStatus = "filled"
	OrderStatusCancelled     OrderStatus = "cancelled"
	OrderStatusRejected      OrderStatus = "rejected"
	OrderStatusExpired       OrderStatus = "expired"
)

// TimeInForce represents the time in force for an order
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "gtc" // Good Till Cancelled
	TimeInForceIOC TimeInForce = "ioc" // Immediate Or Cancel
	TimeInForceFOK TimeInForce = "fok" // Fill Or Kill
	TimeInForceGTD TimeInForce = "gtd" // Good Till Date
)

// Trade represents an executed trade
type Trade struct {
	ID              string                 `json:"id" db:"id"`
	OrderID         string                 `json:"order_id" db:"order_id"`
	AccountID       string                 `json:"account_id" db:"account_id"`
	ExchangeID      string                 `json:"exchange_id" db:"exchange_id"`
	Symbol          string                 `json:"symbol" db:"symbol"`
	Side            OrderSide              `json:"side" db:"side"`
	Quantity        string                 `json:"quantity" db:"quantity"`
	Price           string                 `json:"price" db:"price"`
	Value           string                 `json:"value" db:"value"`
	Fee             string                 `json:"fee" db:"fee"`
	FeeAsset        string                 `json:"fee_asset" db:"fee_asset"`
	IsMaker         bool                   `json:"is_maker" db:"is_maker"`
	ExchangeTradeID string                 `json:"exchange_trade_id" db:"exchange_trade_id"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	ExecutedAt      time.Time              `json:"executed_at" db:"executed_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// Portfolio represents a trading portfolio
type Portfolio struct {
	ID              string                 `json:"id" db:"id"`
	AccountID       string                 `json:"account_id" db:"account_id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	BaseCurrency    string                 `json:"base_currency" db:"base_currency"`
	TotalValue      string                 `json:"total_value" db:"total_value"`
	TotalPnL        string                 `json:"total_pnl" db:"total_pnl"`
	TotalPnLPercent float64                `json:"total_pnl_percent" db:"total_pnl_percent"`
	DayPnL          string                 `json:"day_pnl" db:"day_pnl"`
	DayPnLPercent   float64                `json:"day_pnl_percent" db:"day_pnl_percent"`
	Holdings        []Holding              `json:"holdings" db:"holdings"`
	Performance     PortfolioPerformance   `json:"performance" db:"performance"`
	RiskMetrics     RiskMetrics            `json:"risk_metrics" db:"risk_metrics"`
	IsActive        bool                   `json:"is_active" db:"is_active"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// Holding represents a portfolio holding
type Holding struct {
	Asset           string    `json:"asset"`
	Quantity        string    `json:"quantity"`
	AveragePrice    string    `json:"average_price"`
	CurrentPrice    string    `json:"current_price"`
	MarketValue     string    `json:"market_value"`
	CostBasis       string    `json:"cost_basis"`
	UnrealizedPnL   string    `json:"unrealized_pnl"`
	UnrealizedPnLPercent float64 `json:"unrealized_pnl_percent"`
	Allocation      float64   `json:"allocation"`
	LastUpdated     time.Time `json:"last_updated"`
}

// PortfolioPerformance represents portfolio performance metrics
type PortfolioPerformance struct {
	TotalReturn      string    `json:"total_return"`
	TotalReturnPercent float64 `json:"total_return_percent"`
	AnnualizedReturn float64   `json:"annualized_return"`
	Volatility       float64   `json:"volatility"`
	SharpeRatio      float64   `json:"sharpe_ratio"`
	MaxDrawdown      float64   `json:"max_drawdown"`
	WinRate          float64   `json:"win_rate"`
	ProfitFactor     float64   `json:"profit_factor"`
	TotalTrades      int       `json:"total_trades"`
	WinningTrades    int       `json:"winning_trades"`
	LosingTrades     int       `json:"losing_trades"`
	LastUpdated      time.Time `json:"last_updated"`
}

// RiskMetrics represents portfolio risk metrics
type RiskMetrics struct {
	VaR95           string    `json:"var_95"`
	VaR99           string    `json:"var_99"`
	Beta            float64   `json:"beta"`
	Alpha           float64   `json:"alpha"`
	Correlation     float64   `json:"correlation"`
	ConcentrationRisk float64 `json:"concentration_risk"`
	LiquidityRisk   float64   `json:"liquidity_risk"`
	LastUpdated     time.Time `json:"last_updated"`
}

// Strategy represents a trading strategy
type Strategy struct {
	ID              string                 `json:"id" db:"id"`
	AccountID       string                 `json:"account_id" db:"account_id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	StrategyType    StrategyType           `json:"strategy_type" db:"strategy_type"`
	Status          StrategyStatus         `json:"status" db:"status"`
	Parameters      StrategyParameters     `json:"parameters" db:"parameters"`
	Performance     StrategyPerformance    `json:"performance" db:"performance"`
	RiskLimits      RiskLimits             `json:"risk_limits" db:"risk_limits"`
	Allocation      string                 `json:"allocation" db:"allocation"`
	MaxAllocation   string                 `json:"max_allocation" db:"max_allocation"`
	IsBacktested    bool                   `json:"is_backtested" db:"is_backtested"`
	BacktestResults BacktestResults        `json:"backtest_results" db:"backtest_results"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// StrategyType represents the type of trading strategy
type StrategyType string

const (
	StrategyTypeManual      StrategyType = "manual"
	StrategyTypeAlgorithmic StrategyType = "algorithmic"
	StrategyTypeCopyTrading StrategyType = "copy_trading"
	StrategyTypeGridTrading StrategyType = "grid_trading"
	StrategyTypeDCA         StrategyType = "dca"
	StrategyTypeArbitrage   StrategyType = "arbitrage"
)

// StrategyStatus represents the status of a trading strategy
type StrategyStatus string

const (
	StrategyStatusActive   StrategyStatus = "active"
	StrategyStatusInactive StrategyStatus = "inactive"
	StrategyStatusPaused   StrategyStatus = "paused"
	StrategyStatusStopped  StrategyStatus = "stopped"
	StrategyStatusError    StrategyStatus = "error"
)

// StrategyParameters represents strategy parameters
type StrategyParameters struct {
	Symbols         []string               `json:"symbols"`
	Timeframe       string                 `json:"timeframe"`
	Indicators      map[string]interface{} `json:"indicators"`
	EntryConditions map[string]interface{} `json:"entry_conditions"`
	ExitConditions  map[string]interface{} `json:"exit_conditions"`
	RiskManagement  map[string]interface{} `json:"risk_management"`
	CustomParams    map[string]interface{} `json:"custom_params"`
}

// StrategyPerformance represents strategy performance metrics
type StrategyPerformance struct {
	TotalReturn      string    `json:"total_return"`
	TotalReturnPercent float64 `json:"total_return_percent"`
	WinRate          float64   `json:"win_rate"`
	ProfitFactor     float64   `json:"profit_factor"`
	SharpeRatio      float64   `json:"sharpe_ratio"`
	MaxDrawdown      float64   `json:"max_drawdown"`
	TotalTrades      int       `json:"total_trades"`
	WinningTrades    int       `json:"winning_trades"`
	LosingTrades     int       `json:"losing_trades"`
	AverageTrade     string    `json:"average_trade"`
	BestTrade        string    `json:"best_trade"`
	WorstTrade       string    `json:"worst_trade"`
	LastUpdated      time.Time `json:"last_updated"`
}

// RiskLimits represents risk limits for a strategy
type RiskLimits struct {
	MaxPositionSize  string  `json:"max_position_size"`
	MaxDailyLoss     string  `json:"max_daily_loss"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	StopLossPercent  float64 `json:"stop_loss_percent"`
	TakeProfitPercent float64 `json:"take_profit_percent"`
	MaxLeverage      float64 `json:"max_leverage"`
}

// BacktestResults represents backtest results
type BacktestResults struct {
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	InitialCapital   string    `json:"initial_capital"`
	FinalCapital     string    `json:"final_capital"`
	TotalReturn      string    `json:"total_return"`
	TotalReturnPercent float64 `json:"total_return_percent"`
	AnnualizedReturn float64   `json:"annualized_return"`
	Volatility       float64   `json:"volatility"`
	SharpeRatio      float64   `json:"sharpe_ratio"`
	MaxDrawdown      float64   `json:"max_drawdown"`
	WinRate          float64   `json:"win_rate"`
	TotalTrades      int       `json:"total_trades"`
	WinningTrades    int       `json:"winning_trades"`
	LosingTrades     int       `json:"losing_trades"`
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	AccountID     string                 `json:"account_id" validate:"required"`
	WalletID      string                 `json:"wallet_id" validate:"required"`
	ExchangeID    string                 `json:"exchange_id" validate:"required"`
	Symbol        string                 `json:"symbol" validate:"required"`
	OrderType     OrderType              `json:"order_type" validate:"required"`
	Side          OrderSide              `json:"side" validate:"required"`
	Quantity      string                 `json:"quantity" validate:"required"`
	Price         string                 `json:"price,omitempty"`
	StopPrice     string                 `json:"stop_price,omitempty"`
	TimeInForce   TimeInForce            `json:"time_in_force"`
	ClientOrderID string                 `json:"client_order_id,omitempty"`
	StrategyID    string                 `json:"strategy_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateOrderRequest represents a request to update an order
type UpdateOrderRequest struct {
	Quantity  *string `json:"quantity,omitempty"`
	Price     *string `json:"price,omitempty"`
	StopPrice *string `json:"stop_price,omitempty"`
}

// CancelOrderRequest represents a request to cancel an order
type CancelOrderRequest struct {
	OrderID string `json:"order_id" validate:"required"`
}

// CreateStrategyRequest represents a request to create a strategy
type CreateStrategyRequest struct {
	AccountID     string                 `json:"account_id" validate:"required"`
	Name          string                 `json:"name" validate:"required"`
	Description   string                 `json:"description"`
	StrategyType  StrategyType           `json:"strategy_type" validate:"required"`
	Parameters    StrategyParameters     `json:"parameters" validate:"required"`
	RiskLimits    RiskLimits             `json:"risk_limits" validate:"required"`
	Allocation    string                 `json:"allocation" validate:"required"`
	MaxAllocation string                 `json:"max_allocation" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// OrderListRequest represents a request to list orders
type OrderListRequest struct {
	Page       int         `json:"page" validate:"min=1"`
	Limit      int         `json:"limit" validate:"min=1,max=100"`
	AccountID  string      `json:"account_id,omitempty"`
	ExchangeID string      `json:"exchange_id,omitempty"`
	Symbol     string      `json:"symbol,omitempty"`
	Status     OrderStatus `json:"status,omitempty"`
	Side       OrderSide   `json:"side,omitempty"`
	DateFrom   *time.Time  `json:"date_from,omitempty"`
	DateTo     *time.Time  `json:"date_to,omitempty"`
}

// OrderListResponse represents a response to list orders
type OrderListResponse struct {
	Orders     []Order `json:"orders"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalPages int     `json:"total_pages"`
}
