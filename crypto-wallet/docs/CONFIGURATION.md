# ⚙️ Configuration Guide

Complete configuration guide for the Web3 DeFi Algorithmic Trading Platform.

## 📁 Configuration Files

```
configs/
├── development.yaml    # Development environment
├── staging.yaml       # Staging environment
├── production.yaml    # Production environment
└── local.yaml         # Local development overrides
```

## 🔧 Environment Variables

### Database Configuration

```bash
# PostgreSQL Database
DATABASE_URL=postgres://username:password@localhost:5432/defi_trading
DATABASE_MAX_CONNECTIONS=100
DATABASE_MAX_IDLE_CONNECTIONS=10
DATABASE_CONNECTION_LIFETIME=3600

# Redis Cache
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=your-redis-password
REDIS_MAX_CONNECTIONS=100
REDIS_IDLE_TIMEOUT=300
```

### Blockchain Configuration

```bash
# Ethereum
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_INFURA_KEY
ETHEREUM_WS_URL=wss://mainnet.infura.io/ws/v3/YOUR_INFURA_KEY
ETHEREUM_CHAIN_ID=1

# Binance Smart Chain
BSC_RPC_URL=https://bsc-dataseed.binance.org/
BSC_WS_URL=wss://bsc-ws-node.nariox.org:443/ws/v3/YOUR_KEY
BSC_CHAIN_ID=56

# Polygon
POLYGON_RPC_URL=https://polygon-rpc.com/
POLYGON_WS_URL=wss://polygon-ws.com/
POLYGON_CHAIN_ID=137

# Arbitrum
ARBITRUM_RPC_URL=https://arb1.arbitrum.io/rpc
ARBITRUM_WS_URL=wss://arb1.arbitrum.io/ws
ARBITRUM_CHAIN_ID=42161
```

### DeFi Protocol Configuration

```bash
# Uniswap V3
UNISWAP_V3_FACTORY=0x1F98431c8aD98523631AE4a59f267346ea31F984
UNISWAP_V3_ROUTER=0xE592427A0AEce92De3Edee1F18E0157C05861564
UNISWAP_V3_QUOTER=0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6

# Aave V3
AAVE_LENDING_POOL=0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2
AAVE_DATA_PROVIDER=0x7B4EB56E7CD4b454BA8ff71E4518426369a138a3
AAVE_ORACLE=0x54586bE62E3c3580375aE3723C145253060Ca0C2

# 1inch
ONEINCH_API_URL=https://api.1inch.io/v5.0
ONEINCH_API_KEY=your-1inch-api-key

# Chainlink
CHAINLINK_ETH_USD=0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419
CHAINLINK_BTC_USD=0xF4030086522a5bEEa4988F8cA5B36dbC97BeE88c
```

### Security Configuration

```bash
# JWT Authentication
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
JWT_EXPIRATION=3600
JWT_REFRESH_EXPIRATION=86400

# Encryption
ENCRYPTION_KEY=your-32-byte-encryption-key-here
WALLET_ENCRYPTION_ENABLED=true

# API Security
API_RATE_LIMIT_ENABLED=true
API_RATE_LIMIT_REQUESTS=1000
API_RATE_LIMIT_WINDOW=60

# CORS
CORS_ALLOWED_ORIGINS=https://app.defi-trading.com,https://dashboard.defi-trading.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
```

### Trading Configuration

```bash
# Trading Bot Settings
TRADING_ENABLED=true
TRADING_MAX_POSITION_SIZE=50000
TRADING_MIN_PROFIT_MARGIN=0.005
TRADING_MAX_SLIPPAGE=0.01
TRADING_GAS_PRICE_MULTIPLIER=1.1

# Arbitrage Settings
ARBITRAGE_ENABLED=true
ARBITRAGE_MIN_PROFIT_USD=10
ARBITRAGE_MAX_GAS_COST_USD=50
ARBITRAGE_CONFIDENCE_THRESHOLD=0.8

# Yield Farming Settings
YIELD_FARMING_ENABLED=true
YIELD_FARMING_MIN_APY=0.05
YIELD_FARMING_AUTO_COMPOUND=true
YIELD_FARMING_REBALANCE_THRESHOLD=0.02
```

### Monitoring Configuration

```bash
# Prometheus Metrics
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
PROMETHEUS_PATH=/metrics

# Jaeger Tracing
JAEGER_ENABLED=true
JAEGER_ENDPOINT=http://localhost:14268/api/traces
JAEGER_SERVICE_NAME=defi-trading-platform

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_FILE_PATH=/var/log/defi-trading.log
```

## 📄 YAML Configuration

### Development Configuration (`configs/development.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  host: "localhost"
  port: 5432
  name: "defi_trading_dev"
  user: "postgres"
  password: "password"
  ssl_mode: "disable"
  max_connections: 25
  max_idle_connections: 5

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  max_connections: 10

blockchain:
  ethereum:
    rpc_url: "https://goerli.infura.io/v3/YOUR_KEY"
    ws_url: "wss://goerli.infura.io/ws/v3/YOUR_KEY"
    chain_id: 5
  
trading:
  enabled: true
  max_position_size: 1000
  min_profit_margin: 0.01
  max_slippage: 0.02
  
arbitrage:
  enabled: true
  min_profit_usd: 5
  max_gas_cost_usd: 25
  
yield_farming:
  enabled: true
  min_apy: 0.03
  auto_compound: true

logging:
  level: "debug"
  format: "text"
  output: "stdout"
```

### Production Configuration (`configs/production.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/server.crt"
    key_file: "/etc/ssl/private/server.key"

database:
  host: "${DATABASE_HOST}"
  port: 5432
  name: "${DATABASE_NAME}"
  user: "${DATABASE_USER}"
  password: "${DATABASE_PASSWORD}"
  ssl_mode: "require"
  max_connections: 100
  max_idle_connections: 10
  connection_lifetime: 3600

redis:
  host: "${REDIS_HOST}"
  port: 6379
  password: "${REDIS_PASSWORD}"
  db: 0
  max_connections: 50
  cluster_enabled: true
  cluster_nodes:
    - "${REDIS_NODE_1}:6379"
    - "${REDIS_NODE_2}:6379"
    - "${REDIS_NODE_3}:6379"

blockchain:
  ethereum:
    rpc_url: "${ETHEREUM_RPC_URL}"
    ws_url: "${ETHEREUM_WS_URL}"
    chain_id: 1
    gas_price_multiplier: 1.1
  bsc:
    rpc_url: "${BSC_RPC_URL}"
    ws_url: "${BSC_WS_URL}"
    chain_id: 56
  polygon:
    rpc_url: "${POLYGON_RPC_URL}"
    ws_url: "${POLYGON_WS_URL}"
    chain_id: 137

defi_protocols:
  uniswap_v3:
    factory: "0x1F98431c8aD98523631AE4a59f267346ea31F984"
    router: "0xE592427A0AEce92De3Edee1F18E0157C05861564"
    quoter: "0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6"
  aave_v3:
    lending_pool: "0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2"
    data_provider: "0x7B4EB56E7CD4b454BA8ff71E4518426369a138a3"
  oneinch:
    api_url: "https://api.1inch.io/v5.0"
    api_key: "${ONEINCH_API_KEY}"

trading:
  enabled: true
  max_position_size: 50000
  min_profit_margin: 0.005
  max_slippage: 0.01
  max_daily_trades: 1000
  risk_management:
    stop_loss_enabled: true
    take_profit_enabled: true
    position_size_limit: 0.1  # 10% of portfolio

arbitrage:
  enabled: true
  min_profit_usd: 10
  max_gas_cost_usd: 50
  confidence_threshold: 0.8
  execution_timeout: 30s

yield_farming:
  enabled: true
  min_apy: 0.05
  auto_compound: true
  rebalance_threshold: 0.02
  max_impermanent_loss: 0.1

security:
  jwt:
    secret: "${JWT_SECRET}"
    expiration: 3600
  encryption:
    key: "${ENCRYPTION_KEY}"
    enabled: true
  rate_limiting:
    enabled: true
    requests_per_minute: 1000
    burst_size: 100

monitoring:
  prometheus:
    enabled: true
    port: 9090
    path: "/metrics"
  jaeger:
    enabled: true
    endpoint: "${JAEGER_ENDPOINT}"
    service_name: "defi-trading-platform"
  health_check:
    enabled: true
    path: "/health"
    interval: 30s

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/defi-trading.log"
  rotation:
    enabled: true
    max_size: "100MB"
    max_age: "7d"
    max_backups: 10
```

## 🔒 Security Best Practices

### Environment Variables Security

1. **Never commit secrets to version control**
2. **Use environment-specific .env files**
3. **Rotate secrets regularly**
4. **Use secret management tools (Vault, AWS Secrets Manager)**

### Database Security

```yaml
database:
  ssl_mode: "require"
  connection_encryption: true
  backup_encryption: true
  access_logging: true
```

### API Security

```yaml
security:
  cors:
    allowed_origins: ["https://trusted-domain.com"]
    credentials: true
  headers:
    x_frame_options: "DENY"
    x_content_type_options: "nosniff"
    x_xss_protection: "1; mode=block"
```

## 📊 Performance Tuning

### Database Optimization

```yaml
database:
  max_connections: 100
  max_idle_connections: 10
  connection_lifetime: 3600
  query_timeout: 30s
  slow_query_log: true
  slow_query_threshold: 1s
```

### Redis Optimization

```yaml
redis:
  max_connections: 50
  idle_timeout: 300s
  read_timeout: 5s
  write_timeout: 5s
  pool_size: 20
```

### Trading Performance

```yaml
trading:
  execution_workers: 10
  order_queue_size: 1000
  price_update_interval: 1s
  arbitrage_scan_interval: 5s
```

## 🚀 Deployment Configurations

### Docker Environment

```bash
# Docker Compose Environment
COMPOSE_PROJECT_NAME=defi-trading
COMPOSE_FILE=docker-compose.prod.yml

# Container Resources
CONTAINER_MEMORY_LIMIT=2g
CONTAINER_CPU_LIMIT=1.0
```

### Kubernetes Configuration

```yaml
# ConfigMap for application config
apiVersion: v1
kind: ConfigMap
metadata:
  name: defi-trading-config
data:
  config.yaml: |
    server:
      port: 8080
    database:
      max_connections: 100
    # ... rest of config
```

## 🔧 Configuration Validation

### Validation Rules

```go
// Configuration validation
type Config struct {
    Server   ServerConfig   `validate:"required"`
    Database DatabaseConfig `validate:"required"`
    Trading  TradingConfig  `validate:"required"`
}

type TradingConfig struct {
    MaxPositionSize  decimal.Decimal `validate:"required,gt=0"`
    MinProfitMargin  decimal.Decimal `validate:"required,gte=0,lte=1"`
    MaxSlippage      decimal.Decimal `validate:"required,gte=0,lte=1"`
}
```

### Environment Validation

```bash
# Validate configuration before startup
go run cmd/validate-config/main.go --config=configs/production.yaml
```

---

**🚀 Need help with configuration? Check our [troubleshooting guide](TROUBLESHOOTING.md)!**
