package models

import (
	"time"

	"github.com/google/uuid"
)

// Vendor represents a vendor in the system
type Vendor struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	ContactEmail string    `json:"contact_email" db:"contact_email"`
	ContactPhone string    `json:"contact_phone" db:"contact_phone"`
	Address      string    `json:"address" db:"address"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// VendorInput represents the input for creating or updating a vendor
type VendorInput struct {
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ContactEmail *string `json:"contact_email,omitempty"`
	ContactPhone *string `json:"contact_phone,omitempty"`
	Address      *string `json:"address,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// NewVendor creates a new vendor with default values
func NewVendor(input VendorInput) *Vendor {
	now := time.Now()
	isActive := true

	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	description := ""
	if input.Description != nil {
		description = *input.Description
	}

	contactEmail := ""
	if input.ContactEmail != nil {
		contactEmail = *input.ContactEmail
	}

	contactPhone := ""
	if input.ContactPhone != nil {
		contactPhone = *input.ContactPhone
	}

	address := ""
	if input.Address != nil {
		address = *input.Address
	}

	return &Vendor{
		ID:           uuid.New(),
		Name:         input.Name,
		Description:  description,
		ContactEmail: contactEmail,
		ContactPhone: contactPhone,
		Address:      address,
		IsActive:     isActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
