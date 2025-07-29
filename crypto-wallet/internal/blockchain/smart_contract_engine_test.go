package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger
func createTestLogger() *logger.Logger {
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	return logger.NewLogger(logConfig)
}

func TestNewSmartContractEngine(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()

	engine := NewSmartContractEngine(logger, config)

	assert.NotNil(t, engine)
	assert.Equal(t, config.DefaultChain, engine.config.DefaultChain)
	assert.Equal(t, config.SupportedChains, engine.config.SupportedChains)
	assert.False(t, engine.isRunning)
	assert.NotNil(t, engine.clients)
	assert.NotNil(t, engine.contracts)
	assert.NotNil(t, engine.protocolAdapters)
}

func TestSmartContractEngine_StartStop(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()
	// Use mock endpoints to avoid actual network calls
	config.RPCEndpoints = map[string]string{
		"ethereum": "http://localhost:8545",
	}
	config.SupportedChains = []string{"ethereum"}

	engine := NewSmartContractEngine(logger, config)
	ctx := context.Background()

	// Test starting (will fail due to mock endpoint, but that's expected)
	err := engine.Start(ctx)
	// We expect this to fail since we're using a mock endpoint
	assert.Error(t, err)

	// Test stopping
	err = engine.Stop()
	assert.NoError(t, err)
	assert.False(t, engine.IsRunning())
}

func TestSmartContractEngine_StartDisabled(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()
	config.Enabled = false

	engine := NewSmartContractEngine(logger, config)
	ctx := context.Background()

	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.False(t, engine.IsRunning()) // Should remain false when disabled
}

func TestSmartContractEngine_RegisterContract(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()

	engine := NewSmartContractEngine(logger, config)

	contract := &ContractDefinition{
		Name:     "Test Contract",
		Address:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Chain:    "ethereum",
		Protocol: "test",
		Category: "Testing",
		Version:  "1.0",
		Functions: map[string]FunctionDef{
			"testFunction": {
				Name:            "testFunction",
				Signature:       "testFunction(uint256)",
				StateMutability: "nonpayable",
				GasEstimate:     100000,
				Description:     "Test function",
			},
		},
		SecurityScore: decimal.NewFromFloat(0.8),
		TrustLevel:    "medium",
		IsActive:      true,
	}

	err := engine.RegisterContract(contract)
	assert.NoError(t, err)

	// Test retrieving the contract
	retrievedContract, err := engine.GetContract("ethereum", contract.Address)
	assert.NoError(t, err)
	assert.Equal(t, contract.Name, retrievedContract.Name)
	assert.Equal(t, contract.Address, retrievedContract.Address)
	assert.Equal(t, contract.Protocol, retrievedContract.Protocol)
}

func TestSmartContractEngine_GetContract(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()

	engine := NewSmartContractEngine(logger, config)

	// Test getting non-existent contract
	_, err := engine.GetContract("ethereum", common.HexToAddress("0x1234567890123456789012345678901234567890"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contract not found")
}

func TestSmartContractEngine_ListContracts(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()

	engine := NewSmartContractEngine(logger, config)

	// Initially should be empty
	contracts := engine.ListContracts()
	assert.Empty(t, contracts)

	// Register a contract
	contract := &ContractDefinition{
		Name:     "Test Contract",
		Address:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Chain:    "ethereum",
		Protocol: "test",
		Category: "Testing",
		Version:  "1.0",
		IsActive: true,
	}

	err := engine.RegisterContract(contract)
	require.NoError(t, err)

	// Should now have one contract
	contracts = engine.ListContracts()
	assert.Len(t, contracts, 1)
	assert.Equal(t, contract.Name, contracts[0].Name)
}

func TestTransactionManager(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()

	tm := NewTransactionManager(logger, config)
	ctx := context.Background()

	err := tm.Start(ctx)
	assert.NoError(t, err)

	// Test executing a transaction
	request := &TransactionRequest{
		ID:              "test-tx-1",
		Chain:           "ethereum",
		ContractAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		FunctionName:    "testFunction",
		Inputs:          []interface{}{big.NewInt(100)},
		GasLimit:        100000,
		GasPrice:        big.NewInt(20000000000),
		From:            common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
		Deadline:        time.Now().Add(10 * time.Minute),
		Priority:        PriorityNormal,
		CreatedAt:       time.Now(),
	}

	result, err := tm.ExecuteTransaction(ctx, request)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, request.ID, result.ID)
	assert.Equal(t, StatusConfirmed, result.Status)

	// Test getting transaction status
	status, err := tm.GetTransactionStatus(request.ID)
	assert.NoError(t, err)
	assert.Equal(t, StatusConfirmed, status.Status)

	// Test getting pending transactions (should be empty after completion)
	pending := tm.GetPendingTransactions()
	assert.Empty(t, pending)

	// Test getting completed transactions
	completed := tm.GetCompletedTransactions()
	assert.Len(t, completed, 1)

	err = tm.Stop()
	assert.NoError(t, err)
}

func TestGasOracle(t *testing.T) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()

	gasOracle := NewGasOracle(logger, config)
	ctx := context.Background()

	err := gasOracle.Start(ctx)
	assert.NoError(t, err)

	// Test getting gas price
	gasPrice, err := gasOracle.GetGasPrice(ctx, "ethereum", PriorityNormal)
	assert.NoError(t, err)
	assert.NotNil(t, gasPrice)
	assert.True(t, gasPrice.Cmp(big.NewInt(0)) > 0)

	// Test getting EIP-1559 gas prices
	maxFee, priorityFee, err := gasOracle.GetEIP1559GasPrice(ctx, "ethereum", PriorityHigh)
	assert.NoError(t, err)
	assert.NotNil(t, maxFee)
	assert.NotNil(t, priorityFee)
	assert.True(t, maxFee.Cmp(priorityFee) > 0)

	// Test estimating gas cost
	estimate, err := gasOracle.EstimateGasCost(ctx, "ethereum", 100000, PriorityNormal)
	assert.NoError(t, err)
	assert.NotNil(t, estimate)
	assert.Equal(t, uint64(100000), estimate.GasLimit)
	assert.True(t, estimate.EstimatedCost.GreaterThan(decimal.Zero))

	err = gasOracle.Stop()
	assert.NoError(t, err)
}

func TestNonceManager(t *testing.T) {
	logger := createTestLogger()
	nm := NewNonceManager(logger)
	ctx := context.Background()

	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	chain := "ethereum"

	// Test getting nonce for new address
	nonce, err := nm.GetNonce(ctx, chain, address)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, nonce, uint64(0))

	// Test getting nonce again (should increment)
	nextNonce, err := nm.GetNonce(ctx, chain, address)
	assert.NoError(t, err)
	assert.Equal(t, nonce+1, nextNonce)

	// Test setting nonce
	nm.SetNonce(chain, address, 100)
	currentNonce, exists := nm.GetCurrentNonce(chain, address)
	assert.True(t, exists)
	assert.Equal(t, uint64(100), currentNonce)

	// Test incrementing nonce
	newNonce := nm.IncrementNonce(chain, address)
	assert.Equal(t, uint64(101), newNonce)

	// Test decrementing nonce
	decrementedNonce := nm.DecrementNonce(chain, address)
	assert.Equal(t, uint64(100), decrementedNonce)

	// Test resetting nonce
	err = nm.ResetNonce(ctx, chain, address)
	assert.NoError(t, err)

	// Test syncing nonce
	err = nm.SyncNonce(ctx, chain, address)
	assert.NoError(t, err)

	// Test getting all nonces
	allNonces := nm.GetAllNonces()
	assert.Contains(t, allNonces, chain)
	assert.Contains(t, allNonces[chain], address)

	// Test clearing nonces
	nm.ClearNonces()
	allNonces = nm.GetAllNonces()
	assert.Empty(t, allNonces)
}

func TestProtocolAdapters(t *testing.T) {
	logger := createTestLogger()
	config := ProtocolConfig{
		Enabled: true,
		Version: "1.0",
	}

	// Test Uniswap adapter
	uniswapAdapter := NewUniswapAdapter(logger, config)
	assert.Equal(t, "uniswap", uniswapAdapter.GetProtocolName())
	assert.Contains(t, uniswapAdapter.GetSupportedChains(), "ethereum")
	assert.True(t, uniswapAdapter.SupportsFeature("swapping"))
	assert.False(t, uniswapAdapter.SupportsFeature("lending"))

	// Test Aave adapter
	aaveAdapter := NewAaveAdapter(logger, config)
	assert.Equal(t, "aave", aaveAdapter.GetProtocolName())
	assert.Contains(t, aaveAdapter.GetSupportedChains(), "ethereum")
	assert.True(t, aaveAdapter.SupportsFeature("lending"))
	assert.False(t, aaveAdapter.SupportsFeature("swapping"))

	// Test Curve adapter
	curveAdapter := NewCurveAdapter(logger, config)
	assert.Equal(t, "curve", curveAdapter.GetProtocolName())
	assert.Contains(t, curveAdapter.GetSupportedChains(), "ethereum")
	assert.True(t, curveAdapter.SupportsFeature("swapping"))
	assert.True(t, curveAdapter.SupportsFeature("staking"))
}

func TestGetDefaultSmartContractConfig(t *testing.T) {
	config := GetDefaultSmartContractConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, "ethereum", config.DefaultChain)
	assert.Contains(t, config.SupportedChains, "ethereum")
	assert.Contains(t, config.SupportedChains, "polygon")
	assert.NotNil(t, config.MaxGasPrice)
	assert.True(t, config.GasMultiplier.GreaterThan(decimal.NewFromFloat(1.0)))
	assert.True(t, config.EnableBatchTransactions)
	assert.NotEmpty(t, config.ProtocolConfigs)
	assert.Contains(t, config.ProtocolConfigs, "uniswap")
	assert.Contains(t, config.ProtocolConfigs, "aave")
	assert.Contains(t, config.ProtocolConfigs, "curve")
}

// Benchmark tests
func BenchmarkTransactionExecution(b *testing.B) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()
	tm := NewTransactionManager(logger, config)
	ctx := context.Background()

	tm.Start(ctx)
	defer tm.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := &TransactionRequest{
			ID:              fmt.Sprintf("bench-tx-%d", i),
			Chain:           "ethereum",
			ContractAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
			FunctionName:    "testFunction",
			GasLimit:        100000,
			GasPrice:        big.NewInt(20000000000),
			From:            common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
			Deadline:        time.Now().Add(10 * time.Minute),
			CreatedAt:       time.Now(),
		}

		_, err := tm.ExecuteTransaction(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGasPriceRetrieval(b *testing.B) {
	logger := createTestLogger()
	config := GetDefaultSmartContractConfig()
	gasOracle := NewGasOracle(logger, config)
	ctx := context.Background()

	gasOracle.Start(ctx)
	defer gasOracle.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gasOracle.GetGasPrice(ctx, "ethereum", PriorityNormal)
		if err != nil {
			b.Fatal(err)
		}
	}
}
