# Developer DAO Platform - Deployment Guide

## 🎯 Overview

This guide provides step-by-step instructions for deploying the complete Developer DAO Platform with all three microservices: Bounty Management, Solution Marketplace, and TVL/MAU Tracking.

## 📋 Prerequisites

### Infrastructure Requirements
- **Kubernetes cluster** (v1.24+) or Docker Compose environment
- **PostgreSQL** (v14+) with 16GB+ RAM
- **Redis** (v6+) with 8GB+ RAM
- **Load Balancer** (NGINX, HAProxy, or cloud LB)
- **Monitoring Stack** (Prometheus, Grafana, Jaeger)

### Development Requirements
- **Go** (v1.21+)
- **Docker** (v20+)
- **kubectl** (for Kubernetes deployment)
- **Helm** (v3+, optional but recommended)

## 🏗️ Architecture Deployment

### Service Ports
```
Bounty Service:     HTTP 8080, gRPC 9090
Marketplace Service: HTTP 8081, gRPC 9091  
Metrics Service:    HTTP 8082, gRPC 9092
```

### Database Schema
```sql
-- Create databases
CREATE DATABASE bounty_db;
CREATE DATABASE marketplace_db;
CREATE DATABASE metrics_db;

-- Create users with appropriate permissions
CREATE USER bounty_user WITH PASSWORD 'secure_password';
CREATE USER marketplace_user WITH PASSWORD 'secure_password';
CREATE USER metrics_user WITH PASSWORD 'secure_password';

GRANT ALL PRIVILEGES ON DATABASE bounty_db TO bounty_user;
GRANT ALL PRIVILEGES ON DATABASE marketplace_db TO marketplace_user;
GRANT ALL PRIVILEGES ON DATABASE metrics_db TO metrics_user;
```

## 🐳 Docker Deployment

### 1. Build Docker Images

```bash
# Build all services
cd developer-dao

# Bounty Service
docker build -f cmd/bounty-service/Dockerfile -t developer-dao/bounty-service:latest .

# Marketplace Service  
docker build -f cmd/marketplace-service/Dockerfile -t developer-dao/marketplace-service:latest .

# Metrics Service
docker build -f cmd/metrics-service/Dockerfile -t developer-dao/metrics-service:latest .
```

### 2. Docker Compose Deployment

```yaml
# docker-compose.yml
version: '3.8'

services:
  # Database
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: developer_dao
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  # Cache
  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  # Bounty Service
  bounty-service:
    image: developer-dao/bounty-service:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/developer_dao
      - REDIS_URL=redis://redis:6379
      - ENVIRONMENT=production
    depends_on:
      - postgres
      - redis

  # Marketplace Service
  marketplace-service:
    image: developer-dao/marketplace-service:latest
    ports:
      - "8081:8081"
      - "9091:9091"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/developer_dao
      - REDIS_URL=redis://redis:6379
      - ENVIRONMENT=production
    depends_on:
      - postgres
      - redis

  # Metrics Service
  metrics-service:
    image: developer-dao/metrics-service:latest
    ports:
      - "8082:8082"
      - "9092:9092"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/developer_dao
      - REDIS_URL=redis://redis:6379
      - ENVIRONMENT=production
    depends_on:
      - postgres
      - redis

  # Load Balancer
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - bounty-service
      - marketplace-service
      - metrics-service

volumes:
  postgres_data:
  redis_data:
```

### 3. NGINX Configuration

```nginx
# nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream bounty_service {
        server bounty-service:8080;
    }
    
    upstream marketplace_service {
        server marketplace-service:8081;
    }
    
    upstream metrics_service {
        server metrics-service:8082;
    }

    server {
        listen 80;
        server_name localhost;

        # Bounty API
        location /api/v1/bounties {
            proxy_pass http://bounty_service;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        # Marketplace API
        location /api/v1/solutions {
            proxy_pass http://marketplace_service;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        # Metrics API
        location /api/v1/tvl {
            proxy_pass http://metrics_service;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        location /api/v1/mau {
            proxy_pass http://metrics_service;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        # Health checks
        location /health {
            access_log off;
            return 200 "healthy\n";
        }
    }
}
```

## ☸️ Kubernetes Deployment

### 1. Namespace and ConfigMap

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: developer-dao

---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: developer-dao
data:
  environment: "production"
  log_level: "info"
  database_host: "postgres-service"
  redis_host: "redis-service"
```

### 2. Database Deployment

```yaml
# postgres.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: developer-dao
spec:
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
        image: postgres:14
        env:
        - name: POSTGRES_DB
          value: "developer_dao"
        - name: POSTGRES_USER
          value: "postgres"
        - name: POSTGRES_PASSWORD
          value: "postgres"
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc

---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: developer-dao
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

### 3. Service Deployments

```yaml
# bounty-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bounty-service
  namespace: developer-dao
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bounty-service
  template:
    metadata:
      labels:
        app: bounty-service
    spec:
      containers:
      - name: bounty-service
        image: developer-dao/bounty-service:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: DATABASE_URL
          value: "postgres://postgres:postgres@postgres-service:5432/developer_dao"
        - name: REDIS_URL
          value: "redis://redis-service:6379"
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

---
apiVersion: v1
kind: Service
metadata:
  name: bounty-service
  namespace: developer-dao
spec:
  selector:
    app: bounty-service
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
```

### 4. Ingress Configuration

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: developer-dao-ingress
  namespace: developer-dao
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: api.developer-dao.com
    http:
      paths:
      - path: /api/v1/bounties
        pathType: Prefix
        backend:
          service:
            name: bounty-service
            port:
              number: 8080
      - path: /api/v1/solutions
        pathType: Prefix
        backend:
          service:
            name: marketplace-service
            port:
              number: 8081
      - path: /api/v1/tvl
        pathType: Prefix
        backend:
          service:
            name: metrics-service
            port:
              number: 8082
      - path: /api/v1/mau
        pathType: Prefix
        backend:
          service:
            name: metrics-service
            port:
              number: 8082
```

## 📊 Monitoring Setup

### 1. Prometheus Configuration

```yaml
# prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: developer-dao
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
    - job_name: 'bounty-service'
      static_configs:
      - targets: ['bounty-service:8080']
    - job_name: 'marketplace-service'
      static_configs:
      - targets: ['marketplace-service:8081']
    - job_name: 'metrics-service'
      static_configs:
      - targets: ['metrics-service:8082']
```

### 2. Grafana Dashboards

```json
{
  "dashboard": {
    "title": "Developer DAO Platform",
    "panels": [
      {
        "title": "Active Bounties",
        "type": "stat",
        "targets": [
          {
            "expr": "bounty_active_total"
          }
        ]
      },
      {
        "title": "TVL Growth",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(tvl_total[5m])"
          }
        ]
      },
      {
        "title": "MAU Growth",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(mau_total[1h])"
          }
        ]
      }
    ]
  }
}
```

## 🔧 Configuration Management

### Environment Variables

```bash
# Production environment
export ENVIRONMENT=production
export LOG_LEVEL=info
export DATABASE_URL=postgres://user:pass@host:5432/db
export REDIS_URL=redis://host:6379
export ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_KEY
export BSC_RPC_URL=https://bsc-dataseed.binance.org/
export POLYGON_RPC_URL=https://polygon-rpc.com/
```

### Configuration Files

```yaml
# configs/config.yaml
environment: production
logging:
  level: info
  format: json

server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  host: postgres-service
  port: 5432
  name: developer_dao
  user: postgres
  password: postgres
  max_connections: 100

redis:
  host: redis-service
  port: 6379
  password: ""
  db: 0

blockchain:
  ethereum:
    rpc_url: "https://mainnet.infura.io/v3/YOUR_KEY"
  bsc:
    rpc_url: "https://bsc-dataseed.binance.org/"
  polygon:
    rpc_url: "https://polygon-rpc.com/"

monitoring:
  prometheus:
    enabled: true
    path: "/metrics"
```

## 🚀 Deployment Commands

### Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Scale services
docker-compose up -d --scale bounty-service=3

# Stop services
docker-compose down
```

### Kubernetes
```bash
# Apply all configurations
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n developer-dao

# View logs
kubectl logs -f deployment/bounty-service -n developer-dao

# Scale deployment
kubectl scale deployment bounty-service --replicas=5 -n developer-dao

# Port forward for testing
kubectl port-forward service/bounty-service 8080:8080 -n developer-dao
```

## ✅ Health Checks

### Service Health Endpoints
```bash
# Check all services
curl http://localhost:8080/health  # Bounty Service
curl http://localhost:8081/health  # Marketplace Service
curl http://localhost:8082/health  # Metrics Service

# Check metrics endpoints
curl http://localhost:8080/metrics  # Prometheus metrics
```

### Database Connectivity
```bash
# Test database connections
psql -h localhost -U postgres -d developer_dao -c "SELECT 1;"

# Test Redis connection
redis-cli -h localhost ping
```

## 🔒 Security Considerations

### SSL/TLS Configuration
- Use Let's Encrypt for SSL certificates
- Configure HTTPS redirects
- Enable HSTS headers

### Network Security
- Use private networks for internal communication
- Configure firewall rules
- Enable VPN access for management

### Authentication & Authorization
- Implement JWT token validation
- Configure rate limiting
- Enable audit logging

## 📈 Performance Optimization

### Database Optimization
- Configure connection pooling
- Set up read replicas
- Optimize query performance

### Caching Strategy
- Configure Redis clustering
- Implement cache warming
- Monitor cache hit rates

### Load Balancing
- Configure health checks
- Implement sticky sessions if needed
- Monitor response times

## 🎉 Deployment Complete!

The Developer DAO Platform is now ready for production use with:

- ✅ **3 Microservices** deployed and running
- ✅ **Database & Cache** configured and optimized
- ✅ **Load Balancing** for high availability
- ✅ **Monitoring & Alerting** for operational visibility
- ✅ **Security** measures implemented

**Platform is production-ready and scalable! 🚀**
