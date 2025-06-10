package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// JWTService handles JWT token operations
type JWTService interface {
	// Token generation
	GenerateAccessToken(ctx context.Context, userID, email, role string, metadata map[string]interface{}) (string, *Claims, error)
	GenerateRefreshToken(ctx context.Context, userID string) (string, error)
	GenerateTokenPair(ctx context.Context, userID, email, role string, metadata map[string]interface{}) (*TokenPair, error)
	
	// Token validation
	ValidateToken(ctx context.Context, tokenString string) (*Claims, error)
	ValidateAccessToken(ctx context.Context, tokenString string) (*Claims, error)
	ValidateRefreshToken(ctx context.Context, tokenString string) (*Claims, error)
	
	// Token operations
	RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	RevokeToken(ctx context.Context, tokenID string) error
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)
	
	// Token introspection
	GetTokenClaims(ctx context.Context, tokenString string) (*Claims, error)
	GetTokenExpiry(ctx context.Context, tokenString string) (time.Time, error)
	IsTokenExpired(ctx context.Context, tokenString string) (bool, error)
}

// Claims represents JWT claims
type Claims struct {
	UserID   string                 `json:"user_id"`
	Email    string                 `json:"email"`
	Role     string                 `json:"role"`
	TokenID  string                 `json:"token_id"`
	TokenType string                `json:"token_type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// JWTServiceImpl implements JWTService
type JWTServiceImpl struct {
	config *config.JWTConfig
	logger *logger.Logger
	revokedTokens map[string]time.Time // In production, use Redis or database
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg *config.JWTConfig, logger *logger.Logger) JWTService {
	return &JWTServiceImpl{
		config:        cfg,
		logger:        logger,
		revokedTokens: make(map[string]time.Time),
	}
}

// GenerateAccessToken generates an access token
func (j *JWTServiceImpl) GenerateAccessToken(ctx context.Context, userID, email, role string, metadata map[string]interface{}) (string, *Claims, error) {
	tokenID, err := generateTokenID()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token ID: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(j.config.AccessTokenTTL)

	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenID:   tokenID,
		TokenType: "access",
		Metadata:  metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   userID,
			Issuer:    j.config.Issuer,
			Audience:  jwt.ClaimStrings{j.config.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod(j.config.Algorithm), claims)
	tokenString, err := token.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, claims, nil
}

// GenerateRefreshToken generates a refresh token
func (j *JWTServiceImpl) GenerateRefreshToken(ctx context.Context, userID string) (string, error) {
	tokenID, err := generateTokenID()
	if err != nil {
		return "", fmt.Errorf("failed to generate token ID: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(j.config.RefreshTokenTTL)

	claims := &Claims{
		UserID:    userID,
		TokenID:   tokenID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   userID,
			Issuer:    j.config.Issuer,
			Audience:  jwt.ClaimStrings{j.config.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod(j.config.Algorithm), claims)
	tokenString, err := token.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTServiceImpl) GenerateTokenPair(ctx context.Context, userID, email, role string, metadata map[string]interface{}) (*TokenPair, error) {
	accessToken, claims, err := j.GenerateAccessToken(ctx, userID, email, role, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := j.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.config.AccessTokenTTL.Seconds()),
		ExpiresAt:    claims.ExpiresAt.Time,
	}, nil
}

// ValidateToken validates a JWT token
func (j *JWTServiceImpl) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is revoked
	if revoked, err := j.IsTokenRevoked(ctx, claims.TokenID); err != nil {
		return nil, fmt.Errorf("failed to check token revocation: %w", err)
	} else if revoked {
		return nil, fmt.Errorf("token has been revoked")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (j *JWTServiceImpl) ValidateAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type: expected access, got %s", claims.TokenType)
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (j *JWTServiceImpl) ValidateRefreshToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type: expected refresh, got %s", claims.TokenType)
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (j *JWTServiceImpl) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := j.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if refresh token is close to expiry
	timeUntilExpiry := time.Until(claims.ExpiresAt.Time)
	if timeUntilExpiry < j.config.RefreshThreshold {
		// Generate new token pair
		return j.GenerateTokenPair(ctx, claims.UserID, claims.Email, claims.Role, claims.Metadata)
	}

	// Generate only new access token
	accessToken, accessClaims, err := j.GenerateAccessToken(ctx, claims.UserID, claims.Email, claims.Role, claims.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Keep the same refresh token
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.config.AccessTokenTTL.Seconds()),
		ExpiresAt:    accessClaims.ExpiresAt.Time,
	}, nil
}

// RevokeToken revokes a token
func (j *JWTServiceImpl) RevokeToken(ctx context.Context, tokenID string) error {
	j.revokedTokens[tokenID] = time.Now()
	j.logger.InfoWithFields("Token revoked", logger.String("token_id", tokenID))
	return nil
}

// IsTokenRevoked checks if a token is revoked
func (j *JWTServiceImpl) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	_, revoked := j.revokedTokens[tokenID]
	return revoked, nil
}

// GetTokenClaims extracts claims from a token without validation
func (j *JWTServiceImpl) GetTokenClaims(ctx context.Context, tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GetTokenExpiry returns token expiry time
func (j *JWTServiceImpl) GetTokenExpiry(ctx context.Context, tokenString string) (time.Time, error) {
	claims, err := j.GetTokenClaims(ctx, tokenString)
	if err != nil {
		return time.Time{}, err
	}

	return claims.ExpiresAt.Time, nil
}

// IsTokenExpired checks if a token is expired
func (j *JWTServiceImpl) IsTokenExpired(ctx context.Context, tokenString string) (bool, error) {
	expiryTime, err := j.GetTokenExpiry(ctx, tokenString)
	if err != nil {
		return true, err
	}

	return time.Now().After(expiryTime), nil
}

// generateTokenID generates a unique token ID
func generateTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetTimeUntilExpiryClaims returns time until token expiry
func (c *Claims) GetTimeUntilExpiryClaims() time.Duration {
	return time.Until(c.ExpiresAt.Time)
}

// IsExpired checks if the token is expired
func (c *Claims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt.Time)
}

// IsValidForRefresh checks if the token can be refreshed
func (c *Claims) IsValidForRefresh(threshold time.Duration) bool {
	return c.GetTimeUntilExpiryClaims() < threshold
}
