package security

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret           string        `yaml:"secret"`
	AccessTokenTTL   time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `yaml:"refresh_token_ttl"`
	Issuer           string        `yaml:"issuer"`
	Audience         string        `yaml:"audience"`
}

// JWTService implements JWT token operations
type JWTService struct {
	config *JWTConfig
	logger *logger.Logger
}

// NewJWTService creates a new JWT service
func NewJWTService(config *JWTConfig, logger *logger.Logger) *JWTService {
	return &JWTService{
		config: config,
		logger: logger,
	}
}

// GenerateAccessToken generates an access token for the user
func (j *JWTService) GenerateAccessToken(ctx context.Context, user *domain.User, sessionID string) (string, *domain.TokenClaims, error) {
	return j.generateToken(ctx, user, sessionID, domain.TokenTypeAccess, j.config.AccessTokenTTL)
}

// GenerateRefreshToken generates a refresh token for the user
func (j *JWTService) GenerateRefreshToken(ctx context.Context, user *domain.User, sessionID string) (string, *domain.TokenClaims, error) {
	return j.generateToken(ctx, user, sessionID, domain.TokenTypeRefresh, j.config.RefreshTokenTTL)
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

// generateToken generates a JWT token with the specified type and TTL
func (j *JWTService) generateToken(ctx context.Context, user *domain.User, sessionID string, tokenType domain.TokenType, ttl time.Duration) (string, *domain.TokenClaims, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	// Create token claims
	claims := domain.NewTokenClaims(
		user.ID,
		user.Email,
		user.Role,
		sessionID,
		"", // Token ID will be set after token creation
		tokenType,
		expiresAt,
	)

	// Create JWT claims
	jwtClaims := jwt.MapClaims{
		"user_id":    claims.UserID,
		"email":      claims.Email,
		"role":       string(claims.Role),
		"session_id": claims.SessionID,
		"token_id":   claims.TokenID,
		"type":       string(claims.Type),
		"iat":        claims.IssuedAt.Unix(),
		"exp":        claims.ExpiresAt.Unix(),
		"iss":        j.config.Issuer,
		"aud":        j.config.Audience,
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenString, err := token.SignedString([]byte(j.config.Secret))
	if err != nil {
		j.logger.Error("Failed to sign JWT token", zap.Error(err))
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	j.logger.Debug("Generated JWT token",
		zap.String("user_id", user.ID),
		zap.String("session_id", sessionID),
		zap.String("type", string(tokenType)),
		zap.Time("expires_at", expiresAt),
	)

	return tokenString, claims, nil
}

// ValidateToken validates a JWT token and returns claims
func (j *JWTService) ValidateToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	return j.parseAndValidateToken(ctx, tokenString, true)
}

// ValidateAccessToken validates an access token
func (j *JWTService) ValidateAccessToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	claims, err := j.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != domain.TokenTypeAccess {
		return nil, domain.ErrTokenTypeInvalid
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
		return nil, domain.ErrTokenTypeInvalid
	}

	return claims, nil
}

// ParseToken parses a JWT token and returns claims with validation
func (j *JWTService) ParseToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	return j.parseAndValidateToken(ctx, tokenString, true)
}

// ParseTokenWithoutValidation parses a JWT token without validation
func (j *JWTService) ParseTokenWithoutValidation(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	return j.parseAndValidateToken(ctx, tokenString, false)
}

// parseAndValidateToken parses and optionally validates a JWT token
func (j *JWTService) parseAndValidateToken(ctx context.Context, tokenString string, validate bool) (*domain.TokenClaims, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		j.logger.Error("Failed to parse JWT token", zap.Error(err))
		return nil, domain.ErrTokenMalformed
	}

	// Extract claims
	jwtClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		j.logger.Error("Failed to extract JWT claims")
		return nil, domain.ErrTokenClaims
	}

	// Validate token if required
	if validate && !token.Valid {
		j.logger.Error("JWT token is invalid")
		return nil, domain.ErrTokenInvalid
	}

	// Convert to domain claims
	claims, err := j.mapClaimsToDomain(jwtClaims)
	if err != nil {
		j.logger.Error("Failed to convert JWT claims", zap.Error(err))
		return nil, err
	}

	// Validate claims
	if validate {
		if err := claims.IsValidClaims(); err != nil {
			j.logger.Error("Invalid token claims", zap.Error(err))
			return nil, err
		}
	}

	return claims, nil
}

// mapClaimsToDomain converts JWT claims to domain token claims
func (j *JWTService) mapClaimsToDomain(jwtClaims jwt.MapClaims) (*domain.TokenClaims, error) {
	// Extract required fields
	userID, ok := jwtClaims["user_id"].(string)
	if !ok {
		return nil, domain.ErrTokenClaims
	}

	email, ok := jwtClaims["email"].(string)
	if !ok {
		return nil, domain.ErrTokenClaims
	}

	roleStr, ok := jwtClaims["role"].(string)
	if !ok {
		return nil, domain.ErrTokenClaims
	}

	sessionID, ok := jwtClaims["session_id"].(string)
	if !ok {
		return nil, domain.ErrTokenClaims
	}

	tokenID, ok := jwtClaims["token_id"].(string)
	if !ok {
		tokenID = "" // Token ID might be empty for some tokens
	}

	typeStr, ok := jwtClaims["type"].(string)
	if !ok {
		return nil, domain.ErrTokenClaims
	}

	// Extract timestamps
	iatFloat, ok := jwtClaims["iat"].(float64)
	if !ok {
		return nil, domain.ErrTokenClaims
	}

	expFloat, ok := jwtClaims["exp"].(float64)
	if !ok {
		return nil, domain.ErrTokenClaims
	}

	// Convert to domain types
	role := domain.UserRole(roleStr)
	tokenType := domain.TokenType(typeStr)
	issuedAt := time.Unix(int64(iatFloat), 0)
	expiresAt := time.Unix(int64(expFloat), 0)

	return &domain.TokenClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		SessionID: sessionID,
		TokenID:   tokenID,
		Type:      tokenType,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}, nil
}

// ExtractTokenFromHeader extracts token from Authorization header
func (j *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	// Check for Bearer prefix
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", errors.New("token is empty")
	}

	return token, nil
}

// GetTokenExpiry returns the expiry duration for the specified token type
func (j *JWTService) GetTokenExpiry(tokenType domain.TokenType) time.Duration {
	switch tokenType {
	case domain.TokenTypeAccess:
		return j.config.AccessTokenTTL
	case domain.TokenTypeRefresh:
		return j.config.RefreshTokenTTL
	default:
		return j.config.AccessTokenTTL
	}
}

// IsTokenExpired checks if the token claims are expired
func (j *JWTService) IsTokenExpired(claims *domain.TokenClaims) bool {
	return claims.IsExpiredClaims()
}
