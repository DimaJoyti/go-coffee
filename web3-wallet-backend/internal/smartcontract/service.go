package smartcontract

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/yourusername/web3-wallet-backend/internal/transaction"
	"github.com/yourusername/web3-wallet-backend/internal/wallet"
	"github.com/yourusername/web3-wallet-backend/pkg/blockchain"
	cryptoUtil "github.com/yourusername/web3-wallet-backend/pkg/crypto"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
	"github.com/yourusername/web3-wallet-backend/pkg/models"
)

// Service provides smart contract operations
type Service struct {
	repo           Repository
	walletRepo     wallet.Repository
	txRepo         transaction.Repository
	ethClient      *blockchain.EthereumClient
	bscClient      *blockchain.EthereumClient
	polygonClient  *blockchain.EthereumClient
	keyManager     *cryptoUtil.KeyManager
	logger         *logger.Logger
}

// NewService creates a new smart contract service
func NewService(
	repo Repository,
	walletRepo wallet.Repository,
	txRepo transaction.Repository,
	ethClient *blockchain.EthereumClient,
	bscClient *blockchain.EthereumClient,
	polygonClient *blockchain.EthereumClient,
	keyManager *cryptoUtil.KeyManager,
	logger *logger.Logger,
) *Service {
	return &Service{
		repo:          repo,
		walletRepo:    walletRepo,
		txRepo:        txRepo,
		ethClient:     ethClient,
		bscClient:     bscClient,
		polygonClient: polygonClient,
		keyManager:    keyManager,
		logger:        logger.Named("smartcontract-service"),
	}
}

// DeployContract deploys a new smart contract
func (s *Service) DeployContract(ctx context.Context, req *models.DeployContractRequest) (*models.DeployContractResponse, error) {
	s.logger.Info(fmt.Sprintf("Deploying contract %s on chain %s", req.Name, req.Chain))

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

	// Parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(req.ABI))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to parse ABI: %v", err))
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Parse bytecode
	bytecode := common.FromHex(req.Bytecode)

	// Get nonce
	nonce, err := client.GetNonce(ctx, wallet.Address)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get nonce: %v", err))
		return nil, fmt.Errorf("failed to get nonce: %w", err)
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

	// Create auth
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create auth: %v", err))
		return nil, fmt.Errorf("failed to create auth: %w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = req.Gas
	auth.GasPrice = gasPrice

	// Convert arguments
	var args []interface{}
	for _, arg := range req.Arguments {
		args = append(args, arg)
	}

	// Deploy contract
	var tx *types.Transaction
	var contractAddress common.Address

	// Create contract
	parsed, err := parsedABI.Constructor.Inputs.Pack(args...)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to pack arguments: %v", err))
		return nil, fmt.Errorf("failed to pack arguments: %w", err)
	}

	// Create transaction data
	data := append(bytecode, parsed...)

	// Create transaction
	tx = types.NewContractCreation(
		nonce,
		big.NewInt(0),
		req.Gas,
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

	// Get contract address
	contractAddress = crypto.CreateAddress(common.HexToAddress(wallet.Address), nonce)

	// Create contract record
	contract := &models.Contract{
		ID:        uuid.New().String(),
		UserID:    wallet.UserID,
		Name:      req.Name,
		Address:   contractAddress.Hex(),
		Chain:     string(req.Chain),
		ABI:       req.ABI,
		Bytecode:  req.Bytecode,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save contract to database
	if err := s.repo.CreateContract(ctx, contract); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save contract: %v", err))
		return nil, fmt.Errorf("failed to save contract: %w", err)
	}

	// Create transaction record
	txHash := signedTx.Hash().Hex()
	transaction := &models.Transaction{
		ID:        uuid.New().String(),
		UserID:    wallet.UserID,
		WalletID:  wallet.ID,
		Hash:      txHash,
		From:      wallet.Address,
		To:        "",
		Value:     "0",
		Gas:       req.Gas,
		GasPrice:  gasPrice.String(),
		Nonce:     nonce,
		Data:      "0x" + hex.EncodeToString(data),
		Chain:     wallet.Chain,
		Status:    string(models.TransactionStatusPending),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save transaction to database
	if err := s.txRepo.CreateTransaction(ctx, transaction); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save transaction: %v", err))
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Contract deployed successfully: %s", contract.ID))

	// Return response
	return &models.DeployContractResponse{
		Contract: *contract,
		Transaction: models.Transaction{
			ID:       transaction.ID,
			Hash:     transaction.Hash,
			From:     transaction.From,
			To:       transaction.To,
			Value:    transaction.Value,
			Gas:      transaction.Gas,
			GasPrice: transaction.GasPrice,
			Status:   transaction.Status,
		},
	}, nil
}

// ImportContract imports an existing contract
func (s *Service) ImportContract(ctx context.Context, req *models.ImportContractRequest) (*models.ImportContractResponse, error) {
	s.logger.Info(fmt.Sprintf("Importing contract %s on chain %s", req.Name, req.Chain))

	// Create contract record
	contract := &models.Contract{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Name:      req.Name,
		Address:   req.Address,
		Chain:     string(req.Chain),
		ABI:       req.ABI,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save contract to database
	if err := s.repo.CreateContract(ctx, contract); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save contract: %v", err))
		return nil, fmt.Errorf("failed to save contract: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Contract imported successfully: %s", contract.ID))

	// Return response
	return &models.ImportContractResponse{
		Contract: *contract,
	}, nil
}

// GetContract retrieves a contract by ID
func (s *Service) GetContract(ctx context.Context, req *models.GetContractRequest) (*models.GetContractResponse, error) {
	s.logger.Info(fmt.Sprintf("Getting contract %s", req.ID))

	// Get contract from database
	contract, err := s.repo.GetContract(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error(fmt.Sprintf("Contract not found: %s", req.ID))
			return nil, fmt.Errorf("contract not found: %s", req.ID)
		}
		s.logger.Error(fmt.Sprintf("Failed to get contract: %v", err))
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Contract retrieved successfully: %s", contract.ID))

	// Return response
	return &models.GetContractResponse{
		Contract: *contract,
	}, nil
}

// GetContractByAddress retrieves a contract by address
func (s *Service) GetContractByAddress(ctx context.Context, req *models.GetContractByAddressRequest) (*models.GetContractByAddressResponse, error) {
	s.logger.Info(fmt.Sprintf("Getting contract by address %s on chain %s", req.Address, req.Chain))

	// Get contract from database
	contract, err := s.repo.GetContractByAddress(ctx, req.Address, string(req.Chain))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error(fmt.Sprintf("Contract not found: %s", req.Address))
			return nil, fmt.Errorf("contract not found: %s", req.Address)
		}
		s.logger.Error(fmt.Sprintf("Failed to get contract: %v", err))
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Contract retrieved successfully: %s", contract.ID))

	// Return response
	return &models.GetContractByAddressResponse{
		Contract: *contract,
	}, nil
}

// ListContracts lists all contracts for a user
func (s *Service) ListContracts(ctx context.Context, req *models.ListContractsRequest) (*models.ListContractsResponse, error) {
	s.logger.Info(fmt.Sprintf("Listing contracts for user %s", req.UserID))

	// Get contracts from database
	contracts, total, err := s.repo.ListContracts(ctx, req.UserID, string(req.Chain), string(req.Type), req.Limit, req.Offset)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list contracts: %v", err))
		return nil, fmt.Errorf("failed to list contracts: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Listed %d contracts for user %s", len(contracts), req.UserID))

	// Return response
	return &models.ListContractsResponse{
		Contracts: contracts,
		Total:     total,
	}, nil
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
