'use client'

import useSWR from 'swr'
import { 
  dashboardAPI, 
  coffeeAPI, 
  defiAPI, 
  agentsAPI, 
  scrapingAPI, 
  analyticsAPI 
} from '@/lib/api'

// Generic fetcher function
const fetcher = (fn: () => Promise<any>) => fn()

// Dashboard hooks
export function useDashboardMetrics() {
  return useSWR('dashboard-metrics', () => dashboardAPI.getMetrics(), {
    refreshInterval: 30000, // Refresh every 30 seconds
    revalidateOnFocus: true,
  })
}

export function useDashboardActivity() {
  return useSWR('dashboard-activity', () => dashboardAPI.getActivity(), {
    refreshInterval: 60000, // Refresh every minute
  })
}

// Coffee hooks
export function useCoffeeOrders(params?: any) {
  const key = params ? ['coffee-orders', params] : 'coffee-orders'
  return useSWR(key, () => coffeeAPI.getOrders(params), {
    refreshInterval: 15000, // Refresh every 15 seconds
  })
}

export function useCoffeeInventory() {
  return useSWR('coffee-inventory', () => coffeeAPI.getInventory(), {
    refreshInterval: 60000,
  })
}

// DeFi hooks
export function useDefiPortfolio() {
  return useSWR('defi-portfolio', () => defiAPI.getPortfolio(), {
    refreshInterval: 10000, // Refresh every 10 seconds for crypto data
  })
}

export function useDefiAssets() {
  return useSWR('defi-assets', () => defiAPI.getAssets(), {
    refreshInterval: 10000,
  })
}

export function useDefiStrategies() {
  return useSWR('defi-strategies', () => defiAPI.getStrategies(), {
    refreshInterval: 30000,
  })
}

// AI Agents hooks
export function useAgentsStatus() {
  return useSWR('agents-status', () => agentsAPI.getStatus(), {
    refreshInterval: 5000, // Refresh every 5 seconds for agent status
  })
}

export function useAgentLogs(agentId: string) {
  return useSWR(
    agentId ? ['agent-logs', agentId] : null,
    () => agentsAPI.getAgentLogs(agentId),
    {
      refreshInterval: 10000,
    }
  )
}

// Scraping hooks (Bright Data)
export function useMarketData(params?: any) {
  const key = params ? ['market-data', params] : 'market-data'
  return useSWR(key, () => scrapingAPI.getMarketData(params), {
    refreshInterval: 120000, // Refresh every 2 minutes
  })
}

export function useDataSources() {
  return useSWR('data-sources', () => scrapingAPI.getDataSources(), {
    refreshInterval: 300000, // Refresh every 5 minutes
  })
}

// Analytics hooks
export function useSalesData(timeRange?: string) {
  const key = timeRange ? ['sales-data', timeRange] : 'sales-data'
  return useSWR(key, () => analyticsAPI.getSalesData({ timeRange }), {
    refreshInterval: 60000,
  })
}

export function useRevenueData(timeRange?: string) {
  const key = timeRange ? ['revenue-data', timeRange] : 'revenue-data'
  return useSWR(key, () => analyticsAPI.getRevenueData({ timeRange }), {
    refreshInterval: 60000,
  })
}

export function useTopProducts() {
  return useSWR('top-products', () => analyticsAPI.getTopProducts(), {
    refreshInterval: 300000, // Refresh every 5 minutes
  })
}

export function useLocationPerformance() {
  return useSWR('location-performance', () => analyticsAPI.getLocationPerformance(), {
    refreshInterval: 300000,
  })
}

// Mutation hooks for actions
export function useRefreshMarketData() {
  const { mutate } = useSWR('market-data')
  
  const refresh = async () => {
    try {
      await scrapingAPI.refreshData()
      // Revalidate market data after refresh
      mutate()
      return { success: true }
    } catch (error) {
      return { success: false, error }
    }
  }
  
  return { refresh }
}

export function useToggleAgent() {
  const { mutate } = useSWR('agents-status')
  
  const toggle = async (agentId: string) => {
    try {
      await agentsAPI.toggleAgent(agentId)
      // Revalidate agents status after toggle
      mutate()
      return { success: true }
    } catch (error) {
      return { success: false, error }
    }
  }
  
  return { toggle }
}

export function useToggleStrategy() {
  const { mutate } = useSWR('defi-strategies')
  
  const toggle = async (strategyId: string) => {
    try {
      await defiAPI.toggleStrategy(strategyId)
      // Revalidate strategies after toggle
      mutate()
      return { success: true }
    } catch (error) {
      return { success: false, error }
    }
  }
  
  return { toggle }
}
