package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	signingMethod jwt.SigningMethod
	issuer        string
	logger        Logger
}

// Claims represents JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	TokenType   string   `json:"token_type"`
	jwt.RegisteredClaims
}

// RefreshClaims represents refresh token claims
type RefreshClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(privateKeyPEM, publicKeyPEM []byte, issuer string, logger Logger) (*JWTManager, error) {
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	var err error

	// Parse private key if provided
	if len(privateKeyPEM) > 0 {
		privateKey, err = parsePrivateKey(privateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	// Parse public key if provided
	if len(publicKeyPEM) > 0 {
		publicKey, err = parsePublicKey(publicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	// If no keys provided, generate them
	if privateKey == nil && publicKey == nil {
		privateKey, publicKey, err = generateKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}
		logger.Warn("Generated new RSA key pair for JWT signing. In production, use persistent keys.")
	}

	// If only private key provided, derive public key
	if privateKey != nil && publicKey == nil {
		publicKey = &privateKey.PublicKey
	}

	return &JWTManager{
		privateKey:    privateKey,
		publicKey:     publicKey,
		signingMethod: jwt.SigningMethodRS256,
		issuer:        issuer,
		logger:        logger,
	}, nil
}

// GenerateToken generates a new access token
func (jm *JWTManager) GenerateToken(userID string, roles []string, expiration time.Duration) (string, error) {
	if jm.privateKey == nil {
		return "", fmt.Errorf("private key not available for token generation")
	}

	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Roles:     roles,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Subject:   userID,
			Audience:  []string{"orchestration-engine"},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jm.signingMethod, claims)
	tokenString, err := token.SignedString(jm.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	jm.logger.Debug("Generated access token", "user_id", userID, "expires_at", claims.ExpiresAt)
	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token
func (jm *JWTManager) GenerateRefreshToken(userID string, expiration time.Duration) (string, error) {
	if jm.privateKey == nil {
		return "", fmt.Errorf("private key not available for token generation")
	}

	now := time.Now()
	claims := &RefreshClaims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Subject:   userID,
			Audience:  []string{"orchestration-engine"},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jm.signingMethod, claims)
	tokenString, err := token.SignedString(jm.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	jm.logger.Debug("Generated refresh token", "user_id", userID, "expires_at", claims.ExpiresAt)
	return tokenString, nil
}

// ValidateToken validates and parses an access token
func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	if jm.publicKey == nil {
		return nil, fmt.Errorf("public key not available for token validation")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify token type
	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type: expected access, got %s", claims.TokenType)
	}

	// Verify issuer
	if claims.Issuer != jm.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", jm.issuer, claims.Issuer)
	}

	jm.logger.Debug("Validated access token", "user_id", claims.UserID, "jti", claims.ID)
	return claims, nil
}

// ValidateRefreshToken validates and parses a refresh token
func (jm *JWTManager) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	if jm.publicKey == nil {
		return nil, fmt.Errorf("public key not available for token validation")
	}

	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token claims")
	}

	// Verify token type
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type: expected refresh, got %s", claims.TokenType)
	}

	// Verify issuer
	if claims.Issuer != jm.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", jm.issuer, claims.Issuer)
	}

	jm.logger.Debug("Validated refresh token", "user_id", claims.UserID, "jti", claims.ID)
	return claims, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func (jm *JWTManager) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// GetPublicKeyPEM returns the public key in PEM format
func (jm *JWTManager) GetPublicKeyPEM() ([]byte, error) {
	if jm.publicKey == nil {
		return nil, fmt.Errorf("public key not available")
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(jm.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	return publicKeyPEM, nil
}

// GetPrivateKeyPEM returns the private key in PEM format
func (jm *JWTManager) GetPrivateKeyPEM() ([]byte, error) {
	if jm.privateKey == nil {
		return nil, fmt.Errorf("private key not available")
	}

	privateKeyDER := x509.MarshalPKCS1PrivateKey(jm.privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	})

	return privateKeyPEM, nil
}

// Helper functions

// parsePrivateKey parses a PEM-encoded RSA private key
func parsePrivateKey(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}
}

// parsePublicKey parses a PEM-encoded RSA public key
func parsePublicKey(publicKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	switch block.Type {
	case "PUBLIC KEY":
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA public key")
		}
		return rsaKey, nil
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported public key type: %s", block.Type)
	}
}

// generateKeyPair generates a new RSA key pair
func generateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return privateKey, &privateKey.PublicKey, nil
}

// generateJTI generates a unique JWT ID
func generateJTI() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// TokenBlacklist manages blacklisted tokens
type TokenBlacklist struct {
	blacklistedTokens map[string]time.Time
	logger            Logger
	mutex             sync.RWMutex
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist(logger Logger) *TokenBlacklist {
	return &TokenBlacklist{
		blacklistedTokens: make(map[string]time.Time),
		logger:            logger,
	}
}

// BlacklistToken adds a token to the blacklist
func (tb *TokenBlacklist) BlacklistToken(jti string, expiresAt time.Time) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	tb.blacklistedTokens[jti] = expiresAt
	tb.logger.Debug("Token blacklisted", "jti", jti, "expires_at", expiresAt)
}

// IsBlacklisted checks if a token is blacklisted
func (tb *TokenBlacklist) IsBlacklisted(jti string) bool {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	
	expiresAt, exists := tb.blacklistedTokens[jti]
	if !exists {
		return false
	}
	
	// Remove expired blacklisted tokens
	if time.Now().After(expiresAt) {
		delete(tb.blacklistedTokens, jti)
		return false
	}
	
	return true
}

// CleanupExpired removes expired tokens from the blacklist
func (tb *TokenBlacklist) CleanupExpired() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	now := time.Now()
	for jti, expiresAt := range tb.blacklistedTokens {
		if now.After(expiresAt) {
			delete(tb.blacklistedTokens, jti)
		}
	}
	
	tb.logger.Debug("Cleaned up expired blacklisted tokens")
}

// GetBlacklistedCount returns the number of blacklisted tokens
func (tb *TokenBlacklist) GetBlacklistedCount() int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return len(tb.blacklistedTokens)
}
