package handlers

import (
	"github.com/google/uuid"
)

// generateID generates a unique ID using UUID
func generateID() string {
	return uuid.New().String()
}
