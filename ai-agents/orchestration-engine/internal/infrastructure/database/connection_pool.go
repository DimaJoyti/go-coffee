package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/config"
)

// ConnectionPool manages database connections with optimization
type ConnectionPool struct {
	db     *sql.DB
	config *config.DatabaseConfig
	logger Logger
	stats  *PoolStats
	mutex  sync.RWMutex
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// PoolStats represents connection pool statistics
type PoolStats struct {
	MaxOpenConnections     int           `json:"max_open_connections"`
	OpenConnections        int           `json:"open_connections"`
	InUseConnections       int           `json:"in_use_connections"`
	IdleConnections        int           `json:"idle_connections"`
	WaitCount              int64         `json:"wait_count"`
	WaitDuration           time.Duration `json:"wait_duration"`
	MaxIdleClosed          int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed      int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed      int64         `json:"max_lifetime_closed"`
	TotalQueries           int64         `json:"total_queries"`
	SuccessfulQueries      int64         `json:"successful_queries"`
	FailedQueries          int64         `json:"failed_queries"`
	AverageQueryTime       time.Duration `json:"average_query_time"`
	LastUpdated            time.Time     `json:"last_updated"`
}

// QueryResult represents the result of a database query
type QueryResult struct {
	Rows     *sql.Rows
	Duration time.Duration
	Error    error
}

// TransactionFunc represents a function that executes within a transaction
type TransactionFunc func(tx *sql.Tx) error

// NewConnectionPool creates a new optimized connection pool
func NewConnectionPool(config *config.DatabaseConfig, logger Logger) (*ConnectionPool, error) {
	db, err := sql.Open(config.Driver, config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.ConnectionLifetime)
	db.SetConnMaxIdleTime(config.ConnectionTimeout)

	pool := &ConnectionPool{
		db:     db,
		config: config,
		logger: logger,
		stats: &PoolStats{
			MaxOpenConnections: config.MaxOpenConnections,
			LastUpdated:       time.Now(),
		},
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection pool initialized",
		"max_open_connections", config.MaxOpenConnections,
		"max_idle_connections", config.MaxIdleConnections,
		"connection_lifetime", config.ConnectionLifetime)

	return pool, nil
}

// Query executes a query with performance monitoring
func (cp *ConnectionPool) Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
	start := time.Now()
	
	// Add query timeout if not set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cp.config.QueryTimeout)
		defer cancel()
	}

	rows, err := cp.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	// Update statistics
	cp.updateQueryStats(duration, err == nil)

	if err != nil {
		cp.logger.Error("Database query failed", err,
			"query", query,
			"duration", duration,
			"args_count", len(args))
		return &QueryResult{Duration: duration, Error: err}, err
	}

	cp.logger.Debug("Database query executed",
		"duration", duration,
		"query", query,
		"args_count", len(args))

	return &QueryResult{Rows: rows, Duration: duration}, nil
}

// QueryRow executes a query that returns a single row
func (cp *ConnectionPool) QueryRow(ctx context.Context, query string, args ...interface{}) (*sql.Row, time.Duration, error) {
	start := time.Now()
	
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cp.config.QueryTimeout)
		defer cancel()
	}

	row := cp.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	// Update statistics (assume success for QueryRow as error is checked on Scan)
	cp.updateQueryStats(duration, true)

	cp.logger.Debug("Database query row executed",
		"duration", duration,
		"query", query,
		"args_count", len(args))

	return row, duration, nil
}

// Exec executes a query without returning rows
func (cp *ConnectionPool) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, time.Duration, error) {
	start := time.Now()
	
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cp.config.QueryTimeout)
		defer cancel()
	}

	result, err := cp.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	// Update statistics
	cp.updateQueryStats(duration, err == nil)

	if err != nil {
		cp.logger.Error("Database exec failed", err,
			"query", query,
			"duration", duration,
			"args_count", len(args))
		return nil, duration, err
	}

	cp.logger.Debug("Database exec executed",
		"duration", duration,
		"query", query,
		"args_count", len(args))

	return result, duration, nil
}

// Transaction executes a function within a database transaction
func (cp *ConnectionPool) Transaction(ctx context.Context, fn TransactionFunc) error {
	start := time.Now()
	
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cp.config.QueryTimeout*2) // Double timeout for transactions
		defer cancel()
	}

	tx, err := cp.db.BeginTx(ctx, nil)
	if err != nil {
		cp.logger.Error("Failed to begin transaction", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			cp.logger.Error("Transaction panicked, rolled back", fmt.Errorf("%v", p))
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			cp.logger.Error("Failed to rollback transaction", rollbackErr)
		}
		cp.logger.Error("Transaction failed, rolled back", err, "duration", time.Since(start))
		return err
	}

	if err := tx.Commit(); err != nil {
		cp.logger.Error("Failed to commit transaction", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	cp.logger.Debug("Transaction completed successfully", "duration", time.Since(start))
	return nil
}

// Prepare creates a prepared statement with caching
func (cp *ConnectionPool) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cp.config.ConnectionTimeout)
		defer cancel()
	}

	stmt, err := cp.db.PrepareContext(ctx, query)
	if err != nil {
		cp.logger.Error("Failed to prepare statement", err, "query", query)
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	cp.logger.Debug("Statement prepared", "query", query)
	return stmt, nil
}

// Health checks the health of the connection pool
func (cp *ConnectionPool) Health(ctx context.Context) error {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	if err := cp.db.PingContext(ctx); err != nil {
		cp.logger.Error("Database health check failed", err)
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetStats returns current pool statistics
func (cp *ConnectionPool) GetStats() *PoolStats {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	// Get current database stats
	dbStats := cp.db.Stats()

	// Update pool stats with current values
	cp.stats.OpenConnections = dbStats.OpenConnections
	cp.stats.InUseConnections = dbStats.InUse
	cp.stats.IdleConnections = dbStats.Idle
	cp.stats.WaitCount = dbStats.WaitCount
	cp.stats.WaitDuration = dbStats.WaitDuration
	cp.stats.MaxIdleClosed = dbStats.MaxIdleClosed
	cp.stats.MaxIdleTimeClosed = dbStats.MaxIdleTimeClosed
	cp.stats.MaxLifetimeClosed = dbStats.MaxLifetimeClosed
	cp.stats.LastUpdated = time.Now()

	// Return a copy
	statsCopy := *cp.stats
	return &statsCopy
}

// updateQueryStats updates query performance statistics
func (cp *ConnectionPool) updateQueryStats(duration time.Duration, success bool) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	cp.stats.TotalQueries++
	if success {
		cp.stats.SuccessfulQueries++
	} else {
		cp.stats.FailedQueries++
	}

	// Update average query time (simple moving average)
	if cp.stats.AverageQueryTime == 0 {
		cp.stats.AverageQueryTime = duration
	} else {
		cp.stats.AverageQueryTime = (cp.stats.AverageQueryTime + duration) / 2
	}
}

// Close closes the connection pool
func (cp *ConnectionPool) Close() error {
	cp.logger.Info("Closing database connection pool")
	return cp.db.Close()
}

// GetDB returns the underlying database connection (use with caution)
func (cp *ConnectionPool) GetDB() *sql.DB {
	return cp.db
}

// BatchInsert performs optimized batch insert operations
func (cp *ConnectionPool) BatchInsert(ctx context.Context, table string, columns []string, values [][]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	start := time.Now()
	
	// Build batch insert query
	placeholders := make([]string, len(values))
	args := make([]interface{}, 0, len(values)*len(columns))
	
	for i, row := range values {
		if len(row) != len(columns) {
			return fmt.Errorf("row %d has %d values, expected %d", i, len(row), len(columns))
		}
		
		rowPlaceholders := make([]string, len(columns))
		for j := range columns {
			rowPlaceholders[j] = fmt.Sprintf("$%d", len(args)+j+1)
		}
		placeholders[i] = fmt.Sprintf("(%s)", fmt.Sprintf("%s", rowPlaceholders))
		args = append(args, row...)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		table,
		fmt.Sprintf("%s", columns),
		fmt.Sprintf("%s", placeholders))

	_, _, err := cp.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}

	cp.logger.Info("Batch insert completed",
		"table", table,
		"rows", len(values),
		"duration", time.Since(start))

	return nil
}

// BatchUpdate performs optimized batch update operations
func (cp *ConnectionPool) BatchUpdate(ctx context.Context, updates []BatchUpdateItem) error {
	if len(updates) == 0 {
		return nil
	}

	start := time.Now()

	err := cp.Transaction(ctx, func(tx *sql.Tx) error {
		for _, update := range updates {
			_, err := tx.ExecContext(ctx, update.Query, update.Args...)
			if err != nil {
				return fmt.Errorf("batch update item failed: %w", err)
			}
		}
		return nil
	})

	cp.logger.Info("Batch update completed",
		"updates", len(updates),
		"duration", time.Since(start))
	
	return err
}

// BatchUpdateItem represents a single update operation in a batch
type BatchUpdateItem struct {
	Query string
	Args  []interface{}
}

// ConnectionPoolMonitor monitors connection pool performance
type ConnectionPoolMonitor struct {
	pool   *ConnectionPool
	logger Logger
	stopCh chan struct{}
}

// NewConnectionPoolMonitor creates a new connection pool monitor
func NewConnectionPoolMonitor(pool *ConnectionPool, logger Logger) *ConnectionPoolMonitor {
	return &ConnectionPoolMonitor{
		pool:   pool,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// Start starts monitoring the connection pool
func (cpm *ConnectionPoolMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cpm.stopCh:
			return
		case <-ticker.C:
			cpm.logPoolStats()
		}
	}
}

// Stop stops the connection pool monitor
func (cpm *ConnectionPoolMonitor) Stop() {
	close(cpm.stopCh)
}

// logPoolStats logs current pool statistics
func (cpm *ConnectionPoolMonitor) logPoolStats() {
	stats := cpm.pool.GetStats()
	
	cpm.logger.Info("Database connection pool stats",
		"open_connections", stats.OpenConnections,
		"in_use_connections", stats.InUseConnections,
		"idle_connections", stats.IdleConnections,
		"wait_count", stats.WaitCount,
		"wait_duration", stats.WaitDuration,
		"total_queries", stats.TotalQueries,
		"successful_queries", stats.SuccessfulQueries,
		"failed_queries", stats.FailedQueries,
		"avg_query_time", stats.AverageQueryTime)

	// Check for potential issues
	if stats.InUseConnections >= stats.MaxOpenConnections*80/100 {
		cpm.logger.Warn("High connection usage detected",
			"usage_percent", float64(stats.InUseConnections)/float64(stats.MaxOpenConnections)*100)
	}

	if stats.WaitCount > 0 {
		cpm.logger.Warn("Connection waits detected",
			"wait_count", stats.WaitCount,
			"avg_wait_duration", stats.WaitDuration)
	}

	if stats.TotalQueries > 0 {
		errorRate := float64(stats.FailedQueries) / float64(stats.TotalQueries) * 100
		if errorRate > 5.0 {
			cpm.logger.Warn("High query error rate detected",
				"error_rate_percent", errorRate)
		}
	}
}
