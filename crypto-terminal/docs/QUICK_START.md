# 🚀 Crypto Market Terminal - Quick Start Guide

Welcome to the Crypto Market Terminal! This guide will help you get up and running quickly.

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.22+** - [Download Go](https://golang.org/dl/)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Node.js 18+** - [Download Node.js](https://nodejs.org/) (for frontend development)
- **Git** - [Install Git](https://git-scm.com/downloads)

## 🏃‍♂️ Quick Start (Docker)

The fastest way to get started is using Docker Compose:

### 1. Clone the Repository

```bash
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee/crypto-terminal
```

### 2. Start with Docker Compose

```bash
# Start all services (database, cache, backend)
docker-compose up -d

# Check if services are running
docker-compose ps
```

### 3. Access the Application

- **Backend API**: http://localhost:8090
- **Health Check**: http://localhost:8090/health
- **API Documentation**: http://localhost:8090/api/v1

### 4. Test the API

```bash
# Get market overview
curl http://localhost:8090/api/v1/market/overview

# Get Bitcoin price
curl http://localhost:8090/api/v1/market/prices/bitcoin

# Get portfolio data
curl http://localhost:8090/api/v1/portfolio
```

## 🛠️ Development Setup

For development with hot reloading and debugging:

### 1. Start Infrastructure Services

```bash
# Start only database and cache
docker-compose up -d postgres redis
```

### 2. Install Go Dependencies

```bash
go mod tidy
```

### 3. Run the Backend

```bash
# Using Make (recommended)
make run-dev

# Or directly with Go
go run cmd/terminal/main.go
```

### 4. Setup Frontend (Optional)

```bash
cd web
npm install
npm start
```

The frontend will be available at http://localhost:3000

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the project root:

```bash
# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=crypto_terminal
DATABASE_USER=postgres
DATABASE_PASSWORD=password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=2

# API Keys (optional)
COINGECKO_API_KEY=your_coingecko_api_key

# Logging
LOG_LEVEL=info
```

### Configuration File

The main configuration is in `configs/config.yaml`. Key sections:

```yaml
server:
  port: 8090
  host: "0.0.0.0"

market_data:
  providers:
    coingecko:
      api_key: "${COINGECKO_API_KEY}"
      rate_limit: 50

redis:
  host: "localhost"
  port: 6379
  db: 2

database:
  host: "localhost"
  port: 5432
  name: "crypto_terminal"
```

## 📡 API Endpoints

### Market Data

```bash
# Get all cryptocurrency prices
GET /api/v1/market/prices

# Get specific cryptocurrency price
GET /api/v1/market/prices/{symbol}

# Get price history
GET /api/v1/market/history/{symbol}?timeframe=1h&limit=100

# Get technical indicators
GET /api/v1/market/indicators/{symbol}?timeframe=1h

# Get market overview
GET /api/v1/market/overview

# Get top gainers/losers
GET /api/v1/market/gainers
GET /api/v1/market/losers
```

### Portfolio Management

```bash
# Get user portfolios
GET /api/v1/portfolio

# Get portfolio performance
GET /api/v1/portfolio/{id}/performance?timeRange=30D

# Get portfolio holdings
GET /api/v1/portfolio/{id}/holdings

# Sync portfolio with wallets
POST /api/v1/portfolio/{id}/sync
```

### Alerts

```bash
# Get user alerts
GET /api/v1/alerts

# Create new alert
POST /api/v1/alerts

# Update alert
PUT /api/v1/alerts/{id}

# Delete alert
DELETE /api/v1/alerts/{id}
```

### WebSocket Connections

```bash
# General WebSocket
ws://localhost:8090/ws

# Market data stream
ws://localhost:8090/ws/market

# Portfolio updates
ws://localhost:8090/ws/portfolio

# Alert notifications
ws://localhost:8090/ws/alerts
```

## 🧪 Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run tests with race detection
make test-race

# Run benchmarks
make bench
```

### Load Testing

```bash
# Install k6 (if not already installed)
brew install k6  # macOS
# or
sudo apt install k6  # Ubuntu

# Run load tests (when implemented)
k6 run tests/load/api-load-test.js
```

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check if PostgreSQL is running
   docker-compose ps postgres
   
   # Check logs
   docker-compose logs postgres
   
   # Restart database
   docker-compose restart postgres
   ```

2. **Redis Connection Failed**
   ```bash
   # Check if Redis is running
   docker-compose ps redis
   
   # Test Redis connection
   docker-compose exec redis redis-cli ping
   ```

3. **Port Already in Use**
   ```bash
   # Check what's using port 8090
   lsof -i :8090
   
   # Kill the process
   kill -9 <PID>
   ```

4. **Go Module Issues**
   ```bash
   # Clean module cache
   go clean -modcache
   
   # Re-download dependencies
   go mod download
   go mod tidy
   ```

### Debug Mode

Run the application in debug mode for more verbose logging:

```bash
LOG_LEVEL=debug make run-dev
```

### Health Checks

Check the health of all services:

```bash
# Backend health
curl http://localhost:8090/health

# Database health
docker-compose exec postgres pg_isready -U postgres

# Redis health
docker-compose exec redis redis-cli ping
```

## 📊 Monitoring

### Prometheus & Grafana (Optional)

Start the monitoring stack:

```bash
# Start monitoring services
make monitor-up

# Access Grafana
open http://localhost:3000
# Default credentials: admin/admin

# Access Prometheus
open http://localhost:9090
```

### Logs

View application logs:

```bash
# Docker logs
docker-compose logs -f crypto-terminal

# Development logs
make run-dev
```

## 🔄 Development Workflow

### Making Changes

1. **Backend Changes**:
   ```bash
   # Make your changes
   # Run tests
   make test
   
   # Run the application
   make run-dev
   ```

2. **Frontend Changes**:
   ```bash
   cd web
   # Make your changes
   npm start  # Hot reload enabled
   ```

3. **Database Changes**:
   ```bash
   # Update schema in scripts/init-db.sql
   # Reset database
   make db-reset
   ```

### Code Quality

```bash
# Format code
make fmt

# Lint code
make lint

# Vet code
make vet

# Security scan
make security-scan
```

## 🚀 Production Deployment

### Build for Production

```bash
# Build optimized binary
make prod-build

# Build Docker image
make docker-build

# Deploy (customize for your environment)
make prod-deploy
```

### Environment Setup

For production, ensure you have:

- Proper SSL certificates
- Environment variables configured
- Database backups enabled
- Monitoring and alerting setup
- Log aggregation configured

## 📚 Next Steps

1. **Explore the API** - Use the provided endpoints to build your own integrations
2. **Customize Configuration** - Modify `configs/config.yaml` for your needs
3. **Add Market Data Providers** - Implement additional data sources
4. **Extend Portfolio Features** - Add more portfolio analysis tools
5. **Build Custom Alerts** - Create sophisticated alert conditions
6. **Integrate with DeFi** - Connect to DeFi protocols for yield farming

## 🆘 Getting Help

- **Documentation**: Check the `docs/` directory for detailed guides
- **Issues**: Report bugs on GitHub Issues
- **Discussions**: Join GitHub Discussions for questions
- **API Reference**: See `docs/api-reference.md`

## 🎯 Key Features to Explore

- **Real-time Market Data** - Live cryptocurrency prices and charts
- **Portfolio Tracking** - Comprehensive portfolio management
- **Technical Analysis** - Built-in technical indicators
- **Smart Alerts** - Customizable price and technical alerts
- **DeFi Integration** - Yield farming and arbitrage opportunities
- **WebSocket Streams** - Real-time data updates
- **Risk Analysis** - Portfolio risk metrics and diversification

Happy trading! 🚀📈
