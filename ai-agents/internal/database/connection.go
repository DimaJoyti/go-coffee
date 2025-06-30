package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-coffee-ai-agents/internal/config"
	"go-coffee-ai-agents/internal/observability"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB represents a database connection with observability
type DB struct {
	*sql.DB
	config  config.DatabaseConfig
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
}

// ConnectionManager manages database connections and lifecycle
type ConnectionManager struct {
	db      *DB
	config  config.DatabaseConfig
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
}

// NewConnectionManager creates a new database connection manager
func NewConnectionManager(
	config config.DatabaseConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *ConnectionManager {
	return &ConnectionManager{
		config:  config,
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// Connect establishes a database connection
func (cm *ConnectionManager) Connect(ctx context.Context) (*DB, error) {
	ctx, span := cm.tracing.StartDatabaseSpan(ctx, "CONNECT", "database")
	defer span.End()

	start := time.Now()
	
	// Build connection string
	dsn := cm.config.GetDSN()
	if dsn == "" {
		err := fmt.Errorf("invalid database configuration: empty DSN")
		cm.tracing.RecordError(span, err, "Failed to build DSN")
		return nil, err
	}

	cm.logger.InfoContext(ctx, "Connecting to database",
		"driver", cm.config.Driver,
		"host", cm.config.Host,
		"port", cm.config.Port,
		"database", cm.config.Database)

	// Open database connection
	sqlDB, err := sql.Open(cm.config.Driver, dsn)
	if err != nil {
		cm.tracing.RecordError(span, err, "Failed to open database connection")
		cm.logger.ErrorContext(ctx, "Failed to open database connection", err,
			"driver", cm.config.Driver,
			"host", cm.config.Host)
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	cm.configureConnectionPool(sqlDB)

	// Test connection
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		cm.tracing.RecordError(span, err, "Failed to ping database")
		cm.logger.ErrorContext(ctx, "Failed to ping database", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	duration := time.Since(start)
	
	// Record metrics
	if cm.metrics != nil {
		histograms := cm.metrics.GetHistograms()
		if histograms != nil {
			histograms.DatabaseQueryDuration.Record(ctx, duration.Seconds())
		}
	}

	// Create DB wrapper
	db := &DB{
		DB:      sqlDB,
		config:  cm.config,
		logger:  cm.logger,
		metrics: cm.metrics,
		tracing: cm.tracing,
	}

	cm.db = db
	cm.tracing.RecordSuccess(span, "Database connection established")
	cm.logger.InfoContext(ctx, "Database connection established",
		"duration_ms", duration.Milliseconds(),
		"max_open_conns", cm.config.MaxOpenConns,
		"max_idle_conns", cm.config.MaxIdleConns)

	return db, nil
}

// configureConnectionPool configures the database connection pool
func (cm *ConnectionManager) configureConnectionPool(db *sql.DB) {
	// Set maximum number of open connections
	db.SetMaxOpenConns(cm.config.MaxOpenConns)
	
	// Set maximum number of idle connections
	db.SetMaxIdleConns(cm.config.MaxIdleConns)
	
	// Set maximum lifetime of connections
	db.SetConnMaxLifetime(cm.config.ConnMaxLifetime)
	
	// Set maximum idle time of connections
	if cm.config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(cm.config.ConnMaxIdleTime)
	}

	cm.logger.Debug("Database connection pool configured",
		"max_open_conns", cm.config.MaxOpenConns,
		"max_idle_conns", cm.config.MaxIdleConns,
		"conn_max_lifetime", cm.config.ConnMaxLifetime,
		"conn_max_idle_time", cm.config.ConnMaxIdleTime)
}

// GetDB returns the current database connection
func (cm *ConnectionManager) GetDB() *DB {
	return cm.db
}

// Close closes the database connection
func (cm *ConnectionManager) Close(ctx context.Context) error {
	if cm.db == nil {
		return nil
	}

	ctx, span := cm.tracing.StartDatabaseSpan(ctx, "CLOSE", "database")
	defer span.End()

	cm.logger.InfoContext(ctx, "Closing database connection")
	
	err := cm.db.Close()
	if err != nil {
		cm.tracing.RecordError(span, err, "Failed to close database connection")
		cm.logger.ErrorContext(ctx, "Failed to close database connection", err)
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	cm.tracing.RecordSuccess(span, "Database connection closed")
	cm.logger.InfoContext(ctx, "Database connection closed successfully")
	cm.db = nil
	
	return nil
}

// HealthCheck performs a database health check
func (cm *ConnectionManager) HealthCheck(ctx context.Context) error {
	if cm.db == nil {
		return fmt.Errorf("database connection not established")
	}

	ctx, span := cm.tracing.StartDatabaseSpan(ctx, "PING", "database")
	defer span.End()

	start := time.Now()
	err := cm.db.PingContext(ctx)
	duration := time.Since(start)

	// Record metrics
	if cm.metrics != nil {
		histograms := cm.metrics.GetHistograms()
		if histograms != nil {
			histograms.DatabaseQueryDuration.Record(ctx, duration.Seconds())
		}
	}

	if err != nil {
		cm.tracing.RecordError(span, err, "Database health check failed")
		cm.logger.ErrorContext(ctx, "Database health check failed", err,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("database health check failed: %w", err)
	}

	cm.tracing.RecordSuccess(span, "Database health check passed")
	cm.logger.DebugContext(ctx, "Database health check passed",
		"duration_ms", duration.Milliseconds())
	
	return nil
}

// GetStats returns database connection statistics
func (cm *ConnectionManager) GetStats() sql.DBStats {
	if cm.db == nil {
		return sql.DBStats{}
	}
	return cm.db.Stats()
}

// ExecContext executes a query with observability
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, span := db.tracing.StartDatabaseSpan(ctx, "EXEC", "query")
	defer span.End()

	start := time.Now()
	
	// Add query information to span
	db.tracing.SetAttributes(span,
		observability.Attribute("db.statement", query),
		observability.Attribute("db.operation", "exec"))

	if db.config.EnableLogging {
		db.logger.DebugContext(ctx, "Executing database query",
			"query", query,
			"args_count", len(args))
	}

	result, err := db.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	// Record metrics
	if db.metrics != nil {
		counters := db.metrics.GetCounters()
		histograms := db.metrics.GetHistograms()
		if counters != nil && histograms != nil {
			counters.DatabaseOperations.Add(ctx, 1)
			histograms.DatabaseQueryDuration.Record(ctx, duration.Seconds())
		}
	}

	if err != nil {
		db.tracing.RecordError(span, err, "Database query execution failed")
		db.logger.ErrorContext(ctx, "Database query execution failed", err,
			"query", query,
			"duration_ms", duration.Milliseconds())
		return nil, err
	}

	db.tracing.RecordSuccess(span, "Database query executed successfully")
	
	if db.config.EnableLogging {
		db.logger.DebugContext(ctx, "Database query executed successfully",
			"duration_ms", duration.Milliseconds())
	}

	return result, nil
}

// QueryContext executes a query that returns rows with observability
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, span := db.tracing.StartDatabaseSpan(ctx, "QUERY", "query")
	defer span.End()

	start := time.Now()
	
	// Add query information to span
	db.tracing.SetAttributes(span,
		observability.Attribute("db.statement", query),
		observability.Attribute("db.operation", "query"))

	if db.config.EnableLogging {
		db.logger.DebugContext(ctx, "Executing database query",
			"query", query,
			"args_count", len(args))
	}

	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	// Record metrics
	if db.metrics != nil {
		counters := db.metrics.GetCounters()
		histograms := db.metrics.GetHistograms()
		if counters != nil && histograms != nil {
			counters.DatabaseOperations.Add(ctx, 1)
			histograms.DatabaseQueryDuration.Record(ctx, duration.Seconds())
		}
	}

	if err != nil {
		db.tracing.RecordError(span, err, "Database query failed")
		db.logger.ErrorContext(ctx, "Database query failed", err,
			"query", query,
			"duration_ms", duration.Milliseconds())
		return nil, err
	}

	db.tracing.RecordSuccess(span, "Database query executed successfully")
	
	if db.config.EnableLogging {
		db.logger.DebugContext(ctx, "Database query executed successfully",
			"duration_ms", duration.Milliseconds())
	}

	return rows, nil
}

// QueryRowContext executes a query that returns a single row with observability
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, span := db.tracing.StartDatabaseSpan(ctx, "QUERY_ROW", "query")
	defer span.End()

	start := time.Now()
	
	// Add query information to span
	db.tracing.SetAttributes(span,
		observability.Attribute("db.statement", query),
		observability.Attribute("db.operation", "query_row"))

	if db.config.EnableLogging {
		db.logger.DebugContext(ctx, "Executing database query row",
			"query", query,
			"args_count", len(args))
	}

	row := db.DB.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	// Record metrics
	if db.metrics != nil {
		counters := db.metrics.GetCounters()
		histograms := db.metrics.GetHistograms()
		if counters != nil && histograms != nil {
			counters.DatabaseOperations.Add(ctx, 1)
			histograms.DatabaseQueryDuration.Record(ctx, duration.Seconds())
		}
	}

	db.tracing.RecordSuccess(span, "Database query row executed")
	
	if db.config.EnableLogging {
		db.logger.DebugContext(ctx, "Database query row executed",
			"duration_ms", duration.Milliseconds())
	}

	return row
}

// BeginTx starts a transaction with observability
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	ctx, span := db.tracing.StartDatabaseSpan(ctx, "BEGIN", "transaction")
	defer span.End()

	start := time.Now()
	
	db.logger.DebugContext(ctx, "Starting database transaction")

	sqlTx, err := db.DB.BeginTx(ctx, opts)
	duration := time.Since(start)

	if err != nil {
		db.tracing.RecordError(span, err, "Failed to start transaction")
		db.logger.ErrorContext(ctx, "Failed to start transaction", err,
			"duration_ms", duration.Milliseconds())
		return nil, err
	}

	db.tracing.RecordSuccess(span, "Transaction started successfully")
	db.logger.DebugContext(ctx, "Transaction started successfully",
		"duration_ms", duration.Milliseconds())

	return &Tx{
		Tx:      sqlTx,
		db:      db,
		logger:  db.logger,
		metrics: db.metrics,
		tracing: db.tracing,
	}, nil
}

// Tx represents a database transaction with observability
type Tx struct {
	*sql.Tx
	db      *DB
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
}

// Commit commits the transaction with observability
func (tx *Tx) Commit(ctx context.Context) error {
	ctx, span := tx.tracing.StartDatabaseSpan(ctx, "COMMIT", "transaction")
	defer span.End()

	start := time.Now()
	
	tx.logger.DebugContext(ctx, "Committing database transaction")

	err := tx.Tx.Commit()
	duration := time.Since(start)

	if err != nil {
		tx.tracing.RecordError(span, err, "Failed to commit transaction")
		tx.logger.ErrorContext(ctx, "Failed to commit transaction", err,
			"duration_ms", duration.Milliseconds())
		return err
	}

	tx.tracing.RecordSuccess(span, "Transaction committed successfully")
	tx.logger.DebugContext(ctx, "Transaction committed successfully",
		"duration_ms", duration.Milliseconds())

	return nil
}

// Rollback rolls back the transaction with observability
func (tx *Tx) Rollback(ctx context.Context) error {
	ctx, span := tx.tracing.StartDatabaseSpan(ctx, "ROLLBACK", "transaction")
	defer span.End()

	start := time.Now()
	
	tx.logger.DebugContext(ctx, "Rolling back database transaction")

	err := tx.Tx.Rollback()
	duration := time.Since(start)

	if err != nil {
		tx.tracing.RecordError(span, err, "Failed to rollback transaction")
		tx.logger.ErrorContext(ctx, "Failed to rollback transaction", err,
			"duration_ms", duration.Milliseconds())
		return err
	}

	tx.tracing.RecordSuccess(span, "Transaction rolled back successfully")
	tx.logger.DebugContext(ctx, "Transaction rolled back successfully",
		"duration_ms", duration.Milliseconds())

	return nil
}
