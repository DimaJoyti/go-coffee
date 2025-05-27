package defi

import (
	"math/big"
	"time"

	"github.com/shopspring/decimal"
)

// DeFi Protocol Types
type ProtocolType string

const (
	ProtocolTypeUniswap   ProtocolType = "uniswap"
	ProtocolTypePancakeSwap ProtocolType = "pancakeswap"
	ProtocolTypeQuickSwap ProtocolType = "quickswap"
	ProtocolTypeAave      ProtocolType = "aave"
	ProtocolTypeCompound  ProtocolType = "compound"
	ProtocolTypeChainlink ProtocolType = "chainlink"
	ProtocolType1inch     ProtocolType = "1inch"
	ProtocolTypeSynthetix ProtocolType = "synthetix"
)

// Chain represents supported blockchain networks
type Chain string

const (
	ChainEthereum Chain = "ethereum"
	ChainBSC      Chain = "bsc"
	ChainPolygon  Chain = "polygon"
	ChainArbitrum Chain = "arbitrum"
	ChainOptimism Chain = "optimism"
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
	ID           string          `json:"id"`
	Protocol     ProtocolType    `json:"protocol"`
	Chain        Chain           `json:"chain"`
	Token0       Token           `json:"token0"`
	Token1       Token           `json:"token1"`
	Reserve0     decimal.Decimal `json:"reserve0"`
	Reserve1     decimal.Decimal `json:"reserve1"`
	TotalSupply  decimal.Decimal `json:"total_supply"`
	Fee          decimal.Decimal `json:"fee"`
	APY          decimal.Decimal `json:"apy"`
	TVL          decimal.Decimal `json:"tvl"`
	Address      string          `json:"address"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// SwapQuote represents a token swap quote
type SwapQuote struct {
	ID               string          `json:"id"`
	Protocol         ProtocolType    `json:"protocol"`
	Chain            Chain           `json:"chain"`
	TokenIn          Token           `json:"token_in"`
	TokenOut         Token           `json:"token_out"`
	AmountIn         decimal.Decimal `json:"amount_in"`
	AmountOut        decimal.Decimal `json:"amount_out"`
	MinAmountOut     decimal.Decimal `json:"min_amount_out"`
	PriceImpact      decimal.Decimal `json:"price_impact"`
	Fee              decimal.Decimal `json:"fee"`
	GasEstimate      uint64          `json:"gas_estimate"`
	Route            []string        `json:"route"`
	ExpiresAt        time.Time       `json:"expires_at"`
	CreatedAt        time.Time       `json:"created_at"`
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
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	Protocol        ProtocolType    `json:"protocol"`
	Chain           Chain           `json:"chain"`
	Token           Token           `json:"token"`
	Amount          decimal.Decimal `json:"amount"`
	Type            string          `json:"type"` // "lending" or "borrowing"
	APY             decimal.Decimal `json:"apy"`
	CollateralRatio decimal.Decimal `json:"collateral_ratio,omitempty"`
	LiquidationPrice decimal.Decimal `json:"liquidation_price,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
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
	ID          string          `json:"id"`
	Protocol    ProtocolType    `json:"protocol"`
	Chain       Chain           `json:"chain"`
	Token       Token           `json:"token"`
	Amount      decimal.Decimal `json:"amount"`
	Fee         decimal.Decimal `json:"fee"`
	Available   bool            `json:"available"`
	MaxAmount   decimal.Decimal `json:"max_amount"`
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
	Address      string          `json:"address"`
	Chain        Chain           `json:"chain"`
	TotalSupply  decimal.Decimal `json:"total_supply"`
	CirculatingSupply decimal.Decimal `json:"circulating_supply"`
	Price        decimal.Decimal `json:"price"`
	MarketCap    decimal.Decimal `json:"market_cap"`
	StakingAPY   decimal.Decimal `json:"staking_apy"`
	RewardsPool  decimal.Decimal `json:"rewards_pool"`
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
	Chain    Chain        `json:"chain"`
	Protocol ProtocolType `json:"protocol"`
	Token0   string       `json:"token0"`
	Token1   string       `json:"token1"`
	MinTVL   decimal.Decimal `json:"min_tvl"`
	Limit    int          `json:"limit"`
	Offset   int          `json:"offset"`
}

// GetLiquidityPoolsResponse represents a response with liquidity pools
type GetLiquidityPoolsResponse struct {
	Pools []LiquidityPool `json:"pools"`
	Total int             `json:"total"`
}

// AddLiquidityRequest represents a request to add liquidity
type AddLiquidityRequest struct {
	UserID      string          `json:"user_id" validate:"required"`
	WalletID    string          `json:"wallet_id" validate:"required"`
	PoolID      string          `json:"pool_id" validate:"required"`
	Amount0     decimal.Decimal `json:"amount0" validate:"required"`
	Amount1     decimal.Decimal `json:"amount1" validate:"required"`
	MinAmount0  decimal.Decimal `json:"min_amount0"`
	MinAmount1  decimal.Decimal `json:"min_amount1"`
	Passphrase  string          `json:"passphrase" validate:"required"`
}

// AddLiquidityResponse represents a response after adding liquidity
type AddLiquidityResponse struct {
	TransactionHash string          `json:"transaction_hash"`
	LPTokens        decimal.Decimal `json:"lp_tokens"`
	Status          string          `json:"status"`
}
