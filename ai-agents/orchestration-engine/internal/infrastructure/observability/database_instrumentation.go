package observability

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DatabaseInstrumentation provides OpenTelemetry instrumentation for database operations
type DatabaseInstrumentation struct {
	telemetry *TelemetryManager
	logger    Logger
}

// NewDatabaseInstrumentation creates a new database instrumentation
func NewDatabaseInstrumentation(telemetry *TelemetryManager, logger Logger) *DatabaseInstrumentation {
	return &DatabaseInstrumentation{
		telemetry: telemetry,
		logger:    logger,
	}
}

// InstrumentedDB wraps a database connection with instrumentation
type InstrumentedDB struct {
	db              *sql.DB
	instrumentation *DatabaseInstrumentation
	dbName          string
	dbSystem        string
}

// InstrumentedTx wraps a database transaction with instrumentation
type InstrumentedTx struct {
	tx              *sql.Tx
	instrumentation *DatabaseInstrumentation
	dbName          string
	dbSystem        string
}

// InstrumentedStmt wraps a prepared statement with instrumentation
type InstrumentedStmt struct {
	stmt            *sql.Stmt
	instrumentation *DatabaseInstrumentation
	query           string
	dbName          string
	dbSystem        string
}

// InstrumentedRows wraps query results with instrumentation
type InstrumentedRows struct {
	rows            *sql.Rows
	instrumentation *DatabaseInstrumentation
	span            trace.Span
}

// NewInstrumentedDB creates an instrumented database connection
func (di *DatabaseInstrumentation) NewInstrumentedDB(db *sql.DB, dbName, dbSystem string) *InstrumentedDB {
	return &InstrumentedDB{
		db:              db,
		instrumentation: di,
		dbName:          dbName,
		dbSystem:        dbSystem,
	}
}

// Query executes a query with instrumentation
func (idb *InstrumentedDB) Query(ctx context.Context, query string, args ...interface{}) (*InstrumentedRows, error) {
	start := time.Now()
	
	// Start span for database query
	spanName := fmt.Sprintf("db.query %s", idb.dbName)
	ctx, span := idb.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", idb.dbSystem),
			attribute.String("db.name", idb.dbName),
			attribute.String("db.operation", "query"),
			attribute.String("db.statement", sanitizeQuery(query)),
		),
	)

	// Add query parameters (sanitized)
	if len(args) > 0 {
		span.SetAttributes(attribute.Int("db.parameters.count", len(args)))
	}

	// Execute query
	rows, err := idb.db.QueryContext(ctx, query, args...)
	
	duration := time.Since(start)

	// Record metrics
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		idb.instrumentation.telemetry.RecordError(ctx, "database_query_error", "database")
	}

	// Update span with timing and status
	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	// Record custom metrics
	idb.recordDatabaseMetrics(ctx, "query", status, duration)

	if err != nil {
		span.End()
		return nil, err
	}

	return &InstrumentedRows{
		rows:            rows,
		instrumentation: idb.instrumentation,
		span:            span,
	}, nil
}

// QueryRow executes a query that returns a single row with instrumentation
func (idb *InstrumentedDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	
	// Start span for database query
	spanName := fmt.Sprintf("db.query_row %s", idb.dbName)
	ctx, span := idb.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", idb.dbSystem),
			attribute.String("db.name", idb.dbName),
			attribute.String("db.operation", "query_row"),
			attribute.String("db.statement", sanitizeQuery(query)),
		),
	)
	defer span.End()

	// Add query parameters (sanitized)
	if len(args) > 0 {
		span.SetAttributes(attribute.Int("db.parameters.count", len(args)))
	}

	// Execute query
	row := idb.db.QueryRowContext(ctx, query, args...)
	
	duration := time.Since(start)

	// Update span with timing
	span.SetAttributes(
		attribute.String("db.status", "success"),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	// Record custom metrics
	idb.recordDatabaseMetrics(ctx, "query_row", "success", duration)

	return row
}

// Exec executes a query without returning rows with instrumentation
func (idb *InstrumentedDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	
	// Start span for database execution
	spanName := fmt.Sprintf("db.exec %s", idb.dbName)
	ctx, span := idb.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", idb.dbSystem),
			attribute.String("db.name", idb.dbName),
			attribute.String("db.operation", "exec"),
			attribute.String("db.statement", sanitizeQuery(query)),
		),
	)
	defer span.End()

	// Add query parameters (sanitized)
	if len(args) > 0 {
		span.SetAttributes(attribute.Int("db.parameters.count", len(args)))
	}

	// Execute query
	result, err := idb.db.ExecContext(ctx, query, args...)
	
	duration := time.Since(start)

	// Record metrics
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		idb.instrumentation.telemetry.RecordError(ctx, "database_exec_error", "database")
	} else {
		// Add result information if available
		if rowsAffected, err := result.RowsAffected(); err == nil {
			span.SetAttributes(attribute.Int64("db.rows_affected", rowsAffected))
		}
	}

	// Update span with timing and status
	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	// Record custom metrics
	idb.recordDatabaseMetrics(ctx, "exec", status, duration)

	return result, err
}

// Begin starts a transaction with instrumentation
func (idb *InstrumentedDB) Begin(ctx context.Context) (*InstrumentedTx, error) {
	start := time.Now()
	
	// Start span for transaction begin
	spanName := fmt.Sprintf("db.begin %s", idb.dbName)
	ctx, span := idb.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", idb.dbSystem),
			attribute.String("db.name", idb.dbName),
			attribute.String("db.operation", "begin"),
		),
	)
	defer span.End()

	// Begin transaction
	tx, err := idb.db.BeginTx(ctx, nil)
	
	duration := time.Since(start)

	// Record metrics
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		idb.instrumentation.telemetry.RecordError(ctx, "database_begin_error", "database")
	}

	// Update span with timing and status
	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	// Record custom metrics
	idb.recordDatabaseMetrics(ctx, "begin", status, duration)

	if err != nil {
		return nil, err
	}

	return &InstrumentedTx{
		tx:              tx,
		instrumentation: idb.instrumentation,
		dbName:          idb.dbName,
		dbSystem:        idb.dbSystem,
	}, nil
}

// Prepare creates a prepared statement with instrumentation
func (idb *InstrumentedDB) Prepare(ctx context.Context, query string) (*InstrumentedStmt, error) {
	start := time.Now()
	
	// Start span for statement preparation
	spanName := fmt.Sprintf("db.prepare %s", idb.dbName)
	ctx, span := idb.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", idb.dbSystem),
			attribute.String("db.name", idb.dbName),
			attribute.String("db.operation", "prepare"),
			attribute.String("db.statement", sanitizeQuery(query)),
		),
	)
	defer span.End()

	// Prepare statement
	stmt, err := idb.db.PrepareContext(ctx, query)
	
	duration := time.Since(start)

	// Record metrics
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		idb.instrumentation.telemetry.RecordError(ctx, "database_prepare_error", "database")
	}

	// Update span with timing and status
	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	// Record custom metrics
	idb.recordDatabaseMetrics(ctx, "prepare", status, duration)

	if err != nil {
		return nil, err
	}

	return &InstrumentedStmt{
		stmt:            stmt,
		instrumentation: idb.instrumentation,
		query:           query,
		dbName:          idb.dbName,
		dbSystem:        idb.dbSystem,
	}, nil
}

// Transaction methods for InstrumentedTx

// Query executes a query within a transaction with instrumentation
func (itx *InstrumentedTx) Query(ctx context.Context, query string, args ...interface{}) (*InstrumentedRows, error) {
	start := time.Now()
	
	// Start span for transaction query
	spanName := fmt.Sprintf("db.tx.query %s", itx.dbName)
	ctx, span := itx.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", itx.dbSystem),
			attribute.String("db.name", itx.dbName),
			attribute.String("db.operation", "tx_query"),
			attribute.String("db.statement", sanitizeQuery(query)),
		),
	)

	// Execute query
	rows, err := itx.tx.QueryContext(ctx, query, args...)
	
	duration := time.Since(start)

	// Record metrics and update span
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	if err != nil {
		span.End()
		return nil, err
	}

	return &InstrumentedRows{
		rows:            rows,
		instrumentation: itx.instrumentation,
		span:            span,
	}, nil
}

// Exec executes a query within a transaction with instrumentation
func (itx *InstrumentedTx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	
	// Start span for transaction execution
	spanName := fmt.Sprintf("db.tx.exec %s", itx.dbName)
	ctx, span := itx.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", itx.dbSystem),
			attribute.String("db.name", itx.dbName),
			attribute.String("db.operation", "tx_exec"),
			attribute.String("db.statement", sanitizeQuery(query)),
		),
	)
	defer span.End()

	// Execute query
	result, err := itx.tx.ExecContext(ctx, query, args...)
	
	duration := time.Since(start)

	// Record metrics and update span
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		if rowsAffected, err := result.RowsAffected(); err == nil {
			span.SetAttributes(attribute.Int64("db.rows_affected", rowsAffected))
		}
	}

	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	return result, err
}

// Commit commits the transaction with instrumentation
func (itx *InstrumentedTx) Commit(ctx context.Context) error {
	start := time.Now()
	
	// Start span for transaction commit
	spanName := fmt.Sprintf("db.tx.commit %s", itx.dbName)
	ctx, span := itx.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", itx.dbSystem),
			attribute.String("db.name", itx.dbName),
			attribute.String("db.operation", "commit"),
		),
	)
	defer span.End()

	// Commit transaction
	err := itx.tx.Commit()
	
	duration := time.Since(start)

	// Record metrics and update span
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	return err
}

// Rollback rolls back the transaction with instrumentation
func (itx *InstrumentedTx) Rollback(ctx context.Context) error {
	start := time.Now()
	
	// Start span for transaction rollback
	spanName := fmt.Sprintf("db.tx.rollback %s", itx.dbName)
	ctx, span := itx.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", itx.dbSystem),
			attribute.String("db.name", itx.dbName),
			attribute.String("db.operation", "rollback"),
		),
	)
	defer span.End()

	// Rollback transaction
	err := itx.tx.Rollback()
	
	duration := time.Since(start)

	// Record metrics and update span
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	return err
}

// Prepared statement methods for InstrumentedStmt

// Query executes a prepared statement query with instrumentation
func (istmt *InstrumentedStmt) Query(ctx context.Context, args ...interface{}) (*InstrumentedRows, error) {
	start := time.Now()
	
	// Start span for prepared statement query
	spanName := fmt.Sprintf("db.stmt.query %s", istmt.dbName)
	ctx, span := istmt.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", istmt.dbSystem),
			attribute.String("db.name", istmt.dbName),
			attribute.String("db.operation", "stmt_query"),
			attribute.String("db.statement", sanitizeQuery(istmt.query)),
		),
	)

	// Execute prepared statement
	rows, err := istmt.stmt.QueryContext(ctx, args...)
	
	duration := time.Since(start)

	// Record metrics and update span
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	if err != nil {
		span.End()
		return nil, err
	}

	return &InstrumentedRows{
		rows:            rows,
		instrumentation: istmt.instrumentation,
		span:            span,
	}, nil
}

// Close closes the prepared statement
func (istmt *InstrumentedStmt) Close() error {
	return istmt.stmt.Close()
}

// Rows methods for InstrumentedRows

// Next advances to the next row
func (ir *InstrumentedRows) Next() bool {
	return ir.rows.Next()
}

// Scan copies the columns in the current row into the values pointed at by dest
func (ir *InstrumentedRows) Scan(dest ...interface{}) error {
	return ir.rows.Scan(dest...)
}

// Close closes the rows iterator
func (ir *InstrumentedRows) Close() error {
	defer ir.span.End()
	return ir.rows.Close()
}

// Err returns the error, if any, that was encountered during iteration
func (ir *InstrumentedRows) Err() error {
	return ir.rows.Err()
}

// Helper methods

// recordDatabaseMetrics records database-specific metrics
func (idb *InstrumentedDB) recordDatabaseMetrics(ctx context.Context, operation, status string, duration time.Duration) {
	if !idb.instrumentation.telemetry.config.MetricsEnabled {
		return
	}

	// Record using the telemetry manager's custom metrics
	// Log the database operation for now (in a real implementation, we'd use custom metrics)
	idb.instrumentation.logger.Debug("Database operation recorded",
		"db_system", idb.dbSystem,
		"db_name", idb.dbName,
		"operation", operation,
		"status", status,
		"duration_ms", duration.Milliseconds(),
	)
}

// sanitizeQuery removes sensitive information from SQL queries for logging
func sanitizeQuery(query string) string {
	// In a real implementation, this would remove or mask sensitive data
	// For now, we'll just limit the length
	if len(query) > 200 {
		return query[:200] + "..."
	}
	return query
}

// Close closes the database connection
func (idb *InstrumentedDB) Close() error {
	return idb.db.Close()
}

// Ping verifies a connection to the database is still alive
func (idb *InstrumentedDB) Ping(ctx context.Context) error {
	start := time.Now()
	
	// Start span for ping
	spanName := fmt.Sprintf("db.ping %s", idb.dbName)
	ctx, span := idb.instrumentation.telemetry.StartSpan(ctx, spanName,
		trace.WithAttributes(
			attribute.String("db.system", idb.dbSystem),
			attribute.String("db.name", idb.dbName),
			attribute.String("db.operation", "ping"),
		),
	)
	defer span.End()

	// Ping database
	err := idb.db.PingContext(ctx)
	
	duration := time.Since(start)

	// Record metrics and update span
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.SetAttributes(
		attribute.String("db.status", status),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	return err
}
