package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/DimaJoyti/go-coffee/domain/shared"
)

// TenantAwareDB provides tenant-aware database operations
type TenantAwareDB interface {
	// GetConnection returns a database connection for the tenant
	GetConnection(ctx context.Context, tenantID shared.TenantID) (*sql.DB, error)

	// GetSchema returns the schema name for the tenant
	GetSchema(tenantID shared.TenantID) string

	// GetTableName returns the table name with tenant prefix/suffix
	GetTableName(tenantID shared.TenantID, baseTableName string) string

	// ExecuteInTenantContext executes a function within tenant context
	ExecuteInTenantContext(ctx context.Context, tenantID shared.TenantID, fn func(*sql.DB) error) error
}

// MultiTenantDB implements TenantAwareDB with different isolation strategies
type MultiTenantDB struct {
	isolationLevel   shared.TenantIsolationLevel
	connections      map[string]*sql.DB // For database-per-tenant
	sharedConnection *sql.DB            // For shared database
	schemaPrefix     string             // For schema-per-tenant
}

// NewMultiTenantDB creates a new multi-tenant database instance
func NewMultiTenantDB(isolationLevel shared.TenantIsolationLevel, sharedConnection *sql.DB) *MultiTenantDB {
	return &MultiTenantDB{
		isolationLevel:   isolationLevel,
		connections:      make(map[string]*sql.DB),
		sharedConnection: sharedConnection,
		schemaPrefix:     "tenant_",
	}
}

// GetConnection returns a database connection for the tenant
func (db *MultiTenantDB) GetConnection(ctx context.Context, tenantID shared.TenantID) (*sql.DB, error) {
	switch db.isolationLevel {
	case shared.DatabasePerTenant:
		return db.getDatabasePerTenantConnection(tenantID)
	case shared.SchemaPerTenant, shared.SharedDatabase:
		return db.sharedConnection, nil
	default:
		return nil, fmt.Errorf("unsupported isolation level: %s", db.isolationLevel.String())
	}
}

// GetSchema returns the schema name for the tenant
func (db *MultiTenantDB) GetSchema(tenantID shared.TenantID) string {
	switch db.isolationLevel {
	case shared.SchemaPerTenant:
		return db.schemaPrefix + tenantID.Value()
	case shared.DatabasePerTenant:
		return "public" // Default schema in tenant-specific database
	case shared.SharedDatabase:
		return "public" // Shared schema
	default:
		return "public"
	}
}

// GetTableName returns the table name with tenant context
func (db *MultiTenantDB) GetTableName(tenantID shared.TenantID, baseTableName string) string {
	switch db.isolationLevel {
	case shared.SchemaPerTenant:
		return fmt.Sprintf("%s.%s", db.GetSchema(tenantID), baseTableName)
	case shared.DatabasePerTenant:
		return baseTableName // Table name without schema prefix
	case shared.SharedDatabase:
		return baseTableName // Shared table with tenant_id column
	default:
		return baseTableName
	}
}

// ExecuteInTenantContext executes a function within tenant context
func (db *MultiTenantDB) ExecuteInTenantContext(ctx context.Context, tenantID shared.TenantID, fn func(*sql.DB) error) error {
	conn, err := db.GetConnection(ctx, tenantID)
	if err != nil {
		return err
	}

	// Set schema for schema-per-tenant isolation
	if db.isolationLevel == shared.SchemaPerTenant {
		schema := db.GetSchema(tenantID)
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", schema)); err != nil {
			return fmt.Errorf("failed to set schema: %w", err)
		}
	}

	return fn(conn)
}

// getDatabasePerTenantConnection returns a connection for database-per-tenant isolation
func (db *MultiTenantDB) getDatabasePerTenantConnection(tenantID shared.TenantID) (*sql.DB, error) {
	connectionKey := tenantID.Value()

	if conn, exists := db.connections[connectionKey]; exists {
		return conn, nil
	}

	// In a real implementation, you would create a new connection to the tenant-specific database
	// For now, we'll return an error indicating the connection needs to be configured
	return nil, fmt.Errorf("database connection for tenant %s not configured", tenantID.Value())
}

// BaseTenantAwareRepository provides common functionality for tenant-aware repositories
type BaseTenantAwareRepository struct {
	db         TenantAwareDB
	tableName  string
	entityName string
}

// NewBaseTenantAwareRepository creates a new base tenant-aware repository
func NewBaseTenantAwareRepository(db TenantAwareDB, tableName, entityName string) *BaseTenantAwareRepository {
	return &BaseTenantAwareRepository{
		db:         db,
		tableName:  tableName,
		entityName: entityName,
	}
}

// BuildSelectQuery builds a SELECT query with tenant awareness
func (r *BaseTenantAwareRepository) BuildSelectQuery(tenantID shared.TenantID, columns []string, whereClause string, args []interface{}) (string, []interface{}) {
	tableName := r.db.GetTableName(tenantID, r.tableName)
	columnsStr := strings.Join(columns, ", ")

	query := fmt.Sprintf("SELECT %s FROM %s", columnsStr, tableName)

	// Add tenant filter for shared database
	if multiTenantDB, ok := r.db.(*MultiTenantDB); ok && multiTenantDB.isolationLevel == shared.SharedDatabase {
		tenantFilter := "tenant_id = ?"
		if whereClause != "" {
			whereClause = fmt.Sprintf("(%s) AND %s", whereClause, tenantFilter)
		} else {
			whereClause = tenantFilter
		}
		args = append(args, tenantID.Value())
	}

	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	return query, args
}

// BuildInsertQuery builds an INSERT query with tenant awareness
func (r *BaseTenantAwareRepository) BuildInsertQuery(tenantID shared.TenantID, columns []string, values []interface{}) (string, []interface{}) {
	tableName := r.db.GetTableName(tenantID, r.tableName)

	// Add tenant_id for shared database
	if multiTenantDB, ok := r.db.(*MultiTenantDB); ok && multiTenantDB.isolationLevel == shared.SharedDatabase {
		columns = append(columns, "tenant_id")
		values = append(values, tenantID.Value())
	}

	columnsStr := strings.Join(columns, ", ")
	placeholders := strings.Repeat("?,", len(columns))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columnsStr, placeholders)

	return query, values
}

// BuildUpdateQuery builds an UPDATE query with tenant awareness
func (r *BaseTenantAwareRepository) BuildUpdateQuery(tenantID shared.TenantID, setClause string, whereClause string, args []interface{}) (string, []interface{}) {
	tableName := r.db.GetTableName(tenantID, r.tableName)

	query := fmt.Sprintf("UPDATE %s SET %s", tableName, setClause)

	// Add tenant filter for shared database
	if multiTenantDB, ok := r.db.(*MultiTenantDB); ok && multiTenantDB.isolationLevel == shared.SharedDatabase {
		tenantFilter := "tenant_id = ?"
		if whereClause != "" {
			whereClause = fmt.Sprintf("(%s) AND %s", whereClause, tenantFilter)
		} else {
			whereClause = tenantFilter
		}
		args = append(args, tenantID.Value())
	}

	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	return query, args
}

// BuildDeleteQuery builds a DELETE query with tenant awareness
func (r *BaseTenantAwareRepository) BuildDeleteQuery(tenantID shared.TenantID, whereClause string, args []interface{}) (string, []interface{}) {
	tableName := r.db.GetTableName(tenantID, r.tableName)

	query := fmt.Sprintf("DELETE FROM %s", tableName)

	// Add tenant filter for shared database
	if multiTenantDB, ok := r.db.(*MultiTenantDB); ok && multiTenantDB.isolationLevel == shared.SharedDatabase {
		tenantFilter := "tenant_id = ?"
		if whereClause != "" {
			whereClause = fmt.Sprintf("(%s) AND %s", whereClause, tenantFilter)
		} else {
			whereClause = tenantFilter
		}
		args = append(args, tenantID.Value())
	}

	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	return query, args
}

// ExecuteQuery executes a query within tenant context
func (r *BaseTenantAwareRepository) ExecuteQuery(ctx context.Context, tenantID shared.TenantID, query string, args []interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error

	executeErr := r.db.ExecuteInTenantContext(ctx, tenantID, func(db *sql.DB) error {
		rows, err = db.QueryContext(ctx, query, args...)
		return err
	})

	if executeErr != nil {
		return nil, executeErr
	}

	return rows, err
}

// ExecuteNonQuery executes a non-query (INSERT, UPDATE, DELETE) within tenant context
func (r *BaseTenantAwareRepository) ExecuteNonQuery(ctx context.Context, tenantID shared.TenantID, query string, args []interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	executeErr := r.db.ExecuteInTenantContext(ctx, tenantID, func(db *sql.DB) error {
		result, err = db.ExecContext(ctx, query, args...)
		return err
	})

	if executeErr != nil {
		return nil, executeErr
	}

	return result, err
}

// TenantAwareTransaction provides transaction support with tenant awareness
type TenantAwareTransaction struct {
	tx       *sql.Tx
	tenantID shared.TenantID
	db       TenantAwareDB
}

// NewTenantAwareTransaction creates a new tenant-aware transaction
func NewTenantAwareTransaction(ctx context.Context, db TenantAwareDB, tenantID shared.TenantID) (*TenantAwareTransaction, error) {
	conn, err := db.GetConnection(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Set schema for schema-per-tenant isolation
	if multiTenantDB, ok := db.(*MultiTenantDB); ok && multiTenantDB.isolationLevel == shared.SchemaPerTenant {
		schema := db.GetSchema(tenantID)
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", schema)); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to set schema in transaction: %w", err)
		}
	}

	return &TenantAwareTransaction{
		tx:       tx,
		tenantID: tenantID,
		db:       db,
	}, nil
}

// Commit commits the transaction
func (t *TenantAwareTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *TenantAwareTransaction) Rollback() error {
	return t.tx.Rollback()
}

// ExecContext executes a query within the transaction
func (t *TenantAwareTransaction) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// QueryContext executes a query within the transaction
func (t *TenantAwareTransaction) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row within the transaction
func (t *TenantAwareTransaction) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// TenantContextValidator validates tenant context in requests
type TenantContextValidator struct{}

// NewTenantContextValidator creates a new tenant context validator
func NewTenantContextValidator() *TenantContextValidator {
	return &TenantContextValidator{}
}

// ValidateContext validates that tenant context exists and is valid
func (v *TenantContextValidator) ValidateContext(ctx context.Context) (*shared.TenantContext, error) {
	tenantCtx, err := shared.FromContext(ctx)
	if err != nil {
		return nil, errors.New("tenant context is required")
	}

	if tenantCtx.TenantID().IsEmpty() {
		return nil, errors.New("tenant ID cannot be empty")
	}

	return tenantCtx, nil
}

// ValidateTenantAccess validates that the request has access to the specified tenant
func (v *TenantContextValidator) ValidateTenantAccess(ctx context.Context, targetTenantID shared.TenantID) error {
	tenantCtx, err := v.ValidateContext(ctx)
	if err != nil {
		return err
	}

	if !tenantCtx.TenantID().Equals(targetTenantID) {
		return errors.New("access denied: tenant mismatch")
	}

	return nil
}

// ValidateFeatureAccess validates that the tenant has access to a specific feature
func (v *TenantContextValidator) ValidateFeatureAccess(ctx context.Context, feature string) error {
	tenantCtx, err := v.ValidateContext(ctx)
	if err != nil {
		return err
	}

	if !tenantCtx.HasFeature(feature) {
		return fmt.Errorf("access denied: feature '%s' not available for current subscription", feature)
	}

	return nil
}
