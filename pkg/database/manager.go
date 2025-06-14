package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"go.uber.org/zap"
)

// Manager provides a unified interface for database operations
type Manager struct {
	optimizedDB *OptimizedDatabase
	config      *config.DatabaseConfig
	logger      *zap.Logger
}

// NewManager creates a new database manager with optimized pooling
func NewManager(cfg *config.DatabaseConfig, logger *zap.Logger) (*Manager, error) {
	// Convert existing config to optimized config
	optimizedConfig := &OptimizedConfig{
		WriteConnectionString: buildConnectionString(cfg),
		ReadConnectionStrings: buildReadConnectionStrings(cfg),
		MaxConnections:        50, // Default values since not in config
		MinConnections:        10,
		MaxConnLifetime:       5 * time.Minute,
		MaxConnIdleTime:       2 * time.Minute,
		HealthCheckPeriod:     30 * time.Second,
		DefaultQueryTimeout:   30 * time.Second,
		SlowQueryThreshold:    1 * time.Second,
		ConnectionTimeout:     10 * time.Second,
		ReadReplicaWeight:     []int{1}, // Default weight
		ReadReplicaFailover:   true,
	}

	optimizedDB, err := NewOptimizedDatabase(optimizedConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create optimized database: %w", err)
	}

	return &Manager{
		optimizedDB: optimizedDB,
		config:      cfg,
		logger:      logger,
	}, nil
}

// buildConnectionString creates a connection string from config
func buildConnectionString(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)
}

// buildReadConnectionStrings creates read replica connection strings
func buildReadConnectionStrings(cfg *config.DatabaseConfig) []string {
	// For now, use the same connection string
	// In production, you would configure separate read replicas
	return []string{buildConnectionString(cfg)}
}

// ExecuteWrite executes a write operation
func (m *Manager) ExecuteWrite(ctx context.Context, query string, args ...interface{}) error {
	return m.optimizedDB.ExecuteWrite(ctx, query, args...)
}

// QueryRead executes a read operation
func (m *Manager) QueryRead(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
	rows, err := m.optimizedDB.QueryRead(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return &QueryResult{rows: rows}, nil
}

// QueryResult wraps database query results
type QueryResult struct {
	rows interface{}
}

// GetMetrics returns database performance metrics
func (m *Manager) GetMetrics() DatabaseMetrics {
	return m.optimizedDB.GetMetrics()
}

// Close closes all database connections
func (m *Manager) Close() {
	m.optimizedDB.Close()
}

// HealthCheck performs a health check on all database connections
func (m *Manager) HealthCheck(ctx context.Context) error {
	// Simple health check - try to execute a basic query
	_, err := m.QueryRead(ctx, "SELECT 1")
	return err
}

// Transaction provides transaction support
func (m *Manager) Transaction(ctx context.Context, fn func(tx Transaction) error) error {
	// Begin transaction on write database
	tx, err := m.optimizedDB.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create transaction wrapper
	txWrapper := &transactionWrapper{tx: tx}

	// Execute function
	if err := fn(txWrapper); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			m.logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Transaction interface for database transactions
type Transaction interface {
	Exec(ctx context.Context, query string, args ...interface{}) error
	Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error)
}

// transactionWrapper implements the Transaction interface
type transactionWrapper struct {
	tx *sql.Tx
}

func (tw *transactionWrapper) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := tw.tx.ExecContext(ctx, query, args...)
	return err
}

func (tw *transactionWrapper) Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
	rows, err := tw.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &QueryResult{rows: rows}, nil
}

// PreparedStatement provides prepared statement support
type PreparedStatement struct {
	query string
	stmt  interface{}
}

// Prepare creates a prepared statement
func (m *Manager) Prepare(ctx context.Context, query string) (*PreparedStatement, error) {
	// Implementation for prepared statements
	return &PreparedStatement{
		query: query,
		stmt:  nil, // Would be actual prepared statement
	}, nil
}

// Execute executes a prepared statement
func (ps *PreparedStatement) Execute(ctx context.Context, args ...interface{}) error {
	// Implementation for executing prepared statements
	return fmt.Errorf("prepared statement execution not implemented")
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	// Implementation for closing prepared statements
	return nil
}

// ConnectionPool provides access to connection pool statistics
type ConnectionPool struct {
	manager *Manager
}

// GetConnectionPool returns connection pool interface
func (m *Manager) GetConnectionPool() *ConnectionPool {
	return &ConnectionPool{manager: m}
}

// Stats returns connection pool statistics
func (cp *ConnectionPool) Stats() PoolStats {
	metrics := cp.manager.GetMetrics()
	return PoolStats{
		TotalConnections:  int(metrics.TotalConnections),
		ActiveConnections: int(metrics.ActiveConnections),
		IdleConnections:   int(metrics.IdleConnections),
		QueryCount:        metrics.QueryCount,
		SlowQueryCount:    metrics.SlowQueryCount,
		ConnectionErrors:  metrics.ConnectionErrors,
		AverageQueryTime:  metrics.AverageQueryTime,
	}
}

// PoolStats contains connection pool statistics
type PoolStats struct {
	TotalConnections  int
	ActiveConnections int
	IdleConnections   int
	QueryCount        int64
	SlowQueryCount    int64
	ConnectionErrors  int64
	AverageQueryTime  time.Duration
}

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder struct {
	manager *Manager
	query   string
	args    []interface{}
}

// NewQueryBuilder creates a new query builder
func (m *Manager) NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		manager: m,
		args:    make([]interface{}, 0),
	}
}

// Select starts a SELECT query
func (qb *QueryBuilder) Select(columns string) *QueryBuilder {
	qb.query = "SELECT " + columns
	return qb
}

// From adds a FROM clause
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.query += " FROM " + table
	return qb
}

// Where adds a WHERE clause
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.query += " WHERE " + condition
	qb.args = append(qb.args, args...)
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(column string) *QueryBuilder {
	qb.query += " ORDER BY " + column
	return qb
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query += fmt.Sprintf(" LIMIT %d", limit)
	return qb
}

// Execute executes the built query
func (qb *QueryBuilder) Execute(ctx context.Context) (*QueryResult, error) {
	return qb.manager.QueryRead(ctx, qb.query, qb.args...)
}

// GetQuery returns the built query and arguments
func (qb *QueryBuilder) GetQuery() (string, []interface{}) {
	return qb.query, qb.args
}

// Batch provides batch operation support
type Batch struct {
	manager *Manager
	queries []BatchQuery
}

// BatchQuery represents a single query in a batch
type BatchQuery struct {
	Query string
	Args  []interface{}
}

// NewBatch creates a new batch
func (m *Manager) NewBatch() *Batch {
	return &Batch{
		manager: m,
		queries: make([]BatchQuery, 0),
	}
}

// Add adds a query to the batch
func (b *Batch) Add(query string, args ...interface{}) {
	b.queries = append(b.queries, BatchQuery{
		Query: query,
		Args:  args,
	})
}

// Execute executes all queries in the batch
func (b *Batch) Execute(ctx context.Context) error {
	return b.manager.Transaction(ctx, func(tx Transaction) error {
		for _, query := range b.queries {
			if err := tx.Exec(ctx, query.Query, query.Args...); err != nil {
				return fmt.Errorf("batch query failed: %w", err)
			}
		}
		return nil
	})
}

// Size returns the number of queries in the batch
func (b *Batch) Size() int {
	return len(b.queries)
}

// Clear clears all queries from the batch
func (b *Batch) Clear() {
	b.queries = b.queries[:0]
}
