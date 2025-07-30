package database

import (
	"context"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/hft-bot/pkg/config"
)

// PostgresDB provides a mock PostgreSQL implementation for development
type PostgresDB struct {
	config *config.DatabaseConfig
	logger Logger
	tables map[string][]map[string]interface{} // Mock table storage
	mu     sync.RWMutex                        // Mutex for thread safety
}

// NewPostgresDB creates a new mock PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig, log Logger) (*PostgresDB, error) {
	db := &PostgresDB{
		config: cfg,
		logger: log,
		tables: make(map[string][]map[string]interface{}),
	}

	log.Info("Connected to Mock PostgreSQL database",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database,
	)

	return db, nil
}

// Close closes the database connection (mock - does nothing)
func (db *PostgresDB) Close() error {
	db.logger.Info("Mock PostgreSQL connection closed")
	return nil
}

// HealthCheck performs a health check on the database
func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	db.logger.Debug("Mock PostgreSQL health check - always healthy")
	return nil
}

// ExecContext executes a query without returning rows
func (db *PostgresDB) ExecContext(ctx context.Context, query string, args ...interface{}) error {
	start := time.Now()

	db.logger.Debug("Mock PostgreSQL EXEC operation",
		"query", query,
		"args", args,
		"duration", time.Since(start),
	)

	// Mock implementation - just log the execution
	db.logger.Info("Mock PostgreSQL query executed",
		"query", query,
		"args_count", len(args),
	)

	return nil
}

// QueryContext executes a query that returns rows
func (db *PostgresDB) QueryContext(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	start := time.Now()

	db.logger.Debug("Mock PostgreSQL QUERY operation",
		"query", query,
		"args", args,
		"duration", time.Since(start),
	)

	// Mock implementation - return empty result set
	result := []map[string]interface{}{}

	db.logger.Info("Mock PostgreSQL query executed",
		"query", query,
		"args_count", len(args),
		"rows_returned", len(result),
	)

	return result, nil
}

// QueryRowContext executes a query that returns a single row
func (db *PostgresDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
	start := time.Now()

	db.logger.Debug("Mock PostgreSQL QUERY ROW operation",
		"query", query,
		"args", args,
		"duration", time.Since(start),
	)

	// Mock implementation - return empty row
	result := map[string]interface{}{}

	db.logger.Info("Mock PostgreSQL query row executed",
		"query", query,
		"args_count", len(args),
	)

	return result, nil
}

// BeginTx starts a transaction
func (db *PostgresDB) BeginTx(ctx context.Context) (*MockTx, error) {
	start := time.Now()

	db.logger.Debug("Mock PostgreSQL BEGIN TRANSACTION",
		"duration", time.Since(start),
	)

	tx := &MockTx{
		db:     db,
		logger: db.logger,
	}

	db.logger.Info("Mock PostgreSQL transaction started")
	return tx, nil
}

// MockTx represents a mock database transaction
type MockTx struct {
	db     *PostgresDB
	logger Logger
}

// Commit commits the transaction
func (tx *MockTx) Commit() error {
	start := time.Now()

	tx.logger.Debug("Mock PostgreSQL COMMIT TRANSACTION",
		"duration", time.Since(start),
	)

	tx.logger.Info("Mock PostgreSQL transaction committed")
	return nil
}

// Rollback rolls back the transaction
func (tx *MockTx) Rollback() error {
	start := time.Now()

	tx.logger.Debug("Mock PostgreSQL ROLLBACK TRANSACTION",
		"duration", time.Since(start),
	)

	tx.logger.Info("Mock PostgreSQL transaction rolled back")
	return nil
}

// ExecContext executes a query within the transaction
func (tx *MockTx) ExecContext(ctx context.Context, query string, args ...interface{}) error {
	start := time.Now()

	tx.logger.Debug("Mock PostgreSQL TX EXEC operation",
		"query", query,
		"args", args,
		"duration", time.Since(start),
	)

	tx.logger.Info("Mock PostgreSQL transaction query executed",
		"query", query,
		"args_count", len(args),
	)

	return nil
}

// QueryContext executes a query within the transaction
func (tx *MockTx) QueryContext(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	start := time.Now()

	tx.logger.Debug("Mock PostgreSQL TX QUERY operation",
		"query", query,
		"args", args,
		"duration", time.Since(start),
	)

	// Mock implementation - return empty result set
	result := []map[string]interface{}{}

	tx.logger.Info("Mock PostgreSQL transaction query executed",
		"query", query,
		"args_count", len(args),
		"rows_returned", len(result),
	)

	return result, nil
}

// GetStats returns mock database statistics
func (db *PostgresDB) GetStats() map[string]interface{} {
	db.mu.RLock()
	tableCount := len(db.tables)
	db.mu.RUnlock()

	return map[string]interface{}{
		"connected":      true,
		"tables_count":   tableCount,
		"max_open_conns": db.config.MaxOpenConns,
		"max_idle_conns": db.config.MaxIdleConns,
		"version":        "mock-postgres-1.0.0",
		"host":           db.config.Host,
		"port":           db.config.Port,
		"database":       db.config.Database,
	}
}

// CreateTable creates a mock table
func (db *PostgresDB) CreateTable(tableName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.tables[tableName]; !exists {
		db.tables[tableName] = []map[string]interface{}{}
		db.logger.Info("Mock table created", "table", tableName)
	}

	return nil
}

// InsertMockData inserts mock data into a table
func (db *PostgresDB) InsertMockData(tableName string, data map[string]interface{}) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.tables[tableName]; !exists {
		db.tables[tableName] = []map[string]interface{}{}
	}

	db.tables[tableName] = append(db.tables[tableName], data)
	db.logger.Info("Mock data inserted", "table", tableName, "rows", len(db.tables[tableName]))

	return nil
}

// GetMockData retrieves mock data from a table
func (db *PostgresDB) GetMockData(tableName string) ([]map[string]interface{}, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, exists := db.tables[tableName]; exists {
		return data, nil
	}

	return []map[string]interface{}{}, nil
}
