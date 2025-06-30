package database

import (
	"context"
	"fmt"
	"time"

	"go-coffee-ai-agents/internal/config"
	"go-coffee-ai-agents/internal/observability"
)

// Manager manages database connections, migrations, and repositories
type Manager struct {
	config            config.DatabaseConfig
	connectionManager *ConnectionManager
	migrationManager  *MigrationManager
	db                *DB
	logger            *observability.StructuredLogger
	metrics           *observability.MetricsCollector
	tracing           *observability.TracingHelper
	
	// Repositories
	beverageRepo BeverageRepository
}

// NewManager creates a new database manager
func NewManager(
	config config.DatabaseConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *Manager {
	return &Manager{
		config:  config,
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// Initialize initializes the database connection and runs migrations
func (m *Manager) Initialize(ctx context.Context) error {
	ctx, span := m.tracing.StartDatabaseSpan(ctx, "INITIALIZE", "database")
	defer span.End()

	m.logger.InfoContext(ctx, "Initializing database manager")

	// Create connection manager
	m.connectionManager = NewConnectionManager(m.config, m.logger, m.metrics, m.tracing)

	// Establish database connection
	db, err := m.connectionManager.Connect(ctx)
	if err != nil {
		m.tracing.RecordError(span, err, "Failed to connect to database")
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	m.db = db

	// Create migration manager
	m.migrationManager = NewMigrationManager(m.db, m.logger, m.tracing)

	// Run migrations
	if err := m.migrationManager.Migrate(ctx); err != nil {
		m.tracing.RecordError(span, err, "Failed to run migrations")
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize repositories
	m.initializeRepositories()

	m.tracing.RecordSuccess(span, "Database manager initialized successfully")
	m.logger.InfoContext(ctx, "Database manager initialized successfully")

	return nil
}

// initializeRepositories initializes all repositories
func (m *Manager) initializeRepositories() {
	m.beverageRepo = NewPostgresBeverageRepository(m.db, m.logger, m.metrics, m.tracing)
}

// GetDB returns the database connection
func (m *Manager) GetDB() *DB {
	return m.db
}

// GetBeverageRepository returns the beverage repository
func (m *Manager) GetBeverageRepository() BeverageRepository {
	return m.beverageRepo
}

// BeginTransaction starts a new database transaction
func (m *Manager) BeginTransaction(ctx context.Context) (*Transaction, error) {
	ctx, span := m.tracing.StartDatabaseSpan(ctx, "BEGIN_TRANSACTION", "database")
	defer span.End()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		m.tracing.RecordError(span, err, "Failed to begin transaction")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	m.tracing.RecordSuccess(span, "Transaction started successfully")
	
	return &Transaction{
		tx:           tx,
		manager:      m,
		logger:       m.logger,
		metrics:      m.metrics,
		tracing:      m.tracing,
		beverageRepo: m.beverageRepo.WithTx(tx).(BeverageRepository),
	}, nil
}

// HealthCheck performs a comprehensive database health check
func (m *Manager) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	ctx, span := m.tracing.StartDatabaseSpan(ctx, "HEALTH_CHECK", "database")
	defer span.End()

	start := time.Now()
	status := &HealthStatus{
		Timestamp: start,
		Healthy:   true,
		Details:   make(map[string]interface{}),
	}

	// Check database connection
	if err := m.connectionManager.HealthCheck(ctx); err != nil {
		status.Healthy = false
		status.Error = err.Error()
		status.Details["connection_error"] = err.Error()
		m.tracing.RecordError(span, err, "Database health check failed")
		return status, nil
	}

	// Get connection statistics
	stats := m.connectionManager.GetStats()
	status.Details["connection_stats"] = map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}

	// Test a simple query
	var result int
	err := m.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		status.Healthy = false
		status.Error = fmt.Sprintf("Query test failed: %v", err)
		status.Details["query_error"] = err.Error()
		m.tracing.RecordError(span, err, "Database query test failed")
		return status, nil
	}

	// Check migration status
	appliedMigrations, err := m.migrationManager.getAppliedMigrations(ctx)
	if err != nil {
		status.Details["migration_check_error"] = err.Error()
	} else {
		status.Details["applied_migrations"] = len(appliedMigrations)
		status.Details["latest_migration"] = 0
		if len(appliedMigrations) > 0 {
			status.Details["latest_migration"] = appliedMigrations[len(appliedMigrations)-1]
		}
	}

	status.ResponseTime = time.Since(start)
	status.Details["response_time_ms"] = status.ResponseTime.Milliseconds()

	m.tracing.RecordSuccess(span, "Database health check completed")
	m.logger.DebugContext(ctx, "Database health check completed",
		"healthy", status.Healthy,
		"response_time_ms", status.ResponseTime.Milliseconds())

	return status, nil
}

// GetStatistics returns database usage statistics
func (m *Manager) GetStatistics(ctx context.Context) (*DatabaseStatistics, error) {
	ctx, span := m.tracing.StartDatabaseSpan(ctx, "GET_STATISTICS", "database")
	defer span.End()

	stats := &DatabaseStatistics{
		Timestamp: time.Now(),
	}

	// Get connection pool statistics
	dbStats := m.connectionManager.GetStats()
	stats.ConnectionPool = ConnectionPoolStats{
		OpenConnections:     dbStats.OpenConnections,
		InUse:              dbStats.InUse,
		Idle:               dbStats.Idle,
		WaitCount:          dbStats.WaitCount,
		WaitDuration:       dbStats.WaitDuration,
		MaxIdleClosed:      dbStats.MaxIdleClosed,
		MaxIdleTimeClosed:  dbStats.MaxIdleTimeClosed,
		MaxLifetimeClosed:  dbStats.MaxLifetimeClosed,
	}

	// Get table statistics
	tableStats, err := m.getTableStatistics(ctx)
	if err != nil {
		m.logger.WarnContext(ctx, "Failed to get table statistics", "error", err.Error())
		stats.Tables = make(map[string]TableStats)
	} else {
		stats.Tables = tableStats
	}

	// Get migration statistics
	appliedMigrations, err := m.migrationManager.getAppliedMigrations(ctx)
	if err != nil {
		m.logger.WarnContext(ctx, "Failed to get migration statistics", "error", err.Error())
	} else {
		stats.Migrations = MigrationStats{
			Applied: len(appliedMigrations),
			Latest:  0,
		}
		if len(appliedMigrations) > 0 {
			stats.Migrations.Latest = appliedMigrations[len(appliedMigrations)-1]
		}
	}

	m.tracing.RecordSuccess(span, "Database statistics retrieved")
	return stats, nil
}

// getTableStatistics retrieves statistics for database tables
func (m *Manager) getTableStatistics(ctx context.Context) (map[string]TableStats, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			n_tup_ins as inserts,
			n_tup_upd as updates,
			n_tup_del as deletes,
			n_live_tup as live_tuples,
			n_dead_tup as dead_tuples,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze
		FROM pg_stat_user_tables
		WHERE schemaname = 'public'`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table statistics: %w", err)
	}
	defer rows.Close()

	tableStats := make(map[string]TableStats)
	for rows.Next() {
		var schema, table string
		var stats TableStats
		var lastVacuum, lastAutovacuum, lastAnalyze, lastAutoanalyze *time.Time

		err := rows.Scan(
			&schema,
			&table,
			&stats.Inserts,
			&stats.Updates,
			&stats.Deletes,
			&stats.LiveTuples,
			&stats.DeadTuples,
			&lastVacuum,
			&lastAutovacuum,
			&lastAnalyze,
			&lastAutoanalyze,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table statistics: %w", err)
		}

		if lastVacuum != nil {
			stats.LastVacuum = *lastVacuum
		}
		if lastAutovacuum != nil {
			stats.LastAutovacuum = *lastAutovacuum
		}
		if lastAnalyze != nil {
			stats.LastAnalyze = *lastAnalyze
		}
		if lastAutoanalyze != nil {
			stats.LastAutoanalyze = *lastAutoanalyze
		}

		tableStats[table] = stats
	}

	return tableStats, nil
}

// Migrate runs database migrations
func (m *Manager) Migrate(ctx context.Context) error {
	return m.migrationManager.Migrate(ctx)
}

// Rollback rolls back the last migration
func (m *Manager) Rollback(ctx context.Context) error {
	return m.migrationManager.Rollback(ctx)
}

// Close closes the database connection
func (m *Manager) Close(ctx context.Context) error {
	ctx, span := m.tracing.StartDatabaseSpan(ctx, "CLOSE", "database")
	defer span.End()

	m.logger.InfoContext(ctx, "Closing database manager")

	if m.connectionManager != nil {
		if err := m.connectionManager.Close(ctx); err != nil {
			m.tracing.RecordError(span, err, "Failed to close database connection")
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	m.tracing.RecordSuccess(span, "Database manager closed successfully")
	m.logger.InfoContext(ctx, "Database manager closed successfully")

	return nil
}

// Transaction represents a database transaction with repositories
type Transaction struct {
	tx           *Tx
	manager      *Manager
	logger       *observability.StructuredLogger
	metrics      *observability.MetricsCollector
	tracing      *observability.TracingHelper
	beverageRepo BeverageRepository
}

// GetBeverageRepository returns the beverage repository for this transaction
func (t *Transaction) GetBeverageRepository() BeverageRepository {
	return t.beverageRepo
}

// Commit commits the transaction
func (t *Transaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// HealthStatus represents the health status of the database
type HealthStatus struct {
	Timestamp    time.Time              `json:"timestamp"`
	Healthy      bool                   `json:"healthy"`
	Error        string                 `json:"error,omitempty"`
	ResponseTime time.Duration          `json:"response_time"`
	Details      map[string]interface{} `json:"details"`
}

// DatabaseStatistics represents database usage statistics
type DatabaseStatistics struct {
	Timestamp      time.Time                `json:"timestamp"`
	ConnectionPool ConnectionPoolStats      `json:"connection_pool"`
	Tables         map[string]TableStats    `json:"tables"`
	Migrations     MigrationStats           `json:"migrations"`
}

// ConnectionPoolStats represents connection pool statistics
type ConnectionPoolStats struct {
	OpenConnections     int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed  int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
}

// TableStats represents statistics for a database table
type TableStats struct {
	Inserts         int64     `json:"inserts"`
	Updates         int64     `json:"updates"`
	Deletes         int64     `json:"deletes"`
	LiveTuples      int64     `json:"live_tuples"`
	DeadTuples      int64     `json:"dead_tuples"`
	LastVacuum      time.Time `json:"last_vacuum"`
	LastAutovacuum  time.Time `json:"last_autovacuum"`
	LastAnalyze     time.Time `json:"last_analyze"`
	LastAutoanalyze time.Time `json:"last_autoanalyze"`
}

// MigrationStats represents migration statistics
type MigrationStats struct {
	Applied int `json:"applied"`
	Latest  int `json:"latest"`
}

// Global database manager instance
var globalManager *Manager

// InitGlobalManager initializes the global database manager
func InitGlobalManager(
	config config.DatabaseConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) error {
	globalManager = NewManager(config, logger, metrics, tracing)
	return globalManager.Initialize(context.Background())
}

// GetGlobalManager returns the global database manager
func GetGlobalManager() *Manager {
	return globalManager
}

// CloseGlobalManager closes the global database manager
func CloseGlobalManager(ctx context.Context) error {
	if globalManager == nil {
		return nil
	}
	return globalManager.Close(ctx)
}
