'use client'

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react'
import { getMockWebSocketServer } from '@/lib/websocket'

export interface PriceData {
  symbol: string
  price: number
  change24h: number
  volume: number
  timestamp: number
}

export interface TradeData {
  symbol: string
  price: number
  amount: number
  side: 'buy' | 'sell'
  timestamp: number
}

export interface AIAgentData {
  agentId: string
  status: 'active' | 'inactive' | 'error'
  profit: number
  action: string
  timestamp: number
}

export interface PortfolioData {
  totalValue: number
  change24h: number
  assets: Array<{
    symbol: string
    amount: number
    value: number
    change24h: number
  }>
}

interface RealTimeDataContextType {
  prices: Map<string, PriceData>
  trades: TradeData[]
  aiAgents: Map<string, AIAgentData>
  portfolio: PortfolioData | null
  isConnected: boolean
  connectionStatus: 'connecting' | 'connected' | 'disconnected' | 'error'
  subscribe: (type: string, callback: (data: any) => void) => () => void
  getPrice: (symbol: string) => PriceData | null
  getRecentTrades: (symbol?: string, limit?: number) => TradeData[]
}

const RealTimeDataContext = createContext<RealTimeDataContextType | null>(null)

export function useRealTimeData() {
  const context = useContext(RealTimeDataContext)
  if (!context) {
    throw new Error('useRealTimeData must be used within a RealTimeDataProvider')
  }
  return context
}

interface RealTimeDataProviderProps {
  children: React.ReactNode
}

export function RealTimeDataProvider({ children }: RealTimeDataProviderProps) {
  const [prices, setPrices] = useState<Map<string, PriceData>>(new Map())
  const [trades, setTrades] = useState<TradeData[]>([])
  const [aiAgents, setAIAgents] = useState<Map<string, AIAgentData>>(new Map())
  const [portfolio, setPortfolio] = useState<PortfolioData | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('disconnected')

  // Initialize mock data immediately
  useEffect(() => {
    console.log('🚀 Initializing Go Coffee mock data...')

    // Set initial prices
    const initialPrices = new Map([
      ['BTC/USDT', { symbol: 'BTC/USDT', price: 45234.56, change24h: 2.34, volume: 1234567890, timestamp: Date.now() }],
      ['ETH/USDT', { symbol: 'ETH/USDT', price: 2834.12, change24h: 1.87, volume: 987654321, timestamp: Date.now() }],
      ['SOL/USDT', { symbol: 'SOL/USDT', price: 98.45, change24h: -0.56, volume: 456789123, timestamp: Date.now() }],
      ['COFFEE/USDT', { symbol: 'COFFEE/USDT', price: 12.34, change24h: 5.67, volume: 123456789, timestamp: Date.now() }]
    ])
    setPrices(initialPrices)
    console.log('✅ Initial prices set:', Array.from(initialPrices.entries()))

    // Set initial portfolio
    setPortfolio({
      totalValue: 123456.78,
      change24h: 2.45,
      assets: [
        { symbol: 'BTC', amount: 2.5432, value: 115234.56, change24h: 2.34 },
        { symbol: 'ETH', amount: 45.8901, value: 130234.12, change24h: 1.87 },
        { symbol: 'SOL', amount: 234.5678, value: 23098.45, change24h: -0.56 },
        { symbol: 'COFFEE', amount: 1000.0, value: 12340.00, change24h: 5.67 }
      ]
    })

    // Set initial AI agents
    const initialAgents = new Map([
      ['trading-alpha', { agentId: 'trading-alpha', status: 'active' as const, profit: 2345, action: 'Monitoring markets', timestamp: Date.now() }],
      ['market-analyzer', { agentId: 'market-analyzer', status: 'active' as const, profit: 1892, action: 'Analyzing trends', timestamp: Date.now() }],
      ['risk-manager', { agentId: 'risk-manager', status: 'active' as const, profit: 456, action: 'Managing risk', timestamp: Date.now() }]
    ])
    setAIAgents(initialAgents)

    // Set connection status to connected immediately
    setIsConnected(true)
    setConnectionStatus('connected')
    console.log('Mock data initialization complete')
  }, [])

  // Connect to mock WebSocket server
  useEffect(() => {
    console.log('Starting mock WebSocket server...')
    setConnectionStatus('connecting')

    const mockServer = getMockWebSocketServer()

    const unsubscribe = mockServer.subscribe((data) => {
      console.log('Received mock data:', data)
      handleWebSocketMessage(data)
    })

    mockServer.start()

    // Set connected status after a short delay to simulate connection
    setTimeout(() => {
      setIsConnected(true)
      setConnectionStatus('connected')
      console.log('Mock WebSocket server connected')
    }, 100)

    return () => {
      console.log('Stopping mock WebSocket server...')
      unsubscribe()
      mockServer.stop()
      setIsConnected(false)
      setConnectionStatus('disconnected')
    }
  }, [])

  const handleWebSocketMessage = useCallback((data: any) => {
    const { type, payload } = data

    switch (type) {
      case 'price_update':
        setPrices(prev => {
          const newPrices = new Map(prev)
          newPrices.set(payload.symbol, payload)
          return newPrices
        })
        
        // Update portfolio based on price changes
        setPortfolio(prev => {
          if (!prev) return prev
          
          const updatedAssets = prev.assets.map(asset => {
            const priceData = prices.get(`${asset.symbol}/USDT`)
            if (priceData) {
              return {
                ...asset,
                value: asset.amount * priceData.price,
                change24h: priceData.change24h
              }
            }
            return asset
          })

          const totalValue = updatedAssets.reduce((sum, asset) => sum + asset.value, 0)
          const totalChange24h = updatedAssets.reduce((sum, asset) => sum + (asset.value * asset.change24h / 100), 0)

          return {
            ...prev,
            totalValue,
            change24h: (totalChange24h / totalValue) * 100,
            assets: updatedAssets
          }
        })
        break

      case 'trade_update':
        setTrades(prev => {
          const newTrades = [payload, ...prev.slice(0, 99)] // Keep last 100 trades
          return newTrades
        })
        break

      case 'ai_update':
        setAIAgents(prev => {
          const newAgents = new Map(prev)
          const existingAgent = newAgents.get(payload.agentId)
          if (existingAgent) {
            newAgents.set(payload.agentId, {
              ...existingAgent,
              ...payload
            })
          }
          return newAgents
        })
        break

      default:
        console.log('Unknown message type:', type)
    }
  }, [prices])

  const subscribe = useCallback((type: string, callback: (data: any) => void) => {
    // For mock implementation, we'll just call the callback with current data
    const interval = setInterval(() => {
      switch (type) {
        case 'prices':
          callback(Array.from(prices.values()))
          break
        case 'trades':
          callback(trades.slice(0, 10))
          break
        case 'portfolio':
          callback(portfolio)
          break
        case 'ai_agents':
          callback(Array.from(aiAgents.values()))
          break
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [prices, trades, portfolio, aiAgents])

  const getPrice = useCallback((symbol: string): PriceData | null => {
    return prices.get(symbol) || null
  }, [prices])

  const getRecentTrades = useCallback((symbol?: string, limit: number = 10): TradeData[] => {
    let filteredTrades = trades
    if (symbol) {
      filteredTrades = trades.filter(trade => trade.symbol === symbol)
    }
    return filteredTrades.slice(0, limit)
  }, [trades])

  const contextValue: RealTimeDataContextType = {
    prices,
    trades,
    aiAgents,
    portfolio,
    isConnected,
    connectionStatus,
    subscribe,
    getPrice,
    getRecentTrades
  }

  return (
    <RealTimeDataContext.Provider value={contextValue}>
      {children}
    </RealTimeDataContext.Provider>
  )
}

// Custom hooks for specific data types
export function usePrices() {
  const { prices, getPrice } = useRealTimeData()
  return { prices: Array.from(prices.values()), getPrice }
}

export function useTrades(symbol?: string, limit?: number) {
  const { getRecentTrades } = useRealTimeData()
  return getRecentTrades(symbol, limit)
}

export function usePortfolio() {
  const { portfolio } = useRealTimeData()
  return portfolio
}

export function useAIAgents() {
  const { aiAgents } = useRealTimeData()
  return Array.from(aiAgents.values())
}

export function useConnectionStatus() {
  const { isConnected, connectionStatus } = useRealTimeData()
  return { isConnected, connectionStatus }
}
