package security

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/crypto"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/models"
	"github.com/golang-jwt/jwt/v4"
)

// Service provides security operations
type Service struct {
	keyManager    *crypto.KeyManager
	logger        *logger.Logger
	jwtSecret     string
	jwtExpiry     time.Duration
	refreshExpiry time.Duration
}

// NewService creates a new security service
func NewService(
	keyManager *crypto.KeyManager,
	logger *logger.Logger,
	jwtSecret string,
	jwtExpiry time.Duration,
	refreshExpiry time.Duration,
) *Service {
	return &Service{
		keyManager:    keyManager,
		logger:        logger.Named("security-service"),
		jwtSecret:     jwtSecret,
		jwtExpiry:     jwtExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateKeyPair generates a new key pair
func (s *Service) GenerateKeyPair(ctx context.Context, req *models.GenerateKeyPairRequest) (*models.GenerateKeyPairResponse, error) {
	s.logger.Info(fmt.Sprintf("Generating key pair for chain %s", req.Chain))

	// Generate key pair
	privateKey, publicKey, address, err := s.keyManager.GenerateKeyPair()
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to generate key pair: %v", err))
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Key pair generated successfully for address %s", address))

	// Return response
	return &models.GenerateKeyPairResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}

// EncryptPrivateKey encrypts a private key
func (s *Service) EncryptPrivateKey(ctx context.Context, req *models.EncryptPrivateKeyRequest) (*models.EncryptPrivateKeyResponse, error) {
	s.logger.Info("Encrypting private key")

	// Encrypt private key
	encryptedKey, err := s.keyManager.EncryptPrivateKey(req.PrivateKey, req.Passphrase)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to encrypt private key: %v", err))
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	s.logger.Info("Private key encrypted successfully")

	// Return response
	return &models.EncryptPrivateKeyResponse{
		EncryptedKey: encryptedKey,
	}, nil
}

// DecryptPrivateKey decrypts a private key
func (s *Service) DecryptPrivateKey(ctx context.Context, req *models.DecryptPrivateKeyRequest) (*models.DecryptPrivateKeyResponse, error) {
	s.logger.Info("Decrypting private key")

	// Decrypt private key
	privateKey, err := s.keyManager.DecryptPrivateKey(req.EncryptedKey, req.Passphrase)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decrypt private key: %v", err))
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	s.logger.Info("Private key decrypted successfully")

	// Return response
	return &models.DecryptPrivateKeyResponse{
		PrivateKey: privateKey,
	}, nil
}

// GenerateJWT generates a JWT token
func (s *Service) GenerateJWT(ctx context.Context, req *models.GenerateJWTRequest) (*models.GenerateJWTResponse, error) {
	s.logger.Info(fmt.Sprintf("Generating JWT token for user %s", req.UserID))

	// Calculate expiry times
	now := time.Now()
	expiresAt := now.Add(s.jwtExpiry)
	refreshExpiresAt := now.Add(s.refreshExpiry)

	// Create token claims
	claims := jwt.MapClaims{
		"user_id": req.UserID,
		"email":   req.Email,
		"role":    req.Role,
		"exp":     expiresAt.Unix(),
		"iat":     now.Unix(),
	}

	// Create refresh token claims
	refreshClaims := jwt.MapClaims{
		"user_id": req.UserID,
		"exp":     refreshExpiresAt.Unix(),
		"iat":     now.Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to sign token: %v", err))
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to sign refresh token: %v", err))
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	s.logger.Info(fmt.Sprintf("JWT token generated successfully for user %s", req.UserID))

	// Return response
	return &models.GenerateJWTResponse{
		Token:        tokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

// VerifyJWT verifies a JWT token
func (s *Service) VerifyJWT(ctx context.Context, req *models.VerifyJWTRequest) (*models.VerifyJWTResponse, error) {
	s.logger.Info("Verifying JWT token")

	// Parse token
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	// Check for parsing errors
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to parse token: %v", err))
		return &models.VerifyJWTResponse{
			Valid: false,
		}, nil
	}

	// Check if token is valid
	if !token.Valid {
		s.logger.Error("Token is invalid")
		return &models.VerifyJWTResponse{
			Valid: false,
		}, nil
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.logger.Error("Failed to extract claims")
		return &models.VerifyJWTResponse{
			Valid: false,
		}, nil
	}

	// Extract user ID
	userID, ok := claims["user_id"].(string)
	if !ok {
		s.logger.Error("Failed to extract user ID")
		return &models.VerifyJWTResponse{
			Valid: false,
		}, nil
	}

	// Extract email
	email, ok := claims["email"].(string)
	if !ok {
		s.logger.Error("Failed to extract email")
		return &models.VerifyJWTResponse{
			Valid: false,
		}, nil
	}

	// Extract role
	role, ok := claims["role"].(string)
	if !ok {
		s.logger.Error("Failed to extract role")
		return &models.VerifyJWTResponse{
			Valid: false,
		}, nil
	}

	// Extract expiry
	exp, ok := claims["exp"].(float64)
	if !ok {
		s.logger.Error("Failed to extract expiry")
		return &models.VerifyJWTResponse{
			Valid: false,
		}, nil
	}

	s.logger.Info(fmt.Sprintf("JWT token verified successfully for user %s", userID))

	// Return response
	return &models.VerifyJWTResponse{
		Valid:     true,
		UserID:    userID,
		Email:     email,
		Role:      role,
		ExpiresAt: int64(exp),
	}, nil
}

// GenerateMnemonic generates a mnemonic phrase
func (s *Service) GenerateMnemonic(ctx context.Context, req *models.GenerateMnemonicRequest) (*models.GenerateMnemonicResponse, error) {
	s.logger.Info("Generating mnemonic phrase")

	// Generate mnemonic
	mnemonic, err := s.keyManager.GenerateMnemonic()
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to generate mnemonic: %v", err))
		return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	s.logger.Info("Mnemonic phrase generated successfully")

	// Return response
	return &models.GenerateMnemonicResponse{
		Mnemonic: mnemonic,
	}, nil
}

// ValidateMnemonic validates a mnemonic phrase
func (s *Service) ValidateMnemonic(ctx context.Context, req *models.ValidateMnemonicRequest) (*models.ValidateMnemonicResponse, error) {
	s.logger.Info("Validating mnemonic phrase")

	// Validate mnemonic
	valid := s.keyManager.ValidateMnemonic(req.Mnemonic)

	s.logger.Info(fmt.Sprintf("Mnemonic phrase validation result: %v", valid))

	// Return response
	return &models.ValidateMnemonicResponse{
		Valid: valid,
	}, nil
}

// MnemonicToPrivateKey converts a mnemonic to a private key
func (s *Service) MnemonicToPrivateKey(ctx context.Context, req *models.MnemonicToPrivateKeyRequest) (*models.MnemonicToPrivateKeyResponse, error) {
	s.logger.Info(fmt.Sprintf("Converting mnemonic to private key with path %s", req.Path))

	// Convert mnemonic to private key
	privateKey, err := s.keyManager.MnemonicToPrivateKey(req.Mnemonic, req.Path)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to convert mnemonic to private key: %v", err))
		return nil, fmt.Errorf("failed to convert mnemonic to private key: %w", err)
	}

	// Generate key pair from private key
	_, publicKey, address, err := s.keyManager.GenerateKeyPairFromPrivateKey(privateKey)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to generate key pair from private key: %v", err))
		return nil, fmt.Errorf("failed to generate key pair from private key: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Mnemonic converted to private key successfully for address %s", address))

	// Return response
	return &models.MnemonicToPrivateKeyResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}
