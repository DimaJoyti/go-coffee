package models

import (
	"time"
)

// TradingSymbol represents a trading pair symbol information
type TradingSymbol struct {
	Symbol     string  `json:"symbol"`
	BaseAsset  string  `json:"baseAsset"`
	QuoteAsset string  `json:"quoteAsset"`
	Status     string  `json:"status"`
	MinPrice   float64 `json:"minPrice"`
	MaxPrice   float64 `json:"maxPrice"`
	TickSize   float64 `json:"tickSize"`
	MinQty     float64 `json:"minQty"`
	MaxQty     float64 `json:"maxQty"`
	StepSize   float64 `json:"stepSize"`
}

// MarketDepth represents order book depth
type MarketDepth struct {
	Symbol       string       `json:"symbol"`
	LastUpdateID int64        `json:"lastUpdateId"`
	Bids         []PriceLevel `json:"bids"`
	Asks         []PriceLevel `json:"asks"`
}

// PriceLevel represents a price level in the order book
type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	Count    int     `json:"count,omitempty"`
}

// Trade represents a trade execution
type Trade struct {
	ID       string  `json:"id"`
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	Time     int64   `json:"time"`
	IsBuyer  bool    `json:"isBuyerMaker"`
}

// Ticker24hr represents 24hr ticker statistics
type Ticker24hr struct {
	Symbol             string  `json:"symbol"`
	PriceChange        float64 `json:"priceChange"`
	PriceChangePercent float64 `json:"priceChangePercent"`
	WeightedAvgPrice   float64 `json:"weightedAvgPrice"`
	PrevClosePrice     float64 `json:"prevClosePrice"`
	LastPrice          float64 `json:"lastPrice"`
	LastQty            float64 `json:"lastQty"`
	BidPrice           float64 `json:"bidPrice"`
	BidQty             float64 `json:"bidQty"`
	AskPrice           float64 `json:"askPrice"`
	AskQty             float64 `json:"askQty"`
	OpenPrice          float64 `json:"openPrice"`
	HighPrice          float64 `json:"highPrice"`
	LowPrice           float64 `json:"lowPrice"`
	Volume             float64 `json:"volume"`
	QuoteVolume        float64 `json:"quoteVolume"`
	OpenTime           int64   `json:"openTime"`
	CloseTime          int64   `json:"closeTime"`
	Count              int64   `json:"count"`
}

// Kline represents candlestick data
type Kline struct {
	OpenTime                 int64   `json:"openTime"`
	Open                     float64 `json:"open"`
	High                     float64 `json:"high"`
	Low                      float64 `json:"low"`
	Close                    float64 `json:"close"`
	Volume                   float64 `json:"volume"`
	CloseTime                int64   `json:"closeTime"`
	QuoteAssetVolume         float64 `json:"quoteAssetVolume"`
	NumberOfTrades           int64   `json:"numberOfTrades"`
	TakerBuyBaseAssetVolume  float64 `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume float64 `json:"takerBuyQuoteAssetVolume"`
}

// OrderRequest represents a new order request
type OrderRequest struct {
	Symbol        string  `json:"symbol" binding:"required"`
	Side          string  `json:"side" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	TimeInForce   string  `json:"timeInForce,omitempty"`
	Quantity      float64 `json:"quantity" binding:"required"`
	Price         float64 `json:"price,omitempty"`
	StopPrice     float64 `json:"stopPrice,omitempty"`
	IcebergQty    float64 `json:"icebergQty,omitempty"`
	ClientOrderID string  `json:"newClientOrderId,omitempty"`
	ReduceOnly    bool    `json:"reduceOnly,omitempty"`
	ClosePosition bool    `json:"closePosition,omitempty"`
}

// TradingOrder represents a trading order (different from existing Order model)
type TradingOrder struct {
	OrderID             string    `json:"orderId"`
	ClientOrderID       string    `json:"clientOrderId"`
	Symbol              string    `json:"symbol"`
	Side                string    `json:"side"`
	Type                string    `json:"type"`
	TimeInForce         string    `json:"timeInForce"`
	Quantity            float64   `json:"origQty"`
	Price               float64   `json:"price"`
	StopPrice           float64   `json:"stopPrice,omitempty"`
	Status              string    `json:"status"`
	TransactTime        int64     `json:"transactTime"`
	ExecutedQty         float64   `json:"executedQty"`
	CummulativeQuoteQty float64   `json:"cummulativeQuoteQty"`
	AvgPrice            float64   `json:"avgPrice,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// CancelOrderResult represents the result of canceling an order
type CancelOrderResult struct {
	OrderID       string `json:"orderId"`
	Symbol        string `json:"symbol"`
	Status        string `json:"status"`
	ClientOrderID string `json:"clientOrderId"`
	TransactTime  int64  `json:"transactTime"`
}

// TradingPosition represents a trading position
type TradingPosition struct {
	Symbol           string  `json:"symbol"`
	PositionAmt      float64 `json:"positionAmt"`
	EntryPrice       float64 `json:"entryPrice"`
	MarkPrice        float64 `json:"markPrice"`
	UnRealizedProfit float64 `json:"unRealizedProfit"`
	LiquidationPrice float64 `json:"liquidationPrice"`
	Leverage         float64 `json:"leverage"`
	MaxNotionalValue float64 `json:"maxNotionalValue"`
	MarginType       string  `json:"marginType"`
	IsolatedMargin   float64 `json:"isolatedMargin"`
	IsAutoAddMargin  bool    `json:"isAutoAddMargin"`
	PositionSide     string  `json:"positionSide"`
	Notional         float64 `json:"notional"`
	IsolatedWallet   float64 `json:"isolatedWallet"`
	UpdateTime       int64   `json:"updateTime"`
}

// AccountBalance represents account balance information
type AccountBalance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Locked float64 `json:"locked"`
}

// AccountInfo represents account information
type AccountInfo struct {
	MakerCommission  int64            `json:"makerCommission"`
	TakerCommission  int64            `json:"takerCommission"`
	BuyerCommission  int64            `json:"buyerCommission"`
	SellerCommission int64            `json:"sellerCommission"`
	CanTrade         bool             `json:"canTrade"`
	CanWithdraw      bool             `json:"canWithdraw"`
	CanDeposit       bool             `json:"canDeposit"`
	UpdateTime       int64            `json:"updateTime"`
	AccountType      string           `json:"accountType"`
	Balances         []AccountBalance `json:"balances"`
	Permissions      []string         `json:"permissions"`
}

// CommissionRates represents trading commission rates
type CommissionRates struct {
	Symbol             string `json:"symbol"`
	StandardCommission struct {
		Maker  float64 `json:"maker"`
		Taker  float64 `json:"taker"`
		Buyer  float64 `json:"buyer"`
		Seller float64 `json:"seller"`
	} `json:"standardCommission"`
	TaxCommission struct {
		Maker  float64 `json:"maker"`
		Taker  float64 `json:"taker"`
		Buyer  float64 `json:"buyer"`
		Seller float64 `json:"seller"`
	} `json:"taxCommission"`
	Discount struct {
		EnabledForAccount bool    `json:"enabledForAccount"`
		EnabledForSymbol  bool    `json:"enabledForSymbol"`
		DiscountAsset     string  `json:"discountAsset"`
		Discount          float64 `json:"discount"`
	} `json:"discount"`
}

// MarketSentiment represents market sentiment data
type MarketSentiment struct {
	Symbol    string    `json:"symbol"`
	Sentiment string    `json:"sentiment"` // BULLISH, BEARISH, NEUTRAL
	Score     float64   `json:"score"`     // -1 to 1
	Volume24h float64   `json:"volume24h"`
	Sources   []string  `json:"sources"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MarketHeatmapData represents market heatmap visualization data
type MarketHeatmapData struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Change24h     float64 `json:"change24h"`
	ChangePercent float64 `json:"changePercent"`
	Volume24h     float64 `json:"volume24h"`
	MarketCap     float64 `json:"marketCap"`
	Size          float64 `json:"size"` // For visualization sizing
}

// EnhancedVolumeProfile represents volume profile data
type EnhancedVolumeProfile struct {
	Symbol    string                       `json:"symbol"`
	Timeframe string                       `json:"timeframe"`
	StartTime int64                        `json:"startTime"`
	EndTime   int64                        `json:"endTime"`
	Levels    []EnhancedVolumeProfileLevel `json:"levels"`
	POC       float64                      `json:"poc"` // Point of Control
	ValueArea ValueArea                    `json:"valueArea"`
}

// EnhancedVolumeProfileLevel represents a single level in volume profile
type EnhancedVolumeProfileLevel struct {
	Price      float64 `json:"price"`
	Volume     float64 `json:"volume"`
	BuyVolume  float64 `json:"buyVolume"`
	SellVolume float64 `json:"sellVolume"`
}

// ValueArea represents the value area in volume profile
type ValueArea struct {
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume float64 `json:"volume"`
}

// EnhancedOrderFlowData represents order flow analysis data
type EnhancedOrderFlowData struct {
	Symbol     string                 `json:"symbol"`
	Timeframe  string                 `json:"timeframe"`
	StartTime  int64                  `json:"startTime"`
	EndTime    int64                  `json:"endTime"`
	Bars       []EnhancedOrderFlowBar `json:"bars"`
	Delta      []DeltaPoint           `json:"delta"`
	Imbalances []Imbalance            `json:"imbalances"`
}

// EnhancedOrderFlowBar represents a single bar in order flow
type EnhancedOrderFlowBar struct {
	Time       int64                    `json:"time"`
	Open       float64                  `json:"open"`
	High       float64                  `json:"high"`
	Low        float64                  `json:"low"`
	Close      float64                  `json:"close"`
	Volume     float64                  `json:"volume"`
	BuyVolume  float64                  `json:"buyVolume"`
	SellVolume float64                  `json:"sellVolume"`
	Footprint  map[string]FootprintData `json:"footprint"` // price -> footprint data
}

// FootprintData represents footprint chart data for a price level
type FootprintData struct {
	Price      float64 `json:"price"`
	BuyVolume  float64 `json:"buyVolume"`
	SellVolume float64 `json:"sellVolume"`
	Delta      float64 `json:"delta"`
	Trades     int     `json:"trades"`
}

// DeltaPoint represents a delta analysis point
type DeltaPoint struct {
	Time     int64   `json:"time"`
	Delta    float64 `json:"delta"`
	CumDelta float64 `json:"cumDelta"`
}

// Imbalance represents an order flow imbalance
type Imbalance struct {
	Price    float64 `json:"price"`
	Time     int64   `json:"time"`
	Type     string  `json:"type"` // BID_IMBALANCE, ASK_IMBALANCE
	Ratio    float64 `json:"ratio"`
	Strength string  `json:"strength"` // WEAK, MEDIUM, STRONG
}

// WebSocket message types for real-time updates
type EnhancedMarketDataUpdate struct {
	Symbol           string  `json:"symbol"`
	Price            float64 `json:"price"`
	Change24h        float64 `json:"change24h"`
	ChangePercent24h float64 `json:"changePercent24h"`
	Volume24h        float64 `json:"volume24h"`
	High24h          float64 `json:"high24h"`
	Low24h           float64 `json:"low24h"`
	Timestamp        int64   `json:"timestamp"`
}

type OrderBookUpdate struct {
	Symbol       string       `json:"symbol"`
	Bids         []PriceLevel `json:"bids"`
	Asks         []PriceLevel `json:"asks"`
	LastUpdateID int64        `json:"lastUpdateId"`
	Timestamp    int64        `json:"timestamp"`
}

type TradeUpdate struct {
	ID           string  `json:"id"`
	Symbol       string  `json:"symbol"`
	Price        float64 `json:"price"`
	Quantity     float64 `json:"quantity"`
	Time         int64   `json:"time"`
	IsBuyerMaker bool    `json:"isBuyerMaker"`
	TradeID      int64   `json:"tradeId"`
}

type TickerUpdate struct {
	Symbol             string  `json:"symbol"`
	PriceChange        float64 `json:"priceChange"`
	PriceChangePercent float64 `json:"priceChangePercent"`
	WeightedAvgPrice   float64 `json:"weightedAvgPrice"`
	PrevClosePrice     float64 `json:"prevClosePrice"`
	LastPrice          float64 `json:"lastPrice"`
	LastQty            float64 `json:"lastQty"`
	BidPrice           float64 `json:"bidPrice"`
	BidQty             float64 `json:"bidQty"`
	AskPrice           float64 `json:"askPrice"`
	AskQty             float64 `json:"askQty"`
	OpenPrice          float64 `json:"openPrice"`
	HighPrice          float64 `json:"highPrice"`
	LowPrice           float64 `json:"lowPrice"`
	Volume             float64 `json:"volume"`
	QuoteVolume        float64 `json:"quoteVolume"`
	OpenTime           int64   `json:"openTime"`
	CloseTime          int64   `json:"closeTime"`
	Count              int64   `json:"count"`
	Timestamp          int64   `json:"timestamp"`
}

// Real-time subscription management
type SubscriptionRequest struct {
	Type    string   `json:"type"`
	Channel string   `json:"channel"`
	Symbol  string   `json:"symbol,omitempty"`
	Symbols []string `json:"symbols,omitempty"`
}

type SubscriptionResponse struct {
	Type      string `json:"type"`
	Channel   string `json:"channel"`
	Symbol    string `json:"symbol,omitempty"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Timestamp int64  `json:"timestamp"`
}
