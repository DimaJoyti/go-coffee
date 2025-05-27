# 🚀 CI/CD Pipeline Documentation

## Overview

The CI/CD pipeline for the Web3 Coffee Platform is designed to handle multiple services across different architectures, from legacy Kafka-based services to modern Web3 wallet backend services with AI integration.

## Pipeline Architecture

### 🔄 Workflow Triggers
- **Push to main**: Full build, test, and deployment
- **Pull Request**: Build, test, security scan, and quality checks

### 🏗️ Pipeline Jobs

#### 1. Build and Test (`build-and-test`)
Builds and tests all services in the platform:

**Legacy Services:**
- Producer service (Kafka message producer)
- Consumer service (Kafka message consumer) 
- Streams service (Kafka Streams processing)

**Web3 Wallet Backend:**
- All microservices (wallet, API gateway, DeFi, etc.)
- Telegram bot with AI integration
- Comprehensive test suite

**Accounts Service:**
- GraphQL-based accounts management

#### 2. Build and Push Images (`build-and-push-images`)
Uses matrix strategy to build Docker images for all services:

**Images Built:**
- `coffee-producer` - Legacy producer service
- `coffee-consumer` - Legacy consumer service  
- `coffee-streams` - Legacy streams service
- `web3-coffee-telegram-bot` - Telegram bot with AI
- `web3-coffee-wallet-service` - Wallet management service
- `web3-coffee-api-gateway` - API gateway
- `web3-coffee-defi-service` - DeFi integration service
- `coffee-accounts-service` - Accounts management

**Registry:** GitHub Container Registry (ghcr.io)

#### 3. Deploy (`deploy`)
Multi-stage deployment process:

1. **Legacy Services**: Deployed using Helm charts
2. **Web3 Wallet Backend**: Deployed using Kubernetes manifests
3. **Telegram Bot**: Docker Compose configuration updated
4. **Accounts Service**: Kubernetes deployment

#### 4. Security Scan (`security-scan`)
Runs on pull requests only:
- **Trivy**: Vulnerability scanning for dependencies and containers
- **Gosec**: Go security analyzer for source code
- Results uploaded to GitHub Security tab

#### 5. Code Quality (`code-quality`)
Ensures code quality standards:
- **golangci-lint**: Comprehensive Go linting
- **gofmt**: Code formatting verification
- **go vet**: Static analysis
- **go mod tidy**: Dependency management verification

#### 6. Integration Tests (`integration-tests`)
Runs on pull requests with real services:
- **Redis**: For caching and session management
- **PostgreSQL**: For data persistence
- **Full integration test suite**

## 🔧 Configuration

### Required Secrets
```yaml
KUBECONFIG: Base64 encoded Kubernetes config for deployment
GITHUB_TOKEN: Automatically provided by GitHub Actions
```

### Environment Variables
```yaml
REGISTRY: ghcr.io
REGISTRY_USERNAME: ${{ github.actor }}
REGISTRY_PASSWORD: ${{ secrets.GITHUB_TOKEN }}
REGISTRY_NAMESPACE: ${{ github.repository_owner }}
KUBERNETES_NAMESPACE: coffee-system
```

## 📦 Service Matrix

| Service | Context | Dockerfile | Image Name |
|---------|---------|------------|------------|
| Producer | `./producer` | `./producer/Dockerfile` | `coffee-producer` |
| Consumer | `./consumer` | `./consumer/Dockerfile` | `coffee-consumer` |
| Streams | `./streams` | `./streams/Dockerfile` | `coffee-streams` |
| Telegram Bot | `./web3-wallet-backend` | `./web3-wallet-backend/deployments/telegram-bot/Dockerfile` | `web3-coffee-telegram-bot` |
| Wallet Service | `./web3-wallet-backend` | `./web3-wallet-backend/build/wallet-service/Dockerfile` | `web3-coffee-wallet-service` |
| API Gateway | `./web3-wallet-backend` | `./web3-wallet-backend/build/api-gateway/Dockerfile` | `web3-coffee-api-gateway` |
| DeFi Service | `./web3-wallet-backend` | `./web3-wallet-backend/build/defi-service/Dockerfile` | `web3-coffee-defi-service` |
| Accounts Service | `./accounts-service` | `./accounts-service/Dockerfile` | `coffee-accounts-service` |

## 🚀 Deployment Strategy

### Development
- All services run locally using Docker Compose
- Hot reloading for rapid development
- Local Redis and PostgreSQL instances

### Staging
- Kubernetes deployment with staging configuration
- Integration with external services (limited)
- Automated testing and validation

### Production
- Multi-region Kubernetes deployment
- High availability configuration
- Full monitoring and alerting
- Blue-green deployment strategy

## 📊 Monitoring and Observability

### Metrics Collection
- **Prometheus**: Metrics scraping and storage
- **Grafana**: Visualization and dashboards
- **Custom metrics**: Business and technical KPIs

### Logging
- **Structured logging**: JSON format with correlation IDs
- **Centralized collection**: ELK stack or similar
- **Log levels**: Debug, Info, Warn, Error

### Alerting
- **Critical alerts**: Service down, high error rates
- **Warning alerts**: Performance degradation, resource usage
- **Business alerts**: Transaction failures, security events

## 🔒 Security

### Container Security
- **Base images**: Minimal, regularly updated
- **Vulnerability scanning**: Trivy integration
- **Non-root users**: All containers run as non-root
- **Secret management**: Kubernetes secrets

### Code Security
- **Static analysis**: Gosec for Go security issues
- **Dependency scanning**: Automated vulnerability detection
- **Code review**: Required for all changes
- **SAST/DAST**: Integrated security testing

## 🧪 Testing Strategy

### Unit Tests
- **Coverage**: Minimum 80% code coverage
- **Fast execution**: Under 30 seconds total
- **Isolated**: No external dependencies

### Integration Tests
- **Real services**: Redis, PostgreSQL, Kafka
- **API testing**: End-to-end API workflows
- **Contract testing**: Service interface validation

### Performance Tests
- **Load testing**: Simulated user traffic
- **Stress testing**: Breaking point identification
- **Benchmark tests**: Performance regression detection

## 🔄 Rollback Strategy

### Automatic Rollback
- **Health checks**: Failed deployments auto-rollback
- **Monitoring**: Alert-triggered rollbacks
- **Circuit breakers**: Service protection

### Manual Rollback
```bash
# Rollback to previous version
kubectl rollout undo deployment/service-name -n coffee-system

# Rollback to specific revision
kubectl rollout undo deployment/service-name --to-revision=2 -n coffee-system
```

## 📈 Performance Optimization

### Build Optimization
- **Multi-stage builds**: Minimal production images
- **Layer caching**: GitHub Actions cache
- **Parallel builds**: Matrix strategy for speed

### Deployment Optimization
- **Rolling updates**: Zero-downtime deployments
- **Resource limits**: Proper CPU/memory allocation
- **Horizontal scaling**: Auto-scaling based on metrics

## 🛠️ Troubleshooting

### Common Issues

1. **Build Failures**
   - Check Go version compatibility
   - Verify dependency versions
   - Review build logs

2. **Test Failures**
   - Ensure test services are healthy
   - Check environment variables
   - Review test isolation

3. **Deployment Issues**
   - Verify Kubernetes connectivity
   - Check resource availability
   - Review deployment logs

### Debug Commands
```bash
# Check pipeline status
gh run list --workflow=ci-cd.yaml

# View specific run logs
gh run view <run-id> --log

# Check deployment status
kubectl get deployments -n coffee-system

# View pod logs
kubectl logs -f deployment/service-name -n coffee-system
```

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Helm Documentation](https://helm.sh/docs/)
