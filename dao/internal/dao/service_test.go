package dao

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/config"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

type MockProposalRepository struct {
	mock.Mock
}

func (m *MockProposalRepository) Create(ctx context.Context, proposal *Proposal) error {
	args := m.Called(ctx, proposal)
	return args.Error(0)
}

func (m *MockProposalRepository) GetByID(ctx context.Context, proposalID string) (*Proposal, error) {
	args := m.Called(ctx, proposalID)
	return args.Get(0).(*Proposal), args.Error(1)
}

func (m *MockProposalRepository) List(ctx context.Context, filter *ProposalFilter) ([]*Proposal, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*Proposal), args.Int(1), args.Error(2)
}

func (m *MockProposalRepository) Update(ctx context.Context, proposal *Proposal) error {
	args := m.Called(ctx, proposal)
	return args.Error(0)
}

func (m *MockProposalRepository) Delete(ctx context.Context, proposalID string) error {
	args := m.Called(ctx, proposalID)
	return args.Error(0)
}

func (m *MockProposalRepository) GetStats(ctx context.Context) (*GovernanceStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*GovernanceStats), args.Error(1)
}

type MockDeveloperRepository struct {
	mock.Mock
}

func (m *MockDeveloperRepository) Create(ctx context.Context, profile *DeveloperProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockDeveloperRepository) GetByAddress(ctx context.Context, address string) (*DeveloperProfile, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(*DeveloperProfile), args.Error(1)
}

func (m *MockDeveloperRepository) GetActive(ctx context.Context, limit, offset int) ([]*DeveloperProfile, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*DeveloperProfile), args.Error(1)
}

func (m *MockDeveloperRepository) Update(ctx context.Context, profile *DeveloperProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockDeveloperRepository) UpdateActivity(ctx context.Context, address string) error {
	args := m.Called(ctx, address)
	return args.Error(0)
}

func (m *MockDeveloperRepository) GetTopByReputation(ctx context.Context, limit int) ([]*DeveloperProfile, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*DeveloperProfile), args.Error(1)
}

type MockVoteRepository struct {
	mock.Mock
}

func (m *MockVoteRepository) Create(ctx context.Context, vote *Vote) error {
	args := m.Called(ctx, vote)
	return args.Error(0)
}

func (m *MockVoteRepository) GetByProposal(ctx context.Context, proposalID string, limit, offset int) ([]*Vote, error) {
	args := m.Called(ctx, proposalID, limit, offset)
	return args.Get(0).([]*Vote), args.Error(1)
}

func (m *MockVoteRepository) GetByVoter(ctx context.Context, voterAddress string, limit, offset int) ([]*Vote, error) {
	args := m.Called(ctx, voterAddress, limit, offset)
	return args.Get(0).([]*Vote), args.Error(1)
}

func (m *MockVoteRepository) GetVoteCount(ctx context.Context, proposalID string) (int, error) {
	args := m.Called(ctx, proposalID)
	return args.Int(0), args.Error(1)
}

type MockGovernorClient struct {
	mock.Mock
}

func (m *MockGovernorClient) CreateProposal(ctx context.Context, req *GovernorProposalRequest) (string, string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockGovernorClient) CastVote(ctx context.Context, req *GovernorVoteRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockGovernorClient) GetVotingPower(ctx context.Context, address string) (decimal.Decimal, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockGovernorClient) GetProposalStatus(ctx context.Context, proposalID string) (ProposalStatus, error) {
	args := m.Called(ctx, proposalID)
	return args.Get(0).(ProposalStatus), args.Error(1)
}

func (m *MockGovernorClient) GetTotalSupply(ctx context.Context) (decimal.Decimal, error) {
	args := m.Called(ctx)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

// Ensure MockGovernorClient implements GovernorClientInterface
var _ GovernorClientInterface = (*MockGovernorClient)(nil)

// Test functions

func TestCreateProposal(t *testing.T) {
	// Setup mocks
	mockProposalRepo := &MockProposalRepository{}
	mockDeveloperRepo := &MockDeveloperRepository{}
	mockVoteRepo := &MockVoteRepository{}
	mockGovernorClient := &MockGovernorClient{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	// Create a minimal service for testing
	service := &Service{
		proposalRepo:     mockProposalRepo,
		developerRepo:    mockDeveloperRepo,
		voteRepo:         mockVoteRepo,
		governorClient:   mockGovernorClient,
		logger:           testLogger,
		config:           &config.Config{DAO: config.DAOConfig{ProposalThreshold: 10000}},
		votingPowerCache: make(map[string]decimal.Decimal),
		proposalCache:    make(map[string]*Proposal),
	}

	ctx := context.Background()
	req := &CreateProposalRequest{
		Title:           "Test Proposal",
		Description:     "Test Description",
		Category:        ProposalCategoryGeneral,
		ProposerAddress: "0x1234567890123456789012345678901234567890",
		Targets:         []string{},
		Values:          []string{},
		Calldatas:       []string{},
	}

	// Mock expectations
	votingPower := decimal.NewFromFloat(15000.0) // Above threshold
	mockGovernorClient.On("GetVotingPower", ctx, req.ProposerAddress).Return(votingPower, nil)

	proposalID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	txHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	mockGovernorClient.On("CreateProposal", ctx, mock.AnythingOfType("*dao.GovernorProposalRequest")).Return(proposalID, txHash, nil)

	mockProposalRepo.On("Create", ctx, mock.AnythingOfType("*dao.Proposal")).Return(nil)

	// Execute
	response, err := service.CreateProposal(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, proposalID, response.ProposalID)
	assert.Equal(t, txHash, response.TransactionHash)
	assert.Equal(t, ProposalStatusPending, response.Status)

	// Verify mocks
	mockGovernorClient.AssertExpectations(t)
	mockProposalRepo.AssertExpectations(t)
}

func TestGetProposal(t *testing.T) {
	// Setup mocks
	mockProposalRepo := &MockProposalRepository{}

	service := &Service{
		proposalRepo:  mockProposalRepo,
		proposalCache: make(map[string]*Proposal),
	}

	ctx := context.Background()
	proposalID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	expectedProposal := &Proposal{
		ID:              1,
		ProposalID:      proposalID,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Category:        ProposalCategoryGeneral,
		ProposerAddress: "0x1234567890123456789012345678901234567890",
		Status:          ProposalStatusActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Mock expectations
	mockProposalRepo.On("GetByID", ctx, proposalID).Return(expectedProposal, nil)

	// Execute
	proposal, err := service.GetProposal(ctx, proposalID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, proposal)
	assert.Equal(t, expectedProposal.ProposalID, proposal.ProposalID)
	assert.Equal(t, expectedProposal.Title, proposal.Title)

	// Verify mocks
	mockProposalRepo.AssertExpectations(t)
}

func TestGetVotingPower(t *testing.T) {
	// Setup mocks
	mockGovernorClient := &MockGovernorClient{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	service := &Service{
		governorClient:   mockGovernorClient,
		logger:           testLogger,
		votingPowerCache: make(map[string]decimal.Decimal),
	}

	ctx := context.Background()
	address := "0x1234567890123456789012345678901234567890"
	expectedPower := decimal.NewFromFloat(10000.0)

	// Mock expectations
	mockGovernorClient.On("GetVotingPower", ctx, address).Return(expectedPower, nil)

	// Execute
	power, err := service.GetVotingPower(ctx, address)

	// Assert
	assert.NoError(t, err)
	assert.True(t, expectedPower.Equal(power))

	// Verify mocks
	mockGovernorClient.AssertExpectations(t)
}

func TestVoteOnProposal(t *testing.T) {
	// Setup mocks
	mockVoteRepo := &MockVoteRepository{}
	mockGovernorClient := &MockGovernorClient{}

	// Create a test logger
	testLogger := logger.New("info", "json")

	service := &Service{
		voteRepo:         mockVoteRepo,
		governorClient:   mockGovernorClient,
		logger:           testLogger,
		votingPowerCache: make(map[string]decimal.Decimal),
	}

	ctx := context.Background()
	req := &VoteRequest{
		ProposalID:   "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		VoterAddress: "0x1234567890123456789012345678901234567890",
		Support:      VoteSupportFor,
		Reason:       "I support this proposal",
	}

	votingPower := decimal.NewFromFloat(5000.0)
	txHash := "0xvote1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	// Mock expectations
	mockGovernorClient.On("GetVotingPower", ctx, req.VoterAddress).Return(votingPower, nil)
	mockGovernorClient.On("CastVote", ctx, mock.AnythingOfType("*dao.GovernorVoteRequest")).Return(txHash, nil)
	mockVoteRepo.On("Create", ctx, mock.AnythingOfType("*dao.Vote")).Return(nil)

	// Execute
	response, err := service.VoteOnProposal(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, txHash, response.TransactionHash)
	assert.True(t, votingPower.Equal(response.VotingPower))

	// Verify mocks
	mockGovernorClient.AssertExpectations(t)
	mockVoteRepo.AssertExpectations(t)
}

func TestProposalStatusEnum(t *testing.T) {
	// Test ProposalStatus string conversion
	assert.Equal(t, "PENDING", ProposalStatusPending.String())
	assert.Equal(t, "ACTIVE", ProposalStatusActive.String())
	assert.Equal(t, "EXECUTED", ProposalStatusExecuted.String())
}

func TestVoteSupportEnum(t *testing.T) {
	// Test VoteSupport string conversion
	assert.Equal(t, "AGAINST", VoteSupportAgainst.String())
	assert.Equal(t, "FOR", VoteSupportFor.String())
	assert.Equal(t, "ABSTAIN", VoteSupportAbstain.String())
}

func TestProposalCategoryEnum(t *testing.T) {
	// Test ProposalCategory string conversion
	assert.Equal(t, "GENERAL", ProposalCategoryGeneral.String())
	assert.Equal(t, "BOUNTY", ProposalCategoryBounty.String())
	assert.Equal(t, "TREASURY", ProposalCategoryTreasury.String())
}
