package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Price represents a cryptocurrency price point
type Price struct {
	Symbol    string          `json:"symbol" db:"symbol"`
	Price     decimal.Decimal `json:"price" db:"price"`
	Volume24h decimal.Decimal `json:"volume_24h" db:"volume_24h"`
	Change24h decimal.Decimal `json:"change_24h" db:"change_24h"`
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
	Source    string          `json:"source" db:"source"`
}

// OHLCV represents Open, High, Low, Close, Volume data
type OHLCV struct {
	Symbol    string          `json:"symbol" db:"symbol"`
	Timeframe string          `json:"timeframe" db:"timeframe"`
	Open      decimal.Decimal `json:"open" db:"open"`
	High      decimal.Decimal `json:"high" db:"high"`
	Low       decimal.Decimal `json:"low" db:"low"`
	Close     decimal.Decimal `json:"close" db:"close"`
	Volume    decimal.Decimal `json:"volume" db:"volume"`
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
}

// MarketData represents comprehensive market data for a cryptocurrency
type MarketData struct {
	Symbol           string          `json:"symbol"`
	Name             string          `json:"name"`
	CurrentPrice     decimal.Decimal `json:"current_price"`
	MarketCap        decimal.Decimal `json:"market_cap"`
	MarketCapRank    int             `json:"market_cap_rank"`
	Volume24h        decimal.Decimal `json:"volume_24h"`
	Change24h        decimal.Decimal `json:"change_24h"`
	Change7d         decimal.Decimal `json:"change_7d"`
	Change30d        decimal.Decimal `json:"change_30d"`
	High24h          decimal.Decimal `json:"high_24h"`
	Low24h           decimal.Decimal `json:"low_24h"`
	CirculatingSupply decimal.Decimal `json:"circulating_supply"`
	TotalSupply      decimal.Decimal `json:"total_supply"`
	MaxSupply        decimal.Decimal `json:"max_supply"`
	ATH              decimal.Decimal `json:"ath"`
	ATHDate          time.Time       `json:"ath_date"`
	ATL              decimal.Decimal `json:"atl"`
	ATLDate          time.Time       `json:"atl_date"`
	LastUpdated      time.Time       `json:"last_updated"`
}

// TechnicalIndicator represents a technical analysis indicator
type TechnicalIndicator struct {
	Symbol     string                 `json:"symbol"`
	Timeframe  string                 `json:"timeframe"`
	Indicator  string                 `json:"indicator"`
	Value      decimal.Decimal        `json:"value"`
	Values     map[string]interface{} `json:"values,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Signal     string                 `json:"signal,omitempty"` // BUY, SELL, HOLD
	Confidence decimal.Decimal        `json:"confidence,omitempty"`
}

// TradingSignal represents an AI-generated trading signal
type TradingSignal struct {
	ID          string          `json:"id" db:"id"`
	Symbol      string          `json:"symbol" db:"symbol"`
	Signal      string          `json:"signal" db:"signal"` // BUY, SELL, HOLD
	Confidence  decimal.Decimal `json:"confidence" db:"confidence"`
	Price       decimal.Decimal `json:"price" db:"price"`
	TargetPrice decimal.Decimal `json:"target_price" db:"target_price"`
	StopLoss    decimal.Decimal `json:"stop_loss" db:"stop_loss"`
	Timeframe   string          `json:"timeframe" db:"timeframe"`
	Reasoning   string          `json:"reasoning" db:"reasoning"`
	Source      string          `json:"source" db:"source"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	ExpiresAt   time.Time       `json:"expires_at" db:"expires_at"`
	Status      string          `json:"status" db:"status"` // ACTIVE, EXECUTED, EXPIRED, CANCELLED
}

// MarketOverview represents overall market statistics
type MarketOverview struct {
	TotalMarketCap       decimal.Decimal `json:"total_market_cap"`
	TotalVolume24h       decimal.Decimal `json:"total_volume_24h"`
	MarketCapChange24h   decimal.Decimal `json:"market_cap_change_24h"`
	ActiveCryptocurrencies int           `json:"active_cryptocurrencies"`
	Markets              int             `json:"markets"`
	BTCDominance         decimal.Decimal `json:"btc_dominance"`
	ETHDominance         decimal.Decimal `json:"eth_dominance"`
	FearGreedIndex       int             `json:"fear_greed_index"`
	LastUpdated          time.Time       `json:"last_updated"`
}

// TopGainer represents a top gaining cryptocurrency
type TopGainer struct {
	Symbol    string          `json:"symbol"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Change24h decimal.Decimal `json:"change_24h"`
	Volume24h decimal.Decimal `json:"volume_24h"`
}

// TopLoser represents a top losing cryptocurrency
type TopLoser struct {
	Symbol    string          `json:"symbol"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Change24h decimal.Decimal `json:"change_24h"`
	Volume24h decimal.Decimal `json:"volume_24h"`
}

// NewsItem represents a cryptocurrency news item
type NewsItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	Sentiment   string    `json:"sentiment"` // POSITIVE, NEGATIVE, NEUTRAL
	Relevance   decimal.Decimal `json:"relevance"`
	Symbols     []string  `json:"symbols"`
}

// SentimentData represents market sentiment data
type SentimentData struct {
	Symbol         string          `json:"symbol"`
	OverallScore   decimal.Decimal `json:"overall_score"`
	TwitterScore   decimal.Decimal `json:"twitter_score"`
	RedditScore    decimal.Decimal `json:"reddit_score"`
	NewsScore      decimal.Decimal `json:"news_score"`
	SentimentTrend string          `json:"sentiment_trend"` // BULLISH, BEARISH, NEUTRAL
	LastUpdated    time.Time       `json:"last_updated"`
}

// ArbitrageOpportunity represents a cross-exchange arbitrage opportunity
type ArbitrageOpportunity struct {
	Symbol         string          `json:"symbol"`
	BuyExchange    string          `json:"buy_exchange"`
	SellExchange   string          `json:"sell_exchange"`
	BuyPrice       decimal.Decimal `json:"buy_price"`
	SellPrice      decimal.Decimal `json:"sell_price"`
	ProfitPercent  decimal.Decimal `json:"profit_percent"`
	ProfitUSD      decimal.Decimal `json:"profit_usd"`
	Volume         decimal.Decimal `json:"volume"`
	EstimatedGas   decimal.Decimal `json:"estimated_gas"`
	NetProfit      decimal.Decimal `json:"net_profit"`
	Timestamp      time.Time       `json:"timestamp"`
	Status         string          `json:"status"` // ACTIVE, EXECUTED, EXPIRED
}

// LiquidityPool represents a DeFi liquidity pool
type LiquidityPool struct {
	ID           string          `json:"id"`
	Protocol     string          `json:"protocol"`
	Token0       string          `json:"token0"`
	Token1       string          `json:"token1"`
	TVL          decimal.Decimal `json:"tvl"`
	APY          decimal.Decimal `json:"apy"`
	Volume24h    decimal.Decimal `json:"volume_24h"`
	Fees24h      decimal.Decimal `json:"fees_24h"`
	ImpermanentLoss decimal.Decimal `json:"impermanent_loss"`
	LastUpdated  time.Time       `json:"last_updated"`
}

// YieldFarmingOpportunity represents a yield farming opportunity
type YieldFarmingOpportunity struct {
	Protocol      string          `json:"protocol"`
	Pool          string          `json:"pool"`
	Tokens        []string        `json:"tokens"`
	APY           decimal.Decimal `json:"apy"`
	TVL           decimal.Decimal `json:"tvl"`
	Risk          string          `json:"risk"` // LOW, MEDIUM, HIGH
	MinDeposit    decimal.Decimal `json:"min_deposit"`
	LockPeriod    time.Duration   `json:"lock_period"`
	RewardTokens  []string        `json:"reward_tokens"`
	LastUpdated   time.Time       `json:"last_updated"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Channel   string      `json:"channel"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// WebSocketSubscription represents a WebSocket subscription
type WebSocketSubscription struct {
	Channel string   `json:"channel"`
	Symbols []string `json:"symbols,omitempty"`
	UserID  string   `json:"user_id,omitempty"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	APIResponse
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// HealthCheck represents system health status
type HealthCheck struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version"`
	Uptime    time.Duration     `json:"uptime"`
}
