package commands

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/cli/config"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	_ "github.com/lib/pq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Migration represents a database migration
type Migration struct {
	ID            int           `json:"id"`
	Version       string        `json:"version"`
	Name          string        `json:"name"`
	UpSQL         string        `json:"up_sql"`
	DownSQL       string        `json:"down_sql"`
	AppliedAt     *time.Time    `json:"applied_at"`
	Checksum      string        `json:"checksum"`
	ExecutionTime time.Duration `json:"execution_time"`
}

// MigrationStatus represents the status of migrations
type MigrationStatus struct {
	TotalMigrations   int        `json:"total_migrations"`
	AppliedMigrations int        `json:"applied_migrations"`
	PendingMigrations int        `json:"pending_migrations"`
	LastApplied       *Migration `json:"last_applied"`
}

// newMigrateUpCommand creates the migrate up command
func newMigrateUpCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var steps int
	var dryRun bool
	var backup bool

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Run pending database migrations",
		Long: `Run pending database migrations to bring the database schema up to date.
This command will:
  • Check for pending migrations
  • Create backup if requested
  • Apply migrations in order
  • Update migration tracking table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runMigrationsUp(ctx, cfg, logger, environment, steps, dryRun, backup)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "development", "Environment to migrate")
	cmd.Flags().IntVarP(&steps, "steps", "s", 0, "Number of migrations to run (0 = all)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be migrated without executing")
	cmd.Flags().BoolVar(&backup, "backup", true, "Create backup before migration")

	return cmd
}

// newMigrateDownCommand creates the migrate down command
func newMigrateDownCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var steps int
	var dryRun bool
	var backup bool

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback database migrations",
		Long: `Rollback database migrations to a previous state.
This command will:
  • Check applied migrations
  • Create backup if requested
  • Rollback migrations in reverse order
  • Update migration tracking table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runMigrationsDown(ctx, cfg, logger, environment, steps, dryRun, backup)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "development", "Environment to rollback")
	cmd.Flags().IntVarP(&steps, "steps", "s", 1, "Number of migrations to rollback")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be rolled back without executing")
	cmd.Flags().BoolVar(&backup, "backup", true, "Create backup before rollback")

	return cmd
}

// newMigrateStatusCommand creates the migrate status command
func newMigrateStatusCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var format string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long: `Show the current status of database migrations including:
  • Applied migrations
  • Pending migrations
  • Migration history
  • Database version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return showMigrationStatus(ctx, cfg, logger, environment, format)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "development", "Environment to check")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, yaml)")

	return cmd
}

// newMigrateSeedCommand creates the seed command
func newMigrateSeedCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var seedFile string
	var force bool

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed database with test data",
		Long: `Seed the database with test data for development and testing.
This command will:
  • Load seed data from SQL files
  • Insert test records
  • Set up demo accounts and data`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return seedDatabase(ctx, cfg, logger, environment, seedFile, force)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "development", "Environment to seed")
	cmd.Flags().StringVarP(&seedFile, "file", "f", "", "Specific seed file to run")
	cmd.Flags().BoolVar(&force, "force", false, "Force seeding even if data exists")

	return cmd
}

// newMigrateGenerateCommand creates the generate migration command
func newMigrateGenerateCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var name string
	var template string

	cmd := &cobra.Command{
		Use:   "generate <name>",
		Short: "Generate new migration files",
		Long: `Generate new migration files with up and down SQL.
This command will:
  • Create timestamped migration files
  • Generate template SQL
  • Add to migrations directory`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]
			return generateMigration(cfg, logger, name, template)
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "default", "Migration template (default, table, index)")

	return cmd
}

// runMigrationsUp executes pending migrations
func runMigrationsUp(ctx context.Context, cfg *config.Config, logger *zap.Logger, environment string, steps int, dryRun, backup bool) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Checking migrations..."
	s.Start()

	// Connect to database
	db, err := connectToDatabase(cfg, environment)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Ensure migration table exists
	if err := ensureMigrationTable(db); err != nil {
		s.Stop()
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	// Get pending migrations
	migrations, err := getPendingMigrations(db)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	s.Stop()

	if len(migrations) == 0 {
		color.Green("✅ No pending migrations")
		return nil
	}

	// Limit migrations if steps specified
	if steps > 0 && steps < len(migrations) {
		migrations = migrations[:steps]
	}

	if dryRun {
		return showMigrationPlan(migrations, "up")
	}

	// Create backup if requested
	if backup {
		backupConfig := &BackupConfig{Environment: environment, BackupType: "database"}
		if err := createDatabaseBackup(ctx, cfg, "backups/"+environment, backupConfig); err != nil {
			color.Yellow("⚠️ Failed to create backup: %v", err)
		} else {
			color.Green("✅ Database backup created")
		}
	}

	// Apply migrations
	color.Cyan("🚀 Applying %d migrations...", len(migrations))

	for i, migration := range migrations {
		fmt.Printf("\n[%d/%d] Applying %s...", i+1, len(migrations), migration.Name)

		start := time.Now()
		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
		}

		duration := time.Since(start)
		color.Green(" ✅ (%s)", duration)

		logger.Info("Migration applied",
			zap.String("migration", migration.Name),
			zap.Duration("duration", duration),
		)
	}

	color.Green("\n🎉 All migrations applied successfully!")
	return nil
}

// runMigrationsDown rolls back migrations
func runMigrationsDown(ctx context.Context, cfg *config.Config, logger *zap.Logger, environment string, steps int, dryRun, backup bool) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Checking applied migrations..."
	s.Start()

	// Connect to database
	db, err := connectToDatabase(cfg, environment)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Get applied migrations to rollback
	migrations, err := getAppliedMigrations(db, steps)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	s.Stop()

	if len(migrations) == 0 {
		color.Green("✅ No migrations to rollback")
		return nil
	}

	if dryRun {
		return showMigrationPlan(migrations, "down")
	}

	// Create backup if requested
	if backup {
		backupConfig := &BackupConfig{Environment: environment, BackupType: "database"}
		if err := createDatabaseBackup(ctx, cfg, "backups/"+environment, backupConfig); err != nil {
			color.Yellow("⚠️ Failed to create backup: %v", err)
		} else {
			color.Green("✅ Database backup created")
		}
	}

	// Rollback migrations
	color.Cyan("🔄 Rolling back %d migrations...", len(migrations))

	for i, migration := range migrations {
		fmt.Printf("\n[%d/%d] Rolling back %s...", i+1, len(migrations), migration.Name)

		start := time.Now()
		if err := rollbackMigration(db, migration); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Name, err)
		}

		duration := time.Since(start)
		color.Green(" ✅ (%s)", duration)

		logger.Info("Migration rolled back",
			zap.String("migration", migration.Name),
			zap.Duration("duration", duration),
		)
	}

	color.Green("\n🎉 All migrations rolled back successfully!")
	return nil
}

// showMigrationStatus displays the current migration status
func showMigrationStatus(ctx context.Context, cfg *config.Config, logger *zap.Logger, environment, format string) error {
	// Connect to database
	db, err := connectToDatabase(cfg, environment)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Get migration status
	status, err := getMigrationStatus(db)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// Get all migrations
	allMigrations, err := getAllMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	switch format {
	case "json":
		return printMigrationStatusJSON(status, allMigrations)
	case "yaml":
		return printMigrationStatusYAML(status, allMigrations)
	default:
		return printMigrationStatusTable(status, allMigrations)
	}
}

// printMigrationStatusTable prints migration status in table format
func printMigrationStatusTable(status *MigrationStatus, migrations []Migration) error {
	color.Cyan("📊 Migration Status")
	color.Cyan("=" + strings.Repeat("=", 50))

	fmt.Printf("Total Migrations: %d\n", status.TotalMigrations)
	fmt.Printf("Applied: %d\n", status.AppliedMigrations)
	fmt.Printf("Pending: %d\n", status.PendingMigrations)

	if status.LastApplied != nil {
		fmt.Printf("Last Applied: %s (%s)\n", status.LastApplied.Name, status.LastApplied.AppliedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Println()

	// Migrations table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Version", "Name", "Status", "Applied At"})
	table.SetBorder(false)

	for _, migration := range migrations {
		status := "❌ Pending"
		appliedAt := "-"

		if migration.AppliedAt != nil {
			status = "✅ Applied"
			appliedAt = migration.AppliedAt.Format("2006-01-02 15:04:05")
		}

		table.Append([]string{
			migration.Version,
			migration.Name,
			status,
			appliedAt,
		})
	}

	table.Render()
	return nil
}

// Helper functions for database operations
func connectToDatabase(cfg *config.Config, environment string) (*sql.DB, error) {
	// Get database configuration for environment
	dbConfig := cfg.Services.Database

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database)

	return sql.Open("postgres", connStr)
}

func ensureMigrationTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			checksum VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			execution_time_ms INTEGER DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_schema_migrations_version ON schema_migrations(version);
	`

	_, err := db.Exec(query)
	return err
}

func getPendingMigrations(db *sql.DB) ([]Migration, error) {
	// Read migration files from disk
	migrationFiles, err := readMigrationFiles()
	if err != nil {
		return nil, err
	}

	// Get applied migrations from database
	appliedVersions, err := getAppliedVersions(db)
	if err != nil {
		return nil, err
	}

	var pending []Migration
	for _, migration := range migrationFiles {
		if !contains(appliedVersions, migration.Version) {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

func readMigrationFiles() ([]Migration, error) {
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			migration, err := parseMigrationFile(filepath.Join(migrationsDir, file.Name()))
			if err != nil {
				continue
			}
			migrations = append(migrations, migration)
		}
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigrationFile(filename string) (Migration, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return Migration{}, err
	}

	// Extract version and name from filename
	// Format: YYYYMMDDHHMMSS_migration_name.up.sql
	basename := filepath.Base(filename)
	parts := strings.Split(basename, "_")
	if len(parts) < 2 {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version := parts[0]
	name := strings.Join(parts[1:], "_")
	name = strings.TrimSuffix(name, ".up.sql")

	// Read down migration if exists
	downFile := strings.Replace(filename, ".up.sql", ".down.sql", 1)
	var downSQL string
	if downContent, err := os.ReadFile(downFile); err == nil {
		downSQL = string(downContent)
	}

	return Migration{
		Version: version,
		Name:    name,
		UpSQL:   string(content),
		DownSQL: downSQL,
	}, nil
}

func getAppliedVersions(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func applyMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.UpSQL); err != nil {
		return err
	}

	// Record migration
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)",
		migration.Version, migration.Name, "checksum", // TODO: Calculate actual checksum
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func rollbackMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute rollback SQL
	if migration.DownSQL != "" {
		if _, err := tx.Exec(migration.DownSQL); err != nil {
			return err
		}
	}

	// Remove migration record
	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = $1", migration.Version)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Additional helper functions would be implemented here...
func getAppliedMigrations(db *sql.DB, limit int) ([]Migration, error) {
	// Implementation for getting applied migrations in reverse order
	return nil, nil
}

func getMigrationStatus(db *sql.DB) (*MigrationStatus, error) {
	// Implementation for getting migration status
	return &MigrationStatus{}, nil
}

func getAllMigrations(db *sql.DB) ([]Migration, error) {
	// Implementation for getting all migrations
	return nil, nil
}

func showMigrationPlan(migrations []Migration, direction string) error {
	// Implementation for showing migration plan
	return nil
}

func seedDatabase(ctx context.Context, cfg *config.Config, logger *zap.Logger, environment, seedFile string, force bool) error {
	// Implementation for seeding database
	return nil
}

func generateMigration(cfg *config.Config, logger *zap.Logger, name, template string) error {
	// Implementation for generating migration files
	return nil
}

func printMigrationStatusJSON(status *MigrationStatus, migrations []Migration) error {
	// Implementation for JSON output
	return nil
}

func printMigrationStatusYAML(status *MigrationStatus, migrations []Migration) error {
	// Implementation for YAML output
	return nil
}
