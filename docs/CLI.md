# 🚀 Go Coffee CLI

Next-Generation Cloud-Native Platform Command Line Interface

## 📋 Overview

The Go Coffee CLI is a powerful command-line tool for managing cloud-native microservices, Kubernetes operators, and Infrastructure as Code. It provides a unified interface for all Go Coffee platform operations.

## ✨ Features

### 🏗️ **Multi-Service Orchestration**
- Manage all Go Coffee microservices
- Start, stop, restart, and scale services
- Real-time health monitoring
- Log aggregation and analysis

### ☸️ **Kubernetes Management**
- Deploy and manage custom operators
- Handle CRDs and custom resources
- Monitor workloads and resources
- Apply and manage manifests

### 🌐 **Cloud Infrastructure**
- Multi-cloud deployment automation
- Infrastructure as Code with Terraform
- Cost monitoring and optimization
- Resource lifecycle management

### 🔒 **Security & Compliance**
- Policy enforcement with OPA
- RBAC and access control
- Security scanning and compliance
- Certificate management

### 📊 **Observability**
- Distributed tracing with OpenTelemetry
- Metrics collection and visualization
- Log aggregation and analysis
- Performance monitoring

### 🔄 **GitOps Workflows**
- ArgoCD and Flux integration
- Automated deployments
- Git repository management
- Sync and rollback operations

## 🚀 Quick Start

### Installation

#### Binary Installation
```bash
# Download latest release
curl -L https://github.com/DimaJoyti/go-coffee/releases/latest/download/gocoffee-linux-amd64.tar.gz | tar xz
sudo mv gocoffee /usr/local/bin/

# Verify installation
gocoffee version
```

#### Build from Source
```bash
# Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Build CLI
make -f Makefile.cli build

# Install
make -f Makefile.cli install
```

#### Docker
```bash
# Run in Docker
docker run --rm -it ghcr.io/dimajoyti/gocoffee-cli:latest

# Build locally
make -f Makefile.cli docker-build
```

### Configuration

Create configuration file:
```bash
# Generate example config
gocoffee config init

# Edit configuration
vim ~/.gocoffee/config.yaml
```

Example configuration:
```yaml
log_level: info
telemetry:
  enabled: true
  service_name: gocoffee-cli
kubernetes:
  config_path: ~/.kube/config
  namespace: go-coffee
cloud:
  provider: gcp
  region: us-central1
  project: my-project
```

## 📚 Usage Examples

### Services Management

```bash
# List all services
gocoffee services list

# Start specific services
gocoffee services start api-gateway auth-service

# Start all services
gocoffee services start --all

# Check service health
gocoffee services health

# View service logs
gocoffee services logs api-gateway --follow

# Scale service
gocoffee services scale order-service 5

# Deploy new version
gocoffee services deploy api-gateway --image myregistry/api-gateway:v2.0.0
```

### Kubernetes Operations

```bash
# Get all resources
gocoffee kubernetes get all

# Apply manifests
gocoffee kubernetes apply -f k8s/

# List operators
gocoffee kubernetes operators list

# Install operator
gocoffee kubernetes operators install llm-orchestrator

# Manage custom workloads
gocoffee kubernetes workloads list
gocoffee kubernetes workloads create -f workload.yaml
```

### Cloud Infrastructure

```bash
# Initialize infrastructure
gocoffee cloud init --provider gcp --region us-central1

# Plan changes
gocoffee cloud plan --env production

# Apply changes
gocoffee cloud apply --env production

# List resources
gocoffee cloud resources

# Show costs
gocoffee cloud cost --breakdown
```

### Security & Compliance

```bash
# Security scan
gocoffee security scan

# Policy validation
gocoffee security policy validate

# Certificate management
gocoffee security certs list
gocoffee security certs renew
```

### GitOps Workflows

```bash
# Sync deployments
gocoffee gitops sync

# Rollback deployment
gocoffee gitops rollback --revision 5

# Repository management
gocoffee gitops repo add https://github.com/myorg/configs
```

### Observability

```bash
# Open monitoring dashboard
gocoffee observability dashboard

# View metrics
gocoffee observability metrics

# Trace analysis
gocoffee observability trace
```

## 🔧 Advanced Usage

### Custom Operators

The CLI can manage custom Kubernetes operators:

```bash
# LLM Orchestrator Operator
gocoffee k8s operators install llm-orchestrator
gocoffee k8s workloads create llm-workload.yaml

# Coffee Service Operator
gocoffee k8s operators install coffee-operator
gocoffee k8s workloads create coffee-order.yaml

# Multi-tenant Operator
gocoffee k8s operators install multitenant-operator
```

### Infrastructure as Code

```bash
# Terraform integration
gocoffee cloud init --provider gcp
gocoffee cloud plan --var-file production.tfvars
gocoffee cloud apply --auto-approve

# Helm integration
gocoffee k8s helm install go-coffee ./helm-chart
gocoffee k8s helm upgrade go-coffee ./helm-chart
```

### Multi-Environment Management

```bash
# Environment-specific operations
gocoffee --env development services start --all
gocoffee --env staging cloud plan
gocoffee --env production gitops sync
```

## 🛠️ Development

### Building

```bash
# Install dependencies
make -f Makefile.cli deps

# Build binary
make -f Makefile.cli build

# Run tests
make -f Makefile.cli test

# Run linter
make -f Makefile.cli lint

# Development mode
make -f Makefile.cli dev
```

### Contributing

1. Fork the repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

## 📖 Documentation

- [Architecture Guide](./ARCHITECTURE.md)
- [API Reference](./API.md)
- [Configuration Reference](./CONFIG.md)
- [Troubleshooting](./TROUBLESHOOTING.md)

## 🤝 Support

- GitHub Issues: [Report bugs](https://github.com/DimaJoyti/go-coffee/issues)
- Discussions: [Community support](https://github.com/DimaJoyti/go-coffee/discussions)
- Documentation: [Full docs](https://docs.gocoffee.dev)

## 📄 License

MIT License - see [LICENSE](../LICENSE) file for details.

---

**Go Coffee CLI** - Powering the next generation of cloud-native applications ☕️
