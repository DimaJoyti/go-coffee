package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
)

// EthereumClient defines the interface for Ethereum blockchain operations
type EthereumClient interface {
	// Connection management
	Connect(ctx context.Context, rpcURL string) error
	Close() error
	IsConnected() bool

	// Block operations
	GetLatestBlockNumber(ctx context.Context) (uint64, error)
	GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)

	// Transaction operations
	GetTransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	EstimateGas(ctx context.Context, msg interface{}) (uint64, error)

	// Account operations
	GetBalance(ctx context.Context, address common.Address) (*big.Int, error)
	GetNonce(ctx context.Context, address common.Address) (uint64, error)

	// Contract operations
	CallContract(ctx context.Context, call interface{}, blockNumber *big.Int) ([]byte, error)
	
	// Token operations
	GetTokenBalance(ctx context.Context, tokenAddress, walletAddress common.Address) (*big.Int, error)
	GetTokenInfo(ctx context.Context, tokenAddress common.Address) (*TokenInfo, error)

	// Price operations
	GetTokenPrice(ctx context.Context, tokenAddress common.Address) (decimal.Decimal, error)
}

// SolanaClient defines the interface for Solana blockchain operations
type SolanaClient interface {
	// Connection management
	Connect(ctx context.Context, rpcURL string) error
	Close() error
	IsConnected() bool

	// Account operations
	GetBalance(ctx context.Context, address string) (uint64, error)
	GetTokenBalance(ctx context.Context, tokenMint, walletAddress string) (uint64, error)

	// Transaction operations
	SendTransaction(ctx context.Context, transaction []byte) (string, error)
	GetTransaction(ctx context.Context, signature string) (*SolanaTransaction, error)

	// Token operations
	GetTokenInfo(ctx context.Context, mintAddress string) (*SolanaTokenInfo, error)
	GetTokenPrice(ctx context.Context, mintAddress string) (decimal.Decimal, error)

	// Program operations
	CallProgram(ctx context.Context, programID string, data []byte) ([]byte, error)
}

// TokenInfo represents Ethereum token information
type TokenInfo struct {
	Address     common.Address  `json:"address"`
	Name        string          `json:"name"`
	Symbol      string          `json:"symbol"`
	Decimals    uint8           `json:"decimals"`
	TotalSupply *big.Int        `json:"total_supply"`
	Price       decimal.Decimal `json:"price"`
}

// SolanaTokenInfo represents Solana token information
type SolanaTokenInfo struct {
	MintAddress string          `json:"mint_address"`
	Name        string          `json:"name"`
	Symbol      string          `json:"symbol"`
	Decimals    uint8           `json:"decimals"`
	Supply      uint64          `json:"supply"`
	Price       decimal.Decimal `json:"price"`
}

// SolanaTransaction represents a Solana transaction
type SolanaTransaction struct {
	Signature   string                 `json:"signature"`
	Slot        uint64                 `json:"slot"`
	BlockTime   int64                  `json:"block_time"`
	Meta        *SolanaTransactionMeta `json:"meta"`
	Transaction interface{}            `json:"transaction"`
}

// SolanaTransactionMeta represents Solana transaction metadata
type SolanaTransactionMeta struct {
	Err               interface{} `json:"err"`
	Fee               uint64      `json:"fee"`
	PreBalances       []uint64    `json:"pre_balances"`
	PostBalances      []uint64    `json:"post_balances"`
	InnerInstructions []interface{} `json:"inner_instructions"`
	LogMessages       []string    `json:"log_messages"`
}

// ChainType represents different blockchain networks
type ChainType string

const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypeBSC      ChainType = "bsc"
	ChainTypePolygon  ChainType = "polygon"
	ChainTypeArbitrum ChainType = "arbitrum"
	ChainTypeOptimism ChainType = "optimism"
	ChainTypeSolana   ChainType = "solana"
)

// NetworkConfig represents blockchain network configuration
type NetworkConfig struct {
	ChainID     int64     `json:"chain_id"`
	Name        string    `json:"name"`
	Type        ChainType `json:"type"`
	RPCURL      string    `json:"rpc_url"`
	WSUrl       string    `json:"ws_url,omitempty"`
	ExplorerURL string    `json:"explorer_url"`
	NativeCurrency Currency `json:"native_currency"`
}

// Currency represents a blockchain's native currency
type Currency struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

// ClientFactory creates blockchain clients
type ClientFactory interface {
	CreateEthereumClient(config NetworkConfig) (EthereumClient, error)
	CreateSolanaClient(config NetworkConfig) (SolanaClient, error)
}
