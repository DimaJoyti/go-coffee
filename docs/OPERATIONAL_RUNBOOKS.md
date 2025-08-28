# 📚 Go Coffee Platform - Operational Runbooks

## 📋 Overview

This document contains comprehensive operational runbooks for managing the Go Coffee multi-cloud platform. These runbooks provide step-by-step procedures for common operational tasks, troubleshooting, and incident response.

## 🎯 Runbook Categories

1. **Daily Operations** - Routine maintenance and monitoring tasks
2. **Incident Response** - Emergency procedures and troubleshooting
3. **Deployment Operations** - Application and infrastructure deployment
4. **Security Operations** - Security monitoring and response
5. **Performance Optimization** - System tuning and optimization
6. **Disaster Recovery** - Backup, restore, and failover procedures

---

## 🌅 Daily Operations

### Morning Health Check Routine

#### **System Health Verification (15 minutes)**

1. **Check Overall System Status**
   ```bash
   # Check all Kubernetes clusters
   kubectl get nodes --all-namespaces
   kubectl get pods --all-namespaces | grep -v Running
   
   # Check critical services
   kubectl get services -n go-coffee
   kubectl get ingresses -n go-coffee
   ```

2. **Verify Application Health**
   ```bash
   # Health check endpoints
   curl -f https://api.go-coffee.com/health
   curl -f https://go-coffee.com/api/health
   
   # Database connectivity
   kubectl exec -it postgres-primary-0 -n go-coffee -- pg_isready
   
   # Redis connectivity
   kubectl exec -it redis-0 -n go-coffee -- redis-cli ping
   ```

3. **Review Monitoring Dashboards**
   - Open Grafana: `https://grafana.go-coffee.com`
   - Check "Go Coffee Overview" dashboard
   - Verify all services are green
   - Review error rates and response times

4. **Check Recent Alerts**
   ```bash
   # Check AlertManager
   curl -s https://alertmanager.go-coffee.com/api/v1/alerts | jq '.data[] | select(.state=="firing")'
   
   # Review Slack notifications from last 24 hours
   # Check #alerts channel for any critical issues
   ```

### Weekly Maintenance Tasks

#### **Security Updates (30 minutes)**

1. **Check for Security Vulnerabilities**
   ```bash
   # Run Trivy scans on all images
   trivy image go-coffee/backend:latest
   trivy image go-coffee/frontend:latest
   
   # Check for Kubernetes security issues
   kubectl get vulnerabilityreports -A
   ```

2. **Update Dependencies**
   ```bash
   # Backend dependencies
   cd backend && go mod tidy && go mod download
   
   # Frontend dependencies
   cd web-ui/frontend && npm audit fix
   ```

3. **Review Access Logs**
   ```bash
   # Check for suspicious activity
   kubectl logs -n go-coffee -l app=api-gateway --since=7d | grep -E "(401|403|429)"
   ```

#### **Performance Review (45 minutes)**

1. **Resource Utilization Analysis**
   ```bash
   # Check resource usage
   kubectl top nodes
   kubectl top pods -A --sort-by=cpu
   kubectl top pods -A --sort-by=memory
   ```

2. **Database Performance**
   ```bash
   # Check slow queries
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"
   ```

3. **Cost Analysis**
   - Review KubeCost dashboard
   - Check cloud provider billing
   - Identify optimization opportunities

---

## 🚨 Incident Response

### Critical Service Down

#### **Severity: P0 - Complete Service Outage**

**Detection Signs:**
- Health check endpoints returning 5xx errors
- Multiple user reports of service unavailability
- Monitoring alerts firing for all services

**Immediate Response (0-5 minutes):**

1. **Acknowledge the Incident**
   ```bash
   # Post in #incidents channel
   echo "🚨 P0 INCIDENT: Complete service outage detected. Investigating..."
   
   # Page on-call engineer if not already aware
   ```

2. **Quick Assessment**
   ```bash
   # Check overall cluster health
   kubectl get nodes
   kubectl get pods -n go-coffee | grep -v Running
   
   # Check ingress and load balancer
   kubectl get ingress -n go-coffee
   kubectl describe ingress go-coffee-ingress -n go-coffee
   ```

3. **Check External Dependencies**
   ```bash
   # Database connectivity
   kubectl exec -it postgres-primary-0 -n go-coffee -- pg_isready
   
   # Redis connectivity
   kubectl exec -it redis-0 -n go-coffee -- redis-cli ping
   
   # External API dependencies
   curl -f https://api.stripe.com/v1/charges
   ```

**Investigation and Resolution (5-30 minutes):**

1. **Identify Root Cause**
   ```bash
   # Check recent deployments
   kubectl rollout history deployment/coffee-service -n go-coffee
   
   # Review recent events
   kubectl get events -n go-coffee --sort-by=.metadata.creationTimestamp
   
   # Check application logs
   kubectl logs -n go-coffee -l app=coffee-service --tail=100
   ```

2. **Apply Quick Fix**
   ```bash
   # If recent deployment caused issue, rollback
   kubectl rollout undo deployment/coffee-service -n go-coffee
   
   # If resource issue, scale up
   kubectl scale deployment coffee-service --replicas=5 -n go-coffee
   
   # If database issue, failover to replica
   # (Follow database failover procedure)
   ```

3. **Verify Resolution**
   ```bash
   # Test health endpoints
   curl -f https://api.go-coffee.com/health
   
   # Check service status
   kubectl get pods -n go-coffee -l app=coffee-service
   ```

### High Error Rate

#### **Severity: P1 - Degraded Service Performance**

**Detection Signs:**
- Error rate > 5% for critical endpoints
- Increased response times
- User complaints about slow performance

**Response Procedure:**

1. **Identify Affected Services**
   ```bash
   # Check error rates by service
   kubectl logs -n go-coffee -l app=coffee-service --since=10m | grep ERROR | wc -l
   kubectl logs -n go-coffee -l app=payment-service --since=10m | grep ERROR | wc -l
   ```

2. **Analyze Error Patterns**
   ```bash
   # Check common error types
   kubectl logs -n go-coffee -l app=coffee-service --since=10m | grep ERROR | sort | uniq -c
   
   # Check database errors
   kubectl logs -n go-coffee -l app=postgres --since=10m | grep ERROR
   ```

3. **Apply Mitigation**
   ```bash
   # Scale up affected services
   kubectl scale deployment coffee-service --replicas=3 -n go-coffee
   
   # Enable circuit breaker if available
   # Increase timeout values if needed
   ```

### Database Performance Issues

#### **Severity: P1 - Database Slowdown**

**Response Procedure:**

1. **Check Database Health**
   ```bash
   # Connection count
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT count(*) FROM pg_stat_activity;"
   
   # Active queries
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes';"
   ```

2. **Identify Slow Queries**
   ```bash
   # Top slow queries
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 5;"
   ```

3. **Apply Quick Fixes**
   ```bash
   # Kill long-running queries if safe
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE (now() - pg_stat_activity.query_start) > interval '10 minutes';"
   
   # Increase connection pool if needed
   kubectl patch configmap postgres-config -n go-coffee --patch '{"data":{"max_connections":"200"}}'
   ```

---

## 🚀 Deployment Operations

### Application Deployment

#### **Standard Deployment Process**

1. **Pre-Deployment Checks**
   ```bash
   # Verify staging deployment
   curl -f https://staging-api.go-coffee.com/health
   
   # Check resource availability
   kubectl describe nodes | grep -A 5 "Allocated resources"
   
   # Verify backup completion
   velero backup get | head -5
   ```

2. **Deploy to Production**
   ```bash
   # Using ArgoCD
   argocd app sync go-coffee-production
   
   # Or using kubectl
   kubectl set image deployment/coffee-service \
     coffee-service=go-coffee/backend:v1.2.3 -n go-coffee
   ```

3. **Post-Deployment Verification**
   ```bash
   # Check rollout status
   kubectl rollout status deployment/coffee-service -n go-coffee
   
   # Verify health endpoints
   curl -f https://api.go-coffee.com/health
   
   # Check error rates
   kubectl logs -n go-coffee -l app=coffee-service --since=5m | grep ERROR
   ```

### Infrastructure Updates

#### **Kubernetes Cluster Updates**

1. **Prepare for Update**
   ```bash
   # Backup cluster state
   velero backup create cluster-backup-$(date +%Y%m%d)
   
   # Check node readiness
   kubectl get nodes
   kubectl describe nodes | grep -A 10 "Conditions"
   ```

2. **Update Process**
   ```bash
   # Update control plane (managed service)
   # AWS EKS: Update through AWS Console or CLI
   # GKE: Update through Google Cloud Console or CLI
   
   # Update worker nodes (rolling update)
   kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data
   # Update node
   kubectl uncordon <node-name>
   ```

3. **Verify Update**
   ```bash
   # Check cluster version
   kubectl version
   
   # Verify all pods are running
   kubectl get pods --all-namespaces | grep -v Running
   ```

---

## 🔒 Security Operations

### Security Incident Response

#### **Suspected Security Breach**

1. **Immediate Containment**
   ```bash
   # Isolate affected pods
   kubectl label pod <suspicious-pod> security.policy=quarantine
   
   # Block suspicious IP addresses
   kubectl apply -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: block-suspicious-ip
   spec:
     podSelector: {}
     policyTypes:
     - Ingress
     ingress:
     - from:
       - ipBlock:
           cidr: 0.0.0.0/0
           except:
           - <suspicious-ip>/32
   EOF
   ```

2. **Evidence Collection**
   ```bash
   # Collect logs
   kubectl logs <suspicious-pod> > incident-logs-$(date +%Y%m%d).log
   
   # Collect network traffic
   kubectl exec -it <network-monitoring-pod> -- tcpdump -w incident-traffic.pcap
   ```

3. **Notify Security Team**
   ```bash
   # Send alert to security team
   curl -X POST -H 'Content-type: application/json' \
     --data '{"text":"🚨 SECURITY INCIDENT: Suspicious activity detected. Investigation in progress."}' \
     $SECURITY_SLACK_WEBHOOK
   ```

### Vulnerability Management

#### **Critical Vulnerability Response**

1. **Assess Impact**
   ```bash
   # Scan all images for vulnerability
   trivy image --severity HIGH,CRITICAL go-coffee/backend:latest
   trivy image --severity HIGH,CRITICAL go-coffee/frontend:latest
   ```

2. **Apply Security Patches**
   ```bash
   # Update base images
   docker build --no-cache -t go-coffee/backend:patched .
   
   # Deploy patched version
   kubectl set image deployment/coffee-service \
     coffee-service=go-coffee/backend:patched -n go-coffee
   ```

3. **Verify Patch Effectiveness**
   ```bash
   # Re-scan for vulnerabilities
   trivy image go-coffee/backend:patched
   
   # Check application functionality
   curl -f https://api.go-coffee.com/health
   ```

---

## ⚡ Performance Optimization

### Database Optimization

#### **Query Performance Tuning**

1. **Identify Slow Queries**
   ```bash
   # Enable slow query logging
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "ALTER SYSTEM SET log_min_duration_statement = 1000;"
   
   # Reload configuration
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT pg_reload_conf();"
   ```

2. **Analyze Query Plans**
   ```bash
   # Explain slow queries
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "EXPLAIN ANALYZE SELECT * FROM orders WHERE created_at > NOW() - INTERVAL '1 day';"
   ```

3. **Create Indexes**
   ```bash
   # Create missing indexes
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "CREATE INDEX CONCURRENTLY idx_orders_created_at ON orders(created_at);"
   ```

### Application Performance

#### **Memory Optimization**

1. **Identify Memory Leaks**
   ```bash
   # Check memory usage trends
   kubectl top pods -n go-coffee --sort-by=memory
   
   # Get detailed memory stats
   kubectl exec -it <pod-name> -n go-coffee -- cat /proc/meminfo
   ```

2. **Optimize Resource Limits**
   ```bash
   # Update resource limits
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

---

## 🛡️ Disaster Recovery Operations

### Backup Procedures

#### **Manual Backup Creation**

1. **Database Backup**
   ```bash
   # Create database backup
   kubectl exec postgres-primary-0 -n go-coffee -- \
     pg_dump -U postgres go_coffee > backup-$(date +%Y%m%d).sql
   
   # Upload to S3
   aws s3 cp backup-$(date +%Y%m%d).sql s3://go-coffee-backups/manual/
   ```

2. **Application Backup**
   ```bash
   # Create Velero backup
   velero backup create manual-backup-$(date +%Y%m%d) \
     --include-namespaces go-coffee
   
   # Verify backup
   velero backup describe manual-backup-$(date +%Y%m%d)
   ```

### Restore Procedures

#### **Database Restore**

1. **Prepare for Restore**
   ```bash
   # Scale down applications
   kubectl scale deployment --replicas=0 -n go-coffee --all
   
   # Verify no connections to database
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT count(*) FROM pg_stat_activity WHERE datname='go_coffee';"
   ```

2. **Perform Restore**
   ```bash
   # Download backup
   aws s3 cp s3://go-coffee-backups/manual/backup-20231201.sql .
   
   # Restore database
   kubectl exec -i postgres-primary-0 -n go-coffee -- \
     psql -U postgres go_coffee < backup-20231201.sql
   ```

3. **Verify and Resume**
   ```bash
   # Verify data integrity
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT count(*) FROM orders;"
   
   # Scale up applications
   kubectl scale deployment --replicas=3 -n go-coffee --all
   ```

---

## 📞 Emergency Contacts

### Escalation Matrix

| Severity | Response Time | Primary Contact | Secondary Contact |
|----------|---------------|-----------------|-------------------|
| P0 | 15 minutes | Platform Lead | CTO |
| P1 | 1 hour | DevOps Engineer | Platform Lead |
| P2 | 4 hours | Developer | DevOps Engineer |
| P3 | 24 hours | Team Lead | Developer |

### Contact Information

- **Platform Team**: platform@go-coffee.com
- **Security Team**: security@go-coffee.com
- **On-Call Engineer**: +1-555-ONCALL
- **Emergency Hotline**: +1-555-EMERGENCY

---

## 📚 Additional Resources

- **Monitoring Dashboards**: https://grafana.go-coffee.com
- **Log Analysis**: https://kibana.go-coffee.com
- **ArgoCD**: https://argocd.go-coffee.com
- **Internal Wiki**: https://wiki.go-coffee.com
- **Incident Management**: https://incidents.go-coffee.com

---

*These runbooks are living documents and should be updated regularly based on operational experience and system changes.*
