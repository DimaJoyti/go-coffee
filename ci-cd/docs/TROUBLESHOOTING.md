# 🔧 Go Coffee CI/CD Troubleshooting Guide

## 🎯 Overview

This guide provides solutions to common issues encountered in the Go Coffee CI/CD pipeline, from build failures to deployment problems and runtime issues.

## 🚨 Emergency Quick Reference

### Critical Service Down
```bash
# 1. Check service status
kubectl get pods -n go-coffee-production -l app=<service-name>

# 2. Check recent deployments
kubectl rollout history deployment/<service-name> -n go-coffee-production

# 3. Quick rollback if needed
kubectl rollout undo deployment/<service-name> -n go-coffee-production

# 4. Check logs
kubectl logs -l app=<service-name> -n go-coffee-production --tail=100
```

### Complete System Outage
```bash
# 1. Check cluster health
kubectl get nodes
kubectl get pods --all-namespaces | grep -v Running

# 2. Check ArgoCD status
kubectl get pods -n argocd

# 3. Check monitoring
kubectl get pods -n go-coffee-monitoring

# 4. Switch to DR if needed
kubectl config use-context go-coffee-dr
```

## 🔨 Build & CI Issues

### GitHub Actions Workflow Failures

**Issue**: Workflow fails during build stage
```bash
# Check workflow logs in GitHub UI
# Common causes and solutions:

# 1. Go build failures
Error: "go: module not found"
Solution: 
- Check go.mod and go.sum files
- Run `go mod tidy` locally
- Ensure all dependencies are available

# 2. Docker build failures
Error: "failed to solve with frontend dockerfile.v0"
Solution:
- Check Dockerfile syntax
- Verify base image availability
- Check build context and .dockerignore

# 3. Test failures
Error: "Test suite failed"
Solution:
- Run tests locally: `go test ./...`
- Check test dependencies (database, Redis)
- Review test environment variables
```

**Issue**: Security scanning failures
```bash
# Check SARIF results in GitHub Security tab
# Common issues:

# 1. High severity vulnerabilities
Solution:
- Update dependencies: `go get -u ./...`
- Review and fix code issues
- Add exceptions for false positives

# 2. Container scanning failures
Solution:
- Update base images in Dockerfiles
- Remove unnecessary packages
- Use distroless or minimal images
```

### Docker Build Issues

**Issue**: Multi-stage build failures
```bash
# Debug locally
docker build --target builder -t debug-build .
docker run -it debug-build /bin/sh

# Common solutions:
# 1. Missing build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# 2. Incorrect COPY paths
COPY . .  # Ensure correct context

# 3. Permission issues
USER 1001:1001  # Use non-root user
```

**Issue**: Image size too large
```bash
# Analyze image layers
docker history <image-name>

# Solutions:
# 1. Use multi-stage builds
# 2. Minimize layers
# 3. Use .dockerignore
# 4. Remove unnecessary files
```

## ⚙️ Deployment Issues

### ArgoCD Sync Problems

**Issue**: Application stuck in "OutOfSync" state
```bash
# Check application status
argocd app get go-coffee-core-production

# Common causes and solutions:

# 1. Resource conflicts
argocd app diff go-coffee-core-production
# Solution: Review and fix manifest conflicts

# 2. Permission issues
kubectl auth can-i create deployments --namespace=go-coffee-production
# Solution: Check RBAC permissions

# 3. Resource quotas exceeded
kubectl describe quota -n go-coffee-production
# Solution: Increase quotas or reduce resource requests
```

**Issue**: Sync operation fails
```bash
# Force sync with options
argocd app sync go-coffee-core-production --force --replace

# If still failing, check:
# 1. Kubernetes API server connectivity
kubectl cluster-info

# 2. ArgoCD controller logs
kubectl logs -l app.kubernetes.io/name=argocd-application-controller -n argocd

# 3. Repository access
argocd repo list
```

### Kubernetes Deployment Issues

**Issue**: Pods stuck in "Pending" state
```bash
# Check pod events
kubectl describe pod <pod-name> -n <namespace>

# Common causes:
# 1. Insufficient resources
kubectl top nodes
kubectl describe nodes

# 2. Node selector constraints
kubectl get nodes --show-labels

# 3. Pod security policies
kubectl get psp
```

**Issue**: Pods in "CrashLoopBackOff"
```bash
# Check logs
kubectl logs <pod-name> -n <namespace> --previous

# Common causes:
# 1. Application startup failures
# Check environment variables and secrets

# 2. Health check failures
# Review liveness/readiness probe configuration

# 3. Resource limits
# Check memory/CPU limits vs actual usage
```

**Issue**: Service not accessible
```bash
# Check service and endpoints
kubectl get svc <service-name> -n <namespace>
kubectl get endpoints <service-name> -n <namespace>

# Test connectivity
kubectl run debug --image=nicolaka/netshoot -it --rm -- /bin/bash
# Inside debug pod:
nslookup <service-name>.<namespace>.svc.cluster.local
curl http://<service-name>.<namespace>.svc.cluster.local/health
```

## 🔐 Security & Access Issues

### Secret Management Problems

**Issue**: Secrets not found or invalid
```bash
# Check secret existence
kubectl get secrets -n <namespace>

# Verify secret content (base64 encoded)
kubectl get secret <secret-name> -n <namespace> -o yaml

# Common solutions:
# 1. Recreate secrets
./ci-cd/environments/setup-environments.sh rotate-secrets production

# 2. Check secret mounting
kubectl describe pod <pod-name> -n <namespace>
```

**Issue**: RBAC permission denied
```bash
# Check current permissions
kubectl auth can-i <verb> <resource> --namespace=<namespace>

# Check service account
kubectl get serviceaccount -n <namespace>
kubectl describe serviceaccount <sa-name> -n <namespace>

# Check role bindings
kubectl get rolebindings -n <namespace>
kubectl describe rolebinding <binding-name> -n <namespace>
```

### Network Policy Issues

**Issue**: Services cannot communicate
```bash
# Check network policies
kubectl get networkpolicies -n <namespace>

# Test connectivity
kubectl exec -it <pod-name> -n <namespace> -- nc -zv <target-service> <port>

# Temporary solution (for debugging only):
kubectl delete networkpolicy --all -n <namespace>
```

## 📊 Monitoring & Observability Issues

### Prometheus Issues

**Issue**: Metrics not being scraped
```bash
# Check Prometheus targets
kubectl port-forward svc/prometheus-stack-kube-prom-prometheus 9090:9090 -n go-coffee-monitoring
# Visit http://localhost:9090/targets

# Common solutions:
# 1. Check service annotations
kubectl get svc <service-name> -n <namespace> -o yaml | grep prometheus

# 2. Verify metrics endpoint
kubectl exec -it <pod-name> -n <namespace> -- curl localhost:9090/metrics

# 3. Check ServiceMonitor
kubectl get servicemonitor -n go-coffee-monitoring
```

**Issue**: Grafana dashboards not loading
```bash
# Check Grafana pod status
kubectl get pods -l app.kubernetes.io/name=grafana -n go-coffee-monitoring

# Check datasource configuration
kubectl logs -l app.kubernetes.io/name=grafana -n go-coffee-monitoring

# Reset Grafana admin password
kubectl patch secret prometheus-stack-grafana -n go-coffee-monitoring \
  -p '{"data":{"admin-password":"'$(echo -n "newpassword" | base64)'"}}'
```

### Jaeger Tracing Issues

**Issue**: Traces not appearing
```bash
# Check Jaeger components
kubectl get pods -l app.kubernetes.io/name=jaeger -n go-coffee-monitoring

# Verify trace collection
kubectl logs -l app.kubernetes.io/component=collector -n go-coffee-monitoring

# Check application instrumentation
# Ensure OpenTelemetry is properly configured in services
```

## 🔄 Performance Issues

### High Resource Usage

**Issue**: High CPU usage
```bash
# Check resource usage
kubectl top pods -n <namespace>
kubectl top nodes

# Identify resource-hungry pods
kubectl get pods -n <namespace> --sort-by='.status.containerStatuses[0].restartCount'

# Solutions:
# 1. Increase resource limits
# 2. Optimize application code
# 3. Scale horizontally
kubectl scale deployment <deployment-name> --replicas=5 -n <namespace>
```

**Issue**: High memory usage
```bash
# Check memory usage patterns
kubectl describe pod <pod-name> -n <namespace>

# Check for memory leaks
kubectl exec -it <pod-name> -n <namespace> -- top

# Solutions:
# 1. Increase memory limits
# 2. Implement memory profiling
# 3. Restart pods periodically
```

### Database Performance Issues

**Issue**: Slow database queries
```bash
# Check PostgreSQL performance
kubectl exec -it postgres-0 -n <namespace> -- psql -U go_coffee_user -d go_coffee

# Inside PostgreSQL:
SELECT * FROM pg_stat_activity WHERE state = 'active';
SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;

# Solutions:
# 1. Add database indexes
# 2. Optimize queries
# 3. Increase connection pool size
# 4. Scale database resources
```

**Issue**: Redis performance problems
```bash
# Check Redis metrics
kubectl exec -it redis-0 -n <namespace> -- redis-cli info

# Check slow queries
kubectl exec -it redis-0 -n <namespace> -- redis-cli slowlog get 10

# Solutions:
# 1. Optimize Redis configuration
# 2. Implement proper caching strategies
# 3. Monitor memory usage
```

## 🌐 External Dependencies

### API Integration Issues

**Issue**: External API timeouts
```bash
# Check network connectivity
kubectl exec -it <pod-name> -n <namespace> -- nslookup api.external-service.com

# Test API endpoints
kubectl exec -it <pod-name> -n <namespace> -- curl -v https://api.external-service.com/health

# Solutions:
# 1. Increase timeout values
# 2. Implement retry logic
# 3. Add circuit breakers
# 4. Check API rate limits
```

**Issue**: DNS resolution problems
```bash
# Check DNS configuration
kubectl get configmap coredns -n kube-system -o yaml

# Test DNS resolution
kubectl run debug --image=nicolaka/netshoot -it --rm -- nslookup kubernetes.default

# Solutions:
# 1. Check CoreDNS pods
kubectl get pods -n kube-system -l k8s-app=kube-dns

# 2. Restart CoreDNS if needed
kubectl rollout restart deployment/coredns -n kube-system
```

## 🔧 Recovery Procedures

### Service Recovery

```bash
# 1. Identify the problem
kubectl get pods -n <namespace> | grep -v Running

# 2. Check recent changes
kubectl rollout history deployment/<service-name> -n <namespace>

# 3. Rollback if needed
kubectl rollout undo deployment/<service-name> -n <namespace>

# 4. Scale up if needed
kubectl scale deployment <service-name> --replicas=3 -n <namespace>

# 5. Verify recovery
kubectl get pods -n <namespace> -w
```

### Data Recovery

```bash
# 1. Check backup status
kubectl get pvc -n <namespace>

# 2. Restore from backup (if available)
# This depends on your backup solution

# 3. Verify data integrity
kubectl exec -it postgres-0 -n <namespace> -- psql -U go_coffee_user -d go_coffee -c "SELECT COUNT(*) FROM orders;"
```

## 📞 Escalation Procedures

### When to Escalate

1. **Immediate Escalation**:
   - Complete service outage > 5 minutes
   - Data loss or corruption
   - Security breach detected
   - Multiple services failing

2. **Standard Escalation**:
   - Single service down > 15 minutes
   - Performance degradation > 30 minutes
   - Deployment failures > 3 attempts

### Escalation Contacts

- **Level 1**: DevOps Team (#go-coffee-support)
- **Level 2**: Senior SRE (#go-coffee-escalation)
- **Level 3**: Engineering Manager (phone: +1-555-COFFEE-1)
- **Security Issues**: Security Team (security@gocoffee.dev)

## 📋 Useful Commands Reference

```bash
# Quick health check
kubectl get pods --all-namespaces | grep -v Running

# Resource usage overview
kubectl top nodes && kubectl top pods --all-namespaces

# Recent events
kubectl get events --all-namespaces --sort-by='.lastTimestamp' | tail -20

# ArgoCD status
argocd app list

# Check all services
for ns in go-coffee-staging go-coffee-production; do
  echo "=== $ns ==="
  kubectl get pods -n $ns
done
```

---

**Remember**: When in doubt, check the logs first, then escalate if needed. Document any new issues and solutions for future reference.

**Last Updated**: 2024-01-30
**Maintained By**: Go Coffee DevOps Team
