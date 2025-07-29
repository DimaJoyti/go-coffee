package defi

import (
	"context"
	"testing"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/gagliardetto/solana-go"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRaydiumClient(t *testing.T) {
	// Create logger
	logger := logger.New("test")

	// Create client
	client, err := NewRaydiumClient("https://api.devnet.solana.com", logger)
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Close client
	err = client.Close()
	assert.NoError(t, err)
}

func TestRaydiumClient_GetPools(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Get pools
	pools, err := client.GetPools(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, pools)

	// Verify pool structure
	for _, pool := range pools {
		assert.NotEmpty(t, pool.ID)
		assert.NotEmpty(t, pool.TokenA)
		assert.NotEmpty(t, pool.TokenB)
		assert.True(t, pool.TokenAAmount.GreaterThan(decimal.Zero))
		assert.True(t, pool.TokenBAmount.GreaterThan(decimal.Zero))
		assert.True(t, pool.LPTokens.GreaterThan(decimal.Zero))
		assert.True(t, pool.Fee.GreaterThanOrEqual(decimal.Zero))
		assert.True(t, pool.APY.GreaterThanOrEqual(decimal.Zero))
	}
}

func TestRaydiumClient_GetSwapQuote(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test SOL -> USDC swap
	inputToken := "So11111111111111111111111111111111111111112"   // SOL
	outputToken := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" // USDC
	inputAmount := decimal.NewFromFloat(1.0)

	quote, err := client.GetSwapQuote(ctx, inputToken, outputToken, inputAmount)
	require.NoError(t, err)
	assert.NotNil(t, quote)

	// Verify quote structure
	assert.Equal(t, inputToken, quote.InputToken)
	assert.Equal(t, outputToken, quote.OutputToken)
	assert.Equal(t, inputAmount, quote.InputAmount)
	assert.True(t, quote.OutputAmount.GreaterThan(decimal.Zero))
	assert.True(t, quote.PriceImpact.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, quote.Fee.GreaterThanOrEqual(decimal.Zero))
	assert.NotEmpty(t, quote.Route)
}

func TestRaydiumClient_GetSwapQuote_ReverseSwap(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test USDC -> SOL swap
	inputToken := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" // USDC
	outputToken := "So11111111111111111111111111111111111111112" // SOL
	inputAmount := decimal.NewFromFloat(50.0)

	quote, err := client.GetSwapQuote(ctx, inputToken, outputToken, inputAmount)
	require.NoError(t, err)
	assert.NotNil(t, quote)

	// Verify quote structure
	assert.Equal(t, inputToken, quote.InputToken)
	assert.Equal(t, outputToken, quote.OutputToken)
	assert.Equal(t, inputAmount, quote.InputAmount)
	assert.True(t, quote.OutputAmount.GreaterThan(decimal.Zero))
}

func TestRaydiumClient_GetSwapQuote_ZeroAmount(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test with zero amount
	inputToken := "So11111111111111111111111111111111111111112"
	outputToken := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	inputAmount := decimal.Zero

	quote, err := client.GetSwapQuote(ctx, inputToken, outputToken, inputAmount)
	require.NoError(t, err)
	assert.NotNil(t, quote)

	// Output should also be zero
	assert.True(t, quote.OutputAmount.Equal(decimal.Zero))
}

func TestRaydiumClient_ExecuteSwap(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Create test quote
	quote := &RaydiumSwapQuote{
		InputToken:   "So11111111111111111111111111111111111111112",
		OutputToken:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		InputAmount:  decimal.NewFromFloat(1.0),
		OutputAmount: decimal.NewFromFloat(50.0),
		PriceImpact:  decimal.NewFromFloat(0.001),
		Fee:          decimal.NewFromFloat(0.125),
		Route:        []string{"So11111111111111111111111111111111111111112", "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"},
	}

	// Create test wallet
	userWallet := solana.NewWallet().PrivateKey

	// Execute swap
	signature, err := client.ExecuteSwap(ctx, quote, userWallet)
	require.NoError(t, err)
	assert.NotEmpty(t, signature)
}

func TestRaydiumClient_AddLiquidity(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test add liquidity
	poolID := "58oQChx4yWmvKdwLLZzBi4ChoCc2fqCUWBkwMihLYQo2"
	tokenAAmount := decimal.NewFromFloat(1.0)
	tokenBAmount := decimal.NewFromFloat(50.0)
	userWallet := solana.NewWallet().PrivateKey

	signature, err := client.AddLiquidity(ctx, poolID, tokenAAmount, tokenBAmount, userWallet)
	require.NoError(t, err)
	assert.NotEmpty(t, signature)
}

func TestRaydiumClient_RemoveLiquidity(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test remove liquidity
	poolID := "58oQChx4yWmvKdwLLZzBi4ChoCc2fqCUWBkwMihLYQo2"
	lpTokenAmount := decimal.NewFromFloat(10.0)
	userWallet := solana.NewWallet().PrivateKey

	signature, err := client.RemoveLiquidity(ctx, poolID, lpTokenAmount, userWallet)
	require.NoError(t, err)
	assert.NotEmpty(t, signature)
}

func TestRaydiumClient_GetPoolInfo(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Get pool info
	poolID := "58oQChx4yWmvKdwLLZzBi4ChoCc2fqCUWBkwMihLYQo2"
	pool, err := client.GetPoolInfo(ctx, poolID)
	require.NoError(t, err)
	assert.NotNil(t, pool)

	// Verify pool structure
	assert.Equal(t, poolID, pool.ID)
	assert.NotEmpty(t, pool.TokenA)
	assert.NotEmpty(t, pool.TokenB)
	assert.True(t, pool.TokenAAmount.GreaterThan(decimal.Zero))
	assert.True(t, pool.TokenBAmount.GreaterThan(decimal.Zero))
	assert.True(t, pool.LPTokens.GreaterThan(decimal.Zero))
	assert.True(t, pool.Fee.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, pool.APY.GreaterThanOrEqual(decimal.Zero))
}

func TestRaydiumClient_GetUserPositions(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Get user positions
	userAddress := "11111111111111111111111111111112"
	positions, err := client.GetUserPositions(ctx, userAddress)
	require.NoError(t, err)
	assert.NotNil(t, positions)

	// Verify positions structure
	for _, position := range positions {
		assert.NotEmpty(t, position.ID)
		assert.NotEmpty(t, position.TokenA)
		assert.NotEmpty(t, position.TokenB)
		assert.True(t, position.TokenAAmount.GreaterThanOrEqual(decimal.Zero))
		assert.True(t, position.TokenBAmount.GreaterThanOrEqual(decimal.Zero))
		assert.True(t, position.LPTokens.GreaterThanOrEqual(decimal.Zero))
	}
}

func TestRaydiumClient_SwapQuote_PriceCalculation(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Test different amounts to verify price calculation
	amounts := []float64{0.1, 1.0, 10.0}
	inputToken := "So11111111111111111111111111111111111111112"
	outputToken := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"

	var previousRatio decimal.Decimal
	for i, amount := range amounts {
		quote, err := client.GetSwapQuote(ctx, inputToken, outputToken, decimal.NewFromFloat(amount))
		require.NoError(t, err)

		// Calculate price ratio
		ratio := quote.OutputAmount.Div(quote.InputAmount)

		if i > 0 {
			// Price should be relatively stable (within reasonable bounds)
			diff := ratio.Sub(previousRatio).Abs()
			maxDiff := previousRatio.Mul(decimal.NewFromFloat(0.1)) // 10% tolerance
			assert.True(t, diff.LessThanOrEqual(maxDiff),
				"Price ratio should be relatively stable across different amounts")
		}

		previousRatio = ratio
	}
}

func TestRaydiumClient_SwapQuote_FeeCalculation(t *testing.T) {
	// Create test client
	client := createTestRaydiumClient(t)
	defer client.Close()

	ctx := context.Background()

	// Get quote
	quote, err := client.GetSwapQuote(ctx,
		"So11111111111111111111111111111111111111112",
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		decimal.NewFromFloat(1.0))
	require.NoError(t, err)

	// Fee should be reasonable (less than 1% of output)
	maxFee := quote.OutputAmount.Mul(decimal.NewFromFloat(0.01))
	assert.True(t, quote.Fee.LessThanOrEqual(maxFee),
		"Fee should be reasonable (less than 1% of output)")

	// Fee should be positive
	assert.True(t, quote.Fee.GreaterThan(decimal.Zero), "Fee should be positive")
}

// Helper function to create test client
func createTestRaydiumClient(t *testing.T) *RaydiumClient {
	logger := logger.New("test")
	client, err := NewRaydiumClient("https://api.devnet.solana.com", logger)
	require.NoError(t, err)
	return client
}

// Helper function to create test client for benchmarks
func createBenchmarkRaydiumClient(b *testing.B) *RaydiumClient {
	logger := logger.New("benchmark")
	client, err := NewRaydiumClient("https://api.devnet.solana.com", logger)
	if err != nil {
		b.Fatal(err)
	}
	return client
}

func BenchmarkRaydiumClient_GetPools(b *testing.B) {
	client := createBenchmarkRaydiumClient(b)
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetPools(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRaydiumClient_GetSwapQuote(b *testing.B) {
	client := createBenchmarkRaydiumClient(b)
	defer client.Close()

	ctx := context.Background()
	inputToken := "So11111111111111111111111111111111111111112"
	outputToken := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	amount := decimal.NewFromFloat(1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetSwapQuote(ctx, inputToken, outputToken, amount)
		if err != nil {
			b.Fatal(err)
		}
	}
}
