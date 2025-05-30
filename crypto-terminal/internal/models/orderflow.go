package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Tick represents a single trade tick with buy/sell classification
type Tick struct {
	ID          string          `json:"id" db:"id"`
	Symbol      string          `json:"symbol" db:"symbol"`
	Price       decimal.Decimal `json:"price" db:"price"`
	Volume      decimal.Decimal `json:"volume" db:"volume"`
	Side        string          `json:"side" db:"side"` // BUY, SELL, UNKNOWN
	TradeID     string          `json:"trade_id" db:"trade_id"`
	Exchange    string          `json:"exchange" db:"exchange"`
	Timestamp   time.Time       `json:"timestamp" db:"timestamp"`
	BidPrice    decimal.Decimal `json:"bid_price" db:"bid_price"`
	AskPrice    decimal.Decimal `json:"ask_price" db:"ask_price"`
	IsAggressor bool            `json:"is_aggressor" db:"is_aggressor"`
	Sequence    int64           `json:"sequence" db:"sequence"`
}

// FootprintBar represents aggregated order flow data for a price level and time period
type FootprintBar struct {
	ID              string          `json:"id" db:"id"`
	Symbol          string          `json:"symbol" db:"symbol"`
	Timeframe       string          `json:"timeframe" db:"timeframe"`
	PriceLevel      decimal.Decimal `json:"price_level" db:"price_level"`
	TickSize        decimal.Decimal `json:"tick_size" db:"tick_size"`
	BuyVolume       decimal.Decimal `json:"buy_volume" db:"buy_volume"`
	SellVolume      decimal.Decimal `json:"sell_volume" db:"sell_volume"`
	TotalVolume     decimal.Decimal `json:"total_volume" db:"total_volume"`
	Delta           decimal.Decimal `json:"delta" db:"delta"`
	BuyTrades       int             `json:"buy_trades" db:"buy_trades"`
	SellTrades      int             `json:"sell_trades" db:"sell_trades"`
	TotalTrades     int             `json:"total_trades" db:"total_trades"`
	MaxBuyVolume    decimal.Decimal `json:"max_buy_volume" db:"max_buy_volume"`
	MaxSellVolume   decimal.Decimal `json:"max_sell_volume" db:"max_sell_volume"`
	VolumeImbalance decimal.Decimal `json:"volume_imbalance" db:"volume_imbalance"`
	IsImbalanced    bool            `json:"is_imbalanced" db:"is_imbalanced"`
	IsPointOfControl bool           `json:"is_point_of_control" db:"is_point_of_control"`
	StartTime       time.Time       `json:"start_time" db:"start_time"`
	EndTime         time.Time       `json:"end_time" db:"end_time"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// VolumeProfile represents volume distribution across price levels
type VolumeProfile struct {
	ID              string                 `json:"id" db:"id"`
	Symbol          string                 `json:"symbol" db:"symbol"`
	ProfileType     string                 `json:"profile_type" db:"profile_type"` // VPSV, VPVR
	StartTime       time.Time              `json:"start_time" db:"start_time"`
	EndTime         time.Time              `json:"end_time" db:"end_time"`
	HighPrice       decimal.Decimal        `json:"high_price" db:"high_price"`
	LowPrice        decimal.Decimal        `json:"low_price" db:"low_price"`
	TotalVolume     decimal.Decimal        `json:"total_volume" db:"total_volume"`
	PointOfControl  decimal.Decimal        `json:"point_of_control" db:"point_of_control"`
	ValueAreaHigh   decimal.Decimal        `json:"value_area_high" db:"value_area_high"`
	ValueAreaLow    decimal.Decimal        `json:"value_area_low" db:"value_area_low"`
	ValueAreaVolume decimal.Decimal        `json:"value_area_volume" db:"value_area_volume"`
	PriceLevels     []VolumeProfileLevel   `json:"price_levels,omitempty"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// VolumeProfileLevel represents volume at a specific price level
type VolumeProfileLevel struct {
	Price       decimal.Decimal `json:"price" db:"price"`
	Volume      decimal.Decimal `json:"volume" db:"volume"`
	BuyVolume   decimal.Decimal `json:"buy_volume" db:"buy_volume"`
	SellVolume  decimal.Decimal `json:"sell_volume" db:"sell_volume"`
	Delta       decimal.Decimal `json:"delta" db:"delta"`
	TradeCount  int             `json:"trade_count" db:"trade_count"`
	Percentage  decimal.Decimal `json:"percentage" db:"percentage"`
	IsHVN       bool            `json:"is_hvn" db:"is_hvn"` // High Volume Node
	IsLVN       bool            `json:"is_lvn" db:"is_lvn"` // Low Volume Node
	IsPOC       bool            `json:"is_poc" db:"is_poc"` // Point of Control
	IsValueArea bool            `json:"is_value_area" db:"is_value_area"`
}

// DeltaProfile represents cumulative delta analysis
type DeltaProfile struct {
	ID                string          `json:"id" db:"id"`
	Symbol            string          `json:"symbol" db:"symbol"`
	Timeframe         string          `json:"timeframe" db:"timeframe"`
	StartTime         time.Time       `json:"start_time" db:"start_time"`
	EndTime           time.Time       `json:"end_time" db:"end_time"`
	CumulativeDelta   decimal.Decimal `json:"cumulative_delta" db:"cumulative_delta"`
	DeltaHigh         decimal.Decimal `json:"delta_high" db:"delta_high"`
	DeltaLow          decimal.Decimal `json:"delta_low" db:"delta_low"`
	DeltaRange        decimal.Decimal `json:"delta_range" db:"delta_range"`
	DeltaMomentum     decimal.Decimal `json:"delta_momentum" db:"delta_momentum"`
	DeltaAcceleration decimal.Decimal `json:"delta_acceleration" db:"delta_acceleration"`
	BuyPressure       decimal.Decimal `json:"buy_pressure" db:"buy_pressure"`
	SellPressure      decimal.Decimal `json:"sell_pressure" db:"sell_pressure"`
	NetPressure       decimal.Decimal `json:"net_pressure" db:"net_pressure"`
	DeltaStrength     decimal.Decimal `json:"delta_strength" db:"delta_strength"`
	IsDivergent       bool            `json:"is_divergent" db:"is_divergent"`
	IsExhausted       bool            `json:"is_exhausted" db:"is_exhausted"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
}

// OrderFlowImbalance represents detected order flow imbalances
type OrderFlowImbalance struct {
	ID              string          `json:"id" db:"id"`
	Symbol          string          `json:"symbol" db:"symbol"`
	Price           decimal.Decimal `json:"price" db:"price"`
	ImbalanceType   string          `json:"imbalance_type" db:"imbalance_type"` // BID_STACK, ASK_STACK, VOLUME_IMBALANCE
	Severity        string          `json:"severity" db:"severity"` // LOW, MEDIUM, HIGH, EXTREME
	BuyVolume       decimal.Decimal `json:"buy_volume" db:"buy_volume"`
	SellVolume      decimal.Decimal `json:"sell_volume" db:"sell_volume"`
	ImbalanceRatio  decimal.Decimal `json:"imbalance_ratio" db:"imbalance_ratio"`
	Duration        time.Duration   `json:"duration" db:"duration"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	IsResolved      bool            `json:"is_resolved" db:"is_resolved"`
	ResolutionType  string          `json:"resolution_type" db:"resolution_type"` // ABSORPTION, CONTINUATION, REVERSAL
	DetectedAt      time.Time       `json:"detected_at" db:"detected_at"`
	ResolvedAt      *time.Time      `json:"resolved_at" db:"resolved_at"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// OrderFlowSession represents a trading session's order flow summary
type OrderFlowSession struct {
	ID                  string          `json:"id" db:"id"`
	Symbol              string          `json:"symbol" db:"symbol"`
	SessionType         string          `json:"session_type" db:"session_type"` // ASIAN, LONDON, NEW_YORK, OVERLAP
	StartTime           time.Time       `json:"start_time" db:"start_time"`
	EndTime             time.Time       `json:"end_time" db:"end_time"`
	OpenPrice           decimal.Decimal `json:"open_price" db:"open_price"`
	HighPrice           decimal.Decimal `json:"high_price" db:"high_price"`
	LowPrice            decimal.Decimal `json:"low_price" db:"low_price"`
	ClosePrice          decimal.Decimal `json:"close_price" db:"close_price"`
	TotalVolume         decimal.Decimal `json:"total_volume" db:"total_volume"`
	BuyVolume           decimal.Decimal `json:"buy_volume" db:"buy_volume"`
	SellVolume          decimal.Decimal `json:"sell_volume" db:"sell_volume"`
	SessionDelta        decimal.Decimal `json:"session_delta" db:"session_delta"`
	VolumeWeightedPrice decimal.Decimal `json:"volume_weighted_price" db:"volume_weighted_price"`
	PointOfControl      decimal.Decimal `json:"point_of_control" db:"point_of_control"`
	ValueAreaHigh       decimal.Decimal `json:"value_area_high" db:"value_area_high"`
	ValueAreaLow        decimal.Decimal `json:"value_area_low" db:"value_area_low"`
	ImbalanceCount      int             `json:"imbalance_count" db:"imbalance_count"`
	AbsorptionCount     int             `json:"absorption_count" db:"absorption_count"`
	BreakoutCount       int             `json:"breakout_count" db:"breakout_count"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
}

// OrderFlowConfig represents configuration for order flow analysis
type OrderFlowConfig struct {
	Symbol                  string          `json:"symbol"`
	TickAggregationMethod   string          `json:"tick_aggregation_method"` // TIME, VOLUME, TICK_COUNT
	TicksPerRow             int             `json:"ticks_per_row"`
	VolumePerRow            decimal.Decimal `json:"volume_per_row"`
	TimePerRow              time.Duration   `json:"time_per_row"`
	PriceTickSize           decimal.Decimal `json:"price_tick_size"`
	ImbalanceThreshold      decimal.Decimal `json:"imbalance_threshold"`
	ImbalanceMinVolume      decimal.Decimal `json:"imbalance_min_volume"`
	ValueAreaPercentage     decimal.Decimal `json:"value_area_percentage"`
	HVNThreshold            decimal.Decimal `json:"hvn_threshold"`
	LVNThreshold            decimal.Decimal `json:"lvn_threshold"`
	DeltaSmoothingPeriod    int             `json:"delta_smoothing_period"`
	EnableRealTimeUpdates   bool            `json:"enable_real_time_updates"`
	EnableImbalanceDetection bool           `json:"enable_imbalance_detection"`
	EnableDeltaDivergence   bool            `json:"enable_delta_divergence"`
}

// OrderFlowMetrics represents real-time order flow metrics
type OrderFlowMetrics struct {
	Symbol              string          `json:"symbol"`
	Timestamp           time.Time       `json:"timestamp"`
	CurrentPrice        decimal.Decimal `json:"current_price"`
	BidPrice            decimal.Decimal `json:"bid_price"`
	AskPrice            decimal.Decimal `json:"ask_price"`
	BidSize             decimal.Decimal `json:"bid_size"`
	AskSize             decimal.Decimal `json:"ask_size"`
	LastTradeVolume     decimal.Decimal `json:"last_trade_volume"`
	LastTradeSide       string          `json:"last_trade_side"`
	CumulativeDelta     decimal.Decimal `json:"cumulative_delta"`
	SessionDelta        decimal.Decimal `json:"session_delta"`
	VolumeImbalance     decimal.Decimal `json:"volume_imbalance"`
	BuyPressure         decimal.Decimal `json:"buy_pressure"`
	SellPressure        decimal.Decimal `json:"sell_pressure"`
	OrderFlowMomentum   decimal.Decimal `json:"order_flow_momentum"`
	LiquidityIndex      decimal.Decimal `json:"liquidity_index"`
	MarketDepth         decimal.Decimal `json:"market_depth"`
	ActiveImbalances    int             `json:"active_imbalances"`
	RecentAbsorptions   int             `json:"recent_absorptions"`
}

// OrderFlowAlert represents order flow-based alerts
type OrderFlowAlert struct {
	ID          string          `json:"id" db:"id"`
	UserID      string          `json:"user_id" db:"user_id"`
	Symbol      string          `json:"symbol" db:"symbol"`
	AlertType   string          `json:"alert_type" db:"alert_type"` // IMBALANCE, ABSORPTION, DELTA_DIVERGENCE, POC_BREAK
	Condition   string          `json:"condition" db:"condition"`
	Threshold   decimal.Decimal `json:"threshold" db:"threshold"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	IsTriggered bool            `json:"is_triggered" db:"is_triggered"`
	Message     string          `json:"message" db:"message"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	TriggeredAt *time.Time      `json:"triggered_at" db:"triggered_at"`
}

// OrderFlowWebSocketMessage represents real-time order flow data
type OrderFlowWebSocketMessage struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// FootprintChartData represents data for footprint chart visualization
type FootprintChartData struct {
	Symbol      string         `json:"symbol"`
	Timeframe   string         `json:"timeframe"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	Bars        []FootprintBar `json:"bars"`
	Config      OrderFlowConfig `json:"config"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// VolumeProfileChartData represents data for volume profile visualization
type VolumeProfileChartData struct {
	Symbol       string        `json:"symbol"`
	ProfileType  string        `json:"profile_type"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Profile      VolumeProfile `json:"profile"`
	PriceLevels  []VolumeProfileLevel `json:"price_levels"`
	Config       OrderFlowConfig `json:"config"`
}

// DeltaAnalysisData represents delta analysis results
type DeltaAnalysisData struct {
	Symbol          string        `json:"symbol"`
	Timeframe       string        `json:"timeframe"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	DeltaProfile    DeltaProfile  `json:"delta_profile"`
	DeltaHistory    []DeltaProfile `json:"delta_history"`
	Divergences     []OrderFlowImbalance `json:"divergences"`
	Config          OrderFlowConfig `json:"config"`
}
