package sec

import (
	"fmt"
	"math/big"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/ecc"
)

// SEC1 format constants
const (
	// Uncompressed public key prefix
	UncompressedPrefix = 0x04
	// Compressed public key prefixes
	CompressedEvenPrefix = 0x02
	CompressedOddPrefix  = 0x03
	// Key sizes
	UncompressedKeySize = 65 // 1 + 32 + 32 bytes
	CompressedKeySize   = 33 // 1 + 32 bytes
	PrivateKeySize      = 32 // 32 bytes
)

// EncodePublicKeyUncompressed encodes a public key in uncompressed SEC1 format
func EncodePublicKeyUncompressed(publicKey *ecc.Point) []byte {
	if publicKey.IsInfinity() {
		return nil
	}

	// 65 bytes: 0x04 + 32 bytes x + 32 bytes y
	encoded := make([]byte, UncompressedKeySize)
	encoded[0] = UncompressedPrefix

	// Copy x coordinate (32 bytes, big-endian)
	xBytes := publicKey.X.Bytes()
	copy(encoded[1+32-len(xBytes):33], xBytes)

	// Copy y coordinate (32 bytes, big-endian)
	yBytes := publicKey.Y.Bytes()
	copy(encoded[33+32-len(yBytes):65], yBytes)

	return encoded
}

// EncodePublicKeyCompressed encodes a public key in compressed SEC1 format
func EncodePublicKeyCompressed(publicKey *ecc.Point) []byte {
	if publicKey.IsInfinity() {
		return nil
	}

	// 33 bytes: prefix + 32 bytes x
	encoded := make([]byte, CompressedKeySize)

	// Set prefix based on y coordinate parity
	if new(big.Int).And(publicKey.Y, big.NewInt(1)).Cmp(big.NewInt(0)) == 0 {
		encoded[0] = CompressedEvenPrefix // Even y
	} else {
		encoded[0] = CompressedOddPrefix // Odd y
	}

	// Copy x coordinate (32 bytes, big-endian)
	xBytes := publicKey.X.Bytes()
	copy(encoded[1+32-len(xBytes):33], xBytes)

	return encoded
}

// DecodePublicKey decodes a public key from SEC1 format (compressed or uncompressed)
func DecodePublicKey(encoded []byte) (*ecc.Point, error) {
	if len(encoded) == 0 {
		return nil, fmt.Errorf("empty encoded public key")
	}

	curve := ecc.GetSecp256k1()

	switch encoded[0] {
	case UncompressedPrefix:
		return decodeUncompressedPublicKey(encoded, curve)
	case CompressedEvenPrefix, CompressedOddPrefix:
		return decodeCompressedPublicKey(encoded, curve)
	default:
		return nil, fmt.Errorf("invalid public key prefix: 0x%02x", encoded[0])
	}
}

// decodeUncompressedPublicKey decodes an uncompressed public key
func decodeUncompressedPublicKey(encoded []byte, curve *ecc.Secp256k1) (*ecc.Point, error) {
	if len(encoded) != UncompressedKeySize {
		return nil, fmt.Errorf("uncompressed public key must be %d bytes, got %d", UncompressedKeySize, len(encoded))
	}

	// Extract coordinates
	x := new(big.Int).SetBytes(encoded[1:33])
	y := new(big.Int).SetBytes(encoded[33:65])

	// Create point
	point, err := ecc.NewPoint(x, y, curve.A, curve.B)
	if err != nil {
		return nil, fmt.Errorf("invalid public key point: %w", err)
	}

	return point, nil
}

// decodeCompressedPublicKey decodes a compressed public key
func decodeCompressedPublicKey(encoded []byte, curve *ecc.Secp256k1) (*ecc.Point, error) {
	if len(encoded) != CompressedKeySize {
		return nil, fmt.Errorf("compressed public key must be %d bytes, got %d", CompressedKeySize, len(encoded))
	}

	prefix := encoded[0]
	if prefix != CompressedEvenPrefix && prefix != CompressedOddPrefix {
		return nil, fmt.Errorf("invalid compressed public key prefix: 0x%02x", prefix)
	}

	// Extract x coordinate
	x := new(big.Int).SetBytes(encoded[1:])

	// Calculate y^2 = x^3 + 7 mod p
	ySquared := curve.ModSquare(x)
	ySquared = curve.ModMul(ySquared, x)
	ySquared = curve.ModAdd(ySquared, curve.B)

	// Calculate y = sqrt(y^2) mod p
	y := curve.ModSqrt(ySquared)
	if y == nil {
		return nil, fmt.Errorf("point is not on curve")
	}

	// Choose correct y based on parity
	yParity := new(big.Int).And(y, big.NewInt(1)).Uint64()
	expectedParity := uint64(prefix - CompressedEvenPrefix)

	if yParity != expectedParity {
		y = curve.ModSub(curve.P, y)
	}

	// Create point
	point, err := ecc.NewPoint(x, y, curve.A, curve.B)
	if err != nil {
		return nil, fmt.Errorf("invalid public key point: %w", err)
	}

	return point, nil
}

// EncodePrivateKey encodes a private key as 32 bytes
func EncodePrivateKey(privateKey *big.Int) []byte {
	// 32 bytes, big-endian
	encoded := make([]byte, PrivateKeySize)
	privateKeyBytes := privateKey.Bytes()
	copy(encoded[PrivateKeySize-len(privateKeyBytes):], privateKeyBytes)
	return encoded
}

// DecodePrivateKey decodes a private key from 32 bytes
func DecodePrivateKey(encoded []byte) (*big.Int, error) {
	if len(encoded) != PrivateKeySize {
		return nil, fmt.Errorf("private key must be %d bytes, got %d", PrivateKeySize, len(encoded))
	}

	privateKey := new(big.Int).SetBytes(encoded)

	// Validate private key
	curve := ecc.GetSecp256k1()
	if !curve.IsValidPrivateKey(privateKey) {
		return nil, fmt.Errorf("invalid private key value")
	}

	return privateKey, nil
}

// IsCompressed checks if a public key is in compressed format
func IsCompressed(encoded []byte) bool {
	if len(encoded) == 0 {
		return false
	}
	return encoded[0] == CompressedEvenPrefix || encoded[0] == CompressedOddPrefix
}

// IsUncompressed checks if a public key is in uncompressed format
func IsUncompressed(encoded []byte) bool {
	if len(encoded) == 0 {
		return false
	}
	return encoded[0] == UncompressedPrefix
}

// ValidatePublicKey validates a public key in SEC1 format
func ValidatePublicKey(encoded []byte) error {
	_, err := DecodePublicKey(encoded)
	return err
}

// ValidatePrivateKey validates a private key
func ValidatePrivateKey(encoded []byte) error {
	_, err := DecodePrivateKey(encoded)
	return err
}

// CompressPublicKey converts an uncompressed public key to compressed format
func CompressPublicKey(uncompressed []byte) ([]byte, error) {
	if len(uncompressed) != UncompressedKeySize || uncompressed[0] != UncompressedPrefix {
		return nil, fmt.Errorf("input is not an uncompressed public key")
	}

	// Decode the uncompressed key
	point, err := DecodePublicKey(uncompressed)
	if err != nil {
		return nil, fmt.Errorf("failed to decode uncompressed key: %w", err)
	}

	// Encode as compressed
	return EncodePublicKeyCompressed(point), nil
}

// DecompressPublicKey converts a compressed public key to uncompressed format
func DecompressPublicKey(compressed []byte) ([]byte, error) {
	if len(compressed) != CompressedKeySize {
		return nil, fmt.Errorf("input is not a compressed public key")
	}

	if compressed[0] != CompressedEvenPrefix && compressed[0] != CompressedOddPrefix {
		return nil, fmt.Errorf("input is not a compressed public key")
	}

	// Decode the compressed key
	point, err := DecodePublicKey(compressed)
	if err != nil {
		return nil, fmt.Errorf("failed to decode compressed key: %w", err)
	}

	// Encode as uncompressed
	return EncodePublicKeyUncompressed(point), nil
}

// GetPublicKeyFormat returns the format of a public key
func GetPublicKeyFormat(encoded []byte) string {
	if len(encoded) == 0 {
		return "invalid"
	}

	switch encoded[0] {
	case UncompressedPrefix:
		if len(encoded) == UncompressedKeySize {
			return "uncompressed"
		}
		return "invalid"
	case CompressedEvenPrefix, CompressedOddPrefix:
		if len(encoded) == CompressedKeySize {
			return "compressed"
		}
		return "invalid"
	default:
		return "invalid"
	}
}

// PublicKeyFromPrivateKey generates a public key from a private key
func PublicKeyFromPrivateKey(privateKey *big.Int, compressed bool) ([]byte, error) {
	curve := ecc.GetSecp256k1()

	// Generate public key point
	publicKeyPoint, err := curve.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	// Encode based on compression preference
	if compressed {
		return EncodePublicKeyCompressed(publicKeyPoint), nil
	}
	return EncodePublicKeyUncompressed(publicKeyPoint), nil
}

// KeyPairFromPrivateKey generates both compressed and uncompressed public keys from a private key
func KeyPairFromPrivateKey(privateKey *big.Int) (compressed, uncompressed []byte, err error) {
	curve := ecc.GetSecp256k1()

	// Generate public key point
	publicKeyPoint, err := curve.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	// Encode both formats
	compressed = EncodePublicKeyCompressed(publicKeyPoint)
	uncompressed = EncodePublicKeyUncompressed(publicKeyPoint)

	return compressed, uncompressed, nil
}
