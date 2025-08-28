# 🔧 Go Coffee Platform - Troubleshooting Guide

## 📋 Overview

This comprehensive troubleshooting guide provides systematic approaches to diagnose and resolve common issues in the Go Coffee multi-cloud platform. It includes diagnostic commands, common solutions, and escalation procedures.

## 🎯 Troubleshooting Methodology

### Systematic Approach

1. **Define the Problem**
   - What is the expected behavior?
   - What is the actual behavior?
   - When did the issue start?
   - What changed recently?

2. **Gather Information**
   - Check monitoring dashboards
   - Review recent logs
   - Examine system metrics
   - Identify affected components

3. **Form Hypothesis**
   - Based on symptoms and data
   - Consider recent changes
   - Review similar past incidents

4. **Test and Validate**
   - Apply targeted diagnostics
   - Test hypothesis systematically
   - Document findings

5. **Implement Solution**
   - Apply fix with minimal impact
   - Monitor for improvement
   - Verify complete resolution

6. **Document and Learn**
   - Update runbooks
   - Share knowledge with team
   - Implement preventive measures

---

## 🚨 Common Issues and Solutions

### Application Issues

#### **Issue: Service Returning 5xx Errors**

**Symptoms:**
- Health check endpoints failing
- High error rate in monitoring
- User reports of service unavailability

**Diagnostic Steps:**
```bash
# Check pod status
kubectl get pods -n go-coffee -l app=coffee-service

# Check recent events
kubectl get events -n go-coffee --sort-by=.metadata.creationTimestamp | tail -20

# Check application logs
kubectl logs -n go-coffee -l app=coffee-service --tail=100

# Check resource usage
kubectl top pods -n go-coffee -l app=coffee-service
```

**Common Causes and Solutions:**

1. **Out of Memory (OOMKilled)**
   ```bash
   # Check for OOM events
   kubectl describe pod <pod-name> -n go-coffee | grep -A 5 "Last State"
   
   # Solution: Increase memory limits
   kubectl patch deployment coffee-service -n go-coffee -p '
   {
     "spec": {
       "template": {
         "spec": {
           "containers": [{
             "name": "coffee-service",
             "resources": {
               "limits": {"memory": "1Gi"},
               "requests": {"memory": "512Mi"}
             }
           }]
         }
       }
     }
   }'
   ```

2. **Database Connection Issues**
   ```bash
   # Test database connectivity
   kubectl exec -it postgres-primary-0 -n go-coffee -- pg_isready
   
   # Check connection count
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT count(*) FROM pg_stat_activity;"
   
   # Solution: Restart database connections
   kubectl rollout restart deployment coffee-service -n go-coffee
   ```

3. **Configuration Issues**
   ```bash
   # Check ConfigMap
   kubectl get configmap app-config -n go-coffee -o yaml
   
   # Check Secrets
   kubectl get secret app-secrets -n go-coffee -o yaml
   
   # Solution: Update configuration
   kubectl patch configmap app-config -n go-coffee --patch '{"data":{"LOG_LEVEL":"debug"}}'
   ```

#### **Issue: Slow Response Times**

**Symptoms:**
- Response times > 2 seconds
- Timeout errors
- Poor user experience

**Diagnostic Steps:**
```bash
# Check response time metrics
curl -w "@curl-format.txt" -o /dev/null -s https://api.go-coffee.com/health

# Create curl-format.txt
cat > curl-format.txt << EOF
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
EOF

# Check database performance
kubectl exec -it postgres-primary-0 -n go-coffee -- \
  psql -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"
```

**Solutions:**

1. **Database Query Optimization**
   ```bash
   # Identify slow queries
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE (now() - pg_stat_activity.query_start) > interval '5 seconds';"
   
   # Create indexes for slow queries
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "CREATE INDEX CONCURRENTLY idx_orders_user_id ON orders(user_id);"
   ```

2. **Scale Up Resources**
   ```bash
   # Increase replicas
   kubectl scale deployment coffee-service --replicas=5 -n go-coffee
   
   # Increase resource limits
   kubectl patch deployment coffee-service -n go-coffee -p '
   {
     "spec": {
       "template": {
         "spec": {
           "containers": [{
             "name": "coffee-service",
             "resources": {
               "limits": {"cpu": "1000m", "memory": "1Gi"},
               "requests": {"cpu": "500m", "memory": "512Mi"}
             }
           }]
         }
       }
     }
   }'
   ```

3. **Enable Caching**
   ```bash
   # Check Redis connectivity
   kubectl exec -it redis-0 -n go-coffee -- redis-cli ping
   
   # Monitor cache hit rate
   kubectl exec -it redis-0 -n go-coffee -- redis-cli info stats | grep keyspace
   ```

### Infrastructure Issues

#### **Issue: Pod Stuck in Pending State**

**Symptoms:**
- Pods not starting
- Deployment rollout stuck
- Resource scheduling issues

**Diagnostic Steps:**
```bash
# Check pod status and events
kubectl describe pod <pod-name> -n go-coffee

# Check node resources
kubectl describe nodes | grep -A 5 "Allocated resources"

# Check resource quotas
kubectl describe resourcequota -n go-coffee
```

**Common Causes and Solutions:**

1. **Insufficient Node Resources**
   ```bash
   # Check node capacity
   kubectl top nodes
   
   # Solution: Add more nodes or reduce resource requests
   kubectl patch deployment coffee-service -n go-coffee -p '
   {
     "spec": {
       "template": {
         "spec": {
           "containers": [{
             "name": "coffee-service",
             "resources": {
               "requests": {"cpu": "100m", "memory": "128Mi"}
             }
           }]
         }
       }
     }
   }'
   ```

2. **Node Selector Issues**
   ```bash
   # Check node labels
   kubectl get nodes --show-labels
   
   # Solution: Update node selector or add labels
   kubectl label nodes <node-name> workload-type=general
   ```

3. **Persistent Volume Issues**
   ```bash
   # Check PV and PVC status
   kubectl get pv,pvc -n go-coffee
   
   # Check storage class
   kubectl get storageclass
   
   # Solution: Create or fix storage resources
   kubectl apply -f - <<EOF
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: data-claim
     namespace: go-coffee
   spec:
     accessModes: [ReadWriteOnce]
     resources:
       requests:
         storage: 10Gi
     storageClassName: fast-ssd
   EOF
   ```

#### **Issue: Network Connectivity Problems**

**Symptoms:**
- Services cannot communicate
- DNS resolution failures
- Timeout errors between services

**Diagnostic Steps:**
```bash
# Test DNS resolution
kubectl exec -it <pod-name> -n go-coffee -- nslookup coffee-service.go-coffee.svc.cluster.local

# Test service connectivity
kubectl exec -it <pod-name> -n go-coffee -- curl -v http://coffee-service:8080/health

# Check network policies
kubectl get networkpolicies -n go-coffee

# Check service endpoints
kubectl get endpoints -n go-coffee
```

**Solutions:**

1. **DNS Issues**
   ```bash
   # Check CoreDNS
   kubectl get pods -n kube-system -l k8s-app=kube-dns
   
   # Restart CoreDNS if needed
   kubectl rollout restart deployment/coredns -n kube-system
   ```

2. **Network Policy Issues**
   ```bash
   # Check existing policies
   kubectl describe networkpolicy -n go-coffee
   
   # Create allow-all policy for debugging
   kubectl apply -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-all-debug
     namespace: go-coffee
   spec:
     podSelector: {}
     policyTypes:
     - Ingress
     - Egress
     ingress:
     - {}
     egress:
     - {}
   EOF
   ```

3. **Service Configuration Issues**
   ```bash
   # Check service configuration
   kubectl get service coffee-service -n go-coffee -o yaml
   
   # Verify pod labels match service selector
   kubectl get pods -n go-coffee --show-labels
   ```

### Database Issues

#### **Issue: Database Connection Pool Exhaustion**

**Symptoms:**
- "too many connections" errors
- Application timeouts
- Database performance degradation

**Diagnostic Steps:**
```bash
# Check current connections
kubectl exec -it postgres-primary-0 -n go-coffee -- \
  psql -c "SELECT count(*) FROM pg_stat_activity;"

# Check connection limits
kubectl exec -it postgres-primary-0 -n go-coffee -- \
  psql -c "SHOW max_connections;"

# Check connection sources
kubectl exec -it postgres-primary-0 -n go-coffee -- \
  psql -c "SELECT client_addr, count(*) FROM pg_stat_activity GROUP BY client_addr;"
```

**Solutions:**

1. **Increase Connection Limit**
   ```bash
   # Update PostgreSQL configuration
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "ALTER SYSTEM SET max_connections = 200;"
   
   # Restart PostgreSQL
   kubectl rollout restart statefulset postgres -n go-coffee
   ```

2. **Optimize Application Connection Pooling**
   ```bash
   # Update application configuration
   kubectl patch configmap app-config -n go-coffee --patch '
   {
     "data": {
       "DB_MAX_CONNECTIONS": "20",
       "DB_MAX_IDLE_CONNECTIONS": "5",
       "DB_CONNECTION_TIMEOUT": "30s"
     }
   }'
   
   # Restart application
   kubectl rollout restart deployment coffee-service -n go-coffee
   ```

3. **Kill Long-Running Queries**
   ```bash
   # Identify long-running queries
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes';"
   
   # Kill specific queries
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT pg_terminate_backend(<pid>);"
   ```

#### **Issue: Database Replication Lag**

**Symptoms:**
- Read replicas behind primary
- Data inconsistency
- Replication alerts firing

**Diagnostic Steps:**
```bash
# Check replication status
kubectl exec -it postgres-primary-0 -n go-coffee -- \
  psql -c "SELECT * FROM pg_stat_replication;"

# Check replication lag
kubectl exec -it postgres-replica-0 -n go-coffee -- \
  psql -c "SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()));"
```

**Solutions:**

1. **Network Issues**
   ```bash
   # Test connectivity between primary and replica
   kubectl exec -it postgres-replica-0 -n go-coffee -- \
     ping postgres-primary-0.postgres.go-coffee.svc.cluster.local
   ```

2. **Resource Constraints**
   ```bash
   # Check resource usage on replica
   kubectl top pod postgres-replica-0 -n go-coffee
   
   # Increase resources if needed
   kubectl patch statefulset postgres-replica -n go-coffee -p '
   {
     "spec": {
       "template": {
         "spec": {
           "containers": [{
             "name": "postgres",
             "resources": {
               "limits": {"cpu": "2000m", "memory": "4Gi"},
               "requests": {"cpu": "1000m", "memory": "2Gi"}
             }
           }]
         }
       }
     }
   }'
   ```

### Monitoring and Alerting Issues

#### **Issue: Missing Metrics or Dashboards**

**Symptoms:**
- Grafana dashboards showing no data
- Missing metrics in Prometheus
- Alerts not firing when expected

**Diagnostic Steps:**
```bash
# Check Prometheus targets
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Visit http://localhost:9090/targets

# Check ServiceMonitor resources
kubectl get servicemonitor -n go-coffee

# Check if metrics endpoint is accessible
kubectl exec -it <pod-name> -n go-coffee -- curl http://localhost:9090/metrics
```

**Solutions:**

1. **Fix ServiceMonitor Configuration**
   ```bash
   # Create or update ServiceMonitor
   kubectl apply -f - <<EOF
   apiVersion: monitoring.coreos.com/v1
   kind: ServiceMonitor
   metadata:
     name: coffee-service-monitor
     namespace: go-coffee
   spec:
     selector:
       matchLabels:
         app: coffee-service
     endpoints:
     - port: metrics
       path: /metrics
       interval: 30s
   EOF
   ```

2. **Check Prometheus Configuration**
   ```bash
   # Check Prometheus config
   kubectl get prometheus -n monitoring -o yaml
   
   # Check if namespace is being monitored
   kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorNamespaceSelector}'
   ```

3. **Restart Monitoring Stack**
   ```bash
   # Restart Prometheus
   kubectl rollout restart statefulset prometheus-prometheus -n monitoring
   
   # Restart Grafana
   kubectl rollout restart deployment grafana -n monitoring
   ```

---

## 🔍 Diagnostic Tools and Commands

### Essential Kubernetes Commands

```bash
# Pod diagnostics
kubectl get pods -n <namespace> -o wide
kubectl describe pod <pod-name> -n <namespace>
kubectl logs <pod-name> -n <namespace> --previous
kubectl exec -it <pod-name> -n <namespace> -- /bin/bash

# Service diagnostics
kubectl get services -n <namespace>
kubectl describe service <service-name> -n <namespace>
kubectl get endpoints <service-name> -n <namespace>

# Resource diagnostics
kubectl top nodes
kubectl top pods -n <namespace>
kubectl describe node <node-name>

# Event diagnostics
kubectl get events -n <namespace> --sort-by=.metadata.creationTimestamp
kubectl get events --all-namespaces --sort-by=.metadata.creationTimestamp
```

### Network Diagnostics

```bash
# DNS testing
kubectl exec -it <pod-name> -n <namespace> -- nslookup <service-name>
kubectl exec -it <pod-name> -n <namespace> -- dig <domain-name>

# Connectivity testing
kubectl exec -it <pod-name> -n <namespace> -- curl -v <url>
kubectl exec -it <pod-name> -n <namespace> -- telnet <host> <port>
kubectl exec -it <pod-name> -n <namespace> -- ping <host>

# Network policy testing
kubectl exec -it <pod-name> -n <namespace> -- nc -zv <host> <port>
```

### Database Diagnostics

```bash
# PostgreSQL diagnostics
kubectl exec -it postgres-primary-0 -n go-coffee -- pg_isready
kubectl exec -it postgres-primary-0 -n go-coffee -- psql -c "SELECT version();"
kubectl exec -it postgres-primary-0 -n go-coffee -- psql -c "SELECT * FROM pg_stat_activity;"

# Redis diagnostics
kubectl exec -it redis-0 -n go-coffee -- redis-cli ping
kubectl exec -it redis-0 -n go-coffee -- redis-cli info
kubectl exec -it redis-0 -n go-coffee -- redis-cli monitor
```

---

## 📞 Escalation Procedures

### When to Escalate

1. **Immediate Escalation (P0)**
   - Complete service outage
   - Security breach
   - Data loss or corruption
   - Customer-facing critical issues

2. **Escalate Within 1 Hour (P1)**
   - Significant performance degradation
   - Partial service outage
   - Failed deployments to production
   - Monitoring system failures

3. **Escalate Within 4 Hours (P2)**
   - Non-critical service issues
   - Development environment problems
   - Monitoring alerts without immediate impact

### Escalation Contacts

| Issue Type | Primary Contact | Secondary Contact |
|------------|----------------|-------------------|
| Infrastructure | DevOps Engineer | Platform Lead |
| Application | Development Team | Tech Lead |
| Database | Database Admin | Senior Developer |
| Security | Security Team | CISO |
| Network | Network Engineer | Infrastructure Lead |

### Information to Include

When escalating, provide:
- **Issue Description**: Clear summary of the problem
- **Impact Assessment**: Affected services and users
- **Timeline**: When the issue started and key events
- **Diagnostic Results**: Commands run and outputs
- **Attempted Solutions**: What has been tried
- **Current Status**: Current state and ongoing actions

---

## 📚 Additional Resources

### Monitoring Dashboards
- **Grafana**: https://grafana.go-coffee.com
- **Prometheus**: https://prometheus.go-coffee.com
- **AlertManager**: https://alertmanager.go-coffee.com

### Log Analysis
- **Kibana**: https://kibana.go-coffee.com
- **Jaeger**: https://jaeger.go-coffee.com

### Documentation
- [Architecture Overview](./ARCHITECTURE_OVERVIEW.md)
- [Operational Runbooks](./OPERATIONAL_RUNBOOKS.md)
- [Deployment Guide](./DEPLOYMENT_GUIDE.md)

### External Resources
- [Kubernetes Troubleshooting](https://kubernetes.io/docs/tasks/debug-application-cluster/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Prometheus Troubleshooting](https://prometheus.io/docs/prometheus/latest/troubleshooting/)

---

*This troubleshooting guide is continuously updated based on operational experience. Please contribute improvements and new solutions as they are discovered.*
