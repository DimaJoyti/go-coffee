# Cryptocurrency Automation Platform - Production Deployment Guide

## 🚀 Overview

This guide provides comprehensive instructions for deploying the Cryptocurrency Automation Platform to a production Kubernetes environment. The platform includes advanced features like MEV protection, flash loan arbitrage, cross-chain arbitrage, AI-powered risk management, and market volatility analysis.

## 📋 Prerequisites

### Infrastructure Requirements

- **Kubernetes Cluster**: v1.24+ with at least 3 worker nodes
- **Node Resources**: Minimum 4 CPU cores and 8GB RAM per node
- **Storage**: 500GB+ SSD storage with dynamic provisioning
- **Network**: Load balancer support for ingress
- **SSL/TLS**: cert-manager for automatic certificate management

### Required Tools

- `kubectl` v1.24+
- `helm` v3.8+ (optional, for package management)
- `docker` (for building custom images)
- Access to a container registry (Docker Hub, ECR, GCR, etc.)

### External Services

- **Blockchain RPC Endpoints**: Infura, Alchemy, or self-hosted nodes
- **Flashbots**: Private key for MEV protection
- **DNS**: Domain names for API and monitoring endpoints
- **Monitoring**: Optional external monitoring services

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Load Balancer / Ingress                 │
├─────────────────────────────────────────────────────────────┤
│  api.crypto-automation.com  │  monitoring.crypto-automation.com │
└─────────────────┬───────────────────────────┬───────────────┘
                  │                           │
┌─────────────────▼─────────────────┐ ┌───────▼─────────────────┐
│     Crypto Automation API         │ │    Monitoring Stack    │
│  ┌─────────────────────────────┐   │ │  ┌─────────────────┐   │
│  │  MEV Protection Engine      │   │ │  │   Prometheus    │   │
│  │  Flash Loan Arbitrage       │   │ │  │   Grafana       │   │
│  │  Cross-Chain Arbitrage      │   │ │  │   AlertManager  │   │
│  │  AI Risk Management         │   │ │  └─────────────────┘   │
│  │  Market Volatility Analysis │   │ │                        │
│  │  Hardware Wallet Support    │   │ │                        │
│  └─────────────────────────────┘   │ │                        │
└─────────────────┬─────────────────┘ └────────────────────────┘
                  │
┌─────────────────▼─────────────────┐
│         Data Layer               │
│  ┌─────────────┐ ┌─────────────┐  │
│  │ PostgreSQL  │ │    Redis    │  │
│  │ (Primary DB)│ │   (Cache)   │  │
│  └─────────────┘ └─────────────┘  │
└───────────────────────────────────┘
```

## 🔧 Configuration

### 1. Environment Variables

Create a `.env` file with your configuration:

```bash
# Database Configuration
DATABASE_PASSWORD=your_secure_database_password
REDIS_PASSWORD=your_secure_redis_password

# JWT Configuration
JWT_SECRET_KEY=your_jwt_secret_key_min_32_chars

# Blockchain Configuration
FLASHBOTS_PRIVATE_KEY=0x1234567890abcdef...
INFURA_API_KEY=your_infura_api_key
ALCHEMY_API_KEY=your_alchemy_api_key

# Domain Configuration
API_DOMAIN=api.crypto-automation.com
MONITORING_DOMAIN=monitoring.crypto-automation.com
```

### 2. Kubernetes Secrets

The deployment script will prompt for sensitive information and create Kubernetes secrets automatically. Alternatively, you can create them manually:

```bash
kubectl create secret generic crypto-platform-secrets \
  --from-literal=database-password="$DATABASE_PASSWORD" \
  --from-literal=redis-password="$REDIS_PASSWORD" \
  --from-literal=jwt-secret-key="$JWT_SECRET_KEY" \
  --from-literal=flashbots-private-key="$FLASHBOTS_PRIVATE_KEY" \
  --from-literal=infura-api-key="$INFURA_API_KEY" \
  --from-literal=alchemy-api-key="$ALCHEMY_API_KEY" \
  -n crypto-automation
```

## 🚀 Deployment Steps

### Quick Deployment

For a complete automated deployment:

```bash
cd crypto-wallet/deployment/scripts
./deploy-production.sh deploy
```

### Manual Step-by-Step Deployment

#### 1. Create Namespace

```bash
kubectl create namespace crypto-automation
kubectl label namespace crypto-automation app=crypto-automation-platform environment=production
```

#### 2. Deploy Infrastructure

```bash
kubectl apply -f deployment/production/infrastructure.yaml
```

Wait for infrastructure components to be ready:

```bash
kubectl wait --for=condition=ready pod -l app=postgres -n crypto-automation --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n crypto-automation --timeout=300s
```

#### 3. Deploy Main Application

```bash
kubectl apply -f deployment/production/crypto-automation-platform.yaml
```

Wait for application to be ready:

```bash
kubectl wait --for=condition=available deployment/crypto-automation-api -n crypto-automation --timeout=300s
```

#### 4. Deploy Ingress and Monitoring

```bash
kubectl apply -f deployment/production/ingress-monitoring.yaml
```

#### 5. Verify Deployment

```bash
kubectl get all -n crypto-automation
kubectl get ingress -n crypto-automation
```

## 📊 Monitoring and Observability

### Grafana Dashboard

Access Grafana at `https://monitoring.crypto-automation.com/grafana`

- **Username**: admin
- **Password**: crypto_grafana_admin123 (change after first login)

### Key Metrics to Monitor

1. **API Performance**
   - Request rate and response times
   - Error rates (4xx, 5xx)
   - Active connections

2. **Arbitrage Operations**
   - Active opportunities detected
   - Successful executions
   - Profit/loss tracking
   - Bridge transaction status

3. **AI Risk Management**
   - Risk scores distribution
   - Alert frequency
   - Model performance metrics

4. **Infrastructure Health**
   - Database connections and performance
   - Redis cache hit rates
   - Resource utilization (CPU, memory, disk)

### Prometheus Alerts

The deployment includes pre-configured alerts for:

- High error rates (>5% for 2 minutes)
- High latency (95th percentile >500ms for 5 minutes)
- Database connection failures
- Redis connection failures
- Pod restart loops

## 🔒 Security Considerations

### Network Security

- **Network Policies**: Restrict pod-to-pod communication
- **TLS Termination**: All external traffic encrypted
- **Private Keys**: Stored in Kubernetes secrets
- **RBAC**: Minimal required permissions

### Application Security

- **Rate Limiting**: 100 requests/minute per IP
- **CORS**: Restricted to allowed origins
- **JWT Authentication**: Secure token-based auth
- **Input Validation**: All inputs validated and sanitized

### Operational Security

- **Secret Rotation**: Regular rotation of API keys and passwords
- **Audit Logging**: All administrative actions logged
- **Backup Strategy**: Regular database and configuration backups
- **Incident Response**: Documented procedures for security incidents

## 🔄 Scaling and High Availability

### Horizontal Pod Autoscaler

The deployment includes HPA configuration:

- **Min Replicas**: 3
- **Max Replicas**: 10
- **CPU Target**: 70%
- **Memory Target**: 80%
- **Custom Metrics**: HTTP requests per second

### Database High Availability

For production, consider:

- **PostgreSQL Cluster**: Use PostgreSQL Operator for HA
- **Read Replicas**: Separate read and write workloads
- **Backup Strategy**: Automated daily backups with point-in-time recovery

### Redis High Availability

- **Redis Sentinel**: For automatic failover
- **Redis Cluster**: For horizontal scaling
- **Persistence**: Both RDB and AOF enabled

## 🛠️ Maintenance and Operations

### Regular Maintenance Tasks

1. **Update Dependencies**: Monthly security updates
2. **Monitor Logs**: Daily log review for errors
3. **Performance Review**: Weekly performance analysis
4. **Backup Verification**: Monthly backup restore tests
5. **Security Audit**: Quarterly security reviews

### Troubleshooting Common Issues

#### Pod Startup Issues

```bash
# Check pod status
kubectl get pods -n crypto-automation

# View pod logs
kubectl logs -f deployment/crypto-automation-api -n crypto-automation

# Describe pod for events
kubectl describe pod <pod-name> -n crypto-automation
```

#### Database Connection Issues

```bash
# Test database connectivity
kubectl exec -it deployment/crypto-automation-api -n crypto-automation -- \
  psql -h postgres-service -U crypto_user -d crypto_automation -c "SELECT 1;"
```

#### Performance Issues

```bash
# Check resource usage
kubectl top pods -n crypto-automation
kubectl top nodes

# View HPA status
kubectl get hpa -n crypto-automation
```

## 📈 Performance Tuning

### Application Tuning

- **Connection Pooling**: Optimize database connection pools
- **Caching Strategy**: Implement Redis caching for frequently accessed data
- **Goroutine Management**: Monitor and optimize concurrent operations
- **Memory Management**: Regular garbage collection tuning

### Infrastructure Tuning

- **Node Affinity**: Place workloads on appropriate node types
- **Resource Requests/Limits**: Fine-tune based on actual usage
- **Storage Performance**: Use high-performance SSD storage
- **Network Optimization**: Optimize service mesh configuration

## 🔄 Backup and Disaster Recovery

### Backup Strategy

1. **Database Backups**: Daily automated backups with 30-day retention
2. **Configuration Backups**: Version-controlled Kubernetes manifests
3. **Secret Backups**: Encrypted backup of sensitive data
4. **Application State**: Regular snapshots of application state

### Disaster Recovery Plan

1. **RTO (Recovery Time Objective)**: 4 hours
2. **RPO (Recovery Point Objective)**: 1 hour
3. **Backup Verification**: Monthly restore tests
4. **Failover Procedures**: Documented step-by-step procedures

## 📞 Support and Contact

For deployment issues or questions:

- **Documentation**: Check this README and inline comments
- **Logs**: Review application and infrastructure logs
- **Monitoring**: Check Grafana dashboards for system health
- **Community**: Join our Discord/Slack for community support

## 📝 License

This deployment configuration is part of the Cryptocurrency Automation Platform and is subject to the same license terms.
