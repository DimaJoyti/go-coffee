package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/config"
	"github.com/DimaJoyti/go-coffee/internal/task-cli/repository"
	"github.com/DimaJoyti/go-coffee/internal/task-cli/service"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	cfg         *config.Config
	taskService service.TaskService
	redisClient *redis.Client

	// Version information
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "task-cli",
	Short: "A powerful CLI task manager with Redis backend",
	Long: `Task CLI is a feature-rich command-line task manager that uses Redis for storage.
It supports creating, updating, listing, and managing tasks with advanced features like
filtering, searching, statistics, and bulk operations.

Features:
  • Create and manage tasks with priorities, due dates, and assignments
  • Filter and search tasks with powerful query capabilities
  • View task statistics and analytics
  • Bulk operations for efficient task management
  • Export tasks in multiple formats (JSON, CSV, YAML)
  • Colorful and intuitive command-line interface

Examples:
  task-cli create "Fix login bug" --priority high --assignee john --due 2024-01-15
  task-cli list --status pending --assignee john
  task-cli update 123 --status completed
  task-cli search "bug"
  task-cli stats`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeApp()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		cleanup()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/task-cli/task-cli.yaml)")
	rootCmd.PersistentFlags().String("redis-url", "", "Redis connection URL (overrides config)")
	rootCmd.PersistentFlags().String("output", "", "Output format (table, json, yaml, csv)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")

	// Bind flags to viper
	viper.BindPFlag("redis.url", rootCmd.PersistentFlags().Lookup("redis-url"))
	viper.BindPFlag("cli.output_format", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("cli.color_output", rootCmd.PersistentFlags().Lookup("no-color"))

	// Add subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		fmt.Println("Run 'task-cli config init' to create a default configuration file")
		os.Exit(1)
	}
}

// initializeApp initializes the application dependencies
func initializeApp() error {
	// Initialize Redis client
	var err error
	redisClient, err = initRedisClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Redis client: %w", err)
	}

	// Test Redis connection
	ctx := context.Background()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Initialize repository and service
	taskRepo := repository.NewRedisTaskRepository(redisClient)
	taskService = service.NewTaskService(taskRepo, cfg)

	return nil
}

// initRedisClient initializes the Redis client
func initRedisClient() (*redis.Client, error) {
	// Check for Redis URL override
	if redisURL := viper.GetString("redis.url"); redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			return nil, fmt.Errorf("invalid Redis URL: %w", err)
		}
		return redis.NewClient(opt), nil
	}

	// Use configuration
	return redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}), nil
}

// cleanup performs cleanup operations
func cleanup() {
	if redisClient != nil {
		redisClient.Close()
	}
}

// getOutputFormat returns the output format from flags or config
func getOutputFormat() string {
	if format := viper.GetString("cli.output_format"); format != "" {
		return format
	}
	return cfg.CLI.OutputFormat
}

// isColorEnabled returns whether colored output is enabled
func isColorEnabled() bool {
	if viper.GetBool("no-color") {
		return false
	}
	return cfg.CLI.ColorOutput
}

// isVerboseEnabled returns whether verbose output is enabled
func isVerboseEnabled() bool {
	return viper.GetBool("verbose")
}

// printError prints an error message with optional color
func printError(msg string, args ...interface{}) {
	if isColorEnabled() {
		fmt.Printf("\033[31mError: "+msg+"\033[0m\n", args...)
	} else {
		fmt.Printf("Error: "+msg+"\n", args...)
	}
}

// printSuccess prints a success message with optional color
func printSuccess(msg string, args ...interface{}) {
	if isColorEnabled() {
		fmt.Printf("\033[32m"+msg+"\033[0m\n", args...)
	} else {
		fmt.Printf(msg+"\n", args...)
	}
}

// printWarning prints a warning message with optional color
func printWarning(msg string, args ...interface{}) {
	if isColorEnabled() {
		fmt.Printf("\033[33mWarning: "+msg+"\033[0m\n", args...)
	} else {
		fmt.Printf("Warning: "+msg+"\n", args...)
	}
}

// printInfo prints an info message with optional color
func printInfo(msg string, args ...interface{}) {
	if isColorEnabled() {
		fmt.Printf("\033[36m"+msg+"\033[0m\n", args...)
	} else {
		fmt.Printf(msg+"\n", args...)
	}
}

// SetVersionInfo sets version information from main
func SetVersionInfo(v, bt, gc string) {
	version = v
	buildTime = bt
	gitCommit = gc
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of task-cli",
	Long:  "Print the version number and build information of task-cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Task CLI v%s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		fmt.Println("Built with Go and Redis")
		fmt.Println("https://github.com/DimaJoyti/go-coffee")
	},
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Manage task-cli configuration settings",
}

func init() {
	// Config subcommands
	configCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Create default configuration file",
		Long:  "Create a default configuration file in the user's config directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.CreateDefaultConfig(); err != nil {
				printError("Failed to create config: %v", err)
				return err
			}
			printSuccess("Default configuration created successfully")
			printInfo("Edit the configuration file to customize settings")
			printInfo("Config file location: %s", config.GetConfigPath())
			return nil
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current configuration settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg == nil {
				printError("Configuration not loaded")
				return fmt.Errorf("configuration not loaded")
			}

			fmt.Printf("Configuration file: %s\n\n", config.GetConfigPath())
			fmt.Printf("Redis:\n")
			fmt.Printf("  Host: %s\n", cfg.Redis.Host)
			fmt.Printf("  Port: %d\n", cfg.Redis.Port)
			fmt.Printf("  DB: %d\n", cfg.Redis.DB)
			fmt.Printf("  URL: %s\n", cfg.Redis.URL)
			fmt.Printf("\nCLI:\n")
			fmt.Printf("  Default User: %s\n", cfg.CLI.DefaultUser)
			fmt.Printf("  Date Format: %s\n", cfg.CLI.DateFormat)
			fmt.Printf("  Output Format: %s\n", cfg.CLI.OutputFormat)
			fmt.Printf("  Color Output: %t\n", cfg.CLI.ColorOutput)
			fmt.Printf("  Page Size: %d\n", cfg.CLI.PageSize)
			fmt.Printf("  Sort By: %s\n", cfg.CLI.SortBy)
			fmt.Printf("  Sort Order: %s\n", cfg.CLI.SortOrder)
			fmt.Printf("\nDefaults:\n")
			fmt.Printf("  Priority: %s\n", cfg.Defaults.Priority)
			fmt.Printf("  Status: %s\n", cfg.Defaults.Status)
			fmt.Printf("  Tags: %v\n", cfg.Defaults.Tags)

			return nil
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Long:  "Display the path to the configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.GetConfigPath())
		},
	})
}
