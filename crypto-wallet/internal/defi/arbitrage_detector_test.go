package defi

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// MockRedisClient мок для Redis клієнта
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

// MockUniswapClient мок для Uniswap клієнта
type MockUniswapClient struct {
	mock.Mock
}

func (m *MockUniswapClient) GetSwapQuote(ctx context.Context, req *GetSwapQuoteRequest) (*GetSwapQuoteResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*GetSwapQuoteResponse), args.Error(1)
}

func (m *MockUniswapClient) GetLiquidityPools(ctx context.Context, req *GetLiquidityPoolsRequest) (*GetLiquidityPoolsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*GetLiquidityPoolsResponse), args.Error(1)
}

// MockOneInchClient мок для 1inch клієнта
type MockOneInchClient struct {
	mock.Mock
}

func (m *MockOneInchClient) GetSwapQuote(ctx context.Context, req *GetSwapQuoteRequest) (*GetSwapQuoteResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*GetSwapQuoteResponse), args.Error(1)
}

func TestArbitrageDetector_DetectArbitrageForToken(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}
	mockUniswap := &MockUniswapClient{}
	mockOneInch := &MockOneInchClient{}

	detector := NewArbitrageDetector(logger, mockRedis, mockUniswap, mockOneInch)

	ctx := context.Background()
	testToken := Token{
		Address: "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
		Symbol:  "USDC",
		Chain:   ChainEthereum,
	}

	// Mock Uniswap response
	uniswapQuote := &GetSwapQuoteResponse{
		AmountOut:    decimal.NewFromFloat(2500.0), // $2500 per token
		PriceImpact:  decimal.NewFromFloat(0.001),
		GasEstimate:  decimal.NewFromFloat(0.005),
	}
	mockUniswap.On("GetSwapQuote", ctx, mock.AnythingOfType("*defi.GetSwapQuoteRequest")).Return(uniswapQuote, nil)

	// Mock 1inch response (higher price for arbitrage opportunity)
	oneInchQuote := &GetSwapQuoteResponse{
		AmountOut:    decimal.NewFromFloat(2530.0), // $2530 per token (1.2% higher)
		PriceImpact:  decimal.NewFromFloat(0.001),
		GasEstimate:  decimal.NewFromFloat(0.005),
	}
	mockOneInch.On("GetSwapQuote", ctx, mock.AnythingOfType("*defi.GetSwapQuoteRequest")).Return(oneInchQuote, nil)

	// Act
	opportunities, err := detector.DetectArbitrageForToken(ctx, testToken)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, opportunities)

	if len(opportunities) > 0 {
		opp := opportunities[0]
		assert.Equal(t, testToken.Symbol, opp.Token.Symbol)
		assert.True(t, opp.ProfitMargin.GreaterThan(decimal.Zero))
		assert.True(t, opp.NetProfit.GreaterThan(decimal.Zero))
		assert.NotEmpty(t, opp.ID)
		assert.Equal(t, OpportunityStatusDetected, opp.Status)
	}

	mockUniswap.AssertExpectations(t)
	mockOneInch.AssertExpectations(t)
}

func TestArbitrageDetector_CalculateArbitrageOpportunity(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}
	mockUniswap := &MockUniswapClient{}
	mockOneInch := &MockOneInchClient{}

	detector := NewArbitrageDetector(logger, mockRedis, mockUniswap, mockOneInch)

	testToken := Token{
		Address: "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
		Symbol:  "USDC",
		Chain:   ChainEthereum,
	}

	sourceExchange := Exchange{
		ID:   "uniswap-v3",
		Name: "Uniswap V3",
		Type: ExchangeTypeDEX,
	}

	targetExchange := Exchange{
		ID:   "1inch-aggregator",
		Name: "1inch Aggregator",
		Type: ExchangeTypeDEX,
	}

	sourcePrice := decimal.NewFromFloat(2500.0)
	targetPrice := decimal.NewFromFloat(2530.0) // 1.2% higher

	// Act
	opportunity := detector.calculateArbitrageOpportunity(
		testToken, sourceExchange, targetExchange, sourcePrice, targetPrice)

	// Assert
	require.NotNil(t, opportunity)
	assert.Equal(t, testToken.Symbol, opportunity.Token.Symbol)
	assert.Equal(t, sourceExchange.ID, opportunity.SourceExchange.ID)
	assert.Equal(t, targetExchange.ID, opportunity.TargetExchange.ID)
	assert.Equal(t, sourcePrice, opportunity.SourcePrice)
	assert.Equal(t, targetPrice, opportunity.TargetPrice)

	expectedProfitMargin := targetPrice.Sub(sourcePrice).Div(sourcePrice)
	assert.True(t, opportunity.ProfitMargin.Equal(expectedProfitMargin))
	assert.True(t, opportunity.NetProfit.GreaterThan(decimal.Zero))
	assert.True(t, opportunity.Confidence.GreaterThan(decimal.Zero))
}

func TestArbitrageDetector_MinimumProfitMargin(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}
	mockUniswap := &MockUniswapClient{}
	mockOneInch := &MockOneInchClient{}

	detector := NewArbitrageDetector(logger, mockRedis, mockUniswap, mockOneInch)

	testToken := Token{Symbol: "TEST", Chain: ChainEthereum}
	sourceExchange := Exchange{ID: "source", Name: "Source"}
	targetExchange := Exchange{ID: "target", Name: "Target"}

	// Test case 1: Profit margin below minimum (should return nil)
	sourcePrice1 := decimal.NewFromFloat(1000.0)
	targetPrice1 := decimal.NewFromFloat(1002.0) // 0.2% profit (below 0.5% minimum)

	opportunity1 := detector.calculateArbitrageOpportunity(
		testToken, sourceExchange, targetExchange, sourcePrice1, targetPrice1)

	assert.Nil(t, opportunity1, "Should not create opportunity with profit margin below minimum")

	// Test case 2: Profit margin above minimum (should return opportunity)
	sourcePrice2 := decimal.NewFromFloat(1000.0)
	targetPrice2 := decimal.NewFromFloat(1010.0) // 1.0% profit (above 0.5% minimum)

	opportunity2 := detector.calculateArbitrageOpportunity(
		testToken, sourceExchange, targetExchange, sourcePrice2, targetPrice2)

	assert.NotNil(t, opportunity2, "Should create opportunity with profit margin above minimum")
	assert.True(t, opportunity2.ProfitMargin.GreaterThan(decimal.NewFromFloat(0.005)))
}

func TestArbitrageDetector_RiskCalculation(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}
	mockUniswap := &MockUniswapClient{}
	mockOneInch := &MockOneInchClient{}

	detector := NewArbitrageDetector(logger, mockRedis, mockUniswap, mockOneInch)

	// Test different risk scenarios
	testCases := []struct {
		name           string
		profitMargin   decimal.Decimal
		volume         decimal.Decimal
		gasCost        decimal.Decimal
		expectedRisk   RiskLevel
	}{
		{
			name:         "Low Risk - High profit, low gas",
			profitMargin: decimal.NewFromFloat(0.02), // 2%
			volume:       decimal.NewFromFloat(10.0),
			gasCost:      decimal.NewFromFloat(0.01), // $0.01
			expectedRisk: RiskLevelLow,
		},
		{
			name:         "Medium Risk - Medium profit, medium gas",
			profitMargin: decimal.NewFromFloat(0.01), // 1%
			volume:       decimal.NewFromFloat(5.0),
			gasCost:      decimal.NewFromFloat(0.02), // $0.02
			expectedRisk: RiskLevelMedium,
		},
		{
			name:         "High Risk - Low profit, high gas",
			profitMargin: decimal.NewFromFloat(0.006), // 0.6%
			volume:       decimal.NewFromFloat(1.0),
			gasCost:      decimal.NewFromFloat(0.005), // $0.005
			expectedRisk: RiskLevelHigh,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			risk := detector.calculateRisk(tc.profitMargin, tc.volume, tc.gasCost)

			// Assert
			assert.Equal(t, tc.expectedRisk, risk)
		})
	}
}

func TestArbitrageDetector_ConfidenceCalculation(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}
	mockUniswap := &MockUniswapClient{}
	mockOneInch := &MockOneInchClient{}

	detector := NewArbitrageDetector(logger, mockRedis, mockUniswap, mockOneInch)

	// Test confidence calculation
	testCases := []struct {
		name           string
		profitMargin   decimal.Decimal
		volume         decimal.Decimal
		minConfidence  decimal.Decimal
		maxConfidence  decimal.Decimal
	}{
		{
			name:          "High confidence - High profit margin",
			profitMargin:  decimal.NewFromFloat(0.05), // 5%
			volume:        decimal.NewFromFloat(20.0),
			minConfidence: decimal.NewFromFloat(0.8),
			maxConfidence: decimal.NewFromFloat(1.0),
		},
		{
			name:          "Medium confidence - Medium profit margin",
			profitMargin:  decimal.NewFromFloat(0.01), // 1%
			volume:        decimal.NewFromFloat(5.0),
			minConfidence: decimal.NewFromFloat(0.3),
			maxConfidence: decimal.NewFromFloat(0.7),
		},
		{
			name:          "Low confidence - Low profit margin",
			profitMargin:  decimal.NewFromFloat(0.006), // 0.6%
			volume:        decimal.NewFromFloat(1.0),
			minConfidence: decimal.NewFromFloat(0.0),
			maxConfidence: decimal.NewFromFloat(0.4),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			confidence := detector.calculateConfidence(tc.profitMargin, tc.volume)

			// Assert
			assert.True(t, confidence.GreaterThanOrEqual(tc.minConfidence),
				"Confidence should be >= %s, got %s", tc.minConfidence, confidence)
			assert.True(t, confidence.LessThanOrEqual(tc.maxConfidence),
				"Confidence should be <= %s, got %s", tc.maxConfidence, confidence)
		})
	}
}

// Benchmark tests
func BenchmarkArbitrageDetector_CalculateOpportunity(b *testing.B) {
	logger := logger.New("benchmark")
	mockRedis := &MockRedisClient{}
	mockUniswap := &MockUniswapClient{}
	mockOneInch := &MockOneInchClient{}

	detector := NewArbitrageDetector(logger, mockRedis, mockUniswap, mockOneInch)

	testToken := Token{Symbol: "BENCH", Chain: ChainEthereum}
	sourceExchange := Exchange{ID: "source", Name: "Source"}
	targetExchange := Exchange{ID: "target", Name: "Target"}
	sourcePrice := decimal.NewFromFloat(1000.0)
	targetPrice := decimal.NewFromFloat(1010.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.calculateArbitrageOpportunity(
			testToken, sourceExchange, targetExchange, sourcePrice, targetPrice)
	}
}

func TestArbitrageDetector_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	logger := logger.New("integration-test")
	mockRedis := &MockRedisClient{}
	mockUniswap := &MockUniswapClient{}
	mockOneInch := &MockOneInchClient{}

	detector := NewArbitrageDetector(logger, mockRedis, mockUniswap, mockOneInch)

	ctx := context.Background()

	// Mock successful cache operations
	mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Mock price responses for multiple tokens
	tokens := []Token{
		{Address: "0x1", Symbol: "TOKEN1", Chain: ChainEthereum},
		{Address: "0x2", Symbol: "TOKEN2", Chain: ChainEthereum},
	}

	for _, token := range tokens {
		uniswapQuote := &GetSwapQuoteResponse{
			AmountOut:   decimal.NewFromFloat(1000.0),
			PriceImpact: decimal.NewFromFloat(0.001),
		}
		oneInchQuote := &GetSwapQuoteResponse{
			AmountOut:   decimal.NewFromFloat(1015.0), // 1.5% higher
			PriceImpact: decimal.NewFromFloat(0.001),
		}

		mockUniswap.On("GetSwapQuote", ctx, mock.MatchedBy(func(req *GetSwapQuoteRequest) bool {
			return req.TokenIn == token.Address
		})).Return(uniswapQuote, nil)

		mockOneInch.On("GetSwapQuote", ctx, mock.MatchedBy(func(req *GetSwapQuoteRequest) bool {
			return req.TokenIn == token.Address
		})).Return(oneInchQuote, nil)
	}

	// Act & Assert
	for _, token := range tokens {
		opportunities, err := detector.DetectArbitrageForToken(ctx, token)
		require.NoError(t, err)
		assert.NotEmpty(t, opportunities, "Should find opportunities for token %s", token.Symbol)

		for _, opp := range opportunities {
			assert.True(t, opp.ProfitMargin.GreaterThan(decimal.Zero))
			assert.True(t, opp.NetProfit.GreaterThan(decimal.Zero))
			assert.NotEmpty(t, opp.ID)
		}
	}

	mockUniswap.AssertExpectations(t)
	mockOneInch.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}
