package rpc

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// Client is a high-level RPC client with automatic failover and load balancing
type Client struct {
	logger      *logger.Logger
	nodeManager *NodeManager
	sessionID   string
	chain       string
}

// ClientConfig holds configuration for RPC client
type ClientConfig struct {
	Chain     string `json:"chain" yaml:"chain"`
	SessionID string `json:"session_id" yaml:"session_id"`
}

// NewClient creates a new RPC client
func NewClient(logger *logger.Logger, nodeManager *NodeManager, config ClientConfig) *Client {
	return &Client{
		logger:      logger.Named("rpc-client"),
		nodeManager: nodeManager,
		sessionID:   config.SessionID,
		chain:       config.Chain,
	}
}

// executeWithRetry executes a function with automatic retry and failover
func (c *Client) executeWithRetry(ctx context.Context, operation string, fn func(*ethclient.Client) error) error {
	maxRetries := c.nodeManager.config.MaxRetries
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Get healthy node
		node, err := c.nodeManager.GetHealthyNode(c.sessionID)
		if err != nil {
			c.logger.Warn("Failed to get healthy node",
				zap.String("operation", operation),
				zap.Int("attempt", attempt),
				zap.Error(err))
			
			if attempt == maxRetries {
				return fmt.Errorf("no healthy nodes available after %d attempts: %w", maxRetries, err)
			}
			
			// Wait before retry
			time.Sleep(c.nodeManager.config.FailoverTimeout)
			continue
		}

		// Filter by chain if specified
		if c.chain != "" && node.Config.Chain != c.chain {
			// Try to find node for specific chain
			nodes := c.nodeManager.GetHealthyNodes()
			var chainNode *RPCNode
			for _, n := range nodes {
				if n.Config.Chain == c.chain {
					chainNode = n
					break
				}
			}
			if chainNode != nil {
				node = chainNode
			}
		}

		// Execute operation
		startTime := time.Now()
		err = fn(node.Client)
		duration := time.Since(startTime)

		// Update node metrics
		node.mutex.Lock()
		node.Metrics.TotalRequests++
		node.Metrics.LastRequestTime = time.Now()
		
		if err != nil {
			node.Metrics.FailedRequests++
		} else {
			node.Metrics.SuccessfulRequests++
		}
		
		// Update average latency
		if node.Metrics.TotalRequests > 0 {
			totalLatency := node.Metrics.AverageLatency * time.Duration(node.Metrics.TotalRequests-1)
			node.Metrics.AverageLatency = (totalLatency + duration) / time.Duration(node.Metrics.TotalRequests)
		} else {
			node.Metrics.AverageLatency = duration
		}
		node.mutex.Unlock()

		if err == nil {
			c.logger.Debug("RPC operation successful",
				zap.String("operation", operation),
				zap.String("node_id", node.Config.ID),
				zap.Duration("duration", duration),
				zap.Int("attempt", attempt))
			return nil
		}

		c.logger.Warn("RPC operation failed",
			zap.String("operation", operation),
			zap.String("node_id", node.Config.ID),
			zap.Duration("duration", duration),
			zap.Int("attempt", attempt),
			zap.Error(err))

		// If this was the last attempt, return the error
		if attempt == maxRetries {
			return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, err)
		}

		// Wait before retry
		time.Sleep(c.nodeManager.config.FailoverTimeout)
	}

	return fmt.Errorf("operation failed after %d attempts", maxRetries)
}

// ChainID retrieves the chain ID
func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	var result *big.Int
	err := c.executeWithRetry(ctx, "ChainID", func(client *ethclient.Client) error {
		chainID, err := client.ChainID(ctx)
		if err != nil {
			return err
		}
		result = chainID
		return nil
	})
	return result, err
}

// BlockNumber retrieves the latest block number
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	var result uint64
	err := c.executeWithRetry(ctx, "BlockNumber", func(client *ethclient.Client) error {
		blockNumber, err := client.BlockNumber(ctx)
		if err != nil {
			return err
		}
		result = blockNumber
		return nil
	})
	return result, err
}

// BlockByNumber retrieves a block by number
func (c *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	var result *types.Block
	err := c.executeWithRetry(ctx, "BlockByNumber", func(client *ethclient.Client) error {
		block, err := client.BlockByNumber(ctx, number)
		if err != nil {
			return err
		}
		result = block
		return nil
	})
	return result, err
}

// BlockByHash retrieves a block by hash
func (c *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	var result *types.Block
	err := c.executeWithRetry(ctx, "BlockByHash", func(client *ethclient.Client) error {
		block, err := client.BlockByHash(ctx, hash)
		if err != nil {
			return err
		}
		result = block
		return nil
	})
	return result, err
}

// HeaderByNumber retrieves a block header by number
func (c *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	var result *types.Header
	err := c.executeWithRetry(ctx, "HeaderByNumber", func(client *ethclient.Client) error {
		header, err := client.HeaderByNumber(ctx, number)
		if err != nil {
			return err
		}
		result = header
		return nil
	})
	return result, err
}

// TransactionByHash retrieves a transaction by hash
func (c *Client) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	var result *types.Transaction
	var isPending bool
	err := c.executeWithRetry(ctx, "TransactionByHash", func(client *ethclient.Client) error {
		tx, pending, err := client.TransactionByHash(ctx, hash)
		if err != nil {
			return err
		}
		result = tx
		isPending = pending
		return nil
	})
	return result, isPending, err
}

// TransactionReceipt retrieves a transaction receipt
func (c *Client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	var result *types.Receipt
	err := c.executeWithRetry(ctx, "TransactionReceipt", func(client *ethclient.Client) error {
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err != nil {
			return err
		}
		result = receipt
		return nil
	})
	return result, err
}

// BalanceAt retrieves the balance of an account
func (c *Client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var result *big.Int
	err := c.executeWithRetry(ctx, "BalanceAt", func(client *ethclient.Client) error {
		balance, err := client.BalanceAt(ctx, account, blockNumber)
		if err != nil {
			return err
		}
		result = balance
		return nil
	})
	return result, err
}

// NonceAt retrieves the nonce of an account
func (c *Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	var result uint64
	err := c.executeWithRetry(ctx, "NonceAt", func(client *ethclient.Client) error {
		nonce, err := client.NonceAt(ctx, account, blockNumber)
		if err != nil {
			return err
		}
		result = nonce
		return nil
	})
	return result, err
}

// SendTransaction sends a transaction
func (c *Client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return c.executeWithRetry(ctx, "SendTransaction", func(client *ethclient.Client) error {
		return client.SendTransaction(ctx, tx)
	})
}

// CallContract executes a contract call
func (c *Client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	var result []byte
	err := c.executeWithRetry(ctx, "CallContract", func(client *ethclient.Client) error {
		data, err := client.CallContract(ctx, msg, blockNumber)
		if err != nil {
			return err
		}
		result = data
		return nil
	})
	return result, err
}

// EstimateGas estimates gas for a transaction
func (c *Client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	var result uint64
	err := c.executeWithRetry(ctx, "EstimateGas", func(client *ethclient.Client) error {
		gas, err := client.EstimateGas(ctx, msg)
		if err != nil {
			return err
		}
		result = gas
		return nil
	})
	return result, err
}

// SuggestGasPrice suggests a gas price
func (c *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	var result *big.Int
	err := c.executeWithRetry(ctx, "SuggestGasPrice", func(client *ethclient.Client) error {
		gasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return err
		}
		result = gasPrice
		return nil
	})
	return result, err
}

// SuggestGasTipCap suggests a gas tip cap (EIP-1559)
func (c *Client) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	var result *big.Int
	err := c.executeWithRetry(ctx, "SuggestGasTipCap", func(client *ethclient.Client) error {
		gasTipCap, err := client.SuggestGasTipCap(ctx)
		if err != nil {
			return err
		}
		result = gasTipCap
		return nil
	})
	return result, err
}

// FilterLogs executes a log filter query
func (c *Client) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	var result []types.Log
	err := c.executeWithRetry(ctx, "FilterLogs", func(client *ethclient.Client) error {
		logs, err := client.FilterLogs(ctx, q)
		if err != nil {
			return err
		}
		result = logs
		return nil
	})
	return result, err
}

// SubscribeFilterLogs subscribes to log events
func (c *Client) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	// For subscriptions, we need to use a specific node rather than load balancing
	node, err := c.nodeManager.GetHealthyNode(c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy node for subscription: %w", err)
	}

	// Filter by chain if specified
	if c.chain != "" && node.Config.Chain != c.chain {
		nodes := c.nodeManager.GetHealthyNodes()
		for _, n := range nodes {
			if n.Config.Chain == c.chain {
				node = n
				break
			}
		}
	}

	c.logger.Info("Creating log subscription",
		zap.String("node_id", node.Config.ID),
		zap.String("chain", node.Config.Chain))

	return node.Client.SubscribeFilterLogs(ctx, q, ch)
}

// SubscribeNewHead subscribes to new block headers
func (c *Client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	// For subscriptions, we need to use a specific node rather than load balancing
	node, err := c.nodeManager.GetHealthyNode(c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy node for subscription: %w", err)
	}

	// Filter by chain if specified
	if c.chain != "" && node.Config.Chain != c.chain {
		nodes := c.nodeManager.GetHealthyNodes()
		for _, n := range nodes {
			if n.Config.Chain == c.chain {
				node = n
				break
			}
		}
	}

	c.logger.Info("Creating new head subscription",
		zap.String("node_id", node.Config.ID),
		zap.String("chain", node.Config.Chain))

	return node.Client.SubscribeNewHead(ctx, ch)
}

// GetNodeMetrics returns metrics for the current session's preferred node
func (c *Client) GetNodeMetrics() (*NodeMetrics, error) {
	node, err := c.nodeManager.GetHealthyNode(c.sessionID)
	if err != nil {
		return nil, err
	}

	node.mutex.RLock()
	metrics := *node.Metrics
	node.mutex.RUnlock()

	return &metrics, nil
}

// GetNodeHealth returns health status for the current session's preferred node
func (c *Client) GetNodeHealth() (*NodeHealth, error) {
	node, err := c.nodeManager.GetHealthyNode(c.sessionID)
	if err != nil {
		return nil, err
	}

	node.mutex.RLock()
	health := *node.Health
	node.mutex.RUnlock()

	return &health, nil
}

// SetChain sets the preferred blockchain chain
func (c *Client) SetChain(chain string) {
	c.chain = chain
}

// GetChain returns the current preferred chain
func (c *Client) GetChain() string {
	return c.chain
}

// SetSessionID sets the session ID for sticky sessions
func (c *Client) SetSessionID(sessionID string) {
	c.sessionID = sessionID
}

// GetSessionID returns the current session ID
func (c *Client) GetSessionID() string {
	return c.sessionID
}
