# TradingView Integration & Portfolio Manager Dashboard

## Overview

This document describes the comprehensive TradingView integration and portfolio manager dashboard built using Bright Data MCP for scraping real-time crypto market data and creating stunning visualizations for professional crypto portfolio managers.

## Features

### 🚀 TradingView Data Integration

- **Real-time Market Data**: Live cryptocurrency prices, market caps, and trading volumes
- **Market Overview**: Total market cap, BTC/ETH dominance, Fear & Greed Index
- **Trending Analysis**: Trending coins with social mentions and trend scores
- **Top Movers**: Real-time gainers and losers with percentage changes
- **Technical Ratings**: Buy/Sell/Neutral ratings from TradingView analysis
- **Data Quality Monitoring**: Real-time data quality metrics and validation

### 📊 Portfolio Analytics Engine

- **Comprehensive Portfolio Tracking**: Multi-portfolio support with real-time valuations
- **Performance Metrics**: Sharpe ratio, Sortino ratio, Alpha, Beta, and more
- **Risk Management**: Value at Risk (VaR), stress testing, correlation analysis
- **Asset Allocation**: Dynamic allocation tracking with rebalancing recommendations
- **Diversification Analysis**: Sector, geographic, and market cap diversification metrics

### 🎨 Stunning Visualizations

- **Interactive Market Heatmap**: Size-based market cap visualization with performance colors
- **Portfolio Dashboard**: Real-time portfolio value, returns, and risk metrics
- **Correlation Matrix**: Asset correlation analysis for risk management
- **Sector Performance**: Sector-based performance tracking and analysis
- **Risk Assessment**: Visual risk scoring and stress test scenarios

### 🔍 Advanced Analytics

- **Stress Testing**: Multiple scenario analysis (Market Crash, Regulatory, Exchange Hack)
- **Correlation Analysis**: Real-time correlation tracking between assets
- **Performance Attribution**: Detailed breakdown of portfolio performance drivers
- **Risk-Adjusted Returns**: Comprehensive risk-adjusted performance metrics

## API Endpoints

### TradingView Data Endpoints

```
GET /api/v2/tradingview/market-data          # Complete TradingView market data
GET /api/v2/tradingview/coins                # Cryptocurrency listings
GET /api/v2/tradingview/trending             # Trending cryptocurrencies
GET /api/v2/tradingview/gainers              # Top gaining coins
GET /api/v2/tradingview/losers               # Top losing coins
GET /api/v2/tradingview/market-overview      # Market overview data
```

### Portfolio Analytics Endpoints

```
GET /api/v2/portfolio/{portfolioId}/analytics     # Complete portfolio analytics
GET /api/v2/portfolio/{portfolioId}/risk-metrics  # Risk metrics and VaR
GET /api/v2/portfolio/{portfolioId}/performance   # Performance metrics
```

### Market Visualization Endpoints

```
GET /api/v2/market/heatmap                   # Market heatmap data
GET /api/v2/market/sectors                   # Sector performance data
GET /api/v2/market/correlation               # Asset correlation matrix
```

## Data Models

### TradingView Data Structure

```typescript
interface TradingViewData {
  coins: TradingViewCoin[];
  marketOverview: MarketOverview;
  trendingCoins: TrendingCoin[];
  gainers: TradingViewCoin[];
  losers: TradingViewCoin[];
  lastUpdated: string;
  dataQuality: number;
}

interface TradingViewCoin {
  symbol: string;
  name: string;
  price: number;
  change24h: number;
  marketCap: number;
  volume24h: number;
  socialDominance: number;
  techRating: string;
  rank: number;
  category: string[];
}
```

### Portfolio Analytics Structure

```typescript
interface PortfolioAnalytics {
  portfolioId: string;
  totalValue: number;
  totalReturn: number;
  totalReturnPct: number;
  holdings: PortfolioHolding[];
  performance: PerformanceMetrics;
  risk: RiskMetrics;
  diversification: DiversificationMetrics;
}

interface RiskMetrics {
  var95: number;
  var99: number;
  portfolioVol: number;
  riskScore: number;
  stressTests: StressTestResult[];
}
```

## Components

### Portfolio Manager Dashboard (`PortfolioManagerDashboard.tsx`)

- **Real-time Portfolio Tracking**: Live portfolio value and performance updates
- **Holdings Management**: Detailed view of all portfolio holdings
- **Performance Analytics**: Comprehensive performance metrics display
- **Risk Analysis**: Visual risk assessment and monitoring

### Market Heatmap (`MarketHeatmap.tsx`)

- **Interactive Visualization**: Click-to-explore market heatmap
- **Sector Analysis**: Sector-based performance breakdown
- **Color-coded Performance**: Intuitive color coding for gains/losses
- **Real-time Updates**: Live market data updates

### TradingView Widget (`TradingViewWidget.tsx`)

- **Market Overview**: Complete market statistics and sentiment
- **Coin Listings**: Searchable and filterable cryptocurrency listings
- **Trending Analysis**: Social sentiment and trending analysis
- **Technical Ratings**: Professional buy/sell/neutral ratings

## Bright Data Integration

### Data Sources

- **TradingView.com**: Primary source for market data and technical analysis
- **Real-time Scraping**: Live data collection using Bright Data MCP
- **Data Validation**: Quality checks and validation for all scraped data
- **Caching Strategy**: Redis-based caching for optimal performance

### Scraping Implementation

```go
// ScrapeTradingView scrapes crypto market data from TradingView
func (mi *MarketIntelligence) ScrapeTradingView(ctx context.Context) (*TradingViewData, error) {
    // Uses scrape_as_markdown_Bright_Data MCP function
    tradingViewData := &TradingViewData{
        Coins:       mi.parseTradingViewCoins(),
        MarketOverview: mi.parseMarketOverview(),
        TrendingCoins: mi.parseTrendingCoins(),
        // ... additional parsing
    }
    return tradingViewData, nil
}
```

## Usage

### Accessing the Portfolio Manager

1. Navigate to `/portfolio-manager` in the application
2. Select a portfolio from the available options
3. Explore different tabs: Dashboard, Market Data, Heatmap, Analytics, Data Quality

### Key Features Usage

- **Portfolio Selection**: Choose from multiple portfolios with different risk profiles
- **Real-time Monitoring**: Auto-refresh every 30 seconds for live data
- **Risk Assessment**: Monitor VaR, stress tests, and correlation metrics
- **Performance Tracking**: Track Sharpe ratio, alpha, and other key metrics

## Configuration

### Environment Variables

```bash
BRIGHT_DATA_ENABLED=true
BRIGHT_DATA_UPDATE_INTERVAL=30s
BRIGHT_DATA_CACHE_TTL=5m
BRIGHT_DATA_RATE_LIMIT_RPS=10
```

### Portfolio Configuration

```yaml
portfolios:
  - id: "portfolio-1"
    name: "Institutional Portfolio"
    risk_profile: "moderate"
    rebalance_threshold: 0.05
  - id: "portfolio-2"
    name: "DeFi Growth Fund"
    risk_profile: "aggressive"
    rebalance_threshold: 0.10
```

## Performance Optimization

- **Caching Strategy**: Redis-based caching for frequently accessed data
- **Data Compression**: Efficient data structures for large datasets
- **Lazy Loading**: Progressive loading of heavy components
- **Real-time Updates**: WebSocket connections for live data streams

## Security Considerations

- **Data Validation**: All scraped data is validated before processing
- **Rate Limiting**: Implemented to prevent API abuse
- **Error Handling**: Comprehensive error handling and fallback mechanisms
- **Data Privacy**: No sensitive portfolio data is stored externally

## Future Enhancements

- **Machine Learning Integration**: AI-powered portfolio optimization
- **Advanced Charting**: Interactive charts with technical indicators
- **Alert System**: Custom alerts for portfolio and market events
- **Mobile Optimization**: Responsive design for mobile devices
- **Export Functionality**: PDF reports and data export capabilities

## Support

For technical support or feature requests, please refer to the main project documentation or contact the development team.
