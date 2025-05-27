package blockchain

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/shopspring/decimal"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// SolanaClient represents a Solana blockchain client
type SolanaClient struct {
	rpcClient *rpc.Client
	wsClient  *ws.Client
	config    config.SolanaNetworkConfig
	logger    *logger.Logger
}

// NewSolanaClient creates a new Solana client
func NewSolanaClient(cfg config.SolanaNetworkConfig, logger *logger.Logger) (*SolanaClient, error) {
	// Create RPC client
	rpcClient := rpc.New(cfg.RPCURL)

	// Create WebSocket client (optional)
	var wsClient *ws.Client
	if cfg.WSURL != "" {
		var err error
		wsClient, err = ws.Connect(context.Background(), cfg.WSURL)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to connect to Solana WebSocket: %v", err))
		}
	}

	return &SolanaClient{
		rpcClient: rpcClient,
		wsClient:  wsClient,
		config:    cfg,
		logger:    logger.Named("solana"),
	}, nil
}

// GetBalance retrieves the SOL balance for a given address
func (c *SolanaClient) GetBalance(ctx context.Context, address string) (decimal.Decimal, error) {
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid address: %w", err)
	}

	balance, err := c.rpcClient.GetBalance(ctx, pubKey, rpc.CommitmentConfirmed)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get balance: %w", err)
	}

	// Convert lamports to SOL (1 SOL = 1e9 lamports)
	solBalance := decimal.NewFromUint64(balance.Value).Div(decimal.NewFromInt(1e9))
	
	c.logger.Info(fmt.Sprintf("Retrieved SOL balance for %s: %s", address, solBalance.String()))
	return solBalance, nil
}

// GetTokenBalance retrieves the SPL token balance for a given address and mint
func (c *SolanaClient) GetTokenBalance(ctx context.Context, address, mintAddress string) (decimal.Decimal, uint8, error) {
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return decimal.Zero, 0, fmt.Errorf("invalid address: %w", err)
	}

	mintPubKey, err := solana.PublicKeyFromBase58(mintAddress)
	if err != nil {
		return decimal.Zero, 0, fmt.Errorf("invalid mint address: %w", err)
	}

	// Get token accounts by owner
	tokenAccounts, err := c.rpcClient.GetTokenAccountsByOwner(
		ctx,
		pubKey,
		&rpc.GetTokenAccountsConfig{
			Mint: &mintPubKey,
		},
		&rpc.GetTokenAccountsOpts{
			Commitment: rpc.CommitmentConfirmed,
		},
	)
	if err != nil {
		return decimal.Zero, 0, fmt.Errorf("failed to get token accounts: %w", err)
	}

	if len(tokenAccounts.Value) == 0 {
		return decimal.Zero, 0, nil
	}

	// Get the first token account balance
	tokenAccount := tokenAccounts.Value[0]
	balance, err := c.rpcClient.GetTokenAccountBalance(
		ctx,
		tokenAccount.Pubkey,
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		return decimal.Zero, 0, fmt.Errorf("failed to get token balance: %w", err)
	}

	amount, err := decimal.NewFromString(balance.Value.Amount)
	if err != nil {
		return decimal.Zero, 0, fmt.Errorf("failed to parse token amount: %w", err)
	}

	c.logger.Info(fmt.Sprintf("Retrieved token balance for %s (mint: %s): %s", address, mintAddress, amount.String()))
	return amount, balance.Value.Decimals, nil
}

// SendTransaction sends a transaction to the Solana network
func (c *SolanaClient) SendTransaction(ctx context.Context, transaction *solana.Transaction) (string, error) {
	signature, err := c.rpcClient.SendTransaction(ctx, transaction)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	c.logger.Info(fmt.Sprintf("Transaction sent with signature: %s", signature.String()))
	return signature.String(), nil
}

// ConfirmTransaction waits for transaction confirmation
func (c *SolanaClient) ConfirmTransaction(ctx context.Context, signature string) error {
	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	// Wait for confirmation with timeout
	timeout := 30 * time.Second
	if c.config.Timeout != "" {
		if parsedTimeout, err := time.ParseDuration(c.config.Timeout); err == nil {
			timeout = parsedTimeout
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("transaction confirmation timeout")
		default:
			status, err := c.rpcClient.GetSignatureStatuses(ctx, true, sig)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			if len(status.Value) > 0 && status.Value[0] != nil {
				if status.Value[0].Err != nil {
					return fmt.Errorf("transaction failed: %v", status.Value[0].Err)
				}
				
				if status.Value[0].ConfirmationStatus == rpc.ConfirmationStatusConfirmed ||
					status.Value[0].ConfirmationStatus == rpc.ConfirmationStatusFinalized {
					c.logger.Info(fmt.Sprintf("Transaction confirmed: %s", signature))
					return nil
				}
			}

			time.Sleep(1 * time.Second)
		}
	}
}

// GetRecentBlockhash gets a recent blockhash for transaction creation
func (c *SolanaClient) GetRecentBlockhash(ctx context.Context) (solana.Hash, error) {
	response, err := c.rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return solana.Hash{}, fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	return response.Value.Blockhash, nil
}

// GetMinimumBalanceForRentExemption gets the minimum balance required for rent exemption
func (c *SolanaClient) GetMinimumBalanceForRentExemption(ctx context.Context, dataSize uint64) (uint64, error) {
	balance, err := c.rpcClient.GetMinimumBalanceForRentExemption(ctx, dataSize, rpc.CommitmentConfirmed)
	if err != nil {
		return 0, fmt.Errorf("failed to get minimum balance for rent exemption: %w", err)
	}

	return balance, nil
}

// CreateAccount creates a new Solana account
func (c *SolanaClient) CreateAccount(ctx context.Context, fromAccount, newAccount solana.PrivateKey, space uint64, owner solana.PublicKey) (*solana.Transaction, error) {
	// Get recent blockhash
	blockhash, err := c.GetRecentBlockhash(ctx)
	if err != nil {
		return nil, err
	}

	// Get minimum balance for rent exemption
	minBalance, err := c.GetMinimumBalanceForRentExemption(ctx, space)
	if err != nil {
		return nil, err
	}

	// Create the transaction
	transaction, err := solana.NewTransaction(
		[]solana.Instruction{
			solana.NewCreateAccountInstruction(
				minBalance,
				space,
				owner,
				fromAccount.PublicKey(),
				newAccount.PublicKey(),
			).Build(),
		},
		blockhash,
		solana.TransactionPayer(fromAccount.PublicKey()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Sign the transaction
	_, err = transaction.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if key.Equals(fromAccount.PublicKey()) {
				return &fromAccount
			}
			if key.Equals(newAccount.PublicKey()) {
				return &newAccount
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return transaction, nil
}

// Close closes the Solana client connections
func (c *SolanaClient) Close() error {
	if c.wsClient != nil {
		return c.wsClient.Close()
	}
	return nil
}

// GetCluster returns the configured cluster
func (c *SolanaClient) GetCluster() string {
	return c.config.Cluster
}

// GetCommitment returns the configured commitment level
func (c *SolanaClient) GetCommitment() string {
	return c.config.Commitment
}
