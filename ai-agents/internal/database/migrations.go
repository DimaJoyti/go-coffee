package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"go-coffee-ai-agents/internal/observability"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Name        string
	UpSQL       string
	DownSQL     string
	Description string
}

// MigrationManager manages database migrations
type MigrationManager struct {
	db      *DB
	logger  *observability.StructuredLogger
	tracing *observability.TracingHelper
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(
	db *DB,
	logger *observability.StructuredLogger,
	tracing *observability.TracingHelper,
) *MigrationManager {
	return &MigrationManager{
		db:      db,
		logger:  logger,
		tracing: tracing,
	}
}

// GetMigrations returns all available migrations
func (mm *MigrationManager) GetMigrations() []Migration {
	migrations := []Migration{
		{
			Version:     1,
			Name:        "create_schema_migrations_table",
			Description: "Create schema migrations tracking table",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS schema_migrations (
					version INTEGER PRIMARY KEY,
					name VARCHAR(255) NOT NULL,
					applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					checksum VARCHAR(64)
				);
				
				CREATE INDEX IF NOT EXISTS idx_schema_migrations_applied_at 
				ON schema_migrations(applied_at);
			`,
			DownSQL: `DROP TABLE IF EXISTS schema_migrations;`,
		},
		{
			Version:     2,
			Name:        "create_beverages_table",
			Description: "Create beverages table for storing beverage recipes",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS beverages (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					name VARCHAR(255) NOT NULL,
					description TEXT,
					theme VARCHAR(100) NOT NULL,
					ingredients JSONB NOT NULL,
					metadata JSONB NOT NULL DEFAULT '{}',
					created_by VARCHAR(255) NOT NULL,
					rating DECIMAL(3,2) DEFAULT 0.0 CHECK (rating >= 0.0 AND rating <= 5.0),
					view_count INTEGER DEFAULT 0 CHECK (view_count >= 0),
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);
				
				-- Indexes for performance
				CREATE INDEX IF NOT EXISTS idx_beverages_theme ON beverages(theme);
				CREATE INDEX IF NOT EXISTS idx_beverages_created_by ON beverages(created_by);
				CREATE INDEX IF NOT EXISTS idx_beverages_rating ON beverages(rating DESC);
				CREATE INDEX IF NOT EXISTS idx_beverages_created_at ON beverages(created_at DESC);
				CREATE INDEX IF NOT EXISTS idx_beverages_view_count ON beverages(view_count DESC);
				
				-- GIN index for JSON ingredient search
				CREATE INDEX IF NOT EXISTS idx_beverages_ingredients_gin ON beverages USING GIN (ingredients);
				CREATE INDEX IF NOT EXISTS idx_beverages_metadata_gin ON beverages USING GIN (metadata);
				
				-- Composite index for popular beverages
				CREATE INDEX IF NOT EXISTS idx_beverages_popularity 
				ON beverages(rating DESC, view_count DESC, created_at DESC);
				
				-- Full-text search index on name and description
				CREATE INDEX IF NOT EXISTS idx_beverages_search 
				ON beverages USING GIN (to_tsvector('english', name || ' ' || COALESCE(description, '')));
			`,
			DownSQL: `DROP TABLE IF EXISTS beverages;`,
		},
		{
			Version:     3,
			Name:        "create_beverage_ratings_table",
			Description: "Create table for tracking individual beverage ratings",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS beverage_ratings (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					beverage_id UUID NOT NULL REFERENCES beverages(id) ON DELETE CASCADE,
					user_id VARCHAR(255) NOT NULL,
					rating DECIMAL(3,2) NOT NULL CHECK (rating >= 0.0 AND rating <= 5.0),
					comment TEXT,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					
					-- Ensure one rating per user per beverage
					UNIQUE(beverage_id, user_id)
				);
				
				-- Indexes
				CREATE INDEX IF NOT EXISTS idx_beverage_ratings_beverage_id ON beverage_ratings(beverage_id);
				CREATE INDEX IF NOT EXISTS idx_beverage_ratings_user_id ON beverage_ratings(user_id);
				CREATE INDEX IF NOT EXISTS idx_beverage_ratings_rating ON beverage_ratings(rating DESC);
				CREATE INDEX IF NOT EXISTS idx_beverage_ratings_created_at ON beverage_ratings(created_at DESC);
			`,
			DownSQL: `DROP TABLE IF EXISTS beverage_ratings;`,
		},
		{
			Version:     4,
			Name:        "create_beverage_tags_table",
			Description: "Create table for beverage tags and categories",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS beverage_tags (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					name VARCHAR(100) NOT NULL UNIQUE,
					description TEXT,
					color VARCHAR(7), -- Hex color code
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);
				
				CREATE TABLE IF NOT EXISTS beverage_tag_assignments (
					beverage_id UUID NOT NULL REFERENCES beverages(id) ON DELETE CASCADE,
					tag_id UUID NOT NULL REFERENCES beverage_tags(id) ON DELETE CASCADE,
					assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					assigned_by VARCHAR(255),
					
					PRIMARY KEY (beverage_id, tag_id)
				);
				
				-- Indexes
				CREATE INDEX IF NOT EXISTS idx_beverage_tags_name ON beverage_tags(name);
				CREATE INDEX IF NOT EXISTS idx_beverage_tag_assignments_beverage_id ON beverage_tag_assignments(beverage_id);
				CREATE INDEX IF NOT EXISTS idx_beverage_tag_assignments_tag_id ON beverage_tag_assignments(tag_id);
			`,
			DownSQL: `
				DROP TABLE IF EXISTS beverage_tag_assignments;
				DROP TABLE IF EXISTS beverage_tags;
			`,
		},
		{
			Version:     5,
			Name:        "create_beverage_favorites_table",
			Description: "Create table for user favorite beverages",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS beverage_favorites (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					beverage_id UUID NOT NULL REFERENCES beverages(id) ON DELETE CASCADE,
					user_id VARCHAR(255) NOT NULL,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					
					-- Ensure one favorite per user per beverage
					UNIQUE(beverage_id, user_id)
				);
				
				-- Indexes
				CREATE INDEX IF NOT EXISTS idx_beverage_favorites_beverage_id ON beverage_favorites(beverage_id);
				CREATE INDEX IF NOT EXISTS idx_beverage_favorites_user_id ON beverage_favorites(user_id);
				CREATE INDEX IF NOT EXISTS idx_beverage_favorites_created_at ON beverage_favorites(created_at DESC);
			`,
			DownSQL: `DROP TABLE IF EXISTS beverage_favorites;`,
		},
		{
			Version:     6,
			Name:        "add_beverage_analytics_columns",
			Description: "Add analytics columns to beverages table",
			UpSQL: `
				ALTER TABLE beverages 
				ADD COLUMN IF NOT EXISTS last_viewed_at TIMESTAMP WITH TIME ZONE,
				ADD COLUMN IF NOT EXISTS favorite_count INTEGER DEFAULT 0 CHECK (favorite_count >= 0),
				ADD COLUMN IF NOT EXISTS rating_count INTEGER DEFAULT 0 CHECK (rating_count >= 0),
				ADD COLUMN IF NOT EXISTS average_rating DECIMAL(3,2) DEFAULT 0.0 CHECK (average_rating >= 0.0 AND average_rating <= 5.0);
				
				-- Create index for analytics queries
				CREATE INDEX IF NOT EXISTS idx_beverages_analytics 
				ON beverages(favorite_count DESC, rating_count DESC, average_rating DESC);
			`,
			DownSQL: `
				ALTER TABLE beverages 
				DROP COLUMN IF EXISTS last_viewed_at,
				DROP COLUMN IF EXISTS favorite_count,
				DROP COLUMN IF EXISTS rating_count,
				DROP COLUMN IF EXISTS average_rating;
			`,
		},
		{
			Version:     7,
			Name:        "create_beverage_preparation_logs_table",
			Description: "Create table for tracking beverage preparation attempts",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS beverage_preparation_logs (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					beverage_id UUID NOT NULL REFERENCES beverages(id) ON DELETE CASCADE,
					prepared_by VARCHAR(255) NOT NULL,
					preparation_time_minutes INTEGER,
					actual_cost DECIMAL(10,2),
					notes TEXT,
					success BOOLEAN DEFAULT true,
					photos JSONB DEFAULT '[]',
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);
				
				-- Indexes
				CREATE INDEX IF NOT EXISTS idx_beverage_preparation_logs_beverage_id ON beverage_preparation_logs(beverage_id);
				CREATE INDEX IF NOT EXISTS idx_beverage_preparation_logs_prepared_by ON beverage_preparation_logs(prepared_by);
				CREATE INDEX IF NOT EXISTS idx_beverage_preparation_logs_created_at ON beverage_preparation_logs(created_at DESC);
				CREATE INDEX IF NOT EXISTS idx_beverage_preparation_logs_success ON beverage_preparation_logs(success);
			`,
			DownSQL: `DROP TABLE IF EXISTS beverage_preparation_logs;`,
		},
		{
			Version:     8,
			Name:        "create_beverage_search_functions",
			Description: "Create functions and triggers for enhanced beverage search",
			UpSQL: `
				-- Function to update beverage search vector
				CREATE OR REPLACE FUNCTION update_beverage_search_vector()
				RETURNS TRIGGER AS $$
				BEGIN
					NEW.search_vector := to_tsvector('english', 
						NEW.name || ' ' || 
						COALESCE(NEW.description, '') || ' ' || 
						NEW.theme || ' ' ||
						COALESCE(NEW.ingredients::text, '')
					);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
				
				-- Add search vector column
				ALTER TABLE beverages ADD COLUMN IF NOT EXISTS search_vector tsvector;
				
				-- Create GIN index for full-text search
				CREATE INDEX IF NOT EXISTS idx_beverages_search_vector 
				ON beverages USING GIN (search_vector);
				
				-- Create trigger to update search vector
				DROP TRIGGER IF EXISTS trigger_update_beverage_search_vector ON beverages;
				CREATE TRIGGER trigger_update_beverage_search_vector
					BEFORE INSERT OR UPDATE ON beverages
					FOR EACH ROW EXECUTE FUNCTION update_beverage_search_vector();
				
				-- Update existing records
				UPDATE beverages SET search_vector = to_tsvector('english', 
					name || ' ' || 
					COALESCE(description, '') || ' ' || 
					theme || ' ' ||
					COALESCE(ingredients::text, '')
				);
			`,
			DownSQL: `
				DROP TRIGGER IF EXISTS trigger_update_beverage_search_vector ON beverages;
				DROP FUNCTION IF EXISTS update_beverage_search_vector();
				ALTER TABLE beverages DROP COLUMN IF EXISTS search_vector;
			`,
		},
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations
}

// Migrate runs all pending migrations
func (mm *MigrationManager) Migrate(ctx context.Context) error {
	ctx, span := mm.tracing.StartDatabaseSpan(ctx, "MIGRATE", "schema_migrations")
	defer span.End()

	mm.logger.InfoContext(ctx, "Starting database migration")

	// Get all migrations
	migrations := mm.GetMigrations()

	// Get applied migrations
	appliedVersions, err := mm.getAppliedMigrations(ctx)
	if err != nil {
		mm.tracing.RecordError(span, err, "Failed to get applied migrations")
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Find pending migrations
	var pendingMigrations []Migration
	for _, migration := range migrations {
		if !contains(appliedVersions, migration.Version) {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	if len(pendingMigrations) == 0 {
		mm.logger.InfoContext(ctx, "No pending migrations")
		mm.tracing.RecordSuccess(span, "No pending migrations")
		return nil
	}

	mm.logger.InfoContext(ctx, "Found pending migrations",
		"count", len(pendingMigrations))

	// Apply pending migrations
	for _, migration := range pendingMigrations {
		if err := mm.applyMigration(ctx, migration); err != nil {
			mm.tracing.RecordError(span, err, "Failed to apply migration")
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	mm.tracing.RecordSuccess(span, "Database migration completed")
	mm.logger.InfoContext(ctx, "Database migration completed",
		"applied_count", len(pendingMigrations))

	return nil
}

// Rollback rolls back the last migration
func (mm *MigrationManager) Rollback(ctx context.Context) error {
	ctx, span := mm.tracing.StartDatabaseSpan(ctx, "ROLLBACK", "schema_migrations")
	defer span.End()

	mm.logger.InfoContext(ctx, "Starting migration rollback")

	// Get the last applied migration
	lastVersion, err := mm.getLastAppliedMigration(ctx)
	if err != nil {
		mm.tracing.RecordError(span, err, "Failed to get last applied migration")
		return fmt.Errorf("failed to get last applied migration: %w", err)
	}

	if lastVersion == 0 {
		mm.logger.InfoContext(ctx, "No migrations to rollback")
		mm.tracing.RecordSuccess(span, "No migrations to rollback")
		return nil
	}

	// Find the migration to rollback
	migrations := mm.GetMigrations()
	var migrationToRollback *Migration
	for _, migration := range migrations {
		if migration.Version == lastVersion {
			migrationToRollback = &migration
			break
		}
	}

	if migrationToRollback == nil {
		err := fmt.Errorf("migration version %d not found", lastVersion)
		mm.tracing.RecordError(span, err, "Migration not found")
		return err
	}

	// Apply rollback
	if err := mm.rollbackMigration(ctx, *migrationToRollback); err != nil {
		mm.tracing.RecordError(span, err, "Failed to rollback migration")
		return fmt.Errorf("failed to rollback migration %d: %w", lastVersion, err)
	}

	mm.tracing.RecordSuccess(span, "Migration rollback completed")
	mm.logger.InfoContext(ctx, "Migration rollback completed",
		"version", lastVersion,
		"name", migrationToRollback.Name)

	return nil
}

// applyMigration applies a single migration
func (mm *MigrationManager) applyMigration(ctx context.Context, migration Migration) error {
	ctx, span := mm.tracing.StartDatabaseSpan(ctx, "APPLY_MIGRATION", "schema_migrations")
	defer span.End()

	start := time.Now()

	mm.logger.InfoContext(ctx, "Applying migration",
		"version", migration.Version,
		"name", migration.Name,
		"description", migration.Description)

	// Start transaction
	tx, err := mm.db.BeginTx(ctx, nil)
	if err != nil {
		mm.tracing.RecordError(span, err, "Failed to start transaction")
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Execute migration SQL
	_, err = tx.ExecContext(ctx, migration.UpSQL)
	if err != nil {
		mm.tracing.RecordError(span, err, "Failed to execute migration SQL")
		mm.logger.ErrorContext(ctx, "Failed to execute migration SQL", err,
			"version", migration.Version,
			"name", migration.Name)
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration in schema_migrations table
	_, err = tx.ExecContext(ctx, `
		INSERT INTO schema_migrations (version, name, applied_at, checksum) 
		VALUES ($1, $2, $3, $4)`,
		migration.Version, migration.Name, time.Now(), mm.calculateChecksum(migration.UpSQL))
	if err != nil {
		mm.tracing.RecordError(span, err, "Failed to record migration")
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		mm.tracing.RecordError(span, err, "Failed to commit migration transaction")
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	duration := time.Since(start)
	mm.tracing.RecordSuccess(span, "Migration applied successfully")
	mm.logger.InfoContext(ctx, "Migration applied successfully",
		"version", migration.Version,
		"name", migration.Name,
		"duration_ms", duration.Milliseconds())

	return nil
}

// rollbackMigration rolls back a single migration
func (mm *MigrationManager) rollbackMigration(ctx context.Context, migration Migration) error {
	ctx, span := mm.tracing.StartDatabaseSpan(ctx, "ROLLBACK_MIGRATION", "schema_migrations")
	defer span.End()

	start := time.Now()

	mm.logger.InfoContext(ctx, "Rolling back migration",
		"version", migration.Version,
		"name", migration.Name)

	// Start transaction
	tx, err := mm.db.BeginTx(ctx, nil)
	if err != nil {
		mm.tracing.RecordError(span, err, "Failed to start transaction")
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Execute rollback SQL
	_, err = tx.ExecContext(ctx, migration.DownSQL)
	if err != nil {
		mm.tracing.RecordError(span, err, "Failed to execute rollback SQL")
		mm.logger.ErrorContext(ctx, "Failed to execute rollback SQL", err,
			"version", migration.Version,
			"name", migration.Name)
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove migration record
	_, err = tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = $1`, migration.Version)
	if err != nil {
		mm.tracing.RecordError(span, err, "Failed to remove migration record")
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		mm.tracing.RecordError(span, err, "Failed to commit rollback transaction")
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	duration := time.Since(start)
	mm.tracing.RecordSuccess(span, "Migration rolled back successfully")
	mm.logger.InfoContext(ctx, "Migration rolled back successfully",
		"version", migration.Version,
		"name", migration.Name,
		"duration_ms", duration.Milliseconds())

	return nil
}

// getAppliedMigrations returns a list of applied migration versions
func (mm *MigrationManager) getAppliedMigrations(ctx context.Context) ([]int, error) {
	// First, ensure the schema_migrations table exists
	_, err := mm.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			checksum VARCHAR(64)
		)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	rows, err := mm.db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// getLastAppliedMigration returns the version of the last applied migration
func (mm *MigrationManager) getLastAppliedMigration(ctx context.Context) (int, error) {
	row := mm.db.QueryRowContext(ctx, `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`)

	var version int
	err := row.Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // No migrations applied
		}
		return 0, fmt.Errorf("failed to get last applied migration: %w", err)
	}

	return version, nil
}

// calculateChecksum calculates a simple checksum for migration SQL
func (mm *MigrationManager) calculateChecksum(sql string) string {
	// Simple checksum - in production, use a proper hash function
	return fmt.Sprintf("%x", len(sql))
}

// Helper function to check if slice contains value
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
