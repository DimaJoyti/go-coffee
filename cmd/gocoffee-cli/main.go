package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/DimaJoyti/go-coffee/internal/cli"
	"github.com/DimaJoyti/go-coffee/internal/cli/config"
	"github.com/DimaJoyti/go-coffee/internal/cli/telemetry"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Initialize context with cancellation
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := initLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Initialize telemetry
	telemetryShutdown, err := telemetry.Init(ctx, cfg.Telemetry)
	if err != nil {
		logger.Error("Failed to initialize telemetry", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		if err := telemetryShutdown(ctx); err != nil {
			logger.Error("Failed to shutdown telemetry", zap.Error(err))
		}
	}()

	// Create root command
	rootCmd := cli.NewRootCommand(cfg, logger, version, commit, date)

	// Execute command
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.Error("Command execution failed", zap.Error(err))
		os.Exit(1)
	}
}

func initLogger(level string) (*zap.Logger, error) {
	var config zap.Config
	
	switch level {
	case "debug":
		config = zap.NewDevelopmentConfig()
	case "info", "warn", "error":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(parseLogLevel(level))
	default:
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return config.Build()
}

func parseLogLevel(level string) zap.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}
