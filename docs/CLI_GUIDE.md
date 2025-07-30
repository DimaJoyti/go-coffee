# 🚀 Go Coffee  CLI - Comprehensive Guide

A powerful command-line interface for comprehensive  operations on the Go Coffee platform.

## 📋 Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Commands](#core-commands)
- [Advanced Usage](#advanced-usage)
- [Configuration](#configuration)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## 🎯 Overview

The Go Coffee  CLI provides comprehensive  operations including:

- **🏗️ Project Management** - Initialize and manage project environments
- **🚀 Deployment** - Deploy services with zero-downtime strategies
- **📊 Monitoring** - Real-time health checks and system status
- **📝 Log Management** - Centralized log aggregation and analysis
- **🗄️ Database Operations** - Migrations, seeding, and backup/restore
- **⚙️ Configuration Management** - Environment-specific configurations
- **🧪 Testing** - Automated testing across all service layers
- **🔒 Security** - Built-in security scanning and compliance

## 📦 Installation

### Binary Installation

```bash
# Download latest release
curl -L https://github.com/DimaJoyti/go-coffee/releases/latest/download/coffee-linux-amd64.tar.gz | tar xz
sudo mv coffee /usr/local/bin/

# Verify installation
coffee version
```

### Build from Source

```bash
# Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Build CLI
make -f Makefile.cli build

# Install globally
make -f Makefile.cli install
```

### Docker

```bash
# Run in Docker
docker run --rm -it ghcr.io/dimajoyti/coffee:latest

# Build locally
make -f Makefile.cli docker-build
```

## 🚀 Quick Start

### 1. Initialize New Environment

```bash
# Initialize development environment
coffee  init my-coffee-app --env development

# Initialize with specific template
coffee  init my-coffee-app --env production --template enterprise
```

### 2. Deploy Services

```bash
# Deploy all services
coffee  deploy --env development

# Deploy specific services
coffee  deploy api-gateway auth-service --env staging

# Dry run deployment
coffee  deploy --dry-run --env production
```

### 3. Monitor System Health

```bash
# Check system status
coffee  status --env production

# Watch status continuously
coffee  status --watch --interval 5s

# Check specific service health
coffee  health check --service api-gateway
```

## 🛠️ Core Commands

### Project Initialization

```bash
# Initialize new project
coffee  init <project-name> [flags]

Flags:
  -e, --env string        Environment (development, staging, production)
  -t, --template string   Template (standard, minimal, enterprise)
  -f, --force            Force initialization even if files exist
```

**Example:**
```bash
coffee  init my-app --env development --template standard
```

### Service Deployment

```bash
# Deploy services
coffee  deploy [service...] [flags]

Flags:
  -e, --env string       Target environment
      --dry-run         Show deployment plan without executing
  -p, --parallel        Deploy services in parallel
      --timeout duration Deployment timeout (default 10m)
```

**Examples:**
```bash
# Deploy all services
coffee  deploy --env production

# Deploy specific services in parallel
coffee  deploy api-gateway auth-service --parallel --env staging

# Dry run for production
coffee  deploy --dry-run --env production
```

### System Status and Health

```bash
# Show system status
coffee  status [flags]

Flags:
  -e, --env string       Environment to check
  -f, --format string    Output format (table, json, yaml)
  -w, --watch           Watch status continuously
      --interval duration Watch interval (default 5s)
```

**Examples:**
```bash
# Basic status check
coffee  status --env production

# Watch status with custom interval
coffee  status --watch --interval 10s --env staging

# JSON output for automation
coffee  status --format json --env production
```

### Log Management

```bash
# View logs
coffee  logs [service] [flags]

Flags:
  -e, --env string      Environment
  -f, --follow         Follow log output
      --tail int       Number of lines to show (default 100)
      --since string   Show logs since timestamp
      --grep string    Filter logs by pattern
```

**Examples:**
```bash
# View all service logs
coffee  logs --env production

# Follow specific service logs
coffee  logs api-gateway --follow --env staging

# Filter logs with pattern
coffee  logs --grep "ERROR" --tail 500 --env production
```

### Database Operations

```bash
# Database migrations
coffee  migrate <command> [flags]

Commands:
  up      Run pending migrations
  down    Rollback migrations
  status  Show migration status
  seed    Seed database with test data
  generate Generate new migration
```

**Examples:**
```bash
# Run all pending migrations
coffee  migrate up --env production --backup

# Rollback last 2 migrations
coffee  migrate down --steps 2 --env staging

# Check migration status
coffee  migrate status --env production --format table

# Seed development database
coffee  migrate seed --env development --force
```

### Backup and Restore

```bash
# Create backup
coffee  backup create [flags]

Flags:
  -e, --env string      Environment to backup
  -t, --type string     Backup type (full, database, config, volumes)
  -c, --compress       Compress backup
      --encrypt        Encrypt backup
      --remote         Upload to remote storage
```

**Examples:**
```bash
# Full system backup
coffee  backup create --env production --type full --compress --encrypt

# Database only backup
coffee  backup create --env staging --type database --remote

# List available backups
coffee  backup list --env production

# Restore from backup
coffee  restore --backup backup-20240101-120000 --env staging
```

### Testing

```bash
# Run tests
coffee  test [service] [flags]

Flags:
  -e, --env string      Environment to test against
  -t, --type string     Test type (unit, integration, e2e, performance, security)
  -c, --coverage       Generate coverage report
  -p, --parallel       Run tests in parallel
```

**Examples:**
```bash
# Run all unit tests with coverage
coffee  test --type unit --coverage

# Run integration tests for specific service
coffee  test api-gateway --type integration --env staging

# Run performance tests
coffee  test --type performance --env production
```

### Environment Management

```bash
# Environment operations
coffee  env <command> [flags]

Commands:
  list    List available environments
  switch  Switch to different environment
  create  Create new environment
  delete  Delete environment
```

**Examples:**
```bash
# List environments
coffee  env list

# Switch environment
coffee  env switch staging

# Create new environment
coffee  env create testing --template development

# Delete environment
coffee  env delete testing --force
```

### Health Monitoring

```bash
# Health operations
coffee  health <command> [flags]

Commands:
  check    Run health checks
  monitor  Continuous monitoring
  alerts   Manage health alerts
```

**Examples:**
```bash
# Run health checks
coffee  health check --env production --timeout 30s

# Monitor continuously
coffee  health monitor --env production --interval 30s

# Manage alerts
coffee  health alerts
```

## ⚙️ Configuration

### Global Configuration

Create `~/.coffee/config.yaml`:

```yaml
# Global CLI configuration
log_level: info
default_environment: development

# Service endpoints
services:
  default_port: 8080
  health_check_path: /health
  metrics_path: /metrics
  database:
    host: localhost
    port: 5432
    database: go_coffee
    username: go_coffee_user
    password: go_coffee_password
    ssl_mode: disable

# Cloud providers
cloud:
  provider: gcp
  region: us-central1
  project: go-coffee-project

# Monitoring
telemetry:
  enabled: true
  endpoint: http://localhost:4317
  service_name: coffee
```

### Environment-Specific Configuration

Each environment can have its own configuration in `configs/<environment>/`:

```
configs/
├── development/
│   ├── app.yaml
│   ├── database.yaml
│   ├── monitoring.yaml
│   └── secrets.yaml
├── staging/
│   └── ...
└── production/
    └── ...
```

## 🎯 Best Practices

### 1. Environment Management

- Always use environment-specific configurations
- Test deployments in staging before production
- Use dry-run mode for production deployments
- Maintain separate databases per environment

### 2. Deployment Strategy

```bash
# Recommended production deployment flow
coffee  backup create --env production --type full
coffee  deploy --dry-run --env production
coffee  deploy --env production
coffee  status --env production
coffee  health check --env production
```

### 3. Database Management

```bash
# Safe migration workflow
coffee  migrate status --env production
coffee  backup create --env production --type database
coffee  migrate up --env production --backup
coffee  migrate status --env production
```

### 4. Monitoring and Alerting

- Set up continuous health monitoring
- Configure alerts for critical services
- Monitor resource usage and performance
- Implement log aggregation and analysis

### 5. Security

- Use encrypted backups for production
- Implement proper access controls
- Regular security scans
- Keep secrets in secure storage

## 🔧 Troubleshooting

### Common Issues

#### 1. Connection Errors

```bash
# Check service connectivity
coffee  health check --service api-gateway --timeout 10s

# Verify configuration
coffee config validate --env production
```

#### 2. Deployment Failures

```bash
# Check deployment logs
coffee  logs --grep "ERROR" --since 1h

# Rollback if needed
coffee  deploy --rollback --env staging
```

#### 3. Database Issues

```bash
# Check database connectivity
coffee  migrate status --env production

# Verify migrations
coffee  migrate status --format json
```

#### 4. Performance Issues

```bash
# Run performance tests
coffee  test --type performance --env staging

# Monitor resource usage
coffee  status --watch --interval 5s
```

### Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
coffee --log-level debug  status --env production
```

### Getting Help

```bash
# General help
coffee --help

# Command-specific help
coffee  deploy --help

# Version information
coffee version
```

## 📚 Additional Resources

- [Architecture Guide](./CLEAN_ARCHITECTURE_GUIDE.md)
- [API Reference](./api-reference.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Security Guide](./security.md)
- [Performance Guide](./performance.md)

## 🤝 Support

For support and questions:

- 📧 Email: support@gocoffee.dev
- 💬 Discord: [Go Coffee Community](https://discord.gg/gocoffee)
- 🐛 Issues: [GitHub Issues](https://github.com/DimaJoyti/go-coffee/issues)
- 📖 Documentation: [Official Docs](https://docs.gocoffee.dev)

---

**Happy  with Go Coffee! ☕️**
