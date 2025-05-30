package models

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the system
type Product struct {
	ID          uuid.UUID `json:"id" db:"id"`
	VendorID    uuid.UUID `json:"vendor_id" db:"vendor_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	IsAvailable bool      `json:"is_available" db:"is_available"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	
	// Relationships (not stored in the database)
	Vendor      *Vendor   `json:"vendor,omitempty" db:"-"`
}

// ProductInput represents the input for creating or updating a product
type ProductInput struct {
	VendorID    uuid.UUID `json:"vendor_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Price       float64   `json:"price"`
	IsAvailable *bool     `json:"is_available,omitempty"`
}

// NewProduct creates a new product with default values
func NewProduct(input ProductInput) *Product {
	now := time.Now()
	isAvailable := true

	if input.IsAvailable != nil {
		isAvailable = *input.IsAvailable
	}

	description := ""
	if input.Description != nil {
		description = *input.Description
	}

	return &Product{
		ID:          uuid.New(),
		VendorID:    input.VendorID,
		Name:        input.Name,
		Description: description,
		Price:       input.Price,
		IsAvailable: isAvailable,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
