package models

// GenerateKeyPairRequest represents a request to generate a key pair
type GenerateKeyPairRequest struct {
	Chain Chain `json:"chain"`
}

// GenerateKeyPairResponse represents the response from generating a key pair
type GenerateKeyPairResponse struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

// EncryptPrivateKeyRequest represents a request to encrypt a private key
type EncryptPrivateKeyRequest struct {
	PrivateKey string `json:"private_key"`
	Passphrase string `json:"passphrase"`
}

// EncryptPrivateKeyResponse represents the response from encrypting a private key
type EncryptPrivateKeyResponse struct {
	EncryptedKey string `json:"encrypted_key"`
}

// DecryptPrivateKeyRequest represents a request to decrypt a private key
type DecryptPrivateKeyRequest struct {
	EncryptedKey string `json:"encrypted_key"`
	Passphrase   string `json:"passphrase"`
}

// DecryptPrivateKeyResponse represents the response from decrypting a private key
type DecryptPrivateKeyResponse struct {
	PrivateKey string `json:"private_key"`
}

// GenerateJWTRequest represents a request to generate a JWT token
type GenerateJWTRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// GenerateJWTResponse represents the response from generating a JWT token
type GenerateJWTResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// VerifyJWTRequest represents a request to verify a JWT token
type VerifyJWTRequest struct {
	Token string `json:"token"`
}

// VerifyJWTResponse represents the response from verifying a JWT token
type VerifyJWTResponse struct {
	Valid     bool   `json:"valid"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

// GenerateMnemonicRequest represents a request to generate a mnemonic phrase
type GenerateMnemonicRequest struct {
	Strength int `json:"strength"`
}

// GenerateMnemonicResponse represents the response from generating a mnemonic phrase
type GenerateMnemonicResponse struct {
	Mnemonic string `json:"mnemonic"`
}

// ValidateMnemonicRequest represents a request to validate a mnemonic phrase
type ValidateMnemonicRequest struct {
	Mnemonic string `json:"mnemonic"`
}

// ValidateMnemonicResponse represents the response from validating a mnemonic phrase
type ValidateMnemonicResponse struct {
	Valid bool `json:"valid"`
}

// MnemonicToPrivateKeyRequest represents a request to convert a mnemonic to a private key
type MnemonicToPrivateKeyRequest struct {
	Mnemonic string `json:"mnemonic"`
	Path     string `json:"path"`
}

// MnemonicToPrivateKeyResponse represents the response from converting a mnemonic to a private key
type MnemonicToPrivateKeyResponse struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}