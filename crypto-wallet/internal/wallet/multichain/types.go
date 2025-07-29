package multichain

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// CrossChainTransferRequest represents a cross-chain transfer request
type CrossChainTransferRequest struct {
	SourceChain   string          `json:"source_chain"`
	DestChain     string          `json:"dest_chain"`
	Token         TokenConfig     `json:"token"`
	Amount        decimal.Decimal `json:"amount"`
	FromAddress   common.Address  `json:"from_address"`
	ToAddress     common.Address  `json:"to_address"`
	Slippage      decimal.Decimal `json:"slippage"`
	Deadline      time.Time       `json:"deadline"`
	PrivateKey    string          `json:"private_key"`
	MaxFee        decimal.Decimal `json:"max_fee"`
	Priority      string          `json:"priority"` // "fast", "normal", "slow"
	Metadata      map[string]interface{} `json:"metadata"`
}

// TransactionRequest represents a transaction request
type TransactionRequest struct {
	To         common.Address  `json:"to"`
	Value      decimal.Decimal `json:"value"`
	Data       []byte          `json:"data"`
	GasLimit   uint64          `json:"gas_limit"`
	GasPrice   *big.Int        `json:"gas_price"`
	Nonce      uint64          `json:"nonce"`
	PrivateKey string          `json:"private_key"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TransactionResult represents a transaction result
type TransactionResult struct {
	Hash        string          `json:"hash"`
	Status      string          `json:"status"`
	BlockNumber uint64          `json:"block_number"`
	GasUsed     uint64          `json:"gas_used"`
	GasPrice    *big.Int        `json:"gas_price"`
	Fee         decimal.Decimal `json:"fee"`
	CreatedAt   time.Time       `json:"created_at"`
	ConfirmedAt *time.Time      `json:"confirmed_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Transaction represents a blockchain transaction
type Transaction struct {
	To       common.Address  `json:"to"`
	Value    decimal.Decimal `json:"value"`
	Data     []byte          `json:"data"`
	GasLimit uint64          `json:"gas_limit"`
	GasPrice *big.Int        `json:"gas_price"`
	Nonce    uint64          `json:"nonce"`
}

// BridgeConfig holds bridge configuration
type BridgeConfig struct {
	Enabled         bool                       `json:"enabled" yaml:"enabled"`
	SupportedBridges []string                  `json:"supported_bridges" yaml:"supported_bridges"`
	DefaultBridge   string                     `json:"default_bridge" yaml:"default_bridge"`
	BridgeConfigs   map[string]BridgeProtocolConfig `json:"bridge_configs" yaml:"bridge_configs"`
	MaxSlippage     decimal.Decimal            `json:"max_slippage" yaml:"max_slippage"`
	Timeout         time.Duration              `json:"timeout" yaml:"timeout"`
}

// BridgeProtocolConfig holds configuration for a specific bridge protocol
type BridgeProtocolConfig struct {
	Name            string            `json:"name"`
	Enabled         bool              `json:"enabled"`
	APIEndpoint     string            `json:"api_endpoint"`
	ContractAddress map[string]string `json:"contract_address"` // chain -> address
	SupportedChains []string          `json:"supported_chains"`
	SupportedTokens []string          `json:"supported_tokens"`
	MinAmount       decimal.Decimal   `json:"min_amount"`
	MaxAmount       decimal.Decimal   `json:"max_amount"`
	Fee             decimal.Decimal   `json:"fee"`
	EstimatedTime   time.Duration     `json:"estimated_time"`
	Priority        int               `json:"priority"`
}

// GasConfig holds gas tracking configuration
type GasConfig struct {
	Enabled         bool                    `json:"enabled" yaml:"enabled"`
	UpdateInterval  time.Duration           `json:"update_interval" yaml:"update_interval"`
	ChainConfigs    map[string]GasChainConfig `json:"chain_configs" yaml:"chain_configs"`
	AlertThresholds map[string]decimal.Decimal `json:"alert_thresholds" yaml:"alert_thresholds"`
}

// GasChainConfig holds gas configuration for a specific chain
type GasChainConfig struct {
	Enabled       bool            `json:"enabled"`
	GasStation    string          `json:"gas_station"`
	DefaultGasPrice decimal.Decimal `json:"default_gas_price"`
	MaxGasPrice   decimal.Decimal `json:"max_gas_price"`
	GasMultiplier decimal.Decimal `json:"gas_multiplier"`
	Priority      map[string]decimal.Decimal `json:"priority"` // "slow", "normal", "fast"
}

// PriceOracleConfig holds price oracle configuration
type PriceOracleConfig struct {
	Enabled         bool                        `json:"enabled" yaml:"enabled"`
	Provider        string                      `json:"provider" yaml:"provider"`
	APIKey          string                      `json:"api_key" yaml:"api_key"`
	UpdateInterval  time.Duration               `json:"update_interval" yaml:"update_interval"`
	CacheTimeout    time.Duration               `json:"cache_timeout" yaml:"cache_timeout"`
	SupportedTokens []string                    `json:"supported_tokens" yaml:"supported_tokens"`
	Endpoints       map[string]string           `json:"endpoints" yaml:"endpoints"`
	RateLimit       int                         `json:"rate_limit" yaml:"rate_limit"`
}

// PortfolioConfig holds portfolio tracking configuration
type PortfolioConfig struct {
	Enabled            bool          `json:"enabled" yaml:"enabled"`
	UpdateInterval     time.Duration `json:"update_interval" yaml:"update_interval"`
	HistoryRetention   time.Duration `json:"history_retention" yaml:"history_retention"`
	RiskCalculation    bool          `json:"risk_calculation" yaml:"risk_calculation"`
	PerformanceTracking bool         `json:"performance_tracking" yaml:"performance_tracking"`
	AlertsEnabled      bool          `json:"alerts_enabled" yaml:"alerts_enabled"`
}

// GasPrice represents gas price information
type GasPrice struct {
	Chain       string          `json:"chain"`
	Slow        decimal.Decimal `json:"slow"`
	Normal      decimal.Decimal `json:"normal"`
	Fast        decimal.Decimal `json:"fast"`
	Instant     decimal.Decimal `json:"instant"`
	LastUpdated time.Time       `json:"last_updated"`
	Source      string          `json:"source"`
}

// TokenPrice represents token price information
type TokenPrice struct {
	Symbol      string          `json:"symbol"`
	PriceUSD    decimal.Decimal `json:"price_usd"`
	Change24h   decimal.Decimal `json:"change_24h"`
	Volume24h   decimal.Decimal `json:"volume_24h"`
	MarketCap   decimal.Decimal `json:"market_cap"`
	LastUpdated time.Time       `json:"last_updated"`
	Source      string          `json:"source"`
}

// BridgeRoute represents a bridge route between chains
type BridgeRoute struct {
	ID              string                 `json:"id"`
	Protocol        string                 `json:"protocol"`
	SourceChain     string                 `json:"source_chain"`
	DestChain       string                 `json:"dest_chain"`
	Token           TokenConfig            `json:"token"`
	MinAmount       decimal.Decimal        `json:"min_amount"`
	MaxAmount       decimal.Decimal        `json:"max_amount"`
	Fee             decimal.Decimal        `json:"fee"`
	EstimatedTime   time.Duration          `json:"estimated_time"`
	Liquidity       decimal.Decimal        `json:"liquidity"`
	Success         decimal.Decimal        `json:"success_rate"`
	Enabled         bool                   `json:"enabled"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PortfolioSnapshot represents a portfolio snapshot at a point in time
type PortfolioSnapshot struct {
	Address       common.Address             `json:"address"`
	Timestamp     time.Time                  `json:"timestamp"`
	TotalValueUSD decimal.Decimal            `json:"total_value_usd"`
	Balances      map[string]*UnifiedBalance `json:"balances"`
	Performance   *PerformanceMetrics        `json:"performance"`
	Risk          *RiskMetrics               `json:"risk"`
}

// PerformanceMetrics represents portfolio performance metrics
type PerformanceMetrics struct {
	Return1d      decimal.Decimal `json:"return_1d"`
	Return7d      decimal.Decimal `json:"return_7d"`
	Return30d     decimal.Decimal `json:"return_30d"`
	Return1y      decimal.Decimal `json:"return_1y"`
	Volatility    decimal.Decimal `json:"volatility"`
	SharpeRatio   decimal.Decimal `json:"sharpe_ratio"`
	MaxDrawdown   decimal.Decimal `json:"max_drawdown"`
	WinRate       decimal.Decimal `json:"win_rate"`
}

// Alert represents a portfolio alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Address     common.Address         `json:"address"`
	Chain       string                 `json:"chain"`
	Token       string                 `json:"token"`
	Threshold   decimal.Decimal        `json:"threshold"`
	CurrentValue decimal.Decimal       `json:"current_value"`
	CreatedAt   time.Time              `json:"created_at"`
	Acknowledged bool                  `json:"acknowledged"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ChainStatus represents the status of a blockchain
type ChainStatus struct {
	Chain           string          `json:"chain"`
	ChainID         int64           `json:"chain_id"`
	BlockNumber     uint64          `json:"block_number"`
	BlockTime       time.Duration   `json:"block_time"`
	GasPrice        decimal.Decimal `json:"gas_price"`
	IsHealthy       bool            `json:"is_healthy"`
	LastUpdated     time.Time       `json:"last_updated"`
	RPCLatency      time.Duration   `json:"rpc_latency"`
	SyncStatus      string          `json:"sync_status"`
	PeerCount       int             `json:"peer_count"`
}

// NetworkStats represents network statistics
type NetworkStats struct {
	TotalChains       int                        `json:"total_chains"`
	ActiveChains      int                        `json:"active_chains"`
	TotalTransactions uint64                     `json:"total_transactions"`
	TotalValue        decimal.Decimal            `json:"total_value"`
	ChainStats        map[string]*ChainStatus    `json:"chain_stats"`
	LastUpdated       time.Time                  `json:"last_updated"`
}

// Utility functions

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// IsValidChain checks if a chain is valid
func IsValidChain(chain string) bool {
	validChains := map[string]bool{
		"ethereum":  true,
		"bsc":       true,
		"polygon":   true,
		"arbitrum":  true,
		"optimism":  true,
		"avalanche": true,
		"fantom":    true,
		"gnosis":    true,
	}
	return validChains[chain]
}

// GetChainID returns the chain ID for a given chain name
func GetChainID(chain string) int64 {
	chainIDs := map[string]int64{
		"ethereum":  1,
		"bsc":       56,
		"polygon":   137,
		"arbitrum":  42161,
		"optimism":  10,
		"avalanche": 43114,
		"fantom":    250,
		"gnosis":    100,
	}
	return chainIDs[chain]
}

// GetNativeToken returns the native token for a given chain
func GetNativeToken(chain string) TokenConfig {
	nativeTokens := map[string]TokenConfig{
		"ethereum": {
			Symbol:   "ETH",
			Name:     "Ethereum",
			Decimals: 18,
			IsNative: true,
		},
		"bsc": {
			Symbol:   "BNB",
			Name:     "Binance Coin",
			Decimals: 18,
			IsNative: true,
		},
		"polygon": {
			Symbol:   "MATIC",
			Name:     "Polygon",
			Decimals: 18,
			IsNative: true,
		},
		"arbitrum": {
			Symbol:   "ETH",
			Name:     "Ethereum",
			Decimals: 18,
			IsNative: true,
		},
		"optimism": {
			Symbol:   "ETH",
			Name:     "Ethereum",
			Decimals: 18,
			IsNative: true,
		},
		"avalanche": {
			Symbol:   "AVAX",
			Name:     "Avalanche",
			Decimals: 18,
			IsNative: true,
		},
	}
	
	if token, exists := nativeTokens[chain]; exists {
		return token
	}
	
	// Default to ETH
	return nativeTokens["ethereum"]
}
