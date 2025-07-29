package wallet

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHardwareWalletManager(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	manager := NewHardwareWalletManager(logger, config)

	assert.NotNil(t, manager)
	assert.Equal(t, config.EnabledTypes, manager.config.EnabledTypes)
	assert.Equal(t, 0, len(manager.connectedWallets))
}

func TestHardwareWalletManager_DiscoverWallets(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	manager := NewHardwareWalletManager(logger, config)
	ctx := context.Background()

	devices, err := manager.DiscoverWallets(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, devices)

	// Should find both Ledger and Trezor devices (simulated)
	var foundLedger, foundTrezor bool
	for _, device := range devices {
		switch device.Type {
		case HardwareWalletTypeLedger:
			foundLedger = true
			assert.Equal(t, "Nano S Plus", device.Model)
		case HardwareWalletTypeTrezor:
			foundTrezor = true
			assert.Equal(t, "Model T", device.Model)
		}
	}

	assert.True(t, foundLedger, "Should find Ledger device")
	assert.True(t, foundTrezor, "Should find Trezor device")
}

func TestHardwareWalletManager_ConnectWallet_Ledger(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	manager := NewHardwareWalletManager(logger, config)
	ctx := context.Background()

	deviceID := "ledger_test_001"
	wallet, err := manager.ConnectWallet(ctx, deviceID, HardwareWalletTypeLedger)
	require.NoError(t, err)
	assert.NotNil(t, wallet)

	assert.True(t, wallet.IsConnected())
	assert.Equal(t, HardwareWalletTypeLedger, wallet.GetWalletType())

	// Check that wallet is stored in manager
	connectedWallets := manager.GetConnectedWallets()
	assert.Equal(t, 1, len(connectedWallets))
	assert.Contains(t, connectedWallets, deviceID)

	// Clean up
	err = manager.DisconnectWallet(deviceID)
	assert.NoError(t, err)
}

func TestHardwareWalletManager_ConnectWallet_Trezor(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	manager := NewHardwareWalletManager(logger, config)
	ctx := context.Background()

	deviceID := "trezor_test_001"
	wallet, err := manager.ConnectWallet(ctx, deviceID, HardwareWalletTypeTrezor)
	require.NoError(t, err)
	assert.NotNil(t, wallet)

	assert.True(t, wallet.IsConnected())
	assert.Equal(t, HardwareWalletTypeTrezor, wallet.GetWalletType())

	// Clean up
	err = manager.DisconnectWallet(deviceID)
	assert.NoError(t, err)
}

func TestHardwareWalletManager_ConnectWallet_UnsupportedType(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	manager := NewHardwareWalletManager(logger, config)
	ctx := context.Background()

	deviceID := "unknown_device"
	_, err := manager.ConnectWallet(ctx, deviceID, HardwareWalletType("unknown"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported wallet type")
}

func TestLedgerWallet_GetAddress(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	wallet, err := NewLedgerWallet(logger, "test_device", config)
	require.NoError(t, err)

	ctx := context.Background()
	err = wallet.Connect(ctx)
	require.NoError(t, err)

	derivationPath := "m/44'/60'/0'/0/0"
	address, err := wallet.GetAddress(ctx, derivationPath)
	require.NoError(t, err)
	assert.NotEmpty(t, address)
	assert.True(t, common.IsHexAddress(address))

	// Clean up
	err = wallet.Disconnect()
	assert.NoError(t, err)
}

func TestLedgerWallet_GetAddresses(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	wallet, err := NewLedgerWallet(logger, "test_device", config)
	require.NoError(t, err)

	ctx := context.Background()
	err = wallet.Connect(ctx)
	require.NoError(t, err)

	derivationPaths := []string{
		"m/44'/60'/0'/0/0",
		"m/44'/60'/0'/0/1",
		"m/44'/60'/0'/0/2",
	}

	addresses, err := wallet.GetAddresses(ctx, derivationPaths)
	require.NoError(t, err)
	assert.Equal(t, len(derivationPaths), len(addresses))

	for _, address := range addresses {
		assert.True(t, common.IsHexAddress(address))
	}

	// Clean up
	err = wallet.Disconnect()
	assert.NoError(t, err)
}

func TestLedgerWallet_GetPublicKey(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	wallet, err := NewLedgerWallet(logger, "test_device", config)
	require.NoError(t, err)

	ctx := context.Background()
	err = wallet.Connect(ctx)
	require.NoError(t, err)

	derivationPath := "m/44'/60'/0'/0/0"
	publicKey, err := wallet.GetPublicKey(ctx, derivationPath)
	require.NoError(t, err)
	assert.NotEmpty(t, publicKey)
	assert.Equal(t, 65, len(publicKey)) // Uncompressed public key

	// Clean up
	err = wallet.Disconnect()
	assert.NoError(t, err)
}

func TestLedgerWallet_SignTransaction(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	wallet, err := NewLedgerWallet(logger, "test_device", config)
	require.NoError(t, err)

	ctx := context.Background()
	err = wallet.Connect(ctx)
	require.NoError(t, err)

	// Create a test transaction
	tx := types.NewTransaction(
		0,
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(1000000000000000000), // 1 ETH
		21000,
		big.NewInt(20000000000), // 20 gwei
		nil,
	)

	derivationPath := "m/44'/60'/0'/0/0"
	signedTx, err := wallet.SignTransaction(ctx, tx, derivationPath)
	require.NoError(t, err)
	assert.NotNil(t, signedTx)

	// Clean up
	err = wallet.Disconnect()
	assert.NoError(t, err)
}

func TestTrezorWallet_GetAddress(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	wallet, err := NewTrezorWallet(logger, "test_device", config)
	require.NoError(t, err)

	ctx := context.Background()
	err = wallet.Connect(ctx)
	require.NoError(t, err)

	derivationPath := "m/44'/60'/0'/0/0"
	address, err := wallet.GetAddress(ctx, derivationPath)
	require.NoError(t, err)
	assert.NotEmpty(t, address)
	assert.True(t, common.IsHexAddress(address))

	// Clean up
	err = wallet.Disconnect()
	assert.NoError(t, err)
}

func TestTrezorWallet_SignMessage(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	wallet, err := NewTrezorWallet(logger, "test_device", config)
	require.NoError(t, err)

	ctx := context.Background()
	err = wallet.Connect(ctx)
	require.NoError(t, err)

	message := []byte("Hello, Trezor!")
	derivationPath := "m/44'/60'/0'/0/0"
	signature, err := wallet.SignMessage(ctx, message, derivationPath)
	require.NoError(t, err)
	assert.NotEmpty(t, signature)

	// Clean up
	err = wallet.Disconnect()
	assert.NoError(t, err)
}

func TestValidateDerivationPath(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	wallet, err := NewLedgerWallet(logger, "test_device", config)
	require.NoError(t, err)

	// Valid paths
	validPaths := []string{
		"m/44'/60'/0'/0/0",
		"m/44'/60'/1'/0/5",
		"m/44'/60'/0'/1/10",
	}

	for _, path := range validPaths {
		err := wallet.ValidateDerivationPath(path)
		assert.NoError(t, err, "Path should be valid: %s", path)
	}

	// Invalid paths
	invalidPaths := []string{
		"",
		"m/44/60/0/0/0",    // Missing apostrophes
		"44'/60'/0'/0/0",   // Missing 'm/'
		"m/44'/60'/0'/2/0", // Invalid change value (should be 0 or 1)
		"m/44'/60'/0'",     // Incomplete path
	}

	for _, path := range invalidPaths {
		err := wallet.ValidateDerivationPath(path)
		assert.Error(t, err, "Path should be invalid: %s", path)
	}
}

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	assert.Equal(t, 2, len(config.EnabledTypes))
	assert.Contains(t, config.EnabledTypes, HardwareWalletTypeLedger)
	assert.Contains(t, config.EnabledTypes, HardwareWalletTypeTrezor)
	assert.Equal(t, 30*time.Second, config.ConnectionTimeout)
	assert.Equal(t, 60*time.Second, config.SigningTimeout)
	assert.True(t, config.AutoReconnect)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, "m/44'/60'/0'/0/0", config.DefaultDerivationPath)
}

func TestHardwareWalletManager_SignTransactionWithHardwareWallet(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	manager := NewHardwareWalletManager(logger, config)
	ctx := context.Background()

	// Connect a Ledger wallet
	deviceID := "ledger_test_001"
	_, err := manager.ConnectWallet(ctx, deviceID, HardwareWalletTypeLedger)
	require.NoError(t, err)

	// Create a test transaction
	tx := types.NewTransaction(
		0,
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(1000000000000000000), // 1 ETH
		21000,
		big.NewInt(20000000000), // 20 gwei
		nil,
	)

	derivationPath := "m/44'/60'/0'/0/0"
	signedTx, err := manager.SignTransactionWithHardwareWallet(ctx, deviceID, tx, derivationPath)
	require.NoError(t, err)
	assert.NotNil(t, signedTx)

	// Clean up
	err = manager.DisconnectWallet(deviceID)
	assert.NoError(t, err)
}

func TestHardwareWalletManager_GetAddressFromHardwareWallet(t *testing.T) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	manager := NewHardwareWalletManager(logger, config)
	ctx := context.Background()

	// Connect a Trezor wallet
	deviceID := "trezor_test_001"
	_, err := manager.ConnectWallet(ctx, deviceID, HardwareWalletTypeTrezor)
	require.NoError(t, err)

	derivationPath := "m/44'/60'/0'/0/0"
	address, err := manager.GetAddressFromHardwareWallet(ctx, deviceID, derivationPath)
	require.NoError(t, err)
	assert.NotEmpty(t, address)
	assert.True(t, common.IsHexAddress(address))

	// Clean up
	err = manager.DisconnectWallet(deviceID)
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkLedgerWallet_GetAddress(b *testing.B) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	wallet, err := NewLedgerWallet(logger, "test_device", config)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	err = wallet.Connect(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer wallet.Disconnect()

	derivationPath := "m/44'/60'/0'/0/0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := wallet.GetAddress(ctx, derivationPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTrezorWallet_GetAddress(b *testing.B) {
	logger := logger.New("test")
	config := GetDefaultConfig()

	wallet, err := NewTrezorWallet(logger, "test_device", config)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	err = wallet.Connect(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer wallet.Disconnect()

	derivationPath := "m/44'/60'/0'/0/0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := wallet.GetAddress(ctx, derivationPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}
