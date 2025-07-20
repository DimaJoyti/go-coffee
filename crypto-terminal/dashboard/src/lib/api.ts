import axios from 'axios'
import type {
  Portfolio,
  CoffeeStrategy,
  MarketData,
  Trade,
  ArbitrageOpportunity,
  MarketOverview,
} from '@/types/trading'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090'

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor for auth tokens
api.interceptors.request.use(
  (config) => {
    // Add auth token if available
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized access
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Portfolio API
export const portfolioApi = {
  getPortfolios: async (): Promise<Portfolio[]> => {
    const response = await api.get('/api/v1/coffee-trading/portfolio')
    return response.data.data || []
  },

  getPortfolioPerformance: async (portfolioId: string, timeRange: string = '30D') => {
    const response = await api.get(`/api/v1/coffee-trading/portfolio/${portfolioId}/performance`, {
      params: { timeRange }
    })
    return response.data.data
  },

  getPortfolioRiskMetrics: async (portfolioId: string) => {
    const response = await api.get(`/api/v1/coffee-trading/portfolio/${portfolioId}/risk-metrics`)
    return response.data.data
  },

  syncPortfolio: async (portfolioId: string) => {
    const response = await api.post(`/api/v1/coffee-trading/portfolio/${portfolioId}/sync`)
    return response.data.data
  },
}

// Coffee Strategy API
export const strategyApi = {
  getStrategies: async (): Promise<CoffeeStrategy[]> => {
    const response = await api.get('/api/v1/coffee-trading/strategies')
    return response.data.data || []
  },

  getCoffeeMenu: async () => {
    const response = await api.get('/api/v1/coffee-trading/coffee/menu')
    return response.data.data
  },

  startEspressoStrategy: async (data: { symbol: string; name: string }) => {
    const response = await api.post('/api/v1/coffee-trading/coffee/espresso/start', data)
    return response.data.data
  },

  startLatteStrategy: async (data: { symbol: string; name: string }) => {
    const response = await api.post('/api/v1/coffee-trading/coffee/latte/start', data)
    return response.data.data
  },

  startColdBrewStrategy: async (data: { symbol: string; name: string }) => {
    const response = await api.post('/api/v1/coffee-trading/coffee/cold-brew/start', data)
    return response.data.data
  },

  startCappuccinoStrategy: async (data: { symbol: string; name: string }) => {
    const response = await api.post('/api/v1/coffee-trading/coffee/cappuccino/start', data)
    return response.data.data
  },

  startStrategy: async (strategyId: string) => {
    const response = await api.post(`/api/v1/coffee-trading/strategies/${strategyId}/start`)
    return response.data.data
  },

  stopStrategy: async (strategyId: string) => {
    const response = await api.post(`/api/v1/coffee-trading/strategies/${strategyId}/stop`)
    return response.data.data
  },

  getStrategyRecommendations: async () => {
    const response = await api.get('/api/v1/coffee-trading/coffee/recommendations')
    return response.data.data
  },
}

// Market Data API
export const marketApi = {
  getPrices: async (): Promise<MarketData[]> => {
    const response = await api.get('/api/v1/market/prices')
    return response.data.data || []
  },

  getPrice: async (symbol: string): Promise<MarketData> => {
    const response = await api.get(`/api/v1/market/prices/${symbol}`)
    return response.data.data
  },

  getMarketOverview: async (): Promise<MarketOverview> => {
    const response = await api.get('/api/v1/market/overview')
    return response.data.data
  },

  getTopGainers: async () => {
    const response = await api.get('/api/v1/market/gainers')
    return response.data.data || []
  },

  getTopLosers: async () => {
    const response = await api.get('/api/v1/market/losers')
    return response.data.data || []
  },

  getTradingViewData: async () => {
    const response = await api.get('/api/v1/tradingview/market-data')
    return response.data.data
  },

  getMarketHeatmap: async () => {
    const response = await api.get('/api/v1/market/heatmap')
    return response.data.data
  },
}

// Analytics API
export const analyticsApi = {
  getDashboard: async () => {
    const response = await api.get('/api/v1/coffee-trading/analytics/dashboard')
    return response.data.data
  },

  getPerformance: async () => {
    const response = await api.get('/api/v1/coffee-trading/analytics/performance')
    return response.data.data
  },
}

// Arbitrage API
export const arbitrageApi = {
  getOpportunities: async (): Promise<ArbitrageOpportunity[]> => {
    const response = await api.get('/api/v2/arbitrage/opportunities')
    return response.data.data || []
  },

  getOpportunity: async (symbol: string): Promise<ArbitrageOpportunity> => {
    const response = await api.get(`/api/v2/arbitrage/opportunities/${symbol}`)
    return response.data.data
  },
}

// Enhanced Market Data API
export const enhancedMarketApi = {
  getAggregatedPrice: async (symbol: string) => {
    const response = await api.get(`/api/v2/market/aggregated/${symbol}`)
    return response.data.data
  },

  getBestPrices: async (symbol: string) => {
    const response = await api.get(`/api/v2/market/best-prices/${symbol}`)
    return response.data.data
  },

  getExchangeStatus: async () => {
    const response = await api.get('/api/v2/market/exchanges/status')
    return response.data.data
  },

  getDataQuality: async () => {
    const response = await api.get('/api/v2/market/data-quality')
    return response.data.data
  },
}

export default api
