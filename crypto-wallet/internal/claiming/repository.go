package claiming

import (
	"context"
)

// Repository represents a claiming repository
type Repository interface {
	// GetClaim gets a claim by ID
	GetClaim(ctx context.Context, id string) (*Claim, error)
	
	// GetClaimByOrderID gets a claim by order ID
	GetClaimByOrderID(ctx context.Context, orderID string) (*Claim, error)
	
	// CreateClaim creates a new claim
	CreateClaim(ctx context.Context, claim *Claim) error
	
	// UpdateClaim updates a claim
	UpdateClaim(ctx context.Context, id string, claim *Claim) error
	
	// ListClaims lists claims
	ListClaims(ctx context.Context, userID, orderID, status string, page, pageSize int) ([]*Claim, int, error)
	
	// ClaimOrder claims an order
	ClaimOrder(ctx context.Context, orderID, userID string) (*Claim, error)
}
