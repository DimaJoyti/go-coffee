package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

// KeyManager handles cryptographic key operations
type KeyManager struct {
	keystore *keystore.KeyStore
}

// NewKeyManager creates a new key manager
func NewKeyManager(keystorePath string) *KeyManager {
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	return &KeyManager{
		keystore: ks,
	}
}

// GenerateKeyPair generates a new private/public key pair
func (km *KeyManager) GenerateKeyPair() (privateKey, publicKey, address string, err error) {
	// Generate private key
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Get private key in hex
	privateKey = hexutil.Encode(crypto.FromECDSA(key))

	// Get public key in hex
	publicKeyBytes := crypto.FromECDSAPub(&key.PublicKey)
	publicKey = hexutil.Encode(publicKeyBytes)

	// Get Ethereum address
	address = crypto.PubkeyToAddress(key.PublicKey).Hex()

	return privateKey, publicKey, address, nil
}

// ImportPrivateKey imports a private key
func (km *KeyManager) ImportPrivateKey(privateKeyHex string, passphrase string) (string, error) {
	// Decode private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex[2:]) // Remove "0x" prefix
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Import key to keystore
	account, err := km.keystore.ImportECDSA(privateKey, passphrase)
	if err != nil {
		return "", fmt.Errorf("failed to import private key: %w", err)
	}

	return account.Address.Hex(), nil
}

// ExportPrivateKey exports a private key
func (km *KeyManager) ExportPrivateKey(address, passphrase string) (string, error) {
	// Get account
	account, err := km.getAccount(address)
	if err != nil {
		return "", err
	}

	// Get key file
	keyJSON, err := km.keystore.Export(account, passphrase, passphrase)
	if err != nil {
		return "", fmt.Errorf("failed to export key: %w", err)
	}

	// Parse key file
	key, err := keystore.DecryptKey(keyJSON, passphrase)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt key: %w", err)
	}

	// Get private key in hex
	privateKey := hexutil.Encode(crypto.FromECDSA(key.PrivateKey))

	return privateKey, nil
}

// EncryptPrivateKey encrypts a private key with a passphrase
func (km *KeyManager) EncryptPrivateKey(privateKeyHex, passphrase string) (string, error) {
	// Decode private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex[2:]) // Remove "0x" prefix
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Create temporary account
	account, err := km.keystore.ImportECDSA(privateKey, passphrase)
	if err != nil {
		return "", fmt.Errorf("failed to import private key: %w", err)
	}

	// Get key file
	keyJSON, err := km.keystore.Export(account, passphrase, passphrase)
	if err != nil {
		return "", fmt.Errorf("failed to export key: %w", err)
	}

	// Delete temporary account
	err = km.keystore.Delete(account, passphrase)
	if err != nil {
		return "", fmt.Errorf("failed to delete temporary account: %w", err)
	}

	return string(keyJSON), nil
}

// DecryptPrivateKey decrypts a private key with a passphrase
func (km *KeyManager) DecryptPrivateKey(encryptedKey, passphrase string) (string, error) {
	// Parse key file
	key, err := keystore.DecryptKey([]byte(encryptedKey), passphrase)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt key: %w", err)
	}

	// Get private key in hex
	privateKey := hexutil.Encode(crypto.FromECDSA(key.PrivateKey))

	return privateKey, nil
}

// GenerateMnemonic generates a new mnemonic phrase
func (km *KeyManager) GenerateMnemonic() (string, error) {
	// Generate entropy
	entropy, err := bip39.NewEntropy(256) // 24 words
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	// Generate mnemonic
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// ValidateMnemonic validates a mnemonic phrase
func (km *KeyManager) ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// MnemonicToPrivateKey converts a mnemonic to a private key
func (km *KeyManager) MnemonicToPrivateKey(mnemonic string, path string) (string, error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", errors.New("invalid mnemonic")
	}

	// Generate seed
	seed := bip39.NewSeed(mnemonic, "")

	// Parse derivation path
	derivationPath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse derivation path: %w", err)
	}

	// Derive private key
	privateKey, err := km.derivePrivateKey(seed, derivationPath)
	if err != nil {
		return "", fmt.Errorf("failed to derive private key: %w", err)
	}

	// Get private key in hex
	privateKeyHex := hexutil.Encode(crypto.FromECDSA(privateKey))

	return privateKeyHex, nil
}

// getAccount gets an account by address
func (km *KeyManager) getAccount(address string) (accounts.Account, error) {
	// Get all accounts
	accts := km.keystore.Accounts()

	// Find account by address
	for _, acct := range accts {
		if acct.Address.Hex() == address {
			return acct, nil
		}
	}

	return accounts.Account{}, fmt.Errorf("account not found: %s", address)
}

// derivePrivateKey derives a private key from seed and derivation path
func (km *KeyManager) derivePrivateKey(seed []byte, derivationPath accounts.DerivationPath) (*ecdsa.PrivateKey, error) {
	// Simplified implementation - in production, use proper HD wallet derivation
	// For now, just generate a key from the seed

	// Use seed to generate a deterministic private key
	hash := sha256.Sum256(seed)

	// Generate private key from hash
	privateKey, err := crypto.ToECDSA(hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	return privateKey, nil
}

// EncryptData encrypts data with a key
func EncryptData(data, key []byte) ([]byte, error) {
	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to create nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// DecryptData decrypts data with a key
func DecryptData(data, key []byte) ([]byte, error) {
	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Get nonce size
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// DeriveKey derives a key from a passphrase
func DeriveKey(passphrase string, salt []byte, iterations, keyLen int) []byte {
	return pbkdf2.Key([]byte(passphrase), salt, iterations, keyLen, sha256.New)
}

// GenerateRandomBytes generates random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomHex generates a random hex string
func GenerateRandomHex(n int) (string, error) {
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
