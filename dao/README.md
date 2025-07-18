# DAO Platform - Go Coffee Ecosystem

## 🎯 Overview

The DAO Platform is a comprehensive solution for building market-validated DeFi solutions within the Go Coffee ecosystem. It focuses on TVL growth, MAU expansion, and provides a complete framework for  incentives, governance, and solution delivery.

## 🏗️ Architecture

### Core Objectives
1. **TVL Growth Solutions**: Develop mechanisms to increase Total Value Locked across L1/L2 networks
2. **MAU Expansion**: Create user acquisition and retention strategies for DeFi protocols
3. **Market Validation**: Implement data-driven validation processes for all solutions

### Key Components
- **DAO Governance Structure**: Smart contract-based voting with Coffee Token integration
- **Incentive Programs**: Bounty system with performance-based rewards
- **Solution Delivery Framework**: End-to-end development and deployment pipeline
- **Marketplace Integration**: 3rd party component marketplace with revenue sharing

## 📁 Directory Structure

```
 dao/
├── contracts/                 # Smart contracts
│   ├── DAOGovernor.sol       # Governance proposals and voting
│   ├── BountyManager.sol     # Developer incentive management
│   ├── RevenueSharing.sol    # Performance-based rewards
│   └── SolutionRegistry.sol  # Component marketplace registry
├── cmd/                      # Service entry points
│   ├── dao-governance-service/
│   ├── bounty-service/
│   ├── solution-marketplace/
│   ├── developer-portal/
│   └── metrics-aggregator/
├── internal/                 # Core business logic
│   ├── dao/                  # DAO governance logic
│   ├── bounty/              # Bounty management
│   ├── marketplace/         # Solution marketplace
│   ├── metrics/             # Performance tracking
│   └── revenue/             # Revenue sharing
├── api/                     # API definitions
│   └── proto/               # gRPC definitions
├── web/                     # Frontend components
│   ├── dao-portal/          # Developer portal UI
│   └── governance-ui/       # Governance interface
├── migrations/              # Database migrations
├── configs/                 # Configuration files
├── scripts/                 # Deployment and utility scripts
└── docs/                    # Documentation
```

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- Node.js 18+ (for smart contracts)
- PostgreSQL 13+
- Redis 6+
- Docker & Docker Compose

### Installation
```bash
# Navigate to developer-dao directory
cd developer-dao

# Install Go dependencies
go mod tidy

# Install smart contract dependencies
cd contracts && npm install

# Set up configuration
cp configs/config.yaml.example configs/config.yaml

# Run database migrations
make migrate-up

# Start all services
make run-all
```

## 📊 Success Metrics

### TVL Growth Targets
- 50% increase in TVL across integrated protocols (6 months)
- Solution-specific TVL attribution and tracking
- Developer performance scoring based on TVL impact

### MAU Expansion Goals
- 100% growth in monthly active users (6 months)
- User retention rate improvement
- Community engagement depth analytics

### Developer Ecosystem KPIs
- 50+ active developers in DAO (6 months)
- 20+ validated solutions deployed
- 30% cost reduction through component reuse

## 🔧 Integration with Go Coffee Ecosystem

This platform integrates seamlessly with existing Go Coffee infrastructure:
- **DeFi Service**: Leverages existing Uniswap, Aave, 1inch integrations
- **Coffee Token**: Extends current staking functionality for governance
- **AI Agents**: Integrates with orchestration engine for automated operations
- **Monitoring**: Uses existing Prometheus/Grafana infrastructure

## 🚀 Quick Deployment

### Option 1: Complete Platform (Recommended)

```bash
# Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee/dao

# Setup environment
cp .env.example .env
# Edit .env with your API keys:
# - OPENAI_API_KEY: For AI features
# - GITHUB_TOKEN: For repository analysis
# - WALLET_CONNECT_PROJECT_ID: For Web3 integration

# Start complete platform (backend + frontend)
./start-platform.sh start
```

### Option 2: Backend Only

```bash
# Start only backend services
./run-local.sh start
```

### Option 3: Frontend Only

```bash
# Setup and start frontend applications
cd web
./build-frontend.sh dev
```

### Option 4: Docker Deployment

```bash
# Build Docker images first
docker build -f cmd/bounty-service/Dockerfile -t developer-dao/bounty-service:latest .
docker build -f cmd/marketplace-service/Dockerfile -t developer-dao/marketplace-service:latest .
docker build -f cmd/metrics-service/Dockerfile -t developer-dao/metrics-service:latest .
docker build -f cmd/dao-governance-service/Dockerfile -t developer-dao/dao-governance-service:latest .

# Start infrastructure services
docker-compose up -d postgres redis qdrant

# Start backend services
docker-compose up -d bounty-service marketplace-service metrics-service dao-governance-service
```

### 🔍 Verify Deployment

```bash
# Check platform status
./start-platform.sh status

# Or check individual services
curl http://localhost:8080/health  # Bounty Service
curl http://localhost:8081/health  # Marketplace Service
curl http://localhost:8082/health  # Metrics Service
curl http://localhost:8084/health  # DAO Governance Service
curl http://localhost:3000         # DAO Portal
curl http://localhost:3001         # Governance UI
```

### 🌐 Access URLs

- **DAO Portal**: http://localhost:3000
- **Governance UI**: http://localhost:3001
- **Bounty Service**: http://localhost:8080
- **Marketplace Service**: http://localhost:8081
- **Metrics Service**: http://localhost:8082
- **DAO Governance**: http://localhost:8084
- **API Documentation**: http://localhost:8080/swagger/index.html
- **Monitoring**: http://localhost:9090 (Prometheus), http://localhost:3003 (Grafana)

## 📖 Documentation

- [Smart Contract Documentation](docs/smart-contracts.md)
- [API Reference](docs/api-reference.md)
- [Developer Guide](docs/developer-guide.md)
- [Deployment Guide](docs/deployment.md)

## 🤝 Contributing

Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to participate in the Developer DAO.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
