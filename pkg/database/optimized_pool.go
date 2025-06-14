package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"
)

// OptimizedDatabase provides high-performance database operations using standard library
type OptimizedDatabase struct {
	writeDB     *sql.DB
	readDBs     []*sql.DB
	config      *OptimizedConfig
	logger      *zap.Logger
	metrics     *DatabaseMetrics
	healthCheck *HealthChecker
	mu          sync.RWMutex
}

// OptimizedConfig contains database optimization configuration
type OptimizedConfig struct {
	WriteConnectionString string
	ReadConnectionStrings []string
	MaxConnections        int32
	MinConnections        int32
	MaxConnLifetime       time.Duration
	MaxConnIdleTime       time.Duration
	HealthCheckPeriod     time.Duration
	DefaultQueryTimeout   time.Duration
	SlowQueryThreshold    time.Duration
	ConnectionTimeout     time.Duration
	ReadReplicaWeight     []int
	ReadReplicaFailover   bool
}

// DatabaseMetrics tracks database performance
type DatabaseMetrics struct {
	QueryCount        int64
	SlowQueryCount    int64
	ConnectionErrors  int64
	TotalConnections  int32
	ActiveConnections int32
	IdleConnections   int32
	AverageQueryTime  time.Duration
	mu                sync.RWMutex
}

// HealthChecker monitors database health
type HealthChecker struct {
	db       *OptimizedDatabase
	interval time.Duration
	stopCh   chan struct{}
}

// NewOptimizedDatabase creates a new optimized database instance
func NewOptimizedDatabase(config *OptimizedConfig, logger *zap.Logger) (*OptimizedDatabase, error) {
	db := &OptimizedDatabase{
		config:  config,
		logger:  logger,
		metrics: &DatabaseMetrics{},
	}

	// Initialize write database
	if err := db.initializeWriteDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize write database: %w", err)
	}

	// Initialize read databases
	if err := db.initializeReadDBs(); err != nil {
		return nil, fmt.Errorf("failed to initialize read databases: %w", err)
	}

	// Start health checker
	db.healthCheck = &HealthChecker{
		db:       db,
		interval: config.HealthCheckPeriod,
		stopCh:   make(chan struct{}),
	}
	go db.healthCheck.start()

	// Start metrics collection
	go db.startMetricsCollection()

	return db, nil
}

// initializeWriteDB sets up the write database connection
func (db *OptimizedDatabase) initializeWriteDB() error {
	writeDB, err := sql.Open("postgres", db.config.WriteConnectionString)
	if err != nil {
		return fmt.Errorf("failed to open write database: %w", err)
	}

	// Configure connection pool
	writeDB.SetMaxOpenConns(int(db.config.MaxConnections))
	writeDB.SetMaxIdleConns(int(db.config.MinConnections))
	writeDB.SetConnMaxLifetime(db.config.MaxConnLifetime)
	writeDB.SetConnMaxIdleTime(db.config.MaxConnIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), db.config.ConnectionTimeout)
	defer cancel()

	if err := writeDB.PingContext(ctx); err != nil {
		writeDB.Close()
		return fmt.Errorf("failed to ping write database: %w", err)
	}

	db.writeDB = writeDB
	atomic.StoreInt32(&db.metrics.TotalConnections, db.config.MaxConnections)

	db.logger.Info("Write database initialized successfully")
	return nil
}

// initializeReadDBs sets up read replica connections
func (db *OptimizedDatabase) initializeReadDBs() error {
	if len(db.config.ReadConnectionStrings) == 0 {
		// Use write DB as read DB if no read replicas configured
		db.readDBs = []*sql.DB{db.writeDB}
		return nil
	}

	db.readDBs = make([]*sql.DB, len(db.config.ReadConnectionStrings))

	for i, connStr := range db.config.ReadConnectionStrings {
		readDB, err := sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("failed to open read database %d: %w", i, err)
		}

		// Configure connection pool (smaller for read replicas)
		readDB.SetMaxOpenConns(int(db.config.MaxConnections / 2))
		readDB.SetMaxIdleConns(int(db.config.MinConnections))
		readDB.SetConnMaxLifetime(db.config.MaxConnLifetime)
		readDB.SetConnMaxIdleTime(db.config.MaxConnIdleTime)

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), db.config.ConnectionTimeout)
		if err := readDB.PingContext(ctx); err != nil {
			cancel()
			readDB.Close()
			if db.config.ReadReplicaFailover {
				db.logger.Warn("Read replica failed, continuing with available replicas",
					zap.Int("replica_index", i), zap.Error(err))
				continue
			}
			return fmt.Errorf("failed to ping read database %d: %w", i, err)
		}
		cancel()

		db.readDBs[i] = readDB
	}

	db.logger.Info("Read databases initialized successfully",
		zap.Int("replica_count", len(db.readDBs)))
	return nil
}

// ExecuteWrite executes a write operation
func (db *OptimizedDatabase) ExecuteWrite(ctx context.Context, query string, args ...interface{}) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		db.updateQueryMetrics(duration)
	}()

	// Add timeout to context if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, db.config.DefaultQueryTimeout)
		defer cancel()
	}

	atomic.AddInt64(&db.metrics.QueryCount, 1)

	_, err := db.writeDB.ExecContext(ctx, query, args...)
	if err != nil {
		atomic.AddInt64(&db.metrics.ConnectionErrors, 1)
		return fmt.Errorf("write query failed: %w", err)
	}

	return nil
}

// QueryRead executes a read operation on an available read replica
func (db *OptimizedDatabase) QueryRead(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		db.updateQueryMetrics(duration)
	}()

	// Add timeout to context if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, db.config.DefaultQueryTimeout)
		defer cancel()
	}

	atomic.AddInt64(&db.metrics.QueryCount, 1)

	// Select read database (simple round-robin for now)
	readDB := db.selectReadDB()

	rows, err := readDB.QueryContext(ctx, query, args...)
	if err != nil {
		atomic.AddInt64(&db.metrics.ConnectionErrors, 1)
		return nil, fmt.Errorf("read query failed: %w", err)
	}

	return rows, nil
}

// selectReadDB selects an available read database
func (db *OptimizedDatabase) selectReadDB() *sql.DB {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if len(db.readDBs) == 0 {
		return db.writeDB
	}

	// Simple round-robin selection
	// In production, you might want to implement weighted selection
	index := int(atomic.LoadInt64(&db.metrics.QueryCount)) % len(db.readDBs)
	return db.readDBs[index]
}

// updateQueryMetrics updates query performance metrics
func (db *OptimizedDatabase) updateQueryMetrics(duration time.Duration) {
	db.metrics.mu.Lock()
	defer db.metrics.mu.Unlock()

	// Update average query time
	if db.metrics.AverageQueryTime == 0 {
		db.metrics.AverageQueryTime = duration
	} else {
		db.metrics.AverageQueryTime = (db.metrics.AverageQueryTime + duration) / 2
	}

	// Track slow queries
	if duration > db.config.SlowQueryThreshold {
		atomic.AddInt64(&db.metrics.SlowQueryCount, 1)
		db.logger.Warn("Slow query detected",
			zap.Duration("duration", duration),
			zap.Duration("threshold", db.config.SlowQueryThreshold))
	}
}

// GetMetrics returns current database metrics
func (db *OptimizedDatabase) GetMetrics() DatabaseMetrics {
	db.metrics.mu.RLock()
	defer db.metrics.mu.RUnlock()

	// Update connection stats
	if db.writeDB != nil {
		stats := db.writeDB.Stats()
		atomic.StoreInt32(&db.metrics.ActiveConnections, int32(stats.OpenConnections))
		atomic.StoreInt32(&db.metrics.IdleConnections, int32(stats.Idle))
	}

	return *db.metrics
}

// Close closes all database connections
func (db *OptimizedDatabase) Close() {
	db.logger.Info("Closing database connections")

	// Stop health checker
	if db.healthCheck != nil {
		close(db.healthCheck.stopCh)
	}

	// Close write database
	if db.writeDB != nil {
		db.writeDB.Close()
	}

	// Close read databases
	for _, readDB := range db.readDBs {
		if readDB != nil && readDB != db.writeDB {
			readDB.Close()
		}
	}

	db.logger.Info("Database connections closed")
}

// Health checker methods

// start starts the health checker
func (hc *HealthChecker) start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkHealth()
		case <-hc.stopCh:
			return
		}
	}
}

// checkHealth performs health checks on all database connections
func (hc *HealthChecker) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check write database
	if err := hc.db.writeDB.PingContext(ctx); err != nil {
		hc.db.logger.Error("Write database health check failed", zap.Error(err))
		atomic.AddInt64(&hc.db.metrics.ConnectionErrors, 1)
	}

	// Check read databases
	for i, readDB := range hc.db.readDBs {
		if readDB != hc.db.writeDB {
			if err := readDB.PingContext(ctx); err != nil {
				hc.db.logger.Error("Read database health check failed",
					zap.Int("replica_index", i), zap.Error(err))
				atomic.AddInt64(&hc.db.metrics.ConnectionErrors, 1)
			}
		}
	}
}

// startMetricsCollection starts periodic metrics collection
func (db *OptimizedDatabase) startMetricsCollection() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			db.collectMetrics()
		}
	}
}

// collectMetrics collects and logs database metrics
func (db *OptimizedDatabase) collectMetrics() {
	metrics := db.GetMetrics()

	db.logger.Debug("Database metrics",
		zap.Int64("query_count", metrics.QueryCount),
		zap.Int64("slow_query_count", metrics.SlowQueryCount),
		zap.Int64("connection_errors", metrics.ConnectionErrors),
		zap.Int32("active_connections", metrics.ActiveConnections),
		zap.Int32("idle_connections", metrics.IdleConnections),
		zap.Duration("avg_query_time", metrics.AverageQueryTime),
	)
}
