// +build integration

package accounts

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

type AccountsIntegrationTestSuite struct {
	suite.Suite
	db          *sqlx.DB
	repo        Repository
	service     Service
	container   testcontainers.Container
	ctx         context.Context
}

func (suite *AccountsIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "test_fintech",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)
	suite.container = container

	// Get container host and port
	host, err := container.Host(suite.ctx)
	suite.Require().NoError(err)

	port, err := container.MappedPort(suite.ctx, "5432")
	suite.Require().NoError(err)

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=test_fintech sslmode=disable",
		host, port.Port())

	db, err := sqlx.Connect("postgres", dsn)
	suite.Require().NoError(err)
	suite.db = db

	// Create schema
	suite.createSchema()

	// Initialize repository and service
	suite.repo = NewPostgreSQLRepository(db)
	
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
	mockCache := &MockCache{}
	suite.service = NewService(suite.repo, cfg, logger, mockCache)
}

func (suite *AccountsIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
	}
}

func (suite *AccountsIntegrationTestSuite) SetupTest() {
	// Clean up tables before each test
	suite.db.MustExec("TRUNCATE TABLE accounts.accounts CASCADE")
	suite.db.MustExec("TRUNCATE TABLE accounts.sessions CASCADE")
	suite.db.MustExec("TRUNCATE TABLE accounts.security_events CASCADE")
	suite.db.MustExec("TRUNCATE TABLE accounts.kyc_documents CASCADE")
}

func (suite *AccountsIntegrationTestSuite) createSchema() {
	// Create accounts schema
	suite.db.MustExec(`CREATE SCHEMA IF NOT EXISTS accounts`)

	// Create accounts table
	suite.db.MustExec(`
		CREATE TABLE IF NOT EXISTS accounts.accounts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			phone VARCHAR(50),
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			date_of_birth DATE,
			nationality VARCHAR(3),
			country VARCHAR(3) NOT NULL,
			state VARCHAR(100),
			city VARCHAR(100),
			address TEXT,
			postal_code VARCHAR(20),
			account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('personal', 'business', 'enterprise')),
			account_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (account_status IN ('active', 'inactive', 'suspended', 'closed', 'pending')),
			kyc_status VARCHAR(20) NOT NULL DEFAULT 'not_started' CHECK (kyc_status IN ('not_started', 'pending', 'in_review', 'approved', 'rejected', 'expired')),
			kyc_level VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (kyc_level IN ('none', 'basic', 'standard', 'enhanced')),
			risk_score DECIMAL(3,2) DEFAULT 0.00,
			compliance_flags TEXT[],
			two_factor_enabled BOOLEAN DEFAULT FALSE,
			two_factor_method VARCHAR(20),
			last_login_at TIMESTAMP WITH TIME ZONE,
			last_login_ip INET,
			failed_login_count INTEGER DEFAULT 0,
			account_limits JSONB,
			notification_preferences JSONB,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`)

	// Create sessions table
	suite.db.MustExec(`
		CREATE TABLE IF NOT EXISTS accounts.sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			account_id UUID NOT NULL REFERENCES accounts.accounts(id) ON DELETE CASCADE,
			device_id VARCHAR(255),
			ip_address INET,
			user_agent TEXT,
			location VARCHAR(255),
			session_token VARCHAR(255) UNIQUE NOT NULL,
			refresh_token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`)

	// Create security events table
	suite.db.MustExec(`
		CREATE TABLE IF NOT EXISTS accounts.security_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			account_id UUID REFERENCES accounts.accounts(id) ON DELETE CASCADE,
			event_type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
			description TEXT NOT NULL,
			ip_address INET,
			user_agent TEXT,
			location VARCHAR(255),
			resolved BOOLEAN DEFAULT FALSE,
			resolved_at TIMESTAMP WITH TIME ZONE,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`)

	// Create KYC documents table
	suite.db.MustExec(`
		CREATE TABLE IF NOT EXISTS accounts.kyc_documents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			account_id UUID NOT NULL REFERENCES accounts.accounts(id) ON DELETE CASCADE,
			document_type VARCHAR(50) NOT NULL,
			document_url TEXT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
			uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			verified_at TIMESTAMP WITH TIME ZONE,
			expires_at TIMESTAMP WITH TIME ZONE,
			metadata JSONB DEFAULT '{}'
		)`)
}

func (suite *AccountsIntegrationTestSuite) TestCreateAndGetAccount() {
	req := &CreateAccountRequest{
		Email:       "integration@test.com",
		Phone:       "+1234567890",
		FirstName:   "Integration",
		LastName:    "Test",
		Password:    "password123",
		AccountType: AccountTypePersonal,
		Country:     "USA",
		AcceptTerms: true,
	}

	// Create account
	account, err := suite.service.CreateAccount(suite.ctx, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(account)

	// Verify account was created
	suite.Equal(req.Email, account.Email)
	suite.Equal(req.FirstName, account.FirstName)
	suite.Equal(req.LastName, account.LastName)
	suite.Equal(AccountStatusPending, account.AccountStatus)

	// Get account by ID
	retrievedAccount, err := suite.service.GetAccount(suite.ctx, account.ID)
	suite.Require().NoError(err)
	suite.Equal(account.ID, retrievedAccount.ID)
	suite.Equal(account.Email, retrievedAccount.Email)

	// Get account by email
	retrievedByEmail, err := suite.service.GetAccountByEmail(suite.ctx, account.Email)
	suite.Require().NoError(err)
	suite.Equal(account.ID, retrievedByEmail.ID)
}

func (suite *AccountsIntegrationTestSuite) TestLoginFlow() {
	// First create an account
	createReq := &CreateAccountRequest{
		Email:       "login@test.com",
		Phone:       "+1234567890",
		FirstName:   "Login",
		LastName:    "Test",
		Password:    "password123",
		AccountType: AccountTypePersonal,
		Country:     "USA",
		AcceptTerms: true,
	}

	account, err := suite.service.CreateAccount(suite.ctx, createReq)
	suite.Require().NoError(err)

	// Update account status to active
	account.AccountStatus = AccountStatusActive
	err = suite.repo.UpdateAccount(suite.ctx, account)
	suite.Require().NoError(err)

	// Test login
	loginReq := &LoginRequest{
		Email:    createReq.Email,
		Password: createReq.Password,
		DeviceID: "test-device",
	}

	loginResp, err := suite.service.Login(suite.ctx, loginReq)
	suite.Require().NoError(err)
	suite.Require().NotNil(loginResp)
	suite.NotEmpty(loginResp.AccessToken)
	suite.NotEmpty(loginResp.RefreshToken)

	// Test session validation
	session, err := suite.service.ValidateSession(suite.ctx, loginResp.AccessToken)
	suite.Require().NoError(err)
	suite.Equal(account.ID, session.AccountID)

	// Test logout
	err = suite.service.Logout(suite.ctx, loginResp.AccessToken)
	suite.Require().NoError(err)

	// Verify session is invalidated
	_, err = suite.service.ValidateSession(suite.ctx, loginResp.AccessToken)
	suite.Require().Error(err)
}

func (suite *AccountsIntegrationTestSuite) TestPasswordReset() {
	// Create account
	createReq := &CreateAccountRequest{
		Email:       "reset@test.com",
		Phone:       "+1234567890",
		FirstName:   "Reset",
		LastName:    "Test",
		Password:    "oldpassword123",
		AccountType: AccountTypePersonal,
		Country:     "USA",
		AcceptTerms: true,
	}

	account, err := suite.service.CreateAccount(suite.ctx, createReq)
	suite.Require().NoError(err)

	// Request password reset
	resetReq := &ResetPasswordRequest{
		Email: createReq.Email,
	}

	err = suite.service.ResetPassword(suite.ctx, resetReq)
	suite.Require().NoError(err)

	// Get account to retrieve reset token
	updatedAccount, err := suite.repo.GetAccountByID(suite.ctx, account.ID)
	suite.Require().NoError(err)

	resetToken, ok := updatedAccount.Metadata["password_reset_token"].(string)
	suite.Require().True(ok)
	suite.NotEmpty(resetToken)

	// Confirm password reset
	confirmReq := &ConfirmPasswordResetRequest{
		Token:       resetToken,
		NewPassword: "newpassword123",
	}

	err = suite.service.ConfirmPasswordReset(suite.ctx, confirmReq)
	suite.Require().NoError(err)

	// Verify old password doesn't work
	loginReq := &LoginRequest{
		Email:    createReq.Email,
		Password: "oldpassword123",
		DeviceID: "test-device",
	}

	updatedAccount.AccountStatus = AccountStatusActive
	err = suite.repo.UpdateAccount(suite.ctx, updatedAccount)
	suite.Require().NoError(err)

	_, err = suite.service.Login(suite.ctx, loginReq)
	suite.Require().Error(err)

	// Verify new password works
	loginReq.Password = "newpassword123"
	_, err = suite.service.Login(suite.ctx, loginReq)
	suite.Require().NoError(err)
}

func (suite *AccountsIntegrationTestSuite) TestKYCDocuments() {
	// Create account
	createReq := &CreateAccountRequest{
		Email:       "kyc@test.com",
		Phone:       "+1234567890",
		FirstName:   "KYC",
		LastName:    "Test",
		Password:    "password123",
		AccountType: AccountTypePersonal,
		Country:     "USA",
		AcceptTerms: true,
	}

	account, err := suite.service.CreateAccount(suite.ctx, createReq)
	suite.Require().NoError(err)

	// Submit KYC document
	kycReq := &KYCSubmissionRequest{
		DocumentType: DocumentTypePassport,
		DocumentURL:  "https://example.com/passport.jpg",
	}

	doc, err := suite.service.SubmitKYCDocument(suite.ctx, account.ID, kycReq)
	suite.Require().NoError(err)
	suite.Equal(account.ID, doc.AccountID)
	suite.Equal(DocumentTypePassport, doc.DocumentType)
	suite.Equal(DocumentStatusPending, doc.Status)

	// Get KYC documents
	docs, err := suite.service.GetKYCDocuments(suite.ctx, account.ID)
	suite.Require().NoError(err)
	suite.Len(docs, 1)
	suite.Equal(doc.ID, docs[0].ID)

	// Update KYC status
	err = suite.service.UpdateKYCStatus(suite.ctx, account.ID, KYCStatusApproved, KYCLevelStandard)
	suite.Require().NoError(err)

	// Verify status was updated
	updatedAccount, err := suite.service.GetAccount(suite.ctx, account.ID)
	suite.Require().NoError(err)
	suite.Equal(KYCStatusApproved, updatedAccount.KYCStatus)
	suite.Equal(KYCLevelStandard, updatedAccount.KYCLevel)
}

func TestAccountsIntegrationSuite(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests. Set INTEGRATION_TESTS=1 to run.")
	}

	suite.Run(t, new(AccountsIntegrationTestSuite))
}
