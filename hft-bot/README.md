# High-Frequency Trading (HFT) Bot System

A comprehensive, production-ready High-Frequency Trading bot system built in Go with Clean Architecture principles.

## 🚀 Features

### Core Capabilities
- **Ultra-Low Latency**: Sub-millisecond order execution
- **High Throughput**: 10,000+ orders per second
- **Multi-Exchange Support**: Binance, Coinbase Pro, Kraken
- **Real-time Market Data**: WebSocket-based data feeds
- **Advanced Risk Management**: Position limits, stop-losses, circuit breakers
- **Multiple Trading Strategies**: Arbitrage, Market Making, Momentum, Mean Reversion
- **Real-time Portfolio Management**: P&L tracking and performance analytics

### Technical Features
- **Microservices Architecture**: Clean separation of concerns
- **OpenTelemetry Integration**: Comprehensive observability
- **Lock-free Data Structures**: Optimized for performance
- **Paper Trading Mode**: Safe strategy testing
- **99.9% Uptime**: Robust error handling and failover
- **Comprehensive Testing**: Unit, integration, and load tests

## 🏗️ Architecture

### Services
- **Market Data Service**: Real-time data ingestion and order book management
- **Order Execution Service**: Ultra-low latency order processing
- **Risk Management Service**: Real-time risk monitoring and controls
- **Strategy Engine Service**: Trading strategy execution and management
- **Portfolio Service**: Position tracking and P&L calculation
- **Configuration Service**: Dynamic configuration management
- **API Gateway**: External API access and rate limiting

### Technology Stack
- **Language**: Go 1.23+
- **Databases**: Redis (caching), PostgreSQL (persistence)
- **Messaging**: gRPC for inter-service communication
- **Observability**: OpenTelemetry, Prometheus, Grafana, Jaeger
- **Deployment**: Docker, Kubernetes, Helm
- **Testing**: Go testing framework with comprehensive test suites

## 📁 Project Structure

```
hft-bot/
├── cmd/                    # Service entry points
│   ├── market-data/       # Market data service
│   ├── execution/         # Order execution service
│   ├── risk/              # Risk management service
│   ├── strategy/          # Strategy engine service
│   ├── portfolio/         # Portfolio service
│   └── gateway/           # API gateway
├── internal/               # Core business logic
│   ├── market-data/       # Market data domain
│   ├── execution/         # Order execution domain
│   ├── risk/              # Risk management domain
│   ├── strategy/          # Strategy domain
│   ├── portfolio/         # Portfolio domain
│   └── shared/            # Shared utilities
├── pkg/                   # Shared packages
│   ├── logger/            # OpenTelemetry logger
│   ├── config/            # Configuration management
│   ├── database/          # Database connections
│   ├── metrics/           # Performance metrics
│   └── websocket/         # WebSocket utilities
├── api/                   # API definitions
│   ├── proto/             # gRPC definitions
│   └── openapi/           # REST API specs
├── configs/               # Configuration files
├── deployments/           # Docker and K8s manifests
├── scripts/               # Build and deployment scripts
└── tests/                 # Test suites
```

## 🚦 Getting Started

### Prerequisites
- Go 1.23+
- Docker and Docker Compose
- Redis
- PostgreSQL
- Make

### Quick Start
```bash
# Clone the repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee/hft-bot

# Start dependencies
make deps-up

# Build all services
make build

# Run in paper trading mode
make run-paper

# View real-time dashboard
open http://localhost:8080
```

### Configuration
Copy and modify the configuration file:
```bash
cp configs/config.example.yaml configs/config.yaml
# Edit configs/config.yaml with your exchange API keys and settings
```

## 📊 Performance Targets

- **Latency**: Sub-millisecond order execution
- **Throughput**: 10,000+ orders per second
- **Uptime**: 99.9% availability
- **Recovery**: < 5 seconds failover time
- **Concurrency**: 100+ simultaneous connections

## 🛡️ Safety Features

- **Paper Trading Mode**: Test strategies without real money
- **Risk Management**: Comprehensive position and loss limits
- **Circuit Breakers**: Automatic trading halts on anomalies
- **Emergency Stops**: Manual and automatic kill switches
- **Audit Logging**: Complete trade and decision audit trail
- **Gradual Rollout**: Feature flags for safe deployments

## 📈 Trading Strategies

### Built-in Strategies
1. **Market Making**: Provides liquidity and captures spreads
2. **Arbitrage**: Cross-exchange and statistical arbitrage
3. **Momentum**: Trend-following based on price movements
4. **Mean Reversion**: Contrarian strategy based on price reversals

### Custom Strategies
The system supports custom strategy development through a pluggable framework.

## 🔧 Development

### Building
```bash
make build          # Build all services
make build-service SERVICE=market-data  # Build specific service
```

### Testing
```bash
make test           # Run all tests
make test-unit      # Run unit tests only
make test-integration  # Run integration tests
make test-load      # Run load tests
```

### Deployment
```bash
make deploy-dev     # Deploy to development
make deploy-staging # Deploy to staging
make deploy-prod    # Deploy to production
```

## 📚 Documentation

- [API Documentation](docs/api.md)
- [Architecture Guide](docs/architecture.md)
- [Deployment Guide](docs/deployment.md)
- [Strategy Development](docs/strategies.md)
- [Performance Tuning](docs/performance.md)
- [Troubleshooting](docs/troubleshooting.md)

## 🔐 Security

- API key management and rotation
- Rate limiting and DDoS protection
- Input validation and sanitization
- Audit logging for compliance
- Network security and encryption
- Access control and authentication

## 📞 Support

For questions, issues, or contributions, please refer to our [Contributing Guide](CONTRIBUTING.md).

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
