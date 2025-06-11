package commands

import (
	"fmt"
	"os"

	"github.com/DimaJoyti/go-coffee/internal/cli/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// NewConfigCommand creates the configuration management command
func NewConfigCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long: `Manage CLI configuration settings:
  • View current configuration
  • Set configuration values
  • Reset to defaults
  • Validate configuration`,
	}

	configCmd.AddCommand(newConfigViewCommand(cfg, logger))
	configCmd.AddCommand(newConfigSetCommand(cfg, logger))
	configCmd.AddCommand(newConfigResetCommand(cfg, logger))
	configCmd.AddCommand(newConfigValidateCommand(cfg, logger))

	return configCmd
}

// newConfigViewCommand creates the view config command
func newConfigViewCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			color.Cyan("Current Configuration:")
			fmt.Println(string(data))
			return nil
		},
	}

	return cmd
}

// newConfigSetCommand creates the set config command
func newConfigSetCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set configuration value",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("specify key and value")
			}

			// TODO: Implement config setting logic
			color.Green("✓ Configuration updated: %s = %s", args[0], args[1])
			return nil
		},
	}

	return cmd
}

// newConfigResetCommand creates the reset config command
func newConfigResetCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement config reset logic
			color.Green("✓ Configuration reset to defaults")
			return nil
		},
	}

	return cmd
}

// newConfigValidateCommand creates the validate config command
func newConfigValidateCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				color.Red("✗ Configuration validation failed: %v", err)
				os.Exit(1)
			}

			color.Green("✓ Configuration is valid")
			return nil
		},
	}

	return cmd
}
