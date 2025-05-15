package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/repository"
)

// AccountService handles business logic for accounts
type AccountService struct {
	accountRepo repository.AccountRepository
}

// NewAccountService creates a new account service
func NewAccountService(accountRepo repository.AccountRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
	}
}

// Create creates a new account
func (s *AccountService) Create(ctx context.Context, input models.AccountInput) (*models.Account, error) {
	// Check if username already exists
	existingAccount, err := s.accountRepo.GetByUsername(ctx, input.Username)
	if err == nil && existingAccount != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	existingAccount, err = s.accountRepo.GetByEmail(ctx, input.Email)
	if err == nil && existingAccount != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash the password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the account
	account := models.NewAccount(input, string(passwordHash))

	// Save the account
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// GetByID retrieves an account by ID
func (s *AccountService) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return account, nil
}

// GetByUsername retrieves an account by username
func (s *AccountService) GetByUsername(ctx context.Context, username string) (*models.Account, error) {
	account, err := s.accountRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return account, nil
}

// GetByEmail retrieves an account by email
func (s *AccountService) GetByEmail(ctx context.Context, email string) (*models.Account, error) {
	account, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return account, nil
}

// List retrieves all accounts with optional pagination
func (s *AccountService) List(ctx context.Context, offset, limit int) ([]*models.Account, error) {
	accounts, err := s.accountRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	return accounts, nil
}

// Update updates an existing account
func (s *AccountService) Update(ctx context.Context, id uuid.UUID, input models.AccountInput) (*models.Account, error) {
	// Get the existing account
	account, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Update the account fields
	if input.Username != "" && input.Username != account.Username {
		// Check if username already exists
		existingAccount, err := s.accountRepo.GetByUsername(ctx, input.Username)
		if err == nil && existingAccount != nil && existingAccount.ID != id {
			return nil, fmt.Errorf("username already exists")
		}
		account.Username = input.Username
	}

	if input.Email != "" && input.Email != account.Email {
		// Check if email already exists
		existingAccount, err := s.accountRepo.GetByEmail(ctx, input.Email)
		if err == nil && existingAccount != nil && existingAccount.ID != id {
			return nil, fmt.Errorf("email already exists")
		}
		account.Email = input.Email
	}

	if input.Password != "" {
		// Hash the password
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		account.PasswordHash = string(passwordHash)
	}

	if input.FirstName != nil {
		account.FirstName = *input.FirstName
	}

	if input.LastName != nil {
		account.LastName = *input.LastName
	}

	if input.IsActive != nil {
		account.IsActive = *input.IsActive
	}

	if input.IsAdmin != nil {
		account.IsAdmin = *input.IsAdmin
	}

	// Save the updated account
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

// Delete deletes an account by ID
func (s *AccountService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.accountRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	return nil
}

// Count returns the total number of accounts
func (s *AccountService) Count(ctx context.Context) (int, error) {
	count, err := s.accountRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count accounts: %w", err)
	}
	return count, nil
}

// Authenticate authenticates an account by username and password
func (s *AccountService) Authenticate(ctx context.Context, username, password string) (*models.Account, error) {
	// Get the account by username
	account, err := s.accountRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Check if the account is active
	if !account.IsActive {
		return nil, fmt.Errorf("account is not active")
	}

	// Compare the password
	err = bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	return account, nil
}
