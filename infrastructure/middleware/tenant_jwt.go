package middleware

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/DimaJoyti/go-coffee/domain/shared"
)

// TenantJWTExtractor handles JWT-based tenant ID extraction
type TenantJWTExtractor struct {
	publicKey    *rsa.PublicKey
	issuer       string
	audience     string
	clockSkew    time.Duration
	tenantClaim  string
}

// TenantJWTConfig holds configuration for JWT tenant extraction
type TenantJWTConfig struct {
	PublicKey   *rsa.PublicKey
	Issuer      string
	Audience    string
	ClockSkew   time.Duration
	TenantClaim string // The claim name that contains tenant ID (default: "tenant_id")
}

// NewTenantJWTExtractor creates a new JWT-based tenant extractor
func NewTenantJWTExtractor(config TenantJWTConfig) *TenantJWTExtractor {
	if config.ClockSkew == 0 {
		config.ClockSkew = 5 * time.Minute // Default clock skew
	}
	if config.TenantClaim == "" {
		config.TenantClaim = "tenant_id" // Default claim name
	}

	return &TenantJWTExtractor{
		publicKey:   config.PublicKey,
		issuer:      config.Issuer,
		audience:    config.Audience,
		clockSkew:   config.ClockSkew,
		tenantClaim: config.TenantClaim,
	}
}

// TenantClaims represents the JWT claims structure with tenant information
type TenantClaims struct {
	TenantID     string   `json:"tenant_id"`
	TenantName   string   `json:"tenant_name,omitempty"`
	UserID       string   `json:"user_id,omitempty"`
	Roles        []string `json:"roles,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// ExtractTenantFromJWT extracts tenant ID from JWT token in the request
func (e *TenantJWTExtractor) ExtractTenantFromJWT(r *http.Request) (shared.TenantID, error) {
	// Extract token from Authorization header
	token, err := e.extractTokenFromHeader(r)
	if err != nil {
		return shared.TenantID{}, err
	}

	// Parse and validate the JWT token
	claims, err := e.parseAndValidateToken(token)
	if err != nil {
		return shared.TenantID{}, err
	}

	// Extract tenant ID from claims
	tenantID := claims.TenantID
	if tenantID == "" {
		return shared.TenantID{}, errors.New("tenant_id claim not found in JWT token")
	}

	return shared.NewTenantID(tenantID)
}

// ExtractTenantClaimsFromJWT extracts full tenant claims from JWT token
func (e *TenantJWTExtractor) ExtractTenantClaimsFromJWT(r *http.Request) (*TenantClaims, error) {
	// Extract token from Authorization header
	token, err := e.extractTokenFromHeader(r)
	if err != nil {
		return nil, err
	}

	// Parse and validate the JWT token
	return e.parseAndValidateToken(token)
}

// extractTokenFromHeader extracts JWT token from Authorization header
func (e *TenantJWTExtractor) extractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header not found")
	}

	// Check for Bearer token format
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization header format, expected 'Bearer <token>'")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", errors.New("empty JWT token")
	}

	return token, nil
}

// parseAndValidateToken parses and validates the JWT token
func (e *TenantJWTExtractor) parseAndValidateToken(tokenString string) (*TenantClaims, error) {
	// Parse the token with custom claims
	token, err := jwt.ParseWithClaims(tokenString, &TenantClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return e.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	// Extract and validate claims
	claims, ok := token.Claims.(*TenantClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid JWT token or claims")
	}

	// Validate standard claims
	if err := e.validateClaims(claims); err != nil {
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	return claims, nil
}

// validateClaims validates the JWT claims
func (e *TenantJWTExtractor) validateClaims(claims *TenantClaims) error {
	now := time.Now()

	// Check expiration with clock skew
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Add(e.clockSkew).Before(now) {
		return errors.New("token has expired")
	}

	// Check not before with clock skew
	if claims.NotBefore != nil && claims.NotBefore.Time.Add(-e.clockSkew).After(now) {
		return errors.New("token is not yet valid")
	}

	// Check issued at with clock skew
	if claims.IssuedAt != nil && claims.IssuedAt.Time.Add(-e.clockSkew).After(now) {
		return errors.New("token was issued in the future")
	}

	// Validate issuer
	if e.issuer != "" && claims.Issuer != e.issuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", e.issuer, claims.Issuer)
	}

	// Validate audience
	if e.audience != "" {
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == e.audience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return fmt.Errorf("invalid audience: expected %s", e.audience)
		}
	}

	return nil
}

// CreateTenantJWT creates a JWT token with tenant information (for testing/development)
func (e *TenantJWTExtractor) CreateTenantJWT(tenantID, userID string, roles []string, duration time.Duration) (string, error) {
	if e.publicKey == nil {
		return "", errors.New("cannot create JWT without private key")
	}

	now := time.Now()
	claims := &TenantClaims{
		TenantID: tenantID,
		UserID:   userID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    e.issuer,
			Audience:  []string{e.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	
	// Note: This would need the private key, not public key
	// This is just for demonstration - in practice, token creation
	// would be done by an authentication service with the private key
	return token.SignedString(e.publicKey)
}

// ValidateJWTMiddleware creates a middleware that validates JWT tokens
func (e *TenantJWTExtractor) ValidateJWTMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := e.ExtractTenantClaimsFromJWT(r)
			if err != nil {
				writeJWTError(w, http.StatusUnauthorized, "Invalid JWT token", err.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeJWTError writes a JSON error response for JWT-related errors
func writeJWTError(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"error":   error,
		"message": message,
		"type":    "jwt_error",
	}

	json.NewEncoder(w).Encode(response)
}
