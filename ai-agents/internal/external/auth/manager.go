package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// AuthManager manages authentication for external services
type AuthManager struct {
	providers map[string]AuthProvider
	storage   TokenStorage
	config    *Config
	mutex     sync.RWMutex
	
	// OAuth state management
	stateStore map[string]*OAuthState
	stateMutex sync.RWMutex
}

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	// Authentication
	GetAuthURL(state string, scopes []string) (string, error)
	ExchangeCode(ctx context.Context, code, state string) (*Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
	
	// Token validation
	ValidateToken(ctx context.Context, token *Token) error
	RevokeToken(ctx context.Context, token *Token) error
	
	// Provider info
	GetProviderInfo() *ProviderInfo
}

// TokenStorage defines the interface for token storage
type TokenStorage interface {
	StoreToken(ctx context.Context, userID, provider string, token *Token) error
	GetToken(ctx context.Context, userID, provider string) (*Token, error)
	DeleteToken(ctx context.Context, userID, provider string) error
	ListTokens(ctx context.Context, userID string) ([]*StoredToken, error)
	
	// Token rotation
	RotateToken(ctx context.Context, userID, provider string, newToken *Token) error
	
	// Cleanup
	CleanupExpiredTokens(ctx context.Context) error
}

// Config holds authentication configuration
type Config struct {
	// OAuth settings
	OAuth2Providers map[string]*OAuth2Config `yaml:"oauth2_providers" json:"oauth2_providers"`
	
	// API key settings
	APIKeyProviders map[string]*APIKeyConfig `yaml:"api_key_providers" json:"api_key_providers"`
	
	// Security settings
	StateTimeout    time.Duration            `yaml:"state_timeout" json:"state_timeout"`
	TokenEncryption bool                     `yaml:"token_encryption" json:"token_encryption"`
	EncryptionKey   string                   `yaml:"encryption_key" json:"encryption_key"`
	
	// Refresh settings
	AutoRefresh     bool                     `yaml:"auto_refresh" json:"auto_refresh"`
	RefreshBuffer   time.Duration            `yaml:"refresh_buffer" json:"refresh_buffer"`
	
	// Rate limiting
	RateLimit       *RateLimitConfig         `yaml:"rate_limit" json:"rate_limit"`
}

// OAuth2Config holds OAuth 2.0 configuration
type OAuth2Config struct {
	ClientID     string   `yaml:"client_id" json:"client_id"`
	ClientSecret string   `yaml:"client_secret" json:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url" json:"redirect_url"`
	AuthURL      string   `yaml:"auth_url" json:"auth_url"`
	TokenURL     string   `yaml:"token_url" json:"token_url"`
	Scopes       []string `yaml:"scopes" json:"scopes"`
	
	// Provider-specific settings
	ExtraParams  map[string]string `yaml:"extra_params" json:"extra_params"`
	
	// Security settings
	PKCE         bool     `yaml:"pkce" json:"pkce"`
	State        bool     `yaml:"state" json:"state"`
}

// APIKeyConfig holds API key configuration
type APIKeyConfig struct {
	APIKey       string            `yaml:"api_key" json:"api_key"`
	APIKeyHeader string            `yaml:"api_key_header" json:"api_key_header"`
	APIKeyParam  string            `yaml:"api_key_param" json:"api_key_param"`
	ExtraHeaders map[string]string `yaml:"extra_headers" json:"extra_headers"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int           `yaml:"requests_per_minute" json:"requests_per_minute"`
	BurstSize         int           `yaml:"burst_size" json:"burst_size"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// Token represents an authentication token
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       []string  `json:"scopes,omitempty"`
	
	// Provider-specific data
	Extra        map[string]interface{} `json:"extra,omitempty"`
	
	// Metadata
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StoredToken represents a stored token with metadata
type StoredToken struct {
	UserID       string    `json:"user_id"`
	Provider     string    `json:"provider"`
	Token        *Token    `json:"token"`
	StoredAt     time.Time `json:"stored_at"`
	LastUsed     time.Time `json:"last_used"`
	UsageCount   int       `json:"usage_count"`
}

// OAuthState represents OAuth state information
type OAuthState struct {
	State       string    `json:"state"`
	Provider    string    `json:"provider"`
	UserID      string    `json:"user_id"`
	RedirectURL string    `json:"redirect_url"`
	Scopes      []string  `json:"scopes"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	
	// PKCE support
	CodeVerifier  string `json:"code_verifier,omitempty"`
	CodeChallenge string `json:"code_challenge,omitempty"`
}

// ProviderInfo contains information about an auth provider
type ProviderInfo struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"` // oauth2, api_key
	AuthURL      string   `json:"auth_url,omitempty"`
	TokenURL     string   `json:"token_url,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	Capabilities []string `json:"capabilities"`
}

// AuthResult represents the result of an authentication attempt
type AuthResult struct {
	Success      bool      `json:"success"`
	Token        *Token    `json:"token,omitempty"`
	Error        string    `json:"error,omitempty"`
	RedirectURL  string    `json:"redirect_url,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *Config, storage TokenStorage) *AuthManager {
	return &AuthManager{
		providers:  make(map[string]AuthProvider),
		storage:    storage,
		config:     config,
		stateStore: make(map[string]*OAuthState),
	}
}

// RegisterProvider registers an authentication provider
func (am *AuthManager) RegisterProvider(name string, provider AuthProvider) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if _, exists := am.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}
	
	am.providers[name] = provider
	return nil
}

// GetProvider returns a provider by name
func (am *AuthManager) GetProvider(name string) (AuthProvider, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	provider, exists := am.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	
	return provider, nil
}

// StartOAuthFlow initiates an OAuth flow
func (am *AuthManager) StartOAuthFlow(ctx context.Context, provider, userID string, scopes []string, redirectURL string) (*AuthResult, error) {
	authProvider, err := am.GetProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	
	// Generate state
	state, err := am.generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	
	// Store state
	oauthState := &OAuthState{
		State:       state,
		Provider:    provider,
		UserID:      userID,
		RedirectURL: redirectURL,
		Scopes:      scopes,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(am.config.StateTimeout),
	}
	
	am.stateMutex.Lock()
	am.stateStore[state] = oauthState
	am.stateMutex.Unlock()
	
	// Get auth URL
	authURL, err := authProvider.GetAuthURL(state, scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth URL: %w", err)
	}
	
	return &AuthResult{
		Success:     true,
		RedirectURL: authURL,
		ExpiresAt:   oauthState.ExpiresAt,
	}, nil
}

// CompleteOAuthFlow completes an OAuth flow
func (am *AuthManager) CompleteOAuthFlow(ctx context.Context, code, state string) (*AuthResult, error) {
	// Validate state
	am.stateMutex.Lock()
	oauthState, exists := am.stateStore[state]
	if exists {
		delete(am.stateStore, state)
	}
	am.stateMutex.Unlock()
	
	if !exists {
		return &AuthResult{
			Success: false,
			Error:   "invalid or expired state",
		}, nil
	}
	
	if time.Now().After(oauthState.ExpiresAt) {
		return &AuthResult{
			Success: false,
			Error:   "state expired",
		}, nil
	}
	
	// Get provider
	provider, err := am.GetProvider(oauthState.Provider)
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Sprintf("provider not found: %v", err),
		}, nil
	}
	
	// Exchange code for token
	token, err := provider.ExchangeCode(ctx, code, state)
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Sprintf("failed to exchange code: %v", err),
		}, nil
	}
	
	// Store token
	err = am.storage.StoreToken(ctx, oauthState.UserID, oauthState.Provider, token)
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Sprintf("failed to store token: %v", err),
		}, nil
	}
	
	return &AuthResult{
		Success:   true,
		Token:     token,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

// GetToken retrieves a token for a user and provider
func (am *AuthManager) GetToken(ctx context.Context, userID, provider string) (*Token, error) {
	token, err := am.storage.GetToken(ctx, userID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	
	// Check if token needs refresh
	if am.config.AutoRefresh && am.needsRefresh(token) {
		refreshedToken, err := am.RefreshToken(ctx, userID, provider)
		if err != nil {
			// Return original token if refresh fails
			return token, nil
		}
		return refreshedToken, nil
	}
	
	return token, nil
}

// RefreshToken refreshes a token
func (am *AuthManager) RefreshToken(ctx context.Context, userID, provider string) (*Token, error) {
	// Get current token
	currentToken, err := am.storage.GetToken(ctx, userID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get current token: %w", err)
	}
	
	if currentToken.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}
	
	// Get provider
	authProvider, err := am.GetProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	
	// Refresh token
	newToken, err := authProvider.RefreshToken(ctx, currentToken.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	
	// Store new token
	err = am.storage.RotateToken(ctx, userID, provider, newToken)
	if err != nil {
		return nil, fmt.Errorf("failed to store refreshed token: %w", err)
	}
	
	return newToken, nil
}

// RevokeToken revokes a token
func (am *AuthManager) RevokeToken(ctx context.Context, userID, provider string) error {
	// Get token
	token, err := am.storage.GetToken(ctx, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	
	// Get provider
	authProvider, err := am.GetProvider(provider)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	
	// Revoke token with provider
	err = authProvider.RevokeToken(ctx, token)
	if err != nil {
		// Log error but continue with local deletion
		fmt.Printf("Failed to revoke token with provider: %v\n", err)
	}
	
	// Delete token from storage
	err = am.storage.DeleteToken(ctx, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	
	return nil
}

// ValidateToken validates a token
func (am *AuthManager) ValidateToken(ctx context.Context, userID, provider string) error {
	// Get token
	token, err := am.storage.GetToken(ctx, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	
	// Check expiration
	if time.Now().After(token.ExpiresAt) {
		return fmt.Errorf("token expired")
	}
	
	// Get provider
	authProvider, err := am.GetProvider(provider)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	
	// Validate with provider
	return authProvider.ValidateToken(ctx, token)
}

// GetAuthenticatedClient returns an HTTP client with authentication
func (am *AuthManager) GetAuthenticatedClient(ctx context.Context, userID, provider string) (*http.Client, error) {
	token, err := am.GetToken(ctx, userID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	
	// Create OAuth2 token
	oauth2Token := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.ExpiresAt,
	}
	
	// Get OAuth2 config for provider
	providerConfig, exists := am.config.OAuth2Providers[provider]
	if !exists {
		return nil, fmt.Errorf("OAuth2 config not found for provider %s", provider)
	}
	
	oauth2Config := &oauth2.Config{
		ClientID:     providerConfig.ClientID,
		ClientSecret: providerConfig.ClientSecret,
		RedirectURL:  providerConfig.RedirectURL,
		Scopes:       providerConfig.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  providerConfig.AuthURL,
			TokenURL: providerConfig.TokenURL,
		},
	}
	
	return oauth2Config.Client(ctx, oauth2Token), nil
}

// CleanupExpiredStates removes expired OAuth states
func (am *AuthManager) CleanupExpiredStates() {
	am.stateMutex.Lock()
	defer am.stateMutex.Unlock()
	
	now := time.Now()
	for state, oauthState := range am.stateStore {
		if now.After(oauthState.ExpiresAt) {
			delete(am.stateStore, state)
		}
	}
}

// CleanupExpiredTokens removes expired tokens
func (am *AuthManager) CleanupExpiredTokens(ctx context.Context) error {
	return am.storage.CleanupExpiredTokens(ctx)
}

// generateState generates a random state string
func (am *AuthManager) generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// needsRefresh checks if a token needs to be refreshed
func (am *AuthManager) needsRefresh(token *Token) bool {
	if token.RefreshToken == "" {
		return false
	}
	
	// Refresh if token expires within the buffer time
	refreshTime := token.ExpiresAt.Add(-am.config.RefreshBuffer)
	return time.Now().After(refreshTime)
}

// StartCleanupRoutine starts a routine to clean up expired states and tokens
func (am *AuthManager) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(am.config.StateTimeout / 2)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			am.CleanupExpiredStates()
			am.CleanupExpiredTokens(ctx)
		}
	}
}

// GetProviderList returns a list of all registered providers
func (am *AuthManager) GetProviderList() []*ProviderInfo {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	var providers []*ProviderInfo
	for _, provider := range am.providers {
		providers = append(providers, provider.GetProviderInfo())
	}
	
	return providers
}

// IsTokenValid checks if a token is valid without making external calls
func (am *AuthManager) IsTokenValid(token *Token) bool {
	if token == nil {
		return false
	}
	
	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return false
	}
	
	// Check if access token exists
	if token.AccessToken == "" {
		return false
	}
	
	return true
}
