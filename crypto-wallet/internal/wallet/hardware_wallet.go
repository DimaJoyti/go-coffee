package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// HardwareWalletType represents the type of hardware wallet
type HardwareWalletType string

const (
	HardwareWalletTypeLedger HardwareWalletType = "ledger"
	HardwareWalletTypeTrezor HardwareWalletType = "trezor"
)

// HardwareWalletInterface defines the interface for hardware wallet operations
type HardwareWalletInterface interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool
	GetDeviceInfo() (*DeviceInfo, error)

	// Address derivation
	GetAddress(ctx context.Context, derivationPath string) (string, error)
	GetAddresses(ctx context.Context, derivationPaths []string) ([]string, error)
	GetPublicKey(ctx context.Context, derivationPath string) ([]byte, error)

	// Transaction signing
	SignTransaction(ctx context.Context, tx *types.Transaction, derivationPath string) (*types.Transaction, error)
	SignMessage(ctx context.Context, message []byte, derivationPath string) ([]byte, error)
	SignTypedData(ctx context.Context, typedData []byte, derivationPath string) ([]byte, error)

	// Wallet management
	GetSupportedChains() []string
	ValidateDerivationPath(path string) error
	GetWalletType() HardwareWalletType
}

// DeviceInfo represents information about a hardware wallet device
type DeviceInfo struct {
	Type         HardwareWalletType `json:"type"`
	Model        string             `json:"model"`
	SerialNumber string             `json:"serial_number"`
	FirmwareVersion string          `json:"firmware_version"`
	IsLocked     bool               `json:"is_locked"`
	IsInitialized bool              `json:"is_initialized"`
	Label        string             `json:"label"`
	SupportedApps []string          `json:"supported_apps"`
}

// HardwareWalletManager manages multiple hardware wallets
type HardwareWalletManager struct {
	logger *logger.Logger
	
	// Connected wallets
	connectedWallets map[string]HardwareWalletInterface
	
	// Configuration
	config HardwareWalletConfig
}

// HardwareWalletConfig holds configuration for hardware wallet management
type HardwareWalletConfig struct {
	EnabledTypes        []HardwareWalletType `json:"enabled_types" yaml:"enabled_types"`
	ConnectionTimeout   time.Duration        `json:"connection_timeout" yaml:"connection_timeout"`
	SigningTimeout      time.Duration        `json:"signing_timeout" yaml:"signing_timeout"`
	AutoReconnect       bool                 `json:"auto_reconnect" yaml:"auto_reconnect"`
	MaxRetries          int                  `json:"max_retries" yaml:"max_retries"`
	DefaultDerivationPath string             `json:"default_derivation_path" yaml:"default_derivation_path"`
}

// NewHardwareWalletManager creates a new hardware wallet manager
func NewHardwareWalletManager(logger *logger.Logger, config HardwareWalletConfig) *HardwareWalletManager {
	return &HardwareWalletManager{
		logger:           logger.Named("hardware-wallet-manager"),
		connectedWallets: make(map[string]HardwareWalletInterface),
		config:           config,
	}
}

// DiscoverWallets discovers available hardware wallets
func (hwm *HardwareWalletManager) DiscoverWallets(ctx context.Context) ([]*DeviceInfo, error) {
	hwm.logger.Info("Discovering hardware wallets")

	var devices []*DeviceInfo

	// Discover Ledger devices
	if hwm.isTypeEnabled(HardwareWalletTypeLedger) {
		ledgerDevices, err := hwm.discoverLedgerDevices(ctx)
		if err != nil {
			hwm.logger.Warn("Failed to discover Ledger devices", zap.Error(err))
		} else {
			devices = append(devices, ledgerDevices...)
		}
	}

	// Discover Trezor devices
	if hwm.isTypeEnabled(HardwareWalletTypeTrezor) {
		trezorDevices, err := hwm.discoverTrezorDevices(ctx)
		if err != nil {
			hwm.logger.Warn("Failed to discover Trezor devices", zap.Error(err))
		} else {
			devices = append(devices, trezorDevices...)
		}
	}

	hwm.logger.Info("Hardware wallet discovery completed",
		zap.Int("devices_found", len(devices)))

	return devices, nil
}

// ConnectWallet connects to a specific hardware wallet
func (hwm *HardwareWalletManager) ConnectWallet(ctx context.Context, deviceID string, walletType HardwareWalletType) (HardwareWalletInterface, error) {
	hwm.logger.Info("Connecting to hardware wallet",
		zap.String("device_id", deviceID),
		zap.String("type", string(walletType)))

	// Check if already connected
	if wallet, exists := hwm.connectedWallets[deviceID]; exists {
		if wallet.IsConnected() {
			return wallet, nil
		}
		// Remove stale connection
		delete(hwm.connectedWallets, deviceID)
	}

	// Create wallet instance
	var wallet HardwareWalletInterface
	var err error

	switch walletType {
	case HardwareWalletTypeLedger:
		wallet, err = NewLedgerWallet(hwm.logger, deviceID, hwm.config)
	case HardwareWalletTypeTrezor:
		wallet, err = NewTrezorWallet(hwm.logger, deviceID, hwm.config)
	default:
		return nil, fmt.Errorf("unsupported wallet type: %s", walletType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create wallet instance: %w", err)
	}

	// Connect to the wallet
	connectCtx, cancel := context.WithTimeout(ctx, hwm.config.ConnectionTimeout)
	defer cancel()

	if err := wallet.Connect(connectCtx); err != nil {
		return nil, fmt.Errorf("failed to connect to wallet: %w", err)
	}

	// Store connected wallet
	hwm.connectedWallets[deviceID] = wallet

	hwm.logger.Info("Successfully connected to hardware wallet",
		zap.String("device_id", deviceID),
		zap.String("type", string(walletType)))

	return wallet, nil
}

// DisconnectWallet disconnects from a hardware wallet
func (hwm *HardwareWalletManager) DisconnectWallet(deviceID string) error {
	hwm.logger.Info("Disconnecting hardware wallet", zap.String("device_id", deviceID))

	wallet, exists := hwm.connectedWallets[deviceID]
	if !exists {
		return fmt.Errorf("wallet not connected: %s", deviceID)
	}

	if err := wallet.Disconnect(); err != nil {
		hwm.logger.Error("Failed to disconnect wallet", 
			zap.String("device_id", deviceID),
			zap.Error(err))
		return err
	}

	delete(hwm.connectedWallets, deviceID)

	hwm.logger.Info("Successfully disconnected hardware wallet",
		zap.String("device_id", deviceID))

	return nil
}

// GetConnectedWallets returns all connected hardware wallets
func (hwm *HardwareWalletManager) GetConnectedWallets() map[string]HardwareWalletInterface {
	connected := make(map[string]HardwareWalletInterface)
	
	for deviceID, wallet := range hwm.connectedWallets {
		if wallet.IsConnected() {
			connected[deviceID] = wallet
		} else {
			// Clean up disconnected wallets
			delete(hwm.connectedWallets, deviceID)
		}
	}
	
	return connected
}

// SignTransactionWithHardwareWallet signs a transaction using a hardware wallet
func (hwm *HardwareWalletManager) SignTransactionWithHardwareWallet(
	ctx context.Context,
	deviceID string,
	tx *types.Transaction,
	derivationPath string,
) (*types.Transaction, error) {
	hwm.logger.Info("Signing transaction with hardware wallet",
		zap.String("device_id", deviceID),
		zap.String("derivation_path", derivationPath),
		zap.String("tx_hash", tx.Hash().Hex()))

	wallet, exists := hwm.connectedWallets[deviceID]
	if !exists {
		return nil, fmt.Errorf("wallet not connected: %s", deviceID)
	}

	if !wallet.IsConnected() {
		return nil, fmt.Errorf("wallet is not connected: %s", deviceID)
	}

	// Sign transaction with timeout
	signCtx, cancel := context.WithTimeout(ctx, hwm.config.SigningTimeout)
	defer cancel()

	signedTx, err := wallet.SignTransaction(signCtx, tx, derivationPath)
	if err != nil {
		hwm.logger.Error("Failed to sign transaction",
			zap.String("device_id", deviceID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	hwm.logger.Info("Successfully signed transaction with hardware wallet",
		zap.String("device_id", deviceID),
		zap.String("signed_tx_hash", signedTx.Hash().Hex()))

	return signedTx, nil
}

// GetAddressFromHardwareWallet gets an address from a hardware wallet
func (hwm *HardwareWalletManager) GetAddressFromHardwareWallet(
	ctx context.Context,
	deviceID string,
	derivationPath string,
) (string, error) {
	hwm.logger.Debug("Getting address from hardware wallet",
		zap.String("device_id", deviceID),
		zap.String("derivation_path", derivationPath))

	wallet, exists := hwm.connectedWallets[deviceID]
	if !exists {
		return "", fmt.Errorf("wallet not connected: %s", deviceID)
	}

	if !wallet.IsConnected() {
		return "", fmt.Errorf("wallet is not connected: %s", deviceID)
	}

	address, err := wallet.GetAddress(ctx, derivationPath)
	if err != nil {
		return "", fmt.Errorf("failed to get address: %w", err)
	}

	return address, nil
}

// ValidateDerivationPath validates a derivation path
func (hwm *HardwareWalletManager) ValidateDerivationPath(path string) error {
	// Basic validation for BIP44 derivation paths
	// Format: m/44'/coin_type'/account'/change/address_index
	
	if path == "" {
		return fmt.Errorf("derivation path cannot be empty")
	}

	if path[0] != 'm' {
		return fmt.Errorf("derivation path must start with 'm'")
	}

	// More detailed validation would be implemented here
	return nil
}

// Helper methods

// isTypeEnabled checks if a wallet type is enabled
func (hwm *HardwareWalletManager) isTypeEnabled(walletType HardwareWalletType) bool {
	for _, enabledType := range hwm.config.EnabledTypes {
		if enabledType == walletType {
			return true
		}
	}
	return false
}

// discoverLedgerDevices discovers Ledger devices
func (hwm *HardwareWalletManager) discoverLedgerDevices(ctx context.Context) ([]*DeviceInfo, error) {
	// In a real implementation, this would use the Ledger SDK to discover devices
	// For now, return a simulated device if Ledger is enabled
	
	devices := []*DeviceInfo{
		{
			Type:            HardwareWalletTypeLedger,
			Model:           "Nano S Plus",
			SerialNumber:    "0001",
			FirmwareVersion: "1.0.3",
			IsLocked:        false,
			IsInitialized:   true,
			Label:           "My Ledger",
			SupportedApps:   []string{"Ethereum", "Bitcoin"},
		},
	}

	return devices, nil
}

// discoverTrezorDevices discovers Trezor devices
func (hwm *HardwareWalletManager) discoverTrezorDevices(ctx context.Context) ([]*DeviceInfo, error) {
	// In a real implementation, this would use the Trezor SDK to discover devices
	// For now, return a simulated device if Trezor is enabled
	
	devices := []*DeviceInfo{
		{
			Type:            HardwareWalletTypeTrezor,
			Model:           "Model T",
			SerialNumber:    "0002",
			FirmwareVersion: "2.5.3",
			IsLocked:        false,
			IsInitialized:   true,
			Label:           "My Trezor",
			SupportedApps:   []string{"Ethereum", "Bitcoin"},
		},
	}

	return devices, nil
}

// GetDefaultConfig returns default hardware wallet configuration
func GetDefaultConfig() HardwareWalletConfig {
	return HardwareWalletConfig{
		EnabledTypes:          []HardwareWalletType{HardwareWalletTypeLedger, HardwareWalletTypeTrezor},
		ConnectionTimeout:     30 * time.Second,
		SigningTimeout:        60 * time.Second,
		AutoReconnect:         true,
		MaxRetries:            3,
		DefaultDerivationPath: "m/44'/60'/0'/0/0", // Ethereum default
	}
}
