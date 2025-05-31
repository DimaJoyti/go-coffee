# 🔧 Go Coffee Environment Setup Guide

This guide explains how to set up and manage environment files for the Go Coffee project.

## 📁 Environment Files Overview

The Go Coffee project uses multiple environment files for different purposes:

| File | Purpose | When to Use |
|------|---------|-------------|
| `.env.example` | Template file with all variables | Reference and initial setup |
| `.env` | Main environment file | Local development |
| `.env.local` | Local overrides | Personal development settings |
| `.env.development` | Development environment | Development deployments |
| `.env.production` | Production environment | Production deployments |
| `.env.docker` | Docker Compose | Container deployments |
| `.env.ai-search` | AI Search Engine specific | AI service configuration |
| `.env.web3` | Web3 services specific | Blockchain service configuration |

## 🚀 Quick Start

### Option 1: Automated Setup (Recommended)

Run the interactive setup script:

```bash
./scripts/setup-env.sh
```

This script will:
- Create necessary environment files
- Generate secure secrets
- Configure database and Redis settings
- Set up AI service configuration
- Validate the configuration

### Option 2: Manual Setup

1. **Copy the example file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit the .env file:**
   ```bash
   nano .env  # or your preferred editor
   ```

3. **Generate secure secrets:**
   ```bash
   make env-generate-secrets
   ```

4. **Validate configuration:**
   ```bash
   make env-validate
   ```

## 🔒 Security Configuration

### Critical Security Variables

These variables **MUST** be changed from their default values:

```bash
# JWT Configuration - CRITICAL
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
API_KEY_SECRET=your-api-key-secret-change-this
WEBHOOK_SECRET=your-webhook-secret-key-change-this
ENCRYPTION_KEY=your-32-character-encryption-key!!
```

### Generating Secure Secrets

Use the built-in secret generator:

```bash
make env-generate-secrets
```

Or generate manually:

```bash
# JWT Secret (64 characters)
openssl rand -hex 32

# API Key Secret (48 characters)
openssl rand -hex 24

# Webhook Secret (48 characters)
openssl rand -hex 24

# Encryption Key (32 characters)
openssl rand -hex 16
```

## 🗄️ Database Configuration

### PostgreSQL Setup

```bash
# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=go_coffee
DATABASE_USER=postgres
DATABASE_PASSWORD=your-secure-password
DATABASE_SSL_MODE=disable  # Use 'require' in production
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME=300s
```

### Redis Configuration

```bash
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5
REDIS_URL=redis://localhost:6379
```

## 🤖 AI Services Configuration

### OpenAI Configuration

```bash
# OpenAI
OPENAI_API_KEY=sk-your-openai-api-key
```

### Google Gemini Configuration

```bash
# Google Gemini
GEMINI_API_KEY=your-gemini-api-key
```

### Local AI (Ollama) Configuration

```bash
# Ollama (Local AI)
OLLAMA_URL=http://localhost:11434
```

### AI Search Engine Configuration

```bash
# AI Search Configuration
AI_SEARCH_EMBEDDING_MODEL=coffee_ai_v2
AI_SEARCH_VECTOR_DIMENSIONS=384
AI_SEARCH_SIMILARITY_THRESHOLD=0.7
AI_SEARCH_MAX_RESULTS=50
```

## 🌐 Web3 & Blockchain Configuration

### Ethereum Configuration

```bash
# Ethereum
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/your-project-id
ETHEREUM_TESTNET_RPC_URL=https://goerli.infura.io/v3/your-project-id
ETHEREUM_PRIVATE_KEY=your-ethereum-private-key
ETHEREUM_GAS_LIMIT=21000
ETHEREUM_GAS_PRICE=20000000000
```

### Bitcoin Configuration

```bash
# Bitcoin
BITCOIN_RPC_URL=https://your-bitcoin-node.com
BITCOIN_RPC_USERNAME=your-bitcoin-rpc-username
BITCOIN_RPC_PASSWORD=your-bitcoin-rpc-password
```

### Solana Configuration

```bash
# Solana
SOLANA_RPC_URL=https://api.mainnet-beta.solana.com
SOLANA_TESTNET_RPC_URL=https://api.testnet.solana.com
SOLANA_PRIVATE_KEY=your-solana-private-key
```

## 📊 Monitoring & Observability

### Prometheus Configuration

```bash
# Prometheus
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
PROMETHEUS_METRICS_PATH=/metrics
```

### Jaeger Tracing Configuration

```bash
# Jaeger Tracing
JAEGER_ENABLED=true
JAEGER_ENDPOINT=http://localhost:14268/api/traces
JAEGER_SAMPLER_TYPE=const
JAEGER_SAMPLER_PARAM=1
```

### Sentry Error Tracking

```bash
# Sentry Error Tracking
SENTRY_DSN=your-sentry-dsn
SENTRY_ENVIRONMENT=development
SENTRY_RELEASE=1.0.0
```

## 🔧 Environment Management Commands

### Using Makefile

```bash
# Setup environment files
make env-setup

# Validate configuration
make env-validate

# Test configuration loading
make env-test

# Show current configuration
make env-show

# Check for security issues
make env-check-secrets

# Generate secure secrets
make env-generate-secrets

# Switch environments
make env-dev      # Switch to development
make env-prod     # Switch to production (with confirmation)
make env-docker   # Switch to Docker environment

# Backup and restore
make env-backup   # Backup current environment files
make env-restore  # Restore from backup

# Clean up
make env-clean    # Remove environment files (keeps .env.example)
```

### Using Scripts

```bash
# Interactive setup
./scripts/setup-env.sh

# Generate secrets only
./scripts/setup-env.sh secrets

# Configure database only
./scripts/setup-env.sh database

# Configure Redis only
./scripts/setup-env.sh redis

# Configure AI services only
./scripts/setup-env.sh ai

# Validate configuration only
./scripts/setup-env.sh validate
```

## 🌍 Environment-Specific Configuration

### Development Environment

For local development, use `.env.development`:

```bash
# Development settings
ENVIRONMENT=development
DEBUG=true
LOG_LEVEL=debug
LOG_FORMAT=text

# Relaxed security for development
ENABLE_CORS=true
ENABLE_SWAGGER=true
ENABLE_DEBUG_ENDPOINTS=true
```

### Production Environment

For production, use `.env.production`:

```bash
# Production settings
ENVIRONMENT=production
DEBUG=false
LOG_LEVEL=warn
LOG_FORMAT=json

# Enhanced security for production
TLS_ENABLED=true
SECURITY_HEADERS_ENABLED=true
AUDIT_LOGGING_ENABLED=true
```

### Docker Environment

For Docker deployments, use `.env.docker`:

```bash
# Docker-specific settings
POSTGRES_HOST=postgres
REDIS_HOST=redis
KAFKA_BROKERS=kafka:9092

# Docker network configuration
NETWORK_NAME=go-coffee-network
```

## 🔍 Validation and Testing

### Configuration Test

Run the configuration test utility:

```bash
go run cmd/config-test/main.go
```

This will:
- Load all environment files
- Validate configuration
- Check for missing or invalid values
- Display current configuration (without secrets)
- Show service status

### Environment Validation

```bash
# Validate current environment
make env-validate

# Test configuration loading
make env-test

# Check for security issues
make env-check-secrets
```

## 🚨 Troubleshooting

### Common Issues

1. **Missing environment file:**
   ```bash
   Error: env file .env does not exist
   ```
   **Solution:** Run `make env-setup` or copy from `.env.example`

2. **Invalid configuration:**
   ```bash
   Error: JWT_SECRET must be set to a secure value
   ```
   **Solution:** Generate secure secrets with `make env-generate-secrets`

3. **Database connection failed:**
   ```bash
   Error: failed to connect to database
   ```
   **Solution:** Check database configuration and ensure PostgreSQL is running

4. **Redis connection failed:**
   ```bash
   Error: failed to connect to Redis
   ```
   **Solution:** Check Redis configuration and ensure Redis is running

### Debug Configuration

To debug configuration issues:

```bash
# Show current configuration (without secrets)
make env-show

# Show differences between environment files
make env-diff

# Export configuration to JSON for analysis
make env-export
```

## 📝 Best Practices

### Security Best Practices

1. **Never commit .env files** with real secrets to version control
2. **Use different secrets** for each environment
3. **Rotate secrets regularly** (at least every 90 days)
4. **Use strong, random passwords** for all services
5. **Enable TLS/SSL** in production environments
6. **Limit access** to environment files on servers

### Configuration Management

1. **Use environment-specific files** for different deployments
2. **Keep .env.example updated** with all required variables
3. **Document all environment variables** and their purposes
4. **Validate configuration** before deployment
5. **Backup environment files** before making changes

### Development Workflow

1. **Start with .env.example** for new developers
2. **Use .env.local** for personal overrides
3. **Test configuration changes** before committing
4. **Keep development and production configs separate**
5. **Use feature flags** to enable/disable services

## 📚 Additional Resources

- [Configuration Package Documentation](../pkg/config/README.md)
- [Docker Deployment Guide](./DOCKER_DEPLOYMENT.md)
- [Security Best Practices](./SECURITY.md)
- [Monitoring Setup Guide](./MONITORING.md)

## 🆘 Getting Help

If you encounter issues with environment setup:

1. Check this documentation
2. Run `make env-validate` to identify issues
3. Check the [troubleshooting section](#-troubleshooting)
4. Create an issue in the project repository
