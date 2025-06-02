//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/repository"
)

// These tests require a PostgreSQL database to be running
// They are integration tests and should be run with the -tags=integration flag
// Example: go test -tags=integration ./...

// TestAccountRepository_Integration tests the account repository with a real database
func TestAccountRepository_Integration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Connect to the database
	db, err := sqlx.Connect("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=coffee_accounts_test sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	// Create the database
	database := &Database{db: db}

	// Create the repository
	repo := NewAccountRepository(database)

	// Create a test account
	ctx := context.Background()
	account := &models.Account{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "Test",
		LastName:     "User",
		IsActive:     true,
		IsAdmin:      false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Test Create
	err = repo.Create(ctx, account)
	require.NoError(t, err)

	// Test GetByID
	retrievedAccount, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	assert.Equal(t, account.ID, retrievedAccount.ID)
	assert.Equal(t, account.Username, retrievedAccount.Username)
	assert.Equal(t, account.Email, retrievedAccount.Email)

	// Test GetByUsername
	retrievedAccount, err = repo.GetByUsername(ctx, account.Username)
	require.NoError(t, err)
	assert.Equal(t, account.ID, retrievedAccount.ID)
	assert.Equal(t, account.Username, retrievedAccount.Username)
	assert.Equal(t, account.Email, retrievedAccount.Email)

	// Test GetByEmail
	retrievedAccount, err = repo.GetByEmail(ctx, account.Email)
	require.NoError(t, err)
	assert.Equal(t, account.ID, retrievedAccount.ID)
	assert.Equal(t, account.Username, retrievedAccount.Username)
	assert.Equal(t, account.Email, retrievedAccount.Email)

	// Test Update
	account.FirstName = "Updated"
	account.LastName = "Name"
	err = repo.Update(ctx, account)
	require.NoError(t, err)

	// Verify update
	retrievedAccount, err = repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", retrievedAccount.FirstName)
	assert.Equal(t, "Name", retrievedAccount.LastName)

	// Test List
	accounts, err := repo.List(ctx, 0, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(accounts), 1)

	// Test Count
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)

	// Test Delete
	err = repo.Delete(ctx, account.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, account.ID)
	assert.Error(t, err)
	assert.True(t, repository.IsNotFound(err))
}

// Helper function to check if an error is a not found error
func repository_IsNotFound(err error) bool {
	return err != nil && err.Error() == repository.ErrNotFound.Error()
}
