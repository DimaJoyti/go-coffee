# DeFi Module - Go Coffee Project

## Overview

The DeFi module provides comprehensive decentralized finance functionality for the Go Coffee project, including:

- **Token Price Feeds** - Real-time token pricing from multiple sources
- **DEX Aggregation** - Swap quotes and execution across multiple DEXs
- **Arbitrage Detection** - Automated arbitrage opportunity detection
- **Yield Farming** - Yield optimization and farming strategies
- **Trading Bots** - Automated trading strategies (DCA, Grid, Arbitrage)
- **On-Chain Analysis** - Whale watching and market signal detection
- **Multi-Chain Support** - Ethereum, BSC, Polygon, Solana

## Architecture

The DeFi module follows Clean Architecture principles with clear separation of concerns:

```
internal/defi/
├── models.go              # Domain models and types
├── service.go             # Main DeFi service
├── clients.go             # Protocol client implementations
├── aggregators.go         # Yield and arbitrage aggregators
├── trading_bot.go         # Trading bot implementation
└── service_integration_test.go  # Integration tests
```

### Key Components

1. **Service Layer** (`service.go`)
   - Main orchestrator for all DeFi operations
   - Manages blockchain clients and protocol integrations
   - Provides unified API for DeFi functionality

2. **Protocol Clients** (`clients.go`)
   - Uniswap, Aave, Chainlink, 1inch integrations
   - Solana DEX integrations (Raydium, Jupiter)
   - Mock implementations for testing

3. **Trading Components** (`aggregators.go`, `trading_bot.go`)
   - Arbitrage detection and execution
   - Yield farming optimization
   - Automated trading strategies

## Configuration

DeFi configuration is integrated into the main application config:

```go
type DeFiConfig struct {
    UniswapV3Router      string `json:"uniswap_v3_router"`
    AaveLendingPool      string `json:"aave_lending_pool"`
    CompoundComptroller  string `json:"compound_comptroller"`
    OneInchAPIKey        string `json:"oneinch_api_key"`
    ChainlinkEnabled     bool   `json:"chainlink_enabled"`
    ArbitrageEnabled     bool   `json:"arbitrage_enabled"`
    YieldFarmingEnabled  bool   `json:"yield_farming_enabled"`
    TradingBotsEnabled   bool   `json:"trading_bots_enabled"`
}
```

### Environment Variables

```bash
# DeFi Protocol Configuration
UNISWAP_V3_ROUTER=0xE592427A0AEce92De3Edee1F18E0157C05861564
AAVE_LENDING_POOL=0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9
COMPOUND_COMPTROLLER=0x3d9819210A31b4961b30EF54bE2aeD79B9c9Cd3B
ONEINCH_API_KEY=your-1inch-api-key

# Feature Flags
CHAINLINK_ENABLED=true
ARBITRAGE_ENABLED=true
YIELD_FARMING_ENABLED=true
TRADING_BOTS_ENABLED=false  # Disabled by default for safety

# Blockchain RPC URLs
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/your-project-id
SOLANA_RPC_URL=https://api.mainnet-beta.solana.com
```

## API Endpoints

The DeFi service exposes REST API endpoints:

### Token Operations
- `POST /api/v1/tokens/price` - Get token price
- `POST /api/v1/swaps/quote` - Get swap quote
- `POST /api/v1/swaps/execute` - Execute swap

### Liquidity & Yield
- `GET /api/v1/pools` - Get liquidity pools
- `GET /api/v1/yield/opportunities` - Get yield opportunities

### Arbitrage
- `GET /api/v1/arbitrage/opportunities` - Get arbitrage opportunities

### Trading Bots
- `POST /api/v1/bots` - Create trading bot
- `GET /api/v1/bots` - List all bots
- `GET /api/v1/bots/:id` - Get specific bot
- `POST /api/v1/bots/:id/start` - Start bot
- `POST /api/v1/bots/:id/stop` - Stop bot
- `DELETE /api/v1/bots/:id` - Delete bot
- `GET /api/v1/bots/:id/performance` - Get bot performance

### Analysis
- `GET /api/v1/analysis/signals` - Get market signals
- `GET /api/v1/analysis/whales` - Get whale activity
- `GET /api/v1/analysis/tokens/:address` - Get token analysis

## Usage Examples

### Getting Token Price

```bash
curl -X POST http://localhost:8093/api/v1/tokens/price \
  -H "Content-Type: application/json" \
  -d '{
    "token_address": "0xA0b86a33E6441e6e80D0c4C6C7527d5B8C6B0b0b",
    "chain": "ethereum"
  }'
```

### Creating a Trading Bot

```bash
curl -X POST http://localhost:8093/api/v1/bots \
  -H "Content-Type: application/json" \
  -d '{
    "name": "DCA Bot",
    "strategy": "dca",
    "config": {
      "max_position_size": "1000",
      "min_profit_margin": "0.01",
      "max_slippage": "0.005",
      "risk_tolerance": "medium",
      "auto_compound": true,
      "max_daily_trades": 10,
      "stop_loss_percent": "0.05",
      "take_profit_percent": "0.15",
      "execution_delay": "1s"
    }
  }'
```

## Running the Service

### Standalone Service

```bash
# Build and run the DeFi service
go build -o defi-service cmd/defi-service/main.go
./defi-service
```

### Integration with Main Application

The DeFi service can be integrated into the main Go Coffee application by importing and starting it as a component.

## Testing

Run the integration tests:

```bash
cd internal/defi
go test -v
```

The tests cover:
- Service initialization and configuration
- Token price retrieval
- Swap quote generation
- Trading bot lifecycle
- Arbitrage detection
- Yield farming optimization
- Concurrent operations
- Error handling

## Security Considerations

1. **Private Keys** - Never store private keys in code or config files
2. **API Keys** - Use environment variables for sensitive API keys
3. **Rate Limiting** - Implement rate limiting for external API calls
4. **Input Validation** - Validate all user inputs and transaction parameters
5. **Trading Limits** - Set reasonable limits for automated trading
6. **Monitoring** - Monitor all trading activities and set up alerts

## Future Enhancements

1. **Real Blockchain Integration** - Replace mock clients with real implementations
2. **Advanced Strategies** - Implement more sophisticated trading strategies
3. **Risk Management** - Enhanced risk assessment and management
4. **Portfolio Management** - Multi-asset portfolio optimization
5. **Flash Loans** - Flash loan arbitrage strategies
6. **Cross-Chain** - Cross-chain arbitrage and yield farming
7. **AI Integration** - Machine learning for market prediction

## Dependencies

- `shopspring/decimal` - Precise decimal arithmetic
- `ethereum/go-ethereum` - Ethereum client
- `gagliardetto/solana-go` - Solana client
- `gin-gonic/gin` - HTTP framework
- `go-redis/redis/v8` - Redis client
- `stretchr/testify` - Testing framework

## Contributing

1. Follow the existing code structure and patterns
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure all tests pass before submitting PRs
5. Follow Go best practices and conventions
