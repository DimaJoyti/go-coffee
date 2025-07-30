# 📚 Go Coffee CI/CD Deployment Runbook

## 🎯 Overview

This runbook provides comprehensive guidance for deploying and managing the Go Coffee platform using the CI/CD pipeline. It covers all aspects from initial setup to production deployment and troubleshooting.

## 🏗️ Architecture Overview

The Go Coffee platform consists of:
- **19+ Microservices**: Core business logic, AI services, and infrastructure components
- **Multi-Environment Setup**: Staging, Production, and Disaster Recovery
- **GitOps Deployment**: ArgoCD for automated and manual deployments
- **Comprehensive Monitoring**: Prometheus, Grafana, Jaeger, and AlertManager
- **Security Scanning**: SAST, DAST, container scanning, and compliance checks

## 🚀 Quick Start Deployment

### Prerequisites

1. **Required Tools**:
   ```bash
   # Install required tools
   kubectl version --client
   helm version
   docker version
   git version
   ```

2. **Cluster Access**:
   ```bash
   # Verify cluster access
   kubectl cluster-info
   kubectl get nodes
   ```

3. **Repository Access**:
   ```bash
   # Clone repository
   git clone https://github.com/DimaJoyti/go-coffee.git
   cd go-coffee
   ```

### Initial Setup

1. **Deploy CI/CD Infrastructure**:
   ```bash
   # Deploy ArgoCD and CI/CD stack
   chmod +x ci-cd/deploy-cicd-stack.sh
   ./ci-cd/deploy-cicd-stack.sh deploy
   ```

2. **Setup Environments**:
   ```bash
   # Create staging and production environments
   chmod +x ci-cd/environments/setup-environments.sh
   ./ci-cd/environments/setup-environments.sh setup
   ```

3. **Deploy Monitoring Stack**:
   ```bash
   # Deploy monitoring and observability
   chmod +x ci-cd/monitoring/deploy-monitoring-stack.sh
   ./ci-cd/monitoring/deploy-monitoring-stack.sh deploy
   ```

4. **Generate Kubernetes Manifests**:
   ```bash
   # Generate deployment manifests
   chmod +x ci-cd/kubernetes/generate-manifests.sh
   ./ci-cd/kubernetes/generate-manifests.sh generate
   ```

5. **Generate Docker Images**:
   ```bash
   # Generate Dockerfiles and build images
   chmod +x ci-cd/docker/generate-dockerfiles.sh
   ./ci-cd/docker/generate-dockerfiles.sh generate
   ./build-all-images.sh
   ```

## 🔄 Deployment Workflows

### Staging Deployment (Automatic)

**Trigger**: Push to `develop` branch

**Process**:
1. Code pushed to `develop` branch
2. GitHub Actions triggers build and test workflow
3. Security scanning and quality checks
4. Docker images built and pushed to registry
5. ArgoCD automatically syncs staging environment
6. Health checks and smoke tests executed

**Monitoring**:
```bash
# Check staging deployment status
kubectl get pods -n go-coffee-staging
kubectl get applications -n argocd | grep staging

# View deployment logs
kubectl logs -l app=api-gateway -n go-coffee-staging -f
```

### Production Deployment (Manual)

**Trigger**: Manual workflow dispatch from GitHub Actions

**Process**:
1. Select version/commit to deploy
2. Manual approval required
3. Pre-deployment health checks
4. Blue-green deployment strategy
5. Gradual traffic switching
6. Post-deployment validation
7. Monitoring and alerting

**Commands**:
```bash
# Trigger production deployment
# Go to GitHub Actions → Deploy to Production → Run workflow

# Monitor production deployment
kubectl get pods -n go-coffee-production
argocd app get go-coffee-core-production
```

## 🔧 Service Management

### Core Services Deployment Order

1. **Infrastructure Services**:
   - PostgreSQL
   - Redis
   - Monitoring stack

2. **Authentication & Security**:
   - auth-service
   - security-gateway

3. **Core Business Services**:
   - order-service
   - payment-service
   - kitchen-service
   - user-gateway

4. **AI & ML Services**:
   - ai-service
   - ai-search
   - llm-orchestrator
   - ai-arbitrage-service

5. **Gateway Services**:
   - api-gateway (last)

### Service Health Checks

```bash
# Check all services health
for service in api-gateway auth-service order-service payment-service kitchen-service; do
  echo "Checking $service..."
  kubectl port-forward svc/$service 8080:80 -n go-coffee-production &
  PID=$!
  sleep 2
  curl -f http://localhost:8080/health || echo "❌ $service unhealthy"
  kill $PID
done
```

### Scaling Services

```bash
# Scale specific service
kubectl scale deployment api-gateway --replicas=5 -n go-coffee-production

# Auto-scaling with HPA
kubectl autoscale deployment api-gateway --cpu-percent=70 --min=2 --max=10 -n go-coffee-production
```

## 🔒 Security Operations

### Secrets Management

```bash
# View secrets (without values)
kubectl get secrets -n go-coffee-production

# Rotate secrets
./ci-cd/environments/setup-environments.sh rotate-secrets production

# Update specific secret
kubectl create secret generic go-coffee-api-keys \
  --from-literal=openai-api-key="new-key" \
  --namespace=go-coffee-production \
  --dry-run=client -o yaml | kubectl apply -f -
```

### Security Scanning

```bash
# Run comprehensive security scan
gh workflow run "🔒 Comprehensive Security Scanning" \
  --field scan_type=all \
  --field severity_threshold=HIGH

# Check security scan results
gh run list --workflow="🔒 Comprehensive Security Scanning"
```

## 📊 Monitoring & Observability

### Access Monitoring Tools

```bash
# Prometheus
kubectl port-forward svc/prometheus-stack-kube-prom-prometheus 9090:9090 -n go-coffee-monitoring

# Grafana
kubectl port-forward svc/prometheus-stack-grafana 3000:80 -n go-coffee-monitoring

# Jaeger
kubectl port-forward svc/jaeger-query 16686:16686 -n go-coffee-monitoring

# AlertManager
kubectl port-forward svc/prometheus-stack-kube-prom-alertmanager 9093:9093 -n go-coffee-monitoring
```

### Key Metrics to Monitor

1. **Application Metrics**:
   - Request rate and latency
   - Error rates
   - Service availability

2. **Infrastructure Metrics**:
   - CPU and memory usage
   - Disk and network I/O
   - Pod restart counts

3. **CI/CD Metrics**:
   - Deployment frequency
   - Lead time
   - Change failure rate
   - Mean time to recovery

### Alert Management

```bash
# Check active alerts
curl -s http://localhost:9093/api/v1/alerts | jq '.data[] | select(.status.state=="active")'

# Silence alert
curl -X POST http://localhost:9093/api/v1/silences \
  -H "Content-Type: application/json" \
  -d '{"matchers":[{"name":"alertname","value":"HighErrorRate"}],"startsAt":"2024-01-01T00:00:00Z","endsAt":"2024-01-01T01:00:00Z","comment":"Maintenance window"}'
```

## 🔄 Rollback Procedures

### Automatic Rollback

ArgoCD automatically rolls back on:
- Health check failures
- Deployment timeouts
- Critical alert triggers

### Manual Rollback

```bash
# Rollback using ArgoCD
argocd app rollback go-coffee-core-production

# Rollback using kubectl
kubectl rollout undo deployment/api-gateway -n go-coffee-production

# Rollback to specific revision
kubectl rollout undo deployment/api-gateway --to-revision=2 -n go-coffee-production
```

### Blue-Green Rollback

```bash
# Switch traffic back to blue environment
kubectl patch service api-gateway -n go-coffee-production \
  -p '{"spec":{"selector":{"version":"blue"}}}'
```

## 🛠️ Troubleshooting Guide

### Common Issues

1. **Pod Stuck in Pending**:
   ```bash
   # Check resource constraints
   kubectl describe pod <pod-name> -n <namespace>
   kubectl get events -n <namespace> --sort-by='.lastTimestamp'
   ```

2. **Service Unavailable**:
   ```bash
   # Check service endpoints
   kubectl get endpoints <service-name> -n <namespace>
   kubectl describe service <service-name> -n <namespace>
   ```

3. **ArgoCD Sync Issues**:
   ```bash
   # Check application status
   argocd app get <app-name>
   argocd app diff <app-name>
   argocd app sync <app-name> --force
   ```

4. **High Memory Usage**:
   ```bash
   # Check memory usage
   kubectl top pods -n <namespace>
   kubectl describe pod <pod-name> -n <namespace>
   ```

### Debugging Commands

```bash
# Get pod logs
kubectl logs <pod-name> -n <namespace> -f

# Execute into pod
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh

# Check resource usage
kubectl top nodes
kubectl top pods --all-namespaces

# Network debugging
kubectl run debug --image=nicolaka/netshoot -it --rm -- /bin/bash
```

## 📋 Maintenance Procedures

### Regular Maintenance Tasks

1. **Weekly**:
   - Review security scan results
   - Check resource utilization
   - Validate backup procedures
   - Update dependencies

2. **Monthly**:
   - Rotate secrets
   - Review and update monitoring alerts
   - Performance optimization
   - Disaster recovery testing

3. **Quarterly**:
   - Security audit
   - Capacity planning
   - Update CI/CD tools
   - Documentation review

### Backup Procedures

```bash
# Backup ArgoCD configuration
kubectl get applications -n argocd -o yaml > argocd-apps-backup.yaml

# Backup secrets
kubectl get secrets --all-namespaces -o yaml > secrets-backup.yaml

# Backup persistent volumes
kubectl get pv -o yaml > pv-backup.yaml
```

## 🚨 Emergency Procedures

### Service Outage Response

1. **Immediate Actions**:
   - Check monitoring dashboards
   - Review recent deployments
   - Check infrastructure status
   - Activate incident response team

2. **Investigation**:
   - Analyze logs and metrics
   - Check external dependencies
   - Review recent changes
   - Identify root cause

3. **Resolution**:
   - Implement fix or rollback
   - Monitor recovery
   - Communicate status
   - Document incident

### Disaster Recovery

```bash
# Switch to DR environment
kubectl config use-context go-coffee-dr

# Verify DR environment
kubectl get pods --all-namespaces
kubectl get services --all-namespaces

# Update DNS to point to DR
# (This would be done through your DNS provider)
```

## 📞 Support Contacts

- **DevOps Team**: devops@gocoffee.dev
- **Security Team**: security@gocoffee.dev
- **On-Call Engineer**: +1-555-COFFEE-1
- **Slack Channels**: 
  - #go-coffee-alerts (critical)
  - #go-coffee-deployments (deployments)
  - #go-coffee-support (general support)

## 📚 Additional Resources

- [Architecture Documentation](./ARCHITECTURE.md)
- [Security Guidelines](./SECURITY.md)
- [Monitoring Guide](./MONITORING.md)
- [API Documentation](./API.md)
- [Troubleshooting FAQ](./TROUBLESHOOTING.md)

---

**Last Updated**: 2024-01-30
**Version**: 1.0.0
**Maintained By**: Go Coffee DevOps Team
