package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Database represents a database connection wrapper
type Database struct {
	db     *sqlx.DB
	config *config.DatabaseConfig
	logger *logger.Logger
}

// DatabaseInterface defines the database interface
type DatabaseInterface interface {
	// Connection management
	Ping(ctx context.Context) error
	Close() error
	Stats() sql.DBStats

	// Query operations
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Prepared statements
	Prepare(ctx context.Context, query string) (*sql.Stmt, error)

	// Transactions
	Begin(ctx context.Context) (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Extended operations (using sqlx)
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error)

	// Utility methods
	GetDB() *sqlx.DB
	GetConfig() *config.DatabaseConfig
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig, logger *logger.Logger) (DatabaseInterface, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	// Create connection
	db, err := sqlx.Connect("postgres", cfg.GetDatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to database successfully")

	return &Database{
		db:     db,
		config: cfg,
		logger: logger,
	}, nil
}

// Ping checks database connectivity
func (d *Database) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Stats returns database statistics
func (d *Database) Stats() sql.DBStats {
	return d.db.Stats()
}

// Query executes a query that returns rows
func (d *Database) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.db.QueryContext(ctx, query, args...)
	d.logQuery(query, args, time.Since(start), err)
	return rows, err
}

// QueryRow executes a query that returns at most one row
func (d *Database) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := d.db.QueryRowContext(ctx, query, args...)
	d.logQuery(query, args, time.Since(start), nil)
	return row
}

// Exec executes a query without returning any rows
func (d *Database) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := d.db.ExecContext(ctx, query, args...)
	d.logQuery(query, args, time.Since(start), err)
	return result, err
}

// Prepare creates a prepared statement
func (d *Database) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return d.db.PrepareContext(ctx, query)
}

// Begin starts a transaction
func (d *Database) Begin(ctx context.Context) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, nil)
}

// BeginTx starts a transaction with options
func (d *Database) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}

// Get using sqlx for single row queries
func (d *Database) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := d.db.GetContext(ctx, dest, query, args...)
	d.logQuery(query, args, time.Since(start), err)
	return err
}

// Select using sqlx for multi-row queries
func (d *Database) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := d.db.SelectContext(ctx, dest, query, args...)
	d.logQuery(query, args, time.Since(start), err)
	return err
}

// NamedQuery using sqlx for named parameter queries
func (d *Database) NamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := d.db.NamedQueryContext(ctx, query, arg)
	d.logQuery(query, []interface{}{arg}, time.Since(start), err)
	return rows, err
}

// NamedExec using sqlx for named parameter execution
func (d *Database) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := d.db.NamedExecContext(ctx, query, arg)
	d.logQuery(query, []interface{}{arg}, time.Since(start), err)
	return result, err
}

// GetDB returns the underlying sqlx.DB
func (d *Database) GetDB() *sqlx.DB {
	return d.db
}

// GetConfig returns the database configuration
func (d *Database) GetConfig() *config.DatabaseConfig {
	return d.config
}

// logQuery logs database queries if enabled
func (d *Database) logQuery(query string, args []interface{}, duration time.Duration, err error) {
	if !d.config.LogQueries {
		return
	}

	fields := []logger.Field{
		logger.String("query", query),
		logger.Duration("duration", duration),
	}

	if len(args) > 0 {
		fields = append(fields, logger.Any("args", args))
	}

	if err != nil {
		fields = append(fields, logger.Error(err))
		d.logger.ErrorWithFields("Database query failed", fields...)
	} else if duration > d.config.SlowQueryThreshold {
		d.logger.WarnWithFields("Slow database query", fields...)
	} else {
		d.logger.With(fields...).Debug("Database query executed")
	}
}

// Transaction represents a database transaction wrapper
type Transaction struct {
	tx     *sql.Tx
	logger *logger.Logger
}

// TransactionInterface defines the transaction interface
type TransactionInterface interface {
	// Query operations
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Prepared statements
	Prepare(ctx context.Context, query string) (*sql.Stmt, error)

	// Transaction control
	Commit() error
	Rollback() error
}

// NewTransaction creates a new transaction wrapper
func NewTransaction(tx *sql.Tx, logger *logger.Logger) TransactionInterface {
	return &Transaction{
		tx:     tx,
		logger: logger,
	}
}

// Query executes a query within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := t.tx.QueryContext(ctx, query, args...)
	t.logQuery(query, args, time.Since(start), err)
	return rows, err
}

// QueryRow executes a query that returns at most one row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := t.tx.QueryRowContext(ctx, query, args...)
	t.logQuery(query, args, time.Since(start), nil)
	return row
}

// Exec executes a query without returning any rows within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := t.tx.ExecContext(ctx, query, args...)
	t.logQuery(query, args, time.Since(start), err)
	return result, err
}

// Prepare creates a prepared statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return t.tx.PrepareContext(ctx, query)
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	err := t.tx.Commit()
	if err != nil {
		t.logger.WithError(err).Error("Failed to commit transaction")
	} else {
		t.logger.Debug("Transaction committed successfully")
	}
	return err
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	err := t.tx.Rollback()
	if err != nil {
		t.logger.WithError(err).Error("Failed to rollback transaction")
	} else {
		t.logger.Debug("Transaction rolled back successfully")
	}
	return err
}

// logQuery logs transaction queries
func (t *Transaction) logQuery(query string, args []interface{}, duration time.Duration, err error) {
	fields := []logger.Field{
		logger.String("query", query),
		logger.Duration("duration", duration),
		logger.String("context", "transaction"),
	}

	if len(args) > 0 {
		fields = append(fields, logger.Any("args", args))
	}

	if err != nil {
		fields = append(fields, logger.Error(err))
		t.logger.ErrorWithFields("Transaction query failed", fields...)
	} else {
		t.logger.With(fields...).Debug("Transaction query executed")
	}
}

// DatabaseManager manages multiple database connections
type DatabaseManager struct {
	databases map[string]DatabaseInterface
	logger    *logger.Logger
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(logger *logger.Logger) *DatabaseManager {
	return &DatabaseManager{
		databases: make(map[string]DatabaseInterface),
		logger:    logger,
	}
}

// AddDatabase adds a database connection
func (dm *DatabaseManager) AddDatabase(name string, db DatabaseInterface) {
	dm.databases[name] = db
}

// GetDatabase returns a database connection by name
func (dm *DatabaseManager) GetDatabase(name string) (DatabaseInterface, bool) {
	db, exists := dm.databases[name]
	return db, exists
}

// GetDefaultDatabase returns the default database connection
func (dm *DatabaseManager) GetDefaultDatabase() DatabaseInterface {
	if db, exists := dm.databases["default"]; exists {
		return db
	}
	return nil
}

// CloseAll closes all database connections
func (dm *DatabaseManager) CloseAll() error {
	var lastErr error
	for name, db := range dm.databases {
		if err := db.Close(); err != nil {
			dm.logger.WithError(err).WithField("database", name).Error("Failed to close database")
			lastErr = err
		}
	}
	return lastErr
}

// HealthCheck checks the health of all databases
func (dm *DatabaseManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)
	for name, db := range dm.databases {
		results[name] = db.Ping(ctx)
	}
	return results
}

// GetStats returns statistics for all databases
func (dm *DatabaseManager) GetStats() map[string]sql.DBStats {
	results := make(map[string]sql.DBStats)
	for name, db := range dm.databases {
		results[name] = db.Stats()
	}
	return results
}
