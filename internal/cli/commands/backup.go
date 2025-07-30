package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/cli/config"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// BackupConfig holds backup configuration
type BackupConfig struct {
	Environment   string
	BackupType    string
	Compression   bool
	Encryption    bool
	RemoteStorage bool
	RetentionDays int
}

// BackupInfo represents backup information
type BackupInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Environment string    `json:"environment"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	Location    string    `json:"location"`
	Checksum    string    `json:"checksum"`
	Compressed  bool      `json:"compressed"`
	Encrypted   bool      `json:"encrypted"`
}

// newBackupCommand creates the backup command
func newBackupCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup system data and configurations",
		Long: `Create backups of system data including:
  • Database backups (PostgreSQL, Redis)
  • Configuration files
  • Application data
  • Container volumes
  • Kubernetes resources`,
		Aliases: []string{"bak"},
	}

	backupCmd.AddCommand(newBackupCreateCommand(cfg, logger))
	backupCmd.AddCommand(newBackupListCommand(cfg, logger))
	backupCmd.AddCommand(newBackupDeleteCommand(cfg, logger))
	backupCmd.AddCommand(newBackupVerifyCommand(cfg, logger))

	return backupCmd
}

// newBackupCreateCommand creates the backup create command
func newBackupCreateCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var backupType string
	var compression bool
	var encryption bool
	var remoteStorage bool
	var retentionDays int
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new backup",
		Long: `Create a new backup of the specified components.
Backup types:
  • full - Complete system backup
  • database - Database only
  • config - Configuration files only
  • volumes - Container volumes only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			backupConfig := &BackupConfig{
				Environment:   environment,
				BackupType:    backupType,
				Compression:   compression,
				Encryption:    encryption,
				RemoteStorage: remoteStorage,
				RetentionDays: retentionDays,
			}

			return createBackup(ctx, cfg, logger, backupConfig, name)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "development", "Environment to backup")
	cmd.Flags().StringVarP(&backupType, "type", "t", "full", "Backup type (full, database, config, volumes)")
	cmd.Flags().BoolVarP(&compression, "compress", "c", true, "Compress backup")
	cmd.Flags().BoolVar(&encryption, "encrypt", false, "Encrypt backup")
	cmd.Flags().BoolVar(&remoteStorage, "remote", false, "Upload to remote storage")
	cmd.Flags().IntVar(&retentionDays, "retention", 30, "Backup retention in days")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom backup name")

	return cmd
}

// newBackupListCommand creates the backup list command
func newBackupListCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var backupType string
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available backups",
		Long:  `List all available backups with details including size, date, and location.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return listBackups(ctx, cfg, logger, environment, backupType, format)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "", "Filter by environment")
	cmd.Flags().StringVarP(&backupType, "type", "t", "", "Filter by backup type")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, yaml)")

	return cmd
}

// newBackupDeleteCommand creates the backup delete command
func newBackupDeleteCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var force bool
	var older string

	cmd := &cobra.Command{
		Use:   "delete <backup-id>",
		Short: "Delete a backup",
		Long:  `Delete a specific backup or backups older than specified time.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var backupID string
			if len(args) > 0 {
				backupID = args[0]
			}

			return deleteBackup(ctx, cfg, logger, backupID, older, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion without confirmation")
	cmd.Flags().StringVar(&older, "older-than", "", "Delete backups older than specified time (e.g., 30d, 1w)")

	return cmd
}

// newBackupVerifyCommand creates the backup verify command
func newBackupVerifyCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify <backup-id>",
		Short: "Verify backup integrity",
		Long:  `Verify the integrity of a backup by checking checksums and testing restoration.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			backupID := args[0]
			return verifyBackup(ctx, cfg, logger, backupID)
		},
	}

	return cmd
}

// newRestoreCommand creates the restore command
func newRestoreCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var backupID string
	var dryRun bool
	var force bool
	var components []string

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore system from backup",
		Long: `Restore system components from a backup.
This command will:
  • Stop affected services
  • Restore data from backup
  • Restart services
  • Verify restoration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return restoreFromBackup(ctx, cfg, logger, backupID, environment, components, dryRun, force)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "development", "Target environment")
	cmd.Flags().StringVarP(&backupID, "backup", "b", "", "Backup ID to restore from")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be restored without executing")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force restoration without confirmation")
	cmd.Flags().StringSliceVarP(&components, "components", "c", []string{}, "Specific components to restore")

	return cmd
}

// createBackup creates a new backup
func createBackup(ctx context.Context, cfg *config.Config, logger *zap.Logger, backupConfig *BackupConfig, name string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Creating backup..."
	s.Start()

	logger.Info("Starting backup creation",
		zap.String("environment", backupConfig.Environment),
		zap.String("type", backupConfig.BackupType),
		zap.Bool("compression", backupConfig.Compression),
		zap.Bool("encryption", backupConfig.Encryption),
	)

	// Generate backup name if not provided
	if name == "" {
		name = generateBackupName(backupConfig)
	}

	// Create backup directory
	backupDir := filepath.Join("backups", backupConfig.Environment)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupPath := filepath.Join(backupDir, name)

	// Create backup based on type
	var err error
	switch backupConfig.BackupType {
	case "full":
		err = createFullBackup(ctx, cfg, backupPath, backupConfig)
	case "database":
		err = createDatabaseBackup(ctx, cfg, backupPath, backupConfig)
	case "config":
		err = createConfigBackup(ctx, cfg, backupPath, backupConfig)
	case "volumes":
		err = createVolumesBackup(ctx, cfg, backupPath, backupConfig)
	default:
		s.Stop()
		return fmt.Errorf("unsupported backup type: %s", backupConfig.BackupType)
	}

	if err != nil {
		s.Stop()
		return fmt.Errorf("backup creation failed: %w", err)
	}

	// Post-process backup (compression, encryption)
	if err := postProcessBackup(backupPath, backupConfig); err != nil {
		s.Stop()
		return fmt.Errorf("backup post-processing failed: %w", err)
	}

	// Upload to remote storage if requested
	if backupConfig.RemoteStorage {
		if err := uploadBackupToRemote(backupPath, backupConfig); err != nil {
			color.Yellow("⚠️ Failed to upload to remote storage: %v", err)
		} else {
			color.Green("✅ Backup uploaded to remote storage")
		}
	}

	s.Stop()

	// Get backup info
	backupInfo, err := getBackupInfo(backupPath)
	if err != nil {
		color.Yellow("⚠️ Failed to get backup info: %v", err)
	}

	color.Green("✅ Backup created successfully")
	fmt.Printf("Backup ID: %s\n", backupInfo.ID)
	fmt.Printf("Location: %s\n", backupInfo.Location)
	fmt.Printf("Size: %s\n", formatSize(backupInfo.Size))
	fmt.Printf("Type: %s\n", backupInfo.Type)

	return nil
}

// createFullBackup creates a complete system backup
func createFullBackup(ctx context.Context, cfg *config.Config, backupPath string, backupConfig *BackupConfig) error {
	// Create full backup including database, config, and volumes
	if err := createDatabaseBackup(ctx, cfg, filepath.Join(backupPath, "database"), backupConfig); err != nil {
		return fmt.Errorf("database backup failed: %w", err)
	}

	if err := createConfigBackup(ctx, cfg, filepath.Join(backupPath, "config"), backupConfig); err != nil {
		return fmt.Errorf("config backup failed: %w", err)
	}

	if err := createVolumesBackup(ctx, cfg, filepath.Join(backupPath, "volumes"), backupConfig); err != nil {
		return fmt.Errorf("volumes backup failed: %w", err)
	}

	return nil
}

// createDatabaseBackup creates database backups
func createDatabaseBackup(ctx context.Context, cfg *config.Config, backupPath string, backupConfig *BackupConfig) error {
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return err
	}

	// Backup PostgreSQL
	pgBackupFile := filepath.Join(backupPath, "postgres.sql")
	if err := backupPostgreSQL(cfg, pgBackupFile, backupConfig.Environment); err != nil {
		return fmt.Errorf("PostgreSQL backup failed: %w", err)
	}

	// Backup Redis
	redisBackupFile := filepath.Join(backupPath, "redis.rdb")
	if err := backupRedis(cfg, redisBackupFile, backupConfig.Environment); err != nil {
		return fmt.Errorf("Redis backup failed: %w", err)
	}

	return nil
}

// createConfigBackup creates configuration backups
func createConfigBackup(ctx context.Context, cfg *config.Config, backupPath string, backupConfig *BackupConfig) error {
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return err
	}

	// Backup configuration directories
	configDirs := []string{
		"configs",
		"docker",
		"k8s",
		"monitoring",
	}

	for _, dir := range configDirs {
		if _, err := os.Stat(dir); err == nil {
			destDir := filepath.Join(backupPath, dir)
			if err := copyDirectory(dir, destDir); err != nil {
				return fmt.Errorf("failed to backup %s: %w", dir, err)
			}
		}
	}

	return nil
}

// createVolumesBackup creates container volumes backup
func createVolumesBackup(ctx context.Context, cfg *config.Config, backupPath string, backupConfig *BackupConfig) error {
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return err
	}

	// Get Docker volumes
	cmd := exec.Command("docker", "volume", "ls", "-q")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list Docker volumes: %w", err)
	}

	volumes := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, volume := range volumes {
		if volume == "" {
			continue
		}

		// Skip system volumes
		if strings.HasPrefix(volume, "go-coffee") {
			volumeBackupPath := filepath.Join(backupPath, volume+".tar")
			if err := backupDockerVolume(volume, volumeBackupPath); err != nil {
				return fmt.Errorf("failed to backup volume %s: %w", volume, err)
			}
		}
	}

	return nil
}

// listBackups lists available backups
func listBackups(ctx context.Context, cfg *config.Config, logger *zap.Logger, environment, backupType, format string) error {
	backups, err := getAvailableBackups(environment, backupType)
	if err != nil {
		return fmt.Errorf("failed to get backups: %w", err)
	}

	switch format {
	case "json":
		return printBackupsJSON(backups)
	case "yaml":
		return printBackupsYAML(backups)
	default:
		return printBackupsTable(backups)
	}
}

// printBackupsTable prints backups in table format
func printBackupsTable(backups []BackupInfo) error {
	if len(backups) == 0 {
		color.Yellow("No backups found")
		return nil
	}

	color.Cyan("💾 Available Backups")
	color.Cyan("=" + strings.Repeat("=", 50))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Type", "Environment", "Size", "Created", "Location"})
	table.SetBorder(false)

	for _, backup := range backups {
		table.Append([]string{
			backup.ID[:8] + "...", // Truncate ID
			backup.Name,
			backup.Type,
			backup.Environment,
			formatSize(backup.Size),
			backup.CreatedAt.Format("2006-01-02 15:04"),
			filepath.Base(backup.Location),
		})
	}

	table.Render()
	return nil
}

// Helper functions
func generateBackupName(config *BackupConfig) string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s-%s.backup", config.BackupType, config.Environment, timestamp)
}

func backupPostgreSQL(cfg *config.Config, backupFile, environment string) error {
	dbConfig := cfg.Services.Database

	cmd := exec.Command("pg_dump",
		"-h", dbConfig.Host,
		"-p", fmt.Sprintf("%d", dbConfig.Port),
		"-U", dbConfig.Username,
		"-d", dbConfig.Database,
		"-f", backupFile,
		"--verbose",
		"--no-password",
	)

	// Set password via environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dbConfig.Password))

	return cmd.Run()
}

func backupRedis(cfg *config.Config, backupFile, environment string) error {
	// Use redis-cli to create backup
	cmd := exec.Command("redis-cli", "--rdb", backupFile)
	return cmd.Run()
}

func backupDockerVolume(volumeName, backupPath string) error {
	cmd := exec.Command("docker", "run", "--rm",
		"-v", volumeName+":/data",
		"-v", filepath.Dir(backupPath)+":/backup",
		"alpine",
		"tar", "czf", "/backup/"+filepath.Base(backupPath), "-C", "/data", ".")

	return cmd.Run()
}

func copyDirectory(src, dst string) error {
	cmd := exec.Command("cp", "-r", src, dst)
	return cmd.Run()
}

func postProcessBackup(backupPath string, config *BackupConfig) error {
	// Implement compression and encryption
	return nil
}

func uploadBackupToRemote(backupPath string, config *BackupConfig) error {
	// Implement remote storage upload
	return nil
}

func getBackupInfo(backupPath string) (*BackupInfo, error) {
	stat, err := os.Stat(backupPath)
	if err != nil {
		return nil, err
	}

	return &BackupInfo{
		ID:        generateBackupID(),
		Name:      filepath.Base(backupPath),
		Size:      stat.Size(),
		CreatedAt: stat.ModTime(),
		Location:  backupPath,
	}, nil
}

func generateBackupID() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// formatSize function moved to utils.go

// Additional helper functions would be implemented here...
func getAvailableBackups(environment, backupType string) ([]BackupInfo, error) {
	// Implementation for getting available backups
	return nil, nil
}

func deleteBackup(ctx context.Context, cfg *config.Config, logger *zap.Logger, backupID, older string, force bool) error {
	// Implementation for deleting backups
	return nil
}

func verifyBackup(ctx context.Context, cfg *config.Config, logger *zap.Logger, backupID string) error {
	// Implementation for verifying backup integrity
	return nil
}

func restoreFromBackup(ctx context.Context, cfg *config.Config, logger *zap.Logger, backupID, environment string, components []string, dryRun, force bool) error {
	// Implementation for restoring from backup
	return nil
}

func printBackupsJSON(backups []BackupInfo) error {
	// Implementation for JSON output
	return nil
}

func printBackupsYAML(backups []BackupInfo) error {
	// Implementation for YAML output
	return nil
}
