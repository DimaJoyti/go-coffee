package payment

import (
	"context"
	"fmt"

	"github.com/DimaJoyti/go-coffee/pkg/bitcoin"
	"github.com/DimaJoyti/go-coffee/pkg/bitcoin/ecc"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/ethereum"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/models"
)

// Service handles payment operations
type Service struct {
	config   *config.Config
	logger   *logger.Logger
	bitcoin  *bitcoin.BitcoinUtils
	ethereum *ethereum.EthereumUtils
}

// NewService creates a new payment service
func NewService(cfg *config.Config, log *logger.Logger) (*Service, error) {
	bitcoinUtils := bitcoin.NewBitcoinUtils()

	return &Service{
		config:  cfg,
		logger:  log,
		bitcoin: bitcoinUtils,
	}, nil
}

// CreateWallet creates a new Bitcoin wallet
func (s *Service) CreateWallet(ctx context.Context, testnet bool) (*models.Wallet, error) {
	s.logger.Info("Creating new Bitcoin wallet", "testnet", testnet)

	wallet, err := bitcoin.NewWallet(testnet)
	if err != nil {
		s.logger.Error("Failed to create wallet", "error", err)
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	walletModel := &models.Wallet{
		Address:    wallet.GetAddress(),
		PrivateKey: wallet.GetPrivateKeyWIF(true),
		PublicKey:  fmt.Sprintf("%x", wallet.GetPublicKeyCompressed()),
		Network:    getNetworkString(testnet),
		Type:       "bitcoin",
	}

	s.logger.Info("Wallet created successfully", "address", walletModel.Address)
	return walletModel, nil
}

// ImportWallet imports a wallet from WIF private key
func (s *Service) ImportWallet(ctx context.Context, wif string) (*models.Wallet, error) {
	s.logger.Info("Importing Bitcoin wallet from WIF")

	wallet, err := bitcoin.NewWalletFromWIF(wif)
	if err != nil {
		s.logger.Error("Failed to import wallet", "error", err)
		return nil, fmt.Errorf("failed to import wallet: %w", err)
	}

	// Determine network from WIF
	_, _, testnet, err := s.bitcoin.WIFToPrivateKey(wif)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WIF: %w", err)
	}

	walletModel := &models.Wallet{
		Address:    wallet.GetAddress(),
		PrivateKey: wif,
		PublicKey:  fmt.Sprintf("%x", wallet.GetPublicKeyCompressed()),
		Network:    getNetworkString(testnet),
		Type:       "bitcoin",
	}

	s.logger.Info("Wallet imported successfully", "address", walletModel.Address)
	return walletModel, nil
}

// ValidateAddress validates a Bitcoin address
func (s *Service) ValidateAddress(ctx context.Context, address string) (*models.AddressValidation, error) {
	s.logger.Info("Validating Bitcoin address", "address", address)

	isValid := s.bitcoin.ValidateAddress(address)
	
	var addressType, network string
	if isValid {
		addressType, network, _ = s.bitcoin.GetAddressInfo(address)
	}

	validation := &models.AddressValidation{
		Address: address,
		Valid:   isValid,
		Type:    addressType,
		Network: network,
	}

	s.logger.Info("Address validation completed", "address", address, "valid", isValid)
	return validation, nil
}

// CreateMultisigAddress creates a multisig address
func (s *Service) CreateMultisigAddress(ctx context.Context, req *models.MultisigRequest) (*models.MultisigAddress, error) {
	s.logger.Info("Creating multisig address", "threshold", req.Threshold, "keys", len(req.PublicKeys))

	if req.Threshold < 1 || req.Threshold > len(req.PublicKeys) {
		return nil, fmt.Errorf("invalid threshold: %d (must be between 1 and %d)", req.Threshold, len(req.PublicKeys))
	}

	// Convert hex public keys to Point objects
	publicKeys, err := s.parsePublicKeys(req.PublicKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public keys: %w", err)
	}

	address, err := s.bitcoin.CreateMultisigAddress(publicKeys, req.Threshold, req.Testnet)
	if err != nil {
		s.logger.Error("Failed to create multisig address", "error", err)
		return nil, fmt.Errorf("failed to create multisig address: %w", err)
	}

	multisigAddr := &models.MultisigAddress{
		Address:    address,
		Threshold:  req.Threshold,
		PublicKeys: req.PublicKeys,
		Network:    getNetworkString(req.Testnet),
		Type:       "P2SH",
	}

	s.logger.Info("Multisig address created successfully", "address", address)
	return multisigAddr, nil
}

// SignMessage signs a message with a private key
func (s *Service) SignMessage(ctx context.Context, req *models.SignMessageRequest) (*models.SignMessageResponse, error) {
	s.logger.Info("Signing message", "message_length", len(req.Message))

	wallet, err := bitcoin.NewWalletFromWIF(req.PrivateKey)
	if err != nil {
		s.logger.Error("Failed to create wallet from WIF", "error", err)
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	signature, err := wallet.SignMessage([]byte(req.Message))
	if err != nil {
		s.logger.Error("Failed to sign message", "error", err)
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	response := &models.SignMessageResponse{
		Message:   req.Message,
		Signature: fmt.Sprintf("%x", signature.DER()),
		Address:   wallet.GetAddress(),
	}

	s.logger.Info("Message signed successfully")
	return response, nil
}

// VerifyMessage verifies a message signature
func (s *Service) VerifyMessage(ctx context.Context, req *models.VerifyMessageRequest) (*models.VerifyMessageResponse, error) {
	s.logger.Info("Verifying message signature", "address", req.Address)

	// This is a simplified implementation
	// In a real implementation, you would need to recover the public key from the signature
	// or have the public key provided separately

	response := &models.VerifyMessageResponse{
		Message: req.Message,
		Address: req.Address,
		Valid:   false, // Placeholder - would implement full verification
	}

	s.logger.Info("Message verification completed", "valid", response.Valid)
	return response, nil
}

// Helper functions

func getNetworkString(testnet bool) string {
	if testnet {
		return "testnet"
	}
	return "mainnet"
}

func (s *Service) parsePublicKeys(hexKeys []string) ([]*ecc.Point, error) {
	// This would need to be implemented to convert hex strings to Point objects
	// For now, returning an error as this requires more complex parsing
	return nil, fmt.Errorf("public key parsing not yet implemented")
}

// GetSupportedFeatures returns the supported Bitcoin features
func (s *Service) GetSupportedFeatures(ctx context.Context) []string {
	return bitcoin.GetSupportedFeatures()
}

// GetVersion returns the service version
func (s *Service) GetVersion(ctx context.Context) string {
	return bitcoin.GetVersion()
}
