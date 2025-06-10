package blockchain

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
)

func TestNewSolanaClient(t *testing.T) {
	// Create test config
	cfg := config.SolanaNetworkConfig{
		Network:            "devnet",
		RPCURL:             "https://api.devnet.solana.com",
		WSURL:              "wss://api.devnet.solana.com",
		Cluster:            "devnet",
		Commitment:         "confirmed",
		Timeout:            "30s",
		MaxRetries:         3,
		ConfirmationBlocks: 32,
	}

	// Create logger
	logger := logger.New("test")

	// Create client
	client, err := NewSolanaClient(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Verify configuration
	assert.Equal(t, "devnet", client.GetCluster())
	assert.Equal(t, "confirmed", client.GetCommitment())

	// Close client
	err = client.Close()
	assert.NoError(t, err)
}

func TestSolanaClient_GetBalance_InvalidAddress(t *testing.T) {
	// Create test client
	client := createTestSolanaClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test with invalid address
	_, err := client.GetBalance(ctx, "invalid-address")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestSolanaClient_GetBalance_ValidAddress(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create test client
	client := createTestSolanaClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test with valid address (this might fail if address has no funds, but should not error on format)
	validAddress := "11111111111111111111111111111112" // System program address
	balance, err := client.GetBalance(ctx, validAddress)
	
	// Should not error on valid address format
	if err != nil {
		// If it errors, it should be a network/RPC error, not a format error
		assert.NotContains(t, err.Error(), "invalid address")
	} else {
		// Balance should be non-negative
		assert.True(t, balance.GreaterThanOrEqual(decimal.Zero))
	}
}

func TestSolanaClient_GetTokenBalance_InvalidAddress(t *testing.T) {
	// Create test client
	client := createTestSolanaClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test with invalid address
	_, _, err := client.GetTokenBalance(ctx, "invalid-address", "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestSolanaClient_GetTokenBalance_InvalidMint(t *testing.T) {
	// Create test client
	client := createTestSolanaClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test with invalid mint address
	validAddress := "11111111111111111111111111111112"
	_, _, err := client.GetTokenBalance(ctx, validAddress, "invalid-mint")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mint address")
}

func TestSolanaClient_GetRecentBlockhash(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create test client
	client := createTestSolanaClient(t)
	defer client.Close()

	ctx := context.Background()

	// Get recent blockhash
	blockhash, err := client.GetRecentBlockhash(ctx)
	if err != nil {
		// If it errors, it should be a network error, not a client error
		assert.NotContains(t, err.Error(), "invalid")
		t.Skipf("Network error: %v", err)
	} else {
		// Blockhash should not be empty
		assert.NotEmpty(t, blockhash.String())
	}
}

func TestSolanaClient_GetMinimumBalanceForRentExemption(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create test client
	client := createTestSolanaClient(t)
	defer client.Close()

	ctx := context.Background()

	// Get minimum balance for rent exemption
	balance, err := client.GetMinimumBalanceForRentExemption(ctx, 165) // Typical account size
	if err != nil {
		// If it errors, it should be a network error
		t.Skipf("Network error: %v", err)
	} else {
		// Balance should be positive
		assert.Greater(t, balance, uint64(0))
	}
}

func TestSolanaClient_ConfirmTransaction_InvalidSignature(t *testing.T) {
	// Create test client
	client := createTestSolanaClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test with invalid signature
	err := client.ConfirmTransaction(ctx, "invalid-signature")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestSolanaClient_Configuration(t *testing.T) {
	// Create test config
	cfg := config.SolanaNetworkConfig{
		Network:            "testnet",
		RPCURL:             "https://api.testnet.solana.com",
		WSURL:              "wss://api.testnet.solana.com",
		Cluster:            "testnet",
		Commitment:         "finalized",
		Timeout:            "60s",
		MaxRetries:         5,
		ConfirmationBlocks: 64,
	}

	// Create logger
	logger := logger.New("test")

	// Create client
	client, err := NewSolanaClient(cfg, logger)
	require.NoError(t, err)
	defer client.Close()

	// Verify configuration
	assert.Equal(t, "testnet", client.GetCluster())
	assert.Equal(t, "finalized", client.GetCommitment())
}

func TestSolanaClient_ContextCancellation(t *testing.T) {
	// Create test client
	client := createTestSolanaClient(t)
	defer client.Close()

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This should timeout quickly
	_, err := client.GetBalance(ctx, "11111111111111111111111111111112")
	assert.Error(t, err)
	// Should be a context error
	assert.True(t, 
		err == context.DeadlineExceeded || 
		err == context.Canceled ||
		err.Error() == "context deadline exceeded" ||
		err.Error() == "context canceled")
}

func TestSolanaClient_MultipleOperations(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create test client
	client := createTestSolanaClient(t)
	defer client.Close()

	ctx := context.Background()
	validAddress := "11111111111111111111111111111112"

	// Perform multiple operations
	operations := []func() error{
		func() error {
			_, err := client.GetBalance(ctx, validAddress)
			return err
		},
		func() error {
			_, err := client.GetRecentBlockhash(ctx)
			return err
		},
		func() error {
			_, err := client.GetMinimumBalanceForRentExemption(ctx, 165)
			return err
		},
	}

	// Run operations concurrently
	errors := make(chan error, len(operations))
	for _, op := range operations {
		go func(operation func() error) {
			errors <- operation()
		}(op)
	}

	// Collect results
	var networkErrors int
	for i := 0; i < len(operations); i++ {
		err := <-errors
		if err != nil {
			networkErrors++
			// Should be network errors, not client errors
			assert.NotContains(t, err.Error(), "invalid")
		}
	}

	// If all operations failed, it's likely a network issue
	if networkErrors == len(operations) {
		t.Skip("All operations failed - likely network issue")
	}
}

// Helper function to create test client
func createTestSolanaClient(t *testing.T) *SolanaClient {
	cfg := config.SolanaNetworkConfig{
		Network:            "devnet",
		RPCURL:             "https://api.devnet.solana.com",
		WSURL:              "", // Skip WebSocket for tests
		Cluster:            "devnet",
		Commitment:         "confirmed",
		Timeout:            "30s",
		MaxRetries:         3,
		ConfirmationBlocks: 32,
	}

	logger := logger.New("test")
	client, err := NewSolanaClient(cfg, logger)
	require.NoError(t, err)
	return client
}

func BenchmarkSolanaClient_GetBalance(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	client := createTestSolanaClient(b)
	defer client.Close()

	ctx := context.Background()
	address := "11111111111111111111111111111112"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetBalance(ctx, address)
		if err != nil {
			b.Skip("Network error:", err)
		}
	}
}

func BenchmarkSolanaClient_GetRecentBlockhash(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	client := createTestSolanaClient(b)
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetRecentBlockhash(ctx)
		if err != nil {
			b.Skip("Network error:", err)
		}
	}
}
