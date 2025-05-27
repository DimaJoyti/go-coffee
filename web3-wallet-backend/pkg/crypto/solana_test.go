package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyManager_GenerateSolanaKeyPair(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Generate Solana key pair
	privateKey, publicKey, address, err := km.GenerateSolanaKeyPair()
	require.NoError(t, err)

	// Verify keys are not empty
	assert.NotEmpty(t, privateKey)
	assert.NotEmpty(t, publicKey)
	assert.NotEmpty(t, address)

	// Verify address equals public key (in Solana)
	assert.Equal(t, publicKey, address)

	// Verify private key format (base58)
	assert.True(t, len(privateKey) > 40)
	assert.True(t, len(publicKey) > 40)
}

func TestKeyManager_ImportSolanaPrivateKey(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Generate a key pair first
	privateKey, _, expectedAddress, err := km.GenerateSolanaKeyPair()
	require.NoError(t, err)

	// Import the private key
	address, err := km.ImportSolanaPrivateKey(privateKey)
	require.NoError(t, err)

	// Verify the address matches
	assert.Equal(t, expectedAddress, address)
}

func TestKeyManager_ImportSolanaPrivateKey_InvalidKey(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Try to import invalid private key
	_, err := km.ImportSolanaPrivateKey("invalid-private-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse Solana private key")
}

func TestKeyManager_GenerateSolanaKeyPairFromSeed(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Create test seed (32 bytes)
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}

	// Generate key pair from seed
	privateKey, publicKey, address, err := km.GenerateSolanaKeyPairFromSeed(seed)
	require.NoError(t, err)

	// Verify keys are not empty
	assert.NotEmpty(t, privateKey)
	assert.NotEmpty(t, publicKey)
	assert.NotEmpty(t, address)

	// Verify address equals public key
	assert.Equal(t, publicKey, address)

	// Generate again with same seed - should be deterministic
	privateKey2, publicKey2, address2, err := km.GenerateSolanaKeyPairFromSeed(seed)
	require.NoError(t, err)

	assert.Equal(t, privateKey, privateKey2)
	assert.Equal(t, publicKey, publicKey2)
	assert.Equal(t, address, address2)
}

func TestKeyManager_GenerateSolanaKeyPairFromSeed_ShortSeed(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Create short seed (less than 32 bytes)
	seed := make([]byte, 16)

	// Should fail with short seed
	_, _, _, err := km.GenerateSolanaKeyPairFromSeed(seed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "seed must be at least 32 bytes")
}

func TestKeyManager_SolanaMnemonicToPrivateKey(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Generate valid mnemonic
	mnemonic, err := km.GenerateMnemonic()
	require.NoError(t, err)

	// Convert mnemonic to private key
	privateKey, err := km.SolanaMnemonicToPrivateKey(mnemonic, "m/44'/501'/0'/0'")
	require.NoError(t, err)

	// Verify private key is not empty
	assert.NotEmpty(t, privateKey)

	// Should be deterministic - same mnemonic should produce same key
	privateKey2, err := km.SolanaMnemonicToPrivateKey(mnemonic, "m/44'/501'/0'/0'")
	require.NoError(t, err)
	assert.Equal(t, privateKey, privateKey2)
}

func TestKeyManager_SolanaMnemonicToPrivateKey_InvalidMnemonic(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Try with invalid mnemonic
	_, err := km.SolanaMnemonicToPrivateKey("invalid mnemonic phrase", "m/44'/501'/0'/0'")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mnemonic")
}

func TestKeyManager_SolanaKeyPair_Uniqueness(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Generate multiple key pairs
	keys := make(map[string]bool)
	addresses := make(map[string]bool)

	for i := 0; i < 10; i++ {
		privateKey, _, address, err := km.GenerateSolanaKeyPair()
		require.NoError(t, err)

		// Verify uniqueness
		assert.False(t, keys[privateKey], "Private key should be unique")
		assert.False(t, addresses[address], "Address should be unique")

		keys[privateKey] = true
		addresses[address] = true
	}
}

func TestKeyManager_SolanaKeyPair_ValidFormat(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Generate key pair
	privateKey, publicKey, address, err := km.GenerateSolanaKeyPair()
	require.NoError(t, err)

	// Verify base58 format (should not contain invalid characters)
	validBase58Chars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	
	for _, char := range privateKey {
		assert.Contains(t, validBase58Chars, string(char), "Private key should be valid base58")
	}
	
	for _, char := range publicKey {
		assert.Contains(t, validBase58Chars, string(char), "Public key should be valid base58")
	}
	
	for _, char := range address {
		assert.Contains(t, validBase58Chars, string(char), "Address should be valid base58")
	}
}

func TestKeyManager_SolanaKeyPair_Length(t *testing.T) {
	// Create key manager
	km := NewKeyManager("./test-keystore")

	// Generate key pair
	privateKey, publicKey, address, err := km.GenerateSolanaKeyPair()
	require.NoError(t, err)

	// Verify typical Solana key lengths (base58 encoded)
	// Private key: 64 bytes -> ~88 chars in base58
	// Public key: 32 bytes -> ~44 chars in base58
	assert.True(t, len(privateKey) >= 80 && len(privateKey) <= 90, "Private key length should be around 88 chars")
	assert.True(t, len(publicKey) >= 40 && len(publicKey) <= 50, "Public key length should be around 44 chars")
	assert.Equal(t, publicKey, address, "Address should equal public key")
}

func BenchmarkKeyManager_GenerateSolanaKeyPair(b *testing.B) {
	km := NewKeyManager("./test-keystore")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, err := km.GenerateSolanaKeyPair()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkKeyManager_ImportSolanaPrivateKey(b *testing.B) {
	km := NewKeyManager("./test-keystore")
	
	// Generate a key pair for testing
	privateKey, _, _, err := km.GenerateSolanaKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := km.ImportSolanaPrivateKey(privateKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}
