package models

import (
	"time"

	"github.com/google/uuid"
)

// Account represents a user account in the system
type Account struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	IsAdmin      bool       `json:"is_admin" db:"is_admin"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// AccountInput represents the input for creating or updating an account
type AccountInput struct {
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
	IsAdmin   *bool   `json:"is_admin,omitempty"`
}

// NewAccount creates a new account with default values
func NewAccount(input AccountInput, passwordHash string) *Account {
	now := time.Now()
	isActive := true
	isAdmin := false

	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	if input.IsAdmin != nil {
		isAdmin = *input.IsAdmin
	}

	firstName := ""
	if input.FirstName != nil {
		firstName = *input.FirstName
	}

	lastName := ""
	if input.LastName != nil {
		lastName = *input.LastName
	}

	return &Account{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		IsActive:     isActive,
		IsAdmin:      isAdmin,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
