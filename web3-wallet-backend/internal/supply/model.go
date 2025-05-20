package supply

import (
	"time"
)

// Supply represents a supply
type Supply struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Currency  string    `json:"currency" db:"currency"`
	Amount    float64   `json:"amount" db:"amount"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SupplyCreatedEvent represents a supply created event
type SupplyCreatedEvent struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Currency  string    `json:"currency"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// SupplyUpdatedEvent represents a supply updated event
type SupplyUpdatedEvent struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Currency  string    `json:"currency"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SupplyDeletedEvent represents a supply deleted event
type SupplyDeletedEvent struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
}
