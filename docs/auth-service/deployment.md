# 🚀 Auth Service Deployment Guide

<div align="center">

![Docker](https://img.shields.io/badge/Docker-Ready-blue?style=for-the-badge&logo=docker)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Native-green?style=for-the-badge&logo=kubernetes)
![Production](https://img.shields.io/badge/Production-Ready-red?style=for-the-badge)

**Complete deployment guide for production environments**

</div>

---

## 🎯 Deployment Overview

The Auth Service supports multiple deployment strategies from local development to enterprise-scale production environments. This guide covers all deployment scenarios with best practices and security considerations.

### 🏗️ Deployment Options

<table>
<tr>
<td width="25%">

**🐳 Docker**
- Single container
- Docker Compose
- Swarm mode
- Easy local development

</td>
<td width="25%">

**☸️ Kubernetes**
- Horizontal scaling
- Service mesh ready
- Cloud native
- Enterprise grade

</td>
<td width="25%">

**☁️ Cloud Platforms**
- AWS ECS/EKS
- Google GKE
- Azure AKS
- Managed services

</td>
<td width="25%">

**🖥️ Traditional**
- Virtual machines
- Bare metal
- System services
- Legacy environments

</td>
</tr>
</table>

---

## 🐳 Docker Deployment

### 📦 Container Image

#### Building the Image

```bash
# Build production image
docker build -f cmd/auth-service/Dockerfile -t go-coffee/auth-service:latest .

# Build with specific version
docker build -f cmd/auth-service/Dockerfile -t go-coffee/auth-service:v1.0.0 .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 \
  -f cmd/auth-service/Dockerfile \
  -t go-coffee/auth-service:latest .
```

#### Image Security Features

<table>
<tr>
<td width="50%">

**🔒 Security Hardening**
- Multi-stage build (minimal final image)
- Non-root user execution
- No shell in final image
- Distroless base image option
- Vulnerability scanning

</td>
<td width="50%">

**📊 Image Details**
- **Base**: Alpine Linux (5MB)
- **Final Size**: ~15MB
- **User**: appuser (UID 1001)
- **Exposed Ports**: 8080, 50053
- **Health Check**: Built-in

</td>
</tr>
</table>

### 🐳 Docker Compose Deployment

#### Production Docker Compose

```yaml
version: '3.8'

services:
  auth-service:
    image: go-coffee/auth-service:latest
    container_name: auth-service-prod
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "50053:50053"
    environment:
      - ENVIRONMENT=production
      - JWT_SECRET=${JWT_SECRET}
      - REDIS_URL=redis://redis:6379
      - LOG_LEVEL=info
    volumes:
      - ./config:/app/config:ro
      - ./logs:/app/logs
    depends_on:
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - auth-network
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

  redis:
    image: redis:7-alpine
    container_name: redis-prod
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    environment:
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
      - ./config/redis.conf:/usr/local/etc/redis/redis.conf:ro
    healthcheck:
      test: ["CMD", "redis-cli", "--no-auth-warning", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - auth-network

  nginx:
    image: nginx:alpine
    container_name: nginx-proxy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/ssl:ro
    depends_on:
      - auth-service
    networks:
      - auth-network

volumes:
  redis_data:
    driver: local

networks:
  auth-network:
    driver: bridge
```

#### Environment Configuration

```bash
# .env file for production
ENVIRONMENT=production
JWT_SECRET=your-super-secure-256-bit-secret-key-here
REDIS_PASSWORD=your-secure-redis-password
LOG_LEVEL=info
REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379

# SSL Configuration
SSL_CERT_PATH=/etc/ssl/certs/auth-service.crt
SSL_KEY_PATH=/etc/ssl/private/auth-service.key

# Monitoring
PROMETHEUS_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

---

## ☸️ Kubernetes Deployment

### 📋 Kubernetes Manifests

#### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: auth-service
  labels:
    name: auth-service
    environment: production
```

#### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-service-config
  namespace: auth-service
data:
  config.yaml: |
    server:
      http_port: 8080
      grpc_port: 50053
      host: "0.0.0.0"
    
    redis:
      url: "redis://redis-service:6379"
      db: 0
      max_retries: 3
      pool_size: 10
    
    security:
      access_token_ttl: "15m"
      refresh_token_ttl: "168h"
      bcrypt_cost: 12
      max_login_attempts: 5
      lockout_duration: "30m"
    
    logging:
      level: "info"
      format: "json"
    
    environment: "production"
```

#### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
  namespace: auth-service
type: Opaque
data:
  jwt-secret: <base64-encoded-jwt-secret>
  redis-password: <base64-encoded-redis-password>
```

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: auth-service
  labels:
    app: auth-service
    version: v1.0.0
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: auth-service
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
      - name: auth-service
        image: go-coffee/auth-service:v1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        - containerPort: 50053
          name: grpc
          protocol: TCP
        - containerPort: 9090
          name: metrics
          protocol: TCP
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: jwt-secret
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: redis-password
        - name: REDIS_URL
          value: "redis://:$(REDIS_PASSWORD)@redis-service:6379"
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
      volumes:
      - name: config
        configMap:
          name: auth-service-config
      nodeSelector:
        kubernetes.io/os: linux
      tolerations:
      - key: "auth-service"
        operator: "Equal"
        value: "true"
        effect: "NoSchedule"
```

#### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: auth-service
  labels:
    app: auth-service
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  - port: 50053
    targetPort: 50053
    protocol: TCP
    name: grpc
  selector:
    app: auth-service
```

#### HorizontalPodAutoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: auth-service-hpa
  namespace: auth-service
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: auth-service
  minReplicas: 3
  maxReplicas: 10
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
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

### 🔧 Deployment Commands

```bash
# Apply all manifests
kubectl apply -f k8s/auth-service/

# Check deployment status
kubectl get pods -n auth-service
kubectl get services -n auth-service

# View logs
kubectl logs -f deployment/auth-service -n auth-service

# Scale deployment
kubectl scale deployment auth-service --replicas=5 -n auth-service

# Rolling update
kubectl set image deployment/auth-service auth-service=go-coffee/auth-service:v1.1.0 -n auth-service

# Check rollout status
kubectl rollout status deployment/auth-service -n auth-service
```

---

## ☁️ Cloud Platform Deployment

### 🌩️ AWS EKS Deployment

#### EKS Cluster Setup

```bash
# Create EKS cluster
eksctl create cluster \
  --name go-coffee-cluster \
  --version 1.28 \
  --region us-west-2 \
  --nodegroup-name auth-service-nodes \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 2 \
  --nodes-max 10 \
  --managed

# Configure kubectl
aws eks update-kubeconfig --region us-west-2 --name go-coffee-cluster
```

#### AWS Load Balancer Controller

```bash
# Install AWS Load Balancer Controller
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=go-coffee-cluster \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

#### Ingress Configuration

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: auth-service-ingress
  namespace: auth-service
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-west-2:123456789012:certificate/12345678-1234-1234-1234-123456789012
    alb.ingress.kubernetes.io/ssl-redirect: '443'
spec:
  rules:
  - host: auth.go-coffee.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: auth-service
            port:
              number: 80
```

### 🔵 Google GKE Deployment

```bash
# Create GKE cluster
gcloud container clusters create go-coffee-cluster \
  --zone us-central1-a \
  --num-nodes 3 \
  --enable-autoscaling \
  --min-nodes 2 \
  --max-nodes 10 \
  --machine-type e2-standard-2 \
  --enable-autorepair \
  --enable-autoupgrade

# Get credentials
gcloud container clusters get-credentials go-coffee-cluster --zone us-central1-a
```

### 🔷 Azure AKS Deployment

```bash
# Create resource group
az group create --name go-coffee-rg --location eastus

# Create AKS cluster
az aks create \
  --resource-group go-coffee-rg \
  --name go-coffee-cluster \
  --node-count 3 \
  --enable-addons monitoring \
  --generate-ssh-keys

# Get credentials
az aks get-credentials --resource-group go-coffee-rg --name go-coffee-cluster
```

---

## 📊 Monitoring & Observability

### 🔍 Prometheus Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    
    scrape_configs:
    - job_name: 'auth-service'
      static_configs:
      - targets: ['auth-service:9090']
      metrics_path: /metrics
      scrape_interval: 10s
```

### 📈 Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Auth Service Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(auth_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(auth_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

---

## 🔒 Production Security

### 🛡️ Security Checklist

<table>
<tr>
<td width="50%">

**✅ Network Security**
- [ ] TLS/SSL certificates configured
- [ ] Network policies implemented
- [ ] Firewall rules configured
- [ ] VPN/Private networks
- [ ] Load balancer security groups

</td>
<td width="50%">

**✅ Container Security**
- [ ] Non-root user execution
- [ ] Read-only root filesystem
- [ ] Security contexts configured
- [ ] Image vulnerability scanning
- [ ] Resource limits set

</td>
</tr>
<tr>
<td width="50%">

**✅ Secrets Management**
- [ ] Kubernetes secrets for sensitive data
- [ ] External secret management (Vault)
- [ ] Secret rotation policies
- [ ] No secrets in images
- [ ] Environment variable encryption

</td>
<td width="50%">

**✅ Access Control**
- [ ] RBAC configured
- [ ] Service accounts with minimal permissions
- [ ] Pod security policies
- [ ] Network segmentation
- [ ] Audit logging enabled

</td>
</tr>
</table>

### 🔐 Secret Management

```bash
# Create secrets from files
kubectl create secret generic auth-service-secrets \
  --from-file=jwt-secret=./secrets/jwt-secret.txt \
  --from-file=redis-password=./secrets/redis-password.txt \
  -n auth-service

# Using external secret management (Vault)
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
  namespace: auth-service
spec:
  provider:
    vault:
      server: "https://vault.company.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "auth-service"
EOF
```

---

## 🚀 CI/CD Pipeline

### 🔄 GitHub Actions Workflow

```yaml
name: Deploy Auth Service

on:
  push:
    branches: [main]
    paths: ['cmd/auth-service/**', 'internal/auth/**']

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.24
    - run: make -f Makefile.auth test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Build and push Docker image
      uses: docker/build-push-action@v3
      with:
        context: .
        file: cmd/auth-service/Dockerfile
        push: true
        tags: |
          go-coffee/auth-service:latest
          go-coffee/auth-service:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
    - uses: actions/checkout@v3
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/auth-service \
          auth-service=go-coffee/auth-service:${{ github.sha }} \
          -n auth-service
```

---

## 📋 Troubleshooting

### 🔍 Common Issues

<table>
<tr>
<td width="50%">

**🚨 Service Won't Start**
```bash
# Check logs
kubectl logs deployment/auth-service -n auth-service

# Check events
kubectl get events -n auth-service

# Check configuration
kubectl describe configmap auth-service-config -n auth-service
```

</td>
<td width="50%">

**🔌 Redis Connection Issues**
```bash
# Test Redis connectivity
kubectl exec -it deployment/auth-service -n auth-service -- \
  wget -qO- http://localhost:8080/health

# Check Redis service
kubectl get svc redis-service -n auth-service
```

</td>
</tr>
</table>

### 📊 Performance Tuning

```yaml
# Resource optimization
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"

# JVM tuning for Go (if applicable)
env:
- name: GOGC
  value: "100"
- name: GOMAXPROCS
  value: "2"
```

---

<div align="center">

**🚀 Deployment Documentation**

[🏠 Main README](./README.md) • [📖 API Reference](./api-reference.md) • [🏗️ Architecture](./architecture.md) • [🛡️ Security](./security.md)

</div>
