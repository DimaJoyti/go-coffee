package exchanges

import (
	"time"

	"github.com/shopspring/decimal"
)

// ExchangeType represents different cryptocurrency exchanges
type ExchangeType string

const (
	ExchangeBinance  ExchangeType = "binance"
	ExchangeCoinbase ExchangeType = "coinbase"
	ExchangeKraken   ExchangeType = "kraken"
	ExchangeFTX      ExchangeType = "ftx"
	ExchangeBybit    ExchangeType = "bybit"
	ExchangeOKX      ExchangeType = "okx"
)

// OrderType represents different order types
type OrderType string

const (
	OrderTypeMarket          OrderType = "market"
	OrderTypeLimit           OrderType = "limit"
	OrderTypeStop            OrderType = "stop"
	OrderTypeStopLimit       OrderType = "stop_limit"
	OrderTypeTakeProfit      OrderType = "take_profit"
	OrderTypeTakeProfitLimit OrderType = "take_profit_limit"
	OrderTypeOCO             OrderType = "oco"
	OrderTypeTrailingStop    OrderType = "trailing_stop"
)

// OrderSide represents buy or sell side
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusNew      OrderStatus = "new"
	OrderStatusFilled   OrderStatus = "filled"
	OrderStatusCanceled OrderStatus = "canceled"
	OrderStatusRejected OrderStatus = "rejected"
	OrderStatusExpired  OrderStatus = "expired"
)

// Symbol represents a trading pair
type Symbol struct {
	Base   string `json:"base"`   // e.g., "BTC"
	Quote  string `json:"quote"`  // e.g., "USDT"
	Symbol string `json:"symbol"` // e.g., "BTCUSDT"
}

// Ticker represents real-time ticker data
type Ticker struct {
	Exchange         ExchangeType    `json:"exchange"`
	Symbol           string          `json:"symbol"`
	LastPrice        decimal.Decimal `json:"last_price"`
	BidPrice         decimal.Decimal `json:"bid_price"`
	AskPrice         decimal.Decimal `json:"ask_price"`
	Volume24h        decimal.Decimal `json:"volume_24h"`
	VolumeQuote24h   decimal.Decimal `json:"volume_quote_24h"`
	Change24h        decimal.Decimal `json:"change_24h"`
	ChangePercent24h decimal.Decimal `json:"change_percent_24h"`
	High24h          decimal.Decimal `json:"high_24h"`
	Low24h           decimal.Decimal `json:"low_24h"`
	OpenPrice        decimal.Decimal `json:"open_price"`
	Timestamp        time.Time       `json:"timestamp"`
	Count            int64           `json:"count"` // Number of trades
}

// OrderBook represents market depth data
type OrderBook struct {
	Exchange     ExchangeType     `json:"exchange"`
	Symbol       string           `json:"symbol"`
	Bids         []OrderBookLevel `json:"bids"`
	Asks         []OrderBookLevel `json:"asks"`
	Timestamp    time.Time        `json:"timestamp"`
	LastUpdateID int64            `json:"last_update_id"`
}

// OrderBookLevel represents a single level in the order book
type OrderBookLevel struct {
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
	Count    int             `json:"count,omitempty"` // Number of orders (if available)
}

// Trade represents a completed trade
type Trade struct {
	ID        string          `json:"id"`
	Exchange  ExchangeType    `json:"exchange"`
	Symbol    string          `json:"symbol"`
	Price     decimal.Decimal `json:"price"`
	Quantity  decimal.Decimal `json:"quantity"`
	Side      OrderSide       `json:"side"`
	Timestamp time.Time       `json:"timestamp"`
	IsMaker   bool            `json:"is_maker"`
}

// Kline represents candlestick/OHLCV data
type Kline struct {
	Exchange            ExchangeType    `json:"exchange"`
	Symbol              string          `json:"symbol"`
	Interval            string          `json:"interval"` // 1m, 5m, 1h, 1d, etc.
	OpenTime            time.Time       `json:"open_time"`
	CloseTime           time.Time       `json:"close_time"`
	Open                decimal.Decimal `json:"open"`
	High                decimal.Decimal `json:"high"`
	Low                 decimal.Decimal `json:"low"`
	Close               decimal.Decimal `json:"close"`
	Volume              decimal.Decimal `json:"volume"`
	QuoteVolume         decimal.Decimal `json:"quote_volume"`
	TradeCount          int64           `json:"trade_count"`
	TakerBuyBaseVolume  decimal.Decimal `json:"taker_buy_base_volume"`
	TakerBuyQuoteVolume decimal.Decimal `json:"taker_buy_quote_volume"`
}

// Order represents an order on an exchange
type Order struct {
	ID                string          `json:"id"`
	ClientOrderID     string          `json:"client_order_id"`
	Exchange          ExchangeType    `json:"exchange"`
	Symbol            string          `json:"symbol"`
	Type              OrderType       `json:"type"`
	Side              OrderSide       `json:"side"`
	Status            OrderStatus     `json:"status"`
	Price             decimal.Decimal `json:"price"`
	Quantity          decimal.Decimal `json:"quantity"`
	FilledQuantity    decimal.Decimal `json:"filled_quantity"`
	RemainingQuantity decimal.Decimal `json:"remaining_quantity"`
	AveragePrice      decimal.Decimal `json:"average_price"`
	Commission        decimal.Decimal `json:"commission"`
	CommissionAsset   string          `json:"commission_asset"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	TimeInForce       string          `json:"time_in_force"`
	StopPrice         decimal.Decimal `json:"stop_price,omitempty"`
}

// Balance represents account balance
type Balance struct {
	Exchange  ExchangeType    `json:"exchange"`
	Asset     string          `json:"asset"`
	Free      decimal.Decimal `json:"free"`
	Locked    decimal.Decimal `json:"locked"`
	Total     decimal.Decimal `json:"total"`
	Timestamp time.Time       `json:"timestamp"`
}

// ExchangeInfo represents exchange information
type ExchangeInfo struct {
	Exchange    ExchangeType `json:"exchange"`
	Name        string       `json:"name"`
	Status      string       `json:"status"`
	Symbols     []Symbol     `json:"symbols"`
	ServerTime  time.Time    `json:"server_time"`
	RateLimits  []RateLimit  `json:"rate_limits"`
	Permissions []string     `json:"permissions"`
}

// RateLimit represents API rate limiting information
type RateLimit struct {
	Type     string `json:"type"`     // REQUEST_WEIGHT, ORDERS, RAW_REQUESTS
	Interval string `json:"interval"` // SECOND, MINUTE, DAY
	Limit    int    `json:"limit"`
}

// ArbitrageOpportunity represents a potential arbitrage opportunity
type ArbitrageOpportunity struct {
	Symbol          string          `json:"symbol"`
	BuyExchange     ExchangeType    `json:"buy_exchange"`
	SellExchange    ExchangeType    `json:"sell_exchange"`
	BuyPrice        decimal.Decimal `json:"buy_price"`
	SellPrice       decimal.Decimal `json:"sell_price"`
	PriceDifference decimal.Decimal `json:"price_difference"`
	ProfitPercent   decimal.Decimal `json:"profit_percent"`
	Volume          decimal.Decimal `json:"volume"`
	Timestamp       time.Time       `json:"timestamp"`
	Confidence      float64         `json:"confidence"` // 0-1 confidence score
}

// MarketSummary represents aggregated market data across exchanges
type MarketSummary struct {
	Symbol         string                   `json:"symbol"`
	BestBid        *ExchangePrice           `json:"best_bid"`
	BestAsk        *ExchangePrice           `json:"best_ask"`
	WeightedPrice  decimal.Decimal          `json:"weighted_price"`
	TotalVolume24h decimal.Decimal          `json:"total_volume_24h"`
	PriceSpread    decimal.Decimal          `json:"price_spread"`
	SpreadPercent  decimal.Decimal          `json:"spread_percent"`
	ExchangePrices map[ExchangeType]*Ticker `json:"exchange_prices"`
	Arbitrage      *ArbitrageOpportunity    `json:"arbitrage,omitempty"`
	Timestamp      time.Time                `json:"timestamp"`
	DataQuality    float64                  `json:"data_quality"` // 0-1 quality score
}

// ExchangePrice represents price from a specific exchange
type ExchangePrice struct {
	Exchange  ExchangeType    `json:"exchange"`
	Price     decimal.Decimal `json:"price"`
	Volume    decimal.Decimal `json:"volume"`
	Timestamp time.Time       `json:"timestamp"`
}

// StreamData represents real-time streaming data
type StreamData struct {
	Type      string       `json:"type"` // ticker, trade, orderbook, kline
	Exchange  ExchangeType `json:"exchange"`
	Symbol    string       `json:"symbol"`
	Data      interface{}  `json:"data"`
	Timestamp time.Time    `json:"timestamp"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Stream string      `json:"stream"`
	Data   interface{} `json:"data"`
}

// SubscriptionRequest represents a subscription request
type SubscriptionRequest struct {
	Exchange ExchangeType `json:"exchange"`
	Symbol   string       `json:"symbol"`
	Type     string       `json:"type"`               // ticker, trade, orderbook, kline
	Interval string       `json:"interval,omitempty"` // for kline
}

// DataQualityMetrics represents data quality metrics
type DataQualityMetrics struct {
	Exchange        ExchangeType  `json:"exchange"`
	Symbol          string        `json:"symbol"`
	LastUpdate      time.Time     `json:"last_update"`
	UpdateFrequency float64       `json:"update_frequency"` // updates per second
	Latency         time.Duration `json:"latency"`
	ErrorRate       float64       `json:"error_rate"`
	Availability    float64       `json:"availability"`
	QualityScore    float64       `json:"quality_score"` // 0-1 overall quality
}

// FuturesPositionSide represents futures position side
type FuturesPositionSide string

const (
	PositionSideLong  FuturesPositionSide = "LONG"
	PositionSideShort FuturesPositionSide = "SHORT"
	PositionSideBoth  FuturesPositionSide = "BOTH"
)

// MarginType represents margin type for futures
type MarginType string

const (
	MarginTypeIsolated MarginType = "ISOLATED"
	MarginTypeCrossed  MarginType = "CROSSED"
)

// FuturesOrder represents a futures order
type FuturesOrder struct {
	ID                string              `json:"id"`
	ClientOrderID     string              `json:"client_order_id"`
	Exchange          ExchangeType        `json:"exchange"`
	Symbol            string              `json:"symbol"`
	Type              OrderType           `json:"type"`
	Side              OrderSide           `json:"side"`
	PositionSide      FuturesPositionSide `json:"position_side"`
	Status            OrderStatus         `json:"status"`
	Price             decimal.Decimal     `json:"price"`
	Quantity          decimal.Decimal     `json:"quantity"`
	FilledQuantity    decimal.Decimal     `json:"filled_quantity"`
	RemainingQuantity decimal.Decimal     `json:"remaining_quantity"`
	AveragePrice      decimal.Decimal     `json:"average_price"`
	Commission        decimal.Decimal     `json:"commission"`
	CommissionAsset   string              `json:"commission_asset"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	TimeInForce       string              `json:"time_in_force"`
	StopPrice         decimal.Decimal     `json:"stop_price,omitempty"`
	ActivatePrice     decimal.Decimal     `json:"activate_price,omitempty"`
	CallbackRate      decimal.Decimal     `json:"callback_rate,omitempty"`
	ReduceOnly        bool                `json:"reduce_only"`
	ClosePosition     bool                `json:"close_position"`
}

// FuturesPosition represents a futures position
type FuturesPosition struct {
	Exchange       ExchangeType        `json:"exchange"`
	Symbol         string              `json:"symbol"`
	PositionSide   FuturesPositionSide `json:"position_side"`
	Size           decimal.Decimal     `json:"size"`
	EntryPrice     decimal.Decimal     `json:"entry_price"`
	MarkPrice      decimal.Decimal     `json:"mark_price"`
	UnrealizedPnL  decimal.Decimal     `json:"unrealized_pnl"`
	RealizedPnL    decimal.Decimal     `json:"realized_pnl"`
	Leverage       decimal.Decimal     `json:"leverage"`
	MarginType     MarginType          `json:"margin_type"`
	IsolatedMargin decimal.Decimal     `json:"isolated_margin"`
	Timestamp      time.Time           `json:"timestamp"`
}

// FuturesAccount represents futures account information
type FuturesAccount struct {
	Exchange           ExchangeType      `json:"exchange"`
	TotalWalletBalance decimal.Decimal   `json:"total_wallet_balance"`
	TotalUnrealizedPnL decimal.Decimal   `json:"total_unrealized_pnl"`
	TotalMarginBalance decimal.Decimal   `json:"total_margin_balance"`
	TotalPositionValue decimal.Decimal   `json:"total_position_value"`
	AvailableBalance   decimal.Decimal   `json:"available_balance"`
	MaxWithdrawAmount  decimal.Decimal   `json:"max_withdraw_amount"`
	Assets             []FuturesAsset    `json:"assets"`
	Positions          []FuturesPosition `json:"positions"`
	CanTrade           bool              `json:"can_trade"`
	CanDeposit         bool              `json:"can_deposit"`
	CanWithdraw        bool              `json:"can_withdraw"`
	UpdateTime         time.Time         `json:"update_time"`
}

// FuturesAsset represents a futures asset balance
type FuturesAsset struct {
	Asset                  string          `json:"asset"`
	WalletBalance          decimal.Decimal `json:"wallet_balance"`
	UnrealizedProfit       decimal.Decimal `json:"unrealized_profit"`
	MarginBalance          decimal.Decimal `json:"margin_balance"`
	MaintMargin            decimal.Decimal `json:"maint_margin"`
	InitialMargin          decimal.Decimal `json:"initial_margin"`
	PositionInitialMargin  decimal.Decimal `json:"position_initial_margin"`
	OpenOrderInitialMargin decimal.Decimal `json:"open_order_initial_margin"`
	MaxWithdrawAmount      decimal.Decimal `json:"max_withdraw_amount"`
	CrossWalletBalance     decimal.Decimal `json:"cross_wallet_balance"`
	CrossUnPnl             decimal.Decimal `json:"cross_un_pnl"`
	AvailableBalance       decimal.Decimal `json:"available_balance"`
}

// OCOOrder represents an OCO (One-Cancels-Other) order
type OCOOrder struct {
	ID                string    `json:"id"`
	ListOrderStatus   string    `json:"list_order_status"`
	ListStatusType    string    `json:"list_status_type"`
	ListClientOrderID string    `json:"list_client_order_id"`
	TransactionTime   time.Time `json:"transaction_time"`
	Symbol            string    `json:"symbol"`
	Orders            []Order   `json:"orders"`
	OrderReports      []Order   `json:"order_reports"`
}

// MarginAccount represents margin account information
type MarginAccount struct {
	Exchange            ExchangeType    `json:"exchange"`
	BorrowEnabled       bool            `json:"borrow_enabled"`
	MarginLevel         decimal.Decimal `json:"margin_level"`
	TotalAssetOfBtc     decimal.Decimal `json:"total_asset_of_btc"`
	TotalLiabilityOfBtc decimal.Decimal `json:"total_liability_of_btc"`
	TotalNetAssetOfBtc  decimal.Decimal `json:"total_net_asset_of_btc"`
	TradeEnabled        bool            `json:"trade_enabled"`
	TransferEnabled     bool            `json:"transfer_enabled"`
	UserAssets          []MarginAsset   `json:"user_assets"`
	UpdateTime          time.Time       `json:"update_time"`
}

// MarginAsset represents a margin asset
type MarginAsset struct {
	Asset    string          `json:"asset"`
	Borrowed decimal.Decimal `json:"borrowed"`
	Free     decimal.Decimal `json:"free"`
	Interest decimal.Decimal `json:"interest"`
	Locked   decimal.Decimal `json:"locked"`
	NetAsset decimal.Decimal `json:"net_asset"`
}
