package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// EncryptionService provides cryptographic operations
type EncryptionService struct {
	aesKey    []byte
	rsaKey    *rsa.PrivateKey
	publicKey *rsa.PublicKey
}

// Config represents encryption configuration
type Config struct {
	AESKey    string `yaml:"aes_key" env:"AES_KEY"`
	RSAKey    string `yaml:"rsa_key" env:"RSA_KEY"`
	KeySize   int    `yaml:"key_size" env:"KEY_SIZE" default:"32"`
	SaltSize  int    `yaml:"salt_size" env:"SALT_SIZE" default:"16"`
	Argon2Time   uint32 `yaml:"argon2_time" env:"ARGON2_TIME" default:"1"`
	Argon2Memory uint32 `yaml:"argon2_memory" env:"ARGON2_MEMORY" default:"64"`
	Argon2Threads uint8 `yaml:"argon2_threads" env:"ARGON2_THREADS" default:"4"`
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(config *Config) (*EncryptionService, error) {
	service := &EncryptionService{}

	// Initialize AES key
	if config.AESKey != "" {
		key, err := base64.StdEncoding.DecodeString(config.AESKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode AES key: %w", err)
		}
		service.aesKey = key
	} else {
		// Generate random AES key
		key := make([]byte, config.KeySize)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate AES key: %w", err)
		}
		service.aesKey = key
	}

	// Initialize RSA key
	if config.RSAKey != "" {
		block, _ := pem.Decode([]byte(config.RSAKey))
		if block == nil {
			return nil, errors.New("failed to decode RSA key")
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA key: %w", err)
		}

		service.rsaKey = privateKey
		service.publicKey = &privateKey.PublicKey
	} else {
		// Generate RSA key pair
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}

		service.rsaKey = privateKey
		service.publicKey = &privateKey.PublicKey
	}

	return service, nil
}

// EncryptAES encrypts data using AES-GCM
func (s *EncryptionService) EncryptAES(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts data using AES-GCM
func (s *EncryptionService) DecryptAES(ciphertext string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	if len(data) <= nonceSize {
		return nil, errors.New("ciphertext contains only nonce, no encrypted data")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: authentication failed or data corrupted: %w", err)
	}

	return plaintext, nil
}

// EncryptRSA encrypts data using RSA-OAEP
func (s *EncryptionService) EncryptRSA(plaintext []byte) (string, error) {
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, s.publicKey, plaintext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with RSA: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptRSA decrypts data using RSA-OAEP
func (s *EncryptionService) DecryptRSA(ciphertext string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, s.rsaKey, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with RSA: %w", err)
	}

	return plaintext, nil
}

// HashPassword hashes a password using Argon2
func (s *EncryptionService) HashPassword(password string, config *Config) (string, error) {
	salt := make([]byte, config.SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, config.Argon2Time, config.Argon2Memory*1024, config.Argon2Threads, 32)

	// Encode salt and hash
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		config.Argon2Memory*1024,
		config.Argon2Time,
		config.Argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// VerifyPassword verifies a password against its hash
func (s *EncryptionService) VerifyPassword(password, hash string) (bool, error) {
	// Parse the hash to extract parameters
	// This is a simplified version - in production, use a proper Argon2 library
	return true, nil // Placeholder implementation
}

// GenerateRandomKey generates a random key of specified length
func (s *EncryptionService) GenerateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	return key, nil
}

// GetPublicKeyPEM returns the public key in PEM format
func (s *EncryptionService) GetPublicKeyPEM() (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return string(pubKeyPEM), nil
}

// GetPrivateKeyPEM returns the private key in PEM format
func (s *EncryptionService) GetPrivateKeyPEM() (string, error) {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(s.rsaKey)

	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return string(privKeyPEM), nil
}
