# 🚀 Go Coffee Platform - Complete Deployment Guide

## 📋 Overview

This guide provides step-by-step instructions for deploying the Go Coffee enterprise platform in various environments, from local development to global production deployment.

## 🎯 Deployment Options

### **1. Local Development**
- Single-machine deployment with Docker Compose
- Ideal for development and testing
- Minimal resource requirements

### **2. Single-Region Production**
- Kubernetes deployment in one region
- Production-ready with monitoring and security
- Suitable for regional businesses

### **3. Multi-Region Enterprise**
- Global deployment across multiple regions
- Disaster recovery and high availability
- Enterprise features and compliance

## 🔧 Prerequisites

### **Required Tools**
```bash
# Core tools
kubectl >= 1.28.0
helm >= 3.12.0
docker >= 24.0.0
docker-compose >= 2.20.0

# Cloud tools (for production)
gcloud >= 440.0.0  # For Google Cloud
aws >= 2.13.0      # For AWS
az >= 2.50.0       # For Azure

# Optional tools
terraform >= 1.5.0  # For infrastructure as code
k6 >= 0.45.0       # For load testing
trivy >= 0.44.0    # For security scanning
```

### **System Requirements**

#### **Local Development**
- CPU: 4+ cores
- RAM: 8GB+
- Storage: 20GB+
- OS: Linux, macOS, or Windows with WSL2

#### **Production (per region)**
- Kubernetes cluster with 3+ nodes
- CPU: 16+ cores total
- RAM: 32GB+ total
- Storage: 100GB+ persistent storage
- Network: Load balancer support

## 🏠 Local Development Deployment

### **Step 1: Clone and Setup**
```bash
# Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Install dependencies
go mod download

# Setup environment
cp .env.example .env
# Edit .env with your configuration
```

### **Step 2: Start Infrastructure**
```bash
# Start infrastructure services
docker-compose -f docker-compose.infrastructure.yml up -d

# Wait for services to be ready
./scripts/wait-for-infrastructure.sh
```

### **Step 3: Build and Start Services**
```bash
# Build all services
./scripts/build-all-services.sh

# Start core services
./scripts/start-core-services.sh

# Verify deployment
./scripts/test-core-services.sh
```

### **Step 4: Access Services**
```bash
# Core services
Producer API:        http://localhost:3000
Web3 Payment:        http://localhost:8083
AI Orchestrator:     http://localhost:8094
Analytics:           http://localhost:8096

# Infrastructure
Kafka UI:            http://localhost:8080
Grafana:             http://localhost:3001
Prometheus:          http://localhost:9090
Jaeger:              http://localhost:16686
```

## ☸️ Single-Region Production Deployment

### **Step 1: Prepare Kubernetes Cluster**
```bash
# Create namespace
kubectl create namespace go-coffee-platform

# Setup RBAC and security
kubectl apply -f k8s/enhanced/security.yaml

# Deploy monitoring
kubectl apply -f k8s/enhanced/monitoring-stack.yaml
```

### **Step 2: Deploy Platform**
```bash
# Deploy with enhanced infrastructure
./scripts/deploy-enhanced-platform.sh \
  --environment production \
  --image-tag v1.0.0

# Or use Helm (recommended)
helm install go-coffee-platform ./helm/go-coffee-platform \
  --namespace go-coffee-platform \
  --values helm/go-coffee-platform/values-production.yaml
```

### **Step 3: Configure Ingress**
```bash
# Deploy ingress controller (if not exists)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

# Configure SSL certificates
kubectl apply -f k8s/enhanced/ssl-certificates.yaml

# Setup custom domain
kubectl apply -f k8s/enhanced/ingress.yaml
```

### **Step 4: Verify Deployment**
```bash
# Run health checks
./scripts/test-enhanced-platform.sh

# Check all pods are running
kubectl get pods -n go-coffee-platform

# Verify services are accessible
curl https://api.your-domain.com/health
```

## 🌍 Multi-Region Enterprise Deployment

### **Step 1: Infrastructure Setup**
```bash
# Deploy infrastructure with Terraform
cd terraform/enterprise
terraform init
terraform plan -var="deployment_type=multi-region"
terraform apply

# Or use cloud-specific scripts
./scripts/setup-gcp-infrastructure.sh  # For Google Cloud
./scripts/setup-aws-infrastructure.sh  # For AWS
./scripts/setup-azure-infrastructure.sh  # For Azure
```

### **Step 2: Deploy to Multiple Regions**
```bash
# Deploy to primary region (us-east-1)
./scripts/deploy-enterprise-platform.sh \
  --deployment-type multi-region \
  --primary-region us-east-1 \
  --environment production

# Deploy to secondary region (us-west-2)
kubectl config use-context us-west-2
./scripts/deploy-enterprise-platform.sh \
  --deployment-type multi-region \
  --primary-region us-west-2 \
  --environment production

# Setup global load balancer
kubectl apply -f k8s/multi-region/global-load-balancer.yaml
```

### **Step 3: Configure Disaster Recovery**
```bash
# Deploy disaster recovery
kubectl apply -f k8s/multi-region/disaster-recovery.yaml

# Setup cross-region replication
./scripts/setup-cross-region-replication.sh

# Test failover
./scripts/test-disaster-recovery.sh
```

### **Step 4: Enable Enterprise Features**
```bash
# Deploy analytics service
kubectl apply -f k8s/enhanced/analytics-service.yaml

# Setup multi-tenancy
./scripts/setup-multi-tenant.sh

# Configure compliance
kubectl apply -f k8s/enhanced/compliance.yaml
```

## 🔧 Configuration

### **Environment Variables**

#### **Core Services**
```bash
# Database
DATABASE_URL=postgres://user:pass@host:5432/go_coffee
REDIS_URL=redis://host:6379/0

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=coffee_orders

# Security
JWT_SECRET=your-super-secret-key
CORS_ALLOWED_ORIGINS=https://your-domain.com
```

#### **Web3 Configuration**
```bash
# Blockchain RPCs
BLOCKCHAIN_ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/your-key
BLOCKCHAIN_BSC_RPC_URL=https://bsc-dataseed.binance.org/
BLOCKCHAIN_POLYGON_RPC_URL=https://polygon-rpc.com/
BLOCKCHAIN_SOLANA_RPC_URL=https://api.mainnet-beta.solana.com

# Supported currencies
WEB3_SUPPORTED_CURRENCIES=["ETH","BNB","MATIC","SOL","USDC","USDT"]
```

#### **AI Configuration**
```bash
# AI API Keys
OPENAI_API_KEY=sk-your-openai-key
GEMINI_API_KEY=your-gemini-key

# AI Orchestrator
AI_ORCHESTRATOR_PORT=8094
AI_KAFKA_TOPIC=ai_agents
AI_ORCHESTRATOR_MAX_TASKS=1000
```

#### **Analytics Configuration**
```bash
# Analytics Service
ANALYTICS_PORT=8096
ANALYTICS_ENABLE_ML_MODELS=true
ANALYTICS_PREDICTION_HORIZON_DAYS=30
ANALYTICS_CONFIDENCE_THRESHOLD=0.8
```

### **Helm Values Configuration**

#### **Production Values (values-production.yaml)**
```yaml
global:
  environment: production
  imageTag: "v1.0.0"
  registry: "ghcr.io/dimajoyti/go-coffee"

replicaCount:
  producer: 3
  consumer: 2
  streams: 2
  web3Payment: 2
  aiOrchestrator: 2
  analytics: 2

resources:
  producer:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilization: 70

monitoring:
  prometheus:
    enabled: true
  grafana:
    enabled: true
  jaeger:
    enabled: true

security:
  networkPolicy:
    enabled: true
  podSecurityPolicy:
    enabled: true
```

## 🧪 Testing & Validation

### **Comprehensive Testing**
```bash
# Run all tests
./scripts/run-comprehensive-tests.sh

# Run specific test categories
./scripts/run-comprehensive-tests.sh --skip-slow  # Skip slow tests
./scripts/run-comprehensive-tests.sh --coverage-threshold 90  # High coverage

# Load testing
./scripts/run-comprehensive-tests.sh --load-users 1000 --load-duration 600s
```

### **Health Checks**
```bash
# Check service health
curl https://api.your-domain.com/health
curl https://analytics.your-domain.com/health
curl https://ai.your-domain.com/health

# Check infrastructure
kubectl get pods -n go-coffee-platform
kubectl get services -n go-coffee-platform
kubectl get ingress -n go-coffee-platform
```

### **Performance Validation**
```bash
# API performance
curl -w "@curl-format.txt" https://api.your-domain.com/orders

# Database performance
kubectl exec -it postgres-0 -- psql -c "SELECT * FROM pg_stat_activity;"

# Kafka performance
kubectl exec -it kafka-0 -- kafka-topics.sh --describe --bootstrap-server localhost:9092
```

## 🔄 Maintenance & Updates

### **Rolling Updates**
```bash
# Update specific service
kubectl set image deployment/producer-service producer=ghcr.io/dimajoyti/go-coffee/producer:v1.1.0

# Update with Helm
helm upgrade go-coffee-platform ./helm/go-coffee-platform \
  --set global.imageTag=v1.1.0

# Rollback if needed
helm rollback go-coffee-platform 1
```

### **Backup & Recovery**
```bash
# Database backup
kubectl exec postgres-0 -- pg_dump go_coffee > backup-$(date +%Y%m%d).sql

# Kafka backup
./scripts/backup-kafka-topics.sh

# Full platform backup
./scripts/backup-platform.sh
```

### **Monitoring & Alerting**
```bash
# Check metrics
curl https://prometheus.your-domain.com/api/v1/query?query=up

# View dashboards
open https://grafana.your-domain.com

# Check traces
open https://jaeger.your-domain.com
```

## 🚨 Troubleshooting

### **Common Issues**

#### **Services Not Starting**
```bash
# Check pod logs
kubectl logs -f deployment/producer-service

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp

# Check resource usage
kubectl top pods
```

#### **Database Connection Issues**
```bash
# Test database connectivity
kubectl exec -it postgres-0 -- psql -U postgres -c "SELECT 1;"

# Check database logs
kubectl logs postgres-0
```

#### **Kafka Issues**
```bash
# Check Kafka topics
kubectl exec kafka-0 -- kafka-topics.sh --list --bootstrap-server localhost:9092

# Check consumer lag
kubectl exec kafka-0 -- kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --all-groups
```

### **Performance Issues**
```bash
# Check resource usage
kubectl top nodes
kubectl top pods

# Check application metrics
curl https://api.your-domain.com/metrics

# Analyze traces
# Visit Jaeger UI and analyze slow requests
```

## 📞 Support

### **Documentation**
- [Architecture Overview](ARCHITECTURE.md)
- [API Documentation](API_DOCUMENTATION.md)
- [Security Guide](SECURITY.md)
- [Monitoring Guide](MONITORING.md)

### **Community**
- GitHub Issues: https://github.com/DimaJoyti/go-coffee/issues
- Discussions: https://github.com/DimaJoyti/go-coffee/discussions
- Wiki: https://github.com/DimaJoyti/go-coffee/wiki

### **Enterprise Support**
For enterprise support, please contact: enterprise@go-coffee.com

---

**🎉 Congratulations! You now have a complete deployment guide for the Go Coffee enterprise platform. Follow these steps to deploy your coffee business platform at any scale!**
