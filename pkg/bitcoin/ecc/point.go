package ecc

import (
	"crypto/sha256"
	"fmt"
	"math/big"
)

// Point represents a point on an elliptic curve
type Point struct {
	X *big.Int // X coordinate (nil for point at infinity)
	Y *big.Int // Y coordinate (nil for point at infinity)
	A *big.Int // Curve parameter a
	B *big.Int // Curve parameter b
}

// NewPoint creates a new point on the elliptic curve y^2 = x^3 + ax + b
func NewPoint(x, y, a, b *big.Int) (*Point, error) {
	point := &Point{
		X: new(big.Int),
		Y: new(big.Int),
		A: new(big.Int),
		B: new(big.Int),
	}

	// Handle point at infinity
	if x == nil && y == nil {
		point.X = nil
		point.Y = nil
		point.A.Set(a)
		point.B.Set(b)
		return point, nil
	}

	if x == nil || y == nil {
		return nil, fmt.Errorf("x and y must both be nil (point at infinity) or both be non-nil")
	}

	point.X.Set(x)
	point.Y.Set(y)
	point.A.Set(a)
	point.B.Set(b)

	// Verify the point is on the curve: y^2 = x^3 + ax + b
	if !point.IsOnCurve() {
		return nil, fmt.Errorf("point (%s, %s) is not on the curve", x.String(), y.String())
	}

	return point, nil
}

// IsOnCurve checks if the point is on the elliptic curve
func (p *Point) IsOnCurve() bool {
	// Point at infinity is always on the curve
	if p.X == nil && p.Y == nil {
		return true
	}

	// For secp256k1, we need to do modular arithmetic
	curve := secp256k1

	// Calculate y^2 mod p
	ySquared := new(big.Int).Mul(p.Y, p.Y)
	ySquared.Mod(ySquared, curve.P)

	// Calculate x^3 + ax + b mod p
	xCubed := new(big.Int).Mul(p.X, p.X)
	xCubed.Mul(xCubed, p.X)
	xCubed.Mod(xCubed, curve.P)

	ax := new(big.Int).Mul(p.A, p.X)
	ax.Mod(ax, curve.P)

	rightSide := new(big.Int).Add(xCubed, ax)
	rightSide.Add(rightSide, p.B)
	rightSide.Mod(rightSide, curve.P)

	return ySquared.Cmp(rightSide) == 0
}

// IsInfinity checks if the point is the point at infinity
func (p *Point) IsInfinity() bool {
	return p.X == nil && p.Y == nil
}

// Equal checks if two points are equal
func (p *Point) Equal(other *Point) bool {
	// Check if curve parameters are the same
	if p.A.Cmp(other.A) != 0 || p.B.Cmp(other.B) != 0 {
		return false
	}

	// Both points at infinity
	if p.IsInfinity() && other.IsInfinity() {
		return true
	}

	// One point at infinity, other not
	if p.IsInfinity() || other.IsInfinity() {
		return false
	}

	// Compare coordinates
	return p.X.Cmp(other.X) == 0 && p.Y.Cmp(other.Y) == 0
}

// Add performs elliptic curve point addition
func (p *Point) Add(other *Point) (*Point, error) {
	// Check if points are on the same curve
	if p.A.Cmp(other.A) != 0 || p.B.Cmp(other.B) != 0 {
		return nil, fmt.Errorf("points are not on the same curve")
	}

	// Case 1: p is point at infinity
	if p.IsInfinity() {
		return &Point{
			X: new(big.Int).Set(other.X),
			Y: new(big.Int).Set(other.Y),
			A: new(big.Int).Set(other.A),
			B: new(big.Int).Set(other.B),
		}, nil
	}

	// Case 2: other is point at infinity
	if other.IsInfinity() {
		return &Point{
			X: new(big.Int).Set(p.X),
			Y: new(big.Int).Set(p.Y),
			A: new(big.Int).Set(p.A),
			B: new(big.Int).Set(p.B),
		}, nil
	}

	// Case 3: points have same x coordinate
	if p.X.Cmp(other.X) == 0 {
		// Case 3a: points are additive inverses (different y coordinates)
		if p.Y.Cmp(other.Y) != 0 {
			return NewPoint(nil, nil, p.A, p.B)
		}

		// Case 3b: points are the same (point doubling)
		return p.Double()
	}

	// Case 4: general case (different x coordinates)
	// Calculate slope: s = (y2 - y1) / (x2 - x1) mod p
	curve := secp256k1
	numerator := new(big.Int).Sub(other.Y, p.Y)
	numerator.Mod(numerator, curve.P)

	denominator := new(big.Int).Sub(other.X, p.X)
	denominator.Mod(denominator, curve.P)

	slope := new(big.Int).ModInverse(denominator, curve.P)
	slope.Mul(slope, numerator)
	slope.Mod(slope, curve.P)

	// Calculate x3 = s^2 - x1 - x2 mod p
	x3 := new(big.Int).Mul(slope, slope)
	x3.Sub(x3, p.X)
	x3.Sub(x3, other.X)
	x3.Mod(x3, curve.P)

	// Calculate y3 = s(x1 - x3) - y1 mod p
	y3 := new(big.Int).Sub(p.X, x3)
	y3.Mul(slope, y3)
	y3.Sub(y3, p.Y)
	y3.Mod(y3, curve.P)

	return NewPoint(x3, y3, p.A, p.B)
}

// Double performs elliptic curve point doubling
func (p *Point) Double() (*Point, error) {
	// Point at infinity
	if p.IsInfinity() {
		return NewPoint(nil, nil, p.A, p.B)
	}

	// If y = 0, then 2P = O (point at infinity)
	if p.Y.Cmp(big.NewInt(0)) == 0 {
		return NewPoint(nil, nil, p.A, p.B)
	}

	// Calculate slope: s = (3x^2 + a) / (2y) mod p
	curve := secp256k1
	numerator := new(big.Int).Mul(p.X, p.X)
	numerator.Mul(numerator, big.NewInt(3))
	numerator.Add(numerator, p.A)
	numerator.Mod(numerator, curve.P)

	denominator := new(big.Int).Mul(p.Y, big.NewInt(2))
	denominator.Mod(denominator, curve.P)

	slope := new(big.Int).ModInverse(denominator, curve.P)
	slope.Mul(slope, numerator)
	slope.Mod(slope, curve.P)

	// Calculate x3 = s^2 - 2x mod p
	x3 := new(big.Int).Mul(slope, slope)
	x3.Sub(x3, new(big.Int).Mul(p.X, big.NewInt(2)))
	x3.Mod(x3, curve.P)

	// Calculate y3 = s(x - x3) - y mod p
	y3 := new(big.Int).Sub(p.X, x3)
	y3.Mul(slope, y3)
	y3.Sub(y3, p.Y)
	y3.Mod(y3, curve.P)

	return NewPoint(x3, y3, p.A, p.B)
}

// ScalarMult performs scalar multiplication k*P
func (p *Point) ScalarMult(k *big.Int) (*Point, error) {
	// Handle special cases
	if k.Cmp(big.NewInt(0)) == 0 {
		return NewPoint(nil, nil, p.A, p.B)
	}

	if k.Cmp(big.NewInt(1)) == 0 {
		return &Point{
			X: new(big.Int).Set(p.X),
			Y: new(big.Int).Set(p.Y),
			A: new(big.Int).Set(p.A),
			B: new(big.Int).Set(p.B),
		}, nil
	}

	// Use binary method for scalar multiplication
	result, err := NewPoint(nil, nil, p.A, p.B) // Start with point at infinity
	if err != nil {
		return nil, err
	}

	addend := &Point{
		X: new(big.Int).Set(p.X),
		Y: new(big.Int).Set(p.Y),
		A: new(big.Int).Set(p.A),
		B: new(big.Int).Set(p.B),
	}

	// Make a copy of k to avoid modifying the original
	kCopy := new(big.Int).Set(k)

	// Process each bit of k
	for kCopy.Cmp(big.NewInt(0)) > 0 {
		// If current bit is 1, add current addend to result
		if new(big.Int).And(kCopy, big.NewInt(1)).Cmp(big.NewInt(1)) == 0 {
			result, err = result.Add(addend)
			if err != nil {
				return nil, err
			}
		}

		// Double the addend for next bit
		addend, err = addend.Double()
		if err != nil {
			return nil, err
		}

		// Shift k right by 1 bit
		kCopy.Rsh(kCopy, 1)
	}

	return result, nil
}

// String returns a string representation of the point
func (p *Point) String() string {
	if p.IsInfinity() {
		return "Point(infinity)"
	}
	return fmt.Sprintf("Point(%s, %s)", p.X.String(), p.Y.String())
}

// Keccak256 computes the Keccak-256 hash (used by Ethereum)
func Keccak256(data []byte) []byte {
	// For now, use SHA256 as a placeholder
	// In a real implementation, you would use golang.org/x/crypto/sha3
	hash := sha256.Sum256(data)
	return hash[:]
}
