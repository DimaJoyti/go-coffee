package smartcontract

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
	"github.com/yourusername/web3-wallet-backend/pkg/models"
)

// Repository provides access to contract storage
type Repository interface {
	CreateContract(ctx context.Context, contract *models.Contract) error
	GetContract(ctx context.Context, id string) (*models.Contract, error)
	GetContractByAddress(ctx context.Context, address, chain string) (*models.Contract, error)
	ListContracts(ctx context.Context, userID, chain, contractType string, limit, offset int) ([]models.Contract, int, error)
	UpdateContract(ctx context.Context, contract *models.Contract) error
	DeleteContract(ctx context.Context, id string) error
	CreateContractEvent(ctx context.Context, event *models.ContractEvent) error
	GetContractEvents(ctx context.Context, contractID, eventName string, fromBlock, toBlock uint64, limit, offset int) ([]models.ContractEvent, int, error)
}

// PostgresRepository is a PostgreSQL implementation of Repository
type PostgresRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB, logger *logger.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger.Named("smartcontract-repository"),
	}
}

// CreateContract creates a new contract in the database
func (r *PostgresRepository) CreateContract(ctx context.Context, contract *models.Contract) error {
	query := `
		INSERT INTO contracts (id, user_id, name, address, chain, abi, bytecode, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		contract.ID,
		contract.UserID,
		contract.Name,
		contract.Address,
		contract.Chain,
		contract.ABI,
		contract.Bytecode,
		contract.CreatedAt,
		contract.UpdatedAt,
	)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to create contract: %v", err))
		return fmt.Errorf("failed to create contract: %w", err)
	}

	return nil
}

// GetContract retrieves a contract by ID
func (r *PostgresRepository) GetContract(ctx context.Context, id string) (*models.Contract, error) {
	query := `
		SELECT id, user_id, name, address, chain, abi, bytecode, created_at, updated_at
		FROM contracts
		WHERE id = $1
	`

	var contract models.Contract
	err := r.db.GetContext(ctx, &contract, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.Error(fmt.Sprintf("Failed to get contract: %v", err))
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}

	return &contract, nil
}

// GetContractByAddress retrieves a contract by address
func (r *PostgresRepository) GetContractByAddress(ctx context.Context, address, chain string) (*models.Contract, error) {
	query := `
		SELECT id, user_id, name, address, chain, abi, bytecode, created_at, updated_at
		FROM contracts
		WHERE address = $1 AND chain = $2
	`

	var contract models.Contract
	err := r.db.GetContext(ctx, &contract, query, address, chain)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.Error(fmt.Sprintf("Failed to get contract by address: %v", err))
		return nil, fmt.Errorf("failed to get contract by address: %w", err)
	}

	return &contract, nil
}

// ListContracts lists all contracts for a user
func (r *PostgresRepository) ListContracts(ctx context.Context, userID, chain, contractType string, limit, offset int) ([]models.Contract, int, error) {
	// Build query based on filters
	query := `
		SELECT id, user_id, name, address, chain, abi, bytecode, created_at, updated_at
		FROM contracts
		WHERE user_id = $1
	`
	countQuery := `
		SELECT COUNT(*)
		FROM contracts
		WHERE user_id = $1
	`

	args := []interface{}{userID}
	argIndex := 2

	if chain != "" {
		query += fmt.Sprintf(" AND chain = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND chain = $%d", argIndex)
		args = append(args, chain)
		argIndex++
	}

	if contractType != "" {
		// For contract type, we need to check the ABI for specific interfaces
		// This is a simplified approach; in a real system, you might want to store the contract type explicitly
		switch contractType {
		case string(models.ContractTypeERC20):
			query += " AND abi LIKE '%function transfer(address,uint256)%' AND abi LIKE '%function balanceOf(address)%'"
			countQuery += " AND abi LIKE '%function transfer(address,uint256)%' AND abi LIKE '%function balanceOf(address)%'"
		case string(models.ContractTypeERC721):
			query += " AND abi LIKE '%function safeTransferFrom(address,address,uint256)%' AND abi LIKE '%function ownerOf(uint256)%'"
			countQuery += " AND abi LIKE '%function safeTransferFrom(address,address,uint256)%' AND abi LIKE '%function ownerOf(uint256)%'"
		case string(models.ContractTypeERC1155):
			query += " AND abi LIKE '%function safeTransferFrom(address,address,uint256,uint256,bytes)%' AND abi LIKE '%function balanceOf(address,uint256)%'"
			countQuery += " AND abi LIKE '%function safeTransferFrom(address,address,uint256,uint256,bytes)%' AND abi LIKE '%function balanceOf(address,uint256)%'"
		}
	}

	// Add order by and pagination
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	// Get contracts
	var contracts []models.Contract
	err := r.db.SelectContext(ctx, &contracts, query, args...)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to list contracts: %v", err))
		return nil, 0, fmt.Errorf("failed to list contracts: %w", err)
	}

	// Get total count
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, args[:argIndex-1]...)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to get total contract count: %v", err))
		return nil, 0, fmt.Errorf("failed to get total contract count: %w", err)
	}

	return contracts, total, nil
}

// UpdateContract updates a contract in the database
func (r *PostgresRepository) UpdateContract(ctx context.Context, contract *models.Contract) error {
	query := `
		UPDATE contracts
		SET name = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		contract.Name,
		contract.UpdatedAt,
		contract.ID,
	)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to update contract: %v", err))
		return fmt.Errorf("failed to update contract: %w", err)
	}

	return nil
}

// DeleteContract deletes a contract from the database
func (r *PostgresRepository) DeleteContract(ctx context.Context, id string) error {
	query := `
		DELETE FROM contracts
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to delete contract: %v", err))
		return fmt.Errorf("failed to delete contract: %w", err)
	}

	return nil
}

// CreateContractEvent creates a new contract event in the database
func (r *PostgresRepository) CreateContractEvent(ctx context.Context, event *models.ContractEvent) error {
	query := `
		INSERT INTO contract_events (id, contract_id, transaction_id, event, block_number, log_index, data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		event.ContractID,
		event.TransactionID,
		event.Event,
		event.BlockNumber,
		event.LogIndex,
		event.Data,
		event.CreatedAt,
	)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to create contract event: %v", err))
		return fmt.Errorf("failed to create contract event: %w", err)
	}

	return nil
}

// GetContractEvents retrieves events for a contract
func (r *PostgresRepository) GetContractEvents(ctx context.Context, contractID, eventName string, fromBlock, toBlock uint64, limit, offset int) ([]models.ContractEvent, int, error) {
	// Build query based on filters
	query := `
		SELECT id, contract_id, transaction_id, event, block_number, log_index, data, created_at
		FROM contract_events
		WHERE contract_id = $1
	`
	countQuery := `
		SELECT COUNT(*)
		FROM contract_events
		WHERE contract_id = $1
	`

	args := []interface{}{contractID}
	argIndex := 2

	if eventName != "" {
		query += fmt.Sprintf(" AND event = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND event = $%d", argIndex)
		args = append(args, eventName)
		argIndex++
	}

	if fromBlock > 0 {
		query += fmt.Sprintf(" AND block_number >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND block_number >= $%d", argIndex)
		args = append(args, fromBlock)
		argIndex++
	}

	if toBlock > 0 {
		query += fmt.Sprintf(" AND block_number <= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND block_number <= $%d", argIndex)
		args = append(args, toBlock)
		argIndex++
	}

	// Add order by and pagination
	query += " ORDER BY block_number DESC, log_index DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	// Get events
	var events []models.ContractEvent
	err := r.db.SelectContext(ctx, &events, query, args...)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to list contract events: %v", err))
		return nil, 0, fmt.Errorf("failed to list contract events: %w", err)
	}

	// Get total count
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, args[:argIndex-1]...)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to get total contract event count: %v", err))
		return nil, 0, fmt.Errorf("failed to get total contract event count: %w", err)
	}

	return events, total, nil
}
