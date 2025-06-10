package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/container"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/mux"
)

// TestRunner provides a simple way to test the auth service
type TestRunner struct {
	container *container.Container
	server    *httptest.Server
	logger    *logger.Logger
}

// NewTestRunner creates a new test runner
func NewTestRunner() (*TestRunner, error) {
	// Create test configuration
	config := &container.Config{
		Database: &container.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "go_coffee_auth_test",
			Username: "postgres",
			Password: "postgres",
			SSLMode:  "disable",
			MaxConns: 10,
			MinConns: 2,
		},
		Logger: &logger.Config{
			Level: logger.InfoLevel,
		},
	}

	// Create container
	c, err := container.NewContainer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Create router and register routes
	router := mux.NewRouter()
	c.HTTPServer.GetHandler().RegisterRoutes(router)

	// Create test server
	server := httptest.NewServer(router)

	return &TestRunner{
		container: c,
		server:    server,
		logger:    c.Logger,
	}, nil
}

// Close closes the test runner
func (tr *TestRunner) Close() {
	if tr.server != nil {
		tr.server.Close()
	}
	if tr.container != nil {
		tr.container.Close()
	}
}

// GetBaseURL returns the base URL of the test server
func (tr *TestRunner) GetBaseURL() string {
	return tr.server.URL
}

// TestBasicFlow tests the basic authentication flow
func (tr *TestRunner) TestBasicFlow() error {
	tr.logger.Info("Starting basic auth flow test")

	// Test 1: Register a user
	registerReq := application.RegisterRequest{
		Email:    "test@example.com",
		Password: "TestPassword123!",
		Role:     "user",
	}

	registerResp, err := tr.makeRequest("POST", "/api/v1/auth/register", registerReq)
	if err != nil {
		return fmt.Errorf("register request failed: %w", err)
	}

	var registerResult application.RegisterResponse
	if err := json.Unmarshal(registerResp, &registerResult); err != nil {
		return fmt.Errorf("failed to unmarshal register response: %w", err)
	}

	if registerResult.User == nil {
		return fmt.Errorf("register response missing user")
	}

	if registerResult.AccessToken == "" {
		return fmt.Errorf("register response missing access token")
	}

	tr.logger.Info("✓ User registration successful")

	// Test 2: Login with the registered user
	loginReq := application.LoginRequest{
		Email:    registerReq.Email,
		Password: registerReq.Password,
	}

	loginResp, err := tr.makeRequest("POST", "/api/v1/auth/login", loginReq)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	var loginResult application.LoginResponse
	if err := json.Unmarshal(loginResp, &loginResult); err != nil {
		return fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	if loginResult.User == nil {
		return fmt.Errorf("login response missing user")
	}

	if loginResult.AccessToken == "" {
		return fmt.Errorf("login response missing access token")
	}

	tr.logger.Info("✓ User login successful")

	// Test 3: Validate token
	validateReq := application.ValidateTokenRequest{
		Token: loginResult.AccessToken,
	}

	validateResp, err := tr.makeRequest("POST", "/api/v1/auth/validate", validateReq)
	if err != nil {
		return fmt.Errorf("validate request failed: %w", err)
	}

	var validateResult application.ValidateTokenResponse
	if err := json.Unmarshal(validateResp, &validateResult); err != nil {
		return fmt.Errorf("failed to unmarshal validate response: %w", err)
	}

	if !validateResult.Valid {
		return fmt.Errorf("token validation failed")
	}

	tr.logger.Info("✓ Token validation successful")

	// Test 4: Get user info (protected endpoint)
	userInfoResp, err := tr.makeAuthenticatedRequest("GET", "/api/v1/auth/me", nil, loginResult.AccessToken)
	if err != nil {
		return fmt.Errorf("get user info request failed: %w", err)
	}

	var userInfoResult application.GetUserInfoResponse
	if err := json.Unmarshal(userInfoResp, &userInfoResult); err != nil {
		return fmt.Errorf("failed to unmarshal user info response: %w", err)
	}

	if userInfoResult.User == nil {
		return fmt.Errorf("user info response missing user")
	}

	tr.logger.Info("✓ Get user info successful")

	tr.logger.Info("🎉 All basic auth flow tests passed!")
	return nil
}

// TestErrorCases tests various error scenarios
func (tr *TestRunner) TestErrorCases() error {
	tr.logger.Info("Starting error cases test")

	// Test 1: Register with invalid email
	invalidRegisterReq := application.RegisterRequest{
		Email:    "invalid-email",
		Password: "TestPassword123!",
		Role:     "user",
	}

	_, err := tr.makeRequestExpectingError("POST", "/api/v1/auth/register", invalidRegisterReq, http.StatusBadRequest)
	if err != nil {
		return fmt.Errorf("invalid email test failed: %w", err)
	}

	tr.logger.Info("✓ Invalid email registration rejected")

	// Test 2: Login with wrong credentials
	wrongLoginReq := application.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "WrongPassword123!",
	}

	_, err = tr.makeRequestExpectingError("POST", "/api/v1/auth/login", wrongLoginReq, http.StatusUnauthorized)
	if err != nil {
		return fmt.Errorf("wrong credentials test failed: %w", err)
	}

	tr.logger.Info("✓ Wrong credentials login rejected")

	// Test 3: Validate invalid token
	invalidValidateReq := application.ValidateTokenRequest{
		Token: "invalid.jwt.token",
	}

	_, err = tr.makeRequestExpectingError("POST", "/api/v1/auth/validate", invalidValidateReq, http.StatusUnauthorized)
	if err != nil {
		return fmt.Errorf("invalid token test failed: %w", err)
	}

	tr.logger.Info("✓ Invalid token validation rejected")

	tr.logger.Info("🎉 All error case tests passed!")
	return nil
}

// Helper methods

func (tr *TestRunner) makeRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, tr.server.URL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	respBody := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			respBody = append(respBody, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return respBody, nil
}

func (tr *TestRunner) makeAuthenticatedRequest(method, path string, body interface{}, token string) ([]byte, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, tr.server.URL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	respBody := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			respBody = append(respBody, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return respBody, nil
}

func (tr *TestRunner) makeRequestExpectingError(method, path string, body interface{}, expectedStatus int) ([]byte, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, tr.server.URL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		return nil, fmt.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	respBody := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			respBody = append(respBody, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return respBody, nil
}
