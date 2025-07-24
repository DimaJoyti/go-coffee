# 📈 Crypto Market Terminal

A comprehensive cryptocurrency market terminal integrated with the Go Coffee Web3 ecosystem, providing real-time market data, technical analysis, portfolio tracking, and DeFi integration.

## 🌟 Features

### 📊 **Market Data & Analysis**
- **Real-time Price Feeds** - Live cryptocurrency prices from multiple exchanges
- **Technical Indicators** - RSI, MACD, Bollinger Bands, Moving Averages
- **Chart Analysis** - TradingView integration with multiple timeframes
- **Market Overview** - Top gainers/losers, market cap rankings, volume analysis

### 💼 **Portfolio Management**
- **Real-time Portfolio Tracking** - Live P&L calculations
- **Asset Allocation** - Diversification analysis and pie charts
- **Transaction History** - Complete trading history with performance metrics
- **Multi-wallet Support** - Integration with existing Go Coffee wallet services

### 🚨 **Alert System**
- **Price Alerts** - Custom price level notifications
- **Technical Signals** - AI-generated trading signals
- **DeFi Opportunities** - Yield farming and arbitrage alerts
- **News & Events** - Market-moving news and calendar events

### ⚡ **Advanced Order Flow Toolkit**
- **Configurable Footprint Charts** - Visualize buy/sell volume at each price level
- **Volume Profile Tools** - VPSV & VPVR analysis with Value Area calculation
- **True Tick-Level Data** - Raw, unfiltered market buy/sell volumes
- **Delta Analysis** - Real-time buying/selling pressure measurement
- **Imbalance Detection** - Identify order flow imbalances and absorption patterns
- **Point of Control** - Automatic identification of high-volume price levels

### 🌐 **DeFi Integration**
- **Liquidity Pool Tracking** - Real-time pool information and APY
- **Yield Farming** - Automated yield optimization strategies
- **Arbitrage Scanner** - Cross-DEX arbitrage opportunities
- **Gas Tracker** - Network fee optimization

### ⚡ **High-Frequency Trading (HFT)**
- **Ultra-Low Latency** - Sub-millisecond market data processing
- **Multi-Exchange Connectivity** - Binance, Coinbase, Kraken integration
- **Advanced Order Management** - Complete order lifecycle with smart routing
- **Strategy Engine** - Market making, arbitrage, and momentum strategies
- **Real-Time Risk Management** - Dynamic limits and circuit breakers
- **Performance Analytics** - Comprehensive trading metrics and reporting

## 🏗️ Architecture

```
crypto-terminal/
├── cmd/terminal/           # Main application entry point
├── internal/
│   ├── terminal/          # Core terminal business logic
│   ├── market/           # Market data providers and aggregation
│   ├── analysis/         # Technical analysis and indicators
│   ├── portfolio/        # Portfolio tracking and management
│   ├── alerts/           # Alert system and notifications
│   └── websocket/        # Real-time WebSocket communication
├── web/                  # Frontend React application
│   ├── src/
│   ├── components/
│   └── pages/
├── configs/              # Configuration files
├── docs/                 # Documentation
└── tests/               # Test files
```

## 🆕 Recent Updates

### Enhanced Features
- **OpenTelemetry Integration**: Comprehensive monitoring, tracing, and metrics
- **Enhanced Configuration**: Environment-specific settings with validation
- **Improved Health Monitoring**: Detailed component health checks
- **Real API Integration**: Replaced mock data with live API calls
- **Database Migrations**: Complete database schema with migrations
- **Go Coffee Integration**: Kafka-based ecosystem integration
- **Enhanced WebSocket Hub**: Improved real-time communication with subscriptions
- **Comprehensive Testing**: Unit, integration, and benchmark tests

### Performance Improvements
- **Caching Strategy**: Redis-based caching for market data and user sessions
- **Connection Pooling**: Optimized database and Redis connection management
- **Error Handling**: Robust error handling with fallback mechanisms
- **Security Enhancements**: CORS middleware, rate limiting, and input validation

## 🚀 Quick Start

### Prerequisites

- **Go 1.22+** - Latest Go version
- **Node.js 18+** - For frontend development
- **Redis 8+** - For caching and real-time data
- **PostgreSQL 15+** - For historical data storage

### Installation

1. **Clone and navigate to crypto-terminal**
   ```bash
   cd crypto-terminal
   ```

2. **Install Go dependencies**
   ```bash
   go mod tidy
   ```

3. **Install frontend dependencies**
   ```bash
   cd web && npm install
   ```

4. **Start the services**
   ```bash
   # Terminal 1: Start the backend
   go run cmd/terminal/main.go

   # Terminal 2: Start the frontend
   cd web && npm start
   ```

### Configuration

Create a `configs/config.yaml` file:

```yaml
server:
  port: 8090
  host: "0.0.0.0"

market_data:
  providers:
    coingecko:
      api_key: "${COINGECKO_API_KEY}"
      rate_limit: 50
    binance:
      websocket_url: "wss://stream.binance.com:9443/ws"
    
redis:
  host: "localhost"
  port: 6379
  db: 2

database:
  host: "localhost"
  port: 5432
  name: "crypto_terminal"
  user: "postgres"
  password: "password"

integrations:
  go_coffee:
    defi_service_url: "http://localhost:8082"
    wallet_service_url: "http://localhost:8083"
    kafka_brokers: ["localhost:9092"]
```

## 📡 API Endpoints

### Market Data
- `GET /api/v1/market/prices` - Get current cryptocurrency prices
- `GET /api/v1/market/history/{symbol}` - Get historical price data
- `GET /api/v1/market/indicators/{symbol}` - Get technical indicators

### Portfolio
- `GET /api/v1/portfolio` - Get portfolio overview
- `GET /api/v1/portfolio/performance` - Get portfolio performance metrics
- `POST /api/v1/portfolio/sync` - Sync with wallet services

### Alerts
- `GET /api/v1/alerts` - Get active alerts
- `POST /api/v1/alerts` - Create new alert
- `DELETE /api/v1/alerts/{id}` - Delete alert

### Order Flow
- `GET /api/v1/orderflow/footprint/{symbol}` - Get footprint chart data
- `GET /api/v1/orderflow/volume-profile/{symbol}` - Get volume profile analysis
- `GET /api/v1/orderflow/delta/{symbol}` - Get delta analysis
- `GET /api/v1/orderflow/metrics/{symbol}` - Get real-time order flow metrics
- `GET /api/v1/orderflow/imbalances/{symbol}` - Get active imbalances

### High-Frequency Trading (HFT)
- `GET /api/v1/hft/status` - Get HFT system status and metrics
- `GET /api/v1/hft/latency` - Get latency statistics
- `GET /api/v1/hft/strategies` - List all trading strategies
- `POST /api/v1/hft/strategies/{id}/start` - Start a trading strategy
- `POST /api/v1/hft/strategies/{id}/stop` - Stop a trading strategy
- `GET /api/v1/hft/orders` - Get active HFT orders
- `POST /api/v1/hft/orders` - Place a new HFT order
- `DELETE /api/v1/hft/orders/{id}` - Cancel an HFT order
- `GET /api/v1/hft/positions` - Get HFT positions
- `GET /api/v1/hft/risk/events` - Get risk management events

### WebSocket
- `ws://localhost:8090/ws/market` - Real-time market data
- `ws://localhost:8090/ws/portfolio` - Real-time portfolio updates
- `ws://localhost:8090/ws/alerts` - Real-time alert notifications
- `ws://localhost:8090/ws/orderflow` - Real-time order flow data
- `ws://localhost:8090/ws/hft` - Real-time HFT events and updates

## 🔧 Integration with Go Coffee

The crypto terminal seamlessly integrates with the existing Go Coffee ecosystem:

- **DeFi Service Integration** - Leverages existing DeFi protocols for yield farming and arbitrage
- **Wallet Service Connection** - Uses Go Coffee wallet infrastructure for portfolio tracking
- **AI Agent Coordination** - Receives trading signals from AI agents via Kafka
- **Shared Infrastructure** - Uses existing Redis, PostgreSQL, and monitoring setup

## 📊 Technical Indicators

Supported technical analysis indicators:

- **Trend Indicators**: SMA, EMA, MACD, Parabolic SAR
- **Momentum Indicators**: RSI, Stochastic, Williams %R
- **Volatility Indicators**: Bollinger Bands, ATR
- **Volume Indicators**: OBV, Volume SMA

## 🤖 AI Integration

The terminal integrates with Go Coffee's AI agent network:

- **Trading Signal Agent** - Generates buy/sell signals based on technical analysis
- **Market Sentiment Agent** - Analyzes social media and news sentiment
- **Risk Management Agent** - Monitors portfolio risk and suggests adjustments
- **Arbitrage Agent** - Identifies cross-exchange arbitrage opportunities

## 📈 Performance Metrics

Target performance benchmarks:

| Metric | Target | Current |
|--------|--------|---------|
| WebSocket Latency | < 50ms | TBD |
| API Response Time | < 100ms | TBD |
| Data Update Frequency | 1s | TBD |
| Concurrent Users | 1000+ | TBD |

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...

# Run frontend tests
cd web && npm test
```

## 📚 Documentation

- [Quick Start Guide](docs/QUICK_START.md)
- [Order Flow Toolkit](docs/ORDER_FLOW_TOOLKIT.md)
- [High-Frequency Trading System](docs/HFT_SYSTEM.md)
- [Architecture Guide](docs/architecture.md)
- [API Reference](docs/api-reference.md)
- [Integration Guide](docs/integration.md)
- [Development Guide](docs/development.md)

## 🤝 Contributing

Please see the main [Go Coffee Contributing Guide](../CONTRIBUTING.md) for contribution guidelines.

## 📄 License

This project is part of the Go Coffee ecosystem and is licensed under the MIT License.
