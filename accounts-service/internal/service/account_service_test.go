package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/repository"
)

// MockAccountRepository is a mock implementation of repository.AccountRepository
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(ctx context.Context, account *models.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByUsername(ctx context.Context, username string) (*models.Account, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByEmail(ctx context.Context, email string) (*models.Account, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepository) List(ctx context.Context, offset, limit int) ([]*models.Account, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*models.Account), args.Error(1)
}

func (m *MockAccountRepository) Update(ctx context.Context, account *models.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAccountRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func TestAccountService_Create(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockAccountRepository)

	// Create service
	service := NewAccountService(mockRepo)

	// Test data
	ctx := context.Background()
	input := models.AccountInput{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: stringPtr("Test"),
		LastName:  stringPtr("User"),
	}

	// Setup expectations
	mockRepo.On("GetByUsername", ctx, "testuser").Return(nil, repository.ErrNotFound)
	mockRepo.On("GetByEmail", ctx, "test@example.com").Return(nil, repository.ErrNotFound)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Account")).Return(nil)

	// Call the service
	account, err := service.Create(ctx, input)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, "testuser", account.Username)
	assert.Equal(t, "test@example.com", account.Email)
	assert.Equal(t, "Test", account.FirstName)
	assert.Equal(t, "User", account.LastName)
	assert.True(t, account.IsActive)
	assert.False(t, account.IsAdmin)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestAccountService_GetByID(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockAccountRepository)

	// Create service
	service := NewAccountService(mockRepo)

	// Test data
	ctx := context.Background()
	id := uuid.New()
	expectedAccount := &models.Account{
		ID:        id,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, id).Return(expectedAccount, nil)

	// Call the service
	account, err := service.GetByID(ctx, id)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedAccount, account)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
