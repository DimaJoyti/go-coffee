package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/wallet"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/blockchain"
	cryptoUtil "github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/crypto"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/models"
)

// Service provides transaction operations
type Service struct {
	repo           Repository
	walletRepo     wallet.Repository
	ethClient      *blockchain.EthereumClient
	bscClient      *blockchain.EthereumClient
	polygonClient  *blockchain.EthereumClient
	keyManager     *cryptoUtil.KeyManager
	logger         *logger.Logger
}

// NewService creates a new transaction service
func NewService(
	repo Repository,
	walletRepo wallet.Repository,
	ethClient *blockchain.EthereumClient,
	bscClient *blockchain.EthereumClient,
	polygonClient *blockchain.EthereumClient,
	keyManager *cryptoUtil.KeyManager,
	logger *logger.Logger,
) *Service {
	return &Service{
		repo:          repo,
		walletRepo:    walletRepo,
		ethClient:     ethClient,
		bscClient:     bscClient,
		polygonClient: polygonClient,
		keyManager:    keyManager,
		logger:        logger.Named("transaction-service"),
	}
}

// CreateTransaction creates a new transaction
func (s *Service) CreateTransaction(ctx context.Context, req *models.CreateTransactionRequest) (*models.CreateTransactionResponse, error) {
	s.logger.Info(fmt.Sprintf("Creating transaction from wallet %s to %s", req.WalletID, req.To))

	// Get wallet
	wallet, err := s.walletRepo.GetWallet(ctx, req.WalletID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error(fmt.Sprintf("Wallet not found: %s", req.WalletID))
			return nil, fmt.Errorf("wallet not found: %s", req.WalletID)
		}
		s.logger.Error(fmt.Sprintf("Failed to get wallet: %v", err))
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Get blockchain client
	client, err := s.getBlockchainClient(models.Chain(wallet.Chain))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get blockchain client: %v", err))
		return nil, fmt.Errorf("failed to get blockchain client: %w", err)
	}

	// Get keystore
	keystore, err := s.walletRepo.GetKeystore(ctx, wallet.ID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get keystore: %v", err))
		return nil, fmt.Errorf("failed to get keystore: %w", err)
	}

	// Decrypt private key
	privateKeyHex, err := s.keyManager.DecryptPrivateKey(keystore, "temporary-passphrase")
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decrypt private key: %v", err))
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex[2:]) // Remove "0x" prefix
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to parse private key: %v", err))
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Get nonce
	var nonce uint64
	if req.Nonce > 0 {
		nonce = req.Nonce
	} else {
		nonce, err = client.GetNonce(ctx, wallet.Address)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to get nonce: %v", err))
			return nil, fmt.Errorf("failed to get nonce: %w", err)
		}
	}

	// Get gas price
	var gasPrice *big.Int
	if req.GasPrice != "" {
		gasPrice, _ = new(big.Int).SetString(req.GasPrice, 10)
	} else {
		gasPrice, err = client.GetGasPrice(ctx)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to get gas price: %v", err))
			return nil, fmt.Errorf("failed to get gas price: %w", err)
		}
	}

	// Parse value
	value, ok := new(big.Int).SetString(req.Value, 10)
	if !ok {
		s.logger.Error(fmt.Sprintf("Failed to parse value: %s", req.Value))
		return nil, fmt.Errorf("failed to parse value: %s", req.Value)
	}

	// Parse data
	var data []byte
	if req.Data != "" {
		data = common.FromHex(req.Data)
	}

	// Get gas limit
	var gas uint64
	if req.Gas > 0 {
		gas = req.Gas
	} else {
		gas, err = client.EstimateGas(ctx, wallet.Address, req.To, value, data)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to estimate gas: %v", err))
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}
	}

	// Get chain ID
	var chainID *big.Int
	switch models.Chain(wallet.Chain) {
	case models.ChainEthereum:
		chainID = big.NewInt(1) // Ethereum Mainnet
	case models.ChainBSC:
		chainID = big.NewInt(56) // Binance Smart Chain Mainnet
	case models.ChainPolygon:
		chainID = big.NewInt(137) // Polygon Mainnet
	default:
		s.logger.Error(fmt.Sprintf("Unsupported chain: %s", wallet.Chain))
		return nil, fmt.Errorf("unsupported chain: %s", wallet.Chain)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(req.To),
		value,
		gas,
		gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to sign transaction: %v", err))
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to send transaction: %v", err))
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Create transaction record
	txHash := signedTx.Hash().Hex()
	transaction := &models.Transaction{
		ID:        uuid.New().String(),
		UserID:    wallet.UserID,
		WalletID:  wallet.ID,
		Hash:      txHash,
		From:      wallet.Address,
		To:        req.To,
		Value:     req.Value,
		Gas:       gas,
		GasPrice:  gasPrice.String(),
		Nonce:     nonce,
		Data:      req.Data,
		Chain:     wallet.Chain,
		Status:    string(models.TransactionStatusPending),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save transaction to database
	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save transaction: %v", err))
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Transaction created successfully: %s", transaction.ID))

	// Return response
	return &models.CreateTransactionResponse{
		Transaction: *transaction,
	}, nil
}

// GetTransaction retrieves a transaction by ID
func (s *Service) GetTransaction(ctx context.Context, req *models.GetTransactionRequest) (*models.GetTransactionResponse, error) {
	s.logger.Info(fmt.Sprintf("Getting transaction %s", req.ID))

	// Get transaction from database
	transaction, err := s.repo.GetTransaction(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error(fmt.Sprintf("Transaction not found: %s", req.ID))
			return nil, fmt.Errorf("transaction not found: %s", req.ID)
		}
		s.logger.Error(fmt.Sprintf("Failed to get transaction: %v", err))
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Update transaction status if pending
	if transaction.Status == string(models.TransactionStatusPending) {
		if err := s.updateTransactionStatus(ctx, transaction); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to update transaction status: %v", err))
			// Continue with the current status
		}
	}

	s.logger.Info(fmt.Sprintf("Transaction retrieved successfully: %s", transaction.ID))

	// Return response
	return &models.GetTransactionResponse{
		Transaction: *transaction,
	}, nil
}

// GetTransactionByHash retrieves a transaction by hash
func (s *Service) GetTransactionByHash(ctx context.Context, req *models.GetTransactionByHashRequest) (*models.GetTransactionByHashResponse, error) {
	s.logger.Info(fmt.Sprintf("Getting transaction by hash %s on chain %s", req.Hash, req.Chain))

	// Get transaction from database
	transaction, err := s.repo.GetTransactionByHash(ctx, req.Hash, string(req.Chain))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error(fmt.Sprintf("Transaction not found: %s", req.Hash))
			return nil, fmt.Errorf("transaction not found: %s", req.Hash)
		}
		s.logger.Error(fmt.Sprintf("Failed to get transaction: %v", err))
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Update transaction status if pending
	if transaction.Status == string(models.TransactionStatusPending) {
		if err := s.updateTransactionStatus(ctx, transaction); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to update transaction status: %v", err))
			// Continue with the current status
		}
	}

	s.logger.Info(fmt.Sprintf("Transaction retrieved successfully: %s", transaction.ID))

	// Return response
	return &models.GetTransactionByHashResponse{
		Transaction: *transaction,
	}, nil
}

// ListTransactions lists all transactions for a wallet
func (s *Service) ListTransactions(ctx context.Context, req *models.ListTransactionsRequest) (*models.ListTransactionsResponse, error) {
	s.logger.Info(fmt.Sprintf("Listing transactions for user %s", req.UserID))

	// Get transactions from database
	transactions, total, err := s.repo.ListTransactions(ctx, req.UserID, req.WalletID, string(req.Status), string(req.Chain), req.Limit, req.Offset)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list transactions: %v", err))
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Update pending transactions
	for i, tx := range transactions {
		if tx.Status == string(models.TransactionStatusPending) {
			if err := s.updateTransactionStatus(ctx, &transactions[i]); err != nil {
				s.logger.Error(fmt.Sprintf("Failed to update transaction status: %v", err))
				// Continue with the current status
			}
		}
	}

	s.logger.Info(fmt.Sprintf("Listed %d transactions for user %s", len(transactions), req.UserID))

	// Return response
	return &models.ListTransactionsResponse{
		Transactions: transactions,
		Total:        total,
	}, nil
}

// EstimateGas estimates the gas required for a transaction
func (s *Service) EstimateGas(ctx context.Context, req *models.EstimateGasRequest) (*models.EstimateGasResponse, error) {
	s.logger.Info(fmt.Sprintf("Estimating gas for transaction from %s to %s on chain %s", req.From, req.To, req.Chain))

	// Get blockchain client
	client, err := s.getBlockchainClient(req.Chain)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get blockchain client: %v", err))
		return nil, fmt.Errorf("failed to get blockchain client: %w", err)
	}

	// Parse value
	value := new(big.Int)
	if req.Value != "" {
		var ok bool
		value, ok = value.SetString(req.Value, 10)
		if !ok {
			s.logger.Error(fmt.Sprintf("Failed to parse value: %s", req.Value))
			return nil, fmt.Errorf("failed to parse value: %s", req.Value)
		}
	}

	// Parse data
	var data []byte
	if req.Data != "" {
		data = common.FromHex(req.Data)
	}

	// Estimate gas
	gas, err := client.EstimateGas(ctx, req.From, req.To, value, data)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to estimate gas: %v", err))
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Gas estimated successfully: %d", gas))

	// Return response
	return &models.EstimateGasResponse{
		Gas: gas,
	}, nil
}

// GetGasPrice retrieves the current gas price
func (s *Service) GetGasPrice(ctx context.Context, req *models.GetGasPriceRequest) (*models.GetGasPriceResponse, error) {
	s.logger.Info(fmt.Sprintf("Getting gas price for chain %s", req.Chain))

	// Get blockchain client
	client, err := s.getBlockchainClient(req.Chain)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get blockchain client: %v", err))
		return nil, fmt.Errorf("failed to get blockchain client: %w", err)
	}

	// Get gas price
	gasPrice, err := client.GetGasPrice(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get gas price: %v", err))
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Calculate slow, average, and fast gas prices
	slow := new(big.Int).Mul(gasPrice, big.NewInt(8))
	slow = slow.Div(slow, big.NewInt(10))
	average := gasPrice
	fast := new(big.Int).Mul(gasPrice, big.NewInt(12))
	fast = fast.Div(fast, big.NewInt(10))

	s.logger.Info(fmt.Sprintf("Gas price retrieved successfully: %s", gasPrice.String()))

	// Return response
	return &models.GetGasPriceResponse{
		GasPrice: gasPrice.String(),
		Slow:     slow.String(),
		Average:  average.String(),
		Fast:     fast.String(),
	}, nil
}

// GetTransactionReceipt retrieves a transaction receipt
func (s *Service) GetTransactionReceipt(ctx context.Context, req *models.GetTransactionReceiptRequest) (*models.GetTransactionReceiptResponse, error) {
	s.logger.Info(fmt.Sprintf("Getting transaction receipt for hash %s on chain %s", req.Hash, req.Chain))

	// Get blockchain client
	client, err := s.getBlockchainClient(req.Chain)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get blockchain client: %v", err))
		return nil, fmt.Errorf("failed to get blockchain client: %w", err)
	}

	// Get receipt
	receipt, err := client.GetTransactionReceipt(ctx, req.Hash)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get transaction receipt: %v", err))
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Convert logs
	logs := make([]models.Log, len(receipt.Logs))
	for i, log := range receipt.Logs {
		topics := make([]string, len(log.Topics))
		for j, topic := range log.Topics {
			topics[j] = topic.Hex()
		}

		logs[i] = models.Log{
			Address:     log.Address.Hex(),
			Topics:      topics,
			Data:        common.Bytes2Hex(log.Data),
			BlockNumber: log.BlockNumber,
			TxHash:      log.TxHash.Hex(),
			TxIndex:     log.TxIndex,
			BlockHash:   log.BlockHash.Hex(),
			Index:       log.Index,
			Removed:     log.Removed,
		}
	}

	s.logger.Info(fmt.Sprintf("Transaction receipt retrieved successfully for hash %s", req.Hash))

	// Return response
	return &models.GetTransactionReceiptResponse{
		BlockHash:         receipt.BlockHash.Hex(),
		BlockNumber:       receipt.BlockNumber.Uint64(),
		ContractAddress:   receipt.ContractAddress.Hex(),
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		From:              receipt.From.Hex(),
		GasUsed:           receipt.GasUsed,
		Status:            receipt.Status == 1,
		To:                receipt.To.Hex(),
		TransactionHash:   receipt.TxHash.Hex(),
		TransactionIndex:  receipt.TransactionIndex,
		Logs:              logs,
	}, nil
}

// updateTransactionStatus updates the status of a transaction
func (s *Service) updateTransactionStatus(ctx context.Context, transaction *models.Transaction) error {
	// Get blockchain client
	client, err := s.getBlockchainClient(models.Chain(transaction.Chain))
	if err != nil {
		return fmt.Errorf("failed to get blockchain client: %w", err)
	}

	// Get transaction receipt
	receipt, err := client.GetTransactionReceipt(ctx, transaction.Hash)
	if err != nil {
		// Transaction not yet mined
		return nil
	}

	// Update transaction status
	if receipt.Status == 1 {
		transaction.Status = string(models.TransactionStatusConfirmed)
	} else {
		transaction.Status = string(models.TransactionStatusFailed)
	}

	// Update transaction details
	transaction.BlockNumber = receipt.BlockNumber.Uint64()
	transaction.BlockHash = receipt.BlockHash.Hex()

	// Get latest block number
	latestBlock, err := client.GetLatestBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block number: %w", err)
	}

	// Calculate confirmations
	if transaction.BlockNumber > 0 {
		confirmations := latestBlock.Uint64() - transaction.BlockNumber
		transaction.Confirmations = confirmations
	}

	// Update transaction in database
	if err := s.repo.UpdateTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// getBlockchainClient returns the appropriate blockchain client based on the chain
func (s *Service) getBlockchainClient(chain models.Chain) (*blockchain.EthereumClient, error) {
	switch chain {
	case models.ChainEthereum:
		return s.ethClient, nil
	case models.ChainBSC:
		return s.bscClient, nil
	case models.ChainPolygon:
		return s.polygonClient, nil
	default:
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
}
