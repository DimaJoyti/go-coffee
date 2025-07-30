package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/cli/config"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewDevOpsCommand creates the main DevOps command
func NewDevOpsCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var devopsCmd = &cobra.Command{
		Use:   "devops",
		Short: "DevOps operations for Go Coffee platform",
		Long: color.New(color.FgCyan, color.Bold).Sprint(`
╔═══════════════════════════════════════════════════════════════╗
║                    🚀 DevOps Operations                      ║
║           Comprehensive Platform Management                   ║
╚═══════════════════════════════════════════════════════════════╝

DevOps operations for the Go Coffee platform including:
  • 🏗️  Project initialization and setup
  • 🚀 Service deployment and orchestration
  • 📊 Health monitoring and status checks
  • 📝 Log aggregation and viewing
  • 🗄️  Database migrations and operations
  • ⚙️  Configuration management and validation
  • 💾 Backup and restore operations
  • 🧪 Testing and quality assurance
`),
		Aliases: []string{"ops", "deploy"},
	}

	// Add subcommands
	devopsCmd.AddCommand(newInitCommand(cfg, logger))
	devopsCmd.AddCommand(newDeployCommand(cfg, logger))
	devopsCmd.AddCommand(newStatusCommand(cfg, logger))
	devopsCmd.AddCommand(newLogsCommand(cfg, logger))
	devopsCmd.AddCommand(newMigrateCommand(cfg, logger))
	devopsCmd.AddCommand(newBackupCommand(cfg, logger))
	devopsCmd.AddCommand(newRestoreCommand(cfg, logger))
	devopsCmd.AddCommand(newTestCommand(cfg, logger))
	devopsCmd.AddCommand(newEnvCommand(cfg, logger))
	devopsCmd.AddCommand(newHealthCommand(cfg, logger))

	return devopsCmd
}

// newInitCommand creates the project initialization command
func newInitCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var force bool
	var template string

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize new Go Coffee environment",
		Long: `Initialize a new Go Coffee environment with:
  • Environment-specific configuration files
  • Docker compose setup
  • Kubernetes manifests
  • Database initialization scripts
  • Monitoring and observability setup`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			projectName := "go-coffee"
			if len(args) > 0 {
				projectName = args[0]
			}

			return initializeProject(ctx, cfg, logger, projectName, environment, template, force)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "development", "Environment to initialize (development, staging, production)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force initialization even if files exist")
	cmd.Flags().StringVarP(&template, "template", "t", "standard", "Project template (standard, minimal, enterprise)")

	return cmd
}

// newDeployCommand creates the deployment command
func newDeployCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var services []string
	var dryRun bool
	var parallel bool
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "deploy [service...]",
		Short: "Deploy services to target environment",
		Long: `Deploy one or more services to the target environment:
  • Build and push container images
  • Update Kubernetes deployments
  • Run database migrations
  • Perform health checks
  • Rollback on failure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) > 0 {
				services = args
			}

			return deployServices(ctx, cfg, logger, services, environment, dryRun, parallel, timeout)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "development", "Target environment")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deployed without executing")
	cmd.Flags().BoolVarP(&parallel, "parallel", "p", false, "Deploy services in parallel")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "Deployment timeout")

	return cmd
}

// newStatusCommand creates the status command
func newStatusCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var environment string
	var format string
	var watch bool
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show system health and service status",
		Long: `Display comprehensive system status including:
  • Service health and availability
  • Resource utilization
  • Database connectivity
  • External dependencies
  • Recent deployments and changes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if watch {
				return watchSystemStatus(ctx, cfg, logger, environment, format, interval)
			}

			return showSystemStatus(ctx, cfg, logger, environment, format)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "", "Environment to check (default: current)")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, yaml)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch status continuously")
	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "Watch interval")

	return cmd
}

// newLogsCommand creates the logs command
func newLogsCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var service string
	var environment string
	var follow bool
	var tail int
	var since string
	var grep string

	cmd := &cobra.Command{
		Use:   "logs [service]",
		Short: "View and aggregate service logs",
		Long: `View logs from services with advanced filtering:
  • Real-time log streaming
  • Multi-service log aggregation
  • Pattern matching and filtering
  • Structured log parsing
  • Export to various formats`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) > 0 {
				service = args[0]
			}

			return viewLogs(ctx, cfg, logger, service, environment, follow, tail, since, grep)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "", "Environment to get logs from")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().IntVar(&tail, "tail", 100, "Number of lines to show from the end")
	cmd.Flags().StringVar(&since, "since", "", "Show logs since timestamp (e.g., 2h, 1d)")
	cmd.Flags().StringVar(&grep, "grep", "", "Filter logs by pattern")

	return cmd
}

// newMigrateCommand creates the database migration command
func newMigrateCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Database migration operations",
		Long: `Manage database migrations and schema changes:
  • Run pending migrations
  • Rollback migrations
  • Generate migration files
  • Seed test data
  • Backup before migrations`,
		Aliases: []string{"db", "migration"},
	}

	migrateCmd.AddCommand(newMigrateUpCommand(cfg, logger))
	migrateCmd.AddCommand(newMigrateDownCommand(cfg, logger))
	migrateCmd.AddCommand(newMigrateStatusCommand(cfg, logger))
	migrateCmd.AddCommand(newMigrateSeedCommand(cfg, logger))
	migrateCmd.AddCommand(newMigrateGenerateCommand(cfg, logger))

	return migrateCmd
}

// Implementation functions will be added in separate files for better organization

// initializeProject initializes a new project environment
func initializeProject(ctx context.Context, cfg *config.Config, logger *zap.Logger, projectName, environment, template string, force bool) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Initializing %s environment for %s...", environment, projectName)
	s.Start()
	defer s.Stop()

	logger.Info("Starting project initialization",
		zap.String("project", projectName),
		zap.String("environment", environment),
		zap.String("template", template),
		zap.Bool("force", force),
	)

	// Create project directory structure
	if err := createProjectStructure(projectName, environment, template, force); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	// Generate configuration files
	if err := generateConfigFiles(projectName, environment, template); err != nil {
		return fmt.Errorf("failed to generate config files: %w", err)
	}

	// Initialize Docker setup
	if err := initializeDockerSetup(projectName, environment); err != nil {
		return fmt.Errorf("failed to initialize Docker setup: %w", err)
	}

	// Initialize Kubernetes manifests
	if err := initializeKubernetesManifests(projectName, environment); err != nil {
		return fmt.Errorf("failed to initialize Kubernetes manifests: %w", err)
	}

	s.Stop()

	color.Green("✅ Successfully initialized %s environment for %s", environment, projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Review configuration files in ./%s/\n", projectName)
	fmt.Printf("  2. Run: coffee deploy --env %s\n", environment)
	fmt.Printf("  3. Check status: coffee status --env %s\n", environment)

	return nil
}

// createProjectStructure creates the basic project directory structure
func createProjectStructure(projectName, environment, template string, force bool) error {
	dirs := []string{
		filepath.Join(projectName, "configs", environment),
		filepath.Join(projectName, "docker"),
		filepath.Join(projectName, "k8s", environment),
		filepath.Join(projectName, "scripts"),
		filepath.Join(projectName, "migrations"),
		filepath.Join(projectName, "monitoring"),
		filepath.Join(projectName, "logs"),
		filepath.Join(projectName, "backups"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateConfigFiles generates environment-specific configuration files
func generateConfigFiles(projectName, environment, template string) error {
	// This will be implemented to generate actual config files
	// For now, create placeholder files
	configDir := filepath.Join(projectName, "configs", environment)

	files := map[string]string{
		"app.yaml":        generateAppConfig(environment),
		"database.yaml":   generateDatabaseConfig(environment),
		"monitoring.yaml": generateMonitoringConfig(environment),
		"secrets.yaml":    generateSecretsConfig(environment),
	}

	for filename, content := range files {
		filePath := filepath.Join(configDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write config file %s: %w", filePath, err)
		}
	}

	return nil
}

// Helper functions for generating config content
func generateAppConfig(environment string) string {
	return fmt.Sprintf(`# Go Coffee Application Configuration - %s
app:
  name: go-coffee
  environment: %s
  port: 8080
  log_level: info
  
server:
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s
  
cors:
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowed_headers: ["*"]
`, environment, environment)
}

func generateDatabaseConfig(environment string) string {
	return fmt.Sprintf(`# Database Configuration - %s
database:
  postgres:
    host: localhost
    port: 5432
    database: go_coffee_%s
    username: go_coffee_user
    password: go_coffee_password
    ssl_mode: disable
    max_open_conns: 25
    max_idle_conns: 5
    conn_max_lifetime: 5m
    
  redis:
    host: localhost
    port: 6379
    password: ""
    database: 0
    max_retries: 3
    pool_size: 10
`, environment, environment)
}

func generateMonitoringConfig(environment string) string {
	return fmt.Sprintf(`# Monitoring Configuration - %s
monitoring:
  prometheus:
    enabled: true
    port: 9090
    path: /metrics
    
  jaeger:
    enabled: true
    endpoint: http://localhost:14268/api/traces
    
  logging:
    level: info
    format: json
    output: stdout
`, environment)
}

func generateSecretsConfig(environment string) string {
	return fmt.Sprintf(`# Secrets Configuration - %s
# Note: This file should be encrypted or stored in a secure secret management system
secrets:
  jwt_secret: "your-jwt-secret-here"
  database_password: "your-db-password-here"
  redis_password: "your-redis-password-here"
  api_keys:
    external_service: "your-api-key-here"
`, environment)
}

// initializeDockerSetup initializes Docker configuration
func initializeDockerSetup(projectName, environment string) error {
	dockerDir := filepath.Join(projectName, "docker")

	// Create docker-compose file for the environment
	composeFile := filepath.Join(dockerDir, fmt.Sprintf("docker-compose.%s.yml", environment))
	composeContent := generateDockerCompose(environment)

	if err := os.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose file: %w", err)
	}

	// Create Dockerfile templates
	dockerfileContent := generateDockerfile()
	dockerfilePath := filepath.Join(dockerDir, "Dockerfile")

	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	return nil
}

// initializeKubernetesManifests initializes Kubernetes manifests
func initializeKubernetesManifests(projectName, environment string) error {
	k8sDir := filepath.Join(projectName, "k8s", environment)

	manifests := map[string]string{
		"namespace.yaml":  generateNamespaceManifest(projectName, environment),
		"configmap.yaml":  generateConfigMapManifest(projectName, environment),
		"deployment.yaml": generateDeploymentManifest(projectName, environment),
		"service.yaml":    generateServiceManifest(projectName, environment),
		"ingress.yaml":    generateIngressManifest(projectName, environment),
	}

	for filename, content := range manifests {
		filePath := filepath.Join(k8sDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write manifest %s: %w", filename, err)
		}
	}

	return nil
}

// generateDockerCompose generates docker-compose content
func generateDockerCompose(environment string) string {
	return fmt.Sprintf(`version: '3.8'

services:
  # Go Coffee Application
  app:
    build: .
    container_name: go-coffee-app-%s
    ports:
      - "8080:8080"
    environment:
      - ENV=%s
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    networks:
      - go-coffee-network
    restart: unless-stopped

  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: go-coffee-postgres-%s
    environment:
      POSTGRES_DB: go_coffee_%s
      POSTGRES_USER: go_coffee_user
      POSTGRES_PASSWORD: go_coffee_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data_%s:/var/lib/postgresql/data
    networks:
      - go-coffee-network
    restart: unless-stopped

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: go-coffee-redis-%s
    ports:
      - "6379:6379"
    volumes:
      - redis_data_%s:/data
    networks:
      - go-coffee-network
    restart: unless-stopped

volumes:
  postgres_data_%s:
  redis_data_%s:

networks:
  go-coffee-network:
    driver: bridge
`, environment, environment, environment, environment, environment, environment, environment, environment, environment)
}

// generateDockerfile generates Dockerfile content
func generateDockerfile() string {
	return `# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api-gateway

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./main"]
`
}
