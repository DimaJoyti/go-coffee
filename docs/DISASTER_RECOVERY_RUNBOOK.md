# 🛡️ Go Coffee Disaster Recovery Runbook

## 📋 Overview

This runbook provides comprehensive procedures for disaster recovery scenarios in the Go Coffee platform. It covers detection, response, recovery, and post-incident activities.

## 🎯 Recovery Objectives

- **RTO (Recovery Time Objective)**: 60 minutes
- **RPO (Recovery Point Objective)**: 15 minutes
- **Availability Target**: 99.99% uptime
- **Data Loss Tolerance**: Maximum 15 minutes

## 📞 Emergency Contacts

### Primary Contacts
- **Incident Commander**: platform-lead@go-coffee.com | +1-555-0101
- **Technical Lead**: devops@go-coffee.com | +1-555-0102
- **Database Specialist**: dba@go-coffee.com | +1-555-0103

### Escalation Contacts
- **CTO**: cto@go-coffee.com | +1-555-0001
- **Operations Manager**: ops-manager@go-coffee.com | +1-555-0002

### 24/7 Support
- **PagerDuty**: go-coffee-platform
- **Slack Channel**: #go-coffee-incidents
- **Emergency Hotline**: +1-555-DR-HELP

## 🚨 Disaster Scenarios

### 1. Complete Region Failure

#### Detection Signs
- Multiple service health checks failing across all services
- No response from primary region endpoints (us-east-1)
- Cloud provider status page indicates regional issues
- Monitoring alerts: `PrimaryRegionDown` firing

#### Immediate Response (0-5 minutes)
1. **Assess Situation**
   ```bash
   # Check primary region status
   kubectl get nodes --context=primary-cluster
   curl -f https://api.go-coffee.com/health || echo "Primary region down"
   
   # Check secondary region availability
   kubectl get nodes --context=secondary-cluster
   curl -f https://api-dr.go-coffee.com/health || echo "Secondary region down"
   ```

2. **Notify Incident Response Team**
   ```bash
   # Send alert to Slack
   curl -X POST -H 'Content-type: application/json' \
     --data '{"text":"🚨 CRITICAL: Primary region failure detected. Initiating DR procedures."}' \
     $SLACK_WEBHOOK_URL
   ```

#### Failover Execution (5-20 minutes)
1. **Automated Failover Check**
   ```bash
   # Check if automated failover is active
   kubectl get pods -n disaster-recovery -l app=failover-controller
   kubectl logs -n disaster-recovery -l app=failover-controller --tail=50
   ```

2. **Manual Failover (if automated fails)**
   ```bash
   # Switch DNS to secondary region
   aws route53 change-resource-record-sets \
     --hosted-zone-id Z123456789 \
     --change-batch file://dns-failover.json
   
   # Scale up DR services
   kubectl scale deployment coffee-service --replicas=3 -n go-coffee-dr
   kubectl scale deployment payment-service --replicas=2 -n go-coffee-dr
   kubectl scale deployment user-service --replicas=2 -n go-coffee-dr
   kubectl scale deployment order-service --replicas=2 -n go-coffee-dr
   kubectl scale deployment inventory-service --replicas=2 -n go-coffee-dr
   ```

3. **Database Failover**
   ```bash
   # Promote read replica to primary
   aws rds promote-read-replica --db-instance-identifier go-coffee-db-replica-west
   
   # Update connection strings
   kubectl patch secret database-credentials -n go-coffee-dr \
     -p '{"data":{"url":"'$(echo -n "postgresql://user:pass@new-primary:5432/go_coffee" | base64)'"}}'
   ```

#### Verification (20-30 minutes)
1. **Test Critical User Journeys**
   ```bash
   # Test user registration
   curl -X POST https://api-dr.go-coffee.com/v1/users \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"test123"}'
   
   # Test coffee ordering
   curl -X POST https://api-dr.go-coffee.com/v1/orders \
     -H "Authorization: Bearer $TEST_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"coffee_type":"latte","size":"large"}'
   
   # Test payment processing
   curl -X POST https://api-dr.go-coffee.com/v1/payments \
     -H "Authorization: Bearer $TEST_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"order_id":"123","amount":5.99,"method":"card"}'
   ```

2. **Monitor System Health**
   ```bash
   # Check all pods are running
   kubectl get pods -n go-coffee-dr
   
   # Check service endpoints
   kubectl get endpoints -n go-coffee-dr
   
   # Verify database connectivity
   kubectl exec -it postgres-primary-0 -n go-coffee-dr -- pg_isready
   ```

### 2. Database Failure

#### Detection Signs
- Database connection errors in application logs
- High error rates (>5%) in database queries
- Database monitoring alerts firing
- `DatabaseReplicationLag` or `DatabaseDown` alerts

#### Response Steps (0-10 minutes)
1. **Assess Database State**
   ```bash
   # Check primary database
   kubectl exec -it postgres-primary-0 -n go-coffee -- pg_isready
   
   # Check replication status
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT * FROM pg_stat_replication;"
   
   # Check replica databases
   kubectl exec -it postgres-replica-0 -n go-coffee -- pg_isready
   ```

2. **Determine Recovery Strategy**
   - If primary is recoverable: Restart and repair
   - If primary is lost: Promote replica
   - If all databases lost: Restore from backup

#### Database Recovery (10-30 minutes)
1. **Promote Replica (if primary lost)**
   ```bash
   # Stop replica
   kubectl exec -it postgres-replica-0 -n go-coffee -- \
     su - postgres -c "pg_ctl stop -D /var/lib/postgresql/data"
   
   # Promote replica to primary
   kubectl exec -it postgres-replica-0 -n go-coffee -- \
     su - postgres -c "pg_ctl promote -D /var/lib/postgresql/data"
   
   # Update service to point to new primary
   kubectl patch service postgres-primary -n go-coffee \
     -p '{"spec":{"selector":{"app":"postgres-replica"}}}'
   ```

2. **Restore from Backup (if all databases lost)**
   ```bash
   # Get latest backup
   LATEST_BACKUP=$(velero backup get -o name | head -1)
   
   # Restore database
   velero restore create db-restore-$(date +%s) \
     --from-backup $LATEST_BACKUP \
     --include-resources persistentvolumeclaims,persistentvolumes
   
   # Wait for restore completion
   velero restore describe db-restore-$(date +%s)
   ```

#### Verification (30-40 minutes)
1. **Test Database Connectivity**
   ```bash
   # Test read operations
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT COUNT(*) FROM users;"
   
   # Test write operations
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "INSERT INTO health_check (timestamp) VALUES (NOW());"
   ```

2. **Verify Data Integrity**
   ```bash
   # Check recent transactions
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT * FROM orders WHERE created_at > NOW() - INTERVAL '1 hour';"
   
   # Verify user data
   kubectl exec -it postgres-primary-0 -n go-coffee -- \
     psql -c "SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '1 day';"
   ```

### 3. Application Service Failure

#### Detection Signs
- Service health checks failing
- High error rates in specific services
- Increased response times
- Service-specific alerts firing

#### Response Steps (0-15 minutes)
1. **Identify Failed Services**
   ```bash
   # Check pod status
   kubectl get pods -n go-coffee | grep -v Running
   
   # Check service endpoints
   kubectl get endpoints -n go-coffee
   
   # Check recent events
   kubectl get events -n go-coffee --sort-by=.metadata.creationTimestamp
   ```

2. **Restart Failed Services**
   ```bash
   # Restart specific deployment
   kubectl rollout restart deployment/coffee-service -n go-coffee
   
   # Check rollout status
   kubectl rollout status deployment/coffee-service -n go-coffee
   ```

3. **Scale Services if Needed**
   ```bash
   # Scale up healthy services to handle load
   kubectl scale deployment payment-service --replicas=5 -n go-coffee
   kubectl scale deployment user-service --replicas=4 -n go-coffee
   ```

## 🔄 Recovery Procedures

### Backup Restoration

#### List Available Backups
```bash
# Velero backups
velero backup get

# S3 backups
aws s3 ls s3://go-coffee-backups/production/ --recursive

# Local backups
ls -la /var/backups/go-coffee/production/
```

#### Restore Specific Backup
```bash
# Restore from Velero backup
velero restore create restore-$(date +%s) --from-backup <backup-name>

# Restore from S3 backup
aws s3 cp s3://go-coffee-backups/production/backup.tar.gz /tmp/
cd /tmp && tar -xzf backup.tar.gz
./scripts/disaster-recovery/restore.sh /tmp/backup-folder

# Monitor restore progress
velero restore describe restore-$(date +%s)
```

### Service Recovery

#### Check Service Health
```bash
# Overall cluster health
kubectl get nodes
kubectl get pods --all-namespaces | grep -v Running

# Service-specific health
kubectl get pods -n go-coffee
kubectl get services -n go-coffee
kubectl get ingresses -n go-coffee
```

#### Restart Services
```bash
# Restart all services
kubectl rollout restart deployment -n go-coffee

# Restart specific service
kubectl rollout restart deployment/coffee-service -n go-coffee

# Check rollout status
kubectl rollout status deployment/coffee-service -n go-coffee --timeout=300s
```

#### Scale Services
```bash
# Auto-scale based on load
kubectl autoscale deployment coffee-service --cpu-percent=70 --min=2 --max=10 -n go-coffee

# Manual scaling
kubectl scale deployment coffee-service --replicas=5 -n go-coffee
```

## 📊 Post-Incident Actions

### 1. Immediate Assessment (0-2 hours)
- [ ] Verify all services are operational
- [ ] Confirm data integrity
- [ ] Check system performance metrics
- [ ] Update status page
- [ ] Notify stakeholders of resolution

### 2. Root Cause Analysis (2-24 hours)
- [ ] Collect logs and metrics from incident timeframe
- [ ] Interview team members involved in response
- [ ] Identify contributing factors
- [ ] Document timeline of events
- [ ] Create preliminary incident report

### 3. Documentation and Communication (1-3 days)
- [ ] Complete detailed post-mortem report
- [ ] Share findings with stakeholders
- [ ] Update runbooks based on lessons learned
- [ ] Communicate with customers if needed
- [ ] Schedule team retrospective

### 4. Improvement Actions (1-4 weeks)
- [ ] Implement identified improvements
- [ ] Update monitoring and alerting
- [ ] Enhance automation where possible
- [ ] Conduct additional training if needed
- [ ] Test updated procedures

## 🧪 Testing Schedule

### Monthly Tests
- [ ] Backup integrity verification
- [ ] Database failover test
- [ ] Service restart procedures
- [ ] Monitoring alert validation

### Quarterly Tests
- [ ] Full region failover exercise
- [ ] Complete backup restoration
- [ ] Cross-team incident response drill
- [ ] Business continuity plan review

### Annual Tests
- [ ] Comprehensive disaster recovery exercise
- [ ] Third-party DR audit
- [ ] Business impact assessment
- [ ] Insurance and compliance review

## 📚 Useful Commands

### Cluster Operations
```bash
# Check cluster status
kubectl cluster-info
kubectl get nodes
kubectl top nodes

# View recent events
kubectl get events --sort-by=.metadata.creationTimestamp --all-namespaces

# Check resource usage
kubectl top pods -A
kubectl describe node <node-name>
```

### Velero Operations
```bash
# Create backup
velero backup create manual-backup-$(date +%s)

# List backups
velero backup get

# Restore from backup
velero restore create restore-$(date +%s) --from-backup <backup-name>

# Check backup/restore status
velero backup describe <backup-name>
velero restore describe <restore-name>

# Delete old backups
velero backup delete <backup-name>
```

### Database Operations
```bash
# Connect to database
kubectl exec -it postgres-primary-0 -n go-coffee -- psql -U postgres

# Check replication status
kubectl exec -it postgres-primary-0 -n go-coffee -- \
  psql -c "SELECT * FROM pg_stat_replication;"

# Create manual backup
kubectl exec postgres-primary-0 -n go-coffee -- \
  pg_dump -U postgres go_coffee > backup-$(date +%s).sql
```

### Monitoring and Logs
```bash
# Check Prometheus targets
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Visit http://localhost:9090/targets

# View Grafana dashboards
kubectl port-forward -n monitoring svc/grafana 3000:80
# Visit http://localhost:3000

# Collect logs
kubectl logs -n go-coffee -l app=coffee-service --tail=100
kubectl logs -n go-coffee -l app=payment-service --since=1h
```

---

## 📞 Emergency Escalation

If standard procedures fail or the incident severity increases:

1. **Immediately contact the Incident Commander**
2. **Escalate to CTO if business-critical impact**
3. **Engage external support if needed**
4. **Consider activating business continuity plan**

Remember: **Safety first, then service restoration**

---

*This runbook is a living document. Update it after each incident and during regular reviews.*
