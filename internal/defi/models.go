package defi

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// DeFi Protocol Types
type ProtocolType string

const (
	ProtocolTypeUniswap     ProtocolType = "uniswap"
	ProtocolTypePancakeSwap ProtocolType = "pancakeswap"
	ProtocolTypeQuickSwap   ProtocolType = "quickswap"
	ProtocolTypeAave        ProtocolType = "aave"
	ProtocolTypeCompound    ProtocolType = "compound"
	ProtocolTypeChainlink   ProtocolType = "chainlink"
	ProtocolType1inch       ProtocolType = "1inch"
	ProtocolTypeSynthetix   ProtocolType = "synthetix"
)

// Chain represents supported blockchain networks
type Chain string

const (
	ChainEthereum Chain = "ethereum"
	ChainBSC      Chain = "bsc"
	ChainPolygon  Chain = "polygon"
	ChainArbitrum Chain = "arbitrum"
	ChainOptimism Chain = "optimism"
	ChainSolana   Chain = "solana"
)

// Token represents an ERC-20 token
type Token struct {
	Address  string          `json:"address"`
	Symbol   string          `json:"symbol"`
	Name     string          `json:"name"`
	Decimals int             `json:"decimals"`
	Chain    Chain           `json:"chain"`
	Price    decimal.Decimal `json:"price"`
	LogoURL  string          `json:"logo_url"`
}

// LiquidityPool represents a liquidity pool
type LiquidityPool struct {
	ID          string          `json:"id"`
	Protocol    ProtocolType    `json:"protocol"`
	Chain       Chain           `json:"chain"`
	Token0      Token           `json:"token0"`
	Token1      Token           `json:"token1"`
	Reserve0    decimal.Decimal `json:"reserve0"`
	Reserve1    decimal.Decimal `json:"reserve1"`
	TotalSupply decimal.Decimal `json:"total_supply"`
	Fee         decimal.Decimal `json:"fee"`
	APY         decimal.Decimal `json:"apy"`
	TVL         decimal.Decimal `json:"tvl"`
	Address     string          `json:"address"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// SwapQuote represents a token swap quote
type SwapQuote struct {
	ID           string          `json:"id"`
	Protocol     ProtocolType    `json:"protocol"`
	Chain        Chain           `json:"chain"`
	TokenIn      Token           `json:"token_in"`
	TokenOut     Token           `json:"token_out"`
	AmountIn     decimal.Decimal `json:"amount_in"`
	AmountOut    decimal.Decimal `json:"amount_out"`
	MinAmountOut decimal.Decimal `json:"min_amount_out"`
	PriceImpact  decimal.Decimal `json:"price_impact"`
	Fee          decimal.Decimal `json:"fee"`
	GasEstimate  uint64          `json:"gas_estimate"`
	Route        []string        `json:"route"`
	ExpiresAt    time.Time       `json:"expires_at"`
	CreatedAt    time.Time       `json:"created_at"`
}

// YieldFarm represents a yield farming opportunity
type YieldFarm struct {
	ID           string          `json:"id"`
	Protocol     ProtocolType    `json:"protocol"`
	Chain        Chain           `json:"chain"`
	Name         string          `json:"name"`
	Pool         LiquidityPool   `json:"pool"`
	RewardTokens []Token         `json:"reward_tokens"`
	APY          decimal.Decimal `json:"apy"`
	TVL          decimal.Decimal `json:"tvl"`
	MinDeposit   decimal.Decimal `json:"min_deposit"`
	LockPeriod   time.Duration   `json:"lock_period"`
	Active       bool            `json:"active"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// LendingPosition represents a lending/borrowing position
type LendingPosition struct {
	ID               string          `json:"id"`
	UserID           string          `json:"user_id"`
	Protocol         ProtocolType    `json:"protocol"`
	Chain            Chain           `json:"chain"`
	Token            Token           `json:"token"`
	Amount           decimal.Decimal `json:"amount"`
	Type             string          `json:"type"` // "lending" or "borrowing"
	APY              decimal.Decimal `json:"apy"`
	CollateralRatio  decimal.Decimal `json:"collateral_ratio,omitempty"`
	LiquidationPrice decimal.Decimal `json:"liquidation_price,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// StakingPosition represents a staking position
type StakingPosition struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	Protocol     ProtocolType    `json:"protocol"`
	Chain        Chain           `json:"chain"`
	Token        Token           `json:"token"`
	Amount       decimal.Decimal `json:"amount"`
	RewardTokens []Token         `json:"reward_tokens"`
	APY          decimal.Decimal `json:"apy"`
	LockPeriod   time.Duration   `json:"lock_period"`
	UnlockAt     *time.Time      `json:"unlock_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// PriceOracle represents price feed data
type PriceOracle struct {
	Token     Token           `json:"token"`
	Price     decimal.Decimal `json:"price"`
	Source    string          `json:"source"`
	Timestamp time.Time       `json:"timestamp"`
}

// FlashLoan represents a flash loan opportunity
type FlashLoan struct {
	ID        string          `json:"id"`
	Protocol  ProtocolType    `json:"protocol"`
	Chain     Chain           `json:"chain"`
	Token     Token           `json:"token"`
	Amount    decimal.Decimal `json:"amount"`
	Fee       decimal.Decimal `json:"fee"`
	Available bool            `json:"available"`
	MaxAmount decimal.Decimal `json:"max_amount"`
}

// ArbitrageOpportunity represents an arbitrage opportunity
type ArbitrageOpportunity struct {
	ID           string          `json:"id"`
	Token        Token           `json:"token"`
	BuyExchange  string          `json:"buy_exchange"`
	SellExchange string          `json:"sell_exchange"`
	BuyPrice     decimal.Decimal `json:"buy_price"`
	SellPrice    decimal.Decimal `json:"sell_price"`
	ProfitMargin decimal.Decimal `json:"profit_margin"`
	Volume       decimal.Decimal `json:"volume"`
	GasCost      decimal.Decimal `json:"gas_cost"`
	NetProfit    decimal.Decimal `json:"net_profit"`
	ExpiresAt    time.Time       `json:"expires_at"`
	CreatedAt    time.Time       `json:"created_at"`
}

// CoffeeToken represents the COFFEE utility token
type CoffeeToken struct {
	Address           string          `json:"address"`
	Chain             Chain           `json:"chain"`
	TotalSupply       decimal.Decimal `json:"total_supply"`
	CirculatingSupply decimal.Decimal `json:"circulating_supply"`
	Price             decimal.Decimal `json:"price"`
	MarketCap         decimal.Decimal `json:"market_cap"`
	StakingAPY        decimal.Decimal `json:"staking_apy"`
	RewardsPool       decimal.Decimal `json:"rewards_pool"`
}

// CoffeeStaking represents coffee token staking
type CoffeeStaking struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	Amount       decimal.Decimal `json:"amount"`
	RewardRate   decimal.Decimal `json:"reward_rate"`
	StartTime    time.Time       `json:"start_time"`
	LastClaim    time.Time       `json:"last_claim"`
	TotalRewards decimal.Decimal `json:"total_rewards"`
	Active       bool            `json:"active"`
}

// Request/Response Models

// GetTokenPriceRequest represents a request to get token price
type GetTokenPriceRequest struct {
	TokenAddress string `json:"token_address" validate:"required"`
	Chain        Chain  `json:"chain" validate:"required"`
}

// GetTokenPriceResponse represents a response with token price
type GetTokenPriceResponse struct {
	Token Token           `json:"token"`
	Price decimal.Decimal `json:"price"`
}

// GetSwapQuoteRequest represents a request for swap quote
type GetSwapQuoteRequest struct {
	TokenIn     string          `json:"token_in" validate:"required"`
	TokenOut    string          `json:"token_out" validate:"required"`
	AmountIn    decimal.Decimal `json:"amount_in" validate:"required"`
	Chain       Chain           `json:"chain" validate:"required"`
	Slippage    decimal.Decimal `json:"slippage"`
	UserAddress string          `json:"user_address"`
}

// GetSwapQuoteResponse represents a response with swap quote
type GetSwapQuoteResponse struct {
	Quote SwapQuote `json:"quote"`
}

// ExecuteSwapRequest represents a request to execute a swap
type ExecuteSwapRequest struct {
	QuoteID    string `json:"quote_id" validate:"required"`
	UserID     string `json:"user_id" validate:"required"`
	WalletID   string `json:"wallet_id" validate:"required"`
	Passphrase string `json:"passphrase" validate:"required"`
}

// ExecuteSwapResponse represents a response after executing a swap
type ExecuteSwapResponse struct {
	TransactionHash string `json:"transaction_hash"`
	Status          string `json:"status"`
}

// GetLiquidityPoolsRequest represents a request to get liquidity pools
type GetLiquidityPoolsRequest struct {
	Chain    Chain           `json:"chain"`
	Protocol ProtocolType    `json:"protocol"`
	Token0   string          `json:"token0"`
	Token1   string          `json:"token1"`
	MinTVL   decimal.Decimal `json:"min_tvl"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
}

// GetLiquidityPoolsResponse represents a response with liquidity pools
type GetLiquidityPoolsResponse struct {
	Pools []LiquidityPool `json:"pools"`
	Total int             `json:"total"`
}

// AddLiquidityRequest represents a request to add liquidity
type AddLiquidityRequest struct {
	UserID     string          `json:"user_id" validate:"required"`
	WalletID   string          `json:"wallet_id" validate:"required"`
	PoolID     string          `json:"pool_id" validate:"required"`
	Amount0    decimal.Decimal `json:"amount0" validate:"required"`
	Amount1    decimal.Decimal `json:"amount1" validate:"required"`
	MinAmount0 decimal.Decimal `json:"min_amount0"`
	MinAmount1 decimal.Decimal `json:"min_amount1"`
	Passphrase string          `json:"passphrase" validate:"required"`
}

// AddLiquidityResponse represents a response after adding liquidity
type AddLiquidityResponse struct {
	TransactionHash string          `json:"transaction_hash"`
	LPTokens        decimal.Decimal `json:"lp_tokens"`
	Status          string          `json:"status"`
}

// Trading Strategy Models

// TradingStrategy represents a trading strategy
type TradingStrategy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        TradingStrategyType    `json:"type"`
	Status      TradingStrategyStatus  `json:"status"`
	Config      map[string]interface{} `json:"config"`
	Performance TradingPerformance     `json:"performance"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TradingStrategyType represents the type of trading strategy
type TradingStrategyType string

const (
	StrategyTypeArbitrage    TradingStrategyType = "arbitrage"
	StrategyTypeYieldFarming TradingStrategyType = "yield_farming"
	StrategyTypeDCA          TradingStrategyType = "dca"
	StrategyTypeGridTrading  TradingStrategyType = "grid_trading"
	StrategyTypeRebalancing  TradingStrategyType = "rebalancing"
	StrategyTypeMarketMaking TradingStrategyType = "market_making"
)

// TradingStrategyStatus represents the status of a trading strategy
type TradingStrategyStatus string

const (
	StrategyStatusActive   TradingStrategyStatus = "active"
	StrategyStatusPaused   TradingStrategyStatus = "paused"
	StrategyStatusStopped  TradingStrategyStatus = "stopped"
	StrategyStatusBacktest TradingStrategyStatus = "backtest"
)

// TradingPerformance represents trading strategy performance metrics
type TradingPerformance struct {
	TotalTrades    int             `json:"total_trades"`
	WinningTrades  int             `json:"winning_trades"`
	LosingTrades   int             `json:"losing_trades"`
	WinRate        decimal.Decimal `json:"win_rate"`
	TotalProfit    decimal.Decimal `json:"total_profit"`
	TotalLoss      decimal.Decimal `json:"total_loss"`
	NetProfit      decimal.Decimal `json:"net_profit"`
	ROI            decimal.Decimal `json:"roi"`
	Sharpe         decimal.Decimal `json:"sharpe"`
	MaxDrawdown    decimal.Decimal `json:"max_drawdown"`
	AvgTradeProfit decimal.Decimal `json:"avg_trade_profit"`
	LastUpdated    time.Time       `json:"last_updated"`
}

// ArbitrageDetection represents an enhanced arbitrage opportunity
type ArbitrageDetection struct {
	ID             string            `json:"id"`
	Token          Token             `json:"token"`
	SourceExchange Exchange          `json:"source_exchange"`
	TargetExchange Exchange          `json:"target_exchange"`
	SourcePrice    decimal.Decimal   `json:"source_price"`
	TargetPrice    decimal.Decimal   `json:"target_price"`
	ProfitMargin   decimal.Decimal   `json:"profit_margin"`
	Volume         decimal.Decimal   `json:"volume"`
	GasCost        decimal.Decimal   `json:"gas_cost"`
	NetProfit      decimal.Decimal   `json:"net_profit"`
	Confidence     decimal.Decimal   `json:"confidence"`
	Risk           RiskLevel         `json:"risk"`
	ExecutionTime  time.Duration     `json:"execution_time"`
	ExpiresAt      time.Time         `json:"expires_at"`
	Status         OpportunityStatus `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
}

// Exchange represents a DEX or CEX
type Exchange struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Type     ExchangeType    `json:"type"`
	Chain    Chain           `json:"chain"`
	Protocol ProtocolType    `json:"protocol"`
	Address  string          `json:"address"`
	Fee      decimal.Decimal `json:"fee"`
	Active   bool            `json:"active"`
}

// ExchangeType represents the type of exchange
type ExchangeType string

const (
	ExchangeTypeDEX ExchangeType = "dex"
	ExchangeTypeCEX ExchangeType = "cex"
)

// RiskLevel represents risk assessment
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

// OpportunityStatus represents the status of an opportunity
type OpportunityStatus string

const (
	OpportunityStatusDetected  OpportunityStatus = "detected"
	OpportunityStatusExecuting OpportunityStatus = "executing"
	OpportunityStatusExecuted  OpportunityStatus = "executed"
	OpportunityStatusExpired   OpportunityStatus = "expired"
	OpportunityStatusFailed    OpportunityStatus = "failed"
)

// YieldFarmingOpportunity represents an enhanced yield farming opportunity
type YieldFarmingOpportunity struct {
	ID              string          `json:"id"`
	Protocol        ProtocolType    `json:"protocol"`
	Chain           Chain           `json:"chain"`
	Pool            LiquidityPool   `json:"pool"`
	Strategy        string          `json:"strategy"`
	APY             decimal.Decimal `json:"apy"`
	APR             decimal.Decimal `json:"apr"`
	TVL             decimal.Decimal `json:"tvl"`
	MinDeposit      decimal.Decimal `json:"min_deposit"`
	MaxDeposit      decimal.Decimal `json:"max_deposit"`
	LockPeriod      time.Duration   `json:"lock_period"`
	RewardTokens    []Token         `json:"reward_tokens"`
	Risk            RiskLevel       `json:"risk"`
	ImpermanentLoss decimal.Decimal `json:"impermanent_loss"`
	Active          bool            `json:"active"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// OnChainMetrics represents on-chain analytics data
type OnChainMetrics struct {
	Token           Token           `json:"token"`
	Chain           Chain           `json:"chain"`
	Price           decimal.Decimal `json:"price"`
	Volume24h       decimal.Decimal `json:"volume_24h"`
	Liquidity       decimal.Decimal `json:"liquidity"`
	MarketCap       decimal.Decimal `json:"market_cap"`
	Holders         int64           `json:"holders"`
	Transactions24h int64           `json:"transactions_24h"`
	Volatility      decimal.Decimal `json:"volatility"`
	Timestamp       time.Time       `json:"timestamp"`
}

// PriceProvider defines the interface for getting token prices from exchanges
type PriceProvider interface {
	GetPrice(ctx context.Context, token Token) (decimal.Decimal, error)
	GetExchangeInfo() Exchange
	IsHealthy(ctx context.Context) bool
}

// Yield Strategy Models

// YieldStrategy represents a yield farming strategy
type YieldStrategy struct {
	ID            string                     `json:"id"`
	Name          string                     `json:"name"`
	Type          YieldStrategyType          `json:"type"`
	Opportunities []*YieldFarmingOpportunity `json:"opportunities"`
	TotalAPY      decimal.Decimal            `json:"total_apy"`
	Risk          RiskLevel                  `json:"risk"`
	MinInvestment decimal.Decimal            `json:"min_investment"`
	MaxInvestment decimal.Decimal            `json:"max_investment"`
	AutoCompound  bool                       `json:"auto_compound"`
	RebalanceFreq time.Duration              `json:"rebalance_freq"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
}

// YieldStrategyType represents different yield strategies
type YieldStrategyType string

const (
	YieldStrategyTypeConservative YieldStrategyType = "conservative"
	YieldStrategyTypeBalanced     YieldStrategyType = "balanced"
	YieldStrategyTypeAggressive   YieldStrategyType = "aggressive"
	YieldStrategyTypeCustom       YieldStrategyType = "custom"
)

// OptimalStrategyRequest represents a request for optimal strategy
type OptimalStrategyRequest struct {
	InvestmentAmount decimal.Decimal `json:"investment_amount"`
	RiskTolerance    RiskLevel       `json:"risk_tolerance"`
	PreferredTokens  []string        `json:"preferred_tokens"`
	MinAPY           decimal.Decimal `json:"min_apy"`
	MaxLockPeriod    time.Duration   `json:"max_lock_period"`
	AutoCompound     bool            `json:"auto_compound"`
	Diversification  bool            `json:"diversification"`
}

// On-Chain Analysis Models

// BlockchainEvent represents a significant blockchain event
type BlockchainEvent struct {
	ID          string                 `json:"id"`
	Type        BlockchainEventType    `json:"type"`
	Chain       Chain                  `json:"chain"`
	BlockNumber uint64                 `json:"block_number"`
	TxHash      string                 `json:"tx_hash"`
	Token       Token                  `json:"token"`
	Amount      decimal.Decimal        `json:"amount"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// BlockchainEventType represents different types of blockchain events
type BlockchainEventType string

const (
	EventTypeLargeTransfer   BlockchainEventType = "large_transfer"
	EventTypeLiquidityAdd    BlockchainEventType = "liquidity_add"
	EventTypeLiquidityRemove BlockchainEventType = "liquidity_remove"
	EventTypeSwap            BlockchainEventType = "swap"
	EventTypeStake           BlockchainEventType = "stake"
	EventTypeUnstake         BlockchainEventType = "unstake"
	EventTypeNewToken        BlockchainEventType = "new_token"
	EventTypePriceAlert      BlockchainEventType = "price_alert"
)

// WhaleWatch represents monitoring of whale addresses
type WhaleWatch struct {
	Address    string          `json:"address"`
	Label      string          `json:"label"`
	Chain      Chain           `json:"chain"`
	Balance    decimal.Decimal `json:"balance"`
	LastTx     time.Time       `json:"last_tx"`
	TxCount24h int             `json:"tx_count_24h"`
	Volume24h  decimal.Decimal `json:"volume_24h"`
	Active     bool            `json:"active"`
}

// LiquidityEvent represents a liquidity pool event
type LiquidityEvent struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Pool      string          `json:"pool"`
	Token0    Token           `json:"token0"`
	Token1    Token           `json:"token1"`
	Amount0   decimal.Decimal `json:"amount0"`
	Amount1   decimal.Decimal `json:"amount1"`
	User      string          `json:"user"`
	TxHash    string          `json:"tx_hash"`
	Timestamp time.Time       `json:"timestamp"`
}

// MarketSignal represents a trading signal derived from on-chain analysis
type MarketSignal struct {
	ID         string          `json:"id"`
	Type       SignalType      `json:"type"`
	Token      Token           `json:"token"`
	Strength   decimal.Decimal `json:"strength"`
	Confidence decimal.Decimal `json:"confidence"`
	Direction  SignalDirection `json:"direction"`
	Timeframe  time.Duration   `json:"timeframe"`
	Reason     string          `json:"reason"`
	CreatedAt  time.Time       `json:"created_at"`
	ExpiresAt  time.Time       `json:"expires_at"`
}

// SignalType represents different types of market signals
type SignalType string

const (
	SignalTypeWhaleMovement  SignalType = "whale_movement"
	SignalTypeLiquidityShift SignalType = "liquidity_shift"
	SignalTypeVolumeSpike    SignalType = "volume_spike"
	SignalTypePriceAnomaly   SignalType = "price_anomaly"
	SignalTypeNewListing     SignalType = "new_listing"
)

// SignalDirection represents the direction of a market signal
type SignalDirection string

const (
	SignalDirectionBullish SignalDirection = "bullish"
	SignalDirectionBearish SignalDirection = "bearish"
	SignalDirectionNeutral SignalDirection = "neutral"
)

// TokenAnalysis represents comprehensive token analysis
type TokenAnalysis struct {
	Token           Token             `json:"token"`
	Metrics         OnChainMetrics    `json:"metrics"`
	Signals         []*MarketSignal   `json:"signals"`
	WhaleActivity   []*WhaleWatch     `json:"whale_activity"`
	LiquidityEvents []*LiquidityEvent `json:"liquidity_events"`
	Score           decimal.Decimal   `json:"score"`
	Recommendation  string            `json:"recommendation"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// Trading Bot Models

// TradingBot represents an automated trading bot
type TradingBot struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Strategy    TradingStrategyType `json:"strategy"`
	Status      TradingBotStatus    `json:"status"`
	Config      TradingBotConfig    `json:"config"`
	Performance TradingPerformance  `json:"performance"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// TradingBotStatus represents the status of a trading bot
type TradingBotStatus string

const (
	BotStatusActive  TradingBotStatus = "active"
	BotStatusPaused  TradingBotStatus = "paused"
	BotStatusStopped TradingBotStatus = "stopped"
	BotStatusError   TradingBotStatus = "error"
)

// TradingBotConfig represents trading bot configuration
type TradingBotConfig struct {
	MaxPositionSize   decimal.Decimal `json:"max_position_size"`
	MinProfitMargin   decimal.Decimal `json:"min_profit_margin"`
	MaxSlippage       decimal.Decimal `json:"max_slippage"`
	RiskTolerance     RiskLevel       `json:"risk_tolerance"`
	AutoCompound      bool            `json:"auto_compound"`
	MaxDailyTrades    int             `json:"max_daily_trades"`
	StopLossPercent   decimal.Decimal `json:"stop_loss_percent"`
	TakeProfitPercent decimal.Decimal `json:"take_profit_percent"`
	ExecutionDelay    time.Duration   `json:"execution_delay"`
}

// TradingPosition represents an active trading position
type TradingPosition struct {
	ID            string          `json:"id"`
	BotID         string          `json:"bot_id"`
	Type          PositionType    `json:"type"`
	Token         Token           `json:"token"`
	Amount        decimal.Decimal `json:"amount"`
	EntryPrice    decimal.Decimal `json:"entry_price"`
	CurrentPrice  decimal.Decimal `json:"current_price"`
	UnrealizedPnL decimal.Decimal `json:"unrealized_pnl"`
	StopLoss      decimal.Decimal `json:"stop_loss"`
	TakeProfit    decimal.Decimal `json:"take_profit"`
	OpenedAt      time.Time       `json:"opened_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// PositionType represents the type of trading position
type PositionType string

const (
	PositionTypeLong  PositionType = "long"
	PositionTypeShort PositionType = "short"
)

// TradingOrder represents a trading order
type TradingOrder struct {
	ID         string          `json:"id"`
	BotID      string          `json:"bot_id"`
	Type       OrderType       `json:"type"`
	Token      Token           `json:"token"`
	Amount     decimal.Decimal `json:"amount"`
	Price      decimal.Decimal `json:"price"`
	Status     OrderStatus     `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	ExecutedAt *time.Time      `json:"executed_at,omitempty"`
}

// OrderType represents the type of trading order
type OrderType string

const (
	OrderTypeBuy   OrderType = "buy"
	OrderTypeSell  OrderType = "sell"
	OrderTypeSwap  OrderType = "swap"
	OrderTypeStake OrderType = "stake"
)

// OrderStatus represents the status of a trading order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusExecuting OrderStatus = "executing"
	OrderStatusExecuted  OrderStatus = "executed"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusCancelled OrderStatus = "cancelled"
)
