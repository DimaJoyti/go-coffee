package ecc

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

// Signature represents an ECDSA signature
type Signature struct {
	R *big.Int // R component of the signature
	S *big.Int // S component of the signature
}

// NewSignature creates a new ECDSA signature
func NewSignature(r, s *big.Int) *Signature {
	return &Signature{
		R: new(big.Int).Set(r),
		S: new(big.Int).Set(s),
	}
}

// Sign creates an ECDSA signature for a message hash using a private key
func Sign(privateKey *big.Int, messageHash []byte) (*Signature, error) {
	curve := GetSecp256k1()

	// Validate private key
	if !curve.IsValidPrivateKey(privateKey) {
		return nil, fmt.Errorf("invalid private key")
	}

	// Convert message hash to big.Int
	z := new(big.Int).SetBytes(messageHash)

	// Ensure z is in the valid range [1, n-1]
	if z.Cmp(curve.N) >= 0 {
		z.Mod(z, curve.N)
	}

	for {
		// Generate random k in range [1, n-1]
		k, err := generateRandomK(curve.N)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random k: %w", err)
		}

		// Calculate R = k * G
		generator, err := curve.Generator()
		if err != nil {
			return nil, fmt.Errorf("failed to get generator: %w", err)
		}

		R, err := generator.ScalarMult(k)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate R: %w", err)
		}

		// r = R.x mod n
		r := new(big.Int).Mod(R.X, curve.N)

		// If r = 0, try again with different k
		if r.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		// Calculate s = k^(-1) * (z + r * privateKey) mod n
		kInv := new(big.Int).ModInverse(k, curve.N)
		if kInv == nil {
			continue // k is not invertible, try again
		}

		rPrivateKey := new(big.Int).Mul(r, privateKey)
		rPrivateKey.Mod(rPrivateKey, curve.N)

		zPlusRPrivateKey := new(big.Int).Add(z, rPrivateKey)
		zPlusRPrivateKey.Mod(zPlusRPrivateKey, curve.N)

		s := new(big.Int).Mul(kInv, zPlusRPrivateKey)
		s.Mod(s, curve.N)

		// If s = 0, try again with different k
		if s.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		// Use low-s canonical form (BIP 62)
		halfN := new(big.Int).Div(curve.N, big.NewInt(2))
		if s.Cmp(halfN) > 0 {
			s.Sub(curve.N, s)
		}

		return NewSignature(r, s), nil
	}
}

// Verify verifies an ECDSA signature against a message hash and public key
func (sig *Signature) Verify(publicKey *Point, messageHash []byte) bool {
	curve := GetSecp256k1()

	// Validate signature components
	if sig.R.Cmp(big.NewInt(1)) < 0 || sig.R.Cmp(curve.N) >= 0 {
		return false
	}
	if sig.S.Cmp(big.NewInt(1)) < 0 || sig.S.Cmp(curve.N) >= 0 {
		return false
	}

	// Validate public key
	if publicKey.IsInfinity() || !publicKey.IsOnCurve() {
		return false
	}

	// Convert message hash to big.Int
	z := new(big.Int).SetBytes(messageHash)

	// Ensure z is in the valid range [0, n-1]
	if z.Cmp(curve.N) >= 0 {
		z.Mod(z, curve.N)
	}

	// Calculate w = s^(-1) mod n
	w := new(big.Int).ModInverse(sig.S, curve.N)
	if w == nil {
		return false
	}

	// Calculate u1 = z * w mod n
	u1 := new(big.Int).Mul(z, w)
	u1.Mod(u1, curve.N)

	// Calculate u2 = r * w mod n
	u2 := new(big.Int).Mul(sig.R, w)
	u2.Mod(u2, curve.N)

	// Calculate point = u1 * G + u2 * publicKey
	generator, err := curve.Generator()
	if err != nil {
		return false
	}

	// u1 * G
	u1G, err := generator.ScalarMult(u1)
	if err != nil {
		return false
	}

	// u2 * publicKey
	u2PublicKey, err := publicKey.ScalarMult(u2)
	if err != nil {
		return false
	}

	// u1 * G + u2 * publicKey
	point, err := u1G.Add(u2PublicKey)
	if err != nil {
		return false
	}

	// If point is at infinity, signature is invalid
	if point.IsInfinity() {
		return false
	}

	// Verify that point.x mod n == r
	pointX := new(big.Int).Mod(point.X, curve.N)
	return pointX.Cmp(sig.R) == 0
}

// RecoverPublicKey recovers the public key from a signature and message hash
func (sig *Signature) RecoverPublicKey(messageHash []byte, recoveryID int) (*Point, error) {
	curve := GetSecp256k1()

	// Validate recovery ID (0, 1, 2, or 3)
	if recoveryID < 0 || recoveryID > 3 {
		return nil, fmt.Errorf("invalid recovery ID: %d", recoveryID)
	}

	// Calculate R point from signature
	x := new(big.Int).Set(sig.R)
	if recoveryID >= 2 {
		x.Add(x, curve.N)
	}

	// Check if x is valid
	if x.Cmp(curve.P) >= 0 {
		return nil, fmt.Errorf("invalid x coordinate")
	}

	// Calculate y^2 = x^3 + 7 mod p
	ySquared := curve.ModSquare(x)
	ySquared = curve.ModMul(ySquared, x)
	ySquared = curve.ModAdd(ySquared, curve.B)

	// Calculate y
	y := curve.ModSqrt(ySquared)
	if y == nil {
		return nil, fmt.Errorf("point is not on curve")
	}

	// Choose correct y based on recovery ID
	yParity := new(big.Int).And(y, big.NewInt(1)).Uint64()
	expectedParity := uint64(recoveryID & 1)

	if yParity != expectedParity {
		y = curve.ModSub(curve.P, y)
	}

	// Create R point
	R, err := NewPoint(x, y, curve.A, curve.B)
	if err != nil {
		return nil, fmt.Errorf("failed to create R point: %w", err)
	}

	// Convert message hash to big.Int
	z := new(big.Int).SetBytes(messageHash)
	if z.Cmp(curve.N) >= 0 {
		z.Mod(z, curve.N)
	}

	// Calculate r^(-1) mod n
	rInv := new(big.Int).ModInverse(sig.R, curve.N)
	if rInv == nil {
		return nil, fmt.Errorf("r is not invertible")
	}

	// Calculate public key = r^(-1) * (s * R - z * G)
	// First calculate s * R
	sR, err := R.ScalarMult(sig.S)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate s * R: %w", err)
	}

	// Calculate z * G
	generator, err := curve.Generator()
	if err != nil {
		return nil, fmt.Errorf("failed to get generator: %w", err)
	}

	zG, err := generator.ScalarMult(z)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate z * G: %w", err)
	}

	// Calculate -z * G (negate y coordinate)
	negZG := &Point{
		X: new(big.Int).Set(zG.X),
		Y: curve.ModSub(curve.P, zG.Y),
		A: new(big.Int).Set(zG.A),
		B: new(big.Int).Set(zG.B),
	}

	// Calculate s * R - z * G
	sRMinusZG, err := sR.Add(negZG)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate s * R - z * G: %w", err)
	}

	// Calculate public key = r^(-1) * (s * R - z * G)
	publicKey, err := sRMinusZG.ScalarMult(rInv)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate public key: %w", err)
	}

	return publicKey, nil
}

// generateRandomK generates a cryptographically secure random k in range [1, max-1]
func generateRandomK(max *big.Int) (*big.Int, error) {
	for {
		// Generate random bytes
		kBytes := make([]byte, 32)
		_, err := rand.Read(kBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random bytes: %w", err)
		}

		// Convert to big.Int
		k := new(big.Int).SetBytes(kBytes)

		// Check if valid (0 < k < max)
		if k.Cmp(big.NewInt(0)) > 0 && k.Cmp(max) < 0 {
			return k, nil
		}
	}
}

// DER encodes the signature in DER format
func (sig *Signature) DER() []byte {
	// Encode r
	rBytes := sig.R.Bytes()
	if rBytes[0] >= 0x80 {
		rBytes = append([]byte{0x00}, rBytes...)
	}

	// Encode s
	sBytes := sig.S.Bytes()
	if sBytes[0] >= 0x80 {
		sBytes = append([]byte{0x00}, sBytes...)
	}

	// Build DER sequence
	der := []byte{0x30} // SEQUENCE tag
	
	// r INTEGER
	rDER := append([]byte{0x02, byte(len(rBytes))}, rBytes...)
	
	// s INTEGER
	sDER := append([]byte{0x02, byte(len(sBytes))}, sBytes...)
	
	// Combine
	content := append(rDER, sDER...)
	der = append(der, byte(len(content)))
	der = append(der, content...)

	return der
}

// String returns a string representation of the signature
func (sig *Signature) String() string {
	return fmt.Sprintf("Signature(r=%s, s=%s)", sig.R.String(), sig.S.String())
}

// SignMessage signs a message (not a hash) using SHA256
func SignMessage(privateKey *big.Int, message []byte) (*Signature, error) {
	hash := sha256.Sum256(message)
	return Sign(privateKey, hash[:])
}

// VerifyMessage verifies a signature against a message (not a hash) using SHA256
func (sig *Signature) VerifyMessage(publicKey *Point, message []byte) bool {
	hash := sha256.Sum256(message)
	return sig.Verify(publicKey, hash[:])
}
