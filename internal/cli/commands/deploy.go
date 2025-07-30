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
	"go.uber.org/zap"
)

const (
	// DefaultHealthCheckPath is the default health check endpoint
	DefaultHealthCheckPath = "/health"
)

// DeploymentConfig holds deployment configuration
type DeploymentConfig struct {
	Environment string
	Services    []string
	DryRun      bool
	Parallel    bool
	Timeout     time.Duration
}

// ServiceDeployment represents a service deployment
type ServiceDeployment struct {
	Name         string
	Image        string
	Version      string
	Status       string
	HealthCheck  string
	Dependencies []string
}

// deployServices handles the deployment of services
func deployServices(ctx context.Context, cfg *config.Config, logger *zap.Logger, services []string, environment string, dryRun, parallel bool, timeout time.Duration) error {
	deployConfig := &DeploymentConfig{
		Environment: environment,
		Services:    services,
		DryRun:      dryRun,
		Parallel:    parallel,
		Timeout:     timeout,
	}

	logger.Info("Starting deployment",
		zap.String("environment", environment),
		zap.Strings("services", services),
		zap.Bool("dry_run", dryRun),
		zap.Bool("parallel", parallel),
		zap.Duration("timeout", timeout),
	)

	// Get list of services to deploy
	servicesToDeploy, err := getServicesToDeploy(services, environment)
	if err != nil {
		return fmt.Errorf("failed to get services to deploy: %w", err)
	}

	if dryRun {
		return showDeploymentPlan(servicesToDeploy, deployConfig)
	}

	// Pre-deployment checks
	if err := runPreDeploymentChecks(ctx, cfg, logger, servicesToDeploy, environment); err != nil {
		return fmt.Errorf("pre-deployment checks failed: %w", err)
	}

	// Execute deployment
	if parallel {
		return deployServicesParallel(ctx, cfg, logger, servicesToDeploy, deployConfig)
	}

	return deployServicesSequential(ctx, cfg, logger, servicesToDeploy, deployConfig)
}

// getServicesToDeploy returns the list of services to deploy
func getServicesToDeploy(services []string, environment string) ([]*ServiceDeployment, error) {
	allServices := map[string]*ServiceDeployment{
		"api-gateway": {
			Name:         "api-gateway",
			Image:        "go-coffee/api-gateway",
			Version:      "latest",
			HealthCheck:  DefaultHealthCheckPath,
			Dependencies: []string{"redis", "postgres"},
		},
		"auth-service": {
			Name:         "auth-service",
			Image:        "go-coffee/auth-service",
			Version:      "latest",
			HealthCheck:  DefaultHealthCheckPath,
			Dependencies: []string{"postgres", "redis"},
		},
		"order-service": {
			Name:         "order-service",
			Image:        "go-coffee/order-service",
			Version:      "latest",
			HealthCheck:  DefaultHealthCheckPath,
			Dependencies: []string{"postgres", "redis", "kafka"},
		},
		"kitchen-service": {
			Name:         "kitchen-service",
			Image:        "go-coffee/kitchen-service",
			Version:      "latest",
			HealthCheck:  DefaultHealthCheckPath,
			Dependencies: []string{"postgres", "redis"},
		},
		"payment-service": {
			Name:         "payment-service",
			Image:        "go-coffee/payment-service",
			Version:      "latest",
			HealthCheck:  DefaultHealthCheckPath,
			Dependencies: []string{"postgres", "redis"},
		},
		"producer": {
			Name:         "producer",
			Image:        "go-coffee/producer",
			Version:      "latest",
			HealthCheck:  DefaultHealthCheckPath,
			Dependencies: []string{"kafka", "redis"},
		},
		"consumer": {
			Name:         "consumer",
			Image:        "go-coffee/consumer",
			Version:      "latest",
			HealthCheck:  DefaultHealthCheckPath,
			Dependencies: []string{"kafka", "postgres"},
		},
		"streams": {
			Name:         "streams",
			Image:        "go-coffee/streams",
			Version:      "latest",
			HealthCheck:  DefaultHealthCheckPath,
			Dependencies: []string{"kafka"},
		},
	}

	var servicesToDeploy []*ServiceDeployment

	if len(services) == 0 {
		// Deploy all services
		for _, service := range allServices {
			servicesToDeploy = append(servicesToDeploy, service)
		}
	} else {
		// Deploy specified services
		for _, serviceName := range services {
			if service, exists := allServices[serviceName]; exists {
				servicesToDeploy = append(servicesToDeploy, service)
			} else {
				return nil, fmt.Errorf("unknown service: %s", serviceName)
			}
		}
	}

	return servicesToDeploy, nil
}

// showDeploymentPlan shows what would be deployed in dry-run mode
func showDeploymentPlan(services []*ServiceDeployment, config *DeploymentConfig) error {
	color.Yellow("🔍 Deployment Plan (Dry Run)")
	color.Yellow("=" + strings.Repeat("=", 50))

	fmt.Printf("Environment: %s\n", config.Environment)
	fmt.Printf("Parallel: %t\n", config.Parallel)
	fmt.Printf("Timeout: %s\n\n", config.Timeout)

	fmt.Println("Services to deploy:")
	for i, service := range services {
		fmt.Printf("  %d. %s\n", i+1, service.Name)
		fmt.Printf("     Image: %s:%s\n", service.Image, service.Version)
		fmt.Printf("     Health Check: %s\n", service.HealthCheck)
		if len(service.Dependencies) > 0 {
			fmt.Printf("     Dependencies: %s\n", strings.Join(service.Dependencies, ", "))
		}
		fmt.Println()
	}

	color.Yellow("Note: This is a dry run. No actual deployment will occur.")
	return nil
}

// runPreDeploymentChecks performs pre-deployment validation
func runPreDeploymentChecks(ctx context.Context, cfg *config.Config, logger *zap.Logger, services []*ServiceDeployment, environment string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Running pre-deployment checks..."
	s.Start()
	defer s.Stop()

	// Check Docker daemon
	if err := checkDockerDaemon(); err != nil {
		return fmt.Errorf("Docker daemon check failed: %w", err)
	}

	// Check Kubernetes connectivity (if deploying to k8s)
	if environment != "local" {
		if err := checkKubernetesConnectivity(); err != nil {
			return fmt.Errorf("Kubernetes connectivity check failed: %w", err)
		}
	}

	// Check service dependencies
	for _, service := range services {
		if err := checkServiceDependencies(service); err != nil {
			return fmt.Errorf("dependency check failed for %s: %w", service.Name, err)
		}
	}

	// Validate configuration files
	if err := validateConfigFiles(environment); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	s.Stop()
	color.Green("✅ Pre-deployment checks passed")
	return nil
}

// deployServicesSequential deploys services one by one
func deployServicesSequential(ctx context.Context, cfg *config.Config, logger *zap.Logger, services []*ServiceDeployment, config *DeploymentConfig) error {
	color.Cyan("🚀 Starting sequential deployment...")

	for i, service := range services {
		fmt.Printf("\n[%d/%d] Deploying %s...\n", i+1, len(services), service.Name)

		if err := deployService(ctx, cfg, logger, service, config); err != nil {
			return fmt.Errorf("failed to deploy %s: %w", service.Name, err)
		}

		color.Green("✅ %s deployed successfully", service.Name)
	}

	color.Green("\n🎉 All services deployed successfully!")
	return nil
}

// deployServicesParallel deploys services in parallel
func deployServicesParallel(ctx context.Context, cfg *config.Config, logger *zap.Logger, services []*ServiceDeployment, config *DeploymentConfig) error {
	color.Cyan("🚀 Starting parallel deployment...")

	// Create channels for coordination
	results := make(chan error, len(services))

	// Deploy services in parallel
	for _, service := range services {
		go func(svc *ServiceDeployment) {
			err := deployService(ctx, cfg, logger, svc, config)
			if err != nil {
				results <- fmt.Errorf("failed to deploy %s: %w", svc.Name, err)
			} else {
				results <- nil
			}
		}(service)
	}

	// Wait for all deployments to complete
	var errors []string
	for i := 0; i < len(services); i++ {
		if err := <-results; err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("deployment errors: %s", strings.Join(errors, "; "))
	}

	color.Green("\n🎉 All services deployed successfully!")
	return nil
}

// deployService deploys a single service
func deployService(ctx context.Context, cfg *config.Config, logger *zap.Logger, service *ServiceDeployment, config *DeploymentConfig) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Deploying %s...", service.Name)
	s.Start()
	defer s.Stop()

	logger.Info("Deploying service",
		zap.String("service", service.Name),
		zap.String("environment", config.Environment),
	)

	// Build service image
	if err := buildServiceImage(service, config.Environment); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	// Deploy based on environment
	switch config.Environment {
	case "local", "development":
		return deployToDocker(service, config)
	case "staging", "production":
		return deployToKubernetes(service, config)
	default:
		return fmt.Errorf("unsupported environment: %s", config.Environment)
	}
}

// buildServiceImage builds the Docker image for a service
func buildServiceImage(service *ServiceDeployment, environment string) error {
	dockerfilePath := filepath.Join("docker", fmt.Sprintf("Dockerfile.%s", service.Name))
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		dockerfilePath = filepath.Join("cmd", service.Name, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			// Use generic Dockerfile
			dockerfilePath = "Dockerfile"
		}
	}

	imageTag := fmt.Sprintf("%s:%s", service.Image, service.Version)

	cmd := exec.Command("docker", "build", "-t", imageTag, "-f", dockerfilePath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// deployToDocker deploys service using Docker Compose
func deployToDocker(service *ServiceDeployment, config *DeploymentConfig) error {
	composeFile := "docker/docker-compose.yml"
	if config.Environment != "local" {
		composeFile = fmt.Sprintf("docker/docker-compose.%s.yml", config.Environment)
	}

	cmd := exec.Command("docker-compose", "-f", composeFile, "up", "-d", service.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// deployToKubernetes deploys service to Kubernetes
func deployToKubernetes(service *ServiceDeployment, config *DeploymentConfig) error {
	manifestPath := filepath.Join("k8s", config.Environment, fmt.Sprintf("%s.yaml", service.Name))

	cmd := exec.Command("kubectl", "apply", "-f", manifestPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Helper functions for checks
func checkDockerDaemon() error {
	cmd := exec.Command("docker", "info")
	return cmd.Run()
}

func checkKubernetesConnectivity() error {
	cmd := exec.Command("kubectl", "cluster-info")
	return cmd.Run()
}

func checkServiceDependencies(service *ServiceDeployment) error {
	// This would check if dependencies are available
	// For now, just return nil
	return nil
}

func validateConfigFiles(environment string) error {
	// This would validate configuration files
	// For now, just return nil
	return nil
}
