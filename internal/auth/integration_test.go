package auth

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/cache"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/container"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/jwt"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/password"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/security"
	httpTransport "github.com/DimaJoyti/go-coffee/internal/auth/transport/http"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite provides integration testing for the auth service
type IntegrationTestSuite struct {
	suite.Suite
	container *container.Container
	db        *sql.DB
	server    *httptest.Server
}

// SetupSuite runs once before all tests
func (suite *IntegrationTestSuite) SetupSuite() {
	// Setup test database
	testDB, err := setupTestDatabase()
	require.NoError(suite.T(), err)
	suite.db = testDB

	// Create test configuration
	config := &container.Config{
		Database: &container.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "go_coffee_auth_test",
			Username: "postgres",
			Password: "postgres",
			SSLMode:  "disable",
			MaxConns: 5,
			MinConns: 1,
		},
		Redis: &cache.Config{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       1, // Use different DB for tests
			Prefix:   "auth_test",
			PoolSize: 5,
			Timeout:  5 * time.Second,
		},
		JWT: &jwt.Config{
			SecretKey:          "test-secret-key-for-integration-tests-only",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			Issuer:             "go-coffee-auth-test",
			Audience:           "go-coffee-api-test",
			RefreshTokenLength: 32,
		},
		Password: password.DefaultConfig(),
		Security: security.DefaultConfig(),
		HTTP: &httpTransport.Config{
			Port:         0, // Use random port for testing
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Logger: &logger.Config{
			Level: logger.ErrorLevel, // Reduce log noise in tests
		},
	}

	// Create container
	container, err := container.NewContainer(config)
	require.NoError(suite.T(), err)
	suite.container = container

	// Run database migrations
	err = suite.runMigrations()
	require.NoError(suite.T(), err)

	// Create router and register routes
	router := mux.NewRouter()
	suite.container.HTTPServer.GetHandler().RegisterRoutes(router)

	// Start test server
	suite.server = httptest.NewServer(router)
}

// TearDownSuite runs once after all tests
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}

	if suite.container != nil {
		suite.container.Close()
	}

	if suite.db != nil {
		// Clean up test database
		suite.cleanupTestDatabase()
		suite.db.Close()
	}
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up data before each test
	suite.cleanupTestData()
}

// TestUserRegistrationFlow tests the complete user registration flow
func (suite *IntegrationTestSuite) TestUserRegistrationFlow() {
	// Test user registration
	registerReq := application.RegisterRequest{
		Email:    "integration@test.com",
		Password: "TestPassword123!",
		Role:     "user",
	}

	registerResp := suite.makeRequest("POST", "/api/v1/auth/register", registerReq, http.StatusCreated)

	var registerResult application.RegisterResponse
	err := json.Unmarshal(registerResp, &registerResult)
	require.NoError(suite.T(), err)

	assert.NotNil(suite.T(), registerResult.User)
	assert.Equal(suite.T(), registerReq.Email, registerResult.User.Email)
	assert.NotEmpty(suite.T(), registerResult.User.ID)
	assert.NotEmpty(suite.T(), registerResult.AccessToken)

	// Test duplicate registration
	suite.makeRequest("POST", "/api/v1/auth/register", registerReq, http.StatusBadRequest)
}

// TestUserLoginFlow tests the complete user login flow
func (suite *IntegrationTestSuite) TestUserLoginFlow() {
	// First register a user
	registerReq := application.RegisterRequest{
		Email:    "login@test.com",
		Password: "TestPassword123!",
		Role:     "user",
	}
	suite.makeRequest("POST", "/api/v1/auth/register", registerReq, http.StatusCreated)

	// Test successful login
	loginReq := application.LoginRequest{
		Email:    registerReq.Email,
		Password: registerReq.Password,
	}

	loginResp := suite.makeRequest("POST", "/api/v1/auth/login", loginReq, http.StatusOK)

	var loginResult application.LoginResponse
	err := json.Unmarshal(loginResp, &loginResult)
	require.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), loginResult.AccessToken)
	assert.NotEmpty(suite.T(), loginResult.RefreshToken)
	assert.NotNil(suite.T(), loginResult.User)
	assert.Equal(suite.T(), registerReq.Email, loginResult.User.Email)

	// Test invalid login
	invalidLoginReq := application.LoginRequest{
		Email:    registerReq.Email,
		Password: "WrongPassword",
	}
	suite.makeRequest("POST", "/api/v1/auth/login", invalidLoginReq, http.StatusUnauthorized)
}

// TestTokenValidationFlow tests token validation
func (suite *IntegrationTestSuite) TestTokenValidationFlow() {
	// Register and login to get tokens
	accessToken := suite.registerAndLogin("token@test.com", "TestPassword123!")

	// Test token validation
	validateReq := application.ValidateTokenRequest{
		Token: accessToken,
	}

	validateResp := suite.makeRequest("POST", "/api/v1/auth/validate", validateReq, http.StatusOK)

	var validateResult application.ValidateTokenResponse
	err := json.Unmarshal(validateResp, &validateResult)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), validateResult.Valid)
	assert.NotEmpty(suite.T(), validateResult.UserID)

	// Test invalid token validation
	invalidValidateReq := application.ValidateTokenRequest{
		Token: "invalid.token.here",
	}
	suite.makeRequest("POST", "/api/v1/auth/validate", invalidValidateReq, http.StatusBadRequest)
}

// TestProtectedEndpoints tests accessing protected endpoints
func (suite *IntegrationTestSuite) TestProtectedEndpoints() {
	// Register and login to get tokens
	accessToken := suite.registerAndLogin("protected@test.com", "TestPassword123!")

	// Test accessing protected endpoint with valid token
	req, err := http.NewRequest("GET", suite.server.URL+"/api/v1/auth/me", nil)
	require.NoError(suite.T(), err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Test accessing protected endpoint without token
	req, err = http.NewRequest("GET", suite.server.URL+"/api/v1/auth/me", nil)
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
}

// TestPasswordChangeFlow tests password change functionality
func (suite *IntegrationTestSuite) TestPasswordChangeFlow() {
	// Register and login
	email := "password@test.com"
	oldPassword := "OldPassword123!"
	newPassword := "NewPassword123!"

	accessToken := suite.registerAndLogin(email, oldPassword)

	// Change password
	changeReq := application.ChangePasswordRequest{
		CurrentPassword: oldPassword,
		NewPassword:     newPassword,
	}

	req, err := http.NewRequest("POST", suite.server.URL+"/api/v1/auth/change-password", bytes.NewBuffer(suite.marshal(changeReq)))
	require.NoError(suite.T(), err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Test login with old password (should fail)
	loginReq := application.LoginRequest{
		Email:    email,
		Password: oldPassword,
	}
	suite.makeRequest("POST", "/api/v1/auth/login", loginReq, http.StatusUnauthorized)

	// Test login with new password (should succeed)
	loginReq.Password = newPassword
	suite.makeRequest("POST", "/api/v1/auth/login", loginReq, http.StatusOK)
}

// TestSessionManagement tests session management functionality
func (suite *IntegrationTestSuite) TestSessionManagement() {
	// Register and login to create a session
	accessToken := suite.registerAndLogin("session@test.com", "TestPassword123!")

	// Get user sessions
	req, err := http.NewRequest("GET", suite.server.URL+"/api/v1/auth/sessions", nil)
	require.NoError(suite.T(), err)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var sessionsResp application.GetUserSessionsResponse
	err = json.NewDecoder(resp.Body).Decode(&sessionsResp)
	require.NoError(suite.T(), err)

	assert.Greater(suite.T(), len(sessionsResp.Sessions), 0)
}

// TestHealthCheck tests the health check endpoint
func (suite *IntegrationTestSuite) TestHealthCheck() {
	resp := suite.makeRequest("GET", "/health", nil, http.StatusOK)

	var healthResp map[string]interface{}
	err := json.Unmarshal(resp, &healthResp)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "healthy", healthResp["status"])
	assert.Equal(suite.T(), "auth-service", healthResp["service"])
}

// Helper methods

func (suite *IntegrationTestSuite) registerAndLogin(email, password string) string {
	// Register user
	registerReq := application.RegisterRequest{
		Email:    email,
		Password: password,
		Role:     "user",
	}
	suite.makeRequest("POST", "/api/v1/auth/register", registerReq, http.StatusCreated)

	// Login user
	loginReq := application.LoginRequest{
		Email:    email,
		Password: password,
	}
	loginResp := suite.makeRequest("POST", "/api/v1/auth/login", loginReq, http.StatusOK)

	var loginResult application.LoginResponse
	err := json.Unmarshal(loginResp, &loginResult)
	require.NoError(suite.T(), err)

	return loginResult.AccessToken
}

func (suite *IntegrationTestSuite) makeRequest(method, path string, body interface{}, expectedStatus int) []byte {
	var reqBody []byte
	if body != nil {
		reqBody = suite.marshal(body)
	}

	req, err := http.NewRequest(method, suite.server.URL+path, bytes.NewBuffer(reqBody))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), expectedStatus, resp.StatusCode)

	respBody := make([]byte, 0)
	if resp.ContentLength != 0 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		respBody = buf.Bytes()
	}

	return respBody
}

func (suite *IntegrationTestSuite) marshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	require.NoError(suite.T(), err)
	return data
}

func (suite *IntegrationTestSuite) runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100),
			last_name VARCHAR(100),
			phone_number VARCHAR(20),
			role VARCHAR(20) NOT NULL DEFAULT 'customer',
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
			is_phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
			mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
			mfa_method VARCHAR(20) DEFAULT 'none',
			mfa_secret VARCHAR(255),
			mfa_backup_codes JSONB,
			failed_login_attempts INTEGER NOT NULL DEFAULT 0,
			last_failed_login TIMESTAMP WITH TIME ZONE,
			last_login_at TIMESTAMP WITH TIME ZONE,
			last_password_change TIMESTAMP WITH TIME ZONE,
			security_level VARCHAR(20) NOT NULL DEFAULT 'low',
			risk_score DECIMAL(3,2) NOT NULL DEFAULT 0.0,
			device_fingerprints JSONB,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			token_hash VARCHAR(255) NOT NULL,
			refresh_token_hash VARCHAR(255) NOT NULL,
			device_info JSONB,
			ip_address INET,
			user_agent TEXT,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			refresh_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			last_activity TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
	}

	for _, migration := range migrations {
		_, err := suite.db.Exec(migration)
		if err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}

func (suite *IntegrationTestSuite) cleanupTestData() {
	suite.db.Exec("DELETE FROM sessions")
	suite.db.Exec("DELETE FROM users")
}

func (suite *IntegrationTestSuite) cleanupTestDatabase() {
	suite.db.Exec("DROP TABLE IF EXISTS sessions")
	suite.db.Exec("DROP TABLE IF EXISTS users")
}

func setupTestDatabase() (*sql.DB, error) {
	// This would connect to a test database
	// In a real implementation, you might use Docker or a test database
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=go_coffee_auth_test sslmode=disable"
	return sql.Open("postgres", dsn)
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	// Skip integration tests if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(IntegrationTestSuite))
}
