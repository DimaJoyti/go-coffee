package defi

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRedisClient is already defined in arbitrage_detector_test.go

func TestNewMEVProtectionService(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled:                true,
		Level:                  MEVProtectionAdvanced,
		UseFlashbots:           true,
		UsePrivateMempool:      true,
		MaxSlippageProtection:  decimal.NewFromFloat(0.05),
		SandwichDetection:      true,
		FrontrunDetection:      true,
		MinBlockConfirmations:  1,
		GasPriceMultiplier:     decimal.NewFromFloat(1.1),
		FlashbotsRelay:         "https://relay.flashbots.net",
		PrivateMempoolEndpoint: "https://api.private-mempool.com/v1",
	}

	service := NewMEVProtectionService(config, logger, mockCache)

	assert.NotNil(t, service)
	assert.Equal(t, config.Level, service.config.Level)
	assert.True(t, service.config.Enabled)
	assert.NotNil(t, service.sandwichDetector)
	assert.NotNil(t, service.frontrunDetector)
	assert.NotNil(t, service.mempoolMonitor)
	assert.NotNil(t, service.flashbotsClient)
	assert.NotNil(t, service.privateMempoolClient)
}

func TestMEVProtectionService_Start(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled:           true,
		Level:             MEVProtectionBasic,
		SandwichDetection: true,
		FrontrunDetection: true,
	}

	service := NewMEVProtectionService(config, logger, mockCache)
	ctx := context.Background()

	err := service.Start(ctx)
	assert.NoError(t, err)

	// Clean up
	service.Stop()
}

func TestMEVProtectionService_StartDisabled(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled: false,
	}

	service := NewMEVProtectionService(config, logger, mockCache)
	ctx := context.Background()

	err := service.Start(ctx)
	assert.NoError(t, err) // Should not error when disabled
}

func TestMEVProtectionService_ProtectTransaction_Disabled(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled: false,
	}

	service := NewMEVProtectionService(config, logger, mockCache)
	ctx := context.Background()

	// Create a mock transaction
	tx := types.NewTransaction(
		0,
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(1000000000000000000), // 1 ETH
		21000,
		big.NewInt(20000000000), // 20 gwei
		nil,
	)

	_, err := service.ProtectTransaction(ctx, tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MEV protection is disabled")
}

func TestMEVProtectionService_ProtectTransaction_BasicLevel(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled:            true,
		Level:              MEVProtectionBasic,
		GasPriceMultiplier: decimal.NewFromFloat(1.2), // 20% increase
	}

	service := NewMEVProtectionService(config, logger, mockCache)
	ctx := context.Background()

	// Create a mock transaction
	originalGasPrice := big.NewInt(20000000000) // 20 gwei
	tx := types.NewTransaction(
		0,
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(1000000000000000000), // 1 ETH
		21000,
		originalGasPrice,
		nil,
	)

	protectedTx, err := service.ProtectTransaction(ctx, tx)
	require.NoError(t, err)
	assert.NotNil(t, protectedTx)
	assert.Equal(t, MEVProtectionBasic, protectedTx.ProtectionLevel)
	assert.Equal(t, "standard", protectedTx.SubmissionMethod)
	assert.Equal(t, "pending", protectedTx.Status)

	// Check that gas price was increased
	expectedGasPrice := new(big.Int).Mul(originalGasPrice, big.NewInt(1)) // 1.2 * 20 gwei = 24 gwei
	expectedGasPrice.Mul(expectedGasPrice, big.NewInt(120))
	expectedGasPrice.Div(expectedGasPrice, big.NewInt(100))

	assert.True(t, protectedTx.ProtectedGasPrice.Cmp(originalGasPrice) > 0)
}

func TestMEVProtectionService_GetMetrics(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled: true,
		Level:   MEVProtectionBasic,
	}

	service := NewMEVProtectionService(config, logger, mockCache)

	metrics := service.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalTransactions)
	assert.Equal(t, int64(0), metrics.ProtectedTransactions)
	assert.Equal(t, int64(0), metrics.DetectedAttacks)
	assert.Equal(t, int64(0), metrics.PreventedAttacks)
}

func TestMEVProtectionService_GetDetectedAttacks(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled: true,
		Level:   MEVProtectionBasic,
	}

	service := NewMEVProtectionService(config, logger, mockCache)

	attacks := service.GetDetectedAttacks()
	assert.NotNil(t, attacks)
	assert.Equal(t, 0, len(attacks))
}

func TestMEVProtectionService_RecordDetection(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	// Mock cache operations
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

	config := MEVProtectionConfig{
		Enabled: true,
		Level:   MEVProtectionBasic,
	}

	service := NewMEVProtectionService(config, logger, mockCache)

	detection := &MEVDetection{
		ID:                "test-detection-1",
		Type:              MEVAttackSandwich,
		TargetTransaction: "0xabc123",
		AttackerAddress:   "0x1234567890123456789012345678901234567890",
		VictimAddress:     "0x0987654321098765432109876543210987654321",
		TokenAddress:      "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
		EstimatedLoss:     decimal.NewFromFloat(0.1),
		Confidence:        decimal.NewFromFloat(0.85),
		BlockNumber:       12345678,
		Timestamp:         time.Now(),
		Prevented:         true,
		PreventionMethod:  "flashbots",
	}

	service.recordDetection(detection)

	// Check that detection was recorded
	attacks := service.GetDetectedAttacks()
	assert.Equal(t, 1, len(attacks))
	assert.Contains(t, attacks, detection.ID)

	// Check metrics were updated
	metrics := service.GetMetrics()
	assert.Equal(t, int64(1), metrics.DetectedAttacks)
	assert.Equal(t, int64(1), metrics.PreventedAttacks)
	assert.True(t, metrics.TotalSavings.Equal(detection.EstimatedLoss))

	mockCache.AssertExpectations(t)
}

func TestMEVProtectionService_GenerateBundleID(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled: true,
		Level:   MEVProtectionBasic,
	}

	service := NewMEVProtectionService(config, logger, mockCache)

	bundleID1 := service.generateBundleID()
	bundleID2 := service.generateBundleID()

	assert.NotEmpty(t, bundleID1)
	assert.NotEmpty(t, bundleID2)
	assert.NotEqual(t, bundleID1, bundleID2)
	assert.Equal(t, 32, len(bundleID1)) // 16 bytes * 2 (hex encoding)
}

func TestMEVProtectionService_UpdateMetrics(t *testing.T) {
	logger := logger.New("test")
	mockCache := &MockRedisClient{}

	// Mock cache operations
	mockCache.On("Set", mock.Anything, "mev:metrics", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

	config := MEVProtectionConfig{
		Enabled: true,
		Level:   MEVProtectionBasic,
	}

	service := NewMEVProtectionService(config, logger, mockCache)

	// Add some protected transactions
	service.protectedTxs["tx1"] = &ProtectedTransaction{Hash: "tx1"}
	service.protectedTxs["tx2"] = &ProtectedTransaction{Hash: "tx2"}

	service.updateMetrics()

	metrics := service.GetMetrics()
	assert.Equal(t, int64(2), metrics.TotalTransactions)
	assert.True(t, time.Since(metrics.LastUpdate) < time.Second)

	mockCache.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkMEVProtectionService_ProtectTransaction(b *testing.B) {
	logger := logger.New("benchmark")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled:            true,
		Level:              MEVProtectionBasic,
		GasPriceMultiplier: decimal.NewFromFloat(1.1),
	}

	service := NewMEVProtectionService(config, logger, mockCache)
	ctx := context.Background()

	tx := types.NewTransaction(
		0,
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(1000000000000000000),
		21000,
		big.NewInt(20000000000),
		nil,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ProtectTransaction(ctx, tx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMEVProtectionService_GetMetrics(b *testing.B) {
	logger := logger.New("benchmark")
	mockCache := &MockRedisClient{}

	config := MEVProtectionConfig{
		Enabled: true,
		Level:   MEVProtectionBasic,
	}

	service := NewMEVProtectionService(config, logger, mockCache)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.GetMetrics()
	}
}
