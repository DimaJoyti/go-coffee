package marketplace

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

// MockSolutionRepository implements SolutionRepository interface
type MockSolutionRepository struct {
	mock.Mock
}

func (m *MockSolutionRepository) Create(ctx context.Context, solution *Solution) error {
	args := m.Called(ctx, solution)
	return args.Error(0)
}

func (m *MockSolutionRepository) GetByID(ctx context.Context, solutionID uint64) (*Solution, error) {
	args := m.Called(ctx, solutionID)
	return args.Get(0).(*Solution), args.Error(1)
}

func (m *MockSolutionRepository) List(ctx context.Context, filter *SolutionFilter) ([]*Solution, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*Solution), args.Int(1), args.Error(2)
}

func (m *MockSolutionRepository) Update(ctx context.Context, solution *Solution) error {
	args := m.Called(ctx, solution)
	return args.Error(0)
}

func (m *MockSolutionRepository) Delete(ctx context.Context, solutionID uint64) error {
	args := m.Called(ctx, solutionID)
	return args.Error(0)
}

func (m *MockSolutionRepository) GetByDeveloper(ctx context.Context, developerAddress string, limit, offset int) ([]*Solution, error) {
	args := m.Called(ctx, developerAddress, limit, offset)
	return args.Get(0).([]*Solution), args.Error(1)
}

func (m *MockSolutionRepository) GetByCategory(ctx context.Context, category SolutionCategory, limit, offset int) ([]*Solution, error) {
	args := m.Called(ctx, category, limit, offset)
	return args.Get(0).([]*Solution), args.Error(1)
}

func (m *MockSolutionRepository) GetPopular(ctx context.Context, limit, offset int) ([]*Solution, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*Solution), args.Error(1)
}

func (m *MockSolutionRepository) GetTrending(ctx context.Context, limit, offset int) ([]*Solution, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*Solution), args.Error(1)
}

func (m *MockSolutionRepository) GetRecentlyUpdated(ctx context.Context, since time.Duration, limit, offset int) ([]*Solution, error) {
	args := m.Called(ctx, since, limit, offset)
	return args.Get(0).([]*Solution), args.Error(1)
}

func (m *MockSolutionRepository) IncrementInstallCount(ctx context.Context, solutionID uint64) error {
	args := m.Called(ctx, solutionID)
	return args.Error(0)
}

// MockReviewRepository implements ReviewRepository interface
type MockReviewRepository struct {
	mock.Mock
}

func (m *MockReviewRepository) Create(ctx context.Context, review *Review) error {
	args := m.Called(ctx, review)
	return args.Error(0)
}

func (m *MockReviewRepository) GetByID(ctx context.Context, id int64) (*Review, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Review), args.Error(1)
}

func (m *MockReviewRepository) GetBySolutionID(ctx context.Context, solutionID uint64) ([]*Review, error) {
	args := m.Called(ctx, solutionID)
	return args.Get(0).([]*Review), args.Error(1)
}

func (m *MockReviewRepository) GetByReviewer(ctx context.Context, reviewerAddress string, limit, offset int) ([]*Review, error) {
	args := m.Called(ctx, reviewerAddress, limit, offset)
	return args.Get(0).([]*Review), args.Error(1)
}

func (m *MockReviewRepository) HasReviewed(ctx context.Context, solutionID uint64, reviewerAddress string) (bool, error) {
	args := m.Called(ctx, solutionID, reviewerAddress)
	return args.Bool(0), args.Error(1)
}

func (m *MockReviewRepository) Update(ctx context.Context, review *Review) error {
	args := m.Called(ctx, review)
	return args.Error(0)
}

func (m *MockReviewRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockInstallationRepository implements InstallationRepository interface
type MockInstallationRepository struct {
	mock.Mock
}

func (m *MockInstallationRepository) Create(ctx context.Context, installation *Installation) error {
	args := m.Called(ctx, installation)
	return args.Error(0)
}

func (m *MockInstallationRepository) GetByID(ctx context.Context, id int64) (*Installation, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Installation), args.Error(1)
}

func (m *MockInstallationRepository) GetBySolutionID(ctx context.Context, solutionID uint64) ([]*Installation, error) {
	args := m.Called(ctx, solutionID)
	return args.Get(0).([]*Installation), args.Error(1)
}

func (m *MockInstallationRepository) GetByInstaller(ctx context.Context, installerAddress string, limit, offset int) ([]*Installation, error) {
	args := m.Called(ctx, installerAddress, limit, offset)
	return args.Get(0).([]*Installation), args.Error(1)
}

func (m *MockInstallationRepository) Update(ctx context.Context, installation *Installation) error {
	args := m.Called(ctx, installation)
	return args.Error(0)
}

func (m *MockInstallationRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockSolutionRegistryClient implements SolutionRegistryClientInterface
type MockSolutionRegistryClient struct {
	mock.Mock
}

func (m *MockSolutionRegistryClient) SubmitSolution(ctx context.Context, req *SolutionSubmissionRequest) (uint64, string, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(uint64), args.String(1), args.Error(2)
}

func (m *MockSolutionRegistryClient) ApproveSolution(ctx context.Context, solutionID uint64) (string, error) {
	args := m.Called(ctx, solutionID)
	return args.String(0), args.Error(1)
}

func (m *MockSolutionRegistryClient) GetSolutionInfo(ctx context.Context, solutionID uint64) (*SolutionInfo, error) {
	args := m.Called(ctx, solutionID)
	return args.Get(0).(*SolutionInfo), args.Error(1)
}

// Test functions

func TestCreateSolution(t *testing.T) {
	// Setup mocks
	mockSolutionRepo := &MockSolutionRepository{}
	mockReviewRepo := &MockReviewRepository{}
	mockSolutionRegistryClient := &MockSolutionRegistryClient{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		solutionRepo:           mockSolutionRepo,
		reviewRepo:             mockReviewRepo,
		solutionRegistryClient: mockSolutionRegistryClient,
		logger:                 testLogger,
		solutionCache:          make(map[uint64]*Solution),
		qualityScoreCache:      make(map[uint64]*QualityScore),
		compatibilityCache:     make(map[string]*CompatibilityResult),
	}

	ctx := context.Background()
	req := &CreateSolutionRequest{
		Name:             "Test DeFi Solution",
		Description:      "A comprehensive DeFi analytics platform",
		Category:         SolutionCategoryDeFi,
		Version:          "1.0.0",
		DeveloperAddress: "0x1234567890123456789012345678901234567890",
		RepositoryURL:    "https://github.com/developer/defi-solution",
		DocumentationURL: "https://docs.defi-solution.com",
		DemoURL:          "https://demo.defi-solution.com",
		Tags:             []string{"defi", "analytics", "dashboard"},
	}

	// Setup mock expectations
	mockSolutionRegistryClient.On("SubmitSolution", ctx, mock.AnythingOfType("*marketplace.SolutionSubmissionRequest")).
		Return(uint64(1), "0xsolution123", nil)

	mockSolutionRepo.On("Create", ctx, mock.AnythingOfType("*marketplace.Solution")).
		Return(nil)

	// Mock for quality score calculation
	mockReviewRepo.On("GetBySolutionID", ctx, uint64(1)).Return([]*Review{}, nil)

	// Execute
	response, err := service.CreateSolution(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint64(1), response.SolutionID)
	assert.Equal(t, "0xsolution123", response.TransactionHash)
	assert.Equal(t, SolutionStatusPending, response.Status)

	// Verify mock calls
	mockSolutionRegistryClient.AssertExpectations(t)
	mockSolutionRepo.AssertExpectations(t)
}

func TestGetSolution(t *testing.T) {
	// Setup mocks
	mockSolutionRepo := &MockSolutionRepository{}
	mockReviewRepo := &MockReviewRepository{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		solutionRepo:      mockSolutionRepo,
		reviewRepo:        mockReviewRepo,
		logger:            testLogger,
		solutionCache:     make(map[uint64]*Solution),
		qualityScoreCache: make(map[uint64]*QualityScore),
	}

	ctx := context.Background()
	solutionID := uint64(1)

	expectedSolution := &Solution{
		ID:               1,
		SolutionID:       solutionID,
		Name:             "Test DeFi Solution",
		Description:      "A comprehensive DeFi analytics platform",
		Category:         SolutionCategoryDeFi,
		Version:          "1.0.0",
		DeveloperAddress: "0x1234567890123456789012345678901234567890",
		Status:           SolutionStatusApproved,
		RepositoryURL:    "https://github.com/developer/defi-solution",
		DocumentationURL: "https://docs.defi-solution.com",
		DemoURL:          "https://demo.defi-solution.com",
		Tags:             []string{"defi", "analytics", "dashboard"},
		InstallCount:     25,
		AverageRating:    decimal.NewFromFloat(4.5),
		ReviewCount:      8,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	expectedReviews := []*Review{
		{
			ID:                 1,
			SolutionID:         solutionID,
			ReviewerAddress:    "0x9876543210987654321098765432109876543210",
			Rating:             5,
			Comment:            "Excellent solution with great analytics features",
			SecurityScore:      5,
			PerformanceScore:   4,
			UsabilityScore:     5,
			DocumentationScore: 4,
			CreatedAt:          time.Now(),
		},
	}

	// Setup mock expectations
	mockSolutionRepo.On("GetByID", ctx, solutionID).Return(expectedSolution, nil)
	mockReviewRepo.On("GetBySolutionID", ctx, solutionID).Return(expectedReviews, nil)

	// Mock for installations (empty list)
	service.installationRepo = &MockInstallationRepository{}
	mockInstallationRepo := service.installationRepo.(*MockInstallationRepository)
	mockInstallationRepo.On("GetBySolutionID", ctx, solutionID).Return([]*Installation{}, nil)

	// Execute
	solutionDetails, err := service.GetSolution(ctx, solutionID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, solutionDetails)
	assert.Equal(t, expectedSolution, solutionDetails.Solution)
	assert.Equal(t, expectedReviews, solutionDetails.Reviews)

	// Verify mock calls
	mockSolutionRepo.AssertExpectations(t)
	mockReviewRepo.AssertExpectations(t)
}

func TestReviewSolution(t *testing.T) {
	// Setup mocks
	mockSolutionRepo := &MockSolutionRepository{}
	mockReviewRepo := &MockReviewRepository{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		solutionRepo:      mockSolutionRepo,
		reviewRepo:        mockReviewRepo,
		logger:            testLogger,
		qualityScoreCache: make(map[uint64]*QualityScore),
	}

	ctx := context.Background()
	req := &ReviewSolutionRequest{
		SolutionID:         1,
		ReviewerAddress:    "0x9876543210987654321098765432109876543210",
		Rating:             5,
		Comment:            "Excellent solution with great analytics features",
		SecurityScore:      5,
		PerformanceScore:   4,
		UsabilityScore:     5,
		DocumentationScore: 4,
	}

	expectedSolution := &Solution{
		ID:         1,
		SolutionID: 1,
		Status:     SolutionStatusApproved,
	}

	// Setup mock expectations
	mockSolutionRepo.On("GetByID", ctx, req.SolutionID).Return(expectedSolution, nil)
	mockReviewRepo.On("HasReviewed", ctx, req.SolutionID, req.ReviewerAddress).Return(false, nil)
	mockReviewRepo.On("Create", ctx, mock.AnythingOfType("*marketplace.Review")).Return(nil)
	mockReviewRepo.On("GetBySolutionID", ctx, req.SolutionID).Return([]*Review{}, nil)

	// Execute
	response, err := service.ReviewSolution(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Verify mock calls
	mockSolutionRepo.AssertExpectations(t)
	mockReviewRepo.AssertExpectations(t)
}

func TestSolutionStatusEnum(t *testing.T) {
	assert.Equal(t, "PENDING", SolutionStatusPending.String())
	assert.Equal(t, "APPROVED", SolutionStatusApproved.String())
	assert.Equal(t, "REJECTED", SolutionStatusRejected.String())
	assert.Equal(t, "DEPRECATED", SolutionStatusDeprecated.String())
}

func TestSolutionCategoryEnum(t *testing.T) {
	assert.Equal(t, "DEFI", SolutionCategoryDeFi.String())
	assert.Equal(t, "NFT", SolutionCategoryNFT.String())
	assert.Equal(t, "DAO", SolutionCategoryDAO.String())
	assert.Equal(t, "ANALYTICS", SolutionCategoryAnalytics.String())
	assert.Equal(t, "INFRASTRUCTURE", SolutionCategoryInfrastructure.String())
	assert.Equal(t, "SECURITY", SolutionCategorySecurity.String())
	assert.Equal(t, "UI", SolutionCategoryUI.String())
	assert.Equal(t, "INTEGRATION", SolutionCategoryIntegration.String())
}

func TestInstallationStatusEnum(t *testing.T) {
	assert.Equal(t, "ACTIVE", InstallationStatusActive.String())
	assert.Equal(t, "INACTIVE", InstallationStatusInactive.String())
	assert.Equal(t, "FAILED", InstallationStatusFailed.String())
}
