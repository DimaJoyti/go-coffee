package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
)

// Account represents a GraphQL account resolver
type Account struct {
	account *models.Account
	resolver *Resolver
}

// ID resolves the ID field of an account
func (r *Account) ID() string {
	return r.account.ID.String()
}

// Username resolves the username field of an account
func (r *Account) Username() string {
	return r.account.Username
}

// Email resolves the email field of an account
func (r *Account) Email() string {
	return r.account.Email
}

// FirstName resolves the firstName field of an account
func (r *Account) FirstName() *string {
	if r.account.FirstName == "" {
		return nil
	}
	return &r.account.FirstName
}

// LastName resolves the lastName field of an account
func (r *Account) LastName() *string {
	if r.account.LastName == "" {
		return nil
	}
	return &r.account.LastName
}

// IsActive resolves the isActive field of an account
func (r *Account) IsActive() bool {
	return r.account.IsActive
}

// IsAdmin resolves the isAdmin field of an account
func (r *Account) IsAdmin() bool {
	return r.account.IsAdmin
}

// CreatedAt resolves the createdAt field of an account
func (r *Account) CreatedAt() string {
	return r.account.CreatedAt.Format(time.RFC3339)
}

// UpdatedAt resolves the updatedAt field of an account
func (r *Account) UpdatedAt() string {
	return r.account.UpdatedAt.Format(time.RFC3339)
}

// Orders resolves the orders field of an account
func (r *Account) Orders(ctx context.Context) ([]*Order, error) {
	orders, err := r.resolver.orderService.ListByAccount(ctx, r.account.ID, 0, 100)
	if err != nil {
		return nil, err
	}

	var orderResolvers []*Order
	for _, order := range orders {
		orderResolvers = append(orderResolvers, &Order{
			order:    order,
			resolver: r.resolver,
		})
	}

	return orderResolvers, nil
}

// Account resolves the account query
func (r *Resolver) Account(ctx context.Context, args struct{ ID string }) (*Account, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	account, err := r.accountService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Account{
		account:  account,
		resolver: r,
	}, nil
}

// Accounts resolves the accounts query
func (r *Resolver) Accounts(ctx context.Context, args struct {
	Offset *int32
	Limit  *int32
}) ([]*Account, error) {
	offset := 0
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	limit := 10
	if args.Limit != nil {
		limit = int(*args.Limit)
	}

	accounts, err := r.accountService.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	var accountResolvers []*Account
	for _, account := range accounts {
		accountResolvers = append(accountResolvers, &Account{
			account:  account,
			resolver: r,
		})
	}

	return accountResolvers, nil
}

// AccountsCount resolves the accountsCount query
func (r *Resolver) AccountsCount(ctx context.Context) (int32, error) {
	count, err := r.accountService.Count(ctx)
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

// CreateAccount resolves the createAccount mutation
func (r *Resolver) CreateAccount(ctx context.Context, args struct {
	Input struct {
		Username  string
		Email     string
		Password  string
		FirstName *string
		LastName  *string
		IsActive  *bool
		IsAdmin   *bool
	}
}) (*Account, error) {
	input := models.AccountInput{
		Username:  args.Input.Username,
		Email:     args.Input.Email,
		Password:  args.Input.Password,
		FirstName: args.Input.FirstName,
		LastName:  args.Input.LastName,
		IsActive:  args.Input.IsActive,
		IsAdmin:   args.Input.IsAdmin,
	}

	account, err := r.accountService.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return &Account{
		account:  account,
		resolver: r,
	}, nil
}

// UpdateAccount resolves the updateAccount mutation
func (r *Resolver) UpdateAccount(ctx context.Context, args struct {
	ID    string
	Input struct {
		Username  *string
		Email     *string
		Password  *string
		FirstName *string
		LastName  *string
		IsActive  *bool
		IsAdmin   *bool
	}
}) (*Account, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	input := models.AccountInput{
		FirstName: args.Input.FirstName,
		LastName:  args.Input.LastName,
		IsActive:  args.Input.IsActive,
		IsAdmin:   args.Input.IsAdmin,
	}

	if args.Input.Username != nil {
		input.Username = *args.Input.Username
	}

	if args.Input.Email != nil {
		input.Email = *args.Input.Email
	}

	if args.Input.Password != nil {
		input.Password = *args.Input.Password
	}

	account, err := r.accountService.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return &Account{
		account:  account,
		resolver: r,
	}, nil
}

// DeleteAccount resolves the deleteAccount mutation
func (r *Resolver) DeleteAccount(ctx context.Context, args struct{ ID string }) (bool, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return false, fmt.Errorf("invalid ID: %w", err)
	}

	err = r.accountService.Delete(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}
