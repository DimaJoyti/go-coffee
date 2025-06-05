package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderSide represents the side of an order
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderType represents the type of an order
type OrderType string

const (
	OrderTypeMarket    OrderType = "market"
	OrderTypeLimit     OrderType = "limit"
	OrderTypeStop      OrderType = "stop"
	OrderTypeStopLimit OrderType = "stop_limit"
	OrderTypeIOC       OrderType = "ioc"      // Immediate or Cancel
	OrderTypeFOK       OrderType = "fok"      // Fill or Kill
	OrderTypePost      OrderType = "post"     // Post Only
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending          OrderStatus = "pending"
	OrderStatusNew              OrderStatus = "new"
	OrderStatusPartiallyFilled  OrderStatus = "partially_filled"
	OrderStatusFilled           OrderStatus = "filled"
	OrderStatusCanceled         OrderStatus = "canceled"
	OrderStatusRejected         OrderStatus = "rejected"
	OrderStatusExpired          OrderStatus = "expired"
)

// TimeInForce represents how long an order remains active
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "gtc" // Good Till Canceled
	TimeInForceIOC TimeInForce = "ioc" // Immediate or Cancel
	TimeInForceFOK TimeInForce = "fok" // Fill or Kill
	TimeInForceGTD TimeInForce = "gtd" // Good Till Date
)

// Order represents a trading order in the HFT system
type Order struct {
	ID              string          `json:"id" db:"id"`
	ClientOrderID   string          `json:"client_order_id" db:"client_order_id"`
	StrategyID      string          `json:"strategy_id" db:"strategy_id"`
	Symbol          string          `json:"symbol" db:"symbol"`
	Exchange        string          `json:"exchange" db:"exchange"`
	Side            OrderSide       `json:"side" db:"side"`
	Type            OrderType       `json:"type" db:"type"`
	Quantity        decimal.Decimal `json:"quantity" db:"quantity"`
	Price           decimal.Decimal `json:"price" db:"price"`
	StopPrice       decimal.Decimal `json:"stop_price" db:"stop_price"`
	TimeInForce     TimeInForce     `json:"time_in_force" db:"time_in_force"`
	Status          OrderStatus     `json:"status" db:"status"`
	FilledQuantity  decimal.Decimal `json:"filled_quantity" db:"filled_quantity"`
	RemainingQty    decimal.Decimal `json:"remaining_quantity" db:"remaining_quantity"`
	AvgFillPrice    decimal.Decimal `json:"avg_fill_price" db:"avg_fill_price"`
	Commission      decimal.Decimal `json:"commission" db:"commission"`
	CommissionAsset string          `json:"commission_asset" db:"commission_asset"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	ExpiresAt       *time.Time      `json:"expires_at" db:"expires_at"`
	ExchangeOrderID string          `json:"exchange_order_id" db:"exchange_order_id"`
	ErrorMessage    string          `json:"error_message" db:"error_message"`
	Latency         time.Duration   `json:"latency" db:"latency"`
}

// Fill represents a trade execution
type Fill struct {
	ID              string          `json:"id" db:"id"`
	OrderID         string          `json:"order_id" db:"order_id"`
	TradeID         string          `json:"trade_id" db:"trade_id"`
	Symbol          string          `json:"symbol" db:"symbol"`
	Exchange        string          `json:"exchange" db:"exchange"`
	Side            OrderSide       `json:"side" db:"side"`
	Quantity        decimal.Decimal `json:"quantity" db:"quantity"`
	Price           decimal.Decimal `json:"price" db:"price"`
	Commission      decimal.Decimal `json:"commission" db:"commission"`
	CommissionAsset string          `json:"commission_asset" db:"commission_asset"`
	Timestamp       time.Time       `json:"timestamp" db:"timestamp"`
	IsMaker         bool            `json:"is_maker" db:"is_maker"`
}

// Position represents a trading position
type Position struct {
	ID               string          `json:"id" db:"id"`
	StrategyID       string          `json:"strategy_id" db:"strategy_id"`
	Symbol           string          `json:"symbol" db:"symbol"`
	Exchange         string          `json:"exchange" db:"exchange"`
	Side             OrderSide       `json:"side" db:"side"`
	Size             decimal.Decimal `json:"size" db:"size"`
	EntryPrice       decimal.Decimal `json:"entry_price" db:"entry_price"`
	MarkPrice        decimal.Decimal `json:"mark_price" db:"mark_price"`
	UnrealizedPnL    decimal.Decimal `json:"unrealized_pnl" db:"unrealized_pnl"`
	RealizedPnL      decimal.Decimal `json:"realized_pnl" db:"realized_pnl"`
	Margin           decimal.Decimal `json:"margin" db:"margin"`
	MaintenanceMargin decimal.Decimal `json:"maintenance_margin" db:"maintenance_margin"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// MarketDataTick represents ultra-low latency market data
type MarketDataTick struct {
	Symbol       string          `json:"symbol"`
	Exchange     string          `json:"exchange"`
	Price        decimal.Decimal `json:"price"`
	Quantity     decimal.Decimal `json:"quantity"`
	Side         OrderSide       `json:"side"`
	BidPrice     decimal.Decimal `json:"bid_price"`
	BidQuantity  decimal.Decimal `json:"bid_quantity"`
	AskPrice     decimal.Decimal `json:"ask_price"`
	AskQuantity  decimal.Decimal `json:"ask_quantity"`
	Timestamp    time.Time       `json:"timestamp"`
	ReceiveTime  time.Time       `json:"receive_time"`
	ProcessTime  time.Time       `json:"process_time"`
	Latency      time.Duration   `json:"latency"`
	SequenceNum  uint64          `json:"sequence_num"`
}

// OrderBookLevel represents a level in the order book
type OrderBookLevel struct {
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
	Count    int             `json:"count"`
}

// OrderBook represents the current order book state
type OrderBook struct {
	Symbol      string           `json:"symbol"`
	Exchange    string           `json:"exchange"`
	Bids        []OrderBookLevel `json:"bids"`
	Asks        []OrderBookLevel `json:"asks"`
	Timestamp   time.Time        `json:"timestamp"`
	ReceiveTime time.Time        `json:"receive_time"`
	SequenceNum uint64           `json:"sequence_num"`
}

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

// StrategyStatus represents the status of a strategy
type StrategyStatus string

const (
	StrategyStatusStopped StrategyStatus = "stopped"
	StrategyStatusRunning StrategyStatus = "running"
	StrategyStatusPaused  StrategyStatus = "paused"
	StrategyStatusError   StrategyStatus = "error"
)

// Strategy represents a trading strategy
type Strategy struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Type        StrategyType           `json:"type" db:"type"`
	Status      StrategyStatus         `json:"status" db:"status"`
	Symbols     []string               `json:"symbols" db:"symbols"`
	Exchanges   []string               `json:"exchanges" db:"exchanges"`
	Parameters  map[string]any `json:"parameters" db:"parameters"`
	RiskLimits  RiskLimits             `json:"risk_limits" db:"risk_limits"`
	Performance StrategyPerformance    `json:"performance" db:"performance"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time             `json:"started_at" db:"started_at"`
	StoppedAt   *time.Time             `json:"stopped_at" db:"stopped_at"`
}

// RiskLimits represents risk management limits for a strategy
type RiskLimits struct {
	MaxPositionSize    decimal.Decimal `json:"max_position_size" db:"max_position_size"`
	MaxDailyLoss       decimal.Decimal `json:"max_daily_loss" db:"max_daily_loss"`
	MaxDrawdown        decimal.Decimal `json:"max_drawdown" db:"max_drawdown"`
	MaxOrderSize       decimal.Decimal `json:"max_order_size" db:"max_order_size"`
	MaxOrdersPerSecond int             `json:"max_orders_per_second" db:"max_orders_per_second"`
	MaxExposure        decimal.Decimal `json:"max_exposure" db:"max_exposure"`
	StopLossPercent    decimal.Decimal `json:"stop_loss_percent" db:"stop_loss_percent"`
	TakeProfitPercent  decimal.Decimal `json:"take_profit_percent" db:"take_profit_percent"`
}

// StrategyPerformance represents performance metrics for a strategy
type StrategyPerformance struct {
	TotalPnL        decimal.Decimal `json:"total_pnl" db:"total_pnl"`
	DailyPnL        decimal.Decimal `json:"daily_pnl" db:"daily_pnl"`
	TotalTrades     int64           `json:"total_trades" db:"total_trades"`
	WinningTrades   int64           `json:"winning_trades" db:"winning_trades"`
	LosingTrades    int64           `json:"losing_trades" db:"losing_trades"`
	WinRate         decimal.Decimal `json:"win_rate" db:"win_rate"`
	AvgWin          decimal.Decimal `json:"avg_win" db:"avg_win"`
	AvgLoss         decimal.Decimal `json:"avg_loss" db:"avg_loss"`
	ProfitFactor    decimal.Decimal `json:"profit_factor" db:"profit_factor"`
	SharpeRatio     decimal.Decimal `json:"sharpe_ratio" db:"sharpe_ratio"`
	MaxDrawdown     decimal.Decimal `json:"max_drawdown" db:"max_drawdown"`
	VolumeTraded    decimal.Decimal `json:"volume_traded" db:"volume_traded"`
	AvgLatency      time.Duration   `json:"avg_latency" db:"avg_latency"`
	LastUpdated     time.Time       `json:"last_updated" db:"last_updated"`
}

// Signal represents a trading signal generated by a strategy
type Signal struct {
	ID         string          `json:"id" db:"id"`
	StrategyID string          `json:"strategy_id" db:"strategy_id"`
	Symbol     string          `json:"symbol" db:"symbol"`
	Exchange   string          `json:"exchange" db:"exchange"`
	Side       OrderSide       `json:"side" db:"side"`
	Strength   decimal.Decimal `json:"strength" db:"strength"`
	Price      decimal.Decimal `json:"price" db:"price"`
	Quantity   decimal.Decimal `json:"quantity" db:"quantity"`
	Confidence decimal.Decimal `json:"confidence" db:"confidence"`
	Reason     string          `json:"reason" db:"reason"`
	Metadata   map[string]any `json:"metadata" db:"metadata"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	ExpiresAt  time.Time       `json:"expires_at" db:"expires_at"`
	Executed   bool            `json:"executed" db:"executed"`
}

// RiskEvent represents a risk management event
type RiskEvent struct {
	ID          string                 `json:"id" db:"id"`
	Type        string                 `json:"type" db:"type"`
	Severity    string                 `json:"severity" db:"severity"`
	StrategyID  string                 `json:"strategy_id" db:"strategy_id"`
	Symbol      string                 `json:"symbol" db:"symbol"`
	Description string                 `json:"description" db:"description"`
	Data        map[string]any `json:"data" db:"data"`
	Action      string                 `json:"action" db:"action"`
	Resolved    bool                   `json:"resolved" db:"resolved"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	ResolvedAt  *time.Time             `json:"resolved_at" db:"resolved_at"`
}
