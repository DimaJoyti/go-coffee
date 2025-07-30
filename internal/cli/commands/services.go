package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/cli/config"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Service represents a Go Coffee service
type Service struct {
	Name        string
	Type        string
	Status      string
	Port        int
	Health      string
	Version     string
	Replicas    int
	CPU         string
	Memory      string
	LastUpdated time.Time
}

// NewServicesCommand creates the services management command
func NewServicesCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var servicesCmd = &cobra.Command{
		Use:   "services",
		Short: "Manage Go Coffee microservices",
		Long: `Manage and monitor all Go Coffee microservices including:
  • Coffee ordering services (producer, consumer, streams)
  • AI services (arbitrage, order processing, search)
  • Infrastructure services (auth, gateway, kitchen)
  • Web3 services (payment, DeFi, crypto)`,
		Aliases: []string{"svc", "service"},
	}

	// Add subcommands
	servicesCmd.AddCommand(newServicesListCommand(cfg, logger))
	servicesCmd.AddCommand(newServicesStartCommand(cfg, logger))
	servicesCmd.AddCommand(newServicesStopCommand(cfg, logger))
	servicesCmd.AddCommand(newServicesRestartCommand(cfg, logger))
	servicesCmd.AddCommand(newServicesLogsCommand(cfg, logger))
	servicesCmd.AddCommand(newServicesHealthCommand(cfg, logger))
	servicesCmd.AddCommand(newServicesScaleCommand(cfg, logger))
	servicesCmd.AddCommand(newServicesDeployCommand(cfg, logger))

	return servicesCmd
}

// newServicesListCommand creates the list services command
func newServicesListCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var format string
	var filter string

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all services",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = " Discovering services..."
			s.Start()

			services, err := discoverServices(ctx, cfg, filter)
			s.Stop()

			if err != nil {
				return fmt.Errorf("failed to discover services: %w", err)
			}

			switch format {
			case "json":
				return printServicesJSON(services)
			case "yaml":
				return printServicesYAML(services)
			default:
				return printServicesTable(services)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter services by type or status")

	return cmd
}

// newServicesStartCommand creates the start services command
func newServicesStartCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var all bool
	var parallel bool

	cmd := &cobra.Command{
		Use:   "start [service...]",
		Short: "Start one or more services",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if all {
				return startAllServices(ctx, cfg, logger, parallel)
			}

			if len(args) == 0 {
				return fmt.Errorf("specify service names or use --all flag")
			}

			return startServices(ctx, cfg, logger, args, parallel)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Start all services")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Start services in parallel")

	return cmd
}

// newServicesStopCommand creates the stop services command
func newServicesStopCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var all bool
	var force bool

	cmd := &cobra.Command{
		Use:   "stop [service...]",
		Short: "Stop one or more services",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if all {
				return stopAllServices(ctx, cfg, logger, force)
			}

			if len(args) == 0 {
				return fmt.Errorf("specify service names or use --all flag")
			}

			return stopServices(ctx, cfg, logger, args, force)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Stop all services")
	cmd.Flags().BoolVar(&force, "force", false, "Force stop services")

	return cmd
}

// newServicesRestartCommand creates the restart services command
func newServicesRestartCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart [service...]",
		Short: "Restart one or more services",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) == 0 {
				return fmt.Errorf("specify service names")
			}

			return restartServices(ctx, cfg, logger, args)
		},
	}

	return cmd
}

// newServicesLogsCommand creates the logs command
func newServicesLogsCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var follow bool
	var tail int
	var since string

	cmd := &cobra.Command{
		Use:   "logs [service]",
		Short: "Show service logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) == 0 {
				return fmt.Errorf("specify service name")
			}

			return showServiceLogs(ctx, cfg, logger, args[0], follow, tail, since)
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().IntVar(&tail, "tail", 100, "Number of lines to show from the end")
	cmd.Flags().StringVar(&since, "since", "", "Show logs since timestamp (e.g. 2h, 1h30m)")

	return cmd
}

// newServicesHealthCommand creates the health check command
func newServicesHealthCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health [service...]",
		Short: "Check service health",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) == 0 {
				return checkAllServicesHealth(ctx, cfg, logger)
			}

			return checkServicesHealth(ctx, cfg, logger, args)
		},
	}

	return cmd
}

// newServicesScaleCommand creates the scale command
func newServicesScaleCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale [service] [replicas]",
		Short: "Scale service replicas",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) != 2 {
				return fmt.Errorf("specify service name and replica count")
			}

			return scaleService(ctx, cfg, logger, args[0], args[1])
		},
	}

	return cmd
}

// newServicesDeployCommand creates the deploy command
func newServicesDeployCommand(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	var image string
	var tag string
	var env string

	cmd := &cobra.Command{
		Use:   "deploy [service]",
		Short: "Deploy service with new image",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) == 0 {
				return fmt.Errorf("specify service name")
			}

			return deployServiceLegacy(ctx, cfg, logger, args[0], image, tag, env)
		},
	}

	cmd.Flags().StringVar(&image, "image", "", "Container image")
	cmd.Flags().StringVar(&tag, "tag", "latest", "Image tag")
	cmd.Flags().StringVar(&env, "env", "development", "Environment (development, staging, production)")

	return cmd
}

// Helper functions for service operations
func discoverServices(ctx context.Context, cfg *config.Config, filter string) ([]Service, error) {
	// Mock implementation - in real scenario, this would query Kubernetes, Docker, etc.
	services := []Service{
		{Name: "api-gateway", Type: "gateway", Status: "running", Port: 8080, Health: "healthy", Version: "v1.0.0", Replicas: 2, CPU: "100m", Memory: "256Mi", LastUpdated: time.Now().Add(-1 * time.Hour)},
		{Name: "auth-service", Type: "auth", Status: "running", Port: 8081, Health: "healthy", Version: "v1.0.0", Replicas: 3, CPU: "50m", Memory: "128Mi", LastUpdated: time.Now().Add(-30 * time.Minute)},
		{Name: "order-service", Type: "business", Status: "running", Port: 8082, Health: "healthy", Version: "v1.1.0", Replicas: 5, CPU: "200m", Memory: "512Mi", LastUpdated: time.Now().Add(-15 * time.Minute)},
		{Name: "kitchen-service", Type: "business", Status: "running", Port: 50052, Health: "healthy", Version: "v1.0.0", Replicas: 2, CPU: "150m", Memory: "256Mi", LastUpdated: time.Now().Add(-45 * time.Minute)},
		{Name: "ai-arbitrage", Type: "ai", Status: "running", Port: 8090, Health: "healthy", Version: "v0.9.0", Replicas: 1, CPU: "500m", Memory: "1Gi", LastUpdated: time.Now().Add(-2 * time.Hour)},
		{Name: "payment-service", Type: "web3", Status: "stopped", Port: 8083, Health: "unhealthy", Version: "v1.0.0", Replicas: 0, CPU: "0", Memory: "0", LastUpdated: time.Now().Add(-3 * time.Hour)},
	}

	// Apply filter if specified
	if filter != "" {
		var filtered []Service
		for _, svc := range services {
			if strings.Contains(svc.Type, filter) || strings.Contains(svc.Status, filter) || strings.Contains(svc.Name, filter) {
				filtered = append(filtered, svc)
			}
		}
		services = filtered
	}

	return services, nil
}

func printServicesTable(services []Service) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "Status", "Health", "Port", "Replicas", "CPU", "Memory", "Version", "Last Updated"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, svc := range services {
		status := svc.Status
		health := svc.Health

		// Color code status and health
		if svc.Status == "running" {
			status = color.GreenString("●") + " " + svc.Status
		} else {
			status = color.RedString("●") + " " + svc.Status
		}

		if svc.Health == "healthy" {
			health = color.GreenString("✓") + " " + svc.Health
		} else {
			health = color.RedString("✗") + " " + svc.Health
		}

		table.Append([]string{
			svc.Name,
			svc.Type,
			status,
			health,
			fmt.Sprintf("%d", svc.Port),
			fmt.Sprintf("%d", svc.Replicas),
			svc.CPU,
			svc.Memory,
			svc.Version,
			svc.LastUpdated.Format("15:04:05"),
		})
	}

	table.Render()
	return nil
}

// Additional helper functions for service operations
func printServicesJSON(services []Service) error {
	// TODO: Implement JSON output
	fmt.Println("JSON output not implemented yet")
	return nil
}

func printServicesYAML(services []Service) error {
	// TODO: Implement YAML output
	fmt.Println("YAML output not implemented yet")
	return nil
}

func startAllServices(ctx context.Context, cfg *config.Config, logger *zap.Logger, parallel bool) error {
	// TODO: Implement start all services
	logger.Info("Starting all services", zap.Bool("parallel", parallel))
	return nil
}

func startServices(ctx context.Context, cfg *config.Config, logger *zap.Logger, services []string, parallel bool) error {
	// TODO: Implement start services
	logger.Info("Starting services", zap.Strings("services", services), zap.Bool("parallel", parallel))
	return nil
}

func stopAllServices(ctx context.Context, cfg *config.Config, logger *zap.Logger, force bool) error {
	// TODO: Implement stop all services
	logger.Info("Stopping all services", zap.Bool("force", force))
	return nil
}

func stopServices(ctx context.Context, cfg *config.Config, logger *zap.Logger, services []string, force bool) error {
	// TODO: Implement stop services
	logger.Info("Stopping services", zap.Strings("services", services), zap.Bool("force", force))
	return nil
}

func restartServices(ctx context.Context, cfg *config.Config, logger *zap.Logger, services []string) error {
	// TODO: Implement restart services
	logger.Info("Restarting services", zap.Strings("services", services))
	return nil
}

func showServiceLogs(ctx context.Context, cfg *config.Config, logger *zap.Logger, service string, follow bool, tail int, since string) error {
	// TODO: Implement show logs
	logger.Info("Showing logs", zap.String("service", service), zap.Bool("follow", follow))
	return nil
}

func checkAllServicesHealth(ctx context.Context, cfg *config.Config, logger *zap.Logger) error {
	// TODO: Implement health check all
	logger.Info("Checking health of all services")
	return nil
}

func checkServicesHealth(ctx context.Context, cfg *config.Config, logger *zap.Logger, services []string) error {
	// TODO: Implement health check
	logger.Info("Checking health", zap.Strings("services", services))
	return nil
}

func scaleService(ctx context.Context, cfg *config.Config, logger *zap.Logger, service, replicas string) error {
	// TODO: Implement scale service
	logger.Info("Scaling service", zap.String("service", service), zap.String("replicas", replicas))
	return nil
}

func deployServiceLegacy(ctx context.Context, cfg *config.Config, logger *zap.Logger, service, image, tag, env string) error {
	// TODO: Implement deploy service
	logger.Info("Deploying service",
		zap.String("service", service),
		zap.String("image", image),
		zap.String("tag", tag),
		zap.String("env", env))
	return nil
}
