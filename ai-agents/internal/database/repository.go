package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"go-coffee-ai-agents/internal/observability"
)

// Repository defines the base repository interface
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, entity interface{}) error
	GetByID(ctx context.Context, id interface{}, dest interface{}) error
	Update(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, id interface{}) error
	
	// Query operations
	FindAll(ctx context.Context, dest interface{}) error
	FindWhere(ctx context.Context, condition string, args []interface{}, dest interface{}) error
	Count(ctx context.Context, condition string, args []interface{}) (int64, error)
	
	// Transaction support
	WithTx(tx *Tx) Repository
}

// BaseRepository provides common repository functionality
type BaseRepository struct {
	db        *DB
	tx        *Tx
	tableName string
	logger    *observability.StructuredLogger
	metrics   *observability.MetricsCollector
	tracing   *observability.TracingHelper
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(
	db *DB,
	tableName string,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *BaseRepository {
	return &BaseRepository{
		db:        db,
		tableName: tableName,
		logger:    logger,
		metrics:   metrics,
		tracing:   tracing,
	}
}

// WithTx returns a new repository instance with transaction
func (r *BaseRepository) WithTx(tx *Tx) Repository {
	return &BaseRepository{
		db:        r.db,
		tx:        tx,
		tableName: r.tableName,
		logger:    r.logger,
		metrics:   r.metrics,
		tracing:   r.tracing,
	}
}

// getExecutor returns the appropriate executor (DB or Tx)
func (r *BaseRepository) getExecutor() Executor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Executor interface for database operations
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Create inserts a new entity into the database
func (r *BaseRepository) Create(ctx context.Context, entity interface{}) error {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "INSERT", r.tableName)
	defer span.End()

	start := time.Now()
	
	// Build INSERT query using reflection
	query, args, err := r.buildInsertQuery(entity)
	if err != nil {
		r.tracing.RecordError(span, err, "Failed to build insert query")
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	r.logger.DebugContext(ctx, "Creating entity",
		"table", r.tableName,
		"query", query)

	executor := r.getExecutor()
	result, err := executor.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to create entity")
		r.logger.ErrorContext(ctx, "Failed to create entity", err,
			"table", r.tableName,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to create entity in %s: %w", r.tableName, err)
	}

	// Get the number of affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "Could not get rows affected", "error", err.Error())
	}

	r.tracing.RecordSuccess(span, "Entity created successfully")
	r.logger.DebugContext(ctx, "Entity created successfully",
		"table", r.tableName,
		"rows_affected", rowsAffected,
		"duration_ms", duration.Milliseconds())

	return nil
}

// GetByID retrieves an entity by its ID
func (r *BaseRepository) GetByID(ctx context.Context, id interface{}, dest interface{}) error {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "SELECT", r.tableName)
	defer span.End()

	start := time.Now()
	
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", r.tableName)
	
	r.logger.DebugContext(ctx, "Getting entity by ID",
		"table", r.tableName,
		"id", id)

	executor := r.getExecutor()
	row := executor.QueryRowContext(ctx, query, id)
	
	err := r.scanRow(row, dest)
	duration := time.Since(start)

	if err != nil {
		if err == sql.ErrNoRows {
			r.tracing.RecordError(span, err, "Entity not found")
			r.logger.DebugContext(ctx, "Entity not found",
				"table", r.tableName,
				"id", id,
				"duration_ms", duration.Milliseconds())
			return ErrNotFound
		}
		
		r.tracing.RecordError(span, err, "Failed to get entity by ID")
		r.logger.ErrorContext(ctx, "Failed to get entity by ID", err,
			"table", r.tableName,
			"id", id,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to get entity from %s: %w", r.tableName, err)
	}

	r.tracing.RecordSuccess(span, "Entity retrieved successfully")
	r.logger.DebugContext(ctx, "Entity retrieved successfully",
		"table", r.tableName,
		"id", id,
		"duration_ms", duration.Milliseconds())

	return nil
}

// Update updates an existing entity
func (r *BaseRepository) Update(ctx context.Context, entity interface{}) error {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "UPDATE", r.tableName)
	defer span.End()

	start := time.Now()
	
	// Build UPDATE query using reflection
	query, args, err := r.buildUpdateQuery(entity)
	if err != nil {
		r.tracing.RecordError(span, err, "Failed to build update query")
		return fmt.Errorf("failed to build update query: %w", err)
	}

	r.logger.DebugContext(ctx, "Updating entity",
		"table", r.tableName,
		"query", query)

	executor := r.getExecutor()
	result, err := executor.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to update entity")
		r.logger.ErrorContext(ctx, "Failed to update entity", err,
			"table", r.tableName,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to update entity in %s: %w", r.tableName, err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "Could not get rows affected", "error", err.Error())
	} else if rowsAffected == 0 {
		r.tracing.RecordError(span, ErrNotFound, "No rows affected by update")
		return ErrNotFound
	}

	r.tracing.RecordSuccess(span, "Entity updated successfully")
	r.logger.DebugContext(ctx, "Entity updated successfully",
		"table", r.tableName,
		"rows_affected", rowsAffected,
		"duration_ms", duration.Milliseconds())

	return nil
}

// Delete removes an entity by its ID
func (r *BaseRepository) Delete(ctx context.Context, id interface{}) error {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "DELETE", r.tableName)
	defer span.End()

	start := time.Now()
	
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", r.tableName)
	
	r.logger.DebugContext(ctx, "Deleting entity",
		"table", r.tableName,
		"id", id)

	executor := r.getExecutor()
	result, err := executor.ExecContext(ctx, query, id)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to delete entity")
		r.logger.ErrorContext(ctx, "Failed to delete entity", err,
			"table", r.tableName,
			"id", id,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to delete entity from %s: %w", r.tableName, err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "Could not get rows affected", "error", err.Error())
	} else if rowsAffected == 0 {
		r.tracing.RecordError(span, ErrNotFound, "No rows affected by delete")
		return ErrNotFound
	}

	r.tracing.RecordSuccess(span, "Entity deleted successfully")
	r.logger.DebugContext(ctx, "Entity deleted successfully",
		"table", r.tableName,
		"id", id,
		"rows_affected", rowsAffected,
		"duration_ms", duration.Milliseconds())

	return nil
}

// FindAll retrieves all entities from the table
func (r *BaseRepository) FindAll(ctx context.Context, dest interface{}) error {
	return r.FindWhere(ctx, "", nil, dest)
}

// FindWhere retrieves entities based on a condition
func (r *BaseRepository) FindWhere(ctx context.Context, condition string, args []interface{}, dest interface{}) error {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "SELECT", r.tableName)
	defer span.End()

	start := time.Now()
	
	query := fmt.Sprintf("SELECT * FROM %s", r.tableName)
	if condition != "" {
		query += " WHERE " + condition
	}

	r.logger.DebugContext(ctx, "Finding entities",
		"table", r.tableName,
		"condition", condition,
		"args_count", len(args))

	executor := r.getExecutor()
	rows, err := executor.QueryContext(ctx, query, args...)
	if err != nil {
		duration := time.Since(start)
		r.tracing.RecordError(span, err, "Failed to execute find query")
		r.logger.ErrorContext(ctx, "Failed to execute find query", err,
			"table", r.tableName,
			"condition", condition,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to find entities in %s: %w", r.tableName, err)
	}
	defer rows.Close()

	err = r.scanRows(rows, dest)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to scan rows")
		r.logger.ErrorContext(ctx, "Failed to scan rows", err,
			"table", r.tableName,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to scan rows from %s: %w", r.tableName, err)
	}

	r.tracing.RecordSuccess(span, "Entities found successfully")
	r.logger.DebugContext(ctx, "Entities found successfully",
		"table", r.tableName,
		"condition", condition,
		"duration_ms", duration.Milliseconds())

	return nil
}

// Count returns the number of entities matching the condition
func (r *BaseRepository) Count(ctx context.Context, condition string, args []interface{}) (int64, error) {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "COUNT", r.tableName)
	defer span.End()

	start := time.Now()
	
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.tableName)
	if condition != "" {
		query += " WHERE " + condition
	}

	r.logger.DebugContext(ctx, "Counting entities",
		"table", r.tableName,
		"condition", condition)

	executor := r.getExecutor()
	row := executor.QueryRowContext(ctx, query, args...)
	
	var count int64
	err := row.Scan(&count)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to count entities")
		r.logger.ErrorContext(ctx, "Failed to count entities", err,
			"table", r.tableName,
			"condition", condition,
			"duration_ms", duration.Milliseconds())
		return 0, fmt.Errorf("failed to count entities in %s: %w", r.tableName, err)
	}

	r.tracing.RecordSuccess(span, "Entities counted successfully")
	r.logger.DebugContext(ctx, "Entities counted successfully",
		"table", r.tableName,
		"condition", condition,
		"count", count,
		"duration_ms", duration.Milliseconds())

	return count, nil
}

// Helper methods for query building and scanning

// buildInsertQuery builds an INSERT query using reflection
func (r *BaseRepository) buildInsertQuery(entity interface{}) (string, []interface{}, error) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("entity must be a struct")
	}

	t := v.Type()
	var columns []string
	var placeholders []string
	var args []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !value.CanInterface() {
			continue
		}

		// Get column name from db tag or use field name
		columnName := field.Tag.Get("db")
		if columnName == "" {
			columnName = strings.ToLower(field.Name)
		}

		// Skip id field for insert (assuming auto-increment)
		if columnName == "id" {
			continue
		}

		columns = append(columns, columnName)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)+1))
		args = append(args, value.Interface())
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		r.tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return query, args, nil
}

// buildUpdateQuery builds an UPDATE query using reflection
func (r *BaseRepository) buildUpdateQuery(entity interface{}) (string, []interface{}, error) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("entity must be a struct")
	}

	t := v.Type()
	var setParts []string
	var args []interface{}
	var idValue interface{}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !value.CanInterface() {
			continue
		}

		// Get column name from db tag or use field name
		columnName := field.Tag.Get("db")
		if columnName == "" {
			columnName = strings.ToLower(field.Name)
		}

		if columnName == "id" {
			idValue = value.Interface()
			continue
		}

		setParts = append(setParts, fmt.Sprintf("%s = $%d", columnName, len(args)+1))
		args = append(args, value.Interface())
	}

	if idValue == nil {
		return "", nil, fmt.Errorf("entity must have an id field")
	}

	args = append(args, idValue)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d",
		r.tableName,
		strings.Join(setParts, ", "),
		len(args))

	return query, args, nil
}

// scanRow scans a single row into a destination struct
func (r *BaseRepository) scanRow(row *sql.Row, dest interface{}) error {
	// This is a simplified implementation
	// In a real implementation, you would use reflection to map columns to struct fields
	return row.Scan(dest)
}

// scanRows scans multiple rows into a destination slice
func (r *BaseRepository) scanRows(rows *sql.Rows, dest interface{}) error {
	// This is a simplified implementation
	// In a real implementation, you would use reflection to map columns to struct fields
	// and build a slice of structs
	return fmt.Errorf("scanRows not implemented - use a proper ORM or implement reflection-based scanning")
}

// Common repository errors
var (
	ErrNotFound = fmt.Errorf("entity not found")
	ErrConflict = fmt.Errorf("entity conflict")
)
