# Go Coffee Platform - Infrastructure Deployment Guide

This comprehensive guide covers deploying the Go Coffee platform with the new Clean Architecture infrastructure, consolidated environment files, updated Docker Compose, and Kubernetes manifests.

## 📋 Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Environment Configuration](#environment-configuration)
- [Docker Compose Deployment](#docker-compose-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Monitoring Setup](#monitoring-setup)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)

## 🎯 Overview

The Go Coffee platform has been migrated to a Clean Architecture with:

- **Consolidated environment files** for better configuration management
- **Enhanced Docker Compose** with production-ready configurations
- **Comprehensive Kubernetes manifests** with monitoring and scaling
- **Real-time session management** with Redis backend
- **Advanced monitoring** with Prometheus, Grafana, and Jaeger
- **Event-driven architecture** with pub/sub messaging
- **Production-grade security** with JWT, encryption, and rate limiting

## 🔧 Prerequisites

### Required Tools

- **Docker** (20.10+) and **Docker Compose** (2.0+)
- **Kubernetes** cluster (1.20+) with kubectl configured
- **Helm** (3.0+) for package management
- **Git** for source code management

### System Requirements

| Environment | CPU | RAM | Storage | Network |
|-------------|-----|-----|---------|---------|
| Development | 2+ cores | 4GB+ | 10GB+ | 1Gbps |
| Staging | 4+ cores | 8GB+ | 50GB+ | 1Gbps |
| Production | 8+ cores | 16GB+ | 100GB+ | 10Gbps |

## 🌍 Environment Configuration

### Consolidated Environment Files

The platform now uses three main environment files:

```bash
.env.development    # Development with debug features and hot reload
.env.staging        # Staging environment for testing
.env.production     # Production with security hardening
```

### Key Infrastructure Variables

```bash
# Core Settings
ENVIRONMENT=production
SERVICE_NAME=go-coffee
LOG_LEVEL=info
LOG_FORMAT=json

# Database (PostgreSQL)
DB_HOST=postgres-service
DB_NAME=go_coffee
DB_USER=go_coffee_user
DB_PASSWORD=${DB_PASSWORD}  # From secrets
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10

# Redis Cache & Sessions
REDIS_HOST=redis-service
REDIS_PORT=6379
REDIS_POOL_SIZE=50
REDIS_CLUSTER_MODE=false

# Security & Authentication
JWT_SECRET_KEY=${JWT_SECRET_KEY}  # From secrets
JWT_ACCESS_TOKEN_TTL=5m
JWT_REFRESH_TOKEN_TTL=7d
AES_ENCRYPTION_KEY=${AES_ENCRYPTION_KEY}  # From secrets

# Session Management
SESSION_COOKIE_SECURE=true
SESSION_MAX_AGE=24h
SESSION_IDLE_TIMEOUT=30m

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=1000

# Event Infrastructure
EVENT_STORE_TYPE=redis
EVENT_PUBLISHER_WORKERS=10
EVENT_SUBSCRIBER_WORKERS=10

# Monitoring
METRICS_ENABLED=true
HEALTH_ENABLED=true
JAEGER_ENABLED=true
```

## 🐳 Docker Compose Deployment

### Quick Start (Development)

```bash
# Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Copy and configure environment
cp .env.example .env.development
# Edit .env.development with your settings

# Start infrastructure and services
cd docker
docker-compose up -d

# Check service health
docker-compose ps
curl http://localhost:8080/health
```

### Environment-Specific Deployment

```bash
# Development (with hot reload and debug tools)
docker-compose -f docker-compose.yml -f docker-compose.override.yml up -d

# Production (optimized and secured)
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Enhanced Services

The updated Docker Compose includes:

- **Enhanced Redis** with optimized configuration and persistence
- **PostgreSQL** with performance tuning and health checks
- **Monitoring Stack** with Prometheus, Grafana, Jaeger, AlertManager
- **Exporters** for system, database, and Redis metrics
- **Development Tools** (Adminer, Redis Commander) in override file

### Service Scaling

```bash
# Scale auth service for high availability
docker-compose up -d --scale auth-service=3

# Scale event processing workers
docker-compose up -d --scale consumer=5 --scale streams=3
```

## ☸️ Kubernetes Deployment

### Namespace and Resources

```bash
# Create namespace with resource quotas
kubectl apply -f k8s/namespace.yaml

# Verify namespace creation
kubectl get namespace go-coffee
kubectl describe namespace go-coffee
```

### Secrets Management

```bash
# Apply secrets (update with real values first)
kubectl apply -f k8s/secrets.yaml

# Verify secrets
kubectl get secrets -n go-coffee
```

### Configuration Management

```bash
# Apply configuration maps
kubectl apply -f k8s/configmap.yaml

# Verify configuration
kubectl get configmaps -n go-coffee
kubectl describe configmap go-coffee-config -n go-coffee
```

### Infrastructure Deployment

```bash
# Deploy PostgreSQL with persistence
kubectl apply -f k8s/postgres.yaml

# Deploy Redis with clustering support
kubectl apply -f k8s/redis.yaml

# Wait for infrastructure to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n go-coffee --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n go-coffee --timeout=300s
```

### Application Deployment

```bash
# Deploy Go Coffee services
kubectl apply -f k8s/go-coffee-services.yaml

# Wait for applications to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=go-coffee -n go-coffee --timeout=600s

# Check deployment status
kubectl get pods -n go-coffee
kubectl get services -n go-coffee
```

### Horizontal Pod Autoscaling

```bash
# Enable HPA for auth service
kubectl autoscale deployment auth-service \
  --cpu-percent=70 \
  --min=2 \
  --max=10 \
  -n go-coffee

# Check HPA status
kubectl get hpa -n go-coffee
kubectl describe hpa auth-service -n go-coffee
```

## 📊 Monitoring Setup

### Docker Compose Monitoring

```bash
# Start complete monitoring stack
docker-compose up -d prometheus grafana jaeger alertmanager node-exporter redis-exporter postgres-exporter

# Access monitoring interfaces
echo "Prometheus: http://localhost:9090"
echo "Grafana: http://localhost:3000 (admin/admin)"
echo "Jaeger: http://localhost:16686"
echo "AlertManager: http://localhost:9093"
```

### Kubernetes Monitoring

```bash
# Deploy monitoring stack
kubectl apply -f k8s/monitoring.yaml

# Check monitoring services
kubectl get pods -n go-coffee-monitoring
kubectl get services -n go-coffee-monitoring
```

### Pre-configured Dashboards

The monitoring setup includes dashboards for:

- **Infrastructure Overview**: CPU, memory, disk, network metrics
- **Application Performance**: Request rates, response times, error rates
- **Database Monitoring**: PostgreSQL connections, queries, performance
- **Cache Performance**: Redis memory usage, hit rates, operations
- **Business Metrics**: Order rates, payment success, user activity
- **Security Monitoring**: Failed logins, rate limiting, JWT validation

### Alerting Rules

Comprehensive alerting for:

- **Critical**: Service down, high error rates, payment failures
- **Warning**: High resource usage, slow queries, low order rates
- **Security**: Suspicious login activity, rate limit violations
- **Business**: Low order volume, payment processing issues

## 🚀 Production Deployment

### Pre-Deployment Checklist

- [ ] SSL certificates configured and valid
- [ ] Database backups enabled and tested
- [ ] Monitoring and alerting configured
- [ ] Load balancer configured with health checks
- [ ] DNS records updated and propagated
- [ ] Security scanning completed
- [ ] Performance testing completed
- [ ] Disaster recovery plan documented

### Production Security Configuration

```bash
# Use secrets management (Vault, K8s secrets, etc.)
JWT_SECRET_KEY=${VAULT_JWT_SECRET}
AES_ENCRYPTION_KEY=${VAULT_AES_KEY}
DB_PASSWORD=${VAULT_DB_PASSWORD}

# Production security settings
SESSION_COOKIE_SECURE=true
SESSION_COOKIE_DOMAIN=.go-coffee.com
CORS_ALLOWED_ORIGINS=https://app.go-coffee.com,https://admin.go-coffee.com

# Rate limiting for production
RATE_LIMIT_REQUESTS_PER_MINUTE=1000
RATE_LIMIT_BURST_SIZE=50

# Monitoring with sampling
JAEGER_SAMPLER_PARAM=0.01  # 1% sampling
METRICS_INTERVAL=10s
```

### Production Deployment Steps

```bash
# 1. Prepare environment
export ENVIRONMENT=production
export NAMESPACE=go-coffee

# 2. Deploy infrastructure
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/configmap.yaml

# 3. Deploy data layer
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml

# 4. Wait for data layer
kubectl wait --for=condition=ready pod -l app=postgres -n go-coffee --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n go-coffee --timeout=300s

# 5. Deploy applications
kubectl apply -f k8s/go-coffee-services.yaml

# 6. Deploy monitoring
kubectl apply -f k8s/monitoring.yaml

# 7. Configure ingress/load balancer
kubectl apply -f k8s/ingress.yaml

# 8. Verify deployment
kubectl get all -n go-coffee
curl https://api.go-coffee.com/health
```

### Production Monitoring

```bash
# Monitor cluster health
kubectl get nodes
kubectl top nodes
kubectl top pods -n go-coffee

# Check application metrics
curl https://api.go-coffee.com/metrics

# Monitor logs
kubectl logs -f deployment/auth-service -n go-coffee --tail=100
```

## 🔧 Troubleshooting

### Common Issues

#### Service Health Check Failures

```bash
# Check pod status and events
kubectl describe pod <pod-name> -n go-coffee
kubectl get events -n go-coffee --sort-by='.lastTimestamp'

# Check service logs
kubectl logs <pod-name> -n go-coffee --previous

# Test connectivity
kubectl exec -it <pod-name> -n go-coffee -- curl localhost:8080/health
```

#### Database Connection Issues

```bash
# Test database connectivity
kubectl exec -it deployment/auth-service -n go-coffee -- \
  psql -h postgres-service -U go_coffee_user -d go_coffee -c "SELECT 1;"

# Check database logs
kubectl logs deployment/postgres -n go-coffee

# Verify database configuration
kubectl get configmap postgres-config -n go-coffee -o yaml
```

#### Redis Connection Issues

```bash
# Test Redis connectivity
kubectl exec -it deployment/auth-service -n go-coffee -- \
  redis-cli -h redis-service ping

# Check Redis logs
kubectl logs deployment/redis -n go-coffee

# Monitor Redis performance
kubectl exec -it deployment/redis -n go-coffee -- \
  redis-cli info memory
```

### Performance Debugging

```bash
# Check resource usage
kubectl top pods -n go-coffee
kubectl describe node <node-name>

# Application profiling
kubectl port-forward service/auth-service 8080:8080 -n go-coffee
curl http://localhost:8080/debug/pprof/

# Database performance
kubectl exec -it deployment/postgres -n go-coffee -- \
  psql -c "SELECT * FROM pg_stat_activity WHERE state = 'active';"
```

### Log Analysis

```bash
# Centralized logging
kubectl logs -l app.kubernetes.io/name=go-coffee -n go-coffee --tail=1000

# Error analysis
kubectl logs deployment/auth-service -n go-coffee | grep ERROR

# Performance analysis
kubectl logs deployment/auth-service -n go-coffee | grep "response_time"
```

## 📞 Support and Resources

### Documentation

- [Migration Guide](./MIGRATION_GUIDE.md) - Migrating from Gin to Clean Architecture
- [Infrastructure README](../pkg/infrastructure/README.md) - Detailed infrastructure documentation
- [Monitoring Guide](./MONITORING_GUIDE.md) - Advanced monitoring configuration
- [Security Guide](./SECURITY_GUIDE.md) - Security best practices

### Getting Help

1. Check the troubleshooting section above
2. Review application and infrastructure logs
3. Check monitoring dashboards for anomalies
4. Consult the infrastructure documentation
5. Create an issue with detailed error information and logs

### Useful Commands

```bash
# Quick health check
kubectl get pods -n go-coffee | grep -v Running

# Resource usage overview
kubectl top pods -n go-coffee --sort-by=memory

# Service endpoints
kubectl get endpoints -n go-coffee

# Configuration verification
kubectl get configmap go-coffee-config -n go-coffee -o yaml

# Secret verification (without values)
kubectl get secrets -n go-coffee
```

This deployment guide provides comprehensive instructions for deploying the enhanced Go Coffee platform with the new Clean Architecture infrastructure, consolidated environment management, and production-ready monitoring.
