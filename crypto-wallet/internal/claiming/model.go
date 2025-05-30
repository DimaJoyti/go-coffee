package claiming

import (
	"time"
)

// Claim represents a claim
type Claim struct {
	ID          string    `json:"id" db:"id"`
	OrderID     string    `json:"order_id" db:"order_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Status      string    `json:"status" db:"status"`
	ClaimedAt   time.Time `json:"claimed_at" db:"claimed_at"`
	ProcessedAt time.Time `json:"processed_at,omitempty" db:"processed_at"`
}

// OrderClaimedEvent represents an order claimed event
type OrderClaimedEvent struct {
	ClaimID   string    `json:"claim_id"`
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	ClaimedAt time.Time `json:"claimed_at"`
}

// ClaimProcessedEvent represents a claim processed event
type ClaimProcessedEvent struct {
	ClaimID     string    `json:"claim_id"`
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}
