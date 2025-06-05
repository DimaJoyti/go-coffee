package utils

import (
	"fmt"

	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID string
// Returns a UUID string, panics if generation fails
func GenerateUUID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Errorf("failed to generate UUID: %w", err))
	}
	return id.String()
}

// GenerateUUIDSafe generates a new UUID string
// Returns empty string and error if UUID generation fails
func GenerateUUIDSafe() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return id.String(), nil
}

// ParseUUID validates and parses a UUID string
// Returns error if the string is not a valid UUID
func ParseUUID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}
	return nil
}

// MustGenerateUUID generates a new UUID string
// Panics if UUID generation fails
// Use this only when you're sure UUID generation won't fail
func MustGenerateUUID() string {
	return GenerateUUID()
}
