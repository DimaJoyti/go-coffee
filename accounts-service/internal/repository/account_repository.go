package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
)

// AccountRepository defines the interface for account repository operations
type AccountRepository interface {
	// Create creates a new account
	Create(ctx context.Context, account *models.Account) error
	
	// GetByID retrieves an account by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	
	// GetByUsername retrieves an account by username
	GetByUsername(ctx context.Context, username string) (*models.Account, error)
	
	// GetByEmail retrieves an account by email
	GetByEmail(ctx context.Context, email string) (*models.Account, error)
	
	// List retrieves all accounts with optional pagination
	List(ctx context.Context, offset, limit int) ([]*models.Account, error)
	
	// Update updates an existing account
	Update(ctx context.Context, account *models.Account) error
	
	// Delete deletes an account by ID
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Count returns the total number of accounts
	Count(ctx context.Context) (int, error)
}
