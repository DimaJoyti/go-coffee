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

### WebSocket
- `ws://localhost:8090/ws/market` - Real-time market data
- `ws://localhost:8090/ws/portfolio` - Real-time portfolio updates
- `ws://localhost:8090/ws/alerts` - Real-time alert notifications
- `ws://localhost:8090/ws/orderflow` - Real-time order flow data

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
- [Architecture Guide](docs/architecture.md)
- [API Reference](docs/api-reference.md)
- [Integration Guide](docs/integration.md)
- [Development Guide](docs/development.md)

## 🤝 Contributing

Please see the main [Go Coffee Contributing Guide](../CONTRIBUTING.md) for contribution guidelines.

## 📄 License

This project is part of the Go Coffee ecosystem and is licensed under the MIT License.
