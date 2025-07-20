package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Portfolio represents a user's cryptocurrency portfolio
type Portfolio struct {
	ID              string          `json:"id" db:"id"`
	UserID          string          `json:"user_id" db:"user_id"`
	Name            string          `json:"name" db:"name"`
	Description     string          `json:"description" db:"description"`
	IsPublic        bool            `json:"is_public" db:"is_public"`
	TotalValue      decimal.Decimal `json:"total_value" db:"total_value"`
	TotalCost       decimal.Decimal `json:"total_cost" db:"total_cost"`
	TotalPnL        decimal.Decimal `json:"total_pnl" db:"total_pnl"`
	TotalPnLPercent decimal.Decimal `json:"total_pnl_percent" db:"total_pnl_percent"`
	DayChange       decimal.Decimal `json:"day_change" db:"day_change"`
	DayChangePercent decimal.Decimal `json:"day_change_percent" db:"day_change_percent"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	Holdings        []Holding       `json:"holdings,omitempty"`
}

// Holding represents a cryptocurrency holding in a portfolio
type Holding struct {
	ID              string          `json:"id" db:"id"`
	PortfolioID     string          `json:"portfolio_id" db:"portfolio_id"`
	Symbol          string          `json:"symbol" db:"symbol"`
	Name            string          `json:"name" db:"name"`
	Quantity        decimal.Decimal `json:"quantity" db:"quantity"`
	AveragePrice    decimal.Decimal `json:"average_price" db:"average_price"`
	CurrentPrice    decimal.Decimal `json:"current_price" db:"current_price"`
	TotalCost       decimal.Decimal `json:"total_cost" db:"total_cost"`
	CurrentValue    decimal.Decimal `json:"current_value" db:"current_value"`
	PnL             decimal.Decimal `json:"pnl" db:"pnl"`
	PnLPercent      decimal.Decimal `json:"pnl_percent" db:"pnl_percent"`
	DayChange       decimal.Decimal `json:"day_change" db:"day_change"`
	DayChangePercent decimal.Decimal `json:"day_change_percent" db:"day_change_percent"`
	AllocationPercent decimal.Decimal `json:"allocation_percent" db:"allocation_percent"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	Transactions    []Transaction   `json:"transactions,omitempty"`
}

// Transaction represents a buy/sell transaction
type Transaction struct {
	ID          string          `json:"id" db:"id"`
	PortfolioID string          `json:"portfolio_id" db:"portfolio_id"`
	HoldingID   string          `json:"holding_id" db:"holding_id"`
	Symbol      string          `json:"symbol" db:"symbol"`
	Type        string          `json:"type" db:"type"` // BUY, SELL
	Quantity    decimal.Decimal `json:"quantity" db:"quantity"`
	Price       decimal.Decimal `json:"price" db:"price"`
	TotalAmount decimal.Decimal `json:"total_amount" db:"total_amount"`
	Fee         decimal.Decimal `json:"fee" db:"fee"`
	Exchange    string          `json:"exchange" db:"exchange"`
	TxHash      string          `json:"tx_hash" db:"tx_hash"`
	Notes       string          `json:"notes" db:"notes"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// PortfolioPerformance represents portfolio performance metrics
type PortfolioPerformance struct {
	PortfolioID     string                    `json:"portfolio_id"`
	TimeRange       string                    `json:"time_range"` // 1D, 7D, 30D, 90D, 1Y, ALL
	StartValue      decimal.Decimal           `json:"start_value"`
	EndValue        decimal.Decimal           `json:"end_value"`
	TotalReturn     decimal.Decimal           `json:"total_return"`
	TotalReturnPercent decimal.Decimal        `json:"total_return_percent"`
	AnnualizedReturn decimal.Decimal          `json:"annualized_return"`
	Volatility      decimal.Decimal           `json:"volatility"`
	SharpeRatio     decimal.Decimal           `json:"sharpe_ratio"`
	MaxDrawdown     decimal.Decimal           `json:"max_drawdown"`
	MaxDrawdownPercent decimal.Decimal        `json:"max_drawdown_percent"`
	WinRate         decimal.Decimal           `json:"win_rate"`
	ProfitFactor    decimal.Decimal           `json:"profit_factor"`
	BestDay         decimal.Decimal           `json:"best_day"`
	WorstDay        decimal.Decimal           `json:"worst_day"`
	CalculatedAt    time.Time                 `json:"calculated_at"`
	HistoricalData  []PortfolioHistoricalData `json:"historical_data,omitempty"`
}

// PortfolioHistoricalData represents historical portfolio value
type PortfolioHistoricalData struct {
	PortfolioID string          `json:"portfolio_id" db:"portfolio_id"`
	Date        time.Time       `json:"date" db:"date"`
	TotalValue  decimal.Decimal `json:"total_value" db:"total_value"`
	TotalCost   decimal.Decimal `json:"total_cost" db:"total_cost"`
	PnL         decimal.Decimal `json:"pnl" db:"pnl"`
	PnLPercent  decimal.Decimal `json:"pnl_percent" db:"pnl_percent"`
}

// AssetAllocation represents portfolio asset allocation
type AssetAllocation struct {
	Symbol      string          `json:"symbol"`
	Name        string          `json:"name"`
	Value       decimal.Decimal `json:"value"`
	Percentage  decimal.Decimal `json:"percentage"`
	Category    string          `json:"category"` // LARGE_CAP, MID_CAP, SMALL_CAP, DEFI, NFT, etc.
}

// RiskMetrics represents portfolio risk analysis
type RiskMetrics struct {
	PortfolioID         string          `json:"portfolio_id"`
	VaR95               decimal.Decimal `json:"var_95"` // Value at Risk 95%
	VaR99               decimal.Decimal `json:"var_99"` // Value at Risk 99%
	ConditionalVaR      decimal.Decimal `json:"conditional_var"`
	Beta                decimal.Decimal `json:"beta"`
	Alpha               decimal.Decimal `json:"alpha"`
	CorrelationBTC      decimal.Decimal `json:"correlation_btc"`
	CorrelationETH      decimal.Decimal `json:"correlation_eth"`
	ConcentrationRisk   decimal.Decimal `json:"concentration_risk"`
	LiquidityRisk       decimal.Decimal `json:"liquidity_risk"`
	RiskScore           decimal.Decimal `json:"risk_score"` // 1-10 scale
	RiskLevel           string          `json:"risk_level"` // LOW, MEDIUM, HIGH
	Recommendations     []string        `json:"recommendations"`
	CalculatedAt        time.Time       `json:"calculated_at"`
}

// DiversificationAnalysis represents portfolio diversification metrics
type DiversificationAnalysis struct {
	PortfolioID           string                    `json:"portfolio_id"`
	DiversificationScore  decimal.Decimal           `json:"diversification_score"` // 0-100
	NumberOfAssets        int                       `json:"number_of_assets"`
	EffectiveAssets       decimal.Decimal           `json:"effective_assets"`
	HerfindahlIndex       decimal.Decimal           `json:"herfindahl_index"`
	SectorAllocation      []SectorAllocation        `json:"sector_allocation"`
	MarketCapAllocation   []MarketCapAllocation     `json:"market_cap_allocation"`
	GeographicAllocation  []GeographicAllocation    `json:"geographic_allocation"`
	Recommendations       []string                  `json:"recommendations"`
	CalculatedAt          time.Time                 `json:"calculated_at"`
}

// SectorAllocation represents allocation by sector
type SectorAllocation struct {
	Sector     string          `json:"sector"`
	Value      decimal.Decimal `json:"value"`
	Percentage decimal.Decimal `json:"percentage"`
}

// MarketCapAllocation represents allocation by market cap
type MarketCapAllocation struct {
	Category   string          `json:"category"` // LARGE_CAP, MID_CAP, SMALL_CAP
	Value      decimal.Decimal `json:"value"`
	Percentage decimal.Decimal `json:"percentage"`
}

// GeographicAllocation represents allocation by geography
type GeographicAllocation struct {
	Region     string          `json:"region"`
	Value      decimal.Decimal `json:"value"`
	Percentage decimal.Decimal `json:"percentage"`
}

// PortfolioAlert represents a portfolio-related alert
type PortfolioAlert struct {
	ID          string          `json:"id" db:"id"`
	PortfolioID string          `json:"portfolio_id" db:"portfolio_id"`
	Type        string          `json:"type" db:"type"` // PRICE, PNL, ALLOCATION, RISK
	Symbol      string          `json:"symbol" db:"symbol"`
	Condition   string          `json:"condition" db:"condition"` // ABOVE, BELOW, EQUALS
	Threshold   decimal.Decimal `json:"threshold" db:"threshold"`
	CurrentValue decimal.Decimal `json:"current_value" db:"current_value"`
	Message     string          `json:"message" db:"message"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	IsTriggered bool            `json:"is_triggered" db:"is_triggered"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	TriggeredAt *time.Time      `json:"triggered_at" db:"triggered_at"`
}

// PortfolioSummary represents a summary view of portfolio
type PortfolioSummary struct {
	Portfolio     Portfolio                `json:"portfolio"`
	Performance   PortfolioPerformance     `json:"performance"`
	RiskMetrics   RiskMetrics              `json:"risk_metrics"`
	TopHoldings   []Holding                `json:"top_holdings"`
	Allocation    []AssetAllocation        `json:"allocation"`
	RecentTrades  []Transaction            `json:"recent_trades"`
	ActiveAlerts  []PortfolioAlert         `json:"active_alerts"`
}

// WalletConnection represents a connected wallet
type WalletConnection struct {
	ID          string    `json:"id" db:"id"`
	PortfolioID string    `json:"portfolio_id" db:"portfolio_id"`
	WalletType  string    `json:"wallet_type" db:"wallet_type"` // METAMASK, COINBASE, HARDWARE
	Address     string    `json:"address" db:"address"`
	Network     string    `json:"network" db:"network"` // ETHEREUM, BSC, POLYGON, etc.
	IsActive    bool      `json:"is_active" db:"is_active"`
	LastSynced  time.Time `json:"last_synced" db:"last_synced"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PortfolioSync represents a portfolio synchronization operation
type PortfolioSync struct {
	ID          string    `json:"id" db:"id"`
	PortfolioID string    `json:"portfolio_id" db:"portfolio_id"`
	Status      string    `json:"status" db:"status"` // PENDING, RUNNING, COMPLETED, FAILED
	Progress    int       `json:"progress" db:"progress"` // 0-100
	Message     string    `json:"message" db:"message"`
	StartedAt   time.Time `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	Error       string    `json:"error" db:"error"`
}

// CreatePortfolioRequest represents a request to create a new portfolio
type CreatePortfolioRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"max=500"`
	IsPublic    bool   `json:"is_public"`
}

// UpdatePortfolioRequest represents a request to update an existing portfolio
type UpdatePortfolioRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsPublic    *bool   `json:"is_public,omitempty"`
}
