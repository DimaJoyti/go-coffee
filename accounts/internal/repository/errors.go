package repository

import (
	"errors"
	"strings"
)

// Common repository errors
var (
	// ErrNotFound is returned when an entity is not found
	ErrNotFound = errors.New("entity not found")

	// ErrAlreadyExists is returned when an entity already exists
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrInvalidInput is returned when input is invalid
	ErrInvalidInput = errors.New("invalid input")
)

// IsNotFound returns true if the error is a not found error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNotFound) || strings.Contains(err.Error(), ErrNotFound.Error())
}
