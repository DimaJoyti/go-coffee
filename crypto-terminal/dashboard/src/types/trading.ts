export interface Price {
  symbol: string
  price: number
  volume24h: number
  change24h: number
  timestamp: string
  source: string
}

export interface MarketData {
  symbol: string
  name: string
  currentPrice: number
  marketCap: number
  marketCapRank: number
  volume24h: number
  change24h: number
  change7d: number
  change30d: number
  high24h: number
  low24h: number
  circulatingSupply: number
  totalSupply: number
  maxSupply: number
  ath: number
  athDate: string
  atl: number
  atlDate: string
  lastUpdated: string
}

export interface Portfolio {
  id: string
  userId: string
  name: string
  totalValue: number
  totalCost: number
  totalPnL: number
  totalPnLPercent: number
  dayChange: number
  dayChangePercent: number
  createdAt: string
  updatedAt: string
  holdings: Holding[]
}

export interface Holding {
  id: string
  portfolioId: string
  symbol: string
  name: string
  amount: number
  averagePrice: number
  currentPrice: number
  totalValue: number
  totalCost: number
  pnl: number
  pnlPercent: number
  allocation: number
  lastUpdated: string
}

export interface CoffeeStrategy {
  id: string
  name: string
  type: 'espresso' | 'latte' | 'cold-brew' | 'cappuccino'
  symbol: string
  status: 'active' | 'inactive' | 'paused'
  description: string
  emoji: string
  performance: {
    totalTrades: number
    winRate: number
    totalPnL: number
    totalPnLPercent: number
    avgTradeSize: number
    maxDrawdown: number
  }
  settings: {
    riskLevel: 'low' | 'medium' | 'high'
    maxPositionSize: number
    stopLoss: number
    takeProfit: number
  }
  createdAt: string
  lastTradeAt?: string
}

export interface Trade {
  id: string
  strategyId: string
  symbol: string
  side: 'buy' | 'sell'
  type: 'market' | 'limit' | 'stop'
  amount: number
  price: number
  executedPrice?: number
  status: 'pending' | 'filled' | 'cancelled' | 'rejected'
  pnl?: number
  pnlPercent?: number
  createdAt: string
  executedAt?: string
}

export interface WebSocketMessage {
  type: 'price_update' | 'signal_alert' | 'trade_execution' | 'portfolio_update' | 'risk_alert'
  data: any
  timestamp: string
}

export interface PriceUpdate {
  symbol: string
  price: number
  change: number
  changePercent: number
  volume: number
  timestamp: string
}

export interface SignalAlert {
  strategy: string
  symbol: string
  signal: 'BUY' | 'SELL' | 'HOLD'
  confidence: number
  message: string
  emoji: string
  timestamp: string
}

export interface TradeExecution {
  tradeId: string
  strategy: string
  symbol: string
  side: 'buy' | 'sell'
  amount: number
  price: number
  status: 'filled' | 'rejected'
  message: string
  timestamp: string
}

export interface PortfolioUpdate {
  portfolioId: string
  totalValue: number
  totalPnL: number
  totalPnLPercent: number
  dayChange: number
  dayChangePercent: number
  updatedHoldings: Holding[]
  timestamp: string
}

export interface RiskAlert {
  type: 'position_size' | 'drawdown' | 'volatility' | 'correlation'
  severity: 'low' | 'medium' | 'high' | 'critical'
  message: string
  symbol?: string
  portfolioId?: string
  recommendation: string
  timestamp: string
}

export interface ArbitrageOpportunity {
  symbol: string
  buyExchange: string
  sellExchange: string
  buyPrice: number
  sellPrice: number
  spread: number
  spreadPercent: number
  volume: number
  profitPotential: number
  timestamp: string
}

export interface MarketOverview {
  totalMarketCap: number
  totalVolume24h: number
  marketCapChange24h: number
  volumeChange24h: number
  btcDominance: number
  ethDominance: number
  activeCoins: number
  markets: number
  fearGreedIndex: number
  timestamp: string
}
