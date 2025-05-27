# DeFi Protocol Integration Documentation

## Overview

This document outlines the comprehensive DeFi protocol integration for the go-coffee Web3 wallet backend. The integration enables users to interact with various DeFi protocols while purchasing coffee with cryptocurrency, earning rewards, and participating in yield farming.

## Architecture

### Core Components

1. **DeFi Service** - Central service for DeFi protocol interactions
2. **Protocol Clients** - Specialized clients for each DeFi protocol
3. **Coffee Token** - Utility token for the coffee ecosystem
4. **Price Oracles** - Real-time price feeds via Chainlink
5. **DEX Aggregation** - Best price routing via 1inch

### Supported Protocols

#### 1. Uniswap V3
- **Purpose**: Decentralized token swaps and liquidity provision
- **Features**:
  - Token swaps with optimal routing
  - Liquidity pool management
  - Concentrated liquidity positions
  - Fee tier selection (0.05%, 0.3%, 1%)

#### 2. Aave V3
- **Purpose**: Lending and borrowing protocols
- **Features**:
  - Supply tokens to earn interest
  - Borrow against collateral
  - Flash loans for arbitrage
  - Variable and stable interest rates

#### 3. Chainlink
- **Purpose**: Decentralized price oracles
- **Features**:
  - Real-time price feeds
  - Historical price data
  - Multiple asset support
  - High reliability and accuracy

#### 4. 1inch
- **Purpose**: DEX aggregation for best prices
- **Features**:
  - Multi-DEX price comparison
  - Optimal swap routing
  - Gas optimization
  - Slippage protection

#### 5. Coffee Token (COFFEE)
- **Purpose**: Utility token for coffee ecosystem
- **Features**:
  - Staking rewards (12% APY)
  - Coffee purchase discounts
  - Governance voting rights
  - Loyalty program integration

## Coffee-Specific DeFi Features

### 1. Coffee Token Economics

```
Total Supply: 1,000,000,000 COFFEE
Distribution:
- 40% - Public Sale
- 25% - Coffee Rewards Pool
- 15% - Team & Advisors (vested)
- 10% - Ecosystem Development
- 5% - Marketing & Partnerships
- 5% - Liquidity Provision
```

### 2. Staking Rewards System

- **APY**: 12% annual percentage yield
- **Minimum Stake**: 100 COFFEE tokens
- **Lock Period**: No lock period (flexible staking)
- **Rewards**: Paid in COFFEE tokens
- **Compound**: Auto-compound option available

### 3. Coffee Purchase Benefits

- **Payment Flexibility**: Pay with any ERC-20 token
- **Auto-Swap**: Automatic conversion to payment currency
- **Discounts**: 5% discount when paying with COFFEE tokens
- **Cashback**: 2% cashback in COFFEE tokens for all purchases
- **Loyalty Tiers**: Bronze, Silver, Gold, Platinum based on COFFEE holdings

### 4. Yield Farming Opportunities

- **COFFEE-ETH LP**: 25% APY
- **COFFEE-USDC LP**: 20% APY
- **COFFEE-BTC LP**: 18% APY
- **Single Asset Staking**: 12% APY

## API Endpoints

### Token Price Endpoints

```http
GET /api/v1/defi/prices/{tokenAddress}
GET /api/v1/defi/prices/multiple?tokens=eth,btc,usdc
GET /api/v1/defi/prices/historical/{tokenAddress}?timestamp={unix_timestamp}
```

### Swap Endpoints

```http
POST /api/v1/defi/swap/quote
POST /api/v1/defi/swap/execute
GET /api/v1/defi/swap/history/{userID}
```

### Liquidity Pool Endpoints

```http
GET /api/v1/defi/pools
GET /api/v1/defi/pools/{poolID}
POST /api/v1/defi/pools/{poolID}/add-liquidity
POST /api/v1/defi/pools/{poolID}/remove-liquidity
```

### Lending Endpoints

```http
GET /api/v1/defi/lending/positions/{userID}
POST /api/v1/defi/lending/supply
POST /api/v1/defi/lending/borrow
POST /api/v1/defi/lending/repay
POST /api/v1/defi/lending/withdraw
```

### Coffee Token Endpoints

```http
GET /api/v1/coffee/token/info
GET /api/v1/coffee/token/balance/{address}
POST /api/v1/coffee/staking/stake
POST /api/v1/coffee/staking/unstake
POST /api/v1/coffee/staking/claim-rewards
GET /api/v1/coffee/staking/positions/{userID}
```

## Integration Examples

### 1. Coffee Purchase with Auto-Swap

```javascript
// User wants to buy coffee with ETH, but shop accepts USDC
const purchaseRequest = {
  shopId: "coffee-shop-123",
  items: [
    { productId: "espresso", quantity: 2, price: 3.50 }
  ],
  paymentToken: "ETH",
  paymentAmount: "0.003", // $7.50 worth of ETH
  targetToken: "USDC"
};

// 1. Get swap quote
const swapQuote = await defiService.getSwapQuote({
  tokenIn: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
  tokenOut: "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1", // USDC
  amountIn: "0.003",
  chain: "ethereum"
});

// 2. Execute swap
const swapResult = await defiService.executeSwap({
  quoteId: swapQuote.id,
  userId: "user-123",
  walletId: "wallet-456"
});

// 3. Process coffee payment
const paymentResult = await coffeeService.processPayment({
  orderId: "order-789",
  transactionHash: swapResult.transactionHash,
  amount: swapQuote.amountOut,
  currency: "USDC"
});
```

### 2. Coffee Token Staking

```javascript
// Stake COFFEE tokens for rewards
const stakingRequest = {
  userId: "user-123",
  amount: "1000", // 1000 COFFEE tokens
  chain: "ethereum"
};

const stakingResult = await coffeeTokenService.stake(stakingRequest);

// Check pending rewards
const pendingRewards = await coffeeTokenService.calculatePendingRewards({
  stakingId: stakingResult.stakingId
});

// Claim rewards
const claimResult = await coffeeTokenService.claimRewards({
  stakingId: stakingResult.stakingId
});
```

### 3. Yield Farming

```javascript
// Add liquidity to COFFEE-ETH pool
const liquidityRequest = {
  userId: "user-123",
  poolId: "coffee-eth-pool",
  amount0: "1000", // 1000 COFFEE
  amount1: "0.4",  // 0.4 ETH
  slippage: 0.005  // 0.5%
};

const liquidityResult = await defiService.addLiquidity(liquidityRequest);

// Stake LP tokens for additional rewards
const farmingRequest = {
  userId: "user-123",
  farmId: "coffee-eth-farm",
  lpTokens: liquidityResult.lpTokens
};

const farmingResult = await defiService.stakeInFarm(farmingRequest);
```

## Security Considerations

### 1. Smart Contract Security
- Multi-signature wallet for contract ownership
- Time-locked upgrades with 48-hour delay
- Regular security audits by reputable firms
- Bug bounty program for vulnerability disclosure

### 2. Price Oracle Security
- Multiple price feed sources
- Price deviation checks
- Circuit breakers for extreme price movements
- Fallback oracle mechanisms

### 3. User Fund Protection
- Non-custodial architecture
- Hardware wallet integration
- Multi-factor authentication
- Transaction signing verification

### 4. Protocol Risk Management
- Diversified protocol integration
- Liquidity monitoring
- Slippage protection
- Maximum exposure limits

## Monitoring and Analytics

### 1. Key Metrics
- Total Value Locked (TVL)
- Daily Active Users (DAU)
- Transaction Volume
- Coffee Token Price
- Staking Participation Rate
- Yield Farming APY

### 2. Alerts and Notifications
- Price movement alerts
- Liquidation warnings
- Reward claim reminders
- Protocol upgrade notifications

### 3. Performance Monitoring
- Transaction success rates
- Average confirmation times
- Gas cost optimization
- API response times

## Deployment Guide

### 1. Prerequisites
- Go 1.24+
- PostgreSQL 13+
- Redis 6+
- Ethereum node access
- Infura/Alchemy API keys

### 2. Environment Setup
```bash
# Clone repository
git clone https://github.com/yourusername/go-coffee.git
cd go-coffee/web3-wallet-backend

# Install dependencies
go mod tidy

# Set up configuration
cp config/config.yaml.example config/config.yaml
# Edit config.yaml with your settings

# Run database migrations
go run db/migrate.go -up

# Start DeFi service
go run cmd/defi-service/main.go
```

### 3. Smart Contract Deployment
```bash
# Deploy Coffee Token contract
npx hardhat deploy --network mainnet --tags CoffeeToken

# Verify contract on Etherscan
npx hardhat verify --network mainnet <CONTRACT_ADDRESS>

# Update configuration with deployed addresses
```

## Future Roadmap

### Phase 1 (Current)
- ✅ Core DeFi protocol integration
- ✅ Coffee Token implementation
- ✅ Basic staking rewards
- ✅ Price oracle integration

### Phase 2 (Q2 2024)
- 🔄 Cross-chain bridge integration
- 🔄 Advanced yield farming strategies
- 🔄 Coffee NFT marketplace
- 🔄 DAO governance implementation

### Phase 3 (Q3 2024)
- ⏳ Coffee futures trading
- ⏳ Insurance protocol integration
- ⏳ Mobile app DeFi features
- ⏳ Institutional DeFi services

### Phase 4 (Q4 2024)
- ⏳ Layer 2 scaling solutions
- ⏳ Advanced trading features
- ⏳ Coffee supply chain tracking
- ⏳ Global expansion

## Support and Resources

- **Documentation**: [docs.go-coffee.com](https://docs.go-coffee.com)
- **API Reference**: [api.go-coffee.com](https://api.go-coffee.com)
- **Discord Community**: [discord.gg/go-coffee](https://discord.gg/go-coffee)
- **GitHub Issues**: [github.com/yourusername/go-coffee/issues](https://github.com/yourusername/go-coffee/issues)
- **Email Support**: defi-support@go-coffee.com
