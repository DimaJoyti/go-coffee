# ArbitrageDetector Architecture Documentation

## Overview

The ArbitrageDetector is a core component of the DeFi module that implements Clean Architecture principles to detect profitable arbitrage opportunities across multiple decentralized exchanges (DEXs).

## Architecture Principles

### Clean Architecture Implementation

The ArbitrageDetector follows Clean Architecture principles:

1. **Dependency Inversion**: Uses interfaces (`PriceProvider`) instead of concrete implementations
2. **Single Responsibility**: Each component has a single, well-defined purpose
3. **Interface Segregation**: Small, focused interfaces
4. **Open/Closed**: Extensible through new PriceProvider implementations

### Key Components

#### 1. ArbitrageDetectorInterface

```go
type ArbitrageDetectorInterface interface {
    Start(ctx context.Context) error
    Stop()
    GetOpportunities(ctx context.Context) ([]*ArbitrageDetection, error)
    DetectArbitrageForToken(ctx context.Context, token Token) ([]*ArbitrageDetection, error)
    SetConfiguration(config ArbitrageConfig) error
    GetMetrics() ArbitrageMetrics
}
```

**Purpose**: Defines the contract for arbitrage detection services, enabling easy testing and multiple implementations.

#### 2. PriceProvider Interface

```go
type PriceProvider interface {
    GetPrice(ctx context.Context, token Token) (decimal.Decimal, error)
    GetExchangeInfo() Exchange
    IsHealthy(ctx context.Context) bool
}
```

**Purpose**: Abstracts price fetching from different exchanges, allowing for:
- Easy addition of new exchanges
- Health checking before price requests
- Consistent error handling
- Better testability

#### 3. Configuration Management

```go
type ArbitrageConfig struct {
    MinProfitMargin  decimal.Decimal `json:"min_profit_margin"`
    MaxGasCost       decimal.Decimal `json:"max_gas_cost"`
    ScanInterval     time.Duration   `json:"scan_interval"`
    MaxOpportunities int             `json:"max_opportunities"`
    EnabledChains    []string        `json:"enabled_chains"`
}
```

**Features**:
- Runtime configuration updates
- Validation of configuration parameters
- Environment-specific settings

#### 4. Metrics and Monitoring

```go
type ArbitrageMetrics struct {
    TotalOpportunities      int64         `json:"total_opportunities"`
    ProfitableOpportunities int64         `json:"profitable_opportunities"`
    AverageProfitMargin     decimal.Decimal `json:"average_profit_margin"`
    LastScanDuration        time.Duration `json:"last_scan_duration"`
    ErrorCount              int64         `json:"error_count"`
    LastError               string        `json:"last_error,omitempty"`
    Uptime                  time.Duration `json:"uptime"`
}
```

**Features**:
- Real-time performance metrics
- Error tracking and reporting
- Running averages for profit margins
- Uptime monitoring

## Data Flow

### 1. Initialization

```
Constructor → Configuration → Price Providers → Start Detection Loop
```

### 2. Detection Process

```
Scan Timer → Get Prices → Calculate Opportunities → Filter by Criteria → Store & Cache → Notify
```

### 3. Price Fetching

```
For each PriceProvider:
  Check Health → Get Price → Handle Errors → Store Result
```

## Error Handling Strategy

### 1. Graceful Degradation
- Continue operation if some price providers fail
- Skip unhealthy providers
- Maintain service availability

### 2. Error Metrics
- Track error counts per provider
- Store last error message
- Monitor error rates

### 3. Retry Logic
- Implement exponential backoff for failed requests
- Circuit breaker pattern for consistently failing providers

## Performance Optimizations

### 1. Concurrent Price Fetching
- Fetch prices from multiple providers in parallel
- Use goroutines with proper synchronization

### 2. Caching Strategy
- Cache opportunities for external access
- Implement TTL-based cache invalidation
- Use Redis for distributed caching

### 3. Memory Management
- Limit number of stored opportunities
- Clean up expired opportunities
- Use object pooling for frequent allocations

## Testing Strategy

### 1. Unit Tests
- Mock PriceProvider interfaces
- Test calculation logic in isolation
- Verify error handling scenarios

### 2. Integration Tests
- Test with real price providers
- Verify end-to-end opportunity detection
- Test configuration updates

### 3. Benchmark Tests
- Measure detection performance
- Profile memory usage
- Test under high load

## Usage Examples

### Basic Usage

```go
// Create price providers
uniswapProvider := NewUniswapPriceProvider(client)
oneInchProvider := NewOneInchPriceProvider(client)
providers := []PriceProvider{uniswapProvider, oneInchProvider}

// Create detector
detector := NewArbitrageDetector(logger, cache, providers)

// Start detection
ctx := context.Background()
if err := detector.Start(ctx); err != nil {
    log.Fatal(err)
}

// Get opportunities
opportunities, err := detector.GetOpportunities(ctx)
if err != nil {
    log.Error(err)
}
```

### Custom Configuration

```go
config := ArbitrageConfig{
    MinProfitMargin:  decimal.NewFromFloat(0.01), // 1%
    MaxGasCost:       decimal.NewFromFloat(0.02), // $20
    ScanInterval:     time.Second * 15,           // 15 seconds
    MaxOpportunities: 50,
    EnabledChains:    []string{"ethereum", "bsc"},
}

detector := NewArbitrageDetectorWithConfig(logger, cache, providers, config)
```

## Future Enhancements

### 1. Machine Learning Integration
- Predict price movements
- Optimize profit calculations
- Risk assessment improvements

### 2. Cross-Chain Arbitrage
- Support for bridge protocols
- Multi-chain opportunity detection
- Gas cost optimization across chains

### 3. Advanced Risk Management
- Dynamic risk scoring
- Portfolio-based risk assessment
- Real-time risk monitoring

## Monitoring and Alerting

### Key Metrics to Monitor
- Opportunity detection rate
- Average profit margins
- Error rates per provider
- Detection latency
- Cache hit rates

### Alerting Thresholds
- Error rate > 10%
- No opportunities detected for > 5 minutes
- Average profit margin < configured minimum
- Detection latency > 30 seconds

## Security Considerations

### 1. Input Validation
- Validate all token addresses
- Sanitize configuration inputs
- Verify price data integrity

### 2. Rate Limiting
- Implement rate limits for external API calls
- Prevent abuse of price providers
- Protect against DDoS attacks

### 3. Access Control
- Secure configuration endpoints
- Authenticate metric access
- Audit opportunity access
