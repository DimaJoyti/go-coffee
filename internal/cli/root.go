package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/cli/commands"
	"github.com/DimaJoyti/go-coffee/internal/cli/config"
	"github.com/DimaJoyti/go-coffee/internal/cli/telemetry"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewRootCommand creates the root command for the Go Coffee CLI
func NewRootCommand(cfg *config.Config, logger *zap.Logger, version, commit, date string) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "gocoffee",
		Short: "Go Coffee - Next-Generation Cloud-Native Platform CLI",
		Long: color.New(color.FgCyan, color.Bold).Sprint(`
╔═══════════════════════════════════════════════════════════════╗
║                    🚀 Go Coffee CLI                          ║
║           Next-Generation Cloud-Native Platform              ║
╚═══════════════════════════════════════════════════════════════╝

A powerful CLI for managing cloud-native microservices, 
Kubernetes operators, and Infrastructure as Code.

Features:
  • 🏗️  Multi-Service Orchestration
  • ☸️  Kubernetes Operator Management  
  • 🌐 Cloud Infrastructure Automation
  • 🔒 Security & Policy Enforcement
  • 📊 Observability & Monitoring
  • 🔄 GitOps Workflows
`),
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Record command start time
			start := time.Now()
			ctx := cmd.Context()

			// Store start time in context for later use
			ctx = context.WithValue(ctx, "start_time", start)
			cmd.SetContext(ctx)

			logger.Info("Command started",
				zap.String("command", cmd.CommandPath()),
				zap.Strings("args", args),
			)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Calculate command duration
			start, ok := cmd.Context().Value("start_time").(time.Time)
			if !ok {
				start = time.Now()
			}
			duration := time.Since(start)

			// Record telemetry
			telemetry.RecordCommandExecution(cmd.Context(), cmd.CommandPath(), duration, nil)

			logger.Info("Command completed",
				zap.String("command", cmd.CommandPath()),
				zap.Duration("duration", duration),
			)

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.gocoffee/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(commands.NewServicesCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewKubernetesCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewCloudCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewSecurityCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewGitOpsCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewObservabilityCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewConfigCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewVersionCommand(version, commit, date))

	// 3: Advanced cloud-native commands
	rootCmd.AddCommand(commands.NewMultiCloudCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewEdgeCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewMLOpsCommand(cfg, logger))

	// 4: Future technologies commands
	rootCmd.AddCommand(commands.NewQuantumCommand(cfg, logger))
	rootCmd.AddCommand(commands.NewBlockchainCommand(cfg, logger))

	return rootCmd
}
