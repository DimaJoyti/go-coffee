package base58

import (
	"crypto/sha256"
	"fmt"
	"math/big"
)

// Base58 alphabet used by Bitcoin (excludes 0, O, I, l to avoid confusion)
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var (
	// Base58 alphabet as byte slice for faster lookup
	alphabet = []byte(base58Alphabet)
	// Reverse lookup table for decoding
	decodeTable [256]int
)

func init() {
	// Initialize decode table
	for i := range decodeTable {
		decodeTable[i] = -1
	}
	for i, char := range alphabet {
		decodeTable[char] = i
	}
}

// Encode encodes a byte slice to Base58 string
func Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	// Count leading zeros
	leadingZeros := 0
	for i := 0; i < len(input) && input[i] == 0; i++ {
		leadingZeros++
	}

	// Convert to big integer
	num := new(big.Int).SetBytes(input)
	
	// Convert to base58
	var result []byte
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append(result, alphabet[mod.Int64()])
	}

	// Add leading zeros as '1' characters
	for i := 0; i < leadingZeros; i++ {
		result = append(result, alphabet[0])
	}

	// Reverse the result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// Decode decodes a Base58 string to byte slice
func Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return []byte{}, nil
	}

	// Count leading '1' characters (representing leading zeros)
	leadingOnes := 0
	for i := 0; i < len(input) && input[i] == alphabet[0]; i++ {
		leadingOnes++
	}

	// Convert from base58
	num := big.NewInt(0)
	base := big.NewInt(58)
	
	for _, char := range input {
		if char > 255 || decodeTable[char] == -1 {
			return nil, fmt.Errorf("invalid character in base58 string: %c", char)
		}
		
		num.Mul(num, base)
		num.Add(num, big.NewInt(int64(decodeTable[char])))
	}

	// Convert to bytes
	decoded := num.Bytes()

	// Add leading zeros
	result := make([]byte, leadingOnes+len(decoded))
	copy(result[leadingOnes:], decoded)

	return result, nil
}

// EncodeCheck encodes a byte slice with Base58Check encoding (includes checksum)
func EncodeCheck(input []byte) string {
	// Calculate checksum (double SHA256)
	checksum := calculateChecksum(input)
	
	// Append checksum to input
	payload := append(input, checksum...)
	
	// Encode with Base58
	return Encode(payload)
}

// DecodeCheck decodes a Base58Check encoded string and verifies checksum
func DecodeCheck(input string) ([]byte, error) {
	// Decode from Base58
	decoded, err := Decode(input)
	if err != nil {
		return nil, fmt.Errorf("base58 decode error: %w", err)
	}

	// Check minimum length (payload + 4 byte checksum)
	if len(decoded) < 4 {
		return nil, fmt.Errorf("decoded data too short")
	}

	// Split payload and checksum
	payload := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]

	// Verify checksum
	expectedChecksum := calculateChecksum(payload)
	for i := 0; i < 4; i++ {
		if checksum[i] != expectedChecksum[i] {
			return nil, fmt.Errorf("checksum verification failed")
		}
	}

	return payload, nil
}

// calculateChecksum calculates the 4-byte checksum for Base58Check
func calculateChecksum(data []byte) []byte {
	// Double SHA256
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	
	// Return first 4 bytes
	return second[:4]
}

// IsValid checks if a Base58Check encoded string is valid
func IsValid(input string) bool {
	_, err := DecodeCheck(input)
	return err == nil
}

// ValidateBase58 checks if a string contains only valid Base58 characters
func ValidateBase58(input string) bool {
	for _, char := range input {
		if char > 255 || decodeTable[char] == -1 {
			return false
		}
	}
	return true
}

// Bitcoin address version bytes
const (
	// Mainnet
	MainnetP2PKHVersion byte = 0x00 // 1...
	MainnetP2SHVersion  byte = 0x05 // 3...

	// Testnet
	TestnetP2PKHVersion byte = 0x6F // m... or n...
	TestnetP2SHVersion  byte = 0xC4 // 2...

	// Private key WIF (Wallet Import Format)
	MainnetWIFVersion byte = 0x80 // 5... (uncompressed) or K/L... (compressed)
	TestnetWIFVersion byte = 0xEF // 9... or c...
)

// EncodeAddress encodes a hash160 to a Bitcoin address
func EncodeAddress(hash160 []byte, version byte) string {
	if len(hash160) != 20 {
		return ""
	}
	
	// Prepend version byte
	payload := append([]byte{version}, hash160...)
	
	// Encode with Base58Check
	return EncodeCheck(payload)
}

// DecodeAddress decodes a Bitcoin address to hash160 and version
func DecodeAddress(address string) (hash160 []byte, version byte, err error) {
	// Decode with Base58Check
	decoded, err := DecodeCheck(address)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid address: %w", err)
	}

	// Check length (1 byte version + 20 bytes hash160)
	if len(decoded) != 21 {
		return nil, 0, fmt.Errorf("invalid address length")
	}

	version = decoded[0]
	hash160 = decoded[1:]

	return hash160, version, nil
}

// EncodeWIF encodes a private key in Wallet Import Format
func EncodeWIF(privateKey []byte, compressed bool, testnet bool) string {
	if len(privateKey) != 32 {
		return ""
	}

	// Choose version byte
	version := MainnetWIFVersion
	if testnet {
		version = TestnetWIFVersion
	}

	// Build payload
	payload := append([]byte{version}, privateKey...)
	
	// Add compression flag if compressed
	if compressed {
		payload = append(payload, 0x01)
	}

	// Encode with Base58Check
	return EncodeCheck(payload)
}

// DecodeWIF decodes a private key from Wallet Import Format
func DecodeWIF(wif string) (privateKey []byte, compressed bool, testnet bool, err error) {
	// Decode with Base58Check
	decoded, err := DecodeCheck(wif)
	if err != nil {
		return nil, false, false, fmt.Errorf("invalid WIF: %w", err)
	}

	// Check length (33 bytes for uncompressed, 34 bytes for compressed)
	if len(decoded) != 33 && len(decoded) != 34 {
		return nil, false, false, fmt.Errorf("invalid WIF length")
	}

	version := decoded[0]
	privateKey = decoded[1:33]

	// Check compression flag
	compressed = len(decoded) == 34 && decoded[33] == 0x01

	// Determine network
	testnet = version == TestnetWIFVersion

	// Validate version
	if version != MainnetWIFVersion && version != TestnetWIFVersion {
		return nil, false, false, fmt.Errorf("invalid WIF version: 0x%02x", version)
	}

	return privateKey, compressed, testnet, nil
}

// GetAddressType returns the type of Bitcoin address
func GetAddressType(address string) string {
	hash160, version, err := DecodeAddress(address)
	if err != nil {
		return "invalid"
	}

	if len(hash160) != 20 {
		return "invalid"
	}

	switch version {
	case MainnetP2PKHVersion:
		return "P2PKH-mainnet"
	case MainnetP2SHVersion:
		return "P2SH-mainnet"
	case TestnetP2PKHVersion:
		return "P2PKH-testnet"
	case TestnetP2SHVersion:
		return "P2SH-testnet"
	default:
		return "unknown"
	}
}

// IsMainnetAddress checks if an address is for mainnet
func IsMainnetAddress(address string) bool {
	_, version, err := DecodeAddress(address)
	if err != nil {
		return false
	}
	return version == MainnetP2PKHVersion || version == MainnetP2SHVersion
}

// IsTestnetAddress checks if an address is for testnet
func IsTestnetAddress(address string) bool {
	_, version, err := DecodeAddress(address)
	if err != nil {
		return false
	}
	return version == TestnetP2PKHVersion || version == TestnetP2SHVersion
}
