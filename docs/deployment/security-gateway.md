# Security Gateway Deployment Guide

## Overview

This guide provides comprehensive instructions for deploying the Security Gateway in various environments, from local development to production-scale deployments on Kubernetes.

## Prerequisites

### System Requirements

#### Minimum Requirements
- **CPU**: 2 cores
- **Memory**: 4GB RAM
- **Storage**: 20GB available space
- **Network**: 1Gbps network interface

#### Recommended Requirements (Production)
- **CPU**: 4+ cores
- **Memory**: 8GB+ RAM
- **Storage**: 100GB+ SSD
- **Network**: 10Gbps network interface

#### Software Dependencies
- **Go**: 1.21+ (for building from source)
- **Docker**: 20.10+ (for containerized deployment)
- **Docker Compose**: 2.0+ (for local development)
- **Kubernetes**: 1.25+ (for production deployment)
- **Redis**: 7.0+ (for rate limiting and caching)

## Deployment Methods

### 1. Local Development Deployment

#### Quick Start with Docker Compose

1. **Clone the repository**:
```bash
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee
```

2. **Generate security keys**:
```bash
./scripts/generate-security-keys.sh
```

3. **Setup environment**:
```bash
cp .env.security.example .env
# Edit .env with your configuration
```

4. **Start services**:
```bash
docker-compose -f docker-compose.security-gateway.yml up -d
```

5. **Verify deployment**:
```bash
curl http://localhost:8080/health
```

#### Manual Local Deployment

1. **Install dependencies**:
```bash
# Install Go 1.21+
go version

# Install Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

2. **Build the service**:
```bash
make -f Makefile.security-gateway build
```

3. **Configure environment**:
```bash
export REDIS_URL="redis://localhost:6379"
export LOG_LEVEL="debug"
export AES_KEY="$(openssl rand -base64 32)"
export JWT_SECRET="$(openssl rand -base64 64)"
```

4. **Run the service**:
```bash
./bin/security-gateway
```

### 2. Docker Deployment

#### Build Docker Image

```bash
# Build the image
docker build -t go-coffee/security-gateway:latest -f cmd/security-gateway/Dockerfile .

# Or use the Makefile
make -f Makefile.security-gateway docker-build
```

#### Run with Docker

```bash
# Create a network
docker network create go-coffee-network

# Start Redis
docker run -d \
  --name redis \
  --network go-coffee-network \
  -p 6379:6379 \
  redis:7-alpine

# Start Security Gateway
docker run -d \
  --name security-gateway \
  --network go-coffee-network \
  -p 8080:8080 \
  -e REDIS_URL="redis://redis:6379" \
  -e LOG_LEVEL="info" \
  -e AES_KEY="your-aes-key" \
  -e JWT_SECRET="your-jwt-secret" \
  go-coffee/security-gateway:latest
```

#### Docker Compose Production Setup

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  security-gateway:
    image: go-coffee/security-gateway:latest
    ports:
      - "8080:8080"
    environment:
      - REDIS_URL=redis://redis:6379
      - LOG_LEVEL=info
      - ENVIRONMENT=production
    env_file:
      - .env.production
    depends_on:
      - redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --maxmemory 1gb --maxmemory-policy allkeys-lru
    volumes:
      - redis-data:/data
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - security-gateway
    restart: unless-stopped

volumes:
  redis-data:
```

### 3. Kubernetes Deployment

#### Prerequisites

1. **Kubernetes cluster** (1.25+)
2. **kubectl** configured
3. **Helm** 3.0+ (optional)

#### Create Namespace

```bash
kubectl create namespace go-coffee-security
```

#### Deploy Redis

Create `redis-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: go-coffee-security
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        command: ["redis-server"]
        args: ["--appendonly", "yes", "--maxmemory", "1gb", "--maxmemory-policy", "allkeys-lru"]
        volumeMounts:
        - name: redis-data
          mountPath: /data
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
      volumes:
      - name: redis-data
        persistentVolumeClaim:
          claimName: redis-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: go-coffee-security
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-pvc
  namespace: go-coffee-security
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

#### Deploy Security Gateway

Create `security-gateway-deployment.yaml`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: security-gateway-secrets
  namespace: go-coffee-security
type: Opaque
data:
  aes-key: <base64-encoded-aes-key>
  jwt-secret: <base64-encoded-jwt-secret>
  redis-password: <base64-encoded-redis-password>
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: security-gateway-config
  namespace: go-coffee-security
data:
  config.yaml: |
    server:
      port: 8080
      host: "0.0.0.0"
    logging:
      level: "info"
      format: "json"
    redis:
      url: "redis://redis:6379"
    rate_limit:
      enabled: true
      requests_per_minute: 1000
    waf:
      enabled: true
    monitoring:
      enabled: true
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: security-gateway
  namespace: go-coffee-security
  labels:
    app: security-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: security-gateway
  template:
    metadata:
      labels:
        app: security-gateway
    spec:
      containers:
      - name: security-gateway
        image: go-coffee/security-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: REDIS_URL
          value: "redis://redis:6379"
        - name: AES_KEY
          valueFrom:
            secretKeyRef:
              name: security-gateway-secrets
              key: aes-key
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: security-gateway-secrets
              key: jwt-secret
        - name: LOG_LEVEL
          value: "info"
        - name: ENVIRONMENT
          value: "production"
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: security-gateway-config
---
apiVersion: v1
kind: Service
metadata:
  name: security-gateway
  namespace: go-coffee-security
spec:
  selector:
    app: security-gateway
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: security-gateway-ingress
  namespace: go-coffee-security
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
  - hosts:
    - security.go-coffee.com
    secretName: security-gateway-tls
  rules:
  - host: security.go-coffee.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: security-gateway
            port:
              number: 80
```

#### Deploy with kubectl

```bash
# Deploy Redis
kubectl apply -f redis-deployment.yaml

# Wait for Redis to be ready
kubectl wait --for=condition=available --timeout=300s deployment/redis -n go-coffee-security

# Deploy Security Gateway
kubectl apply -f security-gateway-deployment.yaml

# Wait for Security Gateway to be ready
kubectl wait --for=condition=available --timeout=300s deployment/security-gateway -n go-coffee-security

# Check deployment status
kubectl get pods -n go-coffee-security
kubectl get services -n go-coffee-security
```

#### Helm Deployment

Create `helm/security-gateway/values.yaml`:

```yaml
replicaCount: 3

image:
  repository: go-coffee/security-gateway
  tag: latest
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: security.go-coffee.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: security-gateway-tls
      hosts:
        - security.go-coffee.com

resources:
  limits:
    cpu: 500m
    memory: 1Gi
  requests:
    cpu: 250m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

redis:
  enabled: true
  auth:
    enabled: true
    password: "redis-password"
  master:
    persistence:
      enabled: true
      size: 10Gi

config:
  server:
    port: 8080
  logging:
    level: info
  rate_limit:
    enabled: true
    requests_per_minute: 1000
  waf:
    enabled: true
  monitoring:
    enabled: true

secrets:
  aesKey: "your-aes-key"
  jwtSecret: "your-jwt-secret"
```

Deploy with Helm:

```bash
# Add Helm repository (if using external charts)
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Install Security Gateway
helm install security-gateway ./helm/security-gateway \
  --namespace go-coffee-security \
  --create-namespace \
  --values helm/security-gateway/values.yaml
```

### 4. Cloud Provider Deployments

#### AWS EKS Deployment

1. **Create EKS cluster**:
```bash
eksctl create cluster \
  --name go-coffee-security \
  --region us-west-2 \
  --nodegroup-name workers \
  --node-type m5.large \
  --nodes 3 \
  --nodes-min 1 \
  --nodes-max 10
```

2. **Install AWS Load Balancer Controller**:
```bash
kubectl apply -k "github.com/aws/eks-charts/stable/aws-load-balancer-controller//crds?ref=master"
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  --set clusterName=go-coffee-security \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller \
  -n kube-system
```

3. **Deploy with ALB Ingress**:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: security-gateway-alb
  namespace: go-coffee-security
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-west-2:123456789:certificate/your-cert
spec:
  rules:
  - host: security.go-coffee.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: security-gateway
            port:
              number: 80
```

#### Google GKE Deployment

1. **Create GKE cluster**:
```bash
gcloud container clusters create go-coffee-security \
  --zone us-central1-a \
  --num-nodes 3 \
  --enable-autoscaling \
  --min-nodes 1 \
  --max-nodes 10 \
  --machine-type n1-standard-2
```

2. **Get credentials**:
```bash
gcloud container clusters get-credentials go-coffee-security --zone us-central1-a
```

3. **Deploy with GCE Ingress**:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: security-gateway-gce
  namespace: go-coffee-security
  annotations:
    kubernetes.io/ingress.class: gce
    kubernetes.io/ingress.global-static-ip-name: security-gateway-ip
    networking.gke.io/managed-certificates: security-gateway-ssl-cert
spec:
  rules:
  - host: security.go-coffee.com
    http:
      paths:
      - path: /*
        pathType: ImplementationSpecific
        backend:
          service:
            name: security-gateway
            port:
              number: 80
```

## Monitoring and Observability

### Prometheus Monitoring

Deploy Prometheus to monitor the Security Gateway:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: go-coffee-security
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
    - job_name: 'security-gateway'
      static_configs:
      - targets: ['security-gateway:8080']
      metrics_path: '/metrics'
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: go-coffee-security
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      containers:
      - name: prometheus
        image: prom/prometheus:latest
        ports:
        - containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        args:
          - '--config.file=/etc/prometheus/prometheus.yml'
          - '--storage.tsdb.path=/prometheus'
          - '--web.console.libraries=/etc/prometheus/console_libraries'
          - '--web.console.templates=/etc/prometheus/consoles'
      volumes:
      - name: config
        configMap:
          name: prometheus-config
```

### Grafana Dashboard

Deploy Grafana with pre-configured dashboards:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
  namespace: go-coffee-security
data:
  security-gateway.json: |
    {
      "dashboard": {
        "title": "Security Gateway Dashboard",
        "panels": [
          {
            "title": "Request Rate",
            "type": "graph",
            "targets": [
              {
                "expr": "rate(security_gateway_total_requests[5m])"
              }
            ]
          }
        ]
      }
    }
```

### Logging with ELK Stack

Deploy Elasticsearch, Logstash, and Kibana for centralized logging:

```bash
# Add Elastic Helm repository
helm repo add elastic https://helm.elastic.co
helm repo update

# Install Elasticsearch
helm install elasticsearch elastic/elasticsearch \
  --namespace go-coffee-security \
  --set replicas=1 \
  --set minimumMasterNodes=1

# Install Kibana
helm install kibana elastic/kibana \
  --namespace go-coffee-security

# Install Filebeat for log collection
helm install filebeat elastic/filebeat \
  --namespace go-coffee-security
```

## Security Considerations

### Network Security

1. **Use TLS/SSL**: Always enable TLS in production
2. **Network Policies**: Implement Kubernetes network policies
3. **Firewall Rules**: Configure appropriate firewall rules
4. **VPC/Subnet Isolation**: Use private subnets for backend services

### Secrets Management

1. **Kubernetes Secrets**: Use Kubernetes secrets for sensitive data
2. **External Secret Managers**: Consider AWS Secrets Manager, HashiCorp Vault
3. **Secret Rotation**: Implement automatic secret rotation
4. **Encryption at Rest**: Enable encryption for persistent volumes

### Access Control

1. **RBAC**: Implement Role-Based Access Control
2. **Service Accounts**: Use dedicated service accounts
3. **Pod Security Policies**: Enforce pod security standards
4. **Network Policies**: Restrict network access between pods

## Scaling and Performance

### Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: security-gateway-hpa
  namespace: go-coffee-security
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: security-gateway
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Vertical Pod Autoscaling

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: security-gateway-vpa
  namespace: go-coffee-security
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: security-gateway
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: security-gateway
      maxAllowed:
        cpu: 2
        memory: 4Gi
      minAllowed:
        cpu: 100m
        memory: 128Mi
```

## Backup and Disaster Recovery

### Redis Backup

```bash
# Create Redis backup job
kubectl create job redis-backup --image=redis:7-alpine \
  --namespace go-coffee-security \
  -- redis-cli -h redis BGSAVE

# Schedule regular backups with CronJob
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-backup
  namespace: go-coffee-security
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: redis:7-alpine
            command:
            - /bin/sh
            - -c
            - |
              redis-cli -h redis BGSAVE
              sleep 10
              redis-cli -h redis LASTSAVE
          restartPolicy: OnFailure
```

### Configuration Backup

```bash
# Backup Kubernetes resources
kubectl get all,configmaps,secrets,ingress,pvc \
  -n go-coffee-security \
  -o yaml > security-gateway-backup.yaml

# Restore from backup
kubectl apply -f security-gateway-backup.yaml
```

## Troubleshooting

### Common Issues

1. **Pod not starting**: Check resource limits and node capacity
2. **Service unreachable**: Verify service and ingress configuration
3. **High memory usage**: Adjust Redis and application memory limits
4. **Rate limiting not working**: Check Redis connectivity
5. **SSL/TLS issues**: Verify certificate configuration

### Debug Commands

```bash
# Check pod status
kubectl get pods -n go-coffee-security

# View pod logs
kubectl logs -f deployment/security-gateway -n go-coffee-security

# Describe pod for events
kubectl describe pod <pod-name> -n go-coffee-security

# Execute into pod for debugging
kubectl exec -it <pod-name> -n go-coffee-security -- /bin/sh

# Check service endpoints
kubectl get endpoints -n go-coffee-security

# Test connectivity
kubectl run test-pod --image=curlimages/curl -it --rm -- /bin/sh
```

### Performance Tuning

1. **Resource Limits**: Adjust CPU and memory limits based on usage
2. **Redis Configuration**: Tune Redis memory and persistence settings
3. **Connection Pooling**: Optimize connection pool sizes
4. **Caching**: Enable and tune application-level caching
5. **Load Balancing**: Configure proper load balancing algorithms

---

For more information, see the [Configuration Reference](../configuration/security-gateway.md) and [API Documentation](../api/security-gateway.md).
