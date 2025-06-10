package jwt

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

// JWTService implements the JWTService interface
type JWTService struct {
	config *Config
	logger *logger.Logger
}

// Config represents JWT service configuration
type Config struct {
	SecretKey          string        `yaml:"secret_key"`
	AccessTokenExpiry  time.Duration `yaml:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
	Issuer             string        `yaml:"issuer"`
	Audience           string        `yaml:"audience"`
	RefreshTokenLength int           `yaml:"refresh_token_length"`
}

// Claims represents JWT claims
type Claims struct {
	UserID    string           `json:"user_id"`
	Email     string           `json:"email"`
	Role      domain.UserRole  `json:"role"`
	SessionID string           `json:"session_id"`
	TokenID   string           `json:"token_id"`
	Type      domain.TokenType `json:"type"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service
func NewJWTService(config *Config, logger *logger.Logger) application.JWTService {
	return &JWTService{
		config: config,
		logger: logger,
	}
}

// GenerateAccessToken generates an access token for a user
func (j *JWTService) GenerateAccessToken(ctx context.Context, user *domain.User, sessionID string) (string, *domain.TokenClaims, error) {
	now := time.Now()
	expiresAt := now.Add(j.config.AccessTokenExpiry)
	tokenID := generateJTI()

	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		SessionID: sessionID,
		TokenID:   tokenID,
		Type:      domain.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Audience:  jwt.ClaimStrings{j.config.Audience},
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		j.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to sign access token")
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	tokenClaims := &domain.TokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		SessionID: sessionID,
		TokenID:   tokenID,
		Type:      domain.TokenTypeAccess,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
	}

	j.logger.WithFields(map[string]interface{}{
		"user_id":    user.ID,
		"session_id": sessionID,
		"token_id":   tokenID,
	}).Debug("Access token generated successfully")

	return tokenString, tokenClaims, nil
}

// GenerateRefreshToken generates a refresh token for a user
func (j *JWTService) GenerateRefreshToken(ctx context.Context, user *domain.User, sessionID string) (string, *domain.TokenClaims, error) {
	now := time.Now()
	expiresAt := now.Add(j.config.RefreshTokenExpiry)
	tokenID := generateJTI()

	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		SessionID: sessionID,
		TokenID:   tokenID,
		Type:      domain.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Audience:  jwt.ClaimStrings{j.config.Audience},
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		j.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to sign refresh token")
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	tokenClaims := &domain.TokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		SessionID: sessionID,
		TokenID:   tokenID,
		Type:      domain.TokenTypeRefresh,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
	}

	j.logger.WithFields(map[string]interface{}{
		"user_id":    user.ID,
		"session_id": sessionID,
		"token_id":   tokenID,
	}).Debug("Refresh token generated successfully")

	return tokenString, tokenClaims, nil
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTService) GenerateTokenPair(ctx context.Context, user *domain.User, sessionID string) (accessToken, refreshToken string, accessClaims, refreshClaims *domain.TokenClaims, err error) {
	// Generate access token
	accessToken, accessClaims, err = j.GenerateAccessToken(ctx, user, sessionID)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, refreshClaims, err = j.GenerateRefreshToken(ctx, user, sessionID)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, accessClaims, refreshClaims, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTService) ValidateToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		j.logger.WithError(err).Error("Failed to parse JWT token")
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		j.logger.Error("Invalid JWT token claims")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token is expired")
	}

	tokenClaims := &domain.TokenClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Role:      claims.Role,
		SessionID: claims.SessionID,
		TokenID:   claims.TokenID,
		Type:      claims.Type,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	j.logger.WithFields(map[string]interface{}{
		"user_id":    claims.UserID,
		"session_id": claims.SessionID,
		"token_id":   claims.TokenID,
	}).Debug("Token validated successfully")

	return tokenClaims, nil
}

// ValidateAccessToken validates an access token
func (j *JWTService) ValidateAccessToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	claims, err := j.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != domain.TokenTypeAccess {
		return nil, fmt.Errorf("invalid token type: expected access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (j *JWTService) ValidateRefreshToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	claims, err := j.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != domain.TokenTypeRefresh {
		return nil, fmt.Errorf("invalid token type: expected refresh token")
	}

	return claims, nil
}

// ParseToken parses a token and returns claims (with validation)
func (j *JWTService) ParseToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	return j.ValidateToken(ctx, tokenString)
}

// ParseTokenWithoutValidation parses a token without validation
func (j *JWTService) ParseTokenWithoutValidation(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	tokenClaims := &domain.TokenClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Role:      claims.Role,
		SessionID: claims.SessionID,
		TokenID:   claims.TokenID,
		Type:      claims.Type,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	return tokenClaims, nil
}

// ExtractTokenFromHeader extracts token from Authorization header
func (j *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

// GetTokenExpiry returns the expiry duration for a token type
func (j *JWTService) GetTokenExpiry(tokenType domain.TokenType) time.Duration {
	switch tokenType {
	case domain.TokenTypeAccess:
		return j.config.AccessTokenExpiry
	case domain.TokenTypeRefresh:
		return j.config.RefreshTokenExpiry
	default:
		return j.config.AccessTokenExpiry
	}
}

// IsTokenExpired checks if token claims are expired
func (j *JWTService) IsTokenExpired(claims *domain.TokenClaims) bool {
	return time.Now().After(claims.ExpiresAt)
}

// generateJTI generates a unique JWT ID
func generateJTI() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// DefaultConfig returns default JWT service configuration
func DefaultConfig() *Config {
	return &Config{
		SecretKey:          "your-secret-key-change-in-production",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour, // 7 days
		Issuer:             "go-coffee-auth",
		Audience:           "go-coffee-api",
		RefreshTokenLength: 32,
	}
}

// ValidateConfig validates JWT service configuration
func ValidateConfig(config *Config) error {
	if config.SecretKey == "" {
		return fmt.Errorf("secret key is required")
	}
	if len(config.SecretKey) < 32 {
		return fmt.Errorf("secret key must be at least 32 characters long")
	}
	if config.AccessTokenExpiry <= 0 {
		return fmt.Errorf("access token expiry must be positive")
	}
	if config.RefreshTokenExpiry <= 0 {
		return fmt.Errorf("refresh token expiry must be positive")
	}
	if config.RefreshTokenLength < 16 {
		return fmt.Errorf("refresh token length must be at least 16")
	}
	return nil
}
