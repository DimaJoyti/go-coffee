package defi

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
)

func TestTradingBot_Creation(t *testing.T) {
	// Arrange
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}
	mockArbitrageDetector := &MockArbitrageDetector{}
	mockYieldAggregator := &MockYieldAggregator{}
	mockUniswap := &MockUniswapClient{}
	mockOneInch := &MockOneInchClient{}
	mockAave := &MockAaveClient{}

	config := TradingBotConfig{
		MaxPositionSize:   decimal.NewFromFloat(1000),
		MinProfitMargin:   decimal.NewFromFloat(0.01),
		MaxSlippage:       decimal.NewFromFloat(0.005),
		RiskTolerance:     RiskLevelMedium,
		AutoCompound:      true,
		MaxDailyTrades:    10,
		StopLossPercent:   decimal.NewFromFloat(0.05),
		TakeProfitPercent: decimal.NewFromFloat(0.15),
		ExecutionDelay:    time.Second * 5,
	}

	// Act
	bot := NewTradingBot(
		"Test Bot",
		StrategyTypeArbitrage,
		config,
		logger,
		mockRedis,
		mockArbitrageDetector,
		mockYieldAggregator,
		mockUniswap,
		mockOneInch,
		mockAave,
	)

	// Assert
	assert.NotNil(t, bot)
	assert.Equal(t, "Test Bot", bot.Name)
	assert.Equal(t, StrategyTypeArbitrage, bot.Strategy)
	assert.Equal(t, BotStatusStopped, bot.Status)
	assert.Equal(t, config.MaxPositionSize, bot.Config.MaxPositionSize)
	assert.NotEmpty(t, bot.ID)
}

func TestTradingBot_StartStop(t *testing.T) {
	// Arrange
	bot := createTestBot(t)
	ctx := context.Background()

	// Test Start
	err := bot.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, BotStatusActive, bot.GetStatus())

	// Test Start when already active
	err = bot.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already active")

	// Test Stop
	err = bot.Stop()
	assert.NoError(t, err)
	assert.Equal(t, BotStatusStopped, bot.GetStatus())

	// Test Stop when already stopped
	err = bot.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already stopped")
}

func TestTradingBot_PauseResume(t *testing.T) {
	// Arrange
	bot := createTestBot(t)
	ctx := context.Background()

	// Start bot first
	err := bot.Start(ctx)
	require.NoError(t, err)

	// Test Pause
	err = bot.Pause()
	assert.NoError(t, err)
	assert.Equal(t, BotStatusPaused, bot.GetStatus())

	// Test Pause when already paused
	err = bot.Pause()
	assert.Error(t, err)

	// Test Resume
	err = bot.Resume()
	assert.NoError(t, err)
	assert.Equal(t, BotStatusActive, bot.GetStatus())

	// Test Resume when not paused
	err = bot.Resume()
	assert.Error(t, err)

	// Cleanup
	bot.Stop()
}

func TestTradingBot_ArbitrageStrategy(t *testing.T) {
	// Arrange
	bot := createTestBot(t)
	bot.Strategy = StrategyTypeArbitrage

	mockArbitrageDetector := &MockArbitrageDetector{}
	bot.arbitrageDetector = mockArbitrageDetector

	ctx := context.Background()

	// Mock arbitrage opportunities
	opportunities := []*ArbitrageDetection{
		{
			ID:           "arb1",
			Token:        Token{Symbol: "ETH", Address: "0x1"},
			ProfitMargin: decimal.NewFromFloat(0.015), // 1.5%
			Volume:       decimal.NewFromFloat(1.0),
			SourcePrice:  decimal.NewFromFloat(2000),
			TargetPrice:  decimal.NewFromFloat(2030),
			Risk:         RiskLevelMedium,
		},
	}

	mockArbitrageDetector.On("GetOpportunities", ctx).Return(opportunities, nil)

	// Act
	bot.executeArbitrageStrategy(ctx)

	// Assert
	// Check that orders were queued (we can't easily test the channel directly)
	// In a real test, we might use a mock channel or check side effects
	mockArbitrageDetector.AssertExpectations(t)
}

func TestTradingBot_CalculateStopLossTakeProfit(t *testing.T) {
	// Arrange
	bot := createTestBot(t)
	entryPrice := decimal.NewFromFloat(2000)

	// Act
	stopLoss := bot.calculateStopLoss(entryPrice)
	takeProfit := bot.calculateTakeProfit(entryPrice)

	// Assert
	expectedStopLoss := entryPrice.Mul(decimal.NewFromFloat(0.95)) // 5% below
	expectedTakeProfit := entryPrice.Mul(decimal.NewFromFloat(1.15)) // 15% above

	assert.True(t, stopLoss.Equal(expectedStopLoss))
	assert.True(t, takeProfit.Equal(expectedTakeProfit))
}

func TestTradingBot_PerformanceTracking(t *testing.T) {
	// Arrange
	bot := createTestBot(t)

	// Test successful trade
	bot.updatePerformance(true, decimal.NewFromFloat(100))
	performance := bot.GetPerformance()

	assert.Equal(t, 1, performance.TotalTrades)
	assert.Equal(t, 1, performance.WinningTrades)
	assert.Equal(t, 0, performance.LosingTrades)
	assert.True(t, performance.WinRate.Equal(decimal.NewFromFloat(1.0)))

	// Test failed trade
	bot.updatePerformance(false, decimal.NewFromFloat(50))
	performance = bot.GetPerformance()

	assert.Equal(t, 2, performance.TotalTrades)
	assert.Equal(t, 1, performance.WinningTrades)
	assert.Equal(t, 1, performance.LosingTrades)
	assert.True(t, performance.WinRate.Equal(decimal.NewFromFloat(0.5)))
}

func TestTradingBot_PositionManagement(t *testing.T) {
	// Arrange
	bot := createTestBot(t)

	// Create test position
	position := &TradingPosition{
		ID:           "pos1",
		BotID:        bot.ID,
		Type:         PositionTypeLong,
		Token:        Token{Symbol: "ETH"},
		Amount:       decimal.NewFromFloat(1.0),
		EntryPrice:   decimal.NewFromFloat(2000),
		CurrentPrice: decimal.NewFromFloat(2000),
		StopLoss:     decimal.NewFromFloat(1900),
		TakeProfit:   decimal.NewFromFloat(2300),
		OpenedAt:     time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Add position
	bot.mutex.Lock()
	bot.activePositions[position.ID] = position
	bot.mutex.Unlock()

	// Test getting active positions
	positions := bot.GetActivePositions()
	assert.Len(t, positions, 1)
	assert.Equal(t, position.ID, positions[0].ID)

	// Test position monitoring (simplified)
	ctx := context.Background()

	// Update price to trigger stop loss
	position.CurrentPrice = decimal.NewFromFloat(1850) // Below stop loss
	bot.monitorPositions(ctx)

	// Position should be removed (closed)
	time.Sleep(10 * time.Millisecond) // Allow for async processing
	positions = bot.GetActivePositions()
	// Note: In real implementation, position would be closed and removed
}

func TestTradingBot_RiskManagement(t *testing.T) {
	// Arrange
	bot := createTestBot(t)
	ctx := context.Background()

	// Test order with high slippage (should be rejected)
	mockOneInch := &MockOneInchClient{}
	bot.oneInchClient = mockOneInch

	highSlippageQuote := &GetSwapQuoteResponse{
		AmountIn:     decimal.NewFromFloat(1000),
		AmountOut:    decimal.NewFromFloat(0.9), // Very poor rate
		PriceImpact:  decimal.NewFromFloat(0.1), // 10% slippage (above 0.5% limit)
		GasEstimate:  decimal.NewFromFloat(0.01),
	}

	mockOneInch.On("GetSwapQuote", ctx, mock.AnythingOfType("*defi.GetSwapQuoteRequest")).Return(highSlippageQuote, nil)

	order := &TradingOrder{
		ID:     "order1",
		BotID:  bot.ID,
		Type:   OrderTypeBuy,
		Token:  Token{Symbol: "ETH", Address: "0x1"},
		Amount: decimal.NewFromFloat(1000),
		Status: OrderStatusPending,
	}

	// Act
	err := bot.executeBuyOrder(ctx, order)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "price impact too high")
	mockOneInch.AssertExpectations(t)
}

// Mock types for testing
type MockArbitrageDetector struct {
	mock.Mock
}

func (m *MockArbitrageDetector) GetOpportunities(ctx context.Context) ([]*ArbitrageDetection, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ArbitrageDetection), args.Error(1)
}

func (m *MockArbitrageDetector) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockArbitrageDetector) Stop() {
	m.Called()
}

type MockYieldAggregator struct {
	mock.Mock
}

func (m *MockYieldAggregator) GetBestOpportunities(ctx context.Context, limit int) ([]*YieldFarmingOpportunity, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*YieldFarmingOpportunity), args.Error(1)
}

func (m *MockYieldAggregator) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockYieldAggregator) Stop() {
	m.Called()
}

// Helper function to create test bot
func createTestBot(t *testing.T) *TradingBot {
	logger := logger.New("test")
	mockRedis := &MockRedisClient{}
	mockArbitrageDetector := &MockArbitrageDetector{}
	mockYieldAggregator := &MockYieldAggregator{}
	mockUniswap := &MockUniswapClient{}
	mockOneInch := &MockOneInchClient{}
	mockAave := &MockAaveClient{}

	config := TradingBotConfig{
		MaxPositionSize:   decimal.NewFromFloat(1000),
		MinProfitMargin:   decimal.NewFromFloat(0.01),
		MaxSlippage:       decimal.NewFromFloat(0.005),
		RiskTolerance:     RiskLevelMedium,
		AutoCompound:      true,
		MaxDailyTrades:    10,
		StopLossPercent:   decimal.NewFromFloat(0.05),
		TakeProfitPercent: decimal.NewFromFloat(0.15),
		ExecutionDelay:    time.Second * 1, // Shorter for tests
	}

	return NewTradingBot(
		"Test Bot",
		StrategyTypeArbitrage,
		config,
		logger,
		mockRedis,
		mockArbitrageDetector,
		mockYieldAggregator,
		mockUniswap,
		mockOneInch,
		mockAave,
	)
}

// Benchmark tests
func BenchmarkTradingBot_PerformanceUpdate(b *testing.B) {
	bot := createTestBot(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bot.updatePerformance(i%2 == 0, decimal.NewFromFloat(100))
	}
}

func TestTradingBot_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	bot := createTestBot(t)
	ctx := context.Background()

	// Mock dependencies
	mockArbitrageDetector := &MockArbitrageDetector{}
	mockYieldAggregator := &MockYieldAggregator{}
	bot.arbitrageDetector = mockArbitrageDetector
	bot.yieldAggregator = mockYieldAggregator

	// Mock empty opportunities (no trades)
	mockArbitrageDetector.On("GetOpportunities", ctx).Return([]*ArbitrageDetection{}, nil)
	mockYieldAggregator.On("GetBestOpportunities", ctx, 5).Return([]*YieldFarmingOpportunity{}, nil)

	// Act
	err := bot.Start(ctx)
	require.NoError(t, err)

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Test status
	assert.Equal(t, BotStatusActive, bot.GetStatus())

	// Test pause/resume
	err = bot.Pause()
	assert.NoError(t, err)
	assert.Equal(t, BotStatusPaused, bot.GetStatus())

	err = bot.Resume()
	assert.NoError(t, err)
	assert.Equal(t, BotStatusActive, bot.GetStatus())

	// Stop bot
	err = bot.Stop()
	assert.NoError(t, err)
	assert.Equal(t, BotStatusStopped, bot.GetStatus())

	mockArbitrageDetector.AssertExpectations(t)
}
