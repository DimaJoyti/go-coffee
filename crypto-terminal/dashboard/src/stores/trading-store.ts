import { create } from 'zustand'
import { devtools, subscribeWithSelector } from 'zustand/middleware'
import type {
  Portfolio,
  CoffeeStrategy,
  MarketData,
  PriceUpdate,
  SignalAlert,
  TradeExecution,
  PortfolioUpdate,
  RiskAlert,
  ArbitrageOpportunity,
  MarketOverview,
} from '@/types/trading'

interface TradingState {
  // Portfolio state
  portfolios: Portfolio[]
  selectedPortfolioId: string | null
  portfolioLoading: boolean
  portfolioError: string | null

  // Market data state
  marketData: Record<string, MarketData>
  marketOverview: MarketOverview | null
  marketLoading: boolean
  marketError: string | null

  // Coffee strategies state
  strategies: CoffeeStrategy[]
  strategiesLoading: boolean
  strategiesError: string | null

  // Real-time updates
  priceUpdates: Record<string, PriceUpdate>
  signalAlerts: SignalAlert[]
  tradeExecutions: TradeExecution[]
  riskAlerts: RiskAlert[]

  // Arbitrage opportunities
  arbitrageOpportunities: ArbitrageOpportunity[]

  // WebSocket connection state
  isConnected: boolean
  connectionError: string | null

  // UI state
  theme: 'light' | 'dark'
  sidebarCollapsed: boolean
  activeTab: string
}

interface TradingActions {
  // Portfolio actions
  setPortfolios: (portfolios: Portfolio[]) => void
  setSelectedPortfolioId: (id: string | null) => void
  setPortfolioLoading: (loading: boolean) => void
  setPortfolioError: (error: string | null) => void
  updatePortfolio: (portfolioUpdate: PortfolioUpdate) => void

  // Market data actions
  setMarketData: (data: MarketData[]) => void
  updateMarketData: (symbol: string, data: MarketData) => void
  setMarketOverview: (overview: MarketOverview) => void
  setMarketLoading: (loading: boolean) => void
  setMarketError: (error: string | null) => void
  updatePrice: (priceUpdate: PriceUpdate) => void

  // Strategy actions
  setStrategies: (strategies: CoffeeStrategy[]) => void
  updateStrategy: (strategyId: string, updates: Partial<CoffeeStrategy>) => void
  setStrategiesLoading: (loading: boolean) => void
  setStrategiesError: (error: string | null) => void

  // Real-time update actions
  addSignalAlert: (alert: SignalAlert) => void
  addTradeExecution: (execution: TradeExecution) => void
  addRiskAlert: (alert: RiskAlert) => void
  clearOldAlerts: () => void

  // Arbitrage actions
  setArbitrageOpportunities: (opportunities: ArbitrageOpportunity[]) => void
  updateArbitrageOpportunity: (opportunity: ArbitrageOpportunity) => void

  // WebSocket actions
  setConnectionStatus: (connected: boolean) => void
  setConnectionError: (error: string | null) => void

  // UI actions
  setTheme: (theme: 'light' | 'dark') => void
  toggleSidebar: () => void
  setActiveTab: (tab: string) => void

  // Utility actions
  reset: () => void
}

type TradingStore = TradingState & TradingActions

const initialState: TradingState = {
  portfolios: [],
  selectedPortfolioId: null,
  portfolioLoading: false,
  portfolioError: null,

  marketData: {},
  marketOverview: null,
  marketLoading: false,
  marketError: null,

  strategies: [],
  strategiesLoading: false,
  strategiesError: null,

  priceUpdates: {},
  signalAlerts: [],
  tradeExecutions: [],
  riskAlerts: [],

  arbitrageOpportunities: [],

  isConnected: false,
  connectionError: null,

  theme: 'dark',
  sidebarCollapsed: false,
  activeTab: 'dashboard',
}

export const useTradingStore = create<TradingStore>()(
  devtools(
    subscribeWithSelector((set) => ({
      ...initialState,

      // Portfolio actions
      setPortfolios: (portfolios) => set({ portfolios }),
      setSelectedPortfolioId: (id) => set({ selectedPortfolioId: id }),
      setPortfolioLoading: (loading) => set({ portfolioLoading: loading }),
      setPortfolioError: (error) => set({ portfolioError: error }),
      updatePortfolio: (portfolioUpdate) => {
        set((state) => ({
          portfolios: state.portfolios.map((portfolio) =>
            portfolio.id === portfolioUpdate.portfolioId
              ? {
                  ...portfolio,
                  totalValue: portfolioUpdate.totalValue,
                  totalPnL: portfolioUpdate.totalPnL,
                  totalPnLPercent: portfolioUpdate.totalPnLPercent,
                  dayChange: portfolioUpdate.dayChange,
                  dayChangePercent: portfolioUpdate.dayChangePercent,
                  holdings: portfolioUpdate.updatedHoldings,
                  updatedAt: portfolioUpdate.timestamp,
                }
              : portfolio
          ),
        }))
      },

      // Market data actions
      setMarketData: (data) => {
        const marketDataMap = data.reduce((acc, item) => {
          acc[item.symbol] = item
          return acc
        }, {} as Record<string, MarketData>)
        set({ marketData: marketDataMap })
      },
      updateMarketData: (symbol, data) => {
        set((state) => ({
          marketData: {
            ...state.marketData,
            [symbol]: data,
          },
        }))
      },
      setMarketOverview: (overview) => set({ marketOverview: overview }),
      setMarketLoading: (loading) => set({ marketLoading: loading }),
      setMarketError: (error) => set({ marketError: error }),
      updatePrice: (priceUpdate) => {
        set((state) => {
          const existingMarketData = state.marketData[priceUpdate.symbol]

          return {
            priceUpdates: {
              ...state.priceUpdates,
              [priceUpdate.symbol]: priceUpdate,
            },
            marketData: {
              ...state.marketData,
              [priceUpdate.symbol]: existingMarketData ? {
                ...existingMarketData,
                currentPrice: priceUpdate.price,
                change24h: priceUpdate.changePercent,
                volume24h: priceUpdate.volume,
                lastUpdated: priceUpdate.timestamp,
              } : {
                symbol: priceUpdate.symbol,
                name: priceUpdate.symbol.replace('USDT', ''),
                currentPrice: priceUpdate.price,
                marketCap: 0,
                marketCapRank: 0,
                volume24h: priceUpdate.volume,
                change24h: priceUpdate.changePercent,
                change7d: 0,
                change30d: 0,
                high24h: priceUpdate.price,
                low24h: priceUpdate.price,
                circulatingSupply: 0,
                totalSupply: 0,
                maxSupply: 0,
                ath: priceUpdate.price,
                athDate: priceUpdate.timestamp,
                atl: priceUpdate.price,
                atlDate: priceUpdate.timestamp,
                lastUpdated: priceUpdate.timestamp,
              },
            },
          }
        })
      },

      // Strategy actions
      setStrategies: (strategies) => set({ strategies }),
      updateStrategy: (strategyId, updates) => {
        set((state) => ({
          strategies: state.strategies.map((strategy) =>
            strategy.id === strategyId ? { ...strategy, ...updates } : strategy
          ),
        }))
      },
      setStrategiesLoading: (loading) => set({ strategiesLoading: loading }),
      setStrategiesError: (error) => set({ strategiesError: error }),

      // Real-time update actions
      addSignalAlert: (alert) => {
        set((state) => ({
          signalAlerts: [alert, ...state.signalAlerts].slice(0, 50), // Keep last 50 alerts
        }))
      },
      addTradeExecution: (execution) => {
        set((state) => ({
          tradeExecutions: [execution, ...state.tradeExecutions].slice(0, 100), // Keep last 100 executions
        }))
      },
      addRiskAlert: (alert) => {
        set((state) => ({
          riskAlerts: [alert, ...state.riskAlerts].slice(0, 20), // Keep last 20 risk alerts
        }))
      },
      clearOldAlerts: () => {
        const oneHourAgo = Date.now() - 60 * 60 * 1000
        set((state) => ({
          signalAlerts: state.signalAlerts.filter(
            (alert) => new Date(alert.timestamp).getTime() > oneHourAgo
          ),
          riskAlerts: state.riskAlerts.filter(
            (alert) => new Date(alert.timestamp).getTime() > oneHourAgo
          ),
        }))
      },

      // Arbitrage actions
      setArbitrageOpportunities: (opportunities) => set({ arbitrageOpportunities: opportunities }),
      updateArbitrageOpportunity: (opportunity) => {
        set((state) => ({
          arbitrageOpportunities: state.arbitrageOpportunities.map((opp) =>
            opp.symbol === opportunity.symbol ? opportunity : opp
          ),
        }))
      },

      // WebSocket actions
      setConnectionStatus: (connected) => set({ isConnected: connected }),
      setConnectionError: (error) => set({ connectionError: error }),

      // UI actions
      setTheme: (theme) => set({ theme }),
      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      setActiveTab: (tab) => set({ activeTab: tab }),

      // Utility actions
      reset: () => set(initialState),
    })),
    {
      name: 'trading-store',
    }
  )
)

// Selectors for computed values
export const useSelectedPortfolio = () =>
  useTradingStore((state) =>
    state.portfolios.find((p) => p.id === state.selectedPortfolioId)
  )

export const useActiveStrategies = () =>
  useTradingStore((state) => state.strategies.filter((s) => s.status === 'active'))

export const useRecentSignalAlerts = (limit: number = 10) =>
  useTradingStore((state) => state.signalAlerts.slice(0, limit))

export const useTopArbitrageOpportunities = (limit: number = 5) =>
  useTradingStore((state) =>
    state.arbitrageOpportunities
      .sort((a, b) => b.spreadPercent - a.spreadPercent)
      .slice(0, limit)
  )
