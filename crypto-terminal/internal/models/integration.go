package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// MarketDataUpdate represents a market data update message
type MarketDataUpdate struct {
	Symbol    string          `json:"symbol"`
	Price     decimal.Decimal `json:"price"`
	Volume    decimal.Decimal `json:"volume"`
	Change24h decimal.Decimal `json:"change_24h"`
	Source    string          `json:"source"`
	Timestamp time.Time       `json:"timestamp"`
}


// PortfolioUpdate represents a portfolio update message
type PortfolioUpdate struct {
	PortfolioID      string          `json:"portfolio_id"`
	UserID           string          `json:"user_id"`
	TotalValue       decimal.Decimal `json:"total_value"`
	TotalPnL         decimal.Decimal `json:"total_pnl"`
	TotalPnLPercent  decimal.Decimal `json:"total_pnl_percent"`
	DayChange        decimal.Decimal `json:"day_change"`
	DayChangePercent decimal.Decimal `json:"day_change_percent"`
	UpdatedAt        time.Time       `json:"updated_at"`
}


// DeFiProtocolData represents DeFi protocol integration data
type DeFiProtocolData struct {
	Protocol     string                 `json:"protocol"`
	Network      string                 `json:"network"`
	TVL          decimal.Decimal        `json:"tvl"`
	APY          decimal.Decimal        `json:"apy"`
	Pools        []DeFiPool             `json:"pools"`
	Tokens       []DeFiToken            `json:"tokens"`
	LastUpdated  time.Time              `json:"last_updated"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DeFiPool represents a DeFi liquidity pool
type DeFiPool struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Token0       string          `json:"token0"`
	Token1       string          `json:"token1"`
	Liquidity    decimal.Decimal `json:"liquidity"`
	Volume24h    decimal.Decimal `json:"volume_24h"`
	Fees24h      decimal.Decimal `json:"fees_24h"`
	APR          decimal.Decimal `json:"apr"`
	Price        decimal.Decimal `json:"price"`
	PriceChange  decimal.Decimal `json:"price_change"`
}

// DeFiToken represents a DeFi token
type DeFiToken struct {
	Address     string          `json:"address"`
	Symbol      string          `json:"symbol"`
	Name        string          `json:"name"`
	Decimals    int             `json:"decimals"`
	Price       decimal.Decimal `json:"price"`
	MarketCap   decimal.Decimal `json:"market_cap"`
	Volume24h   decimal.Decimal `json:"volume_24h"`
	Change24h   decimal.Decimal `json:"change_24h"`
	LastUpdated time.Time       `json:"last_updated"`
}

// AIAgentSignal represents a signal from AI trading agents
type AIAgentSignal struct {
	AgentID      string                 `json:"agent_id"`
	AgentName    string                 `json:"agent_name"`
	Symbol       string                 `json:"symbol"`
	Action       string                 `json:"action"` // BUY, SELL, HOLD
	Confidence   decimal.Decimal        `json:"confidence"`
	Reasoning    string                 `json:"reasoning"`
	TechnicalData map[string]interface{} `json:"technical_data,omitempty"`
	SentimentData map[string]interface{} `json:"sentiment_data,omitempty"`
	RiskScore    decimal.Decimal        `json:"risk_score"`
	TimeHorizon  string                 `json:"time_horizon"` // SHORT, MEDIUM, LONG
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
}

// WalletIntegration represents wallet service integration data
type WalletIntegration struct {
	WalletID     string                 `json:"wallet_id"`
	UserID       string                 `json:"user_id"`
	Network      string                 `json:"network"`
	Address      string                 `json:"address"`
	Balance      decimal.Decimal        `json:"balance"`
	Tokens       []WalletToken          `json:"tokens"`
	Transactions []WalletTransaction    `json:"transactions"`
	LastSync     time.Time              `json:"last_sync"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// WalletToken represents a token in a wallet
type WalletToken struct {
	Address     string          `json:"address"`
	Symbol      string          `json:"symbol"`
	Name        string          `json:"name"`
	Balance     decimal.Decimal `json:"balance"`
	Value       decimal.Decimal `json:"value"`
	Price       decimal.Decimal `json:"price"`
	Change24h   decimal.Decimal `json:"change_24h"`
	LastUpdated time.Time       `json:"last_updated"`
}

// WalletTransaction represents a wallet transaction
type WalletTransaction struct {
	Hash        string          `json:"hash"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Value       decimal.Decimal `json:"value"`
	Gas         decimal.Decimal `json:"gas"`
	GasPrice    decimal.Decimal `json:"gas_price"`
	Status      string          `json:"status"`
	BlockNumber int64           `json:"block_number"`
	Timestamp   time.Time       `json:"timestamp"`
}

// NewsIntegration represents news service integration data
type NewsIntegration struct {
	ArticleID   string                 `json:"article_id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Summary     string                 `json:"summary"`
	Source      string                 `json:"source"`
	Author      string                 `json:"author"`
	URL         string                 `json:"url"`
	Sentiment   decimal.Decimal        `json:"sentiment"`
	Symbols     []string               `json:"symbols"`
	Categories  []string               `json:"categories"`
	PublishedAt time.Time              `json:"published_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SocialSentiment represents social media sentiment data
type SocialSentiment struct {
	Symbol      string          `json:"symbol"`
	Platform    string          `json:"platform"` // twitter, reddit, telegram
	Sentiment   decimal.Decimal `json:"sentiment"` // -1 to 1
	Volume      int64           `json:"volume"`
	Mentions    int64           `json:"mentions"`
	Engagement  int64           `json:"engagement"`
	Trending    bool            `json:"trending"`
	Keywords    []string        `json:"keywords"`
	TimeFrame   string          `json:"time_frame"` // 1h, 4h, 24h
	LastUpdated time.Time       `json:"last_updated"`
}


// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	Service         string                 `json:"service"`
	CPUUsage        decimal.Decimal        `json:"cpu_usage"`
	MemoryUsage     decimal.Decimal        `json:"memory_usage"`
	DiskUsage       decimal.Decimal        `json:"disk_usage"`
	NetworkIn       decimal.Decimal        `json:"network_in"`
	NetworkOut      decimal.Decimal        `json:"network_out"`
	ActiveConnections int64                `json:"active_connections"`
	RequestsPerSecond decimal.Decimal      `json:"requests_per_second"`
	ErrorRate       decimal.Decimal        `json:"error_rate"`
	Latency         time.Duration          `json:"latency"`
	Uptime          time.Duration          `json:"uptime"`
	CustomMetrics   map[string]interface{} `json:"custom_metrics,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}
