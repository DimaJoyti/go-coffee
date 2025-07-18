package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

// MockTVLRepository implements TVLRepository interface
type MockTVLRepository struct {
	mock.Mock
}

func (m *MockTVLRepository) Create(ctx context.Context, record *TVLRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockTVLRepository) GetByID(ctx context.Context, id int64) (*TVLRecord, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*TVLRecord), args.Error(1)
}

func (m *MockTVLRepository) GetByProtocol(ctx context.Context, protocol string, limit, offset int) ([]*TVLRecord, error) {
	args := m.Called(ctx, protocol, limit, offset)
	return args.Get(0).([]*TVLRecord), args.Error(1)
}

func (m *MockTVLRepository) GetByChain(ctx context.Context, chain string, limit, offset int) ([]*TVLRecord, error) {
	args := m.Called(ctx, chain, limit, offset)
	return args.Get(0).([]*TVLRecord), args.Error(1)
}

func (m *MockTVLRepository) GetHistory(ctx context.Context, protocol, chain string, since time.Time, limit int) ([]*TVLRecord, error) {
	args := m.Called(ctx, protocol, chain, since, limit)
	return args.Get(0).([]*TVLRecord), args.Error(1)
}

func (m *MockTVLRepository) GetLatest(ctx context.Context, protocol, chain string) (*TVLRecord, error) {
	args := m.Called(ctx, protocol, chain)
	return args.Get(0).(*TVLRecord), args.Error(1)
}

func (m *MockTVLRepository) GetAggregated(ctx context.Context, period string, since time.Time) ([]*AggregatedMetrics, error) {
	args := m.Called(ctx, period, since)
	return args.Get(0).([]*AggregatedMetrics), args.Error(1)
}

// MockMAURepository implements MAURepository interface
type MockMAURepository struct {
	mock.Mock
}

func (m *MockMAURepository) Create(ctx context.Context, record *MAURecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockMAURepository) GetByID(ctx context.Context, id int64) (*MAURecord, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*MAURecord), args.Error(1)
}

func (m *MockMAURepository) GetByFeature(ctx context.Context, feature string, limit, offset int) ([]*MAURecord, error) {
	args := m.Called(ctx, feature, limit, offset)
	return args.Get(0).([]*MAURecord), args.Error(1)
}

func (m *MockMAURepository) GetHistory(ctx context.Context, feature string, since time.Time, limit int) ([]*MAURecord, error) {
	args := m.Called(ctx, feature, since, limit)
	return args.Get(0).([]*MAURecord), args.Error(1)
}

func (m *MockMAURepository) GetLatest(ctx context.Context, feature string) (*MAURecord, error) {
	args := m.Called(ctx, feature)
	return args.Get(0).(*MAURecord), args.Error(1)
}

func (m *MockMAURepository) GetAggregated(ctx context.Context, period string, since time.Time) ([]*AggregatedMetrics, error) {
	args := m.Called(ctx, period, since)
	return args.Get(0).([]*AggregatedMetrics), args.Error(1)
}

// MockDefiLlamaClient implements DefiLlamaClientInterface
type MockDefiLlamaClient struct {
	mock.Mock
}

func (m *MockDefiLlamaClient) GetProtocolTVL(ctx context.Context, protocol string) (decimal.Decimal, error) {
	args := m.Called(ctx, protocol)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockDefiLlamaClient) GetChainTVL(ctx context.Context, chain string) (decimal.Decimal, error) {
	args := m.Called(ctx, chain)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockDefiLlamaClient) GetHistoricalTVL(ctx context.Context, protocol string, days int) ([]TVLDataPoint, error) {
	args := m.Called(ctx, protocol, days)
	return args.Get(0).([]TVLDataPoint), args.Error(1)
}

func (m *MockDefiLlamaClient) GetProtocolList(ctx context.Context) ([]ProtocolInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]ProtocolInfo), args.Error(1)
}

// Test functions

func TestRecordTVL(t *testing.T) {
	// Setup mocks
	mockTVLRepo := &MockTVLRepository{}
	mockDefiLlamaClient := &MockDefiLlamaClient{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		tvlRepo:         mockTVLRepo,
		defiLlamaClient: mockDefiLlamaClient,
		logger:          testLogger,
		metricsCache:    make(map[string]*MetricsSnapshot),
	}

	ctx := context.Background()
	req := &RecordTVLRequest{
		Protocol:    "go-coffee-defi",
		Chain:       "ethereum",
		Amount:      decimal.NewFromFloat(5000000),
		TokenSymbol: "USDC",
		Source:      "defi-llama",
		BlockNumber: func() *int64 { n := int64(18500000); return &n }(),
		TxHash:      "0x1234567890abcdef",
	}

	// Setup mock expectations
	mockTVLRepo.On("Create", ctx, mock.AnythingOfType("*metrics.TVLRecord")).
		Return(nil)

	// Execute
	response, err := service.RecordTVL(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "recorded", response.Status)

	// Verify mock calls
	mockTVLRepo.AssertExpectations(t)
}

func TestRecordMAU(t *testing.T) {
	// Setup mocks
	mockMAURepo := &MockMAURepository{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		mauRepo:      mockMAURepo,
		logger:       testLogger,
		metricsCache: make(map[string]*MetricsSnapshot),
	}

	ctx := context.Background()
	req := &RecordMAURequest{
		Feature:     "defi_trading",
		UserCount:   25000,
		UniqueUsers: 22500,
		Period:      "monthly",
		Source:      "analytics",
	}

	// Setup mock expectations
	mockMAURepo.On("Create", ctx, mock.AnythingOfType("*metrics.MAURecord")).
		Return(nil)

	// Execute
	response, err := service.RecordMAU(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "recorded", response.Status)

	// Verify mock calls
	mockMAURepo.AssertExpectations(t)
}

func TestGetTVLMetrics(t *testing.T) {
	// Setup mocks
	mockDefiLlamaClient := &MockDefiLlamaClient{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		defiLlamaClient: mockDefiLlamaClient,
		logger:          testLogger,
		metricsCache:    make(map[string]*MetricsSnapshot),
	}

	ctx := context.Background()
	req := &GetTVLMetricsRequest{
		Protocol: "go-coffee-defi",
		Chain:    "ethereum",
		Period:   "daily",
	}

	// Execute
	response, err := service.GetTVLMetrics(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, decimal.NewFromFloat(5000000), response.CurrentTVL)
	assert.Equal(t, decimal.NewFromFloat(2.5), response.Growth24h)
	assert.Equal(t, decimal.NewFromFloat(15.2), response.Growth7d)
}

func TestGetMAUMetrics(t *testing.T) {
	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		logger:       testLogger,
		metricsCache: make(map[string]*MetricsSnapshot),
	}

	ctx := context.Background()
	req := &GetMAUMetricsRequest{
		Feature: "defi_trading",
		Period:  "monthly",
	}

	// Execute
	response, err := service.GetMAUMetrics(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 25000, response.CurrentMAU)
	assert.Equal(t, decimal.NewFromFloat(8.5), response.Growth30d)
	assert.Equal(t, decimal.NewFromFloat(75.2), response.Retention)
}

func TestValidateTVLRequest(t *testing.T) {
	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		logger: testLogger,
	}

	// Test valid request
	validReq := &RecordTVLRequest{
		Protocol: "go-coffee-defi",
		Chain:    "ethereum",
		Amount:   decimal.NewFromFloat(5000000),
		Source:   "defi-llama",
	}

	err := service.validateTVLRequest(validReq)
	assert.NoError(t, err)

	// Test invalid request - missing protocol
	invalidReq := &RecordTVLRequest{
		Chain:  "ethereum",
		Amount: decimal.NewFromFloat(5000000),
		Source: "defi-llama",
	}

	err = service.validateTVLRequest(invalidReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "protocol is required")

	// Test invalid request - negative amount
	negativeAmountReq := &RecordTVLRequest{
		Protocol: "go-coffee-defi",
		Chain:    "ethereum",
		Amount:   decimal.NewFromFloat(-1000),
		Source:   "defi-llama",
	}

	err = service.validateTVLRequest(negativeAmountReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be positive")
}

func TestValidateMAURequest(t *testing.T) {
	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		logger: testLogger,
	}

	// Test valid request
	validReq := &RecordMAURequest{
		Feature:   "defi_trading",
		UserCount: 25000,
		Period:    "monthly",
		Source:    "analytics",
	}

	err := service.validateMAURequest(validReq)
	assert.NoError(t, err)

	// Test invalid request - missing feature
	invalidReq := &RecordMAURequest{
		UserCount: 25000,
		Period:    "monthly",
		Source:    "analytics",
	}

	err = service.validateMAURequest(invalidReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "feature is required")

	// Test invalid request - negative user count
	negativeUserCountReq := &RecordMAURequest{
		Feature:   "defi_trading",
		UserCount: -100,
		Period:    "monthly",
		Source:    "analytics",
	}

	err = service.validateMAURequest(negativeUserCountReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user count must be non-negative")
}

func TestMetricTypeEnum(t *testing.T) {
	assert.Equal(t, "TVL", MetricTypeTVL.String())
	assert.Equal(t, "MAU", MetricTypeMAU.String())
	assert.Equal(t, "REVENUE", MetricTypeRevenue.String())
	assert.Equal(t, "TRANSACTIONS", MetricTypeTransactions.String())
	assert.Equal(t, "USERS", MetricTypeUsers.String())
}

func TestAlertTypeEnum(t *testing.T) {
	assert.Equal(t, "THRESHOLD", AlertTypeThreshold.String())
	assert.Equal(t, "GROWTH_RATE", AlertTypeGrowthRate.String())
	assert.Equal(t, "ANOMALY", AlertTypeAnomaly.String())
	assert.Equal(t, "DOWNTIME", AlertTypeDowntime.String())
}

func TestAlertStatusEnum(t *testing.T) {
	assert.Equal(t, "ACTIVE", AlertStatusActive.String())
	assert.Equal(t, "RESOLVED", AlertStatusResolved.String())
	assert.Equal(t, "SUPPRESSED", AlertStatusSuppressed.String())
}

func TestCalculateGrowthRate(t *testing.T) {
	current := decimal.NewFromFloat(1100)
	previous := decimal.NewFromFloat(1000)
	
	growth := CalculateGrowthRate(current, previous)
	expected := decimal.NewFromFloat(10) // 10% growth
	
	assert.True(t, growth.Equal(expected))
	
	// Test with zero previous value
	zeroGrowth := CalculateGrowthRate(current, decimal.Zero)
	assert.True(t, zeroGrowth.Equal(decimal.Zero))
}

func TestFormatMetricValue(t *testing.T) {
	// Test TVL formatting
	tvlValue := decimal.NewFromFloat(5000000.50)
	formatted := FormatMetricValue(tvlValue, MetricTypeTVL)
	assert.Equal(t, "$5000000.50", formatted)
	
	// Test MAU formatting
	mauValue := decimal.NewFromFloat(25000.75)
	formatted = FormatMetricValue(mauValue, MetricTypeMAU)
	assert.Equal(t, "25001", formatted) // Rounded to integer
	
	// Test default formatting
	defaultValue := decimal.NewFromFloat(123.456)
	formatted = FormatMetricValue(defaultValue, MetricType(999))
	assert.Equal(t, "123.46", formatted)
}
