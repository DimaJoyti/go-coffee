# 🚀 Deployment Guide

Complete deployment guide for the Web3 DeFi Algorithmic Trading Platform across different environments.

## 🎯 Deployment Overview

### Supported Platforms
- **🐳 Docker** - Containerized deployment
- **☸️ Kubernetes** - Container orchestration
- **☁️ Cloud Providers** - AWS, GCP, Azure
- **🖥️ Bare Metal** - Direct server deployment

### Environment Types
- **Development** - Local development environment
- **Staging** - Pre-production testing
- **Production** - Live trading environment

## 🐳 Docker Deployment

### Prerequisites

```bash
# Install Docker and Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### Development Deployment

```bash
# Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee/web3-wallet-backend

# Copy environment file
cp .env.example .env
# Edit .env with your configuration

# Start development environment
docker-compose up --build

# Run in background
docker-compose up -d
```

### Production Deployment

```bash
# Use production configuration
docker-compose -f docker-compose.prod.yml up -d

# Scale services
docker-compose -f docker-compose.prod.yml up -d --scale app=3

# View logs
docker-compose logs -f app

# Monitor resources
docker stats
```

### Docker Compose Configuration

**Development (`docker-compose.yml`):**
```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - ENV=development
      - DATABASE_URL=postgres://postgres:password@postgres:5432/defi_trading
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis
    volumes:
      - ./configs:/app/configs
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: defi_trading
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml

volumes:
  postgres_data:
  redis_data:
```

**Production (`docker-compose.prod.yml`):**
```yaml
version: '3.8'

services:
  app:
    image: defi-trading:latest
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 2G
          cpus: '1.0'
        reservations:
          memory: 1G
          cpus: '0.5'
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
    depends_on:
      - app
    restart: always
```

## ☸️ Kubernetes Deployment

### Prerequisites

```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### Namespace Setup

```bash
# Create namespace
kubectl create namespace defi-trading

# Set default namespace
kubectl config set-context --current --namespace=defi-trading
```

### ConfigMap and Secrets

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: defi-trading-config
  namespace: defi-trading
data:
  config.yaml: |
    server:
      port: 8080
    database:
      max_connections: 100
    trading:
      enabled: true
      max_position_size: 50000
---
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: defi-trading-secrets
  namespace: defi-trading
type: Opaque
stringData:
  database-url: "postgres://user:pass@postgres:5432/defi_trading"
  jwt-secret: "your-super-secret-jwt-key"
  encryption-key: "your-32-byte-encryption-key"
```

### Application Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: defi-trading-app
  namespace: defi-trading
spec:
  replicas: 3
  selector:
    matchLabels:
      app: defi-trading-app
  template:
    metadata:
      labels:
        app: defi-trading-app
    spec:
      containers:
      - name: app
        image: defi-trading:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: defi-trading-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: defi-trading-secrets
              key: jwt-secret
        volumeMounts:
        - name: config
          mountPath: /app/configs
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: defi-trading-config
---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: defi-trading-service
  namespace: defi-trading
spec:
  selector:
    app: defi-trading-app
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
---
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: defi-trading-ingress
  namespace: defi-trading
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.defi-trading.com
    secretName: defi-trading-tls
  rules:
  - host: api.defi-trading.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: defi-trading-service
            port:
              number: 80
```

### Database Deployment

```yaml
# postgres.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: defi-trading
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15
        env:
        - name: POSTGRES_DB
          value: defi_trading
        - name: POSTGRES_USER
          value: postgres
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
```

### Deployment Commands

```bash
# Apply all configurations
kubectl apply -f deployments/kubernetes/

# Check deployment status
kubectl get pods -l app=defi-trading-app

# View logs
kubectl logs -f deployment/defi-trading-app

# Scale deployment
kubectl scale deployment defi-trading-app --replicas=5

# Update deployment
kubectl set image deployment/defi-trading-app app=defi-trading:v2.0.0

# Rollback deployment
kubectl rollout undo deployment/defi-trading-app
```

## ☁️ Cloud Provider Deployment

### AWS EKS Deployment

```bash
# Install eksctl
curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin

# Create EKS cluster
eksctl create cluster \
  --name defi-trading-cluster \
  --region us-west-2 \
  --nodegroup-name workers \
  --node-type m5.large \
  --nodes 3 \
  --nodes-min 1 \
  --nodes-max 10 \
  --managed

# Deploy application
kubectl apply -f deployments/kubernetes/
```

### GCP GKE Deployment

```bash
# Install gcloud CLI
curl https://sdk.cloud.google.com | bash

# Create GKE cluster
gcloud container clusters create defi-trading-cluster \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type n1-standard-2 \
  --enable-autoscaling \
  --min-nodes 1 \
  --max-nodes 10

# Get credentials
gcloud container clusters get-credentials defi-trading-cluster --zone us-central1-a

# Deploy application
kubectl apply -f deployments/kubernetes/
```

### Azure AKS Deployment

```bash
# Install Azure CLI
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# Create resource group
az group create --name defi-trading-rg --location eastus

# Create AKS cluster
az aks create \
  --resource-group defi-trading-rg \
  --name defi-trading-cluster \
  --node-count 3 \
  --enable-addons monitoring \
  --generate-ssh-keys

# Get credentials
az aks get-credentials --resource-group defi-trading-rg --name defi-trading-cluster

# Deploy application
kubectl apply -f deployments/kubernetes/
```

## 🔧 Production Checklist

### Pre-Deployment

- [ ] **Environment Configuration**
  - [ ] Environment variables configured
  - [ ] Secrets properly managed
  - [ ] SSL certificates installed
  - [ ] Database migrations applied

- [ ] **Security**
  - [ ] Security audit completed
  - [ ] Penetration testing passed
  - [ ] Access controls configured
  - [ ] Backup strategy implemented

- [ ] **Performance**
  - [ ] Load testing completed
  - [ ] Resource limits configured
  - [ ] Auto-scaling configured
  - [ ] Monitoring stack deployed

### Post-Deployment

- [ ] **Verification**
  - [ ] Health checks passing
  - [ ] API endpoints responding
  - [ ] Trading functionality working
  - [ ] Monitoring alerts configured

- [ ] **Documentation**
  - [ ] Deployment documentation updated
  - [ ] Runbooks created
  - [ ] Team training completed
  - [ ] Incident response plan ready

## 📊 Monitoring and Observability

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'defi-trading'
    static_configs:
      - targets: ['defi-trading-service:8080']
    metrics_path: /metrics
    scrape_interval: 5s
```

### Grafana Dashboards

```bash
# Import pre-built dashboards
kubectl apply -f monitoring/grafana-dashboards.yaml

# Access Grafana
kubectl port-forward svc/grafana 3000:80
```

### Log Aggregation

```yaml
# fluentd-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/*defi-trading*.log
      pos_file /var/log/fluentd-containers.log.pos
      tag kubernetes.*
      format json
    </source>
    
    <match kubernetes.**>
      @type elasticsearch
      host elasticsearch.logging.svc.cluster.local
      port 9200
      index_name defi-trading
    </match>
```

## 🔄 CI/CD Pipeline

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    - run: go test ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Build Docker image
      run: |
        docker build -t defi-trading:${{ github.sha }} .
        docker tag defi-trading:${{ github.sha }} defi-trading:latest
    - name: Push to registry
      run: |
        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
        docker push defi-trading:${{ github.sha }}
        docker push defi-trading:latest

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Deploy to Kubernetes
      run: |
        echo ${{ secrets.KUBECONFIG }} | base64 -d > kubeconfig
        export KUBECONFIG=kubeconfig
        kubectl set image deployment/defi-trading-app app=defi-trading:${{ github.sha }}
        kubectl rollout status deployment/defi-trading-app
```

## 🆘 Troubleshooting

### Common Issues

1. **Pod CrashLoopBackOff**
   ```bash
   kubectl describe pod <pod-name>
   kubectl logs <pod-name> --previous
   ```

2. **Service Not Accessible**
   ```bash
   kubectl get svc
   kubectl describe svc defi-trading-service
   ```

3. **Database Connection Issues**
   ```bash
   kubectl exec -it <pod-name> -- psql $DATABASE_URL
   ```

### Health Checks

```bash
# Application health
curl http://api.defi-trading.com/health

# Database health
kubectl exec -it postgres-0 -- pg_isready

# Redis health
kubectl exec -it redis-0 -- redis-cli ping
```

---

**🚀 Need help with deployment? Check our [troubleshooting guide](TROUBLESHOOTING.md) or contact support!**
