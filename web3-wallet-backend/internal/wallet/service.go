package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/yourusername/web3-wallet-backend/pkg/blockchain"
	"github.com/yourusername/web3-wallet-backend/pkg/crypto"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
	"github.com/yourusername/web3-wallet-backend/pkg/models"
)

// Service provides wallet operations
type Service struct {
	repo           Repository
	ethClient      *blockchain.EthereumClient
	bscClient      *blockchain.EthereumClient
	polygonClient  *blockchain.EthereumClient
	keyManager     *crypto.KeyManager
	logger         *logger.Logger
	keystorePath   string
}

// NewService creates a new wallet service
func NewService(
	repo Repository,
	ethClient *blockchain.EthereumClient,
	bscClient *blockchain.EthereumClient,
	polygonClient *blockchain.EthereumClient,
	keyManager *crypto.KeyManager,
	logger *logger.Logger,
	keystorePath string,
) *Service {
	return &Service{
		repo:          repo,
		ethClient:     ethClient,
		bscClient:     bscClient,
		polygonClient: polygonClient,
		keyManager:    keyManager,
		logger:        logger.Named("wallet-service"),
		keystorePath:  keystorePath,
	}
}

// CreateWallet creates a new wallet
func (s *Service) CreateWallet(ctx context.Context, req *models.CreateWalletRequest) (*models.CreateWalletResponse, error) {
	s.logger.Info(fmt.Sprintf("Creating wallet for user %s on chain %s", req.UserID, req.Chain))

	// Generate key pair
	privateKey, publicKey, address, err := s.keyManager.GenerateKeyPair()
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to generate key pair: %v", err))
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Generate mnemonic
	mnemonic, err := s.keyManager.GenerateMnemonic()
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to generate mnemonic: %v", err))
		return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	// Create wallet
	wallet := &models.Wallet{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Name:      req.Name,
		Address:   address,
		Chain:     string(req.Chain),
		Type:      string(req.Type),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save wallet to database
	if err := s.repo.CreateWallet(ctx, wallet); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save wallet: %v", err))
		return nil, fmt.Errorf("failed to save wallet: %w", err)
	}

	// Encrypt private key
	encryptedKey, err := s.keyManager.EncryptPrivateKey(privateKey, "temporary-passphrase")
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to encrypt private key: %v", err))
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Save encrypted key to keystore
	if err := s.repo.SaveKeystore(ctx, wallet.ID, encryptedKey); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save keystore: %v", err))
		return nil, fmt.Errorf("failed to save keystore: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Wallet created successfully: %s", wallet.ID))

	// Return response
	return &models.CreateWalletResponse{
		Wallet:         *wallet,
		Mnemonic:       mnemonic,
		PrivateKey:     privateKey,
		DerivationPath: "m/44'/60'/0'/0/0", // Default Ethereum derivation path
	}, nil
}

// GetWallet retrieves a wallet by ID
func (s *Service) GetWallet(ctx context.Context, req *models.GetWalletRequest) (*models.GetWalletResponse, error) {
	s.logger.Info(fmt.Sprintf("Getting wallet %s", req.ID))

	// Get wallet from database
	wallet, err := s.repo.GetWallet(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error(fmt.Sprintf("Wallet not found: %s", req.ID))
			return nil, fmt.Errorf("wallet not found: %s", req.ID)
		}
		s.logger.Error(fmt.Sprintf("Failed to get wallet: %v", err))
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Wallet retrieved successfully: %s", wallet.ID))

	// Return response
	return &models.GetWalletResponse{
		Wallet: *wallet,
	}, nil
}

// ListWallets lists all wallets for a user
func (s *Service) ListWallets(ctx context.Context, req *models.ListWalletsRequest) (*models.ListWalletsResponse, error) {
	s.logger.Info(fmt.Sprintf("Listing wallets for user %s", req.UserID))

	// Get wallets from database
	wallets, total, err := s.repo.ListWallets(ctx, req.UserID, string(req.Chain), string(req.Type), req.Limit, req.Offset)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list wallets: %v", err))
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Listed %d wallets for user %s", len(wallets), req.UserID))

	// Return response
	return &models.ListWalletsResponse{
		Wallets: wallets,
		Total:   total,
	}, nil
}

// GetBalance retrieves the balance of a wallet
func (s *Service) GetBalance(ctx context.Context, req *models.GetBalanceRequest) (*models.GetBalanceResponse, error) {
	s.logger.Info(fmt.Sprintf("Getting balance for wallet %s", req.WalletID))

	// Get wallet from database
	wallet, err := s.repo.GetWallet(ctx, req.WalletID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error(fmt.Sprintf("Wallet not found: %s", req.WalletID))
			return nil, fmt.Errorf("wallet not found: %s", req.WalletID)
		}
		s.logger.Error(fmt.Sprintf("Failed to get wallet: %v", err))
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Get blockchain client based on chain
	var client *blockchain.EthereumClient
	var symbol string
	var decimals int

	switch models.Chain(wallet.Chain) {
	case models.ChainEthereum:
		client = s.ethClient
		symbol = "ETH"
		decimals = 18
	case models.ChainBSC:
		client = s.bscClient
		symbol = "BNB"
		decimals = 18
	case models.ChainPolygon:
		client = s.polygonClient
		symbol = "MATIC"
		decimals = 18
	default:
		s.logger.Error(fmt.Sprintf("Unsupported chain: %s", wallet.Chain))
		return nil, fmt.Errorf("unsupported chain: %s", wallet.Chain)
	}

	// Get balance
	var balance *big.Int
	var balanceErr error

	if req.TokenAddress == "" {
		// Get native token balance
		balance, balanceErr = client.GetBalance(ctx, wallet.Address)
	} else {
		// Get ERC-20 token balance
		// TODO: Implement ERC-20 token balance retrieval
		balance = big.NewInt(0)
		symbol = "TOKEN"
		decimals = 18
	}

	if balanceErr != nil {
		s.logger.Error(fmt.Sprintf("Failed to get balance: %v", balanceErr))
		return nil, fmt.Errorf("failed to get balance: %w", balanceErr)
	}

	s.logger.Info(fmt.Sprintf("Balance retrieved successfully for wallet %s: %s", wallet.ID, balance.String()))

	// Return response
	return &models.GetBalanceResponse{
		Balance:      balance.String(),
		Symbol:       symbol,
		Decimals:     decimals,
		TokenAddress: req.TokenAddress,
	}, nil
}

// ImportWallet imports an existing wallet
func (s *Service) ImportWallet(ctx context.Context, req *models.ImportWalletRequest) (*models.ImportWalletResponse, error) {
	s.logger.Info(fmt.Sprintf("Importing wallet for user %s on chain %s", req.UserID, req.Chain))

	// Import private key
	address, err := s.keyManager.ImportPrivateKey(req.PrivateKey, "temporary-passphrase")
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to import private key: %v", err))
		return nil, fmt.Errorf("failed to import private key: %w", err)
	}

	// Create wallet
	wallet := &models.Wallet{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Name:      req.Name,
		Address:   address,
		Chain:     string(req.Chain),
		Type:      string(models.WalletTypeImported),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save wallet to database
	if err := s.repo.CreateWallet(ctx, wallet); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save wallet: %v", err))
		return nil, fmt.Errorf("failed to save wallet: %w", err)
	}

	// Encrypt private key
	encryptedKey, err := s.keyManager.EncryptPrivateKey(req.PrivateKey, "temporary-passphrase")
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to encrypt private key: %v", err))
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Save encrypted key to keystore
	if err := s.repo.SaveKeystore(ctx, wallet.ID, encryptedKey); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save keystore: %v", err))
		return nil, fmt.Errorf("failed to save keystore: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Wallet imported successfully: %s", wallet.ID))

	// Return response
	return &models.ImportWalletResponse{
		Wallet: *wallet,
	}, nil
}

// ExportWallet exports a wallet (private key or keystore)
func (s *Service) ExportWallet(ctx context.Context, req *models.ExportWalletRequest) (*models.ExportWalletResponse, error) {
	s.logger.Info(fmt.Sprintf("Exporting wallet %s", req.WalletID))

	// Get wallet from database
	wallet, err := s.repo.GetWallet(ctx, req.WalletID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error(fmt.Sprintf("Wallet not found: %s", req.WalletID))
			return nil, fmt.Errorf("wallet not found: %s", req.WalletID)
		}
		s.logger.Error(fmt.Sprintf("Failed to get wallet: %v", err))
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Get keystore
	keystore, err := s.repo.GetKeystore(ctx, wallet.ID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get keystore: %v", err))
		return nil, fmt.Errorf("failed to get keystore: %w", err)
	}

	// Decrypt private key
	privateKey, err := s.keyManager.DecryptPrivateKey(keystore, "temporary-passphrase")
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decrypt private key: %v", err))
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Wallet exported successfully: %s", wallet.ID))

	// Return response
	return &models.ExportWalletResponse{
		PrivateKey: privateKey,
		Keystore:   keystore,
	}, nil
}

// DeleteWallet deletes a wallet
func (s *Service) DeleteWallet(ctx context.Context, req *models.DeleteWalletRequest) (*models.DeleteWalletResponse, error) {
	s.logger.Info(fmt.Sprintf("Deleting wallet %s", req.WalletID))

	// Delete wallet from database
	if err := s.repo.DeleteWallet(ctx, req.WalletID); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to delete wallet: %v", err))
		return nil, fmt.Errorf("failed to delete wallet: %w", err)
	}

	// Delete keystore
	if err := s.repo.DeleteKeystore(ctx, req.WalletID); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to delete keystore: %v", err))
		return nil, fmt.Errorf("failed to delete keystore: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Wallet deleted successfully: %s", req.WalletID))

	// Return response
	return &models.DeleteWalletResponse{
		Success: true,
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
