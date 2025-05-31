# 🚀 Go Coffee Production Deployment Guide

## 📋 Overview

Go Coffee - це enterprise-grade мікросервісна архітектура з AI інтеграцією, готова для production deployment. Система включає 7 мікросервісів з повним моніторингом, load balancing та observability.

## 🏗️ Architecture

### Microservices:
- **AI Search Engine** (Port 8092) - Семантичний пошук з AI
- **Auth Service** (Port 8080) - JWT автентифікація
- **User Gateway** (Port 8081) - API Gateway з load balancing
- **Kitchen Service** (gRPC 50052) - Управління кухнею
- **Communication Hub** (gRPC 50053) - Міжсервісна комунікація
- **Redis MCP Server** (Port 8093) - Redis MCP інтеграція

### Infrastructure:
- **Redis 8** - Blazingly fast caching та pub/sub
- **PostgreSQL 15** - Основна база даних
- **Nginx** - Load balancer та reverse proxy
- **Prometheus** - Metrics collection
- **Grafana** - Dashboards та візуалізація
- **Jaeger** - Distributed tracing

## 🚀 Quick Start

### 1. Docker Compose (Recommended for Development)

```bash
# Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Start production environment
./start_production.sh
```

### 2. Kubernetes (Production)

```bash
# Deploy to Kubernetes
./deploy_k8s.sh
```

## 📦 Prerequisites

### For Docker Deployment:
- Docker 20.10+
- Docker Compose 2.0+
- 4GB RAM minimum
- 10GB disk space

### For Kubernetes Deployment:
- Kubernetes 1.20+
- kubectl configured
- Docker for building images
- 8GB RAM minimum
- 20GB disk space

## 🔧 Configuration

### Environment Variables

Create `.env` file:

```bash
# AI Services
GEMINI_API_KEY=your-gemini-api-key-here
OLLAMA_BASE_URL=http://ollama:11434

# Security
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Database
POSTGRES_DB=go_coffee
POSTGRES_USER=go_coffee_user
POSTGRES_PASSWORD=go_coffee_password

# Redis
REDIS_URL=redis://redis:6379

# Logging
LOG_LEVEL=info

# Environment
ENVIRONMENT=production
```

## 🌐 Service Endpoints

### Public APIs (через Load Balancer):
```
http://localhost/api/v1/auth/register    - User registration
http://localhost/api/v1/auth/login       - User login
http://localhost/api/v1/orders           - Order management
http://localhost/api/v1/ai-search/       - AI search
http://localhost/api/v1/redis-mcp/       - Redis MCP
http://localhost/health                  - Health check
```

### Direct Service Access:
```
http://localhost:8080/health             - Auth Service
http://localhost:8081/health             - User Gateway
http://localhost:8092/api/v1/ai-search/health - AI Search
http://localhost:8093/health             - Redis MCP Server
```

### Monitoring:
```
http://localhost:9090                    - Prometheus
http://localhost:3000                    - Grafana (admin/admin)
http://localhost:16686                   - Jaeger Tracing
```

## 📊 Monitoring & Observability

### Prometheus Metrics:
- HTTP request duration
- Request count by status code
- Service health status
- Resource usage (CPU, Memory)
- Redis operations
- Database connections

### Grafana Dashboards:
- Service Overview
- API Performance
- Infrastructure Metrics
- Error Rates
- Business Metrics

### Jaeger Tracing:
- Request flow across services
- Performance bottlenecks
- Error tracking
- Dependency mapping

## 🔍 Health Checks

### Service Health:
```bash
# All services health
curl http://localhost/health

# Individual services
curl http://localhost:8080/health  # Auth
curl http://localhost:8081/health  # Gateway
curl http://localhost:8092/api/v1/ai-search/health  # AI Search
curl http://localhost:8093/health  # Redis MCP
```

### Infrastructure Health:
```bash
# Redis
redis-cli ping

# PostgreSQL
pg_isready -h localhost -p 5432 -U go_coffee_user
```

## 🛠️ Operations

### Docker Compose Commands:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f [service-name]

# Scale service
docker-compose up -d --scale user-gateway=3

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

### Kubernetes Commands:

```bash
# View all resources
kubectl get all -n go-coffee

# View logs
kubectl logs -f deployment/user-gateway -n go-coffee

# Scale deployment
kubectl scale deployment user-gateway --replicas=5 -n go-coffee

# Port forward for local access
kubectl port-forward -n go-coffee service/user-gateway-service 8081:8081

# Delete deployment
kubectl delete namespace go-coffee
```

## 🔐 Security

### Production Security Checklist:
- [ ] Change default JWT secret
- [ ] Use strong database passwords
- [ ] Configure SSL/TLS certificates
- [ ] Set up firewall rules
- [ ] Enable audit logging
- [ ] Configure rate limiting
- [ ] Set up backup strategy

### SSL/TLS Configuration:
```bash
# Generate self-signed certificate (for testing)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout deployments/nginx/ssl/nginx.key \
  -out deployments/nginx/ssl/nginx.crt
```

## 📈 Performance Tuning

### Recommended Resource Limits:

#### Docker Compose:
```yaml
resources:
  limits:
    memory: 512M
    cpus: '0.5'
  reservations:
    memory: 256M
    cpus: '0.25'
```

#### Kubernetes:
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Scaling Guidelines:
- **User Gateway**: 3-5 replicas (high traffic)
- **AI Search**: 2-3 replicas (CPU intensive)
- **Auth Service**: 2-3 replicas (security critical)
- **Kitchen Service**: 1-2 replicas (stateful)
- **Communication Hub**: 1-2 replicas (message broker)

## 🚨 Troubleshooting

### Common Issues:

#### Services not starting:
```bash
# Check logs
docker-compose logs [service-name]

# Check resource usage
docker stats

# Restart service
docker-compose restart [service-name]
```

#### Database connection issues:
```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres psql -U go_coffee_user -d go_coffee
```

#### Redis connection issues:
```bash
# Check Redis logs
docker-compose logs redis

# Test connection
docker-compose exec redis redis-cli ping
```

### Performance Issues:
1. Check Prometheus metrics
2. Review Grafana dashboards
3. Analyze Jaeger traces
4. Scale bottleneck services
5. Optimize database queries

## 📚 API Documentation

### Authentication:
```bash
# Register user
curl -X POST http://localhost/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","name":"Test User"}'

# Login
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Orders:
```bash
# Create order
curl -X POST http://localhost/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"items":[{"name":"Espresso","quantity":1,"price":3.50}]}'

# Get order
curl http://localhost/api/v1/orders/ORDER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### AI Search:
```bash
# Semantic search
curl -X POST http://localhost/api/v1/ai-search/semantic \
  -H "Content-Type: application/json" \
  -d '{"query":"strong espresso","limit":5}'
```

## 🎯 Next Steps

1. **Production Deployment**:
   - Set up CI/CD pipeline
   - Configure external load balancer
   - Set up backup strategy
   - Configure monitoring alerts

2. **Scaling**:
   - Implement horizontal pod autoscaling
   - Set up cluster autoscaling
   - Configure database read replicas
   - Implement caching strategies

3. **Security**:
   - Set up OAuth2/OIDC
   - Implement API rate limiting
   - Configure network policies
   - Set up vulnerability scanning

4. **Monitoring**:
   - Configure alerting rules
   - Set up log aggregation
   - Implement business metrics
   - Set up uptime monitoring

---

## 🏆 Success! 

Ваша Go Coffee мікросервісна архітектура готова до production використання! 

**Enterprise-grade features:**
✅ Microservices architecture  
✅ AI-powered search  
✅ Load balancing  
✅ Auto-scaling  
✅ Monitoring & observability  
✅ Security & authentication  
✅ High availability  
✅ Production-ready deployment  

**Готово обслуговувати тисячі користувачів одночасно!** ☕🚀
