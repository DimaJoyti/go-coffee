package ecc

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Secp256k1 represents the secp256k1 elliptic curve parameters used by Bitcoin
type Secp256k1 struct {
	P  *big.Int // Prime field modulus
	A  *big.Int // Curve parameter a (0 for secp256k1)
	B  *big.Int // Curve parameter b (7 for secp256k1)
	Gx *big.Int // Generator point x coordinate
	Gy *big.Int // Generator point y coordinate
	N  *big.Int // Order of the generator point
	H  *big.Int // Cofactor (1 for secp256k1)
}

// Global secp256k1 instance
var secp256k1 *Secp256k1

func init() {
	secp256k1 = &Secp256k1{}

	// Prime field modulus: p = 2^256 - 2^32 - 2^9 - 2^8 - 2^7 - 2^6 - 2^4 - 1
	secp256k1.P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)

	// Curve parameters for y^2 = x^3 + ax + b
	secp256k1.A = big.NewInt(0)
	secp256k1.B = big.NewInt(7)

	// Generator point coordinates
	secp256k1.Gx, _ = new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	secp256k1.Gy, _ = new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)

	// Order of the generator point
	secp256k1.N, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)

	// Cofactor
	secp256k1.H = big.NewInt(1)
}

// GetSecp256k1 returns the global secp256k1 curve instance
func GetSecp256k1() *Secp256k1 {
	return secp256k1
}

// Generator returns the generator point G
func (curve *Secp256k1) Generator() (*Point, error) {
	return NewPoint(curve.Gx, curve.Gy, curve.A, curve.B)
}

// IsValidPrivateKey checks if a private key is valid (0 < key < n)
func (curve *Secp256k1) IsValidPrivateKey(privateKey *big.Int) bool {
	return privateKey.Cmp(big.NewInt(0)) > 0 && privateKey.Cmp(curve.N) < 0
}

// GeneratePrivateKey generates a cryptographically secure random private key
func (curve *Secp256k1) GeneratePrivateKey() (*big.Int, error) {
	for {
		// Generate random bytes
		privateKeyBytes := make([]byte, 32)
		_, err := rand.Read(privateKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random bytes: %w", err)
		}

		// Convert to big.Int
		privateKey := new(big.Int).SetBytes(privateKeyBytes)

		// Check if valid (0 < privateKey < n)
		if curve.IsValidPrivateKey(privateKey) {
			return privateKey, nil
		}
	}
}

// PrivateKeyToPublicKey converts a private key to its corresponding public key point
func (curve *Secp256k1) PrivateKeyToPublicKey(privateKey *big.Int) (*Point, error) {
	if !curve.IsValidPrivateKey(privateKey) {
		return nil, fmt.Errorf("invalid private key")
	}

	// Get generator point
	generator, err := curve.Generator()
	if err != nil {
		return nil, fmt.Errorf("failed to get generator point: %w", err)
	}

	// Calculate public key = privateKey * G
	publicKey, err := generator.ScalarMult(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate public key: %w", err)
	}

	return publicKey, nil
}

// GenerateKeyPair generates a new private/public key pair
func (curve *Secp256k1) GenerateKeyPair() (*big.Int, *Point, error) {
	// Generate private key
	privateKey, err := curve.GeneratePrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Calculate public key
	publicKey, err := curve.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate public key: %w", err)
	}

	return privateKey, publicKey, nil
}

// ModInverse calculates the modular inverse of a modulo p
func (curve *Secp256k1) ModInverse(a *big.Int) *big.Int {
	return new(big.Int).ModInverse(a, curve.P)
}

// ModAdd performs modular addition (a + b) mod p
func (curve *Secp256k1) ModAdd(a, b *big.Int) *big.Int {
	result := new(big.Int).Add(a, b)
	return result.Mod(result, curve.P)
}

// ModSub performs modular subtraction (a - b) mod p
func (curve *Secp256k1) ModSub(a, b *big.Int) *big.Int {
	result := new(big.Int).Sub(a, b)
	return result.Mod(result, curve.P)
}

// ModMul performs modular multiplication (a * b) mod p
func (curve *Secp256k1) ModMul(a, b *big.Int) *big.Int {
	result := new(big.Int).Mul(a, b)
	return result.Mod(result, curve.P)
}

// ModSquare performs modular squaring (a^2) mod p
func (curve *Secp256k1) ModSquare(a *big.Int) *big.Int {
	result := new(big.Int).Mul(a, a)
	return result.Mod(result, curve.P)
}

// IsOnCurve checks if a point (x, y) is on the secp256k1 curve
func (curve *Secp256k1) IsOnCurve(x, y *big.Int) bool {
	// Calculate y^2 mod p
	ySquared := curve.ModSquare(y)

	// Calculate x^3 + 7 mod p (since a = 0, b = 7 for secp256k1)
	xCubed := curve.ModSquare(x)
	xCubed = curve.ModMul(xCubed, x)
	rightSide := curve.ModAdd(xCubed, curve.B)

	return ySquared.Cmp(rightSide) == 0
}

// CompressPoint compresses a public key point to 33 bytes
func (curve *Secp256k1) CompressPoint(point *Point) []byte {
	if point.IsInfinity() {
		return nil
	}

	// 33 bytes: 1 byte prefix + 32 bytes x coordinate
	compressed := make([]byte, 33)

	// Set prefix based on y coordinate parity
	if new(big.Int).And(point.Y, big.NewInt(1)).Cmp(big.NewInt(0)) == 0 {
		compressed[0] = 0x02 // Even y
	} else {
		compressed[0] = 0x03 // Odd y
	}

	// Copy x coordinate (32 bytes, big-endian)
	xBytes := point.X.Bytes()
	copy(compressed[33-len(xBytes):], xBytes)

	return compressed
}

// DecompressPoint decompresses a 33-byte compressed public key
func (curve *Secp256k1) DecompressPoint(compressed []byte) (*Point, error) {
	if len(compressed) != 33 {
		return nil, fmt.Errorf("compressed point must be 33 bytes")
	}

	prefix := compressed[0]
	if prefix != 0x02 && prefix != 0x03 {
		return nil, fmt.Errorf("invalid compression prefix: %02x", prefix)
	}

	// Extract x coordinate
	x := new(big.Int).SetBytes(compressed[1:])

	// Calculate y^2 = x^3 + 7 mod p
	ySquared := curve.ModSquare(x)
	ySquared = curve.ModMul(ySquared, x)
	ySquared = curve.ModAdd(ySquared, curve.B)

	// Calculate y = sqrt(y^2) mod p using Tonelli-Shanks algorithm
	y := curve.ModSqrt(ySquared)
	if y == nil {
		return nil, fmt.Errorf("point is not on curve")
	}

	// Choose correct y based on parity
	yParity := new(big.Int).And(y, big.NewInt(1)).Uint64()
	expectedParity := uint64(prefix - 0x02)

	if yParity != expectedParity {
		y = curve.ModSub(curve.P, y)
	}

	return NewPoint(x, y, curve.A, curve.B)
}

// ModSqrt calculates the square root of a modulo p using Tonelli-Shanks algorithm
func (curve *Secp256k1) ModSqrt(a *big.Int) *big.Int {
	// For secp256k1, p ≡ 3 (mod 4), so we can use the simple formula
	// sqrt(a) = a^((p+1)/4) mod p
	exp := new(big.Int).Add(curve.P, big.NewInt(1))
	exp.Div(exp, big.NewInt(4))
	
	result := new(big.Int).Exp(a, exp, curve.P)
	
	// Verify the result
	if curve.ModSquare(result).Cmp(a) == 0 {
		return result
	}
	
	return nil
}

// String returns a string representation of the curve parameters
func (curve *Secp256k1) String() string {
	return fmt.Sprintf("secp256k1: y^2 = x^3 + %s (mod %s)", curve.B.String(), curve.P.String())
}
