# Developer DAO Platform - Complete Deployment Guide

## 🎯 Overview

This guide provides comprehensive instructions for deploying the complete Developer DAO Platform, including all backend services and frontend applications.

## 🏗️ Complete Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Developer DAO Platform                       │
├─────────────────────────────────────────────────────────────────┤
│                        Frontend Layer                           │
│  Developer Portal (3000)  │  Governance UI (3001)              │
├─────────────────────────────────────────────────────────────────┤
│                        Backend Services                         │
│  Bounty Service (8080)  │  Marketplace (8081)  │  Metrics (8082)│
├─────────────────────────────────────────────────────────────────┤
│                      Infrastructure Layer                       │
│  PostgreSQL (5432)  │  Redis (6379)  │  NGINX (80/443)        │
└─────────────────────────────────────────────────────────────────┘
```

## 📋 Prerequisites

### System Requirements
- **Docker** 20.10+ and Docker Compose 2.0+
- **Node.js** 18+ (for local development)
- **Go** 1.21+ (for backend development)
- **PostgreSQL** 14+ (if running separately)
- **Redis** 6+ (if running separately)

### Environment Setup
- Minimum 8GB RAM
- 50GB+ disk space
- Network access for external API integrations

## 🚀 Quick Start (Docker Compose)

### 1. Clone and Setup
```bash
# Clone the repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee/developer-dao

# Create environment file
cp .env.example .env
```

### 2. Configure Environment Variables
```bash
# .env file
DATABASE_URL=postgres://postgres:postgres@postgres:5432/developer_dao
REDIS_URL=redis://redis:6379
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_KEY
BSC_RPC_URL=https://bsc-dataseed.binance.org/
POLYGON_RPC_URL=https://polygon-rpc.com/
WALLET_CONNECT_PROJECT_ID=your_wallet_connect_project_id
```

### 3. Build and Deploy
```bash
# Build all backend services
docker build -f cmd/bounty-service/Dockerfile -t developer-dao/bounty-service:latest .
docker build -f cmd/marketplace-service/Dockerfile -t developer-dao/marketplace-service:latest .
docker build -f cmd/metrics-service/Dockerfile -t developer-dao/metrics-service:latest .

# Deploy complete platform
cd web
docker-compose up -d
```

### 4. Verify Deployment
```bash
# Check all services
curl http://localhost:8080/health  # Bounty Service
curl http://localhost:8081/health  # Marketplace Service
curl http://localhost:8082/health  # Metrics Service
curl http://localhost:3000         # Developer Portal
curl http://localhost:3001         # Governance UI
```

## 🔧 Service-by-Service Deployment

### Backend Services

#### 1. Database Setup
```bash
# Start PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_DB=developer_dao \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:14

# Start Redis
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:6-alpine
```

#### 2. Bounty Service
```bash
# Build and run
cd developer-dao
docker build -f cmd/bounty-service/Dockerfile -t bounty-service .
docker run -d \
  --name bounty-service \
  -p 8080:8080 \
  -e DATABASE_URL=postgres://postgres:postgres@host.docker.internal:5432/developer_dao \
  -e REDIS_URL=redis://host.docker.internal:6379 \
  bounty-service
```

#### 3. Marketplace Service
```bash
# Build and run
docker build -f cmd/marketplace-service/Dockerfile -t marketplace-service .
docker run -d \
  --name marketplace-service \
  -p 8081:8081 \
  -e DATABASE_URL=postgres://postgres:postgres@host.docker.internal:5432/developer_dao \
  -e REDIS_URL=redis://host.docker.internal:6379 \
  marketplace-service
```

#### 4. Metrics Service
```bash
# Build and run
docker build -f cmd/metrics-service/Dockerfile -t metrics-service .
docker run -d \
  --name metrics-service \
  -p 8082:8082 \
  -e DATABASE_URL=postgres://postgres:postgres@host.docker.internal:5432/developer_dao \
  -e REDIS_URL=redis://host.docker.internal:6379 \
  metrics-service
```

### Frontend Applications

#### 1. Developer Portal
```bash
cd web/dao-portal

# Install dependencies
npm install

# Build for production
npm run build

# Start production server
npm start
# OR with Docker
docker build -t dao-portal .
docker run -d -p 3000:3000 dao-portal
```

#### 2. Governance UI
```bash
cd web/governance-ui

# Install dependencies
npm install

# Build for production
npm run build

# Start production server
npm start
# OR with Docker
docker build -t governance-ui .
docker run -d -p 3001:3001 governance-ui
```

## ☸️ Kubernetes Deployment

### 1. Namespace and ConfigMaps
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: developer-dao

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: developer-dao
data:
  DATABASE_URL: "postgres://postgres:postgres@postgres-service:5432/developer_dao"
  REDIS_URL: "redis://redis-service:6379"
```

### 2. Database Deployments
```yaml
# k8s/postgres.yaml
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
```

### 3. Backend Services
```yaml
# k8s/bounty-service.yaml
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
        envFrom:
        - configMapRef:
            name: app-config
```

### 4. Frontend Services
```yaml
# k8s/dao-portal.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dao-portal
  namespace: developer-dao
spec:
  replicas: 2
  selector:
    matchLabels:
      app: dao-portal
  template:
    metadata:
      labels:
        app: dao-portal
    spec:
      containers:
      - name: dao-portal
        image: developer-dao/dao-portal:latest
        ports:
        - containerPort: 3000
        env:
        - name: NEXT_PUBLIC_API_URL
          value: "http://api.developer-dao.com"
```

### 5. Ingress Configuration
```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: developer-dao-ingress
  namespace: developer-dao
spec:
  rules:
  - host: portal.developer-dao.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: dao-portal-service
            port:
              number: 3000
  - host: governance.developer-dao.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: governance-ui-service
            port:
              number: 3001
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
```

## 🔒 Production Security

### SSL/TLS Configuration
```bash
# Install Certbot for Let's Encrypt
sudo apt install certbot python3-certbot-nginx

# Generate SSL certificates
sudo certbot --nginx -d portal.developer-dao.com
sudo certbot --nginx -d governance.developer-dao.com
sudo certbot --nginx -d api.developer-dao.com
```

### Environment Security
```bash
# Use Docker secrets for sensitive data
echo "your_database_password" | docker secret create db_password -
echo "your_jwt_secret" | docker secret create jwt_secret -

# Update docker-compose with secrets
services:
  bounty-service:
    secrets:
      - db_password
      - jwt_secret
```

## 📊 Monitoring Setup

### 1. Prometheus Configuration
```yaml
# monitoring/prometheus.yml
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
```bash
# Deploy monitoring stack
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

docker run -d \
  --name grafana \
  -p 3003:3000 \
  grafana/grafana
```

## 🧪 Testing Deployment

### Health Checks
```bash
#!/bin/bash
# health-check.sh

echo "Checking Backend Services..."
curl -f http://localhost:8080/health || exit 1
curl -f http://localhost:8081/health || exit 1
curl -f http://localhost:8082/health || exit 1

echo "Checking Frontend Applications..."
curl -f http://localhost:3000 || exit 1
curl -f http://localhost:3001 || exit 1

echo "Checking Database Connectivity..."
pg_isready -h localhost -p 5432 || exit 1

echo "All services are healthy!"
```

### API Testing
```bash
# Test bounty creation
curl -X POST http://localhost:8080/api/v1/bounties \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Bounty",
    "description": "Test bounty description",
    "category": 0,
    "reward_amount": "1000.00",
    "currency": "USDC",
    "deadline": "2024-12-31T23:59:59Z"
  }'

# Test solution creation
curl -X POST http://localhost:8081/api/v1/solutions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Solution",
    "description": "Test solution description",
    "category": 0,
    "version": "1.0.0",
    "repository_url": "https://github.com/test/solution"
  }'
```

## 🔧 Troubleshooting

### Common Issues

#### 1. Database Connection Issues
```bash
# Check database logs
docker logs postgres

# Test connection
psql -h localhost -U postgres -d developer_dao -c "SELECT 1;"
```

#### 2. Service Discovery Issues
```bash
# Check Docker network
docker network ls
docker network inspect developer-dao_default

# Check service connectivity
docker exec bounty-service ping postgres
```

#### 3. Frontend Build Issues
```bash
# Clear Next.js cache
cd web/dao-portal
rm -rf .next
npm run build

# Check environment variables
echo $NEXT_PUBLIC_API_URL
```

### Performance Optimization
```bash
# Monitor resource usage
docker stats

# Scale services
docker-compose up -d --scale bounty-service=3
docker-compose up -d --scale dao-portal=2
```

## 🎉 Deployment Complete!

The complete Developer DAO Platform is now deployed with:

- ✅ **3 Backend Microservices** (Bounty, Marketplace, Metrics)
- ✅ **2 Frontend Applications** (Developer Portal, Governance UI)
- ✅ **Database & Cache** (PostgreSQL, Redis)
- ✅ **Load Balancing** (NGINX)
- ✅ **Monitoring** (Prometheus, Grafana)
- ✅ **Security** (SSL/TLS, Secrets Management)

**Platform URLs:**
- Developer Portal: http://localhost:3000
- Governance UI: http://localhost:3001
- API Gateway: http://localhost/api/v1
- Monitoring: http://localhost:9090 (Prometheus), http://localhost:3003 (Grafana)

**The Developer DAO Platform is production-ready and scalable! 🚀**
