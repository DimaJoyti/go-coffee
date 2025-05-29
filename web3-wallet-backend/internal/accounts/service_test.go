package accounts

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateAccount(ctx context.Context, account *Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockRepository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockRepository) GetAccountByUserID(ctx context.Context, userID string) (*Account, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockRepository) GetAccountByResetToken(ctx context.Context, token string) (*Account, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*Account), args.Error(1)
}

func (m *MockRepository) UpdateAccount(ctx context.Context, account *Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockRepository) DeleteAccount(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListAccounts(ctx context.Context, req *AccountListRequest) ([]Account, int, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]Account), args.Int(1), args.Error(2)
}

func (m *MockRepository) CreateSession(ctx context.Context, session *Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockRepository) UpdateSession(ctx context.Context, session *Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRepository) InvalidateAllSessions(ctx context.Context, accountID string) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

func (m *MockRepository) CreateSecurityEvent(ctx context.Context, event *SecurityEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockRepository) GetSecurityEvents(ctx context.Context, accountID string, limit int) ([]SecurityEvent, error) {
	args := m.Called(ctx, accountID, limit)
	return args.Get(0).([]SecurityEvent), args.Error(1)
}

func (m *MockRepository) CreateKYCDocument(ctx context.Context, doc *KYCDocument) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockRepository) GetKYCDocuments(ctx context.Context, accountID string) ([]KYCDocument, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]KYCDocument), args.Error(1)
}

func (m *MockRepository) UpdateKYCDocument(ctx context.Context, doc *KYCDocument) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockRepository) DeleteKYCDocument(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCache is a mock implementation of Redis client
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Test setup helper
func setupTestService() (*AccountService, *MockRepository, *MockCache) {
	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	
	cfg := config.AccountsConfig{
		MaxLoginAttempts: 5,
		AccountLimits: config.AccountLimits{
			DailyTransactionLimit:   "10000.00",
			MonthlyTransactionLimit: "100000.00",
			MaxWalletsPerUser:       10,
			MaxCardsPerUser:         5,
		},
		NotificationSettings: config.NotificationSettings{
			EmailEnabled:      true,
			SMSEnabled:        true,
			PushEnabled:       true,
			SecurityAlerts:    true,
			TransactionAlerts: true,
		},
	}

	logger := logger.New("debug", "json")
	service := NewService(mockRepo, cfg, logger, mockCache).(*AccountService)

	return service, mockRepo, mockCache
}

// Test CreateAccount
func TestCreateAccount(t *testing.T) {
	service, mockRepo, _ := setupTestService()
	ctx := context.Background()

	req := &CreateAccountRequest{
		Email:       "test@example.com",
		Phone:       "+1234567890",
		FirstName:   "Test",
		LastName:    "User",
		Password:    "password123",
		AccountType: AccountTypePersonal,
		Country:     "USA",
		AcceptTerms: true,
	}

	// Mock repository calls
	mockRepo.On("GetAccountByEmail", ctx, req.Email).Return(nil, assert.AnError)
	mockRepo.On("CreateAccount", ctx, mock.AnythingOfType("*accounts.Account")).Return(nil)
	mockRepo.On("CreateSecurityEvent", ctx, mock.AnythingOfType("*accounts.SecurityEvent")).Return(nil)

	// Execute
	account, err := service.CreateAccount(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, req.Email, account.Email)
	assert.Equal(t, req.FirstName, account.FirstName)
	assert.Equal(t, req.LastName, account.LastName)
	assert.Equal(t, AccountStatusPending, account.AccountStatus)
	assert.Equal(t, KYCStatusNotStarted, account.KYCStatus)

	mockRepo.AssertExpectations(t)
}

// Test CreateAccount with existing email
func TestCreateAccountExistingEmail(t *testing.T) {
	service, mockRepo, _ := setupTestService()
	ctx := context.Background()

	req := &CreateAccountRequest{
		Email:       "existing@example.com",
		Phone:       "+1234567890",
		FirstName:   "Test",
		LastName:    "User",
		Password:    "password123",
		AccountType: AccountTypePersonal,
		Country:     "USA",
		AcceptTerms: true,
	}

	existingAccount := &Account{
		ID:    uuid.New().String(),
		Email: req.Email,
	}

	// Mock repository calls
	mockRepo.On("GetAccountByEmail", ctx, req.Email).Return(existingAccount, nil)

	// Execute
	account, err := service.CreateAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, account)
	assert.Contains(t, err.Error(), "already exists")

	mockRepo.AssertExpectations(t)
}

// Test Login success
func TestLoginSuccess(t *testing.T) {
	service, mockRepo, _ := setupTestService()
	ctx := context.Background()

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
		DeviceID: "device123",
	}

	account := &Account{
		ID:               uuid.New().String(),
		Email:            req.Email,
		AccountStatus:    AccountStatusActive,
		FailedLoginCount: 0,
		Metadata: map[string]interface{}{
			"password_hash": "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9qK", // "password123"
		},
	}

	// Mock repository calls
	mockRepo.On("GetAccountByEmail", ctx, req.Email).Return(account, nil)
	mockRepo.On("UpdateAccount", ctx, mock.AnythingOfType("*accounts.Account")).Return(nil)
	mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*accounts.Session")).Return(nil)
	mockRepo.On("CreateSecurityEvent", ctx, mock.AnythingOfType("*accounts.SecurityEvent")).Return(nil)

	// Execute
	response, err := service.Login(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, account, response.Account)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "Bearer", response.TokenType)

	mockRepo.AssertExpectations(t)
}

// Test Login with invalid credentials
func TestLoginInvalidCredentials(t *testing.T) {
	service, mockRepo, _ := setupTestService()
	ctx := context.Background()

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
		DeviceID: "device123",
	}

	account := &Account{
		ID:               uuid.New().String(),
		Email:            req.Email,
		AccountStatus:    AccountStatusActive,
		FailedLoginCount: 0,
		Metadata: map[string]interface{}{
			"password_hash": "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9qK", // "password123"
		},
	}

	// Mock repository calls
	mockRepo.On("GetAccountByEmail", ctx, req.Email).Return(account, nil)
	mockRepo.On("UpdateAccount", ctx, mock.AnythingOfType("*accounts.Account")).Return(nil)
	mockRepo.On("CreateSecurityEvent", ctx, mock.AnythingOfType("*accounts.SecurityEvent")).Return(nil)

	// Execute
	response, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid credentials")

	mockRepo.AssertExpectations(t)
}
