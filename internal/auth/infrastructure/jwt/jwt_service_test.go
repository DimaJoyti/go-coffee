package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateAndValidateTokens(t *testing.T) {
	// Setup
	config := &Config{
		SecretKey:          "test-secret-key-that-is-long-enough-for-testing-purposes",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
		Audience:           "test-audience",
		RefreshTokenLength: 32,
	}

	logger := logger.New("test")
	jwtService := NewJWTService(config, logger)

	ctx := context.Background()
	user := &domain.User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  domain.UserRoleUser,
	}
	sessionID := "test-session-id"

	t.Run("GenerateAccessToken", func(t *testing.T) {
		token, claims, err := jwtService.GenerateAccessToken(ctx, user, sessionID)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, claims)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Role, claims.Role)
		assert.Equal(t, sessionID, claims.SessionID)
		assert.Equal(t, domain.TokenTypeAccess, claims.Type)
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		token, claims, err := jwtService.GenerateRefreshToken(ctx, user, sessionID)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, claims)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Role, claims.Role)
		assert.Equal(t, sessionID, claims.SessionID)
		assert.Equal(t, domain.TokenTypeRefresh, claims.Type)
	})

	t.Run("GenerateTokenPair", func(t *testing.T) {
		accessToken, refreshToken, accessClaims, refreshClaims, err := jwtService.GenerateTokenPair(ctx, user, sessionID)

		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NotNil(t, accessClaims)
		assert.NotNil(t, refreshClaims)

		assert.Equal(t, domain.TokenTypeAccess, accessClaims.Type)
		assert.Equal(t, domain.TokenTypeRefresh, refreshClaims.Type)
		assert.Equal(t, user.ID, accessClaims.UserID)
		assert.Equal(t, user.ID, refreshClaims.UserID)
	})

	t.Run("ValidateToken", func(t *testing.T) {
		// Generate a token first
		token, originalClaims, err := jwtService.GenerateAccessToken(ctx, user, sessionID)
		require.NoError(t, err)

		// Validate the token
		claims, err := jwtService.ValidateToken(ctx, token)
		require.NoError(t, err)

		assert.Equal(t, originalClaims.UserID, claims.UserID)
		assert.Equal(t, originalClaims.Email, claims.Email)
		assert.Equal(t, originalClaims.Role, claims.Role)
		assert.Equal(t, originalClaims.SessionID, claims.SessionID)
		assert.Equal(t, originalClaims.Type, claims.Type)
	})

	t.Run("ValidateAccessToken", func(t *testing.T) {
		// Generate access token
		accessToken, _, err := jwtService.GenerateAccessToken(ctx, user, sessionID)
		require.NoError(t, err)

		// Validate as access token
		claims, err := jwtService.ValidateAccessToken(ctx, accessToken)
		require.NoError(t, err)
		assert.Equal(t, domain.TokenTypeAccess, claims.Type)

		// Generate refresh token and try to validate as access token (should fail)
		refreshToken, _, err := jwtService.GenerateRefreshToken(ctx, user, sessionID)
		require.NoError(t, err)

		_, err = jwtService.ValidateAccessToken(ctx, refreshToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token type")
	})

	t.Run("ValidateRefreshToken", func(t *testing.T) {
		// Generate refresh token
		refreshToken, _, err := jwtService.GenerateRefreshToken(ctx, user, sessionID)
		require.NoError(t, err)

		// Validate as refresh token
		claims, err := jwtService.ValidateRefreshToken(ctx, refreshToken)
		require.NoError(t, err)
		assert.Equal(t, domain.TokenTypeRefresh, claims.Type)

		// Generate access token and try to validate as refresh token (should fail)
		accessToken, _, err := jwtService.GenerateAccessToken(ctx, user, sessionID)
		require.NoError(t, err)

		_, err = jwtService.ValidateRefreshToken(ctx, accessToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token type")
	})

	t.Run("ExtractTokenFromHeader", func(t *testing.T) {
		token := "test-token"
		authHeader := "Bearer " + token

		extractedToken, err := jwtService.ExtractTokenFromHeader(authHeader)
		require.NoError(t, err)
		assert.Equal(t, token, extractedToken)

		// Test invalid header
		_, err = jwtService.ExtractTokenFromHeader("Invalid header")
		assert.Error(t, err)

		// Test empty header
		_, err = jwtService.ExtractTokenFromHeader("")
		assert.Error(t, err)
	})

	t.Run("GetTokenExpiry", func(t *testing.T) {
		accessExpiry := jwtService.GetTokenExpiry(domain.TokenTypeAccess)
		refreshExpiry := jwtService.GetTokenExpiry(domain.TokenTypeRefresh)

		assert.Equal(t, config.AccessTokenExpiry, accessExpiry)
		assert.Equal(t, config.RefreshTokenExpiry, refreshExpiry)
	})

	t.Run("IsTokenExpired", func(t *testing.T) {
		// Create expired claims
		expiredClaims := &domain.TokenClaims{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		assert.True(t, jwtService.IsTokenExpired(expiredClaims))

		// Create valid claims
		validClaims := &domain.TokenClaims{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		assert.False(t, jwtService.IsTokenExpired(validClaims))
	})
}

func TestValidateConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		config := &Config{
			SecretKey:          "this-is-a-very-long-secret-key-for-testing",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			RefreshTokenLength: 32,
		}

		err := ValidateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		// Empty secret key
		config := &Config{
			SecretKey:          "",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			RefreshTokenLength: 32,
		}
		err := ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret key is required")

		// Short secret key
		config.SecretKey = "short"
		err = ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret key must be at least 32 characters long")

		// Invalid expiry times
		config.SecretKey = "this-is-a-very-long-secret-key-for-testing"
		config.AccessTokenExpiry = 0
		err = ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access token expiry must be positive")

		config.AccessTokenExpiry = 15 * time.Minute
		config.RefreshTokenExpiry = 0
		err = ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token expiry must be positive")

		// Invalid refresh token length
		config.RefreshTokenExpiry = 7 * 24 * time.Hour
		config.RefreshTokenLength = 8
		err = ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token length must be at least 16")
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.NotEmpty(t, config.SecretKey)
	assert.Greater(t, config.AccessTokenExpiry, time.Duration(0))
	assert.Greater(t, config.RefreshTokenExpiry, time.Duration(0))
	assert.NotEmpty(t, config.Issuer)
	assert.NotEmpty(t, config.Audience)
	assert.Greater(t, config.RefreshTokenLength, 0)

	// Should pass validation
	err := ValidateConfig(config)
	assert.NoError(t, err)
}
