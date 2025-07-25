# 🚀 Epic Crypto Terminal - Deployment Guide

This guide covers deploying the Epic Crypto Terminal stack in various environments.

## 📋 Prerequisites

- Docker & Docker Compose
- Node.js 18+ (for local development)
- Go 1.21+ (for local development)
- PostgreSQL 15+ (if not using Docker)
- Redis 7+ (if not using Docker)

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Next.js       │    │   Go Backend    │    │   PostgreSQL    │
│   Dashboard     │◄──►│   API Server    │◄──►│   Database      │
│   (Port 3001)   │    │   (Port 8090)   │    │   (Port 5432)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │     Redis       │
                    │     Cache       │
                    │   (Port 6379)   │
                    └─────────────────┘
```

## 🐳 Docker Deployment (Recommended)

### Quick Start

1. **Clone and navigate to the project**
   ```bash
   git clone <repository-url>
   cd crypto-terminal
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start the complete stack**
   ```bash
   docker-compose -f docker-compose.epic.yml up -d
   ```

4. **Access the applications**
   - Dashboard: http://localhost:3001
   - API: http://localhost:8090
   - API Docs: http://localhost:8090/swagger/index.html

### Environment Configuration

Create a `.env` file in the project root:

```env
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=crypto_user
DB_PASSWORD=crypto_password
DB_NAME=crypto_terminal

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=redis_password

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# API Configuration
API_RATE_LIMIT=1000
CORS_ORIGINS=http://localhost:3001,https://your-domain.com

# TradingView API (Optional)
TRADINGVIEW_API_KEY=your_tradingview_api_key

# External APIs
COINMARKETCAP_API_KEY=your_coinmarketcap_api_key
COINGECKO_API_KEY=your_coingecko_api_key
```

### Production Deployment

1. **Update environment variables for production**
   ```bash
   # Update .env with production values
   CORS_ORIGINS=https://your-domain.com
   JWT_SECRET=your-production-jwt-secret
   ```

2. **Deploy with production profile**
   ```bash
   docker-compose -f docker-compose.epic.yml --profile production up -d
   ```

3. **Enable monitoring (optional)**
   ```bash
   docker-compose -f docker-compose.epic.yml --profile monitoring up -d
   ```

## 🖥️ Local Development

### Backend (Go)

1. **Install dependencies**
   ```bash
   cd crypto-terminal
   go mod download
   ```

2. **Set up database**
   ```bash
   # Start PostgreSQL and Redis with Docker
   docker-compose up postgres redis -d
   
   # Run migrations
   go run cmd/migrate/main.go
   ```

3. **Start the API server**
   ```bash
   go run cmd/api/main.go
   ```

### Frontend (Next.js)

1. **Install dependencies**
   ```bash
   cd crypto-terminal/dashboard
   npm install
   ```

2. **Set up environment**
   ```bash
   cp .env.example .env.local
   # Edit .env.local with your configuration
   ```

3. **Start development server**
   ```bash
   npm run dev
   ```

## ☁️ Cloud Deployment

### AWS ECS/Fargate

1. **Build and push images**
   ```bash
   # Build images
   docker build -t epic-crypto-api .
   docker build -t epic-crypto-dashboard ./dashboard
   
   # Tag for ECR
   docker tag epic-crypto-api:latest 123456789.dkr.ecr.region.amazonaws.com/epic-crypto-api:latest
   docker tag epic-crypto-dashboard:latest 123456789.dkr.ecr.region.amazonaws.com/epic-crypto-dashboard:latest
   
   # Push to ECR
   docker push 123456789.dkr.ecr.region.amazonaws.com/epic-crypto-api:latest
   docker push 123456789.dkr.ecr.region.amazonaws.com/epic-crypto-dashboard:latest
   ```

2. **Deploy using ECS task definitions**
   - Use the provided `aws/ecs-task-definition.json`
   - Configure RDS PostgreSQL and ElastiCache Redis
   - Set up Application Load Balancer

### Google Cloud Run

1. **Deploy API**
   ```bash
   gcloud run deploy epic-crypto-api \
     --image gcr.io/PROJECT_ID/epic-crypto-api \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated
   ```

2. **Deploy Dashboard**
   ```bash
   gcloud run deploy epic-crypto-dashboard \
     --image gcr.io/PROJECT_ID/epic-crypto-dashboard \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated
   ```

### Kubernetes

1. **Apply Kubernetes manifests**
   ```bash
   kubectl apply -f k8s/namespace.yaml
   kubectl apply -f k8s/configmap.yaml
   kubectl apply -f k8s/secrets.yaml
   kubectl apply -f k8s/postgres.yaml
   kubectl apply -f k8s/redis.yaml
   kubectl apply -f k8s/api.yaml
   kubectl apply -f k8s/dashboard.yaml
   kubectl apply -f k8s/ingress.yaml
   ```

## 🔧 Configuration

### Nginx Reverse Proxy

Create `nginx/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream api {
        server crypto-api:8090;
    }
    
    upstream dashboard {
        server crypto-dashboard:3001;
    }
    
    server {
        listen 80;
        server_name your-domain.com;
        
        location /api/ {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
        
        location / {
            proxy_pass http://dashboard;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
```

### SSL/TLS Configuration

1. **Generate SSL certificates**
   ```bash
   # Using Let's Encrypt
   certbot certonly --webroot -w /var/www/html -d your-domain.com
   ```

2. **Update Nginx configuration**
   ```nginx
   server {
       listen 443 ssl;
       ssl_certificate /etc/nginx/ssl/cert.pem;
       ssl_certificate_key /etc/nginx/ssl/key.pem;
       # ... rest of configuration
   }
   ```

## 📊 Monitoring & Observability

### Prometheus Metrics

The API exposes metrics at `/metrics`:
- HTTP request duration
- Database connection pool stats
- WebSocket connection count
- Custom business metrics

### Grafana Dashboards

Pre-configured dashboards for:
- API performance metrics
- Database performance
- Trading activity
- System resources

### Health Checks

- API: `GET /health`
- Dashboard: `GET /api/health`
- Database: Connection pool status
- Redis: Ping command

## 🔒 Security Considerations

### Production Checklist

- [ ] Change default passwords
- [ ] Use strong JWT secrets
- [ ] Enable HTTPS/TLS
- [ ] Configure CORS properly
- [ ] Set up rate limiting
- [ ] Enable database encryption
- [ ] Use secrets management
- [ ] Configure firewall rules
- [ ] Enable audit logging

### Environment Variables Security

```bash
# Use Docker secrets or external secret management
docker secret create jwt_secret jwt_secret.txt
docker secret create db_password db_password.txt
```

## 🚨 Troubleshooting

### Common Issues

1. **Database connection failed**
   ```bash
   # Check database status
   docker-compose logs postgres
   
   # Test connection
   docker exec -it epic-crypto-postgres psql -U crypto_user -d crypto_terminal
   ```

2. **API not responding**
   ```bash
   # Check API logs
   docker-compose logs crypto-api
   
   # Check health endpoint
   curl http://localhost:8090/health
   ```

3. **Dashboard build failed**
   ```bash
   # Check build logs
   docker-compose logs crypto-dashboard
   
   # Rebuild with verbose output
   docker-compose build --no-cache crypto-dashboard
   ```

### Performance Tuning

1. **Database optimization**
   - Increase shared_buffers
   - Tune work_mem
   - Configure connection pooling

2. **Redis optimization**
   - Set appropriate maxmemory
   - Configure eviction policy
   - Enable persistence if needed

3. **API optimization**
   - Adjust GOMAXPROCS
   - Tune database connection pool
   - Configure rate limiting

## 📈 Scaling

### Horizontal Scaling

1. **API scaling**
   ```bash
   docker-compose up --scale crypto-api=3
   ```

2. **Load balancer configuration**
   - Use Nginx upstream
   - Configure health checks
   - Set up session affinity for WebSockets

### Database Scaling

1. **Read replicas**
   - Configure PostgreSQL streaming replication
   - Route read queries to replicas

2. **Connection pooling**
   - Use PgBouncer
   - Configure pool sizes appropriately

## 📞 Support

For deployment issues:
- 📧 Email: devops@epic-crypto-terminal.com
- 💬 Discord: [Join our community](https://discord.gg/epic-crypto)
- 📖 Documentation: [Full docs](https://docs.epic-crypto-terminal.com)

---

**Happy deploying! 🚀**
