package wallet

import (
	"context"
	"fmt"
	"regexp"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// LedgerWallet implements hardware wallet interface for Ledger devices
type LedgerWallet struct {
	logger   *logger.Logger
	deviceID string
	config   HardwareWalletConfig

	// Connection state
	connected  bool
	deviceInfo *DeviceInfo

	// Ledger-specific fields
	transport LedgerTransport
	appName   string
}

// LedgerTransport represents the transport layer for Ledger communication
type LedgerTransport interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	SendAPDU(command []byte) ([]byte, error)
	GetDeviceInfo() (*DeviceInfo, error)
}

// LedgerUSBTransport implements USB transport for Ledger devices
type LedgerUSBTransport struct {
	logger    *logger.Logger
	deviceID  string
	connected bool
}

// NewLedgerWallet creates a new Ledger wallet instance
func NewLedgerWallet(logger *logger.Logger, deviceID string, config HardwareWalletConfig) (*LedgerWallet, error) {
	transport := &LedgerUSBTransport{
		logger:   logger.Named("ledger-usb"),
		deviceID: deviceID,
	}

	return &LedgerWallet{
		logger:    logger.Named("ledger-wallet"),
		deviceID:  deviceID,
		config:    config,
		transport: transport,
		appName:   "Ethereum",
	}, nil
}

// Connect connects to the Ledger device
func (lw *LedgerWallet) Connect(ctx context.Context) error {
	lw.logger.Info("Connecting to Ledger device", zap.String("device_id", lw.deviceID))

	// Connect transport
	if err := lw.transport.Connect(); err != nil {
		return fmt.Errorf("failed to connect transport: %w", err)
	}

	// Get device info
	deviceInfo, err := lw.transport.GetDeviceInfo()
	if err != nil {
		lw.transport.Disconnect()
		return fmt.Errorf("failed to get device info: %w", err)
	}

	lw.deviceInfo = deviceInfo
	lw.connected = true

	// Verify Ethereum app is available
	if err := lw.verifyEthereumApp(ctx); err != nil {
		lw.logger.Warn("Ethereum app verification failed", zap.Error(err))
		// Continue anyway - user might need to open the app manually
	}

	lw.logger.Info("Successfully connected to Ledger device",
		zap.String("model", deviceInfo.Model),
		zap.String("firmware", deviceInfo.FirmwareVersion))

	return nil
}

// Disconnect disconnects from the Ledger device
func (lw *LedgerWallet) Disconnect() error {
	lw.logger.Info("Disconnecting from Ledger device", zap.String("device_id", lw.deviceID))

	if err := lw.transport.Disconnect(); err != nil {
		return fmt.Errorf("failed to disconnect transport: %w", err)
	}

	lw.connected = false
	lw.deviceInfo = nil

	lw.logger.Info("Successfully disconnected from Ledger device")
	return nil
}

// IsConnected returns whether the wallet is connected
func (lw *LedgerWallet) IsConnected() bool {
	return lw.connected && lw.transport.IsConnected()
}

// GetDeviceInfo returns device information
func (lw *LedgerWallet) GetDeviceInfo() (*DeviceInfo, error) {
	if !lw.connected {
		return nil, fmt.Errorf("device not connected")
	}
	return lw.deviceInfo, nil
}

// GetAddress gets an address from the Ledger device
func (lw *LedgerWallet) GetAddress(ctx context.Context, derivationPath string) (string, error) {
	lw.logger.Debug("Getting address from Ledger",
		zap.String("derivation_path", derivationPath))

	if !lw.connected {
		return "", fmt.Errorf("device not connected")
	}

	if err := lw.ValidateDerivationPath(derivationPath); err != nil {
		return "", fmt.Errorf("invalid derivation path: %w", err)
	}

	// Build APDU command for getting address
	command, err := lw.buildGetAddressCommand(derivationPath, false)
	if err != nil {
		return "", fmt.Errorf("failed to build command: %w", err)
	}

	// Send command to device
	response, err := lw.transport.SendAPDU(command)
	if err != nil {
		return "", fmt.Errorf("failed to send APDU: %w", err)
	}

	// Parse response
	address, err := lw.parseAddressResponse(response)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	lw.logger.Debug("Successfully retrieved address from Ledger",
		zap.String("address", address))

	return address, nil
}

// GetAddresses gets multiple addresses from the Ledger device
func (lw *LedgerWallet) GetAddresses(ctx context.Context, derivationPaths []string) ([]string, error) {
	var addresses []string

	for _, path := range derivationPaths {
		address, err := lw.GetAddress(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("failed to get address for path %s: %w", path, err)
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetPublicKey gets a public key from the Ledger device
func (lw *LedgerWallet) GetPublicKey(ctx context.Context, derivationPath string) ([]byte, error) {
	lw.logger.Debug("Getting public key from Ledger",
		zap.String("derivation_path", derivationPath))

	if !lw.connected {
		return nil, fmt.Errorf("device not connected")
	}

	if err := lw.ValidateDerivationPath(derivationPath); err != nil {
		return nil, fmt.Errorf("invalid derivation path: %w", err)
	}

	// Build APDU command for getting public key
	command, err := lw.buildGetPublicKeyCommand(derivationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build command: %w", err)
	}

	// Send command to device
	response, err := lw.transport.SendAPDU(command)
	if err != nil {
		return nil, fmt.Errorf("failed to send APDU: %w", err)
	}

	// Parse response
	publicKey, err := lw.parsePublicKeyResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return publicKey, nil
}

// SignTransaction signs a transaction with the Ledger device
func (lw *LedgerWallet) SignTransaction(ctx context.Context, tx *types.Transaction, derivationPath string) (*types.Transaction, error) {
	lw.logger.Info("Signing transaction with Ledger",
		zap.String("derivation_path", derivationPath),
		zap.String("tx_hash", tx.Hash().Hex()))

	if !lw.connected {
		return nil, fmt.Errorf("device not connected")
	}

	if err := lw.ValidateDerivationPath(derivationPath); err != nil {
		return nil, fmt.Errorf("invalid derivation path: %w", err)
	}

	// Serialize transaction for signing
	txData, err := lw.serializeTransactionForSigning(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build APDU command for signing
	command, err := lw.buildSignTransactionCommand(derivationPath, txData)
	if err != nil {
		return nil, fmt.Errorf("failed to build command: %w", err)
	}

	// Send command to device (this will prompt user for confirmation)
	response, err := lw.transport.SendAPDU(command)
	if err != nil {
		return nil, fmt.Errorf("failed to send APDU: %w", err)
	}

	// Parse signature response
	signature, err := lw.parseSignatureResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature: %w", err)
	}

	// Apply signature to transaction
	signedTx, err := lw.applySignatureToTransaction(tx, signature)
	if err != nil {
		return nil, fmt.Errorf("failed to apply signature: %w", err)
	}

	lw.logger.Info("Successfully signed transaction with Ledger",
		zap.String("signed_tx_hash", signedTx.Hash().Hex()))

	return signedTx, nil
}

// SignMessage signs a message with the Ledger device
func (lw *LedgerWallet) SignMessage(ctx context.Context, message []byte, derivationPath string) ([]byte, error) {
	lw.logger.Info("Signing message with Ledger",
		zap.String("derivation_path", derivationPath),
		zap.Int("message_length", len(message)))

	if !lw.connected {
		return nil, fmt.Errorf("device not connected")
	}

	// Build APDU command for message signing
	command, err := lw.buildSignMessageCommand(derivationPath, message)
	if err != nil {
		return nil, fmt.Errorf("failed to build command: %w", err)
	}

	// Send command to device
	response, err := lw.transport.SendAPDU(command)
	if err != nil {
		return nil, fmt.Errorf("failed to send APDU: %w", err)
	}

	// Parse signature response
	signature, err := lw.parseSignatureResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature: %w", err)
	}

	return signature, nil
}

// SignTypedData signs typed data with the Ledger device
func (lw *LedgerWallet) SignTypedData(ctx context.Context, typedData []byte, derivationPath string) ([]byte, error) {
	lw.logger.Info("Signing typed data with Ledger",
		zap.String("derivation_path", derivationPath),
		zap.Int("data_length", len(typedData)))

	// For now, treat typed data similar to message signing
	// In a real implementation, this would use EIP-712 specific commands
	return lw.SignMessage(ctx, typedData, derivationPath)
}

// GetSupportedChains returns supported blockchain networks
func (lw *LedgerWallet) GetSupportedChains() []string {
	return []string{"ethereum", "polygon", "bsc", "arbitrum", "optimism"}
}

// ValidateDerivationPath validates a BIP44 derivation path
func (lw *LedgerWallet) ValidateDerivationPath(path string) error {
	// BIP44 format: m/44'/coin_type'/account'/change/address_index
	pattern := `^m/44'/\d+'/\d+'/[01]/\d+$`
	matched, err := regexp.MatchString(pattern, path)
	if err != nil {
		return fmt.Errorf("regex error: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid BIP44 derivation path format: %s", path)
	}
	return nil
}

// GetWalletType returns the wallet type
func (lw *LedgerWallet) GetWalletType() HardwareWalletType {
	return HardwareWalletTypeLedger
}

// Helper methods

// verifyEthereumApp verifies that the Ethereum app is available
func (lw *LedgerWallet) verifyEthereumApp(ctx context.Context) error {
	// In a real implementation, this would check if the Ethereum app is open
	// For now, just simulate success
	return nil
}

// buildGetAddressCommand builds APDU command for getting address
func (lw *LedgerWallet) buildGetAddressCommand(derivationPath string, display bool) ([]byte, error) {
	// Simplified APDU command building
	// In a real implementation, this would properly encode the derivation path
	command := []byte{0xE0, 0x02, 0x00, 0x00} // CLA, INS, P1, P2

	if display {
		command[2] = 0x01 // P1 = 1 for display
	}

	// Add derivation path (simplified)
	pathBytes := []byte(derivationPath)
	command = append(command, byte(len(pathBytes)))
	command = append(command, pathBytes...)

	return command, nil
}

// buildGetPublicKeyCommand builds APDU command for getting public key
func (lw *LedgerWallet) buildGetPublicKeyCommand(derivationPath string) ([]byte, error) {
	// Simplified implementation
	command := []byte{0xE0, 0x02, 0x01, 0x00} // CLA, INS, P1, P2

	pathBytes := []byte(derivationPath)
	command = append(command, byte(len(pathBytes)))
	command = append(command, pathBytes...)

	return command, nil
}

// buildSignTransactionCommand builds APDU command for signing transaction
func (lw *LedgerWallet) buildSignTransactionCommand(derivationPath string, txData []byte) ([]byte, error) {
	// Simplified implementation
	command := []byte{0xE0, 0x04, 0x00, 0x00} // CLA, INS, P1, P2

	// Add derivation path
	pathBytes := []byte(derivationPath)
	command = append(command, byte(len(pathBytes)))
	command = append(command, pathBytes...)

	// Add transaction data
	command = append(command, txData...)

	return command, nil
}

// buildSignMessageCommand builds APDU command for signing message
func (lw *LedgerWallet) buildSignMessageCommand(derivationPath string, message []byte) ([]byte, error) {
	// Simplified implementation
	command := []byte{0xE0, 0x08, 0x00, 0x00} // CLA, INS, P1, P2

	pathBytes := []byte(derivationPath)
	command = append(command, byte(len(pathBytes)))
	command = append(command, pathBytes...)
	command = append(command, message...)

	return command, nil
}

// serializeTransactionForSigning serializes transaction for Ledger signing
func (lw *LedgerWallet) serializeTransactionForSigning(tx *types.Transaction) ([]byte, error) {
	// In a real implementation, this would properly serialize the transaction
	// according to Ledger's expected format
	return tx.Hash().Bytes(), nil
}

// parseAddressResponse parses address from Ledger response
func (lw *LedgerWallet) parseAddressResponse(response []byte) (string, error) {
	// Simplified parsing - in reality would parse the actual APDU response
	if len(response) < 20 {
		return "", fmt.Errorf("invalid response length")
	}

	// Generate a mock address for demonstration
	address := common.BytesToAddress(response[:20])
	return address.Hex(), nil
}

// parsePublicKeyResponse parses public key from Ledger response
func (lw *LedgerWallet) parsePublicKeyResponse(response []byte) ([]byte, error) {
	// Simplified parsing
	if len(response) < 65 {
		return nil, fmt.Errorf("invalid response length")
	}

	return response[:65], nil
}

// parseSignatureResponse parses signature from Ledger response
func (lw *LedgerWallet) parseSignatureResponse(response []byte) ([]byte, error) {
	// Simplified parsing
	if len(response) < 65 {
		return nil, fmt.Errorf("invalid response length")
	}

	return response[:65], nil
}

// applySignatureToTransaction applies signature to transaction
func (lw *LedgerWallet) applySignatureToTransaction(tx *types.Transaction, signature []byte) (*types.Transaction, error) {
	// In a real implementation, this would properly apply the signature
	// For now, return the original transaction
	return tx, nil
}

// USB Transport Implementation

// Connect connects the USB transport
func (lut *LedgerUSBTransport) Connect() error {
	lut.logger.Info("Connecting Ledger USB transport", zap.String("device_id", lut.deviceID))

	// In a real implementation, this would establish USB connection
	lut.connected = true
	return nil
}

// Disconnect disconnects the USB transport
func (lut *LedgerUSBTransport) Disconnect() error {
	lut.logger.Info("Disconnecting Ledger USB transport")

	lut.connected = false
	return nil
}

// IsConnected returns connection status
func (lut *LedgerUSBTransport) IsConnected() bool {
	return lut.connected
}

// SendAPDU sends an APDU command to the device
func (lut *LedgerUSBTransport) SendAPDU(command []byte) ([]byte, error) {
	if !lut.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	lut.logger.Debug("Sending APDU command", zap.Int("command_length", len(command)))

	// Simulate APDU response
	response := make([]byte, 65)
	for i := range response {
		response[i] = byte(i)
	}

	// Add success status word
	response = append(response, 0x90, 0x00)

	return response, nil
}

// GetDeviceInfo gets device information
func (lut *LedgerUSBTransport) GetDeviceInfo() (*DeviceInfo, error) {
	return &DeviceInfo{
		Type:            HardwareWalletTypeLedger,
		Model:           "Nano S Plus",
		SerialNumber:    lut.deviceID,
		FirmwareVersion: "1.0.3",
		IsLocked:        false,
		IsInitialized:   true,
		Label:           "My Ledger",
		SupportedApps:   []string{"Ethereum", "Bitcoin"},
	}, nil
}
