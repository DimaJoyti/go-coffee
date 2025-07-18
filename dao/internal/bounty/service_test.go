package bounty

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

// MockBountyRepository implements BountyRepository interface
type MockBountyRepository struct {
	mock.Mock
}

func (m *MockBountyRepository) Create(ctx context.Context, bounty *Bounty) error {
	args := m.Called(ctx, bounty)
	return args.Error(0)
}

func (m *MockBountyRepository) GetByID(ctx context.Context, bountyID uint64) (*Bounty, error) {
	args := m.Called(ctx, bountyID)
	return args.Get(0).(*Bounty), args.Error(1)
}

func (m *MockBountyRepository) List(ctx context.Context, filter *BountyFilter) ([]*Bounty, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*Bounty), args.Int(1), args.Error(2)
}

func (m *MockBountyRepository) Update(ctx context.Context, bounty *Bounty) error {
	args := m.Called(ctx, bounty)
	return args.Error(0)
}

func (m *MockBountyRepository) Delete(ctx context.Context, bountyID uint64) error {
	args := m.Called(ctx, bountyID)
	return args.Error(0)
}

func (m *MockBountyRepository) GetActive(ctx context.Context, limit, offset int) ([]*Bounty, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*Bounty), args.Error(1)
}

func (m *MockBountyRepository) GetCompleted(ctx context.Context, limit, offset int) ([]*Bounty, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*Bounty), args.Error(1)
}

func (m *MockBountyRepository) GetByCreator(ctx context.Context, creatorAddress string, limit, offset int) ([]*Bounty, error) {
	args := m.Called(ctx, creatorAddress, limit, offset)
	return args.Get(0).([]*Bounty), args.Error(1)
}

func (m *MockBountyRepository) GetByAssignee(ctx context.Context, assigneeAddress string, limit, offset int) ([]*Bounty, error) {
	args := m.Called(ctx, assigneeAddress, limit, offset)
	return args.Get(0).([]*Bounty), args.Error(1)
}

// MockMilestoneRepository implements MilestoneRepository interface
type MockMilestoneRepository struct {
	mock.Mock
}

func (m *MockMilestoneRepository) Create(ctx context.Context, milestone *Milestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *MockMilestoneRepository) GetByID(ctx context.Context, id int64) (*Milestone, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Milestone), args.Error(1)
}

func (m *MockMilestoneRepository) GetByBountyID(ctx context.Context, bountyID uint64) ([]*Milestone, error) {
	args := m.Called(ctx, bountyID)
	return args.Get(0).([]*Milestone), args.Error(1)
}

func (m *MockMilestoneRepository) GetByBountyAndIndex(ctx context.Context, bountyID uint64, index uint64) (*Milestone, error) {
	args := m.Called(ctx, bountyID, index)
	return args.Get(0).(*Milestone), args.Error(1)
}

func (m *MockMilestoneRepository) Update(ctx context.Context, milestone *Milestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *MockMilestoneRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockApplicationRepository implements ApplicationRepository interface
type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) Create(ctx context.Context, application *Application) error {
	args := m.Called(ctx, application)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetByID(ctx context.Context, id int64) (*Application, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Application), args.Error(1)
}

func (m *MockApplicationRepository) GetByBountyID(ctx context.Context, bountyID uint64) ([]*Application, error) {
	args := m.Called(ctx, bountyID)
	return args.Get(0).([]*Application), args.Error(1)
}

func (m *MockApplicationRepository) GetByApplicant(ctx context.Context, applicantAddress string, limit, offset int) ([]*Application, error) {
	args := m.Called(ctx, applicantAddress, limit, offset)
	return args.Get(0).([]*Application), args.Error(1)
}

func (m *MockApplicationRepository) HasApplied(ctx context.Context, bountyID uint64, applicantAddress string) (bool, error) {
	args := m.Called(ctx, bountyID, applicantAddress)
	return args.Bool(0), args.Error(1)
}

func (m *MockApplicationRepository) UpdateStatus(ctx context.Context, bountyID uint64, applicantAddress string, status ApplicationStatus) error {
	args := m.Called(ctx, bountyID, applicantAddress, status)
	return args.Error(0)
}

// MockBountyManagerClient implements BountyManagerClientInterface
type MockBountyManagerClient struct {
	mock.Mock
}

func (m *MockBountyManagerClient) CreateBounty(ctx context.Context, req *BountyRequest) (uint64, string, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(uint64), args.String(1), args.Error(2)
}

func (m *MockBountyManagerClient) AssignBounty(ctx context.Context, bountyID uint64, developer string) (string, error) {
	args := m.Called(ctx, bountyID, developer)
	return args.String(0), args.Error(1)
}

func (m *MockBountyManagerClient) CompleteMilestone(ctx context.Context, bountyID uint64, milestoneIndex uint64) (string, error) {
	args := m.Called(ctx, bountyID, milestoneIndex)
	return args.String(0), args.Error(1)
}

func (m *MockBountyManagerClient) VerifyPerformance(ctx context.Context, bountyID uint64, tvlImpact, mauImpact int64) (string, error) {
	args := m.Called(ctx, bountyID, tvlImpact, mauImpact)
	return args.String(0), args.Error(1)
}

// Test functions

func TestCreateBounty(t *testing.T) {
	// Setup mocks
	mockBountyRepo := &MockBountyRepository{}
	mockMilestoneRepo := &MockMilestoneRepository{}
	mockApplicationRepo := &MockApplicationRepository{}
	mockBountyManagerClient := &MockBountyManagerClient{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		bountyRepo:          mockBountyRepo,
		milestoneRepo:       mockMilestoneRepo,
		applicationRepo:     mockApplicationRepo,
		bountyManagerClient: mockBountyManagerClient,
		logger:              testLogger,
		bountyCache:         make(map[uint64]*Bounty),
		reputationCache:     make(map[string]int),
		performanceCache:    make(map[uint64]*PerformanceMetrics),
	}

	ctx := context.Background()
	req := &CreateBountyRequest{
		Title:          "Test Bounty",
		Description:    "Test bounty description",
		Category:       BountyCategoryTVLGrowth,
		CreatorAddress: "0x1234567890123456789012345678901234567890",
		TotalReward:    decimal.NewFromFloat(1000),
		Deadline:       time.Now().Add(30 * 24 * time.Hour),
		Milestones: []CreateMilestoneRequest{
			{
				Description: "Milestone 1",
				Reward:      decimal.NewFromFloat(500),
				Deadline:    time.Now().Add(15 * 24 * time.Hour),
			},
		},
	}

	// Setup mock expectations
	mockBountyManagerClient.On("CreateBounty", ctx, mock.AnythingOfType("*bounty.BountyRequest")).
		Return(uint64(1), "0xbounty123", nil)

	mockBountyRepo.On("Create", ctx, mock.AnythingOfType("*bounty.Bounty")).
		Return(nil)

	mockMilestoneRepo.On("Create", ctx, mock.AnythingOfType("*bounty.Milestone")).
		Return(nil)

	// Execute
	response, err := service.CreateBounty(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint64(1), response.BountyID)
	assert.Equal(t, "0xbounty123", response.TransactionHash)
	assert.Equal(t, BountyStatusOpen, response.Status)

	// Verify mock calls
	mockBountyManagerClient.AssertExpectations(t)
	mockBountyRepo.AssertExpectations(t)
	mockMilestoneRepo.AssertExpectations(t)
}

func TestGetBounty(t *testing.T) {
	// Setup mocks
	mockBountyRepo := &MockBountyRepository{}
	mockMilestoneRepo := &MockMilestoneRepository{}
	mockApplicationRepo := &MockApplicationRepository{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		bountyRepo:      mockBountyRepo,
		milestoneRepo:   mockMilestoneRepo,
		applicationRepo: mockApplicationRepo,
		logger:          testLogger,
		bountyCache:     make(map[uint64]*Bounty),
	}

	ctx := context.Background()
	bountyID := uint64(1)

	expectedBounty := &Bounty{
		ID:             1,
		BountyID:       bountyID,
		Title:          "Test Bounty",
		Description:    "Test description",
		Category:       BountyCategoryTVLGrowth,
		Status:         BountyStatusOpen,
		CreatorAddress: "0x1234567890123456789012345678901234567890",
		TotalReward:    decimal.NewFromFloat(1000),
		Deadline:       time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	expectedMilestones := []*Milestone{
		{
			ID:          1,
			BountyID:    bountyID,
			Index:       0,
			Description: "Milestone 1",
			Reward:      decimal.NewFromFloat(500),
			Deadline:    time.Now().Add(15 * 24 * time.Hour),
			Completed:   false,
			Paid:        false,
		},
	}

	expectedApplications := []*Application{
		{
			ID:                 1,
			BountyID:           bountyID,
			ApplicantAddress:   "0x9876543210987654321098765432109876543210",
			ApplicationMessage: "I want to work on this bounty",
			ProposedTimeline:   14,
			AppliedAt:          time.Now(),
			Status:             ApplicationStatusPending,
		},
	}

	// Setup mock expectations
	mockBountyRepo.On("GetByID", ctx, bountyID).Return(expectedBounty, nil)
	mockMilestoneRepo.On("GetByBountyID", ctx, bountyID).Return(expectedMilestones, nil)
	mockApplicationRepo.On("GetByBountyID", ctx, bountyID).Return(expectedApplications, nil)

	// Execute
	bountyDetails, err := service.GetBounty(ctx, bountyID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, bountyDetails)
	assert.Equal(t, expectedBounty, bountyDetails.Bounty)
	assert.Equal(t, expectedMilestones, bountyDetails.Milestones)
	assert.Equal(t, expectedApplications, bountyDetails.Applications)

	// Verify mock calls
	mockBountyRepo.AssertExpectations(t)
	mockMilestoneRepo.AssertExpectations(t)
	mockApplicationRepo.AssertExpectations(t)
}

func TestApplyForBounty(t *testing.T) {
	// Setup mocks
	mockBountyRepo := &MockBountyRepository{}
	mockApplicationRepo := &MockApplicationRepository{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		bountyRepo:      mockBountyRepo,
		applicationRepo: mockApplicationRepo,
		logger:          testLogger,
	}

	ctx := context.Background()
	req := &ApplyBountyRequest{
		BountyID:         1,
		ApplicantAddress: "0x9876543210987654321098765432109876543210",
		Message:          "I want to work on this bounty",
		ProposedTimeline: 14,
	}

	expectedBounty := &Bounty{
		ID:       1,
		BountyID: 1,
		Status:   BountyStatusOpen,
	}

	// Setup mock expectations
	mockBountyRepo.On("GetByID", ctx, req.BountyID).Return(expectedBounty, nil)
	mockApplicationRepo.On("HasApplied", ctx, req.BountyID, req.ApplicantAddress).Return(false, nil)
	mockApplicationRepo.On("Create", ctx, mock.AnythingOfType("*bounty.Application")).Return(nil)

	// Execute
	response, err := service.ApplyForBounty(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, ApplicationStatusPending, response.Status)

	// Verify mock calls
	mockBountyRepo.AssertExpectations(t)
	mockApplicationRepo.AssertExpectations(t)
}

func TestBountyStatusEnum(t *testing.T) {
	assert.Equal(t, "OPEN", BountyStatusOpen.String())
	assert.Equal(t, "ASSIGNED", BountyStatusAssigned.String())
	assert.Equal(t, "IN_PROGRESS", BountyStatusInProgress.String())
	assert.Equal(t, "SUBMITTED", BountyStatusSubmitted.String())
	assert.Equal(t, "COMPLETED", BountyStatusCompleted.String())
	assert.Equal(t, "CANCELLED", BountyStatusCancelled.String())
}

func TestBountyCategoryEnum(t *testing.T) {
	assert.Equal(t, "TVL_GROWTH", BountyCategoryTVLGrowth.String())
	assert.Equal(t, "MAU_EXPANSION", BountyCategoryMAUExpansion.String())
	assert.Equal(t, "INNOVATION", BountyCategoryInnovation.String())
	assert.Equal(t, "MAINTENANCE", BountyCategoryMaintenance.String())
	assert.Equal(t, "SECURITY", BountyCategorySecurity.String())
	assert.Equal(t, "INTEGRATION", BountyCategoryIntegration.String())
}

func TestApplicationStatusEnum(t *testing.T) {
	assert.Equal(t, "PENDING", ApplicationStatusPending.String())
	assert.Equal(t, "ACCEPTED", ApplicationStatusAccepted.String())
	assert.Equal(t, "REJECTED", ApplicationStatusRejected.String())
}
