package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthereumClient represents an Ethereum blockchain client
type EthereumClient struct {
	client *ethclient.Client
	config config.BlockchainNetworkConfig
	logger *logger.Logger
}

// NewEthereumClient creates a new Ethereum client
func NewEthereumClient(cfg config.BlockchainNetworkConfig, logger *logger.Logger) (*EthereumClient, error) {
	// Connect to Ethereum node
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	return &EthereumClient{
		client: client,
		config: cfg,
		logger: logger.Named("ethereum"),
	}, nil
}

// GetBalance retrieves the balance of an address
func (c *EthereumClient) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	// Convert address string to common.Address
	addr := common.HexToAddress(address)

	// Get balance
	balance, err := c.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// GetNonce retrieves the nonce of an address
func (c *EthereumClient) GetNonce(ctx context.Context, address string) (uint64, error) {
	// Convert address string to common.Address
	addr := common.HexToAddress(address)

	// Get nonce
	nonce, err := c.client.NonceAt(ctx, addr, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	return nonce, nil
}

// GetGasPrice retrieves the current gas price
func (c *EthereumClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	// Get gas price
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	return gasPrice, nil
}

// EstimateGas estimates the gas required for a transaction
func (c *EthereumClient) EstimateGas(ctx context.Context, from, to string, value *big.Int, data []byte) (uint64, error) {
	// Convert addresses
	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)

	// Create call message
	msg := ethereum.CallMsg{
		From:  fromAddr,
		To:    &toAddr,
		Value: value,
		Data:  data,
	}

	// Estimate gas
	gas, err := c.client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	return gas, nil
}

// SendTransaction sends a signed transaction
func (c *EthereumClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	// Send transaction
	err := c.client.SendTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	return nil
}

// GetTransactionByHash retrieves a transaction by its hash
func (c *EthereumClient) GetTransactionByHash(ctx context.Context, hash string) (*types.Transaction, bool, error) {
	// Convert hash string to common.Hash
	txHash := common.HexToHash(hash)

	// Get transaction
	tx, isPending, err := c.client.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, isPending, nil
}

// GetTransactionReceipt retrieves a transaction receipt by its hash
func (c *EthereumClient) GetTransactionReceipt(ctx context.Context, hash string) (*types.Receipt, error) {
	// Convert hash string to common.Hash
	txHash := common.HexToHash(hash)

	// Get receipt
	receipt, err := c.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	return receipt, nil
}

// GetBlockByNumber retrieves a block by its number
func (c *EthereumClient) GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	// Get block
	block, err := c.client.BlockByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	return block, nil
}

// GetLatestBlockNumber retrieves the latest block number
func (c *EthereumClient) GetLatestBlockNumber(ctx context.Context) (*big.Int, error) {
	// Get header
	header, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block number: %w", err)
	}

	return header.Number, nil
}

// Close closes the client connection
func (c *EthereumClient) Close() {
	c.client.Close()
}
