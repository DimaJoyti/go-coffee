package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// EncryptionService handles encryption and decryption operations
type EncryptionService interface {
	// Symmetric encryption
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
	EncryptBytes(data []byte) ([]byte, error)
	DecryptBytes(data []byte) ([]byte, error)
	
	// Password hashing
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
	
	// Key derivation
	DeriveKey(password, salt []byte) ([]byte, error)
	GenerateSalt() ([]byte, error)
	
	// Utility functions
	GenerateRandomBytes(length int) ([]byte, error)
	GenerateRandomString(length int) (string, error)
	Hash(data []byte) []byte
	HashString(data string) string
}

// EncryptionServiceImpl implements EncryptionService
type EncryptionServiceImpl struct {
	config *config.EncryptionConfig
	logger *logger.Logger
	gcm    cipher.AEAD
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(cfg *config.EncryptionConfig, logger *logger.Logger) (EncryptionService, error) {
	if cfg.AESKey == "" {
		return nil, fmt.Errorf("AES key is required")
	}

	// Derive key from the provided key
	key := sha256.Sum256([]byte(cfg.AESKey))
	
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &EncryptionServiceImpl{
		config: cfg,
		logger: logger,
		gcm:    gcm,
	}, nil
}

// Encrypt encrypts plaintext and returns base64 encoded ciphertext
func (e *EncryptionServiceImpl) Encrypt(plaintext string) (string, error) {
	ciphertext, err := e.EncryptBytes([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64 encoded ciphertext and returns plaintext
func (e *EncryptionServiceImpl) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	
	plaintext, err := e.DecryptBytes(data)
	if err != nil {
		return "", err
	}
	
	return string(plaintext), nil
}

// EncryptBytes encrypts byte data
func (e *EncryptionServiceImpl) EncryptBytes(data []byte) ([]byte, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := e.gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptBytes decrypts byte data
func (e *EncryptionServiceImpl) DecryptBytes(data []byte) ([]byte, error) {
	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// HashPassword hashes a password using bcrypt
func (e *EncryptionServiceImpl) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (e *EncryptionServiceImpl) VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}
	return nil
}

// DeriveKey derives a key from password and salt using scrypt
func (e *EncryptionServiceImpl) DeriveKey(password, salt []byte) ([]byte, error) {
	key, err := scrypt.Key(password, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}
	return key, nil
}

// GenerateSalt generates a random salt
func (e *EncryptionServiceImpl) GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// GenerateRandomBytes generates random bytes
func (e *EncryptionServiceImpl) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// GenerateRandomString generates a random base64 string
func (e *EncryptionServiceImpl) GenerateRandomString(length int) (string, error) {
	bytes, err := e.GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Hash creates a SHA-256 hash of the data
func (e *EncryptionServiceImpl) Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashString creates a SHA-256 hash of the string
func (e *EncryptionServiceImpl) HashString(data string) string {
	hash := e.Hash([]byte(data))
	return base64.StdEncoding.EncodeToString(hash)
}

// PasswordPolicy represents password policy configuration
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
	MaxLength        int  `json:"max_length"`
}

// DefaultPasswordPolicy returns default password policy
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   false,
		MaxLength:        128,
	}
}

// ValidatePassword validates a password against the policy
func (p *PasswordPolicy) ValidatePassword(password string) error {
	if len(password) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters long", p.MinLength)
	}
	
	if p.MaxLength > 0 && len(password) > p.MaxLength {
		return fmt.Errorf("password must be at most %d characters long", p.MaxLength)
	}
	
	if p.RequireUppercase && !containsUppercase(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	
	if p.RequireLowercase && !containsLowercase(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	
	if p.RequireNumbers && !containsNumber(password) {
		return fmt.Errorf("password must contain at least one number")
	}
	
	if p.RequireSymbols && !containsSymbol(password) {
		return fmt.Errorf("password must contain at least one symbol")
	}
	
	return nil
}

// Helper functions for password validation
func containsUppercase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func containsLowercase(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func containsSymbol(s string) bool {
	symbols := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, r := range s {
		for _, sym := range symbols {
			if r == sym {
				return true
			}
		}
	}
	return false
}

// SecureString represents a string that should be handled securely
type SecureString struct {
	value []byte
}

// NewSecureString creates a new secure string
func NewSecureString(value string) *SecureString {
	return &SecureString{
		value: []byte(value),
	}
}

// String returns the string value (use with caution)
func (s *SecureString) String() string {
	return string(s.value)
}

// Bytes returns the byte value (use with caution)
func (s *SecureString) Bytes() []byte {
	return s.value
}

// Clear securely clears the string from memory
func (s *SecureString) Clear() {
	for i := range s.value {
		s.value[i] = 0
	}
	s.value = nil
}

// Length returns the length of the string
func (s *SecureString) Length() int {
	return len(s.value)
}

// IsEmpty checks if the string is empty
func (s *SecureString) IsEmpty() bool {
	return len(s.value) == 0
}

// Equals compares two secure strings
func (s *SecureString) Equals(other *SecureString) bool {
	if s.Length() != other.Length() {
		return false
	}
	
	result := byte(0)
	for i := 0; i < s.Length(); i++ {
		result |= s.value[i] ^ other.value[i]
	}
	
	return result == 0
}
